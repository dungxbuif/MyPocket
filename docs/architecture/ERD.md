---
artifact_type: erd_master
id: ERD-MASTER
status: draft
owner: shared
human_fields: [data_ownership_decisions, migration_approval]
ai_fields: [entities, relationships, constraints, migrations, linked_decisions]
shared_fields: [status, trace]
---

# ERD

## Implemented calendar, jar, and month schema — migrations 000013–000015

Migration `000013` adds `user.timezone` and `timezone_confirmed`, converts `wallets.target_date` to `date`, and replaces budget timestamp bounds with account-calendar `start_date`/`end_date` date columns while preserving existing labels. Migration `000014` adds stable owner-scoped `jars`, initialized month records (`jar_months`), per-month configurations (`jar_month_configs`), and nullable `transactions.jar_id` with a composite same-owner FK. Migration `000015` adds `month_notes`, uniquely keyed by owner and first day of month. Migration `000016` adds nullable `transactions.transfer_id` for atomic paired wallet transfers. Local dev DB is at migration 16, `dirty=false`; repository integration tests use PostgreSQL. Month figures are derived from live ledger data; no report snapshot/close table or cron job exists. [ADR-008](../decisions/ADR-008-account-timezone-and-calendar-dates.md) is proposed; [CORE-03](../work/tickets/CORE-03-TIME-JARS-MONTH-DETAIL_DESIGN.md) is in progress pending UAT and reconciliation.

## Implemented AI entry schema — migrations 000010–000012

Migration `000010` created `ai_entry_sessions`, `ai_entry_messages`, `ai_entry_proposals`, and `ai_entry_requests`. The one-shot API uses the owner-scoped session row internally for process/idempotency only; it writes no messages or exposes history. `ai_entry_messages` is legacy and unused. Migrations `000011`–`000012` add `transaction_attachments` (private object key, owner/process, source metadata, OCR text/state and cleanup deadline) and unique `(transaction_id,attachment_id)` links, plus safe cleanup claim state. Approval writes links in the same SQL transaction as the ledger row; owner/link-scoped download is implemented. Those migrations are part of the current clean local version 15; real DB repository tests pass. Explicit cleanup command is implemented but not run against bucket contents. Live S3/OCR and browser download UAT remain pending. [ADR-005](../decisions/ADR-005-ai-entry-review.md), [ADR-006](../decisions/ADR-006-private-ai-attachments.md), [AI-ENTRY-02](../work/tickets/AI-ENTRY-02-BATCH-ATTACHMENTS-DETAIL_DESIGN.md).

## Implemented budget schema — migrations 000009 and 000013

`budgets`: text UUID id, owner_id FK user, name, positive safe-integer limit_amount, nullable wallet_id/category_id FKs, account-calendar `start_date`/`end_date` SQL `date`, timestamps. Owner/period index. Delete owner/wallet/category cascades matching budget configuration; deleting budget never removes transactions. Spent/remaining days/ended are derived fields, not stored counters. Existing labels were backfilled in `000013` using the existing-account timezone. [ADR-004](../decisions/ADR-004-budget-api-data.md), [ADR-008](../decisions/ADR-008-account-timezone-and-calendar-dates.md), [API design](../work/tickets/API-SCREENS-01-DETAIL_DESIGN.md). No AutoMigrate is used.

## Field Ownership

- Human owns data ownership and migration approval.
- AI maintains entities, relationships, constraints, migrations, and linked decisions.

## Entities

## ASCII — mô hình ví và nhóm để review

`[E]` = đã có trong Go model/migration hiện tại; `[P]` = đề xuất, chưa có bảng/migration.
`1 --- 0..N` = một bản ghi có thể liên kết không hoặc nhiều bản ghi phía còn lại.
Sơ đồ chỉ liệt kê khoá và các trường cần đọc quan hệ, không phải toàn bộ schema triển khai.

Owner chốt tên bảng là `user`. Entity tài khoản đã có; đổi tên bảng sang `user` chưa được migration, nên sơ đồ đánh dấu `[R]` (rename pending).

```text
                      +----------------------------+
| user [E]                   |
                      | PK id                      |
                      | name, email                |
                      +----------------------------+
                           1                 1
                           |                 |
                        0..N              0..N
                           |                 |
       +---------------------------+  +---------------------------+
       | wallets [E]               |  | categories [E]            |
       | PK id                     |  | PK id                     |
       | FK owner_id -> user       |  | FK owner_id -> user       |
       | name, type, is_in_total   |  | FK parent_id -> categories|
       | currency = VND            |  | name, kind, system_key    |
       +---------------------------+  +---------------------------+
                    1                              1
                    |                              |
                 0..N                           0..N
                    |                              |
                    +--------+           +---------+
                             |           |
                    +----------------------------+
                    | category_wallets [E]       |
                    | FK category_id             |
                    | FK wallet_id               |
                    | UNIQUE(category_id,        |
                    |        wallet_id)          |
                    +----------------------------+

       categories (parent) 0..1 ----- 0..N categories (child)
       Parent: parent_id = NULL. Child: parent_id = parent.id.
       Maximum depth: parent -> child (2 levels).
```

- Ví thuộc một tài khoản. `type` là `basic`, `goal`, `credit`; chưa mở rộng sơ đồ sang sao kê/ledger khi thiết kế phần đó chưa hoàn tất.
- Một nhóm áp dụng cho nhiều ví và một ví dùng nhiều nhóm, qua liên kết `category_wallets`. API validate mọi `wallet_id` thuộc owner của nhóm trước khi thay thế liên kết.
- Đề xuất `categories` chứa bản nhóm theo tài khoản; catalog app cũ là nguồn seed với `system_key` ổn định. Chưa chốt giữa sao chép catalog theo account và catalog chung + override; sơ đồ trình bày phương án theo account để review, không xác nhận đã duyệt kiến trúc seed.
- `kind` là phân loại nghiệp vụ của catalog cũ. Chưa coi đây là ý nghĩa đã chốt của trường UI “category”.
- Khi triển khai phải ràng buộc nhóm, nhóm cha và ví liên kết cùng owner ở use case và DB; hai foreign key riêng lẻ chưa đủ chống liên kết chéo tài khoản. Không tự chọn mình làm cha, không tạo vòng lặp hoặc cấp ba.
- Không có dòng trong `category_wallets` nghĩa là nhóm chưa giới hạn theo ví. Ví mới không tự nhận liên kết từ nhóm hiện có.
- Xoá ví đã chốt xoá thực sau xác nhận. Không tự đặt cascade xoá danh mục toàn account từ liên kết ví; quy tắc xoá nhóm và ảnh hưởng lịch sử phải thiết kế riêng.

Nguồn: [UI Quản lý nhóm](../design/system/DESIGN.md#account--quản-lý-nhóm), [ticket nhóm](../work/tickets/TICKET-01-03-tong-vi-danh-muc.md), [thiết kế ví](../work/tickets/TICKET-01-02-DETAIL_DESIGN.md).

## Danh sách thực thể

| Entity/Table | Purpose | Owner | Notes |
| --- | --- | --- | --- |
| `user` | Authenticated account/user | User | Migration `000001` preserves legacy `app_users` IDs when renaming. |
| `wallets` | User-owned basic/goal/credit wallet | User | Migrated; VND enforced for this stage. |
| `categories` | System or user-owned group | System/User | Migrated; owner-approved system catalog is seeded through `000004`. |
| `category_wallets` | Quan hệ nhóm–ví áp dụng | Cùng user với nhóm/ví | Migrated; owner consistency remains use-case enforced. |
| `transactions` | Owner-scoped income/expense ledger entry | User | Migrated; belongs to one wallet and optional visible category. Nullable `transfer_id` links the two rows of an internal transfer. |
| `jars` | Stable owner-scoped jar identity | User | Migration `000014`; month-specific names/allocation live in `jar_month_configs`. |
| `jar_months` | Marks a month whose jar configuration has been initialized | User/month | Migration `000014`; serializes one-time prior-month copy, including empty initialization. |
| `jar_month_configs` | Name, allocation rule and active state for one jar/month | User/month/jar | Migration `000014`; historical rows remain when a later month changes. |
| `month_notes` | User-authored note for one account-local month | User/month | Migration `000015`; independent from calculated report figures. |
| `transaction_attachments` | Private S3 receipt metadata and OCR text/state | User/process | Migration `000011`–`000012`; object key is backend-only; 24-hour unlinked cleanup lifecycle. |
| `transaction_attachment_links` | Approved transaction-to-receipt relation | User via both FKs | Unique transaction/attachment pair; only created inside proposal approval. |

## Relationships

| From | To | Relationship | Notes |
| --- | --- | --- | --- |
| `user` | `wallets` | 1-to-many | `wallets.owner_id` references `user.id`. |
| `user` | `categories` | 1-to-many | `categories.owner_id` references `user.id` for personal groups. |
| `categories` | `categories` | Parent 0..1 / children 0..N | Tối đa hai cấp. |
| `categories` | `wallets` | Many-to-many | Qua `category_wallets`; use case xác minh liên kết cùng owner. |
| `user` | `jars`, `month_notes`, `jar_months` | 1-to-many | Rows are owner-scoped; month and note keys use the first calendar day. |
| `jar_months` | `jar_month_configs` | 1-to-many | Composite `(owner_id, month)` FK; a month is copied/initialized once. |
| `jars` | `jar_month_configs` | 1-to-many | Composite `(owner_id, jar_id)` FK prevents cross-owner configuration. |
| `jars` | `transactions` | 1-to-many optional | Composite owner/jar FK; use case additionally validates ordinary-expense kind and month configuration. |
| `transactions` | `transactions` | Paired 1-to-1 by nullable `transfer_id` | Source expense and destination income share a transfer id; both rows are created atomically and excluded from reports/jars. |
| `transactions` | `transaction_attachments` | Many-to-many | Qua `transaction_attachment_links`; transaction owner and attachment owner are both checked before signed download. |

## Constraints

- `user.id` là khoá chính đích, giữ nguyên ID tài khoản hiện có khi đổi tên bảng.
- `wallets.owner_id` is required and indexed; every repository query scopes by owner.
- `wallets.type` accepts `basic`, `goal`, or `credit`.
- `wallets.currency` is fixed to `VND` for the current stage.
- `user.timezone` must be a valid IANA zone. Instants remain UTC `timestamptz`; calendar dates and month keys have no time-of-day or implicit UTC conversion.
- Jar/month configuration and note repositories are owner-scoped. Jar spend, monthly totals and completion state are derived; they are not stored as counters or immutable close snapshots.
- Xoá thực sau xác nhận đã được owner chốt, kể cả ví có giao dịch. Migration `000003` cascade xoá transaction thuộc ví; không cascade xoá danh mục toàn account từ liên kết ví. Không có unique constraint tên ví. Chi tiết tác động tới ví đối ứng cần thiết kế cùng giao dịch liên kết.

## Migrations

- Versioned SQL migrations are active. [`DATABASE.md`](DATABASE.md) defines the dev/prod command and migration ledger; API startup does not invoke `AutoMigrate`.
- Applied migration sequence: `000013_account_timezone_and_calendar_dates`, `000014_jars`, `000015_month_notes`, `000016_transaction_transfer_links`; local migration ledger reports version 16, clean.
- Rollbacks of `000014`/`000015` refuse to drop user jar/note data. Do not run destructive down migrations against populated environments. `000013` rollback retains account timezone preferences.
- Adjustment ledger, credit/debt semantics, and recurring execution remain separate designs. Paired internal transfers are implemented through migration `000016`; pair edit/delete and replay idempotency remain follow-up work.

## Linked Decisions

- [ADR-008 — Account timezone and calendar dates](../decisions/ADR-008-account-timezone-and-calendar-dates.md) (proposed; implementation evidence recorded, owner decision status remains human-owned).
