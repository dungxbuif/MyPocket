# AI Assistant RAG & Intelligence Upgrade Implementation Plan

## Scope correction after owner clarification — 2026-10-20

The owner wants a more capable assistant that understands the application.
This document is a proposal, not approval to build retrieval infrastructure.
Tasks 2–3 (new ingestion/retrieval infrastructure) are deferred. Task 8 applies
to the maintained iOS app only; archived web implementation is out of scope.
Do not introduce vector storage, embeddings, external crawling or a keyword
router that forces tool selection as a prerequisite for improving the assistant.

Immediate sequence:

1. Repair confirmed request lifecycle and action correctness defects, with
   regression tests: double submit, uncertain submit identity, stale history,
   cancellation, jar preservation and mutation replay.
2. Reconcile the compact app handbook with the actual tool registry and native
   navigation. Include account-local current time and precise tool descriptions.
   Keep existing bounded product-help lookup; finance queries use live tools.
3. Measure real-model routing and multi-turn behavior on synthetic fixtures.
   Provider rejection is a failed availability gate, never a passing quality test.
4. Extend missing finance capabilities only against explicit domain contracts.
   Revisit retrieval infrastructure if measured failures show missing knowledge
   cannot fit in the compact handbook or existing help lookup.

The remainder records candidate future work, not a requirement to implement
all RAG infrastructure. Acceptance is useful, correct in-app assistance.

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make Finance Assistant materially more useful and reliable by upgrading the provider to `DeepSeek-V4.1-Flash`, adding release-aware grounded retrieval, deterministic finance intent routing, stronger context/tool orchestration, and real model quality gates without weakening owner isolation or confirmation safety.

**Architecture:** Keep Go as the only orchestrator and source of financial truth. Product knowledge is indexed into a versioned, signed/validated knowledge manifest and retrieved by a bounded hybrid retriever; financial facts continue to come only from owner-scoped finance tools. The model receives compact retrieved evidence with stable citations, while all writes remain proposal-plus-owner-confirmation actions. Evaluation and observability are separate, fail-closed for quality/security but fail-open for telemetry.

**Tech Stack:** Go, PostgreSQL existing repositories/migrations, OpenAI-compatible chat completions, deterministic Vietnamese tokenizer/ranker, optional embedding provider behind a disabled-by-default interface, Promptfoo HTTP evals, Go contract tests, OpenTelemetry-compatible spans, Markdown/JSON agent-readable docs.

**Spec:** `docs/work/tickets/FEEDBACK-20261019-PRODUCT-HELP-RAG-CONTRACT.md`, `docs/architecture/AI_ADVISOR_TOOLS.md`, `docs/work/tickets/AI-ADVISOR-01-DETAIL_DESIGN.md`, and this plan.

## Global Constraints

- The authenticated `Principal.OwnerID` is always server-supplied; no model argument may select an owner.
- Finance Assistant may read through owner-scoped tools and prepare bounded proposals; it may not directly mutate a ledger, pay, bank-transfer, or delete without explicit owner confirmation.
- Retrieved text is untrusted evidence, never executable instructions; it cannot alter tool policy, system prompt, permissions, or provider configuration.
- Product retrieval must not index transactions, amounts, notes, tokens, API keys, private attachments, or arbitrary external URLs.
- Model and provider changes are backend environment changes only; no secret enters web/iOS bundles or Git.
- Every answer containing a financial number must be traceable to a current tool result; retrieved product docs cannot supply financial facts.
- Existing request/tool/provider bounds remain: six provider requests, eight tool executions, bounded body/response sizes, timeout and idempotent run semantics.
- Production rollout requires image, backup, migration/readback, worker health, provider smoke, and route/schema parity evidence; local tests alone do not authorize production.

## Review Focus

- Vietnamese paraphrases, diacritics, abbreviations, and mixed Vietnamese/English feature names must retrieve the same authoritative section.
- A stale release note must not override a current product contract; conflicting sources must expose source/version and prefer the current contract.
- Prompt injection inside a retrieved document must be rendered as inert evidence and never change tool access.
- A financial question with an explicit period/wallet/category must route to the correct finance tool rather than product-help retrieval or a free-text guess.
- A provider returning malformed tool calls, unsupported tools, huge usage metadata, or an answer with ungrounded amounts must fail safely and leave no write.

---

### Task 1: Model configuration and provider contract

**Files:**
- Modify: `backend/.env.local` (ignored runtime config) — set `AI_MODEL=DeepSeek-V4.1-Flash`.
- Modify: `backend/.env.example` — document the model override without adding credentials.
- Modify: `backend/internal/infrastructure/ai/advisor_client.go` and `client.go` only if the provider smoke reveals wire incompatibility.
- Test: `backend/internal/infrastructure/ai/advisor_client_test.go`, `provider_config_test.go`.
- Modify: `docs/architecture/INTEGRATIONS.md`, `docs/operations/PROVIDER_UAT.md`, `docs/releases/CHANGELOG.md`.

**Interfaces:** Preserve `AdvisorProvider.Chat` and AI-entry extraction contracts. Add a provider capability probe only if required by the model, never a frontend model selector.

- [ ] Write a test that asserts the configured model is passed exactly as `DeepSeek-V4.1-Flash` in an OpenAI-compatible request and that tool schemas remain present.
- [ ] Run the provider wire test and an opt-in bounded live smoke using the ignored environment; record model name, latency, status and schema outcome without recording the key or prompt.
- [ ] Set local runtime configuration and reconcile docs; do not claim production is upgraded until the deployment environment is read back.

### Task 2: Versioned knowledge source manifest and ingestion

**Files:**
- Create: `backend/internal/knowledge/manifest.go` — `KnowledgeDocument`, `KnowledgeChunk`, `KnowledgeManifest`, validation and stable IDs.
- Create: `backend/internal/knowledge/manifest_test.go`.
- Create: `backend/internal/knowledge/source/*.md` or generated checked-in JSON for product screens, API behavior, AI help, release notes and troubleshooting.
- Create: `scripts/knowledge-validate.mjs` and `docs/architecture/AI_KNOWLEDGE_SOURCES.md`.
- Modify: `docs/work/tickets/FEEDBACK-20261019-PRODUCT-HELP-RAG-CONTRACT.md` and `docs/architecture/AI_ADVISOR_TOOLS.md`.

**Interfaces:** `LoadManifest([]byte) (KnowledgeManifest, error)`, `Validate() error`, and stable source metadata: `id`, `kind`, `title`, `text`, `productVersion`, `minRelease`, `maxRelease`, `owner`, `reviewedAt`, `contentHash`.

- [ ] Add failing tests for missing IDs, duplicate IDs, unknown owners, invalid release ranges, oversized chunks, stale review dates and prompt-injection markers treated as text rather than instructions.
- [ ] Implement strict manifest validation and deterministic content hashing; reject malformed manifests atomically.
- [ ] Populate only authoritative, source-controlled docs; mark release notes as secondary and require a link to a current contract before they can answer behavior questions.
- [ ] Validate with Go and script tests; document refresh/rollback and active manifest version.

### Task 3: Hybrid product retrieval with citations

**Files:**
- Create: `backend/internal/knowledge/retriever.go` — token/phrase/alias ranking, source-kind and release filters, bounded top-k.
- Create: `backend/internal/knowledge/retriever_test.go`.
- Modify: `backend/internal/usecase/ai_advisor_help.go` to call the retriever while preserving the `get_product_help` tool schema and response view.
- Modify: `backend/internal/usecase/ai_advisor_context.go` to include only bounded retrieved evidence with citation IDs.

**Interfaces:** `Retriever.Search(ctx context.Context, query string, options SearchOptions) ([]Evidence, error)`, where `Evidence` carries source ID, title, excerpt, version and confidence; no owner ID or finance rows are accepted.

- [ ] Add tests for diacritic-insensitive Vietnamese, aliases (`hũ/jar`, `ví/wallet`), phrase boosts, release preference, deterministic tie breaks, max 5 results and no-match abstention.
- [ ] Implement lexical BM25-like scoring plus exact phrase/alias boosts first; keep embedding retrieval behind an interface and feature flag, disabled until a separately measured provider is approved.
- [ ] Return stable citations and an explicit `not_found`/`conflict` state instead of dumping the corpus.
- [ ] Add injection tests proving retrieved text cannot add tools, change system role, or introduce URLs/owner data.

### Task 4: Finance intent router and context builder

**Files:**
- Create: `backend/internal/usecase/ai_intent.go` — deterministic intent hints, date-period normalization, tool allowlist and ambiguity state.
- Create: `backend/internal/usecase/ai_intent_test.go`.
- Modify: `backend/internal/usecase/ai_advisor_context.go`, `ai_advisor_provider.go`, `ai_advisor_tools.go`.
- Modify: `docs/architecture/AI_ADVISOR_TOOLS.md`.

**Interfaces:** `ClassifyAdvisorIntent(message string, timezone string) IntentHints`; `BuildAdvisorRequestContext(history, hints, evidence) []AdvisorChatMessage`.

- [ ] Add tests for “tháng 9”, explicit date ranges, wallet/category/merchant filters, comparison questions, help/navigation questions and unsupported banking/payment requests.
- [ ] Implement intent hints as constraints, not an autonomous second model: finance intents force current finance tools; product intents force `get_product_help`; ambiguous requests ask one focused clarification.
- [ ] Preserve at most the configured recent conversation window, add a compact summary only for resolved context, and never treat old cards as current ledger truth.
- [ ] Add regression tests that a prompt injection or user-supplied `owner_id` cannot change the route or scope.

### Task 5: Stronger grounded multi-tool answer contract

**Files:**
- Modify: `backend/internal/usecase/ai_advisor_provider.go` and `ai_advisor_parts.go`.
- Create: `backend/internal/usecase/ai_grounding.go` and tests.
- Modify: `backend/internal/entity/ai_advisor.go` / repository only if additive provenance fields are needed.
- Modify: native/web structured response renderers only for additive citation/provenance display.

**Interfaces:** `ValidateAdvisorGrounding(answer AdvisorAnswer, results []entity.FinanceResult) error`; each numeric claim must map to a result provenance key, while product answers must expose source IDs/version.

- [ ] Add failing tests for ungrounded amounts, wrong-period claims, stale tool results, duplicate proposal actions, unsupported tool calls and malformed structured parts.
- [ ] Implement server-side grounding metadata and safe fallback text; do not use an LLM judge to decide money correctness.
- [ ] Keep proposal cards review-first, owner/generation-bound and idempotent; confirmation remains the only mutation path.
- [ ] Add multi-step tests for compare → category breakdown → merchant drill-down and affordance questions with bounded tool calls.

### Task 6: Usage normalization and privacy-safe observability

**Files:**
- Create: `backend/internal/infrastructure/ai/usage.go` and tests.
- Create/modify: `backend/internal/infrastructure/observability/*` for fail-open OpenTelemetry spans.
- Modify: `backend/internal/infrastructure/ai/advisor_client.go`, `ai_advisor.go`, repository usage persistence.
- Create: `docs/architecture/AI_ADVISOR_HARNESS.md`.

**Interfaces:** `NormalizeAdvisorUsage(raw map[string]any) (map[string]any, error)` with allowlisted fields, byte cap, `prompt_version`, `schema_version`; `StartAdvisorTrace`/child spans with redacted attributes.

- [ ] Add tests for unknown/nested/oversized usage, malformed token values, provider request IDs and schema/prompt version presence.
- [ ] Store only bounded normalized usage; never persist prompt, OCR text, notes, amounts, API keys, URLs or raw provider bodies.
- [ ] Add spans `advisor.run`, `provider.chat`, `tool.<name>`, `answer.persist`; collector outage must not fail a user request.
- [ ] Document Langfuse/OTel setup as optional and fail-open, plus retention and redaction rules.

### Task 7: Promptfoo and Go quality harness

**Files:**
- Create: `evals/finance-assistant/promptfooconfig.yaml`.
- Create: `evals/finance-assistant/datasets/{vietnamese,finance-grounding,security-boundary,multi-turn}.yaml`.
- Create: `evals/finance-assistant/assertions/{tool-routing,grounding,response-shape}.js`.
- Create: `backend/internal/usecase/ai_advisor_harness_test.go` and synthetic fixtures.
- Modify: `evals/ai/README.md`, `docs/architecture/AI_ADVISOR_HARNESS.md`, validation matrix.

**Interfaces:** HTTP provider adapter uses staging/synthetic database only; fixture oracle, never an LLM judge, decides owner scope, tool args, money/date correctness and card schema.

- [ ] Add at least 30 synthetic Vietnamese cases covering reads, comparisons, drill-down, proposals, refusal, injection, multi-turn and missing data.
- [ ] Gate owner isolation, forbidden writes, JSON/schema, numeric grounding and response-card validity at 100%; tool-routing target ≥95%; style judge remains advisory.
- [ ] Run deterministic Go harness on every change; run Promptfoo/live model only when explicitly enabled and never against production data.
- [ ] Record model version, prompt/schema versions, case IDs and thresholds in machine-readable output.

### Task 8: Native/web assistant UX and docs reconciliation

**Files:**
- Modify: `mobile-ios/MyPocketIOS/Features/Assistant/AssistantViews.swift` and shared card bases only for citations, ambiguity prompts, and non-jarring loading states.
- Modify: `app/src/atomic/organisms/FinanceAssistantPanel.tsx` and structured part registry additively.
- Test: native assistant unit/UI tests and web AI tests.
- Modify: `docs/guides/finance-assistant-ios.md`, `docs/README.md`, `docs/releases/CHANGELOG.md`, `docs/CONTEXT.md`, `docs/work/BACKLOG.md`.

- [ ] Add tests for citation tap/readability, clarification cards, grounded numeric cards, stale proposal state, offline cached history and provider error recovery.
- [ ] Keep one shared semantic component contract per card type; do not duplicate screen-local cards.
- [ ] Update public Markdown and AI-readable OpenAPI/catalog surfaces only after implementation and verification.

### Task 9: Release verification and production gate

**Files:**
- Modify: `docs/operations/ROLLOUT_READINESS.md`, `docs/releases/CHANGELOG.md`, validation matrix.
- Run: backend tests, Promptfoo synthetic gate, web checks/build, native tests, image build/preflight, provider smoke.

- [ ] Verify local image contains the model configuration contract and all required migrations.
- [ ] Verify staging route/schema parity, backup/readback, migration version, worker/Redis audit health and authenticated UAT before any production service update.
- [ ] Read back the live production model/version after rollout; if manager/backup access is absent, stop at pushed immutable image and record the exact gate rather than claiming deploy.

## Dependency order

Tasks 1–3 establish model/source/retrieval contracts. Task 4 depends on 3. Task 5 depends on 4. Tasks 6–7 can proceed in parallel after the response contract is fixed. Task 8 depends on 5. Task 9 depends on all prior tasks and external staging/production gates.

## Self-review

- Coverage: provider upgrade, RAG corpus/refresh, retrieval quality, finance intelligence, grounding, usage privacy, evals, UI, docs and release gates are each assigned a task.
- No task grants autonomous writes, adds a user usage limit, exposes credentials, or treats RAG as a source of financial truth.
- The plan deliberately starts with deterministic hybrid retrieval; embeddings are an optional measured layer, not an unbounded dependency or a shortcut around source governance.

