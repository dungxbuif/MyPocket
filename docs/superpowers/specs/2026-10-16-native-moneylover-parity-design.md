# Native Money Lover parity design

Status: approved direction; implementation plan pending review  
Date: 2026-10-16

## Goal

Bring the native iOS presentation close to the supplied Money Lover native
screens at pixel level while keeping MyPocket's existing business rules,
owner isolation, routes, API contracts, offline behavior, and AI flows. The
Money Lover app is a visual reference only; no Money Lover data, branding
credentials, or proprietary runtime code is copied.

## Reference contract

- Native reference capture: `/tmp/moneylover-account.png` and the captures
  recorded in `docs/design/reference-survey/2026-10-16-money-lover-iphone.md`.
- MyPocket behavior remains authoritative for wallets, transactions, budgets,
  jars, reports, credit, feedback, authentication, and route guards.
- Every migrated screen gets a reference capture, a named visual checklist,
  and a live-device screenshot before release.

## Shared design system

`mobile-ios/MyPocketIOS/DesignSystem` remains the only place for visual
primitives. The parity pass will consolidate:

- `MPFont`: iOS system typography, Vietnamese metrics, Dynamic Type-safe
  hierarchy and consistent weights.
- `MPColor`: canvas, surface, ink, secondary, action, danger and disabled
  semantic tokens; no screen-local colors.
- `MPMetrics`: page inset, section rhythm, row height, card radius, divider,
  shadow and hit target.
- `MPSurfaceCard`, `MPMenuRow`, `MPProfileCard`, `MPBottomNavigation`,
  `MPButton`, `MPIconButton`, `MPTextField`, and `MPSegmentedControl` as the
  only interactive/layout bases for the migrated screens.
- A small icon registry mapping each Money Lover reference glyph to a native
  SF Symbol or an owned vector asset; emoji and ad-hoc symbols are forbidden.

## Screen rollout

1. Shell and Account: status-safe header, profile/premium card, grouped menu,
   bank-link row, bottom navigation and feedback bubble.
2. Overview, Transactions and Wallet Detail: balance header, wallet scope,
   period controls, transaction rows, add action and detail cards.
3. Budgets, Jars, Reports, Credit and Savings: metric/card anatomy and
   specialized empty/loading/error states.
4. Categories, Recurring, Months, Travel, Portfolio, Notifications and
   Assistant: shared rows, selectors, charts and action confirmations.
5. Auth, AI entry, Feedback, Settings and all route/deep-link states.

Each wave removes screen-local font/radius/padding declarations as it lands.
No API or schema change is allowed for visual parity work.

## Verification

- `git diff --check` and base-component usage audit for every wave.
- Native unit suite and Release `xcodebuild` with warnings-as-errors.
- Screenshot comparison on the iPhone 15 Plus and iPhone 16 Simulator at the
  same orientation and scale as the reference.
- Route sweep verifies auth guards, back behavior, tab selection, add action,
  feedback bubble, and deep links after each shell change.
- Release only after the live-device screenshot checklist is reconciled; do
  not claim 100% parity when a reference screen is not captured or measured.
