---
artifact_type: ticket
id: TICKET-02-01
status: in_review
owner: human
parent: TICKET-02
trace:
  parent: TICKET-02-giao-dich.md
  guide: README.md
  detail_design: TICKET-02-01-DETAIL_DESIGN.md
  verification: TICKET-02-01-VERIFICATION.md
---

# TICKET-02-01 — Ghi thu và chi

Ticket lớn: [Ghi chép và quản lý giao dịch](TICKET-02-giao-dich.md).

## Mục tiêu và phạm vi

Nhập tiền, ví, ngày, danh mục, note và chứng từ; chi thường có thể gắn hũ.

## Tiêu chí nghiệm thu

- [x] Giao dịch hợp lệ cập nhật số dư đúng chiều thu/chi.
- [ ] Chỉ chi thường được gắn tối đa một hũ; chưa gắn hũ vẫn lưu được.
- [ ] Xem lại được ghi chú và chứng từ cùng giao dịch.

Ledger thu/chi cơ bản và ghi chú đã được [verification](TICKET-02-01-VERIFICATION.md). Hũ và chứng từ/OCR vẫn là acceptance còn mở của ticket BA; không bị mô phỏng trong slice này.
