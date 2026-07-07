package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/voocel/ainovel-cli/internal/store"
)

func TestSaveArcSummaryPersistsStyleRulesDialogueObjects(t *testing.T) {
	s := store.NewStore(t.TempDir())
	if err := s.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}

	tool := NewSaveArcSummaryTool(s)
	args, err := json.Marshal(map[string]any{
		"volume":     1,
		"arc":        2,
		"title":      "Vào núi",
		"summary":    "Nhân vật chính hoàn thành thử thách vào núi, xác định hướng truy xét tiếp theo.",
		"key_events": []string{"Vượt qua thử thách", "Phát hiện manh mối án cũ"},
		"character_snapshots": []map[string]any{
			{"name": "Thẩm Uyên", "status": "Sống sót", "motivation": "Truy tra án cũ"},
		},
		"style_rules": map[string]any{
			"prose": []string{"Miêu tả bối cảnh ưu tiên xúc giác và khứu giác", "Cảnh hành động dùng câu ngắn để thúc đẩy", "Miêu tả tâm lý không giải thích kết luận"},
			"dialogue": []map[string]any{
				{"name": "Thẩm Uyên", "rules": []string{"Hội thoại cực kỳ tối giản", "Ít dùng câu hỏi"}},
			},
			"taboos": []string{"Tránh độc thoại dài ở cuối chương"},
		},
	})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	if _, err := tool.Execute(context.Background(), args); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	rules, err := s.World.LoadStyleRules()
	if err != nil {
		t.Fatalf("LoadStyleRules: %v", err)
	}
	if rules == nil || len(rules.Dialogue) != 1 {
		t.Fatalf("expected one dialogue rule, got %+v", rules)
	}
	if rules.Dialogue[0].Name != "Thẩm Uyên" || len(rules.Dialogue[0].Rules) != 2 {
		t.Fatalf("unexpected dialogue rule: %+v", rules.Dialogue[0])
	}
}

func TestSaveArcSummaryRejectsDialogueStringArray(t *testing.T) {
	s := store.NewStore(t.TempDir())
	if err := s.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}

	tool := NewSaveArcSummaryTool(s)
	args, err := json.Marshal(map[string]any{
		"volume":              1,
		"arc":                 2,
		"title":               "Vào núi",
		"summary":             "Nhân vật chính hoàn thành thử thách vào núi, xác định hướng truy xét tiếp theo.",
		"key_events":          []string{"Vượt qua thử thách"},
		"character_snapshots": []map[string]any{},
		"style_rules": map[string]any{
			"prose":    []string{"Miêu tả bối cảnh ưu tiên xúc giác và khứu giác"},
			"dialogue": []string{"Thẩm Uyên hội thoại cực kỳ tối giản"},
		},
	})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	if _, err := tool.Execute(context.Background(), args); err == nil || !strings.Contains(err.Error(), "style_rules.dialogue") {
		t.Fatalf("expected style_rules.dialogue validation error, got %v", err)
	}
}
