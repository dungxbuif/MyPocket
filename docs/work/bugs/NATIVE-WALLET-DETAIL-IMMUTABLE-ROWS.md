---
artifact_type: bug
id: NATIVE-WALLET-DETAIL-IMMUTABLE-ROWS
status: resolved
updated: 2026-10-06
---

# Native wallet detail must preserve immutable transaction semantics

## Observation

`WalletDetailView` rendered every ordinary-wallet row as an editable
`TransactionEditorView`. That included transfer and balance-adjustment rows.
The backend correctly rejects ordinary edits for those rows, but the native UI
offered the wrong action before the request failed. `TransactionDetailView`
also treated `adjustment` rows as editable/deletable and offered Travel Mode
linking, although adjustments are immutable and are not ordinary travel-linked
transactions.

## Required behavior

- Ordinary income/expense rows may open the normal edit/delete actions.
- Transfer rows may only offer paired transfer edit/delete actions.
- Adjustment rows remain read-only and explain that a compensating adjustment
  is required.
- Credit rows remain read-only in the ordinary transaction detail.
- Travel Mode linking is available only for ordinary non-credit rows.

## Resolution and proof

Added `TransactionDetailPolicy`, routed basic/goal wallet-detail rows through
`TransactionDetailView`, passed the complete wallet set to transfer editing,
and corrected adjustment signed amounts using `adjustment_direction`.
Adjustment rows now remain read-only and cannot be linked to Travel Mode;
transfer rows use paired actions; ordinary rows retain edit/delete actions.

Native unit/contract XCTest passes 62/62. The targeted wallet mutation UI
regression passes 1/1, the full API-backed simulator suite passes 15/15 in
310.705 seconds, and the Release/base-component/project/diff gates pass. Local
health remained 200 and fixture-wallet cleanup left no `Fixture*` wallets.
Physical-device, real OAuth, provider-backed AI/OCR and staging/production
checks remain external release gates.
