---
artifact_type: bug
id: NATIVE-TRANSFER-WALLET-DATE-CONTRACT
status: regression_verified
owner: shared
---

# Native transfer wallet/date contract

## Finding

The native transfer editor built its wallet options from every wallet, so a
credit wallet could be selected even though the API rejects credit wallets for
internal transfers. Its date was serialized directly from a device `Date`,
which could move a selected account-local calendar day when the device timezone
differs from the account timezone.

## Intended behavior

- Source and destination selectors contain only non-credit wallets.
- The two selected wallet IDs must remain different.
- A selected transfer date is serialized at account-local noon as RFC3339,
  preserving the displayed account calendar day across device timezones.
- Server balance/ownership and idempotency rules remain authoritative.

## Verification plan

- Add pure tests for the non-credit wallet scope and account-local date round
  trip.
- Run native unit tests, Release build/guards and the signed iPhone 16
  API-backed UI suite.

## Verification result — 2026-10-06

- Native XCTest: 46/46 passed, including credit-wallet exclusion and
  account-local transfer-date round-trip.
- Release build with warnings-as-errors, base-component guard, project
  verification and `git diff --check`: passed.
- Signed iPhone 16 simulator API-backed UI suite: 10/10 passed in 199.978s.
