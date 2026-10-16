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

Explicit-interval CRUD is implemented with overlap protection and ended-period
read/delete only. The 2026-10-16 local candidate also implements canonical
weekly/monthly/quarterly/yearly recurring series with versioned next-period
configuration, stop, idempotency and optional positive carryover. Generated
periods are managed through their series. See the
[design](../../superpowers/specs/2026-10-16-savings-budget-design.md),
[API](../../guides/budget-recurrence-api.md) and
[validation](../VALIDATION_MATRIX.md). Production rollout and owner/device
acceptance remain pending; the full ticket is not marked done.

Tạo/sửa/xóa budget theo ví, danh mục, kỳ và lựa chọn lặp.

## Tiêu chí nghiệm thu

- [ ] Chọn được tuần, tháng, quý, năm hoặc khoảng ngày tùy chỉnh.
- [ ] Kỳ cố định có lựa chọn lặp; khoảng ngày tùy chỉnh không tự lặp.
- [ ] Không tạo budget trùng phạm vi và thời gian chồng lấn; budget kết thúc xem/xóa được.
