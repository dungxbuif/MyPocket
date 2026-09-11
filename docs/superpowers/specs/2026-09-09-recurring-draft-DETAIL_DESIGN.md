---
artifact_type: detail_design
id: SO-TIEN-F2-DRAFT
status: ready
owner: shared
human_approval: Owner authorized continuous implementation, testing and deployment without further approval on 2026-09-10.
implementation_status: approved_for_execution
---

# Recurring draft decisions — bounded F2 contract

Parent: [approved Sổ tiền design](2026-09-09-so-tien-design.md). Trace: [backlog](../../work/BACKLOG.md), [recurring ticket](../../work/tickets/TICKET-013-recurring-schedules-worker.md), [API](../../architecture/API.md), [validation](../../work/VALIDATION_MATRIX.md). This is a proposed contract, not implemented or released behavior. An approved overall redesign does not silently approve these public routes and transaction boundary changes.

## Decision requested

Add owner-scoped confirm/reject for existing recurring drafts. Confirmation allows editing amount and note only; wallet/category/type/date come from the locked occurrence. Confirmation posts accounting exactly once; rejection never affects balances. Keep AI/OCR/import drafts deferred. No migration, dependency, auth-policy or production change is proposed.

## Source evidence and boundary

`backend/migrations/0006_phase004_recurring_schedules.sql` already has pending/confirmed/rejected state, confirmed transaction ID, version, timestamps and unique occurrence identity. Planning list currently omits the confirmed ID. Finance `Repository.CreateTransaction` opens and commits its own SQL transaction: calling it and then marking a draft confirmed would be unsafe partial success.

Extract a transaction-bound finance creation helper without changing ordinary creation behavior. Planning confirmation owns a single SQL transaction, locks the owned draft `FOR UPDATE`, validates status/version, invokes accounting inside that transaction, updates draft status/link/amount/note/version, and commits once. Rollback must cover balances, transaction, idempotency and draft together. Lock wallet IDs in deterministic order using the existing multi-wallet locking helper; this also avoids opposite-direction transfer lock inversion inside this new path.

## Proposed public API

`POST /api/v1/transaction-drafts/{id}/confirm` requires `Idempotency-Key` and JSON `{ "version": 1, "amount_vnd": 3500000, "note": "Thuê nhà" }`. Positive integer VND only, consistent with existing recurring transactions. No user ID, wallet, category or type overrides accepted. Response 200: `{ "status": "ok", "draft": { ...full draft including confirmed_transaction_id... }, "transaction": { ...linked transaction... }, "correlation_id": "..." }`.

`POST /api/v1/transaction-drafts/{id}/reject` accepts `{ "version": 1 }`. Response 200 uses the same envelope without a transaction. Add `confirmed_transaction_id` to the existing list representation so terminal decisions are observable.

Cookie authentication retains CSRF checks; API-key bearer authentication retains owner scoping and existing CSRF exemption. No expansion of key-management permissions.

| Case | Result |
| --- | --- |
| First decision | 200, one terminal state transition |
| Same-action replay | 200, same terminal IDs, no balance/version change; evaluated before stale-version rejection |
| Confirm replay with different amount/note | 409 `DRAFT_ALREADY_RESOLVED` |
| Confirm rejected / reject confirmed | 409 `DRAFT_ALREADY_RESOLVED` |
| Pending stale version | 409 `DRAFT_VERSION_CONFLICT` |
| Malformed input, non-positive amount, missing confirm key, invalid finance references | 400 `VALIDATION_FAILED`, no effects |
| Missing/foreign draft | 403 `FORBIDDEN`, matching existing ownership-hiding behavior |
| Unauthenticated | 401 `AUTH_REQUIRED` |
| Database failure | 503 `INTERNAL_RETRYABLE`, no partial effects |

## Replay details that must not be left implicit

- Draft row locking is the primary exactly-once guard, including concurrent requests with different keys.
- Reuse of a key for unrelated direct transactions or another draft must not return/link the unrelated transaction. Derive a bounded finance replay key from a fixed operation namespace, draft ID and supplied key (SHA-256), not the raw shared finance key. Test identical payloads on two different drafts with the same supplied key.
- Confirmed draft amount/note are the immutable accepted decision. Compare replay against those fields, not against a later edited transaction. Linked transaction readback may reflect subsequent edits/archive; document that replay promises stable IDs/no new effects, not an eternally identical transaction representation.
- Offline confirmation/rejection is not queued in this slice. Retain inputs on failure and offer explicit retry using the same key. Do not show accepted state before server confirmation.

## Tests and reconciliation required after approval

1. PostgreSQL known-answer expense/income/transfer, concurrent confirmations, same/different keys, replay and terminal conflicts; two drafts sharing a client key remain independent.
2. Inject failure between finance insertion and draft update: assert no transaction, wallet, idempotency or terminal-state changes persist.
3. Foreign/archived wallet, inactive/wrong-kind category, stale version and changed references leave pending drafts and balances unchanged.
4. Cookie+CSRF and bearer ownership parity, missing CSRF/key, malformed routes/methods, safe correlation envelopes.
5. Mounted real-draft UI confirm/reject, retry/input retention/double-submit and exact balances after reload. No fake draft rows or direct DB mutations as the production UI workflow.
6. Reconcile API/Docusaurus action mapping, architecture boundary and ADR; add execution evidence to validation/context/backlog/changelog. Do not mark released before a separately authorized release.

Whole Sổ tiền F2 still requires separate decisions for category hierarchy/settings readback, zero/negative adjustment storage, schedule editing and report gaps. This bounded proposal does not approve those implicitly.
