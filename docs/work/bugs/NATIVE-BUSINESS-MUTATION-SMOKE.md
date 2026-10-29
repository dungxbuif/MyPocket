---
artifact_type: bug
id: NATIVE-BUSINESS-MUTATION-SMOKE
status: regression_verified
created: 2026-10-06
updated: 2026-10-06
---

# Native business mutation smoke

## Finding

The native simulator suite initially proved navigation and one ordinary
transaction create/update path, but did not prove the UI-to-API write paths
for wallet creation, atomic transfer, or credit entry/payment. Backend
integration tests cover those contracts independently; that evidence is not
enough to prove that SwiftUI labels, selectors, draft payloads, and refresh
behavior are wired correctly. The first wallet mutation run also exposed two
native-only issues: the wallet editor could be recreated as a create editor
when opening an existing wallet, and a numeric keyboard could cover the credit
note field.

## Intended behavior

- A wallet created from the native editor appears in the wallet list with the
  server-assigned name and can be removed by teardown.
- A transfer submitted from the transaction overflow flow creates the paired
  source/destination rows, refreshes the ledger, can update both linked rows
  without implying that source/destination can be changed, and can be deleted
  as one atomic pair.
- A credit entry and a payment submitted from the credit statement flow refresh
  the statement and expose the resulting rows; the payment must use a
  non-credit source wallet.

## Implementation decision

Extend the opt-in signed `LOCAL_API_UI` XCTest suite with disposable,
owner-scoped fixture wallets and deterministic UI assertions. The fixture API
only creates/deletes test records in `mypocket_stage_v1`; no production or
provider credential is used. Teardown must remove transfer pairs and all
fixture wallets even when an assertion fails.

## Implementation and verification result

- Wallet editor presentation now carries an explicit `.create` or `.edit(wallet)`
  payload, so editing an existing wallet always sends `PATCH /wallets/:id`.
- The shared native amount field exposes a keyboard “Xong” action so a numeric
  keyboard can be dismissed before entering a note or submitting a form.
- Transfer edit mode disables source/destination selectors because the existing
  idempotent pair-update API accepts only amount, date and note. A regression
  test proves those fields update both persisted rows and keeps wallet IDs
  unchanged.
- Signed iPhone 16 Simulator `LOCAL_API_UI` mutation tests pass for wallet
  create/update, atomic transfer create/update/delete, and credit entry/payment.
- The complete signed API-backed UI suite passes 14/14 in 293.192 seconds.
- Disposable Stage v1 fixture counts after the run are wallets=0,
  transactions=0, feedback=0.

## Verification plan

1. Run the focused mutation tests on iPhone 16 Simulator
   `EC80B092-CEF1-4A8E-A5A3-7420298EFD4D` against the local API.
2. Run the complete signed `MyPocketIOSUITests` suite.
3. Query the disposable database for zero `Fixture*` wallets, transactions,
   and feedback rows.
4. Record the result in `docs/work/VALIDATION_MATRIX.md` and this bug's status.

## External gates

Real Google OAuth, provider-backed AI/OCR, authenticated staging/production,
and physical-device signing remain separate release gates.

## 2026-10-08 transfer-edit addendum

The transfer-edit regression was reproduced and fixed after the original smoke
run. A dedicated disposable PostgreSQL database and loopback Redis container
were used (not `mypocket_stage_v1`): the API-backed iPhone 16 test now creates,
edits, reads back both pair rows, verifies the wallet IDs remain unchanged and
deletes the pair. The full current native suite passed 145/145; see the
[validation matrix](../VALIDATION_MATRIX.md#native-transfer-pair-edit-integrity--2026-10-08).
The dedicated DB/container and DerivedData were removed; retained xcresult
bundles contain the verification evidence.
