---
artifact_type: phase
id: PHASE-003
status: in_progress
owner: human
priority: High
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
    - BL-003
  roadmap: ../ROADMAP.md
  detail_design: PHASE-003-detail-design.md
  requirements:
    - REQ-F-004
    - REQ-NF-002
    - REQ-NF-003
    - REQ-NF-008
  tickets:
    - TICKET-008
    - TICKET-009
    - TICKET-010
  bugs: []
  test_verification: ../test-verification/PHASE-003-offline-sync.md
  validation_matrix: ../VALIDATION_MATRIX.md
  adrs: []
  release_notes: ../../releases/CHANGELOG.md
---

# PHASE-003: Offline Synchronization

## Status

- ID: PHASE-003
- Status: in_progress
- Owner: human
- Priority: High
- Created: 2026-08-24
- Updated: 2026-08-31

## Trace Links

- Backlog: [BACKLOG.md](../BACKLOG.md)
- Roadmap: [ROADMAP.md](../ROADMAP.md)
- Detail design: [PHASE-003-detail-design.md](PHASE-003-detail-design.md) — approved 2026-08-30
- Requirements: [REQUIREMENTS.md](../../requirements/REQUIREMENTS.md) — REQ-F-004, REQ-NF-002, REQ-NF-003, REQ-NF-008
- Test verification: [PHASE-003-offline-sync.md](../test-verification/PHASE-003-offline-sync.md) — planned proof target
- Validation matrix: [VALIDATION_MATRIX.md](../VALIDATION_MATRIX.md)
- ADRs: [decisions](../../decisions/README.md)
- Release notes: [CHANGELOG.md](../../releases/CHANGELOG.md)

## Goal

Deliver full offline read/write with deterministic reconciliation and user-reviewed conflicts.

## Scope

- IndexedDB schema for mirrored records, outbox mutations, cursors, tombstones, and conflicts.
- Optimistic local writes with client UUIDs, mutation IDs, base versions, retry state, and reconnect processing.
- Go batch mutation and incremental change APIs with idempotency ledger and per-user monotonic cursor.
- Conflict inbox and keep-server, edit-and-retry, and discard-local resolution flows.
- IndexedDB/API schema migration and full-resync recovery.

## Out Of Scope

- Background AI/provider ingestion.
- Collaborative real-time editing.
- Silent last-write-wins resolution.

## Tickets And Bugs

| ID | Type | Title | Status | Link |
| --- | --- | --- | --- | --- |
| TICKET-008 | Ticket | IndexedDB mirror and offline outbox | in_progress | [TICKET-008](../tickets/TICKET-008-indexeddb-mirror-outbox.md) |
| TICKET-009 | Ticket | Idempotent sync API and change feed | ready | [TICKET-009](../tickets/TICKET-009-sync-api-change-feed.md) |
| TICKET-010 | Ticket | Conflict inbox and recovery flows | ready | [TICKET-010](../tickets/TICKET-010-conflict-inbox-recovery.md) |

## Dependencies

- PHASE-002 versioned finance commands and authoritative records.

## Risks

- Out-of-order replay can violate user intent.
- Browser quota/eviction can remove local state.
- Cursor or tombstone mistakes can resurrect deleted records.

## Success Criteria

- A user can create and edit supported finance data offline and see optimistic results.
- Reconnect applies each mutation at most once and refreshes authoritative versions.
- Stale edits create explicit conflicts and never overwrite server state silently.
- Full resync recovers from invalid cursor or cleared IndexedDB without duplicating accounting.

## Verification Plan

- Unit tests for ordering, retry, cursor, and merge-state reducers.
- PostgreSQL integration tests for mutation replay, versions, cursor isolation, tombstones, and conflict lifecycle.
- Browser-storage tests for hydration, offline actions, reconnect, partial failure, migration, and full resync.
- E2E/UAT with two browser contexts making conflicting offline edits.

## Gate Checklist

- [x] Phase links approved requirements
- [x] Tickets have stable planned IDs and bounded titles
- [x] Risks and dependencies are recorded
- [x] Verification plan is defined
- [x] Release/changelog need is linked
- [x] Detail design is approved
- [x] Ticket artifacts and detailed implementation plan are created
- [x] Phase status is promoted to ready after plan review

## Completion Summary

TICKET-008 implementation is in progress. The frontend now has an IndexedDB mirror/outbox with localStorage migration, wallet/category/transaction queue primitives, degraded read-only UI, and automated build/unit/PWA shell proof. TICKET-009 sync API/change feed, TICKET-010 conflicts, dedicated offline pending-reload E2E, and human UAT remain pending before PHASE-003 review.
