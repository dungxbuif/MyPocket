# Transactions — list and basic income/expense editor

Status: implementation contract for [TICKET-02-01](../../../work/tickets/TICKET-02-01-ghi-thu-chi.md) and its [detail design](../../../work/tickets/TICKET-02-01-DETAIL_DESIGN.md). Route: `/transactions`; the global add action opens the same editor from every primary tab.

## Composition

The page header uses the shared `ScreenHeader` molecule so title/action spacing matches Budgets and Account screens; the options action remains a `BaseButton` chip.

2026-09-24 owner correction: the supplied savings screenshot's “TẤT CẢ CÁC GIAO DỊCH” is a goal-history heading, not a shared All-time tab. Keep one wallet selector at the top. `Tổng cộng`, selected basic wallets and the provisional read-only credit scope share the same ledger layout: account-local Monday–Sunday Week/Custom controls, period cashflow card and date-grouped rows. Selecting a basic wallet filters both totals and rows by wallet ID; it does not create a special body or show all-time history. Goal alone shows distinct balance/remaining/progress and complete monthly history. No API/schema change or server-side week query is introduced. [Correction design](../../../work/tickets/UI-LEDGER-WEEK-DETAIL_DESIGN.md#owner-correction--wallet-specific-ledger-2026-09-24).

2026-09-20: editor uses grouped SurfaceCards for wallet/large amount/category/note and date, retaining the shared inputs and form-sheet shell. Goal category selection is restricted to real savings catalog keys and backend enforces the same rule; category remains optional. Unsupported old goal categories remain visible in history but must be cleared/reselected before editing. Report flag for external entries defaults true. This is not the paired-wallet transfer flow.

| Region | Base implementation | Contract |
| --- | --- | --- |
| Wallet scope filter | `FormField` + `BaseSelect` | Always visible below the header. `Tổng cộng` is the aggregate wallet scope, not an All-time period. Switching a goal wallet selects its specialized ledger body; other scopes use the same selector. |
| Period and week navigation | `SegmentedControl`, `BaseButton`, `Text` | Aggregate/basic/credit: current week by default, previous/next week and custom range; no All-time tab. Goal: hide period control and show full goal history. |
| Date group | `SurfaceCard`, `Text` | Groups real API rows by the user's local calendar date and displays the signed group total. |
| Transaction row | `TransactionItem` | Category icon/title, wallet and optional note, signed VND amount; activation opens edit mode. |
| Editor shell | `BaseBottomSheet` | Modal focus trap, Escape/backdrop cancel, and return focus follow the shared sheet contract. |
| Type, amount and metadata | `AmountField`, `DateField`, `FormSelectorRow`, `CategoryTreeSelector` → `BaseCategoryTree` selection mode, `FormField`/`BaseSelect` for optional jar, `BaseTextInput`, `BaseSwitch` | The global add sheet is ordinary `expense`/`income` only. Transfer uses a separate transfer-only sheet opened from the transactions screen's three-dot options menu; it uses the same amount/date/note bases plus two wallet selects and never shows category, jar, or report controls. |
| Feedback/actions | `StatusMessage`, `BaseButton` | Save has loading/disabled behavior; delete exists only in edit mode and requires confirmation. |

No screen-local color, shape, input, button or card styling is allowed. Debt, receipt/OCR and advanced filter controls stay out of this screen until their own approved slices exist.

## Events and effects

| Event | State/effect |
| --- | --- |
| Enter `/transactions` | Load transactions, wallets and visible groups from their authenticated APIs. Never replace an error with mock rows. |
| Select wallet scope | Recompute the real ledger view locally without another API call; `Tổng cộng` clears the wallet predicate. |
| Select a goal wallet | Show real savings summary/progress and monthly history; do not show ordinary weekly cashflow cards or an All-time period control. |
| Select a basic wallet | Keep the aggregate layout, including Week/Custom control and cashflow card; filter both totals and day rows to the selected wallet and current period. |
| Select an aggregate/basic/credit scope or step week | Start/restart at the current account-local Monday–Sunday week; step through prior weeks. Never navigate into a future week. A timezone change recalculates the week from the account-local current date. |
| Select `Tùy chọn` | Open the existing date-range dialog. A valid range filters both summary and rows; cancel keeps the previous period choice and dates. |
| Return from custom range to week | Show the current account-local week. There is no All-time period tab for ordinary scopes. |
| Open an expense editor | Load jar choices for the transaction's account-local month; show only active monthly configurations, plus the existing historical assignment when editing. |
| Activate global add | Open a clean ordinary editor with `expense`, current local date/time, report inclusion enabled, and the first available wallet. It does not offer transfer mode. |
| Choose “Chuyển tiền đến ví khác” from the three-dot options menu | Open a transfer-only sheet with source/destination wallet selects; submit calls the atomic transfer endpoint and refreshes both wallet balances and the ledger. |
| Hold global add for 500ms | Open [AI entry chat](../assistant/README.md) with persistent prefilled review proposals. Release does not also open the manual editor; moving/cancelling cancels hold. Manual create offers keyboard-accessible “Nhập bằng AI”. |
| Change wallet | Keep the selected group only if it applies to the new wallet; otherwise clear the group. |
| Change type | Keep the selected group only if its kind matches; otherwise clear the group. |
| Change transaction month | Reload that month's jar configurations; keep historical assignment visible for review, but clear or choose an active jar before moving a linked transaction to a different instant/month. |
| Change type to income / select transfer-out category | Clear and hide optional jar selection; backend remains the final eligibility guard. |
| Submit valid editor | Convert local datetime to RFC3339 UTC, include optional `jar_id`, call create/update, close on success, then refresh transactions and wallets. |
| Submit invalid editor | Keep entered data and show field/form feedback; do not call the API. |
| Activate a transaction | Open the editor populated from that API row. |
| Open the group selector | Search the category tree, choose an applicable root/child group, or clear the selection; the picker composes the same `BaseCategoryTree` selection mode used by group management. Non-applicable parents may remain visible only as hierarchy context. |
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
- No matching groups: category remains optional; the transaction can still be saved.

## Validation and source of truth

- Backend owns authorization and final validation. Client validation is recovery guidance, not a security boundary.
- `wallet_id` must belong to the authenticated owner. Basic income/expense accepts `basic` and `goal`; `credit` is excluded until its separate debt/payment ledger is implemented.
- `type` is `income` or `expense`; `amount` is an integer greater than zero.
- Transfer mode requires two different non-credit owner wallets and a positive integer amount; the API creates the paired rows and sets report inclusion false.
- Category is optional. It must be visible, match the transaction type, and either have no wallet restriction or include the selected wallet ID.
- `occurred_at` is sent as an RFC3339 timestamp. Note is trimmed and omitted when empty.
- `included_in_reports` defaults to true.
- `jar_id` is optional and may reference one active owner/month configuration only for ordinary expenses; income and transfer-out rows cannot be assigned. Clearing the selection sends null.
- Wallet current balance is derived from ledger rows: opening balance plus income minus expense. Editing/deleting a row changes the derived balance; it does not mutate opening balance.

APIs: `GET/POST /api/v1/transactions`, `POST /api/v1/transactions/transfer`, `PATCH/DELETE /api/v1/transactions/{id}`, `GET /api/v1/wallets`, `GET /api/v1/categories`, and `GET /api/v1/jars?month=YYYY-MM` for active jar options.

## Copy and formatting

Use Vietnamese product copy. Money uses signed integer VND via the shared formatter. Transaction title prefers category name and falls back to `Khoản thu`/`Khoản chi`; metadata shows wallet then note. Date headings use `Hôm nay`, `Hôm qua`, otherwise a Vietnamese calendar date.

## Proof and residual gaps

Automated proof covers backend owner scope, positive amount/type validation, wallet ownership, category kind and wallet applicability, report flag, optional jar-link validation, and frontend design/build guardrails. `npm run test:transaction-jars` renders the real form and verifies jar selection for ordinary expenses while hiding it for income/transfer-out. Manual UAT still needs to cover create/edit/delete, jar assignment/clear, and refreshed wallet totals.

2026-09-24 weekly-ledger proof: `test:calendar` covers Monday–Sunday and year boundaries in account-local time; `test:transactions` covers wallet + week predicate at UTC boundary; the SSR period-selector fixture covers tab labels/range/future-week disable. Chrome on the local preview showed All → 21–27/9 → 14–20/9 with cashflow/list updates, empty state, next-week disable, and custom-dialog cancel preserving week. Build/design checks pass. Owner mobile visual acceptance remains open. [User guide](../../../guides/transactions.md).

Transfer proof now includes `npm run test:transfer:e2e` (real Chromium UI/state) and `npm run test:transfer:integrated` (Vite proxy/API/PostgreSQL state). Residual gaps remain separate tickets: transfer-pair edit/delete, attachments/OCR, debt/credit ledger, advanced search/filter/pagination, and owner UAT sign-off.
