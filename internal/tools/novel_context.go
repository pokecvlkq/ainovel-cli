package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/voocel/agentcore/schema"
	"github.com/voocel/ainovel-cli/internal/domain"
	"github.com/voocel/ainovel-cli/internal/store"
)

// References Các tài liệu tham khảo được nhúng.
type References struct {
	// V0
	ChapterGuide      string
	HookTechniques    string
	QualityChecklist  string
	OutlineTemplate   string
	CharacterTemplate string
	ChapterTemplate   string
	// V1
	Consistency      string
	ContentExpansion string
	DialogueWriting  string
	// V2
	StyleReference   string // Tài liệu tham khảo bổ sung về phong cách (có thể rỗng)
	LongformPlanning string // Tài liệu tham khảo quy hoạch dài hạn chung
	Differentiation  string // Tài liệu tham khảo thiết kế khác biệt chung
	ArcTemplates     string // Mẫu cốt truyện theo thể loại (tải theo style, có thể rỗng)
	AntiAITone       string // Thư viện tiêu chí loại bỏ giọng điệu AI (dùng chung cho writer/editor, nhúng trong toàn bộ quá trình)
}

// ContextTool Lắp ráp ngữ cảnh cần thiết cho chương hiện tại.
type ContextTool struct {
	store *store.Store
	refs  References
	style string
}

// NewContextTool Tạo công cụ ngữ cảnh.
// user_rules được buildUserRules đọc trực tiếp từ bản chụp của sách (meta/user_rules.json) và đưa vào, không còn phụ thuộc vào các tùy chọn tải.
func NewContextTool(store *store.Store, refs References, style string) *ContextTool {
	return &ContextTool{store: store, refs: refs, style: style}
}

func (t *ContextTool) Name() string { return "novel_context" }
func (t *ContextTool) Description() string {
	return "Lấy trạng thái hiện tại và ngữ cảnh sáng tác của tiểu thuyết." +
		"Nếu không truyền chapter: trả về progress_status (các trường tiến độ như phase/flow/next_chapter/pending_rewrites) + cài đặt cơ bản, dùng để xác định bước tiếp theo nên làm gì." +
		"Nếu truyền chapter=N: trả về thêm ngữ cảnh viết của chương đó như tóm tắt trước đó, phục bút, trạng thái nhân vật, quy tắc phong cách, v.v."
}
func (t *ContextTool) Label() string { return "Tải ngữ cảnh" }

// Công cụ chỉ đọc, có thể được lập lịch đồng thời.
func (t *ContextTool) ReadOnly(_ json.RawMessage) bool        { return true }
func (t *ContextTool) ConcurrencySafe(_ json.RawMessage) bool { return true }

func (t *ContextTool) Schema() map[string]any {
	return schema.Object(
		schema.Property("chapter", schema.Int("Số chương. Nếu không truyền sẽ trả về trạng thái tiến độ và thiết lập cơ bản (Coordinator dùng để xác định bước tiếp theo); nếu truyền sẽ trả về thêm ngữ cảnh viết của chương đó (Writer dùng)")),
	)
}

func (t *ContextTool) Execute(_ context.Context, args json.RawMessage) (json.RawMessage, error) {
	var a struct {
		Chapter int `json:"chapter"`
	}
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, fmt.Errorf("invalid args: %w", err)
	}

	result := make(map[string]any)
	var warnings []string
	seenWarnings := make(map[string]struct{})
	warn := func(scope string, err error) {
		if err == nil || os.IsNotExist(err) {
			return
		}
		msg := fmt.Sprintf("%s Đọc thất bại: %v", scope, err)
		if _, ok := seenWarnings[msg]; ok {
			return
		}
		seenWarnings[msg] = struct{}{}
		warnings = append(warnings, msg)
	}

	if a.Chapter > 0 {
		// Luồng Writer: Tải toàn bộ dữ liệu cơ bản + ngữ cảnh chương
		t.buildBaseContext(result, warn)
		seed := newChapterContextEnvelope()
		state := t.prepareChapterContext(a.Chapter, &seed, warn)
		seed.apply(result)
		t.buildChapterContext(result, state, warn)
		// Ghi chú ngữ nghĩa dữ liệu (trị lỗi lặp lại thông tin): episodic là ghi nhớ đã được viết vào nội dung chính, không phải tài liệu chờ viết.
		// Chỉ treo trong vùng chứa, không vào ảnh cấp cao nhất.
		if epi, ok := result["episodic_memory"].(map[string]any); ok && len(epi) > 0 {
			epi["_usage"] = "Vùng chứa này là bản ghi nhớ các sự kiện đã viết trong nội dung chính (để đối chiếu tính nhất quán và liền mạch); việc lặp lại y nguyên những nội dung này trong phần chính của chương mới là một lỗi lặp lại"
		}
	} else {
		// Luồng Coordinator/Architect: Chỉ trả về trạng thái + dữ liệu cấu trúc, không tải toàn bộ văn bản gốc
		t.buildProgressStatus(result)
		t.buildArchitectContext(result, warn)
	}

	// Đưa vào working_memory.user_rules (đường dẫn chuẩn). Luồng architect ban đầu không có working_memory,
	// do buildUserRules tạo mới khi cần để chỉ chứa vùng chứa user_rules. Khi thiếu bản chụp sẽ lùi về mặc định được tích hợp sẵn,
	// luôn xuất ra cấu trúc ổn định, tránh LLM nhìn thấy user_rules=null sẽ đi vào nhánh bất thường.
	if a.Chapter > 0 {
		t.buildSimulationProfile(result, "working_memory", warn)
	} else {
		t.buildSimulationProfile(result, "planning_memory", warn)
	}

	t.buildUserRules(result)

	if len(warnings) > 0 {
		result["_warnings"] = warnings
	}

	// Ngân sách ưu tiên: tự động cắt giảm dữ liệu độ ưu tiên thấp khi tổng kích thước vượt ngưỡng
	if a.Chapter > 0 {
		trimByBudget(result, 100*1024) // Writer: 100KB
	} else {
		trimByBudget(result, 60*1024) // Coordinator/Architect: 60KB
	}

	result["_loading_summary"] = buildLoadingSummary(result, a.Chapter)
	return json.Marshal(result)
}

// buildLoadingSummary Thống kê lượng dữ liệu các mục từ result đã lắp ráp, tạo một dòng tóm tắt dễ đọc.
func buildLoadingSummary(result map[string]any, chapter int) string {
	var parts []string

	if chapter > 0 {
		parts = append(parts, fmt.Sprintf("ch=%d", chapter))
	} else {
		parts = append(parts, "architect")
	}
	if tier, ok := result["planning_tier"].(domain.PlanningTier); ok && tier != "" {
		parts = append(parts, fmt.Sprintf("tier=%s", tier))
	}

	// Vị trí quyển và phần
	if pos, ok := result["position"].(map[string]any); ok {
		parts = append(parts, fmt.Sprintf("V%dA%d", pos["volume"], pos["arc"]))
	}

	var items []string
	countSlice := func(key string) int {
		if v, ok := result[key]; ok {
			if s, ok := v.([]domain.Character); ok {
				return len(s)
			}
			// Reflection chung cho slice
			return sliceLen(v)
		}
		return 0
	}

	// Nhân vật
	if n := countSlice("character_snapshots"); n > 0 {
		items = append(items, fmt.Sprintf("Nhân vật:%d(Bản chụp)", n))
	} else if n := countSlice("characters"); n > 0 {
		items = append(items, fmt.Sprintf("Nhân vật:%d", n))
	}

	if working, ok := result["working_memory"].(map[string]any); ok && len(working) > 0 {
		items = append(items, fmt.Sprintf("Bộ nhớ làm việc:%d", len(working)))
	}
	if episodic, ok := result["episodic_memory"].(map[string]any); ok && len(episodic) > 0 {
		items = append(items, fmt.Sprintf("Bộ nhớ cốt truyện:%d", len(episodic)))
	}
	if planning, ok := result["planning_memory"].(map[string]any); ok && len(planning) > 0 {
		items = append(items, fmt.Sprintf("Bộ nhớ kế hoạch:%d", len(planning)))
	}
	if foundation, ok := result["foundation_memory"].(map[string]any); ok && len(foundation) > 0 {
		items = append(items, fmt.Sprintf("Bộ nhớ cơ bản:%d", len(foundation)))
	}

	// Tóm tắt phân lớp
	if n := countSlice("volume_summaries"); n > 0 {
		items = append(items, fmt.Sprintf("Tóm tắt quyển:%d", n))
	}
	if n := countSlice("arc_summaries"); n > 0 {
		items = append(items, fmt.Sprintf("Tóm tắt phần:%d", n))
	}
	if n := countSlice("recent_summaries"); n > 0 {
		items = append(items, fmt.Sprintf("Tóm tắt chương:%d", n))
	}

	// Dàn ý phân lớp
	if n := countSlice("layered_outline"); n > 0 {
		items = append(items, fmt.Sprintf("Dàn ý phân lớp:%d quyển", n))
	}

	// Dữ liệu trạng thái
	if n := countSlice("timeline"); n > 0 {
		items = append(items, fmt.Sprintf("Dòng thời gian:%d", n))
	}
	if n := countSlice("foreshadow_ledger"); n > 0 {
		items = append(items, fmt.Sprintf("Phục bút:%d", n))
	}
	if n := countSlice("relationship_state"); n > 0 {
		items = append(items, fmt.Sprintf("Mối quan hệ:%d", n))
	}
	if n := countSlice("recent_state_changes"); n > 0 {
		items = append(items, fmt.Sprintf("Thay đổi trạng thái:%d", n))
	}
	if _, ok := result["previous_tail"]; ok {
		items = append(items, "Đuôi chương trước:ok")
	}
	if _, ok := result["style_rules"]; ok {
		items = append(items, "Quy tắc phong cách:ok")
	}
	if n := sliceLen(result["related_chapters"]); n > 0 {
		items = append(items, fmt.Sprintf("Chương liên quan:%d", n))
	}
	if selected, ok := result["selected_memory"].(map[string]any); ok && len(selected) > 0 {
		if n := sliceLen(selected["story_threads"]); n > 0 {
			items = append(items, fmt.Sprintf("Thu hồi manh mối:%d", n))
		}
		if n := sliceLen(selected["review_lessons"]); n > 0 {
			items = append(items, fmt.Sprintf("Thu hồi đánh giá:%d", n))
		}
	}

	// Tài liệu tham khảo
	if refs, ok := result["references"].(map[string]string); ok && len(refs) > 0 {
		items = append(items, fmt.Sprintf("Tham khảo:%d mục", len(refs)))
	}
	if pack, ok := result["reference_pack"].(map[string]any); ok && len(pack) > 0 {
		items = append(items, fmt.Sprintf("Gói tham khảo:%d", len(pack)))
	}
	if _, ok := result["memory_policy"]; ok {
		items = append(items, "Chiến lược bộ nhớ:ok")
	}
	if _, ok := result["simulation_profile"]; ok {
		items = append(items, "Hồ sơ mô phỏng viết:ok")
	}
	if warnings, ok := result["_warnings"].([]string); ok && len(warnings) > 0 {
		items = append(items, fmt.Sprintf("Cảnh báo:%d", len(warnings)))
	}
	if trimmed, ok := result["_trimmed"].([]string); ok && len(trimmed) > 0 {
		items = append(items, fmt.Sprintf("Cắt xén:%s", strings.Join(trimmed, ",")))
	}

	if len(items) > 0 {
		parts = append(parts, strings.Join(items, " "))
	}
	return strings.Join(parts, " | ")
}

// sliceLen Thử lấy độ dài slice đối với kiểu any.
func sliceLen(v any) int {
	switch s := v.(type) {
	case []domain.ChapterSummary:
		return len(s)
	case []domain.ArcSummary:
		return len(s)
	case []domain.VolumeSummary:
		return len(s)
	case []domain.CharacterSnapshot:
		return len(s)
	case []domain.TimelineEvent:
		return len(s)
	case []domain.ForeshadowEntry:
		return len(s)
	case []domain.RelationshipEntry:
		return len(s)
	case []domain.StateChange:
		return len(s)
	case []domain.VolumeOutline:
		return len(s)
	case []domain.Character:
		return len(s)
	case []domain.RelatedChapter:
		return len(s)
	case []domain.RecallItem:
		return len(s)
	default:
		return 0
	}
}

// loadFilteredCharacters Lọc nhân vật theo Tier và cảnh xuất hiện.
// core/important luôn được trả về; secondary/decorative chỉ trả về khi được nhắc đến trong dàn ý chương hiện tại.
func (t *ContextTool) loadFilteredCharacters(result map[string]any, chapter int, warn func(string, error)) {
	chars, err := t.store.Characters.Load()
	if err != nil {
		warn("characters", err)
		return
	}
	if len(chars) == 0 {
		return
	}

	// Lấy mô tả cảnh của dàn ý chương hiện tại, dùng để khớp với các nhân vật phụ
	entry, err := t.store.Outline.GetChapterOutline(chapter)
	if err != nil {
		warn("current_chapter_outline", err)
		result["characters"] = chars
		return
	}
	sceneText := strings.Join(entry.Scenes, " ") + " " + entry.CoreEvent + " " + entry.Title

	var filtered []domain.Character
	for _, c := range chars {
		switch c.Tier {
		case "secondary", "decorative":
			if matchCharacter(sceneText, c) {
				filtered = append(filtered, c)
			}
		default: // core, important, hoặc chưa thiết lập
			filtered = append(filtered, c)
		}
	}
	result["characters"] = filtered
}

// matchCharacter Kiểm tra xem văn bản cảnh có chứa tên chính thức hoặc bất kỳ biệt danh nào của nhân vật hay không.
func matchCharacter(text string, c domain.Character) bool {
	if strings.Contains(text, c.Name) {
		return true
	}
	for _, alias := range c.Aliases {
		if strings.Contains(text, alias) {
			return true
		}
	}
	return false
}

// loadLayeredSummaries Tải tóm tắt phân lớp: tóm tắt quyển + tóm tắt phần của quyển hiện tại + tóm tắt chương trong phần.
func (t *ContextTool) loadLayeredSummaries(result map[string]any, chapter, summaryWindow int, warn func(string, error)) {
	vol, arc, err := t.store.Outline.LocateChapter(chapter)
	if err != nil {
		warn("layered_outline_position", err)
		// Lùi về chế độ phẳng
		if summaries, err := t.store.Summaries.LoadRecentSummaries(chapter, summaryWindow); err == nil && len(summaries) > 0 {
			result["recent_summaries"] = summaries
		} else {
			warn("recent_summaries", err)
		}
		return
	}

	// 1. Tóm tắt quyển của các quyển đã hoàn thành
	if volSummaries, err := t.store.Summaries.LoadAllVolumeSummaries(); err == nil && len(volSummaries) > 0 {
		result["volume_summaries"] = volSummaries
	} else {
		warn("volume_summaries", err)
	}

	// 2. Tóm tắt phần của các phần đã hoàn thành trong quyển hiện tại (không bao gồm phần hiện tại)
	if arcSummaries, err := t.store.Summaries.LoadArcSummaries(vol); err == nil && len(arcSummaries) > 0 {
		var prior []domain.ArcSummary
		for _, s := range arcSummaries {
			if s.Arc < arc {
				prior = append(prior, s)
			}
		}
		if len(prior) > 0 {
			result["arc_summaries"] = prior
		}
	} else {
		warn("arc_summaries", err)
	}

	// 3. Tóm tắt chương của N chương gần đây trong phần hiện tại
	if summaries, err := t.store.Summaries.LoadRecentSummaries(chapter, summaryWindow); err == nil && len(summaries) > 0 {
		result["recent_summaries"] = summaries
	} else {
		warn("recent_summaries", err)
	}
}

// loadLayeredCharacters Tải nhân vật trong chế độ Layered: ưu tiên dùng bản chụp gần nhất, lùi về thiết lập gốc + lọc theo Tier.
func (t *ContextTool) loadLayeredCharacters(result map[string]any, chapter int, warn func(string, error)) {
	snapshots, err := t.store.Characters.LoadLatestSnapshots()
	if err == nil && len(snapshots) > 0 {
		result["character_snapshots"] = snapshots
		// Đồng thời giữ lại các nhân vật core/important trong thiết lập gốc (bản chụp có thể không chứa nhân vật mới xuất hiện)
		t.loadFilteredCharacters(result, chapter, warn)
		return
	}
	warn("character_snapshots", err)
	// Khi không có bản chụp, lùi về thiết lập gốc
	t.loadFilteredCharacters(result, chapter, warn)
}

// writerReferences Trả về tài liệu tham khảo viết. Chương 1 trả về toàn bộ, các chương tiếp theo sẽ cắt bỏ các mẫu không còn cần thiết.
func (t *ContextTool) writerReferences(chapter int) map[string]string {
	refs := map[string]string{}
	add := func(k, v string) {
		if v != "" {
			refs[k] = v
		}
	}
	// Tải dần dần: luôn giữ lại tham chiếu cốt lõi, 3 chương đầu tải thêm hướng dẫn viết đầy đủ
	add("consistency", t.refs.Consistency)
	add("hook_techniques", t.refs.HookTechniques)
	add("quality_checklist", t.refs.QualityChecklist)
	add("anti_ai_tone", t.refs.AntiAITone) // Tiêu chí loại bỏ giọng AI được đưa vào toàn bộ, không cắt theo chương
	if chapter <= 3 {
		add("chapter_guide", t.refs.ChapterGuide)
		add("dialogue_writing", t.refs.DialogueWriting)
		add("style_reference", t.refs.StyleReference)
	}

	// Tài liệu tham khảo bổ sung chỉ tải ở chương đầu tiên
	if chapter <= 1 {
		add("chapter_template", t.refs.ChapterTemplate)
		add("content_expansion", t.refs.ContentExpansion)
	}
	return refs
}

func (t *ContextTool) architectReferences() map[string]string {
	refs := map[string]string{}
	add := func(k, v string) {
		if v != "" {
			refs[k] = v
		}
	}
	add("outline_template", t.refs.OutlineTemplate)
	add("character_template", t.refs.CharacterTemplate)
	add("longform_planning", t.refs.LongformPlanning)
	add("differentiation", t.refs.Differentiation)
	add("style_reference", t.refs.StyleReference)
	add("arc_templates", t.refs.ArcTemplates)
	add("anti_ai_tone", t.refs.AntiAITone) // Dàn ý architect loại bỏ giọng AI; cũng bao gồm luồng editor qua đường dẫn Chapter=0
	return refs
}

// foundationStatus Kiểm tra tính hoàn thiện của các cài đặt cơ bản, trả về danh sách các mục còn thiếu.
// Dùng chung logic phán đoán store.FoundationMissing với công cụ save_foundation, để đảm bảo LLM thấy
// ready/missing từ novel_context và foundation_ready từ save_foundation
// luôn nhất quán (các chi tiết như yêu cầu compass dài hạn sẽ không bị sai lệch).
func (t *ContextTool) foundationStatus() map[string]any {
	missing := t.store.FoundationMissing()
	status := map[string]any{"ready": len(missing) == 0}
	if len(missing) > 0 {
		status["missing"] = missing
	}
	return status
}

// ContextSummary Trả về tóm tắt ngắn gọn của trạng thái hiện tại (dùng cho log).
func (t *ContextTool) ContextSummary() string {
	var parts []string
	if p, _ := t.store.Outline.LoadPremise(); p != "" {
		parts = append(parts, "premise:ok")
	}
	if o, _ := t.store.Outline.LoadOutline(); o != nil {
		parts = append(parts, fmt.Sprintf("outline:%d chapters", len(o)))
	}
	if c, _ := t.store.Characters.Load(); c != nil {
		parts = append(parts, fmt.Sprintf("characters:%d", len(c)))
	}
	if len(parts) == 0 {
		return "empty"
	}
	return strings.Join(parts, ", ")
}

// trimByBudget Cắt giảm result theo ưu tiên, để tổng kích thước JSON không vượt quá budget byte.
// Ưu tiên (từ thấp đến cao): references < voice_samples < style_anchors < previous_tail < timeline
//
//	< recent_state_changes < foreshadow_ledger < relationship_state < Phần còn lại (không cắt giảm)
//
// Key đã bị cắt giảm sẽ được ghi lại vào result["_trimmed"] để kiểm tra log.
func trimByBudget(result map[string]any, budget int) {
	// Đo kích thước hiện tại trước
	data, err := json.Marshal(result)
	if err != nil || len(data) <= budget {
		return
	}

	// Liệt kê các key có thể cắt giảm theo mức độ ưu tiên từ thấp đến cao
	trimOrder := []string{
		"references",
		"voice_samples",
		"style_anchors",
		"style_rules",
		"style_stats",
		"previous_tail",
		"timeline",
		"recent_state_changes",
		"foreshadow_ledger",
		"relationship_state",
	}

	var trimmed []string
	for _, key := range trimOrder {
		if _, ok := result[key]; !ok {
			continue
		}
		deleteContextKey(result, key)
		trimmed = append(trimmed, key)
		data, err = json.Marshal(result)
		if err != nil || len(data) <= budget {
			break
		}
	}
	if len(trimmed) > 0 {
		result["_trimmed"] = trimmed
	}
}

func deleteContextKey(result map[string]any, key string) {
	delete(result, key)
	for _, containerKey := range []string{
		"working_memory",
		"episodic_memory",
		"planning_memory",
		"foundation_memory",
		"reference_pack",
	} {
		section, ok := result[containerKey].(map[string]any)
		if !ok {
			continue
		}
		delete(section, key)
	}
}

// buildRelatedChapters Dựa trên dữ liệu cấu trúc, tra ngược các chương lịch sử liên quan đến chương hiện tại.
// Đề xuất từ 4 khía cạnh: phục bút, nhân vật xuất hiện, thay đổi trạng thái, và mối quan hệ; sau khi loại bỏ trùng lặp sẽ trả về tối đa 5 mục.
// Tất cả dữ liệu được truyền qua tham số, không làm thêm IO nào khác.
func (t *ContextTool) buildRelatedChapters(
	chapter int,
	entry *domain.OutlineEntry,
	foreshadow []domain.ForeshadowEntry,
	relationships []domain.RelationshipEntry,
	stateChanges []domain.StateChange,
) []domain.RelatedChapter {
	const recentWindow = 10
	const maxResults = 5

	seen := make(map[int]struct{})
	var results []domain.RelatedChapter
	add := func(ch int, reason string) {
		if ch <= 0 || ch >= chapter {
			return
		}
		// Các chương gần đây quá sát, không đề xuất
		if ch > chapter-recentWindow {
			return
		}
		if _, ok := seen[ch]; ok {
			return
		}
		seen[ch] = struct{}{}
		results = append(results, domain.RelatedChapter{Chapter: ch, Reason: reason})
	}

	// Nối văn bản dàn ý để dùng cho việc khớp từ khóa
	outlineText := entry.Title + " " + entry.CoreEvent
	for _, s := range entry.Scenes {
		outlineText += " " + s
	}

	// 1. Tra ngược phục bút: mô tả của phục bút đang hoạt động có liên quan đến dàn ý chương hiện tại không
	for _, f := range foreshadow {
		if strings.Contains(outlineText, f.ID) || containsAny(outlineText, strings.Fields(f.Description)) {
			add(f.PlantedAt, fmt.Sprintf("Chương vùi phục bút %s(%s)", f.ID, truncateRunes(f.Description, 15)))
		}
		if len(results) >= maxResults {
			break
		}
	}

	// 2. Tra ngược nhân vật xuất hiện: duyệt một lần hàng loạt, giảm IO từ O(số nhân vật × số chương) xuống O(số chương)
	chars, _ := t.store.Characters.Load()
	outlineChars := matchOutlineCharacters(outlineText, chars)
	if len(outlineChars) > 0 {
		appearances := t.store.Summaries.FindCharacterAppearances(outlineChars, chapter, recentWindow)
		for _, name := range outlineChars {
			if len(results) >= maxResults {
				break
			}
			if ch, ok := appearances[name]; ok {
				add(ch, fmt.Sprintf("Chương xuất hiện cuối cùng của nhân vật '%s'", name))
			}
		}
	}

	// 3. Tra ngược thay đổi trạng thái: thao tác trên slice đã tải, không IO
	for _, name := range outlineChars {
		if len(results) >= maxResults {
			break
		}
		ch := findLastStateChange(stateChanges, name, chapter)
		if ch > 0 && ch <= chapter-recentWindow {
			add(ch, fmt.Sprintf("Chương thay đổi trạng thái của '%s'", name))
		}
	}

	// 4. Tra ngược mối quan hệ: thay đổi cuối cùng trong mối quan hệ giữa các nhân vật liên quan trong chương hiện tại
	if len(relationships) > 0 && len(outlineChars) >= 2 {
		charSet := make(map[string]struct{}, len(outlineChars))
		for _, c := range outlineChars {
			charSet[c] = struct{}{}
		}
		for _, r := range relationships {
			if len(results) >= maxResults {
				break
			}
			_, aIn := charSet[r.CharacterA]
			_, bIn := charSet[r.CharacterB]
			if aIn && bIn {
				add(r.Chapter, fmt.Sprintf("Thay đổi quan hệ giữa %s và %s", r.CharacterA, r.CharacterB))
			}
		}
	}

	return results
}

// findLastStateChange Tìm số chương thay đổi gần nhất của thực thể trong danh sách thay đổi trạng thái đã được tải.
func findLastStateChange(changes []domain.StateChange, entity string, currentChapter int) int {
	for i := len(changes) - 1; i >= 0; i-- {
		if changes[i].Entity == entity && changes[i].Chapter < currentChapter {
			return changes[i].Chapter
		}
	}
	return 0
}

// matchOutlineCharacters Khớp tên nhân vật xuất hiện từ văn bản dàn ý.
func matchOutlineCharacters(text string, chars []domain.Character) []string {
	var matched []string
	for _, c := range chars {
		if strings.Contains(text, c.Name) {
			matched = append(matched, c.Name)
			continue
		}
		for _, alias := range c.Aliases {
			if strings.Contains(text, alias) {
				matched = append(matched, c.Name)
				break
			}
		}
	}
	return matched
}

// containsAny Kiểm tra xem text có chứa bất kỳ từ nào trong words hay không (chỉ khớp nếu có ít nhất 2 chữ cái để tránh nhiễu).
func containsAny(text string, words []string) bool {
	for _, w := range words {
		if len([]rune(w)) >= 2 && strings.Contains(text, w) {
			return true
		}
	}
	return false
}

func (t *ContextTool) selectStoryThreads(state contextBuildState) []domain.RecallItem {
	if state.currentEntry == nil {
		return nil
	}
	if len(state.foreshadow) < storyThreadRecallThreshold {
		return nil
	}

	const maxThreads = 5
	var items []domain.RecallItem
	seen := make(map[string]struct{})
	picked := make(map[string]struct{}) // Các ID phục bút đã được chọn, để loại bỏ trùng lặp khi bổ sung dựa trên độ cũ
	add := func(item domain.RecallItem) {
		key := item.Kind + "|" + item.Key + "|" + item.Summary
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		picked[item.Key] = struct{}{}
		items = append(items, item)
	}

	// 1. Thu hồi theo tính liên quan: các phục bút có từ khóa trùng lặp với từ khóa tập trung của chương hiện tại.
	focusTerms := recallFocusTerms(state.currentEntry, state.chapterPlan)
	focusText := strings.Join(focusTerms, " ")
	for _, entry := range state.foreshadow {
		if !matchesRecallTerms(entry.ID+" "+entry.Description, focusTerms) && !strings.Contains(focusText, entry.ID) {
			continue
		}
		add(domain.RecallItem{
			Kind:    "story_thread",
			Key:     entry.ID,
			Chapter: entry.PlantedAt,
			Reason:  "Chương hiện tại có thể cần tiếp nối phục bút hiện có",
			Summary: fmt.Sprintf("Phục bút “%s” đã vùi ở chương %d: %s", entry.ID, entry.PlantedAt, truncateRunes(entry.Description, 80)),
		})
		if len(items) >= maxThreads {
			return items
		}
	}

	// 2. Bổ sung dựa trên độ cũ: các phục bút không liên quan đến chương hiện tại nhưng đã để quá lâu chưa thu hồi (ưu tiên cũ nhất), lấp đầy các vị trí còn lại.
	//    Bổ sung cho điểm mù tự nhiên của phương pháp thu hồi theo tính liên quan — một tuyến truyện bị treo quá lâu nhưng lại không vô tình khớp với từ khóa trong chương này.
	for _, entry := range agingForeshadow(state.foreshadow, state.chapter, picked) {
		add(domain.RecallItem{
			Kind:    "story_thread",
			Key:     entry.ID,
			Chapter: entry.PlantedAt,
			Reason:  "Phục bút treo lâu chưa thu hồi, chú ý thúc đẩy hoặc thu hồi kịp thời",
			Summary: fmt.Sprintf("Phục bút “%s” đã vùi ở chương %d, đã %d chương chưa thu hồi: %s", entry.ID, entry.PlantedAt, state.chapter-entry.PlantedAt, truncateRunes(entry.Description, 80)),
		})
		if len(items) >= maxThreads {
			break
		}
	}

	return items
}

// agingForeshadow Trả về các phục bút chưa thu hồi có tuổi thọ ≥ foreshadowAgingChapters, được sắp xếp theo thứ tự cũ nhất ưu tiên,
// bỏ qua các mục trong picked đã được chọn qua thu hồi theo tính liên quan. Tham số đầu vào all đã là danh sách active (chưa thu hồi), nên không cần lọc lại trạng thái.
func agingForeshadow(all []domain.ForeshadowEntry, chapter int, picked map[string]struct{}) []domain.ForeshadowEntry {
	var aging []domain.ForeshadowEntry
	for _, e := range all {
		if _, ok := picked[e.ID]; ok {
			continue
		}
		if e.PlantedAt <= 0 || chapter-e.PlantedAt < foreshadowAgingChapters {
			continue
		}
		aging = append(aging, e)
	}
	sort.SliceStable(aging, func(i, j int) bool {
		return aging[i].PlantedAt < aging[j].PlantedAt
	})
	return aging
}

func (t *ContextTool) selectReviewLessons(chapter int, warn func(string, error)) []domain.RecallItem {
	if chapter <= 1 {
		return nil
	}

	var items []domain.RecallItem
	seen := make(map[string]struct{})
	add := func(item domain.RecallItem) {
		key := item.Summary
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		items = append(items, item)
	}

	appendReview := func(review *domain.ReviewEntry) bool {
		if review == nil {
			return false
		}
		for i, miss := range review.ContractMisses {
			add(domain.RecallItem{
				Kind:    "review_lesson",
				Key:     fmt.Sprintf("review-%d-contract-%d", review.Chapter, i),
				Chapter: review.Chapter,
				Reason:  "Bài đánh giá gần đây chỉ ra thiếu sót hợp đồng",
				Summary: fmt.Sprintf("Chương %d thiếu sót hợp đồng: %s", review.Chapter, miss),
			})
			if len(items) >= 3 {
				return true
			}
		}
		for i, issue := range review.Issues {
			switch issue.Severity {
			case "", "warning", "error", "critical":
				add(domain.RecallItem{
					Kind:    "review_lesson",
					Key:     fmt.Sprintf("review-%d-issue-%d", review.Chapter, i),
					Chapter: review.Chapter,
					Reason:  "Bài đánh giá gần đây chỉ ra cần tránh các vấn đề lặp lại",
					Summary: fmt.Sprintf("Chương %d nhắc nhở đánh giá: %s", review.Chapter, truncateRunes(issue.Description, 80)),
				})
			}
			if len(items) >= 3 {
				return true
			}
		}
		return false
	}

	for ch := chapter - 1; ch >= max(chapter-3, 1); ch-- {
		review, err := t.store.World.LoadReview(ch)
		if err != nil {
			warn("review", err)
			continue
		}
		if appendReview(review) {
			return items
		}
	}

	globalReview, err := t.store.World.LoadLastReview(chapter - 1)
	if err != nil {
		warn("global_review", err)
	} else if appendReview(globalReview) {
		return items
	}
	return items
}

func recallFocusTerms(entry *domain.OutlineEntry, plan *domain.ChapterPlan) []string {
	if entry == nil {
		return nil
	}
	var terms []string
	add := func(v string) {
		v = strings.TrimSpace(v)
		if v != "" {
			terms = append(terms, v)
		}
	}

	add(entry.Title)
	add(entry.CoreEvent)
	add(entry.Hook)
	for _, scene := range entry.Scenes {
		add(scene)
	}
	if plan != nil {
		add(plan.Goal)
		add(plan.Hook)
		for _, point := range plan.Contract.PayoffPoints {
			add(point)
		}
		add(plan.Contract.HookGoal)
	}
	return terms
}

func matchesRecallTerms(text string, terms []string) bool {
	for _, term := range terms {
		term = strings.TrimSpace(term)
		if len([]rune(term)) < 2 {
			continue
		}
		if strings.Contains(text, term) || strings.Contains(term, text) {
			return true
		}
		if hasMeaningfulOverlap(term, text) {
			return true
		}
	}
	return false
}

func hasMeaningfulOverlap(a, b string) bool {
	ar := []rune(strings.TrimSpace(a))
	br := []rune(strings.TrimSpace(b))
	if len(ar) < 5 || len(br) < 5 {
		return false
	}
	shorter := len(ar)
	if len(br) < shorter {
		shorter = len(br)
	}
	threshold := 5
	switch {
	case shorter >= 12:
		threshold = 7
	case shorter >= 9:
		threshold = 6
	}
	return longestCommonSubstringRunes(ar, br) >= threshold
}

const storyThreadRecallThreshold = 6
const storyThreadRecallMinSelected = 2

// foreshadowAgingChapters: Một phục bút kể từ lúc vùi nếu vượt quá ngần này chương mà vẫn chưa thu hồi thì được coi là "treo quá lâu".
// Loại phục bút này ngay cả khi không liên quan đến từ khóa chương hiện tại, cũng sẽ được bổ sung vào story_threads, tránh bị lãng quên hoàn toàn trong một truyện dài
// (Thu hồi theo tính liên quan bản chất chỉ nhìn thấy các tuyến liên quan đến chương này, không thể thấy những tuyến truyện đã treo một mình quá lâu).
// Độ tuổi là sự thật bắt nguồn từ logic code (chương hiện tại - chương vùi phục bút), chỉ thông báo là "đã treo N chương chưa thu hồi", không ra lệnh.
const foreshadowAgingChapters = 30

func longestCommonSubstringRunes(a, b []rune) int {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	prev := make([]int, len(b)+1)
	best := 0
	for i := 1; i <= len(a); i++ {
		curr := make([]int, len(b)+1)
		for j := 1; j <= len(b); j++ {
			if a[i-1] != b[j-1] {
				continue
			}
			curr[j] = prev[j-1] + 1
			if curr[j] > best {
				best = curr[j]
			}
		}
		prev = curr
	}
	return best
}

// truncateRunes Cắt xén chuỗi ký tự theo số lượng rune chỉ định.
func truncateRunes(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return string(runes[:maxRunes]) + "..."
}
