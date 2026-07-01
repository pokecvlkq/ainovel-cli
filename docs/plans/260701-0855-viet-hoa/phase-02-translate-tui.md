# Phase 02: Translate TUI & CLI (Source Code)
Status: ⬜ Pending
Dependencies: Phase 01

## Objective
Dịch tất cả các thông báo, menu, và giao diện dòng lệnh từ tiếng Trung sang tiếng Việt trong mã nguồn Go.

## Requirements
### Functional
- [ ] Thay thế các chuỗi ký tự tiếng Trung sang tiếng Việt trong mã nguồn.
- [ ] Đảm bảo ngữ nghĩa, văn phong tự nhiên.
- [ ] Không làm hỏng các logic định dạng biến của Golang (như `%s`, `%d`, `%v`).

## Implementation Steps
1. [ ] Thực hiện thay thế mã nguồn tại `internal/host/`, `internal/entry/`, v.v...
2. [ ] Thực hiện thay thế các lệnh trợ giúp (Help Menu) tại `cmd/`.

## Files to Create/Modify
- Các file `.go` chứa chuỗi giao diện.

---
Next Phase: Phase 03
