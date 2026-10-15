# Native bottom add affordance when disabled

## Problem

The compact shell navigation still instantiates `MPBottomAddButton` when the
current tab disables the global add action (Account and Assistant). Hiding the
button with opacity leaves the raised circle in the layout/visual hierarchy,
which can appear as a dim circle over the last row or other content.

## Intended behavior

- When `isAddEnabled == true`, render the shared raised add button exactly as
  before, including its accessibility identifier and long-press behavior.
- When `isAddEnabled == false`, do not instantiate the button at all. The shell
  must not reserve a visual/hit-test affordance for an unavailable action.
- Account-specific add actions remain owned by the Account screen and are not
  changed by this fix.

## Scope and decision

This is a base-component-only SwiftUI fix in `MPBottomNavigation`; no API,
storage, navigation route, or production configuration changes are needed.

## Verification

- Focused native tests assert the global add button remains available on
  Transactions and is absent/disabled on Account.
- Run base-component guards, project verification, diff checks, and the native
  focused test suite. Reconcile the validation matrix with the result.

## Reconciliation — 2026-10-15

Implemented in `MPBottomNavigation`: the compact shell now conditionally
creates `MPBottomAddButton` only when `isAddEnabled` is true. Account and
Assistant therefore have no disabled/transparent raised circle in the view
tree, while Transactions keeps the existing add and long-press behavior.

Verification passed:

- `mobile-ios/scripts/check-base-components.sh`
- `mobile-ios/scripts/check-base-components.test.sh`
- `mobile-ios/scripts/verify-project.sh`
- `git diff --check`
- signed simulator focused run: 90 `DesignSystemTests` plus
  `LaunchTests.testGlobalAddLongPressOpensAIEntry`, 0 failures, artifact at
  `/tmp/mypocket-account-add-disabled-20261015/Logs/Test/Test-MyPocketIOS-2026.10.15_20-05-54-+0700.xcresult`

No production, device, API, migration, or runtime configuration was changed.
