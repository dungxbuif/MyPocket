---
artifact_type: phase
id: PHASE-007
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
    - BL-007
  roadmap: ../ROADMAP.md
  detail_design: PHASE-007-detail-design.md
  requirements:
    - REQ-F-013
    - REQ-F-014
    - REQ-F-015
    - REQ-NF-004
    - REQ-NF-006
    - REQ-NF-007
  tickets:
    - TICKET-022
    - TICKET-023
    - TICKET-024
    - TICKET-025
  bugs: []
  test_verification: not_created_phase_not_executed
  validation_matrix: ../VALIDATION_MATRIX.md
  adrs: []
  release_notes: ../../releases/CHANGELOG.md
---

# PHASE-007: Audit, Export, Account Lifecycle, and Production Operations

## Status

- ID: PHASE-007
- Status: draft
- Owner: human
- Priority: High
- Created: 2026-08-24
- Updated: 2026-08-24

## Trace Links

- Backlog: [BACKLOG.md](../BACKLOG.md)
- Roadmap: [ROADMAP.md](../ROADMAP.md)
- Detail design: [PHASE-007-detail-design.md](PHASE-007-detail-design.md) — approved 2026-08-30
- Requirements: [REQUIREMENTS.md](../../requirements/REQUIREMENTS.md) — REQ-F-013, REQ-F-014, REQ-F-015, REQ-NF-004, REQ-NF-006, REQ-NF-007
- Test verification: created during execution
- Validation matrix: [VALIDATION_MATRIX.md](../VALIDATION_MATRIX.md)
- ADRs: [decisions](../../decisions/README.md)
- Release notes: [CHANGELOG.md](../../releases/CHANGELOG.md)

## Goal

Deliver restricted audit debugging, manual exports, safe account lifecycle, and verified homelab production operations.

## Scope

- Append-only state-changing/security audit events with redacted diffs and correlation IDs.
- Hidden read-only audit page and API authorized only by exact `AUDIT_VIEWER_EMAIL`.
- Configurable audit retention with default 180 days and bounded worker purge.
- Manual CSV and Google Sheets-compatible snapshot export.
- Confirmed account reset/delete background jobs including S3 objects and lifecycle audit evidence.
- Production configuration, migrations, health, backup/restore, worker restart, rollback guidance, and release verification.

## Out Of Scope

- Audit of ordinary clicks, page views, filters, or searches.
- Spreadsheet import/synchronization.
- A general administrator role or device management.

## Tickets And Bugs

| ID | Type | Title | Status | Link |
| --- | --- | --- | --- | --- |
| TICKET-022 | Ticket | Append-only audit pipeline and hidden viewer | planned | Created during implementation planning |
| TICKET-023 | Ticket | Manual export jobs | planned | Created during implementation planning |
| TICKET-024 | Ticket | Account reset and deletion | planned | Created during implementation planning |
| TICKET-025 | Ticket | Homelab production hardening and release proof | planned | Created during implementation planning |

## Dependencies

- PHASE-001 operations/security foundation and all domain actions that emit audit records.
- PHASE-006 provider operations requiring audit visibility.
- Production connection string and credentials supplied by the homelab operator for final platform verification.

## Risks

- Over-logging can expose secrets or sensitive payloads.
- Hidden UI without backend authorization would be insecure.
- Deletion and retention jobs are destructive and require exact scoping and backup proof.

## Success Criteria

- All approved state/security events produce redacted append-only records with correlation IDs.
- Only the configured verified Google email can query the hidden audit API/page; the route is absent from navigation.
- The worker purges records older than configured retention in bounded batches without recursive per-row audit entries.
- Exports contain only the requesting user's selected snapshot.
- Reset/delete removes exactly the confirmed scope and associated S3 objects with recoverability expectations documented.
- Homelab deployment passes migration, health, restart, backup, restore, and release checks.

## Verification Plan

- Unit tests for audit action mapping, redaction, export formatting, and deletion plans.
- PostgreSQL integration tests for audit authorization/immutability/retention, export isolation, and deletion scoping.
- E2E/security tests for hidden navigation, forbidden accounts, configured viewer, export, reset, and delete confirmation.
- Platform proof for production-like Compose, migrations, worker restart, S3, backup/restore, and rollback documentation.

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
