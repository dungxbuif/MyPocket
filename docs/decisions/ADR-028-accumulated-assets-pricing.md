---
artifact_type: adr
id: ADR-028
status: accepted_for_implementation
updated: 2026-11-01
---

# ADR-028: Explicit units and provenance for accumulated-asset pricing

## Context

Portfolio holdings already store fixed-point quantities, immutable trades, and
correction receipts. The accumulated-assets slice must make the unit explicit
without guessing what legacy quantities meant, and it must later allow a
system-catalog quote without rewriting purchase history or wallet cash.

## Decision

Add nullable `unit_label`, non-null `unit_confirmed`, and non-null `price_mode`
metadata to each portfolio asset. Existing rows migrate to `manual`, with a
null unconfirmed unit. A version-scoped unit confirmation endpoint records the
owner's interpretation and never converts quantity, price, trades, or
corrections. A confirmed unit cannot change after trades exist. Manual price
updates are rejected for assets in automatic mode; future market quotes remain
separate from asset versioning.

## Alternatives

- Infer gold units from symbol/name: rejected because legacy data has no safe
  unit contract.
- Rewrite quantities during migration: rejected because it changes financial
  history and cannot be proven reversible.
- Add a second portfolio ledger: rejected because the existing correction and
  weighted-average engine already owns this holding history.

## Consequences

The schema/API are additive and preserve legacy reads and writes. Clients must
surface an explicit unit-confirmation step before system price binding. The
initial migration does not provide automatic products; external provider
contract evidence is still required before catalog activation.

## Links

- [Accumulated-assets design](../superpowers/specs/2026-10-31-accumulated-assets-design.md)
- [Implementation plan](../superpowers/plans/2026-10-31-accumulated-assets.md)
- [Portfolio API](../architecture/API.md)
