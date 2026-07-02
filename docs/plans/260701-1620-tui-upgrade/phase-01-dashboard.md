# Phase 01: Bảng Tiến Trình Tác Tử (Agent Dashboard)
Status: ⬜ Pending
Dependencies: Không

## Objective
Hiển thị trực quan trạng thái của từng tác tử (Coordinator, Architect, Writer, Editor) thay vì chỉ hiện log văn bản trôi tuột.

## Requirements
### Functional
- [ ] Phân bổ một khu vực (Panel) riêng trên giao diện để hiển thị trạng thái của các Agent.
- [ ] Thêm thanh tiến trình (Progress bar) cho Writer để báo cáo tiến độ sinh văn bản.
- [ ] Thêm hiệu ứng Spinner động cho các Agent đang suy nghĩ/làm việc.

### Non-Functional
- [ ] Hiệu suất: Đảm bảo TUI vẫn phản hồi mượt mà ở 60fps khi các thành phần động (Spinner, Progress bar) hoạt động.
- [ ] Tương thích: Tương thích với các loại Terminal tiêu chuẩn trên Windows (Windows Terminal, PowerShell).

## Implementation Steps
1. [ ] Bổ sung thư viện UI: Cấu hình `github.com/charmbracelet/bubbles/progress` và `spinner`.
2. [ ] Sửa đổi `panels_activity.go`: Thiết kế lại bố cục (layout) của khung log thành các thẻ (Cards) đại diện cho từng Agent.
3. [ ] Cập nhật `theme.go`: Cấu hình màu sắc, lipgloss style cho các thẻ Agent (Màu đang chạy, Hoàn thành, Lỗi).
4. [ ] Tích hợp trạng thái: Lắng nghe Event từ các tác tử và cập nhật Model để re-render giao diện tương ứng.

## Files to Create/Modify
- `internal/entry/tui/panels_activity.go` - Thay đổi cách hiển thị event log thành Dashboard Cards.
- `internal/entry/tui/theme.go` - Bổ sung các màu sắc động cho thanh tiến trình.
- `internal/entry/tui/model.go` - Cập nhật model state để chứa dữ liệu progress/spinner.

## Test Criteria
- [ ] Giao diện hiển thị đúng 4 Agent chính.
- [ ] Thanh tiến trình của Writer cập nhật liên tục từ 0% đến 100%.
- [ ] Spinner xoay mượt mà khi Coordinator/Architect đang làm việc.

---
Next Phase: `phase-02-editor.md`
