---
artifact_type: detail_design
id: OFFLINE-FINANCE-SYNC-01
status: draft
owner: shared
updated: 2026-10-06
human_fields: [scope_approval, conflict_policy_approval]
ai_fields: [context_loaded, implementation_decisions, verification_plan]
shared_fields: [status, trace]
trace:
  architecture: ../../architecture/ARCHITECTURE.md
  api: ../../architecture/API.md
  native: 2026-10-05-native-ios-swiftui-design.md
  release: 2026-10-01-release-completion-design.md
  plan: ../plans/2026-10-06-offline-financial-sync.md
---

# Offline financial write/sync contract

## Decision requested

Approve the contract below before implementation. The first delivery is an
iOS-safe offline ledger slice. It must never invent a financial result, silently
overwrite a server change, or replay a mutation as a second ledger event.

## V1 scope

Offline writes are limited to:

- ordinary income/expense transaction creation;
- atomic wallet-to-wallet transfer creation.

Wallet/category/budget/jar/credit edits, transaction edit/delete, balance
adjustments, recurring execution, AI approval, feedback and authentication
mutations remain online-only until their own conflict rules are approved.
Reads use a last-known cache but are marked stale when offline. Pending local
rows are visible as pending and are never presented as server-confirmed totals.

## Source of truth and identity

- PostgreSQL remains the only financial source of truth.
- The client generates one UUID `Idempotency-Key` per user intent and persists
  it before attempting the network call. It is never regenerated on retry.
- The server derives `owner_id` from the authenticated session/API key; no
  client field can select an owner.
- The selected `occurred_at` is the user's account-local date/time. Device
  `created_at` is diagnostic only and cannot replace it.
- Outbox data is keyed by the authenticated account. Logout cannot reassign
  pending mutations to another account.

## HTTP mutation contract

Every financial mutation used by the offline client sends:

```http
Idempotency-Key: 2b8c6c98-...
X-Client-Platform: ios
X-Client-Version: 1.0.0
```

`Idempotency-Key` is required, ASCII, 1–128 bytes, and scoped by
`(owner_id, operation, key)`. The existing transaction and transfer create
routes are updated to reject a missing key with `400 MISSING_IDEMPOTENCY_KEY`.

The first request returns the normal resource envelope and:

```json
{
  "data": {
    "id": "server-resource-id",
    "mutation": {
      "idempotency_key": "2b8c6c98-...",
      "status": "applied",
      "replayed": false
    }
  },
  "meta": null
}
```

A byte-equivalent retry for the same owner/key returns the original resource,
`200` (or the original success status), and `mutation.replayed=true`. It never
calls the domain mutation a second time. Reusing a key with a different
operation or request body returns:

```json
{
  "title": "Conflict",
  "status": 409,
  "code": "IDEMPOTENCY_KEY_REUSED",
  "detail": "Idempotency-Key đã được dùng cho nội dung khác.",
  "request_id": "..."
}
```

The response body and logs must not include the request hash, token, receipt,
or raw financial payload beyond the normal resource returned to the caller.

## Durable server receipt

Add one PostgreSQL `financial_mutation_receipts` table (migration) with:

| Column | Rule |
| --- | --- |
| `id` | UUID primary key |
| `owner_id` | required; server-derived |
| `operation` | `transaction.create` or `transfer.create` in V1 |
| `idempotency_key` | required; unique with owner and operation |
| `request_hash` | SHA-256 of canonical operation + JSON body |
| `status` | `processing`, `applied`, or `rejected` |
| `resource_type`, `resource_id` | canonical result reference |
| `response_json` | redacted canonical envelope for replay |
| `created_at`, `updated_at` | audit timestamps |

Reserve/check, domain write, and receipt completion happen in one database
transaction. A unique-key race waits/reloads the first receipt. A process
crash before commit leaves no ledger row; a timeout after commit is recovered
by replaying the same key.

Receipts are retained for the same retention period as the ledger. They are
not evicted by an in-memory cache and are not owned by Redis.

## Outbox state machine

The native client persists an encrypted, account-scoped outbox before enqueue:

```text
pending -> sending -> applied
                    ├-> retry_wait -> pending
                    ├-> conflict
                    ├-> rejected
                    └-> blocked_auth
```

Each record contains `mutation_id`, `account_id`, `operation`, canonical
payload, `idempotency_key`, `created_at`, `attempt_count`, `next_attempt_at`,
`last_error_code`, and optional server resource ID. Payloads contain no bearer
token or provider secret. The sync engine sends one mutation at a time per
account in creation order; a transfer remains one atomic HTTP call.

Retry policy:

- retry network loss, timeout, `408`, `429`, and `5xx` with capped exponential
  backoff (1s, 2s, 4s, 8s, then 15 minutes) plus jitter;
- refresh the access token once for `401`, then retry the same key;
- do not auto-retry `400`, `403`, `404`, `409`, `422`, or a business rejection;
- `409` becomes an explicit conflict card, never last-write-wins;
- an app restart resumes `pending`, `sending` (leased), and `retry_wait` rows.

After an applied mutation the client invalidates affected reads and refetches
the server totals. It may show the acknowledged row immediately, but it must
not fold an unacknowledged row into authoritative report totals.

## Conflict and recovery UX

V1 create operations are append-only, so a normal create has no stale-version
conflict. Conflicts still cover duplicate-key payload reuse, missing/deleted
wallet/category, insufficient domain validity, and account mismatch. The UI
must show the operation, server reason, and actions:

- `Retry` reuses the same key only for a transient failure;
- `Edit and retry` creates a new mutation only after the user changes the
  draft and confirms;
- `Discard` marks the local outbox row cancelled without deleting server data.

Future edit/delete support adds an explicit integer `base_version` and
server-returned `current_version`; it must be designed before enabling those
operations offline. It will use `409 STALE_VERSION` and offer “Keep server” or
“Edit with current data”, never automatic merge for money amounts.

## Sync lifecycle

1. Form validation runs locally and creates an outbox row atomically with the
   local pending projection.
2. A foreground/network-reachability trigger starts the sync actor.
3. The actor drains rows FIFO, applies the retry policy, and records a result.
4. On success it refetches wallet/transaction/report reads and clears the
   pending projection only after the server response is persisted.
5. On logout it stops the actor. If rows remain, the account is marked
   `blocked_auth` and the UI asks the same account to sign in before resume.

No background financial write is silently executed in V1. Background fetch may
only wake the actor when the same account session is valid and iOS grants
execution time; foreground retry remains the guaranteed path.

## Security and observability

- Keychain protects auth tokens and the local database key; the outbox itself
  is account-scoped and wiped only after explicit user confirmation.
- Logs and metrics contain mutation ID, operation, attempt, status, HTTP code,
  latency and error code; never amount, note, wallet name, token or image.
- Server audit records the owner, operation, idempotency key hash/reference,
  result and request ID. Redis may transport diagnostics but is not the source
  of truth.

## Non-goals

No bank synchronization, payment execution, multi-device merge, automatic
last-write-wins, offline credit statement calculation, or offline AI/OCR
processing is introduced by this contract.
