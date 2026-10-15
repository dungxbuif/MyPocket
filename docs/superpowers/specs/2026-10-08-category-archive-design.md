# Category archive and inactive-category visibility

## Goal and context

Complete the category-management behavior shown in Stitch `img_7793` across
web and native iOS. The reference exposes “Hiển thị nhóm không hoạt động”, but
does not show how a personal category is deactivated; the existing web screen
has no inactive state. The current Category API and model have no persisted
active status.

This design assumes reversible soft archival because the application must
preserve financial history. It does not authorize production migration or
deletion. System categories remain immutable and active.

## User-visible behavior

- Personal categories are active by default.
- A personal category editor offers “Lưu trữ nhóm” / “Khôi phục nhóm”.
- Category-management screens show active categories by default and provide a
  persistent “Hiển thị nhóm không hoạt động” filter. When enabled, archived
  categories appear in the existing hierarchy and can be restored.
- Inactive categories cannot be newly selected for transactions, credit
  entries, budgets, recurring schedules, or AI-entry proposals.
- Inactive category rows may remain as non-selectable hierarchy context when
  an active child is shown. Parent and child active states are independent.
- Existing transactions, report totals, month summaries, budgets and recurring
  schedules are never rewritten by archive. Existing budget/schedule records
  may retain their existing inactive category; recurring execution continues
  using its already-saved category. An existing transaction edit may retain
  its current archived category, but may not switch to a different archived
  category.
- An AI proposal whose category becomes inactive before approval must be
  rejected by the same validation path as other invalid proposal categories;
  approval must not write a transaction.
- Existing hard-delete semantics are unchanged. Archive is a reversible
  alternative, not a replacement for the existing delete action.

## API and data contract

- Add `categories.is_active BOOLEAN NOT NULL DEFAULT TRUE`; migration backfills
  all existing personal and system categories as active.
- Category JSON includes `is_active`. `GET /api/v1/categories` continues to
  return all visible categories so historical screens can resolve names/icons;
  clients filter by status for category management and selection.
- Add owner-scoped `PATCH /api/v1/categories/:id/status` with
  `{ "is_active": false }` to archive and `true` to restore. This is
  idempotent, personal-category-only, and may not change system categories.
- The backend validates active status at every new category assignment
  boundary. Existing stored references are preserved per the behavior rules
  above. Ownership, kind and wallet-applicability checks remain in force.
- New assignment attempts with an inactive category return HTTP 400 with
  stable problem code `category_inactive` and do not persist any mutation.
- No API-key permission is broadened. Existing JWT/API-key route policy for
  category reads/writes must be preserved and tested.

## Architecture and boundaries

- Backend entity/repository owns persistence and owner-scoped status changes;
  use cases enforce category invariants; handlers expose typed status input.
- Web and iOS share the server field but own presentation filtering. Selection
  components receive only active selectable IDs while retaining inactive
  ancestors as read-only hierarchy context where needed.
- Reports, transaction details and finance-query outputs remain historical
  reads and must not filter by `is_active`.
- Existing delete behavior and category wallet applicability do not change.

## Failure handling and security

- Missing/invalid JSON returns the standard 400 Problem response.
- Attempts to archive/restore another owner's or system category do not reveal
  the target and return the existing category-not-found shape.
- Attempting to assign an archived category returns a stable 400 validation
  error; no transaction, budget or schedule is written.
- Status mutations are serialized/owner-scoped and never cascade to children.
- Migration is additive and forward-only under the stage-v1 policy; never run
  it against production as part of local implementation or tests.

## Verification

- Go use-case and handler tests cover defaults, archive/restore, owner/system
  isolation, inactive assignment rejection, transaction edit retention,
  budget/schedule preservation, and AI proposal revalidation.
- PostgreSQL migration/repository tests prove existing rows default active,
  owner scope, status persistence, unchanged history and unchanged configured
  budgets/schedules.
- Web tests cover management filter/restore, selection exclusion, and retaining
  an existing archived category in edit forms.
- iOS tests cover decoding/status filtering, hierarchy context, restore route,
  transaction/budget/recurring/credit/AI category selection behavior.
- Regenerate Swagger/OpenAPI and reconcile public API docs, AI-readable docs,
  category screen docs, migration README, validation matrix, backlog, context
  and changelog.
- Run all Go tests, web design/type/build tests, iOS unit and focused API-backed
  UI tests, Release warnings-as-errors, project/base-component guards and
  `git diff --check`.

## Explicit non-goals

- No production deploy, production migration or production DB access.
- No rewriting or reclassification of historical transactions.
- No cascading archive/restore across parent-child categories.
- No change to public API-key permissions, AI write tools or category deletion.
