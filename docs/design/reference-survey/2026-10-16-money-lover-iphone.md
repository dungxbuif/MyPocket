---
artifact_type: visual_reference_survey
id: MONEY-LOVER-PHYSICAL-20261016
status: in_progress
owner: shared
artifact_source: physical_device_reference
reference_app: Money Lover
reference_version: 8.89.0 (9250)
reference_bundle_id: me.moneylover.ios.Money-Lover
updated: 2026-10-16
---

# Money Lover physical iPhone reference survey

## Intent and scope

Owner requested visiting Money Lover functions on the connected iPhone and
using observed UI as the visual reference for MyPocket. Preserve MyPocket's
extended functions and financial/security/API behavior. This document records
actual evidence separately from planned coverage; app discovery alone proves
no visual parity.

Related: [device tooling probe](../../operations/MOBILE-MCP-DEVICE-PROBE.md),
[native screen registry](../platforms/ios-screen-registry.md),
[native base guideline](../platforms/ios-base-guideline.md),
[base contracts](../system/BASE_COMPONENTS.md),
[native feedback work](../../work/tickets/NATIVE-MOBILE-FEEDBACK-20261013.md),
[validation](../../work/VALIDATION_MATRIX.md).

Small task exemption: yes.
Reason: evidence inventory only; no product source or business behavior change.
Impact checked: API=no, DB=no, Security=no, Runtime=no, Standards=no.
Automation helper setup is a separately owned developer-tool operation.

## Survey policy and verification plan

1. Discover the installed reference app by device inventory; never guess its
   bundle ID or assume the App Store version matches this phone.
2. Read the accessibility hierarchy before navigation. Record observed entry,
   title, visible controls, layout, back/dismiss behavior and reachable states.
3. Capture only privacy-safe reference screenshots. Do not commit raw personal
   finances, account identity, credentials, notifications or unrelated screens.
4. Inspect draft/editor/picker UI without saving financial changes. Do not
   submit purchases, sync, transfers, deletes, account changes or transactions.
5. Map each observed region/action to an existing named MyPocket base. Record
   a missing base/variant before implementation; screen-specific overrides do
   not meet the project's reference fidelity rule.
6. Track visited/unvisited/unavailable functions and conditional branches. An
   app paywall, unavailable wallet type or login requirement is not evidence
   of the content behind it.

## Verified connection evidence

On 2026-10-16, the paired USB iPhone is discoverable by `devicectl` and Mobile
MCP 1.0.11. MCP initialize over `http://127.0.0.1:19347/mcp` succeeds.
Device discovery uses the explicit device ID for later calls.

App discovery command (pass):

```sh
rtk proxy xcrun devicectl device info apps \
  --device 00008120-001C28C02131A01E \
  --include-default-apps --search Money
```

Result: Money Lover, `me.moneylover.ios.Money-Lover`, version 8.89.0,
build 9250. `devicectl` defaults to developer-installed apps; omitting
`--include-default-apps` initially showed MyPocket only. That omission must not
be interpreted as Money Lover being absent.

Read-only UI probe (fail, environmental prerequisite):

```sh
rtk proxy /Users/dungxbuif/.npm/_npx/9f1df1bfb1203d38/node_modules/@mobilenext/mobilecli-darwin-arm64/mobilecli-darwin-arm64 \
  snapshot --device 00008120-001C28C02131A01E
```

Exact cause: `failed to start agent on device ...: agent is not installed,
use 'mobilecli agent install --device ... --provisioning-profile <path>' to
install it`.

CLI help confirms physical iOS installation requires
`--provisioning-profile`. The existing profile is app-specific to MyPocket;
do not use it to sign the helper. The helper needs a separately signed identity.
Setup was assigned to the dedicated installer agent. At 21:04 +0700 it verified
Device Kit 0.0.34 build 1 under the separate bundle
`AU442S3MN3.com.mobilenext.devicekit-iosUITests.xctrunner`. The installer then
released device UI control to this survey. Hierarchy reads now pass. This survey
agent performed no signing/install mutations.

## Observed reference coverage

| Evidence | Status | Result |
| --- | --- | --- |
| Physical device discovery | verified | Paired USB phone found |
| Money Lover installed identity/version | verified | Actual bundle/version above |
| Accessibility hierarchy | verified | Helper installation resolved initial prerequisite |
| Reference screens visited | in_progress | See observed subfunction ledger below |
| Privacy-safe screenshots inspected | in_progress | Account navigation crop and recurring editor visually inspected |
| Interaction/function parity accepted | unverified | No UI completion claim |

## Observed subfunction ledger — 21:04–21:11 +0700

Private screenshots are stored under
`/tmp/mypocket-reference-20261016.tpx0QV/`, outside Git. Only labels, geometry
and behavior are recorded here; no raw account identity or transaction details.
The owner explicitly authorized local financial reference screenshots. Unrelated
notification banners are excluded from retained reference evidence.

| ID | Actual action path / state visited | Observed controls and composition | Evidence / pending branches |
| --- | --- | --- | --- |
| ML-01 | Launch → Tổng quan | Balance/header eye, search, bell; wallet block; Money Insider; report with week/month tabs; top categories; recent transactions | AX; viewport 430×932pt; private values omitted. Scrolled chart/detail actions still pending |
| ML-02 | Tổng quan → balance eye → restore | Eye masks only headline balance; wallet-row values remain visible | AX before/after; original visibility restored immediately |
| ML-03 | Account tab → Mở rộng | Support header; profile; grouped wallet/category, linked banking, event/recurring/bill, debt/tools/travel; lower export/store/discover/support/settings/about | `account-navigation-crop.png` visually inspected; identity excluded |
| ML-04 | Account → Giao dịch định kì list | Centered title/back, wallet scope glyph+chevron, plus; explanatory caption and centered empty state | AX; empty list actually observed |
| ML-05 | Recurring list → plus → create | Rounded sheet; Cancel/title/disabled Save; grouped wallet, amount, category, note; sync warning; separate repeat-options row/caption | `recurring-create.png` visually inspected |
| ML-06 | Recurring create → repeat options | Frequency; interval; start; end modes Forever / Until date / occurrence count | AX; draft only |
| ML-07 | Repeat options → frequency | Inline native wheel with daily/weekly/monthly/yearly | AX plus visually inspected transient screenshot; unrelated notification excluded; selected yearly only in unsaved draft |
| ML-08 | Repeat options → count end mode | Checkmark moves; integer text field appears, default 1 occurrence | AX; no keyboard input/save |
| ML-09 | Repeat options → until-date end mode | Checkmark moves; inline 176pt date picker appears | AX; no date committed to ledger |
| ML-10 | Repeat options → interval | Inline native wheel appears under interval row | AX; default 1 years after draft frequency change |
| ML-11 | Recurring list → header wallet scope | Chọn Ví sheet; Close/Edit; total row; title-over-balance wallet rows; selected check; add/service footer | AX; private amounts omitted |
| ML-12 | Wallet selector → Edit → Ví Của Tôi | Header back/sort/Add; TÍNH VÀO TỔNG section; wallet rows | AX; no sort/save executed |
| ML-13 | Wallet management → Add | Four illustrated choices in 2×2 grid: basic, linked, credit, savings; each has info affordance | `wallet-types.png`; linked service not activated |
| ML-14 | Add wallet → Savings draft | Icon/name; currency; target; opening balance; deadline; notification switch default on; exclude-from-total default off | `wallet-savings-create.png`; Cancel back to type choice; no Save |
| ML-15 | Savings draft → deadline | Centered calendar popup; month header with month/year drill affordance; prev/next; Monday-first six-week grid; Cancel | `wallet-date-picker.png`; canceled without date selection |
| ML-16 | Savings/credit draft → currency | Search header; flag + title + code/symbol rows; common currencies above full list | AX; VND selected only in unsaved credit draft; search branch pending |
| ML-17 | Add wallet → Credit draft | Icon/name; currency; notification/exclude switches; Next disabled until currency | AX; initial name matches existing wallet |
| ML-18 | Credit draft → currency VND → Next | Alert Cảnh báo / Tên ví đã tồn tại. / Cancel | Actual safe duplicate-name validation. Further creation step unvisited because no wallet may be created |
| ML-19 | Add wallet → Basic draft | Icon/name/currency/current balance; service-link row; exclude-from-total switch | AX; no Save/service action |
| ML-20 | Basic draft → icon | Chọn Biểu Tượng; Basic/Downloaded tabs; Store action; six-column icon grid, icons 40×40pt, pitch about 74×70pt | `icon-picker-basic.png`; downloaded tab visited, no download/store purchase |
| ML-21 | Wallet management → existing savings detail → Edit | Detail identity/currency/categories/goal/deadline/member role; notification/exclude/archive controls; transfer/adjustment/delete actions below scroll. Edit retains name/icon/currency/goal/deadline; no opening-balance field | `wallet-savings-detail.png` private; existing detail switches not changed |
| ML-22 | Savings edit → target amount | Dedicated amount route with large currency/value, clear, 4-column calculator: C/divide/multiply/backspace; digits/subtract/add; 000/comma; double-height Done | `amount-keypad.png`; no amount edited |
| ML-23 | Wallet detail → transfer draft | From amount/wallet/category/note; destination wallet; date; exclude-from-report on by default; optional transfer-fee switch | `transfer-draft.png`; no Done |
| ML-24 | Transfer draft → fee on / destination picker | Fee adds amount/wallet/category/note fields; destination picker excludes source wallet | AX; all unsaved and canceled |
| ML-25 | Wallet detail → adjust balance | Wallet, current balance amount, exclude-from-report switch on; disabled Save while unchanged | AX; canceled |
| ML-26 | Account → Categories expense hierarchy | Three type segments; wallet scope/search; New Group; root cards contain child connectors, consistent text column, applicability subtitle; bottom show-inactive control | `categories-expense-tree.png` visually inspected |
| ML-27 | Categories → existing child edit | Name/icon; fixed type; parent with clear affordance; Merge; wallet applicability Edit; delete icon; disabled pinned Save | `category-child-edit.png`; no merge/delete/save |
| ML-28 | Child edit → parent picker → search no-match | Searchable category list, selected parent check; no-match becomes empty table; keyboard shown; Close exits search | AX; search text cleared by close, no parent change |
| ML-29 | Child edit → applicability Edit | Separate sheet with search and three wallet rows with on switches | AX; applicability switches not changed |
| ML-30 | Categories → New Group | Empty name/icon, income/expense segment, optional parent, bottom disabled Save | AX; toggled income only in unsaved draft |
| ML-31 | Categories income → system transfer category | Locked glyph in list; editor name/icon disabled; explicit system-category explanation; no delete/save action | AX; read-only system state actually observed |
| ML-32 | Categories debt + show inactive + scope picker | Debt system rows; show/hidden inactive view toggle; shared wallet scope selector | AX; inactive-view toggle restored; no category activated |
| ML-33 | Budgets tab → load → empty | Actual spinner then empty message; Create; explicit View Sample Data action | AX; `budgets-list.png` captured during load transition, not a populated list |
| ML-34 | Budgets → overflow → ended | Overflow offers ended budgets; separate ended page with actual empty state | AX; no budgets in owner dataset |
| ML-35 | Budget create | Category; amount; period; wallet; repeat switch off; explanatory caption; bottom disabled Save | `budget-create.png`; unsaved |
| ML-36 | Budget create → period | Sheet choices week/month/quarter/year/custom, each bounded-period label; Cancel | AX; account's current custom month boundary is a reference preference, not MyPocket financial policy |
| ML-37 | Budget period → custom → start calendar → month/year | Separate start/end route with disabled Done; shared calendar; header opens two native month/year wheels with Back/View | `budget-custom-date.png`; calendar and month/year drilldown AX; no range saved |
| ML-38 | Budget create → repeat on | Switch turns on; same repeat-explanation caption | AX; draft canceled |
| ML-39 | Budget create → category multi-selection | All-categories switch; hierarchical expense rows with switches; selecting parent selects observed children; Done enables | `budget-category-picker.png`; picker/draft canceled |
| ML-40 | Budget empty → built-in sample data | Explicit blue sample disclaimer + Off action; data is not owner's real finances | Sample mode actually activated; no fixture/ledger writes; persists after relaunch |
| ML-41 | Sample budgets list | Month tab; big summary semicircle gauge; limit/spent/days metrics; Create; category cards with remaining/progress/Today marker | `budget-sample-list.png` visually inspected; sample evidence only |
| ML-42 | Sample budget → detail / scroll | Summary amount/spent/remaining/progress/Today/date/scope; cumulative actual + dotted projection + limit graph; recommended daily/forecast/actual daily; category breakdown; filtered transactions; delete | `budget-sample-detail.png`, `budget-sample-forecast.png`; sample forecast visually over limit/red; no delete |
| ML-43 | Sample budget → transaction list → transaction detail | Count/sum; day headers; category/note/signed amount rows; detail has amount/category/note/date/wallet/Edit/Copy/Delete | `transaction-sample-detail.png`; sample only; no Copy/Delete |
| ML-44 | Sample transaction → Edit | Wallet/amount/currency/category/note/date with previous/next; With/location/event/reminder/photo/report exclusion; bottom disabled Save | `transaction-sample-edit.png`; no Save |
| ML-45 | Transaction edit → note | Dedicated growing text-view route; previous-note suggestions with category/amount context below | `transaction-note-editor-crop.png`; no text changed |
| ML-46 | Transaction edit → add image | Sheet choices Photo Library / Camera / Cancel | AX; source sheet canceled, no library/camera permission or media opened |
| ML-47 | Transaction edit → reminder on → time | Draft switch reveals compact date/time controls; time opens native hour/minute wheels in popover | AX; reminder restored off before cancel; no persisted reminder |
| ML-48 | Transaction edit → event → add event | Empty event picker; Add opens draft name/icon/end-date/currency/wallet with disabled Save | AX; no event saved |
| ML-49 | Transaction edit → With | Name entry, Done, contacts list section; empty list in sample | AX; no contacts read/copied, no name entered |
| ML-50 | Transaction edit → location | App unexpectedly returned to Home Screen; relaunch succeeded and sample mode persisted | Location UI unobserved; do not claim location parity. No repeat crash experiment |
| ML-51 | Sample ledger | Wallet capsule, search/overflow, period tabs; income/expense/net summary; report CTA; day-grouped rows | `ledger-sample.png`; sample only |
| ML-52 | Ledger overflow / period / category grouping | Menu period/group/multi-delete/transfer/adjust/sync; period menu day/week/month/quarter/year/all/custom; grouping changes to category total/count with transaction date/note rows | `ledger-overflow.png`, `ledger-month-category-group.png`; sync/delete not submitted; individual period branches still tracked below |
| ML-53 | Ledger → period report overview | Report sheet Close/wallet/calendar/help; period tabs; opening/closing balances; net/income/expense; category report cards; debt/lending/other lower sections | `reports-period-sample.png`; sample only |
| ML-54 | Report → net income detail → week drilldown | Total; chart; per-week income/expense/net rows; row opens filtered count/sums/day-grouped transactions | `reports-net-income-detail.png`; AX drilldown |
| ML-55 | Report → expense detail / Trend | Total/average daily; optional previous-three-month comparison; Detail/Trend segments; category donut/list or period bar chart; drilldown rows | `reports-category-expense.png`, `reports-expense-trend.png` |
| ML-56 | Expense report → show three-month comparison | Expands gray comparison card; actual red cumulative line and gray average trajectory, current marker/value; Hide action | `reports-compare-three-months.png` visually inspected |
| ML-57 | Report → income detail / Trend | Income counterpart with green semantic totals, optional comparison, Detail/Trend and period rows | `reports-category-income.png`; comparison was restored hidden |
| ML-58 | Ledger → simple search → query | Search by tag/category/text, config, Close; actual compatibility warning for some wallets with Update action; sample search produced broad results despite entered query | `ledger-search-controls.png`; no app update; query semantics not inferred from sample response |
| ML-59 | Search → config | Advanced rows wallet/amount/category/date/location/note/tag/With; Search action and Simple mode | AX; no location reattempt after app exit |
| ML-60 | Advanced search → amount / range | Amount popover all/less/greater/exact/range; range opens From/To input route with Done | `search-amount-range.png`; no bounds submitted |
| ML-61 | Advanced search → date | Date popover all/before/after/exact/range | AX; operator-specific results unverified |
| ML-62 | Advanced search → category type | Popover all/all expenses/all income | AX; no filter saved |
| ML-63 | Advanced search → tags | Dedicated tag text-entry route and empty list | AX; no tags created |
| ML-64 | Central plus → new transaction modes | Expense/income/debt segments; wallet/currency/amount/category/note/date-stepper; More Details; amount suggestions above calculator; Save + AI receipt icon | `transaction-create-debt-crop.png`; all three mode selections visited, draft canceled |
| ML-65 | New debt transaction → category picker → lend | Picker has lend/pay debt/receive debt/borrow; selecting lend expands With/location/event/reminder/image fields; report exclusion on and future reminder shown in draft | AX; no debt transaction saved; do not copy reference defaults as MyPocket financial policy |
| ML-66 | Transaction creator → AI receipt | Image-entry sheet Photo Library / Camera / Cancel | AX; canceled before media/provider processing |
| ML-67 | Account → bills | Explanation, wallet scope and add; actual empty state | `bills-list-sample.png`; sample mode also has no bills |
| ML-68 | Bills → add → repeat options | Wallet/amount/category/note; non-sync note; shared repeat-options screen with daily/interval/start/three end modes | `bill-create.png`; Save appeared enabled despite empty fields, deliberately unpressed |
| ML-69 | Account → debt ledger → person | Unpaid/unreceived totals; person count/subtotal; person detail loan subtotal and Receive CTA | `debt-people-list.png`, `debt-person-detail.png`; Receive/payment not submitted |
| ML-70 | Account → Tools | ATM finder, bank finder, tip calculator, interest and personal-income-tax tool rows | AX; individual finder/interest/tax branches pending |
| ML-71 | Tools → tip calculator | Amount, five-person glyph selection with More, five-heart tip selection with More, Calculate | `tools-tip-calculator.png` visually inspected |
| ML-72 | Tip → person / tip More | Each expands inline native wheel beneath its row | AX; no real transaction data entered |
| ML-73 | Tip → Calculate at defaults | Pure calculator result shows total paid/tip/people/per-person and Add Transaction CTA | AX; no Add Transaction; result/edit/close navigation inspected |

Connection checkpoint 21:33–21:36 +0700: after closing the tip form, MCP
hierarchy calls stopped returning. Direct `mobilecli agent status` still
reports the separate helper installed. Direct `snapshot` also waits without
output. Read-only `xcrun devicectl device info lockState --device ...` returns
`passcodeRequired: true`, `unlockedSinceBoot: true`; owner unlock is required
before further UI work. No passcode attempt was made. Sample mode remains
active until the next successful `Tắt` action; restoration is not yet claimed.
No reference financial record was created, changed, paid or deleted.
The pending direct snapshot eventually exited with
`RPC call device.dump.ui failed: Post "http://localhost:8100/rpc": context deadline exceeded`.

### Shared visual observations

Observed account and draft screenshots use a very light neutral/lavender page
background, flat white grouped surfaces, light separators inset to the text
column, black outline navigation icons, green primary/selection accents and
large rounded group/sheet corners. These are actual reference observations;
the existing MyPocket 12pt-card/8pt-button rule is not proof of fidelity to this
new reference. Any token/base change must be centralized and approved in the
parent's implementation scope.

Account rows have leading icons at x36 (24pt), text at x76 and chevrons near
x379; vertically adjacent rows are approximately 50pt apart. Wallet selection
uses 40pt glyphs at x36 with text at x92, a title above balance and approximately
72pt row pitch. Recurring form rows use the same x92 text column; amount is
larger and has a small label above it. Ordinary navigation is y59/h54;
presented-sheet navigation starts at y79 or y89 depending on nested presentation.

The tab bar accessibility frame is x0/y849/w430/h83, but the screenshot shows
the rendered control as a floating translucent capsule above the bottom safe
area, with an elevated central green plus and a pale rounded selected tab.
AX frames alone would miss this visual distinction.

In sample mode the blue disclaimer occupies the bottom 62pt and shifts the
tab-bar frame up to y787/h83. This banner is reference-app demo chrome and
must not be copied into MyPocket production layout. The ledger screenshot
also shows a balance subtitle beneath its wallet capsule, then the period
rail, then the income/expense/net summary. There is no visible large page
title duplicating the wallet capsule, despite a screen-title AX element.

The sampled transaction editor has one large grouped surface containing
wallet, amount, category, note and date controls; With is a separate group;
location/event/reminder are another; image and report exclusion are separate
single-row groups. Date is a previous/center-date/next control within the row.
Save is pinned across the bottom on its own white footer, disabled gray when
unchanged. Large fields are unboxed values inside the group, not a separate
outlined input for every row.

The report overview uses a white net-income group with horizontal income and
expense bars, then a separate category group with two gray summary tiles and
two donut charts with peripheral icon/percentage labels. Expense comparison
uses an expandable gray chart panel; red is the expense data series and green
the income counterpart. Preserve exact MyPocket values and privacy masking;
the sample values are not fixtures for MyPocket calculations.

### Region/action/state → native base mapping

| Observed reference region or interaction | MyPocket base ownership | Required proof / potential shared gap |
| --- | --- | --- |
| Floating capsule navigation, raised plus, selected rounded tab | `MPBottomNavigation` / shell layout | Shared variant required if current rectangular sibling bar differs; safe-area and keyboard proof; retain Assistant entry |
| Capsule back/cancel/save/help/overflow header controls | `MPScreenHeader`, `MPCenteredScreenHeader`, `MPHeaderIconButton`, `MPOverflowMenu`, `MPEditorScaffold` | Header variants own shapes/hit targets; feature views must not override button radii |
| Wallet glyph/title/chevron capsule with balance below | `MPWalletScopeSelector`, `MPWalletScopeGlyph`, `MPMetric` | Aggregate/basic/goal/credit variants; scope selection and balance privacy must stay owner-safe |
| Underlined horizontally scrolling period rail | Shared selection/presentation family; existing `MPSegmentedControl` only if a suitable period variant exists | Rounded segmented control alone does not reproduce this rail; preserve selected period and accessibility |
| Ledger income/expense/net rows and centered report link | `MPMetric`, `MPText`, `MPTextLinkButton`, shared divider | Flat aligned rows, signed amounts, no guessed totals or duplicated card nesting |
| Day-group white surface, large day number, date context, total | `MPSurfaceCard` + a shared day-header composition | Header/body grouping, date timezone, width/large-text proof |
| Transaction category glyph with optional wallet badge; title/note/value | `MPTransactionRow` / `MPTransactionPresentation` | Reference variant needs 40pt glyph and note line; transfer blue/neutral vs expense red/income semantics centralized |
| Icon/name + currency/amount/category/note/date grouped form | `MPFormRow`, `MPAmountField`, `MPTextField`, `MPDateField`, `MPEditorScaffold` | Reuse row-based unboxed presentation; fixed footer and keyboard reachability |
| Calculator digit/operator grid, suggested amounts | `MPAmountKeypad` | Keyboard owns numeric semantics and geometry; suggestions cannot change bound value until chosen |
| Parent/child root group, connectors, smaller child glyph, bottom search | `MPCategoryTree` / `MPCategoryPicker` | Root40/child32 geometry observed; switch/select/management variants and pinned search belong to base |
| Multi-category parent switch propagating to child switches | Selection family / `MPCategoryPicker` variant if supported | Keep reference UX separate from MyPocket API category-scope capabilities; no silent backend contract expansion |
| Calendar modal plus month/year drilldown; recurrence wheels | `MPDateCalendarSheet`, `MPDateField`, `MPNumberField`, `MPOptionSelector` | Shared presentation variants, timezone and integer frequency semantics |
| Budget semicircle summary, progress/Today marker, forecast chart | `MPBudgetSummary`, `MPBudgetGauge`, `MPBudgetCard`, shared chart family | Exact backend spending/projection; clamped geometry without clamping financial values; over-limit state |
| Report total/daily average, detail/trend, comparison panel and drills | `MPMetric`, `MPDonutChart`, `MPTrendChart`, shared selection family | Reference has distinct bar/donut/comparison compositions; a generic line chart is not blanket parity |
| Photo/Camera choice; note/With/reminder subroute | `MPDocumentPicker`, `MPTextEditor`, shared selection/presentation | MyPocket must retain Files/PDF and AI approval; no automatic media/provider submission |
| Empty/spinner/disabled/system-locked/compatibility warning | `MPAsyncStates`, `MPStatusMessage`, `MPButton`, typed field enabled state | Use actually observed states; never manufacture unavailable success/error screenshots |

Automation note: the immediate hierarchy returned inside a batch can show
transitional off-screen y-coordinates after dismissal. Re-read after transition
before using references; do not count a transitional tree as another screen.

## Planned function coverage and base mapping

This function map relates actual observed states to MyPocket's current screen
registry. Partial means one or more branches remain; it is not completion of
that function. See the explicit remaining register below.

| Function to investigate | Reference visit | MyPocket target / reusable bases |
| --- | --- | --- |
| Main navigation and screen headers | partial ML-01/03/51 | `MPBottomNavigation`, `MPScreenHeader`, `MPCenteredScreenHeader`, `MPHeaderIconButton` |
| Wallet switcher/list and aggregate scope | partial ML-11/12 | `MPWalletScopeSelector`, `MPWalletSelectionList`, `MPWalletSelectionRow`, `MPWalletGlyph`, `MPWalletAggregateGlyph` |
| Basic wallet details and management | partial ML-13/19/20/21 | Wallet pages; `MPWalletRow`, `MPMetric`, `MPOverflowMenu`, `MPEditorScaffold` |
| Ledger period/navigation/search | partial ML-51/52/58–63 | Transactions; `MPTransactionRow`, `MPTextField`, `MPSegmentedControl`, `MPDateStepper` |
| Add/edit transaction; note/report inclusion | partial ML-44–49/64–66 | `MPFormRow`, `MPAmountField`, `MPAmountKeypad`, `MPTextEditor`, `MPCheckboxRow` |
| Category search/tree/selection | partial ML-26–32/39/65 | `MPCategoryPicker`, `MPCategoryTree`, `MPFormBottomSheet` |
| Date/calendar/time selection | partial ML-09/15/37/47 | `MPDateField`, `MPDateTimeField`, `MPDateCalendarSheet` |
| Transfer and balance adjustment drafts | partial ML-23–25 | Wallet editor flows; `MPAmountField`, `MPOptionSelector`, `MPDateField` |
| Transaction detail/actions | partial ML-43/44 | `MPTransactionRow`, `MPText`, `MPButton`, `MPConfirmDialog` |
| Reports and period selection | partial ML-53–57 | `MPTrendChart`, `MPDonutChart`, `MPMetric`, `MPOptionSelector` |
| Budget list/detail/editor/category scope | partial ML-33–42 | `MPBudgetSummary`, `MPBudgetCard`, `MPBudgetGauge`, `MPEditorScaffold` |
| Savings/goal list/detail/add-money draft | partial ML-14/21/22; scoped ledger unvisited | `MPSavingsSummary`, `MPSavingsBalanceBlock`, `MPProgressBar`, `MPTransactionRow` |
| Recurring transactions/scheduled bills | partial ML-04–10/67/68 | Recurring; `MPEditorScaffold`, `MPNumberField`, date/category/selector bases |
| Credit/debt/repayment draft | partial ML-17/18/65/69 | `MPCreditSummary`, `MPFormBottomSheet`, shared amount/category fields |
| Categories management/editor | partial ML-26–32 | `MPCategoryTree`, `MPReferenceScreenHeader`, `MPWalletApplicabilityCard` |
| More/profile/settings/navigation | partial ML-03/70–73; settings unvisited | `MPProfileCard`, `MPMenuRow`, `MPText`, shared fields |
| Empty/loading/error/disabled states | partial ML-04/18/28/31/33/34/58/67 | `MPAsyncStates`, `MPStatusMessage`, `MPButton` |

## Remaining function/state register

Every row below is deliberately open. This register and the observed ledger
together define survey coverage; no percentage is claimed without a complete
discovered function inventory.

| Area | Unvisited or incomplete subfunctions | Reason / next safe action |
| --- | --- | --- |
| Connection cleanup | Switch built-in sample mode Off | Phone unlock, hierarchy read, persistent `Tắt` button; verify owner view returns |
| Overview | Week/month dashboard tabs; top-category drilldown; Insider help/category selection; bell notifications; balance-help disclosure | Safe navigation after unlock; trial/subscribe not submitted |
| Wallets | Type help dialogs; linked-wallet landing; sort UI; currency search; existing basic/credit detail/edit; savings scoped ledger/add-money; archive/member actions | Inspect draft/view controls; no member invite/archive/link/bank write |
| Ledger | Explicit day/quarter/year/all/custom period selection; future/previous page; wallet-specific ledger; multi-delete selection UI | Display/selection only; no row deletion or sync |
| Transactions | Expense/income picker variants/search; date previous/next and picker; report switch in ordinary draft; field error states; copy draft behavior | Safe unsaved draft actions; actual writes not needed |
| Location | Location editor and associated map branches | One actual app exit; no screen observed. Investigate only if owner wants reference-app diagnosis |
| Search | Filter-specific result validation, wallet picker, note/With filters, successful narrow result | Compatibility warning/sample broad-result behavior limits evidence; no app update/account reset |
| Categories | Root editor versus child; merge selector/confirmation; search matched hierarchy; inactive list with actual inactive row | No merge/delete/activation; existing child and system read-only states already observed |
| Recurring/bills | Populated rows/detail/edit; frequency weekly/monthly specifics; start-date selection; non-default interval/count; validation | Current real/sample lists empty; owner now permits isolated demo fixtures if needed, but none created yet |
| Budgets | Sample Edit; quarterly/weekly/yearly list state; ended populated detail; category breakdown drill; wallet picker; help dialogs | Read sample draft/detail controls, no save/delete |
| Reports | Income category drill; debt/lending/other lower panels; calendar period choices; graph touch/tooltips; empty period | Safe read-only drills; net/expense/income/trend/comparison already observed |
| Events | Account events list/detail/ended states | Embedded picker/add-event draft observed; standalone list not yet visited |
| Debt | Wallet scope; repayment draft/confirmation | Person list/detail observed; Receive/payment not executed |
| Tools | ATM finder/bank finder; interest calculator; income-tax calculator | Individual rows inventoried; tip form/wheels/result complete for default input only |
| Travel | Landing and currency/date selectors | No travel mode enabled |
| Export | Export setup/Google account gate | No export or external account permission granted |
| Commercial | Store/catalog, Discover, linked-bank info, subscription info | Read informational navigation only; no purchase/trial/bank connection |
| Account | Settings menu and read-only subselectors; About; Support/help content | Avoid secrets/auth edits; no messages, logout, reset or account mutation |

## MyPocket extension preservation register

Reference absence never authorizes deleting a MyPocket capability. Keep these
existing destinations/flows reachable through shared native navigation and
presentations even if no Money Lover counterpart is observed:

- Finance Assistant conversation, structured cards, owner confirmation,
  cancellation/history/reset and idle-expiry behavior.
- AI transaction entry with Photos/Files/PDF, extraction proposals and explicit
  approval/rejection.
- Owner-scoped encrypted offline read cache and pending financial queue,
  rejected/conflict editor, retry and discard.
- Jars, month notes, travel workflows, portfolio assets/trades/price updates,
  and each existing report/credit/savings capability.
- Account timezone, Face ID enrollment/lock/logout, API-key lifecycle,
  notifications, feedback attachments/resolution/changelog.

All source navigation is listed in the
[native screen registry](../platforms/ios-screen-registry.md). Extended screens
reuse the same observed header, row, selection and editor families while
retaining their own data contracts. Do not invent reference financial rules.

## Per-screen evidence record

Use one record for each actual visited screen or distinct state. A function
with an observed landing page but an unvisited editor is only partially visited.

```text
Screen/state ID:
Observed app/version and observation time:
Entry control and previous screen:
Visible title and navigation/back/dismiss behavior:
Regions in top-to-bottom order:
  header / summary / period or filter / list / footer / bottom navigation
Row anatomy:
  icon / title / optional context / signed amount / disclosure or action
Selection, empty, disabled, scroll and keyboard states actually seen:
Draft/picker actions exercised without save:
Shared MyPocket bases mapped by region/action/state:
Missing shared-base variant or behavior:
Privacy-safe screenshot path, or why no screenshot was retained:
Observed deviations in equivalent MyPocket render:
Unvisited branches and why:
```

For visual comparisons record screen dimensions, safe-area relationship,
header/footer placement, horizontal alignment, grouping/divider rhythm,
typographic hierarchy, icon size/tone, selected-state treatment and tap target
geometry. Describe measurements as approximate unless extracted from actual
element bounds. App screenshot pixels alone do not prove point sizes.

## Actionable parity backlog

| Priority | Next action | Acceptance evidence |
| --- | --- | --- |
| Resolved | Dedicated installer signed separate helper | Independent identity, helper status and actual hierarchy verified |
| P0 | Unlock phone and restore owner data view with `Tắt` after remaining sample inspection | Read-only lock-state check currently requires owner passcode; verify sample disclaimer gone |
| P1 | Complete the remaining function/state register | Per-branch action paths, actual hierarchy/screenshot and explicit access gates |
| P1 | Finish wallet-specific savings/credit, settings/travel and remaining tool branches | Entry/back/dismiss plus actual sample/detail/draft states; no real financial saves |
| P1 | Compare MyPocket rendered equivalents | Same device-size/layout/state comparison; list deviations rather than claiming source-only parity |
| P2 | Reconcile native registry reference authority and extension inventory | Owner's new visual-reference scope recorded without changing business contracts |

## Docs reconciliation

Internal: this evidence note is the human- and agent-readable survey surface.
Device setup results belong to the linked tooling probe, owned by the parent.
The parent owns context/backlog/validation reconciliation to avoid concurrent
edits. No public/API/config/schema change is made by this note; public docs and
API/llms surfaces require no update yet. Product visual acceptance remains open.

Docs review/checkpoint: `rtk git diff --check` passes after reconciliation.
This doc-only survey changes no product/runtime code; product unit/integration
tests are not required. Platform evidence consists of actual `devicectl`
discovery/lock state, Mobile MCP/CLI hierarchy and 38 local screenshot files.
No image containing an unrelated notification banner is retained as reference.
