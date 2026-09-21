---
artifact_type: detail_design
id: AI-ENTRY-02
status: ready
owner: shared
updated: 2026-09-21
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

## Scope and boundaries

- Keep the existing private, owner-scoped processing record internally for idempotency and draft approval, but remove conversation/history/new-conversation UI and wording. It is an implementation detail named a batch at the HTTP/UI boundary.
- Submit text plus up to three JPEG, PNG, or PDF files. Store immutable originals in configured S3 before OCR; OCR receives a short-lived presigned read URL or supported provider source URL, never a public object URL.
- Create attachment metadata as `pending` on submission. On each draft approval, atomically link all batch attachments to that created transaction. Attachment bytes are not copied per draft.
- A rejected draft never gets an attachment link. A batch with no approved drafts has pending attachments; a scheduled cleanup deletes only unlinked pending attachments after 24 hours. The first slice provides the cleanup use case/command, not a background scheduler.
- Attachment download is an authenticated API that verifies transaction ownership and returns a short-lived S3 GET URL. Keys, bucket name and provider URLs never reach the browser in normal transaction payloads.
- Object key format is `<S3_PREFIX>/<APP_ENV>/owners/<owner-id>/batches/<batch-id>/<attachment-id>/<safe-filename>`. `APP_ENV` is mandatory outside development; slash/path traversal, control characters and filename ambiguity are rejected. The configured prefix is a neutral root, never itself an environment selector.
- No S3 public ACL, no model image input, no OCR-to-ledger write, no attachment linking before explicit approval, and no internal-transfer support in this slice.

## Architecture

`FE one-shot input -> authenticated batch API -> validate bytes -> S3 private put -> OCR text -> model extract -> persistent proposals -> edit/approve rows -> DB transaction creates ledger row + attachment links`.

The use case owns sequencing and cleanup compensation. A narrow `AttachmentStorage` port owns S3 put, delete and signed read URL; the S3 adapter owns endpoint/path-style configuration. The repository owns metadata and row locking. The model continues to receive OCR text only. Existing session tables may remain for compatibility; HTTP resource names and UI copy become `batches`, and one batch accepts exactly one extraction request.

## Impacted brownfield scope

Backend: AI entry entity/use case/repository/handler, OCR adapter, API composition/config, migrations, S3 infrastructure, transaction attachment read endpoint. Frontend: `AiEntrySheet`, AI service/types, file input validation and proposal-row composition; reuse manual transaction field bases, no screen-specific controls. Docs: API, ERD, architecture, integrations/OCR, assistant screen contract, ADR, validation, context/backlog/changelog.

Direct dependencies inspected: existing AI session/proposal persistence, transaction creation, OCR base64 client, S3 config placeholders, `BaseFileUpload`, `EntryProposalRow`. Known unknown: production S3 bucket name/region are absent. Implementation must have fake-S3 tests; a live bucket test is explicitly pending that configuration.

## API/data contract

- `POST /api/v1/ai/entry/batches`: create one owner batch.
- `POST /api/v1/ai/entry/batches/:id/submit`: multipart `text`, `timezone`, and `files[]`; one submission only, idempotency key required.
- `GET /api/v1/ai/entry/batches/:id`: owner-scoped batch and proposals; no message history.
- Existing proposal edit/approve/reject endpoints remain owner-scoped. Approval returns attachment metadata IDs and transaction ID.
- `GET /api/v1/transactions/:id/attachments/:attachmentId/download`: owner-scoped short-lived download redirect/URL.

Tables: `transaction_attachments` (owner, batch, object key, original filename, MIME, bytes, sha256, OCR document ID/status/text, lifecycle); `transaction_attachment_links` (transaction, attachment), unique on both columns. Do not delete DB attachment metadata when storage deletion fails; mark `delete_failed` for retry.

## Acceptance and verification

1. Holding Add shows a one-shot input, not a conversation; submit renders multiple editable proposal rows built from shared transaction bases.
2. A PDF fixture uploads under the environment-qualified private key, OCR result is persisted, and the model receives text only.
3. Approving one of two rows creates only that ledger entry and links attachments only to it; replay does not duplicate the row or link.
4. Cross-owner batch, proposal, attachment download and object-key traversal attempts fail; bucket/key never appear in normal client payloads.
5. Fake S3 tests prove put/delete/presign behavior and configuration/key construction; real S3 OCR PDF UAT is recorded pending `S3_BUCKET` and model credentials.

## Alternatives rejected

- Browser direct-to-S3 upload: would expose credentials/policy complexity and bypass server validation.
- Persisting attachments only after approval: OCR cannot reliably access private bytes and failure/retry becomes opaque.
- Public S3 URL for OCR/download: violates receipt privacy.
- A full chat/session redesign: the product request is a one-shot batch, not conversational entry.

## Reconciliation plan

Update API/ERD/architecture/integration/OCR contracts and add an ADR for durable private attachment ownership and environment key isolation. Update the assistant screen behavior spec and validation matrix. No new AI advice/tool-call behavior is included.
