# Foreground offline sync engine

Status: in_progress. Implements approved offline Task5, not entire offline/app goal.

Before code: actor owns one active drain, coalesces same-account triggers and
refuses different-account triggers until stopped. Use real encrypted OutboxStore
leases, FinanceAPI owner-bound capability and receipt adapter. Fresh capability
gate before leasing; unsupported/malformed/old server preserves queue unchanged.
Only foreground/network/manual/same-owner-login triggers; no background writes.
Explicit same-owner login resumes auth-blocked rows, never another account.

On success persist applied via lease CAS, then invalidate affected reads through
an injected production callback; no pending sums added to server balances.
Transient transport/408/429/5xx and ambiguous receipt/decoding use same-key retry
with exponential jitter capped900s. 400/403/404/422 reject;409conflict; exhausted401,
missing session/account mismatch block owner queue. Never store raw error bodies.
Logout cancels request task, invalidates generation and blocks queue; late callbacks
cannot apply or invalidate another owner's cache. Persist retry date, no sleeping
loop. Future foreground scheduler/UI handles next wake.

TDD: real SQLite + deterministic clock/jitter and network boundary double prove
apply/FIFO/coalescing/capability gate, timeout then identical key replay, retry
classes/cap, permanent/conflict, auth-block/sameowner resume/logout-late callback.
Full native units/base/project/diff. Human+agent offline guide/current docs/ledger
updated after evidence. No backend schema or public API change, prod deploy,
signed install or release acceptance in this slice.

Regression investigation: full lifecycle gate24753 showed one existing HTTP
adapter assertion failure: equal length228byte request bodies differ on replay.
Shared JSONEncoder.mypocket uses no sortedKeys; OutboxPayload canonical storage
does sort keys. Hypothesis: JSON object iteration order changes wire bytes despite
same frozen fields (server typed hash remains equivalent). Tighten adapter test
against independently sorted JSONSerialization bytes before fixing only outbox
wire encoder; online encoding remains unchanged. No financial/key/content change.

## Local implementation checkpoint

Engine uses real store leases and fresh expected-owner capability before drain;
same-owner active triggers coalesce, different owners cannot share a drain.
FIFO persisted CAS outcomes, explicit same-owner-login resume, auth blocking,
generation/cancellation on stop and owner-tagged invalidation callback implemented.
Retry schedules persist; no sleeps/background scheduler. Policy retries network,
408/429/5xx/unverified acknowledgements, rejects permanent HTTP,409conflict,
401/missing/mismatched auth blocks. Error strings stored are controlled codes.
Ruling: capped exponential1,2,4,8,16…900s with0–25%jitter (hardcap900) follows
plan's capped exponential wording, not a sudden8→900jump in ambiguous spec prose;
cost: more early transient retries than that alternate reading.

RuntimeRED45673 five initial engine cases (first1792 compile typo corrected before
validRED)→54341GREEN228units. Additional lifecycle tests initially24753full gate
failed existing adapter wire-byte assertion: JSONEncoder unsorted order. Explicit
independent sortedJSON assertion runtimeRED80149 reproduced; outbox-only sorted
encoder fix. Final52696exit0,232/232unitsPASS0skip,
`/tmp/mypocket-sync-engine-final2-20261014.xcresult`. Tests include noncooperative
late success after logout, explicit sameowner resume, backoff cap, and real
APIClient/FinanceAPI→receipt→SQLite timeout/replay (fake external HTTP only).
Fresh capability GET on bothdrains, exactkey/body and acknowledgedIDs persisted.
Actual network server financial E2E and real401refresh within engine still need
integration coverage (APIClient refresh independently covered). Coalescing code
exists but simultaneous-trigger behavioral proof remains open.

Full Go race22266exit0 dedicatedPG55434/Redis16379, no backend changes. Base/
guardtests/project/diff pass. XcodeGen included SyncEngine/tests from project.yml;
existing empty development team/config signing settings retained. No schema/API
wire changes or Swagger generation required. Human+agent offline guide updated.
Not app runtime wiring/cache/refetch/UI/foreground scheduler completion; injected
invalidation callback alone is not proof of real refetch. No prod/sign/install/
release/owneracceptance/fullgoalcompletion.
