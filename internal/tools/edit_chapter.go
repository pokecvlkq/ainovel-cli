package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/voocel/agentcore/schema"
	agentcoretools "github.com/voocel/agentcore/tools"
	"github.com/voocel/ainovel-cli/internal/domain"
	"github.com/voocel/ainovel-cli/internal/errs"
	"github.com/voocel/ainovel-cli/internal/store"
)

// EditChapterTool Công cụ thay thế chuỗi tại vị trí cố định cho bản nháp chương, phù hợp với ngữ cảnh trau chuốt.
// So với việc viết lại toàn bộ chương của draft_chapter, token tiết kiệm hơn 10x+.
//
// Giao ước lưu trữ: chỉ sửa drafts/{ch:02d}.draft.md, cấm sửa trực tiếp chapters/ (bản cuối do commit_chapter độc chiếm).
// Ngữ nghĩa Seed: drafts không tồn tại nhưng chapters có → tự động sao chép chapters vào drafts làm điểm bắt đầu.
// Kiểm tra thuộc tính: khi chương đã hoàn thành phải nằm trong hàng đợi PendingRewrites, nếu không sẽ bị từ chối.
//
// Công cụ này là một lớp bọc mỏng của agentcore.EditTool, logic tìm-thay thế (khớp lỗi nhiều cấp, xuất diff, giữ lại cuối dòng/BOM)
// đều sử dụng lại việc triển khai thượng nguồn.
type EditChapterTool struct {
	store *store.Store
	edit  *agentcoretools.EditTool
}

func NewEditChapterTool(s *store.Store) *EditChapterTool {
	return &EditChapterTool{
		store: s,
		edit:  agentcoretools.NewEdit(s.Dir(), nil),
	}
}

func (t *EditChapterTool) Name() string  { return "edit_chapter" }
func (t *EditChapterTool) Label() string { return "Chỉnh sửa chương" }

// ReadOnly Khai báo rõ ràng công cụ ghi (kết hợp với ConcurrencySafeTool để ngăn chặn việc lên lịch đồng thời).
func (t *EditChapterTool) ReadOnly(_ json.RawMessage) bool { return false }

// ConcurrencySafe Cấm đồng thời một cách rõ ràng: gọi edit_chapter nhiều lần trên cùng một chương sẽ gây ra xung đột đọc-sửa-ghi,
// ngay cả khi song song ở các chương khác nhau cũng sẽ làm xáo trộn thứ tự checkpoint. Tuần tự hóa đồng nhất là ổn định nhất.
func (t *EditChapterTool) ConcurrencySafe(_ json.RawMessage) bool { return false }

// ActivityDescription Cung cấp mô tả hoạt động của công cụ hiện tại để hiển thị trên UI/log.
func (t *EditChapterTool) ActivityDescription(_ json.RawMessage) string {
	return "Chỉnh sửa bản nháp chương"
}

func (t *EditChapterTool) Description() string {
	return "Thay thế chuỗi tại vị trí cố định cho bản nháp chương (ưu tiên cho ngữ cảnh trau chuốt, tiết kiệm token hơn việc viết lại toàn bộ chương của draft_chapter)." +
		"Tìm old_string và thay thế bằng new_string, yêu cầu khớp chính xác và duy nhất (nếu khớp nhiều chỗ cần replace_all=true)." +
		"Ghi vào drafts/{ch}.draft.md; khi drafts không tồn tại sẽ tự động lấy từ chapters." +
		"Từ chối thực thi khi chương đã hoàn thành và không nằm trong hàng đợi PendingRewrites. Mỗi lần gọi chỉ sửa một chỗ, nếu cần sửa nhiều chỗ vui lòng gọi nhiều lần."
}

func (t *EditChapterTool) Schema() map[string]any {
	return schema.Object(
		schema.Property("chapter", schema.Int("Số chương")).Required(),
		schema.Property("old_string", schema.String("Đoạn văn bản gốc chính xác cần thay thế, nếu có nhiều dòng cần bao gồm ký tự xuống dòng; nếu không thêm replace_all thì phải xuất hiện duy nhất trong bản nháp")).Required(),
		schema.Property("new_string", schema.String("Văn bản mới sau khi thay thế")).Required(),
		schema.Property("replace_all", schema.Bool("Thay thế tất cả các kết quả khớp (mặc định là false)")),
	)
}

func (t *EditChapterTool) Execute(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	var a struct {
		Chapter    int    `json:"chapter"`
		OldString  string `json:"old_string"`
		NewString  string `json:"new_string"`
		ReplaceAll bool   `json:"replace_all"`
	}
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, fmt.Errorf("invalid args: %w: %w", errs.ErrToolArgs, err)
	}
	if a.Chapter <= 0 {
		return nil, fmt.Errorf("chapter must be > 0: %w", errs.ErrToolArgs)
	}
	if a.OldString == "" {
		return nil, fmt.Errorf("old_string không được để trống: %w", errs.ErrToolArgs)
	}
	if a.OldString == a.NewString {
		return nil, fmt.Errorf("old_string và new_string giống nhau, không cần sửa đổi: %w", errs.ErrToolArgs)
	}

	// Kiểm tra thuộc tính: chương đã hoàn thành phải nằm trong hàng đợi viết lại để tránh làm bẩn bản cuối
	if t.store.Progress.IsChapterCompleted(a.Chapter) {
		progress, _ := t.store.Progress.Load()
		if progress == nil || !slices.Contains(progress.PendingRewrites, a.Chapter) {
			return nil, fmt.Errorf("Chương %d đã hoàn thành và không nằm trong hàng đợi PendingRewrites, không thể chỉnh sửa; nếu cần sửa đổi, vui lòng để editor đánh giá kích hoạt viết lại/trau chuốt trước: %w", a.Chapter, errs.ErrToolPrecondition)
		}
	}

	// Seed: khi drafts không tồn tại, sao chép một bản từ chapters làm điểm bắt đầu
	if err := t.ensureDraft(a.Chapter); err != nil {
		return nil, err
	}

	// Ủy quyền cho agentcore.EditTool hoàn thành việc tìm-thay thế
	subArgs, _ := json.Marshal(map[string]any{
		"path":        fmt.Sprintf("drafts/%02d.draft.md", a.Chapter),
		"file_path":   fmt.Sprintf("drafts/%02d.draft.md", a.Chapter),
		"old_text":    a.OldString,
		"old_string":  a.OldString,
		"new_text":    a.NewString,
		"new_string":  a.NewString,
		"replace_all": a.ReplaceAll,
	})
	result, err := t.edit.Execute(ctx, subArgs)
	if err != nil {
		return nil, fmt.Errorf("apply edit: %w: %w", errs.ErrToolPrecondition, err)
	}

	if _, err := t.store.Checkpoints.AppendArtifact(
		domain.ChapterScope(a.Chapter), "edit",
		fmt.Sprintf("drafts/%02d.draft.md", a.Chapter),
	); err != nil {
		return nil, fmt.Errorf("checkpoint edit: %w: %w", errs.ErrStoreWrite, err)
	}

	// Hướng dẫn bổ sung: cho writer biết các bước tiếp theo, tránh bỏ sót check_consistency / commit_chapter
	var passthrough map[string]any
	if err := json.Unmarshal(result, &passthrough); err != nil {
		return result, nil
	}
	passthrough["chapter"] = a.Chapter
	passthrough["next_step"] = "edit đã được lưu. Nếu vẫn còn lỗi nghiêm trọng có thể gọi lại edit_chapter; nếu không thì check_consistency sau đó commit_chapter"
	return json.Marshal(passthrough)
}

// ensureDraft đảm bảo drafts/{ch}.draft.md tồn tại:
//   - Đã có bản nháp → trả về trực tiếp
//   - Không có bản nháp nhưng có bản cuối → sao chép bản cuối vào drafts làm điểm bắt đầu sửa đổi (thường thấy trong ngữ cảnh trau chuốt)
//   - Cả hai đều không có → báo lỗi, nhắc nhở sử dụng draft_chapter để tạo bản nháp ban đầu trước
func (t *EditChapterTool) ensureDraft(chapter int) error {
	draft, err := t.store.Drafts.LoadDraft(chapter)
	if err != nil {
		return fmt.Errorf("load draft: %w: %w", errs.ErrStoreRead, err)
	}
	if draft != "" {
		return nil
	}
	text, err := t.store.Drafts.LoadChapterText(chapter)
	if err != nil {
		return fmt.Errorf("load chapter: %w: %w", errs.ErrStoreRead, err)
	}
	if text == "" {
		return fmt.Errorf("Chương %d không có bản nháp cũng không có bản cuối, vui lòng gọi draft_chapter(mode=write, chapter=%d) để tạo bản nháp ban đầu trước: %w", chapter, chapter, errs.ErrToolPrecondition)
	}
	if err := t.store.Drafts.SaveDraft(chapter, text); err != nil {
		return fmt.Errorf("seed draft from chapter: %w: %w", errs.ErrStoreWrite, err)
	}
	return nil
}
