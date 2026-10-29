---
artifact_type: audit
status: in_progress
owner: shared
---

# Native UI system audit and new feedback

Owner requests ui-ux-pro-max-guided common bases, all-screen/presentation audit,
per-screen specs and remaining feedback fixes. iOS only; no web maintenance,
no reference cloning, no release/status updates without owner's acceptance.

Source/navigation inventory: [registry](../../design/platforms/ios-screen-registry.md).
Consolidation proposal: [guideline](../../design/platforms/ios-base-guideline.md).
Work queue: [backlog](../BACKLOG.md); [validation](../VALIDATION_MATRIX.md).

Authenticated prod queue refreshed:42 records, latest8 below, all open and no
screenshots. Runtime service token read only on node, never printed/stored.

| Feedback | Finding / next proof |
| --- | --- |
| f7887563 Finance Assistant | Overview/history share scroll; oversized note-editor composer; pin overview, chat fills rest, suggestions inside chat |
| f4597cd2 Card padding/alignment | Separate horizontal12 from vertical8; one heading/row alignment, remove local overrides |
| 26fcad87 Add investment asset | Always-visible inline editor, nested cards; use common long-form editor flow |
| 2d0811c0 Money/feedback border | Audit MPAmountField vs decimal/count semantics; full feedback sheet gutters/field clipping runtime proof needed |
| aa6ca3cf Duplicate notifications | Overview bell and More notification row duplicate same destination; keep bell |
| d098ff72 Face ID missing | Current source has enrollment/button and tests; reproduce real-device state before changing security. Do not assume solved from source alone |
| e5428521 Duplicate account | Profile and More account row same action; keep profile as single entry |
| 75dee199 Photos before Files | MPDocumentPicker directly opens Files; shared Photos-first source selector needed |

Current checkbox task remains separate: [owner resolution](FEEDBACK-OWNER-RESOLUTION.md).
No raw feedback text is treated as authority to change security/business rules.
No full rendered-screen audit claimed yet. First implement/verify bounded
duplicate navigation and attachment/money fixes, then apply reviewed common
editor/layout rules across registry screens with per-screen evidence.

Deployment contract gap found during status review: deployed auth baseline
12572c87 (and isolated feedback candidate1f6a7f7d) registers category CRUD/wallet
scope but not PATCH /categories/:id/status. Current dirty source/native category
archive/restore depends on that extra route. Migration30's is_active column alone
does not expose it. Treat archive/restore as locally implemented, not confirmed
production-ready; require a separately scoped/tested API rollout before claiming
all existing native functions work on production. Do not bundle unrelated dirty
finance changes into a UI-only deployment.

Bounded fixes implemented: removed duplicate account/notifications rows from
More; profile and Overview bell retained. RED UI caught both duplicates;
GREEN1/1 `/tmp/mypocket-more-duplicates-green-20261014.xcresult`. Existing route
test entry points adjusted; broad route suite not rerun yet. Registry/guideline
and six missing native screen docs added; fresh all-screen visual proof remains
pending. Proposed five-family guideline sent for owner review before broader
cross-screen editor/presentation changes. Photos-first, portfolio, chat and
physical Face ID reproduction are still open, not claimed fixed.

Owner approved the written guideline on 2026-10-14, explicitly retaining the
app palette and requesting whole-app UI/UX improvement. Implementation plan:
[native UI consolidation](../../superpowers/plans/2026-10-14-native-ui-system-consolidation.md).
Six inline tasks cover shared layout/rows, typed fields/note, management editor
presentation, Photos-first attachments, full-height chat, all-screen conformance
and build. Plan self-review completed; written plan awaits owner's review per
writing-plans handoff. Do not describe proposal approval as execution completion.
