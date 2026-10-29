# Category tree

[Source IMG-07](../../system/SOURCE_EXTRACTION.md) · [Base contract](../../system/BASE_COMPONENTS.md) · [Page](../../pages/account-groups/README.md).

BaseCategoryTree composes SurfaceCard, BaseButton row, IconBadge and Text. Root and direct children are one card. Children do not get their own cards. Root icon 40px, child 32px, connector 2px with curved branches. Geometry is in ui/variants.ts, never copied into screen.
Current nested variant uses pl-12; line variant changes vertical row spacing. These are existing named implementation variants; a materially different dense tree requires its own reviewed contract, not a local class override.
Current card radius is 12px; source 24px superseded. Global colorful icon catalog replaces legacy dark parent background.

Props: root, children, layout, mode, selectedID, onSelect; each item has ID, name, subtitle, system/editable/selectable flags, optional icon/tone and trailing React content. `navigation` keeps the group-management behavior: editable rows invoke onSelect and system rows show the locked metadata affordance. `selection` renders eligible rows as BaseButton with `aria-pressed`, a shared selected check, and invokes onSelect; ineligible parent rows remain visible as non-interactive hierarchy context. `CategoryTreeSelector` adds search and a clear action while composing this same base for transactions, proposals and budgets.

Native iOS uses the shared `MPCategoryTree` DesignSystem base for category
management, with the same root/child hierarchy, stable row IDs and
parent-context rows. Editable rows navigate as a whole to the dedicated
category editor and show a trailing chevron; list-level pencil/delete controls
are not rendered. Personal-category delete stays in the editor header.
`CategoryListPresentation` owns local kind, wallet and accent-insensitive
search filtering; scope subtitle is derived from the same `wallet_ids` data as
web (`Áp dụng tất cả ví` for an empty set, otherwise `Hoạt động trong N ví`).
Both tree implementations are fixed expanded and preserve stable depth-first
order.

Native transaction, budget, recurring and credit editors compose `MPCategoryPicker` with
`MPCategorySelectionTree` inside the shared form-bottom-sheet base. The picker
accepts optional selectable IDs, so transaction type/wallet rules remain
authoritative and a non-selectable parent can only provide context. Budget
scope uses the existing “Tất cả danh mục” clear action and nil category payload.
Each depth-zero root is rendered as its own shared `MPSurfaceCard`, matching
the web `CategoryTreeSelector` composition of one `BaseCategoryTree` per root;
the picker keeps stable depth-first groups and exposes each group as
`category-picker-group-<root-id>` for accessibility/UI automation.
Picker eligibility is modeled separately from hierarchy/search context: an
eligible ancestor retained only to explain a matching child remains selectable
as in web, while a wallet/type-ineligible parent is context-only. Selectable
rows use `Nhóm mặc định` or `Nhóm cá nhân`; an ineligible parent uses
`Không áp dụng cho lựa chọn này`, matching the web selector rather than
repeating the transaction kind. Budget's `Tất cả danh mục` clear action uses
the transparent, action-colored, leading-aligned ghost button base.
No collapse state exists in current code. Export describes chevron rotation for collapse: planned, not implemented. A root without children keeps horizontal inset and drops card vertical padding.
Category management supports the “Hiển thị nhóm không hoạt động” control.
Inactive rows are visible only when requested; an inactive parent remains
non-selectable hierarchy context for an active child. Transaction, budget,
credit, recurring and AI-entry selection paths exclude inactive categories.
Existing transaction and recurring edits may retain the exact unchanged
archived category reference; this does not make archived rows newly selectable.
Text slots truncate/wrap within available width; connector changes require browser check with long labels, childless root and multiple children.
