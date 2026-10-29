---
artifact_type: ticket
id: TICKET-013
status: in_review
owner: human
priority: high
lane: high-risk
trace:
  backlog_item: BL-004
  requirement: [REQ-F-005, REQ-NF-002, REQ-NF-005]
  phase: PHASE-004
  detail_design: ../phases/PHASE-004-detail-design.md
  implementation_plan: ../../superpowers/plans/2026-08-30-phase-004-planning-automation.md
  test_verification: ../test-verification/PHASE-004-planning-automation.md
  validation_matrix: ../VALIDATION_MATRIX.md
  release_notes: ../../releases/CHANGELOG.md
---

# Ticket: TICKET-013 Recurring Schedules and Worker Occurrences

## Status

- Status: in_review
- Type: feature
- Priority: high
- Phase: PHASE-004

## Context

Recurring payments must produce reviewable drafts exactly once per due occurrence, including across worker retries and restarts.

## Acceptance Criteria

- [x] Recurring schedules support normalized recurrence, timezone, amount, wallet, category, and draft payload fields.
- [x] Worker leases prevent duplicate concurrent processing.
- [x] Each due occurrence creates at most one transaction draft using a deterministic occurrence key.
- [x] Generated drafts never modify wallet balances until explicitly confirmed.
- [x] Mobile schedule UI supports create/archive and draft review entry points; editing an existing schedule is archive-and-create-new for M1.

## Small Task Exemption

- Small task exemption: no
- Reason: Adds worker scheduling, DB idempotency, and draft-generation behavior.
- Impact checked: API=yes, DB=yes, Security=yes, Runtime=yes, Standards=no

## Test Expectations

- Unit recurrence/timezone tests.
- Integration tests for leases, restart recovery, and occurrence idempotency.
- E2E tests for schedule setup and draft confirmation handoff.

## Verification Results

- Command: `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/planning ./internal/worker ./internal/platform/httpapi -run 'Recurring|WorkerLease|Planning|Budgets|Events|Obligations' -count=1` from `backend/`
- Result: passed 2026-08-31
- Notes: Covers normalized schedule creation, worker leases, deterministic draft generation, no-balance-change draft behavior, worker runner lease gate, and authenticated schedule/draft APIs.
- Command: `rtk npm test -- --run` from `frontend/`
- Result: passed 2026-08-31, 4 files / 31 tests
- Notes: Covers schedule creation UI, pending draft review row, and existing planning/offline regressions.
- Command: `rtk npm run build` from `frontend/`
- Result: passed 2026-08-31
- Notes: Production PWA build includes recurring schedule and draft review entry points.
- Command: `rtk npm run test:e2e -- planning-automation.spec.ts` from `frontend/`
- Result: passed 2026-08-31, 2 mobile tests
- Notes: Covers live API budget CRUD, event/debt linking, and recurring schedule setup.
- Command: `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test -p 1 ./... -count=1` from `backend/`
- Result: passed 2026-08-31
- Notes: Full backend regression after recurring migration, repository, API, worker runner, and command wiring.
