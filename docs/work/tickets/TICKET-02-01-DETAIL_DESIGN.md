---
artifact_type: detail_design
id: DESIGN-02-01
status: in_progress
owner: shared
approval: authorized_by_user_full_spec
trace:
  ticket: TICKET-02-01-ghi-thu-chi.md
  backlog: ../BACKLOG.md
  spec: ../../requirements/SPEC.md
  tests: ../../work/VALIDATION_MATRIX.md
  database_operations: ../../architecture/DATABASE.md
  adr: ../../decisions/ADR-001-versioned-database-migrations.md
---

# Thu/chi cơ bản trên ledger

## Scope đã được duyệt

Tạo, xem, sửa và xóa giao dịch thu/chi VND trong một ví sở hữu bởi user. Mỗi giao dịch có ví, loại, số tiền dương, ngày giờ UTC, danh mục tùy chọn, note và cờ tính báo cáo. Số dư ví tính từ opening balance cộng thu trừ chi. Giao dịch chi có thể gắn tối đa một hũ trong slice hũ sau.

## Không nằm trong slice

Chuyển ví, điều chỉnh số dư, nợ/tín dụng, recurring, Travel Mode, chứng từ/OCR và phân trang nâng cao có model/edge case riêng. Không dùng mock để giả lập các hành vi này.

## Contract

`GET/POST /api/v1/transactions`, `PATCH/DELETE /api/v1/transactions/{id}`. Tất cả route protected, owner-scoped, dùng response envelope chung. Chỉ nhận `income` hoặc `expense`; amount phải lớn hơn 0; wallet phải thuộc owner; category nếu có phải visible và đúng kind.

## Proof và reconciliation

Unit/handler tests kiểm tra owner scope, validation và số dư tính đúng. Schema/seed được chạy bằng migration versioned trước API ở cả dev và prod; API không tự `AutoMigrate`. Swagger annotations nằm cạnh handler và được tạo lại qua `go generate ./cmd/api`, không có file cấu hình Swagger riêng. Router chỉ ghép route groups; từng domain tự đăng ký routes của nó.
