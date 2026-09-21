# AI Batch Attachments Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace visible AI conversation entry with a one-shot batch review and retain approved transaction evidence in environment-separated private S3 storage.

**Architecture:** The AI batch use case coordinates validated upload, private S3 storage, OCR, text-only extraction and durable review proposals. A dedicated attachment storage port isolates S3 compatibility; proposal approval creates the ledger row and links already-persisted attachments within the database transaction.

**Tech Stack:** Go/Gin/GORM/PostgreSQL, S3-compatible object storage, React/Vite existing atomic bases, current OCR and OpenAI-compatible adapters.

**Spec:** [AI-ENTRY-02-BATCH-ATTACHMENTS-DETAIL_DESIGN.md](../../work/tickets/AI-ENTRY-02-BATCH-ATTACHMENTS-DETAIL_DESIGN.md)

## Global Constraints

- Files are private, server validated and stored at `<S3_PREFIX>/<APP_ENV>/owners/<owner-id>/batches/<batch-id>/<attachment-id>/<safe-filename>`.
- Model receives OCR text only; no model write tools or public object URLs.
- One batch has one extraction; a proposal writes a ledger row only after owner approval.
- Reuse shared transaction form bases and preserve normal Add/manual entry.
- Live S3 test is pending `S3_BUCKET`; fake-S3 and HTTP-provider tests are mandatory.

## Review Focus

1. PDF MIME spoofing and oversized documents must be rejected before S3 upload.
2. An approved row must link its evidence exactly once under approval replay/concurrency.
3. Unapproved evidence must never be downloadable as a transaction attachment.
4. `APP_ENV` and filenames must not permit cross-environment or path-traversal object keys.
5. OCR/model failure after S3 upload must preserve a recoverable pending batch without a ledger write.

### Task 1: Storage configuration and private S3 adapter

**Files:** config tests, `infrastructure/storage/s3.go`, adapter tests, API composition.

- [ ] Write failing tests for required bucket outside development, normalized environment-qualified keys, traversal rejection, put/delete/presign HTTP behavior.
- [ ] Implement `AttachmentStorage.Put/Delete/SignedGet` and S3 SigV4 adapter with path-style option; never log credentials or signed URLs.
- [ ] Wire storage from config and run adapter/config tests.

### Task 2: Attachment model, migration and approval linking

**Files:** attachment entity/repository, migration 000011, AI entry repository/use-case tests.

- [ ] Write failing PostgreSQL tests for owner scope, pending attachment lifecycle, atomic approval link and replay/concurrent approval.
- [ ] Add attachment metadata/link tables, repository methods and transactional link during approval.
- [ ] Run migration and real PostgreSQL tests.

### Task 3: One-shot batch submission with JPEG/PNG/PDF OCR

**Files:** AI entry handler/usecase, OCR adapter, HTTP tests.

- [ ] Write failing multipart tests for text-only, PDF success, invalid MIME/size, OCR failure and idempotent repeated submit.
- [ ] Implement batch endpoints, byte validation, storage compensation, OCR source handling and one-submission guard. Support PDF only through OCR; model remains text-only.
- [ ] Run provider-boundary and handler tests using the supplied PDF as an explicit local manual fixture without committing it.

### Task 4: Transaction attachment download and cleanup

**Files:** transaction attachment handler/usecase, cleanup command, repository tests.

- [ ] Write failing tests for owner-only signed download, not-linked rejection and pending cleanup deleting only expired unlinked objects.
- [ ] Implement authenticated download and explicit cleanup command with delete-failure lifecycle state.
- [ ] Run integration tests and verify no storage key leaks in API responses.

### Task 5: Frontend one-shot batch review

**Files:** `AiEntrySheet`, AI service/types, file validation, assistant screen spec/tests.

- [ ] Write browser/unit regressions proving no conversation/history/new-conversation UI; long press opens one-shot input and submit shows shared editable proposal rows.
- [ ] Switch API client to batches and allow PDF selection; use existing file-upload and proposal/form bases without visual overrides.
- [ ] Run design guard, AI/browser tests and production build.

### Task 6: Reconciliation and acceptance

**Files:** API/ERD/architecture/integrations/OCR docs, ADR, ticket/validation/context/backlog/changelog.

- [ ] Regenerate Swagger and record exact backend/frontend/migration commands.
- [ ] Update master docs and screen contract to match actual endpoints, lifecycle and environment path.
- [ ] Record live OCR/S3 PDF UAT as pending until bucket/model configuration is present; do not claim it passed from fakes.

## Self-review

- Spec coverage: all accepted requirements map to Tasks 1–6.
- Placeholder scan: no implementation TBD is used; external live-bucket input is explicitly a verification blocker.
- Type consistency: storage interface is internal; HTTP exposes attachment metadata without object keys.
- Review focus is covered by Tasks 1–4 tests.
