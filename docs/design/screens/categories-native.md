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
The reference pass uses the app palette and `MPSurfaceCard(style: .reference)`
for the filter group; no sample financial values are persisted.

## Verification

Focused contracts live in `MyPocketIOSTests/CategoriesRecurringReferenceTests.swift`.
The populated UI must be checked with fixture categories containing at least one
parent and two children plus an archived row.
