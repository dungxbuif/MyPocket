# Finance Assistant V1 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking. The plan started execution on 2026-09-22; unchecked items remain release work, not proof that the local slice is complete.

**Goal:** Một tab trợ lý tài chính chỉ đọc với history, tool calling, số liệu có nguồn và card tương tác trên backend Go hiện có.

**Architecture:** Go use cases own orchestration and typed finance queries; PostgreSQL owns conversation/run/fact state. React renders allowlisted parts through existing bases; model has neither database nor write access.

**Tech Stack:** Existing Go/Gin/GORM/PostgreSQL/Redis and React 19/Vite/Tailwind 4. Optional assistant-ui headless adapter after compatibility spike; no Next.js, Python, vector DB or new agent service.

**Spec:** [AI-ADVISOR-01 detail design](../../work/tickets/AI-ADVISOR-01-DETAIL_DESIGN.md). Read both documents before implementation. The spec owns financial, protocol, limits and permission semantics.

**Technical guide (required):** [Go/SQL/API/SSE/React implementation contracts](2026-09-22-finance-assistant-technical-guide.md). Read the relevant section before each task; it defines exact types, table constraints, algorithms, defaults, errors, fixtures and runtime wiring. All snippets are proposed implementation guidance, not evidence that source exists or tests passed.

## Execution checkpoint — 2026-09-22

The local vertical slice now includes Task 1 semantic query foundation, Task 2 migration/repository baseline, Task 3 read-only registry/provider, a synchronous Task 4/5 orchestration + JWT route slice, and a thin Task 6/7 React tab. Verified locally with the PostgreSQL-backed backend race suite and frontend type/design/build checks. The unchecked gates below remain intentionally open: final fact/card validation, SSE/reconnect and recovery endpoints, Redis audit, user API-key auth, browser/provider E2E, and public deployment. Do not mark the plan complete until those gates have evidence.

## How to execute this plan

Dependency order: Task1 → Task2 → Task3 → Task4 → Task5 → Task6 → Task7 → Task8. Run each task as smaller red/green cycles below; record exact result before continuing. Existing Feedback and transfer changes must stay untouched except separately documented integration points. All paths abbreviated under `entity/`, `repository/`, `usecase/`, `infrastructure/`, `controller/` mean `backend/internal/`.

Read-only baseline before editing:

```sh
rtk git status --short
rtk git branch --show-current
rtk proxy rg --files backend/migrations
```

Capture status in the work note, including tracked/untracked changes. Do not reset/reformat unrelated files. Implementation may proceed after plan review; this file is also the execution ledger for remaining release gates.

Commit unit: one tested behavior, not one arbitrary file. Stage exact paths named by the completed subtask; if existing dirty file overlaps, inspect its diff and stage only the task hunks. Never stage all worktree files. Attach evidence and current limitations to `AI-ADVISOR-01-VERIFICATION.md` as work proceeds.

## Global Constraints

- V1 chỉ đọc; V2 mới đề xuất thay đổi. Không tích hợp ngân hàng/thanh toán.
- Một active conversation/account, không thread picker/branching/regenerate V1.
- No per-user daily/monthly usage quota.
- Loop limits: at most 4 tool executions, at most 6 provider requests including final response/one repair, 30s per provider HTTP call within 90s run deadline.
- PostgreSQL authoritative; no transaction held while waiting on model.
- All UI through shared bases; no screen-only styles replacing base visuals.
- Amounts integer VND; UTC instants/account-local calendar; transfer/report/budget semantics unchanged.
- User keys are a **public-release prerequisite**, not provided by Feedback's developer token.
- Preserve dirty transfer/Feedback work. Do not change their migrations or reuse their partially implemented services as financial write tools.
- Each task updates its relevant internal/public docs before its scoped commit; stage explicit paths only.

## Review Focus

1. More than 50 matches: totals include all rows, not only page one (Task 1).
2. Ledger edited while LLM is thinking: final cards use one fresh fact bundle (Task 1/4).
3. Lost POST response/reload/multi-tab/cancel: no second provider run or resurrected history (Task 2/5).
4. Malicious notes/foreign IDs/revoked key: no privileged tool or financial data leak (Task 3/4/5).
5. Unknown/null card and masked mode: no blank page or unmasked prose/ARIA amounts (Task 6/7).

## Task 1: Semantic finance queries with accounting proof

**Read:** technical guide §§1–4,10. **Subtask exit order:**

- [ ] 1a. Implement date/money/filter types and unit tests: invalid date, DST next-midnight, max12 months, safe integer, zero baseline; no SQL yet.
- [ ] 1b. Implement one-owner report query on G1; prove1000/150/850 and wallet total830 before category/compare cards.
- [ ] 1c. Implement search + independent aggregate on G2; prove page50+1 and total51; cursor rejects scope mismatch.
- [ ] 1d. Implement category subtree/uncategorized/compare with no rollup double count; null previous percent when baseline0.
- [ ] 1e. Add budget/goal and read-only jar views on the same DB handle; test jar month/config row counts unchanged.
- [ ] 1f. Implement final snapshot bundle + writer-barrier test; then wire monthly and budget consumers with response parity tests.

**Files:** create `backend/internal/entity/finance_query.go`, `repository/finance_query.go`, `infrastructure/repository/finance_query_postgres.go`, `usecase/finance_query.go`; tests alongside each new implementation. Modify `controller/http/month_handler.go`, `controller/http/budget_handler.go` and wiring in `backend/cmd/api/main.go`. Keep existing jar mutation path separate; shared helpers may be extracted only with parity tests. Reconcile `docs/requirements/REPORTS.md`, `docs/architecture/API.md` only for implemented changes.

**Interfaces:** `FinanceQueryService.Execute(ctx, ownerID, query) (FinanceResult,error)` and `RefreshBundle(ctx, ownerID, queries) (FactBundle,error)`. `FinanceQuery` is tagged by the eight tool names and holds validated normalized filters; `FinanceResult` carries typed rows/metrics/status/scope. `FactBundle` groups results with one ID and snapshot time. Public DTOs never expose SQL.

- [ ] Add golden PostgreSQL fixture helper in the new test file: two owners, two basic wallets, Food parent/child, 1,000 income, 100+50 expenses, excluded 20 expense, transfer pair 200. Create with disposable IDs and clean only those IDs. Add tests including this assertion body:

```go
if summary.Income != 1000 || summary.Expense != 150 || summary.Net != 850 {
    t.Fatalf("incorrect report: %+v", summary)
}
if food.Amount != 150 { t.Fatalf("subtree duplicated/missing: %+v", food) }
if page.TotalCount != 51 || len(page.Items) != 50 {
    t.Fatalf("aggregate must not use page count: %+v", page)
}
```

- [ ] Run from `backend/`: `rtk go test ./internal/infrastructure/repository -run TestFinanceQuery -count=1`; expect missing implementation/failing behavior, not DB unavailable. Ensure test database env points at isolated DB, never infer production credentials.
- [ ] Implement bounded parameterized owner queries, stable cursor `(occurred_at,id)`, normalized filters, all-page aggregates, category descendants and the existing report exclusions. `RefreshBundle` runs all queries inside one short read-only REPEATABLE READ transaction. Test writer commit between provisional and final reads, ensuring final income/category totals agree. Implement overflow, timezone/DST and baseline-zero cases from spec.
- [ ] Replace monthly income/expense/category calculation with this service while preserving jars/manual note/response fields. Add parity assertions against existing monthly response and budget scope (do not rewrite budget formula independently).
- [ ] Explicitly ban existing `JarRepository.ListMonth` in read tools: it initializes month rows. Implement snapshot SELECT-only jar view with `config_origin=stored|inherited_preview|unconfigured` per technical guide. Extend read-only assertions from transactions alone to wallets/categories/budgets/jars/jar_months/jar_month_configs/month_notes.
- [ ] Run targeted tests and `rtk go test ./...`; record DB tests actually executed. Commit explicit query/month/docs files as `feat: add shared finance query service`.

## Task 2: Durable single conversation and bounded context

**Read:** technical guide §§5–6. **Subtask exit order:**

- [ ] 2a. Add migration + repository conversation lookup and paginated messages, with composite owner FK and uniqueness tests.
- [ ] 2b. Implement StartRun transaction: same-key replay before busy check; one persisted user message under20 concurrent submits.
- [ ] 2c. Implement claim/heartbeat/fencing, atomic event sequence, terminal completion; establish conversation→run lock order.
- [ ] 2d. Implement cancel/clear/purged tombstones; prove late result/late same-key POST cannot resurrect cleared history.
- [ ] 2e. Implement deterministic context anchors +12 recent messages and token budget; no extra uncounted summary-model call.
- [ ] 2f. Exercise up/down-one/up using golang-migrate in disposable DB test; existing migration CLI has no down command.

**Files:** create `entity/ai_advisor.go`, `repository/ai_advisor.go`, `infrastructure/repository/ai_advisor_postgres.go`, `usecase/ai_advisor_context.go`, corresponding `_test.go`; next available migration pair under `backend/migrations/`. Update `docs/architecture/ERD.md` and architecture lifecycle docs.

**Interfaces:** exact `AdvisorRepository` signatures and entity types in guide §5; includes heartbeat, read-run/read-events and interrupt-expired methods. Every operation includes owner and expected generation/lease where relevant. Context builder consumes persisted profile/summary/recent messages; does not query transaction rows itself.

- [ ] Write repository concurrency tests: unique conversation per owner, identical client_request_id returns original run, changed text returns conflict, second active run denied. Test clear history races completion and summary CAS.

```go
if first.RunID != replay.RunID { t.Fatal("replay created a second run") }
if !errors.Is(conflict, repository.ErrAdvisorRequestConflict) { t.Fatal("payload reuse accepted") }
if len(afterClear.Messages) != 0 { t.Fatal("late completion restored history") }
```

- [ ] Run `rtk go test ./internal/infrastructure/repository -run TestAdvisor -count=1` and observe red before implementation.
- [ ] Implement spec tables/unique constraints/indexes and state transitions. Use bounded JSONB parts/facts; no persisted provider secret/raw reasoning. Choose migration number after listing current migrations, not fixed 18 by assumption. Up/down/up proof only on disposable DB.
- [ ] Implement summary watermark + max 12 recent messages/16k input policy; user text and summary bounds. Build a token-budget interface to account for model context size; test summary omits ledger as current truth and “tháng trước” resolves from prior selected period. Invalid/oversized context returns explicit error, not clipped system policy.
- [ ] Run repository + usecase tests; prove zero financial table mutation by history clear. Record migration and retention behavior; commit as `feat: persist advisor conversation and runs`.

## Task 3: Read-only registry and separate provider adapter

**Read:** technical guide §§3,6. **Subtask exit order:**

- [ ] 3a. Add schema source and typed decoders for eight tool inputs; required-field/unknown-field/duplicate-key/enum/range tests precede execution.
- [ ] 3b. Wire each tool to Task1, test owner IDs and scope; unknown/write tools must fail without reaching a repository.
- [ ] 3c. Implement HTTP provider adapter with full tool_call_id roundtrip; fake-server tests assert exact envelope, timeout and redirect policy.
- [ ] 3d. Add separate plan/answer prompts and final structured AnswerPlan schema; verify no effect on existing entry extraction.

**Files:** create `usecase/ai_advisor_tools.go`, `usecase/ai_advisor_provider.go`, `infrastructure/ai/advisor_client.go`, `infrastructure/ai/advisor_plan.v1.txt`, `infrastructure/ai/advisor_answer.v1.txt`, tests alongside. Keep `infrastructure/ai/client.go` extractor behavior unchanged. Create `docs/architecture/AI_ADVISOR_TOOLS.md` as plain Markdown with JSON schema examples.

**Interfaces:** tool `Definition() ToolDefinition`, `Execute(ctx, Principal, json.RawMessage) (FinanceResult,error)`; provider `Chat(ctx, AdvisorRequest) (AdvisorResponse,error)`. Response contains either typed tool calls or answer segments referencing facts. Server supplies owner/scopes; tool argument schema forbids owner_id and unknown fields.

- [ ] Write fake-provider/registry tests for each tool and negative calls (`execute_sql`, `create_transaction`, foreign wallet, invalid dates/negative limit/51 page size, injected note). Test only 8 documented read tools registered and registry arguments schema matches public catalog.

```go
for _, forbidden := range []string{"execute_sql", "create_transaction", "confirm_action"} {
    if _, ok := registry.Lookup(forbidden); ok { t.Fatalf("write/escape tool: %s", forbidden) }
}
```

- [ ] Run `rtk go test ./internal/usecase ./internal/infrastructure/ai -run 'TestAdvisor|TestExtract' -count=1`; new registry tests red, existing entry must remain green.
- [ ] Implement strict schema validation, ID scoping, sanitized tool errors and mapping to Task 1. Separate provider request struct supports OpenAI-compatible tools and tool_call_id messages; preserve request IDs, body/deadline caps and no credential-forwarding redirects. Never downgrade unsupported tool calling into SQL or prose-generated totals.
- [ ] Add `httptest.Server` provider protocol fixtures for multiple tool calls, malformed arguments, provider error, timeout, unsupported tool mode. Provider capability false disables chat with useful UI, not extraction.
- [ ] Verify catalog/examples against fixture schema. Commit as `feat: add read-only advisor tool registry`.

## Task 4: Bounded orchestrator and trustworthy parts

**Read:** technical guide §§5–6. **Subtask exit order:**

- [ ] 4a. Fake model one-tool→answer case, provider/tool counters and all tool_call IDs resolved.
- [ ] 4b. Add multi-step, over-budget and denied-tool cases; reserve final-answer call inside six-call budget.
- [ ] 4c. Discard provisional numeric context, refresh snapshot and build template/card parts from final facts only.
- [ ] 4d. Add context cancellation/credential validation and lease fencing at every publication boundary.
- [ ] 4e. Add runner bounded worker pool, startup recovery and shutdown; no provider replay for expired running leases.

**Files:** create `usecase/ai_advisor.go`, `usecase/ai_advisor_parts.go`, `usecase/ai_advisor_runner.go`, tests; add fixtures `backend/internal/usecase/testdata/advisor/`. Update `docs/architecture/AI_ADVISOR_TOOLS.md` with parts version/provenance.

**Interfaces:** `AdvisorService.Run(ctx, principal, runID, leaseToken) error` uses Task 2 storage, Task 3 provider/registry and Task 1 RefreshBundle. `BuildParts(answer, finalBundle)` returns server-validated versioned parts. `Principal` includes owner, credential ID/type and effective scopes; not model-controlled.

- [ ] Script fake model replies: ask category summary → comparison → answer with fact IDs. Add unknown fact, altered amount, malformed part, extra tool after final bundle, repeated tool loop, summary injection, cancellation and lease-loss cases.

```go
if executedWrites != 0 { t.Fatal("advisor mutated financial data") }
if calls > 6 || toolsExecuted > 4 { t.Fatal("unbounded run") }
if card.Amount != finalBundle.Expense { t.Fatal("model-authored monetary value") }
```

- [ ] Run `rtk go test ./internal/usecase -run TestAdvisor -count=1`, confirm failures attributable to missing orchestrator.
- [ ] Execute tools sequentially V1 under one deadline; append tool status/events without raw prompts/results in audit. Store normalized requests, refresh one final bundle, request final grounded answer, then build quantitative text/cards from fact references. Safe schema repair max once within provider-call budget; no raw numeric model fallback.
- [ ] Implement capability-filtered questions and partial/unsupported/error parts; service never calls write repos. For current-owner accounting ambiguity ask a clarifying question; no defaults that silently change report scope.
- [ ] Tests include new transaction committed during provider delay, final matched facts, no ledger delta, and user text attempting owner override. Commit as `feat: orchestrate grounded finance answers`.

## Task 5: HTTP/SSE, auth boundary and recovery

**Read:** technical guide §§7–8,11. **Subtask exit order:**

- [ ] 5a. Add typed principal alongside legacy userID context; session expiry/revocation tests.
- [ ] 5b. Mount capabilities/overview/history/submit/run/by-request/cancel/clear routes with strict error mapping and max-body checks.
- [ ] 5c. Implement persisted-event SSE/replay/compaction, then reconnect and expired-cursor tests; GET must never call provider.
- [ ] 5d. Implement owner-scoped fact-bundle drilldown; all filters server-resolved, no arbitrary model URL.
- [ ] 5e. Add bounded Redis audit/metrics and feature flag/off recovery; test auth Redis failure versus audit-delivery failure separately.
- [ ] 5f. Complete the user-key prerequisite independently before third-party release; run scope parity and mid-run revocation tests against real middleware.

**Files:** create `controller/http/ai_advisor_handler.go`, `routes_ai_advisor.go`, tests; wire `backend/cmd/api/main.go`; extend config/env example with an advisor-enabled flag default false. Add `infrastructure/repository/advisor_audit_redis.go` and tests only if existing audit interface cannot serve scoped read events safely. Update API/operations docs and regenerate Swagger for actual routes only.

**Interfaces:** routes/envelopes/errors/events exactly in spec §7. HTTP handler resolves principal; runner has app lifecycle context bounded by run deadline, not SSE connection lifetime. Event reader authenticates owner and credential on reconnect; side effects happen only on POST accepted run.

- [ ] Table-test no auth, other owner, same idempotency key, modified payload, busy run, cancel, GET replay, clear-history during run; verify no provider calls from GET. SSE decoder test splits records across arbitrary network chunks and resumes by sequence.

```go
if callsAfterReplay != callsBeforeReplay { t.Fatal("GET replay invoked model") }
if foreignStatus != 404 { t.Fatalf("foreign run exposed: %d", foreignStatus) }
if duplicateSeqs != 0 { t.Fatal("duplicate events reached message reducer") }
```

- [ ] Run `rtk go test ./internal/controller/http -run TestAdvisor -count=1` red, then add handlers/runner claim/reaper behavior, SSE heartbeat/no-cache and startup interruption recovery from spec. Check failover fencing and cancellation before publishing each final result.
- [ ] Integrate JWT initially; **do not enable API-key claims until TICKET-10-01 is implemented**. Before public third-party release complete that separate prerequisite with: one-time secrets + digest storage, owner scopes, immediate revocation, self-info/self-revoke only, audit and isolated cross-owner tests. Add advisor JWT/key parity and revoke-mid-run cases once real key middleware exists. Missing prerequisite means browser pilot only, not public API ready.
- [ ] Read audit uses bounded delivery, safe fields, operational failure counter/log without raw finance data. Test Redis down does not hang request or bypass auth. Financial mutation audit outbox belongs to V2, not needed to claim V1 read-only.
- [ ] Run handler/repository tests and `rtk proxy go generate ./cmd/api`; inspect generated endpoints. Document reverse-proxy SSE settings/run interruption and flag-off fallback; commit as `feat: expose recoverable advisor API`.

## Task 6: Typed frontend adapter and reusable finance cards

**Read:** technical guide §§8–9. **Subtask exit order:**

- [ ] 6a. Parse DTOs/parts as unknown with runtime checks; unknown part fallback before rendering any card.
- [ ] 6b. Implement byte/chunk-safe SSE parser and ordered event reducer; duplicates/gaps/generation changes tested independently.
- [ ] 6c. Implement JSON service + authenticated streaming reader + same-request recovery; never feed SSE to apiRequest.
- [ ] 6d. Write shared component contracts then metric/category/comparison/message/tool-status molecules; mask DOM/ARIA amounts.
- [ ] 6e. Extend AssistantComposer with advisor mode and IME-safe Enter while preserving entry defaults; evaluate optional headless adapter in a fixture only.

**Files:** create `app/src/services/aiAdvisor.ts`, `app/src/services/aiAdvisorParts.ts`, `app/src/services/aiAdvisorStream.ts`, `app/src/atomic/molecules/AssistantMessageList.tsx`, `AssistantToolStatus.tsx`, `AssistantPart.tsx`, `FinanceMetricGroup.tsx`, `FinanceComparisonCard.tsx`, `FinanceCategoryBreakdown.tsx`; modify `AssistantComposer.tsx`; tests `app/scripts/ai-advisor.test.ts`. Reuse existing budget/goal/transaction bases. Add planned page spec `docs/design/pages/finance-assistant/README.md` and update base inventory before JSX.

**Interfaces:** discriminated `AdvisorPart` union validates `parts_version=1`, money/range/arrays; `reduceAdvisorEvent(state,event)` dedupes `(run_id,seq)` and upserts stable part IDs. Service owns POST/recovery/fetch SSE; renderer cannot call arbitrary API URLs from part data.

- [ ] Add parser/reducer tests for null arrays, invalid/overflow amount, unknown version/type, repeated events, run replacement and fragmented SSE. Unknown card fallback remains renderable; malformed known part shows warning without throwing.

```ts
assert.doesNotThrow(() => renderPart(parsePart({type: 'future_card', data: null})))
assert.equal(reduceAdvisorEvent(reduceAdvisorEvent(state, event), event).parts.length, 1)
```

- [ ] Run `rtk proxy node --test scripts/ai-advisor.test.ts` from `app/` and observe failures before implementation. Define helper fixtures in that test file rather than importing runtime mocks.
- [ ] Write base contracts (states/keyboard/tokens/ARIA); implement cards using SurfaceCard/Text/Heading/BaseButton/Progress and domain bases, layout-only screen classes. Keep money source/from/to/as_of and unknown/partial states visible. Drilldown uses typed scope references.
- [ ] Spike assistant-ui ExternalStoreRuntime with one fake local message, one card, cancel, history load and keyboard focus **in a test fixture only**. Record build/design guard and package changes. Adopt pinned compatible version only if it replaces shell lifecycle work without violating base controls; otherwise remove spike dependency and ship thin shell. No simultaneous AI SDK/CopilotKit runtime installation.
- [ ] Run adapter/card tests, `rtk npm run check:design`, `rtk npm run test:design`, `rtk npm run build`. Commit chosen implementation and decision record as `feat: add typed finance assistant cards`.

## Task 7: Real Assistant tab and end-to-end state verification

**Read:** technical guide §§9–10. **Subtask exit order:**

- [ ] 7a. Route/tab/page mapping and measured six-slot navigation; normal Add/hold Add remain working.
- [ ] 7b. Wire overview/history/composer and source-backed cards, no initial model call or runtime fake data.
- [ ] 7c. Real API/DB fake-model browser test G1→Food drilldown→comparison→reload; assert finance state unchanged.
- [ ] 7d. Recover lost POST/SSE, cancel and clear during generation; repeat at mobile widths and keyboard/IME.
- [ ] 7e. Run existing entry/transfer/design regression, capture visual and state proof and record remaining UAT separately.

**Files:** create `app/src/atomic/organisms/FinanceAssistantPanel.tsx`; modify `app/src/router.tsx`, `atomic/pages/FinancePrototypePage.tsx`, `atomic/organisms/BottomNavigation.tsx` as required for `/assistant`; add `app/scripts/ai-advisor-integrated.mjs`, `ai-advisor-browser.test.mjs`, test fixtures under `app/tests/`. Update `app/package.json` scripts and assistant screen spec.

**Interfaces:** panel consumes services/parts from Task 6 and real APIs Task 5; one active conversation, no multi-thread UI. Navigation preserves normal click Add/manual and hold Add/one-shot entry.

- [ ] Write real-browser test with fake model behind real local Go API and isolated PostgreSQL fixtures. Ask monthly spending, click Food drilldown, ask “so tháng trước?”, reload, page history, reconnect/cancel. Assert UI amounts AND API/DB state, not only screenshot presence.

```js
assert.equal(await ui.foodAmount(), '150 ₫')
assert.equal(await api.reportExpense(), 150)
assert.deepEqual(await db.ledgerRows(), beforeChatLedgerRows)
assert.equal(await ui.visibleMessageCountAfterReload(), persistedMessageCount)
```

- [ ] Implement test helpers locally with deterministic fixture IDs/clock; do not reuse a test that edits the owner's real note/timezone. Run script red before mounting route, then implement panel's dashboard/suggestions/history/composer/recovery and shared drilldown.
- [ ] Add mobile 390×844, desktop keyboard, IME, scroll-anchor, reduced motion, masked money in prose/ARIA, unconfigured provider and unknown part regression cases. Attachments remain on quick-add only; no misleading disabled “subscription” prompt in V1.
- [ ] Add package scripts `test:advisor`, `test:advisor:integrated`, `test:advisor:e2e`. Run all three against fixture API/DB, plus existing `test:ai`, `test:transfer:integrated`, `test:transfer:e2e`, design checks and build. Clearly distinguish fixture-model proof from actual provider eval.
- [ ] Record screenshots + request/result IDs + golden expectations without real account data; commit as `feat: add persistent financial assistant tab`.

## Task 8: Provider evaluation, public docs and release readiness

**Read:** technical guide §§10–13. **Subtask exit order:**

- [ ] 8a. Add30 synthetic cases and required completion/fact assertions; fake and actual provider runs reported separately.
- [ ] 8b. Opt-in actual provider eval90 executions; no private data or ledger mutation, record failures without threshold weakening.
- [ ] 8c. Run no-skipped-DB release suite and proxy/mobile/revocation smoke tests, including all new recovery endpoints.
- [ ] 8d. Publish only implemented human/API/agent docs with exact error/event/schema examples and data-sharing controls.
- [ ] 8e. Record release readiness/key prerequisite/rollout flag and owner UAT; do not confuse staging proof with deploy completion.

**Files:** create `backend/internal/infrastructure/ai/advisor_live_eval_test.go`, synthetic eval fixtures, `docs/work/tickets/AI-ADVISOR-01-VERIFICATION.md`, `docs/guides/finance-assistant.md`, `docs/ai/finance-assistant.md`, versioned tool/part schemas under `docs/ai/`; update docs entry/API/architecture/ERD/runbook/validation/context/backlog/changelog.

- [ ] Implement opt-in live eval using existing backend configuration without printing secrets. Use spec's 30 cases ×3; save prompt/schema/model versions and pass/fail counts/latencies, no private raw prompts. Confirm the configured provider actually supports tool calls, not just entry extraction JSON schema.
- [ ] Require 100% fact correctness/no cross-owner or writes and ≥90% answerable-task completion on dataset. Failures keep advisor disabled; do not weaken assertions to pass. Broad real-user accuracy is still not guaranteed by this dataset.
- [ ] Run final backend race + real DB suite and frontend/integrated/browser suites. Verify proxy-delivered SSE on staging, restart recovery, key revocation and provider-unavailable fallback. Local key revocation is proven; public release still waits for Redis audit, configured-provider/browser and deployment evidence.
- [ ] Human docs: getting started, what is read/sent to provider, examples, scopes/timezone, history/delete, stale-card/retry semantics, limits vs no usage quota, unsupported capabilities. AI surface: stable Markdown, JSON schemas and endpoint catalog; generated OpenAPI only for implemented endpoints. Link from `docs/README.md`; inspect actual hosting source before any publication/deploy.
- [ ] Reconcile all trace links and write honest verification matrix, residual gaps and rollout/rollback (flag off, preserve data). No destructive migration rollback as an AI disable mechanism. Commit docs/evidence as `docs: document verified finance assistant v1`.

## Self-review and handoff

Coverage: spec §4→Task1; §6→Task2; §3/5→Task3; §7/9 facts→Task4; §7/8→Task5; §9 UI→Tasks6/7; §11/13→Task8. V2 actions and later roadmap are explicitly outside V1; not marked implemented. Five review-focus risks each have owning tests above. The technical guide adds exact contracts and named tests, including the newly observed jar-read side effect, lost-submit lookup, direct drilldown, deterministic context anchors and migration CLI limitation.

For each subtask, create the named failing test first, run its narrow command, implement the corresponding guide algorithm, then rerun and commit scoped changes with doc evidence. Examples in the task bodies are assertion excerpts; use the complete guide fixture contract and test helper, not undefined UI/DB stubs. No task may be called done on a compile-only or skipped-DB run.

Planning verification checks Markdown targets and whitespace only. None of these runtime tests has run for the unimplemented advisor. Feedback task remains independently unfinished, with its existing files preserved. Next step after plan review is Task 1, not a deploy of this document.
