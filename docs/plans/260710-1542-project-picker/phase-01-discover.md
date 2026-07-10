# Phase 01: Scan & Discover Projects
Status: ✅ Complete

## Objective
Tạo module cho phép quét thư mục `output/` để lấy danh sách các dự án truyện hiện tại (đọc `meta/progress.json`).

## Requirements
### Functional
- [x] Scan thư mục `output/` tìm các subfolder.
- [x] Chỉ nhận diện subfolder nào chứa `meta/progress.json` hợp lệ.
- [x] Trích xuất Tên truyện, Số chữ (RealWordCount), Số chương, Ngày cập nhật.
- [x] Sắp xếp danh sách theo ngày cập nhật mới nhất.

## Implementation Steps
1. [x] Tạo struct `ProjectInfo` trong `internal/store` hoặc `host` để chứa thông tin dự án.
2. [x] Viết hàm `DiscoverProjects(outputDir string) ([]ProjectInfo, error)` để thực hiện việc quét.
3. [x] Viết Unit Test cho `DiscoverProjects`.

## Files to Create/Modify
- `internal/store/discover.go` (hoặc vị trí phù hợp) - Code lấy danh sách.
- `internal/store/discover_test.go` - Test case.

## Test Criteria
- [x] Quét được đúng 2 dự án đang có trong `output/`.
- [x] Không bị lỗi nếu có thư mục rác (không có progress.json).

---
Next Phase: [Phase 02: TUI Project Picker Screen](phase-02-picker-ui.md)
