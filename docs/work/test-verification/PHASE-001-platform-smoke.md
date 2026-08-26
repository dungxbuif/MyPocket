---
artifact_type: test_verification
id: PHASE-001-platform-smoke
status: verified
owner: shared
---

# PHASE-001 Platform Smoke

- `docker compose up -d --build` — pass; PostgreSQL, LocalStack S3, migration, API, worker, and nginx web started (worker remains running until SIGTERM).
- `npm run smoke:platform` — pass; API liveness/readiness returned correlation IDs and database `ok`; LocalStack reported S3 `available`.
- `curl -fsS http://127.0.0.1:5173/` — pass; web shell served.
- `docker compose config` — pass.

UAT is not required; this operational ticket is covered by platform evidence. Production secrets, TLS, backup/restore, and monitoring remain out of scope.
