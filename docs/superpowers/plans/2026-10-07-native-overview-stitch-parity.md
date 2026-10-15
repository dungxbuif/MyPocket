# Native overview Stitch parity implementation plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans or superpowers:subagent-driven-development to implement this plan task-by-task.

**Goal:** Bring the native overview hierarchy closer to the supplied Stitch Money Lover reference while preserving the existing owner-scoped overview/month/insider contracts.

**Architecture:** Keep `OverviewViewModel` as the source of real data. Extract only reusable visual surfaces into existing `MP*` base components, then compose the overview in the same order as the reference: balance header, wallet summary, insider, period report, recent activity, and month detail. No endpoint or payload changes.

**Tech Stack:** SwiftUI, existing MyPocket iOS design system, XCTest/XCUITest.

**Spec:** `docs/design/references/STITCH-MY-POCKET.md`, especially the canonical tokens and overview screen mapping.

## Global Constraints

- Keep API behavior, timezone calculations, masking, refresh and navigation callbacks unchanged.
- Reuse `MPSurfaceCard`, `MPSectionHeader`, `MPSegmentedControl`, `MPWalletRow`, `MPTransactionRow`, and `MPButton`; do not add screen-local colors or radii.
- Preserve accessibility identifiers used by existing UI/API tests.
- Visual parity is measured against the 738×1600 Stitch screenshots; this slice does not claim pixel-level completion without authenticated visual UAT.

## Review Focus

- Empty/error overview states must remain actionable and must not render misleading balances.
- Masked balance mode must mask every amount in the reordered cards.
- Week/month report switching must continue to use the real month and current-week totals.
- Wallet list and recent transaction actions must still route through their existing callbacks.
- A failed insider/month endpoint must not prevent the wallet summary from rendering.

### Task 1: Lock the reference hierarchy

**Files:**
- Modify: `mobile-ios/MyPocketIOSUITests/LaunchTests.swift`
- Modify: `docs/work/PIXEL-PARITY-ACCOUNT-TRANSACTION.md`

- [x] Add a unit contract that names the overview reference sections and fixes wallet summary before report/month detail.
- [x] Run the targeted test and observe the expected compile failure before implementation.

### Task 2: Recompose the overview with shared bases

**Files:**
- Modify: `mobile-ios/MyPocketIOS/Features/Overview/OverviewView.swift`
- Modify: `mobile-ios/MyPocketIOS/DesignSystem/MPFinanceCards.swift` only if a reusable reference surface is missing.

- [x] Move the wallet summary before the insight/report surfaces and give each section a stable accessibility identifier.
- [x] Align the wallet action copy and section spacing to the Stitch hierarchy without changing model calculations or callbacks.
- [x] Keep month detail and recent activity available below the primary reference surfaces.

### Task 3: Verify and reconcile

**Files:**
- Modify: `docs/work/VALIDATION_MATRIX.md`
- Modify: `docs/work/PORT-100-AUDIT.md`
- Modify: `docs/releases/CHANGELOG.md`

- [x] Run the full native unit/UI checks, project/base-component guards, and `git diff --check`.
- [x] Record the actual result and keep authenticated device/pixel UAT explicitly open.

### Task 4: Match the Stitch balance-header grouping

**Files:**
- Modify: `mobile-ios/MyPocketIOS/DesignSystem/MPNavigation.swift`
- Modify: `mobile-ios/MyPocketIOSUITests/LaunchTests.swift` only if the stable
  header identifiers need a regression assertion.

- [x] Keep the balance amount and visibility control in the same leading row,
  with search and notification actions in the trailing cluster, matching the
  supplied Stitch overview capture.
- [x] Preserve the existing accessibility identifiers and notification route.
- [x] Re-run native unit/UI smoke, Release build, project/base guards and
  reconcile the visual-port audit.
