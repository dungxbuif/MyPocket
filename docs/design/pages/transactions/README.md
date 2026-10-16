# Transactions — list and basic income/expense editor

## Native physical-reference composition — 2026-10-16

[Approved wave design](../../../work/tickets/NATIVE-REFERENCE-WAVE1-20261016-DETAIL_DESIGN.md)
maps the actually observed ledger, category-group and overflow captures to these
existing native bases before implementation:

| Reference region | Native composition | Preserved behavior |
| --- | --- | --- |
| Wallet identity and balance | `MPWalletScopeSelector`, `MPText` | Centered owner-wallet scope; balance is the selected wallet's API balance, or the existing `is_in_total` aggregate policy. Money respects privacy masking. |
| Period strip | `MPSegmentedControl`, `MPWeekNavigator`, `MPDateRangePicker` | Existing Monday–Sunday week/custom range and future-week guard; no sample-account day-5 rule. |
| Cashflow and report action | Flat `MPMetricRow` label/value rows, `MPDivider`, `MPTextLinkButton` | Real filtered income, expense and net; existing Reports callback only when supplied. The action says “Xem báo cáo” because the existing route does not accept a week/custom period. No pending-outbox amounts. |
| Day/category group | One `MPSurfaceCard(.reference)` per group, existing `MPText` day heading | Account-local day number, today/weekday and month/year, signed group total; category grouping retains title and row count. |
| Transaction row | Existing `MPTransactionRow` inside `MPButton(.plain)` | Date grouping keeps category/note/amount semantics; category grouping shows account-local date and note below its category/count header. Existing amount/glyph mapping, detail route and bulk-selection eligibility remain shared. |
| Secondary actions | Existing `MPOverflowMenu` and `MPIconButton` add | Group/range/reload/bulk/transfer/adjustment stay available. Search is not an existing supported native action and no inert search control is added. |

This slice changes layout only. Goal and credit retain their specialized ledger
surfaces; offline recovery, owner guards, editor/detail/transfer actions and
account timezone remain authoritative. The bulk-selection panel must coexist
with the grouped rows so selection is reachable. Screen modifiers are layout
only; all typography, color, shapes and control hit targets remain base-owned.

Verification planned: focused balance/day-presentation unit tests, native
base/project guards, signed simulator unit target and Release compile, followed
by populated normal/large-text simulator screenshots compared with the actual
reference. Physical-device acceptance remains a separate gate. API/schema and
public setup/reference docs need no update because their contracts are unchanged;
this Markdown screen contract is the agent-readable implementation surface.

Local implementation checkpoint: `bash mobile-ios/scripts/check-base-components.test.sh`
and `git diff --check` pass. New `TransactionsReferenceTests` cover selected vs
included aggregate balance, local today/weekday, and category-row local date.
Test-first compilation captured missing presentation helpers; concurrent base
tests were also intentionally red while their shared implementations were being
added. The parent integration run owns signed unit/build results and populated
normal/large-text screenshots; visual completion is not claimed here.

## Native queued creates — 2026-10-14

DefaultAuth/root injects FinancialMutationQueue into shared ordinary/transfer
create forms. State UUID per editor, encryptedenqueue before HTTP, no online
fallback after persistedintent. Existing edit/delete/credit/adjustment online.
MPFinancialQueueStatus composes existing card/text/badge/button/overflow menu above ledger,
never merges pending amounts into confirmed cashflow. Receipt+CAS revision reloads
ledger/overview. Peritemrecovery, pendingUI/a11y and full E2E open;
matchingbackend rollout required before installation.
[Work](../../../work/tickets/OFFLINE-RUNTIME-QUEUE-20261014.md).

Foreground lifecycle is now locally wired through AppRoot scene+owner and SDK
reachability: one FIFO retry/lease wake, no writes while inactive; pause differs
from auth-block. Actual Home/server-recovery/foreground auto-resume plus restart/
manual retry and commit/drop replay pass3/3. Existing confirmed totals remain
separate from pending rows. Full recovery/a11y/device UAT remains open.
[Lifecycle](../../../work/tickets/OFFLINE-FOREGROUND-LIFECYCLE-20261014.md).

Ordinary queued creates and changed replacements require a nonempty trimmed
category identifier before any insert/cancel; existing encrypted records remain
readable, not rewritten. Existing frozen replay preserves prior transport/key/body
semantics so server receipt/rejection remains recoverable. Transfer does not use ordinary category input. Backend
still validates actual category visibility/kind/wallet eligibility. See
[validation fix](../../../work/tickets/OFFLINE-CATEGORY-VALIDATION-20261014.md).

Owner-approved compact queue: each row keeps type, amount, wrapped note and status.
Its 44pt MPOverflowMenu (`Thao tác bản nháp`) exposes eligible recovery actions,
disabled while sync/action is active; sending rows have no menu. No per-item
full-width action buttons, no new palette/font/shape. Global Sync remains visible.
Feedback overlay placement is separately blocked, not solved by this adjustment.

Edit eligibility is evaluated per queued item: a goal-wallet transfer draft may
open the existing transfer editor with wallet data alone, while a transaction
draft waits for categories. `MPFinancialQueueStatus` owns this predicate so the
shared menu stays consistent across aggregate, basic and savings ledgers.

Recovery code checkpoint: MPFinancialQueueStatus uses peritem RetryNow, Edit for
conflict/rejected, and shared MPConfirmDialog for local-only discard with uncertain
server-result warning. Edit reuses ordinary/transfer editors, prefills frozen
intent and confirms changed draft with newkey/atomic replacement; original date
preserved when control unchanged. New real API UI proof covers ordinary422/409
unchanged-refusal/amount replacement/previous-day preservation; discard warning
cancel+confirm and restart retain committed server rows or leave rejected writes
absent. Transfer503/commitdrop restart/retry produces one pair, correct balances
and unchanged monthly report aggregates. No silent conflict overwrite. See
[recovery E2E](../../../work/tickets/OFFLINE-RECOVERY-E2E-20261014.md).
Goal-scoped queue entry, rejected transfer editor UI, broader auth/cache/device/
a11y acceptance remain separate gates; this is not full recovery UAT.

Native transaction detail groups its screen with `.accessibilityElement(children:
.contain)`: one `transaction-detail-overlay` identity contains the header,
metrics, note and individual action controls. Do not propagate that screen ID
onto sibling cards/buttons or combine/hide the controls. Confirmed-row recovery
tests open by server row ID and query the note within this named detail group.
This is accessibility structure, not a new style/base family.

## Native cached read — 2026-10-14

Default native Auth composition injects an encrypted owner/server-specific
FinancialSnapshotCache into this ledger. Successful complete references save
confirmed rows/wallets/categories/timezone; transient failures may restore them.
MPStatusMessage displays fetchedAt and explicitly stale status above cashflow,
MPButton offers reload; no new card/color/font. Auth/ownership/malformed errors
never fall back. Pending outbox rows are not mixed into cashflow. Valid saved
session required; credit/goal subroutes, overview/report cache and offline writes
remain separate work. Local unit/model and normalledgerAPIUI proof recorded in
[cache work](../../../work/tickets/OFFLINE-FINANCIAL-CACHE-20261014.md); actualstale
banner/accessibility/landscape UAT not yet accepted.

## Native note editor — 2026-10-14

`transaction-note-row` opens `mpFormBottomSheet` + shared `MPTextEditor` with a
separate note draft. **Lưu ghi chú** copies that draft to the transaction form
only; it does not call the API. Closing/swiping the note sheet discards those
note-only edits. Reopening starts from the current transaction draft. Outer
transaction cancel discards all form changes. **Thêm chi tiết** contains only
the report-inclusion checkbox, not a duplicate note field. API payload remains
the existing optional `note`; multi-line content is preserved and trimmed at save.
SE UI2/2 covers save/cancel and existing editor controls; API persistence proof
is recorded separately in [validation](../../../work/VALIDATION_MATRIX.md).
Global raised Add is hidden on Assistant/Account to avoid obscuring input/menu;
it remains on Overview/Transactions. Historical web layout below is archived.

The transaction options menu also exposes explicit balance adjustments. A user records a positive amount as either `Tăng số dư` or `Giảm số dư`; the row remains visible in the ledger and changes the wallet balance, but never changes income/expense report totals. Adjustment rows are immutable so a correction is made by recording the opposite adjustment.

Status: implementation contract for [TICKET-02-01](../../../work/tickets/TICKET-02-01-ghi-thu-chi.md) and its [detail design](../../../work/tickets/TICKET-02-01-DETAIL_DESIGN.md). Route: `/transactions`; the global add action opens the same editor from every primary tab.

## Composition

The page header uses the shared `ScreenHeader` molecule so title/action spacing matches Budgets and Account screens; the options action remains a `BaseButton` chip.

2026-09-24 owner correction: the supplied savings screenshot's “TẤT CẢ CÁC GIAO DỊCH” is a goal-history heading, not a shared All-time tab. Keep one wallet selector at the top. `Tổng cộng` and selected basic wallets use account-local Monday–Sunday Week/Custom controls, period cashflow and date-grouped rows. Credit wallets use a dedicated debt statement with current debt, available limit, credit rows, and actions for purchase/refund/fee/interest or payment from another non-credit wallet; ordinary period, bulk-delete and transfer/adjustment controls are hidden. Goal alone shows distinct balance/remaining/progress and complete history. [Correction design](../../../work/tickets/UI-LEDGER-WEEK-DETAIL_DESIGN.md#owner-correction--wallet-specific-ledger-2026-09-24).

2026-09-20: editor uses grouped SurfaceCards for wallet/large amount/category/note and date, retaining the shared inputs and form-sheet shell. Goal category selection is restricted to real savings catalog keys and backend enforces the same rule; ordinary income/expense and credit entries require a category, with the first applicable group prefilled when needed. Unsupported old goal categories remain visible in history but must be cleared/reselected before editing. Report flag for external entries defaults true. This is not the paired-wallet transfer flow.

| Region | Base implementation | Contract |
| --- | --- | --- |
| Wallet scope filter | Web: `WalletScopeSelector` + `WalletSelectionList` + `BaseBottomSheet`; iOS: `MPWalletScopeSelector` + `MPMenuSelector` + `MPWalletScopeGlyph` | Always visible below the header as a compact, centered wallet capsule. Web opens the shared `Chọn Ví` sheet with `Tổng cộng`, included/excluded wallet groups, real balances, add-wallet and disabled-link states. iOS keeps the same compact capsule anatomy and native menu selection. `Tổng cộng` is the aggregate wallet scope, not an All-time period. The selector uses the wallet icon and semantic type tone; basic is neutral slate, goal is success, credit is teal. Switching a goal wallet selects its specialized ledger body; other scopes use the same selector. The account wallet-management list uses the same base list, and selecting a wallet routes to `/transactions?wallet_id=<id>` rather than opening a separate wallet-detail screen. |
| Period and week navigation | `SegmentedControl`, `BaseButton`, `Text` | Aggregate/basic: open to the current account-local week when it contains rows, otherwise the newest past week with rows in the selected scope; if no rows exist, remain on the current week. Previous/next week and custom range are available; no All-time tab. Goal and credit: hide ordinary period control. |
| Date group | `SurfaceCard`, `Text` | Groups real API rows by the user's local calendar date and displays the signed group total. The header uses a large day number plus localized weekday and month/year metadata; category grouping keeps its category label/count variant. |
| Transaction row | `TransactionItem` / native `MPTransactionGlyph` + `MPTransactionRow` | Category icon/title, optional note (or short date when the note is empty), signed VND amount; do not repeat the wallet name when the active wallet scope already identifies it. Activation opens edit mode. Income uses the action tone, expenses use danger, and transfer/adjustment rows use neutral tone; adjustment rows cannot be edited or deleted through the ordinary editor. |
| Editor shell | `BaseBottomSheet` | Modal focus trap, Escape/backdrop cancel, and return focus follow the shared sheet contract. |
| Type, amount and metadata | `AmountField`, `DateField`, `FormSelectorRow`, `CategoryTreeSelector` → `BaseCategoryTree` selection mode, `FormField`/`BaseSelect` for optional jar, `BaseTextInput`, `BaseSwitch` | The transaction-specific amount base places a VND chip before a large unboxed value and applies `vi-VN` thousands grouping to display only; bound/API values remain integer digits. Other amount consumers retain the standard bordered-field presentation. The global add sheet is ordinary `expense`/`income` only. Transfer uses a separate transfer-only sheet opened from the transactions screen's three-dot options menu; it uses the standard shared amount/date/note bases plus two wallet selects and never shows category, jar, or report controls. |
| Feedback/actions | `StatusMessage`, `BaseButton` | Save has loading/disabled behavior; delete exists only in edit mode and requires confirmation. |

No screen-local color, shape, input, button or card styling is allowed. Debt, receipt/OCR and advanced filter controls stay out of this screen until their own approved slices exist.

The native transaction editor exposes report inclusion as the shared
checkbox-row base. It is selected by default for a new entry and editing uses
the transaction's persisted report-inclusion value.

## Events and effects

| Event | State/effect |
| --- | --- |
| Enter `/transactions` | Load transactions, wallets and visible groups from their authenticated APIs. Never replace an error with mock rows. |
| Select wallet scope | Recompute the real ledger view locally without another API call; `Tổng cộng` clears the wallet predicate and removes `wallet_id` from the URL. iOS remembers the selected wallet locally per authenticated owner and clears a stale selection if that wallet is no longer available. Selecting a wallet from the account wallet-management screen navigates to `/transactions?wallet_id=<id>`; the transactions screen parses that optional query and selects the current account-local week when it has rows, otherwise the newest past week containing rows for that wallet. |
| Select a goal wallet | Show real savings summary/progress and monthly history; do not show ordinary weekly cashflow cards or an All-time period control. |
| Select a basic wallet | Keep the aggregate layout, including Week/Custom control and cashflow card; filter both totals and day rows to the selected wallet and current period. |
| Select an aggregate/basic scope or step week | Scope changes select the current account-local Monday–Sunday week if it has rows, otherwise the newest past week with rows in that scope; an empty ledger remains on the current week. Step through prior weeks and never navigate into a future week. A timezone change recalculates the week from the account-local current date. Credit uses its statement instead. |
| Select `Tùy chọn` | Open the existing date-range dialog. A valid range filters both summary and rows; cancel keeps the previous period choice and dates. |
| Return from custom range to week | Show the current account-local week. There is no All-time period tab for ordinary scopes. |
| Open an expense editor | Load jar choices for the transaction's account-local month; show only active monthly configurations, plus the existing historical assignment when editing. |
| Activate global add | Open a clean ordinary editor with `expense`, current local date/time, report inclusion enabled, and the first available wallet. It does not offer transfer mode. |
| Choose “Chuyển tiền đến ví khác” from the three-dot options menu | Open a transfer-only sheet with source/destination wallet selects; submit calls the atomic transfer endpoint and refreshes both wallet balances and the ledger. |
| Choose “Điều chỉnh số dư” from the three-dot options menu | Open the shared form sheet, choose an eligible wallet and increase/decrease direction, then call the adjustment endpoint; refresh balances and ledger without changing report totals. |
| Choose “Xóa nhiều giao dịch” from the three-dot options menu | Enter selection mode for the current ledger scope. The UI excludes transfer/adjustment rows, requires confirmation, sends at most 100 unique IDs, and keeps selection/error state when the atomic request fails. Native iOS must expose the same behavior through `MPButton`/icon bases and `POST /api/v1/transactions/bulk-delete`. |
| Hold global add for 500ms | The shared native/web add base opens [AI entry chat](../assistant/README.md) with persistent prefilled review proposals. Release does not also open the manual editor; manual create offers keyboard-accessible “Nhập bằng AI”. |
| Change wallet | Keep the selected group only if it applies to the new wallet; otherwise clear the group. |
| Change type | Keep the selected group only if its kind matches; otherwise clear the group. |
| Change transaction month | Reload that month's jar configurations; keep historical assignment visible for review, but clear or choose an active jar before moving a linked transaction to a different instant/month. |
| Change type to income / select transfer-out category | Clear and hide optional jar selection; backend remains the final eligibility guard. |
| Submit valid editor | Convert local datetime to RFC3339 UTC, include optional `jar_id`, call create/update, close on success, then refresh transactions and wallets. |
| Submit invalid editor | Keep entered data and show field/form feedback; do not call the API. |
| Activate a transaction | Open the editor populated from that API row. |
| Open the group selector | Search the category tree, choose an applicable root/child group, or clear the selection; the picker composes the same `BaseCategoryTree` selection mode used by group management. Non-applicable parents may remain visible only as hierarchy context. Native iOS uses the `MPCategoryPicker` sheet and retains the same required-category and wallet/type eligibility rules. |
| Delete in edit mode | Show destructive confirmation, delete through API, close, then refresh transactions and wallets. |
| Cancel/Escape/backdrop | Close without persistence. No autosave. |
| API failure | Keep current data, show retryable error, and leave the editor open when a save/delete failed. |

## States

- Initial/loading: neutral status; list actions that require loaded data are unavailable.
- Ready: real API rows grouped newest first.
- Empty: explain that no transactions exist and direct the user to the global add action. Use `StatusMessage variant="plain"` without a card border, background or shadow; Overview uses the same variant for its empty transaction message (owner request 2026-09-20).
- Error: danger status with an explicit retry action; stale successful data may remain visible but mock data must not appear.
- Saving/deleting: submit controls disabled/loading; one request per activation.
- No wallets: editor explains that a wallet must be created first and cannot submit.
- No matching groups: the form blocks saving and explains that an applicable group is required. Adjustment rows are created only through the dedicated adjustment flow and may remain uncategorized.

## Validation and source of truth

- Backend owns authorization and final validation. Client validation is recovery guidance, not a security boundary.
- `wallet_id` must belong to the authenticated owner. Basic income/expense accepts `basic` and `goal`; credit entries use the separate debt/payment ledger and dedicated credit endpoints.
- `type` is `income` or `expense`; `amount` is an integer greater than zero.
- Transfer mode requires two different non-credit owner wallets and a positive integer amount; the API creates the paired rows and sets report inclusion false.
- Editing a transfer updates amount, date and note on both linked rows. Source and destination wallets are fixed by the current update API; show them as read-only values, not editable selects. Moving a transfer between wallets is not supported by this contract.
- Category is required for ordinary income/expense and credit entries. It must be visible, match the transaction type, and either have no wallet restriction or include the selected wallet ID. Only adjustment rows may omit it.
- `occurred_at` is sent as an RFC3339 timestamp. Note is trimmed and omitted when empty.
- `included_in_reports` defaults to true.
- `jar_id` is optional and may reference one active owner/month configuration only for ordinary expenses; income and transfer-out rows cannot be assigned. Clearing the selection sends null.
- Wallet current balance is derived from ledger rows: opening balance plus income minus expense, with adjustments adding or subtracting by direction. Credit balance follows the same signed ledger, while available credit is `credit_limit + current_balance`; credit rows are mutated only through the dedicated credit API. Editing/deleting an ordinary row changes the derived balance; an adjustment is immutable and must be compensated by an opposite adjustment.

APIs: `GET/POST /api/v1/transactions`, `POST /api/v1/transactions/transfer`, `PATCH/DELETE /api/v1/transactions/transfer/{transfer_id}`, `POST /api/v1/transactions/adjustment`, `POST /api/v1/transactions/bulk-delete`, `PATCH/DELETE /api/v1/transactions/{id}`, `GET /api/v1/wallets`, `GET /api/v1/categories`, and `GET /api/v1/jars?month=YYYY-MM` for active jar options. Credit uses `POST /api/v1/credit/wallets/{id}/entries`, `POST /api/v1/credit/wallets/{id}/payments`, and `GET /api/v1/credit/wallets/{id}/statement`. Recurring schedules are managed separately from this editor at `/account/recurring`.

The native recurring editor also uses `MPCategoryPicker` for category choice;
it preserves the existing selected category ID and transaction-type/wallet
eligibility policy when the schedule is created or edited.

## Copy and formatting

Use Vietnamese product copy. Money uses signed integer VND via the shared formatter. Transaction title prefers category name and falls back to `Khoản thu`/`Khoản chi`; the ledger subtitle shows the optional note, or the local short date when there is no note. Do not repeat the wallet name inside the ledger row because the active scope already identifies it. The account timezone is the source of truth.

## Proof and residual gaps

Automated proof covers backend owner scope, positive amount/type validation, wallet ownership, category kind and wallet applicability, report flag, optional jar-link validation, and frontend design/build guardrails. `npm run test:transaction-jars` renders the real form and verifies jar selection for ordinary expenses while hiding it for income/transfer-out. Manual UAT still needs to cover create/edit/delete, jar assignment/clear, and refreshed wallet totals.

2026-09-24 weekly-ledger proof: `test:calendar` covers Monday–Sunday and year boundaries in account-local time; `test:transactions` covers wallet + week predicate at UTC boundary; the SSR period-selector fixture covers tab labels/range/future-week disable. Chrome on the local preview showed All → 21–27/9 → 14–20/9 with cashflow/list updates, empty state, next-week disable, and custom-dialog cancel preserving week. Build/design checks pass. Owner mobile visual acceptance remains open. [User guide](../../../guides/transactions.md).

Transfer proof now includes `npm run test:transfer:e2e` (real Chromium UI/state) and `npm run test:transfer:integrated` (Vite proxy/API/PostgreSQL state). Residual gaps remain separate tickets: attachments/OCR, advanced search/filter/pagination, and owner UAT sign-off.
