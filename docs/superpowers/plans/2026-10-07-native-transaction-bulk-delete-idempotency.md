# Native bulk-delete idempotency fix

Date: 2026-10-07
Scope: native iOS transaction ledger bulk deletion.

## Intended behavior

Confirming a valid selection of 1–100 ordinary transactions sends exactly one
`POST /transactions/bulk-delete` request with the unique transaction IDs. The
client removes the rows locally only after that request succeeds. Transfer,
adjustment and credit rows remain excluded by the existing selection policy.

## Technical decision

The existing `TransactionsViewModel.deleteMany` is the owner of the mutation.
It must not retry implicitly or issue a duplicate request. Retry remains an
explicit user action after the error message is shown, preserving the backend's
atomic contract and avoiding duplicate audit/usage events.

## Verification

- Add a fake `FinanceAPI` call-count regression asserting one bulk-delete call.
- Keep the existing 1–100 boundary and selectable-row tests.
- Run the full native unit target, Release warnings-as-errors build, project and
  base-component guards, and `git diff --check`.
- Reconcile `PORT-100-AUDIT.md`, `VALIDATION_MATRIX.md` and the release
  changelog with the resulting evidence.

## Result

- Focused regression: 1/1 passed at `/tmp/mypocket-bulk-delete-green/Logs/Test/`.
- Native unit target: 111/111 passed at `/tmp/mypocket-bulk-delete-unit.xcresult`.
- Release warnings-as-errors build passed at `/tmp/mypocket-bulk-delete-release`.
- Project, base-component and diff guards passed.
- The all-target scheme invocation was stopped after the known Xcode UI-runner
  hang; the unit-only target completed successfully and is the authoritative
  regression gate for this logic change.
