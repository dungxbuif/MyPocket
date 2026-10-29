# AI advisor category confirmation idempotency

Status: verified
Date: 2026-10-20

## Scope

Make confirmed `create_category` actions safe to retry after a lost response and
safe under concurrent confirmation. The same owner, operation, and action key
must return the same category instead of creating a duplicate. A reused key with
different payload must be rejected, and the same key may be used independently
by another owner.

## Root cause

The advisor confirmation handler forwards the action id as `Idempotency-Key`, but
the previous category interactor called the ordinary `CategoryRepository.Create`
path. That path had no durable mutation receipt, so a second confirmation could
insert another category. The existing `transaction_mutations` ledger already
provides the required unique owner/operation/key boundary and is used by wallet,
budget, jar, and transaction writes.

## Design decision

- Reuse `transaction_mutations`; do not add a migration or an in-memory cache.
- Add a narrow `CategoryCreateMutationRepository` capability and a matching
  interactor method. Normal API creates keep their current behavior when no
  idempotency key is supplied.
- The PostgreSQL implementation reserves `category.create` in the ledger and
  writes the category and category-wallet links in the same transaction. It
  stores the complete category response as JSON for deterministic replay.
- The handler validates a single printable ASCII `Idempotency-Key` (1–128
  characters), returns `201` for a first write and `200` for a replay, and maps
  payload conflicts/pending mutations through the existing mutation error
  contract.
- Owner isolation comes from the ledger's `(owner_id, operation,
  idempotency_key)` unique key and all category writes remain owner-scoped.

## Verification plan

1. Handler regression: first confirmation, exact replay, changed-payload
   conflict, and owner separation.
2. PostgreSQL integration: one durable ledger row and one category under
   concurrent same-key confirmations; replay returns the persisted id/fields;
   another owner can use the same key.
3. Run category/controller tests, the full Go suite, `go vet ./...`, and
   `git diff --check`.

## Residual scope

This change covers confirmed category creation. Category update confirmation is
still protected by expected-version CAS but does not yet have a durable replay
receipt; that requires a separate mutation contract so it is not silently
claimed as fixed here.

## Implementation evidence

- `CategoryHandler.CreateCategory` validates the key, dispatches the
  idempotent interactor path, and returns `201` for the first write / `200` for
  a replay.
- `CategoryPostgresRepository.CreateIdempotent` uses the existing
  `transaction_mutations` unique key and commits the category, wallet links,
  and response envelope in one transaction.
- Controller tests cover direct keyed create and persisted advisor action
  confirmation replay. PostgreSQL tests cover payload conflict, owner
  isolation, durable receipt state, and concurrent single-write behavior when
  `TEST_DATABASE_URL` is supplied.
- Verification completed locally: `go test ./... -count=1`, `go vet ./...`,
  and `git diff --check` passed. A disposable PostgreSQL 17 container was
  migrated to version 40 (`dirty=false`), then the category replay/conflict,
  owner-isolation, concurrent-confirmation, category archive, handler, and
  advisor confirmation tests passed both normally and under `go test -race`.
  The disposable container was removed after verification; no production or
  shared database was used.

The final MCP review wave explicitly integrates these authorized category files
and tests, eliminating the previous dependency on uncommitted worktree changes.
Permanent handler proof with the original committed category files overlaid
fails with replay `201` instead of `200`. Added regression coverage includes
changed-payload `409`, invalid keys, owner separation and a real PostgreSQL
transaction with an injected wallet-link insert failure: category, links and
receipt all roll back. Candidate-isolation evidence and exact final counts are
recorded in [MCP-FINAL-REVIEW-20261020](MCP-FINAL-REVIEW-20261020.md).
