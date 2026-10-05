---
id: NATIVE-API-UI-FIXTURE-ISOLATION
status: regression_verified
scope: mobile-ios
---

# Native API-backed UI fixture isolation

## Symptom

`testAPIBackedWalletDetailExposesTransactionEditor` and the credit-wallet
variant assume that the fixture owner already has `Fixture Wallet` and
`Fixture Credit`. The mobile auth fixture only creates/loads the owner; it does
not seed wallets. After a previous run deletes its disposable rows, the test
opens an empty wallet screen and fails before exercising the detail UI.

## Reproduction evidence

- `testAPIBackedAssistantLoadsOverview` passed against local `:8080`.
- The wallet detail test failed with an empty-state accessibility tree and no
  `Fixture Wallet` after the same fixture auth exchange.
- The API-backed test source says “seed the disposable fixture wallet” but did
  not perform that seed itself.

## Intended behavior

Opt-in local API UI tests must seed their own owner-scoped wallet fixture and
delete it during XCTest teardown. They must not depend on simulator Keychain,
previous test order, production data, or manually prepared PostgreSQL rows.

## Verification plan

1. Add a red/green fixture setup path that exchanges the local OAuth fixture,
   creates a named basic/credit wallet through the public API, and registers
   owner-scoped teardown deletion.
2. Re-run the wallet and credit UI tests with `LOCAL_API_UI`.
3. Re-run the Assistant test, the default UI suite, native unit tests, and
   release/guard checks.

## Resolution

`LaunchTests` now seeds a basic or credit wallet through the local public API
after the fixture OAuth exchange, removes any stale same-name rows first, and
registers an async XCTest teardown that deletes the created row. The helper is
compiled only with `LOCAL_API_UI` and always targets `127.0.0.1:8080`, so it
cannot seed production or alter the release app.

Red proof: the wallet test failed with an empty owner wallet list before the
seed path existed. Green proof: wallet detail, credit ledger and Assistant
API-backed tests all pass; the full `LOCAL_API_UI` UI suite is 8/8 with zero
skips, and a post-run API check reports zero `Fixture*` wallets.
