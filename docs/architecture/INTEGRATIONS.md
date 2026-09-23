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
| OCR Platform (`https://ocr.dungxbuif.com`) | Receipt OCR for images/PDFs before text-only LLM extraction; scan enhancement not in current public OpenAPI | `docs/architecture/OCR_API.md` | human-approved provider; local credentials are configured | adapter tests and local receipt smoke pass; production OCR UAT pending |

## Credentials And Secrets

AI entry uses the backend-only OpenAI-compatible chat-completions adapter with `AI_BASE_URL`, `AI_API_KEY`, `AI_MODEL`; base URL should include the provider's API prefix such as `/v1`, not `/chat/completions`. The committed example remains the local oMLX profile at `http://127.0.0.1:1234/v1` with `Qwen3.6-35B-A3B-MLX-4bit`; the active ignored `backend/.env.local` validation override is FPT AI `https://token-api.fpt.ai/v1` with `DeepSeek-V4-Flash`. Provider keys remain in ignored `backend/.env.local` or the process environment and must never be committed. It requests strict `json_schema` structured output (no model tools), explicitly caps each generation at 8,192 tokens (the previous 2,048-token budget truncated nine-image batches), sets temperature to zero, and sends `chat_template_kwargs.enable_thinking=false` for this extraction path. This is required for Qwen/oMLX: omitting `max_tokens` inherits the local server's 1,000,000-token default and can leave a request running until MyPocket's old 30-second model timeout, producing `stage=model`, `code=model_request`, `provider_status=0`. The model timeout is now 180 seconds and the overall extraction budget 210 seconds to cover a cold local model prefill; the model still receives no images/files. The application imposes no draft-count limit; only provider/transport envelopes apply. A synthetic five-case run on the configured Qwen endpoint passed 24/24 scored fields; the FPT model-list and strict JSON chat smoke also passed on 2026-09-23. The bounded owner wallet catalog includes any saved wallet description (max 2 KiB each) as untrusted matching context. No runtime mock fallback, automatic provider fallback, or per-user usage cap is active. Future user-selectable providers, sourced price comparisons, cost consent and usage budgets are a separate draft under [AI-ENTRY-03](../work/tickets/AI-ENTRY-03-PROVIDER-SELECTION.md); only the current server-configured endpoint is integrated.

Local development reads `backend/.env.local` (Git-ignored, mode 0600), with process environment taking precedence. OCR, LLM and S3 credentials are backend-only. S3 config names: `S3_ENDPOINT`, `S3_REGION`, `S3_BUCKET`, `S3_PREFIX`, `S3_FORCE_PATH_STYLE`, `S3_ACCESS_KEY_ID`, `S3_SECRET_ACCESS_KEY`; storage is wired into API composition and files are disabled when it is not configured in development. Originals and OCR text persist in `transaction_attachments`; approval inserts `transaction_attachment_links` atomically. Authenticated download checks transaction owner and attachment link, then redirects to a five-minute private URL. Cleanup of expired unlinked objects and production OCR/S3 UAT remain pending. The one-shot OCR adapter sends validated image bytes as Base64 and uses OCR Platform's private presign flow for PDFs; the LLM receives extracted text only and no conversation text is persisted. [AI-ENTRY-02](../work/tickets/AI-ENTRY-02-BATCH-ATTACHMENTS-DETAIL_DESIGN.md), [ADR-006](../decisions/ADR-006-private-ai-attachments.md).

Document required secret names only. Do not store secret values.

`AI_STORE_USAGE` controls internal model usage retention. When enabled, completed AI-entry processes and Finance Assistant runs store provider request id, model, finish reason, latency, normalized token counts and the provider usage JSON in owner-scoped JSONB metadata. This is operational/cost telemetry, not user financial data, and is excluded from public responses. Detailed AI logs are privacy-safe stage metadata only. Provider streaming is not enabled for the one-shot proposal endpoint: structured JSON must be complete before any draft is persisted. To eliminate browser/gateway timeouts, the next transport evolution is asynchronous processing with polling/SSE, not partial JSON persistence.

- `OCR_API_URL`
- `OCR_API_KEY`

## Failure Modes

- OCR provider unavailable: keep receipt attachment and manual transaction entry available.
- OCR result expired before MyPocket persisted it: ask user to re-run OCR or continue manually.
- Upload too large or unsupported file type: block or explain before submission where possible.
- Presigned upload failure: request a fresh presign URL and do not submit stale `sourceUrl`.
- OCR recognition low quality: show extracted fields as suggestions, never as final transaction data without user confirmation.
