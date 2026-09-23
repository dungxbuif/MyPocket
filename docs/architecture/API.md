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

Extraction policy (2026-09-23): `proposals` are reviewable suggestions, not booked transactions. The model is instructed to merge clear overlapping copies within the submitted batch, retain distinct or uncertain entries, prefill a supplied wallet (best match, otherwise first supplied), and keep incomplete fields with per-row questions. No wallet catalog means blank wallet IDs rather than refusing extraction. Missing dates are defaulted to the current instant in the account timezone and marked with a review question; missing amounts remain zero and must be corrected before approval. Balance-only input may validly return an empty list; uncertain mapping or duplicates alone must not cause an empty result. Deduplication against existing ledger records is not performed. There is no application limit on proposal count; parser, persistence and response contracts accept every recognized draft. Each provider response remains bounded at 32768 tokens and 256 KiB for transport safety; a provider-truncated response fails atomically without partial proposals or automatic retry. JSON API shapes and authorization are unchanged. See [review-first design and verification](../work/bugs/AI-ENTRY-REVIEW-FIRST-DETAIL_DESIGN.md).

The multipart `text` field is sent to the model as `user_instruction` as well as retained in source context for backward-compatible extraction. It may filter recognizable rows, such as “chỉ lấy giao dịch tháng 9”; the model uses source dates when available and leaves uncertain dates for review instead of inventing them. OCR/source text remains untrusted evidence and cannot override application policy. `AI_STORE_USAGE=true` (default) stores normalized provider/model request metadata and the provider's usage object on the owner-scoped process/run; set it to `false` to disable persistence. Usage is never returned in the public process response. AI logs record stage, counts, sizes, latency, status, finish reason and token counts only; prompts, OCR, credentials and financial content are excluded.

Authenticated routes under `/api/v1/ai/entry`: `GET /capabilities`, `POST /process`, `GET /requests/{request_id}`, `PATCH /proposals/{id}`, `POST /proposals/{id}/approve`, and `POST /proposals/{id}/reject`. No session, message, history, or latest-record route is public. `POST /process` accepts one multipart request (`request_id`, `text`, `timezone`, up to 20 `files`); retrying an identical request key returns the stored outcome without repeating provider work. `GET /requests/{request_id}` is read-only recovery. [Stateless contract](../superpowers/specs/2026-09-21-stateless-ai-entry-design.md), [AI-ENTRY-02](../work/tickets/AI-ENTRY-02-BATCH-ATTACHMENTS-DETAIL_DESIGN.md), [ADR-005](../decisions/ADR-005-ai-entry-review.md).

Authenticated `GET /api/v1/transactions/{id}/attachments/{attachmentId}/download` verifies both transaction ownership and an existing approval link, then returns `302` to a five-minute private S3 URL. Approved proposal responses include opaque `attachment_ids`; bucket and object keys remain server-only.

Each file must be JPEG/PNG/PDF, <=5 MiB, and match detected content; invalid files are rejected before provider work with 400. The backend stores originals in private environment-qualified S3, then OCR reads them via short-lived signed URLs. The LLM request contains only user text, OCR text and minimal reference catalogs, never file bytes, base64, URL, or image input. Metadata and OCR text are persisted; proposal approval atomically links process attachments to the approved transaction. Authenticated `GET /api/v1/transactions/{id}/attachments/{attachmentId}/download` redirects to a short-lived private URL only after owner/link verification. Cleanup of expired unlinked files remains pending. Persistent proposals have draft/version/status/questions; editing does not write ledger. Approve validates owner, category visibility/applicability, supported wallet, positive safe-integer VND and RFC3339 date; writes exactly one income/expense transaction and approval receipt atomically. Repeated approval returns the original receipt even if the ledger row was later deleted. Transfer/credit approval is unsupported. Reject has no ledger effect. See [AI-ENTRY-02](../work/tickets/AI-ENTRY-02-BATCH-ATTACHMENTS-DETAIL_DESIGN.md) and [ADR-006](../decisions/ADR-006-private-ai-attachments.md).

Same request ID/payload does not repeat extraction. Different payload with same ID, stale edits and conflicting terminal decisions return 409; invalid input 400, unavailable/unconfigured provider 503, wrong owner 404, missing auth 401. A completed AI entry response includes the model's bounded user-facing `reply` alongside `proposals`; this is persisted for request recovery and does not contain provider diagnostics. There is currently no per-user AI/OCR usage cap; provider-cost controls are a future provider-selection concern. Expired processing lease shows interruption; no automatic resubmission. Environment capability booleans indicate configuration, not successful provider authentication or quality.

## Budget and savings integration — 2026-09-20

Authenticated `GET/POST /api/v1/budgets`, `PATCH/DELETE /api/v1/budgets/{id}` implemented. Explicit calendar interval input uses `start_date` and `end_date` (`YYYY-MM-DD`), inclusive in the account's saved timezone; it is not an RFC3339 instant. Other inputs are name, positive safe-integer `limit_amount`, nullable owner `wallet_id` and visible expense `category_id`. List returns items with ledger-derived spent/days_remaining/ended and active limit_amount/spent summary (transaction IDs deduplicated). Category scope includes descendants. Only report-included expenses count. Exact same scope and overlapping dates returns 409; ended budgets cannot be edited. Delete configuration leaves transactions unchanged. [Design/proof](../work/tickets/API-SCREENS-01-DETAIL_DESIGN.md), [ADR-004](../decisions/ADR-004-budget-api-data.md), [ADR-008](../decisions/ADR-008-account-timezone-and-calendar-dates.md).

Goal transaction categories, when explicitly selected, must use real catalog keys income_transfer_in/income_interest for income or expense_transfer_out for expense, in addition to kind/wallet applicability. Category omission remains allowed. External savings entries affect one wallet and retain report inclusion default true. Internal transfers use the dedicated paired endpoint below.

`POST /api/v1/transactions/transfer` atomically creates two owner-scoped rows: an `expense_transfer_out` row on the source wallet and an `income_transfer_in` row on the destination wallet. The wallets must be different, owned by the caller, and non-credit. The server chooses the system categories; both rows share `transfer_id`, have no jar, and set `included_in_reports=false`. Pair edit/delete and idempotency-key replay are follow-up work.

2026-09-21 ownership correction: system category wallet applicability is read/replaced within the authenticated owner's wallets. Changing a shared category's selection preserves other owners' assignments. AI catalog and confirmation use the same scope.

## Account timezone and calendar dates — Stage v1 baseline

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
| MyPocket local API | HTTP REST | JWT bearer for `/auth/profile` and `/home`; OAuth cookies for Google callback; user `mpk_...` API keys for the read-only advisor surface | implemented for current auth/profile/home/advisor slice | Code-first Swagger annotations in Go handlers; generated UI is added incrementally. API-key secrets are shown once and stored only as hashes. |
| `/api/v1/api-keys` | HTTP REST | JWT bearer only | implemented locally | `POST` creates an owner-scoped key and returns the secret once; `GET` lists metadata; `DELETE /:id` revokes. Supported scopes are `finance:read`, `advisor:read`, and `advisor:chat`; `advisor:chat` implies the two read scopes. |
| `/api/v1/categories` | HTTP REST | JWT bearer | implemented | `GET` returns the owner-approved system catalog and owner-visible groups with `wallet_ids`; `POST`/`PATCH` accept name, kind, parent and owner-scoped applicable wallet IDs; `DELETE` removes a personal group and returns its personal children to root. |
| `/api/v1/categories/:id/wallets` | HTTP REST | JWT bearer | implemented | Replaces applicable-wallet selections for a visible personal or system category. System metadata remains immutable. |
| `/api/v1/wallets` | HTTP REST | JWT bearer | implemented; core UAT verified | Create/list/update/delete owner-scoped VND wallets. List returns `opening_balance` and ledger-derived `current_balance`; goal requires positive target and credit requires positive limit. |
| `/api/v1/transactions` | HTTP REST | JWT bearer | implemented; basic ledger UAT verified | Create/list/update/delete owner-scoped income and expense entries for basic/goal wallets. Optional category must match kind and applicable-wallet scope; credit uses a later dedicated ledger. |
| `/api/v1/transactions/transfer` | HTTP REST | JWT bearer | implemented; automated proof, owner UAT pending | Atomically create paired source-expense/destination-income rows for an internal transfer. Transfers are excluded from reports and jars. |
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
## Feedback and changelog

The owner-scoped feedback lifecycle and the redacted service-token agent surface are documented in [FEEDBACK_API.md](FEEDBACK_API.md). The public changelog does not expose raw feedback. This API is separate from user API-key authentication and Finance Assistant credential scopes; the Feedback service token is never accepted by advisor routes.

## Finance query foundation — local implementation

The shared backend finance query layer now provides owner-scoped, repeatable-read summaries, transaction search with all-scope counts and keyset cursors, category subtree filters, period comparison with zero-baseline semantics, wallet/goal/budget views, and a SELECT-only jar progress view. It is the internal service boundary used by the local advisor surface. Transfer rows and system transfer categories are excluded from report summaries, while search can inspect the full ledger scope.

The first local Finance Assistant surface is mounted behind JWT/session or a user API key: `GET /api/v1/ai/advisor/capabilities`, `GET /api/v1/ai/advisor/overview`, `POST /api/v1/ai/advisor/messages`, `GET /api/v1/ai/advisor/conversation/messages`, `GET /api/v1/ai/advisor/runs/{id}`, and `POST /api/v1/ai/advisor/runs/{id}/cancel`. Overview reads real report data without calling the model; chat persists one owner conversation/run and executes only the eight allowlisted read tools. Advisor auth is audited best-effort to Redis stream `mypocket:audit:advisor`; durable delivery/alerting and replayable SSE remain later gates from the advisor plan.
