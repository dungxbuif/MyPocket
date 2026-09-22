---
artifact_type: adr
id: ADR-006
status: accepted
owner: shared
date: 2026-09-22
---

# Private AI receipt attachment lifecycle

## Context

AI entry accepts user receipts for OCR, but originals must not be sent to the LLM or exposed through public storage. OCR needs retrievable source bytes, and a receipt becomes a transaction attachment only after explicit approval. The API is one-shot and does not persist conversation history. See [AI-ENTRY-02](../work/tickets/AI-ENTRY-02-BATCH-ATTACHMENTS-DETAIL_DESIGN.md), [stateless AI design](../superpowers/specs/2026-09-21-stateless-ai-entry-design.md), [API](../architecture/API.md), and [ERD](../architecture/ERD.md).

## Decision

The backend validates uploaded bytes and stores originals in private environment-qualified S3. It sends validated image bytes to OCR as Base64 and uploads PDFs through OCR Platform's private presign flow; a MyPocket signed URL is not sent to OCR because it is not an OCR Platform-owned `s3://` source. The LLM receives only user text, OCR text and bounded reference catalogs. PostgreSQL stores owner/process/file metadata and OCR text. Each explicit proposal approval creates a ledger row and links the process attachments in the same transaction. Download requires both transaction ownership and an attachment link, then redirects to a five-minute signed URL. Expired, unlinked objects are intended to be removed by an explicit cleanup command; failed deletion retains metadata for retry.

Environment is part of every object key under a neutral configured prefix. S3 credentials remain backend-only; outside development, API startup requires complete storage configuration. No public ACLs or direct browser-to-S3 uploads are allowed.

## Alternatives

- OCR the transient browser upload without retaining it: rejected because receipts would be lost after approval and ambiguously failed batches would not be recoverable.
- Store objects publicly or expose object keys: rejected because receipts contain financial/private information.
- Send file/base64/image content directly to the LLM: rejected; OCR is the only file-reading boundary.
- Persist files only after approval: rejected because the OCR service needs temporary access before the owner can review proposals.

## Consequences

Private object storage and a relational link table add lifecycle/configuration burden. Pending files require running the explicit cleanup command; it is not scheduled automatically. Live provider and bucket UAT remains separate from fake tests. Approval replay is protected by existing proposal locking and unique attachment links. `ai_entry_sessions` remains an internal legacy-named process table pending a separate cleanup migration.

## Trace and verification

Backlog: [BACKLOG](../work/BACKLOG.md) · Validation: [VALIDATION_MATRIX](../work/VALIDATION_MATRIX.md) · Release: [CHANGELOG](../releases/CHANGELOG.md).

The code has fake S3/use-case/OCR boundary tests and explicit migration `000011`; real database migration application, scheduled/CLI cleanup and authenticated live S3/OCR UAT remain open.
