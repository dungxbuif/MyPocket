# Current UI composition and behavior

Scope of UI-BASE-01: normalize shared visual primitives and preserve current route behavior. Product acceptance for backend flows belongs to product tickets.

| Area | Current composition | Behavior / known limits |
| --- | --- | --- |
| Auth gate | Text, StatusMessage, BaseButton | Checking/login/error; OAuth behavior unchanged; primary uses shared brand |
| Shell | MobileAppShell, AppHeader, BottomNavigation with BaseFab/BaseNavigationItem | sticky header, mask toggle, 4 routes; central add callback remains inactive pending transaction slice |
| Overview | SurfaceCard, SectionTitle, WalletCard, TransactionItem, StatusMessage | Authenticated wallet/transaction data plus live account-local month summary; month card links to detail and jars. Retry/masking states are real. |
| Transactions | SurfaceCard, Text, IconButton, TransactionItem, TransactionFields, BaseBottomSheet | Authenticated grouped ledger with search/filter/create/edit/delete, account-local date conversion and optional month-configured jar assignment. |
| Budgets | SurfaceCard, Text, BudgetGauge, MetricBox, BudgetProgressItem, BudgetEditorForm, BaseCalendar | Authenticated budget CRUD/progress with date-only account-calendar bounds, category tree and wallet scope; full owner CRUD UAT remains pending. |
| Account | ProfileHeroCard, BaseButton, SurfaceCard, AccountMenuRow, FormField, BaseSelect, StatusMessage | Profile/logout, wallet/group routes and authenticated account timezone read/update. |
| Wallet management | BaseLink, BaseButton, StatusMessage, SurfaceCard, WalletEditorForm | Fetch/create/edit/delete; delete native confirm; form fields disabled while saving; list/empty/error states |
| Group management | PageBackHeader, SegmentedControl, BaseCategoryTree | [Detailed contract](account-groups/README.md) |
| Reports/QuickAdd | QuickAddSheet, AssistantComposer, AssistantResultCard, base controls/cards/text | QuickAdd and one-shot AI review are mounted/API-backed. Broader reports/charts and financial Q&A remain pending; monthly detail is the current live numeric report. |

Shared QA: app/tests/design.html verifies atoms/tree/gauge/sheet without production data. Product CRUD UAT is not rerun by this fixture; this task does not sign off ledger/budget/business workflows.
