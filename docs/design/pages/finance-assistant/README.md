---
artifact_type: screen_spec
id: FINANCE_ASSISTANT_SCREEN
status: implemented_local
updated: 2026-09-22
---

# Finance Assistant `/assistant`

## Approved native reference correction — 2026-11-01

This section supersedes the historical pinned-overview, continuous-panel,
multi-row composer and visible Feedback descriptions below. Owner explicitly
approved matching the supplied ZIP layout while preserving shared native colors,
font sizes and medium prose weight. Scope: native UI, focused native tests and
this spec only; integration documentation belongs to the coordinator.
Reference inspected: `/tmp/mypocket-assistant-reference.YDtl3S/screen.png`,
`code.html`, `DESIGN.md`. The screenshot is visual input, not financial data.

| Reference region/action/state | Native base mapping and behavior |
| --- | --- |
| Slim back/avatar/title/menu/profile header | `MPChatHeader`, `MPChatAvatar`, `MPIconButton`, `MPOverflowMenu`; title has a single-line normal-size layout, accessible sizes get a separate title row. Reload, overview, keyboard dismissal and confirmed new chat remain in the menu. Profile glyph is decorative identity, not an invented destination. |
| Greeting and financial overview | `MPChatGreeting`, `MPChatOverview`; inline in the scrolling transcript, real income/expense and existing source/details sheet. Missing values remain unavailable; privacy masking applies. No invented name, score, rank or savings claim. |
| Suggested questions | Horizontal `MPButton` row, fills and focuses draft without submitting; hidden when a nonempty draft would be overwritten. |
| Left assistant / right user prose | `MPChatMessage`, `MPChatTextBubble`, `MPChatAvatar`, `MPMarkdown`; separate semantic surfaces and opposite alignment. Shared colors and `MPFont` roles, no export CSS/font sizes. |
| Budget/action cards and financial breakdowns | Existing `AssistantStructuredCard`, `MPAssistantCard`, `MPMetric`, `MPProgressBar`, `MPTransactionRow`; cards remain transcript siblings of prose, without a second prose bubble. Existing typed validators, source/privacy rendering and explicit confirmation remain authoritative. |
| Smart recommendations and offers | Render only server-supplied supported cards; no static subscription/cashback offers, fabricated charts or inferred financial health. |
| Attachments and compact bottom input | `MPChatComposer` with inline `MPDocumentPicker` icon-only variant; one line initially, grows to four (three at accessibility sizes), selected files above only when present. Existing photo/file sheet, bounds, remove, send/stop and uncertain-request locking remain. No unsupported microphone action. |
| Full-screen route | Canvas transcript, pinned composer, existing shell hides both bottom navigation and Feedback; back remains origin-aware. |
| History/loading/errors/retry | Existing paged history and scroll-anchor policy, `MPLoadingView`, `MPStatusMessage`, `MPButton`; no persistence or request lifecycle change. |

Root cause: each independent header action consumed title width; composer base
reserved multiple lines plus a separate action row; `MPChatSurface` forced the
whole transcript into one white panel. Fix the shared anatomy instead of reducing
font size or clipping content. Shared base extension contract: header takes status,
back and menu actions; greeting takes formatted real totals and details callback;
text bubble takes prose/role/privacy; composer accepts an optional attachment
control. All interactive targets remain at least 44pt. Dynamic Type grows content;
focus/dismiss and draft ownership remain unchanged. No API/data/security/runtime
changes or new dependencies.

Verification pending: coordinator will run the combined native build/tests and
compare the rendered screen with the reference at compact phone widths, large
text, light/dark, keyboard and populated history states. This worker must not run
`xcodebuild`, tests, or touch the simulator. Focused regression coverage will check
header/composer measured geometry and existing lifecycle tests remain applicable.
Visual fidelity, provider-backed cards and physical-device acceptance are **not
verified** by this implementation. Links: [native phase](../../../work/phases/PHASE-002-native-ios-swiftui.md),
[base contract](../../system/BASE_COMPONENTS.md),
[validation](../../../work/VALIDATION_MATRIX.md).

Local handoff evidence: `rtk git diff --check` passed. The read-only
`rtk proxy bash mobile-ios/scripts/check-base-components.sh` failed on the
pre-existing `MainTabView.swift:235` raw `.fullScreenCover`; that file was not
changed in this slice. No build, XCTest or simulator command was run. Added
hosted geometry checks for 320/375/430pt headers, accessibility title expansion,
and compact attachment/input growth in `DesignSystemTests`; updated focused
`LaunchTests` for inline overview and menu-based keyboard dismissal. These are
authored checks, not passing runtime evidence. Existing shell policy already
hides navigation and Feedback for Assistant and was left intact. No unsupported
health score, promotional offer, microphone or fabricated timestamp was added.

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

Native run recovery is GET-only: `AdvisorRunObserver` follows a queued or
running `run_id` until a terminal server status, with 1/2/3-second foreground
polling and 3/5/10-second network backoff. It never resubmits the user message.
Run responses may include additive `server_time`, `deadline_at` and
`retry_after_ms`; history `meta` may include `active_run`. Legacy envelopes
remain valid. Capabilities advertise `run_timeout_seconds` and
`grounding_version` when present.

## Native full-screen implementation candidate — 2026-11-01

The native Assistant route is a single full-screen transcript and composer.
The shell feedback bubble is hidden only while this route is visible; Reports,
Overview and other surfaces retain it. The shared composer owns welcome versus
conversation density, Dynamic Type line ranges, 44pt send/stop controls and
focus requests. Typed entry batches are read-only until explicit per-row or
valid-batch decisions, and ordinary proposal editing reuses the shared
transaction fields and jar policy.

Local acceptance coverage now includes run-observer lifecycle/relaunch recovery,
chat submit/cancel state policy, unknown-part read-only behavior, typed entry
batch preview/action state, and canonical transaction-row parity. The focused
native XCTest run passed on the iPhone 16 Simulator; provider, physical-device,
and production evidence remain intentionally unclaimed.

## Contract

- The composer uses `AssistantComposer mode="advisor"`, hides receipt uploads and keeps the shared `BaseTextArea`/`BaseButton` controls.
- User messages are submitted to `POST /api/v1/ai/advisor/messages` with a client request ID; the backend persists a run and assistant message before returning.
- Async submissions use `POST /api/v1/ai/advisor/messages/async`; the native client observes the returned run with `GET /api/v1/ai/advisor/runs/{id}` and can resume an `active_run` discovered from history after foreground/relaunch.
- History loads from `GET /api/v1/ai/advisor/conversation/messages?conversation_id=...`.
- Finance tools in V1 are read-only. A structured wallet-creation proposal is a separate, explicitly user-confirmed UI action through the existing wallet API; its confirmation is disabled while saving and disappears after success, while errors remain retryable. No transaction mutation, bank connection, payment, or arbitrary model URL is exposed.
- When the AI provider is missing, the screen shows a configuration state instead of fake finance numbers.
- Assistant text is rendered through the shared `FinanceAssistantMarkdown` molecule: headings, emphasis, lists, inline code and pipe tables become semantic UI; raw Markdown delimiters are not shown as plain text. The renderer accepts text only and never injects model HTML.
- Structured list cards stay compact like web: transaction results show at most five rows; wallet, goal, budget and jar lists show at most eight. The full server-side total remains visible where supplied; these presentation caps do not paginate or discard backend data. Transaction result/preview rows use the same native reference/grouped card language as `Thêm giao dịch` (shared category glyph footprint, money/date/note hierarchy and action treatment); Assistant conversation chrome must not invent a second transaction-card variant.

## Navigation

The current shell has four navigation items—overview, transactions, assistant
and account—with the Add FAB occupying the centered fifth grid slot. Budgets
remain an account destination rather than a primary tab. The Add FAB remains
owned by `BaseFab`; the assistant slot is a normal `BaseNavigationItem`. The
native shell mirrors this contract through `MPShellNavigation` and
`MPBottomNavigation`.

The markdown contract is covered by `node scripts/ai-advisor-markdown.test.mjs`.
