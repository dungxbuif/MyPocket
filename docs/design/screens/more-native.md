# More / account landing (native iOS)

Status: implemented reference wave 1, 2026-10-16.

The More route is the owner-scoped account/navigation landing surface. It keeps
the existing `MoreDestination` identifiers, shell routing, profile action and
support/feedback entry. The presentation follows the inspected reference
account screen: centered title/support action, compact profile identity and
one rounded grouped surface per navigation section.

## Component contract

- `MPCenteredScreenHeader` owns the centered title and support affordance.
- `MPProfileCard` owns identity, avatar initial and account-route action.
- `MPSurfaceCard(style: .reference)` owns grouped white surfaces and their
  reference radius.
- `MPMenuRow` owns icon/title/chevron geometry and the 44pt hit target.
- `MPDivider` separates rows inside a group; destinations do not invent their
  own controls or colors.

`MoreReferencePresentation.sections` is the stable, testable ordering contract.
It is presentation-only; API/auth ownership and all existing secondary routes
remain unchanged.

## Verification

`MoreFeedbackReferenceTests` verifies group ordering, reference surface style,
and compact row rhythm. The standard native test/build guard remains required
before a release.
