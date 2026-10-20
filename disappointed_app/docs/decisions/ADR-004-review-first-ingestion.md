---
artifact_type: adr
id: ADR-004
status: accepted
owner: shared
human_fields:
  - decision_approval
  - final_status
ai_fields:
  - context
  - alternatives_considered
  - consequences
  - linked_work
shared_fields:
  - decision
  - trace
trace:
  requirements: [REQ-F-007, REQ-F-008, REQ-F-009, REQ-F-010, REQ-F-011]
  phase: PHASE-006
  tickets_or_bugs: [TICKET-018, TICKET-019, TICKET-020, TICKET-021]
  detail_design: ../superpowers/specs/2026-08-23-mypocket-system-design.md
  master_docs:
    - ../requirements/SPEC.md
    - ../architecture/ARCHITECTURE.md
    - ../architecture/API.md
    - ../architecture/ERD.md
  release_notes: ../releases/CHANGELOG.md
---

# ADR-004: Review-First Shared Transaction Ingestion

## Status

- Status: accepted
- Date: 2026-08-24
- Decision approval: user-approved during design review

## Context

Text AI, receipt OCR, AI-chat images, recurring schedules, and bank notifications are probabilistic or automated inputs. Incorrect automatic confirmation would change wallet balances and reports.

## Decision

Make every generated input produce a shared TransactionDraft. Only explicit authenticated confirmation invokes the finance accounting service. Receipt camera uses the external OCR adapter; an image sent inside AI chat uses the OpenAI-compatible multimodal adapter. Voice later reuses the same draft contract.

## Alternatives Considered

- Auto-save high-confidence results: rejected because confidence is provider-specific and still risks accounting errors.
- Auto-save with undo: rejected because downstream sync, analytics, alerts, and exports can observe incorrect data.
- Separate draft types per provider: rejected because confirmation and validation rules must be consistent.

## Consequences

- Positive: one auditable validation/confirmation boundary protects accounting and simplifies provider substitution.
- Negative: every generated entry requires user review.
- Neutral: providers can return confidence metadata for UX but it cannot bypass confirmation.

## Linked Work

- Approved design: [MyPocket System Design](../superpowers/specs/2026-08-23-mypocket-system-design.md)
- Requirements: [REQUIREMENTS.md](../requirements/REQUIREMENTS.md)
- Roadmap: [ROADMAP.md](../work/ROADMAP.md)
- Architecture: [ARCHITECTURE.md](../architecture/ARCHITECTURE.md)
- API: [API.md](../architecture/API.md)
- ERD: [ERD.md](../architecture/ERD.md)
- Integrations: [INTEGRATIONS.md](../architecture/INTEGRATIONS.md)

