---
artifact_type: ticket
id: TICKET-018
status: deferred
owner: human
priority: high
lane: high-risk
trace:
  backlog_item: BL-006
  requirement: [REQ-F-007, REQ-F-011, REQ-NF-004, REQ-NF-008]
  phase: PHASE-006
  detail_design: ../phases/PHASE-006-detail-design.md
  implementation_plan: ../../superpowers/plans/2026-08-30-phase-006-ingestion.md
  test_verification: ../test-verification/PHASE-006-ai-receipt-bank-ingestion.md
  validation_matrix: ../VALIDATION_MATRIX.md
  adrs: [../../decisions/ADR-004-review-first-ingestion.md]
  release_notes: ../../releases/CHANGELOG.md
---

# Ticket: TICKET-018 Shared Drafts and Text AI Chat

## Status

- Status: deferred
- Type: feature
- Priority: high
- Phase: PHASE-006

## Context

Text AI can propose transactions, but all proposals must remain editable drafts until explicit authenticated confirmation.

## Acceptance Criteria

- [ ] Shared transaction draft schema supports source, normalized proposal, validation issues, status, version, and confirmation transaction ID.
- [ ] Text AI adapter uses OpenAI-compatible structured responses with bounded owned wallet/category context.
- [ ] Single, multi-entry, and transfer proposals create editable drafts, not confirmed transactions.
- [ ] Confirmation revalidates ownership/version and calls PHASE-002 accounting exactly once with idempotency.
- [ ] Malformed, timed-out, adversarial, or unresolved provider outputs cannot affect balances.

## Small Task Exemption

- Small task exemption: no
- Reason: Adds provider integration, draft data model, API, and confirmation safety boundary.
- Impact checked: API=yes, DB=yes, Security=yes, Runtime=yes, Standards=no

## Test Expectations

- Unit schema, normalization, prompt-boundary, redaction, and state-machine tests.
- Adapter contract tests for fake OpenAI-compatible success/failure responses.
- Integration/E2E tests for draft create/edit/reject/confirm.

## Verification Results

- Command: not run yet
- Result: pending
- Notes: Deferred to M2/post-M1 by human scope decision on 2026-08-31; no execution evidence claimed.
