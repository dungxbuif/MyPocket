# ADR-018 — Reuse durable financial mutation receipts

Status: local implementation decision; public offline contract/rollout incomplete.
Date:2026-10-14. [Work](../work/tickets/OFFLINE-CREATE-RECEIPTS-20261014.md),
[contract](../superpowers/specs/2026-10-06-offline-financial-sync-contract.md),
[plan](../superpowers/plans/2026-10-06-offline-financial-sync.md).

## Context and decision

Old plan proposed a new table in migration30. Current source already has durable
transaction_mutations from22;30 is category-active and current schema is32.
Reuse this table for transaction.create/transfer.create alongside transfer.patch/
delete. Unique scope owner+operation+key detects changed payload within that scope.
Domain write and completed JSON result commit in one PostgreSQL transaction.
Retries return original IDs, never new financial events. Existing category/jar/
travel/pair validation and report exclusion stay authoritative. No new migration,
Redis durability or production data rewrite.

## Alternatives and consequences

A second table offers new columns but duplicates ownership and adds migration risk.
Redis-only durability cannot survive eviction safely. JSON receipts already carry
resource identities without dedicated resource columns. Failed writes roll back
their receipt; client outbox must retain rejection state. Same key in distinct
operations is independent, matching the existing unique key.

Shared reserve had a candidate-primary-key lookup bug; fresh persisted entity
fixes replay including patch/delete after pair deletion. Real DB tests/full Go race
pass locally, not HTTP/native availability proof. Required headers, canonical
hashing, array-safe response metadata and financial SQL logging privacy are gates
before route activation. No release or production deployment inferred.
