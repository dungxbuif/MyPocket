---
artifact_type: integrations_doc
id: INTEGRATIONS
status: draft
owner: shared
human_fields: [provider_choice, credential_approval, business_owner]
ai_fields: [systems, contracts, secrets_names, failure_modes]
shared_fields: [status, trace]
---

# Integrations

## Field Ownership

- Human owns provider choices, credential approval, and business owner.
- AI documents systems, contracts, secret names, and failure modes.

## External Systems

| System | Purpose | Contract | Owner | Status |
| --- | --- | --- | --- | --- |
| OCR Platform (`https://ocr.dungxbuif.com`) | Receipt OCR for images/PDFs; scan enhancement not in current public OpenAPI | `docs/architecture/OCR_API.md` | human-approved provider pending final credential setup | ready for implementation planning |

## Credentials And Secrets

AI entry uses the backend-only text provider adapter with `AI_BASE_URL`, `AI_API_KEY`, `AI_MODEL`; base URL should include the provider's API prefix such as `/v1`, not `/chat/completions`. It requests JSON object output and sends no images to the model. Owner's endpoint/model/key are pending; no runtime mock fallback.

Local development reads `backend/.env.local` (Git-ignored, mode 0600), with process environment taking precedence. OCR and S3 keys supplied by the owner are stored only there. S3 config names: `S3_ENDPOINT`, `S3_REGION`, `S3_BUCKET`, `S3_PREFIX`, `S3_FORCE_PATH_STYLE`, `S3_ACCESS_KEY_ID`, `S3_SECRET_ACCESS_KEY`. Endpoint is `https://storage.dungxbuif.com`; bucket/region are not supplied yet. Config presence does not mean a storage adapter or retained receipt uploads are implemented. Current images use transient OCR base64; text is persisted in the session. [Current slice](../work/tickets/TICKET-09-01-ENTRY-DETAIL_DESIGN.md).

Document required secret names only. Do not store secret values.

- `OCR_API_URL`
- `OCR_API_KEY`

## Failure Modes

- OCR provider unavailable: keep receipt attachment and manual transaction entry available.
- OCR result expired before MyPocket persisted it: ask user to re-run OCR or continue manually.
- Upload too large or unsupported file type: block or explain before submission where possible.
- Presigned upload failure: request a fresh presign URL and do not submit stale `sourceUrl`.
- OCR recognition low quality: show extracted fields as suggestions, never as final transaction data without user confirmation.
