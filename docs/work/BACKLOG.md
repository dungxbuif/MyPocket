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
updated: 2026-08-24
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
| 1 | BL-001 | phase | high-risk | Platform and Identity | Urgent | open | [PHASE-001](phases/PHASE-001-platform-identity.md) | Auth, Authorization, Data model, Public API, Deployment/runtime | TICKET-001..004 and implementation plan | Active queue focus; no code until ticket/plan review. |
| 2 | BL-002 | phase | high-risk | Finance Core | High | open | [PHASE-002](phases/PHASE-002-finance-core.md) | Data model, Existing behavior, Public API, Multi-domain | TICKET-005..007 and implementation plan | Depends on PHASE-001. |
| 3 | BL-003 | phase | high-risk | Offline Synchronization | High | open | [PHASE-003](phases/PHASE-003-offline-sync.md) | Data model, Data loss, Public API, Multi-domain | TICKET-008..010 and implementation plan | Depends on versioned PHASE-002 domain commands. |
| 4 | BL-004 | phase | high-risk | Planning and Automation | High | open | [PHASE-004](phases/PHASE-004-planning-automation.md) | Data model, Jobs, Web Push, Multi-domain | TICKET-011..014 and implementation plan | Depends on finance and sync contracts. |
| 5 | BL-005 | phase | normal | Analytics and Dashboard | High | open | [PHASE-005](phases/PHASE-005-analytics-dashboard.md) | Existing behavior, Weak proof, Multi-domain | TICKET-015..017 and implementation plan | Depends on finance/planning data. |
| 6 | BL-006 | phase | high-risk | AI, Receipt, and Bank Ingestion | High | open | [PHASE-006](phases/PHASE-006-ai-receipt-bank-ingestion.md) | External provider, Security/privacy, Webhook, Public API | TICKET-018..021 and implementation plan | All generated results remain drafts. |
| 7 | BL-007 | phase | high-risk | Audit, Export, Account Lifecycle, and Production Operations | High | open | [PHASE-007](phases/PHASE-007-audit-export-production.md) | Authorization, Data deletion, Audit/privacy, Deployment/runtime | TICKET-022..025 and implementation plan | Requires backup/restore proof before destructive jobs release. |
| 8 | BL-008 | phase | high-risk | Deferred Voice Input | Low | deferred | [PHASE-008](phases/PHASE-008-deferred-voice.md) | External provider, Privacy, Browser compatibility | TICKET-026 after v1 verification | Explicitly excluded from the initial release. |

