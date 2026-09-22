---
artifact_type: ticket
id: TICKET-04-01
status: in_progress
owner: human
parent: TICKET-04
trace:
  parent: TICKET-04-hu-chi-tieu.md
  guide: README.md
  detail_design: CORE-03-TIME-JARS-MONTH-DETAIL_DESIGN.md
  validation: ../VALIDATION_MATRIX.md
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

Quyết định đã chốt: gỡ cấu hình hũ trong một tháng không xóa giao dịch, ID hũ hoặc cấu hình tháng khác. Phần trăm tính theo thu thực tế đủ điều kiện, loại nhóm chuyển nội bộ. Xem [thiết kế triển khai](CORE-03-TIME-JARS-MONTH-DETAIL_DESIGN.md).
