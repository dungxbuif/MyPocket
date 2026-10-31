# MyPocket icon study

MyPocket ships a small, local-only subset of Phosphor Icons regular 2.1.1 in
`mobile-ios/MyPocketIOS/Assets.xcassets/Icons/`. The semantic catalog in
`MPIconCatalog` maps the existing API/system keys to these assets without
changing persisted keys. Unknown keys still resolve to the visible tag role.

The checked-in [manifest.json](manifest.json) is the source-of-truth inventory:
it pins the upstream package/version/license, maps every `MPIconID` to its
bundled vector, and names the 20 roles shown in the comparison board. The
read-only guard `mobile-ios/scripts/check-icon-catalog.sh` verifies the
manifest, vector-preserving asset entries, tintable SVGs, project resource
inclusion, license, and board coverage.

The subset is intentionally regular-weight and tintable. The `duotone` API
style remains available for future layered assets, but currently reuses the
regular asset so the app does not mix unverified weights. There is no runtime
network fetch and the shipped license is
`mobile-ios/MyPocketIOS/Resources/Phosphor-LICENSE.txt`.

Reference board: [comparison.html](comparison.html). It records the current SF
Symbol fallback name, Phosphor Regular local preview, the unbundled Duotone
state, and the Lucide comparison name at 20/24/32pt beside Vietnamese sample
text on light/dark surfaces. The Duotone choice and final whole-screen
legibility remain owner visual/device approval items; this artifact does not
claim that approval.
