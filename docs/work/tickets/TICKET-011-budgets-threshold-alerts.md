---
artifact_type: ticket
id: TICKET-011
status: ready
owner: human
priority: high
lane: high-risk
trace:
  backlog_item: BL-004
  requirement: [REQ-F-005, REQ-NF-005]
  phase: PHASE-004
  detail_design: ../phases/PHASE-004-detail-design.md
  implementation_plan: ../../superpowers/plans/2026-08-30-phase-004-planning-automation.md
  test_verification: ../test-verification/PHASE-004-planning-automation.md
  validation_matrix: ../VALIDATION_MATRIX.md
  release_notes: ../../releases/CHANGELOG.md
---

# Ticket: TICKET-011 Budgets and Threshold Alerts

## Status

- Status: ready
- Type: feature
- Priority: high
- Phase: PHASE-004

## Context

Users need budget periods and threshold notices that match confirmed spending in `Asia/Ho_Chi_Minh` without changing accounting.

## Acceptance Criteria

- [ ] Budget CRUD supports weekly, monthly, quarterly, yearly, and custom periods.
- [ ] Budget scopes can cover all expense categories or selected categories owned/available to the user.
- [ ] Progress uses confirmed non-archived expenses and excludes transfers, adjustments, and report-excluded transactions.
- [ ] 80% and 100% notices are deduplicated by budget, threshold, and period.
- [ ] Mobile budget screens show progress, period status, empty/error/offline states, and create/edit/archive flows.

## Small Task Exemption

- Small task exemption: no
- Reason: Adds data model, API, worker-visible notices, and user-facing planning behavior.
- Impact checked: API=yes, DB=yes, Security=yes, Runtime=yes, Standards=no

## Test Expectations

- Unit period-boundary and progress formula tests.
- PostgreSQL ownership, threshold, and dedupe integration tests.
- Mobile component/E2E tests for budget CRUD and threshold notice visibility.

## Verification Results

- Command: not run yet
- Result: pending
- Notes: Ready for implementation; no execution evidence claimed.
