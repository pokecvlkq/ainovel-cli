package reminder

import (
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"

	"github.com/voocel/agentcore"
	"github.com/voocel/ainovel-cli/internal/domain"
	"github.com/voocel/ainovel-cli/internal/host/flow"
	"github.com/voocel/ainovel-cli/internal/store"
)

// StopGuard 是"物理不可停机"的最后防线。
// 当 LLM 试图 end_turn 时：
//   - Progress.Phase = Complete → 放行
//   - 否则注入 user message，让 agent 继续下一 turn
//   - 连续阻拦超过 maxConsecutive 次 → Escalate 终止 run（说明 prompt/reminder 严重失灵）
//
// Guard 内部维护 consecutive block 计数；一旦成功放行或成功注入后重置为 0。
// 真正驱动 Coordinator 行为的是 Reminder + Prompt，StopGuard 只是兜底。
const maxConsecutiveBlocks = 5

// NewStopGuard 构造 Coordinator 专用 StopGuard。
// onBlock 可选，非 nil 时每次阻拦调一次，用于审计。
func NewStopGuard(st *store.Store, onBlock func(reason string, consecutive int32)) agentcore.StopGuard {
	var consecutive atomic.Int32
	var lastBlockTurn atomic.Int64 // 上次 block 的 TurnIndex；-1 表示尚未 block 过
	lastBlockTurn.Store(-1)
	return func(_ context.Context, info agentcore.StopInfo) agentcore.StopDecision {
		progress, _ := st.Progress.Load()
		if progress != nil && progress.Phase == domain.PhaseComplete {
			consecutive.Store(0)
			lastBlockTurn.Store(-1)
			return agentcore.StopDecision{Allow: true}
		}
		// 只有"相邻 turn 连续被拦"才累计计数；否则视为新一轮（LLM 已做过 tool call 取得过进展，
		// 或用户注入 / resume 导致 TurnIndex 倒流），重置计数。
		last := lastBlockTurn.Load()
		if last < 0 || int64(info.TurnIndex) != last+1 {
			consecutive.Store(0)
		}
		lastBlockTurn.Store(int64(info.TurnIndex))
		n := consecutive.Add(1)
		if n > maxConsecutiveBlocks {
			slog.Error("stop_guard 连续阻拦超限，升级为终止",
				"module", "host.reminder", "turn", info.TurnIndex, "consecutive", n)
			if onBlock != nil {
				onBlock("escalated", n)
			}
			return agentcore.StopDecision{Allow: false, Escalate: true}
		}
		inject := blockMessage(st, progress)
		if progress != nil && len(progress.PendingRewrites) > 0 {
			inject = fmt.Sprintf("Cấm kết thúc phiên thoại. Hàng đợi viết lại vẫn chưa dọn xong: %v, vui lòng gọi ngay writer để xử lý.", progress.PendingRewrites)
		}
		slog.Warn("stop_guard 拦截 end_turn",
			"module", "host.reminder", "turn", info.TurnIndex, "consecutive", n)
		if onBlock != nil {
			onBlock("blocked", n)
		}
		return agentcore.StopDecision{Allow: false, InjectMessage: inject}
	}
}

func blockMessage(st *store.Store, progress *domain.Progress) string {
	if progress != nil && flow.Route(flow.LoadState(st)) != nil {
		return "Cấm kết thúc phiên thoại. Phase vẫn chưa Complete; vui lòng đợi và thực thi chỉ thị do Host ban hành [Host ra chỉ thị], không tự ý gọi novel_context hoặc subagent."
	}
	return "Cấm kết thúc phiên thoại. Phase vẫn chưa Complete, và hiện tại không có chỉ thị điều hướng nào từ Host; đây là tình huống chờ đánh giá của Coordinator, vui lòng tiếp tục xử lý theo quy tắc của coordinator.md, đừng chờ đợi vô ích chỉ thị của Host."
}
