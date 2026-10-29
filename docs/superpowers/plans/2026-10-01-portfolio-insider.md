# Portfolio and Money Insider implementation plan

## Scope

Add a narrow, auditable portfolio ledger and expose a deterministic Money
Insider summary from the existing finance query layer. Portfolio trades never
mutate wallet balances or ordinary income/expense reports.

## Portfolio contract

- Assets are owner-scoped records with a symbol/name and optional latest price
  snapshot. Trade rows are immutable buy/sell events with decimal quantity,
  VND unit price, fee, timestamp and stable ordering.
- Quantity is accepted as a decimal string and stored as fixed-point units
  (8 decimal places). Prices and fees are positive integer VND.
- Buy adds `(quantity * price) + fee` to moving cost. Sell removes quantity at
  the pre-sale weighted average cost; sell fee reduces realized proceeds.
- A sell cannot exceed the available position. Deleting or editing history is
  not exposed in this first slice; compensating trades preserve auditability.
- A latest manual price is optional. Missing price is unknown, never zero.

## Money Insider contract

`GET /api/v1/reports/insider?month=YYYY-MM` returns report-derived totals,
previous-period comparison, top expense categories, top expense rows, average
daily spend, spending/income ratio with zero-baseline semantics, and an
explicit `estimated` flag for projections. It uses the account timezone and
report-included rows only; no model call or fabricated narrative is required.

## Verification

- Unit tests cover fixed-point parsing and weighted-average buy/sell sequences,
  sell-over-position rejection, zero-income ratio handling, and owner scope.
- PostgreSQL tests cover portfolio persistence and report exclusion.
- Frontend checks cover the shared Portfolio and Insider panels, then full
  design/type/build suites.

## Delivered

- Portfolio is implemented by migration `000027_portfolio_ledger.up.sql`,
  owner-scoped Go repository/handlers, generated Swagger, and the Account
  Portfolio panel. Quantity uses eight fixed-point decimal places; prices and
  fees are integer VND and API validation rejects null/negative values.
- `GET /api/v1/reports/insider` now reads account-local report rows with
  optional wallet/category filters and powers the Money Insider card. The
  report excludes transfers/adjustments, keeps zero-income ratios unknown, and
  exposes an explicit `estimated` flag.
- The web release gate covers AI entry browser behavior and PWA manifest,
  service-worker registration, and user-triggered install prompting. The
  verified commands are documented in `docs/operations/LOCAL_RELEASE_GATE.md`.
