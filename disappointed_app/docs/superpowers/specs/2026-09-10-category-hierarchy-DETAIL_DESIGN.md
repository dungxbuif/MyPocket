---
artifact_type: detail_design
id: SO-TIEN-F2-CATEGORIES
status: ready
owner: shared
human_approval: Owner authorized continuous implementation, testing and deployment without further approval on 2026-09-10.
---

# F2 category hierarchy and per-wallet activation

Extend the existing owner-scoped category APIs so the UI can truthfully create/edit an optional parent category and read, then update, activation per wallet. Existing system categories remain protected and archive retains historical transaction references.

`POST /api/v1/categories` accepts optional `parent_id`; parent must belong to owner, be active, and have the same income/expense kind. `PATCH /api/v1/categories/{id}` supports changing name and parent with optimistic version. Reject self/descendant cycles, foreign/archived/mismatched parent and system category mutation using existing safe validation/forbidden envelopes.

`GET /api/v1/wallets/{id}/category-settings` returns every visible category with `active`, sourced from `wallet_category_settings` with the established default. Existing `PUT` remains the mutation. Owner isolation and API-key/cookie parity are required.

Tests: PostgreSQL hierarchy/owner/cycle/archive/reference cases; HTTP CSRF/bearer/validation; mounted UI read/edit/reload. No schema migration is expected: `categories.parent_id` and `wallet_category_settings` already exist. No production action until all checks pass.
