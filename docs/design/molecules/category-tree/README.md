---
artifact_type: molecule_design_contract
id: CATEGORY-TREE
status: active
owner: shared
---

# Category Tree

Canonical base molecule for parent/child category presentation. It merges the
two former design exports into one implementation contract and owns both
reference variants.

| Variant | Source geometry | Use |
| --- | --- | --- |
| `nested` | aligned rows, compact 0.5 spacing | Account views. |
| `line` | aligned rows, dense 1 spacing | Dense category selector views. |

The component is implemented as `BaseCategoryTree`; consumers pass root,
children, localized labels and optional edit/delete callbacks. A row may supply
an icon and trailing-content slot (for example, a report amount), while
`IconButton`, `IconBadge` and `SurfaceCard` remain atom-owned. Root and child
are not separate cards: every row aligns to the same left edge and text column;
a vertical tree line plus the 40px root icon and 32px child icon communicate
hierarchy. Both root and child rows use the same fixed 56px height. A third
layout may not be introduced locally.

Reference exports live under `nested/` and `line/`, each with its own
`code.html` and `screen.png`. New work must reference this folder and
[`BASE_COMPONENTS.md`](../../system/BASE_COMPONENTS.md).
