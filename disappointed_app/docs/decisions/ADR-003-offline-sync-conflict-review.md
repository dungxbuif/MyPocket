---
artifact_type: adr
id: ADR-003
status: accepted
owner: shared
human_fields:
  - decision_approval
  - final_status
ai_fields:
  - context
  - alternatives_considered
  - consequences
  - linked_work
shared_fields:
  - decision
  - trace
trace:
  requirements: [REQ-F-004, REQ-NF-002, REQ-NF-003]
  phase: PHASE-003
  tickets_or_bugs: [TICKET-008, TICKET-009, TICKET-010]
  detail_design: ../superpowers/specs/2026-08-23-mypocket-system-design.md
  master_docs:
    - ../requirements/SPEC.md
    - ../architecture/ARCHITECTURE.md
    - ../architecture/API.md
    - ../architecture/ERD.md
  release_notes: ../releases/CHANGELOG.md
---

# ADR-003: Offline Sync with Explicit Conflict Review

## Status

- Status: accepted
- Date: 2026-08-24
- Decision approval: user-approved during design review

## Context

The PWA must support full offline read/write across multiple devices. Silent last-write-wins can discard financial edits and server-always-wins can silently lose offline work.

## Decision

Use IndexedDB as a local mirror and outbox. Mutations carry client IDs, idempotency keys, and base versions. PostgreSQL remains authoritative and emits a per-user change cursor. Stale writes create explicit conflicts for keep-server, edit-and-retry, or discard-local resolution.

## Alternatives Considered

- Last write wins: rejected because it silently overwrites concurrent work.
- Server always wins: rejected because it silently discards offline work.
- Offline capture only: rejected because the user approved full offline read/write.

## Consequences

- Positive: no silent data loss and deterministic retry/idempotency behavior.
- Negative: requires cursor, tombstone, IndexedDB migration, and conflict UI complexity.
- Neutral: the server remains authoritative even while local optimistic state drives the offline UX.

## Linked Work

- Approved design: [MyPocket System Design](../superpowers/specs/2026-08-23-mypocket-system-design.md)
- Requirements: [REQUIREMENTS.md](../requirements/REQUIREMENTS.md)
- Roadmap: [ROADMAP.md](../work/ROADMAP.md)
- Architecture: [ARCHITECTURE.md](../architecture/ARCHITECTURE.md)
- API: [API.md](../architecture/API.md)
- ERD: [ERD.md](../architecture/ERD.md)
- Integrations: [INTEGRATIONS.md](../architecture/INTEGRATIONS.md)

