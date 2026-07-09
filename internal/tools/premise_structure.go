package tools

import (
	"strings"

	"github.com/voocel/ainovel-cli/internal/domain"
)

var premiseHeadingAliases = map[string]string{
	"题材定位":    "Định vị đề tài",
	"题材和基调":   "Đề tài và giọng điệu",
	"核心冲突":    "Xung đột cốt lõi",
	"主角目标":    "Mục tiêu của nhân vật chính",
	"结局方向":    "Hướng kết cục",
	"终局方向":    "Hướng kết cục",
	"写作禁区":    "Vùng cấm khi viết",
	"差异化卖点":   "Điểm thu hút khác biệt",
	"差异化钩子":   "Hook khác biệt",
	"核心兑现承诺":  "Cam kết cốt lõi",
	"故事引擎":    "Động cơ câu chuyện",
	"关系/成长主线": "Tuyến chính Quan hệ/Phát triển",
	"升级路径":    "Lộ trình thăng cấp",
	"中段转折":    "Chuyển hướng giữa chừng",
	"中期转向":    "Chuyển hướng giữa chừng",
	"终局命题":    "Mệnh đề chung cuộc",
	"短篇适配性":   "Tính tương thích truyện ngắn",
	"本作为什么适合短篇/单卷收束": "Tính tương thích truyện ngắn",

	// Vietnamese aliases
	// Vietnamese aliases
	"Đề tài và giọng điệu": "Đề tài và giọng điệu",
	"Đề tài & giọng điệu":  "Đề tài và giọng điệu",
	"Định vị đề tài":       "Định vị đề tài",
	"Định vị đề tài (độc giả mục tiêu, điểm tiêu thụ cốt lõi)": "Định vị đề tài",
	"Xung đột cốt lõi":            "Xung đột cốt lõi",
	"Mục tiêu của nhân vật chính": "Mục tiêu của nhân vật chính",
	"Hướng kết cục":               "Hướng kết cục",
	"Hướng kết cục (hướng theo chủ đề, không phải tên tập hay số chương)": "Hướng kết cục",
	"Vùng cấm khi viết":                       "Vùng cấm khi viết",
	"Điểm thu hút khác biệt":                  "Điểm thu hút khác biệt",
	"Điểm thu hút khác biệt (ít nhất 3 điểm)": "Điểm thu hút khác biệt",
	"Hook khác biệt":                          "Hook khác biệt",
	"Hook khác biệt: Điểm độc đáo nhất đáng để tiếp tục theo dõi": "Hook khác biệt",
	"Cam kết cốt lõi": "Cam kết cốt lõi",
	"Cam kết cốt lõi: Cuốn sách liên tục mang lại gì cho độc giả": "Cam kết cốt lõi",
	"Động cơ câu chuyện": "Động cơ câu chuyện",
	"Động cơ câu chuyện: Yếu tố thúc đẩy bên ngoài và bên trong là gì": "Động cơ câu chuyện",
	"Tuyến chính Quan hệ/Phát triển":                                   "Tuyến chính Quan hệ/Phát triển",
	"Tuyến chính Quan hệ / Phát triển":                                 "Tuyến chính Quan hệ/Phát triển",
	"Tuyến chính Quan hệ/Phát triển: Quan hệ và sự trưởng thành của nhân vật phát triển qua các phần như thế nào": "Tuyến chính Quan hệ/Phát triển",
	"Lộ trình thăng cấp": "Lộ trình thăng cấp",
	"Lộ trình thăng cấp: Giai đoạn đầu, giữa, cuối truyện dựa vào đâu để thăng cấp": "Lộ trình thăng cấp",
	"Chuyển hướng giữa chừng": "Chuyển hướng giữa chừng",
	"Chuyển hướng giữa chừng: Khi nào phương pháp đầu truyện mất tác dụng, câu chuyện chuyển hướng thế nào": "Chuyển hướng giữa chừng",
	"Mệnh đề chung cuộc": "Mệnh đề chung cuộc",
	"Mệnh đề chung cuộc: Câu hỏi cuối cùng thực sự cần trả lời ở giai đoạn cuối": "Mệnh đề chung cuộc",
	"Tính tương thích truyện ngắn": "Tính tương thích truyện ngắn",
}

func parsePremiseSections(premise string) map[string]string {
	lines := strings.Split(premise, "\n")
	sections := make(map[string]string)
	var current string
	var body []string

	flush := func() {
		if current == "" {
			return
		}
		text := strings.TrimSpace(strings.Join(body, "\n"))
		if text != "" {
			sections[current] = text
		}
		body = body[:0]
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if heading, ok := canonicalPremiseHeading(trimmed); ok {
			flush()
			current = heading
			continue
		}
		if current != "" {
			body = append(body, line)
		}
	}
	flush()
	return sections
}

func canonicalPremiseHeading(line string) (string, bool) {
	if !strings.HasPrefix(line, "#") {
		return "", false
	}
	title := strings.TrimSpace(strings.TrimLeft(line, "#"))
	if title == "" {
		return "", false
	}
	canonical, ok := premiseHeadingAliases[title]
	return canonical, ok
}

func premiseStructure(premise string, tier domain.PlanningTier) map[string]any {
	sections := parsePremiseSections(premise)
	required := requiredPremiseHeadings(tier)
	found := make([]string, 0, len(required))
	var missing []string
	for _, heading := range required {
		if _, ok := sections[heading]; ok {
			found = append(found, heading)
			continue
		}
		missing = append(missing, heading)
	}

	structure := map[string]any{
		"template_ready": len(missing) == 0,
		"found":          found,
		"missing":        missing,
	}
	if len(sections) > 0 {
		structure["section_count"] = len(sections)
	}
	return structure
}

func requiredPremiseHeadings(tier domain.PlanningTier) []string {
	common := []string{
		"Đề tài và giọng điệu",
		"Định vị đề tài",
		"Xung đột cốt lõi",
		"Mục tiêu của nhân vật chính",
		"Hướng kết cục",
		"Vùng cấm khi viết",
		"Điểm thu hút khác biệt",
		"Hook khác biệt",
		"Cam kết cốt lõi",
	}

	switch tier {
	case domain.PlanningTierLong:
		return append(common,
			"Động cơ câu chuyện",
			"Tuyến chính Quan hệ/Phát triển",
			"Lộ trình thăng cấp",
			"Chuyển hướng giữa chừng",
			"Mệnh đề chung cuộc",
		)
	case domain.PlanningTierMid:
		return append(common,
			"Động cơ câu chuyện",
			"Chuyển hướng giữa chừng",
		)
	case domain.PlanningTierShort:
		return append(common,
			"Tính tương thích truyện ngắn",
		)
	default:
		return common
	}
}
