# Tổng quan tháng

Route `/`; implements the non-AI part of [CORE-03](../../../work/tickets/CORE-03-TIME-JARS-MONTH-DETAIL_DESIGN.md) and [TICKET-07-04](../../../work/tickets/TICKET-07-04-tong-ket-thang-ai.md). Status: implementation in progress.

## Composition

| Region | Base | Behavior |
| --- | --- | --- |
| Current month summary | `SurfaceCard`, `SectionTitle`, `Text`, `BaseButton`, `StatusMessage` | Shows the account-local month, included income/expense totals, completion label, and the user's saved note when non-empty. It is a navigation summary, not an AI narrative or frozen close record. |
| Jar shortcut | `BaseButton` inside the month card | Opens `/jars` for the same account-local month. |
| Wallets and recent transactions | Existing `SurfaceCard`, `SectionTitle`, `WalletCard`, `TransactionItem`, `StatusMessage` | Preserve the live API behavior and existing borderless empty states. |

The screen composes only existing bases. `className` values may position content; colors, typography, card shape, and controls remain base-owned.

## Events and effects

| Event | Effect |
| --- | --- |
| Enter overview | Read the month label using the saved account timezone, then load `GET /api/v1/months/{YYYY-MM}` alongside existing overview data. |
| Activate month summary | Navigate to `/months/{YYYY-MM}`. |
| Activate jar shortcut | Navigate to `/jars?month={YYYY-MM}`. |
| Ledger refresh | Reload live totals and recent rows; the backend recalculates from persisted transactions. |
| Month-summary request fails | Keep wallet/transaction overview usable; show a retryable status in the month card, never zero-valued mock totals. |

## States and validation

- Loading: existing overview loading status; no placeholder amounts.
- Ready: shows the account-local month label, derived totals and `Đang diễn ra` or `Đã hoàn tất` according to the server response.
- Empty: real zero totals are allowed only after a successful API response; wallet and transaction empty states retain their current plain variant.
- Error: retry the month API independently from wallet/transaction data.
- Privacy mask: hides all currency values, including month totals.
- The month label and `is_complete` are server-derived; clients do not close months or calculate totals independently.

## Copy and source of truth

Use Vietnamese labels `Tổng kết tháng`, `Thu`, `Chi`, `Đang diễn ra`, `Đã hoàn tất`, and `Xem chi tiết`. Money uses the shared integer-VND formatter. The note is previewed verbatim and is edited only on the detail route. Completion never blocks edits or backdated transactions.

API: `GET /api/v1/months/{YYYY-MM}`. No AI text, fixture data, or manual close action is shown.

## Proof and gaps

Implementation update: the month card is connected to the authenticated month API, formats the current month using the saved account timezone, masks currency with the privacy toggle, retries report errors independently, and links to month detail/Hũ. No fixture totals are rendered.

Automated proof: `npm run check:design`, `npm run test:design`, and `npm run build` pass. `npm run test:calendar` covers account-local month selection. Runtime proof must still verify totals after a persisted transaction change and note independence. Owner browser UAT is pending; AI narrative and other report modules are outside this screen.
