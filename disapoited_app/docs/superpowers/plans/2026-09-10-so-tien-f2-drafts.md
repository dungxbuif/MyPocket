# Sổ tiền F2 — recurring draft decisions

**Goal:** turn existing recurring transaction drafts into a safe, owner-scoped confirmation/rejection journey that produces accounting exactly once and is usable from both cookie and API-key clients.

**Approved design:** [recurring-draft contract](../specs/2026-09-09-recurring-draft-DETAIL_DESIGN.md). Owner authorized continuous implementation, tests and deployment on 2026-09-10.

**Constraints:** Preserve the dirty checkout; no commits/staging; do not change production data or add dependencies. Use PostgreSQL-backed tests, integer VND, server-derived owner identity, CSRF for cookie mutations and existing bearer semantics for API keys. No offline queue for decisions; retain inputs and offer explicit retry. Do not expose raw errors.

## Task 1 — Atomic backend decision boundary

**Owns:** `backend/internal/finance/repository.go` plus direct finance tests; `backend/internal/planning/{types.go,repository.go}` plus direct planning tests; `backend/internal/platform/httpapi/{planning.go,router.go,planning_test.go}` and directly required stubs/tests only.

1. First write PostgreSQL regressions for confirm expense/income/transfer, same-action replay, conflicting terminal action, stale version, foreign owner, request-key collision across two drafts, and rollback after injected failure. Add HTTP cookie+CSRF/API-key/missing-key envelope tests.
2. Extract transaction-aware finance creation without changing direct `CreateTransaction` semantics. Planning locks the owned draft, resolves terminal replay/conflict, invokes accounting in the same SQL transaction, then atomically updates accepted amount/note/status/link/version.
3. Extend list/response with `confirmed_transaction_id`; add bounded confirm/reject repository/API methods and routes. Derive the finance idempotency key using a draft-scoped hash, never reuse raw client key across drafts.
4. Run focused RED→GREEN, package Go tests and `go test -p 1 ./...` against disposable PostgreSQL. Report exact commands/results and risks.

## Task 2 — Mounted draft decision UI

**Depends on Task 1. Owns:** `frontend/src/app/planning.ts`, `frontend/src/app/App.tsx` only draft action wiring, and new/changed direct App/Planning tests. No style migration.

1. Add API adapters typed to stable envelopes; preserve same idempotency key across user retry but never auto-retry.
2. Mount pending drafts in the existing planning surface with confirm/reject, amount/note edit, busy/disabled state and shared safe `OperationError`. Terminal draft exposes linked transaction identity; server success triggers targeted refresh.
3. Add real App fetch-boundary tests for successful expense/transfer display, failure/input retention/retry, duplicate click, and terminal replay. Run focused tests/typecheck.

## Task 3 — Integration, docs and release gate

**Parent-owned:** PostgreSQL browser E2E for confirm/reject/reload/exact wallet effects; full frontend/build, full Go suite, API docs/OpenAPI reconciliation, Docusaurus workflow/action mapping, context/backlog/validation/changelog, full browser suite and release verification. Deploy only after all fresh checks and known WebKit/device limitations are documented; record image/version/rollback evidence.

This plan addresses only draft decisions. Category hierarchy/settings, adjustment zero/negative semantics, schedule update and remaining report contracts receive separate F2 plans; they are not dropped.
