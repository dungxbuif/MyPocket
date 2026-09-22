---
artifact_type: detail_design
id: AI-ENTRY-02
status: in_progress
owner: shared
updated: 2026-09-22
approval: approved_by_owner_in_conversation
human_fields: [approval, priority, product_decisions]
ai_fields: [approach, impacted_scope, verification, reconciliation]
shared_fields: [status, trace]
trace:
  feedback: [FB-003]
  backlog: AI-ENTRY-02
  parent: TICKET-09-01
  adr: ADR-006
---

# AI batch entry and private transaction attachments

## Decision received

The long-press Add entry must not expose a conversation or resume chat. It accepts one text/file submission, then displays a list of independently editable transaction drafts using the same fields and bases as manual Add. Each approved draft writes one ordinary ledger transaction and links the submitted files as private attachments. The owner also requires an S3 key path that separates environments.

UI reference: follow the interaction layout in the [Money Lover AI entry guide](https://moneylover.zendesk.com/hc/en-us/articles/42320025248409-Add-transactions-faster-and-easier-with-AI-feature): natural-language composer with file/send affordances and a grouped recognized-transactions result. MyPocket intentionally keeps its approved review step: editable proposal cards, remove an item, save one item, or save all. Do not copy Money Lover's immediate auto-save or persistent chat history. Shared bases are `BaseTextArea`, `BaseFileUpload`, `AssistantComposer`, `AssistantResultCard`, `EntryProposalRow` and `TransactionFields`; the screen contract maps each region in [assistant README](../../design/screens/assistant/README.md).

## Scope and boundaries

- Keep the existing private, owner-scoped processing record internally for idempotency and draft approval, but remove conversation/history/new-conversation UI and wording. It is an implementation detail named a batch at the HTTP/UI boundary.
- Submit text plus up to three JPEG, PNG, or PDF files. Store immutable originals in configured S3 before OCR; OCR receives a short-lived presigned read URL or supported provider source URL, never a public object URL.
- Create attachment metadata as `pending` on submission. On each draft approval, atomically link all batch attachments to that created transaction. Attachment bytes are not copied per draft.
- A rejected draft never gets an attachment link. A batch with no approved drafts has pending attachments; the explicit cleanup command deletes only expired unlinked attachments after 24 hours. It is not scheduled automatically.
- Attachment download is an authenticated API that verifies transaction ownership and returns a short-lived S3 GET URL. Keys, bucket name and provider URLs never reach the browser in normal transaction payloads.
- Object key format is `<S3_PREFIX>/<APP_ENV>/owners/<owner-id>/batches/<batch-id>/<attachment-id>/<safe-filename>`. `APP_ENV` is mandatory outside development; slash/path traversal, control characters and filename ambiguity are rejected. The configured prefix is a neutral root, never itself an environment selector.
- No S3 public ACL, no model image input, no OCR-to-ledger write, no attachment linking before explicit approval, and no internal-transfer support in this slice.

## Architecture

`FE one-shot input -> authenticated batch API -> validate bytes -> S3 private put -> OCR text -> model extract -> persistent proposals -> edit/approve rows -> DB transaction creates ledger row + attachment links`.

The use case owns sequencing and cleanup compensation. A narrow `AttachmentStorage` port owns S3 put, delete and signed read URL; the S3 adapter owns endpoint/path-style configuration. The repository owns metadata and row locking. The model continues to receive OCR text only. Existing session tables may remain for compatibility; HTTP resource names and UI copy become `batches`, and one batch accepts exactly one extraction request.

## Impacted brownfield scope

Backend: AI entry entity/use case/repository/handler, OCR adapter, API composition/config, migrations, S3 infrastructure, transaction attachment read endpoint. Frontend: `AiEntrySheet`, AI service/types, file input validation and proposal-row composition; reuse manual transaction field bases, no screen-specific controls. Docs: API, ERD, architecture, integrations/OCR, assistant screen contract, ADR, validation, context/backlog/changelog.

Direct dependencies inspected: existing owner-scoped process/proposal persistence, transaction creation, OCR image/PDF adapters, S3-compatible adapter/config, `BaseFileUpload`, `EntryProposalRow`. The API/UI stateless contract, OCR-before-LLM ordering, private S3 application wiring, attachment metadata/migrations 000011–000012, atomic approval links, authenticated download and explicit cleanup command are implemented in code as of 2026-09-22. Fake/unit/provider and local PostgreSQL proofs pass. Remaining scope: owner UAT against live S3/OCR and browser download. Never auto-retry a previously submitted document.

## API/data contract

- `POST /api/v1/ai/entry/process`: one authenticated multipart request containing `request_id`, `text`, `timezone`, and up to three `files`; it performs validation, private storage, OCR, text-only LLM extraction, and returns the reviewable result. No separate create-session/submit calls.
- `GET /api/v1/ai/entry/requests/:request_id`: owner-scoped read-only processing state and proposals for recovery; never starts provider work.
- Existing proposal edit/approve/reject endpoints remain owner-scoped. Approval returns attachment metadata IDs and transaction ID.
- `GET /api/v1/transactions/:id/attachments/:attachmentId/download`: owner-scoped short-lived download redirect/URL.

The existing migration-000010 session-named table may remain as an internal compatibility store for process/idempotency state, but the API and returned proposal JSON use process/batch terminology. The one-shot path does not write/load conversation messages or send history to the LLM. Browser uploads are raw multipart files; only the backend constructs a transient OCR request. The LLM receives only OCR text, user-entered text, and the minimal reference catalog; never file bytes, base64, signed/public URLs or image input.

Tables: `transaction_attachments` (owner, batch, object key, original filename, MIME, bytes, sha256, OCR status/text, deletion lifecycle); `transaction_attachment_links` (transaction, attachment), unique on both columns. Do not delete DB attachment metadata when storage deletion fails; mark `delete_failed` for retry.

## Acceptance and verification

1. Holding Add shows a one-shot input, not a conversation; submit renders multiple editable proposal rows built from shared transaction bases.
2. A PDF fixture uploads under the environment-qualified private key, OCR result is persisted, and the model receives text only.
3. Approving one of two rows creates only that ledger entry and links attachments only to it; replay does not duplicate the row or link.
4. Cross-owner batch, proposal, attachment download and object-key traversal attempts fail; bucket/key never appear in normal client payloads.
5. Fake S3 tests prove put/delete/presign behavior and configuration/key construction; live S3/OCR and browser attachment UAT remain owner-facing acceptance.

Proof run: migrations 000011–000012 applied to local dev PostgreSQL (version 12, clean); `TEST_DATABASE_URL=... go test ./... -count=1` passed, including real owner-scoped approval/link/download-query and cleanup-claim tests. The 2026-09-22 live synthetic-image evaluation is blocked by [BUG-001](../bugs/BUG-001-s3-presigned-get-signature.md): S3 PUT succeeds but signed GET returns `SignatureDoesNotMatch`, before OCR/LLM. The user's previous PDF was not retried.

## Live evaluation attempt — 2026-09-22

- Added opt-in `TestLiveQwenImageAndTransactionEval` using a synthetic receipt only; it intends to put/read one private S3 object, send its signed URL to OCR, then score five synthetic Qwen extraction cases and report latency/field accuracy. It never approves proposals or writes ledger transactions.
- The owner approved AWS SDK for Go v2. The S3 adapter now uses its S3 presign client for signed GET; PUT/DELETE remain unchanged. Focused and full Go tests pass.
- Live synthetic image test passes private S3 PUT/readback (1,027,882 bytes; 62 ms) and OCR (3.815 s; mock total recognized). In the captured five-case model rerun, all calls failed strict draft-schema validation: 0/15 fields, LLM p50 15.089 s/p95 21.7 s, transfer-safety gate failed. The OCR + receipt extraction pipeline measured 25.516 s. S3 is verified; AI extraction compatibility is tracked separately under AI-ENTRY-01. Cleanup finalization returns conflict if the attachment is no longer in the claimed `deleting` state, so a stale cleanup worker cannot silently mark an unclaimed file deleted. See [BUG-001](../bugs/BUG-001-s3-presigned-get-signature.md), [ADR-007](../../decisions/ADR-007-aws-s3-presigning.md) and [validation evidence](../VALIDATION_MATRIX.md#ai-entry-02-live-provider-evaluation-2026-09-22).

- Follow-up after [AI-ENTRY-01](TICKET-09-01-ENTRY-DETAIL_DESIGN.md): the signed GET issue is fixed; live S3 readback and OCR now pass, and strict JSON Schema LLM evaluation passes 5/5 synthetic cases. Latest timings and scores are recorded in the [validation matrix](../VALIDATION_MATRIX.md#ai-entry-01-openai-compatible-json-schema-resolution--2026-09-22). Browser attachment/download owner UAT remains pending.

## Alternatives rejected

- Browser direct-to-S3 upload: would expose credentials/policy complexity and bypass server validation.
- Persisting attachments only after approval: OCR cannot reliably access private bytes and failure/retry becomes opaque.
- Public S3 URL for OCR/download: violates receipt privacy.
- A full chat/session redesign: the product request is a one-shot batch, not conversational entry.

## Reconciliation plan

Update API/ERD/architecture/integration/OCR contracts, ADR-006/007, assistant screen behavior spec, validation matrix, context, backlog and changelog. No new AI advice/tool-call behavior is included.
