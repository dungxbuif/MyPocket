# ADR-004 — Persist budget scopes, derive progress from ledger

Status: accepted for [API-SCREENS-01](../work/tickets/API-SCREENS-01-DETAIL_DESIGN.md), owner requested API-only screens.

Store owner-scoped budget limits and explicit UTC interval boundaries in a versioned migration. Compute spent from report-included expense transactions and selected category descendants, never from a stored counter or frontend fixtures. Serialize configuration writes per owner to reject exact-scope overlaps safely. Wallet/category deletion cascades tracking configuration instead of changing it to all-wallet/all-category scope. No transaction is deleted by deleting a budget.

Alternatives: localStorage (not shared/persistent backend), mock or stored spent (stale), orphan scopes (silently wider). Consequences: list requires ledger/category reads; recurring scheduling remains separate. Current UI supplies local-date UTC boundaries, pending future account timezone settings. Master [API](../architecture/API.md), [ERD](../architecture/ERD.md), [validation](../work/VALIDATION_MATRIX.md), [release](../releases/CHANGELOG.md), [backlog](../work/BACKLOG.md). Phase none; parent tickets 03-01/03-02 remain partial.
