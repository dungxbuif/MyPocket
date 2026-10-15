# Native categories action parity

## Intent

Align the native Categories/Groups management entry point with the existing web
`/account/groups` contract and the shared iOS base-component rules. The web
contract presents a full-width outline `Nhóm mới` action below the kind
segmented control; the native route currently exposes an icon-only header plus
with different copy.

## Contract

- Keep the native route, owner-scoped category API, form, tree, and edit/delete
  behavior unchanged.
- Expose the create action as a full-width shared `MPButton` outline control
  below the screen header, with the user-facing copy `Nhóm mới`.
- Keep the header reserved for navigation; do not add a screen-local button
  style or duplicate button geometry.
- Add a reusable `outline` `MPButtonStyle` that uses the canonical surface,
  line, action color, hit target and corner-radius tokens.
- Preserve the stable accessibility contract by giving the new action the
  identifier `category-add-group`.

## Verification plan

- Add a focused unit/contract test for copy and the action identifier before
  implementation; run it red.
- Run the focused test green, then the complete native unit/contract target.
- Build Release with warnings as errors, run project/base-component guards and
  `git diff --check`.
- Record the result in validation, audit, iOS design and release notes. No
  backend, migration, owner-scope, or deploy changes are in scope.

## Status

- [x] Contract test added and red run captured
- [x] Base button and Categories screen implemented
- [x] Native tests, Release build and guards pass
- [x] Documentation reconciled

## Final evidence

- Focused contract test: `/tmp/mypocket-categories-action-unit-final2/Logs/Test/`
  (1/1 passed).
- Full native unit/contract target: `/tmp/mypocket-categories-action-full-final/Logs/Test/`
  (102/102 passed).
- Release warnings-as-errors build:
  `/tmp/mypocket-categories-action-release-final` (succeeded).
- Project verification, base-component guard/tests and `git diff --check` all
  passed after the final base-button initializer refactor.
