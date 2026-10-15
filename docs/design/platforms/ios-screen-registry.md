# Native screen and presentation registry

Audit date: 2026-10-14. Scope: all named feature View declarations and opening
controls in `mobile-ios/MyPocketIOS/Features`, including private editor states.
This is a source/navigation audit, **not** a claim that every rendered state,
orientation, keyboard and populated production screen has been visually tested.
No Money Lover/web parity is an authority; these are MyPocket's own contracts.

Shared proposed guideline: [native base families](ios-base-guideline.md).
Work: [new feedback audit](../../work/tickets/NATIVE-UI-SYSTEM-AUDIT-20261014.md).
Existing docs linked below retain domain detail; this registry specifies native
entry/presentation and highlights drift instead of silently blessing it.

## Screen/entry map

Paths in the source column are under `mobile-ios/MyPocketIOS/Features/`.

| Screen/state | Entry and exit | Existing base composition | Behavior/UI spec; source |
| --- | --- | --- | --- |
| Password login | AppRoot unauthenticated; submit→session shell | MPTextField/MPSecureField/MPButton | [Auth](../pages/auth/README.md); Auth/AuthView.swift |
| Biometric lock | Restored enrolled session; unlock→shell; sign out→login | MPButton/status | [Auth](../pages/auth/README.md); Auth/BiometricUnlockView.swift |
| Main shell | Authenticated app root; overview/ledger/assistant/account tabs | MPBottomNavigation/MPFeedbackBubble | [iOS platform](ios.md); Shell/MainTabView.swift |
| Overview | Overview tab; wallet→wallet list, month→month route, bell→notifications | MPBalanceHeader/MPSurfaceCard/MPMetric/MPWalletRow/MPTransactionRow | [Overview](../pages/overview/README.md); Overview/OverviewView.swift |
| Transactions aggregate/basic | Tab or wallet selection; scope selector, week, search, overflow, add | MPScreenHeader/MPWalletScopeSelector/MPTransactionRow | [Transactions](../pages/transactions/README.md); Transactions/TransactionsView.swift |
| Transactions stale read | Transient load failure with matching encrypted snapshot; reload→live data | MPStatusMessage/MPButton, unchanged confirmed rows | [Cache work](../../work/tickets/OFFLINE-FINANCIAL-CACHE-20261014.md); actual stale UI UAT pending |
| Financial pending queue | Ledger with nonterminal current-owner intents; compact type/amount/note/status rows; three-dot eligible Retry/Edit/Discard; active+reachable FIFO retry/lease wake, inactive pause | MPFinancialQueueStatus using shared card/text/badge/button/MPOverflowMenu; 44pt menu, disabled while busy | [Lifecycle proof](../../work/tickets/OFFLINE-FOREGROUND-LIFECYCLE-20261014.md); full recovery/a11y/device UAT open |
| Rejected/conflict draft editor | Queue Edit→existing ordinary/transfer form→confirm changed draft/newkey or cancel; unchanged draft refused | Existing transaction/transfer field bases, MPButton/MPConfirmDialog | [Recovery E2E](../../work/tickets/OFFLINE-RECOVERY-E2E-20261014.md): ordinary422/409 amount/date/key and rejected/committed discard locally verified; transfer rejected-editor/device/a11y UAT open |
| Transaction create/edit | Shell plus or ledger add/edit; full-screen editor; save/cancel→caller | MPFormRow/MPAmountKeypad/MPDateField/MPCheckboxRow | [Transactions](../pages/transactions/README.md); TransactionEditorView |
| Transaction detail | Ledger row; edit/delete actions, back→ledger | Shared row/text/button | [Transactions](../pages/transactions/README.md); TransactionDetailView |
| Transaction category/date subpicker | Editor row→shared form sheet/calendar; select→return preserving draft | MPCategoryTree/MPFormBottomSheet/MPDateField | [Transactions](../pages/transactions/README.md); editor presentation modifiers |
| Transfer editor | Wallet/ledger overflow; save/cancel→wallet | MPAmountField/MPOptionSelector/MPDateField | [Wallets](../pages/wallets/README.md); TransferEditorView |
| Balance adjustment | Wallet overflow; submit→refreshed wallet | MPAmountField/MPDateField | [Wallets](../pages/wallets/README.md); AdjustmentEditorView |
| Wallet list | Overview/all-account wallets; row→scoped ledger, detail→detail, edit/add→editor | MPWalletSelectionList/MPSelectionScreenHeader | [Wallets](../pages/wallets/README.md); Wallets/WalletsView.swift |
| Wallet create/edit | Wallet list state replacement; cancel/save→list | Shared fields/amount/date/selectors | [Wallets](../pages/wallets/README.md); WalletEditorView |
| Basic wallet detail | Wallet detail action; edit full cover, overflow transfer/adjustment, back | MPWalletRow/MPTransactionRow/metrics | [Wallets](../pages/wallets/README.md); WalletDetailView |
| Savings/goal wallet detail | Goal scope/detail uses goal layout and add CTA | MPSavingsBalanceBlock/ProgressBlock/MPTransactionRow | [Savings](../pages/savings/README.md); WalletDetailView and scoped ledger |
| Credit wallet detail | Credit wallet scope/detail; payment/entry sheet | MPCreditSummary/MPFormBottomSheet | [Credit](../pages/credit/README.md); Secondary/CreditView |
| More/account landing | Account tab; profile or destination row→NavigationStack route | MPProfileCard/MPMenuRow/MPSurfaceCard | [More](../pages/more/README.md); More/MoreView.swift |
| Account settings | Profile row only; back→More | Shared fields/key copy/button; Face ID password sheet | [Account](../pages/account-settings/README.md); Assistant/AccountView |
| Face ID enrollment/forget | Account security actions; password sheet/confirm dialog | MPFormBottomSheet/MPSecureField/MPConfirmDialog | [Face ID](../../guides/face-id.md); AccountView/AuthViewModel |
| API key create/copy/revoke | Account key section add→inline form, secret once→copy, revoke action | MPTextField/MPClipboardAction/MPButton | [Account](../pages/account-settings/README.md); AccountView |
| Budgets list/detail | More budgets; card selection→inline detail/editor | MPBudgetSummary/MPBudgetCard/amount/category | [Budgets](../pages/budgets/README.md); Secondary/BudgetsView |
| Budget editor/category picker | Add/edit replaces list with shared editor; category sheet; cancel/save→list | MPEditorScaffold/MPAmountField/MPFormBottomSheet/MPCategoryTree | [Budgets](../pages/budgets/README.md); BudgetsView |
| Reports | More or ledger reports action; month selection→recalculate | MPReportsInsider/MPTrendChart/MPMetric | [Reports](../pages/reports/README.md); Secondary/ReportsView |
| Categories list | More categories; search/kind/tree; row/add→editor state | MPCategoryTree/MPTextField | [Categories](../pages/account-groups/README.md); Secondary/CategoriesView |
| Category editor | Dedicated route state; save/cancel→tree, delete→confirm | MPReferenceScreenHeader/MPWalletApplicabilityCard/MPFormFields | [Categories](../pages/account-groups/README.md); CategoriesView editor |
| Jars list/detail | More jars; month selector, report jar selection | MPJarSummary/MPMetric/MPOptionSelector | [Jars](../pages/jars/README.md); Secondary/JarsView |
| Jar editor | Add/edit replaces list with shared editor; save/cancel→list | MPEditorScaffold/MPAmountField/MPNumberField/MPWalletSelector | [Jars](../pages/jars/README.md); JarsView |
| Month detail/note | Overview month/More months; month change with dirty-note confirmation | MPMonthSummary/MPTextEditor/MPConfirmDialog | [Month](../pages/month-detail/README.md); Secondary/MonthsView |
| Recurring list/editor | More recurring; add/edit shared editor replaces list; delete confirm | MPEditorScaffold/shared amount/date/selectors/category sheet | [Recurring](../pages/recurring/README.md); Secondary/RecurringView |
| Travel list/editor | More travel; dedicated editor; optional account-timezone calendars | MPEditorScaffold/MPOptionalDateField/MPTextField/MPTextEditor | [Travel](../pages/travel/README.md); Secondary/TravelView |
| Portfolio summary/assets | More portfolio; refresh; add/trade open dedicated editors | Flat MPMetric/MPAssistantCard/MPTransactionRow | [Portfolio](../pages/portfolio/README.md); Secondary/PortfolioView |
| Asset create/trade/price | Asset/trade full-screen MPEditorScaffold; price short sheet; success closes/refetches, error retains draft | MPTextField/MPNumberField/MPAmountField/MPDateField | [Portfolio](../pages/portfolio/README.md); PortfolioView |
| Credit statement/entry/payment | More credit/credit-wallet detail; shared bottom sheet | MPOptionSelector/MPAmountField/category sheet | [Credit](../pages/credit/README.md); Secondary/CreditView |
| Notifications | Overview bell; mark-read→persist | Shared text/card/button | [Notifications](../pages/notifications/README.md); Secondary/NotificationsView |
| Finance Assistant | Assistant tab; programmatic secondary route; reload/history, send/cancel; pinned overview→details; keyboard hides tabs, header overview/dismiss; accessibility full-label navigation menu. Locally tested, owner/device acceptance pending | MPChatOverview/MPChatMessage/MPChatComposer/MPMarkdown/MPBottomNavigation/MPMenuSelector/structured cards | [Finance Assistant](../pages/finance-assistant/README.md); Assistant/AssistantViews.swift; [adaptive gate](../../work/tickets/NATIVE-CHAT-ADAPTIVE-20261014.md) |
| AI transaction entry | Long-press plus/More AI entry; attachment, process→drafts | MPDocumentPicker/MPAttachmentThumb/MPButton | [AI entry](../pages/assistant/README.md); Assistant/AIEntryView |
| AI proposal editor | Proposal edit→editor state, approve/reject→API | Shared amount/date/wallet/category fields | [AI entry](../pages/assistant/README.md); EntryProposalEditor |
| Feedback list | More feedback; counted resolution tabs, checkbox→API then move | MPSegmentedControl/MPCheckboxRow/MPSurfaceCard | [Feedback](../pages/feedback/README.md); Feedback/FeedbackView |
| Feedback create | Bubble→captured overlay; header+→shared full cover; send/cancel→caller | FeedbackOverlay/shared image picker/fields | [Feedback](../pages/feedback/README.md); FeedbackOverlay |
| Changelog | Feedback Xem changelog→inline disclosure | MPAssistantCard/MPText | [Feedback](../pages/feedback/README.md); FeedbackView |

## Current conformance failures

- Latest account API-key lifecycle verified against updated local API: create,
  copy confirmation, revoke clears secret, leave/reopen keeps revoked row absent.
  Final219/219unit+selectedAPIUI includes owner-refresh CAS regressions. Inspected
  reopen render nevertheless has Feedback over timezone control; design blocked,
  not full visual/accessibility acceptance. Raw server key metadata retained.

- API-key scope selector removed; shared full-account warning/metadata shown,
  no-scope create wire and selected editor UI verified212/212unit+UI. Inspected
  SE render exposes Feedback over warning text, so no full visual acceptance.
  Full-owner backend policy local; production activation still pending.

- Budgets/jars/recurring/travel/portfolio now use the shared editor scaffold.
  Same-item canceled-edit reopening, jar month invalidation and ended-budget
  preview/delete-cancel now pass in188/188 gate. No durable-draft promise;
  remaining all-screen/error/visual acceptance is not inferred from CRUD alone.
- Portfolio migrated to dedicated editors and flat asset metrics/history;
  compact value scale/8pt rhythm verified191/191 unit/API UI with inspected SE
  render. Feedback covers asset identity; avoidance design and final visual/
  accessibility/device sweep remain pending.
- Assistant core chat geometry is migrated and SE unit/UI/API-overview verified;
  live-provider history/cancel/error, accessibility and device acceptance remain.
- Typed decimal/integer/percentage bases now implemented; unit and actual
  fractional trade/recurrence persistence proofs pass. Full-screen sweep pending.
- MPDocumentPicker now has Photos-first shared source selection and explicit
  Files/PDF tab. Source/capacity/system cancel and real stock-photo append proof
  pass with inspected renders. Partial loader failure and mixed JPEG/PDF wire
  proofs pass at data/HTTP boundaries; live iCloud/OCR/device proof remains open.
- AI entry header/Save-all/Analyze now use shared pinned scaffold;196/196 selected
  gate protects previews/remove/approval too. Visual review still finds Feedback
  covering description placeholder; avoidance slice blocked for design review
  after third fix pass. No whole-screen acceptance or universal overlap guarantee.
- More account/notification/Assistant duplicates removed; Assistant tab is primary.
  Typed programmatic route remains; navigation/draft gate196/196 passes locally.
- Source guard catches raw controls but not nested surfaces, vertical spacing,
  duplicated destinations or mismatched presentation. Add behavior/layout proofs.

## Runtime audit register

Feedback two-tab resolve/reopen: local API-backed Simulator UI PASS (1/1),
server re-read verifies state/status. Density/native170 unit PASS. Historical
screen proofs remain in VALIDATION_MATRIX; a fresh populated all-screen visual
sweep, keyboard and Dynamic Type pass is pending for the new family guideline.
Do not mark every row visually accepted based solely on this source registry.
