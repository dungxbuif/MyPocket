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

Document HTTP endpoints, RPC methods, events, CLI commands, or any other public contract.

| Contract | Type | Auth | Status | Notes |
| --- | --- | --- | --- | --- |
| OCR Platform document recognition | HTTP REST | `Authorization: Bearer sk_ocr_...` for protected requests | ready | See `docs/architecture/OCR_API.md`. |

## Errors

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
