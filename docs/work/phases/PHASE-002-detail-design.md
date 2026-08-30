---
artifact_type: detail_design
id: PHASE-002-DETAIL-DESIGN
status: approved
owner: shared
approval: approved
human_fields:
  - approval
  - constraints
  - scope_decisions
ai_fields:
  - problem
  - context_loaded
  - brownfield_scope
  - proposed_approach
  - design_tradeoffs
  - architecture_overview
  - execution_flow
  - api_data_model
  - security
  - test_plan
  - reconciliation_plan
shared_fields:
  - status
  - trace
  - small_task_exemption
trace:
  backlog_item: BL-002
  requirement:
    - REQ-F-002
    - REQ-F-003
    - REQ-F-011
    - REQ-NF-002
    - REQ-NF-005
  phase: PHASE-002
  ticket_or_bug:
    - TICKET-005
    - TICKET-006
    - TICKET-007
  test_verification: ../test-verification/PHASE-002-finance-core.md
  validation_matrix: ../VALIDATION_MATRIX.md
  docs_review: per-ticket_completion_checklist
  adrs:
    - ../../decisions/ADR-001-react-go-modular-monolith.md
    - ../../decisions/ADR-003-offline-sync-conflict-review.md
    - ../../decisions/ADR-004-review-first-ingestion.md
  master_docs_touched:
    - ../../architecture/API.md
    - ../../architecture/ERD.md
    - ../../architecture/ARCHITECTURE.md
---

# DETAIL DESIGN: PHASE-002 Finance Core

## 1. Context & Scope

### Problem Statement

- Problem statement: MyPocket cannot become a useful finance app until wallets, categories, and confirmed transactions update balances correctly.
- Why now: PHASE-001 identity, platform, PWA shell, local Compose, and browser auth proof are in review, so finance core is the next dependency for offline sync and analytics.
- Success outcome: authenticated users can manage their own wallets/categories and create/search/edit/archive income, expense, transfer, and adjustment transactions with exact VND integer accounting.

### Context Loaded

- `docs/CONTEXT.md`
- `docs/work/BACKLOG.md`
- `docs/standards/README.md`
- `docs/standards/QUALITY_BAR.md`
- `docs/standards/VALIDATION.md`
- `docs/work/phases/PHASE-002-finance-core.md`
- `docs/requirements/REQUIREMENTS.md`
- `docs/architecture/API.md`
- `docs/architecture/ERD.md`
- `docs/architecture/ARCHITECTURE.md`
- `docs/decisions/ADR-001-react-go-modular-monolith.md`
- `docs/decisions/ADR-003-offline-sync-conflict-review.md`
- `docs/decisions/ADR-004-review-first-ingestion.md`

### Brownfield Scope

- Touched modules/files: `backend/internal/finance`, `backend/migrations`, `backend/internal/platform/httpapi`, `frontend/src/app`, `frontend/e2e`, docs under `docs/work`, `docs/architecture`, and `docs/releases`.
- Direct dependencies inspected: identity user context, auth route behavior, DB migration runner, API error envelope, frontend API client, design contract.
- Contracts affected: finance REST API, ERD, validation matrix, mobile PWA workflows.
- Known unknowns: exact production category seed wording may need human review after first UI pass.
- Scope expansion reason: no expansion beyond PHASE-002.

### Small Task Exemption

- Small task exemption: no
- Reason: PHASE-002 introduces finance schema, public API, authorization, and user-facing accounting behavior.
- Impact checked: API=yes, DB=yes, Security=yes, Runtime=no, Standards=no

## 2. Design Considerations & Trade-offs

| Consideration / Alternative | Pros | Cons | Decision |
| --- | --- | --- | --- |
| Store wallet balances as materialized integers updated transactionally | Fast UI/API reads; easier balance proof | Requires exact reversal/reapply on edit/archive | chosen |
| Derive balances from all transactions on every read | Simple write path | Expensive, harder offline versioning, slow for mobile | rejected |
| Use one `transactions` row for transfer | Simple list model | Needs clear source/destination effects | chosen |
| Use separate double-entry ledger rows in PHASE-002 | Strong accounting model | More schema/UI complexity than current phase needs | rejected |
| Implement receipt OCR now | More complete add flow | Violates PHASE-002 out-of-scope provider boundary | rejected |

## 3. Architecture Overview

### Component Responsibilities

| Component | Role |
| --- | --- |
| `backend/internal/finance` | Domain validation, accounting effects, repository transactions, idempotency helpers, DTOs |
| `backend/internal/platform/httpapi` | Authenticated finance route registration, request parsing, stable JSON errors |
| `backend/migrations` | Wallet/category/transaction/receipt/idempotency schema and Vietnamese seed data |
| `frontend/src/app` | Mobile screens for wallets, category selection, add/edit/search transaction workflows |
| `frontend/e2e` | Browser proof for wallet/category and transaction flows |

```text
-------------+       /api/v1 finance JSON        +------------------+
| Frontend    | --------------------------------> | httpapi handlers |
| mobile PWA  | <-------------------------------- | auth + errors    |
+------+------+                                    +---------+--------+
       |                                                     |
       | renders VND state                                  v
       |                                           +------------------+
       |                                           | finance service  |
       |                                           | validation/effects|
       |                                           +---------+--------+
       |                                                     |
       v                                                     v
  IndexedDB later                                   PostgreSQL tables
  PHASE-003                                        PHASE-002 source
```

## 4. Execution Flow

1. Migrations create finance tables, constraints, indexes, and idempotency ledger columns needed by PHASE-003.
2. Finance services validate ownership, active/archive state, positive integer amounts, and category activation.
3. Create/edit/archive commands run in one SQL transaction, lock affected wallets, write transaction state, update balances, and record idempotency result.
4. HTTP handlers derive user ID only from auth context, never from payload ownership fields.
5. Frontend loads wallets/categories/transactions, renders VND integer formatting, and submits commands through the API client with CSRF and idempotency keys.
6. E2E validates mobile flows against the fixture-auth API.

## 5. API & Data Model Design

### API Changes

- `GET /api/v1/wallets`: list user wallets with balance, type, archive state, and default AI flag.
- `POST /api/v1/wallets`: create wallet.
- `PATCH /api/v1/wallets/{id}`: edit wallet metadata, archive flag, include-in-total, and default AI selection.
- `GET /api/v1/categories`: list system/user categories and hierarchy.
- `POST /api/v1/categories`: create user category.
- `PATCH /api/v1/categories/{id}`: edit/archive user category; reject system category mutation.
- `PUT /api/v1/wallets/{wallet_id}/categories/{category_id}`: toggle activation.
- `GET /api/v1/transactions`: list/search user transactions.
- `POST /api/v1/transactions`: create income, expense, transfer, or adjustment.
- `PATCH /api/v1/transactions/{id}`: edit or archive a transaction with balance reversal/reapply.

All state-changing endpoints require `X-CSRF-Token`; retryable creates and edits require `Idempotency-Key`.

### Data Model Changes

- `wallets`: `id`, `user_id`, `name`, `type`, `balance_vnd`, `include_in_total`, `is_default_ai`, `archived_at`, `version`, credit metadata, timestamps.
- `categories`: `id`, nullable `user_id`, nullable `parent_id`, `kind`, `name`, `system_key`, `is_system`, `archived_at`, timestamps.
- `wallet_category_settings`: `wallet_id`, `category_id`, `is_active`, timestamps.
- `transactions`: `id`, `user_id`, `type`, `source_wallet_id`, optional `destination_wallet_id`, optional `category_id`, `amount_vnd`, `occurred_at`, `note`, `with_person`, `event_id`, `receipt_object_id`, `excluded_from_reports`, `archived_at`, `version`, timestamps.
- `finance_idempotency_keys`: `user_id`, `key`, `request_hash`, `response_json`, `created_at`.
- `receipt_objects`: `id`, `user_id`, optional `transaction_id`, `object_key`, `content_type`, `size_bytes`, `checksum_sha256`, timestamps.

Schema migration needed: yes.

## 6. Security & Authorization

- Authentication changes: none; reuse PHASE-001 cookie auth.
- Authorization / Permissions: every finance query filters by authenticated `user_id`; mutations verify all referenced wallets/categories/receipt metadata belong to the same user or are allowed system rows.
- Data Privacy / PII impact: notes, with-person values, and receipt metadata are user-private.
- Input Validation: VND amounts are positive integers; transaction type controls required fields; transfer wallets must differ; archived objects reject new writes except allowed archive/update commands.

## 7. Implementation & Verification Plan

### Impacted Areas

- Code/modules: `backend/internal/finance`, `backend/internal/platform/httpapi`, `frontend/src/app`, `frontend/e2e`.
- Product behavior: wallet/category management and transaction accounting.
- API/contracts: new finance endpoints and idempotency header.
- Data/schema: finance tables, indexes, and seed data.
- Security/auth: ownership scoping for every finance operation.
- Deployment/runtime: migrations only; no new service.
- Docs: API, ERD, validation matrix, context, backlog, changelog.

### Test Plan

- Unit: finance validation and accounting effects for income, expense, transfer, adjustment, edit, archive.
- Integration: migration from empty DB, seed idempotency, constraints, ownership isolation, transaction atomicity, idempotency replay.
- E2E: mobile wallet list/create/edit/archive, category activation, add/search/edit/archive transactions.
- UAT: Vietnamese copy, VND formatting, credit wallet behavior, balance privacy.
- Manual/platform: existing Compose smoke plus finance API readiness through web proxy.
- Docs review: per-ticket checklist and validation matrix rows for REQ-F-002, REQ-F-003, REQ-F-011, REQ-NF-002, REQ-NF-005.

### Validation Matrix Impact

- Update required: yes
- Rows: REQ-F-002, REQ-F-003, REQ-F-011, REQ-NF-002, REQ-NF-005, REQ-F-016

### Verification Results

- Command: not_run
- Result: pending
- Notes: Design artifact only; implementation has not started.

## 8. Reconciliation Plan

- Requirements docs: no change expected unless scope changes.
- Architecture docs: update implemented finance module details.
- API docs: update concrete finance request/response contracts.
- ERD docs: update implemented table definitions and constraints.
- SDD docs: no change expected.
- ADR: no change expected unless accounting/idempotency model changes.
- Context: update after each ticket execution.

### Docs Review Checklist

- Code changed but docs unchanged reason: not applicable; no code in this design step.
- Requirements updated or not needed reason: no change expected.
- Architecture updated or not needed reason: pending execution.
- API updated or not needed reason: pending execution.
- ERD/data updated or not needed reason: pending execution.
- SDD updated or not needed reason: no change expected.
- ADR created or not needed reason: no new durable decision beyond accepted ADRs.
