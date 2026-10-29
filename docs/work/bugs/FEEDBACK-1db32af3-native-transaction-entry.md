---
artifact_type: bug_work_note
id: FEEDBACK-1db32af3-fca8-48d9-8e10-4cef60bcb032
status: in_review
owner: ai
updated: 2026-10-26
---

# Native transaction entry: note and report inclusion

Trace: [backlog](../BACKLOG.md), [native phase](../phases/PHASE-002-native-ios-swiftui.md),
[prior feedback audit](../tickets/NATIVE-FEEDBACK-PROD-AUDIT-20261014.md),
[screen contract](../../design/pages/transactions/README.md),
[public guide](../../guides/transactions.md), [validation](../VALIDATION_MATRIX.md),
[release notes](../../releases/CHANGELOG.md).

## Reproduction and scope

Feedback ID `1db32af3-fca8-48d9-8e10-4cef60bcb032` reports that Note cannot
be tapped and asks for a report-exclusion checkbox defaulting to inclusion, with
exclusion behavior tested. This work owns only the native ordinary transaction
entry view, its UI tests and related docs. No production feedback, financial
data or deployment action is in scope.

The existing `TransactionEditorView` already presents `transaction-note-row` as
an `MPFormRow` that opens a shared note sheet. It initializes
`includedInReports` from an edit/draft or `true` for a new transaction, renders
`MPCheckboxRow`, and submits the value in `TransactionInputModel`. The focused
iPhone 16 Simulator run of `testTransactionNoteHasDedicatedEditorAndCancelDoesNotChangeDraft`
and `testAddTransactionUsesStitchFormBaseControls` passed 2/2 on 2026-10-26.
Thus the reported Note tap failure does not reproduce in this checkout; the
historical report may describe an older deployed build. Current tests verify
checkbox appearance and toggling, but not the value saved from that control.

## Decision and verification plan

Reuse `MPFormRow`, `mpFormBottomSheet`, `MPTextEditor` and `MPCheckboxRow`; no
new base, API, schema or financial rule is needed. Extend the existing
disposable `LOCAL_API_UI` wallet transaction test to turn inclusion off in the
native form, save, read back `included_in_reports=false`, reopen edit to verify
the checkbox reflects the stored value, and retain its note edit/readback proof.
If the new check fails, trace the native state-to-payload path before changing
implementation. Run the focused API UI test against a disposable local fixture,
then the surrounding native checks and reconcile this note, screen contract,
public guide and validation evidence. Physical device and deployed-build
acceptance remain separate.

Small task exemption: yes. Reason: existing native contract and controls are
already implemented; this is bounded regression proof and documentation.
Impact checked: API=no, DB=no, Security=no, Runtime=no, Standards=no.

Known unknown: which app version and device produced the original tap report.

## Result and docs review

The native editor required no production-code change. The signed iPhone 16
Simulator `LOCAL_API_UI` test passed 1/1 on 2026-10-26. It starts with the
checkbox selected, turns it off, saves an ordinary expense to a disposable
database, reads back `included_in_reports=false`, reopens edit with the box
unselected, changes the note, and reads back the same exclusion value. The
transaction remains in the ledger while the account-local monthly report
expense remains zero. The focused note-sheet and checkbox UI checks passed 2/2
before the test edit; surrounding native units passed 380/380. Project/base and
web design/build gates also passed. Exact test commands and result bundles are in the
[validation matrix](../VALIDATION_MATRIX.md).

The first API UI attempt failed before launch because fixture processes ended
with their setup shell. A second attempt reached Save but failed with the
generic save error when `CODE_SIGNING_ALLOWED=NO`; no transaction POST was sent.
Running the fixture and test in one command with simulator signing enabled
passed. The unsigned local queue failure was not separately diagnosed. The
disposable fixture was stopped and its generated database dropped after each
run. No production access occurred.

Docs review: the native screen contract and public transaction guide now say
the checkbox defaults to include and describe the saved exclusion effect. The
guide's Markdown AI-agent table carries the same rule. API, architecture,
schema, setup and release docs need no change because no behavior or contract
changed. Owner acceptance on the affected physical/deployed build remains open;
production feedback status is untouched.
