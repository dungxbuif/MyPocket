# Native Overview — populated reference state

Status: `in_progress`  
Reference: [Money Lover iPhone survey](../reference-survey/2026-10-16-money-lover-iphone.md), ML-01/ML-02  
Owner: native iOS

## Composition

The Overview route is a data-driven dashboard. It keeps the shared balance
header, wallet block, Money Insider, period report, month summary, and recent
transactions in that order. The populated state is the normal state; empty,
loading, cached, and error states remain explicit branches.

| Region | Shared base | Behavior |
| --- | --- | --- |
| Balance and eye/search/bell | `MPBalanceHeader`, `MPIconButton` | Eye masks the headline only; navigation actions stay owner-scoped. |
| Wallet block | `MPSurfaceCard(.reference)`, `MPSectionHeader`, `MPWalletRow` | Shows up to three current wallets and a management action. Descriptions fall back to the wallet type label. |
| Money Insider | `MPMoneyInsiderCard` | Uses the owner-scoped insider response; no sample values are hard-coded. |
| Week/month report | `MPOverviewReportCard`, `MPSegmentedControl`, `MPChart` | The report switches period locally while preserving the API's month and account-week values. |
| Month metrics | `MPMetricGrid`, `MPMetric` | Uses a compact two-column grid at normal type and reflows at narrow/accessibility sizes. |
| Recent ledger | `MPSurfaceCard(.reference)`, `MPTransactionRow` | Uses the common transaction anatomy and privacy mask. |

## States and proof

- `OverviewViewModel.load()` reads wallets, transactions, categories, profile,
  month summary, and insider data using the authenticated API; the cache is
  owner-scoped and is only a fallback after a permitted transport failure.
- No synthetic transaction or balance is created by the screen. Populated
  screenshots must use the disposable API fixture or a user-approved account.
- `OverviewReferenceTests` covers compact metric labels/formatting and wallet
  subtitle fallback. `ReferenceDisplayTests` covers grid reflow and masking.

Known gap: top-category drilldown, Insider help/category selection, and the
reference app's notification/help disclosures remain separate follow-up states.
