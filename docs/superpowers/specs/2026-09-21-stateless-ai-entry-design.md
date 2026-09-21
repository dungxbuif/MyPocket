# Stateless AI Entry and PDF OCR Design

## Intent and approved product behavior

Replace the current conversational/session-shaped AI entry API with a one-shot processing request. The user submits text and optional receipt files once, receives a reviewable list of transaction proposals, and explicitly edits, removes, approves, or saves proposals. No conversation history, message API, or new/resume session UI is exposed. No ledger transaction is created during extraction.

The user's correction on 2026-09-21 is authoritative: each operation is an API action in its own right; AI extraction is not represented as a chat session. The existing approval requirement remains: only an explicit proposal approval creates a ledger transaction.

## Root cause evidence

- Browser reported `POST /api/v1/ai/entry/sessions/{id}/messages` and HTTP 503.
- `AiEntrySheet` calls `createEntrySession()` and then `sendEntryMessage(session.id, ...)`.
- The UI accepts `application/pdf`, but `backend/internal/infrastructure/ai/ocr.go` only allows `image/jpeg` and `image/png`.
- `AIEntryService.Send` maps every extractor error to `ErrAIProvider`; `aiEntryFail` maps that to 503 and returns the generic potentially-billable message.
- Therefore this PDF is rejected by local OCR-input validation before any OCR/model request. The displayed warning overstates the risk for this specific failure path.

## API contract

- `POST /api/v1/ai/entry/process`: authenticated, one-shot extraction request. Accept multipart `request_id`, `text`, `timezone`, and up to three files (`image/jpeg`, `image/png`, or `application/pdf`, each at most 5 MiB). Return an opaque processing/batch ID and editable proposal list only after extraction finishes. Repeating the same idempotency key and payload returns the stored outcome; reusing the key with a different payload conflicts. No automatic provider retry.
- `GET /api/v1/ai/entry/requests/{request_id}`: owner-scoped fetch of processing status and proposals using the client-generated idempotency UUID, including when the original POST response was lost. It never starts extraction.
- Existing proposal operations remain independent: `PATCH /proposals/{id}`, `POST /proposals/{id}/approve`, and `POST /proposals/{id}/reject`. They are scoped to owner and proposal version.
- Remove frontend calls to `/sessions`, `/sessions/latest`, and `/sessions/{id}/messages`. Stop generating conversational history for extraction. Internal persistence may retain an opaque process record for idempotency and proposal lifecycle, but it must not expose or behave as a conversation/session.

## PDF and error behavior

Files are always processed by backend code and OCR before any LLM call. The LLM must never receive the original file, base64 data, a file URL (signed or public), or a multimodal file/image input. Validate file count, MIME, decoded bytes, file signatures, and size before contacting OCR. Store originals privately in environment-separated S3, then pass them to the OCR platform only through its documented private-upload flow. After OCR completes, persist the extracted text and send only that text plus the minimal wallet/category catalog needed for structured extraction to the LLM. OCR failure must prevent model invocation and all ledger writes. Distinguish local validation/configuration errors from upstream OCR/model failures in status/code; do not claim an upstream request may have been billed when validation proves no provider call occurred. Provider errors remain redacted and are not automatically retried.

## Data and lifecycle

Persist one processing record per idempotency key and owner, containing status and result/proposal association, not a sequence of messages. Pending proposals are independently editable/removable and approvable. Approval is idempotent and creates exactly one ordinary ledger transaction. Attachments stay private and are linked only on approval per AI-ENTRY-02's attachment lifecycle and environment-separated S3 key contract. This spec does not authorize public storage or automatic ledger writes.

## UI contract

Keep the approved one-shot composer and result-card layout. Submit once; on ambiguous transport failure, offer only a read-only result refresh by process ID/idempotency key, never auto-resubmit. Show validated PDF support and per-file size limits. Keep each proposal editable and independently saveable/removable, plus save-all where current UI provides it. Do not display chat bubbles, message history, or “new conversation.” Reuse the existing shared assistant and transaction form bases.

## Acceptance and regression tests

1. Frontend sends one request to `/ai/entry/process`, with text/files and an idempotency key; no session route is called. A read-only `GET /requests/{request_id}` recovers the result after an ambiguous timeout without repeating provider work.
2. A valid PDF fixture passes server validation and private OCR upload; a PDF MIME spoof, unsupported file, >5 MiB file, or more than three files is rejected before any provider request.
3. For PDF and image inputs, captured LLM HTTP requests contain extracted OCR text but no file bytes, base64, file URL, or image content. OCR failure does not call the model; extraction failure returns no proposals/ledger writes and is not retried automatically.
4. Replaying an identical request key returns the same process result without a second OCR/model submission; changed content with the same key conflicts.
5. Process reads, proposal edits, rejects, and approvals are owner-scoped. Approval replay creates one ledger row and attachment links only for the approved proposal.
6. Existing manual transaction entry remains unaffected. Frontend design guards/tests/build and backend tests pass.

## Boundaries and dependencies

This is a high-risk API/data-flow change within approved backlog item [AI-ENTRY-02](../../work/tickets/AI-ENTRY-02-BATCH-ATTACHMENTS-DETAIL_DESIGN.md). Preserve its private S3 and attachment approval-link requirements. Before implementation, verify the OCR private-upload contract and current config against [OCR_API.md](../../architecture/OCR_API.md); do not assume base64 PDF support. Update the ticket, implementation plan, API docs, OpenAPI output, validation matrix, context, backlog, changelog, and assistant screen contract with the implemented route names.

## Trace links

- Backlog: [BACKLOG.md](../../work/BACKLOG.md#L136)
- Work item/detail design: [AI-ENTRY-02](../../work/tickets/AI-ENTRY-02-BATCH-ATTACHMENTS-DETAIL_DESIGN.md)
- Implementation plan: [AI batch attachments plan](../plans/2026-09-21-ai-batch-attachments.md)
- Tests: backend AI infrastructure/usecase/controller tests; frontend AI entry unit/browser tests.
- Validation: [VALIDATION_MATRIX.md](../../work/VALIDATION_MATRIX.md)
- Master docs to reconcile: API/OpenAPI, [OCR_API.md](../../architecture/OCR_API.md), assistant screen contract, ADR-006, [CONTEXT.md](../../CONTEXT.md), [CHANGELOG.md](../../releases/CHANGELOG.md).
