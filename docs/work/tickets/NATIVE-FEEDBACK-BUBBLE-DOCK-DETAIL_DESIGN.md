---
artifact_type: detail_design
id: NATIVE-FEEDBACK-BUBBLE-DOCK
status: approved
owner: shared
trace:
  backlog: ../BACKLOG.md
  phase: ../phases/PHASE-002-native-ios-swiftui.md
  bug: ../bugs/NATIVE-LIST-FEEDBACK-BUBBLE-COLLISION.md
  shell: ../../design/platforms/ios.md
  bases: ../../design/system/BASE_COMPONENTS.md
  validation: ../VALIDATION_MATRIX.md
---

# Docked native feedback action — approved detail design

## Intent

Keep feedback globally available on authenticated native tabs while ensuring
the action never paints over list rows, form buttons, or other content. The
owner selected option A: a dedicated shell dock that reserves its own
viewport lane. This approved contract is implemented by the shared SwiftUI
shell and `MPFeedbackDock`; implementation/verification evidence is tracked in
the linked bug, validation matrix and execution plan.

## Proposed composition

Replace the shell-level floating sibling overlay with a vertically partitioned
shell layout:

```text
MainTabView
┌───────────────────────────────┐
│ selected tab / nested route   │  ← bounded content viewport
├───────────────────────────────┤
│                     [feedback]│  ← dedicated 60pt dock row
├───────────────────────────────┤
│ persistent bottom navigation │  ← existing shared navigation base
└───────────────────────────────┘
```

- The selected tab receives only the space above the dock. The feedback button
  remains trailing-aligned, globally available, and keeps its existing label,
  screenshot-capture behavior, and feedback overlay flow.
- The dock participates in the shell's measured layout; it is not a `ZStack`
  overlay, route-local inset, or extra scroll-content padding. Nested
  `NavigationStack` and `ScrollView` content must be bounded by the content
  region so visible rows and editor actions cannot pass behind the dock.
- When the feedback sheet is open or a screenshot capture is pending, retain
  the dock's geometry but hide/disable the trigger as required by the existing
  capture-redaction contract. This avoids a viewport jump during capture.
- Keep the persistent tab bar and its actions in their current order and keep
  the shell add affordance behavior unchanged.

## Shared base contract

Add `MPFeedbackDock` in `DesignSystem/MPNavigation.swift` (or the current
navigation-base file if it moves before implementation). The base owns the
trailing alignment, 60pt minimum row geometry, safe horizontal spacing, and
stable accessibility container. The shell supplies `isVisible` and the button
action; the dock does not own screenshot state, API calls, navigation state,
or feedback presentation. Keep `MPFeedbackBubble` as the visual control used
inside the dock.

Do not introduce a second color, typography, shadow, or touch-target system;
reuse the current bubble and semantic design tokens. The button remains at
least 44pt and exposes `floating-feedback-bubble` for UI automation. The
approved dock is lower-right inside the shell viewport; it is an overlay and
must not reserve a horizontal scroll lane or shift content.

## Existing spacer migration

The shell currently renders `MPBottomNavigationInset` in many feature routes,
while `MainTabView` separately places the persistent bar in a bottom
`safeAreaInset`. Before changing layout, inventory every consumer and classify
it as either:

1. a route mounted inside the shell, where shell-owned viewport bounds make a
   duplicate bar spacer unnecessary; or
2. a modal/full-screen surface outside the shell, where a local inset may
   still be necessary.

Remove or adapt only proven duplicate insets. Preserve keyboard, form-footer,
home-indicator, and full-screen-cover behavior. Add no raw per-screen spacer
heights; `MPBottomNavigationInset` remains the named base if any independent
presentation still needs it.

The narrow Account-tab hit-test blocker must be reevaluated: a non-overlapping
shell partition should make it redundant. Remove it only if a UI regression
proves the selected Account tab no longer falls through to the last More row.

## Scope and compatibility

- Native SwiftUI layout and accessibility only.
- No web behavior, backend/API, database, migration, auth, feedback payload,
  screenshot semantics, or public documentation contract changes.
- Existing `FeedbackOverlay`, submit flow, and capture-redaction behavior stay
  intact.
- The internal renderer and base-component documents, validation matrix,
  backlog, port audit, and changelog are reconciled after implementation.

## Risks and mitigations

- **Nested scroll can still draw outside its proposal:** add viewport-boundary
  UI assertions on More and category routes, before and after scrolling; inspect
  screenshots and row frames. Do not accept only `isHittable` as proof.
- **Viewport becomes shorter on compact phones:** verify scrolling and save /
  cancel reachability on iPhone 16 and a compact iPhone simulator. Avoid
  shrinking the dock below the existing button hit target.
- **Duplicate bottom spacers:** inventory shell and modal consumers and verify
  the last row/footer has no excessive blank region after the migration.
- **Capture layout shifts:** keep a fixed dock row while trigger visibility
  changes; assert the feedback attachment excludes the trigger and underlying
  screen geometry remains stable.
- **Tab-bar behavior changes:** verify Overview, Transactions, Assistant, More,
  add, and long-press AI actions; specifically prove Account-tab taps do not
  activate the final More row.

## Verification required

1. Preserve the current red proof before implementation. Add/retain frame-based
   assertions that no visible interactive More row, category row, or category
   editor save/cancel action intersects the dock or its hit target.
2. Run the regression on iPhone 16 before and after scrolling, then on a
   compact iPhone simulator. Attach screenshots for More, category list, and
   category editor.
3. Exercise tab navigation, add/long-press behavior, feedback open/capture/
   dismiss, and Account-tab hit testing.
4. Audit `MPBottomNavigationInset` consumers and test any independent
   full-screen form that retains it.
5. Run native unit/UI tests, Release warnings-as-errors, base-component and
   project guards, and `git diff --check`.

## Owner review

Approved by the owner on 2026-10-12. The 60pt dock lane, bounded content
viewport, and evidence-based spacer audit are the accepted design contract.
Implementation remains gated on review/approval of the separate implementation
plan. No backend update is required for this issue; backend compatibility
remains in scope for other reproduced native API contract gaps.
