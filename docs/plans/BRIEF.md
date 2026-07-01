# 💡 BRIEF: Việt hoá Giao diện & Tài liệu ainovel-cli

**Ngày tạo:** 01/07/2026
**Mục tiêu:** Bản địa hoá công cụ ainovel-cli sang tiếng Việt, giúp người dùng Việt Nam dễ dàng tiếp cận, cài đặt và thao tác với ứng dụng.

---

## 1. VẤN ĐỀ CẦN GIẢI QUYẾT
Dự án ainovel-cli là một công cụ rất mạnh mẽ để viết tiểu thuyết tự động, nhưng hiện tại giao diện dòng lệnh (CLI), TUI và tài liệu hướng dẫn hoàn toàn bằng tiếng Trung. Điều này tạo rào cản lớn cho người dùng Việt Nam khi muốn cài đặt, đọc hiểu log hệ thống, hay thao tác.

## 2. GIẢI PHÁP ĐỀ XUẤT
Thực hiện "Việt hoá" các thành phần hiển thị với người dùng (User-facing) mà không can thiệp vào core prompts của AI, nhằm giữ nguyên tính logic của engine nhưng tạo giao diện thân thiện cho người Việt.

## 3. ĐỐI TƯỢNG SỬ DỤNG
- **Primary:** Các tác giả, người đam mê tiểu thuyết AI tại Việt Nam muốn tự cài đặt và sử dụng tool.
- **Secondary:** Developers Việt Nam muốn tìm hiểu và đóng góp.

## 4. PHẠM VI VIỆT HOÁ (SCOPE)

### 🚀 MVP (Bắt buộc làm):
- [ ] **Việt hoá Giao diện (TUI/CLI):**
  - Dịch các thông báo hiển thị trên terminal (stdout/stderr).
  - Dịch các menu, nhãn (labels), trạng thái trong giao diện dòng lệnh (TUI).
  - Dịch các thông báo lỗi (Error messages) để người dùng biết cách xử lý.
  - Dịch Help texts (`-h`, `--help`) của các lệnh CLI.
- [ ] **Việt hoá Tài liệu:**
  - Dịch toàn bộ nội dung file `README.md` sang tiếng Việt (giữ nguyên format, markdown, link hình ảnh, cấu trúc).

### 🎁 Phase 2 (NICE-TO-HAVE - Có thể làm sau):
- [ ] Dịch các comment trong file `config.example.jsonc`.
- [ ] Dịch các file tài liệu khác trong thư mục `docs/` (nếu có).

## 5. ƯỚC TÍNH SƠ BỘ (Technical Reality Check)
- **Độ phức tạp:** 🟡 TRUNG BÌNH. 
  - File `README.md` khá dài, cần dịch chau chuốt. 
  - Giao diện CLI nằm rải rác trong các file mã nguồn Go (`cmd/`, `internal/`). Cần phải quét các file `.go` để tìm và thay thế chuỗi text (strings) tiếng Trung sang tiếng Việt một cách cẩn thận để không làm hỏng logic code.
- **Rủi ro kỹ thuật:** Việc dịch text trong source code có thể làm ảnh hưởng đến độ dài của giao diện TUI (tiếng Việt thường dài hơn tiếng Trung), có thể cần tinh chỉnh lại một chút về UI layout nếu bị tràn chữ.

## 6. BƯỚC TIẾP THEO
→ Chạy `/plan` để AI lên kế hoạch cụ thể (liệt kê các file cần sửa) và chia phase thực hiện!
