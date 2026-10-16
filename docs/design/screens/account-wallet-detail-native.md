# Account and wallet detail native presentation

Status: reference-density wave (2026-10-16).

Account settings and wallet detail now use the same grouped reference surfaces
as the rest of the native shell. This is a presentation-only change: API-key
creation/revocation, timezone updates, Face ID enrollment, logout, wallet
transactions, credit statements, transfer editing, queue retry and masking
retain their existing behavior and owner scope.

## Layout contract

- Account identity, timezone, API keys, one-time secret and security are
  grouped with `MPSurfaceCard(style: .reference)`.
- Wallet balance and credit statement metrics use `MPMetricGrid`; transaction
  history remains in grouped reference surfaces with the shared transaction
  row.
- Keep the fixed header, reserved bottom transaction CTA, accessibility IDs,
  confirmation dialogs and editor routes unchanged.

## Verification

`AccountWalletReferenceTests` covers account/wallet route copy and section
ordering. Focused signed tests, project/base guards, `git diff --check` and a
Release simulator build with warnings-as-errors pass.
