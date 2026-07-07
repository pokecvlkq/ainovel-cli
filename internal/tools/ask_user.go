package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/voocel/agentcore/schema"
)

// AskUserResponse Kết quả trả lời của người dùng.
type AskUserResponse struct {
	Answers map[string]string // question text → Câu trả lời do người dùng chọn
	Notes   map[string]string // question text → Nhập tuỳ chỉnh (khi chọn "khác")
}

// AskUserHandler Chặn đợi người dùng trả lời, được triển khai cụ thể bởi CLI hoặc TUI.
type AskUserHandler func(ctx context.Context, questions []Question) (*AskUserResponse, error)

// Question Một câu hỏi đơn lẻ.
type Question struct {
	Question    string   `json:"question"`
	Header      string   `json:"header"`
	Options     []Option `json:"options"`
	MultiSelect bool     `json:"multiSelect"`
}

// Option Tuỳ chọn.
type Option struct {
	Label       string `json:"label"`
	Description string `json:"description"`
}

// AskUserTool Cho phép LLM đặt câu hỏi có cấu trúc cho người dùng.
type AskUserTool struct {
	mu      sync.RWMutex
	handler AskUserHandler
}

func NewAskUserTool() *AskUserTool {
	return &AskUserTool{}
}

// SetHandler Tiêm callback UI, CLI và TUI tự triển khai.
func (t *AskUserTool) SetHandler(h AskUserHandler) {
	t.mu.Lock()
	t.handler = h
	t.mu.Unlock()
}

func (t *AskUserTool) Name() string  { return "ask_user" }
func (t *AskUserTool) Label() string { return "Hỏi người dùng" }

// Công cụ tương tác: chặn đợi người dùng trả lời, rõ ràng không thể lập lịch đồng thời.
func (t *AskUserTool) ReadOnly(_ json.RawMessage) bool        { return false }
func (t *AskUserTool) ConcurrencySafe(_ json.RawMessage) bool { return false }
func (t *AskUserTool) Description() string {
	return "Khi thông tin yêu cầu không đủ và thông tin bị thiếu sẽ ảnh hưởng đáng kể đến hướng lập kế hoạch, hãy hỏi người dùng 1-4 câu hỏi có cấu trúc. Mỗi câu hỏi phải bao gồm header, question và 2-4 tuỳ chọn; người dùng có thể chọn các mục định sẵn, hoặc tự do bổ sung. Kết quả trả về là một tóm tắt tiếng Việt có thể đọc trực tiếp, định dạng tương tự: Người dùng trả lời: [Thời lượng] Dài; [Trọng tâm] Nâng cấp cốt truyện (bổ sung: không harem). Chỉ sử dụng khi không thể đánh giá ổn định thời lượng, trọng tâm cốt truyện, các ràng buộc chính hoặc sở thích rõ ràng; không ném các vấn đề có thể tự suy luận hợp lý cho người dùng."
}

func (t *AskUserTool) Schema() map[string]any {
	option := schema.Object(
		schema.Property("label", schema.String("Văn bản hiển thị tuỳ chọn (1-5 từ)")).Required(),
		schema.Property("description", schema.String("Giải thích ý nghĩa tuỳ chọn")).Required(),
	)
	question := schema.Object(
		schema.Property("question", schema.String("Văn bản câu hỏi đầy đủ")).Required(),
		schema.Property("header", schema.String("Thẻ ngắn (tối đa 12 ký tự)")).Required(),
		schema.Property("options", schema.Array("2-4 tuỳ chọn", option)).Required(),
		schema.Property("multiSelect", schema.Bool("Có cho phép chọn nhiều hay không")),
	)
	return schema.Object(
		schema.Property("questions", schema.Array("1-4 câu hỏi", question)).Required(),
	)
}

type askUserArgs struct {
	Questions []Question `json:"questions"`
}

func (t *AskUserTool) Execute(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	var a askUserArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, fmt.Errorf("invalid args: %w", err)
	}
	if err := validateQuestions(a.Questions); err != nil {
		return json.Marshal(fmt.Sprintf("Xác thực tham số thất bại: %s", err))
	}

	t.mu.RLock()
	h := t.handler
	t.mu.RUnlock()

	if h == nil {
		return json.Marshal("Môi trường hiện tại không hỗ trợ hỏi đáp tương tác, vui lòng tự đưa ra quyết định dựa trên phán đoán của bạn và tiếp tục.")
	}

	resp, err := h(ctx, a.Questions)
	if err != nil {
		return json.Marshal(fmt.Sprintf("Tương tác người dùng thất bại: %s. Vui lòng tự đưa ra quyết định dựa trên phán đoán của bạn và tiếp tục.", err))
	}

	return json.Marshal(formatAnswers(a.Questions, resp))
}

func validateQuestions(questions []Question) error {
	if len(questions) == 0 {
		return fmt.Errorf("Cần ít nhất một câu hỏi")
	}
	if len(questions) > 4 {
		return fmt.Errorf("Tối đa 4 câu hỏi, hiện tại có %d", len(questions))
	}
	for i, q := range questions {
		if q.Question == "" {
			return fmt.Errorf("Câu hỏi %d: Văn bản câu hỏi không được để trống", i+1)
		}
		if q.Header == "" {
			return fmt.Errorf("Câu hỏi %d: header không được để trống", i+1)
		}
		if utf8.RuneCountInString(q.Header) > 12 {
			return fmt.Errorf("Câu hỏi %d: header %q vượt quá 12 ký tự", i+1, q.Header)
		}
		if len(q.Options) < 2 || len(q.Options) > 4 {
			return fmt.Errorf("Câu hỏi %d: Cần 2-4 tuỳ chọn, hiện tại có %d", i+1, len(q.Options))
		}
		for j, opt := range q.Options {
			if opt.Label == "" {
				return fmt.Errorf("Câu hỏi %d tuỳ chọn %d: label không được để trống", i+1, j+1)
			}
			if opt.Description == "" {
				return fmt.Errorf("Câu hỏi %d tuỳ chọn %d: description không được để trống", i+1, j+1)
			}
		}
	}
	return nil
}

func formatAnswers(questions []Question, resp *AskUserResponse) string {
	if resp == nil || len(resp.Answers) == 0 {
		return "Người dùng chưa cung cấp câu trả lời, vui lòng tự đưa ra quyết định dựa trên phán đoán của bạn và tiếp tục."
	}
	var parts []string
	for _, q := range questions {
		answer, ok := resp.Answers[q.Question]
		if !ok {
			continue
		}
		entry := fmt.Sprintf("[%s] %s", q.Header, answer)
		if note, hasNote := resp.Notes[q.Question]; hasNote {
			entry += " (bổ sung: " + note + ")"
		}
		parts = append(parts, entry)
	}
	return fmt.Sprintf("Người dùng trả lời: %s", strings.Join(parts, "; "))
}
