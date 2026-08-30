---
artifact_type: test_verification
id: PHASE-004-planning-automation
status: planned
owner: shared
trace:
  backlog_item: BL-004
  requirement: [REQ-F-005, REQ-F-012, REQ-NF-002, REQ-NF-005]
  phase: PHASE-004
  ticket_or_bug: [TICKET-011, TICKET-012, TICKET-013, TICKET-014]
  detail_design: ../phases/PHASE-004-detail-design.md
  validation_matrix: ../VALIDATION_MATRIX.md
  release_notes: ../../releases/CHANGELOG.md
---

# PHASE-004 Planning and Automation Verification

## Status

- Status: planned
- Owner: shared

## Planned Commands

| Command | Expected Coverage |
| --- | --- |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/planning ./internal/worker -count=1` | Budget periods, debts, recurring schedules, notifications, worker leases. |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test -p 1 ./... -count=1` | Full backend regression. |
| `rtk npm test -- --run` | Mobile planning and inbox component tests. |
| `rtk npm run build` | Production PWA build. |
| `rtk npm run test:e2e -- planning-automation.spec.ts` | Budget, event, debt, recurring draft, inbox, and push-state flows. |

## Verification Results

- Command: not run yet
- Result: pending
- Notes: Planned proof only; no implementation evidence.
