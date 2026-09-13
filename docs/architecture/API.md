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

## Field Ownership

- Human owns public contract intent and approval for contract changes.
- AI maintains contract rows, errors, auth notes, versioning, and linked decisions.

## API Surface

The current local Gin API will use a code-first Swagger contract: each Gin handler owns its Swagger annotations, and generated artifacts are derived from the Go source. Keep the generated surface limited to endpoints that exist in the running implementation; add annotations only when the corresponding handler and verification exist.

Swagger UI is available at `/api/v1/docs/index.html`; generated artifacts live under `backend/docs/` and must be regenerated with `swag init` from the `backend/` module.

Frontend navigation is client-side and does not change the API base path. TanStack Router owns `/`, `/transactions`, `/budgets`, `/reports`, `/account`, `/account/groups`, and `/account/wallets`.

Google login starts from the frontend origin using the relative `/api/v1/auth/google` path. Vite proxies that request during development, and the backend redirects the completed OAuth flow back to the same frontend origin at `/auth/callback`.

Document HTTP endpoints, RPC methods, events, CLI commands, or any other public contract.

| Contract | Type | Auth | Status | Notes |
| --- | --- | --- | --- | --- |
| OCR Platform document recognition | HTTP REST | `Authorization: Bearer sk_ocr_...` for protected requests | ready | See `docs/architecture/OCR_API.md`. |
| MyPocket local API | HTTP REST | JWT bearer for `/auth/profile` and `/home`; OAuth cookies for Google callback | implemented for current auth/profile/home slice | Code-first Swagger annotations in Go handlers; generated UI is added incrementally. |
| `GET /api/v1/categories` | HTTP REST | JWT bearer | implemented | Returns system categories plus categories owned by the authenticated user. |

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
