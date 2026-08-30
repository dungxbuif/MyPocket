# PHASE-003 Offline Synchronization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the temporary localStorage-first transaction outbox with a durable IndexedDB-primary offline mirror, idempotent backend sync APIs, explicit conflict review, and full-resync recovery.

**Architecture:** PostgreSQL remains authoritative. The React PWA stores a local IndexedDB mirror plus ordered mutation outbox and conflict records. The Go backend exposes authenticated `/api/v1/sync/*` routes backed by an idempotency ledger and per-user monotonic change log. Sync results update the local mirror; stale writes create explicit conflict records instead of overwriting server state silently.

**Tech Stack:** Go 1.24, PostgreSQL/pgx, React 19, TypeScript, Vite, Vitest, Playwright, native IndexedDB, Docker Compose.

**Spec:** `docs/work/phases/PHASE-003-detail-design.md`

## Global Constraints

- Use `backend/` for Go commands and `frontend/` for npm commands.
- Every shell command in docs and execution starts with `rtk`.
- Do not trust `user_id` from client sync payloads; derive ownership from auth context only.
- Process mutations in per-device sequence order; preserve independent mutations when another entity conflicts.
- Mutation IDs are idempotency keys; replaying a matching request must not repeat accounting effects.
- Base-version mismatches produce explicit conflict results; no silent last-write-wins.
- Tombstones remain in the change feed through the retained cursor horizon.
- IndexedDB is primary for offline mode; localStorage is only a one-time migration source for the temporary PHASE-002 outbox.
- Browser storage contains private finance data and must be cleared on logout/account deletion paths when those paths exist.
- Do not implement provider ingestion, planning automation, analytics, or production backup in PHASE-003.

---

### Task 1: Sync Schema And Backend Foundation

**Files:**
- Create: `backend/migrations/0003_phase003_sync.sql`
- Create: `backend/internal/sync/types.go`
- Create: `backend/internal/sync/service.go`
- Create: `backend/internal/sync/service_test.go`
- Modify: `backend/internal/platform/db/migrate_test.go`
- Modify during reconciliation: `docs/architecture/ERD.md`

**Interfaces:**
- Produces `sync_changes`, `sync_mutations`, and conflict/change DTOs for HTTP handlers.
- Consumes PHASE-002 finance entity IDs, versions, and archive/tombstone state.

- [ ] **Step 1: Write RED migration and service tests**

Cover expected tables, unique `(user_id, mutation_id)`, ordered cursor indexes, request-hash replay behavior, and stale base-version conflict classification.

Run: `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/platform/db ./internal/sync -count=1`

Expected: FAIL because sync schema/package does not exist.

- [ ] **Step 2: Add migration and sync primitives**

Create `sync_changes` with per-user cursor, entity type/id, operation, version, payload JSON, and tombstone support. Create `sync_mutations` with mutation UUID, request hash, result JSON, status, and timestamps.

- [ ] **Step 3: Implement service-level validation and replay ledger**

Validate UUIDs, supported entity/operation combinations, max batch size, payload bounds, base version, and retry classification. Store matching mutation replay results without reapplying domain effects.

- [ ] **Step 4: Run backend foundation tests**

Run: `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/platform/db ./internal/sync -count=1`

Expected: PASS.

---

### Task 2: IndexedDB Mirror And Local Migration

**Files:**
- Create: `frontend/src/offline/db.ts`
- Create: `frontend/src/offline/types.ts`
- Create: `frontend/src/offline/migrations.ts`
- Create: `frontend/src/offline/outbox.ts`
- Create: `frontend/src/offline/reducers.ts`
- Create: `frontend/src/offline/*.test.ts`
- Modify: `frontend/src/app/finance.ts`
- Modify: `frontend/src/app/App.tsx`
- Modify during reconciliation: `docs/architecture/ARCHITECTURE.md`

**Interfaces:**
- Produces IndexedDB stores: wallets, categories, transactions, outbox, conflicts, tombstones, meta.
- Migrates temporary PHASE-002 localStorage outbox once.

- [ ] **Step 1: Write RED frontend tests**

Cover schema open, migration from localStorage, optimistic mutation append, offline reload hydration, operation ordering, and degraded read-only mode.

Run: `rtk npm test -- --run src/offline`

Expected: FAIL because offline module does not exist.

- [ ] **Step 2: Implement IndexedDB adapter and schema migration**

Use native IndexedDB through a small typed adapter. Include versioned schema creation, transaction helpers, mirror upsert/delete/tombstone helpers, outbox state changes, and meta cursor storage.

- [ ] **Step 3: Wire finance state hydration to IndexedDB**

Hydrate cached records before network refresh, keep existing online fetch behavior for web use, and show pending/stale/offline states in mobile finance screens.

- [ ] **Step 4: Run frontend unit/component tests**

Run: `rtk npm test -- --run`

Expected: PASS.

---

### Task 3: Sync HTTP Routes And Finance Command Dispatch

**Files:**
- Create: `backend/internal/platform/httpapi/sync.go`
- Create: `backend/internal/platform/httpapi/sync_test.go`
- Modify: `backend/internal/platform/httpapi/router.go`
- Modify: `backend/internal/finance` as needed for version-checked command hooks.
- Modify during reconciliation: `docs/architecture/API.md`

**Interfaces:**
- Exposes `POST /api/v1/sync/mutations`, `GET /api/v1/sync/changes`, and `POST /api/v1/sync/resync`.
- Applies wallet/category/transaction mutations through existing authenticated finance command boundaries.

- [ ] **Step 1: Write RED HTTP and integration tests**

Cover auth/CSRF, user isolation, ordered batch behavior, duplicate replay, duplicate hash mismatch, stale base-version conflict, cursor paging, tombstones, and resync.

Run: `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/platform/httpapi -run Sync -count=1`

Expected: FAIL because sync routes do not exist.

- [ ] **Step 2: Implement authenticated sync routes**

Map safe error envelopes, enforce CSRF on mutation/resync routes, parse cursor/limit bounds, and use auth context for ownership.

- [ ] **Step 3: Dispatch supported finance mutations**

Support wallet/category create/update/archive/default/activation and transaction create/update/archive operations. Record sync changes after successful authoritative writes.

- [ ] **Step 4: Run backend regression**

Run: `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test -p 1 ./... -count=1`

Expected: PASS.

---

### Task 4: Reconnect Processor And Change Pull

**Files:**
- Modify: `frontend/src/offline/outbox.ts`
- Create: `frontend/src/offline/syncClient.ts`
- Create: `frontend/src/offline/processor.ts`
- Modify: `frontend/src/app/App.tsx`
- Modify: `frontend/src/app/App.test.tsx`
- Create/modify: `frontend/e2e/offline-sync.spec.ts`

**Interfaces:**
- Sends ordered mutation batches to backend sync API.
- Pulls changes after local cursor and updates IndexedDB mirror.

- [ ] **Step 1: Write RED processor and E2E tests**

Cover offline create/edit, reload offline, reconnect once, duplicate drain prevention, partial retry, and server change pull.

Run: `rtk npm test -- --run src/offline && rtk npm run test:e2e -- offline-sync.spec.ts`

Expected: FAIL because reconnect processor is not implemented.

- [ ] **Step 2: Implement online/offline detection and drain loop**

Process one ordered batch at a time, mark retryable failures, preserve pending mutations, and avoid concurrent drains.

- [ ] **Step 3: Apply sync results and pull changes**

Convert applied/replayed results into mirror updates, store conflicts locally, advance cursor only after successful change application, and surface stale state.

- [ ] **Step 4: Run frontend proof**

Run: `rtk npm test -- --run`

Run: `rtk npm run build`

Run: `rtk npm run test:e2e -- offline-sync.spec.ts`

Expected: PASS.

---

### Task 5: Conflict Inbox And Recovery

**Files:**
- Create: `frontend/src/offline/conflicts.ts`
- Modify: `frontend/src/app/App.tsx`
- Modify: `frontend/src/styles.css`
- Modify: `frontend/src/app/App.test.tsx`
- Modify: `frontend/e2e/offline-sync.spec.ts`

**Interfaces:**
- Renders mobile conflict inbox with keep-server, edit-and-retry, discard-local, and full-resync recovery.
- Consumes conflict results from Task 4 and sync result payloads from Task 3.

- [ ] **Step 1: Write RED conflict tests**

Cover conflict storage, visible inbox row, keep-server, edit-and-retry, discard-local, full resync, and two-context stale edit E2E.

Run: `rtk npm test -- --run && rtk npm run test:e2e -- offline-sync.spec.ts`

Expected: FAIL because conflict UI/recovery does not exist.

- [ ] **Step 2: Implement mobile conflict inbox**

Use existing mobile sheet patterns and concise Vietnamese copy. Do not block normal browsing of unrelated records.

- [ ] **Step 3: Implement recovery actions**

Keep-server mirrors authoritative state and removes local intent. Edit-and-retry creates a new mutation against current server version. Discard-local removes local intent. Full resync refreshes authoritative records while preserving recoverable unsent mutations.

- [ ] **Step 4: Run frontend and backend regression**

Run: `rtk npm test -- --run`

Run: `rtk npm run build`

Run: `rtk npm run test:e2e`

Run: `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test -p 1 ./... -count=1`

Expected: PASS.

---

### Task 6: Reconciliation, UAT, And Commit

**Files:**
- Modify: `docs/work/test-verification/PHASE-003-offline-sync.md`
- Modify: `docs/work/tickets/TICKET-008-indexeddb-mirror-outbox.md`
- Modify: `docs/work/tickets/TICKET-009-sync-api-change-feed.md`
- Modify: `docs/work/tickets/TICKET-010-conflict-inbox-recovery.md`
- Modify: `docs/work/phases/PHASE-003-offline-sync.md`
- Modify: `docs/work/VALIDATION_MATRIX.md`
- Modify: `docs/work/BACKLOG.md`
- Modify: `docs/CONTEXT.md`
- Modify: `docs/releases/CHANGELOG.md`
- Modify as needed: `docs/architecture/API.md`, `docs/architecture/ERD.md`, `docs/architecture/ARCHITECTURE.md`

**Interfaces:**
- Moves PHASE-003 and tickets to `in_review` only after proof is recorded.
- Preserves UAT pending until human sign-off.

- [ ] **Step 1: Record proof**

Update verification artifact with real commands, pass/fail history, attempt log, and evidence notes.

- [ ] **Step 2: Reconcile master docs**

Document sync APIs, database schema, offline module boundary, and recovery policy. Create an ADR only if implementation deviates from approved design.

- [ ] **Step 3: Update status and validation matrix**

Move TICKET-008..010 and PHASE-003 to `in_review` after automated proof passes. Keep UAT pending until user accepts behavior.

- [ ] **Step 4: Final verification**

Run: `rtk git diff --check`

Run: `rtk npm test -- --run`

Run: `rtk npm run build`

Run: `rtk npm run test:e2e`

Run: `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test -p 1 ./... -count=1`

Expected: PASS.

- [ ] **Step 5: Commit**

Run:

```bash
rtk git add backend frontend docs
rtk git commit -m "feat: add offline synchronization"
```
