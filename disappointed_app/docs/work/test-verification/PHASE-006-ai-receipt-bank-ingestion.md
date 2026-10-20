---
artifact_type: test_verification
id: PHASE-006-ai-receipt-bank-ingestion
status: planned
owner: shared
trace:
  backlog_item: BL-006
  requirement: [REQ-F-007, REQ-F-008, REQ-F-009, REQ-F-010, REQ-F-011, REQ-NF-004, REQ-NF-008]
  phase: PHASE-006
  ticket_or_bug: [TICKET-018, TICKET-019, TICKET-020, TICKET-021]
  detail_design: ../phases/PHASE-006-detail-design.md
  validation_matrix: ../VALIDATION_MATRIX.md
  adrs: [../../decisions/ADR-004-review-first-ingestion.md]
  release_notes: ../../releases/CHANGELOG.md
---

# PHASE-006 AI, Receipt, and Bank Ingestion Verification

## Status

- Status: planned
- Owner: shared

## Planned Commands

| Command | Expected Coverage |
| --- | --- |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/ingestion ./internal/providers -count=1` | Draft contract, adapters, HMAC, redaction, confirmation safety. |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test -p 1 ./... -count=1` | Full backend regression. |
| `rtk npm test -- --run` | Draft review, receipt, AI chat, and webhook notice component tests. |
| `rtk npm run build` | Production PWA build. |
| `rtk npm run test:e2e -- ingestion.spec.ts` | Text AI, receipt OCR, image chat, webhook review, edit/reject/confirm, and provider failure flows. |

## Verification Results

- Command: not run yet
- Result: pending
- Notes: Planned proof only; no implementation evidence.
