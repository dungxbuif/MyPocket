---
artifact_type: validation_matrix
id: VALIDATION_MATRIX
status: active
owner: shared
human_fields:
  - proof_override
  - acceptance_signoff
ai_fields:
  - proof_recommendation
  - evidence_links
  - status_updates
shared_fields:
  - matrix_rows
  - validation_status
updated: 2026-08-31
---

# Validation Matrix

## Field Ownership

- Human owns proof overrides and UAT acceptance sign-off.
- AI recommends proof types, adds real evidence links, and updates status.
- All rows are planned; no design or plan is treated as implementation proof.

## Status Values

| Status | Meaning |
| --- | --- |
| planned | Accepted behavior, not implemented |
| in_progress | Actively built or verified |
| implemented | Implementation exists and every required proof has linked evidence |
| changed | Contract or expected proof changed after implementation |
| retired | No longer accepted |

## Matrix

| Requirement | Phase | Ticket/Bug | Contract/Behavior | Unit | Integration | E2E | UAT | Platform/Manual | Docs Review | Status | Evidence |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| REQ-F-001 | PHASE-001 | TICKET-003 | Google OAuth, stateless cookie, and cross-user isolation | yes | yes | yes | yes | yes | yes | implemented | Backend fixture callback, signed cookie, `/api/v1/me`, repository provisioning, browser fixture login, refresh persistence, logout, and forbidden state evidence recorded |
| REQ-F-002 | PHASE-002 | TICKET-005/007 | Wallets, category hierarchy, activation, archive, and seeds | yes | yes | yes | yes | no | yes | in_progress | PHASE-002 migration/schema/seed, backend wallet/category repository, authenticated wallet/category mutation HTTP routes, API-driven mobile display, mobile wallet/category manager component proof, and live Playwright wallet/category create proof recorded in [PHASE-002 finance verification](test-verification/PHASE-002-finance-core.md); full user UAT remains pending |
| REQ-F-003 | PHASE-002 | TICKET-006 | Atomic income, expense, transfer, adjustment, edit, and search | yes | yes | yes | yes | no | yes | in_progress | TICKET-006 domain, repository, HTTP, frontend component proof, and live Playwright proof covers income, expense, transfer, adjustment, atomic wallet balance/version updates, user-owned wallet checks, category activation checks, duplicate idempotent create replay, edit reversal/reapply, archive reversal once, filtered user-scoped search, CSRF-protected transaction mutations, idempotency header mapping, transaction JSON envelopes, mobile add type controls, report exclusion, and mobile edit/archive controls in [PHASE-002 finance verification](test-verification/PHASE-002-finance-core.md); full user UAT remains pending |
| REQ-F-004 | PHASE-003 | TICKET-008..010 | Offline mirror, outbox, sync, tombstones, and conflict review | yes | yes | yes | yes | yes | yes | in_progress | TICKET-008 frontend proof covers IndexedDB mirror/outbox, legacy migration/quarantine, durable sequence, offline cached hydration, cached-auth offline reload with pending mutation visibility, optimistic transaction queueing, wallet/category queue primitives, degraded read-only UI, and PWA shell offline reload. TICKET-009 proof covers sync migration, idempotent mutation replay, user-scoped change feed, authoritative resync, stale-version conflict responses, CSRF-authenticated sync routes, frontend sync API drain, and mobile reconnect-once E2E. TICKET-010 proof covers conflict storage, non-blocking mobile inbox, keep-server, discard-local, edit-and-retry, and full resync preservation in [PHASE-003 offline sync verification](test-verification/PHASE-003-offline-sync.md); human UAT remains pending |
| REQ-F-005 | PHASE-004 | TICKET-011..014 | Budgets, events, recurring drafts, debts, and alerts | yes | yes | yes | yes | yes | yes | planned | PHASE-004 ready tickets, planned verification artifact, and executable plan are linked; no implementation evidence yet |
| REQ-F-006 | PHASE-005 | TICKET-016/017 | Net worth and approved analytics formulas | yes | yes | yes | yes | no | yes | planned | PHASE-005 ready tickets, planned verification artifact, and executable plan are linked; no implementation evidence yet |
| REQ-F-007 | PHASE-006 | TICKET-018 | Text AI creates validated reviewable drafts | yes | yes | yes | yes | no | yes | planned | Deferred to M2/post-M1 by human scope decision on 2026-08-31; PHASE-006 design and tickets remain available, but no implementation evidence yet |
| REQ-F-008 | PHASE-006 | TICKET-019 | Receipt camera uses S3 and external OCR to create a draft | yes | yes | yes | yes | yes | yes | planned | Deferred to M2/post-M1 by human scope decision on 2026-08-31; PHASE-006 design and tickets remain available, but no implementation evidence yet |
| REQ-F-009 | PHASE-006 | TICKET-020 | AI-chat image uses multimodal AI and shared drafts | yes | yes | yes | yes | yes | yes | planned | Deferred to M2/post-M1 by human scope decision on 2026-08-31; PHASE-006 design and tickets remain available, but no implementation evidence yet |
| REQ-F-010 | PHASE-006 | TICKET-021 | HMAC bank webhook rejects replay/duplicates and creates notice | yes | yes | yes | yes | yes | yes | planned | Deferred to M2/post-M1 by human scope decision on 2026-08-31; PHASE-006 design and tickets remain available, but no implementation evidence yet |
| REQ-F-011 | PHASE-006 | TICKET-018..021 | Generated inputs cannot affect accounting before confirmation | yes | yes | yes | yes | no | yes | planned | Deferred to M2/post-M1 by human scope decision on 2026-08-31; PHASE-006 design and tickets remain available, but no implementation evidence yet |
| REQ-F-012 | PHASE-004 | TICKET-014 | Durable inbox and best-effort Web Push | yes | yes | yes | yes | yes | yes | planned | PHASE-004 ready tickets, planned verification artifact, and executable plan are linked; no implementation evidence yet |
| REQ-F-013 | PHASE-007 | TICKET-023 | Manual user-scoped CSV/Sheets-compatible export | yes | yes | yes | yes | yes | yes | planned | Deferred to M2/post-M1 by human scope decision on 2026-08-31; PHASE-007 design and tickets remain available, but no implementation evidence yet |
| REQ-F-014 | PHASE-007 | TICKET-024 | Confirmed reset/delete removes exact database and S3 scope | yes | yes | yes | yes | yes | yes | planned | Deferred to M2/post-M1 by human scope decision on 2026-08-31; PHASE-007 design and tickets remain available, but no implementation evidence yet |
| REQ-F-015 | PHASE-007 | TICKET-022 | Append-only restricted audit viewer and 180-day default retention | yes | yes | yes | yes | yes | yes | planned | Deferred to M2/post-M1 by human scope decision on 2026-08-31; PHASE-007 design and tickets remain available, but no implementation evidence yet |
| REQ-F-016 | PHASE-001/005 | TICKET-001/015..017 | Installable Vietnamese/VND PWA navigation and workflows | yes | yes | yes | yes | yes | yes | in_progress | Mobile shell tests, build, manifest/service worker, screenshots, Playwright offline reload, PHASE-002 mobile finance component workflow tests, and live finance CRUD Playwright proof pass from `frontend/`; later analytics/planning workflows remain pending |
| REQ-F-017 | PHASE-008 | TICKET-026 | Deferred voice transcription reuses draft contract | yes | yes | yes | yes | yes | yes | planned | Deferred to M3 after M2 draft infrastructure |
| REQ-NF-001 | PHASE-001 | TICKET-003 | Backend ownership isolation on every user-owned operation | yes | yes | yes | not_required | no | yes | in_progress | TICKET-003 ownership harness rejects cross-user object access and browser forbidden state is covered; finance-domain ownership checks pending in later phases |
| REQ-NF-002 | PHASE-002/003/004 | TICKET-006/009/013 | Integer money, atomic writes, and idempotent retries | yes | yes | yes | not_required | no | yes | in_progress | TICKET-006 unit, integration, and HTTP proof covers integer VND accounting effects, database transaction balance updates, duplicate idempotent create replay, persisted deltas for exact edit/archive reversal, and browser-allowed idempotency headers; TICKET-009 adds sync mutation idempotency and reconnect-once E2E proof; later planning retry behavior remains pending |
| REQ-NF-003 | PHASE-003 | TICKET-009/010 | No silent last-write-wins conflict handling | yes | yes | yes | yes | yes | yes | in_progress | TICKET-009 backend integration proof returns explicit conflict for stale transaction versions and preserves server state; TICKET-010 frontend proof stores conflict state, shows a mobile inbox entry, and resolves by explicit keep-server, discard-local, or edit-and-retry actions without silent overwrite |
| REQ-NF-004 | PHASE-001/006/007 | TICKET-003/018..022 | Secrets and sensitive provider data are redacted | yes | yes | yes | not_required | yes | yes | in_progress | TICKET-002/TICKET-003 tests cover safe config/object-store/auth errors, no Google provider-token persistence, and safe forbidden UI; provider phases pending |
| REQ-NF-005 | PHASE-002/004/005 | TICKET-006/011/017 | VND integers and Ho Chi Minh reporting boundaries | yes | yes | yes | yes | yes | yes | in_progress | TICKET-006 unit and integration proof covers VND integer balance effects, persisted VND integer transaction amounts, and date-range search inputs; Ho Chi Minh reporting boundary UAT remains pending |
| REQ-NF-006 | PHASE-001/007 | TICKET-002/004/025 | Homelab health, migration, locks, backup, and restore | yes | yes | yes | not_required | yes | yes | in_progress | TICKET-004 compose startup and health/S3 smoke passed after splitting backend/frontend build contexts; locks, backup, and restore remain pending |
| REQ-NF-007 | All phases | All tickets | Required proof and docs review gate implementation status | not_required | not_required | not_required | yes | yes | yes | planned | No execution evidence |
| REQ-NF-008 | PHASE-001/003/006 | TICKET-001/009/018..021 | Stable safe error codes and correlation IDs | yes | yes | yes | not_required | yes | yes | in_progress | TICKET-001/TICKET-003 API tests cover stable error envelopes and correlation IDs for health/auth/CSRF paths; TICKET-009 sync route tests cover authenticated sync envelopes and safe validation/retryable errors; deferred provider API phases pending |

## Evidence Rules

- Replace “No execution evidence” only with links to real test verification, UAT, platform proof, and docs review.
- `not_required` must retain its reason in the owning ticket or verification artifact.
- Security, deletion, synchronization, accounting, provider, and runtime rows cannot reduce proof without human approval.
- A phase status cannot become `verified` until all its required rows are implemented with evidence.

## Update Log

| Date | Updated By | Change |
| --- | --- | --- |
| 2026-08-24 | AI | Migrated approved MyPocket requirements and planned proof into Harness. |
| 2026-08-30 | AI | Recorded backend/frontend folder split verification for frontend PWA build/E2E and local Compose smoke. |
| 2026-08-30 | AI | Recorded PHASE-001 browser auth E2E coverage for fixture login persistence, logout, and forbidden state. |
| 2026-08-30 | AI | Recorded PHASE-002 finance migration/schema/seed RED/GREEN proof and backend regression evidence. |
| 2026-08-30 | AI | Recorded TICKET-005 backend wallet/category validation and repository proof. |
| 2026-08-30 | AI | Recorded TICKET-005 initial wallet/category HTTP route proof and API contract reconciliation. |
| 2026-08-30 | AI | Recorded TICKET-005 API-driven mobile wallet/category display proof. |
| 2026-08-30 | AI | Recorded TICKET-005 wallet/category mutation API and mobile manager proof plus TICKET-006 mobile transaction type/edit/archive proof. |
| 2026-08-30 | AI | Recorded dedicated mobile Playwright finance CRUD proof for live wallet/category creation and transaction create/edit/archive. |
| 2026-08-30 | AI | Created PHASE-003 offline synchronization tickets, planned verification artifact, and executable implementation plan; validation remains planned pending implementation proof. |
| 2026-08-30 | AI | Created PHASE-004 through PHASE-007 ticket artifacts, planned verification artifacts, and executable implementation plans; validation remains planned pending implementation proof. |
| 2026-08-31 | AI | Recorded TICKET-008 frontend IndexedDB mirror/outbox implementation proof, including unit/component coverage and mobile PWA offline reload smoke; PHASE-003 remains in progress pending sync/conflict work. |
| 2026-08-31 | AI | Recorded human scope decision that current M1 release stops at TICKET-017; TICKET-018 through TICKET-025 are deferred to M2/post-M1 work. |
| 2026-08-31 | AI | Recorded TICKET-009 sync API/change-feed implementation proof, including backend integration, frontend sync API outbox drain, and mobile reconnect replay E2E. |
| 2026-08-31 | AI | Recorded TICKET-010 conflict inbox and recovery proof, including conflict persistence, mobile inbox visibility, keep-server/discard-local/edit-retry actions, full resync, and targeted mobile E2E. |
