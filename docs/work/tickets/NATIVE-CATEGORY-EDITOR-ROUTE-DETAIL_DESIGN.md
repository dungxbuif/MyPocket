---
artifact_type: detail_design
id: NATIVE-CATEGORY-EDITOR-ROUTE
status: implemented
owner: shared
trace:
  backlog: ../BACKLOG.md
  audit: ../PORT-100-AUDIT.md
  categories: ../../design/pages/account-groups/README.md
  edit_group: ../../design/molecules/edit-group/README.md
  validation: ../VALIDATION_MATRIX.md
---

# Native category editor route parity — detail design

## Intent

Match the web/Stitch category-management flow: the category list is a distinct
screen, and create/edit opens a dedicated editor destination rather than
inserting a form above the long category tree. Render wallet applicability as
visible multi-select rows rather than a menu. Preserve the existing category
CRUD API, owner scope, category parent rules and empty-`wallet_ids` means
“applies to all wallets” behavior.

## Interaction and base components

- “Nhóm mới” opens a dedicated native editor screen with its own shared header,
  back action, metadata form and footer actions. Selecting a category row opens
  that same editor for the category. Save/cancel returns to the category list;
  save refreshes the owner-scoped list.
- Personal categories expose delete from the editor header; deletion keeps the
  existing confirmation and backend integrity behavior. System categories
  retain read-only name/kind/parent/icon metadata and can only change wallet
  applicability.
- The wallet applicability card uses a shared DesignSystem checkbox row. Empty
  selection represents “Tất cả ví” (every wallet checked). Tapping one wallet
  from the all-wallet state creates an explicit set containing the remaining
  wallets; clearing the last explicit wallet returns to the all-wallet state.
- Kind changes clear an incompatible parent. Parent choices remain same-kind
  roots and exclude the edited node/descendants. Existing icon picker, trimmed
  name validation, save loading and retryable errors remain.
- Use project-owned header, text, field, surface, button, status, icon and new
  checkbox-row bases. Do not introduce raw feature-level `Text`, `Image`,
  button or platform form primitives.

## Backend/API decision

The public API contract does not need to change: `/categories` already accepts
owner-scoped `wallet_ids`, system-category scope has a dedicated update route,
and deletion remains server-authorized. API-backed testing found a repository
implementation gap: category creation stores the category row but discards its
wallet IDs. The backend must persist category creation and its wallet links
atomically, after validating every wallet belongs to the authenticated owner.
Keep the existing wire contract and add a repository regression proving the
selected wallet scope survives a fresh list read. Never move owner scope or
validation into the client.

## Verification plan

- First add failing API-backed iPhone assertions for dedicated create/edit
  screen routing and visible wallet checkbox rows.
- Then verify backend category creation persists selected wallet links; verify
  create with a selected wallet and update/delete through the dedicated route,
  including fresh API reads of `wallet_ids` and category name.
- Run the focused category UI test, native unit/contract suite, backend tests
  against isolated local `mypocket_stage_v1`, Release warnings-as-errors,
  project/base-component guards and `git diff --check`.
- Reconcile category page/component design docs, phase, backlog, port audit,
  validation and changelog. Production data and APIs are out of scope.

## Implementation and evidence — 2026-10-08

- The native route and visible wallet rows are implemented using shared
  DesignSystem bases.
- API-backed RED found `walletIDs` was serialized as `wallet_i_ds`; the Go API
  expected `wallet_ids`. The request model now pins the exact field name and a
  wire-level test prevents regression.
- Category create now writes the category and its owner-validated wallet links
  within one backend transaction. The public API contract remains unchanged.
- The iPhone 16 UI test verifies create, visible selection, server readback of
  selected/excluded wallets, filters, rename and delete. Owner visual/device
  acceptance and inactive-category semantics remain outside this completed
  functional slice.
