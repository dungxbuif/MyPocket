---
artifact_type: test_verification
status: in_review
date: 2026-09-12
trace:
  spec: ../../superpowers/specs/2026-09-12-two-chat-design.md
  plan: ../../superpowers/plans/2026-09-12-two-chat-agent-implementation.md
  validation_matrix: ../VALIDATION_MATRIX.md
---

# Agent Two-Chat Verification

## Scope

Implemented the reviewed two-chat Agent contract:

- Transaction intake endpoints create reviewable income/expense drafts only.
- Advisor endpoint is answer-only/read-only and creates no drafts.
- Existing `/api/v1/agent/messages` remains as a compatibility route for legacy explicit `kind` submissions.
- OCR remains an internal third-party image tool for intake; no bank or voice integration was added.
- OCR adapter was reconciled against the separate `mac-ocr` project contract: capability discovery expects `engine=OCR`/`capabilityVersion=ocr-v1.*`, submit sends `input.base64` with normalized `vi-VN`/`en-US` languages, and read parses `documentId`, `result.text`, `result.pageCount`, `result.pages`, `resultExpiresAt`, and `errorDetail`.

## Commands

| Command | Result | Notes |
| --- | --- | --- |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/platform/httpapi -run 'TestAgent(IntakeEndpoint\|AdvisorEndpoint)' -count=1` | fail, expected RED | New endpoint handlers/kinds did not exist yet. |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/platform/httpapi -run 'TestAgent(IntakeEndpoint\|AdvisorEndpoint\|Messages)' -count=1` | pass | New HTTP endpoints and legacy route contract pass. |
| `rtk env 'MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/agent -run 'TestAgent(Intake\|Advisor)' -count=1` | pass | Intake rejects transfer, persists multiple review drafts, advisor creates no drafts. |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/worker -run 'TestAgent' -count=1` | pass | Worker agent tests pass after schema/kind changes. |
| `rtk npm test -- --run src/app/agent.test.ts` | fail, expected RED; then pass | Client tests first failed because split functions did not exist; after implementation 10/10 pass. |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/platform/httpapi -run 'TestOpenAPI\|TestDocs\|TestAgent' -count=1` | pass | OpenAPI route inventory and agent HTTP docs checks pass. |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/platform/ocr -run TestOCRClientContract -count=1` | fail, expected RED | Existing adapter still expected stale `api`/`input`, `document_id`, top-level `text`/`expires_at`. |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/platform/ocr -count=1` | pass | OCR client now matches `mac-ocr` request/response contract and keeps provider errors secret-safe. |
| `rtk env 'MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/agent ./internal/worker ./internal/platform/httpapi -count=1` | fail | Parallel package run collided on the shared test schema reset; rerun serially below. |
| `rtk env 'MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' GOCACHE=/private/tmp/mypocket-go-cache go test -p 1 ./internal/platform/ocr ./internal/agent ./internal/worker ./internal/platform/httpapi -count=1` | pass | Focused backend impacted packages pass with full migration chain when DB-resetting packages run serially. |
| `rtk npm run build` | pass | Frontend typecheck and production Vite build pass. |
| `rtk npm run build` in `frontend/docs` | pass | Public Docusaurus docs compile after Agent/OCR contract updates. |

## Residual Risk

- No production provider/OCR smoke was run in this slice.
- No full frontend build/E2E or physical-device UAT was run in this slice.
- Intake session history is currently run-based and compatibility-preserving; a richer persisted intake session table remains future work if the UI needs multi-turn browsing beyond run polling.
