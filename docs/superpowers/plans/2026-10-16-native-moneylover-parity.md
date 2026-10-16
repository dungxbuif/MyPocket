# Native Money Lover parity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rebuild the native iOS presentation around the approved Money Lover reference while preserving every MyPocket API, route, auth, owner-scope, and finance behavior.

**Architecture:** Keep `DesignSystem` as the sole visual primitive layer. Migrate the shell first, then Account and feature surfaces in independent waves. Each wave adds deterministic reference assertions and a device screenshot checkpoint before the next wave.

**Tech Stack:** SwiftUI, UIKit bridges for text fields, XCTest, Xcode Release builds, `devicectl` for iPhone 15 Plus validation.

**Spec:** `docs/superpowers/specs/2026-10-16-native-moneylover-parity-design.md`

## Global Constraints

- Money Lover is a visual reference only; no Money Lover runtime code, credentials, or data is copied.
- Existing MyPocket business rules, API contracts, owner isolation, auth guards, routes, offline behavior, and AI flows remain unchanged.
- All interactive/layout primitives live in `mobile-ios/MyPocketIOS/DesignSystem`; screens cannot add local colors, fonts, radii, or raw interactive controls.
- Personal-Team device builds use `MyPocketIOS/Signing/MyPocketIOS.personal.entitlements`; production entitlements remain unchanged.
- Every wave must pass `git diff --check`, unit tests, warnings-as-errors Release build, and a live-device screenshot checkpoint.

## Review Focus

- Dynamic Type and Vietnamese diacritics: system typography must preserve hierarchy without clipping. Test in the typography/base-component task.
- Small iPhone width and safe areas: grouped cards and bottom navigation must not overlap or reserve a feedback lane. Test in shell task.
- Long profile/menu labels: rows must wrap or expand predictably without changing hit targets. Test in Account task.
- Route/auth preservation: visual migration must not bypass pending-route or owner guards. Test in shell/route task.
- Masked financial values: visual changes must not reveal balances or transaction amounts. Test in feature-wave task.

### Task 1: Lock the typography, color, metrics, and icon contracts

**Files:**
- Modify: `mobile-ios/MyPocketIOS/DesignSystem/Tokens.swift`
- Modify: `mobile-ios/MyPocketIOS/DesignSystem/MPIcon.swift`
- Modify: `mobile-ios/MyPocketIOS/DesignSystem/MPText.swift`
- Test: `mobile-ios/MyPocketIOSTests/ReferenceDisplayTests.swift`
- Test: `mobile-ios/MyPocketIOSTests/MoreFeedbackReferenceTests.swift`
- Update: `docs/design/references/STITCH-MY-POCKET.md`

**Interfaces:**
- Consumes: existing `MPFont`, `MPColor`, `MPMetrics`, `MPText`, and icon call sites.
- Produces: one system-font hierarchy, semantic color roles, measured card/row metrics, and a complete reference icon mapping consumed by every later task.

- [x] Step 1: Add failing assertions for the approved system font family, title/body/label sizes, page inset, card radius, row minimum hit target, and required reference icon names.
- [x] Step 2: Run `xcodebuild test` for the focused reference tests and verify the new assertions fail against the old values.
- [x] Step 3: Implement the token/icon changes only in the shared design-system files; remove remaining screen-level font-name use.
- [x] Step 4: Run the focused tests and verify all typography/metric/icon assertions pass.
- [x] Step 5: Reconcile the Stitch/reference docs with the actual native typography decision and commit `feat(native): lock parity design tokens`.

### Task 2: Rebuild the native shell and bottom navigation

**Files:**
- Modify: `mobile-ios/MyPocketIOS/DesignSystem/MPNavigation.swift`
- Modify: `mobile-ios/MyPocketIOS/Features/Shell/MainTabView.swift`
- Modify: `mobile-ios/MyPocketIOS/App/AppRootView.swift`
- Test: `mobile-ios/MyPocketIOSTests/ReferenceDisplayTests.swift`
- Test: `mobile-ios/MyPocketIOSUITests/MyPocketIOSUITests.swift`

**Interfaces:**
- Consumes: Task 1 tokens/icons; existing `AppRouteCoordinator` and `MainTabView` selection bindings.
- Produces: Money Lover-like safe-area shell, centered add action, selected-tab capsule, feedback bubble dock, and route-safe tab transitions without changing route values.

- [x] Step 1: Add failing layout/route assertions for five destinations, centered add hit target, selected tab state, hidden tab bar while keyboard is visible, and no feedback lane reservation.
- [x] Step 2: Run focused XCTest/UI smoke and confirm failures identify the current shell mismatch.
- [x] Step 3: Implement shell changes through `MPBottomNavigation`, shared header, and feedback dock; leave feature content untouched.
- [x] Step 4: Run focused tests plus the route sweep; verify deep links, auth pending-route consumption, back behavior, and tab selection.
- [x] Step 5: Build Release for Simulator with warnings-as-errors and commit `feat(native): align parity shell`.

### Task 3: Port Account screen to the reference anatomy

**Files:**
- Modify: `mobile-ios/MyPocketIOS/Features/More/MoreView.swift`
- Modify/Create: `mobile-ios/MyPocketIOS/DesignSystem/MPFinanceCards.swift` (profile/premium card base only)
- Modify: `mobile-ios/MyPocketIOS/DesignSystem/MPSurfaces.swift`
- Test: `mobile-ios/MyPocketIOSTests/MoreFeedbackReferenceTests.swift`
- Test: `mobile-ios/MyPocketIOSUITests/MyPocketIOSUITests.swift`

**Interfaces:**
- Consumes: Task 1/2 bases and existing `MoreDestination`, `AccountView`, feedback routing, and auth session data.
- Produces: Account tab with reference profile/premium card, grouped wallet/group rows, a clearly informational/disabled bank-link row (no unsupported fake write), events/recurring/bills group, and preserved MyPocket-only destinations behind the same routes.

- [x] Step 1: Add failing assertions for section order, profile-card data source, premium badge visibility, bank-link row style, grouped separators, feedback bubble position, and logout/auth accessibility identifiers.
- [x] Step 2: Run focused tests and confirm the current Account/More implementation fails the reference anatomy assertions.
- [x] Step 3: Implement the screen using only shared `MPProfileCard`, `MPSurfaceCard`, `MPMenuRow`, `MPButton`, and `MPIcon`; map supported actions to existing MyPocket routes and render unsupported Money Lover actions as informational/disabled without fake writes.
- [x] Step 4: Run focused XCTest/UI smoke with long names, masked state, logout, and route navigation.
- [x] Step 5: Capture iPhone 15 Plus screenshot, compare against `/tmp/moneylover-account.png`, record remaining deltas, and commit `feat(native): align account reference screen`.

### Task 4: Align overview, transactions, wallets, and add flow

**Files:**
- Modify: `mobile-ios/MyPocketIOS/Features/Overview/OverviewView.swift`
- Modify: `mobile-ios/MyPocketIOS/Features/Transactions/TransactionsView.swift`
- Modify: `mobile-ios/MyPocketIOS/Features/Wallets/WalletsView.swift`
- Modify: `mobile-ios/MyPocketIOS/DesignSystem/MPSurfaces.swift`
- Test: `mobile-ios/MyPocketIOSTests/OverviewReferenceTests.swift`
- Test: `mobile-ios/MyPocketIOSTests/TransactionsReferenceTests.swift`
- Test: `mobile-ios/MyPocketIOSTests/WalletsReferenceTests.swift`

**Interfaces:**
- Consumes: shell/navigation and token contracts from Tasks 1–3; existing finance API view models.
- Produces: reference balance header, wallet cards, scoped ledger, period controls, transaction rows, transfer overflow, and global add flow with unchanged API calls.

- [x] Step 1: Add/extend failing assertions for card geometry, row density, selected wallet scope, period tabs, add action, transfer action, and masked values.
- [x] Step 2: Run focused tests and verify failures.
- [x] Step 3: Refactor feature views to shared surfaces/rows and apply measured reference spacing.
- [x] Step 4: Run unit/UI smoke for ordinary, goal, credit, transfer, masked, empty, and error states.
- [x] Step 5: Capture device screenshots and commit `feat(native): align ledger and wallet reference screens`.

### Task 5: Align planning/reporting and specialized surfaces

**Files:**
- Modify: `mobile-ios/MyPocketIOS/Features/Secondary/SecondaryFeatureViews.swift`
- Modify: `mobile-ios/MyPocketIOS/Features/Reports/SavingsReportsView.swift`
- Modify: `mobile-ios/MyPocketIOS/Features/Budgets/BudgetReportView.swift`
- Modify: `mobile-ios/MyPocketIOS/DesignSystem/MPFinanceCards.swift`
- Test: `mobile-ios/MyPocketIOSTests/JarsMonthsReferenceTests.swift`
- Test: `mobile-ios/MyPocketIOSTests/CreditNotificationsReferenceTests.swift`
- Test: `mobile-ios/MyPocketIOSTests/BillsDebtTravelReferenceTests.swift`

**Interfaces:**
- Consumes: shared surfaces and navigation from Tasks 1–4; existing report/budget/jar/credit models and API contracts.
- Produces: consistent cards, charts, selectors, empty/error states, and specialized wallet detail surfaces.

- [x] Step 1: Add failing reference assertions for card hierarchy, chart labels, selector states, and action confirmations.
- [x] Step 2: Run focused tests and verify failures.
- [x] Step 3: Replace local visual variants with shared bases and preserve all write confirmations and owner scoping.
- [x] Step 4: Run the complete native unit suite and Release build.
- [x] Step 5: Capture representative device screens and commit `feat(native): align planning and reports surfaces`.

### Task 6: Align settings, auth, AI, feedback, and route states

**Files:**
- Modify: `mobile-ios/MyPocketIOS/Features/Auth/AuthView.swift`
- Modify: `mobile-ios/MyPocketIOS/Features/Auth/BiometricUnlockView.swift`
- Modify: `mobile-ios/MyPocketIOS/Features/Assistant/AssistantViews.swift`
- Modify: `mobile-ios/MyPocketIOS/Features/Feedback/FeedbackView.swift`
- Modify: `mobile-ios/MyPocketIOS/Features/More/MoreView.swift`
- Test: `mobile-ios/MyPocketIOSTests/MoreFeedbackReferenceTests.swift`
- Test: `mobile-ios/MyPocketIOSUITests/MyPocketIOSUITests.swift`

**Interfaces:**
- Consumes: all shared tokens/shell/surfaces and existing auth/AI/feedback route contracts.
- Produces: reference-aligned auth/Face ID, assistant composer/cards, feedback tabs, settings selectors, and deep-link guarded secondary routes.

- [x] Step 1: Add failing assertions for auth state, Face ID lock screen, feedback resolved/unresolved tabs, AI action cards, and route guard behavior.
- [x] Step 2: Run focused tests and verify failures.
- [x] Step 3: Implement visual-only changes through base components; do not broaden permissions or provider calls.
- [x] Step 4: Run full unit suite, UI smoke, route sweep, and API fixture isolation checks.
- [x] Step 5: Capture final device screenshots and commit `feat(native): align auth assistant feedback surfaces`.

### Task 7: Full parity gate and device release

**Files:**
- Update: `docs/design/reference-survey/2026-10-16-money-lover-iphone.md`
- Update: `docs/work/VALIDATION_MATRIX.md`
- Update: `docs/CONTEXT.md`
- Update: `docs/releases/CHANGELOG.md`

**Interfaces:**
- Consumes: all completed visual waves and their screenshot evidence.
- Produces: reconciled parity checklist, release notes, and an installable Personal-Team device build.

- [ ] Step 1: Run `rtk git diff --check` and a base-component audit (`rg` for screen-local font/color/radius/interactive declarations).
- [ ] Step 2: Run the complete native XCTest suite and record pass/fail counts.
- [ ] Step 3: Build Release for Simulator and iPhone with warnings-as-errors; verify the iPhone artifact is signed with `AU442S3MN3` and the personal entitlements fallback.
- [ ] Step 4: Install and launch with `devicectl`, capture every release-gate screen, and reconcile deltas against the reference checklist.
- [ ] Step 5: Update docs/changelog with exact remaining gaps; only then commit `release: native parity checkpoint`.
