# Beta reliability verification and fix design

Status: in_review. Authorization: user requested reliable installation, comprehensive testing and fixes on 2026-09-07. iOS release is not approved without physical-device proof of offline installation/reload.

Links: [backlog](../BACKLOG.md), [validation](../VALIDATION_MATRIX.md), [context](../../CONTEXT.md), [release notes](../../releases/CHANGELOG.md).

Scope: existing beta workflows, real database integration, browser CRUD/offline replay, ownership, cookie handling, and install/build validation. No production deployment or external messages.

Baseline: frontend 48 tests pass; full backend suite fails Secure-cookie and unauthenticated-docs expectations; browser CRUD tests target controls removed by the latest design. Database tests reset the public schema, so use a dedicated newly created `mypocket_verify_20260907` database, never the application database.

Design: use configured HTTPS origin for Secure cookies behind TLS termination, preserve local HTTP development; enforce private docs access and no-store; update browser scenarios to exercise the current UI and explicitly report missing category controls. Verify each reproduced defect with negative cases and surrounding suites. Review owner transitions and offline queues for loss or cross-account replay before release.

Reconciliation: update this evidence record, validation, backlog/context and release notes. Existing API/schema contracts retained. Secure-cookie handling implements the existing HTTPS contract.

Evidence so far: real PostgreSQL suites passed for finance, planning, portfolio, sync, identity, analytics, notification and migrations with `MYPOCKET_TEST_DATABASE_URL` targeting the dedicated database and `go test -p 1 -count=1`.

Verified commands (run in backend or frontend as appropriate):

- `go test -race -p 1 -count=1 ./...` with the dedicated integration database: pass; `go vet ./...`: pass after removing an ineffective self-assignment.
- `npm run typecheck`: pass, includes all screens and tests, strict mode.
- `npm test -- --run`: 54 tests cover expired auth while browser reports offline, full storage, foreign cache, private service-worker requests, known-offline shell navigation, owner-switch pending preservation and startup replay isolation.
- `npm run test:e2e -- --repeat-each=2`: 36 Chromium mobile/desktop runs passed before the final namespace change. Final three-project suite: 26/27 passed; WebKit offline reload failed with an internal browser error. Do not treat this as passed or suppress it with a skip.
- Final-source rerun: `npm run typecheck && npm test -- --run && npm run test:e2e -- --project=mobile --project=desktop`: strict typecheck, 54 unit/component tests and 18/18 Chromium E2E tests passed. The WebKit offline case remains unresolved as described below.
- Docker API and web images build with clean `npm ci`; API container readiness responds with database OK. Image build is not proof of production S3/OAuth/push integration.
- `git diff --check`: pass.
- 2026-09-08 regression review found and corrected three release blockers: the mounted overview now renders its daily expense bars from `/api/v1/reports/daily`; logout preserves the active user's isolated IndexedDB queue; and HTTP proof confirms bearer API keys can list only their owner's wallets. Fresh verification: frontend `npm test -- --run` 55/55 pass, HTTP `go test ./internal/platform/httpapi -count=1` pass, and `npm run build` pass.

Reproduction/fixes: Secure cookies depended only on direct TLS; docs lacked auth; category UI had been removed; refresh completion raced offline category use; Vite omitted type checks and hid runtime references to nonexistent variables; cache fallback hid auth errors and propagated quota failures. Owner switches discarded pending work; per-user databases now retain it ([ADR-007](../../decisions/ADR-007-offline-user-databases.md)).

Re-run prerequisites: create disposable `mypocket_verify_*` and `mypocket_e2e` databases on test PostgreSQL at port 55433, install frontend dependencies and Playwright Chromium/WebKit, then run `MYPOCKET_TEST_DATABASE_URL=<dedicated-url> bash scripts/verify-beta.sh`. The script fails on any failing gate. Playwright starts fresh API/web processes and uses fixture OAuth only.

Residual acceptance: resolve WebKit offline reload; validate installed PWA on physical iPhone/Android; real Google OAuth/HTTPS reverse proxy, private S3 upload/download and Web Push have not been accepted by this local run. Test evidence cannot establish absence of all possible bugs. Legacy unowned offline queues are retained but not replayed without verified ownership.

WebKit control experiment: a standalone local HTTP server serves a minimal page and a service worker which always returns static HTML from `new Response`, without any MyPocket code or network fetch. After `serviceWorker.ready` and controller presence, `context.setOffline(true); page.reload()` fails with the identical WebKit internal error. The MyPocket test additionally proves its controller is active and shell exists in CacheStorage before the failure. This demonstrates a test-engine failure independent of the application, but does not substitute for iPhone installation acceptance. No retry/skip was added to conceal it.

Docs review: context, backlog, validation and changelog reconciled; API/schema retained; owner-storage decision recorded. No production deployment or commit performed.
