---
artifact_type: ticket
id: TICKET-09-01
status: in_review
owner: human
parent: TICKET-09
trace:
  parent: TICKET-09-tro-ly-ai.md
  guide: README.md
---

# TICKET-09-01 — Nhập giao dịch bằng nội dung hoặc chứng từ

Ticket lớn: [Trợ lý nhập liệu và tư vấn](TICKET-09-tro-ly-ai.md). Live synthetic extraction now passes strict JSON Schema evaluation; see [implementation detail/proof](TICKET-09-01-ENTRY-DETAIL_DESIGN.md). Owner UAT remains pending.

Thiết kế tổng thể: [DESIGN-09-AI](TICKET-09-DETAIL_DESIGN.md). Owner đã yêu cầu triển khai lát [AI-ENTRY-01](TICKET-09-01-ENTRY-DETAIL_DESIGN.md): giữ nút thêm → chat → danh sách điền sẵn → sửa/duyệt/từ chối. Code và kiểm thử fixture/DB đã có; live model/OCR quality và owner UAT còn chờ. Không đánh dấu toàn bộ multi-bank/transfer/retained receipts hoàn tất.

## Mục tiêu và phạm vi

Gửi nội dung/ảnh, nhận đề xuất, sửa rồi xác nhận hoặc từ chối.

## Tiêu chí nghiệm thu

- [ ] Thông tin thiếu được hỏi lại; hỗ trợ nhiều đề xuất từ nội dung phù hợp.
- [ ] Đề xuất xem/sửa được, chưa tác động số dư trước xác nhận.
- [ ] Xác nhận lại không tạo trùng; OCR lỗi không tự tạo giao dịch.
- [ ] Chỉ đọc được chứng từ thuộc tài khoản đang sử dụng.
