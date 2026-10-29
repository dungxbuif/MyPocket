# Native category icon-picker parity

## Decision

Port the web category icon picker as a shared native base control instead of
adding a screen-local menu. The picker exposes the same ten canonical keys
used by `CategoryIconPicker` on web, keeps the selected icon in the category
draft, and sends it through the existing `CategoryModelInput.iconKey` field.

## Scope

- Add a reusable SwiftUI icon-picker base with a compact selected preview and
  a grid sheet for canonical category icons.
- Compose it in the Categories create/edit form for editable categories.
- Preserve the read-only system-category path and the existing owner-scoped
  category API contract.

## Verification

- Add a pure key/label/presentation contract test.
- Run native unit tests, Release warnings-as-errors, project/base-component
  guards and the web design guard.
- Keep the API-backed category UI proof as the behavior gate; no production
  data is touched.

## Result

- Added `MPCategoryIconSelector` to the native DesignSystem and wired category
  create/edit state to the selected icon key.
- Full native unit target passes 119/119; Release warnings-as-errors passes.
- No API/schema change; system-category metadata remains read-only.
- Local API-backed category create/update/delete UI flow passes 1/1 with
  `expense_food` persisted as the selected `icon_key`.
