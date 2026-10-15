---
artifact_type: bug
id: NATIVE-REPORT-INVALID-MONTH
status: done
owner: shared
---

# Native Reports must reject invalid month scopes without stale content

## Evidence

The native Reports screen accepted an input such as `2026-99`, sent it to both
report endpoints, then kept the previous summary and rendered zero/old trend
content while hiding the request error. This was reproduced on the disposable
local API fixture; it was not a production or database finding.

## Bounded fix

- Validate `YYYY-MM` and the calendar month before any network request.
- On validation or load failure, clear summary, month detail and trend rows so a
  previous month is never shown under a new scope.
- Reuse the existing base `MPErrorView`/retry state and keep the editable month.
- No API, schema, cache, or environment changes.

## Verification

- Pure unit tests cover valid months, malformed input, and out-of-range months.
- Focused native unit and authenticated UI regressions pass at
  `/tmp/mypocket-report-month-unit-final-20261015.xcresult` and
  `/tmp/mypocket-report-month-ui-final-20261015.xcresult`; the Debug simulator
  build and base/project guards pass with warnings treated as errors.
- Backend `go test ./...` passes across all packages at
  `/tmp/mypocket-backend-regression-20261015.log`; this change does not alter
  the backend contract.
- An unsigned full native attempt reproduced four unrelated simulator Keychain
  failures (`-34018` / `keyUnavailable`) in existing API/outbox tests. A signed
  rerun passed the complete 294-test native suite, confirming the failures are
  signing-mode environment behavior and not Reports/product regressions.
- The existing API-backed UI fixture remains the end-to-end proof when its
  PostgreSQL dependency is available; no environment is started by this task.

## Result

`ReportsMonthPolicy` now rejects malformed/out-of-range months before any API
call, and failed/invalid loads clear the previous report scope. Focused unit and
local authenticated UI regressions pass, and the Debug simulator build plus
guards pass. Production and PostgreSQL-backed UI remain environment gates
outside this code-only fix; the full native suite still has the four pre-existing
simulator Keychain failures listed above.
