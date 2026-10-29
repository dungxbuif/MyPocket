---
artifact_type: detail_design
id: FEEDBACK-20261019-AI
status: in_review
owner: shared
approval: owner authorized local feedback fixes; parent test-first barrier released 2026-10-19
updated: 2026-10-19
---

# Advisor and receipt recovery feedback repair

Source: owner-requested feedback sweep, including jar proposal feedback
`824dea29-0a7b-45ca-976d-038616c12448` and product-help feedback
`1a59aa1b-371f-40fa-9e25-35f99236fd51`. Parent tracks the production feedback
snapshot and aggregate reconciliation. No production status mutation is included.

Trace: [backlog](../BACKLOG.md), [active phase](../phases/PHASE-002-native-ios-swiftui.md),
[prior feedback design](NATIVE-FEEDBACK-NEW-20261015-DETAIL_DESIGN.md),
[advisor architecture](../../architecture/AI_ADVISOR_TOOLS.md),
[OCR architecture](../../architecture/OCR_API.md),
[validation](../VALIDATION_MATRIX.md), [release notes](../../releases/CHANGELOG.md).

## Context and concrete gaps

Read project gateway, context/current backlog, quality/validation/debugging
standards, current advisor catalog, active phase and prior feedback design.
The registry, structured cards and confirmation dispatcher already support
wallet create, jar create/update and budget create/update. They retain strict
arguments, owner/generation-bound persisted proposals and explicit confirmation.
Those features must not be represented as missing or expanded into arbitrary APIs.
Recent history queries read descending sequence before the 12-message context
limit; idle expiry and cancellation already have service/repository coverage.

Remaining reproducible candidates:

1. `get_product_help` has no jar section; the assistant excerpt only describes
   wallet proposals. Querying the supported jar feature cannot retrieve its
   navigation/confirmation instructions. Add a versioned jar section and align
   the assistant help with existing registered proposal capabilities. Weight
   title matches above body mentions so a generic navigation section cannot
   displace a feature-specific match when max_results=1.
2. `ocr.go` wraps invalid completed results from immediate submission but returns
   the same errors raw from polling. Expired/missing result text loses the safe
   `ocr_poll` diagnostic and is classified as a generic model/provider error.
   Apply the same safe wrapper at the polling boundary.
3. `Client.Extract` drops `SourceText` after a later image fails OCR, although it
   retains per-attachment text. Preserve successfully extracted, document-marked
   source evidence for owner recovery; keep no drafts, `OCRComplete=false`, and
   never call the model after incomplete OCR.
4. The combined source-size failure sets `OCRComplete=true` even when later
   images were never processed. Derive completion from the number processed so
   unprocessed attachments cannot be marked completed. Keep the existing input
   and context limits and failure behavior.
5. Confirmation loads only the most recent 50 messages. A still-valid proposal
   displayed through older-history pagination incorrectly returns 404. Search
   backwards with the existing owner-scoped message cursor until found or
   exhausted, retaining generation checks and strict cursor progress. This
   affects confirmation lookup only; model context remains 12 recent messages.
6. Independent review traced incomplete OCR through persistence: ID alignment
   emits empty positional placeholders, but `UpdateAttachmentOCR` marks every
   present slot completed. Preserve alignment and mark only nonblank recognized
   text completed. A real PostgreSQL/usecase regression must verify the failed
   filename stays failed and has no text while successful evidence is retained.
7. Malformed model envelopes and unsupported finish/tool responses return only
   source text, dropping completed attachment text/completion metadata. Return
   the already-collected OCR metadata with the same model error, no drafts and
   no retries. Local HTTP fixtures reproduce each rejected response path.
8. Large attachment batches submit OCR serially, so a 20-file request can
   exhaust the 210-second extraction budget even when each document completes
   within the provider wait window. Use a bounded four-worker OCR fan-out for
   batches of four or more, preserving request order, partial recovery and the
   existing small-batch first-failure/source-limit semantics.

## Scope, alternatives and risk

Owned code: backend advisor help/confirmation and AI OCR/client adapter plus focused tests.
No API shape, schema, migration, runtime setting, dependency, finance mutation,
authorization or confirmation policy changes. No live provider calls/costs,
production writes, external retrieval or autonomous financial writes.

Rejected alternative: retry the entire image batch on partial failure. This
could repeat provider charges and loses the existing recovery evidence. Preserve
evidence and fail closed instead. Rejected unrestricted tool access: existing
confirmation and domain validation remain authoritative.

## Verification plan and acceptance

Tests use the real extraction client against local HTTP fixtures, never live
credentials. First run new focused tests RED before product edits. Cases prove
completed-poll invalid text/expiry has a safe OCR-stage diagnostic; a later
image failure preserves only completed document source with no model execution;
the aggregate size limit does not mark unprocessed images complete. Help tool
tests query the registry and verify a relevant jar source is returned with a
confirmation instruction. Run surrounding advisor/entry/OCR tests and race
tests; parent runs aggregate backend suite/vet after all workers finish.

## Reconciliation plan

Record actual results below. Parent updates shared context/backlog/validation/
release notes and reconciles advisor/OCR architecture plus the human/Markdown
agent guide where behavior changed. OpenAPI and ERD need no edits because no
wire or schema contract changes. No ADR is needed: this repairs existing
documented confirmation, diagnostic and recovery boundaries. Authenticated
provider/device UAT remains separate from deterministic local proof.

## Evidence

RED established before product edits:

`rtk go test ./internal/infrastructure/ai ./internal/usecase -run 'TestOCRCompletedPollInvalidResultKeepsDiagnostic|TestOCRPartialBatchPreservesCompletedSourceForRecovery|TestOCRSourceLimitDoesNotCompleteUnprocessedImages|TestAdvisorProductHelpFindsJarConfirmationInstructions' -count=1`

Result: FAIL in both packages, all four new top-level regression tests fail for
the predicted causes. Poll diagnostics are raw `*errors.errorString`; partial
batch source is empty; source-limit output incorrectly reports complete=true;
jar lookup returns no jar source. No product code changes at the barrier.

The additional pagination regression was also RED before implementation:
`rtk go test ./internal/controller/http -run TestAdvisorConfirmOlderJarProposalKeepsOwnerGenerationAndReplay -count=1`.
Current-owner confirmation returned 404 instead of 201 and stale generation
returned 404 instead of 409; foreign-owner rejection already passed.

Surrounding pre-change baseline: 71 passed across usecase/HTTP/repository with
`rtk go test ./internal/usecase ./internal/controller/http ./internal/infrastructure/repository -run 'TestAdvisor|TestBuildAdvisor|TestAIEntry|TestAIEval' -skip TestAdvisorProductHelpFindsJarConfirmationInstructions -count=1`.
This is deterministic coverage; it does not imply a real PostgreSQL fixture ran.

GREEN: `rtk go test -race ./internal/infrastructure/ai ./internal/usecase ./internal/controller/http -run 'TestOCR|TestExtraction|TestModelFailure|TestAdvisor|TestBuildAdvisor|TestAIEntry|TestAIEval' -count=1`
passed 106 tests across all three packages. No provider or production calls.
The first implementation verification passed OCR/pagination but exposed equal
token scores ranking generic navigation above jar help. Giving title matches
an additional relevance point resolved the remaining test; the second
implementation verification passed. `rtk git diff --check` also passed.

## Actual reconciliation and remaining acceptance

All seven bounded fixes are implemented locally. Help catalog version is 1.2;
system handbook remains 1.3. Stable source IDs for unchanged sections remain
unchanged; updated assistant and new jar sections use v1.2 IDs.

Architecture/OCR Markdown references were reconciled, including older-card
lookup, explicit confirmation, title relevance and partial receipt recovery.
Parent owns the public human/agent Finance Assistant guide and shared
context/backlog/validation/release reconciliation. API/ERD/ADR unchanged for the
reasons above. Regression tests demonstrate local behavior; deployment, live
provider accuracy, physical-device acceptance and production feedback resolution
remain outside this repair. Parent aggregate backend verification is pending.

## Independent-review extension

Real disposable PostgreSQL RED: `TestAIEntryPartialOCRPersistsSuccessfulAndFailedAttachments`
and `TestAIEntryOCRSparseAlignmentKeepsEmptySlotsFailed` fail because empty
ID-aligned positions are persisted as `completed`. The first test uses real
`AIEntryService.Process`, real wallet/category/entry repositories and synthetic
provider/storage boundaries; only UUID-owned fixture rows are created and the
owner cleanup removes them. This proves persistence, not merely a test double.

Adapter RED: all five cases in
`TestModelRejectedResponseRetainsCompletedOCRMetadata` fail for dropped
attachment text/status (invalid JSON, missing choice, filtered finish,
unexpected tool call and legacy function call).

Fix: retain aligned positions but complete only nonblank OCR text; preserve
OCR metadata on the two missing model-error returns. Error strings, retry
policy and draft/approval boundaries remain unchanged. `FailMessage` does not
persist `SourceText` into conversation history; durable recovery evidence lives
on attachment OCR text rows. This clarification prevents overstating recovery.

GREEN with disposable PostgreSQL on `127.0.0.1:55436` (database
`mypocket_savings_budget`, migrated through 035):
`rtk proxy env TEST_DATABASE_URL=<disposable-local-fixture> go test -race ./internal/infrastructure/repository ./internal/infrastructure/ai ./internal/usecase -run 'TestAIEntry|TestOCR|TestModelRejectedResponseRetainsCompletedOCRMetadata|TestModelFailure|TestExtraction|TestAIEval' -count=1`
passed all three packages, including the real database tests without skips.
No live provider credentials or production data were used.

## OCR batch timeout repair

The adapter now uses up to four concurrent OCR workers for batches of four or
more files. One- to three-file requests remain serial so a source-size limit or
the first failed document stops before the next submission, preserving the
partial-recovery contract. Completed text is stored by original index and the
model still receives document markers in upload order; no provider response is
replayed and no draft is produced when OCR remains incomplete.

Focused verification:

`rtk go test -race ./internal/infrastructure/ai -count=1`

passes, including `TestExtractionOCRUsesBoundedConcurrencyForLargeBatches`,
which proves a batch of eight uses at least two and no more than four workers,
submits every document once, and completes faster than serial processing.
The existing partial-failure and source-limit recovery tests also pass. This is
local deterministic evidence only; provider rate limits and live image
accuracy remain the separate UAT gate.
