---
artifact_type: operational_checklist
id: PROVIDER_UAT
status: credential_gated
owner: shared
human_fields: [credential_approval, synthetic_fixture_approval, acceptance_signoff]
ai_fields: [preflight, commands, evidence, failure_modes]
shared_fields: [status, scope, no_production_mutation]
---

# AI/OCR provider UAT checklist

## Deterministic provider contract (2026-10-19)

The provider/OCR boundary is tested without paid or live calls. The contract is:

- OCR completes (and returns non-empty text) before the model request starts.
- Image batches accept 1–20 files, reject the 21st file before any upstream call,
  and fan out OCR with the bounded concurrency configured by the client.
- A malformed model envelope or an incomplete/tool-requesting extraction response
  is reported with a safe, machine-readable model diagnostic (`stage=model`);
  no provider body, URL, source text, or credential is included in the error.
- OCR evidence is retained when the model rejects a response, while no drafts are
  persisted from a rejected response.
- The extraction client is text-only at the model boundary: image bytes, base64,
  and private source URLs never cross into the model request.

These checks use `httptest` and synthetic receipt bytes. They do not call a live
OCR/LLM provider; live probes remain opt-in (`OCR_LIVE_TEST=1` or
`AI_EVAL_LIVE=1`) and are not release gates.

This is the human- and agent-readable gate for configured-provider testing. It
is intentionally separate from the deterministic Go suite. The 2026-10-19
local gate passed without contacting AI, OCR or S3; it does not prove provider
availability or image-recognition quality.

## Scope and safety

- Use a staging endpoint or an explicitly approved local provider.
- Use only synthetic receipts and synthetic wallet/category data.
- Do not use a production receipt, production API key, production database, or
  a real user's financial row.
- Never print `AI_API_KEY`, `OCR_API_KEY`, S3 credentials, response bodies or
  OCR text in logs or committed artifacts.
- The extraction path proposes drafts only. It must not approve a transaction.
- Stop if the configured endpoint is not owned/approved, or if the test would
  incur an unapproved provider charge.

## Preflight

Required server-side variables:

```text
AI_BASE_URL      # origin + provider prefix, for example https://host/v1
AI_API_KEY       # never put this in browser or iOS code
AI_MODEL
OCR_API_URL      # default https://ocr.dungxbuif.com
OCR_API_KEY
S3_ENDPOINT      # required for the private receipt round-trip
S3_REGION
S3_BUCKET
S3_ACCESS_KEY_ID
S3_SECRET_ACCESS_KEY
```

The committed `backend/.env.example` contains names and safe local defaults,
not credentials. `backend/.env.local` is ignored; process environment values
take precedence. Before any live command, verify that the fixture path is
synthetic and that the selected database is not production.

## Commands

Deterministic gate (safe and required for every change):

```sh
cd backend
go test ./internal/infrastructure/ai ./internal/usecase -count=1
```

Optional synthetic provider gate, only after preflight and explicit approval:

```sh
cd backend
AI_EVAL_LIVE=1 \
AI_EVAL_IMAGE=/absolute/path/to/synthetic-receipt.png \
go test ./internal/infrastructure/ai \
  -run '^TestLiveQwenImageAndTransactionEval$' -count=1 -v
```

The focused probes use `AI_LIVE_TEST=1`, `AI_TOOL_CALL_PROBE=1`,
`AI_JSON_SCHEMA_PROBE=1`, `AI_RAG_LIVE=1`, `AI_REVIEW_LIVE=1`, or
`OCR_LIVE_TEST=1` as documented by their test names. They are independent
probes, can incur provider charges, and are never run by default. A skipped
probe is not a pass.

## Evidence to retain

Record only:

- commit/image identifier and non-secret endpoint/model label;
- pass/fail, HTTP status class, latency, token counts and draft/field counts;
- synthetic fixture identifier and whether S3 cleanup succeeded;
- no raw prompt, OCR text, receipt bytes, bearer token or financial row.

Required acceptance checks:

1. AI text extraction returns the expected synthetic type, amount, wallet and
   date, with strict schema validation.
2. OCR image extraction completes, preserves document order and returns text
   only to the model; PDFs use provider presign and exact signed upload headers.
3. OCR/model failures expose only safe stage/code diagnostics and create no
   transaction or auto-approved proposal.
4. The synthetic S3 object is readable during the test and deleted afterward.
5. Provider usage is recorded only as owner-scoped operational metadata when
   `AI_STORE_USAGE=true`.

## Current status

- Deterministic adapter, OCR validation, timeout, failure-redaction and
  persistence tests: **local pass**.
- The 2026-10-19 provider contract additions pass locally: malformed model
  envelopes are classified as `model_response_invalid`, incomplete/tool
  responses as `model_response_incomplete`, and empty model content as a
  schema mismatch without drafts. Verification commands:
  `go test ./internal/infrastructure/ai -count=1` and
  `go test ./internal/usecase ./internal/controller/http -count=1`; the full
  backend suite also passes with `go test ./... -count=1`.
- Attachment storage uses a 45-second request-scoped deadline and at most four
  concurrent private-object uploads; a storage timeout stops before any OCR or
  model charge and rolls back staged objects.
- Current live AI/OCR/S3 probes: **not run in the 2026-10-19 gate**.
- Production/provider UAT and physical-device receipt workflow: **pending**.

See [OCR API contract](../architecture/OCR_API.md),
[integrations](../architecture/INTEGRATIONS.md), and the
[validation matrix](../work/VALIDATION_MATRIX.md) for implementation details
and historical evidence. Historical pass entries must not be reused as proof
of current credentials or deployment health.
## Live advisor conversation UAT

The read-only production regression script intentionally does not submit AI
questions. Use `scripts/mypocket-advisor-uat.sh` when an explicit provider run
is approved. It logs in with a dedicated UAT account, submits one synthetic
question, polls the returned run with a bounded timeout, and verifies the
persisted conversation contains both the user and assistant messages.

Credentials are supplied only through the environment. The script never
prints tokens, passwords, prompt contents, or the provider response body. A
timeout or provider error stops the run and does not resubmit automatically.

```bash
UAT_BASE_URL=https://money.dungxbuif.com \
UAT_EMAIL="$UAT_EMAIL" \
UAT_PASSWORD="$UAT_PASSWORD" \
UAT_ADVISOR_PROMPT='Tổng chi tiêu tháng này của tôi là bao nhiêu?' \
  ./scripts/mypocket-advisor-uat.sh
```

The output contains only the `run_id`, `conversation_id`, terminal status and
message count. Pass the returned conversation ID to the authenticated
read-only harness when the rest of the account surface should be checked:

```bash
UAT_CONVERSATION_ID='<returned id>' ./scripts/mypocket-production-uat-regression.sh
```

This is a provider-dependent UAT, not a deterministic local contract test.
Local tests remain the source of truth for owner isolation, tool routing,
schema validation, idle expiry and duplicate-request behavior.

### Latest evidence (2026-10-19)

- Live async submission completed: `run_id` and `conversation_id` were
  captured in restricted temporary evidence; the persisted history contained
  two messages (user + assistant).
- The authenticated read-only regression then passed all covered endpoints,
  including advisor history, with HTTP 200 responses.
- The first harness attempt failed only while parsing the history envelope
  (`data` is an array, not an object). The parser was corrected and the
  completed run was verified without another provider submission.
