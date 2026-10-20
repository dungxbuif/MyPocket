---
artifact_type: requirement_spec
id: REQ-MASTER
status: active
owner: human
human_fields: [product_summary, goals, non_goals, users_and_stakeholders, acceptance_criteria]
ai_fields: [functional_requirements, non_functional_requirements, constraints, linked_decisions]
shared_fields: [status, trace]
trace:
  source: SRS.md
  design: docs/superpowers/specs/2026-08-23-mypocket-system-design.md
  requirements: docs/requirements/REQUIREMENTS.md
  user_stories: docs/requirements/USER_STORIES.md
  roadmap: docs/work/ROADMAP.md
---

# MyPocket Product Specification

## Field Ownership

- Human-approved intent, scope, priorities, and acceptance are recorded from the design review completed on 2026-08-23 and 2026-08-24.
- AI maintains requirement wording, constraints, trace links, and implementation-grounded status.

## Product Summary

MyPocket is a multi-user, VND-first personal-finance PWA for a self-hosted homelab. It combines reliable wallet accounting, offline read/write operation, budgets and analytics, review-first AI/OCR/bank-notification ingestion, Web Push, manual export, and restricted debugging audit logs.

## Goals

- Let each Google-authenticated user manage isolated financial data from desktop or an installed mobile PWA.
- Keep wallet balances and reports correct across online, offline, retried, and concurrent operations.
- Reduce manual entry through text AI, receipt OCR, AI-chat images, recurring drafts, and signed bank-notification webhooks.
- Require human confirmation before generated inputs affect accounting.
- Run as a maintainable React, Go, PostgreSQL, and S3-compatible homelab deployment.
- Provide one authorized account with a hidden audit-log page for debugging state changes and security events.

## Non-Goals

- Native Android notification listening in this repository.
- Direct bank OAuth connectivity.
- Voice input in the initial release; it is a deferred phase.
- Model training or a first-party OCR model.
- Real-time or bidirectional Google Sheets synchronization.
- Full exchange-rate-based multi-currency valuation.
- Device management, device limits, or a database session table.
- Subscription tiers, ads, paywalls, or payments.

## Users And Stakeholders

- **Finance user:** authenticates with Google and owns wallets, categories, transactions, plans, reports, drafts, notifications, exports, and attachments.
- **Audit viewer:** the one verified Google email configured by `AUDIT_VIEWER_EMAIL`; can read the hidden audit page but gains no ownership over other users' finance records.
- **Homelab operator:** supplies production connection strings and provider credentials, deploys API/worker/web containers, and manages backup and restore.
- **External notification sender:** posts bank notification text to a user-owned HMAC-protected webhook source.

## Functional Requirements

The authoritative requirement index is [`REQUIREMENTS.md`](REQUIREMENTS.md). It covers identity, finance core, offline synchronization, planning, analytics, ingestion, notifications, export, account lifecycle, audit, and production operations.

## Non-Functional Requirements

- PostgreSQL is the authoritative store; IndexedDB is the offline client mirror and outbox.
- Financial writes, balance effects, sync change records, and transfer legs are transactionally consistent.
- All user-owned API access is scoped by authenticated `user_id` in Go.
- Mutations are idempotent and version conflicts require explicit resolution.
- Secrets are environment-provided and excluded from source, API errors, logs, and audit diffs.
- VND uses integer units and reporting boundaries use `Asia/Ho_Chi_Minh`.
- Provider failures cannot create confirmed financial records.
- Production includes health checks, migrations, structured correlation IDs, worker locking, and backup/restore proof.

## Constraints

- Frontend: React + TypeScript PWA using Vite, React Router, and TanStack Query.
- Backend: Go modular monolith exposed as an API process plus a separate worker process sharing domain packages.
- Database: PostgreSQL; production connection string will be supplied by the homelab operator.
- Objects: private S3-compatible storage.
- Authentication: Google OAuth with a signed stateless HTTP-only cookie and no persisted Google token.
- AI: OpenAI-compatible text and multimodal endpoints configured through environment variables.
- Receipt capture: external OCR API from the add-transaction flow.
- Audit retention: configurable through `AUDIT_RETENTION_DAYS`, default `180` days.

## Acceptance Criteria

- A user can complete the core finance, offline, planning, analytics, ingestion, notification, export, and account-lifecycle workflows specified in [`USER_STORIES.md`](USER_STORIES.md).
- Concurrent offline edits never silently overwrite an authoritative server version.
- AI, OCR, webhook, and recurring inputs remain drafts until an authenticated user confirms them.
- Cross-user object access is denied by integration and end-to-end proof.
- The audit route is absent from navigation, forbidden to every non-configured account, and readable by the configured audit viewer.
- Required automated, UAT, platform, and docs-review evidence is recorded in [`../work/VALIDATION_MATRIX.md`](../work/VALIDATION_MATRIX.md) before requirements are marked implemented.

## Linked Decisions

- [`ADR-001`](../decisions/ADR-001-react-go-modular-monolith.md)
- [`ADR-002`](../decisions/ADR-002-stateless-google-oauth.md)
- [`ADR-003`](../decisions/ADR-003-offline-sync-conflict-review.md)
- [`ADR-004`](../decisions/ADR-004-review-first-ingestion.md)
- [`ADR-005`](../decisions/ADR-005-restricted-audit-log.md)
