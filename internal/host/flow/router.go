// Package flow 实现垂类路由：Host 根据事实决定下一个调哪个子代理做什么。
//
// 设计原则：
//   - Route 是纯函数：输入 State，输出 *Instruction。无 IO、无 Store 调用，可单测。
//   - State 由 LoadState（非纯）从 Store 构造，一次性把路由需要的事实读齐。
//   - 返回 nil 是合法的：表示"裁定场景，让 Coordinator LLM 自主决策"。
//
// Router 覆盖的是"查表型"决策（每章下一步、弧末后处理、队列驱动），
// 不覆盖"语义理解型"决策（选规划师、处理用户 Steer、输出总结）。
package flow

import (
	"fmt"

	"github.com/voocel/ainovel-cli/internal/domain"
	storepkg "github.com/voocel/ainovel-cli/internal/store"
)

// Instruction 指示 Host 下一步要求 Coordinator 调用的子代理与任务。
type Instruction struct {
	Agent   string // architect_long / architect_short / writer / editor
	Task    string // 给子代理的任务描述
	Reason  string // 给 Coordinator 看的理由（可选，方便调试与日志）
	Chapter int    // writer 任务涉及的章节号（续写/重写/打磨）；0 表示不涉及（editor/architect 任务）
}

// State 是 Route 的输入：所有事实必须在此显式声明，禁止 Route 内部读 Store。
type State struct {
	Progress *domain.Progress

	// 上一个已完成章节（Progress.CompletedChapters 末尾）；为 0 表示尚未开始写作。
	LastCompleted int

	// 上一章的弧边界信息；IsArcEnd=false 时其他字段无意义。
	// 当 LastCompleted=0 或非 Layered 模式时应为 nil。
	ArcBoundary *storepkg.ArcBoundary

	// 弧末后处理的三个事实：评审 / 弧摘要 / 卷摘要是否已完成。
	HasArcReview     bool
	HasArcSummary    bool
	HasVolumeSummary bool

	// 基础设定缺项（规划阶段的补齐信号）。
	FoundationMissing []string
}

// Route 根据事实返回下一步指令；返回 nil 表示让 Coordinator LLM 自主裁定。
//
// 决策优先级（互斥，自上而下匹配第一个）：
//  1. Phase=Complete        → nil（LLM 输出总结）
//  2. Phase!=Writing        → nil（LLM 裁定规划师选型 / 规划补齐）
//  3. PendingRewrites 非空  → writer 按队列重写/打磨
//  4. Flow=Reviewing        → nil（editor 刚保存 review，verdict 分叉由工具层处理）
//  5. Flow=Steering         → nil（用户干预处理中）
//  6. 弧末评审缺失           → editor(arc review)
//  7. 弧末评审有但弧摘要缺失  → editor(arc summary)
//  8. 卷末弧摘要有但卷摘要缺失 → editor(volume summary)
//  9. 下一弧是骨架           → architect_long(expand_arc)
//
// 10. 卷末需决策下一卷       → architect_long(append_volume / complete_book)
// 11. 其它                  → writer(写 next_chapter)
func Route(s State) *Instruction {
	p := s.Progress
	if p == nil {
		return nil
	}

	// 1. 终态：让 LLM 输出总结
	if p.Phase == domain.PhaseComplete {
		return nil
	}

	// 2. 规划阶段由 Coordinator 裁定（选 architect_long/short + 补齐循环）
	if p.Phase != domain.PhaseWriting {
		return nil
	}

	// 3. 重写/打磨队列优先（事实已在工具层落盘，Router 只照单派发）
	if len(p.PendingRewrites) > 0 {
		ch := p.PendingRewrites[0]
		verb := "Viết lại"
		if p.Flow == domain.FlowPolishing {
			verb = "Gọt giũa"
		}
		return &Instruction{
			Agent:   "writer",
			Task:    fmt.Sprintf("%s chương %d", verb, ch),
			Reason:  fmt.Sprintf("Hàng đợi PendingRewrites còn lại %d chương", len(p.PendingRewrites)),
			Chapter: ch,
		}
	}

	// 4. 审阅中：save_review 刚落盘，verdict 升级/降级由工具层处理，路由不介入
	if p.Flow == domain.FlowReviewing {
		return nil
	}

	// 5. 用户干预处理中：Coordinator 正在裁定，Host 不抢占
	if p.Flow == domain.FlowSteering {
		return nil
	}

	// 6-10. 分层模式的弧末后处理
	if p.Layered && s.ArcBoundary != nil && s.ArcBoundary.IsArcEnd {
		b := s.ArcBoundary
		switch {
		case !s.HasArcReview:
			return &Instruction{
				Agent:  "editor",
				Task:   fmt.Sprintf("Tiến hành đánh giá Hồi đối với Quyển %d Hồi %d (scope=arc)", b.Volume, b.Arc),
				Reason: "Chưa hoàn thành đánh giá Hồi cuối",
			}
		case !s.HasArcSummary:
			return &Instruction{
				Agent:  "editor",
				Task:   fmt.Sprintf("Tạo tóm tắt Hồi cho Quyển %d Hồi %d (save_arc_summary)", b.Volume, b.Arc),
				Reason: "Chưa hoàn thành tóm tắt Hồi",
			}
		case b.IsVolumeEnd && !s.HasVolumeSummary:
			return &Instruction{
				Agent:  "editor",
				Task:   fmt.Sprintf("Tạo tóm tắt Quyển cho Quyển %d (save_volume_summary)", b.Volume),
				Reason: "Chưa hoàn thành tóm tắt Quyển",
			}
		case b.NeedsExpansion && b.NextArc > 0:
			return &Instruction{
				Agent:  "architect_long",
				Task:   fmt.Sprintf("Triển khai Quyển %d Hồi %d (save_foundation type=expand_arc)", b.NextVolume, b.NextArc),
				Reason: "Đang chờ triển khai khung Hồi tiếp theo",
			}
		case b.NeedsNewVolume:
			return &Instruction{
				Agent:  "architect_long",
				Task:   "Sau khi đánh giá, gọi save_foundation type=append_volume (viết tiếp) hoặc type=complete_book (kết thúc toàn thư)",
				Reason: "Cuối quyển cần quyết định thêm quyển mới hoặc kết thúc toàn thư",
			}
		}
	}

	// 12. 正常续写
	next := p.NextChapter()
	if next <= 0 {
		return nil
	}
	return &Instruction{
		Agent:   "writer",
		Task:    fmt.Sprintf("Viết chương %d", next),
		Reason:  "Viết tiếp chương sau",
		Chapter: next,
	}
}

// FormatMessage bọc Instruction thành tin nhắn gửi cho Coordinator.
func FormatMessage(i *Instruction) string {
	return fmt.Sprintf(
		"[Host ra chỉ thị]\nBước tiếp theo: gọi subagent(%s, %q)\nagent: %s\ntask: %q\nLý do: %s\nĐây là chỉ thị rõ ràng của tầng quy trình, vui lòng thực thi ngay lập tức; tham số agent/task của subagent phải được sử dụng nguyên văn như agent/task ở trên, không được viết lại task, không được gọi novel_context trước, không được xuất ra suy luận trước.",
		i.Agent, i.Task, i.Agent, i.Task, i.Reason,
	)
}
