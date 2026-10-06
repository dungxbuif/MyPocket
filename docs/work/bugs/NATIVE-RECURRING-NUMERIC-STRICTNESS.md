---
artifact_type: bug
id: NATIVE-RECURRING-NUMERIC-STRICTNESS
status: regression_verified
owner: shared
---

# Native recurring numeric strictness

## Reproduction

`RecurringDraftValidator.positiveInteger` currently removes every character
that is not recognized as a number before parsing. A malformed draft such as
`1a2` therefore becomes `12` and can be submitted as a different recurring
amount or interval.

## Impact

Recurring schedules are financial writes. Silently changing a user's draft is
unsafe and violates the shared numeric-input contract used by the other native
editors. Valid integer input must be preserved; malformed input must be
rejected before a create/update request is made.

## Intended behavior

- Accept a positive integer after surrounding whitespace is trimmed.
- Preserve the existing recurring display-format compatibility for a valid
  thousands-grouped value such as `5.000.000`.
- Reject letters, signs, decimal/exponent notation, misplaced separators,
  overflow and any other malformed characters; never filter them out.

## Verification result — 2026-10-06

`RecurringDraftValidator` now validates optional thousands grouping instead of
filtering arbitrary characters. `5.000.000` remains accepted, while `1a2`,
`1 2` and `5.00.000` are rejected before the recurring create/update request.
The targeted regression is green; native XCTest passes 62/62, and the signed
iPhone 16 `LOCAL_API_UI` suite passes 77/77 (0 failures). Release
warnings-as-errors, base-component/project guards, no invalid-frame faults and
fixture cleanup also pass.
