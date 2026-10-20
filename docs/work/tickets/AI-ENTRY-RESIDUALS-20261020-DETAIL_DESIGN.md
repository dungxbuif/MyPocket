# AI entry residual feedback — detail design (2026-10-20)

## Scope

Close the two residual production-feedback items recorded in
`docs/work/audits/FEEDBACK-UNRESOLVED-UAT-20261020.md` without touching
production data or changing the AI-entry ownership contract:

1. Make expired, unlinked AI-entry attachments run through the existing
   bounded cleanup use case from the regular worker, while retaining the
   one-shot operator command for diagnostics and recovery.
2. Make AI-entry proposal rows use the canonical `MPTransactionRow` contract,
   so title/category glyph, date/note subtitle, signed amount, privacy mask,
   and Dynamic Type behavior are identical to the ledger and assistant.

## Backend decision

The worker will call `AIEntryAttachmentCleanup.Run` once per pass with a
bounded batch of 100. The repository already claims only ownerless, expired
rows with `SKIP LOCKED`, and records `deleted` or `delete_failed`; this keeps
the pass idempotent and safe when multiple worker instances overlap. The
cleanup command remains the explicit diagnostic path and reports the number
of successful deletions. Worker diagnostics log counts and errors but never
object keys, OCR text, filenames, or owner IDs. Configuration is limited to
the existing worker interval and `WORKER_ONCE` controls; no new production
secret or migration is required.

## iOS decision

Introduce a small presentation adapter that maps an `EntryProposalModel` to
the existing structured initializer of `MPTransactionRow`. The proposal
editor/actions remain outside the row and continue to require explicit user
review. Pending rows show the canonical transaction anatomy followed by the
existing review actions; resolved rows show the same row followed by status.
The adapter supplies the account timezone and privacy-mask environment, and
does not create a persisted `TransactionModel` or invent a wallet/category.

This preserves proposal-specific editing and approval behavior while removing
the duplicate amount/date/note/type rendering that caused visual drift.

## Verification

- Backend unit test: the worker cleanup pass invokes the bounded cleanup
  contract, propagates the cleanup error, and is a no-op when not configured.
- Existing cleanup use-case tests remain green.
- iOS unit test: the proposal presentation adapter delegates to the canonical
  transaction presentation for expense/income and keeps proposal-specific
  type labels available for unknown types.
- Existing AI-entry UI tests and the native design-system test target run after
  the focused tests.

## Implementation status

Implemented locally on 2026-10-20. The worker creates the cleanup use case only
when private storage is configured and logs bounded pass results without
private attachment content. `AIEntryProposalReview.transactionRow` now maps
proposal drafts into `MPTransactionRow`; proposal review controls and
confirmation semantics are unchanged. Production worker scheduling and visual
device acceptance remain pending the parent task's UAT/deployment gate.
