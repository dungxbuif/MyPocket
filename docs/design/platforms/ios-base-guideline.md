# MyPocket native base guideline

Status: owner approved 2026-10-14: keep the app palette, improve UI/UX across
native iOS using the proposed common families. [Screen registry](ios-screen-registry.md).
Goal: a small set of composable families, not a unique component per feature.
Existing Go business/auth contracts remain authoritative; UI must not invent totals.

## Five families

| Family | Shared responsibility | Existing bases to keep/consolidate |
| --- | --- | --- |
| Shell/layout | Bounded viewport, one header anatomy, gutters, pinned footer, section rhythm | MainTabView/MPBottomNavigation/MPScreenHeader/MPFormBottomSheet |
| Surface/list | One flat group surface; reusable title/value/meta/icon row and semantic finance mapping | MPSurfaceCard/MPMenuRow/MPTransactionRow/MPWalletRow/MPMetric |
| Form/input | Labels/errors, text/password, money, decimal quantity, integer count, note, dates | MPTextField/MPSecureField/MPAmountField/MPAmountKeypad/MPTextEditor/MPDateField |
| Selection/presentation | Segmented/filter tabs, checkbox, searchable selection, short sheet vs long editor route, confirmation | MPSegmentedControl/MPCheckboxRow/MPOptionSelector/MPCategoryTree/mpFormBottomSheet/mpFullScreenCover/mpConfirmDialog |
| Feedback/state | Loading/empty/error/disabled, attachment source+thumb, chat composer/messages and action results | MPAsyncStates/MPAttachmentViews/MPMarkdown/shared financial cards |

These are families, not a demand for five giant files/components. Features own
domain data/validation/actions; bases own appearance and interaction geometry.
Do not rename every existing component or duplicate it as a new v2 hierarchy.

## Geometry and visual authority

- Keep current semantic green/neutral palette, Manrope and tokenized text styles.
  No new color/font choice from search output; existing owner choices prevail.
- Page horizontal gutter16; card horizontal inset12, vertical8. This preserves
  compact vertical rhythm while increasing horizontal breathing room requested
  in feedback. Section gap8, internal row gap4. Card corners12/button8.
- Header, group heading and row leading icon use consistent alignment columns;
  row title starts after the shared32 icon+8 gap. Remove per-screen heading inset,
  negative padding and arbitrary horizontal overrides.
- Minimum interactive target44pt. Normal finance row44 minimum, content-driven
  expansion for long copy/Dynamic Type. No fixed maximum height/font shrinking.
- No card-in-card. A section owns one surface; internal metrics/rows/dividers
  are flat content. Cards are not a universal wrapper for every label/action.
- One content scroll area per screen. Navbar bounds content; floating feedback
  does not reserve a lane. Keyboard-safe composer/editor actions remain reachable.
  Feedback hides during keyboard input and returns on dismissal; pinned footer
  height determines its at-rest clearance, without reducing the scroll viewport.
  Current avoidance implementation also reports shared button/preview frames and
  tries alternative edge positions. This remains unaccepted: proposal render
  covers description input despite196/196 selected tests. Review protected input/
  content regions and saturated viewport behavior before another placement fix.

## Interaction rules

- One primary entry per destination: profile→account; bell→notifications;
  Assistant tab→advisor. Remove duplicate More rows, retain route identifiers
  for intentional deep links/navigation callbacks.
- Long create/edit form opens dedicated editor, not expanding above a list.
  Short choice/confirmation opens shared sheet/dialog. Save/cancel/dismiss rules,
  dirty form and pending submit states are the same across consumers.
  In a mounted management screen, cancel leaves input unsaved but keeps the local
  draft for reopening the same item. A different item/create starts a distinct
  draft; success resets it. Jar month changes invalidate the draft. No durable
  draft/restart promise is implied; offline persistence is a separate contract.
- Money is integer minor-unit/VND input formatted through one money base/parser.
  Quantity is decimal, frequency interval integer, allocation percent bounded
  percentage: separate input semantics, same field styling. Never strip decimals
  from investment quantity or show a currency keypad for a recurrence count.
- Attachment chooser defaults Photos; visible Files tab supports PDF/import.
  Thumbnail/remove, remaining count, file errors and loading belong to the base.
- Advisor overview stays pinned/compact; chat takes remaining height with
  suggestions inside the chat region. Composer grows a few lines then scrolls;
  no full-sized note editor for every message. Read-only AI boundary unchanged.
- Resolution checkbox means owner confirmation, not dev status. Face ID state
  must remain visible and independent of app-lock/refresh/session/logout.

## Evidence and guard policy

Every registry screen needs entry/back/save/cancel/error/empty/pending specs;
each migrated screen requires populated small-phone screenshot and UI action
proof, plus API state re-read for writes. Shared-base unit tests cover normal/
long text, Dynamic Type, 44pt targets and no nested card expansion. Static raw
control guard alone is insufficient. No release note until owner acceptance.

UI/UX Pro Max was used for native touch, consistent geometry and presentation
guidance. Two design-system searches returned marketing/enterprise landing
patterns rather than a valid native in-app pattern; those outputs were not
persisted/adopted. This proposal uses verified SwiftUI navigation/spacing advice
and native pro-rules with the project's existing palette, not an invented match.
