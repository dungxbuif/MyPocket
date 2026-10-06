# Native numeric draft parsing

Status: regression_verified

## Observation

Several native financial editors currently turn malformed input such as `12abc3` or
`1_000` into a different amount by applying `filter(\.isNumber)` before decoding.
That is unsafe for financial data: the UI should either accept a valid non-negative
integer or reject the draft, never silently change the value.

## Intended contract

- Trim surrounding whitespace.
- Accept ASCII decimal digits only for integer money fields.
- Reject signs, separators, letters, exponent notation, overflow, and other
  malformed characters.
- Keep explicitly supported formatting behavior (for example recurring-entry
  display normalization) separate from API-bound validation. The recurring
  editor accepts a valid dot-grouped value such as `5.000.000`, but validates
  its grouping and rejects malformed characters instead of filtering them.
- Preserve existing field-specific constraints such as `> 0`, optional empty
  values, or maximum safe integer limits.

## Affected behavior

The shared parser is used by native transaction, transfer, adjustment,
wallet, credit, jar, budget, assistant proposal, and portfolio draft paths where
the backend expects an integer amount. Recurring-entry normalization and decimal
percentage/quantity fields remain field-specific and will not be broadened by
this change.

## Verification result — 2026-10-06

- `MPFormat.parseNonNegativeInteger` now accepts only trimmed ASCII decimal
  digits, rejects malformed characters and overflow, and supports explicit empty
  optional fields. Positive fields use the shared strict parser.
- Native transaction, transfer, adjustment, wallet, credit, jar, budget,
  assistant-proposal and portfolio price/trade paths no longer silently filter
  malformed characters before API submission. Recurring-entry display
  normalization and decimal quantity/percentage validators remain field-specific.
- Native XCTest: 47/47 passed.
- Release build with warnings-as-errors, base-component guard, project
  verification and `git diff --check`: passed.
- Signed iPhone 16 simulator API-backed UI suite: 10/10 passed in 200.110s.
