# MyPocket System Design

**Status:** Approved in design review on 2026-08-23  
**Source:** [`SRS.md`](../../../SRS.md)  
**Planned delivery:** [`../plans/`](../plans/)  
**Risk lane:** High-risk  
**Risk flags:** Authentication, authorization, financial data, data deletion, public API, external providers, webhooks, offline synchronization, deployment/runtime

## 1. Problem and Goal

MyPocket is a multi-user personal-finance progressive web application inspired by MoneyLover. It must support reliable financial recording, offline use, planning, analytics, AI-assisted ingestion, receipt processing, bank-notification ingestion, and homelab deployment without depending on a commercial banking connection.

The system must preserve financial correctness before convenience. AI, OCR, webhook, recurring, and offline inputs therefore create reviewable drafts or validated mutations; they must not silently corrupt balances or overwrite concurrent edits.

## 2. Approved Product Scope

The initial product program includes:

- Google OAuth authentication for multiple users.
- Wallets, credit-card metadata, a two-level category tree, and wallet-specific category activation.
- Income, expense, transfer, and balance-adjustment transactions.
- A five-destination PWA shell for overview, transactions, quick add, budgets, and account settings.
- Transaction search, wallet-scoped views, and a privacy control that hides displayed balances.
- Full offline read/write behavior with synchronization and explicit conflict review.
- Budgets, threshold alerts, events/trips, recurring transactions, and debt/loan tracking.
- Dashboard and analytics for net worth, category composition, daily spending, period comparison, and cumulative trends.
- Text-based AI chat that can extract one or many transactions and transfers.
- A receipt-camera flow that sends an image to an external OCR API.
- A separate AI-chat flow that sends an attached image directly to an OpenAI-compatible multimodal endpoint.
- A signed webhook that accepts bank-notification text from an external Android automation or companion.
- In-app notifications and Web Push.
- Manual CSV and Google Sheets-compatible export.
- Account reset and deletion.
- A hidden audit-log page restricted to one configured Google account.
- Production deployment on a homelab using a PostgreSQL connection string supplied through environment configuration.

Voice input is deliberately deferred to a later phase. The initial AI ingestion scope is text, receipt OCR, AI-chat images, and bank-notification webhooks.

## 3. Explicit Non-Goals

- No native Android notification-listener application in this repository.
- No direct bank OAuth integration.
- No real-time or bidirectional Google Sheets synchronization.
- No model training or fine-tuning.
- No first-party OCR model; the application integrates an external OCR API.
- No device-management UI or device limits.
- No server-side session table.
- No subscription tier, advertisement, paywall, or payment processing.
- No full multi-currency valuation or exchange-rate provider in the initial release.
- No automatic transaction confirmation from AI, OCR, webhooks, or recurring schedules.

## 4. Technology and Repository Architecture

The project uses a single repository with three deployable applications and shared packages:

```text
+----------------------+        REST/JSON         +----------------------+
| React TypeScript PWA | -----------------------> | Go API               |
| IndexedDB + outbox   | <----------------------- | Modular monolith     |
+----------------------+     sync/change feed     +----------+-----------+
                                                           |
                                                           | shared domain and
                                                           | application packages
                                                           v
                                                +----------------------+
                                                | Go Worker            |
                                                | jobs and notifications|
                                                +----------+-----------+
                                                           |
                         +---------------------------------+------------------+
                         |                                 |                  |
                         v                                 v                  v
                  +-------------+                   +-------------+   +-------------+
                  | PostgreSQL  |                   | S3-compatible|   | Providers   |
                  | source truth|                   | receipt store|   | OAuth/AI/OCR|
                  +-------------+                   +-------------+   +-------------+
```

The intended top-level layout is:

```text
apps/
  web/       React, TypeScript, Vite, React Router, TanStack Query, PWA
  api/       Go HTTP API entrypoint
  worker/    Go background-worker entrypoint
internal/
  identity/
  finance/
  planning/
  analytics/
  ingestion/
  sync/
  notification/
  export/
  audit/
  platform/
web/
  src/features/...
  src/shared/...
db/
  migrations/
  seeds/
deployments/
  compose/
```

The API and worker are separate processes but reuse the same Go domain and application packages. PostgreSQL is the authoritative data store. S3-compatible storage, such as MinIO, stores private receipt images. The React client uses IndexedDB for its local mirror, pending outbox, cached queries, and conflict records.

## 5. Domain Boundaries

### 5.1 Identity

Identity owns Google OAuth callback handling, user provisioning, and signed application cookies. It stores only the provider subject, verified email, display metadata needed by the product, and application user identity. It does not retain Google access or refresh tokens.

Authentication state is a signed, secure, HTTP-only cookie with a configurable lifetime. There is no session table. Logout deletes the current browser cookie; logging out all devices is not supported.

### 5.2 Finance Core

Finance owns wallets, categories, wallet-category activation, transactions, transfers, balance adjustments, and receipt attachments.

Money is represented as integer minor units. VND uses a unit scale of `1`. Wallets still carry an ISO currency code so the data model is ready for a later multi-currency phase, but all initial net-worth and cross-wallet reports operate on VND only.

Confirmed transactions are the accounting source of truth. Balance changes and transaction persistence occur atomically. Transfers debit the source wallet and credit the destination wallet within one PostgreSQL transaction. A balance adjustment creates an explicit compensating income or expense transaction rather than rewriting history.

### 5.3 Planning

Planning owns budgets, budget threshold state, events/trips, recurring schedules, and debt/loan records. Recurring occurrences create drafts and notifications; they do not directly affect wallet balances.

### 5.4 Analytics

Analytics reads confirmed, reportable finance data. It calculates:

- Net worth across wallets with `include_in_total = true`.
- Income, expense, and net income for a period.
- Category and subcategory composition.
- Daily spending and daily average.
- Current-period versus previous-period change.
- Current cumulative spending versus the average cumulative curve of the previous three complete months.

`exclude_from_report = true` removes a transaction from spending and income analytics without removing it from wallet accounting.

### 5.5 Ingestion

Ingestion owns AI conversations, provider adapters, receipt extraction, bank webhook intake, normalization, deduplication, and transaction drafts. Every ingestion source converges on a shared `TransactionDraft` contract and the same confirmation service.

### 5.6 Synchronization

Synchronization owns client mutation idempotency, record versions, incremental change cursors, offline outbox replay, tombstones, and conflict resolution. It does not own domain validation; synchronized mutations call the same application services as online mutations.

### 5.7 Notifications

Notifications owns the in-app inbox, Web Push subscriptions, delivery attempts, and user-visible notices for budget thresholds, recurring drafts, bank imports, and items needing review.

### 5.8 Export

Export owns manually requested CSV and Google Sheets-compatible files. PostgreSQL remains authoritative; exported files are snapshots and are never imported back automatically.

### 5.9 Audit

Audit owns append-only records for state-changing actions and security events plus a hidden read-only debug page. UI clicks, page views, searches, and filter changes are not audited.

## 6. Core Data Model

All user-owned records include `user_id`, `created_at`, `updated_at`, and an optimistic `version` where offline mutation applies. UUIDs are generated client-side for offline-capable entities and server-side for server-only operational entities.

Core entities are:

| Entity | Purpose and critical constraints |
| --- | --- |
| `users` | Google subject and verified email are unique; no provider tokens or session rows. |
| `wallets` | Name, cash/bank/credit/e-wallet/savings/debt type, currency, VND balance, include-in-total, AI-default flag, archive state, and optional credit limit/statement/payment-due fields. At most one active default AI wallet per user. |
| `categories` | Expense, income, or debt/loan type; maximum depth is parent plus one child; system categories cannot be renamed or deleted. |
| `wallet_category_settings` | Enables or disables a category for one wallet without deleting category history. |
| `transactions` | Confirmed income, expense, transfer, or adjustment with positive amount, occurred time, note, with-person text, report flag, wallet/category references, optional destination wallet, event, and receipt. |
| `transaction_drafts` | Reviewable candidate with source, normalized fields, raw-text reference, validation status, confidence metadata, and confirmation link. |
| `receipt_objects` | Private S3 object key, content type, size, checksum, and owner. |
| `budgets` | Category or all-category amount; weekly, monthly, quarterly, yearly, or custom period; dates; and 80%/100% threshold state. |
| `events` | Named trip/event window and finished state. |
| `recurring_schedules` | Transaction template, daily/weekly/monthly/yearly frequency, next occurrence, active state, and timezone. |
| `debts` | Borrowed or lent principal, counterparty label, status, and related transactions. |
| `notifications` | In-app message, type, read state, and related entity. |
| `push_subscriptions` | User-owned Web Push endpoint and encrypted subscription keys. |
| `sync_changes` | Monotonic per-user change cursor, entity reference, version, operation, and timestamp. |
| `sync_conflicts` | Server snapshot, attempted client mutation, resolution state, and selected outcome. |
| `webhook_sources` | Named source, secret hash/encrypted secret reference, enabled state, and owner. |
| `webhook_events` | Source event ID, payload digest, replay metadata, processing status, and draft link. |
| `ai_conversations` | User-owned conversation metadata without provider secrets. |
| `ai_messages` | Text/image-reference messages and structured-result references. |
| `audit_logs` | Append-only actor/action/entity/outcome/diff metadata with no application update or delete API. |

System seed data contains the approved Vietnamese expense, income, and debt/loan category taxonomy from the source SRS. Seed identifiers are stable so migrations and tests can reference them safely.

## 7. Critical Business Rules

- Every user-owned lookup and mutation is scoped by the authenticated `user_id` in the backend.
- Amount input must be positive; transaction type determines accounting direction.
- A transfer requires distinct source and destination wallets owned by the same user.
- A transaction category must match the transaction type and be active for the selected wallet.
- AI uses the user's active default wallet only when the input does not resolve an explicit owned wallet.
- Wallets and categories referenced by history are archived rather than cascade-deleted.
- Only a confirmed draft can produce a confirmed transaction, and confirmation is idempotent.
- AI, OCR, webhook, and recurring inputs never update balances before confirmation.
- Financial mutations emit a sync-change record in the same database transaction.
- Analytics uses `Asia/Ho_Chi_Minh` boundaries unless a later user-timezone feature changes the rule.
- Percentage comparison against a zero previous period returns an explicit `no_baseline` state rather than dividing by zero.
- Budget thresholds emit at most one notification per budget, threshold, and period.

## 8. Offline Synchronization

### 8.1 Local Write

The React application writes an optimistic record and a mutation envelope to IndexedDB. The envelope contains a client-generated mutation ID, entity ID, operation, base version, local timestamp, and payload. The UI immediately reads the optimistic state.

### 8.2 Server Reconciliation

When connectivity is available, the client sends ordered outbox batches. The Go API checks the mutation ID for idempotency, validates the base version, applies domain rules, commits the mutation and sync-change row atomically, and returns the authoritative record and version.

### 8.3 Conflict Policy

If the base version is stale, the API returns a conflict response containing the server state and safe metadata about the rejected client mutation. It does not overwrite the server record. The client stores a conflict item and lets the user keep the server state, retry an edited client state against the latest version, or discard the local mutation.

### 8.4 Incremental Pull

The client requests changes after its last per-user cursor. Deletes and archives are represented by tombstones. A full resync is available when the cursor is invalid or local data is reset.

## 9. Ingestion Flows

### 9.1 Text AI Chat

The API receives a user message, loads only the minimum relevant wallet and category context, and calls an OpenAI-compatible endpoint. The provider must return schema-constrained structured output for one or more income, expense, or transfer proposals. The backend validates the provider output, resolves exact owned entities, and creates drafts. The conversation renders draft cards that the user can edit and confirm.

The application plans and implements AI functions, prompts, schemas, validation, and provider integration. It does not train a model.

### 9.2 Receipt Camera

From the add-transaction flow, the browser captures or selects an image and uploads it to a private S3 object using a short-lived presigned URL. The API invokes the configured OCR adapter, normalizes the extracted text, and creates a draft for review. Provider failures leave the image available for retry but create no transaction.

### 9.3 Image in AI Chat

From an active AI conversation, the user attaches an image and asks to add a transaction. The API supplies the image to the OpenAI-compatible multimodal endpoint and requests the same structured proposal schema used by text chat. Valid proposals become conversation-linked drafts. This is separate from the receipt-camera OCR path.

### 9.4 Bank Notification Webhook

An external automation sends notification text, source event ID, timestamp, nonce, and HMAC signature. The API validates the per-source secret, timestamp window, nonce, and payload digest before accepting the event. Duplicate source event IDs or payload digests do not create duplicate drafts. A successful import creates a review notice through the in-app inbox and Web Push.

## 10. Authentication and Security

- Google OAuth is the only initial login provider.
- OAuth state and nonce are verified during callback handling.
- The application cookie is signed, `HttpOnly`, `Secure`, and `SameSite=Lax`; its lifetime is configurable.
- No Google access or refresh token is persisted.
- State-changing cookie-authenticated requests require CSRF protection appropriate to the selected HTTP framework.
- Provider credentials, database strings, signing keys, webhook secrets, and S3 credentials come from environment or mounted secrets, never source control.
- Receipt objects are private and accessed through short-lived presigned URLs.
- Uploads are restricted by allowed content type, maximum size, checksum, and decoded-image validation.
- API authorization is enforced server-side for every object; frontend route hiding is never treated as authorization.
- Account deletion is an authenticated, confirmed, audited background operation that deletes database-owned data and associated S3 objects.

The PWA presents overview, transaction history, central quick-add, budget, and account destinations. Overview shows net worth, included wallets, report summaries, and recent transactions. Users can hide displayed balances without changing stored values, search transactions, and switch analytics between all included wallets and one owned wallet. Account settings expose profile, wallets, category management, exports, account reset, and account deletion; the audit page is intentionally absent.

## 11. Hidden Audit Log

The application records all state-changing domain actions and security events, including:

- Login success/failure and logout.
- Create, update, archive, restore, confirm, reset, delete, and export actions.
- Offline synchronization acceptance, rejection, and conflict resolution.
- AI/OCR/webhook/worker ingestion attempts and outcomes.
- Webhook authentication failures and replay rejection.
- Provider failures relevant to debugging.
- Account reset and deletion lifecycle events.

Each audit record contains actor user ID when known, action code, entity type and ID, redacted before/after diff, source, outcome, request/correlation ID, timestamp, and safe error metadata. Audit storage excludes cookies, OAuth tokens, API keys, webhook secrets, raw receipt bytes, and unredacted provider payloads.

The page is mounted at an undocumented internal route and is absent from navigation. Read access requires both normal authentication and an exact match between the verified Google email and `AUDIT_VIEWER_EMAIL`. The API repeats this authorization check for every query. The page is read-only and supports time, actor, action, entity, outcome, and correlation-ID filters.

Audit records are append-only through application interfaces. Retention is configured by `AUDIT_RETENTION_DAYS` with a default of `180`. The worker deletes expired rows in bounded batches and records a retention-run summary without recursively auditing each deleted audit row.

## 12. Background Work and Notifications

The worker handles recurring occurrences, budget threshold evaluation, Web Push delivery, provider retries, account deletion, audit retention, and other bounded maintenance jobs. PostgreSQL leases or advisory locks prevent concurrent workers from processing the same occurrence. Every occurrence uses a deterministic idempotency key.

Retries use capped exponential backoff and a terminal failure state. Retryable provider outages are distinct from validation errors. In-app notifications remain the durable user-visible record even when Web Push permission is denied or delivery fails.

## 13. Error Contract

The API returns stable machine-readable error codes with a correlation ID. The frontend distinguishes at least:

- Authentication required or forbidden.
- Validation failure.
- Record not found within the current user scope.
- Optimistic-version conflict.
- Duplicate/idempotent request.
- Provider unavailable or timed out.
- Upload rejected.
- Webhook signature, replay, or deduplication rejection.
- Rate limit.
- Internal failure safe to retry or unsafe to retry.

Raw provider errors and stack traces are never returned to the browser. Structured logs and audit records use the shared correlation ID.

## 14. Verification Strategy

### 14.1 Backend

- Unit tests cover money arithmetic, balance effects, transfers, adjustments, category rules, budgets, recurring schedules, analytics, AI schema validation, and audit redaction.
- PostgreSQL integration tests cover migrations, per-user isolation, atomic transfers, idempotency, optimistic conflicts, change cursors, webhook replay protection, audit authorization, audit retention, and worker locks.
- Adapter contract tests use fakes for Google OAuth, OpenAI-compatible AI, OCR, S3, and Web Push.

### 14.2 Frontend

- Component tests cover forms, category trees, draft review, AI proposal cards, analytics states, audit filters, and conflict resolution.
- IndexedDB tests cover hydration, optimistic updates, outbox replay, reconnect, tombstones, full resync, and conflicting edits.
- Route tests prove the audit route is absent from navigation and unauthorized accounts receive a forbidden response from the API.

### 14.3 End-to-End and Security

- Browser tests cover OAuth callback fixtures, online and offline transactions, transfers, draft confirmation, receipt OCR, AI-chat images, bank webhook review, budget alerts, analytics, export, and account deletion.
- Security tests cover cross-user object access, CSRF, cookie flags, HMAC tampering, webhook replay, unsafe uploads, malformed AI output, provider prompt injection boundaries, and audit secret redaction.
- Operational tests cover Compose startup, migrations, health checks, worker restart/idempotency, S3 access, and backup/restore smoke tests.
- UAT covers Vietnamese text, VND formatting, `Asia/Ho_Chi_Minh` period boundaries, installed mobile PWA behavior, offline recovery, conflict review, and Web Push permission denial.

## 15. Delivery Decomposition

The full SRS is implemented through linked plans rather than one oversized execution artifact:

1. **Platform and Identity:** repository foundation, React PWA, Go API/worker, PostgreSQL, Google OAuth, S3 adapter, deployment and observability.
2. **Finance Core:** wallets, categories, Vietnamese seeds, transactions, transfers, adjustments, attachments, and archive behavior.
3. **Offline Sync:** IndexedDB mirror, outbox, idempotent mutation API, change cursor, versions, tombstones, and conflict review.
4. **Planning and Automation:** budgets, alerts, events, recurring drafts, debt/loan tracking, in-app notifications, and Web Push.
5. **Analytics and Dashboard:** navigation, dashboard, wallet selector, reports, comparisons, and trends.
6. **AI, Receipt, and Bank Ingestion:** text chat, shared drafts, receipt OCR, AI-chat images, signed webhook intake, and provider adapters.
7. **Audit, Export, Account Lifecycle, and Production Operations:** hidden audit page, retention, manual export, reset/delete, backups, health checks, and release proof.
8. **Deferred Voice Input:** audio capture, OpenAI-compatible transcription, and transcript-to-draft behavior after the initial release.

Each phase must remain deployable and testable. Later phases consume public application interfaces from earlier phases rather than bypassing their domain rules.

## 16. Reconciliation Plan

Implementation phases must keep these durable documents synchronized with actual behavior:

- `docs/requirements/` for accepted product behavior and UAT criteria.
- `docs/architecture/ARCHITECTURE.md` for runtime and module boundaries.
- `docs/architecture/API.md` for REST, sync, webhook, errors, and authentication contracts.
- `docs/architecture/ERD.md` for tables, relationships, ownership, versions, and deletion behavior.
- `docs/architecture/INTEGRATIONS.md` for Google OAuth, OpenAI-compatible AI, OCR, S3, Web Push, and webhook integrations.
- `docs/architecture/SDD.md` for approved system design as implementation evolves.
- `docs/decisions/` for durable choices affecting authentication, offline sync, financial accounting, audit authorization, and deployment.
- `docs/work/VALIDATION_MATRIX.md`, `docs/CONTEXT.md`, backlog/work status, and release notes for evidence and continuity.

## 17. Approved Decisions Summary

- Frontend: React TypeScript PWA, not Next.js.
- Backend: Go modular monolith with a separate worker process.
- Database: PostgreSQL hosted in the user's homelab production environment.
- Storage: S3-compatible private object storage.
- Authentication: Google OAuth with a stateless signed cookie and no session table.
- Users: multi-user, with strict row ownership.
- Currency: VND-first with a multi-currency-ready schema.
- Offline: full read/write with explicit optimistic-conflict review.
- AI: OpenAI-compatible endpoint; all outputs are reviewable drafts.
- Receipt entry: external OCR API from the add-transaction camera flow.
- AI image entry: direct multimodal AI call from an AI conversation.
- Bank notification entry: secured webhook from an external sender.
- Voice: deferred.
- Google Sheets: manual export only.
- Alerts: in-app inbox and Web Push.
- Audit: append-only state/security events, hidden page restricted by one configured Google email, 180-day default retention.
