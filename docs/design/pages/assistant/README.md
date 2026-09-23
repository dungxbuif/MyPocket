# AI transaction input and proposal review

Scope: approved [AI-ENTRY-01 design](../../../work/tickets/TICKET-09-01-ENTRY-DETAIL_DESIGN.md) and batch follow-up [AI-ENTRY-02](../../../work/tickets/AI-ENTRY-02-BATCH-ATTACHMENTS-DETAIL_DESIGN.md). Global entry from each primary tab; no separate route. Visual pattern follows the [Money Lover AI entry guide](https://moneylover.zendesk.com/hc/en-us/articles/42320025248409-Add-transactions-faster-and-easier-with-AI-feature): natural-language input, attachment/send actions and a clear grouped result. MyPocket keeps explicit review before ledger writes.

## Composition

| Region/state | Reused base and implementation |
| --- | --- |
| Hold entry | BaseFab (`atoms/BaseNavigation.tsx`), [hold contract](../../atoms/long-press/README.md) |
| Keyboard alternative | BaseButton in manual QuickAddSheet |
| AI entry shell | BaseBottomSheet form presentation; focus trap, Escape, close/return focus |
| One-shot input | AssistantComposer molecule; BaseTextArea, BaseFileUpload button variant, BaseButton send; text and files share one submit. Selected files render as compact thumbnails with per-file shared IconButton removal. |
| Result summary/list | AssistantResultCard molecule; neutral muted surface, count and list supplied by screen |
| Editable transaction proposal | EntryProposalRow + TransactionFields; same wallet, amount, category, optional month-configured jar, note, date, and report controls as manual Add |
| Attachments | BaseFileUpload validates accepted type/count/size in consumer before submission, supports picker and drag/drop, and exposes an accessible drop hint; privacy/helper copy remains visible |
| Loading, missing data, provider failures | StatusMessage; BaseButton for explicit reload. During processing the composer is locked, attachments remain visible with a muted overlay, and the result region uses a subtle pulse/shimmer state. |
| Save/reject/send/reload | BaseButton; disabled/loading and confirmation where needed |

## Behavior

Review-first extraction: AI proposes every recognizable entry even when some fields are missing. Clear copies from overlapping images in the same submission are merged; separate dates/references remain separate and uncertain duplicates stay visible with a review question. It does not check for existing-ledger duplicates. With no clear wallet match, prefill the first supplied wallet; with no wallets, keep an incomplete row for the user to fix. Missing dates stay blank, missing amounts stay zero and unknown categories stay empty. Balance/limit/summary figures are not transactions. Check inferred wallets and suspected duplicates before saving; only explicit approval changes the ledger. There is no application cap on proposal count. The provider response still has a bounded token/byte envelope; a truncated response is rejected atomically instead of showing partial rows.

Short Add opens unchanged manual entry. Hold 500ms opens the AI entry sheet; movement/cancel/blur/unmount cancel; hold consumes release click. A keyboard-accessible AI action is available in manual create mode. The sheet presents a single multi-line composer with optional files, followed by an AI result card containing editable transaction proposals. It does not render user/assistant message bubbles, a conversation log or a new-conversation action.

The API adapter loads capabilities and wallets/categories, then submits one multipart request to `POST /api/v1/ai/entry/process`. After an ambiguous result it performs only `GET /api/v1/ai/entry/requests/{request_id}`; it never automatically resubmits. Proposal edits/decisions use their own APIs. The backend keeps an opaque owner-scoped processing record internally for idempotency, but exposes no session, message, history or latest-record API and does not write conversation messages.

Each pending row exposes prefilled type, amount, wallet, category, optional jar for an eligible expense, local date/time, note and report flag. Jar choices are loaded for the proposal's account-local month; extraction never guesses a jar. Wallet selection is mandatory: the sole supplied wallet is used by default, and multiple-wallet requests choose the most plausible supplied wallet while adding a question when the choice was inferred. No default date/amount is supplied for missing values. Save draft PATCHes with version; approval requires saved values and valid supported kind, wallet, positive integer amount, date and compatible optional category/jar. Changing scope never silently remaps category. Per the owner's follow-up, unknown candidates show “Chưa xác định — chọn loại” and allow explicit manual income/expense classification with no default; that change must be saved before approval. This supersedes the earlier unknown-type lock in the linked design. Transfer remains labeled and locked until paired ledger support exists, with no ordinary expense/income conversion. Approved/rejected rows show terminal state and saved values without controls. Only approval invokes onSaved to refresh wallet balances, transactions and budgets. Rejection requires explicit confirmation.

Optional files: up to 20 JPEG/PNG/PDF, <=5 MiB each. Browser sends raw multipart files, not base64, from either picker or drag/drop. Backend validates the detected file signature/type (the client part MIME header is not trusted), stores originals in environment-qualified private S3, then runs OCR before LLM extraction; the LLM receives only user text plus OCR text and a minimal reference catalog, never file bytes, base64, URLs or image input. OCR failure prevents LLM invocation. Approval links evidence to the transaction and authenticated download is owner/link-scoped. Capabilities expose `files_configured` (OCR plus private storage); if false, attachment selection is disabled while text entry remains available. No provider secrets or mocks in runtime.

Successful AI extraction diagnostics are privacy-safe metadata only: provider model, latency, OCR attachment count, draft count and reply byte length. OCR text, source images, prompts, credentials and monetary values are never logged. The model's user-facing `reply` is persisted with the one-shot result and returned alongside proposals, so an empty proposal list can show a clarification instead of appearing blank; it is not treated as a transport/OCR failure.

The submitted text is also a direct `user_instruction` for selection filters such as a requested month. OCR remains evidence and cannot override schema/ownership/approval policy. Model usage persistence is controlled by `AI_STORE_USAGE`; it is internal and not displayed in this review UI. A future async/polling or SSE transport may show progress without persisting incomplete streamed JSON.

All API paths/types follow the linked authoritative design. Backend owns final validation, authorization, concurrency and ledger writes. All controls/cards use shared bases; consumer classes are layout only.

## Proof and limitations

Browser proof uses deterministic HTTP fixtures and real Chrome interactions. It covers hold vs short-press, no conversation UI, multiline composer/file affordance, grouped result card, edit/save one/save all/delete item, transfer lock, date picker cancel/select/time preservation, budget editor, and grouped transactions. No live transaction is created by the fixture suite. Owner UAT and provider extraction quality remain pending.

Frontend implementation is in review; owner UAT and live provider integration remain pending. Current browser proof checks multipart one-shot submit, absence of session/message calls, editable proposal rows and manual-entry regression. No live transaction was created during fixture tests. Backend private S3 retention, Stage v1 attachment schema, approval links, download route and cleanup command are implemented and tested against local PostgreSQL; live bucket/OCR and browser download UAT remain pending.

Commands from `app/` (run with `rtk`):

| Command | Result/evidence |
| --- | --- |
| `node scripts/ai-entry.test.mjs` | PASS: multipart `POST /process`, raw PDF File field, browser does not set multipart Content-Type boundary, independent request fetch, proposal edit/approve/reject, file limits and no conversation API. |
| `node scripts/ai-entry-browser.test.mjs` | PASS: real Chrome hold/short press, one-shot submit, no `/sessions` or `/messages` calls, editable candidates, save/reject/transfer lock, and calendar/budget/list browser flow. |
| `npm run check:design` | PASS: theme/native-control guard; 96 valid local design links. |
| `npm run test:design` | PASS: 7 guardrail fixtures and shared base contracts, including selectable category tree. |
| `npm run test:transactions` | PASS: 5 transaction/category/wallet rules. |
| `npm run test:transaction-jars` | PASS: the real TransactionFields selector appears only for eligible expenses and offers the active month's jars. |
| `npm run test:calendar` | PASS: account-local month derivation, calendar boundaries and DST field behavior. |
| `node --test scripts/form-logic.test.ts` | PASS: 2 calendar/calculator cases. |
| `npm run build` | PASS: TypeScript and Vite production bundle, 1,756 modules. |
| `git diff --check` (repository root) | PASS. |

Browser tests use real React components and Chrome input events with deterministic HTTP fixtures only in `app/tests/ai-entry.tsx`; runtime uses authenticated services with no mock fallback. Rendered 390×844 screenshot `/tmp/mypocket-ui-pIE4DC/ai-entry-final.png` was compared to Money Lover's composer/result reference: multiline input and file/send actions sit at the bottom; results appear as an editable grouped card; MyPocket's brand/primary theme is retained. Physical-device touch UAT, backend ownership/concurrency/ledger guarantees and live AI/OCR extraction are not proven by frontend HTTP fixtures.

Docs review: the screen spec and base inventory map the composer, upload and grouped result card. The API contract is one-shot and no-history; AI-ENTRY-02 remaining scope is live S3/OCR and browser attachment download UAT. Unsent text/files and unsaved row edits are local until submit/save; approved rows refresh wallet/transaction/budget data.
