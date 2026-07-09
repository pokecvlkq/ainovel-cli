# Phase 03: Application Layer (Tools)
Status: ⬜ Pending
Dependencies: Phase 02

## Objective
Cập nhật các tool (draft, commit, edit, read, check) để sử dụng bộ đếm "chữ" mới, đồng thời vẫn truyền đúng cả 2 loại count vào storage.

## Implementation Steps

### 3.1 draft_chapter.go
1. [ ] Thay `utf8.RuneCountInString(full)` → `domain.WordCount(full)` trong phần báo cáo `word_count` trả về cho AI
2. [ ] Tương tự cho append mode: `utf8.RuneCountInString(a.Content)` → `domain.WordCount(a.Content)`
3. [ ] AI nhận "word_count" theo chữ thật → tự điều chỉnh viết dài hơn

### 3.2 commit_chapter.go
4. [ ] Cập nhật tất cả call sites `LoadChapterContent` nhận thêm giá trị `wordCount` (chữ) bên cạnh `charCount` (ký tự)
5. [ ] Cập nhật call sites `MarkChapterComplete` truyền thêm `wordCount`
6. [ ] Cập nhật `CommitResult.WordCount` → dùng word count (chữ) thay vì char count

### 3.3 edit_chapter.go
7. [ ] Cập nhật `LoadDraft` / `LoadChapterContent` call — nhận thêm word count

### 3.4 check_consistency.go
8. [ ] Cập nhật call `LoadChapterContent` nhận thêm giá trị word count

### 3.5 rules/checker.go
9. [ ] Giữ nguyên logic kiểm tra `appendChapterWords` — truyền word count (chữ) thay vì char count vào hàm `Check()`. Các ngưỡng min/max giữ nguyên vì user xác nhận con số từ tiếng Trung → tiếng Việt vẫn phù hợp.

### 3.6 novel_context_builders.go
10. [ ] Cập nhật `total_word_count` trong context signals → dùng `TotalRealWordCount` (chữ)
11. [ ] Cập nhật `LoadChapterContent` call sites

### 3.7 read_chapter.go
12. [ ] Cập nhật `LoadChapterContent` call sites (nếu có)

### 3.8 host/cocreate_stage.go & host/resume.go & host/host.go
13. [ ] Cập nhật tham chiếu `TotalWordCount` → dùng `TotalRealWordCount` trong thông điệp cho user/AI ("khoảng X chữ")

### 3.9 domain/writing.go
14. [ ] Giữ nguyên `CommitResult.WordCount` field name + JSON tag — giá trị bên trong giờ là word count (chữ) thay vì char count

## Files to Create/Modify
- `internal/tools/draft_chapter.go`
- `internal/tools/commit_chapter.go`
- `internal/tools/edit_chapter.go`
- `internal/tools/check_consistency.go`
- `internal/tools/read_chapter.go`
- `internal/tools/novel_context_builders.go`
- `internal/rules/checker.go`
- `internal/host/cocreate_stage.go`
- `internal/host/resume.go`
- `internal/host/host.go`

## Test Criteria
- [ ] `draft_chapter` trả về word_count là số chữ (không phải ký tự)
- [ ] `commit_chapter` ghi đúng cả charCount + wordCount vào progress
- [ ] `checker.Check()` kiểm tra đúng trên word count mới
- [ ] Các context builder gửi đúng `total_word_count` = số chữ cho AI

---
Next Phase: phase-04-ui.md
