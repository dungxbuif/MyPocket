# AI entry chat and proposal review

Scope: approved [AI-ENTRY-01 design](../../../work/tickets/TICKET-09-01-ENTRY-DETAIL_DESIGN.md). Global entry from each primary tab; no separate route. No visual reference supplied. Frontend ownership only; coordinator owns work/global reconciliation.

## Composition

| Region/state | Reused base and implementation |
| --- | --- |
| Hold entry | BaseFab (`atoms/BaseNavigation.tsx`), [hold contract](../../atoms/long-press/README.md) |
| Keyboard alternative | BaseButton in manual QuickAddSheet |
| Chat shell | BaseBottomSheet form presentation; focus trap, Escape, close/return focus |
| Messages, proposals | SurfaceCard, Text, Heading; EntryProposalRow molecule |
| Editable fields | FormField, BaseTextInput, BaseSelect, BaseCheckbox |
| Attachments | [BaseFileUpload](../../atoms/file-upload/README.md) |
| Loading, missing data, provider failures | StatusMessage; BaseButton for explicit reload |
| Save/approve/reject/send/new conversation | BaseButton, disabled/loading while requests run |

## Behavior

Short Add opens unchanged manual entry. Hold 500ms opens chat; movement/cancel/blur/unmount cancel; hold consumes release click. A keyboard-accessible AI action is available in manual create mode.

Every chat mount GETs capabilities and latest session, wallets and categories. Messages/proposals come only from real API responses. Empty latest session is created on first submit or by explicit “Cuộc trò chuyện mới”. New conversation asks confirmation that unsent text, selected files and unsaved proposal edits will be discarded. On confirmation it POSTs `/sessions` with `{}` and switches to the returned session, clearing text, files and uncertain-submission state only on success. Cancel or API failure preserves the existing session and edits. Initial loading, any active message/proposal/session request and server processing disable creation; a synchronous shared request guard also prevents duplicate activation. Creation is available even if AI is unconfigured. Backend caps sessions at 20 user messages; users explicitly start the next session, with no automatic rollover. Latest-session resume still works on reopening. Show processing/error states honestly; explicit reload recovers session after ambiguous request failure without resubmitting. Never automatically retry message POST. Saved session history survives close/reopen; unsent text/files/local row edits do not.

Each pending row exposes prefilled type, amount, wallet, category, local date/time, note and report flag. No default wallet/date/amount is supplied for missing values. Save draft PATCHes with version; approval requires saved values and valid supported kind, wallet, positive integer amount, date and compatible optional category. Changing scope never silently remaps category. Per the owner's follow-up, unknown candidates show “Chưa xác định — chọn loại” and allow explicit manual income/expense classification with no default; that change must be saved before approval. This supersedes the earlier unknown-type lock in the linked design. Transfer remains labeled and locked until paired ledger support exists, with no ordinary expense/income conversion. Approved/rejected rows show terminal state and saved values without controls. Only approval invokes onSaved to refresh wallet balances, transactions and budgets. Rejection requires explicit confirmation.

Optional images: up to 3 JPEG/PNG, <=5 MiB each, validated before FileReader conversion. Submit sends raw base64 under `images` with MIME/name, request ID, text and local IANA timezone. Images are transient OCR inputs, not retained receipt attachments; extracted text remains in session. No provider secrets or mocks in runtime. AI unconfigured disables sending, OCR unconfigured disables images; persisted draft review remains available.

All API paths/types follow the linked authoritative design. Backend owns final validation, authorization, concurrency and ledger writes. All controls/cards use shared bases; consumer classes are layout only.

## Proof and limitations

Follow-up regression: browser tests first failed with both `unknownSelectable: false` and `newConversationAction: false`. Root causes were classification eligibility derived from the original draft and no explicit session-creation control. The implementation now checks edited classification, keeps an independent transfer lock, and uses the existing session API under the shared request guard. Browser tests cover both income and expense recovery through PATCH then versioned approval; new-session confirmation/cancel, API failure retaining text/files/row edits, clearing uncertain state on success, duplicate prevention, disabling during initial loading and PATCH/approve/reject/create, sending to the new ID and latest-session resume. Main/coordinator owns live-browser verification; this worker uses isolated HTTP fixtures and does not mutate live account data.

Frontend implementation verified for coordinator review; owner UAT and live provider integration remain pending. Tests were authored before the new modules. Initial red: `node scripts/ai-entry.test.mjs` could not load the absent long-press implementation. Browser regression red subsequently proved two real defects: native timers called with the scheduler object raised Chrome's `Illegal invocation`, preventing hold activation; optional hold logic suppressed ordinary FAB clicks after blur. Fixed by invoking native timers through closures and bypassing hold suppression when no hold callback is supplied. Both browser assertions are green. SSR fixtures needed browser storage setup; CDP keyboard fixtures needed Enter character text. These test harness adjustments do not change runtime behavior.

Commands from `app/` (run with `rtk`):

| Command | Result/evidence |
| --- | --- |
| `npm run test:ai` | PASS: timer boundary/cancel; draft missing/invalid fields and category compatibility; versioned PATCH/approve/reject payloads; no direct ledger endpoint; provider problem response; upload limits and terminal markup. Chrome interactions prove short/held pointer actions, release suppression, movement/pointercancel/window blur/unmount, keyboard alternative, latest GET on reopening, prefilled list, unsaved approval lock, saved-version approval, rejection, terminal persistence, onSaved only after approve, selected image base64, provider error and unconfigured state. |
| `npm run check:design` | PASS: controls/base ownership, tokens, 88 local design links. |
| `npm run test:design` | PASS: 7 guardrail tests plus existing base contract suite. |
| `npm run test:transactions` | PASS: 5 existing ledger/category/wallet logic tests; Node emits its existing experimental type-stripping warning. |
| `npm run build` | PASS: guardrails, TypeScript, Vite production bundle (1744 modules). |
| `git diff --check` (repository root) | PASS. |

Browser test uses real React components and Chrome input events with deterministic HTTP fixtures only in `app/tests/ai-entry.tsx`; runtime uses authenticated services with no mock fallback. Browser runner defaults to macOS Chrome, overridable through `CHROME_BIN`. `AI_ENTRY_SCREENSHOT` optionally records a screenshot; 390×844 chat rendering was inspected for readable shared controls, card layout, and sheet header. No supplied visual reference exists for fidelity comparison. Physical-device touch UAT, backend ownership/concurrency/ledger guarantees and live AI/OCR extraction are not proven by frontend HTTP fixtures.

Docs review: this screen and new upload/hold base contracts match implemented composition, validation, events, privacy copy and states. No API/schema/architecture change was introduced by this worker. Manual transaction create/update/delete logic is unchanged; only the optional AI entry button was added. Unsent text/files and unsaved row edits are discarded on close; persisted server messages/drafts are reloaded. Global validation/backlog/release updates and final ticket acceptance remain with the coordinator under the assigned ownership boundary. No commit created.
