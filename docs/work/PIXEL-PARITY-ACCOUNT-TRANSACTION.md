# Native account and transaction visual parity

## Decision

Use the Stitch export as the visual reference for the native iOS account
("Mở rộng") and add-transaction surfaces. The web app and existing native
API contracts remain the source of truth for behavior, validation, wallet and
category semantics, and navigation. This is a presentation change only: no
transaction endpoint, payload, ownership rule, or persistence behavior may
change.

## Intended behavior

- The add-transaction screen follows the Stitch hierarchy: close/title header,
  expense/income segmented control, wallet/category rows, amount with VND
  context, optional detail section, date row, and a persistent save action.
- Existing jar, report inclusion, note, edit, validation, loading, and error
  states remain available through the optional detail section.
- The account screen keeps every existing destination and accessibility ID but
  groups them into Money Lover-like account sections with a profile surface and
  stable row hit targets.
- All new visuals use shared `MP*` base components and the existing Stitch
  color/type/radius tokens; feature screens must not introduce ad-hoc controls.

## Verification

1. Add a failing UI contract test before implementation.
2. Run the targeted UI test red, then green after implementation.
3. Re-run all 30 native UI cases (20 API-backed with `LOCAL_API_UI`, 10 smoke
   cases without the flag), 75 native unit tests, project/base-component
   guards, backend tests/vet, and web typecheck/build.
4. Record results in `docs/work/VALIDATION_MATRIX.md` and
   `docs/releases/CHANGELOG.md`. Physical-device and production visual sign-off
   remain explicit external gates.

## Follow-up overview slice

The supplied Stitch overview evidence places the wallet summary before the
insider/report surfaces. Native `OverviewView` now follows that hierarchy and
exposes stable section identifiers (`overview-section-wallets`,
`overview-section-insider`, `overview-section-report`,
`overview-section-month-summary`, `overview-section-recent-transactions`) so
future visual UAT can compare the same regions without changing API behavior.
