# Native reports scope parity

## Status

`verified_fixed` — 2026-10-06.

## Contract

The web report combines the insider metrics with the month-summary ledger
breakdown: category totals/counts and an activity explanation that excludes
transfers and balance adjustments. The native report currently loads only the
insider endpoint, so it cannot show those report details.

## Implementation plan

Load `/months/:month` alongside the insider report, render the category rows
and transaction count with shared native cards/text, and keep the insider
summary as the primary error boundary. A month-summary failure should not hide
otherwise usable insider metrics.

Implemented in `ReportsView`: the insider metrics remain primary, while the
optional month-summary response now renders category/type/count rows and the
ledger activity explanation. A month-summary request failure leaves insider
metrics visible instead of replacing the screen with an error.

## Verification plan

- add model/UI data coverage through the existing month-summary decoding;
- native unit tests remain green;
- Release warnings-as-errors, base/project guards, and authenticated More-route
  smoke pass on the iPhone 16 simulator.

Verification: `MyPocketIOSTests` 55/55 passed; Release warnings-as-errors,
base-component/project guards, and `git diff --check` passed; the authenticated
More-route sweep passed 1/1 in 124.287s.
