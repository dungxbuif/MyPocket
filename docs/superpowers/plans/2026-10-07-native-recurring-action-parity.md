# Native recurring action parity

## Intent

Match the native recurring schedule action anatomy to the web
`RecurringPanel`. The web keeps the route content in a shared card with a
compact `Thêm` action next to the section title; native currently puts an
icon-only plus in the page header.

## Contract

- Keep schedule loading, run-due, CRUD, validation, timezone and owner scope
  unchanged.
- Make the screen header navigation-only and render `Thêm` in the existing
  recurring summary card using shared `MPButton`.
- Preserve the same editor callback and expose stable ID `recurring-add`.
- No backend/API/schema/migration/permission changes.

## Verification plan

- Add the copy/accessibility contract test before implementation and capture a
  red compile.
- Implement the shared-base action placement, run focused/full native tests.
- Build Release with warnings as errors and run project/base-component/diff
  guards.
- Reconcile validation, platform design, audit and changelog docs.

## Status

- [x] Contract test added and red run captured
- [x] Native action placement implemented with shared base
- [x] Native tests, Release build and guards pass
- [x] Documentation reconciled

## Final evidence

- RED compile evidence: `/tmp/mypocket-recurring-action-red`.
- Focused GREEN contract test: `/tmp/mypocket-recurring-action-unit-final/Logs/Test/`
  (1/1 passed).
- Full native unit/contract target at that historical action-parity checkpoint:
  `/tmp/mypocket-recurring-action-full-final/Logs/Test/` (105/105 passed);
  the current suite is 129/129.
- Release warnings-as-errors build:
  `/tmp/mypocket-recurring-action-release-final` (succeeded).
- Project verification, base-component guard/tests and `git diff --check` all
  passed.
- Disposable API-backed create/update/delete/run-due UI flow passes 1/1 at
  `/tmp/mypocket-recurring-ui-final2.xcresult`; the test targets `recurring-add`
  and asserts the current `Thêm` label.
