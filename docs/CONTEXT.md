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

- Status: TICKET-001 through TICKET-017 are implemented and in review after automated package/component/build proof. The owner approved TICKET-027 asset portfolio valuation as a PHASE-005 extension immediately after TICKET-017; backend schema/domain/repository, REST API, dashboard total fields, Account-tab asset UI, IndexedDB asset cache/outbox, sync asset replay, and static provider refresh worker are implemented and in review after automated proof. TICKET-027 still needs human UAT. TICKET-018 through TICKET-026 remain deferred, and prod-hardening/audit logging is separated from TICKET-027.
- Active backlog: [BL-009](work/BACKLOG.md)
- Current queue focus: TICKET-028 production-hardening implementation is complete and in review; human UAT and deployment configuration remain.
- Active phase: [PHASE-007 Production Hardening and Debug Audit](work/phases/PHASE-007-production-hardening-debug-audit-detail-design.md), status `in_review`.
- Active ticket: [TICKET-028](work/tickets/TICKET-028-production-hardening-debug-audit-foundation.md), status `in_review`.
- Active bug: None.

## Current Focus

Execute PHASE-004 planning and automation after PHASE-003 reached review.

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
- Offline reload for authenticated cached data now uses the last successful `/api/v1/me` envelope only when `navigator.onLine` is false; logout clears both this cached user and IndexedDB offline stores.
- TICKET-009 preserves client UUIDs for offline-created wallets, categories, and transactions so optimistic IndexedDB rows reconcile to the same authoritative IDs after reconnect.
- TICKET-010 keeps conflict review local to the PWA for M1: conflict responses are persisted in IndexedDB, the failed mutation remains recoverable until the user chooses an action, keep-server/discard-local remove the pending local intent, edit-and-retry queues a replacement mutation against the authoritative server version, and full resync refreshes the finance mirror without clearing recoverable outbox items.
- TICKET-011 implements budgets as online-authenticated planning records; web/PWA can view progress, but create/edit/archive requires online API access. Threshold notices are deduped in `budget_alerts` by `(budget_id, threshold, period_start)`.
- TICKET-012 implements events and obligations as online-authenticated planning records; event/debt screens are available in the Ngân sách tab, links only attach confirmed owned transactions, event totals exclude report-excluded transactions, and obligation repayments reject totals above principal.
- TICKET-013 implements recurring schedules and transaction drafts as online-authenticated planning records; the worker acquires a recurring lease before processing, occurrence keys are deterministic, generated drafts do not change balances, and the mobile planning tab shows schedule setup plus pending draft review rows.
- TICKET-014 implements a durable notification inbox, push subscription lifecycle, capped delivery retry, expiry cleanup, and mobile permission/offline states.
- TICKET-015 through TICKET-017 implement authenticated search/wallet detail, server-side dashboard totals, normalized analytics reports, category roll-up, daily/cumulative series, comparison baseline handling, and device privacy masking.
- Frontend styling now uses Tailwind CSS v4 via the Vite plugin with MyPocket theme tokens in `frontend/src/styles.css` and reusable base primitives in `frontend/src/app/components.tsx`; current migration preserves stable class names while moving new/iterated UI toward shared primitives.
- Runtime fix: Service Worker now ignores non-HTTP(S) requests (including `chrome-extension://`), and auth payloads are validated before rendering to prevent blank-page crashes.
- OAuth runtime: local Vite proxies `/api` to the Go API; backend supports real Google authorization-code exchange when `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, and `GOOGLE_REDIRECT_URL` are configured, while fixture mode remains the automated-test path.
- Money Insider research is captured in [PHASE-005 Money Insider detail design](work/phases/PHASE-005-money-insider-detail-design.md): replace the hardcoded Home card with authenticated category-frequency, daily-average, prior-period, six-period, budget, projection, and top-expense reporting under TICKET-017; subscription/paywall UI is excluded.
- Owner decision: this is a personal application and all Money Insider functionality is permanently unlocked; do not add trial, premium, subscription, registration, entitlement, or payment gates.
- TICKET-027 keeps market-valued assets separate from wallet accounting. D-01 through D-04 are approved: hybrid automatic/manual pricing, moving weighted-average buy/sell accounting, separate wallet/investment/combined totals, and offline manual mutations with server-online price refresh. Implemented: migration `0008_phase005_asset_portfolio.sql`, package `backend/internal/portfolio`, REST routes under `/api/v1/assets` and `/api/v1/portfolio/summary`, dashboard investment/combined total fields, Account-tab asset UI with manual/automatic pricing mode, IndexedDB asset cache/outbox/tombstones/conflict resync, sync entity replay for asset create/trade/price/archive, and static-price provider refresh worker using `PORTFOLIO_STATIC_PRICES_JSON`. Real PostgreSQL proof covers ownership, ledger replay, manual price history, archive retention, sync resync/mutation ledger, migration chain, and wallet non-interference.
- Auth runtime now supports `ALLOWED_LOGIN_EMAILS` as a comma-separated allowlist; local dev is set to `dungbui.dungbui.00@gmail.com`. Empty allowlist preserves prior login behavior.
- Production-hardening TICKET-028 is in review with automated proof complete. Implemented foundation pieces include structured request logs, panic recovery, `audit_events`, hidden `/api/v1/audit/events` plus `/api/v1/audit/access`, audit retention worker, Account-tab API key management plus `/api/v1/api-keys`, bearer `mpk_...` auth for third-party/AI agents, Redis cache for API key/session token lookups with PostgreSQL fallback and immediate revoke invalidation, user-scoped receipt metadata with S3-compatible presigned upload/download, transaction receipt attachment, durable IndexedDB receipt retry queue, and a whitelist-only audit viewer panel. Production requires `AUDIT_VIEWER_EMAIL`, `AUDIT_HASH_SECRET`, `API_KEY_HASH_SECRET`, `REDIS_URL`, S3 settings, HTTPS `PUBLIC_WEB_URL`, and `LOG_FORMAT=json`.
- Category catalog migration `0011_phase002_category_catalog.sql` expands default system categories into Vietnamese parent/child groups for expense, income, and debt while preserving legacy system keys used by existing transactions.

## Queue Summary

- BL-001 through BL-005 comprise the current M1 release in dependency order; BL-005 now includes TICKET-027 after TICKET-017 and is ready for human UAT.
- BL-008 is deferred voice input.
- BL-006 and BL-007 are deferred to M2 after TICKET-027; BL-008 is deferred after M2.
- PHASE-001 through PHASE-005 have automated implementation proof through TICKET-017 pending human review. TICKET-027 extends PHASE-005 and has backend/API/dashboard/sync/offline/provider-worker/frontend build evidence; human UAT remains pending. Browser visual proof is blocked by local Playwright/Chrome headless availability, not by automated app tests.

## Next Steps

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
