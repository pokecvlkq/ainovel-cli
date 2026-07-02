# Phase 03: Trình Duyệt & So Sánh (Chapter Reviewer)
Status: ⬜ Pending
Dependencies: phase-02-editor.md

## Objective
Cung cấp giao diện trực quan để người dùng kiểm duyệt nội dung sau khi tác tử Editor sửa bài. Cho phép so sánh giữa bản nháp của Writer và bản hoàn thiện của Editor.

## Requirements
### Functional
- [ ] Tích hợp thư viện Diff (`go-diff`) để tính toán văn bản thêm/bớt.
- [ ] Hiển thị văn bản Diff trực tiếp trong TUI (chia 2 cột Split View hoặc Diff View nội tuyến 1 cột).
- [ ] Footer menu tương tác khi duyệt: `[A] Duyệt (Approve)` | `[R] Yêu cầu viết lại (Reject)` | `[E] Mở Editor sửa thủ công`.

### Non-Functional
- [ ] Màu sắc Diff cần tương phản tốt: Chữ bị xóa bôi ĐỎ gạch ngang, chữ thêm mới bôi XANH. Phù hợp với cả Light/Dark theme.

## Implementation Steps
1. [ ] Thêm thư viện `github.com/sergi/go-diff` vào `go.mod`.
2. [ ] Tạo file mới `panels_review.go` và xây dựng logic Diff: nhận 2 đoạn text, chạy thuật toán tính diff, bọc kết quả diff vào thẻ màu lipgloss (Red/Green).
3. [ ] Tạo View Component để hiển thị kết quả Diff có thể cuộn (viewport).
4. [ ] Cập nhật `model_update.go`:
      - Hiển thị thông báo khi có chương mới cần duyệt.
      - Phím tắt kích hoạt quy trình duyệt bản thảo: `A` (Approve), `R` (Reject).

## Files to Create/Modify
- `go.mod`, `go.sum` - Thêm thư viện `go-diff`.
- `internal/entry/tui/panels_review.go` (Mới) - Model hiển thị Diff và Diff Viewport.
- `internal/entry/tui/model_update.go` - Xử lý logic và phím tắt A/R/E.
- `internal/entry/tui/theme.go` - Định nghĩa màu sắc `colorDiffAdd`, `colorDiffDelete`.

## Test Criteria
- [ ] Khi Editor hoàn thành, chương chuyển sang trạng thái "Chờ Duyệt".
- [ ] Mở màn hình Duyệt hiển thị rõ ràng chữ xanh/đỏ.
- [ ] Nhấn `A` thì chương được đánh dấu hoàn thành, `R` thì Agent Editor/Writer tiếp tục sửa lại.
