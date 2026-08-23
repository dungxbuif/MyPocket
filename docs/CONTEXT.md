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
updated: 2026-08-24
---

# Project Context

## Field Ownership

- Human owns product intent, priority overrides, and unresolved product decisions.
- AI owns concise implementation-state refreshes, evidence links, and next-step recommendations.
- This context does not silently override approved scope in the product spec or phase files.

## Current Status

- Status: MyPocket system design and Harness migration verified; no product code exists yet.
- Active backlog: [BL-001](work/BACKLOG.md)
- Current queue focus: create ticket artifacts and a detailed executable plan for PHASE-001.
- Active phase: [PHASE-001 Platform and Identity](work/phases/PHASE-001-platform-identity.md), status `draft`.
- Active ticket: None; stable planned ticket IDs are TICKET-001 through TICKET-004.
- Active bug: None.

## Current Focus

Finish planning PHASE-001 as the first deployable slice of the approved React PWA, Go API/worker, PostgreSQL, S3-compatible, and Google OAuth architecture.

## Recently Touched Areas

- `docs/requirements/`
- `docs/architecture/`
- `docs/decisions/`
- `docs/work/`
- `docs/superpowers/specs/2026-08-23-mypocket-system-design.md`
- `docs/work/DOCS-REVIEW-M0.md`

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

1. Create TICKET-001 through TICKET-004 for PHASE-001.
2. Produce the Superpowers implementation plan linked to those Harness tickets.
3. Review the plan and promote PHASE-001 from `draft` to `ready`.
4. Execute only after ticket/detail-design approval and required gates pass.

## Open Questions

These values are deployment inputs, not architecture blockers:

- Production `DATABASE_URL` and public application URL.
- Google OAuth client ID, secret, and redirect URL.
- S3 endpoint, bucket, region, and credentials.
- OpenAI-compatible base URL, text model, vision model, and key.
- OCR API base URL, authentication method, and response schema.
- Web Push VAPID keys.
- Verified Google email used for `AUDIT_VIEWER_EMAIL`.
