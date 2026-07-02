# Phase 02: Trình Soạn Thảo (TUI Text Editor)
Status: ⬜ Pending
Dependencies: phase-01-dashboard.md

## Objective
Cho phép người dùng nhấn phím tắt (ví dụ `E`) để mở màn hình chỉnh sửa toàn khung hình cho chương truyện hiện tại ngay bên trong TUI, không cần thoát ra mở VS Code.

## Requirements
### Functional
- [ ] Tích hợp thành phần nhập liệu đa dòng (Textarea).
- [ ] Bắt phím tắt (Hotkeys): `E` (từ màn hình chính) để Mở Editor; `Ctrl+S` để Lưu; `Esc` để Hủy/Đóng.
- [ ] Lưu nội dung đã sửa trực tiếp vào bộ nhớ struct tiểu thuyết.

### Non-Functional
- [ ] Hỗ trợ gõ Tiếng Việt Unicode không bị lỗi dấu.
- [ ] Trải nghiệm cuộn trang mượt mà (Scrolling) đối với các chương văn bản dài.

## Implementation Steps
1. [ ] Tạo file mới `panels_editor.go` và định nghĩa Textarea Model (`bubbles/textarea`).
2. [ ] Khởi tạo Textarea component, cấu hình `lipgloss` để nó chiếm toàn màn hình hoặc một Panel lớn.
3. [ ] Sửa đổi vòng lặp sự kiện `model_update.go`:
      - Khi nhận phím `E` -> Chuyển state sang `stateEditing`, nạp nội dung chương vào textarea.
      - Khi ở `stateEditing` -> Route các sự kiện phím sang `textarea.Update()`.
      - Khi nhận `Ctrl+S` -> Lấy nội dung mới từ `textarea.Value()`, lưu vào tiểu thuyết, thoát `stateEditing`.
      - Khi nhận `Esc` -> Thoát không lưu.

## Files to Create/Modify
- `internal/entry/tui/panels_editor.go` (Mới) - Model và view cho Textarea.
- `internal/entry/tui/model.go` - Thêm state mới `stateEditing`.
- `internal/entry/tui/model_update.go` - Xử lý chuyển đổi state và phím tắt.
- `internal/entry/tui/events.go` - Xử lý lưu nội dung.

## Test Criteria
- [ ] Nhấn `E` mở được Textarea với nội dung chương hiện tại.
- [ ] Gõ được văn bản, cuộn lên xuống được.
- [ ] Nhấn `Ctrl+S` lưu thành công, văn bản mới hiển thị trên Outline/Preview.

---
Next Phase: `phase-03-reviewer.md`
