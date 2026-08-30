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
updated: 2026-08-30
---

# Project Context

## Field Ownership

- Human owns product intent, priority overrides, and unresolved product decisions.
- AI owns concise implementation-state refreshes, evidence links, and next-step recommendations.
- This context does not silently override approved scope in the product spec or phase files.

## Current Status

- Status: PHASE-002 finance core is in review after automated implementation proof. Finance schema, Vietnamese seed categories, receipt metadata validation/persistence, backend wallet/category repository behavior, wallet/category mutation API routes, API-driven mobile wallet/category management, live mobile finance CRUD E2E, transaction accounting effect validation, repository create-transaction atomic/idempotent behavior, edit/archive reversal, transaction search filters, transaction HTTP routes, mobile transaction list/search/create/type/edit/archive UI, and a browser transaction outbox with reconnect drain are implemented with automated proof.
- Active backlog: [BL-003](work/BACKLOG.md)
- Current queue focus: PHASE-002 human/UAT review remains pending while PHASE-003 offline synchronization is ready for execution.
- Active phase: [PHASE-003 Offline Synchronization](work/phases/PHASE-003-offline-sync.md), status `ready`.
- Active ticket: [TICKET-008](work/tickets/TICKET-008-indexeddb-mirror-outbox.md), status `ready`.
- Active bug: None.

## Current Focus

Execute PHASE-002 as the finance core slice after PHASE-001 platform, identity, PWA shell, and local stack proof reached review.

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
- `frontend/src/styles.css`
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
- `scripts/smoke-platform.sh`
- `design/DESIGN.md`
- `backend/cmd/api/main.go`
- `backend/cmd/worker/main.go`
- `docs/architecture/API.md`
- `docs/architecture/ERD.md`

## Recent Decisions

- Use React + TypeScript with Vite, React Router, TanStack Query, PWA service worker, and IndexedDB.
- Use a Go modular monolith exposed as separate API and worker processes sharing domain/application packages.
- Use PostgreSQL as source of truth and private S3-compatible storage for receipts/exports.
- Use Google OAuth with a signed stateless cookie, no session table, and no linked-device management.
- Require explicit review before AI, OCR, webhook, or recurring drafts affect wallet balances.
- Use optimistic versions and explicit conflict review for full offline read/write.
- Keep receipt OCR and AI-chat images as distinct user flows.
- Restrict a hidden audit page to `AUDIT_VIEWER_EMAIL`; default retention is 180 days.
- Defer voice input until the seven initial release phases are verified.
- User delegated remaining implementation decisions on 2026-08-30; PHASE-003 through PHASE-007 now have approved detail designs and component contracts.
- PHASE-003 through PHASE-007 now also have ready ticket artifacts, planned verification targets, and executable implementation plans; no implementation evidence is claimed for those phases yet.

## Queue Summary

- BL-001 through BL-007 comprise the initial release in dependency order.
- BL-008 is deferred voice input.
- PHASE-001 and PHASE-002 have automated implementation proof pending human review; PHASE-003 through PHASE-007 are planned and ready for ordered execution.

## Next Steps

1. Execute TICKET-008 to replace the temporary localStorage-first outbox with the approved IndexedDB-primary offline store.
2. Execute TICKET-009 sync API/change-feed backend after the browser mirror and mutation envelope are stable.
3. Execute TICKET-010 conflict inbox/recovery, then run PHASE-003 E2E/UAT and reconcile docs.
4. Review PHASE-002 with the user/UAT criteria and mark verified if accepted.

## Open Questions

These values are deployment inputs, not architecture blockers:

- Production `DATABASE_URL` and public application URL.
- Google OAuth client ID, secret, and redirect URL.
- S3 endpoint, bucket, region, and credentials.
- OpenAI-compatible base URL, text model, vision model, and key.
- OCR API base URL, authentication method, and response schema.
- Web Push VAPID keys.
- Verified Google email used for `AUDIT_VIEWER_EMAIL`.
