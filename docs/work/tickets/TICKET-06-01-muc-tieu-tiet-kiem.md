---
artifact_type: ticket
id: TICKET-06-01
status: in_progress
owner: human
parent: TICKET-06
trace:
  parent: TICKET-06-tiet-kiem-tin-dung-no.md
  guide: README.md
---

# TICKET-06-01 — Theo dõi mục tiêu tiết kiệm

Ticket lớn: [Tiết kiệm, tín dụng và khoản nợ](TICKET-06-tiet-kiem-tin-dung-no.md).

## Mục tiêu và phạm vi

Owner requested implementation from supplied savings screens. [Active design/proof](TICKET-06-01-DETAIL_DESIGN.md). Goal date, summary/history and shared category selection are implemented. Managed-wallet paired transfers are implemented in the shared transaction flow (not fabricated by savings deposit/withdrawal entries); authenticated browser/device UAT remains pending. No unchecked acceptance criterion is silently marked done.

Đặt mục tiêu/hạn cho ví tiết kiệm, nạp/rút và xem tiến độ.

**Implementation status 2026-10-16:** The local candidate adds a dedicated
savings report and paged drilldown. Goal balances/activity and account-wide
surplus are separate, with shared ledger predicates and account-local months.
The configured baseline enters at wallet creation, not before it. See
[API](../../guides/savings-report-api.md),
[design](../../superpowers/specs/2026-10-16-savings-budget-design.md), and
[validation](../VALIDATION_MATRIX.md). Scheduled savings advice, completion-date
forecast, notification fan-out and owner/device acceptance remain pending.

## Tiêu chí nghiệm thu

- [ ] Tiến độ từ số thực trong ví; sửa giao dịch cập nhật tiến độ.
- [ ] Đạt mục tiêu được thể hiện rõ; rút xuống dưới mục tiêu thì tiến độ phản ánh lại.
- [ ] Chuyển giữa các ví không bị tính thu/chi thường lần nữa.
