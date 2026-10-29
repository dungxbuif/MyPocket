---
artifact_type: detail_design
id: FEEDBACK-fc8cae65-investment-account
status: deferred_backlog
updated: 2026-10-26
approval: deferred_by_owner_2026-10-26
source_feedback: fc8cae65
---

# Balance-based investment accounts and gold holdings

## Decision and boundary

Owner's product direction for feedback `fc8cae65` ("Tài sản" / "Chưa có UI
để nhập các tìa khoản biến động"): add a separate **Tài khoản đầu tư theo số
dư** for an account whose total value changes. A value change is asset
valuation, not income. Deposits/withdrawals are transfers. Do not count the
same wealth in both this account and itemized holdings. Gold stays a portfolio
asset with quantity, purchase unit price, current unit price, and the existing
weighted-average cost rule; distinguish gold type and unit.

This document records a possible future design only. On 2026-10-26 the owner
explicitly deferred the feature to backlog and requested no schema, API, UI,
deployment, or feedback-resolution work now. Existing itemized portfolio and
gold behavior remain unchanged.

## Current system and problem

- `portfolio_assets`/`portfolio_trades` represent instruments and immutable
  buy/sell history. The summary derives quantity, cost basis, and market value
  from a per-unit price. Native `Danh mục đầu tư` exposes asset, trade, and
  manual price editors. It does not represent an opaque account balance.
- `wallets` allows only `basic`, `goal`, and `credit`. Its paired transfer API
  moves money between two wallets; portfolio trades and prices intentionally
  do not alter wallet balances or ordinary income/expense reports.
- Requirements describe net worth components, but no current backend net-worth
  composition endpoint or historical portfolio price ledger exists. A new
  account cannot simply be added to a wallet/portfolio total without defining
  ownership of the underlying wealth.

Sources: [feedback work note](../bugs/FEEDBACK-fc8cae65-variable-asset-ui.md),
[portfolio rules](../../requirements/BUSINESS_RULES.md),
[portfolio requirements](../../requirements/SPEC.md),
[report requirements](../../requirements/REPORTS.md),
[current API](../../architecture/API.md), [ERD](../../architecture/ERD.md),
[portfolio screen](../../design/pages/portfolio/README.md),
[native phase](../phases/PHASE-002-native-ios-swiftui.md).

## Goals, non-goals, acceptance

Goals: enter a named VND account's reported total; record later total-value
snapshots; move principal between a managed cash wallet and the account
without income/expense; expose a current wealth breakdown that cannot silently
double-count itemized positions; keep gold on the itemized quantity/price path.

Non-goals: brokerage synchronization, securities lots, automatic gold quotes
or unit conversion, per-holding allocation within an aggregate account,
cross-currency balances, historical net-worth charts, changing wallet type,
and automatic classification of an unexplained value change as taxable income
or realized P/L. Existing cashflow reports and portfolio trade math remain
unchanged. Direct deposit from an untracked external account is not a V1
transfer; the user first records the funding source in a managed wallet or
uses a clearly labeled pre-existing opening value.

Acceptance to prove after approval/implementation:

1. Opening at 0, transferring 100 VND from a cash wallet, recording a 110 VND
   total, then withdrawing 20 VND produces account value 90 VND and valuation
   change +10 VND; transfer steps contribute 0 VND to income/expense and
   preserve combined wealth before valuation.
2. A value snapshot never creates a wallet transaction. The wallet cash leg
   and account transfer leg commit atomically, replay idempotently, and cannot
   be edited/deleted through ordinary transaction routes.
3. A current combined wealth total is returned only when every included
   account has confirmed coverage and every included standalone holding has a
   known price. Whole holdings covered by an aggregate account are excluded
   once, not counted again. Partial holding overlap is rejected in V1.
4. Existing holdings and trades retain their quantities/costs after migration.
   A gold asset displays its type and unit; each trade and current price uses
   that same immutable unit, with weighted-average cost unchanged.
5. Owner isolation, stale-write rejection, reversal, missing-price, overflow,
   and native readback/error states pass focused tests. Physical owner UAT and
   production-connected readback remain separate gates.

## Approaches considered

| Approach | Benefit | Cost / reason |
| --- | --- | --- |
| Add an `investment` wallet type | Reuses wallet list and paired transfers | Wallet balances and ordinary transaction/category/report rules would treat market valuation as spendable cash; high regression and double-count risk. Not recommended. |
| Model one synthetic portfolio asset per account | Reuses price/summary fields | Quantity and buy/sell history become artificial; a deposit could look like a purchase and a value change like a unit-price move. Confuses gold and account semantics. Not recommended. |
| Separate aggregate account, snapshot and cash-transfer ledgers | Names the two different kinds of change, preserves current wallet/holding contracts | New schema/API/report projection and guarded cash-leg integration. Recommended. |

## Proposed model and calculation

New owner-scoped `investment_accounts`: `id`, `owner_id`, name, `currency=VND`,
`version`, `coverage_status` (`unconfirmed`, `disjoint`, `mapped`), active/archive
state, timestamps. Creation requires an opening **reported total** snapshot,
including explicit zero, and an effective timestamp. A nonzero opening value
is labeled pre-existing wealth, not income; funding from a tracked wallet must
instead use the transfer flow. Account names need not be unique; IDs are.
Archive is allowed only after value reaches zero and coverage mappings are
released, so hiding an account cannot silently erase wealth from the total.

Append-only `investment_account_valuations`: account/owner ID, `total_value`
(nonnegative integer VND), effective time, note, created time, stable ID.
An incorrect snapshot is corrected by an immutable superseding valuation with
reason and the same effective time; replay uses the latest valid revision and
retains prior rows for audit. Corrections require account version and an
idempotency key, not direct row update/delete.
Append-only `investment_account_transfers`: account/owner ID, managed wallet
ID, direction (`deposit` or `withdrawal`), positive amount, effective time,
linked wallet transaction ID (unique FK with delete restricted), idempotency
receipt, optional reversal-of ID, created time. Repository transactions verify
both owner scopes, wallet type (`basic`/`goal`, never `credit`), amount limits,
and version. A dedicated
reversal writes opposite legs; original rows remain audit history. No direct
PATCH/DELETE of financial events in V1.

Replay events by `(effective_at, created_at, id)` with a documented tie rule.
A valuation snapshot sets the account's reported value; a later deposit adds
and a later withdrawal subtracts principal. At each snapshot, the derived
valuation change is `new reported total - value immediately before snapshot`.
It is a valuation component, not a ledger transaction or income category.
Reject withdrawal that makes the account negative, negative snapshots,
overflow, events before opening, and stale `expected_version`. Backdated
events after opening trigger deterministic replay; API returns recomputed
current value and valuation history. This calculation must never infer
income or realized P/L from a balance change.

`investment_account_coverage` maps whole `(owner_id, portfolio_asset_id)`
positions to at most one aggregate account, with an owner-checked foreign key
and unique asset ownership within coverage. `disjoint` means the user confirms
the account contains none of their itemized holdings; `mapped` means selected
whole holdings are included in its reported total. A newly created account is
`unconfirmed` until that explicit choice. A holding cannot be split between
an aggregate account and standalone portfolio in V1. Coverage changes are
versioned/audited and require recomputing the current wealth view; historical
coverage is not represented, so V1 does not publish a historical net-worth
series. If coverage is unconfirmed or partial overlap is reported, the API
returns `total=null` with an explicit reason, never a plausible doubled sum.
When complete, `total = included non-credit wallet balances - credit
liability + aggregate account values + priced standalone portfolio values`.
Apply the existing wallet `is_in_total` policy and credit sign convention
exactly once; missing standalone prices also make `total=null`.

Gold remains in `portfolio_assets`: propose `asset_kind` (`generic`, `gold`),
`unit_code` (`unit` for existing generic assets; a selected gold unit), and
`gold_type` as a user-visible instrument label. Migration defaults all existing
assets to `generic`/`unit` without changing trades or prices. A gold asset's
unit is fixed once it has trades; quantity and every VND unit price refer to
that unit. Different gold type/unit combinations are separate instruments.
No conversion or quote-source matching occurs implicitly. Exact allowed gold
units and whether `gold_type` is free text or a controlled list need owner
approval before schema/API implementation.

## API and security proposal

- `GET/POST /api/v1/investment-accounts`, `GET/PATCH
  /api/v1/investment-accounts/{id}`: create/list/read and rename/archive;
  archive retains financial history. No hard delete in V1.
- `GET/POST /api/v1/investment-accounts/{id}/valuations`: append snapshots;
  `POST` requires `expected_version` and `Idempotency-Key`.
- `POST /api/v1/investment-accounts/{id}/valuations/{valuation_id}/correct`:
  immutable, reasoned replacement with version/idempotency guards.
- `GET/POST /api/v1/investment-accounts/{id}/transfers` and `POST
  /api/v1/investment-accounts/{id}/transfers/{transfer_id}/reverse`: atomic
  cash/account legs and audit-safe reversal, with `expected_version` and
  `Idempotency-Key` on writes.
- `PUT /api/v1/investment-accounts/{id}/coverage`: explicit `disjoint` or
  complete whole-position asset ID set, owner-scoped and version-guarded.
- `GET /api/v1/reports/net-worth`: **current-only** component projection,
  not a replacement for existing income/expense or portfolio summaries.
  Return wallet cash per existing `is_in_total` policy, credit liability once,
  aggregate account values, standalone portfolio values, excluded covered
  asset IDs, `as_of`/staleness and completeness reasons. `total` is nullable.

All routes require existing owner JWT/API-key authority. Validate monetary
safe-integer/overflow bounds, timestamps, request body fields, owner FKs,
replay key/hash conflict and version CAS server-side. The unique
`investment_account_transfers.wallet_transaction_id` FK identifies the cash
leg. Ordinary edit/delete/bulk-delete and AI transaction mutations must check
that link and reject the row, as they already guard paired wallet transfers.
The cash leg is report-excluded and jar-ineligible.
Create/reverse must commit account event, wallet row, and replay receipt in one
PostgreSQL transaction. No MCP/AI write tool is included in V1.
Before rollout, classify every new route in the existing MCP/API capability
inventory as an explicit non-tool or obtain separate approval for a tool;
this design does not change that shared inventory.

## Native UI and reporting

Under `Ví của tôi`, add a separate **Tài khoản đầu tư theo số dư** destination
beside **Danh mục đầu tư**. List rows show account name, current total,
snapshot time/staleness, and a clear incomplete-coverage marker. Long-form
create/editor uses existing `mpFullScreenCover`/`MPEditorScaffold`,
`MPTextField`, `MPAmountField`, `MPDateTimeField`, `MPOptionSelector`, and
`MPButton`. Short snapshot entry/correction uses `mpFormBottomSheet`; transfers
use the existing wallet selector and amount field, with explicit source/destination
copy. `mpConfirmDialog` confirms reversal/archive. History separates
**Chuyển tiền** from **Định giá**; a valuation delta is never labeled **Thu
nhập**. Busy/error states retain the draft, and success reloads server data.

Coverage choice is part of onboarding, not a dismissible warning. Show the
portfolio asset list as whole-position choices when `mapped`; reject partial
coverage and keep combined total unavailable until resolved. The report UI
shows cash, credit, aggregate accounts, standalone holdings, excluded covered
holdings, and why `total` is unavailable. It must not silently coerce missing
prices to zero. Gold creation in the existing portfolio editor gains gold
type/unit controls and shows the unit alongside quantity and both prices; it
does not use the balance-account editor. Reuse native MP bases above and
`MPMetric`/`MPSurfaceCard` for readback; any missing base behavior is designed
in the shared system before screen implementation.

## Migration, rollout, and rollback design

Add new tables and constraints in a forward migration; add nullable/defaulted
gold metadata columns to `portfolio_assets`. No existing asset, wallet,
transaction or trade is reclassified/backfilled into an investment account.
Historical `expense_investment` wallet rows are not inferred to be transfers;
onboarding must make their reconciliation explicit rather than silently
rewriting cashflow history.
Indexes cover owner/account and event order; DB checks and uniqueness enforce
positive transfers, nonnegative snapshots, one reversal, and one aggregate
coverage owner per asset. The implementation plan must validate schema on a
disposable migrated PostgreSQL database, including down-migration only while
new tables are empty. Once real account data exists, rollback is feature
disable/forward repair, not destructive table drop. API deploy precedes native
UI release; no production rollout is part of this design task.

## Verification plan (not run in this design task)

- Domain unit tests: event replay example in acceptance, zero vs missing,
  transfer after snapshot, backdated ordering, valuation delta, negative and
  overflow rejection; existing gold weighted-average/fractional-quantity tests.
- PostgreSQL integration: migration/default preservation, owner isolation,
  CAS/idempotency/concurrent transfer, atomic failure rollback, protected
  wallet cash leg, reversal, archive guard, coverage uniqueness and nullable
  combined total.
- HTTP contract: auth, malformed amount/time/coverage, foreign wallet/asset,
  stale version, key replay/conflict, readback of components and reasons.
- Native XCTest/API-backed UI: account create, snapshot, deposit/withdrawal,
  coverage gate, error/draft retention, gold unit/type readback, and wallet +
  report unchanged by valuation. Use disposable local fixtures only.
- Run focused Go tests with migrated disposable PostgreSQL, native model and
  API-backed UI tests on an isolated simulator, base/project guards, and docs
  review. Then perform owner physical-device and production-connected UAT
  separately. No result is claimed here.

## Approval decisions and docs reconciliation

Before implementation, owner must confirm: (1) V1 whole-position overlap
mapping and nullable combined total for unresolved/partial overlap; (2) exact
gold unit set and whether gold type is free text or a catalog; (3) native-first
delivery and managed-wallet-only transfers, with pre-existing opening values
for older external assets. These choices affect schema, API and report meaning.

Impact: API=yes, DB/migration=yes, Security/owner scope=yes, Runtime/rollout=yes,
Standards=no. Small-task exemption: **no**. An approved detail design and
implementation plan are required before code. A new ADR should record the
aggregate-vs-holding ownership and cash/valuation boundary. After actual
implementation, reconcile [SPEC](../../requirements/SPEC.md),
[BUSINESS_RULES](../../requirements/BUSINESS_RULES.md),
[REPORTS](../../requirements/REPORTS.md), [API](../../architecture/API.md),
[ERD](../../architecture/ERD.md), [portfolio screen](../../design/pages/portfolio/README.md),
native human/AI-agent [guide](../../guides/portfolio-ios.md),
[validation matrix](../VALIDATION_MATRIX.md), and release notes. The shared
[context](../../CONTEXT.md) and [backlog](../BACKLOG.md) are intentionally not
edited during this parallel design task; their owner coordinates later updates.
