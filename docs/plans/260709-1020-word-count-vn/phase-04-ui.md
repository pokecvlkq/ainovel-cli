# Phase 04: UI/TUI + Snapshot
Status: ⬜ Pending
Dependencies: Phase 03

## Objective
Hiển thị cả 2 bộ đếm trên giao diện TUI sidebar. Label rõ ràng: "Số ký tự" và "Số chữ".

## Implementation Steps

### 4.1 UISnapshot (`internal/host/events.go`)
1. [ ] Thêm field `TotalRealWordCount int` vào struct `UISnapshot`
2. [ ] Đổi tên (hoặc thêm comment) cho `TotalWordCount` — field này giờ mang nghĩa "char count" nhưng giữ nguyên tên để backward-compat

### 4.2 Host snapshot builder (`internal/host/host.go`)
3. [ ] Trong hàm build snapshot, gán thêm `snap.TotalRealWordCount = progress.TotalRealWordCount`

### 4.3 Sidebar (`internal/entry/tui/panels_sidebar.go`)
4. [ ] Đổi label dòng hiện tại: `"Số từ"` → `"Số ký tự"` — hiển thị `snap.TotalWordCount`
5. [ ] Thêm dòng mới ngay dưới: `"Số chữ"` — hiển thị `snap.TotalRealWordCount`

### 4.4 Outline panel (`internal/entry/tui/panels_outline.go`)
6. [ ] Kiểm tra xem panel outline có hiển thị word count per-chapter không, nếu có thì cập nhật tương tự

### 4.5 App-level snapshot (`app.go`)
7. [ ] Gán `TotalRealWordCount` khi xây dựng UISnapshot từ Progress

### 4.6 Diagnostics (`internal/diag/diag.go`)
8. [ ] Cập nhật `st.TotalWords` → dùng `TotalRealWordCount` (hoặc thêm field `TotalChars` nếu cần)

## Files to Create/Modify
- `internal/host/events.go` — Thêm field vào UISnapshot
- `internal/host/host.go` — Gán giá trị mới cho snapshot
- `internal/entry/tui/panels_sidebar.go` — Đổi label + thêm dòng
- `internal/entry/tui/panels_outline.go` — Kiểm tra + cập nhật
- `app.go` — Gán TotalRealWordCount
- `internal/diag/diag.go` — Cập nhật diagnostics

## Test Criteria
- [ ] Sidebar hiển thị 2 dòng riêng biệt: "Số ký tự: X" và "Số chữ: Y"
- [ ] UISnapshot chứa đúng cả 2 giá trị
- [ ] Diagnostics report đúng word count

---
Next Phase: phase-05-test.md
