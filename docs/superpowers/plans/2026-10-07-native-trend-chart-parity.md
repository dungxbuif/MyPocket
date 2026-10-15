# Native trend chart parity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace the native Reports “Xu hướng” placeholder bars with a real account-timezone daily income/expense trend derived from owner-scoped transactions.

**Architecture:** Keep the backend contract unchanged. A pure Swift aggregation helper converts the already available transaction list into daily points for the selected account-local month; a shared `MPTrendChart` base renders those points and an empty state. `ReportsView` loads transactions alongside its existing report endpoints and switches to the real chart only in trend mode.

**Tech Stack:** SwiftUI, Foundation `Calendar`/`DateFormatter`, existing `FinanceAPI`, shared MyPocket design-system views, XCTest.

**Spec:** `docs/design/references/STITCH-MY-POCKET.md` and the “Trend chart” row in `docs/design/INDEX.md`.

## Global Constraints

- Do not add an API, schema, migration, or provider dependency.
- Filter out transactions with `includedInReports == false` and transfer/adjustment types that are not income/expense.
- Interpret month/day in the profile timezone, never the device timezone.
- Do not fabricate data when the month has no eligible transactions.
- The chart remains a shared base component and exposes a stable accessibility identifier.

## Review Focus

- A transaction near UTC midnight must land on the account-local calendar day.
- Excluded, transfer, and adjustment rows must not affect the trend.
- Income and expense must be accumulated independently and preserve their sign semantics.
- Empty month must render a plain empty state instead of a fake chart.
- A month with a percentage or amount larger than chart height must remain readable without changing the displayed amount.

### Task 1: Pure trend aggregation contract

**Files:**
- Modify: `mobile-ios/MyPocketIOS/DesignSystem/MPFinanceCards.swift`
- Test: `mobile-ios/MyPocketIOSTests/DesignSystemTests.swift`

**Interfaces:**
- Produce `MPTrendPoint(day: Date, income: Int64, expense: Int64)`.
- Produce `MPTrendData.points(transactions:month:timeZone:) -> [MPTrendPoint]`.

- [ ] Write a failing XCTest covering timezone boundary, filtering, independent income/expense totals, empty month, and sorted output.
- [ ] Run the focused test and confirm it fails because the aggregation contract is absent.
- [ ] Implement the smallest pure helper using existing `MPFormat.parseInstant`, `accountMonth`, and a Gregorian calendar configured to the supplied timezone.
- [ ] Re-run the focused test and confirm it passes.

### Task 2: Shared trend chart base

**Files:**
- Modify: `mobile-ios/MyPocketIOS/DesignSystem/MPFinanceCards.swift`
- Test: `mobile-ios/MyPocketIOSTests/DesignSystemTests.swift`

**Interfaces:**
- Produce `MPTrendChart(points:masked:)` with accessibility identifier `reports-trend-chart`.

- [ ] Add a focused view-contract assertion for the identifier and empty-safe rendering policy.
- [ ] Run the focused test red before implementation.
- [ ] Implement a compact SwiftUI chart with aligned day labels and separate action/danger strokes/bars, clamping only drawing geometry and never the money values.
- [ ] Run the focused test green.

### Task 3: Reports integration

**Files:**
- Modify: `mobile-ios/MyPocketIOS/Features/Secondary/SecondaryFeatureViews.swift`
- Test: `mobile-ios/MyPocketIOSUITests/LaunchTests.swift`

**Interfaces:**
- `ReportsView` loads `api.transactions()` in parallel with insider/month summary and derives points from the selected month/profile timezone.

- [ ] Add an API-backed UI assertion that selecting `Xu hướng` exposes `reports-trend-chart` when eligible fixture rows exist and shows the plain empty copy otherwise.
- [ ] Run the UI test red against the placeholder chart behavior.
- [ ] Integrate transactions and replace the trend-mode `MPChart(topExpenses)` branch with `MPTrendChart`/empty copy; keep detail mode unchanged.
- [ ] Run the focused UI test green and then the full native test suite.

### Task 4: Documentation and gates

**Files:**
- Modify: `docs/design/INDEX.md`
- Modify: `docs/design/system/BASE_COMPONENTS.md`
- Modify: `docs/design/platforms/ios.md`
- Modify: `docs/work/PORT-100-AUDIT.md`
- Modify: `docs/work/VALIDATION_MATRIX.md`
- Modify: `docs/releases/CHANGELOG.md`

- [x] Mark native trend chart implemented and link the separate web parity plan; the combined docs now record both surfaces as mounted.
- [x] Run `git diff --check`, native project/base-component guards, full unit/UI test targets, and Release warnings-as-errors.

## Result

- Implemented `MPTrendData` and shared `MPTrendChart` in the native design
  system; Reports now derives the trend from owner-scoped `transactions()`.
- Focused RED/GREEN cycles were observed for the pure contract and API-backed
  Reports UI path.
- `/tmp/mypocket-trend-full-unit.xcresult`: 126/126 native unit/contract tests
  passed.
- `/tmp/mypocket-trend-ui-final.xcresult`: API-backed Reports trend smoke 1/1
  passed.
- `/tmp/mypocket-trend-release.log`: Release warnings-as-errors build passed;
  native project/base-component/diff guards passed.
- The matching web trend surface is now implemented separately in
  `2026-10-07-web-trend-chart-parity.md`; both clients use the same filtering
  contract and no web/backend API change was needed.
