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

## Budget and savings integration — 2026-09-20

Authenticated `GET/POST /api/v1/budgets`, `PATCH/DELETE /api/v1/budgets/{id}` implemented. Explicit interval input: name, positive safe-integer limit_amount, nullable owner wallet_id and visible expense category_id, RFC3339 start_at inclusive/end_at exclusive. List returns items with ledger-derived spent/days_remaining/ended and active limit_amount/spent summary (transaction IDs deduplicated). Category scope includes descendants. Only report-included expenses count. Exact same scope and overlapping dates returns 409; ended budgets cannot be edited. Delete configuration leaves transactions unchanged. [Design/proof](../work/tickets/API-SCREENS-01-DETAIL_DESIGN.md), [ADR-004](../decisions/ADR-004-budget-api-data.md).

Goal transaction categories, when explicitly selected, must use real catalog keys income_transfer_in/income_interest for income or expense_transfer_out for expense, in addition to kind/wallet applicability. Category omission remains allowed. External savings entries affect one wallet and retain report inclusion default true; internal paired transfers are not implemented by this path.

## Wallet goal date

Wallet create/update accepts optional `target_date` as `YYYY-MM-DD` for goal wallets. Invalid calendar dates return 400. Existing `target_date` timestamp column stores UTC midnight; response uses the existing timestamp representation, with clients retaining its first ten date characters. Omission on update preserves the date; empty string clears it. Other wallet types ignore this input. No migration. See [savings design](../work/tickets/TICKET-06-01-DETAIL_DESIGN.md).

## Field Ownership

- Human owns public contract intent and approval for contract changes.
- AI maintains contract rows, errors, auth notes, versioning, and linked decisions.

## API Surface

The current local Gin API will use a code-first Swagger contract: each Gin handler owns its Swagger annotations, and generated artifacts are derived from the Go source. Keep the generated surface limited to endpoints that exist in the running implementation; add annotations only when the corresponding handler and verification exist.

Swagger UI is available at `/api/v1/docs/index.html`; generated artifacts live under `backend/docs/`. Run `go generate ./cmd/api` from `backend/`: the generator declaration stays beside the Go API entry point, and endpoint annotations stay beside their handlers. There is no separate Swagger configuration file.

Frontend navigation is client-side and does not change the API base path. TanStack Router currently exposes `/`, `/transactions`, `/budgets`, `/account`, `/account/groups`, and `/account/wallets`; the global quick-add sheet now persists the implemented basic income/expense contract. Reports remain unmounted until their APIs exist.

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
