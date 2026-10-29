---
artifact_type: bug
id: NATIVE-ENUM-LABEL-PARITY
status: implemented
created: 2026-10-06
updated: 2026-10-06
---

# Native enum-label parity

## Finding

Several native screens still rendered backend enum values directly (for
example `basic`, `expense`, and feedback statuses). The API wire values must
remain unchanged, but user-facing iOS labels should use the shared Vietnamese
label vocabulary, matching the web and the other native screens.

## Intended behavior

- Keep raw enum values only in Codable/request logic.
- Centralize display labels in the design-system label component.
- Apply the shared labels to wallet rows, category rows, credit status/details,
  feedback history, and Assistant terminal-status errors.

## Verification plan

- Add pure label tests for wallet/category/feedback/credit values.
- Run native XCTest, Release warnings-as-errors, base/project guards and the
  signed `LOCAL_API_UI` simulator suite.

## Implementation result

- `MPSelectionLabels` now owns wallet, category, credit, feedback and common
  terminal-status display labels without changing API wire values.
- Wallet rows, category rows, credit detail, feedback history and Assistant
  errors consume the shared label vocabulary.

## Verification result — 2026-10-06

- Native XCTest: 57/57 passed, including shared-label coverage.
- Release simulator build with Swift warnings treated as errors: passed.
- Signed iPhone 16 simulator targeted UI checks: credit-wallet route 1/1 and
  feedback/more route 1/1 passed after the label change.
- Base/project guards and `git diff --check`: passed.
