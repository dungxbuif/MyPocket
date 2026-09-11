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

Latest 2026-09-11 completion run: backend race/integration evidence covers API authorization, lifecycle, Agent and OCR workers; frontend **177/177**, production build and **105/105 browser tests pass**. Agent text/image remains review-first and transfer E2E proves no income/expense report pollution. [Physical device UAT](completion/PHYSICAL-DEVICE-UAT.md) and production provider smoke remain open release gates.

2026-09-11: [portfolio/offline follow-up evidence](test-verification/BUSINESS-LOGIC-AUDIT-2026-09-10.md) adds overflow RED/GREEN, failed trade/price rollback including feed/version, aggregate validation, and pending transaction identity tests. Full browser baseline 79/81; follow-up runs and category cache investigation must be resolved before release claims.

2026-09-10 business-logic/API audit: [live record](test-verification/BUSINESS-LOGIC-AUDIT-2026-09-10.md) adds RED-GREEN proof for checked accounting/analytics arithmetic, exact transaction reversal, wallet/transaction/planning/category optimistic versions, recurring validation, obligation repayment direction/locking, HCM dates, Bearer precedence, Redis-stale revocation and sync race-to-conflict handling. Atomic sync now has local rollback, replay and lost-response E2E proof. Release still requires migration 0012, full client resync and existing physical-device/provider gates; see the updated record.

Beta reliability evidence: [2026-09-07 verification](test-verification/BETA-RELIABILITY.md). Adds strict TypeScript build validation, PostgreSQL/race proof, desktop/mobile browser tests, real two-user/API-key/revocation checks, storage-quota/auth-denial tests and account-switch preservation. Production Google OAuth, S3/Web Push and physical-device installation remain unverified by this local suite.

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

2026-09-10: [Sổ tiền F1 local proof](test-verification/SO-TIEN-F1-2026-09-10.md) records 138 frontend tests, build and 3 desktop browser regressions passing for selected budgets, HCM report/remaining-day dates and asynchronous read failures. F2/F3, full browser/Go rerun, retained WebKit/device gates and visual UI acceptance remain open; no release is implied.

2026-09-09 local functional/accounting: [proof](test-verification/R0-FUNCTIONAL-E2E-2026-09-09.md) adds literal ledger assertions for transfer create/edit/archive and five date/wallet-filtered reports including timezone boundaries, exclusion, adjustment and archive. New 9/9 browser passes are included in full 61/63; two WebKit offline failures retained. Frontend 120 pass, Go/PostgreSQL suite pass, build/typecheck pass. No full-function, public-docs, visual fidelity, physical-device or deployment acceptance.

2026-09-09 21:25 redeployment: [fresh platform proof](test-verification/R0-WEB-DOCS-RELEASE-2026-09-09.md) verifies same-image web/API recreation, healthy API, public readiness/database, JS 200 (333903 bytes), expected unauthenticated docs redirect, and successful registry push/manifest config identity. Public connectivity blocker resolved before this run; no infrastructure fix claimed. Authenticated docs/device UAT and full R0 remain open.

2026-09-09 web/docs rollout: [platform proof](test-verification/R0-WEB-DOCS-RELEASE-2026-09-09.md) records 100 frontend tests, both Docker builds, byte-identical API executables, local readiness/database, new web assets, and expected unauthenticated docs redirect. Production containers updated; worker/migrations untouched. Public HTTPS and remote registry publication blocked by connection closure, authenticated docs rendering and physical-device UAT not accepted. No whole-requirement promotion.

2026-09-09 search feedback: [proof](test-verification/R0-SEARCH-FEEDBACK-2026-09-09.md) adds 8 real-App search tests (100 frontend total) and 6 browser checks across Chrome mobile/desktop + WebKit. API failure/retry, pending/empty and stale response invalidation verified. No whole-requirement promotion; cache freshness, offline index and full navigation/search parity remain separate. Physical PWA/receipt gates unchanged.

2026-09-09 focused receipt readback follow-up: [diagnostic evidence](test-verification/R0-RECEIPT-READBACK-2026-09-09.md) separates a reproducible WebKit offline-emulation read failure from passing byte/entity checks under a separately simulated HTTP outage. Latest targeted suite 8 pass/1 retained emulation failure; unit 92 pass. This adds proof without replacing the original failing test or accepting physical Safari/PWA/provider behavior. No application/storage changes or whole-requirement promotion.

2026-09-09 latest R0 follow-up: [receipt-controls evidence](test-verification/R0-RECEIPT-CONTROLS-2026-09-09.md) records frontend 92 pass and targeted browser 5 pass/1 WebKit stored-image readback failure. Receipt controls and shared footer are implemented, but the slice is blocked under the loop guard pending focused File/Blob readback review. [Earlier error-feedback proof](test-verification/R0-ERROR-FEEDBACK-2026-09-09.md) retains transaction/API-key feedback and stale-response guard evidence. The original WebKit offline-reload gate is separate and unchanged. No whole requirement is promoted; full R0, physical-device UAT and API completeness remain open. [Mounted actions](test-verification/R0-MOUNTED-ACTIONS-2026-09-09.md) distinguish handlers, working browser journeys and missing functionality.

Latest evidence qualification (2026-09-08): [Money Lover parity verification](test-verification/MONEYLOVER-PARITY-2026-09-08.md) supersedes blanket current-UI success interpretations of historical evidence below. Backend: 139 test pass events with dedicated PostgreSQL, 2 provider skips. Frontend: 60 tests/build pass. Mobile: 7 distinct passes across two runs, 2 remaining failures. REQ-F-001/016 logout is obstructed by the PWA prompt; REQ-F-005 named-budget edit/archive E2E is incomplete; REQ-F-002 custom category parent, REQ-F-006 full reporting and REQ-F-018 mounted asset trade flow are not verified. The `yes` proof columns below specify required proof types, not an assertion that every current flow passed. No requirements are promoted to `implemented` or `verified` by this audit.

| Requirement | Phase | Ticket/Bug | Contract/Behavior | Unit | Integration | E2E | UAT | Platform/Manual | Docs Review | Status | Evidence |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| REQ-F-001 | PHASE-001 | TICKET-003 | Google OAuth, stateless cookie, and cross-user isolation | yes | yes | yes | yes | yes | yes | in_progress | Fixture callback remains covered by backend/browser tests; real mode now builds Google authorization URL, validates state, exchanges code, reads userinfo, provisions identity, and issues the signed cookie. Live provider UAT still requires Google credentials, redirect registration, API, and PostgreSQL. |
| REQ-F-002 | PHASE-002 | TICKET-005/007 | Wallets, category hierarchy, activation, archive, and seeds | yes | yes | yes | yes | no | yes | in_progress | PHASE-002 migration/schema/seed, backend wallet/category repository, authenticated wallet/category mutation HTTP routes, API-driven mobile display, mobile wallet/category manager component proof, and live Playwright wallet/category create proof recorded in [PHASE-002 finance verification](test-verification/PHASE-002-finance-core.md); full user UAT remains pending |
| REQ-F-003 | PHASE-002 | TICKET-006 | Atomic income, expense, transfer, adjustment, edit, and search | yes | yes | yes | yes | no | yes | in_progress | TICKET-006 domain, repository, HTTP, frontend component proof, and live Playwright proof covers income, expense, transfer, adjustment, atomic wallet balance/version updates, user-owned wallet checks, category activation checks, duplicate idempotent create replay, edit reversal/reapply, archive reversal once, filtered user-scoped search, CSRF-protected transaction mutations, idempotency header mapping, transaction JSON envelopes, mobile add type controls, report exclusion, and mobile edit/archive controls in [PHASE-002 finance verification](test-verification/PHASE-002-finance-core.md); full user UAT remains pending |
| REQ-F-004 | PHASE-003 | TICKET-008..010 | Offline mirror, outbox, sync, tombstones, and conflict review | yes | yes | yes | yes | yes | yes | in_progress | TICKET-008 frontend proof covers IndexedDB mirror/outbox, legacy migration/quarantine, durable sequence, offline cached hydration, cached-auth offline reload with pending mutation visibility, optimistic transaction queueing, wallet/category queue primitives, degraded read-only UI, and PWA shell offline reload. TICKET-009 proof covers sync migration, idempotent mutation replay, user-scoped change feed, authoritative resync, stale-version conflict responses, CSRF-authenticated sync routes, frontend sync API drain, and mobile reconnect-once E2E. TICKET-010 proof covers conflict storage, non-blocking mobile inbox, keep-server, discard-local, edit-and-retry, and full resync preservation in [PHASE-003 offline sync verification](test-verification/PHASE-003-offline-sync.md); human UAT remains pending |
| REQ-F-005 | PHASE-004 | TICKET-011..014 | Budgets, events, recurring drafts, debts, and alerts | yes | yes | yes | yes | yes | yes | in_progress | TICKET-011 proof covers budget CRUD, period validation, selected/all expense category scopes, confirmed-expense progress formulas, threshold dedupe, and mobile budget CRUD. TICKET-012 proof covers event/debt CRUD, linking, totals, and overpayment rejection. TICKET-013 proof covers recurring schedule setup, worker leases, deterministic no-accounting drafts, and mobile draft review rows in [PHASE-004 planning verification](test-verification/PHASE-004-planning-automation.md); inbox/Web Push and full UAT remain pending |
| REQ-F-006 | PHASE-005 | TICKET-016/017 | Net worth and approved analytics formulas | yes | yes | yes | yes | yes | yes | in_progress | Server-side dashboard/report formulas, category roll-up, daily/cumulative series, comparison handling, and Money Insider Home category-frequency/daily-average comparison implemented; Insider six-period detail and PostgreSQL UAT remain |
| REQ-F-007 | PHASE-006 | completion Task 6 | Text Agent creates validated reviewable drafts | yes | yes | yes | yes | no | yes | in_review | [Agent proof](completion/AGENT-REVIEW-FIRST.md) and [mounted UI proof](completion/AGENT-UI.md); production model-provider smoke remains |
| REQ-F-008 | PHASE-006 | completion Task 7 | Owned receipt uses S3 and external OCR image tool to support a draft | yes | yes | yes | yes | yes | yes | in_review | [OCR tool proof](completion/OCR-AGENT-TOOL.md); physical receipt/OCR and production provider UAT remain |
| REQ-F-009 | PHASE-006 | completion Tasks 7–8 | Agent image uses third-party OCR and OpenAI-compatible structured analysis | yes | yes | yes | yes | yes | yes | in_review | Backend ownership/no-accounting tests and six text/image UI E2E checks pass; provider UAT remains |
| REQ-F-010 | future | none | Bank ingestion | no | no | no | no | no | yes | retired | Explicitly outside the current release; no route or table is advertised |
| REQ-F-011 | PHASE-006 | completion Tasks 6–8 | Generated inputs cannot affect accounting before confirmation | yes | yes | yes | yes | no | yes | in_review | Agent/OCR persistence and E2E prove draft-only behavior; physical/provider UAT remains |
| REQ-F-012 | PHASE-004 | TICKET-014 | Durable inbox and best-effort Web Push | yes | yes | yes | yes | yes | yes | in_progress | Notification migration/repository/API, worker retry cleanup, and mobile permission-state UI implemented; provider delivery/UAT remains |
| REQ-F-013 | PHASE-007 | completion Task 4 | Manual user-scoped CSV/Sheets-compatible export | yes | yes | yes | yes | yes | yes | in_review | [Lifecycle proof](completion/DATA-LIFECYCLE.md); production object-store smoke remains |
| REQ-F-014 | PHASE-007 | completion Task 4 | Confirmed reset/delete removes exact database and S3 scope | yes | yes | yes | yes | yes | yes | in_review | [Lifecycle proof](completion/DATA-LIFECYCLE.md); production backup/restore and destructive UAT remain |
| REQ-F-015 | PHASE-007 | TICKET-022/028 | Append-only restricted audit viewer and 180-day default retention | yes | yes | yes | yes | yes | yes | in_review | Authorization inventory, audit repository and retention tests pass; production retention/permission smoke remains |
| REQ-F-016 | PHASE-001/005 | TICKET-001/015..017 | Installable Vietnamese/VND PWA navigation and workflows | yes | yes | yes | yes | yes | yes | in_progress | Tailwind v4/Vite integration, base UI primitives, mobile shell/search/dashboard/report UI, compact wallet list/sheet with create-wallet entry, liquid-glass bottom nav/sheets, transparent scrollbars, standalone/Apple install metadata, service worker push handling, cached-auth offline reload, React tests, production build, and mobile Playwright offline-reload proof pass; desktop and live analytics UAT remain |
| REQ-F-017 | PHASE-008 | TICKET-026 | Deferred voice transcription reuses draft contract | yes | yes | yes | yes | yes | yes | planned | Deferred to M3 after M2 draft infrastructure |
| REQ-F-018 | PHASE-005 | TICKET-027 | User-owned assets preserve offline buy/sell history, moving-average cost and hybrid price history, calculate realized/unrealized P&L, and never mutate wallet accounting | yes | yes | yes | yes | yes | yes | in_review | Evidence in [TICKET-027 asset portfolio verification](test-verification/TICKET-027-asset-portfolio-valuation.md): portfolio domain tests, real PostgreSQL repository tests, migration proof, backend regression, REST route wiring, dashboard investment totals, sync asset replay, IndexedDB asset cache/outbox, static provider price refresh worker, frontend app/offline tests, build, and whitespace check passed. Human UAT remains pending. |
| REQ-F-015 | PHASE-007 | TICKET-028 | Production debug audit foundation records redacted state/security audit events with correlation IDs, hidden exact-email viewer access, structured safe logs, retention purge, Account-tab/user API keys for third-party/AI-agent access, Redis auth cache with revoke invalidation, and user-scoped receipt media with offline retry | yes | yes | yes | yes | yes | yes | in_review | Evidence in [TICKET-028 verification](test-verification/TICKET-028-production-hardening-debug-audit-foundation.md): targeted backend receipt/audit tests, real PostgreSQL migration/API-key/receipt proof, cache sync/revoke tests, S3-compatible put/delete smoke, full backend regression, frontend app/offline tests, and production build passed. Human UAT remains. |
| REQ-F-002 | PHASE-002 | TICKET-007 | Default Vietnamese category catalog includes parent/child expense, income, and debt groups | yes | yes | yes | yes | yes | yes | in_review | Migration `0011_phase002_category_catalog.sql` seeds the complete catalog idempotently; sequential PostgreSQL migration and finance repository tests pass. Human category-picker UAT remains. |
| REQ-NF-001 | PHASE-001 | TICKET-003 | Backend ownership isolation on every user-owned operation | yes | yes | yes | not_required | no | yes | in_progress | TICKET-003 ownership harness rejects cross-user object access and browser forbidden state is covered; finance-domain ownership checks pending in later phases |
| REQ-NF-002 | PHASE-002/003/004 | TICKET-006/009/012/013 | Integer money, atomic writes, and idempotent retries | yes | yes | yes | not_required | no | yes | in_progress | TICKET-006 unit, integration, and HTTP proof covers integer VND accounting effects, database transaction balance updates, duplicate idempotent create replay, persisted deltas for exact edit/archive reversal, and browser-allowed idempotency headers; TICKET-009 adds sync mutation idempotency and reconnect-once E2E proof; TICKET-012 keeps event/debt money as integer VND and links only confirmed owned transactions without changing accounting; TICKET-013 uses deterministic occurrence keys and unique draft constraints for worker retry idempotency |
| REQ-NF-003 | PHASE-003 | TICKET-009/010 | No silent last-write-wins conflict handling | yes | yes | yes | yes | yes | yes | in_progress | TICKET-009 backend integration proof returns explicit conflict for stale transaction versions and preserves server state; TICKET-010 frontend proof stores conflict state, shows a mobile inbox entry, and resolves by explicit keep-server, discard-local, or edit-and-retry actions without silent overwrite |
| REQ-NF-004 | PHASE-001/006/007 | TICKET-003/018..022 | Secrets and sensitive provider data are redacted | yes | yes | yes | not_required | yes | yes | in_progress | TICKET-002/TICKET-003 tests cover safe config/object-store/auth errors, no Google provider-token persistence, and safe forbidden UI; provider phases pending |
| REQ-NF-005 | PHASE-002/004/005 | TICKET-006/011/012/017 | VND integers and Ho Chi Minh reporting boundaries | yes | yes | yes | yes | yes | yes | in_progress | Finance/planning integer proof plus analytics normalized Ho Chi Minh filters and daily boundaries implemented; broader reporting UAT remains |
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
| 2026-08-31 | AI | Recorded TICKET-011 budgets and threshold proof, including planning migration, API, progress formulas, threshold dedupe, mobile budget screen, and live mobile E2E. |
| 2026-08-31 | AI | Recorded TICKET-012 events/debts proof, including planning migration, authenticated API routes, event totals with report exclusions, repayment overpayment rejection, mobile planning UI, and live mobile E2E. |
| 2026-08-31 | AI | Recorded TICKET-013 recurring schedule proof, including recurring/draft/lease migration, authenticated schedule/draft APIs, worker lease gate, deterministic draft creation, no-balance-change integration test, and mobile schedule setup E2E. |
| 2026-08-31 | AI | Added planned REQ-F-018 proof for human-promoted TICKET-027 asset portfolio valuation immediately after TICKET-017; design approval and implementation evidence remain pending. |
| 2026-08-31 | Human/AI | Approved TICKET-027 D-01 through D-04: hybrid pricing, moving-average buy/sell accounting, separate dashboard totals, and offline manual mutations; implementation evidence remains pending. |
| 2026-08-31 | AI | Recorded TICKET-027 Wave 1 schema/domain/repository proof for asset positions, buy/sell ledger replay, manual prices, migration, ownership, archive retention, and wallet non-interference. |
| 2026-08-31 | AI | Recorded TICKET-027 initial API/dashboard/frontend proof plus login email whitelist config and callback rejection behavior. |
| 2026-08-31 | AI | Completed TICKET-027 implementation proof for sync/offline asset mutation replay, IndexedDB asset cache/outbox, static provider price refresh worker, PostgreSQL sequential integration, full backend regression, frontend component/offline tests, production build, and whitespace check. |
| 2026-08-31 | Human/AI | Promoted TICKET-028 Production Hardening and Debug Audit Foundation before prod/UAT; plan/design created and implementation started without additional confirmation per owner instruction. |
| 2026-08-31 | AI | Recorded TICKET-028 implementation proof for structured logging, panic recovery, audit events/viewer/retention, API keys for third-party agents, Redis auth cache, production config validation, PostgreSQL migration proof, backend regression, and frontend build. |
