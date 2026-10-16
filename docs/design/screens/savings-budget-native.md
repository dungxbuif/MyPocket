# Native savings and recurring budgets

Status: in_progress. Scope: native Task 4 of the [implementation plan](../../superpowers/plans/2026-10-16-savings-budget-implementation.md), following the [approved API and UX design](../../superpowers/specs/2026-10-16-savings-budget-design.md).

Reports opens Savings with account-calendar month and optional goal wallet.
Goal balances/activity and account-wide surplus remain separate sections. Targets
are labeled current configuration. Scoped transaction pages use the server's
matching predicates and explicit total; a failed route shows retry, never zero.

Budgets retain explicit intervals. A recurring editor chooses weekly/monthly/
quarterly/yearly cadence and optional positive carryover. Generated instances
show base/carry/effective limits and open the series editor rather than direct
instance mutation. Series edits require the current version and take effect at
the next boundary; stop preserves existing periods. Requests include mutation
idempotency keys and the initiating owner guard.

Budget report detail shows server-calculated daily spend and forecast status:
estimated, actual, or not started. Nullable projections stay unavailable. All
money respects masking, including accessibility supplied by shared bases.

Composition: MPScreenHeader, MPEditorScaffold, MPSurfaceCard, MPText, MPMetric,
MPProgressBar, MPOptionSelector, MPCheckboxRow, MPDateField, MPAmountField,
MPButton, MPErrorView, MPEmptyView, MPStatusMessage, MPTransactionRow. Preserve
palette and compact 8pt section rhythm; no screen-local control styles.

Verification plan: decoding and presentation contracts for savings examples,
old budget payloads, null forecast, versioned mutations; signed focused XCTest,
native build and project/base guards. Populated small-phone visual, live API CRUD,
VoiceOver/Dynamic Type, device and owner UAT remain separate acceptance gates.

Parent owns public guide, API/architecture/requirements reconciliation and
context/backlog/validation/release artifacts. This Markdown is the internal and
AI-readable native screen contract; no separate native public surface is needed.

## Implementation and evidence — 2026-10-16

Implemented with the named existing bases in `SavingsReportsView.swift`,
`BudgetReportView.swift`, and the existing Reports/Budgets composition. Native
series list decodes `{items}` with `current_config` and `pending_config`;
edit uses pending configuration when present and submits the loaded version.
Creation/edit retries retain the same key for identical encoded input; stop
retains its key for the same series/version. Version errors retain the draft
and expose an explicit discard/reload action. Generated rows have report and
series actions and no individual edit/delete controls.

New report reads are live, owner-guarded and scope-bound; no savings or budget
report cache is added. Transaction drilldown uses server `limit`, `offset` and
`total`, shared rows and section-compatible kind choices. Monetary/progress
values respect the existing privacy mask. Unsupported routes fail visibly.

Decoding regression first failed on absent models, then exposed the shared
decoder's uppercase `BPS` acronym conversion; model properties now match that
existing decoder. Final signed simulator unit command:

```sh
rtk proxy xcodebuild -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS -destination 'platform=iOS Simulator,id=EC80B092-CEF1-4A8E-A5A3-7420298EFD4D' -derivedDataPath /tmp/mypocket-savings-native-20261016 -only-testing:MyPocketIOSTests test
```

Pass: 317/317, zero failures/skips. Artifact:
`/tmp/mypocket-savings-native-20261016/Logs/Test/Test-MyPocketIOS-2026.10.16_20-57-39-+0700.xcresult`.
The suite includes ten new contracts for old budgets, effective limits,
surplus separation/null ratio, goal activity/BPS, future forecast, canonical
cadence, pending configuration, scoped paging, stable versioned mutation wire,
and old-route/wrong-owner failure. Existing malformed-image fixture emits its
ImageIO diagnostic; it is not a failed test. Base-component guard, guard tests,
project verification and whitespace checks pass. New source registration was
generated from existing `project.yml` using xcodegen.

Residual acceptance: populated small-phone visuals, live local API create/edit/
stop/re-read and paging, Dynamic Type/VoiceOver and physical-device owner UAT.
No production rollout or device installation is claimed by this native slice.

Review follow-up: `initial_balance_added` decodes optionally and is displayed
when nonzero as “Số dư khi tạo ví trong kỳ”. This explains a baseline recorded
at wallet creation without showing an invented ledger transaction or placing
the balance in earlier months.
