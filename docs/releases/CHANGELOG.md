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
