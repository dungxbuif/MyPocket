# Native feedback bubble right-side dock

## Problem

The feedback trigger was positioned at the lower-left edge of the shell
viewport. On the simulator this made the content feel shifted left and did not
match the intended floating-action placement.

## Intended behavior

- Dock the 52pt feedback bubble at the lower-right edge of the shell viewport.
- Keep the position stable; obstacles do not move it to another row or side.
- Keep it as an overlay only: it must not reserve a scroll lane or change the
  content viewport width.
- Preserve footer/keyboard clearance and hide it while feedback is captured or
  presented.

## Scope and verification

This is a native shared layout/base-component change only. No API, storage,
permissions, or production configuration changes are required. Update the
geometry unit contract and run the focused feedback UI checks plus the signed
native suite.

## Reconciliation — 2026-10-15

`MPFeedbackBubbleLayout.frame` now computes the x-coordinate from the
viewport's trailing edge (`maxX - gutter - side`). It still ignores obstacles
for placement and uses only footer clearance for the y-coordinate, so the
trigger remains an overlay and the scroll content width is unchanged.

Verification passed:

- signed simulator focused run: 1 geometry unit test and 3 feedback UI tests,
  0 failures, artifact at
  `/tmp/mypocket-feedback-right-dock-20261015/Logs/Test/Test-MyPocketIOS-2026.10.15_20-20-29-+0700.xcresult`;
- the UI checks confirm the bubble remains inside the full viewport, does not
  cover Account security text, More/category actions, or reserve a scroll
  lane;
- prior signed full native regression before this follow-up was green at
  `/tmp/mypocket-full-native-account-add-disabled-20261015/Logs/Test/Test-MyPocketIOS-2026.10.15_20-06-42-+0700.xcresult`;
- base-component guards, project verification and `git diff --check` remain
  required before integration.

Continuation probe — 2026-10-15 20:46 (iPhone 16 simulator):

- `DesignSystemTests/testFloatingFeedbackUsesStableLowerRightDockAndNeverReservesObstacleLane`
  passed with `xcodebuild test` (RC 0) on simulator
  `EC80B092-CEF1-4A8E-A5A3-7420298EFD4D`.
- The contract still computes the bubble at the trailing edge of the full
  content viewport and returns the same frame when obstacle rectangles are
  supplied; no viewport-width reduction or left-side relocation was observed
  in this probe.
- This is a layout-contract result only. It does not claim a signed device
  install or production rollout; those remain open in the reconciliation
  matrix.

No production, device, API, migration, or runtime configuration was changed.
