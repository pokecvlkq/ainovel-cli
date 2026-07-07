package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/voocel/agentcore/schema"
	"github.com/voocel/ainovel-cli/internal/domain"
	"github.com/voocel/ainovel-cli/internal/store"
)

// SaveReviewTool lưu kết quả đánh giá của Editor.
type SaveReviewTool struct {
	store *store.Store
}

func NewSaveReviewTool(store *store.Store) *SaveReviewTool {
	return &SaveReviewTool{store: store}
}

func (t *SaveReviewTool) Name() string { return "save_review" }
func (t *SaveReviewTool) Description() string {
	return "Lưu kết quả đánh giá và cập nhật trạng thái quy trình. verdict là một trong số accept/polish/rewrite." +
		"Công cụ tự động kiểm tra bảng điểm (có thể nâng cấp verdict), trực tiếp cập nhật flow và pending_rewrites của Progress." +
		"Trả về sự thật có cấu trúc: final_verdict / affected_chapters / escalation_reason / next_flow / next_chapter"
}
func (t *SaveReviewTool) Label() string { return "Lưu đánh giá" }

// Công cụ ghi (đồng thời cập nhật reviews/ và PendingRewrites/Flow của Progress), cấm chạy đồng thời.
func (t *SaveReviewTool) ReadOnly(_ json.RawMessage) bool        { return false }
func (t *SaveReviewTool) ConcurrencySafe(_ json.RawMessage) bool { return false }

func (t *SaveReviewTool) Schema() map[string]any {
	issueSchema := schema.Object(
		schema.Property("type", schema.Enum("Loại vấn đề", "consistency", "character", "pacing", "continuity", "foreshadow", "hook", "aesthetic")).Required(),
		schema.Property("severity", schema.Enum("Mức độ nghiêm trọng", "critical", "error", "warning")).Required(),
		schema.Property("description", schema.String("Mô tả vấn đề")).Required(),
		schema.Property("evidence", schema.String("Bằng chứng: Đoạn trích, tình tiết cụ thể hoặc dữ liệu trạng thái")).Required(),
		schema.Property("suggestion", schema.String("Gợi ý chỉnh sửa")),
	)
	dimensionSchema := schema.Object(
		schema.Property("dimension", schema.Enum("Khía cạnh", "consistency", "character", "pacing", "continuity", "foreshadow", "hook", "aesthetic")).Required(),
		schema.Property("score", schema.Int("Điểm số (0-100)")).Required(),
		schema.Property("verdict", schema.Enum("Kết luận theo khía cạnh (có thể bỏ qua: hệ thống tự động suy ra từ score, ≥80 pass / ≥60 warning / <60 fail)", "pass", "warning", "fail")),
		schema.Property("comment", schema.String("Kết luận ngắn gọn cho khía cạnh này; bắt buộc điền cho mỗi khía cạnh, aesthetic phải trích dẫn bản gốc hoặc thống kê cụ thể")).Required(),
	)
	return schema.Object(
		schema.Property("chapter", schema.Int("Số chương đang đánh giá (đánh giá toàn cục điền số chương mới nhất)")).Required(),
		schema.Property("scope", schema.Enum("Phạm vi đánh giá", "chapter", "global", "arc")).Required(),
		schema.Property("dimensions", schema.Array("Chấm điểm theo khía cạnh (7 khía cạnh mỗi khía cạnh 1 mục)", dimensionSchema)).Required(),
		schema.Property("issues", schema.Array("Các vấn đề phát hiện được", issueSchema)).Required(),
		schema.Property("contract_status", schema.Enum("Mức độ hoàn thành khế ước chương", "met", "partial", "missed")),
		schema.Property("contract_misses", schema.Array("Các điều khoản contract chưa hoàn thành hoặc làm trái", schema.String(""))),
		schema.Property("contract_notes", schema.String("Giải thích ngắn gọn về tình hình thực hiện contract")),
		schema.Property("verdict", schema.Enum("Kết luận đánh giá", "accept", "polish", "rewrite")).Required(),
		schema.Property("summary", schema.String("Tóm tắt đánh giá")).Required(),
		schema.Property("affected_chapters", schema.Array("Danh sách các chương cần viết lại hoặc đánh bóng (bắt buộc khi verdict là polish/rewrite)", schema.Int(""))),
	)
}

func (t *SaveReviewTool) Execute(_ context.Context, args json.RawMessage) (json.RawMessage, error) {
	var r domain.ReviewEntry
	if err := json.Unmarshal(args, &r); err != nil {
		return nil, fmt.Errorf("invalid args: %w", err)
	}
	if r.Chapter <= 0 {
		return nil, fmt.Errorf("chapter must be > 0")
	}
	// verdict là hàm thuần túy của score (≥80 pass / ≥60 warning / <60 fail), được tính toán chính xác bằng mã——
	// Không để LLM tự cung cấp rồi lại phải kiểm tra tính nhất quán. Vừa loại bỏ sự thừa thãi, vừa diệt tận gốc mâu thuẫn kiểu "score=85 lại trả về warning".
	for i := range r.Dimensions {
		r.Dimensions[i].Verdict = expectedDimensionVerdict(r.Dimensions[i].Score)
	}
	if err := validateReviewEntry(r); err != nil {
		return nil, err
	}

	// Đánh giá dựa trên bảng điểm — logic nâng cấp được nội suy từ policy/review.go gốc
	finalVerdict := r.Verdict
	var escalationReason string

	if r.Verdict == "accept" {
		// Kiểm tra trạng thái contract
		if r.ContractStatus == "missed" {
			finalVerdict = "rewrite"
			escalationReason = "Trạng thái hợp đồng là missed, nâng cấp lên viết lại (rewrite)"
		} else if r.ContractStatus == "partial" {
			finalVerdict = "polish"
			escalationReason = "Trạng thái hợp đồng là partial, nâng cấp lên đánh bóng (polish)"
		}
		// Kiểm tra bảng điểm
		if finalVerdict == "accept" {
			if gate := evaluateScorecardGate(r.Dimensions); gate != "" {
				if strings.Contains(gate, "rewrite") {
					finalVerdict = "rewrite"
				} else {
					finalVerdict = "polish"
				}
				escalationReason = gate
			}
		}
	}

	affected := r.AffectedChapters
	if finalVerdict == "rewrite" || finalVerdict == "polish" {
		if len(affected) == 0 && r.Chapter > 0 {
			affected = []int{r.Chapter}
		}
		if err := t.store.Progress.ValidatePendingRewrites(affected); err != nil {
			return nil, fmt.Errorf("validate pending rewrites: %w", err)
		}
	}

	if err := t.store.World.SaveReview(r); err != nil {
		return nil, fmt.Errorf("save review: %w", err)
	}

	// Cập nhật Progress dựa trên verdict cuối cùng.
	// Nếu ghi lỗi phải return ngay — phần sau sẽ append checkpoint review, nếu nuốt lỗi ở đây
	// Coordinator sẽ thấy saved:true nhưng Store vẫn ở Flow cũ / thiếu PendingRewrites nửa vời.
	progress, _ := t.store.Progress.Load()
	if finalVerdict == "rewrite" || finalVerdict == "polish" {
		flow := domain.FlowRewriting
		if finalVerdict == "polish" {
			flow = domain.FlowPolishing
		}
		if err := t.store.Progress.SetPendingRewrites(affected, r.Summary); err != nil {
			return nil, fmt.Errorf("set pending rewrites: %w", err)
		}
		if err := t.store.Progress.SetFlow(flow); err != nil {
			return nil, fmt.Errorf("set flow %s: %w", flow, err)
		}
	} else {
		if err := t.store.Progress.SetFlow(domain.FlowWriting); err != nil {
			return nil, fmt.Errorf("set flow writing: %w", err)
		}
	}

	// Lấy snapshot Progress sau khi cập nhật làm sự thật
	latest, _ := t.store.Progress.Load()
	nextFlow := string(domain.FlowWriting)
	nextChapter := 0
	if latest != nil {
		nextFlow = string(latest.Flow)
		nextChapter = latest.NextChapter()
	}

	// Thêm checkpoint
	scope := domain.ChapterScope(r.Chapter)
	if r.Scope == "arc" {
		vol, arc := 0, 0
		if progress != nil {
			vol, arc = progress.CurrentVolume, progress.CurrentArc
		}
		scope = domain.ArcScope(vol, arc)
	}
	artifact := fmt.Sprintf("reviews/%02d.json", r.Chapter)
	if r.Scope == "global" {
		artifact = fmt.Sprintf("reviews/%02d-global.json", r.Chapter)
	}
	if _, err := t.store.Checkpoints.AppendArtifact(scope, "review", artifact); err != nil {
		return nil, fmt.Errorf("checkpoint review: %w", err)
	}

	return json.Marshal(map[string]any{
		"saved":             true,
		"chapter":           r.Chapter,
		"scope":             r.Scope,
		"verdict":           r.Verdict,
		"final_verdict":     finalVerdict,
		"escalation_reason": escalationReason,
		"affected_chapters": affected,
		"issues":            len(r.Issues),
		"next_flow":         nextFlow,
		"next_chapter":      nextChapter,
	})
}

var expectedReviewDimensions = map[string]struct{}{
	"consistency": {},
	"character":   {},
	"pacing":      {},
	"continuity":  {},
	"foreshadow":  {},
	"hook":        {},
	"aesthetic":   {},
}

func validateReviewEntry(r domain.ReviewEntry) error {
	if strings.TrimSpace(r.Scope) == "" {
		return fmt.Errorf("scope is required")
	}
	if strings.TrimSpace(r.Summary) == "" {
		return fmt.Errorf("summary is required")
	}
	for _, issue := range r.Issues {
		if strings.TrimSpace(issue.Description) == "" {
			return fmt.Errorf("issue description is required")
		}
		if strings.TrimSpace(issue.Evidence) == "" {
			return fmt.Errorf("issue evidence is required")
		}
	}
	if err := validateDimensions(r.Dimensions); err != nil {
		return err
	}
	if (r.Verdict == "rewrite" || r.Verdict == "polish") && len(r.AffectedChapters) == 0 {
		return fmt.Errorf("affected_chapters is required when verdict=%s", r.Verdict)
	}
	return nil
}

func validateDimensions(dimensions []domain.DimensionScore) error {
	if len(dimensions) != len(expectedReviewDimensions) {
		return fmt.Errorf("dimensions must contain exactly %d entries", len(expectedReviewDimensions))
	}

	seen := make(map[string]struct{}, len(dimensions))
	for _, dim := range dimensions {
		if _, ok := expectedReviewDimensions[dim.Dimension]; !ok {
			return fmt.Errorf("unknown dimension: %s", dim.Dimension)
		}
		if _, ok := seen[dim.Dimension]; ok {
			return fmt.Errorf("duplicate dimension: %s", dim.Dimension)
		}
		seen[dim.Dimension] = struct{}{}
		if dim.Score < 0 || dim.Score > 100 {
			return fmt.Errorf("invalid score for %s: %d", dim.Dimension, dim.Score)
		}
		if strings.TrimSpace(dim.Comment) == "" {
			return fmt.Errorf("dimension comment is required: %s", dim.Dimension)
		}
	}
	return nil
}

func expectedDimensionVerdict(score int) string {
	switch {
	case score >= 80:
		return "pass"
	case score >= 60:
		return "warning"
	default:
		return "fail"
	}
}

// criticalDimensions định nghĩa các khía cạnh quan trọng kích hoạt nâng cấp verdict.
var criticalDimensions = map[string]struct{}{
	"consistency": {},
	"character":   {},
	"continuity":  {},
}

// evaluateScorecardGate kiểm tra xem bảng điểm có cần nâng cấp verdict không.
// Trả về chuỗi rỗng nghĩa là không cần nâng cấp.
func evaluateScorecardGate(dimensions []domain.DimensionScore) string {
	var criticalFails []string
	var polishIssues []string

	for _, dim := range dimensions {
		_, isCritical := criticalDimensions[dim.Dimension]
		if isCritical && (dim.Verdict == "fail" || dim.Score < 60) {
			criticalFails = append(criticalFails, fmt.Sprintf("%s(%d)", dim.Dimension, dim.Score))
		} else if dim.Verdict == "warning" || (isCritical && dim.Score < 80) {
			polishIssues = append(polishIssues, fmt.Sprintf("%s(%d)", dim.Dimension, dim.Score))
		}
	}

	if len(criticalFails) > 0 {
		return fmt.Sprintf("rewrite: Khía cạnh quan trọng không đạt %v", criticalFails)
	}
	if len(polishIssues) > 0 {
		return fmt.Sprintf("polish: Một số khía cạnh cần đánh bóng %v", polishIssues)
	}
	return ""
}
