---
artifact_type: detail_design
id: DESIGN-02-01
status: in_review
owner: shared
approval: authorized_by_user_full_spec
trace:
  ticket: TICKET-02-01-ghi-thu-chi.md
  backlog: ../BACKLOG.md
  spec: ../../requirements/SPEC.md
  tests: ../../work/VALIDATION_MATRIX.md
  database_operations: ../../architecture/DATABASE.md
  adr: ../../decisions/ADR-001-versioned-database-migrations.md
  ui_spec: ../../design/screens/transactions/README.md
---

# Thu/chi cơ bản trên ledger

## Scope đã được duyệt

Tạo, xem, sửa và xóa giao dịch thu/chi VND trong một ví sở hữu bởi user. Mỗi giao dịch có ví, loại, số tiền dương, ngày giờ UTC, danh mục tùy chọn, note và cờ tính báo cáo. Số dư ví tính từ opening balance cộng thu trừ chi. Giao dịch chi có thể gắn tối đa một hũ trong slice hũ sau.

## Không nằm trong slice

Chuyển ví, điều chỉnh số dư, nợ/tín dụng, recurring, Travel Mode, chứng từ/OCR và phân trang nâng cao có model/edge case riêng. Không dùng mock để giả lập các hành vi này.

## Contract

`GET/POST /api/v1/transactions`, `PATCH/DELETE /api/v1/transactions/{id}`. Tất cả route protected, owner-scoped, dùng response envelope chung. Chỉ nhận `income` hoặc `expense`; amount phải lớn hơn 0; wallet phải thuộc owner; category nếu có phải visible và đúng kind.

Slice thu/chi chỉ chọn ví `basic` hoặc `goal`. Ví `credit` bị chặn ở cả API và UI vì CRD-01…04 định nghĩa chiều dư nợ, mua, hoàn tiền và thanh toán riêng; dùng công thức thu/chi thường sẽ làm sai nghĩa tín dụng.

Nhóm không có `wallet_ids` áp dụng cho mọi ví. Nhóm có `wallet_ids` chỉ được chọn và lưu với một ví nằm trong danh sách đó. API là lớp bảo vệ cuối, UI lọc cùng quy tắc và xóa lựa chọn nhóm không còn hợp lệ khi đổi ví hoặc loại giao dịch.

Số dư hiện tại của ví là giá trị suy ra theo WAL-02: `opening_balance + income - expense`. CRUD giao dịch không ghi đè `opening_balance`; list ví trả thêm `current_balance` để các màn dùng cùng một nguồn dữ liệu.

UI compose từ các base đã có: `BaseBottomSheet`, `SegmentedControl`, `FormField`/`BaseTextInput`/`BaseSelect`, `BaseCheckbox`, `StatusMessage`, `SurfaceCard`, `BaseButton` và `TransactionItem`. Behavior màn được khóa tại [screen spec](../../design/screens/transactions/README.md); không dựng control/card/màu riêng trong organism.

## Proof và reconciliation

Unit/handler tests kiểm tra owner scope, validation và số dư tính đúng. Schema/seed được chạy bằng migration versioned trước API ở cả dev và prod; API không tự `AutoMigrate`. Swagger annotations nằm cạnh handler và được tạo lại qua `go generate ./cmd/api`, không có file cấu hình Swagger riêng. Router chỉ ghép route groups; từng domain tự đăng ký routes của nó.
