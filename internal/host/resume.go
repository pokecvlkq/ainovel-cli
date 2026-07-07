package host

import (
	"fmt"
	"os"
	"strings"

	"github.com/voocel/ainovel-cli/internal/domain"
	storepkg "github.com/voocel/ainovel-cli/internal/store"
)

// buildResumePrompt 基于事实生成 Resume 用的简短 prompt 与 UI 标签。
//
// 重构说明（2026-04-20）：所有"具体下一步应该做什么"的决策已下沉到 Host Flow Router。
// 本函数不再替 Coordinator 规划动作，只做三件事：
//  1. 判断是否需要恢复（Phase=Complete 或无 Progress → 返回空 label 表示新建）
//  2. 生成适合在 UI 展示的 label（"恢复：弧末评审待处理（V2 A3）" 之类）
//  3. 把用户停机期间留下的 PendingSteer 显式传给 Coordinator
//
// 返回 (prompt, label, error)。label 为空表示无可恢复状态（应走新建）。
func buildResumePrompt(store *storepkg.Store) (string, string, error) {
	progress, err := store.Progress.Load()
	if err != nil && !os.IsNotExist(err) {
		return "", "", err
	}
	if progress == nil || progress.Phase == domain.PhaseComplete {
		return "", "", nil
	}

	label := describeResume(store, progress)

	var b strings.Builder
	title := progress.NovelName
	if title == "" {
		title = "Tiểu thuyết hiện tại"
	}
	b.WriteString(fmt.Sprintf("[Khôi phục] Cuốn sách «%s»", title))
	if n := len(progress.CompletedChapters); n > 0 {
		b.WriteString(fmt.Sprintf(" đã hoàn thành %d chương", n))
		if progress.TotalChapters > 0 {
			b.WriteString(fmt.Sprintf(" (tổng cộng %d chương)", progress.TotalChapters))
		}
		b.WriteString(fmt.Sprintf(", tổng cộng %d chữ", progress.TotalWordCount))
	}
	b.WriteString("。\n")
	b.WriteString("Host sẽ đưa ra chỉ thị tiếp theo `[Host ra chỉ thị]` dựa trên thực tế hiện tại. Nhận được thì thực thi ngay, đừng gọi novel_context để suy luận trước.\n")

	if meta, _ := store.RunMeta.Load(); meta != nil && meta.PendingSteer != "" {
		b.WriteString("\nNgười dùng đã để lại một ý kiến can thiệp trong thời gian tạm dừng:\n「")
		b.WriteString(meta.PendingSteer)
		b.WriteString("」\nVui lòng đánh giá và xử lý theo quy tắc can thiệp của người dùng trong coordinator.md trước.")
	}

	return b.String(), label, nil
}

// describeResume 生成人类可读的恢复标签；不影响 Coordinator 的行为。
// 所有执行路由由 Flow Router 按事实推导；这里仅面向 UI 的 "恢复：xxx"。
func describeResume(store *storepkg.Store, progress *domain.Progress) string {
	switch progress.Phase {
	case domain.PhasePremise, domain.PhaseOutline:
		return fmt.Sprintf("Khôi phục: Giai đoạn quy hoạch (%s)", progress.Phase)
	case domain.PhaseWriting:
		// 优先级与 Router 的决策优先级对齐，让 label 与即将派发的指令一致。
		if pending, _ := store.Signals.LoadPendingCommit(); pending != nil {
			return fmt.Sprintf("Khôi phục: Đang gửi chương %d", pending.Chapter)
		}
		if len(progress.PendingRewrites) > 0 {
			verb := "Viết lại"
			if progress.Flow == domain.FlowPolishing {
				verb = "Gọt giũa"
			}
			return fmt.Sprintf("Khôi phục: Đợi %s %d chương", strings.ToLower(verb), len(progress.PendingRewrites))
		}
		if progress.Flow == domain.FlowReviewing {
			return "Khôi phục: Đang xét duyệt"
		}
		if progress.InProgressChapter > 0 {
			return fmt.Sprintf("Khôi phục: Đang viết chương %d", progress.InProgressChapter)
		}
		if label := describeArcEndLabel(store, progress); label != "" {
			return label
		}
		return fmt.Sprintf("Khôi phục: Tiếp tục từ chương %d", progress.NextChapter())
	}
	return "Khôi phục"
}

// describeArcEndLabel 为弧末/卷末的多种中间状态生成贴合 UI 的标签。
// 与 flow.Route 的弧末分支保持同序，保证 label 与 Router 首条指令对齐。
func describeArcEndLabel(store *storepkg.Store, progress *domain.Progress) string {
	if !progress.Layered || len(progress.CompletedChapters) == 0 {
		return ""
	}
	lastCh := progress.CompletedChapters[len(progress.CompletedChapters)-1]
	boundary, err := store.Outline.CheckArcBoundary(lastCh)
	if err != nil || boundary == nil || !boundary.IsArcEnd {
		return ""
	}
	vol, arc := boundary.Volume, boundary.Arc
	switch {
	case !store.World.HasArcReview(lastCh):
		return fmt.Sprintf("Khôi phục: Đợi xét duyệt cuối hồi (Quyển %d Hồi %d)", vol, arc)
	case !store.Summaries.HasArcSummary(vol, arc):
		return fmt.Sprintf("Khôi phục: Đợi tóm tắt hồi (Quyển %d Hồi %d)", vol, arc)
	case boundary.IsVolumeEnd && !store.Summaries.HasVolumeSummary(vol):
		return fmt.Sprintf("Khôi phục: Đợi tóm tắt quyển (Quyển %d)", vol)
	case boundary.NeedsExpansion && boundary.NextArc > 0:
		return fmt.Sprintf("Khôi phục: Đợi triển khai hồi tiếp theo (Quyển %d Hồi %d)", boundary.NextVolume, boundary.NextArc)
	case boundary.NeedsNewVolume:
		return fmt.Sprintf("Khôi phục: Đợi quyết định quyển tiếp theo (Cuối quyển %d)", vol)
	}
	return ""
}
