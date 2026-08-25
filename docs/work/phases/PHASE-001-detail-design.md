---
artifact_type: detail_design
id: PHASE-001-DESIGN
status: ready
owner: ai
approval: approved
human_fields:
  - approval
  - constraints
  - scope_decisions
ai_fields:
  - problem
  - context_loaded
  - brownfield_scope
  - proposed_approach
  - design_tradeoffs
  - architecture_overview
  - execution_flow
  - api_data_model
  - security
  - test_plan
  - reconciliation_plan
shared_fields:
  - status
  - trace
  - small_task_exemption
trace:
  backlog_item: BL-001
  requirement:
    - REQ-F-001
    - REQ-NF-001
    - REQ-NF-004
    - REQ-NF-006
    - REQ-NF-008
  phase: PHASE-001
  ticket_or_bug:
    - TICKET-001
    - TICKET-002
    - TICKET-003
    - TICKET-004
  implementation_plan: ../../superpowers/plans/2026-08-25-phase-001-platform-identity.md
  test_verification: not_created_phase_not_executed
  validation_matrix: ../VALIDATION_MATRIX.md
  docs_review: per-ticket_completion_checklist
  adrs:
    - ADR-001
    - ADR-002
  master_docs_touched:
    - ../../architecture/API.md
    - ../../architecture/ERD.md
    - ../../architecture/ARCHITECTURE.md
---

# DETAIL DESIGN: PHASE-001 Platform and Identity

## Field Ownership

- Human owns approval, constraints, and scope decisions.
- AI owns analysis, architecture, flows, models, test plan, and reconciliation plan.
- Shared fields include status, trace links, and small-task exemption.

## Status

- ID: PHASE-001-DESIGN
- Status: ready
- Ticket/Bug: TICKET-001, TICKET-002, TICKET-003, TICKET-004
- Approval: approved by user instruction "finish whole app" on 2026-08-25
- Author: AI
- Updated: 2026-08-25

## Trace Links

- Backlog item: [BL-001](../BACKLOG.md)
- Requirement: [REQ-F-001](../../requirements/REQUIREMENTS.md), [REQ-NF-001](../../requirements/REQUIREMENTS.md), [REQ-NF-004](../../requirements/REQUIREMENTS.md), [REQ-NF-006](../../requirements/REQUIREMENTS.md), [REQ-NF-008](../../requirements/REQUIREMENTS.md)
- Phase: [PHASE-001](PHASE-001-platform-identity.md)
- Ticket/Bug: [TICKET-001](../tickets/TICKET-001-repository-runtime-foundation.md), [TICKET-002](../tickets/TICKET-002-postgresql-migrations-s3-platform-adapters.md), [TICKET-003](../tickets/TICKET-003-google-oauth-user-isolation.md), [TICKET-004](../tickets/TICKET-004-development-operations-verification-baseline.md)
- Implementation plan: [2026-08-25-phase-001-platform-identity.md](../../superpowers/plans/2026-08-25-phase-001-platform-identity.md)
- Test verification: created during execution
- Validation matrix: [VALIDATION_MATRIX.md](../VALIDATION_MATRIX.md)
- Docs review: per-ticket completion checklist
- ADRs: [ADR-001](../../decisions/ADR-001-react-go-modular-monolith.md), [ADR-002](../../decisions/ADR-002-stateless-google-oauth.md)
- Master docs touched during reconciliation: API, ERD, ARCHITECTURE if implementation adds durable specifics.

---

## 1. Context & Scope

### Problem Statement

- Problem statement: Build the first deployable MyPocket slice: an installable mobile-first React PWA shell, Go API, Go worker, PostgreSQL, S3-compatible storage, Google OAuth, signed stateless cookie auth, CSRF protection, correlation IDs, and ownership-isolation proof.
- Why now: The backlog marks BL-001 as the urgent first dependency for every finance, sync, analytics, ingestion, and operations phase.
- Success outcome: A developer can run the web/API/worker stack locally, authenticate through a deterministic Google callback fixture, access an authenticated PWA shell, reload the shell offline after first load, and see cross-user access denied in automated tests.

### Context Loaded

- `docs/CONTEXT.md`
- `docs/work/BACKLOG.md`
- `docs/standards/README.md`
- `docs/standards/WORKFLOW.md`
- `docs/standards/QUALITY_BAR.md`
- `docs/standards/VALIDATION.md`
- `docs/standards/DOCS.md`
- `docs/standards/GIT.md`
- `docs/work/phases/PHASE-001-platform-identity.md`
- `docs/work/VALIDATION_MATRIX.md`
- `docs/work/TRACEABILITY.md`
- `docs/work/ROADMAP.md`
- `docs/requirements/REQUIREMENTS.md`
- `docs/architecture/ARCHITECTURE.md`
- `docs/architecture/API.md`
- `docs/architecture/ERD.md`
- `docs/architecture/SDD.md`
- `docs/decisions/ADR-001-react-go-modular-monolith.md`
- `docs/decisions/ADR-002-stateless-google-oauth.md`

### Brownfield Scope

- Touched modules/files: no product code exists yet; execution will create `apps/web`, `apps/api`, `apps/worker`, shared Go packages, migrations, platform config, compose/dev docs, and verification artifacts.
- Direct dependencies inspected: planning docs, templates, requirements, architecture masters, and accepted ADRs.
- Contracts affected: `/api/v1` baseline, auth routes, `GET /api/v1/me`, `POST /api/v1/auth/logout`, `GET /api/v1/health/live`, `GET /api/v1/health/ready`, error envelopes, cookie, CSRF header, migration contract, S3 adapter contract.
- Known unknowns: production URLs, Google OAuth credentials, S3 credentials, and final homelab connection strings remain deployment inputs. Tests must use fixtures or local services instead of requiring production secrets.
- Scope expansion reason: user asked to focus on PWA/mobile UI and offline mechanism. This design includes installable PWA shell caching and offline app-shell reload in PHASE-001, while leaving full offline read/write, outbox replay, tombstones, and conflict resolution in PHASE-003.

### Small Task Exemption

- Small task exemption: no
- Reason: PHASE-001 changes architecture boundaries, auth/security, public API, data/migrations, runtime configuration, and user-visible PWA behavior.
- Impact checked: API=yes, DB=yes, Security=yes, Runtime=yes, Standards=no

---

## 2. Design Considerations & Trade-offs

| Consideration / Alternative | Pros | Cons | Decision |
| --- | --- | --- | --- |
| Include PWA app-shell caching in PHASE-001 | Matches user priority for PWA/mobile and provides early offline proof | Does not provide financial offline writes yet | Chosen |
| Move full IndexedDB outbox/sync into PHASE-001 | Satisfies offline end state earlier | Requires wallet/transaction domain contracts that PHASE-001 explicitly excludes | Rejected; keep full offline read/write in PHASE-003 |
| Use deterministic OAuth fixture for automated tests | Proves app user provisioning without real Google credentials | Real-provider callback still needs later environment validation | Chosen |
| Database-backed sessions | Enables server-side revocation | Violates ADR-002 and user-approved no-session scope | Rejected |
| One Go process for API and worker | Simpler startup | Conflicts with ADR-001 independent worker lifecycle | Rejected |
| Add finance placeholder entities for isolation tests | Gives realistic ownership proof early | Risks leaking PHASE-002 data model into PHASE-001 | Rejected; use a minimal test-owned fixture table or harness-only repository |

---

## 3. Architecture Overview

```text
-----------------------------+       /api/v1 JSON       +-----------------------------+
| apps/web React PWA         | -----------------------> | apps/api Go HTTP           |
| - mobile shell/nav         | <----------------------- | - auth middleware          |
| - service worker cache     |    signed cookie + CSRF  | - errors/correlation IDs   |
| - auth state + offline UI  |                          | - health/readiness         |
+--------------+--------------+                          +--------------+--------------+
               |                                                            |
               | local app-shell cache                                      | shared Go packages
               v                                                            v
       Browser Cache/IDB stub                                   internal/platform + identity
                                                                            |
                                                                            v
                                                     +----------------------+----------------+
                                                     | PostgreSQL users + platform metadata |
                                                     | S3-compatible smoke adapter          |
                                                     +----------------------+----------------+
                                                                            |
                                                                            v
                                                            apps/worker Go runtime
```

### Component Responsibilities

| Component | Role |
| --- | --- |
| `apps/web` | React TypeScript PWA shell, mobile layout, route fallback, auth state, offline status, service worker registration, API client baseline |
| `apps/api` | Go HTTP entrypoint, routes, middleware, config, health, auth endpoints, current-user endpoint |
| `apps/worker` | Go worker entrypoint that validates config, connects to platform dependencies, and exposes/logs a runnable baseline |
| `internal/platform/config` | Environment loading and safe validation errors |
| `internal/platform/http` | Router, JSON errors, correlation IDs, request logging, CSRF middleware |
| `internal/platform/db` | PostgreSQL connection and migration runner |
| `internal/platform/objectstore` | S3-compatible adapter and smoke operation |
| `internal/identity` | Google callback fixture/provider boundary, user provisioning, signed cookie claims, CSRF token handling, ownership context |
| `migrations` | Ordered immutable SQL migrations for PHASE-001 identity/platform tables |
| `deployments` or `compose.yaml` | Local PostgreSQL, S3-compatible service, API, worker, and web topology |

---

## 4. Execution Flow

1. Bootstrap repository tooling for Go workspace, React/Vite, test commands, lint/build commands, and local environment examples.
2. Implement API middleware first: correlation IDs, JSON error envelope, config loading, and health routes.
3. Add migration runner and PHASE-001 database tables for users and platform metadata.
4. Add S3-compatible adapter interface and local smoke proof.
5. Implement identity service with deterministic OAuth fixture support, user provisioning, signed cookie, logout, CSRF, and current-user endpoint.
6. Build mobile-first PWA shell with service worker app-shell caching, offline status, authenticated/unauthenticated states, and no finance behavior.
7. Add worker baseline that shares config/db/platform packages and proves independent process startup.
8. Add automated tests and platform smoke scripts, then record real evidence in tickets and validation artifacts during execution.
9. Reconcile API/ERD/architecture docs only where concrete implementation details are durable.

---

## 5. API & Data Model Design

### API Changes

- Endpoint: `GET /api/v1/health/live`
- Request: no body
- Response: `{"status":"ok","correlation_id":"<id>"}`

- Endpoint: `GET /api/v1/health/ready`
- Request: no body
- Response: `{"status":"ok","checks":{"database":"ok"},"correlation_id":"<id>"}` or safe non-200 readiness envelope.

- Endpoint: `GET /api/v1/auth/google`
- Request: browser navigation
- Response: redirect to Google authorization URL in production mode; deterministic fixture redirect in test mode.

- Endpoint: `GET /api/v1/auth/google/callback`
- Request: OAuth `code` and `state`
- Response: sets signed application cookie and redirects to the PWA shell.

- Endpoint: `POST /api/v1/auth/logout`
- Request: CSRF-protected cookie-authenticated request with no body
- Response: clears application cookie and returns `{"ok":true,"correlation_id":"<id>"}`.

- Endpoint: `GET /api/v1/me`
- Request: signed application cookie
- Response: `{"user":{"id":"<uuid>","email":"<verified-email>","display_name":"<name>"},"correlation_id":"<id>"}`.

- Error envelope: `{"error":{"code":"AUTH_REQUIRED|FORBIDDEN|VALIDATION_FAILED|INTERNAL_RETRYABLE|INTERNAL_FAILURE","message":"<safe message>"},"correlation_id":"<id>"}`.

### Data Model Changes

- Table: `users`
- Fields: `id uuid primary key`, `google_subject text unique not null`, `email text not null`, `email_verified boolean not null`, `display_name text not null default ''`, `avatar_url text not null default ''`, `created_at timestamptz not null`, `updated_at timestamptz not null`
- Notes: no session table, no Google provider token columns.

- Table: `schema_migrations`
- Fields: `version bigint primary key`, `name text not null`, `checksum text not null`, `applied_at timestamptz not null`
- Notes: migration runner may use an existing library if it creates equivalent metadata.

- Table: `ownership_test_objects`
- Fields: `id uuid primary key`, `user_id uuid not null references users(id)`, `name text not null`
- Notes: used only to prove ownership scoping until finance tables exist; it must be clearly marked as PHASE-001 test/harness infrastructure or omitted if integration tests can prove the same contract through a repository-level fake.

- Schema migration needed: yes.

---

## 6. Security & Authorization

- Authentication changes: introduce Google OAuth callback and signed stateless application cookie.
- Authorization / Permissions: all authenticated user context derives from the verified cookie; user-owned access must scope by `user_id`; request payloads must never assign ownership.
- Data Privacy / PII impact: store verified email, display name, and avatar URL only; do not store provider tokens; logs and errors must not include secrets, raw provider responses, or stack traces.
- Input Validation: validate OAuth state/nonce, callback fixture mode, cookie signature/expiry, CSRF token/header on cookie-authenticated mutations, environment config, request method/content type, and UUID route parameters.

---

## 7. Implementation & Verification Plan

### Impacted Areas

- Code/modules: new web app, API, worker, shared platform/identity packages, migrations, tests, local runtime files.
- Product behavior: installable mobile PWA shell, offline app-shell reload, login/logout/current-user states.
- API/contracts: auth, current user, health, errors, CSRF.
- Data/schema: PHASE-001 users and migration metadata.
- Security/auth: Google OAuth, signed cookie, CSRF, no persisted provider tokens, ownership isolation harness.
- Deployment/runtime: local compose/dev commands, readiness/liveness, worker process.
- Docs: tickets, test verification, validation matrix, context, backlog, release notes, API/ERD/architecture if implementation adds concrete details.

### Bug Analysis

Not applicable; this is feature/platform planning.

### Test Plan

- Unit: Go config, error envelope, correlation ID, cookie signing/verification, CSRF helpers, OAuth fixture validation; React shell/offline/auth-state components.
- Integration: migrations from empty PostgreSQL, user provisioning, no provider-token persistence, `/me`, logout, CSRF rejection, cross-user isolation harness, S3-compatible adapter smoke.
- E2E: mobile viewport PWA shell, unauthenticated login route, deterministic callback, authenticated shell, refresh persistence, logout, forbidden state, offline reload after first online load.
- UAT: install/mobile shell, login/logout, refresh persistence, forbidden state, offline app-shell reload.
- Manual/platform: documented commands for local web/API/worker/PostgreSQL/S3 startup, health readiness, and object-store smoke.
- Docs review: reconcile API/ERD/architecture/SDD/ADR need, update context/backlog/validation/release notes with evidence after execution.

### Validation Matrix Impact

- Update required: yes after implementation proof exists.
- Row(s): REQ-F-001, REQ-NF-001, REQ-NF-004, REQ-NF-006, REQ-NF-007, REQ-NF-008.
- Reason if no update required: not applicable.

### Verification Results

- Command: not run
- Result: not_started
- Notes: This is a draft design; implementation and verification are approval-gated.

---

## 8. Reconciliation Plan

- Requirements docs: no change expected unless scope changes.
- Architecture docs: update if concrete package names, runtime files, or health topology differ from the current architecture.
- API docs: update with finalized PHASE-001 endpoint schemas, CSRF header name, cookie notes, and health response shape.
- ERD docs: update with concrete PHASE-001 table definitions after migrations are implemented.
- SDD docs: no change expected unless the implementation diverges from approved design.
- ADR: no change expected; create/update only if auth/runtime/offline app-shell policy diverges from ADR-001 or ADR-002.
- Context: update after execution to reflect current status and next steps.

### Docs Review Checklist

- Code changed but docs unchanged reason: not yet executed.
- Requirements updated or not needed reason: expected no change; confirm after execution.
- Architecture updated or not needed reason: confirm after execution.
- API updated or not needed reason: expected update for concrete endpoint schemas.
- ERD/data updated or not needed reason: expected update for concrete PHASE-001 tables.
- SDD updated or not needed reason: expected no change unless design diverges.
- ADR created or not needed reason: expected no new ADR unless durable decision changes.
