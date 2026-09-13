---
name: Financial Clarity System
colors:
   surface: '#fbf9f9'
   surface-dim: '#dbdad9'
   surface-bright: '#fbf9f9'
   surface-container-lowest: '#ffffff'
   surface-container-low: '#f5f3f3'
   surface-container: '#efeded'
   surface-container-high: '#e9e8e7'
   surface-container-highest: '#e3e2e2'
   on-surface: '#1b1c1c'
   on-surface-variant: '#3f4a3c'
   inverse-surface: '#303031'
   inverse-on-surface: '#f2f0f0'
   outline: '#6f7a6b'
   outline-variant: '#becab9'
   surface-tint: '#006e1c'
   primary: '#006e1c'
   on-primary: '#ffffff'
   primary-container: '#4caf50'
   on-primary-container: '#003c0b'
   inverse-primary: '#78dc77'
   secondary: '#556158'
   on-secondary: '#ffffff'
   secondary-container: '#d9e6da'
   on-secondary-container: '#5b675e'
   tertiary: '#bb1614'
   on-tertiary: '#ffffff'
   tertiary-container: '#ff6c5c'
   on-tertiary-container: '#6d0003'
   error: '#ba1a1a'
   on-error: '#ffffff'
   error-container: '#ffdad6'
   on-error-container: '#93000a'
   primary-fixed: '#94f990'
   primary-fixed-dim: '#78dc77'
   on-primary-fixed: '#002204'
   on-primary-fixed-variant: '#005313'
   secondary-fixed: '#d9e6da'
   secondary-fixed-dim: '#bdcabe'
   on-secondary-fixed: '#131e17'
   on-secondary-fixed-variant: '#3e4a41'
   tertiary-fixed: '#ffdad5'
   tertiary-fixed-dim: '#ffb4a9'
   on-tertiary-fixed: '#410001'
   on-tertiary-fixed-variant: '#930005'
   background: '#fbf9f9'
   on-background: '#1b1c1c'
   surface-variant: '#e3e2e2'
typography:
   display-lg:
      fontFamily: Manrope
      fontSize: 32px
      fontWeight: '700'
      lineHeight: 40px
      letterSpacing: -0.02em
   headline-md:
      fontFamily: Manrope
      fontSize: 20px
      fontWeight: '600'
      lineHeight: 28px
   headline-sm:
      fontFamily: Manrope
      fontSize: 17px
      fontWeight: '600'
      lineHeight: 24px
   body-md:
      fontFamily: Manrope
      fontSize: 15px
      fontWeight: '400'
      lineHeight: 22px
   body-sm:
      fontFamily: Manrope
      fontSize: 13px
      fontWeight: '400'
      lineHeight: 18px
   label-lg:
      fontFamily: Manrope
      fontSize: 14px
      fontWeight: '600'
      lineHeight: 20px
   label-md:
      fontFamily: Manrope
      fontSize: 12px
      fontWeight: '500'
      lineHeight: 16px
   currency-display:
      fontFamily: Manrope
      fontSize: 24px
      fontWeight: '600'
      lineHeight: 32px
rounded:
   sm: 0.25rem
   DEFAULT: 0.5rem
   md: 0.75rem
   lg: 1rem
   xl: 1.5rem
   full: 9999px
spacing:
   container-margin: 1rem
   card-padding: 1.25rem
   gutter: 0.75rem
   stack-sm: 0.25rem
   stack-md: 0.5rem
   stack-lg: 1rem
---

## Product Direction

Financial Clarity is a mobile-first personal finance app for everyday spending control. The first implementation should help users record expenses quickly, understand wallet balances, and see whether current spending is still safe for the month.

The product borrows interaction patterns from Money Lover, but the UI should feel calmer, clearer, and more systematized. Data is the main visual subject. Decoration should stay quiet and should never compete with balances, transactions, budgets, or warnings.

## Current Implementation Baseline (2026-09-13)

The current source of truth is the React implementation in `app/src/atomic/`. It is a working, mock-data-first preview connected to the local authentication/profile/home API. The implementation preserves the Financial Clarity visual language: Manrope typography, the token colors above, soft neutral surfaces, rounded cards, mobile width capped at 430px, sticky header, bottom navigation and centered quick-add action.

Implemented component map:

| Layer | Components | Responsibility |
| --- | --- | --- |
| Atoms | `IconButton`, `MetricBox`, `SectionTitle`, `SegmentedControl` | Reusable controls, metrics and section headings |
| Molecules | `WalletCard`, `TransactionItem`, `BudgetProgressItem`, `GoalCard`, `CategoryTreeCard`, `FormSelectorRow` | Repeated financial rows/cards and form cells |
| Organisms | `AppHeader`, `BottomNavigation`, `OverviewPanel`, `TransactionsPanel`, `BudgetsPanel`, `ReportsPanel`, `AccountPanel`, `QuickAddSheet` | Composed screen sections and flows |
| Template/Page | `MobileAppShell`, `FinancePrototypePage` | App shell, authentication state and tab composition |

The overview currently starts with **Ví của bạn**. The header shows **Tổng số dư** with hide/show controls. The former overview total-balance hero card and the “Trang chủ / Ví / Note” notice card are intentionally removed from the current layout.

### Mandatory component reuse rule

Every new or changed screen must use an existing base component whenever the pattern already exists. If the required pattern does not exist, create or update the base component first, then use it from the screen or organism. Do not copy card, icon-badge, button, metric, row, progress or spacing markup into feature components.

Before adding JSX, check `app/src/atomic/atoms/` and `app/src/atomic/molecules/`. New shared patterns belong there; screen-specific composition belongs in `organisms/`. A refactor must update all consumers of a base component in the same change so visual behavior remains consistent.

### Current base-component gaps

The implementation still needs shared primitives for `SurfaceCard`, `IconBadge`, `ProgressBar`, `BaseButton`, `ScreenSection`, and currency display. Until those are introduced, the repeated styles currently visible in organisms and molecules are known technical debt, not separate design variants. New work must not increase that duplication.

**Primary outcome:** a user can open the app, add an expense, choose a category and wallet, save it, and immediately see the effect on overview, transactions, and budget progress.

## Design Principles

- **Fast capture first:** adding a transaction must be the shortest and most polished path in the app.
- **Numbers lead, labels support:** balances, amounts, remaining budget, and percentage changes get the strongest hierarchy.
- **Cards group decisions:** use cards for repeated data units and form groups, not for whole page sections.
- **Status is visible without drama:** warnings are clear, but the interface should not feel alarming unless the user is over budget or action is blocked.
- **Mobile ergonomics:** all primary controls need comfortable touch targets, stable layout, and bottom-safe spacing.

## Color Usage

Use the YAML tokens above as the source of truth.

| Role             | Token                                         | Use                                                        |
| ---------------- | --------------------------------------------- | ---------------------------------------------------------- |
| Primary action   | `primary` / `on-primary`                      | Save, continue, selected tabs, healthy budget progress     |
| Positive value   | `primary` or `inverse-primary`                | Income, savings, under-budget status                       |
| Warning/negative | `tertiary`, `error`                           | Expenses, over-budget status, destructive or failed states |
| App background   | `background`, `surface`                       | Screen base                                                |
| Card background  | `surface-container-lowest`                    | Main cards and sheets                                      |
| Muted surface    | `surface-container`, `surface-container-high` | Chips, secondary rows, input backgrounds                   |
| Text primary     | `on-surface`                                  | Headings, balances, transaction names                      |
| Text secondary   | `on-surface-variant`                          | Metadata, captions, helper labels                          |
| Borders          | `outline-variant`                             | Dividers, inactive controls, subtle component boundaries   |

Avoid introducing new greens or reds in app code unless a chart needs multiple category colors. If chart colors are needed, define them as chart-specific tokens instead of hardcoding them in components.

## Typography

Use **Manrope** across the app.

| Style              | Use                                                                       |
| ------------------ | ------------------------------------------------------------------------- |
| `display-lg`       | Current balance, safe-to-spend amount, empty-state headline when spacious |
| `currency-display` | Transaction amount input, card-level totals, budget amounts               |
| `headline-md`      | Screen title, sheet title, major card title                               |
| `headline-sm`      | Section title, card title, transaction group date                         |
| `body-md`          | Row title, form value, normal content                                     |
| `body-sm`          | Supporting text, notes, timestamps                                        |
| `label-lg`         | Button label, selected tab label                                          |
| `label-md`         | Field label, chip label, small metric caption                             |

Currency formatting must be consistent: use thousands separators and the configured currency suffix, for example `250.000 đ`. Negative values and expenses should use a clear sign or expense color, but never rely on color alone.

## Layout And Spacing

- Use a mobile-first shell with a bottom navigation bar and centered quick-add FAB.
- Main screen content uses `container-margin` on both sides and keeps the bottom safe area clear of navigation/FAB.
- Use vertical stacks with `gutter` between sibling cards and `card-padding` inside cards.
- Card internals follow an 8px rhythm: `stack-sm` for label/value pairs, `stack-md` for row text, `stack-lg` between logical groups.
- Prefer whitespace to dividers. Use dividers only inside dense lists or grouped form rows.
- Fixed-format controls such as keypad keys, tab segments, chart bars, and progress rows must have stable dimensions so state changes do not shift layout.

## Shape And Elevation

- Main cards use `rounded-xl`.
- Nested cards, inputs, and row containers use `rounded-lg`.
- Small chips and icon badges use `rounded-full`.
- Buttons are pill-shaped when they are primary page actions.
- Progress bars and chart arcs use fully rounded caps.
- Use soft shadows only for floating surfaces: bottom sheets, FAB, sticky action areas. Normal cards should rely mostly on surface contrast and spacing.

## Core Navigation

The app has four primary tabs plus one global quick action:

| Area          | Purpose                                                             | Primary components                                 |
| ------------- | ------------------------------------------------------------------- | -------------------------------------------------- |
| Overview      | Current balance, wallet summary, recent transactions, insight cards | Metric cards, wallet rows, transaction rows        |
| Transactions  | Search, filter, and review the ledger                               | Date headers, transaction rows, scope selector     |
| Budgets       | Understand monthly budget health and safe-to-spend amount           | Gauge, progress cards, warning states              |
| Account       | Wallets, settings, data controls                                    | Form rows, switches, destructive rows              |
| Quick Add FAB | Create transaction from anywhere                                    | Transaction form, amount keypad, category selector |

Reports can start as a secondary panel reachable from Overview or Transactions. It does not need to be a first MVP tab unless implementation capacity allows.

## Component Contracts

### Buttons

- **Primary:** `primary` background, `on-primary` text, pill shape, minimum 44px height.
- **Secondary:** tonal surface using `secondary-container`, text using `on-secondary-container`.
- **Ghost:** transparent background, used for cancel, back, or low-emphasis actions.
- Buttons must show disabled, pressed, loading, and error-adjacent states when relevant.

### Cards

- Cards contain one clear unit: one wallet, one budget, one metric, one form group, or one insight.
- Avoid placing a card inside another card.
- Each card should have a clear top-left identity: icon, title, or metric label.
- Financial cards should keep the main amount visually dominant.

### Transaction Row

Pattern: `[category icon] [title + metadata] [amount]`.

Required states:

- Expense, income, transfer/debt when supported.
- Pending/offline sync if the implementation includes offline mode.
- Pressed/selected state for opening details.
- Empty metadata should collapse cleanly without leaving awkward gaps.

### Amount Input And Keypad

The add-transaction flow uses a large amount display plus custom numeric keypad.

Required keys:

- Digits `0-9`
- Decimal or locale separator if supported
- Backspace
- Clear
- `+`, `-`, `x`, `/` calculator operations
- Quick amount chips such as `50.000`, `150.000`, `500.000`
- Done/save action

The amount display must be readable at a glance and must not resize the keypad.

### Form Row Item

Use for category, wallet, date, event, note, receipt, and report-inclusion settings.

Pattern: `[leading icon] [label/value stack] [optional trailing control]`.

Rows must support:

- Placeholder state
- Filled state
- Error/helper text
- Chevron for navigation
- Switch or checkmark for binary choices

### Category Tree

Use nested rows for parent and child categories.

- Parent categories can expand/collapse.
- Child rows align under the parent text block, not under the parent icon.
- Tree connector lines are subtle and never dominate the row.
- Icons remain fixed-size so indentation does not cause layout jitter.

### Budget Gauge And Progress Cards

Budget surfaces must communicate three things quickly:

- total budget
- spent amount
- safe-to-spend amount or remaining amount

Progress color rules:

- Under budget: primary green.
- Near limit: use a caution treatment if a caution token is added; until then, use neutral text plus clear copy.
- Over budget: tertiary/error.

### Charts

Charts are secondary to direct decisions. Use them for comparison and trend explanation, not decoration.

- Bar charts use rounded tops.
- Current period uses primary or alert color depending on metric meaning.
- Previous/reference periods use muted surfaces.
- Donut charts must include a text summary beside or inside the chart.
- Tooltips and legends must be readable on mobile.

### Icons

- Functional navigation icons should be simple stroke icons with consistent weight.
- Category icons can be more colorful, but they must remain legible at small sizes.
- Do not use icon color alone to indicate financial status.

## MVP Screen Priority

Build in this order:

1. **App Shell:** mobile layout, bottom navigation, quick-add FAB, shared header patterns.
2. **Quick Add Transaction:** amount input, keypad, transaction type tabs, category selector, wallet selector, date row, save action.
3. **Overview:** total balance, wallets, recent transactions, simple insight/notice banner.
4. **Transactions:** ledger list grouped by date with filters.
5. **Budgets:** safe-to-spend gauge and category progress cards.
6. **Reports:** comparison bars, trend line, donut breakdown.

This order keeps the first usable slice focused on the highest-frequency workflow.

## Implementation Notes

- Treat this file as the design source of truth for theme tokens and component behavior.
- Use `design/INDEX.md` as the component inventory and product PRD companion.
- Use the existing HTML component experiments under `design/` as visual references, not production code.
- If building from `refereces/disappointed_app/`, map existing components to these contracts before creating new variants.
- The implementation baseline above takes precedence over stale prototype wording elsewhere in this document. Keep the current layout unless a product/design decision explicitly changes it.
- Base-component reuse is mandatory: create or update the shared component before consuming a new visual pattern, and keep tokens/styles centralized wherever practical.
- Any change to core token names, navigation structure, or MVP order should be recorded in Harness docs.

## Acceptance Checklist

- [ ] Theme tokens from the YAML frontmatter are available in the app styling layer.
- [ ] Mobile app shell has bottom navigation and quick-add FAB.
- [ ] Add transaction flow can be completed without layout jumps.
- [ ] Amounts are formatted consistently.
- [ ] Budget status is understandable without relying on color alone.
- [ ] Transaction rows, form rows, cards, charts, and category tree follow the contracts above.
- [ ] Empty, loading, error, pressed, disabled, and over-budget states are represented where relevant.
