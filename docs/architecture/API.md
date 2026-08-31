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
| `GET /api/v1/api-keys`, `POST /api/v1/api-keys`, `POST /api/v1/api-keys/{id}/revoke` | REST/command | User | implemented | Browser-authenticated users can create/list/revoke API keys from the Account tab or API for third-party systems and AI agents; plaintext `mpk_...` key is returned only once on create; key management requires browser cookie + CSRF |
| `GET /wallets`, `POST /wallets`, `PATCH /wallets/{id}`, `POST /wallets/{id}/archive`, `POST /wallets/{id}/default-ai` | REST/command | User | implemented | Lists, creates, edits, archives, and selects the default AI wallet for the authenticated user; mutations require CSRF; `user_id` is always derived from the signed cookie |
| `GET /categories` | REST collection | User | implemented | Lists system and authenticated-user categories; `system_key` is returned for locked system rows |
| `POST /categories`, `PATCH /categories/{id}`, `POST /categories/{id}/archive`, `PUT /wallets/{wallet_id}/categories/{category_id}` | REST/command | User | implemented | Creates user categories, edits/archives user-owned categories, rejects system category mutation, and toggles category activation for a user-owned wallet |
| `GET /transactions`, `POST /transactions`, `PATCH /transactions/{id}`, `POST /transactions/{id}/archive` | REST/command | User | implemented | Income, expense, transfer, adjustment, edit/archive reversal, search filters, and idempotent creates |
| `POST /api/v1/files/presign`, `GET /api/v1/files/{id}/download` | REST/object | User | implemented | Generic file contract validates image metadata (JPEG/PNG/WebP, max 15 MiB), creates user-scoped file metadata, returns a 15-minute private S3 presigned upload URL, and returns a 10-minute download URL. Receipt attachment is a consumer of the returned file ID. |
| `POST /sync/mutations`, `GET /sync/changes`, `POST /sync/resync` | Sync | User | implemented | Idempotent batches, per-user cursors, versions, tombstones, and authoritative resync snapshots |
| `/sync/conflicts` | REST collection | User | planned | Optional server-backed conflict collection; current PHASE-003 conflict review is persisted client-side from sync mutation results |
| `GET /budgets`, `POST /budgets`, `PATCH /budgets/{id}`, `POST /budgets/{id}/archive` | REST/command | User | implemented | Budget CRUD, selected/all expense category scopes, Ho Chi Minh period progress, and 80%/100% alert dedupe |
| `GET /events`, `POST /events`, `PATCH /events/{id}`, `POST /events/{id}/archive`, `POST /events/{id}/transactions/{transaction_id}` | REST/command | User | implemented | Event CRUD and links to owned confirmed transactions; totals preserve report exclusion |
| `GET /obligations`, `POST /obligations`, `PATCH /obligations/{id}`, `POST /obligations/{id}/archive`, `POST /obligations/{id}/repayments/{transaction_id}` | REST/command | User | implemented | Borrow/lend obligation CRUD and repayment links with overpayment protection |
| `GET /recurring-schedules`, `POST /recurring-schedules`, `POST /recurring-schedules/{id}/archive` | REST/command | User | implemented | Recurring schedule setup for deterministic draft generation |
| `GET /dashboard`, `GET /reports/{cash-flow,categories,daily,comparison,cumulative,insider}`, `GET /search` | Query REST | User | implemented | Authenticated server aggregates, normalized Ho Chi Minh date filters, bounded grouped search, generated metadata and data version; Insider Home summary selects the most frequent expense category and compares daily averages |
| `GET /assets`, `POST /assets`, `GET /assets/{id}`, `POST /assets/{id}/archive`, `POST /assets/{id}/trades`, `PATCH /assets/{id}/trades/{trade_id}`, `POST /assets/{id}/trades/{trade_id}/archive`, `POST /assets/{id}/prices`, `GET /portfolio/summary` | REST/command | User | implemented | User-owned market-valued assets, moving-average buy/sell trades, append-only manual/provider VND price snapshots, archive retention, optimistic versions, and separate investment totals |
| `/ai/conversations`, `/ai/conversations/{id}/messages` | REST collection | User | planned | Text and multimodal chat |
| `POST /receipts/{id}/extract` | Command | User | planned | External OCR path from add transaction |
| `GET /transaction-drafts`, `POST /transaction-drafts/{id}/confirm` | REST/command | User | partial | Draft listing implemented; confirmation remains planned |
| `POST /webhooks/bank/{source_id}` | Webhook | HMAC | planned | Timestamp, nonce, signature, deduplication |
| `GET /notifications`, `PATCH /notifications/{id}/read`, `POST /push-subscriptions`, `DELETE /push-subscriptions/{id}` | REST collection | User | implemented | Cursor-bounded inbox, ownership checks, private subscription keys, best-effort worker delivery |
| `POST /exports`, `GET /exports/{id}` | Job REST | User | planned | Manual snapshot export |
| `POST /account/reset`, `DELETE /account` | Command | User | planned | Confirmed background deletion |
| `GET /api/v1/audit/access`, `GET /api/v1/audit/events` | Query REST | Audit viewer | implemented | Access check and read-only filtered audit access; both require signed-cookie auth and exact verified `AUDIT_VIEWER_EMAIL`; the Account UI only renders the log panel after the access check succeeds |
| `GET /api/v1/health/live`, `GET /api/v1/health/ready` | Operations | None/internal | implemented | Liveness is process-only; readiness returns `503 INTERNAL_RETRYABLE` when the required database dependency is unavailable and never exposes sensitive dependency details |

## Errors

| Error | Meaning | Consumer Impact |
| --- | --- | --- |
| `AUTH_REQUIRED` | No valid application cookie | Redirect to Google login |
| `FORBIDDEN` | Authenticated principal lacks access | Show forbidden state; do not retry |
| `VALIDATION_FAILED` | Stable field-level domain validation failed | Display correctable fields |
| `NOT_FOUND` | Object absent within authenticated scope | Display missing/removed state |
| `VERSION_CONFLICT` | `base_version` is stale | Store conflict and request user resolution |
| `ASSET_OVERSELL` | Asset sell quantity exceeds current position quantity | Ask the user to reduce quantity or add/correct buy history |
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

## Asset Portfolio Schemas

### `GET /api/v1/assets`

Response:

```json
{
  "status": "ok",
  "assets": [
    {
      "id": "uuid",
      "type": "gold",
      "symbol": "SJC",
      "exchange": "",
      "name": "SJC 9999",
      "unit": "tael",
      "reporting_currency": "VND",
      "pricing_mode": "manual",
      "include_in_net_worth": true,
      "version": 3,
      "summary": {
        "quantity": "2.5",
        "cost_basis_vnd": 180000000,
        "realized_pnl_vnd": 0,
        "current_unit_price_vnd": 75000000,
        "market_value_vnd": 187500000,
        "unrealized_pnl_vnd": 7500000,
        "unrealized_pnl_percent": "4.1667",
        "unrealized_not_comparable": false,
        "valuation_status": "current"
      }
    }
  ],
  "correlation_id": "req_..."
}
```

### `POST /api/v1/assets`

Headers:

- `X-CSRF-Token`: must match the `mypocket_csrf` cookie.

Request:

```json
{
  "type": "stock",
  "symbol": "FPT",
  "exchange": "HOSE",
  "name": "FPT",
  "unit": "share",
  "pricing_mode": "automatic",
  "provider_key": "static",
  "provider_symbol": "FPT",
  "include_in_net_worth": true
}
```

Response status: `201 Created`; body contains `{ "asset": { ... } }`.

### `POST /api/v1/assets/{id}/trades`

Request:

```json
{
  "side": "buy",
  "quantity": "10",
  "unit_price_vnd": 120000,
  "fee_vnd": 15000,
  "occurred_at": "2026-08-31T10:00:00Z",
  "note": "",
  "base_version": 1
}
```

Response status: `201 Created`; body contains the recalculated asset. `sell` rejects overselling with `ASSET_OVERSELL`.

### `POST /api/v1/assets/{id}/prices`

Request:

```json
{
  "unit_price_vnd": 75000000,
  "priced_at": "2026-08-31T10:00:00Z",
  "source": "manual",
  "base_version": 2
}
```

Response status: `201 Created`; price history is append-only and latest `priced_at` wins valuation.

### `GET /api/v1/portfolio/summary`

Response:

```json
{
  "status": "ok",
  "summary": {
    "investment_market_value_vnd": 187500000,
    "missing_price_count": 0,
    "included_position_count": 1,
    "position_count": 1
  },
  "correlation_id": "req_..."
}
```

### `GET /api/v1/categories`

Returns the seeded Vietnamese parent/child system catalog from migration `0011_phase002_category_catalog.sql` together with the caller's own categories.

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

### `GET /api/v1/budgets`

Returns authenticated-user active budgets with current period progress computed in `Asia/Ho_Chi_Minh`.

```json
{
  "status": "ok",
  "budgets": [
    {
      "budget": {
        "id": "uuid",
        "name": "Ăn uống",
        "period_type": "monthly",
        "amount_vnd": 500000,
        "category_ids": ["uuid"],
        "all_categories": false,
        "version": 1
      },
      "period_start": "2026-08-01",
      "period_end": "2026-08-31",
      "spent_vnd": 410000,
      "remaining_vnd": 90000,
      "percent": 82,
      "alert_80": true,
      "alert_100": false
    }
  ],
  "correlation_id": "req_..."
}
```

### `POST /api/v1/budgets`

Headers:

- `X-CSRF-Token`: must match the `mypocket_csrf` cookie.

Request fields: `name`, `period_type` (`weekly`, `monthly`, `quarterly`, `yearly`, or `custom`), `amount_vnd`, optional `category_ids`, and `custom_start`/`custom_end` as `YYYY-MM-DD` for custom periods. Empty `category_ids` means all expense categories.

Response status: `201 Created`.

### `PATCH /api/v1/budgets/{id}`

Same JSON shape as create. Mutations are scoped to the authenticated owner and return `403 FORBIDDEN` for another user's budget or category scope.

### `POST /api/v1/budgets/{id}/archive`

Archives the budget for the authenticated owner. Archived budgets are hidden from `GET /api/v1/budgets`.

### `GET /api/v1/events`

Returns authenticated-user active events with totals from linked, non-archived, report-included transactions.

### `POST /api/v1/events`

Headers:

- `X-CSRF-Token`: must match the `mypocket_csrf` cookie.

Request fields: `name`, `starts_on`, `ends_on`, and optional `note`. Dates use `YYYY-MM-DD`. Response status: `201 Created`.

### `PATCH /api/v1/events/{id}`

Same JSON shape as create. Mutations are scoped to the authenticated owner.

### `POST /api/v1/events/{id}/archive`

Archives the event for the authenticated owner. Archived events are hidden from `GET /api/v1/events`.

### `POST /api/v1/events/{id}/transactions/{transaction_id}`

Links an owned confirmed transaction to an owned active event. The link does not change wallet accounting. Linked transactions marked `excluded_from_reports` are not counted in event totals.

### `GET /api/v1/obligations`

Returns authenticated-user active borrow/lend obligations with `repaid_vnd` and `remaining_vnd` derived from linked repayment transactions.

### `POST /api/v1/obligations`

Headers:

- `X-CSRF-Token`: must match the `mypocket_csrf` cookie.

Request fields: `direction` (`borrowed` or `lent`), `principal_vnd`, `counterparty`, `due_on`, and optional `note`. Response status: `201 Created`.

### `PATCH /api/v1/obligations/{id}`

Same JSON shape as create. Mutations are scoped to the authenticated owner.

### `POST /api/v1/obligations/{id}/archive`

Archives the obligation for the authenticated owner. Archived obligations are hidden from `GET /api/v1/obligations`.

### `POST /api/v1/obligations/{id}/repayments/{transaction_id}`

Links an owned confirmed transaction as a repayment. The command returns `400 VALIDATION_FAILED` when the linked repayment total would exceed the obligation principal.

### `GET /api/v1/recurring-schedules`

Returns authenticated-user active recurring schedules with normalized `frequency`, `timezone`, `starts_at`, `next_occurs_at`, wallet/category references, integer `amount_vnd`, and draft payload fields.

### `POST /api/v1/recurring-schedules`

Headers:

- `X-CSRF-Token`: must match the `mypocket_csrf` cookie.

Request fields: `name`, `frequency` (`daily`, `weekly`, or `monthly`), `timezone`, `starts_at` as RFC3339, `type`, `source_wallet_id`, optional `destination_wallet_id`, optional `category_id`, `amount_vnd`, and optional `note`. Response status: `201 Created`.

### `POST /api/v1/recurring-schedules/{id}/archive`

Archives the recurring schedule for the authenticated owner. Existing generated drafts remain visible for review.

### `GET /api/v1/transaction-drafts`

Returns authenticated-user transaction drafts generated by recurring schedules. Drafts are reviewable records and do not modify wallet balances until an explicit confirmation API is added.

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
- Third-party callers can authenticate with `Authorization: Bearer mpk_...`; API key requests resolve to the owning user and do not require browser CSRF for business API mutations.
- API key management endpoints still require browser cookie + CSRF; API keys cannot create or revoke other API keys.
- API key hashes are stored in PostgreSQL and cached through Redis by the same HMAC hash so revoke can invalidate the cached key immediately. Redis is an acceleration/cache layer; PostgreSQL remains authoritative.
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
