---
artifact_type: changelog
id: CHANGELOG
status: active
owner: shared
human_fields: [release_approval]
ai_fields: [change_entries, linked_releases]
shared_fields: [status]
---

# Changelog

## Unreleased — 2026-09-11 completion contract

- Added review-first OpenAI-compatible Agent text/analysis endpoints and worker; provider output can create transaction drafts but never confirmed ledger entries.
- Added optional owned-receipt attachment with OCR Platform as a server-side third-party image tool, including leases, checksum/size checks and user/run scoping. No bank integration is included.
- Published the complete route inventory through OpenAPI 3.1 with named operation schemas, Bearer/cookie authorization rules, Redis rate limiting and API-agent skill guidance.
- Mounted Agent text/image UI and consolidated the product into a monochrome shared-component interface with five-tab navigation and non-blocking blurred PWA install prompt.
- Local verification: 177 frontend tests, production build and 105 Playwright E2E pass; physical iPhone/Safari UAT and production provider smoke remain release blockers.

## Unreleased — 2026-09-11 correctness follow-up

- Constrain shared comparison-chart columns and labels to available width. Large seven-day report values no longer expand the mobile viewport and displace fixed navigation/PWA hit targets; complete values remain in label text/title.
- Reject portfolio overflow in rounded quantities × prices, fees, cost basis, realized/unrealized profit and aggregate valuation; rejected commands preserve state/version/change feed.
- Reconcile pending transaction rows by entity ID, overlay edits, hide archives and exclude wallet/category/asset commands from transaction rows.
- Use UUIDs for concurrent offline browser fixtures; retain strict single-transaction assertions.
- Local-only. Full release gates remain tracked in the business-logic audit; not deployment evidence.

## Field Ownership

- Human owns release approval.
- AI maintains planned and released change entries with trace links.

## [Unreleased — business-logic and integration audit]

- Added a repository-owned AI audit skill plus public Docusaurus guidance so future agents can reproduce the finance/planning/analytics/sync/auth review contract.
- Fixed checked balance and reporting arithmetic, exact edit/archive reversals, receipt-preserving idempotency, recurring transaction validation, obligation repayment direction/reuse/locking and HCM calendar-date defaults.
- Added optimistic versions to wallet and transaction update/archive commands, including default-AI wallet selection.
- Hardened API-key authorization: an Authorization header is authoritative, every Bearer request is revalidated in PostgreSQL, and Redis is only a short-lived hint.
- Converted a sync version race into an explicit conflict response with authoritative server state.
- Bounded recurring catch-up by the worker batch limit so a long-offline daily schedule cannot create an unbounded transaction.
- Atomic sync now commits domain/feed/receipt together, serializes concurrent retries, emits direct finance/portfolio and draft-confirmation changes, and reads consistent resync snapshots.
- Ambiguous offline responses retain the original mutation ID instead of falling back to a fresh direct create. Sync preserves category parent clearing and canonical default-wallet responses.
- Local-only: apply migration 0012 before the binary and full client resync for historical feed omissions. No deployment is implied.

## [2026-09-10 prod-2026.09.10.3]

- Deployed production API and web images for the merged category hierarchy + recurring draft slice and e2e selector stability fixes:
  - Backend image: `registry.dungxbuif.com/mypocket-api:prod-2026.09.10.3` (`sha256:f8e1b9301e918bb6b9b7727643f3bf5113611c88c778f5386785179ea010208b`)
  - Web image: `registry.dungxbuif.com/mypocket-web:prod-2026.09.10.3` (`sha256:d36c7dc99cb78d3f9d013b63da96fe4ac92a2584dc59d06b317ad24180d7dacf`)
- Production deploy commands executed under `/Users/dungxbuif/production/mypocket/docker-compose.yml` for `api worker web` via `docker compose up -d --no-deps --pull never --force-recreate --wait --wait-timeout 60`, and migration run via `docker compose run --rm migrate`.
- Production proof:
  - `https://money.dungxbuif.com/api/v1/health/live` returns `{"status":"ok",...}`.
  - `https://money.dungxbuif.com/` returns HTML shell (HTTP 200).
- Verification at deploy time: `go test -p 1 ./... -count=1` (backend, with test DB), `npm test -- --run` (154), `npm run build` (typecheck + Vite build), `npx playwright test --config=playwright.r0.config.ts --workers=1` (72 passed).

## [2026-09-09 web/docs rollout]

### Planned — Sổ tiền redesign after functional completion

- Local-only F1 2026-09-10: corrected independent selected-budget totals, HCM date boundaries and asynchronous read/retry lifecycle feedback. Evidence: [SO-TIEN-F1](../work/test-verification/SO-TIEN-F1-2026-09-10.md). Not released; F2/F3 and redesign acceptance remain open.

- Owner approved the monochrome editorial direction. [Written specification](../superpowers/specs/2026-09-09-so-tien-design.md) is in review: correctness and docs first, shared components and all-screen migration next, safe old-UI cleanup last. No implementation or deployment in this design turn.

### Later local corrections — not deployed

- Empty-wallet detail normalizes null transactions instead of crashing. Transfer/adjustment rows show actual neutral amounts; quick-add exposes transfer between distinct wallets.
- Overview opens five server-backed report types with date/wallet filters, safe retry, privacy and numeric tables; unloaded reports no longer imply zero. Persistent offline report cache/charts/full parity remain absent.
- [Evidence and residual inventory](../work/test-verification/R0-FUNCTIONAL-E2E-2026-09-09.md): 120 frontend pass, full browsers 61 pass/2 retained WebKit failures; all new nine accounting checks pass. No deployment or Docusaurus publication in this run.

- Update at 21:25 Vietnam: redeployed the same web/docs images on explicit owner request. Public HTTPS/readiness/assets now pass, docs enforce login, and both images are published with verified registry manifests. Earlier connectivity/local-only limitations below describe the first rollout. No infrastructure repair, API behavior change or full/device acceptance is claimed. [Fresh evidence](../work/test-verification/R0-WEB-DOCS-RELEASE-2026-09-09.md).

- Owner authorized deployment. Current web deployed as `prod-2026.09.09.1`; API image `prod-2026.09.09.1-docs` updates Docusaurus only and retains previous API executables. Database, migrations and worker unchanged.
- Docusaurus now includes the recent functional fixes, integration links and remaining limits; canonical URL is `https://money.dungxbuif.com/docs/`. Corrected intro endpoint inventory and logout/cache description; existing API audit edits retained.
- [Release evidence](../work/test-verification/R0-WEB-DOCS-RELEASE-2026-09-09.md): 100 frontend tests, builds and local platform checks pass. Public HTTPS/registry connectivity is blocked from this host; new images are local-only, not published remotely. Authenticated public docs and physical Safari/PWA/receipt acceptance remain open.

## [Earlier implementation records]

The dated “not deployed” notes below describe their original verification runs. The web changes are now included in the bounded rollout above; diagnostic test changes and unrelated backend work are not implied to be released.

### Fixed — 2026-09-09 search feedback (not deployed)

- Search distinguishes loading, API failure and successful empty results. Query survives failures and panel toggles; explicit retry uses shared safe error/correlation-ID feedback. Obsolete responses cannot replace the current query's state.
- Reuses shared card/search input/buttons; clear action is named for accessibility. Offline search explicitly requires connectivity. [Verification](../work/test-verification/R0-SEARCH-FEEDBACK-2026-09-09.md): 100 frontend tests and 6 targeted browser checks pass. No API/DB change; full R0 and physical-device acceptance remain open.

### Investigated — 2026-09-09 receipt readback (not deployed)

- [Focused evidence](../work/test-verification/R0-RECEIPT-READBACK-2026-09-09.md) isolates the local WebKit readback symptom to offline emulation: the same stored bytes become readable again without rewriting. Added a separate real HTTP-abort plus offline-signal scenario, passing on Chrome and WebKit; original failing PWA/emulation proof remains active (targeted 8 pass/1 fail).
- No app/storage-format changes. Physical Safari/PWA and receipt-provider acceptance remain open; no release claim.

### In progress — 2026-09-09 R0 receipt controls (blocked, not deployed)

- One shared file picker serves both image actions; selected filename remains visible, removal/reselection is explicit, and debt/pending/read-only states prevent silent attachment loss.
- Four unsupported quick-add detail controls show an explicit unavailable state. Shared `SheetFrame` optionally separates its scroll body and footer so the save toolbar does not cover the remove action.
- [Verification](../work/test-verification/R0-RECEIPT-CONTROLS-2026-09-09.md): frontend 92 pass; receipt browser checks 5 pass/1 fail. WebKit stored image readback remains blocked with `NotReadableError`; no complete receipt-flow, iPhone-readiness or release claim.

### Fixed — 2026-09-09 R0 error feedback (not deployed)

- Existing transaction add/edit/archive keeps input and shows shared safe error feedback, without automatic financial retries. A saved offline transaction is not recreated when the separate receipt queue fails.
- API-key list/create/revoke/copy has loading/error/success feedback and read-only reload. Confirmed mutations survive a failed refresh; old list and copy completions cannot overwrite current feedback/state.
- [Verification](../work/test-verification/R0-ERROR-FEEDBACK-2026-09-09.md): frontend 85 pass; Chrome mobile/desktop 26 pass. Standalone WebKit probe isolates an offline-emulation/reload failure even with a minimal worker; no physical Safari acceptance or worker rewrite. Current UI style and API/schema remain unchanged; full R0 remains open.

### Fixed — 2026-09-09 R0 first slice (not deployed)

- Keep the existing UI style while reserving scroll space for the installation prompt and bottom dock on mobile; add dismiss/installed states and retain install help. Logout works with help expanded, using real browser clicks.
- Show persisted budget names with category context; native reusable action cards support keyboard and offline disabling. Budget gauges no longer replace valid zero values with sample amounts; unloaded data has an explicit unavailable state.
- [Verification](../work/test-verification/R0-FUNCTIONAL-STABILITY-2026-09-09.md): frontend 72 pass, Chrome mobile/desktop 22 pass. WebKit offline reload remains a reproduced failure; neither full R0 nor production readiness is claimed. [Remaining action inventory](../work/test-verification/R0-MOUNTED-ACTIONS-2026-09-09.md).

### Audit — 2026-09-08 Money Lover parity

- Added a source-linked inventory of 96 public Money Lover Help Center articles, 50 code/UI assessment rows, fresh test evidence and ordered remediation. This is original summary/comparison material, not a verbatim vendor-doc mirror.
- Corrected verified API portal drift for file upload/download paths, analytics parameters/envelopes/unsupported Insider fields and category-parent input; marked exports and other unimplemented contracts clearly. Portal snippets are not yet a fully validated OpenAPI contract.
- Recorded unresolved mobile logout-overlay obstruction and budget-name E2E mismatch; no application fix or production deploy included. Existing tests/build do not establish full functional or visual parity.

### Added

- Hardened the beta PWA install metadata for mobile/standalone use, repaired cached-auth recovery after an offline service-worker navigation, and added live mobile offline-reload proof. Financial browser caches and IndexedDB are user-namespaced; account switches preserve each owner's queued work separately. Bearer `mpk_...` keys are explicitly documented as full user-owned business API credentials, while API-key management remains browser-session only.
- Beta reliability fixes: secure cookies for configured HTTPS behind a reverse proxy, private/no-store docs, service-worker exclusion of private/cross-origin requests, restored category creation, wallet-detail runtime fixes, strict TypeScript build checks, and cache failure/expired-session regressions. Docker contexts exclude host dependencies and secrets; repeatable checks are in `scripts/verify-beta.sh`.
- Removed sample values from the newly introduced Overview report chart; it now renders only daily expense values returned by the reporting API or a Vietnamese empty state.

- Added Go API and worker runtime entrypoints with safe configuration validation, stable JSON error envelopes, correlation IDs, and health routes for PHASE-001.
- Added a PostgreSQL migration runner with idempotent ordered migrations, checksum mismatch protection, and the PHASE-001 `users` schema without provider-token or server-session persistence.
- Wired API and worker startup to verify required PostgreSQL availability from `DATABASE_URL`.
- Added an S3-compatible object-store adapter with safe smoke put/delete checks and complete S3 config validation.
- Added backend identity primitives and auth routes for fixture Google login, stateless signed cookies, CSRF enforcement, current-user lookup, and cross-user ownership guard testing.
- Added the mobile-first React PWA shell, manifest, service worker app-shell cache, Vietnamese navigation, quick-add sheet, and component design contract based on provided screenshots.
- Added a Docker Compose local stack for PostgreSQL, LocalStack S3, API, worker, migration, and nginx-served web, plus a repeatable platform smoke command.
- Split source code into `backend/` for the Go API/worker/migrations and `frontend/` for the React PWA, with npm package metadata owned by `frontend/`.
- Added browser auth proof for fixture Google login persistence, CSRF-backed logout, forbidden state rendering, credentialed CORS, and Compose web `/api` proxying.
- Added PHASE-002 PostgreSQL finance core schema for wallets, categories, wallet/category activation, transactions, idempotency keys, receipt metadata, and stable Vietnamese system category seeds.
- Added backend wallet/category validation and repository behavior for user-scoped wallet CRUD, default AI wallet selection, system/user category rules, and wallet/category activation.
- Added initial authenticated wallet/category API routes for listing and creating wallets plus listing categories.
- Added API-driven mobile wallet/category display in the PWA overview and add-transaction sheet.
- Added backend transaction accounting effect validation for income, expense, transfer, and balance adjustment balance deltas.
- Added repository create-transaction behavior with atomic wallet balance updates, version increments, category activation validation, and idempotent replay storage.
- Added repository transaction edit/archive reversal and user-scoped transaction search filters backed by persisted transaction deltas.
- Added authenticated transaction HTTP routes for create, edit, archive, and filtered listing with CSRF and idempotency header support.
- Added API-backed mobile transaction listing, search, and expense creation from the quick-add sheet.
- Added user-scoped receipt metadata persistence with private object-key and SHA-256 validation.
- Added an IndexedDB-compatible browser outbox layer (localStorage fallback) for optimistic offline transactions and ordered reconnect replay.
- Added authenticated wallet/category mutation API routes for wallet edit/archive/default AI selection, category create/edit/archive, and per-wallet category activation.
- Added mobile finance controls for wallet/category management plus income, expense, transfer, adjustment, report exclusion, and transaction edit/archive workflows.
- Added live mobile Playwright coverage for fixture-login finance CRUD through the real API.
- Added the PHASE-003 frontend IndexedDB mirror/outbox foundation with legacy localStorage migration/quarantine, wallet/category/transaction offline mutation queueing, cached-auth offline reload, cached finance hydration, degraded read-only UI, and mobile PWA offline reload proof.
- Added PHASE-003 sync API/change-feed support with PostgreSQL sync cursors/changes/mutation ledger, idempotent mutation replay, stale-version conflict responses, authoritative resync snapshots, frontend sync outbox drain, and mobile reconnect replay proof.
- Added PHASE-003 conflict inbox and recovery support with IndexedDB conflict persistence, non-blocking mobile review UI, keep-server, discard-local, edit-and-retry, full resync preservation, and targeted mobile E2E proof.
- Added PHASE-004 budget and threshold support with planning tables, authenticated budget API routes, Ho Chi Minh period progress, selected/all-category scopes, 80%/100% alert dedupe, and mobile create/edit/archive proof.
- Added PHASE-004 event and debt planning with event/obligation tables, authenticated API routes, event transaction links, repayment links with overpayment protection, mobile Vietnamese planning screens, and live mobile E2E proof.
- Added PHASE-004 recurring schedules with deterministic occurrence keys, transaction draft generation, worker lease processing, authenticated schedule/draft APIs, mobile schedule setup, and pending draft review rows.
- Added TICKET-027 Wave 1 backend portfolio schema and domain repository for user-owned asset positions, ordered buy/sell trades, append-only manual/provider price history, moving-average cost, realized/unrealized P&L, archive retention, and wallet-accounting isolation.
- Added login email allowlisting through `ALLOWED_LOGIN_EMAILS`, plus asset portfolio REST routes, dashboard investment/combined net-worth fields, Account-tab asset UI, IndexedDB asset cache/outbox, sync replay for asset mutations, and a static-price provider worker path for automatic valuation UAT.
- Added TICKET-028 production-hardening foundations: structured API/worker logging, safe panic recovery, PostgreSQL audit events, hidden audit event API, audit retention worker, Account-tab user-managed API keys for AI agents/third-party callers, Redis-backed auth cache with DB fallback and revoke invalidation, and production config validation for audit/API-key/cache secrets.
- Extended TICKET-028 with whitelist-only audit access checking and Account audit-log UI, user-scoped receipt uploads/downloads via S3-compatible presigned URLs, transaction receipt attachment, and IndexedDB offline receipt retry. OCR remains deferred.
- Expanded the default category seed with the complete Vietnamese expense, income, and debt parent/child catalog in migration `0011_phase002_category_catalog.sql`, preserving legacy keys for existing data.

### Planning

- Approved TICKET-027 immediately after TICKET-017 with hybrid automatic/manual pricing, moving-average buy/sell accounting, separate wallet/investment/combined totals, offline manual mutations, an accepted wallet/portfolio boundary ADR, executable plan, and verification target; implementation is now in review pending human UAT.
- Approved executable detail designs for PHASE-003 through PHASE-007 and extended the shared mobile component design contract through the originally planned release scope.
- Approved and recorded the MyPocket React PWA and Go modular-monolith system design.
- Migrated the source SRS into Harness requirements, user stories, architecture, API, ERD, integrations, SDD, ADRs, roadmap, eight phases, backlog, traceability, and validation planning.
- Redefined the current M1 release to stop after TICKET-017; TICKET-018 through TICKET-025 are deferred to M2/post-M1 work, and voice input moves after that.
- Added durable notification inbox/Web Push subscription lifecycle, search and wallet detail APIs, authoritative dashboard totals, and normalized analytics reports through TICKET-017.
- Replaced the hardcoded Money Insider Home card and subscription CTAs with an authenticated report for the most frequent expense category, real spending, elapsed-day average, prior-month comparison, privacy masking, empty state, and refresh action.
- Refined the mobile wallet experience with a compact Money Lover-style wallet sheet, create-wallet entry flow, main-page wallet-only card, liquid-glass navigation/sheets, and transparent scrollbars.
- Added Tailwind CSS v4 through the Vite plugin and introduced reusable frontend base primitives for cards, pills, icon buttons, sheets, section titles, and class composition.
- Fixed blank-page runtime crashes from extension-scheme cache requests and malformed authenticated-user payloads.
- Added configurable Google OAuth authorization-code exchange with state-cookie validation and Google userinfo provisioning; fixture mode remains available for local tests.
- Recorded full offline conflict review, review-first AI/OCR/webhook ingestion, Google OAuth without server sessions, and restricted 180-day-default audit logging.
- Created PHASE-001 ticket artifacts, detail design, and executable implementation plan focused on the first deployable PWA/mobile shell, platform, identity, and offline app-shell cache slice.
- Created PHASE-002 finance-core tickets, detail design, and executable implementation plan for wallet/category, transaction accounting, and Vietnamese seed/receipt metadata slices.
- Created PHASE-003 offline synchronization tickets, verification target, and executable implementation plan for IndexedDB mirror/outbox, idempotent sync APIs, change feed, conflict inbox, and full resync.
- Created PHASE-004 through PHASE-007 ticket artifacts, verification targets, and executable implementation plans for planning automation, analytics/dashboard, AI/OCR/webhook ingestion, audit/export/account lifecycle, and production release proof.
