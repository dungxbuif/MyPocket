---
artifact_type: architectural_spec
id: FEEDBACK-FIX-CHANGELOG-01
status: approved_for_spec_review
owner: shared
created: 2026-09-22
---

# Feedback → Fix → Changelog

## 1. Intent

MyPocket needs a runtime feedback loop that is independent from the AI chat. A user can report a bug, feature request, or improvement; a local/dev agent can fetch actionable feedback; a completed fix can be linked to a versioned changelog entry; and the user can see the resulting release directly on their feedback item.

The existing `docs/work/FEEDBACK_LOG.md` remains the human/agent intake process. This subsystem adds persistent, owner-scoped runtime feedback and a narrow internal agent boundary. It does not grant the agent access to wallets, transactions, or AI conversations.

## 2. Goals and non-goals

### Goals

- Persist feedback with a validated type and lifecycle status.
- Keep raw feedback private to its owner and the authorized internal agent.
- Give the local/dev agent a pollable API protected by a dedicated service credential.
- Atomically publish a changelog entry and mark one or more feedback records fixed.
- Expose sanitized changelog entries publicly and show fixed version/title in the user feedback UI.
- Emit Redis audit events for feedback, status, authorization, and changelog actions without logging descriptions or credentials.
- Document the contract in Markdown and generated Swagger so another AI agent can use it without reading implementation details.

### Non-goals

- No public raw-feedback feed.
- No agent access to financial data APIs or user API-key management in this slice.
- No AI chat history, tool-calling protocol, vector search, or autonomous code execution.
- No automatic deployment or automatic merge after a fix.

## 3. Data model

### `feedback`

| Column | Type | Rules |
| --- | --- | --- |
| `id` | UUID | primary key |
| `user_id` | UUID | required, foreign key to the user, indexed |
| `type` | text | `bug`, `feature`, or `improvement` |
| `title` | text | trimmed, required, max 200 characters |
| `description` | text | trimmed, required, max 10,000 characters |
| `status` | text | `open`, `triaged`, `in_progress`, `fixed`, `rejected`; default `open` |
| `fixed_at` | timestamptz | nullable; set only when status becomes `fixed` |
| `changelog_id` | UUID | nullable foreign key to `changelog` |
| `created_at` | timestamptz | required |
| `updated_at` | timestamptz | required |

Indexes cover `(user_id, created_at desc)` and `(status, created_at asc)` for owner history and agent polling.

### `changelog`

| Column | Type | Rules |
| --- | --- | --- |
| `id` | UUID | primary key |
| `version` | text | required, unique, max 50 |
| `title` | text | trimmed, required, max 200 |
| `description` | text | trimmed, required, max 10,000 |
| `published_at` | timestamptz | required |
| `created_at` | timestamptz | required |

`feedback.changelog_id` is nullable until fixed. A changelog can reference many feedback rows through that foreign key; no join table is needed for the first slice.

## 4. Authorization model

### User routes

Existing JWT bearer authentication is required. All reads are owner-scoped by the authenticated user ID. A user cannot read, update, or delete another user's feedback.

### Agent/internal routes

The first agent credential is a deployment-configured `FEEDBACK_AGENT_TOKEN`. It is sent as `Authorization: Bearer $FEEDBACK_AGENT_TOKEN` and validated with constant-time comparison. The token has only feedback/changelog service permissions; it cannot call wallet, transaction, budget, or AI entry routes. The middleware boundary is named and scoped so the future user API-key implementation from `TICKET-10-01` can plug in without changing feedback handlers.

Agent responses omit `user_id`, email, and other account identifiers. The description is available to the authorized local/dev agent because it is the issue being fixed; no description is written to logs or Redis audit payloads.

### Public routes

Only published changelog records are public. Raw feedback is never public.

## 5. API contract

All routes use the existing `/api/v1` prefix and response/problem envelopes.

### User-facing

- `POST /api/v1/feedback` — create feedback; returns `201` and the owner-scoped record.
- `GET /api/v1/feedback` — list current user's records, newest first.
- `GET /api/v1/feedback/:id` — fetch one owner-scoped record; returns `404` for another owner to avoid disclosure.

Create input:

```json
{
  "type": "bug",
  "title": "Grab bị phân loại sai",
  "description": "Các giao dịch Grab tháng này đang vào nhóm khác."
}
```

### Agent/internal

- `GET /api/v1/agent/feedback?status=open&limit=50` — poll actionable feedback. `status` is restricted to lifecycle values and `limit` is bounded.
- `PATCH /api/v1/internal/feedback/:id/status` — move one item to `triaged`, `in_progress`, or `rejected`. The handler rejects backward transitions, terminal transitions, and direct `fixed`; only changelog publication can finalize `fixed`.
- `POST /api/v1/internal/changelog` — atomically create a changelog and finalize all listed feedback IDs.

Status input:

```json
{ "status": "in_progress" }
```

Changelog input:

```json
{
  "feedback_ids": ["fb-uuid"],
  "version": "1.4.2",
  "title": "Improved transaction categorization",
  "description": "Fixed incorrect categorization for recurring merchants."
}
```

`POST /internal/changelog` returns `409` for a duplicate version, an unknown feedback ID, a feedback owned by a terminal/rejected state, or a feedback already linked to another changelog.

### Public changelog

- `GET /api/v1/changelog` — newest published entries, bounded page size.
- `GET /api/v1/changelog/:id` — one sanitized entry with fixed feedback count/IDs omitted from the public payload.

## 6. Lifecycle and atomicity

Allowed transitions are:

```text
open → triaged → in_progress → fixed
open → triaged → in_progress → rejected
```

An internal status update is a single-row transaction and sets `fixed_at` only for `fixed`; it clears `fixed_at` for non-terminal statuses only when no changelog is linked. A linked/fixed/rejected record is terminal and cannot be moved backward.

Changelog creation runs in one database transaction:

1. Validate version/title/description and lock all requested feedback rows in deterministic ID order.
2. Confirm every row is visible to the agent, not rejected/fixed, and belongs to the requested transition.
3. Insert the unique changelog version.
4. Update every row to `fixed`, set `fixed_at`, and set `changelog_id`.
5. Commit, then emit a single Redis audit event containing actor kind, changelog ID, version, and feedback IDs.

Any validation or insert failure rolls back both the changelog and all feedback updates. Redis audit failure is reported to the operational log after commit and never exposes the description or token; it must not make the user-visible state half-complete.

## 7. UI contract

Add an authenticated Account → Feedback route using existing base components only:

- `BaseSelect` for type and status display;
- `BaseTextInput`/text area atom for title and description;
- `BaseButton`, `SurfaceCard`, `StatusMessage`, and shared loading/error states;
- no screen-local native controls or literal color tokens.

The page supports submit, reload, empty/loading/error states, and owner history. A fixed record renders:

```text
✅ Đã xử lý
Fixed in v1.4.2
Improved transaction categorization
```

The first UI slice does not expose internal agent controls. Changelog browsing can be a read-only section on the same page; it is not coupled to AI chat.

## 8. Audit and privacy

Audit event names are `feedback.created`, `feedback.status_changed`, `feedback.agent_denied`, `feedback.agent_read`, and `changelog.published`. Payloads contain request ID, actor kind, feedback/changelog UUIDs, old/new status, and version only. They never contain title, description, email, JWT, service token, or financial records.

Authorization failures are audited with the reason category, not the credential value. Redis is a transport/cache audit sink, not the source of truth; database state and the API response remain authoritative.

## 9. Verification plan

- Migration/schema tests: apply and rollback, FK/index/unique-version constraints, migration CLI dirty-state check.
- Backend unit tests: input bounds, transition matrix, owner scoping, agent credential scope, public redaction, duplicate version, terminal conflict, audit payload redaction.
- PostgreSQL integration tests: concurrent changelog finalization, deterministic row locks, rollback on one invalid feedback, owner isolation, fixed timestamp/link consistency.
- API integration test through the Vite proxy: user submit/list/detail, agent poll/status, atomic changelog, public changelog, unauthorized and cross-owner cases.
- Chromium E2E: submit feedback from Account, reload, render fixed status/version, ensure no raw other-user feedback is visible.
- Docs checks: generated Swagger, API/ERD/architecture, context/backlog, validation matrix, release changelog, and this AI-agent-readable spec.

## 10. Rollout and compatibility

The migration is additive. Existing JWT routes, AI chat, ledger, and current API-key backlog behavior are unchanged. The service token is optional in development only when agent routes are disabled; non-development startup must reject an enabled agent route without a configured token. No existing data is backfilled.

Implementation follows this spec with a separate plan and TDD red→green cycles. The implementation must not claim that the local agent actually fixed a user issue until a code change, verification evidence, status transition, and changelog transaction all exist.
