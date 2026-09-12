---
artifact_type: ticket
id: TICKET-009
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
- Status: in_review
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

- [x] Given ordered client mutations, when `POST /api/v1/sync/mutations` receives a valid batch, then it applies each mutation at most once and returns per-item `applied`, `replayed`, `rejected`, or `conflict` results.
- [x] Given a duplicate `mutation_id` with the same request hash, when replayed, then the server returns the stored result without applying accounting effects again.
- [x] Given a duplicate `mutation_id` with a different request hash, when submitted, then the server rejects it with a stable safe error.
- [x] Given a stale base version, when the mutation would overwrite a newer server record, then the server preserves authoritative state and returns an explicit conflict result.
- [x] Given `GET /api/v1/sync/changes?after={cursor}&limit={n}`, when the user is authenticated, then the server returns only that user's ordered changes, tombstones, and `next_cursor`.
- [x] Given an invalid cursor or schema recovery request, when `POST /api/v1/sync/resync` is called, then the server returns an authoritative bounded snapshot without duplicating accounting.

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

- Command: `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test -p 1 ./... -count=1` from `backend/`
- Result: passed 2026-08-31 after local-network rerun; initial sandboxed run was blocked from TCP `127.0.0.1:55433`.
- Command: `rtk npm run build` from `frontend/`
- Result: passed 2026-08-31.
- Command: `rtk npm test -- --run` from `frontend/`
- Result: passed 2026-08-31, 3 files / 21 tests.
- Command: `rtk npm run test:e2e -- offline-sync.spec.ts` from `frontend/`
- Result: passed 2026-08-31, 1 mobile Playwright test.
- Notes: Sync migration, idempotency ledger, user-scoped change feed, resync snapshot, stale-version conflict response, authenticated/CSRF-protected sync routes, frontend sync API drain, and reconnect-once mobile E2E have automated proof. Human PHASE-003 UAT remains pending after TICKET-010 conflict inbox.

## Fix/Test Attempt Log

- Same-path failure attempts: 0 / 3
- Total fix/test cycles: 1 / 5
- Blocked by loop guard: no
- Human/design input needed: none; design approval is recorded.

## UAT

- Required: yes
- Reason if not required: not applicable
- Expected behavior: reconnect syncs once, handles stale edits as conflicts, and never leaks another user's data.
- Verified behavior: backend and mobile browser proof covers ordered offline mutation replay, duplicate replay without duplicate accounting, stale update conflict responses, user-scoped change feeds, and authoritative resync. Human PHASE-003 UAT remains pending after TICKET-010 conflict inbox.
- Sign-off: pending.

## Docs Review

- Requirements updated or not needed reason: no requirement change expected.
- Architecture updated or not needed reason: updated sync module/runtime notes already describe the implemented API-backed IndexedDB/outbox boundary.
- API updated or not needed reason: updated `/api/v1/sync/mutations`, `/api/v1/sync/changes`, and `/api/v1/sync/resync` contracts.
- ERD/data updated or not needed reason: updated implemented `sync_cursors`, `sync_changes`, and `sync_mutations` schema notes.
- ADR created or not needed reason: not expected; follows approved PHASE-003 detail design.
- `docs/CONTEXT.md` updated: yes, planning state recorded.

## Completion Checklist

- [x] Implementation complete
- [x] Tests run and recorded
- [x] Fix/test loop guard respected
- [x] Validation matrix updated or explicitly not affected
- [ ] UAT completed or explicitly not required
- [x] Master docs reconciled
- [x] Docs review completed
- [x] ADR created or explicitly not needed
- [x] `docs/CONTEXT.md` updated
- [x] `docs/work/BACKLOG.md` updated
- [x] Trace links updated
