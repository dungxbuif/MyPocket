---
artifact_type: adr
id: ADR-001
status: accepted
owner: shared
trace:
  backlog: ../work/BACKLOG.md#bl-005
  tickets:
    - ../work/tickets/TICKET-01-02-quan-ly-vi.md
    - ../work/tickets/TICKET-02-01-ghi-thu-chi.md
  architecture:
    - ../architecture/DATABASE.md
    - ../architecture/ERD.md
---

# ADR-001: Versioned SQL migrations own database schema

## Context

Wallets, groups and transactions introduce durable PostgreSQL data. GORM
`AutoMigrate` at API startup would change dev or production schema as a side
effect of serving traffic and cannot provide an auditable migration history.

## Decision

Use `golang-migrate` and ordered SQL files under `backend/migrations/`. The
same `go run ./cmd/migrate up` command is used for development and production.
The API process only connects to an already-migrated database. System category
seeding is migration data, not an API startup side effect.

## Alternatives considered

| Alternative | Decision |
| --- | --- |
| Keep GORM `AutoMigrate` in `cmd/api` | Rejected: implicit, unauditable and runtime-coupled schema changes. |
| Add a separate schema-management service | Rejected for current scope: a migration CLI is sufficient and simpler to operate. |

## Consequences

- Deployments must run migrations before starting a new API version.
- Each schema/data change requires an immutable forward migration and migration proof.
- Initial schema/category migration is non-destructive on down; recovery uses backup/restore, avoiding accidental user-data deletion.
