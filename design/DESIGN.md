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

## App Initialization Spec

MyPocket is a mobile-first personal finance app that recreates the useful product patterns of Money Lover while using its own brand, design tokens, information architecture, and implementation. The goal is not a visual skin or brand copy. The goal is feature parity for daily money management, then a stronger personal layer for automation, insight, and owner-specific workflows.

The app must keep the full scope: wallets, transactions, budgets, reports, category hierarchy, recurring items, debt/loan tracking, receipt/photo capture, search/filtering, settings, and personal extensions. Scope is organized by phases so implementation can start without deleting later ambitions.

## Product North Star

A user should be able to understand their money position in seconds, add a transaction in under five seconds, and see the effect on balances, budgets, reports, and insight surfaces immediately.

Core success signals:

- Fast expense capture with minimal taps.
- Accurate wallet and category accounting.
- Clear budget health and safe-to-spend guidance.
- Useful reports that explain what changed, not just what happened.
- Personal automation that reduces manual tracking over time.

## Experience Principles

- **Clone the workflow, not the brand:** reuse proven Money Lover-style flows, but keep MyPocket visual identity and copy.
- **Capture first:** Quick Add is the most important interaction in the app.
- **Every number needs context:** show period, wallet scope, category, trend, or status near financial values.
- **Progress over dashboards:** surfaces should answer what the user can do next.
- **Personal but inspectable:** AI/OCR/automation must show what it inferred and allow manual correction.
- **Offline-tolerant:** finance tracking should keep working during weak network conditions when feasible.

## Product Scope

### Money Management Core

| Module | Required capabilities |
| --- | --- |
| Wallets | Cash, bank, card, savings, debt wallet, archived wallet, balance adjustment |
| Transactions | Expense, income, transfer, debt/loan, note, date, wallet, category, receipt, exclude-from-report |
| Categories | Parent/child categories, icons, colors, expense/income category sets |
| Budgets | Monthly/category budgets, progress, safe-to-spend, over-budget warnings |
| Reports | Category breakdown, period comparison, daily bars, cumulative trend, wallet/category filters |
| Search | Keyword search, amount/date/category/wallet filters, grouped results |
| Recurring | Recurring income/expense, reminder, generated transaction history |
| Settings | Currency, locale, privacy, passcode/biometric slot, backup/export/import |

### Personal Extensions

Keep these in the product design even if implementation is phased:

| Extension | Intent |
| --- | --- |
| Smart categorization | Suggest category, wallet, and note from history or receipt text |
| Receipt OCR | Capture receipt photo and extract merchant, total, date, line hints |
| Cashflow forecast | Estimate end-of-month balance from recurring items, budgets, and recent pace |
| Safe-to-spend coach | Convert monthly budget state into daily allowance and warnings |
| Debt reminders | Track borrow/lend items, due dates, partial payments, and reminders |
| Goals and funds | Track savings goals, sinking funds, and progress toward planned purchases |
| Travel mode | Separate travel spending by trip, currency, and temporary budget |
| Family/shared wallet ready | Design data boundaries so shared wallets can be added later |
| Investment/portfolio ready | Leave navigation and data model room for assets without blocking MVP |

## Information Architecture

```text
MyPocket
|
+-- Overview
|   +-- Total balance and privacy toggle
|   +-- Wallet summary
|   +-- Safe-to-spend / Money Insight
|   +-- Recent transactions
|   +-- Report preview
|
+-- Transactions
|   +-- Ledger grouped by date
|   +-- Search and filters
|   +-- Transaction detail
|   +-- Edit / duplicate / delete
|
+-- Quick Add
|   +-- Expense / Income / Transfer / Debt
|   +-- Amount keypad and calculator
|   +-- Category selector
|   +-- Wallet selector
|   +-- Date, note, receipt, exclude-from-report
|
+-- Budgets
|   +-- Safe-to-spend gauge
|   +-- Category budget cards
|   +-- Budget setup and edit
|   +-- Over-budget details
|
+-- Reports
|   +-- Category donut
|   +-- Period comparison bars
|   +-- Cumulative trend
|   +-- Daily breakdown
|
+-- Account
    +-- Wallet management
    +-- Recurring items
    +-- Debts and loans
    +-- Goals and funds
    +-- Import/export and privacy settings
```

## Navigation Model

Use four persistent tabs plus one global center FAB:

| Surface | Role |
| --- | --- |
| Overview | Default home and financial status |
| Transactions | Ledger, search, review, corrections |
| Budgets | Monthly control and safe-to-spend |
| Account | Wallets, settings, recurring, goals, data controls |
| Center FAB | Quick Add from any main tab |

Reports start as an Overview/Transactions entry point and can become a fifth tab only when report usage justifies the extra navigation weight.

## Design System Rules

Use the YAML tokens in this file as the source of truth.

### Color Roles

| Role | Token | Use |
| --- | --- | --- |
| Primary action | `primary`, `on-primary` | Save, selected tab, healthy progress |
| Positive value | `primary`, `inverse-primary` | Income, savings, under-budget |
| Negative/warning | `tertiary`, `error` | Expenses, over-budget, destructive state |
| App background | `background`, `surface` | Screen background |
| Card surface | `surface-container-lowest` | Wallet, budget, report, and form cards |
| Muted surface | `surface-container`, `surface-container-high` | Chips, inactive controls, nested rows |
| Text primary | `on-surface` | Amounts, titles, critical labels |
| Text secondary | `on-surface-variant` | Metadata, helper text, captions |
| Border | `outline-variant` | Dividers and subtle component boundaries |

Do not hardcode alternate greens/reds in components. Add chart tokens later if reports need more categorical colors.

### Typography

Use **Manrope** everywhere.

| Token | Use |
| --- | --- |
| `display-lg` | Total balance, safe-to-spend hero, major empty state |
| `currency-display` | Amount input, wallet balances, budget totals |
| `headline-md` | Screen and sheet titles |
| `headline-sm` | Card title, section title, date group title |
| `body-md` | Row value, form text, transaction title |
| `body-sm` | Metadata, note preview, helper text |
| `label-lg` | Button label, selected tab |
| `label-md` | Field label, chip, metric caption |

Currency formatting must be consistent, for example `250.000 đ`. Negative values need a sign, label, or context; color alone is not enough.

### Layout

- Mobile-first viewport with safe-area support.
- Main content uses `container-margin`.
- Cards use `card-padding`.
- Sibling content uses `gutter`.
- Internal spacing uses `stack-sm`, `stack-md`, and `stack-lg`.
- Keypads, segmented controls, tab bars, chart slots, and progress rows must have stable dimensions.
- Do not nest UI cards inside other cards.

### Shape And Motion

- Main cards: `rounded-xl`.
- Nested rows and inputs: `rounded-lg`.
- Chips, badges, icon containers, progress caps: `rounded-full`.
- Use soft elevation only for FAB, bottom sheets, sticky actions, and modal overlays.
- Motion should clarify hierarchy: sheet enters from bottom, row press is subtle, chart transitions are quick.

## Component Inventory

The system keeps all 28 base components described in `design/INDEX.md`, grouped into implementation families.

### Navigation And Structure

- Top Bar / Screen Header
- Modal or Sheet Header
- Bottom Navigation Bar
- Center Quick Add FAB
- Segmented Control
- Sub-header Scope Selector

### Inputs And Controls

- Amount Display Input
- Custom Numeric Keypad with calculator operations
- Quick Amount Chips
- Selector Row Item
- Date Picker Row
- Switch Row
- Primary Action Button

### Lists And Rows

- Wallet List Item
- Transaction List Item
- Category Tree Item
- Section Date Header
- Search Result Row

### Cards And Data Display

- Overview Metric Card
- Total Balance Card with privacy toggle
- Budget Progress Card
- Semi-circle Budget Gauge
- Stat Metric Card
- Notice or Insight Banner

### Charts

- Bar Comparison Chart
- Donut Category Chart
- Trend Line Chart
- Daily Breakdown Bar Chart

### Modals And Overlays

- Action Sheet
- Calendar Date Picker Modal
- Context Dropdown Menu
- Wallet Selection Bottom Sheet
- Category Selection Bottom Sheet
- Transaction Detail Sheet

### Badges And Micro-elements

- Status / Percentage Badge
- Category Icon Badge
- Sync / Pending Indicator
- Over-budget Indicator

## Core Flow Contracts

### Quick Add Transaction

Required flow:

1. User taps center FAB.
2. Sheet opens with transaction type segmented control.
3. Amount input is focused by default.
4. User enters or calculates amount.
5. User chooses category, wallet, date, optional note, optional receipt.
6. User saves.
7. App updates wallet balance, transaction ledger, budget progress, and report aggregates.

Required states:

- Empty amount
- Invalid amount
- Missing category or wallet
- Expense, income, transfer, debt/loan
- Receipt attached
- Excluded from reports
- Save loading, success, failure, offline queued

### Category Selection

- Parent categories can expand/collapse.
- Child categories align under the parent text block.
- Tree lines are subtle and fixed.
- Search filters both parent and child categories.
- Recent or suggested categories can appear above the tree.

### Budget Health

- Show total budget, spent, remaining, days left, and safe-to-spend.
- Category budgets show icon, title, limit, spent, progress, and status.
- Over-budget state must use text and color together.
- Near-limit state should be calm and actionable.

### Reports

- Reports always respect selected wallet scope and date period.
- Comparison charts show current period vs previous period.
- Donut charts show category share and the top category summary.
- Trend charts show cumulative spending and comparison baseline.
- Empty report states explain which data is missing.

## Data Domains

Initial implementation should keep these domains explicit:

| Domain | Core entities |
| --- | --- |
| Identity | User, local profile, privacy settings |
| Wallet | Wallet, balance adjustment, wallet group |
| Ledger | Transaction, transfer, debt/loan transaction, receipt attachment |
| Category | Category, category parent, category type, category icon |
| Budget | Budget, budget period, budget category allocation |
| Planning | Recurring item, reminder, goal, fund |
| Analytics | Report period, aggregate, insight, forecast |
| Sync | Local change, pending queue, conflict state |

This design does not require all domains in the first build, but UI and architecture should not block them.

## Personal Feature Backlog

Keep these as first-class roadmap items:

1. AI category suggestion from user history.
2. Receipt OCR and attach-to-transaction workflow.
3. Cashflow forecast and end-of-month projection.
4. Safe-to-spend daily coach.
5. Debt/loan reminders and partial repayment tracking.
6. Savings goals and sinking funds.
7. Travel mode with temporary budget and optional currency handling.
8. Shared wallet/family mode.
9. Investment/asset tracking.
10. Import/export for backup and migration.

## Launch Phases

### Phase 1: Money Lover Core Shell

- App shell, bottom navigation, FAB.
- Overview with total balance, wallets, recent transactions.
- Quick Add for expense/income.
- Basic wallet and category data.
- Transaction ledger grouped by date.

### Phase 2: Budget And Reports

- Category budgets.
- Safe-to-spend gauge.
- Report preview and full reports.
- Period filters and wallet scope selector.

### Phase 3: Full Money Management

- Transfers.
- Debt/loan tracking.
- Recurring transactions.
- Receipt attachment.
- Search and advanced filters.
- Export/import.

### Phase 4: Personal Intelligence

- AI categorization.
- OCR receipt extraction.
- Forecasting.
- Personalized insight cards.
- Goal/fund guidance.

### Phase 5: Expansion Readiness

- Shared wallets.
- Travel mode.
- Investment/portfolio surfaces.
- Multi-device sync and conflict UX.

## Implementation Notes

- Treat this file as the design and product source of truth for app start.
- Use `design/INDEX.md` as the detailed component inventory and PRD companion.
- Use existing HTML experiments under `design/` as visual references, not production source.
- If building from `refereces/disappointed_app/`, map existing screens/components to this spec before creating new variants.
- Record major changes to navigation, data domains, or phase order in Harness docs.

## Acceptance Checklist

- [ ] App shell supports Overview, Transactions, Budgets, Account, and center Quick Add.
- [ ] Quick Add can create expense and income records with category, wallet, date, note, and amount.
- [ ] Wallet balances, transaction ledger, budgets, and reports update from the same ledger source.
- [ ] All 28 base component categories remain represented in design or implementation roadmap.
- [ ] Budget status is understandable without color alone.
- [ ] Reports support wallet/date scope and empty states.
- [ ] Personal extensions are preserved as roadmap items even when not in the first release.
- [ ] Theme tokens from this file are available in the app styling layer.
