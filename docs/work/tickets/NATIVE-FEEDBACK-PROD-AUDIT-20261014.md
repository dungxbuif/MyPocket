---
artifact_type: feedback_audit
status: in_review
source: authenticated GET /api/v1/agent/feedback?limit=100
---

# Production feedback audit

Fresh2026-10-15 authenticated queue:48records, all status=open, response meta=null.
Runtime FEEDBACK_AGENT_TOKEN read/used only on vm100, never copied or printed.
Newest six beyond the42snapshot include portfolio edit/delete/price, sheet/close,
edge-swipe Back, fixed lower-left Feedback+Account Add, row separation, and AI jar
creation/full user-domain tools. [Current source/design assessment](NATIVE-FEEDBACK-NEW-20261015-DETAIL_DESIGN.md).
New requirements are not marked fixed from older tests. No feedback state writes.

Reconciled2026-10-15:34-record table below is a historical authenticated snapshot,
not the latest queue count. Dedicated note editor is now implemented/tested
locally; offline outbox/engine/foreground/fault flows are also implemented locally,
but production capability still404. Current residuals/rollout boundaries are in
[native reconciliation](NATIVE-RECONCILIATION-20261015.md). No feedback status was
changed by this documentation correction.

The runtime service token was used on the production node, without exposing or
copying it into local files. The API returned all 34 records, all status=open,
all screenshot_available=false. Earlier 404/401 checks do not establish a missing
API or an empty queue. No feedback status/changelog was mutated. These local
engineering classifications are not production status updates.

Links: [implementation](NATIVE-FEEDBACK-CONSISTENCY-AUTH-DETAIL_DESIGN.md),
[rollout](NATIVE-DEVICE-AUTH-ROLLOUT-20261014.md), [proof](../VALIDATION_MATRIX.md).

| ID prefix | Report | Current disposition/evidence |
| --- | --- | --- |
| 62e31193 | AI response formatting | Native MPMarkdown renderer exists; exact live-provider prose is not revalidated by deterministic UI tests |
| afee0f86 | AI creating wallets/actions | General V2 write-tool orchestration remains planned; read-only advisor safety boundary is not silently removed |
| 726943be | Notifications, money input | Native inbox/read API and shared formatted amount keypad exist; APNs on free Personal Team is not an implemented remote-push delivery guarantee |
| 1a59aa1b | Assistant product help/RAG | Product-knowledge/release-aware RAG remains an independent AI feature, not claimed fixed by UI polish |
| e11d18e6 | Wallet selection goes to ledger | API UI proof verifies selected scope routing |
| 6b8b32e1 | Required transaction category | Migration28/database and handler contracts already implemented; production was already29 before this rollout |
| 89cca866 | Optional transaction jar | Existing native jar selector/eligibility and payload contracts |
| 7495cd2f | Shared calendar/money controls | DesignSystem bases and guard/tests; no competing native amount/calendar primitive added |
| 52362c1c | Recurring UI style | Uses shared field/amount/date/card bases; common compact tokens now apply |
| 08246b4b | Chat composer pinned | Native composer is a sibling of history ScrollView, not a message-list child |
| 06fb8f5b | Portfolio workflow | Existing asset, manual price and persisted trade history API/UI implementation; earlier native API proofs retained |
| 7ddf77f3 | Native/offline | Native and encrypted outbox/receipt/foreground sync are implemented/tested locally; production rollout/cache/full offline acceptance remain open |
| 559242fe | Assistant shadow/layout | Shared card aliases now flat,12pt padding; no Assistant-only card shape |
| 2eb68f7b | Missing icons | Central category/custom-icon mapping and same MPTransactionRow used across Overview/ledger/wallet/Assistant; unit + API UI proof |
| 5d6557d2 | Navbar too tall | Shared56pt navbar, minimum44pt controls; compact-device proof |
| b63c7727 | Feedback occupies scroll space | Removed lane; explicit viewport above navbar + bubble overlay; compact UI RED/GREEN |
| 4ca27075 | Month summary failure | API UI recalculates owner ledger and checks real state |
| 31b7bcf7 | Account opens AI entry | Viewport sizing/selected-tab shielding regression; actual account tap proof |
| 26e78c7f | Privacy eye only masks total | Shared financial-mask environment/policies and API UI coverage |
| 1b3c25d8 | Newest populated week | Owner/wallet-scoped current-or-newest-past week policy tested |
| 196110ec | Wallet selector in header/persist | Duplicate title removed, selector moved to header, per-owner local preference retained; UI proof |
| 019b2e9c | Local cache smoothness | Durable user-scoped financial cache/offline synchronization is still a separate open feature; no unsafe shared response cache added |
| 534ad994 | Card/content too large | Shared12pt padding, compact typography/metrics, horizontal profile; rendered-size/Dynamic Type tests |
| 4b9a11a5 | Report time/timezone copy | Inclusive account-local day labels, no midnight/exclusive-boundary jargon |
| f82b0108 | Item inset/size | Plain row inset0, shared transaction glyph36/vertical4/min44, same row across screens |
| fd10b2de | Income/expense transition | Shared segmented animation respects Reduce Motion |
| c9fb1fb7 | Thin font | Shared medium body/secondary and semibold emphasis; Manrope retained |
| ecf02904 | Attach feedback image | Shared photo choose/remove/replace, thumbnail and bounded PNG; field/control and multipart tests |
| 1e12aaa1 | Common modal/select components | Native form sheets/selectors use shared bases; native system menu behavior is retained where specified, not an arbitrary new per-screen UI |
| a738ad4a | Tree line geometry | One shared category tree,40/32pt glyphs, connector stops before icon; prior UI proof retained |
| fc8cae65 | Variable-value assets | Existing portfolio price/trade UI; not a new transaction/wallet accounting model |
| 1db32af3 | Note/detail/report checkbox | Checkbox default/persistence fixed; dedicated shared note sheet now implemented with UI RED→GREEN and API note readback. Owner UAT remains |
| 8eeb848f | Transaction row wallet name | Note or local date, no redundant wallet name; shared-row API UI proof |
| 31ef1877 | Face ID after logout | Independent protected enrollment + fresh login; real local API UI8/8 batch includes enroll/logout/Keychain login; auth-only backend deployed |

Do not claim all historical34 reports fixed: older feature requests (AI V2 writes,
product RAG, full offline/cache rollout, push delivery) remain open. Note-editor
implementation is no longer missing. The owner-facing finish list covers bubble/viewport, density,
Face ID, authenticated queue access and physical build/install. Physical UAT
and release acceptance remain gates; no changelog is published yet.
