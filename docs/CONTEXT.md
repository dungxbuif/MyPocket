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
updated: 2026-08-26
---

# Project Context

## Field Ownership

- Human owns product intent, priority overrides, and unresolved product decisions.
- AI owns concise implementation-state refreshes, evidence links, and next-step recommendations.
- This context does not silently override approved scope in the product spec or phase files.

## Current Status

- Status: PHASE-002 finance-core execution is underway after PHASE-001 reached review. Finance schema, Vietnamese seed categories, backend wallet/category repository behavior, initial wallet/category API routes, API-driven mobile wallet/category display, transaction accounting effect validation, repository create-transaction atomic/idempotent behavior, edit/archive reversal, and transaction search filters are implemented with automated proof.
- Active backlog: [BL-002](work/BACKLOG.md)
- Current queue focus: execute PHASE-002 finance core; TICKET-005 wallet/category foundation is partially implemented and TICKET-006 transaction accounting has started.
- Active phase: [PHASE-002 Finance Core](work/phases/PHASE-002-finance-core.md), status `ready`.
- Active ticket: [TICKET-006](work/tickets/TICKET-006-transaction-accounting-engine.md), status `in_progress`.
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
- `docs/superpowers/plans/2026-08-30-phase-002-finance-core.md`
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

## Queue Summary

- BL-001 through BL-007 comprise the initial release in dependency order.
- BL-008 is deferred voice input.
- All product requirements are accepted but remain unimplemented with planned proof only.

## Next Steps

1. Add `/api/v1/transactions` HTTP handlers for create, edit/archive, and list/search using the verified repository behavior.
2. Complete remaining TICKET-005 wallet/category mutation API routes and mobile create/edit controls before PHASE-002 review.
3. Add mobile transaction add/search/edit flows and PHASE-002 UAT proof.

## Open Questions

These values are deployment inputs, not architecture blockers:

- Production `DATABASE_URL` and public application URL.
- Google OAuth client ID, secret, and redirect URL.
- S3 endpoint, bucket, region, and credentials.
- OpenAI-compatible base URL, text model, vision model, and key.
- OCR API base URL, authentication method, and response schema.
- Web Push VAPID keys.
- Verified Google email used for `AUDIT_VIEWER_EMAIL`.
