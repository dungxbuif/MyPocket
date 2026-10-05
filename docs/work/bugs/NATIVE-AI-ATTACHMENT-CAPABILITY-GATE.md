---
id: NATIVE-AI-ATTACHMENT-CAPABILITY-GATE
status: regression_verified
scope: mobile-ios
---

# Native AI-entry attachment capability gate

## Symptom

The native AI-entry screen renders an enabled document picker before checking
whether the backend reports `files_configured`. Submitting then fails with an
OCR/storage configuration error. The approved assistant contract says text
entry remains available, while attachment selection is disabled when private
OCR/storage is unavailable.

## Intended behavior

The picker is disabled while capabilities are loading, when `files_configured`
is false, and while a request is processing. Text-only analysis remains
available when AI itself is configured.

## Verification plan

1. Add a pure policy test covering loading, unavailable, available and busy
   states.
2. Wire the policy around the existing base `MPDocumentPicker` without adding
   screen-local controls.
3. Run native unit/UI suites, release build and base-component guards.

## Resolution

`AIEntryView` now wraps the base document picker with
`AIEntryAttachmentPolicy`: it is enabled only after the server reports
`files_configured=true` and while no request is processing. The text composer
remains available when files are unavailable. A pure XCTest and an API-backed
iPhone 16 UI test cover the disabled state.

Red proof: the pure policy test failed to compile before the policy existed;
the UI contract was then added against the local capability-disabled fixture.
Green proof: native unit tests pass 38/38 and the opt-in UI suite passes 9/9.
