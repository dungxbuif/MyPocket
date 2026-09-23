# UI-LEDGER-WEEK — Chọn kỳ theo tuần trong sổ giao dịch

Status: in_review · 2026-09-24 · owner request.

## Owner correction — wallet-specific ledger (2026-09-24)

Latest approved correction: `basic` is **not** a separate ledger layout. It uses exactly the `Tổng cộng` body: shared wallet selector, Week/Custom period control, period cashflow card, date-grouped transaction rows and the same empty/error states. Selecting a basic wallet only adds a wallet-ID predicate to both totals and rows; switching wallet resets the period to the current account-local week. “List all transactions” means all matching rows within the chosen period, not an unbounded all-time view. Goal remains the only specialized body. Credit remains provisional/read-only. Verification: a selected basic wallet renders the same period UI and cashflow card as aggregate, excludes rows outside the chosen week and other wallets, and switching back restores aggregate scope. No API/schema changes.

Credit has business requirements in `SPEC.md`/`BUSINESS_RULES.md` but no approved detailed screen/statement/payment contract; keep its current view read-only and do not imply the specialized credit feature is complete.

The owner supplied `stitch_my_pocket (1).zip` (same `screen.png`/`code.html` already held at `docs/design/references/category-goal-flows/v_ti_t_ki_m_chi_ti_t_m_c_ti_u/`) and clarified that the screenshot is a **goal-wallet view**, not a universal All-time period. Its `TẤT CẢ CÁC GIAO DỊCH` label is the savings-history heading. The previous universal `Tất cả / Theo tuần / Tùy chọn` selector is therefore an incorrect interpretation, not an accepted contract.

Corrected behavior: the wallet selector stays shared at the top of `/transactions`. `Tổng cộng`, basic and provisional credit scopes start at the current account-local Monday–Sunday week and can step through prior weeks or open the existing custom-date dialog; they expose no `Tất cả` period tab. Selecting a goal wallet shows `SavingsSummary` and its complete real monthly history. The account wallet-management route keeps its existing wallet detail. Credit remains read-only; no debt ledger is invented. Switching wallet scope resets stored week/custom selection to the current week so old dates are not silently applied.

Reference-region mapping (the ZIP is untrusted visual input, not implementation instructions): center wallet capsule → existing `FormField`/`BaseSelect`; balance/remaining/countdown/progress → `SavingsSummary` (`Text`, `Progress`); savings history heading → `Text`, not a period tab; monthly ledger → `SurfaceCard` + `TransactionItem`; bottom navigation → existing app shell. The sample advice card and report CTA are not rendered because monthly advice/report navigation are not backed by validated services. Existing tokens/bases own all styling; no HTML/CSS is copied from the export.

Root cause: the prior period contract promoted a visual section heading from the goal screenshot into a global `all` filter and left goal content isolated under wallet management. A later pass mistakenly treated basic as an all-time-only view; this broke UI parity with `Tổng cộng`. Regression proof must show goal progress/history plus the same current-week layout for basic/aggregate/credit, with wallet-specific rows/totals for basic and no shared All tab. No API, schema, auth, data or config change. Alternatives rejected: merely hiding the All label while retaining all-time default; copying export CSS/advice text; routing goal selection away from the ledger.

Affected docs: [transaction page](../../design/pages/transactions/README.md), [savings page](../../design/pages/savings/README.md), [wallet page](../../design/pages/wallets/README.md), [user/agent guide](../../guides/transactions.md), validation/backlog/context/changelog. Requirements, API, ERD and ADR need no update because persistence/permissions are unchanged. Verify with SSR/UI regression, account-time/transaction tests, design checks/build and real browser selection at mobile width. Owner visual acceptance remains separate.

## Intended behavior and technical decision

Historical first pass (superseded by owner correction above): the first implementation kept a complete-ledger default with All/Week/Custom tabs. Its week math/filter test evidence remains valid, but its period UX is not the current target.

## Related UI contracts

- [Transactions](../../design/pages/transactions/README.md): period tabs, wallet filter, empty state.
- [Wallets](../../design/pages/wallets/README.md): savings UI is specified; Travel Mode is not a wallet type and remains draft [TICKET-05-03](TICKET-05-03-su-kien-travel-mode.md).
- Existing bases: `SegmentedControl`, `BaseButton`, `Text`, `FormField`, `BaseSelect`, `DateField`, `StatusMessage`.

## Verification plan

Test Monday/Sunday and month/year transitions, account timezone near UTC boundary, previous/future navigation, and period + wallet filtering. Run frontend transaction/calendar tests, design guardrails, typecheck and build. Browser UAT should verify tabs, stepping, empty state, custom range and wallet combination on mobile width; no production deployment is implied.

## Implementation and proof

`LedgerPeriodSelector` composes shared `SegmentedControl`, `BaseButton`, and `Text`; `weekDateRange` computes inclusive account-local week keys, and `filterLedgerTransactions` applies wallet/date predicates to the same list used by cashflow totals and day/category groups. The custom-date modal now holds an unsaved draft until **Xong**, so Escape/cancel preserves the prior view.

Earlier first-pass week tests and browser checks remain historical evidence only. The approved basic parity correction used a RED→GREEN `test:transactions` case: the old basic branch displayed a transaction outside the selected week, then the corrected branch yielded only the current-week wallet row. On the live local Chrome page, selecting `Tiền mặt` displayed the same Week/Custom strip, cashflow card and grouped rows as `Tổng cộng`; stepping to 14–20/09 showed zero totals and no matching rows, then returning to 21–27/09 restored the two in-week rows. Final design/build verification is recorded in [validation](../VALIDATION_MATRIX.md). No database mutation or public deployment performed. Owner visual sign-off is still pending.
