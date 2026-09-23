# Quản lý hũ

Route `/jars`; implements [TICKET-04](../../../work/tickets/TICKET-04-hu-chi-tieu.md) and [CORE-03](../../../work/tickets/CORE-03-TIME-JARS-MONTH-DETAIL_DESIGN.md). Status: implementation in progress.

## Composition

| Region | Base | Behavior |
| --- | --- | --- |
| Header/month choice | `PageBackHeader`, `BaseSelect`, `FormField`, `Text` | Back returns to the prior app route; month value is `YYYY-MM` in the saved account timezone. |
| Monthly figures | `SurfaceCard`, `Text`, `Progress`, `StatusMessage` | Displays actual income, ordinary expense, allocated total, unassigned spend, over-allocation and overspend advisories from the API. |
| Jar rows | `SurfaceCard`, `Text`, `Progress`, `BaseButton` | Each saved monthly config has its own name/allocation and API-derived actual spend. Inactive historical configs with transactions remain readable, not editable. |
| Add/edit form | `BaseBottomSheet`, `FormField`, `BaseTextInput`, `BaseSelect`, `BaseButton`, `StatusMessage` | User selects name, allocation mode (`Không đặt`, fixed VND, percent of actual included income), and the corresponding amount. |
| Cumulative view | `SurfaceCard`, `BaseSelect`, `Text`, `Progress`, `StatusMessage` | Select a stable jar from the owner's full jar catalog and view all-history totals/month breakdown; only configured months contribute allocation/difference. No carryover balance is implied. |

All styling, controls, typography, progress and cards are owned by the named bases. Screen classes are layout-only. No screen-local chart, currency palette, or clickable `div` is allowed.

## Events and effects

| Event | Effect |
| --- | --- |
| Enter `/jars` | Fetch selected month summary from the authenticated API; the server initializes that month once from the nearest earlier initialized configuration. |
| Change month | Fetch that month's actual config and ledger-derived spend. Historical months are viewable. |
| Add jar | POST one stable jar identity and the selected month's configuration; refresh month and cumulative options. |
| Edit monthly config | PUT the current month only; prior/future snapshots and the stable identity are not renamed. |
| Remove from month | Confirm, then deactivate only this month's config; keep stable ID, transaction links, and other snapshots. |
| Select a jar for cumulative | Fetch the stable-identity aggregate through the selected month, with all-history default range. |
| Save/API failure | Keep the draft, show retryable error, and do not optimistically change totals. |

## States and validation

- Loading/error: explicit status; no sample jars or fabricated zero allocations.
- Empty month: explain there are no jars configured and provide `Thêm hũ`; unassigned expenses remain valid and visible in summary.
- Ready: display real config and spend; warning states are advisory and never block ledger saves.
- Saving: disable duplicate actions and preserve draft until success.
- Allocation is optional. Fixed VND and percent are non-negative/positive according to API validation; percentages are bounded by backend rules. No allocation is represented as `Không đặt`.
- Only ordinary expense transactions can be assigned; income, transfer-out, and out-of-scope rows do not count as jar spend.
- Removing a monthly config never deletes a jar or linked ledger rows.

## Copy and API

Use `Hũ tháng`, `Chi chưa gắn hũ`, `Phân bổ`, `Chi thực tế`, `Cộng dồn`, `Tháng này`, and advisory copy such as `Tổng phân bổ vượt thu thực tế` / `Chi vượt mức hũ`. Never call cumulative difference “số dư hũ”. Money is VND; percentage is based on actual report-included ordinary income.

APIs: `/api/v1/jars?month=YYYY-MM`, `POST /api/v1/jars`, `PUT/DELETE /api/v1/jars/{jarID}/months/{YYYY-MM}`, and `/api/v1/jars/{jarID}/report?to=YYYY-MM` (optional `from` narrows the range; omission means all history). All access is owner-scoped.

## Proof and gaps

Implementation update: `/jars` now uses the authenticated APIs for month totals, create/edit/remove, prior-month configuration, stable-identity cumulative totals and month breakdown. Empty collections render as real empty states. Manual and AI transaction editors load choices for the transaction's account-local month and submit the optional `jar_id`.

Automated proof: PostgreSQL tests cover timezone rebucketing, empty-array response contracts, concurrent one-time copy of active config, and owner isolation; UI tests cover expense-only selection, hidden income/transfer-out controls, and account-local month derivation. Browser UAT must still verify create/edit/remove, month copy, manual and AI assignment, and cumulative history with owner-approved test data. No mock data, carryover, notifications, or AI inference is in scope.
