---
artifact_type: decision_contract
id: FEEDBACK-20261019-PORTFOLIO-CORRECTION-CONTRACT
status: in_progress
updated: 2026-10-19
---

# Portfolio trade correction contract

Feedback `b017284a` asks for editing/deleting/correcting an investment trade.
The current implementation deliberately has an append-only trade ledger:

- `POST /api/v1/portfolio/assets/:id/trades` is the only trade mutation.
- `GET /api/v1/portfolio/assets/:id/trades` is the history read.
- metadata and latest-price updates never rewrite a trade.
- a sell cannot exceed the position and realized P&L/cost basis are derived
  from the complete ordered history.

That is a safety invariant, not an incomplete HTTP route. Adding `PATCH` or
`DELETE` without accounting semantics would change historical cost basis and
could silently alter reports.

## Safe options

| Option | Ledger behavior | Product decision required |
| --- | --- | --- |
| Compensating trade (recommended default) | Append a user-confirmed opposite trade with a link to the source trade; original rows remain immutable. | Whether correction uses original date/price and how the compensating row is shown. |
| Draft replacement | Mark the source as superseded and append a replacement pair in one transaction. | Whether reports use replacement rows and what audit history is shown. |
| Hard edit/delete | Rewrite or remove source history. | Explicit accounting/audit approval; not recommended for v1. |

## Required API contract before implementation

Any future correction endpoint must define `source_trade_id`, correction
reason, effective date, price/fee treatment, report inclusion, idempotency key,
owner scope, and whether the operation is allowed after later trades exist. It
must return the resulting position and report delta, and preserve an audit
trail. Cross-owner IDs must remain indistinguishable from not-found.

## Implemented contract

`backend/internal/entity/portfolio_correction.go` now projects the latest
immutable replace/void revision over the original rows and replays the full
ordered history. `POST .../trade-corrections/preview` is read-only; the commit
route requires `expected_asset_version`, a non-empty reason and
`Idempotency-Key`, locks the owned asset, appends one revision, increments the
asset version and persists a replay receipt. Stale versions and duplicate keys
are conflicts, and invalid resulting inventory is rejected before the revision
is written. Migration `000038` is additive. Native editor/readback and
production rollout remain separate acceptance gates.

## Audit follow-up — 2026-10-20

The read-only preview enforces the same optimistic-version contract as the
commit route. A preview captured against an older asset version returns the
portfolio conflict rather than presenting a projection that can no longer be
committed. This is a safety correction only: it does not choose accounting
semantics for compensating trades, and it does not change the append-only
revision model. The regression is covered by a repository-level version guard
test; PostgreSQL-backed mutation proof still requires a migrated disposable
database and is not claimed here.
