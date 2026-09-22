---
artifact_type: ticket
id: TICKET-01-01
status: in_progress
owner: human
parent: TICKET-01
trace:
  parent: TICKET-01-tai-khoan-va-vi.md
  guide: README.md
  detail_design: CORE-03-TIME-JARS-MONTH-DETAIL_DESIGN.md
  validation: ../VALIDATION_MATRIX.md
---

# TICKET-01-01 — Thiết lập cá nhân

Ticket lớn: [Tài khoản, ví và danh mục](TICKET-01-tai-khoan-va-vi.md).

## Mục tiêu và phạm vi

Cài đặt múi giờ, cách hiển thị và sử dụng VND.

## Tiêu chí nghiệm thu

- [ ] Các màn hình dùng cùng múi giờ người dùng đã chọn.
- [ ] Giao dịch gần ranh giới ngày/tháng được thống kê đúng kỳ.
- [ ] VND được dùng sẵn; đổi cách hiển thị không làm đổi giá trị tiền.

## Implementation scope

Timezone settings and date-only persistence are implemented under [CORE-03](CORE-03-TIME-JARS-MONTH-DETAIL_DESIGN.md), together with Hũ and monthly overview because all three share account-local month boundaries. Existing accounts remain `Asia/Ho_Chi_Minh`; new accounts initialize from browser IANA timezone once. Automated calendar/API/DB proof passes; owner timezone-change and visual UAT remain pending. See [validation evidence](../VALIDATION_MATRIX.md#core-03-timezone-jars-and-live-monthly-summary--2026-09-22) and proposed [ADR-008](../../decisions/ADR-008-account-timezone-and-calendar-dates.md).
