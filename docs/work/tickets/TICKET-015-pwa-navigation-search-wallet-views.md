---
artifact_type: ticket
id: TICKET-015
status: ready
owner: human
priority: high
lane: normal
trace:
  backlog_item: BL-005
  requirement: [REQ-F-016]
  phase: PHASE-005
  detail_design: ../phases/PHASE-005-detail-design.md
  implementation_plan: ../../superpowers/plans/2026-08-30-phase-005-analytics-dashboard.md
  test_verification: ../test-verification/PHASE-005-analytics-dashboard.md
  validation_matrix: ../VALIDATION_MATRIX.md
  release_notes: ../../releases/CHANGELOG.md
---

# Ticket: TICKET-015 PWA Navigation, Search, and Wallet Views

## Status

- Status: ready
- Type: feature
- Priority: high
- Phase: PHASE-005

## Context

The PWA needs complete mobile navigation, search, wallet detail, and cached read-only offline states beyond the finance editor slice.

## Acceptance Criteria

- [ ] Five-destination shell matches `design/DESIGN.md` with overview, transactions, quick add, budgets, and account.
- [ ] Search returns bounded user-owned wallets, transactions, categories, events, and debts grouped by kind.
- [ ] Wallet detail shows balance, included/excluded status, recent transactions, and filter links.
- [ ] Offline cached query states are visibly stale and read-only for analytics/search results.
- [ ] Mobile and desktop layouts avoid text overlap and preserve Vietnamese/VND formatting.

## Small Task Exemption

- Small task exemption: no
- Reason: Changes major user-facing navigation and query APIs.
- Impact checked: API=yes, DB=no, Security=yes, Runtime=no, Standards=no

## Test Expectations

- Backend search ownership/filter tests.
- Component tests for navigation, wallet view, search states, and offline/stale markers.
- E2E tests on mobile and desktop PWA viewports.

## Verification Results

- Command: not run yet
- Result: pending
- Notes: Ready for implementation; no execution evidence claimed.
