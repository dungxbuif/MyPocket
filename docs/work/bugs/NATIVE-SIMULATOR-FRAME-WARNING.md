---
artifact_type: bug
id: NATIVE-SIMULATOR-FRAME-WARNING
status: fixed
updated: 2026-10-06
---

# Simulator mutation forms emit invalid-frame warnings

## Observation

The API-backed iPhone 16 Simulator suite passes, but XCTest records
`Invalid frame dimension (negative or non-finite)` while scrolling to the save
button in the credit entry/payment, transfer, wallet creation and ordinary
transaction mutation flows. Each affected flow has an active numeric keyboard
from `MPAmountField`; non-mutation route tests do not emit the warning.

## Reproduction

Run the `LOCAL_API_UI` suite with the disposable local API and inspect the
XCTest result bundle runtime warnings. The warning is reported during
`Scroll element to visible` immediately before tapping a save button.

## Rejected hypothesis: keyboard toolbar

The shared number-field keyboard toolbar (`ToolbarItemGroup` containing a
`Spacer` and the dismiss button) was the first common layout boundary while the
scroll view was resized for the keyboard. Replacing that group with one
bounded keyboard toolbar item left the warning present in all four targeted
tests, so the experiment was reverted without changing amount-entry behavior.

## Next hypothesis

`MPButton` currently asks the label's inner `HStack` to be both at least a
fixed width and infinitely wide. This base component is present at each
mutation save boundary, including the single-button wallet and transaction
editors. Move the width constraint to the outer button label container so the
label receives one bounded proposal while retaining the same full-width visual
contract, then rerun the wallet mutation test and inspect the simulator log.

That one-frame variant preserved the UI and removed the warning on the edit
save path, but the same warning still occurred when the create form was
scrolled to its save button. It is therefore not a sufficient fix and was
reverted.

## Next hypothesis

The warning may be emitted by SwiftUI's keyboard toolbar layout itself while
XCTest scrolls a form whose numeric field is first responder. Temporarily
remove the shared number-field toolbar, rerun the same wallet test, and use the
simulator log to distinguish a toolbar issue from a general scroll/layout issue.

Removing the toolbar makes the warning disappear in the wallet create/update
test, confirming that the fault is in SwiftUI's `.keyboard` toolbar layout on
this simulator runtime rather than the form's numeric value. The final fix
must therefore retain the `Xong` keyboard action without using SwiftUI's
keyboard-toolbar placement.

## Final fix design

Render the shared amount field with a small UIKit-backed `UITextField` base
component. It keeps the existing binding, accessibility label, number-pad
keyboard and rounded-border presentation, but installs `Xong` as a UIKit
`inputAccessoryView`. This avoids the SwiftUI toolbar layout path that emitted
the invalid frame while preserving the user-facing behavior.

## Verification

- The UIKit bridge compiles under Swift 6 and keeps the base amount-field
  contract.
- The wallet create/update UI test finds and taps `Xong` on the iPhone 16
  Simulator.
- The four affected mutation tests pass 4/4 in 95.884 seconds.
- The simulator log shows no new `Invalid frame dimension` fault during that
  targeted run.
