---
artifact_type: ticket
id: TICKET-002
status: verified
owner: human
priority: urgent
lane: high-risk
human_fields:
  - title
  - priority
  - acceptance_criteria
  - scope
  - approval
ai_fields:
  - impacted_areas
  - test_expectations
  - verification_results
  - docs_review
  - context_updates
shared_fields:
  - status
  - trace
  - small_task_exemption
trace:
  backlog_item: BL-001
  requirement:
    - REQ-NF-006
  phase: PHASE-001
  detail_design: ../phases/PHASE-001-detail-design.md
  implementation_plan: ../../superpowers/plans/2026-08-25-phase-001-platform-identity.md
  test_verification: ../test-verification/PHASE-001-platform-smoke.md
  validation_matrix: ../VALIDATION_MATRIX.md
  docs_review: per-ticket_completion_checklist
  adrs:
    - ADR-001
  release_notes: ../../releases/CHANGELOG.md
---

# Ticket: TICKET-002 PostgreSQL Migrations and S3 Platform Adapters

## Field Ownership

- Human fills intent, priority, acceptance criteria, scope, and approval.
- AI fills impact analysis, test expectations, verification evidence, docs review, and context/backlog updates.
- Shared fields include status, trace links, and small-task exemption.

## Status

- ID: TICKET-002
- Status: verified
- Type: feature
- Priority: urgent
- Phase: PHASE-001
- Owner: human

## Trace Links

- Backlog item: [BL-001](../BACKLOG.md)
- Requirement: [REQ-NF-006](../../requirements/REQUIREMENTS.md)
- Phase: [PHASE-001](../phases/PHASE-001-platform-identity.md)
- Detail design: [PHASE-001-detail-design.md](../phases/PHASE-001-detail-design.md)
- Implementation plan: [2026-08-25-phase-001-platform-identity.md](../../superpowers/plans/2026-08-25-phase-001-platform-identity.md)
- Test verification: created during execution
- Validation matrix: [VALIDATION_MATRIX.md](../VALIDATION_MATRIX.md)
- Docs review: this ticket completion checklist
- ADRs: [ADR-001](../../decisions/ADR-001-react-go-modular-monolith.md)
- Release notes: [CHANGELOG.md](../../releases/CHANGELOG.md)

## Context

Human fill:

- User/business/system problem: Later finance, sync, receipt, export, and worker phases need a tested PostgreSQL and S3-compatible foundation.
- Source prompt or requirement: PHASE-001 platform scope and REQ-NF-006.
- Out of scope: finance tables, sync tables, receipt upload UI, production backup/restore proof, and account deletion jobs.

AI fill:

- Current repository context read: required hydration docs plus architecture, ERD, SDD, ADR-001.
- Brownfield touched scope, if applicable: no existing implementation; expected touched scope is new migration runner, initial identity/platform migrations, and S3 adapter.

## Acceptance Criteria

- [x] Given an empty PostgreSQL database, when migrations run, then PHASE-001 tables for users and platform metadata are created idempotently through documented commands.
- [x] Given API and worker processes start, when they open database connections, then readiness reflects required database availability.
- [x] Given configured S3-compatible credentials, when the adapter performs a smoke operation, then object-store access is verified without exposing credentials in logs.
- [x] Given missing or invalid required runtime configuration, when API or worker starts, then startup fails with safe, actionable configuration errors.
- [x] UAT requirement is not_required because this ticket has no direct user-facing behavior.

## Small Task Exemption

- Small task exemption: no
- Reason: This introduces database migrations, object-store integration, and runtime configuration.
- Impact checked: API=no, DB=yes, Security=yes, Runtime=yes, Standards=no

## Impacted Areas

- Code: migration files, migration runner, database config, S3 adapter interface/implementation, health readiness hooks.
- Requirements docs: no change expected.
- Architecture docs: no change expected if implementation follows approved architecture.
- API docs: health readiness details may be reconciled.
- ERD/data docs: update only if concrete PHASE-001 table definitions differ from master ERD.
- Decisions: no new ADR expected; follows ADR-001.

## Detail Design

- Required: yes
- Link: [PHASE-001-detail-design.md](../phases/PHASE-001-detail-design.md)
- Approval: approved by user instruction "finish whole app" on 2026-08-25

## Test Expectations

- Unit: config validation and object-key/safe-error helpers.
- Integration: apply migrations to empty database; verify users table shape; exercise S3 adapter against local S3-compatible service.
- E2E: not required.
- UAT: not_required because behavior is platform-only.
- Manual/platform: compose startup proves PostgreSQL and S3-compatible service readiness.
- Docs review: ticket checklist plus any ERD/API reconciliation.

## Verification Results

- Command: `rtk go test ./internal/platform/db`
- Result: fail
- Notes: RED check failed for expected missing package/implementation path after tests were added: `no non-test Go files`.
- Command: `rtk go get github.com/jackc/pgx/v5/stdlib@v5.7.2`
- Result: pass
- Notes: Added PostgreSQL driver dependency required by the approved migration/runtime scope.
- Command: `rtk env 'MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55432/mypocket?sslmode=disable' go test ./internal/platform/db`
- Result: pass
- Notes: Migration integration tests passed against a disposable local PostgreSQL 16 container; verified idempotent application, checksum mismatch rejection, and `users` schema without provider-token/session columns.
- Command: `rtk env 'DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55432/mypocket?sslmode=disable' go run ./apps/migrate`
- Result: pass
- Notes: Documented migration command printed `migrations applied` against the disposable local PostgreSQL 16 container.
- Command: `rtk go test ./internal/platform/db ./internal/platform/httpapi`
- Result: pass
- Notes: 5 tests passed across database migration and API health packages.
- Command: `rtk go test ./...`
- Result: pass
- Notes: 8 tests passed across 5 Go packages for the current backend/platform baseline.
- Command: `rtk git diff --check`
- Result: pass
- Notes: No whitespace errors reported.
- Command: `rtk go test ./internal/platform/config ./internal/platform/objectstore`
- Result: fail
- Notes: RED check failed for intended missing behavior: incomplete S3 config accepted and `internal/platform/objectstore` had no non-test Go files.
- Command: `rtk go get github.com/aws/aws-sdk-go-v2/config@v1.29.7 github.com/aws/aws-sdk-go-v2/credentials@v1.17.60 github.com/aws/aws-sdk-go-v2/service/s3@v1.76.0`
- Result: pass
- Notes: Added AWS SDK v2 S3 client dependencies for the approved S3-compatible adapter.
- Command: `rtk go test ./internal/platform/config ./internal/platform/objectstore`
- Result: pass
- Notes: 5 tests passed for config validation and objectstore default path; S3 lifecycle smoke skipped without env.
- Command: `rtk docker run --rm -d --name mypocket-test-s3 -e SERVICES=s3 -e DEFAULT_REGION=us-east-1 -p 54566:4566 localstack/localstack:3.8.1`
- Result: pass
- Notes: Started cached LocalStack 3.8.1 container for S3-compatible smoke testing.
- Command: `rtk curl -fsS http://127.0.0.1:54566/_localstack/health`
- Result: pass
- Notes: LocalStack reported S3 service `available`.
- Command: `rtk env MYPOCKET_TEST_S3_ENDPOINT=http://127.0.0.1:54566 MYPOCKET_TEST_S3_BUCKET=mypocket-smoke MYPOCKET_TEST_S3_ACCESS_KEY=test MYPOCKET_TEST_S3_SECRET_KEY=test go test ./internal/platform/objectstore`
- Result: pass
- Notes: S3 smoke lifecycle test created the bucket, put the smoke object, and deleted it against LocalStack.
- Command: `rtk go test ./...`
- Result: pass
- Notes: 10 tests passed across 7 Go packages after S3 adapter implementation.
- Command: `rtk go test ./...` from `backend/`
- Result: pass
- Notes: 17 tests passed across 8 Go packages after backend/frontend split and auth browser proof updates.

## Fix/Test Attempt Log

- Same-path failure attempts: 0 / 3
- Total fix/test cycles: 0 / 5
- Blocked by loop guard: no
- Human/design input needed: none for approved PHASE-001 scope.

## UAT

- Required: no
- Reason if not required: platform-only adapter and migration behavior is covered by integration and platform proof.
- Expected behavior: not applicable
- Verified behavior: PostgreSQL migration/readiness and S3-compatible smoke behavior verified by automated integration/unit tests.
- Sign-off: not required

## Docs Review

- Requirements updated or not needed reason: not needed; implementation follows existing PHASE-001/REQ-NF-006 platform scope.
- Architecture updated or not needed reason: updated to record implemented PostgreSQL startup and S3-compatible adapter baseline.
- API updated or not needed reason: updated to record concrete `/api/v1/health/live` and `/api/v1/health/ready` readiness behavior.
- ERD/data updated or not needed reason: updated to record concrete PHASE-001 `users` and `schema_migrations` columns.
- ADR created or not needed reason: not needed; follows ADR-001 and the approved detail design.
- `docs/CONTEXT.md` updated: yes; next steps now point to Google OAuth/user isolation.

## Completion Checklist

- [x] Implementation complete
- [x] Tests run and recorded
- [x] Fix/test loop guard respected
- [x] Validation matrix updated or explicitly not affected
- [x] UAT completed or explicitly not required
- [x] Master docs reconciled
- [x] Docs review completed
- [x] ADR created or explicitly not needed
- [x] `docs/CONTEXT.md` updated
- [x] `docs/work/BACKLOG.md` updated
- [x] Trace links updated
