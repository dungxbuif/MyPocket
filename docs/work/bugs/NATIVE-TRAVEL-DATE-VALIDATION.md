---
artifact_type: bug
id: NATIVE-TRAVEL-DATE-VALIDATION
status: regression_verified
owner: shared
severity: normal
priority: normal
bug_markers: [reproduced]
trace:
  phase: ../phases/PHASE-002-native-ios-swiftui.md
  validation_matrix: ../VALIDATION_MATRIX.md
  release_notes: ../../releases/CHANGELOG.md
---

# Native Travel Mode accepts invalid draft dates

## Reproduction

Open Travel Mode, choose a start date after the end date, or enter only
whitespace as the trip name, then tap `Lưu`.

Expected: the form blocks the request and explains the invalid draft, matching
the web form. Actual: the native view only checks that the raw name is non-empty
and sends the date strings to the API without validating their shape or order.

## Root cause and impact

`TravelView.save()` has no date-range or trimmed-name validation. This allows a
bad draft to reach the public API, produces an avoidable server error, and can
make the native behavior differ from the web contract.

## Fix and regression proof

Added a pure `TravelEventDraftValidator` for optional `yyyy-MM-dd` values and
trimmed required names. XCTest coverage was red before the validation logic,
then green for valid, empty, malformed and reversed ranges; the view now
rejects invalid input before any API call.

## Reconciliation

Updated the native phase, validation matrix, context and changelog after green
unit/UI/build verification. No API, schema, auth or data behavior changes.
