---
artifact_type: bug
id: NATIVE-WALLET-DETAIL-BALANCE-REFRESH
status: regression_verified
owner: shared
---

# Native wallet detail balance refresh

## Finding

`WalletDetailView` received an immutable wallet snapshot from the list. After
adding or editing a transaction, the detail reloaded rows but kept the old
`current_balance` in the header and goal progress. This could show a balance
that no longer matched the API ledger.

## Intended behavior

- Detail load refreshes the selected wallet by its stable ID from the
  owner-scoped wallet list.
- Transaction/category/credit payload failures remain distinguishable; a
  missing optional category payload must not hide valid transaction rows.
- Add/edit transaction reloads both rows and the wallet summary.

## Verification plan

- Add a fresh wallet lookup to detail loading and keep category labels
  optional.
- Run native unit tests, Release build/guards and the signed iPhone 16
  API-backed UI suite.

## Verification result — 2026-10-06

- Native XCTest: 46/46 passed.
- Release build with warnings-as-errors, base-component guard, project
  verification and `git diff --check`: passed.
- Signed iPhone 16 simulator API-backed UI suite: 10/10 passed in 199.887s.

Wallet detail now refreshes the owner-scoped wallet summary after loading and
does not discard valid transaction rows when optional category metadata fails.
