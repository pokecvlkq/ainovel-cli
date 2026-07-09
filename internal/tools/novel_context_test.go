package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/voocel/ainovel-cli/internal/domain"
	"github.com/voocel/ainovel-cli/internal/store"
)

func TestContextToolInjectsStyleStats(t *testing.T) {
	dir := t.TempDir()
	st := store.NewStore(dir)
	if err := st.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}

	progress := &domain.Progress{TotalChapters: 10}
	body := "# Chương N\nHắn không phải do dự, mà là sợ hãi. Trầm mặc vài nhịp thở. Như một đạo ánh sáng.\nBóng đêm buông xuống.\nHắn đi rồi."
	for ch := 1; ch <= 6; ch++ {
		if err := st.Drafts.SaveFinalChapter(ch, body); err != nil {
			t.Fatalf("SaveFinalChapter: %v", err)
		}
		progress.CompletedChapters = append(progress.CompletedChapters, ch)
	}
	if err := st.Progress.Save(progress); err != nil {
		t.Fatalf("Save progress: %v", err)
	}

	tool := NewContextTool(st, References{}, "default")
	args, _ := json.Marshal(map[string]any{"chapter": 7})
	raw, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	var payload struct {
		Episodic map[string]json.RawMessage `json:"episodic_memory"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	statsRaw, ok := payload.Episodic["style_stats"]
	if !ok {
		t.Fatalf("expected episodic_memory.style_stats, got keys %v", keysOf(payload.Episodic))
	}
	var stats struct {
		Chapters int `json:"chapters"`
		Patterns []struct {
			Name  string `json:"name"`
			Total int    `json:"total"`
		} `json:"patterns"`
	}
	if err := json.Unmarshal(statsRaw, &stats); err != nil {
		t.Fatalf("Unmarshal stats: %v", err)
	}
	if stats.Chapters != 6 || len(stats.Patterns) == 0 {
		t.Errorf("stats content: %+v", stats)
	}
	if usage, ok := payload.Episodic["_usage"]; !ok || len(usage) == 0 {
		t.Error("expected episodic_memory._usage annotation")
	}
}

func keysOf(m map[string]json.RawMessage) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func TestContextToolReportsWarningsForCorruptedState(t *testing.T) {
	dir := t.TempDir()
	store := store.NewStore(dir)
	if err := store.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dir, "outline.json"), []byte("{invalid"), 0o644); err != nil {
		t.Fatalf("write outline.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "meta", "progress.json"), []byte("{invalid"), 0o644); err != nil {
		t.Fatalf("write progress.json: %v", err)
	}

	tool := NewContextTool(store, References{}, "default")
	args, err := json.Marshal(map[string]any{"chapter": 2})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	result, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	var payload struct {
		Warnings []string `json:"_warnings"`
		Summary  string   `json:"_loading_summary"`
	}
	if err := json.Unmarshal(result, &payload); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if len(payload.Warnings) == 0 {
		t.Fatal("expected context warnings for corrupted files")
	}
	if !containsWarning(payload.Warnings, "outline") {
		t.Fatalf("expected outline warning, got %v", payload.Warnings)
	}
	if !containsWarning(payload.Warnings, "progress") {
		t.Fatalf("expected progress warning, got %v", payload.Warnings)
	}
	if !strings.Contains(payload.Summary, "Cảnh báo:") {
		t.Fatalf("expected loading summary to contain warning count, got %q", payload.Summary)
	}
}

func containsWarning(warnings []string, key string) bool {
	for _, warning := range warnings {
		if strings.Contains(warning, key) {
			return true
		}
	}
	return false
}

func TestContextToolChapterModeIncludesWorkingAndReferenceFields(t *testing.T) {
	dir := t.TempDir()
	s := store.NewStore(dir)
	if err := s.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if err := s.Outline.SavePremise(`## Đề tài và giọng điệu
Thiếu niên trưởng thành, hơi căng thẳng áp bách.

## Định vị đề tài
Thiếu niên thăng cấp lưu

## Xung đột cốt lõi
Nhân vật chính phải sống sót trong cuộc cạnh tranh của tông môn.

## Mục tiêu của nhân vật chính
Tiến vào nội môn.

## Hướng kết cục
Trở thành người đánh cờ thực sự.

## Vùng cấm khi viết
Không tiết lộ chân tướng của sư tôn trước thời hạn.

## Điểm thu hút khác biệt
Kẻ yếu nghịch tập.

## Hook khác biệt
Mỗi giai đoạn đều phải dùng cái giá lớn hơn để đổi lấy sự trưởng thành.

## Cam kết cốt lõi
Liên tục hiện thực hóa nguy cơ và đột phá.

## Động cơ câu chuyện
Thử thách, tranh đoạt tài nguyên và thăng cấp thân phận cùng thúc đẩy.

## Chuyển hướng giữa chừng
Nhân vật chính buộc phải chuyển sang một con đường tu hành khác.
`); err != nil {
		t.Fatalf("SavePremise: %v", err)
	}
	if err := s.Outline.SaveOutline([]domain.OutlineEntry{
		{Chapter: 1, Title: "Nhập môn", CoreEvent: "Nhân vật chính tiến vào tông môn", Scenes: []string{"Bái sư", "Lập thệ"}},
		{Chapter: 2, Title: "Thử thách", CoreEvent: "Tham gia thử thách ngoại môn", Scenes: []string{"Tập hợp", "Xuất phát"}},
	}); err != nil {
		t.Fatalf("SaveOutline: %v", err)
	}
	if err := s.Characters.Save([]domain.Character{
		{Name: "Lâm Nghiên", Role: "Nhân vật chính", Description: "Thiếu niên tu sĩ", Arc: "Trưởng thành", Traits: []string{"Bình tĩnh"}},
	}); err != nil {
		t.Fatalf("SaveCharacters: %v", err)
	}
	if err := s.World.SaveWorldRules([]domain.WorldRule{
		{Category: "magic", Rule: "Linh khí có thể luyện hóa", Boundary: "Phàm nhân không thể trực tiếp điều khiển"},
	}); err != nil {
		t.Fatalf("SaveWorldRules: %v", err)
	}
	if err := s.Progress.Init("test", 2); err != nil {
		t.Fatalf("InitProgress: %v", err)
	}
	if err := s.Summaries.SaveSummary(domain.ChapterSummary{
		Chapter:    1,
		Summary:    "Nhân vật chính bái nhập tông môn, xác lập mục tiêu.",
		Characters: []string{"Lâm Nghiên"},
		KeyEvents:  []string{"Bái sư"},
	}); err != nil {
		t.Fatalf("SaveSummary: %v", err)
	}
	if err := s.Drafts.SaveFinalChapter(1, "Phần cuối chương 1, để lại huyền cơ thử thách."); err != nil {
		t.Fatalf("SaveFinalChapter: %v", err)
	}
	if err := s.Drafts.SaveChapterPlan(domain.ChapterPlan{
		Chapter: 2,
		Title:   "Thử thách",
		Goal:    "Vượt qua ải thứ nhất",
		Contract: domain.ChapterContract{
			RequiredBeats:    []string{"Bắt buộc để nhân vật chính vượt qua ải thứ nhất", "Bắt buộc cài cắm thư mời thử thách nội môn"},
			ForbiddenMoves:   []string{"Không được tiết lộ thân phận thật của sư tôn trước thời hạn"},
			ContinuityChecks: []string{"Vết thương cũ ở tay trái nhân vật chính vẫn chưa lành"},
			EvaluationFocus:  []string{"Trọng điểm kiểm tra nhịp độ thử thách có rề rà không"},
		},
	}); err != nil {
		t.Fatalf("SaveChapterPlan: %v", err)
	}
	if err := s.World.SaveStyleRules(domain.WritingStyleRules{
		Volume: 1,
		Arc:    1,
		Prose:  []string{"Trần thuật giữ sự kiềm chế"},
	}); err != nil {
		t.Fatalf("SaveStyleRules: %v", err)
	}
	if err := s.RunMeta.SetPlanningTier(domain.PlanningTierLong); err != nil {
		t.Fatalf("SetPlanningTier: %v", err)
	}

	tool := NewContextTool(s, References{
		Consistency:      "Kiểm tra tính nhất quán",
		HookTechniques:   "Kỹ năng hook",
		QualityChecklist: "Danh sách chất lượng",
	}, "default")
	args, err := json.Marshal(map[string]any{"chapter": 2})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	result, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(result, &payload); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	for _, key := range []string{
		"premise",
		"premise_sections",
		"premise_structure",
		"outline",
		"world_rules",
		"memory_policy",
		"planning_tier",
		"working_memory",
		"episodic_memory",
		"reference_pack",
		"current_chapter_outline",
		"recent_summaries",
		"chapter_plan",
		"chapter_contract",
		"previous_tail",
		"style_rules",
		"references",
	} {
		if _, ok := payload[key]; !ok {
			t.Fatalf("expected key %q in chapter context", key)
		}
	}
}

func TestContextToolArchitectModeIncludesPlanningAndFoundation(t *testing.T) {
	dir := t.TempDir()
	s := store.NewStore(dir)
	if err := s.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if err := s.Outline.SavePremise(`## Đề tài và giọng điệu
Quần tượng mạo hiểm, hơi hướng sử thi lạnh lùng.

## Định vị đề tài
Quần tượng mạo hiểm dài tập

## Xung đột cốt lõi
Mọi người phải tìm kiếm trật tự mới trong trật tự cũ đang mất kiểm soát.

## Mục tiêu của nhân vật chính
Chạm đến cốt lõi sự thật.

## Hướng kết cục
Tiết lộ chân tướng cổ xưa và kiến tạo lại trật tự.

## Vùng cấm khi viết
Không dựa vào thiết lập từ trên trời rơi xuống để kết thúc.

## Điểm thu hút khác biệt
Thúc đẩy quan hệ quần tượng.

## Hook khác biệt
Mỗi tập đều thay đổi cấu trúc quan hệ nhóm.

## Cam kết cốt lõi
Liên tục mang đến khám phá, hy sinh và lựa chọn.

## Động cơ câu chuyện
Hành trình tiến bước, điều tra chân tướng và quan hệ nhóm cùng thúc đẩy.

## Tuyến chính Quan hệ/Phát triển
Nhóm từ không tin tưởng nhau dẫn đến chia rẽ rồi lại tái hợp.

## Lộ trình thăng cấp
Từ sự kiện địa phương tiến tới khủng hoảng cấp thế giới.

## Chuyển hướng giữa chừng
Sự thật không phải là kẻ thù, mà bản thân trật tự có vấn đề.

## Mệnh đề chung cuộc
Trật tự nên do ai định nghĩa.
`); err != nil {
		t.Fatalf("SavePremise: %v", err)
	}
	if err := s.Outline.SaveOutline([]domain.OutlineEntry{
		{Chapter: 1, Title: "Điểm xuất phát", CoreEvent: "Hành trình bắt đầu"},
	}); err != nil {
		t.Fatalf("SaveOutline: %v", err)
	}
	if err := s.Characters.Save([]domain.Character{
		{Name: "Thẩm Diệu", Role: "Nhân vật chính", Description: "Kiếm khách lang thang", Arc: "Tìm kiếm sự thật", Traits: []string{"Nhạy bén"}},
	}); err != nil {
		t.Fatalf("SaveCharacters: %v", err)
	}
	if err := s.World.SaveWorldRules([]domain.WorldRule{
		{Category: "society", Rule: "Thành bang mọc lên san sát", Boundary: "Hoàng quyền không thể trực tiếp quản lý vùng biên giới"},
	}); err != nil {
		t.Fatalf("SaveWorldRules: %v", err)
	}
	if err := s.Outline.SaveLayeredOutline([]domain.VolumeOutline{
		{
			Index: 1, Title: "Quyển 1", Theme: "Bước lên hành trình",
			Arcs: []domain.ArcOutline{
				{Index: 1, Title: "Khởi hành", Goal: "Xây dựng đội ngũ", Chapters: []domain.OutlineEntry{{Chapter: 1, Title: "Điểm xuất phát"}}},
				{Index: 2, Title: "Sương mù", Goal: "Tiếp cận bí mật", EstimatedChapters: 5},
			},
		},
	}); err != nil {
		t.Fatalf("SaveLayeredOutline: %v", err)
	}
	if err := s.Outline.SaveCompass(domain.StoryCompass{
		EndingDirection: "Tiết lộ chân tướng cổ xưa",
		EstimatedScale:  "Dự kiến 3 quyển",
	}); err != nil {
		t.Fatalf("SaveCompass: %v", err)
	}
	if err := s.World.SaveStyleRules(domain.WritingStyleRules{
		Volume: 1,
		Arc:    1,
		Prose:  []string{"Giữ sự lạnh lùng và kiềm chế"},
	}); err != nil {
		t.Fatalf("SaveStyleRules: %v", err)
	}
	if err := s.RunMeta.SetPlanningTier(domain.PlanningTierLong); err != nil {
		t.Fatalf("SetPlanningTier: %v", err)
	}

	tool := NewContextTool(s, References{
		OutlineTemplate:   "Template dàn ý",
		CharacterTemplate: "Template nhân vật",
		LongformPlanning:  "Kế hoạch dài tập",
	}, "default")
	args, err := json.Marshal(map[string]any{})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	result, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(result, &payload); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	for _, key := range []string{
		"memory_policy",
		"planning_tier",
		"planning_memory",
		"foundation_memory",
		"reference_pack",
		"premise_sections",
		"premise_structure",
		"characters",
		"layered_outline",
		"skeleton_arcs",
		"compass",
		"style_rules",
		"references",
		"foundation_status",
	} {
		if _, ok := payload[key]; !ok {
			t.Fatalf("expected key %q in architect context", key)
		}
	}
}

func TestTrimByBudgetRemovesMirroredMemoryKeys(t *testing.T) {
	result := map[string]any{
		"references": map[string]string{
			"a": strings.Repeat("x", 200),
			"b": strings.Repeat("y", 200),
		},
		"reference_pack": map[string]any{
			"references": map[string]string{
				"a": strings.Repeat("x", 200),
				"b": strings.Repeat("y", 200),
			},
			"style_rules": []string{"Kiềm chế"},
		},
	}

	trimByBudget(result, 80)

	if _, ok := result["references"]; ok {
		t.Fatal("expected top-level references to be trimmed")
	}
	pack, ok := result["reference_pack"].(map[string]any)
	if !ok {
		t.Fatal("expected reference_pack to remain available")
	}
	if _, ok := pack["references"]; ok {
		t.Fatal("expected mirrored references to be trimmed from reference_pack")
	}
}

func TestContextToolSelectedMemoryRecallsStoryThreadsAndReviewLessons(t *testing.T) {
	dir := t.TempDir()
	s := store.NewStore(dir)
	if err := s.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if err := s.Outline.SaveOutline([]domain.OutlineEntry{
		{Chapter: 1, Title: "Lời mời", CoreEvent: "Trưởng lão âm thầm đưa ra thư mời thử thách nội môn", Scenes: []string{"Mật đàm", "Để lại lệnh bài thử thách"}},
		{Chapter: 2, Title: "Đêm trước thư mời", CoreEvent: "Lâm Nghiên chuẩn bị hồi đáp thư mời của trưởng lão", Hook: "Ai đứng sau thúc đẩy việc này", Scenes: []string{"Sắp xếp manh mối", "Quyết định phó ước"}},
	}); err != nil {
		t.Fatalf("SaveOutline: %v", err)
	}
	if err := s.Progress.Init("test", 8); err != nil {
		t.Fatalf("InitProgress: %v", err)
	}
	if err := s.World.SaveForeshadowLedger([]domain.ForeshadowEntry{
		{ID: "trial_invite", Description: "Mục đích thực sự của thư mời thử thách nội môn", PlantedAt: 1, Status: "planted"},
		{ID: "trial_mastermind", Description: "Ai đứng sau thúc đẩy thử thách này", PlantedAt: 1, Status: "planted"},
		{ID: "trial_rules", Description: "Tàn quyển bia đá quy tắc thử thách", PlantedAt: 1, Status: "planted"},
		{ID: "outer_disciple", Description: "Tranh chấp nợ cũ của đệ tử ngoại môn", PlantedAt: 1, Status: "planted"},
		{ID: "elder_token", Description: "Lai lịch của lệnh bài trong tay trưởng lão", PlantedAt: 1, Status: "planted"},
		{ID: "hidden_gate", Description: "Lối đi bí mật phía sau sơn môn", PlantedAt: 1, Status: "planted"},
		{ID: "trial_bet", Description: "Người đứng sau thao túng kèo cược thử thách", PlantedAt: 1, Status: "planted"},
	}); err != nil {
		t.Fatalf("SaveForeshadowLedger: %v", err)
	}
	if err := s.Drafts.SaveChapterPlan(domain.ChapterPlan{
		Chapter: 2,
		Title:   "Đêm trước thư mời",
		Goal:    "Quyết định xem có nên nhận thư mời không",
		Contract: domain.ChapterContract{
			PayoffPoints: []string{"Hồi đáp thư mời của trưởng lão"},
			HookGoal:     "Mở ra vấn đề ai đứng sau thúc đẩy việc này",
		},
	}); err != nil {
		t.Fatalf("SaveChapterPlan: %v", err)
	}
	if err := s.World.SaveReview(domain.ReviewEntry{
		Chapter:        1,
		Scope:          "chapter",
		Verdict:        "polish",
		Summary:        "Tuyến chính khởi động xong, nhưng phục bút chưa đủ rõ ràng.",
		ContractStatus: "partial",
		ContractMisses: []string{"Chưa cài cắm rõ ràng thư mời thử thách nội môn"},
		Issues: []domain.ConsistencyIssue{
			{Type: "hook", Severity: "warning", Description: "Hook cuối chương chưa đủ cụ thể"},
		},
	}); err != nil {
		t.Fatalf("SaveReview: %v", err)
	}

	tool := NewContextTool(s, References{}, "default")
	args, err := json.Marshal(map[string]any{"chapter": 2})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	result, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	var payload struct {
		Selected struct {
			StoryThreads  []domain.RecallItem `json:"story_threads"`
			ReviewLessons []domain.RecallItem `json:"review_lessons"`
		} `json:"selected_memory"`
		Summary string `json:"_loading_summary"`
	}
	if err := json.Unmarshal(result, &payload); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if len(payload.Selected.StoryThreads) == 0 {
		t.Fatal("expected story thread recall items")
	}
	if len(payload.Selected.ReviewLessons) == 0 {
		t.Fatal("expected review lesson recall items")
	}
	if !containsRecallSummary(payload.Selected.StoryThreads, "thư mời thử thách nội môn") {
		t.Fatalf("expected story thread recall to mention invite, got %+v", payload.Selected.StoryThreads)
	}
	if !containsRecallSummary(payload.Selected.StoryThreads, "thúc đẩy thử thách này") {
		t.Fatalf("expected story thread recall to mention trial mastermind, got %+v", payload.Selected.StoryThreads)
	}
	if containsRecallSummary(payload.Selected.StoryThreads, "Tàn quyển bia đá quy tắc thử thách") {
		t.Fatalf("expected weak-overlap foreshadow to stay out, got %+v", payload.Selected.StoryThreads)
	}
	if containsRecallSummary(payload.Selected.StoryThreads, "Đề nghị xem lại chương") {
		t.Fatalf("expected related_chapters not to be duplicated into story_threads, got %+v", payload.Selected.StoryThreads)
	}
	if !containsRecallSummary(payload.Selected.ReviewLessons, "thiếu sót hợp đồng") {
		t.Fatalf("expected review lesson recall to mention contract miss, got %+v", payload.Selected.ReviewLessons)
	}
	if !strings.Contains(payload.Summary, "Thu hồi manh mối:") || !strings.Contains(payload.Summary, "Thu hồi đánh giá:") {
		t.Fatalf("expected loading summary to report selected memory, got %q", payload.Summary)
	}
}

// Phục bút đã lâu chưa thu hồi, cho dù không liên quan đến từ khóa chương hiện tại, cũng phải được điền vào story_threads dựa trên thời gian tồn tại——
// Đây chính là điểm mù của việc gọi lại theo độ liên quan (tuyến treo đơn độc quá lâu, nhưng lại không trùng với từ khóa trong chương này).
// Phục bút mới gieo gần đây (thời gian tồn tại < ngưỡng) không nên bị đánh dấu sai là "chưa thu hồi".
func TestContextToolSelectedMemorySurfacesAgingForeshadow(t *testing.T) {
	dir := t.TempDir()
	s := store.NewStore(dir)
	if err := s.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	// Chủ đề chương hiện tại không liên quan đến bất kỳ phục bút nào, đảm bảo gọi lại theo độ liên quan là rỗng, chỉ còn điền lại theo thời gian tồn tại có hiệu lực.
	if err := s.Outline.SaveOutline([]domain.OutlineEntry{
		{Chapter: 50, Title: "Ôn dịch", CoreEvent: "Lâm Nghiên cứu chữa bệnh nhân ôn dịch ở y quán phía nam thành", Scenes: []string{"Sắc thuốc", "Phong tỏa đường phố"}},
	}); err != nil {
		t.Fatalf("SaveOutline: %v", err)
	}
	if err := s.Progress.Init("test", 60); err != nil {
		t.Fatalf("InitProgress: %v", err)
	}
	// 6 mục thỏa mãn ngưỡng gọi lại; hai mục đầu thời gian tồn tại ≥30 (treo lâu), bốn mục sau thời gian tồn tại <30 (gần đây).
	if err := s.World.SaveForeshadowLedger([]domain.ForeshadowEntry{
		{ID: "ancient_seal", Description: "Khe nứt của phong ấn thượng cổ", PlantedAt: 3, Status: "planted"},
		{ID: "lost_bloodline", Description: "Lai lịch huyết mạch thất lạc của nhân vật chính", PlantedAt: 5, Status: "advanced"},
		{ID: "market_feud", Description: "Cuộc cãi vã ở khu chợ đêm qua", PlantedAt: 47, Status: "planted"},
		{ID: "rumor_a", Description: "Tin đồn gần đây A", PlantedAt: 48, Status: "planted"},
		{ID: "rumor_b", Description: "Tin đồn gần đây B", PlantedAt: 48, Status: "planted"},
		{ID: "rumor_c", Description: "Tin đồn gần đây C", PlantedAt: 49, Status: "planted"},
	}); err != nil {
		t.Fatalf("SaveForeshadowLedger: %v", err)
	}

	tool := NewContextTool(s, References{}, "default")
	args, err := json.Marshal(map[string]any{"chapter": 50})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	result, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	var payload struct {
		Selected struct {
			StoryThreads []domain.RecallItem `json:"story_threads"`
		} `json:"selected_memory"`
	}
	if err := json.Unmarshal(result, &payload); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	// Hai phục bút treo lâu phải được điền lại, và mang nhãn dán "chưa thu hồi" theo thời gian tồn tại.
	if !containsRecallSummary(payload.Selected.StoryThreads, "Khe nứt của phong ấn thượng cổ") {
		t.Fatalf("expected aging foreshadow to surface despite no relevance, got %+v", payload.Selected.StoryThreads)
	}
	if !containsRecallSummary(payload.Selected.StoryThreads, "huyết mạch thất lạc") {
		t.Fatalf("expected second aging foreshadow to surface, got %+v", payload.Selected.StoryThreads)
	}
	if !containsRecallSummary(payload.Selected.StoryThreads, "chưa thu hồi") {
		t.Fatalf("expected aging item to carry overdue annotation, got %+v", payload.Selected.StoryThreads)
	}
	// Phục bút gần đây (thời gian tồn tại <30 và không liên quan) không nên được điền lại.
	if containsRecallSummary(payload.Selected.StoryThreads, "Cuộc cãi vã ở khu chợ đêm qua") {
		t.Fatalf("recent foreshadow must not be labeled overdue, got %+v", payload.Selected.StoryThreads)
	}
}

func TestContextToolSelectedMemoryIncludesGlobalReviewLessons(t *testing.T) {
	dir := t.TempDir()
	s := store.NewStore(dir)
	if err := s.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if err := s.Outline.SaveOutline([]domain.OutlineEntry{
		{Chapter: 1, Title: "Mở đầu", CoreEvent: "Câu chuyện bắt đầu"},
		{Chapter: 2, Title: "Thúc đẩy", CoreEvent: "Tuyến chính tiếp tục thúc đẩy"},
	}); err != nil {
		t.Fatalf("SaveOutline: %v", err)
	}
	if err := s.Progress.Init("test", 6); err != nil {
		t.Fatalf("InitProgress: %v", err)
	}
	if err := s.World.SaveReview(domain.ReviewEntry{
		Chapter: 1,
		Scope:   "global",
		Verdict: "polish",
		Summary: "Thúc đẩy toàn cục đạt yêu cầu, nhưng biểu đạt mục tiêu nhân vật vẫn chưa đủ ổn định.",
		Issues: []domain.ConsistencyIssue{
			{Type: "character", Severity: "warning", Description: "Biểu đạt mục tiêu nhân vật chính chưa đủ ổn định"},
		},
	}); err != nil {
		t.Fatalf("SaveReview(global): %v", err)
	}

	tool := NewContextTool(s, References{}, "default")
	args, err := json.Marshal(map[string]any{"chapter": 2})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	result, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	var payload struct {
		Selected struct {
			ReviewLessons []domain.RecallItem `json:"review_lessons"`
		} `json:"selected_memory"`
	}
	if err := json.Unmarshal(result, &payload); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !containsRecallSummary(payload.Selected.ReviewLessons, "Biểu đạt mục tiêu nhân vật chính chưa đủ ổn định") {
		t.Fatalf("expected global review lesson to be recalled, got %+v", payload.Selected.ReviewLessons)
	}
}

func TestContextToolKeepsFullForeshadowWhenRecallNotTriggered(t *testing.T) {
	dir := t.TempDir()
	s := store.NewStore(dir)
	if err := s.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if err := s.Outline.SaveOutline([]domain.OutlineEntry{
		{Chapter: 1, Title: "Bắt đầu", CoreEvent: "Câu chuyện bắt đầu"},
		{Chapter: 2, Title: "Thúc đẩy", CoreEvent: "Tiếp tục thúc đẩy"},
	}); err != nil {
		t.Fatalf("SaveOutline: %v", err)
	}
	if err := s.Progress.Init("test", 4); err != nil {
		t.Fatalf("InitProgress: %v", err)
	}
	if err := s.World.SaveForeshadowLedger([]domain.ForeshadowEntry{
		{ID: "small_1", Description: "Phục bút nhỏ thứ nhất", PlantedAt: 1, Status: "planted"},
		{ID: "small_2", Description: "Phục bút nhỏ thứ hai", PlantedAt: 1, Status: "planted"},
	}); err != nil {
		t.Fatalf("SaveForeshadowLedger: %v", err)
	}

	tool := NewContextTool(s, References{}, "default")
	args, err := json.Marshal(map[string]any{"chapter": 2})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	result, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(result, &payload); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if _, ok := payload["foreshadow_ledger"]; !ok {
		t.Fatal("expected full foreshadow ledger to remain when selected recall is not triggered")
	}
	if _, ok := payload["selected_memory"]; ok {
		t.Fatalf("expected no selected_memory for small foreshadow sets, got %+v", payload["selected_memory"])
	}
}

func TestContextToolFallsBackToFullForeshadowWhenSelectionIsTooSparse(t *testing.T) {
	dir := t.TempDir()
	s := store.NewStore(dir)
	if err := s.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if err := s.Outline.SaveOutline([]domain.OutlineEntry{
		{Chapter: 1, Title: "Lời mời", CoreEvent: "Trưởng lão âm thầm đưa ra thư mời thử thách nội môn"},
		{Chapter: 2, Title: "Đêm trước lúc đi xa", CoreEvent: "Lâm Nghiên suy nghĩ về tương lai và cuộc sống bình dị", Scenes: []string{"Sắp xếp hành lý", "Quyết định phó ước"}},
	}); err != nil {
		t.Fatalf("SaveOutline: %v", err)
	}
	if err := s.Progress.Init("test", 8); err != nil {
		t.Fatalf("InitProgress: %v", err)
	}
	if err := s.World.SaveForeshadowLedger([]domain.ForeshadowEntry{
		{ID: "trial_invite", Description: "Mục đích thực sự của thư mời thử thách nội môn", PlantedAt: 1, Status: "planted"},
		{ID: "trial_rules", Description: "Tàn quyển bia đá quy tắc thử thách", PlantedAt: 1, Status: "planted"},
		{ID: "outer_disciple", Description: "Tranh chấp nợ cũ của đệ tử ngoại môn", PlantedAt: 1, Status: "planted"},
		{ID: "elder_token", Description: "Lai lịch của lệnh bài trong tay trưởng lão", PlantedAt: 1, Status: "planted"},
		{ID: "hidden_gate", Description: "Lối đi bí mật phía sau sơn môn", PlantedAt: 1, Status: "planted"},
		{ID: "trial_bet", Description: "Người đứng sau thao túng kèo cược thử thách", PlantedAt: 1, Status: "planted"},
	}); err != nil {
		t.Fatalf("SaveForeshadowLedger: %v", err)
	}

	tool := NewContextTool(s, References{}, "default")
	args, err := json.Marshal(map[string]any{"chapter": 2})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	result, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(result, &payload); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if _, ok := payload["foreshadow_ledger"]; !ok {
		t.Fatal("expected full foreshadow ledger when selection is too sparse")
	}
	if selected, ok := payload["selected_memory"].(map[string]any); ok {
		if _, exists := selected["story_threads"]; exists {
			t.Fatalf("expected sparse story_threads to fall back to full ledger, got %+v", selected["story_threads"])
		}
	}
}

func containsRecallSummary(items []domain.RecallItem, want string) bool {
	for _, item := range items {
		if strings.Contains(item.Summary, want) {
			return true
		}
	}
	return false
}

func TestContextToolInjectsRewriteBriefForPendingRewriteChapter(t *testing.T) {
	dir := t.TempDir()
	s := store.NewStore(dir)
	if err := s.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if err := s.Progress.Init("test", 3); err != nil {
		t.Fatalf("InitProgress: %v", err)
	}
	if err := s.Progress.MarkChapterComplete(2, 3000, 3000, "", ""); err != nil {
		t.Fatalf("MarkChapterComplete: %v", err)
	}
	if err := s.Progress.SetPendingRewrites([]int{2}, "Nhịp độ rề rà, cần nén phần nửa đầu"); err != nil {
		t.Fatalf("SetPendingRewrites: %v", err)
	}
	if err := s.World.SaveReview(domain.ReviewEntry{
		Chapter: 2,
		Scope:   "chapter",
		Verdict: "rewrite",
		Summary: "Phần nửa đầu lót đường quá dài, xung đột chậm trễ không xuất hiện.",
		Issues: []domain.ConsistencyIssue{
			{Type: "pacing", Severity: "error", Description: "2000 chữ đầu không có tiến triển"},
		},
		ContractMisses: []string{"Chưa hiện thực hóa mở đầu thử thách"},
	}); err != nil {
		t.Fatalf("SaveReview: %v", err)
	}

	tool := NewContextTool(s, References{}, "default")
	args, err := json.Marshal(map[string]any{"chapter": 2})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	result, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(result, &payload); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	brief, ok := payload["rewrite_brief"].(map[string]any)
	if !ok {
		t.Fatalf("expected rewrite_brief in chapter context, got %T", payload["rewrite_brief"])
	}
	if got := brief["reason"]; got != "Nhịp độ rề rà, cần nén phần nửa đầu" {
		t.Fatalf("expected rewrite reason, got %v", got)
	}
	if got, _ := brief["review_summary"].(string); !strings.Contains(got, "lót đường quá dài") {
		t.Fatalf("expected review summary from chapter review, got %v", brief["review_summary"])
	}
	if issues, _ := brief["issues"].([]any); len(issues) == 0 {
		t.Fatalf("expected review issues in rewrite_brief, got %v", brief["issues"])
	}
	if misses, _ := brief["contract_misses"].([]any); len(misses) == 0 {
		t.Fatalf("expected contract misses in rewrite_brief, got %v", brief["contract_misses"])
	}
}

func TestContextToolOmitsRewriteBriefForNormalChapter(t *testing.T) {
	dir := t.TempDir()
	s := store.NewStore(dir)
	if err := s.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if err := s.Progress.Init("test", 3); err != nil {
		t.Fatalf("InitProgress: %v", err)
	}

	tool := NewContextTool(s, References{}, "default")
	args, err := json.Marshal(map[string]any{"chapter": 2})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	result, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(result, &payload); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if _, ok := payload["rewrite_brief"]; ok {
		t.Fatal("expected no rewrite_brief for chapter outside PendingRewrites")
	}
}

func TestContextToolDoesNotInjectUserDirectives(t *testing.T) {
	// save_directive đã bị xóa: novel_context không còn tiêm working_memory.user_directives nữa,
	// yêu cầu viết dài hạn được thống nhất vào user_rules. Khóa điều này để tránh hồi quy.
	dir := t.TempDir()
	s := store.NewStore(dir)
	if err := s.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if err := s.Progress.Init("test", 3); err != nil {
		t.Fatalf("InitProgress: %v", err)
	}

	tool := NewContextTool(s, References{}, "default")
	for name, chapter := range map[string]int{"writer": 1, "architect": 0} {
		args, _ := json.Marshal(map[string]any{"chapter": chapter})
		result, err := tool.Execute(context.Background(), args)
		if err != nil {
			t.Fatalf("[%s] Execute: %v", name, err)
		}
		var payload map[string]any
		if err := json.Unmarshal(result, &payload); err != nil {
			t.Fatalf("[%s] Unmarshal: %v", name, err)
		}
		working, ok := payload["working_memory"].(map[string]any)
		if !ok {
			t.Fatalf("[%s] missing working_memory", name)
		}
		if _, exists := working["user_directives"]; exists {
			t.Errorf("[%s] working_memory không nên có user_directives nữa (đã được thống nhất vào user_rules)", name)
		}
		// user_rules vẫn nên được tiêm ổn định
		if _, ok := working["user_rules"].(map[string]any); !ok {
			t.Errorf("[%s] working_memory.user_rules nên được tiêm ổn định", name)
		}
	}
}
