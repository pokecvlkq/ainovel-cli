# Phase 04: Build & Testing
Status: ⬜ Pending
Dependencies: Phase 02, Phase 03

## Objective
Biên dịch lại toàn bộ dự án để xuất ra file `.exe` đã được Việt hoá và kiểm tra xem giao diện có bị lỗi hiển thị hay không.

## Requirements
### Functional
- [ ] File thực thi `ainovel-cli.exe` được build thành công.
- [ ] Khi chạy chương trình, giao diện (TUI) hiển thị tốt, tiếng Việt có dấu không bị vỡ font.
- [ ] Menu tương tác không bị lệch bố cục quá nhiều.

## Implementation Steps
1. [ ] Chạy lệnh `go build -o ainovel-cli.exe ./cmd/ainovel-cli`
2. [ ] Khởi chạy và chụp lại log hoặc kiểm tra kết quả hiển thị.
3. [ ] Căn chỉnh lại mã nguồn nếu giao diện bị vỡ.

## Files to Create/Modify
- `ainovel-cli.exe` (Binaries)

---
Next Phase: Finish
