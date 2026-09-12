---
artifact_type: requirement_spec
id: REQ-MYPOCKET-V2-DRAFT
status: draft
owner: human
approval: pending_human_review
updated: 2026-09-12
human_fields:
  - product_summary
  - goals
  - non_goals
  - users_and_stakeholders
  - acceptance_criteria
ai_fields:
  - functional_requirements
  - non_functional_requirements
  - constraints
  - ambiguity_report
shared_fields:
  - status
  - trace
trace:
  context: ../CONTEXT.md
  backlog: ../work/BACKLOG.md
  roadmap: ../work/ROADMAP.md
  master_spec: SPEC.md
  requirements: REQUIREMENTS.md
  user_stories: USER_STORIES.md
  business_rules: MYPOCKET-V2-BUSINESS-RULES-DRAFT.md
  money_lover_parity: ../research/moneylover/FEATURE-PARITY-2026-09-08.md
  validation_matrix: ../work/VALIDATION_MATRIX.md
  release_notes: ../releases/CHANGELOG.md
  detail_designs:
    - ../superpowers/specs/2026-09-11-mypocket-complete-personal-finance-design.md
    - ../superpowers/specs/2026-09-12-budget-recurring-agent-completion-design.md
    - ../superpowers/specs/2026-09-12-two-chat-design.md
    - ../work/phases/PHASE-005-asset-portfolio-detail-design.md
  adrs:
    - ../decisions/ADR-003-offline-sync-conflict-review.md
    - ../decisions/ADR-004-review-first-ingestion.md
    - ../decisions/ADR-006-separate-asset-portfolio-valuation.md
    - ../decisions/ADR-008-atomic-offline-sync-commit.md
---

# MyPocket V2 Product Specification Draft

## Review Status

- Status: `draft`; this document does not authorize implementation.
- Purpose: consolidate the existing product documents into one reviewable product contract.
- Human direction captured on 2026-09-12: use Money Lover's public personal-finance capabilities as the behavioral baseline, exclude online banking and voice, and retain MyPocket-specific jars, asset tracking, API keys, and AI agents.
- Existing approved behavior remains authoritative until this draft is approved and reconciled into the master requirements.

## Product Summary

MyPocket is a Vietnamese, VND-first, offline-capable personal-finance PWA. It covers the core personal-finance workflows publicly documented by Money Lover while keeping an independent brand, implementation, security model, and user experience. MyPocket differs by assigning spending jars directly to transactions, tracking market-valued assets separately from wallet accounting, exposing scoped API keys for bots and external applications, and providing review-first AI intake plus read-only financial advice through controlled tools.

## Product Goal

A user can record, plan, inspect, automate, and safely integrate all of their personal-finance activity from web or installed PWA, with mathematically correct balances and reports across online, offline, retry, concurrent, API-key, and AI-assisted workflows.

## Source And Precedence

1. This draft records the latest human direction for product scope.
2. MyPocket business invariants and approved ADRs take precedence over Money Lover behavior where the products differ.
3. The Money Lover parity matrix is a capability baseline, not permission to copy branding, assets, text, paywalls, or proprietary implementation.
4. A baseline capability is complete only when its selected parity row is either `verified` with evidence or `explicitly_excluded` with owner approval.

## Users And Stakeholders

- **Finance owner:** owns wallets, categories, transactions, jars, plans, assets, reports, files, agent history, API keys, and offline data.
- **Shared-wallet member:** accesses only wallets explicitly shared with that member and only with the granted role.
- **Bot or external application:** uses a revocable, scoped API key to act within the owner's selected capabilities.
- **Homelab operator:** configures providers and runtime services without receiving access to plaintext user API keys or finance content through logs.
- **Audit viewer:** reads restricted operational audit events without becoming owner of finance data.

## Goals

- Deliver the selected Money Lover core capability baseline with MyPocket-specific business rules and verifiable workflows.
- Keep wallet balances, transfers, debts, jars, assets, and reports correct after edits, archives, retries, conflicts, and offline replay.
- Let users manage wallets, transactions, categories, credit, savings goals, recurring items, events, debts, shared wallets, reports, and exports.
- Make every supported user-owned capability available through a documented API where its security model permits external access.
- Let transaction-intake AI produce reviewable drafts and let advisor AI answer from fresh, tool-sourced financial data.
- Preserve data ownership, least privilege, auditability, and user confirmation for consequential actions.

## Non-Goals

- Online bank connection, bank OAuth, bank synchronization, bank webhook ingestion, or automatic bank transaction import.
- Voice recording, speech-to-text, or voice-triggered transaction entry.
- Money Lover branding, proprietary assets, copied UI, subscription tiers, advertisements, trials, or paywalls.
- Native-only widgets, notification interception, Google Pay capture, or Apple Shortcuts in this product scope.
- Tax filing, brokerage execution, crypto exchange execution, investment recommendations, or automatic trading.
- Full accounting-suite, business bookkeeping, invoicing, or multi-organization administration.

## Product Architecture Boundary

```text
+------------------------+      cookie / API key      +------------------------+
| React PWA              | -------------------------> | Go application API     |
| IndexedDB + outbox     | <------------------------- | domain services        |
+------------------------+                            +-----------+------------+
                                                                |
                           +------------------------------------+-----------------------------------+
                           |                                    |                                   |
                           v                                    v                                   v
                +--------------------+              +--------------------+             +--------------------+
                | PostgreSQL         |              | Private S3 objects |             | Worker processes   |
                | authoritative data |              | receipt media      |             | recurring/OCR/AI   |
                +--------------------+              +--------------------+             +--------------------+
                                                                                                  |
                                                                                     approved providers only
```

## Functional Requirements

### FR-01 Money Lover Capability Baseline

- **Current:** The parity audit maps 96 public Help Center articles into 50 capability/evidence rows; several rows remain partial, missing, deferred, or excluded by the older scope.
- **Target:** MyPocket supports the selected core capabilities for wallets, transactions, categories, credit, savings goals, shared wallets, planning, reports, search, notifications, receipts, export, and account lifecycle under MyPocket rules.
- **Acceptance:** Every row selected for the release in `FEATURE-PARITY-2026-09-08.md` is marked `verified` with linked proof or `explicitly_excluded` with owner approval; no selected row remains `partial`, `missing`, or undocumented.

### FR-02 Wallets And Categories

- **Current:** Core wallet and two-level category behavior exists, but the parity matrix records incomplete history, restore, credit lifecycle, smart suggestions, metadata, and some mounted management journeys.
- **Target:** Users can create, edit, archive, restore, filter, and inspect supported wallet types and two-level categories; category availability can vary by wallet without deleting historical references.
- **Acceptance:** Automated and UAT fixtures prove wallet/category create, update, archive, restore, hierarchy, activation, filtering, credit metadata, and historical readability after reload and offline reconnect.

### FR-03 Transactions And Accounting

- **Current:** Income, expense, transfer, adjustment, edit, archive, and atomic accounting exist in varying degrees of UI/API completeness.
- **Target:** Users can create, edit, duplicate, archive, bulk-select, search, and filter income, expenses, transfers with fees, and balance adjustments while preserving exact accounting effects.
- **Acceptance:** Golden ledger tests prove balance deltas and exact reversal/reapplication for every transaction type, including retries, stale versions, concurrent mutation, transfer neutrality, archive, and restore where supported.

### FR-04 Transaction-Applied Spending Jars

- **Current:** The local model supports optional `budget_id` on expenses and draft/schedule carry-through; older category-derived budget behavior remains in historical documents.
- **Target:** A spending jar is a limit and period whose progress is calculated only from confirmed, report-included expense transactions explicitly assigned to that jar. Category selection may provide suggestions but never determines progress.
- **Acceptance:** Tests prove an expense changes only its assigned jar, unassigned expenses change no jar, transfers/income/adjustments cannot be assigned, and edit/archive/reassignment updates progress exactly once.

### FR-05 Planning And Automation

- **Current:** Budgets, events, debts, recurring schedules, drafts, and notifications exist, but selected Money Lover planning journeys remain incomplete or require reconciliation.
- **Target:** Users can manage jars, events/trips, bills, recurring schedules, debts/loans, repayments, credit statements, and savings goals with explicit lifecycle states and deterministic accounting links.
- **Acceptance:** Each planning object passes create, edit, pause/complete/archive, reload, offline/conflict, notification, and linked-transaction tests appropriate to that object; draft-producing automation never duplicates accounting.

### FR-06 Shared Wallets

- **Current:** User isolation exists; no approved membership, invitation, or shared-wallet authorization model is implemented.
- **Target:** An owner can invite a user to a wallet, grant a defined role, revoke access, and inspect member-attributed activity without exposing any unshared finance data.
- **Acceptance:** Two-user integration and E2E tests prove invitation, acceptance, role enforcement, member reporting, revocation, leaving, and denial of cross-user access outside the shared wallet.

### FR-07 Reports And Financial Insight

- **Current:** Dashboard and several reports exist; the parity audit records gaps in scope consistency, period history, drilldown, forecasting, and some current UI controls.
- **Target:** Dashboard, cash flow, category, trend, comparison, cumulative, jar, debt, credit, asset, and net-worth reports derive from authoritative confirmed data using consistent wallet/date/report-inclusion rules.
- **Acceptance:** A versioned golden dataset produces independently calculated expected totals for every report, with drilldown rows reconciling to totals and `Asia/Ho_Chi_Minh` boundaries verified.

### FR-08 Asset Portfolio And Value Movement

- **Current:** Gold, stock, crypto, foreign-currency, and other positions with trades and price history have local implementation evidence and remain separate from wallet accounting.
- **Target:** Users track quantity, buy/sell history, moving-average cost, manual or provider price snapshots, market value, realized P&L, unrealized P&L, and value changes without turning market movement into income/expense transactions.
- **Acceptance:** Deterministic replay tests prove quantity, cost basis, realized/unrealized P&L, missing/stale price states, historical corrections, offline mutations, and separate wallet/investment/combined net-worth totals.

### FR-09 Offline And Multi-Device Operation

- **Current:** IndexedDB mirror/outbox, incremental sync, conflict handling, and atomic replay exist for core entities; parity coverage varies by newer domains.
- **Target:** Supported finance, planning, jar, shared-wallet, and manual portfolio mutations can be captured offline where safe, then replay idempotently with explicit conflict resolution.
- **Acceptance:** Browser tests prove reload while offline, queued mutation persistence, reconnect, lost-response replay, two-device conflict, logout cleanup, and no duplicate financial effect for each offline-enabled entity family.

### FR-10 API Keys For Bots And Applications

- **Current:** Users can create/list/revoke one-time plaintext `mpk_` keys; keys authenticate as the owner but do not yet have granular scopes or expiration.
- **Target:** Users create named, revocable API keys with explicit least-privilege scopes, optional expiration, safe prefix display, last-used metadata, and access to the same domain services as the PWA.
- **Acceptance:** Contract tests prove one-time secret display, hash-only storage, scope allow/deny, expiration, immediate revocation, per-user ownership, rate limiting, audit correlation, and inability for an API key to create or manage other keys.

### FR-11 Public API Contract

- **Current:** A versioned REST API and OpenAPI document exist, with historical parity findings showing some documentation/runtime drift.
- **Target:** Every externally supported capability has a versioned, machine-readable contract with bearer authentication, stable errors, idempotency, optimistic concurrency, pagination, and examples.
- **Acceptance:** Router-to-OpenAPI coverage is complete; cookie and bearer contract suites pass; public docs contain no route, field, or behavior absent from the running API.

### FR-12 AI Transaction Intake

- **Current:** Intake persists sessions/messages and creates validated income/expense drafts from structured model output; context history, clarification state, and true model-selected tool dispatch are incomplete.
- **Target:** Text or owned-receipt input can produce zero or more editable drafts or a `needs_input` response. Intake may use only allowlisted read tools and a draft-proposal tool; it can never confirm a transaction.
- **Acceptance:** Tests prove multi-draft extraction, missing-field clarification, receipt/OCR failure handling, prompt injection resistance, ownership validation, idempotent retry, human edit, confirm/reject, and zero accounting effect before confirmation.

### FR-13 Read-Only AI Advisor

- **Current:** Advisor returns answer-only JSON from a fixed 30-day aggregate and does not execute model-selected finance tools or consume bounded session history.
- **Target:** Advisor uses allowlisted read-only tools for reports, transactions, jars, obligations, assets, and relevant chat history; every account-specific factual claim is grounded in fresh tool results.
- **Acceptance:** Tests prove correct tool selection for representative questions, source IDs on quantitative claims, bounded history/summary behavior, denial of every write tool, prompt injection resistance, and truthful insufficient-data responses.

### FR-14 Receipt And OCR Assistance

- **Current:** Owned images are stored privately, verified, sent to an external OCR service by a worker, and included only in the matching agent run context.
- **Target:** Receipt OCR remains a deterministic server-orchestrated preprocessing step, not a general public OCR proxy; extracted fields remain untrusted and reviewable.
- **Acceptance:** Provider tests prove ownership, checksum, size/type limit, timeout/retry/expiry, private storage, redacted errors, and no draft/transaction creation from failed or foreign OCR output.

### FR-15 Notifications, Export, And Data Lifecycle

- **Current:** Durable notices, Web Push, export jobs, reset/delete, audit, and retention have implementation or design artifacts with differing release status.
- **Target:** Users receive durable review/threshold/due notices, export an owned snapshot, and perform confirmed reset/delete operations with observable job progress and complete owned-object handling.
- **Acceptance:** Integration and UAT evidence proves durable fallback when push fails, user-only export contents, idempotent lifecycle jobs, private-object cleanup, audit correlation, and recovery behavior for interrupted jobs.

## API-Key Scope Draft

| Scope | Allows | Explicitly denies |
| --- | --- | --- |
| `finance:read` | Wallet, category, transaction and report reads | Any mutation |
| `finance:write` | Finance mutations and explicit draft confirmation permitted by domain rules | API-key management and agent-controlled/implicit confirmation |
| `planning:read` / `planning:write` | Jars, schedules, events, goals, credit and debts | Finance writes without `finance:write` |
| `portfolio:read` / `portfolio:write` | Positions, trades and manual prices | Brokerage/exchange execution |
| `agent:intake` | Submit/read intake sessions and drafts | Advisor and confirmation |
| `agent:advisor` | Submit/read advisor sessions and read tools | Every write tool |
| `sync` | Approved offline synchronization contract | Scope escalation |
| `export` | Request/read owned exports | Import, reset or delete |

## Non-Functional Requirements

- PostgreSQL is authoritative; IndexedDB is a user-scoped offline mirror/outbox.
- VND amounts are signed or unsigned integer units according to domain rules; asset quantities use bounded decimal strings and PostgreSQL numeric storage.
- All retryable mutations are idempotent; mutable entities use optimistic versions and never silently use last-write-wins.
- Every query, mutation, file, session, tool execution, and sync record is scoped to its effective owner or explicit shared-wallet membership.
- Browser-cookie mutations require CSRF; bearer API-key calls use scopes and never fall back to cookie identity.
- Provider/model/OCR output is untrusted, schema-validated, bounded, ownership-checked, and prevented from directly changing confirmed accounting.
- Secrets, raw authorization headers, cookies, prompts, receipt text, and provider payloads are excluded from normal logs and audit metadata.
- Reports and periods use `Asia/Ho_Chi_Minh`; timestamps remain explicit and round-trip safely.
- User-visible capabilities require responsive desktop/mobile PWA workflows, accessible controls, offline/error/loading states, public docs, and automated plus UAT proof.

## Constraints

- Preserve the React/TypeScript PWA, Go modular monolith, PostgreSQL, Redis, private S3-compatible storage, API/worker split, and current deployment boundary unless a later approved ADR changes them.
- Domain services own business invariants; REST handlers, sync adapters, workers, and AI tools reuse them instead of duplicating finance logic.
- External AI/OCR/asset-price providers are replaceable adapters with bounded timeouts, retries, capability checks, and server-only credentials.
- Online banking and voice must not appear as active UI controls, runtime routes, tool definitions, or advertised supported capabilities.

## Product Acceptance Criteria

- [ ] All selected Money Lover parity rows are `verified` or `explicitly_excluded` with traceable owner approval.
- [ ] Golden accounting fixtures pass for transactions, transfers, adjustments, jars, recurring items, debts, credit, goals, and shared-wallet activity.
- [ ] A transaction affects at most one explicitly assigned jar and never affects a jar through category membership alone.
- [ ] Portfolio value movement never changes wallet balance, cash flow, income, or expense totals.
- [ ] API keys enforce scopes, expiration, revocation, ownership, rate limits, and audit correlation through negative as well as positive tests.
- [ ] Intake creates only reviewable drafts; advisor remains read-only; neither can bypass domain services or ownership checks.
- [ ] Offline/retry/concurrency tests show no duplicate or silently overwritten financial effects.
- [ ] Public API and user documentation match the running contracts and selected product scope.
- [ ] Desktop, mobile Chromium, mobile WebKit, and required physical-device UAT pass for release-critical workflows.
- [ ] Online banking and voice are absent from the release runtime and public capability claims.

## Review Decisions

The following are proposed in this draft and require explicit owner approval before they become master requirements:

1. Shared wallets are included in the Money Lover baseline with owner/member roles and member-attributed reporting.
2. Each expense may belong to zero or one spending jar; this prevents double counting and matches the current `budget_id` model.
3. API keys gain explicit scopes and optional expiration; existing unscoped keys require a migration/compatibility decision.
4. Savings goals and credit lifecycle are included; online bank linkage and voice remain excluded.
5. OCR stays deterministic preprocessing when a receipt is attached; model-selected tools are used for finance reads and draft proposals.
6. Automatic asset-price providers remain adapters; provider vendor choice and symbol mapping are operational decisions, not product behavior.

## Delivery Plan Draft

Implementation is gap-first: each phase begins by comparing the running behavior and current evidence with this draft, preserves verified behavior, and implements only the remaining delta.

| Phase | Outcome | Depends on | Exit gate |
| --- | --- | --- | --- |
| 0. Contract reconciliation | Owner decisions become stable requirement IDs, user stories, ADRs, and a selected parity matrix. | Approval of this draft | No unresolved scope conflict; every selected capability has an owner and proof plan. |
| 1. Finance baseline | Wallet/category/transaction/transfer/adjustment, credit, savings goal, and shared-wallet rules are complete and reachable. | Phase 0 | Golden ledger, ownership, sharing, offline and mounted UI journeys pass. |
| 2. Jars and planning | Transaction-applied jars, recurring items, events, bills, debts/repayments and planning notifications reconcile exactly. | Phase 1 | Jar assignment and planning lifecycle fixtures plus E2E/UAT pass. |
| 3. Reports and portfolio | All selected reports reconcile to golden data; asset trades, prices, P&L and net worth remain separate from cash accounting. | Phases 1-2 | Report drilldowns reconcile; deterministic portfolio replay and valuation-state tests pass. |
| 4. Public API and API keys | Stable OpenAPI plus scoped, expiring, revocable keys support approved bot/app workflows. | Phases 1-3 | Cookie/bearer parity, scope-denial, revocation, rate-limit and docs-contract suites pass. |
| 5. Agent tool platform | Intake uses clarification/draft tools and advisor uses evidence-bearing read tools with bounded context. | Phases 3-4 | Tool selection, prompt injection, read/write boundary, history, retry and provider smoke pass. |
| 6. Product reconciliation and release | Offline coverage, data lifecycle, public docs, physical-device UAT, production operations and selected parity rows are closed. | Phases 1-5 | Validation matrix is complete; no selected row remains partial/missing; release checklist passes. |

Each phase requires its own approved detail design before implementation because the combined program changes API, schema, authorization, business rules, runtime providers, and repository master documents.

## Ambiguity Report

| Dimension | Score | Minimum | Status | Notes |
| --- | ---: | ---: | --- | --- |
| Goal clarity | 0.88 | 0.75 | met | Baseline and four MyPocket differentiators are explicit. |
| Boundary clarity | 0.90 | 0.70 | met | Banking, voice, native-only integrations and commercial cloning are excluded. |
| Constraint clarity | 0.78 | 0.65 | met | Existing stack, data authority, review-first and provider boundaries are retained. |
| Acceptance clarity | 0.80 | 0.70 | met | Product and per-requirement pass/fail checks are included. |
| **Ambiguity** | **0.15** | **<= 0.20** | **gate met for review draft** | Review decisions remain intentionally unapproved. |

## Research And Interview Log

| Round | Perspective | Input | Draft decision |
| --- | --- | --- | --- |
| 1 | Researcher | Existing Money Lover audit and MyPocket code/docs | Use the 50-row evidence matrix as baseline inventory. |
| 2 | Simplifier | Reuse old docs; preserve MyPocket differences | Consolidate rather than redesign working domains. |
| 3 | Boundary keeper | Owner excluded online banking and voice | Remove them from runtime, API, agent tools and capability claims. |
| 4 | Failure analyst | Finance correctness and external-agent risk | Keep domain services authoritative, review-first writes and least-privilege keys. |

## Next Artifact After Approval

After owner review, reconcile approved clauses into `SPEC.md`, `REQUIREMENTS.md`, `USER_STORIES.md`, `ROADMAP.md`, relevant ADRs, and a phased implementation plan. No execution should begin from this draft alone.
