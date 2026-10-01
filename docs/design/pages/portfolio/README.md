# Portfolio page contract

The portfolio page is an account-scoped ledger for assets that are not
wallet transactions. It uses the same mobile shell, `SurfaceCard`, `Text`,
`BaseButton`, and form atoms as the rest of the web app.

## Required states

- Empty state explains that adding a buy creates the first position.
- Each asset shows symbol, name, quantity, cost basis, weighted average cost,
  and realised P/L. A latest manual price shows market value and unrealised
  P/L; without a price the value is labelled “Chưa có giá”, never zero.
- The trade history is chronological and labels buy/sell, quantity, price,
  fee, and occurred time. History is append-only; fixes are compensating
  trades.
- Sell validation remains visible before submit and the API rejects an
  over-position sale.

## API boundary

The page consumes `/api/v1/portfolio/*` only. Portfolio data never changes
wallet balances, transaction reports, budgets, or jars. Quantity is a decimal
string and the backend stores eight fixed-point decimal places.
