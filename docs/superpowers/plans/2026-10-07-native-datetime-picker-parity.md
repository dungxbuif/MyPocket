# Native date-time picker parity

Date: 2026-10-07  
Scope: recurring schedule and AI transaction-entry date/time fields.

## Intended behavior

`MPDateTimeField` must use the same project-owned calendar surface as the
date-only transaction, wallet, transfer, portfolio and budget controls. The
field must retain the selected account-local clock time, allow day and
hour/minute selection, and preserve the existing `Date` binding/API wire
values. No feature screen should expose SwiftUI's compact picker chrome.

## Technical decision

- Extend `MPDateCalendarSheet` with an opt-in time section instead of creating
  a second calendar implementation.
- Keep the existing Monday-first Vietnamese grid, tokenized surface/buttons,
  account timezone and date-only accessibility IDs unchanged.
- Use the platform hour/minute wheel only inside the shared project-owned sheet;
  the sheet owns its label, timezone and hit area so feature screens do not
  style native controls independently.
- No backend, schema, timezone or persistence changes.

## Verification plan

1. Add pure display/selection contracts for account-timezone date+time labels
   and preserving the selected clock when changing calendar days.
2. Run the full native unit target, Release warnings-as-errors build,
   project/base-component/diff guards and web design checks.
3. Reconcile the audit, validation matrix, architecture, base-component
   inventory and changelog.

## Verification result

- `DesignSystemTests` focused target: PASS, 43/43 at
  `/tmp/mypocket-datetime-unit-redgreen.xcresult`, including the
  account-timezone date+time label and clock-preservation contracts.
- AI proposal editor date+time sheet UI flow: PASS, 1/1 at
  `/tmp/mypocket-datetime-ui2.xcresult` with the shared title, time control,
  screenshot attachment under `/tmp/mypocket-datetime-ui2-attachments/` and
  cancel behavior.
- Full native unit/contract target: PASS, 123/123 at
  `/tmp/mypocket-calendar-range-final.xcresult`.
- Release simulator build with warnings-as-errors: PASS at
  `/tmp/mypocket-calendar-range-release`.
- Project, base-component, base-component-test, diff and web design guards:
  PASS. Backend `go test ./...` remains PASS (231 tests), and the web
  production build passes with only the existing Vite chunk-size advisory.
