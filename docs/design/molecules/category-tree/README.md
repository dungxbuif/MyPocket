# Category tree

[Source IMG-07](../../system/SOURCE_EXTRACTION.md) · [Base contract](../../system/BASE_COMPONENTS.md) · [Page](../../pages/account-groups/README.md).

BaseCategoryTree composes SurfaceCard, BaseButton row, IconBadge and Text. Root and direct children are one card. Children do not get their own cards. Root icon 40px, child 32px, connector 2px with curved branches. Geometry is in ui/variants.ts, never copied into screen.
Current nested variant uses pl-12; line variant changes vertical row spacing. These are existing named implementation variants; a materially different dense tree requires its own reviewed contract, not a local class override.
Current card radius is 12px; source 24px superseded. Global colorful icon catalog replaces legacy dark parent background.

Props: root, children, layout, mode, selectedID, onSelect; each item has ID, name, subtitle, system/editable/selectable flags, optional icon/tone and trailing React content. `navigation` keeps the group-management behavior: editable rows invoke onSelect and system rows show the locked metadata affordance. `selection` renders eligible rows as BaseButton with `aria-pressed`, a shared selected check, and invokes onSelect; ineligible parent rows remain visible as non-interactive hierarchy context. `CategoryTreeSelector` adds search and a clear action while composing this same base for transactions, proposals and budgets.
No collapse state exists in current code. Export describes chevron rotation for collapse: planned, not implemented. A root without children keeps horizontal inset and drops card vertical padding.
No inactive-filter behavior exists just because the source has “Hiển thị nhóm không hoạt động”.
Text slots truncate/wrap within available width; connector changes require browser check with long labels, childless root and multiple children.
