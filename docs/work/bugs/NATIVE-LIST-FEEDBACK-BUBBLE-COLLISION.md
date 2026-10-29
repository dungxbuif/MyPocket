---
artifact_type: bug
id: NATIVE-LIST-FEEDBACK-BUBBLE-COLLISION
status: done
severity: P2
owner: shared
trace:
  phase: ../phases/PHASE-002-native-ios-swiftui.md
  audit: ../PORT-100-AUDIT.md
  validation: ../VALIDATION_MATRIX.md
  shell: ../../design/platforms/ios.md
---

# Floating feedback bubble overlaps native list actions

## Reproduction

1. Launch the authenticated native app on iPhone 16 Simulator.
2. Open Tài khoản → Danh mục.
3. Show enough category rows that a row crosses the feedback bubble's fixed
   vertical position near the lower-right edge.

## Expected vs actual

- Expected: the feedback action stays globally accessible and list rows remain
  visible and directly tappable.
- Actual: the floating bubble covers the trailing chevron on category rows in
  the long list. The category remains reachable by scrolling, but the current
  visual and touch layout overlaps an interactive target.

## Root cause and scope

The native feedback bubble is a shell-level control positioned above the
persistent bottom navigation. `MoreView` owns a nested `NavigationStack`; its
root menu and pushed `SecondaryScreen` scroll views do not reserve the shell's
feedback row. An API-backed iPhone 16 regression reproduced both overlaps:
the `more-destination-credit` row and a category row intersect the bubble. The
category row frame was `(48, 684.3, 297, 60)` while the button frame was
`(325, 690, 44, 44)`. Screenshot evidence is attached to the category parity
run in `/tmp/mypocket-category-list-parity-green2-attachments/`.

Scope is native shell/list layout and accessibility only. Preserve the owner's
global feedback entry point. Do not change web behavior, API, backend, schema,
or capture semantics as part of the fix.

## Historical investigation and rejected layout attempts

The following notes describe the pre-fix state and are retained as decision
history. They are superseded by the 2026-10-13 Resolution section below.

The RED UI regression checks every accessibility row whose frame intersects
the app viewport; it deliberately does not use `isHittable`, because an
obscured row may report itself non-hittable and disappear from the check. It
reproduces overlap on both `more-destination-credit` and `category-row-*`.

Two layout attempts were ineffective: moving the bubble into the shell's
`safeAreaInset`, then adding a clearance inset to the `MoreView`/destination
wrappers. The iPhone screenshot and measured row/bubble frames remained
overlapped. Do not repeat those wrapper-level approaches.

The latest candidate increased the shared `MPBottomNavigationInset`, but the
focused UI test still found the More wallet row and category-editor cancel
button intersecting the bubble. That ineffective clearance increase has been
removed; the regression test remains. A direct shell/scroll architecture
review is required before another layout attempt. Do not repeat the
route-wrapper/inset approaches.

## Historical owner design review — 2026-10-08

Three layout strategies were tested and rejected: route-local clearance,
a shell safe-area feedback lane, and a larger shared bottom-navigation inset.
The focused iPhone 16 regression still detects overlap on More/category
content. The owner selected a dedicated dock row in the shell, with its own
reserved viewport lane. The detail design is ready for review; implementation
remains gated on approval of that contract. Overall native-port goal remains
active.

Evidence: `testFeedbackBubbleDoesNotCoverMoreRowsOrCategoryEditorActions`
fails in the recorded iPhone 16 results below; a later API-backed archive /
restore test also remains red on the same independent bubble-overlap
assertions. No backend/API/schema change is implicated. See the proposed
[docked-shell detail design](../tickets/NATIVE-FEEDBACK-BUBBLE-DOCK-DETAIL_DESIGN.md).

Historical current-source revalidation on 2026-10-08 still failed the focused regression
at `/tmp/mypocket-feedback-collision-revalidate-20261008.xcresult`: the More
wallets row, the More Assistant row after scrolling, and the category-editor
cancel action intersect the bubble. The recent goal-ledger root-scroll change
did not alter this shell overlap. The dock design remains in review and no
architecture change has been applied.

Historical current-source revalidation on 2026-10-12 independently reproduced the same
three failures on a disposable iPhone 16 / iOS 18.5 simulator:
`/tmp/mypocket-feedback-dock-red-20261012.xcresult`. The assertions identify
the More wallets row, More Assistant row after scrolling, and category-editor
cancel control as intersecting the bubble. Unit/contract tests pass 146/146;
Release warnings-as-errors, project verification, base-component guards and
`git diff --check` pass. This isolates the known failure to the unresolved UI
layout regression; implementation remains gated on owner review of the dock
detail design.

## Owner decision and implementation handoff — 2026-10-12

The owner approved the dock detail design, including the fixed 60pt shell lane,
bounded content viewport and evidence-based `MPBottomNavigationInset` audit.
The implementation plan is
[`2026-10-12-native-feedback-dock.md`](../../superpowers/plans/2026-10-12-native-feedback-dock.md)
and records the execution and verification evidence below.

## Resolution — 2026-10-13

`MainTabView` now clips selected route content to a bounded viewport above the
shared `MPFeedbackDock`. The dock reserves a fixed 60pt trailing lane and keeps
that geometry while the trigger is hidden during capture or Feedback
presentation. Shell-owned duplicate `MPBottomNavigationInset` consumers were
removed; the inset remains on independent modal/full-screen forms and pickers.
The Account-tab blocker remains because its selected-tab hit-test regression
passes with it in place.

Verification on disposable simulators:

- iPhone 16 / iOS 18.5 focused collision regression passes after recording the
  expected RED at `/tmp/mypocket-feedback-dock-red-20261012-2.xcresult`.
- More bottom-row reachability and category/notification bounded-viewport
  regressions pass at `/tmp/mypocket-feedback-dock-spacer-green-20261012-1.xcresult`.
- The category-editor overlap/capture check passes at
  `/tmp/mypocket-feedback-dock-category-green-20261012.xcresult`.
- Feedback-trigger hide/capture geometry, selected Account-tab hit testing,
  authenticated More routes, and editor inset regressions passed in the
  focused Task 1/2/3 results recorded in the plan ledger.

No API, backend, database, schema, migration, or public documentation contract
changed. Authenticated owner visual/device acceptance remains part of the
overall native-port UAT and is not claimed by simulator geometry tests.

## Historical shell/scroll architecture decision — 2026-10-08 (superseded)

The shared shell currently places the feedback bubble as a sibling overlay
over the selected tab's content. More/category screens own nested scroll views,
so their route-local bottom insets do not reserve the bubble's fixed vertical
lane; rows can scroll beneath the overlay. Repeated route-wrapper and inset
increases failed because they only change the end-of-scroll clearance, not the
scroll viewport where the shell overlay is painted.

Architecture hypothesis tested and rejected: compose the bubble hit-target in
a dedicated trailing-aligned row inside the shell's bottom safe-area
reservation. The iPhone 16 UI regression still found More destination rows
intersecting it both before and after scrolling. The nested
`NavigationStack`/`ScrollView` does not honor this outer reservation as a
non-overlap boundary in the current composition. The attempt was removed; do
not repeat the shell safe-area-lane approach without the full viewport
partition described in the proposed design. Preserve the global feedback
entry point and capture semantics. Compact-device proof remains open rather
than inferred.

The originally proposed layout and spacer migration are documented in
[`NATIVE-FEEDBACK-BUBBLE-DOCK-DETAIL_DESIGN.md`](../tickets/NATIVE-FEEDBACK-BUBBLE-DOCK-DETAIL_DESIGN.md).
That proposal was approved and implemented; current status and verification are
recorded in the Resolution section and [validation matrix](../VALIDATION_MATRIX.md).
