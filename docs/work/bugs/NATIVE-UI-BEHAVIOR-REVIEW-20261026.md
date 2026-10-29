---
artifact_type: bug_batch
id: NATIVE-UI-BEHAVIOR-REVIEW-20261026
status: verified
owner: shared
approved_by: human
updated: 2026-10-26
---

# Native UI Behavior Review Fixes

## Context And Scope

The owner approved a narrow regression-fix batch for five reviewed native iOS
behaviors. This work must not read, rewrite, or otherwise mutate transaction
reconciliation data. The implementation is limited to the affected SwiftUI
presentation/state helpers and focused XCTest coverage.

In scope:

1. Hide a revealed one-time API secret whenever the app leaves the active scene.
2. Make the goal-wallet selector actually change the selected wallet.
3. Normalize pasted VND values that contain accepted grouping separators while
   keeping the bound draft as canonical ASCII digits.
4. Preserve the portfolio privacy accessibility contract when values are masked.
5. Reuse one transaction signed-amount policy across ledger and wallet detail.

Out of scope: API/schema changes, transaction data reconciliation, deployment,
new dependencies, visual redesign, or changes to financial accounting rules.

## Root Cause Evidence

- `AccountView` stores reveal state locally but does not observe scene lifecycle,
  so an explicitly revealed secret can remain visible in the app-switcher snapshot.
- The goal detail selector invokes optional `onSelectWallet`, but `WalletsView`
  constructs `WalletDetailView` without supplying that callback. Its binding getter
  is also pinned to the initial wallet ID.
- Grouped `MPAmountField` translates single-key edits against formatted display
  text, while paste falls through and writes the formatted string directly into
  the canonical bound draft.
- Portfolio summary and position items are preformatted with masked glyphs and
  then passed to `MPMetricGrid(masked: false)`, so VoiceOver treats the glyphs as
  literal text instead of announcing that the amount is hidden.
- `TransactionsPresentation` and `WalletDetailTransactionPolicy` implement the
  same income/expense/adjustment sign logic independently.

## Intended Behavior And Fix Strategy

- Add a pure API-secret lifecycle policy and reset reveal state on `.inactive`
  and `.background`; keep the secret transient and unchanged.
- Route goal-wallet selection through the owning `WalletsView.selectedWalletID`.
  Reject aggregate selection for the goal-only selector and keep selection valid
  when wallet data refreshes.
- Add a pure grouped-VND draft normalizer used by paste/change handling. Accept
  ASCII digits plus Vietnamese dot grouping, regular spaces, non-breaking spaces,
  and narrow non-breaking spaces; reject decimal/comma or arbitrary text.
- Pass raw formatted portfolio values together with `masked: true` at the shared
  metric boundary so visible and spoken privacy behavior are consistent.
- Move persisted-transaction sign resolution into `MPTransactionPresentation`
  and delegate both ledger policies and row rendering to it.

## Verification Plan

TDD sequence:

1. Add focused failing tests for secret lifecycle, goal selection routing, grouped
   VND paste, portfolio masked accessibility state, and canonical transaction signs.
2. Run each focused test before production edits and record RED.
3. Implement the smallest fixes, rerun focused tests GREEN, then run the full
   native unit/contract suite, warnings-as-errors build, repository guards, and
   `git diff --check`.

## Documentation Impact

- Update account settings, wallet, portfolio, transaction, and shared base docs
  only where the shipped behavior changes.
- No public API, schema, architecture, runtime, or dependency documentation change
  is expected.
- Reconcile validation matrix, changelog, backlog, and project context after code
  verification. No ADR is required because the fixes enforce existing contracts.

## Verification Evidence

- RED: focused build first failed because `WalletSelectionPolicy`, canonical
  transaction signs, secret lifecycle policy, grouped VND normalization, and
  metric masking state did not exist. The pre-existing grouped-field test then
  exposed the old behavior that kept pasted `1.234` as a noncanonical draft.
- GREEN: focused `DesignSystemTests`, `WalletsReferenceTests`, and
  `BillsDebtTravelReferenceTests` passed in
  `/tmp/mypocket-ui-behavior-green2.xcresult`.
- PASS: signed full `MyPocketIOSTests` executed 415 tests, zero failures/skips
  and zero runtime warnings in
  `/tmp/mypocket-ui-behavior-full-signed.xcresult`.
- PASS: Release simulator build with `SWIFT_TREAT_WARNINGS_AS_ERRORS=YES` at
  `/tmp/mypocket-ui-behavior-release`.
- PASS: `mobile-ios/scripts/verify-project.sh`,
  `mobile-ios/scripts/check-base-components.test.sh`, and `git diff --check`.
- An unsigned full run produced only the expected simulator Keychain entitlement
  failures (`-34018`/`keyUnavailable`); the repository-standard signed rerun is
  the authoritative result above.

## Reconciliation

The account, wallet, portfolio, transaction, and shared-base design contracts
now describe the shipped behavior. Validation, backlog, context, and changelog
are updated. Public API/OpenAPI, schema, architecture, deployment, and external
user setup docs do not change because this batch is native presentation/state
only. No ADR is needed.

## Trace Links

- [Backlog](../BACKLOG.md)
- [Validation matrix](../VALIDATION_MATRIX.md)
- [Account settings design](../../design/pages/account-settings/README.md)
- [Wallet design](../../design/pages/wallets/README.md)
- [Portfolio design](../../design/pages/portfolio/README.md)
- [Transaction design](../../design/pages/transactions/README.md)
- [Shared bases](../../design/system/BASE_COMPONENTS.md)
- [Changelog](../../releases/CHANGELOG.md)
