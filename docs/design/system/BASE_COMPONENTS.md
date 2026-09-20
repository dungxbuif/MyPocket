# Base component contracts

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
| FormField / BaseTextInput / BaseSelect | native props; default/inline/title | label association; focus còn thấy; disabled forwarding; select chevron không ăn click |
| BaseCheckbox | checked/disabled/label/onChange | controlled boolean, native keyboard, hàng min 44px |
| IconBadge | size xs/sm/md/lg; rounded/circle; tone | Một tone map, không chồng class màu. Icon visual size không đồng nghĩa hit target |
| Progress | value/label/danger | clamp thanh 0–100, aria-valuenow, rounded cap; nhãn/số tiền bên ngoài giữ giá trị nghiệp vụ thật |
| BudgetGauge | value/label | SVG arc 180°, clamp arc, label truy cập được; value >100 dùng danger |
| StatusMessage | muted/danger; variant card/plain | card mặc định compose SurfaceCard + Text; plain chỉ Text, không nền/viền/bóng; giữ role status/alert. Transaction empty states ở Tổng quan và Sổ GD dùng plain; không có tương tác/keyboard riêng. |
| Divider | layout className | line token chung |
| Chip | children/layout className | passive pill label; use BaseButton chip for interactive actions |

## Molecules

Wallet empty states in Overview (“Chưa có ví.”) and WalletManagementPanel (“Chưa có ví nào.”) also use `StatusMessage variant="plain"`: semantic status text without a separate card border, background or shadow. Loading/error feedback retains its existing presentation.

| Base | Composition và behavior | Chi tiết |
| --- | --- | --- |
| BaseBottomSheet | Heading + IconButton, modal shell; focus trap, Escape, backdrop, return focus, scroll lock | [Behavior](BEHAVIOR.md) |
| BaseCategoryTree | SurfaceCard + BaseButton row + IconBadge + Text; root/children/trailing content | [Tree](../molecules/category-tree/README.md) |
| WalletCard / TransactionItem | shared row/button, badge và numeric Text; domain tone maps | [Cards](../atoms/base-cards/README.md) |
| FormSelectorRow / InlineControlRow | leading/content/trailing slots; native input chỉ qua base | [Rows](../molecules/transaction-form-rows/README.md) |
| CategoryEditForm / ApplicableWalletsCard | base fields, icon picker, checkbox, status/save | [Edit group](../molecules/edit-group/README.md) |
| BudgetProgressItem / GoalCard | SurfaceCard + IconBadge + Text + Progress | [Budget](../molecules/budget-progress-cards/README.md) |
| PageBackHeader | BaseLink + Heading + trailing slot | một tầng interactive |
| AccountMenuRow / ProfileHeroCard | shared identity/navigation presentation | screen-specific actions thuộc consumer |
| BaseBarChart / BaseDonutChart | values/shares/label; shared visual geometry and theme-based data colors | source preview only; tooltip/drilldown not implemented |

## Contract gate

WalletSelectionList: controlled selectedID/editing, aggregate excludes is_in_total=false, grouped shared rows and selection marks; callbacks own navigation. SavingsSummary: clamped Progress with real remaining/reached state, calendar-day countdown. CategorySelectionList: BaseTextInput search and base rows of API-supplied applicable categories; onSelect returns category ID. These molecules reuse controls and tokens, not custom screen styling.

Wallet follow-up: BaseSwitch owns the native checkbox role=switch, 44px target, checked/focus/disabled visuals and controlled onChange. BaseBottomSheet `presentation=form` owns canvas/tall shell, cancel/title/headerAction; same focus trap, Escape and restoration. WalletCreateForm composes shared cards/inline controls/switch; WalletTypePicker composes row buttons with aria-pressed selection. See [wallet contract](../screens/wallets/README.md).

Trước một biến thể mới, ghi: intent, anatomy, props, event/effect, states, keyboard, tokens, consumers, proof. Sau đó implement base và kiểm thử consumer. API lớp UI có thể mở rộng theo task đã duyệt; không tự tạo palette/interaction model riêng.

Ảnh chỉ chứng minh trạng thái nhìn thấy; trạng thái chưa có trong code được đánh planned/partial. Không copy tiêu đề “Production” trên export thành trạng thái implementation.
