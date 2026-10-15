# Native wallet ledger Stitch parity implementation plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans or superpowers:subagent-driven-development to implement this plan task-by-task.

**Goal:** Align native wallet detail screens with the Stitch wallet-ledger hierarchy without changing wallet-specific business rules.

**Architecture:** Keep wallet data and transaction grouping in `WalletDetailView`. Add a shared reference surface for the wallet balance/summary and stable section IDs, then compose goal/basic/credit content through the existing `MP*` bases. Credit and goal rules remain specialized; no endpoint changes.

**Tech Stack:** SwiftUI, existing MyPocket iOS design system, XCTest/XCUITest.

**Spec:** `docs/design/references/STITCH-MY-POCKET.md` and Stitch states `image.png_5`, `image.png_4`, `img_7696`–`img_7712`.

## Global Constraints

- Do not change wallet type semantics, current-balance calculations, transaction ownership, or mutation routes.
- Use `MPSurfaceCard(.reference)`, `MPWalletRow`, `MPMetricCard`, `MPSectionHeader`, `MPTransactionRow`, and `MPButton` for all new UI.
- Preserve `wallet-detail-add-transaction`, route titles, and existing mutation callbacks.
- Visual UAT on authenticated data remains a separate gate; automated tests must still prove state, not only route entry.

## Review Focus

- Goal wallets must show progress and remaining target without exposing a false completed state.
- Basic wallets must keep all transaction rows and signed amounts intact.
- Credit wallets must continue using the statement endpoint and not reuse ordinary-wallet totals.
- Empty ledgers must render a useful empty state and keep Add transaction reachable.
- Masking and account-local date formatting must remain consistent with overview and transactions.

### Task 1: Lock the wallet-ledger contract

**Files:**
- Modify: `mobile-ios/MyPocketIOSTests/DesignSystemTests.swift`

- [x] Add a failing contract for the wallet-ledger reference sections and run it red.

### Task 2: Compose the wallet reference surfaces

**Files:**
- Modify: `mobile-ios/MyPocketIOS/Features/Wallets/WalletsView.swift`
- Modify: `mobile-ios/MyPocketIOS/DesignSystem/MPSurfaces.swift` only if a reusable base is missing.

- [x] Add the explicit wallet header/balance, summary and all-transactions identifiers.
- [x] Recompose the goal/basic ledger hierarchy with reference cards while preserving current calculations and actions.

### Task 3: Verify and reconcile

**Files:**
- Modify: `docs/design/references/STITCH-MY-POCKET.md`
- Modify: `docs/work/PORT-100-AUDIT.md`
- Modify: `docs/work/VALIDATION_MATRIX.md`
- Modify: `docs/releases/CHANGELOG.md`

- [x] Run targeted and full native unit checks, build, guards, and `git diff --check`.
- [x] Record the result and leave authenticated visual UAT open until it is actually performed.
