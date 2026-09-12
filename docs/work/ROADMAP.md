---
artifact_type: roadmap
id: ROADMAP
status: draft
owner: human
human_fields:
  - milestones
  - priority
  - phase_order
ai_fields:
  - phase_links
  - status_summaries
shared_fields:
  - milestone_status
updated: 2026-09-13
---

# Roadmap

## Field Ownership

- Human owns milestones, priority, and phase order.
- AI maintains phase links and status summaries.

No implementation roadmap has been scheduled for the current product specification. The owner is reviewing the complete function design in [SPEC.md](../requirements/SPEC.md) and its [parent/child business tickets](tickets/README.md). The draft tickets describe business scope; their numbering does not assign implementation phases or priority.

## Roadmap Rules

- A roadmap item becomes executable only after it has a phase file in `docs/work/phases/`.
- Each phase should contain one or more tickets or bugs.
- Completed phases should link to release notes or changelog entries when relevant.

## Milestones

| Milestone | Goal | Status | Phase Files |
| --- | --- | --- | --- |

No project milestones have been assigned. The inherited Harness CLI example is not MyPocket product work and has been removed from this scheduler.
