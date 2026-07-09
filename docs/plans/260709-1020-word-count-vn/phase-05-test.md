# Phase 05: Test & Verify
Status: ⬜ Pending
Dependencies: Phase 04

## Objective
Đảm bảo toàn bộ thay đổi không làm hỏng build và logic đếm đúng.

## Implementation Steps

### 5.1 Fix các test bị ảnh hưởng bởi thay đổi signature
1. [ ] `internal/store/progress_test.go` — cập nhật tất cả call `MarkChapterComplete` thêm param wordCount
2. [ ] `internal/tools/commit_chapter_test.go` — tương tự
3. [ ] `internal/tools/draft_chapter_test.go` — tương tự
4. [ ] `internal/tools/edit_chapter_test.go` — tương tự
5. [ ] `internal/tools/read_draft_test.go` — cập nhật `LoadChapterContent` nhận thêm return value
6. [ ] `internal/host/exp/exporter_test.go` — tương tự
7. [ ] `internal/host/imp/analyzer_test.go` & `runner_test.go` — tương tự
8. [ ] `internal/tools/reopen_book_test.go` — tương tự
9. [ ] `internal/tools/save_foundation_test.go` — tương tự
10. [ ] `internal/tools/novel_context_test.go` — tương tự
11. [ ] `internal/rules/checker_test.go` — kiểm tra assertion vẫn đúng

### 5.2 Chạy toàn bộ test suite
12. [ ] Chạy `go build ./...` — đảm bảo compile thành công
13. [ ] Chạy `go test ./...` — đảm bảo tất cả test pass
14. [ ] Chạy `go vet ./...` — kiểm tra code quality

### 5.3 Manual Verification
15. [ ] Chạy app (`go run .`) và kiểm tra sidebar hiển thị đúng 2 dòng
16. [ ] Kiểm tra file `progress.json` sau commit chapter → có cả 2 trường

## Files to Create/Modify
- Tất cả `*_test.go` files liệt kê ở trên

## Test Criteria
- [ ] `go build ./...` → thành công
- [ ] `go test ./...` → 100% pass
- [ ] `go vet ./...` → không có warning

---
End of phases.
