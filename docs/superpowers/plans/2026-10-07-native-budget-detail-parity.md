# Native budget detail parity implementation plan

## Goal

Match the implemented web `BudgetsPanel` behavior in native iOS without
changing the budget API, date semantics or owner scope.

## Contract

The native budget screen must expose the same real response data as the web:

- summary spent, limit, remaining and overall progress;
- each budget's date/status copy and remaining-day count;
- daily spending suggestion while a budget is active;
- 75% warning and over-budget warning;
- existing create/update/delete controls and privacy masking.

All presentation must use existing `MP*` bases or a new shared base; no
screen-local colors or raw controls are allowed.

## Verification

- [x] Add pure unit coverage for timing, daily suggestion and threshold copy.
- [x] Keep the existing API-backed budget CRUD UI test and assert the new
  status surfaces on an active fixture.
- [x] Run the full native unit suite, focused budget UI, Release
  warnings-as-errors build, project/base-component guards and diff checks.
- [x] Reconcile `VALIDATION_MATRIX.md`, `PORT-100-AUDIT.md`, iOS design docs
  and `CHANGELOG.md`; no API or migration changes were made.

## Shipped evidence

- `MPBudgetCopy` centralizes timing, daily-suggestion and 75%/over-budget
  copy; `BudgetsView` now composes the shared reference summary card,
  `MPProgressBar`, `MPMetricCard` and `MPStatusMessage` bases.
- Focused design-system contract: 23/23 at
  `/tmp/mypocket-budget-detail-unit/Logs/Test/`.
- API-backed budget CRUD/status UI: 1/1 at
  `/tmp/mypocket-budget-detail-ui/Logs/Test/` on `MyPocket-iPhone-16`.
- Full native unit/contract suite: 94/94 at
  `/tmp/mypocket-budget-detail-full/Logs/Test/`.
- Release warnings-as-errors build: `/tmp/mypocket-budget-detail-release`.
- `verify-project.sh`, `check-base-components.test.sh` and `git diff --check`
  pass. No backend/API/database migration changed.
