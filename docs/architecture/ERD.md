---
artifact_type: erd_master
id: ERD-MASTER
status: active
owner: shared
human_fields: [data_ownership_decisions, migration_approval]
ai_fields: [entities, relationships, constraints, migrations, linked_decisions]
shared_fields: [status, trace]
updated: 2026-08-31
---

# ERD

## Field Ownership

- Human-approved ownership: each finance object belongs to one application user; PostgreSQL is authoritative.
- AI maintains planned tables, relationships, constraints, and migration links.
- Concrete column definitions remain subject to the owning phase's approved detail design and migration review.

## Entities

| Entity/Table | Purpose | Owner | Notes |
| --- | --- | --- | --- |
| `users` | Google-backed application identity | Identity | Unique provider subject and verified email; no provider tokens |
| `wallets` | Cash, bank, credit, e-wallet, savings, or debt account | User | VND-first; optional credit metadata; versioned/archiveable |
| `categories` | Two-level expense, income, or debt/loan taxonomy | User/system seed | System rows locked; user rows archiveable |
| `wallet_category_settings` | Category activation per wallet | User | Composite wallet/category uniqueness |
| `transactions` | Confirmed accounting event | User | Positive integer amount; type determines balance effect |
| `transaction_drafts` | Reviewable proposal from manual automation/provider flow | User | Source and confirmation link; no accounting effect |
| `receipt_objects` | Private object metadata | User | S3 key, checksum, size, content type |
| `budgets` | Period category/all-category limits | User | Weekly/monthly/quarterly/yearly/custom periods; versioned/archiveable |
| `budget_categories` | Selected expense category scope for a budget | User | Composite budget/category uniqueness |
| `budget_alerts` | Durable 80%/100% budget threshold event | User | Unique budget/threshold/period dedupe |
| `events` | Trip or event grouping | User | Optional transaction relationship |
| `recurring_schedules` | Template and next occurrence | User | Worker emits deterministic drafts |
| `debts` | Borrowed/lent obligation | User | Related repayment transactions |
| `notifications` | Durable in-app notice | User | Read state and related entity |
| `push_subscriptions` | Browser Web Push subscription | User | Encrypted keys and endpoint |
| `sync_mutations` | Idempotency ledger | User | Unique client mutation ID |
| `sync_changes` | Incremental authoritative change feed | User | Monotonic per-user cursor |
| `sync_conflicts` | Rejected client mutation and server snapshot | User | Explicit resolution lifecycle |
| `webhook_sources` | User-owned bank notification sender | User | Encrypted HMAC material and enable state |
| `webhook_events` | Replay/deduplication ledger | User/source | Unique source event and payload digest |
| `ai_conversations` | AI chat container | User | Provider-independent metadata |
| `ai_messages` | Text/image-reference message | User/conversation | Links structured draft results |
| `audit_logs` | Append-only state/security event | System | Redacted; restricted read; retention-managed |
| `job_leases` | Worker ownership/idempotency state | System | Bounded leases and occurrence keys |
| `exports` | Manual export job and object result | User | Expiring snapshot metadata |

## Relationships

| From | To | Relationship | Notes |
| --- | --- | --- | --- |
| `users` | all user-owned tables | one-to-many | Ownership is required and indexed |
| `categories` | `categories` | optional parent-to-many children | Database and service enforce depth <= 2 |
| `wallets` | `wallet_category_settings` | one-to-many | Settings reference same user's category |
| `wallets` | `transactions` | one-to-many | Source wallet required |
| `transactions` | `wallets` | optional destination | Required only for transfer |
| `transactions` | `categories` | many-to-one | Type and activation must match |
| `transactions` | `events` | optional many-to-one | Event totals do not alter base accounting |
| `transactions` | `receipt_objects` | optional many-to-one | Private attachment |
| `transaction_drafts` | `transactions` | optional one-to-one confirmation | Idempotent confirmation |
| `ai_conversations` | `ai_messages` | one-to-many | Ordered messages |
| `ai_messages` | `transaction_drafts` | one-to-many | Multi-transaction result |
| `webhook_sources` | `webhook_events` | one-to-many | Replay and deduplication scope |
| `webhook_events` | `transaction_drafts` | optional one-to-one | Accepted event result |
| `users` | `sync_changes` | one-to-many ordered cursor | Pull scope is per user |

## Constraints

- Every user-owned foreign-key relationship must reference objects with the same `user_id`; enforce in service logic and database constraints where practical.
- Money is a positive integer; transaction type and wallet role determine signs.
- Transfer source and destination wallets differ and belong to one user.
- At most one active default AI wallet exists per user.
- Category depth is at most two; system categories cannot be renamed or deleted.
- Mutation IDs, confirmation commands, recurring occurrences, and source webhook event IDs are unique in their intended scope.
- Version increments occur only with accepted authoritative mutations.
- Audit rows have no application update/delete endpoint; retention deletion is worker-only.
- Rows referenced by financial history are archived or tombstoned rather than cascade-deleted.
- Account deletion is an explicit background lifecycle, not ordinary cascading API behavior.

## Migrations

- Migrations are ordered, immutable after release, and applied before the API/worker version that depends on them.
- Each phase owns forward migration tests and records any rollback limitation.
- Seed migrations use stable IDs for the approved Vietnamese category taxonomy.
- PostgreSQL integration tests start from an empty database and apply the full migration chain.
- Production backup/restore proof is required before account-deletion and retention jobs are released.

## Implemented Schema

### `schema_migrations`

| Column | Type | Notes |
| --- | --- | --- |
| `version` | `text` | Primary key; migration filename |
| `checksum` | `text` | SHA-256 checksum; mismatch fails startup/test migration |
| `applied_at` | `timestamptz` | Defaults to `now()` |

### `users`

| Column | Type | Notes |
| --- | --- | --- |
| `id` | `uuid` | Primary key; defaults through `gen_random_uuid()` from `pgcrypto` |
| `google_subject` | `text` | Required and unique |
| `email` | `text` | Required |
| `email_verified` | `boolean` | Required; defaults false |
| `display_name` | `text` | Required; defaults empty string |
| `avatar_url` | `text` | Required; defaults empty string |
| `created_at` | `timestamptz` | Required; defaults to `now()` |
| `updated_at` | `timestamptz` | Required; defaults to `now()` |

The implemented identity schema intentionally has no `google_access_token`, `google_refresh_token`, or `session_id` columns.

### `wallets`

| Column | Type | Notes |
| --- | --- | --- |
| `id` | `uuid` | Primary key; defaults through `gen_random_uuid()` |
| `user_id` | `uuid` | Required owner; references `users(id)` |
| `name` | `text` | Required non-empty wallet name |
| `type` | `text` | `cash`, `bank`, `credit`, `e_wallet`, `savings`, or `debt` |
| `balance_vnd` | `bigint` | Required integer VND balance; defaults `0` |
| `include_in_total` | `boolean` | Controls total balance inclusion |
| `is_default_ai` | `boolean` | Partial unique index allows one active default AI wallet per user |
| `credit_limit_vnd`, `statement_day`, `payment_due_day` | nullable numeric fields | Allowed only for `credit` wallets |
| `archived_at` | `timestamptz` | Archive marker |
| `version` | `bigint` | Optimistic version seed; starts at `1` |
| `created_at`, `updated_at` | `timestamptz` | Audit timestamps |

### `categories`

| Column | Type | Notes |
| --- | --- | --- |
| `id` | `uuid` | Primary key; system seed rows use stable UUIDs |
| `user_id` | `uuid` | Nullable owner; `NULL` only for system categories |
| `parent_id` | `uuid` | Optional parent category; references `categories(id)` |
| `kind` | `text` | `expense`, `income`, or `debt` |
| `name` | `text` | Required non-empty Vietnamese/user label |
| `system_key` | `text` | Unique stable key for system rows; user rows keep `NULL` |
| `is_system` | `boolean` | Locks seeded category ownership and key semantics |
| `archived_at` | `timestamptz` | User category archive marker |
| `version` | `bigint` | Optimistic version seed; starts at `1` |
| `created_at`, `updated_at` | `timestamptz` | Audit timestamps |

Seeded system keys include `expense_food`, `expense_shopping`, `expense_transport`, `income_salary`, `income_bonus`, and `debt_loan`.

### `wallet_category_settings`

| Column | Type | Notes |
| --- | --- | --- |
| `wallet_id` | `uuid` | References `wallets(id)`; part of composite primary key |
| `category_id` | `uuid` | References `categories(id)`; part of composite primary key |
| `user_id` | `uuid` | Required owner scope |
| `active` | `boolean` | Wallet/category availability flag |
| `created_at`, `updated_at` | `timestamptz` | Audit timestamps |

### `receipt_objects`

| Column | Type | Notes |
| --- | --- | --- |
| `id` | `uuid` | Primary key |
| `user_id` | `uuid` | Required owner scope |
| `object_key` | `text` | Private S3-compatible object key; unique per user |
| `content_type` | `text` | Required MIME type |
| `size_bytes` | `bigint` | Required positive object size |
| `checksum_sha256` | `text` | Required 64-character SHA-256 checksum |
| `original_filename` | `text` | Optional display filename; defaults empty |
| `created_at` | `timestamptz` | Creation timestamp |

### `transactions`

| Column | Type | Notes |
| --- | --- | --- |
| `id` | `uuid` | Primary key |
| `user_id` | `uuid` | Required owner scope |
| `type` | `text` | `income`, `expense`, `transfer`, or `adjustment` |
| `source_wallet_id` | `uuid` | Required source wallet |
| `destination_wallet_id` | `uuid` | Required only for transfer; must differ from source |
| `category_id` | `uuid` | Optional category reference |
| `receipt_object_id` | `uuid` | Optional private receipt metadata reference |
| `amount_vnd` | `bigint` | Required positive VND integer |
| `balance_after_vnd` | `bigint` | Snapshot for source wallet after posting |
| `source_delta_vnd`, `destination_delta_vnd` | `bigint` | Persisted balance effects used to reverse edits and archives exactly |
| `note`, `with_person`, `event_ref` | `text` | Optional search/display metadata; default empty |
| `occurred_at` | `timestamptz` | Required transaction date/time |
| `excluded_from_reports` | `boolean` | Report exclusion flag |
| `archived_at` | `timestamptz` | Archive marker |
| `version` | `bigint` | Optimistic version seed; starts at `1` |
| `created_at`, `updated_at` | `timestamptz` | Audit timestamps |

### `finance_idempotency_keys`

| Column | Type | Notes |
| --- | --- | --- |
| `user_id` | `uuid` | Required owner; part of primary key |
| `key` | `text` | Idempotency key; part of primary key |
| `request_hash` | `text` | Request identity hash |
| `response_status` | `integer` | Stored HTTP-equivalent response status |
| `response_json` | `jsonb` | Stored replay response |
| `created_at` | `timestamptz` | Creation timestamp |

### `sync_cursors`

| Column | Type | Notes |
| --- | --- | --- |
| `user_id` | `uuid` | Primary key; references `users(id)` |
| `next_cursor` | `bigint` | Next per-user cursor to allocate; must be positive |

### `sync_changes`

| Column | Type | Notes |
| --- | --- | --- |
| `user_id` | `uuid` | Required owner; part of primary key |
| `cursor` | `bigint` | Required positive per-user cursor; part of primary key |
| `entity_type` | `text` | `wallet`, `category`, or `transaction` |
| `entity_id` | `text` | Entity identifier preserved from the client/server command |
| `operation` | `text` | `create`, `update`, `archive`, `set_default_ai`, or `set_category_active` |
| `version` | `bigint` | Authoritative entity version or tombstone version |
| `payload_json` | `jsonb` | Authoritative payload snapshot or tombstone payload |
| `created_at` | `timestamptz` | Creation timestamp |

`sync_changes_user_entity_idx` supports per-user entity/cursor lookups.

### `sync_mutations`

| Column | Type | Notes |
| --- | --- | --- |
| `user_id` | `uuid` | Required owner; part of primary key |
| `mutation_id` | `text` | Client mutation ID; part of primary key |
| `request_hash` | `text` | Stable request identity hash used to reject mutation ID reuse with different content |
| `result_json` | `jsonb` | Stored mutation result for applied, conflict, or rejected replay |
| `created_at` | `timestamptz` | Creation timestamp |

## Linked Decisions

- [ADR-001](../decisions/ADR-001-react-go-modular-monolith.md)
- [ADR-003](../decisions/ADR-003-offline-sync-conflict-review.md)
- [ADR-004](../decisions/ADR-004-review-first-ingestion.md)
- [ADR-005](../decisions/ADR-005-restricted-audit-log.md)
