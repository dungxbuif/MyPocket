# Base component inventory and Atomic Design map

[Gateway](README.md) · [Shared contracts](system/BASE_COMPONENTS.md) · [Organisms](organisms/README.md) · [Templates](templates/README.md) · [Pages](pages/README.md) · [References](references/README.md). Status: implemented = reusable base exists, partial = preview/composition lacks complete behavior, planned = must create base before page use. A working mock is not completed product behavior.

| # | Component target | Current implementation | Status / next behavior |
| --- | --- | --- | --- |
| 1 | Top bar | PageBackHeader, AppHeader | implemented; screen specifies actions |
| 2 | Bottom navigation | BottomNavigation + BaseNavigationItem/BaseFab | implemented; central action depends on transaction slice |
| 3 | Segmented control | SegmentedControl | implemented; click, arrow, Home/End |
| 4 | Scope selector | LedgerPeriodSelector + BaseSelect | implemented for transaction ledger; reuse remains page-owned |
| 5 | Amount input | QuickAddSheet preview | partial; [keypad](atoms/amount-keypad/README.md) |
| 6 | Selector row | FormSelectorRow, InlineControlRow | partial; explicit callbacks needed per selector |
| 7 | Date selector | BaseCalendar + DateField | implemented for configured date/range flows |
| 8 | Switch row | BaseSwitch | implemented; keep switch behavior in the atom |
| 9 | Primary action | BaseButton | implemented |
| 10 | Numeric keypad | QuickAddSheet preview | partial, arithmetic not complete |
| 11 | Wallet list | WalletCard, wallet-management composition | implemented visual; API wallet management separately tested |
| 12 | Transaction list | TransactionItem | implemented visual; source preview data |
| 13 | Category tree | BaseCategoryTree | implemented for navigation and selection; fixed expanded, named nested/line spacing |
| 14 | Date group header | TransactionsPanel + Text | partial; date-header molecule needed for real ledger |
| 15 | Total balance | AppHeader | implemented for authenticated account summary; page data states remain explicit |
| 16 | Budget progress | BudgetProgressItem + Progress | partial; no time marker/warning policy |
| 17 | Budget gauge | BudgetGauge | partial; shared SVG arc exists, real budget data pending |
| 18 | Stat metric | MetricBox | implemented visual |
| 19 | Notice banner | StatusMessage | partial; status/error exists, promo intent requires named variant |
| 20 | Bar comparison | BaseBarChart | partial; visual base extracted, reporting data/interaction pending |
| 21 | Donut category chart | BaseDonutChart | partial; visual base extracted, interactive chart behavior pending |
| 22 | Trend chart | absent | planned |
| 23 | Action/bottom sheet | BaseBottomSheet | implemented; focus, close, scroll lock |
| 24 | Calendar picker | absent | planned |
| 25 | Context menu/popover | absent | planned |
| 26 | Wallet selection sheet | ApplicableWalletsCard + checkbox only | partial; sheet flow needs contract |
| 27 | Percentage/status badge | Text/StatusMessage only | partial; visual badge needs named base |
| 28 | Category icon badge | IconBadge + categoryPresentation | implemented; shared tone catalog |

Additional foundation atoms: SurfaceCard, Text, Heading, FormField, BaseTextInput, BaseSelect, BaseCheckbox, BaseLink, Divider, IconButton, Progress, Chip.
No inventory entry claims all 28 are complete. Existing retained preview components must meet the same automated checks as mounted code.

New shared bases: BaseTextArea and the button variant of BaseFileUpload support the AI composer; AssistantComposer and AssistantResultCard provide the Money Lover inspired input/result layout while leaving submission and approval behavior to the screen. CategoryTreeSelector composes BaseCategoryTree selection mode for transaction, budget and proposal pickers. These use the existing theme tokens and transaction field bases.

Page contracts: [Overview](pages/overview/README.md) · [Account groups](pages/account-groups/README.md) · [Transactions](pages/transactions/README.md) · [Budgets](pages/budgets/README.md) · [Jars](pages/jars/README.md) · [Month detail](pages/month-detail/README.md) · [Assistant](pages/assistant/README.md) · [Finance Assistant](pages/finance-assistant/README.md).
