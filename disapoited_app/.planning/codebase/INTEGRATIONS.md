# External Integrations

**Analysis Date:** 2026-09-10

## APIs & External Services

**Identity:**
- Google OAuth 2.0 / OpenID Connect - browser sign-in and verified profile lookup.
  - SDK/Client: Go standard library HTTP/form clients in `backend/internal/platform/httpapi/auth.go`; no Google SDK is used.
  - Auth: `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `GOOGLE_REDIRECT_URL`.
  - Endpoints: Google authorization, token exchange, and OIDC userinfo URLs are called directly by `backend/internal/platform/httpapi/auth.go`.
  - Persistence: only Google subject/profile identity is stored by `backend/internal/identity/repository.go`; provider access tokens are not retained.

**Object Storage:**
- S3-compatible private object storage - receipt image upload/download, deletion, and platform smoke checks through presigned URLs.
  - SDK/Client: AWS SDK for Go v2 packages declared in `backend/go.mod` and adapter `backend/internal/platform/objectstore/s3.go`.
  - Auth: `S3_ACCESS_KEY`, `S3_SECRET_KEY`; endpoint/bucket are `S3_ENDPOINT` and `S3_BUCKET`.
  - Contract: path-style S3 requests, fixed `us-east-1` signing region, 15-minute upload URLs from `backend/internal/platform/httpapi/receipts.go`, and 10-minute download URLs from `backend/internal/platform/objectstore/s3.go`.

**Portfolio Pricing:**
- Static configured price provider - worker refreshes automatic asset prices from server-side JSON configuration.
  - SDK/Client: internal `portfolio.PriceProvider` boundary in `backend/internal/portfolio/` and `backend/internal/worker/portfolio_prices.go`; no live market API client is implemented.
  - Auth: none; optional data is supplied through `PORTFOLIO_STATIC_PRICES_JSON`.

**Notifications:**
- In-app notifications - durable notices and subscriptions are persisted through `backend/internal/notification/repository.go`.
  - SDK/Client: internal repository/processor in `backend/internal/notification/worker.go`.
  - Auth: application session or MyPocket API key at HTTP boundaries.
- Web Push delivery is not connected: the worker explicitly uses `notification.NoopDelivery` in `backend/cmd/worker/main.go`; subscription persistence exists, but no VAPID/provider SDK or credentials are active.

**Planned Provider Boundaries:**
- OpenAI-compatible text/multimodal APIs, external OCR, and HMAC bank-notification ingestion are documented provider classes in `docs/architecture/INTEGRATIONS.md`, but no active external client or incoming bank webhook route is registered in `backend/internal/platform/httpapi/router.go`.
- Treat these as planned integrations, not available runtime capabilities; confirmed accounting remains behind review/confirmation contracts described in `docs/architecture/API.md`.

## Data Storage

**Databases:**
- PostgreSQL 16 - authoritative identity, finance, planning, portfolio, sync, audit, and notification store; also supplies advisory/lease coordination for workers.
  - Connection: `DATABASE_URL`.
  - Client: `github.com/jackc/pgx/v5` 5.7.2 through `backend/internal/platform/db/`; repositories live under `backend/internal/` domain packages.
  - Schema: sequential SQL migrations in `backend/migrations/`, applied by `backend/cmd/migrate/main.go`.

**File Storage:**
- Private S3-compatible storage for receipt bytes through `backend/internal/platform/objectstore/s3.go`; PostgreSQL receipt metadata and object keys are handled by `backend/internal/finance/repository.go` and `backend/internal/platform/httpapi/receipts.go`.
- LocalStack S3 is the local development provider in `compose.yaml`; production may use another S3-compatible endpoint.
- Browser-local file/cache state uses IndexedDB through `frontend/src/offline/db.ts`; it is an offline mirror/outbox, not the source of truth.

**Caching:**
- Redis caches API-key/session identity lookups through the minimal RESP client in `backend/internal/platform/authcache/cache.go`.
  - Connection: `REDIS_URL`.
  - Client: custom standard-library TCP/RESP implementation; no Redis package is declared in `backend/go.mod`.
  - Behavior: optional outside production, required by production validation in `backend/internal/platform/config/config.go`; PostgreSQL remains authoritative and cache errors fall back through authentication logic in `backend/internal/platform/httpapi/auth_context.go`.
- Browser offline cache/outbox uses IndexedDB in `frontend/src/offline/` and service-worker caching in `frontend/public/sw.js`.

## Authentication & Identity

**Auth Provider:**
- Google OAuth 2.0 / OIDC for browser identity.
  - Implementation: authorization-code redirect, state cookie, direct token exchange, OIDC userinfo fetch, optional allowlist, and local signed application session in `backend/internal/platform/httpapi/auth.go`.
  - Session: signed `mypocket_auth` cookie from `backend/internal/identity/cookie.go`, companion CSRF token from `backend/internal/identity/csrf.go`, and 30-day cookie lifetime configured by `backend/internal/platform/httpapi/auth.go`.
  - Development fixture: `OAUTH_FIXTURE_MODE=true` bypasses Google only outside production; enforcement is in `backend/internal/platform/config/config.go`.
- MyPocket API keys provide third-party bearer access to user-owned business routes.
  - Implementation: key creation/hash/revocation in `backend/internal/identity/api_keys.go` and `backend/internal/identity/repository.go`; bearer resolution and Redis cache integration in `backend/internal/platform/httpapi/auth_context.go`.
  - Auth: `Authorization: Bearer mpk_...`; server-side hashing uses `API_KEY_HASH_SECRET`.
  - Constraint: API-key management remains browser-session plus CSRF protected; business calls authenticated by an API key do not use browser CSRF, as documented in `docs/architecture/API.md`.

## Monitoring & Observability

**Error Tracking:**
- None detected; no Sentry, hosted APM, OpenTelemetry exporter, or error-tracking SDK appears in `backend/go.mod` or `frontend/package.json`.

**Logs:**
- Go standard `log/slog` writes text or JSON to stdout with level selection in `backend/internal/platform/logging/logging.go`; production requires JSON via `backend/internal/platform/config/config.go`.
- HTTP middleware supplies correlation IDs and records safe audit events through `backend/internal/platform/httpapi/router.go` and `backend/internal/audit/`.
- Security/audit records are stored in PostgreSQL and restricted by `AUDIT_VIEWER_EMAIL`; hashing/redaction uses `AUDIT_HASH_SECRET` in `backend/internal/audit/`.
- Health/readiness endpoints are registered by `backend/internal/platform/httpapi/router.go`; readiness checks PostgreSQL through `backend/cmd/api/main.go`.

## CI/CD & Deployment

**Hosting:**
- Docker Compose homelab deployment defined in `compose.yaml`: PostgreSQL 16, Redis 7, S3-compatible LocalStack for the bundled environment, migrator, Go API, Go worker, and Nginx web frontend.
- Container definitions are `backend/Dockerfile`, `backend/Dockerfile.docs-release`, and `frontend/Dockerfile`; internal reverse proxy rules are in `frontend/nginx.conf`.
- Image publication/release state is recorded in `docs/releases/CHANGELOG.md`; edge TLS/registry infrastructure is operationally referenced there but not implemented in this repository.

**CI Pipeline:**
- No checked-in hosted CI workflow is detected; there is no `.github/workflows/` pipeline.
- Verification and release gates are local shell entry points: `scripts/smoke-platform.sh`, `scripts/run-e2e.sh`, and `scripts/verify-beta.sh`.

## Environment Configuration

**Required env vars:**
- Base API/worker: `DATABASE_URL`, `PUBLIC_WEB_URL`, `COOKIE_SECRET`, `CSRF_SECRET` (`backend/internal/platform/config/config.go`).
- Production-only: `AUDIT_VIEWER_EMAIL`, `AUDIT_HASH_SECRET`, `API_KEY_HASH_SECRET`, `REDIS_URL`, `LOG_FORMAT=json` (`backend/internal/platform/config/config.go`).
- Google OAuth when enabled: `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `GOOGLE_REDIRECT_URL`; optional access restriction: `ALLOWED_LOGIN_EMAILS` (`backend/internal/platform/config/config.go`).
- S3 when enabled: `S3_ENDPOINT`, `S3_BUCKET`, `S3_ACCESS_KEY`, `S3_SECRET_KEY` (`backend/internal/platform/config/config.go`).
- Optional/runtime tuning: `APP_ENV`, `OAUTH_FIXTURE_MODE`, `AUDIT_RETENTION_DAYS`, `LOG_LEVEL`, `HTTP_ADDR`, `DOCS_DIR`, `PORTFOLIO_STATIC_PRICES_JSON` (`backend/internal/platform/config/config.go`, `backend/internal/worker/portfolio_prices.go`).
- Frontend: optional `VITE_API_BASE_URL` in `frontend/src/app/apiClient.ts`; local Vite proxy target `API_PROXY_TARGET` in `frontend/vite.config.ts`.
- Test-only integration variables: `MYPOCKET_TEST_DATABASE_URL` in backend repository tests, `MYPOCKET_TEST_REDIS_URL` in `backend/internal/platform/authcache/cache_test.go`, and `MYPOCKET_TEST_S3_ENDPOINT`, `MYPOCKET_TEST_S3_BUCKET`, `MYPOCKET_TEST_S3_ACCESS_KEY`, `MYPOCKET_TEST_S3_SECRET_KEY` in `backend/internal/platform/objectstore/objectstore_test.go`.

**Secrets location:**
- `.env` is present at repository root and contains environment configuration; its contents were not read. `.env.example` is also present and was not read because environment-file contents are excluded from codebase mapping.
- Application code reads secrets only through environment names in `backend/internal/platform/config/config.go`; production secret material must remain outside version-controlled source.
- API keys are stored as hashes by `backend/internal/identity/repository.go`; Google provider tokens are not stored after profile exchange in `backend/internal/platform/httpapi/auth.go`.

## Webhooks & Callbacks

**Incoming:**
- `GET /api/v1/auth/google/callback` receives the Google OAuth authorization-code callback through routing in `backend/internal/platform/httpapi/router.go` and processing in `backend/internal/platform/httpapi/auth.go`.
- A bank-notification route `POST /webhooks/bank/{source_id}` is specified as planned in `docs/architecture/API.md`, but no handler is registered in `backend/internal/platform/httpapi/router.go`; do not build callers against it yet.

**Outgoing:**
- Google OAuth token exchange and OpenID Connect userinfo requests are issued by `backend/internal/platform/httpapi/auth.go`.
- S3 API calls and presigned browser upload/download URLs are produced by `backend/internal/platform/objectstore/s3.go` and exposed through `backend/internal/platform/httpapi/receipts.go`.
- No real Web Push delivery is emitted because `backend/cmd/worker/main.go` wires `notification.NoopDelivery`.
- No live portfolio market-data request is emitted; `backend/internal/worker/portfolio_prices.go` uses locally configured static prices.

---

*Integration audit: 2026-09-10*
