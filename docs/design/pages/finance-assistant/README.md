---
artifact_type: screen_spec
id: FINANCE_ASSISTANT_SCREEN
status: implemented_local
updated: 2026-09-22
---

# Finance Assistant `/assistant`

## Native iOS conversation — 2026-10-14

2026-10-15 persisted history local E2E: backend latest50 + exclusive older cursor,
native chronological rows, Tải tin cũ at top, old page remains visible instead of
scrolling to latest, explicit scroll back and restart preserve saved messages.
Final274/274PASS0skip (273unit+1real DB/API/UI); synthetic fixture cleaned0/0/0.
[Proof](../../../work/tickets/NATIVE-ADVISOR-HISTORY-E2E-20261015.md). Does not
close live cancellation/send races, numeric/privacy or full owner/device UAT.

2026-10-15 approved local correction: exact finance-safe money, unavailable invalid
values, exact count/JSON formatting, masked unknown/malformed fallback. Native
progress-boundary/goal-percent/privacy sensitivity RED→GREEN; final287/287PASS
(279unit+8chatUI, numeric/count/JSON+hosted card checks).
Existing MPMetric/MPText/MPProgressBar/MPAssistantCard/MPWalletRow/MPTransactionRow
reused; no palette/font/shape/base or API change. Physical updated install and
whole-app privacy/visual acceptance remain pending; do not claim those from units.
[Work/review](../../../work/tickets/NATIVE-AI-CARD-NUMERIC-SAFETY-20261015.md).

Historical adaptive gate RED: actual local API overview + landscape keyboard leaves0pt
conversation (fake-session variant7pt); compositor overlaps navigation. Largest
Dynamic Type focused-control tests pass but rendered tab labels/glyphs overflow.
Neither landscape nor full a11y is accepted. One-line input cap did not help and
was reverted. [Evidence/design review](../../../work/tickets/NATIVE-CHAT-ADAPTIVE-20261014.md).
Keyboard chrome/tab accessibility changes are owner-approved and locally tested:
hide tabs while typing, move overview into header, explicit keyboard
dismissal restores tabs, full-label menu at accessibility text sizes. Preserve
draft/source/privacy/base tokens. First approved gate:274/274PASS0skip,268unit+
6UI. Actual landscape focused conversation63pt instead of0/7pt. Owner/device
acceptance and broader accessibility remain outstanding; see work ticket.

Native composition: `MPScreenHeader` → pinned `MPChatOverview` → remaining-height
conversation ScrollView → `MPChatComposer`. Overview uses only exact backend
income/expense values; absent/non-integer data shows unavailable, never invented
zeros. Tap it for the structured data/source card in a shared form sheet.
History scrolling does not move the overview. While keyboard is visible, the
pinned summary is hidden; header chart action opens the same details sheet.
Header keyboard-dismiss action requests focus removal through MPChatComposer,
without clearing draft. Shell removes both tabs and their64pt reservation while
typing, restoring them after keyboard dismissal. Reload remains available after
dismissal. At accessibility text sizes MPBottomNavigation uses MPMenuSelector
with complete destination labels; normal-size tab layout is unchanged.

Composer grows one to four lines then scrolls internally; send/stop targets44pt.
Provider unavailable: allow local drafting, disable submission. Pending work
disables editing; accepted submission clears draft, rejected submission retains
it. Feedback floats above measured composer height without a reserved lane.
Raised Add is hidden on Assistant; Overview/Transactions retain it.

The Assistant tab is the primary entry; More no longer duplicates it. Unsent
input is held only in the authenticated shell's memory and survives tab switches.
Lock/logout removes that shell, and an owner change clears the text. No draft is
written to UserDefaults, sent automatically, or promised after restart. Standalone
programmatic advisor routes keep their own local input. Accepted submission still
clears input through the same binding; rejection retains it. Current navigation
regression196/196 passes locally (188unit+8UI); retained-draft render inspected.
This is not a live-provider, physical-device or release acceptance claim.

Suggestions belong to empty conversation, not another card/duplicate empty state.

Wallet action-card validation reuses the existing shared danger status when the
proposal violates its typed contract; no confirmation button for malformed data.
Never coerce unsupported wallet type to basic or fractional balance to0. Existing
MPAssistantCard/MPText/MPButton geometry/colors and confirmation lifecycle remain.
Local decoder/render fix is complemented by the persisted proposal `action_id`
boundary, but does not provide the full server prepared-action dispatcher or
live provider replay contract.
[Evidence](../../../work/tickets/NATIVE-AI-WALLET-PROPOSAL-VALIDATION-20261014.md).
User prose uses trailing `MPChatMessage`; assistant Markdown/structured cards are
flat. Latest-message identity changes scroll down; prepended history does not.
Existing cancel/history/privacy/API/read-only behavior remains; no streaming.

Proof: SE UI3/3 compact/keyboard/multiline/bubble/prefill; API pinned overview
and details1/1; native units171/171. [Validation](../../../work/VALIDATION_MATRIX.md).
Live provider chat, accessibility/landscape and device acceptance remain gates.
Historical web contract below is archived, not native authority.

The first implementation is a read-only conversation tab. It keeps the conversation in PostgreSQL and stores the conversation ID in the browser only as a pointer for history reload; ledger data remains the backend source of truth.

## Contract

- The composer uses `AssistantComposer mode="advisor"`, hides receipt uploads and keeps the shared `BaseTextArea`/`BaseButton` controls.
- User messages are submitted to `POST /api/v1/ai/advisor/messages` with a client request ID; the backend persists a run and assistant message before returning.
- History loads from `GET /api/v1/ai/advisor/conversation/messages?conversation_id=...`.
- Finance tools in V1 are read-only. A structured wallet-creation proposal is a separate, explicitly user-confirmed UI action through the existing wallet API; its confirmation is disabled while saving and disappears after success, while errors remain retryable. No transaction mutation, bank connection, payment, or arbitrary model URL is exposed.
- When the AI provider is missing, the screen shows a configuration state instead of fake finance numbers.
- Assistant text is rendered through the shared `FinanceAssistantMarkdown` molecule: headings, emphasis, lists, inline code and pipe tables become semantic UI; raw Markdown delimiters are not shown as plain text. The renderer accepts text only and never injects model HTML.
- Structured list cards stay compact like web: transaction results show at most five rows; wallet, goal, budget and jar lists show at most eight. The full server-side total remains visible where supplied; these presentation caps do not paginate or discard backend data.

## Navigation

The current shell has four navigation items—overview, transactions, assistant
and account—with the Add FAB occupying the centered fifth grid slot. Budgets
remain an account destination rather than a primary tab. The Add FAB remains
owned by `BaseFab`; the assistant slot is a normal `BaseNavigationItem`. The
native shell mirrors this contract through `MPShellNavigation` and
`MPBottomNavigation`.

The markdown contract is covered by `node scripts/ai-advisor-markdown.test.mjs`.
