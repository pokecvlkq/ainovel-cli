package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/voocel/ainovel-cli/internal/domain"
	"github.com/voocel/ainovel-cli/internal/store"
)

func TestSaveReviewPersistsContractAssessment(t *testing.T) {
	s := store.NewStore(t.TempDir())
	if err := s.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if err := s.Progress.Init("test", 10); err != nil {
		t.Fatalf("Progress.Init: %v", err)
	}
	if err := s.Progress.MarkChapterComplete(3, 3000, 3000, "", ""); err != nil {
		t.Fatalf("MarkChapterComplete: %v", err)
	}

	tool := NewSaveReviewTool(s)
	args, err := json.Marshal(map[string]any{
		"chapter":           3,
		"scope":             "chapter",
		"dimensions":        []map[string]any{{"dimension": "consistency", "score": 85, "verdict": "pass", "comment": "Cơ bản nhất quán"}, {"dimension": "character", "score": 82, "verdict": "pass", "comment": "Nhân vật ổn định"}, {"dimension": "pacing", "score": 78, "verdict": "warning", "comment": "Hơi chậm"}, {"dimension": "continuity", "score": 84, "verdict": "pass", "comment": "Mạch lạc"}, {"dimension": "foreshadow", "score": 80, "verdict": "pass", "comment": "Bình thường"}, {"dimension": "hook", "score": 76, "verdict": "warning", "comment": "Hook bình thường"}, {"dimension": "aesthetic", "score": 81, "verdict": "pass", "comment": "Văn phong cơ bản hợp lý"}},
		"issues":            []map[string]any{},
		"contract_status":   "partial",
		"contract_misses":   []string{"Chưa thiết lập rõ thư mời tham gia thử thách nội môn"},
		"contract_notes":    "Tuyến chính đã được thúc đẩy, nhưng mục thúc đẩy thứ hai trong hợp đồng chưa được thực hiện.",
		"verdict":           "polish",
		"summary":           "Chương này cơ bản hoàn thành mục tiêu, nhưng hợp đồng vẫn còn thiếu sót.",
		"affected_chapters": []int{3},
	})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	if _, err := tool.Execute(context.Background(), args); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	review, err := s.World.LoadReview(3)
	if err != nil {
		t.Fatalf("LoadReview: %v", err)
	}
	if review == nil {
		t.Fatal("expected review saved, got nil")
	}
	if review.ContractStatus != "partial" {
		t.Fatalf("unexpected contract status: %q", review.ContractStatus)
	}
	if len(review.ContractMisses) != 1 || review.ContractMisses[0] != "Chưa thiết lập rõ thư mời tham gia thử thách nội môn" {
		t.Fatalf("unexpected contract misses: %+v", review.ContractMisses)
	}
	if review.Dimension("aesthetic") == nil {
		t.Fatalf("expected aesthetic dimension persisted, got %+v", review.Dimensions)
	}
}

func TestSaveReviewRejectsMissingDimensions(t *testing.T) {
	s := store.NewStore(t.TempDir())
	if err := s.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if err := s.Progress.Init("test", 10); err != nil {
		t.Fatalf("Progress.Init: %v", err)
	}
	if err := s.Progress.MarkChapterComplete(3, 3000, 3000, "", ""); err != nil {
		t.Fatalf("MarkChapterComplete: %v", err)
	}

	tool := NewSaveReviewTool(s)
	args, err := json.Marshal(map[string]any{
		"chapter":    3,
		"scope":      "chapter",
		"dimensions": []map[string]any{{"dimension": "consistency", "score": 85, "verdict": "pass", "comment": "Cơ bản nhất quán"}},
		"issues":     []map[string]any{},
		"verdict":    "accept",
		"summary":    "ok",
	})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	if _, err := tool.Execute(context.Background(), args); err == nil || !strings.Contains(err.Error(), "dimensions must contain exactly") {
		t.Fatalf("expected dimensions validation error, got %v", err)
	}
}

func TestSaveReviewRejectsDimensionWithoutComment(t *testing.T) {
	s := store.NewStore(t.TempDir())
	if err := s.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if err := s.Progress.Init("test", 10); err != nil {
		t.Fatalf("Progress.Init: %v", err)
	}
	if err := s.Progress.MarkChapterComplete(3, 3000, 3000, "", ""); err != nil {
		t.Fatalf("MarkChapterComplete: %v", err)
	}

	tool := NewSaveReviewTool(s)
	args, err := json.Marshal(map[string]any{
		"chapter": 3,
		"scope":   "chapter",
		"dimensions": []map[string]any{
			{"dimension": "consistency", "score": 85, "comment": "Cơ bản nhất quán"},
			{"dimension": "character", "score": 82, "comment": "Nhân vật ổn định"},
			{"dimension": "pacing", "score": 78},
			{"dimension": "continuity", "score": 84, "comment": "Mạch lạc"},
			{"dimension": "foreshadow", "score": 80, "comment": "Bình thường"},
			{"dimension": "hook", "score": 76, "comment": "Hook bình thường"},
			{"dimension": "aesthetic", "score": 81, "comment": "Văn phong cơ bản hợp lý"},
		},
		"issues":  []map[string]any{},
		"verdict": "accept",
		"summary": "ok",
	})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	if _, err := tool.Execute(context.Background(), args); err == nil || !strings.Contains(err.Error(), "dimension comment is required: pacing") {
		t.Fatalf("expected dimension comment validation error, got %v", err)
	}
}

func TestSaveReviewRejectsUnfinishedAffectedChapter(t *testing.T) {
	s := store.NewStore(t.TempDir())
	if err := s.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if err := s.Progress.Init("test", 80); err != nil {
		t.Fatalf("Progress.Init: %v", err)
	}
	for ch := 1; ch <= 58; ch++ {
		if err := s.Progress.MarkChapterComplete(ch, 3000, 3000, "", ""); err != nil {
			t.Fatalf("MarkChapterComplete(%d): %v", ch, err)
		}
	}

	tool := NewSaveReviewTool(s)
	args, err := json.Marshal(map[string]any{
		"chapter": 58,
		"scope":   "chapter",
		"dimensions": []map[string]any{
			{"dimension": "consistency", "score": 85, "comment": "Cơ bản nhất quán"},
			{"dimension": "character", "score": 82, "comment": "Nhân vật ổn định"},
			{"dimension": "pacing", "score": 58, "comment": "Cần viết lại nhịp độ"},
			{"dimension": "continuity", "score": 84, "comment": "Mạch lạc"},
			{"dimension": "foreshadow", "score": 80, "comment": "Bình thường"},
			{"dimension": "hook", "score": 76, "comment": "Hook bình thường"},
			{"dimension": "aesthetic", "score": 81, "comment": "Văn phong cơ bản hợp lý"},
		},
		"issues":            []map[string]any{},
		"contract_status":   "partial",
		"verdict":           "polish",
		"summary":           "Cần trau chuốt chương 58, không thể đưa chương chưa hoàn thành vào hàng đợi.",
		"affected_chapters": []int{65},
		"contract_misses":   []string{"Nhịp độ vượt quá trách nhiệm của chương này"},
		"contract_notes":    "Chỉ nên xử lý các chương đã hoàn thành.",
	})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	if _, err := tool.Execute(context.Background(), args); err == nil || !strings.Contains(err.Error(), "pending_rewrites chỉ có thể chứa các chương đã hoàn thành") {
		t.Fatalf("expected unfinished affected chapter rejection, got %v", err)
	}
	review, err := s.World.LoadReview(58)
	if err != nil {
		t.Fatalf("LoadReview: %v", err)
	}
	if review != nil {
		t.Fatalf("review should not be saved when pending rewrite validation fails: %+v", review)
	}
	p, _ := s.Progress.Load()
	if p.Flow != domain.FlowWriting && p.Flow != "" {
		t.Fatalf("flow should not enter rewrite/polish, got %s", p.Flow)
	}
	if len(p.PendingRewrites) != 0 {
		t.Fatalf("pending_rewrites should remain empty, got %v", p.PendingRewrites)
	}
}

// TestSaveReviewDerivesVerdictFromScore Kiểm tra: verdict được suy ra một cách chắc chắn từ score, nếu mô hình cung cấp
// verdict không nhất quán (ví dụ score=85 nhưng ghi là warning) sẽ không báo lỗi nữa, mà được ghi đè thành giá trị đúng (pass).
// Issue phòng ngừa lùi lội: Tình trạng điểm/verdict không khớp từ mô hình yếu từng khiến save_review liên tục gặp lỗi.
func TestSaveReviewDerivesVerdictFromScore(t *testing.T) {
	s := store.NewStore(t.TempDir())
	if err := s.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if err := s.Progress.Init("test", 10); err != nil {
		t.Fatalf("Progress.Init: %v", err)
	}
	if err := s.Progress.MarkChapterComplete(3, 3000, 3000, "", ""); err != nil {
		t.Fatalf("MarkChapterComplete: %v", err)
	}

	tool := NewSaveReviewTool(s)
	args, err := json.Marshal(map[string]any{
		"chapter": 3,
		"scope":   "chapter",
		"dimensions": []map[string]any{
			{"dimension": "consistency", "score": 85, "verdict": "pass", "comment": "Nhất quán"},
			{"dimension": "character", "score": 82, "comment": "Ổn định"}, // Bỏ qua verdict
			{"dimension": "pacing", "score": 78, "verdict": "warning", "comment": "Hơi chậm"},
			{"dimension": "continuity", "score": 84, "verdict": "pass", "comment": "Mạch lạc"},
			{"dimension": "foreshadow", "score": 80, "verdict": "pass", "comment": "Bình thường"},
			{"dimension": "hook", "score": 76, "verdict": "warning", "comment": "Hook bình thường"},
			{"dimension": "aesthetic", "score": 85, "verdict": "warning", "comment": "Văn phong hợp lý"}, // Không nhất quán: 85 nhưng ghi warning
		},
		"issues":  []map[string]any{},
		"verdict": "accept",
		"summary": "ok",
	})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	if _, err := tool.Execute(context.Background(), args); err != nil {
		t.Fatalf("Execute should succeed (verdict auto-derived), got %v", err)
	}

	review, err := s.World.LoadReview(3)
	if err != nil || review == nil {
		t.Fatalf("LoadReview: %v", err)
	}
	// 85 → pass (ghi đè warning mô hình đưa ra); 82 bị bỏ qua → pass.
	if d := review.Dimension("aesthetic"); d == nil || d.Verdict != "pass" {
		t.Fatalf("aesthetic verdict should be derived to pass, got %+v", d)
	}
	if d := review.Dimension("character"); d == nil || d.Verdict != "pass" {
		t.Fatalf("character verdict should be derived to pass, got %+v", d)
	}
}

func TestSaveReviewRejectsMissingAffectedChaptersForRewrite(t *testing.T) {
	s := store.NewStore(t.TempDir())
	if err := s.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}

	tool := NewSaveReviewTool(s)
	args, err := json.Marshal(map[string]any{
		"chapter": 3,
		"scope":   "chapter",
		"dimensions": []map[string]any{
			{"dimension": "consistency", "score": 85, "verdict": "pass", "comment": "Cơ bản nhất quán"},
			{"dimension": "character", "score": 82, "verdict": "pass", "comment": "Nhân vật ổn định"},
			{"dimension": "pacing", "score": 78, "verdict": "warning", "comment": "Hơi chậm"},
			{"dimension": "continuity", "score": 84, "verdict": "pass", "comment": "Mạch lạc"},
			{"dimension": "foreshadow", "score": 80, "verdict": "pass", "comment": "Bình thường"},
			{"dimension": "hook", "score": 76, "verdict": "warning", "comment": "Hook bình thường"},
			{"dimension": "aesthetic", "score": 81, "verdict": "pass", "comment": "Văn phong cơ bản hợp lý"},
		},
		"issues":  []map[string]any{},
		"verdict": "rewrite",
		"summary": "Cần viết lại",
	})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	if _, err := tool.Execute(context.Background(), args); err == nil || !strings.Contains(err.Error(), "affected_chapters is required") {
		t.Fatalf("expected affected_chapters validation error, got %v", err)
	}
}

func TestSaveReviewRejectsIssueWithoutEvidence(t *testing.T) {
	s := store.NewStore(t.TempDir())
	if err := s.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}

	tool := NewSaveReviewTool(s)
	args, err := json.Marshal(map[string]any{
		"chapter": 3,
		"scope":   "chapter",
		"dimensions": []map[string]any{
			{"dimension": "consistency", "score": 85, "verdict": "pass", "comment": "Cơ bản nhất quán"},
			{"dimension": "character", "score": 82, "verdict": "pass", "comment": "Nhân vật ổn định"},
			{"dimension": "pacing", "score": 78, "verdict": "warning", "comment": "Hơi chậm"},
			{"dimension": "continuity", "score": 84, "verdict": "pass", "comment": "Mạch lạc"},
			{"dimension": "foreshadow", "score": 80, "verdict": "pass", "comment": "Bình thường"},
			{"dimension": "hook", "score": 76, "verdict": "warning", "comment": "Hook bình thường"},
			{"dimension": "aesthetic", "score": 81, "verdict": "pass", "comment": "Văn phong cơ bản hợp lý"},
		},
		"issues": []map[string]any{
			{"type": "hook", "severity": "warning", "description": "Hook cuối chương hơi yếu"},
		},
		"verdict":           "polish",
		"summary":           "Cần củng cố hook.",
		"affected_chapters": []int{3},
	})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	if _, err := tool.Execute(context.Background(), args); err == nil || !strings.Contains(err.Error(), "issue evidence is required") {
		t.Fatalf("expected issue evidence validation error, got %v", err)
	}
}
