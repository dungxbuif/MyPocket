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

## Event → effect

List load fetches categories, groups roots by kind and displays direct children. Segment click/keyboard changes filter. Nhóm mới routes to create with current kind. Select row routes to edit by ID.
Edit name/icon/kind/parent in base controls; kind change resets parent. Only same-kind roots excluding self are parent options. Wallet checkboxes toggle selected IDs. Save validates nonempty trimmed name for personal group, then screen invokes API. System save updates wallet applicability only.
Delete requires native confirmation before API; API determines scope and integrity. Cancel/back return to group list. Icon sheet choose updates icon and closes; Escape/backdrop close without changing selection.

## States, copy and proof

Loading “Đang tải nhóm...”; empty “Chưa có nhóm nào.”; failure “Không tải được danh sách nhóm.” with “Thử lại”. Save loading disables CTA; errors surfaced. System hint explains only wallets can change.
No wallet restriction displays “Áp dụng tất cả ví”; selected count displays activity count. Tree is fixed expanded; collapse and inactive filter are not implemented.
Base fixture checks keyboard, shared tokens and sheet focus behavior. Existing product ticket is done after owner UI review and owns CRUD/API proof; this normalization preserves that acceptance and does not claim a new CRUD run.
