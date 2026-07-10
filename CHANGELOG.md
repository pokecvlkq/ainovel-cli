# Changelog

Tất cả những thay đổi nổi bật đối với dự án AINovel CLI sẽ được ghi chép tại đây.

## [v1.1.6] - 2026-07-10

### Thêm mới (Added)
- **Tính năng Project Picker (Quản lý nhiều truyện)**:
  - Cho phép người dùng viết và quản lý nhiều dự án truyện độc lập trong thư mục `output/`.
  - Tự động quét các truyện cũ khi khởi động (nếu có) và hiển thị màn hình chọn dự án.
  - Hỗ trợ tuỳ chỉnh tên thư mục truyện: Ấn `N` để tạo mới, sau đó nhập tên mong muốn hoặc ấn `Enter` để tự động tạo theo timestamp, tránh ghi đè dữ liệu.
  - Lệnh `/projects`: Hỗ trợ quay về màn hình chọn dự án bất cứ lúc nào ngay từ giao diện chat.
  - Giao diện (TUI) trực quan: Hiển thị tên truyện, số chương, tổng số chữ và thời gian cập nhật. Dễ dàng di chuyển bằng phím mũi tên và chọn bằng `Enter`.

## [v1.1.5] - 2026-07-09

### Thêm mới (Added)
- **Hệ thống đếm ký tự và đếm chữ**:
  - Tách biệt bộ đếm `CharCount` (ký tự) và `WordCount` (chữ) trong `chapter.go`.
  - Cập nhật `domain.Progress` để theo dõi cả `TotalWordCount` (ký tự, giữ nguyên tên để tương thích ngược) và `TotalRealWordCount` (chữ) cùng với `ChapterRealWordCounts`.
  - Cập nhật hệ thống chẩn đoán (Diagnostics) trong `rules_quality.go` và `diag.go` để sử dụng `TotalRealWordCount` cho các thống kê chữ.

## [v1.1.4] - 2026-07-08

### Thêm mới (Added)
- **Tích hợp Google Vertex AI Custom Provider**: 
  - Triển khai Vertex AI làm nhà cung cấp mô hình chính thức (Phương án 2) thông qua thư viện chuẩn `cloud.google.com/go/vertexai/genai`.
  - Hỗ trợ nạp credentials an toàn từ chuỗi JSON Service Account trực tiếp thông qua biến môi trường trong tệp `.env`.
  - Thiết lập cơ chế tự động parse `project_id` trực tiếp từ dữ liệu JSON Service Account, giúp tinh giản cấu hình.
  - Hỗ trợ cấu hình song song nhiều tài khoản Vertex AI (Ví dụ: `Vertex1-aitnd` và `Vertex2-poke`) trỏ tới các biến môi trường khác nhau trong `.env` để làm dự phòng xoay vòng (fallback) khi một tài khoản hết hạn mức (Quota Exceeded).

### Thay đổi (Changed)
- **Cấu hình biến môi trường**: Tích hợp thư viện `github.com/joho/godotenv` vào hàm khởi chạy chính `main.go` để tự động nạp các cài đặt cấu hình từ file `.env` cục bộ.
- **Bảo mật**: Thêm thư mục `credentials/` vào danh sách loại trừ `.gitignore` tránh nguy cơ rò rỉ Service Account Keys lên kho mã nguồn.

### Sửa lỗi (Fixed)
- **Hỗ trợ Agent Tools**: Fix lỗi thiếu hỗ trợ Tools (Function Calling) trong cấu hình VertexModel. Đã thêm cơ chế `convertSchema` để map chính xác dữ liệu của `agentcore` sang định dạng cấu trúc mà `genai.Schema` yêu cầu, ngăn chặn lỗi crash ứng dụng khi các Agent (như Researcher, Writer) gọi các API đọc, tìm kiếm hệ thống qua công cụ ngoài.


## [v1.1.3] - 2026-07-08

### Thêm mới (Added)
- **Hệ thống Quản lý Quota (Quota Tracker)**: Xây dựng cơ chế theo dõi In-Memory Thread-safe với 3 trạng thái: `Active` (Sẵn sàng), `Cooldown` (Nghỉ 60s), và `Dead` (Hết hạn mức).
- **Giao diện Dashboard TUI nâng cao**: 
  - Bổ sung sidebar `Tài khoản` liệt kê trạng thái thời gian thực của toàn bộ API keys.
  - Hiển thị trực quan `⚡ [provider]` dưới tên từng Agent đang hoạt động ở cả sidebar và màn hình chính.

### Thay đổi (Changed)
- **Cải tiến Load Balancing & Fallback**: Các vai trò giờ đây có thể xoay vòng lệch mảng fallbacks (ví dụ Writer quay vòng 01-07, Editor quay vòng 03-07-01-02) để chia đều tải trọng mạng và API limits.
- **Tối ưu xoay vòng API**: Tự động bỏ qua các models/API keys đang bị lỗi Quota hoặc đang Cooldown trong quá trình thử lại (failover) mà không làm ngắt quãng phiên tạo nội dung.

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
