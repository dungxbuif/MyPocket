---
artifact_type: design_system_contract
id: DESIGN-BASE-COMPONENTS
status: active
owner: shared
trace:
  source_artifacts:
    - ../atoms/base-cards/code.html
    - ../atoms/amount-keypad/code.html
    - ../molecules/budget-meter/code.html
    - ../molecules/transaction-form-rows/code.html
---

# Base component technical contract

This is the implementation contract extracted from the current design export
images and HTML. A screen composes these bases; it must not recreate their
visual CSS or one-off variants.

| Base | Required geometry and tokens | Variants / behavior |
| --- | --- | --- |
| `SurfaceCard` | white card, maximum `rounded-2xl` / 16px, `0 4px 20px rgba(0,0,0,.03)`, subtle slate divider | list, form, metric and settings use the same container; padding/elevation are named variants. |
| `IconBadge` | 28–40px icon hit area; soft tinted background plus matching border | semantic `wallet`, `income`, `expense`, `warning`, `neutral` and global category palette; no screen supplies arbitrary color classes. |
| `BaseButton` | Manrope 600/700; primary green; touch target at least 44px; inline flex centres icon-plus-label content | primary, secondary, ghost, danger; size and loading are named props. |
| `Heading` | semantic `h1`/`h2`/`h3`; slate hierarchy and Manrope weight are global | `screen`, `section`, `field`; caller can supply semantic tag and additive `className`, never local base typography. |
| `FormField`, `BaseTextInput`, `BaseSelect` | label plus 1px slate border, 12px inset control, emerald focus state and `rounded-lg` control | `FormField` owns label association/layout; input/select forward native props and allow additive classes. |
| `IconButton` | 40px circular interaction target, subtle border/shadow, visible focus state | accessible label is required; additive `className` is allowed for named composition contexts. |
| `FormSelectorRow` | fixed 36–40px leading slot, flexible content, trailing value/chevron/switch | row target is 44–48px; selector, date and toggle are props rather than bespoke markup. |
| `Toggle` | iOS style 48×24px | active uses emerald; keyboard and label support are mandatory. |
| `SegmentedControl` | pill tabs, selected state visible | arrow, Home and End keys move the active tab. |
| `BaseCategoryTree` | white `rounded-3xl` tree card; 40px root icon and 32px child icon; 2px vertical connector plus curved branches | `nested` and `line` are the only layout variants. The `nested` geometry follows its source export: card `p-3`, root `p-1.5`, children `pt-1.5 pl-10 pr-1`, 0.5 gap, and no horizontal dividers. A root without children keeps horizontal inset but drops card vertical padding. Icon color/presentation stays sourced from the global category catalog. Optional icon and trailing-content slots support reporting without recreating a tree. |

## Direct artifact rules

- Typography is Manrope 400/500/800; financial values use a stable numeric
  presentation.
- Canvas is the light neutral surface; cards are white; text hierarchy uses
  slate 900/800 then slate 500/400.
- The base-card export names green `#10b981`/`#059669`; amount keypad names
  `#4caf50`; budget/form exports use `#22c55e`. They are semantic variants of
  the global brand palette, not per-screen ad-hoc values.
- Amount keypad supports `000`, arithmetic `+ − × ÷`, and `XONG` to commit a
  formatted VND result.
- Budget semi-circle is 180° SVG. Its time marker is
  `(currentDay / totalDays) × 100`; warning starts above that marker and danger
  above 100%.
- Category tree has two approved layout variants. The `nested` variant follows
  its source geometry: a 2px vertical connector at `left: 37px` beginning at
  the first child row, curved child branches offset right of that trunk, and no horizontal dividers. A
  consuming component selects the named variant; it does not hand-roll a third
  layout.

Every completed screen must update its matching `docs/design/screens/<screen>`
artifact with composition, states and copy, as required by [README](../README.md).

`SurfaceCard` additionally forwards semantic section HTML attributes such as
`role` and `aria-*`; this is the permitted customization path for accessible
status cards. Visual shape, elevation and spacing still use its named variants.
