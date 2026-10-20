# ADR-007: Separate offline databases per user

Status: accepted, 2026-09-07. Authorized by the user's isolation fix and reliability-testing request.

Context: the original shared IndexedDB database required deleting pending work at account switches to avoid cross-user exposure. A regression test reproduced loss of user A's queued mutation after A → B → A.

Decision: bind the authenticated owner before starting hydration/replay; open `mypocket.offline.v1:user:<encoded-user-id>` per owner. Account switches preserve each separate database. Explicit logout retains the existing behavior of clearing that user's local stores. Unowned legacy storage remains untouched and is never automatically attributed to the next login.

Alternative: clearing the shared database at every switch was rejected because it discards unsynced work. A schema-wide composite-key rewrite is unnecessary for this fix.

Consequences: local storage is logically isolated, not encrypted against someone with access to the browser profile. Legacy offline-only data needs explicit owner-verified recovery; online records can be resynchronized. Physical device and concurrent multi-tab acceptance remain release checks.

Proof: `frontend/src/offline/db.test.ts` covers A → B → A with pending-work preservation; App tests cover no replay under a different account. See [verification](../work/test-verification/BETA-RELIABILITY.md), [context](../CONTEXT.md), [architecture](../architecture/ARCHITECTURE.md).
