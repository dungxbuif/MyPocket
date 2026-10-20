---
artifact_type: ticket
id: TICKET-017
status: in_review
owner: human
priority: high
lane: normal
trace:
  backlog_item: BL-005
  requirement: [REQ-F-006, REQ-NF-005, REQ-NF-007]
  phase: PHASE-005
  detail_design: [../phases/PHASE-005-detail-design.md, ../phases/PHASE-005-money-insider-detail-design.md]
  implementation_plan: ../../superpowers/plans/2026-08-30-phase-005-analytics-dashboard.md
  test_verification: ../test-verification/PHASE-005-analytics-dashboard.md
  validation_matrix: ../VALIDATION_MATRIX.md
  release_notes: ../../releases/CHANGELOG.md
---

# Ticket: TICKET-017 Analytics Reports and Cumulative Trends

## Status

- Status: in_review
- Type: feature
- Priority: high
- Phase: PHASE-005

## Context

Reports must use exact server formulas for income, expenses, category shares, daily averages, comparisons, and cumulative trends.

## Acceptance Criteria

- [x] Cash-flow, category, daily, comparison, and cumulative report APIs share one normalized filter contract.
- [x] Reports exclude archived, transfer, adjustment, and report-excluded transactions exactly as designed.
- [x] Category reports roll leaf categories to parents and expose uncategorized explicitly.
- [x] Zero prior baseline returns `not_comparable`, never infinity or misleading percentages.
- [x] Charts expose accessible labels and adjacent numeric summaries; color is not the only signal.
- [x] Money Insider Home uses authenticated aggregates for the most frequent expense category, elapsed-day average, and prior-month comparison; no sample finance values or subscription CTAs remain.
- [ ] Money Insider detail view exposes six periods, wallet/category filters, projection, applicable budget, and top expenses.

## Small Task Exemption

- Small task exemption: no
- Reason: Adds financial reporting formulas and chart UX that require proof.
- Impact checked: API=yes, DB=no, Security=yes, Runtime=no, Standards=no

## Test Expectations

- Unit formula tests for all edge cases.
- Integration tests for filters, timezone, exclusion, and ownership.
- Component/E2E chart tests with seeded expected totals.

## Verification Results

- Command: `GOCACHE=/private/tmp/mypocket-go-cache go test ./...`; `npm test -- --run`; `npm run build`
- Result: pass for package/component/build proof
- Notes: Server formulas normalize Ho Chi Minh date ranges, zero-fill daily series, roll parent categories, and mark zero comparison baselines; PostgreSQL integration remains environment-gated.
