# Feedback inbox and submission (native iOS)

Status: Task 6 reference redesign implemented locally, 2026-10-27.

The Feedback route keeps the existing API-key/JWT owner scope, screenshot
fallback behavior, changelog links and explicit owner-resolution action. The
inbox is presented as two tabs: unresolved and resolved. Ticking an item sends
the existing resolution mutation; the item then appears in the other tab. A
later un-tick is supported for regressions.

## Component contract

- `MPScreenHeader` owns route back and create actions.
- `MPSegmentedControl` owns unresolved/resolved selection.
- `MPSurfaceCard(style: .reference)` groups the visible feedback rows into one
  reference-style surface instead of a card per item.
- `MPCheckboxRow`, `MPFeedbackStatus`, `MPText` and `MPDivider` own row
  geometry, resolution semantics, changelog status and separators.
- `MPEditorScaffold` owns the full-screen submission layout, one scroll region,
  keyboard-safe footer and shared cancel/send behavior.

Backend development status (`open`, `fixed`, etc.) is intentionally distinct
from the owner's resolution checkbox. Screenshot submission still falls back
to text-only only for the documented local storage errors.

The shell-owned Feedback host remains the sole floating capture trigger. It is
anchored in the stable lower-trailing overlay above scrolling feature content,
does not reserve a layout lane, and does not reposition around route obstacles.
More retains one Feedback
destination plus its support header action. This slice does not resolve any
production feedback record or change owner-only API enforcement.

## Verification

`MoreFeedbackReferenceTests` verifies the production presentation constants for
the resolution hint, grouped surface and shell dock geometry. It intentionally
does not duplicate tab or owner policy as detached constants. Existing
APIClient feedback tests remain the authority for owner-scoped resolution,
refresh/mutation races and screenshot fallback behavior. Physical overlap and
owner acceptance remain separate release gates.
