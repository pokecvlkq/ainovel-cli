package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/voocel/ainovel-cli/internal/domain"
	"github.com/voocel/ainovel-cli/internal/store"
)

func TestCommitChapterSchemaDescribesFeedbackAsObject(t *testing.T) {
	tool := NewCommitChapterTool(store.NewStore(t.TempDir()))
	schema := tool.Schema()
	props, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("schema properties missing: %#v", schema["properties"])
	}
	feedback, ok := props["feedback"].(map[string]any)
	if !ok {
		t.Fatalf("feedback schema missing: %#v", props["feedback"])
	}
	desc, _ := feedback["description"].(string)
	if !strings.Contains(desc, "JSON object") || !strings.Contains(desc, "JSON đã string hóa") {
		t.Fatalf("feedback description should warn against stringified JSON, got %q", desc)
	}
	if got := feedback["type"]; got != "object" {
		t.Fatalf("feedback type = %v, want object", got)
	}
}

func TestCommitChapterRejectsNonPendingRewrite(t *testing.T) {
	dir := t.TempDir()
	store := store.NewStore(dir)
	if err := store.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if err := store.Progress.Init("test", 10); err != nil {
		t.Fatalf("InitProgress: %v", err)
	}
	if err := store.Progress.MarkChapterComplete(2, 3000, "", ""); err != nil {
		t.Fatalf("MarkChapterComplete: %v", err)
	}
	if err := store.Progress.SetPendingRewrites([]int{2}, "Kiểm tra viết lại"); err != nil {
		t.Fatalf("SetPendingRewrites: %v", err)
	}
	if err := store.Progress.SetFlow(domain.FlowRewriting); err != nil {
		t.Fatalf("SetFlow: %v", err)
	}
	if err := store.Drafts.SaveDraft(3, "Đây là văn bản của chương sai."); err != nil {
		t.Fatalf("SaveDraft: %v", err)
	}

	tool := NewCommitChapterTool(store)
	args, err := json.Marshal(map[string]any{
		"chapter":         3,
		"summary":         "Commit lỗi",
		"characters":      []string{"Nhân vật chính"},
		"key_events":      []string{"Commit nhầm"},
		"timeline_events": []any{},
	})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	if _, err := tool.Execute(context.Background(), args); err == nil {
		t.Fatal("expected commit to be rejected during rewrite flow")
	}

	if _, err := os.Stat(dir + "/chapters/03.md"); !os.IsNotExist(err) {
		t.Fatalf("chapter should not be persisted, stat err=%v", err)
	}

	progress, err := store.Progress.Load()
	if err != nil {
		t.Fatalf("LoadProgress: %v", err)
	}
	if len(progress.CompletedChapters) != 1 || progress.CompletedChapters[0] != 2 {
		t.Fatalf("completed chapters should only contain original chapter 2, got %v", progress.CompletedChapters)
	}
	if progress.CurrentChapter != 3 {
		t.Fatalf("current chapter should not advance beyond original progress, got %d", progress.CurrentChapter)
	}
}

func TestCommitChapterAllowsPendingRewrite(t *testing.T) {
	dir := t.TempDir()
	store := store.NewStore(dir)
	if err := store.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if err := store.Progress.Init("test", 10); err != nil {
		t.Fatalf("InitProgress: %v", err)
	}
	if err := store.Progress.MarkChapterComplete(2, 3000, "", ""); err != nil {
		t.Fatalf("MarkChapterComplete: %v", err)
	}
	if err := store.Progress.SetPendingRewrites([]int{2}, "Kiểm tra viết lại"); err != nil {
		t.Fatalf("SetPendingRewrites: %v", err)
	}
	if err := store.Progress.SetFlow(domain.FlowRewriting); err != nil {
		t.Fatalf("SetFlow: %v", err)
	}
	if err := store.Drafts.SaveDraft(2, "Đây là văn bản của chương chờ viết lại đúng."); err != nil {
		t.Fatalf("SaveDraft: %v", err)
	}

	tool := NewCommitChapterTool(store)
	args, err := json.Marshal(map[string]any{
		"chapter":         2,
		"summary":         "Commit đúng",
		"characters":      []string{"Nhân vật chính"},
		"key_events":      []string{"Hoàn thành viết lại"},
		"timeline_events": []any{},
	})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	if _, err := tool.Execute(context.Background(), args); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if _, err := os.Stat(dir + "/chapters/02.md"); err != nil {
		t.Fatalf("chapter should be persisted: %v", err)
	}

	progress, err := store.Progress.Load()
	if err != nil {
		t.Fatalf("LoadProgress: %v", err)
	}
	if len(progress.CompletedChapters) != 1 || progress.CompletedChapters[0] != 2 {
		t.Fatalf("unexpected completed chapters: %v", progress.CompletedChapters)
	}
	pending, err := store.Signals.LoadPendingCommit()
	if err != nil {
		t.Fatalf("LoadPendingCommit: %v", err)
	}
	if pending != nil {
		t.Fatalf("expected pending commit cleared, got %+v", pending)
	}
}

// TestCommitChapterUpdatesCastLedger Xác thực: commit_chapter tích luỹ các characters trong chương vào cast_ledger,
// brief_role do cast_intros cung cấp sẽ được sử dụng, và các core characters trong characters.json không đi vào ledger.
func TestCommitChapterUpdatesCastLedger(t *testing.T) {
	dir := t.TempDir()
	s := store.NewStore(dir)
	if err := s.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if err := s.Progress.Init("test", 10); err != nil {
		t.Fatalf("InitProgress: %v", err)
	}
	// Thiết lập tệp nhân vật chính (không được thêm vào cast_ledger)
	if err := s.Characters.Save([]domain.Character{
		{Name: "林墨", Role: "Nhân vật chính", Tier: "core"},
		{Name: "李清砚", Role: "Người hướng dẫn", Tier: "important"},
	}); err != nil {
		t.Fatalf("Save core characters: %v", err)
	}
	if err := s.Drafts.SaveDraft(1, "Văn bản chương đầu, Lâm Mặc gặp ông chủ nhà trọ Lão Chu và tiểu nhị A Vân."); err != nil {
		t.Fatalf("SaveDraft: %v", err)
	}

	tool := NewCommitChapterTool(s)
	args, _ := json.Marshal(map[string]any{
		"chapter":    1,
		"summary":    "Lâm Mặc trọ tại nhà trọ",
		"characters": []string{"林墨", "李清砚", "老周", "阿云"},
		"key_events": []string{"Trọ"},
		"cast_intros": []any{
			map[string]any{"name": "老周", "brief_role": "Ông chủ nhà trọ"},
			map[string]any{"name": "阿云", "brief_role": "Tiểu nhị nhà trọ"},
		},
	})
	if _, err := tool.Execute(context.Background(), args); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	entries, err := s.Cast.Load()
	if err != nil {
		t.Fatalf("Cast.Load: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 ledger entries (老周/阿云), got %d: %+v", len(entries), entries)
	}
	byName := map[string]domain.CastEntry{}
	for _, e := range entries {
		byName[e.Name] = e
	}
	if e, ok := byName["老周"]; !ok || e.BriefRole != "Ông chủ nhà trọ" || e.FirstSeenChapter != 1 {
		t.Errorf("老周 entry wrong: %+v", e)
	}
	if e, ok := byName["阿云"]; !ok || e.BriefRole != "Tiểu nhị nhà trọ" || e.AppearanceCount != 1 {
		t.Errorf("阿云 entry wrong: %+v", e)
	}
	if _, ok := byName["林墨"]; ok {
		t.Errorf("Nhân vật chính Lâm Mặc không nên vào ledger")
	}
	if _, ok := byName["李清砚"]; ok {
		t.Errorf("Nhân vật chính Lý Thanh Nghiên không nên vào ledger")
	}
}

func TestCommitChapterReplayAfterPartialCommitDoesNotDuplicateWorldState(t *testing.T) {
	dir := t.TempDir()
	s := store.NewStore(dir)
	if err := s.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if err := s.Progress.Init("test", 10); err != nil {
		t.Fatalf("InitProgress: %v", err)
	}
	if err := s.Drafts.SaveDraft(1, "Văn bản chương 1, Lâm Mặc gặp hắc ảnh và đột phá."); err != nil {
		t.Fatalf("SaveDraft: %v", err)
	}

	timeline := []domain.TimelineEvent{{
		Chapter:    1,
		Time:       "Sáng sớm",
		Event:      "Lâm Mặc gặp hắc ảnh",
		Characters: []string{"林墨"},
	}}
	stateChanges := []domain.StateChange{{
		Chapter:  1,
		Entity:   "林墨",
		Field:    "realm",
		OldValue: "Người phàm",
		NewValue: "Luyện Khí Kỳ",
	}}
	foreshadow := []domain.ForeshadowUpdate{{
		ID:          "f1",
		Action:      "plant",
		Description: "Danh tính hắc ảnh",
	}}

	// Mô phỏng commit_chapter đã ghi lại trạng thái thế giới, nhưng chưa gọi MarkChapterComplete thì tiến trình bị lỗi.
	if err := s.World.AppendTimelineEvents(timeline); err != nil {
		t.Fatalf("AppendTimelineEvents seed: %v", err)
	}
	if err := s.World.AppendStateChanges(stateChanges); err != nil {
		t.Fatalf("AppendStateChanges seed: %v", err)
	}
	if err := s.World.UpdateForeshadow(1, foreshadow); err != nil {
		t.Fatalf("UpdateForeshadow seed: %v", err)
	}
	if err := s.Signals.SavePendingCommit(domain.PendingCommit{
		Chapter: 1,
		Stage:   domain.CommitStageStateApplied,
		Summary: "Tóm tắt gửi một nửa",
	}); err != nil {
		t.Fatalf("SavePendingCommit: %v", err)
	}

	tool := NewCommitChapterTool(s)
	args, _ := json.Marshal(map[string]any{
		"chapter":            1,
		"summary":            "Lâm Mặc gặp hắc ảnh và đột phá",
		"characters":         []string{"林墨"},
		"key_events":         []string{"Gặp hắc ảnh", "Đột phá"},
		"timeline_events":    timeline,
		"state_changes":      stateChanges,
		"foreshadow_updates": foreshadow,
	})
	if _, err := tool.Execute(context.Background(), args); err != nil {
		t.Fatalf("Execute replay: %v", err)
	}

	events, _ := s.World.LoadTimeline()
	if len(events) != 1 {
		t.Fatalf("timeline duplicated after replay, got %d: %+v", len(events), events)
	}
	changes, _ := s.World.LoadStateChanges()
	if len(changes) != 1 {
		t.Fatalf("state changes duplicated after replay, got %d: %+v", len(changes), changes)
	}
	ledger, _ := s.World.LoadForeshadowLedger()
	if len(ledger) != 1 {
		t.Fatalf("foreshadow duplicated after replay, got %d: %+v", len(ledger), ledger)
	}
	pending, _ := s.Signals.LoadPendingCommit()
	if pending != nil {
		t.Fatalf("pending commit should be cleared, got %+v", pending)
	}
	if cp := s.Checkpoints.LatestByStep(domain.ChapterScope(1), "commit"); cp == nil {
		t.Fatal("commit checkpoint should be written")
	}
}

// TestCommitChapterRejectsPolishWithoutDraftChange Xác thực: Sau khi chương đã hoàn thành đi vào hàng đợi trau chuốt/viết lại,
// Nếu người viết bỏ qua draft_chapter để trực tiếp commit (drafts có cùng nội dung với chapters),
// commit_chapter phải bị từ chối, buộc người viết phải gọi draft_chapter để viết phiên bản mới trước.
// TestCommitChapterNonLayeredRecompletesAfterRework Xác thực cuốn sách không phân lớp sau khi hoàn thành trải qua quá trình reopen làm lại,
// Sau khi sửa xong chương, commit, khi hàng đợi được xả (drain), hệ thống tự động quay lại trạng thái hoàn thành (sửa phân nhánh không phân lớp xác định hoàn thành sau drain).
func TestCommitChapterNonLayeredRecompletesAfterRework(t *testing.T) {
	dir := t.TempDir()
	s := store.NewStore(dir)
	if err := s.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if err := s.Progress.Init("test", 2); err != nil {
		t.Fatalf("InitProgress: %v", err)
	}

	// Hai chương đã hoàn thành. Chương 2 được chuẩn bị cả drafts và chapters, để thử làm lại.
	ch2 := "Văn bản gốc chương 2, dùng để mô phỏng bản cuối đã nộp."
	if err := s.Drafts.SaveDraft(2, ch2); err != nil {
		t.Fatalf("SaveDraft: %v", err)
	}
	if err := s.Drafts.SaveFinalChapter(2, ch2); err != nil {
		t.Fatalf("SaveFinalChapter: %v", err)
	}
	if err := s.Progress.MarkChapterComplete(1, 100, "", ""); err != nil {
		t.Fatalf("MarkChapterComplete(1): %v", err)
	}
	if err := s.Progress.MarkChapterComplete(2, len([]rune(ch2)), "", ""); err != nil {
		t.Fatalf("MarkChapterComplete(2): %v", err)
	}
	if err := s.Progress.MarkComplete(); err != nil {
		t.Fatalf("MarkComplete: %v", err)
	}

	// mở lại chương 2 → phase trở về writing, PendingRewrites=[2], flow=rewriting
	if err := s.Progress.Reopen([]int{2}, "làm lại"); err != nil {
		t.Fatalf("Reopen: %v", err)
	}

	// Nộp sau làm lại (bản nháp phải khác bản cuối mới cho qua)
	if err := s.Drafts.SaveDraft(2, ch2+"\n\nĐoạn bổ sung lúc làm lại."); err != nil {
		t.Fatalf("SaveDraft (reworked): %v", err)
	}
	tool := NewCommitChapterTool(s)
	args, _ := json.Marshal(map[string]any{
		"chapter":    2,
		"summary":    "Tóm tắt sau khi làm lại",
		"characters": []string{"Nhân vật chính"},
		"key_events": []string{"Làm sạch"},
	})
	raw, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("Execute rework commit: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if payload["book_complete"] != true {
		t.Errorf("book_complete = %v, want true", payload["book_complete"])
	}

	p, _ := s.Progress.Load()
	if p.Phase != domain.PhaseComplete {
		t.Errorf("phase = %s, want complete (应自动重新收尾)", p.Phase)
	}
	if len(p.PendingRewrites) != 0 {
		t.Errorf("PendingRewrites = %v, want empty", p.PendingRewrites)
	}
}

// TestCommitChapterLayeredReopenRecompletesDespiteOpenThread Xác thực sự kết thúc: Cuốn sách phân lớp đi qua reopen
// Sau khi làm lại, ngay cả khi compass có những tuyến truyện dài chưa kết thúc (quá trình làm lại có thể gây nhiễu), một khi đã xử lý xong hàng đợi, nó vẫn hoàn thành theo nguyên tắc "cấu trúc hoàn chỉnh" ——
// Không bị kẹt ở writing, và chấm dứt vòng lặp vô tận viết tiếp tục ở phần cuối chương cuối cùng (§6.5 / họ known_outline_exhaustion).
// Chứng minh ngược: nếu lộ trình reopen tiếp tục dùng chất lượng mức layeredBookComplete, open thread trong ví dụ này sẽ dẫn đến kết quả trả về là false,
// book_complete cũng thành false và bài kiểm tra thất bại.
func TestCommitChapterLayeredReopenRecompletesDespiteOpenThread(t *testing.T) {
	dir := t.TempDir()
	s := store.NewStore(dir)
	if err := s.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if err := s.Progress.Init("test", 0); err != nil {
		t.Fatalf("InitProgress: %v", err)
	}

	// Một volume, một arc, hai chương, mở rộng tất cả
	foundation := NewSaveFoundationTool(s)
	layeredArgs, _ := json.Marshal(map[string]any{
		"type": "layered_outline",
		"content": []map[string]any{{
			"index": 1, "title": "Quyển 1", "theme": "Chủ đề",
			"arcs": []map[string]any{{
				"index": 1, "title": "Arc 1", "goal": "Mục tiêu",
				"chapters": []map[string]any{
					{"title": "Chương mở đầu", "core_event": "Khởi đầu", "hook": "Tiếp tục"},
					{"title": "Chương tiếp theo", "core_event": "Phát triển", "hook": "Kết thúc"},
				},
			}},
		}},
		"scale": "long",
	})
	if _, err := foundation.Execute(context.Background(), layeredArgs); err != nil {
		t.Fatalf("Execute layered: %v", err)
	}

	// Hai chương đã hoàn thành và kết thúc
	ch2 := "Văn bản gốc chương hai, mô phỏng bản cuối đã nộp."
	for ch, body := range map[int]string{1: "Văn bản chương 1.", 2: ch2} {
		if err := s.Drafts.SaveDraft(ch, body); err != nil {
			t.Fatalf("SaveDraft %d: %v", ch, err)
		}
		if err := s.Drafts.SaveFinalChapter(ch, body); err != nil {
			t.Fatalf("SaveFinalChapter %d: %v", ch, err)
		}
		if err := s.Progress.MarkChapterComplete(ch, len([]rune(body)), "", ""); err != nil {
			t.Fatalf("MarkChapterComplete %d: %v", ch, err)
		}
	}
	if err := s.Progress.MarkComplete(); err != nil {
		t.Fatalf("MarkComplete: %v", err)
	}

	// Mô phỏng "Làm lại đã thay đổi luồng dài": compass vẫn còn có luồng dài chưa đóng (open thread)
	if err := s.Outline.SaveCompass(domain.StoryCompass{EndingDirection: "Nhân vật chính về quê", OpenThreads: []string{"Kẻ thù cũ chưa diệt"}}); err != nil {
		t.Fatalf("SaveCompass: %v", err)
	}

	// mở lại chương 2 → nộp sau làm lại (bản nháp phải khác bản cuối mới cho qua)
	if err := s.Progress.Reopen([]int{2}, "làm lại"); err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	if err := s.Drafts.SaveDraft(2, ch2+"\n\nĐoạn bổ sung lúc làm lại."); err != nil {
		t.Fatalf("SaveDraft reworked: %v", err)
	}
	tool := NewCommitChapterTool(s)
	args, _ := json.Marshal(map[string]any{
		"chapter": 2, "summary": "Tóm tắt làm lại", "characters": []string{"Nhân vật chính"}, "key_events": []string{"Làm sạch"},
	})
	raw, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("Execute rework commit: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if bc, _ := out["book_complete"].(bool); !bc {
		t.Error("Sau khi reopen làm lại và xả hàng đợi, cuốn sách phải kết thúc lại theo cấu trúc hoàn chỉnh (ngay cả khi tuyến truyện dài chưa kết thúc)")
	}
	p, _ := s.Progress.Load()
	if p.Phase != domain.PhaseComplete {
		t.Errorf("phase = %s, want complete", p.Phase)
	}
	if p.ReopenedFromComplete {
		t.Error("Sau khi hoàn thành lại, ReopenedFromComplete phải bị xóa")
	}
}

func TestCommitChapterRejectsPolishWithoutDraftChange(t *testing.T) {
	dir := t.TempDir()
	s := store.NewStore(dir)
	if err := s.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if err := s.Progress.Init("test", 10); err != nil {
		t.Fatalf("InitProgress: %v", err)
	}

	// Mô phỏng chương 2 đã hoàn thành bình thường: drafts và chapters nội dung giống nhau.
	original := "Văn bản gốc của chương hai, dùng để mô phỏng bản cuối cùng đã gửi."
	if err := s.Drafts.SaveDraft(2, original); err != nil {
		t.Fatalf("SaveDraft: %v", err)
	}
	if err := s.Drafts.SaveFinalChapter(2, original); err != nil {
		t.Fatalf("SaveFinalChapter: %v", err)
	}
	if err := s.Progress.MarkChapterComplete(2, len([]rune(original)), "mystery", "quest"); err != nil {
		t.Fatalf("MarkChapterComplete: %v", err)
	}

	// Vào hàng đợi trau chuốt: Flow=Polishing, PendingRewrites=[2]
	if err := s.Progress.SetPendingRewrites([]int{2}, "Kiểm tra trau chuốt"); err != nil {
		t.Fatalf("SetPendingRewrites: %v", err)
	}
	if err := s.Progress.SetFlow(domain.FlowPolishing); err != nil {
		t.Fatalf("SetFlow: %v", err)
	}

	tool := NewCommitChapterTool(s)
	args, _ := json.Marshal(map[string]any{
		"chapter":    2,
		"summary":    "Giả vờ trau chuốt",
		"characters": []string{"Nhân vật chính"},
		"key_events": []string{"Không có thay đổi"},
	})
	_, err := tool.Execute(context.Background(), args)
	if err == nil {
		t.Fatal("expected commit to be rejected when drafts equals final content")
	}

	// Viết thêm một phiên bản thảo khác → Sẽ được chấp nhận
	polished := original + "\n\nĐoạn văn được thêm sau khi trau chuốt."
	if err := s.Drafts.SaveDraft(2, polished); err != nil {
		t.Fatalf("SaveDraft (polished): %v", err)
	}
	if _, err := tool.Execute(context.Background(), args); err != nil {
		t.Fatalf("Execute after real polish: %v", err)
	}
}

// TestCommitChapterLayeredRejectsOutOfRangeChapter Kiểm tra chế độ phân cấp,
// Việc commit một chương có số nằm ngoài layered_outline phải thất bại hẳn, chứ không phải slog.Warn cho phép đi qua.
// Đây là hệ thống phanh vật lý để ngăn chặn "văn sĩ khỏa thân chạy hoài" sau phán đoán sai (case Phàm cốt ch204..347).
func TestCommitChapterLayeredRejectsOutOfRangeChapter(t *testing.T) {
	dir := t.TempDir()
	s := store.NewStore(dir)
	if err := s.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if err := s.Progress.Init("test", 0); err != nil {
		t.Fatalf("InitProgress: %v", err)
	}

	// Tạo một layered_outline, chỉ có 1 cuốn, 1 arc, 1 chương
	foundation := NewSaveFoundationTool(s)
	layeredArgs, _ := json.Marshal(map[string]any{
		"type": "layered_outline",
		"content": []map[string]any{{
			"index": 1, "title": "Quyển 1", "theme": "Chủ đề",
			"arcs": []map[string]any{{
				"index": 1, "title": "Arc 1", "goal": "Mục tiêu",
				"chapters": []map[string]any{
					{"title": "Chương mở đầu", "core_event": "Khởi đầu", "hook": "Tiếp tục"},
				},
			}},
		}},
		"scale": "long",
	})
	if _, err := foundation.Execute(context.Background(), layeredArgs); err != nil {
		t.Fatalf("Execute layered: %v", err)
	}
	_ = s.Progress.UpdatePhase(domain.PhaseWriting)

	// Commit của chương 2 vượt biên phải thất bại hẳn
	if err := s.Drafts.SaveDraft(2, "Văn bản chương vượt biên, phải bị cản lại."); err != nil {
		t.Fatalf("SaveDraft: %v", err)
	}
	tool := NewCommitChapterTool(s)
	args, _ := json.Marshal(map[string]any{
		"chapter":    2,
		"summary":    "Chương vượt biên",
		"characters": []string{"Nhân vật chính"},
		"key_events": []string{"Không được phép"},
	})
	_, err := tool.Execute(context.Background(), args)
	if err == nil {
		t.Fatal("expected commit to fail when chapter out of layered outline range")
	}

	// Tệp chương không nên được lưu, Progress không được tiến hành
	if _, statErr := os.Stat(dir + "/chapters/02.md"); !os.IsNotExist(statErr) {
		t.Fatalf("chapter 2 should not be persisted, stat err=%v", statErr)
	}
	progress, _ := s.Progress.Load()
	if len(progress.CompletedChapters) != 0 {
		t.Fatalf("CompletedChapters should stay empty, got %v", progress.CompletedChapters)
	}
}

// TestCommitChapterLayeredAutoCompletesWhenDone Xác thực sự hoàn thành xác định trong chế độ phân lớp:
// Đề cương đã được triển khai hoàn toàn và viết xong + Không có cung truyện cơ bản + Không cần làm lại + Số lượng điềm báo hoạt động là không + Tuyến truyện la bàn đã kết thúc,
// Commit của chương cuối cùng tự động đẩy Phase=Complete, không phụ thuộc vào việc kiến trúc sư gọi complete_book.
// Đây là sửa lỗi cho livelock được đưa ra sau khi việc tự động hoàn thành phân lớp bị xoá ở 9bf26a5 (ở phần cuối của tập cuối, model không append
// cũng không complete → Writer rơi vào vòng lặp vượt ngoài giới hạn).
func TestCommitChapterLayeredAutoCompletesWhenDone(t *testing.T) {
	dir := t.TempDir()
	s := store.NewStore(dir)
	if err := s.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if err := s.Progress.Init("test", 0); err != nil {
		t.Fatalf("InitProgress: %v", err)
	}

	// 1 tập, 1 cung truyện, 2 chương, triển khai tất cả (không có cung truyện cơ bản)
	foundation := NewSaveFoundationTool(s)
	layeredArgs, _ := json.Marshal(map[string]any{
		"type": "layered_outline",
		"content": []map[string]any{{
			"index": 1, "title": "Quyển 1", "theme": "Chủ đề",
			"arcs": []map[string]any{{
				"index": 1, "title": "Arc 1", "goal": "Mục tiêu",
				"chapters": []map[string]any{
					{"title": "Chương mở đầu", "core_event": "Khởi đầu", "hook": "Tiếp tục"},
					{"title": "Chương tiếp theo", "core_event": "Phát triển", "hook": "Kết thúc"},
				},
			}},
		}},
		"scale": "long",
	})
	if _, err := foundation.Execute(context.Background(), layeredArgs); err != nil {
		t.Fatalf("Execute layered: %v", err)
	}
	// Tuyến truyện la bàn đã kết thúc (OpenThreads rỗng)
	if err := s.Outline.SaveCompass(domain.StoryCompass{EndingDirection: "Nhân vật chính về quê"}); err != nil {
		t.Fatalf("SaveCompass: %v", err)
	}
	_ = s.Progress.UpdatePhase(domain.PhaseWriting)

	tool := NewCommitChapterTool(s)
	commit := func(ch int) map[string]any {
		if err := s.Drafts.SaveDraft(ch, fmt.Sprintf("Nội dung văn bản chương %d, dùng để kiểm tra hoàn thành xác định.", ch)); err != nil {
			t.Fatalf("SaveDraft %d: %v", ch, err)
		}
		args, _ := json.Marshal(map[string]any{
			"chapter": ch, "summary": "Tóm tắt", "characters": []string{"Nhân vật chính"}, "key_events": []string{"Sự kiện"},
		})
		raw, err := tool.Execute(context.Background(), args)
		if err != nil {
			t.Fatalf("Execute ch%d: %v", ch, err)
		}
		var out map[string]any
		if err := json.Unmarshal(raw, &out); err != nil {
			t.Fatalf("Unmarshal ch%d: %v", ch, err)
		}
		return out
	}

	// Chương 1: Chưa viết xong, không được hoàn thành
	if bc, _ := commit(1)["book_complete"].(bool); bc {
		t.Fatal("Viết xong chương 1 không nên kích hoạt hoàn thành")
	}
	if p, _ := s.Progress.Load(); p.Phase == domain.PhaseComplete {
		t.Fatal("Viết xong chương 1 phase không nên là complete")
	}

	// Chương 2 (chương cuối cùng): Nên tự động hoàn thành
	if bc, _ := commit(2)["book_complete"].(bool); !bc {
		t.Fatal("Viết xong chương cuối cùng nên tự động hoàn thành")
	}
	if p, _ := s.Progress.Load(); p.Phase != domain.PhaseComplete {
		t.Fatalf("expected phase=complete, got %s", p.Phase)
	}
}

// TestCommitChapterLayeredNoAutoCompleteWithOpenThreads Kiểm tra tính bảo thủ: Vẫn còn có các tuyến truyện dài hoạt động
// Ngay cả khi toàn bộ chương đã viết, cũng sẽ không tự động hoàn thành, nhường quyền quyết định "tiếp tục hay không" cho kiến trúc sư.
func TestCommitChapterLayeredNoAutoCompleteWithOpenThreads(t *testing.T) {
	dir := t.TempDir()
	s := store.NewStore(dir)
	if err := s.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if err := s.Progress.Init("test", 0); err != nil {
		t.Fatalf("InitProgress: %v", err)
	}

	foundation := NewSaveFoundationTool(s)
	layeredArgs, _ := json.Marshal(map[string]any{
		"type": "layered_outline",
		"content": []map[string]any{{
			"index": 1, "title": "Quyển 1", "theme": "Chủ đề",
			"arcs": []map[string]any{{
				"index": 1, "title": "Arc 1", "goal": "Mục tiêu",
				"chapters": []map[string]any{{"title": "Chương mở đầu", "core_event": "Khởi đầu", "hook": "Tiếp tục"}},
			}},
		}},
		"scale": "long",
	})
	if _, err := foundation.Execute(context.Background(), layeredArgs); err != nil {
		t.Fatalf("Execute layered: %v", err)
	}
	// Vẫn còn các tuyến truyện dài chưa kết thúc
	if err := s.Outline.SaveCompass(domain.StoryCompass{EndingDirection: "Nhân vật chính về quê", OpenThreads: []string{"Kẻ thù cũ chưa diệt"}}); err != nil {
		t.Fatalf("SaveCompass: %v", err)
	}
	_ = s.Progress.UpdatePhase(domain.PhaseWriting)

	if err := s.Drafts.SaveDraft(1, "Văn bản của chương duy nhất, nhưng tuyến truyện dài vẫn chưa kết thúc."); err != nil {
		t.Fatalf("SaveDraft: %v", err)
	}
	tool := NewCommitChapterTool(s)
	args, _ := json.Marshal(map[string]any{
		"chapter": 1, "summary": "Tóm tắt", "characters": []string{"Nhân vật chính"}, "key_events": []string{"Sự kiện"},
	})
	if _, err := tool.Execute(context.Background(), args); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if p, _ := s.Progress.Load(); p.Phase == domain.PhaseComplete {
		t.Fatal("Khi tuyến truyện dài đang hoạt động chưa kết thúc, không nên tự động hoàn thành")
	}
}
