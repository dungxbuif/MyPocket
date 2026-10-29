# Native overview wallet subtitle parity implementation plan

## Goal

Match the web Overview wallet-card subtitle rule: show the wallet's saved
description when present, otherwise show the semantic wallet-type label.

## Contract

- Extend the shared `MPWalletRow` base with an optional subtitle override.
- Overview passes `description ?? wallet-type label`; other consumers keep the
  existing type-label default.
- Preserve balance masking, icon/tone, accessibility and navigation behavior.

## Verification

- [x] Add a pure unit contract for subtitle fallback selection.
- [x] Run native unit, Release warnings-as-errors, project/base-component
  guards and diff checks.
- [x] Reconcile validation/audit/iOS docs and changelog; no API/schema/migration
  change was made.

## Shipped evidence

- `MPWalletRow` now accepts an optional subtitle; Overview passes the saved
  description with a shared type-label fallback, while other consumers retain
  the original subtitle behavior.
- RED compile evidence is `/tmp/mypocket-overview-wallet-subtitle-red`.
- Focused regression is 1/1 at
  `/tmp/mypocket-overview-wallet-subtitle-unit/Logs/Test/`.
- Full native unit/contract suite is 97/97 at
  `/tmp/mypocket-overview-wallet-subtitle-full/Logs/Test/`.
- Release warnings-as-errors build is
  `/tmp/mypocket-overview-wallet-subtitle-release`; project/base-component
  and diff guards pass.
