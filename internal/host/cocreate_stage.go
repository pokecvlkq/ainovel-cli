package host

import (
	"fmt"
	"strings"

	"github.com/voocel/ainovel-cli/internal/store"
)

// buildStoryStateSummary 组装一段精简的故事现状摘要，供阶段共创助手了解"已经写了什么"。
// 复用 store 访问点，只取规划方向所需的高层事实（进度 / 罗盘 / 最近卷 / 主要人物 / 活跃伏笔）；
// 不拉正文、不喂 novel_context 的全量 JSON——共创是对话，要的是可读概览，不是写作上下文。
// 任一项缺失都跳过（best-effort），返回空串表示尚无可用进度。
func buildStoryStateSummary(s *store.Store) string {
	if s == nil {
		return ""
	}
	var b strings.Builder

	if progress, _ := s.Progress.Load(); progress != nil {
		if name := strings.TrimSpace(progress.NovelName); name != "" {
			fmt.Fprintf(&b, "- Tên sách: 《%s》\n", name)
		}
		fmt.Fprintf(&b, "- Tiến độ: Đã hoàn thành %d chương", len(progress.CompletedChapters))
		if progress.TotalChapters > 0 {
			fmt.Fprintf(&b, " / Quy hoạch %d chương", progress.TotalChapters)
		}
		fmt.Fprintf(&b, ", khoảng %d chữ, chương tiếp theo là chương %d\n", progress.TotalWordCount, progress.NextChapter())
		if progress.Layered && progress.CurrentVolume > 0 {
			fmt.Fprintf(&b, "- Vị trí hiện tại: Quyển %d Arc %d\n", progress.CurrentVolume, progress.CurrentArc)
		}
	}

	if compass, _ := s.Outline.LoadCompass(); compass != nil {
		if dir := strings.TrimSpace(compass.EndingDirection); dir != "" {
			fmt.Fprintf(&b, "- Hướng kết cục: %s\n", dir)
		}
		if compass.EstimatedScale != "" {
			fmt.Fprintf(&b, "- Quy mô ước lượng: %s\n", compass.EstimatedScale)
		}
		if len(compass.OpenThreads) > 0 {
			fmt.Fprintf(&b, "- Tuyến truyện dài hoạt động: %s\n", strings.Join(compass.OpenThreads, "; "))
		}
	}

	// 最近一卷摘要，让助手知道故事刚走到哪
	if vols, _ := s.Summaries.LoadAllVolumeSummaries(); len(vols) > 0 {
		last := vols[len(vols)-1]
		fmt.Fprintf(&b, "- Gần đây 《%s》: %s\n", last.Title, truncate(last.Summary, 200))
	}

	// 主要人物（core/important），最多 8 个
	if chars, _ := s.Characters.Load(); len(chars) > 0 {
		var names []string
		for _, c := range chars {
			if c.Tier == "secondary" || c.Tier == "decorative" {
				continue
			}
			line := c.Name
			if role := strings.TrimSpace(c.Role); role != "" {
				line += "（" + role + "）"
			}
			names = append(names, line)
			if len(names) >= 8 {
				break
			}
		}
		if len(names) > 0 {
			fmt.Fprintf(&b, "- Nhân vật chính: %s\n", strings.Join(names, ", "))
		}
	}

	// 未收伏笔，最多 6 条
	if fs, _ := s.World.LoadActiveForeshadow(); len(fs) > 0 {
		var items []string
		for _, f := range fs {
			items = append(items, truncate(f.Description, 40))
			if len(items) >= 6 {
				break
			}
		}
		fmt.Fprintf(&b, "- Phục bút chưa thu hồi: %s\n", strings.Join(items, "; "))
	}

	return strings.TrimSpace(b.String())
}

// stageSystemPrompt 组装阶段共创的完整系统提示：阶段 prompt + 当前故事状态摘要。
// 摘要作为数据附录挂在末尾（用分隔线与格式规范隔开），呼应 prompt 里"进度见下方"的指引。
func stageSystemPrompt(s *store.Store) string {
	prompt := stageCoCreateSystemPrompt
	if summary := buildStoryStateSummary(s); summary != "" {
		prompt += "\n\n---\n## Trạng thái câu chuyện hiện tại\n(Dưới đây là tóm tắt khách quan của nội dung đã viết, dùng để tham khảo khi quy hoạch phần tiếp theo, không sao chép nguyên văn vào <draft>)\n" + summary
	}
	return prompt
}
