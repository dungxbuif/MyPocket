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

## [Unreleased]

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
