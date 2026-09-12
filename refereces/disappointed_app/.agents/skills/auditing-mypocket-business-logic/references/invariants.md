# MyPocket Business Invariants

## Ledger equations

- `income`: source balance increases by `amount_vnd`.
- `expense`: source balance decreases by `amount_vnd`.
- `transfer`: source decreases and distinct destination increases by the same amount in one database transaction.
- `adjustment`: source balance becomes `target_balance_vnd`; persisted delta equals target minus prior balance.
- Editing reverses the old persisted deltas before applying the new validated effect.
- Archiving reverses the persisted deltas exactly once and excludes the row from active reads.
- All additions, subtractions, totals, percentages and decimal-to-int conversions reject overflow instead of wrapping.

## Reporting equations

- Cash-flow income/expense reports omit transfers, adjustments, archived rows and `excluded_from_reports=true` rows unless an endpoint explicitly documents another contract.
- Date boundaries use the report timezone; the product default is `Asia/Ho_Chi_Minh`.
- A wallet filter changes only the selected scope and must not leak another user's wallet.
- Dashboard totals and reports must state whether wallets with `include_in_total=false` are included. Do not infer equivalence between these scopes.

## Concurrency and replay

- Create requests with the same user + idempotency key + canonical payload return the original result.
- Reusing the key with any business-relevant field changed is rejected.
- Updates and archive commands for versioned wallet, category, transaction and planning records use the caller's `base_version`; stale versions return HTTP 409 `VERSION_CONFLICT` without mutation.
- Domain mutation, change-feed append and mutation receipt must be atomic across crashes and concurrent retries. Inject failures at both feed and receipt insertion and verify complete ledger rollback.
- Direct finance/portfolio commands, draft confirmation and price-worker writes must also emit canonical changes in their domain transaction.
- Retry ambiguous responses with identical mutation ID and payload; never fall back to a fresh direct create. Batches commit per mutation.
- Acquire the shared per-user write lock before accounting entity locks; prove concurrent replay produces one ledger effect.
- Resync data and cursor must share one consistent database snapshot. Migration 0012 enables asset feed CHECK values; historical gaps require a full resync.

## Ownership and external access

- Every object lookup is scoped by authenticated `user_id`; foreign UUIDs behave as not found or forbidden without disclosure.
- Bearer API keys act only as their owning user and support documented business endpoints without CSRF.
- API-key create/list/revoke requires the browser session and CSRF unless the public contract explicitly changes.
- Revocation must invalidate Redis-cached identity; database state remains authoritative on cache miss or Redis failure.

## Planning

- Creating an obligation does not alter wallet balances.
- A repayment links an existing owned transaction, cannot exceed remaining principal, and must match the documented cash direction.
- Recurring income/expense/transfer payloads obey the same wallet/category rules as direct transactions.
- Recurring catch-up work is bounded per worker run; a schedule delayed for years cannot create an unbounded database transaction.
- A pending draft has no accounting effect. Confirm applies one transaction once; reject never changes balances.
