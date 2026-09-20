---
artifact_type: ticket
id: TICKET-03-02
status: in_progress
owner: human
parent: TICKET-03
trace:
  parent: TICKET-03-budget.md
  guide: README.md
---

# TICKET-03-02 — Xem tiến độ, cảnh báo và dự báo budget

Ticket lớn: [Theo dõi budget](TICKET-03-budget.md).

## Mục tiêu và phạm vi

Real ledger-derived progress, descendant matching, deduplicated active summary, daily allowance and on-screen threshold notices implemented in [API-SCREENS-01](API-SCREENS-01-DETAIL_DESIGN.md). Forecasts, notification delivery and full owner acceptance remain pending.

Xem đã chi/còn lại, cảnh báo, dự báo và giao dịch của budget.

## Tiêu chí nghiệm thu

- [ ] Giao dịch đúng ví/danh mục/kỳ tự vào tiến độ; danh mục cha/con không đếm trùng.
- [ ] Có cảnh báo từ 75%, vượt mức, gợi ý chi/ngày và dự báo có nhãn ước tính.
- [ ] Sửa giao dịch nguồn cập nhật tiến độ, kể cả kỳ kết thúc.
