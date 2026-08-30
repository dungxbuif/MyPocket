---
artifact_type: ticket
id: TICKET-006
status: in_progress
owner: human
priority: high
lane: high-risk
human_fields:
  - title
  - priority
  - acceptance_criteria
  - scope
  - approval
ai_fields:
  - impacted_areas
  - test_expectations
  - verification_results
  - docs_review
  - context_updates
shared_fields:
  - status
  - trace
  - small_task_exemption
trace:
  backlog_item: BL-002
  requirement:
    - REQ-F-003
    - REQ-NF-002
    - REQ-NF-005
  phase: PHASE-002
  detail_design: ../phases/PHASE-002-detail-design.md
  implementation_plan: ../../superpowers/plans/2026-08-30-phase-002-finance-core.md
  test_verification: ../test-verification/PHASE-002-finance-core.md
  validation_matrix: ../VALIDATION_MATRIX.md
  docs_review: per-ticket_completion_checklist
  adrs:
    - ADR-001
    - ADR-003
  release_notes: ../../releases/CHANGELOG.md
---

# Ticket: TICKET-006 Transaction Accounting Engine

## Status

- ID: TICKET-006
- Status: in_progress
- Type: feature
- Priority: high
- Phase: PHASE-002
- Owner: human

## Context

Human fill:

- User/business/system problem: Users need reliable income, expense, transfer, adjustment, edit, archive, and search flows with balances that stay correct under retry.
- Source prompt or requirement: REQ-F-003, REQ-NF-002, REQ-NF-005, and PHASE-002.
- Out of scope: offline change feed, analytics aggregates, AI/OCR draft generation, recurring schedules.

AI fill:

- Current repository context read: PHASE-002, requirements, ERD, API, ADR-001, ADR-003, and existing identity/platform code.
- Brownfield touched scope, if applicable: new finance transaction service, migration tables, HTTP handlers, and mobile add/search flows.

## Acceptance Criteria

- [ ] Given an authenticated user and active wallet/category, when an income or expense is created, then the transaction amount is a positive VND integer and the source wallet balance changes atomically with an incremented version.
- [ ] Given two user-owned wallets, when a transfer is created, then source and destination wallet effects commit atomically, cannot use the same wallet twice, and retrying the same idempotency key cannot duplicate effects.
- [ ] Given a balance adjustment, when it is created, then the wallet balance becomes the target amount through an auditable adjustment transaction.
- [ ] Given an existing transaction, when it is edited or archived, then wallet balances are reversed and reapplied exactly once inside one database transaction.
- [ ] Given search filters, when transactions are listed, then results are user-scoped and support wallet, category, type, date range, query text, and report-excluded filters.
- [ ] UAT requirement is required for add/edit/search, transfer, adjustment, report exclusion, and VND integer formatting.

## Small Task Exemption

- Small task exemption: no
- Reason: This ticket implements the core accounting engine and public finance API.
- Impact checked: API=yes, DB=yes, Security=yes, Runtime=no, Standards=no

## Impacted Areas

- Code: finance domain/application code, migrations, HTTP routes, frontend add/search screens.
- Requirements docs: no change expected.
- Architecture docs: reconcile implemented accounting boundary.
- API docs: add concrete transaction contracts and idempotency behavior.
- ERD/data docs: add implemented transaction/idempotency tables.
- Decisions: no new ADR expected unless idempotency or balance strategy diverges from ADR-003.

## Detail Design

- Required: yes
- Link: [PHASE-002-detail-design.md](../phases/PHASE-002-detail-design.md)
- Approval: approved by user instruction to complete the app; design remains available for human review.

## Test Expectations

- Unit: table-driven accounting effects for income, expense, transfer, adjustment, edit, and archive.
- Integration: transaction atomicity, idempotency ledger, ownership isolation, search filters, version increments.
- E2E: add income/expense/transfer/adjustment, edit/archive, and search flows.
- UAT: required; automated browser proof plus human review.
- Manual/platform: not required beyond existing Compose.
- Docs review: API, ERD, validation matrix, context, backlog, changelog.

## Verification Results

- Command: `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/finance -run 'TestApplyAccountingEffect' -count=1`
- Result: pass
- Notes: RED first failed because accounting types and `ApplyAccountingEffect` were missing; GREEN passed after adding transaction type constants, accounting input/effect structs, and income/expense/transfer/adjustment balance effect validation.
- Command: `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/finance -count=1`
- Result: pass
- Notes: Unit-only finance package regression passed after accounting primitives.
- Command: `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/finance -count=1`
- Result: pass
- Notes: Finance package regression passed with DB-backed wallet/category repository tests included.
- Command: `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test -p 1 ./... -count=1`
- Result: pass
- Notes: Backend regression passed across api, migrate, worker, finance, identity, config, db, httpapi, and objectstore packages.
- Command: `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/finance -run 'TestRepositoryCreates.*Transaction|TestRepositoryReplaysDuplicateTransactionIdempotencyKey|TestRepositoryRejectsTransactionForAnotherUsersWallet' -count=1`
- Result: pass
- Notes: RED first failed because `CreateTransaction` and `CreateTransactionInput` were missing; GREEN passed after repository create transaction persistence.
- Command: `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/finance -run 'TestRepository.*Transaction' -count=1`
- Result: pass
- Notes: Repository transaction suite covers income, expense, transfer, adjustment, atomic wallet version/balance updates, duplicate idempotent replay, changed-request idempotency rejection, inactive category rejection, and cross-user wallet rejection.
- Command: `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/finance -count=1`
- Result: pass
- Notes: Unit-only finance package regression passed after transaction repository changes.
- Command: `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/finance -count=1`
- Result: pass
- Notes: Full finance package regression passed with DB-backed transaction repository tests.
- Command: `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test -p 1 ./... -count=1`
- Result: pass
- Notes: Backend regression passed after transaction repository changes.
- Command: `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/finance -run 'TestRepositoryUpdatesTransaction|TestRepositoryArchivesTransaction|TestRepositoryListsTransactions' -count=1`
- Result: pass
- Notes: RED first failed because `UpdateTransaction`, `ArchiveTransaction`, `ListTransactions`, `UpdateTransactionInput`, and `TransactionFilters` were missing; GREEN passed after edit/archive/search repository implementation.
- Command: `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/platform/db -run TestPhase002FinanceTablesAndSeeds -count=1`
- Result: pass
- Notes: Migration test passed after adding transaction delta columns for exact reversal.
- Command: `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/finance -count=1`
- Result: pass
- Notes: Full finance package regression passed after edit/archive/search implementation. A previous parallel DB-backed run failed due shared test schema resets, then passed when rerun sequentially.
- Command: `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test -p 1 ./... -count=1`
- Result: pass
- Notes: Backend regression passed after edit/archive/search implementation.
- Command: `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/platform/httpapi -run 'TestTransactionsAPI|TestCreateTransaction|TestUpdateTransaction|TestArchiveTransaction' -count=1`
- Result: pass
- Notes: RED first failed with 404s because `/api/v1/transactions` routes were missing; GREEN passed after collection and item transaction handlers were added.
- Command: `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/platform/httpapi -count=1`
- Result: pass
- Notes: HTTP package regression passed after transaction routes and `Idempotency-Key` CORS proof.
- Command: `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test -p 1 ./... -count=1`
- Result: pass
- Notes: Backend regression passed after transaction HTTP API implementation.

## Fix/Test Attempt Log

- Same-path failure attempts: 0 / 3
- Total fix/test cycles: 4 / 5
- Blocked by loop guard: no
- Human/design input needed: none before starting approved PHASE-002 plan.

## UAT

- Required: yes
- Reason if not required: not applicable
- Expected behavior: transaction workflows update balances exactly and remain idempotent under retry.
- Verified behavior: domain accounting effect validation, repository create/edit/archive/search flows, and transaction HTTP routes now prove exact income, expense, transfer, and adjustment balance deltas, atomic wallet version/balance updates, category activation validation, duplicate idempotent replay, changed-request idempotency rejection, user-owned wallet enforcement, edit reversal/reapply, archive reversal once, user-scoped filtered listing, CSRF-protected create/edit/archive, authenticated list, idempotency header mapping, and transaction API JSON envelopes. Mobile workflows and UAT remain pending.
- Sign-off: pending.

## Docs Review

- Requirements updated or not needed reason: not needed; accounting behavior follows accepted REQ-F-003/REQ-NF-002/REQ-NF-005 scope.
- Architecture updated or not needed reason: not needed for this slice; repository behavior follows approved finance module boundary.
- API updated or not needed reason: updated `docs/architecture/API.md` with implemented transaction endpoints, filters, body shapes, CSRF, idempotency, and archive semantics.
- ERD/data updated or not needed reason: updated `docs/architecture/ERD.md` with persisted transaction delta columns needed for exact edit/archive reversal.
- ADR created or not needed reason: not needed; no divergence from ADR-003 or approved PHASE-002 design.
- `docs/CONTEXT.md` updated: yes.

## Completion Checklist

- [ ] Implementation complete
- [ ] Tests run and recorded
- [ ] Fix/test loop guard respected
- [ ] Validation matrix updated or explicitly not affected
- [ ] UAT completed or explicitly not required
- [ ] Master docs reconciled
- [ ] Docs review completed
- [ ] ADR created or explicitly not needed
- [ ] `docs/CONTEXT.md` updated
- [ ] `docs/work/BACKLOG.md` updated
- [ ] Trace links updated
