---
artifact_type: audit
id: NATIVE-VISUAL-AUDIT-20261019
status: verified
owner: shared
updated: 2026-10-19
---

# Native visual/base-component audit

## Scope

This audit covers the native iOS feedback about oversized cards/items, the
floating feedback bubble reserving scroll space, bottom-tab presentation,
portrait-only layout, compact typography, and route entry/back behavior. It
does not change an API, financial contract, or server state.

## Current contract

| Area | Authoritative base | Result |
| --- | --- | --- |
| Surface/card | `MPSurfaceCard` and `MPFinanceCard` | All reference cards use the shared web-matching 24pt reference radius and 8pt vertical/12pt horizontal insets. No new per-screen card geometry is introduced. |
| Transaction/item | `MPTransactionRow` | One shared 44pt minimum, content-driven expansion, and the approved 6pt separation inset. Dynamic Type remains content-driven. |
| Feedback bubble | `MPFeedbackBubbleLayout` | Stable lower-right overlay; footer clearance moves only the bubble. The scroll viewport itself is not reduced and no obstacle lane is reserved. |
| Navigation/tabs | `MPBottomNavigation` and `MoreDestination` | Four primary tabs, shared add action, accessibility menu for large Dynamic Type, and typed `NavigationStack` routes for secondary screens. |
| Orientation | `PortraitOnlyApplicationDelegate` plus target `Info.plist` | Portrait is the only supported interface orientation; feature routes do not provide landscape overrides. |
| Typography | `MPFont` and `MPText` | Native SF Pro/System metrics are centralized and Dynamic Type is preserved; feature-local system font overrides are rejected by the base guard. |

## Findings

1. The large-looking reference cards are intentional: `--radius-reference-card`
   in the web token set is 24px, and native uses the same named reference
   token. Reducing it ad hoc would break the approved web/native contract.
2. Transaction rows intentionally retain a 44pt hit target plus the shared
   6pt visual separation. Removing that inset would regress the earlier row
   crowding feedback and would not improve touch safety.
3. The feedback trigger is an overlay, not a layout sibling. Its geometry is
   bounded to the content viewport and excludes the pinned footer clearance;
   there is no reserved scroll lane to remove.
4. Secondary routes have one typed route destination each. Back is supplied by
   the shared route bridge and explicit screen header; editors block the edge
   gesture while busy/dirty.

## Verification additions

`ReferenceDisplayTests` now locks these contracts with geometry and route
identity assertions. The focused suite passed 8/8 with zero failures in
`/tmp/mypocket-native-visual-audit-20261019.xcresult`. The existing portrait,
feedback-frame, transaction-row, base-component, and project verification tests
remain required. A screenshot or physical-device visual sweep is still owner
acceptance evidence and is not claimed by simulator unit tests.

## Reconciliation

No safe product geometry change was necessary in this audit. If owner review
later changes the approved reference-card radius or row inset, update the web
token contract and this document first, then change the shared native token and
rerun the full native suite.
