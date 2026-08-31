---
artifact_type: phase
id: PHASE-005
status: in_review
owner: human
priority: High
human_fields:
  - goal
  - scope
  - out_of_scope
  - priority
  - success_criteria
ai_fields:
  - risks
  - dependencies
  - verification_plan
  - completion_summary
shared_fields:
  - status
  - trace
  - tickets_and_bugs
trace:
  backlog_items:
    - BL-005
  roadmap: ../ROADMAP.md
  detail_design: PHASE-005-detail-design.md
  requirements:
    - REQ-F-006
    - REQ-F-016
    - REQ-NF-005
    - REQ-NF-007
    - REQ-F-018
  tickets:
    - TICKET-015
    - TICKET-016
    - TICKET-017
    - TICKET-027
  bugs: []
  test_verification: ../test-verification/PHASE-005-analytics-dashboard.md
  validation_matrix: ../VALIDATION_MATRIX.md
  adrs: []
  release_notes: ../../releases/CHANGELOG.md
---

# PHASE-005: Analytics and Dashboard

## Status

- ID: PHASE-005
- Status: in_review
- Owner: human
- Priority: High
- Created: 2026-08-24
- Updated: 2026-08-30

## Trace Links

- Backlog: [BACKLOG.md](../BACKLOG.md)
- Roadmap: [ROADMAP.md](../ROADMAP.md)
- Detail design: [PHASE-005-detail-design.md](PHASE-005-detail-design.md) — approved 2026-08-30; [asset portfolio extension](PHASE-005-asset-portfolio-detail-design.md) — approved 2026-08-31
- Requirements: [REQUIREMENTS.md](../../requirements/REQUIREMENTS.md) — REQ-F-006, REQ-F-016, REQ-F-018, REQ-NF-005, REQ-NF-007
- Test verification: [PHASE-005-analytics-dashboard.md](../test-verification/PHASE-005-analytics-dashboard.md) — planned proof target
- Validation matrix: [VALIDATION_MATRIX.md](../VALIDATION_MATRIX.md)
- ADRs: [decisions](../../decisions/README.md)
- Release notes: [CHANGELOG.md](../../releases/CHANGELOG.md)

## Goal

Deliver the installable MoneyLover-inspired PWA navigation, dashboard, search, wallet views, and trustworthy analytics.

## Scope

- Five-destination shell: overview, transactions, central quick add, budgets, and account.
- Net worth, included wallets, recent transactions, balance privacy, search, and wallet selector.
- Net income, category/subcategory donut, daily bars and average, period comparison, and three-month cumulative baseline.
- Responsive installed-PWA behavior, Vietnamese copy, VND formatting, loading/empty/error/offline states.
- Draft extension after TICKET-017: separate asset portfolio valuation for gold, stocks, crypto, foreign currency, and other assets under TICKET-027.

## Out Of Scope

- AI chat and receipt extraction.
- New finance or planning rules.
- Custom 50/30/20 or six-jar analytics beyond the approved SRS.

## Tickets And Bugs

| ID | Type | Title | Status | Link |
| --- | --- | --- | --- | --- |
| TICKET-015 | Ticket | PWA navigation, search, and wallet views | in_review | [TICKET-015](../tickets/TICKET-015-pwa-navigation-search-wallet-views.md) |
| TICKET-016 | Ticket | Overview and net-worth dashboard | in_review | [TICKET-016](../tickets/TICKET-016-overview-net-worth-dashboard.md) |
| TICKET-017 | Ticket | Analytics reports and cumulative trends | in_review | [TICKET-017](../tickets/TICKET-017-analytics-reports-cumulative-trends.md) |
| TICKET-027 | Ticket | Asset portfolio and market valuation | in_progress | [TICKET-027](../tickets/TICKET-027-asset-portfolio-valuation.md) |

## Dependencies

- PHASE-002 confirmed finance data.
- PHASE-003 offline query cache.
- PHASE-004 budget and notification summaries.

## Risks

- Period boundary or exclusion mistakes can make reports disagree with transactions.
- Charts can obscure zero baselines, missing history, or offline stale data.
- Mobile PWA behavior varies by browser.

## Success Criteria

- Dashboard totals match authoritative confirmed transactions and included wallets.
- All report formulas handle zero baselines and excluded transactions explicitly.
- Wallet and period filtering produce consistent detail totals and chart aggregates.
- Responsive installed PWA passes Vietnamese/VND/mobile UAT and exposes stale/offline states clearly.

## Verification Plan

- Go unit tests for every aggregate and edge condition.
- PostgreSQL integration tests for user, wallet, category, report-exclusion, and timezone filters.
- React chart/component tests for values, accessibility labels, empty/error/offline states, and privacy toggle.
- E2E/UAT comparing seeded transactions with dashboard and report outputs.

## Gate Checklist

- [x] Phase links approved requirements
- [x] Tickets have stable planned IDs and bounded titles
- [x] Risks and dependencies are recorded
- [x] Verification plan is defined
- [x] Release/changelog need is linked
- [x] Detail design is approved
- [x] Ticket artifacts and detailed implementation plan are created
- [x] Phase status is promoted to ready after plan review

## Completion Summary

TICKET-015 through TICKET-017 are in review. Search, wallet detail, dashboard aggregates, normalized reports, category roll-up, daily/cumulative series, comparison baseline handling, privacy masking, and mobile UI have automated package/component/build proof. TICKET-027 is in progress as the extension sequenced immediately after TICKET-017; Wave 1 backend schema/domain/repository proof passed, while API, sync/offline, provider worker, UI/dashboard, E2E, and UAT remain pending. PostgreSQL integration and desktop UAT remain for the earlier implemented scope.
