---
artifact_type: ticket
id: TICKET-001
status: in_progress
owner: human
priority: urgent
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
  backlog_item: BL-001
  requirement:
    - REQ-NF-008
  phase: PHASE-001
  detail_design: ../phases/PHASE-001-detail-design.md
  implementation_plan: ../../superpowers/plans/2026-08-25-phase-001-platform-identity.md
  test_verification: not_created_phase_not_executed
  validation_matrix: ../VALIDATION_MATRIX.md
  docs_review: per-ticket_completion_checklist
  adrs:
    - ADR-001
  release_notes: ../../releases/CHANGELOG.md
---

# Ticket: TICKET-001 Repository and Runtime Foundation

## Field Ownership

- Human fills intent, priority, acceptance criteria, scope, and approval.
- AI fills impact analysis, test expectations, verification evidence, docs review, and context/backlog updates.
- Shared fields include status, trace links, and small-task exemption.

## Status

- ID: TICKET-001
- Status: in_progress
- Type: feature
- Priority: urgent
- Phase: PHASE-001
- Owner: human

## Trace Links

- Backlog item: [BL-001](../BACKLOG.md)
- Requirement: [REQ-NF-008](../../requirements/REQUIREMENTS.md)
- Phase: [PHASE-001](../phases/PHASE-001-platform-identity.md)
- Detail design: [PHASE-001-detail-design.md](../phases/PHASE-001-detail-design.md)
- Implementation plan: [2026-08-25-phase-001-platform-identity.md](../../superpowers/plans/2026-08-25-phase-001-platform-identity.md)
- Test verification: created during execution
- Validation matrix: [VALIDATION_MATRIX.md](../VALIDATION_MATRIX.md)
- Docs review: this ticket completion checklist
- ADRs: [ADR-001](../../decisions/ADR-001-react-go-modular-monolith.md)
- Release notes: [CHANGELOG.md](../../releases/CHANGELOG.md)

## Context

Human fill:

- User/business/system problem: MyPocket needs the first runnable React PWA and Go backend repository foundation before any finance or sync behavior can be implemented.
- Source prompt or requirement: PHASE-001 scope and approved React PWA plus Go modular-monolith design.
- Out of scope: Google OAuth callback, persistent data migrations, S3 behavior, full offline sync, wallets, transactions, analytics, AI, OCR, and production credentials.

AI fill:

- Current repository context read: `docs/CONTEXT.md`, `docs/work/BACKLOG.md`, standards, PHASE-001, requirements, architecture, API, ERD, SDD, ADR-001, ADR-002.
- Brownfield touched scope, if applicable: no product code exists; expected touched scope is new repository scaffold only.

## Acceptance Criteria

- [ ] Given a clean checkout, when dependencies are installed, then `apps/web`, `apps/api`, `apps/worker`, and shared Go packages have documented commands that run locally.
- [ ] Given an API request, when the request is handled, then responses include stable JSON error envelopes and correlation IDs without stack traces or secrets.
- [ ] Given the React shell loads on mobile width, when the app renders, then it shows an installable PWA shell with mobile-safe layout and no finance behavior.
- [ ] Given the browser is offline after the shell has loaded once, when the app is reopened, then the app shell loads from the service worker cache and shows an offline state.
- [ ] UAT requirement is required for installability, mobile shell layout, and offline app-shell reload.

## Small Task Exemption

- Small task exemption: no
- Reason: This establishes runtime, public API error behavior, major dependencies, and user-visible PWA behavior.
- Impact checked: API=yes, DB=no, Security=no, Runtime=yes, Standards=no

## Impacted Areas

- Code: create `apps/web`, `apps/api`, `apps/worker`, `internal/platform`, repository config, package manifests.
- Requirements docs: no change expected unless scope changes.
- Architecture docs: no change expected if implementation follows approved SDD.
- API docs: may need endpoint schema details for health/error envelopes.
- ERD/data docs: no change expected.
- Decisions: no new ADR expected; follows ADR-001.

## Detail Design

- Required: yes
- Link: [PHASE-001-detail-design.md](../phases/PHASE-001-detail-design.md)
- Approval: approved by user instruction "finish whole app" on 2026-08-25

## Test Expectations

- Unit: Go tests for config parsing and error envelope helpers; React tests for shell/offline state components.
- Integration: API health endpoint and correlation-ID middleware tests.
- E2E: mobile viewport shell render, route fallback, service worker registration.
- UAT: installed/mobile PWA shell and offline reload after first online load.
- Manual/platform: command-level smoke for local web/API/worker startup.
- Docs review: ticket checklist plus context/backlog reconciliation after execution.

## Verification Results

- Command: `rtk go test ./internal/platform/config` after adding tests before implementation
- Result: fail
- Notes: RED check failed for expected missing-package reason: `no non-test Go files`.
- Command: `rtk go test ./internal/platform/httpapi` after adding tests before implementation
- Result: fail
- Notes: RED check failed for expected missing-package reason: `no non-test Go files`.
- Command: `rtk go test ./internal/platform/config`
- Result: pass
- Notes: 3 config tests passed after minimal implementation.
- Command: `rtk go test ./internal/platform/httpapi`
- Result: pass
- Notes: 4 HTTP API tests passed after minimal implementation.
- Command: `rtk go test ./...`
- Result: pass
- Notes: 7 tests passed across 4 Go packages for the Task 1 backend baseline.

## Fix/Test Attempt Log

- Same-path failure attempts: 0 / 3
- Total fix/test cycles: 0 / 5
- Blocked by loop guard: no
- Human/design input needed: none for approved PHASE-001 scope.

## UAT

- Required: yes
- Reason if not required: not applicable
- Expected behavior: app installs or qualifies as installable, renders mobile shell, and reopens while offline after first load.
- Verified behavior: not verified; no implementation exists.
- Sign-off: pending

## Docs Review

- Requirements updated or not needed reason: not yet executed.
- Architecture updated or not needed reason: not yet executed.
- API updated or not needed reason: not yet executed.
- ERD/data updated or not needed reason: not yet executed.
- ADR created or not needed reason: not yet executed.
- `docs/CONTEXT.md` updated: pending execution/dehydration.

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
