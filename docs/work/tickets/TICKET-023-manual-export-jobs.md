---
artifact_type: ticket
id: TICKET-023
status: ready
owner: human
priority: high
lane: high-risk
trace:
  backlog_item: BL-007
  requirement: [REQ-F-013, REQ-NF-004, REQ-NF-007]
  phase: PHASE-007
  detail_design: ../phases/PHASE-007-detail-design.md
  implementation_plan: ../../superpowers/plans/2026-08-30-phase-007-production.md
  test_verification: ../test-verification/PHASE-007-audit-export-production.md
  validation_matrix: ../VALIDATION_MATRIX.md
  release_notes: ../../releases/CHANGELOG.md
---

# Ticket: TICKET-023 Manual Export Jobs

## Status

- Status: ready
- Type: feature
- Priority: high
- Phase: PHASE-007

## Context

Users need portable CSV/Google Sheets-compatible exports scoped exactly to their own data.

## Acceptance Criteria

- [ ] Export request creates an immutable user-scoped snapshot job with selected datasets/date range.
- [ ] Worker writes UTF-8 CSV with stable headers, ISO timestamps, integer VND, and no secrets.
- [ ] Export download uses short-lived private S3 presigned URLs.
- [ ] Export progress/failure/retry states appear in the account screen.
- [ ] Cross-user export access is forbidden and covered by tests.

## Small Task Exemption

- Small task exemption: no
- Reason: Adds export APIs, jobs, private object output, and security-sensitive data access.
- Impact checked: API=yes, DB=yes, Security=yes, Runtime=yes, Standards=no

## Test Expectations

- Unit CSV formatting tests.
- Integration export isolation/job/idempotency tests.
- E2E export request, progress, and download-state tests.

## Verification Results

- Command: not run yet
- Result: pending
- Notes: Ready for implementation; no execution evidence claimed.
