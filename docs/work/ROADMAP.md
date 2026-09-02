---
artifact_type: roadmap
id: ROADMAP
status: active
owner: human
human_fields:
  - milestones
  - priority
  - phase_order
ai_fields:
  - phase_links
  - status_summaries
shared_fields:
  - milestone_status
updated: 2026-08-31
---

# Roadmap

## Field Ownership

- Human approved the milestone scope and phase order during design review.
- AI maintains phase links and evidence-backed status summaries.x`

## Milestones

| Milestone | Goal | Status | Phase Files |
| --- | --- | --- | --- |
| M0 | Approve system design and migrate product planning into Harness | done | [SDD](../architecture/SDD.md), [requirements](../requirements/REQUIREMENTS.md), [traceability](TRACEABILITY.md), [docs review](DOCS-REVIEW-M0.md) |
| M1 | Deliver the current MyPocket release through TICKET-017 plus the approved TICKET-027 asset-portfolio extension | planned | [PHASE-001](phases/PHASE-001-platform-identity.md), [PHASE-002](phases/PHASE-002-finance-core.md), [PHASE-003](phases/PHASE-003-offline-sync.md), [PHASE-004](phases/PHASE-004-planning-automation.md), [PHASE-005](phases/PHASE-005-analytics-dashboard.md) |
| M2 | Deliver post-M1 ingestion, audit, export, account lifecycle, and production-hardening work | deferred | [PHASE-006](phases/PHASE-006-ai-receipt-bank-ingestion.md), [PHASE-007](phases/PHASE-007-audit-export-production.md) |
| M3 | Add deferred voice transaction entry after post-M1 draft infrastructure is verified | deferred | [PHASE-008](phases/PHASE-008-deferred-voice.md) |

## Approved Phase Order

1. PHASE-001 Platform and Identity
2. PHASE-002 Finance Core
3. PHASE-003 Offline Synchronization
4. PHASE-004 Planning and Automation
5. PHASE-005 Analytics and Dashboard
6. PHASE-006 AI, Receipt, and Bank Ingestion (deferred to M2)
7. PHASE-007 Audit, Export, Account Lifecycle, and Production Operations (deferred to M2)
8. PHASE-008 Deferred Voice Input (deferred to M3)

## Milestone Gates

- M0 closed on 2026-08-24 after Harness link, YAML, placeholder, whitespace, and docs-review validation passed.
- M1 begins only after PHASE-001 ticket artifacts and implementation plan are reviewed.
- Each M1 phase must be independently deployable and verified before the next dependent phase is promoted to `ready`.
- M1 closes after PHASE-001 through PHASE-005 meet completion rules, including TICKET-017 and the human-promoted TICKET-027 asset-portfolio extension.
- TICKET-027 executes immediately after TICKET-017 only after its detail design is approved; TICKET-018 through TICKET-026 remain deferred.
- M2 remains deferred until a human explicitly promotes TICKET-018 through TICKET-025 after M1.
- M3 remains deferred until M2 draft infrastructure is verified and a human explicitly promotes voice work.
