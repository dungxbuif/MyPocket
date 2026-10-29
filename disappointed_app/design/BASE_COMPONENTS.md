# MyPocket - Base Components Specification

This document defines the production-ready component contracts for the MyPocket mobile web/PWA application, derived strictly from the 36 UI screenshot assets (`design/IMG_7696.PNG` through `design/IMG_7731.PNG`).

Every component includes its **Props Interface (TypeScript)**, **Visual Tokens & CSS Rules**, **Interactive States**, **Accessibility & Tap Targets**, and **Screenshot References**.

---

## Table of Contents

1. [Design System Tokens Quick Reference](#1-design-system-tokens-quick-reference)
2. [Navigation & Layout Components](#2-navigation--layout-components)
   - [AppShell / SafeAreaLayout](#appshell--safearealayout)
   - [BottomTabBar](#bottomtabbar)
   - [SubpageHeader](#subpageheader)
   - [ModalHeader](#modalheader)
   - [WalletFilterChip](#walletfilterchip)
   - [SegmentedControl](#segmentedcontrol)
3. [Containers & Grouped Cards](#3-containers--grouped-cards)
   - [GroupedCard](#groupedcard)
   - [FormRowItem](#formrowitem)
   - [SwitchRow](#switchrow)
   - [RadioCheckItem](#radiocheckitem)
4. [Inputs & Form Controls](#4-inputs--form-controls)
   - [AmountInputHero](#amountinputhero)
   - [DateNavigationRow](#datenavigationrow)
   - [SearchInputField](#searchinputfield)
   - [PhotoAttachmentPicker](#photoattachmentpicker)
5. [Action Buttons & Feedback](#5-action-buttons--feedback)
   - [PrimaryButton](#primarybutton)
   - [SecondaryButton](#secondarybutton)
   - [CircleIconButton](#circleiconbutton)
   - [DestructiveActionRow](#destructiveactionrow)
6. [Data Display & Financial Visualizations](#6-data-display--financial-visualizations)
   - [TransactionRow](#transactionrow)
   - [WalletRow](#walletrow)
   - [CategoryTreeRow](#categorytreerow)
   - [GaugeArcSummary](#gaugearcsummary)
   - [ProgressBarWithMarker](#progressbarwithmarker)
   - [ComparisonBarChart](#comparisonbarchart)
   - [TrendAreaLineChart](#trendarealinechart)
   - [DonutChartBreakdown](#donutchartbreakdown)
7. [Feedback, Overlays & Placeholders](#7-feedback-overlays--placeholders)
   - [EmptyStateView](#emptystateview)
   - [NoticeBanner](#noticebanner)
   - [BottomSheetContainer](#bottomsheetcontainer)
   - [TooltipPopover](#tooltippopover)

---

## 1. Design System Tokens Quick Reference

```css
:root {
  /* Surfaces & Canvas */
  --bg-app: #f2f3f8;
  --bg-card: #ffffff;
  --bg-subtle: #f8f9fb;
  --bg-track: #e9eaef;

  /* Typography Colors */
  --text-primary: #111111;
  --text-secondary: #8e8e93;
  --text-tertiary: #b0b0b8;
  --text-inverse: #ffffff;

  /* Brand & Financial Colors */
  --brand-green: #2dbd4f;
  --brand-green-soft: #e8f7ed;
  --brand-green-dark: #1c8535;
  --color-expense: #ff5a66;
  --color-expense-soft: #ffebee;
  --color-income: #32a9df;
  --color-income-soft: #e3f2fd;
  --color-accent-orange: #ff8800;
  --color-lock-gray: #8e8e93;

  /* Borders & Dividers */
  --border-divider: #e8e8ec;
  --border-subtle: #f0f0f4;

  /* Border Radii */
  --radius-card: 24px;
  --radius-sheet: 28px;
  --radius-pill: 9999px;
  --radius-icon: 50%;
  --radius-sm: 12px;

  /* Elevations */
  --shadow-card: 0 2px 10px rgba(0, 0, 0, 0.03);
  --shadow-nav: 0 -4px 20px rgba(0, 0, 0, 0.06);
  --shadow-floating: 0 6px 20px rgba(45, 189, 79, 0.35);

  /* Font Stack */
  --font-family-sans: -apple-system, BlinkMacSystemFont, "San Francisco", "Helvetica Neue", sans-serif;
}
```

---

## 2. Navigation & Layout Components

### `AppShell` / `SafeAreaLayout`
- **Screens**: Universal root layout for all screens (`IMG_7696` – `IMG_7731`).
- **Description**: Outer viewport shell managing dynamic iOS safe areas (`env(safe-area-inset-top)`, `env(safe-area-inset-bottom)`), preventing rubber-band overscroll issues, and handling top status bar background styling.
- **Props Interface**:
  ```typescript
  export interface AppShellProps {
    children: React.ReactNode;
    backgroundColor?: string; // Default: var(--bg-app) #f2f3f8
    statusBarStyle?: 'dark-content' | 'light-content'; // Default: 'dark-content'
    bottomBar?: React.ReactNode; // Slot for BottomTabBar
  }
  ```
- **CSS Rules**:
  - `min-height: 100dvh; max-width: 440px; margin: 0 auto;`
  - `padding-top: max(16px, env(safe-area-inset-top));`
  - `padding-bottom: max(90px, calc(env(safe-area-inset-bottom) + 70px));` (when bottom bar is visible).
- **Accessibility**: Landmark `<main role="main">` with semantic `<header>` and `<nav>`.

---

### `BottomTabBar`
- **Screens**: `IMG_7696`, `IMG_7706`, `IMG_7710`, `IMG_7714`.
- **Description**: 5-slot bottom floating navigation bar with a raised, circular center `FloatingAddButton`. The active tab is framed with a soft gray elongated pill.
- **Props Interface**:
  ```typescript
  export type TabKey = 'overview' | 'transactions' | 'budgets' | 'account';

  export interface BottomTabBarProps {
    activeTab: TabKey;
    onTabSelect: (tab: TabKey) => void;
    onAddClick: () => void;
  }
  ```
- **Slots & Visual Hierarchy**:
  1. `overview`: Icon `Home`, Label `Tổng quan`
  2. `transactions`: Icon `Wallet`, Label `Sổ giao dịch`
  3. `add`: Center raised green circle (`FloatingAddButton`) with white `+` glyph. Diameter: `52px`, elevated with `--shadow-floating`.
  4. `budgets`: Icon `Ledger / Target`, Label `Ngân sách`
  5. `account`: Icon `User / Settings`, Label `Tài khoản`
- **States**:
  - `Active Tab`: Gray pill background (`#eef0f4`), dark primary text (`#111111`), bold font.
  - `Inactive Tab`: Transparent background, muted text (`#8e8e93`), regular font.
  - `Center Add Button Pressed`: Scale `0.94`, background `#25a443`.
- **CSS Tokens**:
  - Height: `64px`, border-radius: `9999px`, background: `rgba(255, 255, 255, 0.94)`, backdrop-filter: `blur(16px)`.
  - Position: `fixed; bottom: max(16px, env(safe-area-inset-bottom)); left: 16px; right: 16px;`.

---

### `SubpageHeader`
- **Screens**: `IMG_7697`, `IMG_7715`, `IMG_7716`, `IMG_7717`, `IMG_7728`, `IMG_7729`.
- **Description**: Standard header for secondary and pushed screens. Contains back button, centered title, optional filter chip, and optional right action.
- **Props Interface**:
  ```typescript
  export interface SubpageHeaderProps {
    title: string;
    onBack: () => void;
    backLabel?: string; // Default: "Quay lại"
    centerExtra?: React.ReactNode; // e.g. WalletFilterChip
    rightAction?: {
      label?: string; // e.g. "Thêm"
      icon?: React.ReactNode; // e.g. Search, Menu, Help
      onClick: () => void;
      disabled?: boolean;
    };
  }
  ```
- **Structure**:
  - Left: Back pill button with chevron `<` and localized label (`Quay lại`). Tap target min `44px x 44px`.
  - Center: Truncated bold title (`font-size: 17px; font-weight: 600; text-align: center;`).
  - Right: Action button or icon (Search `🔍`, Menu `≡`, text action `Thêm`).

---

### `ModalHeader`
- **Screens**: `IMG_7703`, `IMG_7705`, `IMG_7707`, `IMG_7712`, `IMG_7731`.
- **Description**: Standard header for bottom sheets, overlays, and full-sheet creation flows.
- **Props Interface**:
  ```typescript
  export interface ModalHeaderProps {
    title: string;
    onDismiss: () => void;
    dismissLabel?: string; // "Huỷ" or "Đóng" (Default: "Huỷ")
    rightAction?: {
      label?: string; // "Xong", "Sửa", "Lưu"
      icon?: React.ReactNode; // Calendar icon
      onClick: () => void;
      variant?: 'primary' | 'muted';
    };
  }
  ```
- **Visuals**:
  - Dismiss button: Soft gray pill button (`#eef0f4`), text `#111111`, `padding: 6px 14px`.
  - Title: Centered, `17px`, bold font.
  - Right action: Text action (`Xong` with green `#2dbd4f` bold font, or `Sửa` with muted font).

---

### `WalletFilterChip`
- **Screens**: `IMG_7697`, `IMG_7706`, `IMG_7710`, `IMG_7717`, `IMG_7728`.
- **Description**: Compact pill button displaying the currently active wallet scope.
- **Props Interface**:
  ```typescript
  export interface WalletFilterChipProps {
    walletName: string; // e.g. "Tổng cộng", "Techcombank", "Ví tín dụng"
    iconType?: 'globe' | 'wallet' | 'bank'; // Default: 'globe'
    onClick: () => void;
  }
  ```
- **Visuals**:
  - Container: Height `32px`, border-radius `9999px`, background `#eef0f4`, padding `4px 12px`.
  - Content: Leading icon (`🌐` or wallet logo) + wallet label + trailing up/down arrows (`▲▼` or `⌄`).

---

### `SegmentedControl`
- **Screens**: `IMG_7696`, `IMG_7703`, `IMG_7707`, `IMG_7717`.
- **Description**: Sliding horizontal pill switch for switching views or categories.
- **Props Interface**:
  ```typescript
  export interface SegmentOption<T extends string = string> {
    key: T;
    label: string;
  }

  export interface SegmentedControlProps<T extends string = string> {
    options: SegmentOption<T>[];
    selectedKey: T;
    onChange: (key: T) => void;
    size?: 'sm' | 'md' | 'lg'; // sm for card filters, md for tabs, lg for modal switchers
  }
  ```
- **Variants in App**:
  1. 2-segment: `Tuần` | `Tháng` (Dashboard card filter).
  2. 2-segment: `Chi tiết` | `Xu hướng` (Expense breakdown sheet).
  3. 3-segment: `Khoản chi` | `Khoản thu` | `Vay/Nợ` (Transaction and Category management).
- **Styling**:
  - Container: Background `#e5e6eb`, border-radius `9999px`, padding `3px`.
  - Active segment: White pill (`#ffffff`), `box-shadow: 0 2px 6px rgba(0,0,0,0.08)`, smooth transition `transform 0.2s cubic-bezier(0.4, 0, 0.2, 1)`.

---

## 3. Containers & Grouped Cards

### `GroupedCard`
- **Screens**: `IMG_7696`, `IMG_7705`, `IMG_7707`, `IMG_7713`, `IMG_7729`.
- **Description**: White rounded card aggregating rows of settings, transaction inputs, or wallet metadata.
- **Props Interface**:
  ```typescript
  export interface GroupedCardProps {
    title?: string; // Optional section title displayed above the card (e.g., "TÍNH VÀO TỔNG", "HIỂN THỊ")
    headerAction?: {
      label: string; // e.g. "Xem tất cả"
      onClick: () => void;
    };
    children: React.ReactNode;
    className?: string;
  }
  ```
- **CSS Rules**:
  - Card background: `#ffffff`, border-radius: `24px` to `28px`, padding: `14px 16px`.
  - Section title: Uppercase `12px` bold, color `#8e8e93`, letter-spacing `0.5px`, margin `0 0 8px 12px`.
  - Direct children rows automatically receive `#e8e8ec` bottom dividers (except the last child).

---

### `FormRowItem`
- **Screens**: `IMG_7707`, `IMG_7708`, `IMG_7713`, `IMG_7729`.
- **Description**: Interactive row element for selection, navigation, or parameter display.
- **Props Interface**:
  ```typescript
  export interface FormRowItemProps {
    leadingIcon?: React.ReactNode;
    leadingBadge?: React.ReactNode;
    label: string;
    description?: string;
    value?: string | React.ReactNode; // Right-aligned current selection (e.g. "Techcombank", "23/08/2026")
    showChevron?: boolean; // Trailing '>' chevron (Default: true)
    onClick?: () => void;
    disabled?: boolean;
  }
  ```
- **CSS Rules**:
  - Height: Min `52px`, flex align center, justify space-between.
  - Hover / Active state: Background `#f8f9fa` with active press feedback.
  - Chevron: Color `#c7c7cc`, `font-size: 14px`.

---

### `SwitchRow`
- **Screens**: `IMG_7708`, `IMG_7712`, `IMG_7729`, `IMG_7731`.
- **Description**: Setting row containing a label, explanatory subtext, and an iOS-style toggle switch.
- **Props Interface**:
  ```typescript
  export interface SwitchRowProps {
    title: string;
    description?: string;
    checked: boolean;
    onChange: (checked: boolean) => void;
    disabled?: boolean;
  }
  ```
- **CSS Rules**:
  - Switch Track: Width `51px`, Height `31px`, border-radius `15.5px`.
  - Checked: Track background `#2dbd4f`.
  - Unchecked: Track background `#e5e5ea`.
  - Thumb: White circle `27px` diameter, `transform: translateX(20px)` when checked.

---

### `RadioCheckItem`
- **Screens**: `IMG_7705`.
- **Description**: Selectable row within a single-choice picker (e.g., Wallet Selector).
- **Props Interface**:
  ```typescript
  export interface RadioCheckItemProps {
    leadingIcon?: React.ReactNode;
    label: string;
    sublabel?: string;
    selected: boolean;
    onSelect: () => void;
  }
  ```
- **Display**: When `selected === true`, renders a bold green checkmark (`✓`) at the right trailing edge.

---

## 4. Inputs & Form Controls

### `AmountInputHero`
- **Screens**: `IMG_7707`, `IMG_7708`, `IMG_7712`.
- **Description**: Hero-scale numeric currency input with leading `VND` badge, tabular bold numerals, and auto-formatting thousands separator.
- **Props Interface**:
  ```typescript
  export interface AmountInputHeroProps {
    value: number; // Value in minor or base units
    currency?: string; // Default: 'VND'
    onChange: (value: number) => void;
    placeholder?: string; // Default: "0"
    autoFocus?: boolean;
    readOnly?: boolean;
  }
  ```
- **Structure & Styling**:
  - Container: Row flex, align-items center, gap `12px`, padding `12px 0`.
  - Currency Badge: `#eef0f4` pill, text `13px` bold, color `#111111`, padding `4px 10px`.
  - Number Field: `font-size: 34px; font-weight: 700; font-variant-numeric: tabular-nums; color: #111111;`
  - Zero state: Renders `0` in bold muted `#8e8e93` or `#111111`.
  - Auto-formatting: Formats inputs as `250.000` or `45.075.000`.

---

### `DateNavigationRow`
- **Screens**: `IMG_7707`, `IMG_7708`.
- **Description**: Inline date bar featuring step-back (`<`) and step-forward (`>`) day buttons around a centered date pill.
- **Props Interface**:
  ```typescript
  export interface DateNavigationRowProps {
    currentDate: Date | string; // ISO or Date object
    onDateChange: (date: Date) => void;
    onOpenCalendarModal?: () => void;
  }
  ```
- **Display**:
  - Leading: Calendar glyph `📅`.
  - Left Step Button: `<` (Decrements by 1 day).
  - Center Date Pill: `Chủ Nhật, 23/08/2026` (Localized Vietnamese day of week and DD/MM/YYYY). Tapping opens full datepicker.
  - Right Step Button: `>` (Increments by 1 day).

---

### `SearchInputField`
- **Screens**: `IMG_7696`, `IMG_7717`.
- **Description**: Pill-shaped text search bar with search glyph and quick-clear action.
- **Props Interface**:
  ```typescript
  export interface SearchInputFieldProps {
    value: string;
    placeholder?: string; // e.g. "Tìm kiếm giao dịch, nhóm..."
    onChange: (value: string) => void;
    onClear: () => void;
    autoFocus?: boolean;
  }
  ```
- **Visuals**: Background `#eef0f4`, border-radius `9999px`, height `40px`, padding `0 14px`, leading search icon `🔍`, trailing clear button `✕` when `value.length > 0`.

---

### `PhotoAttachmentPicker`
- **Screens**: `IMG_7708`, `IMG_7709`.
- **Description**: Component handling receipt or transaction photo attachments and triggering source options.
- **Props Interface**:
  ```typescript
  export interface PhotoAttachmentPickerProps {
    attachedImages: string[]; // URLs / local URIs
    onAddClick: () => void; // Opens source dialog
    onRemoveImage: (index: number) => void;
  }
  ```
- **Actions Dialog (`IMG_7709`)**:
  - Option 1: `Mở Thư Viện Ảnh` (Open Photo Library).
  - Option 2: `Mở Máy ảnh` (Open Camera).
  - Option 3: `Huỷ` (Cancel dismiss button).

---

## 5. Action Buttons & Feedback

### `PrimaryButton`
- **Screens**: `IMG_7696`, `IMG_7707`, `IMG_7710`, `IMG_7712`.
- **Description**: Full-width or card-level primary call-to-action pill button.
- **Props Interface**:
  ```typescript
  export interface PrimaryButtonProps {
    label: string; // e.g. "Lưu", "Tạo Ngân sách", "Đăng ký ngay"
    onClick: () => void;
    disabled?: boolean;
    loading?: boolean;
    fullWidth?: boolean; // Default: true
  }
  ```
- **CSS Rules**:
  - Height: `50px` – `52px`, border-radius: `9999px`.
  - Normal state: Background `#2dbd4f`, color `#ffffff`, font-weight `600`, font-size `16px`.
  - Pressed state: Background `#25a443`, scale `0.98`.
  - Disabled state: Background `#e0e0e0`, color `#8e8e93`, cursor `not-allowed`.

---

### `SecondaryButton`
- **Screens**: `IMG_7696`.
- **Description**: Supporting pill button with soft green background.
- **Props Interface**:
  ```typescript
  export interface SecondaryButtonProps {
    label: string; // e.g. "Dùng thử miễn phí"
    onClick: () => void;
    disabled?: boolean;
  }
  ```
- **CSS Rules**:
  - Background `#e8f7ed`, text color `#1c8535`, border-radius `9999px`, font-weight `600`.

---

### `CircleIconButton`
- **Screens**: `IMG_7696`, `IMG_7697`, `IMG_7706`, `IMG_7707`.
- **Description**: Circular button component for utility actions (Back, Camera, Search, Options menu, Notification bell).
- **Props Interface**:
  ```typescript
  export interface CircleIconButtonProps {
    icon: React.ReactNode;
    onClick: () => void;
    badgeCount?: number; // e.g. 4 unread notifications
    ariaLabel: string;
    variant?: 'default' | 'filled' | 'camera';
  }
  ```
- **Visuals**:
  - Dimensions: `40px x 40px` (or `44px x 44px` touch zone).
  - Camera variant (`IMG_7707`): Background `#2dbd4f`, white camera glyph, elevated shadow.
  - Notification badge: Small red circle `#ff5a66` at top right with white count numeral.

---

### `DestructiveActionRow`
- **Screens**: `IMG_7715`.
- **Description**: Danger-level action button or row reserved for irreversible operations.
- **Props Interface**:
  ```typescript
  export interface DestructiveActionRowProps {
    label: string; // "Đăng xuất", "Đặt lại tài khoản", "Xóa tài khoản"
    onClick: () => void;
    variant?: 'pill' | 'row';
  }
  ```
- **Visuals**: High-contrast red font (`#ff5a66`). If `variant === 'pill'`, renders as full pill with white or light pink background.

---

## 6. Data Display & Financial Visualizations

### `TransactionRow`
- **Screens**: `IMG_7702`, `IMG_7706`.
- **Description**: Itemized row representing a financial transaction log.
- **Props Interface**:
  ```typescript
  export interface TransactionRowProps {
    id: string;
    categoryName: string; // e.g. "Cơm Bữa", "Ăn vặt", "Trả nợ"
    categoryIcon: React.ReactNode;
    walletName?: string; // e.g. "Techcombank"
    walletBadgeIcon?: React.ReactNode; // Mini wallet badge overlaid on bottom right of category icon
    note?: string; // Optional user note or payee
    amount: number; // Positive for income, negative for expense
    currency?: string; // Default: 'VND'
    dateFormatted?: string; // e.g. "23/08/2026"
    isExcludedFromReport?: boolean;
    onClick?: () => void;
  }
  ```
- **Visuals**:
  - Leading: `44px` circular category icon container (`#29495a` or themed background) with a `16px` mini wallet emblem at bottom-right corner.
  - Middle: Category name in bold `15px`, optional note or wallet name in `13px` muted text.
  - Right: Formatted currency. Expenses in red (`-90.000 đ`), Income in blue (`+5.000.000 đ`), Debt/Loan in neutral dark.

---

### `WalletRow`
- **Screens**: `IMG_7696`, `IMG_7705`, `IMG_7716`.
- **Description**: Row displaying wallet balance and institutional identity.
- **Props Interface**:
  ```typescript
  export interface WalletRowProps {
    id: string;
    name: string; // "Ví tín dụng", "Techcombank", "Tiền Mặt"
    balance: number;
    icon: React.ReactNode; // Credit card glyph, bank logo, cash icon
    currency?: string; // Default: "VND"
    selected?: boolean; // For picker view
    onClick?: () => void;
  }
  ```
- **Display**:
  - Balance supports negative credit display: `-3.711.104 đ` in red or dark text.
  - When used inside `WalletPickerSheet`, renders trailing checkmark `✓` if `selected === true`.

---

### `CategoryTreeRow`
- **Screens**: `IMG_7717` – `IMG_7727`.
- **Description**: Hierarchical category tree item supporting parent-child branching lines, usage count, and locked system status.
- **Props Interface**:
  ```typescript
  export interface CategoryTreeRowProps {
    id: string;
    name: string;
    icon: React.ReactNode;
    isChild?: boolean; // Indents row and draws connecting branch line
    isLastChild?: boolean;
    walletCount?: number; // e.g. 3 -> "Hoạt động trong 3 ví"
    isSystemLocked?: boolean; // Renders 🔒 icon
    onLockedClick?: () => void; // Shows lock explanation tooltip
    onClick?: () => void;
  }
  ```
- **Tree Line CSS**:
  - Child row has `padding-left: 36px; position: relative;`.
  - Pseudo-element `::before`: L-shaped tree branch border (`width: 16px; height: 50%; border-left: 1.5px solid #d1d1d6; border-bottom: 1.5px solid #d1d1d6; position: absolute; left: 16px; top: 0;`).

---

### `GaugeArcSummary`
- **Screens**: `IMG_7710`, `IMG_7711`.
- **Description**: Circular arc gauge visualizing available budget proportion and summary metrics.
- **Props Interface**:
  ```typescript
  export interface GaugeArcSummaryProps {
    availableAmount: number; // e.g. 45075000
    totalBudget: number; // e.g. 65000000
    totalSpent: number; // e.g. 19920000
    daysRemaining: number; // e.g. 12
    currency?: string;
  }
  ```
- **Structure**:
  - Top SVG Arc: 200° progress gauge with colored arc fill (green or warning yellow/red) and indicator head dot.
  - Center Hero: `Số tiền bạn có thể chi` with large bold amount (`45.075.000 đ`).
  - 3-Column Metrics Footer:
    - Col 1: `65 M đ` - `Tổng ngân sách`
    - Col 2: `19,92 M đ` - `Tổng đã chi`
    - Col 3: `12 ngày` - `Đến cuối tháng`

---

### `ProgressBarWithMarker`
- **Screens**: `IMG_7710`, `IMG_7711`.
- **Description**: Horizontal progress bar visualizing category budget exhaustion with a real-time calendar day marker labeled `Hôm nay`.
- **Props Interface**:
  ```typescript
  export interface ProgressBarWithMarkerProps {
    spentPercent: number; // 0 to 100
    dayProgressPercent: number; // 0 to 100 (where current day falls in the month)
    spentAmountFormatted: string; // e.g. "2.190.000 đ"
    remainingAmountFormatted: string; // e.g. "Còn lại 5.810.000 đ"
    color?: string; // Default: var(--brand-green) #2dbd4f
  }
  ```
- **Visuals**:
  - Bar Track: Height `8px`, border-radius `4px`, background `#e9eaef`.
  - Fill: Progress fill proportional to `spentPercent`.
  - Floating Marker: Small vertical pill flag labeled `Hôm nay` positioned at `left: ${dayProgressPercent}%`.

---

### `ComparisonBarChart`
- **Screens**: `IMG_7699`, `IMG_7701`, `IMG_7704`.
- **Description**: Bar chart evaluating spending between periods (`Tuần trước` vs `Tuần này`, `Tháng trước` vs `Tháng này`) or daily spend logs.
- **Props Interface**:
  ```typescript
  export interface BarDataPoint {
    label: string; // "Tuần trước", "Tuần này" or "10", "11", "12"
    amount: number;
    amountFormatted: string;
    isCurrentPeriod?: boolean;
  }

  export interface ComparisonBarChartProps {
    data: BarDataPoint[];
    maxAmount?: number;
    height?: number; // Default: 160px
  }
  ```
- **Visuals**:
  - Previous period: Light shaded bar (`#ffd2d6`).
  - Current period: Solid high-contrast red bar (`#ff5a66`).
  - Baseline: Subtle horizontal `#e8e8ec` zero axis with top amount labels.

---

### `TrendAreaLineChart`
- **Screens**: `IMG_7700`.
- **Description**: Cumulative expenditure curve plotted against 3-month historical baseline.
- **Props Interface**:
  ```typescript
  export interface TrendPoint {
    dateLabel: string;
    actualAmount: number;
    baselineAmount: number;
  }

  export interface TrendAreaLineChartProps {
    points: TrendPoint[];
    currentDateIndex?: number;
  }
  ```
- **Visuals**:
  - Baseline: Shaded soft gray/red zone representing historical 3-month average.
  - Actual: Bold solid red curve (`#ff5a66`) with milestone data points and interactive touch tooltip showing exact day cumulative sum.

---

### `DonutChartBreakdown`
- **Screens**: `IMG_7703`.
- **Description**: Donut chart illustrating percentage distribution of expenses across categories.
- **Props Interface**:
  ```typescript
  export interface DonutSlice {
    categoryKey: string;
    categoryName: string;
    percentage: number; // 0 - 100
    color: string;
    icon?: React.ReactNode;
  }

  export interface DonutChartBreakdownProps {
    slices: DonutSlice[];
    dominantCategoryIcon?: React.ReactNode;
    dominantPercentage?: number; // e.g. 100%
  }
  ```
- **Visuals**: Segmented SVG ring (`stroke-width: 18px`). Center displays the dominant category icon and badge pill (`100%`).

---

## 7. Feedback, Overlays & Placeholders

### `EmptyStateView`
- **Screens**: `IMG_7702`, `IMG_7728`.
- **Description**: Placeholder displayed when transaction lists, budgets, or recurring items are empty.
- **Props Interface**:
  ```typescript
  export interface EmptyStateViewProps {
    illustration?: React.ReactNode; // 3D vector illustration or emoji (e.g. 🙌)
    title: string; // e.g. "Không có giao dịch định kỳ nào", "Nhóm chi tiêu nhiều nhất sẽ hiển thị ở đây"
    description?: string;
    actionButton?: {
      label: string;
      onClick: () => void;
    };
  }
  ```
- **Styling**: Centered column layout, padding `40px 20px`, text color `#8e8e93`, max-width `280px`.

---

### `NoticeBanner`
- **Screens**: `IMG_7710`, `IMG_7711`.
- **Description**: Informational banner anchored at the bottom of the screen (e.g., sample demonstration data notice).
- **Props Interface**:
  ```typescript
  export interface NoticeBannerProps {
    message: string; // e.g. "Dữ liệu trên là ví dụ minh họa..."
    dismissLabel?: string; // Default: "Tắt"
    onDismiss: () => void;
  }
  ```
- **Styling**: Background `#2f80ed`, color `#ffffff`, border-radius `14px`, margin `12px 16px`, padding `10px 14px`, flex justify space-between.

---

### `BottomSheetContainer`
- **Screens**: `IMG_7703`, `IMG_7705`, `IMG_7707`, `IMG_7712`, `IMG_7731`.
- **Description**: Modal bottom sheet container with backdrop blur, drag handle, and spring animation.
- **Props Interface**:
  ```typescript
  export interface BottomSheetContainerProps {
    isOpen: boolean;
    onClose: () => void;
    children: React.ReactNode;
    header?: React.ReactNode; // ModalHeader
    maxHeight?: string; // Default: "92vh"
  }
  ```
- **CSS Rules**:
  - Backdrop: `background: rgba(0, 0, 0, 0.45); backdrop-filter: blur(4px);`
  - Sheet Container: `background: #f2f3f8; border-radius: 28px 28px 0 0; overflow: hidden;`
  - Drag handle: Width `36px`, height `4px`, border-radius `2px`, background `#d1d1d6`, margin `8px auto 4px auto`.

---

### `TooltipPopover`
- **Screens**: `IMG_7725`.
- **Description**: Dark floating tooltip popover explaining non-editable system categories.
- **Props Interface**:
  ```typescript
  export interface TooltipPopoverProps {
    isOpen: boolean;
    text: string; // "Đây là danh mục của hệ thống nên bạn không thể chỉnh sửa hoặc xóa bỏ."
    targetRect?: DOMRect;
    onClose: () => void;
  }
  ```
- **Styling**: Background `#22242a`, text `#ffffff`, border-radius `10px`, font-size `12px`, line-height `1.4`, padding `8px 12px`, box-shadow `0 4px 14px rgba(0,0,0,0.25)`.
