# Offline receipt authentication integration

Status: in_review (existing policy verified; new API-key write design open). Parent [receipt work](OFFLINE-CREATE-RECEIPTS-20261014.md).
This slice verifies existing policy, not a security permission expansion.

Source evidence: NewRouter registers financial routes under RequireAuth (JWT +
live session verification). User API-key catalog has finance:read, advisor and
feedback scopes, no financial write scope. Do not repurpose finance:read as write
or claim API-key financial creates work. A separate transactions:write design
choice was presented to owner asynchronously; existing behavior remains until
explicit security design approval. Other approved offline work continues.

Verification before code: real PostgreSQL fixtures, bcrypt login/real JWT and
actual Redis adapter on dedicated local16379. Test-only namespacing wrapper maps
every cache key to a unique prefix and deletes only tracked exact keys, no flush.
Use production NewRouter/RequireAuth/VerifySession, not manually supplied userID.
Valid session writes/replays; other owner's session cannot replay; missing/bad/
expired JWT and logged-out session cannot even replay a stored receipt. Generate
real hashed user API key in PostgreSQL, prove finance:read validity in key service
and denial on financial write routes; check expired/revoked key state separately.
No credentials printed, no production access/schema changes/product auth changes.

Run focused with TEST_DATABASE_URL + TEST_REDIS_URL, then full Go race/vet. Update
API/guide/current docs to state JWT-only writes and outstanding API-key write design.
This is characterization of intended existing policy; passing initially is not
claimed as a product bug fix. Provider/bubble/outbox/capability/prod/device gates
remain open; no release acceptance or feedback status mutation.

Focused71010 exit0 on DB55434 + Redis16379. Final full
`TEST_DATABASE_URL=<dedicated PostgreSQL> TEST_REDIS_URL=redis://127.0.0.1:16379/0 go test -race ./... -count=1`
48372 exit0, vet/diff0. Actual router/login/bcrypt/JWT/Redis VerifySession proves
owner create/replay; other session404, missing/bad/expired JWT401, actual Redis
TTL expiry401, logout/family revocation401 even for stored receipt. Ledger and
receipt counts remain1each. Real PostgreSQL key valid for finance:read, wrong
service scope denied/expired/revoked invalid; financial routes reject valid read
key401 because JWT-only. This is NOT positive API-key financial-write support.
Redis audit stream test also runs in full suite with configured Redis.

Readback fixture owners0/Redis test:offline-auth:* keys0/schema32false. Only tracked
namespaced cleanup; no flush, prod secrets/access/migration/auth policy/native
product change. API/OpenAPI/guide/current docs reconciled. New transactions:write
choice awaits owner design decision; unrelated outbox work may continue.
