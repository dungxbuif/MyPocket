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

## Native E2E bổ sung — gắn/gỡ chuyến thủ công

- **Behavior:** từ chi tiết một giao dịch thường, chọn một Travel Mode rồi lưu;
  tải lại qua API phải cho thấy đúng `travel_event_id`. Chọn “Không gắn chuyến”
  và lưu lần nữa; đọc lại qua API phải xác nhận `travel_event_id` đã rỗng.
- **Kỹ thuật:** giữ nguyên endpoint owner-scoped
  `PATCH /transactions/{id}/travel`; test chỉ dùng fixture API cục bộ của user
  thử nghiệm, không thay đổi backend/schema/migration hay API contract.
- **Verification:** tạo event, wallet, category và transaction fixture riêng;
  thao tác qua native UI, xác nhận server state sau từng lần lưu, rồi xóa fixture.
  Chạy test trên iPhone 16 Simulator với local API và database dùng riêng cho
  kiểm thử; production không nằm trong luồng này.
- **Docs:** cập nhật validation matrix, native port audit và backlog sau khi có
  kết quả. Public API/OpenAPI/AI endpoint catalog không đổi vì contract hiện hữu
  được giữ nguyên.
- **Result (2026-10-08):** PASS trên iPhone 16 Simulator tại
  `/tmp/mypocket-travel-link-final-20261008.xcresult`. Test tạo chuyến, ví,
  giao dịch; thao tác chọn/gỡ chuyến từ chi tiết native; đọc lại `travel_event_id`
  qua API sau mỗi mutation. Backend contract tests xác nhận tự gắn giao dịch mới
  và semantics xóa liên kết đều pass (2 tests). Test dùng bản sao DB disposable
  `mypocket_travel_link_verify_20261008`; đã xác nhận 0 fixture wallet/event/
  transaction, dừng API 8081 và xóa đúng DB disposable. Production không truy cập.

## Cần chốt

Nhập bù, offline hoặc xác nhận đề xuất sau khi đổi chuyến sẽ xác định chuyến theo thời điểm nào. Xem các tình huống còn mở trong [quy tắc nghiệp vụ](../../requirements/BUSINESS_RULES.md#các-tình-huống-cần-thống-nhất-tiếp).
