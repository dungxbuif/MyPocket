---
artifact_type: detail_design
id: TICKET-02-02-TRANSFER
status: in_review
owner: shared
trace:
  parent: TICKET-02-02-giao-dich-chuyen-vi.md
  backlog: ../BACKLOG.md
  validation: ../VALIDATION_MATRIX.md
  release_notes: ../../releases/CHANGELOG.md
---

# TICKET-02-02 — Chuyển tiền giữa hai ví

## Context and decision

The existing ledger only accepts one-wallet income/expense rows. That cannot
represent a transfer atomically: a retry can leave only one side, balances can
drift, and reports must not count the movement as income or expense. This slice
implements an authenticated internal-transfer endpoint and the matching Quick
Add flow. Agent chat, public API keys/scopes, credit ledgers, and external bank
connectors remain out of scope.

## Contract

- `POST /api/v1/transactions/transfer` accepts `source_wallet_id`,
  `destination_wallet_id`, positive integer `amount`, RFC3339 `occurred_at`,
  optional `note`.
- Both wallets must belong to the caller, be different, and not be credit
  wallets. The server selects system categories `expense_transfer_out` and
  `income_transfer_in`; callers cannot supply arbitrary categories.
- The server creates two ordinary ledger rows in one database transaction,
  sharing a `transfer_id`, with `included_in_reports=false` and no jar.
- A transfer is editable/deletable as two linked rows only in a follow-up
  ticket; the existing single-row editor must not mutate a linked pair.
- Wallet balances derive from the two rows. Reports and jar totals already
  exclude the system transfer categories and the report flag.

## Implementation scope

- Add nullable `transactions.transfer_id` and an index in migration 000016.
- Add an internal transfer repository contract and a PostgreSQL implementation
  that validates ownership again inside the transaction and inserts both rows.
- Add the HTTP handler, route, Swagger annotations, public API docs, and a
  small frontend transfer mode composed from existing base fields/selects.
- Keep the id and timestamps generated server-side; no client-generated pair
  IDs are trusted.

## Alternatives rejected

- Reusing two independent `POST /transactions` calls: not atomic and unsafe on
  retries.
- Adding `type=transfer` to the existing table: current balance, reports,
  jar rules, and downstream consumers already model signed income/expense rows;
  the paired rows preserve those invariants with a stable link.
- Persisting a separate transfer aggregate only: would require changing every
  balance/report query and would duplicate ledger data.

## Risks and safeguards

- Cross-account or credit-wallet movement is rejected at the HTTP boundary and
  rechecked in the repository transaction.
- Same-wallet transfers are rejected to avoid a no-op pair.
- A unique transfer id links exactly two rows; a database transaction prevents
  half-created pairs. A later repair audit can count pair cardinality.
- Existing single-row edit/delete remains unchanged for unlinked rows;
  linked-pair mutation is explicitly deferred rather than silently partial.

## Verification plan

- Unit/handler tests for amount, same-wallet, owner-scope, credit-wallet and
  missing system-category rejection.
- PostgreSQL integration test proves two rows, shared `transfer_id`, excluded
  reporting, atomic rollback and derived balances.
- Frontend design/build checks plus a browser smoke test for opening transfer
  mode and submitting the two-wallet form.
- The browser proof is now checked in as `app/scripts/transfer-browser.test.mjs`
  and the API/state proof as `app/scripts/transfer-roundtrip.mjs`; both are
  exposed through `npm run test:transfer:e2e` and
  `npm run test:transfer:integrated`.
- Run migration version/clean checks and update the validation matrix, API,
  ERD/architecture notes, context, backlog, and changelog.

Approval: owner instruction “Implement còn lại” authorizes this bounded core
ledger slice; worker/recurring and other high-risk slices still require their
own designs.
