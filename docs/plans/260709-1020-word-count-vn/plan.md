# Plan: Nâng cấp bộ đếm Ký Tự & Chữ (Vietnamese Word Count)
Created: 26-07-09 10:20
Status: 🟡 Chờ duyệt

## Overview
Hiện tại hệ thống đếm "từ" (`TotalWordCount`) thực chất là đếm **ký tự** (Unicode runes) qua `utf8.RuneCountInString`. Điều này gây ra:
- UI hiển thị "4030 từ" nhưng thực tế là 4030 ký tự → misleading.
- AI được yêu cầu viết N "từ" nhưng đếm theo ký tự → chương quá ngắn.

**Giải pháp:**
1. Đổi label "Số từ" → "Số ký tự" cho bộ đếm cũ (giữ nguyên logic `utf8.RuneCountInString`).
2. Thêm bộ đếm mới "Số chữ" đếm bằng `strings.Fields` (tách bằng khoảng trắng).
3. AI viết chuyện sẽ nhận phản hồi theo "chữ" (word count) để tự điều chỉnh độ dài.

## Tech Stack
- Ngôn ngữ: Go (Golang)
- Không thêm dependency ngoài (dùng stdlib `strings.Fields`)

## Phases

| Phase | Name | Status | Progress |
|-------|------|--------|----------|
| 01 | Domain & Counting Functions | ⬜ Pending | 0% |
| 02 | Storage Layer (Progress + Drafts) | ⬜ Pending | 0% |
| 03 | Application Layer (Tools) | ⬜ Pending | 0% |
| 04 | UI/TUI + Snapshot | ⬜ Pending | 0% |
| 05 | Test & Verify | ⬜ Pending | 0% |

## Quick Commands
- Start Phase 1: `/code phase-01`
- Check progress: `/next`
- Save context: `/save-brain`
