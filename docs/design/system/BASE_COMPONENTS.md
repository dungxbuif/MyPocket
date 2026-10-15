# Base component contracts

Photos-first implementation (in review): `MPDocumentPicker` retains Files-only
URL callback compatibility; optional `onPhotoImport([MPPickedImage])` opens the
shared source sheet, default Ảnh and explicit Tệp. Each source button anchors its
own system presentation; no screen-specific upload control. `MPPickedImage`
contains normalized name/Data/MIME; `MPDocumentImagePolicy` preserves valid
JPEG/PNG <=5MiB, converts other supported image representations to bounded JPEG
(orientation applied, <=4096px), rejects invalid/>25MiB sources. Sequential photo
loads append successful items and show per-item errors; no existing-selection
replacement. Count is remaining20-currentCount in AI entry; zero capacity cannot
open a picker. Source switch/capacity and actual system Photos/Files cancel pass
on SE in189/189 gate; screenshots inspected. An intentionally absent sheet footer
must be bounded (`Color.clear.frame(height:0)`), not an unbounded EmptyView plus
safe-area background, which covered rendered fields and swallowed touches.
Real selected-photo append now passes: stock Photos selection adds one normalized
image while retaining all19 existing files and reaches20 capacity. Shared
`MPDocumentImagePolicy.loadBatch(count:capacity:load:)` is the same sequential
pipeline used by PhotosPicker; data-boundary tests verify valid/throwing/invalid
items, continued results, ordering, one active load and zero-capacity no-load.
Mixed normalized JPEG/PDF multipart bytes/session/text/timezone are verified at
the HTTP transport boundary. This is not a live iCloud/OCR test or device UAT.

Typed native fields: `MPNumberField(title:text:kind:)` reuses `MPTextField` and
UIKit keyboard/focus/clear/accessory styling. Decimal/integer/percentage choose
decimalPad/numberPad and syntax feedback; raw drafts are never stripped.
`MPNumberInputKind.normalized` emits ASCII decimal strings (comma→dot), rejects
mixed separators/signs/exponents/non-ASCII numerals. Integer rejects fractional
or money grouping syntax; percentage checks0…100 without floating rounding.
Feature validators retain positivity/scale rules; money stays `MPAmountField`.

`MPEditorScaffold(title:isBusy:onCancel:fields:footer:)` owns long-form header,
one scroll area, keyboard-safe pinned actions and pending edit/dismiss prevention.
Features own drafts, requests and errors; full-screen wrapper owns presentation.
Short manual-price/choice forms use shared form sheet instead.
The typed `init(isBusy:header:fields:footer:)` allows AI-entry's existing custom
header/Save-all without AnyView or a separate editor hierarchy. Header min52pt
preserves44pt controls with breathing room and grows with content. The shell
uses a stable lower-right Feedback dock; it does not move between rows or sides
based on obstacle reports and does not reserve a scroll lane. Footer clearance
keeps the dock above the visible composer/navigation. The old obstacle-report
plumbing remains source-compatible for screen consumers, but is no longer used
to reposition the shared bubble.
The footer reports its actual height through `MPFloatingActionClearanceKey` to
keep floating actions clear, without a reserved lane. The shell observes keyboard
visibility with `mpObserveKeyboardVisibility`; Feedback temporarily hides while
typing and restores on dismissal without changing draft state.

`MPOptionalDateField(title:value:timeZone:accessibilityIdentifier:)` binds an
optional calendar date as an empty/`YYYY-MM-DD` string. Enabling defaults to
account-local today, disabling clears it. It composes the existing checkbox/date
field and shared calendar, not a free-text date input. Wire semantics remain a
feature/API-adapter responsibility.

Native chat consolidation: `MPChatComposer(text:isBusy:isEnabled:onSend:onCancel:)`
owns1–4-line input and44pt send/stop. Service unavailable permits drafting but
never submitting. `MPChatMessage` gives user prose a trailing bubble and leaves
assistant parts flat. `MPChatOverview` pins exact tool-derived income/expense;
details use shared form sheet. `MPFloatingActionClearanceKey` reports actual
composer height to move Feedback, not reserve a scroll lane. Raised Add hidden
on Assistant/Account; Overview/Transactions retain it. See native screen spec.

Current surface inset override: horizontal12pt, vertical8pt. Other geometry
below is historical where it conflicts; rollout verification remains in progress.

Native density override — 2026-10-14: card padding8pt, control/menu/category
minimum44pt, transaction/wallet glyph32pt with no outer row padding, page
section gap8pt/row gap4pt. Profile is one tappable identity row (avatar36), not
identity plus a second management row. These supersede the older geometry in
the inventory below; button corners8/card corners12 and typography unchanged.
Account-tab disabled center add action is fully hidden. Dynamic Type remains
content-driven; no maximum row height. Urgent test execution skipped by owner;
compile/signature/install checks still required. No backend/API changes.

[Gateway](../README.md) · [Tokens](TOKENS.md) · [Inventory](../INDEX.md) · [Work item](../../work/tickets/UI-BASE-01-DETAIL_DESIGN.md)

## Atoms đang dùng

| Base | Props/variants | Behavior và ownership |
| --- | --- | --- |
| SurfaceCard | padding none/sm/md/lg; radius sm/md/lg hiện cùng 12px; elevation flat/subtle/raised; tone default/muted/danger | Một background/border/radius/elevation; forward section role/aria; screen không override shape |
| BaseButton | primary/secondary/ghost/danger/outline/row/key/chip; size sm/md/lg/row; loadingLabel | min 44px; type mặc định button; loading → disabled + aria-busy; inline icon/label; CTA pill, row control radius |
| IconButton | surface/bare; shape circle/control; selected | accessible label bắt buộc; disabled; selected ring tại base; min 44px |
| BaseLink | to/label/content | Một link tương tác duy nhất; không bọc button trong anchor |
| BaseNavigationItem / BaseFab | active / children, action | aria-current, named selected color; FAB 56px; consumer chỉ quyết định vị trí |
| Text / Heading | semantic tag, size, weight, tone; numeric | Manrope, type scale và màu chung; className chỉ layout |
| FormField / BaseTextInput / BaseTextArea / BaseSelect | native props; default/inline/title/composer | label association; focus còn thấy; disabled forwarding; textarea hỗ trợ multiline composer theo token; select chevron không ăn click |
| MPMultiOptionSelector (SwiftUI) | `options`, `Binding<Set<String>>`, `allLabel` | Generic shared multi-option menu; category applicability uses the visible `MPWalletApplicabilityCard` instead, with stable checkbox-row IDs and empty-set all-wallet semantics |
| MPButton (SwiftUI) | primary/outline/secondary/ghost/segmentedSelected/key/destructive/plain, loading, optional accessibility identifier | Shared controls expose their full rendered hit shape and stable underlying Button identifier. Native rectangular actions use the single 8pt button-radius token (owner refinement 2026-10-13), 48pt control height and minimum 44pt hit target; card/circular-icon geometry is independent. Primary uses dark action green with white text; outline uses surface/action, secondary success-soft/action, ghost transparent/action, key muted row, segmentedSelected white surface. Plain rows add no horizontal padding. No screen-local shape overrides. |
| MPTextField / MPSecureField / MPAmountField (SwiftUI) | title, placeholder, binding, error, accessibility identifier | Shared fields own the web card/line/accent focus treatment, 48pt control height and 12pt horizontal padding; `MPTextField` uses a UIKit-backed delegate binding plus a shared clear action so editing cannot lose characters or restore a stale buffer; `MPAmountField` exposes the shared integer-safe `MPAmountKeypad`; screens do not apply `.textFieldStyle` or literal borders |
| MPFormRow / MPFormRowLabel / MPOptionSelector / MPDateStepper / MPDateField / MPDateTimeField / MPDateCalendarSheet (SwiftUI) | icon/title/value row, enabled/disabled state, stable ID; account-local date/date+time with previous/next actions and Monday-first calendar confirmation | Transaction-editor rows own the Stitch icon/value/chevron geometry and 44pt hit target; disabled option selectors are non-interactive and omit the disclosure chevron; date-only and date+time fields open the same account-timezone calendar surface, with an opt-in shared hour/minute section; consumers supply only selection/date actions |
| MPFont (SwiftUI) | display/headline/body/label/currency roles | Owns the measured Stitch type scale for native base components; screens do not invent per-screen font sizes |
| MPText (SwiftUI) | title/body/secondary/emphasis/action/danger roles | Owns semantic typography color roles for native feature copy; screens choose a semantic role instead of applying local foreground/font styling |
| MPOverflowMenu (SwiftUI) | accessible label, `[MPOverflowAction]` | Shared three-dot screen action trigger owns the 44px hit target and accessible label; consumers supply only semantic action titles/closures |
| MPMenuRow (SwiftUI) | title, optional subtitle, SF Symbol, stable accessibility ID, action | Shared 48pt secondary-navigation row owns padding, chevron, hit shape and accessibility; More/account consumers do not create raw navigation buttons |
| MPProfileCard / MPHeaderSupportButton (SwiftUI) | identity, email, account subtitle, manage action / support action | Shared compact horizontal 48pt avatar/name/email identity, account-management row, support action with 8pt button corners and stable route IDs; screens provide session data and callbacks only |
| MPReferenceScreenHeader / MPActionIconButton / MPFullScreenCover (SwiftUI) | form title/close action; compact filled image action; boolean presentation binding + content | Shared Stitch form header owns the close capsule and centered title; compact footer actions use one 56pt circular geometry; reference editors use the base modal wrapper so shell navigation cannot occlude save actions |
| MPCenteredScreenHeader / MPHeaderIconButton / MPTextLinkButton / MPWalletGlyph / MPWalletAggregateGlyph / MPSavingsSummary (SwiftUI) | centered screen title; compact circular header action; report/navigation text action; typed wallet identity badge; distinct teal globe for aggregate scope; goal balance/remaining/deadline/progress in `card` or `ledger` layout | Shared headers keep title and icon hit targets stable; individual wallet glyphs retain basic/goal/credit semantic tones and aggregate uses a teal globe; savings calculations and deadline copy stay in one base for explicit detail and embedded goal-ledger surfaces; all displayed money values accept shared privacy-mask state |
| MPMenuSelector / MPInfoAlert (SwiftUI) | custom visual menu label + stable option IDs; informational alert binding/title/message | Goal-wallet scope keeps menu actions and help feedback inside shared bases; feature screens provide only options and state callbacks |
| MPWalletScopeSelector / MPWalletScopeGlyph (SwiftUI) | wallet list + selected wallet ID | Ordinary ledger owns the centered, compact scope capsule and wallet-type icon tone; the selector uses neutral slate for aggregate/basic, success for goal and teal for credit, while the consumer retains scope/filter state |
| MPWalletSelectionList / MPWalletSelectionRow / MPWalletManagementPolicy / MPWalletManagementRow / MPDivider (SwiftUI) | wallet list, editing state and semantic callbacks | Owns the “Tổng cộng” aggregate, `is_in_total` groups, Stitch title-over-balance row anatomy, selected-scope checkmarks, narrow-phone-safe geometry, compact trailing detail action, goal/credit supporting values, edit affordances and add/service actions; total excludes non-included balances; callbacks remain in WalletsView; composes MPButton, MPTextLinkButton, MPWalletGlyph, MPSurfaceCard, MPText, MPProgressBar and shared divider |
| MPSegmentedControl (SwiftUI) | option IDs/titles, bound selection, optional accessibility ID | Shared selection changes animate briefly and honor Reduce Motion; consumers do not implement their own selection transition |
| MPCheckboxRow / MPWalletApplicabilityCard (SwiftUI) | title, selected value, stable ID/action / owner wallet list + selected wallet ID binding | Shared 44pt accessible checkbox rows and visible category-to-wallet applicability card; empty set means all wallets and full-set normalization returns to empty |
| MPCategoryTree / MPCategoryTreeItem (SwiftUI) | stable category nodes with depth, hierarchy-context and branch flags; select callback | Owns shared root/child rows, branch connector, icon/title/subtitle/chevron geometry and lazy rendering; editable rows navigate as one target, hierarchy context stays disabled, and delete remains in the category editor; `CategoryListPresentation` supplies filtered nodes while feature screens own navigation and API calls |
| MPCategoryPicker / MPCategoryTree (SwiftUI) | category candidates, kind, optional eligible-ID set, selected ID, optional clear action | One tree owns management and selection row geometry, localized hierarchy-preserving search, selectable-row check, disabled parent context, one shared surface per root and stable IDs; root/child glyphs 40/32pt, connectors terminate before the icon; host owns sheet lifecycle and payload |
| MPFeedbackBubble / MPBottomNavigationInset (SwiftUI) | bubble action; inset has no consumer props | Bubble is a 52pt stable lower-right overlay with an 8pt bottom margin inside the full scroll viewport; no dedicated lane, spacer or obstacle-driven jumping. MainTabView explicitly sizes route content above its 56pt sibling navbar, so nested NavigationStack scroll regions cannot extend behind it. Footer/composer clearance keeps the dock above protected chrome; the bubble hides before capture and during feedback presentation without resizing content. MPBottomNavigationInset is only for independent editors/modals, not shell routes. |
| MPSurfaceCard / MPMetric / MPMetricCard (SwiftUI) | shared surface style alias; semantic metric title/value/detail/masked | All card styles share 12pt padding, 12pt radius and flat surface; children are accessibility-contained. Metrics share one presentation; screens do not build competing cards/metric fonts. |
| MPTransactionRow / MPTransactionPresentation (SwiftUI) | persisted transaction/category or validated structured fields, timezone, masked | Canonical semantic mapping for category/icon, note-or-local-date, signed/masked amount and neutral transfer/adjustment; used by Overview, ledger, wallet/credit history and Assistant. Default glyph 36pt, 4pt vertical padding, min44pt content height; amounts stack for accessibility text sizes. |
| MPImageAttachmentPicker / MPAttachmentThumb (SwiftUI) | optional bound image bytes | Shared choose/remove/replace image, loading/error states and thumbnail; source limit25MiB, bounded decoding, conversion to PNG ≤1MiB for existing feedback multipart contract. Both feedback entry points reuse this base. |
| BaseFileUpload | field/button | Chọn nhiều JPEG/PNG/PDF; giới hạn và xử lý tệp do consumer cung cấp; button variant kích hoạt input ẩn có nhãn truy cập được |
| BaseCheckbox | checked/disabled/label/onChange | controlled boolean, native keyboard, hàng min 44px |
| IconBadge | size xs/sm/md/lg; rounded/circle; tone | Một tone map, không chồng class màu. Icon visual size không đồng nghĩa hit target |
| Progress | value/label/danger | clamp thanh 0–100, aria-valuenow, rounded cap; nhãn/số tiền bên ngoài giữ giá trị nghiệp vụ thật |
| BudgetGauge / MPBudgetGauge | value/label | SVG/SwiftUI arc 180°, clamp arc, label truy cập được; value >100 dùng danger; summary consumes real spent/limit data |
| StatusMessage / MPStatusMessage | muted/danger; variant card/plain | card mặc định compose SurfaceCard + Text; plain chỉ Text, không nền/viền/bóng; giữ role status/alert. Transaction empty states ở Tổng quan và Sổ GD dùng plain; không có tương tác/keyboard riêng. |
| StatusBadge / MPStatusBadge | label; success/muted/danger | compact semantic status/percentage pill with stable status accessibility; caller resolves business state and copy |
| Divider | layout className | line token chung |
| Chip | children/layout className | passive pill label; use BaseButton chip for interactive actions |

## Molecules

Wallet empty states in Overview (“Chưa có ví.”) and WalletManagementPanel (“Chưa có ví nào.”) also use `StatusMessage variant="plain"`: semantic status text without a separate card border, background or shadow. Loading/error feedback retains its existing presentation.

| Base | Composition và behavior | Chi tiết |
| --- | --- | --- |
| BaseBottomSheet | Heading + IconButton, modal shell; focus trap, Escape, backdrop, return focus, scroll lock | [Behavior](BEHAVIOR.md) |
| BaseCategoryTree | SurfaceCard + BaseButton row + IconBadge + Text; root/children/trailing content; navigation/selection mode | [Tree](../molecules/category-tree/README.md) |
| WalletCard / TransactionItem | shared row/button, badge và numeric Text; domain tone maps | [Cards](../atoms/base-cards/README.md) |
| FormSelectorRow / InlineControlRow | leading/content/trailing slots; native input chỉ qua base | [Rows](../molecules/transaction-form-rows/README.md) |
| CategoryEditForm / ApplicableWalletsCard | base fields, icon picker, shared checkbox rows, status/save | [Edit group](../molecules/edit-group/README.md) |
| BudgetProgressItem / GoalCard | SurfaceCard + IconBadge + Text + Progress | [Budget](../molecules/budget-progress-cards/README.md) |
| DateField / AmountField | BaseButton, FormSelectorRow, BaseCalendar, BaseModal | date stepping/calendar and integer-safe calculator shared by transaction and budget forms; native `MPAmountField` supports `standard` and Stitch `transaction` presentations, where VND precedes a large unboxed value and display grouping never changes bound integer digits |
| LedgerPeriodSelector | SegmentedControl + BaseButton + Text | Ledger Week/Custom tabs for aggregate and basic scopes; inclusive account-local week label and previous/next controls; consumer owns filter state and date dialog. Goal and credit statement panels hide this control. |
| PageBackHeader | BaseLink + Heading + trailing slot | một tầng interactive |
| AccountMenuRow / ProfileHeroCard | shared identity/navigation presentation | screen-specific actions thuộc consumer |
| BaseBarChart / BaseDonutChart / MPDonutChart | values/shares/label; shared visual geometry and theme-based data colors; native period-comparison ring keeps the actual percentage while clamping stroke geometry | BaseBarChart remains mounted for reports with real owner-scoped summary values; BaseDonutChart is mounted in the web Money Insider card and MPDonutChart in native; tooltip/drilldown is not implemented |
| BaseTrendChart / MPTrendChart | account-timezone daily income/expense points, separate semantic strokes, empty-safe rendering and stable accessibility contract; native exposes chronological daily totals and honors amount masking | mounted in web and native Reports trend mode; values come from owner-scoped transactions and exclude non-report/transfer/adjustment rows |
| AssistantComposer / AssistantResultCard | SurfaceCard, BaseTextArea, BaseFileUpload, BaseButton, Text/Heading | vùng soạn một yêu cầu nhiều dòng có thể kèm tệp; thẻ kết quả AI chỉ trình bày nhóm đề xuất và action từ consumer, không sở hữu hội thoại/history hoặc lưu giao dịch |
| MPAssistantCard / AssistantCardType | elevated MPSurfaceCard reference surface, MPText, MPMetricCard, MPTransactionRow, MPProgressBar | Native structured-card shell uses the shared reference elevation; alias mapping and unknown-kind fallback belong to the renderer, while action mutation is allowed only through consumer confirmation |

## Contract gate

WalletSelectionList: aggregate excludes is_in_total=false; grouped rows stack wallet title over balance and show the active aggregate/wallet check; edit mode suppresses selection and sends rows to the editor; callbacks own navigation. SavingsSummary: clamped Progress with real remaining/reached state, calendar-day countdown. CategoryTreeSelector: BaseTextInput search, clear action and BaseCategoryTree selection mode; retains a non-applicable parent as context when an applicable child exists, and only selectable IDs invoke onSelect. Transaction, proposal and budget pickers share this tree used in group management.

Wallet follow-up: BaseSwitch owns the native checkbox role=switch, 44px target, checked/focus/disabled visuals and controlled onChange. BaseBottomSheet `presentation=form` owns canvas/tall shell, cancel/title/headerAction; same focus trap, Escape and restoration. WalletCreateForm composes shared cards/inline controls/switch; WalletTypePicker composes row buttons with aria-pressed selection. See [wallet contract](../pages/wallets/README.md).

Trước một biến thể mới, ghi: intent, anatomy, props, event/effect, states, keyboard, tokens, consumers, proof. Sau đó implement base và kiểm thử consumer. API lớp UI có thể mở rộng theo task đã duyệt; không tự tạo palette/interaction model riêng.

Ảnh chỉ chứng minh trạng thái nhìn thấy; trạng thái chưa có trong code được đánh planned/partial. Không copy tiêu đề “Production” trên export thành trạng thái implementation.
