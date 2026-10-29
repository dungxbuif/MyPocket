---
artifact_type: detail_design
status: approved
owner: shared
approval: owner pre-approved autonomous implementation in active goal
links:
  backlog: ../../work/BACKLOG.md
  context: ../../CONTEXT.md
  validation_matrix: ../../work/VALIDATION_MATRIX.md
  api_docs: ../../architecture/API.md
  erd_docs: ../../architecture/ERD.md
---

# Budget, Recurring, and Agent Completion Design

## Scope

This design implements the owner-approved B/C/D completion slice:

- B. Budget assignment is transaction-level.
- C. Recurring schedules support edit, pause/resume, optional end date, and optional auto-post.
- D. Agent chat has durable history and action cards for transaction drafts.
- Public/internal docs and OpenAPI are reconciled after implementation.

Out of scope:

- A. Wallet behavior model cleanup (`basic`/`goal`/`credit`, removing pseudo wallet types) is intentionally deferred.
- Bank webhook or bank integration.
- UAT and production deployment.

## Problem

The current planning model still treats budgets as category-scope aggregations. That cannot represent owner-requested "hũ chi tiêu" behavior where a user chooses which confirmed expenses belong to a budget. Recurring schedules can create reviewable drafts but cannot be edited, paused, ended, or auto-posted. The current Agent API has two chat surfaces but persists only runs, so it cannot show a durable chat history or action cards independent of polling a run.

## Design decisions

### Budget assignment

Budgets remain reusable limits with period metadata and optional category hints, but spending is counted only from confirmed expense transactions explicitly assigned to the budget.

- Add nullable `budget_id` to `transactions`.
- Add nullable `budget_id` to `transaction_drafts` and `recurring_schedules` so scheduled/Agent proposals can carry an assignment until confirmation.
- `budget_id` is valid only for expense transactions and pending expense drafts.
- The referenced budget must belong to the authenticated user and be active.
- Transfers, income and adjustments reject `budget_id`.
- Budget progress sums active, report-included expense transactions where `transactions.budget_id = budgets.id` and `occurred_at` falls within the current budget period.
- Existing `budget_categories` is retained as category hints/filtering metadata. It no longer defines progress.

Rejected alternatives:

- Keep category-derived progress and add UI-only budget tagging. Rejected because overlapping budgets would still double-count and would not satisfy transaction-level assignment.
- Replace budgets with wallet-like jars. Rejected for this slice because the owner decided the feature belongs in budgets, while wallet behavior redesign is deferred.

### Recurring schedules

Recurring schedules remain review-first by default and gain optional controlled automation.

- Add `posting_mode`: `draft` or `auto_post`, default `draft`.
- Add `paused_at`: non-null means the worker skips the schedule.
- Add `ends_at`: optional stop timestamp. A due occurrence after `ends_at` is not created.
- Add update endpoint: `PATCH /api/v1/recurring-schedules/{id}` with `base_version`.
- Add pause/resume endpoints: `POST /api/v1/recurring-schedules/{id}/pause` and `/resume`, both versioned.
- Draft mode creates `transaction_drafts` as today, including any `budget_id`.
- Auto-post mode creates a confirmed finance transaction directly using an occurrence-derived idempotency key and records the occurrence once.
- Worker catch-up remains bounded per run.

Rejected alternatives:

- Only archive-and-create-new. Rejected because the owner explicitly asked for edit/pause/end.
- Auto-post as default. Rejected because review-first remains safer; auto-post must be explicit.

### Agent durable chat

Agent has two persisted singleton session types per user:

- `intake`: text/image to reviewable transaction draft actions only.
- `advisor`: read-only financial Q&A only.

Data model:

- `agent_sessions`: user, kind, title, context summary, status, timestamps.
- `agent_messages`: session, role, text, optional run, optional action payload, timestamps.
- `agent_runs.session_id`: optional link for compatibility and worker correlation.

API behavior:

- `POST /api/v1/agent/intakes` and `/agent/intakes/{session_id}/messages` append a user message, queue a run, and return the session plus latest message/run.
- `POST /api/v1/agent/advisor/messages` appends to the advisor singleton session and returns the session plus latest message/run.
- `GET /api/v1/agent/intakes/{session_id}` and a new `GET /api/v1/agent/advisor` return bounded message history.
- Existing `/api/v1/agent/messages` and `/api/v1/agent/runs/{id}` remain compatibility routes.
- On run completion, the worker appends an assistant message. If drafts were created, the message includes an action card payload containing `draft_ids`.

Context management:

- Model prompts include authoritative wallet/category references, recent aggregates, OCR results for the run, the session summary when present, and a bounded recent-message window.
- The first implementation does not need a separate summarizer worker; the schema reserves `context_summary` for future compaction without changing public routes.

Rejected alternatives:

- Store only run rows and reconstruct a thread in the UI. Rejected because a run is not a chat session and cannot support stable history/action cards.
- Allow advisor write tools. Rejected because the user requested a separate read-only financial chat.

## Touched surfaces

- Database migrations: transactions, recurring schedules, drafts, agent sessions/messages.
- Backend finance domain/repository/API request and response shapes.
- Backend planning domain/repository/API/worker behavior.
- Backend agent repository/service/worker/API behavior.
- Frontend API clients and UI surfaces for transaction budget selection, recurring edit/pause/resume/auto-post, and Agent chat cards.
- Public docs: API pages, ERD, skill docs, OpenAPI route list, guides.
- Internal docs: context, backlog, validation matrix, changelog, evidence record.

## Security and business invariants

- Every budget/session/message lookup is scoped by authenticated `user_id`.
- Bearer API keys must support the same business endpoints without CSRF where existing middleware permits API-key business access.
- Cross-user budget IDs, schedule IDs, session IDs, run IDs, wallet IDs and category IDs must not mutate or disclose data.
- Pending drafts and Agent action cards do not alter wallet balances.
- Auto-post recurring transactions must be idempotent per occurrence.
- Confirming the same draft remains exactly-once accounting.
- Transfers remain omitted from budget progress and reports.

## Verification plan

RED/GREEN implementation will add or update tests in this order:

1. Backend integration: transaction-level `budget_id` progress, ownership rejection, transfer/income rejection, edit/archive progress reversal, draft confirmation carries budget assignment.
2. Backend integration: recurring update/pause/resume/end/auto-post with bounded/idempotent worker behavior.
3. Backend integration/API: Agent session/message persistence, intake action cards, advisor read-only behavior, compatibility route behavior, cross-user isolation.
4. Frontend unit tests: API clients and UI states for budget assignment, recurring controls, and chat history/action cards.
5. Browser E2E: create wallet/category/budget, create/edit/archive assigned expense, transfer excluded from reports/budget, recurring draft and auto-post, Agent intake action card, advisor read-only.
6. Full relevant regression ladder before handoff: backend `go test -p 1 ./...`, frontend tests/build, Docusaurus build, and required Playwright business suites.

## Docs reconciliation plan

After code matches tests:

- `docs/architecture/API.md`: update live route table and endpoint payloads.
- `docs/architecture/ERD.md`: add new columns/tables/relationships.
- `frontend/docs/docs/api/*.mdx`: update budgets, planning, transactions, agent.
- `frontend/docs/docs/erd/*.mdx`: update ERD pages.
- `skills/mypocket-api/SKILL.md` and references: switch to current two-chat/session/action-card API.
- `docs/work/VALIDATION_MATRIX.md`, `docs/CONTEXT.md`, `docs/work/BACKLOG.md`, `docs/releases/CHANGELOG.md`: record local-only proof and remaining UAT/deploy exclusion.

No ADR is required unless implementation changes the approved security model or introduces a new external provider boundary.
