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

Known gaps: automatic repeating periods, server notifications and forecasts remain unsupported; no fake controls/data for them. Reports not mounted. Browser/owner visual acceptance pending.

Metadata edits retain the original timestamp boundaries when corresponding visible date fields are unchanged, even after switching browser timezone. Intentional date edits use the current browser-local calendar. Owner account timezone persistence remains separate. API persistence proof and real empty/editor browser inspection recorded in the linked design; full browser CRUD/owner sign-off pending.
