# Native base typography audit

## Intent

Ensure every native base component consumes the bundled Manrope roles rather
than SwiftUI/UIKit system-font defaults. The Stitch/web contract centralizes
display, headline, body, metadata, label and currency roles; direct `.body`,
`.caption`, `.footnote`, `.headline` and `.subheadline` modifiers in bases can
otherwise make screens drift even when feature views use `MPText`.

## Scope

- Replace direct system-font modifiers in `DesignSystem` with `MPFont` roles.
- Give UIKit-backed text fields the matching Manrope UIFont fallback path.
- Preserve all control geometry, accessibility labels, APIs and business logic.
- Keep markdown code blocks monospaced, but use the shared small role for size.

## Verification

1. Add/extend a DesignSystem source guard so feature code remains base-only.
2. Run native unit/contract, auth/add UI smoke, project/base guards and the
   Release warnings-as-errors build.
3. Record the final result in the port audit, validation matrix and changelog.
