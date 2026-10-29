---
artifact_type: bug
id: NATIVE-SYSTEM-CATEGORY-SCOPE
status: regression_verified
created: 2026-10-06
updated: 2026-10-06
---

# Native system-category edit scope

## Finding

The native category tree exposed the same edit and delete controls for system
categories as for personal categories. The backend contract intentionally
rejects the personal-category update/delete routes for system rows; only the
applicable-wallet relationship may be changed through
`PATCH /api/v1/categories/{id}/wallets`.

## Intended behavior

- System categories remain read-only for name, kind, parent, and icon.
- System categories may update only their applicable wallet IDs.
- Personal categories keep the normal edit and delete actions.

## Implementation and verification result

- Native system rows now show a lock state instead of delete, render metadata
  as read-only, and save only wallet applicability through the dedicated
  `/categories/{id}/wallets` endpoint.
- Personal categories retain create/edit/delete behavior and the native kind
  selector now includes `debt`, matching the web contract.
- Native unit/contract tests pass 49/49; the signed simulator route sweep
  passes 1/1 for all 13 More destinations.

## Verification plan

1. ~~Add the wallet-scope API method and a small policy test for system rows.~~
2. ~~Route the native editor to the wallet-only endpoint for system categories;
   hide delete and disable metadata editing.~~
3. ~~Run native unit tests, signed API-backed iPhone Simulator UI tests, and
   reconcile this note, validation matrix, context, and changelog.~~

No production data or provider credentials are used for this fix.
