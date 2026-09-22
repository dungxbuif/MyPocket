---
artifact_type: api_contract
id: API-OCR-001
status: ready
owner: shared
human_fields:
  - business_contract_intent
  - approval
ai_fields:
  - request_response
  - errors
  - auth_notes
  - sequence_diagram
  - versioning_notes
shared_fields:
  - status
  - trace
trace:
  requirements:
    - docs/design/system/DESIGN.md
  phases: []
  tickets_or_bugs: []
  detail_designs: []
  test_verification: not-run-docs-only
  docs_review: docs/architecture/API.md
  adrs: []
  release_notes: not-required-unreleased
---

# OCR API Contract

## Field Ownership

- Human owns the decision to use `https://ocr.dungxbuif.com/`, credentials, quota policy, and which OCR output becomes MyPocket business data.
- AI maintains this contract, usage notes, failure modes, and implementation guidance.

## Status

- 2026-09-22 runtime update: one-shot AI upload accepts JPEG/PNG/PDF. Backend validates the uploaded bytes, stores the original in private environment-qualified MyPocket S3, and sends OCR Platform a short-lived signed source URL. OCR completes before text extraction; the LLM receives only user/OCR text (never file bytes, base64, URL or image input). Provider adapter tests also retain coverage for OCR Platform's own private PDF presign flow, but the app upload path uses MyPocket S3 source URLs for both images and PDFs. A live provider probe currently returns HTTP 401 `UNAUTHORIZED` for the configured credential; this is an external secret provisioning/rotation blocker, not a file-type or client MIME failure.
- ID: API-OCR-001
- Status: implemented locally; live provider credential blocked
- Owner: shared
- Public contract checked 2026-09-20: OpenAPI and capabilities returned HTTP 200. Scan-only routes are not advertised; do not implement against older scan guidance. This check did not run an authenticated OCR job.
- Source documentation:
  - `https://ocr.dungxbuif.com/`
  - `https://ocr.dungxbuif.com/guides/onboarding`
  - `https://ocr.dungxbuif.com/api/OCR_RESPONSE`
  - `https://ocr.dungxbuif.com/api/MCP_INTEGRATION`

## Business Contract Intent

MyPocket will use OCR Platform to read receipt photos or PDFs attached to transactions. OCR is an assistant workflow, not the source of truth: the user must be able to review and correct merchant, amount, date, category, and note before saving a transaction.

The first implementation should support:

- Receipt image or PDF upload.
- OCR job submission.
- Polling known `documentId`.
- Persisting normalized text needed by MyPocket before OCR result expiry.
- Scan enhancement is outside the verified public contract and is not an implementation dependency.

## External API Summary

| Contract | Type | Auth | Status | Notes |
| --- | --- | --- | --- | --- |
| `GET /v1/ocr/capabilities` | HTTP | none documented for quickstart | ready | Discover accepted file types, recognition options, and active limits. |
| `POST /v1/documents` | HTTP | Bearer API key | ready | Submit OCR job from URL/Base64/presigned source URL. Returns `202 Accepted` and `documentId`. |
| `GET /v1/documents/{documentId}` | HTTP | Bearer API key | ready | Poll one known document. Completed response includes `result`; no separate result endpoint. |
| `POST /v1/uploads/presign` | HTTP | Bearer API key | ready | Request signed upload URL for large/private files. |
| object storage `PUT uploadUrl` | HTTP | signed URL only | ready | Upload bytes directly; do not send OCR API key. |
| `/mcp` | MCP / JSON-RPC / SSE | Bearer API key | future | Agent integration. Use later for MCP clients, not initial FE receipt flow. |

## Authentication

Protected requests use:

```http
Authorization: Bearer sk_ocr_replace_with_real_key
```

Required environment variables for MyPocket implementation:

```bash
OCR_API_URL=https://ocr.dungxbuif.com
OCR_API_KEY=sk_ocr_replace_with_real_key
```

Rules:

- Never commit `OCR_API_KEY`.
- Do not expose `OCR_API_KEY` in browser code.
- MyPocket frontend should upload through MyPocket backend or a controlled server action later; do not call protected OCR endpoints directly from public FE.
- Do not send the API key when uploading bytes to the returned object-storage `uploadUrl`; the signed URL already carries temporary authorization.

## Recommended MyPocket Receipt Flow

```mermaid
sequenceDiagram
    actor User
    participant FE as MyPocket FE
    participant API as MyPocket API
    participant OCR as OCR Platform
    participant Store as MyPocket Storage

    User->>FE: Attach receipt image/PDF
    FE->>API: Upload receipt candidate
    API->>OCR: POST /v1/uploads/presign
    OCR-->>API: uploadUrl, sourceUrl, signed headers
    API->>OCR: PUT uploadUrl with file bytes
    API->>OCR: POST /v1/documents with sourceUrl
    OCR-->>API: documentId, queued status
    API-->>FE: receiptOcrJobId/documentId
    FE->>API: Poll receipt OCR status
    API->>OCR: GET /v1/documents/{documentId}
    OCR-->>API: queued/processing/completed/failed
    API->>Store: Persist extracted text and parsed draft fields
    API-->>FE: OCR draft for user review
    User->>FE: Correct and save transaction
```

## Basic OCR Submission

Use a public HTTPS URL only when the file can safely be public:

```bash
curl --fail --silent --show-error \
  -X POST "$OCR_API_URL/v1/documents" \
  -H "Authorization: Bearer $OCR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "input": {"url": "https://files.example.com/receipt.pdf"},
    "options": {
      "recognitionLevel": "accurate",
      "languages": ["vi-VN", "en-US"],
      "usesLanguageCorrection": true
    }
  }'
```

Expected initial response:

```json
{
  "documentId": "doc_18f673199c0",
  "status": "queued",
  "createdAt": "2026-08-15T08:30:00Z",
  "links": [
    {"rel": "self", "href": "https://ocr.dungxbuif.com/v1/documents/doc_18f673199c0"}
  ]
}
```

Implementation rule: persist `documentId`. Public document listing is intentionally unavailable.

## Private Or Large File Upload

Use presigned upload for private receipts and larger files.

### 1. Request signed upload

```bash
curl --fail --silent --show-error \
  -X POST "$OCR_API_URL/v1/uploads/presign" \
  -H "Authorization: Bearer $OCR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "filename": "receipt.jpg",
    "sizeBytes": 123456,
    "contentType": "image/jpeg"
  }'
```

The response includes:

- `method`: expected to be `PUT`.
- `uploadUrl`: temporary object-storage URL for bytes.
- `sourceUrl`: account-owned source URL to submit to OCR.
- `headers`: exact `Content-Length` and `Content-Type` to use on upload.

### 2. Upload bytes

```bash
curl --fail --silent --show-error \
  -X PUT "$UPLOAD_URL" \
  -H "Content-Length: $UPLOAD_LENGTH" \
  -H "Content-Type: $UPLOAD_TYPE" \
  --data-binary "@receipt.jpg"
```

Rules:

- Use the exact headers returned by presign.
- Do not submit `sourceUrl` if the byte upload failed.
- Do not construct `s3://` or storage URLs manually.

## 401 troubleshooting

If `/api/v1/ai/entry/process` returns `503 ai_entry_unavailable` and the server diagnostic contains `stage=ocr_submit`, `code=ocr_submit_request`, and `provider_status=401`, the OCR platform rejected the configured `OCR_API_KEY`. The adapter already sends the documented `Authorization: Bearer` header and fails closed without saving transactions. Rotate or re-enable the key in the backend secret store, restart the API, and verify the provider with a server-side request; never place the key in browser code or bypass OCR authentication. A 401 cannot be repaired by changing the uploaded image MIME header.

### 3. Submit source URL

```bash
curl --fail --silent --show-error \
  -X POST "$OCR_API_URL/v1/documents" \
  -H "Authorization: Bearer $OCR_API_KEY" \
  -H "Content-Type: application/json" \
  -d "{\"input\":{\"url\":\"$SOURCE_URL\"}}"
```

## Poll Document Status

```bash
curl --fail --silent --show-error \
  "$OCR_API_URL/v1/documents/$DOCUMENT_ID" \
  -H "Authorization: Bearer $OCR_API_KEY"
```

Status handling:

| OCR Status | MyPocket Behavior |
| --- | --- |
| `queued` | Show pending OCR state. Poll later. |
| `processing` | Show processing state. Respect `Retry-After` if present. |
| `completed` | Persist needed OCR result fields and present a transaction draft. |
| `failed` | Keep receipt attachment, show retry/manual-entry option, log diagnostic detail. |
| `cancelled` | Mark processing stopped; allow manual entry or an explicit new attempt. This status does not imply a public cancel endpoint exists. |

Completed document response includes `result` directly:

```json
{
  "documentId": "doc_18f673199c0",
  "status": "completed",
  "inputContentType": "application/pdf",
  "inputSizeBytes": 248190,
  "resultExpiresAt": "2026-08-22T08:30:04Z",
  "result": {
    "text": "Invoice 1042\nTotal: $82.00",
    "pageCount": 1,
    "pages": [
      {
        "pageNumber": 1,
        "text": "Invoice 1042\nTotal: $82.00",
        "blocks": [
          {
            "text": "Total: $82.00",
            "confidence": 0.9612,
            "bbox": [0.091, 0.781, 0.284, 0.041]
          }
        ]
      }
    ]
  }
}
```

## OCR Result Consumption

Use these fields first:

| Field | MyPocket Use |
| --- | --- |
| `result.text` | Raw searchable receipt text; persist if OCR succeeded. |
| `result.pageCount` | Receipt/PDF preview metadata. |
| `result.pages[].text` | Page-level parsing and future highlights. |
| `result.pages[].blocks[].text` | Optional highlight/search building block. |
| `result.pages[].blocks[].confidence` | Ranking signal only; do not treat as business truth. |
| `result.pages[].blocks[].bbox` | Optional visual highlights; normalized `[x, y, width, height]`. |
| `resultExpiresAt` | Deadline to persist needed OCR data into MyPocket. |

Consumer rules:

- `pageNumber` is one-based and should be treated as authoritative.
- `pages` and `blocks` may be missing or empty; code defensively.
- Do not reconstruct business text by sorting bounding boxes unless tested against receipt layouts.
- For MyPocket v1, parse draft fields from `result.text` and let the user confirm.

## Scan-Only Availability

The [public OpenAPI](https://ocr.dungxbuif.com/api/v1/openapi.json), checked 2026-09-20, does not list `/v1/scans` or `/v1/scans/{scanId}`. Earlier scan-only guidance is superseded. Reverify provider support before designing this enhancement; the AI entry plan requires OCR text only.

## Errors

| Error / Status | Meaning | MyPocket Consumer Impact |
| --- | --- | --- |
| `400 INVALID_INPUT` | Request body or options are invalid. | Show developer-safe user error and log request context without secrets. |
| `401/403` | Missing, invalid, or unauthorized API key/request. | Backend configuration issue; disable OCR and alert maintainers. |
| `404 NOT_FOUND` | Unknown document, wrong owner, or retained metadata gone. | Mark OCR job unavailable; keep manual receipt entry path. |
| `410 RESULT_EXPIRED` | OCR result TTL expired. | Use persisted MyPocket OCR data if available; otherwise request re-scan. |
| `413 URL_CONTENT_TOO_LARGE` | Upload exceeds deployment limit. | Show max size from `limits.maxUploadBytes`; ask user to reduce/split file. |
| `415 UNSUPPORTED_MEDIA_TYPE` | Unsupported input media. | Reject unsupported upload in UI before submission when possible. |
| object-storage `403` XML | Signed upload headers/body mismatch or expired upload URL. | Request a fresh presign URL; do not submit stale `sourceUrl`. |
| OCR `failed` status | Worker could not complete recognition. | Keep attachment, allow retry, and support manual transaction entry. |

## Versioning And Retention

- Base URL for current integration: `https://ocr.dungxbuif.com`.
- Public REST path version: `/v1`.
- Live OpenAPI 3.1 contract is advertised at `/api/v1/openapi.json`.
- OCR results are temporary. MyPocket must persist any needed text or draft fields before `resultExpiresAt`.
- Public API has no document list, delete, cancel, or separate result endpoint.

## Implementation Plan Notes

Initial implementation should add a MyPocket backend adapter with these internal responsibilities:

1. Hold `OCR_API_KEY` server-side.
2. Request presigned upload for receipt bytes.
3. Upload bytes with exact signed headers.
4. Submit OCR job and persist `documentId`.
5. Poll status or run background job.
6. Persist normalized OCR result text and draft extraction.
7. Return a reviewable transaction draft to FE.

Suggested MyPocket internal data additions:

| Entity | Fields |
| --- | --- |
| `receipt_attachment` | id, local object id, content type, size, transaction id nullable |
| `receipt_ocr_job` | id, receipt id, provider `ocr.dungxbuif.com`, document id, status, result expires at |
| `receipt_ocr_result` | job id, raw text, page count, parsed merchant, parsed total, parsed date, confidence notes |

## Linked Decisions

- [Two-flow AI review plan](../work/tickets/TICKET-09-DETAIL_DESIGN.md): OCR-first/text-only model input, source provenance, private upload and retention proposals. Review pending; no integration implemented by this docs update.
- No ADR yet. Create an ADR when the team approves OCR Platform as the durable OCR provider or when secrets/runtime architecture is finalized.
