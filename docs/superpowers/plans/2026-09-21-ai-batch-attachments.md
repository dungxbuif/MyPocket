# AI Batch Attachments Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the session-shaped AI entry contract with one-shot processing, OCR all uploaded files in backend code before the LLM sees extracted text, and retain approved transaction evidence in environment-separated private S3 storage.

**Architecture:** `POST /api/v1/ai/entry/process` validates and stores a one-shot upload, runs private OCR on every file, then sends only OCR text and the minimal reference catalog to the LLM. It persists an opaque process record for idempotency/proposal lifecycle, not conversation messages. `GET /requests/{request_id}` is read-only recovery; proposal edit/approve/reject remain separate operations. A dedicated attachment storage port isolates S3 compatibility; proposal approval creates the ledger row and links already-persisted attachments within the database transaction.

**Tech Stack:** Go/Gin/GORM/PostgreSQL, S3-compatible object storage, React/Vite existing atomic bases, current OCR and OpenAI-compatible adapters.

**Spec:** [Stateless AI Entry and PDF OCR](../specs/2026-09-21-stateless-ai-entry-design.md), with private attachment requirements from [AI-ENTRY-02-BATCH-ATTACHMENTS-DETAIL_DESIGN.md](../../work/tickets/AI-ENTRY-02-BATCH-ATTACHMENTS-DETAIL_DESIGN.md).

## Global Constraints

- Files are private, server validated and stored at `<S3_PREFIX>/<APP_ENV>/owners/<owner-id>/batches/<batch-id>/<attachment-id>/<safe-filename>`.
- Model receives OCR text only; no model write tools or public object URLs.
- LLM request must never contain original bytes, base64, file URLs (signed or public), or image/multimodal file content.
- OCR validation/upload/extraction completes before model invocation; any OCR failure prevents the LLM call and ledger writes.
- `POST /ai/entry/process` performs one extraction; `GET /ai/entry/requests/{request_id}` only reads existing state. Never expose session/message/history routes to the frontend.
- One batch has one extraction; a proposal writes a ledger row only after owner approval.
- Reuse shared transaction form bases and preserve normal Add/manual entry.
- Fake-S3 and HTTP-provider tests are mandatory. Live S3/OCR PDF UAT requires an explicit owner-triggered test sample; do not automatically reuse an earlier user document.

## Review Focus

1. PDF MIME spoofing and oversized documents must be rejected before S3 upload.
2. An approved row must link its evidence exactly once under approval replay/concurrency.
3. Unapproved evidence must never be downloadable as a transaction attachment.
4. `APP_ENV` and filenames must not permit cross-environment or path-traversal object keys.
5. OCR/model failure after S3 upload must preserve a recoverable pending batch without a ledger write.

### Task 1: Storage configuration and private S3 adapter

**Files:** config tests, `infrastructure/storage/s3.go`, adapter tests, API composition.

- [x] Write failing tests for required bucket outside development, normalized environment-qualified keys, traversal rejection, put/delete/presign HTTP behavior.
- [x] Implement `AttachmentStorage.Put/Delete/SignedGet` and S3 SigV4 adapter with path-style option; never log credentials or signed URLs.
- [x] Wire storage from config and run adapter/config tests.

### Task 2: Attachment model, migration and approval linking

**Files:** attachment entity/repository, migration 000011, AI entry repository/use-case tests.

- [x] Write failing PostgreSQL tests for owner scope, pending attachment lifecycle, atomic approval link and replay/concurrent approval.
- [x] Add attachment metadata/link tables, repository methods and transactional link during approval.
- [x] Run migrations 000011–000012 and real PostgreSQL tests (dev schema version 12, clean).

### Task 3: Stateless one-shot API and OCR-first JPEG/PNG/PDF pipeline

**Files:** AI entry handler/usecase, OCR adapter, HTTP tests.

- [x] Write a failing HTTP regression asserting `POST /api/v1/ai/entry/process` stores one request and returns proposals without creating a chat/message row; assert `GET /requests/{request_id}` is read-only and owner-scoped.
- [x] Write a failing OCR adapter test for private PDF/image source submission and text-only LLM; retain JPEG/PNG coverage. Reject MIME spoof, >5 MiB and >3 files before any provider request.
- [x] Write a failing provider-boundary assertion capturing the LLM request: only recognized OCR text and bounded wallet/category catalog may be present; no bytes, base64, URL, or image field.
- [x] Implement direct process and request-read routes, request-key idempotency, internal process persistence without conversation history, PDF-capable OCR upload, and ordering that prevents the LLM call when OCR fails.
- [x] Run focused OCR/use-case/HTTP tests and frontend request-contract tests. Do not use the user PDF for an automatic/live retry; any live manual test requires explicit user action after code is deployed.

### Task 4: Transaction attachment download and cleanup

**Files:** transaction attachment handler/usecase, cleanup command, repository tests.

- [x] Write failing tests for owner-only signed download, not-linked rejection and pending cleanup deleting only expired unlinked objects.
- [x] Implement authenticated download and explicit cleanup command with delete-failure lifecycle state.
- [x] Run integration tests and verify no storage key leaks in API responses.

### Task 5: Frontend one-shot batch review without session APIs

**Files:** `AiEntrySheet`, AI service/types, file validation, assistant screen spec/tests.

- [x] Write browser/unit regressions proving long press opens one-shot input, submit calls only `/process`, result recovery calls only read-only `/requests/{request_id}`, and no session/message/history API is called.
- [x] Assert PDF selection is sent as a file for server OCR (not converted to image/base64 in the browser); use existing file-upload and proposal/form bases without visual overrides.
- [x] Switch API client to process/request endpoints, support idempotency-key recovery after ambiguous timeout, and show the process result as shared editable proposal rows.
- [x] Run design guard, AI/browser tests and production build.

### Task 6: Reconciliation and acceptance

**Files:** API/ERD/architecture/integrations/OCR docs, ADR, ticket/validation/context/backlog/changelog.

- [x] Regenerate Swagger and record exact backend/frontend/migration commands for process, read-only request recovery, attachment download and proposal endpoints.
- [x] Update master docs and screen contract to remove session/message semantics and state that OCR runs in backend before text-only LLM extraction.
- [x] Record live OCR/S3 PDF UAT as pending; do not claim it passed from fakes.

## Self-review

- Spec coverage: all accepted requirements map to Tasks 1–6.
- Placeholder scan: no implementation TBD is used; external live-bucket input is explicitly a verification blocker.
- Type consistency: storage interface is internal; HTTP exposes attachment metadata without object keys.
- Review focus is covered by Tasks 1–4 tests.
