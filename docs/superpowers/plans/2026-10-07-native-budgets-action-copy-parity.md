# Native budgets action copy parity

## Intent

Align the native Budgets entry action with the existing web `BudgetsPanel`
contract. The web route exposes a compact `Thêm` action in the shared screen
header; native currently exposes an icon-only plus control with different copy.

## Contract

- Keep budget loading, owner scope, validation, CRUD and editor behavior
  unchanged.
- Expose `Thêm` in the native header using the shared `MPButton` base with a
  compact non-full-width treatment; do not create screen-local button geometry.
- Preserve the existing route and accessibility behavior while adding a stable
  identifier for the action.
- No backend/API/schema/migration/permission changes.

## Verification plan

- Add a focused copy/accessibility contract test before implementation and run
  it red.
- Implement the shared-base header action, run the focused test green, then the
  complete native unit target.
- Build Release with warnings as errors and run project/base-component/diff
  guards.
- Reconcile validation, platform design, audit and changelog docs.

## Status

- [x] Contract test added and red run captured
- [x] Native header action implemented with shared base
- [x] Native tests, Release build and guards pass
- [x] Documentation reconciled

## Final evidence

- RED compile evidence: `/tmp/mypocket-budgets-action-red`.
- Focused GREEN contract test: `/tmp/mypocket-budgets-action-unit-final/Logs/Test/`
  (1/1 passed).
- Full native unit/contract target: `/tmp/mypocket-budgets-action-full-final/Logs/Test/`
  (103/103 passed).
- Release warnings-as-errors build:
  `/tmp/mypocket-budgets-action-release-final` (succeeded).
- Project verification, base-component guard/tests and `git diff --check` all
  passed.
