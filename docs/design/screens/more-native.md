# More / account landing (native iOS)

Status: Task 6 reference redesign implemented locally, 2026-10-27.

The More route is the owner-scoped account/navigation landing surface. It keeps
the existing `MoreDestination` identifiers, shell routing, profile action and
support/feedback entry. The presentation follows the inspected reference
account screen: centered title/support action, compact profile identity and
one rounded grouped surface per navigation section.

## Component contract

- `MPCenteredScreenHeader` owns the centered title and support affordance.
- `MPProfileCard` owns identity, avatar initial and account-route action.
- `MPSurfaceCard(style: .grouped)` owns each flat account-style destination
  group.
- `MPGroupedRow` and `MPMenuRow` own inset row geometry,
  icon/title/chevron composition and the 44pt minimum hit target.
- `MPInsetDivider` separates rows inside a group; destinations do not invent
  controls, colors, padding or separators.

`MoreReferencePresentation.sections` is the stable, testable ordering contract.
The informational bank-link policy is also consumed by the rendered row, so its
disabled-state test exercises production wiring rather than a detached value.
The shell tab remains the primary Finance Assistant entry and Overview remains
the notification entry, so More does not duplicate those rows. AI transaction
entry, account profile and Feedback remain reachable. This is presentation-only;
API/auth ownership and all typed secondary routes remain unchanged.

## Verification

`MoreFeedbackReferenceTests` verifies group ordering, grouped surface style,
the rendered bank-link enablement policy and the absence of duplicate
Assistant/notification rows. Signed simulator, compiler and guard evidence is
recorded in the validation matrix. Physical owner acceptance remains a separate
release gate.
