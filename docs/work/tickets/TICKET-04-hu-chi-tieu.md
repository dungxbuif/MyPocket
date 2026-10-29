---
artifact_type: ticket
id: TICKET-04
status: in_progress
owner: human
kind: parent
requirements: [REQ-06]
trace:
  guide: README.md
  detail_design: CORE-03-TIME-JARS-MONTH-DETAIL_DESIGN.md
  validation: ../VALIDATION_MATRIX.md
---

# TICKET-04 — Theo dõi chi tiêu bằng hũ

## Mục tiêu

Phân nhóm chi linh hoạt và nhìn lại từng tháng hoặc nhiều tháng.

## Ticket con

| Ticket | Phạm vi |
| --- | --- |
| [TICKET-04-01 — Quản lý hũ theo tháng](TICKET-04-01-cau-hinh-hu-thang.md) | Thêm/sửa/xóa hũ; đặt mức tham khảo theo % thu thực tế hoặc tiền nhập tay. |
| [TICKET-04-02 — Gắn hũ và xem chi tiêu tháng](TICKET-04-02-chi-tieu-hu-thang.md) | Gắn/gỡ hũ cho chi thường; xem tổng chi, giao dịch và mức tham khảo trong tháng. |
| [TICKET-04-03 — Xem hũ cộng dồn](TICKET-04-03-hu-cong-don.md) | Chọn hũ và khoảng tháng hoặc toàn bộ lịch sử để xem tổng chi và diễn biến. |

## Hoàn thành khi

- Các ticket con đáp ứng tiêu chí nghiệm thu và được người dùng xác nhận.
- Số liệu và thao tác giữa các ticket con thống nhất với [đặc tả nghiệp vụ](../../requirements/BUSINESS_RULES.md).

## Implementation

API, cấu hình theo tháng, gắn giao dịch tùy chọn, báo cáo cộng dồn và màn hình dùng dữ liệu thật đã được triển khai theo [detail design](CORE-03-TIME-JARS-MONTH-DETAIL_DESIGN.md). Kiểm thử PostgreSQL, selector chi tiêu và build pass; quyết định xóa chỉ bỏ cấu hình tháng đó, giữ ID/lịch sử/liên kết. Ticket vẫn `in_progress` đến khi owner UAT xác nhận; xem [validation evidence](../VALIDATION_MATRIX.md#core-03-timezone-jars-and-live-monthly-summary--2026-09-22).

Nguồn: [REQ-06](../../requirements/REQUIREMENTS.md). Quy ước trạng thái và liên kết chung: [Danh sách ticket](README.md).
