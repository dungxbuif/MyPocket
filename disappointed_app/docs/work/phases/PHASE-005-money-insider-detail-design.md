---
artifact_type: detail_design
id: PHASE-005-MONEY-INSIDER-DETAIL-DESIGN
status: approved
owner: shared
approval: approved
approved_on: 2026-08-31
trace:
  backlog_item: BL-005
  phase: PHASE-005
  requirements: [REQ-F-006, REQ-F-016, REQ-NF-005, REQ-NF-007]
  tickets: [TICKET-017]
  validation_matrix: ../VALIDATION_MATRIX.md
  source_research: https://moneylover.zendesk.com/hc/en-us/articles/35757921060121-Money-Insider-Definition-usage-and-purchase-instructions
  master_docs_touched: [../../architecture/API.md, ../../architecture/ARCHITECTURE.md]
---

# DETAIL DESIGN: Money Insider Reporting

## 1. Problem and Outcome

The current Home `Money Insider` card contains fixed category, amount, daily average, percentage, and subscription copy. It can misrepresent a new user's finances as real data.

Replace it with an authenticated, user-owned report derived from confirmed transactions. The Home card provides one useful insight; a dedicated detail view supports wallet, category, period, chart, budget, and transaction exploration.

- In scope: Home insight, weekly/monthly detail report, category/wallet filters, six-period comparison, income ratio, projection, budget context, expense statistics, and top expenses.
- Out of scope: subscriptions, trials, payments, premium entitlements, AI-generated advice, and provider data.
- Source behavior: Money Lover's official Money Insider documentation, reviewed 2026-08-31.

## 2. Product Decisions

| Decision | MyPocket behavior |
| --- | --- |
| Home category | Expense category with the highest transaction count in the active month; tie-break by total amount, then stable category ID |
| Parent categories | Leaf expenses roll up to the parent category, matching existing category reports |
| Current average | Category spending divided by elapsed calendar days for the active period; completed periods use full period days |
| Comparison | Current daily average vs equal previous period; zero previous value returns `not_comparable` |
| Six-period chart | Selected week/month plus the five immediately preceding periods |
| Spending/income ratio | `expense / income * 100`; absent when income is zero |
| Projection | `spent / elapsed_days * period_days` only for the current active period with transactions; otherwise label the value `Đã chi` |
| Budget | Reuse applicable active category budget; do not invent a limit when none exists |
| Top categories | Three categories with highest transaction counts; `Xem thêm` opens the full category selector |
| Top expenses | Five highest confirmed expense transactions in the selected category and period |
| Monetization | All Money Insider functionality is unlocked; MyPocket has no trial, registration, subscription, premium entitlement, or payment gate |

All calculations use integer VND on the server. Percentages are presentation values derived from server totals. Transfers, adjustments, archived transactions, and `excluded_from_reports=true` are excluded.

## 3. API Contract

Add one authenticated endpoint:

`GET /api/v1/reports/insider?period=monthly|weekly&anchor=YYYY-MM-DD&wallet_id=&category_id=`

Response shape:

```text
report
  selected_category: id, name, transaction_count
  period: kind, start, end, elapsed_days, total_days, active
  metrics: spent_vnd, average_daily_vnd, prior_average_daily_vnd,
           change_percent, not_comparable, income_vnd,
           spending_income_percent, projected_vnd, projection_available
  budget: nullable id, name, amount_vnd, spent_vnd, remaining_vnd, percent
  periods[6]: start, end, label, spent_vnd, income_vnd,
              spending_income_percent nullable, budget_vnd nullable, over_budget
  top_categories[3]: id, name, transaction_count, spent_vnd
  top_expenses[5]: transaction id, note, amount_vnd, occurred_at
  generated_at, timezone, data_version
```

Validation:

- `period` defaults to `monthly`; only `monthly` and `weekly` are accepted.
- `anchor` defaults to today in `Asia/Ho_Chi_Minh`.
- Wallet/category IDs must belong to the authenticated user.
- Unknown or inaccessible IDs return safe `404`; invalid dates/periods return `400`.
- Empty data returns zero metrics and empty arrays, never sample values.

## 4. Backend Design

Extend `backend/internal/analytics` with an Insider query/service boundary. Keep SQL bounded and user-scoped.

1. Normalize the selected period and five prior periods in Ho Chi Minh time.
2. Aggregate confirmed expenses by rolled-up category and transaction count.
3. Resolve the Home category using the deterministic selection rule.
4. Aggregate six period totals and income in one bounded query where practical.
5. Resolve the applicable category budget from planning data without duplicating budget formulas.
6. Query the top five expenses for the selected category/period.
7. Assemble comparison, ratio, projection, and empty-state semantics in pure functions.

No schema migration is required for the first slice. Add indexes only if PostgreSQL query plans show the existing user/date/category indexes are insufficient.

## 5. Frontend Design

### Home Card

- Replace every fixed value in the current card with `/reports/insider` data.
- Loading uses stable skeleton dimensions; error shows a compact retry action; offline shows the last successful report with a stale timestamp.
- Empty state: `Chưa đủ dữ liệu chi tiêu tháng này` and a transaction quick-add action.
- Display selected category, total spent, average/day, and previous-period comparison.
- Remove `Dùng thử miễn phí` and `Đăng ký ngay`.
- The refresh icon reloads the report; selecting the card opens the detail view.

### Detail View

- Wallet menu, top-three category tabs plus `Xem thêm`, and weekly/monthly segmented control.
- Column chart: six periods, spending bars, optional budget line, explicit over-budget state.
- Ratio view: spending-to-income percentage per period; hide ratio when income is zero.
- Numeric summary adjacent to every chart for accessibility.
- Sections for projection/spent, budget, transaction count, average per transaction, average/day, previous-period comparison, and top five expenses.
- Desktop uses a centered report surface; mobile remains a full-height app view.

Cache successful responses by authenticated user plus normalized filters. Logout clears Insider cache with the existing offline store.

## 6. Implementation Slices

1. **Remove misleading Home content:** delete fixed metrics and subscription CTAs; add loading/error/empty states.
2. **Formula and repository layer:** period boundaries, selection/tie-break, averages, comparisons, ratios, projection, six-period series, budget mapping, and top expenses.
3. **Authenticated API:** route, validation, ownership, envelopes, and safe errors.
4. **Home integration:** real data, retry/refresh, privacy masking, offline cache, and stale state.
5. **Detail report UI:** filters, charts, summaries, top categories, and top expenses.
6. **Performance and reconciliation:** PostgreSQL plans, indexes only if needed, docs/API/validation/changelog/context updates.

## 7. Verification

- Unit: month/week/leap boundaries, current vs completed periods, zero prior, zero income, tie-breaks, parent roll-up, projection, and budget/no-budget cases.
- Integration: user isolation, filters, exclusions, six-period series, applicable budget, and top-five ordering against seeded PostgreSQL.
- Component: no hardcoded finance values, loading/error/empty/offline states, privacy masking, long category names, and responsive charts.
- E2E: create known income/expenses/budget, verify Home insight and detail totals on mobile and desktop, reload offline, then confirm stale labeling.
- UAT: compare visible values with the seeded transaction ledger and sign off on Vietnamese labels.

## 8. Risks and Known Unknowns

- `Most frequent` means transaction count, not highest spend; this follows the official Home behavior and must be explicit in tests.
- Category budgets can target several categories or all categories. The first slice should show a budget only when one unambiguously applies to the selected rolled-up category; ambiguous budget allocation requires a later product decision.
- The official product includes paid access. The owner explicitly decided on 2026-08-31 that this personal application unlocks all functionality, so no Money Insider capability may depend on payment or entitlement state.
