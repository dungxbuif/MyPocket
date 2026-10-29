---
artifact_type: ticket
id: TICKET-022
status: deferred
owner: human
priority: high
lane: high-risk
trace:
  backlog_item: BL-007
  requirement: [REQ-F-015, REQ-NF-004, REQ-NF-007]
  phase: PHASE-007
  detail_design: ../phases/PHASE-007-detail-design.md
  implementation_plan: ../../superpowers/plans/2026-08-30-phase-007-production.md
  test_verification: ../test-verification/PHASE-007-audit-export-production.md
  validation_matrix: ../VALIDATION_MATRIX.md
  release_notes: ../../releases/CHANGELOG.md
---

# Ticket: TICKET-022 Append-Only Audit Pipeline and Hidden Viewer

## Status

- Status: deferred
- Type: feature
- Priority: high
- Phase: PHASE-007

## Context

Sensitive state/security operations need redacted audit records and a backend-authorized hidden viewer restricted to `AUDIT_VIEWER_EMAIL`.

## Acceptance Criteria

- [ ] Approved state/security actions emit append-only redacted audit events with correlation IDs.
- [ ] Audit storage rejects update/delete through application paths.
- [ ] `/api/v1/audit/events` and hidden viewer require exact normalized verified Google email from `AUDIT_VIEWER_EMAIL`.
- [ ] Viewer is absent from navigation and remains read-only with bounded filters/pagination.
- [ ] Retention purge defaults to 180 days and uses bounded worker batches without recursive per-row audit.

## Small Task Exemption

- Small task exemption: no
- Reason: Adds security-sensitive audit data, authorization, retention, and hidden UI.
- Impact checked: API=yes, DB=yes, Security=yes, Runtime=yes, Standards=no

## Test Expectations

- Unit redaction/action mapping tests.
- Integration tests for immutability, authorization, retention.
- E2E/security tests for allowed/denied viewer access and hidden navigation.

## Verification Results

- Command: not run yet
- Result: pending
- Notes: Deferred to M2/post-M1 by human scope decision on 2026-08-31; no execution evidence claimed.
