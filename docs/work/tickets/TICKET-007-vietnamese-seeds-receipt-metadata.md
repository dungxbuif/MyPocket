---
artifact_type: ticket
id: TICKET-007
status: in_review
owner: human
priority: high
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
  backlog_item: BL-002
  requirement:
    - REQ-F-002
    - REQ-F-011
  phase: PHASE-002
  detail_design: ../phases/PHASE-002-detail-design.md
  implementation_plan: ../../superpowers/plans/2026-08-30-phase-002-finance-core.md
  test_verification: ../test-verification/PHASE-002-finance-core.md
  validation_matrix: ../VALIDATION_MATRIX.md
  docs_review: per-ticket_completion_checklist
  adrs:
    - ADR-001
    - ADR-004
  release_notes: ../../releases/CHANGELOG.md
---

# Ticket: TICKET-007 Vietnamese Seed Data and Receipt Metadata

## Status

- ID: TICKET-007
- Status: in_review
- Type: feature
- Priority: high
- Phase: PHASE-002
- Owner: human

## Context

Human fill:

- User/business/system problem: Users need stable Vietnamese categories and receipt metadata foundations before OCR and AI ingestion can produce reviewable drafts.
- Source prompt or requirement: REQ-F-002, REQ-F-011, PHASE-002, and ADR-004.
- Out of scope: OCR provider calls, receipt upload UI, AI draft parsing, and confirmed draft flows.

AI fill:

- Current repository context read: PHASE-002, requirements, ERD, integrations, ADR-004, and existing S3 adapter.
- Brownfield touched scope, if applicable: finance seed migration, seed tests, receipt metadata table/service, and API docs.

## Acceptance Criteria

- [ ] Given a fresh database, when migrations run, then the approved Vietnamese category taxonomy is installed with stable IDs and locked system rows.
- [ ] Given repeated migrations, when seed data already exists, then category IDs and names remain stable and no duplicates are created.
- [ ] Given receipt metadata is attached to a transaction, when it is stored, then the record is user-scoped and references a private S3 object key without exposing credentials or raw file bytes.
- [ ] Given AI/OCR phases are not implemented, when receipt metadata exists, then it does not create confirmed accounting effects by itself.
- [ ] UAT requirement is not_required for seed and metadata foundation; later OCR UI requires UAT in PHASE-006.

## Small Task Exemption

- Small task exemption: no
- Reason: This ticket introduces durable seed data and receipt metadata schema used by later provider phases.
- Impact checked: API=yes, DB=yes, Security=yes, Runtime=no, Standards=no

## Impacted Areas

- Code: migrations, finance seed repository, receipt metadata repository, S3 key validation.
- Requirements docs: no change expected.
- Architecture docs: reconcile receipt metadata boundary if implementation differs.
- API docs: add receipt metadata attachment contract only if exposed in PHASE-002.
- ERD/data docs: add implemented seed and receipt metadata tables.
- Decisions: no new ADR expected; follows ADR-004.

## Detail Design

- Required: yes
- Link: [PHASE-002-detail-design.md](../phases/PHASE-002-detail-design.md)
- Approval: approved by user instruction to complete the app; design remains available for human review.

## Test Expectations

- Unit: seed constants are stable and receipt object keys validate.
- Integration: migrations install seeds idempotently and receipt metadata remains user-scoped.
- E2E: not required unless receipt metadata is surfaced in PHASE-002 UI.
- UAT: not_required because user-facing receipt capture belongs to PHASE-006.
- Manual/platform: not required beyond existing S3 smoke.
- Docs review: ERD, API if exposed, validation matrix, context, backlog, changelog.

## Verification Results

- Command: `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test -p 1 ./internal/finance -count=1`
- Result: pass
- Notes: Receipt metadata validation, persistence, and user-scoped reads are covered by PostgreSQL-backed repository tests; seeded categories remain covered by the existing migration test.

## Fix/Test Attempt Log

- Same-path failure attempts: 0 / 3
- Total fix/test cycles: 0 / 5
- Blocked by loop guard: no
- Human/design input needed: none before starting approved PHASE-002 plan.

## UAT

- Required: no
- Reason if not required: seed and metadata foundation is not directly user-facing in this phase.
- Expected behavior: not applicable.
- Verified behavior: receipt object keys, content metadata, positive size, SHA-256 format, persistence, and cross-user access isolation are verified; metadata storage does not create accounting effects.
- Sign-off: not required.

## Docs Review

- Requirements updated or not needed reason: pending execution.
- Architecture updated or not needed reason: pending execution.
- API updated or not needed reason: pending execution.
- ERD/data updated or not needed reason: pending execution.
- ADR created or not needed reason: pending execution.
- `docs/CONTEXT.md` updated: pending execution.

## Completion Checklist

- [x] Implementation complete
- [x] Tests run and recorded
- [ ] Fix/test loop guard respected
- [ ] Validation matrix updated or explicitly not affected
- [ ] UAT completed or explicitly not required
- [ ] Master docs reconciled
- [ ] Docs review completed
- [ ] ADR created or explicitly not needed
- [x] `docs/CONTEXT.md` updated
- [x] `docs/work/BACKLOG.md` updated
- [ ] Trace links updated
