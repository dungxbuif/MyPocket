---
artifact_type: ticket
id: TICKET-014
status: ready
owner: human
priority: high
lane: high-risk
trace:
  backlog_item: BL-004
  requirement: [REQ-F-012, REQ-NF-004]
  phase: PHASE-004
  detail_design: ../phases/PHASE-004-detail-design.md
  implementation_plan: ../../superpowers/plans/2026-08-30-phase-004-planning-automation.md
  test_verification: ../test-verification/PHASE-004-planning-automation.md
  validation_matrix: ../VALIDATION_MATRIX.md
  release_notes: ../../releases/CHANGELOG.md
---

# Ticket: TICKET-014 In-App Inbox and Web Push

## Status

- Status: ready
- Type: feature
- Priority: high
- Phase: PHASE-004

## Context

In-app notifications are authoritative. Web Push is a best-effort delivery channel that may be denied, unsupported, or expired.

## Acceptance Criteria

- [ ] Notifications persist in a user-scoped inbox with read/unread state and bounded pagination.
- [ ] Budget thresholds, due obligations, and recurring drafts can create durable notices.
- [ ] Push subscriptions are created/deleted with private endpoint/key handling and redacted logs.
- [ ] Worker delivery retries are capped and expired endpoints are cleaned up.
- [ ] Mobile inbox handles permission denied, unsupported, offline, and delivery-failed states.

## Small Task Exemption

- Small task exemption: no
- Reason: Adds notification APIs, push runtime configuration, private endpoint storage, and worker delivery behavior.
- Impact checked: API=yes, DB=yes, Security=yes, Runtime=yes, Standards=no

## Test Expectations

- Unit retry/redaction tests.
- Integration tests for notification dedupe, ownership, and push subscription lifecycle.
- E2E tests for inbox, permission states, and read marking.

## Verification Results

- Command: not run yet
- Result: pending
- Notes: Ready for implementation; no execution evidence claimed.
