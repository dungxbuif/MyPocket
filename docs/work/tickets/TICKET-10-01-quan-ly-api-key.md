---
artifact_type: ticket
id: TICKET-10-01
status: draft
owner: human
parent: TICKET-10
trace:
  parent: TICKET-10-ket-noi-cong-cu.md
  guide: README.md
---

# TICKET-10-01 — Tạo và thu hồi API key

Ticket lớn: [Kết nối công cụ cá nhân](TICKET-10-ket-noi-cong-cu.md).

## Mục tiêu và phạm vi

Đặt tên, tạo key, xem thông tin sử dụng và thu hồi quyền của công cụ.

## Tiêu chí nghiệm thu

- [ ] Secret chỉ hiển thị khi tạo; danh sách sau đó không lộ lại secret.
- [ ] Thu hồi khiến công cụ không tiếp tục dùng key đó được.
- [ ] Key không tạo/quản lý key khác; được xem thông tin/thu hồi chính key đang dùng.
