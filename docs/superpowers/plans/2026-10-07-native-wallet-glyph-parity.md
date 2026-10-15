# Native wallet glyph parity

## Intent

Match the shared web `WalletCard` contract in native Overview and wallet rows:
the same semantic wallet type maps to a 48pt circular badge, stable SF Symbol
glyph and tone (`basic` warning, `goal` success, `credit` teal).

## Contract

- Keep wallet API values, owner scope, labels, balances and navigation unchanged.
- Put icon geometry and semantic colors in the design system; feature views may
  only compose `MPWalletGlyph`.
- Preserve row accessibility by letting the existing wallet row combine the
  glyph, name, subtitle and balance.

## Verification

- Add/adjust design-system token assertions for the new semantic tones.
- Run full native unit tests, Release warnings-as-errors and base/project/diff
  guards.
- Re-run web typecheck/build/design checks because the web contract is the
  source of truth, although no web code is changed.

## Status

- [x] Shared glyph and semantic tokens implemented
- [x] Native/web verification passed
- [x] Documentation reconciled

## Evidence

- Focused semantic-token/icon contract passed 1/1 at
  `/tmp/mypocket-wallet-glyph-contract-final/Logs/Test/Test-MyPocketIOS-2026.10.07_16-08-46-+0700.xcresult`.
- Full native unit/contract target passed 110/110 at
  `/tmp/mypocket-wallet-glyph-final-full/Logs/Test/Test-MyPocketIOS-2026.10.07_16-09-21-+0700.xcresult`.
- Release warnings-as-errors build passed at `/tmp/mypocket-wallet-glyph-final-release`;
  project/base-component/diff guards and web typecheck/build/design checks
  also passed.
