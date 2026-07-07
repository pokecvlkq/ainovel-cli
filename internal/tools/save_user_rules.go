package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/voocel/agentcore/schema"
	"github.com/voocel/ainovel-cli/internal/errs"
	"github.com/voocel/ainovel-cli/internal/rules"
	"github.com/voocel/ainovel-cli/internal/userrules"
)

// SaveUserRulesTool lưu trữ lâu dài các yêu cầu "phong cách viết/chất lượng" của người dùng (chỉ Coordinator giữ).
//
// Đây là cổng thống nhất cho các quy tắc viết dài hạn: các quy tắc phong cách/chất lượng ràng buộc ngòi bút của writer,
// luôn có giá trị (ví dụ: "mỗi chương 1500 chữ", "ít dùng phép ẩn dụ", "cấm dùng 'ở một mức độ nào đó'", "tỉ lệ đối thoại cao", "nhân vật chính tổng thể bình tĩnh, kiềm chế")
// được LLM chuẩn hóa thành ràng buộc có cấu trúc và ghi vào snapshot của sách tại meta/user_rules.json,
// novel_context tiêm vào working_memory.user_rules, commit_chapter dựa vào đó kiểm tra cơ học.
// Chỉnh sửa cốt truyện/cấu trúc/nhân vật/giai đoạn đi qua architect, các chương đã viết làm lại trước tiên đưa vào hàng đợi của editor,
// sau đó Host cử writer viết lại.
//
// Chuẩn hóa thất bại không báo lỗi (hạ cấp thành raw preferences), chỉ khi ghi đĩa thất bại mới trả về tool error——
// Chi tiết kỹ thuật không nên ném lại cho Coordinator như là lỗi quy trình.
type SaveUserRulesTool struct {
	svc *userrules.Service
}

func NewSaveUserRulesTool(svc *userrules.Service) *SaveUserRulesTool {
	return &SaveUserRulesTool{svc: svc}
}

func (t *SaveUserRulesTool) Name() string  { return "save_user_rules" }
func (t *SaveUserRulesTool) Label() string { return "Lưu quy tắc viết" }

func (t *SaveUserRulesTool) Description() string {
	return "Chuẩn hóa các yêu cầu phong cách/chất lượng viết dài hạn của người dùng thành các quy tắc có cấu trúc cho cuốn sách này và lưu trữ lâu dài" +
		"(ví dụ: \"mỗi chương khoảng 1500 chữ\", \"ít dùng ẩn dụ và điệp ngữ\", \"cấm dùng 'ở một mức độ nào đó'\")." +
		"Sau khi lưu, tất cả các sub-agent trong mỗi chương sẽ thấy trong working_memory.user_rules, writer dựa vào đó để viết, commit_chapter tự kiểm tra dựa vào đó, có hiệu lực qua các lần khởi động lại." +
		"text là bắt buộc, chỉ cần thuật lại nguyên văn yêu cầu của người dùng, việc trích xuất cấu trúc sẽ do hệ thống hoàn thành." +
		"Trả về các ràng buộc có cấu trúc hiểu được lần này và các ràng buộc toàn bộ có hiệu lực hiện tại——vui lòng hiển thị lại cho người dùng để xác nhận xem đã hiểu đúng chưa." +
		"Chỉ lưu các quy tắc phong cách/chất lượng viết \"luôn đúng\"; điểu chỉnh cốt truyện/cấu trúc/hướng nhân vật, độ dài theo giai đoạn (ví dụ: \"thêm 10 chương\", \"quyển này viết nhiều về chiến đấu\") đi qua architect, làm lại chương đã viết đi qua editor, tất cả không lưu ở đây."
}

// Công cụ ghi, cấm chạy đồng thời.
func (t *SaveUserRulesTool) ReadOnly(_ json.RawMessage) bool        { return false }
func (t *SaveUserRulesTool) ConcurrencySafe(_ json.RawMessage) bool { return false }

func (t *SaveUserRulesTool) ActivityDescription(_ json.RawMessage) string {
	return "Lưu quy tắc viết"
}

func (t *SaveUserRulesTool) Schema() map[string]any {
	return schema.Object(
		schema.Property("text", schema.String("Yêu cầu viết dài hạn của người dùng (thuật lại nguyên văn, có thể cô đọng), hệ thống sẽ chuẩn hóa thành các ràng buộc có cấu trúc")).Required(),
	)
}

func (t *SaveUserRulesTool) Execute(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	var a struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, fmt.Errorf("invalid args: %w: %w", errs.ErrToolArgs, err)
	}
	text := strings.TrimSpace(a.Text)
	if text == "" {
		return nil, fmt.Errorf("text không được bỏ trống: %w", errs.ErrToolArgs)
	}

	// Chuẩn hóa thất bại sẽ chỉ làm giảm mục nhập đó thành raw preferences (không báo lỗi); chỉ khi lưu thất bại mới trả về tool error.
	snap, cand, err := t.svc.AddRuntimeRule(ctx, text)
	if err != nil {
		return nil, fmt.Errorf("Lưu quy tắc viết thất bại: %w", err)
	}

	return json.Marshal(map[string]any{
		"saved":      true,
		"status":     snap.Status,
		"understood": userRuleUnderstanding(cand), // Hiểu lần này, cung cấp để phản hồi xác nhận
		"in_effect":  snap.Payload(),              // Các ràng buộc toàn bộ có hiệu lực hiện tại
	})
}

// userRuleUnderstanding biến ứng viên chuẩn hóa lần này thành một chế độ xem thực tế cho LLM——
// Coordinator dựa vào đó để trả lại cho người dùng "tôi đã hiểu câu này của bạn là gì", giúp sửa lỗi kịp thời.
func userRuleUnderstanding(c rules.Candidate) map[string]any {
	m := map[string]any{"degraded": c.Degraded}
	if !c.Structured.IsEmpty() {
		m["structured"] = c.Structured
	}
	if p := strings.TrimSpace(c.Preferences); p != "" {
		m["preferences"] = p
	}
	if len(c.Uncertain) > 0 {
		m["uncertain"] = c.Uncertain // Các mục cố ý không nâng cấp thành kiểm tra cứng + lý do
	}
	return m
}
