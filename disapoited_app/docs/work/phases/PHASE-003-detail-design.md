---
artifact_type: detail_design
id: PHASE-003-DETAIL-DESIGN
status: approved
owner: shared
approval: approved
approved_on: 2026-08-30
trace:
  backlog_item: BL-003
  phase: PHASE-003
  requirements: [REQ-F-004, REQ-NF-002, REQ-NF-003, REQ-NF-008]
  tickets: [TICKET-008, TICKET-009, TICKET-010]
  validation_matrix: ../VALIDATION_MATRIX.md
  master_docs_touched: [../../architecture/API.md, ../../architecture/ERD.md, ../../architecture/ARCHITECTURE.md]
---

# DETAIL DESIGN: PHASE-003 Offline Synchronization

## 1. Context and Scope

MyPocket must remain useful without a network connection while PostgreSQL stays authoritative. This phase replaces the initial localStorage transaction outbox with a versioned IndexedDB mirror, a durable mutation queue, incremental server changes, and explicit conflict review.

- In scope: wallets, categories, transactions, planning-ready sync envelopes, tombstones, cursor recovery, conflict inbox, reconnect replay, and full resync.
- Out of scope: collaborative live editing, silent last-write-wins, provider ingestion, and background sync guarantees when the browser is fully terminated.
- Approval: the user delegated remaining design decisions on 2026-08-30.
- Small task exemption: no; API, data, conflict, and offline behavior change.

## 2. Decisions and Trade-offs

| Decision | Chosen approach | Rejected alternative |
| --- | --- | --- |
| Browser storage | Native IndexedDB adapter with schema migrations and localStorage only as emergency migration source | Keep localStorage as primary; insufficient capacity and transactionality |
| Mutation identity | Client UUID plus mutation UUID, entity base version, ordered per-device sequence | Timestamp ordering; nondeterministic under clock skew |
| Server authority | PostgreSQL change log with per-user monotonic cursor | Client-to-client merge or server timestamps |
| Conflicts | Preserve server record and create conflict item for review | Last-write-wins or automatic field merge |
| Replay | One ordered batch at a time; stop affected entity after conflict, continue independent entities | Blind parallel replay |
| Deletion | Tombstones retained through cursor horizon | Hard delete from change feed |

## 3. Architecture

```text
+------------------+       batch mutations       +------------------+
| React PWA        | ---------------------------> | Sync HTTP API    |
| query + commands | <--------------------------- | auth/idempotency |
+--------+---------+       results/conflicts      +--------+---------+
         |                                                  |
         v                                                  v
+------------------+                               +------------------+
| IndexedDB        |       incremental cursor      | PostgreSQL       |
| mirror/outbox/   | <---------------------------- | domain + changes |
| conflicts/meta   |                               +------------------+
+------------------+
```

- `frontend/src/offline`: IndexedDB schema, repository, optimistic reducers, outbox processor, cursor hydration, and conflict state.
- `backend/internal/sync`: mutation envelope validation, domain command dispatch, idempotency, change feed, and resync tokens.
- `backend/internal/platform/httpapi`: authenticated `/api/v1/sync/*` routes.

## 4. Data and API

IndexedDB stores: `wallets`, `categories`, `transactions`, `outbox`, `conflicts`, `tombstones`, and `meta`. Each mirrored record includes `id`, `version`, `updated_at`, `sync_state`, and server payload. Outbox rows include `mutation_id`, `device_id`, `sequence`, `entity_type`, `entity_id`, `operation`, `base_version`, payload, attempts, and state.

PostgreSQL adds `sync_changes(user_id, cursor bigint, entity_type, entity_id, operation, version, payload_json, created_at)`, `sync_mutations(user_id, mutation_id, request_hash, result_json, created_at)`, and indexed retention metadata.

- `POST /api/v1/sync/mutations`: accepts at most 100 ordered mutations; returns applied, replayed, rejected, or conflict results per item.
- `GET /api/v1/sync/changes?after={cursor}&limit={n}`: returns user-scoped changes and `next_cursor`.
- `POST /api/v1/sync/resync`: issues a bounded snapshot token and authoritative collections after invalid cursor/schema recovery.
- All mutations require CSRF, authenticated user context, UUID validation, payload size bounds, and idempotent mutation IDs.

## 5. Runtime Flow

1. Hydrate IndexedDB before rendering finance data; show cached data with a stale/offline marker.
2. Offline commands validate locally, update the mirror optimistically, and append one outbox row atomically.
3. On reconnect, process ordered batches, apply server results, then pull changes after the last cursor.
4. A base-version mismatch creates a conflict with local intent and server state; no server overwrite occurs.
5. Keep-server removes the local mutation; edit-and-retry creates a new mutation against the server version; discard-local restores the server record.
6. Invalid cursor or schema migration failure triggers explicit full resync after preserving unsent mutations.

## 6. Security and Failure Policy

- User ID never appears as trusted mutation ownership input; it comes from auth context.
- IndexedDB contains private finance data and is cleared on logout/account deletion.
- Logs contain mutation IDs and correlation IDs, never full notes or receipt payloads.
- Exponential retry is capped; validation/authorization failures are terminal; network/5xx failures remain retryable.
- Quota or IndexedDB failures enter read-only degraded mode with a visible recovery action.

## 7. Verification and Reconciliation

- Unit: optimistic reducers, ordering, retry classification, schema migration, cursor merge, and conflict actions.
- Integration: mutation replay, user isolation, cursor monotonicity, tombstones, base-version conflict, and resync.
- E2E/UAT: create/edit offline, reload offline, reconnect once, two-context conflict, cleared storage, and quota/error states.
- Commands: frontend Vitest/build/Playwright and backend `go test -p 1 ./...` with PostgreSQL.
- Update API, ERD, architecture, validation rows, context, backlog, changelog, and create an ADR only if this design changes.

