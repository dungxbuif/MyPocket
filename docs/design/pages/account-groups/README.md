---
artifact_type: screen_design_contract
id: ACCOUNT-GROUPS
status: implemented
owner: shared
---

# Account — Quản lý nhóm

[Ticket](../../../work/tickets/TICKET-01-03-tong-vi-danh-muc.md) · [Normalization](../../../work/tickets/UI-BASE-01-DETAIL_DESIGN.md) · [Tree](../../molecules/category-tree/README.md) · [Edit group](../../molecules/edit-group/README.md).

## Routes and composition

/account/groups: PageBackHeader back to account, SegmentedControl expense/income/debt, outline BaseButton “Nhóm mới”, BaseCategoryTree per root.
New/edit routes: PageBackHeader, CategoryEditForm and ApplicableWalletsCard. Personal group exposes delete IconButton in header; system metadata is read-only. Icon picker uses BaseBottomSheet. The whole edit form is not inside a sheet.

On native iOS, create/edit uses a dedicated editor destination rather than an
inline form above the category tree. A selected row opens the editor; the shared
header returns to the list, and personal-category deletion is available from
the editor header. The web route remains the behavioral/API source of truth.
Wallet applicability is a visible `MPWalletApplicabilityCard` composed from
shared `MPCheckboxRow`s. Empty `wallet_ids` means all owner wallets; a partial
selection is sent as explicit owner wallet IDs and survives a fresh API read.

The native list uses the shared kind selector, then the add action, wallet
scope, search and category tree. The native kind/wallet/search controls are
kept even where the web route does not render them, matching the Stitch
reference without changing owner-scoped data.
Filtering is local over the owner-scoped category response: an empty
`wallet_ids` list applies to every wallet, matching children retain their root
as non-editable hierarchy context, and results are shown in stable depth-first
order. The tree is virtualized, so an unfiltered row outside the viewport is
materialized only when scrolled into view; UI tests should search or scroll
before asserting row actions. Inactive categories are hidden by default and
appear when “Hiển thị nhóm không hoạt động” is enabled; an inactive parent
needed to show an active child remains non-editable hierarchy context.

Management rows are full-row navigation actions to the dedicated editor.
Their subtitle mirrors web: empty `wallet_ids` shows “Áp dụng tất cả ví”, and
an explicit scope shows “Hoạt động trong N ví”. Delete is available only in
the personal-category editor header; list rows do not expose pencil/trash
controls.

## Event → effect

List load fetches categories, groups roots by kind and displays direct children. Segment click/keyboard changes filter. Nhóm mới routes to create with current kind. Select row routes to edit by ID.
Edit name/icon/kind/parent in base controls; kind change resets parent. Only same-kind roots excluding self are parent options. Visible shared-base wallet checkbox rows toggle selected IDs; empty selection means all wallets. Save validates nonempty trimmed name for personal group, then screen invokes API. System save updates wallet applicability only.
Delete requires native confirmation before API; API determines scope and integrity. Cancel/back return to group list. Icon sheet choose updates icon and closes; Escape/backdrop close without changing selection.

## States, copy and proof

Loading “Đang tải nhóm...”; empty “Chưa có nhóm nào.”; failure “Không tải được danh sách nhóm.” with “Thử lại”. Save loading disables CTA; errors remain visible in the editor for retry. System hint explains only wallets can change.
No wallet restriction displays “Áp dụng tất cả ví”; selected count displays
the number of selected wallets. Tree is fixed expanded. Personal categories
can be archived/restored from their editor; the “Hiển thị nhóm không hoạt
động” control reveals inactive categories with a “Đã lưu trữ” marker that is
also present in the row's accessible label. Archive state is independent for
parents and children: an inactive parent remains hierarchy context when a
child is active. Historical transactions, reports, budgets and recurring
schedules keep their category reference. New assignments exclude inactive
categories; an existing transaction or recurring schedule can retain its
unchanged archived category while editing.
API-backed iPhone 16 visual/behavior evidence is recorded in
`docs/work/VALIDATION_MATRIX.md`: the screenshot captures compare the list and
editor to Stitch, and the CRUD flow verifies row navigation, exact wallet-scope
copy, filters, rename and editor-only delete. The global feedback bubble still
overlays a category row in the captured long list; it can be reached by
scrolling but remains a separate mobile-overlay parity item for overall UAT.
