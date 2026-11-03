# Tổng quan tháng

Route `/`; implements the non-AI part of [CORE-03](../../../work/tickets/CORE-03-TIME-JARS-MONTH-DETAIL_DESIGN.md) and [TICKET-07-04](../../../work/tickets/TICKET-07-04-tong-ket-thang-ai.md). Status: implementation in progress.

## Native daily-money reference composition — 2026-10-26

The native Overview keeps the existing owner-scoped load, cache fallback,
privacy mask and navigation behavior, but presents the daily-money scan through
the shared Money Lover-derived grammar. It does not add a second dashboard data
model or screen-owned visual constants.

| Region | Native composition | Preserved behavior |
| --- | --- | --- |
| Balance header | `MPBalanceHeader` | Included-wallet total, mask toggle and notifications remain owner-scoped. |
| Wallets | `MPSurfaceCard(.grouped)`, `MPGroupedRow`, `MPInsetDivider`, `MPSectionHeader`, `MPWalletRow` | Up to three real wallets; description falls back to the localized wallet type; every visible balance respects masking. |
| Money Insider and period report | Existing `MPMoneyInsiderCard`, `MPOverviewReportCard` | Server month summary and account-local week calculations remain unchanged. |
| Month summary | `MPSurfaceCard(.grouped)`, `MPGroupedRow`, `MPInsetDivider`, `MPMetricGrid` | Independent loading/error/retry, note preview, month-detail and jar routes are preserved; all money remains maskable. |
| Recent transactions | `MPSurfaceCard(.grouped)`, `MPGroupedRow`, `MPInsetDivider`, `MPTransactionRow` | Four newest loaded rows keep category/note/date/signed-amount semantics and account timezone. |

The page canvas, section spacing, grouped insets, separators, row height,
typography, colors and corner geometry come only from shared tokens/bases.
Loading, empty, stale-cache and retry states stay explicit. This migration does
not change API requests, snapshot contents, totals, report inclusion, routes or
offline behavior.

Implementation checkpoint: wallet, month-summary and recent-transaction regions
now use the shared grouped surface/row/divider composition. Overview money uses
one tested mask-aware presentation helper, while the existing `MPBalanceHeader`,
`MPMetricGrid` and `MPTransactionRow` retain their accessibility masking. The
focused daily-money unit slice passed 10/10 before final review; the post-review
complete signed native unit target passed 409/409 with no failures, skips or
runtime warnings. The focused shell/editor/Overview privacy UI slice passed 4/4,
and the Release simulator build passed with warnings treated as errors. Native
base/project guards and `git diff --check` pass. Populated small-phone,
large-text, dark-mode and physical-device visual acceptance remain release/UAT
gates and are not claimed by this implementation checkpoint.

## Composition

| Region | Base | Behavior |
| --- | --- | --- |
| Current month summary | `SurfaceCard`, `SectionTitle`, `Text`, `BaseButton`, `StatusMessage` | Shows the account-local month, included income/expense totals, completion label, and the user's saved note when non-empty. It is a navigation summary, not an AI narrative or frozen close record. |
| Week/month report card | `OverviewReportCard`, `SegmentedControl`, `BaseBarChart` | The weekly view uses the account-local Monday–Sunday week and sums only report-included income/expense transactions. Transfers, balance adjustments, excluded rows, invalid timestamps, and rows outside the week do not contribute. The monthly view continues to use the server month summary. |
| Jar shortcut | `BaseButton` inside the month card | Opens `/jars` for the same account-local month. |
| Wallets and recent transactions | Existing `SurfaceCard`, `SectionTitle`, `WalletCard`, `TransactionItem`, `StatusMessage` | Preserve the live API behavior and existing borderless empty states. Recent transaction metadata includes the account-local date and time before wallet/note context. |

The screen composes only existing bases. `className` values may position content; colors, typography, card shape, and controls remain base-owned.

## Events and effects

| Event | Effect |
| --- | --- |
| Enter overview | Read the month label using the saved account timezone, then load `GET /api/v1/months/{YYYY-MM}` alongside existing overview data. |
| Select weekly report | Derive the current Monday–Sunday date range from the saved account timezone; apply the same report-inclusion and income/expense-only rules on web and iOS. |
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
- The month label and `is_complete` are server-derived; clients do not close months or recalculate monthly totals independently. Weekly card totals are derived from the transaction feed under the rule above.

## Copy and source of truth

Use Vietnamese labels `Tổng kết tháng`, `Thu`, `Chi`, `Đang diễn ra`, `Đã hoàn tất`, and `Xem chi tiết`. Money uses the shared integer-VND formatter. The note is previewed verbatim and is edited only on the detail route. Completion never blocks edits or backdated transactions.

Recent transaction timestamps use the saved account timezone and the stable
`dd/MM/yyyy HH:mm` display format; they must not be derived from the browser's
timezone.

API: `GET /api/v1/months/{YYYY-MM}`. No AI text, fixture data, or manual close action is shown.

## Proof and gaps

Implementation update: the month card is connected to the authenticated month API, formats the current month using the saved account timezone, masks currency with the privacy toggle, retries report errors independently, and links to month detail/Hũ. No fixture totals are rendered.

Automated proof: `npm run check:design`, `npm run test:design`, and `npm run build` pass. `npm run test:calendar` covers account-local month selection. Weekly report totals are additionally covered by deterministic web and native tests for timezone boundaries and excluded transaction kinds. Runtime proof must still verify totals after a persisted transaction change and note independence. Owner browser UAT is pending; AI narrative and other report modules are outside this screen.

## Hũ trong Tổng quan — planned 2026-11-03

Thẻ tháng thêm hai `MPGroupedRow`: "Hũ: còn có thể chi tổng X" và "Chưa phân bổ tháng này: Y", dữ liệu từ `GET /jars?month=`; nút "Xem hũ tháng" giữ nguyên. Xem [JARS-AS-WALLET-CARDS](../../../work/tickets/JARS-AS-WALLET-CARDS-20261103-DETAIL_DESIGN.md).
