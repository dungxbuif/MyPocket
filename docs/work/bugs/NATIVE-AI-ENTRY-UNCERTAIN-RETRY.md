---
artifact_type: bug
id: NATIVE-AI-ENTRY-UNCERTAIN-RETRY
status: verified
owner: shared
updated: 2026-10-08
---

# Native AI Entry: recover uncertain submissions without duplicate requests

## Observed behavior

The web AI-entry flow retains its `request_id` when a submitted POST fails with
a transport error or server error. It blocks another submission and asks the
user to fetch that request's result. A `404` explicitly resolves the ambiguity
and permits a deliberate retry; a found request is rendered without reposting.

Native iOS currently retains `requestID` but does not track ambiguity. After a
transport/server error, the composer remains usable and “Soạn nội dung mới” is
available, clearing the old ID. The server may already have accepted the first
request, so a second request can cause duplicate provider work or duplicate
proposals. The native error also incorrectly claims that no transaction was
saved even though an AI-entry request is only a proposal and its outcome may be
unknown.

## Intended behavior and decision

- Treat transport failures and HTTP/problem responses with status `>= 500`
  after submission as an uncertain request, not as proof that it was rejected.
- Preserve the same `request_id` and original input/files while outcome is
  uncertain. Never automatically POST again.
- Disable editing, attachment changes/removal, analysis, and “new content” while
  the request is uncertain. Show the web-equivalent ambiguity explanation and
  a “Tải lại kết quả” action.
- The reload action performs only `GET /ai/entry/requests/{request_id}`. A found
  pending/completed request stays on that request. A `404` clears uncertainty
  and the old ID, explicitly allowing the user to submit again.
- A successful POST or GET clears uncertainty. A completed request may clear
  the composer after the result is received; proposal decisions remain explicit
  and unchanged.
- Do not say “không có giao dịch nào được lưu” for an uncertain AI processing
  request; proposals are not ledger transactions until the user approves them.

## Scope

- Native AI-entry view and deterministic UI/unit coverage.
- No backend/API/schema/database/provider changes. Existing POST and GET
  endpoints and request-ID semantics are retained.
- Public endpoint docs and API-readable endpoint catalogs do not change because
  the contract is unchanged. Reconcile the assistant design page, port audit,
  validation matrix, and release changelog after implementation.

## Verification plan

1. Add a failing test for retry eligibility/uncertain-state transitions and a
   UI assertion that uncertain state exposes GET recovery while blocking a new
   POST and composer edits.
2. Implement the smallest state/presentation change that meets the behavior
   above; prove GET recovery never invokes POST again.
3. Run focused iOS unit and UI tests on an isolated simulator/API fixture (not
   the owner's production-test simulator), native project/base-component
   guards, and `git diff --check`.
4. Reconcile docs and record exact test evidence. Do not claim the complete
   native port from this focused fix.

## Implemented and verified — 2026-10-08

- Native iOS now preserves the original request ID and draft after ambiguous
  POST failures, blocks editing/attachments/resubmission, and offers GET-only
  recovery. A found request is rendered; GET 404 clears only the stale request
  state and enables an explicit retry with the preserved draft.
- Removed the misleading assertion that no transaction was saved when an AI
  request outcome is unknown. AI proposals still require the existing explicit
  approval before ledger writes.
- RED: the new policy tests initially failed to compile because
  `AIEntryRequestRecoveryPolicy` did not exist yet.
- GREEN: all 146 native unit tests pass, and the isolated iPhone 16 UI test
  `testAIEntryUncertainRequestRequiresGETResolutionBeforeResubmission` passes
  (1/1) at `/tmp/mypocket-ai-entry-recovery-final-20261008.xcresult`.
  Re-run of the entire unit target passes 146/146 at
  `/tmp/mypocket-ai-entry-recovery-unit-full-20261008.xcresult`.
- No API, backend, schema, migration, provider, or production data changed.
