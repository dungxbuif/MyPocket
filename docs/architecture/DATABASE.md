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

The dedicated native fixture database on local55434 is verified version `32`, `dirty=false`;
the stage-v1 baseline and later additive migrations own the complete schema.

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
