---
artifact_type: project_context
id: CONTEXT
status: active
owner: shared
human_fields:
  - current_focus
  - open_questions
  - priority_override
ai_fields:
  - recently_touched_areas
  - recent_decisions
  - next_steps
  - queue_summary
shared_fields:
  - current_status
  - active_backlog
  - current_queue_focus
  - active_phase
  - active_ticket
  - active_bug
updated: 2026-08-31
---

# Project Context

## Field Ownership

- Human owns product intent, priority overrides, and unresolved product decisions.
- AI owns concise implementation-state refreshes, evidence links, and next-step recommendations.
- This context does not silently override approved scope in the product spec or phase files.

## Current Status

- Latest verification 2026-09-11: full backend race suite on dedicated PostgreSQL, Go vet, 161 frontend tests, typecheck/production build and **87/87 E2E pass**. [Mobile layout diagnosis](work/phases/MOBILE-OVERLAY-DETAIL_DESIGN.md) resolved chart min-content overflow that displaced fixed navigation; prior 14 mobile failures are cleared. Portfolio/offline/layout slices are locally in_review. Physical-device/provider acceptance and migration 0012/full resync release preparation remain separate; no deployment.

- 2026-09-11 continuation: [portfolio checked-money](work/phases/PORTFOLIO-MONEY-DETAIL_DESIGN.md) fixes reproduced overflow in costs, P&L and valuation; PostgreSQL rollback and neighboring race suites pass. [Offline reconciliation](work/phases/OFFLINE-RECONCILIATION-DETAIL_DESIGN.md) fixes pending rows merged by mutation ID instead of entity ID and non-transaction commands rendered as transactions. Full release baseline found 79/81 browser passes; follow-up validation remains in progress. No deploy.

- New owner direction 2026-09-09: approved “Sổ tiền” monochrome editorial concept and the order finish functions/tests/docs first, then replace all UI using shared base components. Owner subsequently requested “Imple all đi”. [F1 local proof](work/test-verification/SO-TIEN-F1-2026-09-10.md) is in_review: budget/HCM date and async lifecycle corrections pass 138 frontend tests, build and 3 desktop E2E journeys. F2 contracts, full test/device/provider gates and the later UI replacement remain open; no deployment occurred in this run.

- Active business-logic/API audit 2026-09-10: [evidence](work/test-verification/BUSINESS-LOGIC-AUDIT-2026-09-10.md) and the repo-owned [agent skill](../.agents/skills/auditing-mypocket-business-logic/SKILL.md) cover finance, transfer, reports, planning, sync, API-key/Redis authorization and ownership. Local corrections include checked arithmetic, exact reversals, receipt/idempotency preservation, HCM dates, wallet/transaction/planning/category optimistic versions, repayment locking/direction, Redis-stale revoke safety and sync race-to-conflict handling. Atomic sync is implemented locally with rollback/replay and lost-response browser proof; migration 0012 and full client resync are required on upgrade. Broader device/provider gates remain open; no new deployment is implied.

- Local functional run 2026-09-09 after the production blank-page report: [evidence and remaining inventory](work/test-verification/R0-FUNCTIONAL-E2E-2026-09-09.md). Empty-wallet null crash, zero transfer display, mounted transfer entry and five report controls corrected locally. Frontend 120 pass; Go/PostgreSQL suite pass; full browser 61 pass/2 retained WebKit offline failures. New accounting cases 9/9 pass within that total. No deployment; current public web still lacks these changes. Docusaurus completeness audit found missing workflow guides/per-action API–UI–release mapping/full machine-readable contract; public docs not updated in this run. Whole app/R0 remains in_progress.

- Redeployment 2026-09-09 21:25 Vietnam: owner requested deploy again. Same web/docs images recreated, API healthy, public HTTPS readiness/database and new web assets verified. Public docs correctly redirect unauthenticated requests to login; authenticated rendering remains untested. Both images now pushed to registry and remote manifest config digests verified. [Evidence](work/test-verification/R0-WEB-DOCS-RELEASE-2026-09-09.md). Worker/DB/config unchanged; earlier public/SSH blockers below are historical, and no infrastructure repair is claimed. Device/UAT gates remain open.

- Public-path follow-up 2026-09-09: [diagnosis](work/test-verification/R0-WEB-DOCS-RELEASE-2026-09-09.md#follow-up-public-path-diagnosis--2026-09-09-2054-vietnam) confirms local/LAN web and database readiness, but documented Pi5 TLS port 443 refuses connections while SSH port 22 is reachable. Live edge inspection blocked because the configured 1Password agent cannot sign the Pi key. No fixes or restarts; next input is unlock/restore existing SSH access, then inspect Caddy/Rathole. Do not infer a stopped service from the port check alone.

- Deployment 2026-09-09: owner requested web/docs release. [Production evidence](work/test-verification/R0-WEB-DOCS-RELEASE-2026-09-09.md): web `prod-2026.09.09.1`, API docs-only `prod-2026.09.09.1-docs` running locally; API executables, worker and DB unchanged. Frontend 100 pass, both image builds pass, local readiness/database and new assets pass. Docusaurus updated at `/docs/`; public HTTPS and registry connections close from this host, so external access and image publication remain unresolved. Physical Safari/PWA/receipt gates still open. Historical “no deployment” entries below refer to those earlier runs.

- Latest implementation, 2026-09-09: [search feedback](work/test-verification/R0-SEARCH-FEEDBACK-2026-09-09.md) is in_review after explicit owner approval. Shared loading/error/retry/clear controls retain query and reject obsolete search responses; offline state no longer claims an empty result. Frontend 100 pass, targeted search browsers 6 pass (Chrome mobile/desktop + WebKit). No API/DB/deploy changes; full R0 and physical receipt/PWA acceptance remain open.

- Latest investigation, 2026-09-09: [receipt readback follow-up](work/test-verification/R0-RECEIPT-READBACK-2026-09-09.md) is in_review after owner approval. Minimal same-record readback passes/fails/passes as WebKit offline emulation is toggled; no input-reset or storage migration fix is justified. Separate HTTP-outage receipt proof passes on all three projects. Frontend 92 pass, full Chrome 32 pass, targeted receipt 8 pass/1 retained emulation failure. No app-code changes or deployment; physical Safari/PWA/provider acceptance remains blocked. The earlier stop/hypothesis below is historical.

- Latest continuation, 2026-09-09: [R0 receipt-controls slice](work/test-verification/R0-RECEIPT-CONTROLS-2026-09-09.md) is **blocked on WebKit stored-image readback**. Shared picker/footer, persistent selection/removal, debt guard and truthful unavailable controls are implemented; 92 frontend tests pass, receipt browser checks 5 pass/1 fail. WebKit picker works, but offline stored bytes throw `NotReadableError` even with a disposable persistent profile. Stopped corrective work under the repo loop guard; focused file-lifetime/storage review is required before more fixes. No deployment; full R0 remains in_progress.

- Latest implementation, 2026-09-09: current UI retained; [R0 error-feedback slice](work/test-verification/R0-ERROR-FEEDBACK-2026-09-09.md) is in review, **full R0 remains in_progress**. Add/edit/archive transaction errors preserve input; API-key lifecycle feedback preserves confirmed results and rejects stale list/copy completions. Frontend 85 pass, Chrome mobile/desktop 26 pass. Minimal-worker diagnostics reproduce WebKit offline-emulation failure while server-down cache reload passes; physical Safari remains unverified. No deployment. Next: receipt footer/unsupported controls and remaining R0 error paths, then finance/report work.

- Latest audit, 2026-09-08: [Money Lover parity audit](research/moneylover/PARITY-AUDIT-2026-09-08.md) is complete as a docs/code comparison, **not a product completion claim**. 96 public Help Center articles are mapped to 50 assessment rows. Fresh proof: 139 backend test pass events with isolated PostgreSQL (2 Redis/S3 skips), 60 frontend tests and build pass; mobile evidence has 7 distinct passes across two runs and 2 unresolved failures (PWA overlay obstructs logout; budget name differs from test/user-entered name). Unmounted prototype screens, incomplete draft confirmation, report drilldowns, category-parent creation and other parity gaps supersede any inference of full UI completion from the historical statuses below. No application changes/deploy in this audit.

- Status: TICKET-001 through TICKET-017 are implemented and in review after automated package/component/build proof. The owner approved TICKET-027 asset portfolio valuation as a PHASE-005 extension immediately after TICKET-017; backend schema/domain/repository, REST API, dashboard total fields, Account-tab asset UI, IndexedDB asset cache/outbox, sync asset replay, and static provider refresh worker are implemented and in review after automated proof. TICKET-027 still needs human UAT. TICKET-018 through TICKET-026 remain deferred, and prod-hardening/audit logging is separated from TICKET-027.
- Active backlog: [BL-009](work/BACKLOG.md)
- Current queue focus: TICKET-028 production-hardening implementation is complete and in review; human UAT and deployment configuration remain.
- Active phase: [PHASE-007 Production Hardening and Debug Audit](work/phases/PHASE-007-production-hardening-debug-audit-detail-design.md), status `in_review`.
- Active ticket: [TICKET-028](work/tickets/TICKET-028-production-hardening-debug-audit-foundation.md), status `in_review`.
- Active bug: None.

## Current Focus

Complete the active business-logic/API audit and its release gates before further UI replacement or deployment.

## Recently Touched Areas

- `docs/requirements/`
- `docs/architecture/`
- `docs/decisions/`
- `docs/work/`
- `docs/work/tickets/TICKET-001-repository-runtime-foundation.md`
- `docs/work/tickets/TICKET-002-postgresql-migrations-s3-platform-adapters.md`
- `docs/work/tickets/TICKET-003-google-oauth-user-isolation.md`
- `docs/work/tickets/TICKET-004-development-operations-verification-baseline.md`
- `docs/work/phases/PHASE-001-detail-design.md`
- `docs/work/phases/PHASE-002-finance-core.md`
- `docs/work/phases/PHASE-002-detail-design.md`
- `docs/work/tickets/TICKET-005-wallet-category-domain.md`
- `docs/work/tickets/TICKET-006-transaction-accounting-engine.md`
- `docs/work/tickets/TICKET-007-vietnamese-seeds-receipt-metadata.md`
- `docs/work/tickets/TICKET-008-indexeddb-mirror-outbox.md`
- `docs/work/tickets/TICKET-009-sync-api-change-feed.md`
- `docs/work/tickets/TICKET-010-conflict-inbox-recovery.md`
- `docs/work/tickets/TICKET-011-budgets-threshold-alerts.md`
- `docs/work/tickets/TICKET-012-events-debts-repayments.md`
- `docs/work/tickets/TICKET-013-recurring-schedules-worker.md`
- `docs/work/tickets/TICKET-014-inbox-web-push.md`
- `docs/work/tickets/TICKET-015-pwa-navigation-search-wallet-views.md`
- `docs/work/tickets/TICKET-016-overview-net-worth-dashboard.md`
- `docs/work/tickets/TICKET-017-analytics-reports-cumulative-trends.md`
- `docs/work/tickets/TICKET-027-asset-portfolio-valuation.md`
- `docs/work/phases/PHASE-005-asset-portfolio-detail-design.md`
- `docs/superpowers/plans/2026-08-31-ticket-027-asset-portfolio-valuation.md`
- `docs/decisions/ADR-006-separate-asset-portfolio-valuation.md`
- `docs/work/tickets/TICKET-018-shared-drafts-text-ai-chat.md`
- `docs/work/tickets/TICKET-019-receipt-capture-ocr-adapter.md`
- `docs/work/tickets/TICKET-020-multimodal-image-chat.md`
- `docs/work/tickets/TICKET-021-signed-bank-webhook.md`
- `docs/work/tickets/TICKET-022-audit-pipeline-hidden-viewer.md`
- `docs/work/tickets/TICKET-023-manual-export-jobs.md`
- `docs/work/tickets/TICKET-024-account-reset-deletion.md`
- `docs/work/tickets/TICKET-025-homelab-production-release-proof.md`
- `docs/superpowers/plans/2026-08-30-phase-002-finance-core.md`
- `docs/superpowers/plans/2026-08-30-phase-003-offline-sync.md`
- `docs/superpowers/plans/2026-08-30-phase-004-planning-automation.md`
- `docs/superpowers/plans/2026-08-30-phase-005-analytics-dashboard.md`
- `docs/superpowers/plans/2026-08-30-phase-006-ingestion.md`
- `docs/superpowers/plans/2026-08-30-phase-007-production.md`
- `docs/superpowers/plans/2026-08-25-phase-001-platform-identity.md`
- `docs/superpowers/specs/2026-08-23-mypocket-system-design.md`
- `docs/work/DOCS-REVIEW-M0.md`
- `backend/internal/platform/db/`
- `backend/migrations/0001_phase001_identity.sql`
- `backend/migrations/0002_phase002_finance_core.sql`
- `backend/internal/finance/`
- `backend/internal/platform/httpapi/finance.go`
- `backend/internal/platform/httpapi/finance_test.go`
- `frontend/src/app/finance.ts`
- `frontend/src/app/App.tsx`
- `frontend/src/app/App.test.tsx`
- `frontend/src/app/components.tsx`
- `frontend/src/app/outbox.ts`
- `frontend/src/app/outbox.test.ts`
- `frontend/src/offline/`
- `frontend/src/test/setup.ts`
- `frontend/src/styles.css`
- `frontend/vite.config.ts`
- `frontend/package.json`
- `frontend/package-lock.json`
- `docs/work/test-verification/PHASE-002-finance-core.md`
- `docs/work/test-verification/PHASE-003-offline-sync.md`
- `docs/work/test-verification/PHASE-004-planning-automation.md`
- `docs/work/test-verification/PHASE-005-analytics-dashboard.md`
- `docs/work/test-verification/PHASE-006-ai-receipt-bank-ingestion.md`
- `docs/work/test-verification/PHASE-007-audit-export-production.md`
- `backend/internal/platform/objectstore/`
- `backend/internal/identity/`
- `backend/internal/platform/httpapi/auth.go`
- `backend/internal/platform/httpapi/router.go`
- `frontend/`
- `compose.yaml`
- `backend/Dockerfile`
- `frontend/Dockerfile`
- `frontend/nginx.conf`
- `frontend/e2e/auth-shell.spec.ts`
- `frontend/e2e/offline-sync.spec.ts`
- `scripts/smoke-platform.sh`
- `design/DESIGN.md`
- `backend/cmd/api/main.go`
- `backend/cmd/worker/main.go`
- `docs/architecture/API.md`
- `docs/architecture/ERD.md`
- `backend/internal/sync/`
- `backend/internal/platform/httpapi/sync.go`
- `backend/migrations/0003_phase003_sync.sql`
- `frontend/src/offline/syncApi.ts`
- `frontend/src/offline/conflicts.ts`
- `frontend/src/offline/conflicts.test.ts`
- `backend/internal/planning/`
- `backend/internal/platform/httpapi/planning.go`
- `backend/migrations/0004_phase004_planning_budgets.sql`
- `backend/migrations/0005_phase004_events_obligations.sql`
- `backend/migrations/0006_phase004_recurring_schedules.sql`
- `backend/migrations/0007_phase004_notifications.sql`
- `backend/internal/notification/`
- `backend/internal/analytics/`
- `frontend/src/app/notifications.ts`
- `frontend/src/app/analytics.ts`
- `backend/internal/worker/`
- `frontend/src/app/planning.ts`
- `frontend/e2e/planning-automation.spec.ts`

## Recent Decisions

- Use React + TypeScript with Vite, React Router, TanStack Query, PWA service worker, and IndexedDB.
- Use a Go modular monolith exposed as separate API and worker processes sharing domain/application packages.
- Use PostgreSQL as source of truth and private S3-compatible storage for receipts/exports.
- Use Google OAuth with a signed stateless cookie, no session table, and no linked-device management.
- Require explicit review before AI, OCR, webhook, or recurring drafts affect wallet balances.
- Use optimistic versions and explicit conflict review for full offline read/write.
- Keep receipt OCR and AI-chat images as distinct user flows.
- Restrict a hidden audit page to `AUDIT_VIEWER_EMAIL`; default retention is 180 days.
- Defer TICKET-018 through TICKET-026; explicitly promote TICKET-027 asset portfolio valuation immediately after TICKET-017 as the only current-release exception.
- Defer voice input until post-M1 draft infrastructure is verified.
- User delegated remaining implementation decisions on 2026-08-30; PHASE-003 through PHASE-007 now have approved detail designs and component contracts.
- PHASE-003 through PHASE-005 remain in the current release queue; PHASE-006 and PHASE-007 keep their ready design artifacts but are deferred to M2.
- TICKET-008 uses `frontend/src/offline/` as the IndexedDB boundary for wallets, categories, transactions, outbox mutations, tombstones, conflicts, and sync meta; `frontend/src/app/outbox.ts` remains a compatibility wrapper for the previous localStorage-first transaction outbox.
- Offline reload uses the last successful `/api/v1/me` envelope after a transport failure only; explicit authentication errors never fall back to cached identity. Financial `localStorage` caches are namespaced by authenticated user and cleared at logout. IndexedDB uses a separate database per user, selected before data-loading/replay effects start. Switching accounts preserves the previous owner's pending work without exposing it; logout clears the active owner's local stores. Legacy unowned offline data is not assigned to the next account.
- TICKET-009 preserves client UUIDs for offline-created wallets, categories, and transactions so optimistic IndexedDB rows reconcile to the same authoritative IDs after reconnect.
- TICKET-010 keeps conflict review local to the PWA for M1: conflict responses are persisted in IndexedDB, the failed mutation remains recoverable until the user chooses an action, keep-server/discard-local remove the pending local intent, edit-and-retry queues a replacement mutation against the authoritative server version, and full resync refreshes the finance mirror without clearing recoverable outbox items.
- TICKET-011 implements budgets as online-authenticated planning records; web/PWA can view progress, but create/edit/archive requires online API access. Threshold notices are deduped in `budget_alerts` by `(budget_id, threshold, period_start)`.
- TICKET-012 implements events and obligations as online-authenticated planning records; event/debt screens are available in the Ngân sách tab, links only attach confirmed owned transactions, event totals exclude report-excluded transactions, and obligation repayments reject totals above principal.
- TICKET-013 implements recurring schedules and transaction drafts as online-authenticated planning records; the worker acquires a recurring lease before processing, occurrence keys are deterministic, generated drafts do not change balances, and the mobile planning tab shows schedule setup plus pending draft review rows.
- TICKET-014 implements a durable notification inbox, push subscription lifecycle, capped delivery retry, expiry cleanup, and mobile permission/offline states.
- TICKET-015 through TICKET-017 implement authenticated search/wallet detail, server-side dashboard totals, normalized analytics reports, category roll-up, daily/cumulative series, comparison baseline handling, and device privacy masking.
- Beta PWA hardening now includes standalone/Apple metadata and a transport-failure cached-auth fallback: a service-worker-cached navigation may report `navigator.onLine=true` while the API is unavailable, so cached authenticated identity is used only after a transport failure and never to bypass a `FORBIDDEN` response. Mobile Playwright offline reload passes.
- Frontend styling now uses Tailwind CSS v4 via the Vite plugin with MyPocket theme tokens in `frontend/src/styles.css` and reusable base primitives in `frontend/src/app/components.tsx`; current migration preserves stable class names while moving new/iterated UI toward shared primitives.
- Runtime fix: Service Worker now ignores non-HTTP(S) requests (including `chrome-extension://`), and auth payloads are validated before rendering to prevent blank-page crashes.
- OAuth runtime: local Vite proxies `/api` to the Go API; backend supports real Google authorization-code exchange when `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, and `GOOGLE_REDIRECT_URL` are configured, while fixture mode remains the automated-test path.
- Money Insider research is captured in [PHASE-005 Money Insider detail design](work/phases/PHASE-005-money-insider-detail-design.md): replace the hardcoded Home card with authenticated category-frequency, daily-average, prior-period, six-period, budget, projection, and top-expense reporting under TICKET-017; subscription/paywall UI is excluded.
- Owner decision: this is a personal application and all Money Insider functionality is permanently unlocked; do not add trial, premium, subscription, registration, entitlement, or payment gates.
- TICKET-027 keeps market-valued assets separate from wallet accounting. D-01 through D-04 are approved: hybrid automatic/manual pricing, moving weighted-average buy/sell accounting, separate wallet/investment/combined totals, and offline manual mutations with server-online price refresh. Implemented: migration `0008_phase005_asset_portfolio.sql`, package `backend/internal/portfolio`, REST routes under `/api/v1/assets` and `/api/v1/portfolio/summary`, dashboard investment/combined total fields, Account-tab asset UI with manual/automatic pricing mode, IndexedDB asset cache/outbox/tombstones/conflict resync, sync entity replay for asset create/trade/price/archive, and static-price provider refresh worker using `PORTFOLIO_STATIC_PRICES_JSON`. Real PostgreSQL proof covers ownership, ledger replay, manual price history, archive retention, sync resync/mutation ledger, migration chain, and wallet non-interference.
- Auth runtime now supports `ALLOWED_LOGIN_EMAILS` as a comma-separated allowlist; local dev is set to `dungbui.dungbui.00@gmail.com`. Empty allowlist preserves prior login behavior.
- Production-hardening TICKET-028 is in review with automated proof complete. Implemented foundation pieces include structured request logs, panic recovery, `audit_events`, hidden `/api/v1/audit/events` plus `/api/v1/audit/access`, audit retention worker, Account-tab API key management plus `/api/v1/api-keys`, bearer `mpk_...` auth for third-party/AI agents, Redis hints with PostgreSQL validation on every Bearer request, revoke invalidation, user-scoped receipt metadata with S3-compatible presigned upload/download, transaction receipt attachment, durable IndexedDB receipt retry queue, and a whitelist-only audit viewer panel. Production requires `AUDIT_VIEWER_EMAIL`, `AUDIT_HASH_SECRET`, `API_KEY_HASH_SECRET`, `REDIS_URL`, S3 settings, HTTPS `PUBLIC_WEB_URL`, and `LOG_FORMAT=json`.
- Category catalog migration `0011_phase002_category_catalog.sql` expands default system categories into Vietnamese parent/child groups for expense, income, and debt while preserving legacy system keys used by existing transactions.

## Queue Summary

- BL-001 through BL-005 comprise the current M1 release in dependency order; BL-005 now includes TICKET-027 after TICKET-017 and is ready for human UAT.
- BL-008 is deferred voice input.
- BL-006 and BL-007 are deferred to M2 after TICKET-027; BL-008 is deferred after M2.
- PHASE-001 through PHASE-005 have automated implementation proof through TICKET-017 pending human review. TICKET-027 extends PHASE-005 and has backend/API/dashboard/sync/offline/provider-worker/frontend build evidence; human UAT remains pending. Browser visual proof is blocked by local Playwright/Chrome headless availability, not by automated app tests.

## Next Steps

- Atomic-sync implementation is locally in review: see [design](work/phases/ATOMIC-SYNC-DETAIL_DESIGN.md) and [test evidence](work/test-verification/BUSINESS-LOGIC-AUDIT-2026-09-10.md). Next release integration must apply migration 0012, arrange full client resync for historical feed gaps, and retain existing physical-device/provider gates. Continue broader business audit without treating this slice as whole-app completion.

- Search feedback is now [in review](work/test-verification/R0-SEARCH-FEEDBACK-2026-09-09.md). Next bounded R0 slice: notification mark-read errors, then wallet-detail/Insider and planning/audit errors; preserve confirmed results and use shared feedback. Existing cache-freshness and full search/navigation parity remain separate from this fix.

- Focused receipt investigation is now [in review](work/test-verification/R0-RECEIPT-READBACK-2026-09-09.md): the same stored File reads pass/fail/pass as WebKit offline emulation is toggled, without rewriting. Separate HTTP-outage receipt proof passes on Chrome and WebKit; original emulation test stays failing (targeted total 8 pass/1 fail). No storage or application-code changes. Physical Safari/PWA and provider acceptance remain blocked; next functional work can continue with remaining R0 error paths without claiming receipt release readiness.

- Receipt release gate: physical Safari/PWA and provider UAT following the [focused investigation](work/test-verification/R0-RECEIPT-READBACK-2026-09-09.md). Owner approved the investigation; storage changes are not justified by the measured emulation-only differential. Then continue remaining planning/search/audit/notification errors and finance/report journeys.

- 2026-09-09 owner direction: **keep current UI, prioritize functional correctness**; approved sequential plan and implementation. [R0 design](work/phases/R0-DETAIL_DESIGN.md), [first-slice verification](work/test-verification/R0-FUNCTIONAL-STABILITY-2026-09-09.md), [mounted-action inventory](work/test-verification/R0-MOUNTED-ACTIONS-2026-09-09.md). Local PWA/logout clearance and budget naming/zero-data/keyboard fixes are in review; full R0 is not done. Final frontend tests 72 pass, Chrome mobile/desktop 22 pass. Earlier expanded suite had WebKit 10 pass/1 fail; offline reload internal error reproduced unchanged and remains a release risk, not a diagnosed engine-only defect. No deployment. Next: isolate that failure and complete existing write/API-key error feedback and inert-action handling before advancing finance/report functionality.

- User's current focus (2026-09-08): use the [parity matrix](research/moneylover/FEATURE-PARITY-2026-09-08.md) and R0–R5 remediation in the audit as the basis for approaching Money Lover. First address the reproduced mobile regression and docs/API truth, then close functional journeys using shared base components. Do not equate research/docs completeness or passing package tests with acceptance; prior exclusions/deferred implementation need explicit reconciliation in the relevant design.

- Reliability verification is recorded in [BETA-RELIABILITY](work/test-verification/BETA-RELIABILITY.md); strict TypeScript is now a build gate, browser E2E uses a dedicated database and fresh test servers. The mounted Overview uses API-derived daily bars, logout preserves user-owned IndexedDB pending work, and bearer keys have wallet-route HTTP proof. Production provider/device acceptance remains separate from local automated proof.

1. Run human UAT for TICKET-027 on direct web/PWA using one gold, one stock, and one crypto fixture.
2. Run human UAT for TICKET-028 production-hardening: real login, API key create/list/revoke, bearer `/api/v1/me`, receipt upload/download on configured S3, offline receipt retry, hidden audit access, and correlation-ID lookup.
3. Run remaining mobile/desktop UAT for TICKET-001 through TICKET-017.
4. Mark PHASE-004 and PHASE-005 verified only after the promoted TICKET-027 UAT is accepted.
5. Keep TICKET-018 through TICKET-026 deferred unless the user separately promotes them.

## Open Questions

These values are deployment inputs, not architecture blockers:

- Production `DATABASE_URL` and public application URL.
- Google OAuth client ID, secret, and redirect URL.
- S3 endpoint, bucket, region, and credentials.
- OpenAI-compatible base URL, text model, vision model, and key.
- OCR API base URL, authentication method, and response schema.
- Web Push VAPID keys.
- Verified Google email used for `AUDIT_VIEWER_EMAIL`.
