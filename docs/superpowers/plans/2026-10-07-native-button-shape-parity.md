# Native button-shape parity

## Intent

Match the shared native action-button shape to the web/Stitch contract. Web
`BaseButton` uses a full pill for primary, secondary and destructive actions;
native `MPButton` currently uses the generic 12pt control radius.

## Affected contract

- `mobile-ios/MyPocketIOS/DesignSystem/MPButton.swift`
- `mobile-ios/MyPocketIOS/DesignSystem/Tokens.swift`
- The convenience-label initializer must inherit the base `MPFont.label`
  instead of applying a system `.body` override.
- No changes to row buttons, icon buttons, API payloads or routing

## Verification

1. Add a DesignSystem assertion for the shared pill radius.
2. Apply it only to filled `MPButton` styles; keep `.plain` transparent and
   preserve `MPMenuRow`/`MPIconButton` geometry.
3. Run native unit/contract, launch UI, project/base guards and Release
   warnings-as-errors build.
4. Reconcile audit, validation and changelog docs.
