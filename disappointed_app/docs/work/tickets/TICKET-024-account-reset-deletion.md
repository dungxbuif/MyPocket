---
artifact_type: ticket
id: TICKET-024
status: deferred
owner: human
priority: high
lane: high-risk
trace:
  backlog_item: BL-007
  requirement: [REQ-F-014, REQ-NF-004, REQ-NF-007]
  phase: PHASE-007
  detail_design: ../phases/PHASE-007-detail-design.md
  implementation_plan: ../../superpowers/plans/2026-08-30-phase-007-production.md
  test_verification: ../test-verification/PHASE-007-audit-export-production.md
  validation_matrix: ../VALIDATION_MATRIX.md
  release_notes: ../../releases/CHANGELOG.md
---

# Ticket: TICKET-024 Account Reset and Deletion

## Status

- Status: deferred
- Type: feature
- Priority: high
- Phase: PHASE-007

## Context

Reset/delete are destructive operations and must require precise confirmation, exact scoping, resumable jobs, and lifecycle evidence.

## Acceptance Criteria

- [ ] Reset/delete commands require recent auth, typed confirmation, CSRF, idempotency, and affected-count preview.
- [ ] Reset removes finance/planning/provider data while preserving identity and lifecycle audit evidence.
- [ ] Delete disables access immediately, queues data/S3 deletion, and leaves only documented tombstone evidence.
- [ ] Jobs are resumable, user-scoped, and fail closed on ambiguous DB rows or object keys.
- [ ] Account screen shows cancellation, confirmation, progress, terminal success, and failure recovery states.

## Small Task Exemption

- Small task exemption: no
- Reason: Adds destructive data lifecycle operations and S3 deletion behavior.
- Impact checked: API=yes, DB=yes, Security=yes, Runtime=yes, Standards=no

## Test Expectations

- Unit deletion-plan tests.
- Integration exact-scope DB/S3 cleanup and idempotent resume tests.
- E2E reset/delete confirmation and cancellation tests.

## Verification Results

- Command: not run yet
- Result: pending
- Notes: Deferred to M2/post-M1 by human scope decision on 2026-08-31; no execution evidence claimed.
