package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/voocel/ainovel-cli/internal/host"
	"github.com/voocel/ainovel-cli/internal/store"
)

// renderTopBar 渲染顶部状态栏。
// 左侧：provider/model，中间：书名，右侧：状态胶囊。
func renderTopBar(snap host.UISnapshot, width int, spinnerFrame, version string) string {
	novelName := snap.NovelName
	if novelName == "" {
		novelName = "Chưa có tên truyện"
	}

	var infoParts []string
	if version != "" {
		infoParts = append(infoParts, "ainovel-cli "+version)
	}
	if snap.Provider != "" {
		infoParts = append(infoParts, snap.Provider)
	}
	if snap.ModelName != "" {
		if w := formatContextWindow(snap.ModelContextWindow); w != "" {
			infoParts = append(infoParts, snap.ModelName+"("+w+")")
		} else {
			infoParts = append(infoParts, snap.ModelName)
		}
	}
	if snap.Style != "" && snap.Style != "default" {
		infoParts = append(infoParts, snap.Style)
	}
	leftText := strings.Join(infoParts, " · ")

	label := snap.StatusLabel
	if label == "" {
		label = "READY"
	}
	color, ok := statusColors[label]
	if !ok {
		color = colorDim
	}
	disp, ok := statusDisplay[label]
	if !ok {
		disp = struct {
			icon  string
			label string
		}{"○", strings.ToLower(label)}
	}
	icon := disp.icon
	if snap.IsRunning && spinnerFrame != "" {
		icon = spinnerFrame
	}
	var status string
	if icon != "" {
		status = statusIconStyle.Foreground(color).Render(icon) + " " + statusLabelStyle.Render(disp.label)
	} else {
		status = statusLabelStyle.Render(disp.label)
	}

	innerW := max(12, width-2)
	titleText := truncate(novelName, max(8, innerW/3))
	centerW := max(16, lipgloss.Width(titleText)+6)
	if centerW > innerW-24 {
		centerW = max(8, innerW-24)
	}
	sideTotal := innerW - centerW
	if sideTotal < 0 {
		sideTotal = 0
		centerW = innerW
	}
	leftW := sideTotal / 2
	rightW := innerW - centerW - leftW

	leftCell := lipgloss.NewStyle().
		Width(leftW).
		AlignHorizontal(lipgloss.Left).
		Foreground(colorDim).
		Render(truncate(leftText, leftW))
	centerCell := lipgloss.NewStyle().
		Width(centerW).
		AlignHorizontal(lipgloss.Center).
		Bold(true).
		Foreground(bodyTextColor).
		Render(titleText)
	rightCell := lipgloss.NewStyle().
		Width(rightW).
		AlignHorizontal(lipgloss.Right).
		Render(status)

	content := leftCell + centerCell + rightCell
	return topBarStyle.Width(width).
		Border(baseBorder, false, false, true, false).
		BorderForeground(colorDim).
		Render(content)
}

// renderStatePanel 把状态侧栏内容(已在 stateVP 中)包进左侧带右边框的盒子。
// 与 renderDetailPanel 对称：内容由 renderStateContent 生成并喂进 viewport，这里只负责框。
// MaxHeight 钳高，防止窗口缩小时溢出比右栏高（见 panels_test.go 的高度契约）。
func renderStatePanel(vp viewport.Model, width, height int, focused bool) string {
	borderColor := colorDim
	if focused {
		borderColor = colorAccent
	}
	style := lipgloss.NewStyle().
		Width(width).
		Height(height).
		MaxHeight(height).
		Border(baseBorder, false, true, false, false).
		BorderForeground(borderColor).
		Padding(1, 1)
	return style.Render(vp.View())
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// renderDetailPanel 渲染右侧可滚动详情面板。
func renderDetailPanel(vp viewport.Model, width, height int, focused bool) string {
	borderColor := colorDim
	if focused {
		borderColor = colorAccent
	}
	style := lipgloss.NewStyle().
		Width(width).
		Height(height).
		MaxHeight(height).
		Border(baseBorder, false, false, false, true).
		BorderForeground(borderColor).
		Padding(0, 1)

	return style.Render(vp.View())
}

// renderWelcome 渲染新建态首屏。
func renderWelcome(width, height int, errMsg string, mode startupMode) string {
	// 简洁标题
	title := lipgloss.NewStyle().
		Foreground(colorAccent).
		Bold(true).
		Render("A I N O V E L")

	// 副标题
	subtitle := lipgloss.NewStyle().
		Foreground(colorMuted).
		Italic(true).
		Render("AI-Powered Novel Creation Engine")

	// 分隔线
	divW := 44
	if divW > width-8 {
		divW = width - 8
	}
	divider := lipgloss.NewStyle().Foreground(colorDim).
		Render(strings.Repeat("~", divW))

	// 功能亮点
	features := []struct{ icon, label, desc string }{
		{">>", "Nhiều model kết hợp", "Architect Lên KH / Writer Viết / Editor Duyệt"},
		{"::", "Khôi phục", "Tự động viết tiếp sau khi gián đoạn"},
		{"<>", "Can thiệp realtime", "Điều chỉnh cốt truyện bất kỳ lúc nào"},
		{"##", "Truyện dài phân tầng", "Hỗ trợ cấu trúc Quyển-Phần-Chương"},
	}
	iconStyle := lipgloss.NewStyle().Foreground(colorAccent2).Bold(true)
	featLabelStyle := lipgloss.NewStyle().Foreground(bodyTextColor)
	descStyle := lipgloss.NewStyle().Foreground(colorDim)
	var featLines []string
	for _, f := range features {
		line := iconStyle.Render(f.icon) + " " +
			featLabelStyle.Render(f.label) + "  " +
			descStyle.Render(f.desc)
		featLines = append(featLines, line)
	}
	feats := strings.Join(featLines, "\n")

	// 输入提示
	prompt := lipgloss.NewStyle().Foreground(bodyTextColor).Render("Nhập yêu cầu tiểu thuyết bên dưới để bắt đầu")

	modeLine := lipgloss.NewStyle().
		Foreground(colorMuted).
		Render("Chế độ hiện tại: " + mode.label() + " · " + mode.subtitle())

	// 示例
	examples := []string{
		"Viết tiểu thuyết trinh thám đô thị 12 chương, nữ chính là pháp y",
		"Viết truyện Tiên hiệp, nam chính tu luyện phi thăng",
		"Viết truyện ngắn Sci-Fi về đạo đức AI",
	}
	exStyle := lipgloss.NewStyle().Foreground(colorAccent)
	dotStyle := lipgloss.NewStyle().Foreground(colorDim)
	var exLines []string
	for _, ex := range examples {
		exLines = append(exLines, dotStyle.Render("  . ")+exStyle.Render(ex))
	}
	exBlock := strings.Join(exLines, "\n")

	// 组装
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(title)
	b.WriteString("\n")
	b.WriteString(subtitle)
	b.WriteString("\n\n")
	b.WriteString(divider)
	b.WriteString("\n\n")
	b.WriteString(feats)
	b.WriteString("\n\n")
	b.WriteString(divider)
	b.WriteString("\n\n")
	b.WriteString(modeLine)
	b.WriteString("\n\n")
	b.WriteString(prompt)
	b.WriteString("\n\n")
	b.WriteString(exBlock)
	b.WriteString("\n\n")
	b.WriteString(lipgloss.NewStyle().Foreground(colorDim).Italic(true).
		Render("Tab Chuyển chế độ · Bắt đầu nhanh: Enter Viết luôn · Lên kế hoạch: Enter để chat"))

	if errMsg != "" {
		b.WriteString("\n\n")
		b.WriteString(lipgloss.NewStyle().Foreground(colorError).Bold(true).Render("! " + errMsg))
	}

	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		AlignHorizontal(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Render(b.String())
}

// renderProjectPicker hiển thị màn hình chọn dự án.
func renderProjectPicker(width, height int, projects []store.ProjectInfo, projectIdx int) string {
	title := lipgloss.NewStyle().Foreground(colorAccent).Bold(true).Render("CHỌN DỰ ÁN TRUYỆN")
	subtitle := lipgloss.NewStyle().Foreground(colorMuted).Render("Danh sách các dự án đang viết")

	divW := 60
	if divW > width-8 {
		divW = width - 8
	}
	divider := lipgloss.NewStyle().Foreground(colorDim).Render(strings.Repeat("~", divW))

	var listLines []string
	for i, p := range projects {
		prefix := "  "
		style := lipgloss.NewStyle().Foreground(bodyTextColor)
		if i == projectIdx {
			prefix = "> "
			style = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
		}

		name := p.NovelName
		if name == "" {
			name = "Chưa có tên truyện (" + p.DirName + ")"
		}

		timeStr := p.LastUpdated.Format("02/01/2006 15:04")
		info := fmt.Sprintf("%d chương · %d chữ · Cập nhật: %s", p.ChapterCount, p.TotalRealWordCount, timeStr)

		line := prefix + name + "  " + lipgloss.NewStyle().Foreground(colorDim).Render(info)
		listLines = append(listLines, style.Render(line))
	}

	listBlock := strings.Join(listLines, "\n")

	prompt := lipgloss.NewStyle().Foreground(colorDim).Italic(true).
		Render("↑/↓ Chọn dự án · Enter Tiếp tục · N Tạo mới truyện")

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(title)
	b.WriteString("\n")
	b.WriteString(subtitle)
	b.WriteString("\n\n")
	b.WriteString(divider)
	b.WriteString("\n\n")
	b.WriteString(listBlock)
	b.WriteString("\n\n")
	b.WriteString(divider)
	b.WriteString("\n\n")
	b.WriteString(prompt)

	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		AlignHorizontal(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Render(b.String())
}
