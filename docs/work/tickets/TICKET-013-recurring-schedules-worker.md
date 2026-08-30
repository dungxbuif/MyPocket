---
artifact_type: ticket
id: TICKET-013
status: ready
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

- Status: ready
- Type: feature
- Priority: high
- Phase: PHASE-004

## Context

Recurring payments must produce reviewable drafts exactly once per due occurrence, including across worker retries and restarts.

## Acceptance Criteria

- [ ] Recurring schedules support normalized recurrence, timezone, amount, wallet, category, and draft payload fields.
- [ ] Worker leases prevent duplicate concurrent processing.
- [ ] Each due occurrence creates at most one transaction draft using a deterministic occurrence key.
- [ ] Generated drafts never modify wallet balances until explicitly confirmed.
- [ ] Mobile schedule UI supports create/edit/archive and draft review entry points.

## Small Task Exemption

- Small task exemption: no
- Reason: Adds worker scheduling, DB idempotency, and draft-generation behavior.
- Impact checked: API=yes, DB=yes, Security=yes, Runtime=yes, Standards=no

## Test Expectations

- Unit recurrence/timezone tests.
- Integration tests for leases, restart recovery, and occurrence idempotency.
- E2E tests for schedule setup and draft confirmation handoff.

## Verification Results

- Command: not run yet
- Result: pending
- Notes: Ready for implementation; no execution evidence claimed.
