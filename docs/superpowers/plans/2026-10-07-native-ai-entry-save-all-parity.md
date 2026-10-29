# Native AI entry save-all parity implementation plan

## Goal

Match the web AI-entry review surface where pending transaction proposals can
be approved from the sheet header with `Lưu tất cả`, while preserving the
existing per-proposal edit/reject/approve flow and owner-scoped versioned API.

## Contract

- Show a shared-base header action only when at least one proposal is pending.
- Approve pending proposals sequentially through the existing
  `decideEntryProposal` endpoint and proposal version; never write directly or
  bypass validation.
- Stop on the first API error, retain already-updated proposal statuses and
  show the user-facing error; do not auto-retry or duplicate requests.
- Keep loading/disabled state visible while the batch is running.

## Verification

- [x] Add a pure unit contract for pending-proposal selection.
- [x] Add an API-backed/native UI assertion that the header action is exposed
  for a pending fixture.
- [x] Run the full native unit suite, focused UI, Release warnings-as-errors,
  project/base-component guards and diff checks.
- [x] Reconcile `VALIDATION_MATRIX.md`, `PORT-100-AUDIT.md`, iOS design docs
  and `CHANGELOG.md`; no backend/API/schema/migration change was made.

## Shipped evidence

- `AIEntryProposalReview.pending` owns the pending-only selection contract;
  native `AIEntryView` exposes the shared-base `Lưu tất cả` header action and
  sequentially calls the existing versioned approve endpoint, stopping on the
  first error without retrying.
- RED compile evidence is `/tmp/mypocket-ai-entry-save-all-red`.
- Focused unit evidence is 24/24 at
  `/tmp/mypocket-ai-entry-save-all-final-unit/Logs/Test/`.
- Pending-review UI evidence is 1/1 at
  `/tmp/mypocket-ai-entry-save-all-final-ui/Logs/Test/`.
- Full native unit/contract evidence is 95/95 at
  `/tmp/mypocket-ai-entry-save-all-full/Logs/Test/`.
- Release warnings-as-errors build is `/tmp/mypocket-ai-entry-save-all-final-release`;
  project/base-component and diff guards pass. No backend/API/schema/migration
  changed.
