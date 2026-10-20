---
artifact_type: implementation_note
id: AI-ENTRY-OCR-PROVIDER-AUDIT-20261020
status: reconciled
owner: shared
updated: 2026-10-20
---

# AI-entry/OCR provider contract audit

This bounded audit covers the existing one-shot receipt contract only. It does
not call a live provider, change production configuration, add retries, or
infer new financial behavior.

## Verified contract

- The upload boundary accepts at most 20 JPEG/PNG/PDF files, each at most 5
  MiB. The multipart envelope is bounded separately and validation happens
  before OCR/model calls.
- OCR uses `mode=wait` with `waitTimeoutMs=15000`; a terminal response is
  consumed inline and a queued response is polled with bounded cancellation.
- OCR is completed before the model is called. A partial/failed batch returns
  recovered document text and attachment status, but never drafts or ledger
  writes.
- User text is sent as the separate `user_instruction` field while OCR text is
  marked as untrusted source evidence. The model receives text only, never
  image bytes, Base64, or storage URLs.
- Strict JSON Schema rejects malformed/empty model content and unsupported
  tool/finish responses. Empty drafts are accepted only when the model follows
  the source contract that no recognizable transaction exists.

## Change in this audit

The stale design references that still said “three files” were reconciled to
the shipped 20-file contract. A regression test now covers text instruction +
image OCR in one request, including source markers and no image-byte leakage.

## Verification

```text
go test ./internal/infrastructure/ai -run 'TestExtractionCarriesUserInstructionAlongsideOCRSource|TestOCR|TestExtraction' -count=1
```

The command passed locally. Full backend/race verification remains owned by
the parent task. Live OCR/model accuracy, credentials, and production UAT are
separate external gates and are intentionally not claimed here.
