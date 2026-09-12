---
artifact_type: detail_design
id: PHASE-004-DETAIL-DESIGN
status: approved
owner: shared
approval: approved
approved_on: 2026-08-30
trace:
  backlog_item: BL-004
  phase: PHASE-004
  requirements: [REQ-F-005, REQ-F-012, REQ-NF-002, REQ-NF-005]
  tickets: [TICKET-011, TICKET-012, TICKET-013, TICKET-014]
  validation_matrix: ../VALIDATION_MATRIX.md
  master_docs_touched: [../../architecture/API.md, ../../architecture/ERD.md, ../../architecture/ARCHITECTURE.md]
---

# DETAIL DESIGN: PHASE-004 Planning and Automation

## 1. Context and Scope

This phase adds budgets, events, debts/loans, recurring drafts, in-app notices, and best-effort Web Push without bypassing finance accounting. It consumes confirmed PHASE-002 transactions and PHASE-003 versioned commands.

- In scope: planning CRUD, deterministic period calculations in `Asia/Ho_Chi_Minh`, worker leases, threshold notices, reviewable recurring drafts, repayments, inbox, and push subscriptions.
- Out of scope: automatic transaction confirmation, email, AI/OCR, and dashboard-wide analytics.
- Approval: user-delegated decisions on 2026-08-30. Small task exemption: no.

## 2. Locked Decisions

| Area | Decision |
| --- | --- |
| Budget periods | Weekly, monthly, quarterly, yearly, and custom inclusive date ranges; server computes boundaries in Ho Chi Minh time |
| Progress | Sum confirmed non-archived expenses not excluded from reports; transfers and adjustments excluded |
| Alerts | Durable dedupe key `(budget_id, threshold, period_start)` for 80% and 100% |
| Events | Optional transaction association; grouping never changes wallet effects |
| Recurrence | RRULE-like normalized schedule with explicit timezone; occurrences create drafts only |
| Worker | PostgreSQL lease with `locked_until`, deterministic occurrence key, capped retry, and restart safety |
| Debts | Direction `borrowed` or `lent`; repayments link confirmed transactions and cannot exceed remaining principal without explicit adjustment |
| Notifications | In-app inbox is authoritative; Web Push is optional delivery with endpoint expiry cleanup |

## 3. Components and Data

- `backend/internal/planning`: periods, budget evaluation, events, obligations, schedules, occurrence generator, notice service.
- `backend/internal/worker`: lease acquisition, due occurrence processing, notice delivery, retry classification.
- `frontend/src/planning`: budget/event/debt/schedule forms, progress, draft review, inbox, and push permission states.

PostgreSQL tables: `budgets`, `budget_categories`, `budget_alerts`, `events`, `event_transactions`, `obligations`, `obligation_repayments`, `recurring_schedules`, `recurring_occurrences`, `transaction_drafts`, `notifications`, `push_subscriptions`, and `worker_leases`. Every user record has ownership, timestamps, archive state where applicable, and optimistic `version`.

## 4. API Contract

- CRUD under `/api/v1/budgets`, `/events`, `/obligations`, `/recurring-schedules` with CSRF and version checks.
- `GET /api/v1/planning/summary` returns current budget progress, due obligations, and pending drafts.
- `GET/PATCH /api/v1/notifications` lists and marks read; `POST/DELETE /api/v1/push-subscriptions` manages endpoints.
- `POST /api/v1/drafts/{id}/confirm` delegates exactly once to the PHASE-002 accounting command; recurring worker never calls accounting directly.
- Stable errors: validation, forbidden, conflict, not-found, retryable provider/delivery failure.

## 5. User Experience

- Budget tab becomes a working planning surface with progress bars, period selector, threshold state, create/edit/archive controls, and empty/error/offline states.
- Events and debts use compact mobile lists and bottom sheets consistent with `design/DESIGN.md`.
- Drafts show source, scheduled date, wallet/category resolution, editable amount/note, and explicit Confirm/Reject commands.
- Notification inbox remains usable when push permission is denied, unsupported, or delivery fails.

## 6. Security and Reliability

- All linked finance objects must belong to the authenticated user.
- Push endpoints and encryption keys are private; logs redact them.
- Occurrence and alert dedupe constraints are authoritative under concurrent workers.
- No schedule, webhook, or notice can modify balances without explicit draft confirmation.

## 7. Verification and Reconciliation

- Unit: period boundaries, threshold crossings, recurrence generation, debt remaining amount, and retry policy.
- Integration: ownership, alert/occurrence dedupe, leases, restart recovery, repayment linkage, and confirmation idempotency.
- E2E/UAT: create budgets/events/debts/schedules, cross 80/100 thresholds, review recurring draft, inbox/push states, and timezone boundaries.
- Reconcile API, ERD, architecture, validation matrix, context, backlog, changelog, and provider/runtime docs.

