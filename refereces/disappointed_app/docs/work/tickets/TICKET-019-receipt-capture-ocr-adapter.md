---
artifact_type: ticket
id: TICKET-019
status: deferred
owner: human
priority: high
lane: high-risk
trace:
  backlog_item: BL-006
  requirement: [REQ-F-008, REQ-F-011, REQ-NF-004, REQ-NF-008]
  phase: PHASE-006
  detail_design: ../phases/PHASE-006-detail-design.md
  implementation_plan: ../../superpowers/plans/2026-08-30-phase-006-ingestion.md
  test_verification: ../test-verification/PHASE-006-ai-receipt-bank-ingestion.md
  validation_matrix: ../VALIDATION_MATRIX.md
  adrs: [../../decisions/ADR-004-review-first-ingestion.md]
  release_notes: ../../releases/CHANGELOG.md
---

# Ticket: TICKET-019 Receipt Capture and OCR Adapter

## Status

- Status: deferred
- Type: feature
- Priority: high
- Phase: PHASE-006

## Context

Receipt camera/upload must store images privately and use the OCR provider path to create reviewable drafts.

## Acceptance Criteria

- [ ] Receipt upload validates size, content type, checksum, ownership, and private S3 object key.
- [ ] OCR adapter maps provider success, malformed, timeout, oversized, retryable, and terminal failures safely.
- [ ] Extracted values create one or more shared drafts with explicit unresolved fields.
- [ ] Receipt images are never exposed through public URLs and logs redact object keys where needed.
- [ ] Mobile quick-add receipt flow supports capture/upload, pending extraction, retry, edit, reject, and confirm.

## Small Task Exemption

- Small task exemption: no
- Reason: Adds private object handling, external OCR integration, and sensitive draft creation.
- Impact checked: API=yes, DB=yes, Security=yes, Runtime=yes, Standards=no

## Test Expectations

- S3/object validation tests.
- OCR fake adapter contract tests.
- Integration/E2E tests for upload, extract, draft review, and failure states.

## Verification Results

- Command: not run yet
- Result: pending
- Notes: Deferred to M2/post-M1 by human scope decision on 2026-08-31; no execution evidence claimed.
