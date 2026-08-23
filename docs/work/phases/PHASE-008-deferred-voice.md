---
artifact_type: phase
id: PHASE-008
status: draft
owner: human
priority: Low
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
    - BL-008
  roadmap: ../ROADMAP.md
  requirements:
    - REQ-F-017
    - REQ-F-011
    - REQ-NF-004
    - REQ-NF-007
  tickets:
    - TICKET-026
  bugs: []
  test_verification: not_created_phase_not_executed
  validation_matrix: ../VALIDATION_MATRIX.md
  adrs: []
  release_notes: ../../releases/CHANGELOG.md
---

# PHASE-008: Deferred Voice Input

## Status

- ID: PHASE-008
- Status: draft
- Owner: human
- Priority: Low
- Created: 2026-08-24
- Updated: 2026-08-24

## Trace Links

- Backlog: [BACKLOG.md](../BACKLOG.md)
- Roadmap: [ROADMAP.md](../ROADMAP.md)
- Requirements: [REQUIREMENTS.md](../../requirements/REQUIREMENTS.md) — REQ-F-017, REQ-F-011, REQ-NF-004, REQ-NF-007
- Test verification: created during execution
- Validation matrix: [VALIDATION_MATRIX.md](../VALIDATION_MATRIX.md)
- ADRs: [decisions](../../decisions/README.md)
- Release notes: [CHANGELOG.md](../../releases/CHANGELOG.md)

## Goal

Add recorded voice transaction entry through OpenAI-compatible transcription without changing the established review-first draft contract.

## Scope

- Browser audio capture/upload with permission, size, type, cancellation, and retry states.
- OpenAI-compatible transcription adapter configured by environment.
- Transcript preview and submission through the existing text AI draft pipeline.
- Privacy, deletion, provider-error, component, integration, E2E, and UAT proof.

## Out Of Scope

- Initial release scope.
- Real-time voice assistant or continuous listening.
- Automatic transaction confirmation.

## Tickets And Bugs

| ID | Type | Title | Status | Link |
| --- | --- | --- | --- | --- |
| TICKET-026 | Ticket | Voice capture, transcription, and draft integration | planned | Created during implementation planning |

## Dependencies

- PHASE-006 shared AI conversation and draft pipeline.
- Explicit product decision to start deferred voice work after PHASE-001 through PHASE-007 are verified.

## Risks

- Browser recording support and permissions vary.
- Audio contains sensitive information and increases provider/storage privacy surface.

## Success Criteria

- Supported browsers record or upload audio and return a transcript preview.
- Confirmed transcript requests create the same editable drafts as text chat.
- Denied permission, unsupported format, timeout, and malformed transcription do not create financial records.

## Verification Plan

- Unit/contract tests for transcription adapter and error mapping.
- React component tests for permissions, capture, preview, cancel, and retry.
- E2E/UAT on supported desktop and installed mobile PWA browsers.
- Docs review of privacy, provider, and operational changes.

## Gate Checklist

- [x] Phase links approved requirements
- [x] Tickets have stable planned IDs and bounded titles
- [x] Risks and dependencies are recorded
- [x] Verification plan is defined
- [x] Release/changelog need is linked
- [ ] Ticket artifacts and detailed implementation plan are created
- [ ] Phase status is promoted to ready after plan review

## Completion Summary

No implementation has started. Completion evidence will be recorded after the phase reaches execution.

