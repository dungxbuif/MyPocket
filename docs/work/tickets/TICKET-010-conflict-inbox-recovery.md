---
artifact_type: ticket
id: TICKET-010
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
- Status: ready
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

- [ ] Given a sync conflict result, when the frontend receives it, then the app stores server state, local intent, entity identity, operation, and recovery status in the local conflict store.
- [ ] Given a conflict exists, when the user opens the mobile app, then an inbox entry is visible without blocking unrelated data viewing.
- [ ] Given the user chooses keep-server, when confirmed, then the local mutation is removed and the authoritative server record is mirrored locally.
- [ ] Given the user chooses edit-and-retry, when they submit revised values, then a new mutation is created against the current server version.
- [ ] Given the user chooses discard-local, when confirmed, then the pending local intent is removed without applying it to the server.
- [ ] Given IndexedDB is cleared, cursor is invalid, or local data is stale beyond recovery, when the user triggers full resync, then the app restores authoritative records and preserves unsent recoverable mutations.

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
- Expected behavior: conflicts are explicit, readable on mobile, and resolvable without silent overwrite.
- Verified behavior: pending implementation.
- Sign-off: pending.

## Docs Review

- Requirements updated or not needed reason: no requirement change expected.
- Architecture updated or not needed reason: pending implementation reconciliation.
- API updated or not needed reason: pending final sync result payload reconciliation.
- ERD/data updated or not needed reason: not directly affected beyond TICKET-009.
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
