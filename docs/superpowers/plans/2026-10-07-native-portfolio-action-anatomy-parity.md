# Native portfolio action anatomy parity

## Intent

Match the native `PortfolioView` to the web `PortfolioPanel` action anatomy
captured in the supplied Stitch/web reference: the header is navigation plus
refresh only, while asset creation and portfolio trade entry are visible
domain cards instead of a header plus action or hidden conditional forms.

## Contract

- Keep `/api/v1/portfolio/*`, owner scoping, validation, price refresh and
  append-only trade behavior unchanged.
- Compose `MPSurfaceCard`, `MPSectionHeader`, `MPTextField`,
  `MPOptionSelector`, `MPWalletSelector`, `MPDateField` and `MPButton`; no
  screen-local control styling or raw interactive controls.
- Preserve per-asset `Giao dịch` as a convenience that selects the shared
  trade card, while the canonical save action is the card-level `Lưu giao
  dịch` action used by the web form.
- Expose stable IDs for refresh, asset creation and trade submission.

## Verification plan

- Add a copy/ID contract test and capture RED before implementation.
- Make the native form/card placement match the web route and run focused and
  full native unit tests.
- Run the disposable API portfolio UI smoke, Release warnings-as-errors,
  project/base-component/diff guards and the app design checks.
- Reconcile the iOS renderer contract, validation matrix, port audit and
  changelog.

## Status

- [x] Contract test added and red run captured
- [x] Native action/card anatomy implemented with shared bases
- [x] Native unit, web design/build, Release and guard checks pass
- [x] Documentation reconciled

## Final evidence

- RED compile evidence: `/tmp/mypocket-portfolio-action-red`.
- Focused GREEN contract test: `/tmp/mypocket-portfolio-action-unit-final`
  (1/1 passed).
- Full native unit/contract target:
  `/tmp/mypocket-portfolio-action-full-final` (107/107 passed).
- Targeted unauthenticated launch smoke:
  `/tmp/mypocket-portfolio-action-ui-launch` (1/1 passed).
- Release warnings-as-errors build:
  `/tmp/mypocket-portfolio-action-release-final` (succeeded).
- Web design/type/build checks, project verification, base-component guard and
  `git diff --check` passed. The all-UI invocation was terminated after the
  known Xcode runner hang; the bounded launch smoke is the available UI proof
  for this slice. Live disposable portfolio API mutation should be rerun when
  the local backend is available.
