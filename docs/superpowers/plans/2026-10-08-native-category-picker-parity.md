# Native category picker parity plan

1. Add failing pure presentation/base tests for hierarchical eligible category
   selection, parent context, search, selected-ID marker, and budget clear.
2. Implement the native picker base using existing shared fields, buttons,
   surfaces and category-tree rows; keep filtering/selection policy in a pure
   helper and finance/API decisions in callers.
3. Replace transaction, budget and recurring flat selectors with the shared
   picker while preserving current IDs, wallet/type eligibility and `__all__`.
4. Add API-backed iPhone UI coverage for transaction, budget and recurring
   selection; assert persisted state through owner-scoped reads.
5. Run native unit/contract, focused API UI, Release warnings-as-errors,
   project/base guards and diff check; reconcile all docs.

## Boundaries

- No backend/schema/API changes unless investigation proves the existing
  owner-scoped category contract cannot support the described flow.
- Do not edit web or unrelated in-progress workspace files.
- Keep the inactive category toggle and management route/layout decision out of
  this picker slice.

## Follow-up — one selection card per root category (2026-10-08)

Current-source review found a visual parity gap: web `CategoryTreeSelector`
composes one `BaseCategoryTree` surface per root, while native
`MPCategorySelectionTree` wraps every root in a single `MPSurfaceCard`.

1. Add a failing pure presentation test proving roots and descendants are
   partitioned into stable per-root groups without dropping hierarchy-context
   nodes or changing their order.
2. Render each group through the existing shared card/tree bases and expose a
   stable accessibility identifier per root group for screen-level proof.
3. Extend the API-backed picker UI flow to assert separate root surfaces,
   attach a named simulator screenshot of the unfiltered picker for visual
   comparison, and retain category selection/persisted-ID assertions.
4. Update category-tree, picker, iOS renderer, port-audit, validation, backlog,
   context and changelog docs. No backend/API/schema/migration behavior changes.

Verification: focused XCTest RED/GREEN, budget/transaction/recurring API UI
regressions, inspect the exported picker screenshot against the web/Stitch
contract, full native unit/contract target, Release warnings-as-errors,
project/base-component guards and `git diff --check`.

## Follow-up — picker copy and clear-action parity (2026-10-08)

The captured native picker screenshot showed transaction-kind subtitles on
every row and a filled, centered budget clear action. Match the web tree's
system/personal/ineligible subtitles and its transparent leading-aligned ghost
clear action through shared model/base components. Preserve selection,
hierarchy, API payloads and financial behavior. Add failing API-backed label
assertions and pure presentation tests before implementation, then rerun picker
UI/API flows, unit/Release checks, base/project guards and diff validation.

Status: complete. API-backed budget CRUD verifies the exact selectable-row
accessibility label and captures each root card in
`/tmp/mypocket-picker-parity-budget-ui-attachments-20261008/`; recurring and
wallet-detail transaction API UI regressions pass 2/2. Native units pass
144/144, Release warnings-as-errors and project/base/diff guards pass. Evidence
and limitations are recorded in `docs/work/VALIDATION_MATRIX.md`.
