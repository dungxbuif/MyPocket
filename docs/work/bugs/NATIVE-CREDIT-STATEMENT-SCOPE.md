# Native credit statement scope parity

## Status

`implemented` — 2026-10-06.

## Contract

Credit statement UI must make debt semantics understandable: localized credit
entry kinds, signed ledger rows, payment status and due date, current debt as a
positive absolute amount, and optional notes. Raw API enum values are wire
contract details, not user-facing labels.

## Implementation plan

Keep the existing credit entry/payment API flow, but normalize the native
presentation with shared metric/text/transaction components. Do not alter
server accounting or wallet ownership rules.

## Verification plan

- native unit tests remain green;
- Release warnings-as-errors, base/project guards, and authenticated More-route
  smoke pass on the iPhone 16 simulator.

## Evidence

- `CreditView` now presents positive current debt, localized payment status,
  due-date copy, localized credit entry kinds, and note-first row subtitles
  through shared metric/text/transaction components.
- `MyPocketIOSTests` passes 55/55 after the change.
- The warnings-as-errors simulator build passes and the authenticated More-route
  sweep passes 1/1 in 125.111s on iPhone 16 Simulator
  `EC80B092-CEF1-4A8E-A5A3-7420298EFD4D`.
- Server accounting, credit wallet ownership and API enums were not changed.
