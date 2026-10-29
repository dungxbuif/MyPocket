---
artifact_type: bug
id: NATIVE-SAVINGS-SUMMARY-SCOPE
status: regression_verified
created: 2026-10-06
updated: 2026-10-06
---

# Native goal-wallet savings summary scope

## Finding

The native goal-wallet detail rendered a generic balance metric and progress
bar, but did not expose the web savings contract: remaining amount, reached
state, account-local days to the target date, and the explicit target amount.

## Intended behavior

- Keep the shared balance metric and transaction detail.
- Show `ĐÃ ĐẠT MỤC TIÊU` or `CẦN THÊM` with the corresponding VND amount.
- Calculate the remaining/overdue day count from the account timezone and the
  calendar-only target date, not the device instant.
- Reuse the base surface/progress components and keep zero/invalid targets
  safe and non-dividing.

## Verification plan

1. Add a pure account-timezone savings summary helper and boundary tests.
2. Render the helper through existing native base components in goal-wallet
   detail.
3. Run native unit tests, Release warnings-as-errors build, and the signed
   simulator route sweep; reconcile the matrix, context, phase note and
   changelog.

No production data or provider credentials are used for this fix.

## Implementation result

- Added shared `MPSavingsProgress` and account-local calendar-day helpers.
- Goal-wallet detail now renders reached/remaining state, target-day status,
  progress and target amount through existing surface/text/progress components.

## Verification result — 2026-10-06

- PASS: native XCTest is 52/52, including positive, reached, zero-target and
  account-timezone boundary coverage.
- PASS: Release simulator build succeeds with
  `SWIFT_TREAT_WARNINGS_AS_ERRORS=YES`.
- PASS: signed iPhone 16 simulator More-route sweep is 1/1 in 124.194s.
- PASS: no production data or provider credential was used.
