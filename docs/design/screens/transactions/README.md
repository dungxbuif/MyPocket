# Transactions — list and basic income/expense editor

Status: implementation contract for [TICKET-02-01](../../../work/tickets/TICKET-02-01-ghi-thu-chi.md) and its [detail design](../../../work/tickets/TICKET-02-01-DETAIL_DESIGN.md). Route: `/transactions`; the global add action opens the same editor from every primary tab.

## Composition

2026-09-20: editor uses grouped SurfaceCards for wallet/large amount/category/note and date, retaining the shared inputs and form-sheet shell. Goal category selection is restricted to real savings catalog keys and backend enforces the same rule; category remains optional. Unsupported old goal categories remain visible in history but must be cleared/reselected before editing. Report flag for external entries defaults true. This is not the paired-wallet transfer flow.

| Region | Base implementation | Contract |
| --- | --- | --- |
| Search/filter summary | `Chip`, `IconButton` | Passive until the search/filter ticket is implemented; it must not imply working filtering. |
| Date group | `SurfaceCard`, `Text` | Groups real API rows by the user's local calendar date and displays the signed group total. |
| Transaction row | `TransactionItem` | Category icon/title, wallet and optional note, signed VND amount; activation opens edit mode. |
| Editor shell | `BaseBottomSheet` | Modal focus trap, Escape/backdrop cancel, and return focus follow the shared sheet contract. |
| Type, amount and metadata | `SegmentedControl`, `FormField`, `BaseTextInput`, `BaseSelect`, `BaseCheckbox` | Only `expense` and `income` are enabled in this slice. Amount is a positive integer VND value. |
| Feedback/actions | `StatusMessage`, `BaseButton` | Save has loading/disabled behavior; delete exists only in edit mode and requires confirmation. |

No screen-local color, shape, input, button or card styling is allowed. Transfer, debt, receipt/OCR and advanced filter controls stay out of this screen until their own approved slices exist.

## Events and effects

| Event | State/effect |
| --- | --- |
| Enter `/transactions` | Load transactions, wallets and visible groups from their authenticated APIs. Never replace an error with mock rows. |
| Activate global add | Open a clean editor with `expense`, current local date/time, report inclusion enabled, and the first available wallet. |
| Change wallet | Keep the selected group only if it applies to the new wallet; otherwise clear the group. |
| Change type | Keep the selected group only if its kind matches; otherwise clear the group. |
| Submit valid editor | Convert local datetime to RFC3339 UTC, call create/update, close on success, then refresh transactions and wallets. |
| Submit invalid editor | Keep entered data and show field/form feedback; do not call the API. |
| Activate a transaction | Open the editor populated from that API row. |
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
- Category is optional. It must be visible, match the transaction type, and either have no wallet restriction or include the selected wallet ID.
- `occurred_at` is sent as an RFC3339 timestamp. Note is trimmed and omitted when empty.
- `included_in_reports` defaults to true.
- Wallet current balance is derived from ledger rows: opening balance plus income minus expense. Editing/deleting a row changes the derived balance; it does not mutate opening balance.

APIs: `GET/POST /api/v1/transactions`, `PATCH/DELETE /api/v1/transactions/{id}`, `GET /api/v1/wallets`, `GET /api/v1/categories`.

## Copy and formatting

Use Vietnamese product copy. Money uses signed integer VND via the shared formatter. Transaction title prefers category name and falls back to `Khoản thu`/`Khoản chi`; metadata shows wallet then note. Date headings use `Hôm nay`, `Hôm qua`, otherwise a Vietnamese calendar date.

## Proof and residual gaps

Automated proof must cover backend owner scope, positive amount/type validation, wallet ownership, category kind and wallet applicability, report flag, and frontend design/build guardrails. Manual UAT covers create, edit and delete for one income and one expense and verifies wallet balances refresh in both transaction and wallet views.

Residual gaps remain separate tickets: attachments/OCR, transfer/adjustment, debt/credit ledger, advanced search/filter/pagination, and owner UAT sign-off.
