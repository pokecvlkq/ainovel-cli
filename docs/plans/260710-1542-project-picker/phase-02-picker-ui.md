# Phase 02: TUI Project Picker Screen
Status: ? Pending
Dependencies: Phase 01

## Objective
T?o giao di?n TUI hi?n th? danh sách các d? án (Project Picker) khi kh?i d?ng.

## Requirements
### Functional
- [ ] Khi kh?i d?ng (`mode = modeProjectPicker`), g?i hàm `DiscoverProjects` d? l?y danh sách.
- [ ] N?u không có d? án nào, chuy?n th?ng sang `modeNew`.
- [ ] Render danh sách các d? án. Cung c?p thông tin: Tên truy?n, Chuong hi?n t?i, T?ng ch?, Ngày c?p nh?t.
- [ ] S? d?ng ?? d? di chuy?n cursor. Highlight d? án du?c ch?n.
- [ ] Nh?n Enter d? m? d? án du?c ch?n (`mode = modeRunning` ho?c ch? d? ch? l?nh, thay d?i OutputDir thành thu m?c tuong ?ng và g?i `bootstrapRuntime`).
- [ ] Nh?n N d? t?o d? án m?i (chuy?n sang `modeNew`).

## Implementation Steps
1. [ ] C?p nh?t `internal/entry/tui/model.go`: thêm tr?ng thái `modeProjectPicker`, tru?ng `projects`, `projectIdx`.
2. [ ] S?a d?i lu?ng kh?i t?o `NewModel()` ho?c logic ? màn hình d?u d? vào `modeProjectPicker`.
3. [ ] Vi?t hàm `renderProjectPicker` trong `internal/entry/tui/panels.go`.
4. [ ] C?p nh?t `Update()` trong `model_update.go` d? b?t phím up/down/enter/n trong `modeProjectPicker`.

## Files to Create/Modify
- `internal/entry/tui/model.go`
- `internal/entry/tui/model_update.go`
- `internal/entry/tui/panels.go`

## Test Criteria
- [ ] Hi?n th? d?p m?t, s? d?ng component list ho?c t? render view.
- [ ] Phím t?t ho?t d?ng mu?t mà.
- [ ] Khôi ph?c du?c d? án cu b?ng cách ch?n và Enter.

---
Next Phase: [Phase 03: Auto Rename Workspace](phase-03-auto-rename.md)
