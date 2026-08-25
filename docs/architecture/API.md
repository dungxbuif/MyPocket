---
artifact_type: api_contract_index
id: API-MASTER
status: active
owner: shared
human_fields: [contract_intent, approval]
ai_fields: [contract_rows, errors, auth_notes, versioning, linked_decisions]
shared_fields: [status, trace]
updated: 2026-08-24
---

# API

## Field Ownership

- Human-approved contract intent is recorded in the product spec and system design.
- AI maintains endpoint families, errors, authentication notes, versioning, and ticket-level schemas.

## Contract Rules

- Base prefix: `/api/v1`.
- Browser contracts use JSON except presigned object upload and generated export download.
- Authenticated finance routes derive `user_id` from the signed cookie and never accept ownership from request payloads.
- State-changing requests require CSRF protection, an idempotency key where retryable, and optimistic `base_version` where offline-capable.
- Detailed request and response schemas are added by the owning phase before implementation.

## API Surface

| Contract | Type | Auth | Status | Notes |
| --- | --- | --- | --- | --- |
| `GET /api/v1/auth/google`, `GET /api/v1/auth/google/callback`, `POST /api/v1/auth/logout`, `GET /api/v1/me` | OAuth/HTTP | Mixed | implemented | Fixture callback is available only when explicitly enabled; sets stateless signed `mypocket_auth` cookie and browser-readable `mypocket_csrf`; no session API |
| `/wallets` | REST collection | User | planned | CRUD/archive and default-AI selection |
| `/categories`, `/wallets/{id}/categories` | REST collection | User | planned | Two-level taxonomy and activation |
| `/transactions` | REST collection | User | planned | Income, expense, transfer, adjustment, search |
| `/receipts/uploads`, `/receipts/{id}` | REST/object | User | planned | Presigned upload and private metadata |
| `POST /sync/mutations`, `GET /sync/changes` | Sync | User | planned | Idempotent batches, cursors, versions, tombstones |
| `/sync/conflicts` | REST collection | User | planned | Read and resolve explicit conflicts |
| `/budgets`, `/events`, `/recurring-schedules`, `/debts` | REST collection | User | planned | Planning and automation |
| `/analytics/*` | Query REST | User | planned | Dashboard, categories, periods, trends |
| `/ai/conversations`, `/ai/conversations/{id}/messages` | REST collection | User | planned | Text and multimodal chat |
| `POST /receipts/{id}/extract` | Command | User | planned | External OCR path from add transaction |
| `/transaction-drafts`, `POST /transaction-drafts/{id}/confirm` | REST/command | User | planned | Shared review and idempotent confirmation |
| `POST /webhooks/bank/{source_id}` | Webhook | HMAC | planned | Timestamp, nonce, signature, deduplication |
| `/notifications`, `/push-subscriptions` | REST collection | User | planned | Inbox and Web Push |
| `POST /exports`, `GET /exports/{id}` | Job REST | User | planned | Manual snapshot export |
| `POST /account/reset`, `DELETE /account` | Command | User | planned | Confirmed background deletion |
| `GET /internal/audit-logs` | Query REST | Audit viewer | planned | Hidden read-only filtered audit access |
| `GET /api/v1/health/live`, `GET /api/v1/health/ready` | Operations | None/internal | implemented | Liveness is process-only; readiness returns `503 INTERNAL_RETRYABLE` when the required database dependency is unavailable and never exposes sensitive dependency details |

## Errors

| Error | Meaning | Consumer Impact |
| --- | --- | --- |
| `AUTH_REQUIRED` | No valid application cookie | Redirect to Google login |
| `FORBIDDEN` | Authenticated principal lacks access | Show forbidden state; do not retry |
| `VALIDATION_FAILED` | Stable field-level domain validation failed | Display correctable fields |
| `NOT_FOUND` | Object absent within authenticated scope | Display missing/removed state |
| `VERSION_CONFLICT` | `base_version` is stale | Store conflict and request user resolution |
| `IDEMPOTENT_REPLAY` | Mutation was already applied | Reconcile returned authoritative result |
| `PROVIDER_UNAVAILABLE` | AI/OCR/Push provider failed transiently | Keep draft/input retryable |
| `UPLOAD_REJECTED` | File failed size/type/checksum/decode rules | Ask for a valid image |
| `WEBHOOK_REJECTED` | Signature, timestamp, nonce, or replay check failed | Sender must not treat event as imported |
| `RATE_LIMITED` | Caller exceeded a bounded operation limit | Retry after returned interval |
| `INTERNAL_RETRYABLE` | Safe transient internal failure | Retry with the same idempotency key |
| `INTERNAL_FAILURE` | Unclassified failure | Show correlation ID; do not expose internals |

## Authentication And CSRF

- Google OAuth state and nonce are verified at callback.
- The application cookie is signed, `HttpOnly`, `Secure`, and `SameSite=Lax`.
- No Google provider token is stored.
- Cookie-authenticated mutations require the framework-selected CSRF token/header contract.
- Cookie-authenticated mutations send `X-CSRF-Token`, matching the browser-readable `mypocket_csrf` cookie.
- Audit queries additionally require verified email equality with `AUDIT_VIEWER_EMAIL`.
- Webhooks authenticate through a per-source HMAC contract rather than the browser cookie.

## Versioning

- The initial contract is `v1` under the URL prefix.
- Additive response fields are backward-compatible.
- Removing or changing field meaning requires a new version or explicit coordinated client migration.
- IndexedDB schema and API versions are coordinated by the Offline Sync phase.
- Every error response contains a stable code and correlation ID.

## Linked Decisions

- [ADR-002](../decisions/ADR-002-stateless-google-oauth.md)
- [ADR-003](../decisions/ADR-003-offline-sync-conflict-review.md)
- [ADR-004](../decisions/ADR-004-review-first-ingestion.md)
- [ADR-005](../decisions/ADR-005-restricted-audit-log.md)
