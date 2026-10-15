---
artifact_type: uat_checklist
id: NATIVE-PORT-100-OWNER-UAT
status: open
owner: shared
---

# Native iOS port — owner UAT checklist

This checklist closes acceptance that simulator/unit tests cannot establish:
real-account rendering, end-user interaction, and visual comparison with the
web/Stitch reference. The iPhone 16 Simulator is currently installed with a
Debug build explicitly pointed at `https://money.dungxbuif.com/api/v1`. The
owner enters credentials; agents must not record them or copy financial values
into screenshots or logs.

## P0 — session and financial integrity

- [ ] Sign in with email/password. Confirm the authenticated home appears and
  a relaunch restores the session; explicitly log out and confirm the next
  launch returns to sign-in. If Face ID is enabled, confirm it unlocks the
  saved session rather than bypassing logout/expiry.
- [ ] Compare Overview wallet totals and recent activity with the web account.
  Toggle amount privacy and confirm values hide/show without changing totals.
- [ ] Open the transaction ledger, switch wallet scope and week, then return
  to all-wallet scope. Confirm transfers and balance adjustments are not
  counted as income/expense.
- [ ] Open a transaction and confirm its type, wallet, category, date/time,
  note, jar and report-inclusion state match the web record. Do not edit or
  delete a real transaction just to test the form.
- [ ] Check each wallet type: ordinary wallet shows its ledger; savings shows
  goal progress and scoped history; credit shows statement, entry/payment
  actions and its specialized balance. Confirm displayed balances match web.
- [ ] Verify a transfer from the transaction overflow flow (not Add
  Transaction): source/destination and both ledger legs must agree. Use a
  clearly marked disposable transfer only if the owner intends to create and
  remove test data in production.
- [ ] Compare one Reports month with the web: income, expense, net, category
  totals, trend and Money Insider. A transaction excluded from reports must
  remain outside report totals; transfer/adjustment pairs must not inflate
  income or expense.

## P1 — account configuration and recurring behavior

- [ ] Verify wallet create/edit/delete confirmation and included/excluded
  total behavior only with an explicitly disposable test wallet.
- [ ] Verify category search, kind and wallet filters; archive/restore must
  preserve historical transaction/report references and archived categories
  must not be offered for new assignments.
- [ ] Compare budget period, category/wallet scope, spent amount and progress
  against web; verify date boundaries around the account timezone.
- [ ] Compare monthly jars, assignments, unassigned spending and cumulative
  history. Confirm over-allocation is an advisory and does not change ledger
  balances.
- [ ] Open month detail, verify report-period boundaries/timezone, save a
  disposable note and reload it.
- [ ] Run a due recurring schedule only when its generated transaction is
  clearly identifiable and safe to remove; verify the schedule and generated
  ledger row agree.
- [ ] In Travel Mode, verify the active event and a transaction's linked state
  in detail; clear the link only on a disposable/owner-selected row.
- [ ] Verify portfolio positions and trade history read correctly. Do not
  create or execute trades against production data for this visual UAT.

## P2 — AI, feedback, and device presentation

- [ ] Verify Finance Assistant loads history and structured cards; do not
  treat an answer as verified unless its underlying values match the app's
  finance data.
- [ ] Verify AI transaction entry attachment thumbnails, proposal edits and
  explicit review before saving. Do not approve generated proposals into the
  real ledger during a visual-only check.
- [ ] Verify Feedback opens from the global shell, captures the current
  screen, and does not cover a row, editor action, or bottom navigation.
  **Automated status:** the approved 60pt shell dock is implemented. Focused
  iPhone 16 and compact iPhone SE simulator tests cover row/editor clearance,
  scrolling, capture geometry, tab hit-testing, and independent form insets;
  see the [validation matrix](VALIDATION_MATRIX.md). **Owner UAT remains open**
  for authenticated visual review and physical-device touch/capture behavior.
- [ ] Compare the same authenticated screens side-by-side with web/Stitch at
  the same content state: Overview, transactions, each wallet type, budgets,
  reports, categories, jars, month, recurring, travel, portfolio, Assistant,
  AI entry, feedback and account. Note any differences in layout, color,
  typography, content order or action placement.
- [ ] If the physical iPhone is available, repeat touch-target, keyboard,
  safe-area and Face ID checks there. Simulator evidence does not replace
  physical-device acceptance.

## Acceptance record

Record only pass/fail, route, build, device, and a redacted screenshot when
needed; never include passwords, API keys, balances, merchant details, or other
personal financial data. Mark overall owner UAT complete only after all
applicable P0/P1 items pass and remaining pixel differences are resolved or
explicitly accepted. The dock overlap is resolved in automated simulator
coverage, but still needs owner visual/device acceptance.

Related evidence: [port audit](PORT-100-AUDIT.md),
[validation matrix](VALIDATION_MATRIX.md), and
[feedback dock design](tickets/NATIVE-FEEDBACK-BUBBLE-DOCK-DETAIL_DESIGN.md).
