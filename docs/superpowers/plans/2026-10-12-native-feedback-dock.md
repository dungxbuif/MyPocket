# Native Feedback Dock Implementation Plan

> **Execution status (2026-10-13):** Owner approved and requested implementation. Tasks 1–3 are implemented; Task 4 final compact-device/full-suite verification is in progress. Follow checkboxes and evidence below.

**Goal:** Keep Feedback globally available without covering interactive content by reserving a dedicated, measured dock lane in the authenticated native shell.

**Architecture:** Replace the shell-level floating sibling overlay with a vertically partitioned shell: bounded selected-tab content, a 60pt trailing-aligned feedback dock row, then the existing persistent tab navigation. Add `MPFeedbackDock` as the shared base; migrate only those `MPBottomNavigationInset` uses proven redundant inside shell-hosted routes, retaining independent modal/full-screen clearance.

**Tech Stack:** SwiftUI, XCTest/XCUITest, Xcode Simulator, project/base-component shell scripts.

**Spec:** `docs/work/tickets/NATIVE-FEEDBACK-BUBBLE-DOCK-DETAIL_DESIGN.md`

## Global Constraints

- The dock row has 60pt minimum geometry; the button has at least a 44pt hit target.
- Keep the current feedback label, `floating-feedback-bubble` accessibility identifier, screenshot capture behavior, and feedback presentation flow.
- Keep dock geometry while feedback is presented or screenshot capture is pending.
- Keep persistent tab order and the shell add/long-press behavior unchanged.
- Reuse the current bubble and semantic tokens; add no new color, typography, shadow, or hit-target system.
- Do not change web behavior, backend/API, database, migrations, authentication, feedback payload, screenshot semantics, or public API documentation.
- Do not remove the narrow Account-tab hit-test blocker unless a regression proves it redundant.
- Do not add raw per-screen spacer heights; retain `MPBottomNavigationInset` where a modal/full-screen surface still needs it.

## Files and responsibilities

- `mobile-ios/MyPocketIOS/DesignSystem/MPNavigation.swift`: shared `MPFeedbackDock` base; existing `MPFeedbackBubble` remains the visual control.
- `mobile-ios/MyPocketIOS/Features/Shell/MainTabView.swift`: shell viewport partition, dock visibility/action wiring, existing feedback capture state and tab navigation.
- `mobile-ios/MyPocketIOS/Features/More/MoreView.swift`, `Features/Overview/OverviewView.swift`, `Features/Wallets/WalletsView.swift`, `Features/Transactions/TransactionsView.swift`, `Features/Assistant/AssistantViews.swift`, `Features/Secondary/SecondaryFeatureViews.swift`, `Features/Feedback/FeedbackView.swift`, and `DesignSystem/MPFields.swift`: classify every `MPBottomNavigationInset` consumer as shell-hosted or independent modal/full-screen; remove/adapt only proven duplicate shell clearance.
- `mobile-ios/MyPocketIOSUITests/LaunchTests.swift`: frame-based dock, scroll, tab hit-test, capture, and route regressions.
- `docs/design/system/BASE_COMPONENTS.md`, `docs/design/platforms/ios.md`, `docs/work/bugs/NATIVE-LIST-FEEDBACK-BUBBLE-COLLISION.md`, `docs/work/VALIDATION_MATRIX.md`, `docs/work/PORT-100-AUDIT.md`, `docs/work/BACKLOG.md`, and `docs/releases/CHANGELOG.md`: reconcile the shared-base and shell contracts, verification evidence, task status, and unreleased change after implementation.

## Review Focus

1. **Nested `NavigationStack`/scroll boundary:** visible More/category row frames must never intersect the dock before or after scrolling. Pin in `testFeedbackBubbleDoesNotCoverMoreRowsOrCategoryEditorActions`.
2. **Compact viewport and pinned form controls:** category save/cancel and the final scroll row remain reachable above the shell navigation on a compact iPhone. Pin in the same regression plus a compact-device run.
3. **Screenshot redaction and layout stability:** hide the trigger before capture while keeping the dock lane fixed, then present/dismiss Feedback without shifting the underlying route. Pin in `testFeedbackBubbleHidesBeforeFeedbackOverlayCapture` and an added frame-stability assertion.
4. **Account tab hit-testing:** tapping the selected Account tab must not activate the final More row; retain or remove the narrow blocker only according to the test result. Pin in an added selected-account-tab regression.
5. **Independent modal/full-screen inset behavior:** forms and sheets that are not shell viewport children retain their own necessary safe-area clearance and keyboard/footer behavior. Pin in existing transaction detail, transfer/adjustment, AI proposal editor, and Feedback route tests; add a focused test only where no stable proof exists.

---

### Task 1: Strengthen frame-based shell regression

**Files:**
- Modify: `mobile-ios/MyPocketIOSUITests/LaunchTests.swift`
- Test: `mobile-ios/MyPocketIOSUITests/LaunchTests.swift`

**Interfaces:**
- Consumes: existing `floating-feedback-bubble`, `more-destination-*`, `Lưu nhóm`, `Hủy`, `openMore`, `reveal`, and screenshot helpers.
- Produces: stable frame-based proof for the 60pt dock row, no intersection with visible interactive route controls, fixed geometry during trigger hide/capture, and Account-tab hit testing.

- [x] **Step 1: Extend the existing failing test** `testFeedbackBubbleDoesNotCoverMoreRowsOrCategoryEditorActions` to assert the dock's stable accessibility identifier/frame, no intersection with each visible More row before and after scrolling, and no intersection with category save/cancel controls.
- [x] **Step 2: Add `testSelectedAccountTabDoesNotActivateLastMoreRow`**: after opening More, tap the selected Account tab and assert the current route remains the More root rather than activating its last row.
- [x] **Step 3: Extend capture coverage** in `testFeedbackBubbleHidesBeforeFeedbackOverlayCapture` to assert the dock frame is unchanged when the trigger hides and Feedback opens; retain the existing assertion that the trigger is not left over the Feedback surface.
- [x] **Step 4: Run the focused test on an isolated iPhone 16 simulator**; expected RED was observed at `/tmp/mypocket-feedback-dock-red-20261012-2.xcresult`, then GREEN after implementation. Preserve the `.xcresult` path in the validation matrix.

### Task 2: Add the shared dock and partition the shell viewport

**Files:**
- Modify: `mobile-ios/MyPocketIOS/DesignSystem/MPNavigation.swift`
- Modify: `mobile-ios/MyPocketIOS/Features/Shell/MainTabView.swift`
- Test: `mobile-ios/MyPocketIOSUITests/LaunchTests.swift`

**Interfaces:**
- Consumes: `MPFeedbackBubble(action:)`, `MPBottomNavigation`, existing `feedbackPresented`, `feedbackCapturePending`, `feedbackScreenshot`, and the selected-tab content switch.
- Produces: `MPFeedbackDock(isVisible:action:)`, a 60pt-minimum trailing shell lane containing the existing 44pt-or-larger bubble; dock visibility does not own capture, feedback, navigation, or API state.

- [x] **Step 1: Implement `MPFeedbackDock`** in `MPNavigation.swift`; expose a stable `feedback-dock` container identifier and preserve `floating-feedback-bubble` on its child control. Keep the row's height fixed when `isVisible` is false.
- [x] **Step 2: Replace the root feedback `ZStack` sibling overlay** in `MainTabView` with a measured vertical composition. Give selected-tab content the flexible bounded region, place the dock row below it, and leave the existing bottom `safeAreaInset` navigation ordering/actions intact.
- [x] **Step 3: Wire existing capture state through the dock**: `isVisible` is false while `feedbackPresented` or `feedbackCapturePending`; do not collapse the reserved row in either state. Keep the 100ms pre-capture hide and `FeedbackOverlay` transition behavior unchanged.
- [x] **Step 4: Run the focused RED test** from Task 1 and `testFeedbackBubbleHidesBeforeFeedbackOverlayCapture`; expected GREEN is no frame intersection, stable dock geometry through capture, and no duplicate/missing feedback trigger.

### Task 3: Audit and migrate shell-owned bottom insets

**Files:**
- Modify only proven duplicate consumers in: `Features/More/MoreView.swift`, `Features/Overview/OverviewView.swift`, `Features/Wallets/WalletsView.swift`, `Features/Transactions/TransactionsView.swift`, `Features/Assistant/AssistantViews.swift`, `Features/Secondary/SecondaryFeatureViews.swift`, `Features/Feedback/FeedbackView.swift`, and `DesignSystem/MPFields.swift`.
- Test: `mobile-ios/MyPocketIOSUITests/LaunchTests.swift`

**Interfaces:**
- Consumes: Task 2's bounded shell viewport and the existing `MPBottomNavigationInset` base.
- Produces: each inset call site classified as shell-hosted or independent modal/full-screen, with only verified duplicate shell insets removed/adapted and independent safe-area/keyboard/footer clearance retained.

- [x] **Step 1: Record every `MPBottomNavigationInset` call site** with its presentation owner and classify it as shell-hosted or independent sheet/cover/modal. `MPFields.swift` picker and transaction/AI editor overlays remain independent.
- [x] **Step 2: Remove only shell-hosted duplicate clearance** after verifying the route's scroll content is bounded by `MainTabView`; keep the shared inset base for independent presentations. No literal replacement spacers were added.
- [x] **Step 3: Run targeted route/form tests**. More routes, transfer/adjustment forms and shell-clipped category/notification routes pass. Wallet-editor, AI-proposal-editor, and API-backed Overview tests are fixture-gated and were not counted as passing; full API fixture acceptance remains separate.
- [x] **Step 4: Re-run the strengthened Feedback collision regression**; focused GREEN evidence covers More root, scrolled More and category editor.

### Task 4: Verify compact layout, navigation, and documentation

**Files:**
- Modify: `docs/design/system/BASE_COMPONENTS.md`
- Modify: `docs/design/platforms/ios.md`
- Modify: `docs/work/bugs/NATIVE-LIST-FEEDBACK-BUBBLE-COLLISION.md`
- Modify: `docs/work/VALIDATION_MATRIX.md`
- Modify: `docs/work/PORT-100-AUDIT.md`
- Modify: `docs/work/BACKLOG.md`
- Modify: `docs/releases/CHANGELOG.md`
- Test: `mobile-ios/MyPocketIOSUITests/LaunchTests.swift`

**Interfaces:**
- Consumes: implemented shared base, shell layout, and migrated inset classification.
- Produces: verified default/compact-device behavior and reconciled internal release/port evidence; no public API document changes because no API contract changes.

- [x] **Step 1: Run collision and capture regressions on iPhone 16 and iPhone SE (3rd generation)** using disposable simulator IDs. Default-device focused runs pass; compact collision passes in `/tmp/mypocket-feedback-dock-se-green-20261013.xcresult`, and compact capture/More reachability/Account hit-testing pass in `/tmp/mypocket-feedback-dock-se-20261013.xcresult`. Do not use the owner's visible device.
- [x] **Step 2: Run shell interaction regressions** `testAuthenticatedMoreRoutesAndFeedbackEntry`, `testGlobalAddLongPressOpensAIEntry`, `testSelectedAccountTabDoesNotActivateLastMoreRow`, and `testFeedbackBubbleHidesBeforeFeedbackOverlayCapture`. These pass in focused iPhone 16/SE results; add long-press is 1/1 at `/tmp/mypocket-feedback-dock-add-action-20261013.xcresult`.
- [x] **Step 3: Run native unit suite, Release warnings-as-errors build, project/base guards, and `git diff --check`**. PASS: 146/146 unit tests; Release build with Swift warnings treated as errors; `verify-project.sh`, `check-base-components.sh`, `check-base-components.test.sh`, and `git diff --check` all pass.
- [x] **Step 4: Reconcile documentation** with simulator IDs, test names, `.xcresult` paths, inset decisions, and limitations; obtain the required independent scoped review. No actionable findings. Bug status is resolved by default/compact-device regressions; authenticated owner visual UAT remains distinct/open.

## Plan self-review

- Spec coverage: global availability, fixed geometry, base ownership, visibility during capture, inset inventory/migration, Account hit testing, compact-phone behavior, tab/action preservation, screenshots, tests, and docs each map to a task above.
- Scope: native SwiftUI/UI test/internal docs only; API/backend/database/auth/business behavior are explicitly excluded.
- Review-focus tests are assigned to Tasks 1, 3, and 4; independent modal clearance is explicitly preserved and tested.
- The approved detail design is implemented in this worktree. Do not claim deployment, commit, production validation, or completion of the broader native-port goal from this scoped fix.
