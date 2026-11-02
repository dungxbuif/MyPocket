---
artifact_type: detail_design
id: TRANSACTION-DRAFT-SAFETY-20261102
status: in_review
owner: shared
---

# Transaction draft and save interaction

Owner approved the product priority in conversation: improve ordinary transaction
entry/edit first, preserving shared bases, palette and typography. This bounded
slice protects an entered draft and makes save failures visible next to Save.

Trace: [backlog](../BACKLOG.md), [native phase](../phases/PHASE-002-native-ios-swiftui.md),
[screen](../../design/pages/transactions/README.md), [bases](../../design/system/BASE_COMPONENTS.md),
[validation](../VALIDATION_MATRIX.md), [changelog](../../releases/CHANGELOG.md).

## Evidence and approach

Source inspection: MPEditorScaffold disables header/fields during requests but
not the footer. TransactionEditorView's AI footer action can therefore leave a
pending save. Cancel immediately calls the parent even after amount/note changes.
Save errors are at the end of the scrolling fields and may be offscreen.

- Disable the shared scaffold footer while busy, covering secondary actions in
  all its consumers (including wallet delete). Existing loading labels remain.
- Extend shared confirmation copy with a configurable cancel label, retaining
  the default for all existing consumers.
- Snapshot the ordinary editor's initial field values after initial category
  normalization, before asynchronous jar loading. Compare all editable values,
  including amount, type, wallet, category, jar, note, report flag and date.
- Cancel or switching to AI with changed values asks whether to discard; staying
  keeps all fields. Untouched forms close immediately. Confirmed exits invoke
  only their intended destination. Block interactive dismissal for dirty or busy drafts.
- Show validation/save errors beside the pinned Save action; clear the previous
  error when another valid submission starts. Pending copy says “Đang lưu…”.

Durable draft storage and automatic retry were considered but excluded: they
need separate lifecycle/reconciliation design. Existing encrypted outbox and
financial semantics remain authoritative. This slice does not change API, schema,
auth, runtime, colors, fonts or dependencies; no new ADR is required.

## Acceptance and verification

Untouched cancel exits; changed amount/note/date prompts; “Tiếp tục nhập” keeps
the form; “Bỏ thay đổi” exits; switching to AI follows the same protection.
While busy all scaffold actions are disabled. Invalid save shows its error above
the footer without clearing the draft. Save success retains its existing callback.

Run native compile and static checks. Automated test execution remains skipped
under the owner's earlier manual-testing preference; add focused UI scenarios
for later integrated execution. Device/visual UAT remains open. Docs review must
update screen/base contracts, context, backlog, validation and changelog.

## Implementation and reconciliation — 2026-11-02

The current diff implements the bounded design in `MPNavigation.swift` and
`TransactionsView.swift`; `LaunchTests.swift` adds
`testTransactionExitKeepsDraftUntilDiscardConfirmed` and
`testTransactionInvalidSaveShowsFooterErrorAndPreservesDraft`. The first covers
untouched Cancel, changed-note Cancel/AI stay and confirmed Cancel discard; the
second covers the visible pinned validation error and retained draft. Neither
scenario has been executed, by owner preference.

Implementation-turn evidence: build-for-testing used
`/tmp/mypocket-draft-safety-20261102`. Its `Logs/Build/LogStoreManifest.plist`
records build `7BEDA300-AC75-46F8-A418-D161062C069F` with `highLevelStatus=S`,
zero errors and zero warnings. `verify-project.sh` and `git diff --check` PASS.
`check-base-components.sh` FAILS at the existing, unchanged
`MainTabView.swift:235` `fullScreenCover`. This is compilation/static evidence,
not a passing UI test run. Exact full build invocation was not supplied to this
docs reconciliation; the manifest is the retained build evidence.

Runtime/device UAT remains OPEN: review dirty-field changes, retained values,
confirmed AI destination, swipe prevention, busy secondary actions and
failed/successful saves. Status is `in_review`, not verified or done.

Docs review: screen and shared-base contracts now match the source diff;
[context](../../CONTEXT.md), backlog, validation and unreleased changelog record
the bounded candidate and residual gates. No API, schema, authentication,
financial-rule, runtime or dependency change requires master-contract or ADR
updates. Broader transaction/product work remains open. Rich-data seeding is
handled independently and is not evidence for this slice.
