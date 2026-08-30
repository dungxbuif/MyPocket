---
artifact_type: ticket
id: TICKET-012
status: ready
owner: human
priority: high
lane: high-risk
trace:
  backlog_item: BL-004
  requirement: [REQ-F-005, REQ-NF-002]
  phase: PHASE-004
  detail_design: ../phases/PHASE-004-detail-design.md
  implementation_plan: ../../superpowers/plans/2026-08-30-phase-004-planning-automation.md
  test_verification: ../test-verification/PHASE-004-planning-automation.md
  validation_matrix: ../VALIDATION_MATRIX.md
  release_notes: ../../releases/CHANGELOG.md
---

# Ticket: TICKET-012 Events, Debts, and Repayments

## Status

- Status: ready
- Type: feature
- Priority: high
- Phase: PHASE-004

## Context

Users need trip/event grouping and borrow/lend tracking while wallet balances remain governed by confirmed finance transactions.

## Acceptance Criteria

- [ ] Events can be created, edited, archived, and linked to owned transactions without changing wallet accounting.
- [ ] Event totals use linked confirmed transactions and preserve report exclusion rules.
- [ ] Obligations support borrowed/lent direction, principal, counterparty, due date, notes, and archive.
- [ ] Repayments link to confirmed owned transactions and cannot exceed remaining principal without explicit adjustment.
- [ ] Mobile event/debt screens provide Vietnamese copy, VND formatting, and recovery states.

## Small Task Exemption

- Small task exemption: no
- Reason: Adds planning data, finance links, and user-facing repayment semantics.
- Impact checked: API=yes, DB=yes, Security=yes, Runtime=no, Standards=no

## Test Expectations

- Unit tests for debt remaining amount and event aggregation.
- Integration tests for ownership, repayment linkage, and overpayment rejection.
- E2E tests for event creation, debt creation, repayment linking, and archive.

## Verification Results

- Command: not run yet
- Result: pending
- Notes: Ready for implementation; no execution evidence claimed.
