---
artifact_type: ticket
id: TICKET-03-01
status: in_progress
owner: human
parent: TICKET-03
trace:
  parent: TICKET-03-budget.md
  guide: README.md
---

# TICKET-03-01 — Thiết lập và quản lý budget

Ticket lớn: [Theo dõi budget](TICKET-03-budget.md).

## Mục tiêu và phạm vi

Explicit-interval CRUD implemented and API-verified in [API-SCREENS-01](API-SCREENS-01-DETAIL_DESIGN.md), with overlap protection and ended-period read/delete only. Recurring/fixed-period choices and owner acceptance remain pending; full ticket not complete.

Tạo/sửa/xóa budget theo ví, danh mục, kỳ và lựa chọn lặp.

## Tiêu chí nghiệm thu

- [ ] Chọn được tuần, tháng, quý, năm hoặc khoảng ngày tùy chỉnh.
- [ ] Kỳ cố định có lựa chọn lặp; khoảng ngày tùy chỉnh không tự lặp.
- [ ] Không tạo budget trùng phạm vi và thời gian chồng lấn; budget kết thúc xem/xóa được.
