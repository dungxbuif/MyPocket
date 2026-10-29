# PHASE-007 Audit, Export, Account Lifecycle, and Production Operations Implementation Plan

**Goal:** Add redacted auditability, portable exports, exact-scope account lifecycle jobs, and production release proof.

**Spec:** `docs/work/phases/PHASE-007-detail-design.md`

## Global Constraints

- Every shell command starts with `rtk`.
- Hidden UI is not a security boundary; backend exact-email authorization is authoritative.
- Destructive jobs require recent auth, typed confirmation, CSRF, idempotency, and previewed affected counts.
- Export/download and S3 deletion must be user-scoped.
- No destructive down migrations; rollback guidance uses forward fixes.

## Tasks

- [ ] **Task 1: Audit Pipeline and Hidden Viewer**
  - Files: audit migration, `backend/internal/audit`, audit HTTP routes, hidden frontend page.
  - Implement append-only redacted audit events, exact `AUDIT_VIEWER_EMAIL` authorization, retention purge.
  - Verify with: `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/audit -count=1`

- [ ] **Task 2: Manual Export Jobs**
  - Files: export migration/domain/worker/API, account export UI, private S3 output.
  - Implement snapshot jobs, CSV/ZIP formatting, presigned downloads, progress/failure states.
  - Verify with: `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/export -count=1`

- [ ] **Task 3: Account Reset and Deletion**
  - Files: lifecycle migration/domain/worker/API, account danger-zone UI.
  - Implement recent-auth checks, typed confirmation, exact DB/S3 deletion plan, resumable jobs, tombstone evidence.
  - Verify with: `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/lifecycle -count=1`

- [ ] **Task 4: Production Hardening**
  - Files: config validation, health checks, worker restart handling, deployment docs/scripts.
  - Implement required env validation, readiness/liveness split, backup/restore scripts or documented commands, release smoke.
  - Verify with: `rtk ./scripts/smoke-platform.sh`

- [ ] **Task 5: Reconciliation and Release Proof**
  - Run: `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test -p 1 ./... -count=1`
  - Run: `rtk npm test -- --run`
  - Run: `rtk npm run build`
  - Run: `rtk npm run test:e2e -- audit-export-production.spec.ts`
  - Run: `rtk ./scripts/smoke-platform.sh`
  - Update verification, tickets, phase, validation matrix, backlog, context, changelog, API, ERD, deployment, architecture.
