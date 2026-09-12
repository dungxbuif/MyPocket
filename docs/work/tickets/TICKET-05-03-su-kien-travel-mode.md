---
artifact_type: ticket
id: TICKET-05-03
status: draft
owner: human
parent: TICKET-05
trace:
  parent: TICKET-05-dinh-ky-du-lich.md
  guide: README.md
---

# TICKET-05-03 — Quản lý sự kiện và bật Travel Mode

Ticket lớn: [Giao dịch định kỳ và du lịch](TICKET-05-dinh-ky-du-lich.md).

## Mục tiêu và phạm vi

Tạo/sửa sự kiện/chuyến đi, bật/tắt Travel Mode và xem giao dịch liên quan.

## Tiêu chí nghiệm thu

- [ ] Mỗi account chỉ có một Travel Mode bật; giao dịch mới đủ điều kiện tự gắn chuyến đó.
- [ ] Tắt/đổi chuyến không viết lại giao dịch cũ; user được sửa/gỡ liên kết.
- [ ] Recurring không tự gắn chuyến; liên kết không ghi đè note user.
- [ ] Chuyến và giao dịch có thật được dùng làm bối cảnh báo cáo tháng.

## Cần chốt

Nhập bù, offline hoặc xác nhận đề xuất sau khi đổi chuyến sẽ xác định chuyến theo thời điểm nào. Xem các tình huống còn mở trong [quy tắc nghiệp vụ](../../requirements/BUSINESS_RULES.md#các-tình-huống-cần-thống-nhất-tiếp).
