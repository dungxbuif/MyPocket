---
artifact_type: bug
id: NATIVE-PORTFOLIO-TRADE-VALIDATION
status: regression_verified
owner: shared
---

# Native portfolio trade validation

## Finding

The native portfolio form only checked that quantity was non-empty and parsed
the price/fee by stripping non-digits. The backend contract is stricter:
`side` must be `buy` or `sell`, quantity must be a positive decimal with no
more than eight fractional digits, and unit price/fee must be integer VND
values that are not negative. Invalid or silently rewritten drafts should be
rejected before an API request.

## Intended behavior

- Accept only `buy`/`sell` sides.
- Accept quantity `> 0` with up to eight decimal places and no sign or
  exponent notation.
- Accept unit price as a non-negative integer VND value.
- Accept fee as an empty/zero value or a non-negative integer VND value.
- Preserve the entered quantity text instead of silently changing its value.

The server remains authoritative, including the insufficient-position check for
sell trades.

## Verification plan

- Add a pure native validator and boundary tests for quantity precision,
  signs, side, price and fee.
- Run native unit tests, release build/guards and the signed API-backed iPhone
  16 simulator suite.

## Verification result — 2026-10-06

- Pure native XCTest: 45/45 passed, including invalid side, zero quantity,
  excessive precision, exponent notation, decimal price and negative fee.
- Release build with Swift warnings treated as errors, base-component guard,
  project verification and `git diff --check`: passed.
- Signed iPhone 16 simulator API-backed UI suite: 10/10 passed in 199.878s.

The native editor now rejects malformed drafts without silently stripping
characters into a different quantity/price/fee before the request is sent.
