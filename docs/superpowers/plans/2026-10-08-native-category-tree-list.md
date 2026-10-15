# Plan: Native category tree list parity

## Goal

Implement the approved Stitch/web category list controls and hierarchical
presentation using existing APIs, without inventing category archive state.

## Files

- `mobile-ios/MyPocketIOS/Core/Models/FinanceModels.swift`: pure list filtering
  and flattening policy.
- `mobile-ios/MyPocketIOS/DesignSystem/MPCategoryTree.swift`: reusable
  root/child tree base.
- `mobile-ios/MyPocketIOS/Features/Secondary/SecondaryFeatureViews.swift`:
  compose filters and tree in `CategoriesView`.
- `mobile-ios/MyPocketIOSTests/FinanceModelTests.swift` and
  `mobile-ios/MyPocketIOSTests/DesignSystemTests.swift`: deterministic policy
  and base contracts.
- `mobile-ios/MyPocketIOSUITests/LaunchTests.swift`: category list interaction
  regression if the existing authenticated fixture harness supports it without
  adding new fixture semantics.
- `docs/design/`, `docs/work/`, `docs/releases/CHANGELOG.md`: reconcile design,
  evidence, queue and release notes.

## Tasks

1. Add failing pure tests for kind/wallet/search filtering, child/root
   retention and stable depth-first order; implement the smallest pure policy.
2. Add the shared native category tree base and failing DesignSystem contracts;
   verify parent context is not an edit action and child edit callbacks retain
   stable IDs.
3. Compose kind, wallet and search controls in `CategoriesView`, using existing
   API data and the shared tree. Add focused UI proof for filter state and
   category visibility.
4. Reconcile screen/base docs, validation matrix, port audit, backlog and
   changelog; run native tests, Release build, guards and diff check.

## Acceptance

- Expense/income/debt tabs filter the visible tree.
- Wallet scope honors empty `wallet_ids` as all wallets and explicit IDs as
  restricted applicability.
- Search is localized, case/diacritic-insensitive, and matching children keep
  their root context.
- Hierarchy order is deterministic; edit/delete actions remain owner/API-backed.
- No fake inactive toggle, API/schema change, or production call is added.

## Out of scope

Dedicated native editor navigation, transaction/budget tree pickers, and
inactive/archive semantics. These remain separate port backlog items.
