---
artifact_type: ticket
id: TICKET-010
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
    - REQ-NF-003
    - REQ-F-016
  phase: PHASE-003
  detail_design: ../phases/PHASE-003-detail-design.md
  implementation_plan: ../../superpowers/plans/2026-08-30-phase-003-offline-sync.md
  test_verification: ../test-verification/PHASE-003-offline-sync.md
  validation_matrix: ../VALIDATION_MATRIX.md
  docs_review: per-ticket_completion_checklist
  adrs: []
  release_notes: ../../releases/CHANGELOG.md
---

# Ticket: TICKET-010 Conflict Inbox and Recovery Flows

## Status

- ID: TICKET-010
- Status: in_review
- Type: feature
- Priority: high
- Phase: PHASE-003
- Owner: human

## Context

Human fill:

- User/business/system problem: When offline changes conflict with newer server state, the user must review and resolve the issue explicitly.
- Source prompt or requirement: REQ-F-004, REQ-NF-003, REQ-F-016, PHASE-003.
- Out of scope: automatic multi-field merges, collaborative live editing, and provider-generated draft conflict resolution.

AI fill:

- Current repository context read: `docs/CONTEXT.md`, `docs/work/BACKLOG.md`, standards, PHASE-003, PHASE-003 detail design, validation matrix, and frontend mobile patterns.
- Brownfield touched scope, if applicable: frontend conflict inbox UI, offline state, sync result handling, E2E specs, and docs reconciliation.

## Acceptance Criteria

- [x] Given a sync conflict result, when the frontend receives it, then the app stores server state, local intent, entity identity, operation, and recovery status in the local conflict store.
- [x] Given a conflict exists, when the user opens the mobile app, then an inbox entry is visible without blocking unrelated data viewing.
- [x] Given the user chooses keep-server, when confirmed, then the local mutation is removed and the authoritative server record is mirrored locally.
- [x] Given the user chooses edit-and-retry, when they submit revised values, then a new mutation is created against the current server version.
- [x] Given the user chooses discard-local, when confirmed, then the pending local intent is removed without applying it to the server.
- [x] Given IndexedDB is cleared, cursor is invalid, or local data is stale beyond recovery, when the user triggers full resync, then the app restores authoritative records and preserves unsent recoverable mutations.

## Small Task Exemption

- Small task exemption: no
- Reason: This ticket changes conflict UX, recovery behavior, and data-loss prevention paths.
- Impact checked: API=no, DB=no, Security=yes, Runtime=yes, Standards=no

## Impacted Areas

- Code: frontend conflict state, mobile inbox UI, sync result handling, E2E tests.
- Requirements docs: no requirement change expected.
- Architecture docs: reconcile conflict and recovery flow.
- API docs: reference sync result payloads implemented by TICKET-009.
- ERD/data docs: not directly affected beyond TICKET-009 server data.
- Decisions: no ADR expected unless approved conflict policy changes.

## Detail Design

- Required: yes
- Link: [PHASE-003-detail-design.md](../phases/PHASE-003-detail-design.md)
- Approval: approved 2026-08-30.

## Test Expectations

- Unit: conflict reducer actions, keep-server/edit-retry/discard-local, full-resync state transitions.
- Component: conflict inbox rows, resolution buttons, stale/offline indicators, degraded recovery action.
- E2E: two browser contexts create a stale edit conflict, then resolve it; cleared storage/full resync flow.
- UAT: required for conflict copy and mobile ergonomics.
- Manual/platform: optional storage inspection for recovery edge cases.
- Docs review: validation matrix, context, backlog, changelog, API references if payload docs change.

## Verification Results

- Command: `rtk npm test -- --run` from `frontend/`
- Result: passed 2026-08-31, 4 files / 28 tests
- Command: `rtk npm run build` from `frontend/`
- Result: passed 2026-08-31
- Command: `rtk npm run test:e2e -- offline-sync.spec.ts` from `frontend/`
- Result: passed 2026-08-31, 2 mobile tests
- Command: `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test -p 1 ./... -count=1` from `backend/`
- Result: passed 2026-08-31
- Notes: Conflict results are stored in IndexedDB with local/server payloads, the mobile inbox remains non-blocking, keep-server/discard-local remove the pending mutation, edit-and-retry queues a new mutation against the server version, and full resync restores authoritative mirror data while preserving recoverable outbox mutations.

## Fix/Test Attempt Log

- Same-path failure attempts: 0 / 3
- Total fix/test cycles: 0 / 5
- Blocked by loop guard: no
- Human/design input needed: none; design approval is recorded.

## UAT

- Required: yes
- Reason if not required: not applicable
- Expected behavior: conflicts are explicit, readable on mobile, and resolvable without silent overwrite.
- Verified behavior: automated proof covers mobile conflict visibility, keep-server recovery, conflict storage, edit-and-retry mutation replacement, discard-local mutation removal, and full resync mirror restoration with outbox preservation.
- Sign-off: pending.

## Docs Review

- Requirements updated or not needed reason: no requirement change expected.
- Architecture updated or not needed reason: updated to record conflict persistence and explicit recovery after sync conflict results.
- API updated or not needed reason: updated sync mutation result payload notes for conflict entries consumed by the frontend inbox.
- ERD/data updated or not needed reason: not directly affected beyond TICKET-009.
- ADR created or not needed reason: not expected; follows approved PHASE-003 detail design.
- `docs/CONTEXT.md` updated: yes, implementation and next queue state recorded.

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
