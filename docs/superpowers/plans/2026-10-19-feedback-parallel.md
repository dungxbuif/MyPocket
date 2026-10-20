# Feedback and Backlog Parallel Implementation Plan

**Goal:** Reconcile all fetched feedback against the native app, repair proven
defects with regression evidence, and preserve explicit remaining feature and
release gates.

**Architecture:** Three independent implementation lanes share agreed existing
contracts. The coordinator owns integration and documentation. PostgreSQL
row locking protects portfolio history; current owner/generation/action replay
contracts protect AI confirmation; native bases own common interactions.

**Spec:** [Batch detail design](../../work/tickets/FEEDBACK-20261019-DETAIL_DESIGN.md).
**Execution:** Owner requested parallel plan/tests, parallel implementation,
then combined verification. No additional execution-method choice is required.
**Branch:** `codex/feedback-backlog-20261019` from `f10cff45`.

## Dependencies and review focus

- No shared file has multiple writers. The coordinator owns context/backlog,
  validation and changelog; lanes own their implementation and detail-design docs.
- Tests must distinguish a real defect from missing local tooling. Existing
  passing behavior remains intact; status `open` alone is not a reproduction.
- Portfolio deletion must serialize with trade creation; malformed price input
  must leave the stored price unchanged.
- Older AI actions must retain owner/generation checks and idempotent replay.
  Looking up a card must not increase model context.
- Partial OCR must keep completed evidence without fabricating completed files
  or drafts. Safe diagnostics must not expose document/provider secrets.
- Back gestures must not discard edits or escape the root route. Account Add
  returns to the caller. Shared row spacing must support larger text.

## Wave 1: Plan and regression tests

- [x] Read the production snapshot and assign all 48 IDs in the batch design.
- [x] Native: reproduce missing Account Add, disabled edge Back and row density.
  [Native design/tests](../../work/tickets/FEEDBACK-20261019-NATIVE-DETAIL_DESIGN.md).
- [x] Finance: reproduce omitted-price erasure and concurrent asset/trade loss
  using real disposable PostgreSQL.
  [Finance design/tests](../../work/tickets/FEEDBACK-20261019-FINANCE-DETAIL_DESIGN.md).
- [x] AI: reproduce completed-poll diagnostics, partial source loss, inaccurate
  batch completion, missing jar help and older-action confirmation failure.
  [AI design/tests](../../work/tickets/FEEDBACK-20261019-AI-DETAIL_DESIGN.md).
- [x] Coordinator reviews contracts and releases the implementation barrier.

## Wave 2: Parallel implementation

- [x] Finance: require explicit price field; lock asset before checking trades.
- [x] AI: preserve source/completion/diagnostics, add jar help, page confirmation
  lookup without enlarging provider history.
- [x] Native: Account Add/caller restoration, scoped interactive Back with dirty
  editor protection, shared row separation.
- [x] Independent review: review native gesture lifecycle and implementation
  interactions; fix findings before aggregate acceptance.
- [x] Platform audit: assess existing worker/auth/offline contracts and identify
  actual remaining feature gaps, using current source rather than stale rows.

## Wave 3: Aggregate proof and reconciliation

- [x] Backend `go test -race -json ./... -count=1` against migration-35 disposable
  PostgreSQL and a dedicated ephemeral Redis; count pass/fail/skip explicitly.
- [x] Backend `go vet ./...`.
- [x] Native focused UI/units then full native unit target, Release
  warnings-as-errors build, base/project guards. One simulator driver at a time.
- [x] Review all diffs and `git diff --check`; verify docs links.
- [x] Reconcile guides (human + Markdown agent sections), architecture,
  context/backlog/validation and unreleased changelog.
- [x] Report actual remaining contracts, live-provider, production and physical
  acceptance gates without marking all feedback resolved.

Final local proof: backend 448 passing test/subtest results, nine opt-in live
probes skipped; native 354 unit + 6 selected UI tests passed, zero failures or
skips; Release simulator build and guards passed. See the
[validation matrix](../../work/VALIDATION_MATRIX.md) and
[remaining deliverables](../../work/tickets/FEEDBACK-20261019-REMAINING.md).
The completed boxes describe this finite repair batch, not every future
feature or physical/production acceptance gate.

## Resource boundaries

Local PostgreSQL: existing disposable `mypocket-savings-budget-test-20261016`,
port 55436, schema35; tests use UUID-scoped synthetic rows and cleanup.
Local Redis: temporary `mypocket-feedback-redis-20261019`, loopback-only dynamic
port, persistence disabled and Docker auto-removal enabled. Stopped after final
verification; its anonymous image volume is disposable. Never use production
financial rows as fixtures. Physical iPhone and production remain unchanged.
