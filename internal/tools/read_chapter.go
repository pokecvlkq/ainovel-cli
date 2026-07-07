package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/voocel/agentcore/schema"
	"github.com/voocel/ainovel-cli/internal/store"
)

// ReadChapterTool Đọc văn bản chương, cho phép Agent đọc lại văn bản của mình và phần trước đó.
type ReadChapterTool struct {
	store *store.Store
}

func NewReadChapterTool(store *store.Store) *ReadChapterTool {
	return &ReadChapterTool{store: store}
}

func (t *ReadChapterTool) Name() string { return "read_chapter" }
func (t *ReadChapterTool) Description() string {
	return "Đọc văn bản gốc của chương. Có thể đọc bản cuối cùng, bản nháp, hoặc trích xuất đoạn hội thoại của nhân vật"
}
func (t *ReadChapterTool) Label() string { return "Đọc chương" }

// Công cụ chỉ đọc, có thể được lên lịch đồng thời (thường đọc nhiều chương cùng lúc khi editor kiểm duyệt).
func (t *ReadChapterTool) ReadOnly(_ json.RawMessage) bool        { return true }
func (t *ReadChapterTool) ConcurrencySafe(_ json.RawMessage) bool { return true }

func (t *ReadChapterTool) Schema() map[string]any {
	return schema.Object(
		schema.Property("chapter", schema.Int("Số chương (Bắt buộc khi đọc một chương)")),
		schema.Property("from", schema.Int("Số chương bắt đầu (Dùng khi đọc một khoảng)")),
		schema.Property("to", schema.Int("Số chương kết thúc (Dùng khi đọc một khoảng)")),
		schema.Property("source", schema.Enum("Nguồn", "final", "draft")).Required(),
		schema.Property("character", schema.String("Tên nhân vật (Dùng khi trích xuất đoạn hội thoại)")),
		schema.Property("max_runes", schema.Int("Số lượng ký tự tối đa mỗi chương (Cắt bớt khi đọc một khoảng, mặc định 2000)")),
	)
}

func (t *ReadChapterTool) Execute(_ context.Context, args json.RawMessage) (json.RawMessage, error) {
	var a struct {
		Chapter   int    `json:"chapter"`
		From      int    `json:"from"`
		To        int    `json:"to"`
		Source    string `json:"source"`
		Character string `json:"character"`
		MaxRunes  int    `json:"max_runes"`
	}
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, fmt.Errorf("invalid args: %w", err)
	}

	// Chế độ 1: Trích xuất hội thoại nhân vật
	if a.Character != "" {
		chars, _ := t.store.Characters.Load()
		var aliases []string
		for _, c := range chars {
			if c.Name == a.Character {
				aliases = c.Aliases
				break
			}
		}
		var maxCompleted int
		if p, _ := t.store.Progress.Load(); p != nil {
			maxCompleted = maxCompletedChapter(p.CompletedChapters)
		}
		samples := t.store.Drafts.ExtractDialogue(a.Character, aliases, 8, maxCompleted)
		result := map[string]any{
			"character": a.Character,
			"samples":   samples,
		}
		if len(samples) == 0 {
			result["hint"] = "Nhân vật này chưa có mẫu hội thoại nào, không cần thử lại, hãy chuyển thẳng sang bước tiếp theo"
		}
		return json.Marshal(result)
	}

	// Chế độ 2: Đọc theo phạm vi
	if a.From > 0 && a.To > 0 {
		maxRunes := a.MaxRunes
		if maxRunes <= 0 {
			maxRunes = 2000
		}
		texts, err := t.store.Drafts.LoadChapterRange(a.From, a.To, maxRunes)
		if err != nil {
			return nil, fmt.Errorf("load chapter range: %w", err)
		}
		return json.Marshal(map[string]any{
			"chapters": texts,
			"from":     a.From,
			"to":       a.To,
		})
	}

	// Chế độ 3: Đọc chương đơn lẻ
	if a.Chapter <= 0 {
		return nil, fmt.Errorf("chapter is required")
	}

	var content string
	var err error
	switch a.Source {
	case "draft":
		content, err = t.store.Drafts.LoadDraft(a.Chapter)
	default: // final
		content, err = t.store.Drafts.LoadChapterText(a.Chapter)
		if err == nil && content == "" {
			slog.Warn("read_chapter Bản cuối cùng đọc được rỗng, quay lại bản nháp", "module", "tool", "chapter", a.Chapter)
			content, err = t.store.Drafts.LoadDraft(a.Chapter)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("read chapter %d: %w", a.Chapter, err)
	}
	if content == "" {
		return json.Marshal(map[string]any{
			"chapter": a.Chapter,
			"exists":  false,
			"hint":    "Chương này chưa được viết, nếu cần viết vui lòng gọi draft_chapter trước",
		})
	}

	return json.Marshal(map[string]any{
		"chapter":    a.Chapter,
		"content":    content,
		"word_count": len([]rune(content)),
	})
}

// maxCompletedChapter Trả về số chương lớn nhất trong danh sách các chương đã hoàn thành.
func maxCompletedChapter(completed []int) int {
	m := 0
	for _, ch := range completed {
		if ch > m {
			m = ch
		}
	}
	return m
}
