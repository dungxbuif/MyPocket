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
| `nested` | Reference geometry: card `p-3`; root `p-1.5` with 40px icon; child list `pt-1.5 pl-10 pr-1 space-y-0.5`; 32px child icon; 2px vertical and curved branch connectors | Account views. |
| `line` | aligned rows, dense 1 spacing | Dense category selector views. |

The component is implemented as `BaseCategoryTree`; consumers pass root,
children, localized labels and optional edit/delete callbacks. A row may supply
an icon and trailing-content slot (for example, a report amount), while
`IconButton`, `IconBadge` and `SurfaceCard` remain atom-owned. Root and child
are not separate cards: the child list is indented by `pl-10`; a 2px vertical
line at `left: 37px`, beginning at the first child row, and a curved branch offset right of the trunk into each child icon
communicate hierarchy. The global colorful icon catalog is retained even though
the legacy export used dark parent badges. A third layout may not be introduced
locally.

A root with no children uses the same horizontal `p-3` card inset but drops the
card's vertical padding; its row keeps only the root row's own `p-1.5` spacing.

Reference exports live under `nested/` and `line/`, each with its own
`code.html` and `screen.png`. New work must reference this folder and
[`BASE_COMPONENTS.md`](../../system/BASE_COMPONENTS.md).
