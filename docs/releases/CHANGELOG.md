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

All notable changes should be recorded here.

- Prepared the Stage v1 reset: removed the checked-in `refereces/` tree, cleaned its documentation links, and squashed the historical database sequence into `backend/migrations/000001_stage_v1.up.sql` with a non-destructive down file. Fresh staging databases apply at version 1; existing databases must be recreated/reset in staging before the baseline is used.
- Local staging AI now targets the verified oMLX OpenAI-compatible endpoint with `Qwen3.6-35B-A3B-MLX-4bit`; the endpoint and model are documented in `backend/.env.example`, while the API key remains process/local-env only.

- Started local Finance Assistant V1 vertical slice: the Stage v1 baseline stores one owner conversation/run with replay and busy guards; shared owner-scoped finance queries, eight read-only tools, bounded provider orchestration, JWT routes and `/assistant` base-first UI are implemented. SSE/reconnect, public API-key auth, configured-provider browser proof and deployment remain release gates. [AI-ADVISOR-01](../work/tickets/AI-ADVISOR-01-DETAIL_DESIGN.md)

- Added local user API-key authentication for the read-only Finance Assistant: one-time `mpk_...` secret issuance with digest-only storage, owner-scoped JWT key management, scope checks (`finance:read`, `advisor:read`, `advisor:chat`), expiry/revocation and advisor middleware tests. Redis audit, configured-provider browser proof and production deployment remain open.

- Added best-effort Redis advisor access audit on the `mypocket:audit:advisor` stream. Events include request/credential metadata, route, decision, status and latency; prompts, notes, tokens and monetary rows are excluded. Durable delivery/alerting is still a production gate.

## [Unreleased]

- Fixed receipt upload validation to use the detected file signature instead of trusting an empty or incorrect browser MIME header; valid PNG/JPEG/PDF uploads now reach OCR while spoofed content remains rejected. Added a shared floating feedback bubble that captures the current mounted app view (excluding its overlay) and submits an optional private screenshot through the feedback API. Transfer creation is now exposed from the transactions three-dot menu. Live OCR remains blocked only by the configured provider credential returning HTTP 401; see [OCR_API](../architecture/OCR_API.md#401-troubleshooting).

- Fixed Finance Assistant tool contracts and lifecycle safety: all read tools expose typed model schemas, OpenAI-compatible tool-call history is serialized correctly, four-call loops can finish with a final answer, revoked/expired credentials are revalidated during execution, and cancelled runs reject late assistant writes. Added wallet/budget/goal/jar assistant cards and corrected history reload/count formatting. [AI-ADVISOR-01](../work/tickets/AI-ADVISOR-01-DETAIL_DESIGN.md)
- Cancellation now propagates to the in-process provider context while retaining the database lease guard; cross-process interruption remains explicitly unreleased.
- Added bounded advisor history pagination: the UI loads older messages through `before_seq`, deduplicates page boundaries and preserves chronological order without clearing the visible conversation.
- Bounded provider context to the 12 most recent persisted text messages; historical card payloads are not reused as current financial facts.

- Completed local Feedback → Fix → Changelog wiring: backend routes are registered with JWT owner scope and dedicated agent token/audit, lifecycle transitions are locked, changelog publication requires `in_progress` feedback and is atomic, Swagger is regenerated, and Account → Phản hồi UI uses real APIs. Browser feedback E2E and production deployment remain pending. [FEEDBACK_API](../architecture/FEEDBACK_API.md)
- Documentation follow-up: expanded the Finance Assistant plan with [technical implementation instructions](../superpowers/plans/2026-09-22-finance-assistant-technical-guide.md), detailed subtask gates, SQL/Go/API/SSE/UI contracts and isolated accounting fixtures. The docs now distinguish the implemented local JSON/overview/typed-part slice from remaining release gates.

- Design history: added [Finance Assistant design](../work/tickets/AI-ADVISOR-01-DETAIL_DESIGN.md) and [V1 task plan](../superpowers/plans/2026-09-22-finance-assistant-v1.md), grounded in the Go/React app. The implementation is now a local read-only pilot; recovery/security/audit release gates and staged confirmed actions remain separate.

- Added internal wallet transfers: `POST /api/v1/transactions/transfer` creates a source expense and destination income atomically, links them with `transfer_id`, excludes both from reports/jars, and adds a Quick Add transfer mode using shared base fields. Automated handler/PostgreSQL/design/build proof passes; owner browser UAT and pair edit/delete remain pending. [TICKET-02-02](../work/tickets/TICKET-02-02-TRANSFER-DETAIL_DESIGN.md)

- Removed the hardcoded AI entry per-user request cap. MyPocket currently does not enforce user AI/OCR usage limits; idempotent request replay still avoids repeating provider work. Future cost consent, provider pricing and usage budgets remain in [AI-ENTRY-03](../work/tickets/AI-ENTRY-03-PROVIDER-SELECTION.md), with the current decision captured in [AI-USAGE-01](../work/tickets/AI-USAGE-01-DETAIL_DESIGN.md).

- Hardened private attachment cleanup finalization: a stale cleanup worker now gets a conflict if an attachment is no longer in the claimed `deleting` state, instead of silently marking success. Regression evidence is recorded in [AI-ENTRY-02](../work/tickets/AI-ENTRY-02-BATCH-ATTACHMENTS-DETAIL_DESIGN.md).

- Implemented account IANA timezone settings, date-only wallet/budget dates, optional transaction-to-jar assignment, per-month jar configurations/cumulative calculations, and live monthly totals with an independent note and automatic account-local completion (migrations `000013`–`000015`). Added `/jars` and `/months/YYYY-MM` real-API screens using documented shared bases. Backend/PostgreSQL and frontend design/calendar/jar/build checks pass; owner UAT remains pending. No cron close or immutable report snapshot was added. [CORE-03](../work/tickets/CORE-03-TIME-JARS-MONTH-DETAIL_DESIGN.md), [ADR-008](../decisions/ADR-008-account-timezone-and-calendar-dates.md).

- Captured the requested dedicated shared-code cron worker as high-risk draft [WORKER-01](../work/tickets/WORKER-01-cron-service.md), including recurring transaction generation and month-end reporting/close as candidate jobs. No worker process, schema, dependency, or deployment configuration has been added; report snapshot semantics and recurring edge rules need owner review first.

- Fixed live LLM extraction compatibility by switching the existing OpenAI-compatible `net/http` adapter to strict JSON Schema output. Synthetic live evaluation now passes 5/5 cases and 24/24 scored fields with transfer-safety intact; wallet descriptions are included as bounded, untrusted matching context. No additional SDK/provider or ledger side effect was introduced. Broader model UAT remains pending. [AI-ENTRY-01](../work/tickets/TICKET-09-01-ENTRY-DETAIL_DESIGN.md).

- Captured future owner request for selectable AI providers/models with sourced, time-stamped pricing and request-cost estimates. Draft only; no extra provider or billing integration is enabled. [AI-ENTRY-03](../work/tickets/AI-ENTRY-03-PROVIDER-SELECTION.md).

- Fixed private S3 signed GET compatibility by using the AWS SDK for Go v2 S3 presigner with the existing endpoint, region, path-style mode and five-minute expiry. Live S3 readback and OCR pass. At that earlier evaluation point, the five-case Qwen run scored 0/15; the subsequent strict JSON Schema fix and successful rerun are recorded above. [ADR-007](../decisions/ADR-007-aws-s3-presigning.md), [AI-ENTRY-02](../work/tickets/AI-ENTRY-02-BATCH-ATTACHMENTS-DETAIL_DESIGN.md).

- Added privacy-safe structured diagnostics for rejected LLM JSON: schema paths, allowlisted schema/catalog field names/types, model, finish reason, response size and latency; arbitrary keys and prompt/OCR/model response values are not logged. Historical diagnostics found root arrays/input-context-shaped objects; strict JSON Schema resolved the synthetic extraction contract mismatch. [AI-ENTRY-01](../work/tickets/TICKET-09-01-ENTRY-DETAIL_DESIGN.md).

- Replaced AI conversation/session endpoints with a one-shot multipart processing API and read-only result recovery by request ID. Uploaded PDF/image originals are retained in private environment-separated S3, OCR runs before LLM extraction, and the LLM receives text only. Approval atomically links evidence to the transaction; authenticated download is owner/link-scoped. Added explicit cleanup for expired unlinked objects (not scheduled automatically); live provider/browser UAT remains under [AI-ENTRY-02](../work/tickets/AI-ENTRY-02-BATCH-ATTACHMENTS-DETAIL_DESIGN.md).

- Added shared multiline AI composer and grouped result card based on Money Lover's AI entry layout while keeping MyPocket's edit/remove/approve-one/approve-all flow. Replaced the old flat transaction/budget group picker with `CategoryTreeSelector` backed by `BaseCategoryTree` selection mode. Automated UI proof is being reconciled in [UI-FORMS-03](../work/tickets/UI-FORMS-03-DETAIL_DESIGN.md); AI provider and stateless batch API remain separate follow-up work.

- Added hold-Add AI entry with persisted prefilled proposal list, edit-before-approve and reject, version conflict handling and atomic approval replay protection (migration 000010). Normal Add remains manual. Added text/OCR adapters and private local env configuration; real model output previously failed strict schema validation. The one-shot composer/result UI is tracked in [UI-FORMS-03](../work/tickets/UI-FORMS-03-DETAIL_DESIGN.md); retained file storage/linking was subsequently delivered in the AI-ENTRY-02 update above. [Scope and proof](../work/tickets/TICKET-09-01-ENTRY-DETAIL_DESIGN.md), [ADR-005](../decisions/ADR-005-ai-entry-review.md).
- Corrected shared system-category wallet applicability to stay within each account when reading or replacing selections; AI context and confirmation cannot inherit another account's wallet assignments. OCR text is preserved when subsequent model extraction fails.

- Added [two-flow AI feasibility and implementation plan](../work/tickets/TICKET-09-DETAIL_DESIGN.md) for owner review: OCR-first entry, multi-bank reconciliation, atomic confirmation, read-only Q&A, prompt evaluations and jar dependencies. Documentation only, implementation approval pending. Corrected stale OCR scan endpoint guidance against the current public OpenAPI.

- Replaced mounted budget mock data with authenticated persisted CRUD and ledger-derived progress via migration 000009. Explicit date intervals, wallet/category scope, child groups, overlap protection, report exclusion, masking, empty/error/retry states. Recurrence/notifications remain unsupported. Added FE-proxy API roundtrip proof and clean-up of test-only fixtures. [Design](../work/tickets/API-SCREENS-01-DETAIL_DESIGN.md).
- Savings now restricts the picker/API to actual savings catalog groups; external money requires no counterpart wallet. History shows real group names/icons; shared transaction form uses grouped bases. API client decodes problem+json errors into readable messages instead of raw JSON.

- Rebuilt wallet selection into grouped included/excluded lists with aggregate and edit mode; creation now uses BaseSelect. Added goal-date API validation/persistence, savings progress/history and shared category selection. Savings transfer semantics, report/notification delivery and full UAT remain pending.

- Standardized wallet-management screen specification: list/create/edit, base mapping, select-only creation type and separate basic/goal/credit behaviors, with Money Lover source references and explicit implementation gaps.

- Rebuilt Add Wallet as a shared form sheet with grouped inputs, type picker and exclude-total switch. Normalized owner-supplied Add Wallet/Wallet Selector references. Tightened card/control base checks and documented reference fidelity in [ADR-003](../decisions/ADR-003-wallet-reference-enforcement.md). Existing-wallet selector remains planned; visual acceptance pending.

- Removed the optional note field from Add Wallet per owner request.

- Applied the same borderless empty-state presentation to “Chưa có ví” in Overview and wallet management.

- Removed the card border/background/shadow from empty transaction messages in Overview and Transactions using the shared `StatusMessage` plain variant. [UI-EMPTY-01](../work/tickets/UI-EMPTY-01.md).

- Connected the basic income/expense ledger end to end: global add, real transaction list, edit/delete sheet, wallet/category validation, signed VND groups and synchronized Header/Overview/Wallet balances. Added the durable [transaction screen contract](../design/screens/transactions/README.md).
- Wallet list responses now expose ledger-derived `current_balance`; totals honor `is_in_total`. Goal wallets require a positive target and credit wallets require a positive limit. Permanent-delete warnings show dependent transaction count and report impact.
- Wallet metadata edits no longer overwrite opening balance; the edit form makes that field read-only until the dedicated adjustment transaction is implemented.
- Wallet type is immutable after creation to prevent existing basic/goal ledger rows from being reinterpreted with credit semantics.
- Enforced category applicable-wallet scope in transaction validation and excluded credit wallets from the basic ledger until credit purchase/payment semantics are implemented. Added Go and frontend rule tests plus real PostgreSQL/Chrome UAT evidence.

- Normalized `docs/design` to Markdown component/token/behavior specifications from seven screenshots. Removed 7 PNG and 8 HTML exports after extraction (originals recoverable in Git); screen behavior is documented as each screen is implemented.
- Wired the canonical theme and refactored shared controls, text, cards, statuses, navigation, progress/gauge and chart compositions. Added build guardrails against literal colors, native controls outside atoms, local card/text recreation and base visual overrides. Added keyboard tab selection and bottom-sheet focus trapping/Escape/return-focus. [UI-BASE-01 proof](../work/tickets/UI-BASE-01-VERIFICATION.md).

- Aligned the center create button vertically inside the bottom navigation frame with the four tab actions.

- Closed Account → Nhóm management: shared base-composed create/edit cards now support icon, name, type, parent and applicable-wallet selection. System categories lock metadata but allow wallet selection; deleting a personal parent returns its personal children to root. The system icon catalog is applied by migration `000008` for every deployment.

- Reduced active bottom navigation to Tổng quan, Sổ GD, Ngân sách and Tài khoản. Reports remain unmounted; the center add button now opens the verified basic income/expense editor.

- Refined the shared category-tree base: all seeded groups use a global colorful icon catalog while the nested reference geometry is retained—40px parent / 32px child icons, `pl-10` child list, 2px central trunk and curved child branches. Personal-category deletion is available from the shared edit form after confirmation.

- Fixed shared button alignment: icon-plus-label content is now centered by `BaseButton`, repairing the Account → Nhóm back and create actions without screen-local styles.

- Reduced shared card, form and row corner radii by one token step while preserving semantic circular badges and pill controls.

- Re-seeded the default expense, income and debt category catalog from the owner-approved hierarchy. PostgreSQL migration `000004` is applied in dev.

- Added real Account → Nhóm management: owner-scoped category list/create/update/delete, system category read-only protection, two-level parent validation, child-delete protection, owner-validated applicable-wallet selection and loading/error/form states in the React screen.

- Replaced API-startup GORM auto-sync and category seeding with versioned PostgreSQL migrations. The same migration CLI is used in development and production; the local dev database is at clean version 2.
- Split Gin route registration into public auth, account, category, wallet and transaction route groups. Swagger remains code-first: handler annotations plus `go generate ./cmd/api`; no standalone Swagger config exists.
- Added the base-component technical contract from current `docs/design` image/HTML artifacts, including card, keypad, form-row, budget-meter and category-tree constraints.

- Added TanStack Router route tree for finance sections and account management paths.
- Added GORM schema migration for `user`, wallets, categories, and category-wallet assignments, plus idempotent system category seeding.

- Wallet/category ERD now names the account table `user`, per owner direction. Preserving existing account data during table rename is documented; runtime mapping remains unchanged pending migration.

### Category management design

- Added an [ASCII ERD](../architecture/ERD.md#ascii--mô-hình-ví-và-nhóm-để-review) separating the existing user model from proposed wallet/category relationships. Seed ownership and wallet applicability defaults remain under review.

- Added the Account → Quản lý nhóm → select → Edit flow with parent and applicable-wallet selectors to [design guidelines](../design/system/DESIGN.md#account--quản-lý-nhóm) and TICKET-01-03. The requested “category” field remains explicitly unresolved. Documentation only; no UI, API or migration implemented.

### Wallet design review — 2026-09-13

- Corrected the already-approved hard-delete policy and recorded duplicate wallet names as allowed.
- Linked official Money Lover research for adjustment transactions and goal/credit wallet fields in [DESIGN-01-02](../work/tickets/TICKET-01-02-DETAIL_DESIGN.md); implementation and migration approval remain pending.
- Pinned default category reuse to the old app's Vietnamese catalog. No wallet code, seed or migration executed.

### Runtime — 2026-09-13

- Enabled development CORS for the FE origins `http://localhost:4173` and `http://127.0.0.1:4173`, including credentials required by the Google OAuth callback flow. Origins can be overridden with `CORS_ALLOWED_ORIGINS`.
- Updated the app header to show the current total balance in place of the MyPocket/“Tài chính hôm nay” branding, following the reviewed design direction.
- Added implementation-grounded technical documentation for wallet management: clean-architecture boundaries, proposed ERD/migration sequence, and code-first Swagger serving at `/api/v1/docs`.

### Business Tickets — 2026-09-13

- Added [11 parent and 31 child tickets](../work/tickets/README.md) covering the product contract, with concise business scope, acceptance criteria and unresolved questions attached to affected children.
- Linked the draft tickets to the current queue and validation matrix. This supersedes the earlier no-ticket direction; no implementation or product UAT was performed.

### Documentation — 2026-09-13

- Consolidated product specification, business rules, report/Insider design, requirements and review stories into [canonical product docs](../requirements/README.md); removed the two superseded V2 drafts. This is design under review, not a shipped runtime change.
- Recorded editable month-end reporting, independent monthly user notes, automatic context/AI summaries, flexible monthly/cumulative jars and UTC/account-timezone semantics.
- Moved existing standards unchanged to `docs/standards/`; removed the unused Harness CLI phase example and its scheduled milestones. No new implementation tasks were created.
- Runtime tests/UAT were not run for this docs-only update. Documentation evidence is recorded in [VALIDATION_MATRIX.md](../work/VALIDATION_MATRIX.md).

## [1.1.1] - 2026-06-07

### Changed
- **Debugging Standards Enhancements**: Upgraded `docs/standards/DEBUGGING.md` to include 5 major proactive bug-fixing practices for AI agents. The detailed rationale for these changes is as follows:
  - **Assess Impact (Blast Radius Analysis)**:
    - *Problem:* Fixing a local bug in a shared library or core utility can cause a chain reaction, breaking multiple other features.
    - *Solution:* Added a mandatory `Assess Impact` phase before implementing fixes, requiring agents to find references and verify that the proposed change will not break dependent features.
  - **Horizontal Fix & Proactive Search**:
    - *Problem:* Fixing a single instance of a bug (e.g., a missing null check) often leaves similar undiscovered vulnerabilities elsewhere in the codebase.
    - *Solution:* Added a `Proactive Search` phase. After fixing the local bug, agents must scan the entire codebase for the same problematic pattern and address it proactively.
  - **Telemetry First for Opaque Bugs**:
    - *Problem:* The previous standard mandated blocking a bug if it could not be reproduced locally, which could stall progress on environment-specific production issues.
    - *Solution:* If a bug lacks local reproduction steps but occurs in higher environments, the agent must create a PR/Ticket to add Telemetry/Logs/Tracing to gather data before marking it as blocked.
  - **Strict Red-Green Testing**:
    - *Problem:* Writing regression proofs after the fix or without explicit validation can lead to tests that always pass, providing a false sense of security.
    - *Solution:* Enforced strict Red-Green testing. The regression test MUST be written and fail (Red) with recorded output *before* any fix code is written, ensuring the test is actually catching the bug.
  - **Anti-Patterns Documentation (Post-mortem)**:
    - *Problem:* Valuable knowledge gained from complex architectural bugs is often lost after the fix is merged.
    - *Solution:* Required documenting root causes and gotchas as *Anti-Patterns* during the Reconcile phase so future agents avoid repeating the same mistakes.

## [1.1.0] - 2026-06-05

### Added
- **Feedback Intake Flow**: Added `PHASE 0: FEEDBACK INTAKE & TRIAGE` to the Agent Lifecycle.
- **Feedback Funnel**: Created `docs/work/FEEDBACK_LOG.md` to serve as the raw intake funnel for all user feedback before entering the backlog.
- **Feedback Template**: Created `docs/templates/FEEDBACK.md` for deep-dive analysis of complex user feedback.
- **SDD Template**: Added `docs/templates/SDD.md` (Software Design Document) to track high-level system architecture and required its synchronization upon detail design changes in `DOCS.md`.

### Changed
- **Triage Types**: Allowed triage types for user feedback in `WORKFLOW.md` now explicitly include `Enhancement`.
- **Detail Design Revamp**: Revamped `docs/templates/DETAIL_DESIGN.md` structure. It now integrates visual components (Architecture Overview, Execution Flow, API & Data Model Design, Security) while retaining strict Harness lifecycle metadata and reconciliation rules.
- **Agent Rules**: Updated `AGENTS.md` (Rule 9: Completion Rules) to strictly enforce writing framework and user-facing changes to `CHANGELOG.md`.
- **Docs Guide**: Updated `docs/README.md` to include the new templates and feedback intake funnel.

## [1.0.0] - 2026-06-05

### Added
- Initial SDLC agent framework scaffold.
- Reconciled Overview cards with the exported base-cards design: shared `SurfaceCard` and `IconBadge` atoms now drive wallet rows, transaction rows, insight and report cards.
- Centralized reusable component variants in the global UI token module; removed local variant maps from atoms and list molecules.
