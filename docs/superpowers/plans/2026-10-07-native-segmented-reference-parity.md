# Native segmented-control reference parity

## Decision

The Stitch and web `SegmentedControl` contract uses a light line/surface track,
with the active segment rendered as a white pill with action-colored text. The
native `MPSegmentedControl` was incorrectly using the solid green primary
button style, which made transaction type, report period and detail/trend
switches visually heavier than the reference.

The shared button base will own a `segmentedSelected` variant. The shared
segmented control will use that variant for the selected option and a plain
label for inactive options; feature screens will not add local colors or
shapes.

## Scope

- `mobile-ios/MyPocketIOS/DesignSystem/MPButton.swift`
- `mobile-ios/MyPocketIOS/DesignSystem/MPNavigation.swift`
- Native design-system contract tests and validation docs.

No API, schema, navigation or data behavior changes are expected.

## Verification

- Add a pure base-policy test for selected/inactive segment styles.
- Run the focused test and the full native unit target.
- Run Release warnings-as-errors, project/base-component guards and diff
  checks.
- Re-run the transaction-editor simulator screenshot and record the evidence.
