# API-key owner contract (2026-10-19)

## Scope

This contract covers user API-key authentication for all owner-scoped routes. A
key is a full-access credential for its own owner (legacy scope labels are
metadata only), but it must never produce a principal without a valid owner.

## Invariants

- API-key creation trims and validates the owner identifier before persistence.
- Authentication rejects a persisted key whose owner identifier is blank. This
  fails closed instead of allowing a handler to execute with an empty owner.
- The authenticated principal's owner comes only from the persisted key; request
  bodies, query strings, and path values cannot override it.
- Expiry, revocation, constant-time secret comparison, and owner-scoped revoke
  remain enforced.
- This change does not add permissions, rewrite legacy scope metadata, or mutate
  production data.

## Verification

The use-case contract tests cover owner normalization and rejection of a
malformed ownerless persisted credential. Existing API-key HTTP tests continue
to cover JWT/API-key parity, same-owner full access, cross-owner isolation,
idempotency, and revocation.
