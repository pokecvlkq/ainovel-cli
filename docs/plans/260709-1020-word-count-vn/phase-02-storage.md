# Phase 02: Storage Layer (Progress + Drafts)
Status: ⬜ Pending
Dependencies: Phase 01

## Objective
Cập nhật lớp lưu trữ để ghi nhận song song cả "ký tự" (char count) và "chữ" (word count).

## Implementation Steps

### 2.1 Struct Progress (`internal/domain/runtime.go`)
1. [ ] Giữ nguyên `TotalWordCount` + `ChapterWordCounts` (JSON tag không đổi) — bây giờ đây là bộ đếm **ký tự** (backward-compatible với dữ liệu cũ)
2. [ ] Thêm trường `TotalRealWordCount int` (JSON: `total_real_word_count`)
3. [ ] Thêm trường `ChapterRealWordCounts map[int]int` (JSON: `chapter_real_word_counts,omitempty`)

### 2.2 ProgressStore (`internal/store/progress.go`)
4. [ ] Cập nhật hàm `MarkChapterComplete(chapter, charCount, wordCount int, hookType, dominantStrand string)` — thêm param `wordCount`
5. [ ] Trong hàm đó, cập nhật logic dồn cho cả `TotalWordCount` (char) và `TotalRealWordCount` (word)

### 2.3 DraftStore (`internal/store/drafts.go`)
6. [ ] Cập nhật `LoadChapterContent` trả về `(string, int, int, error)` — thêm giá trị word count
7. [ ] Bên trong dùng cả `domain.CharCount()` và `domain.WordCount()`

## Files to Create/Modify
- `internal/domain/runtime.go` — Thêm trường mới vào struct Progress
- `internal/store/progress.go` — Thêm param + logic cho MarkChapterComplete
- `internal/store/drafts.go` — Cập nhật signature LoadChapterContent

## Backward Compatibility
- JSON cũ không có `total_real_word_count` → Go tự động đặt zero-value (0) khi deserialize.
- `TotalWordCount` giữ nguyên JSON key → dữ liệu cũ vẫn đọc được.

## Test Criteria
- [ ] `MarkChapterComplete` cập nhật đúng cả 2 bộ đếm
- [ ] `LoadChapterContent` trả về 4 giá trị đúng
- [ ] Đọc file progress.json cũ (không có `total_real_word_count`) → không crash

---
Next Phase: phase-03-tools.md
