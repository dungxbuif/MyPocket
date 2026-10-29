---
artifact_type: test_verification
id: PHASE-005-analytics-dashboard
status: in_review
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

- Status: in_review
- Owner: shared

## Planned Commands

| Command | Expected Coverage |
| --- | --- |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/analytics -count=1` | Aggregate formulas and edge cases. |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test -p 1 ./... -count=1` | Full backend regression. |
| `rtk npm test -- --run` | Dashboard/search/report component and accessibility tests. |
| `rtk npm run build` | Production PWA build. |
| `rtk npm run test:e2e -- analytics-dashboard.spec.ts` | Seeded dashboard/report totals on mobile and desktop. |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./...`; `rtk npm test -- --run`; `rtk npm run build` | Pass 2026-08-31 after replacing hardcoded Money Insider Home metrics with authenticated category-frequency, elapsed-day average, and prior-period comparison. |

## Verification Results

- Command: `GOCACHE=/private/tmp/mypocket-go-cache go test ./...` (backend), `npm test -- --run`, `npm run build` (frontend)
- Result: pass for package/component/build proof; PostgreSQL integration and E2E need local services
- Notes: Analytics package, authenticated search/dashboard/report handlers, privacy masking, and report preview UI are implemented through TICKET-017.
