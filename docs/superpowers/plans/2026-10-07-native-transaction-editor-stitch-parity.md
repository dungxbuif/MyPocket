# Native Transaction Editor Stitch Parity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Bring the native transaction editor's high-frequency fields and bottom action bar in line with the Stitch reference while preserving existing API behavior and accessibility routes.

**Architecture:** Add shared transaction-form primitives to the iOS design system: a row selector with icon/title/value/chevron and a date stepper with calendar affordance, previous/next controls, and a localized date pill. The editor composes those primitives inside the existing reference surface and keeps all selection, validation, loading, and save logic in `TransactionEditorView`.

**Tech Stack:** SwiftUI, XCTest/XCUITest, existing MyPocketIOS design tokens and FinanceAPI.

**Spec:** `docs/design/references/STITCH-MY-POCKET.md`, Stitch reference `image.png_2/screen.png`.

## Global Constraints

- Do not change API endpoints, request models, validation policy, or owner scoping.
- All new controls must be base components under `mobile-ios/MyPocketIOS/DesignSystem`.
- Preserve existing identifiers: `transaction-type-selector`, `transaction-details-toggle`, and `Lưu giao dịch`.
- Keep the form keyboard/safe-area behavior and disabled/empty states accessible.
- Use existing `MPColor`, `MPFont`, `MPMetrics`, and `MPButton` tokens; no feature-local raw styling.

## Review Focus

- Empty category or jar option: the row must remain visibly disabled/empty and must not send an invalid ID.
- Long wallet/category names: values must truncate without moving the chevron outside the hit target.
- Date boundaries: previous/next controls must preserve the account timezone and never mutate the API timestamp format.
- Loading/save failure: the footer must retain the shared loading state and error message without losing entered values.
- Small iPhone viewport and keyboard: the fixed footer must stay reachable and the scroll content must not be hidden behind it.

### Task 1: Add shared transaction-form base controls

**Files:**
- Modify: `mobile-ios/MyPocketIOS/DesignSystem/MPFields.swift`
- Modify: `mobile-ios/MyPocketIOS/DesignSystem/MPNavigation.swift`
- Test: `mobile-ios/MyPocketIOSTests/DesignSystemTests.swift`

**Interfaces:**
- Produce `MPFormRow(title:iconName:value:isEnabled:accessibilityIdentifier:action:)` for semantic selector rows.
- Produce `MPDateStepper(date:timeZone:onPrevious:onNext:accessibilityIdentifier:)` for account-local date navigation.

- [x] **Step 1: Write failing unit assertions** for the new base contracts: default tokens use `MPFont.body`, the row exposes a 44pt hit target, and the date stepper formats a localized day label from the supplied timezone.
- [x] **Step 2: Run `xcodebuild test ... -only-testing:MyPocketIOSTests/DesignSystemTests`** and confirm the new symbols/assertions fail because the primitives do not exist (`/tmp/mypocket-transaction-red`, missing-symbol compile failure).
- [x] **Step 3: Implement the two primitives** with existing tokens, icon fallback, truncation, disabled styling, stable accessibility IDs, and deterministic `Calendar`/`DateFormatter` timezone handling.
- [x] **Step 4: Re-run the focused unit target** and confirm it passes (13/13 at `/tmp/mypocket-transaction-green-unit`).

### Task 2: Compose the Stitch transaction editor layout

**Files:**
- Modify: `mobile-ios/MyPocketIOS/Features/Transactions/TransactionsView.swift`
- Test: `mobile-ios/MyPocketIOSUITests/LaunchTests.swift`

**Interfaces:**
- Consume `MPFormRow` for wallet/category/note rows and `MPDateStepper` for the date row.
- Preserve `MPOptionSelector` menu semantics behind the new row actions.

- [x] **Step 1: Extend `testAddTransactionUsesStitchFormBaseControls`** with identifiers for `transaction-wallet-row`, `transaction-category-row`, `transaction-note-row`, `transaction-date-stepper`, and `transaction-save-footer`; the first run failed only at the missing date-stepper button identifier.
- [x] **Step 2: Replace the editor's ad-hoc selector/date stack** with the shared rows, keeping the current bindings, jar visibility rules, date state, and validation unchanged.
- [x] **Step 3: Keep “Thêm chi tiết” and note/report controls behind the existing toggle**, and place the primary save button in a safe-area-aware footer with the existing `isSaving` behavior.
- [x] **Step 4: Run the focused UI test** and confirm the new controls are discoverable and the existing save/details assertions remain green (`/tmp/mypocket-transaction-ui2`, 1/1).

### Task 3: Verify and reconcile documentation

**Files:**
- Modify: `docs/work/PORT-100-AUDIT.md`
- Modify: `docs/work/VALIDATION_MATRIX.md`
- Modify: `docs/design/system/BASE_COMPONENTS.md`
- Modify: `docs/design/platforms/ios.md`
- Modify: `docs/releases/CHANGELOG.md`

- [x] **Step 1: Run the complete native unit suite, focused transaction UI test, Release warnings-as-errors build, project/base guards, and `git diff --check`.**
- [x] **Step 2: Record exact evidence paths and the remaining live/provider/device gates** in the validation matrix.
- [x] **Step 3: Document `MPFormRow` and `MPDateStepper` ownership and update the transaction screen mapping/changelog.**
