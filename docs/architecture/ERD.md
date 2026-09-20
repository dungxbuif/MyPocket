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

## Implemented budget schema — migration 000009

`budgets`: text UUID id, owner_id FK user, name, positive safe-integer limit_amount, nullable wallet_id/category_id FKs, start_at/end_at timestamptz with end > start, timestamps. Owner/period index. Delete owner/wallet/category cascades matching budget configuration; deleting budget never removes transactions. Spent/remaining days/ended are derived fields, not stored counters. [ADR-004](../decisions/ADR-004-budget-api-data.md), [API design](../work/tickets/API-SCREENS-01-DETAIL_DESIGN.md). Dev DB migrated successfully to version 9; no AutoMigrate introduced. Historical rename-pending note below is superseded by explicit migrations already applied.

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
| `transactions` | Owner-scoped income/expense ledger entry | User | Migrated; belongs to one wallet and optional visible category. |

## Relationships

| From | To | Relationship | Notes |
| --- | --- | --- | --- |
| `user` | `wallets` | 1-to-many | `wallets.owner_id` references `user.id`. |
| `user` | `categories` | 1-to-many | `categories.owner_id` references `user.id` for personal groups. |
| `categories` | `categories` | Parent 0..1 / children 0..N | Tối đa hai cấp. |
| `categories` | `wallets` | Many-to-many | Qua `category_wallets`; use case xác minh liên kết cùng owner. |

## Constraints

- `user.id` là khoá chính đích, giữ nguyên ID tài khoản hiện có khi đổi tên bảng.
- `wallets.owner_id` is required and indexed; every repository query scopes by owner.
- `wallets.type` accepts `basic`, `goal`, or `credit`.
- `wallets.currency` is fixed to `VND` for the current stage.
- Xoá thực sau xác nhận đã được owner chốt, kể cả ví có giao dịch. Migration `000003` cascade xoá transaction thuộc ví; không cascade xoá danh mục toàn account từ liên kết ví. Không có unique constraint tên ví. Chi tiết tác động tới ví đối ứng cần thiết kế cùng giao dịch liên kết.

## Migrations

- Versioned SQL migrations are active. [`DATABASE.md`](DATABASE.md) defines the dev/prod command and migration ledger; API startup does not invoke `AutoMigrate`.
- Wallet migration (after design approval): add `wallets` with UUID primary key, owner foreign key/index, type, name, opening balance, currency, total-inclusion flag, description and timestamps.
- Migration vẫn chờ detail design được review: bổ sung trường riêng goal/credit và xác định ledger cho số dư khởi tạo/điều chỉnh trước khi chốt schema. Chỉ tạo migration cho phần triển khai; việc xoá giao dịch hai vế cần ma trận ảnh hưởng riêng.
- Rollback: disable wallet routes and retain the additive table; do not drop user data automatically.

## Linked Decisions

- TBD
