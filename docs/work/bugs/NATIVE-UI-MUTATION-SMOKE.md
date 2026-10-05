# Native UI mutation smoke

Status: `regression_verified`

## Intended behavior

On a signed-in iOS build backed by the disposable local stage API, a user can:

1. open a basic wallet;
2. create an expense with an amount and note;
3. see the persisted note in the wallet ledger after the editor closes; and
4. reopen that row, edit the note, save, and see the updated value.

The test must use the same SwiftUI base components and the real API client. It
must not assert only that a modal opened: the persisted row and edited value are
the observable state contract.

## Technical decision

Use a disposable fixture wallet created through the local API fixture helper and
delete it in teardown. The test uses the wallet-detail editor because it exercises
the native create/update transaction path without depending on seeded production
data or a fragile global transaction list. The existing API-backed UI suite keeps
the `LOCAL_API_UI` gate so normal simulator/unit runs remain offline-safe.

## Verification plan

- Run the focused signed UI test against `mypocket_stage_v1`.
- Run the complete signed API-backed UI suite on iPhone 16 Simulator.
- Record the result in `docs/work/VALIDATION_MATRIX.md`, the phase log, and the
  unreleased changelog entry.

## Result — 2026-10-06

- PASS: `testAPIBackedWalletDetailCreatesAndUpdatesTransaction` created a
  disposable expense through the native editor, observed the persisted ledger
  row, reopened it, changed the note, and observed the updated row.
- PASS: the latest signed simulator execution against `mypocket_stage_v1`
  completed in 18.348 seconds; fixture teardown removed the wallet after the
  test.
