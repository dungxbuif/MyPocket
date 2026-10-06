---
artifact_type: bug
id: NATIVE-FEEDBACK-DRAFT-SCOPE
status: regression_verified
created: 2026-10-06
updated: 2026-10-06
---

# Native feedback draft scope

## Finding

The native feedback form checked trimmed emptiness but sent the original
untrimmed title/description to the API. It also had no pure guard for an
unsupported feedback type outside the picker UI.

## Intended behavior

- Normalize title and description before JSON or multipart submission.
- Accept only `bug`, `feature`, or `improvement` as feedback types.
- Keep the existing screenshot attachment and shared form controls.

## Verification plan

1. Add a pure feedback draft normalizer and unit coverage.
2. Use it for both screenshot and JSON submission paths.
3. Run native XCTest, Release build and signed simulator route coverage, then
   reconcile the validation matrix, context, phase note and changelog.

No production data or provider credentials are used for this fix.

## Implementation result

- Added `FeedbackDraftValidator` to normalize title/description and restrict
  the type enum before either JSON or screenshot multipart submission.

## Verification result — 2026-10-06

- PASS: native XCTest is 54/54, including normalized and unsupported-type
  coverage.
- PASS: Release simulator build succeeds with
  `SWIFT_TREAT_WARNINGS_AS_ERRORS=YES`.
- PASS: signed iPhone 16 feedback-entry UI test is 1/1 in 9.821s.
- PASS: no production data or provider credential was used.
