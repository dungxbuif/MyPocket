---
artifact_type: ticket
id: TICKET-07-04
status: in_progress
owner: human
parent: TICKET-07
trace:
  parent: TICKET-07-bao-cao-tong-ket.md
  guide: README.md
  detail_design: CORE-03-TIME-JARS-MONTH-DETAIL_DESIGN.md
  validation: ../VALIDATION_MATRIX.md
---

# TICKET-07-04 — Ghi nhớ tháng với note, bối cảnh và AI

Ticket lớn: [Báo cáo, Money Insider và tổng kết tháng](TICKET-07-bao-cao-tong-ket.md).

## Mục tiêu và phạm vi

Xem realtime, tổng kết cuối tháng, ghi note, tự đính kèm bối cảnh và nhận tổng kết AI.

## Tiêu chí nghiệm thu

- [ ] Sửa giao dịch sau tổng kết vẫn cập nhật số của tháng đó.
- [ ] Một note/account/tháng; chỉ thao tác chủ động của user mới sửa/xóa note.
- [ ] Chuyến đi/bối cảnh có thật tự đính kèm; AI tổng kết riêng, không ghi đè note hoặc tạo số.
- [ ] Nguồn đổi thì AI cũ được nhận biết/cập nhật; AI lỗi không cản xem số hoặc ghi note.

## Implementation scope

Phần realtime overview, tổng kết số tự tính theo timezone, trạng thái tháng tự hoàn tất và note riêng đang thực hiện theo [CORE-03](CORE-03-TIME-JARS-MONTH-DETAIL_DESIGN.md). AI narrative/bối cảnh sự kiện là phần còn lại của ticket, không đánh dấu xong theo lát cắt này.
