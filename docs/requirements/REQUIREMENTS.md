---
artifact_type: requirements_index
id: REQUIREMENTS
status: active
owner: shared
human_fields: [priority, acceptance, requirement_source]
ai_fields: [requirement_rows, status_updates, trace_links]
shared_fields: [functional_requirements, non_functional_requirements]
updated: 2026-08-24
---

# Requirements

## Field Ownership

- Human-approved priority and acceptance derive from [`SPEC.md`](SPEC.md) and the linked design review.
- AI maintains stable IDs, phase links, proof links, and implementation status.

## Functional Requirements

| ID | Requirement | Priority | Source | Status |
| --- | --- | --- | --- | --- |
| REQ-F-001 | Authenticate multiple users with Google OAuth, provision isolated user identities, and maintain login with a signed stateless cookie. | High | [SPEC](SPEC.md) | accepted |
| REQ-F-002 | Manage wallets, credit-card metadata, the two-level Vietnamese category taxonomy, and per-wallet category activation without deleting referenced history. | High | [SRS](../../SRS.md) | accepted |
| REQ-F-003 | Create, edit, archive, and search income, expense, transfer, and balance-adjustment transactions with atomic wallet balance effects. | High | [SRS](../../SRS.md) | accepted |
| REQ-F-004 | Support full offline read/write through IndexedDB, an idempotent outbox, incremental sync, tombstones, optimistic versions, and explicit conflict resolution. | High | [Design](../superpowers/specs/2026-08-23-mypocket-system-design.md) | accepted |
| REQ-F-005 | Manage budgets with 80% and 100% alerts, events/trips, recurring transaction drafts, and debt/loan records. | High | [SRS](../../SRS.md) | accepted |
| REQ-F-006 | Present net worth, recent transactions, wallet filtering, net income, category composition, daily spending, period comparison, and three-month cumulative trends. | High | [SRS](../../SRS.md) | accepted |
| REQ-F-007 | Parse text AI conversations into one or more validated income, expense, or transfer drafts using an OpenAI-compatible endpoint. | High | [Design](../superpowers/specs/2026-08-23-mypocket-system-design.md) | accepted |
| REQ-F-008 | Capture a receipt from add-transaction, store it privately, call an external OCR API, and create a reviewable draft. | High | [Design](../superpowers/specs/2026-08-23-mypocket-system-design.md) | accepted |
| REQ-F-009 | Accept an image inside AI chat, call an OpenAI-compatible multimodal endpoint, and create a conversation-linked draft. | High | [Design](../superpowers/specs/2026-08-23-mypocket-system-design.md) | accepted |
| REQ-F-010 | Accept HMAC-signed bank-notification webhooks with replay protection and deduplication, then notify the user about the resulting draft. | High | [Design](../superpowers/specs/2026-08-23-mypocket-system-design.md) | accepted |
| REQ-F-011 | Require review and confirmation before any AI, OCR, webhook, or recurring draft changes wallet accounting. | Urgent | [Design](../superpowers/specs/2026-08-23-mypocket-system-design.md) | accepted |
| REQ-F-012 | Deliver durable in-app notices and best-effort Web Push for budgets, recurring drafts, imports, and review-required states. | Medium | [SPEC](SPEC.md) | accepted |
| REQ-F-013 | Generate manual CSV and Google Sheets-compatible snapshot exports without importing spreadsheet changes. | Medium | [SPEC](SPEC.md) | accepted |
| REQ-F-014 | Let users reset or delete their account through confirmed, audited background operations that include owned S3 objects. | High | [SPEC](SPEC.md) | accepted |
| REQ-F-015 | Record append-only state-changing and security audit events, retain them for a configurable default of 180 days, and expose a hidden read-only page only to `AUDIT_VIEWER_EMAIL`. | High | [Design](../superpowers/specs/2026-08-23-mypocket-system-design.md) | accepted |
| REQ-F-016 | Provide an installable five-destination PWA with quick add, transaction search, wallet views, Vietnamese copy, VND formatting, and a balance-privacy toggle. | High | [SRS](../../SRS.md) | accepted |
| REQ-F-017 | Add voice capture and OpenAI-compatible transcription only after the initial seven phases are delivered and verified. | Low | [SPEC](SPEC.md) | deferred |

## Non-Functional Requirements

| ID | Requirement | Category | Status |
| --- | --- | --- | --- |
| REQ-NF-001 | Enforce authenticated per-user ownership on every user-owned backend query and mutation. | Security | accepted |
| REQ-NF-002 | Store monetary values as integers, make finance operations atomic, and make every retryable mutation idempotent. | Correctness | accepted |
| REQ-NF-003 | Never resolve concurrent record changes through silent last-write-wins behavior. | Data integrity | accepted |
| REQ-NF-004 | Keep secrets and raw sensitive provider data out of source control, client errors, structured logs, and audit diffs. | Privacy/security | accepted |
| REQ-NF-005 | Use `Asia/Ho_Chi_Minh` reporting boundaries and VND integer units in the initial release. | Localization | accepted |
| REQ-NF-006 | Deploy web, API, worker, PostgreSQL, and S3 integration with health, migration, locking, backup, and restore checks suitable for a homelab. | Operations | accepted |
| REQ-NF-007 | Prove accepted behavior with the unit, integration, E2E, UAT, platform, security, and docs-review evidence assigned in the validation matrix. | Quality | accepted |
| REQ-NF-008 | Use stable machine-readable API error codes and correlation IDs without exposing provider errors or stack traces. | Operability | accepted |
