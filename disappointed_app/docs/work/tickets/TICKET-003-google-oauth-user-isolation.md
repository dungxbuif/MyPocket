---
artifact_type: ticket
id: TICKET-003
status: verified
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
    - REQ-F-001
    - REQ-NF-001
    - REQ-NF-004
  phase: PHASE-001
  detail_design: ../phases/PHASE-001-detail-design.md
  implementation_plan: ../../superpowers/plans/2026-08-25-phase-001-platform-identity.md
  test_verification: ../test-verification/PHASE-001-platform-smoke.md
  validation_matrix: ../VALIDATION_MATRIX.md
  docs_review: per-ticket_completion_checklist
  adrs:
    - ADR-002
  release_notes: ../../releases/CHANGELOG.md
---

# Ticket: TICKET-003 Google OAuth and User Isolation

## Field Ownership

- Human fills intent, priority, acceptance criteria, scope, and approval.
- AI fills impact analysis, test expectations, verification evidence, docs review, and context/backlog updates.
- Shared fields include status, trace links, and small-task exemption.

## Status

- ID: TICKET-003
- Status: verified
- Type: feature
- Priority: urgent
- Phase: PHASE-001
- Owner: human

## Trace Links

- Backlog item: [BL-001](../BACKLOG.md)
- Requirement: [REQ-F-001](../../requirements/REQUIREMENTS.md), [REQ-NF-001](../../requirements/REQUIREMENTS.md), [REQ-NF-004](../../requirements/REQUIREMENTS.md)
- Phase: [PHASE-001](../phases/PHASE-001-platform-identity.md)
- Detail design: [PHASE-001-detail-design.md](../phases/PHASE-001-detail-design.md)
- Implementation plan: [2026-08-25-phase-001-platform-identity.md](../../superpowers/plans/2026-08-25-phase-001-platform-identity.md)
- Test verification: created during execution
- Validation matrix: [VALIDATION_MATRIX.md](../VALIDATION_MATRIX.md)
- Docs review: this ticket completion checklist
- ADRs: [ADR-002](../../decisions/ADR-002-stateless-google-oauth.md)
- Release notes: [CHANGELOG.md](../../releases/CHANGELOG.md)

## Context

Human fill:

- User/business/system problem: Sensitive personal finance data requires authenticated user identities and hard ownership isolation before finance features exist.
- Source prompt or requirement: REQ-F-001, REQ-NF-001, REQ-NF-004, ADR-002.
- Out of scope: linked devices, database sessions, persisted Google tokens, non-Google identity providers, finance object authorization.

AI fill:

- Current repository context read: required hydration docs plus requirements, API, ERD, SDD, ADR-002.
- Brownfield touched scope, if applicable: no existing implementation; expected touched scope is identity module, auth routes, cookie/CSRF helpers, current-user endpoint, and isolation test harness.

## Acceptance Criteria

- [x] Given a valid Google callback fixture, when the callback is processed, then an application user is provisioned or found without persisting provider tokens.
- [x] Given a successful callback, when the response is returned, then it sets a signed `HttpOnly`, `Secure`, `SameSite=Lax` application cookie.
- [x] Given an authenticated request, when `GET /api/v1/me` is called, then the response returns the current application user without exposing provider tokens.
- [x] Given two authenticated users and a user-owned fixture object, when user B requests user A's object through the authorization harness, then the API returns `FORBIDDEN` or scoped `NOT_FOUND`.
- [x] Given a cookie-authenticated mutation, when the CSRF header/token is absent or invalid, then the API rejects it with a stable safe error code.
- [ ] UAT requirement is required for login, refresh persistence, logout, and forbidden state behavior.

## Small Task Exemption

- Small task exemption: no
- Reason: This ticket changes authentication, authorization, security, and public API behavior.
- Impact checked: API=yes, DB=yes, Security=yes, Runtime=yes, Standards=no

## Impacted Areas

- Code: identity domain/application packages, auth HTTP routes, cookies, CSRF, user repository, current-user endpoint, frontend auth state.
- Requirements docs: no change expected.
- Architecture docs: no change expected if implementation follows ADR-002.
- API docs: reconcile concrete request/response schemas and CSRF header name.
- ERD/data docs: reconcile concrete `users` columns if needed.
- Decisions: no new ADR expected unless execution diverges from stateless-cookie design.

## Detail Design

- Required: yes
- Link: [PHASE-001-detail-design.md](../phases/PHASE-001-detail-design.md)
- Approval: approved by user instruction "finish whole app" on 2026-08-25

## Test Expectations

- Unit: OAuth state/nonce validation, cookie signing/verification, CSRF helper behavior, redaction helpers.
- Integration: callback fixture provisions user, no session/provider token persisted, cross-user harness denies access.
- E2E: login callback fixture, authenticated shell, logout, forbidden state.
- UAT: login, refresh persistence, logout, forbidden state.
- Manual/platform: verify cookie flags in browser/devtools or Playwright context output.
- Docs review: API/ERD/architecture reconciliation and ADR check.

## Verification Results

- Command: `rtk go test ./internal/identity`
- Result: fail
- Notes: RED check failed for expected missing identity implementation: `no non-test Go files`.
- Command: `rtk env 'MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55432/mypocket?sslmode=disable' go test ./internal/identity`
- Result: pass
- Notes: Signed cookie, tamper rejection, CSRF missing-header rejection, Google profile provisioning, no provider-token/session columns, and cross-user ownership rejection passed against PostgreSQL.
- Command: `rtk go test ./internal/platform/httpapi`
- Result: fail
- Notes: RED check failed for expected missing auth route contracts and later for CSRF response without correlation ID.
- Command: `rtk env 'MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55432/mypocket?sslmode=disable' go test ./internal/identity ./internal/platform/httpapi`
- Result: pass
- Notes: Identity and auth route tests passed, including fixture callback cookie flags, `/api/v1/me`, logout CSRF envelope, and PostgreSQL-backed repository behavior.
- Command: `rtk go test ./...`
- Result: pass
- Notes: 16 tests passed across 8 Go packages after backend identity implementation.
- Command: `rtk npm run test:e2e` from `frontend/`
- Result: pass
- Notes: Browser E2E verifies fixture Google login, authenticated email display, session persistence after reload, CSRF-backed logout, return to login state, and a safe forbidden state with correlation ID.

## Fix/Test Attempt Log

- Same-path failure attempts: 0 / 3
- Total fix/test cycles: 0 / 5
- Blocked by loop guard: no
- Human/design input needed: none for approved PHASE-001 scope.

## UAT

- Required: yes
- Reason if not required: not applicable
- Expected behavior: login succeeds through fixture, session survives refresh, logout clears cookie, and forbidden states do not leak data.
- Verified behavior: backend fixture login, signed cookie, current-user endpoint, CSRF rejection, ownership harness, browser fixture login, refresh persistence, logout, and forbidden state are verified by automated tests.
- Sign-off: automated UAT evidence recorded; human review pending.

## Docs Review

- Requirements updated or not needed reason: not needed; implementation follows REQ-F-001, REQ-NF-001, and REQ-NF-004.
- Architecture updated or not needed reason: not needed; follows ADR-002 stateless cookie design.
- API updated or not needed reason: updated with implemented auth routes and CSRF header contract.
- ERD/data updated or not needed reason: not needed; users schema already reconciled in TICKET-002 and no provider-token/session columns were added.
- ADR created or not needed reason: not needed; follows ADR-002.
- `docs/CONTEXT.md` updated: pending after auth browser E2E commit.

## Completion Checklist

- [x] Implementation complete
- [x] Tests run and recorded
- [x] Fix/test loop guard respected
- [x] Validation matrix updated or explicitly not affected
- [x] UAT completed or explicitly not required
- [x] Master docs reconciled
- [x] Docs review completed
- [x] ADR created or explicitly not needed
- [x] `docs/CONTEXT.md` updated
- [x] `docs/work/BACKLOG.md` updated
- [x] Trace links updated
