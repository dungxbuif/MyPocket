# Feedback inbox and submission (native iOS)

Status: implemented reference wave 1, 2026-10-16.

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

## Verification

`MoreFeedbackReferenceTests` verifies tab IDs/copy, the resolution hint and
the single grouped reference surface. Existing APIClient feedback tests remain
the authority for refresh/mutation races and screenshot fallback behavior.
