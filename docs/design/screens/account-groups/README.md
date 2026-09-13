---
artifact_type: screen_design_contract
id: ACCOUNT-GROUPS
status: implemented_owner_uat_pending
owner: shared
trace:
  ticket: ../../../work/tickets/TICKET-01-03-tong-vi-danh-muc.md
  system_design: ../../system/DESIGN.md
  tree: ../../molecules/category-tree/README.md
---

# Account — Quản lý nhóm

## Composition

- Header: balance header hidden; centered `Heading`, pill back `BaseButton`, then `SegmentedControl` for Expense/Income/Debt and a full-width “Nhóm mới” base button.
- Group list: one `BaseCategoryTree` per root using the `nested` layout, with system-key marker mapping and owner-wallet activity text. Root and child rows use compact shared 48px height with reduced text; child rows are visibly indented and use smaller icons. The vertical connector starts at the centre of the parent icon and each horizontal divider begins at the indented child-content column.
- Form: `BaseBottomSheet`, `FormField`, `BaseTextInput`, `BaseSelect`, and `BaseCheckbox` for wallet applicability. The same form composes create and edit; edit provides a danger `BaseButton` for deletion.
- Feedback: `SurfaceCard` plus `BaseButton`; no screen-local card, field or icon-button styling.

## States and copy

The screen handles loading, loaded tree, empty list, load failure/retry, local
required-name validation, saving, API failure, and native destructive-delete
confirmation. System items are visually read-only. A personal group may select
multiple wallets; an empty selection means it is not limited by wallet.

## Runtime proof

- `/account/groups` loaded in the authenticated Vite session on 2026-09-13;
  system roots/children, header and expand controls were present.
- Automated build proof and category API proof are recorded in the linked
  ticket and validation matrix. Owner UAT for create/edit/delete remains open.
