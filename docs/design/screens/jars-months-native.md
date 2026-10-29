# Jars and month summary native presentation

Status: reference-density wave (2026-10-16).

The Jars and Months routes keep MyPocket's own monthly allocation, note and
report contracts while using the same compact grouped surfaces as the other
native routes. A page-level metric grid replaces separate metric cards; jar,
category and note sections use one grouped reference surface each. This keeps
the vertical rhythm usable on iPhone while preserving all CRUD, masking,
timezone and owner-scoped API behavior.

## Shared layout contract

- Use `MPSurfaceCard(style: .reference)` for each logical section.
- Use `MPMetricGrid` for income/expense/allocation and month summary values.
- Keep progress, note editing, month switching, error/loading/empty states and
  existing accessibility identifiers unchanged.
- Do not change the jar allocation rules or month-report date boundaries.

## Verification

`JarsMonthsReferenceTests` covers stable create-action identifiers and the
account-timezone inclusive month period. Base-component and project guards
pass; no backend, schema, production or physical-device state changed.
