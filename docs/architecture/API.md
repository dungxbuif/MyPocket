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

## Offline implementation status — 2026-10-15

Offline is not available on iOS yet and the native client still fails closed
until it sees an authenticated capability response. The server contract is now
deployed in API/worker image `release-20261016-7846945` at digest
`sha256:71fdd4b4155972f3e6380ae02de846264e503347ab63b9693f65d2f0c713db43`.
Local HTTP/DB
tests now verify create requests with Idempotency-Key: first201, replay200, unchanged
data object/array plus meta.mutation (operation/key/status/replayed). Typed JSON
is hashed before generated defaults; completed receipt lookup precedes mutable
wallet/category validation. Invalid explicit keys fail400; changed body409
IDEMPOTENCY_KEY_REUSED; missing durable capability fails closed, not a plain write.
No-header requests retain legacy online behavior/meta:null and are unsafe to
automatically retry. Real router/PostgreSQL/Redis verification confirms JWT/session
owner binding and expiry/logout rejection before replay. Per owner's2026-10-14
decision, all user routes now also accept full-owner API keys, including financial
writes/replay; no per-key scopes. Revocation/expiry/owner isolation remain enforced.
The server route is deployed and production returns it behind auth; required-key
rollout and native cache/sync/UI remain gates; encrypted outbox storage/leases
are implemented locally. An owner-authenticated production capability response
is still required before enabling the native sender.
[Human/agent guide](../guides/offline-sync-api.md), [auth proof](../work/tickets/OFFLINE-AUTH-HTTP-20261014.md).

GET /api/v1/sync/capabilities is locally implemented under JWT/session or owner key
auth: contract_version1,enabled,operations,receipt_metadata=/meta/mutation,
idempotency_key_required_for_offline=true,legacy_unkeyed_online=true. Operations
are transaction.create/transfer.create only when durable repository+transfer
writer configured; otherwise enabled=false/operations=[]. Read-only, no DB health
claim. Old404/unsupported version/policy blocks future
outbox sends; revalidate each drain, not stale cache after server rollback.
Native getter/policy tests pass; no sync engine/UI wiring or prod activation yet.
[Capability work](../work/tickets/OFFLINE-SERVER-CAPABILITY-20261014.md).

2026-10-14 frozen-create addition (local, not deployed): POST /transactions uses
create-only auto_link_travel boolean. Omitted/null/true retains active-trip fallback;
false uses only validated explicit travel_event_id or no trip. New DTO embeds the
old update DTO and omits nil policy from canonical receipt hash; pre-change
receipts replay unchanged. A changed explicit policy conflicts under the same key.
Capability transaction_create_context="explicit_travel_v1" is advertised only
with configured durable writer; iOS ordinary-create sends require this exact value.
Transfers have no active-trip fallback. Outbox freezes false before persistence,
retaining supplied trip/time. No sync/UI or production availability claim.
[Proof](../work/tickets/OFFLINE-FROZEN-TRAVEL-20261014.md).

Full-owner API keys are implemented locally through common RequireAuth, without
scope checks in authentication/advisor/feedback. [Design/proof](../work/tickets/API-KEY-FULL-ACCESS-20261014.md)
and [human/agent guide](../guides/api-keys.md). Existing valid keys gain the same
owner's capabilities when rolled out; no key regeneration or SQL migration.
[Work/evidence](../work/tickets/OFFLINE-CREATE-RECEIPTS-20261014.md).

Shared SQL diagnostic redaction is locally verified (not deployed): statement
class/timing/rows/SQLSTATE only, no query text/parameters/raw driver error values.
The logging change does not alter responses/errors/auth. This does not establish safety of every
application log call. [Logging contract](../decisions/ADR-019-financial-sql-log-redaction.md).

## AI entry review — 2026-09-21

Extraction policy (2026-09-23): `proposals` are reviewable suggestions, not booked transactions. The model is instructed to merge clear overlapping copies within the submitted batch, retain distinct or uncertain entries, prefill a supplied wallet (best match, otherwise first supplied), and keep incomplete fields with per-row questions. No wallet catalog means blank wallet IDs rather than refusing extraction. Missing dates are defaulted to the current instant in the account timezone and marked with a review question; missing amounts remain zero and must be corrected before approval. Balance-only input may validly return an empty list; uncertain mapping or duplicates alone must not cause an empty result. Deduplication against existing ledger records is not performed. There is no application limit on proposal count; parser, persistence and response contracts accept every recognized draft. Each provider response remains bounded at 32768 tokens and 256 KiB for transport safety; a provider-truncated response fails atomically without partial proposals or automatic retry. JSON API shapes and authorization are unchanged. See [review-first design and verification](../work/bugs/AI-ENTRY-REVIEW-FIRST-DETAIL_DESIGN.md).

The multipart `text` field is sent to the model as `user_instruction` as well as retained in source context for backward-compatible extraction. It may filter recognizable rows, such as “chỉ lấy giao dịch tháng 9”; the model uses source dates when available and leaves uncertain dates for review instead of inventing them. OCR/source text remains untrusted evidence and cannot override application policy. `AI_STORE_USAGE=true` (default) stores normalized provider/model request metadata and the provider's usage object on the owner-scoped process/run; set it to `false` to disable persistence. Usage is never returned in the public process response. AI logs record stage, counts, sizes, latency, status, finish reason and token counts only; prompts, OCR, credentials and financial content are excluded.

Authenticated routes under `/api/v1/ai/entry`: `GET /capabilities`, `POST /process`, `GET /requests/{request_id}`, `PATCH /proposals/{id}`, `POST /proposals/{id}/approve`, and `POST /proposals/{id}/reject`. No session, message, history, or latest-record route is public. `POST /process` accepts one multipart request (`request_id`, `text`, `timezone`, up to 20 `files`); retrying an identical request key returns the stored outcome without repeating provider work. `GET /requests/{request_id}` is read-only recovery. Client retry rule: after a transport/decoding failure or server error where POST acceptance is uncertain, keep the original `request_id` and fetch `GET /requests/{request_id}`; do not generate a new ID or automatically POST again. If GET finds the request, continue with that result (poll GET while `processing=true`). A confirmed GET 404 means that request ID was not accepted/found and allows the client to offer an explicit retry using the retained draft. [Stateless contract](../superpowers/specs/2026-09-21-stateless-ai-entry-design.md), [AI-ENTRY-02](../work/tickets/AI-ENTRY-02-BATCH-ATTACHMENTS-DETAIL_DESIGN.md), [ADR-005](../decisions/ADR-005-ai-entry-review.md).

Authenticated `GET /api/v1/transactions/{id}/attachments/{attachmentId}/download` verifies both transaction ownership and an existing approval link, then returns `302` to a five-minute private S3 URL. Approved proposal responses include opaque `attachment_ids`; bucket and object keys remain server-only.

Each file must be JPEG/PNG/PDF, <=5 MiB, and match detected content; invalid files are rejected before provider work with 400. The backend stores originals in private environment-qualified S3. Validated images are submitted to OCR as Base64; PDFs use the OCR provider's private presign flow and provider-owned `s3://` source URL. The LLM request contains only user text, OCR text and minimal reference catalogs, never file bytes, Base64, URL or image input. Metadata and OCR text are persisted; proposal approval atomically links process attachments to the approved transaction. Authenticated `GET /api/v1/transactions/{id}/attachments/{attachmentId}/download` redirects to a short-lived private URL only after owner/link verification. Cleanup of expired unlinked files remains pending. Persistent proposals have draft/version/status/questions; editing does not write ledger. Approve validates owner, category visibility/applicability, supported wallet, positive safe-integer VND and RFC3339 date; writes exactly one income/expense transaction and approval receipt atomically. Repeated approval returns the original receipt even if the ledger row was later deleted. Transfer/credit approval is unsupported. Reject has no ledger effect. See [AI-ENTRY-02](../work/tickets/AI-ENTRY-02-BATCH-ATTACHMENTS-DETAIL_DESIGN.md) and [ADR-006](../decisions/ADR-006-private-ai-attachments.md).

Same request ID/payload does not repeat extraction. Different payload with same ID, stale edits and conflicting terminal decisions return 409; invalid input 400, unavailable/unconfigured provider 503, wrong owner 404, missing auth 401. A completed AI entry response includes the model's bounded user-facing `reply` alongside `proposals`; this is persisted for request recovery and does not contain provider diagnostics. There is currently no per-user AI/OCR usage cap; provider-cost controls are a future provider-selection concern. Expired processing lease shows interruption; no automatic resubmission. Environment capability booleans indicate configuration, not successful provider authentication or quality.

## Budget and savings integration — 2026-09-20

Authenticated `GET/POST /api/v1/budgets`, `PATCH/DELETE /api/v1/budgets/{id}` implemented. Explicit calendar interval input uses `start_date` and `end_date` (`YYYY-MM-DD`), inclusive in the account's saved timezone; it is not an RFC3339 instant. Other inputs are name, positive safe-integer `limit_amount`, nullable owner `wallet_id` and visible expense `category_id`. List returns items with ledger-derived spent/days_remaining/ended and active limit_amount/spent summary (transaction IDs deduplicated). Category scope includes descendants. Only report-included expenses count. Exact same scope and overlapping dates returns 409; ended budgets cannot be edited. Delete configuration leaves transactions unchanged. [Design/proof](../work/tickets/API-SCREENS-01-DETAIL_DESIGN.md), [ADR-004](../decisions/ADR-004-budget-api-data.md), [ADR-008](../decisions/ADR-008-account-timezone-and-calendar-dates.md).

Goal transaction categories must use real catalog keys income_transfer_in/income_interest for income or expense_transfer_out for expense, in addition to kind/wallet applicability. Ordinary income/expense and credit entries require a visible active category; clients prefill the first applicable group when the user has not chosen one. Only adjustment rows are intentionally uncategorized. External savings entries affect one wallet and retain report inclusion default true. Internal transfers use the dedicated paired endpoint below. Category archival is soft and reversible: it never rewrites ledger or report history. The request fails with `400 category_inactive` when a new assignment uses an archived category; an edit may retain the exact category already attached to its existing transaction.

`POST /api/v1/transactions/transfer` atomically creates two owner-scoped rows: an `expense_transfer_out` row on the source wallet and an `income_transfer_in` row on the destination wallet. The wallets must be different, owned by the caller, and non-credit. The server chooses the system categories; both rows share `transfer_id`, have no jar, and set `included_in_reports=false`. `PATCH` and `DELETE /api/v1/transactions/transfer/{transfer_id}` mutate the exact pair atomically and require an `Idempotency-Key`; replaying the same key/payload returns the stored result, while a changed payload returns 409.

The transfer `PATCH` body is `{amount, occurred_at, note}`; both linked rows update atomically, while source and destination wallet IDs stay fixed for that transfer ID. A transfer cannot be moved between wallets by this endpoint.

`POST /api/v1/transactions/adjustment` records an explicit balance correction for an owned basic or goal wallet. The body is `{wallet_id, amount, direction, occurred_at?, note?}`, where `amount` is a positive integer and `direction` is `increase` or `decrease`. The server stores `type=adjustment`, always sets `included_in_reports=false`, and excludes the row from income/expense reports while applying its signed effect to wallet balances. Adjustment rows are immutable through ordinary transaction edit/delete routes; create a compensating adjustment instead.

`POST /api/v1/transactions/bulk-delete` accepts `{transaction_ids: string[]}` with 1–100 unique IDs. The owner-scoped repository locks and validates every requested row before deleting any row. If an ID is missing/not owned, a transfer pair is included, or an immutable adjustment is included, the entire request is rejected and no row is deleted. The endpoint returns `204` only after all rows are deleted.

Recurring schedules are exposed at `GET/POST /api/v1/recurring`, `PATCH/DELETE /api/v1/recurring/{id}`, and `POST /api/v1/recurring/run-due`. Schedules are owner-scoped ordinary income/expense templates with a positive amount, frequency (`daily|weekly|monthly|yearly`), interval, next local-time occurrence and optional end. `run-due` is bounded to 100 occurrences and records a unique `(schedule_id, due_at)` marker in the same transaction as the generated row; retries do not duplicate transactions. Pausing/deleting a schedule leaves generated rows intact, and generated notes default to `Giao dịch định kỳ — {name}`.

Credit wallets use a separate debt ledger. `POST /api/v1/credit/wallets/{wallet_id}/entries` records a `purchase`, `refund`, `fee`, or `interest` row; purchases/fees/interest are expenses and refunds are income. A category is required, must match the entry direction and its applicable-wallet scope, and is rejected when restricted to another wallet. `POST /api/v1/credit/wallets/{wallet_id}/payments` atomically records an excluded source-wallet expense paired with an excluded credit-wallet `payment` income row. The source must be a different non-credit wallet. `GET /api/v1/credit/wallets/{wallet_id}/statement` returns the credit limit, signed outstanding balance, available credit (`limit + balance`), configured statement balance/cycle, amount due, payment status (`not_configured|due|partial|paid|overdue`) and credit rows, with optional `from`/`to` `YYYY-MM-DD` filters. Credit wallet create/update accepts optional `last_statement_balance`, `statement_day` and `payment_due_day` (days 1–31). Credit rows cannot be created through ordinary transaction, adjustment, transfer, or bulk-delete routes; this preserves payment pairing and statement semantics. All routes are owner-scoped and accept JWT/session or a full-owner user API key.

Travel Mode is an owner-scoped event dimension, not a fourth wallet type. `GET/POST/PATCH/DELETE /api/v1/travel/events`, `POST /api/v1/travel/events/{id}/activate`, and `POST /api/v1/travel/events/{id}/deactivate` manage events with at most one active event per owner. Ordinary income/expense creation auto-links the active event when no explicit event is supplied unless auto_link_travel=false; recurring, transfer, adjustment and AI approval paths do not auto-link. `PATCH /api/v1/transactions/{id}/travel` links or clears one ordinary transaction. Deleting an event clears links without deleting transactions or changing report inclusion. See [Travel Mode page contract](../design/pages/travel/README.md) and [implementation plan](../superpowers/plans/2026-10-01-travel-mode.md).

2026-09-21 ownership correction: system category wallet applicability is read/replaced within the authenticated owner's wallets. Changing a shared category's selection preserves other owners' assignments. AI catalog and confirmation use the same scope.

## Account timezone and calendar dates — Stage v1 baseline

`GET/PATCH /api/v1/auth/profile` reads or updates the authenticated account's IANA `timezone`; invalid zone names return 400. The new-account browser-zone initialization is conditional and does not silently change an existing confirmed setting. Timestamp fields such as `transactions.occurred_at` remain RFC3339 instants; date labels are `YYYY-MM-DD` and month keys are `YYYY-MM`. Month/report intervals are derived from the account timezone as half-open UTC instants.

Wallet create/update accepts optional `target_date` as `YYYY-MM-DD` for goal wallets. PostgreSQL stores it as `date`; omission preserves the date and an empty string clears it. Budget `start_date`/`end_date` are also date-only calendar labels interpreted in the account timezone. [ADR-008](../decisions/ADR-008-account-timezone-and-calendar-dates.md) records the migration and timezone rules.

Goal-wallet create/update also accepts `goal_notifications_enabled` (default
`false`). The field is accepted only for goal wallets and is owner-scoped. The
current delivery surface is the in-app inbox: `GET /api/v1/notifications`
returns the authenticated owner's newest-first items and
`POST /api/v1/notifications/{id}/read` is an idempotent owner-scoped operation.
Push/APNs and scheduled notification fan-out are not claimed yet.

`POST /api/v1/transactions` and transaction edits require an applicable category for ordinary income/expense rows and accept optional `jar_id` (null clears it). `included_in_reports` defaults to true and the create route persists an explicit false value; false keeps the transaction in the ledger and wallet balance but excludes it from report totals and trend aggregation. A jar is valid only for an ordinary expense and an active configuration for that owner's transaction month. Income and transfer-out categories cannot be assigned. Adjustment rows remain the only uncategorized transaction type.

Portfolio routes are owner-scoped and intentionally separate from wallet ledger balances: `GET /api/v1/portfolio/summary`, `GET/POST /api/v1/portfolio/assets`, `PATCH /api/v1/portfolio/assets/{id}/price`, and `GET/POST /api/v1/portfolio/assets/{id}/trades`. Assets hold an optional manual latest-price snapshot. Trades are immutable `buy`/`sell` events with decimal-string quantity (up to eight fractional places), integer VND unit price and fee. The server stores fixed-point quantity, applies weighted-average cost before every sale, rejects sales above the available position with `409 PORTFOLIO_INSUFFICIENT_QUANTITY`, and never mutates wallet balances or ordinary reports. Clearing a price returns valuation to an explicit unknown state; it is not treated as zero. Corrections use compensating trades rather than editing history.

Authenticated jar routes: `GET/POST /api/v1/jars`, `PUT/DELETE /api/v1/jars/{jar_id}/months/{month}`, and `GET /api/v1/jars/{jar_id}/report?from=YYYY-MM&to=YYYY-MM`. Month routes: `GET /api/v1/months/{month}`, `PUT /api/v1/months/{month}/note`, and `DELETE /api/v1/months/{month}/note`. Month figures are live ledger calculations; completion is derived from the current account-local month, and the note is independently user-authored. There is no scheduled or immutable close/report snapshot in CORE-03.

## Field Ownership

- Human owns public contract intent and approval for contract changes.
- AI maintains contract rows, errors, auth notes, versioning, and linked decisions.

## API Surface

The current local Gin API will use a code-first Swagger contract: each Gin handler owns its Swagger annotations, and generated artifacts are derived from the Go source. Keep the generated surface limited to endpoints that exist in the running implementation; add annotations only when the corresponding handler and verification exist.

Swagger UI is available at `/api/v1/docs/index.html`; the unauthenticated public
entry point `/docs/` (and `/docs`) redirects there for stable links. Generated
artifacts live under `backend/docs/`. Run `go generate ./cmd/api` from
`backend/`: the generator declaration stays beside the Go API entry point, and
endpoint annotations stay beside their handlers. There is no separate Swagger
configuration file.

Frontend navigation is client-side and does not change the API base path. TanStack Router exposes `/`, `/transactions`, `/budgets`, `/reports`, `/assistant`, `/account`, `/account/groups`, `/account/wallets`, `/account/feedback`, `/account/recurring`, `/account/travel`, `/account/portfolio`, `/jars`, and `/months/{YYYY-MM}`. The global quick-add sheet persists the implemented basic income/expense contract with optional jar assignment. Reports are mounted at `/reports` and use the owner-scoped insider/month/transaction endpoints; no fixture totals are rendered.

Google login starts from the frontend origin using the relative `/api/v1/auth/google` path. Vite proxies that request during development, and the backend redirects the completed OAuth flow back to the same frontend origin at `/auth/callback`.

Document HTTP endpoints, RPC methods, events, CLI commands, or any other public contract.

| Contract | Type | Auth | Status | Notes |
| --- | --- | --- | --- | --- |
| `/api/v1/auth/devices`, `/auth/device-login`, `/auth/devices/{id}`, `/auth/logout` | HTTP REST | Enrollment/revoke accept JWT or owner key; device login uses its own credential | auth deployed; full-key extension local | Password confirmation still required to enroll. JWT logout revokes family; key logout revokes itself only. See [AUTH_API](AUTH_API.md). |
| OCR Platform document recognition | HTTP REST | `Authorization: Bearer sk_ocr_...` for protected requests | ready | See `docs/architecture/OCR_API.md`. |
| MyPocket local API | HTTP REST | JWT + rotating refresh, or full-owner mpk_ user key | local full-key implementation, rollout pending | See [authentication contract](AUTH_API.md); key secrets shown once/stored as hashes. Login/refresh/device exchange retain own credential contracts. |
| `/api/v1/api-keys` | HTTP REST | JWT or owner API key | implemented locally, rollout pending | `POST` name/optional expires_at creates full-owner key and returns secret once; `GET` metadata only; `DELETE /:id` revokes. No scopes; legacy input ignored, old stored metadata has no authority. |
| `/api/v1/auth/mobile/google/exchange` | HTTP REST | Public route; exact `platform=ios`/PKCE/redirect allowlist | implemented on native branch; production deploy pending | Exchanges a native iOS Google authorization code for the existing JWT + rotating refresh session. No client secret or token is sent in the app callback URL. Invalid profile is `401 GOOGLE_MOBILE_LOGIN_INVALID`; local owner persistence failure is `503 GOOGLE_USER_STORE_UNAVAILABLE`. |
| `/api/v1/categories` | HTTP REST | JWT or owner API key | implemented | `GET` returns all owner-visible categories, including inactive rows for historical labels and hierarchy; each row includes `is_active`. `POST`/`PATCH` accept name, kind, parent and owner-scoped applicable wallet IDs (empty means all owner wallets); category creation persists the category and selected links atomically; `DELETE` removes a personal group and returns its personal children to root. |
| `/api/v1/categories/{id}/status` | HTTP REST | JWT or owner API key | implemented locally | `PATCH {"is_active":false}` archives a personal category; `true` restores it. The response is the complete category model, with `wallet_ids` always an array (empty when no explicit wallet links exist), so native clients can decode it directly. Idempotent and owner-scoped; system or other-owner categories return not found. Existing transactions, reports, budgets and recurring schedules are preserved. New transaction, credit, budget, recurring and AI-entry assignments reject inactive categories with `400 category_inactive`; edits may retain an unchanged historical reference. |
| `/api/v1/categories/:id/wallets` | HTTP REST | JWT or owner API key | implemented | Replaces applicable-wallet selections for a visible personal or system category. System metadata remains immutable. |
| `/api/v1/wallets` | HTTP REST | JWT or owner API key | implemented; core UAT verified | Create/list/update/delete owner-scoped VND wallets. List returns `opening_balance` and ledger-derived `current_balance`; goal requires positive target and credit requires positive limit. A confirmed create may send exactly one ASCII `Idempotency-Key` (1–128 bytes); with the PostgreSQL mutation ledger, same-key/same-payload replays the original wallet, while changed payload returns `409`. Invalid or duplicate headers return `400`. A keyed request cannot fall back to a legacy adapter: missing durable support returns `503 mutation_unavailable`. Missing key keeps legacy online behavior. |
| `/api/v1/notifications` | HTTP REST | JWT or owner API key | implemented locally | Owner-scoped in-app inbox; `GET` lists newest-first and `POST /:id/read` is idempotent. No push delivery. |
| `/api/v1/transactions` | HTTP REST | JWT or owner API key | implemented; basic ledger UAT verified | Create/list/update/delete owner-scoped income and expense entries for basic/goal wallets. Category is required for ordinary rows and must match kind/applicable-wallet scope; credit mutations use the dedicated ledger. Adjustment rows are read through this collection but cannot be edited/deleted here. |
| `/api/v1/transactions/transfer` | HTTP REST | JWT or owner API key | implemented; automated proof, owner UAT pending | Atomically create paired source-expense/destination-income rows for an internal transfer; pair edit/delete require an idempotency key. Transfers are excluded from reports and jars. |
| `/api/v1/transactions/adjustment` | HTTP REST | JWT or owner API key | implemented; automated proof, owner UAT pending | Record a signed increase/decrease correction for a basic or goal wallet. Corrections affect wallet balance but never report totals and are immutable. |
| `/api/v1/transactions/bulk-delete` | HTTP REST | JWT or owner API key | implemented; automated proof, owner UAT pending | All-or-nothing deletion of 1–100 ordinary owner-scoped transactions. Transfer pairs and adjustments are rejected explicitly. |
| `/api/v1/recurring` | HTTP REST | JWT or owner API key | implemented; automated proof, owner UAT pending | CRUD recurring ordinary income/expense schedules and bounded idempotent due materialization. Generated rows remain normal ledger transactions. |
| `/api/v1/credit/wallets/{id}/entries` | HTTP REST | JWT or owner API key | implemented; automated backend/frontend checks | Dedicated purchase/refund/fee/interest entries for credit wallets. |
| `/api/v1/credit/wallets/{id}/payments` | HTTP REST | JWT or owner API key | implemented; automated backend checks | Atomic payment pair from a non-credit source wallet; both rows are excluded from reports. |
| `/api/v1/credit/wallets/{id}/statement` | HTTP REST | JWT or owner API key | implemented; automated backend checks | Owner-scoped credit statement with debt balance and available limit. |
| `/api/v1/travel/events` | HTTP REST | JWT or owner API key | implemented; automated backend checks | Owner-scoped event CRUD with one active event; deleting an event clears transaction links without deleting ledger rows. |
| `/api/v1/transactions/{id}/travel` | HTTP REST | JWT or owner API key | implemented; automated backend checks | Link or clear one ordinary income/expense transaction; transfers, adjustments and credit rows are rejected. |
| `/api/v1/portfolio/*` | HTTP REST | JWT or owner API key | implemented; unit/API contract checks | Owner-scoped assets, append-only buy/sell trades, weighted-average positions and optional manual valuation. Portfolio rows never mutate wallet/report ledger. |
| `/api/v1/reports/insider` | HTTP REST | JWT or full-owner API key | implemented; unit/API contract checks | Deterministic Money Insider summary for an account-local month: report-included totals, previous-period comparison, average/day, zero-safe spending/income ratio, top three categories and top five expenses. Optional owner wallet/category filters. |
| `/api/v1/auth/profile` | HTTP REST | JWT or owner API key | implemented | Reads/updates account IANA timezone; initial browser-zone adoption is initialize-only. |
| `/api/v1/jars` and `/api/v1/jars/{id}/...` | HTTP REST | JWT or owner API key | implemented; owner UAT pending | Stable jar IDs, month snapshots, optional ordinary-expense assignment and cumulative live report. `POST /api/v1/jars` accepts an optional ASCII `Idempotency-Key`; advisor confirmations use operation `jar.create` in the shared mutation ledger. Config removal does not delete transactions or historic snapshots. |
| `/api/v1/months/{YYYY-MM}` | HTTP REST | JWT or owner API key | implemented; owner UAT pending | Live account-local totals and independent note CRUD; completion is derived, not a close operation. |

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
### Travel optional-field PATCH semantics

For `PATCH /api/v1/travel/events/{id}`, omitted or JSON-null `context`,
`starts_on`, and `ends_on` retain the current value. Send an empty string to
explicitly clear one of these fields. Dates otherwise use `YYYY-MM-DD`.
Example: `{"context":"","starts_on":"","ends_on":"2026-10-31"}` clears
description/start while setting the end. Native full-form updates encode this
intent explicitly; create may omit optional values. Owner scope is unchanged.

## Feedback and changelog

Owner confirmation is separate from dev lifecycle: PATCH
`/api/v1/feedback/{id}/resolution` with explicit `resolved` boolean, JWT or
full-owner user key. Owner-scoped, reversible and idempotent, retaining
dev/changelog history. See feedback reference for failure codes.

The owner-scoped feedback lifecycle and the redacted service-token agent surface are documented in [FEEDBACK_API.md](FEEDBACK_API.md). User feedback routes accept full-owner API keys with no scopes. Public changelog never exposes raw feedback. The Feedback service token remains separate: user keys cannot access service/admin routes or all-user feedback.

## Finance query foundation — local implementation

Lifecycle addition deployed in the current API/worker image: `POST /api/v1/ai/advisor/conversation/reset`
requires `Idempotency-Key`, rotates the owner conversation generation, and stores
only a content-free reset ledger. It never touches wallets, transactions,
reports, or confirmed financial actions.

The shared backend finance query layer now provides owner-scoped, repeatable-read summaries, transaction search with all-scope counts and keyset cursors, category subtree filters, period comparison with zero-baseline semantics, wallet/goal/budget views, and a SELECT-only jar progress view. It is the internal service boundary used by the local advisor surface. Transfer rows and system transfer categories are excluded from report summaries, while search can inspect the full ledger scope.

Prepared wallet proposal confirmations use `POST /api/v1/ai/advisor/actions/{action_id}/confirm` with only `conversation_id` and `generation`. The backend resolves the owner-bound persisted proposal, rejects stale generations and client-supplied financial fields, and delegates to the durable wallet idempotency ledger. Legacy cards without action context retain the ordinary wallet endpoint fallback. This dispatcher is deployed; owner-authenticated confirmation UAT remains open.
Jar proposal confirmations use the same dispatcher. `propose_create_jar` returns
an explicit `jar_creation_proposal` and uses the `jar.create` mutation ledger;
`propose_update_jar` requires a UUID `jar_id` and returns a
`jar_update_proposal`. After confirmation, update dispatches to the
owner-scoped `PUT /api/v1/jars/{jar_id}/months/{month}` route, replacing that
month's configuration without creating a new jar. Both are deployed; owner-
authenticated confirmation UAT remains open.
Budget proposals use the same owner/generation-bound dispatcher. The
read-only `propose_create_budget` tool returns a `budget_creation_proposal`;
confirmation delegates to `POST /api/v1/budgets` with `budget.create` mutation
ledger protection. The client sends only the action binding, while the server
revalidates the existing budget/wallet/category/date contract.

`propose_update_budget` returns a `budget_update_proposal` containing a UUID
budget ID and replacement fields. Confirmation still sends only the
conversation/generation binding; the dispatcher resolves the persisted
proposal and delegates to the owner-scoped `PATCH /api/v1/budgets/{id}` route.

The Finance Assistant surface is mounted in production behind JWT/session or a user API key: `GET /api/v1/ai/advisor/capabilities`, `GET /api/v1/ai/advisor/overview`, `POST /api/v1/ai/advisor/messages`, `GET /api/v1/ai/advisor/conversation/messages`, `POST /api/v1/ai/advisor/conversation/reset`, `GET /api/v1/ai/advisor/runs/{id}`, and `POST /api/v1/ai/advisor/runs/{id}/cancel`. Overview reads real report data without calling the model; the `/reports` web route consumes this summary plus the owner-scoped month category breakdown and never renders fixture numbers. Chat persists one owner conversation/run and executes only the eight allowlisted read tools. Advisor history now has a server-authoritative 30-minute idle retention boundary: reads, polling and assistant output do not extend it; expired content is purged by generation and old request retries return `410 advisor_history_expired` without provider work. Financial tables and confirmed writes are never purged. The worker performs bounded cleanup for owners who do not return. Reset is owner-scoped, idempotent and content-free; it also cancels queued/running provider contexts in the current process, while generation/lease guards reject late writes. This additive behavior is deployed in the current API/worker image. History reads may add lifecycle metadata in `meta` (generation, last accepted user time, expiry and server time); old clients may ignore it and reads never extend the deadline. Advisor auth is audited best-effort to Redis stream `mypocket:audit:advisor`; durable delivery/alerting and replayable SSE remain later gates from the advisor plan.
