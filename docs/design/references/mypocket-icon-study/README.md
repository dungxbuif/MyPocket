# MyPocket icon study

The native catalog uses stable `MPIconID` semantic roles and the
`MPIconStyle.regular` / `.duotone` presentation boundary. The current
implementation intentionally ships SF Symbols as the verified fallback:
`MPIconCatalog.usesVerifiedAssets == false` until a pinned vector subset is
reviewed and added to the iOS asset catalog.

The visual direction is Money Lover-inspired in density and category breadth,
with Phosphor-style regular/duotone roles as an implementation target. No
Money Lover assets, fonts, dependency, or proprietary artwork are bundled.
Persisted API icon keys remain unchanged and unknown keys resolve to the
visible `tag` fallback.

## Provenance and license

This study contains semantic mappings and platform fallbacks only. It does
not redistribute Phosphor or Money Lover artwork. If verified Phosphor vectors
are added later, record the exact package version, selected asset paths, and
license notice beside the asset catalog before setting `usesVerifiedAssets` to
true.

See [comparison.html](comparison.html) for the compact role matrix.
