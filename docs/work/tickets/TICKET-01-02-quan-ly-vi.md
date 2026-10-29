---
artifact_type: ticket
id: TICKET-01-02
status: in_review
owner: human
parent: TICKET-01
trace:
  parent: TICKET-01-tai-khoan-va-vi.md
  guide: README.md
  detail_design: TICKET-01-02-DETAIL_DESIGN.md
  verification: TICKET-01-02-VERIFICATION.md
---

# TICKET-01-02 — Quản lý ví

Ticket lớn: [Tài khoản, ví và danh mục](TICKET-01-tai-khoan-va-vi.md).

## Mục tiêu và phạm vi

Thêm, sửa, xóa ví; xem số dư và loại ví thường, tiết kiệm, tín dụng.

## Tiêu chí nghiệm thu

- [x] Tạo và sửa được thông tin phù hợp với loại ví trong core slice.
- [x] Mỗi ví có số dư và lịch sử riêng để đối chiếu.
- [x] Trước khi xóa, thấy dữ liệu liên quan và ảnh hưởng đến báo cáo.

Core CRUD đã được [verification](TICKET-01-02-VERIFICATION.md) bằng unit, HTTP và browser UAT. Metadata kỳ sao kê và ledger tín dụng vẫn thuộc slice tín dụng riêng, không được đánh dấu hoàn tất tại đây.

## Cần chốt

Đã chốt: xoá thực ví và dữ liệu liên quan sau cảnh báo/xác nhận; cho phép trùng tên ví. Nghiên cứu điều chỉnh số dư và trường riêng ba loại ví nằm trong [detail design](TICKET-01-02-DETAIL_DESIGN.md#quyết-định-đã-chốt-và-kết-quả-nghiên-cứu).

Xử lý giao dịch chuyển tiền/trả thẻ liên kết với ví khác khi xóa ví. Xem các tình huống còn mở trong [quy tắc nghiệp vụ](../../requirements/BUSINESS_RULES.md#các-tình-huống-cần-thống-nhất-tiếp).
