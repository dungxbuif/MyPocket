---
artifact_type: detail_design
id: NATIVE-CATEGORY-TREE-PARITY
status: in_progress
owner: shared
trace:
  backlog: ../BACKLOG.md
  phase: ../phases/PHASE-002-native-ios-swiftui.md
  audit: ../PORT-100-AUDIT.md
  screen: ../../design/pages/account-groups/README.md
  base: ../../design/molecules/category-tree/README.md
  validation: ../VALIDATION_MATRIX.md
  editor_route: NATIVE-CATEGORY-EDITOR-ROUTE-DETAIL_DESIGN.md
---

# Native category list/tree parity — detail design

> **Status reconciliation (2026-10-08):** The initial list/tree-only scope
> below predates approval and implementation of reversible category archival.
> Its archived-category deferral is superseded by
> [the category archive contract](../../superpowers/specs/2026-10-08-category-archive-design.md).
> Current behavior is documented in the account-groups and category-tree
> screen contracts and verified in the category-archive section of
> `VALIDATION_MATRIX.md`.

## Intent and scope

Bring the native category-management list closer to the supplied Stitch screen
and the web category-tree contract. The list will expose category kind, wallet
scope and text search, then render the matching root/child tree. This uses the
existing owner-scoped category and wallet APIs; no backend/schema change is
needed for these controls.

The current `CategoryModel.walletIDs` semantics remain authoritative: an empty
list means all wallets. Selecting a specific wallet filters the list to groups
whose wallet scope is empty or includes that wallet. Search is case- and
diacritic-insensitive and preserves tree ancestry for matching children.

The original design deferred inactive-category behavior until its semantics
were approved. That decision was superseded: the approved soft-archive contract
adds persisted `is_active`, the owner-scoped status endpoint, selector
eligibility rules and history-preservation behavior. See the linked archive
contract and validation evidence; do not use the initial deferral below as the
current product decision.

## Visual audit follow-up — 2026-10-08

API-backed iPhone 16 captures were compared with the Stitch category-list
screen and the web `GroupManagementPanel`/`BaseCategoryTree`. They show three
remaining mismatches in the management route: native places “Nhóm mới” before
the kind selector (web places it after), native rows expose pencil/delete
buttons instead of opening the editor from the row, and native subtitles show
the transaction kind (“Chi”) instead of the wallet applicability already
present in `wallet_ids`. The category editor is a dedicated route, so list
delete also bypasses the specified editor action placement.

The follow-up matches the web row/action order and full-row navigation,
removes list-level delete controls, and derives subtitle copy from existing
wallet scope (`Áp dụng tất cả ví` for empty `wallet_ids`, otherwise
`Hoạt động trong N ví`). Keep the wallet/search/kind controls and their
behavior. This portion was a native presentation/navigation change and did not
need an API update. Category archive was subsequently specified and implemented
as a separate, approved backend/web/native contract; see the linked archive
specification.

Proof: the API-backed iPhone UI test now asserts row navigation, editor delete
placement, scope subtitle and filter order; screenshots are attached for Stitch
comparison. Focused UI/unit tests, Release warnings-as-errors,
project/base-component guards and diff checks passed at the implementation
checkpoint below. The category archive follow-up has its own backend/web/native
verification record in the validation matrix.

## Components and behavior

- Add a native `MPCategoryTree` DesignSystem base for root/child rows, using
  project tokens and `MPSurfaceCard`, `MPText`, `MPIcon`, and shared row action
  bases. It supports navigation rows and hierarchy-only parent context; the
  category screen owns API callbacks and edit/delete confirmation.
- Add a pure category-list presentation policy that filters by kind, wallet,
  and search, then emits a stable depth-first root/child sequence. A matching
  child retains its root as context even when the root label does not match.
- `CategoriesView` starts on expense groups, offers expense/income/debt kind
  segments, a wallet filter whose first option is all wallets, and a search
  field. Changing filters is local and does not issue another API request.
- Existing create/edit/delete API behavior and owner scope are unchanged in
  this list-filter slice. Dedicated editor routing and visible wallet rows are
  specified separately in
  [NATIVE-CATEGORY-EDITOR-ROUTE](NATIVE-CATEGORY-EDITOR-ROUTE-DETAIL_DESIGN.md).

## Alternatives considered

- Filtering only flat rows: rejected because it loses parent/child context and
  does not match the shared tree used on web.
- Add inactive status in this slice: rejected because it requires a public API
  and data-model decision, and the current API has no reliable inactive source.

## Impact and risks

- Code: native category presentation policy, DesignSystem tree base, category
  management view, XCTest coverage.
- API/data/security/runtime: no change; reads continue through the existing
  session-scoped `/categories` and `/wallets` APIs.
- User-visible behavior: list filtering only; category create/edit/delete and
  transaction/budget behavior are unchanged.
- Main risk: a search result for a child must retain its root context; tests
  cover parent and child matches, empty results, wallet scope and stable order.

## Verification

- RED/GREEN pure XCTest for kind/wallet/search/tree filtering.
- Native category-tree base contract tests and focused category UI test for
  visible filter controls and resulting rows.
- Attach simulator screenshots for the category list and dedicated editor to
  the API-backed UI result bundle, then compare hierarchy, spacing, type and
  safe-area treatment against the matching Stitch captures. This is visual
  evidence only; do not change API/backend behavior without a reproduced
  contract or persistence gap.
- API-backed UI test found that clearing search before editing is not a valid
  visibility proof: `MPCategoryTree` uses `LazyVStack`, so a newly created row
  outside the viewport is not materialized in the accessibility tree until it
  is searched for or scrolled into view. The search assertion itself passed;
  keep the row filtered through edit/delete instead of treating lazy
  materialization as a product failure.
- Native full unit/contract target, Release warnings-as-errors build,
  project/base-component guards, and `git diff --check`.
- Update account-groups and base-component docs, validation matrix, port audit,
  backlog and changelog to match shipped behavior.

## Approval and deferred decision

The list/tree presentation slice follows the owner's request to port the
supplied Stitch UI and web behavior. At the time of this detail design, archive
semantics were a separate unresolved decision; the later approved archive
specification supersedes that deferral and owns the backend/API behavior.

## Verification record — 2026-10-08

- API-backed category create/update/delete and kind/search/wallet controls:
  PASS, 1/1 on iPhone 16 Simulator, `/tmp/mypocket-category-tree-api-ui2.xcresult`.
- Native unit/contract target: PASS, 134/134, no skips or failures,
  `/tmp/mypocket-category-tree-full-unit.xcresult`.
- Release warnings-as-errors build: PASS at
  `/tmp/mypocket-category-tree-release`.
- `check-base-components.sh`, its test, `verify-project.sh` and
  `git diff --check`: PASS.
- At this checkpoint, no backend or public API doc change was needed for list
  filtering because the owner-scoped category response already included
  `wallet_ids`. This statement does not cover the later archive implementation.

## Visual parity implementation — 2026-10-08

- The first API-backed visual run passed the pre-existing CRUD assertions but
  the new parity assertions failed: web order is kind before add; native had
  list-level pencil/delete actions; row subtitle did not show wallet scope.
- Native `MPCategoryTree` now exposes full-row editor navigation, a trailing
  chevron, and no list-level delete affordance. `CategoriesView` places kind
  before add and derives the row subtitle from the existing `wallet_ids`
  contract. `MPButton` accepts an optional accessibility ID applied to its
  underlying SwiftUI `Button` for stable row targeting.
- RED/GREEN API-backed iPhone 16 UI: 1/1 passes at
  `/tmp/mypocket-category-list-parity-green2.xcresult`; screenshots attached
  under `/tmp/mypocket-category-list-parity-green2-attachments/`.
- Native unit/contract: 138/138, zero skips/failures at
  `/tmp/mypocket-category-list-parity-unit.xcresult`. Release build with
  warnings-as-errors, `verify-project.sh`, both base-component guards/tests,
  and `git diff --check` pass.
- Test isolation: local `mypocket_stage_v1`, Redis DB 15 and loopback API
  8081 only; fixture wallet/category counts are zero after the final run. No
  production data or unrelated port 8080 was used. No backend, public API,
  schema, migration, ERD or public API documentation change was needed.
- Superseded: inactive-category/archive behavior no longer lacks a contract; it
  is implemented per the linked archive specification and verified in the
  validation matrix. Remaining: overall owner visual/device acceptance and the
  independent feedback-bubble overlap regression.
