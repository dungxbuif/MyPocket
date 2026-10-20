# PHASE-004 Planning and Automation Implementation Plan

**Goal:** Deliver budgets, events, debts, recurring drafts, in-app notices, and best-effort Web Push without bypassing finance accounting.

**Spec:** `docs/work/phases/PHASE-004-detail-design.md`

## Global Constraints

- Every shell command starts with `rtk`.
- Server computes time periods in `Asia/Ho_Chi_Minh`.
- Planning never modifies wallet balances directly; confirmed drafts must call PHASE-002 accounting.
- In-app notifications are authoritative; Web Push is optional best-effort delivery.
- Worker actions must be idempotent under retries and restarts.

## Tasks

- [ ] **Task 1: Budgets and Threshold Alerts**
  - Files: `backend/internal/planning`, budget migrations, `backend/internal/platform/httpapi`, `frontend/src/planning`.
  - Implement budget CRUD, category scopes, progress formulas, threshold dedupe, mobile screens.
  - Verify with: `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/planning -run Budget -count=1`

- [ ] **Task 2: Events, Debts, and Repayments**
  - Files: planning migrations/domain/API/frontend planning screens.
  - Implement event grouping, obligation CRUD, repayment links to confirmed transactions, ownership checks.
  - Verify with: `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/planning -run 'Event|Obligation|Repayment' -count=1`

- [ ] **Task 3: Recurring Schedules and Worker Occurrences**
  - Files: `backend/internal/planning`, `backend/internal/worker`, worker command, frontend schedule/draft UI.
  - Implement normalized recurrence, worker leases, deterministic occurrence keys, draft generation.
  - Verify with: `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/planning ./internal/worker -run 'Recurring|Lease|Occurrence' -count=1`

- [ ] **Task 4: In-App Inbox and Web Push**
  - Files: notifications/push migrations, planning notice service, worker delivery, frontend inbox.
  - Implement durable inbox, subscription lifecycle, redaction, capped retry, expired endpoint cleanup.
  - Verify with: `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/planning ./internal/worker -run 'Notification|Push' -count=1`

- [ ] **Task 5: Reconciliation and Release Proof**
  - Run: `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test -p 1 ./... -count=1`
  - Run: `rtk npm test -- --run`
  - Run: `rtk npm run build`
  - Run: `rtk npm run test:e2e -- planning-automation.spec.ts`
  - Update verification, tickets, phase, validation matrix, backlog, context, changelog, API, ERD, architecture.
