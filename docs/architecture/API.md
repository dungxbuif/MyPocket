---
artifact_type: api_contract_index
id: API-MASTER
status: active
owner: shared
human_fields: [contract_intent, approval]
ai_fields: [contract_rows, errors, auth_notes, versioning, linked_decisions]
shared_fields: [status, trace]
updated: 2026-08-31
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
| `GET /wallets`, `POST /wallets`, `PATCH /wallets/{id}`, `POST /wallets/{id}/archive`, `POST /wallets/{id}/default-ai` | REST/command | User | implemented | Lists, creates, edits, archives, and selects the default AI wallet for the authenticated user; mutations require CSRF; `user_id` is always derived from the signed cookie |
| `GET /categories` | REST collection | User | implemented | Lists system and authenticated-user categories; `system_key` is returned for locked system rows |
| `POST /categories`, `PATCH /categories/{id}`, `POST /categories/{id}/archive`, `PUT /wallets/{wallet_id}/categories/{category_id}` | REST/command | User | implemented | Creates user categories, edits/archives user-owned categories, rejects system category mutation, and toggles category activation for a user-owned wallet |
| `GET /transactions`, `POST /transactions`, `PATCH /transactions/{id}`, `POST /transactions/{id}/archive` | REST/command | User | implemented | Income, expense, transfer, adjustment, edit/archive reversal, search filters, and idempotent creates |
| `/receipts/uploads`, `/receipts/{id}` | REST/object | User | planned | Presigned upload and private metadata |
| `POST /sync/mutations`, `GET /sync/changes`, `POST /sync/resync` | Sync | User | implemented | Idempotent batches, per-user cursors, versions, tombstones, and authoritative resync snapshots |
| `/sync/conflicts` | REST collection | User | planned | Optional server-backed conflict collection; current PHASE-003 conflict review is persisted client-side from sync mutation results |
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

## Wallet And Category Schemas

### `GET /api/v1/wallets`

Response:

```json
{
  "status": "ok",
  "wallets": [
    {
      "id": "uuid",
      "name": "Tiền mặt",
      "type": "cash",
      "balance_vnd": 120000,
      "include_in_total": true,
      "is_default_ai": false,
      "version": 1
    }
  ],
  "correlation_id": "req_..."
}
```

### `POST /api/v1/wallets`

Headers:

- `X-CSRF-Token`: must match the `mypocket_csrf` cookie.

Request:

```json
{
  "name": "Ngân hàng",
  "type": "bank",
  "credit_limit_vnd": null,
  "statement_day": null,
  "payment_due_day": null
}
```

Response status: `201 Created`.

### `PATCH /api/v1/wallets/{id}`

Headers:

- `X-CSRF-Token`: must match the `mypocket_csrf` cookie.

Request:

```json
{
  "name": "Ví chính",
  "include_in_total": true
}
```

Response status: `200 OK`; body uses the single-wallet envelope from create.

### `POST /api/v1/wallets/{id}/archive`

Headers:

- `X-CSRF-Token`: must match the `mypocket_csrf` cookie.

Response status: `200 OK`; archived wallets are removed from ordinary list results and cannot remain the default AI wallet.

### `POST /api/v1/wallets/{id}/default-ai`

Headers:

- `X-CSRF-Token`: must match the `mypocket_csrf` cookie.

Response status: `200 OK`; the backend clears any previous active default AI wallet for the authenticated user.

### `GET /api/v1/categories`

Response:

```json
{
  "status": "ok",
  "categories": [
    {
      "id": "uuid",
      "kind": "expense",
      "name": "Ăn uống",
      "system_key": "expense_food",
      "is_system": true,
      "version": 1
    }
  ],
  "correlation_id": "req_..."
}
```

### `POST /api/v1/categories`

Headers:

- `X-CSRF-Token`: must match the `mypocket_csrf` cookie.

Request:

```json
{
  "kind": "expense",
  "name": "Cafe"
}
```

Response status: `201 Created`; body contains `{ "category": { ... } }`.

### `PATCH /api/v1/categories/{id}`

Headers:

- `X-CSRF-Token`: must match the `mypocket_csrf` cookie.

Request:

```json
{
  "name": "Cafe"
}
```

Response status: `200 OK`; system categories return `VALIDATION_FAILED`.

### `POST /api/v1/categories/{id}/archive`

Headers:

- `X-CSRF-Token`: must match the `mypocket_csrf` cookie.

Response status: `200 OK`; system categories return `VALIDATION_FAILED`.

### `PUT /api/v1/wallets/{wallet_id}/categories/{category_id}`

Headers:

- `X-CSRF-Token`: must match the `mypocket_csrf` cookie.

Request:

```json
{
  "active": false
}
```

Response status: `200 OK`; activation is scoped to the authenticated user's wallet and the visible system/user category.

## Sync Schemas

### `POST /api/v1/sync/mutations`

Headers:

- `X-CSRF-Token`: must match the `mypocket_csrf` cookie.

Request:

```json
{
  "mutations": [
    {
      "mutation_id": "uuid",
      "device_id": "browser-device-id",
      "sequence": 1,
      "entity_type": "transaction",
      "entity_id": "uuid",
      "operation": "create",
      "base_version": 0,
      "payload": {
        "type": "expense",
        "source_wallet_id": "uuid",
        "category_id": "uuid",
        "amount_vnd": 45000,
        "occurred_at": "2026-08-31T00:00:00Z",
        "note": "Cafe"
      }
    }
  ]
}
```

Response status: `200 OK`; each result has `state` of `applied`, `replayed`, `rejected`, or `conflict`. Duplicate `mutation_id` with the same request hash replays the stored result; a different request hash is rejected without reapplying accounting. A `conflict` result includes the entity identity, operation, stale `base_version`, current server version, the local payload, and the authoritative server payload so the PWA can persist an explicit local conflict entry for keep-server, discard-local, or edit-and-retry recovery.

### `GET /api/v1/sync/changes`

Query filters:

- `after`: last observed cursor, defaults to `0`.
- `limit`: maximum change rows, defaults to `100` and is capped by the service.

Response:

```json
{
  "status": "ok",
  "changes": [
    {
      "cursor": 1,
      "entity_type": "transaction",
      "entity_id": "uuid",
      "operation": "create",
      "version": 1,
      "payload": {}
    }
  ],
  "next_cursor": 1,
  "correlation_id": "req_..."
}
```

The feed is scoped to the authenticated user and ordered by monotonic per-user cursor.

### `POST /api/v1/sync/resync`

Headers:

- `X-CSRF-Token`: must match the `mypocket_csrf` cookie.

Response status: `200 OK`; body contains `snapshot.wallets`, `snapshot.categories`, `snapshot.transactions`, and `snapshot.next_cursor` from authoritative PostgreSQL state.

## Transaction Schemas

### `GET /api/v1/transactions`

Query filters:

- `wallet_id`
- `category_id`
- `type`: `income`, `expense`, `transfer`, or `adjustment`
- `date_from`, `date_to`: RFC3339 timestamps
- `q`: searches note, with-person, and event reference
- `excluded_from_reports`: boolean
- `include_archived`: boolean

Response:

```json
{
  "status": "ok",
  "transactions": [
    {
      "id": "uuid",
      "type": "expense",
      "source_wallet_id": "uuid",
      "category_id": "uuid",
      "amount_vnd": 45000,
      "balance_after_vnd": 955000,
      "occurred_at": "2026-08-30T10:00:00Z",
      "note": "Cafe",
      "with_person": "",
      "event_ref": "",
      "excluded_from_reports": false,
      "version": 1
    }
  ],
  "correlation_id": "req_..."
}
```

### `POST /api/v1/transactions`

Headers:

- `X-CSRF-Token`: must match the `mypocket_csrf` cookie.
- `Idempotency-Key`: required for retry-safe create.

Request:

```json
{
  "type": "income",
  "source_wallet_id": "uuid",
  "destination_wallet_id": "",
  "category_id": "uuid",
  "amount_vnd": 500000,
  "target_balance_vnd": null,
  "occurred_at": "2026-08-30T10:00:00Z",
  "note": "Lương",
  "with_person": "",
  "event_ref": "",
  "excluded_from_reports": false
}
```

Response status: `201 Created`.

### `PATCH /api/v1/transactions/{id}`

Headers:

- `X-CSRF-Token`: must match the `mypocket_csrf` cookie.

Request uses the same body shape as `POST /api/v1/transactions`; the backend reverses the old stored deltas and reapplies the new effect in one database transaction.

### `POST /api/v1/transactions/{id}/archive`

Headers:

- `X-CSRF-Token`: must match the `mypocket_csrf` cookie.

Response status: `200 OK`; the backend reverses the stored balance effect exactly once.

## Authentication And CSRF

- Google OAuth state and nonce are verified at callback.
- The application cookie is signed, `HttpOnly`, `Secure`, and `SameSite=Lax`.
- No Google provider token is stored.
- Cookie-authenticated mutations require the framework-selected CSRF token/header contract.
- Cookie-authenticated mutations send `X-CSRF-Token`, matching the browser-readable `mypocket_csrf` cookie.
- Local and preview browser calls may use credentialed CORS only from the configured `PUBLIC_WEB_URL`; Compose serves same-origin `/api/` through the web nginx proxy.
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
