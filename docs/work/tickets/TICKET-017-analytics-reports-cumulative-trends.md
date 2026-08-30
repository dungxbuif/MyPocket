---
artifact_type: ticket
id: TICKET-017
status: ready
owner: human
priority: high
lane: normal
trace:
  backlog_item: BL-005
  requirement: [REQ-F-006, REQ-NF-005, REQ-NF-007]
  phase: PHASE-005
  detail_design: ../phases/PHASE-005-detail-design.md
  implementation_plan: ../../superpowers/plans/2026-08-30-phase-005-analytics-dashboard.md
  test_verification: ../test-verification/PHASE-005-analytics-dashboard.md
  validation_matrix: ../VALIDATION_MATRIX.md
  release_notes: ../../releases/CHANGELOG.md
---

# Ticket: TICKET-017 Analytics Reports and Cumulative Trends

## Status

- Status: ready
- Type: feature
- Priority: high
- Phase: PHASE-005

## Context

Reports must use exact server formulas for income, expenses, category shares, daily averages, comparisons, and cumulative trends.

## Acceptance Criteria

- [ ] Cash-flow, category, daily, comparison, and cumulative report APIs share one normalized filter contract.
- [ ] Reports exclude archived, transfer, adjustment, and report-excluded transactions exactly as designed.
- [ ] Category reports roll leaf categories to parents and expose uncategorized explicitly.
- [ ] Zero prior baseline returns `not_comparable`, never infinity or misleading percentages.
- [ ] Charts expose accessible labels and adjacent numeric summaries; color is not the only signal.

## Small Task Exemption

- Small task exemption: no
- Reason: Adds financial reporting formulas and chart UX that require proof.
- Impact checked: API=yes, DB=no, Security=yes, Runtime=no, Standards=no

## Test Expectations

- Unit formula tests for all edge cases.
- Integration tests for filters, timezone, exclusion, and ownership.
- Component/E2E chart tests with seeded expected totals.

## Verification Results

- Command: not run yet
- Result: pending
- Notes: Ready for implementation; no execution evidence claimed.
