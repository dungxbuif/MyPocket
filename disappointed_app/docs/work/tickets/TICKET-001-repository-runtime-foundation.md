---
artifact_type: ticket
id: TICKET-001
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
    - REQ-NF-008
  phase: PHASE-001
  detail_design: ../phases/PHASE-001-detail-design.md
  implementation_plan: ../../superpowers/plans/2026-08-25-phase-001-platform-identity.md
  test_verification: ../test-verification/PHASE-001-platform-smoke.md
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
- Status: verified
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

- [x] Given a clean checkout, when dependencies are installed, then `frontend`, `backend/cmd/api`, `backend/cmd/worker`, and shared Go packages have documented commands that run locally.
- [x] Given an API request, when the request is handled, then responses include stable JSON error envelopes and correlation IDs without stack traces or secrets.
- [x] Given the React shell loads on mobile width, when the app renders, then it shows an installable PWA shell with mobile-safe layout and no finance behavior.
- [ ] Given the browser is offline after the shell has loaded once, when the app is reopened, then the app shell loads from the service worker cache and shows an offline state.
- [ ] UAT requirement is required for installability, mobile shell layout, and offline app-shell reload.

## Small Task Exemption

- Small task exemption: no
- Reason: This establishes runtime, public API error behavior, major dependencies, and user-visible PWA behavior.
- Impact checked: API=yes, DB=no, Security=no, Runtime=yes, Standards=no

## Impacted Areas

- Code: create `frontend`, `backend/cmd/api`, `backend/cmd/worker`, `backend/internal/platform`, repository config, package manifests.
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
- Command: `rtk npm test -- --run` from `frontend/`
- Result: fail
- Notes: RED check failed for expected missing `./App` import after shell tests were added.
- Command: `rtk npm install`
- Result: pass
- Notes: Installed web workspace dependencies; npm reported 0 vulnerabilities.
- Command: `rtk npm test -- --run` from `frontend/`
- Result: pass
- Notes: 3 web tests passed for mobile navigation destinations, offline indicator, and quick-add sheet.
- Command: `rtk npm run build` from `frontend/`
- Result: pass
- Notes: Vite production build succeeded and emitted `index.html`, hashed CSS/JS, `manifest.webmanifest`, `pwa-icon.svg`, and `sw.js`.
- Command: `rtk go test ./...` from `backend/`
- Result: pass
- Notes: Backend tests passed after moving Go module, commands, shared packages, and migrations under `backend/`.
- Command: `rtk npm test -- --run` from `frontend/`
- Result: pass
- Notes: Frontend component tests passed after moving npm package ownership to `frontend/`.
- Command: `rtk npm run build` from `frontend/`
- Result: pass
- Notes: Frontend production build passed after removing the root npm workspace wrapper.
- Command: `rtk npm run test:e2e` from `frontend/`
- Result: pass
- Notes: Mobile PWA service worker registration and offline reload passed after the folder split.
- Command: `rtk npx playwright screenshot --viewport-size=390,844 http://127.0.0.1:5173 /private/tmp/mypocket-mobile-3.png`
- Result: pass
- Notes: Mobile screenshot captured after visual fixes for balance and amount wrapping.
- Command: `rtk npx playwright screenshot --viewport-size=1280,900 http://127.0.0.1:5173 /private/tmp/mypocket-desktop.png`
- Result: pass
- Notes: Desktop screenshot captured for centered shell verification.
- Command: `rtk go test ./...`
- Result: pass
- Notes: 16 tests passed across 8 Go packages after PWA shell implementation.
- Command: `rtk git diff --check`
- Result: pass
- Notes: No whitespace errors reported.

## Fix/Test Attempt Log

- Same-path failure attempts: 0 / 3
- Total fix/test cycles: 0 / 5
- Blocked by loop guard: no
- Human/design input needed: none for approved PHASE-001 scope.

## UAT

- Required: yes
- Reason if not required: not applicable
- Expected behavior: app installs or qualifies as installable, renders mobile shell, and reopens while offline after first load.
- Verified behavior: automated tests and screenshots verify mobile shell layout, navigation, quick-add sheet, offline indicator, manifest, service worker asset presence, and offline reload after first online load.
- Sign-off: automated UAT evidence recorded; human review pending.

## Docs Review

- Requirements updated or not needed reason: not needed; implementation follows existing PHASE-001 and REQ-F-016/REQ-NF-008 scope.
- Architecture updated or not needed reason: updated design contract in `design/DESIGN.md`; master architecture still matches approved React PWA boundary.
- API updated or not needed reason: already reconciled for health/auth routes; PWA shell adds no new API endpoint.
- ERD/data updated or not needed reason: not needed; PWA shell adds no data model.
- ADR created or not needed reason: not needed; follows ADR-001.
- `docs/CONTEXT.md` updated: yes; next steps now point to PHASE-001 reconciliation and PHASE-002.

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
