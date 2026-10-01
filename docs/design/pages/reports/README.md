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
