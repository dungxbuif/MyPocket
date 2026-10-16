# Savings report API

Implementation and verification note, 2026-10-16:
[approved contract](../superpowers/specs/2026-10-16-savings-budget-design.md),
[Task 1](../superpowers/plans/2026-10-16-savings-budget-implementation.md).

The savings report reads owner-scoped wallet and transaction repositories,
derive goal balances from every ledger effect, and derive account surplus from
report-included ordinary income/expense. No schema or financial write is needed.
The public Markdown below also serves as the machine-readable reference surface.

Verification plan: pure ledger examples (20M income − 12M expense = 8M surplus;
5M goal transfer separate), internal goal transfers, account-local month bounds,
missing/future dates, signed adjustments, zero income, checked integer arithmetic,
identical paged drilldown predicates, owner and handler validation failures.

Parent integration owns API registration, generated Swagger and shared project
reconciliation. Deployment and owner-authenticated native UAT remain separate.

## Read a monthly report

Call `GET /api/v1/reports/savings?month=2026-10` with the existing bearer JWT
or owner API key. Omit `month` to select the account-local current month.
Add `wallet_id=<goal UUID>` to show one owned goal. The wallet filter affects
the goal section only: `surplus.scope` always remains `account`.

The standard response is `{ "data": <report>, "meta": null }`. The report
has this structured surface (money values are integer VND):

| Field | Meaning |
| --- | --- |
| `month` | Account calendar month, `YYYY-MM` |
| `timezone` | Current account IANA timezone |
| `as_of` | Server evaluation instant, RFC3339 UTC |
| `goals.opening_balance` | Sum of selected goal opening balances and all earlier ledger effects |
| `goals.closing_balance` | Sum of selected goal balances at month end or server now |
| `goals.deposits` | Ordinary income credited to selected goals |
| `goals.withdrawals` | Ordinary expense debited from selected goals, positive amount |
| `goals.adjustments` | Signed balance adjustments |
| `goals.net_change` | Closing minus opening balance |
| `goals.internal_transfers` | Signed paired-transfer effects on selected goals |
| `goals.initial_balance_added` | Configured starting balance of goals created during the month; separate from transactions |
| `goals.items` | Goal items, ordered by name then wallet ID |
| `surplus.scope` | Literal `account` |
| `surplus.income`, `surplus.expense` | Report-included ordinary account income and expense |
| `surplus.net` | Income minus expense; may be negative |
| `surplus.savings_rate_bps` | `net * 10000 / income`, integer toward zero; null when income is zero |
| `surplus.transaction_count` | Number of rows used in account surplus |

Every goal item contains those money metrics above plus `wallet_id`,
`name`, `target_amount`, `target_date`, `remaining`, `progress_bps`, `reached`
and `transaction_count`. Target fields describe the current goal configuration,
including when viewing historical months. `target_date` is `YYYY-MM-DD` or null;
target amount is nullable for legacy data. For a positive target, remaining is
`max(target - closing, 0)`, progress is `max(closing, 0) * 10000 / target` and
may exceed 10000, and reached means closing balance meets the target. Missing
positive targets yield remaining/progress zero and reached false.

For example, 20,000,000 income and 12,000,000 expense produce surplus
8,000,000 and savings rate 4000 bps. Moving 5,000,000 from a basic wallet into
a goal produces goal `internal_transfers=5000000` and raises its closing
balance; it does not add income to surplus. Moving money between two selected
goals nets to zero in the aggregate while each goal shows its signed movement.
Goal activity and account surplus are separate measures and must not be added.

A goal created in October with an initial balance does not place that balance
into September's history or October's opening. The baseline enters October's
`initial_balance_added` and closing balance at wallet creation. It is source
configuration, not an invented transaction: transaction drilldown therefore
does not include a fake deposit/adjustment for this amount. Reconcile as:
`closing = opening + initial_balance_added + deposits - withdrawals + adjustments + internal_transfers`.
Historical ledger rows entered later remain visible on their occurrence dates.
Legacy metadata without a wallet creation date keeps its prior baseline.

## Drill into matching transactions

Call `GET /api/v1/reports/savings/transactions?month=2026-10&section=goals`.
The response envelope contains `items`, `total`, `limit`, `offset`, `month`,
`timezone` and `as_of`. Items use the existing transaction JSON schema. `total`
counts all matching rows before pagination; a page beyond the end returns
`items: []` and the unchanged total. Ordering is `occurred_at DESC, id DESC`.

| Query parameter | Supported values |
| --- | --- |
| `month`, `wallet_id` | Same as report; wallet filter requires section `goals` |
| `section` | Required: `goals` or `surplus` |
| `kind` for goals | `all` (default), `deposit`, `withdrawal`, `adjustment`, `transfer` |
| `kind` for surplus | `all` (default), `income`, `expense` |
| `limit` | Integer 1–100, default 50 |
| `offset` | Nonnegative integer, default 0 |

Totals and drilldown use the same owner, wallet, month, cutoff and row-kind
predicates. Goal activity includes ordinary rows even when excluded from
ordinary reports, so its balances reconcile the ledger. Transfers are
recognized by `transfer_id`, and adjustments by their explicit direction.
Surplus requires `included_in_reports=true` and excludes paired transfers,
adjustments, credit payment kinds and credit-payment-linked rows. Interest is
ordinary income if recorded that way; note text does not change classification.
Transactions whose wallet is outside the owner's existing wallet set and rows
with missing occurrence dates are excluded from both report and drilldown.

All month boundaries follow the account timezone, including daylight-saving
changes. The next month's start is exclusive. Rows after the server instant
are excluded from current results. Historical reports recalculate from the
current ledger after edits; they are not frozen snapshots. The implementation
reads the complete existing owner ledger; it does not cap or truncate totals.

## Errors and availability

Invalid month, incompatible section/kind, duplicate recognized query parameters,
malformed wallet ID or invalid pagination returns 400 `SAVINGS_INVALID_QUERY`.
Future months return 400 `SAVINGS_FUTURE_MONTH`. Foreign, deleted and non-goal
wallet filters return 404 `SAVINGS_WALLET_NOT_FOUND`. Missing/invalid credentials
use the existing 401 authentication response. Repository/timezone failures
return 500 `SAVINGS_FAILED`, and checked arithmetic overflow returns
500 `SAVINGS_OVERFLOW`; neither produces a fabricated empty report.

## Verification evidence

`rtk go test ./internal/entity ./internal/usecase ./internal/controller/http -run TestSavings -count=1`
passed 10 tests on 2026-10-16. Coverage includes separate surplus/transfers,
internal goal transfer net zero, month defaults, leap-year and daylight-saving
boundaries, missing/future rows, signed adjustments, exclusion flags, current
targets, zero income, negative balances, overflow, pagination, fail-closed
dependencies, owner isolation and actual JWT/API-key authentication middleware.
Parent integration owns the full backend suite and registered-route proof.
