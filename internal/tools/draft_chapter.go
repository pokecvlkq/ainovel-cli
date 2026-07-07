package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"unicode/utf8"

	"github.com/voocel/agentcore/schema"
	"github.com/voocel/ainovel-cli/internal/domain"
	"github.com/voocel/ainovel-cli/internal/errs"
	"github.com/voocel/ainovel-cli/internal/store"
)

// DraftChapterTool Ghi bản nháp toàn chương, thay thế luồng xử lý cũ write_scene + polish_chapter.
// Agent tự quyết định viết xong một lần hay viết tiếp thành từng đợt.
type DraftChapterTool struct {
	store *store.Store
}

func NewDraftChapterTool(store *store.Store) *DraftChapterTool {
	return &DraftChapterTool{store: store}
}

func (t *DraftChapterTool) Name() string { return "draft_chapter" }
func (t *DraftChapterTool) Description() string {
	return "Viết nội dung chương. mode=write ghi đè toàn chương, mode=append ghi thêm vào bản nháp hiện có (viết tiếp/sửa đổi)"
}
func (t *DraftChapterTool) Label() string { return "Viết chương" }

// Công cụ ghi, cấm chạy đồng thời (vấn đề đọc-sửa-ghi).
func (t *DraftChapterTool) ReadOnly(_ json.RawMessage) bool        { return false }
func (t *DraftChapterTool) ConcurrencySafe(_ json.RawMessage) bool { return false }

func (t *DraftChapterTool) Schema() map[string]any {
	// mode đánh dấu là required để tương thích với OpenAI strict tool calling —— chế độ strict
	// yêu cầu tất cả các thuộc tính phải nằm trong danh sách required. Ban đầu "bỏ qua mode sẽ mặc định là write"
	// Hành vi "mặc định" bây giờ yêu cầu mô hình phải truyền trực tiếp mode="write", nhánh default của Execute giữ nguyên.
	return schema.Object(
		schema.Property("chapter", schema.Int("Số chương")).Required(),
		schema.Property("content", schema.String("Nội dung chương")).Required(),
		schema.Property("mode", schema.Enum("Chế độ ghi", "write", "append")).Required(),
	)
}

// StrictSchema kích hoạt OpenAI strict tool calling, yêu cầu mô hình phải tuân thủ nghiêm ngặt
// schema: tất cả các trường required phải điền, arguments không được "kết thúc sớm (EOT)" và xuất hiện một object rỗng.
// litellm chuyển trực tiếp trường strict; các backend như OpenAI / xAI hỗ trợ tính năng này sẽ thi hành một cách bắt buộc, các backend khác
// sẽ bỏ qua các trường chưa biết theo quy ước HTTP/JSON. Anthropic/Gemini/Bedrock thực hiện luồng chuyển đổi riêng
// tất nhiên sẽ không thấy trường này.
func (t *DraftChapterTool) StrictSchema() bool { return false }

func (t *DraftChapterTool) Execute(_ context.Context, args json.RawMessage) (json.RawMessage, error) {
	var a struct {
		Chapter int    `json:"chapter"`
		Content string `json:"content"`
		Mode    string `json:"mode"`
	}
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, fmt.Errorf("invalid args: %w: %w", errs.ErrToolArgs, err)
	}
	if a.Chapter <= 0 {
		return nil, fmt.Errorf("chapter must be > 0: %w", errs.ErrToolArgs)
	}
	if a.Content == "" {
		return nil, fmt.Errorf("content must not be empty: %w", errs.ErrToolArgs)
	}
	if err := t.store.Progress.ValidateChapterWork(a.Chapter); err != nil {
		return nil, err
	}
	if err := EnsureChapterExpanded(t.store, a.Chapter); err != nil {
		return nil, err
	}
	if t.store.Progress.IsChapterCompleted(a.Chapter) {
		// Đường dẫn trau chuốt/viết lại: Mặc dù chương đã hoàn thành nhưng vẫn còn trong pending_rewrites, vẫn cho phép ghi đè bản nháp
		progress, _ := t.store.Progress.Load()
		inRewriteQueue := progress != nil && slices.Contains(progress.PendingRewrites, a.Chapter)
		if !inRewriteQueue {
			return json.Marshal(map[string]any{
				"chapter":   a.Chapter,
				"skipped":   true,
				"completed": true,
				"reason":    fmt.Sprintf("Chương %d đã được gửi xong, không thể ghi đè", a.Chapter),
			})
		}
	}
	if err := t.store.Progress.StartChapter(a.Chapter); err != nil {
		return nil, fmt.Errorf("mark chapter in progress: %w", err)
	}

	switch a.Mode {
	case "append":
		if err := t.store.Drafts.AppendDraft(a.Chapter, a.Content); err != nil {
			return nil, fmt.Errorf("append draft: %w", err)
		}
		full, err := t.store.Drafts.LoadDraft(a.Chapter)
		if err != nil {
			return nil, fmt.Errorf("load draft after append: %w", err)
		}
		if _, err := t.store.Checkpoints.AppendArtifact(
			domain.ChapterScope(a.Chapter), "draft",
			fmt.Sprintf("drafts/%02d.draft.md", a.Chapter),
		); err != nil {
			return nil, fmt.Errorf("checkpoint draft: %w", err)
		}
		return json.Marshal(map[string]any{
			"written":    true,
			"chapter":    a.Chapter,
			"mode":       "append",
			"word_count": utf8.RuneCountInString(full),
			"next_step":  "Gọi read_chapter(source=draft) để đọc lại bản nháp, sau đó gọi check_consistency, cuối cùng là commit_chapter",
		})
	default: // write
		if err := t.store.Drafts.SaveDraft(a.Chapter, a.Content); err != nil {
			return nil, fmt.Errorf("save draft: %w", err)
		}
		if _, err := t.store.Checkpoints.AppendArtifact(
			domain.ChapterScope(a.Chapter), "draft",
			fmt.Sprintf("drafts/%02d.draft.md", a.Chapter),
		); err != nil {
			return nil, fmt.Errorf("checkpoint draft: %w", err)
		}
		return json.Marshal(map[string]any{
			"written":    true,
			"chapter":    a.Chapter,
			"mode":       "write",
			"word_count": utf8.RuneCountInString(a.Content),
			"next_step":  "Gọi read_chapter(source=draft) để đọc lại bản nháp, sau đó gọi check_consistency, cuối cùng là commit_chapter",
		})
	}
}
