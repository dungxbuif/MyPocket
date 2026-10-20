---
artifact_type: traceability_matrix
id: TRACEABILITY
status: active
owner: shared
human_fields:
  - requirement_source
  - release_scope
ai_fields:
  - trace_links
  - evidence_links
  - adr_links
shared_fields:
  - matrix_rows
updated: 2026-08-31
---

# Traceability

## Field Ownership

- Human-approved requirement source and release scope are recorded in the product spec and roadmap.
- AI maintains execution, proof, decision, docs, and release links.

## Trace Matrix

| Requirement | Phase | Ticket/Bug | Detail Design | Test Verification | Docs Review | ADR/Docs | Release |
| --- | --- | --- | --- | --- | --- | --- | --- |
| [REQ-F-001](../requirements/REQUIREMENTS.md) | [PHASE-001](phases/PHASE-001-platform-identity.md) | [TICKET-003](tickets/TICKET-003-google-oauth-user-isolation.md) | [PHASE-001 detail design](phases/PHASE-001-detail-design.md), [approved design](../superpowers/specs/2026-08-23-mypocket-system-design.md) | Created during execution | Per-ticket Harness checklist | ADR-002 | M1 |
| [REQ-F-002](../requirements/REQUIREMENTS.md) | [PHASE-002](phases/PHASE-002-finance-core.md) | TICKET-005, TICKET-007 (planned) | [Approved design](../superpowers/specs/2026-08-23-mypocket-system-design.md) | Created during execution | Per-ticket Harness checklist | ADR-001 | M1 |
| [REQ-F-003](../requirements/REQUIREMENTS.md) | [PHASE-002](phases/PHASE-002-finance-core.md) | TICKET-006 (planned) | [Approved design](../superpowers/specs/2026-08-23-mypocket-system-design.md) | Created during execution | Per-ticket Harness checklist | ADR-001 | M1 |
| [REQ-F-004](../requirements/REQUIREMENTS.md) | [PHASE-003](phases/PHASE-003-offline-sync.md) | TICKET-008..010 (planned) | [Approved design](../superpowers/specs/2026-08-23-mypocket-system-design.md) | Created during execution | Per-ticket Harness checklist | ADR-003 | M1 |
| [REQ-F-005](../requirements/REQUIREMENTS.md) | [PHASE-004](phases/PHASE-004-planning-automation.md) | TICKET-011..014 (planned) | [Approved design](../superpowers/specs/2026-08-23-mypocket-system-design.md) | Created during execution | Per-ticket Harness checklist | ADR-001 | M1 |
| [REQ-F-006](../requirements/REQUIREMENTS.md) | [PHASE-005](phases/PHASE-005-analytics-dashboard.md) | TICKET-015..017 (planned) | [Approved design](../superpowers/specs/2026-08-23-mypocket-system-design.md) | Created during execution | Per-ticket Harness checklist | Architecture/API/ERD | M1 |
| [REQ-F-007](../requirements/REQUIREMENTS.md) | [PHASE-006](phases/PHASE-006-ai-receipt-bank-ingestion.md) | TICKET-018 (deferred) | [Approved design](../superpowers/specs/2026-08-23-mypocket-system-design.md) | Created during execution | Per-ticket Harness checklist | ADR-004 | M2 |
| [REQ-F-008](../requirements/REQUIREMENTS.md) | [PHASE-006](phases/PHASE-006-ai-receipt-bank-ingestion.md) | TICKET-019 (deferred) | [Approved design](../superpowers/specs/2026-08-23-mypocket-system-design.md) | Created during execution | Per-ticket Harness checklist | ADR-004 | M2 |
| [REQ-F-009](../requirements/REQUIREMENTS.md) | [PHASE-006](phases/PHASE-006-ai-receipt-bank-ingestion.md) | TICKET-020 (deferred) | [Approved design](../superpowers/specs/2026-08-23-mypocket-system-design.md) | Created during execution | Per-ticket Harness checklist | ADR-004 | M2 |
| [REQ-F-010](../requirements/REQUIREMENTS.md) | [PHASE-006](phases/PHASE-006-ai-receipt-bank-ingestion.md) | TICKET-021 (deferred) | [Approved design](../superpowers/specs/2026-08-23-mypocket-system-design.md) | Created during execution | Per-ticket Harness checklist | ADR-004 | M2 |
| [REQ-F-011](../requirements/REQUIREMENTS.md) | [PHASE-006](phases/PHASE-006-ai-receipt-bank-ingestion.md) | TICKET-018..021 (deferred) | [Approved design](../superpowers/specs/2026-08-23-mypocket-system-design.md) | Created during execution | Per-ticket Harness checklist | ADR-004 | M2 |
| [REQ-F-012](../requirements/REQUIREMENTS.md) | [PHASE-004](phases/PHASE-004-planning-automation.md) | TICKET-014 (planned) | [Approved design](../superpowers/specs/2026-08-23-mypocket-system-design.md) | Created during execution | Per-ticket Harness checklist | Integrations | M1 |
| [REQ-F-013](../requirements/REQUIREMENTS.md) | [PHASE-007](phases/PHASE-007-audit-export-production.md) | TICKET-023 (deferred) | [Approved design](../superpowers/specs/2026-08-23-mypocket-system-design.md) | Created during execution | Per-ticket Harness checklist | Architecture/API | M2 |
| [REQ-F-014](../requirements/REQUIREMENTS.md) | [PHASE-007](phases/PHASE-007-audit-export-production.md) | TICKET-024 (deferred) | [Approved design](../superpowers/specs/2026-08-23-mypocket-system-design.md) | Created during execution | Per-ticket Harness checklist | ADR-005 | M2 |
| [REQ-F-015](../requirements/REQUIREMENTS.md) | [PHASE-007](phases/PHASE-007-audit-export-production.md) | TICKET-022 (deferred) | [Approved design](../superpowers/specs/2026-08-23-mypocket-system-design.md) | Created during execution | Per-ticket Harness checklist | ADR-005 | M2 |
| [REQ-F-016](../requirements/REQUIREMENTS.md) | [PHASE-005](phases/PHASE-005-analytics-dashboard.md) | TICKET-015..017 (planned) | [Approved design](../superpowers/specs/2026-08-23-mypocket-system-design.md) | Created during execution | Per-ticket Harness checklist | Architecture | M1 |
| [REQ-F-017](../requirements/REQUIREMENTS.md) | [PHASE-008](phases/PHASE-008-deferred-voice.md) | TICKET-026 (planned) | [Approved design](../superpowers/specs/2026-08-23-mypocket-system-design.md) | Created during execution | Per-ticket Harness checklist | SDD | M3 |
| [REQ-NF-001](../requirements/REQUIREMENTS.md) | [PHASE-001](phases/PHASE-001-platform-identity.md) | [TICKET-003](tickets/TICKET-003-google-oauth-user-isolation.md) | [PHASE-001 detail design](phases/PHASE-001-detail-design.md), [approved design](../superpowers/specs/2026-08-23-mypocket-system-design.md) | Created during execution | Per-ticket Harness checklist | ADR-002 | M1 |
| [REQ-NF-002](../requirements/REQUIREMENTS.md) | [PHASE-002](phases/PHASE-002-finance-core.md)/003/004 | TICKET-006, TICKET-009, TICKET-013 (planned) | [Approved design](../superpowers/specs/2026-08-23-mypocket-system-design.md) | Created during execution | Per-ticket Harness checklist | ADR-003 | M1 |
| [REQ-NF-003](../requirements/REQUIREMENTS.md) | [PHASE-003](phases/PHASE-003-offline-sync.md) | TICKET-009..010 (planned) | [Approved design](../superpowers/specs/2026-08-23-mypocket-system-design.md) | Created during execution | Per-ticket Harness checklist | ADR-003 | M1 |
| [REQ-NF-004](../requirements/REQUIREMENTS.md) | [PHASE-001](phases/PHASE-001-platform-identity.md)/006/007 | [TICKET-003](tickets/TICKET-003-google-oauth-user-isolation.md), TICKET-018..022 (deferred) | [PHASE-001 detail design](phases/PHASE-001-detail-design.md), [approved design](../superpowers/specs/2026-08-23-mypocket-system-design.md) | Created during execution | Per-ticket Harness checklist | ADR-002/004/005 | M1/M2 |
| [REQ-NF-005](../requirements/REQUIREMENTS.md) | [PHASE-002](phases/PHASE-002-finance-core.md)/004/005 | TICKET-006, TICKET-011, TICKET-017 (planned) | [Approved design](../superpowers/specs/2026-08-23-mypocket-system-design.md) | Created during execution | Per-ticket Harness checklist | SDD | M1 |
| [REQ-NF-006](../requirements/REQUIREMENTS.md) | [PHASE-001](phases/PHASE-001-platform-identity.md)/007 | [TICKET-002](tickets/TICKET-002-postgresql-migrations-s3-platform-adapters.md), [TICKET-004](tickets/TICKET-004-development-operations-verification-baseline.md), TICKET-025 (deferred) | [PHASE-001 detail design](phases/PHASE-001-detail-design.md), [approved design](../superpowers/specs/2026-08-23-mypocket-system-design.md) | Created during execution | Per-ticket Harness checklist | ADR-001 | M1/M2 |
| [REQ-NF-007](../requirements/REQUIREMENTS.md) | All phases | All planned tickets (planned) | [Approved design](../superpowers/specs/2026-08-23-mypocket-system-design.md) | Created during execution | Per-ticket Harness checklist | Validation matrix | M1 |
| [REQ-NF-008](../requirements/REQUIREMENTS.md) | [PHASE-001](phases/PHASE-001-platform-identity.md)/003/006 | [TICKET-001](tickets/TICKET-001-repository-runtime-foundation.md), TICKET-009, TICKET-018..021 (deferred) | [PHASE-001 detail design](phases/PHASE-001-detail-design.md), [approved design](../superpowers/specs/2026-08-23-mypocket-system-design.md) | Created during execution | Per-ticket Harness checklist | API | M1/M2 |

## Rules

- Planned ticket IDs become links when their Harness ticket artifacts are created.
- Test verification remains unlinked until real commands and outcomes exist.
- Requirements cannot be marked implemented from design or plan evidence alone.
- Every durable implementation divergence updates the SDD/master docs and an ADR before execution continues.
