# Base component inventory and Atomic Design map

[Gateway](README.md) · [Shared contracts](system/BASE_COMPONENTS.md) · [Organisms](organisms/README.md) · [Templates](templates/README.md) · [Pages](pages/README.md) · [References](references/README.md). Status: implemented = reusable base exists, partial = preview/composition lacks complete behavior, planned = must create base before page use. A working mock is not completed product behavior.

| # | Component target | Current implementation | Status / next behavior |
| --- | --- | --- | --- |
| 1 | Top bar | PageBackHeader, AppHeader | implemented; screen specifies actions |
| 2 | Bottom navigation | BottomNavigation + BaseNavigationItem/BaseFab | implemented; central action depends on transaction slice |
| 3 | Segmented control | SegmentedControl | implemented; click, arrow, Home/End |
| 4 | Scope selector | WalletScopeSelector + WalletSelectionList + BaseBottomSheet + LedgerPeriodSelector | implemented for transaction ledger; wallet scope opens the shared picker, period remains page-owned |
| 5 | Amount input | QuickAddSheet / `MPAmountField` + `MPAmountKeypad` | implemented; integer-safe calculator/keypad is a shared native base; [keypad](atoms/amount-keypad/README.md) |
| 6 | Selector row | FormSelectorRow, InlineControlRow / `MPFormRow`, `MPOptionSelector` | implemented; controlled callbacks/identity IDs are supplied by each consumer |
| 7 | Date selector | BaseCalendar + DateField + DateTimeField | implemented; date-only and date+time flows use the shared account-timezone sheet |
| 8 | Switch row | BaseSwitch | implemented; keep switch behavior in the atom |
| 9 | Primary action | BaseButton | implemented |
| 10 | Numeric keypad | `MPAmountKeypad` | implemented; integer-safe arithmetic is covered by the native contract test |
| 11 | Wallet list | WalletCard, wallet-management composition | implemented visual; API wallet management separately tested |
| 12 | Transaction list | TransactionItem | implemented visual; source preview data |
| 13 | Category tree | BaseCategoryTree / `MPCategoryTree` / `MPCategoryPicker` | implemented for web and native management/selection; native kind/wallet/search filtering, full-row editor navigation and wallet-scope subtitle; fixed expanded, stable depth-first tree |
| 14 | Date group header | TransactionsPanel + Text | implemented; account-timezone day number, weekday, month/year and signed group total |
| 15 | Total balance | AppHeader | implemented for authenticated account summary; page data states remain explicit |
| 16 | Budget progress | BudgetProgressItem + Progress | implemented; clamped progress, over-budget tone, timing and daily guidance are mounted in budgets |
| 17 | Budget gauge | BudgetGauge / `MPBudgetGauge` | implemented; real spent/limit summary data, clamped arc and accessible value on web/native |
| 18 | Stat metric | MetricBox | implemented visual |
| 19 | Notice banner | StatusMessage / `MPStatusMessage` | implemented for loading, empty, info and error semantics; promotional notices remain out of scope until a real product contract exists |
| 20 | Bar comparison | BaseBarChart / `MPChart` | implemented; mounted reports use owner-scoped summary data; drilldown remains consumer-owned |
| 21 | Donut / comparison ring | BaseDonutChart / `MPDonutChart` | implemented for the Money Insider period-comparison ring; category drilldown/tooltip remains intentionally unsupported |
| 22 | Trend chart | `BaseTrendChart` / `MPTrendChart` | implemented; web and native Reports use account-timezone daily ledger points with an empty state |
| 23 | Action/bottom sheet | BaseBottomSheet | implemented; focus, close, scroll lock |
| 24 | Calendar picker | `MPDateCalendarSheet` | implemented; `MPDateStepper`, `MPDateField` and `MPDateTimeField` use the shared Stitch-style Monday-first grid with month navigation and confirm/cancel |
| 25 | Context menu/popover | `MPOverflowMenu` / `MPMenuSelector` | implemented for transaction and wallet secondary actions; screen-specific actions remain consumer-owned |
| 26 | Wallet selection sheet | WalletSelectionList + WalletScopeSelector | implemented for transaction scope and wallet management; add/edit routes remain page-owned |
| 27 | Percentage/status badge | StatusBadge / `MPStatusBadge` | implemented; named semantic badge owns tone, compact geometry and status accessibility |
| 28 | Category icon badge | IconBadge + categoryPresentation | implemented; shared tone catalog |

Additional foundation atoms: SurfaceCard, Text, Heading, FormField, BaseTextInput, BaseSelect, BaseCheckbox, BaseLink, Divider, IconButton, Progress, Chip.
No inventory entry claims all 28 are complete. Existing retained preview components must meet the same automated checks as mounted code.

New shared bases: BaseTextArea and the button variant of BaseFileUpload support the AI composer; AssistantComposer and AssistantResultCard provide the Money Lover inspired input/result layout while leaving submission and approval behavior to the screen. CategoryTreeSelector composes BaseCategoryTree selection mode for transaction, budget and proposal pickers. These use the existing theme tokens and transaction field bases.

Page contracts: [Overview](pages/overview/README.md) · [Account groups](pages/account-groups/README.md) · [Transactions](pages/transactions/README.md) · [Budgets](pages/budgets/README.md) · [Jars](pages/jars/README.md) · [Month detail](pages/month-detail/README.md) · [Assistant](pages/assistant/README.md) · [Finance Assistant](pages/finance-assistant/README.md).
