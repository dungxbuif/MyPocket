# Reports and Money Insider page contract

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
