# Plan: Nâng Cấp TUI UI/UX (Terminal UI)
Created: 260701-1620
Status: 🟡 In Progress

## Overview
Nâng cấp trải nghiệm người dùng (UX) và giao diện dòng lệnh (TUI) cho ứng dụng `ainovel-cli` dựa trên nền tảng thư viện Bubble Tea. Giải quyết các vấn đề:
1. Thiếu trực quan trong việc theo dõi tiến trình tác tử.
2. Thiếu khả năng chỉnh sửa trực tiếp nội dung chương truyện.
3. Không hỗ trợ xem và duyệt bản thảo (diff viewer) một cách trực quan.

## Tech Stack
- Khung giao diện: `charmbracelet/bubbletea`
- Hiển thị và Bố cục: `charmbracelet/lipgloss`
- Thành phần UI mở rộng: `charmbracelet/bubbles` (progress, textarea, spinner)
- Diff engine: `sergi/go-diff`

## Phases

| Phase | Name | Status | Progress |
|-------|------|--------|----------|
| 01 | Bảng Tiến Trình Tác Tử (Agent Dashboard) | ⬜ Pending | 0% |
| 02 | Trình Soạn Thảo (TUI Text Editor) | ⬜ Pending | 0% |
| 03 | Trình Duyệt & So Sánh (Chapter Reviewer) | ⬜ Pending | 0% |

## Quick Commands
- Start Phase 1: `/code phase-01`
- Check progress: `/next`
- Save context: `/save-brain`
