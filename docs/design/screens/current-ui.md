# Current UI composition and behavior

Scope of UI-BASE-01: normalize shared visual primitives and preserve current route behavior. Product acceptance for backend flows belongs to product tickets.

| Area | Current composition | Behavior / known limits |
| --- | --- | --- |
| Auth gate | Text, StatusMessage, BaseButton | Checking/login/error; OAuth behavior unchanged; primary uses shared brand |
| Shell | MobileAppShell, AppHeader, BottomNavigation with BaseFab/BaseNavigationItem, FeedbackFloatingBubble | sticky header, mask toggle, four primary routes; the floating feedback bubble is available on every authenticated screen and captures the mounted `.mypocket-app-root` view without the feedback overlay |
| Overview | SurfaceCard, SectionTitle, WalletCard, TransactionItem, StatusMessage | Authenticated wallet/transaction data plus live account-local month summary; month card links to detail and jars. Retry/masking states are real. |
| Transactions | SurfaceCard, Text, IconButton, TransactionItem, TransactionFields, BaseBottomSheet | Authenticated grouped ledger with search/filter/create/edit/delete, account-local date conversion and optional month-configured jar assignment. |
| Budgets | SurfaceCard, Text, BudgetGauge, MetricBox, BudgetProgressItem, BudgetEditorForm, BaseCalendar | Authenticated budget CRUD/progress with date-only account-calendar bounds, category tree and wallet scope; full owner CRUD UAT remains pending. |
| Account | ProfileHeroCard, BaseButton, SurfaceCard, AccountMenuRow, FormField, BaseSelect, StatusMessage | Profile/logout, wallet/group routes and authenticated account timezone read/update. |
| Wallet management | BaseLink, BaseButton, StatusMessage, SurfaceCard, WalletEditorForm | Fetch/create/edit/delete; delete native confirm; form fields disabled while saving; list/empty/error states |
| Group management | PageBackHeader, SegmentedControl, BaseCategoryTree | [Detailed contract](account-groups/README.md) |
| Reports/QuickAdd | QuickAddSheet, AssistantComposer, AssistantResultCard, base controls/cards/text | QuickAdd and one-shot AI review are mounted/API-backed. Broader reports/charts and financial Q&A remain pending; monthly detail is the current live numeric report. Transfer creation is opened from the transactions three-dot menu. |

## Feedback bubble

`FeedbackFloatingBubble` is mounted by `MobileAppShell`, so it follows the authenticated app shell rather than a single route. Opening the bubble starts a best-effort `html2canvas` capture of the `mypocket-app-root` ref. Elements marked `data-feedback-overlay="true"` are excluded, and the user may remove the preview before submitting. The form uses the shared base modal, inputs, select, textarea, button, icon button, and status components; the optional PNG is sent through the feedback multipart API.

Shared QA: app/tests/design.html verifies atoms/tree/gauge/sheet without production data. Product CRUD UAT is not rerun by this fixture; this task does not sign off ledger/budget/business workflows.
