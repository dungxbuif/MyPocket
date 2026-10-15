# Native budget date picker parity

Date: 2026-10-07  
Scope: `BudgetsView` date range editor.

## Intended behavior

The web budget form exposes a shared period selector with date-only pickers.
Native currently uses raw `YYYY-MM-DD` text fields, which is functional but
not the reference mobile interaction. Replace those fields with the shared
account-timezone `MPDateField` calendar base while keeping the existing API
wire values, validation and owner-scoped save/update behavior unchanged.

## Verification plan

- Convert loaded/editing budget date strings to account-local `Date` state and
  serialize them back as `yyyy-MM-dd` at save time.
- Run the native unit target, Release warnings-as-errors build, project/base
  guards and web design guard.
- Reconcile the native audit, validation matrix, architecture and changelog.

## Verification result

- Native unit/contract target: PASS, 121/121 at
  `/tmp/mypocket-budget-date-unit-final.xcresult`.
- Release simulator build with warnings-as-errors: PASS at
  `/tmp/mypocket-budget-date-release-final`.
- API-backed budget create/update/delete UI flow: PASS, 1/1 at
  `/tmp/mypocket-budget-date-ui2.xcresult` against the disposable local API
  fixture. The test now targets the stable `budget-add` identifier rather than
  the localized button label.
- Project, base-component, base-component-test, diff and web design guards:
  PASS.
