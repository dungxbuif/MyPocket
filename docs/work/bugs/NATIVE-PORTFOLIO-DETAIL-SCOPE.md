---
artifact_type: bug
id: NATIVE-PORTFOLIO-DETAIL-SCOPE
status: implemented
created: 2026-10-06
updated: 2026-10-06
---

# Native portfolio detail parity

## Finding

The native portfolio route loads the same summary payload as the web route but
only renders quantity and cost basis for each position. The product contract
also requires weighted average cost, realised/unrealised P/L, current market
value when priced, and chronological buy/sell history with quantity, price,
fee and date.

## Intended behavior

- Keep the shared metric, surface and transaction-row base components.
- Show per-position cost basis, weighted average, realised P/L, current value
  and unrealised P/L; preserve “Chưa có giá” when no manual price exists.
- Show each asset's append-only trade history in chronological order with
  localized buy/sell labels and quantity/price/fee/date context.
- Do not alter the portfolio API, accounting or wallet boundaries.

## Verification plan

- Add/extend pure model coverage for position/trade presentation helpers.
- Run native XCTest, Release warnings-as-errors, base/project guards and the
  authenticated More-route simulator smoke.

## Implementation result

- Native asset cards now render weighted average cost, realised P/L, current
  market value and unrealised P/L; unpriced values remain explicitly unknown.
- The same cards render newest-first buy/sell history with localized labels,
  quantity, unit price, fee, account-timezone date and optional note.
- All presentation continues to use the shared metric, surface, assistant and
  transaction-row bases; API/accounting contracts are unchanged.

## Verification result — 2026-10-06

- Native XCTest: 56/56 passed, including trade filtering/order and label tests.
- Release simulator build with Swift warnings treated as errors: passed.
- Base/project guards and `git diff --check`: passed.
- Signed iPhone 16 simulator More-route smoke: 1/1 in 121.948s.
