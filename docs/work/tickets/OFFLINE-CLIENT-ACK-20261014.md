# Owner-bound offline HTTP acknowledgement adapter

Status: in_progress. Approved [offline Task5](../../superpowers/plans/2026-10-06-offline-financial-sync.md),
[full goal](NATIVE-FULL-COMPLETION-20261014.md). Persistence/lease/Travel/full-key
server prerequisites verified; no app sync activation yet.

Beforecode: APIClient currently discards meta and FinanceAPI generates a new UUID
for each create call. Sync cannot use those helpers. Add sendEnvelope preserving
decoded meta with existing one-refresh behavior, and optional expectedOwnerID on
APIRequest. Session-authenticated owner-bound request checks loaded identity before
transport, before refresh, and before accepting response. Refresh must refuse a
changed local account or a server response for another owner before saving any
token. Mismatch is typed auth-blocking, not an automatic retry under new account.
Do not send owner-bound requests using arbitrary API-key auth lacking local owner
attestation. Normal online methods keep their API and request semantics.

FinanceAPI outbox adapter uses persisted idempotencyKey and frozen payload, not
UUID generation; caller must have a sending lease. Verify meta.mutation exact
operation/key/status=applied and typed replayed boolean; validate returned resource
owner, ID, wallet/type/amount/date/report/trip against saved intent. Transfer is
exactly2distinct IDs, paired transferID, sourceexpense/destinationincome and both
excluded. Invalid/missing receipt/data leaves the intent uncertain for same-key
replay, never applies/deletes/rekeys it. Only acknowledged IDs leave adapter.

Tests first: preserved metadata, stable wirekey across first/replay, missing/wrong
receipt refusal, ordinary and two-row array acknowledgements, changed owner before
send/during response/refresh refuses transport or credential replacement. Use real
APIClient/typed payload and SQLite leases; deterministic fake HTTP transport is
external boundary only. Full native unit regression, base/project/diff and docs.
Sync actor retry/backoff/stop/cache invalidation/UI remain following Task5/6 work;
do not claim offline availability from HTTP adapter alone. No backend/prod schema,
deployment, installation, release or user data change.

Owner transport checkpoint: runtimeRED18476 fivecases; current check-before/save
implementation passes full native+APIUI54401. Remaining check/save race must be
closed before adapter use: extend SessionStore with atomic replace/clear only if
owner/access/refresh credentials still match. Keychain instances share an app-
process recursive lock for all load/save/clear/CAS; memory store uses existing
lock. No cross-process/hardware atomicity claim. Reject replacement with foreign
owner, preserve cleared/new same-owner/new other-owner credentials. Conditional
clear on refresh401 must not invalidate a later login. Test real Simulator
Keychain across2instances and memory store; keep test doubles contract-aligned.
Only then use CAS in refresh. Metadata/UI/profile changes don't change credential
identity; compare tuple(ownerID,access,refresh), not full object equality.

## Foundation result — financial adapter still pending

Initial exhaustive APIError compile finding corrected: accountMismatch behaves
like unauthorized in AIEntry, never resolves old-owner request via new account.
Valid runtimeRED18476 five owner/envelope cases→54401full217/217pass. CAS RED3343
then expanded86604 two actual-store cases against fail-closed scaffold. Added
global app-process Keychain lock/memory lock/conditional replace+clear; test
stores aligned, unknown stores fail closed. Final58452exit0,219/219PASS0skip
(218unit+actualAPI UI), bundle /tmp/mypocket-owner-session-cas-final-20261014.xcresult.
StandardSE LOCAL_API_UI warnings-as-errors command; MyPocketIOSTests plus
LaunchTests/testAPIBackedAccountAPIKeyCreateAndRevoke. Tests16Keychain contenders/
onewinner, foreign refresh owner, switch on200/401, clear/new same-owner login,
stable original body/key/requestID through one valid refresh and retainedmeta.
Existing restore/device/multipart regressions green. Base/guardtests/project/diff
pass. ADR022/auth/native/public offline docs reconciled; no backend wire/schema
change, so Swagger/server ERD unchanged. Financial receipt/transfer checks above
NOT implemented yet; next validator then actor/backoff/cache/UI/events. No offline
activation/prod/sign/install/release or whole-goal completion claim.

## Receipt adapter checkpoint

Added `FinanceAPI.applyOutbox` and `OutboxAcknowledgement`: sending/current-time
unexpired lease and frozen ordinary travel policy required before HTTP; persisted
key/body and expectedOwnerID reused, no UUID regenerated. Validate exact receipt
operation/key/applied status/boolean replay marker, one ordinary resource or two
distinct paired transfer resources. Owner/wallet/type/amount/date instant/report
flags and ordinary category/jar/trip must match. Transfer order is not assumed;
ack IDs return source then destination. Notes may be normalized by server.
Adapter has no SQLite state transition: subsequent engine must CAS current lease,
so delayed acknowledgement cannot overwrite auth-blocked/released work.

RuntimeRED63069 validator accepted cases fail against fail-closed scaffold;
GREEN23224 full220 units. Adapter fixtures initially had compile errors (await
inside XCTUnwrap, private store, String expiry and throwing save), corrected before
valid runtimeRED38814 two adapter cases. Final70626exit0,223/223units PASS0skip,
`/tmp/mypocket-ack-http-final-20261014.xcresult`, standardSE LOCAL_API_UI warnings-
as-errors command. Real encrypted SQLite lease + real APIClient with external
HTTP double prove stable body/key, metadata rejection, transfer array and no
transport for pending/expired/unfrozen work. Pure validator negatives cover
foreign owner, changed resource fields/cardinality and malformed receipt.
Base/guardtests/project pass. No new files/packages/backendwire/schema so no
XcodeGen/Swagger/ERD changes. Public offline guide updated for human/agent readers.
Not live backend financial E2E, timeout/retry/cancellation engine or UI acceptance.
Next Task5 sync actor/backoff/capability/owner lifecycle; cache/UI/events follow.
Full goal remains active; no prod/sign/install/release or feedback acceptance.
