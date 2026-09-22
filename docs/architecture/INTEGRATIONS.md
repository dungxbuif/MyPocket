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
| OCR Platform (`https://ocr.dungxbuif.com`) | Receipt OCR for images/PDFs before text-only LLM extraction; scan enhancement not in current public OpenAPI | `docs/architecture/OCR_API.md` | human-approved provider; local credentials are configured | adapter tests pass; live OCR UAT pending |

## Credentials And Secrets

AI entry uses the backend-only OpenAI-compatible chat-completions adapter with `AI_BASE_URL`, `AI_API_KEY`, `AI_MODEL`; base URL should include the provider's API prefix such as `/v1`, not `/chat/completions`. It requests strict `json_schema` structured output (no model tools) and sends no images/files to the model. A synthetic five-case run on the configured Qwen endpoint passed 24/24 scored fields; real owner UAT remains pending. The bounded owner wallet catalog includes any saved wallet description (max 2 KiB each) as untrusted matching context. No runtime mock fallback or automatic provider fallback. Future user-selectable providers and sourced price comparisons are a separate draft under [AI-ENTRY-03](../work/tickets/AI-ENTRY-03-PROVIDER-SELECTION.md); only the current server-configured endpoint is integrated.

Local development reads `backend/.env.local` (Git-ignored, mode 0600), with process environment taking precedence. OCR, LLM and S3 credentials are backend-only. S3 config names: `S3_ENDPOINT`, `S3_REGION`, `S3_BUCKET`, `S3_PREFIX`, `S3_FORCE_PATH_STYLE`, `S3_ACCESS_KEY_ID`, `S3_SECRET_ACCESS_KEY`; storage is wired into API composition and files are disabled when it is not configured in development. Originals and OCR text persist in `transaction_attachments`; approval inserts `transaction_attachment_links` atomically. Authenticated download checks transaction owner and attachment link, then redirects to a five-minute private URL. Cleanup of expired unlinked objects and live OCR/S3 UAT remain pending. The one-shot OCR adapter processes backend-provided private URLs before LLM; no conversation text is persisted. [AI-ENTRY-02](../work/tickets/AI-ENTRY-02-BATCH-ATTACHMENTS-DETAIL_DESIGN.md), [ADR-006](../decisions/ADR-006-private-ai-attachments.md).

Document required secret names only. Do not store secret values.

- `OCR_API_URL`
- `OCR_API_KEY`

## Failure Modes

- OCR provider unavailable: keep receipt attachment and manual transaction entry available.
- OCR result expired before MyPocket persisted it: ask user to re-run OCR or continue manually.
- Upload too large or unsupported file type: block or explain before submission where possible.
- Presigned upload failure: request a fresh presign URL and do not submit stale `sourceUrl`.
- OCR recognition low quality: show extracted fields as suggestions, never as final transaction data without user confirmation.
