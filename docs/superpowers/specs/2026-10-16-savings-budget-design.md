---
artifact_type: detail_design
id: SAVINGS-BUDGET-20261016
status: ready
approval: user delegated remaining decisions and requested implementation
---

# Savings report, recurring budgets and forecast

## Intent and decisions

User selected both goal-wallet reporting and income/expense surplus. User
delegated remaining business decisions and explicitly requested implementation.
Preserve existing financial records and native colors/base components. Ship
three independently verifiable slices in the native app and Go API.

Trace: [backlog](../../work/BACKLOG.md), [reports](../../requirements/REPORTS.md),
[budget setup](../../work/tickets/TICKET-03-01-thiet-lap-budget.md),
[budget progress](../../work/tickets/TICKET-03-02-tien-do-budget.md),
[savings](../../work/tickets/TICKET-06-01-muc-tieu-tiet-kiem.md),
[validation](../../work/VALIDATION_MATRIX.md).

## Savings report contract

`GET /api/v1/reports/savings?month=YYYY-MM&wallet_id=<optional goal UUID>`
returns the standard data envelope, scoped by authenticated owner (JWT or API
key). Month defaults to account-local current month; invalid month is 400,
foreign/non-goal wallet filter is 404. A wallet filter affects goal activity
only; the surplus is always account-wide and explicitly labeled as such.

The result contains `month`, `timezone`, `as_of` (RFC3339), `goals` and
`surplus`. `goals` contains `opening_balance`, `closing_balance`, `deposits`,
`withdrawals`, `adjustments`, `net_change`, `internal_transfers`, and `items`.
Each item has the same metrics plus `wallet_id`, `name`, `target_amount`,
`target_date`, `remaining`, `progress_bps`, `reached` and `transaction_count`.
`internal_transfers` is a signed amount, shown separately. Sum of item balances
is the aggregate; moving money between selected goal wallets nets to zero.

The wallet's configured baseline takes effect at `wallet.created_at`. It enters
opening balance only when created before month start; when created during the
month it is exposed separately as `initial_balance_added` in goals and each
item, never fabricated as an income/adjustment transaction. Zero creation dates
in legacy metadata retain baseline-before-history behavior. Backdated ledger
rows remain visible independently of wallet metadata. Opening balance also
includes every signed ledger effect strictly before month start. Closing includes rows before next month
start, capped at server now for the current month. Future month balances are
not fabricated: return 400 `SAVINGS_FUTURE_MONTH`. Historical target/progress
uses current goal configuration, labeled current target; targets have no
historical version table. Future-dated rows after now do not affect current
report. Balance reconciliation uses all rows irrespective of report inclusion.

Ordinary income increases deposits, ordinary expense increases withdrawals;
adjustments are signed separately. Paired transfers are recognized by
`transfer_id`, excluded from ordinary report income/expense, and shown as
signed internal transfers. A transfer into a goal increases its balance even
though `included_in_reports=false`. Interest remains ordinary income unless
the ledger explicitly classifies it differently; do not infer from note text.

`surplus` contains `scope:"account"`, `income`, `expense`, `net`,
`savings_rate_bps` (nullable when income <= 0), and `transaction_count`.
Only report-included ordinary income/expense rows count; exclude transfers,
adjustments and credit principal/payment effects using the existing report
semantics. Surplus and goal balances are never added together.

`GET /api/v1/reports/savings/transactions` accepts the same month/wallet scope,
`section=goals|surplus`, `kind=all|deposit|withdrawal|adjustment|transfer|income|expense`,
`limit` (1..100, default 50), `offset` (>=0). Return `items`, `total`,
`limit`, `offset`, `month`, `timezone`, `as_of`; order occurred_at DESC, id DESC.
Filters must use the identical report predicates; incompatible section/kind
or surplus + wallet filter returns 400. Do not silently truncate totals.

## Recurring budget contract

Existing `/budgets` remains explicit-interval CRUD. Add owner-scoped:

- `GET /api/v1/budget-series`
- `POST /api/v1/budget-series` (requires Idempotency-Key)
- `PATCH /api/v1/budget-series/:id` (requires Idempotency-Key and version)
- `POST /api/v1/budget-series/:id/stop` (requires Idempotency-Key and version)

Creation input: name, limit_amount, wallet_id?, category_id?, cadence
(`weekly|monthly|quarterly|yearly`), start_date, carryover_enabled (default false).
Use integer VND with existing 9,007,199,254,740,991 limit. Weekly starts Monday,
monthly first day, quarterly Jan/Apr/Jul/Oct 1, yearly Jan 1. Custom intervals
cannot recur. Initial start must be canonical and current/future, not an
already-ended period. No partial/prorated initial period.

Series stores owner, cadence, start date, active, version; versioned config rows
store effective_start_date, name, base limit, wallet/category scopes and
carryover flag. PATCH replaces configuration effective at the next local
period boundary; cadence/start_date cannot change. Repeated edit replaces the
pending next-period revision, never historical config. Stop prevents future
period creation and preserves existing periods. There is no resume/delete-series
in this slice. Deleting wallet/category stops dependent series before FK cascade;
archived categories remain usable by existing series but not new selections.

Generate one ordinary budget instance per series/period with a DB uniqueness
constraint `(series_id,start_date)`, preserving base/config snapshot fields.
Catch up missed periods under the existing owner lock, both in worker and
on budget list access. Each pass can process a bounded batch; repeated passes
continue from last generated period. On overlap with explicit same-scope
budget, stop generation for that series and expose `generation_error` without
silently skipping a period. No financial transactions are generated.

Generated instances have `series_id`, `base_limit_amount`, `carryover_amount`,
`effective_limit_amount`; existing `limit_amount` retains base meaning. PATCH
or DELETE on a generated instance returns 409 `BUDGET_SERIES_MANAGED`, avoiding
regeneration/deletion ambiguity. Stop/edit the series through its routes.

For carry-enabled consecutive periods:
`carry[n] = max(0, base[n-1] + carry[n-1] - spent[n-1])`.
Disabled period has carry=0; first period has carry=0. Overspend is not debt.
Changing wallet/category scope resets carryover at that revision boundary.
Historical ledger edits recalculate carry forward on reads, so no frozen stale
balances are persisted. Effective limits are used by progress/thresholds/forecast;
summary limits sum effective limits; existing deduplicated spent stays unchanged.
Checked arithmetic must reject overflow rather than wrap. Timezone uses current
account timezone for scheduling and inclusion; stored date labels do not shift.

## Budget forecast/detail contract

`GET /api/v1/budgets/:id/report` returns `budget_id`, `timezone`, `as_of`,
`start_date`, `end_date`, `base_limit_amount`, `carryover_amount`,
`effective_limit_amount`, `spent`, `remaining`, `forecast`, `daily`.
`forecast` has `status=estimated|actual|not_started`, `method=linear_daily_v1`,
`elapsed_days`, `total_days`, `projected_spent` (nullable before start),
`projected_remaining` (nullable), `over_budget` (nullable before start).
Days use account-local calendar dates inclusive of today, including zero-spend
days. Active estimate: ceil(spent_so_far * total_days / elapsed_days), with
checked integer arithmetic. No future-dated rows enter an active estimate.
Ended: projected_spent=actual spent, status=actual. Future: no projection.
Daily rows `{date,spent}` include zero-spend dates up to today/end.

`GET /api/v1/budgets/:id/transactions?limit=50&offset=0` provides identical
expense/category-descendant/wallet/report-inclusion/time predicates and paged
items + total. Neither the budget report nor savings forecast calls an LLM.
No savings completion-date forecast or scheduled push is included in this slice.

## Native UX and compatibility

Savings is reachable from Reports via a labeled action, with month selector,
goal-wallet filter, separate goal/surplus sections and shared transaction rows.
Budget editor adds cadence, recurring and carryover controls through MP bases.
Generated cards show base + carryover and series edit/stop actions. Detail shows
daily breakdown, estimate label and matching transactions. Reuse MPText,
MPSurfaceCard, MPMetric, MPProgressBar, MPOptionSelector, MPToggle, MPButton,
MPErrorView, MPTransactionRow and current shared shell insets; if an exact base
name differs, reuse the existing corresponding base rather than screen styling.
New fields decode optionally against older APIs. New route failures show a
feature error, not fabricated zero data; cached reports remain owner/month scoped.

## Verification, rollout and documentation

Pure Go tests cover internal goal transfers, exclusion flags, zero income,
leap years, timezone edges, future dates, ended/future forecast, carry chains,
edits and overflow. Real PostgreSQL/Redis tests cover concurrent generation,
idempotent replay, same-key conflict, version conflicts, overlap, scoped deletes,
owner isolation through JWT/API key. Signed native tests cover additive decoding,
editor mapping and report states; API-backed UI verifies real rows/totals.

Use additive migration 035 after a validated backup on deployment. Existing
budget/financial rows remain unchanged. Public docs: API guide, Swagger and
agent-readable Markdown; internal: requirements, architecture, ERD, ADR,
backlog, validation and changelog. All new surfaces remain labeled planned until
implementation and command evidence reconcile them.
