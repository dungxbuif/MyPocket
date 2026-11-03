# Reports and Money Insider page contract

## Evidence-backed findings — local in_review, 2026-11-02

The [approved design](../../../work/tickets/REPORTS-AI-USEFULNESS-20261102-DETAIL_DESIGN.md)
extends the November 1 candidate below. Backend-owned budget periods/carryover
and matched comparison intervals supply typed findings. The provider selects
eligible IDs/actions; the server renders names, amounts and claims. Findings are
distinct, capped at three, and do not infer purchase causes from category totals.
Non-comparable periods show “Chưa đủ dữ liệu tương ứng để so sánh”.

`MPReportInsightCard` renders title, supplied metrics, period, explanation and
actions; compact Insider omits the duplicated full explanation.
`MPReportAnalysisArticle` reuses these cards and keeps one saved-state indicator.
“Xem giao dịch” uses the saved snapshot and exact target, including cross-month
budget periods and matched-day category intervals. Expired/revision-stale targets
show update guidance; they never silently substitute a live-month query.

Opening/restoring reads saved analysis without generation. Same-day day/week
changes reuse unchanged evidence fingerprints; persisted evidence endpoints stay
exact. Historical payloads remain readable with explicit stale/unknown provenance
where applicable. “Hỏi tiếp” preserves the draft and carries analysis/snapshot/
finding IDs as hidden context; its neutral visible prompt contains no amounts or
raw IDs. The shared fullscreen Assistant hides navigation and Feedback. Budget
cards consume server-filtered items and disclose any remaining relevant count.

Targeted native/backend checks and three source reviews pass. Live November and
October provider results, day/week reuse, the exact 1,850,000 VND vacuum drilldown,
and two legible saved November findings were observed. Report Feedback remains a
floating overlay with no reserved container. Follow-up returned correct overrun
prose and one relevant budget card, but its optional transaction card included
18 broadly searched rows, including transfers. This remains an acceptance gap;
not all Assistant UX is accepted. Owner acceptance stays open. Full evidence and
non-green backend/web gates: [validation](../../../work/VALIDATION_MATRIX.md).

## New native local candidate — 2026-11-01

The owner requests a visual redesign and saved AI analysis restoration.
[Requirements and technical design](../../../work/tickets/REPORTS-INSIDER-VISUAL-20261101-DETAIL_DESIGN.md)
and [execution plan](../../../superpowers/plans/2026-11-01-reports-insider-visual-and-saved-analysis.md)
define cumulative/weekly trends, category allocation, comparable-period drivers,
shared Insider facts and resumable saved analysis. The local iOS/backend candidate
now implements this contract; production rollout and the provider-backed analysis
UAT gate remain open.
For that candidate, the new design supersedes the older native web-parity,
top-transaction-list and comparison-donut layout directions below. Historical
verification statements below describe their original changes only.

## Existing behavior and historical implementation notes

Reports render facts from the owner-scoped ledger. Money Insider is a
deterministic report card, not a model-generated financial conclusion.

The page uses the shared `SurfaceCard`, `SectionTitle`, `Text`, chart and
status atoms. It shows current month totals, top three expense categories,
top five expense rows, average daily spend, previous-period change and the
spending/income ratio. A missing or zero income baseline renders “Chưa có thu
nhập để tính tỷ lệ”; it must never render Infinity or a fabricated percentage.

The API accepts an account-local `YYYY-MM` month and optional owner wallet or
category filters. Only `included_in_reports` income/expense rows count;
transfers, adjustments and credit-payment system rows are excluded. The
`estimated` flag is explicit and is false for the current ledger-backed
slice. Budget references and projections remain separate, labeled contracts.

The mounted trend surface uses `BaseTrendChart` on web and `MPTrendChart` on
native. Both group the existing owner-scoped transaction response by
account-local day, render income and expense separately, and show a plain empty
state when the selected month has no eligible rows. This adds no endpoint or
schema; the chart is a presentation of the same ledger contract.

## Native/web layout parity — 2026-10-13

The current web Reports panel is the layout source of truth: summary with an
income/expense comparison chart and three metrics; a separate, always-visible
monthly trend card; category totals; monthly activity; and Money Insider. The
Money Insider card includes a comparison donut when a previous-period
percentage exists. Native previously placed a category chart inside the
summary, hid the trend behind a `Chi tiết / Xu hướng` segment, and omitted the
comparison donut from its Reports-only Money Insider card (the donut existed
only on Overview).

Native now uses a shared DesignSystem income/expense comparison chart, renders
the trend as its own always-mounted card without a mode toggle, and reuses
`MPDonutChart` in the Reports Money Insider card. Comparison-bar heights are
normalized against the greatest positive total; zero and negative values render
as zero-height bars. Category totals use the month summary response and are
rendered as rows rather than a duplicate chart. The trend retains account-local
day grouping, report-inclusion filtering and amount masking; its VoiceOver value
exposes the same plotted values. No endpoint, schema, or unfiltered-ledger
calculation was added.

Verification: native unit/contract tests pass 147/147; focused API-backed
Reports UI tests pass 1/1 for Money Insider/comparison/trend hierarchy and 1/1
for persisted income/expense values with the excluded row omitted. Screenshots
are recorded in `/tmp/mypocket-reports-ui-verified-attachments/`. Release
simulator and generic iOS-device builds pass; project/base-component guards
and `git diff --check` pass. Tests used a disposable local API/DB and Redis
database only. No public API or AI-agent reference changes are required.

Native `MPTrendChart` exposes one VoiceOver value in chronological order with
the daily income and expense totals from those same plotted points. When account
amount masking is enabled, every exposed amount uses the shared masked value;
the accessibility summary does not recalculate or broaden report scope.

## Native insider implementation candidate — 2026-11-01

`MonthlyReportsView` owns the additive aggregate-first path: exact-scope
in-memory caching, weekly/category bars, selection drilldown, masking and
on-demand `ReportAnalysisView`. It does not request `/transactions` while
loading a report. When `/reports/monthly` is not deployed, the view falls back
only to the existing owner/month aggregate endpoints and labels that state.
The report analysis handoff pre-fills Assistant without sending a message.

Automated tests were intentionally skipped for this owner-requested implementation
pass. Backend and simulator builds pass; manual rich-data UAT confirms summary,
30-day comparison, cumulative trend, category allocation, contributors, Insider
and snapshot-bound category drill-down. AI provider creation/reopen and physical
device acceptance remain open.

## Native savings-report hierarchy — Task 5 reference redesign

The savings report composes shared finance bases only. `MPFinanceSummaryCard`
owns the title/context, primary outcome, chart slot, supporting metrics and
footer/action order. `MPFinanceComparisonChart` owns normalized comparison-bar
geometry, semantic financial tones and masking-aware VoiceOver values. The
screen supplies existing API values and actions; it does not supply colors,
radii, spacing or chart geometry.

| Region | Shared base | Existing values |
| --- | --- | --- |
| Report filter | `MPFinanceCard`, `MPTextField`, `MPOptionSelector`, `MPButton` | Month and optional goal-wallet scope |
| Goal overview | `MPFinanceSummaryCard`, `MPFinanceComparisonChart`, `MPMetricGrid` | Closing balance as the primary outcome; opening/closing balance comparison; existing deposit, withdrawal, adjustment, transfer and net-change totals |
| Goal detail | `MPFinanceCard`, `MPMetricGrid`, `MPProgressBar`, `MPStatusBadge` | Existing per-wallet balance, target, progress and transaction count |
| Account surplus | `MPFinanceSummaryCard`, `MPFinanceComparisonChart`, `MPMetricGrid` | Net surplus as the primary outcome; existing income/expense comparison, rate and transaction count |
| Drilldown | `ReportTransactionsView`, `MPTransactionRow` | Existing report transaction endpoint and paging |

The two concepts remain separate: goal-wallet balance movement is not account
income/expense surplus. Internal goal-wallet transfers keep their current
netting behavior. The wallet filter still applies only to the goal section.
Masking covers every primary and supporting amount, including chart VoiceOver;
empty and nil-rate states keep their existing copy. No API, payload, schema or
report arithmetic changes are part of this migration.

Verification plan: focused native tests cover zero/negative comparison values
and masking; the report/model tests keep the existing
goal-versus-surplus decoding assertions; the focused iOS test target, base
component guard and warnings-as-errors build must pass. Small-phone, landscape,
Dynamic Type and dark-mode visual acceptance remain part of integrated Task 7/8
UAT when no dedicated Task 5 screenshot harness is available.

Implementation evidence (2026-10-26): the isolated
`JarsMonthsReferenceTests` run passed 7/7 and the complete
`MyPocketIOSTests` target passed 408/408 on the iPhone 16 iOS 18.5 simulator.
The Release simulator build passed with `SWIFT_TREAT_WARNINGS_AS_ERRORS=YES`;
`mobile-ios/scripts/check-base-components.sh`,
`mobile-ios/scripts/verify-project.sh` and `git diff --check` also passed.
No dedicated savings-report screenshot fixture exists in this task scope, so
small-phone/landscape/Dynamic Type/dark-mode visual acceptance remains an
explicit integrated UAT item rather than an inferred pass.

## Native report hierarchy — organization/planning continuation

The main native report route uses `MPFinanceCard` for filter, trend, breakdown,
and activity sections, and `MPFinanceSummaryCard` for the month
income/expense/net outcome. `ReportsReferencePresentation` supplies section
spacing and the summary accessibility key consumed by the rendered shared card.
Category values use `MPMetricRow`; charts keep
their existing report-filtered inputs and masking behavior.

Month validation, three-read completeness, exact-month offline fallback,
snapshot persistence, report inclusion rules, timezone grouping, savings-route
navigation, and all API payloads remain unchanged. Focused tests cover the
exact-month persisted-cache eligibility policy and exercise report reducers
with an ordinary expense plus both legs of a paired transfer; only the ordinary expense
contributes to trend and weekly totals. This is reducer-level acceptance of the
decoded `includedInReports` contract, not an end-to-end API persistence claim
or rendered-hierarchy test. Integrated visual/device acceptance remains a later
redesign gate.

Review remediation evidence (2026-10-27): the four owned reference test classes
passed 24/24 and the normally signed `MyPocketIOSTests` target passed 422/422 on
the iPhone 16 iOS 18.5 simulator. The Release warnings-as-errors build,
base-component guard, and project verification also passed.

## Hũ trong Báo cáo — not applicable on iOS 2026-11-03 (the iOS Reports screen has no jar section; plan wording lives in the Jars, Month and jar detail screens)

"Hũ cộng dồn" giữ dữ liệu; nhãn "Chênh lệch" → "Kế hoạch còn lại"; thêm dòng "Chưa phân bổ" để tổng theo hũ khớp tổng chi tháng. Không thêm biểu đồ. Xem [JARS-AS-WALLET-CARDS](../../../work/tickets/JARS-AS-WALLET-CARDS-20261103-DETAIL_DESIGN.md).
