# Offline create receipt integration

Status: in_progress. Scope approved by owner's full-gap completion request;
implements backend prerequisite of [offline plan](../../superpowers/plans/2026-10-06-offline-financial-sync.md),
not completed offline UX or a production rollout.

Current evidence: migration22 already owns `transaction_mutations`, unique by
owner/operation/key, pending/completed status and JSON response. Transfer patch/
delete reserve and complete in one PostgreSQL transaction. Create endpoints
currently ignore keys although native sends them. Migration30 is occupied by
category-active and production is32. Never create/overwrite migration30.

Decision: extend existing durable receipt mechanism for `transaction.create` and
`transfer.create`, reuse existing category/jar/travel/report-exclusion/atomic-pair
domain writes within the receipt transaction. No second receipt table or migration
needed. Old plan's separate table is superseded for this implementation detail;
financial invariants/retention/owner isolation unchanged. Operation-scoped key
means same key in different operations is independent; changed payload within the
same operation conflicts. No Redis durability or request payload logs.

First slice only adds repository capability and real DB tests; no handler/header/
response change yet, no production writes. Next contract gate must resolve required
keys/coordinated client rollout and metadata placement for array transfer responses
before native outbox wiring. Existing plan's `data.mutation` cannot directly apply
to transfer's array; do not silently ship incompatible envelopes.

Verification: real PostgreSQL on dedicated local55434/schema32, UUID test owners
and cleanup predicates, no user records touched. Runtime type assertion produces
valid RED for missing repository capability, not missing-symbol compile failure.
Cover ordinary replay, payload conflict, owner isolation, concurrent single write,
failure/outer rollback and pair replay/exactly2rows/report exclusion. Run focused
tests, focused race, full Go suite and vet with local-only TEST_DATABASE_URL.
Reconcile API/database/internal docs with capability-not-route status; no Swagger
change until routes change, no public offline availability claim. Broader offline
contract/cache/outbox/sync/UI/E2E and final review remain required.

First implementation gate66476 exposes an existing shared reserve bug: after
INSERT conflict, GORM's First(&mutation) adds candidate mutation.ID as an implicit
PK condition, alongside owner/operation/key. Candidate ID is newly generated and
cannot match stored receipt, so every replay fails record-not-found. Real query
and ordinary concurrent/transfer replay tests reproduce it. Read into a fresh
zero-ID persisted entity before checking hash/status. This fixes the shared
patch/delete replay path too; verify those actual PostgreSQL cases, not just new
capability tests. No schema change or financial rewrite. This is root-cause fix
attempt2, not a timeout/concurrency tuning workaround.

Verification:80988 runtimeRED3/3;66476 real replay/concurrent/pair RED. Five final
tests cover ordinary replay/conflict/owner isolation,8-call concurrent single
write/outer rollback/failure retry, pair identity/exclusion, invalid-key/foreign
wallet and shared patch/delete replay after deletion without resurrection.
Focused race73167 exit0; full Go86668 exit0; final full Go race85540 exit0; vet0.
Commands with `TEST_DATABASE_URL=<dedicated local55434/mypocket_feedback>`:
`go test ./internal/infrastructure/repository -run '^TestOffline.*Receipt' -count=1`,
focused `go test -race` with same selector, `go test ./... -count=1`,
`go test -race ./... -count=1`, `go vet ./...`. Remaining fixture owners0,
schema32 dirtyfalse. No production secrets printed or user financial deletion.
Full suite passing is not live-provider/native offline proof.

Reconcile ADR018/DATABASE/API and offline spec/plan capability status. Swagger
unchanged because handlers/envelopes unchanged. No auth/runtime/deploy change.
Open: header rollout/metadata/hash, HTTP integration, encrypted cache/outbox/sync/
UI/review. Global GORM Info logger uses bound SQL: privacy gate before new HTTP
activation; repository tests alone do not prove absence of financial-data logs.

Shared SQL privacy prerequisite now passes local real-DB/full-race/vet verification;
[ticket](SQL-LOG-PRIVACY-20261014.md) and ADR019 record Scan-recorder bypass and
class-only trace policy. No HTTP wiring or deployed image change. Other arbitrary
application logs, HTTP contract/native offline gates still require verification.

HTTP transitional slice design before code: requests with one valid ASCII1–128
Idempotency-Key use durable create/replay; absent key retains legacy online-only
behavior until coordinated key-required rollout. Explicit blank/duplicate/invalid
header is400, never silently unprotected. This does not complete the spec's final
required-key gate. Body data remains object/array; meta.mutation carries operation,
idempotency_key,status=applied,replayed. First201, replay200, changed typed request
409 IDEMPOTENCY_KEY_REUSED. Hash decoded typed input before generated timestamp/
IDs/active-travel defaults; JSON order/spacing irrelevant, changed fields conflict.
Peek committed receipt before mutable wallet/category validation; a retry after
archive/delete must return original result without resurrection. Reservation at
write still resolves concurrent races. Owner comes only from existing auth context.
No new auth scope/router/migration. Missing capability with a supplied key fails
closed503, not unprotected fallback. Real DB/HTTP tests cover replay, reference
changes, concurrent pair, body conflict, owner isolation and invalid headers.
JWT/API-key middleware and native stable-outbox-key E2E remain later gates.

HTTP evidence:56514 validRED3cases (missing metadata, concurrent retry created
different pairs, invalid key ignored). Keyed handler integration/early receipt
lookup and typed-input hash now pass4cases: original ID/date/report flag after
category archive and wallet delete without resurrection,8concurrent transfers
exactly2excludedrows, invalid/duplicate keys + owner body ignored/foreign owner
404/unauth401, legacy meta:null + missing capability fail-closed.38145 focused
pass; full Go race66898 and final45501 pass, vet0. These use existing handler
auth context directly, not a claim of full JWT/API-key middleware E2E.
`TEST_DATABASE_URL=<dedicated local55434/mypocket_feedback> go test -race ./... -count=1`;
focused selector `^TestOffline(Create|Transfer)HTTP`. Fixtures0/schema32false.

Swagger regenerated via `go generate ./cmd/api`; inspected optional key header,
200/201 typed response envelope and409/503 responses. Public human+AI guide added.
Native compatibility characterization passes full189/189 units0skip at
`/tmp/mypocket-receipt-envelope-native-final-20261014.xcresult`, terminal70137 exit0;
typed object and2-row transfer array accept non-null meta without changing values.
No native product code changed, API restart/prod/sign/install/release occurred.
Required-key rollout, real middleware E2E, encrypted outbox/cache/sync/native
network recovery and final review remain open. This is transitional keyed-path
verification, not final completion of all offline contract tasks.
