---
artifact_type: architecture_doc
id: ARCH-MASTER
status: active
owner: shared
human_fields: [approved_boundaries, architectural_constraints, tradeoff_approval]
ai_fields: [overview, modules, diagrams, flows, dependencies, risks]
shared_fields: [status, linked_decisions]
updated: 2026-08-31
---

# Architecture

## Field Ownership

- Human-approved boundaries: React PWA, Go modular backend and worker, PostgreSQL, S3-compatible storage, Google OAuth, review-first ingestion, and explicit offline conflicts.
- AI maintains implementation-grounded modules, flows, dependencies, and risks.

## Overview

MyPocket is a modular monolith deployed as three application processes: a static React TypeScript PWA, a Go HTTP API, and a Go worker. The API and worker share domain/application packages. PostgreSQL is authoritative; Redis accelerates API key/session auth lookups; IndexedDB provides offline state and an outbox; S3-compatible storage keeps private receipt objects.

```text
+----------------------+       HTTPS REST/JSON       +----------------------+
| React PWA            | --------------------------> | Go API               |
| Router / Query       | <-------------------------- | Modular monolith     |
| IndexedDB / Outbox   |       sync/change feed      +----------+-----------+
+----------------------+                                         |
                                                                 | shared packages
                                                      +----------v-----------+
                                                      | Go Worker            |
                                                      | leased background jobs|
                                                      +----------+-----------+
                                                                 |
             +----------------------+----------------------------+---------------------+
             |                      |                            |                     |
             v                      v                            v                     v
      +-------------+       +--------------+           +----------------+    +----------------+
      | PostgreSQL  |       | S3-compatible|           | Google OAuth   |    | AI / OCR / Push|
      | source truth|       | private files|           | identity only  |    | provider adapters|
      +-------------+       +--------------+           +----------------+    +----------------+
```

## System Boundaries

- The browser owns presentation, local IndexedDB state, offline outbox scheduling, and interactive conflict resolution.
- The Go API owns authentication, authorization, validation, business operations, sync reconciliation, provider orchestration, and public contracts.
- The Go worker owns leased background work: recurring drafts, thresholds, push delivery, retries, deletion, audit retention, and maintenance.
- PostgreSQL owns authoritative relational state and atomicity.
- Redis owns short-lived auth lookup cache only; cache misses and outages fall back to PostgreSQL-backed auth.
- S3-compatible storage owns private receipt bytes; PostgreSQL stores metadata and object keys.
- External Android automation is outside this repository and only calls the signed bank webhook.
- OpenAI-compatible and OCR services extract proposals; they never own or confirm financial records.

## Modules

| Module | Responsibility | Key Files | Notes |
| --- | --- | --- | --- |
| identity | Google callback, user provisioning, signed cookie, current-user context | `backend/internal/identity/` | No session table or Google token persistence |
| finance | Wallets, categories, transactions, transfers, adjustments, attachments | `backend/internal/finance/` | Authoritative accounting boundary |
| portfolio | Gold, stock, crypto, foreign-currency, and other market-valued assets | `backend/internal/portfolio/` | Separate from wallet accounting; stores positions, trades, price history, and valuation formulas |
| sync | Idempotency, versions, cursors, tombstones, conflicts | `backend/internal/sync/` | Calls finance/portfolio application services for offline replay |
| planning | Budgets, events, recurring schedules, debts/loans | `backend/internal/planning/` | Recurring occurrences produce drafts |
| analytics | Net worth and report aggregations | `backend/internal/analytics/` | Reads confirmed reportable transactions |
| ingestion | Drafts, AI chat, OCR, multimodal images, bank webhooks | `backend/internal/ingestion/` | Shared `TransactionDraft` contract |
| notification | In-app inbox and Web Push | `backend/internal/notification/` | Inbox remains authoritative when push fails |
| export | Manual CSV/Sheets-compatible snapshots | `backend/internal/export/` | No spreadsheet import |
| audit | Append-only audit writes and restricted queries | `backend/internal/audit/` | Viewer email is environment-configured |
| platform | HTTP, config, database, object store, logging, jobs | `backend/internal/platform/` | Provider interfaces and adapters |
| web app | Mobile PWA workflows and finance screens | `frontend/src/app/` | Online path uses the Go API; offline write acceptance depends on the offline adapter |
| web offline | Browser mirror, local outbox, migration, and degraded storage handling | `frontend/src/offline/` | IndexedDB stores wallets, categories, transactions, assets, outbox records, tombstones, conflicts, and sync meta |

## Data Flow

1. Online and offline UI actions create client mutation IDs and optimistic local state.
2. The sync API validates idempotency, ownership, base versions, and domain rules.
3. Domain writes, accounting effects, audit records, and sync changes commit transactionally where required.
4. Stale offline writes return explicit conflict payloads; the PWA stores the conflict locally and requires keep-server, discard-local, or edit-and-retry before removing the pending mutation.
5. The client pulls authoritative changes using a per-user cursor, or requests a full resync snapshot when local mirror state is stale or cleared.
6. AI/OCR/webhook/recurring sources create drafts; confirmation invokes the same finance service used by manual entry.
7. Worker jobs acquire PostgreSQL leases before creating occurrences, sending push, refreshing configured portfolio prices, retrying providers, or purging audit rows.

## Runtime Flow

- Web assets are served over HTTPS and call the API on an approved origin.
- The browser caches the last successful current-user envelope only to keep authenticated IndexedDB data visible when `navigator.onLine` is false; logout clears the cached user and local offline stores.
- API and worker connect to the homelab PostgreSQL instance using environment configuration.
- API and worker startup currently requires a successful PostgreSQL ping through `DATABASE_URL`.
- Receipt upload uses short-lived presigned URLs and private S3 objects; the platform layer now includes an S3-compatible smoke adapter for endpoint/bucket access verification.
- Structured logs and API errors share a correlation ID.
- Portfolio price refresh uses a provider boundary; current UAT supports static configured quotes via `PORTFOLIO_STATIC_PRICES_JSON`, while manual price entry remains available for every asset.
- API liveness does not depend on optional AI/OCR/Push providers; readiness reflects required database/config state.

## Dependencies

- React, TypeScript, Vite, React Router, TanStack Query, PWA service worker, and IndexedDB.
- Go HTTP, PostgreSQL driver/query layer, migrations, OAuth/OIDC primitives, S3 SDK, and Web Push implementation.
- PostgreSQL and S3-compatible object storage.
- Google OAuth, OpenAI-compatible text/multimodal API, external OCR API, and Web Push endpoints.

## Risks And Tradeoffs

- Stateless cookies simplify storage but cannot support server-side global logout or per-device revocation.
- Full offline support adds conflict, cursor, tombstone, migration, and browser-storage complexity.
- A modular monolith reduces homelab operations while requiring strict package-boundary discipline.
- External provider availability affects draft creation but must never affect confirmed accounting correctness.
- Append-only audit data helps debugging but increases privacy exposure; redaction, restricted access, and 180-day default retention limit that risk.

## Linked Decisions

- [ADR-001](../decisions/ADR-001-react-go-modular-monolith.md)
- [ADR-002](../decisions/ADR-002-stateless-google-oauth.md)
- [ADR-003](../decisions/ADR-003-offline-sync-conflict-review.md)
- [ADR-004](../decisions/ADR-004-review-first-ingestion.md)
- [ADR-005](../decisions/ADR-005-restricted-audit-log.md)
