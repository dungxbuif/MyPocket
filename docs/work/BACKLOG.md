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
updated: 2026-08-31
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
| 2 | BL-002 | phase | high-risk | Finance Core | High | in_review | [PHASE-002](phases/PHASE-002-finance-core.md), [detail design](phases/PHASE-002-detail-design.md), [plan](../superpowers/plans/2026-08-30-phase-002-finance-core.md), [TICKET-005](tickets/TICKET-005-wallet-category-domain.md), [TICKET-006](tickets/TICKET-006-transaction-accounting-engine.md), [TICKET-007](tickets/TICKET-007-vietnamese-seeds-receipt-metadata.md) | Data model, Existing behavior, Public API, Multi-domain | Human/UAT review, then mark verified if accepted | Wallet/category and transaction backend/API, mobile finance manager/editor component workflows, live finance CRUD E2E, offline transaction outbox, and receipt metadata foundation have automated proof. |
| 3 | BL-003 | phase | high-risk | Offline Synchronization | High | in_progress | [PHASE-003](phases/PHASE-003-offline-sync.md), [detail design](phases/PHASE-003-detail-design.md), [plan](../superpowers/plans/2026-08-30-phase-003-offline-sync.md), [TICKET-008](tickets/TICKET-008-indexeddb-mirror-outbox.md), [TICKET-009](tickets/TICKET-009-sync-api-change-feed.md), [TICKET-010](tickets/TICKET-010-conflict-inbox-recovery.md) | Data model, Data loss, Public API, Multi-domain | Execute TICKET-010 conflict inbox/recovery | TICKET-008 frontend IndexedDB mirror/outbox and TICKET-009 backend sync API/change feed have automated proof and are in review; conflict inbox, final replay/conflict UAT, and PHASE-003 verification remain pending. |
| 4 | BL-004 | phase | high-risk | Planning and Automation | High | ready | [PHASE-004](phases/PHASE-004-planning-automation.md), [detail design](phases/PHASE-004-detail-design.md), [plan](../superpowers/plans/2026-08-30-phase-004-planning-automation.md), [TICKET-011](tickets/TICKET-011-budgets-threshold-alerts.md), [TICKET-012](tickets/TICKET-012-events-debts-repayments.md), [TICKET-013](tickets/TICKET-013-recurring-schedules-worker.md), [TICKET-014](tickets/TICKET-014-inbox-web-push.md) | Data model, Jobs, Web Push, Multi-domain | Execute after PHASE-003 verification | Ready design/plan locks budget periods/alerts, events, debts, recurring drafts, worker leases, inbox, and best-effort push. |
| 5 | BL-005 | phase | normal | Analytics and Dashboard | High | ready | [PHASE-005](phases/PHASE-005-analytics-dashboard.md), [detail design](phases/PHASE-005-detail-design.md), [plan](../superpowers/plans/2026-08-30-phase-005-analytics-dashboard.md), [TICKET-015](tickets/TICKET-015-pwa-navigation-search-wallet-views.md), [TICKET-016](tickets/TICKET-016-overview-net-worth-dashboard.md), [TICKET-017](tickets/TICKET-017-analytics-reports-cumulative-trends.md) | Existing behavior, Weak proof, Multi-domain | Execute after PHASE-004 verification | Ready design/plan locks server-side formulas, report APIs, cached stale state, accessible charts, and mobile/desktop behavior. |
| 6 | BL-006 | phase | high-risk | AI, Receipt, and Bank Ingestion | High | deferred | [PHASE-006](phases/PHASE-006-ai-receipt-bank-ingestion.md), [detail design](phases/PHASE-006-detail-design.md), [plan](../superpowers/plans/2026-08-30-phase-006-ingestion.md), [TICKET-018](tickets/TICKET-018-shared-drafts-text-ai-chat.md), [TICKET-019](tickets/TICKET-019-receipt-capture-ocr-adapter.md), [TICKET-020](tickets/TICKET-020-multimodal-image-chat.md), [TICKET-021](tickets/TICKET-021-signed-bank-webhook.md) | External provider, Security/privacy, Webhook, Public API | Resume in M2 after TICKET-017 verification | Deferred by human scope decision on 2026-08-31. Design and tickets remain available, but TICKET-018 onward is outside the current M1 release. |
| 7 | BL-007 | phase | high-risk | Audit, Export, Account Lifecycle, and Production Operations | High | deferred | [PHASE-007](phases/PHASE-007-audit-export-production.md), [detail design](phases/PHASE-007-detail-design.md), [plan](../superpowers/plans/2026-08-30-phase-007-production.md), [TICKET-022](tickets/TICKET-022-audit-pipeline-hidden-viewer.md), [TICKET-023](tickets/TICKET-023-manual-export-jobs.md), [TICKET-024](tickets/TICKET-024-account-reset-deletion.md), [TICKET-025](tickets/TICKET-025-homelab-production-release-proof.md) | Authorization, Data deletion, Audit/privacy, Deployment/runtime | Resume in M2 after PHASE-006 is promoted | Deferred by human scope decision on 2026-08-31. Design and tickets remain available, but TICKET-022 onward is outside the current M1 release. |
| 8 | BL-008 | phase | high-risk | Deferred Voice Input | Low | deferred | [PHASE-008](phases/PHASE-008-deferred-voice.md) | External provider, Privacy, Browser compatibility | TICKET-026 after post-M1 draft infrastructure | Explicitly excluded from the current M1 release and now sequenced after M2. |
