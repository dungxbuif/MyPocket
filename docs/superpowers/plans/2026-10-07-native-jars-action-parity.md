# Native jars action parity

## Intent

Match the native Hũ chi tiêu action anatomy to the web `JarManagementPanel`.
The web keeps the page header navigation-only and places the compact `Thêm hũ`
action beside the `Hũ tháng` section heading; native currently puts an
icon-only plus action in the page header.

## Contract

- Keep month loading, jar summary, cumulative report, CRUD and owner scope
  unchanged.
- Make the screen header navigation-only and render `Thêm hũ` beside the
  monthly jar section title using the shared compact `MPButton` base.
- Preserve the same action callback and add stable accessibility ID
  `jar-add-month`.
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

- RED compile evidence: `/tmp/mypocket-jars-action-red`.
- Focused GREEN contract test: `/tmp/mypocket-jars-action-unit-final/Logs/Test/`
  (1/1 passed).
- Full native unit/contract target: `/tmp/mypocket-jars-action-full-final/Logs/Test/`
  (104/104 passed).
- Release warnings-as-errors build:
  `/tmp/mypocket-jars-action-release-final` (succeeded).
- Project verification, base-component guard/tests and `git diff --check` all
  passed.
