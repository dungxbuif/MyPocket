# Native recurring-transaction scope parity

## Status

`verified_fixed` — 2026-10-06.

## Contract

The native recurring screen must expose the same operational context as the web
screen:

- explain that schedules create ordinary editable transactions and are run in
  the account timezone;
- show type, frequency, wallet, amount, active state, and the next-run date;
- preserve optional notes and end-date controls in the editor;
- require confirmation before deletion because already-created transactions
  remain in the ledger;
- keep the existing server-side catch-up action and shared base controls.

## Implementation plan

Add the missing summary copy, wallet/date/note presentation, and delete
confirmation to `RecurringView`. Add a shared account-timezone date-time
formatter for the next-run label without changing the API contract.

Implemented in `mobile-ios/MyPocketIOS/Features/Secondary/SecondaryFeatureViews.swift`
and `MPFormat.swift`: recurring cards now show localized type/frequency,
wallet, amount, next run, note, and active state; the shared confirmation
dialog protects deletion; and the screen explains account-timezone execution
and ordinary-transaction semantics.

## Verification plan

- native unit tests and date-format coverage;
- Release build with warnings-as-errors;
- base-component and project checks;
- simulator authenticated More-route smoke test reaches the recurring screen.

Verification: `MyPocketIOSTests` 55/55 passed, including account-timezone
next-run formatting; Release warnings-as-errors, base-component/project
guards, and `git diff --check` passed; and the authenticated More-route sweep
passed 1/1 on the iPhone 16 simulator in 124.207s.
