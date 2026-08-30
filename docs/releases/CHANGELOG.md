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

### Planning

- Approved and recorded the MyPocket React PWA and Go modular-monolith system design.
- Migrated the source SRS into Harness requirements, user stories, architecture, API, ERD, integrations, SDD, ADRs, roadmap, eight phases, backlog, traceability, and validation planning.
- Defined the seven-phase initial release and deferred voice-input phase.
- Recorded full offline conflict review, review-first AI/OCR/webhook ingestion, Google OAuth without server sessions, and restricted 180-day-default audit logging.
- Created PHASE-001 ticket artifacts, detail design, and executable implementation plan focused on the first deployable PWA/mobile shell, platform, identity, and offline app-shell cache slice.
- Created PHASE-002 finance-core tickets, detail design, and executable implementation plan for wallet/category, transaction accounting, and Vietnamese seed/receipt metadata slices.
