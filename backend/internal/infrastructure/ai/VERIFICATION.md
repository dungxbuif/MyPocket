# AI entry provider adapter handoff

Scope: this directory only. Design: [approved entry slice](../../../../docs/work/tickets/TICKET-09-01-ENTRY-DETAIL_DESIGN.md). OCR contract: [OCR API](../../../../docs/architecture/OCR_API.md).

`NewClient(Config)` exposes `Configured`, `OCRConfigured`, and `Extract`. Input/output/draft/image/history types alias the coordinator-owned entities. Provider message JSON uses a private wire type so it does not depend on entity JSON tags. There are no persistence operations or provider tools.

Limits: 32 KiB input text; latest 12 history messages, each at most 8 KiB; 128 KiB serialized message context; 64 KiB combined source; 256 KiB provider responses; no application draft-count cap; nonnegative integer amounts through 9007199254740991. Catalogs project reference fields only, without owner IDs or balances. IDs and date/business validity remain coordinator validation responsibilities; the adapter never generates replacement IDs.

Images: up to twenty JPEG/PNG/PDF files, each at most 5 MiB decoded and 25 million pixels where applicable. Every image is validated before any submission. OCR uses `input.base64`, then polls the returned bounded document ID under the 210-second extraction context. Pending polling respects Retry-After and cancellation. Each HTTP operation is bounded to 30 seconds; the model has an explicit 180-second request timeout and the extraction context is 210 seconds. Redirects are rejected. There is no application retry or ambiguous-submission resubmission. Failure of any image prevents model extraction. Successful OCR source remains available in `Output.SourceText` if the subsequent model fails.

Schema validation rejects unknown/missing fields, null required fields, duplicate JSON keys, trailing content, invalid types and unsafe amounts. Empty drafts with clarification are accepted. Provider diagnostics, URLs and credentials are never propagated as errors or logged. The adapter emits structured, privacy-safe diagnostics for OCR/model start, completion and failure (stage, bounded status/code, latency, byte counts and counts only). When `AI_STORE_USAGE=true`, normalized provider request IDs, finish reason, latency and token usage are returned to the coordinator for internal persistence in `ai_entry_sessions`/`advisor_runs`; raw prompts, images, OCR text and credentials are never logged or returned in public responses. `user_instruction` is sent separately from untrusted OCR/source text so date filters such as “chỉ lấy giao dịch tháng 9” are applied as selection guidance without allowing source text to override schema or ownership rules.

## Verification — 2026-09-21

Commands run in `backend/` (through the required `rtk proxy`):

- `go test ./internal/infrastructure/ai`: initial RED, adapter symbols absent.
- `go test ./internal/infrastructure/ai -run TestTextOnlyRequestAndUnverifiedIDs -count=1 -timeout=15s`: RED, shared history entities serialized uppercase provider keys. Fixed with the private wire type.
- Initial all-adapter run exposed a timeout fixture cleanup hang because its handler waited on a request context without consuming the body; fixture now drains the body and has an explicit cleanup release.
- `go test ./internal/infrastructure/ai -count=1 -timeout=20s`: PASS.
- `go test -race ./internal/infrastructure/ai -count=1 -cover -timeout=45s`: PASS, 84.5% statement coverage.
- `go vet ./internal/infrastructure/ai`: PASS.
- `go test ./... -count=1`: adapter, HTTP controller, entity and infrastructure repository tests PASS; suite FAIL because concurrent coordinator tests in `internal/usecase/ai_entry_test.go` referenced not-yet-defined `AIEntryService`, `AIEntryMessageInput`, `ErrAIUnavailable`, and `ErrAIProvider`.

Tests cover text-only requests, exact HTTP paths/auth separation, shared aliases, history bounds, catalog minimization, strict schema, transfer/unknown IDs preserved for coordinator validation, zero-draft clarification, OCR-before-model, twenty-file input validation, wrong MIME/content/pixel/size/count, partial failures, cancelled/failed/expired OCR, source preservation on model failure, upstream error redaction, redirect/no-resubmission, response bounds and caller timeout/cancellation, separated user instruction and usage normalization, and acceptance of more than thirty drafts.

No live provider call, credentials, commit, app change or shared-doc reconciliation performed by this adapter task. End-to-end persistence/UAT and shared status/validation/release docs belong to the coordinator. Live provider behavior remains unverified; all HTTP evidence uses local `httptest` servers and synthetic keys.

Coordinator follow-up: the previously unfinished usecase symbols are now implemented. Full `go test -race ./... -count=1` with `TEST_DATABASE_URL` passes after independent review fixes; see the approved entry slice for combined HTTP/PostgreSQL/frontend evidence and remaining live-provider limitations.
