package tui

import (
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/lipgloss"
)

// renderEditorScreen hiển thị trình soạn thảo văn bản chiếm toàn màn hình.
func renderEditorScreen(width, height int, ed textarea.Model, err error) string {
	// Tiêu đề
	header := lipgloss.NewStyle().
		Width(width).
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("230")).
		Padding(0, 1).
		Bold(true).
		Render("🖋 Trình soạn thảo trực tiếp (Nhấn Ctrl+S để lưu, Esc để đóng)")

	// Thông báo lỗi nếu có
	var errBox string
	if err != nil {
		errBox = lipgloss.NewStyle().
			Width(width).
			Background(lipgloss.Color("196")).
			Foreground(lipgloss.Color("230")).
			Padding(0, 1).
			Render("Lỗi: " + err.Error())
	}

	// Nội dung
	ed.SetWidth(width)

	// Tính toán chiều cao khả dụng
	availHeight := height - lipgloss.Height(header)
	if errBox != "" {
		availHeight -= lipgloss.Height(errBox)
	}

	ed.SetHeight(availHeight)

	// Hiển thị
	content := ed.View()

	if errBox != "" {
		return lipgloss.JoinVertical(lipgloss.Left, header, errBox, content)
	}
	return lipgloss.JoinVertical(lipgloss.Left, header, content)
}
