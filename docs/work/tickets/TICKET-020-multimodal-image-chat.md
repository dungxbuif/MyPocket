---
artifact_type: ticket
id: TICKET-020
status: ready
owner: human
priority: high
lane: high-risk
trace:
  backlog_item: BL-006
  requirement: [REQ-F-009, REQ-F-011, REQ-NF-004, REQ-NF-008]
  phase: PHASE-006
  detail_design: ../phases/PHASE-006-detail-design.md
  implementation_plan: ../../superpowers/plans/2026-08-30-phase-006-ingestion.md
  test_verification: ../test-verification/PHASE-006-ai-receipt-bank-ingestion.md
  validation_matrix: ../VALIDATION_MATRIX.md
  adrs: [../../decisions/ADR-004-review-first-ingestion.md]
  release_notes: ../../releases/CHANGELOG.md
---

# Ticket: TICKET-020 Multimodal Image Chat

## Status

- Status: ready
- Type: feature
- Priority: high
- Phase: PHASE-006

## Context

AI-chat images use a multimodal AI path, separate from receipt OCR, and converge on the same review-first draft contract.

## Acceptance Criteria

- [ ] Image chat accepts private image references and user text, then calls an OpenAI-compatible multimodal endpoint.
- [ ] Provider context is minimized to owned wallet/category labels and safe system instructions.
- [ ] Structured image-chat results create editable drafts with validation issues surfaced.
- [ ] Provider failures, unsafe content, oversized responses, and unresolved proposals never affect balances.
- [ ] Mobile chat UI distinguishes AI image proposals from receipt OCR extraction.

## Small Task Exemption

- Small task exemption: no
- Reason: Adds multimodal provider integration and sensitive data handling.
- Impact checked: API=yes, DB=yes, Security=yes, Runtime=yes, Standards=no

## Test Expectations

- Adapter contract tests for fake multimodal responses.
- Unit tests for prompt/data minimization and response validation.
- E2E tests for image chat proposal review and provider failure.

## Verification Results

- Command: not run yet
- Result: pending
- Notes: Ready for implementation; no execution evidence claimed.
