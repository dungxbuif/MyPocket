---
artifact_type: ticket
id: TICKET-005
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
  backlog_item: BL-002
  requirement:
    - REQ-F-002
  phase: PHASE-002
  detail_design: ../phases/PHASE-002-detail-design.md
  implementation_plan: ../../superpowers/plans/2026-08-30-phase-002-finance-core.md
  test_verification: ../test-verification/PHASE-002-finance-core.md
  validation_matrix: ../VALIDATION_MATRIX.md
  docs_review: per-ticket_completion_checklist
  adrs:
    - ADR-001
  release_notes: ../../releases/CHANGELOG.md
---

# Ticket: TICKET-005 Wallet and Category Domain

## Status

- ID: TICKET-005
- Status: ready
- Type: feature
- Priority: high
- Phase: PHASE-002
- Owner: human

## Context

Human fill:

- User/business/system problem: Users need wallets and categories that match their real accounts and Vietnamese spending habits before transactions can be correct.
- Source prompt or requirement: REQ-F-002 and PHASE-002.
- Out of scope: transaction posting, offline sync, analytics, OCR, AI, and planning budgets.

AI fill:

- Current repository context read: `docs/CONTEXT.md`, `docs/work/BACKLOG.md`, standards, PHASE-002, requirements, architecture, ERD, API, ADR-001, ADR-003.
- Brownfield touched scope, if applicable: new `backend/internal/finance`, finance migrations, HTTP route registration, and frontend wallet/category screens under `frontend/src`.

## Acceptance Criteria

- [ ] Given an authenticated user, when wallets are created, edited, archived, listed, and selected as default AI wallet, then all operations are scoped to that user and at most one active default AI wallet exists.
- [ ] Given supported wallet types, when credit-card metadata is supplied, then metadata is accepted only for credit wallets and ordinary wallets keep VND integer balance fields.
- [ ] Given system and user categories, when categories are listed or modified, then depth is limited to two levels, system categories are locked, user categories are archiveable, and referenced history is not deleted.
- [ ] Given a wallet/category activation change, when the user toggles availability, then the setting applies only to that user's wallet and category.
- [ ] UAT requirement is required for Vietnamese copy, VND formatting, wallet archive/default behavior, and category activation.

## Small Task Exemption

- Small task exemption: no
- Reason: This ticket introduces finance data model, public API, authorization, and user-facing wallet/category behavior.
- Impact checked: API=yes, DB=yes, Security=yes, Runtime=no, Standards=no

## Impacted Areas

- Code: finance domain package, migration, HTTP routes, frontend wallet/category views.
- Requirements docs: no change expected.
- Architecture docs: reconcile implemented finance module files.
- API docs: add concrete wallet/category contracts.
- ERD/data docs: add implemented wallet/category tables.
- Decisions: no new ADR expected; follows ADR-001.

## Detail Design

- Required: yes
- Link: [PHASE-002-detail-design.md](../phases/PHASE-002-detail-design.md)
- Approval: approved by user instruction to complete the app; design remains available for human review.

## Test Expectations

- Unit: wallet/category validation, default wallet selection, category depth and lock rules.
- Integration: PostgreSQL constraints, ownership scoping, archive semantics, activation uniqueness.
- E2E: mobile wallet/category list and edit flows.
- UAT: required; automated browser proof plus human review.
- Manual/platform: not required beyond existing Compose.
- Docs review: API, ERD, validation matrix, context, backlog, changelog.

## Verification Results

- Command: not_run
- Result: pending
- Notes: Implementation has not started.

## Fix/Test Attempt Log

- Same-path failure attempts: 0 / 3
- Total fix/test cycles: 0 / 5
- Blocked by loop guard: no
- Human/design input needed: none before starting approved PHASE-002 plan.

## UAT

- Required: yes
- Reason if not required: not applicable
- Expected behavior: wallet and category management match Vietnamese finance copy and VND formatting.
- Verified behavior: pending implementation.
- Sign-off: pending.

## Docs Review

- Requirements updated or not needed reason: pending execution.
- Architecture updated or not needed reason: pending execution.
- API updated or not needed reason: pending execution.
- ERD/data updated or not needed reason: pending execution.
- ADR created or not needed reason: pending execution.
- `docs/CONTEXT.md` updated: pending execution.

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
- [ ] Trace links updated
