# Offline Financial Write/Sync Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a safe, owner-scoped iOS offline outbox for ordinary transaction and atomic transfer creation, with durable server idempotency, explicit recovery states, and no duplicate ledger writes.

**Architecture:** PostgreSQL remains the source of truth. The Go API records a durable mutation receipt in the same transaction as the ledger mutation; the Swift client persists an encrypted, account-scoped outbox and drains it FIFO through the existing financial endpoints. V1 is append-only and online-authoritative: pending rows are visibly pending, successful sync invalidates/refetches reads, and conflicts are surfaced for user resolution rather than merged silently.

**Tech Stack:** Go/Gin/GORM/PostgreSQL migrations and repository tests; Swift 6.4/SwiftUI/iOS 17, URLSession async/await, Keychain and the existing `APIClient`; XCTest and PostgreSQL integration fixtures.

**Spec:** `docs/superpowers/specs/2026-10-06-offline-financial-sync-contract.md`

## Global Constraints

- PostgreSQL is the only financial source of truth; Redis is never used for mutation durability.
- V1 offline writes are limited to ordinary income/expense creation and atomic wallet-to-wallet transfer creation.
- Every offline mutation persists one UUID `Idempotency-Key` before its first attempt and reuses it for every retry.
- `owner_id` is derived from JWT/API-key auth; client payloads cannot select an owner.
- Pending local rows are not included in authoritative server totals.
- No automatic last-write-wins, bank/payment action, offline credit math, or offline AI/OCR is introduced.
- No token, provider secret, or raw financial payload is written to diagnostics.
- Existing web clients must continue to work through an explicit API migration or a coordinated key-required rollout; missing keys must never silently duplicate a write.

## Review Focus

- Same key after a timeout must replay one committed transaction: covered by Task 2 concurrent PostgreSQL tests.
- Same key with a changed body must return `409` without a second write: covered by Task 2 hash/replay tests.
- Two accounts using the same key must not see or replay each other’s result: covered by Task 2 owner-scope tests and Task 6 API tests.
- App termination during `sending` must resume safely: covered by Task 4 store/lease tests.
- Offline UI must not present a pending amount as confirmed balance: covered by Task 5 SwiftUI state tests and Task 6 simulator flow.

### Task 1: Freeze the public contract and fixtures

**Files:**
- Create: `docs/superpowers/specs/2026-10-06-offline-financial-sync-contract.md`
- Create: `docs/superpowers/plans/2026-10-06-offline-financial-sync.md`
- Modify: `docs/architecture/API.md`
- Modify: `docs/architecture/DATABASE.md`
- Test: `backend/internal/controller/http/offline_contract_test.go`

**Interfaces:**
- Consumes: existing transaction/transfer routes and Problem Details envelope.
- Produces: exact operation names, required headers, replay/conflict bodies, V1 scope, and fixture payloads used by later tasks.

- [x] **Step 1: Record the approved contract.** Keep the spec in `draft` until owner approves V1 scope and conflict wording; do not implement from an ambiguous “sync everything” requirement.
- [ ] **Step 2: Add contract fixtures and failing route assertions.** Assert required `Idempotency-Key`, response `mutation.replayed`, `IDEMPOTENCY_KEY_REUSED`, and the two allowed V1 operations.
- [ ] **Step 3: Run the focused contract test.** Run `cd backend && go test ./internal/controller/http -run OfflineContract -count=1`; Expected: FAIL because the new headers/fields are not implemented.
- [ ] **Step 4: Reconcile public API/database docs after the contract is accepted.** Do not hand-edit generated Swagger; mark generated output for the implementation task.

### Task 2: Add durable owner-scoped idempotency receipts in Go

**Files:**
- Create: `backend/migrations/000030_financial_mutation_receipts.up.sql`
- Create: `backend/migrations/000030_financial_mutation_receipts.down.sql`
- Create: `backend/internal/entity/financial_mutation_receipt.go`
- Create: `backend/internal/repository/financial_mutation_receipt.go`
- Create: `backend/internal/infrastructure/repository/financial_mutation_receipt_postgres.go`
- Modify: `backend/internal/repository/transaction_mutation.go`
- Test: `backend/internal/infrastructure/repository/financial_mutation_receipt_postgres_test.go`

**Interfaces:**
- Consumes: `ownerID`, operation, idempotency key, canonical request hash, and a transaction callback.
- Produces: `ReserveOrReplay(ctx, ownerID, operation, key, hash)`, `Complete(ctx, receiptID, resourceType, resourceID, responseJSON)`, and typed conflict/replay results.

- [ ] **Step 1: Write failing PostgreSQL tests.** Cover first reservation, identical replay, changed-body conflict, same key across owners, concurrent same-key callers, rollback before commit, and retention metadata.
- [ ] **Step 2: Run tests to verify the receipt layer fails.** Run `cd backend && go test ./internal/infrastructure/repository -run FinancialMutationReceipt -count=1`; Expected: missing migration/repository behavior.
- [ ] **Step 3: Implement the migration and repository.** Use a unique `(owner_id, operation, idempotency_key)`, SHA-256 canonical JSON hash, row locking, and completion JSON stored in the same SQL transaction as the ledger write.
- [ ] **Step 4: Run the focused and race tests.** Run `go test ./internal/infrastructure/repository -run FinancialMutationReceipt -count=1` and `go test -race ./internal/infrastructure/repository -run FinancialMutationReceipt -count=1`; Expected: PASS with one committed resource for concurrent replay.
- [ ] **Step 5: Regenerate Swagger and update `DATABASE.md`.** Record migration version and rollback behavior.

### Task 3: Make transaction and transfer creation replay-safe

**Files:**
- Modify: `backend/internal/controller/http/transaction_handler.go`
- Modify: `backend/internal/controller/http/routes_transaction.go`
- Modify: `backend/internal/usecase/transaction.go`
- Modify: `backend/internal/repository/transaction.go`
- Modify: `backend/internal/repository/transaction_mutation.go`
- Modify: `backend/internal/infrastructure/repository/transaction_postgres.go`
- Test: `backend/internal/controller/http/transaction_handler_test.go`
- Test: `backend/internal/infrastructure/repository/transaction_postgres_test.go`

**Interfaces:**
- Consumes: Task 2 receipt repository and existing owner-scoped transaction/transfer use cases.
- Produces: required-key `POST /transactions` and `POST /transactions/transfer` contracts with replay metadata; unchanged atomic transfer semantics.

- [ ] **Step 1: Write RED handler/repository tests.** Cover missing key, first create, identical replay, changed body `409`, cross-owner key isolation, and transfer replay with exactly two rows.
- [ ] **Step 2: Run the focused tests and confirm failure.** Run `go test ./internal/controller/http ./internal/infrastructure/repository -run 'Transaction.*Idempot|Transfer.*Idempot|Offline' -count=1`.
- [ ] **Step 3: Implement canonical request hashing and receipt-wrapped writes.** Hash operation plus normalized body, reserve under owner scope, execute one domain transaction, complete the receipt, and return the original response on replay. Never log body/amount.
- [ ] **Step 4: Run focused tests, full Go tests, race and vet.** Commands: `go test ./... -count=1`, `go test -race ./... -count=1`, `go vet ./...`; Expected: all PASS.
- [ ] **Step 5: Regenerate Swagger/OpenAPI and reconcile `docs/architecture/API.md`.** Include required header, error codes, and replay response.

### Task 4: Build the Swift account-scoped outbox store

**Files:**
- Create: `mobile-ios/MyPocketIOS/Core/Offline/OutboxModels.swift`
- Create: `mobile-ios/MyPocketIOS/Core/Offline/OutboxStore.swift`
- Create: `mobile-ios/MyPocketIOS/Core/Offline/OutboxDatabase.swift`
- Modify: `mobile-ios/MyPocketIOS/Core/Auth/` (only the existing logout/session boundary)
- Test: `mobile-ios/MyPocketIOSTests/OutboxStoreTests.swift`

**Interfaces:**
- Consumes: `OutboxMutation(accountID:operation:payload:idempotencyKey:)` and the existing Keychain/session owner.
- Produces: `enqueue`, `leaseNext`, `markRetry`, `markApplied`, `markConflict`, `markRejected`, `resumeLeases`, `blockForAuth`, and `pendingCount` with FIFO ordering.

- [ ] **Step 1: Write failing XCTest cases.** Cover atomic enqueue-before-send, stable key across restart, FIFO per account, account isolation, lease recovery after termination, and no bearer token in serialized payload.
- [ ] **Step 2: Run the focused XCTest target to verify RED.** Run `xcodebuild test -scheme MyPocketIOS -destination 'platform=iOS Simulator,name=iPhone 16' -only-testing:MyPocketIOSTests/OutboxStoreTests`.
- [ ] **Step 3: Implement persistent storage.** Use the project’s iOS 17 persistence choice (SwiftData/SQLite boundary), encrypt at rest with a Keychain-held database key, and keep storage behind `OutboxStore` so the sync engine never touches tables directly.
- [ ] **Step 4: Run focused and full native unit tests.** Expected: all existing tests plus the new outbox tests pass.

### Task 5: Implement the sync actor and typed recovery states

**Files:**
- Create: `mobile-ios/MyPocketIOS/Core/Offline/SyncEngine.swift`
- Create: `mobile-ios/MyPocketIOS/Core/Offline/SyncPolicy.swift`
- Modify: `mobile-ios/MyPocketIOS/Core/API/APIClient.swift`
- Modify: `mobile-ios/MyPocketIOS/Core/API/APIError.swift`
- Modify: `mobile-ios/MyPocketIOS/Core/Models/FinanceModels.swift`
- Test: `mobile-ios/MyPocketIOSTests/SyncEngineTests.swift`
- Test: `mobile-ios/MyPocketIOSTests/APIClientTests.swift`

**Interfaces:**
- Consumes: Task 4 outbox store, reachability events, auth refresh, and Task 3 HTTP response contract.
- Produces: `SyncEngine.trigger(reason:)`, typed `pending/applied/retryWait/conflict/rejected/blockedAuth` state, and affected-read invalidation after acknowledgement.

- [ ] **Step 1: Write failing fake-transport tests.** Cover `201` apply, timeout then replay, `401` one refresh, `429/5xx` backoff, `400/422` rejection, `409` conflict without retry, and cancellation on logout.
- [ ] **Step 2: Run the focused tests and confirm RED.** Run the `SyncEngineTests` XCTest target.
- [ ] **Step 3: Implement one-account FIFO draining.** Retry only the specified transient classes, cap backoff at 15 minutes with jitter, reuse the exact idempotency key, and never auto-resolve conflicts.
- [ ] **Step 4: Add cache invalidation hooks.** On success, refetch wallet/transactions/reports; pending projections remain separate from authoritative totals.
- [ ] **Step 5: Run native unit tests, race-free backend tests and static guards.** Expected: no test uses real credentials or writes production data.

### Task 6: Add native UI states using base components

**Files:**
- Create: `mobile-ios/MyPocketIOS/DesignSystem/MPOfflineStatus.swift`
- Modify: `mobile-ios/MyPocketIOS/Features/Transactions/TransactionsView.swift`
- Modify: `mobile-ios/MyPocketIOS/Features/Overview/OverviewView.swift`
- Modify: `mobile-ios/MyPocketIOS/Features/More/MoreView.swift`
- Test: `mobile-ios/MyPocketIOSUITests/OfflineSyncUITests.swift`

**Interfaces:**
- Consumes: Task 5 sync state and existing design-system primitives.
- Produces: pending badge, stale-read banner, retry/conflict/discard action card, and account-safe logout prompt; no feature-local raw buttons where a base component exists.

- [ ] **Step 1: Write UI RED tests.** Assert a queued transaction is visibly pending, excluded from confirmed totals, retry is offered for transient error, and conflict offers edit/discard but no silent overwrite.
- [ ] **Step 2: Run the UI target to confirm RED.** Use the simulator with a deterministic fake transport fixture.
- [ ] **Step 3: Implement the shared status/card primitives and wire feature views.** Preserve the existing native semantic tokens and accessibility labels.
- [ ] **Step 4: Run the native UI suite.** Expected: all existing parity tests plus offline UI tests pass with zero network dependency.

### Task 7: End-to-end, migration and release verification

**Files:**
- Create: `backend/internal/controller/http/offline_sync_integration_test.go`
- Modify: `docs/architecture/API.md`
- Modify: `docs/architecture/DATABASE.md`
- Modify: `docs/architecture/ARCHITECTURE.md`
- Modify: `docs/work/VALIDATION_MATRIX.md`
- Modify: `docs/CONTEXT.md`
- Modify: `docs/releases/CHANGELOG.md`
- Modify: `docs/superpowers/specs/2026-10-06-offline-financial-sync-contract.md` (status only after approval/verification)

**Interfaces:**
- Consumes: Tasks 2–6 and a disposable PostgreSQL database plus iPhone 16 Simulator.
- Produces: reproducible proof for offline enqueue, restart, replay, conflict, read refetch, and account isolation.

- [ ] **Step 1: Run migration clean/version checks.** Apply the new migration to a disposable database and verify `dirty=false` and rollback on a disposable clone.
- [ ] **Step 2: Run API integration scenarios.** Disconnect the fake transport, enqueue two mutations, restart the app, reconnect, assert two server rows and one receipt per key; replay each key and assert no row count increase.
- [ ] **Step 3: Run failure scenarios.** Assert changed-body key reuse returns `409`, cross-owner key use cannot replay, and a missing/deleted wallet becomes a user-visible rejected row.
- [ ] **Step 4: Run required quality gates.** `go test ./...`, `go test -race ./...`, `go vet ./...`, native unit/UI suites, `check-base-components`, `verify-project`, `git diff --check`, and web build/design checks.
- [ ] **Step 5: Reconcile docs and release notes.** Record exact commands/results, update API/OpenAPI and database docs, and keep the contract `draft` or mark `verified` only after owner approval and evidence.

## Execution order and gate

Tasks 1–3 are one backend contract gate; Tasks 4–6 are one native client gate;
Task 7 is the release gate. Do not implement Tasks 4–6 against mocked
idempotency semantics. If the owner wants offline edits/deletes, credit,
budgets, or background sync, create a second detail design with explicit
`base_version` and conflict-resolution rules before extending this plan.
