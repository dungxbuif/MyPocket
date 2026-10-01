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
- AI maintains change entries and links to release notes.

This file is the first-release baseline for the current MyPocket snapshot. Earlier
development notes remain available in Git history and are intentionally not presented
as active release entries here.

## [1.0.7] - 2026-10-01

### Added

- Travel Mode now has owner-scoped event CRUD, one active event per account,
  ordinary transaction auto-linking, and explicit transaction link/unlink.
  Deleting an event clears links without deleting ledger rows or changing
  report inclusion. Recurring, transfer, adjustment and AI approval paths do
  not auto-attach events.
- Credit entry categories now enforce applicable-wallet scope in both the
  handler and repository transaction.

### Verification

- `go test ./...`, `go generate ./cmd/api`, `npm run check:design`,
  `npm run typecheck`, `npm run test:design`, `npm run test:transactions`,
  `npm run test:calendar`, `npm run test:transaction-jars`, and `npm run build`
  pass locally. PostgreSQL integration tests remain environment-gated by
  `TEST_DATABASE_URL`.

### Deployment

- No production deployment is claimed by this local implementation commit.

## [1.0.6] - 2026-09-30

### Fixed

- Transaction ledger wallet scope now follows the supplied wallet-selector reference:
  the compact wallet capsule opens the shared `Chọn Ví` sheet with aggregate,
  included/excluded wallets, real balances, add-wallet and disabled-link states.
  Selecting a wallet keeps the existing account-local week/custom ledger behavior.

### Verification

- `npm run test:wallet-scope`, `npm run check:design`, `npm run typecheck`,
  `npm run test:design`, `npm run test:transactions`, `npm run test:calendar`,
  `npm run test:transaction-jars`, and `npm run build` pass locally.

### Deployment

- Production web service converged on
  `registry.dungxbuif.com/mypocket-web:feedback-ui-wallet-design-20260930-0700`
  (`sha256:933dd62630e3ee5ae7917c6ddaab36b3d858bc58f2fdb0abb87cc2dd5e36d80e`).
  Public HTML serves the new bundle and `/api/v1/health` returns 200.

## [1.0.4] - 2026-09-30

### Fixed

- AI transaction-entry proposals now normalize date-only model output (`dd/MM/yyyy`,
  `dd-MM-yyyy`, or `yyyy-MM-dd`) in the account timezone instead of silently replacing
  a recognized date with today; the missing-time assumption is shown for review.
- Exact duplicate proposals within one OCR/text submission are merged while preserving
  review questions from the duplicate rows.

### Deployment

- API image `registry.dungxbuif.com/mypocket-api:feedback-api-ai-date-20260930-0011`
  was rolled out to production and verified healthy.

## [1.0.5] - 2026-09-30

### Added

- Login and Google OAuth responses now include rotating refresh credentials;
  `POST /api/v1/auth/refresh` consumes a refresh token once and issues a new
  access/refresh pair.
- The web client retries one expired-session request through the refresh route
  and keeps API-key requests out of the refresh flow.

### Deployment

- API `feedback-api-refresh-20260930-0041` and web
  `feedback-ui-refresh-20260930-0032` were rolled out and their public health
  and invalid-refresh checks passed.

## [1.0.2] - 2026-09-30

### Fixed

- Increased the frontend API proxy read/send timeouts to 240 seconds so the
  bounded OCR/model processing window does not turn into an HTML 504 response.

## [1.0.3] - 2026-09-30

### Fixed

- Category search now promotes a parent group when the query matches a child
  category, so matching groups are visible at the top of the result.

## [1.0.1] - 2026-09-29

### Fixed

- Recent transactions on the overview now show the account-local date and time
  before wallet and note metadata.
- Feedback routes accept owner-scoped API keys with `feedback:read` and
  `feedback:write` scopes.

### Deployment

- Production API and web services were rolled out to the verified registry
  images after the backend and frontend checks passed.

## [1.0.0] - 2026-09-24

### Deployment note

- Homelab `money.dungxbuif.com` was redeployed on 2026-09-24 from the
  `feature/wallet-management` branch. Production PostgreSQL schema was reset
  and migrated to version 21; Redis DB 1 was flushed. A pre-reset PostgreSQL
  dump was kept on Pi5 under `/home/dungxbuif/backups/mypocket/`.
- Production now uses Docker images `homelab/mypocket-api:0ff84828-deploy-20260924015529`
  and `homelab/mypocket-web:0ff84828-deploy-20260924015529`. The old
  `mypocket-worker` Swarm service is scaled to zero because this code snapshot
  has no `cmd/worker` binary.
- Production AI entry, receipt OCR, and Finance Assistant are enabled through
  server-side environment variables. MyPocket reuses the shared homelab LLM and
  OCR credentials; no provider keys are exposed to the browser or committed docs.
- The production web proxy accepts the same 101 MiB AI-entry upload envelope as
  the backend so receipt images/PDFs are not rejected before OCR processing.
- The production client now normalizes account and AI-entry timezones before
  submission, falling back to `Asia/Ho_Chi_Minh` when a browser/webview reports
  an empty or non-IANA timezone such as `Local`.

### Added

- Authenticated wallet, category, transaction, budget, jar, account-timezone, feedback,
  and local Finance Assistant slices are present in the current Stage v1 codebase.
- AI transaction entry supports text and receipt review proposals with wallet/category
  suggestions, duplicate handling, explicit user review, and bulk save controls.
- Finance Assistant exposes owner-scoped read-only finance tools, conversation history,
  typed cards, local API-key scopes, and best-effort Redis access audit metadata.
- Transaction ledger supports aggregate/basic-wallet weekly and custom periods, savings
  wallet history, provisional credit history, wallet filtering, and internal transfer
  entry from the transaction options menu.
- Design documentation now follows `system / atoms / molecules / organisms / templates /
  pages / references`; owner PNG/HTML bundles are retained as non-runtime evidence.

### Changed

- The active release narrative is reset to `1.0.0`; old release prose is retained only
  in Git commits, not carried forward as a second active release history.
- Design contracts use page terminology for route-level specifications and a dedicated
  references tree for visual evidence.
- AI provider/model configuration remains environment-driven; credentials are not stored
  in documentation or source control.

### Fixed

- Corrected wallet-scoped weekly ledger filtering so totals and rows use the same
  account-local period predicate.
- Corrected AI-entry empty/503 failure paths for bounded model output and nullable usage
  initialization; incomplete proposals remain reviewable instead of being auto-saved.
- Corrected OCR attachment handling for validated image/PDF inputs and bounded wait/poll
  behavior without exposing provider secrets or financial payloads in logs.

### Documentation

- Added the approved release-baseline detail design and inventory.
- Reconciled design indexes, page contracts, system behavior links, work tickets, backlog,
  validation matrix, context, and agent-readable transaction guidance.

### Known release gates

- Production deployment and public release approval are not claimed by this entry.
- Finance Assistant SSE/reconnect, durable Redis audit delivery/alerting, configured-
  provider browser proof, and full owner UAT remain open where the validation matrix says
  pending or not run.
- Credit-card statement/payment accounting, recurring transactions, advanced reports,
  and Travel Mode runtime remain separate incomplete scopes.
- Frontend `npm run check:design`, lint, build, and tests must be rerun in an environment
  with Node installed; this cleanup environment reported `env: node: No such file or
  directory`.
