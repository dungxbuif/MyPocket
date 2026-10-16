---
artifact_type: project_context
id: CONTEXT
status: active
owner: shared
human_fields:
  - current_focus
  - open_questions
  - priority_override
ai_fields:
  - recently_touched_areas
  - recent_decisions
  - next_steps
  - queue_summary
shared_fields:
  - current_status
  - active_backlog
  - current_queue_focus
  - active_phase
  - active_ticket
  - active_bug
updated: 2026-10-16
---

## Current checkpoint — 2026-10-16: Savings/Budget and physical automation

Owner delegated financial design details and requested implementation. Local
candidate now has both Savings Report sections (goal-wallet activity and
account-wide surplus), paged source transactions, versioned recurring budgets
for week/month/quarter/year, optional positive carryover and deterministic
budget forecast/daily detail. Migration 035 is additive and verified on a
disposable PostgreSQL 17 database. Generated OpenAPI contains 85 paths.
Full backend race suite runs against that real database and passes; native
signed unit target passes 317 tests and Release warnings-as-errors build passes.
API-backed native UI verification passes 2/2 with unique synthetic data and
real state readback, including the Stop confirmation. Review follow-up native
unit rerun also passes 317/317. A Personal Team signed Release device artifact
builds and verifies its signature with a documented Universal Links entitlement
override; the owner's installed app remains unchanged.

Independent review fixed wallet-baseline chronology and deletion of superseded
scope revisions. Scope changes reset carryover. AI single-budget read now loads
the prior owned chain before selecting output so carryover stays consistent.
See [design](superpowers/specs/2026-10-16-savings-budget-design.md),
[plan](superpowers/plans/2026-10-16-savings-budget-implementation.md),
[ADR-023](decisions/ADR-023-ledger-savings-recurring-budgets.md).
Production still runs the previous 78-path image and migration 34; the new
runtime/schema and native candidate have not been deployed or installed.

Physical automation is now working on the owner's paired iPhone. Device Kit
0.0.34 is separately signed and installed; MCP UI reads pass. Existing
MyPocket 1.1.0 (2) remains installed. Helper profile expires October 23;
[installation evidence](operations/IOS-AUTOMATION-INSTALL-20261016.md).
The owner explicitly reintroduced Money Lover as the native visual reference;
an agent is visiting individual functions/drafts/pickers on that device,
recording actual and unvisited states in the
[reference ledger](design/reference-survey/2026-10-16-money-lover-iphone.md).
Extended MyPocket functions are preserved. Survey completion and 100% UI
reproduction are separate acceptance gates, neither is claimed yet.

Reference waves 1–4 are now implemented locally for populated overview,
wallets, savings/budget reports, transaction ledger, credit and notifications.
Shared surfaces
own the reference radius, compact metric grid, adaptive daily rows, masking
and accessibility reading order; screen code only supplies data. Focused
signed simulator tests pass 14/14 for waves 1–3 plus 2/2 for wave 4, the
native unit target passes 331/331 before wave 4, and
the Release simulator build passes with warnings treated as errors. A
deterministic sorted-key encoder was added so identical budget-series retries
produce identical request bytes. The unscoped full xcodebuild run was stopped
after a simulator-launch hang, so no full UI-suite pass is claimed. See the
reference-wave detail design and the validation matrix for evidence and
remaining populated-screen capture work.

## Previous checkpoint — 2026-10-16 20:45 +0700

The async advisor and AI-entry multi-image context fixes are now deployed to
production. API and worker run
`registry.dungxbuif.com/mypocket-api:release-20261016-7846945` at digest
`sha256:71fdd4b4155972f3e6380ae02de846264e503347ab63b9693f65d2f0c713db43`;
both services are `1/1`, health returns `200`, and production Swagger reports
78 paths including `POST /api/v1/ai/advisor/messages/async`. The authenticated
sync capability route is present. Before enabling the idle expiry worker, a
validated `pg_dump -Fc` was written to
`/tmp/mypocket-pre-migration-20261016.dump` on `vm100` (SHA-256
`60bd27c795f510854d9ca8c6e16a6285ab1a9ab2f46b3edde1b5a71e160d3a3d`), then
migrations 033 and 034 were applied from the API image. Migration version is
34 and clean; no financial rows were reset or deleted. The worker is running
and no new `last_user_message_at` schema error appeared after the migration.

The multi-image fix preserves request order and wraps each OCR result in a
document boundary before model extraction. It does not claim provider OCR
accuracy; owner-authenticated AI-entry UAT with the configured provider remains
open. The web service remains `0/0` by the native-only rollout decision, and
the final signed-device/Face ID/visual gates remain open.

Gap-closure verification then completed locally. Native Reports now carries
the owner guard and fails closed when the month summary is unavailable
(`089fd656`); signed iOS Simulator XCTest passed 307/307. Offline transfer
receipt replay coverage was added (`54087c8`) and deterministic AI/OCR
fixtures plus the savings/budget scope audit were added (`982f8c48`). Backend
`go test ./...`, `go test -race ./...`, and `go vet ./...` pass. The fixtures
cover duplicate OCR rows, missing dates, wallet inference, owner-safe tool
arguments, and no-auto-approval. Savings reports and budget recurrence /
forecast remain explicitly partial because no stable API contract exists yet;
no new migration or production behavior was introduced by these local commits.

## Previous checkpoint — 2026-10-16 19:46 +0700

The local async Finance Assistant cancellation slice is now verified. Backend
`go test ./...` and `go test -race ./...` both pass 296/296 across 16 packages;
`go vet ./...` is clean. The local OpenAPI surface is 78 paths and includes
`POST /api/v1/ai/advisor/messages/async`. The synchronous advisor endpoint is
unchanged; the async endpoint reserves an owner-scoped run before provider work
and returns `202`, allowing the first request to be cancelled through the
existing run-cancel route. The native composer now uses this endpoint and polls
the existing run resource. The iOS simulator unit target passes 304/304 with
zero skips, failures, or runtime warnings at
`/tmp/mypocket-async-unit-20261016.xcresult`.

The async slice is committed locally as `fc173237`. A follow-up AI-entry
multi-image context fix is also committed locally as `6b616f6`: model-facing OCR
blocks now preserve request order and carry document markers, while persistence
maps text by attachment ID. Backend tests pass 299/299 after that fix. Production
still serves the previously deployed 77-path image; no production rollout,
database change, provider call, or device install was performed for either local
slice. Owner-authenticated production UAT and physical-device acceptance remain
open.

## Previous checkpoint — 2026-10-15 22:54 +0700

Latest authoritative state: release tag `v1.1.0` is deployed and
`mypocket-api` plus `mypocket-worker` both run
`registry.dungxbuif.com/mypocket-api:release-1.1.0-20261015-2212` at digest
`sha256:bef8040e7a372774234fd96d69e690215dd5d823f4520f901fd5cf1638fb37d2`.
The public health endpoint is 200. Production Swagger JSON and UI return 200,
`/docs/` redirects to the versioned UI, and the public path set is 77/77 equal
to local generated Swagger. Sync capabilities, advisor reset/confirm and
feedback-resolution routes are mounted behind auth (unauthenticated probes
return 401, not 404). The capability body is still unverified without a valid
owner credential, so native offline sending remains fail-closed.

The jar feedback hardening is deployed: the v1.3 advisor handbook explicitly
states that the model is allowed to prepare `propose_create_jar` proposals and
must render a confirmation card; direct writes still require owner confirmation.
Feedback remains open pending owner-authenticated UAT. Backend full tests and
vet pass; the focused native queue/cache/design suite passed 123/123 at
`/tmp/mypocket-goal-contracts-20261015/Logs/Test/Test-MyPocketIOS-2026.10.15_22-01-14-+0700.xcresult`.
The current native unit target then passed 304/304 with zero skips, failures,
or runtime warnings at `/tmp/mypocket-unit-current-20261015.xcresult`.
The savings 422 E2E artifact is present and reports 1/1 passed at
`/tmp/mypocket-offline-savings-20261015e/Logs/Test/Test-MyPocketIOS-2026.10.15_20-37-31-+0700.xcresult`.

The connected iPhone now has a Personal Team-signed Release build installed
and launched (`com.mypocket.ios`, version `1.1.0 (2)`), with the production API
base embedded. Password login and physical Face ID enrollment/logout still need
owner interaction on that device.

Remaining release gates are authenticated production capability and AI UAT,
synchronous first-request cancellation UX, visual/device acceptance, APNs, and
configuration-dependent Google/Travel/Insider/Credit provider paths. These are
not inferred as complete from local tests or unauthenticated probes.

The deployed-state reconciliation also updated the AI advisor, Feedback API and
offline-sync public references: production route presence is documented, while
owner-authenticated payload/UAT remains explicitly open. Backend tests were
re-run at 295/295 with `go vet` clean; the installed Release app was relaunched
on the connected iPhone after verification.

## Historical checkpoint — 2026-10-15 (superseded)

Latest local gate re-run against the current worktree: backend `go test ./...`
295/295 and `go vet ./...` passed; web design contracts, TypeScript and Vite
production build passed with only the existing chunk-size advisory.
The backend race suite also passed 295/295 across 16 packages with no race
reports.
The focused HTTP contract suite passed 22/22 under the race detector,
covering capability advertisement, full-owner API keys, feedback resolution,
advisor lifecycle/cancellation and device/offline auth boundaries.
An additional 15 focused race tests for Google exchange classification, Travel
Mode, Money Insider and Credit validation passed; missing provider/configuration
remains an intentional runtime gate rather than an unhandled write or panic.
The full native scheme then passed on the iPhone 16 simulator with 330 passed,
64 documented skips, 0 failures and 0 runtime warnings; artifact:
`/tmp/mypocket-native-current-20261015/Logs/Test/Test-MyPocketIOS-2026.10.15_20-51-59-+0700.xcresult`.
The focused offline capability suite passed 3/3, proving the client rejects
404/disabled/unknown capability policies and preserves pending writes without
leasing or sending them until a compatible backend advertises support:
`/tmp/mypocket-offline-capability-20261015/Logs/Test/Test-MyPocketIOS-2026.10.15_21-04-22-+0700.xcresult`.
An unsigned physical-device build also completed for the online iPhone with
`CODE_SIGNING_ALLOWED=NO`, producing an arm64 app under
`/tmp/mypocket-ios-device-nosign-20261015/Build/Products/Debug-iphoneos/`.
This separates source/build correctness from the missing provisioning profile;
the app was not installed.
The stricter Release device build also passed with warnings treated as errors
and signing disabled; it produced the Release app under
`/tmp/mypocket-ios-device-release-nosign-20261015/Build/Products/Release-iphoneos/`.
The API and worker were also cross-compiled locally for `linux/amd64` with
static ELF output; hashes and non-deployed status are recorded in
`docs/operations/DEPLOYMENT.md`. No registry push or production mutation was
performed.
Read-only SSH to vm100 confirms the running tasks are still API
`auth-12572c87` and worker `release-73c4c37`; vm100 is a Swarm worker and cannot
inspect or update services. Manager-level rollout remains intentionally open.

Continuation verification: backend `go test ./...` passed 295 tests across 16
packages, `go test -race ./...` passed 295/295 and `go vet ./...` passed; the
focused feedback refresh/resolution race regression passed at
`/tmp/mypocket-feedback-race-green-20261015/Logs/Test/Test-MyPocketIOS-2026.10.15_21-14-02-+0700.xcresult`.
Refresh families now also carry an absolute expiry locally; rotation/logout
TTL tests pass and the static Linux artifacts are staged but not deployed.
Read-only production probes remain unchanged: sync capabilities 404, versioned
Swagger 200, legacy `/docs/` 404 and `/api/v1/openapi.json` 404. No production,
database or device state changed.
The public Swagger `doc.json` exposes 72 paths versus 77 locally; the five
local-only paths are advisor action confirmation, advisor conversation reset,
category status, feedback resolution and sync capabilities. This is the exact
backend/docs rollout delta, not an unverified assumption.
The configured-provider synthetic overlap smoke passed 1/1 at
`/tmp/mypocket-live-provider-overlap-20261015.log` with no DB/ledger writes.
The follow-up live provider contract probes passed 3/3 at
`/tmp/mypocket-live-provider-contract-20261015.log`: extraction, forced
tool-calling schema and five JSON-schema transaction cases all passed against
`DeepSeek-V4-Flash` using synthetic inputs only. No financial rows were written.
The opt-in product-help grounding probe also passed at
`/tmp/mypocket-live-rag-grounding-20261015.log`: the provider selected
`get_product_help` and grounded a Vietnamese answer on a synthetic handbook
result. This is provider/tool-result grounding, not live document retrieval or
production UAT.
OCR/PDF remains unrun because `OCR_API_KEY` is empty; external live document
retrieval/RAG still has no configured evaluation fixture.
The env audit found only `backend/.env.example` and `backend/.env.local`; both
declare the OCR variable, and both leave it empty. No additional project env
file provides a credential.
The installed Apple Development identity/team is present, but a device build
still stops before compilation because no provisioning profile matches
`com.mypocket.ios`; connected iPhones are offline. No provisioning update or
Apple account/device mutation was attempted.
The latest device probe at 20:50 +0700 sees `DungxBuif` online, but the local
provisioning-profile directory is still empty, so signed installation remains
blocked by signing configuration rather than device connectivity.
The 21:09 +0700 follow-up is unchanged: the Apple Development identity is
valid, `DungxBuif` is online, the second phone is offline, and no provisioning
profile exists. The unsigned arm64 device build remains the latest build proof.

Authoritative scope/status: [native reconciliation](work/tickets/NATIVE-RECONCILIATION-20261015.md).
Follow-up audit reran the non-environment gates after the latest native fixes:
backend `go test ./...` and `go vet ./...`, web type/design/transaction checks,
native project/base-component guards, and `git diff --check` all pass. No local
service was restarted and no production/device state changed. The approved
savings 422 UI rerun is now green against the disposable local
PostgreSQL/API/fault-proxy stack; the authoritative 1/1 artifact is recorded
below. The earlier `127.0.0.1:55434` fixture failure and the run that bypassed
the fault proxy are historical harness failures, not current product evidence.
Auth error classification is now corrected locally: user-store lookup/create
failures from Google exchange return `503 GOOGLE_USER_STORE_UNAVAILABLE` with a
generic detail, while invalid OAuth data remains `401`; use-case/HTTP tests and
regenerated Swagger pass. The already-running API process was not restarted,
so its old response must not be used to judge the new contract.
Finance Assistant prepared jar actions are now implemented locally: the
read-only `propose_create_jar` tool validates month/name/allocation, persisted
cards map to `jar_confirmation`, and the owner/generation-bound confirmation
dispatcher reuses the existing mutation ledger with `jar.create`. Backend
full tests/vet and the iOS simulator build pass after this slice. The fresh
unit-test invocation reached CoreSimulator diagnostics and was stopped after
the runner stalled; this is an environment gate, not a test pass. No service,
production, migration or device state was changed; rollout/UAT remain open.
The same local slice now includes `propose_update_jar`: it validates a UUID
jar ID and monthly replacement fields, maps to the existing confirmation card,
and dispatches to the owner-scoped jar PUT handler without creating a second
jar. Focused Go contracts and DesignSystemTests (90/90) pass; production,
device and provider-backed UAT remain open.
The advisor action slice now also includes `propose_create_budget` and
`propose_update_budget`: strict
budget/date/scope validation, a native `budget_confirmation` card and a
server-bound `budget.create` mutation ledger protect confirmation retries.
Focused backend tests and DesignSystemTests 90/90 pass; no production or
device state changed.
After the card copy assertion, the signed simulator full native suite passed
294/294 with no runtime warnings at
`/tmp/mypocket-full-native-signed-20261015.xcresult`. The unsigned simulator
attempt's four Keychain `-34018` failures were reproduced as a signing-mode
environment issue and passed 4/4 under signed execution; no source workaround
or entitlement weakening was introduced.
The advisor system handbook had stale wording that told the model it could not
create jars at all. Local source now explicitly tells the model it is allowed
to prepare a `propose_create_jar` proposal, instead of saying “không có quyền”;
the provider request also asserts that this non-mutating tool is exposed.
Focused provider-handbook, tool-definition and full backend tests are green.
The prompt-hardening API image was rolled out without a migration or database
change as `ai-jar-capability-20261015-215712` (`1/1`, public health `200`);
owner UAT and feedback resolution remain open.
The Feedback shell anchor was also corrected locally: obstacle-driven movement
is gone, the bubble stays in the lower-right dock above footer clearance, and
the scroll viewport is not reserved for it. The focused regression and full
DesignSystem target (90/90) plus warnings-as-errors build/guards pass; populated
screen and device acceptance remain open.
Assistant numeric/card safety is also reconciled locally: seven focused signed
simulator regressions pass for exact money/progress/count conversion, nontrapping
fallback formatting and privacy masking. Evidence is
`/tmp/mypocket-numeric-safety-20261015/Logs/Test/Test-MyPocketIOS-2026.10.15_20-23-53-+0700.xcresult`;
physical-device/UAT and release acceptance remain open.
Read-only production probing also confirms `/api/v1/docs/index.html` serves
Swagger UI (GET 200), but `/docs/` and `/api/v1/sync/capabilities` still return
404. The local redirect/capability implementation has not been deployed; no
authenticated data or production mutation was performed.
Category helper/API-first/last-drag correction280/280PASS; numeric/chat287/287PASS.
Full12business+279unit regression11171 passed291/291,0skip,0 runtime warnings
after the local helper/fixture corrections. Evidence:
`/tmp/mypocket-reconcile-full-green-20261015.xcresult`. Category keyboard
boundary, nested credit container selection and recurring fixture wallet proof
are now green. This is a local simulator/API gate, not production or device
acceptance; broader product/lifecycle/release gates remain open.
Fresh realPG/Redis full Go race32091PASS and vetPASS. Production APIauth-12572c87
healthy, workerrelease-73c4c37; sync capability404, no new-client install before
matching backend rollout. vm100 read SSH works; manager pi requires1Password.
48openfeedback freshly fetched, newest6assessed. Note editor/device-auth status
corrected; [new feedback design](work/tickets/NATIVE-FEEDBACK-NEW-20261015-DETAIL_DESIGN.md)
remains review-gated. The advisor idle-retention foundation is now local:
`000033`, generation-safe access purge, bounded worker pass and `410` mapping;
native expiry metadata, real database proof and production rollout remain open.
The explicit Chat mới reset is implemented locally through migration `000034`,
an owner-scoped content-free idempotency ledger, and the native overflow action.
The reset service also cancels queued/running provider contexts in this process;
replayed reset keys do not cancel newer runs. History reads now expose optional
server-authoritative lifecycle metadata (`conversation_id`, `generation`,
`last_user_message_at`, `expires_at`, `server_time`) without changing the
message array or extending TTL; backend/native contract tests and a warnings-as-
errors simulator build pass locally. No new deployment/history deletion/
financial reset/feedback resolution/release.
Full original goal remains in_progress, not complete. Web production build and
design/type checks are green; browser tests requiring a live Vite/fixture remain
environment-gated and are not claimed as passed.

Portfolio bounded slice: empty assets now expose a shared-confirmed
`portfolio-delete-{assetID}` action backed by the existing owner-scoped delete
route; assets with trades remain append-only. Debug simulator build passed at
`/tmp/mypocket-portfolio-delete-build-20261015`. The first API-backed delete
rerun is pending because the disposable PostgreSQL fixture port
`127.0.0.1:55434` was unavailable, so the fixture login returned
`GOOGLE_MOBILE_LOGIN_INVALID` before the UI was reached. No environment was
changed to work around it.

Reports invalid-month bug fixed locally: `ReportsMonthPolicy` rejects malformed
or out-of-range `YYYY-MM` before network calls, and invalid/failed loads clear
the previous summary/detail/trend instead of showing stale content. Unit and
authenticated UI regressions pass. Overview/Reports now also reuse the encrypted
owner-scoped snapshot for transient offline reads; exact-month report fallback,
cache merge and Overview recovery are covered by 9 focused tests and a Debug
build. PostgreSQL-backed UI, device UAT and production rollout remain separate.

Advisor cancellation reconciliation: the deterministic PostgreSQL stale-status
insert race is fixed locally using the approved conversation→run lock/recheck;
the focused real-DB suite is 4/4 and the full Go race suite is green. This does
not claim that the initial native Stop control can interrupt the synchronous
submit request, nor that the patch is deployed. Those remain separate code and
rollout gates.

Savings queue follow-up: the shared queue menu now evaluates edit eligibility
per item. Goal-wallet transfer drafts can open `Sửa bản nháp` with wallet data
alone; ordinary transaction drafts still wait for categories. Policy unit proof
is `/tmp/mypocket-goal-queue-policy-green-20261015.xcresult`. The approved
fault-injected savings 422 UI rerun is green against disposable local PostgreSQL
version 34 and an API→fault-proxy stack: pending 100,000 đ / remaining 900,000
đ stay unchanged, restart restores the draft, editing to 22,500 đ commits one
transfer pair, reloads 122,500 đ / 877,500 đ, and leaves reports unchanged.
Evidence: `/tmp/mypocket-offline-savings-20261015e/Logs/Test/Test-MyPocketIOS-2026.10.15_20-37-31-+0700.xcresult`.
The earlier failure was discarded as a harness misconfiguration where the app
bypassed the fault proxy and received 201. No production, device or release
state changed.

### Historical checkpoint log — superseded by the matrix above

20261015 owner asks reconcile+finish all. [Current matrix](work/tickets/NATIVE-RECONCILIATION-20261015.md)
supersedes stale snapshots: note editor implemented; device auth backend already
deployed; offline engine/outbox/foreground locally implemented but production
capability freshly404/public OpenAPI absent. Current native ordinary creates use
queue, so NO further new-client install before compatible backend rollout. Prior
installed candidate must not be treated as proof of production offline readiness.
vm100 private-key read-only SSH works, APIauth-12572c87healthy/workerrelease-73c4c37;
Swarm managerfalse. Manager pi rejects known local homelab identity,1Password
agent needed; do not alter trust/permissions/promote node. Category reviewed
API-first lazy-row pass25223 running; first13913 compile-only await correction,
no behavioral result claimed. Fullscope remains open; no prod mutation/release.

20261015 final local numeric/chat48656exit0:287/287PASS0skip/no runtime warnings,
279units+8chatUI. Numeric range/goal percentage and fallback privacy sensitivity
RED→GREEN; captures inspected, base guards/tests/project/diffPASS. Human/agent
guide/catalog/screen reconciled. Business pinned-save probe/wallet CRUD/transaction
PASS, category remainsRED after3renewed helper attempts, stopped for review.
Actual checkbox fixture previously selected3wallets; now singleton selection is
explicit, but latest run stops at post-save lazy-name wait before API readback.
No fourth attempt before review. Idle lifecycle spec9ab9248d review pending,
30minapproved but deletion not implemented. Releasebuild18960 waiting codesign,
iPhone now unavailable; previous installed candidate unchanged. No release/DB
reset/backend deploy/feedback resolution. Fullgoal still unachieved.

20261015 human “Chốt” resumes work:30-minute idle interval plus bounded numeric/
privacy fix and pinned/AX helper refinement approved. Numeric native boundary
RED→GREEN; first expanded unit run277PASS/1FAIL caused only expected literal
regular-space versus existing formatter NBSP, corrected test expectation without
changing shared formatter. Current full units+4actualAPI/UI business gate running
`/tmp/mypocket-approved-business-numeric-20261015.xcresult`, no final claim yet.
Idle lifecycle written spec committed9ab9248d, awaiting review before implementation.
No production/chat deletion, new device install or release publication this turn.

20261015 resumed goal blocked audit: quick iPhone candidate delivered, not full
completion. Idle TTL and other specific design gates unanswered across current
user turn+two automatic continuations; independent native RED/replay investigation
already complete. No live build/test/codesign/proxy; SSH identities0. Remaining
product changes cross explicit review/loop guards. Goal blocked pending human/
access state change, not paused/complete. No repeated known-green tests, release
publication, history deletion or production financial mutation. Preserve fullgoal.

20261015 next resumed goal: prior design-only TTL question was no-progress; now
new actual native numeric regression RED1/1 at `/tmp/mypocket-native-progress-range-red-20261015.xcresult`
proves unsafe progress amount acceptance, no product fix. Idle-reset investigation
found replay depends on old user messages and generation is not yet enforced
against current conversation on answer insert: do not blindly delete history.
[Idle lifecycle work](work/tickets/NATIVE-ADVISOR-IDLE-RESET-20261015.md) records
scope, timing and required tests;30min/24h still unanswered. iPhone installed
candidate unchanged, no production/history/finance writes. Whole goal remains open.

20261015 candidate DELIVERED: Release com.mypocket.ios installed in place on
connected iPhone15Plus00008120-001C28C02131A01E; launched PID22311 and confirmed
running from the new installation path. Production API Info.plist verified,
health200. Initial codesign sealed temporary-file reference corrected by re-sign;
strict/deep verificationPASS before final install. Evidence `/tmp/mypocket-stable-device-install-verified-20261015.json`
and `/tmp/mypocket-stable-device-launch-20261015.json`. BinarySHA256
e09f3884c4fa54d3639c218462890d2c5642f9ee71000a524338e0d4c6e7c379.
No uninstall/reset/DB mutation; owner UAT pending, not full stable-release signoff.
Backend cancellation fix remains local/not deployed; known numeric/privacy and
other backlog gates retained. No release note publication before owner acceptance.

20261015 owner requests quick production-connected iPhone candidate, not full
backlog completion. Approved cancellation conversation→run lock/recheck applied
locally: real PostgreSQL gate4/4PASS (previously3/1RED), full `go test -race ./...`
PASS, fresh native273/273PASS0skip/no runtime warnings. Base/project/diff guards
PASS. Release build uses production API; signing/install verification in progress.
Backend patch is NOT deployed; financial data unchanged. Numeric/privacy,
business test refinements, initial Stop and broader gates remain open. No release
publication or feedback acceptance before owner UAT.

20261015 goal blocked audit: full objective retained, not complete/paused. No live
test/proxy/go-test; SSH agent has no identities. Same specific design review
condition persists over >=3consecutive goal turns despite independent proof work.
Await numeric/privacy, cancellation transaction and pinned/AX test refinement
answers; Feedback/cache/goal422/transport decisions remain separate. Do not repeat
known green gates as replacement for product fixes. Latest actual proofs:chat8/8,
history274/274, business273/2RED, cancellation3/1RED. No production/install/release.
[Full audit](work/tickets/NATIVE-FULL-COMPLETION-20261014.md).

Latest20261015 cancellation realDB70798exit1:3PASS/1raceFAIL0skip,
`/tmp/mypocket-advisor-cancel-db-red-20261015.log`. Actual controlled conversation
lock proves late assistant insert after cancellation commit. Sequential/owner/
new-run and blocking-provider stubservice PASS, not universal cancellation proof.
Exact synthetic-owner teardown0remaining. Bounded lock/recheck design sent, pending;
no product/backend/runtime patch. Source also shows syncPOST returns runID too late
for initial native Stop; separate transport/UI verification needed. Human+agent
guide/catalog corrected stale no-late-answer/Stop claims. Fullgoal active, no
prod/device/release; numeric/business/Feedback/cache reviews remain pending.

20261015 persisted native history E2E27531exit0:274/274PASS0skip/no runtime warnings
(273unit+1actual DB/API/UI), `/tmp/mypocket-advisor-history-final-20261015.xcresult`.
51synthetic messages prove latest50/older1/native nojump/scrollback/restart/API
unchanged. First40442 failed offscreenAX assumption, corrected test only; no
product patch. Capture76D78B35 inspected. Exact fixture cleanup0convo/0run/0msg;
normal8081 defaultlaunch47423, guards/project/diff pass. Human+agent history guide/
screen/ticket/ledger reconciled, no API/auth/schema/release/prod/device action.
Fullgoal active; live cancellation/races/numeric/business/Feedback/cache gates open.

20261015 native advisor wire coverage added (tests only). Target24082 PASS2/2;
final72133exit0 native273/273PASS0skip/no runtime warnings,
`/tmp/mypocket-advisor-wire-units-20261015.xcresult`. History request cursor/auth/
chronological decoded parts and cancel exactPOST/noresubmission checked through
real client+adapter with transport-only fake. Characterization, not productfix
or native history/cancel/provider E2E. Guards/tests/project/diff pass. No tests
running, product/API/prod/device unchanged. Publicdocs no change (wire unchanged).
Numeric/business/Feedback/cache review gates and full objective remain open.

20261015 current-helper chat88653exit0:8/8PASS0skip/no runtime warnings,
`/tmp/mypocket-chat-current-helper-20261015.xcresult`. Current keyboard/AX/
landscape/draft/APIoverview regressions pass. CapturesA80B6F81/FD268710/F165D9F3
inspected: landscapeconvo63pt, screenshotcrop persists, AXFeedback oversized/text
overlap unaccepted. Not populatedhistory/cancel/provider proof. History source
limit50 coherent; cancellation task-overlap risk source-only, needs actual race
reproduction. No product/test source changed. Fullgoal active; numeric/business/
Feedback/cache review gates remain, no release/prod/device action.

Latest numeric diagnostics20261015: actual Envelope.swift JSON decode→raw fallback
12500PASS/1e100SIGTRAP; extracted actual count helper42PASS/Double(Int.max)SIGTRAP;
actual progressTarget admits unrepresentable Double(Int64.max). Swift CLI proof,
not iOS render/E2E or fix. Fallback branches ignore masked in source; guide human+
agent/catalog/screen now state limitation instead of overclaiming universalprivacy.
Numeric design pending; no product/test code, provider/prod/device change. No live
UI/fault process confirmed. Prior goal turn progress; fullgoal active.

Independent20261015 numeric-safety audit: actual source money/progress/count and
JSONValue.prettyDescription can truncate/trap; Swift diagnostic confirms max
Double guard admits nonrepresentable Int64, finite1e100 JSON and1.5→1. Not actual
card crash reproduction. Bounded [design](work/tickets/NATIVE-AI-CARD-NUMERIC-SAFETY-20261015.md)
sent, approval pending; no product/test implementation. Previous goal turn is
progress, no live tests/proxy; fullgoal active, business refinement review pending.

Latest20261015 regression80693terminal65:273PASS/2FAIL0skip/no runtime warnings;
all271units+literalprobe+walletCRUD PASS. Category exact create/API scope PASS,
UI test reads subtitle in label instead of base accessibilityValue; later filter/
edit failures require independent proof. Transaction helper flags pinned saves
outside shell611 but inside window667; not financial-write failure proof.
Stop further guesses, bounded pinned-control/AX-filter test refinement sent for
review; no third fix yet. [Ticket](work/tickets/NATIVE-CURRENT-BUSINESS-SWEEP-20261014.md).
No tests running; full business remains RED/fullgoal active, goal422 separately
blocked. Public API/human/agent docs unchanged (test-only); no release/prod/device.

2026-10-15 owner approved semantic E2E reveal correction (test-only). Draft-only
RED54467 isolates appended names to twenty helper swipes on QuickType ScrollView
(0,407,375,44), before focus/save. No product autocorrection root proven; leave
base/API/keyboard settings unchanged. Existing targets resolve their scroll
ancestor by ID/label, lazy targets use only app viewport, ambiguity fails with
geometry. Probe46637 preserves names but full swipe overshoots139pt viewport;
bounded content drag probe34635exit0,1/1PASS0skip, no runtime warnings. No save in
probe. Exact persisted assertions retained; units+wallet/category/transaction
CRUD80693 now running. Captures inspected; compressed wallet form/Feedback still
not visually accepted. Guards/guardtests/project/diff pass.
Goal422 menu remains separately blocked; no production/device/release action.

Latest2026-10-15 diagnostic7487exit65:272PASS/2FAIL0skip (271unit+correct-wallet
transactionCRUD PASS), no runtime warnings. Category/wallet names match directly
after typing but append prose after focus/save; source/AX/videos recorded. No
product keyboard fix guessed; literal-name-only design sent for owner decision,
notes/chat unchanged. Normal8081 simulator defaultlaunch25086; proxy20834 stopped,
18081no listener, no running test sessions. Guards/tests/project/diff pass.
Business sweep still RED; goal422 menu review blocked, goal503+strong credit
locally verified. Confirmation-summary design pending; fullgoal remains active.

2026-10-15 terminal sweep86110:280PASS/3FAIL0skip (271unit+9UI), no runtime
warnings. Category root type assertion wrong; category/wallet names have appended
prose not in typed literals; transaction test tapped first Chi tiết of wrong
wallet. Source/AX/video inspected, not product-input root cause resolved.
Normal diagnostic7487 running with exact-name assertions and explicit fixture
detail ID; no keyboard/product fix guessed. [Sweep](work/tickets/NATIVE-CURRENT-BUSINESS-SWEEP-20261014.md).

Approved goal rerun94298 terminal65:273PASS/1FAIL0skip,271unit+strong credit+
goal503 PASS. Selector review fixed test assumption; goal503 same-key recovery,
API/rendered112500/remaining887500 and unchanged report proved. Credit explicit
source, pair, both balance deltas and unchanged report proved. Goal422 menu tap
opens ordinary Add; now blocked for root/design review, no fourth placement fix.
Captures98B2E756/76F1A2EF inspected; full visual/credit-label config audit remains.
Proxy20834 stopped after cwd/command verification; normal8081 diagnostic running.
First combined build26896 optional test-balance compile failure fixed via
XCTUnwrap, not a financial-value fallback. No prod/device/install/reset/release.

Current business sweep86110 confirmed running (normal8081 SE):271native units+
12critical API/UI scenarios, `/tmp/mypocket-current-business-sweep-20261014.xcresult`.
Credit test-source audit found default-source assumption; updated test only to
unique/select explicit source, pair/balance/month-report readbacks and scoped
cleanup. Current sweep compiled older test; updated credit must run separately.
Owner explicitly approved savings test-helper correction; implemented selector
name check/current-goal fallback without weakening assertions. Goal fault gates
will run only after current simulator owner finishes; no concurrent UI runners.
Confirmation-summary design still pending. No product/backend/prod/install/reset.

AI next-step audit: full confirmation-summary design sent, awaiting owner reply;
no summary implementation yet. Focused Go advisor/proposal race tests92901exit0
in two packages, log `/tmp/mypocket-advisor-current-audit-20261014.log`. The
versioned handbook now has a bounded deterministic `get_product_help` retrieval
slice and native/web `product_help` cards; it is still not live RAG or arbitrary
document retrieval. Its navigation was reconciled on 2026-10-15 to describe
Trợ lý as the primary tab; the stale Assistant-in-More wording is covered by a
provider regression. Catalog now reflects current8tool/6provider per-run safety
bounds instead of4; no user usage quota introduced. Stub tests are not liveLLM/
real-DB/prod proof. The approved wallet confirmation summary is now implemented
locally with a focused goal/credit contract; next safe work is durable action
replay or design/reproduction of remaining AI correctness gaps.

AI wallet-proposal validation local in_review: actual rendered RED four invalid
cards offered confirmation (150.5pt vs86pt invalid status). Strict shared decoder
now used by card; no coercion of invalid type/money, validates target/credit/date/
day and safe exact integer range, retains valid defaults/nulls. Final8709exit0,
279/279PASS0skip (271unit+8chatUI), no runtime warnings; guards/project/diff pass.
[Work](work/tickets/NATIVE-AI-WALLET-PROPOSAL-VALIDATION-20261014.md). Human/agent
guide/catalog/screen reconciled. The wallet confirmation summary and persisted
proposal `action_id` replay boundary are now implemented locally; the full
prepared-action dispatcher and unsafe numeric conversions elsewhere remain open.
No live creation/provider/prod/device/release or fullgoal claim.

Rejected transfer editor independently verified:52081exit0,269/269PASS0skip
(268unit+1actualAPI UI), no runtime warnings. Restart/unchanged-confirm refusal,
changed22500/newkey, one pair, source77500/destination122500, unchanged monthly
report and scoped detail pass; capture0134D5FF inspected, guards pass.
[Work](work/tickets/OFFLINE-RECOVERY-E2E-20261014.md). No product fix needed.
Savings-specific queue still blocked. Proxy8206 stopped; normal8081 build15493
exit0, installed without reset and default launch9608. APIhealth200;18081 no
listener confirmed. AI catalog corrected: nine non-mutating
tools including wallet proposal; UI confirmation now has a persisted proposal
identity, but is not proof of a general prepared-action dispatcher. No
prod/install-to-iPhone/release/whole-goal completion.

Owner approved adaptive chat. Implementation is now local in_review: keyboard hides
shell tabs without reserving their height, overview moves to header, explicit
dismissal preserves draft; accessibility navigation uses shared full-label menu.
New interaction RED verified (tabs still present/missing dismiss action before
code). Final20967exit0:276/276PASS0skip (268units+8UI), no runtime warnings,
`/tmp/mypocket-chat-approved-final-20261014.xcresult`. AX menu Account→Assistant
and normal tab roundtrip retain draft; actual API landscape conversation63pt.
Guards/project/diff pass; captures inspected with crop/transition limits in ticket.
No production/device/whole-goal completion claim.
This supersedes the pending-approval wording in the historical evidence below.

Adaptive chat RED:26992 landscape7pt; one-line cap10130 unchanged7pt, reverted.
82726actualAPIoverview0pt (fake diagnostic absentframe bug corrected). Final88928
exit65,272PASS/2landscapeFAIL0skip,268units+4otherchatUI pass,
`/tmp/mypocket-chat-adaptive-final-diagnostic-20261014.xcresult`. Frame9E8C4D1E
composer y103.5/h53 overlaps tab y129/h37, convo0; fake7pt.843995F5 inspected.
Largest-text focusedcontrol PASS but tabglyph/label overflow notaccepted.
[Work](work/tickets/NATIVE-CHAT-ADAPTIVE-20261014.md) blocked for bounded design
review: keyboard tabs hide/explicitdismiss/overviewheader; AX navigation full-label
menu throughexistingbase. No approval yet/no newUIcode.4invocations,1reverted
productattempt, no further guesses. Guards/project/diffpass, simulatorportrait
normal8081launch99198. No prod/device/reset/release/fullgoalclaim; cache/goalqueue/
Feedback review gates unchanged. Publichuman/agent guide+registry updatedlimits.

Readcache manualRED: DeviceHub exactSE/localDEBUGfixture→Reports2026-10 then
2026-99+View. Both GETs40023:02:55; oldsummary/zerochart retained, no visibleerror.
Inspected `/tmp/mypocket-report-invalid-month-repro-20261014.png`.
Confirms invalidscope hiddenerror/stale retention only, not nonzero corruption/
offline/Months note incident. ReturnedfixtureOverview, no financialwrites/reset/
production/newbuild. [Audit](work/tickets/OFFLINE-READ-SCOPE-AUDIT-20261014.md).
Typed coherent snapshot approach approval stillpending→spec/plan; no codechange.

Read-cache audit: Overview has no cache injection, sequential partial publication;
Reports keeps old summary ahead of error while month input mutates, optional
metadata/rows fail into empty output; Months retains old summary/note on read
failure. Code risks, not runtime-reproduced incidents. [Audit](work/tickets/OFFLINE-READ-SCOPE-AUDIT-20261014.md).
Owner asked to approve typed query-scoped encrypted coherent snapshots, not local
report recomputation/generic HTTP cache; approach response pending. Architectural
brainstorming gate: then written spec/reviewed plan, no implementation this audit.
Fullgoal unchanged, no prod/device/reset/release; goalqueue/Feedback blocks remain.

Featureprivacy correction: APIError shared allowlist class/status;6featureprints
and localfixturedebuglabel no arbitraryreflection. RED93397 3/3→GREEN66267exit0,
272/272PASS0skip (268units+4chatUI), no runtime warnings,
`/tmp/mypocket-feature-privacy-green-20261014.xcresult`. Guards/project/diffpass,
APIhealth200. [Work](work/tickets/NATIVE-FEATURE-LOG-PRIVACY-20261014.md).
97F7C806 keyboardrender inspected, usablecomposer; tabbar height UX observation,
not defect/approval to hide (plan requires usable navigation). No per-screen
fault-injection/liveLLM/wholeUI acceptance inferred. Publichuman/agent guide+
architecture reconciled. Goalqueue/Feedback blocks unchanged, no prod/device/
reset/release/fullgoalclaim. Next independent cache/auth/AIactions/help/APNs gates.

Independent native HTTP privacy: actualstdout200/422 body/query RED39064→DEBUG
metadata-only. Final92498exit0,267/267PASS0skip (265units+ordinarycreate/update+
transferCRUD APIUI), no runtime warnings; base guards/project/diffpass,APIhealth200.
Bundle `/tmp/mypocket-http-privacy-green-20261014.xcresult`.
[Work](work/tickets/NATIVE-HTTP-LOG-PRIVACY-20261014.md). Assistant/Budgets/local
fixture reflected-error logs still open; APIClient fix is not wholeapp privacy.
Goalqueue/Feedback review blocks unchanged. No prod/install/release/fullgoalclaim.

Goal queue checkpoint blocked for Rule10 review: baseline94206 RED, compiler
11119/88671 failures resolved by typed editor factoring58326 buildPASS;27662
263unitPASS/2goalUIFAIL0skip. Both UI failures: restart retains selected goal,
test incorrectly requires aggregate transaction-wallet-selector. Proposed
test-only helper correction needs review before another cycle; financial refresh/
rejected-transfer editor gates still unverified. [Work](work/tickets/OFFLINE-GOAL-QUEUE-20261014.md).
No product reset/prod/device/release/fullgoal claim.

Latest recoveryE2E: ordinary422/409 changed-only/newkey/date, rejected/committed
discard and transfer503/drop onepair/correctbalances/unchangedmonthlyreport
verified. Proxy5/5; recovery28248 3/3,transfer4135 2/2,regression32155 267/2670skip.
Scoped detail37055 2UIRED/263unitPASS exposed5 siblingIDs; existingcontain pattern
on detailroot→35239exit0,265/2650skip,uniquegroup/scopednote for ordinary409/transfer
drop,/tmp/mypocket-recovery-contained-final-20261014.xcresult. B8F61E91 inspected
22500/day13/note, unchangedstyle. Finalnormal44956exit0,265/2650skip (263unit+2
APIUI) /tmp/mypocket-recovery-normal-final-20261014.xcresult. Guards/project/diff
pass; normal8081build restored, ownedproxy67789 stopped. [Work](work/tickets/OFFLINE-RECOVERY-E2E-20261014.md).
Next rejectedtransfereditor/goalqueue/cache/report/auth/a11y/AIactions/help/APNs/
prod/device. UnderlyingledgerAX visibility + Feedback stillopen; no prod/device/
reset/release/fullgoalclaim.

Latest foreground lifecycle: AppRoot owner/scene + defaultSDK NWPathMonitor→
queue, one FIFO retry/lease timer; inactive/path loss suspends enginegeneration,
preservesuncertainlease/key, logoutseparateauthblock. RED72468 startup/network/
timer→258units; pauseRED18554 lateack+nextwrite→85017exit0,263/263 (261unit+2APIUI).
FinalUI13577exit0,3/3PASS0skip,/tmp/mypocket-lifecycle-ui-final-20261014.xcresult:
Home/serverrecovery no backgroundwrite, foregroundautoapply, restart/menuRetry,
commitdrop replay all exact1serverrow/key. Finalunits94868exit0,263/263PASS0skip,
/tmp/mypocket-lifecycle-unit-final-20261014.xcresult includes earlylease/FIFO/
cooldown/auth/stalewake gates. Guards/project/diffpass; normal8081build restored,
ownedproxy stopped. AppliedD63C86B9 inspected; Feedback stillcovers row, blocked.
[Work](work/tickets/OFFLINE-FOREGROUND-LIFECYCLE-20261014.md). Next fullrecovery/
transferfault/goalqueue/overviewreportcache/auth/a11y/AIactions/help/APNs/prod/device.
No production/install/release/fullgoalclaim.

Latest validation/transfer checkpoint: RED67864 missingcategory allowed at new
enqueue/replacement; compatibilityRED69099 localinvalidPayload before legacy
HTTP. validateNewIntent now creation-only, frozen replay semantics unchanged.
Final40360exit0,256/256PASS0skip (255unit+1realAPI native transferUI),
/tmp/mypocket-category-compat-final-20261014.xcresult. Explicit source/destination,
balances100000→87500/112500→77500/122500→100000, reportflagsfalse and serverpair
deletion verified. No aggregate report UI proof inferred. Guards/project/diffpass,
APIhealth200. [Work](work/tickets/OFFLINE-CATEGORY-VALIDATION-20261014.md).
Publichuman+agent/design reconciled. Foreground/reachability/timer/backgroundpause
wiring still next, not implemented; fullgoal open, no prod/install/release.

Latest fault/compact checkpoint: owner approved three-dot recovery menu, existing
MPOverflowMenu44pt + card/text/badge retained. RED34418 missingmenu; GREEN23703
1/2 exposed known Feedback tap interception, not a new placement fix. Recovery
test scrolls clear without hiding overlay; final73398exit0,2/2PASS0skip,
/tmp/mypocket-compact-green2-20261014.xcresult. Real503 and committed-lost-response
restart/retry prove oneintent/key and one ledger row; commit/drop201/200/200.
Proxy3/3 and guards/project/diffpass. Pending375pt screenshot3CF00F85 inspected;
short rows compact, Feedback still overlays selector (blocked). Publichuman+
agent/screen docs updated. [Work](work/tickets/OFFLINE-FAULT-E2E-20261014.md).
Finalunit29293exit0,252/252PASS0skip,/tmp/mypocket-compact-unit-20261014.xcresult;
normal8081build restored, temporaryproxy stopped, APIhealth200.
No prod/install/release/fullgoalclaim; scheduler/fullrecovery/a11y/device remain.

Newestrecovery localcheckpoint: retryNow/discard/atomicreplaceRejected + optional
encrypted supersedesID. RED93734/6944→248units; coordinatorRED88919→251units;
timestampsemanticRED46389→252units. Final5403exit0,252/252units0skip,
/tmp/mypocket-recovery-editor-20261014.xcresult. SQLiteinsertfailure rollback and
twostoreidempotent replacement proven; state/owner/key/content checks. Sharedqueue
Retry/Edit/Discard+warning, reused prefilledordinary/transfer editors/confirmnewkey,
preserveoriginaldate when unchanged. Guards/project/diffpass. [Work](work/tickets/OFFLINE-QUEUE-RECOVERY-20261014.md).
Normalcreate/editUI58897exit0,1/1PASS0skip,
/tmp/mypocket-recovery-normal-ui-20261014.xcresult; pending/conflict/date/retry/discard/accessibility/
restart/networkE2E not yet proved. Goal-scopedqueuevisibility stillgap. Nextthese
gates + foreground/network/background/auth/allremaininggaps. No prod/install/release.

Newestqueue runtime: defaultAuth/root/env→sharedcreateforms StateUUID/durablequeue,
separatependingbaseUI/acknowledgedrevision ledger+overview subscriptions. RED3560
→242units; actualcreate/editUI45495exit0,1/1 /tmp/mypocket-queue-create-ui-20261014.xcresult,
APIreads category/amount/note. TestappSQLite1row/sealed563bytes container449392A1,
no decryption. Screenshot4685B100detail inspected, not pendingUIproof. Concurrency
RED41241→drainEpoch, final23701exit0,243/243units0skip,
/tmp/mypocket-queue-final-20261014.xcresult. Guards/project/diffpass.
[Work](work/tickets/OFFLINE-RUNTIME-QUEUE-20261014.md). Nextperitemrecovery/pending
E2E/network/background scheduler/authlifecycle/remaincache/fullnativegaps. Matching
backend rollout required, no prod/install/release/goalclaim.

Newestcache checkpoint: AESGCM/atomic/Keychain owner+server snapshot, default
Auth→Transactions injection, ownerboundreads and transientonlyfallback/cachedAt
sharedstatus+retry. RED95867cache/90809model/41489HTTPstatus;503{} exposed optional
ProblemDetails.status bug, JSON/multipart now preserve authoritativeHTTPstatus.
Final13769exit0,239/239(238units+actualAPIkeyUI)0skip,
/tmp/mypocket-cache-final-20261014.xcresult. LedgerUI17267exit0,1/1PASS0skip,
/tmp/mypocket-cache-ledger-ui-20261014.xcresult. Screenshot59D42ADA inspected,
Feedback stillcovers lastamount knownblockeddesign. Syntheticlastknowncache may
retain testrows afterservercleanup. Guards/project/diffpass, XcodeGennewfiles.
[Work](work/tickets/OFFLINE-FINANCIAL-CACHE-20261014.md),
[publichuman+agent guide](guides/offline-sync-api.md). Next fullruntimeoutbox/forms/
queue/foreground/auth/refetch, overview/reportcache and staleUIaccessibility proof.
No offlinewriteactivation/prod/sign/install/release/fullgoalclaim.

Newest syncengine localcheckpoint: Task5 actor/policy implemented, not composed
into app. RuntimeRED45673→228unitGREEN; lifecycle exposed unsortedJSONwire body
regression24753, deterministicRED80149→outboxonly sortedencoder. Final52696exit0
232/232unitPASS0skip,/tmp/mypocket-sync-engine-final2-20261014.xcresult. Fresh
ownerboundcapability, FIFOleaseCAS, retry/permanent/conflict/auth, explicit
sameownerresume/logoutlatecallback, actualAPIClientadapter/SQLite timeoutreplay
with externalHTTPfake. GoRace22266exit0,guards/project/diffpass. XcodeGen newfiles.
[Work](work/tickets/OFFLINE-SYNC-ENGINE-20261014.md). Coalescing+engine401live
integration proof pending; invalidationcallback not realrefetch. Next runtime/
cache/outboxUI/foregroundevents and fullnativegoal. No prod/sign/install/release.

Newest offline acknowledgement checkpoint: FinanceAPI.applyOutbox uses persisted
lease/key/frozen payload and expected-owner transport. Receipt/resource validator
checks ordinary owner/wallet/category/jar/trip/type/amount/report/date instant,
transfer distinct paired source/destination excluded rows. Valid runtimeRED63069
and38814, final70626exit0,223/223unitPASS0skip, bundle
/tmp/mypocket-ack-http-final-20261014.xcresult. RealSQLitelease/APIClient with
externalHTTPdouble, not live financial backendE2E. Base/guardtests/projectpass.
[Work](work/tickets/OFFLINE-CLIENT-ACK-20261014.md),
[human/agent guide](guides/offline-sync-api.md). No new backendwire/schema/package.
Next sync actor/backoff/capability/lifecycle, then cache/UI/events and all gaps.
No offline activation/prod/install/release/wholegoal completion.

Newest local runtime/APIUI and owner-refresh checkpoint: API8081 rebuilt from
current source, PID19887/session46368; dedicatedPG55434/Redis16379, prod untouched.
ActualUI keycreate/copy/revoke pass; extended51955RED reproduced stale secret and
revoked row resurrection,70976GREEN212/212 after transient keyID/filter fix.
Owner envelope/refresh RED18476fivecases; CAS RED86604twoactualstore cases. Final
58452exit0,219/2190skip(218unit+actualAPIUI), bundle
/tmp/mypocket-owner-session-cas-final-20261014.xcresult. Keychain16contenders one
winner; old refresh never overwrites new same/otherowner login or cleared session.
sendEnvelope retainsmeta, expectedOwnerID localguard, no new owner wire field.
[Work](work/tickets/OFFLINE-CLIENT-ACK-20261014.md), [ADR022](decisions/ADR-022-owner-bound-native-refresh.md).
Actual account re-open render FF8AB1F2 inspected: revoked key/secret absent,
Feedback over timezone control confirms known blocked design. Guard/project/diff
pass;0active synthetic testkeys/0wallet/0tx/schema32false. Next Finance outbox
receipt+resource validator, then sync actor/cache/UI/events and all other gaps.
No financial adapter/sync activation/prod/install/release/fullgoal completion.

Owner full-key decision implemented local: common RequireAuth JWT/activeownerkey,
no advisor/feedback scope gates, create no scopes, old SQL arrays ignored/preserved,
metadata compatibility[] only. APIkeylogout selfrevoke, JWTlogout unchanged.
Native selector removed/shared warning. FinalGoRace18813/vet0; native97143exit0
212/2120skip211unit+1UI. [Work](work/tickets/API-KEY-FULL-ACCESS-20261014.md),
[guide](guides/api-keys.md), [ADR021](decisions/ADR-021-full-owner-api-keys.md).
ActualDB write/replay/transfer/owner/keymanagement/expiry/revoke proof. Inspected
SE screenshot has Feedback covering helper; known design block, not visual UAT.
No APIrestart/prod/install/release. Next updatedlocalAPI/accountmutationUI proof,
then sync/cache/UI and entire remaining goal. No transactions:write question.

FrozenTravel gate locally verified: optional create-only auto_link_travel=false
prevents delayed active-trip fallback; omitted/null/true online unchanged. Old
canonical receipt hash preserved, changed policy409, explicit inactive owner trip
accepted/deleted or foreign new intent404; completed replay before lookup retains
original date/trip without resurrection. Cap explicit_travel_v1 mandatory for
iOS ordinary offline writes; outbox freezes false before persisting. Native
211/2110skip59173; fullGoRace51567/vet0 after regeneratedSwagger. [Ticket](work/tickets/OFFLINE-FROZEN-TRAVEL-20261014.md).
No production/install/release/API-key change. Next implement owner's full-key
decision, then sync actor/cache/UI/session and remaining all-gap objective.

Newest owner decision: remove all per-key scopes; user API key must support the
same account operations as JWT. [Full-key design](work/tickets/API-KEY-FULL-ACCESS-20261014.md)
ready, implementation next after current Travel gate. No key permission code or
production activation yet. Preserve owner isolation/expiry/revocation/password
confirmation and service-token boundary; old scopes do not constrain target.
Do not ask again for transactions:write; that question is superseded.

Encrypted outbox now implemented locally, no screen/sync instantiation. SQLite3
WAL/FULL/IMMEDIATE + AES-GCM + per-account/server device-only Keychain key. Stable
enqueue/FIFO/isolation/lost-key refusal and transactional leases/outcomes/auth
block/resume verified. RuntimeRED70442 six missing-persistence tests,43035 four
missing-lease tests; final88836 exit0,209/209units0skip at
`/tmp/mypocket-outbox-store-final-20261014.xcresult`,16store cases. Base guards,
project verification and diff pass. [Ticket](work/tickets/OFFLINE-ENCRYPTED-STORE-20261014.md),
[ADR020](decisions/ADR-020-encrypted-native-outbox.md). No prod/install/release/API
change. Next: freeze entry-time Travel contract/capability, implement sync/cache/
UI/session integration; retain full AI actions/RAG/APNs/UI/rollout/device goal.
Typed markOutcome replaces four planned setters, not sync retry policy. Missing
Keychain key never causes ciphertext wipe/replacement. Full task/goal still open.

Latest: GET /sync/capabilities + typed iOS getter/fail-closed policy implemented
local, no sync wiring/prod activation. Backend404 RED76809, focused1027/full Go
race41946/vet pass; actual JWT/Redis200beforelogout/401after. Native runtimeRED
2fail/1pass with fail-closed scaffold, final193/1930skip terminal77223 bundle
`/tmp/mypocket-sync-capability-native-final-20261014.xcresult`. Tests known2create
operations/version1/key/meta policy and old4041GET/no mutation; unknown/disabled
blocks. OpenAPI inspected, guide/API/spec reconciled; guard/project/diff pass.
DB fixture/Redis namespace0/schema32false; no restart/prod/install/release. Next
encrypted account-scoped cache/outbox store and actual sync actor consuming fresh
capability, while transactions:write security choice and all UI/AI/prod gates stay
open. [Capability work](work/tickets/OFFLINE-SERVER-CAPABILITY-20261014.md).

Latest auth proof: actual NewRouter/bcrypt/JWT + PostgreSQL55434/Redis16379,
full Go race48372/vet0. Owner create/replay; other owner404, missing/bad/expired
JWT401, real Redis session expiry401, logout/family revocation401 before receipt.
Valid finance:read key authenticates in service, but financial routes JWT-only
deny its write/replay401; wrong scope/expired/revoked key state verified. No
positive API-key financial-write support claim. New transactions:write choice
requested, security policy unchanged. DB fixtures/Redis test namespace0/schema
32false; full suite also exercises Redis audit stream. [Work](work/tickets/OFFLINE-AUTH-HTTP-20261014.md).
No prod/schema/install/release; continue capability/native encrypted outbox/sync
while write-scope decision pending. Full UI/AI/prod/device goal unchanged.

Latest: keyed create HTTP receipt branch implemented and locally verified;
201 first/200 replay, unchanged data object/array + meta.mutation, typed request
hash excludes generated defaults, receipt lookup before mutable catalog checks.
No-header legacy path unchanged/unprotected: required-key rollout not complete.
HTTP/DB validRED56514 then4cases pass, final full Go race45501/vet0. Covers
archive/delete retry/no resurrection,8-concurrent pair exactly2excludedrows,
invalid keys/owner isolation/context auth/legacy/failclosed. Actual JWT/API-key
middleware E2E still open. Native189/1890skip compatibility units terminal70137
pass (object+2-row array metadata), not an offline UI test. Swagger regenerated/
inspected; public human+agent guide/docs reconciled. Local fixture owners0/schema
32false. No API restart/prod/sign/install/release. Next native durable key/cache/
outbox/sync, server-capability/activation gate and remaining full UI/AI/prod scope.

Latest offline prerequisite: shared SQL logger privacy locally verified, not
deployed. Default constructor leaked bound values + cast-error text (validRED2).
Filter-only wrapper failed because GORM Scan recorder expands SQL first; class-only
Trace + sanitized SQLSTATE/generic errors now passes4 actual read-only DB tests,
including LogMode, Row/generic events and not-found semantics. Final full Go race
6719 exit0 and vet/diff pass. [Work](work/tickets/SQL-LOG-PRIVACY-20261014.md),
[ADR019](decisions/ADR-019-financial-sql-log-redaction.md). No route/auth/schema/
app UI or production changes, no historic log deletion. Next remains HTTP key/
metadata/hash contract + integration before encrypted cache/outbox/sync; full UI/
AI actions/RAG/APNs/prod/device objective active.

Latest backend offline prerequisite: reusable create receipt capability added
without HTTP wiring or migrations. Reuses transaction_mutations22; fixes existing
GORM candidate-ID lookup bug affecting all shared receipt replays. Real DB5cases
cover replay/conflict/owner/concurrency/rollback/pair exclusion and old patch/delete
after deletion without resurrection. Full Go suite86668, focused race73167 and
final full Go race85540 pass; vet0. Local fixture owners0/schema32false. No API
restart/production changes, no offline availability claim. [Work](work/tickets/OFFLINE-CREATE-RECEIPTS-20261014.md),
[ADR018](decisions/ADR-018-reuse-durable-financial-receipts.md).
Next: freeze HTTP header compatibility/array-safe metadata/hash, financial SQL
logging privacy gate, route integration before encrypted native cache/outbox/sync.
UI blocked bubble design/provider/a11y/full-device/prod scope remains active.

Latest portfolio density slice: asset position values now reuse MPMetric compact
like summary; section/grid rhythm uses8pt token, no fields/unknown prices/history
hidden, no base font/palette/API changes. Real UI RED66252: asset cost line27.5pt
versus compact20.5pt; final71961 terminal0,
`/tmp/mypocket-portfolio-metric-density-green-20261014.xcresult`191/191 PASS0skip
(188unit+3API UI: metric consistency, asset error/cancel/create, price readback).
CaptureF514F91B inspected: values consistent and price/history reachable, but
Feedback covers asset title/quantity. This is the already blocked content-region
avoidance issue; do not make fourth placement attempt or accept whole screen.
Task1/6 density/all-screen gates still open; no prod/sign/install/release.

Latest independent navigation slice: removed duplicate More advisor item while
retaining typed route; primary Assistant tab keeps unsent input across tab changes
through volatile shell binding (standalone routes local fallback). RED7177
duplicate and96308 lost-draft observed. Initial195/196 failed only because test
keyboard helper inserted Return/newline; corrected test switches tab directly.
Final41071 terminal0, `/tmp/mypocket-assistant-navigation-final-20261014.xcresult`
196/196 PASS0skip (188unit+8UI), includes all remaining More destinations/back,
secondary shell, actual API overview, suggestions, keyboard and draft. Retained
draft screenshot inspected; auth error in fake-token fixture is expected and not
provider/prod evidence. Guard/project/diff pass. No signed install/prod/release;
full goal and previously blocked bubble-design slice remain open.

Newest checkpoint: AI-entry pinning/avoidance gate
`/tmp/mypocket-floating-complete-controls2-20261014.xcresult` completed196/196,
0 failures/skips (188 units+8 selected UI). Exported and inspected header/footer,
proposal and actual imported-photo captures. Header/Analyze/Save-all and preview/
remove/approval controls are protected in these cases. However the proposal render
places Feedback over the description text field placeholder. This is NOT visual
acceptance. Third product-fix pass exhausted: this avoidance slice is blocked for
design review under Rule10; do not add another speculative placement patch.
Other work remains active: Task3 evidence gaps, Task1/5/6 full-screen conformance,
offline/AI actions/RAG/APNs/prod/device scope. No signed install/prod/release.
Independent management regression completed55284 exit0,7/7 PASS,0skip at
`/tmp/mypocket-management-current-regression-20261014.xcresult`: budget CRUD,
ended-budget no-write/delete-cancel, jar month/draft CRUD, recurring CRUD/run-due,
travel CRUD/toggle, optional date/context clear and transaction link/unlink.
Read summary and each test node; create-error/all-screen visual/a11y remain open.

Latest: UI plan Task4 locally complete, ledger script75502 terminal0 and final
`/tmp/mypocket-task4-complete-20261014.xcresult`193/193,0skip (187unit+6UI),
after prior identical193/193 gate. Real stock Photos image appended without
replacing19 prior files, full20 bound; actual Photos/Files cancel/switch, uncertain
request recovery and disabled capability preserved. Data-loader characterization
refactor proves2valid/2errors/cap4/one active load/cap0; mixed normalized JPEG/PDF
wire preserves raw bytes and Vietnamese filter/timezone/session. No OCR/model/
financial writes. Screenshot of imported flower thumbnail inspected.
Unsigned prod-URL Release compile46172 exit0; do not install unsigned output.
Guards/project/diff pass. Human+agent-readable native/public docs reconciled;
no new API/schema/auth changes, no signing/install/prod/release, no live handles.
Next: Task3 remaining create/error/capture proofs, then Task1/5/6 all-screen
conformance/provider/keyboard/accessibility; AI entry header/Save-all scrolls out
and Feedback partly covers process CTA (captured Task6 finding). Full offline/
AI actions/RAG/APNs/prod/device objective remains active, not a UI-only completion.

Next active slice: Task6 AI-entry pinned header/Save-all/Analyze, detail approved
under common-family plan. Tests0/2 validRED terminal61061 at
`/tmp/mypocket-ai-entry-pinned-red-20261014.xcresult`; no scaffold implementation
changed yet. Reuse MPEditorScaffold via typed custom-header initializer, preserving
all existing title/onCancel consumers. Additional assertion now protects last
attachment remove action from bubble displacement; this assertion not yet run.
Do not fix CTA overlap by blocking that X or adding a reserved lane. Shared
compact-control avoidance may be needed; contract before code. No live handles.

Newest continuation: Task4 Photos-first source base/image policy implemented.
Source default/switch/count20/removal/cancel and actual Photos+Files system cancel
now pass; blank body root was unbounded EmptyView footer covering sheet fields,
fixed by explicit zero-height absent footer within base. Button-host/timing
hypotheses alone failed; final185units+4UI189/189 zero skips terminal53512,
`/tmp/mypocket-photos-footer-bounds-20261014.xcresult`. All3 source/system
screenshots inspected. Real selected-photo append/partial-cloud/mixed upload
proof and device acceptance remain open, so Task4 is not complete.
Portfolio/price/trade+credit route regression4/5 initially passed; credit failed
old purchase label plus payment transfergeneric title. Shared row payment title
RED→GREEN; final credit UI verifies amounts/category/pair/report exclusion.
Synthetic local trade fixture removed exactly after test; no prod mutation.
Guards/project/diff pass. No live test/signing/install or release publication.
Next: actual photo import/partial/multipart proof, remaining Task3 create/error/
capture evidence, then all-screen/chat provider/accessibility and broader goal.

New continuation: canceled-edit loss reproduced4/4, resume guards implemented
for budgets/jars/recurring/travel with fresh target metadata. Subsequent gate
184/185 passes180units, Budget/Recurring/Travel retention and ended-budget preview;
Jar month selector needed native menu scrolling. Corrected test then reproduced
real cross-month draft retention; month-change resets target/inputs now. Final
180unit+8UI gate20421 terminal exit0,188/188 PASS,0 skipped at
`/tmp/mypocket-draft-retention-final-20261014.xcresult`. Ended screenshot inspected;
final check compares every decoded transaction field, not IDs only. No live tests.
SSH retry14936 terminal255
still SSH-agent signing communication error; no prod change. Goal remains active.

Latest continuation: management budgets/jars/recurring/travel now use the shared
editor route/scaffold; travel dates use optional account-timezone calendars.
Full-form travel PATCH explicitly sends empty strings to clear optional fields;
omitted/null still preserves them on backend. Footer reports floating clearance;
Feedback hides while typing and returns after keyboard dismissal, without lane.
Mixed API gate185/186 and recovery183/184 are NOT green suites: recurring was
recovered, remaining expanded Travel context test had unsafe keyboard gestures.
Its date-clear and180-unit wire checks pass independently. Selection-menu test
correction passed1/1 at `/tmp/mypocket-travel-selection-20261014.xcresult`:
create/context/date and clear/context/date actual API readbacks pass.
Aggregate unit/management/keyboard gate terminal94622 exit0:187/187,0 skipped
at `/tmp/mypocket-management-aggregate-20261014.xcresult`. Jar selector was
misspelled; explicit test terminal53497 exit0,1/1 PASS,0 skips at
`/tmp/mypocket-jar-explicit-20261014.xcresult`. No test process remains live.
Earlier pending retention/ended-budget proof is superseded by188/188 above.
No new signing/install/prod mutation or release publication.

Latest: UI-plan Task2 implemented/verified, uncommitted under scoped-worktree
ruling. Typed quantity/count/percent fields, locale-safe decimal normalization,
strict recurrence count, dedicated note editor. Full179 units pass; actual API
fractional trade/recurrence, report exclusion, note persistence/cancel proofs pass.
Task3 portfolio slice implemented: shared MPEditorScaffold, separate asset/trade
editors, short price sheet, flat asset metrics/history, busy guard and retained
draft on failure/cancel. API create/error/cancel/trade/price readbacks pass;
keyboard/footer and retry screenshots inspected. Release prod-URL compile passes
unsigned, not installed. Remaining Task3 budgets/jars/recurring/travel editors,
Task1 broader flattening/density, Photos, chat provider/accessibility, final sweep,
offline/AI actions/RAG/APNs/prod rollout/device acceptance remain mandatory.
See [validation](work/VALIDATION_MATRIX.md) for mixed-run failure and final proofs.

Owner approved ALL listed gaps. Execution is active, not awaiting plan review;
full scope in [completion matrix](work/tickets/NATIVE-FULL-COMPLETION-20261014.md).
Chat slice implemented: shared1–4-line composer, pinned compact overview/details,
flat assistant messages/trailing user bubbles, feedback clearance without lane,
hidden raised Add on Assistant. Local draft remains editable when service is
unavailable; submission remains gated. SE UI3/3 and actual local API overview1/1
plus full units171/171 passed. No live-provider/physical acceptance claim.
Task1 common card insets and flat metrics now in progress; final full-screen
sweep, offline/AI V2/RAG/APNs, prod activation and device build remain open.
See validation for exact evidence. Old signing process for stale source was
cancelled; do not install its product. No release/status publication.

Task1 partial: asymmetric surface padding12horizontal/8vertical and flat AI
metrics verified, full units173/173 on SE. Task2 note slice: dedicated shared
editor with save/cancel draft semantics; UI2/2, actual API create/update+DB note
readback passes on iPhone16. Combined iPhone16 run exposed3 ambient-device-state
AuthRestore failures; test-only preferences/device-store isolation fixes them,
full units173/173 on same simulator now pass without warnings. At that checkpoint numeric fields,
remaining flat-card consumers/editor routes and all-screen acceptance still open.
Unsigned Release verification passes with existing orientation packaging warning;
generated product is unsigned and must NOT be installed. Final signed app not built/installed.
Fresh prod read-only SSH attempt still failed SSH-agent communication, no change.

Owner approved retaining current palette and applying the five-family guideline
throughout iOS. [Execution plan](superpowers/plans/2026-10-14-native-ui-system-consolidation.md)
has six tasks and is approved/executing; inline/native execution is
the preserved method. No new product code/API/schema change in this planning
checkpoint. Don't re-ask palette/style questions; keep approved Manrope/colors.
Existing feedback API activation/signing approvals remain separate blockers.

Current priority: [native UI system audit](work/tickets/NATIVE-UI-SYSTEM-AUDIT-20261014.md).
Prod intake now42 reports (latest8 recorded). [Native screen/presentation registry](design/platforms/ios-screen-registry.md)
and [five-family guideline draft](design/platforms/ios-base-guideline.md) created;
six missing native screen specs added. This is source/navigation coverage, not
completed rendered all-screen UAT. Duplicate More account/notifications entries
removed with RED/GREEN UI1/1; broader guideline now approved and executing.

Owner-resolution feedback implemented locally: counted tabs, reversible
checkbox, independent of dev status/history. Backend current+isolated candidate
race suites pass; native171 units + sequential API UI1/1 pass; stale refresh
race independently reviewed/reproduced/fixed. [Rollout checkpoint](work/tickets/FEEDBACK-OWNER-RESOLUTION.md).
Backend candidate1f6a7f7d pushed/image built; prod migration32 applied after
validated backup with unchanged content hashes/counts (42feedback). **API image
activation blocked by pi SSH/1Password agent**, existing healthy auth image
remains running. Final device build waits codesign/Keychain; not installed yet.
Owner asked to approve SSH and guideline. Do not claim route deployed or publish
release notes; preserve installed density build and financial data.

Urgent installed-build density feedback: owner requests smaller card/item heights
and immediate reinstall without tests. Shared card padding is now8, controls44,
transaction/wallet glyph32 with no outside vertical padding, category row min44,
profile single identity row, section gaps8/row gaps4; disabled Account FAB hidden.
Fonts and financial/API logic unchanged. Prior173+8 evidence predates this pass;
updated geometry assertions are not executed at owner's request. Release signing/
installation and physical acceptance must be recorded separately after completion.

Density replacement now compiled Release successfully, strict signature verified,
production API URL checked and installed without uninstalling on owner's iPhone.
Tests deliberately skipped; no release publication. Owner checks compactness.

Native feedback finish pass uses flat 12pt-padded shared cards, compact medium
typography, unified transaction rows, horizontal profile, 8pt button corners,
ledger wallet selector in the header and optional feedback photo choose/remove/
replace. The old feedback dock is removed: an explicitly sized content viewport
ends at the 56pt navbar, with the 52pt bubble overlaid and no reserved lane.
169 unit + 4 compact UI tests pass; 8 API-backed UI tests pass, including actual
Face ID enrollment/logout/device-login, row consistency, reports/month state,
wallet scope, key copy/revoke and feedback. See [validation](work/VALIDATION_MATRIX.md).

Authenticated production queue fetch succeeded: 34 open records, no screenshots;
all are classified in [audit](work/tickets/NATIVE-FEEDBACK-PROD-AUDIT-20261014.md).
Do not repeat earlier 404/401 intake conclusions or claim all 34 fixed. Note-editor
separation, durable offline/cache, AI V2 writes/product RAG and remote push remain
open. A dim disabled Account-tab add circle also needs a focused reachability/
visual follow-up. No production feedback statuses or release notes changed.

Owner approved additive backend deployment preserving data. Auth-only commit
`12572c87` is pushed and deployed; migrations 29→31 clean, validated backup and
unchanged financial counts/checksums. API health200/API1/1/worker1/1/web0/0.
See [rollout](work/tickets/NATIVE-DEVICE-AUTH-ROLLOUT-20261014.md). Legacy session
logout can sign out other pre-upgrade sessions once; new session families remain
independent. No financial records or passwords were reset.

Paired iPhone 15 Plus uses a current Personal Team profile through 2026-10-20.
After owner Mac unlock, repaired the stale interrupted-signing manifest with the
same identity; strict/deep verification passed. Production-API Release app was
installed without uninstall and devicectl launch succeeded. Hardware Face ID
and owner visual UAT remain pending. Do not weaken Keychain or treat installation
as acceptance of every feature.
Keep unrelated dirty web/backend edits untouched; active branch remains
`codex/native-ios-port`. Release publication awaits owner acceptance.

## Historical checkpoints

2026-10-13 owner clarified reduced rounding means buttons, not cards.
Native shared button radius is now 8pt rather than 999pt, preserving 48pt
height/44pt hit targets and circular feedback actions. Token RED reproduced
999 versus 8; full native unit suite passes 164/164 on iPhone SE at
`/tmp/mypocket-button-radius-green-20261013.xcresult`. This is not completion
of the feedback audit: owner now wants scroll content to reach the navbar
without a reserved feedback lane, and reports excessive card/content density.
Those follow-ups remain open; do not claim all feedback fixed or publish
release notes before owner review.

2026-10-13 owner explicitly redirected follow-up work to the native iOS app,
not ongoing web maintenance. The remaining mobile feedback audit/fix loop is
active under [NATIVE-MOBILE-FEEDBACK-20261013](work/tickets/NATIVE-MOBILE-FEEDBACK-20261013.md).
Worktree contains unrelated uncommitted web/backend/design changes; preserve
them and scope implementation to `mobile-ios/` plus native design/work docs.
Native iOS feedback fixes now cover ledger row copy, global amount masking,
report inclusion control, month-period copy, remembered wallet scope, newest
populated week fallback, segmented-control motion, and Assistant card
elevation. Unit suite passes 156/156 (fresh result:
`/tmp/mypocket-ios-feedback-unit-final7-20261013.xcresult`) and focused UI
regressions pass. Release warnings-as-errors build and project/base guards pass.
The broad UI scheme stalled on More;
remaining screenshot-dependent polish, Face ID/logout intent, and queue
completeness still require resolution.
The direct production feedback URL returned 404 in this turn, so the previous
authenticated intake is usable but queue completeness is not established. Do
not publish changelog/release notes until owner approval. Re-fetch the queue
before claiming all reports are exhausted.

2026-10-08 native category picker visual parity: each root has its own shared
card and row copy/clear action match web. Eligibility is independent of
hierarchy context. API-backed budget, recurring and transaction flows verify
selection/persistence; full native unit/contract is 144/144. Release
warnings-as-errors, project/base guards and diff check pass. Local disposable
DB/API only; no endpoint/schema changes. Physical-device/owner acceptance and
feedback bubble collision remain open; owner selected option A (dedicated
shell dock), and its detailed layout/spacer migration is awaiting review.
Evidence is in `work/VALIDATION_MATRIX.md`.

2026-10-08 current-source verification refresh: backend without DB config
passes 240 tests across 16 packages. After applying existing migration 000030
to the local stage DB only (version 29 → 30, additive; no reset),
`go test -race ./... -count=1` passes 265 tests across 16 packages against
both local stage and a fresh disposable database; the disposable DB was
removed. Production was not accessed. Native iOS unit/contract target passes
142/142 with zero skips/failures on iPhone 16 Simulator, and Release
warnings-as-errors plus base/project guards pass. The feedback collision UI
regression remains unresolved and its dock detail design awaits owner review.
Physical iPhone 15 Plus currently reports unavailable, so this checkpoint is
Simulator-only. Evidence is recorded in `work/VALIDATION_MATRIX.md`.

2026-10-08 native category-list visual parity: API-backed iPhone capture
exposed wrong add/kind order, per-row pencil/delete controls and category-kind
subtitles. Native now follows the web action order, opens the editor from the
whole row, keeps delete in the editor, and derives `Áp dụng tất cả ví` /
`Hoạt động trong N ví` from the existing API `wallet_ids`. Focused API UI
passes 1/1, unit/contract tests 138/138, Release warnings-as-errors and base
guards pass. No backend/API/schema update was needed because scope data already
exists. Category screenshots and proof are in `work/VALIDATION_MATRIX.md`;
overall device UAT remains open, including the feedback bubble overlapping a
long-list row, now tracked in
`work/bugs/NATIVE-LIST-FEEDBACK-BUBBLE-COLLISION.md`.

2026-10-08 native category editor parity is implemented: create/edit now uses
a dedicated route with visible reusable wallet checkbox rows. The iPhone API
flow uncovered a real wire bug: Swift encoded `walletIDs` as `wallet_i_ds`, so
the Go API silently treated selected scope as empty/all. `CategoryModelInput`
now explicitly emits `wallet_ids`; category plus owner-validated wallet links
are persisted atomically. Wire test, repository test and API-backed iPhone CRUD
pass. No endpoint/schema/migration changed. Continue the remaining category
parity audit; screenshot/device acceptance and inactive-category semantics stay
open. All backend verification is on isolated local stage, never production.

2026-10-08 native credit entry now uses the shared searchable category tree
inside its existing action sheet. Type/wallet eligibility and draft values are
preserved; API-backed iPhone UI verifies the selected category ID after save.
The existing Go credit/category contract already satisfies the flow, so no
backend/schema change is required. Continue the native port audit with the next
unresolved behavior/visual parity item; production data remains out of scope.

2026-10-08 port audit found one explicit web/Stitch vs native discrepancy not
captured as an open backlog item: category create/edit is a dedicated route on
web but inline above the tree on native, and wallet applicability is a visible
card on web/Stitch vs a menu on native. CRUD remains implemented and API-backed;
existing proof does not cover visual/route parity. A draft backlog item and
validation record were created. This audit snapshot is superseded by the
implementation entry above: the owner authorized the dedicated route and it is
now implemented with API-backed proof.

2026-10-08 native wallet management parity: the iOS list now matches web
grouping and aggregate semantics, Sửa/Xong, row-to-edit, in-form delete, add
and disabled service actions while retaining Chi tiết. Simulator UI verifies
included/excluded wallets, included-only total, unscoped aggregate routing and
confirmed API deletion.
That test found GORM was converting explicitly excluded wallets to included;
repository creation now persists false transactionally. Backend `go test ./...`,
native 130/130, API UI 1/1, Release warnings-as-errors and base/project guards
pass. Stage fixtures were cleaned; temporary 8081 stopped and existing 8080
untouched. Owner visual/device acceptance remains open; no API/schema change.

2026-10-08 native wallet row visual parity: compared the Stitch capture and
web WalletSelectionList directly. Native selection rows now stack wallet name
over current balance and show the selected aggregate/wallet check through a
shared base row. API-backed iPhone 16 regression passed 1/1; screenshot and
RED/GREEN evidence are recorded in the port audit and validation matrix. Native
Chi tiết is retained as a compact trailing row action, while goal/credit details
remain. Pixel-level physical-device approval and wallet-glyph artwork parity
remain open; no public API docs were changed.

Wallet aggregate icon is now distinct from a basic wallet: native uses the
globe/categoryTeal treatment specified by the web contract. The API-backed
iPhone 16 regression and screenshot pass; remaining wallet-specific icon art
and authenticated device visual acceptance remain open.

2026-10-08 native notification inbox proof: the seeded unread-item UI path now
passes against an isolated local stage API on port 8081, including persisted
owner-scoped mark-read verification. The full native unit/contract target is
127/127; project/base guards and web design/type/build checks pass. The test
fixture was deleted and the temporary API stopped. The already-running API on
8080 was found to use its local default `postgres` database, so test work now
uses the explicit stage API rather than that service. The configured
DeepSeek-V4-Flash synthetic text extraction, tool-call wire probe and
provider-backed Finance Assistant send/history flow pass; usage keys are
persisted and the synthetic conversation was cleaned. Local OCR and S3
credentials are missing, so image/PDF live proof is blocked. Authenticated
visual/device acceptance remains open.

Native API fixture helpers are now explicitly routed to the same isolated
stage API by default (`127.0.0.1:8081`, overridable with
`MYPOCKET_TEST_API_BASE_URL`). A wallet create/edit simulator flow passed 1/1
with the app configured for that endpoint; its synthetic wallets were absent
after teardown. The temporary API was stopped and the existing port 8080
service was left untouched.

Portfolio write-path UAT exposed a real native history gap: the backend's
aggregate portfolio summary calculates positions from trades but returns an
empty history field. Native now loads each asset's immutable history from the
existing owner-scoped trades endpoint. The iPhone 16 simulator passed a buy,
position/history refresh and fresh API verification of fractional quantity,
price and fee (1/1); native unit contracts remain 127/127 and Release
warnings-as-errors plus project/base guards pass. The exact synthetic asset
and immutable trade were removed from stage DB after testing. Device visual
approval and live OCR/S3 remain open.

2026-10-07 native goal-wallet savings-ledger parity: the scoped goal route
now switches from the generic transaction shell to a Stitch-style savings
surface with shared compact wallet/search/overflow controls, centered balance
and remaining target, deadline/progress/advice, owner-scoped month ledger and
add-transaction CTA. The ledger ordering now places the transaction scope
between the centered balance and progress block through shared base views. The
focused API-backed simulator proof is 2/2 at
`/tmp/mypocket-goal-focused-final-4/Logs/Test/Test-MyPocketIOS-2026.10.07_17-15-08-+0700.xcresult` (2/2);
visual inspection is recorded
in the validation matrix, with the ordering regression at
`/tmp/mypocket-goal-order-green4/Logs/Test/`. Device/provider/OCR/Google UAT remains open.

2026-10-07 native goal CTA shell clearance: a red simulator regression caught
the scoped goal ledger's green “Thêm giao dịch” label underneath the shell's
raised add affordance. The CTA now uses a nested safe-area region around the
shared `MPButton` with a 104pt bottom cushion and 56pt visual lift. The
focused API-backed route is green 1/1 at
`/tmp/mypocket-goal-cta-green3/Logs/Test/`, with inspected screenshot
`/tmp/mypocket-goal-cta-green3-attachments/603309EC-18BF-451E-BAB3-FB1D3F7732B7.png`.
The post-fix full native unit/contract target is 110/110 at
`/tmp/mypocket-goal-cta-full-unit/Logs/Test/`; Release
warnings-as-errors and project/base/diff guards pass. Physical-device,
provider/OCR/Google and production UAT remain open.

2026-10-08 populated goal-ledger verification found that the previous CTA proof
used an empty ledger. On the Transactions route, an outer transactions
`ScrollView` wrapped the goal detail's own `ScrollView`; its bottom CTA was
therefore laid out against content rather than the visible viewport and
overlapped the shell with a persisted row present. The goal detail is now
rendered as the Transactions screen root. iPhone 16 API-backed scoped-ledger
and explicit-detail tests pass (2/2); the scoped test scrolls the real row fully
above the fixed CTA and checks both shell boundaries. Result:
`/tmp/mypocket-goal-ledger-reachability-final-20261008.xcresult`; inspected
capture: `/tmp/mypocket-goal-ledger-reachability-attachments/D2901CC9-C45F-4EEC-8CED-92C14702B4F5.png`.
The local fixture cleanup was verified (zero matching wallets/transactions).
Native unit/contract tests pass 142/142; ordinary weekly ledger and wallet
selection regressions pass 2/2; Release warnings-as-errors, project/base
guards and diff check pass. No API, schema, migration, backend, production data
or endpoint changed. Full API-backed UI suite and owner/device visual UAT remain
open.

2026-10-07 native authenticated visual capture: More/account and transaction
editor screenshots were exported and inspected against the supplied Stitch
reference. The actionable editor mismatch was fixed by presenting transaction
forms full-screen above shell navigation; the shared `MPReferenceScreenHeader`
now owns the close capsule and `MPActionIconButton` exposes the existing AI
receipt-entry route from the save footer. Focused editor smoke and local
API-fixture editor smoke both pass 1/1 at `/tmp/mypocket-transaction-editor-guarded`
and `/tmp/mypocket-transaction-api-guarded`; the full native unit target is
110/110 at `/tmp/mypocket-visual-full-tests-2`, with Release/base/project/web
guards green. Physical-device pixel approval and live
provider/OCR/Google/production feedback remain open.

The More visual follow-up also added the shared `MPCenteredScreenHeader` so
`Mở rộng` stays centered beside the support pill; the focused More route smoke
passes 1/1 at `/tmp/mypocket-more-centered-header`.

2026-10-07 native Auth layout parity: `AuthView` now matches the web/Stitch
AuthGate anatomy with shared `MPBrandMark`, direct canvas title/subtitle, flat
fields, full-width email sign-in, divider and compact Google action. Copy is
centralized in `AuthReferenceCopy`; password/refresh/Google PKCE behavior and
IDs are unchanged. Focused contract and unauthenticated UI pass, full native
unit target is 109/109, Release/guards pass, and screenshot evidence is
`/tmp/mypocket-auth-parity.png`. Real Google/device exchange and authenticated
pixel-level UAT remain open.

2026-10-07 native wallet glyph parity: `MPWalletRow` now composes the shared
`MPWalletGlyph` base with 48pt geometry and web-aligned basic/goal/credit
warning/success/teal tones. Wallet API/data/accessibility are unchanged; full
native tests are 110/110 and Release/base/web guards pass. Bespoke Stitch bank
artwork remains an explicit visual-UAT gap.

2026-10-07 native credit summary-card parity: the native Credit route now uses
one shared reference card for current debt, available credit, limit and payment
status/due details, with web-equivalent entry/payment action IDs. The focused
contract is 1/1, the full native unit gate is 108/108, and Release/base/web
checks pass. The bounded authenticated credit API UI smoke passes 1/1;
the follow-up keeps actions before the statement and renders the empty credit
message as plain text outside the list card. Pixel-level visual UAT remains
open. No API, schema, migration or production data changed.

2026-10-07 native portfolio summary-card parity: Portfolio now shows the web
supporting subtitle and a shared `Tổng danh mục` reference card around headline
cost/value/P&L metrics. The focused contract is 1/1, the full native unit gate
is 108/108 and Release/base/web checks pass. A new disposable UI rerun hit the
known Xcode runner hang; prior mutation proof remains recorded and authenticated
visual UAT is still open. No API, schema, migration or production data changed.

2026-10-07 native account API-key UI proof: the disposable local API flow for
creating and revoking a key passed 1/1 on the iPhone 16 Simulator at
`/tmp/mypocket-account-key-ui/Logs/Test/Test-MyPocketIOS-2026.10.07_15-40-22-+0700.xcresult`.
No production credential was created or revoked; production credential UAT
remains open.

2026-10-07 native Reports API UI proof: the disposable local API flow for
Money Insider summary/detail passed 1/1 on the iPhone 16 Simulator at
`/tmp/mypocket-reports-ui/Logs/Test/Test-MyPocketIOS-2026.10.07_15-41-41-+0700.xcresult`.
Report APIs and owner scope are unchanged; authenticated visual UAT remains
open.

2026-10-07 native Finance Assistant overview API UI proof: the disposable local
API route passed 1/1 on the iPhone 16 Simulator at
`/tmp/mypocket-assistant-overview-ui/Logs/Test/Test-MyPocketIOS-2026.10.07_15-42-43-+0700.xcresult`.
Provider-backed send/history/action behavior remains an external UAT gate; no
model/provider call was made by this proof.

2026-10-07 native feedback submit API UI proof: the floating bubble and submit
flow passed 1/1 against the disposable local API at
`/tmp/mypocket-feedback-ui/Logs/Test/Test-MyPocketIOS-2026.10.07_15-43-30-+0700.xcresult`,
with screenshot storage unavailable. Production credential/storage UAT remains
open.

2026-10-07 native AI-entry attachment UI proof: the deterministic 20-file
fixture renders thumbnails, disables the picker at capacity and re-enables it
after removal; the UI test passed 1/1 at
`/tmp/mypocket-ai-entry-attachments-ui/Logs/Test/Test-MyPocketIOS-2026.10.07_15-45-02-+0700.xcresult`.
Live OCR/provider decision UAT remains open.

2026-10-07 native AI-entry proposal review proof: the simulator fixture
requires explicit income/expense selection before saving and passed 1/1 at
`/tmp/mypocket-ai-entry-review-ui/Logs/Test/Test-MyPocketIOS-2026.10.07_15-45-57-+0700.xcresult`.
No proposal is auto-approved; live provider/OCR decision UAT remains open.

2026-10-07 notification persistence proof: `TestNotificationPostgresOwnerScopeAndIdempotentRead`
passes against the isolated local `mypocket_stage_v1` database, covering
owner-scoped list/read, idempotent mark-read retry and cross-owner rejection.
No notification-create route or production mutation was introduced; seeded
inbox UI UAT remains open.

2026-10-07 release-wide parity verification: backend `go test -race ./...`
passed with the isolated local stage DB; native unit/contract target passed
108/108 at `/tmp/mypocket-native-release-audit/Logs/Test/Test-MyPocketIOS-2026.10.07_15-50-40-+0700.xcresult`;
native Release warnings-as-errors, web typecheck/build/design and project/base/
diff guards pass. Provider/OCR, production credentials, physical signing and
pixel-level authenticated UAT remain open.

2026-10-07 portfolio action-anatomy follow-up: the native Portfolio route now
matches the web card structure for refresh, asset creation and trade entry;
the full native unit/contract gate is 107/107 and the Release/base/web guards
pass. The bounded unauthenticated launch smoke is green, while the disposable
authenticated portfolio API mutation and pixel-level visual UAT remain open.

2026-10-07 native Stitch port checkpoint: the supplied archive is recorded as
the visual reference while web/API contracts remain authoritative. Native
SwiftUI routes cover the implemented web capabilities and compose shared base
components; the final iPhone 16 Simulator unit/contract gate is 83/83 and the
warnings-as-errors Release build passes. Authenticated provider/OCR,
production feedback credentials, physical-device signing and pixel-level
owner UAT remain release gates rather than silently inferred completion.

2026-10-06 production backend rollout: a validated pre-migration custom-format
backup was retained on `vm100-prod`; the live database advanced from migration
27 to 29 without reset/drop/truncate; and `mypocket-api` was rolled to the
immutable release `release-73c4c37` at digest `sha256:d1f3f3b1bba9648192feccb814bd0823b9bc7df2c80518bf29a89cf77df691c2`.
Public health and OpenAPI probes returned 200. Redis is now reached through
the additive attachable overlay `mypocket-redis-net` and stable
`redis-mypocket` alias, verified with `PONG` from `vm100-prod`; no Redis
restart was needed. Commits `75f7f07` and `73c4c37` on
`codex/native-ios-port` are the source baseline for this deployed release.

2026-10-06 recurring worker enablement: commit `73c4c37` added `/app/worker`
to the amd64 API image. A one-shot production pass completed on `vm100-prod`,
then `mypocket-worker` was scaled from `0/0` to `1/1` with the same production
environment and stable Redis overlay. The worker remains responsible only for
idempotent recurring schedules; report snapshots and notification fan-out are
not enabled by this change.

2026-10-06 clean simulator verification: a fresh DerivedData build on the
booted iPhone 16 Simulator ran 68/68 native unit/contract tests and all 19/19
`LOCAL_API_UI` tests with zero skips or failures against a disposable local
API/database. The run covered wallet/credit/transaction/transfer/portfolio/
AI-entry/Assistant/feedback/route flows, including wallet-row → scoped-ledger
selection, the separate wallet detail/editor path, and API-backed portfolio
asset price refresh. Portfolio and wallet fixtures were removed after the run.
This is simulator-only evidence; production authentication/deployment and a
physical device remain external gates.

2026-10-07 native auth reconciliation: the iOS login surface uses the existing
`POST /api/v1/login` email/password contract ("username" means normalized
account email) and now also exposes the existing Google PKCE client through
`/auth/mobile/google/exchange`. The same access/refresh session store is used
for both paths; a placeholder iOS client ID fails locally without a provider
call. The final native unit/contract gate is 83/83, the focused auth launch/unit
smoke is 2/2, and the Release warnings-as-errors simulator build passes. Real
Google Cloud redirect registration, provider exchange and physical-device UAT
remain external gates. No production password was seeded or printed.

2026-10-06 wallet-scope reconciliation complete: web routes a wallet selection
to `/transactions?wallet_id=...`; native wallet management now uses the same
primary action. A separate explicit “Chi tiết” action keeps the wallet/credit
detail surfaces available without making the wallet picker and transaction
scope disagree.

2026-10-06 portfolio verification complete: native portfolio summary decoding
now handles `asset_id`, `realized_pnl`, `unrealized_pnl` and total P/L keys
without dropping the screen into an error state. The API-backed simulator
fixture creates an owner-scoped asset, refreshes its manual price through the
native UI and verifies the refreshed value; the unused fixture asset is
deleted with HTTP 204. Trade math and sell-over-position rules remain covered
by backend entity tests, while immutable trade fixtures are not deleted by the
UI harness.

# Project Context

2026-10-06 native simulator auth fixture correction: the API-backed UI harness
now uses a DEBUG-only in-memory session store when launched with
`-UITestAuthenticatedAPI`. The previous unsigned `CODE_SIGNING_ALLOWED=NO`
run reached the disposable OAuth exchange with HTTP 200 but stayed on the login
screen because Keychain returned `errSecMissingEntitlement (-34018)`. Production
and normal app builds continue to use `KeychainSessionStore`; no token values
are logged. The full iPhone 16 `LOCAL_API_UI` suite passes 16/16 after the
correction.

2026-10-06 native transaction editor parity: ordinary expense editing now
exposes the account-month hũ selector, report-inclusion toggle, and account-
timezone date/time field. Active hũ options are limited to eligible basic/goal
expense wallets; an existing inactive hũ remains visible with an explicit save
error instead of being silently remapped. Type, wallet and category changes
clear incompatible hũ assignments. Native XCTest passes 61/61, the signed
iPhone 16 `LOCAL_API_UI` suite passes 15/15, the Release warnings-as-errors
build and base/project guards pass.

2026-10-06 native AI proposal parity: the proposal editor now preserves the
account-local date/time, keeps an inactive existing hũ visible for explicit
resolution, rejects stale hũ values instead of silently clearing them, and
excludes the transfer-out category from hũ choices. Report inclusion remains
part of the versioned proposal patch; deterministic proposal review and policy
coverage pass.

## Latest AI entry correction — 2026-09-23

[Review-first bug](work/bugs/AI-ENTRY-REVIEW-FIRST-DETAIL_DESIGN.md) is in_review pending owner UAT. Replaced conflicting clarification-first prompt rules with partial draft extraction, within-batch duplicate merging and inferred/default supplied wallets. Real stored nine-image OCR replay changed from zero drafts to 30 with owned wallets after raising the truncated 2048-token response budget; the current provider envelope is 32768 tokens. The application no longer caps proposal count; only provider response/transport envelopes remain. Held-out duplicate/distinct-date/missing-date fixture, 31-draft parser test, full Go tests, focused race tests and backend build pass. Local API restarted; backend health and frontend :4173 both HTTP 200. No ledger writes from replay, no production deployment; owner still reviews amounts/duplicates and existing requests retain cached outcomes.

2026-09-23 diagnostics follow-up: AI entry and Finance Assistant provider calls now emit privacy-safe stage/latency/status/count diagnostics, and `AI_STORE_USAGE=true` persists normalized request/finish/latency/token metadata internally without exposing it in public responses. User text is sent as a separate `user_instruction` field, so explicit filters such as “chỉ lấy giao dịch tháng 9” guide selection while OCR/source text remains untrusted evidence. User-facing repository reads were audited for owner predicates; the only owner-independent operation is the internal attachment cleanup sweeper, which does not return financial data. Streaming remains deferred; the synchronous JSON contract and bounded 180s model/210s extraction timeouts are the current behavior.
2026-09-23 runtime regression fix: initializing empty `model_usage` maps on new AI entry/advisor rows prevents the non-null JSONB columns from receiving SQL `NULL`, which had surfaced as a generic 500 before provider work. The database default remains as a safety net; creation and integration tests now cover the path.
2026-09-30 production feedback fix: AI entry normalizes recognized date-only model output (`yyyy-MM-dd`, `dd/MM/yyyy`, `dd-MM-yyyy`) in the account timezone and explicitly flags the missing time for review. Exact duplicate drafts from one OCR/text submission are merged while preserving duplicate-row questions. Backend package tests pass (176 tests); production API image `feedback-api-ai-date-20260930-0011` is healthy.
2026-09-30 auth feedback fix: login and Google OAuth now issue short-lived JWT access tokens plus rotating opaque refresh tokens stored hashed in Redis; `POST /api/v1/auth/refresh` consumes each refresh token once. The browser retries one expired session request, never API-key requests. Production API `feedback-api-refresh-20260930-0041` and web `feedback-ui-refresh-20260930-0032` are healthy.
2026-09-23 wallet/ledger UI follow-up: wallet selection now opens a shared detail panel for basic, goal and credit wallets; the transaction ledger has a top wallet scope select with `Tổng cộng` or an owner wallet. Credit detail is read-only until its dedicated debt ledger exists.
2026-09-24 ledger correction: `/transactions` has one wallet selector. Basic and `Tổng cộng` now share Week/Custom controls, period cashflow card and grouped rows; basic adds only a wallet-ID predicate. Goal alone shows progress and full monthly history; provisional credit keeps the shared period layout read-only. No shared All-time tab. Live Chrome verified basic/aggregate parity and previous-week empty state after the correction; owner mobile sign-off and deployment are not claimed. Credit statement/payment/detail contract is not approved; its visible rows cannot open the ordinary editor. Travel Mode is now an implemented event linkage, not a wallet type; the `/account/travel` panel manages one active owner-scoped event and ordinary transaction auto-linking, while report narrative remains a follow-up. See [ticket](work/tickets/UI-LEDGER-WEEK-DETAIL_DESIGN.md) and [Travel Mode contract](design/pages/travel/README.md).
2026-09-30 wallet UI reconciliation: the transaction ledger scope control now uses a compact `WalletScopeSelector` trigger and the shared `WalletSelectionList` inside `BaseBottomSheet`, matching the supplied wallet-selector reference while preserving aggregate/basic/goal/credit filtering semantics. The selector can route to wallet management for add/edit; no API or schema change.
2026-10-05 simulator/integration audit: iPhone 16 native unit/UI smoke (19 + 3) passed, the Go backend race suite passed against disposable PostgreSQL + Redis, transfer integration/browser probes now cover idempotent pair update/delete, and feedback was verified through JWT, scoped API key, service-token agent lifecycle and Redis audit events. Local fixtures were cleaned. Provider-backed AI/OCR, real Google iOS OAuth, physical-device and authenticated staging/production proof remain open release gates.
2026-10-06 final simulator/provider pass: the booted iPhone 16 Simulator completed 65/65 native contract tests and 17/17 signed `LOCAL_API_UI` tests. Backend race/vet passed. A fresh live DeepSeek-V4-Flash S3 → OCR Base64 → LLM probe passed 5/5 cases and 24/24 semantic fields (100%); the extraction prompt now classifies paid receipts as expenses unless income/refund/reimbursement received is explicit. Real Google iOS OAuth, production-authenticated feedback, physical hardware and deployment approval remain external gates.
2026-10-06 AI browser regression: a fixed September/October fixture date had drifted outside the current account-local week and caused a false browser failure. The fixture now derives today/yesterday in `Asia/Ho_Chi_Minh`; the complete AI contract + Chromium flow is green again.
2026-10-06 transaction category integrity: migration `000028_transaction_category_required` backfills existing ordinary rows with `expense_general`/`income_other` and rejects uncategorized ordinary income/expense rows at the database boundary. HTTP handlers, recurring/credit repositories, AI approval validation, web forms and native editors now require an applicable category and prefill the first valid group; adjustment rows remain the explicit exception. The simulator red-to-green category policy test passes.
2026-10-06 category integrity simulator rerun: Release warnings-as-errors build passed; fresh iPhone 16 Simulator unit/contract run is 66/66 and the default UI target is 6 passed with 11 intentional `LOCAL_API_UI` skips and 0 failures. The API-backed fixture suite remains a separate opt-in gate.
2026-10-06 quality gate: backend AI/feedback test fixtures were normalized with `gofmt`; `go vet ./...`, targeted Go tests and a Swift Debug simulator build with warnings-as-errors pass.
2026-10-06 native parity fix: session restore now refreshes expired access tokens before Face ID and clears invalid sessions safely. Category editing now uses a shared multi-select base and preserves every selected wallet ID instead of collapsing a multi-wallet category to its first scope. Refresh rejection during an authenticated request now invalidates the AuthViewModel instead of leaving a stale shell. Shared native buttons now use their full rendered hit shape, and the simulator route sweep opened all 13 “Thêm” destinations. Malformed ledger timestamps are excluded from period filters, wallet detail renders the complete API result, and assistant polling keeps a timed-out run cancellable. iPhone 16 Simulator unit tests pass 25/25 and UI smoke passes 4/4.
2026-10-06 native parity fix: session restore now refreshes expired access tokens before Face ID and clears invalid sessions safely. Category editing now uses a shared multi-select base and preserves every selected wallet ID instead of collapsing a multi-wallet category to its first scope. Refresh rejection during an authenticated request now invalidates the AuthViewModel instead of leaving a stale shell. Shared native buttons now use their full rendered hit shape, and the simulator route sweep opened all 13 “Thêm” destinations. Malformed ledger timestamps are excluded from period filters, wallet detail renders the complete API result, and assistant polling keeps a timed-out run cancellable. iPhone 16 Simulator unit tests pass 24/24 and UI smoke passes 4/4.
2026-10-06 native simulator fixture: Debug launch argument `-UITestAuthenticatedAPI` now exchanges against the disposable local OAuth fixture, stores the real JWT in Keychain and renders the overview through live URLSession/API requests. Manual iPhone 16 screenshot `/tmp/mypocket-ios-simulator-api-auth.png` shows the authenticated empty-owner state; this is local validation only and does not certify production OAuth.
2026-10-06 native transaction parity: bulk-delete now matches the web contract (base selection, no transfer/credit/adjustment rows, max 100 IDs, atomic confirmation and retained error state). The typed API client now handles backend `204 No Content` responses for all empty mutations; native XCTest count is 28/28 and UI smoke remains 4/4.
2026-10-06 native Assistant parity: Finance Assistant now restores one owner-keyed opaque conversation ID, reloads the latest 50 messages and read-only monthly overview on entry, pages older messages, and keeps message text out of UserDefaults. The iPhone 16 API-backed UI smoke passed against the disposable local API with the capability-disabled/overview-200 contract; native XCTest count is 29/29 and standard UI smoke remains 4/4.
2026-10-06 native credit parity: the credit ledger entry/payment form now preserves entry kind, applicable category, source wallet, amount and optional note through the native base selectors/inputs. Category kind/wallet filtering is covered by XCTest; native XCTest count is 30/30.
2026-10-06 native wallet-detail parity: basic/goal wallet detail now loads categories, renders account-timezone monthly groups with signed amounts and labels, shows income/expense totals, and opens the shared transaction editor for add/edit. The API-backed simulator smoke seeded and removed a disposable wallet after opening the editor.
2026-10-06 native transaction-ledger parity: the shared overflow menu now owns period/grouping/transfer/adjustment/bulk actions; rows include category and wallet context with signed amounts, and goal/credit wallet scopes suppress ordinary period controls. Standard simulator UI smoke now passes 5/5.
2026-10-06 native selector/jar parity: API enum values now render localized labels through the shared selector base, jar mode changes clear stale drafts, and zero-percent allocations are rejected like the web editor. The clean iPhone 16 simulator run passes 35/35 native unit tests and 5/5 standard UI tests (3 API-backed tests also pass with disposable local fixtures; all fixtures were deleted with HTTP 204).
2026-10-06 native transaction boundary parity: ordinary and AI proposal editors now exclude credit wallets, invalid categories clear when wallet/type changes, and aggregate credit rows remain read-only outside the dedicated credit ledger.
2026-10-06 native scope parity: budget categories are expense-only and budget/adjustment wallet selectors exclude credit wallets, matching the ordinary ledger boundary.
2026-10-06 native budget parity: ended budgets are view/delete-only and no longer expose an edit/save path after their end date.
2026-10-06 native portfolio date parity: trade date selection now follows the account timezone and serializes at local noon to avoid device-timezone drift.
2026-10-06 native category scope parity: system categories now expose a lock/read-only metadata state, hide delete, and update only applicable wallet IDs through the dedicated wallet-scope route; personal categories retain full editing. Native XCTest is 49/49 and the signed simulator More-route sweep is 1/1.
2026-10-06 native jars month/cumulative parity: Hũ chi tiêu now uses a bounded account-local month selector, actual/allocated/unassigned metrics, advisory over-allocation warnings, active/history action gating, and selected-jar cumulative reports that clear stale state on month changes. Native XCTest is 50/50, Release build passes with warnings-as-errors, and the signed More-route sweep is 1/1.
2026-10-06 native savings-summary parity: goal-wallet detail now shows reached/remaining amount, account-local target-day status, bounded progress and target amount through shared base components. Native XCTest is 52/52 and the signed More-route sweep is 1/1 in 124.194s.
2026-10-06 native portfolio draft parity: asset symbol/name drafts are trimmed and whitespace-only values are rejected before API submission; the portfolio header now exposes the shared refresh action. Native XCTest is 53/53 and the signed More-route sweep is 1/1 in 124.144s.
2026-10-06 native feedback draft parity: title/description are normalized and feedback type is restricted before JSON or screenshot multipart submission. Native XCTest is 54/54 and the signed feedback-entry UI test is 1/1 in 9.821s.
2026-10-06 native month-summary parity: Tháng now decodes jar usage, uses bounded account-local month selection, renders period/status/category/jar details, protects dirty notes and links to the selected month’s Hũ view. Native XCTest remains 54/54 and the signed More-route sweep is 1/1 in 122.295s.
2026-10-06 native simulator release check: the Release simulator build passed with warnings treated as errors, project/base-component guards passed, and `com.mypocket.ios` launched on iPhone 16. The post-commit cross-surface rerun passed 48/48 native unit tests, 14/14 signed API-backed UI tests, backend race tests, and the web/API/Chromium/PWA gates; the disposable fixture was empty afterward. Provider-backed AI/OCR, real Google OAuth, authenticated staging/production and physical-device validation remain open release gates; no production OAuth claim is made.
2026-10-06 native AI proposal review follow-up: review now requires an explicit income/expense selection for unknown model output, exposes active account-month hũ choices only for ordinary expense rows, filters/clears incompatible categories, and lets the user choose report inclusion through shared base selectors. The red-to-green contract passes; the deterministic iPhone 16 review fixture passes and the full signed `LOCAL_API_UI` suite is 15/15. Provider-backed extraction remains an external release gate.
2026-10-06 native bug audit: wallet detail month headers now use validated calendar interpolation instead of literal placeholder text; Travel Mode trims required names and rejects malformed/reversed optional date ranges before API submission. The red-to-green regression tests and iPhone 16 run now pass 37/37 native unit tests and 5/5 standard UI tests.
2026-10-06 native UI test isolation: `testAppLaunches` now starts with the DEBUG-only `-UITestUnauthenticated` fixture, clearing only simulator Keychain/biometric state so a prior local login cannot change the assertion. Targeted red-to-green proof and the full iPhone 16 default UI suite pass 9/9 (5 passed, 4 explicit API-fixture skips).
2026-10-06 native API UI fixture isolation: `LOCAL_API_UI` now creates/deletes its own basic and credit wallets through local `:8080` after fixture OAuth. Wallet detail, credit ledger, Assistant and AI-entry capability tests pass, the full opt-in UI suite is 9/9, and a post-run owner check finds zero leftover `Fixture*` wallets.
2026-10-06 native AI-entry capability gate: the base document picker is disabled until `files_configured=true` and while processing; text-only entry remains usable when OCR/private storage is unavailable. Native unit tests are 38/38 and the local API-backed capability UI assertion passes.
2026-10-06 native ledger week-period parity: aggregate/basic scopes now use account-local Monday–Sunday `Theo tuần` plus bounded previous/next navigation and `Tùy chọn`; the incorrect all-time tab was removed and the summary uses `Tiền vào`/`Tiền ra`. Native unit tests are 41/41 and the API-backed ledger UI assertion passes.
2026-10-06 native goal target-date parity: wallet editor date-only parsing and serialization now use the account IANA timezone instead of UTC/device time. Native unit tests are 42/42.

## Field Ownership

- Human owns project intent, priority overrides, and unresolved product questions.
- AI owns concise state refreshes after work: touched areas, recent decisions, next steps, and queue summary.
- Shared fields can be updated by either human or AI, but AI must not silently override human priority.

## Current Status

- 2026-09-22 Stage v1 reset: the historical incremental SQL files were squashed into `backend/migrations/000001_stage_v1.up.sql` with a non-destructive down file. A fresh disposable database applies cleanly at version 1. Existing databases that still report an older historical version must be recreated/reset in the staging environment before using this baseline; production data is never dropped automatically. The former legacy reference tree was removed; this repository is now the canonical implementation and documentation source.

- 2026-09-22 planning follow-up (before implementation): owner requested detailed writing-plan and technical instructions. Added [technical guide](superpowers/plans/2026-09-22-finance-assistant-technical-guide.md) and expanded V1 into ordered subtask gates: concrete Go ports/SQL/API/SSE/UI contracts, isolated fixtures and operational checks. Code inspection found `JarRepository.ListMonth` writes month config; the advisor runtime now uses a read-only jar path. The later implementation checkpoint below is the current proof source.
- 2026-09-22 Finance Assistant implementation has started from the approved plan: semantic query foundation, advisor conversation/run and user API-key schema, read-only tool registry/provider loop, JWT/API-key endpoints, best-effort Redis advisor audit and `/assistant` UI are present locally. Replayable SSE, durable audit delivery/alerting, configured-provider proof, public deployment and merchant normalization remain release gaps; V2 confirmed actions and later intelligence remain roadmap.
- Feedback → Fix → Changelog is wired locally: persistence, service-token middleware/audit, owner-scoped HTTP API, JWT/API-key user auth (`feedback:read`/`feedback:write`), lifecycle locking, Swagger and Account → Phản hồi UI are implemented and verified by backend/frontend suites. The latest local agent lifecycle and Redis audit probe passed; owner visual sign-off and production deployment remain pending. The service token remains separate from user API keys and has no finance access. See [Feedback API](architecture/FEEDBACK_API.md) and [validation evidence](work/VALIDATION_MATRIX.md).
- 2026-09-22 transfer implementation checkpoint: the Stage v1 baseline includes nullable `transactions.transfer_id`; `POST /api/v1/transactions/transfer` atomically creates paired source-expense/destination-income rows with transfer categories and report exclusion. Quick Add has a shared-base transfer mode. The 2026-10-05 local harness now verifies pair update/delete with idempotency keys; owner visual UAT and external deployment proof remain pending.
- 2026-09-22 CORE-03 implementation checkpoint: the Stage v1 baseline includes account IANA timezone/date-only fields, jar/month APIs and screens, optional transaction jar selection, and live month totals/notes. Owner UAT remains pending; see [validation evidence](work/VALIDATION_MATRIX.md#core-03-timezone-jars-and-live-monthly-summary--2026-09-22). Monthly completion is derived, with no cron close/snapshot.
- 2026-09-22 AI usage policy follow-up: owner clarified there should be no user usage limit right now. The hardcoded AI-entry request cap was removed; idempotent replay still prevents repeating provider work for the same request key. Future provider pricing/cost consent/usage budgets remain draft [AI-ENTRY-03](work/tickets/AI-ENTRY-03-PROVIDER-SELECTION.md), and the current no-limit decision is captured in [AI-USAGE-01](work/tickets/AI-USAGE-01-DETAIL_DESIGN.md).
- Dedicated shared-code scheduler request is separately captured as high-risk draft [WORKER-01](work/tickets/WORKER-01-cron-service.md), with recurring transactions and month-end reporting/close as candidate jobs. No worker/runtime/schema implementation is approved; report artifact semantics and recurring edge rules need resolution first.
- 2026-09-21 implementation follow-up: [AI-ENTRY-01](work/tickets/TICKET-09-01-ENTRY-DETAIL_DESIGN.md) implements normal Add/manual vs hold/AI entry, API-backed proposals, editing/approve/reject and atomic replay-safe ledger writes. Migration 000010 applied; real DB HTTP/concurrency tests pass. The 2026-09-22 strict JSON Schema follow-up now passes a synthetic five-case live extraction benchmark; owner UAT remains pending.
- 2026-09-22 AI-ENTRY-02 follow-up: one-shot multipart `POST /api/v1/ai/entry/process`, read-only recovery and independent proposal APIs are implemented. Backend stores originals in environment-qualified private S3, OCRs via a short-lived signed URL before LLM, and sends OCR text only. The Stage v1 baseline owns attachment/OCR metadata, approval links and cleanup claims; approval links are atomic and authenticated download checks owner/link. Explicit cleanup command is implemented but not run against bucket contents; browser download UAT remains pending. Do not retry the user's prior PDF automatically.
- 2026-09-22 LLM resolution: JSON-object mode previously scored 0/15; strict OpenAI-compatible JSON Schema mode now passes 5/5 synthetic cases and 24/24 scored fields. Live S3 PUT/readback 83 ms, OCR 4.034 s, LLM p50 19.769 s/p95 25.655 s, OCR+receipt-case pipeline 29.689 s; transfer-safety passes. Provider calls use the existing standard `net/http` adapter, with no additional SDK. Saved wallet descriptions now enter the bounded, untrusted wallet-matching context (2 KiB each). No transactions were approved or written. Broader model evaluation and owner UAT remain pending. Future multi-provider selection and sourced price transparency are captured in draft [AI-ENTRY-03](work/tickets/AI-ENTRY-03-PROVIDER-SELECTION.md); no additional provider was enabled.
- Local test servers restarted with latest code: frontend `http://localhost:4173/` returns 200 and backend `http://localhost:8080/api/v1/health` returns 200. Keep them running for owner UAT.
- 2026-10-06 native simulator regression: wallet detail now fails closed until
  transaction/category metadata is loaded. A custom decoder bug for plural
  `wallet_ids` (`walletIDs` Swift property) was fixed and covered by XCTest;
  the iPhone 16 local API UI suite now passes 17/17, native contracts 67/67,
  and the default UI suite passes with only its intentional provider/API skips.
- 2026-10-06 native Assistant readability: persisted assistant prose and
  one-shot AI-entry replies now render through the shared `MPMarkdown` base
  component. Headings, emphasis/code, lists and pipe tables are native SwiftUI
  content; model HTML/view code is never evaluated. The parser contract is
  covered in the 68/68 native unit suite.
- 2026-10-06 Finance Assistant grounding: every advisor run now receives a
  bounded/versioned product handbook before the owner conversation. It maps
  actual routes and feature boundaries, requires tool-grounded financial facts,
  and only permits a validated wallet-creation proposal followed by explicit
  confirmation through the owner-scoped wallet API; category, budget, jar and
  transaction writes remain unsupported.
- 2026-10-06 wallet form input: web wallet creation now uses the shared
  `AmountField` base for opening balance, savings target, credit limit and
  statement balance, preserving integer validation while adding formatted
  display and calculator input.
- 2026-10-06 notification slice: migration `000029_notifications` adds the
  persisted goal-wallet preference and an owner-scoped in-app inbox with
  idempotent mark-read. Web and native goal-wallet editors now use base controls;
  native More exposes the inbox. Push/APNs, scheduling and fan-out remain
  explicitly outside scope.

- 2026-09-21: owner requested feasibility assessment and a reviewable file for the two AI flows. [DESIGN-09-AI](work/tickets/TICKET-09-DETAIL_DESIGN.md) is `in_review`, approval pending: text/OCR proposals, internal-transfer reconciliation, read-only financial Q&A, prompt/eval contracts and T0–T7 delivery plan. No AI/transfer/jar runtime implemented in this documentation task.

- 2026-09-20 latest: owner requested API-only screens. All mounted data screens now use real APIs, including new budget explicit-interval CRUD/progress. Savings catalog filtering enforced FE/BE; category names/icons resolved in goal history. API roundtrip through FE proxy passed with PostgreSQL, test fixtures cleaned. Migration 000009 applied. [API-SCREENS-01](work/tickets/API-SCREENS-01-DETAIL_DESIGN.md) and savings remain in review/partial product scope; recurring budgets, internal paired transfers, reports and notifications are not completed. Earlier mock-budget/counterpart-pending notes below are historical.

- Latest owner request (2026-09-13): finish wallet management and adding transactions, then verify wallet–transaction–category logic and UI. Core wallet and basic income/expense slices are in review with [wallet proof](work/tickets/TICKET-01-02-VERIFICATION.md), [ledger proof](work/tickets/TICKET-02-01-VERIFICATION.md), and a durable [page contract](design/pages/transactions/README.md).

- Status: The root `app/` Vite React Tailwind app renders the Financial Clarity preview and is connected to the dev Gin API; Google OAuth and development CORS are enabled for local testing. PostgreSQL uses the explicit Stage v1 baseline migration; API startup does not mutate schema.
- Active backlog: `docs/work/BACKLOG.md`
- Current queue focus: transfer owner UAT and CORE-03 reconciliation; WORKER-01 remains draft pending report and recurrence semantics. AI-ENTRY-01/02 UAT, AI-ENTRY-03 owner design review, UI-FORMS-03 visual review, transfer pair edit/delete/idempotency and credit ledger remain separate open work.
- Active phase: None.
- Active ticket: [CORE-03](work/tickets/CORE-03-TIME-JARS-MONTH-DETAIL_DESIGN.md) is in progress through docs reconciliation and owner UAT. [AI-ENTRY-02](work/tickets/AI-ENTRY-02-BATCH-ATTACHMENTS-DETAIL_DESIGN.md) and [UI-FORMS-03](work/tickets/UI-FORMS-03-DETAIL_DESIGN.md) remain in progress/review for their independent UAT.
- Verified bug: [BUG-001](work/bugs/BUG-001-s3-presigned-get-signature.md), fixed with AWS SDK Go v2; [ADR-007](decisions/ADR-007-aws-s3-presigning.md) records the choice.

## Current Focus

Latest-turn override: implementation/reconciliation of [AI-ADVISOR-01](work/tickets/AI-ADVISOR-01-DETAIL_DESIGN.md). Earlier implementation queues below remain open; the local advisor slice is not a public deployment claim.

Owner requested continuing the approved timezone/jars/month goal. [CORE-03](work/tickets/CORE-03-TIME-JARS-MONTH-DETAIL_DESIGN.md) now has real API-backed screens and automated verification; its schema is included in the Stage v1 baseline. Next is owner UAT and reconciliation of remaining acceptance gaps. Monthly reports are live/recalculable, notes are independent, and automatic completion does not freeze ledger data. The new request for a cron worker is tracked separately in [WORKER-01](work/tickets/WORKER-01-cron-service.md); no implementation starts until month-end artifact semantics and recurrence rules are reviewed.

The independent UI-FORMS-03/AI flows still use the base-first contracts: `CategorySelectionList` is removed and transaction/budget/savings use `CategoryTreeSelector`; Money Lover guides AI input/result layout while MyPocket retains review/approve. Their visual/provider/browser UAT remains separate. AI entry is one-shot, stores no conversation, OCRs files before LLM, retains private S3 evidence and links it on approval.

## Recently Touched Areas

- `app/src/services/{accountTime,months,jars}.ts`, timezone context, jar/month API-backed screens and transaction/AI jar selectors; the Stage v1 baseline and owner-scoped calendar/jar/month repositories.
- `docs/work/tickets/CORE-03-TIME-JARS-MONTH-DETAIL_DESIGN.md`, `WORKER-01-cron-service.md`, API/ERD/ADR-008, report/business rules, validation matrix, backlog and changelog.
- `app/src/atomic/atoms/BaseCategoryTree.tsx`, `molecules/CategoryTreeSelector.tsx`, AI composer/result-card bases, transaction/budget form composition and browser fixture tests.
- `docs/design/system/BASE_COMPONENTS.md`, category-tree contract, assistant/transaction/budget screen specs, UI-FORMS-03.

- `backend/internal/controller/http/transaction_handler.go`, wallet ledger balance repository/model, handler tests and generated Swagger; `backend/internal/infrastructure/ai/` now enforces strict JSON Schema output and includes bounded wallet descriptions for model disambiguation.
- `app/src/services/transactions.ts`, transaction/wallet rule tests, real Overview/Header/Transactions/Quick Add/Wallet consumers.
- `docs/design/pages/transactions/`, wallet/transaction verification, API/backlog/validation/changelog reconciliation.

- `docs/work/tickets/`: 11 parent tickets, 32 children and a business-oriented index; newly captured provider-choice draft remains pending owner review.
- `docs/requirements/SPEC.md`, `BUSINESS_RULES.md`, `REPORTS.md`, `REQUIREMENTS.md`, `USER_STORIES.md`
- `docs/work/VALIDATION_MATRIX.md`, `docs/work/ROADMAP.md`, `docs/releases/CHANGELOG.md`
- Standards relocated unchanged from `docs/templates/standards/` to the mandatory `docs/standards/` path; unused Harness CLI phase example removed.
- `AGENTS.md`
- `docs/`
- `docs/templates/`
- `docs/standards/`
- `design/`
- `design/system/DESIGN.md`
- `app/`
- `app/src/atomic/`

## Recent Decisions

- CORE-03 month completion is account-local and derived; monthly figures stay recalculable and user notes remain independent. A cron close/snapshot would change this contract, so WORKER-01 is intake only until the owner chooses report artifact semantics.
- Basic income/expense entries may use `basic` and `goal` wallets. `credit` is intentionally rejected until CRD-01…04 purchase/refund/payment semantics have an approved ledger; applying ordinary expense direction would corrupt debt meaning.
- Category selection obeys both transaction kind and applicable-wallet scope at UI and API. Empty `wallet_ids` means all owner wallets.
- Wallet API list exposes ledger-derived `current_balance`; `opening_balance` remains unchanged by transaction CRUD. Header total uses current balances only for `is_in_total` wallets.
- Global add and transaction edit/delete now use real APIs. Header, Overview, Transactions and Wallet management share refresh propagation; browser UAT found and fixed two stale-balance branches.

- `docs/design` now retains Markdown specifications only; 7 PNG/8 HTML exports were extracted and removed by owner request, recoverable at Git commit `1bc013d`. Shared behavior and per-component constraints are in `docs/design/system/` and component READMEs. Write screen-specific behavior as each screen is implemented.
- UI primitives use the imported `app/src/ui/theme.css`; component variants use `ui/variants.ts`, with atom tokens only re-exporting. Build checks base ownership, literal colors and common visual overrides. See ADR-002. CTA uses brand, action uses emerald; current default card/control radius remains 12px.
- Shared text, controls, cards, progress/gauge, feedback, navigation and chart compositions were normalized. Keyboard segmented selection and bottom-sheet focus/close/scroll behavior are shared. Prototype data and unmounted incomplete keypad/report flows remain explicitly partial.

- Owner chốt tên bảng tài khoản là `user`. ERD dùng tên đích này; GORM mapping và rename migration chưa triển khai, phải giữ dữ liệu/ID tài khoản hiện có.

- Wallet review: xoá thực sau xác nhận đã chốt, cho phép trùng tên. Money Lover adjustment tạo giao dịch mới bằng chênh lệch, mặc định loại khỏi báo cáo; thiết kế ledger MyPocket cần review theo kết quả này.
- Nhóm mặc định theo catalog owner chốt ngày 2026-09-13; migration `000004` là nguồn runtime cho tên, system key và quan hệ cha–con.
- Schema startup now renames legacy `app_users` to `user`, migrates wallets/categories/category-wallet assignments, and idempotently seeds stable system groups.
- The two category-tree exports are consolidated under `docs/design/molecules/category-tree/`; `BaseCategoryTree` is the canonical atom-composed molecule, with `nested` and `line` named variants.
- Category APIs return `wallet_ids`; create/update validates every selected wallet belongs to the authenticated owner. Swagger is regenerated from handler annotations.
- Unified Gin responses with `{data, meta}`, RFC 9457-style errors, request IDs, and shared constants/helpers; profile route is `/api/v1/auth/profile`.
- Database schema and system category data are now owned by the single Stage v1 SQL baseline in `backend/migrations/000001_stage_v1.up.sql`, executed with `go run ./cmd/migrate up` in dev and prod. `AutoMigrate`/startup system seed were removed; a fresh disposable Stage v1 database applies at version 1 clean. Wallet deletion cascades its transaction rows by the approved permanent-delete policy.
- Gin route registration is split by public-auth, account, category, wallet and transaction domain. Swagger is code-first: annotations are beside Go handlers and `go generate ./cmd/api` regenerates only implemented endpoints.
- Group management uses owner scoping, valid kind/parent rules, max two levels, child re-parenting on personal-parent deletion and owner-wallet validation. System metadata is read-only, but an owner may choose its applicable wallets through a dedicated endpoint. `/account/groups` has real list/create/edit/delete states, not mock data.
- Added `/account/groups` UI using the real categories/wallets APIs and reusable bases (`BaseCategoryTree`, `Heading`, `FormField`, `BaseCheckbox`, `SurfaceCard`, `IconBadge`, `BaseBottomSheet`). The canonical tree follows `docs/design/molecules/category-tree/nested`; personal-category edit exposes confirmed deletion. Account links to wallets/groups are visible. Reports and quick-add are retained source but deliberately not mounted until their APIs exist.
- TICKET-01-03 (group management) is complete; CRUD group remains separately scoped from wallet management.
- The bottom-navigation create button is vertically aligned inside the shared navigation frame; its center slot remains intentionally unlabeled.
- Vite dev proxy now forwards `/api` to Gin, allowing FE and BE to be tested through one browser origin.
- Base-cards export tại `docs/design/atoms/base-cards/` là nguồn tham chiếu mới; Overview đã dùng `SurfaceCard`, `IconBadge`, `WalletCard` và `TransactionItem` để khớp các biến thể card/list, không lặp markup card tại màn hình.
- Variant maps của surface, badge, button, wallet và transaction được tập trung trong `app/src/atomic/atoms/tokens.ts`; consumer chỉ tham chiếu token global.
- `BaseButton` owns inline-flex icon-plus-label alignment; screen callers may use `gap` but must not recreate this layout locally.
- Base card/form/row corners are reduced one token step; circular icon, progress and pill CTA geometry remain semantic exceptions. In `BaseCategoryTree`, the trunk begins at the first child row rather than crossing the parent row.

- Divide the full business contract into parent/child tickets with concise scope and acceptance criteria. Keep unresolved decisions in the affected child; do not silently resolve them during breakdown.
- Current product contract is `docs/requirements/SPEC.md`, with `BUSINESS_RULES.md` and `REPORTS.md`; older design inputs do not override these decisions.
- Budget and Jar are separate. Jars are optional expense-group tracking with monthly copied configuration, advisory allocation/warnings, no balance carryover, and a cumulative reporting view.
- Monthly reports remain recalculable after month-end. User notes are independent; related context and AI summaries are generated separately and cannot overwrite manual notes.
- Timestamps use UTC; account timezone drives query boundaries; calendar-only dates/months keep their explicit labels. Only VND is seeded; multi-currency has no designed contract.
- Travel Mode automatically links eligible new transactions; recurring-generated transactions do not inherit it. Recurring creates ordinary editable transactions with a default note.
- Native recurring numeric drafts now preserve valid thousands grouping while rejecting malformed characters/separators before create/update; see [NATIVE-RECURRING-NUMERIC-STRICTNESS](work/bugs/NATIVE-RECURRING-NUMERIC-STRICTNESS.md).
- The following older bullets describe the existing preview/history, not authority over the current product specification.
- Keep `AGENTS.md` at the repository root for agent discovery.
- Keep shared state in `docs/CONTEXT.md`.
- Use markdown-only enforcement for v1.
- Use `docs/work/phases/` for multi-ticket work.
- Use pay-as-you-go documentation for brownfield projects.
- Treat `design/system/DESIGN.md` and `design/INDEX.md` as the current design/product inputs for the next app-start phase.
- `design/system/DESIGN.md` now defines Money Lover-style app parity, personal extension backlog, component inventory, data domains, launch phases, and acceptance checklist.
- Design artifact folders under `design/` are exported source references; base implementation follows the component inventory in `design/INDEX.md`.
- Frontend preview uses the old app as logic/data-shape reference only; UI is new and follows `docs/design/system/DESIGN.md`.
- New app source lives in root `app/`; runtime behavior and API contracts are documented in this repository.
- The standalone FE preview now includes Overview, Transactions, Budgets, Reports, Account, Quick Add sheet, goals/funds, quick personal actions, and a nested category report mock.
- OCR Platform docs have been captured in `docs/architecture/OCR_API.md` for future receipt OCR implementation against `https://ocr.dungxbuif.com/`.

## Next Steps

Native Travel Mode parity was restored on 2026-10-06: active context/date
scope, safe deletion confirmation, and the ordinary-income/expense-only
auto-link warning are now covered by the shared SwiftUI components. Native
XCTest 54/54, Release warnings-as-errors, base/project guards, and the signed
authenticated More-route sweep 1/1 in 122.031s pass for this correction.

Native recurring parity was then restored: account-timezone scheduling copy,
next-run/wallet/note/type details, and safe delete confirmation now match the
web contract. Native XCTest passes 55/55, the Release warnings-as-errors
build and guards pass, and the signed More-route sweep passes 1/1 in 124.207s.

Native Reports now also loads the optional month-summary breakdown, showing
category/type/count rows and the ledger activity exclusions beside Insider
metrics. Native XCTest remains 55/55, Release/guards pass, and the signed
More-route sweep passes 1/1 in 124.287s.

Native credit statement parity was then restored: current debt is shown as a
positive amount, payment status and credit entry kinds are localized, due-date
context is visible when configured, and row notes take precedence over dates.
The existing credit accounting/API contract is unchanged. Native XCTest
remains 55/55, the warnings-as-errors simulator build passes, and the signed
More-route sweep passes 1/1 in 125.111s.

Native portfolio detail parity was then restored: each asset now shows weighted
average cost, realized/unrealized P/L, current value or explicit unknown
pricing, and newest-first account-timezone buy/sell history with quantity,
price, fee and note context. Native XCTest passes 56/56, Release/guards pass,
and the signed More-route sweep passes 1/1 in 121.948s.

The complete signed `LOCAL_API_UI` simulator regression was rerun after the
credit/portfolio batch: 14/14 passed in 293.119s, including wallet, transfer,
credit entry/payment, Assistant, AI-entry capability, feedback and secondary
route coverage. Credit assertions now intentionally verify localized display
labels rather than raw API enum strings.

The following enum-label parity pass centralized the remaining iOS display
labels for wallet/category/credit/feedback/Assistant statuses. Native XCTest
passes 57/57, Release/guards pass, and the signed credit-wallet plus
feedback/more targeted UI checks pass 1/1 each. Wire enum values are unchanged.

The native feedback screenshot follow-up now matches the backend attachment
contract end to end: captures are PNG, named `screen.png`, sent as
`image/png`, and progressively downscaled to stay at or below the 1 MiB limit;
the existing text-only fallback remains available. The final iPhone 16
Simulator gate passes 65/65 native unit tests and 16/16 signed
`LOCAL_API_UI` tests, with Release warnings-as-errors, base/project guards,
no invalid-frame fault, and zero disposable fixture wallets remaining.

The feedback-bubble redaction follow-up now hides the floating trigger for one
rendered frame before capturing the feedback screenshot, preventing the
control from appearing in its own attachment. The same simulator gate remains
green after the UI regression; physical device, real Google OAuth,
provider-backed AI/OCR and authenticated staging/production remain external
release gates.

The latest native AI-entry attachment regression adds a DEBUG-only twenty-file
fixture: the simulator renders valid PNG thumbnails, disables selection at the
cap, and re-enables it after removing one file. The complete signed
`LOCAL_API_UI` suite passes 17/17 (82/82 combined with unit/contract tests),
with the fixture isolated from production behavior and the existing twenty-file
API contract preserved.

The follow-up base-component audit moved the last feature-level raw text/icon
nodes into `MPText`/`MPIcon` and extended the static guard to reject raw
`Text`/`Image` as well as native controls. A red accessibility regression from
decorative icon labels was fixed; the full 17/17 signed UI rerun, 65/65 unit
tests, Release warnings-as-errors build and simulator launch are green.

- Review the [dedicated advisor design](work/tickets/AI-ADVISOR-01-DETAIL_DESIGN.md) and [V1 task sequence](superpowers/plans/2026-09-22-finance-assistant-v1.md). Feedback's browser round-trip/owner UAT and deployment checks remain separate. Finance Assistant has a local JWT/API-key synchronous JSON vertical slice with best-effort Redis access audit; next gates are typed fact cards, SSE/recovery, durable audit delivery/alerting and configured-provider browser proof.
- Finance Assistant contract/lifecycle fixes are now locally verified: typed schemas, OpenAI-compatible tool history, bounded final-answer loop, credential revalidation, cancelled-run commit guard, and wallet/budget/goal/jar cards. Remaining gates are configured-provider browser proof, SSE/reconnect, durable audit delivery/alerting and owner UAT.
- Finance Assistant cancellation now interrupts provider work handled by the same backend process; the database lease guard remains the source of truth. Do not describe cross-process immediate cancellation as shipped until durable pub/sub or worker ownership is implemented.
- Finish CORE-03 owner UAT at the local frontend (`/account` timezone setting, `/jars`, current/past `/months/YYYY-MM`, transaction and AI proposal jar assignment). Confirm data behavior visually and with an account the owner is comfortable testing; automated evidence is in [VALIDATION_MATRIX](work/VALIDATION_MATRIX.md).
- Clarify [WORKER-01](work/tickets/WORKER-01-cron-service.md): whether “chốt sổ/tạo report” means a persisted month-end artifact or only a scheduled trigger/cache over recalculable figures, plus recurring catch-up/idempotency/timezone behavior. Then prepare detail design for review before code/runtime changes.
- Keep AI-ENTRY-01/02 live/provider/attachment UAT and UI-FORMS-03 visual review as separate backlog work; broader AI model evaluation remains required beyond the five synthetic cases.
- Preserve base-first screen rules and run `npm run check:design`, `npm run test:design`, and `npm run build` for frontend changes. API/DB/runtime changes require PostgreSQL tests and migration evidence.

Native simulator validation on 2026-10-06 also hardened the UI-test harness:
`-UITestAuthenticated` and `-UITestAuthenticatedAPI` never reuse Keychain
credentials, and the More-route sweep reuses one authenticated process instead
of repeatedly terminating/relaunching it. The iPhone 16 simulator passed 65/65
native contracts, 8 default UI smoke tests, and 9 isolated local API-backed UI
flows. A multi-test `LOCAL_API_UI` xcodebuild operation still triggers an Xcode
27 runner restart; treat the isolated per-test command list as the current gate.

Production feedback #62e31193 (`AI respone`) was reproduced in the web Finance
Assistant: assistant text was passed directly to the base `Text` component, so
Markdown tables and headings were visible as raw characters. The web now uses
`FinanceAssistantMarkdown` with no HTML injection; the regression contract is
`npm run test:ai-advisor` and the screen spec records the structured rendering
behavior.

Production feedback #e11d18e6 (`Requirements`) was reproduced in the wallet
management flow: selecting a wallet opened a separate detail screen instead of
scoping the transaction ledger. The web now routes the shared wallet selection
to `/transactions?wallet_id=<id>`; the transactions page consumes the optional
query and resets to the current week. The route contract is covered by
`npm run test:wallet-scope` and the transaction page design contract.

Native ledger visual capture now proves ordinary transaction rows remain
reachable above the persistent shell after scrolling; evidence is recorded in
`docs/work/VALIDATION_MATRIX.md` and the ledger capture plan. Secondary native
routes now share `MPBottomNavigationInset` for Categories, Portfolio, Credit
and Notifications, and the secondary host hides SwiftUI's duplicate default
back button so the custom Stitch header remains singular. Local red/green
evidence is recorded in the validation matrix; device and production visual
UAT remain open.

The dedicated Notifications shell smoke is green, and the seeded inbox test now
proves unread state → native mark-read tap → persisted `read_at` through the
owner-scoped API. No public create route was added; test data is inserted and
removed only in the disposable local stage DB. Pixel-level owner/device UAT
remains open.

The follow-up shell audit found and fixed missing bottom insets on the
scrollable Finance Assistant, AI Entry, Feedback and Account routes. A red
focused regression reproduced the gap and the green route check now passes for
all four using the shared base; no screen-local spacer or API change was added.

The same geometry audit fixed the Wallets list/editor branch: the shared inset
now keeps the wallet save footer above the persistent shell. The focused
simulator contract is recorded in the validation matrix.

Transaction overflow forms for chuyển tiền and điều chỉnh số dư now also use
the shared shell inset; a focused simulator smoke opens both forms and passes.

The AI entry proposal editor overlay now also owns a scoped
`MPBottomNavigationInset`; a RED/green focused test proved that the editor's
save action stays above the persistent shell. The screenshot is deterministic
local fixture evidence only; live OCR/provider and device visual UAT remain
open.

The wallet transaction-detail overlay now follows the same scoped shell
contract. An API-backed mutation flow reproduced the missing inset, then passed
after `TransactionDetailView` composed the shared base spacer; the captured
frame is local fixture evidence, not production data.

2026-10-08 native secondary-route visual checkpoint: all 15 More destinations
were captured after navigation against the isolated local API on iPhone 16
(1/1 UI test; empty stage account). No network-error placeholders appeared.
The Account screenshot revealed the global feedback bubble covering the end of
the Face ID explanation; its text now wraps clear of the button and the focused
API-backed simulator assertion passes.
Populated-state comparison, Stitch/web pixel review, and owner device acceptance
remain open, so the 100% native port is not signed off.

2026-10-08 OCR gate recheck: earlier notes claimed credentials were missing,
but the ignored canonical backend environment has OCR, private S3 and model
credentials. The opt-in synthetic evaluation passes S3 private PUT/readback and
cleanup, OCR recognition of 45,000 VND, and 24/24 expected DeepSeek fields over
five cases. It created no proposals or ledger records. iOS attachment picker,
approval/download and owner device UAT remain open; see the validation matrix.

## Open Questions

- 2026-10-13 native API-key copy feedback: Account's one-time secret card now
  reuses shared MP bases and offers explicit local-only clipboard copy with a
  five-minute expiry and visible/VoiceOver confirmation. Native suite 150/150
  and API-backed UI regression 1/1 pass. Physical iPhone install is blocked by
  missing Xcode team sign-in/provisioning profile; no install was attempted.
  Owner device review and release-note approval remain open. See the API-key
  copy detail design and validation matrix.

- Native category-list parity is in progress: the management screen now has
  kind/wallet/search filters and a shared tree over the existing API contract;
  transaction/budget/recurring pickers also use a searchable tree and their
  selected IDs were verified in API-backed UI, including the generated ledger
  transaction. Web list wallet/search remain. Inactive-category semantics still
  need a product/API decision and must not be inferred from local usage. No
  backend migration was necessary for these flows.

- WORKER-01 month-end semantics: CORE-03 says live/recalculable report and derived completion; the new request names a cron job that closes/generates a report. Decide whether any stored artifact is authoritative or merely rebuildable/cache output. Recurring month-end/catch-up/duplicate prevention/timezone changes also need rules.
- Deletion of linked objects and paired-wallet effects; Travel Mode for backdated/offline/delayed confirmation.
- Exact income/refund classification for jar allocation; credit overpayment/refund/statement allocation; portfolio funding and historical deletion behavior.
