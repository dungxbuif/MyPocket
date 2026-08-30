---
artifact_type: ticket
id: TICKET-009
status: ready
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
  backlog_item: BL-003
  requirement:
    - REQ-F-004
    - REQ-NF-002
    - REQ-NF-003
    - REQ-NF-008
  phase: PHASE-003
  detail_design: ../phases/PHASE-003-detail-design.md
  implementation_plan: ../../superpowers/plans/2026-08-30-phase-003-offline-sync.md
  test_verification: ../test-verification/PHASE-003-offline-sync.md
  validation_matrix: ../VALIDATION_MATRIX.md
  docs_review: per-ticket_completion_checklist
  adrs: []
  release_notes: ../../releases/CHANGELOG.md
---

# Ticket: TICKET-009 Idempotent Sync API and Change Feed

## Status

- ID: TICKET-009
- Status: ready
- Type: feature
- Priority: high
- Phase: PHASE-003
- Owner: human

## Context

Human fill:

- User/business/system problem: Offline writes must sync once, in order, without losing data or overwriting server changes silently.
- Source prompt or requirement: REQ-F-004, REQ-NF-002, REQ-NF-003, REQ-NF-008, PHASE-003.
- Out of scope: frontend conflict inbox UI, provider ingestion, collaborative real-time sync, and silent last-write-wins.

AI fill:

- Current repository context read: `docs/CONTEXT.md`, `docs/work/BACKLOG.md`, standards, PHASE-003, PHASE-003 detail design, API/ERD architecture docs, and PHASE-002 finance implementation state.
- Brownfield touched scope, if applicable: backend migrations, finance repository/domain boundaries, sync package, HTTP routes, API docs, ERD docs, and backend tests.

## Acceptance Criteria

- [ ] Given ordered client mutations, when `POST /api/v1/sync/mutations` receives a valid batch, then it applies each mutation at most once and returns per-item `applied`, `replayed`, `rejected`, or `conflict` results.
- [ ] Given a duplicate `mutation_id` with the same request hash, when replayed, then the server returns the stored result without applying accounting effects again.
- [ ] Given a duplicate `mutation_id` with a different request hash, when submitted, then the server rejects it with a stable safe error.
- [ ] Given a stale base version, when the mutation would overwrite a newer server record, then the server preserves authoritative state and returns an explicit conflict result.
- [ ] Given `GET /api/v1/sync/changes?after={cursor}&limit={n}`, when the user is authenticated, then the server returns only that user's ordered changes, tombstones, and `next_cursor`.
- [ ] Given an invalid cursor or schema recovery request, when `POST /api/v1/sync/resync` is called, then the server returns an authoritative bounded snapshot without duplicating accounting.

## Small Task Exemption

- Small task exemption: no
- Reason: This ticket adds public sync APIs, database schema, idempotency semantics, and security-sensitive user isolation.
- Impact checked: API=yes, DB=yes, Security=yes, Runtime=yes, Standards=no

## Impacted Areas

- Code: backend migrations, new sync domain/service, finance command integration, HTTP routing, tests.
- Requirements docs: no requirement change expected.
- Architecture docs: update with sync module and runtime flow.
- API docs: add `/api/v1/sync/*` contracts and error codes.
- ERD/data docs: add `sync_changes`, `sync_mutations`, and related indexes/tombstone retention.
- Decisions: no ADR expected unless conflict/idempotency policy changes.

## Detail Design

- Required: yes
- Link: [PHASE-003-detail-design.md](../phases/PHASE-003-detail-design.md)
- Approval: approved 2026-08-30.

## Test Expectations

- Unit: sync envelope validation, request hashing, retry classification, result serialization.
- Integration: mutation replay, user isolation, cursor monotonicity, tombstones, base-version conflicts, resync snapshot.
- E2E: reconnect applies queued offline mutations once and refreshes authoritative versions.
- UAT: required with frontend flows after TICKET-010.
- Manual/platform: inspect migration and local PostgreSQL behavior through backend regression.
- Docs review: API, ERD, architecture, validation matrix, context, backlog, changelog.

## Verification Results

- Command: not run yet
- Result: pending
- Notes: Ticket is ready for implementation; no execution evidence is claimed.

## Fix/Test Attempt Log

- Same-path failure attempts: 0 / 3
- Total fix/test cycles: 0 / 5
- Blocked by loop guard: no
- Human/design input needed: none; design approval is recorded.

## UAT

- Required: yes
- Reason if not required: not applicable
- Expected behavior: reconnect syncs once, handles stale edits as conflicts, and never leaks another user's data.
- Verified behavior: pending implementation.
- Sign-off: pending.

## Docs Review

- Requirements updated or not needed reason: no requirement change expected.
- Architecture updated or not needed reason: pending implementation reconciliation.
- API updated or not needed reason: pending sync route implementation.
- ERD/data updated or not needed reason: pending sync migration implementation.
- ADR created or not needed reason: not expected; follows approved PHASE-003 detail design.
- `docs/CONTEXT.md` updated: yes, planning state recorded.

## Completion Checklist

- [ ] Implementation complete
- [ ] Tests run and recorded
- [ ] Fix/test loop guard respected
- [ ] Validation matrix updated or explicitly not affected
- [ ] UAT completed or explicitly not required
- [ ] Master docs reconciled
- [ ] Docs review completed
- [ ] ADR created or explicitly not needed
- [ ] `docs/CONTEXT.md` updated
- [ ] `docs/work/BACKLOG.md` updated
- [x] Trace links updated
