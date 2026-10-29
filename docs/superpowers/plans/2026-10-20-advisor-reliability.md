# Finance Assistant Reliability Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking. Parallel execution is optional only when the owner requests it; the dependency waves below do not dispatch agents by themselves.

**Goal:** Khép kín contract đọc–proposal–confirmation và làm chat tài chính phục hồi được, có nguồn số liệu, theo dõi ổn định trên iOS.

**Architecture:** Giữ Go orchestration và PostgreSQL làm nguồn trạng thái. API executor có lease/heartbeat, worker reaper kết thúc run mất executor; completion là một transaction. Tools trả đủ catalog/version; typed evidence và server templates cung cấp financial narrative, iOS quan sát run bằng GET và giữ context references có giới hạn.

**Tech Stack:** Go 1.26/Gin/GORM/PostgreSQL, OpenAI-compatible provider hiện có, Redis auth/audit, SwiftUI/URLSession/XCTest; không thêm agent framework/vector DB.

**Spec:** [2026-10-20-advisor-reliability-design.md](../specs/2026-10-20-advisor-reliability-design.md).

**Status:** implementation_candidate; các slice lõi đã có trong local candidate
và đã qua deterministic backend/native checks. Integrated DB/provider/device
gates còn mở; production rollout chưa được thực hiện.

## Execution reconciliation — 2026-10-20

- Task 1: catalog/version/snapshot code and focused tests are committed in
  `e434241c`.
- Task 2: runtime lease repository, migration 000041 and atomic completion are
  committed in `89a9b799`.
- Task 3: policy bounds, independent worker reaper and feature-flagged service
  claim/heartbeat/atomic completion are implemented locally; full race and
  kill-executor acceptance remains open.
- Task 4: iOS observer/metadata and Xcode target registration are committed in
  `90088f0c` and `fe56a5de`; focused simulator tests pass.
- Task 5: bounded UTF-8 context and persisted completion state are implemented
  locally for both the fenced and legacy service paths; repository-level
  summary inspection remains open.
- Task 6: evidence indexing, typed facts for supported finance cards, optional
  `finish_answer`, one bounded repair and card-only fallback are implemented
  locally; live-model repair metrics remain open.
- Task 7: the 31-case matrix is now loaded by a deterministic contract test and
  backend/native suites are green locally. Disposable DB/race artifact bundle,
  live provider and device gates are still required before a release claim.

Review correction: loading 31 cases validates the dataset only. It does not
execute those prompts or measure intent accuracy. The schema test is now named
`TestAdvisorReliabilityFixtureSchema` and checks declared tools against the
production registry. Production-flow regressions are tracked separately in
[review closure](../../work/tickets/AI-ADVISOR-REVIEW-FIX-20261020.md).

### Acceptance snapshot

| Gate | Evidence | Status |
| --- | --- | --- |
| Compile/unit/integration package suite | `go test ./... -count=1` | PASS |
| Static analysis | `go vet ./...`, `git diff --check` | PASS |
| Backend race subset | usecase/repository/worker race suite | PASS |
| Native observer/API lifecycle | simulator focused tests | PASS |
| Base-component/project guards | `check-base-components.sh`, `verify-project.sh` | PASS |
| Reliability dataset schema | 31 cases, metadata/ID/tool-registry check only | PASS (not model acceptance) |
| Disposable PostgreSQL migration + lease/completion race | migration 000041, repository tests with `TEST_DATABASE_URL`, `go test -race` | PASS |
| Kill-executor/reaper recovery | expired claimed lease integration test; terminal `executor_lost`, no replay path | PASS (repository simulation) |
| Live provider matrix | configured model, 30 x 3 | OPEN |
| Physical iPhone/device UAT | isolated UAT owner | OPEN |
| Production rollout | explicit owner approval required | NOT STARTED |

## Global Constraints

- Principal.OwnerID luôn server-supplied; JWT/API key cùng quyền owner, không scope/quota mới.
- Một conversation/user; idle TTL 30 phút, accepted user message mới gia hạn.
- Mọi AI write là persisted proposal + explicit confirmation + canonical CAS/idempotency.
- Không ngân hàng/external payment, không thay AI entry/OCR, không maintain archived web.
- Giữ 6 provider calls/8 financial tools; deadline 240s, provider timeout 30s.
- Lease 45s, heartbeat 10s, reaper 15s/batch 100; DB heartbeat 2s, terminal/reaper 5s.
- Lock order conversation -> run; không auto replay provider sau crash/timeout.
- History context 12 messages/24 KiB; refs 4 KiB; tool transcript 96 KiB UTF-8.
- Không secret/raw prompt/financial payload trong logs hoặc artifacts.
- Không sửa migration đã deploy; chọn migration number tiếp theo ở Task 2.
- Full spec là authority cho các giới hạn, copy, compatibility và release gates.

## Review Focus

- UUID hợp lệ nhưng foreign hoặc system category: tools/confirmation không lộ dữ liệu hoặc cho sửa.
- DB commit thành công nhưng mất response: GET/retry receipt không thêm message/action hoặc gọi provider lại.
- Reaper/reset/cancel tranh completion ở đúng boundary: chỉ một terminal outcome.
- iOS background/offline/lost POST response: giữ owner/run/request identity, GET recovery rồi mới cho gửi lại.
- Model bịa số bằng chữ, dùng fact cũ hoặc trả sai protocol: không publish financial narrative chưa được render từ evidence.

## File map và ownership

| Slice | Existing files | New files đề xuất |
| --- | --- | --- |
| Resource reads | entity/finance_query.go, usecase/ai_advisor_tools.go, infrastructure/repository/finance_query_postgres.go | usecase/ai_advisor_resources.go; các test tương ứng |
| Run lifecycle | entity/ai_advisor.go, repository/ai_advisor.go, usecase/ai_advisor.go, repository/ai_advisor_postgres.go | entity/advisor_runtime.go, usecase/advisor_runtime.go, repository/advisor_runtime_postgres.go |
| Context | usecase/ai_advisor_context.go, conversation.summary | usecase/advisor_context_state.go và tests |
| Evidence/final | usecase/ai_advisor_provider.go, ai_advisor_parts.go, ai/advisor_client.go | entity/advisor_evidence.go, usecase/advisor_evidence.go, advisor_grounding.go và tests |
| Worker | cmd/worker/main.go | cmd/worker/advisor_reaper.go và tests |
| Native observer | AssistantViews.swift, AIModels.swift, APIClient.swift | Features/Assistant/AdvisorRunObserver.swift; MyPocketIOSTests/AdvisorRunObserverTests.swift |
| Evaluation | evals/finance-assistant/cases.json, ai/advisor_quality_eval_test.go | evals/finance-assistant/reliability-cases.json |

All backend paths above relative to `backend/internal/` unless `cmd/` stated.
Không tái cấu trúc các feature ngoài scope để làm sạch code. Reuse canonical
mutation handlers, finance reader và shared native bases hiện có.

## Task 1: Catalog danh mục, version và snapshot validation (AC1, AC2, AC9)

**Files:** Modify `backend/internal/entity/finance_query.go`,
`backend/internal/infrastructure/repository/finance_query_postgres.go`,
`backend/internal/usecase/ai_advisor_tools.go`; create
`backend/internal/usecase/ai_advisor_resources.go` và matching `_test.go`;
extend repository/tools/provider/HTTP tests hiện có.

**Interfaces:** Add `CategoryCatalogFilter` to entity with fields tương ứng
tool args trong spec; add `CategoryFilter *CategoryCatalogFilter` to
`NormalizedQuery`. Keep `FinanceQueryService.Execute` unchanged. Add tool
names `list_categories`, `get_category`; transaction view gains `version`.
`AdvisorResourceSnapshots.Record(result entity.FinanceResult)` and
`ValidateProposal(result entity.FinanceResult) error` track only this run's
validated resource IDs/versions; create references also require owned catalog.

- [ ] Viết RED tests `TestAdvisorTransactionToolsExposePersistedVersion`,
  `TestAdvisorCategoryCatalogScopesPaginationAndMutability`,
  `TestAdvisorProposalRequiresReadSnapshot`. Assertions: stored version=7
  returns 7; proposal 6/unknown ID fails; same-name categories remain choices;
  foreign wallet/owner cannot enumerate; system readable but not editable.
- [ ] Run `cd backend && rtk go test ./internal/usecase ./internal/infrastructure/repository ./internal/controller/http -run 'Advisor.*(Version|Categor|Snapshot)|Finance.*(Version|Categor)' -count=1`.
  Use migrated disposable `TEST_DATABASE_URL`; missing fixture is a skipped
  integration gate, not RED proof. Confirm failures are the missing contract.
- [ ] Implement catalog pagination with stable cursor bound owner+filters,
  limits 50/100 and archived semantics. Include snapshot validation before a
  proposal enters results, and typed repairable errors in orchestrator.
- [ ] Run same tests GREEN; test real HTTP confirmation stale version=409 and
  unchanged ledger, not only mocked tool schema. Update tool catalog, handbook
  versions, capabilities list and public Markdown examples together.
- [ ] Commit only this slice with `feat: complete advisor resource snapshots`.

## Task 2: Fenced runtime repository and atomic completion (AC3–5, AC9–10)

**Files:** Modify `backend/internal/entity/ai_advisor.go`,
`backend/internal/repository/ai_advisor.go`; create
`backend/internal/entity/advisor_runtime.go`,
`backend/internal/infrastructure/repository/advisor_runtime_postgres.go`;
extend `ai_advisor_postgres_test.go`; create next numbered forward migration
plus disposable-only down migration and migration tests.

**Interfaces:** Add entity types `AdvisorRunLease{OwnerID,RunID,Generation,
Token,DeadlineAt}`, `AdvisorCompletion{Parts,Usage,FactBundle,ContextState}`,
`AdvisorCompletionResult{Message,Run,Replayed}`. Types use existing entity
types where possible and JSON-safe bounded ContextState. Add repository:

```go
type AdvisorRuntimeRepository interface {
    ClaimRun(context.Context, string, string, string) (entity.AdvisorRunLease, error)
    HeartbeatRun(context.Context, entity.AdvisorRunLease) error
    CompleteRun(context.Context, entity.AdvisorRunLease, entity.AdvisorCompletion) (entity.AdvisorCompletionResult, error)
    FinishRun(context.Context, entity.AdvisorRunLease, string, string, map[string]any) error
    ReapExpiredRuns(context.Context, int) (int, error)
}
```

DB time is authoritative; tests control stored timestamps and transactions,
not application sleeps. `FinishRun` is executor-fenced failed/interrupted;
user cancellation keeps existing owner-authenticated API. Completion events
use existing `run.completed` payload shape; no raw provider data.

- [ ] Write RED PostgreSQL tests `TestAdvisorLeaseSingleClaim`,
  `TestAdvisorCompletionAtomicRollback`, `TestAdvisorCompletionAckLostReplay`,
  `TestAdvisorCompletionRejectsCancelledExpiredOrForeignLease`,
  `TestAdvisorReaperMatchesConversationRunLockOrder`. Inject failure after
  message/event/usage writes; assert none persist. Two connections race
  cancel/reset/reaper against completion; assert one winning outcome.
- [ ] Run `cd backend && rtk go test -race ./internal/infrastructure/repository -run 'Advisor.*(Lease|Completion|Reaper)' -count=1` with real disposable DB; confirm RED.
- [ ] Add additive deadline/index/unique-assistant migration. Preflight duplicate
  rows is read-only and must fail before index DDL if duplicates exist.
  Implement runtime methods, StartRun owner-local stale check, queued terminal
  transitions and idempotent completion receipt lookup.
- [ ] Test fresh migration and upgrade from current schema, preserving fixture
  owners/wallet balances/messages. Assert existing active old rows become
  terminal through reaper without provider replay. Run repository race GREEN.
- [ ] Document proposed ADR for lease/atomic completion and ERD/API additions;
  commit `feat: fence advisor runs and commit answers atomically`.

## Task 3: Executor deadlines, reaper and error diagnostics (AC3–5, AC10)

**Files:** Modify `backend/internal/usecase/ai_advisor.go`,
`ai_advisor_provider.go`, `backend/cmd/api/main.go`, `backend/cmd/worker/main.go`,
`backend/internal/infrastructure/ai/advisor_client.go`, config/config.go;
create `usecase/advisor_runtime.go`, `cmd/worker/advisor_reaper.go` and tests.

**Interfaces:** `AdvisorRuntimePolicy` in `entity/advisor_runtime.go` owns
240s/45s/10s constants. `AdvisorService` requires `AdvisorRuntimeRepository`
when leases flag enabled, no unsafe optional fallback. `StartAdvisorReaper(ctx,
store repository.AdvisorRuntimeRepository) func()` starts independent timer;
returned function stops loop. Config booleans named exactly as spec.

- [ ] Write RED fake-clock tests `TestAdvisorExecutorDeadlineCancelsTools`,
  `TestAdvisorHeartbeatFailureStopsCompletion`,
  `TestAdvisorReaperRunsWhileRecurringPassBlocked`,
  `TestAdvisorUsageFailureCannotRetroactivelyFailAnswer`,
  `TestAdvisorProviderHTTPFailuresHaveSanitizedClasses`.
- [ ] Run `cd backend && rtk go test ./internal/usecase ./internal/infrastructure/ai ./cmd/worker -run 'Advisor.*(Deadline|Heartbeat|Reaper|Usage|HTTPFailures)' -count=1`; confirm RED.
- [ ] Wire claim/heartbeat/cancellation with bounded contexts, route both sync
  and async through atomic CompleteRun, persist known usage on failures.
  Keep inference transport no-retry. Reaper independent of business-job pass;
  startup/shutdown cancel goroutines cleanly. Old flag-off path stays explicit.
- [ ] Run GREEN plus PostgreSQL kill-executor simulation: kill first executor
  after claim, drive clock past lease, reap, allow next run, verify provider
  call count did not rise from recovery. Verify logs contain IDs/error classes
  but no fixture secrets/prompts or notes.
- [ ] Update config examples/runbook/ADR; commit `fix: recover interrupted advisor runs`.

## Task 4: Native run observer and metadata (AC6, AC9)

**Files:** Modify `backend/internal/controller/http/ai_advisor_handler.go`,
repository history metadata, `mobile-ios/MyPocketIOS/Core/Models/AIModels.swift`,
`mobile-ios/MyPocketIOS/Core/API/APIClient.swift`,
`Features/Assistant/AssistantViews.swift`; create `AdvisorRunObserver.swift`
and `AdvisorRunObserverTests.swift`; extend APIClientTests/LaunchTests.

**Interfaces:** Add optional run metadata `server_time`, `deadline_at`,
`retry_after_ms`; history meta `active_run`; capabilities timeout field.
`@MainActor final class AdvisorRunObserver: ObservableObject` with
`observe(ownerID: String, runID: String) async`, `suspend()`, `reset()`;
injected `fetchRun: (String) async throws -> AdvisorRunModel`,
`sleep: (Duration) async throws -> Void`. Published state holds active run,
terminal result and connection state; observer never submits POST.

- [ ] Write RED tests `testAdvisorObserverWaitsForNinetySecondRun`,
  `testAdvisorObserverForegroundRecoversByGET`,
  `testAdvisorObserverRejectsPreviousOwnerResponse`,
  `testAdvisorObserverNetworkFailureKeepsRunIdentity`. Assert 1/2/3s cadence,
  3/5/10s network backoff and no 60-iteration stop. Metadata decode works with
  new and legacy envelopes. Held stale fetch cannot erase newer state.
- [ ] Run focused native tests using simulator command below and backend
  metadata tests; confirm RED before implementation.
- [ ] Implement observer and lifecycle wiring; remove screen-owned fixed loop.
  On foreground/relaunch restore server active run from history; retain
  uncertain client_request_id until run discovery, never resubmit with new ID.
  Keep base components/style and existing SSE replay endpoints.
- [ ] Run GREEN and native UI test: run delayed >30s still shows progress;
  offline shows recovery, resume loads one answer, stop becomes cancelled.
  Reuse MPComposer/MPStatusMessage/MPButton; no screen-local controls.
- [ ] Update public guide/screen contract/OpenAPI/MCP shape tests; commit
  `fix: observe advisor runs through reconnect and app resume`.

## Task 5: Compact structured conversation context (AC7, AC9)

**Files:** Modify `backend/internal/usecase/ai_advisor_context.go`,
`ai_advisor.go`, runtime completion mapping; create
`usecase/advisor_context_state.go` and tests; extend repository expiry tests.

**Interfaces:** `AdvisorContextState` in entity holds fields in spec;
`BuildAdvisorContextState(previous entity.AdvisorContextState, generation,
throughSeq int64, results []entity.FinanceResult) (entity.AdvisorContextState,error)`;
`BuildAdvisorRequestContext(history []entity.AdvisorMessage,
state entity.AdvisorContextState) ([]AdvisorChatMessage,error)`.
Owner revalidation occurs through FinanceQueryService before system injection;
state never provides reusable financial amounts or expected_version.

- [ ] Write RED tests `TestAdvisorContextInheritsExplicitFiltersOnly`,
  `TestAdvisorContextBoundsUTF8AndCurrentMessage`,
  `TestAdvisorContextResetPurgesReferences`,
  `TestAdvisorContextAmbiguousReferencesRequestClarification`.
  Include month/year/DST boundary, deleted/foreign category, failed prior run,
  many old cards and exact 30-minute idle boundary.
- [ ] Run `cd backend && rtk go test ./internal/usecase ./internal/infrastructure/repository -run 'Advisor.*Context|Advisor.*(Idle|Reset)' -count=1`; confirm RED for new assertions.
- [ ] Build bounded context from recent messages+validated references; use
  existing summary version/generation fields, update only in CompleteRun.
  Add transcript byte accounting before provider requests; preserve valid JSON.
- [ ] Run GREEN; inspect persisted summary to confirm no money/notes/version
  snapshots. Update handbook/context contract; commit `feat: retain bounded advisor query references`.

## Task 6: Evidence protocol and server-rendered financial answers (AC8–10)

**Files:** Modify `ai_advisor_provider.go`, `ai_advisor_parts.go`,
`ai_advisor_tools.go`, `finance_query_postgres.go`, `ai/advisor_client.go`;
create entity/advisor_evidence.go, usecase/advisor_evidence.go,
usecase/advisor_grounding.go and tests. Update affected eval fixtures.

**Interfaces:** `AdvisorEvidenceIndex.Add(callID string,result entity.FinanceResult)
error`; `Resolve(ref string) (entity.Fact,error)`; `AdvisorFinalAnswer` has
mode/template/claim fields in spec; `ValidateAdvisorFinal(answer
entity.AdvisorFinalAnswer,index *AdvisorEvidenceIndex) error`;
`RenderAdvisorFinal(answer entity.AdvisorFinalAnswer,index
*AdvisorEvidenceIndex) (string,error)`.
Expose `finish_answer` as terminal protocol tool, handled separately from
finance registry and unable to write data. Empty tool batches/duplicate call
IDs/finish mixed with other calls are rejected as protocol errors.

- [ ] Write RED tests `TestAdvisorFactsHaveUniqueRunCallReferences`,
  `TestAdvisorGroundingRejectsForgedStaleAndWrongScopeFacts`,
  `TestAdvisorGroundingRendersSignedAmountsAndZeroBaseline`,
  `TestAdvisorLegacyFinancialTextUsesSafeFallback`,
  `TestAdvisorGroundingRepairCountsTowardsProviderBudget`.
  Include “ba triệu” fabricated prose, correct number wrong category, large
  integers, duplicate q1, not-found vs true zero and invalid finish calls.
- [ ] Run `cd backend && rtk go test ./internal/usecase ./internal/infrastructure/ai -run 'Advisor.*(Fact|Ground|LegacyFinancial)' -count=1`; confirm RED.
- [ ] Add typed facts for supported card kinds, unique evidence keys, strict
  final schema and templates. One repair max within six calls; safe fallback
  contains validated cards/template only. Keep protocol off by default and
  old parts/card envelope compatible. Persist fact bundle in CompleteRun.
- [ ] Run GREEN, cross-owner/lease/cancel regression and provider mocked wire
  tests. Replace misleading existing parts test with real invalid-claim
  assertions. Document guarantee boundary: server-rendered metrics are checked;
  semantic intent correctness still needs measured live evaluation.
- [ ] Update proposed grounding ADR, human guide and agent catalog; commit
  `feat: ground advisor financial responses in typed evidence`.

## Task 7: Integrated evaluation, documentation and rollout package (AC1–10)

**Files:** evals/finance-assistant/{cases.json,reliability-cases.json,README.md},
backend/internal/infrastructure/ai/advisor_quality_eval_test.go,
docs/operations/{PROVIDER_UAT,DEPLOYMENT,ROLLOUT_READINESS}.md,
docs/architecture/{AI_ADVISOR_TOOLS,API,ARCHITECTURE,ERD,INTEGRATIONS}.md,
docs/guides/finance-assistant-ios.md, native screen spec, OpenAPI artifacts,
MCP reference, docs/CONTEXT.md, docs/work/{BACKLOG,VALIDATION_MATRIX}.md,
CHANGELOG and new ADRs selected at implementation time.

**Interfaces:** Reuse production orchestrator with synthetic finance reader;
test pack carries intent/tools/date/scope/expected facts/action policy, no real
user data. No provider calls during default Go suite.

- [ ] Author >=30 synthetic scenarios (Vietnamese, mixed English, typos,
  missing/duplicate names, transfers/report exclusions, calendar edge,
  multi-turn, jar/budget proposals, CAS, unsupported banking, injection).
  Each case maps to an AC and expected deterministic result.
- [ ] Run deterministic tests/race/vet, disposable PostgreSQL migration tests,
  native unit/UI and base guards. Record attempted/pass/skipped counts,
  artifact paths and exact source SHA; do not use skips as acceptance.
- [ ] Run existing opt-in live harness with configured model, synthetic reader,
  30 scenarios x 3. Require >=95% correct tool/intent, zero cross-owner/direct
  write and 100% rendered fact consistency. Report latency/token/repair/fallback
  metrics against same-model baseline. A failure remains a release blocker.
- [ ] Reconcile every doc class listed in spec; mark changes as implemented
  only after evidence, keep current release v1.2.0 untouched by draft claims.
  Regenerate Swagger; verify route/catalog parity through shared router tests.
- [ ] Prepare concrete deployment package: immutable API/worker image, migration
  preflight, backup policy, feature flag order, fencing-aware rollback,
  health/worker/route/provider gates. Deploy only when separately requested.
- [ ] Staging kill/reconnect/DB-fault tests first; after authorized prod rollout,
  physical iPhone UAT uses isolated UAT owner and verifies persisted state and
  displayed cards, CAS conflict and cancel. Face ID/offline manual gates are
  reported separately. Do not mutate owner's feedback resolution state.
- [ ] Commit evidence/docs and propose next semantic release only after owner
  accepts the user-facing change.

## Dependency waves and integration order

`Task 1` and `Task 2` are independent implementation slices.
`Task 3` depends on 2. `Task 4` depends on 2/3 metadata. `Task 5` depends on
1/2 atomic context. `Task 6` depends on 1/2/3/5. `Task 7` integrates all.
When parallel work is explicitly authorized, separate resource reads, runtime
and native observer ownership; coordinator owns shared orchestrator/contracts,
main.go, docs index and migrations. Reserve migration number once.

## Verification commands

```sh
# Repo root; use an explicitly disposable migrated DB for integration tests.
rtk proxy bash -c 'cd backend && go test ./... -count=1'
rtk proxy bash -c 'cd backend && go test -race ./... -count=1'
rtk proxy bash -c 'cd backend && go vet ./...'
rtk proxy bash mobile-ios/scripts/check-base-components.sh
rtk proxy bash mobile-ios/scripts/verify-project.sh
rtk proxy xcodebuild test -project mobile-ios/MyPocketIOS.xcodeproj \
  -scheme MyPocketIOS \
  -destination 'platform=iOS Simulator,id=EC80B092-CEF1-4A8E-A5A3-7420298EFD4D' \
  -only-testing:MyPocketIOSTests -parallel-testing-enabled NO \
  -collect-test-diagnostics never
rtk proxy bash -c 'cd backend && AI_ADVISOR_QUALITY_LIVE=1 go test ./internal/infrastructure/ai -run TestLiveAdvisorQualityMatrix -count=1 -v'
rtk git diff --check
```

Native focused tests replace `-only-testing:MyPocketIOSTests` with the new
observer class or the owning existing suite. Run API-backed XCUITest only
against explicit disposable local API/DB; physical production smoke only with
isolated UAT owner. Do not source or print production env in evidence.

## Plan self-review and handoff

- All ACs map to Tasks 1–7; reaper fencing/partial commit/intent grounding are
  explicit tests, not implied by old regression counts.
- Existing lease/summary/fact tables and current SSE are reused; no duplicate
  retrieval project or new auth policy.
- Spec fixes the maximum context, deadlines, errors, recovery and rollback.
- Plain Markdown spec/plan are readable by people and agents; current public
  API docs stay current until implementation publishes contract additions.
- No runtime test is claimed for this planning-only change; validation is
  internal-link/file-path/spec-coverage review and `git diff --check`.
- Product-code implementation is present behind explicit runtime/grounding
  flags. This plan remains open until the three release gates above have
  reproducible evidence; local green tests do not imply production rollout.
