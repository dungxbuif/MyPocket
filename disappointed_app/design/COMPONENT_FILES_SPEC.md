# MyPocket - File-by-File Base Components Specification

> **Purpose**: This document serves as the self-contained, definitive engineering guide for all base UI components in MyPocket. It is designed so that any engineer can implement pixel-perfect, interactive components **without requiring access to the original screenshot images**.

---

## Architecture & Directory Structure

```
frontend/src/components/
├── layout/
│   ├── AppShell.tsx                     # Viewport, safe-areas, status bar
│   └── BottomSheet.tsx                  # Sliding modal sheet with backdrop blur
├── navigation/
│   ├── BottomTabBar.tsx                 # 5-slot floating tab bar + raised center '+'
│   ├── SubpageHeader.tsx                # Secondary screen header (< Back, Title, Actions)
│   ├── ModalHeader.tsx                  # Modal/Sheet header (Cancel/Done buttons)
│   ├── WalletFilterChip.tsx             # Interactive wallet pill selector (🌐 Tổng cộng)
│   └── SegmentedControl.tsx             # Sliding pill tabs (2-way & 3-way switchers)
├── cards/
│   ├── GroupedCard.tsx                  # White container card with subtle dividers
│   ├── FormRowItem.tsx                  # Interactive menu/input row with chevron
│   ├── SwitchRow.tsx                    # iOS boolean toggle row with description
│   ├── RadioCheckItem.tsx               # Selection row with active checkmark (✓)
│   └── DestructiveActionRow.tsx         # Red-highlighted destructive operation row
├── inputs/
│   ├── AmountInputHero.tsx              # Large 34px currency input with VND badge
│   ├── DateNavigationRow.tsx            # < [Day, DD/MM/YYYY] > step navigation
│   ├── SearchInputField.tsx             # Pill search bar with clear button
│   └── PhotoAttachmentPicker.tsx        # Attachment list + photo source sheet
├── buttons/
│   ├── PrimaryButton.tsx                # Green pill CTA button (#2dbd4f)
│   ├── SecondaryButton.tsx              # Soft green pill button (#e8f7ed)
│   └── CircleIconButton.tsx             # Circular touch target with badge support
├── finance/
│   ├── TransactionRow.tsx               # Category icon, wallet badge, amount, note
│   ├── WalletRow.tsx                    # Institution logo, name, VND balance
│   └── CategoryTreeRow.tsx              # Indented parent-child tree + lock 🔒 icon
├── charts/
│   ├── GaugeArcSummary.tsx              # Semi-circular budget gauge + 3-stat footer
│   ├── ProgressBarWithMarker.tsx        # Spend progress bar with 'Hôm nay' flag
│   ├── ComparisonBarChart.tsx           # Period comparison bars (Tuần/Tháng trước vs này)
│   ├── TrendAreaLineChart.tsx           # Cumulative spending curve + 3-month baseline
│   └── DonutChartBreakdown.tsx          # Expense category donut ring + center badge
└── feedback/
    ├── EmptyStateView.tsx               # 3D illustration / emoji placeholder
    ├── NoticeBanner.tsx                 # Full-width blue dismissible alert bar
    └── TooltipPopover.tsx               # Floating dark tooltip for system lock info
```

---

## 1. Layout Components

### File: `frontend/src/components/layout/AppShell.tsx`

#### Visual & Layout Description
- Global application wrapper designed for mobile viewports (`375px` to `440px`).
- Background color is canvas gray `#f2f3f8`.
- Seamlessly handles top status bar padding using `env(safe-area-inset-top)` and bottom bar clearance using `env(safe-area-inset-bottom)`.
- Prevents horizontal scrollbars (`overflow-x: hidden`).

#### DOM Anatomy
```html
<div class="app-shell">
  <div class="app-shell__status-bar-spacer" />
  <main class="app-shell__content">
    {children}
  </main>
  {bottomBar && <div class="app-shell__bottom-bar-slot">{bottomBar}</div>}
</div>
```

#### TypeScript Interface & Props
```typescript
export interface AppShellProps {
  children: React.ReactNode;
  bottomBar?: React.ReactNode;
  backgroundColor?: string; // Default: '#f2f3f8'
  className?: string;
}
```

#### CSS Specification
```css
.app-shell {
  min-height: 100dvh;
  max-width: 440px;
  margin: 0 auto;
  background-color: var(--bg-app, #f2f3f8);
  display: flex;
  flex-direction: column;
  position: relative;
  overflow-x: hidden;
  box-sizing: border-box;
}

.app-shell__content {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: max(12px, env(safe-area-inset-top)) 16px max(90px, calc(env(safe-area-inset-bottom) + 76px));
}

.app-shell__bottom-bar-slot {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  max-width: 440px;
  margin: 0 auto;
  pointer-events: none;
  z-index: 100;
}

.app-shell__bottom-bar-slot > * {
  pointer-events: auto;
}
```

---

### File: `frontend/src/components/layout/BottomSheet.tsx`

#### Visual & Layout Description
- Sliding modal card anchoring to the bottom of the screen.
- Top corners are heavily rounded with a `28px` radius.
- Background is canvas gray `#f2f3f8` or pure white `#ffffff`.
- Features a top center pill-shaped drag handle: width `36px`, height `4px`, color `#d1d1d6`, rounded `2px`.
- Backdrop dimming layer: `rgba(0, 0, 0, 0.45)` with `backdrop-filter: blur(4px)`.
- Maximum height is `92vh`, scrollable content within.

#### DOM Anatomy
```html
<div class="bottom-sheet-backdrop" onClick={onClose}>
  <div class="bottom-sheet" onClick={(e) => e.stopPropagation()}>
    <div class="bottom-sheet__handle-bar">
      <span class="bottom-sheet__handle" />
    </div>
    {header && <div class="bottom-sheet__header">{header}</div>}
    <div class="bottom-sheet__body">{children}</div>
  </div>
</div>
```

#### TypeScript Interface & Props
```typescript
export interface BottomSheetProps {
  isOpen: boolean;
  onClose: () => void;
  header?: React.ReactNode; // Typically ModalHeader
  children: React.ReactNode;
  maxHeight?: string; // Default: '92vh'
}
```

#### CSS Specification
```css
.bottom-sheet-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
  backdrop-filter: blur(4px);
  -webkit-backdrop-filter: blur(4px);
  z-index: 1000;
  display: flex;
  align-items: flex-end;
  justify-content: center;
  animation: fadeIn 0.2s ease-out;
}

.bottom-sheet {
  width: 100%;
  max-width: 440px;
  max-height: 92vh;
  background: #f2f3f8;
  border-radius: 28px 28px 0 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  box-shadow: 0 -8px 30px rgba(0, 0, 0, 0.15);
  animation: slideUp 0.28s cubic-bezier(0.2, 0.9, 0.3, 1);
}

.bottom-sheet__handle-bar {
  width: 100%;
  display: flex;
  justify-content: center;
  padding: 8px 0 4px;
}

.bottom-sheet__handle {
  width: 36px;
  height: 4px;
  background: #d1d1d6;
  border-radius: 2px;
}

.bottom-sheet__body {
  flex: 1;
  overflow-y: auto;
  padding: 12px 16px calc(24px + env(safe-area-inset-bottom));
}

@keyframes slideUp {
  from { transform: translateY(100%); }
  to { transform: translateY(0); }
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}
```

---

## 2. Navigation Components

### File: `frontend/src/components/navigation/BottomTabBar.tsx`

#### Visual & Layout Description
- Floating pill-shaped bar positioned `16px` above the bottom edge.
- Background: Semi-transparent white `rgba(255, 255, 255, 0.94)` with `backdrop-filter: blur(16px)` and subtle shadow `0 4px 24px rgba(0, 0, 0, 0.08)`.
- Height: `64px`, border-radius: `9999px`.
- Contains 5 items in a horizontal flex layout:
  1. `Tổng quan` (Home icon)
  2. `Sổ giao dịch` (Wallet icon)
  3. **Center Button**: Green `#2dbd4f` circle, diameter `52px`, elevated with `box-shadow: 0 6px 16px rgba(45, 189, 79, 0.4)`. Contains a centered white `+` icon.
  4. `Ngân sách` (Ledger/Target icon)
  5. `Tài khoản` (User/Gear icon)
- Active Tab state: Enclosed within a light gray pill container (`#eef0f4`), icon and label colored `#111111` with bold weight.
- Inactive Tab state: Transparent, icon and label colored muted gray `#8e8e93`.

#### DOM Anatomy
```html
<nav class="bottom-tab-bar" aria-label="Bottom Navigation">
  <button class="tab-item active" onClick={() => onTabSelect('overview')}>
    <div class="tab-item__pill">
      <svg class="tab-item__icon" />
      <span class="tab-item__label">Tổng quan</span>
    </div>
  </button>
  <button class="tab-item" onClick={() => onTabSelect('transactions')}>
    <div class="tab-item__pill">
      <svg class="tab-item__icon" />
      <span class="tab-item__label">Sổ giao dịch</span>
    </div>
  </button>

  <button class="tab-item__center-add" aria-label="Thêm giao dịch" onClick={onAddClick}>
    <span class="center-add__glyph">+</span>
  </button>

  <button class="tab-item" onClick={() => onTabSelect('budgets')}>
    <div class="tab-item__pill">
      <svg class="tab-item__icon" />
      <span class="tab-item__label">Ngân sách</span>
    </div>
  </button>
  <button class="tab-item" onClick={() => onTabSelect('account')}>
    <div class="tab-item__pill">
      <svg class="tab-item__icon" />
      <span class="tab-item__label">Tài khoản</span>
    </div>
  </button>
</nav>
```

#### TypeScript Interface & Props
```typescript
export type TabId = 'overview' | 'transactions' | 'budgets' | 'account';

export interface BottomTabBarProps {
  activeTab: TabId;
  onTabSelect: (tab: TabId) => void;
  onAddClick: () => void;
}
```

#### CSS Specification
```css
.bottom-tab-bar {
  margin: 0 16px max(16px, env(safe-area-inset-bottom));
  height: 64px;
  background: rgba(255, 255, 255, 0.94);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  border-radius: 9999px;
  display: flex;
  align-items: center;
  justify-content: space-around;
  padding: 0 8px;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.08);
  border: 1px solid rgba(255, 255, 255, 0.6);
}

.tab-item {
  background: transparent;
  border: none;
  padding: 0;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  outline: none;
}

.tab-item__pill {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 2px;
  padding: 6px 12px;
  border-radius: 9999px;
  transition: all 0.2s ease;
}

.tab-item.active .tab-item__pill {
  background: #eef0f4;
}

.tab-item__icon {
  width: 20px;
  height: 20px;
  stroke: #8e8e93;
}

.tab-item.active .tab-item__icon {
  stroke: #111111;
}

.tab-item__label {
  font-size: 11px;
  font-weight: 500;
  color: #8e8e93;
}

.tab-item.active .tab-item__label {
  color: #111111;
  font-weight: 700;
}

.tab-item__center-add {
  width: 52px;
  height: 52px;
  border-radius: 50%;
  background: #2dbd4f;
  border: none;
  color: #ffffff;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  box-shadow: 0 6px 16px rgba(45, 189, 79, 0.4);
  margin-top: -12px;
  transition: transform 0.15s ease, background 0.15s ease;
}

.tab-item__center-add:active {
  transform: scale(0.92);
  background: #25a443;
}

.center-add__glyph {
  font-size: 30px;
  font-weight: 400;
  line-height: 1;
  margin-top: -2px;
}
```

---

### File: `frontend/src/components/navigation/SubpageHeader.tsx`

#### Visual & Layout Description
- Top header for secondary screens (My Wallets, Account Security, Categories, Settings).
- Height: `48px`.
- Left: Back pill button with `< Quay lại`, soft gray background `#eef0f4` or transparent, radius `9999px`, padding `6px 12px`.
- Center: Screen title (bold `17px`, text color `#111111`), optionally accompanied by a small wallet badge `🌐`.
- Right: Optional action button (e.g. search icon `🔍`, menu icon `≡`, or text `Thêm`).

#### TypeScript Interface & Props
```typescript
export interface SubpageHeaderProps {
  title: string;
  onBack: () => void;
  backLabel?: string; // Default: "Quay lại"
  centerExtra?: React.ReactNode; // e.g. WalletFilterChip
  rightAction?: {
    label?: string;
    icon?: React.ReactNode;
    onClick: () => void;
    disabled?: boolean;
  };
}
```

---

### File: `frontend/src/components/navigation/ModalHeader.tsx`

#### Visual & Layout Description
- Top bar of bottom sheets and modals.
- Height: `52px`, flex align center, justify space-between.
- Left: Dismiss pill button labeled `Huỷ` or `Đóng`, background `#eef0f4`, text `#111111`, font-size `14px`, font-weight `600`.
- Center: Modal title (`Thêm Giao Dịch`, `Chọn Ví`, `Chi tiết khoản chi`), bold `17px`.
- Right: Action button (`Xong`, `Sửa`, or Calendar icon button). `Xong` has green text `#2dbd4f` and bold weight.

#### TypeScript Interface & Props
```typescript
export interface ModalHeaderProps {
  title: string;
  onDismiss: () => void;
  dismissLabel?: string; // Default: "Huỷ"
  rightAction?: {
    label?: string; // "Xong", "Sửa", "Lưu"
    icon?: React.ReactNode;
    onClick: () => void;
    isPrimary?: boolean; // If true, green text
  };
}
```

---

### File: `frontend/src/components/navigation/WalletFilterChip.tsx`

#### Visual & Layout Description
- Compact rounded pill chip for switching active wallet scope.
- Height: `32px`, border-radius `9999px`.
- Background: `#eef0f4`, border `1px solid transparent`.
- Content: Leading globe `🌐` (or bank glyph), wallet label (`Tổng cộng`), and trailing selector arrows `▲▼` (color `#8e8e93`).
- Tap target: Opens `WalletPickerSheet`.

#### TypeScript Interface & Props
```typescript
export interface WalletFilterChipProps {
  walletName: string; // e.g. "Tổng cộng"
  icon?: React.ReactNode; // Default: 🌐
  onClick: () => void;
}
```

---

### File: `frontend/src/components/navigation/SegmentedControl.tsx`

#### Visual & Layout Description
- Pill-shaped switcher track with a sliding white pill selector.
- Track height: `36px` to `40px`, background `#e5e6eb`, border-radius `9999px`, padding `3px`.
- Selected segment: Solid white `#ffffff`, border-radius `9999px`, shadow `0 2px 6px rgba(0, 0, 0, 0.08)`.
- Text: Bold `13px`–`14px`, dark `#111111` when selected, muted `#6c6c70` when unselected.
- Transition: Smooth sliding animation (`transition: transform 0.2s cubic-bezier(0.4, 0, 0.2, 1)`).

#### TypeScript Interface & Props
```typescript
export interface SegmentOption<T extends string = string> {
  key: T;
  label: string;
}

export interface SegmentedControlProps<T extends string = string> {
  options: SegmentOption<T>[];
  selectedKey: T;
  onChange: (key: T) => void;
  size?: 'sm' | 'md' | 'lg';
}
```

---

## 3. Containers & Grouped Cards

### File: `frontend/src/components/cards/GroupedCard.tsx`

#### Visual & Layout Description
- Standard iOS-style grouped card container.
- Surface color: Pure white `#ffffff`.
- Border-radius: `24px` to `28px`.
- Padding: `14px 16px`.
- Box-shadow: `0 2px 10px rgba(0, 0, 0, 0.03)`.
- Optional uppercase section label above the card (e.g. `TÍNH VÀO TỔNG`, `HIỂN THỊ`), color `#8e8e93`, font-size `12px`, letter-spacing `0.5px`.
- Automatic thin dividers (`1px solid #e8e8ec`) between child rows.

#### TypeScript Interface & Props
```typescript
export interface GroupedCardProps {
  title?: string;
  headerAction?: {
    label: string; // e.g. "Xem tất cả"
    onClick: () => void;
  };
  children: React.ReactNode;
  className?: string;
}
```

---

### File: `frontend/src/components/cards/FormRowItem.tsx`

#### Visual & Layout Description
- Standard list row item used in settings, transaction inputs, and details.
- Minimum height: `52px`.
- Flex align center, justify space-between.
- Left: Optional circular icon container (`36px` circle), primary label (bold `15px`, color `#111111`), optional sublabel (`13px`, color `#8e8e93`).
- Right: Right-aligned current selection value (`15px`, color `#6c6c70` or `#111111`), followed by a subtle chevron `>` (`#c7c7cc`).
- Touch feedback: Background flash `#f8f9fa` on press.

#### TypeScript Interface & Props
```typescript
export interface FormRowItemProps {
  leadingIcon?: React.ReactNode;
  label: string;
  description?: string;
  value?: string | React.ReactNode;
  showChevron?: boolean; // Default: true
  onClick?: () => void;
  disabled?: boolean;
}
```

---

### File: `frontend/src/components/cards/SwitchRow.tsx`

#### Visual & Layout Description
- Row containing a setting label, multi-line explanation, and an iOS toggle switch.
- Left block: Bold title `15px`, subtext `13px` `#8e8e93` with line-height `1.35`.
- Right: iOS-style toggle switch:
  - Width: `51px`, height: `31px`, radius `15.5px`.
  - Active (checked): Track `#2dbd4f`.
  - Inactive (unchecked): Track `#e5e5ea`.
  - Knob: Pure white circle `27px` diameter, smooth spring slide.

#### TypeScript Interface & Props
```typescript
export interface SwitchRowProps {
  title: string;
  description?: string;
  checked: boolean;
  onChange: (checked: boolean) => void;
  disabled?: boolean;
}
```

---

### File: `frontend/src/components/cards/RadioCheckItem.tsx`

#### Visual & Layout Description
- Single-select item row used inside `WalletPickerSheet`.
- Left: Icon (globe or wallet logo) and wallet label.
- Right: Bold green checkmark (`✓`, color `#2dbd4f`, font-size `18px`, font-weight `700`) displayed only when `selected === true`.

#### TypeScript Interface & Props
```typescript
export interface RadioCheckItemProps {
  leadingIcon?: React.ReactNode;
  label: string;
  sublabel?: string;
  selected: boolean;
  onSelect: () => void;
}
```

---

### File: `frontend/src/components/cards/DestructiveActionRow.tsx`

#### Visual & Layout Description
- Dedicated row or standalone pill button for dangerous operations (Logout, Reset account, Delete account).
- Text color: Vibrant danger red `#ff5a66`, font-weight `600`, font-size `15px`.
- When rendered as a standalone pill: Background white `#ffffff`, border `1px solid #ffd2d6`, height `48px`.

#### TypeScript Interface & Props
```typescript
export interface DestructiveActionRowProps {
  label: string;
  onClick: () => void;
  variant?: 'row' | 'pill';
}
```

---

## 4. Input Components

### File: `frontend/src/components/inputs/AmountInputHero.tsx`

#### Visual & Layout Description
- Prominent hero numeric input row in transaction and budget sheets.
- Left: Rounded pill badge labeled `VND` (background `#eef0f4`, text `#111111`, bold `13px`, padding `4px 10px`).
- Right / Center: Large tabular numeric display (`34px` bold font, color `#111111`).
- Default zero state displays `0`.
- Supports live typing with thousands separator formatting: e.g. `250.000` or `45.075.000`.

#### TypeScript Interface & Props
```typescript
export interface AmountInputHeroProps {
  value: number;
  currency?: string; // Default: "VND"
  onChange: (value: number) => void;
  placeholder?: string; // Default: "0"
  autoFocus?: boolean;
}
```

---

### File: `frontend/src/components/inputs/DateNavigationRow.tsx`

#### Visual & Layout Description
- Inline date selection bar featuring step-decrement (`<`) and step-increment (`>`) buttons.
- Left: Calendar glyph icon.
- Decrement button `<`: Circular touch target, radius `50%`, background `#f2f3f8`.
- Center Date Pill: Background `#eef0f4`, text: `Chủ Nhật, 23/08/2026` (localized Vietnamese). Tapping triggers calendar modal.
- Increment button `>`: Circular touch target, radius `50%`, background `#f2f3f8`.

#### TypeScript Interface & Props
```typescript
export interface DateNavigationRowProps {
  currentDate: Date | string;
  onDateChange: (date: Date) => void;
  onOpenCalendar?: () => void;
}
```

---

### File: `frontend/src/components/inputs/SearchInputField.tsx`

#### Visual & Layout Description
- Full pill search bar.
- Height: `40px`, border-radius `9999px`, background `#eef0f4`.
- Left: Leading search icon `🔍` (`#8e8e93`).
- Center: Text input (`15px`, placeholder `Tìm kiếm...`).
- Right: Circular clear button `✕` visible when text is non-empty.

#### TypeScript Interface & Props
```typescript
export interface SearchInputFieldProps {
  value: string;
  placeholder?: string;
  onChange: (val: string) => void;
  onClear: () => void;
  autoFocus?: boolean;
}
```

---

### File: `frontend/src/components/inputs/PhotoAttachmentPicker.tsx`

#### Visual & Layout Description
- Button and preview list for receipt photos in `AddTransactionSheet`.
- Primary trigger: Pill button labeled `Thêm Hình Ảnh` with a camera icon.
- Tapping triggers a centered floating dialog card (`PhotoSourceActionDialog`) with 3 stacked pill buttons:
  1. `Mở Thư Viện Ảnh` (Open Photo Library)
  2. `Mở Máy ảnh` (Open Camera)
  3. `Huỷ` (Cancel, muted font)
- Thumbnails: `56px x 56px` rounded image previews with a red `✕` delete badge at the top-right corner.

#### TypeScript Interface & Props
```typescript
export interface PhotoAttachmentPickerProps {
  images: string[];
  onAddFromLibrary: () => void;
  onAddFromCamera: () => void;
  onRemove: (index: number) => void;
}
```

---

## 5. Buttons & Actions

### File: `frontend/src/components/buttons/PrimaryButton.tsx`

#### Visual & Layout Description
- Main primary CTA pill button (`Lưu`, `Tạo Ngân sách`, `Đăng ký ngay`).
- Height: `50px`, border-radius: `9999px`.
- Background: Green `#2dbd4f`, text color `#ffffff`, font-size `16px`, font-weight `600`.
- Active/pressed state: Background `#25a443`, scale `0.98`.
- Disabled state: Background `#e0e0e0`, text `#8e8e93`, cursor `not-allowed`.

#### TypeScript Interface & Props
```typescript
export interface PrimaryButtonProps {
  label: string;
  onClick: () => void;
  disabled?: boolean;
  loading?: boolean;
  fullWidth?: boolean; // Default: true
}
```

---

### File: `frontend/src/components/buttons/SecondaryButton.tsx`

#### Visual & Layout Description
- Secondary supporting pill button (`Dùng thử miễn phí`).
- Height: `44px` to `48px`, border-radius `9999px`.
- Background: Soft green `#e8f7ed`, text color `#1c8535`, font-weight `600`.

#### TypeScript Interface & Props
```typescript
export interface SecondaryButtonProps {
  label: string;
  onClick: () => void;
  disabled?: boolean;
}
```

---

### File: `frontend/src/components/buttons/CircleIconButton.tsx`

#### Visual & Layout Description
- Circular icon touch target (`40px x 40px` or `44px x 44px`).
- Used for back navigation, search, camera, help `?`, and notifications.
- Notification bell variant: Features an overlaid red badge `#ff5a66` at top right showing unread count `4`.
- Camera variant: Background `#2dbd4f`, white camera glyph.

#### TypeScript Interface & Props
```typescript
export interface CircleIconButtonProps {
  icon: React.ReactNode;
  onClick: () => void;
  badgeCount?: number;
  ariaLabel: string;
  variant?: 'default' | 'primary' | 'soft';
}
```

---

## 6. Financial Visualizations & Data Display

### File: `frontend/src/components/finance/TransactionRow.tsx`

#### Visual & Layout Description
- Item row representing an individual transaction log.
- Left: `44px` circular category icon (`#29495a` container) with a `16px` mini wallet emblem at bottom right.
- Center: Category title (bold `15px`), secondary line with payee/note and date/time (`13px`, `#8e8e93`).
- Right: Formatted currency amount in bold `15px`:
  - Expense: Red font `#ff5a66` with negative prefix: `-90.000 đ`.
  - Income: Blue font `#32a9df` with positive prefix: `+5.000.000 đ`.
  - Transfer: Dark neutral font.

#### TypeScript Interface & Props
```typescript
export interface TransactionRowProps {
  id: string;
  categoryName: string;
  categoryIcon: React.ReactNode;
  walletName?: string;
  walletIcon?: React.ReactNode;
  note?: string;
  amount: number;
  currency?: string; // Default: "VND"
  dateFormatted?: string;
  onClick?: () => void;
}
```

---

### File: `frontend/src/components/finance/WalletRow.tsx`

#### Visual & Layout Description
- Row displaying a wallet and its current balance.
- Left: Circular icon representing wallet type (credit card, bank logo, cash wallet).
- Center: Wallet name (`Ví tín dụng`, `Techcombank`, `Tiền Mặt`).
- Right: Balance in VND (`-3.711.104 đ`, `4.710.200 đ`, `84.000 đ`). Negative balances are formatted clearly with minus sign.

#### TypeScript Interface & Props
```typescript
export interface WalletRowProps {
  id: string;
  name: string;
  balance: number;
  icon: React.ReactNode;
  currency?: string; // Default: "VND"
  selected?: boolean;
  onClick?: () => void;
}
```

---

### File: `frontend/src/components/finance/CategoryTreeRow.tsx`

#### Visual & Layout Description
- Hierarchical category row in `CategoryManagementScreen`.
- Supports 2 levels: Parent and Indented Children.
- Child row indentation: `padding-left: 36px`.
- Visual L-shaped tree branch connecting line: Drawn with `::before` pseudo-element (`border-left: 1.5px solid #d1d1d6; border-bottom: 1.5px solid #d1d1d6; width: 16px; height: 50%`).
- Metadata subtitle: `Hoạt động trong 3 ví` (muted `12px` font).
- System locked icon 🔒: Displays a gray lock glyph. Tapping displays `TooltipPopover`.
- Trailing chevron `>`.

#### TypeScript Interface & Props
```typescript
export interface CategoryTreeRowProps {
  id: string;
  name: string;
  icon: React.ReactNode;
  isChild?: boolean;
  isLastChild?: boolean;
  walletCount?: number;
  isSystemLocked?: boolean;
  onLockedClick?: () => void;
  onClick?: () => void;
}
```

---

### File: `frontend/src/components/charts/GaugeArcSummary.tsx`

#### Visual & Layout Description
- Semi-circular SVG gauge arc visualizing remaining spendable budget proportion.
- Sweep angle: `200°`, stroke-width: `14px`, rounded stroke caps.
- Background track: Soft gray `#e9eaef`.
- Fill track: Vibrant green `#2dbd4f` (switches to yellow/orange if >80%, red if >100%).
- Center hero text:
  - Subtitle: `Số tiền bạn có thể chi` (`13px`, `#8e8e93`).
  - Hero amount: `45.075.000 đ` (bold `30px`, `#111111`).
- 3-column key metrics footer:
  - Col 1: `65 M đ` - `Tổng ngân sách`
  - Col 2: `19,92 M đ` - `Tổng đã chi`
  - Col 3: `12 ngày` - `Đến cuối tháng`

#### TypeScript Interface & Props
```typescript
export interface GaugeArcSummaryProps {
  availableAmount: number;
  totalBudget: number;
  totalSpent: number;
  daysRemaining: number;
  currency?: string;
}
```

---

### File: `frontend/src/components/charts/ProgressBarWithMarker.tsx`

#### Visual & Layout Description
- Horizontal progress bar displaying category budget exhaustion.
- Track height: `8px`, border-radius `4px`, background `#e9eaef`.
- Elapsed fill: Green `#2dbd4f` or custom category color.
- Sliding vertical flag badge: Labeled `Hôm nay` (Today) positioned at current day percentage.
- Row labels: Left `2.190.000 đ`, Right `Còn lại 5.810.000 đ`.

#### TypeScript Interface & Props
```typescript
export interface ProgressBarWithMarkerProps {
  spentPercent: number; // 0 to 100
  dayProgressPercent: number; // 0 to 100
  spentAmountFormatted: string;
  remainingAmountFormatted: string;
  color?: string; // Default: #2dbd4f
}
```

---

### File: `frontend/src/components/charts/ComparisonBarChart.tsx`

#### Visual & Layout Description
- Comparative vertical bar chart used in Dashboard monthly reports (`Tuần trước` vs `Tuần này`, `Tháng trước` vs `Tháng này`).
- Bar width: `36px`, border-radius: `8px 8px 0 0`.
- Previous period bar: Soft light red `#ffd2d6`.
- Current period bar: Solid high-contrast red `#ff5a66`.
- Baseline axis: Subtle `#e8e8ec` line with currency labels above bars.

#### TypeScript Interface & Props
```typescript
export interface BarItem {
  label: string;
  amount: number;
  amountFormatted: string;
  isCurrent?: boolean;
}

export interface ComparisonBarChartProps {
  data: BarItem[];
  height?: number; // Default: 160px
}
```

---

### File: `frontend/src/components/charts/TrendAreaLineChart.tsx`

#### Visual & Layout Description
- Cumulative spending curve compared against historical baseline.
- Baseline: Soft gray shaded band representing 3-month historical average.
- Actual spending: Solid red curve line (`#ff5a66`, `stroke-width: 3px`) with milestone dots at key calendar days.
- Interactive tooltip on touch displaying date and exact cumulative VND total.

#### TypeScript Interface & Props
```typescript
export interface TrendPoint {
  dateLabel: string;
  actualAmount: number;
  baselineAmount: number;
}

export interface TrendAreaLineChartProps {
  points: TrendPoint[];
}
```

---

### File: `frontend/src/components/charts/DonutChartBreakdown.tsx`

#### Visual & Layout Description
- Circular donut percentage ring used in `ExpenseDetailSheet`.
- Diameter: `180px`, ring stroke-width: `18px`.
- Segmented arcs colored by category.
- Center displays dominant category icon and percentage pill (`100%`, background `#eef0f4`, bold `12px`).

#### TypeScript Interface & Props
```typescript
export interface DonutSlice {
  key: string;
  label: string;
  percentage: number;
  color: string;
}

export interface DonutChartBreakdownProps {
  slices: DonutSlice[];
  centerIcon?: React.ReactNode;
  centerBadgeText?: string; // e.g. "100%"
}
```

---

## 7. Feedback, Overlays & Placeholders

### File: `frontend/src/components/feedback/EmptyStateView.tsx`

#### Visual & Layout Description
- Centered container displayed when transaction lists, budgets, or recurring items are empty.
- Illustration: 3D vector graphic or emoji (e.g. folded hands 🙌).
- Heading: Bold `16px`, color `#111111` (e.g. `"Nhóm chi tiêu nhiều nhất sẽ hiển thị ở đây"`).
- Subtitle: Regular `13px`, color `#8e8e93`.
- Optional CTA pill button.

#### TypeScript Interface & Props
```typescript
export interface EmptyStateViewProps {
  icon?: React.ReactNode;
  title: string;
  description?: string;
  actionButton?: {
    label: string;
    onClick: () => void;
  };
}
```

---

### File: `frontend/src/components/feedback/NoticeBanner.tsx`

#### Visual & Layout Description
- Blue informational banner anchored at the bottom of the Budgets screen.
- Background: `#2f80ed`, text color `#ffffff`, border-radius: `14px`.
- Margin: `12px 16px`, padding: `10px 14px`.
- Left: Informative message: `"Dữ liệu trên là ví dụ minh họa..."`.
- Right: Small pill dismiss button labeled `Tắt` (`padding: 4px 10px`, background `rgba(255, 255, 255, 0.2)`).

#### TypeScript Interface & Props
```typescript
export interface NoticeBannerProps {
  message: string;
  dismissLabel?: string; // Default: "Tắt"
  onDismiss: () => void;
}
```

---

### File: `frontend/src/components/feedback/TooltipPopover.tsx`

#### Visual & Layout Description
- Dark floating bubble popover displayed when tapping a locked system category 🔒.
- Background: Dark charcoal `#22242a`, text `#ffffff`, border-radius `10px`.
- Text: `"Đây là danh mục của hệ thống nên bạn không thể chỉnh sửa hoặc xóa bỏ."`.
- Arrow indicator pointing to target element.

#### TypeScript Interface & Props
```typescript
export interface TooltipPopoverProps {
  isOpen: boolean;
  text: string;
  targetRect?: DOMRect;
  onClose: () => void;
}
```
