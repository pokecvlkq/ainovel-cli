# Phase 01: Domain & Counting Functions
Status: ⬜ Pending
Dependencies: Không

## Objective
Tạo các hàm đếm rõ ràng trong `domain` package, phân biệt giữa đếm ký tự (rune) và đếm chữ (word).

## Implementation Steps
1. [ ] Đổi tên hàm `WordCount()` → `CharCount()` trong `internal/domain/chapter.go` (logic giữ nguyên `utf8.RuneCountInString`)
2. [ ] Thêm hàm `WordCount()` mới dùng `len(strings.Fields(content))` — đếm "chữ" (tiếng) tách bằng khoảng trắng
3. [ ] Cập nhật comment/docstring cho cả 2 hàm

## Files to Create/Modify
- `internal/domain/chapter.go` — Đổi tên hàm cũ + thêm hàm mới

## Backward Compatibility
- Tất cả caller cũ đang gọi `domain.WordCount()` sẽ cần cập nhật thành `domain.CharCount()`.
- Hàm `domain.WordCount()` mới có cùng signature `(string) int` nhưng logic khác (đếm chữ thay vì ký tự).

## Test Criteria
- [ ] `CharCount("Xin chào")` == 8 (8 runes)
- [ ] `WordCount("Xin chào")` == 2 (2 chữ)
- [ ] `WordCount("")` == 0
- [ ] `WordCount("   ")` == 0

---
Next Phase: phase-02-storage.md
