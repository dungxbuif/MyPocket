# Account and wallet detail native presentation

Status: Task 6 account-settings redesign implemented and verified (2026-10-27).

Account settings and wallet detail now use the same grouped reference surfaces
as the rest of the native shell. This is a presentation-only change: API-key
creation/revocation, timezone updates, Face ID enrollment, logout, wallet
transactions, credit statements, transfer editing, queue retry and masking
retain their existing behavior and owner scope.

## Layout contract

- Account identity, timezone, API keys and security are grouped with
  `MPSurfaceCard(style: .reference)` and shared section spacing.
- The one-time API-key secret is a flat subsection of the API-key surface,
  separated by `MPDivider`. `MPSensitiveValue` owns masked/revealed text and
  its VoiceOver-safe toggle; leaving the active scene remasks the value before
  an app-switcher snapshot. Copy remains exact, local-only and five-minute
  pasteboard-expiring. Revocation clears the transient secret.
- Wallet balance and credit statement metrics use `MPMetricGrid`; transaction
  history remains in grouped reference surfaces with the shared transaction
  row.
- Keep the fixed header, reserved bottom transaction CTA, accessibility IDs,
  confirmation dialogs and editor routes unchanged.

## Verification

`MoreFeedbackReferenceTests` covers the account reference anatomy and shared
surface contract. Existing API-key clipboard/API tests and `DeviceSignInTests`
remain the behavior authority for exact-secret copy, owner-only key operations,
Face ID enrollment/revocation and login. Focused signed tests, project/base
guards, `git diff --check` and a Release simulator build with warnings-as-errors
are required. No API, auth policy or wallet behavior changes in this slice.
