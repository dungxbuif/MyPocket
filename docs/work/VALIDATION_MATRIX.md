---
artifact_type: validation_matrix
id: VALIDATION_MATRIX
status: active
owner: shared
human_fields:
  - proof_override
  - acceptance_signoff
ai_fields:
  - proof_recommendation
  - evidence_links
  - status_updates
shared_fields:
  - matrix_rows
  - validation_status
updated: 2026-10-16
---

## 2026-10-16 Savings/Budget implementation — local candidate

Scope: [delegated design](../superpowers/specs/2026-10-16-savings-budget-design.md)
and [implementation plan](../superpowers/plans/2026-10-16-savings-budget-implementation.md).

| Gate | Evidence | Result |
| --- | --- | --- |
| Fresh additive schema | Disposable PostgreSQL 17 at localhost:55436, migrations through035 | PASS; no production schema change |
| Backend with real PostgreSQL | `TEST_DATABASE_URL=<disposable> go test -race ./... -count=1` | PASS; Redis-dependent tests still need separate fixture |
| Savings/report API integration | `TestSavingsBudgetReportsRealLedgerAndOwnerKey` through real PostgreSQL and owner API-key middleware | PASS; 20M−12M=8M surplus,5M goal transfer separate, budget/report pages agree |
| Recurrence persistence | Repository+HTTP race tests,8-way generation/mutation replay, revision/overlap/deletion cases | PASS |
| Review regressions | Wallet baseline creation boundaries; superseded scope deletion; selected AI budget carry chain | PASS |
| Signed native unit | `/tmp/mypocket-savings-native-20261016/Logs/Test/Test-MyPocketIOS-2026.10.16_21-24-27-+0700.xcresult` | PASS 317/317 after review follow-ups |
| Native Release build | `/tmp/mypocket-savings-native-release.log` | BUILD SUCCEEDED; warnings treated as errors |
| API-backed UI | [live UI evidence](tickets/SAVINGS-BUDGET-UI-EVIDENCE-20261016.md), `/tmp/mypocket-savings-budget-ui-fifth-20261016.xcresult` | PASS 2/2; real server stop confirmed before teardown |
| Device automation | [separate signed helper](../operations/IOS-AUTOMATION-INSTALL-20261016.md), physical MCP UI read | PASS; not product acceptance |
| Production/new iPhone candidate | New routes/schema/native install | Pending |

Personal Team Release artifact also builds successfully with a documented
developer-only Universal Links entitlement override. `codesign --verify --deep
--strict` passes; application identifier remains `com.mypocket.ios`, team
`AU442S3MN3`. It has not been installed over the owner's existing app.
[Personal build instructions](../operations/IOS-PERSONAL-BUILD.md).

Public human+agent references: [savings](../guides/savings-report-api.md),
[recurrence](../guides/budget-recurrence-api.md), [forecast](../guides/budget-report-api.md),
and generated Swagger. Scheduled push/goal completion forecast remain separate.

## Earlier 2026-10-16 savings and budget scope audit (superseded above)

The existing contracts are now covered by focused backend tests for account-
local inclusive budget date boundaries, one explicit `start_date`/`end_date`
interval, authenticated owner propagation through notification list/read, and
goal-only notification preference persistence. Existing tests continue to
cover goal progress/category rules, single-wallet savings entries, and paired
managed-wallet transfer exclusion.

The following remain **partial**, with no new behavior inferred: recurring
budget generation, budget/goal forecast fields, producer-driven budget/goal
notifications or push delivery, and a dedicated savings report/drill-down
endpoint. The inbox routes are owner-scoped read/mark-read only.

Evidence command: `rtk go test ./internal/entity ./internal/controller/http ./internal/infrastructure/repository` in `backend` — 142 passed, 33 skipped.
The full Go suite and `git diff --check` remain the final verification gate for
this slice.

## 2026-10-16 offline/sync contract verification

The focused backend command
`go test ./internal/controller/http -run 'Test(SyncCapabilities|Offline|FullAPIKey)' -count=1`
passes in the current worktree. Its PostgreSQL/Redis-backed cases are skipped
when `TEST_DATABASE_URL` or `TEST_REDIS_URL` is not supplied; this run therefore
proves package compilation and test discovery only, not real database receipt
persistence or Redis session expiry. The selected contract tests cover
capability enablement/authentication, receipt metadata and replay, changed-body
idempotency conflicts, concurrent transfer retry, explicit travel-context
replay, and JWT/API-key owner isolation. The transfer regression also asserts
exactly one durable receipt for concurrent retries.

No owner credential was available for an authenticated production capability
probe, and no signed physical-device XCTest was run in this backend-only slice.
Production capability payload/UAT and physical-device acceptance remain
pending; the client must stay fail-closed until the authenticated response is
verified.

## 2026-10-15 offline queue/cache regression recheck

The focused simulator contract run passed 123/123 tests with 0 failures on
`MyPocket-iPhone-16` (FinancialMutationQueue, SyncEngine,
FinancialSnapshotCache and DesignSystem targets). Artifact:
`/tmp/mypocket-goal-contracts-20261015/Logs/Test/Test-MyPocketIOS-2026.10.15_22-01-14-+0700.xcresult`.
This verifies the local fail-closed capability gate, FIFO/retry/conflict/auth
behavior and encrypted owner-scoped snapshot fallback. It does not prove the
production capability response is enabled; that still requires an
owner-authenticated probe.

## 2026-10-15 production API/worker rollout recheck

Both `mypocket-api` and `mypocket-worker` now run immutable image
`registry.dungxbuif.com/mypocket-api:ai-jar-capability-20261015-215712`
(`sha256:20fee27c5da418ecb23709484e9d8a844cbf1e9471592df74415ca0e831d6da0`).
The public health endpoint returned 200. Unauthenticated probes correctly
returned 401 for sync capabilities, advisor reset/confirm and feedback
resolution, proving the routes are mounted behind auth rather than missing.
Production Swagger JSON is 200 and has 77 paths, exactly matching local
generated Swagger (77/77). This proves rollout and public contract parity, but
not owner-authenticated capability `enabled:true` or provider-backed UAT.

## 2026-10-15 native route/auth gate

Added `AppRouteParser` and `AppRouteCoordinator` for scheme/universal-link
navigation. The focused simulator suite passed 10/10 (8 parser cases, including
all secondary scenes, and 2 pending-route/auth-phase cases) at
`/tmp/mypocket-route-final3-20261015/Logs/Test/Test-MyPocketIOS-2026.10.15_21-43-03-+0700.xcresult`.
The build registers the `mypocket` URL scheme and includes the associated-domain
entitlement. Production universal links still require hosting the matching
`apple-app-site-association` file and a signed device build.

## 2026-10-15 auth-family expiry hardening

Refresh records now carry an absolute family deadline. Rotation never extends
that deadline, and logout keeps the revocation tombstone through it with a
small clock-skew cushion. Legacy records migrate to a bounded deadline without
changing their stable family ID. Focused refresh/logout/TTL interleaving tests
pass 6/6; the full backend suite passes 295 tests in 16 packages, including
`go test -race ./...` (295/295) and `go vet ./...`. This is local code and test
evidence only; production rollout and migration execution remain open.
Linux/amd64 static artifacts built from this source are staged locally:
`/tmp/mypocket-api-linux-amd64-authfamily-20261015` (SHA-256
`ed96fa001da3f2a2612bda8be826141cad10097031b886dc3941e23f9678c108`) and
`/tmp/mypocket-worker-linux-amd64-authfamily-20261015` (SHA-256
`c2288fc0f956eb033b0193dafc68a4cbb93cf9348905a769405b42e2bc9773c9`).

## 2026-10-15 continuation verification

Latest re-run against the current worktree: `go test ./...` passed 293 tests in
16 packages, `go vet ./...` passed, and the web design/type/build gate passed
(7 design suites, all base contracts, TypeScript and Vite production build).
The Vite output retains only the known chunk-size advisory; no functional
failure was observed.
The same backend suite under `go test -race ./...` also passed 293 tests in 16
packages, with no race reports.
The focused HTTP contract suite under the race detector passed 22 tests,
covering sync capabilities, full-owner API-key authorization, feedback
resolution, advisor lifecycle/cancellation and device/offline auth contracts.
Configuration-dependent route contracts also passed 15 focused race tests for
Google exchange classification, Travel Mode, Money Insider and Credit wallet
validation. Missing external provider/configuration remains an intentional
runtime gate, not an unhandled panic or silent financial write.
The full native scheme also passed on `MyPocket-iPhone-16` / iOS 18.5:
330 passed, 64 documented skips, 0 failures, 0 runtime warnings, total 394;
artifact `/tmp/mypocket-native-current-20261015/Logs/Test/Test-MyPocketIOS-2026.10.15_20-51-59-+0700.xcresult`.
The focused offline capability guard also passed 3/3: disabled/unknown/malformed
capabilities are rejected fail-closed, known advertised operations are the only
ones enabled, and an unsupported capability preserves the pending intent without
leasing or sending a financial write. Artifact:
`/tmp/mypocket-offline-capability-20261015/Logs/Test/Test-MyPocketIOS-2026.10.15_21-04-22-+0700.xcresult`.
An unsigned physical-device build also completed successfully for the online
connected iPhone (`CODE_SIGNING_ALLOWED=NO`), producing an arm64 app at
`/tmp/mypocket-ios-device-nosign-20261015/Build/Products/Debug-iphoneos/MyPocketIOS.app`.
This isolates source/build correctness from the remaining provisioning-profile
blocker; the app was not installed.
The stricter Release device build also passed with
`SWIFT_TREAT_WARNINGS_AS_ERRORS=YES` and signing disabled; its app is at
`/tmp/mypocket-ios-device-release-nosign-20261015/Build/Products/Release-iphoneos/MyPocketIOS.app`.

Current-worktree checks: backend `go test ./...` passed 293 tests across 16
packages and `go vet ./...` reported no issues. Web design contracts and
production build passed (`npm run test:design`, `npm run build`); Vite emitted
only the existing large-chunk advisory. The focused native feedback
refresh/resolution race test passed 1/1 after a stale delayed GET regression was
made deterministic at
`/tmp/mypocket-feedback-race-green-20261015/Logs/Test/Test-MyPocketIOS-2026.10.15_21-14-02-+0700.xcresult`.
The subsequent full native unit rerun was not accepted as a release gate:
four unrelated Keychain/secure-outbox isolation tests failed in the simulator
(`testConcurrentKeychainRefreshCandidatesHaveOnlyOneWinner`,
`testSessionStoresAtomicallyRefuseStaleRefreshAndClearAcrossInstances`,
`testKeychainOutboxKeyIsDeviceOnlyStableAndNotRecreatedWhenMissing`, and
`testMissingKeyWithExistingRowsDoesNotGenerateReplacementOrLoseIntent`). The
feedback race test remains green; the Keychain failures require a separate
isolated simulator/keychain reset investigation.

Production recheck after the AI jar rollout now finds the sync capability route
present (unauthenticated `401`, not `404`), versioned Swagger UI 200, `/docs/`
redirecting to the versioned UI, and public `doc.json` with 77 paths. The public
Swagger path set matches local generated Swagger 77/77, including advisor
confirmation/reset, category status, feedback resolution and sync capability
routes. The capability body is intentionally not claimed here because this
read-only probe has no owner credential; native offline sending must remain
disabled until an authenticated owner response proves `enabled:true` and the
expected operations. No production data, database or device was changed by
this read-only verification.

The opt-in configured-provider overlap smoke also passed 1/1 using synthetic
Vietnamese transaction text only:
`/tmp/mypocket-live-provider-overlap-20261015.log`. It preserved three
distinct drafts, collapsed the repeated reference, retained separate dates,
and prefilled the sole wallet; it performed no database or ledger writes.
The configured OpenAI-compatible provider contract probes then passed 3/3:
plain extraction (one synthetic expense), one forced tool call with a valid
schema, and five JSON-schema cases covering expense, multiple dates, income,
transfer safety and missing-wallet review defaults. Evidence is
`/tmp/mypocket-live-provider-contract-20261015.log`; provider logs reported
`DeepSeek-V4-Flash`, valid tool-call schema and all five JSON cases passing.
These probes use synthetic data and do not persist financial rows.
Live product-help/RAG grounding and OCR/PDF remain separate gates because the
local `OCR_API_KEY` is empty and no live RAG evaluation fixture is configured.
An env-file audit found only `backend/.env.example` and `backend/.env.local`;
both declare `OCR_API_KEY`, but it is empty in each. No additional project env
file contains an OCR credential, so the OCR gate cannot be promoted by local
configuration discovery.
The next opt-in probe will exercise the real provider with the existing
`get_product_help` tool boundary and a synthetic handbook result; it must not
be interpreted as production RAG or owner-data evidence.
That probe has now passed: the model selected `get_product_help`, accepted
the bounded synthetic handbook result and returned a non-empty Vietnamese
answer mentioning the requested Trợ lý surface. Evidence:
`/tmp/mypocket-live-rag-grounding-20261015.log` (no DB/ledger access).
This validates provider/tool-result grounding only; it is not live document
retrieval or production UAT.

Signing check: an iOS device build with the installed Apple Development
identity/team (`GYRA8D4M54`) still fails before compilation because no
`com.mypocket.ios` provisioning profile is present; the connected iPhones are
also currently offline. `-allowProvisioningUpdates` was not used, so no Apple
Developer account or device state was changed.

## 2026-10-15 feedback dock right-side follow-up

Owner feedback reported that the lower-left feedback bubble made the simulator
content feel shifted left. The shared layout now docks the 52pt bubble at the
lower-right edge of the shell viewport. It remains an overlay: no obstacle
search, reserved lane, or content-width change was introduced.

Evidence:

- signed simulator focused run: 1 geometry unit test and 3 feedback UI tests,
  0 failures, artifact at
  `/tmp/mypocket-feedback-right-dock-20261015/Logs/Test/Test-MyPocketIOS-2026.10.15_20-20-29-+0700.xcresult`;
- Account security, More/category action, full viewport, and no-scroll-lane
  checks all pass;
- the signed full native regression immediately before this right-dock change
  passed 294 native unit tests and 100 UI tests (64 documented skips, 0
  failures) at
  `/tmp/mypocket-full-native-account-add-disabled-20261015/Logs/Test/Test-MyPocketIOS-2026.10.15_20-06-42-+0700.xcresult`.

This closes the local right-side placement slice only. Physical-device visual
acceptance, provider/production checks, and the broader native backlog remain
open; no environment was changed.

## 2026-10-15 Assistant numeric/card safety follow-up

The previously diagnosed unsafe monetary/progress/count conversions are covered
by the current bounded helpers and masking fallback. A signed simulator run of
seven focused `DesignSystemTests` cases passed with 0 failures at
`/tmp/mypocket-numeric-safety-20261015/Logs/Test/Test-MyPocketIOS-2026.10.15_20-23-53-+0700.xcresult`.
This closes the local presentation regression only; signed physical-device/UAT,
provider-backed responses and whole-release acceptance remain separate gates.

## 2026-10-15 cross-surface regression recheck

The current dirty worktree was rechecked without changing runtime state:
backend `go test ./...` and `go vet ./...` passed; web design guards, typecheck,
production Vite build, transaction/calendar/jar, AI-entry and feedback contract
tests all passed. The Vite build emitted only the existing bundle-size advisory;
no functional failure or deployment was inferred from this local run.

Production read-only probe after the API and worker rollout: `/api/v1/docs/doc.json`
and `/api/v1/docs/index.html` return 200, `/docs/` redirects (302), and
`/api/v1/sync/capabilities` is route-present (401 without credentials, not
404). Production Swagger has 77 paths, matching local generated Swagger 77/77.
The sync capability body remains unverified because this probe has no owner
credential; native offline sending must remain disabled until an authenticated
owner response proves `enabled:true` and the expected operations. No production
data or mutation was made.

## 2026-10-15 disabled Account add affordance follow-up

The compact shell previously instantiated `MPBottomAddButton` on Account and
Assistant, then hid it with opacity/disabled state. The resulting raised
circle could remain visible as a dim overlay over the final content row. The
base now conditionally renders the button only when `isAddEnabled` is true;
Account-specific add actions remain owned by the Account screen.

Evidence:

- `mobile-ios/scripts/check-base-components.sh`: PASS
- `mobile-ios/scripts/check-base-components.test.sh`: PASS
- `mobile-ios/scripts/verify-project.sh`: PASS
- `rtk proxy git diff --check`: PASS
- Signed simulator focused run: 90 `DesignSystemTests` plus
  `LaunchTests.testGlobalAddLongPressOpensAIEntry`, 0 failures, artifact at
  `/tmp/mypocket-account-add-disabled-20261015/Logs/Test/Test-MyPocketIOS-2026.10.15_20-05-54-+0700.xcresult`.
  The UI test confirms the global add button is absent on Account and still
  opens the transaction form from Transactions.

This closes the local base-component regression only. Physical-device visual
acceptance, production/provider checks, and the broader native backlog remain
open and were not changed by this slice.

## 2026-10-15 prepared jar update action

The red contract tests first failed because `propose_update_jar` was absent
from the registry/dispatcher and the native decoder only accepted create
actions. The local implementation now validates a UUID jar ID, month, name and
mode-specific allocation fields without writing; `jar_update_proposal` maps to
the existing confirmation card; and confirmation dispatches to the
owner-scoped jar PUT handler. Focused backend tests pass, including persisted
proposal binding and owner/generation checks. Native `DesignSystemTests` pass
90/90 at `/tmp/mypocket-designsystem-jar-update-final-20261015.xcresult` with no
runtime warnings. Production, device UAT and full provider-backed acceptance
remain separate gates.

Full native signed-simulator regression was rerun after the jar update card
change: 294/294 tests passed with 0 failures/skips and no runtime warnings at
`/tmp/mypocket-full-native-signed-20261015.xcresult`. The earlier unsigned
run's four `-34018`/`keyUnavailable` failures were simulator Keychain
entitlement behavior under `CODE_SIGNING_ALLOWED=NO`; the same four tests pass
4/4 when run signed at `/tmp/mypocket-keychain-signed-20261015.xcresult`.
This is still local evidence; device and production gates remain open.

## 2026-10-15 prepared budget create action

The red contract tests first failed because the budget proposal tool and
dispatcher were absent. The local implementation now validates name, safe
positive limit, calendar interval and optional UUID scopes without writing;
`budget_creation_proposal` maps to a native confirmation card; and the
confirmed action delegates to the existing budget validator through the
`budget.create` mutation ledger. Focused Go contracts and the native
DesignSystem target pass 90/90 at
`/tmp/mypocket-budget-action-designsystem-20261015.xcresult`; the signed
native regression passes 294/294 with zero failures/skips and no runtime
warnings at `/tmp/mypocket-full-native-budget-20261015.xcresult`. Backend
`go test ./...`, `go vet ./...`, the focused advisor race test, base-component
guards and `git diff --check` pass. Production/provider/device acceptance
remains open; this evidence is local-only.

## 2026-10-15 prepared budget update action

The local advisor now also exposes `propose_update_budget`. It validates the
owner-bound budget UUID, replacement name/limit/date/scope fields without
writing, maps `budget_update_proposal` to the shared native confirmation card,
and dispatches confirmation to the existing owner-scoped budget PATCH handler.
Focused Go tool/dispatcher contracts, the full backend test/vet/race gates and
native DesignSystem validation cover the new action. Production/provider/device
acceptance remains open; no environment was changed.

Focused native `DesignSystemTests` passed 90/90 at
`/tmp/mypocket-budget-update-designsystem-20261015/Logs/Test/Test-MyPocketIOS-2026.10.15_19-24-17-+0700.xcresult`.
The full signed simulator run completed 294 native unit tests with 0 failures,
then executed 100 UI tests with 64 environment-gated skips and one unrelated
pre-existing failure: `LaunchTests.testMoreProfileCardOpensAccountSettings`
expects the account subtitle in the More profile card. The full result is at
`/tmp/mypocket-full-native-budget-update-20261015/Logs/Test/Test-MyPocketIOS-2026.10.15_19-25-04-+0700.xcresult`.
This failure was recorded rather than masked; the budget-update slice had no
reported failure. Production/provider/device acceptance and the profile-card
fix were still open at that point.

The profile-card failure was fixed in the shared `MPProfileCard` base: it now
renders the supplied `accountSubtitle` and includes it in the accessible label.
The focused signed UI test now passes 1/1 at
`/tmp/mypocket-more-profile-subtitle-20261015/Logs/Test/Test-MyPocketIOS-2026.10.15_19-37-51-+0700.xcresult`.
The earlier full-suite result remains historical evidence from before this
follow-up fix; the final rerun below supersedes it for the native regression
gate.

That final signed rerun is now green: 294 native unit tests plus 100 UI tests
passed, with 64 documented environment-gated skips and no failures or runtime
warnings. Result:
`/tmp/mypocket-full-native-budget-update-profile-final-20261015/Logs/Test/Test-MyPocketIOS-2026.10.15_19-52-13-+0700.xcresult`.

## 2026-10-15 bounded product-help retrieval

The advisor registry now exposes the read-only `get_product_help` tool with
strict `query`/`max_results` bounds. It searches a source-controlled handbook
catalog (version `1.1`) using deterministic token matching and returns stable
source IDs, titles and bounded excerpts. No owner data, arbitrary URLs or model
markup are accepted; no-match returns `not_found`. Backend package tests and
vet pass. Native `product_help` decoding/card rendering passed the focused
XCTest artifact `/tmp/mypocket-product-help-native-green-20261015.xcresult`;
the warnings-as-errors simulator build, base-component guards and project
verification pass. Web typecheck/build, design contracts, feedback/auth and
transaction-jar contracts pass. This closes only the bounded local retrieval
slice: the full prepared-action dispatcher/real-provider replay contract, live
RAG behavior, device UAT and production rollout remain open.

## 2026-10-15 advisor jar handbook contradiction

The system handbook previously said the Assistant could not create jars even
though `propose_create_jar` and the explicit `create_jar` confirmation action
were already implemented. The handbook now describes jar creation as a
reviewable proposal only, and continues to reject direct edit/delete/payment
writes. The red provider-handbook assertion failed on the stale wording; after
the wording and assertion were reconciled, the focused usecase run passed:
`rtk proxy go test ./internal/usecase -run
'^TestAdvisorOrchestrator(PrependsProductHandbook|HandbookMatchesPrimaryAssistantTab)$'
-count=1`. No API, schema, permission, provider, database or runtime change.

## 2026-10-15 stable feedback dock (superseded placement checkpoint)

Root cause of the remaining feedback placement regression was the shared
obstacle-driven candidate search: every layout update could move the bubble to
another side or row, and the reported obstacle was not a guarantee that the
target control would remain reachable. The earlier checkpoint used a stable
lower-left dock and only honored footer clearance; it did not reserve a scroll
lane. The right-side follow-up above now supersedes that placement while
preserving the same overlay/no-lane contract. The test was red before the
change and green after it. Full
`DesignSystemTests` passed 90/90 with no runtime warnings at
`/tmp/mypocket-designsystem-feedback-final-20261015.xcresult`; the warnings-as-
errors simulator build and base/project guards also passed again. Authenticated
simulator checks for account security, More/category rows, lower-edge tapping
and no reserved scroll lane also passed individually. This is local layout
proof; populated-screen, dynamic-type, landscape and device/UAT overlap proof
remain open.

## 2026-10-15 advisor history lifecycle metadata

The history envelope keeps the existing message array in `data` and may now
return server-authoritative `meta` fields: `conversation_id`, `generation`,
`last_user_message_at`, `expires_at` and `server_time`. Reads still enforce
expiry without extending it. Backend package tests and native API contract
tests passed; native warnings-as-errors simulator build, base-component guard
and project verification passed. Evidence:
`/tmp/mypocket-advisor-history-native-20261015.xcresult`. Real PostgreSQL
rollout, production compatibility and owner-device UAT remain open.

## 2026-10-15 web local verification boundary

`npm run build`, `npm run test:design`, feedback/auth/transaction-jar unit
contracts passed. Browser tests that require an active Vite server or disposable
fixture were not counted: the current run hit `ECONNREFUSED` for the transfer
fixture and Router/Vite fixture setup errors for feedback/ledger. No server or
fixture was started or modified to force a result.

## 2026-10-15 savings queue edit eligibility

The shared `MPFinancialQueueStatus` menu now accepts an item-level `canEdit`
predicate. `WalletQueueEditPolicy` permits rejected/conflicted transfer drafts
in goal wallets when wallet data is available, while still requiring categories
for ordinary transaction drafts and refusing credit-wallet edits. The policy
unit passed at `/tmp/mypocket-goal-queue-policy-green-20261015.xcresult`.
The approved fault-injected savings 422 UI flow is now green against the
disposable local stack. PostgreSQL `mypocket_feedback_local` was migrated to
version 34 on `127.0.0.1:5432`; API `127.0.0.1:8081` was fronted by
`mobile-ios/scripts/offline-fault-proxy.mjs` on `127.0.0.1:18081`, and the app
was launched with `MP_API_BASE_URL=http://127.0.0.1:18081/api/v1`. The test
proved: rejected transfer draft appears in the savings scope; confirmed
100,000 đ / remaining 900,000 đ stay unchanged while pending; restart restores
the draft; edit to 22,500 đ and pass-through commit produces one transfer pair,
confirmed 122,500 đ balance / 877,500 đ remaining target, and no report-total
change. Evidence:
`/tmp/mypocket-offline-savings-20261015e/Logs/Test/Test-MyPocketIOS-2026.10.15_20-37-31-+0700.xcresult` (1/1 passed).
The preceding failed attempt was discarded as an environment wiring defect:
the app bypassed port 18081 and received 201, so no pending draft could exist.
No production, device, API schema, migration or release state was changed.

## 2026-10-15 assistant wallet action retry key

Native wallet confirmation cards now keep one idempotency key for their mounted
card lifetime and pass it to the existing wallet-create API. This prevents a
same-session timeout/retry from silently becoming a different request. The
focused API contract passed at
`/tmp/mypocket-wallet-action-key-green-20261015.xcresult`; no endpoint, schema,
permission, database or environment changed. New proposal cards now carry a
server-generated `action_id`; native extraction proof passed at
`/tmp/mypocket-wallet-action-id-green-20261015.xcresult`, so the same persisted
card keeps its replay identity across app restarts. Legacy cards retain the
mounted-card fallback.

## 2026-10-15 wallet confirmation summary

The native `wallet_confirmation` card now renders the goal/credit fields that
will be submitted (target/date, credit limit/statement/due details), plus the
total-inclusion and description fields when present. The shared masking and
base-text components are reused; the proposal decoder and confirmation API
boundary are unchanged. The focused contract passed at
`/tmp/mypocket-wallet-summary-green-20261015.xcresult`.

## 2026-10-15 durable confirmed wallet write

`POST /wallets` now forwards an optional `Idempotency-Key` to the PostgreSQL
wallet repository's existing `transaction_mutations` ledger. The owner-scoped
`wallet.create` operation atomically reserves, creates and stores the response;
same-key/same-payload replays, while changed payloads return conflict. Handler
tests also prove ASCII/duplicate-header validation and a keyed request fails
closed with `503 mutation_unavailable` when its adapter lacks durable support.
The
PostgreSQL replay/conflict test is present but skipped when
`TEST_DATABASE_URL` is unavailable; no database or environment was started.
The proposal identity is now persisted in the assistant message; the card
does not create a financial row until explicit confirmation. The PostgreSQL
replay/conflict test remains skipped without `TEST_DATABASE_URL`, so real-DB
execution and production rollout are still unverified.

## 2026-10-15 portfolio unused-asset deletion slice

The native portfolio now exposes `portfolio-delete-{assetID}` only for an asset
whose owner-scoped trade history is empty. Tapping it uses the shared
confirmation dialog and calls the existing `DELETE /api/v1/portfolio/assets/{id}`
route; success reloads summary/history, while traded assets remain append-only
and keep no delete control. The new API-backed UI test is
`testAPIBackedPortfolioUnusedAssetCanBeDeletedWithConfirmation`; the trade-flow
UI test also asserts that a traded asset has no `portfolio-delete-{assetID}`
control.

The first red run was blocked before UI interaction because the disposable local
PostgreSQL fixture at `127.0.0.1:55434` was unavailable, producing
`GOOGLE_MOBILE_LOGIN_INVALID` from the fixture login. This is an environment
availability blocker, not a failed product assertion. The app target still
passes a Debug simulator build with `SWIFT_TREAT_WARNINGS_AS_ERRORS=YES` at
`/tmp/mypocket-portfolio-delete-build-20261015`.

## 2026-10-15 Reports invalid-month stale-content fix

`ReportsMonthPolicy` now accepts only four-digit-year/two-digit-month values
with calendar months `01`–`12`; invalid input such as `2026-99` is rejected
before any request. Reports clears the prior summary, month detail and trend
rows before a new load and on failure, so a failed scope cannot display stale
totals. Unit proof passed at
`/tmp/mypocket-report-month-unit-final-20261015.xcresult`; the authenticated
UI regression `testReportsRejectInvalidMonthBeforeNetworkAndShowRetryState`
passed at `/tmp/mypocket-report-month-ui-final-20261015.xcresult`. The PostgreSQL-backed
valid-report UI fixture remains environment-dependent and was not started or
claimed by this code-only fix. The Debug simulator build and base/project
guards pass. A separate full native unit attempt ran 280 tests with four
pre-existing simulator Keychain failures (`-34018`/`keyUnavailable`) in API and
outbox tests; no Reports assertion failed and no environment or product-code
workaround was applied. Backend `go test ./...` passes across all packages;
log: `/tmp/mypocket-backend-regression-20261015.log`.

## 2026-10-15 Overview/Reports encrypted read snapshot fallback

The existing owner/server-namespaced encrypted `FinancialSnapshot` now keeps
optional server-returned `MonthSummaryModel` and `InsiderSummaryModel` values.
Overview restores the ledger and report cards after transient transport/408/429/
5xx failures; Reports restores only when both cached report values match the
requested `YYYY-MM`. No local financial recomputation or write/outbox mutation
is performed. The focused cache/fallback suite passed 9/9 at
`/tmp/mypocket-report-cache-fallback-green-20261015`; Debug build, base/project
guards and `git diff --check` pass. A separate full native unit attempt ran 283
tests with the four existing simulator Keychain failures listed above; no cache
assertion failed.

## 2026-10-16 Reports owner scope and summary fail-safe

`FinanceAPI.monthSummary` and `FinanceAPI.insider` now accept the optional
owner ID and pass it to the existing `expectedOwnerID` request guard. Reports
passes its owner context to both reads and awaits both payloads as one atomic
report load; a month-summary failure therefore clears the whole report and
uses the existing retry/error state instead of rendering insider-only or empty
breakdown/activity sections. The focused encrypted-cache/report suite passes
12/12 on the signed iPhone 16 simulator at
`/tmp/mypocket-reports-hardening-signed2`. Build-for-testing also passes with
signing disabled at `/tmp/mypocket-reports-hardening-build`. No backend/API
schema, offline/outbox, or cache-format change was made.

## 2026-10-15 Feedback resolution tab behavior

The owner-resolution flow remains native-only and uses the existing feedback
PATCH contract: ticking a row moves it from `Chưa giải quyết` to `Đã giải
quyết`, and unchecking it reopens the same row. The concurrency regression
`APIClientTests/testFeedbackRefreshAndResolutionCannotOverwriteEachOther`
passed 1/1 at `/tmp/mypocket-feedback-unit-validation-20261015.xcresult`.
The API-backed UI proof was attempted but the disposable fixture rejected its
test identity with `GOOGLE_MOBILE_LOGIN_INVALID` before the screen opened;
this is an environment/fixture gate, not a product assertion failure. No
production feedback status, environment or auth configuration was changed.

## 2026-10-15 Advisor idle retention foundation

The explicit “Chat mới” lifecycle slice is implemented locally: backend route
`POST /api/v1/ai/advisor/conversation/reset` requires `Idempotency-Key`, rotates
the owner generation in one transaction, purges advisor-only content and keeps a
content-free reset ledger. The service cancels queued/running in-process
provider contexts returned by the first reset; replayed keys do not cancel new
runs and replay returns the original server timestamp. The internal purged-run
list is omitted from the HTTP result. Native Assistant exposes the action through the shared overflow menu,
persists the key for uncertain retries, confirms before resetting, and never
submits the composer draft automatically. Native API/expiry contract proof
passed 2/2 at `/tmp/mypocket-reset-ios-test-20261015-r2.xcresult`; backend full
Go tests and vet pass locally, and the focused race packages pass. The native
base/project guards and simulator Debug build also pass at
`/tmp/mypocket-final-debug-build-20261015-r4`.
Production migration and rollout remain intentionally open.

The backend now records `last_user_message_at` through additive migration
`000033_advisor_idle_reset.up.sql`, backfilled only from accepted user messages.
`AdvisorConversation` expires at the inclusive 30-minute server-time boundary.
History/run access and new submits purge advisor messages, events and facts while
keeping minimal run tombstones for idempotency; stale assistant/event writes are
rejected by generation checks. The worker uses a bounded 100-row `SKIP LOCKED`
pass, and purged request retries map to `410 advisor_history_expired`.
Pure entity TTL, worker contract, advisor handler mapping and existing backend
advisor tests pass locally at `/tmp/mypocket-advisor-idle-foundation-20261015.log`.
Native iOS expiry handling now has a focused 1/1 unit proof at
`/tmp/mypocket-assistant-expiry-test-20261015-r5/result.xcresult`: a 410
`advisor_history_expired` or other typed 410 clears the stale local chat, while
ordinary 400 advisor errors do not. The send/history paths preserve the current
composer text and require an explicit resend; they do not auto-submit. Real
PostgreSQL migration/expiry race proof, expiry metadata/manual reset, and
production rollout remain separate gates; no database or environment was
started or modified by this change.

## 2026-10-14 offline financial acknowledgement checkpoint

20261015 reconciliation: category84393exit0,280/280PASS0skip/no runtime warnings
(279unit+actual API/native category CRUD/archive/restore/delete),
`/tmp/mypocket-reconciled-category3-20261015.xcresult`. Reviewed post-save
API/exact-ID reveal and20drag/21check helper closes prior false failures.
Strengthened delete API collection absence and scope preservation are part of
final12business+279unit gate11171: exit0,291PASS/0skip/0 runtime warnings,
`/tmp/mypocket-reconcile-full-green-20261015.xcresult` (824.704s). The gate
includes the earlier isolated category proof plus credit and recurring flows.
The three prior failures were fixed in test infrastructure only: dismiss the
keyboard before revealing the category icon, select the unique innermost visible
ScrollView for nested controls, and explicitly select/read back the recurring
fixture wallet. Exact persisted-data assertions remain unchanged.
Full Go race32091exit0 against
dedicated realPG55434/Redis16379 and vetexit0; no backend source changes this turn.
Fresh authenticated production feedback48open, no status writes or credential
copy. Production sync404, publicdocs/docs404 but Swagger index200; vm100 API
auth-12572c87healthy/workerrelease-73c4c37, managerfalse. Maclocked/manager auth
requires1Password. No install/rollout/data reset; canonical reconciliation is the
active status index, not older “missing note editor/offline not implemented” text.

20261015 approved numeric final48656exit0:287/287PASS0skip/no runtime warnings,
`/tmp/mypocket-numeric-chat-final-20261015.xcresult` (279unit+8chatUI). Range,
goal invalid percent and masking mutationRED→GREEN; capturesCA7766AF/5CFE537F
inspected, base/project/diff guardsPASS. Earlier business53977:281PASS/1category
FAIL;27187 scope diagnostic shows3wallets actually selected, last77512 singleton
fixture run1FAIL post-save lazy-name wait. Three renewed helper attempts stopped,
no fourth guess or all-business-green claim. Production API Release sign18960
live waiting Keychain; phoneunavailable, no newinstall. Idle30minapproved/spec
committed9ab9248d, no deletion/implementation proof yet; fullgoal remains open.

20261015 resumed completion audit: installed iPhone candidate/local cancellation
fix are verified partial deliverables; full native goal unachieved. Native numeric
regression1FAIL, business prior2FAIL, privacy/Feedback/cache/initial Stop and idle
reset lack required implementation/acceptance. No live verified wait remains;
process check terminal and SSH identities0. Same unanswered design/access gates
for3current turns; goal blocked awaiting input, not complete or user-paused.

20261015 actual native numeric diagnostic: `/tmp/mypocket-native-progress-range-red-20261015.xcresult`
exit65,0PASS/1FAIL0skip/no runtime warnings. Actual progressTarget accepts
9007199254740992 beyond finance safe range; expected XCTAssertNil fails. Not a
rendered crash test/fix; simulator-only test, installed Release unchanged. Idle
reset source audit documents replay/generation interactions, no expiry/deletion
proof or changed product behavior claimed.

20261015 device delivery: build exit0; initial strict signature check failed on
sealed .cstemp reference, explicit re-sign exit0 and strict/deep verifyPASS.
Final device install exit0, launch exit0, process listing confirmsPID22311 in
new app container. `/tmp/mypocket-stable-device-install-verified-20261015.json`,
`/tmp/mypocket-stable-device-launch-20261015.json`. Production API configured,
health200; existing app replaced without uninstall/reset. Physical business/FaceID
UAT not performed; backend patch not deployed and full-goal acceptance remains open.

20261015 candidate verification: `/tmp/mypocket-stable-native-units-20261015.xcresult`
273/273PASS0skip/no runtime warnings; `/tmp/mypocket-stable-backend-race-20261015.log`
full Go race suitePASS (DB-dependent coverage is separately exercised below).
`/tmp/mypocket-advisor-cancel-db-green-20261015.log` real PostgreSQL4/4PASS,
including deterministic cancellation while answer waits, foreign-owner/new-run
and provider-context tests; earlier RED retained. Base/project/diff guardsPASS.
Production `/api/v1/health`200; Release API base verified in generated Info.plist.
Signature/install pending; no full product acceptance or backend deployment claim.

20261015 blocked audit (read-only): no live test/proxy/go-test; ssh-add-l exits1
no identities. Current goal explicitly unachieved: native chat/history local
gates pass, business2FAIL and realDB cancellation raceFAIL remain. Numeric/privacy
actual-source traps and fallback masking gap uncorrected. Specific design gates
unanswered over>=3goalturns; no product/runtime/deployment change in this audit.
Wholegoal blocked, not accepted/complete; retain prior evidence and full scope.

20261015 cancellation70798exit1 realDB+service race-gate:3PASS/1FAIL0skip/no Go
data-race report. `/tmp/mypocket-advisor-cancel-db-red-20261015.log`. Controlled
conversation lock: cancellation commits, late answer insertion erroneously
succeeds/historyseq2assistant persists. Sequential cancel/foreign-owner/new-run
and blocking-provider context test pass. Test-ownedusers cleanup0remaining. This
is reproduced transactional race, no productfix; bounded review pending. Initial
native Stop unavailable before syncPOST returnsrunID is source-only finding.

20261015 persisted native history27531exit0:274/274PASS0skip/no runtime warnings,
`/tmp/mypocket-advisor-history-final-20261015.xcresult` (273unit+1realDB/API/UI).
51synthetic messages/latest50/older1, preserve old viewport, scrollback one51,
restart and independently identical persisted API rows pass. Capture76D78B35
inspected. Initial40442 offscreen AX assumption corrected without product/data
assertion relaxation; failure retained. Fixture exact cleanup0/0/0, duplicate
create refuses overwrite, remote Docker override refused. Guards/project/diff pass.
Not live provider cancellation/send race/full a11y/wholegoal acceptance.

20261015 native advisor adapter:24082exit0 PASS2/2, final72133exit0 PASS273/273
unit0skip/no runtime warnings, `/tmp/mypocket-advisor-wire-units-20261015.xcresult`.
New realclient/FinanceAPI tests assert chronological history+parts/cursor/auth and
explicit cancelPOST/session/idempotency/noresubmit, transport boundary fake only.
Characterization of existing behavior, not productfix/nativeUI/liveprovider/DB
pagination or cancellation race acceptance. Guards/tests/project/diff pass.

20261015 current-helper chat88653exit0:8/8PASS0skip, no runtime warnings,
`/tmp/mypocket-chat-current-helper-20261015.xcresult`. Normal/AX/landscape/draft/
suggestions/disabled-send/APIoverview actualUI cases pass. GeometryA80B6F81 gives
63pt landscape conversation/44ptsend. FD268710 cropped capture, F165D9F3 AXFeedback
overlap remain unaccepted. Not populated history/cancel/liveprovider E2E proof.
No product/test source change; latest fullunit271PASS is separate80693 gate, not
claimed as same-run combined total. Public contract unchanged this test-only pass.

20261015 actual-source diagnostic: Envelope.swift decode/raw prettyDescription
valid12500PASS; JSON1e100SIGTRAP Int overflow. Exact extracted transactionCount
42PASS/Double(Int.max)SIGTRAP; progressTarget accepts nonrepresentable maxDouble.
No timeout. This is Swift CLI, NOT native rendered regression/E2E/fix. Fallback
masked omission source-only; human/agent guide/catalog/screen caveat reconciled.
No product/test source changed, bounded design pending. Fullgoal remains active.

20261015 numeric audit: source+Swift diagnostic exit0 prove rounded-bound
representability mismatch and fractional truncation; finite1e100 JSON accepted.
This is NOT a rendered regression/crash/green fix. Numeric/card bounded design
pending approval; actual RED→GREEN/unit/chat/visual proof still required.
[Work](tickets/NATIVE-AI-CARD-NUMERIC-SAFETY-20261015.md).

Latest80693exit65:273PASS/2FAIL0skip (271units+literalprobe+walletCRUD PASS), no
runtime warnings, `/tmp/mypocket-reveal-crud-regression-20261015.xcresult`.
Category exact create/name/API scope PASS; subtitle asserted on wrong AX property,
subsequent filter/edit failures unaccepted. Wallet transaction data checks run
but helper adds two pinned-save false-visibility failures (window667/shell611),
pending bounded design review. Not a green suite or fullbusiness acceptance.
No runtime/base/API patch; guards/guardtests/project/diff pass before this gate.

2026-10-15 reveal ROOT/RED:54467exit65, draft-only literal probe fails immediately
after helper scrolling, before focus/save. Trace records20swipes on QuickType
ScrollView(0,407,375,44), not editor(0,0,375,139). No POST. Owner approved semantic
test-only correction; probe46637 literal boundaries PASS but full swipe overshoots
139pt container. Bounded drag probe34635exit0 PASS1/1,0skip/no runtime warnings;
`/tmp/mypocket-reveal-bounded-probe-20261015.xcresult`. Units+3CRUD80693 running.
No product keyboard fix or fullbusiness green claim.
Exact name/data assertions unchanged; public API/human/agent docs unaffected.

Latest diagnostic7487exit65:272PASS/2FAIL0skip (271units+transactionCRUD correct
fixture nowPASS), no runtime warnings, `/tmp/mypocket-business-diagnostic-20261015.xcresult`.
Category/wallet exact-name assertions after inputPASS; afterfocus/save create
assertions FAIL, AX BD6CAF77/BD47573A confirms appended prose. No product/root fix
or fullbusiness acceptance; literal-name design pending. Normal8081 defaultlaunch
25086, no18081listener/tests running; guards/tests/project/diff pass.

2026-10-15 current sweep86110exit65:280PASS/3FAIL0skip (271unit+9UI pass),
`/tmp/mypocket-current-business-sweep-20261014.xcresult`, no runtime warnings.
CategoryCRUD/walletCRUD unexpected appended names; transactionCRUD wrong first
wallet-detail selection. Traces inspected, product/input root unresolved. Test
diagnostics/scoping corrections now run separately7487; no broad-green claim.
[Work](tickets/NATIVE-CURRENT-BUSINESS-SWEEP-20261014.md).

Approved savings rerun94298exit65:273PASS/1FAIL0skip,271unit+strong credit+goal503
PASS; goal422 menu opens normal Add, blocked for review. Bundle
`/tmp/mypocket-goal-queue-approved-final-20261015.xcresult`, no runtime warnings.
503 API/rendered112500/remaining887500, same key/one pair/unchanged report proves
receipt refresh without manual reload; credit explicit source/pair/balances/report
pass. Captures98B2E756/76F1A2EF inspected, not full visual acceptance. Proxy stopped.

AI current audit92901exit0: focused advisor/proposal tests in usecase and HTTP
packages pass under race detector, log
`/tmp/mypocket-advisor-current-audit-20261014.log`. Stub provider/reader proof only,
not real-DB/liveLLM/production. Source confirms static versioned handbook, no
retrieval in inspected advisor modules. The Assistant-in-More wording was fixed
on 2026-10-15 with a provider regression; product-help/RAG remains partial and
confirmation summary design awaits owner reply.

AI wallet proposal validation: RED18261 four malformed confirmations→strict
decoder; final8709exit0,279/279PASS0skip (271unit+8chatUI), no runtime warnings.
`/tmp/mypocket-wallet-proposal-final-20261014.xcresult`. Valid payload/default/null
and malformed money/date/day/type tests pass; guards/project/diff pass. Not live
creation/replay/provider/device proof. [Work](tickets/NATIVE-AI-WALLET-PROPOSAL-VALIDATION-20261014.md).

Rejected transfer editor:52081exit0,269/269PASS0skip (268unit+1UI), no runtime
warnings, `/tmp/mypocket-rejected-transfer-20261014.xcresult`. Restart, unchanged
refusal/keycount1, changed22500/keycount2, one pair, balances77500/122500,
unchanged monthly report, scoped detail verified; capture0134D5FF inspected.
No product fix needed; savings gate still blocked. Guards/project/diff pass.
[Work](tickets/OFFLINE-RECOVERY-E2E-20261014.md). Normal8081 build15493exit0.

Superseding approved chat gate: final20967exit0,276/276PASS0skip
(268units+8UI), no runtime warnings. Keyboard tabs hidden/restored, normal/AX
menu roundtrip retains draft, actualAPI overview header sheet and landscape
conversation63pt pass. Bundle `/tmp/mypocket-chat-approved-final-20261014.xcresult`.
Base/guardtests/project/diff pass. Captures inspected; landscape export crop and
compact transition image do not prove full pixel acceptance. Owner/device and
broader a11y gates pending. [Work](tickets/NATIVE-CHAT-ADAPTIVE-20261014.md).

Chat adaptive gate:26992 1PASS/1FAIL (landscape7pt),10130 one-linefixstill7pt→
reverted.82726 realAPI0pt, fake diagnostic absentoverview querybug corrected.
Final88928exit65,272PASS/2FAIL0skip,no runtimewarnings,
`/tmp/mypocket-chat-adaptive-final-diagnostic-20261014.xcresult`.268nativeunits+
4otherUI pass; actual/fake landscape fail0/7pt respectively. Numericgeometry
9E8C4D1E+114B7397 and843995F5 render inspected; composeroverlaps tab onrealAPI.
Largest Dynamic Type control checks PASS, renderednavoverflow NOTaccepted.
Guards/project/diffpass. [Work](tickets/NATIVE-CHAT-ADAPTIVE-20261014.md) awaits
adaptivechrome design; no newproductfix/prod/device/fullUI acceptance.

New manualRED Reports: localSEDEBUGfixture valid2026-10→invalid2026-99+load,
GETmonths/insider40023:02:55; inspectedcapture
`/tmp/mypocket-report-invalid-month-repro-20261014.png` shows retained oldsummary
and no visiblefailure. NativeUI throughDeviceHub; no writes/production/reset,
returnedOverview. Not nonzero corruption/offline/Months note proof. No automatic
regression/fix yet. [Audit](tickets/OFFLINE-READ-SCOPE-AUDIT-20261014.md).

Read-cache audit ONLY: inspected Overview/Reports/Months/cache/injection/API
sources show missing overviewcache, partial publication and stale scope risks.
No UI failure reproduction, new tests or completion evidence from this inspection.
[Audit](tickets/OFFLINE-READ-SCOPE-AUDIT-20261014.md). Owner approach review pending;
runtime scope/fallback/owner/month/note-write safety gates required before acceptance.

Feature diagnostics: RED93397 3/3 actual formatter leaks→allowlisted class/status;
6callsites+localfixturelabel wired. Final66267exit0,272/272PASS0skip (268units+
4chatUI),no runtime warnings,`/tmp/mypocket-feature-privacy-green-20261014.xcresult`.
Problemfields/transport/decode/NSError canaries omitted; userMessage unchanged.
Chat overview/details/draft/navigation/keyboard/suggestions regression pass;
97F7C806 capture inspected composerabove keyboard. Not injectedviewerror/liveLLM/
wholeUI/prod/device acceptance. Guard/guardtests/project/diffpass,APIhealth200.
[Work](tickets/NATIVE-FEATURE-LOG-PRIVACY-20261014.md). Goalqueue/Feedback unchanged.

HTTP diagnostic privacy:39064 validRED2/2 exposes200 query and422 body/query.
Metadata-only correction→92498exit0,267/267PASS0skip,265units+ordinary wallet
create/update and transferCRUD APIUI,no runtime warnings. Bundle:
`/tmp/mypocket-http-privacy-green-20261014.xcresult`. Typed422 details unchanged;
stdout retains method/status/generated requestID, no canary. Guard/guardtests/
project/diffpass,APIhealth200. [Work](tickets/NATIVE-HTTP-LOG-PRIVACY-20261014.md).
Separate reflected-error logs/goalqueue/fullgoal/prod/device acceptance remain.

Goalqueue gate27662exit65:263unitPASS/2UIFAIL0skip,
`/tmp/mypocket-goal-queue-green3-20261014.xcresult`. After restart the actual goal
selector is present; test incorrectly queries aggregate selector. Pending queue
was reached before restart, but receipt-driven savings refresh and rejected
transfer edit are NOT verified. Rule10 review required before next cycle.
[Work](tickets/OFFLINE-GOAL-QUEUE-20261014.md). No acceptance/fullgoal claim.

Latest recoveryE2E:28248 3/3 ordinary422/date/unchangedkey/discardbothstates;
4135 2/2 transfer503/drop samepair/balances/unchangedmonthincomeexpensecount;
32155 267/267 (263unit+409/old3faultUI),0skip. Detail59472 4/4 opens actualserver
rows, globalnotequery weak; scoped37055 2UIRED/263unitPASS revealed5 siblingIDs.
Contain grouping→35239exit0,265/265PASS0skip/no runtime warnings,
`/tmp/mypocket-recovery-contained-final-20261014.xcresult`: unique detailcontainer
and scoped note for ordinary409/transferdrop. B8F61E91 inspected22500/day13/note,
unchangedlayout. Finalnormal44956exit0,265/2650skip (263unit+ordinarycreateedit/
transferCRUD),`/tmp/mypocket-recovery-normal-final-20261014.xcresult`. Proxy5/5/
guards/project/diffpass; normal8081build restored, testproxy stopped.
[Work](tickets/OFFLINE-RECOVERY-E2E-20261014.md). UnderlyingledgerAX/Feedback,
transferrejectededitor/goalqueue/cache/auth/device/allgoal gates remainopen.

Latest lifecycle: startup/network/timerRED72468→258units; pauseRED18554 lateack+
nextwrite→85017exit0,263/263 (261unit+2APIUI create/edit+transfer). FinalUI13577
exit0,3/3PASS0skip/no runtime warnings,
`/tmp/mypocket-lifecycle-ui-final-20261014.xcresult`: Home/server-recovery no
backgroundwrites over2s, foreground automatic apply, samekey restart/menuRetry,
real commitdrop replay; each exactly1serverrow. Finalunits94868exit0,263/2630skip,
`/tmp/mypocket-lifecycle-unit-final-20261014.xcresult`: FIFO100s/early99s, no early
lease steal/60s lease timer, unsupported30s, loggedoutstale timer/authblocked wakes.
Guards/project/diffpass, normal8081build restored, testproxy stopped. Screenshot
D63C86B9 inspected; known Feedback row overlap remainsblocked. Not radio-switch/
device/a11y/fullrecovery/prod acceptance. [Work](tickets/OFFLINE-FOREGROUND-LIFECYCLE-20261014.md).

Latest category/replay/transfer gate: creationRED67864 (nil-category durable
insert/invalid replacement) and compatibilityRED69099 (legacy localinvalidPayload)
→final40360exit0,256/256PASS0skip/no runtime warnings,
`/tmp/mypocket-category-compat-final-20261014.xcresult` (255units + transferUI).
RealSQLite invalid-create/no-child/originalpreserved; actualAPIClient legacy400
preserveskey/body. UI explicitly picks source/destination; actualAPIreads prove
100000→87500/112500→77500/122500→100000, bothrowsreportfalse and finalserverpair
absent. Not aggregate reportUI/offline transferfault proof. Guards/project/diffpass,
APIhealth200; [work](tickets/OFFLINE-CATEGORY-VALIDATION-20261014.md). Lifecycle/
remaining fullnative/prod/device/UAT stillopen.

Latest compact/fault: RED34418 missingmenu; GREEN23703 1/2 (Feedback intercepted
503 menu tap; known blockedplacement). Scroll real menu clear, no overlay hiding.
Final73398exit0,2/2PASS0skip/no runtime warnings,
`/tmp/mypocket-compact-green2-20261014.xcresult`. Actual503/realcommitdrop→sameapp
restart→sameintent→menuRetry→one ledger row, samekey/amount/note/category; final
commitdrop201/200/200,2drops,1hashedkey. Proxy3/3 and guards/project/diffPASS.
Pending3CF00F85 inspected375pt: compact rows; bubble still overlays selector,
not visual acceptance. [Work](tickets/OFFLINE-FAULT-E2E-20261014.md).
Finalunit29293exit0,252/252PASS0skip/no runtime warnings,
`/tmp/mypocket-compact-unit-20261014.xcresult`; normal8081build, testproxy stopped.
Fullrecovery/transfer/scheduler/background/auth/a11y/device/prod remainopen.

Newestrecoverygate: RED93734/6944store,88919coordinator,46389timecomparison→final
5403exit0,252/252units0skip,/tmp/mypocket-recovery-editor-20261014.xcresult.
RealSQLite rollbacktrigger/twostore concurrent samekeyreplacement, allowed/refused
retry/discard/edit states/owner/key/body and timezoneequivalence tested. Shared
recovery UI/editor code compiled, not pending/conflict/date visual/E2E/a11y proof.
Guards/project/diffpass. Normalcreate/editAPIUI58897exit0,1/1PASS0skip,
/tmp/mypocket-recovery-normal-ui-20261014.xcresult.
[Work](tickets/OFFLINE-QUEUE-RECOVERY-20261014.md).

Newestqueue runtime: RED3560→242units; concurrencyRED41241→drainEpoch. Final23701
exit0,243/243units0skip,/tmp/mypocket-queue-final-20261014.xcresult. Persistbefore
send/stablekey/unsupported/auth/accountswitch/noncooperativeoldcallback tested.
Actualcreate/editUI45495exit0,1/1,/tmp/mypocket-queue-create-ui-20261014.xcresult
plus readonlytestappSQLite1row563sealedbytes. Detail4685B100 inspected, not pending
UIproof. Guards/project/diffpass. [Runtime](tickets/OFFLINE-RUNTIME-QUEUE-20261014.md).
Recovery/scheduler/networkE2E/background/auth/prod/device remainopen.

Newestcache/readgate:13769exit0,239/239(238units+actualAPIkeyUI)0skip,
`/tmp/mypocket-cache-final-20261014.xcresult`; ledgerAPIUI17267exit0,1/1,
`/tmp/mypocket-cache-ledger-ui-20261014.xcresult`. REDcache95867/model90809/
HTTPstatus41489→GREEN. Encryptedrestart/owner+server isolation/wrongkey/corrupt
nonoverwrite/fallbackclasses/modelcache/auth denial andHTTPstatus authority tested.
Normalledger screenshot59D42ADA inspected; bubblecovers lastamount knownblock.
StaleUI/accessibility/landscape and outboxruntime/overview/refetch stillopen.
[Cachework](tickets/OFFLINE-FINANCIAL-CACHE-20261014.md). Guards/project/diffpass.

NewestTask5 engine gate: final52696exit0,232/232units0skip bundle
`/tmp/mypocket-sync-engine-final2-20261014.xcresult`. RuntimeRED45673 initialfive
enginecases; deterministicJSONRED80149 fixed onlyoutboxencoder. Logout noncooperative
latecallback/explicitresume/backoffcap and realAPIClientadapter/store timeoutreplay
covered with externalHTTPfake. FullGoRace22266exit0; guards/project/diffpass.
Coalescing/engine401integration, runtimecache/refetch/UI/scheduler stillopen.
[Enginework](tickets/OFFLINE-SYNC-ENGINE-20261014.md). No prod/device/releaseclaim.

PASS local: validator runtimeRED63069→220unitGREEN; HTTP adapter runtimeRED38814
→70626exit0,223/223units0skip, bundle
`/tmp/mypocket-ack-http-final-20261014.xcresult`. Existing standard SE LOCAL_API_UI
warnings-as-errors gate, MyPocketIOSTests only. Real encrypted SQLite lease plus
APIClient and external HTTP double; exact receipt/resource checks, stable saved
body/key across201/200 replay, transfer array, no pending/expired/unfrozen send.
Base/guardtests/project pass. Not live backend financial E2E or UI acceptance;
sync actor/backoff/cache/outbox UI/prod/signed install remain unverified.
[Ticket](tickets/OFFLINE-CLIENT-ACK-20261014.md).

## Native chat consolidation — 2026-10-14 execution checkpoint

Local full-key API UI: exact owned10344 stopped after current binary build,
19887/session46368 healthy8081, dedicatedPG55434/Redis16379, existing fixture seed
unchanged. No prod/8080/user5432/6379 access. Initial53912PASS212/212; extended
51955RED keysecret/copy remains and revoked row returns on reopen. After transient
createdKeyID/filter,70976PASS212/2120skip bundle
/tmp/mypocket-key-revoke-reload-final-20261014.xcresult. APItest creates unique
synthetic name, copies viaUI confirmation, revokes, leaves/reopens, asserts no
row/secret/copy. RawAPI revocation metadata retained,0active synthetic keys,
0wallets/0transactions/schema32false. Screenshot FF8AB1F2 inspected in
/tmp/mypocket-key-revoke-reload-visual2-20261014: expected empty key state, bubble
nevertheless covers timezone control. Not universal visual acceptance.

Owner-bound HTTP/CAS: compile finding exhaustiveAPIError corrected before valid
18476RED5cases;54401PASS217/217. CAS failclosed scaffold3343RED,86604RED2cases.
Implemented app-process atomic Keychain/memory replace+clear and refreshed owner/
credential tuple guard. Final58452exit0,219/219PASS0skip(218units+actualkeyAPIUI),
/tmp/mypocket-owner-session-cas-final-20261014.xcresult via standardSE
MyPocketIOSTests + LaunchTests/testAPIBackedAccountAPIKeyCreateAndRevoke,
LOCAL_API_UI/base8081/warnings-as-errors.16actualKeychain candidates one winner;
stale replace/clear/foreign return/new sameowner/new otherowner/clearedsession,
401 afterswitch cannot clear new credentials. Metadata preserved and one valid
refresh retains original body/key/requestID. Existing restore/device/multipart
green. Base/guardtests/project/diff pass, human+agent/auth/native docs reconciled.
[ADR022](../decisions/ADR-022-owner-bound-native-refresh.md). No backendwire/schema
change, Swagger/ERD notneeded. Receipt/resource adapter/engine/cache/UI still open;
no physical protection/prod/install/release/allgoal completion claim.

Full-owner keys: validRED52399 four auth/service cases→21202GREEN. RealPGHTTP
65981REDselflogout401→49557GREEN; refined foreign-revoke78012RED500→typed404.
Final18813exit0 with dedicatedPG55434+Redis16379 `go test -race ./... -count=1`,
`go vet ./...`0 after regeneratedSwagger. Log:
`/tmp/mypocket-full-key-final2-go-20261014.log`. Old stored feedback:read key
now creates/replays ordinary row and atomic transfer (3rows,2excluded), foreign
wallet/receipt/read/revoke denied, keymanagement/selflogout and otherowner
isolation, no secret/lookup/foreignIDs/rights in List; oldSQLscope arrayunchanged.
RealJWT/Redis TTL/logout and revokedexpiredkey route401 preserved. Generic
repository error sentinel restored after compile dependency finding; not scope.
NativeRED59411 wire scopes/selector/missingwarning→97143exit0,212/2120skip
`/tmp/mypocket-full-key-native-final-20261014.xcresult`,211unit+selected
LaunchTests/testAccountAPIKeyEditorUsesStableBaseControls onSE LOCAL_API_UI
warnings-as-errors. ScreenshotC1E878F7 inspected: selector absent, base form/warning
present; Feedback obscures helper. Fake-session UI, not updatedAPIkey create/
copy/revoke or nooverlap/physicalacceptance proof. [Ticket](tickets/API-KEY-FULL-ACCESS-20261014.md),
[ADR021](../decisions/ADR-021-full-owner-api-keys.md). Docs human/agent/API/auth
reconciled; no schema/APIrestart/prod/install/release. Full-goal still active.

FrozenTravel contract: backendRED66771exit1 (falseignored/wrongactive trip,
changedpolicyreplayed, missingcapfield); legacy receipt goldenhash baselinePASS.
Focused12705exit0 then fullGoRace51567exit0/vet0 after `go generate ./cmd/api`.
Commands with dedicated PostgreSQL55434+Redis16379 env: `go test
./internal/controller/http -run 'TestOfflineCreateTravel|TestSyncCapabilities'
-count=1`, `go test -race ./... -count=1`, `go vet ./...`.
NativeRED34132 twofail oldcapunsafe and missingfalsequeuepayload; final59173exit0
`/tmp/mypocket-frozen-travel-native-final-20261014.xcresult`211/211PASS0skip via
standard MyPocketIOSTests SE LOCAL_API_UI warnings-as-errors command. New tests
block missing/unknowncontextordinarycreate but permit independentlysafe transfers;
stored trip/no-trip payload emitsfalse and duplicateenqueue retainsID. ActualDB
HTTP proves old activeA→newB, false none, explicit inactiveA, legacy/null/trueB,
fixed date server readback, policyconflict, deletionreplaybeforevalidation/foreign
new404 and exactlyone row/receipt, independent oldJSONhash replay. One initial
test-interface compile error corrected before runtimeRED; not counted as behavior.
Swagger generated and inspected create-only property, PATCH unchanged. Guard/
guardtests/project/diff pass. No prod/migration/install/release or API-key policy
implementation; new owner full-key decision queued. [Work](tickets/OFFLINE-FROZEN-TRAVEL-20261014.md).

Encrypted outbox store: persistence runtimeRED70442 sixfail/0pass; after SQLite/
AESGCM/Keychain6/6pass3540, security expansion11/11pass40086. Lease runtimeRED43035
fourfail/11pass (nil lease), then15/15pass63674. Final88836 exit0,
`/tmp/mypocket-outbox-store-final-20261014.xcresult`209/209PASS0skip,16store tests.
Command: `rtk proxy xcodebuild test -project mobile-ios/MyPocketIOS.xcodeproj
-scheme MyPocketIOS -destination 'platform=iOS Simulator,id=182D41A2-4FC6-445B-A592-70FDCB5C662E'
-derivedDataPath /tmp/mypocket-feedback-resolution-tests
-resultBundlePath /tmp/mypocket-outbox-store-final-20261014.xcresult
-only-testing:MyPocketIOSTests -collect-test-diagnostics never
'SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI'
MP_API_BASE_URL=http://127.0.0.1:8081/api/v1 SWIFT_TREAT_WARNINGS_AS_ERRORS=YES -quiet`.
Real tempSQLite/CryptoKit and actual SimulatorKeychain, cleanup exact testpaths/
service: persisted restart/nonplaintext files/owner+server/FIFO/samekeyconflict,
two-store enqueue and claim races, missing/wrongkey nonwipe, copiedciphertext
denial, unknownschema nonreset, transfer fixedinstant/singleintent/operationscope,
invalidinput, retrydelay, expiredlease/staleresponse/terminal/authblock/foreign
outcome refusal. Two initial compile errors in test setup corrected before valid
RED, never counted as behavioral proof. No dependency/concurrency warning bypass.
Baseguard/guardtests/project/diff pass. [ADR020](../decisions/ADR-020-encrypted-native-outbox.md).
No runtime UI/session/sync wiring, physical protection claim, backend/migration/
prod/install/release. Full-goal UAT and delayedTravel contract remain open.

Server capability gate: HTTP76809 valid404RED durable/legacy/nil config; focused
1027 pass and full Go race41946/vet with localPG55434+Redis16379 pass. Actual
JWT/Redis integration200beforelogout/401after, unauth401; unsupported writer false
with nonnull empty operations. OpenAPI generated/inspectedtyped200/BearerAuth/401.
Native69552 runtime2fail/1pass baseline explicit fail-closed scaffold, not a
compile failure. Final77223 terminal0,
`/tmp/mypocket-sync-capability-native-final-20261014.xcresult`193/1930skip via same
full MyPocketIOSTests LOCAL_API_UI/SE/base8081 template.4new tests: known create-only
authority, version/disabled/missing-op/key/meta/malformed denial, GET/session wire,
old404 propagated with exactly1GET/no financial fallback. No runtime actor/outbox
wiring or actual network-to-local-API UI proof. Guard/project/diff pass; fixture
owners/Redisnamespace0/schema32false. Human+agent Markdown/Swagger reconciled;
no migration/API restart/prod/sign/install/release/UI style change. Whole goal open.
Post-generation full Go race75975 exit0/vet0 verifies finalized Swagger/source;
no concurrent generation or live handles in this final gate.

Offline existing auth characterization: focused71010 pass; final
TEST_DATABASE_URL=<dedicated55434> TEST_REDIS_URL=redis://127.0.0.1:16379/0
`go test -race ./... -count=1`48372 exit0, vet/diff0. Actual NewRouter/bcrypt/JWT,
PostgreSQL user/API key/ledger/receipt, network Redis VerifySession with unique
tracked prefix/no flush. Owner create/replay; foreign404; missing/bad/expired JWT
401; real1ms Redis TTL observed within bounded1s then401; logout401 even with
receipt. Denied calls leave1ledger/1receipt. Valid read key authenticates via
service, ungranted scope denied/expired/revoked invalid; financial write route
denies even valid key401 (JWT-only), NOT API-key write support. Full suite also
executes Redis audit stream proof. Fixture owners/Redis namespace0/schema32false.
No auth policy/prod/migration/install/release change. Guide/OpenAPI/docs updated;
transactions:write design pending, native outbox/full goal open. Auth ticket links
exact test and policy evidence.

Keyed-create HTTP transitional gate:56514 validRED3cases (metadata absent,
different pairs from same key, invalid key accepted).38145 focusedpass; final
4HTTP/realDB cases after strengthening pass in full Go race45501, vet0. Existing
auth context source, body owner_id cannot select owner; not full middleware E2E.
Same local55434 DSN/testcommand template, focused selector
`^TestOffline(Create|Transfer)HTTP`; full `go test -race ./... -count=1`.
Proof: first201/replay200; same implicit date/ID/exclusion after archive/delete;
changed typed body409;8 concurrent keyed transfers exactly2excludedrows; explicit
invalid/duplicate keys400; foreign404/unauth401; legacy meta:null; missing durable
capability503 with no fallback write. Fixture owners0/schema32false readback.
Swagger generation exit0 and inspected key/header/envelope/200/201/409/503 schemas.
Native full189/189units0skip (terminal70137 exit0,
`/tmp/mypocket-receipt-envelope-native-final-20261014.xcresult`) includes actual
JSONDecoder envelope object+2-row transfer array with meta. This is compatibility
characterization, no native product fix/live offline UI proof. Public human+agent
Markdown/API/spec/plan reconciled; no migration/API restart/prod/sign/install/
release. Key-required rollout/middleware/capability/outbox/cache/sync remain open.

SQL privacy prerequisite: original NewPostgres2realDB tests RED (parameters and
driver error echo); first ParamsFilter-only wrapper remainedRED due GORM Scan
recorder expansion. Attempt2 class-only Trace + sanitized errors/generic events
passes4actual read-only PostgreSQL tests in internal/infrastructure/db/postgres_test.go:
Scan values unchanged/no log payload, cast error unchanged to caller/logSQLSTATE
only after LogMode, Row/generic events redacted, not-found sentinel/rows retained.
Focused `go test ./internal/infrastructure/db -run '^TestPostgresSQLLog' -count=1`
passes; final `TEST_DATABASE_URL=<local55434/mypocket_feedback> go test -race ./... -count=1`
6719 exit0 across all packages, `go vet ./...`/diff-check pass. No migrations/writes/
API restart/prod activation/log deletion. Scope shared SQL boundary only, not all
application logging/HTTP/native offline acceptance. ADR019/runbook/API/parent
ledger reconciled; Swagger/native UI unchanged because no route/UI contracts.

Offline backend prerequisite2026-10-14:80988 runtimeRED3/3 missing create capability;
66476 real replayRED candidate primary-key query bug. Fresh persisted receipt
lookup fixes it;5finaltests cover ordinary replay/conflict/owner,8concurrentcalls/
rollback/failure retry, pair identity/exclusion, invalid keys/foreign wallet and
shared patch/delete after deletion without resurrection. Focused race73167 exit0,
full Go86668 exit0, final full Go race85540 exit0, vet0. Commands with local-only
TEST_DATABASE_URL to dedicated55434/mypocket_feedback: focused selector
`go test -race ./internal/infrastructure/repository -run '^TestOffline.*Receipt' -count=1`,
`go test ./... -count=1`, `go test -race ./... -count=1`, `go vet ./...`.
Post-test readback fixture owners0, schema32 dirtyfalse. No prod data mutation/
migration/API restart, no native/offline/HTTP replay or live-provider acceptance.
Docs/API status/ADR018 reconciled; Swagger unchanged (no handler contract change).
Logging privacy and actual HTTP integration remain gates. See receipt work ticket.

Portfolio metric density RED66252 exit65 at
`/tmp/mypocket-portfolio-metric-density-red-20261014.xcresult`: actual asset cost
line27.5pt > summary compact20.5pt+1, valid typography mismatch, not compile error.
Final71961 exit0 at `/tmp/mypocket-portfolio-metric-density-green-20261014.xcresult`
191/191 PASS0skip (188units+3API UI). Same LOCAL_API_UI/base8081 xcodebuild
template with full MyPocketIOSTests and selectors
`testAPIBackedPortfolioPositionUsesSameCompactMetricScaleAsSummary`,
`testAPIBackedPortfolioEditorsPreserveDraftOnFailureAndCancelWithoutWrites`,
`testAPIBackedPortfolioAssetCreateAndPriceRefresh`. Actual unused fixture asset
via owner API/teardown; no trade fixture. Price readback120000, duplicate-create
error retains draft/cancel unchanged IDs, success/reset tested.
ExportedF514F91B-2EE3-4E36-8ACB-BA88954F66EC.png under
`/tmp/mypocket-portfolio-metric-density-capture-20261014` inspected: compact metric
scale, price/history reachable; Feedback covers asset identity/quantity. Visual
acceptance remains FAIL for known avoidance-design blocker, no fourth guess fix.
Guard/guardtests/project/diff pass. No hidden financial fields, API/schema/auth/
palette/base-font changes. Native/public Markdown reconciled. No deployment,
signature/install/release or full accessibility/landscape acceptance claim.

Advisor single-entry/draft: RED7177 duplicate menu; RED96308 switches tabs and
returns placeholder instead of question. After visible menu removal + volatile
shell input binding, gate84341 was195/196, not all green: test helper pressed
Return adding newline. Removed helper call, direct tab switch preserves exact
literal question. Final41071 terminal exit0 at
`/tmp/mypocket-assistant-navigation-final-20261014.xcresult`196/196 PASS0skip,
188unit+8UI. Same xcodebuild LOCAL_API_UI/base8081 template, full MyPocketIOSTests
plus selectors `testAssistantTabIsSolePrimaryEntryAndRetainsUnsentDraftAcrossTabs`,
`testMoreDoesNotDuplicateProfileAndOverviewNotificationDestinations`,
`testAuthenticatedMoreMenuOpensEveryDestination`,
`testRemainingScrollableSecondaryRoutesReserveShellInset`,
`testAPIBackedAssistantLoadsOverview`,
`testAssistantShowsSuggestedQuestionsBeforeFirstMessage`,
`testAssistantComposerIsCompactAndReachableWithKeyboard`,
`testAssistantFeedbackBubbleDoesNotCoverComposerAndDraftSurvivesDisabledSend`.
Summary inspected; exported retained-draft capture
`/tmp/mypocket-assistant-retained-draft-20261014/4EDDCDE0-6E36-4964-9F81-DA055FB6CA82.png`
inspected. Fake auth fixture error expected; no actual provider response proof.
No durable restart-draft promise. Guard/guardtests/project/diff pass. API/schema/
auth unchanged; native/public Markdown reconciled, release withheld. Existing
bubble visual failure and remaining full goal gates are not closed by this suite.

Independent current management regression:55284 terminal exit0,
`/tmp/mypocket-management-current-regression-20261014.xcresult`7/7 PASS,0skips,
summary and all7 individual test nodes inspected. Exact run uses `rtk proxy
xcodebuild test -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS
-destination 'platform=iOS Simulator,id=182D41A2-4FC6-445B-A592-70FDCB5C662E'
-derivedDataPath /tmp/mypocket-feedback-resolution-tests -resultBundlePath
/tmp/mypocket-management-current-regression-20261014.xcresult
-collect-test-diagnostics never 'SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI'
MP_API_BASE_URL=http://127.0.0.1:8081/api/v1 SWIFT_TREAT_WARNINGS_AS_ERRORS=YES -quiet`
plus one `-only-testing:MyPocketIOSUITests/LaunchTests/` argument for each:
`testAPIBackedBudgetCreateUpdateDeleteFromNativeUI`,
`testAPIBackedEndedBudgetPreviewHasNoWriteControlsAndDeleteCancelPreservesData`,
`testAPIBackedJarCreateUpdateRemoveFromNativeUI`,
`testAPIBackedRecurringCreateUpdateDeleteFromNativeUI`,
`testAPIBackedTravelOptionalDatesUseCalendarAndPreserveNulls`,
`testAPIBackedTravelCreateToggleUpdateDeleteFromNativeUI`,
`testAPIBackedTravelLinkAndUnlinkTransactionFromNativeUI`.
Disposable API health200 at `/api/v1/health`; plain `/health`404 was an incorrect
probe path, not an API outage. Source reviewed: real writes/readbacks and canceled
draft retention; selected run does not prove every error/accessibility/visual
state. Static base guard, guard tests, project verification and diff-check pass.

AI-entry pinning third pass: `xcrun xcresulttool get test-results summary --path
/tmp/mypocket-floating-complete-controls2-20261014.xcresult --format json` confirms
terminal Passed196/196 (188units+8UI),0failed/0skip. Same LOCAL_API_UI/base8081
xcodebuild template; selectors header/analyze, Save-all, common management routes,
portfolio error/cancel/save, real selected-photo append, proposal explicit type,
Assistant draft/bubble, uncertain GET recovery. Earlier195/195 did not cover
thumbnail/approval; stronger assertions failed0/2 before third pass.
Export `xcrun xcresulttool export attachments --path
/tmp/mypocket-floating-complete-controls2-20261014.xcresult --output-path
/tmp/mypocket-floating-final-visual-20261014 --filter '*.png'` inspected:
26FBDC5B header/footer,8E34D9EC proposal,43ADBC44 actual imported-photo PNGs.
Header/Analyze/preview/remove/approval protected, but proposal bubble covers the
description placeholder. Visual FAIL; slice blocked for design review under
third-pass loop guard. Tests do not register every input/text region. No final
all-screen/physical acceptance, signing/install/prod/release claim.

Task4 selected-photo/batch: `/tmp/mypocket-photos-import-final-20261014.xcresult`
terminal14719 exit0,193/193 PASS,0 skipped (187units+6UI). Native selectors:
`testAIEntryAttachmentSourcesDefaultToPhotosAndCancelPreservesFiles`,
`testAIEntrySystemPhotosAndFilesCancelPreserveExistingAttachments`,
`testAIEntryConfiguredAttachmentLimitAndThumbnails`,
`testAIEntrySelectingSystemPhotoAppendsWithoutReplacingExistingFiles`,
`testAIEntryUncertainRequestRequiresGETResolutionBeforeResubmission`,
`testAPIBackedAIEntryDisablesAttachmentsWhenFilesUnavailable`.
Same LOCAL_API_UI/base8081 command below. Real system stock Photos image appended,
19 prior names preserved, count20 disables add; screenshot
`/tmp/mypocket-selected-photo-capture-20261014/178F9773-87A2-4252-92CC-FEC5F2275515.png`
inspected. Batch pipeline187unit proof (standalone exit0 bundle
`/tmp/mypocket-photo-batch-contract-20261014.xcresult`):2valid/2errors within4
remaining, order0…3, one active load, cap0 no data load. Multipart real adapter
with transport capture verifies JPEG/PDF raw bytes, headers/session, Vietnamese
instruction/timezone/request_id, no owner_id/base64. No actual OCR/model/ledger
write; unavailable media is simulated at the data boundary, not live iCloud.
Hierarchy capture first compile failed test variable `add` shadowing XCTest add;
qualified self.add, rerun `/tmp/mypocket-system-photo-hierarchy2-20261014.xcresult`
passes actual system cancellation and provides observed PXGGridLayout-Info target.
Unsigned Release compile46172 terminal0 for prod API URL; not signed/installable.
Guard/guard tests/project/diff exit0. Task-done repeat75502 terminal0,
`/tmp/mypocket-task4-complete-20261014.xcresult`193/193 PASS,0 skip; ledger Task4
complete under verified-worktree/deferred-commit ruling. No live handles.
Task6 gaps: AI entry header scrolls out and Feedback partly covers process CTA;
current screenshots prove import behavior, not whole-screen UI acceptance.
Pinning RED: `/tmp/mypocket-ai-entry-pinned-red-20261014.xcresult`, terminal61061
exit65,0/2 fail as intended. Selectors `testAIEntryHeaderAndAnalyzeRemainPinnedWhileAttachmentsScroll`
and `testAIEntrySaveAllStaysReachableAfterProposalScroll`, same LOCAL_API_UI/base8081
command template. Analyze initially below20 files, header/Save-all moves offscreen
and bubble covers Analyze at bottom. Last-X collision assertion added after this
compiled run, pending. No implementation change yet; this is the next TDD step,
not a failed193/193 attachment gate or a whole-suite green claim.

Photos/credit final: `/tmp/mypocket-photos-footer-bounds-20261014.xcresult`,
terminal53512 exit0,189/189 PASS,0 skipped (185 units+4UI). Same common unit/UI
command below with LOCAL_API_UI/base8081, selectors:
`testAIEntryAttachmentSourcesDefaultToPhotosAndCancelPreservesFiles`,
`testAIEntrySystemPhotosAndFilesCancelPreserveExistingAttachments`,
`testAIEntryConfiguredAttachmentLimitAndThumbnails`,
`testAPIBackedCreditEntryAndPaymentRefreshStatementFromNativeUI`.
Source originalRED71722 Files direct; code gate186/187 source tab failed.
Wait98111 disproved timing; button-host25465 remained187/189. Explicit zero-height
absent footer fixes body rendering/hit-testing; final actual Photos/Files open and
Cancel preserves19 files, tab/default/capacity assert,20-cap removal proof pass.
Image units: exact JPEG/PNG bytes, TIFF→JPEG signature, invalid/25MiB source
rejection,5MiB accepted/one-byte-over rejected. Existing multipart/session tests
run in full suite. This is not actual selected-photo/mixed/cloud failure/OCR proof.
Source screenshot `/tmp/mypocket-photos-final-captures-20261014/F37BD999-F4D6-4E45-B1F0-40A4620B2063.png`
and system Photos/Files screenshots in `/tmp/mypocket-photos-system-captures-20261014/`
inspected. Credit payment title unitRED46341 Nhận chuyển ví→Thanh toán; API UI
verifies12000 purchase/category,7500 payment/pair/report-exclusion, inspected
`/tmp/mypocket-credit-row-green-capture-20261014/FE5A69FB-D023-4D22-8339-2062C43C596D.png`.
Prior portfolio/credit regression14223 was4/5, not all green; portfolio3 and
credit-route pass, credit mutation subsequently recovered in final189/189.
Guard/guard tests/project/diff exit0; native/public human+agent-readable Markdown
updated. No public API/schema/auth/provider changes, so reference/ERD unchanged.
No signing/install/prod change or release publication. Task3/4/full goal stay open.

Final retention gate: `/tmp/mypocket-draft-retention-final-20261014.xcresult`,
terminal20421 exit0,188/188 PASS,0 skipped (180 units+8 UI). Selectors: Budget
CRUD/cancel/reopen; Jar create/update/remove/cancel/reopen/month-switch; Recurring
CRUD/cancel/reopen/interval/run-due; Travel CRUD/cancel/reopen; ended-budget
preview/delete-cancel; Travel optional-date/context clear; common management
route/footer; chat Feedback/draft. Same LOCAL_API_UI/base8081 command below.
Jar trueRED `/tmp/mypocket-jar-month-red-20261014.xcresult` retained Discarded
month draft after previous/current switch; month-change reset resolves it. Final
ended-budget check compares full decoded transactions (IDs, amounts, scopes,
notes, dates and flags), not IDs alone. Guards/guard tests/project/diff exit0.
No backend/schema/auth contract change; API docs unchanged for this input-state
slice. Native human+agent-readable screen docs reconciled. No release publication,
production rollout or device install. Task3 full portfolio/credit/error/capture
scope and broader goal remain open;188 narrow proofs do not mean all functions.

Canceled-edit retention: `/tmp/mypocket-edit-draft-red-20261014.xcresult`
0/4 FAIL as intended; all4 management editors discarded canceled input on reopen.
Command uses LOCAL_API_UI unit/UI template below with selectors Budget CRUD,
Jar create/update/remove, Recurring CRUD and Travel CRUD. Added explicit API
readbacks that cancellation does not persist input. Resume guards implemented;
`/tmp/mypocket-draft-month-scope-20261014.xcresult` running72911, also180units,
new ended-budget read-only/delete-cancel ledger proof and jar month-switch case.
Do not claim GREEN or month isolation from source alone.
Terminal72911:184 passed/1 failed,0 skipped. All180 units pass; Budget,
Recurring, Travel cancel/reopen pass, ended-budget read-only/delete-cancel pass.
Screenshot `/tmp/mypocket-ended-budget-capture-20261014/507BAF08-414E-4469-A254-E91AFA546338.png`
inspected: read-only details, no input/Save, Delete reachable, bubble above footer.
Jar test failed before selection: offscreen month absent in virtualized native
menu. Test-only bounded CollectionView scrolling added; focused run86474
`/tmp/mypocket-jar-month-red-20261014.xcresult` pending. Full financial-field
comparison for ended-budget ledger strengthened after that compiled run; final
gate must execute the strengthened check too.

Management continuation: same xcodebuild command below with LOCAL_API_UI/base8081.
`/tmp/mypocket-management-api-final-20261014.xcresult`:185 passed/1 failed
(Recurring update interaction accidentally opened Feedback). Recovery
`/tmp/mypocket-management-recovery-green-20261014.xcresult`:183 passed/1 failed
(Travel description received predictive-keyboard swipe text);180 units pass,
recurring CRUD, typing-bubble restore and chat-draft tests pass. Neither suite
is wholly green despite the historical recovery filename. Optional-date create
and clear passed in the first run. Wire regression verifies explicit empty-string
clearing without changing backend partial-update semantics.
`/tmp/mypocket-travel-clear-recovery-20261014.xcresult`:Travel helper failed after
a second normalized-coordinate tap returned to list before deletion. Test-only
selection-menu correction running in `/tmp/mypocket-travel-selection-20261014.xcresult`;
terminal exit0,1/1 PASS,0 skipped: create/context/date and clear/context/date
readback all pass. Local synthetic data only. Aggregate recovery running at
`/tmp/mypocket-management-aggregate-20261014.xcresult`:terminal exit0,
187/187 PASS,0 skipped (180 units +7 UI): management route/footer, Budget CRUD,
Recurring CRUD, Travel CRUD, Travel context/date create/clear, keyboard bubble
hide/restore, chat draft. Jar's misspelled selector did not count as coverage;
explicit `testAPIBackedJarCreateUpdateRemoveFromNativeUI` follow-up running in
`/tmp/mypocket-jar-explicit-20261014.xcresult` terminal53497 exit0,1/1 PASS,
0 skipped (create/edit/remove UI). Guard, guard tests,
project verification and diff-check all exit0. No full-screen UAT/signing/prod claim.

UI command: `rtk proxy xcodebuild test -project mobile-ios/MyPocketIOS.xcodeproj
-scheme MyPocketIOS -destination 'platform=iOS Simulator,id=182D41A2-4FC6-445B-A592-70FDCB5C662E'
-derivedDataPath /tmp/mypocket-feedback-resolution-tests -collect-test-diagnostics never
SWIFT_TREAT_WARNINGS_AS_ERRORS=YES -quiet`, with unique result bundle and selectors below.

- RED `/tmp/mypocket-chat-compact-red-20261014.xcresult`: input153.5pt >76.
- GREEN `/tmp/mypocket-chat-feedback-green-20261014.xcresult`:3/3 selectors
  `MyPocketIOSUITests/LaunchTests/testAssistantComposerIsCompactAndReachableWithKeyboard`,
  `testAssistantFeedbackBubbleDoesNotCoverComposerAndDraftSurvivesDisabledSend`,
  `testAssistantShowsSuggestedQuestionsBeforeFirstMessage`. Bubble collision separately
  reproduced before fix in `/tmp/mypocket-chat-feedback-red-20261014.xcresult`.
- `/tmp/mypocket-chat-api-and-unit-20261014.xcresult`:172/172 (171 units +
  `MyPocketIOSUITests/LaunchTests/testAPIBackedAssistantLoadsOverview`), with
  `SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI`
  `MP_API_BASE_URL=http://127.0.0.1:8081/api/v1`. Real disposable DB/API, no prod writes.
  Tests scroll overview, open/close details. Exported pinned/keyboard/multiline
  screenshots inspected in `/tmp/mypocket-chat-api-attachments-20261014` and
  `/tmp/mypocket-chat-feedback-green-attachments-20261014`.
- First GREEN compile failed wrong shared-sheet signature; corrected. Next
  keyboard assertion exposed provider-disabled input; approved UX allows local
  draft, submission still disabled. No provider chat success inferred from fixture UI.
- Card insets RED116 vs124pt in `/tmp/mypocket-card-insets-red-20261014.xcresult`;
  nested metrics RED215pt vs flat reference183pt. Arbitrary160pt ceiling rejected
  rather than shrinking fonts. Revised reference test RED in
  `/tmp/mypocket-flat-reference-red-20261014.xcresult`; GREEN full native173/173
  `/tmp/mypocket-flat-reference-green-20261014.xcresult`. Base guard/guard tests,
  project verification and diff whitespace checks exit0.

Remaining proof: provider-backed history/cancel/error, accessibility/landscape,
all other registered screens, physical iPhone acceptance and final signed build.
No full goal completion or release claim.

Note editor: same UI command above, selectors
`MyPocketIOSUITests/LaunchTests/testTransactionNoteHasDedicatedEditorAndCancelDoesNotChangeDraft`
and `testAddTransactionUsesStitchFormBaseControls`. RED old note row only expanded
details in `/tmp/mypocket-note-editor-red-20261014.xcresult`; GREEN2/2 in
`/tmp/mypocket-note-editor-green-20261014.xcresult` proves multi-line save,
discard on close, no duplicate details input, existing controls. API-backed
create/update and DB note re-read on iPhone16 passes in
`/tmp/mypocket-note-api-green-20261014.xcresult`; that combined run is overall
FAIL (3 AuthRestore tests read residual device metadata). Do not call it a green
suite. Failures: expired access+refresh gate, refresh rejection invalidation,
timezone/session synchronization. Test isolation fixed via per-test preferences
and explicit unenrolled device fixture, preserving actual simulator Keychain.
First isolation run passes but emits XCTest nonisolated lifecycle warnings;
replaced setup/teardown overrides with actor-local lazy suite + per-test defer.
Final full unit173/173, zero warnings in output:
`/tmp/mypocket-auth-isolation-green2-20261014.xcresult`, same iPhone16 and local API
compile conditions. No auth implementation/contract changed for this harness fix.

Release compile: `rtk proxy xcodebuild build -project mobile-ios/MyPocketIOS.xcodeproj
-scheme MyPocketIOS -configuration Release -destination 'generic/platform=iOS'
-derivedDataPath /tmp/mypocket-owner-device-20261014 CODE_SIGNING_ALLOWED=NO
MP_API_BASE_URL=https://money.dungxbuif.com/api/v1 SWIFT_TREAT_WARNINGS_AS_ERRORS=YES -quiet`
exits0. Known packaging warning: all interface orientations must be supported
unless app requires full screen. This is an unsigned compile verification, not
a signed/installable handoff; generated product's old signature/profile removed
by Xcode. Existing installed iPhone app/data remain unchanged. API/config/ERD
public references unchanged because these slices alter UI and test isolation only.
Human guides and stable agent-readable Markdown updated; no release publication.

## Native typed fields and portfolio editors — latest execution

Same xcodebuild test base command above, `SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI`
and `MP_API_BASE_URL=http://127.0.0.1:8081/api/v1` for API runs; SE destination.

| Proof | Exact selector/result bundle | Result |
| --- | --- | --- |
| Locale quantity RED | MyPocketIOSTests/FinanceModelTests/testPortfolioQuantityAcceptsLocalizedDecimalWithoutAcceptingUnicodeNumerals; `/tmp/mypocket-number-locale-red-20261014.xcresult` | FAIL: comma rejected, non-ASCII accepted |
| Count RED | MyPocketIOSTests/FinanceModelTests/testRecurringIntervalDoesNotUseCurrencyThousandsGrouping; `/tmp/mypocket-interval-red-20261014.xcresult` | FAIL: currency grouping accepted as1000 periods |
| Typed fields full units | MyPocketIOSTests; `/tmp/mypocket-typed-fields-green-20261014.xcresult` |178/178 PASS |
| Fractional/count persistence | LaunchTests/testAPIBackedPortfolioTradeRefreshesPositionAndHistory + testAPIBackedRecurringCreateUpdateDeleteFromNativeUI; `/tmp/mypocket-typed-fields-api-20261014.xcresult` |2/2 PASS; `1,25`→`1.25`, interval2 read back |
| Dedicated editor RED | LaunchTests/testAPIBackedPortfolioEditorsPreserveDraftOnFailureAndCancelWithoutWrites; `/tmp/mypocket-portfolio-editor-red-20261014.xcresult` | FAIL: inline fields/missing editor entry |
| Asset error/cancel/create and trade | Same editor selector + testAPIBackedPortfolioTradeRefreshesPositionAndHistory + units; `/tmp/mypocket-portfolio-editors-full-20261014.xcresult` |180 PASS/1 FAIL overall; price test used wrong implicit header ID |
| Price and report exclusion | MyPocketIOSTests + LaunchTests/testAPIBackedPortfolioAssetCreateAndPriceRefresh + testAPIBackedReportsTrendExposesPersistedIncomeAndExpenseValues; `/tmp/mypocket-portfolio-price-green2-20261014.xcresult` |180/180 PASS; price120000 API readback |
| Editor screenshots/keyboard | Editor failure/cancel/create + fractional trade selectors; `/tmp/mypocket-portfolio-editor-captures-20261014.xcresult` |2/2 PASS, images inspected |
| Note save then outer cancel | LaunchTests/testAPIBackedTransactionCancelDoesNotPersistNoteDraft; `/tmp/mypocket-transaction-cancel-api-20261014.xcresult` |1/1 PASS, DB transaction IDs unchanged |
| Final Task2 full unit gate | MyPocketIOSTests; `/tmp/mypocket-task2-final-money-20261014.xcresult` |179/179 PASS including real money input1234567→display1.234.567, unchanged binding |

UIKit disabled probe first failed because helper mounted no window (0 inputs),
not a product RED. Real window characterization passes with original disabled
propagation; speculative forwarding removed. No auth/API/schema change.

Screenshots inspected: asset retry error
`/tmp/mypocket-portfolio-editor-error-capture-20261014/8EA717E2-002D-4BFC-838D-EA3CFCBBD792.png`,
trade keyboard/footer
`/tmp/mypocket-portfolio-trade-keyboard-capture-20261014/8E6468D6-154D-4564-87AD-438A49063803.png`.
Populated flat list inspected; its long metrics still need Task6 density review,
not blanket visual acceptance. Other management editors/accessibility remain open.

Exact synthetic trade+asset IDs were verified then removed from isolated local
DB only with owner+asset predicates (recorded in plan ledger); count of named trade
fixtures now0. Public append-only cleanup cannot delete trade-bearing assets, so
do not infer cleanup from ignored HTTP teardown errors. No prod data changed.
Release command above rerun after new source, unsigned build exits0. Base/guard
tests, project verification and diff check pass; physical signing/install and
release acceptance remain pending. Human+agent Markdown guides reconciled;
API/ERD unchanged because input normalization and presentation keep wire types.

## Native feedback + device auth — earlier 2026-10-14 checkpoint

Latest owner-resolution task has independent evidence in
[FEEDBACK-OWNER-RESOLUTION](tickets/FEEDBACK-OWNER-RESOLUTION.md): backend full
race current+candidate pass, native171/171 after stale-refresh fix, API-backed
resolve/reopen UI1/1. Duplicate More entries UI RED→GREEN1/1. Tests and screenshot
paths, initial interrupted/compile failures, backup/migration32 and pending
SSH/Keychain activation/install are recorded there. Earlier density test-skip
applies to that historical immediate pass, not a blanket waiver for new API.
Source/native screen registry and guideline draft are documentation/audit output;
they do not establish visual acceptance of every screen/presentation state.

Urgent density follow-up: owner explicitly waived test suite execution. Card8,
control44, no outer transaction/wallet row padding, single profile row and smaller
section gaps now supersede sizes validated below. Geometry assertions updated,
**not run**. Prior173/8 results do not prove this new pass. Release build/signature/
install outcome to be recorded; owner on-device visual review remains required.

Density pass delivery: final `xcodebuild build` Release/generic iOS with
SWIFT_TREAT_WARNINGS_AS_ERRORS=YES, Personal Team AU442S3MN3 and production API
exited0 (existing interface-orientation packaging warning remains). Strict/deep
codesign verification passed; Info.plist API URL is production. devicectl install
exited0 on iPhone15 Plus, new bundle container84D5A542-95FE-4C2E-AEB5-FB52AEC6D6D9,
without uninstalling. Diff whitespace check passes. No unit/UI tests executed
for this urgent pass by explicit owner request; do not reuse earlier pass counts
as this revision's test evidence. Physical appearance acceptance pending.

This entry supersedes older dock/intake/signing claims below, not their recorded
historical failures. Work remains in review, not fully accepted/released.

- Native compact/shell artifact `/tmp/mypocket-native-final-compact-se-20261014.xcresult`:
  **173/173 pass** (169 unit + 4 UI), zero failures/skips/runtime warnings,
  iPhone SE3/iOS18.5. Covers compact card/row/profile, Dynamic Type, wallet header,
  overlay/no-dock viewport and final More/category action reachability.
- API-backed artifact `/tmp/mypocket-native-final-api-ui-20261014.xcresult`:
  **8/8 pass**, zero failures/skips/runtime warnings, iPhone16/iOS18.5,
  disposable local API8081/PostgreSQL55434/Redis16379. Includes enrollment →
  logout → protected Keychain login, row consistency across screens, recalculated
  month/report, wallet scope, key copy/revoke and feedback capture/text fallback.
- Fresh persisted-result recheck:
  `rtk proxy xcrun xcresulttool get test-results summary --path /tmp/mypocket-native-final-compact-se-20261014.xcresult`
  and the same command for the API UI artifact return Passed with these counts.
  These read existing results; they are not a second test execution.
- RED artifacts: `/tmp/mypocket-compact-overlay-red-20261013.xcresult` caught
  oversized card/row and reserved lane;
  `/tmp/mypocket-feedback-attachment-profile-red-20261013.xcresult` caught
  large profile/missing attachment action. Face ID UI2 exposed overscrolling
  NavigationStack viewport; explicit route height/sibling navbar fixed it.
  UI3 exposed native label boundary tolerance10.33pt vs8pt; final runs pass
  two-sided viewport/action assertions. Earlier interrupted/failing runs are
  not claimed successful or hardware proof.
- Inspected compact screenshot:
  `/tmp/mypocket-native-final-compact-se-attachments-20261014/086ABC26-B353-43DD-8114-D944507A6311.png`.
  Ledger screenshot:
  `/tmp/mypocket-native-final-api-ui-attachments-20261014/A356F732-FC34-46BB-AABE-BA8EBF8530D1.png`.
  A dim disabled Account add circle still overlaps last-row presentation;
  keep this visual follow-up open.
- Fresh checks: `rtk proxy bash mobile-ios/scripts/check-base-components.sh`,
  `rtk proxy bash mobile-ios/scripts/check-base-components.test.sh`, and
  `rtk git diff --check`: pass; previous EOF whitespace removed.
- Backend independent review reproduced legacy-family/delayed-issuance races;
  deterministic TTL-aware regressions and `go test -race ./... -count=1` passed
  against disposable PostgreSQL, including isolated auth-only candidate.
- Production candidate `12572c87` pushed/deployed after approved backup;
  migrations29→31 dirty=false, unchanged financial counts/checksums. Health200,
  API1/1/worker1/1/web0/0; four auth paths in Swagger, invalid device401.
  No agent financial writes/password reset. [Rollout](tickets/NATIVE-DEVICE-AUTH-ROLLOUT-20261014.md).
- Authenticated feedback API returned34 open records, all without screenshots.
  [Per-record audit](tickets/NATIVE-FEEDBACK-PROD-AUDIT-20261014.md) separates fixes,
  existing coverage and pending work; no production status updates.
- Device installation: after owner unlock, original bundle failed strict
  verification (missing sealed MyPocketIOS.cstemp from interrupted signing).
  Same-identity re-sign with preserved identifier/entitlements/flags exited0;
  `rtk proxy codesign --verify --deep --strict --verbose=2 /tmp/mypocket-owner-device-20261014/Build/Products/Release-iphoneos/MyPocketIOS.app`
  passed. Info.plist confirms production API URL. `rtk proxy xcrun devicectl device install app --device 00008120-001C28C02131A01E /tmp/mypocket-owner-device-20261014/Build/Products/Release-iphoneos/MyPocketIOS.app`
  exited0 (com.mypocket.ios installed), and `rtk proxy xcrun devicectl device process launch --device 00008120-001C28C02131A01E com.mypocket.ios`
  exited0/reported launched. No uninstall; existing data/Keychain retained.
- Pending: physical biometric/visual owner acceptance. Broad orientations/dark/largest accessibility UI, live
  OCR/provider and manual photo selection are not fully covered by these runs.
  Note editor, offline/cache, AI V2/RAG and APNs remain open. No all-features/
  all-feedback completion or release publication claim.
- Docs reconciled: auth API/OpenAPI/ADR017/public Face ID+feedback Markdown,
  native architecture/platform/tokens/bases and work/context/backlog. UI-only
  changes reuse financial/feedback APIs; no financial schema change needed.

## Native button-radius refinement — 2026-10-13

- Scope: owner clarified button rounding only. Shared MPButton radius is 8pt;
  card radius, circular feedback actions and 48pt/44pt control/hit geometry
  remain unchanged by this refinement. UI/UX Pro Max guidance used to retain
  accessible iOS touch targets, not to choose the owner's visual radius.
- RED: token regression failed with actual 999 versus expected 8 at
  `/tmp/mypocket-button-radius-red-20261013.xcresult`.
- GREEN: full native unit target 164/164, zero failures/skips, warnings-as-errors,
  iPhone SE (3rd generation), iOS 18.5:
  `/tmp/mypocket-button-radius-green-20261013.xcresult`.
- UI: auth launch 1/1 passes with local API configuration at
  `/tmp/mypocket-button-radius-ui-20261013.xcresult`. Rendered 375pt-wide
  portrait inspected: `/tmp/mypocket-button-radius-se-20261013.png`; button
  shows reduced corners without changed input/card geometry. This does not
  establish every screen/orientation/Dynamic Type/dark-mode acceptance.
- Pending: scroll-to-navbar bubble-overlay request, card/content density,
  full production feedback queue verification and owner visual acceptance.
  Current production feedback read credential returned 401 TOKEN_INVALID in
  the preceding audit; do not claim the queue has been exhausted.
- Separate Face ID enrollment UI attempt failed in password-field interaction
  and xcodebuild stalled after the failure; interrupted by agent. Result path:
  `/tmp/mypocket-native-face-id-ui-20261013.xcresult`. No end-to-end biometric
  success is claimed from this attempt. Unit auth proof is not hardware proof.
- Global diff check reports a pre-existing extra EOF blank line in
  `OverviewView.swift:252`; button-refinement files are checked separately.
- Public native token/base contracts and internal work/context/backlog updated;
  API/config/schema docs need no change for button shape. Release notes remain
  deferred at owner's explicit request; no production deployment/device install.

## Native mobile feedback follow-up — 2026-10-13

- PASS: transaction-row subtitle policy regression and wallet-balance privacy
  mask policy tests pass 2/2 on iPhone 16 Simulator (iOS 18.5) in
  `/tmp/mypocket-ios-feedback-mask-green/Logs/Test/Test-MyPocketIOS-2026.10.13_21-23-37-+0700.xcresult`.
- Behavior: the ledger subtitle shows the note when present, otherwise a local
  short date; it does not repeat the current wallet name. The shell's existing
  privacy toggle now reaches native financial screens through a shared SwiftUI
  environment. API/data calculations remain unchanged; edit fields stay
  readable for entering drafts.
- Report inclusion uses the shared checkbox base, defaults on for a new
  transaction, and restores the persisted transaction value while editing.
- PASS: month period presentation regression passes 1/1 on iPhone 16 Simulator
  (iOS 18.5) at
  `/tmp/mypocket-ios-feedback-month-green2-20261013/Logs/Test/Test-MyPocketIOS-2026.10.13_21-36-53-+0700.xcresult`;
  UI now presents inclusive local dates while server queries stay half-open.
- PASS: ledger newest-week fallback policy regression passed 1/1 on iPhone 16
  Simulator at
  `/tmp/mypocket-ios-feedback-week-green-20261013/Logs/Test/Test-MyPocketIOS-2026.10.13_21-43-01-+0700.xcresult`.
- PASS: transaction editor report-checkbox UI regression passed 1/1 at
  `/tmp/mypocket-ios-feedback-checkbox-ui-20261013/Logs/Test/Test-MyPocketIOS-2026.10.13_21-37-31-+0700.xcresult`;
  privacy-mask accessibility UI regression passed 1/1 at
  `/tmp/mypocket-ios-feedback-mask-ui-final2-20261013/Logs/Test/Test-MyPocketIOS-2026.10.13_21-41-04-+0700.xcresult`.
- PASS: segmented-control Reduce Motion policy and elevated Assistant surface
  policy regressions pass; full native unit/contract suite is 156/156 with zero
  failures/skips at
  `/tmp/mypocket-ios-feedback-unit-final7-20261013.xcresult`.
- PASS: wallet-scope preference policy regression passes 1/1 at
  `/tmp/mypocket-ios-feedback-scope-green-20261013/Logs/Test/Test-MyPocketIOS-2026.10.13_21-52-40-+0700.xcresult`;
  preference is owner-keyed, device-local, and aggregate scope clears it.
- Fresh final verification after documentation reconciliation: project
  verification, shared base-component guard/tests, `git diff --check`, and
  Release warnings-as-errors build at
  `/tmp/mypocket-ios-feedback-release-final4-20261013` pass. The full unit/
  contract target passes 156/156 with zero failures/skips at
  `/tmp/mypocket-ios-feedback-unit-final7-20261013.xcresult`.
- INCOMPLETE: the broad unit+UI test scheme sat on the More screen for over ten
  minutes without new runner output and was interrupted (exit 130). This is not
  a pass; unit/contract and focused UI suites are being run separately.
- Pending: full UI suite (broad run stalled; focused checkbox and privacy UI
  cases pass), category connector / typography / density / bottom-navigation
  visual acceptance, and owner decision on Face ID after explicit logout. The
  current simulator is at the Home Screen, not an authenticated app route.
  Production feedback queue is not confirmed complete because its direct route
  returned 404 and no agent service token is configured locally.
- No release-note or changelog update; owner approval is still required.

## Native API-key secret copy — 2026-10-13

- RED: the clipboard policy regression first failed to compile because the
  production helper `APIKeySecretClipboard` did not exist.
- PASS: full native iOS unit/contract suite is 150/150 with zero failures or
  skips at `/tmp/mypocket-api-key-copy-unit-full-20261013.xcresult`. This
  includes exact-secret clipboard writer and empty-secret no-op contracts.
- PASS: API-backed account UI regression is 1/1 at
  `/tmp/mypocket-api-key-copy-ui-final-20261013.xcresult`; it creates a key,
  verifies the one-time secret and shared copy button, taps copy, and verifies
  success confirmation. It does not read UIPasteboard across the runner/app
  boundary, which triggers iOS paste permission UI.
- PASS: native project verification, base-component guard/tests and
  `git diff --check` pass. The screen reuses `MPAssistantCard`, `MPText`,
  `MPButton`, and `MPStatusMessage`; no new base component was created.
- BLOCKED: signed iPhone build/install was attempted only after local tests.
  Xcode reports no signed-in account for Team `GYRA8D4M54` and no provisioning
  profile for `com.mypocket.ios`; nothing was installed or launched. The
  simulator Release build and unsigned device compile remain to be recorded.
- Clipboard entry is local-only with five-minute expiry, and no secret is
  logged or persisted. Backend/API/schema/config are unchanged. Changelog and
  release note are intentionally deferred until owner approval; public API docs
  need no update.

## Current native-port verification and Ad Hoc signing readiness — 2026-10-13

- Fresh native unit/contract suite: **148/148 passed**, zero failures or skips,
  on the iPhone 16 Simulator. Result bundle:
  `/tmp/mypocket-port100-resume-unit-20261013/Logs/Test/`.
- Release build passes for generic iOS Simulator and generic iOS device with
  `SWIFT_TREAT_WARNINGS_AS_ERRORS=YES`. Project verification, base-component
  guard and guard tests pass.
- Web transaction tests pass 8/8; password-login contract and all design/base
  contracts pass. `npm run build` passes design guards, docs links, TypeScript
  typecheck and Vite production build; the existing large-chunk advisory is
  emitted. `git diff --check` passes.
- Backend `go test ./... -count=1` passes with `DATABASE_URL` and
  `TEST_DATABASE_URL` unset. PostgreSQL integration tests requiring an explicit
  test DSN are skipped; production and shared local databases were not used.
- The paired iPhone 15 Plus is currently connected and unlocked with Developer
  Mode enabled. Ad Hoc signing cannot proceed: its locally installed Apple
  Development certificate expired 2025-01-12 and there are no local
  provisioning profiles. No device installation, Developer Portal mutation,
  or production request/write was attempted. A current Developer Program team
  identity and an Ad Hoc profile containing this device UDID are required.
- Authenticated owner visual/business UAT remains open. Public API docs,
  machine-readable API references and release notes are unchanged because this
  verification changed no product behavior, endpoint, schema or configuration.

## Native Reports web-layout parity — 2026-10-13

- RED: the comparison-chart policy regression first failed to compile because
  `MPReportComparisonChart` did not exist; the API-backed UI assertion also
  exposed that parent section identifiers masked child chart identifiers.
- PASS: native unit/contract XCTest is 147/147, zero failures/skips, on the
  disposable iPhone 16 simulator at
  `/tmp/mypocket-reports-unit-final-20261013.xcresult`. This includes safe
  zero/negative/imbalanced comparison-bar scaling and report section order.
- PASS: `testAPIBackedReportsExposeMoneyInsiderDetails` passes 1/1 at
  `/tmp/mypocket-reports-ui-verified-20261013.xcresult`; it verifies the
  summary chart, always-mounted trend with no mode toggle, a prior-period
  Money Insider ring and that the ring can be scrolled fully inside the clipped
  shell viewport. Captures:
  `/tmp/mypocket-reports-ui-verified-attachments/A6CBC563-D20F-48BA-B2C1-5BBFDFEAF763.png`
  (summary/trend) and
  `/tmp/mypocket-reports-ui-verified-attachments/2A2F09BE-31C8-46E8-8081-3ABCEBAEAF6E.png`
  (Money Insider ring).
- PASS: `testAPIBackedReportsTrendExposesPersistedIncomeAndExpenseValues`
  passes 1/1 at
  `/tmp/mypocket-reports-trend-api-ui-20261013.xcresult`; account-local
  included income/expense values are exposed while the excluded 888,888 VND
  row is omitted. The test no longer taps a trend-mode control.
- PASS: Release simulator build and generic iOS device arm64 build complete
  with Swift warnings-as-errors; `verify-project.sh`, both base-component
  guards/tests and `git diff --check` pass. The generic-device build emits an
  existing orientation-support advisory unrelated to Swift compilation.
- Isolation: only the fresh local `mypocket_report_native_20261013` database,
  loopback API `127.0.0.1:8081` and Redis DB 14 were used. Wallet and
  transaction fixture counts were zero after test teardown; the API was
  stopped, the exact database dropped, Redis DB 14 returned to zero, and the
  disposable simulator was removed. Production and the user's booted simulator
  were not touched. No API/schema/migration or AI-agent reference changed;
  public API docs therefore need no update.
- This closes the current Reports layout/trend/donut parity gap only; it does
  not prove the full native port, authenticated owner visual acceptance or
  physical-iPhone installation.

## Overview week-report financial parity — 2026-10-13

- Root cause: web Overview used host-local `Date` boundaries and counted every
  non-adjustment, non-income row as expense; iOS used the account-local week
  range and excluded transfers but did not honor `included_in_reports=false`.
- Intended contract: the week card uses the saved account timezone and its
  Monday–Sunday week; only report-included income/expense rows contribute.
  Transfers, adjustments, excluded rows and instants outside that local week
  are ignored on both clients.
- RED: the web contract test failed because `buildWeekReportTotals` did not
  exist; the iOS test target failed compilation because `MPWeekReportData` did
  not exist. Both failures identified the missing production policy.
- PASS: `app/scripts/week-report.test.mjs` verifies a timezone boundary at
  Monday start, Sunday end, adjacent-week exclusion, excluded transaction,
  transfer and adjustment; `npm run test:design`, TypeScript typecheck and Vite
  production build pass.
- PASS: `testOverviewWeekReportTotalsUseAccountWeekAndExcludeNonReportRows`
  and the complete native unit/contract suite pass 148/148 on iPhone 16
  Simulator at
  `/tmp/mypocket-week-report-unit-full-20261013/Logs/Test/`.
- PASS: Release Simulator build with warnings-as-errors, Xcode project
  verification, both base-component guards/tests and `git diff --check` pass.
  Vite emits the existing large-chunk advisory; build succeeds.
- No API, schema, migration or provider contract changed, so public API docs
  and AI-agent endpoint references need no update. Physical-device install and
  owner-authenticated visual UAT remain open.

## Current-source cross-surface and production-simulator checkpoint — 2026-10-08

- iOS Debug Simulator build passes from the current worktree with
  `MP_API_BASE_URL=https://money.dungxbuif.com/api/v1`; the built bundle embeds
  that exact API URL and was installed/launched on iPhone 16 Simulator. The
  Device Hub screen shows the password-login gate. Production `GET
  /api/v1/health` returned `200 {"data":{"status":"ok"}}`; no credentials
  were entered and no authenticated or financial request/write was made.
- Current-source quality checks pass: native project/base-component
  verification, web design guard + design tests + TypeScript typecheck + Vite
  production build, and `go test ./...` across backend packages. The web build
  emits only the existing large-chunk advisory; it succeeds.
- This checkpoint does not verify production sign-in, authenticated data
  rendering, production feedback/AI/OCR, pixel parity, or physical-iPhone
  behavior. At this 2026-10-08 checkpoint the feedback dock collision was
  still open; the 2026-10-13 implementation and simulator proof are recorded
  below. Owner-authenticated visual/device acceptance remains open.
- Prioritized remaining owner checks are listed in
  [NATIVE-PORT-100-OWNER-UAT](NATIVE-PORT-100-OWNER-UAT.md).
- Public endpoint/API behavior and application source were not changed by this
  checkpoint; public API docs and changelog do not need updates.

## Pre-implementation checkpoint — 2026-10-12 (superseded by the following section)

- Native unit/contract suite passes **146/146** on an isolated iPhone 16
  Simulator (`MyPocket-Validation-iPhone-16`), with Swift warnings treated as
  errors. Evidence: `/tmp/mypocket-port100-current-unit-20261012.xcresult`.
- Release Simulator build passes with warnings-as-errors. Project and base
  component guards pass, as does `git diff --check`.
- The focused `testFeedbackBubbleDoesNotCoverMoreRowsOrCategoryEditorActions`
  still fails on current source: the `more-destination-wallets` row, the
  `more-destination-assistant` row after scrolling, and category-editor cancel
  intersect the feedback bubble. Evidence:
  `/tmp/mypocket-feedback-dock-red-20261012.xcresult`.
- Validation ran on a disposable simulator and did not use an API, database,
  production account, or financial data. The owner's separate iPhone 16
  Simulator remains booted at the MyPocket email/password login screen in
  Device Hub for manual testing; no credentials were entered.
- At this checkpoint, shell implementation had not started. These results are
  retained as the before-state; the 2026-10-13 section below records the dock
  fix and subsequent verification. Neither checkpoint establishes complete
  feature parity or authenticated physical-device UAT.

## Native Feedback Dock implementation — 2026-10-13

- PASS: feedback collision and capture geometry on disposable iPhone 16,
  including category editor save/cancel clearance:
  `/tmp/mypocket-feedback-dock-category-green-20261012.xcresult`.
- PASS: More final-row reachability plus category and notification bounded
  viewport regressions:
  `/tmp/mypocket-feedback-dock-spacer-green-20261012-1.xcresult`.
- PASS: shell route, transfer/adjustment inset and feedback capture tests are
  also recorded in `.superpowers/sdd/2026-10-12-native-feedback-dock/progress.md`.
- PASS: compact iPhone SE (3rd generation, iOS 18.5) capture, More final-row
  reachability and selected Account-tab hit-test pass in
  `/tmp/mypocket-feedback-dock-se-20261013.xcresult` (3/4 selected tests; the
  remaining assertion was corrected because it measured an offscreen action).
  The follow-up category collision test passes at
  `/tmp/mypocket-feedback-dock-se-green-20261013.xcresult`, after scrolling
  save/cancel fully into the bounded viewport and checking dock clearance.
- The initial focused RED is retained as historical evidence at
  `/tmp/mypocket-feedback-dock-red-20261012-2.xcresult`; the follow-up after a
  category-editor accessibility-query correction passes at
  `/tmp/mypocket-feedback-dock-category-green-20261012.xcresult`.
- PASS: `MyPocketIOSTests` 146/146 at
  `/tmp/mypocket-feedback-dock-unit-20261013.xcresult`; Release simulator build
  with `SWIFT_TREAT_WARNINGS_AS_ERRORS=YES` succeeded; `verify-project.sh`,
  both base-component guards/tests and `git diff --check` passed. The shell
  add-action long-press regression also passes 1/1 at
  `/tmp/mypocket-feedback-dock-add-action-20261013.xcresult`.
- Only disposable simulators and synthetic UI state were used; no API,
  database, production account or financial data was accessed. Owner device
  and authenticated visual UAT remain untouched/open.

## Physical iPhone release readiness — 2026-10-13

- PASS: current-source generic Release simulator build with Swift warnings
  treated as errors at `/tmp/mypocket-port100-release-refresh-20261013`.
- PASS: current-source Release build for `generic/platform=iOS` (arm64) with
  code signing disabled at `/tmp/mypocket-port100-iphone-release-20261013`.
  This proves device-SDK compilation and packaged resources, not installability.
- PASS: `verify-project.sh`, the base-component guard/tests, and
  `git diff --check`.
- A local signed-build probe targeted the paired iPhone 15 Plus with the
  available development team override but without `-allowProvisioningUpdates`.
  Xcode timed out before signing because the device is locked and reports that
  development services require unlock. The local provisioning-profile folder
  is empty; the project does not persist a development team. No app was
  installed and no provisioning profile or Developer Portal resource was
  created or changed.
- Next physical-device gate: unlock the iPhone, select the correct Team in
  Xcode Signing & Capabilities, let Xcode create/use its development profile,
  then install and run the current Debug build for authenticated owner UAT.
  Do not treat the simulator Release build as a device-installable artifact.
- The device-SDK build also reports Xcode's interface-orientation validation
  warning: all orientations are expected unless the app requires full screen.
  Resolve whether this remains an iPhone-only product, supports iPad
  multitasking, or intentionally requires full screen before App Store-style
  distribution; it does not prevent direct iPhone testing.

## Current native validation refresh — 2026-10-08

| Check | Command / evidence | Result |
| --- | --- | --- |
| Native unit/contract suite | `rtk xcodebuild test -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS -destination 'platform=iOS Simulator,id=EC80B092-CEF1-4A8E-A5A3-7420298EFD4D' -only-testing:MyPocketIOSTests -resultBundlePath /tmp/mypocket-port100-current-unit-20261008.xcresult` | PASS: 144/144 on iPhone 16 Simulator, zero failures/skips; `xcresulttool` reports `Passed`. |
| Release simulator build | `rtk xcodebuild build -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS -configuration Release -destination 'platform=iOS Simulator,id=EC80B092-CEF1-4A8E-A5A3-7420298EFD4D' CODE_SIGNING_ALLOWED=NO -resultBundlePath /tmp/mypocket-port100-current-release-20261008.xcresult` | PASS: `BUILD SUCCEEDED`; simulator-only, code signing disabled. |
| Project/base guards | `rtk bash mobile-ios/scripts/verify-project.sh && rtk bash mobile-ios/scripts/check-base-components.test.sh && rtk git diff --check` | PASS: project verification, base-component guard/tests, and whitespace check. |

These checks establish current-source compilation and native unit/contract
health only. They do not close API-backed UI flows, screenshot/pixel parity,
the open feedback dock collision, populated-data visual acceptance, or
physical-device UAT. No API/database/provider was called by this validation
refresh.

## iOS AppIcon packaging — 2026-10-08

| Gate | Verification | Result |
| --- | --- | --- |
| RED project guard | Added AppIcon catalog/resource/dimension assertions before adding the asset | PASS: guard failed because the catalog was absent |
| Artwork source guard | Regenerate from canonical PWA SVG and byte-compare committed PNG | PASS: deterministic output matches the packaged image |
| Release build | `rtk xcodebuild build -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS -configuration Release -destination 'platform=iOS Simulator,id=EC80B092-CEF1-4A8E-A5A3-7420298EFD4D' -derivedDataPath /tmp/mypocket-appicon-final-deriveddata-20261008 -resultBundlePath /tmp/mypocket-appicon-final2-release-20261008.xcresult CODE_SIGNING_ALLOWED=NO SWIFT_TREAT_WARNINGS_AS_ERRORS=YES` | PASS: `BUILD SUCCEEDED` after regenerating Xcode project from `project.yml` |
| Compiled asset | `rtk xcrun assetutil --info .../MyPocketIOS.app/Assets.car` | PASS: `Name=AppIcon`, `RenditionName=AppIcon.png`, 1024×1024 pixels |
| Project guard | `rtk bash mobile-ios/scripts/verify-project.sh` | PASS: verifies catalog, PNG dimensions and XcodeGen resource path |

The iOS icon reuses `app/public/pwa-icon.svg` wallet artwork with a square
background so the operating system supplies its normal icon mask. The
reproducible transform is `mobile-ios/scripts/generate-app-icon.sh`; the
checked-in PNG is the compiled asset input. This closes packaging only, not the
remaining feedback dock, populated-state visual review or physical-device
UAT. No API, backend, database, provider or production service was used.

## Native core business-flow UI proof — 2026-10-08

| Check | Verification | Result |
| --- | --- | --- |
| Persisted reports | API-backed iPhone 16 UI test creates income, included expense and excluded expense, reloads trend reports and asserts exact persisted values | PASS |
| Wallet CRUD | Native UI creates and edits a wallet against a fresh local API/PostgreSQL database | PASS |
| Wallet-scoped transaction CRUD | Native wallet detail creates and updates a transaction and verifies the persisted result through API-backed UI | PASS |
| Internal transfer lifecycle | Native UI submits a transfer, verifies the transfer row/detail and deletes the atomic pair | PASS |
| Combined regression bundle | `xcodebuild test`, selected four API-backed UI cases, iPhone 16 Simulator, isolated PostgreSQL `mypocket_coreflow_20261008`, dedicated Redis container on loopback port 16386 | PASS: 4/4, zero failures; xcresult `/tmp/mypocket-core-business-20261008.xcresult` |

The test database and dedicated Redis container were removed after the run;
DerivedData was removed while the xcresult was retained. Provider credentials
were disabled. This proves these specific core flows against the current local
API; it does not close the full feature matrix, feedback dock collision, visual
owner sign-off, or physical-device UAT. Production and shared Redis DB 15 were
not used.

## Native transfer pair-edit integrity — 2026-10-08

- RED: the pre-rename transfer lifecycle UI test (same test body, before edit
  coverage was added) exposed that edit mode rendered both wallet selectors
  as enabled although the existing `PATCH` payload has no source/destination
  fields.
- Root cause: `TransferEditorView` reused create-mode selectors during edit;
  `FinanceAPI.updateTransfer` correctly sent only amount, date and note. The
  pair update succeeded in the API, but the UI implied wallet reassignment was
  supported.
- Fix: edit-mode selectors are disabled and render without a disclosure arrow;
  create mode remains selectable. No backend behavior, schema or endpoint
  changed.
- GREEN: `testAPIBackedTransferCreatesEditsAndDeletesAtomicPairFromNativeUI`
  verifies on iPhone 16 Simulator that source/destination are disabled,
  edits amount/note, reads both persisted rows from the owner-scoped API to
  confirm the same wallets and shared pair fields, then deletes the pair. PASS
  1/1 at `/tmp/mypocket-transfer-edit-green-20261008.xcresult`; RED evidence is
  `/tmp/mypocket-transfer-edit-red-20261008.xcresult`.
- Run used a fresh version-30 disposable PostgreSQL database and a dedicated
  loopback Redis container; both and the test DerivedData were removed. No
  production or shared Redis DB 15 access.

## Native secondary-route navigation follow-up — 2026-10-08

| Check | Verification | Result |
| --- | --- | --- |
| Focused AI-entry back navigation | `rtk xcodebuild test -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS -destination 'platform=iOS Simulator,id=EC80B092-CEF1-4A8E-A5A3-7420298EFD4D' -only-testing:MyPocketIOSUITests/LaunchTests/testAIEntryScreenBackActionHasStableIdentifier -resultBundlePath /tmp/mypocket-ai-entry-back-ui-20261008.xcresult` | PASS 1/1 in 23 seconds; opens the route, finds/hits `screen-back`, returns to More. This confirms the shared `MPIconButton` already provides the expected identifier; no production-code change was needed. |
| All 15 More destinations | `testAuthenticatedMoreMenuOpensEveryDestination` | INCONCLUSIVE: after 328 seconds, manually interrupted Xcode while the UI test was still progressing through routes. The partial `.xcresult` reports cancellation, not a product assertion failure; do not count it as pass/fail. |
| Ordered 15-route sweep | `rtk xcodebuild test -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS -destination 'platform=iOS Simulator,id=EC80B092-CEF1-4A8E-A5A3-7420298EFD4D' -only-testing:MyPocketIOSUITests/LaunchTests/testAuthenticatedMoreMenuOpensEveryDestination -resultBundlePath /tmp/mypocket-port100-secondary-routes-optimized-20261008.xcresult -parallel-testing-enabled NO` | PASS 1/1 in 240 seconds for navigation/title/back actions only. This was a non-`LOCAL_API_UI` build using an in-memory synthetic session; data-backed screens showed API-load errors. Screenshots are not API/rendering/visual-parity evidence. |
| Ordered 15-route sweep with real local API | `rtk proxy env MYPOCKET_TEST_API_BASE_URL='http://127.0.0.1:8081/api/v1' xcodebuild test -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS -destination 'platform=iOS Simulator,id=EC80B092-CEF1-4A8E-A5A3-7420298EFD4D' -derivedDataPath /tmp/mypocket-port-routes-api-derivedData-20261008 -resultBundlePath /tmp/mypocket-port-routes-api-20261008.xcresult -parallel-testing-enabled NO -only-testing:MyPocketIOSUITests/LaunchTests/testAuthenticatedMoreMenuOpensEveryDestination SWIFT_ACTIVE_COMPILATION_CONDITIONS='DEBUG LOCAL_API_UI' MP_API_BASE_URL='http://127.0.0.1:8081/api/v1'` | PASS 1/1 in 240.705 seconds; 15 route/title/back assertions and attachments. A fresh blank DB migrated to version 30 plus a dedicated Redis container; inspected non-provider routes show valid empty states. |

The broad sweep now completes with screen-order traversal and preserved menu
scroll position; retain the canceled first attempt as historical evidence, not
as a failure. At capture time there was no local API listener on ports 8080 or
8081, so those first screenshots show network errors; the earlier API-backed
description is withdrawn. A separate fresh `LOCAL_API_UI` run authenticated
through the local fixture API and captured valid empty states on an isolated,
version-30 database. The sweep still asserts route/title/back behavior rather
than populated data, pixel parity, or owner/device acceptance. AI provider/OCR
capabilities were intentionally disabled for isolation, so Finance Assistant
and AI Entry require provider-backed verification. No runtime/API behavior
changed.

## Report exclusion persistence and native trend accessibility — 2026-10-08

| Gate | Verification | Result |
| --- | --- | --- |
| Repository regression | `TEST_DATABASE_URL=<disposable local PostgreSQL clone> rtk go test ./internal/infrastructure/repository -run '^TestCreatePersistsExplicitReportExclusion$' -count=1` | RED reproduced `requested=false persisted=true`; GREEN verifies both explicit false and true through repository Create/readback |
| Backend regression | `TEST_DATABASE_URL=.../mypocket_report_verify_20261008 rtk go test ./... -count=1` | PASS across all packages; clone created from local stage and dropped after run |
| Native API-backed UI | iPhone 16 Simulator `testAPIBackedReportsTrendExposesPersistedIncomeAndExpenseValues` | PASS 1/1 at `/tmp/mypocket-report-trend-green-20261008.xcresult`; persisted included income 123,456 and expense 34,567 VND, excluded expense 888,888 VND, exact VoiceOver chart value confirmed |
| Native unit and Release gates | 144 native tests; Release Simulator build with warnings-as-errors; base-component and project guards; `git diff --check` | PASS |
| Isolation | cloned local disposable DB `mypocket_report_accessibility_20261008` for API/UI and `mypocket_report_verify_20261008` for full backend suite; Redis DB 15, API 8081 | PASS; SQL confirmed 0 fixture wallets and 0 fixture transactions; API stopped and both disposable DBs dropped. Production and API 8080 untouched |

The repository bug was GORM's default-tag behavior: a zero-value bool was
omitted from `Create`, so PostgreSQL applied `DEFAULT true` despite the API
request carrying false. `TransactionPostgresRepository.Create` now explicitly
updates the caller-selected flag inside the same transaction. The existing
finance-query integration fixture now uses repository `Create` for owner rows,
so future regressions are not hidden by raw SQL repair. The chart's
accessibility value reuses existing daily points and honors account masking;
the public API guide now documents the existing field semantics. No route,
schema or migration contract changed.

### Cross-flow regression after transaction-create persistence fix — 2026-10-08

- PASS: `testAPIBackedWalletDetailCreatesAndUpdatesTransaction` and
  `testAPIBackedRecurringCreateUpdateDeleteFromNativeUI` passed 2/2 on the
  iPhone 16 Simulator at
  `/tmp/mypocket-create-regression-ui-20261008.xcresult`. This covers the
  standard transaction `Create` path and recurring create/update/delete with a
  generated ledger row.
- PASS: teardown confirmed zero fixture wallets, transactions, and recurring
  schedules. Tests used cloned local database `mypocket_port_regression_20261008`,
  Redis DB 15 and API 8081; API stopped and clone dropped. Production and API
  8080 were untouched.
- No further code or public-contract changes were needed. The prior report fix's
  full backend suite and native 144-test/Release/guard gates remain the current
  verification evidence above.

## Current native source verification — 2026-10-08

- Travel Mode manual transaction link/unlink passes 1/1 on iPhone 16 Simulator
  at `/tmp/mypocket-travel-link-final-20261008.xcresult`. The test creates a
  trip and ordinary transaction through the supported UI/API fixtures, then
  asserts persisted event ID and null after clearing through the native detail
  selector. Backend auto-link/clear contract tests pass 2/2; native
  unit/contract tests pass 144/144 at
  `/tmp/mypocket-travel-link-unit-20261008.xcresult`.
- API-backed test used only local loopback API 8081, Redis DB 15 and cloned
  disposable `mypocket_travel_link_verify_20261008`. Teardown verification found
  0 fixture wallets, events and transactions; the API stopped and the exact
  disposable database was removed. Production and API 8080 were untouched.
- First UI attempts exposed test-only navigation/accessibility locator
  assumptions; after targeting the detail selector label and returning from
  the Travel subroute, the focused test passed. No app/backend/API/schema
  behavior changed. Public API and AI-readable endpoint documentation remain
  accurate and require no edit.

- Category-picker parity: API-backed iPhone 16 budget create/update/delete UI
  passes 1/1 at `/tmp/mypocket-category-root-cards-ui-pass-20261008.xcresult`.
  The test asserts multiple root groups, searches/selects a category, and
  verifies its persisted category ID. App and helper both used local API
  `127.0.0.1:8081` against a disposable DB clone; production was not accessed.
  Recurring create/update/delete/run-due and wallet-detail transaction
  create/update also pass 2/2 at
  `/tmp/mypocket-category-picker-crossflow-ui-final-20261008.xcresult`, with
  category persistence asserted through the API.
  A first test attempt showed UIAutomation did not expose the filtered card
  container consistently even though the matching category row remained
  accessible. The brittle accessibility-count assertion was removed; pure
  group/search tests retain exact partition assertions. The final API/UI test
  passes.
- After this picker change, the native unit/contract target passes 143/143 at
  `/tmp/mypocket-category-picker-unit-20261008.xcresult`; Release
  warnings-as-errors and project/base-component guards plus `git diff --check`
  pass.
- Current-source refresh: `rtk go test ./...` without a database DSN passes
  240 tests across 16 backend packages. After configuring the local stage DB,
  the database-backed `go test -race ./... -count=1` passes 265 tests across
  16 packages (recorded below). The iOS `MyPocketIOSTests` target passes 142/142 with zero
  skips/failures on iPhone 16 Simulator at
  `/tmp/mypocket-port100-unit-20261008.xcresult`.
- Release Simulator build passes with warnings treated as errors at
  `/tmp/mypocket-port100-release-20261008`.
- `mobile-ios/scripts/verify-project.sh`, `check-base-components.sh`, and
  `check-base-components.test.sh` pass. The focused feedback-collision UI test
  has not been rerun because the shell layout is unchanged; its recorded red
  assertions remain the authoritative UI status.
- `xcrun devicectl list devices` currently reports the iPhone 15 Plus as
  unavailable. This verification is Simulator-only; physical-device visual
  acceptance remains open.
- Feedback-bubble regression: the focused UI test ran 1 test and failed 2
  overlap assertions (More wallets row and category-editor cancel button).
  Xcode then stalled finalizing its result bundle; its session log contains the
  assertion failures, but the xcresult is incomplete. The ineffective inset
  increase was removed. No GREEN UI result is claimed.
- Historical post-revert evidence was 138/138 native tests and 231 backend
  tests across 16 packages. It is superseded for current source counts by the
  refresh above. An earlier attempt mistakenly passed a database name where
  tests require a full DSN; that result is discarded.
- No source/API contract changed in this verification pass. Public API and
  AI-readable endpoint docs already describe the category active/archive field
  and endpoint; no public-doc edit was needed. The overlap fix remains
  unverified pending detail-design approval, implementation, and a focused
  simulator UI rerun.
- The local test database originally lacked migration `000030` (schema version
  29); an initial DB-backed run therefore failed on missing `categories.is_active`.
  Migration `000030` was applied additively to the local stage database only,
  advancing it to version 30 without resetting data. The fresh disposable DB
  and the local stage DB both pass `go test -race ./... -count=1` (265 tests,
  16 packages). The disposable DB was dropped after verification. Production
  was not accessed or changed. The precise DB-backed evidence is recorded in
  the category-archive section below.
- The earlier pre-attempt native unit result remains recorded at
  `/tmp/mypocket-port100-unit-final.xcresult`; the post-revert result above is
  authoritative for the current source.

- Native unit/contract target passes 138/138 with zero failures/skips on the
  booted iPhone 16 Simulator at
  `/tmp/mypocket-port100-unit-final.xcresult`.
- Release simulator build passes with
  `SWIFT_TREAT_WARNINGS_AS_ERRORS=YES` at
  `/tmp/mypocket-port100-release-current`.
- `mobile-ios/scripts/verify-project.sh`,
  `check-base-components.sh`, `check-base-components.test.sh`, and
  `git diff --check` pass.
- This is source/build and unit-contract evidence only; it does not replace
  API-backed E2E, populated-state visual comparison, or device owner UAT. The
  feedback-bubble/list collision regression remains intentionally unresolved
  pending the shell-vs-scroll-container layout decision recorded in
  [its bug note](bugs/NATIVE-LIST-FEEDBACK-BUBBLE-COLLISION.md).
- No runtime behavior, API, schema, migration, or configuration changed in
  this verification pass. Public documentation and release notes do not need
  changes.

## Current web design and production-build verification — 2026-10-08

- `npm run test:design` passes: seven design-rule tests plus base contracts,
  wallet-scope, report-trend, reports-trend and status-badge contracts.
- `npm run test:auth` passes for the password-login request/response contract.
- `npm run build` passes the design guard, 132 local documentation links,
  TypeScript typecheck and Vite production build.
- Vite reports the main JavaScript chunk is 732.61 kB minified (210.23 kB
  gzip), above its 500 kB advisory threshold. Build succeeds, but chunk
  splitting/performance remains a follow-up and is not evidence of port parity.
- No public API or business contract changed during this verification pass.
  UI design docs and release notes already cover these in-progress web changes;
  no additional user-facing/API documentation change is indicated by running
  these checks.

## Native category list route and wallet-scope parity — 2026-10-08

- RED: API-backed iPhone 16 category CRUD exposed that list controls were in
  the wrong order, rows had separate pencil/delete actions, and the subtitle
  omitted wallet applicability. The first parity assertion failed at the
  ordering check; a follow-up run then exposed that the new row ID needed to be
  attached to the underlying shared `MPButton`.
- GREEN: `LaunchTests.testAPIBackedCategoryCreateUpdateDeleteFromNativeUI`
  passes 1/1 at
  `/tmp/mypocket-category-list-parity-green2.xcresult`. It verifies kind/add/
  wallet/search order, full-row navigation to edit, no list-level delete,
  scope subtitle from `wallet_ids`, category filters, rename, editor delete
  and server readback.
- Captures: category list, editor and wallet-scope editor screenshots are
  attached in `/tmp/mypocket-category-list-parity-green2-attachments/` and
  compared with the Stitch references.
- Native unit/contract target: 138/138, zero skips/failures at
  `/tmp/mypocket-category-list-parity-unit.xcresult`. Release simulator build
  with warnings-as-errors passes; project verification and both base
  component guard/test scripts pass; `git diff --check` passes.
- Isolation: local `mypocket_stage_v1`, Redis DB 15 and API `127.0.0.1:8081`;
  fixture wallet/category counts are zero after test cleanup. No production
  data or API 8080 was used.
- Backend/API/schema/migration: unchanged because existing owner-scoped
  `/categories` response already supplies `wallet_ids`. Public API and
  AI-readable endpoint docs need no update.

## Native credit-entry category tree — 2026-10-08

- RED: the API-backed iPhone 16 test failed because tapping the credit “Nhóm”
  menu did not expose the shared searchable category picker:
  `/tmp/mypocket-credit-category-red/Logs/Test/Test-MyPocketIOS-2026.10.08_10-21-00-+0700.xcresult`.
- GREEN: the same entry/payment flow passes 1/1 at
  `/tmp/mypocket-credit-category-green/Logs/Test/Test-MyPocketIOS-2026.10.08_10-24-41-+0700.xcresult`.
  It preserves amount and note drafts while switching views, persists the
  selected eligible category, and reads that category ID back through the
  fixture owner's credit statement API. The payment path still completes.
- Isolation/cleanup: only local `mypocket_stage_v1` and Redis DB 15 were used
  through `127.0.0.1:8081`. Post-test fixture wallet count and associated
  transaction count are both 0; the temporary API is stopped. Production and
  existing port 8080 were untouched.
- Native unit/contract: 137/137, zero skips/failures at
  `/tmp/mypocket-credit-category-unit/Logs/Test/Test-MyPocketIOS-2026.10.08_10-25-38-+0700.xcresult`.
- Release warnings-as-errors build: PASS at
  `/tmp/mypocket-credit-category-release`; project verification, both base
  component guards/tests, and `git diff --check` pass.
- Backend review: the actual existing categories and credit statement APIs
  satisfy the selected-ID and applicability contract; no backend, schema,
  public API or AI-readable endpoint catalog change was required. Backend
  `go test ./...` also passes with `TEST_DATABASE_URL` isolated to local
  `mypocket_stage_v1`.

## Native category editor route, wallet scope and wire contract — 2026-10-08

- RED: API-backed iPhone 16 CRUD failed at the fresh-read assertion for the
  selected wallet. DB staging showed the category was created without links.
  A wire-level XCTest then proved `JSONEncoder.convertToSnakeCase` encoded
  `walletIDs` as `wallet_i_ds`, not the Go API field `wallet_ids`:
  `/tmp/mypocket-category-wire-red/Logs/Test/Test-MyPocketIOS-2026.10.08_10-44-01-+0700.xcresult`.
- GREEN: `CategoryModelInput` now explicitly emits `wallet_ids`; the exact-key
  XCTest passes at `/tmp/mypocket-category-wire-green2/Logs/Test/Test-MyPocketIOS-2026.10.08_10-45-15-+0700.xcresult`.
  Backend category creation now stores the category plus owner-validated wallet
  links atomically; repository scope regression passes against local stage.
- API-backed iPhone 16 create/filter/update/delete flow passes 1/1 at
  `/tmp/mypocket-category-editor-final/Logs/Test/Test-MyPocketIOS-2026.10.08_10-45-32-+0700.xcresult`.
  It verifies visible checkbox rows, a one-wallet selection, fresh API readback
  (A selected, B excluded), income/expense filters, rename and confirmed delete.
- Isolation: only `mypocket_stage_v1`, Redis DB 15 and local API 8081 were used.
  Exact `9A76A17B` wallet/category/transaction fixture counts are all zero after
  test teardown; API 8081 is stopped. Production and API 8080 were untouched.
- Native unit/contract suite: 137/137, zero failures/skips at
  `/tmp/mypocket-category-unit-final/Logs/Test/Test-MyPocketIOS-2026.10.08_10-49-14-+0700.xcresult`.
- Backend `go test ./...` passes with `TEST_DATABASE_URL` pinned to local
  `mypocket_stage_v1`; Release warnings-as-errors, project verification, both
  base-component guards/tests and `git diff --check` pass.
- Public endpoint, request schema and migrations are unchanged; the API
  reference already documents owner-scoped `wallet_ids`. Internal category
  design, native architecture, phase, backlog, audit and changelog are updated.

## Full LOCAL_API_UI suite — repeatable fixture rerun — 2026-10-08

| Check | Evidence | Result |
| --- | --- | --- |
| Full sequential iPhone 16 unit + UI suite | `/tmp/mypocket-port-local-api-ui-current/Logs/Test/Test-MyPocketIOS-2026.10.08_09-14-09-+0700.xcresult` | PASS; 51/51, 0 failures, 1,730.041 seconds; `xcodebuild` exit 0 |
| Notification test repeatability | `mobile-ios/scripts/manage-local-notification-fixture.sh prepare` followed by `cleanup` | PASS; exact synthetic row reset before run and removed after run |
| API isolation | loopback API `127.0.0.1:8081`, local `mypocket_stage_v1`, Redis DB 15 | PASS; production and API port 8080 untouched |
| Release simulator build with warnings-as-errors | `/tmp/mypocket-port-release-current` | PASS; `SWIFT_TREAT_WARNINGS_AS_ERRORS=YES` |
| Native project/base guards and source whitespace | `verify-project.sh`, `check-base-components.sh`, `check-base-components.test.sh`, `git diff --check` | PASS |

The run explicitly embedded `http://127.0.0.1:8081/api/v1` in the test app.
The fixture helper is restricted to a loopback PostgreSQL DSN and the local
`mypocket_stage_v1` database; it does not alter app or backend behavior. The
earlier notification-fixture failure and the previous portfolio selector
collision are superseded by this clean full-suite result. Exact notification
fixture cleanup completed after the run. This establishes the current local
unit/API-backed UI suite only; external physical-device, OCR credentials,
production API-key/feedback and final pixel-parity gates remain open in the
port audit. No public API, user-facing behavior, schema or migration changed,
so public docs and release notes do not need changes.

## Full LOCAL_API_UI suite checkpoint — 2026-10-08

### Verified sequential rerun

| Check | Evidence | Result |
| --- | --- | --- |
| Full sequential iPhone 16 UI suite | `/tmp/mypocket-port-local-api-ui-verified.xcresult` | 1,740.4 seconds; one unique failure: `testAPIBackedPortfolioAssetCreateAndPriceRefresh()` |
| Failure classification | portfolio test activity and captured accessibility hierarchy | Test fixture collision: the fixed symbol `FXT` matched two persisted cards, so `app.textFields["Giá hiện tại"]` was ambiguous; no product behavior failure established |
| API isolation | temporary API on `127.0.0.1:8081`, database `mypocket_stage_v1`, Redis DB 15 | PASS; production and port 8080 untouched |

Planned correction before the final rerun: give each portfolio asset's current-
price input a stable ID derived from its asset ID, make the fixture symbol unique
per invocation, then rerun this focused case and the complete suite. The stable
ID is a UI accessibility/testability contract and does not change visible layout
or API behavior. Keep verification open until both pass and fixture cleanup is
confirmed.

The first focused rerun after rebuilding the test bundle is invalid: the build
used `Debug.xcconfig`'s default app API URL (`127.0.0.1:8080`) instead of the
isolated fixture API (`127.0.0.1:8081`). The test only navigated/read and did not
submit a UI mutation, but do not use that result as verification. All subsequent
API-backed UI builds/runs must explicitly set `MP_API_BASE_URL` to the loopback
fixture endpoint; port 8080 remains out of scope.

The price field and update button now expose asset-ID-scoped accessibility
identifiers. With the app bundle verified to embed `http://127.0.0.1:8081/api/v1`,
the unique-fixture create-and-price-refresh test passes at
`/tmp/mypocket-port-portfolio-a11y-id-rerun.xcresult`. The full suite is pending.
This accessibility-only contract does not alter public API, user-visible copy,
or business behavior, so public docs and release notes do not change.

The later full-suite attempt was interrupted when the conversation turn was
preempted; its live session ended without a finalized result bundle, so it is
not counted as a pass or fail. Rebuilt test products explicitly against 8081;
the native unit/contract target then passed 131/131 on iPhone 16 Simulator at
`/tmp/mypocket-port-unit-after-a11y-retry.xcresult`. Full `LOCAL_API_UI` sweep
still needs a fresh run after the accessibility-ID change.

### Notification fixture repeatability — implementation planned

The next full `LOCAL_API_UI` run exposed that the fixed synthetic notification
survives between runs and can already have `read_at` set. The test requires a
real unread-to-read transition, so its setup must re-seed/reset only the
documented notification row on the loopback `mypocket_stage_v1` database before
the suite; after verification, remove only that exact synthetic row. No
production route, app behavior, migration, or public API contract should change.

| Check | Evidence | Result |
| --- | --- | --- |
| Build-for-testing with `LOCAL_API_UI` and API base `127.0.0.1:8081` | `/tmp/mypocket-port-local-api-ui-derivedData` | PASS; app Info.plist embeds only the isolated loopback API URL |
| Full sequential iPhone 16 UI suite | `/tmp/mypocket-port-local-api-ui-final.xcresult` | FAIL after 1,607.5 seconds; see failure classification below |
| API isolation | temporary API configured for `mypocket_stage_v1`, Redis DB 15, fixture owner | PASS; stopped after the suite; no port 8081 listener remains; port 8080 and production untouched |
| Financial fixture cleanup | pending read-only inventory | OPEN; verify exact test-created records before next run |

The suite has nine unique failing cases (Xcode lists retry attempts separately):

| Area | Failure evidence | Preliminary classification |
| --- | --- | --- |
| Finance Assistant | `LaunchTests.swift:1225` expects the disabled-state copy | Test/config mismatch: the local environment has provider configuration; reconcile expected enabled/disabled capability instead of asserting a fixed flag |
| Jar create/update/remove | `:1755` Save is not hittable; failure hierarchy shows the jar was created in the cumulative selector but not attached to the active month | UI layout or test workflow mismatch; test must exercise the explicit add-to-month action and verify behavior |
| Ledger and Reports | `:1555`, `:1559`, `:1321` require real transaction rows/chart, but the fixture owner has no baseline ledger data | Missing isolated synthetic seed, not yet evidence of a rendering defect |
| Month note | `:1865` Save is not hittable after editing | Reproduce as an actionable UI/layout defect |
| Notifications | `:1093` documented synthetic unread notification is absent | Missing isolated fixture seed |
| Travel Mode | test searches button label `Thêm chuyến đi`; rendered accessibility label is `Thêm` (`travel-add`) | Stale test selector; preserve the current visible label unless web/Stitch evidence says otherwise |
| Wallet editor | `:1165` uses `-UITestAuthenticated` against the API-backed build and has no wallet fixture | Wrong test launch mode/fixture; use a seeded API wallet or keep it in the non-API test suite |
| Wallet detail transaction editor | `:2081` cannot find static text `Ngày giao dịch` | Verify whether the date control is missing or the test uses an outdated accessibility query |
| Wallet create/update | `:2578` expects a direct row edit control while wallet management requires entering its explicit edit mode | Test workflow mismatch with the currently documented grouped/edit-mode contract |

Next verification plan: first add deterministic, uniquely named, owner-scoped baseline fixtures for ledger/report/notification checks with exact teardown; correct stale UI test assumptions against current web/Stitch behavior; then fix only reproduced UI defects (including non-hittable Save controls and the transaction-date field). Re-run focused tests before another full sequential sweep. The category editor route/wallet-scope gate has since been authorized and verified in the focused 2026-10-08 section above; remaining category list/picker and overall visual parity are still open. This run does not establish full iOS functional parity or owner acceptance.

## Focused failure rerun — 2026-10-08

- The ten formerly failing cases were rerun against the same loopback API;
  seven now pass, including the ledger with twelve unique owner-scoped rows,
  notification mark-read, Assistant capability states, jar mutation, Travel
  Mode, wallet detail date picker and wallet editor route.
- Three failures remain: the month-note assertion expects `isHittable` before
  the accessibility tap performs its own scroll-to-visible, although the
  editor value changes, Save is enabled, and the API save/delete assertions
  continue successfully; the Reports test queries a nested chart ID that
  SwiftUI replaces with the enclosing summary-card ID (the empty-state text is
  present); and wallet management exposes edit by tapping the row in edit mode,
  not by a separate button named “Sửa ví …”.
- The next focused run removes only the stale month-note `isHittable`
  precondition (it retains enabled-state, tap, reload/persistence and delete
  assertions), queries report chart/empty-state by visible labels, and resolves
  the wallet ID through the owner API before tapping `wallet-scope-<id>` in
  edit mode. These are test-contract and accessibility-query corrections; no
  product code change was indicated.
- Exact result: `/tmp/mypocket-port-failure-focused-rerun.xcresult`;
  7/10 pass, 3/10 fail. The temporary API is still isolated on 8081 and will be
  stopped after the follow-up; production remains untouched.

## Fresh cross-surface checkpoint — 2026-10-07

- Native unit/contract target: PASS, 127/127, 0 failures/skips on
  `MyPocket-iPhone-16`; latest evidence is
  `/tmp/mypocket-auth-unit-rerun.xcresult`. The prior 129-test checkpoint
  preceded removal of two obsolete Google-auth tests.
- Native Release simulator build with warnings-as-errors: PASS; latest derived
  data evidence is `/tmp/mypocket-auth-release`.
- Native project guard, shared base-component guard/test, and `git diff --check`:
  PASS.
- Backend `go test ./...`: PASS.
- Web `npm run build`: PASS (design guard, 130 local design links, TypeScript,
  Vite production build). Vite still reports the pre-existing bundle-size
  advisory; it is not a build failure.

## Native ordinary-ledger wallet selector parity — 2026-10-08

| Check | Evidence | Result |
| --- | --- | --- |
| Native wallet scope glyph contract | `DesignSystemTests.testWalletScopeGlyphMatchesWebSelectorIdentityAndTone` | PASS |
| Native unit/contract target | `/tmp/mypocket-wallet-scope-selector-unit/Logs/Test/Test-MyPocketIOS-2026.10.08_03-36-07-+0700.xcresult` | PASS, 129/129, zero skips/failures |
| Release warnings-as-errors | `/tmp/mypocket-wallet-scope-selector-release` | PASS |
| Project/base/diff guards | `verify-project.sh`, `check-base-components.sh`, `check-base-components.test.sh`, `git diff --check` | PASS |
| API-backed wallet-selection UI | `/tmp/mypocket-wallet-scope-api-ui/Logs/Test/Test-MyPocketIOS-2026.10.08_03-41-37-+0700.xcresult` | PASS, 1/1 on iPhone 16; asserts selected wallet and capsule-before-week-control order |
| Rendered screenshot | `/tmp/mypocket-wallet-scope-attachments/B8303648-1B89-4E52-A027-F4269B21D189.png` | PASS; selector capsule, week/custom controls and cashflow render in expected order |
| Fixture cleanup/isolation | `mypocket_stage_v1`, Redis DB 15, temporary loopback API 8081 | PASS; zero fixture wallets/transactions after teardown; API 8081 stopped; API 8080 untouched |

Native ordinary-wallet scope now composes a centered capsule directly below
the screen header like web, with a neutral aggregate/basic glyph and semantic
goal/credit accents. Selection state, transaction filtering and period
controls are unchanged. The screenshot was inspected and confirms the
selector/period/cashflow hierarchy. The screen spec, port audit, backlog and
Unreleased changelog were updated. No API/schema or machine-readable endpoint
docs changed.

## Native wallet management parity and included-total persistence — 2026-10-08

| Check | Evidence | Result |
| --- | --- | --- |
| Backend persistence regression | `repository.TestWalletCreatePersistsExplicitlyExcludedFromTotal` against `mypocket_stage_v1` | RED before repository fix; PASS after fix |
| Backend full suite | `TEST_DATABASE_URL=mypocket_stage_v1 go test ./...` | PASS |
| Native unit/contract target | `/tmp/mypocket-wallet-management-unit-final.xcresult` | PASS, 130/130; zero failures/skips |
| API-backed management UI | `/tmp/mypocket-wallet-management-api-ui-delete.xcresult` | PASS, 1/1 on iPhone 16; confirms aggregate route, edit and persisted delete |
| Screenshot inspected | `/tmp/mypocket-wallet-management-delete-attachments/46DF2A08-454E-45E5-BE0C-8167142E8351.png` | PASS; grouped rows are readable at phone width |
| Release warnings-as-errors | `/tmp/mypocket-wallet-management-release` | PASS |
| Project/base/diff guards | `verify-project.sh`, `check-base-components.sh`, `check-base-components.test.sh`, `git diff --check` | PASS |
| Stage isolation | `mypocket_stage_v1`, Redis DB 15, API 127.0.0.1:8081 | PASS; fixtures 0 after teardown; API 8081 stopped, 8080 unchanged |

The UI regression confirms included/excluded groups, included-only aggregate
amount, aggregate routing to unscoped transactions, Sửa/Xong and row-to-edit,
confirmed delete through the editor, retained Chi tiết, add action and disabled
service-link action. It exposed GORM's default-tag behavior changing an explicit
false to true; wallet creation now writes the supplied value transactionally.
This fixes persisted behavior without changing the request, schema, default or
public API documentation contract.

## Native Account shell navigation regression — 2026-10-08

- RED: after long-pressing the global add button to open AI Entry and returning
  to More, the UI hierarchy showed the Account tab still selected after tapping
  Transactions. `MainTabView` had disabled hit-testing on the entire bottom
  navigation whenever Account was selected.
- FIX: shell tab navigation remains interactive on Account; the center add
  action remains disabled there, and a narrow blocker still prevents taps on
  the already-selected Account tab from falling through to the final menu row.
  `MoreView` keeps its shared bottom-navigation inset.
- PASS: iPhone 16 Simulator, 3/3 in
  `/tmp/mypocket-longpress-fixed.xcresult`: long-press → AI Entry → Back →
  Transactions → normal add editor; normal add-form controls; and all 15 More
  destinations. The earlier 3/3-count includes each named UI test once.
- PASS: native unit/contract target is 127/127 with zero skips/failures at
  `/tmp/mypocket-shell-navigation-unit.xcresult`; Release simulator build with
  warnings-as-errors passes at `/tmp/mypocket-shell-navigation-release`.
- PASS: `verify-project.sh`, `check-base-components.test.sh` and
  `git diff --check` pass.
- SCOPE: native shell interaction only. No API, schema, owner scope, finance
  rules, or public API documentation changed.
- This checkpoint does not close the explicitly external gates: authenticated
  physical-device/pixel UAT, live OCR image credentials, and production
  feedback/API-key verification.
- Post-harness rerun: the prior recurring selector checkpoint passed 129/129;
  the current password-only auth run passes 127/127 in
  `/tmp/mypocket-auth-unit-rerun.xcresult`; Release warnings-as-errors passes
  at `/tmp/mypocket-auth-release`, with project/base/diff guards green.

## Earlier secondary-route visual-sweep claim — withdrawn 2026-10-08

| Check | Evidence | Result |
| --- | --- | --- |
| Route screenshot bundle | `/tmp/mypocket-port100-secondary-routes-optimized-20261008.xcresult` and `/tmp/mypocket-port100-secondary-routes-attachments-20261008/` | Route/title/back assertions passed; screenshots display API-load errors |
| Runtime used | Non-`LOCAL_API_UI` test build with in-memory `-UITestAuthenticated` session | Not API-backed; no local API listener on 8080 or 8081 at capture time |
| Earlier claim | This subsection previously described the capture as API-backed | WITHDRAWN: invalid attribution; do not use these images as visual/empty-state evidence |

This does not prove API-backed route rendering, valid empty states,
pixel-level Stitch/web parity, physical-device acceptance or production
behavior. A fresh isolated API-backed run is required before screenshot review.
The separate Account Face ID/floating-feedback geometry test remains evidence
only for that specific captured geometry; it does not validate the overall
feedback dock design or remove the open More/category collision.

| Focused feedback collision regression | Evidence | Result |
| --- | --- | --- |
| API-backed RED | `/tmp/mypocket-feedback-collision-api-red/Logs/Test/Test-MyPocketIOS-2026.10.08_04-38-12-+0700.xcresult` | FAIL as expected; bubble frame intersected Face ID explanation |
| API-backed GREEN | `/tmp/mypocket-feedback-collision-api-green/Logs/Test/Test-MyPocketIOS-2026.10.08_04-39-10-+0700.xcresult` | PASS, 1/1 on iPhone 16; frame no longer intersects |
| Screenshot inspected | `/tmp/mypocket-feedback-collision-api-green-attachments/D93BF6B9-DB9F-4F60-829C-24BE51B9497D.png` | PASS; explanation wraps before the floating bubble |

## Native Reports trend chart parity — 2026-10-07

| Check | Evidence | Result |
| --- | --- | --- |
| Account-timezone aggregation | `DesignSystemTests.testTrendDataUsesAccountTimezoneAndExcludesNonReportRows` | PASS |
| Shared base contract | `DesignSystemTests.testTrendChartUsesStableAccessibilityIdentifier` | PASS |
| API-backed Reports trend mode | `/tmp/mypocket-trend-ui-final.xcresult` | PASS, 1/1; trend toggle exposes the real chart/empty state |
| Web trend aggregation/base | `app/scripts/report-trend.test.mjs`, `app/scripts/reports-trend-design.test.mjs` | PASS; `npm run test:design` and `npm run build` |
| Release warnings-as-errors | `/tmp/mypocket-trend-release.log` | PASS |

Both trend surfaces use only owner-scoped transaction data and exclude rows
outside reports plus transfer/adjustment types. No API, schema or migration was
changed.

## Shared base-component completion — 2026-10-07

| Check | Evidence | Result |
| --- | --- | --- |
| Native gauge/status contracts | `DesignSystemTests` in `/tmp/mypocket-auth-unit-rerun.xcresult` | PASS, 127/127 total; clamping, accessibility IDs and semantic status tones covered |
| Native budget/recurring mounting | `BudgetsView`, `RecurringView` | PASS; real API values feed `MPBudgetGauge`/`MPStatusBadge` |
| Web status badge contract | `app/scripts/base-component-design.test.mjs` | PASS; `npm run test:design` |
| Web production/type/design checks | `npm run typecheck`, `npm run build`, `npm run check:design` | PASS; 130 valid local links; existing chunk advisory only |
| API/schema/migration impact | No backend or schema change | N/A |

## Native recurring API-backed UI regression — 2026-10-07

| Check | Evidence | Result |
| --- | --- | --- |
| Create/update/delete/run-due flow | `/tmp/mypocket-recurring-ui-final2.xcresult` | PASS, 1/1 on `MyPocket-iPhone-16` |
| Add action targeting | `recurring-add` identifier with asserted label `Thêm` | PASS; stale `Thêm lịch` lookup removed |
| Backend/API impact | Disposable local fixture only; no endpoint/schema change | N/A |

## Native password-only authentication — 2026-10-07

| Check | Evidence | Result |
| --- | --- | --- |
| Native sign-in UI | `/tmp/mypocket-auth-green.xcresult`, screenshot `/tmp/mypocket-auth-password-only.png` | PASS; email/password controls present, Google action/divider absent |
| Test-first regression | `/tmp/mypocket-auth-red.xcresult` | RED as expected: both obsolete Google controls were present before implementation |
| Native production auth code/config | `rg` over app source, plist, xcconfigs and Xcode project | PASS; no Google PKCE client, OAuth callback scheme or client configuration |
| Backend impact | Existing backend/web Google routes; only DEBUG local fixture uses synthetic exchange | No API/schema/migration change |
| Physical iPhone install | `xcodebuild` Debug device build with `DEVELOPMENT_TEAM=GYRA8D4M54` | BLOCKED: Xcode reports no signed-in account for the team and no App Development profile for `com.mypocket.ios`; requires Xcode account setup |

## Native portfolio trade contract — 2026-10-07

| Check | Evidence | Result |
| --- | --- | --- |
| Owner-scoped trade path/body/idempotency | `APIClientTests.testPortfolioTradeUsesOwnerPathSnakeCaseBodyAndIdempotency` | PASS, 1/1 at `/tmp/mypocket-portfolio-trade-redgreen.xcresult` |
| Native feature call site | `PortfolioView.saveTrade()` → `FinanceAPI.createPortfolioTrade` | PASS; API-backed native UI creates the buy, refreshes the position and shows history |
| Persisted trade history | `GET /portfolio/assets/{id}/trades` | PASS; native now fetches per-asset history instead of relying on the empty aggregate `summary.trades` field |
| Live append-only mutation/device visual UAT | production/device run | OPEN external gate |

### Native portfolio trade execution and history UI — 2026-10-08

The initial focused interaction exposed the missing history refresh: the buy
persisted and position quantity updated, but `summary.trades` was empty. Native
now fetches history from the existing owner-scoped per-asset trades route.
`testAPIBackedPortfolioTradeRefreshesPositionAndHistory` passes on the iPhone 16
simulator at `/tmp/mypocket-portfolio-trade-ui4.xcresult`: rendered quantity
and trade title are asserted, then a fresh API read confirms side, quantity,
unit price and fee. The API intentionally refuses deleting an asset with
immutable trade history, so the one random synthetic asset and its trade were
removed by exact owner + asset-ID SQL from `mypocket_stage_v1`; the post-cleanup
count is 0. The temporary API used loopback 8081 / Redis DB 15 and is stopped;
8080 and production were untouched. Native unit tests pass 127/127 at
`/tmp/mypocket-portfolio-trade-unit.xcresult`; Release warnings-as-errors,
project/base-component guards and `git diff --check` pass. No API/schema
contract changed. Physical-device pixel approval and provider/OCR gates remain
open.

## Native amount keypad parity — 2026-10-07

| Check | Evidence | Result |
| --- | --- | --- |
| Integer-safe calculator contract | `FinanceModelTests.testAmountCalculatorUsesIntegerVNDArithmeticWithoutEval` in `/tmp/mypocket-native-keypad-final-unit.xcresult` | PASS |
| Shared base integration | `MPAmountField` opens `MPAmountKeypad`; base/project guards | PASS |
| Transaction editor keypad flow | `/tmp/mypocket-keypad-ui6.xcresult` | PASS, 1/1: enter `1,000,000 + 500,000`, evaluate, commit and return |
| Release warnings-as-errors | `/tmp/mypocket-native-keypad-final-release` | PASS |
| API/schema/migration impact | No backend or schema change | N/A |
| Physical-device pixel approval | iPhone visual comparison | OPEN external gate |

## Native calendar picker parity — 2026-10-07

| Check | Evidence | Result |
| --- | --- | --- |
| Shared calendar base | `MPDateCalendarSheet` from `MPDateStepper` using the shared surface/buttons and Monday-first Vietnamese grid | PASS |
| Standalone date fields | Wallet goal, transfer/adjustment and portfolio `MPDateField` consumers reuse the sheet with account timezone | PASS (compile/contract) |
| Transaction custom range | `MPDateRangePicker` forwards `TransactionsViewModel.accountTimezone` to both fields | PASS |
| Transaction editor calendar flow | `/tmp/mypocket-calendar-custom-ui2.xcresult` | PASS, 1/1: open the custom Stitch grid, assert title/picker, capture screenshot, cancel, reopen and confirm back to the form |
| Transaction custom-range UI flow | `/tmp/mypocket-calendar-range-ui.xcresult` | PASS, 1/1: switch to custom period, open start-date calendar, cancel and return to range form |
| Full native unit/contract target | `/tmp/mypocket-native-calendar-field-final-unit.xcresult` | PASS, 121/121 |
| Release warnings-as-errors | `/tmp/mypocket-native-calendar-field-final-release` | PASS |
| Account-local serialization/API impact | Existing form callback and date contract; no backend/schema change | PASS / N/A |
| Physical-device pixel approval | iPhone visual comparison | OPEN external gate |

## Native budget date picker parity — 2026-10-07

| Check | Evidence | Result |
| --- | --- | --- |
| Budget start/end controls | `BudgetsView` uses account-timezone `MPDateField` for both date-only boundaries | PASS |
| Date-only API round-trip | `MPFormat.parseCalendarDate` / `calendarDateString` preserve `yyyy-MM-dd` wire values | PASS |
| API-backed create/update/delete | `/tmp/mypocket-budget-date-ui2.xcresult` | PASS, 1/1 against disposable local API fixture |
| Full native unit/contract target | `/tmp/mypocket-budget-date-unit-final.xcresult` | PASS, 121/121 |
| Release warnings-as-errors | `/tmp/mypocket-budget-date-release-final` | PASS |
| Stable add-action targeting | `budget-add` identifier in native UI test | PASS; avoids localized-label coupling |
| Physical-device pixel approval | iPhone visual comparison | OPEN external gate |

## Native date-time picker parity — 2026-10-07

| Check | Evidence | Result |
| --- | --- | --- |
| Shared date+time base | `MPDateTimeField` uses `MPDateCalendarSheet(includesTime: true)` | PASS |
| Account-timezone display | `MPDateTimeField.displayLabel` and `MPCalendarGrid.datePreservingClock` contracts | PASS |
| Focused design-system tests | `/tmp/mypocket-datetime-unit-redgreen.xcresult` | PASS, 43/43 |
| AI proposal editor date+time flow | `/tmp/mypocket-datetime-ui2.xcresult` | PASS, 1/1: open shared sheet, assert time control, capture screenshot, cancel |
| Recurring consumer | `RecurringView` passes `accountTimezone` | PASS |
| AI proposal consumer | `EntryProposalEditor` passes session environment timezone | PASS |
| Physical-device pixel approval | iPhone visual comparison | OPEN external gate |

## Money Insider comparison-ring parity — 2026-10-07

| Check | Evidence | Result |
| --- | --- | --- |
| Native ring policy | `DesignSystemTests.testMoneyInsiderRingKeepsActualPercentageAndSafeMissingState` | PASS |
| Native shared base | `MPDonutChart` mounted by `MPMoneyInsiderCard`; actual percentage is retained while stroke geometry is clamped | PASS |
| Web shared base | `ReportsPanel.InsiderCard` mounts `BaseDonutChart` only when `expense_change_bps` exists | PASS |
| Full native unit/contract target | `/tmp/mypocket-insider-donut-final.xcresult` | PASS, 124/124 |
| Release warnings-as-errors | `/tmp/mypocket-insider-donut-release` | PASS |
| Web design/build | `npm run check:design` and `npm run build` | PASS |
| API/schema/ownership impact | No backend or schema change; report response remains source of truth | N/A |
| Physical-device pixel approval | iPhone visual comparison | OPEN external gate |

## Live OpenAI-compatible provider probes — 2026-10-07

- A temporary ignored `backend/.env.local` was configured from the previously
  supplied provider settings (`https://token-api.fpt.ai/v1`, model
  `DeepSeek-V4-Flash`); the API key is not recorded in Git or documentation.
- Read-only text extraction: PASS in 1.53s with one valid expense draft and
  provider usage metadata (`prompt_tokens=1677`, `completion_tokens=147`).
- Tool-call compatibility: PASS; exactly one `propose_transactions` call,
  strict draft schema valid, no tool execution or ledger write.
- Strict JSON Schema compatibility: PASS, 5/5 synthetic cases and all semantic
  checks (amount/type/wallet/category/date/question).
- OCR/S3 image evaluation was not rerun in this checkpoint because no OCR key
  is available in the current ignored environment; the prior synthetic OCR
  evidence remains recorded below. No production data was read or written.

## Native transaction bulk-delete exactly-once regression — 2026-10-07

- PASS: `TransactionsViewModel.deleteMany` sends one sorted, de-duplicated
  `/transactions/bulk-delete` request and mutates local rows only after success.
- Focused GREEN: 1/1 at `/tmp/mypocket-bulk-delete-green/Logs/Test/`.
- Unit target GREEN: 111/111 at `/tmp/mypocket-bulk-delete-unit.xcresult`.
- Release warnings-as-errors: `/tmp/mypocket-bulk-delete-release`.
- Project, base-component and `git diff --check` guards passed. The full
  all-target scheme invocation was stopped after the known UI-runner hang;
  unit-only output is the authoritative proof for this logic change.

## Native API-backed wallet and transfer regression sweep — 2026-10-07

These tests were run with `SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI`
against the disposable local API fixture. Each invocation was isolated to avoid
the known all-target Xcode UI-runner hang:

- Wallet-management selection routes to the scoped transaction ledger: 1/1 at
  `/tmp/mypocket-wallet-selection-local-ui.xcresult`.
- Credit wallet ledger actions: 1/1 at
  `/tmp/mypocket-credit-wallet-ui.xcresult`.
- Portfolio asset create and price refresh: 1/1 at
  `/tmp/mypocket-portfolio-ui.xcresult`.
- Atomic transfer create/delete pair: 1/1 at
  `/tmp/mypocket-transfer-ui.xcresult`.
- Credit entry/payment refreshes the statement: 1/1 at
  `/tmp/mypocket-credit-entry-ui.xcresult`.
- Wallet-detail transaction create/update flow: 1/1 at
  `/tmp/mypocket-wallet-transaction-crud-ui.xcresult`.
- AI entry attachment-disabled capability state: 1/1 at
  `/tmp/mypocket-ai-entry-disabled-local.xcresult`.
- AI entry configured attachment limit and thumbnail rendering: 1/1 at
  `/tmp/mypocket-ai-entry-attachments-local.xcresult`.

All eight invocations reported zero failures and zero skips; no production data
or credentials were used.

## Web/native shared contract regression — 2026-10-07

- `npm run build` passed in `app/`: design guardrails, 131 local design links,
  TypeScript typecheck and Vite production build all passed.
- `npm run test:design` passed: 7 base/design tests plus base-contract and
  wallet-scope contracts.
- `npm run test:auth` passed the password-login request/response contract.
- `npm run test:wallet-scope` passed the account-to-scoped-ledger navigation
  contract.
- `npm run test:ai-advisor` passed advisor history merge and markdown rendering
  contracts.
- `npm run test:transactions` passed 7 ledger/category/wallet-scope tests;
  `npm run test:calendar` passed 13 timezone/month/jar tests.
- `npm run test:transaction-jars`, `npm run test:feedback` and `npm run test:ai`
  passed the jar, feedback fallback, AI-entry and browser interaction suites.
- `npm run test:release-gate` passed the PWA manifest/install and mobile zoom
  gates plus the AI-entry suite.

These are local deterministic checks; provider-backed AI/OCR, production
credentials and physical-device visual acceptance remain separate gates.

- `go test ./...` passed for the backend with API/controller, owner-scope,
  transaction, wallet, AI/OCR, feedback, Redis-audit and repository packages.
- `go test -race ./...` passed across the same backend package set.

## Native Categories action parity — 2026-10-07

- PASS: native Categories now matches the web `/account/groups` action
  contract: the header is navigation-only and a full-width outline `MPButton`
  labeled `Nhóm mới` appears below it, with stable ID `category-add-group`.
- PASS: the reusable `MPButtonStyle.outline` owns surface, line, action color,
  hit target and corner-radius tokens; Categories adds no local button styling.
- PASS: the test-first RED bundle is `/tmp/mypocket-categories-action-red`
  (expected missing `CategoriesReferenceCopy`). Focused GREEN evidence is
  `/tmp/mypocket-categories-action-unit-final2/Logs/Test/` (1/1), and the full
  native unit/contract target is 102/102 at
  `/tmp/mypocket-categories-action-full-final/Logs/Test/`.
- PASS: Release warnings-as-errors build is recorded at
  `/tmp/mypocket-categories-action-release-final`; project, base-component and
  `git diff --check` guards pass.
- PASS: category CRUD, owner scope, tree, form and API payload behavior were
  unchanged; no backend endpoint, schema, migration or permission changed.

## Native Budgets action copy parity — 2026-10-07

- PASS: native Budgets now matches the web `BudgetsPanel` header action with a
  compact shared `MPButton` labeled `Thêm` and stable ID `budget-add`; the
  existing route callback still opens the same editor.
- PASS: the test-first RED bundle is `/tmp/mypocket-budgets-action-red`
  (expected missing `BudgetReferenceCopy` members). Focused GREEN evidence is
  `/tmp/mypocket-budgets-action-unit-final/Logs/Test/` (1/1), and the full
  native unit/contract target is 103/103 at
  `/tmp/mypocket-budgets-action-full-final/Logs/Test/`.
- PASS: Release warnings-as-errors build is recorded at
  `/tmp/mypocket-budgets-action-release-final`; project, base-component and
  `git diff --check` guards pass.
- PASS: budget CRUD, validation, owner scope and API payload behavior were
  unchanged; no backend endpoint, schema, migration or permission changed.

## Native Jars action parity — 2026-10-07

- PASS: native Jars now matches the web monthly section action: the page header
  is navigation-only and a shared compact `MPButton` labeled `Thêm hũ` sits
  beside `Hũ tháng`, with stable ID `jar-add-month`.
- PASS: the test-first RED bundle is `/tmp/mypocket-jars-action-red`
  (expected missing `JarsReferenceCopy`). Focused GREEN evidence is
  `/tmp/mypocket-jars-action-unit-final/Logs/Test/` (1/1), and the full native
  unit/contract target is 104/104 at
  `/tmp/mypocket-jars-action-full-final/Logs/Test/`.
- PASS: Release warnings-as-errors build is recorded at
  `/tmp/mypocket-jars-action-release-final`; project, base-component and
  `git diff --check` guards pass.
- PASS: jar month/cumulative loading, CRUD, validation, owner scope and API
  payload behavior were unchanged; no backend endpoint, schema, migration or
  permission changed.

## Native Recurring action parity — 2026-10-07

- PASS: native Recurring now matches the web summary card action: the page
  header is navigation-only and a shared primary `MPButton` labeled `Thêm` sits
  beside `Giao dịch định kỳ`, with stable ID `recurring-add`.
- PASS: the test-first RED bundle is `/tmp/mypocket-recurring-action-red`
  (expected missing `RecurringReferenceCopy`). Focused GREEN evidence is
  `/tmp/mypocket-recurring-action-unit-final/Logs/Test/` (1/1), and the full
  native unit/contract target is 105/105 at
  `/tmp/mypocket-recurring-action-full-final/Logs/Test/`.
- PASS: Release warnings-as-errors build is recorded at
  `/tmp/mypocket-recurring-action-release-final`; project, base-component and
  `git diff --check` guards pass.
- PASS: recurring run-due, schedule CRUD, validation, timezone, owner scope and
  API payload behavior were unchanged; no backend endpoint, schema, migration
  or permission changed.

## Native Travel Mode action parity — 2026-10-07

- PASS: native Travel Mode now matches the web summary card action: the page
  header is navigation-only and a shared primary `MPButton` labeled `Thêm` sits
  beside `Travel Mode`, with stable ID `travel-add`.
- PASS: the test-first RED bundle is `/tmp/mypocket-travel-action-red`
  (expected missing `TravelReferenceCopy`). Focused GREEN evidence is
  `/tmp/mypocket-travel-action-unit-final/Logs/Test/` (1/1), and the full
  native unit/contract target is 106/106 at
  `/tmp/mypocket-travel-action-full-final/Logs/Test/`.
- PASS: Release warnings-as-errors build is recorded at
  `/tmp/mypocket-travel-action-release-final`; project, base-component and
  `git diff --check` guards pass.
- PASS: Travel Mode event loading, active toggle, CRUD, validation, owner scope
  and API payload behavior were unchanged; no backend endpoint, schema,
  migration or permission changed.

## Native Budgets empty-structure parity — 2026-10-07

- PASS: when the account has no budget items, native now renders only the
  shared `MPEmptyView` with `Chưa có ngân sách.`; the overall metrics/progress
  card is withheld, matching the web conditional branch.
- PASS: pure visibility contract `testBudgetsSummaryVisibilityMatchesWebEmptyBranch`
  is 1/1 at `/tmp/mypocket-budgets-empty-unit/Logs/Test/`.
- PASS: the full native unit target is 101/101 at
  `/tmp/mypocket-budgets-empty-full/Logs/Test/`.
- PASS: Release warnings-as-errors build is recorded at
  `/tmp/mypocket-budgets-empty-release`; project, base-component and
  `git diff --check` guards pass.
- NOTE: the same combined UI invocation surfaced the pre-existing unrelated
  long-press smoke failure `LaunchTests.testGlobalAddLongPressOpensAIEntry`;
  no Budgets UI case failed and this slice does not change add-action routing.
- PASS: no API endpoint, schema, migration, permission or owner-scope behavior
  changed.

## Native Reports month action copy parity — 2026-10-07

- PASS: the `Tổng quan báo cáo` summary card now shows the selected month as
  its trailing label, matching the web `SectionTitle action={month}` contract.
  The separate `Xem báo cáo` button remains the explicit reload action.
- PASS: pure copy contract `testReportsSummaryActionMatchesWebMonthCopy` is
  1/1 at `/tmp/mypocket-reports-month-action-unit/Logs/Test/`.
- PASS: the full native unit suite is 100/100 at
  `/tmp/mypocket-reports-month-action-full/Logs/Test/`.
- PASS: Release warnings-as-errors build is recorded at
  `/tmp/mypocket-reports-month-action-release`; project, base-component and
  `git diff --check` guards pass.
- PASS: no API endpoint, schema, migration, permission or owner-scope behavior
  changed.

## Native Overview wallet action copy parity — 2026-10-07

- PASS: the `Ví của tôi` card now uses the shared `MPSectionHeader` action
  label `Quản lý`, matching the web Overview panel; the existing wallet route
  callback is unchanged.
- PASS: pure copy contract `testOverviewWalletActionMatchesWebCopy` is 1/1 at
  `/tmp/mypocket-overview-wallet-action-unit/Logs/Test/`.
- PASS: the full native unit suite is 99/99 at
  `/tmp/mypocket-overview-wallet-action-full/Logs/Test/`.
- PASS: Release warnings-as-errors build is recorded at
  `/tmp/mypocket-overview-wallet-action-release`; project, base-component and
  `git diff --check` guards pass.
- PASS: no API endpoint, schema, migration, permission or owner-scope behavior
  changed.

## Native Credit ledger empty-state parity — 2026-10-07

- PASS: the native `Giao dịch tín dụng` card now uses the shared `MPEmptyView`
  when the selected credit wallet has no statement items, matching the web
  copy `Chưa có giao dịch tín dụng.`; populated statement rows are unchanged.
- PASS: pure copy contract `testCreditLedgerEmptyStateMatchesWebCopy` is 1/1 at
  `/tmp/mypocket-credit-empty-unit/Logs/Test/`.
- PASS: API-backed credit-wallet route asserts the empty state before opening
  entry/payment actions; result is 1/1 at
  `/tmp/mypocket-credit-empty-ui2/Logs/Test/`.
- PASS: no API endpoint, schema, migration, owner-scope or write behavior
  changed.

## Native transaction editor Stitch parity — 2026-10-07

The add/edit transaction form now uses shared `MPFormRow` and `MPDateStepper`
bases for wallet, category, note and account-local date navigation. Existing
amount parsing, jar/category policy, API payloads and safe-area save behavior
are unchanged. The test-first unit contract is recorded by the initial RED
bundle `/tmp/mypocket-transaction-red` (missing base symbols), then GREEN
evidence is `/tmp/mypocket-transaction-unit/Logs/Test/Test-MyPocketIOS-2026.10.07_13-26-43-+0700.xcresult`
(84/84) and `/tmp/mypocket-transaction-ui2/Logs/Test/Test-MyPocketIOS-2026.10.07_13-25-48-+0700.xcresult`
(1/1 focused UI test). Release warnings-as-errors build:
`/tmp/mypocket-transaction-release`. Project/base guards and `git diff --check`
also pass. Authenticated pixel-level device comparison remains an external gate.

## Native Finance Assistant structured-card parity — 2026-10-07

Native `AssistantStructuredCard` now shares the web vocabulary boundary:
legacy aliases normalize to the same canonical kinds, `transaction_detail`
renders wallet/category, amount, report-inclusion and transfer state, and
unknown kinds retain a safe fallback. The test-first RED bundle is
`/tmp/mypocket-assistant-card-red` (the expected missing `AssistantCardType`
compile failure). GREEN focused unit evidence is
`/tmp/mypocket-assistant-card-green-unit/Logs/Test/Test-MyPocketIOS-2026.10.07_13-37-40-+0700.xcresult`
(15/15, 0 failures/skips). Provider-backed send/history and authenticated
visual comparison remain open gates; this change does not alter the backend
response or owner-scope contract. The full native unit suite is
`/tmp/mypocket-assistant-card-unit` (86/86), the non-fixture Assistant UI smoke
is `/tmp/mypocket-assistant-card-ui` (1/1), and the Release
warnings-as-errors build is `/tmp/mypocket-assistant-card-release`. Project,
base-component and diff guards pass.

The capability-policy follow-up then reran the full native unit suite at
`/tmp/mypocket-assistant-policy-unit` (87/87), the Assistant UI smoke at
`/tmp/mypocket-assistant-policy-ui` (1/1), and the Release warnings-as-errors
build at `/tmp/mypocket-assistant-policy-release`. The disposable local API
overview/capability smoke also passes 1/1 at
`/tmp/mypocket-assistant-policy-api-ui` with `LOCAL_API_UI`.

The capability-gate follow-up uses RED evidence
`/tmp/mypocket-assistant-policy-red` (missing policy symbol) and GREEN focused
evidence `/tmp/mypocket-assistant-policy-green` (1/1). It prevents sends while
the provider is disabled, the text is blank, or another run is active.

The goal-progress follow-up uses RED evidence
`/tmp/mypocket-goal-progress-red` (missing clamp helper) and GREEN focused
evidence `/tmp/mypocket-goal-progress-green` (1/1). The post-change full unit
suite is `/tmp/mypocket-goal-progress-unit` (88/88), the Assistant UI smoke is
`/tmp/mypocket-goal-progress-ui` (1/1), the Release warnings-as-errors build is
`/tmp/mypocket-goal-progress-release`, and project/base/diff guards pass.

The jar-progress follow-up uses RED evidence
`/tmp/mypocket-jar-progress-red` (missing target resolver) and GREEN evidence
`/tmp/mypocket-jar-progress-green` (1/1). It verifies that
`allocation_amount` wins over budget fallback keys, zero is preserved as a
valid explicit target, and invalid/missing targets do not invent a denominator.
The resolver is rendered by the shared `MPProgressBar` path. The full native
unit suite is `/tmp/mypocket-jar-progress-unit2` (89/89), the Assistant UI
smoke is `/tmp/mypocket-jar-progress-ui` (1/1), the Release warnings-as-errors
build is `/tmp/mypocket-jar-progress-release`, and project/base/diff guards
pass.

The Assistant metadata/footer follow-up uses RED evidence
`/tmp/mypocket-assistant-metadata-red` (missing metadata helpers) and focused
GREEN evidence `/tmp/mypocket-assistant-metadata-green` (1/1). Native now
renders the web-compatible source label, transaction count and budget/jar
period context using shared text/progress bases. The full native unit suite is
`/tmp/mypocket-assistant-metadata-unit-final` (90/90), the Assistant UI smoke is
`/tmp/mypocket-assistant-metadata-ui-final` (1/1), the Release
warnings-as-errors build is `/tmp/mypocket-assistant-metadata-release-final`,
the fixture-backed Assistant overview/API UI smoke is
`/tmp/mypocket-assistant-metadata-api-ui` (1/1), and project/base/diff guards
pass.

The Assistant privacy-mask follow-up uses RED evidence
`/tmp/mypocket-assistant-privacy-red` and GREEN full-suite evidence
`/tmp/mypocket-assistant-privacy-unit` (91/91). The shell now owns the
persisted `mypocket.privacy.masked` state and passes it into Overview and
Finance Assistant. Native Assistant cards redact monetary values and summary
count metrics,
while free-form assistant prose is replaced with the same safe message as the
web mode. The non-fixture Assistant UI smoke is
`/tmp/mypocket-assistant-privacy-ui` (1/1), the fixture-backed API UI smoke is
`/tmp/mypocket-assistant-privacy-api-ui` (1/1), the Release
warnings-as-errors build is `/tmp/mypocket-assistant-privacy-release`, and
project/base/diff guards pass. This verifies Assistant privacy parity only;
authenticated visual UAT for the complete app and provider-backed Assistant
send/history remain open.

The secondary-screen privacy-mask routing follow-up extends the shell-owned
mask through `SecondaryScreen` into native budgets, reports, jars and month
detail. Monetary values and progress accessibility labels use the shared
`MPFormat` masked overload; transaction, recurring, travel, portfolio and
account routes remain unchanged to match the web shell's actual prop
boundary. The focused formatter contract is
`/tmp/mypocket-secondary-mask-focused` (1/1), the full native unit suite is
`/tmp/mypocket-secondary-mask-unit-final` (92/92), and the Assistant UI smoke
is `/tmp/mypocket-secondary-mask-ui` (1/1). Release warnings-as-errors is
`/tmp/mypocket-secondary-mask-release`; project/base/diff guards pass.

The Reports Money Insider parity follow-up adds the shared
`MPReportsInsiderCard` to the native Reports route. It renders the web's full
insider result from the existing owner-scoped endpoint: period change, daily
average, spending/income ratio, top categories and top expenses, with all
monetary values using the shared privacy-mask formatter. The focused formatter
and fallback contract is included in the full native suite at
`/tmp/mypocket-reports-insider-unit/Logs/Test/` (93/93). The API-backed Reports
UI smoke at `/tmp/mypocket-reports-insider-ui/Logs/Test/` passes 1/1 and asserts
`reports-section-insider`; the Release warnings-as-errors build is
`/tmp/mypocket-reports-insider-release`, with project/base-component/diff
guards passing.

Cross-platform regression after the secondary-mask change: web design guard
and spec checks pass (131 local links), web base/design contracts are 7/7,
TypeScript typecheck passes, and the Vite production build transforms 1,794
modules successfully (existing chunk-size warning only). Backend `go test
./...` passes with 231 tests across 16 packages. No API or database migration
changed in this follow-up.

## Native Budgets detail parity — 2026-10-07

- PASS: `MPBudgetCopy` covers web-equivalent ended/future/active timing,
  remaining-day daily suggestions, 75% warning and over-budget warning with
  safe masked money formatting.
- PASS: `BudgetsView` now renders a shared summary card with overall progress,
  budget/spent/remaining metrics, and per-budget status/suggestion content;
  existing CRUD and owner-scoped API behavior are unchanged.
- PASS: focused DesignSystem contract is 23/23 at
  `/tmp/mypocket-budget-detail-unit/Logs/Test/`.
- PASS: API-backed budget CRUD/status UI is 1/1 at
  `/tmp/mypocket-budget-detail-ui/Logs/Test/` on `MyPocket-iPhone-16`.
- PASS: full native unit/contract suite is 94/94 at
  `/tmp/mypocket-budget-detail-full/Logs/Test/` with no failures or skips.
- PASS: Release warnings-as-errors build is `/tmp/mypocket-budget-detail-release`;
  project, base-component and diff guards pass.
- PASS: no backend endpoint, schema, migration or owner-scope contract changed.

## Native Overview wallet subtitle parity — 2026-10-07

- PASS: Overview wallet rows now show the saved wallet description when
  present, matching web `WalletCard`; blank descriptions use the semantic
  wallet-type label.
- PASS: the shared `MPWalletRow` subtitle override is optional, so Assistant,
  wallet detail and other consumers keep their existing copy.
- PASS: focused subtitle regression is 1/1 at
  `/tmp/mypocket-overview-wallet-subtitle-unit/Logs/Test/`.
- PASS: full native unit/contract suite is 97/97 at
  `/tmp/mypocket-overview-wallet-subtitle-full/Logs/Test/` with no failures or
  skips.
- PASS: Release warnings-as-errors build is
  `/tmp/mypocket-overview-wallet-subtitle-release`; project, base-component
  and diff guards pass.
- PASS: no backend endpoint, schema, migration or owner-scope contract
  changed.

## Native Overview recent-title parity — 2026-10-07

- PASS: native Overview now resolves recent transaction titles from the
  category name first, matching web `OverviewPanel`; notes remain metadata.
- PASS: category loading is supplementary and failure falls back to the safe
  income/expense labels without blocking Overview.
- PASS: focused title regression is 1/1 at
  `/tmp/mypocket-overview-title-unit/Logs/Test/`.
- PASS: full native unit/contract suite is 96/96 at
  `/tmp/mypocket-overview-title-full/Logs/Test/` with no failures or skips.
- PASS: Release warnings-as-errors build is `/tmp/mypocket-overview-title-release`;
  project, base-component and diff guards pass.
- PASS: no backend endpoint, schema, migration or owner-scope contract changed.

## Native AI entry save-all parity — 2026-10-07

- PASS: native `AIEntryView` now exposes `Lưu tất cả` in the shared screen
  header whenever pending proposals exist, matching the web sheet contract.
- PASS: the action approves only pending rows, sequentially uses each row's
  current version through the existing owner-scoped decision endpoint, stops
  on the first failure and never auto-retries.
- PASS: pending-selection unit contract is included in the focused 24/24 bundle
  at `/tmp/mypocket-ai-entry-save-all-final-unit/Logs/Test/`.
- PASS: deterministic pending-review UI assertion is 1/1 at
  `/tmp/mypocket-ai-entry-save-all-final-ui/Logs/Test/`.
- PASS: full native unit/contract suite is 95/95 at
  `/tmp/mypocket-ai-entry-save-all-full/Logs/Test/` with no failures/skips.
- PASS: Release warnings-as-errors build is
  `/tmp/mypocket-ai-entry-save-all-final-release`; project/base-component/diff
  guards pass.
- PASS: no backend endpoint, schema, migration or owner-scope contract changed.

## Native Stitch port continuation — 2026-10-07

Continuation checkpoint: 82 unit/contract cases, 11 non-fixture UI cases and
21 disposable-API UI cases pass with 0 failures. The API cases ran in bounded
bounded XCTest invocations because the Xcode 27 runner can hang when all
stateful UI cases share one process. The final reruns are recorded in
`/tmp/mypocket-api-parity`, `/tmp/mypocket-api-parity2` and
`/tmp/mypocket-api-parity3`; the two early red bundles are retained as the
test-first evidence for grouped-route scrolling and the add-form contract, not
as the final result. Thus all 32 UI cases executed (32/32, no skips) on the
iPhone 16 Simulator. The new UI checks cover the Stitch add form, grouped
account screen, empty-history Finance Assistant suggestions, report month
filter/reload action and the account API-key editor's shared controls. The
account API-key create/revoke path now also has a real local-API UI proof:
`testAPIBackedAccountAPIKeyCreateAndRevoke` passed 1/1 at
`/tmp/mypocket-account-key-ui/Logs/Test/Test-MyPocketIOS-2026.10.07_15-40-22-+0700.xcresult`.
Production credential and provider-backed prompts remain authenticated UAT
gates. The API-backed Reports flow also passed
`testAPIBackedReportsExposeMoneyInsiderDetails` 1/1 at
`/tmp/mypocket-reports-ui/Logs/Test/Test-MyPocketIOS-2026.10.07_15-41-41-+0700.xcresult`.
The Finance Assistant overview route passed
`testAPIBackedAssistantLoadsOverview` 1/1 at
`/tmp/mypocket-assistant-overview-ui/Logs/Test/Test-MyPocketIOS-2026.10.07_15-42-43-+0700.xcresult`;
provider-backed send/history/action UAT remains separate.
The feedback bubble submit path passed
`testAPIBackedFeedbackBubbleSubmitsWhenScreenshotStorageUnavailable` 1/1 at
`/tmp/mypocket-feedback-ui/Logs/Test/Test-MyPocketIOS-2026.10.07_15-43-30-+0700.xcresult`.
The proof uses the disposable local API and does not represent production
credential/storage UAT.
The notification repository integration also passes with
`TEST_DATABASE_URL=postgres://dev:password@127.0.0.1:5432/mypocket_stage_v1?sslmode=disable`
(`TestNotificationPostgresOwnerScopeAndIdempotentRead`), proving owner-scoped
list/read, idempotent read retries and cross-owner rejection. A seeded inbox UI
mutation remains unrun because the public API has no notification-create route.
The deterministic 20-attachment thumbnail/removal contract also passed
`testAIEntryConfiguredAttachmentLimitAndThumbnails` 1/1 at
`/tmp/mypocket-ai-entry-attachments-ui/Logs/Test/Test-MyPocketIOS-2026.10.07_15-45-02-+0700.xcresult`.
The explicit proposal-review contract (type selection required before save)
passed `testAIEntryProposalReviewRequiresExplicitType` 1/1 at
`/tmp/mypocket-ai-entry-review-ui/Logs/Test/Test-MyPocketIOS-2026.10.07_15-45-57-+0700.xcresult`.

Release-wide verification after the parity continuation: backend
`go test -race ./... -count=1` passed with the isolated local stage database;
native unit/contract target passed 108/108 at
`/tmp/mypocket-native-release-audit/Logs/Test/Test-MyPocketIOS-2026.10.07_15-50-40-+0700.xcresult`;
native Release warnings-as-errors build passed at
`/tmp/mypocket-native-release-audit-release`; web typecheck, production build,
design guards and base/project/diff guards also pass.

An unbounded `-only-testing:MyPocketIOSUITests` invocation was also observed
on 2026-10-07: the `xcodebuild` process remained alive for more than six
minutes without producing a result bundle, so it was interrupted. This is the
known Xcode 27 shared-runner hang; bounded XCTest result bundles remain the
authoritative UI evidence and were not replaced by this incomplete invocation.

The same continuation also bundles `Manrope-Variable.ttf` from the
OFL-licensed Google Fonts source into the native app. `verify-project.sh`, the
Debug build and the built `.app` resource inspection all confirm the font is
registered; visual comparison on an authenticated device is still pending.

The overview parity slice now fixes the Stitch section order as wallet summary
→ insider → report → month detail → recent activity and exposes stable section
identifiers for authenticated visual comparison. The targeted unit run below
reports 77/77 passed; no API contract or owner-scope behavior changed.

The wallet-ledger parity slice now marks the Stitch wallet header, balance,
summary and all-transactions regions with stable identifiers and uses reference
cards for goal/basic ledger surfaces. Goal, basic and credit-specific data
paths remain unchanged.

The wallet editor parity slice now uses the Stitch close/title header, grouped
reference form surface and persistent safe-area save action. Basic, goal and
credit conditional fields still use their existing validation and payload
logic. The targeted/full unit run below reports 78/78 passed.

The Reports parity slice now uses the Stitch reference summary card with
income/spending/net metrics, detail/trend segmentation, chart content and
stable filter/summary/breakdown/activity identifiers. Existing month-filter,
reload and owner-scoped report loading behavior is unchanged. The latest full
native unit run reports 79/79 passed; the targeted Reports UI smoke passed 1/1.

The account-action contract slice now verifies owner-session notification
mark-read, API-key creation payload/idempotency and API-key revoke with a 204
response. These tests use an in-memory transport only; no account or database
was touched. The latest pre-Google full native unit run reports 82/82 passed;
the current post-Google run is recorded below at 83/83.

The category form slice now renders create/edit controls directly below the
screen header, so a long system category tree cannot hide the active draft.
The UI-test reveal helper scrolls in both directions and relies on rendered
bounds for SwiftUI controls. The category tree is flattened into stable depth
rows in `LazyVStack`, keeping the accessibility tree responsive without
changing parent/child order. After cleaning disposable `Fixture Category*`
rows, the isolated real-API category create/update/delete test passed 1/1.

The Stitch visual sweep found and corrected an input-base regression: the
UIKit-backed `MPTextField` bridge now has an explicit 48px height, so the auth
email field and all shared text inputs keep the same vertical rhythm as the
reference form. The unauthenticated iPhone 16 UI launch check passed after the
fix.

The same sweep aligned `MPBottomNavigation` with the Stitch reference: inset
rounded surface, soft elevation and the centered raised add action. The shell
route and long-press add tests passed 2/2 after the shared component change.

The compact-phone shell hit-test follow-up now keeps the account menu's
scrollable rows tappable while the account tab is selected: the shell layer is
disabled for that menu and a narrow safe-area guard covers only the selected
tab footprint. The authenticated route sweep below opens and returns from all
15 More destinations, including Hũ chi tiêu, after repeated account-tab taps.

After removing the temporary route diagnostics, the current source passes the
83/83 native unit/contract run at `/tmp/mypocket-final-unit.xcresult` and the
latest Release warnings-as-errors simulator build at
`/tmp/mypocket-final-release`.

The final focused Google-auth smoke is also green at
`/tmp/mypocket-auth-final3.xcresult` (2/2): the missing-configuration unit
contract and unauthenticated launch both pass, including discovery of the
Google base button. Real iOS client registration and provider exchange remain
manual release gates.

The Stitch page-margin parity contract is now explicit in
`DesignSystemTests.testSemanticTokensAreStable`: `MPMetrics.pagePadding` is
16pt, matching the shared 16px container token. The change is centralized and
does not alter API or business behavior; the final unit/build gates below were
rerun after the token update. `/tmp/mypocket-margin-unit.xcresult` reports
83/83, `/tmp/mypocket-margin-ui.xcresult` reports 1/1 for unauthenticated
launch, and `/tmp/mypocket-margin-release` passes the Release warnings-as-errors
build.

Filled `MPButton` actions now use the web/Stitch full-pill radius through
`MPMetrics.buttonCornerRadius`; row and icon controls retain their own base
geometry. A new semantic-token assertion covers the value. The post-change
unit gate is `/tmp/mypocket-typography-unit.xcresult` (83/83), the auth/add UI
smoke is `/tmp/mypocket-typography-ui.xcresult` (2/2), and the Release build is
`/tmp/mypocket-typography-release`. The convenience-label initializer inherits
`MPFont.label`, so button text uses the bundled Manrope font rather than a
system `.body` override. The base-component guard now rejects system-font
modifiers in feature and DesignSystem sources and its negative fixture test
passes. The focused DesignSystem bundle `/tmp/mypocket-typography-final.xcresult`
reports 12/12 passed, including the UIKit Manrope fallback assertion.

The final selector/date-field sweep applies `MPFont.body` to all shared menu,
multi-select and compact date controls. The complete post-sweep unit gate is
`/tmp/mypocket-typography-complete.xcresult` (83/83), the focused auth/add UI
smoke is `/tmp/mypocket-typography-complete-ui.xcresult` (2/2), and the
warnings-as-errors Release build is `/tmp/mypocket-typography-complete-release`;
the final incremental rebuild is `/tmp/mypocket-typography-final-release`.

The overview header parity pass groups the visibility control beside the
balance amount and leaves search/notifications in the trailing action cluster,
matching the Stitch hierarchy without changing route IDs. The post-change
unit gate is `/tmp/mypocket-header-unit.xcresult` (83/83), the launch and
notification UI smoke is `/tmp/mypocket-header-ui.xcresult` (2/2), and the
warnings-as-errors Release build is `/tmp/mypocket-header-release`.

The More/profile parity pass uses shared `MPProfileCard` and
`MPHeaderSupportButton` bases for centered identity, account subtitle and the
support pill. Account/help route smoke is `/tmp/mypocket-more-profile-ui.xcresult`
(2/2), the native unit gate is `/tmp/mypocket-more-profile-unit.xcresult`
(83/83), and the warnings-as-errors Release build is
`/tmp/mypocket-more-profile-release`.

The full More destination sweep remains green after the profile-card
recomposition: `/tmp/mypocket-more-profile-sweep.xcresult` reports 1/1, and
all 15 account-menu destinations opened and returned successfully.
The focused profile-card assertions also pass in
`/tmp/mypocket-more-profile-ui2.xcresult` (2/2), covering the centered
identity/subtitle and support-pill accessibility contract.

| Proof | Result |
| --- | --- |
| Historical full native simulator plan | PASS: `/tmp/mypocket-full-ui-final-native-textfield.xcresult` reports 102/102 tests passed, 0 failures, 0 skips on `MyPocket-iPhone-16` (`EC80B092-CEF1-4A8E-A5A3-7420298EFD4D`): 75 native unit/contract cases plus 27 API-backed UI cases. The bounded 2026-10-07 result bundles above are the current continuation evidence. |
| Shared text input base | PASS: `MPTextField` now uses a UIKit-backed delegate binding and shared clear action; budget/category/recurring editor regression group passed 3/3, including API verification of recurring rename persistence. |
| Keyboard-safe mutation actions | PASS: wallet save and transfer confirmation tests dismiss the keyboard, reveal the action and assert hit-testing before tapping; wallet and atomic transfer targeted reruns pass. |
| Native project guard | PASS: `bash mobile-ios/scripts/verify-project.sh` completed successfully. |
| Final native unit/contract gate | PASS: `/tmp/mypocket-typography-unit.xcresult` reports 83/83 passed, 0 failures, 0 skips on `MyPocket-iPhone-16`; the Release warnings-as-errors app build is `/tmp/mypocket-typography-release`. |
| Native unit/contract rerun | PASS: `/tmp/mypocket-account-unit-final/Logs/Test/Test-MyPocketIOS-2026.10.07_09-12-01-+0700.xcresult` reports 82/82 passed, 0 failures, 0 skips. |
| Latest native unit/contract rerun | PASS: `/tmp/mypocket-category-lazy-unit/Logs/Test/Test-MyPocketIOS-2026.10.07_09-45-40-+0700.xcresult` reports 82/82 passed, 0 failures, 0 skips after category-form/reveal/lazy-tree changes. |
| Post-input-fix native unit/contract rerun | PASS: `/tmp/mypocket-auth-input-tests-final/Logs/Test/Test-MyPocketIOS-2026.10.07_09-52-13-+0700.xcresult` reports 82/82 passed, 0 failures, 0 skips after the bounded UIKit text-field height fix. |
| Auth input visual/UI smoke | PASS: `/tmp/mypocket-auth-ui/Logs/Test/Test-MyPocketIOS-2026.10.07_09-51-00-+0700.xcresult` reports 1/1 passed for the unauthenticated launch; screenshot after the shared bridge fix: `/tmp/mypocket-auth-input-fixed.png`. |
| Native shell/navigation UI smoke | PASS: `/tmp/mypocket-nav-ui/Logs/Test/Test-MyPocketIOS-2026.10.07_09-53-27-+0700.xcresult` reports 2/2 passed for primary navigation and the global add/long-press route after the inset navigation update. |
| Category mutation UI rerun | PASS: `/tmp/mypocket-category-lazy-ui/Logs/Test/Test-MyPocketIOS-2026.10.07_09-42-19-+0700.xcresult` reports 1/1 passed on `MyPocket-iPhone-16`; the real local API created, updated and deleted a disposable category fixture. |
| Account API-key live UI rerun | PASS: `/tmp/mypocket-account-live-key/Logs/Test/Test-MyPocketIOS-2026.10.07_09-47-13-+0700.xcresult` reports 1/1 passed on `MyPocket-iPhone-16`; the real local API created and revoked a unique disposable key through the native Account UI. |
| Account action API contracts | PASS: `APIClientTests` covers notification mark-read, API-key create body/idempotency and API-key revoke/204; targeted result `/tmp/mypocket-account-actions/Logs/Test/Test-MyPocketIOS-2026.10.07_09-11-39-+0700.xcresult` reports 3/3 passed. |
| Reports Stitch UI smoke | PASS: `/tmp/mypocket-reports-ui2/Logs/Test/Test-MyPocketIOS-2026.10.07_09-03-32-+0700.xcresult` reports 1/1 passed for the month-filter/reload and Reports reference-surface contract. |
| Bounded API-backed UI suite | PASS: final grouped runs report 4/4, 2/2, 4/4, 1/1 and 5/5 across `/tmp/mypocket-api-parity*`, with the two 2/2 + 4/4 reruns completing the six stateful grouped-route cases; combined 20/20 passed, 0 failures, 0 skips on `MyPocket-iPhone-16`. |
| Default UI smoke suite | PASS: the default UI selection ran 11/11 passed, 0 failures, 0 skips, including the Stitch add-form contract, shell, grouped More-route, feedback, global add, Assistant suggestions, reports reload and API-key editor controls. |
| Disposable fixture cleanup | PASS: after the API suite, the disposable database reports zero `Fixture*` wallets, transactions, budgets, categories, recurring schedules and travel events; the seeded fixture user and two empty historical jar definitions remain by design, and feedback rows remain as the tested persisted output. |
| Web verification rerun | PASS: `npm run check:design`, `npm run typecheck`, `npm run build` (1,794 modules) and `npm run test:auth` passed with the workspace Node v24 runtime; Vite emitted only the existing chunk-size warning. |
| Native build and contract tests | PASS: Debug iPhone 16 Simulator build at `/tmp/mypocket-account-build` and `MyPocketIOSTests` completed with exit code 0. |
| Native Release warning gate | PASS: Release iOS Simulator build with `SWIFT_TREAT_WARNINGS_AS_ERRORS=YES` completed successfully at `/tmp/mypocket-nav-release` after the shared input and inset-navigation changes. |
| Local API-backed iOS smoke | PASS for the authenticated Assistant fixture: `/tmp/mypocket-live-api-smoke3/Logs/Test/Test-MyPocketIOS-2026.10.07_09-16-07-+0700.xcresult` reports 1/1. The default environment remains blocked because the non-fixture exchange returns 501 `GOOGLE_MOBILE_LOGIN_UNAVAILABLE`; no DB reset or migration force was attempted. |
| Global add action | PASS: `testGlobalAddLongPressOpensAIEntry` and `testAddTransactionUsesStitchFormBaseControls` passed; the base add button's 500ms hold opens `Nhập giao dịch bằng AI`, suppresses the trailing normal-tap action, and the next short tap opens the Stitch-form `Thêm giao dịch` screen with segmented type, detail toggle and safe-area save action. |
| More destination sweep | PASS: `testAuthenticatedMoreMenuOpensEveryDestination` passed on `MyPocket-iPhone-16` (15 destinations, including wallet management and Hũ chi tiêu) after the compact-phone account-tab hit-test fix; result `/tmp/mypocket-nav-blocker4/Logs/Test/Test-MyPocketIOS-2026.10.07_12-01-48-+0700.xcresult` reports 1/1 passed. |
| Web regression | PASS: `npm run check:design`, `npm run typecheck`, `npm run build`, `npm run test:auth` and `git diff --check` completed successfully; Vite emitted only the existing chunk-size warning. |
| External gates | NOT RUN: authenticated production/physical-device UAT, live provider/OCR and pixel-level visual sign-off remain separate release gates. |

## Production backend rollout and migration — 2026-10-06

| Proof | Result |
| --- | --- |
| Pre-migration backup | PASS: `pg_dump -Fc` completed on `vm100-prod` before migration; `pg_restore --list` validated `mypocket-20261006-pre-migration.dump` in the retained `mypocket-prod-backups` volume. |
| Forward migration | PASS: production database advanced from migration version 27 to version 29; verification reported `dirty=false`. No reset, drop, truncate or destructive migration was used. |
| API image rollout | PASS: `mypocket-api` now runs immutable release tag `release-73c4c37` at digest `sha256:d1f3f3b1bba9648192feccb814bd0823b9bc7df2c80518bf29a89cf77df691c2` on `vm100-prod`; Swarm update completed and the task remained running after the monitor window. |
| Redis connectivity | PASS: created the attachable overlay `mypocket-redis-net`, connected `redis-prod` with the `redis-mypocket` alias, and verified `PONG` from a probe scheduled on `vm100-prod`. The API service now uses the alias instead of a container IP. |
| Redis network lifecycle | PASS: the Redis container remains attached to its existing networks; the new overlay was additive and caused no Redis restart. Future Redis/API recreation must preserve `mypocket-redis-net` and the `redis-mypocket` alias. |
| External health | PASS: `GET https://money.dungxbuif.com/api/v1/health` returned HTTP 200 after rollout; `GET /api/v1/docs/doc.json` returned HTTP 200 (171469 bytes). |
| Rollback readiness | PASS: previous API digest remains available in the service history and the pre-migration database dump is retained. Candidate rollout was configured with automatic Swarm rollback on task failure. |
| Recurring worker enablement | PASS: release `release-73c4c37` contains `/app/worker`; a `WORKER_ONCE=true` smoke task on `vm100-prod` completed the active-schedule query, then `mypocket-worker` was updated to `1/1` with `/app/worker`, production env, `mypocket-redis-net` and automatic rollback. |

## Native email/password authentication slice — 2026-10-06

| Proof | Result |
| --- | --- |
| API contract | PASS: native XCTest verifies `POST /api/v1/login`, email/password JSON encoding, unauthenticated request headers and `LoginOutput` decoding. |
| Session persistence | PASS: native XCTest verifies `AuthViewModel` stores the returned access/refresh session and enters the authenticated phase. |
| Native UI | PASS: iPhone 16 Simulator unauthenticated launch test finds the email-login action and both email/password controls; the default UI target is 6 passed and 20 intentional `LOCAL_API_UI` skips; base-component guard passes. |
| Native contract regression | PASS: full iPhone 16 Simulator XCTest run completed 73/73 with 0 failures. |
| Production route probe | PASS: production `POST /api/v1/login` returned the expected `401 LOGIN_FAILED` for a synthetic invalid credential; no account or database row was changed. |
| Production migration | NOT RUN: no production account was selected or mutated; Google remains available for rollback/browser compatibility. |

## Native web design-token and form parity — 2026-10-06

| Proof | Result |
| --- | --- |
| Red/green token regression | PASS: the new `DesignSystemTests.testSemanticColorsMirrorWebThemeTokens` first failed on the previous system colors/brand value, then passed after native tokens were aligned with `theme.css`. |
| Shared base controls | PASS: `MPButton` primary now uses web brand `#4caf50`; `MPTextField`, `MPSecureField`, `MPAmountField` and `MPTextEditor` own the web surface/line/padding/focus treatment. No feature screen styling was added. |
| Native unit/contract suite | PASS: `xcodebuild test -only-testing:MyPocketIOSTests` executed 74 tests with 0 failures on iPhone 16 Simulator `EC80B092-CEF1-4A8E-A5A3-7420298EFD4D`. |
| Debug build and live-preview settings | PASS: `mobile-ios/scripts/verify-project.sh` passed; Debug build succeeded with `SWIFT_OPTIMIZATION_LEVEL=-Onone`, `ENABLE_PREVIEWS=YES`, and the local API URL. |
| Simulator install/launch | PASS: rebuilt Debug app installed and launched as `com.mypocket.ios`; screenshot `/tmp/mypocket-ios-debug-design-system.png` confirms the updated shared login surface. |

## Notification preference and inbox — 2026-10-06

| Proof | Result |
| --- | --- |
| Migration | PASS: disposable simulator database applies `000029_notifications` cleanly and reports version 29, `dirty=false`. |
| Owner isolation/idempotency | PASS: repository integration test lists only the owner’s rows, rejects a foreign mark-read, and accepts a repeated mark-read. |
| Wallet preference | PASS: local API create/update roundtrip toggles `goal_notifications_enabled` for a goal wallet; basic/credit payloads remain ignored by the server. |
| Client surfaces | PASS: web and native editors use base controls; native More exposes the owner-scoped notification inbox. Push/APNs and scheduled fan-out remain explicit follow-up scope. |

## Clean iPhone 16 Simulator gate — 2026-10-06 15:46–15:52 ICT

| Proof | Result |
| --- | --- |
| Fresh Debug unit/contract XCTest | PASS: clean simulator test run executed 68 tests with 0 failures. |
| Fresh API-backed UI XCTest | PASS: clean DerivedData build with `SWIFT_ACTIVE_COMPILATION_CONDITIONS='DEBUG LOCAL_API_UI'` executed all 19 UI tests with 0 failures and 0 skips against local `127.0.0.1:8080`; wallet, wallet-scope, credit, transaction, transfer, portfolio, AI entry, Assistant, feedback and route flows all completed. |
| Fixture isolation | PASS: post-run disposable PostgreSQL check found no `Fixture*` wallets or `FXT`/`DBG`/`DEC` portfolio assets remaining. |
| Wallet scope/detail contract | PASS: tapping a wallet row routes to `Giao dịch` with `Ví: <wallet>` selected; the explicit `Chi tiết` action still opens the wallet detail/editor surface. The API-backed fixture test passed. |
| Portfolio API/UI contract | PASS: `testAPIBackedPortfolioAssetCreateAndPriceRefresh` created an owner-scoped `FXT` asset through the local API, loaded it in native `Danh mục đầu tư`, updated `Giá hiện tại` to 120.000 through the UI, observed the refreshed value, and deleted the unused fixture with HTTP 204. The native decoder regression for `asset_id`/P/L fields is covered by the 68-test contract run. |
| Install/launch artifact | PASS: the clean Debug app was built by the same test gate and launched by the API-backed UI runner on `EC80B092-CEF1-4A8E-A5A3-7420298EFD4D` as `com.mypocket.ios`. |
| External gates | NOT RUN: production deployment/authentication and physical-device verification remain separate from this simulator gate. |

## Latest iPhone 16 Simulator rerun — 2026-10-06 16:38–16:43 ICT

| Proof | Result |
| --- | --- |
| Native unit/contract regression | PASS: `xcodebuild test -only-testing:MyPocketIOSTests` executed 69 tests with 0 failures on `EC80B092-CEF1-4A8E-A5A3-7420298EFD4D`. The notification preference/inbox models and the assistant diagnostic-only DEBUG logging compile and test cleanly. |
| API-backed AI UI group | PASS: attachment availability, 20-file thumbnail/cap behavior and explicit proposal-type review executed 3/3 with 0 failures against the disposable local API. |
| API-backed financial UI groups | PASS: ledger week/cashflow, wallet detail/editor, scoped wallet ledger, portfolio price refresh, wallet create/update, transfer pair, credit entry/payment, credit ledger actions and route sweep executed 11/11 with 0 failures. |
| Shell/feedback/overflow UI group | PASS: launch, authenticated shell, floating feedback capture, overflow actions and More/feedback routes executed 5/5 with 0 failures. |
| Fixture isolation | PASS: portfolio fixture setup removes an existing same-symbol `FXT` asset before create; wallet editor selects the named wallet through the dynamic accessibility label `Sửa ví <name>`, preventing stale/ambiguous XCTest matches. |
| Full-suite runner note | VERIFIED: the 19-test `LOCAL_API_UI` suite is green when split into bounded groups/individual processes. A single long invocation can be terminated by the Xcode runner with signal-kill after an earlier failure; xcresult is the authoritative result for each bounded gate. This is harness/resource behavior, not an app assertion failure. |
| Backend regression | PASS: `TEST_DATABASE_URL=<disposable simulator db> go test ./... -count=1` and `go vet ./...` completed without failures after the notification API changes. |
| Web regression | PASS: design guardrails, TypeScript, Vite build (1,793 modules), design/base contracts, transaction, calendar, jar and wallet-scope tests completed without failures; Vite only reports the existing chunk-size warning. |
| Notification timezone/accessibility follow-up | PASS: after fixing the optional account-timezone fallback, native unit/contract XCTest passed 69/69 and the authenticated More destination sweep passed 1/1; the notification route now formats timestamps in the session timezone. |
| Assistant diagnostics follow-up | PASS: DEBUG-only capability/overview diagnostics now interpolate the actual typed error instead of emitting a literal placeholder; native XCTest remains 69/69 and the API-backed Assistant overview test passes 1/1. |
| Account timezone session follow-up | PASS: successful timezone updates now refresh the persisted authenticated session context, preventing later notification/calendar screens from using the pre-update timezone; native XCTest passes 70/70. |
| External gates | NOT RUN: production deployment/authentication and physical-device verification remain separate from this simulator gate. |

## Wallet selection scope parity — 2026-10-06

| Proof | Result |
| --- | --- |
| Contract decision | PASS: selecting a wallet from management routes to the transaction ledger with that wallet scoped, matching the web route; a separate detail action remains available. |
| Verification | PASS: `testAPIBackedWalletSelectionRoutesToScopedLedger` passed in the clean `LOCAL_API_UI` run; `testAPIBackedWalletDetailExposesTransactionEditor` and the wallet mutation test separately cover the explicit detail surface. |

## Transaction category integrity — 2026-10-06

| Proof | Result |
| --- | --- |
| Backend/API boundary | PASS: ordinary transaction, recurring and credit handlers reject missing categories; repository checks and AI proposal validation preserve the same invariant. |
| Database boundary | PASS: migration `000028_transaction_category_required` backfills ordinary legacy rows with canonical fallback groups and adds a check constraint; adjustment rows remain uncategorized by design. |
| Web/native editor parity | PASS: editors prefill the first applicable category, remove the empty option, reject incompatible categories and send a non-null category ID. |
| iPhone 16 Simulator regression | PASS: fresh iPhone 16 Simulator run completed 66/66 native unit/contract tests and the default UI target completed 6 passed, 11 intentional `LOCAL_API_UI` skips, 0 failures; the red-to-green required-category policy test is included. |

# Validation Matrix

## Native simulator and local integration regression — 2026-10-06

Latest rerun: the first unsigned simulator UI invocation exposed a harness
defect rather than an app/API defect. The local OAuth exchange returned HTTP
200, but Keychain rejected persistence with `errSecMissingEntitlement
(-34018)`. A DEBUG-only in-memory session store is now used for
`-UITestAuthenticatedAPI`; production and normal builds still use Keychain.
After the correction the complete UI suite passed 17/17 on the booted iPhone
16 Simulator. The latest rerun is 65/65 native unit/contract tests and 17/17
signed API-backed UI tests. No token values are written to logs.

After the wallet-scope web fix (`cfc170b7`), the same booted iPhone 16
Simulator was rerun before continuing feedback work: native contracts passed
65/65 and the default UI target passed 17/17 XCTest cases (11 intentionally
skipped because they require `LOCAL_API_UI`, 0 failures). The isolated signed
API-backed command list remains the authoritative local-data UI gate because a
single multi-test `LOCAL_API_UI` invocation can trigger an Xcode 27 runner
restart.

| Proof | Result |
| --- | --- |
| Native simulator frame-warning regression | PASS: the shared UIKit-backed amount field exposes and executes `Xong`; wallet, wallet-detail transaction, transfer and credit mutation UI tests pass 4/4, and the latest iPhone 16 Simulator log contains no `Invalid frame dimension` fault. |
| Production feedback route exposure | PASS: `https://money.dungxbuif.com/api/v1/health` returned 200 and unauthenticated `GET /api/v1/feedback` returned the expected 401 `AUTH_INVALID_FORMAT` problem with a request ID. Authenticated production feedback create/list, screenshot storage and agent lifecycle remain unrun because no production credential was used in this probe. |
| Local authenticated feedback roundtrip | PASS: the running local API fixture exchanged the iOS fixture OAuth payload, created one owner-scoped feedback item, listed it and fetched it by ID; the exact `sim-feedback-*` row was removed from disposable PostgreSQL afterward. |
| Local feedback API-key scope | PASS: a disposable owner key with `feedback:read`/`feedback:write` listed the authenticated owner's feedback with HTTP 200, then was revoked with HTTP 204; the key secret was not logged. |
| iPhone 16 Simulator `EC80B092-CEF1-4A8E-A5A3-7420298EFD4D`: native unit/contract XCTest | PASS: 65 tests, 0 failures; includes the bounded PNG screenshot regression alongside the immutable wallet-row action policy and adjustment-direction display, transaction-editor hũ eligibility, inactive-selection and report/date-time policy coverage, expired access-token refresh, invalid-session clearing, malformed ledger timestamp exclusion, iOS PKCE encoding, safe 204 mutation decoding, immutable bulk-delete boundaries, URL query preservation, structured Advisor data, localized selectors, jar/portfolio/recurring/Travel Mode/AI-entry validation and account-local period/date behavior |
| iPhone 16 Simulator: authenticated route/feedback UI smoke | PASS: the latest signed opt-in `LOCAL_API_UI` suite is 82/82 tests passed (65 unit/contract + 17 API-backed UI) with Assistant, AI-entry capability/review/attachment cap, ledger week/cashflow, wallet, credit-wallet, transfer and feedback fixtures seeded/removed through local API. The unsigned simulator test build uses the DEBUG-only in-memory fixture session store; normal and Release builds retain Keychain-backed sessions. |
| Native AI-entry attachment parity | PASS: a deterministic DEBUG-only fixture renders 20 valid PNG thumbnails, disables the picker at the configured cap, and re-enables it after a close/remove action; the targeted and full signed simulator UI tests pass. |
| Native feedback bubble screenshot context | PASS: the floating trigger is hidden for one rendered frame before capture; the dedicated UI regression confirms only the feedback-sheet action remains visible while the overlay is open, so the attached screenshot cannot include the trigger. |
| Native business mutation smoke | PASS: signed UI created and renamed a basic wallet, created/deleted an atomic transfer pair, and created both a credit purchase and payment while observing refreshed rows. The shared numeric amount field now exposes a “Xong” keyboard action so credit notes remain reachable on iPhone. Post-run disposable counts: wallets=0, transactions=0, feedback=0. |
| iPhone 16 Simulator: authenticated “Thêm” route sweep | PASS: all 15 secondary destinations, including the wallet-management route, opened and returned to the menu; wallet management exposes an explicit back action and selecting a wallet can hand off to the scoped ledger |
| Native category scope parity | PASS: category create/edit preserves multiple selected wallet IDs; empty selection remains the API's all-wallet scope; shared `MPMultiOptionSelector` owns the UI |
| Native ledger/detail/assistant edge cases | PASS: malformed timestamps are excluded from period filters, wallet detail renders the complete API result, assistant polling preserves a still-running run after the local wait budget, Travel Mode uses stable event IDs, and balance adjustments preserve the account-local calendar day |
| Native transaction bulk-delete parity | PASS: selection mode uses base buttons, excludes transfer/credit/adjustment rows, enforces the 100-ID cap, confirms atomically, removes successful rows and keeps a visible error plus selection on failure |
| Native empty mutation responses | PASS: APIClient contract test accepts `204 No Content` for the bulk-delete mutation without envelope decoding failures; the same path covers typed delete methods |
| Native Finance Assistant history/overview parity | PASS: owner-keyed opaque conversation ID, latest 50 message reload, older-message pagination, read-only monthly overview card and structured source data are covered by API/model tests; the API-backed iPhone 16 UI smoke returned 200 from local `:8080` |
| Native credit ledger form parity | PASS: credit entry/payment controls now preserve kind, applicable category, source wallet, amount and optional note through the same base selectors/inputs as the web contract; category wallet/kind filtering is covered by XCTest |
| Native credit wallet detail action parity | PASS: selecting a credit wallet from both the main Giao dịch tab and Ví tab routes to the shared credit statement plus entry/payment actions; the API-backed iPhone 16 smoke passed against the disposable local fixture |
| Native API query-path contract | PASS: APIClient keeps query strings in the URL query instead of escaping `?` into the path; credit statement and shared report/jar/advisor query paths are covered by the URL contract test |
| Native request correlation | PASS: APIClient attaches one UUID `X-Request-ID` to normal and multipart requests and reuses it across the single JWT refresh retry; refresh itself receives a separate request ID |
| Native wallet detail parity | PASS: basic/goal detail now shows account-timezone monthly groups, category/note labels, income/expense totals and shared transaction add/edit; an API-backed iPhone 16 UI smoke opened the seeded fixture wallet and editor, then the fixture was deleted |
| Native wallet detail mutation state | PASS: detail reloads the selected owner-scoped wallet after transaction mutations so balance/goal progress stays current; category metadata remains optional while transaction rows stay visible |
| Native transaction ledger options parity | PASS: the base overflow menu owns period/grouping/transfer/adjustment/bulk actions; rows include category and wallet context with signed amounts; goal/credit selections suppress ordinary period controls and switch to their detail/credit action surfaces; standard UI smoke opens the overflow actions |
| Native transfer wallet/date boundary | PASS: transfer selectors exclude credit wallets and serialize the selected date at account-local noon; source/destination equality and server balance/ownership rules remain enforced by the API |
| Native portfolio parity | PASS: native portfolio now exposes cost/current/unrealized metrics, per-asset current-price drafts that can be cleared, and an explicit trade date; trade drafts enforce buy/sell, positive decimal quantity up to eight places, and non-negative integer VND price/fee contracts before submission |
| Native recurring wallet/category parity | PASS: recurring schedules exclude credit wallets and category options follow the selected expense/income type; changing type clears an incompatible category selection |
| Native selector and jar validation parity | PASS: shared selectors localize API enum values without changing payloads; jar mode changes clear stale drafts; fixed allocations enforce the backend safe-integer ceiling and percentages require finite values with at most two decimal places |
| Native ordinary transaction/credit boundary parity | PASS: ordinary and AI proposal editors exclude credit wallets, invalid categories clear when wallet/type changes, and aggregate credit rows do not expose ordinary edit/delete actions |
| Native budget/adjustment wallet scope parity | PASS: budget categories are expense-only and both budget and adjustment wallet selectors exclude credit wallets |
| Native ended-budget mutation parity | PASS: ended budgets do not expose the edit/save action, while their history and delete path remain available |
| Native recurring draft validation parity | PASS: shared validator requires positive amount/interval, accepts only valid integer/thousands-grouped input (never silently filters malformed characters), supported type/frequency, trimmed name and non-reversed optional end date before create/update; server ownership/category checks remain authoritative |
| Native portfolio account-date parity | PASS: portfolio loads the account timezone and serializes selected trade dates at account-local noon, preventing UTC/device-day drift |
| Native wallet month/Travel Mode validation parity | PASS: wallet detail renders `Tháng MM năm YYYY`; Travel Mode trims required names, validates optional `yyyy-MM-dd` dates and rejects reversed ranges before API submission |
| Native base-component and project guards | PASS: `check-base-components.test.sh` now rejects raw feature `Button`, `TextField`, `SecureField`, `DatePicker`, `ProgressView`, `Text` and `Image` declarations; `verify-project.sh` also passes. The full signed UI suite remains 17/17 after the base-icon accessibility correction. |
| Native static analysis | PASS: `xcodebuild analyze` for the Release iPhone Simulator target completed with `SWIFT_TREAT_WARNINGS_AS_ERRORS=YES`, exit code 0, and no app compiler/analyzer diagnostics. The Apple metadata extraction notice is tool-level output and does not fail analysis. |
| Native Release runtime smoke | HISTORICAL: `com.mypocket.ios` launched on the booted iPhone 16 Simulator and showed the login gate with Google action before the later password-only migration; the current auth launch proof is in the fresh checkpoint above. |
| Native authenticated runtime smoke | PASS: a Debug `LOCAL_API_UI` build was installed and launched with `-UITestAuthenticatedAPI`; the real local API exchange rendered the authenticated `Tổng quan` screen, empty-wallet/empty-ledger states and floating feedback bubble. The follow-up process-log scan found no fatal, crash, unhandled or invalid-frame diagnostic. |
| Native Travel Mode parity | PASS: header explains context/status, event cards expose active state and date scope, deletion uses shared confirmation, and the editor warns that recurring/transfer/adjustment transactions do not auto-link; `MyPocketIOSTests` 54/54, Release warnings-as-errors, base/project guards, and authenticated More-route 1/1 in 122.031s all pass after the change |
| Native recurring parity | PASS: summary explains account-timezone scheduling, cards show type/frequency/wallet/amount/next run/note/active state, delete is confirmed, and `MPFormat.shortDateTime` preserves account-local presentation; `MyPocketIOSTests` 55/55, Release warnings-as-errors, guards, and authenticated More-route 1/1 in 124.207s pass |
| Native reports parity | PASS: report loads optional month-summary data beside Insider metrics, renders category/type/count rows and ledger activity exclusions, and preserves Insider output if month-summary fails; `MyPocketIOSTests` 55/55, Release warnings-as-errors, guards, and authenticated More-route 1/1 in 124.287s pass |
| Native credit statement parity | PASS: credit detail presents positive current debt, localized payment status/kind labels, optional due-date and note-first row context through shared base components; `MyPocketIOSTests` 55/55, warnings-as-errors simulator build, and authenticated More-route 1/1 in 125.111s pass |
| Native portfolio detail parity | PASS: asset cards now show weighted average, realized/unrealized P/L, current value or explicit unknown pricing, and account-timezone buy/sell history with quantity/price/fee/note; `MyPocketIOSTests` 56/56, Release warnings-as-errors, guards, and authenticated More-route 1/1 in 121.948s pass |
| Native enum-label parity | PASS: shared labels hide wallet/category/credit/feedback/status wire enums on user-facing iOS surfaces; `MyPocketIOSTests` 59/59, Release warnings-as-errors and guards pass; signed credit-wallet and feedback/more UI checks pass after the change |
| Native AI proposal review parity | PASS: unknown/transfer output renders `Chưa xác định — chọn loại` and cannot save until the user selects income/expense; category/type/wallet scope clears incompatible values, only active account-month hũ options are offered to eligible expense wallets, and report inclusion is editable. Pure `AIEntryProposalReview` tests pass; the signed iPhone 16 deterministic review fixture and full `LOCAL_API_UI` suite pass 1/1 and 15/15. Provider-backed extraction remains an external gate. |
| Native AI proposal field parity | PASS: proposal edit preserves account-local date/time, keeps an inactive existing hũ visible for explicit resolution, rejects stale hũ values on save, and excludes the transfer-out category from hũ choices while retaining report inclusion in the PATCH. Deterministic proposal UI and policy regression pass. |
| Native ordinary transaction editor parity | PASS: manual expense editor shows active account-month hũ options for eligible wallets, preserves inactive existing hũ selections for explicit resolution, sends `included_in_reports`, and serializes account-local date/time; incompatible hũ values clear on type/wallet/category changes. Policy tests and the API-backed editor UI assertion pass; full signed `LOCAL_API_UI` is 15/15. |
| Native Release build | PASS: iPhone 16 simulator Release build with `SWIFT_TREAT_WARNINGS_AS_ERRORS=YES`; app installed and launched as `com.mypocket.ios` after the build |
| Simulator API-backed authenticated smoke fixture | PASS: DEBUG-only `-UITestAuthenticatedAPI` exchanged fixture PKCE fields against local `:8080`, stored the disposable JWT in the in-memory UI harness session store, and rendered the authenticated overview with live empty wallets/ledger; screenshot `/tmp/mypocket-ios-simulator-api-auth.png` |
| Go backend disposable PostgreSQL + Redis race suite | PASS: final rerun `go test -race ./... -count=1` completed successfully across API, worker, controller, entity, AI, cache, repository, storage and usecase packages; no data race or package failure |
| Go formatting/static quality | PASS: `gofmt -l backend` is clean; `go vet ./...` and Swift Debug build with warnings-as-errors pass |
| FE-proxy transfer integration | PASS: paired creation, balance reconciliation, report/jar exclusion, transfer PATCH and pair DELETE with idempotency keys, auth guard |
| Real Chromium 390×844 transfer flow | PASS: actual UI rendered both wallet options, saved, refreshed and matched persisted API state; stale fixture cleanup now deletes transfer pairs atomically |
| AI entry browser flow | PASS: contract + Chromium fixture; the fixture mounts the route-bound transactions panel under a memory router, derives transaction dates from the current account-local week, and verifies shared form/card editing, save-all, transfer blocking and grouped ledger rendering |
| Feedback JWT + scoped user API key + agent lifecycle | PASS: create/list/revoke, agent open → triaged → in_progress → changelog/fixed, owner-scoped reads, Redis `mypocket:audit:feedback` events observed; native screenshot regressions assert PNG bytes, `screen.png`/`image/png` multipart metadata and the 1 MiB bound |
| Web design/typecheck/build/API roundtrips | PASS: final local design guard, base contracts, wallet-scope contract, TypeScript check and Vite build passed (1,791 modules); Vite only reports the existing large-chunk warning |
| Provider-backed AI/OCR | PASS: configured DeepSeek-V4-Flash text/tool/schema probes and a fresh S3 → OCR Base64 → LLM receipt pipeline pass; 5/5 cases, 24/24 semantic fields, 100.0%, no ledger writes |
| Real Google iOS OAuth, authenticated staging/production and physical device | NOT RUN; these external release gates remain explicit because no iOS client credential, production auth credential or online physical device was supplied |

The disposable database was `mypocket_ios_local`; generated feedback/changelog
rows and transfer fixtures were removed after the probes. No production data or
secrets were used.

## UI-LEDGER-WEEK — 2026-09-24

- Latest owner correction: basic uses the same Week/Custom, period cashflow and date-grouped list as aggregate, with both totals and rows wallet-scoped. `test:transactions` RED first caught the old basic branch including an August row in a September week; GREEN 7/7 after the range correction. `test:calendar` 11/11, period/goal SSR scripts, `test:design` 7/7 plus base contracts, `check:design`, TypeScript typecheck and Vite build all PASS. Live Chrome local page confirmed selecting `Tiền mặt` shows the shared period strip/cashflow/list, previous week 14–20/09 shows zero totals and no rows, and returning to 21–27/09 restores in-week rows. Goal browser selection/mobile visual UAT remain NOT RUN because that user account has no goal wallet; credit specialized accounting remains out of scope.

- [Design and scope](tickets/UI-LEDGER-WEEK-DETAIL_DESIGN.md), [transaction page](../design/pages/transactions/README.md), [user/agent guide](../guides/transactions.md).
- RED→GREEN: week/calendar and combined wallet/date tests failed before implementation; `npm run test:calendar` 11/11 and `npm run test:transactions` 6/6 now pass. SSR period selector confirms tab labels, inclusive range and disabled future-week control.
- `npm run test:design` 7/7 plus base contracts PASS; `npm run build` PASS (1,776 modules, existing bundle-size warning); `git diff --check` PASS.
- Real Chrome local preview: All → current week 21–27/9 (spending 30,000 đ) → previous week 14–20/9 (empty, 0 đ) → current week; next-week disabled, custom-date cancel preserved week, then All restored. No data writes. Owner mobile visual sign-off and deployment remain open.

## AI entry review-first regression — 2026-09-23

[Design, reproduction, docs review and limits](bugs/AI-ENTRY-REVIEW-FIRST-DETAIL_DESIGN.md).

| Proof | Result |
| --- | --- |
| Configured provider, old prompt | RED: synthetic overlap lost cropped-date row; nine-image stored OCR replay returned 0 drafts |
| Held-out overlap fixture, corrected prompt | PASS: 3 drafts, repeated reference merged, distinct dates preserved, missing date blank with question, sole wallet prefilled, balance excluded |
| Nine-image stored OCR replay, output cap 32768 | PASS: 30 owned-wallet drafts, 40.682s model call; no proposal/ledger writes; not a ground-truth completeness assertion |
| Full Go suite; focused AI/usecase race tests; backend build | PASS; env-gated integration/live tests not implied by default suite |
| Truncated provider result; complete 30-draft response | PASS: reject partial output without retry; accept completed batch |
| Restarted local backend and frontend HTTP checks | PASS: /api/v1/health and :4173 return 200 |
| Fresh browser upload/save, full financial reconciliation, production deploy | NOT RUN; owner UAT pending |

### Unlimited proposal-count correction — 2026-09-23

| Proof | Result |
| --- | --- |
| Strict parser accepts 31 drafts | PASS: no application count rejection |
| Usecase/repository proposal count guards | PASS: `>30` guards removed; each valid draft is persisted for review |
| Provider safety envelope | PASS: 32768 output tokens and 256 KiB response remain transport protections; truncated JSON is rejected atomically |
| Full backend tests and race-focused AI/usecase tests | PASS |

### AI provider timeout regression — 2026-09-23

| Proof | Result |
| --- | --- |
| Failed real request timing | RED: AI process returned 503 after ~91s while `/api/v1/health` remained 200; failure was model timeout, not service/database outage |
| Timeout budget correction | PASS: model HTTP/client budget 180s; enclosing extraction/usecase budget 210s; no automatic retry |
| Focused AI/usecase tests and backend build | PASS |
| Restarted local API health check | PASS: HTTP 200 |

### AI diagnostics, model usage and instruction scope — 2026-09-23

| Proof | Result |
| --- | --- |
| AI extraction logs | PASS: start/OCR/model/schema/success/failure stages log counts, sizes, latency, status and token counts only |
| User instruction contract | PASS: provider fixture sees separate `user_instruction` for month filtering; OCR/source remains untrusted context |
| Usage retention | PASS: provider usage is normalized and stored only when `AI_STORE_USAGE=true`; public process response excludes it |
| Owner isolation audit | PASS for user-facing AI entry/advisor queries: owner predicates or owner-verified joins cover sessions, proposals, attachments, runs, messages, events and finance tools. Attachment cleanup is an internal global sweeper by design and exposes no data. |
| Streaming timeout decision | VERIFIED DESIGN: partial streamed JSON is unsafe for proposal persistence; async job + existing request polling/SSE is the follow-up transport, not enabled in this correction |

## Field Ownership

- Human owns proof overrides and acceptance sign-off.
- AI recommends proof types, links evidence, and updates status from verification.

This file maps accepted behavior and work items to proof.

## Full implemented-scope regression audit — 2026-09-22

- Backend: `TEST_DATABASE_URL=<local dev PostgreSQL> go test -race ./... -count=1` PASS — 110 tests in 15 packages; migration CLI reports `version: 16 dirty: false`.
- Frontend: `npm run check:design`, `npm run test:design` (7), `node --test scripts/form-logic.test.ts` (2), `npm run test:transactions` (5), `npm run test:calendar` (9), `npm run test:transaction-jars`, `npm run test:ai`, `npm run test:api:full`, `npm run test:transfer:integrated`, `npm run test:transfer:e2e`, and `npm run build` PASS. Production build transformed 1,764 Vite modules. AI/transfer browser suites were run sequentially; a deliberately parallel attempt produced a fixture timeout from shared Vite/Chrome ports and passed on the clean serial rerun.
- API harness: `TEST_API_ORIGIN=http://127.0.0.1:4173 TEST_DATABASE_URL=<local dev PostgreSQL> npm run test:api:full` PASS through the real FE proxy. It verifies authenticated profile/home, timezone validation, category CRUD and wallet scope, jar create/list/update/report/remove, month summary and note save/delete, AI capabilities, Swagger, and unauthenticated guards. It removes only its generated category, jar/configuration, and note; no ledger rows are created. The seeded account timezone value is unchanged (the valid same-timezone call may confirm an unconfirmed account, as the endpoint contract requires).
- Transfer harnesses: integrated FE-proxy/PostgreSQL roundtrip and real Chromium 390×844 flow both PASS; paired rows, balances, report/jar exclusion, auth guard, rendered wallet options, submit, refresh and persisted state were asserted. Only generated wallets/transactions are removed.
- Fixture correction: `app/scripts/api-roundtrip.mjs` now uses the documented budget contract `start_date`/`end_date` (the stale `start_at`/`end_at` fixture was the only initial red).
- Residual human/platform proof remains explicit: owner visual sign-off, mobile/PWA install, Google OAuth with real credentials, live S3/OCR/LLM, and deployment verification are outside this local regression run. AI browser/provider tests intentionally do not claim live OCR/LLM quality or proposal approval.

## TICKET-02-02 atomic wallet transfer — 2026-09-22

- Scope and trace: [detail design](tickets/TICKET-02-02-TRANSFER-DETAIL_DESIGN.md), [transaction page](../design/pages/transactions/README.md), [API contract](../architecture/API.md), [ERD](../architecture/ERD.md).
- Schema proof: migration `000016_transaction_transfer_links` applied locally; `go run ./cmd/migrate version` reports `version: 16 dirty: false`.
- Backend proof: `TEST_DATABASE_URL=<local dev database> go test ./... -count=1` PASS. Handler tests cover valid paired creation and same-wallet rejection. PostgreSQL integration proves exactly two linked rows, explicit `included_in_reports=false`, owner/credit guards and atomic persistence.
- Frontend proof: `npm run check:design` PASS; `npm run test:design` PASS (7 guardrails plus base contracts); `npm run build` PASS (1,764 Vite modules). TransferFields composes `BaseSelect`, `AmountField`, `DateField`, `BaseTextInput` and `SurfaceCard`; no native controls or literal colors were added outside bases.
- Integrated proof: `npm run test:transfer:integrated` PASS through Vite proxy → Gin → PostgreSQL; verifies persisted pair, balances, report/jar exclusion and unauthenticated 401, then cleans only test IDs. `npm run test:transfer:e2e` PASS in real Chromium at 390×844; opens the actual transaction page, selects both rendered wallet options, submits the real sheet, waits for the refreshed UI note, re-reads API state and saves `/tmp/mypocket-transfer-e2e.png`.
- Residual UAT: owner visual sign-off remains human-owned. Pair edit/delete and idempotency-key replay are intentionally follow-up scope.

## CORE-03 timezone, jars, and live monthly summary — 2026-09-22

- Scope and trace: [CORE-03 design](tickets/CORE-03-TIME-JARS-MONTH-DETAIL_DESIGN.md), [execution plan](../superpowers/plans/2026-09-22-timezone-jars-month.md), [ADR-008](../decisions/ADR-008-account-timezone-and-calendar-dates.md), [account settings](../design/pages/current-ui.md), [jar page](../design/pages/jars/README.md), [month detail](../design/pages/month-detail/README.md), [overview](../design/pages/overview/README.md), [transaction form](../design/pages/transactions/README.md).
- Persistence/API proof: additive migrations `000013`–`000015` are applied locally; `go run ./cmd/migrate version` reports version 15, `dirty=false`. PostgreSQL repository tests cover month-note owner scope/clear, stable jar month initialization under concurrent calls and owner isolation, empty summary collections, and account-local calendar behavior. Backend full suite `TEST_DATABASE_URL=<local dev database> go test ./... -count=1` PASS; database URL intentionally omitted from this record.
- Frontend proof: `npm run check:design` PASS (105 local design links and base/color/native-control checks); `npm run test:design` PASS (7 design guardrails plus base contracts); `npm run test:transactions` PASS (5); `npm run test:ai` PASS; `npm run test:calendar` PASS (9); `npm run test:transaction-jars` PASS; `npm run build` PASS (1,763 Vite modules). The Vite-backed AI and jar tests were run sequentially to avoid shared dev-server port contention.
- Integration/runtime proof: backend health and frontend root returned HTTP 200 on the already-running local servers. No owner ledger was changed. Migration backfill, live transaction CRUD with jars, month note editing, timezone switching and visual acceptance still require owner UAT; do not mark the feature complete on automated tests alone.
- Reporting semantics: month figures remain recalculable from live ledger; completion is derived by account-local month, with no close snapshot or cron job. The separately requested worker is draft [WORKER-01](tickets/WORKER-01-cron-service.md); its requested month-end report/close behavior must be reconciled with this contract before detail design.

## UI-FORMS-03 reference follow-up — 2026-09-21

- Scope and trace: [UI-FORMS-03](tickets/UI-FORMS-03-DETAIL_DESIGN.md), [transaction page](../design/pages/transactions/README.md), [budget page](../design/pages/budgets/README.md), [assistant page](../design/pages/assistant/README.md), [BaseCategoryTree](../design/molecules/category-tree/README.md).
- Owner decisions: Money Lover guides the AI composer/result presentation; MyPocket keeps per-item edit/remove/save and save-all confirmation. Group pickers use the same BaseCategoryTree as Account → Manage Groups; no screen-local category list.
- Frontend fixture proof: `rtk proxy npm run test:ai` in `app/` PASS after checking AI textarea/attachment/result card, transaction and budget tree pickers, parent/child selection, and transaction category ID payload. `rtk proxy node scripts/base-contract.test.mjs` PASS for tree selection and selected mark. Earlier first red was a test fixture omission (no `onSelect` callback), then isolated rerun passed; no code fix was needed for that assertion.
- `rtk proxy npm run check:design` in `app/`: PASS (theme/native-control guard and 96 local links); `rtk proxy npm run test:design`: PASS (7 guardrail fixtures + base contracts); `rtk proxy npm run test:transactions`: PASS (5 cases); `rtk proxy node --test scripts/form-logic.test.ts`: PASS (2 cases); `rtk proxy npm run build`: PASS (TypeScript and 1,756 Vite modules); `rtk proxy git diff --check`: PASS.
- Visual proof: browser fixture at 390×844 compared Money Lover composer/result screenshots to `/tmp/mypocket-ui-pIE4DC/ai-entry-final.png`; layout/structure and MyPocket theme pass. AI browser fixture passed after the final UI adjustment. Owner visual UAT remains pending. AI-ENTRY-02 still owns replacing current session/message persistence with one-shot batch APIs and private receipt linkage.

## AI-ENTRY-02 stateless/OCR-first follow-up — 2026-09-22

- RED→GREEN: backend OCR adapter initially rejected `application/pdf`; a regression proved zero OCR/model calls. PDF validation and OCR Platform's private presigned-upload flow now pass. Captured LLM request tests prove OCR text is present and PDF bytes/base64/source URL/image input are absent. Invalid file validation now returns a client error before request/provider claim.
- Backend route contract test proves `POST /api/v1/ai/entry/process` and read-only `GET /api/v1/ai/entry/requests/:request_id` exist and no public `/sessions` or `/messages` routes remain. Process response uses `process_id`; repository writes no conversation messages and LLM extraction type no longer carries history.
- Frontend AI service test proves a raw PDF `File` is sent as multipart without a base64 field or manually set Content-Type. Browser fixture passed real Chrome one-shot submit and asserts no session/message calls. `npm run typecheck`, AI service/browser tests, `npm run check:design`, `npm run test:design`, and `npm run build` pass.
- `go test ./... -count=1` in `backend/`: PASS; rerun with the local dev `TEST_DATABASE_URL`: PASS, including real PostgreSQL approval/link/download authorization, rejection of unlinked downloads, expired-unlinked cleanup claims and stable OCR-text-to-file mapping. Migration CLI: `go run ./cmd/migrate up && go run ./cmd/migrate version` → version 12, `dirty=false`. Fake-store/use-case proof confirms original bytes are retained and only the signed URL reaches OCR; OCR adapter proof confirms only recognized text reaches the LLM and storage failure stops before extraction. The explicit cleanup command exists but was not run against bucket contents. Live S3/OCR and browser download UAT remain pending. Do not auto-retry the user's prior PDF.
- Startup config: `go test ./cmd/api -run TestAttachmentStorageIsOptionalOnlyInDevelopment -count=1` PASS; empty storage is allowed only in development and complete config initializes the adapter. S3 boundary regressions reject control characters/traversal in environment, prefix, bucket, region and key IDs.
- Local runtime restarted after implementation: `GET http://localhost:4173/` → 200 and `GET http://localhost:8080/api/v1/health` → 200. No live OCR/S3 request was made.

## AI-ENTRY-02 live provider evaluation — 2026-09-22

- Scope: synthetic receipt image only; no user PDF, no proposal approval, and no ledger write. Harness: `backend/internal/infrastructure/ai/live_eval_test.go`.
- Command: `rtk proxy env AI_EVAL_LIVE=1 AI_EVAL_IMAGE=/tmp/mypocket-ai-eval-receipt.png go test ./internal/infrastructure/ai -run '^TestLiveQwenImageAndTransactionEval$' -count=1 -v`.
- Regression history: before the fix, five live runs reproduced S3 PUT success + signed GET `403 SignatureDoesNotMatch`; alternate `us-east-1` signing also failed.
- Fix/evidence: [ADR-007](../decisions/ADR-007-aws-s3-presigning.md) replaces handwritten signed GET with AWS SDK for Go v2 S3 presigning. `rtk proxy go test ./internal/infrastructure/storage -count=1` and `rtk proxy go test ./... -count=1` in `backend/`: PASS. Live private PUT+GET readback PASS (1,027,882 bytes, 62 ms); the test deferred object cleanup.
- OCR: PASS against the configured live provider, 3.815 s; synthetic mock total of 45,000 VND recognized.
- LLM: all five cases failed strict draft-schema validation; field score 0/15 (0%). LLM latency p50 15.089 s, p95 21.7 s. Transfer-safety gate failed; OCR + receipt-case pipeline measured 25.516 s. AI-ENTRY-01 schema compatibility is not verified and is not a successful intelligence score.
- Safe diagnostic log follow-up: `rtk proxy go test ./internal/infrastructure/ai -run '^(TestSchemaMismatchLogReportsStructureWithoutLoggingModelContent|TestSchemaDiagnosticShowsArrayElementShapeWithoutValues|TestSchemaValidation)$' -count=1 -v` PASS; full `rtk proxy go test ./... -count=1` in `backend/` PASS; `rtk proxy git diff --check` PASS. A live run showed root arrays with wallet-like item fields (`currency,id,name,type`) and objects with input-context keys (`categories,now,timezone,untrusted_source_text,wallets`); one of five calls timed out. Five-case metrics: LLM p50 21.138 s, p95 30.001 s; OCR 3.805 s; OCR + receipt pipeline 33.806 s; 0/15 fields. Latest text-only live call logged `finish_reason=stop`, root array/item shape with input-context keys, 363 response bytes, and 14.706 s latency. Logs retain only allowlisted schema/catalog keys, redact arbitrary key names and exclude values/full response bodies. See [AI-ENTRY-01](tickets/TICKET-09-01-ENTRY-DETAIL_DESIGN.md).
- Cleanup regression: `TestAttachmentDeleteStatusRequiresClaimedDeletingState` red→green; stale or unclaimed attachment cleanup finalization now returns conflict instead of silently reporting success.
- Synthetic fixture only; no user PDF, proposal approval, or ledger write. See [BUG-001](bugs/BUG-001-s3-presigned-get-signature.md).

## AI-ENTRY-01 model timeout regression — 2026-09-23

- Root cause: MyPocket omitted `max_tokens` from the OpenAI-compatible request. oMLX inherited its local 1,000,000-token default; Qwen generation was cancelled at the 30-second model-client timeout and logged `stage=model`, `code=model_request`, `provider_status=0`.
- Fix history: the first production correction sent `max_tokens=2048`, `temperature=0`, and `chat_template_kwargs.enable_thinking=false`; the later multi-image correction raised the current extraction envelope to `max_tokens=32768`. The provider boundary test asserts the current cap; `go test ./... -count=1` passes.
- Live verification after restart: text-only extraction returned HTTP 200 in 10.73 seconds; a real 1,027,882-byte receipt image returned HTTP 200 in 13.97 seconds with one review proposal. Capabilities reported `ai_configured=true`, `ocr_configured=true`, and `files_configured=true`.
- Negative probe: an intentionally blank 1×1 PNG returned the expected generic 503 with safe log diagnostic `stage=ocr_poll`, `code=ocr_failed`; this is OCR content rejection, not the model timeout regression. No ledger transaction was auto-written by either probe.

## AI-ENTRY-01 OpenAI-compatible JSON Schema resolution — 2026-09-22

- Comparison: prior `response_format: json_object` production path scored 0/15 and returned input-context-shaped roots. A forced tool/function-calling probe reached response streaming but did not complete within the 30 s HTTP-client timeout; tool calling is not used because extraction must never execute a tool. Strict `response_format: {type: json_schema, json_schema: {strict: true, ...}}` is supported by the configured OpenAI-compatible endpoint/model.
- Implementation: production adapter now uses strict JSON Schema and retains the strict application parser, human review, and no-ledger-before-approval boundary. No SDK/dependency was needed: standard OpenAI-compatible `net/http` transport successfully exercised the provider contract. Optional owner wallet descriptions are included as bounded (2 KiB each), explicitly untrusted disambiguation context; no DB/API change.
- Live command: `rtk proxy env AI_EVAL_LIVE=1 AI_EVAL_IMAGE=/tmp/mypocket-ai-eval-receipt.png go test ./internal/infrastructure/ai -run '^TestLiveQwenImageAndTransactionEval$' -count=1 -v` in `backend/`.
- Live result: private S3 PUT/readback PASS (1,027,882 bytes, 83 ms); OCR PASS (4.034 s, synthetic total recognized); LLM synthetic evaluation PASS 5/5 cases, 24/24 scored fields (100%), transfer-safety PASS. LLM p50 19.769 s, p95 25.655 s; OCR + receipt-case pipeline 29.689 s. Cases: single expense, two expenses/relative date, income/large VND, internal transfer, OCR receipt.
- Tests: `rtk proxy go test -race ./internal/infrastructure/ai -count=1` PASS; `rtk proxy go test ./... -count=1` PASS in `backend/`; `rtk proxy git diff --check` PASS. Synthetic-only; no proposal approval or ledger write. A five-case benchmark is not broad model certification; owner review and broader edge-case UAT remain required.
- Wallet-description and future provider/pricing intent: [AI-ENTRY-01](tickets/TICKET-09-01-ENTRY-DETAIL_DESIGN.md), [AI-ENTRY-03](tickets/AI-ENTRY-03-PROVIDER-SELECTION.md) and its draft design.

### AI-ENTRY-02 current provider probe — 2026-10-06

- Synthetic receipt generation and private S3 PUT/signed-readback: PASS
  (`731,920` bytes); the object was scheduled for cleanup by the test.
- OCR submission: PASS against the configured remote service with
  `mode=wait`, `waitTimeoutMs=15000`; the synthetic receipt total was
  recognized in 1.496 s.
- LLM extraction: PASS, 5/5 cases and 24/24 semantic fields (100.0%); the
  receipt is classified as an expense, wallet prefilled, and transfer safety
  retains a review question. Provider model: `DeepSeek-V4-Flash`.
- End-to-end synthetic S3 → OCR → LLM pipeline: PASS in 2.656 s for the
  receipt case. No proposal, transaction, or production row was written.
- The harness deliberately sends validated image bytes as Base64 to OCR. The
  private S3 signed-readback is a storage check only; provider-owned
  `s3://` URLs remain reserved for the PDF private-presign contract. This
  avoids sending an app-owned signed URL that the OCR service cannot
  dereference.

### AI-ENTRY-01 current provider compatibility regression — 2026-10-06

- Read-only live text extraction: PASS with the configured `DeepSeek-V4-Flash`
  OpenAI-compatible endpoint; one synthetic expense returned HTTP 200-equivalent
  extraction in 1.531 s with one valid draft and 1,726 reported tokens.
- Tool-call capability probe: PASS; exactly one `propose_transactions` call was
  returned, arguments passed the strict draft schema, and no tool was executed.
- Strict JSON Schema probe: PASS, 5/5 synthetic cases and all semantic checks;
  the unclear-wallet case now intentionally prefills the first supplied wallet
  (`wallet-cash`) while retaining a review question, matching the approved
  review-first/default-wallet contract. The transfer case preserves one
  explicitly stated amount as one transfer draft (no model splitting). No
  ledger or proposal writes occurred.
- Regression fixture correction: the previous expectation that an unclear
  wallet must remain empty contradicted the shipped default-wallet policy; the
  live test now asserts default prefill plus review question.

## AI plan review — 2026-09-21

### AI-ENTRY-01 implementation follow-up

- [Approved slice and execution proof](tickets/TICKET-09-01-ENTRY-DETAIL_DESIGN.md), [UI contract/proof](../design/pages/assistant/README.md), [ADR-005](../decisions/ADR-005-ai-entry-review.md). Entry scope is implemented for review; this does not complete all REQ-17, financial Q&A, transfer matching or retained receipts.
- Backend red→green tests: missing validator/repository/service/config functionality; actual regressions caught hidden category-template acceptance and too-large history blocking follow-up. PostgreSQL fixtures prove zero writes before approval, eight concurrent approvals producing one row, stale version conflict, cross-owner rejection, terminal reject, deleted-ledger replay safety, explicit report=false preservation and the current no-user-usage-limit policy for AI entry. HTTP fixture test exercises real provider adapter → service → DB → edit/approve/reject and reload.
- Independent review remediation: shared category wallet scopes no longer disclose/overwrite other owners' assignments; OCR source survives model failure; post-claim history includes the previous completion. All three regressions observed red then green; new fixture category/wallets removed after verification.
- Final post-review run: full `go test -race ./... -count=1` with local `TEST_DATABASE_URL` passed; frontend `test:ai`, `check:design`, `test:design`, `test:transactions`, `build` passed. Local documentation scan 184 links/22 Markdown files: pass. Git-visible files contain none of the supplied credentials; ignored local env mode is 0600. Runtime restarted with final source.
- `rtk go run ./cmd/migrate up` pass, schema 000010; `rtk go generate ./cmd/api` pass. Go suite/race with `TEST_DATABASE_URL` runs real DB tests; fixture records are cleaned, owner data untouched.
- Frontend `test:ai`, `check:design`, `test:design`, `test:transactions`, `build` pass; Chrome test fixtures prove hold/cancel/release suppression and review flow. Live Chrome through localhost:4173 verifies normal Add/manual form, AI entry alternative, configured OCR/unconfigured AI display. No artificial proposals were inserted into the owner's account.
- Live proxy auth guard returns 401, Swagger 200 with eight entry paths. OCR credentials are configured but public capability/auth-format probes do not prove actual OCR extraction. The Qwen extraction test previously failed strict schema validation; live AI/OCR quality, physical touch and owner acceptance remain pending. Private S3 configuration passed put/presign smoke checks, while download roundtrip and application attachment-link implementation remain pending.

- [DESIGN-09-AI](tickets/TICKET-09-DETAIL_DESIGN.md): docs-only feasibility/design/plan delivered for review. Existing two AI tickets remain draft; product approval, provider quality, ledger integration and UAT are pending.
- Read-only evidence: transaction entity/handler/repository confirm current income/expense-only ledger and missing AI confirmation path. OCR public OpenAPI and capabilities returned HTTP 200; current contract does not list scan endpoints. Authenticated OCR/model calls were not run.
- Docs validation: `rtk proxy node --input-type=module -e '…'` checked 117 local Markdown links across the 10 touched documents: pass, zero missing targets. `rtk git diff --check`: pass. Placeholder scan in the new plan found no TODO/TBD. App/Go tests skipped because no runtime code changed. Planned unit, PostgreSQL concurrency, provider contract, model eval and browser UAT gates are in T0–T7; none are claimed passed by this planning work.

## Runtime Verification — 2026-09-13

### Follow-up 2026-09-20

- Budget review follow-up: real-Postgres disappearing-row update regression red (nil error/resurrection) → green (not found/no insert fallback); metadata-date preservation regression passes with browser timezone UTC. `go test ./...` with TEST_DATABASE_URL executes optional integration test. Production API restarted and full `api-roundtrip.mjs` rerun passes after fixes. `npm run build` passes; `git diff --check` clean. Independent review evidence in [API-SCREENS-01](tickets/API-SCREENS-01-DETAIL_DESIGN.md).

- API-SCREENS-01: `rtk proxy go run ./cmd/migrate up` pass (000009); `rtk proxy node scripts/api-roundtrip.mjs` in app pass through localhost:4173 to PostgreSQL: goal date/deposit/withdrawal/interest/edit/delete/current_balance, invalid goal catalog rejection; budget CRUD/exact overlap409/child category/report exclusion/recalculation and unauthenticated401. Only test-created IDs cleaned, no owner data removed. `go test ./...` pass; design checks/SSR/build pass. New tests fail before implementation for goal filter, category history, budget derivation/validation, masked amounts and problem+json parsing. Browser Google re-login via FE proxy, budget empty/editor data and grouped transaction form inspected; drafts canceled. Owner acceptance and full browser persistence UAT pending. [Evidence and limitations](tickets/API-SCREENS-01-DETAIL_DESIGN.md).

- Savings slice: target-date red/green regression, update omit/clear checks, `go test ./...`, Swagger generation pass. Frontend design, shared SSR (goal progress/selection/form), ledger tests and build pass. Browser confirmed grouped selector, goal form and Save/cancel states without saving test data. Full goal create/reload/transaction UAT pending. Transfer/report/notifications are not verified features. [Work evidence](tickets/TICKET-06-01-DETAIL_DESIGN.md).

- Wallet screen spec normalized with source-backed Money Lover references and owner select/immutable-type decisions. Docs review complete; UAT not required for this docs-only update. Runtime select change and per-type detail UI are pending; earlier tests do not cover them.

- UI-WALLET-02: check:design, test:design (7 guard regressions plus SSR form/switch/type contracts), test:transactions (4), build all pass. Add Wallet uses shared sheet/card/switch/type picker; normalized both supplied references. Browser visual and new-form API UAT pending. [Evidence](tickets/UI-WALLET-02-DETAIL_DESIGN.md).

- Add Wallet note removal: `rtk proxy npm run check:design`, `rtk proxy npm run test:design`, `rtk proxy npm run build` passed in `app/`. SSR confirms creation omits “Ghi chú” and editing retains it. Owner visual UAT pending. Docs review: wallet design, context, backlog and changelog reconciled; API/schema unchanged, no ADR required.

- Wallet empty-state follow-up: Overview and WalletManagementPanel now use the tested `StatusMessage` plain variant as requested; design checks/tests/build rerun, visual UAT pending.

- Owner reports previously tested bugs are OK; this is user-reported acceptance of tested fixes, not exhaustive product verification.
- [UI-EMPTY-01](tickets/UI-EMPTY-01.md): `npm run check:design`, `npm run test:design`, and `npm run build` pass in `app/`. SSR regression checks cover plain status without card decoration and preserved danger feedback. New empty-state visual UAT pending owner review.
- Local Google OAuth start endpoint verified HTTP 302 with `redirect_uri=http://localhost:4173/api/v1/auth/google/callback` via FE proxy; session-only launch configuration, no tracked credentials.

### UI-BASE-01 normalization

- Status: implemented / verified for shared UI foundation and standards. [Exact evidence and docs review](tickets/UI-BASE-01-VERIFICATION.md).
- `rtk proxy npm run test:design` in `app/`: pass (6 guardrail regression fixtures and shared SSR contracts).
- `rtk proxy npm run build` in `app/`: pass (source guard, docs guard, typecheck, Vite). Docs guard: 17 Markdown specs, 57 valid links, no PNG/HTML exports.
- Chrome fixture: theme primary color/min touch height, selected keyboard tab, disabled/loading CTA, dialog open/focus trap/Escape/return-focus and visual card/tree/gauge checks passed.
- Root cause reproduced before migration: unimported theme, duplicate literal token values, native controls outside atoms. Owner authorized specs + guardrail + refactor; this supersedes prior UI-BASE-01 removal proposal.
- Product-level CRUD/OAuth/ledger acceptance is unchanged; fixture is shared UI proof, not a new business UAT sign-off.

| Area | Proof | Result | Evidence |
| --- | --- | --- | --- |
| Base component styling | TypeScript + production build and authenticated browser inspection | pass | `npm run typecheck && npm run build` in `app/`; canonical `BaseCategoryTree` now powers both Group management and the Reports adapter. Browser proof at `/account/groups`: `nested` source geometry with global colorful icon catalog, 40px parent / 32px child icons, `pl-10` children, a 2px trunk starting at the first child row and curved branches. Regression proof: `BaseButton` now owns inline flex alignment; Quay lại and Nhóm mới both render icon plus label in one centered row. Base card/form/row radii were reduced one token step; semantic circles and pills remain intact. Browser check at `/`: the center create button is vertically contained in the bottom-navigation frame. |
| Database migrations | Dev PostgreSQL migration CLI | pass | `go run ./cmd/migrate up && go run ./cmd/migrate version` in `backend/`; version `8`, `dirty=false`; SQL migrations own schema and the system icon catalog. |
| Default category catalog | Dev PostgreSQL migration, Go test and browser list | pass | `go run ./cmd/migrate up && go run ./cmd/migrate version && go test ./...` in `backend/`; version `8`, `dirty=false`; migration `000008` assigns catalog icon keys for deployment as well as development. |
| API routing and Swagger generation | Go test + code generation | pass | `go generate ./cmd/api && go test ./...` in `backend/`; generated Swagger includes the implemented applicable-wallet endpoint. |
| Group management CRUD | Use-case tests, FE build and owner closure | pass | `go test ./...`, `go generate ./cmd/api`, `npm run typecheck && npm run build` pass. Use case proves owner scope, level/cycle validation and re-parenting children after a personal-parent deletion. System metadata `PATCH` is refused while the dedicated applicable-wallet endpoint remains permitted. Owner requested ticket closure after UI review. |

| Area | Proof | Result | Evidence |
| --- | --- | --- | --- |
| Frontend routing | TypeScript compile check | pass | `npm run typecheck` in `app/` |
| Schema and system catalog | Backend unit test | pass | `go test ./...` in `backend/` |
| Categories API auth guard | HTTP manual check | pass | `GET /api/v1/categories` without bearer returned `401` |
| Unified response contract | Go unit tests | pass | `TestOKUsesDataEnvelope`, `TestFailUsesProblemDetails`; `go test ./...` |
| Group management UI | Frontend production build | pass | `npm run typecheck && npm run build` in `app/`; runtime requires authenticated browser session |

- CORS preflight: `OPTIONS /api/v1/profile` with `Origin: http://localhost:4173` returned `204` with allow-origin, credentials, methods and headers.
- API health: `GET /api/v1/health` returned `{"status":"ok"}`.
- Google OAuth start: `GET /api/v1/auth/google` returned `302` to Google with redirect URI `http://localhost:4173/api/v1/auth/google/callback`; Vite proxies the callback to the backend on port 8080.
- Swagger UI: `GET /api/v1/docs/index.html` and `GET /api/v1/docs/doc.json` returned `200`; generated paths cover only the implemented auth/profile/home/health endpoints.

Policy lives in `docs/standards/VALIDATION.md`. This matrix is runtime project state and should change as work is planned, implemented, changed, or retired.

### Wallet and basic ledger runtime verification

- Status: core wallet management and the approved basic income/expense ledger are implemented and verified for review. [Wallet evidence](tickets/TICKET-01-02-VERIFICATION.md) · [Ledger evidence](tickets/TICKET-02-01-VERIFICATION.md).
- `rtk go test ./...` in `backend/`: pass, including owner-scoped handler rules, wallet type validation, category-wallet applicability, credit-ledger exclusion and derived balances.
- `rtk go run ./cmd/migrate up` and real HTTP UAT against dev PostgreSQL: pass. Restricted category returned `400`; create/update/delete balance sequence was 115,000 → 113,000 → 93,000; deleting the wallet cascaded its final transaction.
- `rtk go generate ./cmd/api`: pass; Swagger includes wallet `current_balance` and current transaction endpoints.
- `npm run test:transactions`, `npm run test:design`, and `npm run build` in `app/`: pass.
- Chrome UAT through the Vite proxy: create wallet, create expense, list, edit, synchronized Header/Wallet balance, and cleanup all passed. Two stale cross-panel refresh gaps were found during UAT, fixed, and rechecked.
- UAT boundary: receipt/OCR, jar assignment, transfer/adjustment and credit ledger are not required for the approved basic slice and remain open BA/follow-up scope. Human acceptance sign-off remains human-owned.

## Status Values

| Status | Meaning |
| --- | --- |
| planned | Accepted as intended behavior, not implemented |
| in_progress | Actively being built or verified |
| implemented | Implemented and evidence exists |
| changed | Contract or expected proof changed after earlier implementation |
| retired | No longer part of the accepted project contract |

## Matrix

| Requirement | Phase | Ticket/Bug | Contract/Behavior | Unit | Integration | E2E | UAT | Platform/Manual | Docs Review | Status | Evidence |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| REQ-02 | not_applicable | [TICKET-01-02](tickets/TICKET-01-02-quan-ly-vi.md) | Core wallet CRUD, type fields, permanent-delete impact, derived/current total balance | yes | yes | yes | yes | not_required | yes | implemented | [Wallet verification](tickets/TICKET-01-02-VERIFICATION.md); human review remains open for later statement/adjustment scope |
| REQ-04 | not_applicable | [TICKET-02-01](tickets/TICKET-02-01-ghi-thu-chi.md) | Basic income/expense CRUD with wallet/category compatibility and synchronized balances | yes | yes | yes | yes | not_required | yes | implemented | [Ledger verification](tickets/TICKET-02-01-VERIFICATION.md); receipt/jar acceptance remains outside verified slice |
| REQ-01, NFR-04 | not_applicable | [CORE-03](tickets/CORE-03-TIME-JARS-MONTH-DETAIL_DESIGN.md) | UTC/date-only/account timezone and fixed VND | yes | yes | yes | yes | not_required | yes | in_progress | Automated calendar/API/DB proof and migration 15 recorded above; owner timezone-switch and visual UAT pending |
| REQ-02–05, REQ-09–11 | not_applicable | [BA tickets](tickets/README.md) | Wallet, transaction, budget, savings/credit/debt accounting | yes | yes | yes | yes | not_required | yes | planned | Business rules documented; open linked-deletion/credit scenarios remain |
| REQ-06 | not_applicable | [CORE-03](tickets/CORE-03-TIME-JARS-MONTH-DETAIL_DESIGN.md) | Optional jar grouping, soft warnings, monthly configuration and cumulative view | yes | yes | yes | yes | not_required | yes | in_progress | PostgreSQL initialization/ownership tests, expense-only selector test and build pass; real owner CRUD/history UAT pending |
| REQ-07–08 | not_applicable | [BA tickets](tickets/README.md) | Ordinary recurring and automatic Travel Mode excluding recurring | yes | yes | yes | yes | not_required | yes | planned | REC/TRV rules; backdating/catch-up decisions still open |
| REQ-12–13 | not_applicable | [CORE-03](tickets/CORE-03-TIME-JARS-MONTH-DETAIL_DESIGN.md), [TICKET-07-04](tickets/TICKET-07-04-tong-ket-thang-ai.md) | Recalculable monthly reports, independent notes, context/AI and Insider | yes | yes | yes | yes | not_required | yes | in_progress | Live numeric month API/note and automatic local completion implemented; UAT pending; AI narrative/context and Insider remain unimplemented |
| REQ-14 | not_applicable | [BA tickets](tickets/README.md) | Moving weighted-average portfolio | yes | yes | yes | yes | not_required | yes | planned | AST formulas documented; funding/historical deletion open |
| REQ-15–18, NFR-01–03, NFR-05–07 | not_applicable | [BA tickets](tickets/README.md) | Keys/API/AI/offline/data lifecycle | yes | yes | yes | yes | yes | yes | planned | Product contract consolidated; no runtime implementation evidence added |

Proof columns above describe required product validation, not test execution. The current [BA ticket index](tickets/README.md) maps these rules to draft parent/child tickets. Phases remain not applicable; ticket creation does not count as implementation proof. Owner document review is pending; no product UAT has been marked passed.

## Documentation Review — 2026-09-13

- Scope: canonical product/business/report docs, requirements/stories, docs guide, context/queue, roadmap cleanup and standards relocation.
- Docs review: passed for captured decisions, trace targets, obsolete product wording and unchanged relocated standards. Open business questions are explicitly retained in [BUSINESS_RULES.md](../requirements/BUSINESS_RULES.md).
- `rtk git diff --check`: pass, exit 0.
- Read-only Node check: 17 docs inspected, 49 Markdown/frontmatter file targets checked, zero missing targets; six product docs have expected review metadata (README excepted).
- Read-only comparison with Git HEAD: all nine relocated standards are byte-identical to their former tracked files.
- Runtime tests/UAT: not run; this update changes documentation only. Product validation remains planned and owner review remains pending.
- API/ERD/runtime docs: no runtime or schema implementation changed; existing unrelated architecture edits were preserved. Technical ADRs remain a prerequisite where future implementation introduces durable architecture/schema decisions, not evidence supplied by this review.
- Reconciliation: requirements, report contract, stories, context, queue note, roadmap and changelog updated. No implementation ticket/phase created.
- Removed material: two superseded product drafts and one inherited Harness CLI phase example. Canonical product content is in `docs/requirements/`; tracked originals can be recovered from Git history. Standards were moved, not discarded.

### Reproduce Link And Trace Check

```sh
rtk proxy node -e 'const fs=require("fs"),path=require("path");let files=["docs/README.md","docs/work/ROADMAP.md",...fs.readdirSync("docs/standards").map(n=>"docs/standards/"+n),...fs.readdirSync("docs/requirements").map(n=>"docs/requirements/"+n)];let failures=[],count=0;for(const file of files){let s=fs.readFileSync(file,"utf8");for(const m of s.matchAll(/\[[^\]]*\]\(([^)]+)\)/g)){const ref=m[1].split("#")[0];if(ref&&!/^[a-z]+:/i.test(ref)){count++;if(!fs.existsSync(path.resolve(path.dirname(file),ref)))failures.push(file+": "+ref);}}let fm=s.match(/^---\n([\s\S]*?)\n---/);if(fm)for(const m of fm[1].matchAll(/^\s+[a-z_]+:\s+(\S+\.md)\s*$/gm)){count++;if(!fs.existsSync(path.resolve(path.dirname(file),m[1])))failures.push(file+": trace "+m[1]);}}console.log(JSON.stringify({files:files.length,linksAndTraces:count,failures},null,2));if(failures.length)process.exit(1);'
```

## BA ticket review

- Owner request: parent tickets with smaller children, concise BA scope and acceptance criteria.
- Result: 11 parents and 31 children, all `draft`; REQ-01 through REQ-18 covered. Common quality requirements are referenced by the shared ticket guide.
- Check: read-only ticket inventory verified unique IDs, draft status, reciprocal parent/child links, at least three acceptance criteria per child, no placeholders/trailing whitespace, and 106 valid Markdown file links. Pass, zero failures.
- `rtk git diff --check`: pass, exit 0.
- Docs review: scope and acceptance use business language; open questions remain in affected tickets; Reports/AI responsibilities avoid separate conflicting formulas.
- UAT: pending future implementation and owner acceptance. Runtime tests not required for this documentation-only breakdown.
- Reconciliation: ticket index, backlog, context, product docs entry point, roadmap note and changelog updated. No runtime, API/schema or architecture change; no technical ADR needed for the breakdown.

## Rules

### Category UI and ASCII ERD docs review

- Naming review: owner chọn `user`; sơ đồ và FK dùng tên đích, hiện trạng mapping cũ chỉ ghi trong mục migration. Kiểm tra `rtk git diff --check`; chưa chạy migration hoặc runtime tests cho thay đổi tên trong docs.

- Phạm vi: Tài khoản → Quản lý nhóm → chọn nhóm → Sửa; nhóm cha, category và ví áp dụng. [UI design](../design/system/DESIGN.md#account--quản-lý-nhóm) liên kết [ASCII ERD](../architecture/ERD.md#ascii--mô-hình-ví-và-nhóm-để-review) và [ticket](tickets/TICKET-01-03-tong-vi-danh-muc.md).
- Docs review: đã phân biệt user model hiện có và quan hệ nhóm/ví đề xuất; giữ catalog cũ; không tự chốt ý nghĩa category, scope rỗng hoặc chiến lược seed. Base component reuse bắt buộc.
- Kiểm tra: `rtk git diff --check` và kiểm tra file đích của Markdown links bằng Node. Runtime/UAT không chạy vì thay đổi chỉ ở docs; UI/API/migration nhóm chưa implement. Review mô hình dữ liệu còn pending; chưa có quyết định kiến trúc được áp dụng cần ADR implementation.

### Wallet research and docs review — 2026-09-13

- Evidence: official Money Lover support for balance adjustment, wallet management, goal/credit wallets; exact URLs and limits recorded in [DESIGN-01-02](tickets/TICKET-01-02-DETAIL_DESIGN.md).
- Reviewed old category catalog `0011_phase002_category_catalog.sql` and research sections 17–18; pinned reuse in wallet/category tickets.
- Corrected deletion/duplicate-name decisions across design, backlog, ERD, architecture and context. Research is not app UAT; no wallet API/schema/seed executed. Full technical design still needs review.
- Documentation-only verification: diff whitespace and local Markdown link checks; product tests/UAT not run for this research update.

### Finance Assistant planning and implementation — 2026-09-22

- [AI-ADVISOR-01](tickets/AI-ADVISOR-01-DETAIL_DESIGN.md) and [V1 task plan](../superpowers/plans/2026-09-22-finance-assistant-v1.md): design remains in review; advisor implementation is **partially shipped locally** as a JWT/API-key synchronous JSON pilot.
- Docs proof: `rtk git diff --check` passes; Node local-link check reports 8 valid relative targets across the new design/plan. Reviewed current code evidence and official chat-library docs; no dependency installed.
- Runtime/provider/E2E/UAT: backend race/DB and frontend type/design/build are proven for the local slice; local user API-key unit, PostgreSQL, middleware and Redis stream audit proof is now present. Configured provider, typed cards, SSE/recovery, durable audit delivery/alerting, browser state and public deployment remain unproven. Detailed gates are specified in the design/plan.
- API/schema/runtime changed for the local advisor slice (Stage v1 baseline plus JWT routes/UI). Existing transfer work remains independent; Feedback implementation and deployment checks remain separate.

Technical-plan follow-up:

- [Implementation guide](../superpowers/plans/2026-09-22-finance-assistant-technical-guide.md) added; [task plan](../superpowers/plans/2026-09-22-finance-assistant-v1.md) expanded to 8 tasks / 42 ordered subtask gates. Covers SQL/Go ports, tool defaults, recovery/drilldown APIs, SSE/React, exact accounting fixtures and operational limits.
- Code evidence from plan review: jar `ListMonth` can initialize config, API helper buffers response text, composer copy is entry-specific, navigation had only five existing slots, and migrate CLI has no down command. The runtime slice now addresses the jar path, composer/navigation wiring, query/overview boundary, local API-key auth and best-effort Redis access audit; SSE/recovery/durable audit remain open.
- Docs validation: local Node checker over design/plan/guide verified 12 relative link targets, 4 JSON examples and 27 paired fenced blocks; no failures. `rtk git diff --check` passes. Placeholder scan of plan/guide returned no matches.
- Runtime/code snippets in the guide remain instructional; the actual local implementation is verified in the vertical-slice checkpoint below. Public docs now include the implemented tool catalog, routes and API-key contract; proposed SSE/recovery contracts remain marked as unreleased.

Feedback implementation completion checkpoint — 2026-09-22:

- PASS: backend `rtk go test -race ./...` after wiring `main.go`, lifecycle locking, audit context, handler validation and Swagger annotations.
- PASS: frontend `rtk npm run typecheck`, `rtk npm run check:design`, `rtk npm run build` after Account → Feedback route/panel/service.
- PASS: `rtk go generate ./cmd/api`; generated Swagger includes feedback/changelog models and routes.
- PASS: public/agent API contract documented in [FEEDBACK_API](../architecture/FEEDBACK_API.md). Redis audit remains best-effort after commit; no finance access is granted to the service token.
- Historical as-of 2026-09-22: browser feedback E2E and real API roundtrip were not run because no dedicated script existed at that checkpoint. This residual risk is superseded by the 2026-10-06 feedback contract and browser-proxy E2E evidence below; it does not represent the current validation state.

Feedback storage-fallback regression — 2026-10-06:

- PASS: `cd app && npm run test:feedback` proves a screenshot multipart request retries exactly once as text-only for `FEEDBACK_SCREENSHOT_INVALID` and `FEEDBACK_SCREENSHOT_STORAGE_UNAVAILABLE`, while unrelated provider failures are not retried.
- PASS: `cd app && TEST_API_ORIGIN=http://127.0.0.1:4173 TEST_DATABASE_URL=postgres://dev:password@127.0.0.1:5432/mypocket_simulator_20261006?sslmode=disable npm run test:feedback:e2e` covers the real Chromium flow through Vite → Gin → PostgreSQL, controlled storage failure, JSON fallback, persisted owner row and cleanup.
- PASS: native `testAPIBackedFeedbackBubbleSubmitsWhenScreenshotStorageUnavailable` covers the same controlled fallback on the iPhone 16 simulator; full native UI suite remains 20/20.

Finance Assistant Task 1 foundation — 2026-09-22:

- PASS: entity normalization tests cover invalid/reversed/overlong calendar ranges, DST next-local-midnight conversion, JavaScript-safe money bounds, filter defaults/deduplication and zero-baseline percentage semantics.
- PASS: PostgreSQL fixture through `TEST_DATABASE_URL=postgres://dev:password@127.0.0.1:5432/postgres?sslmode=disable go test ./internal/infrastructure/repository -run TestFinanceQueryPostgresSummary -count=1 -v` proves owner scope, 1000/150/850 report totals, 55 full-scope rows versus a 50-row page, cursor replay/scope rejection, category subtree filtering, baseline-zero comparison and no jar/config row mutation.
- PASS: shared `FinanceQueryService` delegates validated owner-scoped bundles; full backend suite remains green. Advisor persistence/provider/HTTP/UI proof is recorded in the vertical-slice checkpoint below; SSE is still open.

Finance Assistant local vertical slice — 2026-09-22:

- PASS: migration `000018_advisor_conversations` applied locally (`version: 18 dirty: false`); PostgreSQL proof covers one conversation per owner, same-request replay, changed-payload conflict, active-run busy guard and cancellation.
- PASS: tool registry/provider/orchestrator/parts tests cover eight read-only tools, `get_transaction` owner isolation, unknown/owner override rejection, provider tool-call round trip, bounded loop, server-built view parts and no write/SQL/confirmation tools.
- PASS: JWT advisor routes now include a model-free `/overview` summary, real Account navigation slot, read-only composer mode, owner-scoped localStorage, masked assistant values, typecheck/design/build and full backend race/DB tests. The local provider-backed chat remains disabled when AI credentials are absent, while overview/history remain available.
- PASS: local API smoke with fixture login against `127.0.0.1:18080` returned `enabled=false` without AI credentials and a real account-local September overview (`income=0`, `expense=0`, `net=0`, `count=0`) without a model call; the route was then stopped. No provider secret was logged or committed.
- PASS: local user API-key creation/list/revoke, digest-only persistence, scope implication/denial, expiry/revocation and advisor middleware owner-isolation tests; the Stage v1 baseline applied to a fresh disposable database (`version: 1 dirty: false`).
- PASS: advisor auth audit writes redacted access metadata to the dedicated Redis stream `mypocket:audit:advisor`; integration proof ran against local Redis.
- NOT RUN: real browser chat round-trip through a configured provider, SSE/reconnect transport, durable audit delivery/alerting and production deployment. These are release gates, not claimed by this local slice.

- Add or update a row when a requirement, ticket, bug, public contract, or accepted behavior is created or changed.
- Mark proof columns `yes`, `no`, or `not_required`.
- Link evidence to `docs/templates/TEST_VERIFICATION.md`, ticket verification sections, UAT, docs review, release notes, or command output summaries.
- Do not set `implemented` until required proof has evidence.
- If a proof type is `not_required`, record the reason in the linked ticket, bug, or verification artifact.

### AI Advisor bug-fix verification — 2026-09-22

| Area | Evidence | Result |
| --- | --- | --- |
| Tool contract | `go test ./internal/usecase -run TestAdvisorToolDefinitionsExposeModelParameters` | PASS: all eight read tools expose model-visible properties; range/month required fields are explicit |
| Provider wire contract | `go test ./internal/infrastructure/ai -run TestAdvisorClientEncodesOpenAICompatibleToolCalls` | PASS: assistant calls use `type=function` and nested function name/arguments; tool result keeps `tool_call_id` |
| Orchestration cap | `go test ./internal/usecase -run TestAdvisorOrchestratorAllowsFinalAnswerAfterFourToolCalls` | PASS: four read calls may be followed by one final answer; further calls remain rejected |
| Credential lifecycle | `go test ./internal/usecase -run 'TestAdvisorOrchestratorRevalidatesPrincipalBeforeProviderAndTool|TestUserAPIKeyValidatePrincipalRejectsRevokedKey'` | PASS: provider/tool boundaries re-read principal; revoked key is rejected |
| Cancellation commit guard | `TEST_DATABASE_URL=... go test ./internal/infrastructure/repository -run TestAdvisorPostgresStartRunIsIdempotentAndBusyScoped` | PASS: cancelled run rejects late assistant message with lease-lost error |
| In-process cancellation | `go test -race ./internal/usecase -run TestAdvisorServiceCancelInterruptsInFlightProvider` | PASS: cancel endpoint path cancels the provider context and submit returns an error |
| History pagination | `npm run test:ai-history` | PASS: older pages merge without duplicate boundary messages and remain chronological |
| Context bound | `go test ./internal/usecase -run TestBuildAdvisorContextKeepsOnlyTwelveRecentMessages` | PASS: provider context keeps the newest 12 messages in chronological order |
| Assistant cards | `npm run typecheck && npm run check:design && npm run build` | PASS: normalized summary/comparison/search/budget plus wallet/goal/jar cards compile and satisfy base/design guards |

The above is local regression proof. Configured-provider browser chat, SSE/reconnect, durable Redis audit delivery/alerting and production deployment remain unverified release gates.

### Native simulator mutation smoke — 2026-10-06

- PASS: `testAPIBackedWalletDetailCreatesAndUpdatesTransaction` created an
  expense through the signed SwiftUI editor, observed the persisted row in
  Wallet Detail, edited its note, and observed the updated row.
- PASS: focused test completed in 18.348 seconds against fresh disposable
  `mypocket_stage_v1` PostgreSQL with fixture wallet teardown.
- PASS: the complete signed API-backed iPhone 16 UI suite passed 11/11 in
  218.318 seconds. This extends the existing 47/47 native XCTest proof with a
  real create/update UI mutation assertion.
- NOT RUN: physical-device, real Google OAuth, provider-backed AI/OCR and
  authenticated staging/production gates remain external checks.

### Native simulator rerun — 2026-10-06 06:26–06:30 ICT

- PASS: signed `xcodebuild test` against iPhone 16 Simulator
  `EC80B092-CEF1-4A8E-A5A3-7420298EFD4D` ran the API-backed
  `MyPocketIOSUITests` suite with 11/11 tests passing in 219.468 seconds
  (0 failures).
- PASS: the suite exercised authenticated overview, AI entry capability,
  Assistant overview, credit wallet ledger/actions, weekly ledger/cashflow,
  wallet transaction create/update, all secondary routes, feedback entry,
  primary tabs, and transfer/adjustment overflow actions.
- PASS: the disposable Stage v1 API remained healthy (`GET /api/v1/health`
  returned 200) and post-run fixture queries returned zero `Fixture*` wallets,
  transactions, or feedback rows.
- PASS: the signed debug app was reinstalled and launched with the local
  authenticated fixture on the same simulator for manual testing; screenshot
  evidence: `/tmp/mypocket-ios-simulator-ready.png`.
- NOT RUN: physical-device, real Google OAuth, provider-backed AI/OCR and
  authenticated staging/production gates remain external checks.

### Native business mutation rerun — 2026-10-06 06:50–06:59 ICT

- PASS: the wallet editor now uses an explicit create/edit presentation payload;
  signed UI created `Fixture UI Wallet`, renamed it to
  `Fixture UI Wallet Updated`, and observed the server-backed list refresh.
- PASS: signed transfer UI created and deleted the paired source/destination
  rows through the overflow action; signed credit UI created a purchase and a
  payment and observed both statement rows after refresh.
- PASS: complete signed `LOCAL_API_UI` `MyPocketIOSUITests` suite passed 14/14
  in 293.192 seconds on iPhone 16 Simulator.
- PASS: local API health stayed 200 and the disposable Stage v1 database was
  empty after teardown (`wallets=0`, `transactions=0`, `feedback=0`).
- NOT RUN: physical-device, real Google OAuth, provider-backed AI/OCR and
  authenticated staging/production gates remain external checks.

### Native Travel/Adjustment regression rerun — 2026-10-06 07:05–07:16 ICT

- PASS: unit/contract XCTest passed 48/48, including account-local adjustment
  timestamp coverage across a timezone boundary.
- PASS: signed API-backed `MyPocketIOSUITests` passed 14/14 in 292.296 seconds
  after replacing name-based Travel Mode selection with stable IDs and routing
  adjustment timestamps through the account timezone.
- NOT RUN: physical-device, real Google OAuth, provider-backed AI/OCR and
  authenticated staging/production gates remain external checks.

### Native simulator and cross-surface gate rerun — 2026-10-06 07:13–07:23 ICT

- PASS: `TEST_DATABASE_URL=postgres://dev:password@127.0.0.1:5432/mypocket_stage_v1?sslmode=disable go test -race ./... -count=1` in `backend/` completed with every package green.
- PASS: web `check:design`, `test:design` (7/7), typecheck, transaction/calendar/AI/history tests, production build, full FE-proxy API roundtrip, integrated transfer, real Chromium transfer E2E, and PWA release gate all passed.
- PASS: native unit/contract XCTest passed 48/48, the signed `LOCAL_API_UI` simulator suite passed 14/14 in 293.319s, and the Release build passed with `SWIFT_TREAT_WARNINGS_AS_ERRORS=YES`.
- PASS: base-component guard, project verification, API `/api/v1/health` (200), and post-run disposable counts (`wallets=0`, `transactions=0`, `feedback=0`).
- NOT RUN: physical-device, real Google OAuth, provider-backed AI/OCR and authenticated staging/production gates remain external checks.

### Native system-category scope fix — 2026-10-06 07:28–07:39 ICT

- PASS: system-category policy coverage now distinguishes read-only metadata
  and non-deletable system rows from personal categories; the native editor
  routes wallet applicability to `PATCH /api/v1/categories/{id}/wallets`.
- PASS: native unit/contract XCTest passed 49/49 after the policy and API
  contract change.
- PASS: signed iPhone 16 Simulator route sweep
  `testAuthenticatedMoreMenuOpensEveryDestination` passed 1/1 in 121.909s,
  including the Danh mục route and all 13 More destinations.
- NOT RUN: the full post-fix API-backed suite was not repeated because the
  simulator diagnostic service hung once; the prior full 14/14 signed suite
  remains valid for unchanged API-backed flows, and the targeted route sweep
  plus unit proof covers this change.

### Native spending-jars month/cumulative scope fix — 2026-10-06

| Area | Evidence | Result |
| --- | --- | --- |
| Month selector | `FinanceModelTests.testMonthOptionsStayBoundedAndIncludeAnchor` | PASS: account-local bounded options include the requested month and adjacent navigation range |
| Summary semantics | `JarsView` renders shared metric/status/progress components | PASS: actual income, actual spending, allocated, unassigned and advisory over-allocation states are visible |
| Active/history behavior | `JarsView` gates edit/remove actions by `item.active` | PASS: historical configurations stay read-only while ledger data remains visible |
| Cumulative report | selected-jar `jarReport` path plus empty-month clearing | PASS: report is scoped to the selected jar and stale results are cleared on month changes |
| Native contract | `xcodebuild test ... -only-testing:MyPocketIOSTests` | PASS: 50/50 tests |
| Release build | `xcodebuild build ... -configuration Release ... SWIFT_TREAT_WARNINGS_AS_ERRORS=YES` | PASS |
| Simulator route | signed `testAuthenticatedMoreMenuOpensEveryDestination` | PASS: 1/1 in 123.999s |

The prior complete signed API-backed simulator suite remains 14/14; it was not
repeated after this UI-only Jars change because a subsequent full run hit the
Xcode diagnostic service. This targeted route, unit, and Release evidence is
the current post-change proof.

### Native goal-wallet savings summary fix — 2026-10-06

| Area | Evidence | Result |
| --- | --- | --- |
| Savings math | `FinanceModelTests.testSavingsProgressClampsTargetAndBalanceSafely` | PASS: negative/zero targets are safe; remaining and progress are clamped |
| Target-day semantics | `FinanceModelTests.testSavingsTargetDayUsesAccountTimezoneAndCalendarDate` | PASS: date-only deadline uses the account timezone across UTC boundary |
| Goal detail UI | `WalletDetailView` uses shared metric/surface/text/progress components | PASS: reached/remaining, due-day state, progress and target amount are rendered |
| Native contract | `xcodebuild test ... -only-testing:MyPocketIOSTests` | PASS: 52/52 tests |
| Release build | `xcodebuild build ... -configuration Release ... SWIFT_TREAT_WARNINGS_AS_ERRORS=YES` | PASS |
| Simulator route | signed `testAuthenticatedMoreMenuOpensEveryDestination` | PASS: 1/1 in 124.194s |

The full API-backed suite remains 14/14 from the preceding simulator gate; this
goal summary change was covered by the targeted route, unit and Release checks.

### Native portfolio draft/refresh fix — 2026-10-06

| Area | Evidence | Result |
| --- | --- | --- |
| Asset draft | `FinanceModelTests.testPortfolioAssetDraftTrimsAndRejectsWhitespaceOnlyValues` | PASS: whitespace-only drafts reject and valid values are normalized before API submission |
| Refresh affordance | `PortfolioView` uses the shared `MPIconButton` header base | PASS: current summary can be refreshed without leaving the route |
| Native contract | `xcodebuild test ... -only-testing:MyPocketIOSTests` | PASS: 53/53 tests |
| Release build | `xcodebuild build ... -configuration Release ... SWIFT_TREAT_WARNINGS_AS_ERRORS=YES` | PASS |
| Simulator route | signed `testAuthenticatedMoreMenuOpensEveryDestination` | PASS: 1/1 in 124.144s |

The complete API-backed suite remains 14/14 from the preceding simulator gate;
this portfolio presentation/input change is covered by the targeted route,
unit and Release checks.

### Native feedback draft normalization fix — 2026-10-06

| Area | Evidence | Result |
| --- | --- | --- |
| Feedback payload | `FinanceModelTests.testFeedbackDraftTrimsAndRestrictsType` | PASS: title/description are normalized and unsupported type values reject |
| JSON/multipart parity | `FeedbackViewModel.submit` passes the normalized model to both client paths | PASS |
| Native contract | `xcodebuild test ... -only-testing:MyPocketIOSTests` | PASS: 54/54 tests |
| Release build | `xcodebuild build ... -configuration Release ... SWIFT_TREAT_WARNINGS_AS_ERRORS=YES` | PASS |
| Simulator feedback flow | `testAuthenticatedMoreRoutesAndFeedbackEntry` | PASS: 1/1 in 9.821s |

### Native month-summary scope fix — 2026-10-06

| Area | Evidence | Result |
| --- | --- | --- |
| Typed payload | `MonthSummaryModel.jarSummary` | PASS: backend `jar_summary` is decoded without losing existing fields |
| Month selection | shared `MPFormat.monthOptions` and `MPOptionSelector` | PASS: free-form month entry is replaced by bounded account-local options |
| Summary coverage | `MonthsView` | PASS: status, boundaries, category counts, jar usage and note controls render through base components |
| Dirty-note safety | `MonthsView` pending-month alert and failed-save retention | PASS: changing month requires confirmation when the note draft is dirty |
| Native contract | `xcodebuild test ... -only-testing:MyPocketIOSTests` | PASS: 54/54 tests |
| Release build | `xcodebuild build ... -configuration Release ... SWIFT_TREAT_WARNINGS_AS_ERRORS=YES` | PASS |
| Simulator route | signed `testAuthenticatedMoreMenuOpensEveryDestination` | PASS: 1/1 in 122.295s |

### Native simulator gate after transaction/AI proposal parity — 2026-10-06 10:08–10:14 ICT

- PASS: native unit/contract XCTest completed 61/61 on the booted iPhone 16
  Simulator, including transaction hũ/report/time fields and AI proposal
  hũ/date/report validation.
- PASS: full `LOCAL_API_UI` API-backed XCTest completed 15/15 in 306.876s;
  the suite covered wallet create/update, transaction create/update, credit
  purchase/payment, transfer pair create/delete, Finance Assistant overview,
  AI proposal review, all More destinations, feedback entry and primary route
  navigation.
- PASS: local fixture session and wallet cleanup left no `Fixture*` wallets
  after teardown; local `/api/v1/health` remained 200.
- PASS: base-component guard, project verification, `git diff --check`, and
  Release simulator build with `SWIFT_TREAT_WARNINGS_AS_ERRORS=YES` all passed.
- NOT RUN: physical-device, real Google OAuth, provider-backed AI/OCR and
  authenticated staging/production gates remain external checks. The simulator
  AI/OCR attachment test intentionally verifies the unavailable-provider state
  on the disposable local API; it is not evidence of provider availability.

### Native wallet-detail immutable-row fix — 2026-10-06 10:23–10:30 ICT

- PASS: native unit/contract XCTest completed 62/62 on the booted iPhone 16
  Simulator. `TransactionDetailPolicy` now keeps ordinary, transfer,
  adjustment and credit actions within their server-supported boundaries, and
  adjustment amounts honor `adjustment_direction`.
- PASS: targeted wallet transaction create/edit UI regression completed 1/1;
  the test now follows the user-visible row → detail → edit flow.
- PASS: full `LOCAL_API_UI` API-backed XCTest completed 15/15 in 310.705s,
  including wallet/transaction/credit/transfer mutations, AI proposal review,
  Finance Assistant overview, feedback entry and all secondary destinations.
- PASS: local health returned 200 and fixture-wallet cleanup returned no
  `Fixture*` wallets after teardown.
- PASS: Release simulator build with
  `SWIFT_TREAT_WARNINGS_AS_ERRORS=YES`, base-component guard, project
  verification and `git diff --check` all passed.
- NOT RUN: physical-device signing, real Google OAuth, provider-backed AI/OCR
  and authenticated staging/production gates remain external release checks.

### Native simulator deterministic-session regression — 2026-10-06

| Area | Evidence | Result |
| --- | --- | --- |
| Authenticated UI isolation | `AuthViewModel.makeDefaultStore()` uses `InMemorySessionStore` for both `-UITestAuthenticated` and `-UITestAuthenticatedAPI` | PASS: generic launch tests cannot read a stale Keychain session or production data |
| More-route sweep | `testAuthenticatedMoreMenuOpensEveryDestination` on iPhone 16 Simulator | PASS: 1/1; all 14 destinations open and return without terminating/relaunching the app between routes |
| Native contracts | `xcodebuild test ... -only-testing:MyPocketIOSTests` | PASS: 65/65 |
| Default UI smoke | `xcodebuild test ... -only-testing:MyPocketIOSUITests` without `LOCAL_API_UI` | PASS: 8 executable UI tests; provider/API fixtures are skipped as designed |
| API-backed UI | Each of the 9 `LOCAL_API_UI` tests run as an isolated xcodebuild invocation against the disposable local API | PASS: 9/9; wallet, transaction, transfer, credit, Assistant, AI entry and route coverage |

The Xcode 27 test runner can hang/restart when all API-backed cases are passed
as one multi-test operation after a prior UI run. The same cases pass in
isolated invocations with the local fixture; release gating must use the
isolated command list until the runner issue is resolved upstream.

### Native wallet-detail category-label assertion correction — 2026-10-06

The wallet-detail mutation test must assert the persisted amount/row identity,
not the fallback label `Khoản chi`. Ordinary rows now render their selected
category name when category metadata is available (the required-category
contract), so a fallback-label assertion can report a false failure even when
the transaction is persisted and visible. The verification will use the
formatted amount as the stable UI signal, then continue through row → detail →
edit → save and assert the updated note.

The wallet-detail editor must not be opened until its transaction and category
metadata load has completed. The add action is therefore gated by a local
`detailsReady` state; this prevents a fast simulator tap from opening an editor
with no applicable category and silently leaving the transaction unsaved.

### Native wallet-detail local API regression — 2026-10-06 14:21–14:32 ICT

- Root cause found and fixed: the custom native snake-case decoder converted
  `wallet_ids` to `walletIds` instead of `walletIDs`, so category metadata failed
  to decode and the wallet detail entered its error state after a successful
  transaction create. The decoder now handles plural `ids`, and
  `FinanceModelTests.testCategoryPayloadDecodesPluralIDKey` locks the contract.
- PASS: targeted wallet-detail create → row → detail → edit → update UI test,
  1/1 on the booted iPhone 16 Simulator. The assertion taps the actual row
  button and verifies the persisted formatted amount before opening the detail
  action; it does not depend on the uncategorized fallback label.
- PASS: native unit/contract XCTest, 67/67.
- PASS: full `LOCAL_API_UI` suite, 17/17 in 287.945s, including wallet,
  transaction, transfer, credit, Assistant, AI entry, feedback and route
  coverage; all disposable fixtures were created through the local API flow.
- PASS: default UI suite, 17 executed with 11 intentional provider/API skips and
  0 failures; Release simulator build with warnings-as-errors, base-component
  guard, project verification and `git diff --check` all pass.
- PASS: local backend health remained 200 throughout the simulator run.

### Native Finance Assistant response readability — 2026-10-06

The iOS Assistant must render assistant prose as safe, readable structured
content instead of showing raw Markdown delimiters. The renderer will be a
shared native base component (no HTML or model-generated view code): headings,
bold/code spans, lists, fenced code blocks and pipe tables are parsed into
SwiftUI text/table primitives, with a plain-text fallback for malformed input.
User messages remain plain text.

- PASS: `MPMarkdownParser` contract covers headings, bold inline content,
  unordered lists and pipe tables; malformed content falls back to plain text.
- PASS: native unit/contract XCTest now runs 68/68, including the assistant
  Markdown parser regression.
- The same base renderer is used for persisted Finance Assistant assistant
  messages and one-shot AI-entry replies; user messages remain plain text.

### Native/web Assistant proposal and simulator regression — 2026-10-06 14:41–14:51 ICT

- PASS: `propose_create_wallet` is strict, owner-independent and no-write; Go
  `go test -race ./... -count=1` passed across all backend packages.
- PASS: web `npm run test:ai-advisor`, `npm run check:design`,
  `npm run typecheck` and `npm run build` all passed. The web card calls the
  existing wallet API only from an explicit confirmation action.
- PASS: the wallet form base-contract test and web typecheck/design/build pass
  after all VND money inputs moved to `AmountField`.
- PASS: native `xcodebuild test ... -only-testing:MyPocketIOSTests` passed
  68/68; Debug build, Release analyze/build with warnings-as-errors,
  base-component guard and project verification passed.
- PASS: default native UI suite passed 17/17 with 11 intentional provider/API
  skips; `LOCAL_API_UI` passed 17/17 against the disposable API and cleaned up
  its fixtures. The native confirmation card is wired to the owner-scoped
  wallet API and never writes from the model/tool loop.

### Finance Assistant product-context grounding — 2026-10-06

Every advisor run will prepend a bounded, versioned product handbook to the
model context. It explains the actual navigation, supported feature boundaries,
read-only tool contract and how to answer “where/how do I use this?” questions.
Financial facts still come only from owner-scoped tools; the handbook contains
no user data. A contract test will assert the system message is present and
forbids invented writes, unsupported banking claims and ungrounded amounts.

- PASS: `BuildAdvisorSystemMessage` prepends the bounded handbook before the
  twelve-message conversation context and includes the account timezone.
- PASS: `TestAdvisorOrchestratorPrependsProductHandbook` verifies the system
  role, navigation map, read-only boundary and tool-grounding rule.
- The current handbook explicitly reports wallet/category/budget/jar/transaction
  writes as unsupported; no assistant message may claim a mutation completed.

### Advisor wallet creation proposal — 2026-10-06

The wallet-creation feedback is closed without allowing a model to mutate the
ledger directly. A bounded `propose_create_wallet` tool validates a wallet
draft and returns an action card; the user must confirm in the native/web card,
which then calls the existing owner-scoped wallet API. The tool itself never
writes.

- PASS: strict proposal schema and type-specific validation cover basic, goal
  and credit wallets; unknown fields are rejected and irrelevant fields are
  stripped before the card is returned.
- PASS: focused Go usecase/controller tests cover proposal output and the
  existing owner-scoped wallet API boundary.
- PASS: web typecheck and native simulator build cover the confirmation-card
  integration; native action cards call `FinanceAPI.createWallet` only after
  explicit confirmation.

### iPhone 16 Simulator bounded UI rerun — 2026-10-06 17:06 ICT

- PASS: `xcodebuild test ... -only-testing:MyPocketIOSTests` executed 70/70
  native unit/contract tests with zero failures.
- PASS: all 19 API-backed UI cases were rerun as isolated xcodebuild
  invocations against the booted iPhone 16 Simulator and disposable local API:
  shell/feedback/overflow (5/5), More/AI entry/Assistant (5/5),
  ledger/wallet/portfolio (5/5), and credit/transfer/wallet mutation (4/4).
- PASS: each bounded `xcresult` reported `TEST SUCCEEDED`; no app crash or
  test failure was observed. The simulator remains the pre-device validation
  gate; physical signing, real Google OAuth, provider-backed AI/OCR and
  authenticated production feedback remain external gates.
- The portfolio UI fixture cleanup was also corrected to interpolate the
  owner-scoped asset ID when removing a pre-existing same-symbol fixture;
  this prevents stale test data from surviving a rerun.
- Feedback multipart handling now normalizes an unconfigured development
  storage pointer instead of passing a typed-nil S3 interface (which previously
  panicked with HTTP 500). The native client falls back to text-only feedback
  for the explicit screenshot-storage error codes; production keeps PNG capture
  when private storage is configured. JSON feedback/list roundtrip passed on
  the disposable API; multipart without storage now returns controlled 400.
- PASS: the new feedback fallback contract brings native unit/contract XCTest
  to 71/71. The API-backed feedback UI test covers capture → multipart storage
  error → text-only fallback → persisted owner feedback on the simulator.
- Counting the prior 19 bounded parity cases plus the new feedback case, the
  current simulator UI evidence is 20/20 with zero failures; the portfolio
  cleanup case was rerun after its interpolation fix as well.

### Full simulator and web regression rerun — 2026-10-06 17:32–17:37 ICT

- PASS: the complete `LOCAL_API_UI` XCTest suite ran on the booted iPhone 16
  Simulator (`EC80B092-CEF1-4A8E-A5A3-7420298EFD4D`), executing 20 tests with
  0 failures. This includes the feedback screenshot-storage fallback and the
  owner-scoped mutation/cleanup flows; the disposable API remained healthy.
- PASS: `go test ./... -count=1`, `go test -race ./... -count=1`, `go vet ./...`,
  `git diff --check`, native base-component guard and project verification all
  passed after the UI run.
- PASS: Release `xcodebuild analyze` with warnings-as-errors and the complete
  web design/type/domain/build suite passed. Vite still reports the existing
  advisory bundle-size warning only; it is not a build failure.
- NOT RUN: physical-device signing, real Google OAuth/OCR/provider calls and
  authenticated production feedback remain external gates and were not
  substituted with simulator evidence.
- PASS: web `npm run test:feedback` proves the browser retries exactly once as
  JSON for the two explicit screenshot-storage codes and never retries an
  unrelated feedback failure.
- PASS: `TEST_API_ORIGIN=http://127.0.0.1:4173 npm run test:feedback:e2e`
  exercised the real Chromium UI through Vite → Gin → PostgreSQL: capture,
  controlled local-storage error, JSON fallback, owner-row persistence and
  cleanup all passed.

### Portfolio refresh assertion race fix — 2026-10-06 17:57–18:03 ICT

- PASS: the targeted portfolio UI test first passed after changing the
  assertion to wait for the refreshed rendered value (`Giá hiện tại: 120.000`)
  rather than the static `Giá trị hiện tại` card title.
- PASS: the complete booted iPhone 16 Simulator `LOCAL_API_UI` suite then ran
  20 tests with 0 failures. The failure was in test synchronization only; no
  portfolio production code change was required.

### Confirmed simulator-first regression — 2026-10-06 18:08–18:14 ICT

- PASS: the complete `LOCAL_API_UI` XCTest suite was rerun on the booted
  iPhone 16 Simulator (`EC80B092-CEF1-4A8E-A5A3-7420298EFD4D`) against the
  disposable local API and database. All 20 UI tests passed with 0 failures
  in 344.875 seconds.
- PASS: the run covered shell/auth, feedback capture and fallback, all More
  destinations, AI-entry attachment/proposal review, Assistant state,
  week-scoped ledger, wallet selection/detail/create/edit, transaction
  create/edit, transfer pair deletion, credit entry/payment, and portfolio
  price refresh.
- PASS: no app crash, assertion failure, or owner-scoping fixture leak was
  observed. The simulator remains the authoritative pre-device gate for this
  iteration.
- NOT RUN: physical-device signing, real Google OAuth, provider-backed
  AI/OCR, and authenticated production feedback remain external gates; this
  simulator run does not substitute for them.
- PASS: the post-UI native unit/contract run executed 71 tests with 0
  failures; backend `go test ./... -count=1`, `go test -race ./... -count=1`,
  and `go vet ./...` also completed successfully.

### Stateful secondary-screen simulator regression — 2026-10-06 18:45–18:56 ICT

- PASS: the full `LOCAL_API_UI` suite on iPhone 16 simulator executed 26 UI
  tests with 0 failures in 547.331 seconds.
- PASS: the added stateful flows verify budget/category create-edit-delete,
  jar create-edit-remove, month note save-delete, recurring create-edit-run-
  due-delete and Travel Mode create-activate-edit-delete. Assertions wait for
  rendered state after each API mutation.
- PASS: unique owner-scoped fixtures and teardown remove generated recurring
  transactions before deleting the disposable wallet; no fixture leak was
  observed.
- PASS: the budget/secondary forms reserve safe-area hit space so controls do
  not overlap the shell's persistent bottom navigation on the compact phone
  viewport.

### Final simulator gate after recurring run-due coverage — 2026-10-06 18:58–19:08 ICT

- PASS: the complete `LOCAL_API_UI` XCTest suite executed 26/26 cases with
  zero failures on iPhone 16 Simulator `EC80B092-CEF1-4A8E-A5A3-7420298EFD4D`
  in 547.401 seconds; the latest `.xcresult` marks the test plan `Passed`.
- PASS: stateful flows cover budget/category create-edit-delete, jar
  create-edit-remove, month-note save-delete, recurring create-edit-run-due-
  delete with generated transaction cleanup, Travel Mode create-activate-
  edit-delete, plus the existing auth/shell, wallet, ledger, transfer,
  credit, portfolio, AI-entry and feedback cases.
- PASS: native unit/contract XCTest passed 71/71 cases with 0 failures;
  base-component guard, project verification and `git diff --check` passed.
  Backend `go test ./...` passed 231/231 in normal and race modes, and
  `go vet ./...` found no issues.
- PASS: post-run fixture checks returned empty wallets, budgets, recurring
  schedules and travel events. Empty jar identities are retained by design to
  preserve jar history after month removal.
- PASS: the unsigned Release simulator artifact built, installed and launched
  on the same iPhone 16 simulator; the Google sign-in gate rendered without a
  crash. This is a runtime smoke check, not real OAuth validation.
- PASS: the web PWA global button base now uses `touch-action: manipulation`
  to prevent accidental double-tap zoom while preserving pinch zoom; the new
  release-gate regression test is green alongside the full web build/design
  checks.
- BLOCKED EXTERNAL: the paired iPhone 15 Plus is discoverable and a local Apple
  Development identity exists, but physical signing fails because Xcode has
  invalid account credentials and no provisioning profile for
  `com.mypocket.ios`; provisioning updates cannot run without a valid Xcode
  account. Physical-device behavior therefore remains unverified.
- PASS: read-only production probes confirm `GET /api/v1/health` and
  `/docs/` are reachable, production OpenAPI is version 2.0 and contains the
  feedback/changelog, API-key and AI route families. Invalid API-key probes for
  feedback, advisor capabilities and API-key listing correctly return 401;
  no production mutation was performed.
## Unified visual-port checkpoint — 2026-10-07

- PASS: web `check:design`, design-doc link checks, TypeScript typecheck and
  production Vite build after adding the shared reference-card tokens and
  overview cards.
- PASS: native project verifier, simulator build and native unit/contract
  target. A simulator screenshot was captured at `/tmp/mypocket-native-auth-port.png`;
  the bottom navigation now stays at the safe-area edge even when the API is
  unavailable, and the balance/add-action hierarchy is visible.
- PASS: targeted native authenticated-shell and route smoke checks after
  updating the navigation assertions to the canonical four web routes. The
  full `testAuthenticatedMoreMenuOpensEveryDestination` loop opened and
  returned from all 15 secondary destinations, including wallet management, on
  the iPhone 16 simulator.
- PASS: the native bottom-navigation base now has an explicit intrinsic height;
  its safe-area background and centered add-action no longer cover More-menu
  rows. Reports, categories and every lower menu destination are physically
  tappable, not just accessible through an XCTest action.
- PASS: the Stitch archive is available at
  `/Users/dungxbuif/Downloads/stitch_my_pocket.zip`. Its exact token contract
  is mirrored by web/native primitives and documented in
  `docs/design/references/STITCH-MY-POCKET.md`; reference-card screenshots and
  the Money Lover device evidence (`/tmp/moneylover-iphone.png`) are treated as
  visual inputs only.
- OPEN: perform authenticated web/native visual UAT with live data. No
  production deploy or database mutation is included in this slice.

### Corrected compact-phone hit-test checkpoint — 2026-10-07

- PASS: `testAPIBackedWalletDetailExposesTransactionEditor` passed after the
  wallet-detail CTA moved into a nested safe-area inset with a 64px bottom
  cushion; the button no longer sits inside the shell navigation hit region.
- PASS: `testAPIBackedWalletDetailCreatesAndUpdatesTransaction` passed with
  real local API create/edit assertions and fixture cleanup.
- PASS: `testAPIBackedJarCreateUpdateRemoveFromNativeUI` passed after account
  destination rows were given hit-test ownership and the shell add/navigation
  layer was disabled for the account tab. This fixes the “Hũ chi tiêu” row at
  the compact iPhone bottom edge without changing route or API contracts.
- PASS: native unit/contract remains 82/82. These three corrected UI cases
  were rerun in isolated result bundles because the full Xcode UI runner is
  stateful and previously reported the pre-fix failures.
- OPEN: run a fresh all-in-one UI suite only if Xcode runner stability is
  needed for release evidence; the bounded result bundles are the authoritative
  regression artifacts for this iteration.
### Native profile-card account action — 2026-10-07

- Added the shared-base `Quản lý tài khoản` action to the native More profile
  card and verified that it opens `AccountView`.
- Command: `xcodebuild -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS -destination 'platform=iOS Simulator,id=EC80B092-CEF1-4A8E-A5A3-7420298EFD4D' -derivedDataPath /tmp/mypocket-profile-ui -parallel-testing-enabled NO SWIFT_ACTIVE_COMPILATION_CONDITIONS='DEBUG' -only-testing:MyPocketIOSUITests/LaunchTests/testMoreProfileCardOpensAccountSettings test`.
- Result: 1/1 passed, xcresult:
  `/tmp/mypocket-profile-ui/Logs/Test/Test-MyPocketIOS-2026.10.07_12-14-12-+0700.xcresult`.

### Native profile-card help action — 2026-10-07

- PASS: `testMoreProfileHelpOpensFeedback` taps the shared-base `Trợ giúp`
  button in the More profile card and opens `Phản hồi & changelog`.
- Result: targeted simulator UI test passed 1/1; xcresult:
  `/tmp/mypocket-profile-help-ui/Logs/Test/Test-MyPocketIOS-2026.10.07_12-22-42-+0700.xcresult`.

### Native overview notification action — 2026-10-07

- PASS: `testOverviewNotificationBellOpensNotifications` taps the shared
  header bell and opens the existing `NotificationsView` route.
- Result: targeted simulator UI test passed 1/1; xcresult:
  `/tmp/mypocket-overview-notification-ui/Logs/Test/Test-MyPocketIOS-2026.10.07_12-23-42-+0700.xcresult`.

### Native shared scroll safe-area — 2026-10-07

- PASS: `MPBottomNavigationInset` is the shared SwiftUI base for Overview, More
  and secondary feature scroll surfaces. It reserves the persistent shell
  navigation area through `safeAreaInset`; editor-specific save bars remain
  separate by contract.
- PASS: Debug simulator build with warnings-as-errors completed at
  `/tmp/mypocket-safearea-build`.
- PASS: Release simulator build with warnings-as-errors completed at
  `/tmp/mypocket-safearea-release`.
- PASS: `testOverviewNotificationBellOpensNotifications` remained green after
  the inset refactor; result:
  `/tmp/mypocket-safearea-ui/Logs/Test/Test-MyPocketIOS-2026.10.07_12-27-39-+0700.xcresult`.
- PASS: `testAPIBackedOverviewFinalSectionStaysAboveShellNavigation` scrolls a
  disposable fixture-backed Overview to the final recent-transactions section
  and asserts the section frame is above the shell navigation. Result:
  `/tmp/mypocket-safearea-reachability/Logs/Test/Test-MyPocketIOS-2026.10.07_12-30-45-+0700.xcresult`.
- PASS: `verify-project.sh`, `check-base-components.test.sh` and
  `git diff --check` completed without errors.

### Native Google authentication wiring — 2026-10-07

> Historical implementation record only. Superseded by the password-only
> migration entry at the current checkpoint; Google login is absent from the
> native app.

- PASS: `AuthView` exposes the shared-base `auth-google` action and
  `AuthViewModel.signInWithGoogle()` delegates to `GooglePKCEClient`, persists
  the returned access/refresh session, and enters the authenticated phase.
- PASS: placeholder/missing iOS OAuth configuration returns a user-facing
  configuration error without making a network request; unit/UI result:
  `/tmp/mypocket-google-auth-tests/Logs/Test/Test-MyPocketIOS-2026.10.07_12-32-35-+0700.xcresult`
  (2/2).
- OPEN: register a real iOS Google client with the exact
  `mypocket://oauth/callback` URI and run an authenticated device exchange.
- PASS: the complete native unit/contract target remains 83/83 after the
  Google auth wiring; result:
  `/tmp/mypocket-google-unit-full/Logs/Test/Test-MyPocketIOS-2026.10.07_12-33-25-+0700.xcresult`.
- PASS: the subsequent centralized Overview typography cleanup also passes the
  warnings-as-errors Debug build at `/tmp/mypocket-base-audit-build` and the
  full 83/83 unit target at
  `/tmp/mypocket-base-audit-unit/Logs/Test/Test-MyPocketIOS-2026.10.07_12-34-43-+0700.xcresult`.

### Native routing follow-up unit gate — 2026-10-07

- PASS: the complete native unit/contract target remains 82/82 after the
  profile-help and overview-notification callback changes.
- Result: xcresult:
  `/tmp/mypocket-profile-notification-unit/Logs/Test/Test-MyPocketIOS-2026.10.07_12-24-15-+0700.xcresult`.

### Web parity verification — 2026-10-07

- PASS: `check-design.mjs`, `check-design-docs.mjs`, TypeScript `tsc --noEmit`
  and the production Vite build all pass using the bundled Node runtime.
- Note: the host shell's `node` is unavailable by default; this verification
  used `/Users/dungxbuif/.cache/codex-runtimes/codex-primary-runtime/dependencies/node/bin/node`.

### Native More-menu route sweep after profile parity — 2026-10-07

- PASS: `testAuthenticatedMoreMenuOpensEveryDestination` opened and returned
  from all 15 secondary destinations, including the profile-card/account path.
- Command: `xcodebuild -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS -destination 'platform=iOS Simulator,id=EC80B092-CEF1-4A8E-A5A3-7420298EFD4D' -derivedDataPath /tmp/mypocket-more-sweep-profile -parallel-testing-enabled NO SWIFT_ACTIVE_COMPILATION_CONDITIONS='DEBUG' -only-testing:MyPocketIOSUITests/LaunchTests/testAuthenticatedMoreMenuOpensEveryDestination test`.
- Result: 1/1 passed in 344.968 seconds, xcresult:
  `/tmp/mypocket-more-sweep-profile/Logs/Test/Test-MyPocketIOS-2026.10.07_12-15-17-+0700.xcresult`.

### Native portfolio action anatomy parity — 2026-10-07

- PASS: the focused portfolio copy/ID contract passed 1/1 at
  `/tmp/mypocket-portfolio-action-unit-final`.
- PASS: the full native unit/contract target passed 107/107 at
  `/tmp/mypocket-portfolio-action-full-final`.
- PASS: web design guards, design-doc links, typecheck and Vite production
  build passed with the bundled Node runtime; native Release warnings-as-errors
  build passed at `/tmp/mypocket-portfolio-action-release-final`.
- PASS: project verification, base-component guard/tests and `git diff --check`
  passed. The bounded unauthenticated launch smoke passed 1/1 at
  `/tmp/mypocket-portfolio-action-ui-launch`.
- OPEN: the all-UI Xcode invocation was stopped after the known runner hang;
  disposable authenticated portfolio API mutation should be rerun when the
  local backend is available. Authenticated visual comparison remains open.

### Native credit summary-card parity — 2026-10-07

- PASS: focused credit summary/action contract at
  `/tmp/mypocket-credit-summary-unit-final` (1/1).
- PASS: full native unit/contract target at
  `/tmp/mypocket-credit-summary-full-final` (108/108).
- PASS: Release warnings-as-errors build at
  `/tmp/mypocket-credit-summary-release-final`; project verification,
  base-component guard/tests, `git diff --check`, web design/type/build checks
  also pass.
- PASS: bounded authenticated credit API UI smoke `testAPIBackedCreditWalletOpensLedgerActions`
  passed 1/1 at
  `/tmp/mypocket-credit-order-ui-final/Logs/Test/Test-MyPocketIOS-2026.10.07_15-34-55-+0700.xcresult` after the section-order fix.
- PASS: follow-up section-order/empty-state contract passed 1/1 at
  `/tmp/mypocket-credit-order-green`; the full native rerun passed 108/108 at
  `/tmp/mypocket-credit-order-full` and the Release warnings-as-errors build
  passed at `/tmp/mypocket-credit-order-release`.
- OPEN: authenticated visual UAT; the implementation preserves the existing
  statement, entry and payment API contracts and no production mutation was
  performed.

### Native portfolio summary-card parity — 2026-10-07

- PASS: focused subtitle/summary-copy contract at
  `/tmp/mypocket-portfolio-summary-green` (1/1).
- PASS: full native unit/contract rerun at
  `/tmp/mypocket-portfolio-summary-full` (108/108) and Release
  warnings-as-errors build at `/tmp/mypocket-portfolio-summary-release`.
- PASS: project/base-component guards, `git diff --check` and web design checks.
- OPEN: the disposable API portfolio UI rerun was stopped by the known Xcode
  runner hang; prior asset create/price-refresh UI evidence remains recorded.
  Authenticated visual UAT remains open.

### Native Auth layout parity — 2026-10-07

> Historical visual checkpoint. Its Google action assertions were removed by
> the later password-only migration decision; see the fresh auth checkpoint at
> the top of this matrix.

- PASS: RED/GREEN copy contract `testAuthCopyMatchesWebMobileContract`; GREEN
  result at `/tmp/mypocket-auth-layout-green/Logs/Test/Test-MyPocketIOS-2026.10.07_15-54-26-+0700.xcresult` (1/1).
- PASS: unauthenticated `testAppLaunches` now asserts the title, subtitle,
  email/password fields and Google action; result at
  `/tmp/mypocket-auth-ui-parity/Logs/Test/Test-MyPocketIOS-2026.10.07_15-54-42-+0700.xcresult` (1/1).
- PASS: full native unit/contract target 109/109 at
  `/tmp/mypocket-auth-full-tests/Logs/Test/Test-MyPocketIOS-2026.10.07_15-55-29-+0700.xcresult`.
- PASS: Release warnings-as-errors build at `/tmp/mypocket-auth-release`,
  project verification, base-component guard/tests and `git diff --check`.
- PASS: visual screenshot `/tmp/mypocket-auth-parity.png` shows the native
  circular mark, direct canvas title/subtitle, flat fields, primary sign-in,
  divider and compact Google action aligned to the web AuthGate screenshot.
- PASS: web typecheck, Vite production build, design guard and design-doc link
  checks pass with the bundled Node runtime (`dist` generated with the existing
  web sources).
- OPEN: this slice does not prove real Google OAuth, physical-device signing,
  authenticated visual UAT or provider-backed flows.

### Native authenticated Overview visual checkpoint — 2026-10-07

- PASS: launched the DEBUG native app with the local disposable API fixture
  (`-UITestAuthenticatedAPI`) and captured `/tmp/mypocket-overview-api.png`,
  then repeated after the glyph update at
  `/tmp/mypocket-overview-wallet-glyph.png`.
  The frame contains the owner-scoped balance header, wallet section with
  `Quản lý`, Money Insider card, report section and persistent feedback/add
  shell. The fixture used only local data; no production request or mutation
  was made.
- OPEN: the supplied Stitch Overview uses richer wallet icon artwork and a
  populated multi-wallet fixture; this screenshot proves native anatomy and
  API-backed rendering, not pixel approval for every icon/data variant.

### Native wallet glyph parity — 2026-10-07

- PASS: `MPWalletGlyph` centralizes 48pt circular wallet badge geometry and
  web-aligned basic/goal/credit tones (warning/success/teal); `MPWalletRow`
  now composes that base rather than owning a gray local icon treatment.
- PASS: focused semantic-token contract 1/1 at
  `/tmp/mypocket-wallet-glyph-contract-final/Logs/Test/Test-MyPocketIOS-2026.10.07_16-08-46-+0700.xcresult`.
- PASS: full native unit/contract target 110/110 at
  `/tmp/mypocket-wallet-glyph-final-full/Logs/Test/Test-MyPocketIOS-2026.10.07_16-09-21-+0700.xcresult`;
  Release warnings-as-errors at `/tmp/mypocket-wallet-glyph-final-release`,
  project/base/diff guards and web typecheck/build/design checks pass.
- OPEN: supplied Stitch uses bespoke bank/cash artwork; native symbols are the
  bounded platform equivalent until those assets are explicitly licensed or
  supplied.

### Native authenticated transaction-editor visual capture — 2026-10-07

- PASS: More/account screenshot attachment exported and inspected from
  `/tmp/mypocket-more-visual-capture/Logs/Test/Test-MyPocketIOS-2026.10.07_16-12-40-+0700.xcresult`;
  attachment `/tmp/mypocket-more-visual-attachments/8AD2A57B-93FC-4A57-9136-B1A71ADB0D62.png`.
- PASS: fixture-backed transaction editor screenshot exported and inspected at
  `/tmp/mypocket-transaction-api-ai-final-attachments/1C3FFCC1-E76B-4E97-A570-8E384B43E280.png`.
  The fixture is local and disposable; no production data was read or mutated.
- PASS: More now centers `Mở rộng` independently of the `Hỗ trợ` pill through
  `MPCenteredScreenHeader`; the focused profile route smoke remains 1/1 at
  `/tmp/mypocket-more-centered-header/Logs/Test/Test-MyPocketIOS-2026.10.07_16-35-18-+0700.xcresult`,
  with inspected attachment
  `/tmp/mypocket-more-centered-header-attachments/3B1118C1-AE96-4736-B481-71CAD1C76490.png`.
- PASS: editor no longer renders below the shell navigation. The shared
  full-screen modal is used by both the ledger and wallet-detail entry points;
  the shared footer also exposes the existing AI-entry route through a filled
  image action button.
- PASS: focused editor UI smoke 1/1 at
  `/tmp/mypocket-transaction-editor-guarded/Logs/Test/Test-MyPocketIOS-2026.10.07_16-30-49-+0700.xcresult`.
- PASS: focused API-backed editor UI smoke 1/1 at
  `/tmp/mypocket-transaction-api-guarded/Logs/Test/Test-MyPocketIOS-2026.10.07_16-31-24-+0700.xcresult`.
- PASS: full native unit/contract target remains 110/110 at
  `/tmp/mypocket-visual-full-tests-2/Logs/Test/Test-MyPocketIOS-2026.10.07_16-30-20-+0700.xcresult`;
  Release warnings-as-errors, project/base-component/diff guards and web
  typecheck/build/design checks are green.
- OPEN: visual approval on a physical owner device, licensed bespoke Stitch
  artwork, and live provider/OCR/Google/production feedback validation remain
  release gates.
# Native goal-wallet Stitch savings-ledger parity — 2026-10-07

- PASS: the API-backed disposable goal-wallet scope test now renders the
  Stitch-style savings ledger on the iPhone 16 simulator: compact wallet
  header, centered balance/remaining target, deadline/progress, advice,
  owner-scoped transaction state and add-transaction action. The test passed
  2/2 at `/tmp/mypocket-goal-focused-final-4/Logs/Test/Test-MyPocketIOS-2026.10.07_17-15-08-+0700.xcresult`.
- PASS: visual evidence was exported and inspected at
  `/tmp/mypocket-goal-focused-final-4-attachments/10590C2D-C217-43EE-972A-2D4C82B3D657.png` and
  `/tmp/mypocket-goal-focused-final-4-attachments/08298850-047B-44F4-9584-B40DF5E57694.png`.
- PASS: the Stitch ordering regression places the transaction scope before
  savings progress; the focused API-backed test is 1/1 at
  `/tmp/mypocket-goal-order-green4/Logs/Test/` with inspected capture
  `/tmp/mypocket-goal-order-green4-attachments/3F89495A-7555-47CD-B986-00A49E1449E6.png`.
- PASS: native unit/contract regression is 110/110 at
  `/tmp/mypocket-goal-order-full-unit/Logs/Test/`; Release
  warnings-as-errors build is `/tmp/mypocket-goal-order-release` and project,
  base-component and diff guards pass.
- Contract: goal scope removes ordinary week/custom period controls and uses
  shared `MPHeaderIconButton`, `MPWalletGlyph`, `MPSavingsBalanceBlock` and
  `MPSavingsProgressBlock`; search
  filters category/note text and overflow reloads or returns to aggregate.
- Remaining gate: authenticated device/browser visual UAT and populated
  provider/OCR/Google flows remain external gates; this fixture uses local API
  data only.

### Native goal-wallet shell CTA clearance — 2026-10-07

- RED: the first CTA clearance assertion reproduced the compact-phone overlap;
  the scoped goal CTA ended below the shell add affordance
  (`/tmp/mypocket-goal-cta-red2/Logs/Test/`, maxY 755.33 vs minY 730).
- PASS: the goal CTA now uses a nested safe-area region with the shared
  `MPButton`, canvas background, 104pt bottom cushion and 56pt visual lift.
  The focused API-backed goal route passed 1/1 at
  `/tmp/mypocket-goal-cta-green3/Logs/Test/` with both shell-clearance
  assertions green.
- PASS: inspected screenshot
  `/tmp/mypocket-goal-cta-green3-attachments/603309EC-18BF-451E-BAB3-FB1D3F7732B7.png`
  shows the entire “Thêm giao dịch” label above the shell navigation and
  centered add affordance.
- PASS: full native unit/contract target is 110/110 at
  `/tmp/mypocket-goal-cta-full-unit/Logs/Test/`; Release
  warnings-as-errors build is `/tmp/mypocket-goal-cta-release`, and project,
  base-component and diff guards pass.
- Physical-device and populated-provider UAT remain open.

### Native transaction ledger visual capture — 2026-10-07

- PASS: the ordinary fixture-backed ledger test renders wallet scope, week
  controls, cashflow totals and date-grouped rows, then scrolls the final rows
  above the persistent shell. Focused UI evidence is
  `/tmp/mypocket-ledger-visual-final/Logs/Test/` (1/1).
- PASS: inspected screenshot
  `/tmp/mypocket-ledger-visual-final-attachments/7F6A03B6-627B-4664-949F-B70A4E3AD26D.png`.
  No feature-local safe-area change was needed; the shared shell inset and
  row identifiers provide the reachability contract.
- OPEN: authenticated device pixel approval, populated production/provider
  data and live OCR/Google/provider gates remain external.

### Native secondary-route shell visual capture — 2026-10-07

- RED: category route lacked the shared bottom-navigation inset at
  `/tmp/mypocket-category-safe-area-red/Logs/Test/`.
- RED: SwiftUI rendered a duplicate native `Back` button above the custom
  Stitch header at `/tmp/mypocket-secondary-header-red/Logs/Test/`.
- PASS: `CategoriesView`, `PortfolioView`, `CreditView` and
  `NotificationsView` now use the shared `MPBottomNavigationInset`; the
  secondary destination host hides the native back button while preserving
  the custom header. Focused route proof is 1/1 at
  `/tmp/mypocket-secondary-header-green/Logs/Test/`.
- PASS: inspected category screenshot is
  `/tmp/mypocket-secondary-header-green-attachments/C1E759A4-27DF-4C37-AA20-D23F897175D.png`.
  The local fixture returned a category-data error, but the shell/header/inset
  assertions passed; this is not production data evidence.
- OPEN: full authenticated visual UAT across populated secondary routes and
  physical-device pixel approval remain open.
- PASS: a dedicated Notifications route smoke is 1/1 at
  `/tmp/mypocket-notifications-shell/Logs/Test/` with inspected attachment
  `/tmp/mypocket-notifications-shell-attachments/A02D696F-328A-4306-A51D-AC925C4E6316.png`.
  The local fixture returned a safe load-error card, so this proves shell and
  header anatomy only; seeded inbox data and owner mark-read UAT remain open.
- RED: the remaining scrollable secondary screens initially lacked the shared
  inset; `/tmp/mypocket-secondary-remaining-red2/Logs/Test/` failed for
  Assistant, AI Entry, Feedback and Account.
- PASS: all four now compose `MPBottomNavigationInset`; the focused regression
  is 1/1 at `/tmp/mypocket-secondary-remaining-green/Logs/Test/`. This closes
  the shell-clearance contract for every scrollable secondary screen without
  adding screen-local spacing.
- PASS: post-fix native unit target is 110/110 at
  `/tmp/mypocket-secondary-inset-full/Logs/Test/`; Release warnings-as-errors
  is `/tmp/mypocket-secondary-inset-release`, and project/base/diff guards
  pass.
- RED: Wallet management's editor initially had no inset and its save footer
  ended below the shell; `/tmp/mypocket-wallet-editor-inset-red/Logs/Test/`
  recorded maxY 802.0 versus shell top 766.33.
- PASS: the Wallets list and editor now use `MPBottomNavigationInset`; the
  focused contract is 1/1 at
  `/tmp/mypocket-wallet-editor-inset-green/Logs/Test/`.
- PASS: post-wallet-fix native unit target is 110/110 at
  `/tmp/mypocket-wallet-inset-full/Logs/Test/`; Release warnings-as-errors is
  `/tmp/mypocket-wallet-inset-release`, and project/base/diff guards pass.
- RED: transfer and balance-adjustment overlays opened from the transaction
  overflow menu initially lacked the shared inset; the focused test failed at
  `/tmp/mypocket-transfer-overlay-red/Logs/Test/`.
- PASS: both overlays now reserve the shell inset; focused regression is 1/1 at
  `/tmp/mypocket-transfer-overlay-green/Logs/Test/`.
- PASS: final full native unit target after all shell-clearance fixes is
  110/110 at `/tmp/mypocket-shell-all-full/Logs/Test/`; Release
  warnings-as-errors is `/tmp/mypocket-shell-all-release`, and project/base/diff
  guards pass.
- RED: the AI entry proposal editor overlay initially inherited the parent
  screen's shell but did not reserve its own inset; the scoped regression
  failed at `/tmp/mypocket-proposal-editor-red2/Logs/Test/`.
- PASS: `EntryProposalEditor` now uses `MPBottomNavigationInset` and keeps
  `Lưu đề xuất` above the persistent shell; focused proof is 1/1 at
  `/tmp/mypocket-proposal-editor-green3/Logs/Test/` with inspected screenshot
  `/tmp/mypocket-proposal-editor-green3-attachments/3E6684AD-F02A-4958-AEAB-7AEA7C9A69E4.png`.
- PASS: post-editor native unit target is 110/110 at
  `/tmp/mypocket-proposal-editor-full/Logs/Test/`; Release
  warnings-as-errors is `/tmp/mypocket-proposal-editor-release`, and project,
  base-component and diff guards pass.
- RED: the wallet-detail transaction detail overlay initially lacked its
  scoped shell inset; the API-backed mutation flow failed at
  `/tmp/mypocket-transaction-detail-red/Logs/Test/`.
- PASS: `TransactionDetailView` now uses `MPBottomNavigationInset`; focused
  API-backed proof is 1/1 at
  `/tmp/mypocket-transaction-detail-green3/Logs/Test/`, with inspected
  screenshot `/tmp/mypocket-transaction-detail-green3-attachments/1B7B79A5-7294-4BAD-97F1-732E6C1C50EF.png`.
- PASS: post-detail native unit target is 110/110 at
  `/tmp/mypocket-transaction-detail-full/Logs/Test/`; Release
  warnings-as-errors is `/tmp/mypocket-transaction-detail-release`, and
  project, base-component and diff guards pass.

### Provider and attachment capability wiring — 2026-10-07

- PASS: backend configuration keeps AI/OCR/provider secrets in process
  environment (or ignored local `.env.local` for development); no mobile or
  web artifact receives a provider key.
- PASS: `files_configured` is derived from both `OCRConfigured()` and the
  private attachment store, so a missing storage credential cannot be masked
  by a valid OCR credential.
- PASS: native AI Entry gates file selection and submission on the backend
  capability response; Release code has no fixture fallback. DEBUG-only UI
  fixtures set capabilities explicitly for deterministic attachment tests.
- PASS: configuration/controller/AI package tests pass (`135` tests across
  three Go packages); native project/base-component guards and `git diff
  --check` pass. The web design and local-link guards also pass (`131` valid
  local links).
- OPEN: current local environment has no OCR credential, so live image/S3 OCR
  was not rerun in this checkpoint. The prior synthetic OCR pipeline evidence
  remains recorded above; production OCR/provider and physical-device UAT need
  deployment credentials and a reachable device.

### Native feedback/changelog content parity — 2026-10-07

- RED: the native feedback list omitted the submitted description and the
  changelog description, and had no linked `Fixed in <version>` presentation.
- PASS: `FeedbackPresentation` resolves fixed feedback to its linked changelog
  and `MPFeedbackStatus` owns the shared status/version presentation. Native
  feedback now renders title, type, description, status, fixed date, linked
  release and full changelog descriptions using existing owner-scoped APIs.
- PASS: focused RED→GREEN contract ran, then the full native unit target passed
  `114/114` at `/tmp/mypocket-port-after-transaction/Logs/Test/`; a fresh
  `-configuration Release SWIFT_TREAT_WARNINGS_AS_ERRORS=YES build` also
  passed at `/tmp/mypocket-port-after-transaction-release`.
- OPEN: authenticated production feedback/storage and physical-device visual
  approval remain external gates.

### Native transaction day-group header parity — 2026-10-07

- RED: the new pure formatter contract initially failed because the native
  ledger had no three-part day-group presentation helper.
- PASS: `TransactionsPresentation.dayGroup` now formats day number, weekday
  and month/year in the account timezone, and the ordinary native ledger uses
  the same visual hierarchy as the web day cards. Category grouping keeps its
  existing category label path.
- PASS: focused RED→GREEN test
  `testTransactionDayGroupPresentationMatchesWebHierarchyAndTimezone` passed
  at `/tmp/mypocket-transaction-group-green/Logs/Test/`.
- OPEN: this is simulator/unit evidence; authenticated device pixel approval
  and production data UAT remain external gates.

### Native transaction category glyph parity — 2026-10-07

- RED: changing `MPTransactionRow` to the richer glyph contract initially
  exposed legacy call sites that still passed `isIncome`; the compatibility
  initializer was retained so overview/reports/credit surfaces keep their
  existing behavior while the ledger opts into category mapping.
- PASS: `TransactionsPresentation.glyph` covers the canonical web category
  families and `MPTransactionGlyph` owns the icon badge geometry/tone. Ordinary
  ledger rows now use category glyphs, income/expense/neutral amount tones and
  the same transfer/adjustment semantics as web.
- PASS: focused mapping test is 1/1 at
  `/tmp/mypocket-transaction-glyph-final/Logs/Test/`; full native unit target
  is 116/116 at `/tmp/mypocket-port-budget-final-unit/Logs/Test/`, Release
  warnings-as-errors is `/tmp/mypocket-port-budget-final-release`, and project,
  base-component, design and diff guards pass.
- OPEN: simulator evidence does not replace authenticated device pixel review.

### Native ended-budget preview parity — 2026-10-07

- RED: ended budget cards exposed only a terminal label; native could not open
  the read-only details that web exposes as `Xem / Xóa`.
- PASS: ended cards now open a read-only preview with name, limit, spent,
  dates, wallet/category scope and delete confirmation. Save/edit fields stay
  unavailable; active budgets retain the existing editor.
- PASS: policy contract `testEndedBudgetUsesReadOnlyPreviewAction` is 1/1 at
  `/tmp/mypocket-budget-preview-red/Logs/Test/`; full native unit target is
  116/116 at `/tmp/mypocket-port-budget-final-unit/Logs/Test/`, and Release
  warnings-as-errors is `/tmp/mypocket-port-budget-final-release`.
- OPEN: authenticated device visual approval remains external.

### Native wallet card metadata/tone parity — 2026-10-07

- RED: the shared wallet row dropped wallet descriptions and always rendered
  balances in the default ink tone, so the management list diverged from the
  web card contract.
- PASS: `MPWalletRowPolicy` trims descriptions, falls back to the canonical
  wallet type label, and selects the danger tone for negative balances.
  Overview and wallet-management callers pass the real description and sign;
  masking and owner-scoped API behavior are unchanged.
- PASS: focused policy test is 1/1 at
  `/tmp/mypocket-wallet-row-final/Logs/Test/`; the full native unit-only target
  is 117/117 at `/tmp/mypocket-port-wallet-final-unit-only/Logs/Test/` and the
  Release warnings-as-errors build is `/tmp/mypocket-port-wallet-final-release`.
- OPEN: authenticated physical-device visual/pixel approval remains external.

### Native reference segmented-control parity — 2026-10-07

- RED: native reused the solid green primary button for active segments, while
  the Stitch/web contract uses a muted track and white active pill.
- PASS: `MPButtonStyle.segmentedSelected` and
  `MPSegmentedControl.buttonStyle(isSelected:)` centralize the active/inactive
  styles; transaction, report and detail/trend consumers remain base-only.
- PASS: focused test is 1/1 at
  `/tmp/mypocket-segmented-reference/Logs/Test/`; full unit-only target is
  118/118 at `/tmp/mypocket-port-segmented-final-unit/Logs/Test/`; Release
  warnings-as-errors is `/tmp/mypocket-port-segmented-final-release`.
- PASS: authenticated local API editor screenshot was regenerated with the
  correct `SWIFT_ACTIVE_COMPILATION_CONDITIONS='DEBUG LOCAL_API_UI'` build
  setting; result image is
  `/tmp/mypocket-visual-wallet-editor-final-attachments/2DA367F3-AFFC-4C51-9BE5-5FE060546DA2.png`.
- OPEN: physical-device/pixel approval remains external.

### Native transaction-editor category icon parity — 2026-10-07

- PASS: `TransactionsPresentation.glyph` resolves canonical category
  `system_key` values and user-created `icon_key` fallbacks; the selected
  icon is rendered by the shared `MPOptionSelector` in the editor.
- PASS: glyph contract is covered by the 118/118 unit target at
  `/tmp/mypocket-native-final-unit3.xcresult`; Release warnings-as-errors build is
  `/tmp/mypocket-native-final-release2`.
- PASS: API-backed editor UI test ran 1/1 with `DEBUG LOCAL_API_UI`; inspected
  screenshot is
  `/tmp/mypocket-visual-wallet-editor-icon2-attachments/0BFB1ECA-9040-426B-ABE8-1F95A365B1DC.png`.
- OPEN: no provider-specific wallet logo asset contract exists yet, so native
  intentionally keeps the shared wallet symbol; physical-device pixel review
  remains open.

### Native category icon-picker parity — 2026-10-07

- PASS: the category create/edit form composes the shared
  `MPCategoryIconSelector` base and sends the selected canonical `icon_key`
  through the existing category API payload; list rows reuse that glyph.
- PASS: ten-key catalog/fallback contract is covered by the 119/119 native
  unit target at `/tmp/mypocket-native-final-unit3.xcresult`; Release
  warnings-as-errors build is `/tmp/mypocket-native-final-release2`.
- PASS: local API-backed category create/update/delete UI flow selects and
  persists `expense_food` as `icon_key`, 1/1 at
  `/tmp/mypocket-category-picker-ui4.xcresult`.
- OPEN: authenticated category picker visual approval remains external.

### Native notification inbox mark-read proof — 2026-10-08

- RED: the new UI flow first exposed a missing local notification lookup
  helper. After that helper was added, a seeded run showed the already-running
  `127.0.0.1:8080` API uses the default `postgres` database, not
  `mypocket_stage_v1`; no financial data was written. The test was isolated on
  API port `127.0.0.1:8081`, using the disposable stage database and Redis DB
  15.
- PASS: `testAPIBackedNotificationCanBeMarkedRead` loads the synthetic unread
  item, taps its stable `notification-mark-read-<id>` action, and confirms
  `read_at` through a fresh owner-scoped API list. iPhone 16 Simulator, 1/1,
  zero skips/failures; xcresult:
  `/tmp/mypocket-notification-green3.xcresult`.
- PASS: only synthetic notification ID
  `native-ui-notification-inbox-fixture-20261008` was removed; post-test count
  is 0. The temporary API on port 8081 was stopped. No production endpoint or
  database was used.
- PASS: native unit/contract target 127/127, zero skips/failures; xcresult:
  `/tmp/mypocket-notification-unit.xcresult`. Project/base guards and
  `git diff --check` also pass.
- PASS: web design guardrails/docs (130 valid local links), design-contract
  tests 12/12, TypeScript `tsc --noEmit`, and Vite production build (1,796
  modules). Existing chunk-size warning remains. Tests reported an occupied
  Vite websocket port 24678, but all assertions passed.
- No API, schema, auth, or user-visible behavior changed. Public API docs and
  changelog are unchanged by design; the contract intentionally has no public
  notification-create route.

### Configured AI model live probes — 2026-10-08

- PASS: `AI_LIVE_TEST=1 AI_TOOL_CALL_PROBE=1 rtk go test ./internal/infrastructure/ai -run 'TestLiveExtractionWhenExplicitlyEnabled|TestLiveToolCallingCompatibility' -count=1 -v` returned 2 passed. It used the configured `DeepSeek-V4-Flash` endpoint with synthetic transaction text; the tool probe only asks for a proposal and does not execute or persist the tool.
- BLOCKED: live OCR/PDF and image-entry evaluation were not run. `backend/.env.local` has an OCR URL but no `OCR_API_KEY`; private PDF evaluation additionally lacks `S3_ENDPOINT`, bucket and credentials. No production secret was read or substituted.

### Finance Assistant provider-backed stage flow — 2026-10-08

- PASS: the temporary API on `127.0.0.1:8081` used the isolated `mypocket_stage_v1` database and Redis DB 15. Before the synthetic request, the fixture owner had 0 advisor conversations and 0 messages.
- PASS: authenticated `POST /api/v1/ai/advisor/messages` with a synthetic monthly-summary question returned `completed` with non-empty assistant text. A follow-up `GET /api/v1/ai/advisor/conversation/messages` returned both `user` and `assistant` roles. The run stored `prompt_tokens`, `completion_tokens`, `total_tokens` and `provider_calls` in `model_usage`.
- PASS: the synthetic conversation/run and cascaded messages/events/facts were removed; owner conversation and matching run counts are both 0. Temporary API stopped. No transaction write tool was invoked by this test, and no production endpoint/database was used.
- LIMIT: this proves provider-backed send/history and model-usage persistence through the local API; it is not authenticated iOS pixel UAT or production acceptance.

### Live OCR gate revalidation — 2026-10-08

- The earlier `TestLiveQwenImageAndTransactionEval` entry marks the run
  blocked because environment credentials were missing. The canonical ignored
  backend env now has the OCR, private S3 and model variables set; values are
  intentionally not printed or stored in tracked docs.
- PASS: `AI_EVAL_LIVE=1 AI_EVAL_IMAGE=/tmp/mypocket-ai-eval-receipt.png go test ./internal/infrastructure/ai -run '^TestLiveQwenImageAndTransactionEval$' -count=1 -v` in `backend/` returned 1 passed in 6.230s. Values were loaded from the ignored canonical local env and were never printed.
- PASS: private S3 put + signed GET readback, 60,915 bytes, 4 ms; OCR recognized the synthetic 45,000 VND receipt in 1.061s; the configured DeepSeek-V4-Flash model scored 24/24 expected fields across five expense, income, multiple-expense, OCR and transfer-safety cases (100%), p50 880ms / p95 1.141s, OCR+receipt pipeline 2.202s.
- PASS: the test's deferred cleanup deleted its one synthetic private object. No proposal was approved and no ledger/database records were created; stage API/DB and production MyPocket API were not used.
- LIMIT: this validates the provider adapter/evaluation, not iOS image-picker UI, proposal approval/receipt attachment download, production data path or owner device acceptance. Public API docs need no update because there is no contract change.

### Isolated native API fixture routing — 2026-10-08

- RED: `LocalAPIFixture` defaulted to port 8080, whose running local API uses
  the default `postgres` database rather than the disposable stage database.
  This could make otherwise isolated UI fixture helpers write to the wrong
  local database.
- PASS: all fixture helpers now default to `127.0.0.1:8081` and can be
  overridden with `MYPOCKET_TEST_API_BASE_URL`; the app-under-test is built
  with `MP_API_BASE_URL` set to that same disposable API. The seeded
  notification test uses the shared fixture URL rather than a separate
  hard-coded endpoint.
- PASS: `testAPIBackedWalletCreateAndUpdateFromNativeEditor` passed 1/1 on the
  iPhone 16 simulator at `/tmp/mypocket-isolated-wallet-ui.xcresult`. The
  synthetic owner had zero same-name wallets before and after teardown. The
  temporary API used `mypocket_stage_v1` and Redis DB 15, and was stopped after
  the test; the existing API on 8080 was not modified.
- No API/schema or user-visible behavior changed. Public API docs and
  changelog are unchanged by design; this is test-harness isolation only.

### Native transaction amount Stitch parity — 2026-10-08

| Proof | Result |
| --- | --- |
| RED regression | PASS: before the layout correction, `testAddTransactionUsesStitchFormBaseControls` failed on the intended position check (VND center x=292.17 was not left of amount input x=36.00). |
| Focused iPhone 16 editor UI | PASS: 1/1 at `/tmp/mypocket-amount-layout-final-ui/Logs/Test/Test-MyPocketIOS-2026.10.08_02-25-09-+0700.xcresult`; verifies leading VND chip, `1.500.000` display grouping, date/calendar, and calculator entry/commit. Screenshot: `/tmp/mypocket-amount-layout-final-attachments/E8F2528D-A885-4256-8136-A78F5B870A0C.png`. |
| Native unit/contract | PASS: 127/127 at `/tmp/mypocket-amount-layout-unit/Logs/Test/Test-MyPocketIOS-2026.10.08_02-27-11-+0700.xcresult`. |
| Release warnings-as-errors | PASS: `xcodebuild -quiet ... SWIFT_TREAT_WARNINGS_AS_ERRORS=YES build`, derived data `/tmp/mypocket-amount-layout-release`. |
| Project/base/diff guards | PASS: `verify-project.sh`, `check-base-components.sh` and `git diff --check`. |
| Scope | No API, schema, migration, provider or production-data changes. Standard amount-field consumers remain unchanged; full app visual acceptance remains open. |

### Native recurring run-due ledger state — 2026-10-08

- PASS: focused `LOCAL_API_UI` test passed 1/1 on iPhone 16 Simulator at
  `/tmp/mypocket-recurring-state-ui/Logs/Test/Test-MyPocketIOS-2026.10.08_02-43-21-+0700.xcresult`.
- PASS: the assertions verified the generated transaction's wallet, amount and
  recurring note through the fixture owner session and verified that row
  appears in the native ledger. Teardown removed the fixture-wallet
  transactions, schedule and wallet.
- PASS: the temporary API used only `mypocket_stage_v1` and Redis DB 15 on
  `127.0.0.1:8081`, and was stopped after the test. The existing API at 8080
  was not used. No API/schema/public-doc or changelog change is needed.

### Native Finance Assistant wallet confirmation parity — 2026-10-08

- RED: the new state contract did not compile before
  `AssistantWalletConfirmationState` existed.
- PASS: focused state regression verifies single submission during an active
  request, permanent lock after success and retry after failure.
- PASS: full native unit/contract target is 128/128 at
  `/tmp/mypocket-wallet-confirmation-full/Logs/Test/Test-MyPocketIOS-2026.10.08_02-53-46-+0700.xcresult`.
- PASS: Release warnings-as-errors build at
  `/tmp/mypocket-wallet-confirmation-release`; project verification,
  base-component guard/tests and `git diff --check` pass.
- SCOPE: no API, schema, migration or backend behavior changed. The shared
  Assistant screen contract and Unreleased changelog were reconciled.

### Native month-summary live ledger state — 2026-10-08

- PASS: API-backed iPhone 16 Simulator UI test passed 1/1 at
  `/tmp/mypocket-month-summary-green/Logs/Test/Test-MyPocketIOS-2026.10.08_03-02-16-+0700.xcresult`.
- PASS: inspected screenshot
  `/tmp/mypocket-month-summary-attachments/6C29E050-644F-4899-A4EF-69B3E1D7B3BE.png`
  shows the same API-derived `321.000 ₫` in the month expense metric and
  category aggregate.
- PASS: the same fixture owner’s `/months/{YYYY-MM}` expense and transaction
  count increased by exactly the synthetic expense; the native month screen
  rendered the updated expense total.
- PASS: test used `mypocket_stage_v1`, Redis DB 15 and temporary API
  `127.0.0.1:8081`. Post-cleanup SQL found 0 fixture wallets and 0 fixture
  transactions; API was stopped. Existing port 8080 and production were not
  used.
- SCOPE: this adds proof for an existing owner-scoped report read path; no API,
  schema, migration or business rule changed. Web/browser and owner visual UAT
  are separate gates.

### Native credit action form-sheet parity — 2026-10-08

- RED: the API-backed iPhone 16 test failed after tapping `Thêm giao dịch`:
  no shared form-sheet surface or close action existed because native inserted
  the form inline in the statement.
- PASS: `testAPIBackedCreditWalletOpensLedgerActions` now verifies both add-entry
  and payment actions open the shared form sheet, exposes the native dismiss
  control, and cancels without writing. Focused result:
  `/tmp/mypocket-credit-sheet-green/Logs/Test/Test-MyPocketIOS-2026.10.08_03-19-37-+0700.xcresult`;
  screenshot:
  `/tmp/mypocket-credit-sheet-attachments-green/01BF787C-CB37-4C43-BDB1-9C9357671A7B.png`.
- PASS: existing `testAPIBackedCreditEntryAndPaymentRefreshStatementFromNativeUI`
  still passes 1/1 at
  the combined two-case result above, proving entry and payment still write
  and refresh the real statement.
- PASS: native unit/contract target passes 128/128 at
  `/tmp/mypocket-credit-sheet-unit/Logs/Test/Test-MyPocketIOS-2026.10.08_03-20-59-+0700.xcresult`;
  Release warnings-as-errors builds at `/tmp/mypocket-credit-sheet-release`;
  `verify-project.sh`, both base-component guards/tests and `git diff --check`
  pass.
- PASS: test used only `mypocket_stage_v1`, Redis DB 15 and loopback API 8081.
  Post-test SQL reports 0 named fixture wallets and 0 associated transactions;
  API 8081 was stopped. Existing API 8080 and production were not used.
- SCOPE: only native presentation changed. Credit endpoints, payloads, category
  rules, atomic payment pair, persistence, owner scope and public API docs are
  unchanged; human-readable changelog and wallet design docs were reconciled.

### Native wallet selection-row Stitch parity — 2026-10-08

- RED: API-backed `testAPIBackedWalletManagementMatchesGroupedWebContract`
  failed on missing title/balance geometry identifiers and the aggregate
  selected check in `/tmp/mypocket-wallet-rows-red.xcresult`.
- GREEN: the same iPhone 16 Simulator test passed 1/1 in
  `/tmp/mypocket-wallet-rows-final-ui.xcresult`. It verifies the included-only
  aggregate, aggregate selected mark, title-above-balance order and shared
  leading edges for aggregate and wallet rows, while retaining the existing
  grouped CRUD/detail/scope flow.
- Screenshot inspected: `/tmp/mypocket-wallet-rows-final-attachments/4F92374E-3F0E-49FF-A097-675913C3EB34.png`.
- Native unit/contract tests pass 130/130 at
  `/tmp/mypocket-wallet-rows-unit.xcresult`; Release warnings-as-errors build
  passes at `/tmp/mypocket-wallet-rows-release-final`; project, base-component
  and `git diff --check` guards pass. Stage fixture-wallet count after teardown
  is 0; API 8081 is stopped and port 8080 was not used for this test.
- Density follow-up RED/GREEN: the focused API-backed wallet-management test
  failed when `Chi tiết` occupied a separate row at
  `/tmp/mypocket-wallet-detail-row-red.xcresult`, then passed after moving it
  to the same row at `/tmp/mypocket-wallet-detail-row-green.xcresult`; it
  asserts the selection and detail actions' vertical centers differ by less
  than 8pt. Inspected screenshot:
  `/tmp/mypocket-wallet-detail-row-attachments/1DD156EC-692D-4DC9-8CCE-B541B2E74EEF.png`.
- Final post-density verification: native unit/contract suite passes 130/130
  at `/tmp/mypocket-wallet-density-unit.xcresult`; Release
  warnings-as-errors build passes at `/tmp/mypocket-wallet-density-release`;
  project/base-component/diff guards pass. Stage fixture wallet count is 0 and
  temporary API 8081 is stopped; production and the existing API 8080 were not
  used.
- Scope: native shared row presentation and transient selection state only;
  no endpoint/schema/migration, owner scope or balance calculation changed.
  Public API docs and AI-readable API catalog are not affected.

### Native wallet aggregate glyph parity — 2026-10-08

- RED: API-backed `testAPIBackedWalletManagementMatchesGroupedWebContract`
  failed on the absent `wallet-total-glyph` semantic element at
  `/tmp/mypocket-total-glyph-red.xcresult`.
- GREEN: the same iPhone 16 Simulator test passes 1/1 at
  `/tmp/mypocket-total-glyph-green.xcresult`; screenshot inspected:
  `/tmp/mypocket-total-glyph-attachments/3D11766A-DDD7-459D-BC11-C16FA65867C9.png`.
- Aggregate scope now renders a globe in the shared categoryTeal tone; wallet
  balance, included/excluded grouping and selection remain unchanged. No public
  API or machine-readable endpoint documentation changed.

### Native category editor route/layout parity — 2026-10-08

- OPEN: web exposes dedicated `/account/groups/new` and
  `/account/groups/:categoryId/edit` routes, while native embeds its editor
  above the category tree. Web/Stitch show visible wallet applicability rows;
  native currently uses a menu selector. Existing category CRUD API-backed UI
  evidence does not prove route/layout parity.
- Await owner approval before replacing the previously intentional inline
  layout. Proposed change is presentation/navigation only; preserve owner scope,
  system metadata restrictions, category parent validation and request payloads.
- No public API docs changed: no endpoint or contract difference was found.
- Full default UI suite initially exposed one fixture-classification issue: `testWalletEditorReservesShellInset` requires a successful wallet-list API response but was running without `LOCAL_API_UI`. It now explicitly skips outside the disposable API fixture. Focused result: 1 skipped, 0 failed at `/tmp/mypocket-wallet-editor-skip-final.xcresult`. Full UI rerun passes with 20 passed, 31 expected fixture-dependent skips, 0 failed at `/tmp/mypocket-wallet-glyph-default-ui-fixed.xcresult`. Existing API-backed wallet-management/editor tests remain the behavior proof. Native unit/contract suite passes 130/130 at `/tmp/mypocket-wallet-glyph-final-unit2.xcresult`; Release warnings-as-errors passes at `/tmp/mypocket-wallet-glyph-final-release`; project/base-component guards and diff check pass.
## Finance Assistant structured-card density parity — 2026-10-08

- Focused iPhone 16 Simulator contracts pass 2/2 at
  `/tmp/mypocket-card-cap-green-derivedData/Logs/Test/Test-MyPocketIOS-2026.10.08_05-42-38-+0700.xcresult`. They cover normalized aliases, five-row
  transaction caps, eight-row wallet/goal/budget/jar caps, visible-prefix
  preservation and server-total precedence. Missing transaction totals fall
  back to the first five valid rows, matching web behavior.
- Full native unit/contract target passes 131/131 with no skips or failures at
  `/tmp/mypocket-card-cap-full-unit.xcresult`.
- Release warnings-as-errors build, `mobile-ios/scripts/verify-project.sh`,
  both base-component guards/tests and `git diff --check` pass. No API,
  database, provider, or public endpoint behavior changed.

## Native category list/tree filters — 2026-10-08

- RED: API-backed `testAPIBackedCategoryCreateUpdateDeleteFromNativeUI` exposed
  a test visibility assumption after clearing search: the newly created row was
  outside the `LazyVStack` viewport and therefore absent from the current
  accessibility snapshot. Search had already found the row; this was not a
  filter/API defect. The test now keeps the row filtered through edit/delete.
- PASS: the API-backed create/update/delete flow, category-kind segment,
  category search, wallet-scope control and empty result after deletion pass
  together on iPhone 16 Simulator, 1/1, zero skips/failures at
  `/tmp/mypocket-category-tree-api-ui2.xcresult`.
- Scope: local presentation filtering over owner-scoped `/categories` and
  `/wallets` data; `wallet_ids` is already returned by the API. No backend,
  migration, public API or AI-agent endpoint catalog change was needed. Test
  used the isolated local `mypocket_stage_v1` database and Redis DB 15, not
  production.

## Native transaction/budget category tree picker — 2026-10-08

- RED: two selection-policy tests first failed to compile because the
  selectable-ID/parent-context policy did not exist,
  `/tmp/mypocket-category-picker-red.xcresult`.
- PASS: selection policy tests now cover kind eligibility, selected IDs,
  Vietnamese diacritic-insensitive child search and retained parent context,
  `/tmp/mypocket-category-picker-green-model.xcresult`.
- PASS: API-backed iPhone 16 flows create a transaction and a budget from the
  shared searchable tree, then verify the stored `category_id` via fresh
  owner-scoped API reads; 2/2 at
  `/tmp/mypocket-category-picker-api-ui.xcresult`.
- PASS: recurring create selects a category from the shared tree, verifies the
  schedule's persisted category ID, runs the due schedule and verifies the
  generated ledger transaction retains that category, 1/1 at
  `/tmp/mypocket-category-picker-recurring-ui.xcresult`.
- No backend code, migration, public endpoint contract or AI-readable API
  catalog changed: current endpoints already expose hierarchy and accept the
  selected IDs.
- PASS: full native unit/contract target is 137/137, zero skips/failures at
  `/tmp/mypocket-category-picker-final-unit.xcresult`; Release
  warnings-as-errors passes at `/tmp/mypocket-category-picker-final-release`.
- PASS: native base-component guard/tests, project verification and
  `git diff --check`.

### Category-picker subtitle and clear-action parity follow-up — 2026-10-08

| Surface | Verification | Result |
| --- | --- | --- |
| API-backed accessibility regression | `testAPIBackedBudgetCreateUpdateDeleteFromNativeUI` on iPhone 16 Simulator | RED before implementation (`Ăn uống` lacked `Nhóm cá nhân`); PASS 1/1 after fix, exact category label and persisted category selection checked at `/tmp/mypocket-picker-parity-budget-ui-20261008.xcresult` |
| Visual review | Named screenshot attached to the same UI test | PASS; inspected `/tmp/mypocket-picker-parity-budget-ui-attachments-20261008/E29DE286-CD42-42B0-BCED-F9CE25C52337.png`; root cards, category subtitles and leading ghost clear action visible |
| Cross-flow regressions | Recurring create/update/delete/run-due and wallet-detail transaction create/update | PASS 2/2 at `/tmp/mypocket-picker-parity-crossflow-ui-20261008.xcresult` |
| Native unit/contract | `MyPocketIOSTests` on iPhone 16 Simulator | PASS 144/144, zero skips/failures at `/tmp/mypocket-picker-parity-unit-final-20261008.xcresult` |
| Build and guards | Release build with `SWIFT_TREAT_WARNINGS_AS_ERRORS=YES`; project and both base-component scripts; `git diff --check` | PASS |
| Scope | Local API on 8081 with disposable local database/Redis fixture | Production not accessed; no backend/API/schema/migration change |

The implementation keeps selectable eligibility separate from hierarchy/search
context, so an eligible parent retained for a matching child stays selectable;
an ineligible parent remains informational context only. Budget clear uses the
shared `.ghost` button and leading alignment. Physical-device and owner visual
approval remain open for the overall 100% port.

## Category archive backend + web + native contract — 2026-10-08

| Surface | Verification | Result |
| --- | --- | --- |
| Backend | `TEST_DATABASE_URL=<isolated local PostgreSQL 15 DSN> rtk go test ./...` | PASS; 265 tests across 16 packages |
| Web selection regression | `rtk npm run test:transactions` | PASS; 8 tests |
| Web production build | `rtk npm run build` | PASS; design guard, typecheck and Vite build (existing large-chunk advisory remains) |
| Native model/API contracts | iPhone 16 Simulator: inactive visibility, parent context, transaction edit retention, authenticated status PATCH; `/Users/dungxbuif/Library/Developer/Xcode/DerivedData/MyPocketIOS-cgilrnvqcdzmbzgtuefjqpthbajm/Logs/Test/Test-MyPocketIOS-2026.10.08_12-29-39-+0700.xcresult` | PASS; 4/4 focused tests |
| Native test build | `xcodebuild build-for-testing` | PASS |
| Repository diff | `rtk git diff --check` | PASS |

Backend contract now persists `is_active`, serves owner-scoped
`PATCH /api/v1/categories/{id}/status`, and rejects new inactive-category
assignments across transaction, credit, budget, recurring and AI-entry paths.
Existing financial references remain unchanged. The web and native management
screens support archive/restore and inactive visibility; new selectors exclude
inactive categories while preserving allowed unchanged historical references.
Swagger and the human API reference are updated.

An isolated PostgreSQL 15 cluster was initialized under `/tmp` on port 55432,
and the complete migration chain was applied there. It contains 75 seed
categories; all migrated categories are active and `is_active` is `NOT NULL`
with default `true`. The focused repository integration tests pass 2/2,
covering the migration default, owner/system isolation, idempotent archive and
restore, and preservation of existing transaction references. A
handler-to-PostgreSQL integration test passes 1/1: it archives through HTTP,
confirms archived-category visibility, rejects a new transaction with
`category_inactive`, preserves transaction/budget/recurring category
references, runs the existing recurring schedule while the category is
archived, restores through HTTP, and then accepts a new transaction. The full
backend suite passes 265 tests across 16 packages with that isolated database
DSN.

Additional current-source verification on 2026-10-08: the existing local stage
DB was at schema version 29 and lacked `categories.is_active`; applying the
existing additive migration `000030` advanced it to version 30 without reset
or data deletion. `TEST_DATABASE_URL=<local-stage-DSN> rtk go test -race ./... -count=1`
then passed 265 tests across 16 packages. The same command also passed against
a fresh disposable PostgreSQL database migrated to version 30; that exact
scratch database was dropped afterward. This stage migration is local-only.

Not established by this run: the full native UI suite or visual/device UAT.
No production database, API, deployment or remote service was used; production
was not accessed or changed. The temporary scratch cluster/database has been
removed.

## Feedback bubble shell/scroll overlap — architecture regression — 2026-10-08

- Fresh iPhone 16 Simulator RED:
  `testFeedbackBubbleDoesNotCoverMoreRowsOrCategoryEditorActions` fails
  because a visible More destination row and the category editor cancel action
  intersect the bubble. Result:
  `/Users/dungxbuif/Library/Developer/Xcode/DerivedData/MyPocketIOS-cgilrnvqcdzmbzgtuefjqpthbajm/Logs/Test/Test-MyPocketIOS-2026.10.08_12-33-04-+0700.xcresult`.
- The test now also inspects More rows after scrolling. Moving the bubble into
  a shell safe-area lane did not solve the overlap before or after scrolling;
  that implementation attempt was removed. Its failed run is recorded at
  `/Users/dungxbuif/Library/Developer/Xcode/DerivedData/MyPocketIOS-cgilrnvqcdzmbzgtuefjqpthbajm/Logs/Test/Test-MyPocketIOS-2026.10.08_12-36-09-+0700.xcresult`.
- Status remains **unfixed**. More's nested `NavigationStack` and descendant
  scroll views need an architecture-level ownership change; do not add more
  route-local or global inset padding without a new hypothesis. This is a
  native layout issue, not an API/backend gap.
- Current-source revalidation on iPhone 16 Simulator:
  `testFeedbackBubbleDoesNotCoverMoreRowsOrCategoryEditorActions` still fails
  3 assertions at `/tmp/mypocket-feedback-collision-revalidate-20261008.xcresult`:
  the feedback action intersects the More wallets row, the More Assistant row
  after scrolling, and the category-editor cancel action. No shell architecture
  change was made pending owner review of the dock design.

## Category archive response compatibility — 2026-10-08

| Surface | Verification | Result |
| --- | --- | --- |
| Regression proof | `TEST_DATABASE_URL=<isolated PostgreSQL 15 DSN> rtk go test ./internal/controller/http -run '^TestCategoryArchiveHTTPFlowWithPostgresAndTransactionAPI$' -count=1` | RED before fix (`wallet_ids:null`), PASS after fix; archive and restore both return a non-null array |
| Local API probe | Fixture exchange → create category → `PATCH /categories/{id}/status` against disposable PostgreSQL | Confirmed old response had `wallet_ids:null`; fixture category removed after probe |
| Native API-backed UI | iPhone 16 Simulator, `testAPIBackedCategoryCreateUpdateArchiveRestoreDeleteFromNativeUI` against disposable local API/database | Archive/restore, default inactive filtering, archived-row rendering and API state assertions pass; test remains RED on the two independently tracked feedback-bubble overlaps (More row and category row). Result: `/tmp/mypocket-category-archive-ui-verified.xcresult` |
| Native API contract | `testCategoryArchiveUsesAuthenticatedStatusPatchContract` | PASS 1/1 at `/tmp/mypocket-category-api-contract.xcresult` |
| Backend regression suite | `TEST_DATABASE_URL=<isolated PostgreSQL 15 DSN> rtk go test ./...` | PASS across all backend packages |
| Native quality gates | `check-base-components.sh`, `check-base-components.test.sh`, `verify-project.sh`, Release Simulator build with `SWIFT_TREAT_WARNINGS_AS_ERRORS=YES` | PASS |
| Whitespace | `rtk git diff --check` | PASS |
| Scope | Local disposable PostgreSQL on port 55433; local Redis/API only | No production database, API or migration touched |

Root cause: `CategoryPostgresRepository.SetActive` updated and reloaded the
database row but did not hydrate the derived `WalletIDs` field. Go serialized
the nil slice as `null`, which could not decode into the native client's
non-optional `[String]`; the server state had changed, but the client caught
the response-decoding error and remained in the editor. The repository now
returns an owner-scoped wallet-link array, initialized to `[]` when empty.

## Populated goal-wallet ledger CTA and scroll reachability — 2026-10-08

- RED with a real local API fixture: `testAPIBackedGoalWalletScopedLedgerRendersSavingsPanel`
  found the CTA frame at y=780.67, below the shell navigation (y=767) and
  raised shell add affordance (y=730). Screenshot showed the CTA anchored at
  the end of the populated embedded ledger. A cushion-only experiment (104pt
  to 160pt) left the frame unchanged.
- Root cause: the Transactions route wrapped the goal detail's own `ScrollView`
  in the screen's outer `ScrollView`, so the goal CTA was not viewport-anchored.
  The goal detail now renders as the screen root; other transaction scopes keep
  the existing outer scroller.
- GREEN: the scoped goal test and explicit goal-detail test pass 2/2 at
  `/tmp/mypocket-goal-ledger-final-20261008.xcresult`. The scoped test now
  scrolls the persisted row fully above the fixed CTA and checks it remains
  above the shell add affordance. Latest reachability proof passes 1/1 at
  `/tmp/mypocket-goal-ledger-reachability-final-20261008.xcresult`; inspected
  populated capture:
  `/tmp/mypocket-goal-ledger-reachability-attachments/D2901CC9-C45F-4EEC-8CED-92C14702B4F5.png`.
- Fixture wallet and both uniquely noted transactions verified absent after
  teardown. Local API 8081, local stage DB and Redis DB 15 only; production was
  not used. No API/backend/schema/migration changes.
- Native unit/contract target passes 142/142 at
  `/tmp/mypocket-goal-ledger-unit-20261008.xcresult`; focused API-backed UI
  checks for populated goal detail, scoped goal ledger, ordinary week ledger
  and wallet selection pass 5/5 across
  `/tmp/mypocket-goal-ledger-final-20261008.xcresult`,
  `/tmp/mypocket-goal-ledger-reachability-final-20261008.xcresult` and
  `/tmp/mypocket-transaction-nongoal-regression-20261008.xcresult`. Release
  warnings-as-errors, `verify-project.sh`, both base-component scripts and
  `git diff --check` pass. The full API-backed UI suite and device/owner visual
  sign-off remain open for the overall port.

## Reference wave integration gate — 2026-10-16

The native source now composes the populated ledger, savings report and budget
report with the shared reference bases. A focused signed simulator run passed
7/7 (`ReferenceDisplayTests` + `TransactionsReferenceTests`) and the Release
simulator build passed with `SWIFT_TREAT_WARNINGS_AS_ERRORS=YES` at
`/tmp/mypocket-reference-wave1-release-20261016`.

The full unit target passed after making the shared JSON encoder use sorted
keys. This stabilizes identical retry bodies for versioned budget-series
mutations and fixes the deterministic idempotency assertion; it does not alter
the JSON schema or business payload. Artifact/log:
`/tmp/mypocket-reference-wave1-unit2-20261016` and
`/tmp/mypocket-reference-wave1-unit2.log`.

After Overview/Wallets integration, the complete native unit target passes
329/329 with zero failures, skips or runtime warnings at
`/tmp/mypocket-reference-wave2-unit-20261016`.

An unscoped `xcodebuild test` was stopped after it remained in simulator
launch orchestration for more than three minutes. This is an environmental
runner issue, not a passing full UI-suite claim. The existing API-backed
populated savings/budget evidence remains the authoritative live fixture proof;
new reference UI screens still require a separately configured fixture run for
screenshots.

## Reference wave 2 integration gate — 2026-10-16

Overview and Wallets now consume the shared compact metric grid and the shared
wallet-selection title contract. Focused signed simulator tests for reference
display, ledger, overview and wallets passed 12/12 with no failures, skips or
runtime warnings. Artifact:
`/tmp/mypocket-reference-wave2-focused-20261016/Logs/Test/`.

The Release simulator build passed with warnings treated as errors at
`/tmp/mypocket-reference-wave2-release-20261016`. No API, database, provider,
production or physical-device state changed. The sample reference app remains
separate from MyPocket's test fixture; its remaining phone-only branches are
listed in the reference-survey ledger.

## Reference wave 3 integration gate — 2026-10-16

Categories and Recurring now use compact reference surfaces while preserving
the existing category tree, CRUD, recurrence editor and overflow actions.
Focused signed simulator tests for all reference waves pass 14/14 with no
failures, skips or runtime warnings at
`/tmp/mypocket-reference-wave3-focused-20261016/Logs/Test/`.

The complete native unit target passes 331/331 at
`/tmp/mypocket-reference-wave3-unit-20261016`, and the Release simulator
build passes with warnings treated as errors at
`/tmp/mypocket-reference-wave3-release-20261016`.

## Reference wave 4 integration gate — 2026-10-16

Credit and Notifications now use shared compact metric/row surfaces. Credit
summary masking, statement counts, owner-scoped payment state, unread/read
grouping and mark-read actions remain unchanged. The agent-owned signed
focused run passes 2/2 at
`/tmp/mypocket-credit-notifications-20261016/Logs/Test/Test-MyPocketIOS-2026.10.16_22-01-17-+0700.xcresult`.

The integrated Release simulator build passes with warnings treated as errors
at `/tmp/mypocket-reference-wave4-release-final-20261016`. A second local test
invocation was stopped after Xcode remained in `simctl diagnose` cleanup for
over three minutes; this is an environment runner hang, not a product test
failure. No API, data or production state changed.

## Reference wave 5 integration gate — 2026-10-16

More and Feedback now use grouped reference surfaces. Feedback keeps the
unresolved/resolved tabs, explicit tick/un-tick flow, changelog status and
owner-scoped API; submission uses the shared editor scaffold and keyboard-safe
footer. The focused signed test set passes 4/4 with no failures or skips in
the agent's recorded run.

Base-component and project guards pass after replacing the new notification
icon with `MPIcon`; the final Release simulator build with warnings treated as
errors passes at `/tmp/mypocket-reference-wave5-release-20261016`.
No API, database, provider, production or device state changed.

## Reference wave 6 integration gate — 2026-10-16

Bills/debt-adjacent credit, Travel Mode and Portfolio now use the shared
reference-density surfaces. Travel rows retain stable toggle/edit/delete
accessibility identifiers and existing event semantics. Portfolio keeps the
owner-scoped asset/trade actions, masking and history while using the shared
metric grid and compact row composition. Recurring's introductory surface now
uses the same reference radius token.

The focused signed simulator run passed 3/3 with no failures, skips or runtime
warnings at `/tmp/mypocket-reference-wave6-focused-20261016/Logs/Test/`.
The complete native unit target passes 340/340 with no failures, skips or
runtime warnings at `/tmp/mypocket-reference-wave6-unit-20261016/Logs/Test/`.
The Release simulator build passed with
`SWIFT_TREAT_WARNINGS_AS_ERRORS=YES` at
`/tmp/mypocket-reference-wave6-release-20261016`.
Base-component and project guards plus `git diff --check` pass. No API,
database, provider, production or physical-device state changed.

## Reference wave 7 integration gate — 2026-10-16

Jars and Months now use grouped reference surfaces and the shared adaptive
metric grid. Monthly allocation rows, cumulative jar report, category summary,
month notes, timezone period text, CRUD and masking behavior remain intact.
The focused signed simulator test passes 2/2 with no failures, skips or
runtime warnings at `/tmp/mypocket-jars-months-focused-20261016/Logs/Test/`.
The complete native unit target passes 342/342 with no failures, skips or
runtime warnings at `/tmp/mypocket-jars-months-unit-20261016/Logs/Test/`.
The Release simulator build passes with warnings treated as errors at
`/tmp/mypocket-jars-months-release-verify-20261016`. Base-component and
project guards plus `git diff --check` pass. No API, database, provider,
production or device state changed.

## Reference wave 8 integration gate — 2026-10-16

Account settings and Wallet Detail now use grouped reference surfaces. API
key lifecycle, timezone, Face ID, logout, wallet ledger, credit statement,
queue/editor routes, masking and accessibility identifiers are unchanged.
`AccountWalletReferenceTests` passes 2/2 with no failures, skips or runtime
warnings at `/tmp/mypocket-account-wallet-focused-20261016/Logs/Test/`.
The complete native unit target passes 344/344 with no failures, skips or
runtime warnings at `/tmp/mypocket-account-wallet-unit-20261016/Logs/Test/`.
Base-component and project guards plus `git diff --check` pass. The Release
simulator build passes with warnings treated as errors at
`/tmp/mypocket-account-wallet-release-20261016`.
No API, database, provider, production or physical-device state changed.

## Reference wave 9 integration gate — 2026-10-16

Reports, savings drilldown and transaction detail now use the shared reference
surface family. Trend/breakdown/activity sections and transaction amount/detail
cards are compact without changing owner scope, report inclusion, travel-link,
transfer or mutation semantics. `TransactionsReferenceTests` passes 3/3 at
`/tmp/mypocket-reports-transaction-focused-20261016/Logs/Test/`; the complete
native unit target passes 344/344 at
`/tmp/mypocket-reports-transaction-unit-20261016/Logs/Test/`.
Base-component/project guards and `git diff --check` pass. Release build with
warnings-as-errors passes at `/tmp/mypocket-reports-transaction-release-20261016`.
No API, database, provider, production or physical-device state changed.

## Native AI-entry uncertain request recovery — 2026-10-08

| Check | Command / evidence | Result |
| --- | --- | --- |
| RED policy regression | Added error-classification and submit-eligibility assertions before the recovery policy existed | EXPECTED FAIL: compiler reports missing `AIEntryRequestRecoveryPolicy` |
| Native unit suite | `xcodebuild test ... -only-testing:MyPocketIOSTests` on isolated iPhone 16 Simulator `F03CAAC4-A5FF-49E7-AAF2-B6DCE8561BB9` | PASS: 146/146 at `/tmp/mypocket-ai-entry-recovery-unit-full-20261008.xcresult` |
| Recovery UI | `testAIEntryUncertainRequestRequiresGETResolutionBeforeResubmission`, `DEBUG LOCAL_API_UI`, deterministic GET-404 fixture; isolated simulator, no backend | PASS: 1/1 at `/tmp/mypocket-ai-entry-recovery-final-20261008.xcresult`; uncertain state blocks analysis/input, GET 404 preserves draft and re-enables explicit submit |

No production simulator interaction, account data, AI provider, backend, API
contract, database, schema or migration was involved. This defect fix does not
close the remaining port audit or owner UAT.

## Current-source native build and project guard refresh — 2026-10-08

- PASS: generic iOS Simulator Debug and Release builds from the current
  worktree; Release used `SWIFT_TREAT_WARNINGS_AS_ERRORS=YES`. Build products
  were isolated under `/tmp/mypocket-port100-validation-20261008` and were not
  installed over the owner's simulator session.
- PASS: `mobile-ios/scripts/verify-project.sh`,
  `mobile-ios/scripts/check-base-components.sh`,
  `mobile-ios/scripts/check-base-components.test.sh`, and `git diff --check`.
- Runtime note: the already-booted iPhone 16 simulator accepted
  `simctl launch com.mypocket.ios` and rendered the login screen in a
  `simctl` screenshot. Although the standalone `Simulator.app` bundle is
  absent, opening Device Hub shows the live `MyPocket-iPhone-16` screen and
  makes the login UI available for owner testing. No account credentials were
  entered.
- This is a build/guard refresh only. It does not resolve feedback-dock
  overlap, authenticated visual UAT, or complete the native port.
