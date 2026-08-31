---
artifact_type: detail_design
id: PHASE-007-PRODUCTION-HARDENING-DEBUG-AUDIT
status: approved
owner: shared
approval: approved
approved_on: 2026-08-31
trace:
  backlog_item: BL-009
  requirement: [REQ-F-015, REQ-NF-004, REQ-NF-007]
  phase: PHASE-007
  ticket_or_bug: TICKET-028
  test_verification: ../test-verification/TICKET-028-production-hardening-debug-audit-foundation.md
  validation_matrix: ../VALIDATION_MATRIX.md
  docs_review: ../DOCS-REVIEW-TICKET-028.md
  adrs: [../../decisions/ADR-005-restricted-audit-log.md]
  master_docs_touched:
    - ../../architecture/API.md
    - ../../architecture/ERD.md
    - ../../architecture/ARCHITECTURE.md
    - ../../operations/DEPLOYMENT.md
---

# DETAIL DESIGN: Production Hardening and Debug Audit Foundation

## Problem

UAT/prod bugs need a reliable path from a user-visible `correlation_id` to API logs and domain/security audit events. Current code has correlation IDs in responses but lacks structured request logs, panic recovery, persistent audit records, hidden audit inspection, and retention.

## Context Loaded

- `docs/CONTEXT.md`
- `docs/work/BACKLOG.md`
- `docs/standards/README.md`
- `docs/standards/QUALITY_BAR.md`
- `docs/standards/VALIDATION.md`
- `docs/standards/DEBUGGING.md`
- `docs/work/phases/PHASE-007-detail-design.md`
- `docs/work/tickets/TICKET-022-audit-pipeline-hidden-viewer.md`
- `docs/decisions/ADR-005-restricted-audit-log.md`
- `backend/internal/platform/httpapi/router.go`
- `backend/internal/platform/httpapi/auth.go`

## Brownfield Scope

- Backend: new `backend/internal/audit`, migration `0009`, API key migration `0010`, HTTP middleware/router/auth/sync instrumentation, worker instrumentation, Redis auth cache, config validation.
- Frontend: Account tab contains a viewer-only log panel. It first calls the backend access-check API; no audit controls or event data render for unauthorized users.
- Runtime: `.env.example`, Compose env, production docs.
- Media: private S3-compatible presigned receipt upload/download with user-scoped metadata and IndexedDB retry queue for offline-selected receipts.
- Master docs: API, ERD, architecture, validation matrix, changelog, context.

## Proposed Approach

Implement two complementary layers:

1. Operational structured logs go to stdout/stderr. They are JSON in production and contain correlation ID, method, path, status, duration, safe actor ID hash, and panic/error outcome.
2. Persistent audit events go to PostgreSQL. They are append-only through the application API and capture security/state actions with redacted metadata. They do not store raw request bodies, cookies, tokens, image bytes, provider secrets, or free-form transaction notes.

Audit writes use a best-effort writer from HTTP middleware for generic mutating routes and explicit writers in auth/sync/worker flows for high-value events. Failure to write an audit event is logged but does not fail the user mutation in this foundation slice; domain atomic audit can be tightened in a later deeper phase.

Add a basic third-party authentication layer for AI agents:

1. Users manage API keys from authenticated browser sessions under `/api/v1/api-keys` and the Account tab.
2. Key material uses the `mpk_` prefix, is generated server-side, and is returned only once.
3. PostgreSQL stores only HMAC-SHA256 hashes with `API_KEY_HASH_SECRET`, plus non-secret prefixes for display.
4. API requests with `Authorization: Bearer mpk_...` resolve to the owning user and use the same user-scoped repositories as web/PWA calls.
5. API-key-authenticated calls can perform business API mutations without CSRF because they are not browser-cookie requests.
6. API key management remains browser cookie + CSRF only, so a leaked API key cannot mint more keys.
7. Redis caches `api_key_hmac_hash -> user_id` and signed-cookie token digests for short TTLs; PostgreSQL remains the source of truth. Create syncs the new key hash into Redis; revoke removes that hash from Redis immediately.

## Audit Event Shape

```text
audit_events
- id uuid primary key
- occurred_at timestamptz
- correlation_id text
- actor_user_id uuid null
- actor_email_hash text
- action text
- entity_type text
- entity_id text
- outcome text: success | failure | denied | conflict | replayed
- severity text: info | warn | error | security
- source text: web | pwa | api | worker | sync | auth | system
- error_code text
- request_method text
- request_path text
- ip_hash text
- user_agent_hash text
- metadata_json jsonb
- created_at timestamptz
```

Indexes cover `occurred_at`, `actor_user_id`, `correlation_id`, `action`, and `severity/outcome`.

## Security

- Viewer APIs are `/api/v1/audit/access` and `/api/v1/audit/events`; both are backend-authorized and the event panel is only rendered after the access check succeeds.
- API key management API is `/api/v1/api-keys`; it is intended for owner use and not linked as a public admin surface in this foundation slice.
- Authorization requires normal app auth plus `user.email_verified = true` and normalized email exactly equals `AUDIT_VIEWER_EMAIL`.
- `AUDIT_VIEWER_EMAIL` is required in production.
- `AUDIT_HASH_SECRET` is required in production and used to HMAC email/IP/user-agent.
- `API_KEY_HASH_SECRET` and `REDIS_URL` are required in production.
- Structured logs never include cookies, auth headers, OAuth codes, provider secrets, request bodies, or raw SQL errors.

## Debug Workflow

1. User sees an error and reports `correlation_id`.
2. Operator filters `/api/v1/audit/events?correlation_id=req_...`.
3. Operator checks structured API/worker logs with the same correlation ID.
4. Audit event shows action/outcome/entity path; logs show status/duration/panic/provider failure.
5. Fix work starts from the correlated event set, not guesses.

## Alternatives

| Alternative | Decision |
| --- | --- |
| Log all request bodies | Rejected: privacy and secret exposure risk. |
| Audit only domain repositories transactionally in this slice | Deferred: stronger but touches every repository transaction boundary; foundation starts with HTTP/sync/worker coverage. |
| Hidden route without backend email authorization | Rejected: route obscurity is not security. |
| External SaaS tracing first | Rejected: homelab/UAT needs local proof and no vendor lock-in. |

## Verification Plan

- `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/audit ./internal/platform/config ./internal/platform/httpapi ./internal/sync ./internal/worker -count=1`
- `rtk env 'MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' GOCACHE=/private/tmp/mypocket-go-cache go test -p 1 ./internal/audit ./internal/platform/db ./internal/platform/httpapi -count=1`
- `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./... -count=1`
- `rtk npm test -- --run src/app frontend/src/offline`
- `rtk npm run build`

## Reconciliation Plan

- Add audit API and schema to architecture docs.
- Add prod env variables to deployment docs and env examples.
- Update backlog, validation matrix, context, docs review, and changelog.
