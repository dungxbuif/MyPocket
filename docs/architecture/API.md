---
artifact_type: api_contract_index
id: API-MASTER
status: draft
owner: shared
human_fields: [contract_intent, approval]
ai_fields: [contract_rows, errors, auth_notes, versioning, linked_decisions]
shared_fields: [status, trace]
---

# API

## AI entry review — 2026-09-21

Authenticated routes under `/api/v1/ai/entry`: `GET /capabilities`, `POST /process`, `GET /requests/{request_id}`, `PATCH /proposals/{id}`, `POST /proposals/{id}/approve`, and `POST /proposals/{id}/reject`. No session, message, history, or latest-record route is public. `POST /process` accepts one multipart request (`request_id`, `text`, `timezone`, up to three `files`); retrying an identical request key returns the stored outcome without repeating provider work. `GET /requests/{request_id}` is read-only recovery. [Stateless contract](../superpowers/specs/2026-09-21-stateless-ai-entry-design.md), [AI-ENTRY-02](../work/tickets/AI-ENTRY-02-BATCH-ATTACHMENTS-DETAIL_DESIGN.md), [ADR-005](../decisions/ADR-005-ai-entry-review.md).

Authenticated `GET /api/v1/transactions/{id}/attachments/{attachmentId}/download` verifies both transaction ownership and an existing approval link, then returns `302` to a five-minute private S3 URL. Approved proposal responses include opaque `attachment_ids`; bucket and object keys remain server-only.

Each file must be JPEG/PNG/PDF, <=5 MiB, and match detected content; invalid files are rejected before provider work with 400. The backend stores originals in private environment-qualified S3, then OCR reads them via short-lived signed URLs. The LLM request contains only user text, OCR text and minimal reference catalogs, never file bytes, base64, URL, or image input. Metadata and OCR text are persisted; proposal approval atomically links process attachments to the approved transaction. Authenticated `GET /api/v1/transactions/{id}/attachments/{attachmentId}/download` redirects to a short-lived private URL only after owner/link verification. Cleanup of expired unlinked files remains pending. Persistent proposals have draft/version/status/questions; editing does not write ledger. Approve validates owner, category visibility/applicability, supported wallet, positive safe-integer VND and RFC3339 date; writes exactly one income/expense transaction and approval receipt atomically. Repeated approval returns the original receipt even if the ledger row was later deleted. Transfer/credit approval is unsupported. Reject has no ledger effect. See [AI-ENTRY-02](../work/tickets/AI-ENTRY-02-BATCH-ATTACHMENTS-DETAIL_DESIGN.md) and [ADR-006](../decisions/ADR-006-private-ai-attachments.md).

Same request ID/payload does not repeat extraction. Different payload with same ID, stale edits and conflicting terminal decisions return 409; invalid input 400, unavailable/unconfigured provider 503, wrong owner 404, missing auth 401. Account cap 20 extraction attempts/24 hours returns 429. Expired processing lease shows interruption; no automatic resubmission. Environment capability booleans indicate configuration, not successful provider authentication or quality.

## Budget and savings integration — 2026-09-20

Authenticated `GET/POST /api/v1/budgets`, `PATCH/DELETE /api/v1/budgets/{id}` implemented. Explicit calendar interval input uses `start_date` and `end_date` (`YYYY-MM-DD`), inclusive in the account's saved timezone; it is not an RFC3339 instant. Other inputs are name, positive safe-integer `limit_amount`, nullable owner `wallet_id` and visible expense `category_id`. List returns items with ledger-derived spent/days_remaining/ended and active limit_amount/spent summary (transaction IDs deduplicated). Category scope includes descendants. Only report-included expenses count. Exact same scope and overlapping dates returns 409; ended budgets cannot be edited. Delete configuration leaves transactions unchanged. [Design/proof](../work/tickets/API-SCREENS-01-DETAIL_DESIGN.md), [ADR-004](../decisions/ADR-004-budget-api-data.md), [ADR-008](../decisions/ADR-008-account-timezone-and-calendar-dates.md).

Goal transaction categories, when explicitly selected, must use real catalog keys income_transfer_in/income_interest for income or expense_transfer_out for expense, in addition to kind/wallet applicability. Category omission remains allowed. External savings entries affect one wallet and retain report inclusion default true; internal paired transfers are not implemented by this path.

2026-09-21 ownership correction: system category wallet applicability is read/replaced within the authenticated owner's wallets. Changing a shared category's selection preserves other owners' assignments. AI catalog and confirmation use the same scope.

## Account timezone and calendar dates — migrations 000013–000015

`GET/PATCH /api/v1/auth/profile` reads or updates the authenticated account's IANA `timezone`; invalid zone names return 400. The new-account browser-zone initialization is conditional and does not silently change an existing confirmed setting. Timestamp fields such as `transactions.occurred_at` remain RFC3339 instants; date labels are `YYYY-MM-DD` and month keys are `YYYY-MM`. Month/report intervals are derived from the account timezone as half-open UTC instants.

Wallet create/update accepts optional `target_date` as `YYYY-MM-DD` for goal wallets. PostgreSQL stores it as `date`; omission preserves the date and an empty string clears it. Budget `start_date`/`end_date` are also date-only calendar labels interpreted in the account timezone. [ADR-008](../decisions/ADR-008-account-timezone-and-calendar-dates.md) records the migration and timezone rules.

`POST /api/v1/transactions` and transaction edits accept optional `jar_id` (null clears it). A jar is valid only for an ordinary expense and an active configuration for that owner's transaction month. Income and transfer-out categories cannot be assigned.

Authenticated jar routes: `GET/POST /api/v1/jars`, `PUT/DELETE /api/v1/jars/{jar_id}/months/{month}`, and `GET /api/v1/jars/{jar_id}/report?from=YYYY-MM&to=YYYY-MM`. Month routes: `GET /api/v1/months/{month}`, `PUT /api/v1/months/{month}/note`, and `DELETE /api/v1/months/{month}/note`. Month figures are live ledger calculations; completion is derived from the current account-local month, and the note is independently user-authored. There is no scheduled or immutable close/report snapshot in CORE-03.

## Field Ownership

- Human owns public contract intent and approval for contract changes.
- AI maintains contract rows, errors, auth notes, versioning, and linked decisions.

## API Surface

The current local Gin API will use a code-first Swagger contract: each Gin handler owns its Swagger annotations, and generated artifacts are derived from the Go source. Keep the generated surface limited to endpoints that exist in the running implementation; add annotations only when the corresponding handler and verification exist.

Swagger UI is available at `/api/v1/docs/index.html`; generated artifacts live under `backend/docs/`. Run `go generate ./cmd/api` from `backend/`: the generator declaration stays beside the Go API entry point, and endpoint annotations stay beside their handlers. There is no separate Swagger configuration file.

Frontend navigation is client-side and does not change the API base path. TanStack Router exposes `/`, `/transactions`, `/budgets`, `/account`, `/account/groups`, `/account/wallets`, `/jars`, and `/months/{YYYY-MM}`. The global quick-add sheet persists the implemented basic income/expense contract with optional jar assignment. General reports remain unmounted until their broader APIs exist.

Google login starts from the frontend origin using the relative `/api/v1/auth/google` path. Vite proxies that request during development, and the backend redirects the completed OAuth flow back to the same frontend origin at `/auth/callback`.

Document HTTP endpoints, RPC methods, events, CLI commands, or any other public contract.

| Contract | Type | Auth | Status | Notes |
| --- | --- | --- | --- | --- |
| OCR Platform document recognition | HTTP REST | `Authorization: Bearer sk_ocr_...` for protected requests | ready | See `docs/architecture/OCR_API.md`. |
| MyPocket local API | HTTP REST | JWT bearer for `/auth/profile` and `/home`; OAuth cookies for Google callback | implemented for current auth/profile/home slice | Code-first Swagger annotations in Go handlers; generated UI is added incrementally. |
| `/api/v1/categories` | HTTP REST | JWT bearer | implemented | `GET` returns the owner-approved system catalog and owner-visible groups with `wallet_ids`; `POST`/`PATCH` accept name, kind, parent and owner-scoped applicable wallet IDs; `DELETE` removes a personal group and returns its personal children to root. |
| `/api/v1/categories/:id/wallets` | HTTP REST | JWT bearer | implemented | Replaces applicable-wallet selections for a visible personal or system category. System metadata remains immutable. |
| `/api/v1/wallets` | HTTP REST | JWT bearer | implemented; core UAT verified | Create/list/update/delete owner-scoped VND wallets. List returns `opening_balance` and ledger-derived `current_balance`; goal requires positive target and credit requires positive limit. |
| `/api/v1/transactions` | HTTP REST | JWT bearer | implemented; basic ledger UAT verified | Create/list/update/delete owner-scoped income and expense entries for basic/goal wallets. Optional category must match kind and applicable-wallet scope; credit uses a later dedicated ledger. |
| `/api/v1/auth/profile` | HTTP REST | JWT bearer | implemented | Reads/updates account IANA timezone; initial browser-zone adoption is initialize-only. |
| `/api/v1/jars` and `/api/v1/jars/{id}/...` | HTTP REST | JWT bearer | implemented; owner UAT pending | Stable jar IDs, month snapshots, optional ordinary-expense assignment and cumulative live report. Config removal does not delete transactions or historic snapshots. |
| `/api/v1/months/{YYYY-MM}` | HTTP REST | JWT bearer | implemented; owner UAT pending | Live account-local totals and independent note CRUD; completion is derived, not a close operation. |

## Errors

All MyPocket success responses use `{data, meta}`. Errors use RFC 9457-style Problem Details with `code` and `request_id`; handlers use shared helpers in `backend/internal/controller/http/response.go`.

| Error | Meaning | Consumer Impact |
| --- | --- | --- |
| `400 INVALID_INPUT` | OCR request/options invalid. | Show OCR failure and keep manual receipt entry available. |
| `404 NOT_FOUND` | OCR document ID unknown, expired metadata, or wrong owner. | Mark OCR job unavailable; retain attachment/manual entry. |
| `410 RESULT_EXPIRED` | OCR result TTL expired. | Use persisted MyPocket result if available or request re-scan. |
| `413 URL_CONTENT_TOO_LARGE` | File exceeds OCR deployment limit. | Show max upload size and ask user to reduce/split file. |
| `415 UNSUPPORTED_MEDIA_TYPE` | Unsupported input type. | Block unsupported upload in UI where possible. |

## Versioning

External public contracts use provider versioning where available. OCR Platform uses `/v1` REST paths. MyPocket should wrap provider calls behind an internal adapter so provider changes do not leak into frontend components.

## Linked Decisions

- `docs/architecture/OCR_API.md`
