---
artifact_type: bug
id: NATIVE-JARS-MONTH-SCOPE
status: regression_verified
created: 2026-10-06
updated: 2026-10-06
---

# Native spending-jars month and cumulative scope

## Finding

The native Hũ chi tiêu screen exposed only a free-form month text field and a
single allocation metric. The web contract provides a bounded month selector,
actual/unassigned spending metrics, over-allocation warnings, active/history
semantics, and a selected-jar cumulative report.

## Intended behavior

- Select a valid month from a bounded account-local list and reload on change.
- Show actual income, total spent, allocated amount, and unassigned spending;
  warn when allocations exceed income or 100%.
- Render per-jar allocation progress, with edit/remove controls only for active
  month configurations.
- Select a jar and load its cumulative month report without changing ledger
  balances.

## Implementation result

- `MPFormat.monthOptions` now provides an account-local bounded month list.
- `JarsView` now renders shared selector/metric/status/card primitives for
  actual income, spending, allocation, unassigned spending and advisory
  over-allocation warnings.
- Active month rows expose edit/remove actions; historical rows remain
  read-only. The cumulative report is scoped to a selected jar and clears when
  the selected month has no jar data, so stale report state cannot leak across
  month changes.

## Verification plan

1. Add shared month-option generation and its unit coverage.
2. Port the summary, warning, active/history, and cumulative-report states to
   `JarsView` using existing base components.
3. Run native unit tests, Release warnings-as-errors build, signed simulator
   route coverage, and reconcile this note, matrix, context, and changelog.

## Verification result — 2026-10-06

- PASS: native XCTest is 50/50, including bounded month-option coverage.
- PASS: Release simulator build succeeds with `SWIFT_TREAT_WARNINGS_AS_ERRORS=YES`.
- PASS: signed iPhone 16 simulator More-route sweep is 1/1 in 123.999s after
  the Jars change.
- PASS: the local Stage v1 API remained healthy; no production data or provider
  credential was used.
- NOTE: the previously verified full API-backed UI suite remains 14/14; this
  change was covered by the targeted route sweep because the full post-change
  diagnostic service run was not repeated.

No production data or provider credentials are used for this fix.
