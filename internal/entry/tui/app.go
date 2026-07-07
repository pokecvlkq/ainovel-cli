package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/voocel/ainovel-cli/assets"
	"github.com/voocel/ainovel-cli/internal/bootstrap"
	"github.com/voocel/ainovel-cli/internal/host"
	"github.com/voocel/ainovel-cli/internal/logger"
)

// Run khởi chạy TUI.
// Quy ước phân tầng chế độ khởi động:
// 1. Chế độ nhanh, chế độ đồng sáng tạo thuộc về "điều phối khởi động";
// 2. Phiên sáng tác chính thức đi vào host.Host;
// 3. Trong tương lai nếu thêm các chế độ chia sẻ như "viết tiếp tiểu thuyết có sẵn", thống nhất đưa vào internal/entry/startup.
func Run(cfg bootstrap.Config, bundle assets.Bundle, version string) error {
	rt, err := host.New(cfg, bundle)
	if err != nil {
		return err
	}
	bridge := newAskUserBridge()
	rt.AskUser().SetHandler(bridge.handler)
	cleanup := logger.SetupFile(rt.Dir(), "tui.log", false)
	defer cleanup()
	defer rt.Close()

	m := NewModel(rt, bridge, version)
	// Không bật báo cáo chuột toàn cục khi khởi động: trang chào mừng không dùng chuột, tắt báo cáo có thể giữ nguyên
	// thao tác kéo chọn sao chép gốc của terminal. Khi vào không gian làm việc (modeRunning) mới do enterRunning bật báo cáo,
	// để hỗ trợ nhấp chuyển bảng / cuộn chuột / kéo thanh bên.
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}
