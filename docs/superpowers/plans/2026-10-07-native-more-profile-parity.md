# Native More profile-card Stitch parity plan

## Goal

Align the native More/account landing surface with the supplied Stitch
`image.png_1` reference: centered avatar and identity block, account-status
badge, support action in the header, and a dedicated account-management row.

## Constraints

- Preserve `screen-title-Thêm`, `more-profile-help` and
  `more-profile-account` accessibility identifiers used by UI tests.
- Keep `MoreDestination` routing and the existing owner/session data unchanged.
- Use shared base components; no raw interactive screen-local controls,
  colors or radii.
- Do not add provider, billing or bank-link behavior that is not present in
  the current API contract. Status/badge copy is presentation-only.

## Implementation

1. Add a reusable `MPProfileCard` base surface for identity, status and the
   account-management action.
2. Add a shared `MPHeaderSupportButton` base for the Stitch pill support action.
3. Compose both in `MoreView`; keep the existing route callbacks and IDs.
4. Verify native unit/UI tests, Release warnings-as-errors, base/project guards
   and reconcile the visual-port docs.

Implementation status: complete. The verification evidence is recorded in
`docs/work/PORT-100-AUDIT.md` and `docs/work/VALIDATION_MATRIX.md`; physical
device and authenticated visual sign-off remain release gates.

The focused assertions additionally cover the centered fixture identity,
account subtitle and support-pill identifier in
`/tmp/mypocket-more-profile-ui2.xcresult` (2/2).

## Verification target

The profile-card route tests remain green; visual sign-off still requires
authenticated device comparison because the supplied Stitch capture contains
account data and status imagery that local fixtures cannot prove.
