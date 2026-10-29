---
artifact_type: detail_design
id: FEEDBACK-20261019-FINANCE
status: in_review
owner: shared
approval: owner requested all-feedback fixes; parent released shared test-first barrier
---

# Finance feedback audit and portfolio integrity

Trace: [backlog](../BACKLOG.md), [phase](../phases/PHASE-002-native-ios-swiftui.md),
[original feedback work](NATIVE-MOBILE-FEEDBACK-20261013.md),
[API](../../architecture/API.md), [portfolio screen](../../design/pages/portfolio/README.md),
[validation](../VALIDATION_MATRIX.md), [release notes](../../releases/CHANGELOG.md).

## Evidence and bounded scope

Read context, backlog, quality/validation/debugging/testing/docs standards, active
native phase, original feedback ticket, portfolio handler/repository/entity,
savings/budget report implementations and the October 16 savings/budget plan and
[ADR-023](../../decisions/ADR-023-ledger-savings-recurring-budgets.md).
Feedback snapshot: `/tmp/mypocket-feedback-all-20261019.json`.

- `b017284a-940c-4533-8950-7c7995f956ea`: portfolio quantity is already fixed-point
  and costs are unit price multiplied by quantity. Current API intentionally
  appends trades; edit/delete of trade history is not implemented. Deletion only
  applies to unused assets. This slice cannot close the broad edit/delete request.
- `6b8b32e1-3ea2-4d3b-b126-e236a7c30591`: migration 028 already enforces categories
  for ordinary transactions; adjustments are the documented exception.
- `1db32af3-fca8-48d9-8e10-4cef60bcb032`: ordinary report exclusion is implemented;
  transfers are stored excluded, savings goal balances deliberately include all
  ledger effects, and surplus/budget reports filter excluded rows and transfers.
- Savings sections, paged drilldown, recurring weekly/monthly/quarterly/yearly
  budgets, positive carryover and deterministic forecasts already exist locally.
  Older partial status is superseded by October 16 implementation, not new work.

## Reproduction, root cause and intended fix

1. Send `{}` or top-level `null` to the existing portfolio price PATCH for a
   priced asset. A `*int64` decodes omitted and explicit null identically, so the
   handler currently clears saved price. Require explicit `latest_price` presence;
   retain explicit null to clear, zero as a known zero price, and integer price.
   Preserve the existing response/error shapes and route.
2. Start a trade transaction holding its asset row lock, then delete that asset
   concurrently. Delete reads/counts without the lock used by CreateTrade, sees
   no committed trade, and later cascades the new trade away. Acquire the same
   asset UPDATE lock before checking history. Whichever operation gets the lock
   first defines the serial outcome; a committed trade must never be erased.

Files: `backend/internal/controller/http/portfolio_handler.go`, new matching
`portfolio_handler_test.go`; `backend/internal/infrastructure/repository/portfolio_postgres.go`
and new `portfolio_postgres_test.go`. No migration, new route, financial model,
auth change, paid provider or production access. Rejected alternatives: adding
mutable trade history (changes approved accounting policy); deleting/correcting
existing rows (unnecessary and unsafe); a count-only recheck (still races).

## Verification plan

Use local migrated PostgreSQL on port 55436 and unique synthetic owner/asset IDs.
HTTP integration asserts malformed/omitted price leaves persisted valuation
untouched and explicit null/zero/positive values work. Concurrency integration
holds the trade lock, waits for the delete to block using PostgreSQL activity,
commits the trade, and asserts conflict plus intact asset/history. Narrow RED
first, then minimum code, GREEN, surrounding finance tests, and parent full suite.
No human UAT substitutes for these data-integrity assertions; portfolio interaction
acceptance and missing edit/delete semantics remain separate open work.

## Reconciliation plan

Parent owns context/backlog/validation/changelog and shared master updates.
API documentation should clarify required presence of `latest_price` while
retaining explicit-null clearing. Public portfolio guide and AI-readable API
reference should agree. Requirements/ERD/architecture/ADR need no new decision:
existing price and immutable history semantics are preserved; no schema/boundary
changes. Parent should record this as partial portfolio integrity coverage only.

## Verification evidence

RED confirmed 2026-10-19 with local migration-35 PostgreSQL:

```sh
rtk proxy env 'TEST_DATABASE_URL=postgres://mypocket_test:local-fixture-only@127.0.0.1:55436/mypocket_savings_budget?sslmode=disable' go test ./internal/controller/http ./internal/infrastructure/repository -run 'TestPortfolio(PriceRequiresExplicitFieldAndPreservesNullClearing|DeleteCannotCascadeConcurrentTrade)$' -count=1
```

Fail as expected: omitted field, unrelated field and top-level null each returned
200 and persisted a null price; concurrent deletion returned nil and erased both
asset and committed trade (`assets=0 trades=0`). Explicit null/zero/positive and
invalid scalar cases passed. Tests clean up their UUID-scoped synthetic owners.
Parent released the combined planning/test barrier after this RED evidence.
Implemented explicit price presence validation and the shared asset UPDATE lock.
The same command then passed both packages; eight HTTP cases retain null/zero/
positive semantics and reject omitted/null-body/unrelated/invalid price inputs.

Surrounding finance regressions passed across all four packages:

```sh
rtk proxy env 'TEST_DATABASE_URL=postgres://mypocket_test:local-fixture-only@127.0.0.1:55436/mypocket_savings_budget?sslmode=disable' go test ./internal/entity ./internal/usecase ./internal/controller/http ./internal/infrastructure/repository -run 'Portfolio|Savings|Budget|Transfer|Wallet|IncludedInReports|ReportExclu' -count=1
rtk proxy env 'TEST_DATABASE_URL=postgres://mypocket_test:local-fixture-only@127.0.0.1:55436/mypocket_savings_budget?sslmode=disable' go test -race ./internal/controller/http ./internal/infrastructure/repository -run 'TestPortfolio' -count=1
rtk git diff --check
```

All passed. Tests use real migrated PostgreSQL with unique synthetic owners;
cleanup errors are checked, and concurrency synchronization observes blocking
by the test's exact transaction PID. Parent owns the final combined suite.

Docs review: updated [public guide and its AI-agent reference](../../guides/portfolio-ios.md).
API.md shared reconciliation is handed to parent. ERD, architecture, requirements,
screen design and ADR remain unchanged because this restores existing price/
immutable-history guarantees without a schema, feature or layout change.
Parent context/backlog/validation/changelog reconciliation remains outstanding.
Broader portfolio asset/trade editing/deletion feedback remains open because
the current approved history model is append-only; unused-asset deletion and
unit-price math must not be represented as completing that request.

## Trade correction follow-up — 2026-10-19

The existing accounting semantics support a narrowly bounded compensating event:
the client can prepare an explicit opposite-side trade and submit it through the
existing `POST /portfolio/assets/{id}/trades` route after user confirmation.
The domain test `TestComputePortfolioPositionSupportsExplicitCompensatingTradeWithoutRewritingHistory`
proves that an exact, fee-free buy/sell reversal can neutralize the position
while preserving both immutable rows and their order. This is a proof of the
existing append-only contract, not a claim that every historical correction is
neutral: fee, price, backdated occurrence and later trades can change realized
P&L or cost basis.

No dedicated correction endpoint was added. A generic endpoint would need an
approved schema for correction linkage, chronological recomputation, fee
treatment, realized P&L, and drafts/idempotency; inferring those rules would
silently change investment accounting. Public API and iOS guidance now require
an explicit draft plus confirmation, identify the original trade in the note,
and prohibit row deletion or free-form implicit reversals. The remaining
portfolio correction backlog therefore stays open as a product/accounting
contract item, with owner isolation and append-only history preserved.
