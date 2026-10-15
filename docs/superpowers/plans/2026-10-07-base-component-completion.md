# Base component completion plan

## Intent

Close the remaining `docs/design/INDEX.md` partial rows that are already required
by mounted product flows. The shared web and native bases must expose the same
semantic state, clamp unsafe values at render time, and keep business data in the
feature layer. No preview-only or fabricated financial data is introduced.

## Scope

1. Add a native `MPBudgetGauge` that consumes the real budget summary/item
   values and has the same clamped progress and accessibility contract as the
   web `BudgetGauge`.
2. Add a named cross-platform status/percentage badge base. It is a semantic
   status surface only; it does not infer a marketing or promotional state.
3. Reconcile the selector and bar-chart inventory against their actual callback,
   data, and accessibility behavior. If a row is already complete, update the
   inventory rather than adding duplicate components.
4. Add unit/design tests before implementation (RED then GREEN), run native and
   web gates, and update the architecture, design, validation, and changelog
   docs with the shipped behavior and remaining external UAT gates.

## Verification

- Native focused tests for gauge clamping, status badge labels, and budget
  composition; full native unit suite and Release warnings-as-errors build.
- Web design/typecheck/build/link checks and existing chart/component tests.
- `git diff --check` and repository base-component guards.

## Non-goals

- No production deploy, database mutation, provider/OCR change, or fake budget
  records.
- No claim of pixel-level or physical-device parity until device UAT is run.

## Result

- Added `MPBudgetGauge` and `MPStatusBadge`; mounted them in native budget and
  recurring surfaces. Added the matching web `StatusBadge` to recurring rows.
- Reconciled selector, notice, bar-chart and percentage/status inventory rows
  against their actual mounted callbacks/data. No promotional banner was
  invented because no product/API contract exists for one.
- RED/GREEN focused tests pass. The native full unit/contract suite is 129/129
  in `/tmp/mypocket-final-full.xcresult`, the Release warnings-as-errors build
  passes from `/tmp/mypocket-final-release`, web design/typecheck/build/link
  gates pass, and native project/base-component/diff guards pass.
- The recurring API-backed UI harness now resolves the add action by its stable
  `recurring-add` identifier and asserts the current shared-base label `Thêm`;
  this removes the stale pre-parity `Thêm lịch` lookup without changing the
  user-facing contract.
