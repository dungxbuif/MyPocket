---
artifact_type: ticket
id: TICKET-04-02
status: in_progress
owner: human
parent: TICKET-04
trace:
  parent: TICKET-04-hu-chi-tieu.md
  guide: README.md
  detail_design: CORE-03-TIME-JARS-MONTH-DETAIL_DESIGN.md
  validation: ../VALIDATION_MATRIX.md
---

# TICKET-04-02 — Gắn hũ và xem chi tiêu tháng

Ticket lớn: [Theo dõi chi tiêu bằng hũ](TICKET-04-hu-chi-tieu.md).

## Mục tiêu và phạm vi

Gắn/gỡ hũ cho chi thường; xem tổng chi, giao dịch và mức tham khảo trong tháng.

## Tiêu chí nghiệm thu

- [ ] Một chi thường có tối đa một hũ; chưa phân hũ vẫn ghi nhận được.
- [ ] Có chi mà chưa có mức phân bổ vẫn có report, không coi mức chưa đặt là 0.
- [ ] Chi vượt mức vẫn lưu được và có cảnh báo; sửa/gỡ giao dịch cập nhật thống kê.
