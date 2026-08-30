---
artifact_type: backlog
id: BACKLOG
status: active
owner: shared
human_fields:
  - priority_override
  - rank
  - blocker_decisions
ai_fields:
  - risk_flags
  - lane_recommendation
  - next_artifact
  - notes
shared_fields:
  - queue_items
  - status
updated: 2026-08-30
---

# Backlog

## Field Ownership

- Human approved the product scope and phase order.
- Human may override rank or block a phase.
- AI maintains readiness, risk flags, links, and next-artifact recommendations.

## Queue Rules

- Rank is dependency-aware; urgent bugs or human priority overrides may reorder ready work.
- High-risk phases require tickets, approved detail design, and verification planning before execution.
- A phase is not executable directly from this table.
- Status follows `draft -> ready -> in_progress -> blocked -> in_review -> verified -> done`.

## Items

| Rank | ID | Type | Lane | Title | Priority | Status | Links | Risk Flags | Next Artifact | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 1 | BL-001 | phase | high-risk | Platform and Identity | Urgent | in_review | [PHASE-001](phases/PHASE-001-platform-identity.md), [detail design](phases/PHASE-001-detail-design.md), [plan](../superpowers/plans/2026-08-25-phase-001-platform-identity.md) | Auth, Authorization, Data model, Public API, Deployment/runtime | Human review, then mark done | Runtime/API foundation, PostgreSQL migrations, S3-compatible adapter, backend identity, mobile PWA shell, offline/auth browser E2E, backend/frontend split, and local compose smoke are verified. |
| 2 | BL-002 | phase | high-risk | Finance Core | High | in_progress | [PHASE-002](phases/PHASE-002-finance-core.md), [detail design](phases/PHASE-002-detail-design.md), [plan](../superpowers/plans/2026-08-30-phase-002-finance-core.md), [TICKET-005](tickets/TICKET-005-wallet-category-domain.md), [TICKET-006](tickets/TICKET-006-transaction-accounting-engine.md), [TICKET-007](tickets/TICKET-007-vietnamese-seeds-receipt-metadata.md) | Data model, Existing behavior, Public API, Multi-domain | Add dedicated live finance CRUD E2E/UAT, then move PHASE-002 to review | Wallet/category and transaction backend/API, mobile finance manager/editor component workflows, offline transaction outbox, and receipt metadata foundation have automated proof. |
| 3 | BL-003 | phase | high-risk | Offline Synchronization | High | open | [PHASE-003](phases/PHASE-003-offline-sync.md), [detail design](phases/PHASE-003-detail-design.md) | Data model, Data loss, Public API, Multi-domain | Create TICKET-008..010 artifacts and implementation plan | Approved design locks IndexedDB mirror, ordered mutation API, change cursor, tombstones, explicit conflicts, and full resync. |
| 4 | BL-004 | phase | high-risk | Planning and Automation | High | open | [PHASE-004](phases/PHASE-004-planning-automation.md), [detail design](phases/PHASE-004-detail-design.md) | Data model, Jobs, Web Push, Multi-domain | Create TICKET-011..014 artifacts and implementation plan | Approved design locks budget periods/alerts, events, debts, recurring drafts, worker leases, inbox, and best-effort push. |
| 5 | BL-005 | phase | normal | Analytics and Dashboard | High | open | [PHASE-005](phases/PHASE-005-analytics-dashboard.md), [detail design](phases/PHASE-005-detail-design.md) | Existing behavior, Weak proof, Multi-domain | Create TICKET-015..017 artifacts and implementation plan | Approved design locks server-side formulas, report APIs, cached stale state, accessible charts, and mobile/desktop behavior. |
| 6 | BL-006 | phase | high-risk | AI, Receipt, and Bank Ingestion | High | open | [PHASE-006](phases/PHASE-006-ai-receipt-bank-ingestion.md), [detail design](phases/PHASE-006-detail-design.md) | External provider, Security/privacy, Webhook, Public API | Create TICKET-018..021 artifacts and implementation plan | Approved design routes every untrusted provider source through one review-first draft and explicit confirmation contract. |
| 7 | BL-007 | phase | high-risk | Audit, Export, Account Lifecycle, and Production Operations | High | open | [PHASE-007](phases/PHASE-007-audit-export-production.md), [detail design](phases/PHASE-007-detail-design.md) | Authorization, Data deletion, Audit/privacy, Deployment/runtime | Create TICKET-022..025 artifacts and implementation plan | Approved design locks append-only redacted audit, exact-email viewer, exports, safe lifecycle jobs, and backup/restore release proof. |
| 8 | BL-008 | phase | high-risk | Deferred Voice Input | Low | deferred | [PHASE-008](phases/PHASE-008-deferred-voice.md) | External provider, Privacy, Browser compatibility | TICKET-026 after v1 verification | Explicitly excluded from the initial release. |
