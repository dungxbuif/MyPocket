# Native Feedback

Entry: More feedback route; global bubble opens creation overlay with optional
captured screenshot; header plus opens FeedbackOverlay in shared full-screen
presentation. Submission uses existing validation/upload and preserves text
fallback when screenshot storage is unavailable.

List starts immediately below MPScreenHeader. MPSegmentedControl has unresolved/
resolved tabs with counts. Each flat MPSurfaceCard contains one MPCheckboxRow,
description, dev status/type and optional dates/changelog. No enclosing card
around the collection. Checkbox pending disables duplicate taps; server success
changes the item/tab/count, failure leaves it in place. Pull-to-refresh and
resolution writes must be serialized so a stale list cannot undo saved state.
Changelog opens from shared MPButton; empty/loading/errors use shared states.

Owner resolution is independent of dev status; unchecking retains all original
content/history. [Public behavior/API](../../../guides/feedback-ios.md),
[implementation](../../../work/tickets/FEEDBACK-OWNER-RESOLUTION.md).
Header/input alignment and screenshot-preview boundaries are part of the next
global base sweep, not claimed accepted solely by list action tests.
