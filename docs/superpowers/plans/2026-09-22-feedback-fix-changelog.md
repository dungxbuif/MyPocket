# Feedback Fix Changelog Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax.

**Goal:** Ship an owner-scoped feedback workflow that a narrow local/dev agent can triage and finalize into a versioned changelog without exposing raw feedback publicly or touching financial APIs.

**Architecture:** Add additive PostgreSQL `feedback` and `changelog` tables with repository/use-case boundaries. User routes use the existing JWT middleware; agent/internal routes use a dedicated constant-time service-token middleware and a Redis audit sink. A single transaction creates a changelog and marks all referenced feedback fixed. The frontend adds an Account → Feedback route using shared form/status/card components.

**Tech Stack:** Go, Gin, GORM, PostgreSQL migrations, Redis Streams, generated Swagger, React/TypeScript, TanStack Router, Node test scripts, real Chromium CDP E2E.

**Spec:** `docs/superpowers/specs/2026-09-22-feedback-fix-changelog-design.md`

## Global Constraints

- No public raw-feedback endpoint; only owner JWT routes and the scoped agent service token may read descriptions.
- Agent responses omit `user_id`, email, and account identifiers; audit payloads omit descriptions, credentials, JWTs, and financial records.
- Lifecycle is `open → triaged → in_progress → fixed|rejected`; direct `fixed` status updates are rejected and changelog publication is the only finalization path.
- Changelog publication locks feedback IDs in sorted order, creates the unique version, updates all feedback rows, then emits one post-commit Redis audit event.
- Migration is additive; existing ledger, AI, JWT, and current API-key backlog behavior must remain compatible.
- Frontend screens use existing base components; no screen-local native controls or literal color tokens.
- Tests must run with `rtk` command prefixes and evidence must be recorded in the validation matrix.

## Review Focus

- Concurrent changelog calls for the same feedback/version must not double-fix rows or create duplicate versions — covered in Task 3 PostgreSQL concurrency tests.
- Cross-owner JWT reads and agent responses must not disclose `user_id` or another user's description — covered in Task 4 handler/integration tests.
- Invalid or missing service tokens must not reach internal handlers, and audit events must never contain token/description — covered in Task 2 middleware/audit tests.
- A partial feedback ID list must roll back both changelog and all feedback status updates — covered in Task 3 transaction test.
- Browser reload after a submitted feedback must show the persisted status and fixed version without AI-chat state — covered in Task 6 Chromium E2E.

---

### Task 1: Domain schema, entities, and repository contracts

**Files:**
- Create: `backend/migrations/000017_feedback_changelog.up.sql`
- Create: `backend/migrations/000017_feedback_changelog.down.sql`
- Create: `backend/internal/entity/feedback.go`
- Create: `backend/internal/entity/changelog.go`
- Create: `backend/internal/entity/feedback_test.go`
- Create: `backend/internal/repository/feedback.go`
- Create: `backend/internal/repository/changelog.go`
- Create: `backend/internal/repository/audit.go`
- Create: `backend/internal/infrastructure/repository/feedback_postgres.go`
- Create: `backend/internal/infrastructure/repository/changelog_postgres.go`
- Test: `backend/internal/infrastructure/repository/feedback_postgres_test.go`

**Interfaces:**
- Produces `entity.Feedback`, `entity.Changelog`, status/type constants, and repository interfaces for owner reads, agent polling, status transitions, and atomic finalization.
- Produces `repository.AuditSink` with `Record(context.Context, AuditEvent) error`; existing auth/cache interfaces remain unchanged.

- [ ] **Step 1: Write failing entity and migration tests.** Assert valid enum values, invalid transitions, title/description bounds, migration version 17, unique changelog version, owner/status indexes, and nullable `changelog_id`.
- [ ] **Step 2: Run the focused tests and verify RED.** Run `rtk go test ./internal/entity ./internal/infrastructure/repository -run 'Feedback|Changelog|Migration' -count=1` from `backend/`. Expected: missing types/tables or migration version failure.
- [ ] **Step 3: Add entity types and constants.** Define JSON-safe response structs, status/type validation helpers, and timestamps without exposing internal owner IDs in agent/public DTOs.
- [ ] **Step 4: Add migration 000017.** Create `changelogs` first, then `feedback` with UUID FKs, check constraints, default status, indexes, and reversible down SQL. Do not alter existing tables except the new FK dependency.
- [ ] **Step 5: Define repository interfaces and audit event shape.** Include owner list/get/create, agent list/status update, sorted-lock finalization, public changelog list/get, and the minimal audit fields from the spec.
- [ ] **Step 6: Implement PostgreSQL repository methods and run focused tests.** Use transactions and `FOR UPDATE` in deterministic ID order; verify rollback, owner scope, unique version, and exact timestamps.
- [ ] **Step 7: Run migration and repository verification.** Run `rtk env TEST_DATABASE_URL='postgres://dev:password@127.0.0.1:5432/postgres?sslmode=disable' go run ./cmd/migrate up`, the same command with `version`, and `rtk env TEST_DATABASE_URL='postgres://dev:password@127.0.0.1:5432/postgres?sslmode=disable' go test ./internal/infrastructure/repository -run 'Feedback|Changelog' -count=1`.
- [ ] **Step 8: Commit the schema slice.** Stage the migration, entity, repository-contract, PostgreSQL implementation, and test files listed above, then run `rtk git commit -m "feat: add feedback and changelog persistence"`.

### Task 2: Agent credential boundary and Redis audit sink

**Files:**
- Modify: `backend/internal/config/config.go`
- Modify: `backend/.env.example`
- Create: `backend/internal/controller/http/agent_middleware.go`
- Create: `backend/internal/controller/http/agent_middleware_test.go`
- Create: `backend/internal/infrastructure/cache/audit.go`
- Create: `backend/internal/infrastructure/cache/audit_test.go`
- Modify: `backend/internal/repository/audit.go`
- Modify: `backend/cmd/api/main.go`

**Interfaces:**
- Consumes `FEEDBACK_AGENT_TOKEN` and `repository.AuditSink` from Task 1.
- Produces `RequireFeedbackAgent`, a middleware that sets an internal actor marker and aborts invalid/missing credentials with `401`; produces Redis Stream key `mypocket:audit:feedback` with redacted fields only.

- [ ] **Step 1: Write failing middleware tests.** Cover missing header, wrong scheme, wrong token, valid token, constant-time comparison boundary, and no leakage of credential values in response/log fields.
- [ ] **Step 2: Verify RED.** Run `rtk go test ./internal/controller/http ./internal/infrastructure/cache -run 'Agent|Audit' -count=1`; expected missing middleware/config/audit method failures.
- [ ] **Step 3: Add config loading and startup validation.** Add `FeedbackAgentToken` to `config.Config`; default empty only for development with internal routes disabled; fail non-development startup when the routes are wired without a token.
- [ ] **Step 4: Implement middleware.** Trim `Authorization`, require `Bearer`, compare with `crypto/subtle.ConstantTimeCompare`, set actor kind `agent`, and return the existing problem envelope.
- [ ] **Step 5: Implement Redis Stream audit.** Add `Record` to `cache.Redis` using `XADD` with request ID/action/actor/status/version/IDs only; never serialize title/description/token. Make the sink best-effort after DB commit.
- [ ] **Step 6: Wire dependency construction without widening existing auth.** Pass the audit sink and token middleware to the router; leave JWT middleware unchanged.
- [ ] **Step 7: Run focused green tests and commit.** Run `rtk go test ./internal/controller/http ./internal/infrastructure/cache -run 'Agent|Audit' -count=1`; stage the config, middleware, Redis audit, router wiring, and test files listed above, then run `rtk git commit -m "feat: add scoped feedback agent auth and audit"`.

### Task 3: Feedback/changelog use cases and atomic workflow

**Files:**
- Create: `backend/internal/usecase/feedback.go`
- Create: `backend/internal/usecase/feedback_test.go`
- Create: `backend/internal/usecase/changelog.go`
- Create: `backend/internal/usecase/changelog_test.go`
- Modify: `backend/internal/repository/feedback.go`
- Modify: `backend/internal/repository/changelog.go`

**Interfaces:**
- Consumes repositories and audit sink from Tasks 1–2.
- Produces `FeedbackService.Create/List/Get`, `FeedbackService.AgentList/SetStatus`, and `ChangelogService.Publish/List/Get`; all methods accept an explicit actor/owner ID and context.

- [ ] **Step 1: Write failing use-case tests.** Cover create validation, owner list/get, transition matrix, direct-fixed rejection, rejected/fixed terminal conflicts, agent DTO redaction, duplicate version, partial-list rollback, and audit event contents.
- [ ] **Step 2: Verify RED.** Run `rtk go test ./internal/usecase -run 'Feedback|Changelog' -count=1`; expected missing service methods/errors.
- [ ] **Step 3: Implement validation and owner operations.** Trim and bound fields, generate UUIDs, default `open`, and map repository not-found/conflict errors to stable domain errors.
- [ ] **Step 4: Implement status transitions.** Permit only `triaged`, `in_progress`, or `rejected` through the status method; reject backward/terminal/direct-fixed changes.
- [ ] **Step 5: Implement publish orchestration.** Validate all IDs, call the repository’s sorted-lock transaction, and emit exactly one redacted `changelog.published` event after commit.
- [ ] **Step 6: Run green unit tests and commit.** Run `rtk go test ./internal/usecase -run 'Feedback|Changelog' -count=1`; commit `feat: add feedback lifecycle services`.

### Task 4: HTTP routes, Swagger, and API integration

**Files:**
- Create: `backend/internal/controller/http/feedback_handler.go`
- Create: `backend/internal/controller/http/changelog_handler.go`
- Create: `backend/internal/controller/http/routes_feedback.go`
- Create: `backend/internal/controller/http/feedback_handler_test.go`
- Create: `backend/internal/controller/http/changelog_handler_test.go`
- Modify: `backend/internal/controller/http/router.go`
- Modify: `backend/cmd/api/main.go`
- Modify: `backend/docs/docs.go`
- Modify: `backend/docs/swagger.json`
- Modify: `backend/docs/swagger.yaml`
- Create: `app/scripts/feedback-api-roundtrip.mjs`
- Modify: `app/package.json`

**Interfaces:**
- Consumes use-case services and `RequireFeedbackAgent` from Tasks 2–3.
- Produces the exact `/api/v1/feedback`, `/api/v1/agent/feedback`, `/api/v1/internal/feedback`, `/api/v1/internal/changelog`, and public `/api/v1/changelog` contracts in the spec.

- [ ] **Step 1: Write failing handler/route tests.** Assert route registration, problem envelopes, JWT owner scope, agent auth, public redaction, body bounds, list limits, status conflicts, duplicate version, and no raw feedback on public endpoints.
- [ ] **Step 2: Verify RED.** Run `rtk go test ./internal/controller/http -run 'Feedback|Changelog' -count=1`; expected route/handler failures.
- [ ] **Step 3: Implement handlers and route groups.** Reuse `OK`, `Created`, `NoContent`, `Fail`; register JWT user routes, agent/internal groups, and unauthenticated changelog reads without exposing owner fields.
- [ ] **Step 4: Generate Swagger and update API docs.** Run `rtk go generate ./cmd/api`; describe auth schemes, inputs, status codes, redaction and examples.
- [ ] **Step 5: Write the FE-proxy API roundtrip first as a failing integration script.** It logs in, creates feedback, lists/gets it, drives triaged/in-progress, publishes a changelog with the agent token, confirms fixed/version, verifies public changelog, cross-owner/unauthorized denial, and cleans only test rows through the API/database-safe test helper.
- [ ] **Step 6: Run green HTTP/integration tests.** Run `rtk env TEST_DATABASE_URL='postgres://dev:password@127.0.0.1:5432/postgres?sslmode=disable' FEEDBACK_AGENT_TOKEN='feedback-test-token' go test ./internal/controller/http -run 'Feedback|Changelog' -count=1` and `rtk env TEST_API_ORIGIN=http://127.0.0.1:4173 FEEDBACK_AGENT_TOKEN='feedback-test-token' TEST_DATABASE_URL='postgres://dev:password@127.0.0.1:5432/postgres?sslmode=disable' npm run test:feedback:api` from `app/`.
- [ ] **Step 7: Commit the backend API slice.** Commit generated Swagger, handlers, route wiring, script and package script as `feat: expose feedback and changelog APIs`.

### Task 5: Frontend service, route, and shared UI

**Files:**
- Create: `app/src/services/feedback.ts`
- Create: `app/src/atomic/organisms/FeedbackPanel.tsx`
- Create: `app/src/atomic/molecules/FeedbackStatus.tsx`
- Create: `app/scripts/feedback-ui.test.ts`
- Create: `app/scripts/feedback-browser.test.mjs`
- Modify: `app/src/router.tsx`
- Modify: `app/src/atomic/pages/FinancePrototypePage.tsx`
- Modify: `app/src/atomic/organisms/AccountPanel.tsx`
- Modify: `app/package.json`

**Interfaces:**
- Consumes user feedback/changelog endpoints from Task 4 and existing `apiRequest`, `getStoredToken`, base components, and Account route conventions.
- Produces an owner-only `/account/feedback` page with create form, list, loading/error/empty states, fixed version card, and public changelog summary.

- [ ] **Step 1: Write failing service/logic tests.** Cover payload trimming, allowed type/status labels, fixed version rendering, error mapping, and no service-token use in the browser.
- [ ] **Step 2: Verify RED.** Run `rtk node --test scripts/feedback-ui.test.ts` from `app/`; expected missing service/formatter failures.
- [ ] **Step 3: Implement service types/functions.** Add `Feedback`, `Changelog`, `FeedbackInput`, `fetchFeedback`, `createFeedback`, and `fetchChangelog`; use JWT storage only.
- [ ] **Step 4: Implement shared status molecule and panel.** Compose `BaseSelect`, `BaseTextInput`, `BaseButton`, `SurfaceCard`, `StatusMessage`, and existing loading/error patterns. Keep fixed copy exactly `Fixed in {version}` and Vietnamese status label.
- [ ] **Step 5: Register route/navigation.** Add `/account/feedback`, Account action entry, and route rendering; preserve all existing tabs and auth behavior.
- [ ] **Step 6: Run green frontend checks.** Run `rtk node --test scripts/feedback-ui.test.ts`, `rtk npm run check:design`, `rtk npm run test:design`, `rtk npm run typecheck`, and `rtk npm run build`.
- [ ] **Step 7: Commit the UI slice.** Commit as `feat: add user feedback status UI`.

### Task 6: Real browser E2E and full reconciliation

**Files:**
- Modify: `app/scripts/feedback-browser.test.mjs`
- Modify: `app/package.json`
- Modify: `docs/architecture/API.md`
- Modify: `docs/architecture/ERD.md`
- Modify: `docs/architecture/INTEGRATIONS.md`
- Modify: `docs/CONTEXT.md`
- Modify: `docs/work/BACKLOG.md`
- Modify: `docs/work/VALIDATION_MATRIX.md`
- Modify: `docs/releases/CHANGELOG.md`
- Create: `docs/architecture/FEEDBACK_API.md`
- Create: `docs/decisions/ADR-009-feedback-agent-audit.md`

**Interfaces:**
- Consumes all runtime behavior from Tasks 1–5.
- Produces repeatable browser proof, machine-readable API/agent guidance, security decision record, and release note.

- [ ] **Step 1: Write the failing Chromium flow.** Authenticate through the existing callback, navigate Account → Feedback, submit a bug, reload, assert persisted open state, and assert fixed/version rendering from a test-finalized record; assert no agent controls or other-owner raw data appear.
- [ ] **Step 2: Verify RED.** Run `rtk npm run test:feedback:e2e` from `app/`; expected missing route/UI/API failures.
- [ ] **Step 3: Implement deterministic setup/cleanup.** Use generated test-owned feedback/changelog rows and remove only those rows; never modify wallets or transactions.
- [ ] **Step 4: Run green E2E and full regression.** Run `rtk npm run test:feedback:e2e`, all existing frontend scripts, `rtk env TEST_DATABASE_URL='postgres://dev:password@127.0.0.1:5432/postgres?sslmode=disable' go test -race ./... -count=1`, migration version, `rtk git diff --check`, and docs-link checks.
- [ ] **Step 5: Reconcile docs and publish release note.** Record exact commands/results, residual live-provider/manual UAT, agent token configuration, API examples, audit redaction, and the `Fixed in v1.4.2` display behavior. Do not claim an external issue fixed without a real linked changelog transaction.
- [ ] **Step 6: Commit final docs/tests.** Commit only owned files as `docs: document feedback fix changelog workflow` after all checks pass.

## Plan self-review

- Spec coverage: data model (Task 1), auth/audit (Task 2), lifecycle/atomicity (Task 3), API (Task 4), UI (Task 5), verification/privacy/docs (Task 6).
- No public raw feedback: enforced in Tasks 2, 4, 5, and 6.
- No direct fixed status: enforced in Tasks 3 and 4.
- No unfinished placeholders or unassigned work: every step has an owner file, test command, expected result, or commit boundary.
- Existing dirty worktree: each commit stages only the files listed in its task; unrelated prior changes remain untouched.
