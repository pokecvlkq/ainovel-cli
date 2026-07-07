package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/voocel/agentcore/schema"
	"github.com/voocel/ainovel-cli/internal/domain"
	"github.com/voocel/ainovel-cli/internal/errs"
	"github.com/voocel/ainovel-cli/internal/store"
)

// SaveFoundationTool lưu thiết lập cơ bản (premise/outline/characters), dành riêng cho Architect.
type SaveFoundationTool struct {
	store *store.Store
}

func NewSaveFoundationTool(store *store.Store) *SaveFoundationTool {
	return &SaveFoundationTool{store: store}
}

func (t *SaveFoundationTool) Name() string { return "save_foundation" }
func (t *SaveFoundationTool) Description() string {
	return "Lưu thiết lập cơ bản của tiểu thuyết (premise/outline/characters/world_rules/compass...). **Đây là lối vào duy nhất để lưu trữ vĩnh viễn**: Nội dung chưa được công cụ này lưu sẽ không vào store, chỉ xuất dưới dạng Markdown/JSON trong tin nhắn đồng nghĩa với việc bị mất. Tham số cố định là {type, content, scale?, volume?, arc?}. type tùy chọn premise / outline / layered_outline / characters / world_rules / expand_arc / append_volume / update_compass / complete_book. Với premise, content phải là chuỗi Markdown; các loại khác ưu tiên truyền thẳng JSON array hoặc object. expand_arc dùng để mở rộng chi tiết chương của một arc (cần volume + arc); append_volume thêm tập mới (content là JSON VolumeOutline đầy đủ); update_compass cập nhật hướng đi cuối (content là JSON StoryCompass); complete_book khai báo kết thúc toàn bộ truyện (content truyền {} rỗng, đẩy thẳng Phase=Complete; trước khi gọi phải qua danh sách kiểm tra tập cuối và không có hàng đợi làm lại). scale tùy chọn, chỉ cho phép short / mid / long."
}
func (t *SaveFoundationTool) Label() string { return "Lưu thiết lập" }

// Công cụ ghi (cập nhật chéo Outline/Progress/Characters), không hỗ trợ đồng thời.
func (t *SaveFoundationTool) ReadOnly(_ json.RawMessage) bool        { return false }
func (t *SaveFoundationTool) ConcurrencySafe(_ json.RawMessage) bool { return false }

func (t *SaveFoundationTool) Schema() map[string]any {
	return schema.Object(
		schema.Property("type", schema.Enum("Loại thiết lập", "premise", "outline", "layered_outline", "characters", "world_rules", "expand_arc", "append_volume", "update_compass", "complete_book")).Required(),
		schema.Property("content", map[string]any{
			"description": "Nội dung. premise truyền chuỗi Markdown; các loại khác truyền thẳng mảng hoặc đối tượng JSON, hỗ trợ cả chuỗi JSON. Khi expand_arc thì truyền mảng chương.",
		}).Required(),
		schema.Property("scale", schema.Enum("Cấp độ quy hoạch", "short", "mid", "long")),
		schema.Property("volume", schema.Int("Số thứ tự tập đích (chỉ bắt buộc khi expand_arc)")),
		schema.Property("arc", schema.Int("Số thứ tự arc đích (chỉ bắt buộc khi expand_arc)")),
	)
}

func (t *SaveFoundationTool) Execute(_ context.Context, args json.RawMessage) (json.RawMessage, error) {
	var a struct {
		Type    string          `json:"type"`
		Content json.RawMessage `json:"content"`
		Scale   string          `json:"scale"`
		Volume  int             `json:"volume"`
		Arc     int             `json:"arc"`
	}
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, fmt.Errorf("invalid args: %w: %w", errs.ErrToolArgs, err)
	}
	content, err := normalizeFoundationContent(a.Content)
	if err != nil {
		return nil, err
	}
	if a.Scale != "" {
		switch domain.PlanningTier(a.Scale) {
		case domain.PlanningTierShort, domain.PlanningTierMid, domain.PlanningTierLong:
		default:
			return nil, fmt.Errorf("invalid scale %q, expected short/mid/long: %w", a.Scale, errs.ErrToolArgs)
		}
		if err := t.store.RunMeta.SetPlanningTier(domain.PlanningTier(a.Scale)); err != nil {
			return nil, fmt.Errorf("save planning tier: %w: %w", errs.ErrStoreWrite, err)
		}
	}

	result := map[string]any{"saved": true, "type": a.Type, "scale": a.Scale}

	// Trong giai đoạn viết, cấm ghi đè toàn bộ đại cương, chỉ cho phép thao tác tăng dần (expand_arc / append_volume)
	if (a.Type == "outline" || a.Type == "layered_outline") && t.isWriting() {
		return nil, fmt.Errorf(
			"Giai đoạn viết cấm dùng %s để ghi đè toàn bộ đại cương. Vui lòng dùng expand_arc để mở rộng chi tiết hoặc append_volume để thêm tập mới: %w", a.Type, errs.ErrToolPrecondition)
	}

	decode := func(typeName string, out any) error {
		return decodeFoundationJSON(typeName, content, out)
	}

	switch a.Type {
	case "premise":
		name := domain.ExtractNovelNameFromPremise(content)
		if err := t.store.Outline.SavePremise(content); err != nil {
			return nil, fmt.Errorf("save premise: %w: %w", errs.ErrStoreWrite, err)
		}
		if name != "" {
			_ = t.store.Progress.SetNovelName(name)
			result["novel_name"] = name
		}
		_ = t.store.Progress.UpdatePhase(domain.PhasePremise)

	case "outline":
		var entries []domain.OutlineEntry
		if err := decode("outline", &entries); err != nil {
			return nil, err
		}
		if err := t.store.Outline.SaveOutline(entries); err != nil {
			return nil, fmt.Errorf("save outline: %w: %w", errs.ErrStoreWrite, err)
		}
		_ = t.store.Progress.UpdatePhase(domain.PhaseOutline)
		_ = t.store.Progress.SetTotalChapters(len(entries))
		if domain.PlanningTier(a.Scale) != domain.PlanningTierLong {
			_ = t.store.Progress.SetLayered(false)
			_ = t.store.Progress.UpdateVolumeArc(0, 0)
			_ = t.store.Outline.ClearLayeredOutline()
		}
		result["chapters"] = len(entries)

	case "layered_outline":
		var volumes []domain.VolumeOutline
		if err := decode("layered_outline", &volumes); err != nil {
			return nil, err
		}
		if err := t.store.Outline.SaveLayeredOutline(volumes); err != nil {
			return nil, fmt.Errorf("save layered_outline: %w: %w", errs.ErrStoreWrite, err)
		}
		flat := domain.FlattenOutline(volumes)
		if err := t.store.Outline.SaveOutline(flat); err != nil {
			return nil, fmt.Errorf("save flattened outline: %w: %w", errs.ErrStoreWrite, err)
		}
		total := domain.TotalChapters(volumes)
		_ = t.store.Progress.UpdatePhase(domain.PhaseOutline)
		_ = t.store.Progress.SetTotalChapters(total)
		_ = t.store.Progress.SetLayered(true)
		if len(volumes) > 0 && len(volumes[0].Arcs) > 0 {
			_ = t.store.Progress.UpdateVolumeArc(volumes[0].Index, volumes[0].Arcs[0].Index)
		}
		result["volumes"] = len(volumes)
		result["chapters"] = total

	case "characters":
		var chars []domain.Character
		if err := decode("characters", &chars); err != nil {
			return nil, err
		}
		if err := t.store.Characters.Save(chars); err != nil {
			return nil, fmt.Errorf("save characters: %w: %w", errs.ErrStoreWrite, err)
		}
		result["count"] = len(chars)

	case "world_rules":
		var rules []domain.WorldRule
		if err := decode("world_rules", &rules); err != nil {
			return nil, err
		}
		if err := t.store.World.SaveWorldRules(rules); err != nil {
			return nil, fmt.Errorf("save world_rules: %w: %w", errs.ErrStoreWrite, err)
		}
		result["count"] = len(rules)

	case "expand_arc":
		if a.Volume <= 0 || a.Arc <= 0 {
			return nil, fmt.Errorf("expand_arc requires volume and arc parameters: %w", errs.ErrToolArgs)
		}
		var chapters []domain.OutlineEntry
		if err := decode("expand_arc chapters", &chapters); err != nil {
			return nil, err
		}
		if err := t.store.ExpandArc(a.Volume, a.Arc, chapters); err != nil {
			return nil, fmt.Errorf("expand arc: %w: %w", errs.ErrStoreWrite, err)
		}
		result["volume"] = a.Volume
		result["arc"] = a.Arc
		result["chapters"] = len(chapters)

	case "append_volume":
		if p, _ := t.store.Progress.Load(); p != nil && p.Phase == domain.PhaseComplete {
			return nil, fmt.Errorf("Toàn bộ truyện đã hoàn thành (phase=complete), không được phép thêm tập mới: %w", errs.ErrToolPrecondition)
		}
		var vol domain.VolumeOutline
		if err := decode("append_volume", &vol); err != nil {
			return nil, err
		}
		if err := t.store.AppendVolume(vol); err != nil {
			return nil, fmt.Errorf("append volume: %w: %w", errs.ErrStoreWrite, err)
		}
		result["volume"] = vol.Index
		result["arcs"] = len(vol.Arcs)
		chCount := 0
		for _, arc := range vol.Arcs {
			chCount += len(arc.Chapters)
		}
		if chCount > 0 {
			result["chapters"] = chCount
		}

	case "complete_book":
		// Lối vào duy nhất để hoàn thành truyện: đẩy thẳng Phase=Complete.
		// Chỉ được phép gọi trong giai đoạn Writing, ngăn giai đoạn Planning gọi nhầm bỏ qua toàn bộ việc viết.
		// Từ chối khi có hàng đợi làm lại —— đảm bảo PendingRewrites chạy xong mới kết thúc.
		progress, perr := t.store.Progress.Load()
		if perr != nil {
			return nil, fmt.Errorf("load progress: %w: %w", errs.ErrStoreRead, perr)
		}
		if progress == nil {
			return nil, fmt.Errorf("progress chưa khởi tạo: %w", errs.ErrToolPrecondition)
		}
		if progress.Phase != domain.PhaseWriting {
			return nil, fmt.Errorf("complete_book chỉ có thể gọi trong giai đoạn writing (phase hiện tại=%s): %w", progress.Phase, errs.ErrToolPrecondition)
		}
		if len(progress.PendingRewrites) > 0 {
			return nil, fmt.Errorf("còn %d chương trong hàng đợi làm lại, xử lý xong mới gọi complete_book: %w", len(progress.PendingRewrites), errs.ErrToolPrecondition)
		}
		if err := t.store.Progress.MarkComplete(); err != nil {
			return nil, fmt.Errorf("mark complete: %w: %w", errs.ErrStoreWrite, err)
		}
		result["book_complete"] = true
		result["phase"] = string(domain.PhaseComplete)

	case "update_compass":
		var compass domain.StoryCompass
		if err := decode("compass", &compass); err != nil {
			return nil, err
		}
		// Ở tầng công cụ ép buộc ghi đè LastUpdated bằng số chương đã hoàn thành, không tin tưởng LLM tự điền.
		// LLM thường quên điền hoặc để 0, khiến diag.CompassDrift báo sai, Router định hướng sai.
		if p, _ := t.store.Progress.Load(); p != nil {
			compass.LastUpdated = p.LatestCompleted()
		}
		if err := t.store.Outline.SaveCompass(compass); err != nil {
			return nil, fmt.Errorf("save compass: %w: %w", errs.ErrStoreWrite, err)
		}
		result["ending_direction"] = compass.EndingDirection
		result["last_updated"] = compass.LastUpdated

	default:
		return nil, fmt.Errorf("unknown type %q, expected premise/outline/layered_outline/characters/world_rules/expand_arc/append_volume/update_compass/complete_book: %w", a.Type, errs.ErrToolArgs)
	}

	// checkpoint
	scope := domain.GlobalScope()
	switch a.Type {
	case "expand_arc":
		scope = domain.ArcScope(a.Volume, a.Arc)
	case "append_volume":
		scope = domain.GlobalScope()
	}
	if _, err := t.store.Checkpoints.AppendArtifact(scope, a.Type, foundationArtifact(a.Type)); err != nil {
		return nil, fmt.Errorf("checkpoint foundation %s: %w: %w", a.Type, errs.ErrStoreWrite, err)
	}

	// Trả về các mục còn thiếu, hướng dẫn Architect tiếp tục hoặc kết thúc;
	// Khi đầy đủ thì đẩy thẳng phase sang writing, tránh việc Coordinator phân công lại.
	remaining := t.store.FoundationMissing()
	ready := len(remaining) == 0
	result["remaining"] = remaining
	result["foundation_ready"] = ready
	if ready {
		if p, _ := t.store.Progress.Load(); p != nil &&
			p.Phase != domain.PhaseWriting && p.Phase != domain.PhaseComplete {
			_ = t.store.Progress.UpdatePhase(domain.PhaseWriting)
			result["phase"] = string(domain.PhaseWriting)
		}
	}
	return json.Marshal(result)
}

func foundationArtifact(t string) string {
	switch t {
	case "premise":
		return "premise.md"
	case "outline":
		return "outline.json"
	case "layered_outline", "expand_arc", "append_volume":
		return "layered_outline.json"
	case "complete_book":
		return "meta/progress.json"
	case "characters":
		return "characters.json"
	case "world_rules":
		return "world_rules.json"
	case "update_compass":
		return "meta/compass.json"
	default:
		return ""
	}
}

// decodeFoundationJSON phân giải trường content của save_foundation, khi thất bại sẽ kèm vị trí dòng cột
// và gợi ý sửa chữa phổ biến nhất, giúp LLM lần thử sau định vị trực tiếp chứ không đoán mò.
func decodeFoundationJSON(typeName, content string, out any) error {
	err := json.Unmarshal([]byte(content), out)
	if err == nil {
		return nil
	}
	hint := `Nguyên nhân thường gặp: Dấu ngoặc kép trong chuỗi chưa được escape (\"), xuống dòng chưa escape (\n), hoặc thiếu dấu phẩy giữa các trường. Vui lòng tạo lại toàn bộ đoạn.`
	if se, ok := err.(*json.SyntaxError); ok {
		line, col := offsetToLineCol(content, int(se.Offset))
		return fmt.Errorf("parse %s JSON (line %d col %d): %w — %s", typeName, line, col, err, hint)
	}
	return fmt.Errorf("parse %s JSON: %w — %s", typeName, err, hint)
}

func offsetToLineCol(s string, offset int) (int, int) {
	if offset < 0 {
		offset = 0
	}
	if offset > len(s) {
		offset = len(s)
	}
	line, col := 1, 1
	for i := 0; i < offset; i++ {
		if s[i] == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}
	return line, col
}

func normalizeFoundationContent(raw json.RawMessage) (string, error) {
	if len(raw) == 0 {
		return "", fmt.Errorf("content is required: %w", errs.ErrToolArgs)
	}

	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return text, nil
	}

	if !json.Valid(raw) {
		return "", fmt.Errorf("invalid content: expected Markdown string or valid JSON value: %w", errs.ErrToolArgs)
	}
	return string(raw), nil
}

func (t *SaveFoundationTool) isWriting() bool {
	p, _ := t.store.Progress.Load()
	return p != nil && p.Phase == domain.PhaseWriting
}
