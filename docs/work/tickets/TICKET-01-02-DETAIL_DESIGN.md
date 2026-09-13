---
artifact_type: detail_design
id: DESIGN-01-02
status: ready
owner: ai
approval: pending
human_fields:
  - approval
  - constraints
  - scope_decisions
ai_fields:
  - problem
  - context_loaded
  - brownfield_scope
  - proposed_approach
  - design_tradeoffs
  - architecture_overview
  - execution_flow
  - api_data_model
  - security
  - test_plan
  - reconciliation_plan
shared_fields:
  - status
  - trace
  - small_task_exemption
trace:
  backlog_item: ../BACKLOG.md#bl-005
  requirement: ../../requirements/REQUIREMENTS.md#req-02
  phase: null
  ticket_or_bug: TICKET-01-02
  test_verification: ../VALIDATION_MATRIX.md
  validation_matrix: ../VALIDATION_MATRIX.md
  docs_review: ../VALIDATION_MATRIX.md
  adrs: []
  master_docs_touched:
    - ../../requirements/SPEC.md
    - ../../requirements/BUSINESS_RULES.md
    - ../../architecture/API.md
---

# DETAIL DESIGN: Quản lý ví

## Status

- ID: DESIGN-01-02
- Status: ready for human review
- Ticket: [TICKET-01-02](TICKET-01-02-quan-ly-vi.md)
- Approval: pending
- Updated: 2026-09-13

## 1. Context & Scope

### Problem statement

Người dùng cần tạo, xem, sửa và xoá các ví thuộc tài khoản để theo dõi số dư riêng. Đây là nền tảng cho giao dịch, budget, hũ, báo cáo và tổng ví.

### Scope for this slice

- Ví thuộc đúng một account/user.
- Loại ví: `basic` (thường), `goal` (tiết kiệm), `credit` (tín dụng).
- Tên ví, số dư ban đầu, ghi chú/mô tả và trạng thái được hiển thị theo loại ví.
- Danh sách ví và tổng số dư của các ví được chọn.
- Tạo, sửa, xem chi tiết và xoá ví.
- Chỉ VND ở giai đoạn hiện tại; currency là giá trị hệ thống cố định, chưa mở API đa tiền tệ.

### Out of scope

- Giao dịch, chuyển ví, trả thẻ, budget, hũ và báo cáo chi tiết.
- Archive/soft-delete; sản phẩm chỉ có ví và xoá.
- Shared wallet.
- Tự xử lý các vế giao dịch liên kết khi xoá ví cho đến khi owner chốt.

### Quyết định đã chốt và kết quả nghiên cứu

- Bộ nhóm mặc định tái sử dụng từ [catalog app cũ](../../../refereces/disappointed_app/backend/migrations/0011_phase002_category_catalog.sql), đối chiếu [nghiên cứu cũ](../../../refereces/disappointed_app/docs/research/moneylover/moneylover-full-research.md#17-default-categories). Giữ tên tiếng Việt, `system_key`, loại thu/chi/vay-nợ và quan hệ cha/con; khi migrate phải đối chiếu cả seed tiền nhiệm mà catalog tham chiếu. Không tự thay catalog bằng bộ nhóm Money Lover mới hoặc mock FE. Chỉ port seed khi triển khai danh mục; không chạy SQL của app cũ vào DB mới.

- Owner đã chốt xoá thực ví và dữ liệu liên quan sau cảnh báo/xác nhận. Không hỏi lại lựa chọn chặn xoá chỉ vì ví có giao dịch.
- Owner đồng ý cho phép trùng tên ví; phân biệt bằng ID, không có unique constraint cho tên.
- Phạm vi tác động tới vế chuyển tiền/trả thẻ nằm ở ví còn tồn tại vẫn cần ma trận riêng trước khi triển khai liên kết đó. Đây là chi tiết xử lý liên kết, không phải mở lại quyết định xoá ví.

#### Money Lover — đối chiếu nguồn chính thức ngày 2026-09-13

| Phần | Hành vi được tài liệu chính thức mô tả | Nguồn |
| --- | --- | --- |
| Điều chỉnh số dư | Nhập số dư đích; lưu tạo giao dịch thu/chi mới bằng chênh lệch, mặc định loại khỏi báo cáo; người dùng có thể đổi cờ này. Không mô tả tự sửa giao dịch cũ. | [Adjust wallet's balance](https://moneylover.zendesk.com/hc/en-us/articles/35569781199001-Adjust-wallet-s-balance) |
| Tạo ví | Basic có icon, tên, tiền tệ, số dư đầu. Goal thêm mục tiêu, ngày kết thúc. Credit có hạn mức, số dư sao kê gần nhất, ngày sao kê và ngày thanh toán. Xoá ví xoá dữ liệu liên quan. | [Wallet management](https://moneylover.zendesk.com/hc/en-us/articles/34974522779417-Create-edit-archive-and-delete-wallets) |
| Tiết kiệm | Theo dõi đã tiết kiệm, còn thiếu, tiến độ và dòng tiền; hướng dẫn chưa hỗ trợ tính lãi tiết kiệm ngân hàng. | [Goal wallet](https://moneylover.zendesk.com/hc/en-us/articles/36976160006809-How-to-manage-your-financial-goals-savings-effectively) |
| Tín dụng | Số dư mang dấu âm khi nợ, có thể dương khi trả dư; hạn mức khả dụng = hạn mức + inflow − outflow. Báo cáo hiển thị số dư, hạn trả, hạn mức và dòng tiền. | [Credit wallet](https://moneylover.zendesk.com/hc/en-us/articles/37006112179609-How-to-manage-credit-cards-effectively) |

Đây là nghiên cứu tài liệu hỗ trợ chính thức, chưa phải kiểm thử app Money Lover trực tiếp. Nguồn chưa xác định đầy đủ cách phân bổ hoàn tiền/thanh toán giữa các sao kê, ngày 29–31 và tác động xoá tới ví đối ứng.

#### Đề xuất áp dụng vào MyPocket — chờ review thiết kế

- Giữ VND và giao diện hiện tại. Trường riêng của `goal`/`credit` phải được thiết kế cùng form tạo ví, không chỉ có trường `type`.
- `goal`: `target_amount` > 0, `target_date` tuỳ chọn theo SPEC hiện có; hiển thị số hiện có, còn thiếu và tiến độ từ giao dịch. Không triển khai công thức lãi ngân hàng từ nghiên cứu này.
- `credit`: `credit_limit` > 0, `last_statement_balance`, `statement_day`, `payment_due_day`. Phải phân biệt số dư sao kê gần nhất với dư nợ hiện tại; không mặc định hai số này luôn bằng nhau. Quy tắc kỳ và phân bổ trả nợ vẫn thuộc thiết kế tín dụng cần hoàn thiện.
- Sửa metadata ví không ghi đè số dư hiện tại. “Điều chỉnh số dư” tạo một giao dịch mới có nguồn adjustment: `delta = target_balance - current_balance`. Ví dụ 1.000.000 → 800.000 tạo khoản giảm 200.000; chênh lệch 0 không cần tạo giao dịch. Không tự chọn giao dịch cũ để sửa.
- Sửa một giao dịch sai là luồng riêng; số dư tính lại theo WAL-02/TX-04. Cách biểu diễn số dư khởi tạo trong sổ giao dịch cần được thiết kế cùng ledger để không đếm hai lần.
- MyPocket hiện quy định adjustment không tự thành thu/chi thường (TX-03). Mặc định loại khỏi báo cáo phù hợp quy tắc này; khả năng bật lại cờ như Money Lover chưa được coi là quyết định đã duyệt.
- Điều chỉnh cần ledger và thao tác DB nguyên tử, kiểm soát gửi lại/đồng thời. Vì vậy không thể gọi nó đã hoàn thiện chỉ bằng CRUD `opening_balance`; thiết kế hiện tại chưa đủ để triển khai đầy đủ chức năng số dư.

### Small task exemption

- Small task exemption: no
- Reason: thay đổi database, API public và ảnh hưởng dữ liệu giao dịch.
- Impact checked: API=yes, DB=yes, Security=yes, Runtime=yes, Standards=no

## 2. Design considerations

| Alternative | Pros | Cons | Decision |
| --- | --- | --- | --- |
| Một bảng `wallets` với `type` enum/string | Truy vấn đơn giản, dùng chung số dư và ownership | Logic credit/goal sẽ mở rộng theo type | Recommended |
| Ba bảng riêng cho từng loại ví | Ràng buộc riêng rõ | Trùng CRUD, khó tổng hợp và mở rộng | Rejected |
| Xoá thực sau xác nhận | Theo quyết định owner | Phải nêu ảnh hưởng dữ liệu và xử lý liên kết nhất quán | Đã chốt nguyên tắc; ma trận ví đối ứng còn mở |

## 3. Architecture overview

- `wallet` entity/model: dữ liệu ví và ownership.
- `wallet repository`: CRUD bằng GORM, không raw SQL.
- `wallet usecase`: kiểm tra input, ownership và quy tắc xoá.
- `wallet handler`: REST JSON, gắn JWT middleware.
- FE `WalletCard`/wallet management organism: dùng base components hiện có; nếu thiếu base component thì tạo trước khi lắp màn hình.

## 4. Execution flow

1. FE gửi request kèm `Authorization: Bearer <JWT>`.
2. Middleware xác định user/account hiện tại.
3. Use case xác thực loại ví, tên, số dư và currency VND.
4. Repository truy vấn theo `owner_id` bằng GORM.
5. Handler trả resource ví hoặc lỗi chuẩn.
6. FE cập nhật danh sách, tổng ví và trạng thái loading/error.

## 5. API & data model design

### Planned API

- `GET /api/v1/wallets` — danh sách ví và tổng số dư được chọn.
- `POST /api/v1/wallets` — tạo ví.
- `GET /api/v1/wallets/:id` — xem chi tiết ví.
- `PATCH /api/v1/wallets/:id` — sửa thông tin ví.
- `DELETE /api/v1/wallets/:id` — xoá thực sau xác nhận; hợp đồng dữ liệu liên quan phải được mô tả theo các bảng đã implement.

All endpoints require JWT and only return the authenticated user's wallets.

### Proposed fields

Tên bảng tài khoản được owner chốt là `user`; `owner_id` tham chiếu `user.id`. [ERD](../../architecture/ERD.md) ghi bước rename bảo toàn dữ liệu từ bảng hiện tại và cập nhật GORM mapping trước migration ví.

`wallets`:

- `id` UUID primary key
- `owner_id` UUID, indexed, required
- `name` string, required
- `type` string: `basic | goal | credit`
- `currency` string, fixed `VND` in this stage
- `opening_balance` integer VND, required
- `is_in_total` boolean, default true
- `description` nullable string
- `created_at`, `updated_at`

Indexes: `(owner_id)`, `(owner_id, name)` non-unique nếu cần lookup. Cho phép trùng tên đã được owner xác nhận. Danh sách fields trên mới là phần chung; fields riêng goal/credit và ledger trong đề xuất phía trên chưa có migration/API được phê duyệt.

## 6. Security & authorization

- JWT required on every wallet endpoint.
- Repository queries must always include authenticated `owner_id`; never trust an owner ID from the request body.
- Validate type against the three supported values. Không tự bổ sung lệnh cấm số dư âm cho basic/goal: giới hạn này chưa được nghiệp vụ chốt. Credit phải biểu diễn được cả dư nợ và trả dư.
- Treat deletion as destructive and require explicit confirmation in the UI.

## 7. Verification plan

### Test cases for the first implementation slice

| ID | Scenario | Expected result |
| --- | --- | --- |
| DB-01 | Start API against an existing database containing `app_users` | Table is renamed to `user`; existing rows and IDs remain readable. |
| DB-02 | Restart API with an already migrated database | No duplicate tables, columns, or system categories are created. |
| DB-03 | Inspect system catalog after startup | VND wallet schema exists and system roots include expense, income, and debt groups; child parent links are valid. |
| DB-04 | Seed runs twice | Same `system_key` keeps one row and preserves the stable ID. |
| DB-05 | Create a wallet with `basic`, `goal`, or `credit` type | Type is stored; currency defaults to `VND`. |
| DB-06 | Try an unsupported wallet type or currency | Validation rejects the input; no row is written. |
| DB-07 | Query a wallet by another user | Repository returns not found and does not disclose the other user's row. |
| DB-08 | Assign a system/custom group to wallets | Junction rows are unique per group-wallet pair; duplicate assignment is harmless. |
| DB-09 | Delete a wallet after confirmation | Wallet row is removed; no archive state is introduced. |
| DB-10 | Delete without confirmation | Request is rejected and wallet remains unchanged. |

- Unit: type validation, VND validation, total calculation, ownership checks.
- Integration: GORM CRUD against dev PostgreSQL; user A cannot read/update/delete user B's wallet.
- API: auth failures, validation errors, pagination/list shape and destructive-delete response.
- E2E/UAT: create each wallet type, edit balance/name, toggle inclusion in total and review the delete warning.
- Swagger: add annotations and generated docs only together with each implemented endpoint.

## 8. Reconciliation plan

- Update `docs/architecture/API.md` and generated Swagger for implemented routes only.
- Update `docs/work/VALIDATION_MATRIX.md` when the first endpoint is verified.
- Update `docs/CONTEXT.md`, backlog status and changelog after implementation.
- Create an ADR only if the owner chooses a durable deletion/cascade policy or a schema strategy that changes this design.
