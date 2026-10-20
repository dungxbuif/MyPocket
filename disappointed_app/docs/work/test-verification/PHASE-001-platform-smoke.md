---
artifact_type: test_verification
id: PHASE-001-platform-smoke
status: verified
owner: shared
---

# PHASE-001 Platform Smoke

- `docker compose up -d --build` — pass; PostgreSQL, LocalStack S3, migration, API, worker, and nginx web started (worker remains running until SIGTERM).
- `bash scripts/smoke-platform.sh` — pass; API liveness/readiness returned correlation IDs and database `ok`; LocalStack reported S3 `available`.
- `curl -fsS http://127.0.0.1:5173/` — pass; web shell served.
- `docker compose config` — pass.
- `cd frontend && npm run test:e2e` — pass; mobile shell exposed manifest metadata, registered the service worker, and reloaded offline.
- `rtk go test ./...` from `backend/` — pass after moving the Go module and packages into `backend/`.
- `rtk npm test -- --run` from `frontend/` — pass after moving npm package ownership into `frontend/`.
- `rtk npm run build` from `frontend/` — pass after moving the PWA into `frontend/`.
- `rtk npm run test:e2e` from `frontend/` — pass after the folder split; offline reload still works.
- `rtk docker compose config` — pass after backend and frontend build contexts were split.
- `rtk docker compose up -d --build` — pass after rebuilding images from `backend/` and `frontend/`.
- `rtk bash scripts/smoke-platform.sh` — pass after the split; API readiness, web-origin `/api` proxy health, and LocalStack S3 health remain reachable.
- `rtk npm run test:e2e` from `frontend/` — pass with 3 mobile browser tests covering service worker offline reload, fixture login persistence/logout, and forbidden auth state.

UAT is not required; this operational ticket is covered by platform evidence. Production secrets, TLS, backup/restore, and monitoring remain out of scope.
