# Offline server capability gate

Status: in_review (local gate verified, sync wiring/rollout pending). Parent: [offline plan](../../superpowers/plans/2026-10-06-offline-financial-sync.md).
Approved full offline scope needs a fail-closed server-version boundary before
outbox sends, because older deployed create routes ignore Idempotency-Key.

Design before code: GET /api/v1/sync/capabilities under existing JWT/live-session
middleware, read-only/no new financial permissions. Response data:
contract_version=1,enabled=(durable create repository + transfer writer configured),
operations=[transaction.create,transfer.create] or empty array, receipt_metadata
=/meta/mutation,idempotency_key_required_for_offline=true,legacy_unkeyed_online=true.
Does not claim DB health or availability; errors still require retry policy.

Native typed model accepts only supported version1,enabled,required-key policy,
known requested operation advertised and exact metadata location. Unknown version,
unknown/missing operation, disabled/malformed policy or old-server404 fail closed.
Never infer support from health200/header acceptance. Fetch with current session
before each sync drain; future offline queue/cache may retain last-known capability
for draft UX but must not use stale capability to authorize a network write after
rollback. No bank/credit/AI/offline edit/delete added. Key-required global rollout
and API-key-write design remain separate open gates.

Proof: runtime HTTP404 RED on real router, supported/unsupported/nil repository
cases, auth401 including revoked session via real PostgreSQL/Redis integration.
Native policy runtimeRED using a fail-closed stub then implementation, decoder/
invalid-version/operation/policy tests and authenticated GET wire capture. Run Go
race/vet/generate + native units/guards/project; reconcile OpenAPI/public human+
agent guide, native API/architecture/context/backlog/matrix. No DB schema, live API
restart, production rollout or app installation in this slice. Full offline
store/cache/outbox/sync/UI/recovery remains required; capability alone is not it.

Post-generation verification75975: full Go race with localPG/Redis exit0, vet0.
This runs after generated Swagger/current source were finalized, not concurrently
with generation. No live test process remains for this slice.

Evidence:76809 valid HTTP404 RED in durable/legacy/nil cases.1027 focused Go pass;
41946 full Go race/vet with dedicatedPG55434/Redis16379 pass. Real JWT/Redis auth
integration now checks capability200 before logout/401 after. OpenAPI generated
and inspected200 typed data +401 +BearerAuth. Native69552 runtimeRED: two positive
policy/wire tests fail against explicit fail-closed policy scaffold, negative
test passes; no compile failure counted. Implemented strict known/version/policy
predicate. Final77223 terminal0 at
`/tmp/mypocket-sync-capability-native-final-20261014.xcresult`193/193 PASS0skip,
including disabled/missing/version/unknown-operation/key/metadata cases and old404
read-only wire with1GET/no mutation. Guard/project/diff pass. Fixture owners and
Redis test namespace0/schema32false. No API restart, schema/prod/sign/install/
release change. Public human+agent guide/API/spec/current docs reconciled; no UI
screen/base styling change, so screen docs unchanged. Store/sync execution has
NOT yet consumed this gate; never claim runtime offline protection from it alone.
