---
name: mypocket-api
description: Use MyPocket's public API to manage an authenticated user's personal-finance data and reviewable AI drafts.
---

# MyPocket API

Use this skill only for data belonging to the user represented by the supplied MyPocket API key.

## Start safely

1. Fetch `GET /api/v1/openapi.json` from the target deployment as the authoritative route inventory.
2. Obtain a user-created key from the Account screen. Never put it in logs, source, documentation, screenshots, or retained chat.
3. Send `Authorization: Bearer <user-api-key>`. Bearer calls do not send CSRF headers.
4. Read the current entity and its `version` before update/archive operations. Send the documented `base_version` unchanged.
5. Generate one stable `Idempotency-Key` for each logical mutation and reuse it only when retrying the exact same body.

## Operating rules

- Never send or trust a payload `user_id`; ownership comes from the authenticated principal.
- Prefer review-first flows: create or inspect drafts, then explicitly confirm or reject.
- Treat transfers as balance movement, not income/expense. Read report endpoints instead of calculating authoritative totals from pages.
- Paginate notifications with `before`, sync changes with `after`, and preserve server cursors.
- On sync retry, reuse the same mutation ID and payload. Resolve `409` from server state instead of overwriting blindly.
- Do not automatically confirm imports, destructive operations, OCR-derived values, or uncertain drafts.

## Error handling

- `401`: key missing, invalid, revoked, or account disabled. Stop and request a fresh user-authorized key.
- `403`: special permission or browser-only recent-auth requirement. Do not retry with broader access.
- `404`: resource absent; user-owned APIs also hide foreign ownership this way.
- `409`: refetch, compare versions, and request review when intent conflicts.
- `429`: wait for `Retry-After`, then retry the identical request.
- `5xx`: retry only boundedly with jitter while preserving idempotency keys and mutation IDs.

Read [references/api-quick-reference.md](references/api-quick-reference.md) for common calls and lifecycle boundaries.
