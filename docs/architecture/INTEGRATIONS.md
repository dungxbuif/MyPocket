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
| OCR Platform (`https://ocr.dungxbuif.com`) | Receipt OCR for images/PDFs, optional future scan enhancement | `docs/architecture/OCR_API.md` | human-approved provider pending final credential setup | ready for implementation planning |

## Credentials And Secrets

Document required secret names only. Do not store secret values.

- `OCR_API_URL`
- `OCR_API_KEY`

## Failure Modes

- OCR provider unavailable: keep receipt attachment and manual transaction entry available.
- OCR result expired before MyPocket persisted it: ask user to re-run OCR or continue manually.
- Upload too large or unsupported file type: block or explain before submission where possible.
- Presigned upload failure: request a fresh presign URL and do not submit stale `sourceUrl`.
- OCR recognition low quality: show extracted fields as suggestions, never as final transaction data without user confirmation.
