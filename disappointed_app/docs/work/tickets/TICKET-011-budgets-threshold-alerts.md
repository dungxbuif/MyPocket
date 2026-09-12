---
artifact_type: ticket
id: TICKET-011
status: in_review
owner: human
priority: high
lane: high-risk
trace:
  backlog_item: BL-004
  requirement: [REQ-F-005, REQ-NF-005]
  phase: PHASE-004
  detail_design: ../phases/PHASE-004-detail-design.md
  implementation_plan: ../../superpowers/plans/2026-08-30-phase-004-planning-automation.md
  test_verification: ../test-verification/PHASE-004-planning-automation.md
  validation_matrix: ../VALIDATION_MATRIX.md
  release_notes: ../../releases/CHANGELOG.md
---

# Ticket: TICKET-011 Budgets and Threshold Alerts

## Status

- Status: in_review
- Type: feature
- Priority: high
- Phase: PHASE-004

## Context

Users need budget periods and threshold notices that match confirmed spending in `Asia/Ho_Chi_Minh` without changing accounting.

## Acceptance Criteria

- [x] Budget CRUD supports weekly, monthly, quarterly, yearly, and custom periods.
- [x] Budget scopes can cover all expense categories or selected categories owned/available to the user.
- [x] Progress uses confirmed non-archived expenses and excludes transfers, adjustments, and report-excluded transactions.
- [x] 80% and 100% notices are deduplicated by budget, threshold, and period.
- [x] Mobile budget screens show progress, period status, empty/error/offline states, and create/edit/archive flows.

## Small Task Exemption

- Small task exemption: no
- Reason: Adds data model, API, worker-visible notices, and user-facing planning behavior.
- Impact checked: API=yes, DB=yes, Security=yes, Runtime=yes, Standards=no

## Test Expectations

- Unit period-boundary and progress formula tests.
- PostgreSQL ownership, threshold, and dedupe integration tests.
- Mobile component/E2E tests for budget CRUD and threshold notice visibility.

## Verification Results

- Command: `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/planning ./internal/platform/httpapi -count=1` from `backend/`
- Result: passed 2026-08-31
- Command: `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test -p 1 ./... -count=1` from `backend/`
- Result: passed 2026-08-31
- Command: `rtk npm test -- --run` from `frontend/`
- Result: passed 2026-08-31, 4 files / 29 tests
- Command: `rtk npm run build` from `frontend/`
- Result: passed 2026-08-31
- Command: `rtk npm run test:e2e -- planning-automation.spec.ts` from `frontend/`
- Result: passed 2026-08-31, 1 mobile test
- Notes: Budget CRUD, category scope validation, Ho Chi Minh period boundaries, progress formula exclusions, 80/100 threshold dedupe, and mobile budget create/edit/archive are implemented and covered by automated proof. UAT remains pending.
