# Phase 03: Auto Rename Workspace
Status: ? Pending
Dependencies: Phase 02

## Objective
T? d?ng d?t tên thu m?c theo tên truy?n khi t?o d? án m?i.

## Requirements
### Functional
- [ ] Khi t?o d? án m?i, m?c d?nh dùng m?t tên thu m?c ch?a timestamp ho?c ID ng?u nhiên (ví d? `novel-202607101542`).
- [ ] N?u quá trình nh?p li?u "B?t d?u nhanh" cung c?p luôn tên truy?n, có th? dùng tên dó ngay. Tuy nhiên thu?ng AI s? sinh ra tên ? bu?c Foundation.
- [ ] Sau khi AI sinh xong Foundation và có `NovelName`, g?i tính nang d?i tên thu m?c.
- [ ] Ho?c cho phép user Enter d? t? d?ng dùng tên sinh ra, n?u không t? d?i tên. (Theo yêu c?u m?i: t? d?t theo tên truy?n, n?u user không can thi?p).

## Implementation Steps
1. [ ] S?a `OutputDir` m?c d?nh khi kh?i t?o d? án m?i thành tên duy nh?t thay vì `novel`.
2. [ ] Implement hàm d?i tên thu m?c `RenameWorkspace` (dã có ho?c c?n thêm) d? h? tr? vi?c thay d?i tên an toàn mà không làm h?ng ti?n trình (chú ý du?ng d?n file dang m? n?u có).
3. [ ] C?p nh?t `internal/host/host.go` (ho?c `runtime`) d? th?c hi?n bu?c d?i tên này sau khi xác d?nh du?c `NovelName`.

## Files to Create/Modify
- `internal/entry/tui/model_update.go` (logic t?o thu m?c m?i).
- `internal/host/host.go` ho?c các file x? lý liên quan.

## Test Criteria
- [ ] Thu m?c không b? d?ng d? (conflict) khi t?o nhi?u d? án.
- [ ] Sau khi vi?t xong bu?c d?u, thu m?c mang dúng tên truy?n (ví d? `novel-Tien-Do-Thieu-Nu`).

---
End of Plan
