---
artifact_type: ticket
id: TICKET-028
status: in_review
owner: human
priority: urgent
lane: high-risk
trace:
  backlog_item: BL-009
  requirement: [REQ-F-015, REQ-NF-004, REQ-NF-007]
  phase: PHASE-007
  detail_design: ../phases/PHASE-007-production-hardening-debug-audit-detail-design.md
  test_verification: ../test-verification/TICKET-028-production-hardening-debug-audit-foundation.md
  validation_matrix: ../VALIDATION_MATRIX.md
  docs_review: ../DOCS-REVIEW-TICKET-028.md
  adrs: [../../decisions/ADR-005-restricted-audit-log.md]
  release_notes: ../../releases/CHANGELOG.md
---

# Ticket: TICKET-028 Production Hardening and Debug Audit Foundation

## Status

- Status: in_review
- Type: feature / production hardening
- Priority: urgent
- Phase: PHASE-007 production-hardening slice
- Approval: user requested plan and implementation without additional confirmation on 2026-08-31

## Context

The app is preparing for UAT/prod. Bugs in production will be difficult to trace unless each request, security outcome, state mutation, worker action, and safe error can be correlated across API responses, structured logs, and a restricted audit view.

## Scope

- Structured JSON request logs for API and worker with correlation IDs, method/path/status/duration, safe user identity, and stable error outcome.
- Panic recovery that returns safe JSON with correlation ID.
- Append-only PostgreSQL audit events for auth/security, state-changing API requests, sync mutations, and worker jobs.
- Hidden read-only audit API restricted by exact verified Google email in `AUDIT_VIEWER_EMAIL`.
- Retention purge with `AUDIT_RETENTION_DAYS`, default 180, in bounded worker batches.
- Basic user-managed API keys for third-party systems and AI agents, accepted as `Authorization: Bearer mpk_...`, with a compact Account-tab management UI.
- Redis-backed auth cache for API key and signed-cookie token lookups, with create/revoke cache sync and database fallback when cache misses.
- Production config validation for audit viewer, audit hash secret, log format/level, and client URL security.
- Documentation of prod-debug workflow and env variables.

## Out Of Scope

- General admin roles.
- API key scopes, expiration policies, rotation schedules, per-key rate limits, and organization/team ownership.
- Logging every UI click, page view, filter, or raw request body.
- Export, account reset/delete, receipt OCR/upload implementation, and external log aggregation.
- Vendor-specific tracing SaaS.

## Acceptance Criteria

- [x] Every API response keeps or generates `X-Correlation-ID`; errors include `correlation_id`.
- [x] API requests are logged as structured JSON in production without cookies, authorization headers, provider secrets, or request bodies.
- [x] Panic recovery emits a safe `INTERNAL_FAILURE` response and a structured error log with correlation ID.
- [x] Mutating API requests append redacted audit events with actor user ID when available, action, entity path, status outcome, source, request path, and correlation ID.
- [x] Google auth start/callback/logout/whitelist denied events are auditable.
- [x] Sync mutation batches append per-mutation audit events for applied/replayed/rejected/conflict outcomes.
- [x] Worker recurring/notification/portfolio refresh runs append safe audit events.
- [x] Authenticated users can create, list, and revoke API keys from API and Account-tab UI; plaintext key material is returned only once at creation.
- [x] Third-party callers can authenticate existing API routes with `Authorization: Bearer mpk_...` without browser cookies.
- [x] API key hashes and signed-cookie token lookups are cached in Redis and fall back to PostgreSQL on cache miss or cache outage.
- [x] API-key-authenticated requests bypass browser CSRF, while API key management itself remains browser cookie + CSRF only.
- [x] `/api/v1/audit/events` is absent from navigation and returns data only to the verified email configured by `AUDIT_VIEWER_EMAIL`.
- [x] Audit query supports bounded pagination and filters by correlation ID, action, severity, and date range.
- [x] Retention purge defaults to 180 days and deletes expired rows in bounded batches without recursive audit spam.
- [x] Tests prove redaction, hidden viewer authorization, audit writes, retention, panic recovery, API key auth, and config validation.

## Small Task Exemption

- Small task exemption: no
- Reason: Adds security-sensitive audit data, API, DB schema, runtime config, worker behavior, and production logging.
- Impact checked: API=yes, DB=yes, Security=yes, Runtime=yes, Standards=no

## Test Expectations

- Unit: redaction/hash behavior, config parsing/validation, action mapping.
- Integration: migration, append-only audit writes, viewer allow/deny, retention purge.
- HTTP: request logging/correlation, panic recovery safe response, mutating request audit event, API key lifecycle, bearer auth.
- Worker: safe audit for recurring/notification/portfolio refresh outcomes.
- Regression: full backend test suite, frontend build/test unchanged.

## Verification Results

- Command: `GOCACHE=/private/tmp/mypocket-go-cache go test ./... -count=1`
- Result: pass
- Command: `MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable GOCACHE=/private/tmp/mypocket-go-cache go test -p 1 ./internal/identity ./internal/audit ./internal/platform/db ./internal/platform/httpapi -count=1 -v`
- Result: pass
- Command: `npm test -- --run src/app frontend/src/offline`
- Result: pass
- Command: `npm run build`
- Result: pass
- Notes: Automated proof complete including Account-tab API key UI build; human UAT for production deployment remains.
