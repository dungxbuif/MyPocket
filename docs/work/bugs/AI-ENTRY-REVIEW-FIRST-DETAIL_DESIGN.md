# AI entry: extract first, deduplicate and prefill for review

Status: in_review; regression_verified locally, owner UAT pending. Approval: owner explicitly requested automatic duplicate filtering and inferred/default wallets in this conversation (2026-09-23).

## Cause and scope

The system prompt simultaneously required drafts and allowed empty clarification responses; it told the model to ask when OCR was incomplete or duplicated. The latest provider reply follows that latter instruction and refuses the entire batch. Previous single-transaction live probes did not exercise overlapping bank screenshots. Wallet fallback runs only after drafts exist, so it cannot repair an empty result.

## Implementation

Rewrite the extraction instructions as one consistent review-first policy: extract recognizable entries even with incomplete fields; collapse clear overlapping copies within the submitted batch; preserve distinct dates/references and ambiguous repeats for review; select the best existing wallet or the first supplied default. Missing wallet catalogs still produce incomplete drafts. Missing/unclear dates default to the current account-local instant and receive a review question; amounts remain grounded and missing amounts stay incomplete. Do not infer transactions from balances, advertising amounts or credit limits. Transfers remain transfer drafts and never become reportable expenses just to bypass approval checks. Keep amounts grounded in the source and use draft questions for missing data. No provider retry, schema change or automatic ledger write.

Verification: an opt-in live synthetic overlap case must produce three distinct entries from four transaction views (one duplicate), including an incomplete entry. A separate read-only replay test loads the stored OCR and owner catalogs of a selected failed process, calls the configured model once, checks nonempty drafts and owner-wallet assignment, and prints only counts. No OCR re-upload or proposal/ledger mutation. Run surrounding Go tests; restart the local API after verification.

Batch-output finding: after the prompt correction, the nine-image replay reached the existing 2048-token output cap (`finish_reason=length`, 4857 content bytes), rather than completing JSON. The bounded extraction output budget is now 32768 tokens; this is only a provider-response envelope, not an application draft-count limit. Retain request deadlines, 256 KiB transport safety, strict parsing and no automatic retries. Log only the known truncation reason, configured cap and byte count. This is part of making the requested batch extraction work, not a user usage limit or a new provider integration.

Runtime regression found 2026-09-23: adding the non-null `model_usage` JSONB columns caused new AI entry/advisor rows to send SQL `NULL` from a nil Go map, producing a generic HTTP 500 before provider work. Creation paths now initialize an empty usage map; the DB default remains a migration safety net. This failure is persistence/configuration, not OCR or model behavior.

Reconcile [AI entry page](../../design/pages/assistant/README.md), [API](../../architecture/API.md), [changelog](../../releases/CHANGELOG.md), [validation](../VALIDATION_MATRIX.md), [context](../../CONTEXT.md), [backlog](../BACKLOG.md). Parent: [AI-ENTRY-02](../tickets/AI-ENTRY-02-BATCH-ATTACHMENTS-DETAIL_DESIGN.md); no active phase. Existing API/schema/security boundaries unchanged. Markdown API and page docs remain human/agent-readable sources; no new endpoint is introduced.

## Verification — 2026-09-23

- RED, configured DeepSeek-V4-Flash: old prompt returned 2 instead of 3 synthetic drafts (lost the cropped-date row); replay of nine stored OCR attachments returned zero drafts.
- Prompt-only correction: overlap case passed; the first real replay failed at the old `finish_reason=length`, 2048-token cap and 4857 response bytes. The transport envelope was increased to 32768 tokens; no automatic retry was introduced.
- GREEN after output-budget correction: held-out overlap fixture (different merchant, references, dates and amounts from the prompt example) produced three drafts; duplicate reference collapsed, two distinct purchase dates preserved, cropped date defaulted to account-local today with review question, sole owner wallet prefilled, balance excluded.
- GREEN: nine stored OCR attachments produced 30 drafts with owned wallet IDs, provider latency 40.682 seconds. This replay used SELECT-only local DB access and one model request; no OCR re-upload, proposal persistence, approval or ledger writes. It proves nonempty extraction and wallet prefill, not completeness or amount-by-amount ground truth. No application draft-count cap remains.
- `go test ./...`: PASS. Opt-in provider tests and database integration tests requiring explicit env are not implied by this command.
- `go test -race ./internal/infrastructure/ai ./internal/usecase`: PASS. Provider adapter tests verify complete 30- and 31-draft acceptance and rejection of truncated batches without partial proposals or retry.
- `go build ./cmd/api`: PASS. Generated tracked binary restored to its prior committed state; runtime restarted with `go run ./cmd/api` so the new embedded prompt is active.
- Backend `/api/v1/health` and frontend `http://localhost:4173/`: HTTP 200 after restart. Browser upload/save UAT was not repeated; existing requests remain idempotent and retain their old outcome. Test the new prompt with a new submission.

## Reconciliation and lessons

Internal business rule AI-01, ADR-005 and integration runtime notes updated. Public/agent-readable Markdown API reference and review-screen behavior updated with partial drafts, within-batch dedup, default wallet and transport-only output limits. No API shape, schema, auth, OpenAPI or frontend component change in this correction, so no new schema generation or frontend build required. Changelog and validation/context/backlog updated. Public-site publishing and production deployment are not claimed.

Do not mix mandatory extraction with clarification-first escape clauses for the same uncertainty. Do not use one complete text transaction as proof that overlapping multi-image batches work. Raising extraction recall also requires testing output-token capacity; provider success on a tiny input is insufficient. Keep uncertain repeats for human review and do not confuse within-submission dedup with ledger reconciliation.

## Follow-up timeout regression — 2026-09-23

A real submission returned 503 after approximately 91 seconds while `/health` stayed 200. The backend log showed the AI process completed with 503 and recovery remained available; this is a model request timeout, not a service/database outage. The previous usecase context and HTTP client both stopped at 90 seconds even though the extraction path had a longer cold-start budget. The regression test now requires at least 120 seconds; production model timeout is 180 seconds and the enclosing extraction/usecase budget is 210 seconds. No automatic provider retry was added.

## Debugging, usage and instruction scope — 2026-09-23

Owner requested detailed but privacy-safe AI diagnostics, optional model usage retention, and reliable user text filters. The decision is to log stage transitions and bounded metadata only (process/request id, counts, byte sizes, provider/model, status, finish reason, latency and safe error codes); never log API keys, prompts, OCR, notes, categories, merchants or amounts. `AI_STORE_USAGE` controls persistence of provider usage JSON plus normalized token/latency metadata on the owner-scoped process; usage is internal and is not returned in the public process response. The model receives `user_instruction` separately from `untrusted_source_text`; the instruction may filter which recognizable rows become drafts but cannot override schema, ownership, approval or privacy rules. Existing AI reads and writes must carry the authenticated owner; global attachment cleanup is intentionally an internal cross-owner job and does not expose data.
