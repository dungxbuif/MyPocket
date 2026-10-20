---
artifact_type: test_verification
id: TICKET-028-production-hardening-debug-audit-foundation
status: in_review
owner: shared
trace:
  backlog_item: BL-009
  requirement: [REQ-F-015, REQ-NF-004, REQ-NF-007]
  phase: PHASE-007
  ticket_or_bug: [TICKET-028]
  detail_design: ../phases/PHASE-007-production-hardening-debug-audit-detail-design.md
  validation_matrix: ../VALIDATION_MATRIX.md
  docs_review: ../DOCS-REVIEW-TICKET-028.md
  release_notes: ../../releases/CHANGELOG.md
---

# TICKET-028 Production Hardening and Debug Audit Verification

## Status

- Status: in_review
- Evidence: automated backend/frontend proof passed; human UAT remains.

## Planned Commands

| Command | Expected Coverage |
| --- | --- |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/audit ./internal/platform/config ./internal/platform/httpapi ./internal/sync ./internal/worker -count=1` | Unit and HTTP proof for audit/logging/config/sync/worker instrumentation. |
| `rtk env 'MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' GOCACHE=/private/tmp/mypocket-go-cache go test -p 1 ./internal/audit ./internal/platform/db ./internal/platform/httpapi -count=1` | Real PostgreSQL migration, append-only events, retention, and hidden viewer auth. |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./... -count=1` | Full backend regression. |
| `rtk npm test -- --run src/app frontend/src/offline` | Frontend regression. |
| `rtk npm run build` | Production frontend build. |

## Verification Results

| Date | Command | Result | Coverage |
| --- | --- | --- | --- |
| 2026-08-31 | `GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/identity ./internal/platform/authcache ./internal/platform/config ./internal/platform/httpapi ./internal/audit ./internal/portfolio ./internal/worker ./cmd/api ./cmd/worker -count=1 -v` | Pass | API key lifecycle, bearer auth, config validation, audit redaction, hidden viewer auth, panic recovery, worker retention, portfolio lease SQL compile proof. Integration tests requiring `MYPOCKET_TEST_DATABASE_URL` were skipped because the env var was not set in this run. |
| 2026-08-31 | `GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/platform/httpapi -count=1 -v -run 'TestBearerAPIKey\|TestAPIKeys\|TestAPIKeyMutating'` | Pass | API key create syncs cache, revoke deletes cache, list hides plaintext/hash, cached bearer auth avoids DB auth, and API-key mutating requests include actor in audit. |
| 2026-08-31 | `MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable GOCACHE=/private/tmp/mypocket-go-cache go test -p 1 ./internal/identity ./internal/audit ./internal/platform/db ./internal/platform/httpapi -count=1 -v` | Pass | Real PostgreSQL migration chain, identity repository, API key hash-only storage/authenticate/revoke, migration directory checks, and HTTP audit/API key behavior. |
| 2026-08-31 | `GOCACHE=/private/tmp/mypocket-go-cache go test ./... -count=1` | Pass | Full backend regression. |
| 2026-08-31 | `npm test -- --run src/app frontend/src/offline` | Pass | Frontend app/offline regression after Account-tab API key UI addition. |
| 2026-08-31 | `npm run build` | Pass | Production frontend bundle after Account-tab API key UI addition. |
| 2026-08-31 | `MYPOCKET_TEST_REDIS_URL=redis://127.0.0.1:6379/0 GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/platform/authcache -count=1 -v` | Pass | Real Redis SETEX/GET/DEL round trip and deterministic token digest proof. |
| 2026-08-31 | `curl -fsS http://127.0.0.1:8080/api/v1/health/live` and `curl -fsS -I http://127.0.0.1:5173/` | Pass | Direct local API and frontend runtime smoke checks; API returned 200 with correlation ID and Vite returned 200. |
| 2026-08-31 | `GOCACHE=/private/tmp/mypocket-go-cache go test ./... -count=1` | Pass | Full backend regression after receipt upload/presign, transaction attachment, audit-access API, and docs-serving changes. |
| 2026-08-31 | `npm test -- --run src/app frontend/src/offline && npm run build` | Pass | Frontend regression/build after receipt picker, upload client, audit viewer, and IndexedDB receipt queue changes. |
| 2026-08-31 | `GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/platform/httpapi -run TestReceiptUploadAndDownloadAreUserScoped -count=1 -v` | Pass | Receipt upload/download route authorization, metadata validation, presigned URL wiring, and cross-user ownership isolation. |
| 2026-08-31 | `MYPOCKET_TEST_S3_SKIP_BUCKET_CREATE=true GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/platform/objectstore -run TestS3SmokeObjectLifecycle -count=1 -v` | Pass | S3-compatible private bucket put/delete smoke using deployment-provided endpoint and credentials; bucket creation is intentionally skipped because the provisioned account lacks create-bucket permission. |

## Remaining Proof

- Human UAT against the running UAT deployment.
