---
artifact_type: phase
id: PHASE-002
status: draft
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
    - BL-002
  roadmap: ../ROADMAP.md
  requirements:
    - REQ-F-002
    - REQ-F-003
    - REQ-F-011
    - REQ-NF-002
    - REQ-NF-005
  tickets:
    - TICKET-005
    - TICKET-006
    - TICKET-007
  bugs: []
  test_verification: not_created_phase_not_executed
  validation_matrix: ../VALIDATION_MATRIX.md
  adrs: []
  release_notes: ../../releases/CHANGELOG.md
---

# PHASE-002: Finance Core

## Status

- ID: PHASE-002
- Status: draft
- Owner: human
- Priority: High
- Created: 2026-08-24
- Updated: 2026-08-24

## Trace Links

- Backlog: [BACKLOG.md](../BACKLOG.md)
- Roadmap: [ROADMAP.md](../ROADMAP.md)
- Requirements: [REQUIREMENTS.md](../../requirements/REQUIREMENTS.md) — REQ-F-002, REQ-F-003, REQ-F-011, REQ-NF-002, REQ-NF-005
- Test verification: created during execution
- Validation matrix: [VALIDATION_MATRIX.md](../VALIDATION_MATRIX.md)
- ADRs: [decisions](../../decisions/README.md)
- Release notes: [CHANGELOG.md](../../releases/CHANGELOG.md)

## Goal

Deliver correct wallet accounting, category management, transaction workflows, and stable Vietnamese seed data.

## Scope

- Wallet types, credit-card metadata, archive behavior, include-in-total, and one active default AI wallet.
- Two-level category taxonomy, stable Vietnamese system seeds, system locks, and wallet-category activation.
- Income, expense, atomic transfer, balance adjustment, edit/archive/search, report exclusion, with-person, event reference, and receipt metadata.
- Idempotent finance commands, version fields, audit-event interface, and finance REST contracts.

## Out Of Scope

- Offline IndexedDB and change-cursor synchronization.
- Planning modules and analytics charts.
- OCR or AI provider calls.

## Tickets And Bugs

| ID | Type | Title | Status | Link |
| --- | --- | --- | --- | --- |
| TICKET-005 | Ticket | Wallet and category domain | planned | Created during implementation planning |
| TICKET-006 | Ticket | Transaction accounting and transfer engine | planned | Created during implementation planning |
| TICKET-007 | Ticket | Vietnamese seed data and receipt metadata | planned | Created during implementation planning |

## Dependencies

- PHASE-001 platform, identity, database, object-store, and authorization primitives.

## Risks

- Incorrect sign or transfer behavior can corrupt balances.
- Archive and edit semantics can make historical reports inconsistent.
- Seed identifiers must remain stable across environments.

## Success Criteria

- Wallet and category CRUD obey ownership, depth, lock, activation, default-wallet, and archive rules.
- Income, expense, transfer, and adjustment commands produce exact atomic balances and remain idempotent under retry.
- Transaction search and wallet scope return only authenticated-user data.
- System seed migration installs the approved Vietnamese taxonomy with stable identifiers.

## Verification Plan

- Table-driven Go unit tests for every accounting effect and validation branch.
- PostgreSQL integration tests for atomicity, idempotency, ownership, constraints, archive behavior, and seeds.
- React component/E2E tests for wallet, category, add/edit/search, transfer, and adjustment flows.
- UAT for Vietnamese copy, VND formatting, credit wallet behavior, and balance privacy.

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

