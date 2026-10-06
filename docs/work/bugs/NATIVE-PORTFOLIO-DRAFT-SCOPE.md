---
artifact_type: bug
id: NATIVE-PORTFOLIO-DRAFT-SCOPE
status: regression_verified
created: 2026-10-06
updated: 2026-10-06
---

# Native portfolio draft scope

## Finding

The web portfolio form trims asset symbol/name before creation and exposes an
explicit refresh action. The native form validated only non-empty raw strings,
so whitespace could be sent to the API and the current summary had no direct
refresh control.

## Intended behavior

- Trim symbol and name before validation and API submission.
- Keep the shared header/icon-button base for an explicit summary refresh.

## Verification plan

1. Add a pure draft validator test for whitespace-only and trimmed values.
2. Update the native form and rerun native XCTest, Release build and signed
   simulator route coverage.
3. Reconcile the matrix, context, phase note and changelog.

No production data or provider credentials are used for this fix.

## Implementation result

- Native portfolio asset creation now trims symbol/name before validation and
  submission, matching the web form.
- The shared screen header now exposes a base icon-button refresh action.

## Verification result — 2026-10-06

- PASS: native XCTest is 53/53, including whitespace-only and normalized draft
  coverage.
- PASS: Release simulator build succeeds with
  `SWIFT_TREAT_WARNINGS_AS_ERRORS=YES`.
- PASS: signed iPhone 16 simulator More-route sweep is 1/1 in 124.144s.
- PASS: no production data or provider credential was used.
