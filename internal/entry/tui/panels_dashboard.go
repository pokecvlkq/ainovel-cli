package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"
	"github.com/voocel/ainovel-cli/internal/host"
)

// renderAgentDashboard hiển thị Dashboard ngang chứa 4 thẻ Agent.
func renderAgentDashboard(snap host.UISnapshot, prog progress.Model, width, spinnerFrame int) string {
	if width < 80 {
		width = 80 // Min width for dashboard
	}
	// Cấu hình 4 agents chính
	agents := []string{"coordinator", "architect", "writer", "editor"}
	
	var cards []string
	cardWidth := (width - 6) / 4 // 4 cards with some padding/gap

	for _, agentName := range agents {
		// Tìm trạng thái hiện tại của Agent
		var agentSnap *host.AgentSnapshot
		for i, a := range snap.Agents {
			if strings.ToLower(a.Name) == agentName {
				agentSnap = &snap.Agents[i]
				break
			}
		}

		card := renderAgentCard(agentName, agentSnap, prog, cardWidth, spinnerFrame, snap)
		cards = append(cards, card)
	}

	dashboard := lipgloss.JoinHorizontal(lipgloss.Top, cards...)
	return dashboard
}

func renderAgentCard(agentName string, snap *host.AgentSnapshot, prog progress.Model, width, spinnerFrame int, uiSnap host.UISnapshot) string {
	// Lấy màu đặc trưng của Agent
	agentColor := eventAgentColor(agentName)
	
	// Icon và Name
	titleStyle := lipgloss.NewStyle().Foreground(agentColor).Bold(true)
	nameLabel := agentDisplayName(agentName)
	
	var stateIcon, stateLabel string
	var stateColor lipgloss.AdaptiveColor
	
	if snap == nil || snap.State == "idle" {
		stateIcon = "○"
		stateLabel = "Chờ lệnh"
		stateColor = colorDim
	} else {
		stateIcon = agentStateIcon(snap.State)
		stateLabel = agentStateLabel(snap.State)
		stateColor = taskStatusColor(snap.State)
		if snap.State == "running" {
			stateIcon = runningSpinner(spinnerFrame)
		}
	}
	
	header := titleStyle.Render(nameLabel) + " " + lipgloss.NewStyle().Foreground(stateColor).Render(stateIcon+" "+stateLabel)

	// Lấy Task hiện tại
	taskDesc := "Đang nghỉ ngơi..."
	if snap != nil && snap.State != "idle" {
		if snap.Summary != "" {
			taskDesc = truncate(snap.Summary, width-4)
		} else if snap.Tool != "" {
			taskDesc = truncate("Đang dùng: "+snap.Tool, width-4)
		}
	}

	body := lipgloss.NewStyle().Foreground(colorMuted).Render(taskDesc)
	
	// Riêng Writer thì hiển thị thanh tiến trình
	footer := ""
	if agentName == "writer" {
		var percent float64 = 0
		if uiSnap.TotalChapters > 0 {
			percent = float64(uiSnap.CompletedCount) / float64(uiSnap.TotalChapters)
			if percent > 1.0 {
				percent = 1.0
			}
		}
		footer = "\n" + prog.ViewAs(percent)
	} else if snap != nil && snap.Context.ContextWindow > 0 {
		// Hiển thị phần trăm context
		ctxStr := fmt.Sprintf("Context: %.1f%%", snap.Context.Percent*100)
		footer = "\n" + lipgloss.NewStyle().Foreground(colorDim).Render(ctxStr)
	} else {
		footer = "\n" + lipgloss.NewStyle().Foreground(colorDim).Render(" ")
	}

	content := header + "\n" + body + footer

	return lipgloss.NewStyle().
		Width(width).
		Height(4).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(agentColor).
		Render(content)
}
