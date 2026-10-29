# Durable wallet confirmation boundary

Status: local implementation complete; production/real-DB rollout remains open. The native confirmation card now keeps a
same-session idempotency key, but the ordinary wallet endpoint still creates a
new row unless its storage adapter participates in the durable mutation ledger.

## Decision

Reuse the existing owner-scoped `transaction_mutations` ledger with operation
`wallet.create` rather than adding a second idempotency table. When a wallet
create request supplies `Idempotency-Key`, the PostgreSQL wallet repository will
reserve the key, create the wallet and persist the response in one transaction.
Same-key/same-payload requests replay the original wallet; same-key/different
payload returns conflict; a pending reservation fails closed. Requests without
the header retain legacy online behavior for existing web clients. Exactly one
header is allowed; keys are ASCII values from 1–128 bytes, and invalid or
duplicate headers return `400`. If a keyed
request reaches an adapter without durable mutation support, the handler fails
closed with `503 mutation_unavailable` instead of downgrading to an unsafe
create.

The proposal payload now also carries a server-generated `action_id`. The
native card reuses that identity as the idempotency key, so a proposal loaded
again from persisted assistant history remains replay-safe before the first
confirmation attempt. Older messages without the field keep the mounted-card
fallback key for backward compatibility.

This closes durable replay of the confirmed wallet write and the identity
boundary for an unconfirmed proposal, without persisting a new financial row
before confirmation. No model tool gains write access, and owner/auth/type
validation remains in the HTTP handler.

## Verification plan

- Handler contract tests cover key forwarding and conflict mapping.
- Handler contract tests cover the missing-adapter fail-closed path.
- PostgreSQL repository tests cover create, replay, payload conflict and atomic
  rollback using the existing real-DB fixture when available.
- Native API tests already prove the caller-supplied key is preserved.
- Native design-system proof covers extracting and reusing a persisted
  proposal `action_id` after history reload.
- No production migration or environment action is part of this code change;
  the existing mutation-ledger migration must be present before rollout.

Local evidence: wallet handler tests pass; the PostgreSQL replay/conflict test
is available at `internal/infrastructure/repository/wallet_postgres_test.go`
but is skipped when `TEST_DATABASE_URL` is unset. Full backend `go test ./...`
and `go vet ./...` pass locally.
