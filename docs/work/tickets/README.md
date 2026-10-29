---
artifact_type: ticket_index
id: TICKET-INDEX
status: draft
owner: shared
updated: 2026-09-13
human_fields: [priority, acceptance, business_decisions]
ai_fields: [ticket_breakdown, trace_links, review_evidence]
shared_fields: [status]
trace:
  backlog: ../BACKLOG.md
  requirements: ../../requirements/REQUIREMENTS.md
  spec: ../../requirements/SPEC.md
  business_rules: ../../requirements/BUSINESS_RULES.md
  reports: ../../requirements/REPORTS.md
  validation_matrix: ../VALIDATION_MATRIX.md
  docs_review: ../VALIDATION_MATRIX.md#ba-ticket-review
  release_notes: ../../releases/CHANGELOG.md
  phase: null
  detail_design: TICKET-01-02-DETAIL_DESIGN.md
  test_verification: null
  adrs: []
---

# Ticket nghiệp vụ MyPocket

11 ticket lớn, 32 ticket nhỏ. Ticket lớn gom một nhóm nhu cầu; ticket nhỏ mô tả một phần việc có thể nghiệm thu riêng. Tất cả đang là bản nháp để review phạm vi, chưa phải quyết định bắt đầu triển khai.

## Công việc nền tảng ngoài bộ ticket nghiệp vụ

- [WORKER-01 — Dịch vụ worker chạy cronjob](WORKER-01-cron-service.md): yêu cầu riêng về worker dùng chung code backend, với giao dịch định kỳ và chốt/tạo report tháng làm ví dụ. Đây là work item high-risk ở trạng thái `draft`, không nằm trong thống kê 11 ticket lớn/32 ticket con. Cần chốt report là live/rebuildable hay snapshot/immutable close và quy tắc recurring trước detail design; hiện CORE-03 chưa có cron close.

## Danh sách ticket lớn

| Ticket lớn | Số ticket nhỏ | Yêu cầu |
| --- | ---: | --- |
| [TICKET-01 — Tài khoản, ví và danh mục](TICKET-01-tai-khoan-va-vi.md) | 3 | REQ-01, REQ-02, REQ-03 |
| [TICKET-02 — Ghi chép và quản lý giao dịch](TICKET-02-giao-dich.md) | 3 | REQ-04 |
| [TICKET-03 — Theo dõi budget](TICKET-03-budget.md) | 2 | REQ-05 |
| [TICKET-04 — Theo dõi chi tiêu bằng hũ](TICKET-04-hu-chi-tieu.md) | 3 | REQ-06 |
| [TICKET-05 — Giao dịch định kỳ và du lịch](TICKET-05-dinh-ky-du-lich.md) | 3 | REQ-07, REQ-08 |
| [TICKET-06 — Tiết kiệm, tín dụng và khoản nợ](TICKET-06-tiet-kiem-tin-dung-no.md) | 4 | REQ-09, REQ-10, REQ-11 |
| [TICKET-07 — Báo cáo, Money Insider và tổng kết tháng](TICKET-07-bao-cao-tong-ket.md) | 4 | REQ-12, REQ-13 |
| [TICKET-08 — Theo dõi danh mục tài sản](TICKET-08-danh-muc-tai-san.md) | 2 | REQ-14 |
| [TICKET-09 — Trợ lý nhập liệu và tư vấn](TICKET-09-tro-ly-ai.md) | 3 | REQ-17 |
| [TICKET-10 — Kết nối công cụ cá nhân](TICKET-10-ket-noi-cong-cu.md) | 2 | REQ-15, REQ-16 |
| [TICKET-11 — Sử dụng liên tục và quản lý dữ liệu](TICKET-11-su-dung-du-lieu.md) | 3 | REQ-18 |

Thứ tự trong bảng để dễ đọc, chưa phải thứ tự ưu tiên triển khai. Mở ticket lớn để xem danh sách và đường dẫn từng ticket nhỏ.

## Cách đọc và nghiệm thu

- Ticket nhỏ chỉ gồm mục tiêu/phạm vi, tiêu chí nghiệm thu và phần cần chốt nếu có.
- Mỗi ticket nhỏ liên kết về ticket lớn; ticket lớn liên kết về yêu cầu nguồn và toàn bộ ticket nhỏ.
- Các điểm chưa chốt vẫn giữ nguyên là câu hỏi nghiệp vụ, không coi một đề xuất là quyết định đã duyệt.
- Tiêu chí có ô trống vì chưa thực hiện nghiệm thu. Chỉ hoàn thành ticket lớn khi các ticket nhỏ đã được nghiệm thu.
- Ngôn ngữ ticket mô tả kết quả người dùng nhận được. Thiết kế kỹ thuật và kế hoạch triển khai được xác định khi có yêu cầu riêng.

## Quy tắc dùng chung

Các ticket cùng tuân theo [đặc tả](../../requirements/SPEC.md) và [nghiệp vụ](../../requirements/BUSINESS_RULES.md): VND, dữ liệu riêng theo tài khoản, múi giờ người dùng, số liệu nhất quán, không tạo trùng khi gửi lại, trạng thái thiếu dữ liệu rõ ràng. Khi dùng trên điện thoại/máy tính, người dùng phải đọc và thực hiện được các thao tác trong phạm vi ticket.

Budget, hũ, tiết kiệm, tín dụng, nợ và tài sản sở hữu quy tắc tính của chính chức năng. [TICKET-07](TICKET-07-bao-cao-tong-ket.md) tập trung trình bày và so sánh báo cáo bằng các số liệu đó. Tổng kết AI tháng ở TICKET-07-04; hỏi đáp AI ở TICKET-09-02.

## Liên kết và trạng thái chung

Các ticket dẫn về trang này để dùng chung [backlog](../BACKLOG.md), [bằng chứng/nghiệm thu](../VALIDATION_MATRIX.md), [ghi nhận thay đổi](../../releases/CHANGELOG.md) và tài liệu nguồn. Các trường để trống trong phần liên kết có nghĩa chưa lập thiết kế/kiểm thử triển khai hoặc không gắn phase; không phải liên kết hỏng. ADR không phát sinh chỉ từ việc chia ticket nghiệp vụ.

Chủ sản phẩm chốt ưu tiên, câu hỏi nghiệp vụ và nghiệm thu. AI duy trì cấu trúc, liên kết và bằng chứng kiểm tra tài liệu. Chia ticket không tự phê duyệt chức năng hoặc thay đổi quy tắc nghiệp vụ.

## Review tài liệu

Detail design cho quản lý ví đã được tạo tại [TICKET-01-02-DETAIL_DESIGN.md](TICKET-01-02-DETAIL_DESIGN.md) và đang chờ owner review; chưa có schema/API implementation.

Đối chiếu độ bao phủ yêu cầu, quan hệ cha–con và liên kết được ghi tại [BA ticket review](../VALIDATION_MATRIX.md#ba-ticket-review). Không có kiểm thử chức năng hoặc UAT sản phẩm trong lần chia ticket này.
