# AI transaction input and proposal review

Scope: approved [AI-ENTRY-01 design](../../../work/tickets/TICKET-09-01-ENTRY-DETAIL_DESIGN.md) and batch follow-up [AI-ENTRY-02](../../../work/tickets/AI-ENTRY-02-BATCH-ATTACHMENTS-DETAIL_DESIGN.md). Global entry from each primary tab; no separate route. Visual pattern follows the [Money Lover AI entry guide](https://moneylover.zendesk.com/hc/en-us/articles/42320025248409-Add-transactions-faster-and-easier-with-AI-feature): natural-language input, attachment/send actions and a clear grouped result. MyPocket keeps explicit review before ledger writes.

## Composition

| Region/state | Reused base and implementation |
| --- | --- |
| Hold entry | BaseFab (`atoms/BaseNavigation.tsx`), [hold contract](../../atoms/long-press/README.md) |
| Keyboard alternative | BaseButton in manual QuickAddSheet |
| AI entry shell | BaseBottomSheet form presentation; focus trap, Escape, close/return focus |
| One-shot input | AssistantComposer molecule; BaseTextArea, BaseFileUpload button variant, BaseButton send; text and files share one submit |
| Result summary/list | AssistantResultCard molecule; neutral muted surface, count and list supplied by screen |
| Editable transaction proposal | EntryProposalRow + TransactionFields; same wallet, amount, category, optional month-configured jar, note, date, and report controls as manual Add |
| Attachments | BaseFileUpload validates accepted type/count/size in consumer before submission; privacy/helper copy remains visible |
| Loading, missing data, provider failures | StatusMessage; BaseButton for explicit reload |
| Save/reject/send/reload | BaseButton; disabled/loading and confirmation where needed |

## Behavior

Short Add opens unchanged manual entry. Hold 500ms opens the AI entry sheet; movement/cancel/blur/unmount cancel; hold consumes release click. A keyboard-accessible AI action is available in manual create mode. The sheet presents a single multi-line composer with optional files, followed by an AI result card containing editable transaction proposals. It does not render user/assistant message bubbles, a conversation log or a new-conversation action.

The API adapter loads capabilities and wallets/categories, then submits one multipart request to `POST /api/v1/ai/entry/process`. After an ambiguous result it performs only `GET /api/v1/ai/entry/requests/{request_id}`; it never automatically resubmits. Proposal edits/decisions use their own APIs. The backend keeps an opaque owner-scoped processing record internally for idempotency, but exposes no session, message, history or latest-record API and does not write conversation messages.

Each pending row exposes prefilled type, amount, wallet, category, optional jar for an eligible expense, local date/time, note and report flag. Jar choices are loaded for the proposal's account-local month; extraction never guesses a jar. No default wallet/date/amount is supplied for missing values. Save draft PATCHes with version; approval requires saved values and valid supported kind, wallet, positive integer amount, date and compatible optional category/jar. Changing scope never silently remaps category. Per the owner's follow-up, unknown candidates show “Chưa xác định — chọn loại” and allow explicit manual income/expense classification with no default; that change must be saved before approval. This supersedes the earlier unknown-type lock in the linked design. Transfer remains labeled and locked until paired ledger support exists, with no ordinary expense/income conversion. Approved/rejected rows show terminal state and saved values without controls. Only approval invokes onSaved to refresh wallet balances, transactions and budgets. Rejection requires explicit confirmation.

Optional files: up to 3 JPEG/PNG/PDF, <=5 MiB each. Browser sends raw multipart files, not base64. Backend validates and stores originals in environment-qualified private S3, then runs OCR before LLM extraction; the LLM receives only user text plus OCR text and a minimal reference catalog, never file bytes, base64, URLs or image input. OCR failure prevents LLM invocation. Approval links evidence to the transaction and authenticated download is owner/link-scoped. Capabilities expose `files_configured` (OCR plus private storage); if false, attachment selection is disabled while text entry remains available. No provider secrets or mocks in runtime.

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
