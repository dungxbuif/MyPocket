---
artifact_type: detail_design
id: AI-ENTRY-01
status: in_review
owner: shared
approval: owner_requested_implementation_2026-09-21
---

# AI entry: hold add, chat, review list

Owner approved this slice with “Làm đi” after specifying long-press Add opens entry chat, prefilled proposal list, edit then approve, or reject. This supersedes the two-mode assistant entry point in [umbrella plan](TICKET-09-DETAIL_DESIGN.md). Financial advice is a separate entry point/future slice; this authorization does not approve unrelated transfer deletion/retention policy.

## Scope and decisions

- Short activation of Add keeps manual QuickAdd. Hold 500ms opens AI entry; pointer movement/cancel/blur cancels, release after hold suppresses click. Provide keyboard-accessible alternative from manual editor.
- Chat persists owner-scoped sessions/messages/proposals in PostgreSQL. Model returns zero or more prefilled drafts plus clarification. No ledger effects before explicit approve. Reject is terminal, no ledger effect. Editing increments version; approve checks version and ownership/validation again under a transaction. Repeated approve returns original result, including after that ledger row was later deleted; never recreate it.
- Each row can be edited, approved or rejected independently. Supported ledger kinds remain income/expense/basic/goal with existing category restrictions. Transfer/unknown candidates remain reviewable but approval is blocked until paired ledger exists. No automatic mapping of bank identity to wallet.
- OpenAI-compatible text adapter from server env; no API key means explicit unavailable UI, never mock rows. OCR always precedes model for images. Bounded synchronous processing for this slice (3 JPEG/PNG files, 5 MiB each); bytes are transient and not retained as receipt attachments. Extracted text remains in the owner-scoped session. Durable receipt storage/worker/retention policies stay outside this slice and require the broader plan.
- No automatic retry of ambiguous provider submissions. Session processing lease guards concurrent messages; late completion cannot overwrite a newer request. Server/model failures stay visible and do not produce ledger writes.

## API contract (all responses data envelope, problem+json errors)

- GET `/api/v1/ai/entry/capabilities` -> `{ai_configured, ocr_configured}`.
- POST `/api/v1/ai/entry/sessions` with `{}` -> Session.
- GET `/api/v1/ai/entry/sessions/latest` -> Session or null; GET `/sessions/:id` -> Session.
- POST `/sessions/:id/messages` JSON `{request_id,text,timezone,images?:[{name,mime_type,base64}]}` -> Session after processing. Body limit 22 MiB. Same request ID cannot generate proposals twice.
- PATCH `/api/v1/ai/entry/proposals/:id` `{version,draft}` -> Proposal.
- POST `/proposals/:id/approve` or `/reject` `{version}` -> Proposal. Errors 409 stale/busy/terminal, 400 invalid fields, 503 provider unconfigured/unavailable, 404 wrong owner.
- Session: `{id,messages:[{id,role,content,created_at}],proposals:Proposal[],processing:boolean,error?:string}`.
- Proposal: `{id,session_id,version,status:'pending'|'approved'|'rejected',draft:Draft,questions:string[],transaction_id?:string}`.
- Draft: `{type:string,amount:number,wallet_id:string,category_id:string|null,occurred_at:string,note:string,included_in_reports:boolean}`. Empty wallet/date or zero amount means needs user input; type may be transfer/unknown and cannot approve as ordinary ledger. Fields ready only when validated. Dates RFC3339 or empty; user edits datetime locally and sends offset/UTC explicitly.

## Code, bases and tests

Main: backend entities, repository, transaction-safe proposal approval, HTTP routes/wiring/config, migration, reconciliation docs. Independent provider adapter: `backend/internal/infrastructure/ai/`, uses `net/http`, text-only prompts and OCR HTTP adapter, tests `httptest`. Frontend: `app/src/services/ai.ts`, chat organism/proposal molecule, navigation hold handling, UI tests/spec.

Reused bases: `BaseFab`, `BaseButton`, `FormField`/`BaseTextInput`/`BaseSelect`, `Text`, `Heading`, `SurfaceCard`, `StatusMessage`, `BaseCheckbox`, `BaseBottomSheet`, existing transaction editor fields. New upload atom owns native file control. No custom screen colors/shapes.

Tests: hold fires once/no normal click, cancel/move; review list missing values, edit then approve/reject; DB no write before approve, concurrent approval once, stale version, cross-owner, invalid wallet/category/type, rejection terminal, source error no rows; provider text-only, timeout/error/malformed JSON, injection output validation. Frontend check:design/test:design/build, Go suite, DB fixture integration, browser where available. AI live extraction remains pending actual AI endpoint/model/key; OCR key supplied privately must not enter tracked docs/code.

## Trace and reconciliation

[Ticket](TICKET-09-01-nhap-lieu-chung-tu.md) · [Backlog](../BACKLOG.md) · [Validation](../VALIDATION_MATRIX.md) · [Architecture](../../architecture/ARCHITECTURE.md) · [API](../../architecture/API.md) · [ERD](../../architecture/ERD.md) · [Release](../../releases/CHANGELOG.md) · [ADR registry](../../decisions/README.md). No active phase. Proof/docs review recorded in this file after execution; UI spec in `docs/design/screens/assistant/README.md`. No commit requested in this turn.

## Superseding one-shot follow-up — 2026-09-22

The prior session/message HTTP contract and conversation persistence described below are superseded for the owner-approved AI-ENTRY-02 follow-up. Current routes are `POST /api/v1/ai/entry/process` and read-only `GET /api/v1/ai/entry/requests/{request_id}`, plus independent proposal APIs. There is no public session/message/latest endpoint; the one-shot path writes no conversation messages and does not send history to the model. Browser sends raw multipart files. Backend validates and OCRs first; LLM receives only text and minimal catalogs. Private S3 receipt retention/linking remains incomplete and is tracked separately in [AI-ENTRY-02](AI-ENTRY-02-BATCH-ATTACHMENTS-DETAIL_DESIGN.md).

## Execution log

- Design/authorization: owner-directed entry/review slice; preserve broader plan as proposed scope. Bound OCR to transient extraction; no storage credentials needed for this slice. No AI configuration available yet; request sent while implementation continues.
- Implemented migration 000010, persistent messages/proposals, owner-scoped APIs, edited versions, terminal rejection and approval receipt/ledger atomic transaction. Eight concurrent confirmations produce one row; deleted ledger replay does not recreate it. Report exclusion false is preserved despite GORM defaults.
- Shared system-category wallet assignments were found to include other owners. Corrected read/replace queries to owner-owned wallet links and serialized replacement against approval validation; no migration or deletion of existing owner data. This is the direct category dependency needed for owner-private model context and approval. Regression uses two fixture owners and a fixture shared category.
- Provider adapter uses embedded versioned prompt, strict JSON schema validation, bounded text/OCR input, no image/model tools, no automatic POST retries. Session claim re-reads latest history after lease acquisition; successful OCR text is retained when model extraction fails. History passed to model is explicitly bounded excerpts, so lengthy old evidence may require clarification.
- Runtime limits: the earlier pilot request cap was superseded by [AI-USAGE-01](AI-USAGE-01-DETAIL_DESIGN.md); there is currently no per-user AI/OCR usage limit. One in-flight process is still guarded by the processing lease, and errors/config state remain visible. Secrets saved to ignored `.env.local` with mode 0600; S3 bucket/region and AI endpoint/model/key remain pending user input. S3 storage config is not an implemented receipt-upload flow.

## Verification and docs review — 2026-09-21

- RED→GREEN: new backend validators/repository/service/config; hidden category-template rejection; bounded history; no-user-usage-limit regression after the owner clarified policy; independent review regressions for category owner scoping, OCR evidence on model failure and history after processing claim.
- `rtk go run ./cmd/migrate up` (backend): pass, migration 000010 applied. Integration tests create and clean only their fixture IDs, including a leftover fixture category from a deliberately failing regression. No owner records removed.
- `rtk proxy env 'TEST_DATABASE_URL=postgres://dev:password@127.0.0.1:5432/postgres?sslmode=disable' go test -race ./... -count=1`: pass after final review fixes, including real PostgreSQL tests. `rtk go generate ./cmd/api`: pass, eight new routes documented.
- `TestAIEntryHTTPReviewEditApproveRejectWithRealDatabase`: real DB and HTTP provider boundary fixture, text→two drafts, zero early ledger writes, duplicate request no second extraction, edit/version conflict, approve once, reject, reload and wrong-owner read. This proves pipeline mechanics, not live model quality.
- Frontend tests/design/build pass; detailed commands and browser fixture evidence in [screen spec](../../design/screens/assistant/README.md). Live Chrome through FE origin verified normal click/manual form and AI chat configuration/empty state, without writing owner transactions.
- Final frontend rerun: `rtk npm run test:ai`, `rtk npm run check:design`, `rtk npm run test:design`, `rtk npm run test:transactions`, `rtk npm run build`: all pass. Build 1744 modules; design guard 89 links. New-conversation and explicit unknown-type classification browser regressions pass. Final local-doc scan: 184 links across 22 touched Markdown files, no missing targets; secret scan of Git-visible changed/new files found no supplied credentials; local env mode 0600. `rtk git diff --check`: pass. API restarted with final fixes and existing FE-proxy Google callback preserved.
- Independent backend review completed; important findings fixed with red→green regression tests. No commits made. Existing pending docs changes from the earlier plan were preserved.
- Docs reconciliation: SPEC entry behavior, screen/base specs, API, ERD, architecture, integrations/OCR status, ADR-005, context/backlog, validation matrix and changelog updated. Separate SDD not required: bounded module addition is recorded in architecture + ADR; broad two-flow plan remains future scope.
- UAT remains in review: strict JSON Schema now passes the synthetic 5-case model evaluation, but broader model edge cases and owner review of real proposals remain pending. Browser attachment download and owner acceptance are separate AI-ENTRY-02 UAT. No claim of full TICKET-09, paired transfer ledger, hũ or financial advice completion. The one-shot input/result presentation in [AI-ENTRY-02](AI-ENTRY-02-BATCH-ATTACHMENTS-DETAIL_DESIGN.md) supersedes the session/chat UI shape described earlier in this ticket.

## Live model schema diagnostics — 2026-09-22

- Added structured warning logs on rejected model output: model name, finish reason, model latency, response byte count, and bounded JSON shape issues (path, expected/present fields, JSON types). Only an allowlist of known schema/catalog field names is retained; arbitrary key names are redacted. Never log prompt/OCR/transaction values or raw response; client-facing error remains generic.
- Current implementation uses `net/http` against an OpenAI-compatible chat-completions endpoint with strict `response_format: json_schema`; no SDK dependency or provider-specific URL was needed. Tool/function calling remains disabled: extraction may propose drafts but must never execute ledger actions. JSON Schema capability was verified live against the configured endpoint/model before switching from JSON-object mode; the public API, strict parser, human review and no-ledger-before-approval boundary are unchanged.
- Unit tests prove malformed model values and prompt text do not appear in logs; tests assert useful paths/field names do.
- Live test command: `rtk proxy env AI_EVAL_LIVE=1 AI_EVAL_IMAGE=/tmp/mypocket-ai-eval-receipt.png go test ./internal/infrastructure/ai -run '^TestLiveQwenImageAndTransactionEval$' -count=1 -v`.
- Historical evidence: prior JSON-object mode scored 0/15 with root arrays/input-context-shaped objects and one timeout; a text-only call returned a 363-byte root array (`finish_reason=stop`, 14.706 s). A forced tool-call probe did not finish streaming within the 30 s HTTP-client timeout. These were diagnostic modes only; parser strictness was not relaxed.

## Live strict JSON Schema resolution and wallet context — 2026-09-22

- Owner-directed follow-up: try alternate OpenAI-compatible request methods and use existing wallet descriptions to improve wallet matching. No extra provider or SDK was added.
- Strict JSON Schema `response_format` passed the full production system prompt/catalog and a five-case live run: 5/5 cases, 24/24 scored fields (100%), including multiple transactions/relative date, income, internal-transfer classification and OCR-derived expense. Transfer-safety gate passed. LLM p50 19.769 s/p95 25.655 s; OCR 4.034 s; S3 private PUT/readback 83 ms; OCR + receipt-case pipeline 29.689 s. Exact command and full result: [validation matrix](../VALIDATION_MATRIX.md#ai-entry-01-openai-compatible-json-schema-resolution--2026-09-22).
- Adapter sends any saved wallet `description` as bounded reference context (max 2 KiB per wallet); the system prompt explicitly marks wallet names/descriptions as untrusted and allows descriptions only as matching evidence. This reuses the existing wallet field; no migration/API change. Descriptions that exceed the bound are rejected before provider submission.
- Tests: `rtk proxy go test -race ./internal/infrastructure/ai -count=1` PASS; `rtk proxy go test ./... -count=1` PASS. Probe/production live evaluations use synthetic content only, with no proposal approval or ledger write.
- Future provider/model selection and price transparency is tracked separately in draft [AI-ENTRY-03](AI-ENTRY-03-PROVIDER-SELECTION.md); current UAT remains with the single configured server endpoint and human review of every proposal.
