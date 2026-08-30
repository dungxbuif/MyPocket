---
artifact_type: ticket
id: TICKET-025
status: ready
owner: human
priority: high
lane: high-risk
trace:
  backlog_item: BL-007
  requirement: [REQ-NF-006, REQ-NF-007]
  phase: PHASE-007
  detail_design: ../phases/PHASE-007-detail-design.md
  implementation_plan: ../../superpowers/plans/2026-08-30-phase-007-production.md
  test_verification: ../test-verification/PHASE-007-audit-export-production.md
  validation_matrix: ../VALIDATION_MATRIX.md
  release_notes: ../../releases/CHANGELOG.md
---

# Ticket: TICKET-025 Homelab Production Hardening and Release Proof

## Status

- Status: ready
- Type: feature
- Priority: high
- Phase: PHASE-007

## Context

The release needs production-like configuration validation, migration/health proof, backup/restore, restart, and rollback guidance.

## Acceptance Criteria

- [ ] Startup validates public URL, database, cookie security, Google OAuth, S3, providers, VAPID, audit viewer, and retention config.
- [ ] Health separates liveness and readiness for PostgreSQL/migrations without blocking on external providers.
- [ ] Release docs define backup, migrate, start, readiness, smoke, expose traffic, restart, and rollback steps.
- [ ] Backup/restore proof verifies PostgreSQL rows and S3 object inventory in an isolated restore target.
- [ ] Worker restart preserves leases and does not duplicate jobs.

## Small Task Exemption

- Small task exemption: no
- Reason: Changes deployment/runtime configuration and production release gates.
- Impact checked: API=no, DB=yes, Security=yes, Runtime=yes, Standards=no

## Test Expectations

- Unit env validation and worker lease tests.
- Compose smoke, migration, restart, backup, restore, and rollback documentation checks.
- Release proof recorded in PHASE-007 verification artifact.

## Verification Results

- Command: not run yet
- Result: pending
- Notes: Ready for implementation; no execution evidence claimed.
