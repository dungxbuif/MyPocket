# Bills, debt, travel and portfolio native presentation

Status: in progress — reference-density wave (2026-10-16).

This screen contract records the populated-state composition for the native
secondary features. The physical Money Lover reference shows grouped white
surfaces on a light canvas, compact rows, a single primary heading per route,
and overflow actions for row-level management. MyPocket keeps its own API,
owner scope, transaction semantics and privacy masking; sample values are not
copied into the product.

## Shared layout rules

- Use `MPSurfaceCard(style: .reference)` for one grouped surface per logical
  section; do not wrap each metric or row in a separate card.
- Use existing `MPText`, `MPMetricGrid`, `MPTransactionGlyph`,
  `MPStatusBadge`, `MPOverflowMenu` and `MPButton` components. Screen code may
  compose them but must not introduce local fonts, colors or corner radii.
- Keep rows compact while preserving the 44pt hit target, full monetary values,
  masked values and VoiceOver labels.
- Mutations, CRUD, wallet ownership, timezone formatting and API contracts are
  unchanged by this visual wave.

## Route contracts

### Recurring transactions

The route has a title/header, one explanatory action surface, then one grouped
list. Each row contains a direction glyph, name/status, type and frequency,
wallet/next-run context, amount, and an overflow menu for edit/delete. Empty,
loading and error states remain explicit.

### Travel mode

The route has one compact status summary followed by one grouped list of
events. Each event row shows name, date range, active state and an overflow
menu. The existing explicit toggle identifier remains available for
accessibility and UI tests; no transaction is moved by this presentation.

### Portfolio

The route has one compact summary group for cost basis, current value and
unrealized P/L. Assets are shown as grouped compact rows with symbol/name,
quantity/price context, a compact metric grid and existing price/trade/delete
actions. Trade history stays in the same group as the asset and uses the shared
transaction row.

### Credit/debt ledger

The route retains the selected credit wallet, four-value statement summary,
entry/payment actions and statement rows. The summary uses the shared adaptive
metric grid so populated credit data remains readable on compact and large-text
devices. Credit mutations retain their existing confirmation and owner scope.

## Verification

Focused native tests assert the deterministic populated presentation contracts;
the signed unit suite and release build are required before this wave is
reported as complete. No backend, database, production or reference-app data
is changed.
