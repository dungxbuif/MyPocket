---
artifact_type: test_verification
id: PHASE-002-finance-core
status: ready
owner: shared
human_fields:
  - uat_sign_off
  - manual_acceptance_notes
ai_fields:
  - commands
  - automated_tests
  - manual_checks
  - failure_summary
  - evidence_notes
  - attempt_log
shared_fields:
  - status
  - trace
  - uat
trace:
  backlog_item: BL-002
  requirement:
    - REQ-F-002
    - REQ-F-003
    - REQ-F-011
    - REQ-NF-002
    - REQ-NF-005
  phase: PHASE-002
  ticket_or_bug:
    - TICKET-005
    - TICKET-006
    - TICKET-007
  detail_design: ../phases/PHASE-002-detail-design.md
  validation_matrix: ../VALIDATION_MATRIX.md
  docs_review: per-ticket_completion_checklist
  release_notes: ../../releases/CHANGELOG.md
---

# PHASE-002 Finance Core Verification

## Status

- ID: PHASE-002-finance-core
- Status: ready
- Owner: shared

## Scope

- Ticket/Bug/Phase: PHASE-002, TICKET-005, TICKET-006, TICKET-007
- Tested behavior: wallet/category domain, transaction accounting, idempotency, Vietnamese seed data, receipt metadata foundation, and mobile finance workflows.

## Trace Links

- Backlog item: [BL-002](../BACKLOG.md)
- Requirement: [REQ-F-002, REQ-F-003, REQ-F-011, REQ-NF-002, REQ-NF-005](../../requirements/REQUIREMENTS.md)
- Phase: [PHASE-002](../phases/PHASE-002-finance-core.md)
- Tickets: [TICKET-005](../tickets/TICKET-005-wallet-category-domain.md), [TICKET-006](../tickets/TICKET-006-transaction-accounting-engine.md), [TICKET-007](../tickets/TICKET-007-vietnamese-seeds-receipt-metadata.md)
- Detail design: [PHASE-002-detail-design.md](../phases/PHASE-002-detail-design.md)
- Validation matrix: [VALIDATION_MATRIX.md](../VALIDATION_MATRIX.md)
- Release notes: [CHANGELOG.md](../../releases/CHANGELOG.md)

## Commands

| Command | Result | Notes |
| --- | --- | --- |
| `rtk env 'MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/platform/db -run TestPhase002FinanceTablesAndSeeds -count=1` | pass | RED first failed because `wallets` did not exist; GREEN passed after adding `0002_phase002_finance_core.sql`. |
| `rtk env 'MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/platform/db -count=1` | pass | Full migration package regression after PHASE-002 migration. |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/finance -count=1` | pass | Unit-only finance validation proof; 4 tests passed when PostgreSQL URL was absent. |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache 'MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/finance -count=1` | pass | Wallet/category unit and PostgreSQL integration proof for TICKET-005 backend domain/repository behavior. |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/platform/httpapi -run 'TestWalletsAPI|TestCreateWallet|TestCategoriesAPI' -count=1` | pass | HTTP handler proof for wallet/category auth requirement, authenticated user scoping, CSRF requirement, create wallet request mapping, and category listing. |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache 'MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test -p 1 ./... -count=1` | pass | Backend regression after finance repository; API, worker, finance, identity, config, db, httpapi, objectstore packages passed. `-p 1` avoids package-level races because integration helpers reset the same local test schema. |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/finance -run 'TestApplyAccountingEffect' -count=1` | pass | RED first failed because accounting types/functions were missing; GREEN passed after adding income, expense, transfer, adjustment effect validation. |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/finance -count=1` | pass | Unit-only finance regression after TICKET-006 accounting primitives. |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/finance -count=1` | pass | Finance package regression with DB-backed repository tests after accounting primitives. |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test -p 1 ./... -count=1` | pass | Backend regression after TICKET-006 accounting primitives. |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/finance -run 'TestRepositoryCreates.*Transaction\|TestRepositoryReplaysDuplicateTransactionIdempotencyKey\|TestRepositoryRejectsTransactionForAnotherUsersWallet' -count=1` | pass | RED first failed because `CreateTransaction`/`CreateTransactionInput` were missing; GREEN passed after repository transaction create implementation. |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/finance -run 'TestRepository.*Transaction' -count=1` | pass | Repository transaction suite covers create income/expense/transfer/adjustment, idempotent replay, idempotency conflict rejection, inactive category rejection, and cross-user wallet rejection. |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/finance -count=1` | pass | Unit-only finance regression after repository transaction create implementation. |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/finance -count=1` | pass | Finance package regression with DB-backed transaction repository tests. |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test -p 1 ./... -count=1` | pass | Backend regression after repository transaction create implementation. |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/finance -run 'TestRepositoryUpdatesTransaction\|TestRepositoryArchivesTransaction\|TestRepositoryListsTransactions' -count=1` | pass | RED first failed because edit/archive/search repository APIs were missing; GREEN passed after implementation. |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/platform/db -run TestPhase002FinanceTablesAndSeeds -count=1` | pass | Migration proof after adding transaction delta columns. Parallel DB-backed commands briefly failed from shared schema-reset contention, then passed sequentially. |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/finance -count=1` | pass | Finance package regression after edit/archive/search repository implementation. |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test -p 1 ./... -count=1` | pass | Backend regression after edit/archive/search repository implementation. |
| `rtk npm test -- --run src/app/App.test.tsx` | pass | RED first failed because wallet/category UI still used hardcoded data; GREEN passed after API-driven finance state was added. |
| `rtk npm test -- --run` | pass | Full frontend component suite; 5 tests passed. |
| `rtk npm run build` | pass | Production Vite build completed. |
| `rtk npm run test:e2e` | pass | Existing mobile PWA/auth/offline browser smoke passed after finance UI changes; 3 tests passed. |

## Fix/Test Attempt Log

| Attempt | Change Made | Command | Result | Failure Summary |
| --- | --- | --- | --- | --- |
| 1 | Added PHASE-002 migration, then changed `categories_system_key_unique` from partial to full unique index so `ON CONFLICT (system_key)` is valid. | `rtk env 'MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/platform/db -run TestPhase002FinanceTablesAndSeeds -count=1` | pass | Initial GREEN attempt failed with SQLSTATE 42P10 because PostgreSQL cannot use a partial unique index for that conflict target. |
| 2 | Added finance validation and repository methods after RED tests exposed missing package/methods; adjusted `UpdateCategoryInput` to match the test-facing API. | `rtk env GOCACHE=/private/tmp/mypocket-go-cache 'MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/finance -count=1` | pass | Initial finance GREEN compile failed because `UpdateCategoryInput.Name` was a pointer while tests and plan used a string field. |
| 3 | Reran backend regression sequentially after package-parallel integration tests raced on shared schema reset. | `rtk env GOCACHE=/private/tmp/mypocket-go-cache 'MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test -p 1 ./... -count=1` | pass | Plain package-parallel `go test ./...` can reset the shared local test schema while another package is using it. |
| 4 | Added finance HTTP dependency, wallet/category handlers, and API runtime wiring after RED tests exposed missing finance dependency/route support. | `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/platform/httpapi -run 'TestWalletsAPI|TestCreateWallet|TestCategoriesAPI' -count=1`; backend regression command above | pass | No remaining HTTP handler failure for implemented wallet/category routes. |
| 5 | Added frontend finance client and API-driven mobile wallet/category rendering after RED component tests showed hardcoded wallet/category data. | `rtk npm test -- --run`; `rtk npm run build`; `rtk npm run test:e2e` | pass | One assertion was adjusted after the UI correctly showed the same total in the header and wallet summary. |
| 6 | Added TICKET-006 accounting effect tests and primitives for income, expense, transfer, and adjustment. | `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/finance -run 'TestApplyAccountingEffect' -count=1`; `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test -p 1 ./... -count=1` | pass | Initial RED compile failed because `TransactionType`, `AccountingInput`, and `ApplyAccountingEffect` did not exist. |
| 7 | Added repository create transaction persistence with wallet locks, atomic balance/version updates, category activation checks, and idempotency replay storage. | `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/finance -run 'TestRepository.*Transaction' -count=1`; `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test -p 1 ./... -count=1` | pass | Initial RED compile failed because `CreateTransaction` and `CreateTransactionInput` did not exist. |
| 8 | Added transaction delta persistence plus repository edit reversal/reapply, archive reversal once, and filtered listing. | `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/finance -run 'TestRepositoryUpdatesTransaction\|TestRepositoryArchivesTransaction\|TestRepositoryListsTransactions' -count=1`; `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test -p 1 ./... -count=1` | pass | Initial RED compile failed because edit/archive/search DTOs and repository methods did not exist. |

Loop guard:

- Same-path failure attempts: 1 / 3
- Active work-item fix/test cycles: TICKET-006 is 3 / 5. Earlier TICKET-005 cycles are tracked in that ticket.
- Blocked: no
- Human/design input needed: none before starting approved PHASE-002 plan.

## Automated Tests

- Passed: PHASE-002 migration/schema/seed test, full `internal/platform/db` package, wallet/category finance package tests, wallet/category HTTP handler tests, TICKET-006 accounting effect tests, TICKET-006 repository create/idempotency/edit/archive/search tests, full backend regression, frontend component tests, frontend build, and mobile PWA/auth/offline E2E smoke.
- Failed: none remaining for migration, backend wallet/category domain, implemented wallet/category API route, and API-driven mobile display slices.
- Skipped: remaining wallet/category mutation routes, transaction API/mobile workflows, and full wallet/category/transaction UAT pending later PHASE-002 work.

## Manual Checks

- Pending mobile UI implementation.

## UAT

- Required: yes for wallet/category and transaction workflows; not required for receipt metadata foundation until PHASE-006.
- Reason if not required: partial exception applies only to non-user-facing receipt metadata foundation.
- Expected behavior: finance workflows use Vietnamese copy, exact VND integer formatting, user-owned data, and correct wallet balances.
- Verified behavior: wallet/category API-driven display, backend accounting effects, repository create transaction/idempotency behavior, edit/archive reversal, and transaction search filters have automated proof; transaction API/mobile workflows remain pending implementation.
- Sign-off: pending.

## Failures And Follow-Up

- Migration GREEN attempt 1 exposed a partial-index conflict-target issue; fixed in the migration before committing.
- Finance validation GREEN attempt 1 exposed a test-facing type mismatch; fixed before repository implementation continued.
- Package-parallel backend regression with a shared DB URL exposed test harness schema-reset contention; rerun with `-p 1` passed.

## Evidence Notes

- PHASE-002 migration test was written and observed failing before production migration code existed: `wallets table missing id column: map[string]bool{}`.
- After adding migration `0002_phase002_finance_core.sql` and fixing the category unique index, `internal/platform/db` tests passed against the local Compose PostgreSQL test URL.
- Wallet/category repository tests were written and observed failing on missing methods before repository code existed, then passed against local Compose PostgreSQL after implementation.
- Wallet/category HTTP tests were written and observed failing on missing dependency support before route code existed, then passed after handlers and runtime wiring were added.
- Frontend wallet/category tests were written and observed failing on hardcoded display behavior, then passed after `frontend/src/app/finance.ts` and App state wiring were added.
- TICKET-006 accounting tests were written and observed failing on missing finance accounting types/functions, then passed after `backend/internal/finance/transactions.go` and related types were added.
- TICKET-006 repository tests were written and observed failing on missing `CreateTransaction` APIs, then passed after transaction persistence, wallet balance updates, category checks, and idempotency storage were added.
- TICKET-006 edit/archive/search repository tests were written and observed failing on missing methods/types, then passed after persisted transaction deltas and repository methods were added.
