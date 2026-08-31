---
artifact_type: ticket
id: TICKET-012
status: in_review
owner: human
priority: high
lane: high-risk
trace:
  backlog_item: BL-004
  requirement: [REQ-F-005, REQ-NF-002]
  phase: PHASE-004
  detail_design: ../phases/PHASE-004-detail-design.md
  implementation_plan: ../../superpowers/plans/2026-08-30-phase-004-planning-automation.md
  test_verification: ../test-verification/PHASE-004-planning-automation.md
  validation_matrix: ../VALIDATION_MATRIX.md
  release_notes: ../../releases/CHANGELOG.md
---

# Ticket: TICKET-012 Events, Debts, and Repayments

## Status

- Status: in_review
- Type: feature
- Priority: high
- Phase: PHASE-004

## Context

Users need trip/event grouping and borrow/lend tracking while wallet balances remain governed by confirmed finance transactions.

## Acceptance Criteria

- [x] Events can be created, edited, archived, and linked to owned transactions without changing wallet accounting.
- [x] Event totals use linked confirmed transactions and preserve report exclusion rules.
- [x] Obligations support borrowed/lent direction, principal, counterparty, due date, notes, and archive.
- [x] Repayments link to confirmed owned transactions and cannot exceed remaining principal without explicit adjustment.
- [x] Mobile event/debt screens provide Vietnamese copy, VND formatting, and recovery states.

## Small Task Exemption

- Small task exemption: no
- Reason: Adds planning data, finance links, and user-facing repayment semantics.
- Impact checked: API=yes, DB=yes, Security=yes, Runtime=no, Standards=no

## Test Expectations

- Unit tests for debt remaining amount and event aggregation.
- Integration tests for ownership, repayment linkage, and overpayment rejection.
- E2E tests for event creation, debt creation, repayment linking, and archive.

## Verification Results

- Command: `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/planning -run 'Event|Obligation' -count=1` from `backend/`
- Result: passed 2026-08-31
- Notes: Covers linked event totals excluding report-excluded transactions, owned transaction enforcement, obligation repayment totals, and overpayment rejection.
- Command: `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/platform/httpapi -run 'Planning|Budgets|Events|Obligations' -count=1` from `backend/`
- Result: passed 2026-08-31
- Notes: Covers authenticated event/obligation list/create/link API scoping.
- Command: `rtk npm test -- --run` from `frontend/`
- Result: passed 2026-08-31, 4 files / 30 tests
- Notes: Covers mobile event/debt rendering and event create flow with existing app/offline regressions.
- Command: `rtk npm run build` from `frontend/`
- Result: passed 2026-08-31
- Notes: Production PWA build includes planning client and screens.
- Command: `rtk npm run test:e2e -- planning-automation.spec.ts` from `frontend/`
- Result: passed 2026-08-31, 2 mobile tests
- Notes: Covers live API budget CRUD plus event creation and debt creation with linked transactions.
- Command: `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test -p 1 ./... -count=1` from `backend/`
- Result: passed 2026-08-31
- Notes: Full backend regression after event/obligation migration, repository, and API wiring.
