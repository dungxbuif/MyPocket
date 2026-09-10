# Technology Stack

**Analysis Date:** 2026-09-10

## Languages

**Primary:**
- Go 1.23 module language - HTTP API, background worker, database migrations, domain logic, and platform adapters under `backend/`; declared by `backend/go.mod`.
- TypeScript 5.9 - React PWA, offline data layer, component tests, and browser-test configuration under `frontend/src/`, `frontend/e2e/`, and `frontend/vite.config.ts`.

**Secondary:**
- SQL (PostgreSQL dialect) - ordered schema migrations in `backend/migrations/`.
- JavaScript - service worker and small browser runtime assets such as `frontend/public/sw.js`.
- Shell - local platform smoke, E2E orchestration, and release verification in `scripts/smoke-platform.sh`, `scripts/run-e2e.sh`, and `scripts/verify-beta.sh`.
- MDX/Markdown - Docusaurus public documentation in `frontend/docs/docs/` and engineering documentation in `docs/`.

## Runtime

**Environment:**
- Go toolchain: module targets Go 1.23 in `backend/go.mod`; production builds currently use `golang:1.24-alpine` in `backend/Dockerfile`.
- Node.js 22 Alpine builds the PWA and public docs in `frontend/Dockerfile` and `backend/Dockerfile`; the docs package accepts Node.js 20+ in `frontend/docs/package.json`.
- Nginx 1.27 Alpine serves the compiled PWA and proxies `/api/` and `/docs/` in `frontend/Dockerfile` and `frontend/nginx.conf`.
- Browser target: modern evergreen browsers with an ES2022 TypeScript target in `frontend/tsconfig.json`; the installable/offline shell is implemented by `frontend/public/manifest.webmanifest` and `frontend/public/sw.js`.

**Package Manager:**
- Go modules - backend dependency resolution via `backend/go.mod` and `backend/go.sum`.
- npm - frontend and docs dependency resolution via `frontend/package.json` and `frontend/docs/package.json`.
- Lockfiles: present at `backend/go.sum`, `frontend/package-lock.json`, and `frontend/docs/package-lock.json`.

## Frameworks

**Core:**
- Go standard library `net/http` - API server and routing in `backend/cmd/api/main.go` and `backend/internal/platform/httpapi/router.go`; no third-party HTTP framework is used.
- React 19.1 - single-page PWA UI rooted in `frontend/src/app/App.tsx`.
- Vite 7.1 - frontend development server, API/docs proxy, production bundling, and test integration in `frontend/vite.config.ts`.
- Tailwind CSS 4.3 with `@tailwindcss/vite` - utility styling pipeline configured in `frontend/vite.config.ts`, alongside application styles in `frontend/src/styles.css`.
- Docusaurus 3.10.2 - compiled public/user documentation under `frontend/docs/`, embedded into the API container by `backend/Dockerfile`.

**Testing:**
- Go `testing` - unit, repository, HTTP contract, PostgreSQL integration, Redis, and S3 smoke tests colocated under `backend/internal/`.
- Vitest 3.2 with jsdom 25, Testing Library React 16.1, jest-dom 6.6, user-event 14.5, and fake-indexeddb 6.2 - frontend unit/component/offline tests configured by `frontend/vite.config.ts` and `frontend/src/test/setup.ts`.
- Playwright 1.62.1 - desktop, mobile, and WebKit browser workflows in `frontend/e2e/` with configuration in `frontend/playwright.config.ts` and `frontend/playwright.r0.config.ts`.

**Build/Dev:**
- TypeScript compiler 5.9 - strict, no-emit frontend type checking through `npm run typecheck` and `frontend/tsconfig.json`.
- Docker multi-stage builds - static web image in `frontend/Dockerfile`; API, worker, migrate binaries and Docusaurus site in `backend/Dockerfile`.
- Docker Compose - local/homelab topology defined by `compose.yaml` with `postgres`, `s3`, `redis`, `migrate`, `api`, `worker`, and `web` services.
- Nginx 1.27 - SPA fallback and same-origin proxying configured in `frontend/nginx.conf`.

## Key Dependencies

**Critical:**
- `github.com/jackc/pgx/v5` 5.7.2 - PostgreSQL driver and pool used by `backend/internal/platform/db/` and all persistent repositories.
- `github.com/aws/aws-sdk-go-v2` 1.36.2 plus `service/s3` 1.76.0 - private S3-compatible object access and presigned upload/download URLs in `backend/internal/platform/objectstore/s3.go`.
- `react` / `react-dom` 19.1 - browser application rendering from `frontend/src/`.
- `vite` 7.1 and `@vitejs/plugin-react` 5.0 - frontend development and production build pipeline in `frontend/vite.config.ts`.
- `typescript` 5.9 - compile-time contract enforcement across `frontend/src/`.

**Infrastructure:**
- PostgreSQL 16 Alpine - authoritative data store and worker coordination service declared in `compose.yaml`; migrations live in `backend/migrations/`.
- Redis 7 Alpine - optional development cache and required production cache for API-key/session identity, declared in `compose.yaml` and implemented without a third-party client in `backend/internal/platform/authcache/cache.go`.
- LocalStack 3.8.1 - development S3-compatible service declared in `compose.yaml`; the application itself accepts any compatible endpoint through `backend/internal/platform/objectstore/s3.go`.
- `lucide-react` 0.468 - icon set used by UI modules under `frontend/src/`.
- `clsx` 2.1 and `tailwind-merge` 3.6 - class composition helpers used by shared UI components under `frontend/src/components/`.

## Configuration

**Environment:**
- Backend configuration is loaded and validated centrally by `backend/internal/platform/config/config.go`; required base variables are `DATABASE_URL`, `PUBLIC_WEB_URL`, `COOKIE_SECRET`, and `CSRF_SECRET`.
- Production additionally requires `AUDIT_VIEWER_EMAIL`, `AUDIT_HASH_SECRET`, `API_KEY_HASH_SECRET`, `REDIS_URL`, HTTPS in `PUBLIC_WEB_URL`, and JSON logging; these checks are enforced in `backend/internal/platform/config/config.go`.
- Google OAuth uses `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, and `GOOGLE_REDIRECT_URL`; `OAUTH_FIXTURE_MODE` is development/test-only and rejected in production by `backend/internal/platform/config/config.go`.
- Optional S3 support is enabled by `S3_ENDPOINT` and then requires `S3_BUCKET`, `S3_ACCESS_KEY`, and `S3_SECRET_KEY`; the adapter fixes region to `us-east-1` and path-style addressing in `backend/internal/platform/objectstore/s3.go`.
- Other backend settings are `APP_ENV`, `ALLOWED_LOGIN_EMAILS`, `AUDIT_RETENTION_DAYS`, `LOG_FORMAT`, `LOG_LEVEL`, `HTTP_ADDR`, and `DOCS_DIR` in `backend/internal/platform/config/config.go`.
- Static portfolio price refresh uses optional `PORTFOLIO_STATIC_PRICES_JSON` in `backend/internal/worker/portfolio_prices.go`.
- Frontend API routing uses `VITE_API_BASE_URL` at build/runtime access in `frontend/src/app/apiClient.ts`; local proxying uses `API_PROXY_TARGET` in `frontend/vite.config.ts`.
- `.env` and `.env.example` are present at repository root for environment configuration; their contents are intentionally not part of this map.

**Build:**
- Frontend configuration: `frontend/vite.config.ts`, `frontend/tsconfig.json`, `frontend/package.json`, `frontend/playwright.config.ts`, and `frontend/playwright.r0.config.ts`.
- Documentation configuration: `frontend/docs/docusaurus.config.ts`, `frontend/docs/sidebars.ts`, and `frontend/docs/package.json`.
- Container/runtime configuration: `compose.yaml`, `backend/Dockerfile`, `backend/Dockerfile.docs-release`, `frontend/Dockerfile`, and `frontend/nginx.conf`.

## Platform Requirements

**Development:**
- Use Go compatible with the 1.23 module directive in `backend/go.mod`, Node.js 22/npm for parity with `frontend/Dockerfile`, Docker Compose for service dependencies in `compose.yaml`, and modern browsers for the PWA.
- Run PostgreSQL-sensitive backend suites against a dedicated test database with serial package execution as prescribed by `.agents/skills/auditing-mypocket-business-logic/SKILL.md`.
- Use `scripts/smoke-platform.sh` for PostgreSQL/S3 platform health, `scripts/run-e2e.sh` for browser orchestration, and `scripts/verify-beta.sh` for release-level verification.

**Production:**
- Homelab container deployment: Nginx-served web, Go API, Go worker, one-shot migrator, PostgreSQL, Redis, and an S3-compatible object store are represented in `compose.yaml`.
- Published API and web images are tracked in `docs/releases/CHANGELOG.md`; TLS/edge routing is operational infrastructure outside the application source, while `frontend/nginx.conf` handles internal same-origin proxying.
- The worker is a single minute-ticker process for recurring drafts, notifications, static portfolio price refresh, and audit retention in `backend/cmd/worker/main.go`.

---

*Stack analysis: 2026-09-10*
