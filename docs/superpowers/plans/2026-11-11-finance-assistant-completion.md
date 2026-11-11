# Finance Assistant Completion Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Complete the external Finance Assistant with safe model-selected read tools and verified recovery while preserving owner isolation, memory and channel-neutral delivery.

**Architecture:** Keep the approved Python, LangGraph, LangMem and PostgreSQL service boundary. Add a bounded model/tool loop whose executor owns credentials and accepts only a fixed read catalog; native rich parts and Telegram text continue to derive from the same result. PostgreSQL remains authoritative and Markdown files remain reviewable projections.

**Tech Stack:** Python 3.12, FastAPI, LangGraph 1.0.1, LangChain OpenAI, LangMem, PostgreSQL/psycopg, pytest.

**Spec:** `docs/work/tickets/FINANCE-ASSISTANT-COMPLETION-AUDIT-20261111.md`

## Global Constraints

- MyPocket is the only authentication and ledger authority.
- The assistant receives only owner-scoped `mpa_` delegation; no full owner API key is stored.
- Expose only `home`, `wallets`, `transactions`, `budgets`, and `jars`; no write tool exists.
- App and adapter transcripts stay separate while trusted memory is owner-shared.
- Tool rounds and returned prompt payloads are bounded.
- Production release uses one owner-created `prod-<semver>` tag and the frozen homelab receipt pipeline.

## Review Focus

- A model requests an unknown or write operation: reject it without an HTTP request.
- A tool call supplies owner ID, token, path, or unbounded pagination: ignore identity fields and validate only operation-specific public parameters.
- A model repeatedly requests tools: stop at the configured maximum and return a safe answer.
- A read endpoint fails: expose a concise unavailable result to the model without leaking credentials.
- Multiple tool calls generate rich parts: preserve deterministic ordering and channel-neutral DTOs.

---

### Task 1: Read-tool catalog and executor

**Files:**
- Modify: `finance-assistant/src/finance_assistant/mypocket.py`
- Modify: `finance-assistant/tests/test_mypocket.py`

**Interfaces:**
- Produces: `FinanceToolCall`, `FINANCE_READ_TOOLS`, and `LiveFinanceContext.execute(owner_id, calls, delegated_credential)`.
- Consumes: existing `MyPocketClient.read` and credential vault.

- [ ] Add failing tests for catalog names, exact allow-list, safe parameter bounds, unknown/write rejection, missing delegation and deterministic rich parts.
- [ ] Run `python -m pytest -q finance-assistant/tests/test_mypocket.py` and confirm failure.
- [ ] Implement immutable tool metadata, validated call arguments and the executor; keep `MyPocketClient` read-only.
- [ ] Re-run the focused tests and confirm pass.

### Task 2: Bounded LangGraph model/tool loop

**Files:**
- Modify: `finance-assistant/src/finance_assistant/agent.py`
- Modify: `finance-assistant/tests/test_api.py`

**Interfaces:**
- Consumes: Task 1 executor and catalog.
- Produces: model protocol with a structured tool-decision response and an `AgentRuntime` loop capped at three tool rounds.

- [ ] Add failing tests for no-tool chat, one/multiple tool calls, repeated-call cap, read failure and owner/delegation propagation.
- [ ] Run the focused tests and confirm failure.
- [ ] Implement model tool binding/decision normalization and the bounded graph node loop; never include owner or credential in model-visible arguments.
- [ ] Re-run focused tests and confirm pass.

### Task 3: Capability and compatibility contract

**Files:**
- Modify: `finance-assistant/src/finance_assistant/api.py`
- Modify: `finance-assistant/tests/test_api.py`
- Modify: `mobile-ios/MyPocketIOS/Features/Assistant/AssistantModels.swift` only if current decoding rejects the catalog.

**Interfaces:**
- Consumes: Task 1 public catalog.
- Produces: truthful `/api/v1/ai/advisor/capabilities` read-tool names with `write=false` and `direct_writes=false`.

- [ ] Change the capability test to expect the five read tools and retained write prohibitions; run it and confirm failure.
- [ ] Return the stable catalog without changing the native envelope.
- [ ] Run API and native model tests and confirm pass.

### Task 4: Recovery and complete regression proof

**Files:**
- Modify: `finance-assistant/scripts/backup.sh` or `restore.sh` only if the isolated drill exposes a defect.
- Modify: `finance-assistant/tests/test_postgres_integration.py` only for missing automated recovery assertions.

**Interfaces:**
- Consumes: existing backup/restore scripts and PostgreSQL schema.
- Produces: reproducible evidence that database authority and workspace projection survive restore.

- [ ] Run all Finance Assistant tests including the PostgreSQL integration test.
- [ ] Run an isolated disposable backup/restore drill covering one memory, separate app/Telegram history, owner isolation and workspace archive permissions.
- [ ] If a defect appears, add a failing regression test, patch minimally and repeat the drill.
- [ ] Run backend proxy tests and the repository verification gates.

### Task 5: Documentation, release and E2E

**Files:**
- Modify: `docs/architecture/FINANCE_ASSISTANT_API.md`
- Modify: `docs/guides/finance-assistant-ios.md`
- Modify: `docs/operations/FINANCE_ASSISTANT_SERVICE.md`
- Modify: `docs/work/VALIDATION_MATRIX.md`
- Modify: `docs/releases/CHANGELOG.md`
- Modify: `docs/work/tickets/FINANCE-ASSISTANT-COMPLETION-AUDIT-20261111.md`

**Interfaces:**
- Consumes: verified outputs from Tasks 1-4.
- Produces: human and AI-readable shipped contract plus release evidence.

- [ ] Reconcile every claim with executed evidence and leave unrun production/user gates explicit.
- [ ] Commit and merge reviewed changes to `main`; verify runner/protected secret/tag rules.
- [ ] Create exactly one new `prod-<semver>` tag and observe the app and ops receipt pipelines.
- [ ] Verify native app chat, Telegram tool-selected read, memory recall, transcript separation, service health and unrelated stack health without logging secrets or financial data.
- [ ] Record the remaining user-only action to rotate the disclosed Telegram token; do not claim it complete until the owner replaces it.

