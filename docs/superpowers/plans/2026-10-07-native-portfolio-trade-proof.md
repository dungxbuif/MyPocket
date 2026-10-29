# Native portfolio trade contract proof

## Intent

Close the last weak automated-evidence point in the native portfolio parity
matrix. The screen already renders the shared trade form and calls
`FinanceAPI.createPortfolioTrade`; add a transport-level contract test so the
real write path cannot drift from the Go/web API.

## Contract

- POST `/portfolio/assets/{assetID}/trades` under the owner session.
- JSON uses backend snake_case keys (`side`, `quantity`, `unit_price`, `fee`,
  `occurred_at`, optional `note`).
- Native sends an idempotency key for the immutable write.
- No wallet balance or ordinary report mutation is implied by this endpoint.

## Verification

Add the test first (RED for the missing fixture transport/test), implement only
the test fixture needed for the existing API method, then run the full native
suite, Release warnings-as-errors build, base guards and diff/docs checks.

## Non-goals

Do not add a destructive trade-delete route or mutate production data. Portfolio
trades are append-only; cleanup remains compensating-trade/domain policy.

## Result

Added `APIClientTests.testPortfolioTradeUsesOwnerPathSnakeCaseBodyAndIdempotency`.
The focused native contract passes 1/1 at
`/tmp/mypocket-portfolio-trade-redgreen.xcresult`. The existing API-backed
portfolio asset/price UI flow also reran 1/1 at `/tmp/mypocket-portfolio-ui.xcresult`;
no API or schema behavior was changed.
