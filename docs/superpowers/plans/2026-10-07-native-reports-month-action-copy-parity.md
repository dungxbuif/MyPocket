# Native reports month action copy parity implementation plan

## Goal

Match the web Reports summary-card trailing label: the selected account-local
month is displayed as the card action label, while the explicit report button
continues to own reload behavior.

## Contract

- The native `Tổng quan báo cáo` card uses the shared `MPSectionHeader` base.
- Its trailing label is the selected `month` value, matching the web
  `SectionTitle action={month}` contract.
- The existing `Xem báo cáo` button remains the only explicit reload action in
  the filter surface; API, month parsing, owner scope and report data remain
  unchanged.

## Verification

- [x] Add a pure copy contract before implementation.
- [x] Run focused and full native unit suites, Release warnings-as-errors,
  project/base-component guards and `git diff --check`.
- [x] Reconcile validation/audit/iOS docs and changelog.

## Shipped evidence

- RED contract: `/tmp/mypocket-reports-month-action-red-test` failed as
  expected before `ReportsReferenceCopy` existed.
- Focused GREEN contract: `/tmp/mypocket-reports-month-action-unit/Logs/Test/`
  (1/1 passed).
- Full native unit suite:
  `/tmp/mypocket-reports-month-action-full/Logs/Test/` (100/100 passed).
- Release warnings-as-errors build:
  `/tmp/mypocket-reports-month-action-release` (`BUILD SUCCEEDED`).
- `verify-project.sh`, `check-base-components.test.sh` and `git diff --check`
  all pass.
