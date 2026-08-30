---
artifact_type: detail_design
id: TICKET-001-FOLDER-SPLIT-DESIGN
status: approved
owner: shared
human_fields:
  - approval
ai_fields:
  - context
  - proposed_approach
  - verification_plan
shared_fields:
  - trace
trace:
  backlog_item: BL-001
  phase: PHASE-001
  ticket: TICKET-001
  validation_matrix: ../VALIDATION_MATRIX.md
  release_notes: ../../releases/CHANGELOG.md
---

# Detail Design: Backend and Frontend Folder Split

## Context

The repository currently keeps the runnable entrypoints under `apps/` and shared Go packages at root `internal/`. The user requested separate backend and frontend folders before supplying environment configuration.

## Approval

Approved by direct user instruction on 2026-08-30: "move code be vaf FE vao 2 folder rieng".

## Proposed Approach

- Move the Go module into `backend/`.
- Move API, worker, and migration entrypoints from `apps/` to `backend/cmd/`.
- Move shared Go packages to `backend/internal/`.
- Move SQL migrations to `backend/migrations/`.
- Move the React PWA from the prior app workspace into `frontend/`.
- Keep npm package metadata and scripts inside `frontend/`; keep Go module metadata inside `backend/`.
- Update Docker Compose build contexts to use `backend/` and `frontend/`.

## Impact

- API behavior: no route or response contract change.
- Database: no schema or migration content change.
- Security: no auth or authorization contract change.
- Runtime: Dockerfile paths, compose build contexts, and local commands change to explicit `backend/` and `frontend/` working directories.
- Docs: local development, architecture paths, backlog/context, verification artifact, and changelog need reconciliation.

## Verification Plan

- `rtk go test ./...` from `backend/`
- `rtk npm test -- --run` from `frontend/`
- `rtk npm run build` from `frontend/`
- `rtk npm run test:e2e` from `frontend/`
- `rtk docker compose config`
- `rtk bash scripts/smoke-platform.sh`
