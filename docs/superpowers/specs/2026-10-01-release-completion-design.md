# MyPocket Release Completion Design

**Status:** approved for staged implementation (2026-10-01)

## Outcome

Bring the web application from the current basic ledger slice to a releasable
stage without inventing fake UI behavior. Each missing domain is delivered as
a vertical slice with an owner-scoped API, persistence rules, shared web
components, automated tests, reconciled documentation, and explicit browser or
provider evidence where the environment permits it.

The existing uncommitted worktree is preserved. Native/mobile work remains
outside this web completion stream.

## Scope and order

1. **Ledger completion:** linked transfer pair edit/delete with replay safety,
   balance adjustment as a normal ledger event, and safe bulk deletion.
2. **Credit wallet:** purchases, refunds, fees/interest, statement periods,
   amount due, payment allocation, remaining balance, due/partial/paid/overdue
   status, and credit-specific UI.
3. **Recurring and worker:** recurring transaction/budget definitions,
   account-timezone due evaluation, idempotent catch-up, retries, and a worker
   process that shares domain code with the API.
4. **Reports and secondary domains:** Money Insider narrative boundaries,
   portfolio valuation and weighted-average cost, and Travel Mode as event
   linkage rather than a wallet type.
5. **AI/PWA release gates:** live provider/browser proof, attachment recovery,
   durable audit delivery, offline queue/conflict semantics, notifications, and
   production verification.

The slices are independent where possible. A later slice cannot silently widen
an earlier API or reinterpret existing income/expense rows.

## Invariants

- Every read and write is owner-scoped in the use case and rechecked in the
  repository transaction.
- Money remains integer VND; no multi-currency or exchange-rate behavior is
  introduced by this design.
- A transfer is two linked rows and is excluded from reports and jars. A linked
  pair can only be edited or deleted atomically.
- Balance adjustment is an explicit immutable ledger event, not a mutation of
  `opening_balance` and not an unlabelled expense/income.
- Credit rows never use the basic wallet balance formula. Payment allocation,
  statement status, and due status are derived from credit-domain rows.
- Recurring execution is idempotent by schedule occurrence and account
  timezone; retries cannot create a second occurrence.
- AI suggestions remain drafts until explicit user approval. Provider output,
  OCR text, and attachments are untrusted input; no automatic ledger write is
  introduced.
- A feature is not called complete from unit tests alone: browser/provider/UAT
  status is recorded separately in the validation matrix.

## Architecture

The Go API remains the source of truth. Domain use cases expose narrow
interfaces to repositories, and PostgreSQL migrations own durable state. The
React web app composes existing atomic components (`BaseButton`, `BaseSelect`,
`BaseBottomSheet`, `SurfaceCard`, `StatusMessage`, and the transaction row)
instead of adding local controls. Swagger is generated from handlers and public
API docs mirror the implemented contract.

The worker is a separate executable only after recurring/month-close semantics
are approved; it reuses existing config, repositories, and use cases rather
than duplicating business logic. Redis remains an audit/coordination mechanism,
never the source of financial truth.

## Slice 1 contract: ledger completion

### Transfer pair mutation

- `PATCH /api/v1/transactions/transfer/{transfer_id}` updates amount,
  occurred-at, note, and (where allowed) the source/destination wallets as one
  transaction. The pair must contain exactly two owner-scoped rows with the
  expected system categories.
- `DELETE /api/v1/transactions/transfer/{transfer_id}` deletes both rows in one
  transaction and is idempotent for an already-deleted pair.
- Both routes accept an `Idempotency-Key`; the same owner/key/body returns the
  original result, while a changed body returns `409`.
- The existing single-row editor rejects linked rows and points callers to the
  pair flow.

### Balance adjustment

- `POST /api/v1/transactions/adjustment` records a signed adjustment event for
  one basic or goal wallet with amount, occurred-at, note, and an explicit
  direction. It is excluded from ordinary income/expense category totals but is
  visible in the wallet ledger and balance reconciliation.
- Adjustments are immutable after creation; correction means a compensating
  adjustment. Credit wallets are not accepted by this slice.

### Bulk delete

- `POST /api/v1/transactions/bulk-delete` accepts a bounded list of transaction
  IDs owned by the caller. Linked transfers are rejected unless the request uses
  the transfer endpoint; the operation is all-or-nothing and returns deleted
  IDs plus skipped/rejected reasons.
- The web menu enables the action only when selection and confirmation are
  present. No silent partial deletion is allowed.

## Error and recovery behavior

Validation errors use the existing Problem Details envelope with a stable code
and request ID. Cross-owner IDs, malformed pair cardinality, stale idempotency
body, credit/unsupported wallet kinds, and partial bulk requests fail without
writes. Timeouts after a committed mutation are recovered with the same
idempotency key or a read endpoint; the UI never blindly resubmits a financial
mutation.

## Verification gates

- RED/GREEN unit and handler tests for each new contract.
- PostgreSQL tests for owner scope, atomicity, pair cardinality, idempotency,
  derived balances, report exclusion, and rollback.
- Frontend design/type/build tests plus browser state verification for transfer
  mutation, adjustment, and bulk-delete confirmation.
- `go test ./...`, migration clean/version checks, frontend design/type/build,
  and focused API/browser scripts before each slice is marked implemented.
- Validation matrix distinguishes automated proof, local browser proof, live
  provider proof, owner UAT, and deployment proof.

## Explicit non-goals for this approval

This design does not add bank connectors, payment execution, multi-currency,
automatic AI approval, a second mobile codebase, or an immutable month-close
snapshot. Those require separate product decisions and remain documented as
follow-ups.
