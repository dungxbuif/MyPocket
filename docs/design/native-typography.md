# Native typography decision

The iOS surface uses the system San Francisco face through the shared `MPFont` token and UIKit field helpers. This matches native iOS controls and the supplied device reference more closely than the previously bundled Manrope face, especially for Vietnamese diacritics and Dynamic Type metrics.

The change is intentionally centralized in `mobile-ios/MyPocketIOS/DesignSystem/Tokens.swift` and the shared amount field in `MPFields.swift`; feature screens must continue to use `MPFont` rather than adding local font names.
