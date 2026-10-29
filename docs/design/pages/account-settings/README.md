---
artifact_type: screen_design_contract
id: ACCOUNT-SETTINGS
status: implemented
owner: shared
---

# Account settings

Route: More → Tài khoản. This is an authenticated, owner-scoped settings
surface; API keys are managed with JWT/session or a full-owner API key. This
new auth policy is implemented locally, not yet activated in production.

## Composition and bases

- Fixed `MPScreenHeader` and scrollable account settings content.
- Profile/timezone, API-key, and security sections compose
  `MPSurfaceCard(style: .reference)` with `MPMetrics.sectionSpacing`.
- API-key creation/revocation uses the existing shared `MPButton` controls.
- The one-time secret is a flat subsection of the API-key surface, separated by
  `MPDivider`. `MPSensitiveValue` owns masking, VoiceOver-safe reveal/hide and
  Dynamic Type layout; the shared `MPButton` copy action and `MPStatusMessage`
  own the remaining interaction feedback. There is no nested card, local eye
  toggle or screen-local control/card styling.

## API key behavior

- All keys have full same-account read/write authority. No scope selector.
  Before creation show the shared-base warning that a key can read/change all
  account data; existing rows say "Toàn quyền tài khoản". Keep palette/base styles.
- Create sends name and optional expiry, no scopes; old scope metadata/input is
  ignored. Existing keys do not require regeneration. Expiry/revoke/owner checks
  still apply. [Public/agent guide](../../../guides/api-keys.md).

- `GET /api/v1/api-keys` lists only metadata; it never returns a prior secret.
- Create returns a secret once. Keep it masked in transient view state and show
  the warning that it is one-time only. Revealing changes presentation only and
  never persists the secret. The revealed value disables ordinary iOS text
  selection, keeping clipboard handling on the bounded copy action. When the
  scene becomes inactive or enters the background, reset reveal state; returning
  to the app keeps the same secret masked.
- Tapping “Sao chép API key” copies the exact currently visible secret to the
  system pasteboard. It is local-only and expires after five minutes; the app
  does not log or persist it. Announce “Đã sao chép API key.” to VoiceOver and
  show the same success status in the card.
- Leave/reload the route and the one-time secret is no longer available. A new
  key can be created; an existing key cannot be revealed again.
- Revoke remains an explicit destructive action and removes the key from the
  visible list after the API succeeds.
- Reopening filters revoked metadata out of the native active-key list; raw API
  history remains. Revoking the just-created key clears its transient secret/copy
  confirmation. Actual API UI verifies create/copy/revoke/leave/reopen against
  rebuilt local backend, separately from fake-session warning layout coverage.

## States and verification

The API-backed native UI regression reveals the one-time secret, confirms a
long press does not expose the ordinary iOS text-selection menu, sends the app
through Home/background, reactivates it and requires the same value to be
masked. This verifies the interaction on the signed iPhone 16 simulator; visual
inspection of a physical app-switcher snapshot remains an owner acceptance gate.

The copy control is present only while the one-time secret is in memory. The
copy action and confirmation reuse controls with accessible labels and at
least a 44pt touch target. Unit tests verify the exact clipboard payload and
empty-secret no-op; API-backed UI regression verifies that tapping the screen
action presents confirmation. The test runner does not read UIPasteboard across
the iOS process boundary. The shell now reserves a dedicated Feedback dock, but
physical-device layout and VoiceOver acceptance remain open. This remediation
does not change API, auth, persistence or server schema. See
[validation](../../../work/VALIDATION_MATRIX.md) and
[API-key ticket](../../../work/tickets/TICKET-10-01-quan-ly-api-key.md).
