---
artifact_type: ticket
id: TICKET-021
status: deferred
owner: human
priority: high
lane: high-risk
trace:
  backlog_item: BL-006
  requirement: [REQ-F-010, REQ-F-011, REQ-NF-004, REQ-NF-008]
  phase: PHASE-006
  detail_design: ../phases/PHASE-006-detail-design.md
  implementation_plan: ../../superpowers/plans/2026-08-30-phase-006-ingestion.md
  test_verification: ../test-verification/PHASE-006-ai-receipt-bank-ingestion.md
  validation_matrix: ../VALIDATION_MATRIX.md
  adrs: [../../decisions/ADR-004-review-first-ingestion.md]
  release_notes: ../../releases/CHANGELOG.md
---

# Ticket: TICKET-021 Signed Bank-Notification Webhook

## Status

- Status: deferred
- Type: feature
- Priority: high
- Phase: PHASE-006

## Context

Bank notifications enter without user cookies, so per-source HMAC, replay protection, and deduplication are the safety boundary.

## Acceptance Criteria

- [ ] Webhook endpoint verifies source, HMAC, timestamp window, nonce, and raw-body signature.
- [ ] Valid source configuration maps to exactly one user and never trusts user identity from payload content.
- [ ] Duplicate payloads and replayed nonces do not create additional drafts.
- [ ] Invalid/replayed requests return stable safe errors and redacted audit evidence.
- [ ] Valid notifications create reviewable drafts or inbox notices, never confirmed transactions.

## Small Task Exemption

- Small task exemption: no
- Reason: Adds unauthenticated provider webhook surface, HMAC security, and draft ingestion.
- Impact checked: API=yes, DB=yes, Security=yes, Runtime=yes, Standards=no

## Test Expectations

- Unit HMAC, timestamp, nonce, and redaction tests.
- Integration tests for dedupe, user mapping, draft creation, and audit events.
- E2E/manual webhook fixture review flow.

## Verification Results

- Command: not run yet
- Result: pending
- Notes: Deferred to M2/post-M1 by human scope decision on 2026-08-31; no execution evidence claimed.
