---
artifact_type: phase
id: PHASE-004
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
    - BL-004
  roadmap: ../ROADMAP.md
  detail_design: PHASE-004-detail-design.md
  requirements:
    - REQ-F-005
    - REQ-F-012
    - REQ-NF-002
    - REQ-NF-005
  tickets:
    - TICKET-011
    - TICKET-012
    - TICKET-013
    - TICKET-014
  bugs: []
  test_verification: ../test-verification/PHASE-004-planning-automation.md
  validation_matrix: ../VALIDATION_MATRIX.md
  adrs: []
  release_notes: ../../releases/CHANGELOG.md
---

# PHASE-004: Planning and Automation

## Status

- ID: PHASE-004
- Status: in_progress
- Owner: human
- Priority: High
- Created: 2026-08-24
- Updated: 2026-08-30

## Trace Links

- Backlog: [BACKLOG.md](../BACKLOG.md)
- Roadmap: [ROADMAP.md](../ROADMAP.md)
- Detail design: [PHASE-004-detail-design.md](PHASE-004-detail-design.md) — approved 2026-08-30
- Requirements: [REQUIREMENTS.md](../../requirements/REQUIREMENTS.md) — REQ-F-005, REQ-F-012, REQ-NF-002, REQ-NF-005
- Test verification: [PHASE-004-planning-automation.md](../test-verification/PHASE-004-planning-automation.md) — planned proof target
- Validation matrix: [VALIDATION_MATRIX.md](../VALIDATION_MATRIX.md)
- ADRs: [decisions](../../decisions/README.md)
- Release notes: [CHANGELOG.md](../../releases/CHANGELOG.md)

## Goal

Deliver budgets, events, recurring drafts, debts/loans, durable notices, and best-effort Web Push.

## Scope

- Weekly, monthly, quarterly, yearly, and custom budgets with category/all-category scope.
- 80% and 100% threshold evaluation with per-budget/threshold/period deduplication.
- Events/trips and reportable grouped totals.
- Recurring schedules whose deterministic occurrences create drafts, not confirmed transactions.
- Borrow/lend obligations and repayment links.
- In-app notification inbox, Web Push subscriptions, capped delivery retry, and worker leases.

## Out Of Scope

- AI/OCR/webhook ingestion.
- Dashboard visualization beyond planning-specific progress.
- Email notifications.

## Tickets And Bugs

| ID | Type | Title | Status | Link |
| --- | --- | --- | --- | --- |
| TICKET-011 | Ticket | Budgets and threshold alerts | in_review | [TICKET-011](../tickets/TICKET-011-budgets-threshold-alerts.md) |
| TICKET-012 | Ticket | Events, debts, and repayments | ready | [TICKET-012](../tickets/TICKET-012-events-debts-repayments.md) |
| TICKET-013 | Ticket | Recurring schedules and worker occurrences | ready | [TICKET-013](../tickets/TICKET-013-recurring-schedules-worker.md) |
| TICKET-014 | Ticket | In-app inbox and Web Push | ready | [TICKET-014](../tickets/TICKET-014-inbox-web-push.md) |

## Dependencies

- PHASE-002 finance/draft interfaces.
- PHASE-003 sync for offline planning mutations where supported.

## Risks

- Timezone boundaries can duplicate or skip scheduled occurrences.
- Push permission and endpoint expiry are outside server control.
- Debt accounting can diverge if repayment links bypass finance rules.

## Success Criteria

- Budget progress and 80%/100% notices are correct and deduplicated per period.
- Events group transactions without changing wallet accounting.
- Each recurring occurrence creates at most one reviewable draft across worker retries/restarts.
- Debt/loan records and repayments reconcile to linked confirmed transactions.
- In-app notices remain available when Web Push is denied or fails.

## Verification Plan

- Go unit tests for periods, thresholds, schedules, and debt calculations.
- PostgreSQL integration tests for leases, occurrence idempotency, notification deduplication, and ownership.
- React component/E2E tests for planning forms, inbox, push-permission states, and draft confirmation.
- Timezone UAT at month/day boundaries in `Asia/Ho_Chi_Minh`.

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

TICKET-011 budgets and threshold alerts are in review with automated backend, frontend, and mobile E2E proof. Remaining PHASE-004 work is TICKET-012 events/debts/repayments, TICKET-013 recurring schedules/worker occurrences, and TICKET-014 in-app inbox/Web Push.
