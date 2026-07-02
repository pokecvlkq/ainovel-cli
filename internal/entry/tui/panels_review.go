package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sergi/go-diff/diffmatchpatch"
)

var (
	diffAddStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Background(lipgloss.Color("22")) // Green
	diffDelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Strikethrough(true)             // Red
	diffEqStyle  = lipgloss.NewStyle()
)

type ReviewModel struct {
	viewport viewport.Model
	diffs    []diffmatchpatch.Diff
	ready    bool
}

func NewReviewModel() ReviewModel {
	return ReviewModel{}
}

func (m *ReviewModel) SetSize(width, height int) {
	if !m.ready {
		m.viewport = viewport.New(width, height-2)
		m.ready = true
	} else {
		m.viewport.Width = width
		m.viewport.Height = height - 2
	}
}

func (m *ReviewModel) SetDiff(oldText, newText string) {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(oldText, newText, false)
	m.diffs = dmp.DiffCleanupSemantic(diffs)
	m.viewport.SetContent(m.renderDiff())
}

func (m ReviewModel) renderDiff() string {
	var sb strings.Builder
	for _, diff := range m.diffs {
		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			sb.WriteString(diffAddStyle.Render(diff.Text))
		case diffmatchpatch.DiffDelete:
			sb.WriteString(diffDelStyle.Render(diff.Text))
		case diffmatchpatch.DiffEqual:
			sb.WriteString(diffEqStyle.Render(diff.Text))
		}
	}
	return sb.String()
}

func (m ReviewModel) Init() tea.Cmd {
	return nil
}

func (m ReviewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m ReviewModel) View() string {
	if !m.ready {
		return "Đang tải dữ liệu Diff..."
	}
	footer := "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("[A] Duyệt (Approve)  |  [R] Yêu cầu viết lại (Reject)  |  [E] Sửa thủ công  |  [Esc] Thoát")
	return fmt.Sprintf("%s\n%s", m.viewport.View(), footer)
}
