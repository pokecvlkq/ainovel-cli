package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/voocel/agentcore/schema"
	"github.com/voocel/ainovel-cli/internal/domain"
	"github.com/voocel/ainovel-cli/internal/errs"
	"github.com/voocel/ainovel-cli/internal/rules"
	"github.com/voocel/ainovel-cli/internal/store"
)

// CommitChapterTool Gửi chương: Tải văn bản → Lưu bản cuối cùng → Tạo tóm tắt → Cập nhật trạng thái → Cập nhật tiến độ.
type CommitChapterTool struct {
	store *store.Store
}

func NewCommitChapterTool(store *store.Store) *CommitChapterTool {
	return &CommitChapterTool{store: store}
}

// commitOutput nhúng các trường mở rộng lên trên domain.CommitResult để giữ gói domain không phụ thuộc vào rules.
// Vì các trường nhúng sẽ được JSON marshaler nâng cấp (promoted), kết quả tuần tự hoá tương đương với cấu trúc phẳng.
type commitOutput struct {
	domain.CommitResult
	RuleViolations []rules.Violation `json:"rule_violations,omitempty"`
}

func (t *CommitChapterTool) Name() string { return "commit_chapter" }
func (t *CommitChapterTool) Description() string {
	return "Gửi bản cuối của chương. Tải bản nháp lưu thành bản cuối, cập nhật dòng thời gian, foreshadowing, mối quan hệ, trạng thái nhân vật và tiến độ." +
		"Trả về các thông tin có cấu trúc: next_chapter / review_required / arc_end / volume_end / needs_expansion / book_complete / flow v.v."
}
func (t *CommitChapterTool) Label() string { return "Gửi chương" }

// Công cụ ghi (thao tác nguyên tử xuyên miền: bản nháp → bản cuối → tóm tắt → tiến độ → checkpoint), cấm đồng thời.
func (t *CommitChapterTool) ReadOnly(_ json.RawMessage) bool        { return false }
func (t *CommitChapterTool) ConcurrencySafe(_ json.RawMessage) bool { return false }

func (t *CommitChapterTool) Schema() map[string]any {
	timelineSchema := schema.Object(
		schema.Property("time", schema.String("Thời gian trong truyện")).Required(),
		schema.Property("event", schema.String("Mô tả sự kiện")).Required(),
		schema.Property("characters", schema.Array("Nhân vật liên quan", schema.String(""))),
	)
	foreshadowSchema := schema.Object(
		schema.Property("id", schema.String("ID foreshadowing")).Required(),
		schema.Property("action", schema.Enum("Thao tác", "plant", "advance", "resolve")).Required(),
		schema.Property("description", schema.String("Mô tả foreshadowing (chỉ bắt buộc khi plant)")),
	)
	relationshipSchema := schema.Object(
		schema.Property("character_a", schema.String("Nhân vật A")).Required(),
		schema.Property("character_b", schema.String("Nhân vật B")).Required(),
		schema.Property("relation", schema.String("Mô tả mối quan hệ hiện tại")).Required(),
	)
	stateChangeSchema := schema.Object(
		schema.Property("entity", schema.String("Tên nhân vật hoặc tên thực thể")).Required(),
		schema.Property("field", schema.String("Thuộc tính thay đổi")).Required(),
		schema.Property("old_value", schema.String("Giá trị trước khi thay đổi")),
		schema.Property("new_value", schema.String("Giá trị sau khi thay đổi")).Required(),
		schema.Property("reason", schema.String("Lý do thay đổi")),
	)
	feedbackSchema := schema.Object(
		schema.Property("deviation", schema.String("Mô tả sự sai lệch so với đề cương")).Required(),
		schema.Property("suggestion", schema.String("Đề xuất điều chỉnh cho đề cương tiếp theo")).Required(),
	)
	feedbackSchema["description"] = "Đối tượng gợi ý cho đề cương tiếp theo; phải truyền trực tiếp JSON object, đừng truyền JSON đã string hóa"
	return schema.Object(
		schema.Property("chapter", schema.Int("Số chương")).Required(),
		schema.Property("summary", schema.String("Tóm tắt nội dung chương này (dưới 200 chữ)")).Required(),
		schema.Property("characters", schema.Array("Tên các nhân vật xuất hiện trong chương này", schema.String(""))).Required(),
		schema.Property("key_events", schema.Array("Các sự kiện chính trong chương này", schema.String(""))).Required(),
		schema.Property("timeline_events", schema.Array("Các sự kiện dòng thời gian trong chương này", timelineSchema)),
		schema.Property("foreshadow_updates", schema.Array("Thao tác foreshadowing", foreshadowSchema)),
		schema.Property("relationship_changes", schema.Array("Thay đổi mối quan hệ", relationshipSchema)),
		schema.Property("state_changes", schema.Array("Thay đổi trạng thái nhân vật/thực thể", stateChangeSchema)),
		schema.Property("cast_intros", schema.Array("Giới thiệu nhân vật phụ xuất hiện lần đầu trong chương này và có thể xuất hiện lại sau này (không bao gồm nhân vật chính và các nhân vật đã có trong characters.json)", schema.Object(
			schema.Property("name", schema.String("Tên nhân vật")).Required(),
			schema.Property("brief_role", schema.String("Định vị bằng một câu (ví dụ: ông chủ nhà trọ/bảo kê sòng bạc)")).Required(),
		))),
		schema.Property("hook_type", schema.Enum("Loại hook cuối chương", "crisis", "mystery", "desire", "emotion", "choice")),
		schema.Property("dominant_strand", schema.Enum("Tuyến tự sự chính trong chương", "quest", "fire", "constellation")),
		schema.Property("feedback", feedbackSchema),
	)
}

func (t *CommitChapterTool) Execute(_ context.Context, args json.RawMessage) (json.RawMessage, error) {
	var a struct {
		Chapter             int                        `json:"chapter"`
		Summary             string                     `json:"summary"`
		Characters          []string                   `json:"characters"`
		KeyEvents           []string                   `json:"key_events"`
		TimelineEvents      []domain.TimelineEvent     `json:"timeline_events"`
		ForeshadowUpdates   []domain.ForeshadowUpdate  `json:"foreshadow_updates"`
		RelationshipChanges []domain.RelationshipEntry `json:"relationship_changes"`
		StateChanges        []domain.StateChange       `json:"state_changes"`
		CastIntros          []domain.CastIntro         `json:"cast_intros"`
		HookType            string                     `json:"hook_type"`
		DominantStrand      string                     `json:"dominant_strand"`
		Feedback            *domain.OutlineFeedback    `json:"feedback"`
	}
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, fmt.Errorf("invalid args: %w: %w", errs.ErrToolArgs, err)
	}
	if a.Chapter <= 0 {
		return nil, fmt.Errorf("chapter must be > 0: %w", errs.ErrToolArgs)
	}
	if t.store.Progress.IsChapterCompleted(a.Chapter) {
		// Xoá các PendingCommit có thể còn sót lại (sự cố xảy ra sau ProgressMarked, trước ClearPendingCommit)
		if pending, _ := t.store.Signals.LoadPendingCommit(); pending != nil && pending.Chapter == a.Chapter {
			if err := t.appendCommitCheckpoint(a.Chapter); err != nil {
				return nil, fmt.Errorf("checkpoint commit: %w: %w", errs.ErrStoreWrite, err)
			}
			_ = t.store.Signals.ClearPendingCommit()
		}
		// Đường dẫn trau chuốt/viết lại: mặc dù chương đã hoàn thành nhưng vẫn ở trong pending_rewrites, cho phép ghi đè và xả (drain) hàng đợi
		progress, _ := t.store.Progress.Load()
		if progress != nil && slices.Contains(progress.PendingRewrites, a.Chapter) {
			return t.executeRewriteCommit(a.Chapter, a.Summary, a.Characters, a.KeyEvents,
				a.HookType, a.DominantStrand, progress)
		}
		return t.buildSkipResult(a.Chapter, progress)
	}
	existingPending, err := t.store.Signals.LoadPendingCommit()
	if err != nil {
		return nil, fmt.Errorf("load pending commit: %w: %w", errs.ErrStoreRead, err)
	}
	if existingPending != nil && existingPending.Chapter != a.Chapter {
		return nil, fmt.Errorf("Có commit chương chưa được khôi phục: chương %d (giai đoạn %s), vui lòng khôi phục hoặc gửi lại chương này trước: %w", existingPending.Chapter, existingPending.Stage, errs.ErrToolConflict)
	}
	if err := t.store.Progress.ValidateChapterWork(a.Chapter); err != nil {
		// Xung đột hàng đợi giữ nguyên (đã có phân loại ErrToolConflict); các lỗi IO khác được phân vào Precondition.
		if errors.Is(err, errs.ErrToolConflict) {
			return nil, err
		}
		return nil, fmt.Errorf("Chương hiện không được phép gửi: %w: %w", errs.ErrToolPrecondition, err)
	}

	// Chặn vượt quá giới hạn trong chế độ phân lớp: Phải thực hiện trước bất kỳ thao tác ghi nào, nếu không commit vượt giới hạn sẽ làm hỏng tệp chương, tóm tắt,
	// Tiến độ đều bị làm hỏng. boundary được tái sử dụng cho bước 6b bên dưới để tính tín hiệu arc/volume.
	var boundary *store.ArcBoundary
	if progress, perr := t.store.Progress.Load(); perr == nil && progress != nil && progress.Layered {
		b, bErr := t.store.Outline.CheckArcBoundary(a.Chapter)
		if bErr != nil {
			return nil, fmt.Errorf("Phát hiện biên arc thất bại chapter=%d: %w: %w", a.Chapter, errs.ErrStoreRead, bErr)
		}
		if b == nil {
			return nil, fmt.Errorf(
				"Chương %d không nằm trong phạm vi đề cương phân lớp: việc viết phải gọi expand_arc (mở rộng arc) hoặc append_volume (thêm volume) trước; nếu toàn bộ cuốn sách đã hoàn thành, vui lòng gọi save_foundation type=complete_book: %w",
				a.Chapter, errs.ErrToolPrecondition)
		}
		boundary = b
	}

	// 1. Tải nội dung văn bản của chương
	content, wordCount, err := t.store.Drafts.LoadChapterContent(a.Chapter)
	if err != nil {
		return nil, fmt.Errorf("load chapter content: %w: %w", errs.ErrStoreRead, err)
	}
	if content == "" {
		return nil, fmt.Errorf("no content found for chapter %d: %w", a.Chapter, errs.ErrToolPrecondition)
	}

	now := time.Now().Format(time.RFC3339)
	pending := domain.PendingCommit{
		Chapter:        a.Chapter,
		Stage:          domain.CommitStageStarted,
		Summary:        a.Summary,
		HookType:       a.HookType,
		DominantStrand: a.DominantStrand,
		StartedAt:      now,
		UpdatedAt:      now,
	}
	if err := t.store.Signals.SavePendingCommit(pending); err != nil {
		return nil, fmt.Errorf("save pending commit: %w: %w", errs.ErrStoreWrite, err)
	}

	// 2. Lưu bản cuối cùng
	if err := t.store.Drafts.SaveFinalChapter(a.Chapter, content); err != nil {
		return nil, fmt.Errorf("save final chapter: %w: %w", errs.ErrStoreWrite, err)
	}

	// 3. Lưu tóm tắt
	summary := domain.ChapterSummary{
		Chapter:    a.Chapter,
		Summary:    a.Summary,
		Characters: a.Characters,
		KeyEvents:  a.KeyEvents,
	}
	if err := t.store.Summaries.SaveSummary(summary); err != nil {
		return nil, fmt.Errorf("save summary: %w: %w", errs.ErrStoreWrite, err)
	}

	// 4. Cập nhật gia số trạng thái
	if len(a.TimelineEvents) > 0 {
		for i := range a.TimelineEvents {
			a.TimelineEvents[i].Chapter = a.Chapter
		}
		if err := t.store.World.AppendTimelineEvents(a.TimelineEvents); err != nil {
			return nil, fmt.Errorf("append timeline: %w: %w", errs.ErrStoreWrite, err)
		}
	}
	if len(a.ForeshadowUpdates) > 0 {
		if err := t.store.World.UpdateForeshadow(a.Chapter, a.ForeshadowUpdates); err != nil {
			return nil, fmt.Errorf("update foreshadow: %w: %w", errs.ErrStoreWrite, err)
		}
	}
	if len(a.RelationshipChanges) > 0 {
		for i := range a.RelationshipChanges {
			a.RelationshipChanges[i].Chapter = a.Chapter
		}
		if err := t.store.World.UpdateRelationships(a.RelationshipChanges); err != nil {
			return nil, fmt.Errorf("update relationships: %w: %w", errs.ErrStoreWrite, err)
		}
	}
	if len(a.StateChanges) > 0 {
		for i := range a.StateChanges {
			a.StateChanges[i].Chapter = a.Chapter
		}
		if err := t.store.World.AppendStateChanges(a.StateChanges); err != nil {
			return nil, fmt.Errorf("append state changes: %w: %w", errs.ErrStoreWrite, err)
		}
	}

	// 4b. Tích luỹ danh sách nhân vật phụ: các nhân vật không cốt lõi xuất hiện trong chương này sẽ vào cast_ledger để novel_context có thể gọi lại.
	// Khi thất bại chỉ cảnh báo (warn) chứ không chặn commit —— danh sách nhân vật là dữ liệu thứ yếu, có thể tự phục hồi thông qua commit ở chương tiếp theo.
	if len(a.Characters) > 0 {
		coreNames := loadCoreCharacterNameSet(t.store)
		if err := t.store.Cast.MergeAppearances(a.Chapter, a.Characters, a.CastIntros, coreNames); err != nil {
			slog.Warn("Tích luỹ danh sách nhân vật phụ thất bại, bỏ qua", "module", "commit", "chapter", a.Chapter, "err", err)
		}
	}

	pending.Stage = domain.CommitStageStateApplied
	pending.UpdatedAt = time.Now().Format(time.RFC3339)
	if err := t.store.Signals.SavePendingCommit(pending); err != nil {
		return nil, fmt.Errorf("update pending commit stage: %w: %w", errs.ErrStoreWrite, err)
	}

	// 5. Cập nhật tiến độ
	if err := t.store.Progress.MarkChapterComplete(a.Chapter, wordCount, a.HookType, a.DominantStrand); err != nil {
		return nil, fmt.Errorf("mark chapter complete: %w: %w", errs.ErrStoreWrite, err)
	}

	// 6. Đánh giá xem có cần xét duyệt hay không
	progress, err := t.store.Progress.Load()
	if err != nil {
		return nil, fmt.Errorf("load progress: %w: %w", errs.ErrStoreRead, err)
	}
	completedCount := 0
	if progress != nil {
		completedCount = len(progress.CompletedChapters)
	}

	// 6b. Tín hiệu arc/volume của chế độ tiểu thuyết dài: boundary đã được xác thực trước ở lối vào, đảm bảo không nil khi Layered
	var arcEnd, volumeEnd, needsExpansion, needsNewVolume bool
	var vol, arc, nextVol, nextArc int
	if progress != nil && progress.Layered && boundary != nil {
		arcEnd = boundary.IsArcEnd
		volumeEnd = boundary.IsVolumeEnd
		vol = boundary.Volume
		arc = boundary.Arc
		needsExpansion = boundary.NeedsExpansion
		needsNewVolume = boundary.NeedsNewVolume
		nextVol = boundary.NextVolume
		nextArc = boundary.NextArc
		_ = t.store.Progress.UpdateVolumeArc(vol, arc)
	}

	var reviewRequired bool
	var reviewReason string
	if progress != nil && progress.Layered {
		reviewRequired, reviewReason = domain.ShouldArcReview(arcEnd, volumeEnd, vol, arc)
	} else {
		reviewRequired, reviewReason = domain.ShouldReview(completedCount)
	}

	// 7. Cấu trúc tín hiệu được xây dựng
	result := domain.CommitResult{
		Chapter:        a.Chapter,
		Committed:      true,
		WordCount:      wordCount,
		NextChapter:    a.Chapter + 1,
		ReviewRequired: reviewRequired,
		ReviewReason:   reviewReason,
		HookType:       a.HookType,
		DominantStrand: a.DominantStrand,
		Feedback:       a.Feedback,
		ArcEnd:         arcEnd,
		VolumeEnd:      volumeEnd,
		Volume:         vol,
		Arc:            arc,
		NeedsExpansion: needsExpansion,
		NeedsNewVolume: needsNewVolume,
		NextVolume:     nextVol,
		NextArc:        nextArc,
	}

	// 8. Đánh giá trạng thái hoàn thành: Viết xong chương cuối cùng ở chế độ không phân lớp / chương cuối của tập cuối cùng ở chế độ phân lớp → MarkComplete
	if t.applyCompletion(&result, progress) {
		result.BookComplete = true
	}
	if p, _ := t.store.Progress.Load(); p != nil {
		result.Flow = string(p.Flow)
	}

	pending.Stage = domain.CommitStageProgressMarked
	pending.Result = &result
	pending.UpdatedAt = time.Now().Format(time.RFC3339)
	if err := t.store.Signals.SavePendingCommit(pending); err != nil {
		return nil, fmt.Errorf("update pending commit result: %w: %w", errs.ErrStoreWrite, err)
	}

	// 9. Bổ sung checkpoint. Phải thực hiện trước khi xoá pending_commit, đảm bảo
	// pending_commit có thể nhìn thấy sau khi khởi động lại luôn có thể điều khiển chạy lại để bù đắp checkpoint bị thiếu.
	if err := t.appendCommitCheckpoint(a.Chapter); err != nil {
		return nil, fmt.Errorf("checkpoint commit: %w: %w", errs.ErrStoreWrite, err)
	}

	// 10. Xoá trạng thái trung gian của tiến độ
	if err := t.store.Progress.ClearInProgress(); err != nil {
		return nil, fmt.Errorf("clear in-progress: %w: %w", errs.ErrStoreWrite, err)
	}
	if err := t.store.Signals.ClearPendingCommit(); err != nil {
		return nil, fmt.Errorf("clear pending commit: %w: %w", errs.ErrStoreWrite, err)
	}

	// 11. Kiểm tra quy tắc cơ học (chỉ trả về sự thật, không chặn)
	violations := t.checkRules(content, wordCount)
	return json.Marshal(commitOutput{CommitResult: result, RuleViolations: violations})
}

func (t *CommitChapterTool) appendCommitCheckpoint(chapter int) error {
	_, err := t.store.Checkpoints.AppendArtifact(
		domain.ChapterScope(chapter), "commit",
		fmt.Sprintf("chapters/%02d.md", chapter),
	)
	return err
}

// checkRules thực hiện kiểm tra cơ học trên nội dung chương: Lint đường cơ sở của sản phẩm tích hợp (cơ chế còn sót lại, luôn thực thi)
// + User rules Check (đọc snapshot structured của sách; nếu snapshot bị thiếu, lùi về mặc định tích hợp sẵn để đảm bảo đường cơ sở cơ học luôn tồn tại).
func (t *CommitChapterTool) checkRules(text string, wordCount int) []rules.Violation {
	violations := rules.Lint(text)
	structured := rules.SystemDefaults().Structured
	if snap, err := t.store.UserRules.Load(); err == nil && snap != nil {
		structured = snap.Structured
	}
	return append(violations, rules.Check(text, wordCount, structured)...)
}

// executeRewriteCommit xử lý việc gửi các chương trau chuốt/viết lại: ghi đè bản cuối cùng và bản tóm tắt, cập nhật số từ, xả (drain) hàng đợi.
// Bỏ qua tất cả các thao tác bổ sung trạng thái thế giới (timeline / foreshadow / relationship / state_changes) và kiểm tra ranh giới arc,
// những thứ này đã được áp dụng trong lần gửi ban đầu của chương.
func (t *CommitChapterTool) executeRewriteCommit(
	chapter int,
	summary string,
	characters, keyEvents []string,
	hookType, dominantStrand string,
	progress *domain.Progress,
) (json.RawMessage, error) {
	// 1. Tải văn bản đã được trau chuốt
	content, wordCount, err := t.store.Drafts.LoadChapterContent(chapter)
	if err != nil {
		return nil, fmt.Errorf("rewrite: load chapter content: %w: %w", errs.ErrStoreRead, err)
	}
	if content == "" {
		return nil, fmt.Errorf("no content found for chapter %d: %w", chapter, errs.ErrToolPrecondition)
	}

	// 2. Xác thực cứng: drafts hoàn toàn giống với bản cuối cùng hiện tại → xác định là chưa thực sự trau chuốt/viết lại (writer đã bỏ qua draft_chapter)
	// Từ chối commit, buộc người viết gọi draft_chapter(mode=write) trước để viết một phiên bản mới.
	existingFinal, _ := t.store.Drafts.LoadChapterText(chapter)
	if existingFinal != "" && existingFinal == content {
		mode := "viết lại"
		if progress != nil && progress.Flow == domain.FlowPolishing {
			mode = "trau chuốt"
		}
		return nil, fmt.Errorf("Chương %d drafts có nội dung hoàn toàn giống với chapters, chưa phát hiện thay đổi %s. Vui lòng gọi draft_chapter(mode=write, chapter=%d) trước để ghi văn bản mới sau khi %s, sau đó gọi commit_chapter: %w",
			chapter, mode, chapter, mode, errs.ErrToolPrecondition)
	}

	// 3. Ghi đè bản cuối cùng
	if err := t.store.Drafts.SaveFinalChapter(chapter, content); err != nil {
		return nil, fmt.Errorf("rewrite: save final chapter: %w: %w", errs.ErrStoreWrite, err)
	}

	// 3. Ghi đè bản tóm tắt
	if err := t.store.Summaries.SaveSummary(domain.ChapterSummary{
		Chapter:    chapter,
		Summary:    summary,
		Characters: characters,
		KeyEvents:  keyEvents,
	}); err != nil {
		return nil, fmt.Errorf("rewrite: save summary: %w: %w", errs.ErrStoreWrite, err)
	}

	// 4. Cập nhật số từ (MarkChapterComplete là idempotic đối với các chương đã hoàn thành: thay thế word count, slice.Contains ngăn cản việc vào hàng đợi nhiều lần)
	if err := t.store.Progress.MarkChapterComplete(chapter, wordCount, hookType, dominantStrand); err != nil {
		return nil, fmt.Errorf("rewrite: update word count: %w: %w", errs.ErrStoreWrite, err)
	}

	// 5. Xả (Drain) hàng đợi đang chờ xử lý; khi hàng đợi rỗng, CompleteRewrite sẽ tự động chuyển dòng (flow) trở lại quá trình viết (writing)
	if err := t.store.Progress.CompleteRewrite(chapter); err != nil {
		return nil, fmt.Errorf("rewrite: complete rewrite: %w: %w", errs.ErrStoreWrite, err)
	}

	// 6. Checkpoint
	if _, err := t.store.Checkpoints.AppendArtifact(
		domain.ChapterScope(chapter), "commit",
		fmt.Sprintf("chapters/%02d.md", chapter),
	); err != nil {
		return nil, fmt.Errorf("rewrite: checkpoint commit: %w: %w", errs.ErrStoreWrite, err)
	}

	// 7. Đọc snapshot Progress sau khi drain, làm thông tin thực tế trả về
	mode := "rewrite"
	if progress.Flow == domain.FlowPolishing {
		mode = "polish"
	}
	latest, _ := t.store.Progress.Load()
	remaining := []int{}
	nextChapter := chapter + 1
	flow := string(domain.FlowWriting)
	if latest != nil {
		remaining = append(remaining, latest.PendingRewrites...)
		nextChapter = latest.NextChapter()
		flow = string(latest.Flow)
	}
	drained := len(remaining) == 0

	// Sau khi hàng đợi rỗng, đánh giá trạng thái hoàn thành một lần nữa: commit của việc làm lại (rework) không đi qua đường dẫn chính applyCompletion, việc hoàn thành chỉ có thể được kích hoạt ở đây.
	//   - Phân lớp + viết tiến lên: sử dụng mức chất lượng layeredBookComplete (yêu cầu thu gọn manh mối), nếu không đáp ứng thì chuyển sang cho architect.
	//   - Phân lớp + reopen làm lại (ReopenedFromComplete): làm lại chỉ thay đổi các chương hiện có, không thêm bớt cấu trúc, theo tính toàn vẹn cấu trúc
	//     tức là hoàn thành lại - nếu do làm lại làm xáo trộn một luồng nào đó khiến quá trình bị kẹt ở phần viết (writing), cuối cùng sẽ rơi vào vòng lặp vô tận của việc tiếp tục viết vượt ranh giới.
	//   - Không phân lớp: Viết đủ TotalChapters tức là hoàn thành (làm lại không tăng hay giảm số chương, ban đầu đã đầy).
	bookComplete := false
	if drained && latest != nil {
		reComplete := false
		switch {
		case latest.Layered && latest.ReopenedFromComplete:
			reComplete = t.layeredStructurallyComplete(latest)
		case latest.Layered:
			reComplete = t.layeredBookComplete(latest)
		default:
			reComplete = latest.TotalChapters > 0 && len(latest.CompletedChapters) >= latest.TotalChapters
		}
		if reComplete {
			if cerr := t.store.Progress.MarkComplete(); cerr == nil {
				bookComplete = true
				if p, _ := t.store.Progress.Load(); p != nil {
					flow = string(p.Flow)
				}
			}
		}
	}

	// Cùng một lộ trình chính: rewrite/polish cũng thực hiện kiểm tra cơ học và đính kèm rule_violations
	violations := t.checkRules(content, wordCount)
	return json.Marshal(map[string]any{
		"chapter":         chapter,
		"rewritten":       true,
		"mode":            mode,
		"word_count":      wordCount,
		"remaining_queue": remaining,
		"queue_drained":   drained,
		"next_chapter":    nextChapter,
		"flow":            flow,
		"book_complete":   bookComplete,
		"rule_violations": violations,
	})
}

// buildSkipResult xây dựng dữ liệu trả về tương ứng với một commit bình thường đối với trường hợp "gửi lặp lại chương đã hoàn thành".
// Điều phối viên căn cứ vào đó để đưa ra quyết định tiếp theo (gửi cho người viết/biên tập/kiến trúc sư), chứ không phải bị ảo giác vì nhận được gợi ý nội dung chữ.
func (t *CommitChapterTool) buildSkipResult(chapter int, progress *domain.Progress) (json.RawMessage, error) {
	_, wordCount, _ := t.store.Drafts.LoadChapterContent(chapter)

	result := domain.CommitResult{
		Chapter:     chapter,
		Committed:   true,
		WordCount:   wordCount,
		NextChapter: chapter + 1,
	}

	if progress != nil && progress.Layered {
		if boundary, _ := t.store.Outline.CheckArcBoundary(chapter); boundary != nil {
			result.ArcEnd = boundary.IsArcEnd
			result.VolumeEnd = boundary.IsVolumeEnd
			result.Volume = boundary.Volume
			result.Arc = boundary.Arc
			result.NeedsExpansion = boundary.NeedsExpansion
			result.NeedsNewVolume = boundary.NeedsNewVolume
			result.NextVolume = boundary.NextVolume
			result.NextArc = boundary.NextArc
		}
		result.ReviewRequired, result.ReviewReason = domain.ShouldArcReview(result.ArcEnd, result.VolumeEnd, result.Volume, result.Arc)
	} else if progress != nil {
		result.ReviewRequired, result.ReviewReason = domain.ShouldReview(len(progress.CompletedChapters))
	}

	if progress != nil {
		if progress.Phase == domain.PhaseComplete {
			result.BookComplete = true
		}
		result.Flow = string(progress.Flow)
	}

	return json.Marshal(result)
}

// loadCoreCharacterNameSet tải tập hợp tên nhân vật đã có trong characters.json (bao gồm cả bí danh).
// Được sử dụng làm tập lọc "cốt lõi đã biết" cho cast_ledger —— nhân vật cốt lõi không được đưa vào danh sách nhân vật phụ.
// Trả về nil khi quá trình tải thất bại (trong lúc merge, mọi nhân vật đều được đưa vào ledger, có thể chấp nhận được).
func loadCoreCharacterNameSet(s *store.Store) map[string]bool {
	chars, err := s.Characters.Load()
	if err != nil || len(chars) == 0 {
		return nil
	}
	set := make(map[string]bool, len(chars)*2)
	for _, c := range chars {
		if c.Name != "" {
			set[c.Name] = true
		}
		for _, alias := range c.Aliases {
			if alias != "" {
				set[alias] = true
			}
		}
	}
	return set
}

// applyCompletion xác định xem commit này có kết thúc toàn bộ cuốn sách hay không, nếu có thì MarkComplete và trả về true.
//   - Không phân lớp: Cuốn sách kết thúc khi hoàn thành số lượng chương đã thỏa thuận.
//   - Phân lớp: Kiến trúc sư gọi tường minh save_foundation type=complete_book là luồng chính; ở đây thêm một lớp bảo vệ:
//     kiểm tra xác định —— khi toàn bộ sách đạt đủ điều kiện hoàn thành một cách khách quan (xem layeredBookComplete) thì tự động kết thúc.
//     Ngăn chặn trường hợp model đến điểm cuối mà không append_volume hay complete_book, dẫn đến "người viết cứ chạy ra ngoài giới hạn chương →
//     Vệ binh giới hạn chặn → Thử lại liên tục" tạo thành livelock (Nguyên nhân cốt lõi trong case《Phàm Cốt》ch204..347).
func (t *CommitChapterTool) applyCompletion(result *domain.CommitResult, progress *domain.Progress) bool {
	if progress == nil {
		return false
	}
	if progress.Layered {
		if t.layeredBookComplete(progress) {
			_ = t.store.Progress.MarkComplete()
			return true
		}
		return false
	}
	if progress.TotalChapters > 0 && result.NextChapter > progress.TotalChapters {
		_ = t.store.Progress.MarkComplete()
		return true
	}
	return false
}

// layeredStructurallyComplete đánh giá xem tiểu thuyết dài phân lớp đã "được viết xong về mặt cấu trúc" hay chưa: hàng đợi rework trống + không có arc khung cần triển khai
// + Tất cả các chương đã mở rộng đều được viết xong. Đây là sự thật rõ ràng của trạng thái kết thúc, không mang các phán đoán ngữ nghĩa như foreshadowing/tuyến truyện dài —— dùng làm "bảo vệ trạng thái kết thúc
// Vòng lặp vô tận" lưới an toàn (sau khi làm rỗng rework, việc kết thúc lại dựa vào điều này).
func (t *CommitChapterTool) layeredStructurallyComplete(progress *domain.Progress) bool {
	// 1. Phải làm trống hàng đợi rework
	if len(progress.PendingRewrites) > 0 {
		return false
	}
	volumes, err := t.store.Outline.LoadLayeredOutline()
	if err != nil || len(volumes) == 0 {
		return false
	}
	// 2. Không thể có cung khung chờ mở rộng (vẫn còn nội dung được lên kế hoạch viết)
	for i := range volumes {
		for j := range volumes[i].Arcs {
			if !volumes[i].Arcs[j].IsExpanded() {
				return false
			}
		}
	}
	// 3. Phải viết tất cả các chương đã được mở rộng
	expanded := len(domain.FlattenOutline(volumes))
	return expanded > 0 && len(progress.CompletedChapters) >= expanded
}

// layeredBookComplete sử dụng thông tin khách quan để xác định xem tiểu thuyết phân lớp đã được viết xong hoàn toàn hay chưa, đối chiếu với tiêu chí đánh giá hoàn thành architect-long.md
// Vài mục có thể định lượng trong danh sách + các sự kiện có tính cấu trúc. Ngoài ra, tính hoàn thiện cấu trúc đòi hỏi không còn foreshadowing chưa mở và các luồng truyện dài đều kết thúc —— nếu thiếu bất kỳ điều kiện nào
// sẽ chuyển lại cho architect tiếp tục expand_arc / append_volume, và tuyệt đối không kết thúc khi câu chuyện chưa được viết xong. Cực kỳ bảo thủ khi không có compass
// sẽ được đánh giá là chưa hoàn thành. Đây là đánh giá hoàn thành "chất lượng" của việc viết tịnh tiến, nghiêm ngặt hơn layeredStructurallyComplete.
func (t *CommitChapterTool) layeredBookComplete(progress *domain.Progress) bool {
	if !t.layeredStructurallyComplete(progress) {
		return false
	}
	// 4. Foreshadowing đang hoạt động phải về không (lời hứa đã được thực hiện)
	if active, aerr := t.store.World.LoadActiveForeshadow(); aerr != nil || len(active) > 0 {
		return false
	}
	// 5. Các tuyến truyện dài theo la bàn phải kết thúc (nếu không có compass / tuyến truyện dài chưa kết thúc, giao lại cho architect quyết định)
	compass, cerr := t.store.Outline.LoadCompass()
	if cerr != nil || compass == nil || len(compass.OpenThreads) > 0 {
		return false
	}
	return true
}
