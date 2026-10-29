# Reports and transaction detail native presentation

Status: reference-density wave (2026-10-16).

Report trend, category and activity sections, savings report drilldowns and
transaction detail now share grouped reference surfaces. The existing report
scope, period validation, owner filtering, travel linkage, edit/delete rules,
transfer handling and masking are unchanged.

## Layout contract

- Keep one compact reference surface per report section.
- Use shared chart/metric/transaction bases; feature views supply only data and
  accessibility identifiers.
- Keep transaction detail's fixed header, amount summary, metadata, travel
  selector and mutation actions in the existing order.

## Verification

The existing transaction reference tests pass 3/3. Base-component/project
guards and Release simulator build with warnings-as-errors pass; no backend or
production state changed.
