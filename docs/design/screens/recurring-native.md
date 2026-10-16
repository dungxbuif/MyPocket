# Recurring native screen

## Reference contract

The Recurring route keeps the existing create/edit/delete/run-due semantics and
renders populated schedules as a compact grouped list. Each row shows the
schedule name, signed domain amount, cadence, wallet, next-run date, optional
note, and active state. Secondary actions are exposed by the shared overflow
menu so the list does not become a stack of oversized cards.

The create/edit route remains the existing dedicated editor. Frequency,
interval, start/end mode, category applicability, timezone, and validation are
unchanged. No schedule is created merely by loading the screen.

## States to verify

- populated expense and income schedules;
- daily/weekly/monthly/yearly cadence and interval greater than one;
- active and paused rows;
- next-run and optional end date;
- create/edit/delete and run-due actions;
- empty, loading, and error states.

## Implementation note

`RecurringView` uses the shared reference surface, transaction glyph, compact
text styles, and `MPOverflowMenu`. The list composition does not introduce a
new per-feature card or change API/model/schema behavior.

## Verification

Focused contracts live in `MyPocketIOSTests/CategoriesRecurringReferenceTests.swift`.
The populated UI must be checked with fixture schedules covering two cadences,
one paused schedule, and one optional note.
