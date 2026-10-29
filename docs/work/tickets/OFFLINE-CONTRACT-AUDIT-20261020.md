# Offline sync contract audit — auth, owner switching, and ambiguous acknowledgements

Status: verified_local
Owner: shared
Date: 2026-10-20

## Scope

This audit checks the implemented native outbox against the approved offline
contract for four failure boundaries: expired access tokens, account changes
while a request is in flight, an acknowledgement that cannot be trusted, and
retry/idempotency preservation.

## Required behavior

- A `401` on a session-authenticated mutation is refreshed at most once by
  `APIClient`; the original body, request id, and `Idempotency-Key` are reused.
- If the authenticated owner changes before or after transport, the mutation
  is not published or acknowledged into the new account. It remains
  `blockedAuth` and must be resumed explicitly by the original owner.
- Missing or malformed mutation receipt/resource data is an ambiguous outcome.
  The intent remains retryable with the same key; it must not be marked
  applied, re-keyed, discarded, or counted as a second local write.
- Transient transport/HTTP failures use the existing capped retry policy;
  terminal `4xx` and conflicts do not retry automatically. Every replay uses
  the persisted key and canonical payload.

## Verification plan

Run deterministic XCTest cases using the real `APIClient` and encrypted SQLite
outbox with an in-memory HTTP transport. Assert request sequence/headers/body,
owner isolation, state transitions, and stable key. No production endpoint,
credential, database, or migration is used.

## Reconciliation

The existing unit suite already covers most store transitions and timeout
replay. This audit adds an explicit owner-switch-after-transport case so the
end-to-end boundary is guarded rather than inferred from separate APIClient
tests. The existing `APIClientTests` cover the 401 refresh sequence (same
request body/key and one refresh); this audit confirms the sync actor preserves
that adapter contract. The focused simulator run passed 9/9 tests:

```text
xcodebuild ... -only-testing:MyPocketIOSTests/SyncEngineTests test
TEST SUCCEEDED — 9 tests, 0 failures
```

The 401-refresh sequence remains covered by the real `APIClient` transport
tests; no production endpoint or credentials were used. Physical touch/UAT,
production rollout, and populated report/device acceptance remain separate
gates.
