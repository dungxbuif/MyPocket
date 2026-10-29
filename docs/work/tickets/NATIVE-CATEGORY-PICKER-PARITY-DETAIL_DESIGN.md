---
artifact_type: detail_design
id: NATIVE-CATEGORY-PICKER-PARITY
status: in_progress
owner: shared
trace:
  backlog: ../BACKLOG.md
  audit: ../PORT-100-AUDIT.md
  transactions: ../../design/pages/transactions/README.md
  budgets: ../../design/pages/budgets/README.md
  tree: ../../design/molecules/category-tree/README.md
  validation: ../VALIDATION_MATRIX.md
---

# Native category picker parity — detail design

## Intent and scope

Replace flat category menus in native transaction, budget, recurring and credit-entry editors with a
searchable root/child category picker matching the shared web
`CategoryTreeSelector` contract. Preserve category IDs, transaction eligibility,
budget-wide clear semantics, API calls and financial rules.

Transaction picker candidates stay limited to the current transaction kind.
Only IDs allowed by `TransactionEditorPolicy.applicableCategories` can be
selected; a non-applicable parent may remain visible only as hierarchy context
when an applicable child is present. Budget scope lists expense categories and
offers “Tất cả danh mục” as an explicit clear action. Search is local,
case/diacritic insensitive, and retains ancestry for a matching child.

## Component and interaction

- Add a shared native selection variant to the category-tree design system.
  It renders rows as shared-base buttons, marks the selected ID, prevents
  selecting hierarchy-only parents and does not expose management edit/delete
  actions.
- Compose the selection tree with shared search/clear controls inside the
  existing shared form-sheet presentation. Callers own selected-ID state and
  decide when to close; the picker does not make API calls.
- Transaction form preserves its current row and glyph. Tapping opens the
  category sheet; selecting an eligible category updates the same `categoryID`
  then closes. Wallet/type changes continue to use the existing
  `requiredCategoryID` rule.
- Budget form preserves the current scope label. Tapping opens the category
  sheet; selecting an expense category sets `categoryScopeID`, while the clear
  action restores the existing `__all__` sentinel and closes.
- Recurring form preserves `categoriesForType` as its eligibility source.
  Changing type/wallet continues to reset an invalid selection through the
  existing `requiredCategoryID` policy.
- Credit-entry form keeps the current action sheet and all draft fields. Its
  category row switches the content of that same sheet to the shared picker
  (never a nested sheet); selection returns to the preserved entry form.
  Purchase/fee/interest use expense categories and refund uses income
  categories, further restricted by the existing credit-wallet applicability
  filter. Payment remains unchanged.

## Backend and data contract

No backend, migration or public API change is expected: `/categories` already
returns owner-scoped `parent_id`, `kind`, `wallet_ids`, and identity; existing
transaction/budget/recurring/credit-entry write endpoints accept the same
category IDs. Keep server validation authoritative. Do not add inactive
semantics or new transaction types. Change backend only if implementation
verification reveals a contract mismatch; any such change must retain
owner-scoping and server-side category applicability checks.

## Verification

- RED/GREEN presentation tests cover hierarchy context, type/eligibility,
  search normalization, stable order, selected state and clear semantics.
- API-backed iPhone UI tests select an eligible category for a transaction and
  a scoped category for a budget, then verify the submitted category IDs through
  a fresh owner-scoped API read.
- Run full native unit/contract suite, focused API-backed UI flows, Release
  warnings-as-errors, project/base-component guards and `git diff --check`.
- Reconcile screen/base docs, port audit, backlog, validation matrix, context
  and changelog after implementation.

## Visual parity follow-up — captured iPhone 16 picker (2026-10-08)

The first attached native picker screenshot confirms separate root cards, but
reveals two remaining differences against the existing web component contract:

- Native repeats the transaction kind (`Chi`) below every row. Web uses
  `Nhóm mặc định` / `Nhóm cá nhân` for selectable categories and
  `Không áp dụng cho lựa chọn này` for an ineligible parent.
- Native renders `Tất cả danh mục` as a centered green-soft filled control.
  Web renders this as a transparent, action-colored, leading-aligned ghost
  control across the available row width.

Keep these differences in the existing shared picker/base: carry eligibility
separately from hierarchy context, derive the category subtitle from system
and selectable state, expose a reusable native `ghost` button style, and render
the clear label leading-aligned. No API, persistence, finance policy or schema
change is intended. Verify labels through the real API-backed UI accessibility
tree and capture/inspect the picker screenshot; unit tests cover the pure
subtitle/state mapping.

## Exclusions

Management category filters are covered by
`NATIVE-CATEGORY-TREE-PARITY-DETAIL_DESIGN.md`. This slice does not change the
web app, category management edit route, API, schema, permissions or inactive
status.

## Verification record — 2026-10-08

- RED/GREEN selection-policy tests cover kind/eligible IDs, Vietnamese
  diacritic-insensitive child search and parent-only context. Initial RED:
  `/tmp/mypocket-category-picker-red.xcresult`; model tests pass at
  `/tmp/mypocket-category-picker-green-model.xcresult`.
- API-backed iPhone 16 UI verifies transaction category and budget
  category-scope persistence, 2/2 at
  `/tmp/mypocket-category-picker-api-ui.xcresult`.
- Recurring schedule category selection, persisted schedule category ID and
  generated transaction category ID pass 1/1 at
  `/tmp/mypocket-category-picker-recurring-ui.xcresult`.
- Credit-entry UI opens the shared hierarchy in the existing action sheet,
  preserves amount/type/note while choosing, and verifies the saved category
  ID through a fresh statement/API read. Payment remains a regression-tested
  unchanged path.
- RED: the new picker assertion failed on iPhone 16 because the credit form
  exposed a flat `Menu` rather than the shared search field at
  `/tmp/mypocket-credit-category-red/Logs/Test/Test-MyPocketIOS-2026.10.08_10-21-00-+0700.xcresult`.
- GREEN: the API-backed credit entry/payment test passes 1/1 at
  `/tmp/mypocket-credit-category-green/Logs/Test/Test-MyPocketIOS-2026.10.08_10-24-41-+0700.xcresult`;
  it verifies amount/note draft preservation and the persisted category ID
  through a fresh statement read.
- Full native unit/contract target after the credit picker: PASS, 137/137,
  no skips/failures at
  `/tmp/mypocket-credit-category-unit/Logs/Test/Test-MyPocketIOS-2026.10.08_10-25-38-+0700.xcresult`.
- Release warnings-as-errors build: PASS at
  `/tmp/mypocket-credit-category-release`; base-component tests/guards,
  project verification and `git diff --check` pass.
- The local API used only `mypocket_stage_v1` and Redis DB 15 on loopback
  `127.0.0.1:8081`; post-test fixture wallet and transaction counts are both
  zero, and the temporary API is stopped. No backend/schema change was needed
  because the current API already carries hierarchy and accepts selected IDs.
- Backend compatibility suite: `TEST_DATABASE_URL` targeting only local
  `mypocket_stage_v1` — `go test ./...` passes; production was not contacted.

### Copy and clear-action parity follow-up — 2026-10-08

- RED: the new API-backed UI assertion failed on iPhone 16 because the native
  accessibility label omitted the web subtitle (`Ăn uống` instead of
  `Ăn uống, Nhóm cá nhân`); the prior screenshot also showed transaction-kind
  copy and a filled, centered clear control.
- PASS: eligibility is now carried separately from hierarchy/search context;
  system, personal and ineligible subtitles match web. Eligible ancestry
  remains selectable, while truly ineligible parent context stays disabled.
  Budget clear uses the shared transparent/action-colored `MPButton.ghost`
  style and a leading-aligned label.
- PASS: API-backed budget create/update/delete UI passes 1/1 at
  `/tmp/mypocket-picker-parity-budget-ui-20261008.xcresult`; the captured
  unfiltered picker image was inspected at
  `/tmp/mypocket-picker-parity-budget-ui-attachments-20261008/E29DE286-CD42-42B0-BCED-F9CE25C52337.png`.
  Recurring and wallet-detail transaction regressions pass 2/2 at
  `/tmp/mypocket-picker-parity-crossflow-ui-20261008.xcresult`.
- PASS: native unit/contract tests 144/144 at
  `/tmp/mypocket-picker-parity-unit-20261008.xcresult`; Release
  warnings-as-errors, `verify-project.sh`, both base-component scripts and
  `git diff --check` pass. Local API/database/Redis fixture only; no public API,
  backend or migration change. Physical-device/owner acceptance remains open.
