---
artifact_type: database_operations
id: DB-OPERATIONS
status: active
owner: shared
trace:
  ticket: ../work/tickets/TICKET-02-01-ghi-thu-chi.md
  detail_design: ../work/tickets/TICKET-02-01-DETAIL_DESIGN.md
  adr: ../decisions/ADR-001-versioned-database-migrations.md
---

# Database migrations

PostgreSQL schema is changed only through ordered SQL files in
`backend/migrations/`. Gin API startup never calls GORM `AutoMigrate`, never
seeds system data, and never changes a schema implicitly.

## One command for development and production

Run from `backend/`, after the target environment supplies `DATABASE_URL`:

```sh
go run ./cmd/migrate up
go run ./cmd/migrate version
```

`MIGRATIONS_PATH` is optional when migrations are packaged outside the current
directory. The CLI applies only forward migrations. Rollback of initial schema
or category seeds is intentionally not automated because it could delete
referenced user data; restore from a database backup is the recovery path.

## Current migration history

| Version | Purpose | Safe repeat behavior |
| --- | --- | --- |
| `000001` | Renames legacy `app_users` to `user` when needed; creates user, wallets, categories, assignments and transactions schema. | The migration ledger records completion; `CREATE ... IF NOT EXISTS` supports existing dev tables during the first adoption. |
| `000002` | Adds AI entry error code. | Additive column with empty default. |
| `000003` | Adds optional feedback screenshot metadata/index. | Additive columns and partial index. |
| `000022` | Durable transaction mutation receipts. | Unique owner/operation/key; pending/completed JSON result. |
| `000029` | Adds goal-wallet notification preference and owner-scoped in-app notifications with dedupe/read indexes. | Additive; existing wallets default to notifications off. |
| `000030` | Adds category active flag. | Existing categories remain active. |
| `000031` | Adds protected device sign-in credentials. | Additive owner-linked table. |
| `000032` | Adds owner feedback resolution fields. | Existing feedback defaults unresolved. |
| `000033` | Adds advisor idle timestamps and the expiry scan index. | Backfills from existing user messages; no financial rows are deleted. |
| `000034` | Adds idempotent advisor reset requests. | Content-free owner ledger; no financial rows are changed. |
| `000035` | Adds versioned recurring budget series and generated-period state. | Additive; existing budgets and ledger rows remain unchanged. |
| `000036` | Adds owner-scoped iOS push-token registrations. | Additive; registration only, no APNs delivery or token response. |
| `000037` | Adds the redacted audit delivery outbox. | Additive; unique stream/event keys, leased retries, no financial content. |
| `000038` | Adds optimistic portfolio asset versions and immutable trade correction revisions. | Additive; original trades remain unchanged, corrections are owner-scoped and replayed from the full history. |
| `000039` | Adds durable owner-scoped APNs delivery hand-off rows. | Additive; deduped route payloads only, no provider call or token exposure. |
| `000040` | Adds optimistic-lock versions to ordinary transactions and personal categories. | Additive; existing rows default to version `1`; stale writes fail closed with a conflict. |

The dedicated native fixture database on local55434 is verified version `32`, `dirty=false`;
the stage-v1 baseline and later additive migrations own the complete schema.

The current unreleased candidate must advance a version-34 production database
through `000035` → `000036` → `000037` → `000038` → `000039` → `000040` in order.
The API image must be compatible with the schema before the worker is enabled.
Validate a custom
format backup and a clean migration version (`dirty=false`) before the change;
do not use `force` as a normal rollout or rollback mechanism. The read-only
checkout preflight is [`scripts/mypocket-rollout-preflight.sh`](../../scripts/mypocket-rollout-preflight.sh);
the operational sequence and rollback evidence are in
[`ROLLOUT_READINESS.md`](../operations/ROLLOUT_READINESS.md).

## Financial replay receipts

Migration22 transaction_mutations retains owner, operation, key, request hash,
pending/completed status, response JSON and created/updated timestamps; unique
scope is owner+operation+key. Transfer patch/delete and locally tested create
capabilities reuse it; create HTTP handlers now consume it for supplied valid keys
in locally verified source (not deployed). No-header legacy writes do not.
Reservation/domain
write/completion share one SQL transaction; rollback leaves no receipt or ledger
row. No Redis durability dependency/new offline migration. Do not overwrite30.
[Decision](../decisions/ADR-018-reuse-durable-financial-receipts.md).
