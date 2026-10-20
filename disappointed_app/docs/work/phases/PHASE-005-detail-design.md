---
artifact_type: detail_design
id: PHASE-005-DETAIL-DESIGN
status: approved
owner: shared
approval: approved
approved_on: 2026-08-30
trace:
  backlog_item: BL-005
  phase: PHASE-005
  requirements: [REQ-F-006, REQ-F-016, REQ-NF-005, REQ-NF-007]
  tickets: [TICKET-015, TICKET-016, TICKET-017]
  validation_matrix: ../VALIDATION_MATRIX.md
  master_docs_touched: [../../architecture/API.md, ../../architecture/ARCHITECTURE.md]
---

# DETAIL DESIGN: PHASE-005 Analytics and Dashboard

## 1. Context and Scope

This phase replaces sample dashboard values with trustworthy server aggregates and completes the mobile PWA navigation, wallet detail, search, and report experience.

- In scope: overview, net worth, recent transactions, global search, wallet views, net income, category donut, daily bars/average, period comparison, three-month cumulative trend, privacy, stale/offline states.
- Out of scope: new accounting rules, custom budgeting systems, AI, and provider ingestion.
- Approval: user-delegated decisions on 2026-08-30. Small task exemption: no.

## 2. Reporting Rules

| Metric | Formula and exclusions |
| --- | --- |
| Net worth | Sum active wallets with `include_in_total=true`; credit balances retain their signed value |
| Income/expense | Confirmed, non-archived transactions in selected Ho Chi Minh date range; exclude transfers, adjustments, and `excluded_from_reports` |
| Net income | Income minus expense |
| Category share | Expense grouped by leaf category with parent roll-up; uncategorized is explicit |
| Daily average | Total expense divided by elapsed calendar days for active period, or full days for completed period |
| Comparison | Current period vs equal prior period; zero prior baseline returns `not_comparable`, never infinity |
| Cumulative trend | Daily cumulative net income over current and previous two months with zero-filled missing days |

All formulas live in `backend/internal/analytics`; frontend charts display returned values and do not independently recompute financial totals.

## 3. API and Cache

- `GET /api/v1/dashboard?wallet_id=&from=&to=` returns net worth, wallet summaries, recent transactions, and planning summary.
- `GET /api/v1/reports/cash-flow`, `/categories`, `/daily`, `/comparison`, and `/cumulative` use the same normalized filter contract.
- `GET /api/v1/search?q=&wallet_id=&from=&to=` returns bounded user-owned transactions, wallets, categories, events, and debts grouped by kind.
- Responses include `generated_at`, normalized timezone/date range, and `data_version` for offline staleness display.
- IndexedDB caches successful query envelopes by user/filter key; cached reports are read-only offline.

## 4. Mobile UI Contract

- Preserve the five destinations and raised quick-add action in `design/DESIGN.md`.
- Overview shows privacy-controlled net worth, included wallets, recent transactions, budget progress, and concise report previews.
- Transactions screen provides search, wallet/type/category/date filters, edit/archive entry points, loading/empty/error/offline states, and virtualized pagination if lists exceed 100 rows.
- Reports use accessible SVG/canvas charts with adjacent numeric summaries and screen-reader labels; color is never the only indicator.
- Desktop uses the same information architecture in a constrained two-column work surface; mobile remains primary.

## 5. Reliability, Privacy, and Performance

- Aggregate queries always filter authenticated user ownership and use integer VND until formatting.
- Privacy toggle masks balances in DOM-visible text and is stored per device.
- Query limits and indexes cover user/date/wallet/category predicates; target p95 under 500 ms for one year of typical data.
- Cached results visibly show offline/stale timestamps; no cached aggregate is presented as current after local unsynced mutations without a pending badge.

## 6. Verification and Reconciliation

- Unit: every formula, zero baseline, leap/month boundary, exclusion, parent roll-up, and signed wallet edge.
- Integration: ownership and filter consistency against seeded PostgreSQL data.
- Component/accessibility: chart values, labels, privacy masking, responsive text, and all empty/error/offline states.
- E2E/UAT: compare known transactions to each dashboard/report total on mobile and desktop installed-PWA viewports.
- Reconcile API, architecture, validation matrix, design contract if components evolve, context, backlog, and changelog.

