# Codebase Concerns

**Analysis Date:** 2026-09-10

## Tech Debt

**Frontend application monolith:**
- Issue: The primary React composition, state orchestration, data loading, mutation handling, and most screen routing remain concentrated in a 2,000-line component, while global presentation rules occupy a 2,200-line stylesheet.
- Files: `frontend/src/app/App.tsx`, `frontend/src/styles.css`, `frontend/src/app/components.tsx`
- Impact: Small UI changes have a large regression surface, shared state transitions are difficult to isolate, and mobile layout defects can propagate across unrelated journeys.
- Fix approach: Extract feature controllers/hooks and route-level containers from `frontend/src/app/App.tsx`; split feature styles beside components while retaining shared tokens and primitives in the existing UI layer.

**Large repository and HTTP adapter modules:**
- Issue: Finance and planning persistence and HTTP translation are implemented as broad files containing many unrelated commands and queries.
- Files: `backend/internal/finance/repository.go`, `backend/internal/planning/repository.go`, `backend/internal/platform/httpapi/finance.go`, `backend/internal/platform/httpapi/planning.go`
- Impact: Transaction boundaries, authorization predicates, and error mappings are easy to change inconsistently; reviews require loading large files with multiple domains.
- Fix approach: Split by aggregate/command without changing package boundaries, and keep transaction-scoped helpers explicit through `backend/internal/platform/commandtx/`.

**Incomplete product and operations surface:**
- Issue: AI/OCR/bank ingestion, export, account reset/deletion, complete production operations, and voice input remain deferred; several mounted quick-add actions intentionally expose unavailable states.
- Files: `docs/work/BACKLOG.md`, `docs/work/tickets/TICKET-018-shared-drafts-text-ai-chat.md`, `docs/work/tickets/TICKET-023-manual-export-jobs.md`, `docs/work/tickets/TICKET-024-account-reset-deletion.md`, `frontend/src/components/feedback/UnavailableAction.tsx`
- Impact: The application cannot yet claim full Money Lover-style parity, complete data portability, or complete account lifecycle handling.
- Fix approach: Preserve the explicit unavailable UI and implement deferred tickets only behind approved designs, public API documentation, privacy review, and end-to-end acceptance.

**Migration-dependent atomic sync rollout:**
- Issue: Atomic domain-write/change-feed behavior depends on migration 0012, and pre-existing change-feed omissions are not backfilled.
- Files: `backend/migrations/0012_atomic_sync_asset_feed.sql`, `backend/internal/platform/commandtx/`, `backend/internal/platform/changefeed/`, `docs/decisions/ADR-008-atomic-offline-sync-commit.md`
- Impact: Deploying the binary before the migration can break writes; upgraded clients can retain an incomplete local mirror if they continue from an old cursor.
- Fix approach: Apply `backend/migrations/0012_atomic_sync_asset_feed.sql` before the new API/worker binaries and require a one-time full client resync as part of the rollout.

## Known Bugs

**WebKit offline-emulation receipt readback:**
- Symptoms: A receipt `File` can be written to IndexedDB in a persistent WebKit profile, but reading the stored bytes while Playwright offline emulation is active can throw `NotReadableError`; the same record reads after emulation is disabled.
- Files: `frontend/src/offline/db.ts`, `frontend/e2e/receipt-controls.spec.ts`, `frontend/scripts/probe-receipt-storage.mjs`, `docs/work/test-verification/R0-RECEIPT-READBACK-2026-09-09.md`
- Trigger: Select an image, queue the transaction offline under the WebKit persistent-profile scenario, then call `arrayBuffer()` on the stored file while browser offline emulation remains enabled.
- Workaround: The HTTP-outage scenario is the accepted automated substitute; physical Safari/PWA testing is still required. Do not hide the retained failing emulation assertion or migrate the storage format without design review.

**Offline/browser interception differs when a service worker controls requests:**
- Symptoms: Playwright page routing does not reliably intercept service-worker-originated requests in WebKit, which can cause fixtures or injected failures to miss the application.
- Files: `frontend/playwright.config.ts`, `frontend/e2e/operation-feedback.spec.ts`, `frontend/e2e/accounting-correctness.spec.ts`, `docs/work/test-verification/R0-ERROR-FEEDBACK-2026-09-09.md`
- Trigger: Run a route-interception test with service workers enabled under WebKit.
- Workaround: Block service workers only in tests whose purpose is controlled network injection; retain real workers for PWA/offline acceptance.

## Security Considerations

**HTTP server and OAuth client lack explicit timeouts:**
- Risk: Slow-client connections can consume API resources, and Google token/profile calls can hang indefinitely because they use package-level HTTP helpers.
- Files: `backend/cmd/api/main.go`, `backend/internal/platform/httpapi/auth.go`
- Current mitigation: Deployment proxying and process-level supervision can bound some failures, but the application creates `http.ListenAndServe` and `http.PostForm` without application-level deadlines.
- Recommendations: Use an `http.Server` with read-header/read/write/idle timeouts and a dedicated OAuth `http.Client` with request timeout and context propagation.

**Request-body limits are inconsistent:**
- Risk: Authenticated endpoints that decode JSON without a byte cap can be used for memory/CPU pressure with oversized bodies.
- Files: `backend/internal/platform/httpapi/finance.go`, `backend/internal/platform/httpapi/planning.go`, `backend/internal/platform/httpapi/portfolio.go`, `backend/internal/platform/httpapi/sync.go`, `backend/internal/platform/httpapi/receipts.go`
- Current mitigation: Receipt metadata uses `http.MaxBytesReader`, sync limits mutation/change batches, and handlers validate domain values.
- Recommendations: Apply a shared maximum JSON body wrapper to every mutation endpoint, reject trailing JSON, and add oversized-body HTTP tests.

**Browser-local financial and identity caches are readable by same-origin script:**
- Risk: An XSS flaw would expose cached user identity, dashboard/asset data, offline transactions, conflicts, and pending mutations stored in localStorage/IndexedDB.
- Files: `frontend/src/app/auth.ts`, `frontend/src/app/userDataCache.ts`, `frontend/src/offline/db.ts`, `frontend/src/offline/migrations.ts`
- Current mitigation: IndexedDB data is namespaced per user, logout clears active-user state, API auth uses HttpOnly cookies, and API responses are not cached by the service worker.
- Recommendations: Maintain a strict Content-Security-Policy at the edge, avoid injecting untrusted HTML, minimize cached sensitive fields, document device-at-rest exposure, and include cache clearing in account deletion.

**Login allowlist is permissive when omitted outside production validation:**
- Risk: An accidentally exposed non-production deployment accepts any verified Google account when `ALLOWED_LOGIN_EMAILS` is empty.
- Files: `backend/internal/platform/httpapi/auth.go`, `backend/internal/platform/config/config.go`
- Current mitigation: Production requires HTTPS and several security secrets/services, and fixture OAuth is rejected in production.
- Recommendations: Require a non-empty allowlist for every remotely reachable environment or add an explicit opt-in flag for open registration.

## Performance Bottlenecks

**Unpaginated finance and planning collections:**
- Problem: Several list endpoints return every active user row and the frontend commonly reloads full collections after mutations.
- Files: `backend/internal/finance/repository.go`, `backend/internal/planning/repository.go`, `frontend/src/app/App.tsx`, `frontend/src/app/finance.ts`, `frontend/src/app/planning.ts`
- Cause: Wallet/category/transaction/budget/event/obligation APIs do not share cursor pagination; transaction listing orders the complete matching set.
- Improvement path: Add stable cursor pagination and bounded defaults to high-growth resources, preserve filter indexes, and update offline reconciliation to merge pages by entity identity.

**Analytics queries recompute aggregates on demand:**
- Problem: Dashboard and report endpoints repeatedly scan and aggregate transaction history; daily reports generate one row per calendar day in the requested range.
- Files: `backend/internal/analytics/repository.go`, `backend/migrations/0002_phase002_finance_core.sql`
- Cause: Reports are direct PostgreSQL aggregates without cached/materialized summaries, and query ranges are not visibly capped at the repository boundary.
- Improvement path: Enforce maximum report ranges, measure `EXPLAIN ANALYZE` on production-shaped data, add targeted composite indexes where evidence supports them, and introduce rollups only after measurement.

**Service worker runtime cache grows without an eviction policy:**
- Problem: Every successful cacheable same-origin GET outside `/api/` and `/docs` is added to a single cache until the cache version changes.
- Files: `frontend/public/sw.js`
- Cause: Runtime caching has no entry count, age limit, or URL allowlist beyond origin/path/method checks.
- Improvement path: Restrict runtime caching to fingerprinted assets/navigation shell and add bounded expiration or explicit old-entry cleanup.

## Fragile Areas

**Offline outbox, sync feed, and optimistic rendering:**
- Files: `frontend/src/offline/db.ts`, `frontend/src/offline/outbox.ts`, `frontend/src/app/pendingTransactions.ts`, `backend/internal/sync/service.go`, `backend/internal/sync/repository.go`, `backend/internal/platform/commandtx/`
- Why fragile: Correctness spans browser persistence, entity/type/operation identity, idempotency, PostgreSQL transactions, response-loss retries, cursors, tombstones, and conflict resolution.
- Safe modification: Preserve mutation IDs for retry identity but reconcile UI rows by entity ID/type/operation; keep domain write, receipt metadata, and change-feed insert in one transaction; require full-resync compatibility review for schema changes.
- Test coverage: Automated unit, PostgreSQL, race, and multi-browser coverage exists, but physical-device offline behavior and migration rollout recovery remain unaccepted.

**Financial arithmetic and reversal paths:**
- Files: `backend/internal/finance/transactions.go`, `backend/internal/finance/repository.go`, `backend/internal/portfolio/decimal.go`, `backend/internal/portfolio/ledger.go`, `backend/internal/analytics/repository.go`
- Why fragile: VND integer overflow, moving-average cost calculations, edit/archive reversals, and cross-wallet transfers can silently corrupt balances if arithmetic or transaction ordering changes.
- Safe modification: Use checked arithmetic before narrowing big integers, store original accounting deltas for exact reversal, lock affected wallets consistently, and test rollback after every downstream failure.
- Test coverage: Focused overflow, rollback, idempotency, stale-version, race, and browser accounting suites exist; production-scale values and live provider data remain outside current acceptance.

**Shared mobile sheet and fixed-overlay geometry:**
- Files: `frontend/src/app/App.tsx`, `frontend/src/app/components.tsx`, `frontend/src/components/charts/ComparisonBarChart.tsx`, `frontend/src/styles.css`, `frontend/e2e/mobile-layout.spec.ts`
- Why fragile: Wide chart labels or long content can enlarge the document beyond the viewport and displace fixed hit targets, causing unrelated buttons to be intercepted by overlays.
- Safe modification: Keep grid children shrinkable, bound visible labels while retaining full accessible/title values, and test zero/large datasets on mobile Chromium and WebKit without forced clicks.
- Test coverage: Controlled mobile layout and full browser suites cover the repaired chart path; physical installed-PWA viewport behavior is still pending.

**Shared integration-test database:**
- Files: `backend/internal/finance/repository_test.go`, `backend/internal/planning/repository_test.go`, `backend/internal/sync/repository_test.go`, `backend/internal/portfolio/repository_test.go`
- Why fragile: Package helpers reset shared schema state, so parallel package execution against one database produces invalid failures or false evidence.
- Safe modification: Run database packages serially (`-p 1`) or provision an isolated database/schema per package before enabling parallel execution.
- Test coverage: Serial full-suite and race runs are documented; parallel isolation is not implemented.

## Scaling Limits

**Single homelab deployment:**
- Current capacity: A single API process, worker process, PostgreSQL database, Redis cache, and S3-compatible object store are the documented deployment shape.
- Limit: There is no demonstrated capacity target, load test, multi-node failover, or horizontal deployment proof.
- Scaling path: Establish request/job/storage SLOs, add production-shaped load tests, measure database pool saturation, and validate lease/idempotency behavior with multiple API and worker instances.

**Worker batch processing:**
- Current capacity: Portfolio refresh defaults to 100 candidates, recurring processing defaults to 25 attempts, notification delivery defaults to 50, and audit purge defaults to 1,000 rows per run.
- Limit: Backlogs can grow faster than the minute worker cadence, while provider or database latency delays all jobs sharing the worker process.
- Scaling path: Instrument queue age and batch duration, make bounded concurrency explicit, separate independent job loops/processes when measured, and preserve database leases for multi-worker safety.

**Offline change history:**
- Current capacity: Change reads are capped at 500 and default to 100 entries per request.
- Limit: Long-disconnected clients require repeated pulls, and migration 0012 does not reconstruct historical omissions.
- Scaling path: Add cursor-lag telemetry, compact/tombstone retention policy, and a tested snapshot/full-resync threshold for stale clients.

## Dependencies at Risk

**Browser File/Blob persistence behavior:**
- Risk: IndexedDB `File`/`Blob` readback differs across WebKit automation/offline modes and is not a portable guarantee for physical Safari.
- Impact: Offline receipt attachments may be queued but unreadable for later upload on affected clients.
- Migration plan: Complete physical Safari acceptance; if failure reproduces, design and version an `ArrayBuffer`-based storage migration with memory/quota analysis and backward-compatible reads.

**Static-only portfolio price provider:**
- Risk: Automatic pricing currently relies on configured static quotes rather than a production market-data adapter.
- Impact: Portfolio valuations become stale or require manual entry and do not establish provider reliability/rate-limit behavior.
- Migration plan: Add a provider adapter registry, timeouts, rate-limit/backoff policy, stale-price UX, and live-provider contract tests while keeping manual prices permanent.

## Missing Critical Features

**Backup, restore, and disaster-recovery proof:**
- Problem: The release matrix keeps database locks, backup, and restore acceptance open.
- Blocks: Production readiness and a credible rollback/recovery claim.
- Files: `docs/work/VALIDATION_MATRIX.md`, `docs/operations/RELEASE_CHECKLIST.md`, `docs/work/tickets/TICKET-025-homelab-production-release-proof.md`

**Account export, reset, and deletion:**
- Problem: Users cannot complete the documented data portability and lifecycle flows; the phase remains deferred.
- Blocks: Complete privacy/account lifecycle support and safe removal of browser/server/object-store data.
- Files: `docs/work/tickets/TICKET-023-manual-export-jobs.md`, `docs/work/tickets/TICKET-024-account-reset-deletion.md`, `docs/work/BACKLOG.md`

**Production provider and physical-device acceptance:**
- Problem: Static portfolio prices and automated browser tests do not validate live providers, installed Safari/PWA storage, push delivery, attachment upload/readback, or all authenticated production paths.
- Blocks: Whole-app release acceptance despite green local unit/integration/browser suites.
- Files: `docs/work/test-verification/BUSINESS-LOGIC-AUDIT-2026-09-10.md`, `docs/work/test-verification/TICKET-027-asset-portfolio-valuation.md`, `docs/work/test-verification/R0-RECEIPT-READBACK-2026-09-09.md`, `docs/operations/RELEASE_CHECKLIST.md`

## Test Coverage Gaps

**Physical installed-PWA and receipt lifecycle:**
- What's not tested: Real iOS Safari/installed-PWA offline reload, durable receipt byte readback, reconnect upload, object-store download, and device quota/eviction behavior.
- Files: `frontend/e2e/pwa-shell.spec.ts`, `frontend/e2e/receipt-controls.spec.ts`, `frontend/src/offline/db.ts`, `frontend/src/app/receipts.ts`
- Risk: Automation can pass while real devices lose attachment availability or pending work.
- Priority: High

**Production migration and recovery drill:**
- What's not tested: Applying migration 0012 to production-shaped data, forcing the required client full resync, rollback boundaries, backup restoration, and recovery from a partially completed release.
- Files: `backend/migrations/0012_atomic_sync_asset_feed.sql`, `docs/decisions/ADR-008-atomic-offline-sync-commit.md`, `docs/operations/RELEASE_CHECKLIST.md`
- Risk: A correct local binary can deploy into an incompatible schema or leave clients with incomplete mirrors.
- Priority: High

**Load and resource-exhaustion behavior:**
- What's not tested: Concurrent slow clients, oversized JSON across all handlers, high-volume transaction/report queries, worker backlog growth, and cache/object-store failure under sustained load.
- Files: `backend/cmd/api/main.go`, `backend/cmd/worker/main.go`, `backend/internal/platform/httpapi/`, `backend/internal/analytics/repository.go`
- Risk: Homelab production can become unavailable or accumulate delayed jobs without correctness tests detecting it.
- Priority: Medium

**Deferred lifecycle and integration phases:**
- What's not tested: AI/OCR/bank webhook security, exports, deletion, restore, live market pricing, and voice input because those implementations are deferred or incomplete.
- Files: `docs/work/tickets/TICKET-018-shared-drafts-text-ai-chat.md`, `docs/work/tickets/TICKET-019-receipt-capture-ocr-adapter.md`, `docs/work/tickets/TICKET-021-signed-bank-webhook.md`, `docs/work/tickets/TICKET-023-manual-export-jobs.md`, `docs/work/tickets/TICKET-024-account-reset-deletion.md`
- Risk: Planning artifacts may be mistaken for shipped capability, and later integrations can introduce unreviewed privacy/authentication failure modes.
- Priority: Medium

---

*Concerns audit: 2026-09-10*
