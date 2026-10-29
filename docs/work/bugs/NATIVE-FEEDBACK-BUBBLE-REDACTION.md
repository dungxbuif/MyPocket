---
artifact_type: bug
id: NATIVE-FEEDBACK-BUBBLE-REDACTION
status: regression_verified
owner: shared
trace:
  backlog: ../BACKLOG.md
  phase: ../phases/PHASE-002-native-ios-swiftui.md
  validation: ../VALIDATION_MATRIX.md
  release_notes: ../../releases/CHANGELOG.md
---

# Native feedback bubble capture redaction

## Reproduction

Open the floating feedback bubble on an authenticated native screen. The
current tap handler captures the key window immediately, then presents the
feedback overlay. Because SwiftUI has not rendered the state transition yet,
the bubble can be present in the screenshot that is attached to the feedback.

## Impact

The screenshot-assisted feedback contract promises the screen behind the
feedback UI. Including the trigger bubble is noisy and can expose an overlay
control that should not be part of the report context.

## Root cause

The original tap handler called `ScreenSnapshot.capture()` in the same SwiftUI
state mutation that presented the feedback overlay. The window was therefore
captured before the trigger's removal had been rendered.

## Intended behavior

The bubble must be hidden for one rendered frame before `ScreenSnapshot.capture`
runs. The feedback overlay can then present with the captured background image;
text-only fallback and the PNG/1 MiB backend contract remain unchanged.

## Regression risk

The delayed capture must not create a second trigger, lose the overlay action,
or leave the bubble hidden after the sheet is dismissed. The dedicated UI
regression checks the trigger/sheet label cardinality; the existing PNG,
multipart and full simulator tests cover the attachment and route behavior.

## Verification result — 2026-10-06

`MainTabView` now hides the trigger for one rendered frame before capturing the
window, then presents the feedback overlay with that underlying-screen image.
The simulator UI regression taps the bubble and verifies that exactly one
`Gửi phản hồi` button remains while the overlay is visible (the sheet action,
not the floating trigger). The full signed `LOCAL_API_UI` suite passes 16/16,
native XCTest passes 65/65, the Release warnings-as-errors build and
base/project guards pass, and the simulator log has no invalid-frame fault.
