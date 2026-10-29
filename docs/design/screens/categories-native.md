# Categories native screen

## Reference contract

The Categories route keeps MyPocket's existing category CRUD, wallet scope, and
system-category rules. Its populated state presents a compact reference-style
filter group followed by one shared hierarchy surface. Root categories use the
larger shared glyph; child rows retain the connector and smaller glyph supplied
by `MPCategoryTree`. Search and the inactive toggle remain visible and do not
change the API contract.

## States to verify

- expense/income/debt filter with populated parent/child data;
- wallet scope and search filtering, including no-match;
- archived/inactive visibility;
- system category read-only editor;
- create, edit, archive, and delete preserving existing owner/API behavior.

## Implementation note

`CategoriesView` owns only screen composition. Selection, row geometry, icons,
minimum hit targets, and category hierarchy are provided by the existing
`MPCategoryTree`, `MPCategoryPicker`, `MPOptionSelector`, and surface tokens.
The reference pass uses the app palette and `MPSurfaceCard(style: .grouped)`
for the filter group; no sample financial values are persisted.

## Verification

Focused contracts live in `MyPocketIOSTests/CategoriesRecurringReferenceTests.swift`.
The populated UI must be checked with fixture categories containing at least one
parent and two children plus an archived row.

## Organization reference composition — 2026-10-26

The remaining reference pass moves the filter controls into the shared grouped
surface grammar and retains `MPCategoryTree` as the shared populated hierarchy
surface. `CategoriesReferencePresentation` supplies the geometry and tree
accessibility key consumed by the rendered view. `MPCategoryTree` continues to own hierarchy connectors and
selection; `CategoriesView` does not duplicate row geometry or category logic.

This is presentation-only. Search, kind/wallet filtering, inactive visibility,
system-category metadata locking, wallet applicability, CRUD, archive and delete
continue to use the existing state and API paths. Focused unit tests cover the
system-category editing policy and the recurring editor's mounted draft-session
behavior; they
do not inspect the rendered SwiftUI hierarchy. The rendered tree key is wiring
for later UI acceptance, not unit proof of composition. Integrated visual/device
acceptance remains a later redesign gate.

Review remediation evidence (2026-10-27): the four owned reference test classes
passed 24/24 and the normally signed `MyPocketIOSTests` target passed 422/422 on
the iPhone 16 iOS 18.5 simulator. The Release warnings-as-errors build,
base-component guard, and project verification also passed.
