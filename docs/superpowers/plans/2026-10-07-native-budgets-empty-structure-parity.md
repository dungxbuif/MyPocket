# Native budgets empty-structure parity implementation plan

## Goal

Match the web Budgets panel when the account has no budget items: show only the
plain `Chưa có ngân sách.` empty state below the shared screen header. Do not
render the overall summary card until at least one budget exists.

## Contract

- The native screen continues to use the shared `MPEmptyView` base for the
  empty state.
- The summary metrics/progress card is rendered only when `summary.items` is
  non-empty, matching the web conditional branch.
- The create-budget action, loading/error states, CRUD APIs, owner scope and
  non-empty budget status behavior remain unchanged.

## Verification

- [x] Add a pure visibility contract before implementation.
- [x] Run focused and full native unit suites, Release warnings-as-errors,
  project/base-component guards and `git diff --check`.
- [x] Reconcile validation/audit/iOS docs and changelog.

## Shipped evidence

- RED contract: `/tmp/mypocket-budgets-empty-red-test` failed as expected
  before `BudgetReferenceCopy` existed.
- Focused GREEN contract: `/tmp/mypocket-budgets-empty-unit/Logs/Test/`
  (1/1 passed).
- Full native unit target: `/tmp/mypocket-budgets-empty-full/Logs/Test/`
  (101/101 unit tests passed). The combined UI invocation also exposed the
  existing unrelated long-press smoke flake
  `LaunchTests.testGlobalAddLongPressOpensAIEntry`; it does not exercise the
  Budgets branch.
- Release warnings-as-errors build:
  `/tmp/mypocket-budgets-empty-release` (`BUILD SUCCEEDED`).
- `verify-project.sh`, `check-base-components.test.sh` and `git diff --check`
  all pass.
