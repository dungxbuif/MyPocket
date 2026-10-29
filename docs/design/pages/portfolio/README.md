# Portfolio page contract

## Native editor workflow — 2026-10-14

List owns summary and flat asset sections only. Add/trade actions open
`mpFullScreenCover` + `MPEditorScaffold`; per-asset trade uses the same editor.
Manual price uses short `mpFormBottomSheet`. No creation fields permanently
above the list. Cancel retains local draft with no API write; errors keep editor
open. Busy state prevents editing, repeated submit and dismissal; success closes
then refreshes. Assets use flat `MPMetric` and history rows inside one surface.
Quantity uses `MPNumberField(.decimal)` with locale normalization and unchanged
positive/eight-decimal financial contract. [Guide](../../../guides/portfolio-ios.md).
Native portfolio VND unit-price, fee, correction and manual-price fields opt into
`MPAmountField(groupsDigits: true)`: canonical integer drafts show Vietnamese dot
grouping while bound values remain ungrouped ASCII digits. Invalid pasted money
strings remain visible for validation; quantity does not use money grouping.
Summary and asset position values use existing compact MPMetric consistently,
through `MPMetricGrid` with shared section/grid rhythm. No metric is hidden;
unknown value still reads
Chưa có giá. Actual SE UI reproduced hero/compact mismatch then verified matching
line heights;191/191 unit/API UI gate passes. Render still has Feedback overlap
on asset identity; floating-content design remains blocked and whole-screen UAT
is not accepted. See completion ticket/validation for exact evidence.
SE API editor error/cancel/create/trade/price proofs pass with server readback;
retry and keyboard/footer screenshots inspected. Final accessibility/device
acceptance and asset-list density sweep remain open. Earlier inline-form proofs
alone do not establish this editor workflow's acceptance.

The portfolio page is an account-scoped ledger for assets that are not
wallet transactions. It uses the same mobile shell, `SurfaceCard`, `Text`,
`BaseButton`, and form atoms as the rest of the web app.

## Required states

- Empty state explains that adding a buy creates the first position.
- Each asset shows symbol, name, quantity, cost basis, weighted average cost,
  and realised P/L. A latest manual price shows market value and unrealised
  P/L; without a price the value is labelled “Chưa có giá”, never zero.
- The trade history is chronological and labels buy/sell, quantity, unit
  price, fee, and occurred time. History is append-only; the native correction
  editor previews then commits an immutable replace/void revision after the
  user confirms it. The original row remains in audit history.
- Sell validation remains visible before submit and the API rejects an
  over-position sale.

## API boundary

The page consumes `/api/v1/portfolio/*` only. Portfolio data never changes
wallet balances, transaction reports, budgets, or jars. Quantity is a decimal
string and the backend stores eight fixed-point decimal places.

Native position/valuation data comes from `/portfolio/summary`; immutable
history loads from `/portfolio/assets/{id}/trades` because the summary
repository currently does not populate its optional `trades` field. A
successful buy/sell reloads both sources before the page presents refreshed
position and history.

Asset cards expose a prefilled `Sửa tài sản` editor using the existing
owner-scoped metadata PATCH. The editor changes symbol/name only, keeps the
same asset ID and trade history, retains the draft after an API error, and
reloads summary/history after success. It reuses `mpFullScreenCover`,
`MPEditorScaffold`, `MPTextField`, `MPStatusMessage` and `MPButton`.

Unused-asset deletion uses the existing owner-scoped
`DELETE /portfolio/assets/{id}` contract. The native card shows the shared
destructive action and confirmation only when that asset has no trade history;
success reloads the summary and history. Assets with any buy/sell remain
append-only and do not show a delete action. Historic trade replace/void uses
the preview/commit correction routes, which require a matching server rollout;
there is no direct trade-row PATCH or DELETE.

`Giá đơn vị` is an integer VND price per unit. A buy of quantity `1.5` at
unit price `100000` contributes `150000` VND before fee to cost basis. The
client sends quantity and unit price as separate fields; it never divides an
entered total by quantity.

Task 6 remains presentation-only: summary values are mapped into
`PortfolioReferencePresentation.summaryItems` and rendered by `MPMetricGrid`;
privacy mode marks each populated financial metric as masked at the shared
metric boundary. The screen renders mask glyphs while VoiceOver announces
`Đã ẩn số tiền`; unknown values continue to read `Chưa có giá`.
Asset/trade/price/correction editors continue to use the existing shared editor,
field and confirmation bases and retain their owner/version checks.

## Simulator/API verification contract

The native gate must exercise the real owner-scoped API, not only render the
route: create a disposable unused asset, load it in the native portfolio, set a
manual price, reload the summary and assert the refreshed value card. The
fixture is deleted through the owner-scoped asset delete route after the test.
Trade execution is separately verified through an API-backed native UI buy
that asserts the refreshed position, visible history and persisted trade
fields. Because trade-bearing assets are intentionally immutable and cannot be
deleted through the public cleanup route, remove only that synthetic asset and
trade by exact owner + asset ID from the disposable stage database after the
test. No portfolio fixture may remain in the local database. Append-only trade
math and sell-over-position rules are covered by backend entity tests. The
unused-asset delete UI case is covered by
`testAPIBackedPortfolioUnusedAssetCanBeDeletedWithConfirmation`; its first
rerun is pending because the disposable local PostgreSQL fixture was
unavailable, not because of a product assertion. The trade-flow regression also
asserts that a traded asset has no destructive delete control.
