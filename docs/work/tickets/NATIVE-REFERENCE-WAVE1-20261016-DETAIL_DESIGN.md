---
artifact_type: detail_design
id: NATIVE-REFERENCE-WAVE1-20261016
status: in_progress
approval: owner requested full native UI reproduction and delegated execution
---

# Physical reference UI — first implementation wave

## Intent and evidence

Owner explicitly selected the installed Money Lover UI as reference and asked
to inspect individual subfunctions, then enabled use of built-in sample data.
The [survey ledger](../../design/reference-survey/2026-10-16-money-lover-iphone.md)
records actually visited screens and screenshots outside Git. Preserve existing
MyPocket finance/API/auth/offline/AI/feedback features and owned data semantics.
This wave implements ledger layout and reusable display bases; subsequent
reference routes remain queued, not implicitly accepted.

Observed evidence: physical 430×932pt viewport; ledger header has wallet scope,
balance, search/overflow, period tabs, incoming/outgoing/net summary, report
action, then one white rounded group per day with day number/date/net header.
Rows have category glyph/title/note and a right-aligned amount. Reports and
budget detail use compact label/value metrics and clear grouped progress.
The sample account's day5 period preference is not MyPocket business logic.

Parent visually read `ledger-sample.png`, `reports-period-sample.png` and
`budget-sample-detail.png`. A current new Savings Report capture reveals metrics
stacked vertically and consuming excessive height; use a shared adaptive metric
grid rather than bespoke feature rows.

## Component contracts and ownership

| Work | Owned source | Required reuse |
| --- | --- | --- |
| Display base fidelity + compact report metrics | DesignSystem/MPSurfaces.swift, MPFinanceCards.swift; new base tests; new SavingsReportsView/BudgetReportView composition only | MPText, MPMetric(compact), MPProgressBar, existing tokens; implement reference surface radius variant and shared adaptive grid |
| Ledger composition | Features/Transactions/TransactionsView.swift; transaction tests only | Existing MPWalletScopeSelector, period controls, MPTransactionRow, MPSurfaceCard(.reference), day heading and overflow bases |
| Integration | Parent | Existing shell/feedback/keyboard contracts; signed fixtures and screenshots |

Typography, primary app palette and money semantics remain centralized in
existing tokens. No screen-local colors/fonts/corners, cloned base wrappers or
new raw controls. Reference variant uses existing `referenceCornerRadius`;
standard/button shape stays as currently configured. A metric grid uses compact
MPMetric in two columns at normal text size, one column for accessibility text
or insufficient width; rows never clip monetary values or hide labels. Feature
helpers own data lists, the base owns layout. Preserve privacy masking.

Ledger changes layout only: retain wallet-specific views, current date/range
semantics, transfer/credit/adjustment actions, bulk selection, offline queue,
retry, search, detail/editor routes, account timezone and owner guard. Do not
invent unsupported reference actions or copy demo amounts into product views.
New header/row region must map to an existing base or a documented shared base.

## Verification and acceptance

- Signed native unit suite, existing base/project guards and Release compile.
- Existing live ledger CRUD/bulk/offline tests remain strict.
- Populated screenshot at normal and accessibility text sizes; use actual
  reference screenshot to compare grouping/alignment and scroll reachability.
- Shared metric grid labels/values remain VoiceOver readable and no monetary
  truncation is introduced; no fixed maximum content height.
- Public APIs, DB and production are unaffected by this UI wave. Internal screen
  contract, base docs, context, validation and unreleased changelog are updated.

UI/UX Pro Max guidance used: `compact financial metrics readability --domain ux`
returned number-formatting/contrast/line-height rules; `adaptive grid dynamic
type --stack swiftui` returned scalable layout/type guidance. Reference evidence
and existing tokens are the visual authority; no unrelated palette was adopted.
