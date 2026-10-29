---
artifact_type: ticket
id: TICKET-09-02
status: draft
owner: human
parent: TICKET-09
trace:
  parent: TICKET-09-tro-ly-ai.md
  guide: README.md
---

# TICKET-09-02 — Hỏi đáp tài chính cá nhân

Ticket lớn: [Trợ lý nhập liệu và tư vấn](TICKET-09-tro-ly-ai.md).

Thiết kế và kế hoạch đề xuất: [DESIGN-09-AI](TICKET-09-DETAIL_DESIGN.md), T6/T7; chờ owner review, chưa triển khai.

2026-09-22: [AI-ADVISOR-01](AI-ADVISOR-01-DETAIL_DESIGN.md) và [plan V1](../../superpowers/plans/2026-09-22-finance-assistant-v1.md) cập nhật thiết kế advisor theo yêu cầu tab riêng, một history/account, tool calling và cards. V1 vẫn chỉ đọc; write proposals chỉ ở V2 có xác nhận, không đổi tiêu chí read-only dưới đây. Các miền chưa hoàn chỉnh phải báo unsupported; không coi danh sách phạm vi là năng lực đã có.

## Mục tiêu và phạm vi

Hỏi trợ lý về giao dịch, reports, budget, hũ, nợ và tài sản.

## Tiêu chí nghiệm thu

- [ ] Trả lời dùng dữ liệu đúng account, có kỳ/phạm vi và nguồn cho con số.
- [ ] Thiếu dữ liệu thì nói rõ, không dùng hội thoại cũ thay số hiện tại.
- [ ] Tư vấn chỉ đọc, không tự sửa tiền, xác nhận giao dịch hoặc thay note user.

Ghi chú: Tổng kết AI tự động theo tháng thuộc TICKET-07-04.
