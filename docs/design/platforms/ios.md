# iOS renderer contract

This document defines how the existing MyPocket screen contracts render in
SwiftUI. It does not redefine API or finance behavior.

## Renderer rules

- Feature screens live under `mobile-ios/MyPocketIOS/Features`.
- User-facing controls and visual behavior live under
  `mobile-ios/MyPocketIOS/DesignSystem`.
- Feature views compose named bases; raw controls are allowed only inside the
  base that owns their accessibility, focus, validation and styling.
- Safe-area, keyboard, Dynamic Type, VoiceOver and 44pt hit-target behavior are
  owned by the base components and verified in previews/UI tests.
- Web and iOS use the same semantic token roles and business rules, but do not
  share DOM, React Native or CSS implementation files.
- Native base components use `MPFont` for all text roles; system-font
  modifiers are rejected by `scripts/check-base-components.sh`. UIKit-backed
  fields use the registered Manrope UIFont with a safe preview/test fallback.

## Screen mapping

| Semantic screen | SwiftUI feature | Required base composition |
| --- | --- | --- |
| Overview | `Features/Overview` | `MPBalanceHeader`, `MPSurfaceCard(.reference)`, `MPSectionHeader` with web action copy (`Ví của tôi` → `Quản lý`), `MPWalletRow` with description/type subtitle fallback, `MPMoneyInsiderCard`, `MPOverviewReportCard`, category-first recent transaction titles, async states |
| Transactions | `Features/Transactions` | `MPWalletSelector`, `MPSegmentedControl`, `MPFormRow`, `MPDateStepper`, `MPTransactionRow`, `MPBottomSheet`; `TransactionsPresentation` owns account-timezone day-group header formatting |
| Wallets | `Features/Wallets` | `MPWalletRow`, `MPMenuRow`, `MPAmountField`, `MPDateField`, `MPConfirmDialog`; More-route back and wallet-row scoped-ledger handoff |
| Budgets/reports | `Features/Secondary/SecondaryFeatureViews.swift` | `MPMetricCard`, `MPProgressBar`, `MPBudgetGauge`, `MPChart`, `MPTrendChart`, `MPDonutChart`, `MPReportsInsiderCard`, `MPStatusMessage`, `MPStatusBadge`, `MPSegmentedControl`, reference summary cards, stable month/reload and section IDs |
| Jars/months | `Features/Secondary/SecondaryFeatureViews.swift` | `MPProgressBar`, `MPTextField`, `MPBottomSheet` |
| Categories/groups | `Features/Secondary/SecondaryFeatureViews.swift` | `MPScreenHeader`, `MPSegmentedControl`, outline `MPButton`, `MPOptionSelector`, `MPWalletSelector`, `MPMultiOptionSelector`, flattened category tree |
| AI entry | `Features/Assistant/AssistantViews.swift` | `MPAttachmentPicker`, `MPAttachmentThumb`, `MPActionCard`, shared header `Lưu tất cả`, async states |
| Assistant | `Features/Assistant` | pinned `MPChatOverview`, flat `MPChatMessage`/`MPMarkdown`/structured cards, remaining-height conversation, 1–4 line `MPChatComposer`, suggested-question `MPButton`s |
| Feedback | `Features/Feedback` | `MPFeedbackBubble`, `MPBottomSheet`, `MPAttachmentThumb`, `MPFeedbackStatus`, `MPStatusView` |
| Account/More | `Features/More/MoreView.swift` | `MPScreenHeader`, `MPHeaderSupportButton`, `MPProfileCard`, `MPSurfaceCard(.reference)`, `MPMenuRow`, `MPButton`, `MPConfirmDialog`, `MPStatusView` |

The persistent shell tabs stay interactive while Account/More is selected so a
user can leave any secondary destination through Overview, Transactions or
Assistant. The center add affordance is disabled on Account to keep its raised
hit region off the long menu; a narrow blocker protects the already-selected
Account tab. Raised Add is also hidden on Assistant to avoid covering its input;
Overview/Transactions retain Add. Composer reports measured floating-action
clearance so Feedback floats above Send with no reserved content lane.
`MainTabView` explicitly sizes each route to the available height
minus its 56pt sibling navbar. The 52pt `MPFeedbackBubble` overlays the trailing
bottom corner with an 8pt margin; no dedicated feedback lane is reserved. The
viewport does not change while the bubble is hidden for screenshot capture or
feedback presentation. Shell-hosted routes must not add a duplicate
`MPBottomNavigationInset`; retain that base only for independent sheets,
full-screen editors and overlays that own their own presentation boundary.

`MPWalletRow` composes `MPWalletGlyph`, a shared 48pt circular badge whose
SF Symbol and semantic tone mirror the web `WalletCard`: basic wallets use the
warning tone, goal wallets success, and credit wallets teal. Wallet data and
accessibility behavior remain unchanged.

Authentication uses `AuthView` with the shared email/password fields and a
shared `MPBrandMark`, direct canvas title/subtitle, flat fields and full-width
primary action. The native app does not expose Google login or include a Google
client/callback configuration. It stores the access/refresh session from
`POST /api/v1/login`; session restore, refresh and optional Face ID unlock use
the same backend session contract. Independent remembered-device enrollment
adds Face ID sign-in after logout/refresh expiry through the four device-auth
endpoints in [AUTH_API](../../architecture/AUTH_API.md). Credentials are
biometryCurrentSet/ThisDeviceOnly protected, never saved passwords. Server-side Google OAuth remains available
to existing web clients.

The native transaction editor uses `MPFormRow`/`MPDateStepper` inside the
reference surface: wallet, category, note and account-local date navigation
share the same icon/value/chevron base. Its `MPAmountField` transaction variant
shows a leading VND chip and large unboxed value with Vietnamese thousands
grouping for display; the bound and API values stay unformatted integer digits.
Other screens use the standard bordered amount-field variant. Jar rules, detail
toggle and safe-area save footer retain their API/validation behavior. New/edit forms are presented full-screen above the shell navigation
so the footer cannot be occluded. `MPReferenceScreenHeader` owns the Stitch
close capsule and `MPActionIconButton` owns the image/AI-entry footer action;
both are shared bases rather than screen-local controls. The focused editor UI
proof covers all stable row identifiers and a disposable API-backed fixture.

Goal-wallet scope now follows the supplied Stitch savings ledger anatomy: the
native transaction route removes the ordinary title/period controls for a goal,
renders a compact help/wallet/search/overflow header, centers balance and
remaining target, shows the shared savings progress calculation and advice,
then renders the owner-scoped month ledger, report navigation and
add-transaction action. The
wallet selector and goal summary are interactive/data-backed; search filters
category/note text locally and overflow reloads or returns to the aggregate
scope. `MPSavingsSummary`, `MPWalletGlyph` and `MPHeaderIconButton` own the
geometry and semantic tones for both embedded and explicit goal-wallet routes.

AI Entry keeps the web review contract: the AI produces editable proposals,
and the shared header exposes `Lưu tất cả` only while pending rows exist. The
batch action uses the same versioned approve endpoint as individual rows and
stops on the first error; it never writes directly to the ledger or retries
automatically.

Finance Assistant structured parts use `AssistantCardType` as the native
vocabulary gate. It normalizes the web aliases (`finance_summary`,
`spending_comparison`, `transaction_search`, `budget_progress` and
`wallet_creation_proposal`) before selecting a shared `MPAssistantCard`, and
renders the web `transaction_detail` contract with wallet/category, amount,
report-inclusion and transfer flags. Unknown kinds remain a safe readable
fallback; no backend response or tool contract is changed. The composer also
uses `AssistantComposerPolicy` to keep the editor/send action disabled until
the capability response enables the provider and no run is active. Canonical
`goal_progress` cards additionally render a clamped `MPProgressBar` and
percentage from each goal's current/target values; ordinary wallet-balance
cards do not show goal-only progress. `jar_progress` resolves the backend
`allocation_amount` before budget fallback keys and renders spent-only text
when the hũ has no configured allocation, matching the web card contract.
Metric, transaction-list and wallet cards also retain web-compatible source
provenance/count labels, while budget/jar cards keep their period totals.
Card rows are bounded like the web renderer: transaction lists show up to five
rows; wallet, goal, budget and jar lists show up to eight. Transaction totals
continue to use the server-provided `total_count` when available, rather than
the visible row count.
The shell persists the privacy-mask toggle in `mypocket.privacy.masked` and
provides it through the shared `mpFinancialValuesMasked` SwiftUI environment.
When enabled, all visible financial values in Overview, wallet selection and
details, transaction list and details, recurring, portfolio, credit, budgets,
reports, jars, month detail, AI entry proposals and Finance Assistant render as
`••••••`. Accessibility labels use the same masked value; amounts in editable
form fields remain visible so users can enter or verify a draft. API payloads,
calculations and owner scope are unchanged. The environment defaults to
unmasked outside the authenticated shell.

The native account screen uses the Stitch grouping contract in
`docs/work/PIXEL-PARITY-ACCOUNT-TRANSACTION.md`; route IDs remain stable while
the visible title is the Money Lover-style “Mở rộng”; `MPCenteredScreenHeader`
keeps that title centered independently of the support pill. The transaction editor
uses the same shared bases plus `MPReferenceScreenHeader` and a safe-area
save bar so the primary action remains reachable on compact phones.

Reports use the Stitch order `header → filter → summary → breakdown → activity`.
The summary card owns income, spending, net and counted activity metrics plus a
detail/trend segmented control. Detail uses the existing category bars; trend
uses the shared `MPTrendChart` with daily account-timezone points from the
owner-scoped transaction list, excluding non-report/transfer/adjustment rows.
Its VoiceOver value reads each plotted day's income/expense totals in order and
uses the same masked display value as the rest of the account when amounts are
hidden.
The subsequent Money Insider card renders the period delta, daily average,
spending/income ratio, top categories and top expenses from the same response,
with shared masking. `ReportsView` keeps the existing month filter and
owner-scoped report endpoints unchanged; the summary card trailing label is the
selected month (matching web), while `Xem báo cáo` remains the explicit reload
action. The web Reports panel mounts the matching `BaseTrendChart` and uses the
same filtering contract.

Month detail mirrors the web report metadata: the selected month summary shows
its account timezone and renders the full local start/next-start timestamps,
with the exclusive next boundary explicitly labelled. `MonthsPresentation` owns
that formatting so the screen does not duplicate timezone/date rules.

Feedback cards render the submitted description and the full changelog
description. `MPFeedbackStatus` owns the status label and shows `Fixed in
<version>` only when the feedback is fixed and its linked changelog entry is
present; API payloads and owner scope are unchanged.

Budgets use the shared reference summary card for overall progress and
budget/spent/remaining metrics. Each budget also exposes the same timing,
active daily-spend suggestion and 75%/over-budget status copy as the web
panel, including masked-money accessibility values; CRUD and owner scope stay
on the existing API contract. When there are no budget items, the summary card
is withheld and the shared `MPEmptyView` shows `Chưa có ngân sách.`, matching
the web empty branch. The header now uses the shared compact `MPButton` with
copy `Thêm` and stable ID `budget-add`, matching the web `BudgetsPanel` action;
the budget editor and API behavior are unchanged.

Credit ledger keeps the shared statement-card treatment for populated rows, but
when a credit wallet has no statement rows it renders the web-equivalent
`Chưa có giao dịch tín dụng.` as plain text outside the list card. The action row
appears immediately after the summary and before the statement region; the
entry/payment actions remain shared controls.

The credit summary now uses one shared `MPSurfaceCard(.reference)` matching the
web anatomy: current debt and available credit are the primary pair, followed
by limit and configured payment-status/due-date details. The action row keeps
the web copy `Thêm giao dịch` and `Thanh toán thẻ` with stable IDs
`credit-add-entry` and `credit-add-payment`; statement selection, entry/payment
payloads, empty-state copy and account-timezone formatting are unchanged.

Categories keeps the existing owner-scoped category/tree and editor behavior,
but matches the web Groups route's action anatomy: the header is navigation-only
and `Nhóm mới` is a full-width outline `MPButton` below it. The outline style is
owned by the shared button base (surface, line, action color and canonical hit
target), with accessibility ID `category-add-group`; no screen-local button
geometry or API contract is introduced.

Transaction, budget, recurring and credit category pickers use the shared searchable hierarchy;
each root group is its own `MPSurfaceCard`, matching the web selector's
per-root-card composition. Search retains matching ancestry, and selection
eligibility and persisted category IDs remain owned by the existing API-backed
form flows.

Jars keeps the header navigation-only and places the shared compact `MPButton`
`Thêm hũ` beside `Hũ tháng`, matching the web monthly section action with stable
ID `jar-add-month`; month/cumulative data and jar API behavior are unchanged.

Recurring keeps the page header navigation-only and places the shared compact
`MPButton` `Thêm` beside the existing recurring summary title, matching the web
card action with stable ID `recurring-add`; run-due, schedule CRUD and timezone
behavior are unchanged.

Travel Mode keeps the page header navigation-only and places the shared primary
`MPButton` `Thêm` beside the summary title, matching the web card action with
stable ID `travel-add`; event CRUD, active state and date validation are
unchanged.

Portfolio keeps the page header navigation-only apart from the shared compact
`MPButton` `Làm mới` (`portfolio-refresh`). The supporting subtitle
`Theo dõi mua bán độc lập với số dư ví` and the `Tổng danh mục` reference card
match the web panel. The web-equivalent `Thêm tài sản` and `Ghi giao dịch` cards are always visible when the portfolio summary is
loaded, and compose `MPTextField`, `MPOptionSelector`, `MPWalletSelector`,
`MPDateField` and shared `MPButton` bases. The per-asset `Giao dịch` action
selects the shared trade card rather than creating a second screen-local form;
portfolio API payloads, validation, account timezone serialization and
owner-scope behavior are unchanged. Stable IDs are `portfolio-add-asset`,
`portfolio-trade-asset` and `portfolio-save-trade`.

Native trade history is loaded from the existing owner-scoped
`GET /api/v1/portfolio/assets/{id}/trades` endpoint for each displayed asset;
the aggregate summary response is authoritative for positions/valuation but
does not currently populate its optional `trades` field. After a buy/sell, the
screen reloads both summary and history so the position and immutable trade row
are shown from persisted API state.

## States and accessibility

Every feature exposes loading, ready, empty and error states. Write forms also
expose dirty, submitting, success and conflict states. Async state bases must
provide a visible label and accessibility value; masked money must not leak the
unmasked amount through VoiceOver. UI tests use stable accessibility IDs rather
than screen text alone.
