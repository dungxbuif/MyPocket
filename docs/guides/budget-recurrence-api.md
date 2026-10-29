---
artifact_type: public_api_guide
status: implemented_local
updated: 2026-10-16
---

# Recurring budgets

A recurring budget creates a separate budget for each calendar period. It does
not create or change financial transactions. Existing `/api/v1/budgets` requests
continue to create explicit date intervals.

All routes require `Authorization: Bearer <owner JWT or API key>`. Responses use
the normal `{ "data": ... }` envelope. This Markdown page is also the stable
agent-readable recurrence reference. See the [approved design](../superpowers/specs/2026-10-16-savings-budget-design.md)
for the related savings and forecast contracts.

## Create a series

Send `POST /api/v1/budget-series` with a unique `Idempotency-Key` header and:

```json
{
  "name": "Food",
  "limit_amount": 3000000,
  "wallet_id": null,
  "category_id": null,
  "cadence": "monthly",
  "start_date": "2026-10-01",
  "carryover_enabled": false
}
```

Money is integer VND, from 1 through 9,007,199,254,740,991. Omitted wallet or
category means all owned wallets or all expense categories. Selected wallets
must belong to the owner and cannot be credit wallets. Categories must be
visible expense categories; a new selection must be active.

The start date must be canonical and within the current or a future account
calendar period. The example is valid during October 2026. No partial or
prorated initial period is generated.

| Cadence | Required first date | Inclusive period end |
| --- | --- | --- |
| `weekly` | Monday | Sunday |
| `monthly` | First day of the month | Last day of the month |
| `quarterly` | January, April, July or October 1 | Last day of the quarter |
| `yearly` | January 1 | December 31 |

Custom intervals cannot recur. Scheduling and transaction inclusion use the
current account timezone; changing timezone does not shift stored date labels.

Creation returns 201. Reuse the same key and payload after a timeout to receive
the original result with 200, including when the original start period has
since ended. A key is 1–128 printable ASCII characters. Reusing a key for a
different payload returns a mutation conflict.

## Read, edit and stop

`GET /api/v1/budget-series` returns `data.items`. Each series contains:

```json
{
  "id": "series UUID",
  "owner_id": "owner ID",
  "cadence": "monthly",
  "start_date": "2026-10-01",
  "active": true,
  "version": 1,
  "generation_error": null,
  "current_config": {
    "effective_start_date": "2026-10-01",
    "name": "Food",
    "limit_amount": 3000000,
    "wallet_id": null,
    "category_id": null,
    "carryover_enabled": false
  },
  "pending_config": null,
  "created_at": "2026-10-16T12:00:00Z",
  "updated_at": "2026-10-16T12:00:00Z"
}
```

Before a future series begins, `current_config` describes its initial
configuration. `pending_config` describes an additional scheduled replacement,
when present.

Send `PATCH /api/v1/budget-series/:id` with a fresh `Idempotency-Key` and the
latest version. Supply the complete replacement configuration:

```json
{
  "version": 1,
  "name": "Food and coffee",
  "limit_amount": 3500000,
  "wallet_id": null,
  "category_id": null,
  "carryover_enabled": true
}
```

The edit takes effect at the next period boundary, or at the initial start when
that start is later. Repeated edits before the boundary replace the pending
revision. Previously generated periods retain their name, scope, base amount
and carryover setting. Cadence and start date cannot be changed. Existing
archived category references remain usable, but an archived category cannot
be newly selected.

Send `POST /api/v1/budget-series/:id/stop` with a fresh `Idempotency-Key` and
`{"version": 2}` using the latest version. Stopping preserves existing periods
and prevents further generation. This release has no resume or series-delete
route. A stale version returns 409 `BUDGET_SERIES_VERSION_CONFLICT`; a new edit
or stop of an already stopped series returns 409 `BUDGET_SERIES_STOPPED`.

## Period generation and carryover

The worker and budget/series list requests catch up due periods under the same
owner lock. Each pass creates at most 100 instances for the owner; later passes
continue the cursor. A database uniqueness constraint prevents duplicate
`(series_id, start_date)` instances during concurrent API and worker execution.

An overlapping budget with the exact same wallet/category scope blocks that
series at the conflicting period. `generation_error` becomes `BUDGET_OVERLAP`.
The period is not skipped. Removing the conflicting explicit budget lets a
later pass retry. Deleting a wallet/category stops a series only when the scope
is needed by its next ungenerated period or a later revision. This includes
missed periods awaiting catch-up. The stop sets `generation_error` to
`"scope_deleted"` before existing budget scope cascades run. A superseded scope
used only by already-generated periods does not stop generation under the new
scope. Revision scope snapshots are preserved; a deleted reference never
becomes an all-wallet or all-category scope.

Generated budget rows expose `series_id`, `base_limit_amount`,
`carryover_amount`, `effective_limit_amount` and `carryover_enabled` on budget
list/report reads. `limit_amount` continues to mean the base amount. Direct
PATCH or DELETE of a generated instance returns 409 `BUDGET_SERIES_MANAGED`;
edit or stop its series instead.

Carryover defaults to false. When enabled for the next consecutive period with
the same wallet and category scope:

```text
carry[next] = max(0, effective_limit[previous] - spent[previous])
effective_limit[next] = base_limit[next] + carry[next]
```

The first period, disabled periods, gaps and scope changes have zero incoming
carry. Overspending never becomes debt. Backdated transaction changes
recalculate carry forward on reads. For example, a 3,000,000 base budget with
2,000,000 spent gives the next carry-enabled period 1,000,000 carry. Editing
the prior expense to 2,500,000 changes carry to 500,000.

Progress and active summary limits use effective limits. Expense selection is
shared with budget report/drilldown, including descendant categories, report
inclusion, owner, wallet and calendar dates; transfers, credit payments and
future rows after the server read time are excluded. Active summary expense
totals remain deduplicated across overlapping category scopes. Checked money
overflow returns 422 `BUDGET_AMOUNT_OVERFLOW` instead of a wrapped value.

## Verification and rollout

Local pure and real PostgreSQL race tests cover calendar boundaries, carry
recalculation and scope reset, positive-only carry, safe money limits,
idempotent replay, version conflicts, simultaneous generation/list access,
snapshot revisions, scope deletion, overlap recovery, and JWT/API-key owner
isolation. Migration `000035_budget_series.up.sql` is additive. Production
migration/deployment and owner/device acceptance are separate release gates.
