# Ledger Completion Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Complete the web ledger's high-risk mutation gaps with atomic linked-transfer mutation, explicit balance adjustments, and safe bulk deletion.

**Architecture:** Keep PostgreSQL as the financial source of truth and enforce owner scope inside repository transactions. Add narrow repository contracts and handlers, then compose the existing atomic web controls for selection, confirmation, and result states. Mutation replay protection uses a durable owner/key/request-hash record rather than process memory.

**Tech Stack:** Go, Gin, GORM, PostgreSQL migrations, React 19, TypeScript, Vite, Node test scripts.

**Spec:** `docs/superpowers/specs/2026-10-01-release-completion-design.md`

## Global Constraints

- Money remains positive integer VND for ordinary rows; direction is explicit for adjustments.
- Every mutation is owner-scoped in both handler/use case and repository transaction.
- Linked transfers mutate exactly two rows or no rows; ordinary row mutation rejects linked rows.
- Adjustments are visible in wallet reconciliation but excluded from ordinary income/expense reports.
- Bulk deletion is all-or-nothing and never partially deletes a linked transfer.
- Use existing base components; no local native controls or one-off visual tokens.
- Generated Swagger and public docs must match implemented routes.

## Review Focus

- A transfer ID with zero, one, or more than two rows must fail without changing data — repository and handler tests in Tasks 1–2.
- A valid transfer mutation cannot cross owners or change a credit wallet — repository test in Task 1.
- Reusing an idempotency key with different JSON must return `409` and not replay a mutation — Task 2.
- An adjustment must change wallet balance without appearing in income/expense totals — Task 3.
- A bulk request containing one linked row must delete nothing — Task 4.

### Task 1: Atomic transfer pair mutation

**Files:**
- Modify: `backend/internal/repository/transfer.go`
- Modify: `backend/internal/infrastructure/repository/transaction_postgres.go`
- Test: `backend/internal/infrastructure/repository/transaction_postgres_test.go`
- Test: `backend/internal/controller/http/transaction_handler_test.go`

**Interfaces:**
- Consumes: existing `TransferID`, system transfer categories, and owner-scoped transaction repository.
- Produces: `UpdateTransfer(ownerID, transferID string, updates TransferUpdate) ([]entity.Transaction, error)` and `DeleteTransfer(ownerID, transferID string) error` with exact-pair validation.

- [ ] Write failing repository tests for amount/note/date update of both rows, owner rejection, malformed cardinality, and atomic delete.
- [ ] Run `go test ./internal/infrastructure/repository -run 'Transfer'` and verify the new tests fail for missing methods.
- [ ] Implement the narrow transfer contract and a PostgreSQL transaction that locks the pair, verifies both expected system categories/types, updates both rows, and deletes both rows atomically.
- [ ] Add handler-level tests proving linked single-row `PATCH/DELETE` is rejected and transfer mutation returns both rows.
- [ ] Run the focused repository and HTTP tests and verify they pass.
- [ ] Commit `feat: mutate linked transfer pairs atomically`.

### Task 2: Durable mutation idempotency and routes

**Files:**
- Create: `backend/migrations/000022_transaction_mutation_idempotency.up.sql`
- Create: `backend/internal/entity/transaction_mutation.go`
- Create: `backend/internal/repository/transaction_mutation.go`
- Modify: `backend/internal/infrastructure/repository/transaction_postgres.go`
- Modify: `backend/internal/controller/http/transaction_handler.go`
- Modify: `backend/internal/controller/http/routes_transactions.go`
- Modify: `backend/cmd/api/main.go`
- Modify: `backend/docs/docs.go`, `backend/docs/swagger.json`, `backend/docs/swagger.yaml`
- Test: `backend/internal/controller/http/transaction_handler_test.go`

**Interfaces:**
- Consumes: Task 1 transfer mutation methods.
- Produces: `PATCH/DELETE /api/v1/transactions/transfer/:transfer_id`, an `Idempotency-Key` header contract, and owner/key/body replay protection.

- [ ] Write failing handler tests for first mutation, same-key replay, changed-body conflict, and missing-key rejection.
- [ ] Run the focused tests and verify the route/record behavior fails before implementation.
- [ ] Add an additive migration with a unique `(owner_id, operation, idempotency_key)` record and request hash/response metadata.
- [ ] Implement reserve/replay/complete semantics inside the same database transaction as the transfer mutation; never log body or financial values.
- [ ] Register routes, add Swagger annotations, regenerate docs, and preserve the existing single-row linked-pair guard.
- [ ] Run focused HTTP tests, `go generate ./cmd/api`, and `go test ./...`.
- [ ] Commit `feat: add idempotent transfer mutation routes`.

### Task 3: Explicit balance adjustments

**Files:**
- Create: `backend/migrations/000023_transaction_adjustments.up.sql`
- Modify: `backend/internal/entity/wallet.go`
- Modify: `backend/internal/repository/transaction.go`
- Modify: `backend/internal/infrastructure/repository/transaction_postgres.go`
- Modify: `backend/internal/infrastructure/repository/finance_query_postgres.go`
- Modify: `backend/internal/controller/http/transaction_handler.go`
- Modify: `backend/internal/controller/http/routes_transactions.go`
- Modify: `app/src/services/transactions.ts`
- Modify: `app/src/atomic/organisms/TransactionsPanel.tsx`
- Test: `backend/internal/controller/http/transaction_handler_test.go`
- Test: `backend/internal/infrastructure/repository/transaction_postgres_test.go`
- Test: `app/scripts/transactions.test.ts`

**Interfaces:**
- Consumes: owner-scoped wallet and transaction APIs.
- Produces: `POST /api/v1/transactions/adjustment` with immutable adjustment rows and wallet reconciliation semantics.

- [ ] Write failing tests for positive/negative direction, credit rejection, balance effect, report exclusion, and immutable edit/delete behavior.
- [ ] Run focused backend/frontend tests and verify the adjustment contract is absent.
- [ ] Add the adjustment discriminator/direction schema and update balance/report queries without changing ordinary income/expense behavior.
- [ ] Implement the protected API and a shared confirmation sheet in the ledger menu; show adjustment rows distinctly.
- [ ] Run backend suite, frontend transaction/design/type tests, and build.
- [ ] Commit `feat: record explicit wallet balance adjustments`.

### Task 4: Safe bulk deletion

**Files:**
- Modify: `backend/internal/repository/transaction.go`
- Modify: `backend/internal/infrastructure/repository/transaction_postgres.go`
- Modify: `backend/internal/controller/http/transaction_handler.go`
- Modify: `backend/internal/controller/http/routes_transactions.go`
- Modify: `app/src/services/transactions.ts`
- Modify: `app/src/atomic/organisms/TransactionsPanel.tsx`
- Test: `backend/internal/controller/http/transaction_handler_test.go`
- Test: `app/scripts/transactions.test.ts`

**Interfaces:**
- Consumes: Task 1 transfer identity rules and Task 3 adjustment visibility.
- Produces: `POST /api/v1/transactions/bulk-delete` with all-or-nothing owner-scoped deletion and explicit rejection reasons.

- [ ] Write failing tests for empty/oversized lists, cross-owner IDs, linked transfers, and successful all-or-nothing deletion.
- [ ] Run focused tests and verify they fail before the route exists.
- [ ] Implement a single repository transaction that locks all requested rows, validates the full set, then deletes or rejects the entire request.
- [ ] Add selection state, confirmation, loading/error feedback, and refresh behavior using existing base components.
- [ ] Run focused tests, full backend tests, frontend checks, and build.
- [ ] Commit `feat: add safe bulk transaction deletion`.

### Task 5: Contract and release reconciliation

**Files:**
- Modify: `docs/architecture/API.md`
- Modify: `docs/design/pages/transactions/README.md`
- Modify: `docs/work/VALIDATION_MATRIX.md`
- Modify: `docs/work/BACKLOG.md`
- Modify: `docs/releases/CHANGELOG.md`

- [ ] Document transfer mutation, adjustment, bulk deletion, errors, idempotency, and UI states.
- [ ] Record automated evidence and remaining browser/UAT/deployment gates.
- [ ] Run `git diff --check`, `go generate ./cmd/api`, `go test ./...`, `npm run check:design`, `npm run typecheck`, and `npm run build`.
- [ ] Commit `docs: reconcile ledger completion contract and proof`.
