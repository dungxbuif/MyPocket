---
artifact_type: ticket
id: AI-ENTRY-03
status: draft
owner: human
kind: product
parent: TICKET-09
trace:
  parent: TICKET-09-tro-ly-ai.md
  detail_design: AI-ENTRY-03-PROVIDER-SELECTION-DETAIL_DESIGN.md
  backlog: ../BACKLOG.md
  validation: ../VALIDATION_MATRIX.md
  docs_review: ../../standards/README.md
  adr_registry: ../../decisions/README.md
  release_notes: ../../releases/CHANGELOG.md
---

# AI-ENTRY-03 — Chọn nhà cung cấp AI và minh bạch chi phí

Yêu cầu tương lai của owner (2026-09-22): có thể tích hợp nhiều nhà cung cấp/model để người dùng lựa chọn và xem/so sánh chi phí trước khi dùng. Đây là yêu cầu thiết kế; chưa cho phép kết nối thêm provider hoặc phát sinh chi phí.

## Mục tiêu

Người dùng hiểu dữ liệu nào được gửi tới dịch vụ nào, chọn được provider/model được hỗ trợ, và thấy giá cùng ước tính chi phí trước khi gửi nội dung xử lý.

## Phạm vi đề xuất

- Lập danh mục provider/model và năng lực đã kiểm chứng cho flow trích xuất giao dịch (ví dụ strict JSON Schema, giới hạn context, độ ổn định và latency).
- Hiển thị giá đầu vào/đầu ra, đơn vị tiền, nguồn giá chính thức và thời điểm kiểm tra; tách phí OCR nếu có.
- Ước tính chi phí của yêu cầu dựa trên token/định dạng đầu vào; nêu rõ đó là ước tính, không phải báo giá cuối cùng.
- Cho phép chọn provider/model; không tự động chuyển sang provider khác khi lỗi hoặc giá thay đổi.
- Công khai dữ liệu gửi đi, vùng xử lý/lưu giữ theo thông tin provider và yêu cầu quyền riêng tư trước khi bật lựa chọn.

## Tiêu chí nghiệm thu cần xác nhận trong thiết kế

- Người dùng có thể so sánh provider/model đang được hỗ trợ cùng giá và ngày xác minh nguồn giá.
- Trước khi gửi, UI nêu provider/model được chọn và ước tính chi phí; khi không thể ước tính thì nói rõ.
- Thay đổi giá cũ/hết hạn không được trình bày như giá hiện hành.
- Provider lỗi không âm thầm chạy lại qua provider khác.
- Không gửi khóa provider xuống trình duyệt; không gửi dữ liệu giao dịch/chứng từ trước khi người dùng chủ động gửi.

## Quyết định còn mở

- Dùng khóa do ứng dụng quản lý hay cho người dùng tự cấu hình khóa (BYOK), hoặc hỗ trợ cả hai?
- Người dùng chọn theo từng lần xử lý hay đặt mặc định trong cài đặt?
- Cần giới hạn ngân sách/lượt xử lý và xác nhận bổ sung ở ngưỡng nào?
- Danh sách provider/model và chính sách vùng dữ liệu được phép là gì?

## Trạng thái

Ticket và [detail design](AI-ENTRY-03-PROVIDER-SELECTION-DETAIL_DESIGN.md) ở trạng thái `draft`, chờ owner review các quyết định trên. Không triển khai trong AI-ENTRY-01/02.

Liên kết: [TICKET-09](TICKET-09-tro-ly-ai.md) · [AI-ENTRY-01](TICKET-09-01-ENTRY-DETAIL_DESIGN.md) · [Backlog](../BACKLOG.md) · [Validation](../VALIDATION_MATRIX.md) · [ADR registry](../../decisions/README.md) · [Changelog](../../releases/CHANGELOG.md).
