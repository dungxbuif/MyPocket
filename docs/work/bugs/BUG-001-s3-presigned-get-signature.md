---
artifact_type: bug
id: BUG-001
status: verified
owner: shared
severity: TBD
priority: TBD
bug_markers: [reproduced]
human_fields: [symptoms, expected_behavior, priority, severity]
ai_fields: [reproduction, actual_behavior, root_cause, impact_scope, fix_strategy, regression_tests, verification_results]
shared_fields: [status, trace, docs_review]
trace:
  backlog_item: AI-ENTRY-02
  requirement: AI-03
  phase: none
  detail_design: ../tickets/AI-ENTRY-02-BATCH-ATTACHMENTS-DETAIL_DESIGN.md
  test_verification: ../VALIDATION_MATRIX.md#ai-entry-02-live-provider-evaluation-2026-09-22
  validation_matrix: ../VALIDATION_MATRIX.md
  docs_review: this document
  adrs: [../../decisions/ADR-006-private-ai-attachments.md]
  release_notes: not applicable until verified
---

# Bug: BUG-001 S3 private signed GET returns SignatureDoesNotMatch

## Status

- ID: BUG-001
- Status: verified
- Bug markers: reproduced
- Severity: TBD (human-owned)
- Priority: TBD (human-owned)
- Phase: none
- Owner: shared

## Trace Links

- Backlog item: [AI-ENTRY-02](../tickets/AI-ENTRY-02-BATCH-ATTACHMENTS-DETAIL_DESIGN.md)
- Requirement: AI-03 (private receipt OCR)
- Detail design: [AI-ENTRY-02](../tickets/AI-ENTRY-02-BATCH-ATTACHMENTS-DETAIL_DESIGN.md)
- Test verification: [Validation matrix](../VALIDATION_MATRIX.md#ai-entry-02-live-provider-evaluation-2026-09-22)
- ADR: [ADR-006](../../decisions/ADR-006-private-ai-attachments.md)
- Release notes: not applicable until fix is verified

## Symptoms

The live S3-compatible upload succeeds, but fetching the uploaded object through the application's signed GET URL returns HTTP 403 `SignatureDoesNotMatch`. OCR cannot be verified with the intended private-S3 URL until this boundary works.

## Reproduction

1. Create a synthetic receipt PNG (no personal data).
2. Set `AI_EVAL_LIVE=1` and `AI_EVAL_IMAGE` to that image, then run from `backend/`:
   `rtk proxy env AI_EVAL_LIVE=1 AI_EVAL_IMAGE=/tmp/mypocket-ai-eval-receipt.png go test ./internal/infrastructure/ai -run '^TestLiveQwenImageAndTransactionEval$' -count=1 -v`
3. Observe the S3 Put succeed, then signed GET fail with HTTP 403 / `SignatureDoesNotMatch`.

## Expected Behavior

The configured private S3 object is readable through a short-lived signed GET URL, allowing authenticated OCR/download consumers to read the object without public ACLs.

## Actual Behavior

PUT succeeds. Signed GET fails with the provider's `SignatureDoesNotMatch`. Testing `us-east-1` instead of the configured signing region did not make the request succeed. No OCR or LLM call was reached by this test run.

## Root Cause

The handwritten SigV4 query signer generated a URL the configured S3-compatible endpoint rejected. Replacing it with the AWS SDK for Go v2 S3 presigner resolved the live failure. The exact low-level canonicalization difference was not isolated; the SDK now owns this protocol implementation.

## Impact Scope

- Code: `backend/internal/infrastructure/storage/s3.go`, AI attachment upload/OCR orchestration, authenticated attachment download.
- Users: private receipt OCR and attachment downloads may fail when they consume signed GET URLs.
- Data: the temporary test object was deleted by test cleanup; no ledger rows were written.
- API/contracts: no API contract change identified.
- Related modules: OCR provider, private object storage, attachment lifecycle.

## Fix Strategy

Use AWS SDK for Go v2 only for `SignedGet`, preserving endpoint, region, bucket, static credentials, path-style mode and five-minute expiry. Keep handwritten PUT/DELETE unchanged in this bug.

## Regression Tests

- Keep `TestLiveQwenImageAndTransactionEval` as an opt-in integration regression proving private PUT + signed GET/readback + OCR + model evaluation, with cleanup.
- `TestS3PutAndSignedGetNeverUsePublicURL` checks path-style endpoint/path, SigV4 query parameters and five-minute expiry.
- After a fix, run the live readback and full `go test ./...` suite.

## Verification Results

- Commands: `rtk proxy go test ./internal/infrastructure/storage -count=1` and `rtk proxy go test ./... -count=1` from `backend/` — PASS.
- Live command: `rtk proxy env AI_EVAL_LIVE=1 AI_EVAL_IMAGE=/tmp/mypocket-ai-eval-receipt.png go test ./internal/infrastructure/ai -run '^TestLiveQwenImageAndTransactionEval$' -count=1 -v`.
- Live evidence (first run): private S3 PUT + SDK signed GET/readback PASS (1,027,882 bytes; 236 ms); OCR PASS (3.759 s, synthetic 45,000 VND total recognized). That run's output detached before the full summary.
- Live evidence (captured rerun): S3 PUT + signed GET/readback PASS (1,027,882 bytes; 62 ms); OCR PASS (3.815 s, synthetic total recognized). All five Qwen calls failed draft-schema validation. Harness score 0/15 fields; LLM p50 15.089 s, p95 21.7 s; transfer-safety gate failed. OCR + receipt-case pipeline was 25.516 s. This is an AI-ENTRY-01 model/schema issue, not an S3 regression.

## Fix/Test Attempt Log

- Previous audit: 5 same-path live failures before any production fix; stopped at loop guard.
- Resumed after explicit owner approval to use AWS SDK for Go v2. One SDK implementation cycle; deterministic suite and live S3 readback now pass.
- Status: verified; remaining LLM schema failures are outside this S3 bug.

## Docs Review

- Requirements: unchanged; AI-03 already requires private OCR processing.
- Architecture/API/ERD: unchanged; no contract change.
- ADRs: [ADR-006](../../decisions/ADR-006-private-ai-attachments.md) and [ADR-007](../../decisions/ADR-007-aws-s3-presigning.md).
- `docs/CONTEXT.md`, backlog and validation matrix updated.
