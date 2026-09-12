# Codebase Structure

**Analysis Date:** 2026-09-10

## Directory Layout

```text
MyPocket/
├── backend/
│   ├── cmd/                    # API, worker, and migration process entry points
│   ├── internal/               # Go domain modules and platform adapters
│   │   ├── finance/            # Wallet/category/transaction accounting
│   │   ├── planning/           # Budgets, events, obligations, schedules, drafts
│   │   ├── portfolio/          # Asset positions, trades, prices, valuation
│   │   ├── sync/               # Offline mutation replay and change feed access
│   │   ├── analytics/          # Dashboard and reporting queries
│   │   ├── identity/           # Users, cookies, CSRF, API keys, ownership
│   │   ├── notification/       # Inbox, push subscriptions, delivery processing
│   │   ├── audit/              # Safe append-only operational audit records
│   │   ├── worker/             # Background runner orchestration
│   │   └── platform/           # HTTP, DB, config, logging, cache, object adapters
│   └── migrations/             # Ordered PostgreSQL schema migrations
├── frontend/
│   ├── src/
│   │   ├── app/                # App shell, shared UI, API feature clients
│   │   ├── screens/            # Page-level feature screens
│   │   ├── offline/            # IndexedDB mirror, outbox, sync, conflicts
│   │   ├── lib/                # Small generic frontend utilities
│   │   └── test/               # Shared frontend test setup/component tests
│   ├── e2e/                    # Playwright user-flow tests
│   └── public/                 # Service worker, manifest, PWA assets
├── docs/
│   ├── architecture/           # Stable architecture/API/ERD/integration masters
│   ├── decisions/              # Durable ADRs
│   ├── engineering/            # Local setup and troubleshooting
│   ├── operations/             # Deployment, environments, release checks
│   ├── requirements/           # Product requirements and user stories
│   ├── standards/              # Repository workflow/quality/testing rules
│   └── work/                   # Active phases, tickets, evidence, backlog
├── design/                     # Approved UI system and component specifications
├── scripts/                    # Smoke, E2E, and beta verification scripts
├── compose.yaml                # Local/homelab service composition
├── AGENTS.md                   # Mandatory repository agent gateway
└── README.md                   # Project entry documentation
```

## Directory Purposes

**`backend/cmd/`:**
- Purpose: Define executable composition roots only.
- Contains: `main` packages for HTTP API, worker, and migration runner.
- Key files: `backend/cmd/api/main.go`, `backend/cmd/worker/main.go`, `backend/cmd/migrate/main.go`.

**`backend/internal/`:**
- Purpose: Hold all non-public Go application code, separated by business domain and platform concern.
- Contains: Domain entities, validation, SQL repositories, service orchestration, HTTP adapters, worker runners.
- Key files: `backend/internal/platform/httpapi/router.go`, `backend/internal/finance/repository.go`, `backend/internal/sync/service.go`.

**`backend/internal/platform/`:**
- Purpose: Isolate technical infrastructure and cross-domain mechanisms from business modules.
- Contains: `authcache/`, `changefeed/`, `commandtx/`, `config/`, `db/`, `httpapi/`, `logging/`, `objectstore/`.
- Key files: `backend/internal/platform/commandtx/transaction.go`, `backend/internal/platform/changefeed/change.go`, `backend/internal/platform/config/config.go`.

**`backend/migrations/`:**
- Purpose: Define the authoritative PostgreSQL schema as forward-only ordered SQL.
- Contains: Phase-numbered migrations from identity through atomic sync asset feed.
- Key files: `backend/migrations/0001_phase001_identity.sql`, `backend/migrations/0012_atomic_sync_asset_feed.sql`.

**`frontend/src/app/`:**
- Purpose: Compose the browser application and expose feature-oriented API clients and shared UI.
- Contains: `App.tsx`, auth/API clients, finance/planning/portfolio/analytics modules, shared components, caches, service-worker registration.
- Key files: `frontend/src/app/App.tsx`, `frontend/src/app/apiClient.ts`, `frontend/src/app/components.tsx`.

**`frontend/src/screens/`:**
- Purpose: Keep route/tab-level presentation outside the application shell.
- Contains: Overview, transactions, budgets, reports, and account screens with colocated component tests.
- Key files: `frontend/src/screens/OverviewScreen.tsx`, `frontend/src/screens/TransactionsScreen.tsx`, `frontend/src/screens/AccountScreen.tsx`.

**`frontend/src/offline/`:**
- Purpose: Own durable browser persistence and synchronization semantics.
- Contains: IndexedDB access/schema, migrations, mutation builders, reducers, sync API, conflicts, receipt queue, offline types.
- Key files: `frontend/src/offline/db.ts`, `frontend/src/offline/outbox.ts`, `frontend/src/offline/syncApi.ts`, `frontend/src/offline/types.ts`.

**`frontend/e2e/`:**
- Purpose: Verify complete browser workflows across desktop/mobile/WebKit configurations.
- Contains: Playwright specs for auth, finance, offline sync, accounting, planning, PWA, feedback, receipts, and isolation.
- Key files: `frontend/e2e/business-core.spec.ts`, `frontend/e2e/accounting-correctness.spec.ts`, `frontend/e2e/offline-sync.spec.ts`.

**`docs/`:**
- Purpose: Store durable product/technical truth plus execution trace and evidence.
- Contains: Master architecture, requirements, ADRs, standards, work queue, tickets, phase designs, validation evidence, operations and engineering guides.
- Key files: `docs/CONTEXT.md`, `docs/architecture/ARCHITECTURE.md`, `docs/work/BACKLOG.md`, `docs/work/VALIDATION_MATRIX.md`.

**`design/`:**
- Purpose: Define the shared frontend visual/component contract.
- Contains: Product design direction, base component rules, component/file mapping.
- Key files: `design/DESIGN.md`, `design/BASE_COMPONENTS.md`, `design/COMPONENT_FILES_SPEC.md`.

## Key File Locations

**Entry Points:**
- `frontend/src/main.tsx`: Browser bootstrap and provider wiring.
- `backend/cmd/api/main.go`: API dependency composition and HTTP listener.
- `backend/cmd/worker/main.go`: Background processor loop.
- `backend/cmd/migrate/main.go`: Schema migration executable.
- `frontend/public/sw.js`: PWA service worker runtime.

**Configuration:**
- `backend/internal/platform/config/config.go`: Environment-driven backend configuration and validation.
- `frontend/vite.config.ts`: Frontend build, test, and PWA-related Vite configuration.
- `frontend/tsconfig.json`: TypeScript compiler settings.
- `frontend/playwright.config.ts`: Browser test configuration.
- `compose.yaml`: PostgreSQL, Redis, object storage, API, worker, and web service composition.
- `backend/Dockerfile`, `frontend/Dockerfile`: Production image definitions.

**Core Logic:**
- `backend/internal/finance/`: Wallet/category/transaction rules and persistence.
- `backend/internal/planning/`: Budget/event/obligation/recurrence/draft rules.
- `backend/internal/portfolio/`: Asset ledger, decimal/money checks, persistence, provider boundary.
- `backend/internal/sync/service.go`: Atomic offline replay and conflict orchestration.
- `backend/internal/analytics/repository.go`: Dashboard/report read models.
- `frontend/src/app/App.tsx`: Top-level authenticated workflow composition.
- `frontend/src/offline/`: Client mirror, outbox, conflict, and resync behavior.

**Testing:**
- `backend/internal/**/*_test.go`: Colocated Go unit and PostgreSQL integration tests.
- `frontend/src/**/*.test.ts`, `frontend/src/**/*.test.tsx`: Colocated Vitest unit/component tests.
- `frontend/e2e/*.spec.ts`: Playwright end-to-end tests.
- `scripts/verify-beta.sh`: Repository release-grade verification ladder.

**Documentation:**
- `docs/architecture/`: Current system, API, ERD, integration, and SDD truth.
- `docs/decisions/`: Architectural decisions including modular monolith, offline identity, and atomic sync.
- `docs/work/phases/`, `docs/work/tickets/`: Approved execution scope and detailed design.
- `docs/work/test-verification/`: Command evidence and remaining risks.

## Naming Conventions

**Files:**
- Go uses lowercase concern names, splitting large domains by behavior: `backend/internal/finance/transactions.go`, `backend/internal/planning/events_obligations.go`.
- Go tests are colocated with `_test.go`: `backend/internal/sync/atomic_test.go`.
- React components/screens use PascalCase: `frontend/src/screens/BudgetsScreen.tsx`, `frontend/src/app/SearchPanel.tsx`.
- TypeScript service/helper modules use camelCase or concise lowercase feature names: `frontend/src/app/apiClient.ts`, `frontend/src/app/reportQueries.ts`, `frontend/src/offline/syncApi.ts`.
- Frontend tests are colocated as `.test.ts(x)`; browser flows live separately as `.spec.ts` under `frontend/e2e/`.
- PostgreSQL migrations use four-digit ordering plus phase/purpose: `backend/migrations/0012_atomic_sync_asset_feed.sql`.
- Durable docs use uppercase stable names or dated descriptive work artifacts: `docs/architecture/ARCHITECTURE.md`, `docs/work/test-verification/BUSINESS-LOGIC-AUDIT-2026-09-10.md`.

**Directories:**
- Go domain/platform packages use singular lowercase names: `backend/internal/finance/`, `backend/internal/objectstore/` is not used; object storage correctly lives at `backend/internal/platform/objectstore/`.
- Frontend organization is role-based at the top level (`app`, `screens`, `offline`, `lib`, `test`) and feature-oriented inside `app`.
- Documentation separates stable truth (`docs/architecture/`, `docs/requirements/`, `docs/decisions/`) from active execution state (`docs/work/`).

## Where to Add New Code

**New Backend Feature:**
- Primary code: Add business rules, types, and persistence to the owning package under `backend/internal/<domain>/`; create a new domain directory only for a genuinely separate ownership boundary.
- HTTP adapter: Add or extend a focused file under `backend/internal/platform/httpapi/`, declare the narrow consumed interface in that package, and wire the concrete repository in `backend/cmd/api/main.go`.
- Tests: Colocate unit/integration tests as `backend/internal/<domain>/*_test.go`; add HTTP contract proof under `backend/internal/platform/httpapi/*_test.go`.
- Data changes: Add the next ordered SQL file under `backend/migrations/` and reconcile `docs/architecture/ERD.md`.

**New Frontend Feature:**
- Primary screen: Add a PascalCase component under `frontend/src/screens/` when it represents a page/tab-level workflow.
- API contract/client: Add a focused feature module under `frontend/src/app/` and route all requests through `frontend/src/app/apiClient.ts`.
- Shell integration: Keep routing/navigation/composition changes small in `frontend/src/app/App.tsx`.
- Tests: Colocate `.test.tsx` with the screen/module and add cross-boundary journeys to `frontend/e2e/`.

**New Offline-Capable Mutation:**
- Shared types/state: Extend `frontend/src/offline/types.ts`.
- Optimistic command: Add to `frontend/src/offline/outbox.ts` and persist via `frontend/src/offline/db.ts`.
- Reconciliation: Update `frontend/src/offline/reducers.ts`, `frontend/src/offline/conflicts.ts`, and `frontend/src/offline/syncApi.ts` as applicable.
- Server replay: Extend `backend/internal/sync/types.go` and `backend/internal/sync/service.go`, calling the owning domain repository inside the shared command transaction.
- Schema/feed: Update `backend/migrations/` only when the persisted contract changes; verify direct and replayed paths emit the same canonical change.

**New Background Job:**
- Runner implementation: Add a focused file under `backend/internal/worker/` when it orchestrates a domain repository; keep domain-specific delivery mechanics in the owning domain package.
- Process wiring: Register the bounded runner in `backend/cmd/worker/main.go`.
- Coordination: Use PostgreSQL leases through an owning repository; do not add process-local-only coordination for correctness-sensitive work.
- Tests: Add a colocated runner test under `backend/internal/worker/` and repository integration proof in the owning domain.

**New Component/Module:**
- Shared UI primitive: `frontend/src/app/components.tsx` when it is part of the established base component set; split into a focused peer module if growth would further enlarge this file.
- Page-specific component: Colocate with its screen in `frontend/src/screens/`.
- Domain pure helper: Place beside the owning Go domain, following `backend/internal/portfolio/ledger.go` and `backend/internal/finance/transactions.go`.
- Infrastructure adapter: Add beneath `backend/internal/platform/` and expose a small interface to consumers.

**Utilities:**
- Shared frontend helpers: `frontend/src/lib/` only for domain-neutral utilities; keep finance/planning behavior in feature modules.
- Shared backend technical helpers: `backend/internal/platform/`; keep business rules in the owning domain package.
- Avoid generic dumping grounds. Prefer the narrowest owning directory evidenced by existing call sites.

**Documentation for Technical Changes:**
- Architecture/API/data changes: Reconcile `docs/architecture/ARCHITECTURE.md`, `docs/architecture/API.md`, `docs/architecture/ERD.md`, or `docs/architecture/INTEGRATIONS.md` as applicable.
- Durable boundary decisions: Add an ADR under `docs/decisions/`.
- Approved implementation scope and evidence: Use `docs/work/phases/`, `docs/work/tickets/`, and `docs/work/test-verification/` according to `AGENTS.md`.
- User/developer operations: Update `README.md`, `docs/engineering/`, or `docs/operations/` and maintain an agent-readable Markdown surface.

## Special Directories

**`.planning/codebase/`:**
- Purpose: Generated GSD reference maps for future planning/execution agents.
- Generated: Yes.
- Committed: Project/workflow dependent; preserve as planning artifacts unless the orchestrator decides otherwise.

**`backend/migrations/`:**
- Purpose: Ordered, append-only PostgreSQL schema evolution.
- Generated: No.
- Committed: Yes.

**`frontend/public/`:**
- Purpose: Static PWA files copied into the frontend build.
- Generated: No.
- Committed: Yes.

**`frontend/dist/`:**
- Purpose: Vite production build output when present.
- Generated: Yes.
- Committed: No; do not place source changes here.

**`frontend/playwright-report/` and `frontend/test-results/`:**
- Purpose: Browser test reports and failure artifacts when present.
- Generated: Yes.
- Committed: No; use summarized evidence under `docs/work/test-verification/`.

**`docs/work/`:**
- Purpose: Active scheduler, approved designs, tickets, validation state, and verification records.
- Generated: No.
- Committed: Yes.

---

*Structure analysis: 2026-09-10*
