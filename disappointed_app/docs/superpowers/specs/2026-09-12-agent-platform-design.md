# MyPocket Agent Platform Design

Status: ready for owner review
Date: 2026-09-12
Scope: replace the current one-shot Agent MVP with a durable, review-first, API-key-capable assistant platform for MyPocket. This document is design only; implementation starts after owner approval and a separate implementation plan.

## Goal

MyPocket needs an Agent that third-party AI clients and the browser can use safely for finance assistance. The Agent may understand text, use an OCR image tool, read authorized finance context, analyze spending, classify transactions, and prepare actions. It must not mutate wallets, transactions, reports, budgets, debts, recurring schedules, or account state without an explicit confirmation step through MyPocket's deterministic backend.

The product behavior should feel like a finance assistant, not a hidden automation layer. The user or API client sends a message, the backend orchestrates model/tool work, the result is either read-only analysis or one or more typed pending actions, and confirmation applies those actions through the same domain services used by normal UI/API flows.

## Current State

The existing implementation is a useful MVP but not the final Agent platform:

- `POST /api/v1/agent/messages` creates an asynchronous one-shot run for `transaction_draft` or `analysis`.
- `GET /api/v1/agent/runs/{id}` returns the owned run.
- Optional receipt image input is uploaded first and passed as a receipt reference.
- The model adapter calls an OpenAI-compatible `/v1/chat/completions` endpoint with a strict JSON schema.
- The worker validates one proposed transaction and can create a pending `transaction_drafts` row.
- Existing finance confirmation remains the accounting boundary.
- OCR is a third-party image-processing tool; it is not a bank integration and not a generic MyPocket OCR proxy.

The gaps are:

- No durable conversation or message model.
- No general tool loop.
- No multi-action proposal model.
- No transfer/budget/obligation/recurring/event action support from Agent.
- No Agent-specific API-key scopes.
- Limited read context and no search/report tools.
- Browser UI is one-shot rather than a persistent conversation with action review.
- Production config can leave Agent work queued forever when optional AI/OCR capability is disabled instead of exposing capability failure early.

## Non-Goals

- No bank connection, bank sync, bank webhook, or automatic bank import in this phase.
- No autonomous destructive actions. The Agent cannot delete accounts, wallets, categories, transactions, budgets, audit records, API keys, or receipt media.
- No direct SQL, arbitrary HTTP, shell execution, or browser-side provider calls by the model.
- No provider secret in browser bundles, public docs, logs, audit metadata, OpenAPI examples, or database rows.
- No generic OCR proxy endpoint. MyPocket uses the owner's OCR Platform only as a server-side Agent image tool.
- No voice input in this phase.

## Recommended Architecture

Use an application-layer Agent orchestrator inside the existing Go modular monolith.

The HTTP API authenticates cookie or Bearer API-key identity, validates transport shape, enforces CSRF for cookie mutations, creates messages/runs, and returns state. The Agent service owns conversation state, tool planning, action drafting, and confirmation handoff. The worker leases queued runs and performs provider/tool work outside request transactions. Finance, planning, analytics, portfolio, receipt, audit, and identity services remain the only owners of their invariants.

The model receives a constrained system prompt, user message history, selected read context, and tool result summaries. It never receives provider credentials. It may request typed tools through the server orchestrator, but the server decides which tools are allowed for that run, validates every argument, enforces ownership, records tool provenance, and redacts outputs before returning them to the model.

PostgreSQL remains authoritative for conversations, messages, runs, tool runs, actions, confirmations, audit events, and idempotency. Redis is used only for rate limits, short capability/cache hints, and circuit-breaker state. Redis is never the source of truth for Agent decisions.

## Data Model

Add durable tables with user ownership on every row:

- `agent_conversations`: conversation ID, user ID, title, state, created/updated timestamps, last run summary.
- `agent_messages`: conversation ID, user ID, role (`user`, `assistant`, `tool`, `system_summary`), content, attachment references, run ID, timestamps.
- `agent_runs`: run ID, conversation ID, user ID, state, provider metadata, lease fields, attempts, idempotency key, created/updated timestamps, redacted error code.
- `agent_tool_runs`: run ID, conversation ID, user ID, tool name, input metadata, output metadata, state, attempts, provider/upstream reference, timestamps, redacted error.
- `agent_actions`: conversation ID, run ID, user ID, action type, status, typed payload, validation result, confirmation idempotency key, applied entity references, timestamps.

Keep the existing one-shot run table only as a compatibility bridge if migration simplicity requires it. The target model is conversation-first.

## Tools

Tools are server-owned functions, not model-owned plugins.

Read tools:

- `list_wallets`: active and optionally archived user-owned wallets with balance metadata.
- `list_categories`: user-owned categories with type and hierarchy.
- `search_transactions`: bounded search by date, wallet, category, amount, type, and text.
- `get_financial_report`: deterministic report slices for income, expense, net, categories, wallets, and transfer neutrality.
- `list_budgets`: current and historical budgets with progress.
- `list_obligations`: debts, loans, repayments, due dates, and settlement state.
- `list_recurring_schedules`: recurring rules and latest generated draft state.
- `get_portfolio`: portfolio positions and valuation snapshot when enabled.
- `ocr_receipt`: submit and poll an owned receipt/image through the third-party OCR Platform.

Draft tools:

- `draft_transactions`: one or more income/expense transaction proposals.
- `draft_transfer`: transfer proposal that must stay neutral in reports.
- `draft_budget`: budget creation or adjustment proposal.
- `draft_obligation`: debt/loan/repayment proposal.
- `draft_recurring_schedule`: recurring rule proposal.
- `draft_financial_event`: event/tag proposal for future grouping.

Each draft tool creates an `agent_actions` row. It does not apply domain mutations.

## Action Confirmation Contract

Every proposed action has four states: `pending`, `confirmed`, `rejected`, or `expired`.

Confirmation is a backend command, not a model continuation. The command:

1. Re-validates user ownership and current entity versions.
2. Re-validates money, dates, wallet/category compatibility, transfer shape, and business-specific invariants.
3. Uses an idempotency key.
4. Calls the existing finance/planning service.
5. Records applied entity IDs and audit metadata.

If state changed since drafting, confirmation returns a conflict with enough safe context for the user/client to revise. The Agent does not auto-repair and apply the action.

Transfers must produce balanced source/destination ledger effects and remain excluded from income/expense reports. Reports and analytics must continue deriving from confirmed ledger state only, never pending Agent actions.

## API Design

Add the conversation-first Agent API while preserving current MVP endpoints as compatibility adapters.

- `POST /api/v1/agent/conversations`: create a conversation.
- `GET /api/v1/agent/conversations`: list owned conversations.
- `GET /api/v1/agent/conversations/{conversation_id}`: return messages, runs, tool summaries, and pending actions.
- `POST /api/v1/agent/conversations/{conversation_id}/messages`: append a user message, optional attachment reference, and enqueue a run.
- `GET /api/v1/agent/runs/{run_id}`: return run state and safe result metadata.
- `POST /api/v1/agent/runs/{run_id}/cancel`: cancel queued/processing work when possible.
- `GET /api/v1/agent/actions`: list pending/recent owned actions.
- `GET /api/v1/agent/actions/{action_id}`: return one owned action.
- `POST /api/v1/agent/actions/{action_id}/confirm`: apply a pending action through deterministic services.
- `POST /api/v1/agent/actions/{action_id}/reject`: reject a pending action.

Compatibility:

- `POST /api/v1/agent/messages` creates or reuses a short-lived conversation and maps the result to current run shape.
- `GET /api/v1/agent/runs/{id}` remains stable for old clients.

Stable errors include `AUTH_REQUIRED`, `FORBIDDEN`, `VALIDATION_FAILED`, `CAPABILITY_UNAVAILABLE`, `IDEMPOTENCY_CONFLICT`, `CONFLICT`, `NOT_FOUND`, `INTERNAL_RETRYABLE`, and `INTERNAL_FAILURE`.

## API-Key Authorization

API keys must support Agent use without granting broad account control.

Add scopes:

- `finance:read`: permits read tools and report context.
- `agent:use`: permits conversations, messages, runs, OCR tool initiation through Agent, and read-only analysis.
- `agent:confirm`: permits confirming/rejecting Agent actions when paired with the required domain write scope.
- `finance:write`: permits applying confirmed finance mutations through existing deterministic APIs.

Rules:

- Browser sessions can manage API keys; API keys cannot mint or rotate other API keys.
- Invalid Bearer credentials never fall back to cookies in the same request.
- Existing unscoped keys keep backward compatibility through a migration-defined legacy policy, but new keys should be scope-explicit.
- Each Agent endpoint must be represented in the authorization matrix for cookie, valid key, insufficient scope, revoked key, and foreign ownership.

## OCR Tool Boundary

OCR remains a third-party image preprocessing tool for Agent runs.

The browser uploads an owned image/receipt through existing private media flow. The backend verifies ownership, size, content type, checksum, and retention state before submitting bytes to the OCR Platform. The OCR provider response is stored as untrusted tool output linked to the run. The model may use it to prepare analysis or draft actions, but OCR output alone never creates a confirmed transaction.

Runtime behavior:

- If OCR is disabled or misconfigured, image Agent requests fail with `CAPABILITY_UNAVAILABLE`.
- If OCR provider is temporarily unavailable, the run enters retryable failure state with redacted metadata.
- OCR credentials and raw provider payloads are never exposed to the browser, OpenAPI examples, Docusaurus pages, logs, or audit events.

## Capability and Config Behavior

Expose Agent capabilities explicitly so clients do not enqueue work that cannot finish.

Backend startup and health:

- Production fails fast if `MYPOCKET_AI_ENABLED=true` but base URL, API key, model, timeout, or HTTPS requirements are invalid.
- Production fails fast if `MYPOCKET_OCR_ENABLED=true` but OCR URL/key or safe timeout/polling limits are invalid.
- If AI is disabled, Agent message creation returns `CAPABILITY_UNAVAILABLE` instead of accepting a permanently queued run.
- If OCR is disabled, text-only Agent analysis can still work when AI is enabled; image requests fail early.
- Health/capability metadata is safe and redacted.

Also reconcile production S3 environment variable names so API and worker use the same documented names.

## Frontend Design

Replace the one-shot Agent screen with a persistent conversation workspace built from base components.

Required components:

- conversation list
- conversation thread
- message bubble
- composer
- attachment picker
- run status
- tool activity row
- action card
- action editor
- confirmation/rejection controls
- empty/error/offline states

Behavior:

- The UI shows pending actions inline in the conversation.
- Users can edit safe fields before confirmation.
- Confirmation uses deterministic action endpoints, not a free-form model reply.
- Text/image input remains usable on mobile and PWA.
- Offline mode preserves draft input but does not pretend provider work can run offline.
- The old one-shot visual flow can be removed only after the conversation UI has equivalent or better test coverage.

## Public Documentation

Before and after implementation, update:

- internal architecture docs: Agent architecture, ERD, API, integrations, requirements, release notes
- public Docusaurus docs: Agent guide, image/OCR behavior, API authentication/scopes, endpoint reference, errors, examples
- machine-readable surfaces: OpenAPI JSON, endpoint catalog if present, `llms.txt`, `llms-full.txt`, and the reusable MyPocket API skill

Docs must state:

- Agent is review-first.
- API keys can use Agent endpoints when scoped correctly.
- OCR is a third-party tool, not a MyPocket generic OCR API.
- Bank integration is out of scope.
- Pending Agent actions do not affect reports/balances until confirmed.
- Provider secrets are configured server-side only.

## Security and Audit

Threats to cover:

- prompt injection through user text, transaction notes, category names, receipt text, and OCR output
- provider hallucination of wallet/category/action IDs
- cross-user data leakage
- API-key scope bypass
- stale or duplicate confirmations
- receipt/object ownership bypass
- provider secret leakage
- raw prompt/provider payload leakage in audit/logs
- replayed idempotency keys

Mitigations:

- All tool inputs are typed and server-validated.
- All tool outputs are treated as untrusted context.
- User ownership is checked at every read, tool run, draft, and confirmation.
- Raw provider prompts/responses are not audit payloads.
- Audit records include actor, auth mode, endpoint/action, result, target IDs, correlation ID, and redacted metadata.
- Redis rate limits apply to Agent message creation, tool runs, and confirmations; PostgreSQL audit remains durable.

## Verification Plan

Backend unit tests:

- model schema rejection
- tool argument validation
- multi-action draft validation
- transfer neutrality
- stale confirmation conflict
- capability-unavailable behavior

Backend integration tests:

- two-user isolation for conversations, messages, runs, tools, and actions
- cookie vs API-key auth
- insufficient scope rejection
- revoked key rejection
- idempotent message replay and confirmation replay
- OCR provider success, timeout, quota, expired result, and capability drift through fakes
- reports exclude pending actions and transfers

Frontend tests:

- conversation rendering
- message submit and polling
- image attachment state
- pending action edit/confirm/reject
- offline preserved composer state
- mobile layout

E2E tests:

- create wallet, add/edit/delete transaction, transfer between wallets, confirm Agent transaction, confirm Agent transfer, verify report neutrality
- Agent receipt image to pending draft to explicit confirmation
- API-key client creates conversation, sends message, reads action, confirms with required scopes

Release gates:

- backend full integration/race/vet
- frontend unit/type/build
- multi-browser E2E including mobile WebKit
- Docusaurus build
- OpenAPI contract drift test
- reusable skill validation
- production-shaped AI/OCR smoke with protected secrets
- physical iPhone/Safari UAT for PWA and image input
- backup/restore and rollback evidence

## Implementation Sequence

1. Write and approve this spec.
2. Write a detailed implementation plan.
3. Add conversation/message/action schema and repositories.
4. Add capability checks and fail-closed Agent/OCR behavior.
5. Add API-key scopes and authorization matrix coverage.
6. Add tool orchestration with deterministic fake provider tests.
7. Add read tools.
8. Add action drafting and confirmation engine.
9. Add transaction and transfer actions first.
10. Add budget, obligation, recurring schedule, and event actions.
11. Add OCR tool integration into the conversation model.
12. Replace the frontend Agent screen with conversation/action UI.
13. Update OpenAPI, public docs, `llms` surfaces, and skill.
14. Run full verification and only then prepare deployment.

## Open Questions for Owner Review

- Should legacy unscoped API keys retain Agent access during migration, or should Agent require explicit new scopes from day one?
- Should Agent conversations expire/archive automatically, or remain user-managed history?
- Should the first implementation include all draft action types, or should transaction and transfer actions ship first behind the same architecture?

Default recommendation: require explicit scopes for new keys, keep old keys compatible only for already-documented non-Agent endpoints, keep conversations user-managed, and implement transaction/transfer first before expanding action types.
