# MoneyLover Complete Feature Research

> Historical research (2026-08-31), not implementation proof. Superseded for current comparison by the [2026-09-08 audit](PARITY-AUDIT-2026-09-08.md), [96-article catalog](SOURCE-CATALOG-2026-09-08.md) and [code/UI feature matrix](FEATURE-PARITY-2026-09-08.md). Source conflicts and retired features must be resolved there before implementation.

> Crawled from moneylover.zendesk.com/hc/en-us  
> Date: 2026-08-31  
> Purpose: Reference for MyPocket implementation gap analysis

---

## Table of Contents

1. [Wallet Types & Management](#1-wallet-types--management)
2. [Transactions](#2-transactions)
3. [Categories](#3-categories)
4. [Budgets](#4-budgets)
5. [Reports](#5-reports)
6. [Home Screen](#6-home-screen)
7. [Money Insider](#7-money-insider)
8. [Debt & Loan](#8-debt--loan)
9. [Goals / Savings](#9-goals--savings)
10. [Credit Cards](#10-credit-cards)
11. [Shared Wallets](#11-shared-wallets)
12. [Linked Bank Wallets](#12-linked-bank-wallets)
13. [Export](#13-export)
14. [Premium Features](#14-premium-features)
15. [AI & Auto Features](#15-ai--auto-features)
16. [Widget & UI](#16-widget--ui)
17. [Default Categories](#17-default-categories)
18. [Category Groups & Hierarchy](#18-category-groups--hierarchy)

---

## 1. Wallet Types & Management

### 1.1 Wallet Types

| Type | Description | Premium? |
|------|-------------|----------|
| **Basic Wallet** | Manual income/expense recording. Create multiple wallets for different funding sources (cash, bank account). Free users get limited count. | Unlimited for Premium |
| **Goal Wallet** | Manual savings tracking with savings goals and deposit recording. No interest calculation supported. | Yes |
| **Linked Wallet** | Auto-sync transaction history from bank accounts. Read-only mode (RSA-2048 encrypted). Updates every 6-8 hours. Separate subscription from Premium. | Subscription |
| **Credit Wallet** | Manual credit card transaction recording. Tracks credit limits, balances, due dates. | Yes |
| **Total Wallet** | Aggregation of wallets marked "include from total". Sum of all included wallet balances (credit cards shown as negative). | Premium for reports |

### 1.2 Wallet CRUD

- Create: Account tab > My Wallets > New wallet
- Edit: Name, type, initial balance, currency
- Archive: Hide without deleting
- Delete: Permanent removal
- Sort order: Drag-and-drop reordering in wallet list
- Default wallet: Long-press wallet > Set as Default Wallet (used by AI assistant)

### 1.3 Wallet Balance Display Modes

- Opening/Ending Balance mode
- Inflow/Outflow mode
- Configurable in Account > Settings > Overview Display Mode

**Formulas:**
- Opening Balance = Total income before period - Total expense before period
- Ending Balance = Total income up to period end - Total expense up to period end
- Inflow = Income + Transactions excluded from report
- Outflow = Expense + Transactions excluded from report

### 1.4 Adjust Balance

- Feature to reconcile actual money vs app balance
- Enter actual amount, app auto-adjusts (adds or subtracts)
- Useful when forgetting to record transactions

### 1.5 Transfer Between Wallets

**Access:** Transactions tab > three-dot menu > "Transfer money"

**Form fields:**
- Amount
- From wallet (nguồn)
- Category (auto-filled = "Outgoing Transfer", có thể đổi)
- Note
- To wallet (đích)
- Date

**Behavior:**
- Sau khi save, tự động tạo **2 transactions**: 1 expense (outgoing) + 1 income (incoming)
- Transfer transactions auto-excluded from report (có thể tắt exclude)

**Transfer Fee:**
- Có thể thêm phí giao dịch (ví dụ: phí rút ATM, phí chuyển tiền)
- Khi có fee, app tự động tạo **3 transactions**: chuyển tiền, nhận tiền, và phí
- Ví dụ: rút tiền từ ATM có phí → app ghi nhận: rút tiền + nhận tiền mặt + phí ATM

---

## 2. Transactions

### 2.1 Create Transaction

**Required fields:**
- Amount
- Category (from Expense, Income, or Debt/Loan)
- Wallet

**Optional fields:**
- Note
- Date (default: today)
- Currency (auto-convert to wallet currency)
- With (person involved)
- Location (iOS only)
- Event (must pre-create events)
- Reminder
- Photo (one only)
- Exclude from report toggle

**Multi-currency:** Enter in different currency, auto-convert to wallet currency.

### 2.2 Edit Transaction

- Tap transaction > Edit button
- Direct tap on category/amount/date fields from detail screen
- Edit any field

### 2.3 Delete Transaction

- Single: Select transaction > Delete > Confirm
- Bulk: Press and hold any transaction > Select > Choose multiple > Delete
- Deleted transactions cannot be recovered

### 2.4 Suggested Transactions

- **Amount-based:** Same amount + same category in same wallet >= 3 times → auto-suggest category
- **Note-based:** Previously used note → auto-suggest

### 2.5 View Transaction History

**Time range options:**
- Day, Week, Month, Quarter, Year, All, Custom
- Up to 20 tabs for Day/Week/Month/Quarter/Year
- Single tab for All/Custom

**Display modes:**
- View by Transaction: Most recent first
- View by Category: Categories with most recent transactions shown first

**Sort categories:** Alphabetically or by frequency of use

**Three-dot menu:**
- Time range selection
- View by Category / Transaction toggle
- Transfer between wallets
- Adjust balance
- Sync wallet (manual)

### 2.6 Search Transactions

- Search bar for finding transactions

---

## 3. Categories

### 3.1 Category Structure

- Three classification types: **Expense**, **Income**, **Debt/Loan**
- Hierarchy: Parent category → Child (sub) category
- User creates categories with icon, name, parent category
- Categories can be active/inactive per wallet

### 3.2 Default Categories (Cannot edit/delete/merge)

**Expense:** Other expense, Outgoing transfer, Pay Interest, Uncategorized
**Income:** Other Income, Incoming transfer, Collect Interest, Uncategorized
**Debt/Loan:** Loan, Repayment, Debt Collection, Debt

### 3.3 Category Management

- View all: Account tab > Categories
- Create: Name, icon, parent category (optional)
- Edit: Name, icon, parent category
- Delete: Remove permanently
- Merge: Combine two categories - all transactions from merged category transfer to target, child categories become children of target, recurring bills disappear
- Merge is irreversible
- Active/inactive per wallet toggle

### 3.4 Category Usage

- Transaction classification
- Report grouping
- Budget targets
- View transactions by category

---

## 4. Budgets

### 4.1 Budget Definition

- A limit for spending in a specific category or all categories
- Create for individual wallet or total wallets (active + included in total)
- Only for Expense and Debt/Loan categories
- Only for Basic Wallet, Linked Wallet, Credit Wallet (active, not archived)

### 4.2 Budget Types

- **Single category budget:** For one specific category
- **All Categories budget:** Overall spending limit
- **Multiple categories budget:** Combine categories across different groups (new feature)
- **Other Categories Budget:** Auto-generated for categories not budgeted (when All Categories budget exists)
- **Repeating budgets:** Auto-renew weekly/monthly/quarterly/yearly

### 4.3 Budget Fields

- Category selection
- Amount
- Time range (This week/month/quarter/year/custom)
- Wallet selection
- Repeating toggle

### 4.4 Budget Progress Bar Colors

| State | Color |
|-------|-------|
| No transactions | Gray |
| < 75% of budget | Green |
| >= 75% but not exceeded | Orange |
| Reached or exceeded | Red |

### 4.5 Budget Detail Screen

- Budget name (category name)
- Amount set
- Spent
- Left (remaining)
- Overspent
- Time period
- Wallet
- Progress bar
- Recommended daily spending
- Projected spending (estimated end-of-period)
- Actual daily spending
- Progress + projection chart
- Spent by category breakdown
- Transactions list (excludes exclude-from-report txns)

### 4.6 Budget Notifications

- Auto-notify when spending >= 75% of budget

### 4.7 Finished Budgets

- View in "Finished budgets" (from three-dot menu)
- Can only delete, cannot edit

### 4.8 Budget Overlap Rules

- Cannot create overlapping budgets (same category + time frame + wallet)

### 4.9 Total Budget Display Cases

- **Case 1:** Sum of parent categories (when parent+child budgets exist)
- **Case 2:** Sum of all categories (when child budgets exist without parent budgets)
- **Case 3:** The All Categories budget amount

---

## 5. Reports

### 5.1 Report Overview

**Access:**
- From Transaction screen > "View reports for this period"
- From Home screen > "See reports"

**Wallet selection:** Total wallet or individual wallet
**Time range:** Day, Week, Month, Quarter, Year, All, Custom

**Displayed info (Total/Basic/Linked wallets):**
- Opening balance
- Ending balance
- Net income (amount + column chart)
- Expense/Income breakdown (pie chart + amounts)
- Debt amount
- Loan amount
- Other amount (excluded transactions)
- Category breakdown
- Trends

### 5.2 Goal Wallet Report

- Goal value
- End date
- Saved amount
- Remaining amount
- Progress bar
- Income/Expense chart
- Inflow and Outflow

### 5.3 Credit Wallet Report

- Balance (negative = owe)
- Due date (notification 5 days before)
- Available Credit, Credit Limit
- Cash flow (Inflow/Outflow)
- No available credit/limit shown for previous months

### 5.4 Net Income Detail

- Total net income
- Chart of income/expenses included in report
- Detailed transaction list per time interval
- Calendar icon to change period

### 5.5 Expense/Income Breakdown

**Pie chart options:**
- Subcategories grouped under parent categories
- All categories at same level (ungrouped)
- Shows percentage and amount per category
- Daily average (expense only)

**Bar chart:**
- Trend over time
- Selectable time intervals
- Drill into transactions per interval

### 5.6 Category Report Detail

- Select category (expense or income tab)
- Subcategory reports: bar chart
- Parent category reports: bar or pie chart
- Trend info, comparison to 3-month average
- Estimated daily spending (daily average) for expenses

### 5.7 Trends

- View spending trends by quarter, year, all, or custom
- Bar chart by month
- Accessed from expense/income detail > select time range

### 5.8 Shared Wallet Member Reports

- Members section in report
- Per member: transaction count, income, expense
- Bar chart by time period
- Category-based pie charts per member
- Drill into member's transactions

### 5.9 Report This Month (Home Screen)

- Trending graph on Home screen
- Comparison: current month spending vs. average of 3 months ago
- Red/blue line: actual expense/income
- Gray line: 3-month average
- X-axis: days of month
- Three-month average = cumulative / number of months with data
- Configurable first day of month (Account > Settings)

---

## 6. Home Screen

### 6.1 Sections

1. **My Wallets:** Shows max 3 wallets included in total; can reorder
2. **Money Insider:** Reporting template card (movable to bottom)
3. **Report this month:** Trending graph (current vs 3-month average)
4. **Top spending:** Highest amount transactions
5. **Recent transactions:** Most recently added

---

## 7. Money Insider

### 7.1 Definition

- Premium add-on (subscription separate from Premium)
- Comprehensive reporting template
- Compares monthly expenses, expenses vs income and budget

### 7.2 Home Screen Card

- Most frequent spending category + total amount
- Average VND/day for that category
- Comparison to previous month's daily average (+/- %)

### 7.3 Main Screen

**Wallet and Category selection:**
- Select wallet + expense category
- Top 3 categories with most transactions
- "More" button for other categories

**Chart types:**
- **Column chart:** 6 columns - rightmost=selected period, other 5=previous periods
  - Spending within budget: blue columns
  - Spending exceeds budget: red columns
  - No budget set: all red
  - Orange line = budget line
  - Hover shows: amount spent, % of income
- **Line chart:** Spending-to-income ratio dots
  - Formula: (Monthly/Weekly Spending / Income) x 100%
  - Dot color matches column color

**Other info:**
- **Projected spending:** (Spent so far / Days elapsed) x Days in period
  - Shows "Spent" for past periods or periods with no transactions
- **Budget:** Previously set budget amount
- **Your expenses:** Period-to-date, total transactions, avg/transaction, avg/day, % vs previous period
- **Top Expenses of Month:** Top 5 highest-value transactions

### 7.4 Pricing

- 7-day free trial
- Auto-renewable subscription
- Available to Premium accounts

---

## 8. Debt & Loan

### 8.1 Categories

- Default categories: Loan, Repayment, Debt Collection, Debt
- Tracks debts (you owe) and loans (others owe you)
- Part of the Debt/Loan transaction type

### 8.2 Features

- Log debts and loans
- Track owed amounts
- Reminders for payment/collection

---

## 9. Goals / Savings

### 9.1 Goal Wallet

- Manual savings tracking
- Set savings goal amount
- Set ending date
- Record deposits
- Monitor progress
- Multiple savings wallets per goal
- No interest calculation (noted as future improvement)

### 9.2 Report

- Goal value, ending date
- Saved, remaining
- Progress bar
- Chart: income and expenses of savings wallet

---

## 10. Credit Cards

### 10.1 Credit Wallet

- Manual credit card transaction recording
- Fields: credit limit, current balance

**Formulas:**
- Available Credit = Credit Limit - Balance
- Balance = Total Inflow (all periods) - Total Outflow (all periods)
- Balance is typically negative (amount owed)

**Due date:** Notification 5 days before payment due

---

## 11. Shared Wallets

### 11.1 Definition

- Share wallet with other MoneyLover users
- Owner invites via email
- Member accepts from Account > Awaiting shared wallet

### 11.2 Permissions

- **Owner only:** Add/edit/delete categories, delete wallet, remove members
- **Members:** Can leave anytime, record transactions
- Shared wallet excluded from member's Total (cannot change)
- No shared budgets between owners and members

### 11.3 Limitations

- Budgets cannot be shared
- Cannot include in member's total wallet

---

## 12. Linked Bank Wallets

### 12.1 Definition

- Auto-sync transaction history from bank accounts
- Read-only mode (cannot use/modify account)
- RSA-2048 encryption
- Google/Amazon server storage
- Separate subscription from Premium

### 12.2 Setup

- Account tab > Connect to banks > Select bank
- Enter banking credentials
- Select account to link
- Wait 2-3 minutes for initial sync
- Manual update: Transactions tab > Update Transactions button
- Update account password: three-dot menu > Update account

### 12.3 Usage

- Auto-update every 6-8 hours
- Manual update button available
- Categorize synced transactions
- Rename wallet during setup

---

## 13. Export

### 13.1 Export to CSV

- Account > Settings > Export CSV (iOS)
- Account > Tools > Export CSV (Android)
- Options: wallet, category, time range, delimiter (comma/semicolon/tab)

### 13.2 Export to Google Sheets

- Account > Export to Google Sheet (iOS)
- Account > Tools > Export to Google Sheet (Android)
- Connect with Google account
- Options: wallet, category, time range

---

## 14. Premium Features

### 14.1 Included

- Unlimited wallets
- Unlimited budgets (with recurring budgets)
- Unlimited recurring transactions, events, bills, debts, loans
- Export to CSV/Google Sheets
- No ads
- Report for total wallet

### 14.2 Not Included

- Linked Wallet (separate subscription)
- Money Insider (separate subscription)

### 14.3 Purchase

Account > Store > Go Premium

---

## 15. AI & Auto Features

### 15.1 MoneyLover Assistant (AI Chat)

- Beta feature
- Enter transaction as natural language chat
- AI recognizes amount, category, notes
- Supports multiple transactions at once
- Supports multiple dates
- Requires default wallet (if wallet not mentioned in text)
- Transfers must be entered separately
- Access: Long-press Add button or AI icon on Add Transaction screen

### 15.2 Receipt OCR (Auto-scan)

- Beta feature
- Add Image icon on Add Transaction screen
- Choose from library or scan with camera
- AI extracts: Amount, Category, Note, Location (iOS only)
- Scan up to 10 receipts at once
- Review/edit before saving

### 15.3 Google Pay Auto-Sync

- Beta feature (v8.72.0+)
- Settings > Google Pay Tracking
- Auto-records contactless POS terminal payments
- Online transactions NOT supported
- Requires internet during payment
- Auto-select category/wallet or set defaults

### 15.4 Apple Pay Auto-Sync

- Similar to Google Pay
- Auto-tracks Apple Pay spending

---

## 16. Widget & UI

### 16.1 Widget

- Add transactions from phone home screen
- Quick transaction entry without opening app

### 16.2 Duplicate Transaction

- Copy existing transaction for quick entry

### 16.3 Settings

- Overview display mode (Opening/Ending vs Inflow/Outflow)
- First day of month adjustment
- Currency settings
- PIN/Fingerprint security

### 16.4 Tab Navigation

- Home, Transactions, Reports (or Budgets), Account
- Tab structure varies by version

---

## 17. Default Categories

### Expense Default Categories (Cannot edit/delete/merge)

| Category | Note |
|----------|------|
| Other expense | Catch-all expense |
| Outgoing transfer | Transfer to another wallet |
| Pay Interest | Interest payment |
| Uncategorized | Uncategorized transactions |

### Income Default Categories (Cannot edit/delete/merge)

| Category | Note |
|----------|------|
| Other Income | Catch-all income |
| Incoming transfer | Transfer from another wallet |
| Collect Interest | Interest received |
| Uncategorized | Uncategorized transactions |

### Debt/Loan Default Categories (Cannot edit/delete/merge)

| Category | Note |
|----------|------|
| Loan | Money borrowed |
| Repayment | Paying back debt |
| Debt Collection | Collecting owed money |
| Debt | Generic debt |

---

## 18. Category Groups & Hierarchy

### 18.1 User-Created Categories

Categories are fully user-customizable:
- Users create their own parent categories and subcategories
- Each category has: icon, name, parent (optional)
- No fixed pre-set groups enforced (except defaults above)
- Categories can be reordered in the selector

### 18.2 Typical Vietnamese Category Structure

Based on common usage patterns observed in the app:

**Expense Groups (typical):**
- Ăn uống (Food & Beverage)
- Di chuyển (Transportation)
- Mua sắm (Shopping)
- Hóa đơn & Tiện ích (Bills & Utilities)
- Giải trí (Entertainment)
- Sức khỏe (Health/Medical)
- Giáo dục (Education)
- Nhà ở (Housing)
- Bảo hiểm (Insurance)
- Du lịch (Travel)
- Quà tặng & Từ thiện (Gifts & Charity)
- Tiết kiệm (Savings)
- Đầu tư (Investment)

**Income Groups (typical):**
- Lương (Salary)
- Thưởng (Bonus)
- Lãi suất (Interest)
- Đầu tư (Investment returns)
- Quà tặng (Gifts)
- Bán hàng (Selling)
- Khác (Other)

### 18.3 Category Usage Rules

- Only Expense and Debt/Loan categories can be used for budgets
- Exclude from report: transactions not shown in reports, budgets, or Overview screen
- Categories can be active/inactive per wallet
- Merging categories transfers all transactions to target category

---

## Feature Status Summary

| Feature | MoneyLover Status | Notes |
|---------|-------------------|-------|
| Wallets (Basic) | Released | Free limited, Premium unlimited |
| Wallets (Goal) | Released | Premium |
| Wallets (Credit) | Released | Premium |
| Wallets (Linked/Bank) | Released | Subscription (separate) |
| Shared Wallets | Released | Premium |
| Transactions CRUD | Released | |
| Bulk Delete | Released | |
| Multi-currency | Released | Auto-convert |
| Search | Released | |
| Categories | Released | User-customizable hierarchy |
| Category Merge | Released | Irreversible |
| Budgets | Released | Premium for unlimited |
| Multi-category Budgets | Released | New feature |
| Recurring Budgets | Released | Premium |
| Reports | Released | Premium for total wallet |
| Report Trends | Released | |
| Money Insider | Released | Subscription add-on |
| Home Dashboard | Released | |
| Debt/Loan | Released | |
| Export CSV | Released | Premium |
| Export Google Sheets | Released | Premium |
| AI Chat Assistant | Beta | Free trial |
| Receipt OCR | Beta | |
| Google Pay Sync | Beta | v8.72.0+ |
| Apple Pay Sync | Beta | |
| Widget | Released | |
| Duplicate Transaction | Released | |
| Pin/Fingerprint | Released | |
| Sync across devices | Released | |
| Web version | Released | Recently relaunched |
| iOS 15 support | Ending | Discontinuing soon |
