---
artifact_type: phase
id: PHASE-006
status: draft
owner: human
priority: High
human_fields:
  - goal
  - scope
  - out_of_scope
  - priority
  - success_criteria
ai_fields:
  - risks
  - dependencies
  - verification_plan
  - completion_summary
shared_fields:
  - status
  - trace
  - tickets_and_bugs
trace:
  backlog_items:
    - BL-006
  roadmap: ../ROADMAP.md
  detail_design: PHASE-006-detail-design.md
  requirements:
    - REQ-F-007
    - REQ-F-008
    - REQ-F-009
    - REQ-F-010
    - REQ-F-011
    - REQ-NF-004
    - REQ-NF-008
  tickets:
    - TICKET-018
    - TICKET-019
    - TICKET-020
    - TICKET-021
  bugs: []
  test_verification: not_created_phase_not_executed
  validation_matrix: ../VALIDATION_MATRIX.md
  adrs: []
  release_notes: ../../releases/CHANGELOG.md
---

# PHASE-006: AI, Receipt, and Bank Ingestion

## Status

- ID: PHASE-006
- Status: draft
- Owner: human
- Priority: High
- Created: 2026-08-24
- Updated: 2026-08-24

## Trace Links

- Backlog: [BACKLOG.md](../BACKLOG.md)
- Roadmap: [ROADMAP.md](../ROADMAP.md)
- Detail design: [PHASE-006-detail-design.md](PHASE-006-detail-design.md) — approved 2026-08-30
- Requirements: [REQUIREMENTS.md](../../requirements/REQUIREMENTS.md) — REQ-F-007, REQ-F-008, REQ-F-009, REQ-F-010, REQ-F-011, REQ-NF-004, REQ-NF-008
- Test verification: created during execution
- Validation matrix: [VALIDATION_MATRIX.md](../VALIDATION_MATRIX.md)
- ADRs: [decisions](../../decisions/README.md)
- Release notes: [CHANGELOG.md](../../releases/CHANGELOG.md)

## Goal

Deliver text AI, receipt OCR, AI-chat images, signed bank webhooks, and one safe draft review/confirmation contract.

## Scope

- Provider-independent transaction draft schema, validation, edit, rejection, and idempotent confirmation.
- OpenAI-compatible text chat for single/multiple transactions and transfers using owned wallet/category context.
- Add-transaction camera/upload to private S3, external OCR extraction, and draft creation.
- AI-conversation image attachment sent to an OpenAI-compatible multimodal endpoint.
- Per-source HMAC bank webhook with timestamp/nonce/replay checks, deduplication, redacted auditing, and review notice.
- Provider timeouts, capped retries, rate limits, structured errors, and prompt-injection boundaries.

## Out Of Scope

- Voice recording/transcription.
- Model training or fine-tuning.
- Automatic confirmation at any confidence threshold.

## Tickets And Bugs

| ID | Type | Title | Status | Link |
| --- | --- | --- | --- | --- |
| TICKET-018 | Ticket | Shared drafts and text AI chat | planned | Created during implementation planning |
| TICKET-019 | Ticket | Receipt capture and OCR adapter | planned | Created during implementation planning |
| TICKET-020 | Ticket | Multimodal image chat | planned | Created during implementation planning |
| TICKET-021 | Ticket | Signed bank-notification webhook | planned | Created during implementation planning |

## Dependencies

- PHASE-002 finance confirmation service and receipt metadata.
- PHASE-004 inbox/Web Push.
- Production-compatible provider URLs, models, and keys supplied through environment configuration during deployment.

## Risks

- Untrusted model/provider output can violate domain rules or expose prompt context.
- Receipt images and notification text contain sensitive financial data.
- Retries or webhook replay can duplicate drafts.

## Success Criteria

- Text chat extracts one or many editable proposals without confirming them.
- Receipt camera uses the OCR provider path; AI-chat images use the multimodal AI path; both converge on shared drafts.
- Malformed, unresolved, timed-out, or adversarial provider results cannot affect wallet balances.
- Invalid/replayed webhooks are rejected and duplicates produce no additional draft.
- Only explicit authenticated confirmation invokes accounting exactly once.

## Verification Plan

- Go unit tests for schemas, normalization, resolution, redaction, HMAC, replay windows, and confirmation rules.
- Adapter contract tests for text AI, multimodal AI, OCR, S3, and failure mapping.
- PostgreSQL integration tests for drafts, idempotent confirmation, webhook deduplication, ownership, and audit events.
- E2E/UAT for text, multi-entry, transfer, receipt-camera, AI-chat image, webhook review, edit, confirm, and provider failure.

## Gate Checklist

- [x] Phase links approved requirements
- [x] Tickets have stable planned IDs and bounded titles
- [x] Risks and dependencies are recorded
- [x] Verification plan is defined
- [x] Release/changelog need is linked
- [x] Detail design is approved
- [ ] Ticket artifacts and detailed implementation plan are created
- [ ] Phase status is promoted to ready after plan review

## Completion Summary

No implementation has started. Completion evidence will be recorded after the phase reaches execution.
