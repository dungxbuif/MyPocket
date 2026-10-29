# Native Travel Mode action parity

## Intent

Match native Travel Mode to the web `TravelModePanel`: keep the page header
navigation-only and place the compact `Thêm` action in the summary card beside
the Travel Mode title.

## Contract

- Preserve event loading, active toggle, CRUD, validation and owner scope.
- Use the shared primary `MPButton` with copy `Thêm` and stable ID
  `travel-add` in the existing summary card.
- No backend/API/schema/migration/permission changes.

## Verification plan

- Add the copy/accessibility contract test and capture RED.
- Implement action placement, run focused/full native tests.
- Build Release with warnings as errors and run project/base-component/diff
  guards.
- Reconcile validation, platform design, audit and changelog docs.

## Status

- [x] Contract test added and red run captured
- [x] Native action placement implemented with shared base
- [x] Native tests, Release build and guards pass
- [x] Documentation reconciled

## Final evidence

- RED compile evidence: `/tmp/mypocket-travel-action-red`.
- Focused GREEN contract test: `/tmp/mypocket-travel-action-unit-final/Logs/Test/`
  (1/1 passed).
- Full native unit/contract target: `/tmp/mypocket-travel-action-full-final/Logs/Test/`
  (106/106 passed).
- Release warnings-as-errors build:
  `/tmp/mypocket-travel-action-release-final` (succeeded).
- Project verification, base-component guard/tests and `git diff --check` all
  passed.
