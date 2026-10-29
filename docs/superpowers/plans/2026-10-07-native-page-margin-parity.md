# Native page-margin parity

## Intent

Align the shared SwiftUI page inset with the supplied Stitch token contract.
The reference defines a 16px container margin; the native renderer currently
uses a 20pt screen inset, which makes every route wider than the web/reference
surface.

## Affected contract

- `mobile-ios/MyPocketIOS/DesignSystem/Tokens.swift` (`MPMetrics.pagePadding`)
- All native screens that already compose `MPMetrics.pagePadding`
- No API, persistence, routing or business-rule changes

## Verification

1. Add a DesignSystem regression assertion for the 16pt page margin.
2. Change only the shared token, leaving screen-local composition untouched.
3. Run native unit/contract tests, the unauthenticated UI smoke, project/base
   guards and the warnings-as-errors Release build.
4. Record the result in the native port audit, validation matrix and changelog.
