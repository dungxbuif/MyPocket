---
artifact_type: requirements_index
id: REQUIREMENTS
status: in_review
owner: shared
human_fields: [priority, acceptance, requirement_source]
ai_fields: [requirement_rows, status_updates, trace_links]
shared_fields: [functional_requirements, non_functional_requirements]
updated: 2026-09-13
trace:
  spec: SPEC.md
  business_rules: BUSINESS_RULES.md
  reports: REPORTS.md
  stories: USER_STORIES.md
  validation: ../work/VALIDATION_MATRIX.md
---

# Danh mục yêu cầu MyPocket

## Field Ownership

- Human owns priority, acceptance, and requirement source.
- AI maintains rows, trace links, and status updates.

## Functional Requirements

Các yêu cầu là hợp đồng đang review, chưa phải trạng thái triển khai. Mã nhóm quy tắc trỏ đến [BUSINESS_RULES.md](BUSINESS_RULES.md); báo cáo chi tiết tại [REPORTS.md](REPORTS.md).

| ID | Yêu cầu | Nhóm quy tắc |
| --- | --- | --- |
| REQ-01 | Account riêng, VND seed cố định, timezone theo user | ACC, TIME |
| REQ-02 | Ví thường, tiết kiệm, tín dụng và tổng ví | WAL |
| REQ-03 | Nhóm danh mục mặc định và danh mục cá nhân | CAT |
| REQ-04 | Tạo/sửa/xóa/nhân bản/lọc giao dịch, số dư nhất quán | TX |
| REQ-05 | Budget theo ví/danh mục/kỳ với cơ sở Money Lover | BUD |
| REQ-06 | Hũ nhóm chi thường, cấu hình tháng, cảnh báo mềm, view cộng dồn | JAR |
| REQ-07 | Recurring tự tạo giao dịch thường, note mặc định, chống trùng | REC |
| REQ-08 | Sự kiện và Travel Mode tự gắn giao dịch, loại recurring | TRV |
| REQ-09 | Tiết kiệm: mục tiêu, nạp/rút, tiến độ từ ví | SAV |
| REQ-10 | Credit: mua/hoàn/sao kê/trả nợ/đến hạn/quá hạn | CRD |
| REQ-11 | Vay/cho vay và các lần thanh toán liên kết | DEBT |
| REQ-12 | Reports, Money Insider, drilldown cùng số liệu | REP |
| REQ-13 | Tổng kết tháng cập nhật được, note độc lập, bối cảnh và AI | MONTH, AI |
| REQ-14 | Portfolio dùng bình quân gia quyền di động | AST |
| REQ-15 | API key quyền tương đương user, giới hạn quản lý key khác | KEY |
| REQ-16 | API có hợp đồng, chống trùng, phân trang và báo xung đột | API |
| REQ-17 | AI nhập liệu có xác nhận, OCR, tư vấn chỉ đọc | AI |
| REQ-18 | Offline, sync, thông báo, export và xóa dữ liệu | SYNC, LIFE |

## Non-Functional Requirements

| ID | Yêu cầu |
| --- | --- |
| NFR-01 | Phân tách account ở query, mutation, file, sync và AI tools |
| NFR-02 | VND chính xác; hiệu ứng nhiều vế nguyên tử |
| NFR-03 | Retry không trùng; phiên bản cũ có xung đột rõ ràng |
| NFR-04 | UTC cho thời điểm, date thuần riêng, query theo timezone account |
| NFR-05 | Số report truy về nguồn; AI phân biệt với số liệu tính toán |
| NFR-06 | PWA desktop/mobile có trạng thái tải, rỗng, lỗi, chưa đồng bộ |
| NFR-07 | Secret và nội dung tài chính riêng không đi vào log thường |

## Tình huống còn cần thống nhất

Xem mục cuối [BUSINESS_RULES.md](BUSINESS_RULES.md). Chưa đánh dấu các tình huống xóa liên kết, travel nhập bù, recurring cuối tháng, phân loại tổng thu, credit và liên kết portfolio là đã duyệt hoặc kiểm chứng.
