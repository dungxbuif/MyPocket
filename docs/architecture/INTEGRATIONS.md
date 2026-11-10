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
| OpenAI-compatible model API | Finance Assistant reasoning and memory candidate extraction | `FINANCE_ASSISTANT_MODEL_BASE_URL`, `FINANCE_ASSISTANT_MODEL_NAME`, secret `finance_assistant_model_api_key` | deployment-managed; never exposed to iOS/Telegram | local contract tests; production release pending |
| Telegram Bot API | Optional per-owner channel adapters and webhooks | one encrypted bot token per adapter; secret-token webhook | each owner supplies and may revoke each bot | local adapter tests; production release pending |

## Credentials And Secrets

### External Finance Assistant model and Telegram

The Python service is the only new component that calls the model. It uses the
same OpenAI-compatible boundary as MyPocket, but has separate deployment
configuration. Local defaults remain `http://127.0.0.1:1234/v1` and
`Qwen3.6-35B-A3B-MLX-4bit`; the production manifest selects
`https://token-api.fpt.ai/v1` and `DeepSeek-V4.1-Flash`. This is a deployment
selector, not evidence that the production release has run. The client sends
`max_completion_tokens=4096`, uses a 240-second timeout, disables library
retries and has no provider fallback. Base URLs fail closed unless they are
absolute HTTP(S) URLs without userinfo, query or fragment.

Model prompts contain the owner's recent assistant conversation, recalled
memory and the finance data returned by the owner-bounded MyPocket delegation.
This means configured provider processing includes personal financial data;
deployment approval must treat the provider as a data processor. Provider keys
are Docker secrets and are never placed in native builds, Telegram messages,
Markdown projections or logs.

An owner adds a Telegram bot token in the native app. The service validates it
with `getMe`, installs an HTTPS webhook with a generated secret, encrypts the
token at rest, and returns only adapter metadata. Pairing uses a single-use
ten-minute deep-link code. Deleting an adapter first removes its webhook and
revokes its bindings. The bot token pasted in a chat or ticket is compromised
and must be revoked with BotFather before use; it must not be copied into a
database, environment file or CI variable.

Current production verification (2026-10-20): the FPT AI deployment selector is
`DeepSeek-V4.1-Flash`. The older `DeepSeek-V4-Flash` text in the historical
checkpoint below is retained as evidence from that earlier run, not as the
current production selector.

AI entry uses the backend-only OpenAI-compatible chat-completions adapter with `AI_BASE_URL`, `AI_API_KEY`, `AI_MODEL`; base URL should include the provider's API prefix such as `/v1`, not `/chat/completions`. The committed example remains the local oMLX profile at `http://127.0.0.1:1234/v1` with `Qwen3.6-35B-A3B-MLX-4bit`; the historical FPT validation override used `https://token-api.fpt.ai/v1` with `DeepSeek-V4-Flash`. The current production selector is `DeepSeek-V4.1-Flash`. Provider keys remain in ignored `backend/.env.local` or the process environment and must never be committed. Startup/client configuration validation now fails closed unless the endpoint is absolute HTTP(S) without userinfo, query or fragment and both model and key are present; validation messages contain field labels only and never credential material. Extraction and advisor clients use this same guard, send credentials only in the Authorization header, and never retry or silently fall back to another provider. It requests strict `json_schema` structured output (no model tools), explicitly caps each generation at 32,768 tokens (the previous 2,048-token budget truncated nine-image batches), sets temperature to zero, and sends `chat_template_kwargs.enable_thinking=false` for this extraction path. This is required for Qwen/oMLX: omitting `max_tokens` inherits the local server's 1,000,000-token default and can leave a request running until MyPocket's old 30-second model timeout, producing `stage=model`, `code=model_request`, `provider_status=0`. The model timeout is now 180 seconds and the overall extraction budget 210 seconds to cover a cold local model prefill; the model still receives no images/files. The application imposes no draft-count limit; only provider/transport envelopes apply. A synthetic five-case run on the configured Qwen endpoint passed 24/24 scored fields; the FPT model-list and strict JSON chat smoke also passed on 2026-09-23. The bounded owner wallet catalog includes any saved wallet description (max 2 KiB each) as untrusted matching context. No runtime mock fallback, automatic provider fallback, or per-user usage cap is active. Future user-selectable providers, sourced price comparisons, cost consent and usage budgets are a separate draft under [AI-ENTRY-03](../work/tickets/AI-ENTRY-03-PROVIDER-SELECTION.md); only the current server-configured endpoint is integrated.

Local deterministic ingestion candidate (2026-10-27): supported bank XLSX files are parsed before the provider with bounded `archive/zip` and `encoding/xml`; no spreadsheet execution library is used. Canonical rows are enriched in batches of at most 25 and exact source-ID set validation fails atomically. The provider cannot alter parser-owned date, amount, direction or currency. XLSX and OCR-backed files cannot be mixed in one request. This candidate is locally tested and is not a production-deployment claim.

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

Local development reads `backend/.env.local` (Git-ignored, mode 0600), with process environment taking precedence. OCR, LLM and S3 credentials are backend-only. S3 config names: `S3_ENDPOINT`, `S3_REGION`, `S3_BUCKET`, `S3_PREFIX`, `S3_FORCE_PATH_STYLE`, `S3_ACCESS_KEY_ID`, `S3_SECRET_ACCESS_KEY`; storage is wired into API composition and files are disabled when it is not configured in development. Originals and OCR text persist in `transaction_attachments`; approval inserts `transaction_attachment_links` atomically. Authenticated download checks transaction owner and attachment link, then redirects to a five-minute private URL. The regular worker calls a bounded expired-unlinked cleanup pass only when its private S3 configuration is complete; failed deletions remain retryable. A read-only 2026-10-26 check found the deployed worker lacks the S3 credential environment variables, so scheduled cleanup is currently disabled there. Production OCR/S3 UAT remains pending. The one-shot OCR adapter sends validated image bytes as Base64 and uses OCR Platform's private presign flow for PDFs; the LLM receives extracted text only and no conversation text is persisted. [AI-ENTRY-02](../work/tickets/AI-ENTRY-02-BATCH-ATTACHMENTS-DETAIL_DESIGN.md), [feedback audit](../work/bugs/FEEDBACK-baca2415-ai-entry-attachments.md), [ADR-006](../decisions/ADR-006-private-ai-attachments.md).

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
