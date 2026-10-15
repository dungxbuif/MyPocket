# Native credit summary-card parity

## Intent

Match the native credit ledger summary to the web `CreditWalletPanel`: show
current debt and available credit as the primary pair, then limit and payment
status/due-date details in one shared reference card instead of a five-cell
metric grid.

## Contract

- Preserve the selected owner-scoped credit wallet, statement endpoint,
  entry/payment actions, empty statement copy and mutation payloads.
- Compose `MPSurfaceCard(.reference)`, `MPText`, `MPMetricCard`/shared text
  roles, `Divider`, `MPButton` and existing selector/form bases only.
- Keep the web action copy and stable IDs for adding a credit entry and paying
  the card; no API, schema, migration or permission change.
- Preserve explicit `not_configured` guidance and account-timezone due-date
  formatting.

## Verification plan

- Add a summary/action copy contract test and capture RED.
- Implement the shared reference card and run focused/full native tests.
- Run web design/type/build checks, Release warnings-as-errors, project/base
  guards and the bounded credit UI smoke when the disposable API is running.
- Reconcile iOS renderer, validation matrix, port audit, context and changelog.

## Status

- [x] Contract test added and red run captured
- [x] Native summary/action anatomy implemented with shared bases
- [x] Native unit, Release and guards/web checks pass
- [x] Documentation reconciled

## Final evidence

- RED compile evidence: `/tmp/mypocket-credit-summary-red`.
- Focused GREEN summary/action contract: `/tmp/mypocket-credit-summary-unit-final`
  (1/1).
- Full native unit/contract target: `/tmp/mypocket-credit-summary-full-final`
  (108/108).
- Follow-up section-order/empty-state contract: `/tmp/mypocket-credit-order-green`
  (1/1); full native rerun: `/tmp/mypocket-credit-order-full` (108/108).
- Follow-up Release warnings-as-errors build: `/tmp/mypocket-credit-order-release`
  (`BUILD SUCCEEDED`).
- Project verification, base-component guard/tests, `git diff --check`, web
  design checks, typecheck and Vite build all pass.
- Bounded authenticated credit API UI smoke passed 1/1 at
  `/tmp/mypocket-credit-order-ui-final/Logs/Test/Test-MyPocketIOS-2026.10.07_15-34-55-+0700.xcresult` after the section-order fix.
  Authenticated visual UAT remains open; no backend/API/schema/migration or
  production mutation was performed.
