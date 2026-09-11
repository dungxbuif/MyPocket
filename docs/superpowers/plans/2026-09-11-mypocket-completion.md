# MyPocket Production Completion Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Complete, verify, document, and deploy MyPocket as a production-grade Vietnamese personal-finance PWA with trustworthy accounting, public API-key access, an OpenAI-compatible agent, and third-party OCR image tooling.

**Architecture:** Preserve the existing React PWA plus Go modular-monolith architecture. PostgreSQL remains authoritative; IndexedDB is a user-scoped offline mirror/outbox; Redis is non-authoritative cache/coordination; S3-compatible storage holds private user artifacts; OpenAI-compatible and OCR Platform calls stay behind backend adapters and never mutate accounting without a validated user confirmation.

**Tech Stack:** Go 1.23, PostgreSQL, Redis, AWS SDK v2/S3-compatible storage, React 19, TypeScript, Vite, IndexedDB, Vitest, Playwright, Docusaurus, Docker Compose.

**Spec:** `docs/superpowers/specs/2026-09-11-mypocket-complete-personal-finance-design.md`

## Global Constraints

- Preserve all existing user work in the dirty tree; never reset or overwrite an unrelated edit.
- Use checked integer VND accounting and reject overflow before persistence.
- PostgreSQL is authoritative; Redis and IndexedDB cannot preserve revoked or deleted server state.
- Browser-cookie and Bearer API-key callers converge on the same authenticated user and domain authorization path.
- Invalid Bearer credentials never fall back to cookies; cookie writes require CSRF, API-key writes do not.
- Every user-owned query filters by authenticated `user_id`; request bodies never select ownership.
- AI and OCR outputs are untrusted and can create only reviewable drafts or read-only analysis.
- `mac-ocr` is a separately deployed third-party agent tool, not a MyPocket receipt-domain component or public MyPocket OCR proxy.
- Bank integration and voice input remain out of scope.
- Repeated interactions use base components; `frontend/src/app/App.tsx` remains composition rather than gaining new feature logic.
- Every task follows RED → GREEN → full relevant regression → focused commit.
- Deployment is forbidden until migration, full resync, backup/restore, provider smoke, physical-device UAT, public docs, and rollback gates pass.
- This plan supersedes conflicting deferred/bank-specific instructions in older Phase 006/007 plans; those files remain historical evidence, while bank integration stays out of scope.

---

### Task 1: Reconcile the Brownfield Baseline

**Files:**
- Create: `docs/work/completion/BASELINE-2026-09-11.md`
- Modify only after ownership review: currently changed backend, frontend, migration, test, and docs files listed by `git status --short`
- Test: existing backend, frontend, Playwright, and Docusaurus suites

**Interfaces:**
- Consumes: current `main` at design commit `a7bc52b` plus the complete dirty worktree
- Produces: reviewable feature commits and a baseline evidence document containing exact commands, results, and unresolved gates

- [x] **Step 1: Capture an immutable inventory before changing code**

```bash
rtk proxy git status --short
rtk proxy git diff --name-status
rtk proxy git diff --check
rtk proxy git log -8 --oneline
```

Write `BASELINE-2026-09-11.md` with the HEAD SHA, every dirty path grouped as `accounting-sync`, `frontend-pwa`, `docs`, `generated`, or `unrelated/needs-owner-preservation`, and never stage the final group.

- [x] **Step 2: Run the existing proof before reconciliation**

```bash
rtk proxy env GOCACHE=/private/tmp/mypocket-go-cache 'MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' MYPOCKET_TEST_REDIS_URL=redis://127.0.0.1:6380/0 go test -race -p 1 ./... -count=1
rtk proxy go vet ./...
rtk npm test -- --run
rtk npm run build
rtk npm run test:e2e
rtk npm --prefix docs run build
```

Run Go commands from `backend/` and npm commands from `frontend/`. Record exact pass/fail counts; a failure is a real baseline finding, not permission to weaken a test.

- [x] **Step 3: Reconcile related changes in dependency order**

Commit only reviewed groups in this order: migration/shared transaction primitives; finance/planning/portfolio/sync domain changes and tests; HTTP/API changes and tests; frontend/offline/PWA changes and tests; docs/evidence. Generated Docusaurus output is committed only if the repository already tracks that exact output path.

```bash
rtk proxy git diff --cached --check
rtk proxy git diff --cached --stat
rtk proxy git commit -m "fix: preserve atomic finance and sync invariants"
```

- [x] **Step 4: Re-run the full baseline at reconciled HEAD**

Use the six commands from Step 2 and update the evidence document with the new commit SHAs. The task is complete only when the tree contains no unclassified change and unrelated preserved edits remain unstaged.

- [x] **Step 5: Commit the evidence**

```bash
rtk proxy git add docs/work/completion/BASELINE-2026-09-11.md
rtk proxy git commit -m "docs: record completion baseline"
```

---

### Task 2: Prove Migration 0012, Full Resync, Backup, and Restore

**Files:**
- Modify: `backend/internal/platform/db/migrate.go`
- Modify: `backend/internal/platform/db/migrate_test.go`
- Modify: `backend/internal/sync/service.go`
- Modify: `backend/internal/sync/service_test.go`
- Modify: `frontend/src/offline/db.ts`
- Modify: `frontend/src/offline/db.test.ts`
- Modify: `frontend/src/offline/syncApi.ts`
- Modify: `frontend/e2e/offline-sync.spec.ts`
- Create: `scripts/backup-production.sh`
- Create: `scripts/restore-drill.sh`
- Create: `docs/operations/backup-restore.md`
- Create: `docs/work/completion/DATA-SAFETY.md`

**Interfaces:**
- Consumes: migration `backend/migrations/0012_atomic_sync_asset_feed.sql`, `GET /api/v1/sync/resync`, user-scoped IndexedDB database names
- Produces: migration lock/checksum proof, one-time client full-resync marker, encrypted/permission-restricted PostgreSQL plus S3 backup, isolated restore verification

- [x] **Step 1: Write failing migration and resync tests**

Add tests proving concurrent migrators serialize, modified applied migrations fail checksum validation, migration 0012 upgrades a production-shaped fixture, and clients with a pre-0012 schema marker perform exactly one full resync while retaining outbox/conflicts.

```go
func TestMigration0012UpgradesProductionShapeAndIsIdempotent(t *testing.T) {
    // migrate through 0011, seed wallet/transaction/asset rows, then migrate twice
    // assert schema_migrations checksum and canonical sync feed constraints
}
```

```ts
it("performs one post-0012 full resync without deleting pending intent", async () => {
  await seedLegacyMirrorWithOutbox();
  await reconcile({ requiredServerEpoch: "atomic-sync-v1" });
  expect(await readPendingMutations()).toHaveLength(1);
  expect(await readMeta("server_epoch")).toBe("atomic-sync-v1");
});
```

- [x] **Step 2: Run focused tests and verify RED**

```bash
rtk proxy go test ./internal/platform/db ./internal/sync -count=1
rtk npm test -- --run src/offline/db.test.ts
```

- [x] **Step 3: Implement the migration epoch and full-resync gate**

Return a stable server epoch from resync metadata and persist it only after the complete authoritative snapshot commits to IndexedDB. Keep pending mutations, receipt bytes, and conflicts in separate stores during snapshot replacement.

```go
type Snapshot struct {
    ServerEpoch string `json:"server_epoch"`
    Cursor      int64  `json:"cursor"`
    // existing authoritative collections remain unchanged
}

const ServerEpoch = "atomic-sync-v1"
```

- [x] **Step 4: Implement backup and isolated restore scripts**

`backup-production.sh` must use `pg_dump --format=custom --no-owner`, capture an S3 object inventory, write SHA-256 checksums, create files with mode `0600`, and fail on an unset destination. `restore-drill.sh` must reject a production database URL, restore into an explicit isolated database/bucket prefix, compare row counts and object checksums, then leave evidence for manual inspection.

```bash
: "${MYPOCKET_BACKUP_DIR:?set an explicit backup directory}"
umask 077
pg_dump --format=custom --no-owner --file "$MYPOCKET_BACKUP_DIR/database.dump" "$DATABASE_URL"
```

- [x] **Step 5: Run focused and browser proof**

```bash
rtk proxy go test ./internal/platform/db ./internal/sync -count=1
rtk npm test -- --run src/offline
rtk npm run test:e2e -- offline-sync.spec.ts
rtk proxy ./scripts/restore-drill.sh
```

- [x] **Step 6: Commit**

```bash
rtk proxy git add backend/internal/platform/db backend/internal/sync frontend/src/offline frontend/e2e/offline-sync.spec.ts scripts/backup-production.sh scripts/restore-drill.sh docs/operations/backup-restore.md docs/work/completion/DATA-SAFETY.md
rtk proxy git commit -m "feat: prove atomic migration and recoverable resync"
```

---

### Task 3: Close Finance, Planning, Portfolio, Analytics, and Notification Gaps

**Files:**
- Modify: `backend/internal/finance/*_test.go`
- Modify: `backend/internal/planning/*_test.go`
- Modify: `backend/internal/portfolio/*_test.go`
- Modify: `backend/internal/analytics/repository_test.go`
- Create: `backend/internal/notification/webpush.go`
- Create: `backend/internal/notification/webpush_test.go`
- Modify: `backend/internal/notification/worker.go`
- Modify: `backend/internal/notification/retry_test.go`
- Modify: `backend/internal/platform/config/config.go`
- Modify: `backend/internal/platform/config/config_test.go`
- Modify: `backend/cmd/worker/main.go`
- Modify: `backend/go.mod`
- Modify: `backend/go.sum`
- Create: `backend/migrations/0013_notification_delivery_state.sql`
- Modify: `backend/internal/platform/httpapi/*_test.go`
- Modify: `frontend/Dockerfile`
- Modify: `frontend/src/app/notifications.ts`
- Create: `frontend/e2e/notifications.spec.ts`
- Modify: `compose.yaml`
- Modify: `frontend/e2e/accounting-correctness.spec.ts`
- Modify: `frontend/e2e/business-core.spec.ts`
- Modify: `frontend/e2e/planning-automation.spec.ts`
- Modify: `frontend/src/screens/ReportsPanel.tsx`
- Modify: `frontend/src/screens/ReportsPanel.test.tsx`
- Create: `docs/work/completion/BUSINESS-CORRECTNESS.md`

**Interfaces:**
- Consumes: existing domain repositories and versioned REST routes
- Produces: executable proof for wallet/category CRUD, all transaction types, exact reversal, neutral transfer reporting, recurring drafts, budgets, events, obligations, portfolio, six-period analytics, durable inbox, and real Web Push delivery

- [x] **Step 1: Add one ledger fixture with independently calculated expectations**

```go
type ledgerExpectation struct {
    WalletBalances map[string]int64
    IncomeVND      int64
    ExpenseVND     int64
    NetVND         int64
}
```

The fixture must cover income, expense, transfer, adjustment, excluded expense, edited amount/wallet/category, archived transaction, repayment, and portfolio trade. Assert wallet balances independently from analytics queries.

- [x] **Step 2: Run domain/integration tests and verify any RED result is reproducible**

```bash
rtk proxy env 'MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test -p 1 ./internal/finance ./internal/planning ./internal/portfolio ./internal/analytics ./internal/notification ./internal/platform/httpapi -count=1
```

- [x] **Step 3: Correct only proven invariant failures**

Use checked helpers for every add/subtract/multiply and preserve the existing atomic command transaction. A representative boundary remains:

```go
func checkedAdd(left, right int64) (int64, error) {
    if (right > 0 && left > math.MaxInt64-right) || (right < 0 && left < math.MinInt64-right) {
        return 0, ErrMoneyOverflow
    }
    return left + right, nil
}
```

- [x] **Step 4: Complete six-period report behavior and mounted UI states**

`ReportsPanel` must render loading, error with retry, empty, privacy-masked, current-period, six-period comparison, cumulative series, and drilldown states from server data. Do not calculate authoritative totals from DOM-visible rows.

```ts
export type ReportQuery = {
  from: string;
  to: string;
  wallet_id?: string;
  category_id?: string;
  include_excluded?: false;
};
```

- [x] **Step 5: Replace no-op push delivery with production Web Push**

Add `github.com/SherClockHolmes/webpush-go` at reviewed release `v1.4.0` and wrap it behind the existing `notification.Delivery` interface. Read `VAPID_PUBLIC_KEY`, `VAPID_PRIVATE_KEY`, and `VAPID_SUBJECT`; production rejects partial configuration and never logs private keys, endpoints, `p256dh`, or auth secrets. Classify HTTP `404/410` as expired subscriptions; retry only transient failures using the existing bounded schedule.

```go
type WebPushDelivery struct {
    PublicKey  string
    PrivateKey string
    Subject    string
    Client     *http.Client
}
```

Expose `VITE_WEB_PUSH_PUBLIC_KEY` as a Docker build argument sourced from the same public key, and keep `NoopDelivery` only when push is explicitly disabled outside production.

- [x] **Step 6: Expand real-browser journeys without force clicks**

Create data only through the real API/UI. Assert a transfer changes both wallet balances but not income/expense reports; an excluded expense changes its wallet but not reports; edit and archive reverse the exact original effect; recurring execution creates a draft and never changes a balance before confirmation.

```bash
rtk npm run test:e2e -- accounting-correctness.spec.ts business-core.spec.ts planning-automation.spec.ts notifications.spec.ts
```

- [x] **Step 7: Run full regressions and commit**

```bash
rtk proxy go test -race -p 1 ./... -count=1
rtk npm test -- --run
rtk npm run test:e2e
rtk proxy git add backend/internal/finance backend/internal/planning backend/internal/portfolio backend/internal/analytics backend/internal/notification backend/internal/platform/config backend/internal/platform/httpapi backend/cmd/worker/main.go backend/go.mod backend/go.sum frontend/Dockerfile frontend/src/app/notifications.ts frontend/src/screens/ReportsPanel.tsx frontend/src/screens/ReportsPanel.test.tsx frontend/e2e/accounting-correctness.spec.ts frontend/e2e/business-core.spec.ts frontend/e2e/planning-automation.spec.ts frontend/e2e/notifications.spec.ts compose.yaml docs/work/completion/BUSINESS-CORRECTNESS.md
rtk proxy git commit -m "fix: close personal finance correctness gaps"
```

---

### Task 4: Add Import, Export, and Account Lifecycle Jobs

**Files:**
- Create: `backend/migrations/0014_data_lifecycle.sql`
- Create: `backend/internal/lifecycle/types.go`
- Create: `backend/internal/lifecycle/repository.go`
- Create: `backend/internal/lifecycle/repository_test.go`
- Create: `backend/internal/lifecycle/export.go`
- Create: `backend/internal/lifecycle/export_test.go`
- Create: `backend/internal/lifecycle/import.go`
- Create: `backend/internal/lifecycle/import_test.go`
- Modify: `backend/internal/platform/objectstore/s3.go`
- Modify: `backend/internal/platform/objectstore/objectstore_test.go`
- Create: `backend/internal/platform/httpapi/lifecycle.go`
- Create: `backend/internal/platform/httpapi/lifecycle_test.go`
- Create: `backend/internal/worker/lifecycle.go`
- Create: `backend/internal/worker/lifecycle_test.go`
- Modify: `backend/internal/platform/httpapi/router.go`
- Modify: `backend/cmd/api/main.go`
- Modify: `backend/cmd/worker/main.go`
- Modify: `frontend/src/screens/AccountScreen.tsx`
- Modify: `frontend/src/screens/AccountScreen.test.tsx`
- Create: `frontend/src/app/lifecycle.ts`
- Create: `frontend/e2e/account-lifecycle.spec.ts`
- Create: `frontend/e2e/import-export.spec.ts`

**Interfaces:**
- Produces: `POST /api/v1/imports`, `GET /api/v1/imports/{id}`, `POST /api/v1/imports/{id}/confirm`, `POST /api/v1/exports`, `GET /api/v1/exports/{id}`, `GET /api/v1/exports/{id}/download`, `POST /api/v1/account/reset`, `POST /api/v1/account/delete`, `GET /api/v1/account/jobs/{id}`
- Produces: `lifecycle.Repository`, `lifecycle.ImportParser`, `lifecycle.ExportFormatter`, `ObjectStore.PutObject`, and `worker.LifecycleRunner`

- [x] **Step 1: Write the failing schema and repository tests**

```sql
CREATE TABLE data_jobs (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL REFERENCES users(id),
  kind text NOT NULL CHECK (kind IN ('import','export','reset','delete')),
  status text NOT NULL CHECK (status IN ('queued','running','awaiting_confirmation','completed','failed')),
  idempotency_key text NOT NULL,
  request jsonb NOT NULL,
  result jsonb,
  result_object_key text,
  attempts integer NOT NULL DEFAULT 0,
  version bigint NOT NULL DEFAULT 1,
  error_code text,
  next_attempt_at timestamptz NOT NULL DEFAULT now(),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (user_id, kind, idempotency_key)
);
```

Tests must prove two-user isolation, import preview/confirmation, immutable export snapshot selection, idempotent replay, exact reset scope, immediate access disable for delete, resumable jobs, and failure on foreign S3 prefixes.

- [x] **Step 2: Verify RED**

```bash
rtk proxy go test ./internal/lifecycle ./internal/platform/httpapi ./internal/worker -count=1
```

- [x] **Step 3: Implement CSV export and private download**

```go
type ExportRequest struct {
    From     *time.Time `json:"from,omitempty"`
    To       *time.Time `json:"to,omitempty"`
    Datasets []string   `json:"datasets"`
}

type ExportFormatter interface {
    Write(ctx context.Context, snapshot Snapshot, out io.Writer) error
}
```

Use UTF-8 CSV, stable Vietnamese-neutral machine headers, RFC3339 timestamps, integer VND, deterministic ordering, and short-lived S3 download URLs. Never export credentials, hashes, audit IP/user-agent hashes, or provider payloads.

Extend the object-store adapter with a bounded streaming upload used by the worker; keep presigned browser upload behavior unchanged.

```go
func (s *S3Store) PutObject(ctx context.Context, key, contentType string, body io.Reader, size int64) error
```

- [x] **Step 4: Implement review-first CSV import**

Before destructive operations, implement import as a review-first job. Accept only a private owned CSV object up to 15 MiB and 10,000 rows; require UTF-8 stable headers, parse dates/time zones explicitly, accept integer VND only, resolve wallet/category references to owned records, and return a preview containing valid rows plus row-numbered errors. Confirmation revalidates the preview version and applies valid rows through existing finance commands with stable per-row idempotency keys; any validation error prevents the entire import from posting.

```go
type ImportPreview struct {
    ID          string           `json:"id"`
    Version     int64            `json:"version"`
    ValidRows   []ImportRow      `json:"valid_rows"`
    Errors      []ImportRowError `json:"errors"`
    Confirmable bool             `json:"confirmable"`
}
```

- [x] **Step 5: Implement reset/delete planning and resumable cleanup**

Require recent authentication, exact typed confirmation, CSRF for cookie callers, `Idempotency-Key`, and a server-produced affected-count preview token. Reset preserves identity and lifecycle audit evidence. Delete disables authentication first, revokes API keys/cache, then deletes only `users/{userID}/` objects and documented database rows.

```go
type DestructiveRequest struct {
    Confirmation string `json:"confirmation"`
    PreviewToken string `json:"preview_token"`
}
```

- [x] **Step 6: Mount Account UI and E2E states**

Use existing base dialog/button/error components for import preview/errors/confirm, export progress/download, destructive preview, typed confirmation, cancel, terminal failure, retry, and success. Tests must prove import preview never mutates balances, cancel is non-mutating, confirmed valid import applies once, and cross-user job/download URLs return `404`.

```bash
rtk npm test -- --run src/screens/AccountScreen.test.tsx
rtk npm run test:e2e -- account-lifecycle.spec.ts import-export.spec.ts
```

- [x] **Step 7: Run regressions and commit**

```bash
rtk proxy go test -race -p 1 ./... -count=1
rtk npm test -- --run
rtk proxy git add backend/migrations/0014_data_lifecycle.sql backend/internal/lifecycle backend/internal/platform/objectstore/s3.go backend/internal/platform/objectstore/objectstore_test.go backend/internal/platform/httpapi/lifecycle.go backend/internal/platform/httpapi/lifecycle_test.go backend/internal/platform/httpapi/router.go backend/internal/worker/lifecycle.go backend/internal/worker/lifecycle_test.go backend/cmd/api/main.go backend/cmd/worker/main.go frontend/src/app/lifecycle.ts frontend/src/screens/AccountScreen.tsx frontend/src/screens/AccountScreen.test.tsx frontend/e2e/account-lifecycle.spec.ts frontend/e2e/import-export.spec.ts
rtk proxy git commit -m "feat: add export and account lifecycle jobs"
```

---

### Task 5: Publish a Complete API Contract and Reusable Agent Skill

**Files:**
- Create: `backend/internal/platform/httpapi/openapi.json`
- Create: `backend/internal/platform/httpapi/openapi.go`
- Create: `backend/internal/platform/httpapi/openapi_test.go`
- Create: `backend/internal/platform/httpapi/authorization_matrix_test.go`
- Create: `backend/internal/platform/ratelimit/redis.go`
- Create: `backend/internal/platform/ratelimit/redis_test.go`
- Modify: `backend/internal/platform/config/config.go`
- Modify: `backend/internal/platform/config/config_test.go`
- Modify: `backend/internal/platform/httpapi/router.go`
- Modify: `backend/cmd/api/main.go`
- Create: `skills/mypocket-api/SKILL.md`
- Create: `skills/mypocket-api/references/api-quick-reference.md`
- Create: `frontend/docs/docs/api/openapi.mdx`
- Create: `frontend/docs/docs/skills/mypocket-api.mdx`
- Modify: `frontend/docs/sidebars.ts`
- Modify: existing `frontend/docs/docs/api/*.mdx`

**Interfaces:**
- Produces: unauthenticated `GET /api/v1/openapi.json`, API-key rate limiting, exhaustive authorization evidence, and public Docusaurus API/skill pages
- Produces: repository-versioned `mypocket-api` skill whose examples authenticate with `Authorization: Bearer <user-api-key>`

- [x] **Step 1: Write a failing router-contract test**

```go
func TestOpenAPICoversEveryVersionedRoute(t *testing.T) {
    routes := RegisteredRoutesForTest()
    contract := LoadOpenAPIForTest(t)
    for _, route := range routes {
        if route.Path == "/api/v1/health/live" || route.Path == "/api/v1/health/ready" { continue }
        requireOpenAPIOperation(t, contract, route.Method, route.Path)
    }
}
```

Also validate that every mutating operation documents cookie CSRF and Bearer API-key security, idempotency where implemented, stable error envelopes, and correlation IDs.

Add a table-driven authorization matrix generated from the same route descriptors. For every user-owned route, prove unauthenticated `401`, invalid/revoked Bearer `401` without cookie fallback, valid cookie behavior, valid API-key behavior, foreign ownership `404`, malformed input `400`, stale version `409`, and audit actor/action/outcome/correlation metadata where the operation changes state.

- [x] **Step 2: Verify OpenAPI, authorization-matrix, and limiter tests are RED**

```bash
rtk proxy go test ./internal/platform/httpapi ./internal/platform/ratelimit -count=1
```

- [x] **Step 3: Add the embedded OpenAPI 3.1 contract and route inventory**

```go
//go:embed openapi.json
var openAPIDocument []byte

func openAPI(w http.ResponseWriter, _ *http.Request) {
    w.Header().Set("Content-Type", "application/vnd.oai.openapi+json")
    _, _ = w.Write(openAPIDocument)
}
```

Define explicit route descriptors beside router registration so contract tests compare normalized `{id}` templates rather than scraping handler source.

- [x] **Step 4: Add API-key rate limiting**

Before documentation, add a Redis-backed per-user/per-key rate limiter at the authenticated boundary. `API_RATE_LIMIT_PER_MINUTE` defaults to `120`, production requires a positive bounded value, responses return `429` plus `Retry-After`, and Bearer requests fail with a retryable `503` if the limiter cannot enforce its limit. Browser-cookie traffic keeps its existing UI-oriented protections and does not inherit an unavailable Redis dependency.

```go
type Limiter interface {
    Allow(ctx context.Context, key string, limit int, window time.Duration) (allowed bool, retryAfter time.Duration, err error)
}
```

- [x] **Step 5: Write and validate the repository skill**

The skill must tell agents how to discover OpenAPI, create/revoke user API keys, authenticate, paginate, send idempotency/version headers, handle `401/403/404/409/429/5xx`, use sync safely, confirm drafts, and avoid logging keys. It must not contain a real key or production private data.

```markdown
---
name: mypocket-api
description: Use MyPocket's public API to manage an authenticated user's personal-finance data and reviewable AI drafts.
---
```

- [x] **Step 6: Build docs and run secret/contract/authorization checks**

```bash
rtk proxy go test ./internal/platform/httpapi -count=1
rtk proxy go test ./internal/platform/ratelimit -count=1
rtk npm --prefix docs run build
rtk proxy rg -n "sk_[A-Za-z0-9_-]{16,}|BEGIN (RSA|OPENSSH|EC) PRIVATE KEY" skills frontend/docs backend/internal/platform/httpapi/openapi.json
```

The secret scan must return no real secret; documented placeholders are written as `<user-api-key>`.

- [x] **Step 7: Commit**

```bash
rtk proxy git add backend/internal/platform/httpapi/openapi.json backend/internal/platform/httpapi/openapi.go backend/internal/platform/httpapi/openapi_test.go backend/internal/platform/httpapi/authorization_matrix_test.go backend/internal/platform/httpapi/router.go backend/internal/platform/ratelimit backend/internal/platform/config backend/cmd/api/main.go skills/mypocket-api frontend/docs
rtk proxy git commit -m "feat: publish and protect the public API contract"
```

---

### Task 6: Add the OpenAI-Compatible Agent and Review-First Drafts

**Files:**
- Create: `backend/migrations/0015_agent.sql`
- Create: `backend/internal/agent/types.go`
- Create: `backend/internal/agent/validation.go`
- Create: `backend/internal/agent/validation_test.go`
- Create: `backend/internal/agent/repository.go`
- Create: `backend/internal/agent/repository_test.go`
- Create: `backend/internal/agent/service.go`
- Create: `backend/internal/agent/service_test.go`
- Create: `backend/internal/platform/openai/client.go`
- Create: `backend/internal/platform/openai/client_test.go`
- Create: `backend/internal/platform/httpapi/agent.go`
- Create: `backend/internal/platform/httpapi/agent_test.go`
- Modify: `backend/internal/platform/httpapi/openapi.json`
- Modify: `backend/internal/platform/httpapi/openapi_test.go`
- Create: `backend/internal/worker/agent.go`
- Create: `backend/internal/worker/agent_test.go`
- Modify: `backend/internal/platform/config/config.go`
- Modify: `backend/internal/platform/config/config_test.go`
- Modify: `backend/internal/platform/httpapi/router.go`
- Modify: `backend/cmd/api/main.go`
- Modify: `backend/cmd/worker/main.go`
- Modify: `.env.example`
- Modify: `compose.yaml`

**Interfaces:**
- Produces: `POST /api/v1/agent/messages`, `GET /api/v1/agent/runs/{id}`, and existing draft confirm/reject interoperability
- Produces: `agent.Model`, `agent.Service`, a leased `worker.AgentRunner`, and an OpenAI-compatible Responses/Chat Completions adapter selected by configuration

- [x] **Step 1: Write config and provider contract tests**

```go
type Model interface {
    Generate(ctx context.Context, request ModelRequest) (ModelResponse, error)
}

type ModelRequest struct {
    System string
    Input  string
    Schema json.RawMessage
}
```

Tests must cover custom base URL, model, Bearer key, timeout, bounded retry only for retryable status, maximum response bytes, malformed JSON, refusal, and redacted errors.

- [x] **Step 2: Verify RED**

```bash
rtk proxy go test ./internal/platform/config ./internal/platform/openai ./internal/agent ./internal/platform/httpapi -count=1
```

- [x] **Step 3: Add production-safe configuration**

```go
type AIConfig struct {
    Enabled    bool
    BaseURL    string
    APIKey     string
    Model      string
    Timeout    time.Duration
    MaxRetries int
}
```

Read `AI_ENABLED`, `AI_BASE_URL`, `AI_API_KEY`, `AI_MODEL`, `AI_TIMEOUT`, and `AI_MAX_RETRIES`. Production with `AI_ENABLED=true` must reject missing URL/key/model, non-HTTPS URL, timeout outside `1s..120s`, or retries outside `0..3`. Never include the key in formatted config/errors.

Migration `0015_agent.sql` adds an immutable agent-run request/provenance row and extends existing `transaction_drafts` with nullable agent source/provenance fields without changing recurring-draft behavior:

```sql
CREATE TABLE agent_runs (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  idempotency_key text NOT NULL,
  kind text NOT NULL CHECK (kind IN ('transaction_draft','analysis')),
  status text NOT NULL CHECK (status IN ('queued','processing','completed','failed')),
  request_text text NOT NULL CHECK (length(request_text) BETWEEN 1 AND 8000),
  response_text text,
  error_code text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (user_id, idempotency_key)
);
ALTER TABLE transaction_drafts ADD COLUMN IF NOT EXISTS agent_run_id uuid REFERENCES agent_runs(id) ON DELETE SET NULL;
ALTER TABLE transaction_drafts ADD COLUMN IF NOT EXISTS provenance jsonb NOT NULL DEFAULT '{}'::jsonb;
```

- [x] **Step 4: Implement typed queued agent runs and finance-context minimization**

```go
type ProposedTransaction struct {
    Type                string  `json:"type"`
    AmountVND           int64   `json:"amount_vnd"`
    SourceWalletID      string  `json:"source_wallet_id"`
    DestinationWalletID *string `json:"destination_wallet_id,omitempty"`
    CategoryID          *string `json:"category_id,omitempty"`
    OccurredAt          string  `json:"occurred_at"`
    Note                string  `json:"note"`
}
```

The HTTP service stores an idempotent queued run and returns immediately. `worker.AgentRunner` claims due runs with a PostgreSQL lease/`FOR UPDATE SKIP LOCKED`, commits the claim, calls the provider outside the transaction, then persists the terminal result. Send only owned active wallet/category IDs plus display names and the minimum requested aggregate context. Validate schema, ownership, transaction shape, integer money, dates, and stale versions before inserting a pending draft. Analysis responses are read-only and cite the date/filter scope used.

- [x] **Step 5: Prove confirmation is the only accounting boundary**

Tests must assert provider success, retry, timeout, injection text, foreign IDs, overflow, unknown fields, and duplicate requests leave wallet balances unchanged. Only existing `ConfirmTransactionDraft` may call the finance command, with user ownership, optimistic version, and idempotency.

```bash
rtk proxy env 'MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test -p 1 ./internal/agent ./internal/planning ./internal/finance ./internal/platform/httpapi -count=1
```

- [x] **Step 6: Commit**

```bash
rtk proxy git add backend/migrations/0015_agent.sql backend/internal/agent backend/internal/platform/openai backend/internal/worker/agent.go backend/internal/worker/agent_test.go backend/internal/platform/httpapi/agent.go backend/internal/platform/httpapi/agent_test.go backend/internal/platform/httpapi/openapi.json backend/internal/platform/httpapi/openapi_test.go backend/internal/platform/httpapi/router.go backend/internal/platform/config backend/cmd/api/main.go backend/cmd/worker/main.go .env.example compose.yaml
rtk proxy git commit -m "feat: add review-first OpenAI-compatible agent"
```

---

### Task 7: Add OCR Platform as a Third-Party Agent Image Tool

**Files:**
- Create: `backend/migrations/0016_agent_image_tools.sql`
- Create: `backend/internal/platform/ocr/client.go`
- Create: `backend/internal/platform/ocr/client_test.go`
- Create: `backend/internal/agent/image_tool.go`
- Create: `backend/internal/agent/image_tool_test.go`
- Modify: `backend/internal/agent/types.go`
- Modify: `backend/internal/agent/repository.go`
- Modify: `backend/internal/agent/service.go`
- Modify: `backend/internal/platform/objectstore/s3.go`
- Modify: `backend/internal/platform/objectstore/objectstore_test.go`
- Modify: `backend/internal/platform/httpapi/agent.go`
- Modify: `backend/internal/platform/httpapi/agent_test.go`
- Modify: `backend/internal/platform/httpapi/openapi.json`
- Modify: `backend/internal/platform/httpapi/openapi_test.go`
- Create: `backend/internal/worker/agent_tools.go`
- Create: `backend/internal/worker/agent_tools_test.go`
- Modify: `backend/internal/platform/config/config.go`
- Modify: `backend/internal/platform/config/config_test.go`
- Modify: `backend/cmd/api/main.go`
- Modify: `backend/cmd/worker/main.go`
- Modify: `.env.example`
- Modify: `compose.yaml`

**Interfaces:**
- Consumes: OCR Platform `POST /v1/documents`, `GET /v1/documents/{documentId}`, `GET /v1/ocr/capabilities`
- Produces: `agent.ImageTool`, `ocr.Client`, `ObjectStore.GetObject`, and leased `worker.AgentToolRunner`
- Does not produce: a generic MyPocket OCR proxy or receipt-owned OCR lifecycle

- [x] **Step 1: Write fake-provider and private-object RED tests**

```go
type ImageTool interface {
    Submit(ctx context.Context, input ImageInput) (ToolSubmission, error)
    Read(ctx context.Context, providerID string) (ToolResult, error)
}

type ImageInput struct {
    ContentType string
    Bytes       []byte
    Languages   []string
}
```

Cover capabilities drift, correct Bearer header, 202 submission, pending `Retry-After`, completed result, failed/cancelled, 404, 410 expiry, 429 quota, timeout, oversized response, and secret-free errors. Object-store tests must fail reads over 15 MiB and checksum mismatch before OCR submission.

- [x] **Step 2: Verify RED**

```bash
rtk proxy go test ./internal/platform/ocr ./internal/platform/objectstore ./internal/agent ./internal/worker ./internal/platform/httpapi -count=1
```

- [x] **Step 3: Implement size-bounded S3 reads and OCR adapter**

```go
func (s *S3Store) GetObject(ctx context.Context, key string, maxBytes int64) ([]byte, error) {
    // GetObject, read through io.LimitReader(maxBytes+1), reject overflow, close body
}
```

Submit JSON Base64 because MyPocket's 15 MiB limit is below OCR Platform's 25 MiB decoded limit. Keep the external client inside `platform/ocr`; agent code depends only on `ImageTool`.

- [x] **Step 4: Add migration 0016, persist generic agent tool runs, and process them with a lease**

Migration `0016_agent_image_tools.sql` creates the tool-run storage after Task 6's general agent schema. Store owning user, image object ID, optional agent/receipt association, provider document ID, state, attempt count, timestamps, result expiry, safe result/provenance, and redacted error code. Claim rows with `FOR UPDATE SKIP LOCKED`; never hold a database transaction across an HTTP provider call.

```sql
CREATE TABLE agent_tool_runs (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  agent_run_id uuid REFERENCES agent_runs(id) ON DELETE CASCADE,
  receipt_object_id uuid REFERENCES receipt_objects(id) ON DELETE SET NULL,
  provider text NOT NULL CHECK (provider = 'ocr'),
  provider_document_id text,
  status text NOT NULL CHECK (status IN ('queued','submitting','processing','completed','failed','cancelled','expired')),
  attempts integer NOT NULL DEFAULT 0 CHECK (attempts >= 0),
  result jsonb,
  result_expires_at timestamptz,
  error_code text,
  next_attempt_at timestamptz NOT NULL DEFAULT now(),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX agent_tool_runs_due_idx ON agent_tool_runs (next_attempt_at, id)
  WHERE status IN ('queued','submitting','processing');
```

```go
type ToolRunStatus string
const (
    ToolQueued ToolRunStatus = "queued"
    ToolSubmitting ToolRunStatus = "submitting"
    ToolProcessing ToolRunStatus = "processing"
    ToolCompleted ToolRunStatus = "completed"
    ToolFailed ToolRunStatus = "failed"
    ToolCancelled ToolRunStatus = "cancelled"
    ToolExpired ToolRunStatus = "expired"
)
```

- [x] **Step 5: Feed completed OCR into agent context and optional receipt drafts**

The agent may summarize/general-analyze OCR output. If `receipt_id` is supplied and owned, it may request the typed transaction schema from Task 6; the result remains a pending draft. Tests prove a completed tool run alone creates no transaction or wallet change.

- [x] **Step 6: Add fail-fast/degraded configuration behavior**

Read `OCR_ENABLED`, `OCR_BASE_URL`, `OCR_API_KEY`, `OCR_TIMEOUT`, `OCR_POLL_INTERVAL`, and `OCR_MAX_PROCESSING_TIME`. Production release requires enabled valid HTTPS configuration, but a runtime outage marks only OCR capability/jobs unavailable while `/health/ready` continues to represent MyPocket's core database readiness.

- [x] **Step 7: Run provider-contract and regression tests, then commit**

```bash
rtk proxy go test -race ./... -count=1
rtk proxy go vet ./...
rtk proxy git add backend/migrations/0016_agent_image_tools.sql backend/internal/platform/ocr backend/internal/platform/objectstore backend/internal/agent backend/internal/worker/agent_tools.go backend/internal/worker/agent_tools_test.go backend/internal/platform/httpapi/agent.go backend/internal/platform/httpapi/agent_test.go backend/internal/platform/httpapi/openapi.json backend/internal/platform/httpapi/openapi_test.go backend/internal/platform/httpapi/router.go backend/internal/platform/config backend/cmd/api/main.go backend/cmd/worker/main.go .env.example compose.yaml
rtk proxy git commit -m "feat: add third-party OCR agent tool"
```

---

### Task 8: Mount Agent and Optional Receipt-Result UI

**Files:**
- Create: `frontend/src/app/agent.ts`
- Create: `frontend/src/screens/AgentScreen.tsx`
- Create: `frontend/src/screens/AgentScreen.test.tsx`
- Create: `frontend/src/components/agent/AgentMessage.tsx`
- Create: `frontend/src/components/agent/AgentRunStatus.tsx`
- Modify: `frontend/src/components/inputs/FilePickerInput.tsx`
- Modify: `frontend/src/app/App.tsx`
- Modify: `frontend/src/screens/sheets/AddTransactionSheet.tsx`
- Modify: `frontend/src/screens/sheets/AddTransactionSheet.test.tsx`
- Create: `frontend/e2e/agent.spec.ts`
- Create: `frontend/e2e/agent-image.spec.ts`

**Interfaces:**
- Consumes: agent message/run APIs from Tasks 6–7 and existing transaction draft confirmation routes
- Produces: text/image agent workflow, provider-independent statuses, draft review, and optional receipt-result reuse

- [x] **Step 1: Write failing component tests**

Test text submission, image validation, queued/processing/completed/failed/expired states, retry, safe error copy, result-to-draft review, reject, edit, and confirm. Verify agent output is rendered as text, never unsafe HTML.

```ts
export type AgentRunView = {
  id: string;
  status: "queued" | "submitting" | "processing" | "completed" | "failed" | "cancelled" | "expired";
  answer?: string;
  draft_ids: string[];
  retryable: boolean;
};
```

- [x] **Step 2: Verify RED**

```bash
rtk npm test -- --run src/screens/AgentScreen.test.tsx src/screens/sheets/AddTransactionSheet.test.tsx
```

- [x] **Step 3: Implement focused API/state modules and base-component UI**

Keep networking and polling in `frontend/src/app/agent.ts`; keep rendering in focused agent components; add only route/sheet composition to `App.tsx`. Reuse `ActionButton`, `FilePickerInput`, `OperationError`, card, select, dialog/sheet, and status primitives.

- [x] **Step 4: Implement optional receipt reuse**

An owned receipt can be attached to an agent request. A completed OCR result can populate a visible editable draft, but the existing Save/Confirm action remains the only ledger mutation. Failure leaves the original receipt accessible.

- [x] **Step 5: Run real-browser E2E**

```bash
rtk npm run test:e2e -- agent.spec.ts agent-image.spec.ts
```

Assert cookie and API-key flows, foreign image isolation, provider timeout/retry, no balance change before confirmation, exactly one change after confirmation, and no forced clicks.

- [x] **Step 6: Commit**

```bash
rtk proxy git add frontend/src/app/agent.ts frontend/src/screens/AgentScreen.tsx frontend/src/screens/AgentScreen.test.tsx frontend/src/components/agent frontend/src/components/inputs/FilePickerInput.tsx frontend/src/app/App.tsx frontend/src/screens/sheets frontend/e2e/agent.spec.ts frontend/e2e/agent-image.spec.ts
rtk proxy git commit -m "feat: mount text and image agent workflows"
```

---

### Task 9: Consolidate the Minimal Monochrome UI and Physical PWA Behavior

**Files:**
- Modify: `frontend/src/app/App.tsx`
- Modify: `frontend/src/app/components.tsx`
- Modify: `frontend/src/components/ui/*`
- Modify: `frontend/src/components/feedback/*`
- Modify: `frontend/src/screens/*`
- Modify: `frontend/src/styles.css`
- Modify: `frontend/public/manifest.webmanifest`
- Modify: `frontend/public/sw.js`
- Modify: `frontend/e2e/mobile-layout.spec.ts`
- Modify: `frontend/e2e/install-prompt.spec.ts`
- Modify: `frontend/e2e/pwa-shell.spec.ts`
- Modify: `frontend/e2e/receipt-controls.spec.ts`
- Create: `docs/work/completion/PHYSICAL-DEVICE-UAT.md`

**Interfaces:**
- Consumes: all mounted workflows from Tasks 3–8
- Produces: one black/white-first responsive design system, no legacy duplicate UI, accessible install/receipt/offline behavior

- [ ] **Step 1: Inventory repeated controls and legacy duplicates**

Record every button, input, select, dialog/sheet, card, table/list row, status, error, empty state, skeleton, toast, and install prompt. Each repeated interaction must map to an existing base component or a new focused base component before screen changes.

- [ ] **Step 2: Write visual-contract and accessibility tests**

```ts
test("mobile sheets keep the primary action above the safe area and keyboard", async ({ page }) => {
  await openTransactionSheet(page);
  await expect(page.getByRole("button", { name: "Lưu" })).toBeInViewport();
});
```

Add keyboard navigation, focus restore, visible labels, contrast, 44px touch targets, reduced motion, safe-area, overflow, and loading/error/empty assertions. Install tests must call the captured `beforeinstallprompt.prompt()` only after a user click and fall back to iOS instructions when the event is unavailable.

- [ ] **Step 3: Verify RED and implement base-component migration**

```bash
rtk npm test -- --run src/test/baseComponents.test.tsx
rtk npm run test:e2e -- mobile-layout.spec.ts install-prompt.spec.ts pwa-shell.spec.ts receipt-controls.spec.ts
```

Replace legacy screen-local primitives only after each replacement has equivalent test coverage. Keep semantic structure and monochrome tokens centralized; use color only for essential state meaning.

- [ ] **Step 4: Run physical iPhone/Safari UAT**

On a real device verify install instructions, installed launch, standalone navigation, safe areas, camera/file receipt selection, airplane-mode save, reconnect upload, exact receipt byte recovery, agent image submission, keyboard/focus, and no blank page. Record device/OS/browser, build SHA, screenshots, and outcome in `PHYSICAL-DEVICE-UAT.md`.

- [ ] **Step 5: Remove proven legacy UI and run full frontend proof**

```bash
rtk proxy rg -n "legacy|deprecated|old-ui" frontend/src
rtk npm test -- --run
rtk npm run build
rtk npm run test:e2e
```

Remove only paths proven replaced; any remaining search match must be explained in the UAT document.

- [ ] **Step 6: Commit**

```bash
rtk proxy git add frontend docs/work/completion/PHYSICAL-DEVICE-UAT.md
rtk proxy git commit -m "feat: complete monochrome responsive PWA UI"
```

---

### Task 10: Update Public Docs, Requirements, and Evidence

**Files:**
- Modify: `docs/requirements/REQUIREMENTS.md`
- Modify: `docs/work/BACKLOG.md`
- Modify: `docs/work/VALIDATION_MATRIX.md`
- Modify: `docs/releases/CHANGELOG.md`
- Modify: `docs/architecture/API.md`
- Modify: `docs/architecture/ERD.md`
- Modify: `frontend/docs/docs/intro.mdx`
- Modify: `frontend/docs/docs/api/*.mdx`
- Create: `frontend/docs/docs/guides/agent-and-images.mdx`
- Create: `frontend/docs/docs/guides/export-and-account-lifecycle.mdx`
- Create: `frontend/docs/docs/updates/2026-09-11-completion.mdx`
- Modify: `frontend/docs/sidebars.ts`

**Interfaces:**
- Consumes: implemented routes, schemas, UI behavior, verification artifacts, and public skill
- Produces: public docs that describe the deployed product exactly and trace every requirement to current evidence

- [ ] **Step 1: Write a docs-to-router drift checklist**

For every route descriptor, record its OpenAPI operation, Docusaurus page, authentication modes, success schema, failure codes, idempotency/version semantics, and a curl example. Marking a requirement complete requires a linked test/UAT artifact, not a code-path assertion.

- [ ] **Step 2: Update requirements and remove stale scope statements**

Change old references that call OCR receipt-owned, AI deferred, export deferred, or bank active. Preserve historical verification documents as historical evidence; add a dated superseding note rather than rewriting old test outcomes.

- [ ] **Step 3: Build docs and validate links/content**

```bash
rtk npm --prefix docs run build
rtk proxy rg -n "OCR deferred|bank webhook|implementation pending" docs frontend/docs/docs skills
rtk proxy git diff --check
```

Every remaining match must be either removed or explicitly identified as historical/out-of-scope context.

- [ ] **Step 4: Commit**

```bash
rtk proxy git add docs frontend/docs skills
rtk proxy git commit -m "docs: publish complete MyPocket product and API guidance"
```

---

### Task 11: Run Release Gates, Deploy, Smoke, and Rollback Drill

**Files:**
- Modify: `.env.example`
- Modify: `/Users/dungxbuif/production/mypocket/docker-compose.yml`
- Modify: `scripts/verify-beta.sh`
- Modify: `scripts/smoke-platform.sh`
- Create: `scripts/smoke-production.sh`
- Create: `docs/operations/release-runbook.md`
- Create: `docs/work/completion/RELEASE-2026-09-11.md`

**Interfaces:**
- Consumes: tested application images, production database/Redis/S3, OpenAI-compatible provider, OCR Platform, DNS/TLS, public docs
- Produces: deployed MyPocket and docs URLs, authenticated web/API-key/provider smoke evidence, rollback tag and exercised recovery procedure

- [ ] **Step 1: Add a fail-closed release verifier**

```bash
required_files=(
  docs/work/completion/BASELINE-2026-09-11.md
  docs/work/completion/DATA-SAFETY.md
  docs/work/completion/BUSINESS-CORRECTNESS.md
  docs/work/completion/PHYSICAL-DEVICE-UAT.md
)
for file in "${required_files[@]}"; do test -s "$file" || exit 1; done
```

The verifier must also require clean diff checks, migration status 0015 or later, backup/restore evidence, complete backend/frontend/docs/E2E suites, physical UAT pass, and configured provider smoke. It must not read or print secret values.

- [ ] **Step 2: Run the complete local/production-shaped gate**

```bash
rtk proxy ./scripts/verify-beta.sh
rtk proxy ./scripts/restore-drill.sh
rtk proxy ./scripts/smoke-platform.sh
rtk proxy git diff --check
```

- [ ] **Step 3: Create immutable rollback points and backups**

Record the current production image digests and database migration version, run `backup-production.sh`, verify checksums, and create a signed/annotated release tag only from the exact tested commit. Never migrate before a verified backup exists.

- [ ] **Step 4: Deploy in safe order**

Update `/Users/dungxbuif/production/mypocket/docker-compose.yml` with the exact tested image tags and the AI/OCR/S3 environment bindings required by both API and worker. Deploy migration job, API, worker, frontend, and embedded docs in that order. Wait for each bounded health/readiness condition; do not expose the new frontend until API and worker are on the compatible schema. Use this existing production stack owning `money.dungxbuif.com`; do not invent a second stack.

- [ ] **Step 5: Run authenticated production smoke**

`smoke-production.sh` must verify TLS, public PWA assets/manifest/service worker, authenticated user load, API-key wallet/category/transaction/report read/write in a dedicated smoke account, revoked-key rejection despite Redis cache, export lifecycle, agent text draft, OCR image tool, receipt-result reuse, docs/OpenAPI/skill reachability, and exactly-once cleanup of smoke data.

```bash
rtk proxy ./scripts/smoke-production.sh https://money.dungxbuif.com
```

- [ ] **Step 6: Exercise rollback**

Roll back application images to the recorded digests. If schema compatibility permits application-only rollback, verify old-version health and authenticated reads. If database rollback is required, restore only into an isolated target first, compare counts/checksums, then follow the documented maintenance-window procedure. Record timings and results.

- [ ] **Step 7: Re-deploy the verified release and publish evidence**

Re-deploy the tested release, repeat production smoke, record exact commit/image/migration/provider versions and public links in `RELEASE-2026-09-11.md`, then update validation status only for gates with direct evidence.

- [ ] **Step 8: Commit release artifacts**

```bash
rtk proxy git add .env.example scripts docs/operations docs/work/completion docs/work/VALIDATION_MATRIX.md docs/releases/CHANGELOG.md
rtk proxy git commit -m "release: verify and deploy MyPocket completion"
```

## Final Completion Audit

Before claiming completion, independently map every item in the design spec's `Approved Boundaries`, `Verification Contract`, and `Completion Definition` to a current file, test output, physical UAT record, or production smoke result. The milestone remains incomplete if any mapping is missing, historical-only, skipped, or inferred from a narrower test.

Run the final evidence commands from the exact release commit:

```bash
rtk proxy env GOCACHE=/private/tmp/mypocket-go-cache 'MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' MYPOCKET_TEST_REDIS_URL=redis://127.0.0.1:6380/0 go test -race -p 1 ./... -count=1
rtk proxy go vet ./...
rtk npm test -- --run
rtk npm run build
rtk npm run test:e2e
rtk npm --prefix docs run build
rtk proxy ./scripts/restore-drill.sh
rtk proxy ./scripts/smoke-production.sh https://money.dungxbuif.com
rtk proxy git diff --check
```

Only after every command and manual gate passes may the release evidence mark the milestone complete.
