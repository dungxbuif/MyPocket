---
artifact_type: detail_design
id: PHASE-006-DETAIL-DESIGN
status: approved
owner: shared
approval: approved
approved_on: 2026-08-30
trace:
  backlog_item: BL-006
  phase: PHASE-006
  requirements: [REQ-F-007, REQ-F-008, REQ-F-009, REQ-F-010, REQ-F-011, REQ-NF-004, REQ-NF-008]
  tickets: [TICKET-018, TICKET-019, TICKET-020, TICKET-021]
  validation_matrix: ../VALIDATION_MATRIX.md
  adrs: [../../decisions/ADR-004-review-first-ingestion.md]
  master_docs_touched: [../../architecture/API.md, ../../architecture/ERD.md, ../../architecture/INTEGRATIONS.md, ../../architecture/ARCHITECTURE.md]
---

# DETAIL DESIGN: PHASE-006 AI, Receipt, and Bank Ingestion

## 1. Context and Boundary

All AI, OCR, image, and webhook inputs are untrusted proposal sources. They converge on one durable transaction-draft contract and can never directly affect wallet balances.

- In scope: text AI, multi-transaction extraction, transfers, receipt capture/OCR, multimodal chat images, signed bank webhook, draft edit/reject/confirm, provider resilience.
- Out of scope: voice, model training, automatic confirmation, and storing provider secrets in user-visible data.
- Approval: user-delegated decisions on 2026-08-30. Small task exemption: no.

## 2. Shared Draft Contract

`transaction_drafts` stores user, source (`ai_text`, `receipt_ocr`, `ai_image`, `bank_webhook`, `recurring`), source reference, normalized proposal JSON, raw-provider object reference where retention permits, resolution errors, status (`pending`, `rejected`, `confirmed`, `expired`), confirmation transaction ID, version, and timestamps.

- Proposals use the same transaction types and integer VND rules as PHASE-002.
- Wallet/category references are resolved only from the authenticated user's allowed context; unresolved values remain explicit validation issues.
- Confirmation requires an authenticated CSRF-protected command and idempotency key, revalidates current ownership/version, and calls finance accounting once.
- Provider confidence is informational and never changes confirmation rules.

## 3. Provider Boundaries

| Source | Adapter behavior | Stored sensitive data |
| --- | --- | --- |
| Text AI | OpenAI-compatible structured JSON response; bounded owned wallet/category context | Redacted request metadata and proposal |
| Receipt OCR | Presigned private S3 upload, checksum/size/type validation, configurable OCR HTTP adapter | Private object key and extracted proposal |
| AI image | Private image reference sent to multimodal OpenAI-compatible endpoint | Private object key and proposal |
| Bank webhook | Per-source HMAC, timestamp window, nonce/replay ledger, payload normalization | Redacted event metadata and hash |

Providers receive the minimum required data. HTTP clients use strict timeouts, response size caps, capped retries only for safe transient failures, structured errors, and correlation IDs.

## 4. API and Workflow

- `POST /api/v1/ai/chat` and `/ai/images` create one or more drafts from user input.
- `POST /api/v1/files/presign` returns a short-lived presigned upload for any supported file; receipt OCR/extraction is a later consumer of the returned file ID.
- `POST /api/v1/webhooks/banks/{source}` authenticates by HMAC without user cookies and maps the configured source to one user.
- `GET/PATCH /api/v1/drafts`, `POST /drafts/{id}/confirm`, and `/reject` provide the shared review surface.
- Provider work longer than the request budget runs as a worker job with status polling and notification completion.

## 5. Security Model

- Treat prompts, images, OCR text, and webhook content as hostile data; never interpolate them into system instructions or SQL.
- Use JSON schema validation plus domain validation; unknown fields are ignored or rejected by adapter contract.
- Webhook HMAC covers timestamp, nonce, and raw bytes; nonces are unique per source inside the replay window.
- S3 objects are private, user-prefixed, content-limited, checksum-verified, and accessed through short-lived presigned URLs.
- Logs and audit events redact secrets, raw financial notification text, image bytes, and provider authorization headers.
- Rate limits apply per authenticated user and webhook source.

## 6. User Experience

- AI chat shows editable proposal cards, unresolved fields, provider failure/retry state, and explicit confirm/reject actions.
- Receipt camera/upload is part of quick-add but remains a separate OCR path from AI image chat.
- Multiple proposals are independently editable and selectable; confirmation shows deterministic success/replay state.
- The UI never labels a draft as a confirmed transaction before finance returns success.

## 7. Verification and Reconciliation

- Unit: schemas, normalization, prompt boundaries, redaction, HMAC, replay window, and confirmation state machine.
- Contract: fake OpenAI-compatible, OCR, S3, and webhook fixtures for success, malformed, timeout, retry, oversized, and adversarial responses.
- Integration: user isolation, draft idempotency, confirmation exactly once, webhook dedupe, and audit evidence.
- E2E/UAT: text single/multi/transfer, receipt OCR, image chat, webhook review, edit/reject/confirm, and every provider failure state.
- Reconcile API, ERD, integrations, architecture, validation matrix, security notes, context, backlog, and changelog.
