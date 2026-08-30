---
artifact_type: ticket
id: TICKET-016
status: ready
owner: human
priority: high
lane: normal
trace:
  backlog_item: BL-005
  requirement: [REQ-F-006, REQ-NF-005]
  phase: PHASE-005
  detail_design: ../phases/PHASE-005-detail-design.md
  implementation_plan: ../../superpowers/plans/2026-08-30-phase-005-analytics-dashboard.md
  test_verification: ../test-verification/PHASE-005-analytics-dashboard.md
  validation_matrix: ../VALIDATION_MATRIX.md
  release_notes: ../../releases/CHANGELOG.md
---

# Ticket: TICKET-016 Overview and Net-Worth Dashboard

## Status

- Status: ready
- Type: feature
- Priority: high
- Phase: PHASE-005

## Context

Dashboard totals must come from authoritative server aggregates, not frontend recomputation or sample values.

## Acceptance Criteria

- [ ] Net worth sums active included wallets and preserves signed credit/debt balances.
- [ ] Overview shows wallet summaries, recent transactions, planning summary, stale state, and privacy masking.
- [ ] Privacy toggle masks DOM-visible balance text and persists per device.
- [ ] Dashboard API responses include generated timestamp, normalized date range, timezone, and data version.
- [ ] Totals match seeded transaction/wallet fixtures in automated proof.

## Small Task Exemption

- Small task exemption: no
- Reason: Adds financial aggregate APIs and user-facing dashboard behavior.
- Impact checked: API=yes, DB=no, Security=yes, Runtime=no, Standards=no

## Test Expectations

- Go unit/integration tests for net worth and ownership.
- Frontend component tests for privacy, empty/error/offline states, and values.
- E2E comparison against seeded fixtures.

## Verification Results

- Command: not run yet
- Result: pending
- Notes: Ready for implementation; no execution evidence claimed.
