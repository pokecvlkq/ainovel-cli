# ainovel-cli

Engine sáng tác tiểu thuyết dài kỳ tự động bằng AI. Coordinator điều khiển 3 sub-agents (Architect / Writer / Editor) trong một Prompt duy nhất để hoàn thành toàn bộ cuốn sách. Host chỉ làm nhiệm vụ khởi động, khôi phục và quan sát. Từ một câu yêu cầu đến tiểu thuyết hoàn chỉnh, toàn bộ quá trình không cần sự can thiệp của con người.

<p align="center">
  <img src="scripts/sample.gif" alt="ainovel-cli demo" width="800">
  <img src="scripts/novel.png" alt="ainovel-cli bg" width="800">
</p>

## Tính năng

- **Đa đặc vụ phối hợp (Multi-Agent)** — Coordinator điều phối Architect / Writer / Editor trong một vòng lặp dài, tự chủ quyết định quy trình sáng tác.
- **LLM điều khiển vòng lặp dài** — Viết cả cuốn sách trong một Prompt duy nhất, Host không can thiệp điều phối. Càng đơn giản càng ổn định, loại bỏ các luồng biên đạo phức tạp.
- **Khôi phục cấp độ Bước (Step)** — Ghi checkpoint sau mỗi lần thực thi công cụ thành công, khôi phục chính xác từng bước plan/draft/check/commit sau khi bị gián đoạn.
- **Lập kế hoạch cuộn hai lớp Quyển-Phần** — Không lập dàn ý toàn bộ chương ngay từ đầu. Ban đầu chỉ lập khung 2 quyển đầu + các chương chi tiết của phần 1, các phần/quyển sau sẽ được Architect mở rộng dần khi viết tới, luôn tham khảo tóm tắt trước đó và trạng thái nhân vật.
- **Đề xuất chương liên quan thông minh** — Tự động đề xuất các chương lịch sử liên quan từ 4 chiều: phục bút, sự xuất hiện của nhân vật, thay đổi trạng thái, và mối quan hệ để đảm bảo tính liên tục của bộ truyện dài 500+ chương.
- **Chiến lược ngữ cảnh tự thích ứng** — Tự động chuyển đổi giữa toàn văn / cửa sổ trượt / tóm tắt phân tầng dựa trên số chương, hỗ trợ truyện dài 500+ chương.
- **Đánh giá chất lượng 7 chiều** — Editor đánh giá từ: tính nhất quán, hành vi nhân vật, nhịp độ, mạch truyện, phục bút, điểm nhấn (hook), và chất lượng thẩm mỹ.
- **Can thiệp thời gian thực** — Nhập ý kiến sửa đổi trực tiếp vào ô nhập liệu khi đang viết (không cần tạm dừng), hệ thống tự đánh giá phạm vi ảnh hưởng và viết lại các chương liên quan.
- **Giao diện TUI thống nhất** — Theo dõi tiến độ realtime, hỗ trợ khởi chạy ngay với một câu yêu cầu.
- **Hỗ trợ nhiều LLM** — Chuyển đổi dễ dàng giữa OpenRouter / Anthropic / Gemini / OpenAI / DeepSeek...

## Kiến trúc

Thiết kế cốt lõi: **LLM điều khiển, Host phục vụ**. Coordinator tự quyết định toàn bộ quy trình trong một lần Run. Host chỉ khởi động, khôi phục và quan sát.

```
┌─────────────────────────────────────────────────┐
│                Host (Vỏ bọc mỏng)                 │
│      Khởi động / Khôi phục / Theo dõi / Can thiệp   │
└──────────────────────┬──────────────────────────┘
                       │ Một Prompt duy nhất
┌──────────────────────▼──────────────────────────┐
│              Coordinator (LLM Long Loop)          │
│ Đọc novel_context → Gọi sub-agent → Đọc kết quả → Tiếp│
└────┬──────────┬──────────┬──────────────────────┘
     │          │          │
 ┌───▼────┐ ┌───▼───┐ ┌────▼────┐
 │Architect│ │Writer │ │ Editor  │
 └───┬────┘ └───┬───┘ └────┬────┘
     └──────────┼──────────┘
                │ Gọi Tools (IO + checkpoint)
┌───────────────▼─────────────────────────────────┐
│                   Store                         │
│  Progress / Checkpoint / Outline / Drafts / ... │
└─────────────────────────────────────────────────┘
```

- **Host** — Khởi động Coordinator, khôi phục sự cố, hiển thị cho TUI. Không đưa ra quyết định điều phối.
- **Coordinator** — Người quyết định duy nhất, điều khiển toàn bộ quá trình lập kế hoạch → viết → duyệt → tóm tắt.
- **SubAgents** — Architect / Writer / Editor hoạt động với context độc lập, tương tác qua các công cụ trong Store.
- **Tools** — Công cụ thực thi IO + ghi checkpoint, chỉ trả về JSON, không chứa lệnh ẩn.

### Trách nhiệm của Agent

| Agent | Trách nhiệm | Công cụ |
|--------|------|------|
| **Coordinator** | Điều phối toàn cục, xử lý quyết định đánh giá và can thiệp của người dùng | `subagent` `novel_context` |
| **Architect** | Tạo tiền đề, đề cương, hồ sơ nhân vật, quy tắc thế giới | `novel_context` `save_foundation` |
| **Writer** | Tự động hoàn thành dàn ý, viết, tự kiểm duyệt và nộp chương | `novel_context` `read_chapter` `plan_chapter` `draft_chapter` `check_consistency` `commit_chapter` |
| **Editor** | Đọc bản gốc, duyệt từ góc độ cấu trúc và thẩm mỹ | `novel_context` `read_chapter` `save_review` `save_arc_summary` `save_volume_summary` |

### Quy trình Sáng tác

```
Yêu cầu của bạn → Architect lên khung + phần 1 → Writer viết từng chương → Editor duyệt phần
                                                   ↑                   │
                                                   ├── Viết lại/Sửa ◄──┘
                                                   │
                                            Architect mở rộng phần/quyển sau
                                           (Tham khảo tóm tắt + trạng thái NV)
```

Writer hoàn thành mỗi chương theo thứ tự cố định (Nội dung tự chủ, nhưng thứ tự công cụ là bắt buộc):

1. `novel_context` — Tải ngữ cảnh (tóm tắt, phục bút, trạng thái nhân vật, gợi ý).
2. `read_chapter` — Đọc lại phần trước để lấy nhịp độ.
3. `plan_chapter` — Lên ý tưởng mục tiêu chương, xung đột, nhịp cảm xúc.
4. `draft_chapter` — Viết bản nháp toàn bộ chương.
5. `check_consistency` — Kiểm tra tính nhất quán (phải làm sau draft).
6. `commit_chapter` — Nộp bản cuối, trả về kết quả cho Reminder điều khiển tiếp.

### Quy tắc Chuyển đổi Trạng thái

Hệ thống chia trạng thái chạy thành 2 lớp:

- **Phase** — Giai đoạn lớn (đang thiết lập, đang viết, hay đã hoàn thành)
- **Flow** — Luồng hoạt động hiện tại (viết bình thường, duyệt, viết lại, gọt giũa hay xử lý can thiệp)

#### Phase

Quy tắc "chỉ tiến không lùi":

```text
init -> premise -> outline -> writing -> complete
  \-------> outline ------^
  \--------------> writing
```

Ý nghĩa:

- `init` — Đã tạo task, chưa có thiết lập ổn định.
- `premise` — Đã lưu tiền đề câu chuyện.
- `outline` — Đã lưu đề cương, có thể bắt đầu viết.
- `writing` — Đang trong thời kỳ viết chương.
- `complete` — Hoàn tất toàn bộ sách.

#### Flow

Chỉ mô tả luồng hoạt động trong giai đoạn viết, cho phép chuyển đổi:

```text
writing   -> reviewing / rewriting / polishing / steering / writing
reviewing -> writing / rewriting / polishing / steering / reviewing
rewriting -> writing / steering / rewriting
polishing -> writing / steering / polishing
steering  -> writing / reviewing / rewriting / polishing / steering
```

Ý nghĩa:

- `writing` — Viết chương tiếp theo.
- `reviewing` — Editor đang duyệt.
- `rewriting` — Xử lý các chương phải viết lại.
- `polishing` — Xử lý các chương cần gọt giũa.
- `steering` — Đang đánh giá và xử lý can thiệp của người dùng.

### Lập kế hoạch Cuộn cho Truyện dài

Giải pháp truyền thống lập kế hoạch tất cả một lần dễ làm đề cương rỗng. Hệ thống dùng **La bàn + Lập kế hoạch theo tầm nhìn**, mô phỏng quá trình của tác giả thật:

```
Lập kế hoạch ban đầu       Khi kết thúc Phần              Khi kết thúc Quyển
┌────────────────────┐    ┌─────────────────────┐    ┌─────────────────────┐
│ Hướng đi (La bàn)  │    │ Editor duyệt phần   │    │ Editor duyệt quyển  │
│ Bắt đầu 2 quyển    │    │ Tóm tắt phần + NV   │    │ Tóm tắt quyển       │
│ Chương chi tiết p1 │ →  │ Architect mở phần 2 │ →  │ Architect tự tạo    │
│ Nhân vật + TG      │    │ Writer tiếp tục viết│    │ Quyển mới + La bàn  │
└────────────────────┘    └─────────────────────┘    └─────────────────────┘
```

- **La bàn (Compass)** — Hướng kết cục + tuyến dài hạn, được Architect cập nhật mỗi khi kết thúc quyển.
- **Tạo theo nhu cầu** — Viết xong quyển hiện tại, Architect tự tạo quyển tiếp theo dựa trên nội dung đã viết.
- **Khung phần (Arc)** — Chỉ có mục tiêu + ước lượng số chương. Tới nơi mới mở rộng chi tiết.
- **Tinh chỉnh dần dần** — Càng viết về sau càng chính xác nhờ tham khảo tóm tắt trước đó.

### Quản lý Ngữ cảnh Truyện dài

Tiểu thuyết 500+ chương dùng tóm tắt 3 cấp độ + 4 cấp độ nén + đề xuất thông minh:

```
Quyển (Volume) → Tóm tắt quyển
└── Phần (Arc) → Tóm tắt phần + Ảnh chụp NV + Quy tắc phong cách
    └── Chương (Chapter) → Tóm tắt chương (cửa sổ trượt 3 chương gần nhất)
```

- **Tóm tắt phân tầng** — Gần dùng tóm tắt chương, xa dùng tóm tắt quyển.
- **Đề xuất chương liên quan** — Tự tìm lịch sử từ phục bút, xuất hiện NV, thay đổi trạng thái để Writer đọc lại.
- **Dự báo chương tiếp theo** — Load đề cương chương tới giúp Writer làm mồi nhử (hook) cuối chương.

#### Đường ống Nén Ngữ cảnh

Khi hội thoại vượt quá giới hạn Context Window, sẽ nén theo từng cấp độ:

```
ToolResultMicrocompact → LightTrim → StoreSummaryCompact → FullSummary
```

- **StoreSummaryCompact** — Dùng tóm tắt có sẵn trong store thay thế cho tin nhắn cũ, không tốn token LLM.
- **FullSummary** — Bắt LLM tóm tắt lại nếu thực sự cạn kiệt context, giữ lại trạng thái nhân vật và phục bút.
- **Bơm lại dữ liệu** — Sau khi nén, bơm lại kế hoạch chương hiện tại để Writer không bị "mất trí nhớ".
- **Giao diện TUI** — Thanh sức khỏe ngữ cảnh: Xanh (<70%) → Vàng (70-85%) → Đỏ (>85%).

## Bắt đầu nhanh

Dự án cung cấp 2 phiên bản: **CLI (Giao diện dòng lệnh)** và **GUI (Giao diện Web App thân thiện)**.

### 1. Phiên bản GUI (Khuyên dùng)
Giao diện trực quan, dễ thao tác, hỗ trợ Dark mode, Monaco Editor, và giám sát Agent theo thời gian thực.
- **Cách sử dụng:** Khởi động file `ainovel-gui.exe`. Giao diện Web UI sẽ tự động mở lên, cho phép bạn thiết lập API Key, chọn Model và bắt đầu sáng tác một dự án tiểu thuyết mới.
- **Tải sẵn:** Tải file `ainovel-gui.exe` (Windows) hoặc bản build tương ứng tại trang [Releases](https://github.com/voocel/ainovel-cli/releases/latest).
- **Build từ source:** Cần cài đặt [Go](https://go.dev/) và [Node.js](https://nodejs.org/).
  ```bash
  # Cài đặt Wails CLI
  go install github.com/wailsapp/wails/v2/cmd/wails@latest
  
  # Build ứng dụng GUI (Windows)
  wails build
  ```

### 2. Phiên bản CLI
Phù hợp chạy nền (headless), trên server, hoặc qua Docker.
```bash
# Cài đặt tự động (macOS / Linux, không cần Go)
curl -fsSL https://raw.githubusercontent.com/voocel/ainovel-cli/main/scripts/install.sh | sh

# Cài đặt qua Go
go install github.com/voocel/ainovel-cli/cmd/ainovel-cli@latest

# Chạy lần đầu tiên để thiết lập API Key và Model
ainovel-cli
```

> Windows hoặc cài đặt thủ công: Vui lòng tải phiên bản từ [Releases](https://github.com/voocel/ainovel-cli/releases/latest).

### Docker

Dùng Docker chạy nền hoặc hiển thị TUI:

```bash
mkdir -p config workspace

# TUI
docker run --rm -it \
  -v "$PWD/config:/root/.ainovel" \
  -v "$PWD/workspace:/workspace" \
  ghcr.io/voocel/ainovel-cli:latest

# Chạy ngầm (Headless)
docker run --rm \
  -v "$PWD/config:/root/.ainovel" \
  -v "$PWD/workspace:/workspace" \
  ghcr.io/voocel/ainovel-cli:latest \
  --headless --prompt "Viết một cuốn tiểu thuyết huyền huyễn phương Đông"
```

### Quản lý nhiều cuốn tiểu thuyết

Mỗi thư mục đại diện cho một cuốn tiểu thuyết. Chạy lệnh ở thư mục nào thì truyện sẽ được lưu vào `output/novel/` của thư mục đó. `cd` về thư mục và gõ `ainovel-cli` để tiếp tục viết từ Checkpoint.

### File Cấu hình

Khi chạy lần đầu, ứng dụng sẽ hướng dẫn bạn tạo `~/.ainovel/config.json`.
Bạn có thể thiết lập nhiều Provider (OpenRouter, OpenAI, Gemini...) và chỉ định các model khác nhau cho Architect, Writer, và Editor.

```jsonc
{
  "provider": "openrouter",
  "model": "google/gemini-2.5-flash",
  "reasoning_effort": "medium",
  "providers": {
    "openrouter": {
      "api_key": "sk-or-v1-xxx",
      "base_url": "https://openrouter.ai/api/v1",
      "models": ["google/gemini-2.5-flash", "google/gemini-2.5-pro"]
    }
  },
  "style": "default"
}
```

#### Phân chia vai trò (Roles)

Bạn có thể để Writer dùng Claude 3.5 Sonnet, nhưng Architect dùng Gemini 2.5 Pro để tiết kiệm chi phí:

```jsonc
{
  "roles": {
    "writer": { "provider": "anthropic", "model": "claude-sonnet-4", "reasoning_effort": "high" },
    "architect": { "provider": "openrouter", "model": "google/gemini-2.5-pro", "reasoning_effort": "low" }
  }
}
```

### Phong cách viết

Sửa cấu hình `style`:
- `default` — Mặc định
- `suspense` — Hồi hộp, trinh thám
- `fantasy` — Giả tưởng, tiên hiệp
- `romance` — Ngôn tình

### Quy tắc loại bỏ văn phong AI

Hệ thống có sẵn bộ lọc từ khóa "công nghiệp" của AI (vd: "Không thể phủ nhận", "Nói tóm lại", "Ánh mắt kiên định"). Bạn có thể thêm file Markdown bất kỳ vào `~/.ainovel/rules/` để yêu cầu thêm các quy tắc cá nhân hóa. Ví dụ: Tạo file `my-rules.md` ghi "Nhân vật chính không được quá thánh mẫu, viết văn dùng nhiều từ chỉ cảm giác". AI sẽ tự động phân tích và áp dụng.

## Báo cáo Chẩn đoán

Gõ `/diag` trong TUI để quét lỗi toàn bộ truyện:
- **Luồng** — Kẹt lặp lại, chương bị nhảy số.
- **Chất lượng** — Điểm đánh giá thấp liên tục.
- **Kế hoạch** — Phục bút bị quên, La bàn lỗi thời.
- **Ngữ cảnh** — Nhân vật mất tích, lỗi timeline.

## Hồ sơ mô phỏng văn phong (Simulate)

Đặt một file `.txt` chứa văn phong mẫu vào thư mục `simulate/`, gõ `/simulate`. AI sẽ phân tích và trích xuất cấu trúc câu, nhịp điệu, cách dùng từ để bắt chước viết theo phong cách đó, lưu tại `output/novel/meta/simulation_profile.json`.

## Nhập tiểu thuyết (Import)

Nhập một cuốn tiểu thuyết có sẵn để viết tiếp:
```
/import ~/tieuthuyet.txt
```
AI sẽ tự động đọc, phân chia chương, trích xuất nhân vật, làm tóm tắt và tiếp tục viết phần tiếp theo. 

## Xuất tiểu thuyết (Export)

Gõ `/export` trong TUI để xuất toàn bộ nội dung đã viết ra TXT hoặc EPUB.
```
/export ~/TruyenCuaToi.epub
```

## Can thiệp thời gian thực (Steer)

Nhập ý kiến thẳng vào TUI khi AI đang viết:
```
❯ Đẩy nhanh tuyến tình cảm lên chương 4 nhé
```
Hệ thống sẽ lưu lại và tự động yêu cầu AI đánh giá, cập nhật kịch bản và sửa lại các chương đã viết nếu cần thiết.

## Cấu trúc thư mục Output

Mọi dữ liệu được lưu vào `output/`:

```
output/{novel_name}/
├── chapters/           # Bản chính thức (Markdown)
├── summaries/          # Tóm tắt chương (JSON)
├── drafts/             # Bản nháp
├── reviews/            # Đánh giá của Editor
├── meta/
│   ├── premise.md      # Tiền đề
│   ├── outline.json    # Đề cương phẳng
│   ├── layered_outline.json # Đề cương phân tầng
│   ├── compass.json    # La bàn định hướng
│   ├── characters.json # Hồ sơ nhân vật
│   ├── world_rules.json# Quy tắc thế giới
│   ├── checkpoints.jsonl # Checkpoint khôi phục
│   └── characters.md   # Hồ sơ nhân vật bản dễ đọc
```

## Giấy phép (License)

MIT License.
Dự án được xây dựng với Go 1.25, Agentcore, LiteLLM và Bubble Tea.
