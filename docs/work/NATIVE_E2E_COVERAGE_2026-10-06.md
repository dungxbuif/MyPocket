# Native simulator E2E coverage follow-up — 2026-10-06

## Purpose

The initial iPhone 16 simulator suite proves navigation plus the highest-risk
wallet, transaction, transfer, credit, portfolio-price, AI-entry and feedback
flows. This follow-up closes the remaining stateful screen gaps before the
native port can be considered behavior-complete.

## Required state assertions

- budget: create, edit, delete and persisted summary row;
- category: create, edit and delete a user category;
- jar: create a monthly configuration, edit its allocation and remove the
  month assignment;
- month: save and delete a month note;
- recurring: create, edit, run due, then delete the schedule; generated
  transactions are removed from the disposable fixture wallet during teardown;
- travel: create, activate/deactivate, edit and delete an event;
- notifications: render the owner-scoped inbox/empty state; mark-read is
  covered by the backend repository/API contract because this phase does not
  include a notification producer or fan-out worker;
- account: update timezone, create an API key, then revoke it;
- portfolio: create an asset and verify a buy trade updates the rendered
  position (price refresh is already covered).

## Test strategy

The tests run against the disposable local API and PostgreSQL database with
`UITestAuthenticatedAPI`. Fixtures are owner-scoped and cleaned in teardown;
no production data or provider-backed credentials are used. Each test must
assert rendered UI state after the API mutation, not only that a button can be
tapped. Tests use the existing `MP*` accessibility labels and base controls.

## Verification plan

1. Add focused UI cases to `LaunchTests.swift` in small stateful groups.
2. Run each new case against the booted iPhone 16 simulator while the local
   API is healthy.
3. Rerun the complete `LOCAL_API_UI` suite, native unit/contract tests,
   backend normal/race/vet checks, base-component/project guards and
   `git diff --check`.
4. Record exact counts and any uncovered external gates in
   `docs/work/VALIDATION_MATRIX.md`.

## Confirmed simulator coverage — 2026-10-06 18:45–18:54 ICT

- PASS: the complete `LOCAL_API_UI` XCTest suite ran on the booted iPhone 16
  simulator and executed 26 tests with 0 failures in 547.331 seconds.
- PASS: follow-up targeted runs covered the previously missing stateful
  screens: budgets and categories (create/edit/delete), jars (create/edit/
  remove month configuration), month notes (save/delete), recurring schedules
  (create/edit/run-due/delete) and Travel Mode (create/activate/edit/delete).
- PASS: the budget form now reserves bottom safe-area hit space so the Save
  action cannot be obscured by the persistent tab bar; the same inset is
  applied to the long secondary forms used by reports, jars, months,
  recurring and Travel Mode.
- PASS: fixtures use unique names, owner-scoped teardown and explicit UI-state
  assertions after each mutation. No disposable wallet, recurring schedule,
  month note, budget, category, jar-month configuration or travel event was
  left by the targeted runs.

### Final simulator gate after recurring run-due coverage — 2026-10-06 18:58–19:08 ICT

- PASS: the complete `LOCAL_API_UI` XCTest suite ran on the booted iPhone 16
  Simulator (`EC80B092-CEF1-4A8E-A5A3-7420298EFD4D`), executing 26 tests with
  0 failures in 547.401 seconds. The `.xcresult` reports every test case as
  `Passed`.
- PASS: the final suite includes recurring schedule create/edit/run-due/delete,
  generated-transaction cleanup, and all previously covered wallet, ledger,
  transfer, credit, portfolio, AI-entry, feedback, budget, category, jar,
  month-note and Travel Mode flows.
- PASS: native unit/contract XCTest passed 71/71 cases with 0 failures;
  base-component and project guards plus `git diff --check` also passed.
  Backend verification then passed 231/231 tests in normal and race modes and
  `go vet` reported no issues.
- PASS: the fixture account has no wallets, budgets, recurring schedules or
  travel events remaining after teardown. Jar identities remain only as empty
  historical definitions after the tested month assignment is removed, which
  is the documented preservation behavior.
- PASS: the unsigned Release simulator artifact built cleanly, installed on
  the same iPhone 16 simulator and launched to the native Google sign-in gate;
  the release screen rendered without a crash.
- BLOCKED EXTERNAL: the paired iPhone 15 Plus is visible to `devicectl` and a
  local Apple Development identity exists, but a physical build cannot be
  signed yet. Xcode reports invalid stored account credentials and no
  provisioning profile for `com.mypocket.ios`; `-allowProvisioningUpdates`
  also reports no usable Xcode account. No source or bundle-ID workaround was
  applied.
- PASS: read-only production probes returned health/docs 200; the live
  OpenAPI lists feedback/changelog, API-key and AI routes, while invalid
  bearer probes correctly returned 401. No production write or feedback row
  was created.
