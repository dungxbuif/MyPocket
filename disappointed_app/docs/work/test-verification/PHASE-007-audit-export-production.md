---
artifact_type: test_verification
id: PHASE-007-audit-export-production
status: planned
owner: shared
trace:
  backlog_item: BL-007
  requirement: [REQ-F-013, REQ-F-014, REQ-F-015, REQ-NF-004, REQ-NF-006, REQ-NF-007]
  phase: PHASE-007
  ticket_or_bug: [TICKET-022, TICKET-023, TICKET-024, TICKET-025]
  detail_design: ../phases/PHASE-007-detail-design.md
  validation_matrix: ../VALIDATION_MATRIX.md
  release_notes: ../../releases/CHANGELOG.md
---

# PHASE-007 Audit, Export, Account Lifecycle, and Production Operations Verification

## Status

- Status: planned
- Owner: shared

## Planned Commands

| Command | Expected Coverage |
| --- | --- |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/audit ./internal/export ./internal/lifecycle -count=1` | Audit, export, reset/delete, retention, env validation. |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test -p 1 ./... -count=1` | Full backend regression. |
| `rtk npm test -- --run` | Audit viewer, account export/reset/delete component tests. |
| `rtk npm run build` | Production PWA build. |
| `rtk npm run test:e2e -- audit-export-production.spec.ts` | Viewer allow/deny, export, reset/delete confirmation, and release smoke UI flows. |
| `rtk ./scripts/smoke-platform.sh` | Production-like Compose health, S3, and runtime smoke. |

## Verification Results

- Command: not run yet
- Result: pending
- Notes: Planned proof only; no implementation evidence.
