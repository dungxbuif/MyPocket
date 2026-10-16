---
artifact_type: ticket
id: TICKET-06-01
status: in_progress
owner: human
parent: TICKET-06
trace:
  parent: TICKET-06-tiet-kiem-tin-dung-no.md
  guide: README.md
---

# TICKET-06-01 — Theo dõi mục tiêu tiết kiệm

Ticket lớn: [Tiết kiệm, tín dụng và khoản nợ](TICKET-06-tiet-kiem-tin-dung-no.md).

## Mục tiêu và phạm vi

Owner requested implementation from supplied savings screens. [Active design/proof](TICKET-06-01-DETAIL_DESIGN.md). Goal date, summary/history and shared category selection are implemented. Managed-wallet paired transfers are implemented in the shared transaction flow (not fabricated by savings deposit/withdrawal entries); authenticated browser/device UAT remains pending. No unchecked acceptance criterion is silently marked done.

Đặt mục tiêu/hạn cho ví tiết kiệm, nạp/rút và xem tiến độ.

**Gap-closure status 2026-10-16:** Goal progress and the existing paired
transfer/single-wallet entry contracts are covered by current API/native tests.
Scheduled savings advice, forecast, notification fan-out and a dedicated
savings report remain **partial** because no current API/entity contract
defines them; this ticket does not infer those behaviors from the goal fields.

## Tiêu chí nghiệm thu

- [ ] Tiến độ từ số thực trong ví; sửa giao dịch cập nhật tiến độ.
- [ ] Đạt mục tiêu được thể hiện rõ; rút xuống dưới mục tiêu thì tiến độ phản ánh lại.
- [ ] Chuyển giữa các ví không bị tính thu/chi thường lần nữa.
