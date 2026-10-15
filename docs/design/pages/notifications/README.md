# Native notification inbox

Entry: Overview bell currently routes through shell to NotificationsView;
More duplicate entry is queued for removal. Back returns to caller's More/root
state using the existing callback. Owner-scoped API loads items; reading an item
persists mark-read then refreshes state. No APNs delivery claim.

Use MPScreenHeader, shared surface/text/action bases, local account-timezone date
format, unread semantic indicator not color alone. Loading/error/retry/empty
must remain visible. Item/title/time should share row alignment, with no nested
cards or extra glyph-sized padding. Native inbox and remote push are different
features. [Registry](../../platforms/ios-screen-registry.md).
