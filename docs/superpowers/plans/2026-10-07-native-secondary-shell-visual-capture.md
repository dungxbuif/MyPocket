# Native secondary-route shell visual capture

## Intent

Close the shared shell regressions found while comparing authenticated native
secondary routes with the Stitch/web reference: duplicate native navigation
back controls and content finishing underneath the persistent bottom shell.

## Contract

- Secondary routes own the Stitch-style header/back capsule; SwiftUI's default
  navigation back button must not render above it.
- All scrollable screens hosted by the secondary shell (including Wallets,
  Assistant, AI Entry, Feedback and Account) reserve the shared persistent
  bottom-navigation inset instead of adding screen-local spacing.
- The shared `MPBottomNavigationInset` remains the only base spacer and keeps
  the existing tab/add/feedback shell behavior unchanged.
- The check uses the disposable local API/UI fixture only; it must not read or
  mutate production data.

## Verification

- RED evidence: `/tmp/mypocket-category-safe-area-red/Logs/Test/` exposed the
  missing inset; `/tmp/mypocket-secondary-header-red/Logs/Test/` exposed the
  duplicate native `Back` button.
- GREEN evidence: focused category route 1/1 at
  `/tmp/mypocket-secondary-header-green/Logs/Test/` and inspected screenshot
  `/tmp/mypocket-secondary-header-green-attachments/C1E759A4-27DF-4C37-AA20-D23F897175D.png`.
- The ledger capture remains independently recorded in
  `2026-10-07-native-transaction-ledger-visual-capture.md`.
- Before calling the slice complete, rerun the full native unit target,
  warnings-as-errors Release build, project/base-component/diff guards and
  reconcile the validation matrix, audit and release notes.
- PASS: final native unit target is 110/110 at
  `/tmp/mypocket-secondary-final-unit/Logs/Test/`; Release
  warnings-as-errors build is `/tmp/mypocket-secondary-final-release` and
  project/base-component/diff guards pass.
- PASS: the dedicated notification route smoke is 1/1 at
  `/tmp/mypocket-notifications-shell/Logs/Test/`; inspected attachment
  `/tmp/mypocket-notifications-shell-attachments/A02D696F-328A-4306-A51D-AC925C4E6316.png`.
  The local unauthenticated fixture returned a safe load-error card, while
  header, custom back and shell-inset assertions passed.
- RED: `/tmp/mypocket-secondary-remaining-red2/Logs/Test/` failed on the
  missing inset for `assistant`, `aiEntry`, `feedback` and `account`.
- PASS: those four screens now compose `MPBottomNavigationInset`; the focused
  route regression is 1/1 at
  `/tmp/mypocket-secondary-remaining-green/Logs/Test/`.
- Follow-up scope: verify the Wallets list/editor footer separately because it
  switches between a scrollable list and a form-owned save footer.
- RED: `/tmp/mypocket-wallet-editor-inset-red/Logs/Test/` reproduced both the
  missing inset and a save footer below the shell (maxY 802.0 vs shell top
  766.33).
- PASS: Wallets list and `WalletEditorView` now compose the shared inset; the
  focused editor contract is 1/1 at
  `/tmp/mypocket-wallet-editor-inset-green/Logs/Test/`.
- PASS: final post-wallet-fix native unit target is 110/110 at
  `/tmp/mypocket-wallet-inset-full/Logs/Test/`; Release warnings-as-errors is
  `/tmp/mypocket-wallet-inset-release`, and project/base-component/diff guards
  pass.
- Follow-up scope: verify the transfer and balance-adjustment overlays opened
  from the transaction overflow menu; both are scrollable forms rendered above
  the primary shell.
- RED: `/tmp/mypocket-transfer-overlay-red/Logs/Test/` failed the inset
  assertion for both transfer and adjustment forms.
- PASS: both overlay forms now compose `MPBottomNavigationInset`; focused
  regression is 1/1 at
  `/tmp/mypocket-transfer-overlay-green/Logs/Test/`.
- PASS: final full native unit target after overlay fixes is 110/110 at
  `/tmp/mypocket-shell-all-full/Logs/Test/`; Release warnings-as-errors is
  `/tmp/mypocket-shell-all-release`, and project/base-component/diff guards
  pass.
- PASS: final post-fix native unit target is 110/110 at
  `/tmp/mypocket-secondary-inset-full/Logs/Test/`; Release
  warnings-as-errors is `/tmp/mypocket-secondary-inset-release`, and project,
  base-component and diff guards pass.
- Follow-up scope: add a deterministic shell-inset assertion for the four
  remaining scrollable secondary screens before applying the shared base.
- Follow-up scope: verify the AI entry proposal editor overlay separately. It
  is presented above `AIEntryView` and must reserve the same persistent shell
  clearance as the parent scrollable screen.
- RED: `/tmp/mypocket-proposal-editor-red2/Logs/Test/` failed when the editor
  omitted its scoped shell inset.
- PASS: `EntryProposalEditor` now composes the shared inset and exposes a
  scoped accessibility contract; focused proof is 1/1 at
  `/tmp/mypocket-proposal-editor-green3/Logs/Test/`, with inspected screenshot
  `/tmp/mypocket-proposal-editor-green3-attachments/3E6684AD-F02A-4958-AEAB-7AEA7C9A69E4.png`.
- PASS: post-editor native unit target is 110/110 at
  `/tmp/mypocket-proposal-editor-full/Logs/Test/`; Release
  warnings-as-errors is `/tmp/mypocket-proposal-editor-release`, and project,
  base-component and diff guards pass.
- Follow-up scope: inspect the transaction-detail overlay rendered from wallet
  rows; its action surface is not a `ScrollView`, but it is still mounted above
  the persistent shell and needs an explicit scoped clearance contract.
- RED: the wallet-detail mutation flow failed the scoped
  `transaction-detail-shell-inset` assertion at
  `/tmp/mypocket-transaction-detail-red/Logs/Test/`.
- PASS: `TransactionDetailView` now composes the shared inset and focused
  API-backed proof is 1/1 at
  `/tmp/mypocket-transaction-detail-green3/Logs/Test/`, with inspected
  screenshot `/tmp/mypocket-transaction-detail-green3-attachments/1B7B79A5-7294-4BAD-97F1-732E6C1C50EF.png`.
- PASS: post-detail native unit target is 110/110 at
  `/tmp/mypocket-transaction-detail-full/Logs/Test/`; Release
  warnings-as-errors is `/tmp/mypocket-transaction-detail-release`, and
  project, base-component and diff guards pass.

## Non-goals

This slice does not claim populated production data, provider/OCR/Google
behavior, physical-device pixel approval or completion of the overall native
port.
