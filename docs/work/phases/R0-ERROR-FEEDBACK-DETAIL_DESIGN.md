# R0 — existing-operation error feedback

Status: in_review for this slice; full R0 remains in_progress. Approval: continuation “Tiếp” of the approved functionality-first R0 plan on 2026-09-09. [Verification](../test-verification/R0-ERROR-FEEDBACK-2026-09-09.md), [parent design](R0-DETAIL_DESIGN.md), [inventory](../test-verification/R0-MOUNTED-ACTIONS-2026-09-09.md), [backlog](../BACKLOG.md), [API contract](../../architecture/API.md), [validation](../VALIDATION_MATRIX.md), [release notes](../../releases/CHANGELOG.md).

## Brownfield scope and root cause

`App.tsx` inline AddTransactionSheet/EditTransactionSheet have `try/finally` but no catch for save/archive; promise rejection escapes and form gives no explanation. `AccountScreen.tsx` swallows API-key load/create/revoke errors, and clipboard optional chaining can silently do nothing. Initial key-list failure looks like an empty list. A successful key creation followed by a failed list refresh is conflated with mutation failure.

Touch only these boundaries and their shared feedback base plus tests. Keep existing layout/theme, API/key authorization, accounting and schema. Do not automatically retry mutations or claim uncertain network failures mean the server did not save. Current per-call transaction idempotency keys do not establish safe retry for an ambiguous response; instruct checking history before resubmission. Any partial receipt failure after a transaction succeeds must report the transaction as saved and prevent duplicate resubmission in the same form.

## Implementation

1. Add a shared `OperationError` feedback component: role=alert, Vietnamese action-specific fallback text, safe correlation ID when provided by APIClientError. Do not display arbitrary exception messages or secrets. Optional retry button is only for idempotent read/copy actions, never automatic mutation retries.
2. Add/edit/archive transaction catches keep entered fields and form open, release pending state and show feedback. Clear stale error at explicit next action. Preserve successful mutation before handling attachment-queue errors; no second transaction create after a known success. Buttons touched use a base preserving their existing classes.
3. API-key list tracks loading/error distinctly from empty, with read-only reload. Create/revoke errors retain user input/current list; successful create stores plaintext and list summary before separate refresh. Successful revoke marks the known result and clears any displayed matching plaintext; refresh failure cannot invite revoking or creating twice. Copy has explicit success/unavailable/failure feedback.
4. Test HTTP failures at fetch boundary with real components; add/retry/edit/archive keep values; key-list failure and retry, create failure, success followed by refresh failure, revoke failure, copy failure/success. Do not add backend or authentication changes.

## WebKit diagnosis boundary

Read-only standalone fixture script `frontend/scripts/probe-offline-webkit.mjs` compares Chromium/WebKit, minimal cache-only/current app service worker, automation/page-initiated reload and server-down transport failure. No production access and no MyPocket financial data. Keep original WebKit E2E failure visible; physical Safari UAT remains required. No dependency upgrade or worker rewrite without causal evidence.

## Alternatives and acceptance

Rejected: hiding failures, clearing forms before confirmed success, auto-retrying financial writes, logging keys/raw exception messages, redesigning forms, marking WebKit skipped to obtain a green suite.

Acceptance: reproducible red tests become green; 72 existing tests remain green; live Chrome mobile/desktop regressions unchanged. Failed writes do not trigger success callbacks; confirmed key creation remains accessible if subsequent read fails. API/schema/ADR unchanged; update context/backlog/validation/changelog plus verification record. Full R0 remains open for other inert actions and error paths.

Review refinement: initial/manual/post-mutation key reads share a generation guard for data/error/busy updates across reconnect; delayed clipboard completion is invalidated when its key is replaced or revoked. Both refinements have red-to-green deferred-response tests. Final unit suite 85 pass; Chrome mobile/desktop 26 pass. New behavior uses shared native action/feedback components and preserves existing styling. No release/deploy approval is implied.
