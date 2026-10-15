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
- Profile/timezone, API-key, and security sections compose `MPSurfaceCard`.
- API-key creation/revocation uses the existing shared `MPButton` controls.
- The one-time secret uses `MPAssistantCard` + `MPText`, with a shared
  `MPButton` copy action and `MPStatusMessage` confirmation. No screen-local
  control/card styling or new UI base is introduced.

## API key behavior

- All keys have full same-account read/write authority. No scope selector.
  Before creation show the shared-base warning that a key can read/change all
  account data; existing rows say "Toàn quyền tài khoản". Keep palette/base styles.
- Create sends name and optional expiry, no scopes; old scope metadata/input is
  ignored. Existing keys do not require regeneration. Expiry/revoke/owner checks
  still apply. [Public/agent guide](../../../guides/api-keys.md).

- `GET /api/v1/api-keys` lists only metadata; it never returns a prior secret.
- Create returns a secret once. Keep it in transient view state and show the
  warning that it is one-time only.
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

The copy control is present only while the one-time secret is in memory. The
copy action and confirmation reuse controls with accessible labels and at
least a 44pt touch target. Unit tests verify the exact clipboard payload and
empty-secret no-op; API-backed UI regression verifies that tapping the screen
action presents confirmation. The test runner does not read UIPasteboard across
the iOS process boundary. Full-key follow-up:212/212unit+selectedUI pass with
inspected SE capture; no selector/warning/base fields present. Known floating
Feedback overlaps the warning, so full visual/device acceptance remains open;
do not claim universal non-overlap. Auth changed, server schema unchanged. See
[validation](../../../work/VALIDATION_MATRIX.md) and
[API-key ticket](../../../work/tickets/TICKET-10-01-quan-ly-api-key.md).
