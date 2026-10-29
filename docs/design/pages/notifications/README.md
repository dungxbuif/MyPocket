# Native notification inbox

Entry: the Overview bell is the single visible notification-inbox entry.
It routes through the shell's internal Account destination to
`NotificationsView`; Account/More does not show a second notification row.
The internal destination remains available for routing and deep links. Back
returns to the Account/root state using the existing callback. Owner-scoped API
loads items; reading an item persists mark-read then refreshes state. No APNs
delivery claim.

Use `MPScreenHeader` and one `MPSurfaceCard(style: .grouped)` per non-empty
read-state section. Each notification composes `MPGroupedRow`,
`MPTransactionGlyph`, `MPStatusBadge`, shared text/actions and `MPInsetDivider`;
the account-timezone date remains unchanged. Unread meaning is conveyed by the
section, badge and mark-read action, not color alone. Loading/error/retry/empty
must remain visible. Item/title/time share row alignment, with no nested cards
or local glyph padding. Native inbox and remote push are different features.
[Registry](../../platforms/ios-screen-registry.md).
