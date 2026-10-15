# Frozen Travel context for delayed transaction creation

Status: in_progress. Approved scope: full-gap/offline contract, preserving online
compatibility and all production data. [Plan](../../superpowers/plans/2026-10-06-offline-financial-sync.md),
[store](OFFLINE-ENCRYPTED-STORE-20261014.md), [completion](NATIVE-FULL-COMPLETION-20261014.md).

Before implementation: CreateTransaction currently auto-links the trip active
at request arrival when travel_event_id is nil. A queued September expense could
therefore join a different October trip. Add create-only optional boolean
auto_link_travel: omitted/null/true keeps online default; false prevents fallback
lookup. Explicit owner-scoped travel_event_id is still validated/used, even if
inactive. Transfer already has no automatic Travel association. No schema change.

Use transactionCreateInput embedding existing transactionInput; new field uses
omitempty. Legacy omitted-field request hashes must be byte-for-byte unchanged
so preexisting receipt replay works. false/true are distinct intents and reuse
of a key with changed policy conflicts. Replay must precede active-trip/default
and mutable explicit-trip validation as before; no resurrection after deletion.
PATCH schema/behavior unchanged; use dedicated link/unlink API for existing rows.

Capabilities advertise transaction_create_context="explicit_travel_v1" only
for configured durable writer. iOS transaction.create support must fail closed
without that exact value; transfer.create remains separately gated by existing
version/key/meta contract. Future outbox entry freezes selected/cached trip plus
auto_link_travel=false before first send; no fallback to a later active trip.
This slice must not activate sync/UI by itself or expand API-key permissions.

TDD: actual HTTP/PostgreSQL active-trip A→B, false/no-link, explicit inactive A,
legacy/true auto-B, changed-policy conflict, receipt replay after trip deletion,
foreign/missing trip denial and unchanged legacy hash literal fixture. Native
old-context JSON must block ordinary writes but not safe transfer; correct
context accepted, unknown context refused. Regenerate Swagger, full Go race/vet,
native units, guards/project/diff. Reconcile API/Travel/offline human+agent docs,
spec/current ledger and validation. No production migration/deploy/install/release.

Result: valid backend runtimeRED66771 (earlier29558 same) explicit-none wrong trip,
changed policy not conflict and missing context capability. Legacy golden receipt
already passed as baseline. Implemented create-only embedded DTO/omitempty,
fallback condition/capability; focused12705pass. Native runtimeRED34132 twofail;
policy gating and frozen enqueue added; final59173exit0,211/211units0skip bundle
`/tmp/mypocket-frozen-travel-native-final-20261014.xcresult`. Existing valid
capability fixtures updated to advertise new context, so old negative version/
key/meta tests still isolate their own policy defect instead of missing context.
Full Go race51567exit0/vet0 after Swagger generation, dedicated55434/16379. New
actualDB HTTP tests cover no-trip/original inactive trip/current online trip,
fixed persisted date/trip, conflict, deleted-trip replay/new404/foreign404 and
pre-change canonical receipt hash including omitted/null policy compatibility.
Base/project/diff pass. Human+agent API/offline/Travel/spec docs reconciled.
No server schema/migration or API-key auth change. Full-goal sync/cache/UI and
production/device gates remain; not complete offline feature or release.
