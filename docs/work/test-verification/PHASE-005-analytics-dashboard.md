---
artifact_type: test_verification
id: PHASE-005-analytics-dashboard
status: planned
owner: shared
trace:
  backlog_item: BL-005
  requirement: [REQ-F-006, REQ-F-016, REQ-NF-005, REQ-NF-007]
  phase: PHASE-005
  ticket_or_bug: [TICKET-015, TICKET-016, TICKET-017]
  detail_design: ../phases/PHASE-005-detail-design.md
  validation_matrix: ../VALIDATION_MATRIX.md
  release_notes: ../../releases/CHANGELOG.md
---

# PHASE-005 Analytics and Dashboard Verification

## Status

- Status: planned
- Owner: shared

## Planned Commands

| Command | Expected Coverage |
| --- | --- |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/analytics -count=1` | Aggregate formulas and edge cases. |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test -p 1 ./... -count=1` | Full backend regression. |
| `rtk npm test -- --run` | Dashboard/search/report component and accessibility tests. |
| `rtk npm run build` | Production PWA build. |
| `rtk npm run test:e2e -- analytics-dashboard.spec.ts` | Seeded dashboard/report totals on mobile and desktop. |

## Verification Results

- Command: not run yet
- Result: pending
- Notes: Planned proof only; no implementation evidence.
