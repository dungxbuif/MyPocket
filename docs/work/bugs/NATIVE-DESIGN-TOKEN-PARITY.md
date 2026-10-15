# Native design-token parity

## Intent

The native SwiftUI client must use the same primary/action color and form
control treatment as the mounted web client and the attached Stitch reference.
The Simulator previously rendered a green/teal action and native text fields
relied on SwiftUI's default rounded-border styling.

## Decision

- Treat `app/src/ui/theme.css` as the source of truth for shared semantic
  colors.
- Keep the native `DesignSystem` as the only consumer boundary: update
  `MPColor.action` and `MPTextField`/`MPSecureField` rather than patching
  feature screens.
- Match the web control contract with a white surface, the web line token,
  12px-equivalent radius, 48pt control height, and a visible action-colored
  focus ring.
- Debug builds use the existing local API URL and Xcode's normal SwiftUI
  preview/live-debug workflow; Release configuration remains production-only.

## Affected flow/config

- Native login and every form consuming `MPTextField`, `MPSecureField`, or
  `MPAmountField`.
- Native shared buttons, focus indicators, and selected states consuming
  `MPColor.action`.
- `mobile-ios/Config/Debug.xcconfig` remains the hot-reload/debug endpoint;
  no production API or secret is changed.

## Verification plan

1. Add a failing DesignSystem test for the web action token and control
   geometry.
2. Implement the shared token and field styling only in `DesignSystem`.
3. Run the focused XCTest, then the complete native test suite.
4. Build the Debug app for the iPhone 16 Simulator, launch it, and compare
   the login primary button and inputs with the web surface.
5. Run `git diff --check` and record the actual build/test result in the
   validation matrix and release changelog.

## Outcome

Implemented in the shared native DesignSystem. The primary CTA now uses the
Stitch/web brand token (`#4caf50`), action/focus uses the Stitch primary action
token (`#006e1c`), system colors were replaced with the web canvas/card/text
primitives, and the shared text,
secure, amount and editor controls use the web border/padding/focus contract.
Debug Live Preview settings are explicit. The focused and full native test
suites, project verification script, Debug build and Simulator install/launch
passed on 2026-10-06.
