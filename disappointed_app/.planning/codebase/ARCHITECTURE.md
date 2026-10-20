<!-- refreshed: 2026-09-10 -->
# Architecture

**Analysis Date:** 2026-09-10

## System Overview

```text
+------------------------ Browser / Presentation -------------------------+
| React PWA shell       Feature screens       API/offline adapters        |
| `frontend/src/app/`   `frontend/src/screens/` `frontend/src/offline/`   |
+-------------------------------+-----------------------------------------+
                                | HTTPS REST/JSON + sync change feed
                                v
+-------------------------- Go Application -------------------------------+
| HTTP boundary                                                        |
| `backend/internal/platform/httpapi/`                                  |
|       |                                                               |
|       v                                                               |
| Domain repositories/services: finance, planning, portfolio, sync,     |
| analytics, identity, notification, audit                              |
| `backend/internal/{finance,planning,portfolio,sync,...}/`              |
|       |                                                               |
|       v                                                               |
| Shared platform transactions, change feed, DB, auth cache, objects    |
| `backend/internal/platform/`                                           |
+-------------------------------+-----------------------------------------+
                                |
              +-----------------+------------------+
              v                 v                  v
        PostgreSQL          Redis cache      S3-compatible storage
        authoritative       auth only        private receipts
              ^
              | shared domain repositories
+-------------+---------------- Go Worker --------------------------------+
| recurring drafts, notifications, price refresh, audit retention        |
| `backend/cmd/worker/`, `backend/internal/worker/`                       |
+--------------------------------------------------------------------------+
```

## Component Responsibilities

| Component | Responsibility | File |
|-----------|----------------|------|
| PWA bootstrap | Mount React, install router/query providers, register service worker | `frontend/src/main.tsx` |
| Application shell | Authentication gate, navigation, screen orchestration, sheets, conflict and notification UI | `frontend/src/app/App.tsx` |
| Feature screens | Render primary overview, transaction, budget, report, and account workflows | `frontend/src/screens/` |
| API clients | Define browser-side contracts and translate feature calls to `/api/v1` | `frontend/src/app/apiClient.ts`, `frontend/src/app/finance.ts`, `frontend/src/app/planning.ts` |
| Offline subsystem | Maintain user-scoped IndexedDB mirrors, mutation outbox, receipts, tombstones, and conflicts | `frontend/src/offline/` |
| HTTP composition root | Construct repositories, optional adapters, sync service, and router | `backend/cmd/api/main.go` |
| HTTP boundary | Route, authenticate, authorize, validate/translate, and serialize API traffic | `backend/internal/platform/httpapi/` |
| Domain modules | Own business types, validation, commands, persistence, and invariants | `backend/internal/finance/`, `backend/internal/planning/`, `backend/internal/portfolio/` |
| Synchronization | Replay mutations atomically, record receipts, emit changes, serve incremental/full snapshots | `backend/internal/sync/` |
| Shared command infrastructure | Bind repositories to one SQL transaction, enforce per-user write locking, append canonical changes | `backend/internal/platform/commandtx/transaction.go`, `backend/internal/platform/changefeed/change.go` |
| Worker composition root | Schedule recurring, notification, portfolio-price, and retention runners | `backend/cmd/worker/main.go` |
| Database migrations | Apply ordered embedded SQL migrations | `backend/cmd/migrate/main.go`, `backend/internal/platform/db/migrate.go`, `backend/migrations/` |

## Pattern Overview

**Overall:** Modular monolith with separate API and worker processes, repository-backed domain modules, and an offline-first PWA.

**Key Characteristics:**
- Keep business rules and persistence together in domain packages such as `backend/internal/finance/`; HTTP files are adapters, not the accounting boundary.
- Compose concrete repositories in `backend/cmd/api/main.go`, but expose narrow interfaces from `backend/internal/platform/httpapi/router.go` for handler isolation and testing.
- Treat PostgreSQL as the source of truth. IndexedDB in `frontend/src/offline/` is a user-scoped mirror and retry queue; Redis in `backend/internal/platform/authcache/` is only an acceleration layer.
- Route offline-capable writes through shared command transactions so domain state, `sync_changes`, and mutation receipts have one outcome.
- Reuse the same domain repositories from the API, sync service, draft confirmation, and worker paths rather than duplicating business logic.

## Layers

**Presentation and workflow layer:**
- Purpose: Own user interaction, screen state, online/offline selection, and optimistic feedback.
- Location: `frontend/src/app/`, `frontend/src/screens/`
- Contains: React components, query orchestration, feature API modules, shared UI primitives.
- Depends on: `frontend/src/offline/`, browser APIs, `/api/v1` contracts.
- Used by: `frontend/src/main.tsx`.

**Browser persistence and synchronization layer:**
- Purpose: Preserve local mirrors and pending mutations, reconcile server changes, and surface conflicts.
- Location: `frontend/src/offline/`
- Contains: IndexedDB schema/access, migrations, outbox commands, sync client, reducers, receipt upload queue.
- Depends on: Types from `frontend/src/app/finance.ts` and `frontend/src/app/portfolio.ts`, API fetch wrappers.
- Used by: `frontend/src/app/App.tsx` and compatibility adapters in `frontend/src/app/offline.ts`, `frontend/src/app/outbox.ts`.

**Transport layer:**
- Purpose: Enforce HTTP method/auth/CSRF/CORS concerns, decode requests, call domain interfaces, and map errors.
- Location: `backend/internal/platform/httpapi/`
- Contains: Router, middleware, resource handlers, health and embedded documentation handlers.
- Depends on: Domain types and repository interfaces, `backend/internal/platform/config/`.
- Used by: `backend/cmd/api/main.go`.

**Domain and persistence layer:**
- Purpose: Enforce ownership, versions, money rules, validation, and atomic persistence.
- Location: `backend/internal/finance/`, `backend/internal/planning/`, `backend/internal/portfolio/`, `backend/internal/identity/`, `backend/internal/notification/`, `backend/internal/audit/`, `backend/internal/analytics/`
- Contains: Input/entity types, pure rules, repository commands and queries, provider boundaries.
- Depends on: PostgreSQL and shared platform transaction/change-feed helpers.
- Used by: HTTP handlers, sync service, worker runners.

**Platform layer:**
- Purpose: Own technical adapters and cross-domain infrastructure.
- Location: `backend/internal/platform/`
- Contains: Config, DB opening/migration, HTTP, logging, Redis auth cache, S3 object store, SQL transaction and change-feed helpers.
- Depends on: External SDKs and standard library.
- Used by: API/worker composition roots and domain repositories.

**Background processing layer:**
- Purpose: Run bounded leased jobs outside request lifetimes.
- Location: `backend/internal/worker/`, `backend/internal/notification/worker.go`, `backend/cmd/worker/main.go`
- Contains: Recurring-draft, static portfolio-price, notification, and audit-retention runners.
- Depends on: Domain repositories and PostgreSQL leases.
- Used by: Worker process only.

## Data Flow

### Primary Request Path

1. `frontend/src/main.tsx` mounts `App`, whose screen or sheet calls a feature client such as `frontend/src/app/finance.ts`.
2. `frontend/src/app/apiClient.ts` sends authenticated JSON to a route registered in `backend/internal/platform/httpapi/router.go`.
3. Router middleware attaches correlation/auth context; the resource handler validates and calls a narrow repository interface in `backend/internal/platform/httpapi/`.
4. A domain command in `backend/internal/finance/commands.go`, `backend/internal/finance/repository.go`, `backend/internal/planning/repository.go`, or `backend/internal/portfolio/commands.go` enforces ownership and invariants inside PostgreSQL.
5. The handler serializes the authoritative result; the PWA refreshes or reconciles its rendered state.

### Offline Mutation and Reconciliation

1. A feature queues an optimistic entity and mutation through `frontend/src/offline/outbox.ts` and persists both via `frontend/src/offline/db.ts`.
2. `frontend/src/offline/syncApi.ts` posts pending mutations to `/api/v1/sync/mutations`.
3. `backend/internal/platform/httpapi/sync.go` delegates to `backend/internal/sync/service.go`.
4. The service binds sync plus finance/portfolio repositories to one transaction using `backend/internal/platform/commandtx/transaction.go`; domain state, change feed, and mutation receipt commit together.
5. The client removes applied items, retains stable conflicts through `frontend/src/offline/conflicts.ts`, and applies incremental changes or a full resync snapshot.

### Background Job Flow

1. `backend/cmd/worker/main.go` wakes once per minute and invokes bounded runners.
2. Runners in `backend/internal/worker/` and `backend/internal/notification/worker.go` acquire PostgreSQL-backed leases or due work through domain repositories.
3. Successful work writes domain state; failures are logged and best-effort audit events are appended through `backend/internal/audit/repository.go`.

**State Management:**
- Server truth is relational state under `backend/migrations/`; repositories scope reads/writes by authenticated `user_id`.
- React owns ephemeral UI state; server-derived data is loaded through feature clients and query orchestration.
- IndexedDB owns durable browser mirrors, outbox state, conflicts, tombstones, receipt uploads, and cursors in `frontend/src/offline/db.ts`.
- `frontend/src/app/userDataCache.ts` provides user-keyed local cache for selected read models; it is not authoritative.

## Key Abstractions

**Repository:**
- Purpose: Combine domain persistence and invariants behind interfaces consumed by HTTP, sync, and workers.
- Examples: `backend/internal/finance/repository.go`, `backend/internal/planning/repository.go`, `backend/internal/portfolio/repository.go`.
- Pattern: Concrete SQL repository with `NewRepository` composition and, where cross-module atomicity is required, `NewRepositoryInTx` binding.

**Command transaction:**
- Purpose: Share a PostgreSQL transaction and stable per-user lock ordering across domain and sync work.
- Examples: `backend/internal/platform/commandtx/transaction.go`, `backend/internal/finance/commands.go`, `backend/internal/portfolio/commands.go`.
- Pattern: Execute mutation callbacks with a transaction-bound handle, then append change-feed state before commit.

**HTTP dependency interface:**
- Purpose: Keep handlers dependent on the exact operations they call and allow tests to substitute fakes.
- Examples: `backend/internal/platform/httpapi/router.go`.
- Pattern: Interfaces live at the consuming transport boundary and concrete implementations are injected by `backend/cmd/api/main.go`.

**Offline mutation:**
- Purpose: Represent retryable, versioned, idempotent browser writes independently of connectivity.
- Examples: `frontend/src/offline/types.ts`, `frontend/src/offline/outbox.ts`, `backend/internal/sync/types.go`.
- Pattern: Client mutation ID plus entity/operation/base-version payload, with explicit pending, retryable, conflict, synced, and quarantined states.

**Provider boundary:**
- Purpose: Isolate optional external infrastructure from core business rules.
- Examples: `backend/internal/platform/objectstore/objectstore.go`, `backend/internal/portfolio/provider.go`, `backend/internal/notification/worker.go`.
- Pattern: Small interfaces with concrete S3/static/no-op adapters assembled at process startup.

## Entry Points

**Web application:**
- Location: `frontend/src/main.tsx`
- Triggers: Browser loading Vite-built assets.
- Responsibilities: Create React root, router, query client, render `App`, register service worker.

**HTTP API:**
- Location: `backend/cmd/api/main.go`
- Triggers: API container/process startup.
- Responsibilities: Load config/logging, connect PostgreSQL and optional adapters, compose repositories/router, listen for HTTP.

**Worker:**
- Location: `backend/cmd/worker/main.go`
- Triggers: Worker container/process startup.
- Responsibilities: Load dependencies, process bounded job cycles, respond to termination signals.

**Migration executable:**
- Location: `backend/cmd/migrate/main.go`
- Triggers: Explicit migration command/deployment step.
- Responsibilities: Open PostgreSQL and apply ordered SQL from `backend/migrations/`.

**Service worker:**
- Location: `frontend/public/sw.js`
- Triggers: Registration from `frontend/src/app/serviceWorker.ts`.
- Responsibilities: Cache installable application assets and serve offline-capable shell behavior.

## Architectural Constraints

- **Threading:** Go request handlers execute concurrently; mutable financial writes must serialize through database transactions and per-user locks in `backend/internal/platform/commandtx/transaction.go`. The browser uses the single-threaded event loop plus asynchronous IndexedDB/network operations.
- **Global state:** Avoid new mutable package globals. Process-wide configuration and dependencies are constructed in `backend/cmd/api/main.go` and `backend/cmd/worker/main.go`; browser durable state belongs in user-scoped IndexedDB databases managed by `frontend/src/offline/db.ts`.
- **Circular imports:** Go's compiler forbids cycles; preserve the dependency direction `cmd -> platform/httpapi -> domain -> platform primitives`. Keep cross-domain orchestration in `backend/internal/sync/` or a dedicated application service, not mutual domain imports.
- **Ownership:** Every database lookup and mutation must accept and filter by authenticated `userID`; do not accept ownership from request payloads. See `backend/internal/identity/ownership.go` and repository queries under `backend/internal/finance/`.
- **Money:** Store and calculate VND as checked `int64`; keep asset decimal quantities in the portfolio decimal boundary. See `backend/internal/finance/transactions.go`, `backend/internal/portfolio/decimal.go`.
- **Offline correctness:** Do not silently overwrite stale data. Preserve mutation IDs, base versions, stable conflicts, and user-scoped storage across `frontend/src/offline/` and `backend/internal/sync/`.
- **Optional infrastructure:** Redis is not authoritative, and optional S3/provider failure must not invalidate core auth/accounting truth. Adapter setup belongs under `backend/internal/platform/`.

## Anti-Patterns

### Business Logic in HTTP Handlers

**What happens:** A handler in `backend/internal/platform/httpapi/` calculates balances, validates ownership independently, or performs a multi-step mutation itself.
**Why it's wrong:** Direct API and offline sync paths can diverge, while atomicity and neighboring invariants become untestable at one shared boundary.
**Do this instead:** Add the rule to the owning domain repository/service, following `backend/internal/finance/repository.go`, and keep the handler as request/response translation.

### Independent Writes for One Logical Mutation

**What happens:** Domain state, sync feed, or replay receipt is written in separate SQL transactions.
**Why it's wrong:** Crashes and retries can produce committed money changes without a feed/receipt or double application.
**Do this instead:** Bind all participants with `backend/internal/platform/commandtx/transaction.go` and append through `backend/internal/platform/changefeed/change.go`, as orchestrated by `backend/internal/sync/service.go`.

### Treating Browser Storage as Server Truth

**What happens:** UI code assumes cached or optimistic IndexedDB entities are authoritative after conflicts, logout, or account switching.
**Why it's wrong:** It risks stale presentation, cross-account replay, or lost pending work.
**Do this instead:** Namespace storage by user, retain conflicts, and reconcile with changes/resync using `frontend/src/offline/db.ts`, `frontend/src/offline/syncApi.ts`, and `frontend/src/offline/conflicts.ts`.

### Expanding the Application Shell

**What happens:** New full-page workflows and data logic are embedded directly into the already large `frontend/src/app/App.tsx`.
**Why it's wrong:** Navigation, state orchestration, feature behavior, and presentation become tightly coupled and difficult to test independently.
**Do this instead:** Put page-level UI in `frontend/src/screens/`, reusable primitives in `frontend/src/app/components.tsx`, and transport/state logic in focused `frontend/src/app/*.ts` or `frontend/src/offline/` modules; leave `App.tsx` as the composition shell.

## Error Handling

**Strategy:** Validate at boundaries, return typed/domain errors, translate them once at HTTP/API edges, and preserve retry/conflict semantics in the offline client.

**Patterns:**
- Domain sentinel/typed errors live in files such as `backend/internal/finance/errors.go`, `backend/internal/planning/errors.go`, and `backend/internal/sync/errors.go`.
- HTTP error mapping and consistent JSON responses live in `backend/internal/platform/httpapi/errors.go`; correlation/observability middleware lives in `backend/internal/platform/httpapi/observability.go`.
- Frontend transport failures become `APIClientError` values in `frontend/src/app/apiClient.ts`; screens render explicit operation errors without discarding confirmed state.
- Offline retryable failures, stable conflicts, and quarantine states remain distinct in `frontend/src/offline/types.ts`.

## Cross-Cutting Concerns

**Logging:** Configure structured logging once through `backend/internal/platform/logging/logging.go`; propagate correlation IDs via `backend/internal/platform/httpapi/observability.go`; redact audit metadata through `backend/internal/audit/redact.go`.
**Validation:** Decode/shape checks occur in HTTP/TypeScript adapters, but ownership, money, category/wallet compatibility, versions, and atomicity belong in domain repositories under `backend/internal/`.
**Authentication:** Signed Google-backed browser identity and API-key bearer identity converge in `backend/internal/platform/httpapi/auth_context.go`; cookie mutations additionally pass CSRF protection from `backend/internal/identity/csrf.go`; Redis lookup failure falls back to PostgreSQL identity state.

---

*Architecture analysis: 2026-09-10*
