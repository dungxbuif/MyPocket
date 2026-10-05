---
artifact_type: bug
id: NATIVE-WALLET-MONTH-LABEL
status: regression_verified
owner: shared
severity: normal
priority: high
bug_markers: [reproduced]
trace:
  phase: ../phases/PHASE-002-native-ios-swiftui.md
  validation_matrix: ../VALIDATION_MATRIX.md
  release_notes: ../../releases/CHANGELOG.md
---

# Native wallet detail renders an invalid month label

## Reproduction

1. Launch the native app and open a basic or goal wallet with at least one
   transaction month.
2. Inspect the month header in the wallet detail history.

Expected: a calendar label such as `Tháng 10 năm 2026`, matching the web
month-label contract. Actual: the SwiftUI view displays the interpolation
source text `Tháng (parts[1]) (parts[0])`.

## Root cause

`WalletDetailView.monthLabel` returns a string literal instead of interpolating
the parsed month and year components. The bug is presentation-only but affects
every non-empty wallet detail history.

## Fix and regression proof

Extracted the shared calendar label formatter into `MPFormat.monthLabel`, made
the wallet detail use it, and added an XCTest for valid and malformed months.
The regression test was red against the literal output, then green after the
interpolation and shape validation correction. The native unit suite and UI
smoke also pass.

## Reconciliation

Updated the native phase, validation matrix, context and changelog after green
verification. No API, schema, auth or data behavior changes.
