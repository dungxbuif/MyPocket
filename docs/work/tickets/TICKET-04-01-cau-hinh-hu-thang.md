---
artifact_type: ticket
id: TICKET-04-01
status: draft
owner: human
parent: TICKET-04
trace:
  parent: TICKET-04-hu-chi-tieu.md
  guide: README.md
---

# TICKET-04-01 — Quản lý hũ theo tháng

Ticket lớn: [Theo dõi chi tiêu bằng hũ](TICKET-04-hu-chi-tieu.md).

## Mục tiêu và phạm vi

Thêm/sửa/xóa hũ; đặt mức tham khảo theo % thu thực tế hoặc tiền nhập tay.

## Tiêu chí nghiệm thu

- [ ] Tự chọn tên/số lượng; có thể dùng hũ mà không đặt mức phân bổ.
- [ ] Tháng mới lấy cấu hình gần nhất; sửa tháng cũ không ghi đè tháng đã có cấu hình.
- [ ] Phân bổ vượt thu hoặc tổng % vượt 100% chỉ cảnh báo; không chuyển dư sang tháng sau.

## Cần chốt

Xóa hũ đang có giao dịch và phân loại khoản nào thuộc tổng thu thực tế. Xem các tình huống còn mở trong [quy tắc nghiệp vụ](../../requirements/BUSINESS_RULES.md#các-tình-huống-cần-thống-nhất-tiếp).
