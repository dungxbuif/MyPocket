# AI Entry Deterministic Financial Ingestion Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the supplied XLSX statements exact and reviewable while retaining `DeepSeek-V4.1-Flash` only for semantic enrichment.

**Architecture:** Parse structured statements into canonical source rows before provider work. Send opaque row IDs and bounded row text to the model, then merge model semantics onto parser-owned amount/date/type/currency fields with fail-closed source-ID validation.

**Tech Stack:** Go 1.26 standard-library OOXML parsing (`archive/zip`, `encoding/xml`), existing Gin multipart API, existing OpenAI-compatible adapter, PostgreSQL review proposals.

**Spec:** `docs/superpowers/specs/2026-10-26-ai-entry-deterministic-financial-ingestion-design.md`

## Global Constraints

- No automatic approval or ledger write.
- No raw source rows, account numbers, prompts or credentials in logs or git.
- Preserve existing request idempotency and private attachment storage.
- Existing dirty files belong to concurrent work; do not revert them.
- No production deployment in this plan.

## Review Focus

- ZIP/XLSX MIME spoofing must fail before storage/provider work.
- Debit and credit both populated must fail closed, never guess direction.
- Two identical source rows must keep two distinct source IDs.
- Model output with missing/duplicate/unknown source IDs must be rejected.
- Formula/external-link/macro workbooks must not execute active content.

---

### Task 1: Canonical workbook parser

**Files:**
- Create: `backend/internal/infrastructure/financialsource/xlsx.go`
- Create: `backend/internal/infrastructure/financialsource/xlsx_test.go`

**Interfaces:**
- Produces: `ParseWorkbook(name string, data []byte, timezone string) (Document, error)` and canonical `Row` values.

- [x] Write RED tests using minimal in-memory workbooks for Techcombank headers, separate identical rows, invalid debit+credit, decimal/non-VND values and unsupported layouts.
- [x] Run `go test ./internal/infrastructure/financialsource -count=1` and confirm RED.
- [x] Implement the required OOXML subset with archive/member/row/cell bounds, byte/MIME validation and stable sheet/row source IDs; reject macros, external links and formulas.
- [x] Run the focused tests and confirm GREEN.

### Task 2: Multipart and private attachment acceptance

**Files:**
- Modify: `backend/internal/controller/http/ai_entry_handler.go`
- Modify: `backend/internal/controller/http/ai_entry_handler_test.go`
- Modify: `backend/internal/controller/http/mcp_operation_catalog.go`
- Modify: `backend/internal/controller/http/mcp_operation_catalog_test.go`
- Modify: `backend/internal/usecase/ai_entry.go`
- Modify: `backend/internal/usecase/ai_entry_test.go`

**Interfaces:**
- Consumes: `financialsource.ParseWorkbook`.
- Produces: validated XLSX `AIImage` attachments and canonical structured rows before provider extraction.

- [x] Add RED handler/catalog tests for a real XLSX signature and spoofed ZIP/filename cases.
- [x] Add RED use-case tests proving XLSX does not require OCR and remains privately stored.
- [x] Extend MIME validation, storage extension mapping and structured parsing without weakening JPEG/PNG/PDF checks.
- [x] Run focused HTTP/use-case tests and confirm GREEN.

### Task 3: Source-bound semantic extraction

**Files:**
- Modify: `backend/internal/entity/ai_entry.go`
- Modify: `backend/internal/infrastructure/ai/client.go`
- Modify: `backend/internal/infrastructure/ai/schema.go`
- Modify: `backend/internal/infrastructure/ai/extract.v1.txt`
- Modify: `backend/internal/infrastructure/ai/client_test.go`
- Modify: `backend/internal/usecase/ai_entry.go`
- Modify: `backend/internal/usecase/ai_entry_test.go`

**Interfaces:**
- Consumes: canonical structured rows.
- Produces: one semantic response per source ID with deterministic financial-field overlay.

- [x] Add RED schema/use-case tests for changed amount/date/type, missing IDs, duplicate IDs and preserved identical rows.
- [x] Add structured row blocks to the prompt and require `source_id` for structured extraction.
- [x] Merge only category/note/questions/wallet semantics; overwrite parser-owned fields and reject source-set mismatch.
- [x] Bound structured batches and aggregate usage/latency without changing request idempotency.
- [x] Run focused AI and use-case tests and confirm GREEN.

### Task 4: Three-source evaluation harness

**Files:**
- Create: `backend/internal/infrastructure/ai/financial_ingestion_eval_test.go`
- Modify: `docs/work/tickets/AI-ENTRY-THREE-FILE-BENCH-20261026.md`
- Modify: `docs/work/VALIDATION_MATRIX.md`

**Interfaces:**
- Consumes: supplied private files via opt-in environment paths.
- Produces: aggregate counts, exact structured-field score, semantic review score and stage latency; never persists raw rows.

- [x] Add an opt-in test that parses both XLSX files and validates source counts without calling production.
- [x] Add review-only configured-provider evaluation with zero approvals and privacy-safe aggregate output.
- [x] Compare against the 55/87 and 110.97-second baseline.

### Task 5: Contract and documentation reconciliation

**Files:**
- Modify: `docs/architecture/API.md`
- Modify: `docs/architecture/INTEGRATIONS.md`
- Modify: `docs/guides/ai-entry.md`
- Modify: `docs/work/BACKLOG.md`
- Modify: `docs/CONTEXT.md`
- Modify: `docs/releases/CHANGELOG.md`
- Modify: generated/public API surfaces when required by repository commands.

- [x] Document XLSX limits, deterministic ownership and remaining PDF/OCR limitations.
- [x] Run focused tests, full backend tests/race checks required by the validation standard, documentation link checks and `git diff --check`.
- [x] Record exact evidence and leave production rollout explicitly open.
