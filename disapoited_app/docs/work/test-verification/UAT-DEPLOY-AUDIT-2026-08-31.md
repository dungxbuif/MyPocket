---
artifact_type: deploy_uat_audit
id: UAT-DEPLOY-AUDIT-2026-08-31
status: in_progress
owner: shared
trace:
  backlog_item: BL-005
  phase: PHASE-005
  ticket_or_bug: [TICKET-027]
  validation_matrix: ../VALIDATION_MATRIX.md
  release_notes: ../../releases/CHANGELOG.md
---

# UAT Deploy Audit 2026-08-31

## Passed Checks

| Area | Evidence | Result |
| --- | --- | --- |
| Backend regression | `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./... -count=1` | pass |
| Portfolio PostgreSQL proof | `rtk env 'MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/portfolio -count=1 -v` | pass |
| Migration proof | `rtk env 'MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/platform/db -count=1 -v` | pass |
| Auth whitelist regression | `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/platform/config ./internal/platform/httpapi -run 'Load|OAuth|CurrentUser|CORS|Logout' -count=1` | pass |
| Frontend app regression | `rtk npm test -- --run src/app` | pass |
| Frontend production build | `rtk npm run build` | pass |
| Whitespace | `rtk git diff --check` | pass |

## Current Prod/UAT Gaps

| Gap | Status | Required Before Prod |
| --- | --- | --- |
| Real Google credentials in local `.env` | found | Do not commit `.env`; rotate the secret if it was exposed outside the machine; configure prod secret manager/env separately. |
| Login allowlist | implemented | Set `ALLOWED_LOGIN_EMAILS` in prod to comma-separated approved users. Current local dev value is `dungbui.dungbui.00@gmail.com`. |
| Asset portfolio sync/offline mutation replay | pending | Add sync entity support and IndexedDB outbox before claiming full PWA offline writes for assets. |
| Asset price provider job | pending | Add provider adapter registry, env config, leased refresh worker, and provider tests; manual price entry works first. |
| Portfolio E2E/UAT | pending | Run mobile and desktop flows for create asset, buy, sell, update price, archive, privacy masking, and dashboard totals. |
| Receipt image upload | implemented (OCR deferred) | Metadata persistence, user-scoped presigned S3 upload/download, transaction attachment, and durable IndexedDB retry queue are implemented. Run UAT with production S3 configuration; OCR extraction remains deferred. |
| S3 runtime | partial | S3-compatible adapter exists; prod needs endpoint, bucket, credentials, region/path-style choice, backup/retention policy, and smoke proof. |
| Domain/runtime config | partial | Replace local `127.0.0.1` URLs with prod `PUBLIC_WEB_URL`, backend callback URL, CORS origin, database URL, and object-store endpoint. |

## Hardcode Audit Notes

- Local URLs in `.env.example` are development defaults, not prod values.
- `OAUTH_FIXTURE_MODE=true` remains test-only and is already rejected in production config.
- `fixture` strings remain in tests and fixture login path only.
- Receipt upload and transaction attachment are implemented; OCR extraction remains deferred to the ingestion phase.
