# Native AI proposal cards: feedback c86b669d

Status: in_review. Scope: native AI-entry review UI, its focused tests, and
the AI-entry screen spec only. Production feedback status, ordinary transaction
entry, finance rules, API, storage, and deployment are outside this work.

## Reproduction and root cause

The current `AIEntryView` loops over proposals and gives each a
`MPSurfaceCard` with a canonical `MPTransactionRow` and review actions. Its
deterministic UI fixture contains only one proposal, so no UI test proves that
multiple imports remain distinct cards. `MPTransactionRow` shows the note *or*
the date in its subtitle. A proposal with a note therefore hides the transaction
date at the point of review. This follows directly from
`MPTransactionPresentation.rowSubtitle` and the existing fixture's nonempty
note. The earlier residual design and audit establish the intended shared row,
but leave visual UAT open.

## Intended behavior and decision

Show every imported proposal as its own flat card. Keep the canonical
`MPTransactionRow` for glyph, category/title, signed or masked amount, note,
and Dynamic Type. Add proposal-specific type and account-local date metadata
within that same card, followed by the existing edit/skip/save or resolved
status controls. An unknown type remains explicit. Reuse `MPSurfaceCard`,
`MPTransactionRow`, `MPText`, and `MPButton`; add no screen styling or nested
card. The row/metadata must remain readable when a note is present. No API,
schema, config, or approval flow changes.

## Verification plan

First extend the deterministic native fixture to two distinct proposals and
add a focused simulator assertion for separate card identities, each proposal's
amount/note/date/type and own actions. Run it red against current rendering,
then make the smallest AI review change and run it green. Run native project and
base guards plus the focused native unit/UI checks. Inspect a populated
simulator screenshot; physical-device and live-provider visual acceptance remain
open. Reconcile the AI-entry screen spec and record exact evidence here.

Trace: [feedback audit](../audits/FEEDBACK-UNRESOLVED-UAT-20261020.md),
[prior residual design](../tickets/AI-ENTRY-RESIDUALS-20261020-DETAIL_DESIGN.md),
[screen spec](../../design/pages/assistant/README.md),
[native phase](../phases/PHASE-002-native-ios-swiftui.md),
[validation matrix](../VALIDATION_MATRIX.md).

## Local evidence (2026-10-26)

- RED: the two-proposal `LOCAL_API_UI` fixture and
  `testAIEntryImportShowsDistinctTransactionReviewCards` failed in
  `/tmp/mypocket-ai-cards-red-focused.xcresult`: missing card identifiers and
  type/date labels at `LaunchTests.swift:3202-3205`. The standard scheme first
  hit another worker's in-progress `FinanceAPI.updatePortfolioAsset` compile
  mismatch; a temporary UI-only scheme isolated this test without editing their
  files.
- GREEN: the new test and the existing explicit-type and pinned Save-all UI
  tests passed 3/3 on iPhone 16 Simulator (iOS 18.5) in
  `/tmp/mypocket-ai-cards-green-neighbors.xcresult`. After adding per-card
  amount assertions, the final focused run passed 1/1 in
  `/tmp/mypocket-ai-cards-final.xcresult`. The ordinary scheme's canonical row
  unit test passed 1/1 in `/tmp/mypocket-ai-cards-unit.xcresult` once the
  concurrent API edit was complete.
- `rtk proxy bash mobile-ios/scripts/check-base-components.sh` and
  `rtk proxy bash mobile-ios/scripts/verify-project.sh`: PASS.
- In `app/`, `rtk npm run check:design`, `rtk npm run test:design` (7/7),
  and `rtk npm run build`: PASS. `rtk git diff --check`: PASS.
- Screenshot review: two flat cards show distinct amounts, notes, dates and
  actions. The first capture showed floating Feedback partly covering the lower
  approval action. The [5ff74529 shell follow-up](FEEDBACK-5ff74529-native-feedback-position.md)
  now separates the trigger into its own dock in focused iPhone 16/SE simulator
  checks; physical-device visual UAT remains open.

Docs review: the internal AI-entry screen spec now matches the shipped native
presentation. No public guide, API/agent reference, schema, setup or migration
text changed because this corrects only native review presentation and adds no
new workflow or contract. The shared validation matrix/context/backlog are
being edited by other workers; this bounded note holds the focused evidence.
Production feedback status remains unchanged.
