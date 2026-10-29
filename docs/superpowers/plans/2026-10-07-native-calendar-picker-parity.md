# Native calendar picker parity

Date: 2026-10-07  
Scope: shared native date controls used by transaction, wallet, budget, jar,
recurring, travel, portfolio and month flows.

## Intended behavior

The Stitch transaction reference exposes a calendar modal from the date row.
Native `MPDateStepper` must keep the existing previous/next actions while
making the date label open a reusable calendar surface. Standalone date fields
used by wallet, portfolio, transfer and transaction custom-range forms must
open that same surface rather than a platform compact picker. Selecting a day
and confirming updates the bound screen date; cancelling leaves it unchanged.
Every consumer supplies the account timezone and keeps the existing date
formatting contract.

## Technical decision

- Add `MPDateCalendarSheet` and the date-button affordance to the existing
  DesignSystem layer in `MPFields.swift`.
- Reuse `MPButton`, `MPSurfaceCard`, tokenized colors and base hit targets; the
  feature screens only provide the date and callbacks.
- Use a shared calendar-grid model with Monday-first Vietnamese weekday labels,
  adjacent-month day cells, month navigation and account-timezone date-only
  selection. The native platform picker is not used because its chrome does
  not match the supplied Stitch calendar reference.
- No API, schema, timezone or persistence changes.

## Verification plan

1. Add pure calendar-grid contracts for month title, six-week cells and
   account-timezone date selection.
2. Add a UI contract that opens the calendar from the transaction editor,
   exposes the shared modal title and supports cancel/confirm without leaving
   the form.
3. Run the full native unit target, Release warnings-as-errors build,
   project/base-component/diff guards and web design/build checks.
3. Reconcile the base-component inventory, validation matrix and changelog.

## Scope boundary

This adds the shared calendar entry point. Date validation and account-local
serialization remain owned by the existing feature forms.

## Verification result

- Focused transaction-editor calendar UI flow: PASS, 1/1 at
  `/tmp/mypocket-calendar-custom-ui2.xcresult` (open, cancel, reopen and
  confirm; screenshot attachment exported under
  `/tmp/mypocket-calendar-custom-ui2-attachments/`).
- API-backed ledger custom-range calendar flow: PASS, 1/1 at
  `/tmp/mypocket-calendar-range-ui.xcresult` (switch to custom period, open
  start date, cancel and return to the range form).
- Standalone `MPDateField` consumers now open the same calendar base for goal
  wallet, transfer/adjustment, portfolio and transaction custom-range
  date-only inputs; the transaction range passes the model's account timezone
  instead of silently using the device timezone.
- Full native unit/contract target: PASS, 121/121 at
  `/tmp/mypocket-native-calendar-field-final-unit.xcresult`.
- Release warnings-as-errors build: PASS at
  `/tmp/mypocket-native-calendar-field-final-release`.
- Project, base-component and diff guards: PASS.
