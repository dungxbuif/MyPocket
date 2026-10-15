# Native credit empty-state parity implementation plan

## Goal

Match the web credit ledger when the selected credit wallet has no statement
items: keep the shared card heading and show an explicit empty-state message
instead of rendering a blank card.

## Contract

- Keep `MPAssistantCard` as the shared/base surface for the ledger section.
- When `statement.items` is empty, render `MPEmptyView` with the web copy
  `Chưa có giao dịch tín dụng.`.
- Keep existing transaction rows, credit actions, owner-scoped API calls,
  privacy behavior and accessibility identifiers unchanged.

## Verification

- [x] Add a pure unit contract for the empty-state copy/branch.
- [x] Add or update the authenticated credit UI contract to assert the empty
  state before creating the first credit entry.
- [x] Run focused/full native tests, Release warnings-as-errors, project and
  base-component guards, and `git diff --check`.
- [x] Reconcile validation/audit/iOS docs and changelog; no API/schema/
  migration change is expected.

## Implementation notes

The web route already exposes this copy for an empty statement. Native should
not invent a new credit-specific card or bypass the shared empty-state base.

## Shipped evidence

- `CreditLedgerCopy.emptyStatement` and the shared `MPEmptyView` branch are
  implemented in `CreditView`.
- RED compile evidence is `/tmp/mypocket-credit-empty-red`.
- Focused copy regression is 1/1 at
  `/tmp/mypocket-credit-empty-unit/Logs/Test/`.
- API-backed credit route is 1/1 at
  `/tmp/mypocket-credit-empty-ui2/Logs/Test/`.
- Full native, Release and guard evidence is recorded in the validation
  matrix after the follow-up run.
