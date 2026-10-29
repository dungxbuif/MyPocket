---
artifact_type: ticket
id: TICKET-014
status: in_review
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

- Status: in_review
- Type: feature
- Priority: high
- Phase: PHASE-004

## Context

In-app notifications are authoritative. Web Push is a best-effort delivery channel that may be denied, unsupported, or expired.

## Acceptance Criteria

- [x] Notifications persist in a user-scoped inbox with read/unread state and bounded pagination.
- [x] Budget thresholds, due obligations, and recurring drafts can create durable notices.
- [x] Push subscriptions are created/deleted with private endpoint/key handling and redacted logs.
- [x] Worker delivery retries are capped and expired endpoints are cleaned up.
- [x] Mobile inbox handles permission denied, unsupported, offline, and delivery-failed states.

## Implementation Decision

- Use a dedicated `notification` package backed by PostgreSQL. Notification creation is idempotent on `(user_id, dedupe_key)` and list pagination is cursor-based with a server-side maximum of 50 rows.
- Store push endpoint and encryption keys only in the server database; API responses expose subscription metadata but never return keys. Delivery is injected behind a small interface so the worker can classify success, expired endpoints, and retryable failures without coupling domain code to a provider.
- Push is best effort: durable inbox insertion succeeds independently, retries stop after five attempts, and expired/invalid subscriptions are deleted.

## Small Task Exemption

- Small task exemption: no
- Reason: Adds notification APIs, push runtime configuration, private endpoint storage, and worker delivery behavior.
- Impact checked: API=yes, DB=yes, Security=yes, Runtime=yes, Standards=no

## Test Expectations

- Unit retry/redaction tests.
- Integration tests for notification dedupe, ownership, and push subscription lifecycle.
- E2E tests for inbox, permission states, and read marking.

## Verification Results

- Command: `GOCACHE=/private/tmp/mypocket-go-cache go test ./...` (backend), `npm test -- --run`, `npm run build` (frontend)
- Result: pass for package/component/build proof; PostgreSQL integration requires local test database
- Notes: Notification API ownership/read tests and retry/redaction unit tests pass; mobile UI exposes denied, unsupported, offline, and failed states.
