# PHASE-006 AI, Receipt, and Bank Ingestion Implementation Plan

**Goal:** Route text AI, receipt OCR, multimodal image chat, and bank webhooks into one review-first draft contract that never confirms accounting automatically.

**Spec:** `docs/work/phases/PHASE-006-detail-design.md`

## Global Constraints

- Every shell command starts with `rtk`.
- Treat prompts, OCR text, image content, and webhook payloads as hostile.
- Provider confidence is informational only.
- Draft confirmation is the only path to PHASE-002 accounting and requires auth, CSRF, idempotency, and revalidation.
- Logs and audit records redact secrets, raw provider payloads, raw notification text, image bytes, tokens, and credentials.

## Tasks

- [ ] **Task 1: Shared Draft Contract and Confirmation**
  - Files: draft migration, `backend/internal/ingestion`, draft HTTP routes, frontend draft review UI.
  - Implement draft schema, edit/reject/confirm, validation issues, idempotent accounting handoff.
  - Verify with: `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/ingestion -run Draft -count=1`

- [ ] **Task 2: Text AI Chat**
  - Files: OpenAI-compatible provider adapter, AI chat routes, mobile proposal cards.
  - Implement structured output parsing, owned wallet/category context, single/multi/transfer proposals, failure mapping.
  - Verify with: `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/providers ./internal/ingestion -run TextAI -count=1`

- [ ] **Task 3: Receipt Capture and OCR**
  - Files: receipt upload/extract routes, S3 integration, OCR adapter, quick-add camera/upload UI.
  - Implement private upload, OCR extraction, draft creation, retry/failure UI.
  - Verify with: `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/providers ./internal/ingestion -run Receipt -count=1`

- [ ] **Task 4: Multimodal Image Chat**
  - Files: multimodal adapter, image chat route, frontend image proposal flow.
  - Implement private image reference handling, structured multimodal output, draft convergence.
  - Verify with: `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/providers ./internal/ingestion -run Image -count=1`

- [ ] **Task 5: Signed Bank Webhook**
  - Files: webhook routes, HMAC/nonce ledger, redacted audit hooks, inbox notice integration.
  - Implement raw-body HMAC, timestamp/nonce replay protection, source-to-user mapping, draft/notice creation.
  - Verify with: `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/ingestion -run Webhook -count=1`

- [ ] **Task 6: Reconciliation and Release Proof**
  - Run: `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test -p 1 ./... -count=1`
  - Run: `rtk npm test -- --run`
  - Run: `rtk npm run build`
  - Run: `rtk npm run test:e2e -- ingestion.spec.ts`
  - Update verification, tickets, phase, validation matrix, backlog, context, changelog, API, ERD, integrations, architecture.
