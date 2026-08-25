---
artifact_type: phase
id: PHASE-001
status: in_progress
owner: human
priority: Urgent
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
    - BL-001
  roadmap: ../ROADMAP.md
  requirements:
    - REQ-F-001
    - REQ-NF-001
    - REQ-NF-004
    - REQ-NF-006
    - REQ-NF-008
  tickets:
    - TICKET-001
    - TICKET-002
    - TICKET-003
    - TICKET-004
  bugs: []
  detail_design: PHASE-001-detail-design.md
  implementation_plan: ../../superpowers/plans/2026-08-25-phase-001-platform-identity.md
  test_verification: not_created_phase_not_executed
  validation_matrix: ../VALIDATION_MATRIX.md
  adrs: []
  release_notes: ../../releases/CHANGELOG.md
---

# PHASE-001: Platform and Identity

## Status

- ID: PHASE-001
- Status: in_progress
- Owner: human
- Priority: Urgent
- Created: 2026-08-24
- Updated: 2026-08-24

## Trace Links

- Backlog: [BACKLOG.md](../BACKLOG.md)
- Roadmap: [ROADMAP.md](../ROADMAP.md)
- Requirements: [REQUIREMENTS.md](../../requirements/REQUIREMENTS.md) — REQ-F-001, REQ-NF-001, REQ-NF-004, REQ-NF-006, REQ-NF-008
- Detail design: [PHASE-001-detail-design.md](PHASE-001-detail-design.md)
- Implementation plan: [2026-08-25-phase-001-platform-identity.md](../../superpowers/plans/2026-08-25-phase-001-platform-identity.md)
- Test verification: created during execution
- Validation matrix: [VALIDATION_MATRIX.md](../VALIDATION_MATRIX.md)
- ADRs: [decisions](../../decisions/README.md)
- Release notes: [CHANGELOG.md](../../releases/CHANGELOG.md)

## Goal

Deliver a deployable React PWA, Go API, Go worker, PostgreSQL and S3 foundation with secure Google authentication and strict user isolation.

## Scope

- Repository and build foundation for `apps/web`, `apps/api`, `apps/worker`, shared Go packages, migrations, and deployment files.
- React PWA shell, API client baseline, Go HTTP/config/logging baseline, PostgreSQL migration runner, and S3 adapter.
- Google OAuth callback, user provisioning, signed stateless cookie, CSRF protection, current-user endpoint, and ownership test harness.
- Docker Compose development topology, liveness/readiness endpoints, structured correlation IDs, and CI-quality commands.

## Out Of Scope

- Wallet/category/transaction behavior.
- Offline synchronization.
- Production credentials or the final homelab connection string.

## Tickets And Bugs

| ID | Type | Title | Status | Link |
| --- | --- | --- | --- | --- |
| TICKET-001 | Ticket | Repository and runtime foundation | draft | [TICKET-001](../tickets/TICKET-001-repository-runtime-foundation.md) |
| TICKET-002 | Ticket | PostgreSQL migrations and S3 platform adapters | draft | [TICKET-002](../tickets/TICKET-002-postgresql-migrations-s3-platform-adapters.md) |
| TICKET-003 | Ticket | Google OAuth and user isolation | draft | [TICKET-003](../tickets/TICKET-003-google-oauth-user-isolation.md) |
| TICKET-004 | Ticket | Development operations and verification baseline | draft | [TICKET-004](../tickets/TICKET-004-development-operations-verification-baseline.md) |

## Dependencies

- Approved SDD and ADR-001/ADR-002.
- Google OAuth test credentials or a deterministic callback fixture for automated tests.

## Risks

- Incorrect cookie/CSRF behavior can expose authenticated mutations.
- Missing ownership scoping can leak multi-user data.
- Runtime choices here constrain every later phase.

## Success Criteria

- Web, API, worker, PostgreSQL, and S3-compatible test service start through documented commands.
- Google callback fixture provisions a user and returns a secure application cookie without a session row or persisted provider token.
- Cross-user integration tests deny access to a second user's objects.
- Migration, health, structured logging, and baseline frontend/backend test commands pass.

## Verification Plan

- Go unit tests for config, cookie claims, and authorization helpers.
- PostgreSQL integration tests for migrations, user provisioning, and ownership isolation.
- React component/E2E checks for login callback, authenticated shell, logout, and forbidden states.
- Platform smoke test for Compose startup, health endpoints, and object-store access.

## Gate Checklist

- [x] Phase links approved requirements
- [x] Tickets have stable planned IDs and bounded titles
- [x] Risks and dependencies are recorded
- [x] Verification plan is defined
- [x] Release/changelog need is linked
- [x] Ticket artifacts and detailed implementation plan are created
- [x] Phase status is promoted to ready after plan review
- [x] Phase execution started after user approval

## Completion Summary

No implementation has started. Completion evidence will be recorded after the phase reaches execution.
