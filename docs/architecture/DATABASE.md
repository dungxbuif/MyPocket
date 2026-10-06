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
| `000002` | Seeds the reviewed Vietnamese system category catalog. | `system_key` conflict handling keeps one stable row per catalog item. |
| `000003` | Makes wallet deletion cascade to its transaction rows. | Matches the owner-approved permanent-delete policy. |
| `000029` | Adds goal-wallet notification preference and owner-scoped in-app notifications with dedupe/read indexes. | Additive; existing wallets default to notifications off. |

The current disposable development database reaches version `29`, `dirty=false`;
the stage-v1 baseline and later additive migrations own the complete schema.
