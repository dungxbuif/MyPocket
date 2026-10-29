# MyPocket Mobile UI/UX Design Specification & Component Contract

Source assets: `design/IMG_7696.PNG` through `design/IMG_7731.PNG`.

This document is the authoritative specification for the MyPocket mobile web/PWA interface, detailing the visual design tokens, reusable base components, hierarchical screen architecture (Parent – Child – Modals/Sheets), and individual screen breakdowns derived directly from the 36 production UI screenshots.

> - For the detailed TypeScript props interfaces, component CSS rules, and state contracts, see [BASE_COMPONENTS.md](file:///Users/dungxbuif/workspace/MyPocket/design/BASE_COMPONENTS.md).
> - For the exhaustive file-by-file visual description, DOM anatomy, and CSS rules designed to build components without images, see [COMPONENT_FILES_SPEC.md](file:///Users/dungxbuif/workspace/MyPocket/design/COMPONENT_FILES_SPEC.md).

---

## I. Design Tokens & Visual Guidelines

### 1. Color Palette Tokens

| Token Name | Hex Code | Visual Application & Context |
|---|---|---|
| **App Canvas / Background** | `#f2f3f8` | Global application viewport canvas and page backdrop |
| **Surface / Card Background** | `#ffffff` | Grouped cards, bottom sheets, floating navigation bar, dialog cards |
| **Text Primary** | `#111111` | Headings, hero amounts, wallet names, active category labels |
| **Text Secondary / Muted** | `#8e8e93` | Subtitles, input placeholders, timestamp labels, wallet association count ("Hoạt động trong 3 ví") |
| **Divider / Border** | `#e8e8ec` | Subtle row separators within grouped cards, chart baseline axes |
| **Primary Green** | `#2dbd4f` | Primary CTA pills (`Lưu`, `Tạo Ngân sách`, `Đăng ký ngay`), active navigation tabs, positive balances |
| **Soft Green** | `#e8f7ed` | Secondary button backgrounds, light accents, safe budget threshold indicators |
| **Expense Red / Destructive** | `#ff5a66` | Expense amounts, spending comparison bars, growth metrics, destructive buttons (`Đăng xuất`, `Xóa tài khoản`) |
| **Income Blue** | `#32a9df` | Income metrics, "Tổng thu" labels |
| **Icon Neutral Dark** | `#29495a` | Default circular category and wallet icon container backgrounds |
| **Warning / Premium Accent** | `#ff8800` | `TÀI KHOẢN PREMIUM` ribbon banner, alert badges |
| **Lock Gray** | `#8e8e93` | System default category lock icon 🔒 indicator |

### 2. Typography & Layout Rules
- **Typography Stack**: Native iOS system font stack (`-apple-system`, `BlinkMacSystemFont`, `San Francisco`, `Helvetica Neue`, `sans-serif`).
  - Hero Numbers: `32px` – `36px` Bold, tabular numbers (`font-variant-numeric: tabular-nums`).
  - Section Titles: `18px` – `22px` Bold.
  - Body & Row Labels: `15px` – `17px` Regular / Medium.
  - Subtext & Micro-labels: `12px` – `13px` Regular.
- **Corner Radii**:
  - Container Cards / Bottom Sheets: `24px` to `28px`.
  - Buttons, Search Chips, Filter Pills: Full pill (`9999px`).
  - Circular Icon Buttons: `50%` circle.
- **Spacing & Padding**:
  - Screen Horizontal Inset: `16px` to `18px`.
  - Grouped Card Vertical Gap: `14px` to `20px`.
  - Card Internal Padding: `14px` to `16px`.
- **Bottom Safe Area**: The floating navigation bar is pinned above the iOS Home Indicator and maintains a consistent backdrop blur/elevation over scrolling content.

---

## II. Reusable Base Components Specification

### 1. Navigation & App Bar Components

#### `BottomTabBar` *(IMG_7696, IMG_7706, IMG_7710, IMG_7714)*
- **Description**: Persistent, floating 5-slot bottom navigation bar with a raised center quick-action button.
- **Slots & Destinations**:
  - Tab 1: `Tổng quan` (Home icon)
  - Tab 2: `Sổ giao dịch` (Wallet icon)
  - Center: `FloatingAddButton` (Raised green circular `+` button)
  - Tab 3: `Ngân sách` (Budget/Ledger icon)
  - Tab 4: `Tài khoản` (Profile/User icon)
- **Visual States**:
  - `Active`: Surrounds the active tab icon and label with an elongated soft gray pill background.
  - `Inactive`: Muted gray icon and label.

#### `SubpageHeader` *(IMG_7697, IMG_7715, IMG_7716, IMG_7717, IMG_7728, IMG_7729)*
- **Description**: Standard header component for all secondary and tertiary push screens.
- **Properties / Slots**:
  - `LeftAction`: Pill button containing `< Quay lại` (`Back`) icon and label.
  - `Title`: Bold, centered screen title.
  - `RightAction` (optional): Action icon button (e.g., Search `🔍`, Tip lightbulb `💡`, Menu `≡`, or text action `Thêm`).
  - `CenterFilterChip` (optional): Interactive wallet selector chip (`🌐`) positioned next to or below the title.

#### `ModalHeader` *(IMG_7703, IMG_7705, IMG_7707, IMG_7712, IMG_7731)*
- **Description**: Standardized header for bottom sheets and overlay modals.
- **Properties / Slots**:
  - `LeftButton`: Dismiss pill button labeled `Huỷ` (`Cancel`) or `Đóng` (`Close`).
  - `Title`: Centered modal title (`Thêm Giao Dịch`, `Chọn Ví`, `Chi tiết khoản chi`, etc.).
  - `RightButton` (optional): Text action (`Xong`, `Sửa`) or utility icon (Calendar button).

#### `WalletFilterChip` *(IMG_7697, IMG_7706, IMG_7710, IMG_7717, IMG_7728)*
- **Description**: Compact rounded pill chip displaying the current wallet scope.
- **Display**: Leading globe icon `🌐` or specific wallet logo, wallet name (default: `Tổng cộng`), and subtle up/down selector arrows (`▲▼`).

---

### 2. Grouped Containers & Form Controls

#### `GroupedCard` *(IMG_7696, IMG_7705, IMG_7707, IMG_7713, IMG_7729)*
- **Description**: White card container aggregating rows of related settings, metrics, or navigation entries.
- **Attributes**:
  - Surface color `#ffffff`, border radius `24px` to `28px`.
  - Optional uppercase section label above the card (`TÍNH VÀO TỔNG`, `HIỂN THỊ`, `HỆ THỐNG`).
  - Internal rows are separated by thin `#e8e8ec` dividers, omitting the divider on the last row.

#### `FormRowItem` *(IMG_7707, IMG_7708, IMG_7713, IMG_7729)*
- **Description**: Generic interactive row for input collection, selection, or navigation.
- **Properties**:
  - `Leading`: Icon container, badge, or avatar.
  - `Label`: Field or menu item name (`Chọn nhóm`, `Ghi chú`, `Thiết bị`, `Cài đặt`).
  - `Value`: Currently selected or configured value aligned to the right.
  - `Trailing`: Chevron `>` icon indicating forward navigation.

#### `AmountInputHero` *(IMG_7707, IMG_7708, IMG_7712)*
- **Description**: Hero-scale numeric input component for transaction and budget creation forms.
- **Structure**:
  - Leading `VND` currency badge pill.
  - Large tabular numeric input display (`32px`–`36px` bold font), defaulting to `0`.

#### `DateNavigationRow` *(IMG_7707, IMG_7708)*
- **Description**: Inline date selection bar featuring quick one-day increment/decrement controls.
- **Structure**:
  - Leading calendar glyph icon.
  - Previous day button (`<`).
  - Centered date badge pill displaying full localized date: `Chủ Nhật, 23/08/2026`.
  - Next day button (`>`).

#### `SegmentedControl` *(IMG_7696, IMG_7703, IMG_7707, IMG_7717)*
- **Description**: Pill-shaped horizontal selector for switching view modes or entity classifications.
- **Variants**:
  - 2-Segment: `Tuần` | `Tháng` (Dashboard), `Chi tiết` | `Xu hướng` (Expense Breakdown).
  - 3-Segment: `Khoản chi` | `Khoản thu` | `Vay/Nợ` (Transaction and Category management).
- **Styling**: Gray background container track with a sliding white pill representing the selected segment.

#### `SwitchRow` *(IMG_7708, IMG_7712, IMG_7729, IMG_7731)*
- **Description**: Toggle setting row combining label, description, and an iOS-style boolean switch.
- **Structure**:
  - Primary title (e.g., `Không tính vào báo cáo`, `Lặp lại ngân sách này`, `Bật Chế Độ Du Lịch`).
  - Multi-line explanatory subtext.
  - Right-aligned toggle switch (Green `#2dbd4f` when enabled, muted gray when disabled).

#### `RadioCheckItem` *(IMG_7705)*
- **Description**: Selectable row within a single-choice list.
- **Display**: When selected, displays an active colored checkmark (`✓`) at the right trailing edge.

---

### 3. Action Buttons & Feedback

#### `PrimaryButton` *(IMG_7696, IMG_7707, IMG_7710, IMG_7712)*
- **Description**: Primary call-to-action button spanning the card width or modal action bar.
- **Visuals**: Full pill border-radius, background `#2dbd4f`, bold white text. Switches to muted gray in `Disabled` state.

#### `SecondaryButton` *(IMG_7696)*
- **Description**: Supporting action button.
- **Visuals**: Full pill shape, soft green background (`#e8f7ed`), dark green text.

#### `DestructiveActionRow` *(IMG_7715)*
- **Description**: Danger-level action buttons or grouped list rows.
- **Visuals**: High-contrast red typography (`#ff5a66`) for irreversible operations (`Đăng xuất`, `Đặt lại tài khoản`, `Xóa tài khoản`).

#### `CircleIconButton` *(IMG_7696, IMG_7697, IMG_7706, IMG_7707)*
- **Description**: Circular touch targets enclosing icons.
- **Applications**: Back navigation, image attachment/camera trigger, search trigger, help icon `?`, and notification bell with unread badge.

---

### 4. Data Display & Visualizations

#### `TransactionRow` *(IMG_7702, IMG_7706)*
- **Description**: Item row representing an individual financial transaction.
- **Structure**:
  - Leading circular category icon with a mini wallet logo badge overlaid on the bottom-right corner.
  - Middle block: Category title (`Trả nợ`, `Ăn vặt`, `Cơm Bữa`) and date/payee details.
  - Right block: Formatted currency amount (red for expenses with negative sign, green/blue for income).

#### `WalletRow` *(IMG_7696, IMG_7705, IMG_7716)*
- **Description**: List row displaying wallet entity metadata and balance.
- **Structure**:
  - Leading circular icon (credit card, bank logo, cash wallet).
  - Wallet name (`Ví tín dụng`, `Techcombank`, `Tiền Mặt`).
  - Right-aligned balance in VND (supports negative amounts e.g., `-3.711.104 đ`).

#### `CategoryTreeRow` *(IMG_7717 -> IMG_7727)*
- **Description**: Hierarchical category row supporting parent-child indentation.
- **Structure**:
  - Parent categories appear at the card root.
  - Child categories are indented with branching tree lines connecting to the parent node.
  - Status subtext: `Hoạt động trong 3 ví`.
  - Lock icon 🔒 for non-editable system categories.
  - Trailing chevron `>`.

#### `GaugeArcSummary` *(IMG_7710, IMG_7711)*
- **Description**: Circular arc gauge visualizing available budget proportion.
- **Display**:
  - Colored progress arc with an active indicator dot.
  - Hero center text: `Số tiền bạn có thể chi 45.075.000 đ`.
  - 3-column key metrics: `Tổng ngân sách` | `Tổng đã chi` | `Đến cuối tháng`.

#### `ProgressBarWithMarker` *(IMG_7710, IMG_7711)*
- **Description**: Two-tone horizontal progress bar visualizing category budget exhaustion.
- **Display**: Colored elapsed fill, remaining budget track, and a sliding vertical badge labeled `Hôm nay` indicating current calendar day progress.

#### `ComparisonBarChart` *(IMG_7699, IMG_7701, IMG_7704)*
- **Description**: Comparative bar chart evaluating spending between adjacent periods or across individual days.
- **Display**: Light-shaded bar (previous period) vs solid red bar (current period) with baseline zero axis and top amount values.

#### `TrendAreaLineChart` *(IMG_7700)*
- **Description**: Cumulative expenditure curve plotted against historical baseline.
- **Display**: Bold red curve with milestone nodes, shaded baseline comparison region (3-month historical average), and interactive date tooltip.

#### `DonutChartBreakdown` *(IMG_7703)*
- **Description**: Donut chart illustrating percentage distribution of expenses across categories.
- **Display**: Segmented circular ring with dominant category icon and percentage badge (`100%`) positioned at the center bottom.

#### `EmptyStateView` *(IMG_7702, IMG_7728)*
- **Description**: Placeholder visual container displayed when entity collections are empty.
- **Display**: 3D vector illustration or friendly icon, primary bold status line, and supporting descriptive subtext.

#### `NoticeBanner` *(IMG_7710, IMG_7711)*
- **Description**: Full-width or anchored alert banner.
- **Display**: Blue bar warning about sample demonstration data with an inline `Tắt` dismiss button.

---

## III. Screen Hierarchy (Parent – Child – Modals – Popovers)

The application navigation architecture follows a strict tree hierarchy:

```
App Shell (Root Application Frame)
│
├── [Tab 1] Overview / Dashboard (`DashboardScreen`) [IMG_7696, 7699, 7700, 7701, 7702]
│    │
│    ├───> [Sub-screen] My Wallets (`WalletManagementScreen`) [IMG_7716] (via "Xem tất cả" in Wallets card)
│    │
│    ├───> [Sub-screen] Spending Trend Analytics (`SpendingTrendScreen`) [IMG_7697, 7698] (via "Xem báo cáo" / "Xem chi tiết")
│    │
│    ├───> [Bottom Sheet] Expense Detail Breakdown (`ExpenseDetailSheet`) [IMG_7703, 7704] (via Top Spending category tap)
│    │      ├── Tab 1: Category breakdown donut chart & percentages
│    │      └── Tab 2: Daily spending bar chart & daily logs
│    │
│    ├───> [Bottom Sheet] Wallet Scope Picker (`WalletPickerSheet`) [IMG_7705] (via Top Bar wallet selector)
│    │
│    └───> [Bottom Sheet] Transaction Detail (via recent transaction row tap)
│
├── [Tab 2] Transaction History (`TransactionListScreen`) [IMG_7706]
│    │
│    ├───> [Bottom Sheet] Wallet Scope Picker (`WalletPickerSheet`) [IMG_7705] (via "Tổng cộng" filter chip)
│    │
│    ├───> [Popover Menu] More Options Menu (`MoreOptionsMenuPopover`) [IMG_7706] (via "..." header button)
│    │      ├── Khoảng thời gian (Date range selector)
│    │      ├── Xem theo nhóm (Group by category)
│    │      ├── Xoá nhiều giao dịch (Bulk delete)
│    │      ├── Chuyển tiền đến ví khác (Transfer between wallets)
│    │      ├── Điều chỉnh số dư (Adjust balance)
│    │      └── Đồng bộ ví (Sync wallet)
│    │
│    └───> [Bottom Sheet] Transaction Detail (via individual list row tap)
│
├── [Action Center] Add Transaction Bottom Sheet (`AddTransactionSheet`) [IMG_7707, 7708]
│    │  (Triggered via raised floating `+` button in BottomTabBar)
│    │
│    ├───> [Bottom Sheet] Wallet Picker (`WalletPickerSheet`) [IMG_7705] (via wallet selection row)
│    │
│    ├───> [Sub-screen] Category Picker (`CategoryManagementScreen`) [IMG_7717 -> 7727] (via "Chọn nhóm" row)
│    │
│    └───> [Floating Dialog] Photo Attachment Source (`PhotoSourceActionDialog`) [IMG_7709] (via camera button)
│           ├── Mở Thư Viện Ảnh (Open photo library)
│           ├── Mở Máy ảnh (Open camera)
│           └── Huỷ (Cancel)
│
├── [Tab 3] Budgets Overview (`BudgetsScreen`) [IMG_7710, 7711]
│    │
│    ├───> [Bottom Sheet] Wallet Scope Picker (`WalletPickerSheet`) [IMG_7705] (via header wallet chip)
│    │
│    ├───> [Bottom Sheet] Add Budget (`AddBudgetSheet`) [IMG_7712] (via "Tạo Ngân sách" or "+" button)
│    │      └───> [Sub-screen] Category Picker (`CategoryManagementScreen`) [IMG_7717 -> 7727]
│    │
│    └───> [Sub-screen / Sheet] Category Budget Detail (via category budget card tap)
│
└── [Tab 4] Account & Extensions (`AccountMoreScreen`) [IMG_7713, 7714]
     │
     ├───> [Sub-screen] Account Security & Devices (`AccountSecurityScreen`) [IMG_7715]
     │      ├── Password change flow
     │      ├── Active devices list (iPhone - This device)
     │      └── Destructive actions: Logout, Reset account, Delete account
     │
     ├───> [Sub-screen] My Wallets (`WalletManagementScreen`) [IMG_7716]
     │      ├── Active wallets list included in total
     │      └── Add wallet & configuration actions
     │
     ├───> [Sub-screen] Category Management (`CategoryManagementScreen`) [IMG_7717 -> 7727]
     │      ├── [Tab 1] Expense Categories (Parent-child hierarchy) [IMG_7717 - 7724]
     │      ├── [Tab 2] Income Categories [IMG_7725, 7726]
     │      ├── [Tab 3] Debt / Loan Categories [IMG_7727]
     │      ├── [Tooltip Popover] System Lock Explanation 🔒 [IMG_7725]
     │      ├── [Sub-screen] Create New Category (via "+ Nhóm mới")
     │      └── [Toggle Action] Toggle inactive categories visibility
     │
     ├───> [Sub-screen] Recurring Transactions (`RecurringTransactionsScreen`) [IMG_7728]
     │      ├── Wallet filter chip
     │      ├── 3D illustrated empty state
     │      └── Create recurring transaction schedule
     │
     ├───> [Sub-screen] Settings (`SettingsScreen`) [IMG_7729, 7730]
     │      ├── Display settings (Currency format, Date format, Language, First day of week/month)
     │      ├── System settings (Daily transaction reminder, Biometric security)
     │      └── Database settings (Export CSV, Update exchange rates)
     │
     └───> [Bottom Sheet] Travel Mode (`TravelModeSheet`) [IMG_7731]
            ├── Travel mode toggle switch
            └── Default travel event selector
```

---

## IV. Detailed Screen Specifications

### 1. Overview / Dashboard Screen (`DashboardScreen`)
- **Assets**: `IMG_7696.PNG`, `IMG_7699.PNG`, `IMG_7700.PNG`, `IMG_7701.PNG`, `IMG_7702.PNG`.
- **Navigation Location**: Tab 1 of `BottomTabBar`.
- **Components & Layout**:
  1. **Top Bar Header**:
     - Large total balance display (`1.083.096 đ`) with `Tổng số dư (?)` subtitle.
     - Eye icon button toggling balance privacy masking.
     - Search icon button `🔍` and notification bell icon with badge count `4`.
  2. **"Ví của tôi" (My Wallets) Card**:
     - Header link `Xem tất cả` leading to `WalletManagementScreen`.
     - Grouped wallet rows: `Ví tín dụng` (`-3.711.104 đ`), `Techcombank` (`4.710.200 đ`), `Tiền Mặt` (`84.000 đ`).
  3. **"Money Insider" Card**:
     - Spending highlight for category `Ăn uống` (`250.000 đ`).
     - Daily average metric (`13.157,89 đ/ngày`) with comparison badge `214% Cao hơn tháng trước`.
     - Dual CTA buttons: `Dùng thử miễn phí` (secondary) and `Đăng ký ngay` (primary).
  4. **"Báo cáo tháng này" (Monthly Report) Carousel Card**:
     - Header action `Xem báo cáo` leading to `SpendingTrendScreen`.
     - Segmented control toggling `Tuần` and `Tháng`.
     - **Slide 1**: Weekly comparison bar chart (`Tuần trước` vs `Tuần này`) (IMG_7699).
     - **Slide 2**: Monthly comparison bar chart (`Tháng trước` vs `Tháng này`) (IMG_7701).
     - **Slide 3**: Cumulative spending trend curve compared to 3-month historical baseline (IMG_7700).
     - Carousel dot pagination controls `< Báo cáo chi tiêu / Báo cáo xu hướng >`.
  5. **"Chi tiêu nhiều nhất" (Top Spending) Card**:
     - Segmented switch for `Tuần` / `Tháng`.
     - Header link `Xem chi tiết` opening `ExpenseDetailSheet`.
     - Empty state illustration with folded hands emoji 🙌 and label `"Nhóm chi tiêu nhiều nhất sẽ hiển thị ở đây"`.
  6. **"Giao dịch gần đây" (Recent Transactions) Section**:
     - Header link `Xem tất cả` switching to `TransactionListScreen`.
     - Chronological list of recent `TransactionRow` items.

---

### 2. Add Transaction Bottom Sheet (`AddTransactionSheet`)
- **Assets**: `IMG_7707.PNG`, `IMG_7708.PNG`, `IMG_7709.PNG`.
- **Trigger**: Center raised `FloatingAddButton` (`+`) in `BottomTabBar`.
- **Basic Form State (IMG_7707)**:
  - Header: `Huỷ` dismiss pill, centered title `Thêm Giao Dịch`.
  - 3-segment control: `Khoản chi` | `Khoản thu` | `Vay/Nợ`.
  - Form Fields Card:
    - Wallet selector row (default: `Techcombank >`) -> opens `WalletPickerSheet`.
    - Hero amount input row (`AmountInputHero`) with `VND` indicator, default `0`.
    - Category picker row (`Chọn nhóm >`) -> opens `CategoryManagementScreen`.
    - Note input row (`Ghi chú >`).
    - Date selection row (`DateNavigationRow`) with `<` `>` and centered date pill.
  - Secondary pill button: `Thêm chi tiết` (Expands advanced fields).
  - Bottom sticky bar: Full-width `Lưu` button (disabled when amount is 0) and green circular camera button.
- **Expanded Details State (IMG_7708)**:
  - Person/Payee association row (`Với >`).
  - Metadata group: `Đặt vị trí >` (Location), `Chọn sự kiện >` (Event), `Đặt nhắc nhở >` (Reminder).
  - Photo attachment button: `Thêm Hình Ảnh`.
  - Report exclusion toggle: `Không tính vào báo cáo` with explanatory text.
- **Photo Source Dialog (IMG_7709)**:
  - Centered floating modal with 3 stacked pill actions: `Mở Thư Viện Ảnh`, `Mở Máy ảnh`, `Huỷ`.

---

### 3. Budgets Screen (`BudgetsScreen`) & Add Budget Sheet
- **Assets**: `IMG_7710.PNG`, `IMG_7711.PNG`, `IMG_7712.PNG`.
- **Navigation Location**: Tab 3 of `BottomTabBar`.
- **Budgets Overview Screen (IMG_7710, 7711)**:
  - Header: Title `Ngân sách Đang áp dụng`, wallet chip `🌐`, options menu `...`, help icon `?`.
  - Tab selector: `Tháng này`.
  - **Summary Hero Card**:
    - Target scope: `🌐 Tất cả các nhóm >`.
    - Circular arc gauge (`GaugeArcSummary`) showing proportion of remaining funds.
    - Hero value: `Số tiền bạn có thể chi 45.075.000 đ`.
    - 3 metrics: `65 M đ - Tổng ngân sách` | `19,92 M đ - Tổng đã chi` | `12 ngày - Đến cuối tháng`.
    - Primary button: `Tạo Ngân sách` -> opens `AddBudgetSheet`.
  - **Category Budgets List**:
    - Group cards for `Mua sắm` (8M VND, remaining 5.81M VND) and `Ăn uống` (5M VND, remaining 2.005M VND).
    - Group card for `Các nhóm còn lại (?)` (52M VND, remaining 37.26M VND).
    - Progress bars with real-time `Hôm nay` indicator markers.
  - **Bottom Alert Banner**: Blue bar indicating sample test data with a `Tắt` dismiss button.
- **Add Budget Sheet (`AddBudgetSheet` - IMG_7712)**:
  - Header: `Huỷ` pill, title `Thêm ngân sách`.
  - Grouped card: `Chọn nhóm >`, `AmountInputHero` (`VND 0`), Period selector (`Tháng này (05/08 - 04/09) >`), Wallet selector (`🌐 Tổng cộng >`).
  - Recurring toggle card: `Lặp lại ngân sách này` with recurrence explanation.
  - Bottom CTA: Full-width `Lưu` button.

---

### 4. Expense Detail Breakdown Bottom Sheet (`ExpenseDetailSheet`)
- **Assets**: `IMG_7703.PNG`, `IMG_7704.PNG`.
- **Trigger**: Tapping a category in the Dashboard top spending section or report view.
- **Layout & Structure**:
  - Header: `Đóng` pill, title `Chi tiết khoản chi`, right calendar icon.
  - Horizontal week pagination bar: Date range `03/08/2026 - 09/08/2026`, tabs `TUẦN TRƯỚC`, `TUẦN NÀY`, fast-forward button `>>`.
  - Aggregated metrics: `Tổng cộng 250.000 đ` (red font), `Trung bình hàng ngày 35.714,29 đ`.
  - 2-segment control: `Chi tiết` (Breakdown) | `Xu hướng` (Trend).
- **Tab "Chi tiết" (IMG_7703)**:
  - Centered donut breakdown chart (`DonutChartBreakdown`) with dominant category icon and `100%` badge.
  - Category breakdown list: `Ăn uống` row showing `250.000 đ` and forward chevron `>`.
- **Tab "Xu hướng" (IMG_7704)**:
  - Daily spending bar chart across days 10 through 16.
  - Day-by-day itemized list showing daily totals (e.g., day 13: `90.000 đ`, day 14: `160.000 đ`).

---

### 5. Category Management Screen (`CategoryManagementScreen`)
- **Assets**: `IMG_7717.PNG` through `IMG_7727.PNG`.
- **Navigation Location**: Pushed from Account tab or Add Transaction category picker.
- **Header**: `< Quay lại`, Title `Nhóm`, Wallet filter chip `🌐`, Search icon `🔍`.
- **Segmented Control Tabs**:
  1. **"Khoản chi" (Expense) Tab (IMG_7717 - IMG_7724)**:
     - Top action pill: `+ Nhóm mới` (Create new category).
     - Hierarchical categories with parent-child connection lines:
       - `Ăn uống` (Parent) -> `Ăn vặt`, `Cà phê`, `Cơm Bữa` (Children).
       - `Hoá đơn & Tiện ích` -> `Hoá đơn điện thoại`, `nước`, `điện`, `gas`, `TV`, `internet`, `Thuê nhà`, `tiện ích khác`.
       - `Mua sắm` -> `Đồ dùng cá nhân`, `Đồ gia dụng`, `Làm đẹp`.
       - `Gia đình` -> `Sửa & trang trí nhà`, `Dịch vụ gia đình`, `Vật nuôi`.
       - `Di chuyển` -> `Bảo dưỡng xe`, `Gửi xe`, `Xăng dầu`, `Taxi`.
       - `Sức khỏe` -> `Khám sức khoẻ`, `Thể dục thể thao`, `Thuốc`.
       - `Giáo dục`, `Giải trí`, `Quà tặng & Quyên góp`, `Bảo hiểm`, `Đầu tư`.
       - Locked system categories (🔒): `Các chi phí khác`, `Tiền chuyển đi`, `Trả lãi`, `Khoản chi chưa phân loại`, `Rút tiền`.
  2. **"Khoản thu" (Income) Tab (IMG_7725, IMG_7726)**:
     - Standard categories: `Lương`, `Thưởng`, `Bán đồ`.
     - Locked system categories (🔒): `Thu nhập khác`, `Tiền chuyển đến`, `Thu lãi`, `Khoản thu chưa phân loại`, `Được tặng`.
     - Dark popover tooltip when tapping a locked category: `"Đây là danh mục của hệ thống nên bạn không thể chỉnh sửa hoặc xóa bỏ."`.
  3. **"Vay/Nợ" (Debt & Loan) Tab (IMG_7727)**:
     - Locked system categories: `Cho vay`, `Trả nợ`, `Đi vay`, `Thu nợ`.
- **Bottom Footer Action**: Pill button with eye icon `👁 Hiển thị nhóm không hoạt động`.

---

### 6. Account Security & Device Screen (`AccountSecurityScreen`)
- **Assets**: `IMG_7715.PNG`.
- **Navigation Location**: Pushed from Account tab -> `Quản lý tài khoản`.
- **Structure**:
  - Header: `< Quay lại`, Title `Quản Lý Tài Khoản`.
  - User profile badge card: Initials avatar `D`, orange ribbon `TÀI KHOẢN PREMIUM`, username, Google email.
  - Action pill button: `Thay đổi mật khẩu` (Change password).
  - Devices section (`THIẾT BỊ 1/5`): Device row with Apple logo, `iPhone`, `Thiết bị này`.
  - Standalone red action pill: `Đăng xuất` (Log out).
  - Destructive operations card:
    - Row: `Đặt lại tài khoản` (Reset account, red text).
    - Divider.
    - Row: `Xóa tài khoản` (Delete account, red text).

---

### 7. Settings Screen (`SettingsScreen`)
- **Assets**: `IMG_7729.PNG`, `IMG_7730.PNG`.
- **Navigation Location**: Pushed from Account tab -> `Cài đặt`.
- **Grouped Sections**:
  - **"Hiển thị" (Display) Card**:
    - `Kiểu hiển thị số tiền >`
    - `Định dạng thời gian` (`23/08/2026 >`)
    - `Chọn ngôn ngữ` (`Tiếng Việt >`)
    - `Đơn vị tiền cho ví Tổng` (`VND >`)
    - `Chọn ngày đầu tuần` (`Thứ hai >`)
    - `Đặt ngày đầu tiên của tháng` (`5 >`)
    - `Chọn tháng đầu tiên của năm` (`Tháng Một >`)
    - `Thiết lập khoảng thời gian...` (`Trong 3 tháng tiếp theo >`)
    - `Chế độ hiển thị Tổng quan` (`Hiện theo tiền và... >`)
    - Switch row: `Không tính vào báo cáo` with description.
  - **"Hệ thống" (System) Card**:
    - Switch row: `Nhắc thêm giao dịch hàng ngày` (Daily transaction reminder toggle).
    - Row: `Bảo mật >` (Biometric & passcode settings, not synced across devices).
  - **"Cơ sở dữ liệu" (Database) Card**:
    - Row: `Xuất file CSV >`.
    - Row: `Chạm để cập nhật tỷ giá` with last updated date (`2026-06-24`) and refresh icon `🔄`.

---

### 8. Supporting Auxiliary Screens & Modals

#### My Wallets Management (`WalletManagementScreen` - IMG_7716)
- Header: `< Quay lại`, Title `Ví của tôi`, Menu icon `≡`, Action button `Thêm`.
- Section `TÍNH VÀO TỔNG`: Card enclosing wallet items with balances.

#### Wallet Scope Picker Sheet (`WalletPickerSheet` - IMG_7705)
- Header: `Đóng`, Title `Chọn Ví`, Action `Sửa`.
- Selection card: `Tổng cộng` row with active checkmark `✓`.
- Section `TÍNH VÀO TỔNG`: List of specific wallets for isolated view.
- Bottom actions: `(+) Thêm ví` and `(🔗) Liên kết dịch vụ`.

#### Recurring Transactions (`RecurringTransactionsScreen` - IMG_7728)
- Header: `< Quay lại`, Title `Giao dịch định kì`, Wallet chip `🌐`, Action button `+`.
- Subtitle: `"Tạo ra định kỳ các giao dịch sẽ được tự động thêm trong tương lai."`.
- 3D illustrated empty state view: `"Không có giao dịch định kỳ nào"`.

#### Travel Mode Sheet (`TravelModeSheet` - IMG_7731)
- Header: `Huỷ`, Title `Chế Độ Du Lịch`, Action `Xong` (green text).
- Toggle card: `Bật Chế Độ Du Lịch` with explanation text.
- Event card: `Chọn sự kiện >`.

---

## V. Phase-by-Phase Component Contracts

This UI specification integrates with the backend synchronization and financial calculation engine across phases:

### 1. PHASE-003 Offline Components
- `SyncStatusPill`: Compact sync indicator (synced, syncing, offline, conflict) positioned in the top bar; non-blocking for local cached reads.
- `PendingMutationBadge`: Micro badge rendered on optimistic wallet, transaction, and budget rows pending server confirmation.
- `ConflictInbox`: Full-width resolution view grouped by entity displaying local intent, server canonical value, timestamp, and 3 actions: Keep server, Edit and retry, Discard local.
- `OfflineRecoverySheet`: Explains quota limits, invalid replication cursors, or full re-sync requirements with single-tap recovery.

### 2. PHASE-004 Planning Components
- `BudgetSummary`: Current period amount, used/remaining values, 80%/100% threshold states, progress bar, and period selector.
- `BudgetEditorSheet`: Name, period, amount, category scope, alert toggles, version conflict indicators, save/archive commands.
- `EventList` and `EventEditorSheet`: Date range, location, aggregated expense totals, and transaction linkage.
- `ObligationCard`: Borrow/lend direction, counterparty name, principal, repaid, remaining, due date, and repayment log.
- `RecurringDraftCard`: Source schedule, trigger date, resolved wallet/category, editable proposal, confirm/reject actions.
- `NotificationInbox`: Unread filter, durable alerts, read states, and push permission banner.

### 3. PHASE-005 Dashboard & Report Components
- `NetWorthHeader`: Privacy masking toggle, selected wallet aggregate, offline/stale status, and wallet picker trigger.
- `RecentTransactions`: Live API data, pending-sync badges, empty/error fallbacks, and direct edit sheet triggers.
- `GlobalSearchSheet`: Grouped multi-entity search results (transactions, wallets, categories, events, debts).
- `MetricSummary`: Net balance, income, expense, and period-over-period comparison (displays `Không thể so sánh` if base is 0).
- `CategoryDonut`, `DailyBars`, and `CumulativeTrend`: Accessible charts backed by structured tabular numbers.

### 4. PHASE-006 Ingestion & Draft Components
- `AIComposer`: Text input, receipt image attachment, send/cancel, model processing state, and rate-limit handler.
- `ProposalCard`: Ingestion source, transaction type, amount, candidate wallet/category, editable fields, confidence score, confirm/reject actions.
- `ReceiptCaptureSheet`: Camera and photo library triggers, secure private upload progress, image preview, OCR status.
- `DraftReviewQueue`: Aggregates pending AI, OCR, bank webhook, and recurring proposals with independent confirmation.

### 5. PHASE-007 Account, Export & Audit Components
- `ExportSheet`: Dataset/date range picker, snapshot description, job progress, retry on failure, and expiring download link.
- `DestructiveActionSheet`: Reset/delete impact preview, re-authentication validation, typed confirmation, and progress bar; destructive red styling reserved exclusively for the final execution command.
- `AuditViewer`: Dense, read-only system audit log with timestamp, action type, actor, target entity, correlation ID, and masked details.
- `OperationsStatus`: Administrative health dashboard for production runtime monitoring.

---

## VI. Verification & Quality Assurance Rules

- **Mobile Viewport Adherence**:
  - All screens must strictly adhere to standard iPhone dimensions (`375px` – `430px` width) with fluid layout responsiveness.
  - Text and numerical amounts must never overflow pill buttons, cards, tabs, or the bottom navigation bar.
- **Localization & Formatting**:
  - Vietnamese text copy and VND currency formatting (thousands separator with `.`, e.g., `1.083.096 đ`) are mandatory UI defaults.
  - Raw unformatted integers must never be presented to the user.
- **State Handling Requirements**:
  - Every list and data-bound component must implement all 5 lifecycle states: Loading (`loading`), Empty (`empty`), Error (`error`), Offline Cache (`offline-cache`), and End-of-list (`pagination-end`).
  - Data mutating buttons must visually convey saving (`saving`), queued offline (`queued-offline`), conflict (`conflict`), retry (`retry`), and success (`success`) states without causing visual layout shifts (CLS).
