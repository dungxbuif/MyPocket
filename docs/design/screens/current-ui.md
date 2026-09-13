# Current UI composition and behavior

Scope of UI-BASE-01: normalize shared visual primitives and preserve current route behavior. Product acceptance for backend flows belongs to product tickets.

| Area | Current composition | Behavior / known limits |
| --- | --- | --- |
| Auth gate | Text, StatusMessage, BaseButton | Checking/login/error; OAuth behavior unchanged; primary uses shared brand |
| Shell | MobileAppShell, AppHeader, BottomNavigation with BaseFab/BaseNavigationItem | sticky header, mask toggle, 4 routes; central add callback remains inactive pending transaction slice |
| Overview | SurfaceCard, SectionTitle, WalletCard, TransactionItem | Current preview arrays; not real ledger proof |
| Transactions | SurfaceCard, Text, IconButton, TransactionItem | Mock daily groups/filter display; real search/filter/create slice pending |
| Budgets | SurfaceCard, Text, BudgetGauge, MetricBox, BudgetProgressItem | Mock budget data and preview forecast; shared arc/progress now used |
| Account | ProfileHeroCard, BaseButton, SurfaceCard, AccountMenuRow | Profile + logout + routes to wallets/groups |
| Wallet management | BaseLink, BaseButton, StatusMessage, SurfaceCard, WalletEditorForm | Fetch/create/edit/delete; delete native confirm; form fields disabled while saving; list/empty/error states |
| Group management | PageBackHeader, SegmentedControl, BaseCategoryTree | [Detailed contract](account-groups/README.md) |
| Reports/QuickAdd | Retained source using base controls/cards/text | Not mounted; complete chart/keypad behaviors remain pending |

Shared QA: app/tests/design.html verifies atoms/tree/gauge/sheet without production data. Product CRUD UAT is not rerun by this fixture; this task does not sign off ledger/budget/business workflows.
