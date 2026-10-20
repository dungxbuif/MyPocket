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
| OCR Platform (`https://ocr.dungxbuif.com`) | Receipt OCR for images/PDFs before text-only LLM extraction; scan enhancement not in current public OpenAPI | `docs/architecture/OCR_API.md` | human-approved provider; credentials are deployment-managed | adapter tests pass locally; current live OCR/S3 UAT is credential-gated and pending |

## Credentials And Secrets

AI entry uses the backend-only OpenAI-compatible chat-completions adapter with `AI_BASE_URL`, `AI_API_KEY`, `AI_MODEL`; base URL should include the provider's API prefix such as `/v1`, not `/chat/completions`. The committed example remains the local oMLX profile at `http://127.0.0.1:1234/v1` with `Qwen3.6-35B-A3B-MLX-4bit`; the active ignored `backend/.env.local` validation override is FPT AI `https://token-api.fpt.ai/v1` with `DeepSeek-V4-Flash`. Provider keys remain in ignored `backend/.env.local` or the process environment and must never be committed. Startup/client configuration validation now fails closed unless the endpoint is absolute HTTP(S) without userinfo, query or fragment and both model and key are present; validation messages contain field labels only and never credential material. Extraction and advisor clients use this same guard, send credentials only in the Authorization header, and never retry or silently fall back to another provider. It requests strict `json_schema` structured output (no model tools), explicitly caps each generation at 32,768 tokens (the previous 2,048-token budget truncated nine-image batches), sets temperature to zero, and sends `chat_template_kwargs.enable_thinking=false` for this extraction path. This is required for Qwen/oMLX: omitting `max_tokens` inherits the local server's 1,000,000-token default and can leave a request running until MyPocket's old 30-second model timeout, producing `stage=model`, `code=model_request`, `provider_status=0`. The model timeout is now 180 seconds and the overall extraction budget 210 seconds to cover a cold local model prefill; the model still receives no images/files. The application imposes no draft-count limit; only provider/transport envelopes apply. A synthetic five-case run on the configured Qwen endpoint passed 24/24 scored fields; the FPT model-list and strict JSON chat smoke also passed on 2026-09-23. The bounded owner wallet catalog includes any saved wallet description (max 2 KiB each) as untrusted matching context. No runtime mock fallback, automatic provider fallback, or per-user usage cap is active. Future user-selectable providers, sourced price comparisons, cost consent and usage budgets are a separate draft under [AI-ENTRY-03](../work/tickets/AI-ENTRY-03-PROVIDER-SELECTION.md); only the current server-configured endpoint is integrated.

The 2026-10-07 native-port checkpoint reran the same ignored FPT override with
`DeepSeek-V4-Flash`: read-only extraction passed, the strict
`propose_transactions` tool-call compatibility probe returned exactly one
schema-valid call, and five strict JSON Schema cases passed all semantic
checks. The current worktree has no OCR credential, so this checkpoint does not
claim a new image-provider run; the earlier synthetic S3→OCR→LLM evidence is
retained in the validation matrix.

Current evidence rule (2026-10-19): the historical FPT/Qwen probes above are
not proof that the ignored local environment or any deployment currently has
those credentials. The current deterministic gate is provider-free. Treat AI,
OCR and S3 availability as unverified until [PROVIDER_UAT.md](../operations/PROVIDER_UAT.md)
is completed with approved synthetic inputs; never copy credentials into this
document or into client code.

Local development reads `backend/.env.local` (Git-ignored, mode 0600), with process environment taking precedence. OCR, LLM and S3 credentials are backend-only. S3 config names: `S3_ENDPOINT`, `S3_REGION`, `S3_BUCKET`, `S3_PREFIX`, `S3_FORCE_PATH_STYLE`, `S3_ACCESS_KEY_ID`, `S3_SECRET_ACCESS_KEY`; storage is wired into API composition and files are disabled when it is not configured in development. Originals and OCR text persist in `transaction_attachments`; approval inserts `transaction_attachment_links` atomically. Authenticated download checks transaction owner and attachment link, then redirects to a five-minute private URL. Cleanup of expired unlinked objects and production OCR/S3 UAT remain pending. The one-shot OCR adapter sends validated image bytes as Base64 and uses OCR Platform's private presign flow for PDFs; the LLM receives extracted text only and no conversation text is persisted. [AI-ENTRY-02](../work/tickets/AI-ENTRY-02-BATCH-ATTACHMENTS-DETAIL_DESIGN.md), [ADR-006](../decisions/ADR-006-private-ai-attachments.md).

Document required secret names only. Do not store secret values.

`AI_STORE_USAGE` controls internal model usage retention. When enabled, completed AI-entry processes and Finance Assistant runs store provider request id, model, finish reason, latency, normalized token counts and the provider usage JSON in owner-scoped JSONB metadata. This is operational/cost telemetry, not user financial data, and is excluded from public responses. Detailed AI logs are privacy-safe stage metadata only. Provider streaming is not enabled for the one-shot proposal endpoint: structured JSON must be complete before any draft is persisted. To eliminate browser/gateway timeouts, the next transport evolution is asynchronous processing with polling/SSE, not partial JSON persistence.

The Advisor OpenAI-compatible adapter maps upstream HTTP failures, malformed
responses, oversized responses, and transport failures to the stable
`advisor_unavailable`/HTTP 503 contract. It never forwards the upstream body,
endpoint, prompt, or credential material. This keeps transient provider errors
distinct from malformed user requests (400) and prevents provider outages from
appearing as opaque internal 500 errors in the client.

- `OCR_API_URL`
- `OCR_API_KEY`

## Failure Modes

- OCR provider unavailable: keep receipt attachment and manual transaction entry available.
- OCR result expired before MyPocket persisted it: ask user to re-run OCR or continue manually.
- Upload too large or unsupported file type: block or explain before submission where possible.
- Presigned upload failure: request a fresh presign URL and do not submit stale `sourceUrl`.
- OCR recognition low quality: show extracted fields as suggestions, never as final transaction data without user confirmation.
