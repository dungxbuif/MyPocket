# Native Finance Assistant Card Parity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Make native Finance Assistant structured cards accept the same normalized view kinds as the web assistant and render transaction details instead of falling back to raw JSON.

**Architecture:** Keep the backend response contract unchanged. Add one shared, testable `AssistantCardType` normalization/title map in the native Assistant feature, use it before rendering both overview and history parts, and add a read-only transaction-detail card using existing finance bases.

**Tech Stack:** SwiftUI, XCTest, existing `MPAssistantCard`, `MPTransactionRow`, `MPMetricCard`, `MPText`, `MPStatusMessage` and `JSONValue` model.

**Spec:** Web source `app/src/atomic/molecules/FinanceAssistantPart.tsx`; native consumer `mobile-ios/MyPocketIOS/Features/Assistant/AssistantViews.swift`.

## Global Constraints

- Do not change the AI backend response schema or tool behavior.
- Keep card rendering read-only; only the existing wallet-confirmation card may mutate after explicit confirmation.
- Preserve owner-scoped conversation/history and all existing route/accessibility IDs.
- Use shared native base components; no screen-local raw colors, font modifiers or buttons.

## Review Focus

- Alias view kinds from the web must resolve to the same native renderer.
- Unknown kinds must still show a safe fallback rather than crash or silently drop data.
- Transaction detail values may be absent; the card must render a localized fallback without force-unwrapping.
- Transfer/report inclusion flags must remain visible and consistent. Native
  masking is owned by the surrounding screen contract; this card slice does not
  invent a second masking state or persist raw values outside the response.

### Task 1: Normalize and test card kinds

**Files:**
- Modify: `mobile-ios/MyPocketIOS/Features/Assistant/AssistantViews.swift`
- Test: `mobile-ios/MyPocketIOSTests/DesignSystemTests.swift`

- [x] Add failing assertions for all five web aliases and the stable transaction-detail kind.
- [x] Run the focused unit target and confirm the new symbol/assertions fail.
- [x] Implement `AssistantCardType.normalized(_:)` and `title(for:)`, then use the normalized kind in the structured-card title/switch.
- [x] Re-run the focused target and confirm all alias assertions pass (15/15).

### Task 2: Render the missing transaction-detail card

**Files:**
- Modify: `mobile-ios/MyPocketIOS/Features/Assistant/AssistantViews.swift`
- Test: `mobile-ios/MyPocketIOSTests/DesignSystemTests.swift`

- [x] Add a deterministic detail-label helper test for wallet/category fallback, report inclusion and transfer marker.
- [x] Implement the read-only detail renderer with existing `MPText`/`MPMetricCard` bases and safe amount/report/transfer fallbacks.
- [x] Run the full native unit suite, focused Assistant UI smoke, Release warnings-as-errors build, guards and diff check.

### Task 3: Reconcile docs

**Files:**
- Modify: `docs/design/platforms/ios.md`
- Modify: `docs/design/system/BASE_COMPONENTS.md`
- Modify: `docs/work/PORT-100-AUDIT.md`
- Modify: `docs/work/VALIDATION_MATRIX.md`
- Modify: `docs/releases/CHANGELOG.md`

- [x] Record the normalized card vocabulary, evidence paths and remaining live provider/device gates.

Verification note: the focused GREEN bundle is
`/tmp/mypocket-assistant-card-green-unit/Logs/Test/Test-MyPocketIOS-2026.10.07_13-37-40-+0700.xcresult`.
The full native unit bundle is `/tmp/mypocket-assistant-card-unit` (86/86),
the Assistant UI smoke is `/tmp/mypocket-assistant-card-ui` (1/1), the Release
warnings-as-errors build is `/tmp/mypocket-assistant-card-release`, and
`verify-project.sh`, base-component checks and `git diff --check` pass.

## Follow-up: Assistant capability gate

The web composer disables input/send when the provider capability is disabled.
Native must enforce the same boundary in the shared feature policy and in the
send guard, rather than relying on the button state alone. This is a read-only
UI/validation change: it does not alter advisor routes, provider payloads or
conversation ownership.

- [x] Add the failing composer-policy contract test.
- [x] Implement the policy and apply it to native input/send controls.
- [x] Re-run unit/UI/Release/guard evidence and reconcile the docs.

Follow-up evidence: RED `/tmp/mypocket-assistant-policy-red`, focused GREEN
`/tmp/mypocket-assistant-policy-green` (1/1), full native unit
`/tmp/mypocket-assistant-policy-unit` (87/87), Assistant UI smoke
`/tmp/mypocket-assistant-policy-ui` (1/1), Release warnings-as-errors
`/tmp/mypocket-assistant-policy-release`, plus project/base/diff guards.
The disposable local API capability/overview smoke is
`/tmp/mypocket-assistant-policy-api-ui` (1/1, `LOCAL_API_UI`).

## Follow-up: goal-progress card parity

The web renderer adds a clamped progress bar and percentage for
`goal_progress`; native currently renders only the wallet balance. The native
renderer will reuse `MPProgressBar`, clamp invalid/over-target values, and
leave ordinary `wallet_balances` unchanged. No API or provider change is
needed.

- [x] Add the failing goal-progress clamp contract.
- [x] Render the progress base only for the canonical `goal_progress` kind.
- [x] Re-run the native gates and reconcile parity docs.

Goal-progress evidence: RED `/tmp/mypocket-goal-progress-red`, focused GREEN
`/tmp/mypocket-goal-progress-green` (1/1), full native unit
`/tmp/mypocket-goal-progress-unit` (88/88), Assistant UI smoke
`/tmp/mypocket-goal-progress-ui` (1/1), Release warnings-as-errors
`/tmp/mypocket-goal-progress-release`, plus project/base/diff guards.

## Follow-up: jar allocation contract parity

The web `jar_progress` card reads `allocation_amount` and intentionally omits
the progress bar when a jar has no configured allocation. Native currently
only checks budget keys (`limit_amount`/`calculated_allocation`), which can
produce a false denominator and a misleading progress bar for jar cards. The
native target resolver will prefer `allocation_amount`, then budget keys, and
will render a text-only spent value when no finite positive target exists.

- [x] Add the failing target-resolution contract tests.
- [x] Render jar progress from `allocation_amount` without changing budget behavior.
- [x] Re-run native unit/UI/Release/guard evidence and reconcile parity docs.

Jar-progress evidence: RED `/tmp/mypocket-jar-progress-red`, focused GREEN
`/tmp/mypocket-jar-progress-green` (1/1), full native unit
`/tmp/mypocket-jar-progress-unit2` (89/89), Assistant UI smoke
`/tmp/mypocket-jar-progress-ui` (1/1), Release warnings-as-errors build
`/tmp/mypocket-jar-progress-release`, plus project/base/diff guards.

## Follow-up: assistant metadata/card footer parity

The web card vocabulary exposes source provenance on metric, transaction and
wallet cards, shows total-count metadata for transaction searches, and keeps
budget/jar period totals visible. Native currently renders the core values but
omits these non-mutating context labels. The follow-up will add the same
metadata using `MPText`/`MPProgressBar` without changing finance values or API
contracts.

- [x] Add deterministic metadata-label contract tests.
- [x] Render source/count/period-total context in the shared card renderer.
- [x] Re-run native unit/UI/Release/guard evidence and reconcile parity docs.

Metadata/footer evidence: RED `/tmp/mypocket-assistant-metadata-red`, focused
GREEN `/tmp/mypocket-assistant-metadata-green` (1/1), full native unit
`/tmp/mypocket-assistant-metadata-unit-final` (90/90), Assistant UI smoke
`/tmp/mypocket-assistant-metadata-ui-final` (1/1), Release warnings-as-errors
build `/tmp/mypocket-assistant-metadata-release-final`, plus
fixture-backed API UI smoke `/tmp/mypocket-assistant-metadata-api-ui` (1/1),
and project/base/diff guards.

## Follow-up: structured-card list density parity

The web intentionally caps transaction results at five rows and wallet,
goal-progress, budget, and jar lists at eight rows. Native currently renders
every returned item, which makes the same response substantially taller and
can produce an unbounded chat card. Match the web caps while retaining the
server-provided transaction total and all existing card summaries. Keep the
API response unchanged; do not add pagination or actions in this slice.

- [x] Add a failing unit contract for normalized card-kind limits and prefix
  behavior.
- [x] Apply the caps in native transaction, wallet/goal, budget and jar views.
- [x] Verify native tests/build/base guards, then reconcile the page spec,
  audit, backlog, validation matrix and changelog.

Verification: the focused Assistant item-cap and card-metadata tests pass 2/2
on iPhone 16 Simulator at
`/tmp/mypocket-card-cap-green-derivedData/Logs/Test/Test-MyPocketIOS-2026.10.08_05-42-38-+0700.xcresult`.
Release warnings-as-errors build, project verification, both base-component
guards/tests and `git diff --check` pass.

## Follow-up: privacy-mask parity for Assistant

The web Finance Assistant receives the shell privacy-mask state and hides both
structured monetary values and free-form user/assistant text when enabled. The
native overview mask is currently local to its view model and Assistant cards
do not receive it. The shell will own the persisted mask state, pass it to
Overview and Assistant, and the native card/message renderers will redact
money and prose consistently without changing the API payload.

- [x] Add failing privacy-display and Assistant wiring contract tests.
- [x] Thread shell mask state into Overview/Assistant and redact cards/messages.
- [x] Re-run native unit/UI/Release/guard evidence and reconcile parity docs.

Privacy evidence: RED `/tmp/mypocket-assistant-privacy-red`; full native unit
`/tmp/mypocket-assistant-privacy-unit` (91/91); non-fixture Assistant UI
`/tmp/mypocket-assistant-privacy-ui` (1/1); fixture-backed Assistant API UI
`/tmp/mypocket-assistant-privacy-api-ui` (1/1); Release
`/tmp/mypocket-assistant-privacy-release`; and project/base/diff guards.

## Follow-up: secondary-screen privacy-mask routing

The web shell passes the same privacy mask into budgets, reports, jars and
month detail. Native now owns the persisted value at the shell and threads it
through `SecondaryScreen` into those four renderers. Transaction, recurring,
travel, portfolio and account routes remain unchanged because the web
reference does not pass the mask to those panels.

- [x] Add a regression contract for the secondary display formatters.
- [x] Pass the shell mask into budgets/reports/jars/months and redact all web-
  masked monetary labels and progress labels.
- [x] Re-run native unit/UI/Release/guard evidence and reconcile parity docs.

Secondary-screen evidence: focused formatter contract
`/tmp/mypocket-secondary-mask-focused` (1/1), full native unit
`/tmp/mypocket-secondary-mask-unit-final` (92/92), Assistant UI smoke
`/tmp/mypocket-secondary-mask-ui` (1/1), Release warnings-as-errors build
`/tmp/mypocket-secondary-mask-release`, and project/base/diff guards.
