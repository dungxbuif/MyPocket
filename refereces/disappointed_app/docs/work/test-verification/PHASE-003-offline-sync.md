---
artifact_type: test_verification
id: PHASE-003-offline-sync
status: in_review
owner: shared
human_fields:
  - uat_sign_off
  - manual_acceptance_notes
ai_fields:
  - commands
  - automated_tests
  - manual_checks
  - failure_summary
  - evidence_notes
  - attempt_log
shared_fields:
  - status
  - trace
  - uat
trace:
  backlog_item: BL-003
  requirement:
    - REQ-F-004
    - REQ-NF-002
    - REQ-NF-003
    - REQ-NF-008
    - REQ-F-016
  phase: PHASE-003
  ticket_or_bug:
    - TICKET-008
    - TICKET-009
    - TICKET-010
  detail_design: ../phases/PHASE-003-detail-design.md
  validation_matrix: ../VALIDATION_MATRIX.md
  docs_review: per-ticket_completion_checklist
  release_notes: ../../releases/CHANGELOG.md
---

# PHASE-003 Offline Synchronization Verification

## Status

- ID: PHASE-003-offline-sync
- Status: in_review
- Owner: shared

## Scope

- Ticket/Bug/Phase: PHASE-003, TICKET-008, TICKET-009, TICKET-010
- Tested behavior: IndexedDB mirror, offline outbox, idempotent sync API, change feed, tombstones, conflict inbox, and recovery flows.

## Trace Links

- Backlog item: [BL-003](../BACKLOG.md)
- Requirement: [REQ-F-004, REQ-NF-002, REQ-NF-003, REQ-NF-008, REQ-F-016](../../requirements/REQUIREMENTS.md)
- Phase: [PHASE-003](../phases/PHASE-003-offline-sync.md)
- Tickets: [TICKET-008](../tickets/TICKET-008-indexeddb-mirror-outbox.md), [TICKET-009](../tickets/TICKET-009-sync-api-change-feed.md), [TICKET-010](../tickets/TICKET-010-conflict-inbox-recovery.md)
- Detail design: [PHASE-003-detail-design.md](../phases/PHASE-003-detail-design.md)
- Validation matrix: [VALIDATION_MATRIX.md](../VALIDATION_MATRIX.md)
- Release notes: [CHANGELOG.md](../../releases/CHANGELOG.md)

## Planned Commands

| Command | Expected Coverage |
| --- | --- |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/sync -count=1` | Sync envelope validation, idempotency, cursor, tombstone, and conflict service behavior. |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test -p 1 ./... -count=1` | Full backend regression with sync routes and migrations. |
| `rtk npm test -- --run` | IndexedDB adapter, optimistic reducers, outbox processor, conflict reducer, and mobile component behavior. |
| `rtk npm run build` | Production PWA build after offline sync implementation. |
| `rtk npm run test:e2e -- offline-sync.spec.ts` | Mobile offline transaction create, reconnect-once replay, and reload after authoritative sync. |
| `rtk npm run test:e2e` | Full frontend browser regression. |

## Verification Results

| Command | Result | Coverage |
| --- | --- | --- |
| `rtk npm run build` from `frontend/` | Passed 2026-08-31 | Production PWA build after IndexedDB/offline UI changes. |
| `rtk npm test -- --run` from `frontend/` | Passed 2026-08-31, 3 files / 21 tests | IndexedDB mirror, legacy localStorage migration/quarantine, durable sequence, sync API outbox drain, transaction outbox fallback, wallet/category mutation queueing, offline cached hydration, cached-auth offline reload with pending mutation visibility, offline transaction create queueing, and degraded read-only UI. |
| `rtk npm run test:e2e -- pwa-shell.spec.ts` from `frontend/` | Passed 2026-08-31 after local-network rerun | Mobile service-worker registration and app-shell reload while offline. |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test -p 1 ./... -count=1` from `backend/` | Passed 2026-08-31 after local-network rerun | Full backend regression with sync migration, sync service integration, authenticated sync HTTP routes, user isolation, stale conflict handling, cursor feed, resync snapshot, and existing finance/identity/platform behavior. |
| `rtk npm run build` from `frontend/` | Passed 2026-08-31 after sync API wiring | Production PWA build after frontend sync API drain changes. |
| `rtk npm test -- --run` from `frontend/` | Passed 2026-08-31, 3 files / 21 tests after sync API wiring | IndexedDB/outbox regression including sync API result application to the local mirror. |
| `rtk npm run test:e2e -- offline-sync.spec.ts` from `frontend/` | Passed 2026-08-31, 1 mobile test | Offline transaction is queued, displayed locally, replayed once through sync API after reconnect, and still visible after reload. |
| `rtk npm test -- --run` from `frontend/` | Passed 2026-08-31, 4 files / 28 tests after conflict recovery wiring | IndexedDB conflict storage, outbox conflict preservation, keep-server, discard-local, edit-and-retry, full resync preservation, and mobile conflict inbox component behavior. |
| `rtk npm run build` from `frontend/` | Passed 2026-08-31 after conflict recovery wiring | Production PWA build after conflict inbox and recovery UI changes. |
| `rtk npm run test:e2e -- offline-sync.spec.ts` from `frontend/` | Passed 2026-08-31, 2 mobile tests | Offline transaction reconnect replay plus stale offline edit conflict creation and keep-server resolution in the mobile UI. |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test -p 1 ./... -count=1` from `backend/` | Passed 2026-08-31 after conflict recovery frontend wiring | Full backend regression after TICKET-010, including existing sync conflict API behavior. |

Notes: First `pwa-shell.spec.ts` attempt failed because the sandbox denied TCP to local PostgreSQL at `127.0.0.1:55433`; rerun with local network permission passed. TICKET-009 backend sync API/change feed and reconnect replay E2E have automated proof. TICKET-010 conflict inbox and recovery flows now have automated unit/component/E2E proof. Final human UAT remains pending before PHASE-003 can be verified.

## Fix/Test Attempt Log

| Attempt | Change Made | Command | Result | Failure Summary |
| --- | --- | --- | --- | --- |
| 1 | Added IndexedDB mirror/outbox and wired finance UI offline hydration/queue paths. | `rtk npm run build`; `rtk npm test -- --run` | failed then passed after fixes | Fixed async default parameter syntax, add-sheet async wallet initialization, and defensive manager callbacks. |
| 2 | Tightened migration sequence/quarantine tests, added degraded-mode UI proof, and added cached-auth offline reload proof. | `rtk npm run build`; `rtk npm test -- --run`; `rtk npm run test:e2e -- pwa-shell.spec.ts` | passed | Initial E2E attempt was sandbox-blocked; rerun with local network access passed. |
| 3 | Added sync migration/service/routes plus frontend sync API outbox drain and mobile reconnect replay E2E. | `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test -p 1 ./... -count=1`; `rtk npm run build`; `rtk npm test -- --run`; `rtk npm run test:e2e -- offline-sync.spec.ts` | passed | Initial backend rerun was sandbox-blocked from local Postgres; after local-network rerun, backend/frontend/E2E passed. |
| 4 | Added frontend conflict persistence, mobile conflict inbox, keep-server/discard-local/edit-and-retry actions, full resync, and conflict E2E proof. | `rtk npm test -- --run`; `rtk npm run build`; `rtk npm run test:e2e -- offline-sync.spec.ts`; `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test -p 1 ./... -count=1` | passed | Initial TICKET-010 E2E rerun had a strict text selector collision with fixture wallet names containing "Offline"; narrowed the assertion to the exact status pill and reran successfully. |

Loop guard:

- Same-path failure attempts: 0 / 3
- Total fix/test cycles: 4 / 5
- Blocked: no
- Human/design input needed: none before starting approved PHASE-003 plan.

## Automated Tests

- Passed: `backend/internal/sync`, `backend/internal/platform/httpapi`, full backend regression, `frontend/src/offline/db.test.ts`, `frontend/src/app/outbox.test.ts`, `frontend/src/app/App.test.tsx`, `frontend/e2e/pwa-shell.spec.ts`, and `frontend/e2e/offline-sync.spec.ts`.
- Failed: none yet.
- Skipped: full frontend browser regression remains available before release verification; targeted mobile conflict/replay E2E passed.

## Manual Checks

- Planned: mobile offline UX review, conflict copy review, and storage recovery inspection if automated assertions cannot cover browser quota/degraded mode.

## UAT

- Required: yes.
- Reason if not required: not applicable.
- Expected behavior: mobile PWA supports read/write offline, reconnect syncs each mutation once, and conflicts are reviewed explicitly.
- Verified behavior: TICKET-008 frontend proof covers IndexedDB hydration, local outbox durability/order, legacy migration/quarantine, optimistic offline create for transactions, wallet/category queue primitives, cached-auth offline reload with pending mutation visibility, degraded read-only UI, and PWA shell offline reload. TICKET-009 proof covers idempotent sync mutation replay, user-scoped change feed, authoritative resync, stale-version conflict response, CSRF-authenticated sync routes, frontend sync API drain, and mobile reconnect-once replay. TICKET-010 proof covers stored sync conflicts, non-blocking mobile conflict inbox visibility, keep-server resolution, discard-local mutation removal, edit-and-retry mutation replacement, and full resync mirror restoration while preserving recoverable unsent mutations.
- Sign-off: pending.

## Evidence Notes

- TICKET-008, TICKET-009, and TICKET-010 have automated evidence and are in review. PHASE-003 is in review pending human UAT/sign-off for the offline and conflict recovery experience.
