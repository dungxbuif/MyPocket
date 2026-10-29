# AI Advisor Quality Harness Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans (or superpowers:subagent-driven-development) to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a deterministic Go contract harness, a synthetic Promptfoo regression suite, and optional OpenTelemetry/Langfuse tracing so Finance Assistant quality is measurable without removing runtime safety limits or exposing user financial data.

**Architecture:** Keep `AdvisorOrchestrator`, the typed read-only registry, and owner-scoped finance services as the production boundary. The Go harness proves business and authorization invariants with fake providers/fixtures; Promptfoo evaluates real model tool selection against synthetic facts; OpenTelemetry traces the production run/provider/tool/persistence spans and exports asynchronously when explicitly enabled. No external agent framework replaces the existing Go use case.

**Tech Stack:** Go 1.26, existing `go test`/`httptest`/PostgreSQL fixtures, Promptfoo `0.123.1` as an evaluation-only Node 22+ dependency, OpenTelemetry Go `v1.47.0` plus `go.opentelemetry.io/contrib/exporters/autoexport v0.72.0`, OTLP/HTTP, and optional self-hosted or cloud Langfuse.

**Spec:** [AI Advisor tool catalog](../../architecture/AI_ADVISOR_TOOLS.md), [Finance Assistant design](../../work/tickets/AI-ADVISOR-01-DETAIL_DESIGN.md), and the approved harness requirements in the requesting conversation. The implementation must reconcile the catalog's stale four-call wording with the current eight-call runtime budget.

## Global Constraints

- Finance Assistant V1 remains read-only; the registered tool set contains no transaction mutation, bank, payment, SQL, arbitrary URL, or confirmation tool.
- The authenticated `Principal.OwnerID` is the only owner scope; the model never supplies or overrides it.
- No per-user daily or monthly usage quota is introduced.
- The production circuit breaker remains bounded at eight tool executions and six provider requests per assistant turn; these are reliability guards, not user limits.
- Amounts are integer VND and account-local calendar ranges; transfers, report scope, budgets, goals, and jars keep existing semantics.
- Go contract tests use isolated synthetic fixtures and must not connect to production or use real user rows.
- Promptfoo is a dev/evaluation dependency only; no Promptfoo or Langfuse client is shipped to the browser or mobile app.
- Telemetry is fail-open: a missing/unreachable collector must never fail an advisor request.
- Traces, eval artifacts, logs, and release docs must not contain API keys, prompts, OCR text, notes, raw financial rows, or unredacted user identifiers.
- Deterministic business/security assertions are mandatory gates; LLM-as-judge may report language quality but cannot decide monetary correctness.
- Every implementation task updates the directly affected internal/public docs; a production release updates `CHANGELOG.md` and a versioned release note with fresh evidence.

## Review Focus

1. A plausible final answer uses the wrong tool or wrong month range — Promptfoo trajectory and argument assertions must fail it.
2. A prompt attempts `owner_id`, a foreign wallet, SQL, or a write action — Go contract tests must reject it before repository access.
3. Unknown fields, duplicate/trailing JSON, invalid dates, or oversized limits — strict argument/schema tests must fail closed.
4. The model returns eight-plus calls, repeats a call, times out, or the provider returns malformed JSON — the run must terminate within the existing budgets with a safe error.
5. OTel/Langfuse is disabled or unavailable — the answer must still complete and no sensitive attribute may be exported.

---

### Task 1: Publish a single tool-catalog contract

**Files:**
- Modify: `backend/internal/usecase/ai_advisor_tools.go`
- Create: `backend/internal/usecase/ai_advisor_tools_catalog_test.go`
- Create: `backend/cmd/ai-advisor-catalog/main.go`
- Create: `evals/finance-assistant/tools.json` (generated output)
- Modify: `docs/architecture/AI_ADVISOR_TOOLS.md`

**Interfaces:**
- Produce `func (r *AdvisorToolRegistry) Definitions() []ToolDefinition`, returning a new slice sorted by `ToolDefinition.Name` and never exposing the registry map.
- Produce a catalog command that writes stable, indented JSON containing the eight OpenAI-compatible function definitions and no credentials or owner fields.
- Preserve `Lookup` and `Execute` behavior; catalog generation must not instantiate or mutate a finance repository.

- [ ] **Step 1: Write failing catalog tests.** Assert that `Definitions()` is sorted, contains exactly the eight documented read tools, every schema has `type: object`, `additionalProperties: false`, and no property named `owner_id`, `user_id`, `sql`, or `url`.
- [ ] **Step 2: Run the focused test to verify the contract is not yet exposed.**

  Run: `cd backend && go test ./internal/usecase -run 'TestAdvisorToolDefinitions' -count=1`

  Expected: FAIL because `Definitions` does not exist yet.
- [ ] **Step 3: Implement `Definitions()` and the catalog command.** Reuse each registered tool's existing `Definition()`; do not duplicate schemas in the command. Use `NewAdvisorToolRegistry(nil)` because definition generation must not execute a query.
- [ ] **Step 4: Generate and verify the checked-in catalog.**

  Run: `cd backend && go run ./cmd/ai-advisor-catalog > ../evals/finance-assistant/tools.json && go test ./internal/usecase -run 'TestAdvisorToolDefinitions' -count=1`

  Expected: PASS; the generated JSON is deterministic on a second run (`cmp` has no diff).
- [ ] **Step 5: Reconcile the tool catalog docs.** Replace the stale “after four tool executions” sentence with the eight-execution budget, link the harness plan, and document that `tools.json` is generated from Go rather than hand-edited.
- [ ] **Step 6: Commit the independently testable catalog contract.**

  ```bash
  git add backend/internal/usecase/ai_advisor_tools.go backend/internal/usecase/ai_advisor_tools_catalog_test.go backend/cmd/ai-advisor-catalog/main.go evals/finance-assistant/tools.json docs/architecture/AI_ADVISOR_TOOLS.md
  git commit -m "test: publish advisor tool catalog contract"
  ```

### Task 2: Build the Go contract harness for business and security invariants

**Files:**
- Create: `backend/internal/usecase/ai_advisor_harness_test.go`
- Modify: `backend/internal/usecase/ai_advisor_provider_test.go`
- Modify: `backend/internal/usecase/ai_advisor_tools_test.go`
- Modify: `backend/internal/usecase/ai_advisor_parts_test.go`
- Create: `backend/internal/usecase/testdata/advisor_harness/finance_cases.json`
- Create: `docs/architecture/AI_ADVISOR_HARNESS.md`

**Interfaces:**
- Test-only `advisorHarnessFixture` owns two principals, deterministic `financeReaderStub`, a fake provider, and before/after ledger counters. It must call the real `AdvisorOrchestrator`, `AdvisorToolRegistry`, `DecodeAdvisorToolArgs`, `BuildAdvisorParts`, and `BuildAdvisorContext`.
- Each case records `name`, `prompt`, expected tool names/arguments, expected owner, expected date range, expected fact values, and whether a repository mutation is allowed. Mutation is always false for this V1 harness.

- [ ] **Step 1: Add red tests for the review-focus cases.** Add exact tests:
  - `TestAdvisorHarnessUsesAuthenticatedOwnerForEveryTool`
  - `TestAdvisorHarnessRejectsOwnerOverrideAndForeignIDs`
  - `TestAdvisorHarnessRejectsUnknownAndWriteToolsBeforeReader`
  - `TestAdvisorHarnessRejectsMalformedAdvisorArguments`
  - `TestAdvisorHarnessPreservesSeptemberRangeFromToolArguments`
  - `TestAdvisorHarnessLeavesLedgerUnchanged`
  - `TestAdvisorHarnessCapsEightToolsAndSixProviderCalls`
  - `TestAdvisorHarnessBuildsOnlyAllowlistedParts`
  - `TestAdvisorHarnessFinalAnswerUsesFixtureFacts`

  The owner test must assert every reader call receives `owner-a`, even when the prompt contains `owner_id: owner-b`; the foreign-ID test must assert no reader call occurs for an unresolved wallet/category/transaction ID.
- [ ] **Step 2: Run the focused tests before implementation.**

  Run: `cd backend && go test ./internal/usecase -run 'TestAdvisorHarness|TestAdvisorOrchestrator' -count=1`

  Expected: new tests fail only at the missing harness/contract behavior; existing advisor tests remain green.
- [ ] **Step 3: Implement the fixture and deterministic oracle in `_test.go`.** Seed synthetic facts for September 2026, a previous period, two owners, wallets, categories, budgets, goals, and jars. Do not use production credentials, random wall-clock ranges, or real user text.
- [ ] **Step 4: Pin tool/argument/security behavior.** Use the real registry and fake provider to assert exact normalized `{from,to}` ranges, `limit <= 50`, unknown fields rejected, unsupported/write tool names rejected, and all monetary values come from `FinanceResult.View` rather than model prose.
- [ ] **Step 5: Pin reliability behavior.** Exercise one response with five parallel calls, a repeated-call loop beyond eight, six provider requests, context cancellation, provider error, malformed empty answer, and principal revocation between provider/tool boundaries. Assert safe typed errors and zero ledger writes.
- [ ] **Step 6: Pin response-card behavior.** Feed known, unknown, malformed, `not_found`, and `unsupported` views through `BuildAdvisorParts`; assert only allowlisted part types are returned and malformed views degrade to the safe text fallback.
- [ ] **Step 7: Run the full Go suite and commit.**

  Run: `cd backend && go test ./...`

  Expected: zero failures; database-backed tests must use an explicitly isolated test database when enabled.

  ```bash
  git add backend/internal/usecase/*advisor*_test.go backend/internal/usecase/testdata/advisor_harness docs/architecture/AI_ADVISOR_HARNESS.md
  git commit -m "test: add finance advisor contract harness"
  ```

### Task 3: Add the Promptfoo synthetic model-evaluation suite

**Files:**
- Create: `evals/finance-assistant/package.json`
- Create: `evals/finance-assistant/package-lock.json`
- Create: `evals/finance-assistant/promptfooconfig.yaml`
- Create: `evals/finance-assistant/provider.mjs`
- Create: `evals/finance-assistant/oracle.mjs`
- Create: `evals/finance-assistant/prompts/advisor-system.md`
- Create: `evals/finance-assistant/datasets/finance-core.yaml`
- Create: `evals/finance-assistant/datasets/security-boundary.yaml`
- Create: `evals/finance-assistant/datasets/vietnamese-prompts.yaml`
- Create: `evals/finance-assistant/assertions/tool-routing.mjs`
- Create: `evals/finance-assistant/assertions/tool-arguments.mjs`
- Create: `evals/finance-assistant/assertions/finance-grounding.mjs`
- Create: `evals/finance-assistant/assertions/response-shape.mjs`
- Create: `evals/finance-assistant/README.md`
- Modify: `docs/architecture/AI_ADVISOR_HARNESS.md`

**Interfaces:**
- `provider.mjs` implements Promptfoo's `ApiProvider` interface: `id()` and `async callApi(prompt, context, options)` returning `{output, tokenUsage, metadata}`.
- The provider reads `AI_BASE_URL`, `AI_API_KEY`, and `AI_MODEL` from the eval environment, loads generated `tools.json`, and runs at most six provider requests/eight synthetic tool executions.
- `oracle.mjs` executes no database calls. It maps only the documented tool names and validated arguments to fixed synthetic facts; any unknown/write tool returns an evaluation error.
- `metadata.toolCalls` contains `{name,input,output,is_error}` and `metadata.fixtureFacts` contains only synthetic values so assertions can inspect trajectory without exposing production data.

- [ ] **Step 1: Add a config/provider smoke test before live calls.** Test the oracle with fixture prompts and a fake fetch response; assert it rejects missing env, unknown tools, owner override, and malformed JSON without sending a provider request.
- [ ] **Step 2: Install the pinned evaluation dependency.** From `evals/finance-assistant/`, install `promptfoo@0.123.1` with Node 22+ and commit the lockfile. Keep it outside the production `app` and backend dependency graphs.
- [ ] **Step 3: Implement the multi-turn custom provider.** Send the system prompt, user prompt, generated OpenAI-compatible tools, and tool results from `oracle.mjs`; preserve tool-call IDs; stop on final text; return usage and bounded trajectory metadata. Never send real account IDs or raw financial text.
- [ ] **Step 4: Create the 30-case synthetic dataset.** Include 12 finance cases (summary, search, compare, wallets, budget, goal, jar), 10 Vietnamese date/filter variants, and 8 security/negative cases (owner override, foreign ID, SQL, write request, unsupported scope, malformed range, repeated tool, empty answer).
- [ ] **Step 5: Implement deterministic assertions.**
  - `tool-routing.mjs`: expected tool set/order and zero forbidden tools.
  - `tool-arguments.mjs`: exact account-local ranges, IDs, limits, and no owner-controlled fields.
  - `finance-grounding.mjs`: final numeric tokens are drawn from `fixtureFacts`; wrong or invented amounts fail.
  - `response-shape.mjs`: non-empty answer, no SQL/credentials, safe refusal for unsupported/write requests.
  LLM-rubric may be reported for clarity, but it must not gate amounts, ownership, or tool correctness.
- [ ] **Step 6: Add scripts and run modes.** Define `npm run eval:unit`, `npm run eval:ci`, `npm run eval:live`, and `npm run eval:view`. Unit mode never calls a provider; CI/live require an explicit `AI_EVAL_LIVE=1`, staging endpoint, synthetic-only credentials, and `--no-cache --ci`.
- [ ] **Step 7: Run the eval suite and record thresholds.** Mandatory deterministic thresholds: owner/security 100%, forbidden writes 0, schema/argument checks 100%, grounded monetary output 100%, routing at least 95%, response shape 100%. A live-provider failure is recorded with model/request metadata and never hidden by lowering thresholds.
- [ ] **Step 8: Commit the evaluation harness.**

  ```bash
  git add evals/finance-assistant docs/architecture/AI_ADVISOR_HARNESS.md
  git commit -m "test: add synthetic finance advisor evaluations"
  ```

### Task 4: Add fail-open OpenTelemetry/Langfuse tracing

**Files:**
- Modify: `backend/go.mod`
- Modify: `backend/go.sum`
- Create: `backend/internal/infrastructure/observability/otel.go`
- Create: `backend/internal/infrastructure/observability/otel_test.go`
- Modify: `backend/internal/config/config.go`
- Modify: `backend/internal/config/local_env_test.go`
- Modify: `backend/cmd/api/main.go`
- Modify: `backend/internal/usecase/ai_advisor.go`
- Modify: `backend/internal/usecase/ai_advisor_provider.go`
- Modify: `backend/internal/infrastructure/ai/advisor_client.go`
- Modify: `docs/operations/ENVIRONMENTS.md`
- Modify: `docs/architecture/AI_ADVISOR_HARNESS.md`

**Interfaces:**
- Add `observability.Config{Enabled bool, ServiceName string}` and `func Setup(ctx context.Context, cfg Config) (shutdown func(context.Context) error, err error)`.
- When enabled, `Setup` creates an OTLP/HTTP batch tracer using standard `OTEL_*` environment variables; when disabled it returns a no-op shutdown function and must not attempt network I/O.
- Safe spans are named `advisor.run`, `provider.chat`, `advisor.tool.<allowlisted-name>`, and `advisor.answer.persist`. Attributes are bounded metadata only: model, tool count/name, provider call count, status, latency, error kind, and trace ID.
- `AdvisorAnswer.Usage` receives the trace ID only when a valid trace exists; existing `AI_STORE_USAGE` persistence remains the source of truth and no migration is added.

- [ ] **Step 1: Add red observability tests.** Use `tracetest.InMemoryExporter` to assert the four span names, parent/child relationships, status/error attributes, and absence of prompt, note, OCR, amount, owner ID, API key, or raw tool argument attributes. Add a disabled-mode test proving no exporter is created.
- [ ] **Step 2: Add pinned OTel dependencies.** Add `go.opentelemetry.io/otel v1.47.0`, `go.opentelemetry.io/otel/sdk v1.47.0`, `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.47.0`, and `go.opentelemetry.io/contrib/exporters/autoexport v0.72.0`; run `go mod tidy` and keep the production binary behavior unchanged when tracing is disabled.
- [ ] **Step 3: Implement `Setup` and config loading.** Add `OTEL_TRACES_ENABLED` defaulting to false, `OTEL_SERVICE_NAME` defaulting to `mypocket-api`, and use standard `OTEL_EXPORTER_OTLP_ENDPOINT`, `OTEL_EXPORTER_OTLP_HEADERS`, and timeout/compression settings. Do not store Langfuse secret/public keys in source or send them to the frontend.
- [ ] **Step 4: Wire the provider/run/tool/persistence spans.** Start the root span in `AdvisorService.Submit`, child provider spans in `AdvisorClient.Chat`, allowlisted tool spans around `AdvisorTool.Execute`, and a persistence span around assistant message/event/usage writes. Record errors and close spans on every return path, including cancellation and budget failure.
- [ ] **Step 5: Wire lifecycle shutdown.** Initialize tracing before serving HTTP in `cmd/api/main.go`, defer bounded shutdown, and ensure exporter/collector errors are logged safely without changing the request result.
- [ ] **Step 6: Run the focused and full tests.**

  Run: `cd backend && go test ./internal/infrastructure/observability ./internal/infrastructure/ai ./internal/usecase -run 'Test.*Advisor|Test.*Trace|Test.*Observ' -count=1 && go test ./...`

  Expected: spans are captured in tests, disabled mode makes no network call, and all existing AI/DB tests pass.
- [ ] **Step 7: Commit the fail-open tracing layer.**

  ```bash
  git add backend/go.mod backend/go.sum backend/internal/infrastructure/observability backend/internal/config backend/cmd/api/main.go backend/internal/usecase/ai_advisor.go backend/internal/usecase/ai_advisor_provider.go backend/internal/infrastructure/ai/advisor_client.go docs/operations/ENVIRONMENTS.md docs/architecture/AI_ADVISOR_HARNESS.md
  git commit -m "feat: add fail-open advisor tracing"
  ```

### Task 5: Release gate, staging proof, and public/agent documentation

**Files:**
- Modify: `app/package.json` only if a root release script is needed; do not add Promptfoo to the app bundle.
- Modify: `docs/operations/LOCAL_RELEASE_GATE.md`
- Modify: `docs/operations/RELEASE_CHECKLIST.md`
- Modify: `docs/architecture/API.md`
- Modify: `docs/releases/CHANGELOG.md` at release time
- Create: `docs/releases/REL-<version>-ai-advisor-harness.md` at release time
- Modify: `docs/README.md` and `docs/architecture/AI_ADVISOR_HARNESS.md`

**Interfaces:**
- The release gate exposes three separate commands and evidence classes: Go contract tests, Promptfoo synthetic/live eval, and OTel smoke/collector proof. A passing Go suite cannot be reported as a passing live model evaluation.
- Public API docs describe only the existing advisor routes and safe data behavior. Internal eval routes/fixtures, if later added, remain non-public and API-key scoped.

- [ ] **Step 1: Document the harness contract before rollout.** `AI_ADVISOR_HARNESS.md` must contain architecture, dataset privacy, exact env vars, commands, thresholds, span names/attributes, failure behavior, and a “not shipped” section for future features.
- [ ] **Step 2: Add the local release gate commands.** Require `cd backend && go test ./...`, the generated catalog diff check, `cd evals/finance-assistant && npm run eval:unit`, and existing app design/type/build gates. Keep live eval explicitly staging-only and opt-in.
- [ ] **Step 3: Run staging proof with synthetic data.** Configure `AI_EVAL_LIVE=1` against staging, run all 30 cases with the deployed model, set `OTEL_TRACES_ENABLED=true` against a non-production Langfuse project/collector, verify trace hierarchy and usage, and confirm no raw sensitive data appears in captured spans or artifacts.
- [ ] **Step 4: Add operational rollback/disable proof.** Set `OTEL_TRACES_ENABLED=false` and verify advisor behavior is unchanged; simulate collector outage and confirm request success; disable live eval without changing production runtime.
- [ ] **Step 5: Reconcile public and AI-agent-readable docs.** Update tool catalog, API reference, env/runbook, harness doc, changelog, and a versioned release note with exact commit/image digest, test outputs, known issues, and model/data-sharing controls. Do not claim Langfuse, Promptfoo, or live thresholds are deployed until staging evidence exists.
- [ ] **Step 6: Run final verification and commit docs.**

  Run:

  ```bash
  cd backend && go test ./...
  cd ../app && npm run check:design && npm run typecheck && npm run build
  cd ../evals/finance-assistant && npm run eval:unit
  git diff --check
  ```

  Then record staging/live evidence separately and commit:

  ```bash
  git add docs/operations/LOCAL_RELEASE_GATE.md docs/operations/RELEASE_CHECKLIST.md docs/architecture/API.md docs/architecture/AI_ADVISOR_HARNESS.md docs/architecture/AI_ADVISOR_TOOLS.md docs/README.md docs/releases/CHANGELOG.md docs/releases/REL-<version>-ai-advisor-harness.md
  git commit -m "docs: publish finance advisor quality gates"
  ```

## Dependency and Rollout Order

1. Task 1 publishes the single source of truth for model-facing tools.
2. Task 2 proves Go business/security invariants and remains mandatory even if the provider changes.
3. Task 3 evaluates model routing on synthetic facts; it depends on the generated catalog but not on Langfuse.
4. Task 4 adds optional production traces without changing answer semantics; it can be rolled out disabled first.
5. Task 5 runs all gates in staging, records evidence, and only then updates the release note/deploy artifact.

No production deployment is authorized by this plan until Tasks 1–5 have fresh verification, the live-provider run is explicitly labeled staging, and the owner approves the resulting release note.

## Self-Review Coverage

- Tool catalog drift is caught by the generated JSON diff in Task 1.
- Owner leakage, writes, malformed args, date filters, and budget limits are pinned in Task 2.
- Model trajectory, argument selection, synthetic grounding, and refusal behavior are pinned in Task 3.
- Sensitive telemetry and collector failure are pinned in Task 4.
- Release evidence, rollback/disable behavior, and documentation reconciliation are pinned in Task 5.

## External References

- [Promptfoo custom JavaScript provider](https://www.promptfoo.dev/docs/providers/custom-api/)
- [Promptfoo assertions and tool trajectory checks](https://www.promptfoo.dev/docs/configuration/expected-outputs/)
- [OpenTelemetry Go](https://opentelemetry.io/docs/languages/go/)
- [OpenTelemetry Go exporters and standard environment variables](https://opentelemetry.io/docs/languages/go/exporters/)
- [Langfuse OpenTelemetry/other-language integration](https://langfuse.com/docs/observability/sdk/overview)
