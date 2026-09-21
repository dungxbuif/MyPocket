---
artifact_type: ticket
id: TICKET-09
status: draft
owner: human
kind: parent
requirements: [REQ-17]
trace:
  guide: README.md
---

# TICKET-09 — Trợ lý nhập liệu và tư vấn

## Mục tiêu

Ghi chép nhanh từ nội dung/chứng từ và hỏi về tài chính cá nhân.

## Kế hoạch để review

[Đánh giá tính khả thi và DETAIL_DESIGN hai flow](TICKET-09-DETAIL_DESIGN.md), ngày 2026-09-21: proposal/OCR, đối chiếu chuyển nội bộ, hỏi đáp chỉ đọc, prompt/eval và thứ tự triển khai. Thiết kế đang `in_review`, chưa duyệt implementation; không đổi trạng thái nghiệm thu các ticket con.

## Ticket con

| Ticket | Phạm vi |
| --- | --- |
| [TICKET-09-01 — Nhập giao dịch bằng nội dung hoặc chứng từ](TICKET-09-01-nhap-lieu-chung-tu.md) | Gửi nội dung/ảnh, nhận đề xuất, sửa rồi xác nhận hoặc từ chối. |
| [TICKET-09-02 — Hỏi đáp tài chính cá nhân](TICKET-09-02-hoi-dap-tai-chinh.md) | Hỏi trợ lý về giao dịch, reports, budget, hũ, nợ và tài sản. |

## Hoàn thành khi

- Các ticket con đáp ứng tiêu chí nghiệm thu và được người dùng xác nhận.
- Số liệu và thao tác giữa các ticket con thống nhất với [đặc tả nghiệp vụ](../../requirements/BUSINESS_RULES.md).

Nguồn: [REQ-17](../../requirements/REQUIREMENTS.md). Quy ước trạng thái và liên kết chung: [Danh sách ticket](README.md).
