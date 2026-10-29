# Ngân sách — dữ liệu API

Native override (2026-10-14): Tài khoản → Ngân sách. Add/edit replaces the list
with `MPEditorScaffold`, shared header, single scroll and pinned Save. Category
remains a shared sheet; pending disables edits/cancel, failure keeps the draft.
Expired budgets remain a read-only preview with delete confirmation. Summary
uses flat `MPMetric` inside one surface. Web descriptions below are historical;
native palette/base contracts are authoritative. Hủy does not save; reopening
the same budget in this mounted screen retains input. Server metadata is refreshed,
so an ended budget is still read-only. Leaving the screen/restarting is not a
durable-draft guarantee. Same-item retention and ended-preview pass in final
gate188/188, including decoded ledger equality after delete cancellation.
Ended screenshot inspected; device UAT remains.

Canonical route `/account/budgets`; `/budgets` remains a compatibility route. [Design](../../../work/tickets/API-SCREENS-01-DETAIL_DESIGN.md). Uses the shared ScreenHeader plus SurfaceCard, MetricBox, BudgetGauge/`MPBudgetGauge`, BudgetProgressItem, Text, BaseButton, StatusMessage and StatusBadge/`MPStatusBadge`. The summary gauge is driven by real spent/limit values and clamps only its visual arc; displayed money remains the unmodified API value. Editor uses BaseBottomSheet, BudgetEditorForm, AmountField, DateField and CategoryTreeSelector backed by the same BaseCategoryTree selection mode as group management and transaction forms. Native iOS uses the searchable shared `MPCategoryPicker` selection tree for category scope; “Tất cả danh mục” preserves the existing nil-category API contract. No screen-local visual variants.

List reads GET budgets with derived spent and deduplicated summary. Loading is not zero; failures show retry, never examples. Empty state is plain “Chưa có ngân sách.” Add opens name/limit/wallet/category/start/end-date form. Explicit intervals only in this slice; end date inclusive in UI, exclusive timestamp in API. Optional wallet/category mean all. Choose expense groups only. Group choice opens searchable `CategoryTreeSelector` backed by `BaseCategoryTree`, matching the group-management root/child presentation; “Tất cả nhóm chi” clears the selection. Save waits for API; draft survives errors. Active rows edit/delete via API; expired rows expose a read-only `Xem / Xóa` preview and destructive delete confirmation, never editable fields. Mask conceals all monetary values. Source transaction save/delete refreshes this screen. Delete budget does not change ledger. Wallet/category delete removes associated tracking configuration.

Native recurring periods, deterministic forecasts and report drilldowns are
implemented. Server notifications remain unsupported, and browser/owner visual
acceptance is still pending; no fake controls or data are shown for either gap.

Metadata edits retain the original timestamp boundaries when corresponding visible date fields are unchanged, even after switching browser timezone. Intentional date edits use the current browser-local calendar. Owner account timezone persistence remains separate. API persistence proof and real empty/editor browser inspection recorded in the linked design; full browser CRUD/owner sign-off pending.

## Native budget-report hierarchy — Task 5 reference redesign

The budget report uses the same shared finance hierarchy as Savings Reports.
`MPFinanceSummaryCard` owns the title/context, primary outcome, chart slot,
supporting metrics and footer order. Its chart slot mounts the existing
`MPBudgetGauge`; the screen supplies only the real `spent` and
`effective_limit_amount` values returned by the report API.

| Region | Shared base | Existing values/state |
| --- | --- | --- |
| Budget outcome | `MPFinanceSummaryCard`, `MPBudgetGauge`, `MPMetricGrid` | Remaining amount as the primary outcome; spent/effective-limit gauge; base limit, carryover, effective limit and spent metrics |
| Forecast | `MPFinanceCard`, `MPMetricGrid`, `MPStatusBadge` | Existing actual/estimated/not-started status, projected values and over-budget state |
| Transactions | `MPButton`, `ReportTransactionsView` | Existing budget drilldown endpoint and paging |
| Daily activity | `MPFinanceCard`, `MPMetricRow` | Existing daily values and empty state |

Gauge geometry clamps visually while displayed money remains the exact API
value. A zero or invalid denominator produces a zero visual percentage rather
than a non-finite value; it does not rewrite `spent`, `remaining`, carryover or
forecast arithmetic. Privacy masking applies to the primary outcome, gauge
VoiceOver label and all metrics. Loading, retry-with-last-snapshot, empty daily,
estimated and over-limit states retain their behavior. No API request, payload,
schema or date-boundary behavior changes.

Verification plan: focused native tests cover safe percentage calculation,
over-limit values and privacy-safe summary presentation;
existing forecast decode tests remain the arithmetic contract. Run the focused
iOS test target, base-component guard and warnings-as-errors build. Small-phone,
landscape, Dynamic Type and dark-mode visual acceptance remain an integrated
Task 7/8 UAT gate when no dedicated Task 5 screenshot harness is available.

Implementation evidence (2026-10-26): the isolated
`JarsMonthsReferenceTests` run passed 7/7 and the complete
`MyPocketIOSTests` target passed 408/408 on the iPhone 16 iOS 18.5 simulator.
The Release simulator build passed with `SWIFT_TREAT_WARNINGS_AS_ERRORS=YES`;
`mobile-ios/scripts/check-base-components.sh`,
`mobile-ios/scripts/verify-project.sh` and `git diff --check` also passed.
No dedicated budget-report screenshot fixture exists in this task scope, so
small-phone/landscape/Dynamic Type/dark-mode visual acceptance remains an
explicit integrated UAT item rather than an inferred pass.

## Parent budget reference composition — 2026-10-26

The parent budget route uses the same shared finance hierarchy as its report:
`MPFinanceSummaryCard` owns aggregate limit/spend/remaining, while each budget
and recurring series uses `MPFinanceCard`. `BudgetReferencePresentation`
supplies section spacing and the summary accessibility key consumed by the
rendered shared card.

Editors, report routing, stop/delete confirmations, carryover state, recurring
series idempotency, calculations, and API payloads remain unchanged. Focused
unit tests cover safe over-limit percentage behavior and masking. They do not
inspect SwiftUI composition or prove editor/API persistence. Integrated
visual/device acceptance remains a later redesign gate.

Review remediation evidence (2026-10-27): the four owned reference test classes
passed 24/24 and the normally signed `MyPocketIOSTests` target passed 422/422 on
the iPhone 16 iOS 18.5 simulator. The Release warnings-as-errors build,
base-component guard, and project verification also passed.
