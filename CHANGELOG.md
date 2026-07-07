# Changelog

Tất cả những thay đổi nổi bật đối với dự án AINovel CLI sẽ được ghi chép tại đây.

## [v1.1.2] - 2026-07-07

### Thay đổi (Changed)
- **Sửa lỗi Strict Tool Calling cho Gemini**: Tắt chế độ `StrictSchema` trong `DraftChapterTool` để tránh lỗi biên dịch/gọi API của Gemini khi sử dụng vai trò `writer` chạy trên các model Gemini.
- **Bản địa hóa 100% giao diện TUI/CLI**:
  - Dịch hoàn toàn các chuỗi hiển thị tiếng Trung còn sót lại trong backend lúc chạy (như `设定` -> `Thiết lập`, `本弧` -> `hồi này`, `全局` -> `toàn cục`, `对话` -> `đối thoại`).
  - Việt hóa hoàn toàn các token can thiệp hệ thống (`[用户干预]` -> `[Can thiệp người dùng]`, `[阶段规划]` -> `[Quy hoạch giai đoạn]`) trong cả mã nguồn Go và file cấu hình Prompts để đồng bộ.
  - Sửa nhãn trạng thái và tiến trình khôi phục sáng tác sang tiếng Việt.
- **Biên dịch đồng bộ các file thực thi (Executable Binaries)**:
  - Rebuild lại đồng thời cả `ainovel-cli.exe`, `ainovel-tui.exe` và bản giao diện đồ họa `ainovel-gui.exe` để các thay đổi Việt hóa có hiệu lực.
- **Tối ưu hóa luồng gọi Model cho Writer**:
  - Cấu hình lại `config.json` để vai trò `writer` chạy hết hạn ngạch (quota) của 7 tài khoản Gemini API (`gemini-3.5-flash`) theo thứ tự ưu tiên trước khi fallback về mô hình local Ollama (`qwen3.6:27b`).

### Sửa lỗi (Fixed)
- **Sửa lỗi không xoay key khi hết quota Gemini**: Hàm `pickNextFallback` trong `models.go` giờ nhận diện lỗi quota (`"You exceeded your current quota"`) là failover-eligible, cho phép tự động xoay sang API key/provider tiếp theo. Nguyên nhân gốc: thư viện upstream `agentcore` (1) pattern `"quota exceeded"` không khớp thứ tự từ trong thông báo Gemini, và (2) `ErrProviderQuota` không nằm trong danh sách `IsFailoverEligible`.
- **Sửa lỗi không xoay key khi Gemini quá tải (High Demand)**: Bổ sung nhận diện lỗi quá tải (`"This model is currently experiencing high demand"`, `"Spikes in demand"`) là failover-eligible để tự động xoay sang key khác thay vì dừng cứng.

## [v1.1.1] - 2026-07-06

### Thay đổi (Changed)
- **Bản địa hóa hoàn toàn (Localisation)**: Việt hóa triệt để 100% các file prompts đặc vụ (`coordinator.md`, `architect-short.md`, `architect-long.md`), dọn sạch các từ khóa tiếng Trung còn sót lại để tránh gây nhiễu cho mô hình.
- **Việt hóa giao diện & định dạng xuất bản**:
  - Bản địa hóa toàn bộ tiến trình nhiệm vụ và trạng thái hiển thị trên Terminal/TUI (như "Viết chương 1", "Quyển %d · Hồi %d").
  - Đổi định dạng tiêu đề, mục lục và trang bìa trong file xuất bản (TXT, EPUB) sang tiếng Việt chuẩn, cấu hình thẻ ngôn ngữ EPUB thành `vi-VN`.
- **Tương thích parser**: Bổ sung ánh xạ song ngữ tiêu đề trong `premise_structure.go` giúp AI có thể sinh dàn ý bằng tiếng Việt mà không làm hỏng cú pháp phân tích của Go backend.
- **Cập nhật Regex**: Tối ưu regex phân tích số chương (`chapterRe` và `chapterTaskRe`) để hỗ trợ tốt cả hai định dạng `"第 N 章"` và `"Chương N"`.

## [v1.1.0] - 2026-07-02

### Thêm mới (Added)
- **Giao diện Web App (Wails GUI)**: Ra mắt phiên bản GUI thay thế/bổ sung cho TUI cũ, cung cấp giao diện trực quan và chuyên nghiệp hơn.
- **Frontend React 18 & Tailwind CSS v4**: Xây dựng UI hiện đại với các tính năng Dark Mode, Editor tích hợp Markdown (Monaco), và Split View để Review nội dung.
- **Wails Bindings Layer**: Tích hợp chặt chẽ giữa Go Backend (`host.Host`) và React Frontend thông qua Event Bridge, cho phép theo dõi log, token stream và trạng thái các Agents (Architect, Writer, Editor) theo thời gian thực.
- **Quản lý cấu hình trực quan**: Hỗ trợ thay đổi model, cấu hình API Key và tham số ứng dụng ngay từ giao diện Settings của GUI.

### Thay đổi (Changed)
- **Cấu trúc dự án**: Chuyển đổi kiến trúc sang Wails project (`app.go`, `wails.json`, thư mục `frontend/`), biến TUI cũ thành phiên bản chạy nền (headless) hoặc CLI độc lập.

## [v1.0.0] - 2026-07-02

### Thêm mới (Added)
- **Giao diện TUI (Terminal User Interface)**: Hệ thống giao diện tương tác mới sử dụng `charmbracelet/bubbletea` và `lipgloss`, mang lại trải nghiệm người dùng hiện đại và trực quan hơn hẳn so với phiên bản dòng lệnh cũ.
- **Hệ thống AI Đa Tác Tử (Multi-Agent System)**: Bổ sung cấu trúc làm việc theo nhóm tác tử (Architect, Coordinator, Editor, Writer, v.v...) để xử lý quá trình viết tiểu thuyết có hệ thống.
- **Hệ thống Config tiên tiến**: Hỗ trợ nạp cấu hình qua file `config.yaml` và các thiết lập nhà cung cấp AI linh hoạt (Gemini, Ollama, DeepSeek...).
- **Cơ chế mô hình dự phòng (Fallback)**: Đảm bảo khả năng phục hồi và tính sẵn sàng khi một mô hình AI gặp sự cố (Rate limit, Timeout, v.v...).
- **Ghi log lỗi Startup (Startup Error Log)**: Lưu lỗi chí mạng vào file `last-error.log` nếu giao diện TUI không kịp hiển thị lỗi.

### Thay đổi (Changed)
- **Việt Hóa System Prompts**: Chuyển đổi toàn bộ bộ System Prompts của các tác tử (Architect, Editor, Writer, Coordinator...) từ Tiếng Trung sang Tiếng Việt. Giúp AI duy trì tư duy bằng tiếng Việt, tránh lỗi sinh văn bản và code lẫn lộn Anh/Trung.
- **Tối ưu hóa Code (Refactor)**: 
  - Khắc phục lỗi shallow copy (sao chép nông) của Map cấu hình trong `configfile.go`.
  - Cải tiến thuật toán sắp xếp cảnh báo lỗi tại module `diag.go` để xử lý các cờ mức độ bất định.
  - Sửa đổi vòng lặp fallback trong module `models.go` đảm bảo thử toàn bộ model dự phòng thay vì chỉ model đầu tiên.

### Sửa lỗi (Fixed)
- Sửa lỗi cú pháp biên dịch trong giao diện TUI do thay đổi cấu trúc tham chiếu con trỏ trong các file `model.go`, `model_update.go` và `panels_review.go`.
- Xóa bỏ hoàn toàn các lỗi type assertion và thiếu thư viện khi cập nhật Go module.

### Loại bỏ (Removed)
- Loại bỏ tính năng truyền tham số và viết truyện trực tiếp từ arguments dòng lệnh (CLI flags), bắt buộc người dùng trải nghiệm không gian nhập liệu qua TUI để tối ưu luồng tác tử.
