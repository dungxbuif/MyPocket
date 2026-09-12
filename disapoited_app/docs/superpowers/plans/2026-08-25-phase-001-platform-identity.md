# PHASE-001 Platform and Identity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the first deployable MyPocket slice: mobile-first installable React PWA shell, Go API, Go worker, PostgreSQL migration baseline, S3-compatible adapter, Google OAuth fixture flow, stateless cookie auth, CSRF, ownership isolation proof, and local operations commands.

**Architecture:** The browser is a Vite React PWA that owns presentation, service-worker app-shell caching, auth state, and offline-status UI. The backend is a Go modular monolith exposed as separate API and worker processes sharing `internal/platform` and `internal/identity`. PostgreSQL is authoritative for users and migration state; S3-compatible storage is accessed through a platform adapter.

**Tech Stack:** React, TypeScript, Vite, React Router, TanStack Query, PWA service worker, Go HTTP, PostgreSQL, S3-compatible object storage, Docker Compose or equivalent local services.

**Spec:** `docs/work/phases/PHASE-001-detail-design.md`

## Global Constraints

- Every shell command in this repository must be prefixed with `rtk`.
- No product code execution starts until `docs/work/phases/PHASE-001-detail-design.md` approval is `approved`.
- Use React + TypeScript with Vite for the PWA.
- Use Go for API and worker processes.
- Use PostgreSQL as source of truth and S3-compatible private object storage.
- Use Google OAuth with a signed stateless cookie; do not create a session table and do not persist Google access or refresh tokens.
- Full offline read/write, outbox replay, tombstones, and conflict resolution remain out of scope for PHASE-001.
- PHASE-001 must still support offline app-shell reload after first online load.
- Use stable JSON API error codes and correlation IDs without exposing provider errors, stack traces, or secrets.
- Keep all user-owned backend access scoped by authenticated `user_id`.
- Update validation, context, backlog, release notes, and docs review only from real execution evidence.

---

## File Structure

- Create: `package.json` for root scripts that call web and Go commands.
- Create: `apps/web/` for Vite React PWA source, mobile shell, service worker, tests, and static manifest.
- Create: `apps/api/` for Go API entrypoint and route registration.
- Create: `apps/worker/` for Go worker entrypoint.
- Create: `internal/platform/config/` for shared environment parsing and safe config errors.
- Create: `internal/platform/httpapi/` for router, error envelopes, correlation IDs, health routes, and CSRF middleware.
- Create: `internal/platform/db/` for PostgreSQL pool setup and migration runner.
- Create: `internal/platform/objectstore/` for S3-compatible adapter interface and smoke operation.
- Create: `internal/identity/` for user model, repository, OAuth fixture/provider boundary, signed cookies, CSRF, and ownership guard.
- Create: `migrations/` for ordered PHASE-001 SQL.
- Create: `deployments/` or root `compose.yaml` for local PostgreSQL/S3/API/worker/web startup.
- Modify: `docs/work/tickets/TICKET-001-repository-runtime-foundation.md` with real verification evidence during execution.
- Modify: `docs/work/tickets/TICKET-002-postgresql-migrations-s3-platform-adapters.md` with real verification evidence during execution.
- Modify: `docs/work/tickets/TICKET-003-google-oauth-user-isolation.md` with real verification evidence during execution.
- Modify: `docs/work/tickets/TICKET-004-development-operations-verification-baseline.md` with real verification evidence during execution.
- Modify: `docs/work/VALIDATION_MATRIX.md`, `docs/CONTEXT.md`, `docs/work/BACKLOG.md`, `docs/releases/CHANGELOG.md`, and affected master docs during reconciliation/dehydration.

---

### Task 1: Repository Tooling and API Error Foundation

**Files:**
- Create: `go.mod`
- Create: `apps/api/main.go`
- Create: `apps/worker/main.go`
- Create: `internal/platform/config/config.go`
- Create: `internal/platform/config/config_test.go`
- Create: `internal/platform/httpapi/errors.go`
- Create: `internal/platform/httpapi/errors_test.go`
- Create: `internal/platform/httpapi/router.go`
- Create: `internal/platform/httpapi/health.go`
- Create: `internal/platform/httpapi/health_test.go`
- Create: `package.json`

**Interfaces:**
- Produces: `config.Load(env map[string]string) (config.Config, error)`
- Produces: `httpapi.ErrorEnvelope(code string, message string, correlationID string) map[string]any`
- Produces: `httpapi.NewRouter(cfg config.Config, deps httpapi.Dependencies) http.Handler`

- [x] **Step 1: Write config tests**

```go
func TestLoadRequiresDatabaseURL(t *testing.T) {
	_, err := config.Load(map[string]string{"APP_ENV": "test"})
	if err == nil || !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Fatalf("expected missing DATABASE_URL error, got %v", err)
	}
}

func TestLoadRedactsSecretValues(t *testing.T) {
	_, err := config.Load(map[string]string{"DATABASE_URL": "postgres://user:secret@localhost/db"})
	if err != nil && strings.Contains(err.Error(), "secret") {
		t.Fatalf("config error leaked secret: %v", err)
	}
}
```

- [x] **Step 2: Run config tests to verify they fail**

Run: `rtk go test ./internal/platform/config`

Expected: FAIL because the package does not exist.

- [x] **Step 3: Implement `internal/platform/config`**

Create `Config` with `AppEnv`, `DatabaseURL`, `PublicWebURL`, `CookieSecret`, `CSRFSecret`, `S3Endpoint`, `S3Bucket`, `S3AccessKey`, `S3SecretKey`, and `OAuthFixtureMode`. `Load` must return named missing-variable errors and must not include secret values in error strings.

- [x] **Step 4: Run config tests to verify they pass**

Run: `rtk go test ./internal/platform/config`

Expected: PASS.

- [x] **Step 5: Write API error and health tests**

```go
func TestErrorEnvelopeIncludesStableCodeAndCorrelationID(t *testing.T) {
	body := httpapi.ErrorEnvelope("AUTH_REQUIRED", "Authentication required", "req_123")
	if body["correlation_id"] != "req_123" {
		t.Fatalf("missing correlation id: %#v", body)
	}
}

func TestLiveHealthReturnsOK(t *testing.T) {
	handler := httpapi.NewRouter(config.Config{AppEnv: "test"}, httpapi.Dependencies{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health/live", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}
}
```

- [x] **Step 6: Run API tests to verify they fail**

Run: `rtk go test ./internal/platform/httpapi`

Expected: FAIL because the package or functions do not exist.

- [x] **Step 7: Implement API router, error envelopes, correlation IDs, and health route**

Use `X-Correlation-ID` when present; otherwise generate a request ID. Every JSON response includes `correlation_id`. Error responses include `error.code` and a safe message.

- [x] **Step 8: Run API tests**

Run: `rtk go test ./internal/platform/config ./internal/platform/httpapi`

Expected: PASS.

- [x] **Step 9: Add minimal API and worker entrypoints**

`apps/api/main.go` loads config, creates the router, and starts HTTP. `apps/worker/main.go` loads config and logs startup without running jobs yet.

- [x] **Step 10: Commit Task 1**

```bash
rtk git add go.mod apps/api apps/worker internal/platform package.json
rtk git commit -m "feat: add runtime and API foundation"
```

---

### Task 2: PostgreSQL Migration Runner and User Schema

**Files:**
- Create: `migrations/0001_phase001_identity.sql`
- Create: `internal/platform/db/db.go`
- Create: `internal/platform/db/migrate.go`
- Create: `internal/platform/db/migrate_test.go`
- Modify: `internal/platform/httpapi/health.go`
- Modify: `internal/platform/httpapi/health_test.go`

**Interfaces:**
- Consumes: `config.Config.DatabaseURL`
- Produces: `db.Open(ctx context.Context, cfg config.Config) (*sql.DB, error)`
- Produces: `db.Migrate(ctx context.Context, conn *sql.DB, migrations fs.FS) error`
- Produces: readiness dependency hook used by `httpapi.NewRouter`

- [x] **Step 1: Write migration tests**

```go
func TestMigrateCreatesUsersWithoutProviderTokens(t *testing.T) {
	conn := openTestPostgres(t)
	err := db.Migrate(context.Background(), conn, os.DirFS("../../../migrations"))
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}
	columns := loadColumns(t, conn, "users")
	if columns["google_access_token"] || columns["google_refresh_token"] || columns["session_id"] {
		t.Fatalf("users table contains forbidden auth persistence columns: %#v", columns)
	}
}
```

- [x] **Step 2: Run migration tests to verify they fail**

Run: `rtk go test ./internal/platform/db`

Expected: FAIL because migration code does not exist or test PostgreSQL is not configured. If PostgreSQL is missing, record the environment failure in TICKET-002 before continuing with local service setup in Task 6.

- [x] **Step 3: Implement migration SQL**

Create `schema_migrations` and `users` with UUID primary key, unique `google_subject`, verified email fields, display fields, and timestamps. Do not add session or provider-token columns.

- [x] **Step 4: Implement migration runner**

Apply ordered `.sql` files once, store version/checksum, and fail on checksum mismatch.

- [x] **Step 5: Add readiness dependency**

Readiness returns OK only when database ping succeeds; liveness remains independent from database.

- [x] **Step 6: Run database and health tests**

Run: `rtk go test ./internal/platform/db ./internal/platform/httpapi`

Expected: PASS when local PostgreSQL test configuration is available.

- [x] **Step 7: Commit Task 2**

```bash
rtk git add migrations internal/platform/db internal/platform/httpapi
rtk git commit -m "feat: add PostgreSQL migrations"
```

---

### Task 3: S3-Compatible Object Store Adapter

**Files:**
- Create: `internal/platform/objectstore/objectstore.go`
- Create: `internal/platform/objectstore/s3.go`
- Create: `internal/platform/objectstore/objectstore_test.go`
- Modify: `internal/platform/config/config.go`
- Modify: `internal/platform/config/config_test.go`

**Interfaces:**
- Produces: `objectstore.Store` with `PutSmokeObject(ctx context.Context) error` and `DeleteSmokeObject(ctx context.Context) error`
- Consumes: S3 endpoint, bucket, access key, and secret from `config.Config`

- [x] **Step 1: Write config and adapter tests**

```go
func TestS3ConfigRequiresBucketWhenEndpointConfigured(t *testing.T) {
	_, err := config.Load(map[string]string{
		"DATABASE_URL": "postgres://localhost/db",
		"S3_ENDPOINT":  "http://localhost:9000",
	})
	if err == nil || !strings.Contains(err.Error(), "S3_BUCKET") {
		t.Fatalf("expected S3_BUCKET error, got %v", err)
	}
}
```

- [x] **Step 2: Run tests to verify they fail**

Run: `rtk go test ./internal/platform/config ./internal/platform/objectstore`

Expected: FAIL because object store package or config validation is incomplete.

- [x] **Step 3: Implement object store interface and S3 adapter**

Use the selected Go S3-compatible SDK. The adapter must never log access keys, secret keys, presigned URLs, or object payloads.

- [x] **Step 4: Add smoke test path**

Add a test that runs only when local S3 environment variables are present. It should create and delete a small private smoke object.

- [x] **Step 5: Run adapter tests**

Run: `rtk go test ./internal/platform/config ./internal/platform/objectstore`

Expected: PASS, with smoke test skipped when S3 env is absent and passing when local S3 is available.

- [x] **Step 6: Commit Task 3**

```bash
rtk git add internal/platform/config internal/platform/objectstore
rtk git commit -m "feat: add object store adapter"
```

---

### Task 4: Identity, OAuth Fixture, Cookie, CSRF, and Ownership Guard

**Files:**
- Create: `internal/identity/user.go`
- Create: `internal/identity/repository.go`
- Create: `internal/identity/oauth.go`
- Create: `internal/identity/cookie.go`
- Create: `internal/identity/csrf.go`
- Create: `internal/identity/ownership.go`
- Create: `internal/identity/identity_test.go`
- Modify: `internal/platform/httpapi/router.go`
- Modify: `internal/platform/httpapi/errors.go`

**Interfaces:**
- Produces: `identity.User`
- Produces: `identity.Repository.FindOrCreateGoogleUser(ctx, profile) (User, error)`
- Produces: `identity.SignCookie(userID string) (string, error)`
- Produces: `identity.VerifyCookie(value string) (Claims, error)`
- Produces: `identity.RequireUser(next http.Handler) http.Handler`
- Produces: `identity.RequireCSRF(next http.Handler) http.Handler`
- Produces: `identity.RequireOwner(ctx, userID, objectID string) error`

- [x] **Step 1: Write identity unit tests**

```go
func TestSignedCookieRoundTrip(t *testing.T) {
	signer := identity.NewCookieSigner([]byte("01234567890123456789012345678901"))
	value, err := signer.Sign("user_123", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	claims, err := signer.Verify(value)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserID != "user_123" {
		t.Fatalf("wrong user id: %s", claims.UserID)
	}
}

func TestCSRFMissingHeaderRejected(t *testing.T) {
	handler := identity.RequireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil))
	if res.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", res.Code)
	}
}
```

- [x] **Step 2: Run identity tests to verify they fail**

Run: `rtk go test ./internal/identity`

Expected: FAIL because identity package does not exist.

- [x] **Step 3: Implement user model, repository, cookie signer, and CSRF guard**

Cookie claims include user ID, expiry, and issued-at. CSRF uses a signed token exposed through a non-HttpOnly CSRF cookie or equivalent double-submit contract and required header for mutations.

- [x] **Step 4: Write OAuth fixture and repository integration tests**

```go
func TestOAuthFixtureProvisionsUserWithoutProviderTokens(t *testing.T) {
	conn := migratedTestPostgres(t)
	repo := identity.NewRepository(conn)
	profile := identity.GoogleProfile{Subject: "google-sub-1", Email: "a@example.com", EmailVerified: true, DisplayName: "A"}
	user, err := repo.FindOrCreateGoogleUser(context.Background(), profile)
	if err != nil {
		t.Fatal(err)
	}
	if user.Email != "a@example.com" {
		t.Fatalf("wrong email: %s", user.Email)
	}
	assertNoProviderTokenColumns(t, conn)
}
```

- [x] **Step 5: Implement auth routes**

Add `GET /api/v1/auth/google`, `GET /api/v1/auth/google/callback`, `POST /api/v1/auth/logout`, and `GET /api/v1/me`. Fixture mode must be explicit through config and unavailable by accident in production mode.

- [x] **Step 6: Write ownership guard integration test**

Create two users and one user-owned harness object. Verify user B cannot access user A's object and receives a stable `FORBIDDEN` or scoped `NOT_FOUND` error.

- [x] **Step 7: Run identity and API tests**

Run: `rtk go test ./internal/identity ./internal/platform/httpapi`

Expected: PASS.

- [x] **Step 8: Commit Task 4**

```bash
rtk git add internal/identity internal/platform/httpapi
rtk git commit -m "feat: add Google OAuth identity"
```

---

### Task 5: Mobile-First React PWA Shell and Offline App-Shell Cache

**Files:**
- Create: `apps/web/package.json`
- Create: `apps/web/index.html`
- Create: `apps/web/vite.config.ts`
- Create: `apps/web/src/main.tsx`
- Create: `apps/web/src/app/App.tsx`
- Create: `apps/web/src/app/apiClient.ts`
- Create: `apps/web/src/app/auth.ts`
- Create: `apps/web/src/app/offline.ts`
- Create: `apps/web/src/app/serviceWorker.ts`
- Create: `apps/web/src/app/App.test.tsx`
- Create: `apps/web/src/styles.css`
- Create: `apps/web/public/manifest.webmanifest`
- Create: `apps/web/public/pwa-icon.svg`
- Create: `apps/web/public/sw.js`

**Interfaces:**
- Consumes: `GET /api/v1/me`, auth redirect routes, and logout route.
- Produces: mobile shell route structure, auth state, offline status, service worker registration, cached app shell.

- [x] **Step 1: Write React shell tests**

```tsx
it("renders mobile navigation destinations without finance data", async () => {
  render(<App />);
  expect(screen.getByRole("navigation")).toBeInTheDocument();
  expect(screen.getByLabelText("Overview")).toBeInTheDocument();
  expect(screen.getByLabelText("Transactions")).toBeInTheDocument();
  expect(screen.getByLabelText("Quick add")).toBeInTheDocument();
  expect(screen.getByLabelText("Wallets")).toBeInTheDocument();
  expect(screen.getByLabelText("Account")).toBeInTheDocument();
});

it("shows offline state when the browser is offline", () => {
  mockNavigatorOnline(false);
  render(<App />);
  expect(screen.getByText("Offline")).toBeInTheDocument();
});
```

- [x] **Step 2: Run web tests to verify they fail**

Run: `rtk npm test --workspace apps/web -- --run`

Expected: FAIL because the web app does not exist yet.

- [x] **Step 3: Implement Vite React app**

Create a real app shell as the first screen. Use compact mobile-first layout, fixed bottom navigation on small screens, a wider sidebar/top layout on desktop, Vietnamese-facing copy where visible, VND-safe placeholder formatting, and no marketing page.

- [x] **Step 4: Implement API client and auth state**

`apiClient` sends credentials, propagates `X-CSRF-Token` for mutations, and normalizes stable error codes. `auth.ts` loads `/api/v1/me`, exposes unauthenticated/authenticated/loading states, and supports logout.

- [x] **Step 5: Implement service worker**

Cache the app shell assets, `manifest.webmanifest`, and PWA icons. Do not cache authenticated API JSON as authoritative offline finance data in PHASE-001. Show offline state when network is unavailable.

- [x] **Step 6: Run web unit tests**

Run: `rtk npm test --workspace apps/web -- --run`

Expected: PASS.

- [x] **Step 7: Build the web app**

Run: `rtk npm run build --workspace apps/web`

Expected: PASS and generated assets include manifest and service worker.

- [x] **Step 8: Commit Task 5**

```bash
rtk git add apps/web package.json
rtk git commit -m "feat: add mobile PWA shell"
```

---

### Task 6: Local Operations, Compose, and Platform Smoke

**Files:**
- Create: `compose.yaml`
- Create: `.env.example`
- Create: `docs/engineering/LOCAL_DEVELOPMENT.md`
- Create: `docs/work/test-verification/PHASE-001-platform-smoke.md`
- Modify: `package.json`
- Modify: `apps/api/main.go`
- Modify: `apps/worker/main.go`

**Interfaces:**
- Consumes: API, worker, database, S3 adapter, web app from earlier tasks.
- Produces: documented startup and smoke commands for local development.

- [ ] **Step 1: Add compose services**

Define local PostgreSQL, S3-compatible service, API, worker, and web services or documented equivalent commands. Do not commit real secrets.

- [ ] **Step 2: Add root scripts**

Add scripts for `dev:web`, `dev:api`, `dev:worker`, `test:go`, `test:web`, `build:web`, and `smoke:platform`.

- [ ] **Step 3: Write platform verification document skeleton**

Record commands that will be run: dependency install, Go tests, web tests, web build, compose startup, health checks, S3 smoke, PWA/mobile/offline E2E.

- [ ] **Step 4: Run backend tests**

Run: `rtk go test ./...`

Expected: PASS.

- [ ] **Step 5: Run frontend tests and build**

Run: `rtk npm test --workspace apps/web -- --run`

Expected: PASS.

Run: `rtk npm run build --workspace apps/web`

Expected: PASS.

- [ ] **Step 6: Run local platform smoke**

Run: `rtk npm run smoke:platform`

Expected: PASS; health endpoints return stable JSON and local S3 smoke succeeds. If local Docker or S3 is unavailable, record `skipped` with residual risk in TICKET-004.

- [ ] **Step 7: Commit Task 6**

```bash
rtk git add compose.yaml .env.example package.json docs/engineering/LOCAL_DEVELOPMENT.md docs/work/test-verification apps/api apps/worker
rtk git commit -m "chore: add local operations baseline"
```

---

### Task 7: Browser E2E and PWA Offline Verification

**Files:**
- Create: `apps/web/e2e/pwa-shell.spec.ts`
- Create: `apps/web/playwright.config.ts`
- Modify: `apps/web/package.json`
- Modify: `docs/work/test-verification/PHASE-001-platform-smoke.md`

**Interfaces:**
- Consumes: running web and API test fixture.
- Produces: evidence for mobile shell, login fixture, refresh persistence, logout, forbidden state, service worker registration, and offline reload.

- [ ] **Step 1: Write Playwright tests**

Cover mobile viewport shell, service worker registration, deterministic login callback, `/me` authenticated state, logout, forbidden state, and offline reload after first load.

- [ ] **Step 2: Run E2E test to verify expected initial failures**

Run: `rtk npm run test:e2e --workspace apps/web`

Expected: FAIL until API fixture and dev server wiring are complete.

- [ ] **Step 3: Wire API fixture and dev server**

Start API in fixture mode with deterministic OAuth profile. Ensure Playwright uses a mobile viewport and toggles offline after the app shell is cached.

- [ ] **Step 4: Run E2E tests**

Run: `rtk npm run test:e2e --workspace apps/web`

Expected: PASS.

- [ ] **Step 5: Record PWA/manual evidence**

Record installability check, mobile screenshot path if captured, offline reload result, and any skipped manual proof with reason.

- [ ] **Step 6: Commit Task 7**

```bash
rtk git add apps/web docs/work/test-verification/PHASE-001-platform-smoke.md
rtk git commit -m "test: verify PWA shell offline behavior"
```

---

### Task 8: Reconciliation and Dehydration

**Files:**
- Modify: `docs/work/tickets/TICKET-001-repository-runtime-foundation.md`
- Modify: `docs/work/tickets/TICKET-002-postgresql-migrations-s3-platform-adapters.md`
- Modify: `docs/work/tickets/TICKET-003-google-oauth-user-isolation.md`
- Modify: `docs/work/tickets/TICKET-004-development-operations-verification-baseline.md`
- Modify: `docs/work/phases/PHASE-001-platform-identity.md`
- Modify: `docs/work/VALIDATION_MATRIX.md`
- Modify: `docs/work/TRACEABILITY.md`
- Modify: `docs/CONTEXT.md`
- Modify: `docs/work/BACKLOG.md`
- Modify: `docs/releases/CHANGELOG.md`
- Modify: `docs/architecture/API.md`
- Modify: `docs/architecture/ERD.md`
- Modify: `docs/architecture/ARCHITECTURE.md` only if runtime details need durable reconciliation.

**Interfaces:**
- Consumes: all verification evidence from Tasks 1-7.
- Produces: Harness completion state for PHASE-001.

- [ ] **Step 1: Update ticket verification results**

For each ticket, record exact commands, `pass`/`fail`/`skipped`, notes, UAT status, docs review, and completion checklist.

- [ ] **Step 2: Update validation matrix**

Rows REQ-F-001, REQ-NF-001, REQ-NF-004, REQ-NF-006, REQ-NF-007, and REQ-NF-008 must link to real evidence before moving to `implemented`.

- [ ] **Step 3: Reconcile master docs**

Update API with concrete endpoint schemas, CSRF header name, cookie behavior, health payloads, and error envelope. Update ERD with concrete PHASE-001 table definitions. Update architecture only if file/module/runtime specifics are durable enough for master docs.

- [ ] **Step 4: Check ADR need**

If implementation follows ADR-001 and ADR-002, record no new ADR needed. If it diverges in auth, runtime topology, or offline policy, stop and create/update ADR before claiming completion.

- [ ] **Step 5: Update context, backlog, phase, and changelog**

Move tickets and phase through the status model only as evidence supports. Set `docs/CONTEXT.md` next focus to PHASE-002 only after PHASE-001 is verified and dehydrated.

- [ ] **Step 6: Run final documentation checks**

Run: `rtk rg -n "TBD|not_started|pending execution|No execution evidence" docs/work/tickets/TICKET-00*.md docs/work/phases/PHASE-001-platform-identity.md docs/work/VALIDATION_MATRIX.md`

Expected: only intentionally unresolved future-phase entries remain; no PHASE-001 completion field claims unresolved evidence.

- [ ] **Step 7: Commit Task 8**

```bash
rtk git add docs
rtk git commit -m "docs: reconcile PHASE-001 execution evidence"
```

---

## Self-Review

- Spec coverage: The plan covers repository/runtime foundation, PWA shell, offline app-shell caching, API/config/error baseline, migrations, S3 adapter, OAuth/stateless cookie, CSRF, user isolation, worker baseline, local operations, verification, and reconciliation.
- Intentional deferral: Full offline read/write, IndexedDB mutation outbox, incremental sync, tombstones, and conflict resolution remain in PHASE-003 because PHASE-001 excludes wallet/category/transaction behavior.
- Placeholder scan: No `TBD`, `TODO`, `implement later`, or "similar to" placeholders are used in executable task steps.
- Type consistency: Shared interfaces are introduced before downstream tasks consume them.

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-08-25-phase-001-platform-identity.md`. Execution remains gated until the human approves `docs/work/phases/PHASE-001-detail-design.md`.

Two execution options after approval:

1. Subagent-Driven: dispatch a fresh subagent per task and review between tasks.
2. Inline Execution: execute tasks in this session with checkpoints.
