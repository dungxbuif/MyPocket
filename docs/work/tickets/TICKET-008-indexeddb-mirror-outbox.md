---
artifact_type: ticket
id: TICKET-008
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

# Ticket: TICKET-008 IndexedDB Mirror and Offline Outbox

## Status

- ID: TICKET-008
- Status: in_review
- Type: feature
- Priority: high
- Phase: PHASE-003
- Owner: human

## Context

Human fill:

- User/business/system problem: The PWA must remain useful on mobile when the network is unavailable, including after reload.
- Source prompt or requirement: REQ-F-004, REQ-F-016, PHASE-003.
- Out of scope: backend sync API, server conflict resolution, provider ingestion, and browser background sync guarantees after the app is terminated.

AI fill:

- Current repository context read: `docs/CONTEXT.md`, `docs/work/BACKLOG.md`, standards, PHASE-003, PHASE-003 detail design, validation matrix, and PHASE-002 finance verification.
- Brownfield touched scope, if applicable: `frontend/src/app`, new `frontend/src/offline`, frontend tests, and mobile/PWA behavior.

## Acceptance Criteria

- [x] Given an authenticated user has finance data, when the PWA hydrates, then wallets, categories, transactions, sync meta, conflicts, tombstones, and outbox records are loaded from IndexedDB before network refresh.
- [x] Given the browser is offline, when the user creates or edits supported finance data, then the UI updates optimistically and appends a durable ordered mutation with `mutation_id`, `device_id`, `sequence`, entity identity, operation, base version, and payload.
- [x] Given the app reloads offline, when the user opens finance screens, then cached data and pending mutations remain visible without requiring the API.
- [x] Given legacy temporary localStorage outbox entries exist, when IndexedDB initializes, then valid entries migrate once and invalid entries are quarantined with a recovery note.
- [x] Given IndexedDB quota or open failure, when the app cannot persist writes safely, then it enters read-only degraded mode with a recovery action instead of accepting unsafe offline writes.

## Small Task Exemption

- Small task exemption: no
- Reason: This ticket changes user-facing offline behavior, private local storage, and PHASE-003 data contracts.
- Impact checked: API=no, DB=no, Security=yes, Runtime=yes, Standards=no

## Impacted Areas

- Code: frontend offline adapter, finance client state, PWA service integration, component tests, E2E tests.
- Requirements docs: no contract change expected.
- Architecture docs: update when implementation finalizes the offline module boundary.
- API docs: not affected in this ticket.
- ERD/data docs: browser storage design may be referenced during reconciliation.
- Decisions: no ADR expected unless the approved IndexedDB strategy changes.

## Detail Design

- Required: yes
- Link: [PHASE-003-detail-design.md](../phases/PHASE-003-detail-design.md)
- Approval: approved 2026-08-30.

## Test Expectations

- Unit: IndexedDB schema migration, localStorage migration, ordering, retry classification, optimistic reducers, degraded mode.
- Integration/component: finance screens hydrate from local cache, queue offline mutations, preserve pending state across reload simulation.
- E2E: offline create/edit and offline reload before reconnect.
- UAT: required for mobile offline UX clarity.
- Manual/platform: browser storage inspection if automated proof cannot assert quota/degraded state.
- Docs review: validation matrix, context, backlog, changelog, architecture notes.

## Verification Results

- Command: `rtk npm run build` from `frontend/`
- Result: passed 2026-08-31
- Command: `rtk npm test -- --run` from `frontend/`
- Result: passed 2026-08-31, 3 files / 20 tests.
- Command: `rtk npm run test:e2e -- pwa-shell.spec.ts` from `frontend/`
- Result: passed 2026-08-31 after rerun with local network access; initial sandboxed run was blocked from TCP `127.0.0.1:55433`.
- Notes: IndexedDB mirror/outbox, localStorage migration/quarantine, durable sequence, offline transaction UI queueing, wallet/category queueing, cached-auth offline reload with pending mutation visibility, degraded read-only UI, and PWA service-worker offline reload have automated proof. Human mobile UAT remains pending before `verified`.

## Fix/Test Attempt Log

- Same-path failure attempts: 0 / 3
- Total fix/test cycles: 2 / 5
- Blocked by loop guard: no
- Human/design input needed: none; design approval is recorded.

## UAT

- Required: yes
- Reason if not required: not applicable
- Expected behavior: mobile PWA finance screens stay usable offline and make pending write state obvious.
- Verified behavior: cached finance data hydrates offline, offline transaction create queues durably and updates pending status, pending mutations remain visible after offline reload through cached auth, IndexedDB failure disables unsafe offline writes, and PWA shell reloads offline. Human mobile UAT remains pending.
- Sign-off: pending.

## Docs Review

- Requirements updated or not needed reason: no requirement change expected.
- Architecture updated or not needed reason: updated browser offline module boundary after implementation.
- API updated or not needed reason: not affected by this frontend-local ticket.
- ERD/data updated or not needed reason: PostgreSQL ERD not affected; IndexedDB store shape remains documented in this ticket/design.
- ADR created or not needed reason: not expected; follows approved PHASE-003 detail design.
- `docs/CONTEXT.md` updated: yes, implementation progress recorded.

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
