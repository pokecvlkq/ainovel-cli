package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/voocel/agentcore/schema"
	"github.com/voocel/ainovel-cli/internal/domain"
	"github.com/voocel/ainovel-cli/internal/errs"
	"github.com/voocel/ainovel-cli/internal/store"
)

// ReopenBookTool Mở lại sách đã hoàn thành để quay lại trạng thái làm lại (chỉ Coordinator được dùng).
// Sau khi hoàn thành cuốn sách, completePhaseGate chặn tất cả việc giao nhiệm vụ subagent, người dùng không thể làm lại các chương đã viết.
// Công cụ này không phải là subagent, giai đoạn complete có thể gọi được: nó sẽ chuyển phase về writing một cách nguyên tử, đưa chương mục tiêu
// vào PendingRewrites, flow=rewriting, sau đó Flow Router sẽ theo hàng đợi làm lại để giao cho writer viết lại từng chương,
// khi hàng đợi chạy xong commit_chapter sẽ tự động kết thúc lại. Các logic cốt lõi như Gate / Router / edit / commit đều không cần thay đổi.
type ReopenBookTool struct {
	store *store.Store
}

func NewReopenBookTool(s *store.Store) *ReopenBookTool {
	return &ReopenBookTool{store: s}
}

func (t *ReopenBookTool) Name() string  { return "reopen_book" }
func (t *ReopenBookTool) Label() string { return "Mở lại làm lại" }

func (t *ReopenBookTool) Description() string {
	return "Mở lại sách đã hoàn thành (phase=complete) để chuyển sang trạng thái làm lại, dùng khi người dùng yêu cầu viết lại/đánh bóng một số chương sau khi hoàn thành sách." +
		"chapters là danh sách số chương đã hoàn thành cần làm lại; sau khi gọi công cụ này, các chương sẽ vào hàng đợi viết lại, Host sẽ cử writer làm lại từng chương, khi xong tất cả sẽ tự động hoàn tất trở lại." +
		"Chỉ dùng khi toàn bộ cuốn sách đã hoàn thành, và người dùng yêu cầu rõ ràng sửa các chương đã viết; người dùng muốn thêm cốt truyện/mở rộng độ dài không thuộc trường hợp làm lại, không dùng công cụ này."
}

// Công cụ ghi, cấm chạy đồng thời.
func (t *ReopenBookTool) ReadOnly(_ json.RawMessage) bool        { return false }
func (t *ReopenBookTool) ConcurrencySafe(_ json.RawMessage) bool { return false }

func (t *ReopenBookTool) ActivityDescription(_ json.RawMessage) string {
	return "Mở lại toàn bộ cuốn sách để làm lại"
}

func (t *ReopenBookTool) Schema() map[string]any {
	return schema.Object(
		schema.Property("chapters", schema.Array("Danh sách các số chương đã hoàn thành cần làm lại (ít nhất một chương)", schema.Int(""))).Required(),
		schema.Property("reason", schema.String("Lý do làm lại (tùy chọn, ví dụ \"dọn dẹp ký tự đặc biệt\")")),
	)
}

func (t *ReopenBookTool) Execute(_ context.Context, args json.RawMessage) (json.RawMessage, error) {
	var a struct {
		Chapters []int  `json:"chapters"`
		Reason   string `json:"reason"`
	}
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, fmt.Errorf("invalid args: %w: %w", errs.ErrToolArgs, err)
	}
	if len(a.Chapters) == 0 {
		return nil, fmt.Errorf("chapters không thể trống, phải chỉ định chương cần làm lại: %w", errs.ErrToolArgs)
	}

	progress, err := t.store.Progress.Load()
	if err != nil {
		return nil, fmt.Errorf("load progress: %w: %w", errs.ErrStoreRead, err)
	}
	if progress == nil {
		return nil, fmt.Errorf("progress chưa được khởi tạo: %w", errs.ErrToolPrecondition)
	}
	// Chỉ có thể làm lại các chương đã viết; các chương không nằm trong danh sách đã hoàn thành là viết tiếp/vượt quá giới hạn, từ chối rõ ràng và hướng dẫn người dùng điều chỉnh độ dài.
	var invalid []int
	for _, ch := range a.Chapters {
		if !slices.Contains(progress.CompletedChapters, ch) {
			invalid = append(invalid, ch)
		}
	}
	if len(invalid) > 0 {
		return nil, fmt.Errorf("Chương %v chưa viết xong, reopen chỉ có thể làm lại các chương đã hoàn thành (nếu thêm/mở rộng cốt truyện vui lòng điều chỉnh độ dài): %w", invalid, errs.ErrToolPrecondition)
	}

	// Xác nhận trước về phase sẽ được store.Reopen lo (chỉ cho phép gọi ở complete).
	if err := t.store.Progress.Reopen(a.Chapters, a.Reason); err != nil {
		return nil, fmt.Errorf("reopen: %w: %w", errs.ErrStoreWrite, err)
	}

	// checkpoint: đối xứng với complete_book (GlobalScope + meta/progress.json).
	if _, err := t.store.Checkpoints.AppendArtifact(domain.GlobalScope(), "reopen", "meta/progress.json"); err != nil {
		return nil, fmt.Errorf("checkpoint reopen: %w: %w", errs.ErrStoreWrite, err)
	}

	return json.Marshal(map[string]any{
		"reopened":         true,
		"phase":            string(domain.PhaseWriting),
		"pending_rewrites": a.Chapters,
		"next_step":        "Đã mở lại và đưa các chương mục tiêu vào hàng đợi. Vui lòng đợi Host chỉ đạo phân phối writer làm lại từng chương; khi sửa xong toàn bộ sẽ tự động hoàn thành trở lại.",
	})
}
