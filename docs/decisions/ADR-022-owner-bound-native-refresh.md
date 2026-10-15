# ADR-022 — Owner-bound transport and atomic refresh replacement

Date2026-10-14. Status: locally implemented/verified within approved offline/full
goal; no physical installation or production activation. [Work](../work/tickets/OFFLINE-CLIENT-ACK-20261014.md).

APIClient discarded receipt metadata and could overwrite a login or resurrect a
logout occurring during refresh. sendEnvelope now preserves data/meta; existing
send still returns data with unchanged204/one-refresh/body/key/requestID behavior.

Optional APIRequest.expectedOwnerID is local, never an owner HTTP override. A
session-bound queued request verifies identity before send/refresh and accepting
response. Key/anonymous auth cannot attest this local queue identity; public
full-owner API-key authority is unchanged. Foreign refresh output is refused.

SessionStore conditional replace/clear atomically compares (ownerID,accessToken,
refreshToken), not profile metadata. New same-owner login also invalidates older
refresh. Keychain instances share an app-process recursive lock for all storage
operations/CAS; memory store uses existing lock. Unknown implementations reject
conditional writes by default, not a non-atomic load/save fallback. Refresh401
cannot clear a later login or trigger its session invalidation. accountMismatch
stops old-owner resolution/retry; outbox is never reassigned to another account.

Check-then-save and per-instance-only locks were rejected because they permit
concurrent auth writers. This is NOT inter-process atomicity; future app extensions
sharing the item must revisit coordination. Simulator verifies real Keychain API
behavior, not physical biometric/hardware guarantees.

Five transport REDs and2CAS REDs→219/219(218unit+actual API-key UI),0skip at
/tmp/mypocket-owner-session-cas-final-20261014.xcresult. Includes16concurrent
Keychain candidates/one winner, foreign response/new same/other-owner login/clear
while refreshing,401 after switch, stable original intent through one refresh and
preserved metadata. Existing restore/multipart/device units remain green.

No backend wire/schema change. Finance receipt/resource validator, sync actor,
cache/UI/session events and full device/provider/prod delivery remain open.
[Native](../architecture/MOBILE_IOS.md), [auth](../architecture/AUTH_API.md),
[validation](../work/VALIDATION_MATRIX.md). Not an offline availability claim.

## Refresh-family expiry invariant — 2026-10-15

Each refresh family now carries one absolute expiry shared by its access-session
record and rotated refresh records. Rotation may shorten the next refresh TTL
to that deadline but may not extend the family. Logout tombstones the family
through the same deadline, so a refresh racing with logout cannot outlive the
revocation marker. Legacy records without this field receive a bounded
one-refresh migration deadline and keep their stable legacy family ID. This is
an additive cache-record change; no database migration or route change is
required. Unit coverage must include a refresh that would otherwise extend the
family and a logout/refresh interleaving with TTL-aware cache behavior.
