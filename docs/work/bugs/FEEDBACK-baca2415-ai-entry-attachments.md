# Feedback baca2415: multi-file diagnosis and unused-photo cleanup

Status: in_review (read-only audit complete; design/release decisions open).
Owner request: bounded follow-up on `baca2415`, verbatim:
`khi njoeeuf files ko biết vấn đề ở đâu\nĐã xử lý xoá ảnh sau 1 tjowif gian ko dùng chưa`.
Scope: AI-entry multi-file failure visibility and private-attachment cleanup.
No production mutation, provider call, financial write, shell change, or code
edit was made by this audit. The owner separately ran the review-only live
benchmark recorded below.

## Expected and observed

- Expected: a failed batch identifies which selected file needs attention,
  without exposing private document content, storage keys, or provider secrets.
  The user can tell whether unused uploaded originals are removed after a
  bounded retention period.
- Native source renders each selected thumbnail with independent removal, up
  to 20 files, but holds one global `error: String`. Local import errors say
  size/type/duplicate without the filename. After submit, the UI renders one
  process error; `EntryProcessModel` has no attachment-status field.
- The backend logs safe OCR `attachment_index`, stage and code, and persists
  per-attachment OCR status/text internally. The process response exposes only
  batch `error`/`error_code`, reply and proposals. An OCR failure is returned as
  a provider-level error; the failing file is not identified to the user.
  Browser `AiEntrySheet` likewise shows one batch error. This is a real
  visibility gap, not a missing thumbnail/remove control.
- Originals successfully uploaded for AI processing get `delete_after` at
  upload time plus 24 hours. Cleanup claims only expired attachments without
  any transaction link; a successful deletion becomes `deleted`, and failure
  becomes `delete_failed` for retry. Approved attachments remain linked and
  are not selected by this cleanup. Merely selected, unsent native files stay
  in local view state and are not private-storage objects.

## Current-source and deployed-worker evidence

- [Approved AI-ENTRY-02 design](../tickets/AI-ENTRY-02-BATCH-ATTACHMENTS-DETAIL_DESIGN.md)
  established private storage, 24-hour unlinked cleanup, and an operator
  command. The later [approved residual design](../tickets/AI-ENTRY-RESIDUALS-20261020-DETAIL_DESIGN.md)
  added one bounded cleanup call per regular worker pass, conditional on S3
  configuration. It did not approve a per-file API/UI diagnostic contract.
- Current `backend/cmd/worker/main.go`, and deployed release commit
  `837bd505` (`git show`), both construct the cleanup runner only when
  `storage.NewS3` succeeds, then call it with limit 100 each pass. Current
  repository claim uses `SKIP LOCKED`, excludes linked rows, and can reclaim
  stale `deleting` rows. The one-shot command remains available.
- Read-only `ssh vm100` Docker listing on 2026-10-26 showed
  `mypocket-worker` on `release-1.2.1-837bd505`. Its startup log states:
  `attachment cleanup unavailable; configure private storage to enable
  scheduled cleanup: S3 storage is not configured` (2026-10-26 00:52:50).
  Read-only container environment-name inspection showed the worker has
  `APP_ENV`, `S3_ENDPOINT`, `S3_REGION`, `S3_BUCKET`, but lacks
  `S3_ACCESS_KEY_ID` and `S3_SECRET_ACCESS_KEY`; the API container lists both.
  No credential values, database rows, object keys, or private images were read.
  Therefore the currently deployed worker cannot run scheduled cleanup,
  despite the scheduling code being in its image. The read-only audit cannot
  establish how many expired unlinked objects, if any, exist.

## Focused verification

- PASS: `rtk go test ./cmd/worker ./internal/usecase ./internal/infrastructure/ai -run 'TestAttachmentCleanup|TestAIEntry.*(OCR|Storage|Attachment)|TestOCR' -count=1`
  in `backend/`: 28 tests in three packages. These verify bounded worker
  invocation, cleanup state transitions, OCR failure boundaries and storage
  compensation locally; they do not prove the deployed worker is configured.
- PASS: isolated iPhone 16 Simulator
  `testAIEntryConfiguredAttachmentLimitAndThumbnails` (1/1, zero skips),
  `/tmp/mypocket-feedback-baca2415-native.xcresult`. It checks selection
  count and independent removal, not per-file provider error visibility.
- PASS: `rtk npm run check:design`, `rtk npm run test:design` (7/7 design
  tests plus base checks), and `rtk npm run build` from `app/`; the build has
  the existing large-chunk warning. `rtk git diff --check` passed. These
  checks validate documentation links and current client build, not the
  proposed per-file contract.

## Owner-supplied live three-PDF benchmark — 2026-10-26

The owner completed a review-only, authenticated run against unchanged
production using three PDFs. Total elapsed time was 111 seconds; serial OCR
per file took 2.044, 0.647 and 1.000 seconds (3.691 seconds combined), while
the model took 106.326 seconds and produced 18,664 completion tokens. OCR
concurrency is therefore **not this run's principal latency bottleneck**;
optimizing it cannot explain the measured model time. This run is distinct
from the local 20-thumbnail fixture and does not verify deployed cleanup.

The independent canonical comparison found exact date/type/amount on 55 of
87 external candidates. Twenty-eight held and 20 excluded source rows were
potentially proposed; 23 proposals were dated today versus one source bank
row dated today. These are review-safety and source-attribution concerns, not
evidence that a particular file failed OCR. Do not auto-approve or treat 158
proposals as 158 valid ledger entries. Source-to-proposal provenance, model
output limits and extraction quality require a separate local design and
regression/evaluation gate, informed by the
[three-file benchmark ticket](../tickets/AI-ENTRY-THREE-FILE-BENCH-20261026.md)
and [PDF safety work](AI-ENTRY-PDF-REVIEW-SAFETY-20261026.md). No broad
architecture or production change is authorized by this feedback audit.

## Decision needed before a fix

1. **Per-file diagnosis:** approve a small owner-scoped, privacy-safe API
   response/recovery contract (for example stable attachment position plus
   stage/status and safe error code) and native presentation for the affected
   file. Do not expose OCR text, original bytes, object keys, provider URLs or
   raw upstream errors. Define how a failed file remains associated with the
   request after an uncertain submit; do not auto-retry uploads or OCR. This
   changes a public response and error UX, so it is outside the approved
   cleanup/card detail designs. Test first with a deterministic two-file OCR
   failure: backend API/recovery must identify only the failed position, and
   native UI must show it next to that file while retaining successful evidence.
2. **Deployed cleanup:** owner-approved worker service configuration/release
   must provide its existing S3 credentials through the established secret
   path, then verify a healthy worker pass. Review non-sensitive logs and a
   bounded, owner-authorized cleanup probe; do not run the one-shot deletion
   command or inspect production objects as part of this audit. No code change
   or new cleanup architecture is justified by the present evidence.
3. **Benchmark quality/latency:** diagnose the 106.326-second model stage and
   incorrect/unsafe rows with reproducible, source-aligned local tests before
   proposing output-budget, prompt, batching, provenance or concurrency
   changes. The three-file OCR timings specifically rule out claiming that
   OCR fan-out fixes this benchmark's main latency. Keep any wider change
   under its own approved design and review-only acceptance; do not deploy.

Docs review: the [assistant screen contract](../../design/pages/assistant/README.md),
[integration contract](../../architecture/INTEGRATIONS.md),
[AI-entry API reference](../../architecture/API.md), and
[deployment runbook](../../operations/DEPLOYMENT.md) distinguish implemented
conditional scheduling from the currently unconfigured deployed worker. The
existing Markdown AI-entry reference is both human- and agent-readable. No
schema, configuration default, or API behavior changed in this audit; a new
per-file public response remains gated on the first decision. The in-flight
`docs/guides/ai-entry.md` is unchanged because this audit did not ship new
user behavior; the current-limit statement lives in the existing API and
screen references.

Trace: [unresolved feedback audit](../audits/FEEDBACK-UNRESOLVED-UAT-20261020.md),
[validation matrix](../VALIDATION_MATRIX.md),
[deployment record](../../operations/DEPLOYMENT.md),
[private-attachment ADR](../../decisions/ADR-006-private-ai-attachments.md).
