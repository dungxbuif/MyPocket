---
artifact_type: bug
id: NATIVE-MONTH-SUMMARY-SCOPE
status: regression_verified
created: 2026-10-06
updated: 2026-10-06
---

# Native month-summary scope

## Finding

The native Tháng screen used a free-form `YYYY-MM` field and omitted web
summary fields (`is_current`, `is_complete`, period boundaries and
`jar_summary`). It also lacked bounded month navigation, dirty-note protection
and a route action to open Hũ chi tiêu for the selected month.

## Intended behavior

- Select a valid account-local month from the shared bounded options.
- Show period status/boundaries, income/expense/net/count, category counts,
  and jar usage when the API returns it.
- Preserve unsaved note text when a save fails and confirm before discarding it
  while changing month.
- Use a clear month-to-jars action while keeping ledger balances unchanged.

## Verification plan

1. Extend the typed month summary model with the backend `jar_summary` field.
2. Replace free-form month entry with shared selector/base cards and wire
   month-to-jars navigation via the existing More route callback.
3. Reuse bounded month-option coverage and run native XCTest, Release build and
   signed simulator route coverage; reconcile all docs.

No production data or provider credentials are used for this fix.

## Implementation result

- `MonthSummaryModel` now decodes the backend `jar_summary` payload.
- `MonthsView` uses the account-local bounded month selector, renders period
  status/boundaries, category counts, jar usage, note save/clear/reset actions,
  and opens the selected month in Hũ chi tiêu through the existing route.
- Month changes with dirty notes require explicit confirmation; failed saves
  keep the draft in place.

## Verification result — 2026-10-06

- PASS: native XCTest is 54/54 and the existing bounded month-option contract
  remains green.
- PASS: Release simulator build succeeds with
  `SWIFT_TREAT_WARNINGS_AS_ERRORS=YES`.
- PASS: signed iPhone 16 More-route sweep is 1/1 in 122.295s, including Tháng
  and Hũ chi tiêu navigation.
- PASS: no production data or provider credential was used.
