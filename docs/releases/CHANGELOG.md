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

## Field Ownership

- Human owns release approval.
- AI maintains planned and released change entries with trace links.

## [Unreleased]

### Added

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
- Added the PHASE-003 frontend IndexedDB mirror/outbox foundation with legacy localStorage migration/quarantine, wallet/category/transaction offline mutation queueing, cached offline hydration, degraded read-only UI, and mobile PWA offline reload proof.

### Planning

- Approved executable detail designs for PHASE-003 through PHASE-007 and extended the shared mobile component design contract through the full M1 release.
- Approved and recorded the MyPocket React PWA and Go modular-monolith system design.
- Migrated the source SRS into Harness requirements, user stories, architecture, API, ERD, integrations, SDD, ADRs, roadmap, eight phases, backlog, traceability, and validation planning.
- Defined the seven-phase initial release and deferred voice-input phase.
- Recorded full offline conflict review, review-first AI/OCR/webhook ingestion, Google OAuth without server sessions, and restricted 180-day-default audit logging.
- Created PHASE-001 ticket artifacts, detail design, and executable implementation plan focused on the first deployable PWA/mobile shell, platform, identity, and offline app-shell cache slice.
- Created PHASE-002 finance-core tickets, detail design, and executable implementation plan for wallet/category, transaction accounting, and Vietnamese seed/receipt metadata slices.
- Created PHASE-003 offline synchronization tickets, verification target, and executable implementation plan for IndexedDB mirror/outbox, idempotent sync APIs, change feed, conflict inbox, and full resync.
- Created PHASE-004 through PHASE-007 ticket artifacts, verification targets, and executable implementation plans for planning automation, analytics/dashboard, AI/OCR/webhook ingestion, audit/export/account lifecycle, and production release proof.
