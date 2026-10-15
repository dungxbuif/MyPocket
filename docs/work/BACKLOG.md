---
artifact_type: backlog
id: BACKLOG
status: active
owner: shared
human_fields:
  - priority_override
  - rank
  - blocker_decisions
ai_fields:
  - risk_flags
  - lane_recommendation
  - next_artifact
  - notes
shared_fields:
  - queue_items
  - status
updated: 2026-10-15
---

# Backlog

## Authoritative native queue — 2026-10-15

[Reconciliation matrix](tickets/NATIVE-RECONCILIATION-20261015.md) is the active
cross-ticket status index. Category280/280PASS and numeric/chat287/287PASS local;
broader12business gate live. Full Go race with realPG/Redis and vetPASS. Fresh
production48openfeedback; [newest six](tickets/NATIVE-FEEDBACK-NEW-20261015-DETAIL_DESIGN.md)
include fixed Feedback/Account Add/Back/spacing/portfolio/AI jar-tool requirements.
Priority: compatible scoped backend rollout before new native install; report/
cache and savings correctness; reviewed AI lifecycle/actions/help; full-screen/
feedback/owner UAT/release gates. No production financial or feedback mutation.
The jar create/update and budget create/update proposals are now reconciled in
the advisor handbook as explicit-confirmation capabilities; jar/budget delete,
other transaction writes and live-provider proof remain separate backlog work.
The shared Feedback dock now has a stable lower-right anchor and no obstacle
driven lane; full-screen/device overlap acceptance remains queued.
Assistant numeric/card safety is now locally verified by seven focused native
regressions (exact money/progress/count conversion, nontrapping fallback and
masking). Signed device/UAT and broader release acceptance remain open; this
does not change provider/API/database behavior.
Older checkpoint text and draft requirement rows below are historical/planning
state, not evidence that currently implemented native features are absent. The
advisor idle-retention foundation and explicit Chat mới reset are now local
code: migrations `000033`/`000034`, generation-safe access purge, bounded worker
pass, `410` mapping, an idempotent reset ledger, in-process cancellation of
purged provider runs and additive native/server expiry metadata; production
rollout remains open.

2026-10-15 verification: backend `go test ./...` and `go vet ./...`, native
history metadata tests, warnings-as-errors simulator build, base-component and
project guards, plus web production build/typecheck/design/unit suites pass
locally. Browser suites requiring a live Vite/fixture remain environment-gated;
no environment was started or changed. The bounded local `get_product_help`
retrieval and native/web `product_help` cards are also verified; durable action
replay, external-document RAG, and real-provider/device/production proof remain
open. The opt-in live provider/tool-result grounding probe now passes with
synthetic handbook data; it is not production RAG evidence.

2026-10-15 savings queue follow-up: the shared queue menu now evaluates edit
eligibility per item. Goal-wallet transfer drafts can open `Sửa bản nháp` with
wallet data alone; ordinary transaction drafts still wait for categories. Policy
unit proof: `/tmp/mypocket-goal-queue-policy-green-20261015.xcresult`. The
fault-injected savings 422 UI rerun is green against the disposable local
PostgreSQL/API/fault-proxy stack; evidence is
`/tmp/mypocket-offline-savings-20261015e/Logs/Test/Test-MyPocketIOS-2026.10.15_20-37-31-+0700.xcresult`.
No API, DB, environment, device or production state was changed.

2026-10-15 auth follow-up: Google mobile exchange now distinguishes invalid OAuth
profiles (`401 GOOGLE_MOBILE_LOGIN_INVALID`) from local owner persistence outages
(`503 GOOGLE_USER_STORE_UNAVAILABLE`). The earlier savings rerun was traced to
the disposable PostgreSQL dependency at `127.0.0.1:55434` not listening, not to
the fixture flag; the approved rerun later passed on a healthy local stack.
Use-case/HTTP/race tests and generated Swagger pass; the running API was not
restarted.

### Historical checkpoint log — not the active scheduler

20261015 owner-approved numeric/privacy correction locally verified287/287PASS
(279unit+8chatUI), in_review pending device/UAT. Release build awaiting Keychain,
iPhone unavailable. Category test3renewed attemptsRED, stop/review post-save lazy
row ordering; wallet/probe/transaction testsPASS.30minidleapproved, written
lifecycle design9ab9248d review pending; no deletion enabled. Fullscope retained.

20261015 resumed blocked audit: installed candidate remains available; no live
test/build/codesign/proxy, SSH identities0. Specific TTL/numeric/privacy/business
and previous architecture/placement decisions remain unanswered for3current turns.
Further product changes require input/access; fullgoal blocked, not complete.
Do not repeat existing verification or silently delete chat records to bypass
the pending lifecycle/idempotency design. Original backlog scope retained.

20261015 user requests idle chat deletion to avoid stale context. [Idle reset](tickets/NATIVE-ADVISOR-IDLE-RESET-20261015.md)
awaits interval; source audit requires preserving request replay and preventing
old-generation answers/actions after purge. No deletion implemented. Numeric
actual native RED1/1 confirms unsafe progress range acceptance; bounded fix review
still pending, installed iPhone candidate unchanged.

20261015 iPhone candidate installed and running PID22311 on owner iPhone15Plus;
verified production-API Release and corrected/verified code signature. Owner UAT
next. This satisfies requested immediate installation, not entire backlog or
production backend cancellation rollout. Preserve numeric/privacy and other gates.

20261015 priority override: owner asks quick iPhone candidate. Cancellation
bounded fix approved and locally green4/4 realDB plus full Go race suite; native
fresh273/273PASS. Production-connected Release signing/install in progress;
no backend deployment or full-backlog completion implied. Remaining numeric/
privacy, business/helper, Feedback/cache and initial Stop gates retained.

20261015 runtime goal blocked pending specific design replies/access, objective
unchanged. No live tests; SSH agent identities0. Pending numeric/privacy, cancel
transaction recheck, business pinned/AX refinement; Feedback/cache/goal422 and
initial Stop transport remain separate. Current verification is not full release
acceptance. Do not spend further turns repeating known checks or infer design
approval from automatic goal continuation. Resume on actual human/access change.

Priority cancellation correctness: [realDB race](tickets/NATIVE-ADVISOR-CANCEL-20261015.md)
was reproduced in 70798 (3PASS/1FAIL) and is now fixed locally by the approved
conversation→run lock/recheck; the green rerun is 4/4 in
`/tmp/mypocket-advisor-cancel-db-green-20261015.log`, with the full Go race suite
passing. This closes only the stale-status insert race. The initial iOS Stop
gap (synchronous POST does not return a run ID until the provider finishes),
deployment and broader completion/cancel atomicity remain separate gates. No
production change was made.

Persisted native history proof added:27531exit0 final274/274PASS0skip (273unit+
1actualDB/API/UI). Latest50→older1/nojump/scrollback/restart+APIunchanged verified;
synthetic51messages/26runs cleaned exactly, counts0/0/0. No product patch. Human/
agent guide+screen updated. Live cancellation/races and pending numeric/business/
Feedback/cache reviews still open; no release/prod/device delivery claim.

Native advisor adapter tests added under existing history/cancel verification
plan:2/2 focused PASS, final273/273unit PASS72133,0skip/no runtime warnings.
Actual FinanceAPI/client decoding/cursor/auth/cancelroute proven with transport
fixtures, not DB/provider/UI. Product unchanged; history scroll/merge/cancelrace
remain open. Numeric/business/Feedback/cache reviews still pending, no deploy.

Current-helper chat8/8PASS88653,0skip/no runtime warnings. Do not infer fullchat:
AXFeedback overlaps text, landscapecapture cropped, populatedhistory/cancel/live
provider still unproven. Source confirms50-message pagination bound alignment;
cancel/send task overlap remains source-only reproduction target. Numeric and
business bounded designs pending; no product/test implementation this pass.

Numeric followup now actual-source reproduced (Swift CLI): rawfallback1e100 and
countDouble(Int.max) SIGTRAP, valid controls12500/42PASS; progress bound admits
unrepresentable Int64. No native render/fix yet. Same fallback ignores masked
in source; public human/agent docs corrected. Existing bounded design review
pending; no product implementation or new auth/API permission change.

Historical next-slice note: [numeric safety](tickets/NATIVE-AI-CARD-NUMERIC-SAFETY-20261015.md)
is now locally verified by the focused native regression recorded above. The
source/Swift diagnostics below preserve the pre-fix provenance; they are not a
current pending implementation claim. Keep product/API/base unchanged pending
the separate signed-device/UAT gate; do not restart blocked business
or savings/Feedback placement tests without their separate review decisions.

Latest80693:273PASS/2FAIL0skip,271units+probe+walletCRUD PASS. Exact category name/
API scope now correct; UI assertion must inspect base accessibilityValue, filter
transition still unproven. New helper rejects pinned save outside shell but in
window; bounded pinned/AX refinement review pending, no more speculative patch.
Business ticket blocked for this review, fullgoal active. No prod/install/release.

2026-10-15 reveal helper design approved. RED54467 proves QuickType swipes mutate
the draft before focus/save; not evidence for disabling product autocorrection.
Test-only semantic ancestor/closed-failure correction implemented; full swipe
overshoot corrected by bounded drag. Probe34635PASS1/1; CRUD+unit80693 running.
Preserve all literal/data assertions and
old failures. Full business/goal422/whole-native acceptance remains open.

Latest diagnostic7487:272PASS/2FAIL0skip, correct-wallet transactionCRUD now PASS.
Typed names initially exact but later append prose in category/wallet CRUD.
No product fix yet; literal-name-only base contract awaiting owner decision.
Normal8081 launched25086, no faultproxy/tests active. Full business gate remains
RED; do not deploy from narrow chat/unit passes. Goal422 menu design/root review
blocked; goal503/credit scoped financial proofs passed, not fullvisual acceptance.

2026-10-15 current business sweep RED:280PASS/3FAIL0skip. Category/wallet exact
typed names changed before persistence; transaction test wrong first wallet.
Test scoping/root-type/early exact-input diagnostics7487 running, no product
keyboard fix yet. Validate owned malformed fixture IDs before cleanup; do not
purge by a broad prefix. [Work](tickets/NATIVE-CURRENT-BUSINESS-SWEEP-20261014.md).
Savings selector approved correction works; goal503+strong credit gate PASS,
goal422 opens wrong ordinary Add on menu tap and is blocked for review. No more
speculative hit-testing/Feedback changes; current full goal stays in_progress.

Savings helper review unblocked by explicit owner reply: test-only correction
implemented; await current business sweep before full503/422 goal verification.
Current sweep normal8081 runs12critical business/route tests+allunits; credit test
strengthened separately to choose its own source and verify pair/balances/report,
requires dedicated rerun (current binary predates test edit). No fullgoal claim.

AI audit current: focused advisor/proposal Go race tests pass in both packages.
Static handbook navigation is now reconciled to the primary Trợ lý tab with a
regression; bounded product-help retrieval and native/web cards are locally
verified, but not live RAG. Catalog8tool/6provider limits reconciled. Wallet
confirmation summary is now implemented locally; durable action replay and
full-goal matrix remain in_progress. No unapproved UI/action API implementation
was added.

AI wallet proposal validation local in_review: four malformed cards reproduced
rendering confirmations; strict decoder final279/279PASS0skip (271units+8chatUI),
no runtime warnings. [Work](tickets/NATIVE-AI-WALLET-PROPOSAL-VALIDATION-20261014.md).
Followup: other structured-card money/progress direct Int64 conversions remain
unsafe on malformed nonfinite/out-of-range provider data; reproduce independently.
The wallet proposal `action_id` now preserves history replay identity and the
confirmation summary includes goal/credit/total attributes; the full prepared
action dispatcher and live-provider proof remain open.

Aggregate/basic rejected transfer editor local in_review:269/269PASS0skip,
restart→unchanged refusal→22500/newkey→one pair, correct balances/unchanged report/
scoped detail. No product fix; savings helper gate remains blocked. AI catalog
source mismatch corrected; the general prepared-action dispatcher remains open.
Full goal remains active.

Adaptive chat approach approved; local in_review, final276/276PASS0skip
(268unit+8UI), normal/AX draft roundtrips and realAPI landscape63pt pass. Earlier
blocked-design entry below is historical. No other design gate is approved by
this reply; no release/install/production write performed.

2026-10-14 chatadaptive blocked designreview: final88928exit65,272PASS/2landscape
FAIL0skip (actualAPI0pt/fake7pt conversation). One-line cap ineffective/reverted.
[Work](tickets/NATIVE-CHAT-ADAPTIVE-20261014.md). Actual frames show fixedheader/
overview/composer/navigation exceed keyboard-reducedspace; largesttype tab visual
overflow remains despitefocusedcontrol pass. Keyboardchrome/AXnav approach sent
for approval, no unapproved UI change. Fullgoal, other review gates stayactive.

2026-10-15 fixed the reproduced Reports invalid-month stale-content bug: local
`ReportsMonthPolicy` rejects malformed/out-of-range `YYYY-MM` before network,
failed loads clear the prior summary/detail/trend, and unit plus authenticated UI
regressions pass (`/tmp/mypocket-report-month-unit-final-20261015.xcresult`,
`/tmp/mypocket-report-month-ui-final-20261015.xcresult`). The reviewed encrypted
Overview/Reports read snapshot is now implemented with exact-month, server-returned
report payloads and no local recomputation; the remaining gate is fixture/device
proof only. No production or DB change was made. [Bug](bugs/NATIVE-REPORT-INVALID-MONTH-20261015.md)
and [offline cache](bugs/NATIVE-OFFLINE-REPORT-CACHE-20261015.md).

2026-10-14 readcache audit drafted: [scope risks](tickets/OFFLINE-READ-SCOPE-AUDIT-20261014.md)
in Overview/Reports/Months and ownerbound aggregate getters; no runtime repro or
cache implementation claimed. Typed scoped snapshot approach awaiting owner
review→written spec/plan before implementation. Preserve full report correctness,
legacycache/data/basepalette; no smaller-cache completion claim. Fullgoal active.

2026-10-14 featureprivacy localin_review: shared class/status allowlist replaces
6reflectederrorprints and fixturedebuglabel. RED93397→66267exit0,272/272PASS0skip
(268unit+4chatUI),guards/project/diffpass. [Work](tickets/NATIVE-FEATURE-LOG-PRIVACY-20261014.md).
This supersedes previous reflected-error followup. Not all-screen injectedfault/
liveprovider/device proof. Goalqueue/Feedback review blocks unchanged; fullgoal
cache/auth/a11y/AIactions/help/APNs/prod/device gates remain.

2026-10-14 independent HTTPprivacy localin_review: DEBUG metadata-only, real
stdout RED→267/267PASS0skip (265unit+2APIUI), guards/project/diffpass.
[Work](tickets/NATIVE-HTTP-LOG-PRIVACY-20261014.md). Follow up reflected-error
logs in Assistant/Budgets/localfixture with actual flow repro; not wholeapp
privacy or fullgoal acceptance. Goalqueue/Feedback review blocks unchanged.

2026-10-14 goalqueue gate blocked Rule10 review:263unitsPASS/2UIFAIL at restart
selection (goal retained, test expects aggregate selector), not receipt evidence.
Review test-only helper correction before another cycle. [Work](tickets/OFFLINE-GOAL-QUEUE-20261014.md).
Savings refresh/rejected transfer UI/fullgoal remain open; no prod/device/release.

2026-10-14 latest recoveryE2E: ordinary422/409 changedonly/date, rejected/committed
discard, transfer503/drop onepair/balances/monthreportunchanged verified. Proxy5/5;
regression267/267; duplicate detailAXidentity RED→contain fix265/265; finalnormal
44956 265/265 (263unit+2APIUI),0skip. [Work](tickets/OFFLINE-RECOVERY-E2E-20261014.md).
Next rejectedtransfereditor/goalqueue/cache/report/auth/a11y/AIactions/help/APNs/
prod/device. CoveredledgerAX visibility and Feedbackdesign remain; no fullgoalclaim.

2026-10-14 latest lifecycle implementedlocal: foreground/SDKreachability/FIFO
timer + inactive suspension, notauthblock. Final263units +3fault/lifecycleUI0skip;
normalcreate/edit+transferAPIUI also pass. [Work](tickets/OFFLINE-FOREGROUND-LIFECYCLE-20261014.md).
Next rejectededit/discard/transferfault UI, goalqueue/cache/report/auth/a11y,
AIactions/help/APNs/prod/device gates. Radio-loss/deviceUAT and overlaydesign
remainopen; no fullTask5/6/7/goalcompletion or releasepublication.

2026-10-14 latest prerequisite: required ordinary category enforced at new
enqueue/replacement, not old immutable replay. RED67864/69099→final40360,
256/256PASS0skip,255units+native transferUI with actual wallet direction/balances/
reportflags/serverdelete. Guards/project/diffpass. [Work](tickets/OFFLINE-CATEGORY-VALIDATION-20261014.md).
Next foreground/reachability/timer/backgroundpause; then full fault/recovery/cache/
auth/a11y/AIactions/help/APNs/prod/device gates. No whole-goal completion.

2026-10-14 latest fault/compact gate: owner approved shared three-dot queue menu;
real503 + committed-lost-response restart/retry E2E2/2PASS0skip (73398), one key
and one server row each. Proxy3/3, guards/project/diffpass. Feedback still covers
content: test scrolls menu clear, not placement acceptance. See
[fault work](tickets/OFFLINE-FAULT-E2E-20261014.md). Fullrecovery/transfer/scheduler/
auth/background/a11y/prod/device and fullnative goal remainopen.

2026-10-14 recoverycode: RetryNow/localdiscard/atomicchangedreplacement/reused
editors local252unitPASS0skip. Pending/conflict/date/recoveryUI/a11y/networkrestart
E2E, goal-scopedqueuevisibility and scheduler/background/auth remainopen.
[Recoverywork](tickets/OFFLINE-QUEUE-RECOVERY-20261014.md), no prod/device/fullgoalclaim.

2026-10-14 newestqueue runtime: sharedcreateforms→durableownerqueue/receipt wired
local, separatependingbaseUI/refetch subscriptions. 243unitPASS0skip, actualcreate/
editUI1/1PASS0skip and encryptedSQLite1row. Peritemretry/discard/conflictedit,
pendingUI/networkE2E/scheduler/background/auth remain. No prod/device rollout.
[Runtimework](tickets/OFFLINE-RUNTIME-QUEUE-20261014.md), fullnativegoal open.

2026-10-14 newestcache gate: defaultAuth→ledger cachedread localimplemented,
239/239units+keyUI and ledgerAPIUI1/1PASS0skip. Encrypted owner/server snapshot,
stale+retrybase UI; status503{} classification fixed. Follow-up 2026-10-15 now
adds server-returned summary/insider fields and exact-month Overview/Reports
fallback (9/9 focused tests, Debug build/guards green). Stalebanner visual/
accessibility, fullruntimeoutbox/forms/queue/scheduler/auth/refetch remain.
[Work](tickets/OFFLINE-FINANCIAL-CACHE-20261014.md),
[report fallback](bugs/NATIVE-OFFLINE-REPORT-CACHE-20261015.md). Fullnativegoal open.

2026-10-14 newestTask5 checkpoint: SyncEngine actor/policy local232unitsPASS,
fullGoRacePASS, guards/project/diffPASS. Not composed into app; cache/refetch/UI/
foregroundscheduler and simultaneoustrigger/engine401integration proof remain.
[Work](tickets/OFFLINE-SYNC-ENGINE-20261014.md). No wholeoffline/goalcompletion.

2026-10-14 latest offline gate: receipt/resource HTTP adapter implemented locally,
223/223units pass0skip (`/tmp/mypocket-ack-http-final-20261014.xcresult`). This is
not full offline completion: sync actor, retry/cancellation/capability lifecycle,
cache/outbox UI/events, live E2E and deployment remain. See
[work ticket](tickets/OFFLINE-CLIENT-ACK-20261014.md). Full native all-gap goal open.

## Field Ownership

- Human owns priority overrides, rank changes, and blocker decisions.
- AI may recommend lane, risk flags, next artifact, and notes.
- Shared fields include queue rows and item status.

Use this file as the runtime work queue.

## Current Native iOS Work

Historical implementation checkpoints below. Use the authoritative matrix at the
top for current implementation/rollout/verification status; do not restart work
from superseded “not implemented yet” lines.

Actual API account key E2E complete local: fresh8081 API19887, create/copy/revoke/
leave/reopen. Fixed revoked row resurrection and lingering one-time secret via
transient keyID/filter, no API history deletion. Owner-bound HTTP preservesmeta;
refresh CAS prevents overwriting new login/logout across Keychain instances.
Final58452exit0,219/2190skip218unit+APIUI;guard/project/diffpass. [Next receipt
adapter](tickets/OFFLINE-CLIENT-ACK-20261014.md) remains in_progress, then sync/cache/
UI/allgap scope. No prod/install/release; bubble covers timezone in inspected
reopen render, protected-content design block unchanged.

Full-owner API keys local/in_review: common user auth, no scope gates/selector,
oldDBscope data preserved, fullaccess helper/selfkeylogout. FinalGoRace18813/vet0,
native212/2120skip97143. [Work](tickets/API-KEY-FULL-ACCESS-20261014.md). Production
policy unchanged until rollout; old keys gain fullrights then. Next updatedlocal
API/accountmutationUI proof, then sync/cache/UI and full goal. Feedback covers
helper in inspected SE render; protected-content review still open, no release.

FrozenTravel prerequisite locally verified211/211native0skip, fullGoRace51567/
vet0 + generatedSwagger. Optional create policy false, oldreceipt compatibility,
explicittrip/no-trip/fixeddate/foreigndeleted rejection and capgate; outbox freezes
context but no sync/UI connection. [Work](tickets/OFFLINE-FROZEN-TRAVEL-20261014.md).
Next full-owner API keys per fresh approval, then entire remaining goal.

Owner2026-10-14 decided API keys have full same-owner JWT-equivalent capabilities,
no scope system. [Implementation design](tickets/API-KEY-FULL-ACCESS-20261014.md)
ready; not implemented/deployed yet. Supersedes pending transactions:write choice
and old limited-key policy as target, not a claim that current middleware changed.
Preserve owner separation/revocation/expiry, no admin service-token privilege.

Latest native store gate:209/209unit0skip terminal88836,16SQLite/Keychain outbox
tests, persistence and leases RED→GREEN. Encrypted FIFO/exclusive sending/expired
recovery/stale-outcome/auth blocking implemented, not connected to UI/session/sync.
See [store work](tickets/OFFLINE-ENCRYPTED-STORE-20261014.md). Next delayed-Travel
contract/capability, sync actor/cache/UI; no prod/install/release. Keep full gap
scope, pending bubble design/write-key scope and provider/device/rollout gates.

Capability gate: `GET /sync/capabilities` is deployed in the current API image
under JWT/full-owner key; Swift known/version/key/meta predicate + authenticated
GET/old404 refusal, native 193/1930skip, Go race/vet pass. The native sync engine
still does not consume it and production `enabled:true` remains unproven without
an owner-authenticated probe.
Next encrypted owner-scoped cache/outbox/sync, retain server revalidation-before-
send/rollback safety and required-key rollout/write-scope decision/full-goal gates.

Actual JWT/session/PostgreSQL/Redis receipt integration verified (full Go race
48372/vet0); TTL/logout/owner checks before replay, fixtures cleaned. Financial
API keys still cannot write/replay; transactions:write decision requested, no
security expansion. Continue capability/encrypted native cache/outbox/sync and
all outstanding UI/AI/prod/device/review gates, not a goal completion.

Keyed HTTP create path verified local (4DB/HTTPcases, full Go race45501/vet0,
native189/189 compatibility units0skip). Keeps data shape, meta.mutation replay,
early receipt lookup/no resurrection, concurrent one-pair safety. No-header path
still legacy online-only, not final key-required rollout or offline availability.
Next: real auth/API-key E2E, server capability/activation gate and native encrypted
cache/outbox/sync. Docs/Swagger updated; no API restart/prod/device/release change.

Offline logging prerequisite locally verified: shared SQL boundary retains class/
rows/time/SQLSTATE without SQL text, bound values or driver-error echoes.4real DB
tests/full Go race6719/vet/diff pass; no prod activation or historic-log deletion.
HTTP key/metadata/hash integration then native encrypted cache/outbox/sync remain
next. UI/bubble/provider/device/full-goal scope unchanged, not completed.

Latest full-goal backend progress: offline create receipt capability + shared
replay lookup fix verified with real local PostgreSQL,5focusedcases, full Go/race/
vet pass. No new schema, handler wiring or production change; offline still not
available. [Work](tickets/OFFLINE-CREATE-RECEIPTS-20261014.md). Next HTTP metadata/
key/hash/privacy contract gate, then native encrypted cache/outbox/sync. Existing
UI/bubble/provider/device/rollout requirements stay open; not a smaller UI goal.

Portfolio secondary metrics now compact,8pt rhythm, verified191/191 (188unit+
3API UI), RED asset27.5pt vs summary20.5pt reproduced. Capture inspected; bubble
still covers asset identity. Keep full visual acceptance/blocked avoidance design
and broader goal open; no release/deploy/device install. No finance logic changed.

Latest independent Task6 slice: Advisor More duplicate removed, typed destination
retained. Unsent chat draft survives tab switches in volatile shell memory. Final
navigation gate41071 exit0,196/196 PASS0skip (188unit+8UI), retained draft render
inspected; fake-token auth error is expected fixture state, not prod failure.
Full Task6/goal not complete: bubble design review, density, provider/a11y/device
and broader offline/AI actions/RAG/APNs/rollout scope remain. No release/install.

Latest: pinning/avoidance selected gate196/196 PASS,0skip, but inspected proposal
capture exposes Feedback covering the description field. Block this slice for
design review (third product-fix pass), not the full goal. Header/Analyze/Save-all,
preview/remove/approval tests pass; not universal non-overlap or UI acceptance.
Continue independent Task3 evidence/all-screen/chat and non-UI work. No rollout,
device replacement or release. See completion ticket and validation for evidence.
Independent current-source management regression7/7 PASS0skip (55284 exit0):
budget/ended-budget/jar/recurring/travel CRUD/date/link-unlink. No Task3 completion
inferred; remaining failure/visual/accessibility coverage still required.

Task4 complete locally: final193/193 zero skip at
`/tmp/mypocket-task4-complete-20261014.xcresult`,187unit+6UI including real
selected-photo append, retained19 prior files,20 cap, source/system cancellation,
capability and uncertain recovery. Loader/mixed multipart boundary proofs pass;
no live OCR/cloud/provider acceptance claim. Unsigned Release compiles, not
installed. Task3 remaining create/error/capture, then full Task1/5/6 proof and
non-UI goal work continue. New Task6 AI-entry findings: scrolling header/Save-all
and bubble partly covering process CTA. No release/status acceptance or prod change.

Active next: Task6 AI entry header/Save-all/Analyze pinning validRED0/2, no new
scaffold implementation yet. Preserve shared title/onCancel API via typed-header
overload; protect both Analyze and last attachment X from Feedback without lane/
spacer. Added last-X assertion pending run; use shared avoidance if necessary.
Task3 remaining evidence and entire non-UI/full-device scope stay open.

Newest: Photos-first source/image-policy base added, source/system Photos+Files
cancel and20-cap/thumb proof pass; final189/189 (185unit+4UI,0skip) and inspected
renders. Blank sheet caused by unbounded absent footer; bounded in base. Task4
still open for actual selected-photo import/partial/mixed upload/device proof.
Credit payment row now retains its specific title with neutral transfer semantics;
actual API UI verifies12000 purchase/7500 payment/pair/report exclusion. Task3
remaining create/error/capture and full conformance scope is not waived. No release,
prod activation or device replacement. Full offline/AI actions/RAG/APNs goal active.

Current: retain canceled same-item drafts in four management editors; invalidate
jar draft on month change, refresh server eligibility metadata. ValidRED4/4 for
loss and focused Jar month RED observed; final gate20421 terminal188/188 PASS,
0 skips (180unit+8UI). Ended-budget preview rendered/inspected; final ledger
field-equality proof passes. Docs/registry reconciled; Task3 portfolio/credit
regression and remaining create/error/capture scope must not be skipped before
Photos and full all-screen/offline/AI/RAG/APNs/prod/device work.

Continuation checkpoint: four management editors implemented with common pinned
actions; optional Travel calendar and explicit-clear wire adapter implemented.
Keyboard hides Feedback without reserving space. Units180 passed in mixed runs;
recurring CRUD and chat/typing regressions recovered. Travel context UI test is
now passes1/1 with API context/date readback; final aggregate187/187 passes
(180 unit+7UI,0 skips), explicit Jar follow-up1/1 passes; later188/188 proves
canceled-edit retention/month isolation and ended-budget read-only. Full goal still includes
all UI tasks, offline, AI actions/RAG/APNs, production activation and device UAT.

Latest execution: Task2 typed fields/note complete locally (179 units; API
fractional trade, interval, report exclusion, note write/cancel proof). Task3
portfolio editor slice verified with actual API failure/cancel/save and screenshots.
Next implement common-editor flow for budgets, jars, recurring and travel, then
Photos-first attachments and full screen/accessibility/provider sweep. Asset-list
metrics still tall on SE despite flat cards; re-evaluate density/disclosure in
Task6 rather than claiming visual acceptance from passing actions alone.
Release compile passed unsigned; no new device install/release/prod change.

Authoritative execution update: owner approves all gaps; [full completion matrix](tickets/NATIVE-FULL-COMPLETION-20261014.md)
preserves offline/AI actions/RAG/APNs/rollout/device requirements. Chat Task5 core
implemented/verified on SE (UI3 + API overview1 + unit171). Task1 asymmetric
surface insets/flat metrics underway; remaining task evidence and full native
UAT are not complete. Do not repeat the historical awaiting-plan-review state.

Owner has approved common native guideline, explicitly keeping app colors.
Execution: [six-task execution plan](../superpowers/plans/2026-10-14-native-ui-system-consolidation.md)
is approved and underway across all registry screens. No new style selection
or alternate component hierarchy. Current feedback rollout awaits SSH/Keychain.

2026-10-14 latest priority: whole-native screen/base audit and8 new feedback.
[Registry](../design/platforms/ios-screen-registry.md) maps all named screens and
editor/presentation states; [guideline draft](../design/platforms/ios-base-guideline.md)
defines five approved common families. Duplicate account/inbox
menu entries fixed with UI proof. Photos-first, portfolio editor, chat layout,
money semantics and physical Face ID reproduction remain next.

[Owner resolution](tickets/FEEDBACK-OWNER-RESOLUTION.md) feature passes backend
race/native171/API UI1. Migration32 on prod clean and preserved data; image
activation awaits SSH/1Password approval; final codesign awaits Keychain.
Do not install new app against absent prod route or claim fully deployed.

Urgent owner follow-up 2026-10-14: installed card/items still too tall. Native
density bases revised to card padding8, control44, transaction/wallet row44
minimum without outer vertical padding, one profile row, section gaps8. Disabled
Account FAB hidden. Owner explicitly requests skipping tests and immediate
installation; previous test counts predate this density pass. Preserve release
acceptance gate; API/data unchanged. Build/sign/install evidence belongs in
VALIDATION_MATRIX before handing off the replacement device build.

Density replacement built/signed/installed successfully on owner's iPhone;
tests intentionally not executed. Pending owner visual acceptance, not released.

Authoritative checkpoint 2026-10-14: compact shared bases and overlay-to-navbar
layout are implemented and verified (169 unit + 4 compact UI; 8 API-backed UI).
The production queue was fetched successfully: 34 records classified in
[audit](tickets/NATIVE-FEEDBACK-PROD-AUDIT-20261014.md), not all fixed.
Auth-only backend was approved, backed up and deployed with migrations31 clean
and unchanged financial state; see [rollout](tickets/NATIVE-DEVICE-AUTH-ROLLOUT-20261014.md).

Next: owner device UAT. After Mac unlock, the same signing identity repaired a
stale interrupted-signing manifest; strict verification passed. Production-API
native app installed without uninstalling and launched on the paired iPhone.
Keep note-editor separation and disabled Account FAB visual gap open; older
offline/cache, AI V2 writes/RAG and APNs feature requests are not part of the
verified UI fixes. No release note or feedback status transition before owner
acceptance. Current Personal Team profile expires 2026-10-20; earlier missing/
expired certificate and Ad Hoc-only statements below are historical, superseded
by owner's later Personal Team choice.

### Historical checkpoint — 2026-10-13

Owner refinement 2026-10-13: reduced rounding is for buttons only. Shared
native button token is 8pt with unchanged 48pt height and 44pt hit targets;
unit verification passes 164/164. Keep the task in progress: replace the
reserved feedback lane with a bubble overlay so scroll reaches the navbar,
then review card/content density on actual native screens. Previous dock
approval is superseded by this latest owner direction; do not treat prior
dock collision tests as acceptance of the newly requested layout.

The 2026-10-13 feedback-dock implementation passes focused iPhone 16 and iPhone
SE collision/capture, scroll-boundary, shell-action, unit, Release, project and
base-component gates; authenticated owner visual UAT remains open. See
[validation](VALIDATION_MATRIX.md) and the
[collision bug](bugs/NATIVE-LIST-FEEDBACK-BUBBLE-COLLISION.md).

The owner requested that the remaining production feedback be fixed on the
native iOS app only. Intake is recorded in
[NATIVE-MOBILE-FEEDBACK-20261013](tickets/NATIVE-MOBILE-FEEDBACK-20261013.md).
Audit current code before treating a report as open; work through financial and
auth correctness, ledger workflows, shared-base visual polish, then screenshot
feedback. No web/backend/deployment changes; no release note until owner
approval. Production feedback route currently returned 404 on direct open, so
the live queue must be re-fetched before claiming all feedback is exhausted.

The current-source 148/148 unit suite and unsigned Release builds for Simulator
and device pass. Physical-device Ad Hoc installation is not yet available: the
paired iPhone 15 Plus is connected, but the local Apple Development certificate
expired on 2025-01-12 and no provisioning profile is installed. See the
[port audit](PORT-100-AUDIT.md) for the exact signing gate; do not substitute a
7-day Personal Team build for the owner's approved Ad Hoc distribution choice.

Historical checkpoint (2026-10-08, before the dock implementation): native
route verification was corrected and rerun. The earlier 15-route
capture used a synthetic in-memory session with no local API listener, so its
screenshots showed network errors and are not valid rendering evidence. A fresh
`LOCAL_API_UI` build then completed the same 15-route/title/back sweep against a
version-30 blank PostgreSQL database and dedicated loopback Redis container;
exported screenshots show valid empty states for the non-provider routes.
Finance Assistant/OCR provider behavior remains unverified because credentials
were intentionally disabled. Test DB/container were removed; stage, Redis DB 15,
8080 and production were untouched. At that checkpoint screenshots still
showed the feedback bubble overlapping long content; the approved dock
implementation above resolves that collision. See [validation](VALIDATION_MATRIX.md) and
[port audit](PORT-100-AUDIT.md).

2026-10-08 Reports audit fixed a financial correctness bug: `Create` previously
lost an explicit `included_in_reports=false` because of GORM/PostgreSQL default
handling. Repository readback and full backend suite now pass; API-backed iPhone
16 Reports E2E confirms excluded amounts stay out of totals and the trend chart
VoiceOver value exposes daily totals with masking. No public contract or schema
changed. The full port and owner/device acceptance remain open; see
[validation](VALIDATION_MATRIX.md) and [audit](PORT-100-AUDIT.md).

2026-10-08 Travel Mode E2E gap closed for manual transaction association:
native detail can attach and clear an ordinary transaction, and an API-backed
iPhone 16 test verifies `travel_event_id` persistence after both writes. Existing
backend auto-link/clear contract tests pass; no backend/API change. This does
not close remaining Travel Mode acceptance questions or the full port. See
[ticket](tickets/TICKET-05-03-su-kien-travel-mode.md),
[validation](VALIDATION_MATRIX.md) and [port audit](PORT-100-AUDIT.md).

2026-10-08 native category-picker root-card parity is implemented. iOS now
renders a separate shared surface per root, matching the web selector; pure
tests preserve root/child order and search ancestry, and API-backed budget CRUD
verifies category-ID persistence (1/1 on iPhone 16 Simulator). No backend/API
change; physical-device/owner visual acceptance and the independent feedback
bubble collision remain open. See the
[port audit](PORT-100-AUDIT.md) and [validation](VALIDATION_MATRIX.md).

2026-10-08 populated savings-ledger regression is fixed: goal detail is now
the root route in Transactions rather than a nested ScrollView. Local
API-backed iPhone 16 proof confirms the real row scrolls fully above its fixed
CTA (2 focused routes pass); the disposable wallet/transactions were cleaned
up. This is a UI composition issue, not a backend/API gap. The independent
feedback bubble collision is resolved with the approved shell dock; compact/full
verification and owner-authenticated visual acceptance remain separate gates.

2026-10-08 backend verification refresh: DB-backed tests initially exposed
that the local stage database was one additive migration behind source
(version 29 vs. required version 30). Applied existing migration `000030`
locally without reset/data loss; the stage DB now passes
`go test -race ./... -count=1` (265 tests, 16 packages). A fresh disposable DB
also passed and was removed. No production access or change occurred. Current
native unit/build evidence and the still-open feedback-bubble UI regression
are tracked in [validation](VALIDATION_MATRIX.md) and the [port audit](PORT-100-AUDIT.md).

2026-10-08 iOS release packaging now includes the 1024px AppIcon derived from
the PWA wallet mark. The project guard asserts the asset catalog wiring and
dimensions; a Release build and compiled `Assets.car` inspection verify the
icon is shipped. Feedback-dock collision, populated-state visual review and
device UAT remain open.

2026-10-08 native credit-entry category parity is implemented: the existing
credit action sheet switches in-place to the shared searchable hierarchy,
preserves its draft and returns on selection. The iPhone 16 API-backed flow
verified category-ID persistence and the existing entry/payment actions against
the isolated local stage API. Backend route, scope validation and payload are
already compatible; no backend or migration change is needed. See the
[wallet design](../design/pages/wallets/README.md),
[detail design](tickets/NATIVE-CATEGORY-PICKER-PARITY-DETAIL_DESIGN.md),
[validation](VALIDATION_MATRIX.md) and [port audit](PORT-100-AUDIT.md).

2026-10-08 notification mark-read UI proof is complete: the local-stage
authenticated simulator test loads one synthetic unread row, taps the native
read action and confirms persisted state through the owner-scoped API. The
fixture is deleted afterward; no public notification-create route or
production DB write was added. The temporary API is isolated on port 8081 and
Redis DB 15 because the existing port 8080 process uses the default local
`postgres` database. The configured DeepSeek text extraction and provider
tool-call wire probes pass with synthetic data; a Finance Assistant stage
send/history run also passes and persists usage metadata before full fixture
cleanup. The remaining audit gates are OCR/PDF proof (blocked because local
`OCR_API_KEY` and S3 credentials are missing) and authenticated pixel-level/
device UAT. All native API fixture helpers now default to the isolated stage
API (override via `MYPOCKET_TEST_API_BASE_URL`); a wallet create/edit UI test
passed against that API and left zero matching synthetic wallets after
teardown. The existing 8080 API was not modified.

2026-10-08 full-suite follow-up: after making the unread-notification fixture
repeatable, the complete sequential iPhone 16 unit + `LOCAL_API_UI` suite passes
51/51 with zero failures against the isolated stage API (1,730.041 seconds;
xcresult recorded in [validation](VALIDATION_MATRIX.md)). The exact synthetic
notification row was removed after the run and API 8081 stopped. This closes
the local automated-suite gate for this checkpoint, not authenticated device,
populated-data, production-provider or pixel-level owner acceptance.

2026-10-08 portfolio trade audit found native reading history from the
aggregate `summary.trades` field, which the backend does not populate. Native
now loads each asset's history through the existing owner-scoped trades route.
The simulator verified buy → updated position/history and fresh API state
(side, fractional quantity, unit price, fee) 1/1. The immutable synthetic
trade/asset was removed from the disposable stage database by exact owner and
asset ID; no API/schema change was made. See [validation](VALIDATION_MATRIX.md)
and [port audit](PORT-100-AUDIT.md).

2026-10-08 ordinary-ledger wallet scope now uses the web-matching centered
capsule directly under the header in native instead of the generic form
selector row below period controls. It is backed by DesignSystem selector/glyph
components and keeps wallet scope behavior intact. Native tests (129/129),
Release warnings-as-errors, API-backed simulator route/screenshot (1/1) and
base/project guards pass. Disposable stage fixtures were cleaned; API 8081 was
stopped and API 8080 left untouched. See [transaction screen](../design/pages/transactions/README.md),
[validation](VALIDATION_MATRIX.md), and [port audit](PORT-100-AUDIT.md).

2026-10-07 unified visual-port slice: owner requested the attached Stitch/Money
Lover-inspired UI to be ported to both native iOS and web, while preserving the
existing MyPocket API and all extended finance/AI/feedback features. The archive
is available at `/Users/dungxbuif/Downloads/stitch_my_pocket.zip` and its
`financial_clarity_system/DESIGN.md` tokens are now mirrored by both clients.
Shared reference cards/bases cover the balance, wallet, Money Insider, monthly
report and center-add hierarchy. The screen inventory and behavior mapping are
recorded in [STITCH-MY-POCKET](../design/references/STITCH-MY-POCKET.md).
Bounded native UI reruns and local API-backed sweeps now pass; remaining proof
is authenticated owner visual UAT plus live provider/device gates, not a
missing design input. The native login screen now follows the owner's later
password-only decision; its launch test proves the Google action/divider are
absent. The iPhone 15 Plus is visible to Xcode over the network, but device
install is waiting on an Apple Developer account in Xcode and an App Development
provisioning profile for `com.mypocket.ios`.

2026-10-08 core-flow checkpoint: native API-backed Simulator UI verifies
wallet create/update, wallet-detail transaction create/update, atomic transfer
create/delete, and report trend values from persisted eligible rows (4/4).
Evidence: [validation matrix](VALIDATION_MATRIX.md) and
[port audit](PORT-100-AUDIT.md). This does not close the full port goal or the
separate feedback-dock, visual owner-acceptance, and physical-device gates.

2026-10-05: owner approved a native SwiftUI port of the current frontend and
requested iOS Simulator-first validation. [NATIVE-IOS-01](../superpowers/specs/2026-10-05-native-ios-swiftui-design.md)
and its [implementation plan](../superpowers/plans/2026-10-05-native-ios-swiftui-port.md)
define the isolated `codex/native-ios-port` lane. Keep `app/` and the Expo
migration branch unchanged. No production deployment or production database
mutation is part of this phase. The first blocker to check is whether Xcode
has an installed iOS Simulator runtime.

2026-10-06 native parity audit fixed two presentation/input defects: wallet
detail month labels no longer expose interpolation source text, and Travel
Mode validates trimmed names plus optional calendar ranges before submission.
Both have regression tests and remain limited to the native client.

2026-10-06 native ledger audit: removed the incorrect all-time period from the
SwiftUI ledger, added account-local Monday–Sunday week navigation and matched
the web `Tiền vào`/`Tiền ra` cashflow card. Goal and credit specialized scopes
remain unchanged; owner visual UAT is still separate.

2026-10-06 native date audit: goal-wallet target dates now round-trip through
the account IANA timezone as calendar dates; simulator regression proof is
recorded in the validation matrix.

2026-10-06 native feedback audit: the floating feedback trigger is hidden for
one rendered frame before the automatic screenshot is captured, preventing
the control from appearing in its own attachment. [NATIVE-FEEDBACK-BUBBLE-REDACTION](bugs/NATIVE-FEEDBACK-BUBBLE-REDACTION.md)
is regression_verified; the dedicated UI test and full simulator gate pass.

2026-10-06 final simulator/provider pass: the booted iPhone 16 Simulator is
green at 65/65 native tests and 17/17 signed API-backed UI tests. Backend
race/vet is green. The configured DeepSeek-V4-Flash S3 → OCR Base64 → LLM
probe is green at 5/5 cases and 24/24 semantic fields; receipt images now
default to expense classification unless explicit income/refund evidence is
present. Physical device install/signing and authenticated staging/production
remain external gates. The native Google login path was subsequently removed
per the owner's migration decision; backend/web OAuth remains separate.

## Current Owner Direction

2026-09-24 wallet/ledger UI follow-up: savings detail has an existing goal/progress spec; Travel Mode is still an event-linkage draft and not a new wallet type. [UI-LEDGER-WEEK](tickets/UI-LEDGER-WEEK-DETAIL_DESIGN.md) is in_review after owner correction: basic uses the same Week/Custom summary/list UI as `Tổng cộng`, filtered by wallet; goal shows complete savings history, and provisional credit remains read-only. Basic selection and week navigation are implemented and verified on the local Chrome page; goal selection/mobile visual UAT remain open. The web credit statement/payment contract remains a separate product gap; the native port now has a tested credit route and must not be regressed. No deployment has been performed for this slice.

2026-09-23: owner requires AI entry to suggest reviewable drafts, auto-merge clear repeated screenshot rows and prefill inferred/default wallets rather than refusing the batch. [Review-first correction](bugs/AI-ENTRY-REVIEW-FIRST-DETAIL_DESIGN.md): in_review, regression_verified locally. Proposal count is no longer capped in parser/usecase/repository; only provider response/transport envelopes remain. Nine stored OCR attachments yield 30 drafts; held-out overlap/partial-row and 31-draft tests pass. Go tests/race/build pass and local API restarted. Next: owner upload/review UAT with a new request; do not claim all source rows reconciled or production deployed.
2026-09-23: diagnostics/usage follow-up adds privacy-safe AI stage logs, optional internal model-usage persistence (`AI_STORE_USAGE`, enabled by default), explicit `user_instruction` filtering guidance, and a user-isolation audit. Streaming is deliberately deferred until an async/polling contract is separately designed; current JSON extraction remains bounded at 180s model/210s overall.

2026-09-22 follow-up: detailed [technical instructions](../superpowers/plans/2026-09-22-finance-assistant-technical-guide.md) accompany the V1 plan. The local implementation now has query/history/tools/JWT/API-key JSON/overview/UI and best-effort Redis audit proof; jar-read side effects, deterministic context anchors, submit recovery/drilldown, SSE, durable audit delivery and browser/provider gates remain open.

2026-09-22 latest turn: owner requested implementation of a separate comprehensive Finance Assistant tab. [AI-ADVISOR-01](tickets/AI-ADVISOR-01-DETAIL_DESIGN.md) separates current one-shot AI entry from persistent advisory chat, with a local V1 read-only slice and remaining release gates. Feedback → Fix → Changelog is locally wired; its browser/production proof remains separate. Public third-party advisor launch depends on Redis-audited provider/browser/deployment evidence; local user API-key authentication is now implemented, not Feedback's service token.

2026-09-21 latest UI follow-up: follow the Money Lover AI entry layout with shared composer/result bases while keeping explicit edit/remove/save-one/save-all behavior. Remove the old transaction group list and reuse BaseCategoryTree selection mode across transaction and budget group pickers. [UI-FORMS-03](tickets/UI-FORMS-03-DETAIL_DESIGN.md) is active; no automatic ledger writes are introduced by the visual reference.

Earlier 2026-09-22 AI follow-up: owner approved AWS SDK Go v2 for S3 presigning. [BUG-001](bugs/BUG-001-s3-presigned-get-signature.md) is verified: private S3 readback and live OCR pass. The initial JSON-object-mode evaluation scored 0/15; the later strict JSON Schema rerun and current status are recorded below and in validation.

2026-09-22 LLM follow-up: strict OpenAI-compatible JSON Schema output fixed the Qwen extraction evaluation (synthetic 5/5 cases, 24/24 fields); current backend sends optional wallet descriptions as bounded untrusted context. Keep AI-ENTRY-01 in review pending owner UAT. Owner also requested future provider/model selection with visible, sourced prices; [AI-ENTRY-03](tickets/AI-ENTRY-03-PROVIDER-SELECTION.md) and its draft design capture this for review only. No extra provider or billing integration is authorized.

2026-09-22 AI usage follow-up: owner clarified there is no user usage limit at this time. [AI-USAGE-01](tickets/AI-USAGE-01-DETAIL_DESIGN.md) records the decision; the current implementation should not enforce a per-user AI/OCR request cap. Future cost consent, pricing, and usage-budget choices stay in [AI-ENTRY-03](tickets/AI-ENTRY-03-PROVIDER-SELECTION.md).

2026-09-21 implementation follow-up: owner explicitly requested the hold-Add AI input/review-list slice. [AI-ENTRY-01](tickets/TICKET-09-01-ENTRY-DETAIL_DESIGN.md) records that approval and implementation. Real extraction fails strict output-schema validation; private attachment lifecycle was subsequently delivered under AI-ENTRY-02 and remains in browser UAT. Broader Q&A/paired transfer/hũ are not claimed delivered by this slice. This supersedes the earlier plan-only scheduling note below for entry review scope only.

2026-09-21: current task is assessment and planning for the two AI flows, requested for owner review. [DESIGN-09-AI](tickets/TICKET-09-DETAIL_DESIGN.md) is `in_review`; no implementation approval inferred. Proposed sequence: shared transfer/confirmation foundation → text input → OCR/multi-image reconciliation → read-only Q&A → pilot UAT. Jar APIs remain a linked independent slice. Small task exemption: yes for docs-only plan authoring. Reason: review artifact without runtime changes. Impact checked: API=no, DB=no, Security=no, Runtime=no, Standards=no.

2026-09-20 latest: owner approved savings continuation and requested no mock data in screens. [API-SCREENS-01](tickets/API-SCREENS-01-DETAIL_DESIGN.md) delivers budget explicit-date API/persistence/progress, mounted mock removal, savings catalog/history and error parsing; review pending. [Savings](tickets/TICKET-06-01-DETAIL_DESIGN.md) external-entry clarification resolved, no counterpart required. The shared managed-wallet transfer editor is now implemented in web and native; remaining work is owner visual/device UAT. Recurring budget roadmap remains partial. Historical pending-select and counterpart questions below no longer schedule work.

2026-09-22: continue approved [CORE-03](tickets/CORE-03-TIME-JARS-MONTH-DETAIL_DESIGN.md) timezone/jars/month screens through reconciliation and owner UAT. Implementation and automated checks are present; do not mark complete until migration/API/UI behavior is tested by the owner. Owner also requested a dedicated cron worker sharing backend code, with recurring transactions and month-end report/close as examples. [WORKER-01](tickets/WORKER-01-cron-service.md) is captured as high-risk draft; detail design is gated on report artifact semantics and recurring-job edge rules.

2026-10-08 wallet management parity: native now matches the web's included/excluded groups, included-only aggregate, Sửa/Xong mode, row-to-edit flow, confirmed API deletion, add/service actions, and preserves the native Chi tiết extension. API-backed simulator test exposed and fixed persistence of explicit `is_in_total=false`; schema/API contract is unchanged. `go test ./...`, 130 native unit tests, API UI 1/1 (including persisted deletion), Release warnings-as-errors and base/project guards pass. Owner visual/device acceptance remains pending.

Active: TICKET-06-01 in_progress, [design/proof](tickets/TICKET-06-01-DETAIL_DESIGN.md). Owner authorized savings UI implementation from five references. Goal create/date/details, common category picker and managed-wallet transfer editor are implemented; authenticated visual/device acceptance remains pending. UI-WALLET-02 selector layout, grouped management list and create BaseSelect are implemented; full acceptance pending.

Latest docs normalization: canonical wallet screen now defines list/create/edit, BaseSelect-only type choice during creation and per-type gaps. UI-WALLET-02 remains in_review for authenticated visual/browser UAT; implementation includes the existing-wallet selection and downstream scope/filter contract. Documentation changes do not authorize new API/schema scope.

Current priority: [UI-WALLET-02](tickets/UI-WALLET-02-DETAIL_DESIGN.md), `in_review`: complete authenticated owner visual/browser and physical-device acceptance for normalized Add Wallet and Wallet Selector references, grouped wallet management, base composition, Add Wallet form/type selection and existing-wallet scope/filter flow. Automated local-stage API-backed behavior passes; owner visual review remains pending. Existing-wallet selection/filter behavior is implemented, not a planned downstream slice.

Wallet UI follow-up: owner requests no note field in Add Wallet. Implemented in shared WalletEditorForm; in_review. Small task exemption: yes. Reason: owner-directed removal of optional create-form field. Impact checked: API=no, DB=no, Security=no, Runtime=no, Standards=no. Reuses FormField/BaseTextInput; existing edit data preserved. Proof: validation matrix; release: changelog; design: TICKET-01-02-DETAIL_DESIGN.md.

UI-EMPTY-01 follow-up includes wallet empty messages in Overview and wallet management; retains `in_review` pending visual acceptance.

2026-09-20: owner reports the bugs they tested are OK and asks to continue implementation. [UI-EMPTY-01](tickets/UI-EMPTY-01.md) is `in_review`: transaction empty messages now use a borderless shared base variant; automated design/build checks passed, owner visual review pending. This does not close unimplemented receipt/jar/transfer/credit scope. Next product slice: prepare transfer/adjustment detail design for review.

Latest request 2026-09-13: normalize design to Markdown contracts, add automated base-first checks and refactor the existing frontend. [UI-BASE-01](tickets/UI-BASE-01-DETAIL_DESIGN.md), from [FB-002](FEEDBACK_LOG.md), supersedes the earlier screen-removal proposal. Ledger remains the next product slice after this UI foundation work.

Naming đã chốt: bảng tài khoản là `user`. Khi triển khai migration, rename bảo toàn dữ liệu và đổi GORM mapping đồng bộ theo [ERD](../architecture/ERD.md); hiện chỉ cập nhật thiết kế.

Quản lý nhóm [TICKET-01-03](tickets/TICKET-01-03-tong-vi-danh-muc.md) đã hoàn tất: catalog system và nhóm cá nhân, cây cha–con, create/edit/delete cá nhân, và chọn ví áp dụng. Tiếp theo là giao diện ledger cơ bản; không đưa Reports/quick-add vào runtime trước khi có API.

Ưu tiên hiện tại: review [DESIGN-01-02](tickets/TICKET-01-02-DETAIL_DESIGN.md) cho quản lý ví trên branch `feature/wallet-management`. Đã chốt xoá thực và trùng tên; nghiên cứu Money Lover xác nhận adjustment tạo giao dịch mới. Bộ nhóm mặc định dùng lại catalog app cũ. Các quyết định này thay ghi chú chờ chọn hành vi xoá ở bản trước; chưa duyệt migration/implementation ví.

Review the complete contract and its [11 parent / 31 child BA tickets](tickets/README.md). The owner requested concise business-language tickets; all are `draft`, without an implementation phase or technical plan. The ticket list covers REQ-01 through REQ-18; common quality rules apply across the groups. Existing queue rows below remain historical context and do not override this review focus.

Documentation checks and remaining product proof are recorded in [VALIDATION_MATRIX.md](VALIDATION_MATRIX.md). Standards now live at the required `docs/standards/` location, with their existing content preserved.

## Current Business Ticket Queue

| Group | Review item | Status |
| --- | --- | --- |
| TICKET-01 | [Tài khoản, ví và danh mục](tickets/TICKET-01-tai-khoan-va-vi.md) | draft |
| TICKET-02 | [Giao dịch](tickets/TICKET-02-giao-dich.md) | draft |
| TICKET-03 | [Budget](tickets/TICKET-03-budget.md) | draft |
| TICKET-04 | [Hũ chi tiêu](tickets/TICKET-04-hu-chi-tieu.md) | draft |
| TICKET-05 | [Định kỳ và du lịch](tickets/TICKET-05-dinh-ky-du-lich.md) | draft |
| TICKET-06 | [Tiết kiệm, tín dụng và nợ](tickets/TICKET-06-tiet-kiem-tin-dung-no.md) | draft |
| TICKET-07 | [Báo cáo và tổng kết](tickets/TICKET-07-bao-cao-tong-ket.md) | draft |
| TICKET-08 | [Danh mục tài sản](tickets/TICKET-08-danh-muc-tai-san.md) | draft |
| TICKET-09 | [Trợ lý AI](tickets/TICKET-09-tro-ly-ai.md) | draft |
| TICKET-10 | [Kết nối công cụ](tickets/TICKET-10-ket-noi-cong-cu.md) | draft |
| TICKET-11 | [Sử dụng và dữ liệu](tickets/TICKET-11-su-dung-du-lieu.md) | draft |

Group numbering is for navigation, not an assigned implementation priority. Each parent links to its children and the relevant product requirements.

The backlog decides what should be worked on next. It does not replace tickets, bugs, requirements, phases, detail designs, or verification artifacts.

## Queue Rules

- Backlog order MAY change at runtime based on priority, severity, dependency, or new information.
- Bugs MAY preempt feature tickets when severity or user impact is higher.
- New requirements MAY enter the backlog before they become requirements docs, phases, or tickets.
- Maintenance and framework work MAY appear in the same queue as product work.
- The active queue focus MUST be reflected in `docs/CONTEXT.md`.
- Do not process all tickets before bugs by default; process the highest-priority queue item that is ready and appropriately scoped.

## Intake Rules

- Keep each item short and actionable.
- Link to source context, requirement, phase, ticket, bug, or decision when available.
- Classify each item by lane: `tiny`, `normal`, `high-risk`, or `blocked`.
- Promote non-tiny backlog items into a requirement slice, phase, ticket, or bug before execution.
- Do not execute directly from backlog unless the item is clearly tiny and records a small-task exemption.
- High-risk items require detail design approval before implementation.
- Blocked items must record the missing decision, dependency, or input.

## Lane Guide

| Lane | Use When | Required Next Artifact |
| --- | --- | --- |
| tiny | Low-risk docs/copy/naming/narrow edits with no contract or runtime impact | Backlog row may be enough with small-task exemption |
| normal | Bounded story-sized work, bug fix, or maintenance task | Ticket or bug |
| high-risk | Auth, authorization, data, security, public contract, external provider, migration, major dependency, or multi-domain impact | Ticket/bug plus detail design and approval |
| blocked | Work cannot proceed because input, decision, dependency, or environment is missing | Blocker note and owner |

## Risk Flags

Mark risk flags in the `Risk Flags` column when relevant:

- Auth
- Authorization
- Data model
- Migration/data loss
- Audit/security/privacy
- External system/provider
- Public contract/API
- Existing behavior
- Weak proof
- Multi-domain
- Deployment/runtime
- Standards change

## Items

| Rank | ID | Type | Lane | Title | Priority | Status | Links | Risk Flags | Next Artifact | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| active | AI-ADVISOR-01 | product | high-risk | Dedicated Finance Assistant: persistent chat, semantic tools and interactive cards | Owner Implement | in_progress | [Design](tickets/AI-ADVISOR-01-DETAIL_DESIGN.md), [V1 plan](../superpowers/plans/2026-09-22-finance-assistant-v1.md), [Page](../design/pages/finance-assistant/README.md), [User keys](tickets/TICKET-10-01-quan-ly-api-key.md) | Authorization, Data model, Public contract/API, External system/provider, Multi-domain | Add SSE/recovery, durable audit delivery, provider browser proof and full UAT | Local read-only vertical slice plus typed tool/card/lifecycle fixes and same-process cancellation are implemented/tested; cross-process cancellation, V2 writes and public release gates remain. |
| historical | UI-FORMS-03 | product | normal | Prior web form consolidation (not active native implementation authority) | Owner request | in_review | [Historical design](tickets/UI-FORMS-03-DETAIL_DESIGN.md), [Current native Assistant](../design/pages/assistant/README.md), [Native reconciliation](tickets/NATIVE-RECONCILIATION-20261015.md) | Existing behavior, Weak proof | Use current native plan/owner UAT | Retain old proof, but use the app's own native base guideline and current screen contracts; no archived web maintenance |
| active | AI-ENTRY-01 | product | high-risk | Hold Add → AI input → edit/approve/reject list | Owner request | in_review | [Design/proof](tickets/TICKET-09-01-ENTRY-DETAIL_DESIGN.md), [ADR-005](../decisions/ADR-005-ai-entry-review.md) | External system/provider, Authorization, Weak proof | Owner UAT with live app; expand evaluation beyond synthetic cases | Production path now uses strict JSON Schema; live synthetic benchmark 5/5, 24/24 fields. Human review of real proposals remains required; no auto-approval. |
| active | AI-ENTRY-02 | product | high-risk | One-shot AI batch, OCR-first PDF and private S3 transaction attachments | Owner request | in_progress | [Design](tickets/AI-ENTRY-02-BATCH-ATTACHMENTS-DETAIL_DESIGN.md), [Bug](bugs/BUG-001-s3-presigned-get-signature.md), [ADR-007](../decisions/ADR-007-aws-s3-presigning.md), [Stateless spec](../superpowers/specs/2026-09-21-stateless-ai-entry-design.md), [Plan](../superpowers/plans/2026-09-21-ai-batch-attachments.md) | External system/provider, Deployment/runtime, Weak proof | Finish browser attachment/download UAT | Live synthetic S3→OCR→LLM pipeline passes; browser download and owner attachment UAT remain. |
| active | CORE-03 | product | high-risk | Account timezone, Hũ, and automatic monthly overview/note | Owner-approved design | in_progress | [Design](tickets/CORE-03-TIME-JARS-MONTH-DETAIL_DESIGN.md), [Plan](../superpowers/plans/2026-09-22-timezone-jars-month.md), [Timezone](tickets/TICKET-01-01-thiet-lap-ca-nhan.md), [Jars](tickets/TICKET-04-hu-chi-tieu.md), [Month](tickets/TICKET-07-04-tong-ket-thang-ai.md), [ADR-008](../decisions/ADR-008-account-timezone-and-calendar-dates.md) | Data model, Migration/data loss, Public contract/API, Authorization, Multi-domain | Owner UAT: timezone edit, jar CRUD/assignment/history, month note and boundary behavior | Migrations 13–15, APIs/screens and automated checks are implemented; owner UAT and final docs review remain. |
| active | TICKET-02-02 | product | high-risk | Implement atomic internal transfer between wallets | Owner request | in_review | [Design](tickets/TICKET-02-02-TRANSFER-DETAIL_DESIGN.md), [Transaction page](../design/pages/transactions/README.md), [API reference](../architecture/API.md) | Data model, Public contract/API, Authorization, Existing behavior | Owner visual UAT | Migration 000016, endpoint, handler/PostgreSQL tests, integrated roundtrip, real Chromium Quick Add E2E, and API-backed native create/edit/delete pair proof pass. Current PATCH keeps source/destination fixed and updates both legs' amount/date/note; native edit selectors now communicate that constraint. Owner sign-off remains. |
| future | WORKER-01 | framework | high-risk | Dedicated cron worker sharing backend domain code | Owner request | draft | [Ticket](tickets/WORKER-01-cron-service.md), [Recurring](tickets/TICKET-05-02-giao-dich-den-ky.md), [Monthly](tickets/TICKET-07-04-tong-ket-thang-ai.md), [CORE-03](tickets/CORE-03-TIME-JARS-MONTH-DETAIL_DESIGN.md) | Data model, Multi-domain, Deployment/runtime, Weak proof | Decide live/rebuildable report vs persisted month-end artifact and recurring catch-up rules; then owner review detail design | Intake only; no scheduler, schema, dependency, or runtime changes approved. |
| future | AI-ENTRY-03 | design | high-risk | Select AI provider/model with transparent price and data sharing | Owner request (future) | draft | [Ticket](tickets/AI-ENTRY-03-PROVIDER-SELECTION.md), [Draft design](tickets/AI-ENTRY-03-PROVIDER-SELECTION-DETAIL_DESIGN.md), [AI-ENTRY-01](tickets/TICKET-09-01-ENTRY-DETAIL_DESIGN.md) | External system/provider, Privacy, Pricing, Public contract/API | Owner review of credential ownership, selection scope, providers and cost-consent policy | Design-only request; do not onboard providers, change credentials or incur new provider charges until separately approved. |
| review | DESIGN-09-AI | design | high-risk | Review feasibility and plan for two AI flows | Owner request | in_review | [Plan](tickets/TICKET-09-DETAIL_DESIGN.md), [AI tickets](tickets/TICKET-09-tro-ly-ai.md) | Data model, External system/provider, Authorization, Public contract/API | Owner design review | Plan only; T0–T7 and jar dependency are proposed, not executed |
| 0 | UI-BASE-01 | framework | high-risk | Normalize design specs, enforce base-first and refactor UI | High | done | [Design](tickets/UI-BASE-01-DETAIL_DESIGN.md), [Proof](tickets/UI-BASE-01-VERIFICATION.md), [ADR](../decisions/ADR-002-design-contract-enforcement.md) | Standards change, Existing behavior | Next UI slice follows base contract | Build, regression checks and shared browser acceptance passed; product contracts preserved |
| 1 | BL-001 | product | normal | Edit Financial Clarity design into implementation-ready source of truth | High | done | `design/system/DESIGN.md`, `design/INDEX.md` | Existing behavior | Docs review | Historical design consolidation retained as evidence. Current visual authority is the app's native base guideline/screen registry and owner-approved palette, not a third-party clone target. |
| 2 | BL-006 | maintenance | tiny | Align center create button within bottom navigation | Low | done | `app/src/atomic/organisms/BottomNavigation.tsx` | Existing behavior | Not required | Small task exemption: yes. Reason: one shared CSS positioning change; impact checked: API=no, DB=no, Security=no, Runtime=no, Standards=no. TypeScript and browser visual check passed. |
| 3 | BL-005 | product | high-risk | Implement wallet management | High | in_review | [Ticket](tickets/TICKET-01-02-quan-ly-vi.md), [Design](tickets/TICKET-01-02-DETAIL_DESIGN.md), [Proof](tickets/TICKET-01-02-VERIFICATION.md) | Data model, Migration/data loss, Public contract/API, Authorization | Human review or next credit/adjustment slice | Core CRUD, derived balances, delete impact and basic/goal/credit creation passed Go, PostgreSQL and Chrome UAT; statement and adjustment behavior remain separate. |
| 4 | BL-007 | product | high-risk | Implement basic income/expense ledger | High | in_review | [Ticket](tickets/TICKET-02-01-ghi-thu-chi.md), [Design](tickets/TICKET-02-01-DETAIL_DESIGN.md), [Proof](tickets/TICKET-02-01-VERIFICATION.md), [UI spec](../design/pages/transactions/README.md) | Public contract/API, Authorization, Multi-domain | Human review or receipt/jar follow-up | Real create/list/edit/delete UI and wallet/category rules are verified; receipts/OCR and jar assignment remain outside this slice. |
| 5 | BL-003 | integration | normal | Implement receipt OCR adapter | Medium | open | `docs/architecture/OCR_API.md`, `docs/architecture/INTEGRATIONS.md` | External system/provider, Public contract/API, Data model | Ticket plus detail design | Use OCR Platform for receipt image/PDF recognition; keep API key server-side. |
| 6 | BL-004 | framework | normal | Finalize first project-specific standards | Medium | open | `docs/standards/` | Standards change | TBD | Add human-maintained rules in `docs/standards/`. |
| done | NATIVE-FEEDBACK-BUBBLE-REDACTION | bug | tiny | Exclude the floating feedback trigger from captured feedback screenshots | High | done | [Bug](bugs/NATIVE-FEEDBACK-BUBBLE-REDACTION.md), [Phase](phases/PHASE-002-native-ios-swiftui.md), [Validation](VALIDATION_MATRIX.md), [Release](../releases/CHANGELOG.md) | Existing behavior, Weak proof | None; retain regression test | Simulator regression passes; no API or schema change. |
| done | NATIVE-FEEDBACK-BUBBLE-COLLISION | bug | normal | Prevent floating feedback button from covering Account content | High | done | [Port audit](PORT-100-AUDIT.md), [Validation](VALIDATION_MATRIX.md), [Visual capture plan](../superpowers/plans/2026-10-07-native-authenticated-visual-capture.md) | Existing behavior | None; retain focused regression | API-backed iPhone 16 test reproduced and then verified the fix; Account security explanation wraps before the bubble while global feedback remains available. |
| done | NATIVE-WALLET-ROW-PARITY | bug | normal | Match native wallet-selection rows to Stitch/web anatomy | High | done | [Port audit](PORT-100-AUDIT.md), [Wallet contract](../design/pages/wallets/README.md), [Validation](VALIDATION_MATRIX.md) | Existing behavior, Weak proof | Continue to next pixel-parity gap | Shared row base stacks title and balance, marks selected scope, and keeps native Chi tiết compact on the same row; API-backed iPhone 16 regression passes. Goal/credit extra values remain intact. |
| done | NATIVE-WALLET-AGGREGATE-GLYPH | bug | tiny | Use the approved globe/teal glyph for wallet aggregate | High | done | [Port audit](PORT-100-AUDIT.md), [Wallet reference](../design/references/wallet-selector/README.md), [Validation](VALIDATION_MATRIX.md) | Existing behavior | Continue to next pixel-parity gap | API-backed iPhone 16 regression distinguishes aggregate from an individual basic wallet; only presentation changed. |
| done | NATIVE-CATEGORY-EDITOR-PARITY | product | normal | Match native category create/edit route and applicable-wallet layout to web/Stitch | Owner request | done | [Audit](PORT-100-AUDIT.md), [Page spec](../design/pages/account-groups/README.md), [Ticket](tickets/TICKET-01-03-tong-vi-danh-muc.md), [Detail design](tickets/NATIVE-CATEGORY-EDITOR-ROUTE-DETAIL_DESIGN.md), [Validation](VALIDATION_MATRIX.md) | Existing behavior, Shared component, Backend contract | Continue category list/filter audit; visual device sign-off remains part of overall port | Dedicated editor and visible shared checkbox rows are implemented. iPhone 16 API test covers create/scope/readback/filter/rename/delete. It caught and fixed `wallet_i_ds` vs `wallet_ids`; backend create now persists category and owner-validated wallet links atomically. No public API/schema/migration change. |
| done | NATIVE-LIST-FEEDBACK-BUBBLE-COLLISION | bug | normal | Keep the global feedback bubble from covering native list actions | PORT-100 screenshot audit | done | [Bug](bugs/NATIVE-LIST-FEEDBACK-BUBBLE-COLLISION.md), [Dock detail design](tickets/NATIVE-FEEDBACK-BUBBLE-DOCK-DETAIL_DESIGN.md), [Implementation plan](../superpowers/plans/2026-10-12-native-feedback-dock.md), [Port audit](PORT-100-AUDIT.md), [Validation](VALIDATION_MATRIX.md) | Existing behavior, Weak proof, Shared component | Finish compact-device/full verification; keep authenticated owner visual acceptance in the broader port UAT | Fixed with shared 60pt shell dock and bounded/clipped route viewport. Focused default-device geometry/reachability and modal clearance pass; compact/full gates are recorded separately. No backend/API change. Overall native-port goal remains active. |
| done | NATIVE-FEEDBACK-RIGHT-DOCK | bug | tiny | Dock feedback bubble on the simulator's right edge without shifting content | Owner feedback | done | [Bug](bugs/NATIVE-FEEDBACK-RIGHT-DOCK-20261015.md), [Base components](../design/system/BASE_COMPONENTS.md), [Validation](VALIDATION_MATRIX.md) | Existing behavior, Shared component, Weak proof | Physical-device visual acceptance remains part of the broader native port UAT | Shared layout now computes a stable trailing-edge overlay; focused geometry/UI tests pass and no API/schema/runtime change was made. |
| in_progress | NATIVE-CATEGORY-TREE-PARITY | product | normal | Match category management and selection trees to Stitch/web | Owner request | in_progress | [Audit](PORT-100-AUDIT.md), [Category page spec](../design/pages/account-groups/README.md), [Base tree spec](../design/molecules/category-tree/README.md), [Validation](VALIDATION_MATRIX.md), [Detail design](tickets/NATIVE-CATEGORY-TREE-PARITY-DETAIL_DESIGN.md), [Editor route design](tickets/NATIVE-CATEGORY-EDITOR-ROUTE-DETAIL_DESIGN.md), [Picker design](tickets/NATIVE-CATEGORY-PICKER-PARITY-DETAIL_DESIGN.md), [Archive spec](../superpowers/specs/2026-10-08-category-archive-design.md) | Existing behavior, Shared component, Weak proof, Public contract/API | Continue category screen visual parity; root-card, row copy and budget-clear parity now have API-backed Simulator proof | Native management, transaction, budget, recurring and credit-entry selectors use shared hierarchy/search bases. Archive/restore and inactive assignment validation are implemented across backend/web/native; migration and HTTP→PostgreSQL tests verify rejected new assignments, preserved transaction/budget/recurring references, existing recurring execution, and restore. Picker labels/ghost action and persistence regressions pass. Feedback overlap is resolved by the shared dock (see validation); populated-data/device UAT remains open. |
| done | NATIVE-ASSISTANT-CARD-DENSITY | bug | tiny | Bound native assistant result-card rows to web limits | High | done | [Audit](PORT-100-AUDIT.md), [Page spec](../design/pages/finance-assistant/README.md), [Plan](../superpowers/plans/2026-10-07-native-assistant-card-parity.md), [Validation](VALIDATION_MATRIX.md) | Existing behavior, Weak proof | Keep total metadata independent from visible prefix | Transactions show 5; wallet/goal/budget/jar show 8, matching web. No API/schema change. |
| in_review | NATIVE-API-KEY-COPY | bug | tiny | Add a safe, accessible copy action for the one-time API key secret | Owner feedback | in_review | [Ticket](tickets/TICKET-10-01-quan-ly-api-key.md), [Detail design](tickets/API-KEY-COPY-DETAIL_DESIGN.md), [Account spec](../design/pages/account-settings/README.md), [Validation](VALIDATION_MATRIX.md) | Existing behavior, Weak proof | Owner review on iPhone; resolve Xcode signing account/profile before install | Shared MP bases reused; 150 unit tests and API-backed copy-confirmation UI test pass. No API/schema change. Release notes await owner approval. |
| in_progress | NATIVE-MOBILE-FEEDBACK-20261013 | integration | high-risk | Audit and fix remaining production feedback on iOS native only | Owner “fix hết feedback khác” | in_progress | [Ticket/design](tickets/NATIVE-MOBILE-FEEDBACK-20261013.md), [Validation](VALIDATION_MATRIX.md), [Port audit](PORT-100-AUDIT.md) | Existing behavior, Weak proof, Security/auth, Shared component | Fixed wallet-name repetition and propagated privacy masking; continue audit against each report; do not change logout/ledger semantics without evidence and regression tests | Native-only scope; latest authenticated production audit contains 48 open records, while owner/device UAT and release acceptance remain open. Release notes await owner approval. |
| done | NATIVE-AI-ENTRY-UNCERTAIN-RETRY | bug | normal | Recover ambiguous native AI-entry submissions without duplicate POSTs | PORT-100 API parity audit | done | [Bug/design](bugs/NATIVE-AI-ENTRY-UNCERTAIN-RETRY.md), [Assistant page](../design/pages/assistant/README.md), [API guidance](../architecture/API.md), [Validation](VALIDATION_MATRIX.md) | Existing behavior, Weak proof | Continue with the next unblocked port gap | iOS retains request ID and draft after uncertain transport/5xx outcomes, locks resubmission and uses GET recovery; GET 404 permits explicit retry. 146 unit + isolated recovery UI test, Release and base/project guards pass; no API/schema change. |
