# Native completion reconciliation

Status: in_progress. Owner: “Đối soát và làm hết đi”. This is the current
cross-ticket index; older checkpoint paragraphs are historical evidence, not
current completion claims. [Full goal](NATIVE-FULL-COMPLETION-20261014.md).

## Before changes

Reconcile internal backlog/tickets/validation with code and actual artifacts;
correct public/agent guides only where behavior changed or a claim is stale.
Do not mark feedback resolved, publish release notes, delete production data or
install a queue-enabled client against an incompatible production backend.
Continue approved tests/fixes; preserve API/owner/auth/financial semantics unless
the applicable reviewed design explicitly changes them. No archived web work.

## Authoritative matrix

Fresh production queue contains48open records, not34/42. [Latest six assessment](NATIVE-FEEDBACK-NEW-20261015-DETAIL_DESIGN.md)
adds actual portfolio edit/delete and shared Back/spacing/fixed Feedback/Account
Add and AI jar-tool requirements. No production feedback status changed.

| Area | Local evidence/current implementation | Production/device gap | Next gate |
| --- | --- | --- | --- |
| Shared bases, compact rows, photos-first | UI plan Tasks2/4 locally verified; numeric/chat287/287PASS | full-screen/owner UAT incomplete | populated registry/visual sweep, independent review |
| Dedicated transaction note editor | TransactionsView note sheet; note RED→GREEN2/2 and API note readback in validation | owner UAT only, not missing implementation | correct stale feedback ticket, retain API/current CRUD regression |
| Remembered-device Face ID | protected device credential flow and local API/UI proof | auth backend already deployed12572c87; physical biometrics UAT pending | owner device acceptance, not auth reimplementation |
| API-key copy/revoke | actual local account key create/copy/revoke/reopen proof | updated account/full-owner policy not fully rolled out | full-owner rollout + owner UAT |
| Full-owner API keys | common RequireAuth, no scope gates, realPG/JWT/Redis security tests | old auth-only API image still live | scoped immutable backend release, isolation/security gates |
| Offline receipts/outbox/engine/foreground | encrypted store, acknowledgements, retry/lease/restart/fault E2E implemented locally | fresh production capability404; public OpenAPI lacks route | deploy matching backend BEFORE another new native install |
| Financial overview/report cache | encrypted owner-scoped snapshot now includes optional server-returned MonthSummary/Insider payloads; Overview and exact-month Reports transient fallback show cached timestamp; invalid-month stale-content bug fixed | PostgreSQL-backed UI fixture and device UAT unavailable; full native unit remains green | real API/UI offline read regression when fixture is available |
| Savings rejected draft |503same-key/pair/balance/report proofPASS; transfer edit eligibility and the approved fault-injected 422 savings UI flow are green locally | production capability/API rollout and device UAT remain open | keep the local fixture command repeatable; re-run only against the compatible backend candidate before release |
| Critical business regression | isolated category280/280PASS; final12business+280unit **292/292PASS**,0skip,0 warnings (latest additive dispatcher-context assertion included) | local gate is green; physical/device, prod compatibility and lifecycle gates remain | retain test-only corrections, then continue non-environment product gates |
| AI numeric/privacy | exact amounts, no invented zero, exact count/JSON, masked fallback;287/287PASS | updated Release pending signing, phone unavailable | signed compatible-backend install + owner UAT |
| AI cancellation | approved conversation→run recheck, realPG4/4 + Go racePASS | patch not deployed; initial Stop unavailable until synchronousPOST returnsID | backend rollout, initial transport/action design |
| AI confirmed actions/help | wallet, jar create/update and budget create/update proposal parser/history local proof; bounded `get_product_help` retrieval and native/web `product_help` cards are locally verified; persisted proposals now use the owner/generation-bound confirmation dispatcher, reject stale/tampered requests, and replay through the mutation ledger; legacy cards retain fallback | live grounded help/RAG, real-provider proof and compatible backend rollout remain incomplete | deploy matching backend, then real-provider eval and owner UAT |
| Idle chat expiry |30min access purge, worker cleanup, 410 mapping, idempotent Chat mới reset, in-process cancellation of purged provider runs, and additive history metadata implemented locally | real-DB proof, production rollout and owner UAT open | compatible backend rollout plus expiry/reset E2E and owner UAT |
| Feedback placement/completeness | bubble/no reserved lane, owner resolve/reopen support, and stale refresh/PATCH snapshot protection are locally verified | production OpenAPI lacks owner-resolution route/migration32; protected-region overlap and new fixed-anchor request open | scoped backend rollout plus reviewed fixed-anchor/UI proof |
| Notifications | inbox/read API and native screen exist | no APNs provider/delivery proof; current Personal Team lacks entitlement | supported signing/delivery decision, not inbox-only completion |
| Public documentation | human/agent Markdown and generated Swagger maintained locally; backend now redirects `/docs` and `/docs/` to `/api/v1/docs/index.html` with a regression test | deployed versioned Swagger UI is reachable at `/api/v1/docs/index.html`, but legacy `/docs/` redirect is still 404; OpenAPI/agent-surface rollout remains unverified | deploy matching redirect and recheck public guide/agent surface without restarting archived web |
| Release | unsigned Debug/Release device builds compile locally; no final signed install | provisioning profile missing; no final whole-branch review/UAT | compatible backend, strict signing, install/launch, owner acceptance |

Fresh production read-only evidence: `/api/v1/sync/capabilities`404; downloaded
public OpenAPI lists device/auth and advisor/feedback/key routes but no sync
capability. Known vm100 SSH identity works without agent; Docker reports API
auth-12572c87 healthy and workerrelease-73c4c37. vm100 ControlAvailable=false;
cannot update Swarm there or promote it. Manager pi rejected the explicit known
homelab identity; normal pi uses a1Password-backed public identity requiring
agent authorization. Do not chmod a public key or alter SSH/Keychain policy.

Fresh unauthenticated probe on 2026-10-15: `GET /api/v1/sync/capabilities`
returned 404; `GET /api/v1/docs/index.html` returned 200 and served Swagger UI;
the legacy `GET /docs/` redirect still returned 404. This confirms the deployed
image is ahead of the old docs-page snapshot but still lacks the compatibility
redirect and offline capability route. No authenticated data or production state
was changed.

Read-only recheck at 2026-10-15 20:48 +0700 returned the same contract: sync
capabilities 404, versioned Swagger UI/doc.json 200, legacy `/docs/` 404 and
`/api/v1/openapi.json` 404. The deployed Swagger document still exposes 72
paths versus 77 in the local generated document. No rollout or production
mutation was attempted.

Device/signing recheck at 2026-10-15 20:50 +0700: the connected iPhone
`DungxBuif` is now online, but the local provisioning-profile directory is
empty. The Apple Development identity exists; a signed device build therefore
still cannot start until a profile for `com.mypocket.ios` is available. No
device install or provisioning update was attempted.

Follow-up signing probe at 2026-10-15 21:09 +0700 is unchanged: the same
development identity is valid, the provisioning-profile directory remains
empty, `DungxBuif` is online and the second iPhone remains offline. The
unsigned arm64 device build is green; no signed build/install was attempted.

Read-only vm100 probe at 2026-10-15: the node is a Swarm worker, so service
listing is manager-only. Its running task containers are nevertheless
authoritative for the current image gap: `mypocket-api` is
`registry.dungxbuif.com/mypocket-api:auth-12572c87` at
`sha256:1e1c686ce90bfd7c333dcc7f9206aff23117ed9f9d95839c8612a1a0fbc26fce`,
and `mypocket-worker` is `release-73c4c37` at
`sha256:d1f3f3b1bba9648192feccb814bd0823b9bc7df2c80518bf29a89cf77df691c2`.
Both were up/healthy for roughly 21 hours. No manager promotion, service
update, migration or registry push was attempted.

The approved savings 422 rerun is now green against a disposable local stack.
PostgreSQL `mypocket_feedback_local` was migrated to version 34 on
`127.0.0.1:5432`; the API ran on `127.0.0.1:8081` and the native app was
explicitly pointed at the fault proxy `127.0.0.1:18081`. The proxy rejected the
first transfer with 422, the app persisted an editable rejected draft without
inflating confirmed savings/report totals, restart restored the draft, editing
to 22,500 đ and switching the proxy to pass committed exactly one transfer pair,
reloaded the goal balance (122,500 đ) and remaining target (877,500 đ), and
left the transfer out of report aggregates. Evidence:
`/tmp/mypocket-offline-savings-20261015e/Logs/Test/Test-MyPocketIOS-2026.10.15_20-37-31-+0700.xcresult`.
The preceding failed run was a harness wiring error: the app used API port
8081 directly and therefore received 201 instead of the injected 422; it was
not used as product evidence.

Mac currently locked (native app-control tool confirms manual unlock required),
so1Password/manager and Keychain approvals cannot be completed by the agent.
Do not automatically retry auth/prompts or report a deployment from worker access.

Compatibility activation must preserve the encrypted owner outbox: queued intents
may drain on the owner's next foreground after capability becomes available.
Never clear/rekey these pending requests or send legacy writes as a fallback.
Validate receipt replay/unchanged existing financial rows before activating the
new client; synthetic rollout proof must not use the real owner's finances.

## Execution order

Follow-up local verification on 2026-10-15 remains green: backend `go test ./...`
and `go vet ./...` passed; base-component and project guards passed; the full
native unit target executed 292 tests with 0 failures at
`/tmp/mypocket-reconcile-unit-final-20261015.xcresult`; the focused DesignSystem
target executed 89 tests with 0 failures at
`/tmp/mypocket-designsystem2-20261015.xcresult`; and a Release build
with `SWIFT_TREAT_WARNINGS_AS_ERRORS=YES CODE_SIGNING_ALLOWED=NO` passed at
`/tmp/mypocket-reconcile-release-nosign-20261015.xcresult`. The signed Release
invocation stopped before compilation because this checkout has no selected
development team, so it was not treated as a code failure. No signing, device
install, deployment, migration, or production data mutation was performed by
this verification pass.

### Final broad test checkpoint

The latest full native run ended with 292 total: 292 PASS/0 skip/0 runtime
warnings. Artifact `/tmp/mypocket-reconcile-unit-final-20261015.xcresult`.
The latest DesignSystem target is 89/89; backend and web checks remain green.

The three failures from session18287 were test-infrastructure defects, not
production financial logic: the category icon was below the keyboard-compressed
viewport, credit controls had shell+nested ScrollView ancestors, and the
recurring fixture did not select its newly-created wallet. The correction dismisses
the keyboard at the category boundary, chooses the unique smallest visible
containing ScrollView (ties still fail), and selects/asserts the recurring wallet.
Exact API/readback/ledger assertions were retained. Focused proofs passed:
category280/280 (`/tmp/mypocket-category-fix-redgreen2-20261015.xcresult`) and
credit+recurring2/2 (`/tmp/mypocket-credit-recurring-fix-20261015.xcresult`).

No deployment, device install, production mutation or environment change follows
this local gate. Remaining product gates are tracked below and are not silently
closed by the test pass.
Public human/agent API docs need no test-only change; standalone public docs
publishing remains an independent open rollout requirement.

1. Reconcile stale ticket claims and finish bounded category proof.
2. Establish compatible scoped backend candidate/access; backup/test/rollout
   approved additive changes without importing unrelated dirty code.
3. Financial report/cache correctness and savings422 recovery; preserve server
   source of truth and visible failed/stale scopes.
4. AI idle lifecycle, cancellation/action replay/help; no unlimited raw history
   or unconfirmed financial mutation.
5. All-screen/feedback/physical tests, independent review, compatible install.
   Owner acceptance precedes feedback resolution/release publication.

Links: [backlog](../BACKLOG.md), [validation](../VALIDATION_MATRIX.md),
[UI plan](../../superpowers/plans/2026-10-14-native-ui-system-consolidation.md),
[offline plan](../../superpowers/plans/2026-10-06-offline-financial-sync.md).
