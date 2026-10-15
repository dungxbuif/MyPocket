# Native overview wallet action copy parity implementation plan

## Goal

Match the current web Overview wallet-card action label exactly while keeping
the existing owner-scoped wallet-management route.

## Contract

- The `Ví của tôi` section uses the shared `MPSectionHeader` base.
- Its trailing action label is `Quản lý`, matching the web Overview panel.
- Tapping the action still invokes the existing native wallet-management
  callback; no API, schema, permission or route contract changes.

## Verification

- [x] Add a pure copy contract before implementation.
- [x] Run the focused and full native unit suites, Release warnings-as-errors,
  project/base-component guards and `git diff --check`.
- [x] Reconcile validation/audit/iOS docs and changelog.

## Shipped evidence

- RED compile: `/tmp/mypocket-overview-wallet-action-red` (missing
  `OverviewReferenceCopy`, expected before implementation).
- Focused GREEN unit: `/tmp/mypocket-overview-wallet-action-unit/Logs/Test/`
  (1/1).
- Full native unit suite:
  `/tmp/mypocket-overview-wallet-action-full/Logs/Test/`
  (99/99, 0 failures).
- Release warnings-as-errors build: `/tmp/mypocket-overview-wallet-action-release`.
- Guards: `verify-project.sh`, `check-base-components.test.sh` and
  `git diff --check` pass.
