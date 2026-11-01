---
id: ASSISTANT-FIDELITY-AI-20261101
status: in_progress
owner: shared
approval: owner explicitly approved UI fidelity and AI corrections in conversation
---

# Assistant reference fidelity and AI corrections

Links: [backlog](../BACKLOG.md), [context](../../CONTEXT.md),
[screen contract](../../design/pages/finance-assistant/README.md),
[validation](../VALIDATION_MATRIX.md), [changelog](../../releases/CHANGELOG.md).
Phase: bounded corrective work after release; no new release authorized here.

## Evidence and scope

Owner reference: Downloads/stitch_my_pocket (1).zip, containing screen.png,
code.html and DESIGN.md; extracted reference visually inspected on 2026-11-01.
Current simulator reproduces header title squeezed vertically, oversized
composer, transcript enclosed in one panel, and verbose create-jar explanation
duplicating the proposal card. Real request: “Tạo hũ Du lịch, mỗi tháng dành
2 triệu đồng.” Result proposes a month-specific jar; recurring scheduling must
not be implied by this action.

Earlier PDF benchmark: six-file batch failed after OCR; single-file summaries
reported incorrect counts and absent wallets despite five catalog wallets.
These are reproduction evidence, not proof of the current root cause: source
contains existing batching/fallback logic which must be traced before changing.

## Approved implementation

Native: compose slim header, distinct transcript messages/action cards, inline
overview/suggestions and compact growing composer through shared bases. Keep
base typography/colors and hide primary navigation/feedback on Assistant.
Map reference regions in the screen contract; use real data or honest empty
states instead of invented scores/offers. Preserve history, file attachments,
send/cancel, confirmation and dismissal behavior.

Extraction: reconcile response text with validated persisted results; investigate
batch failure without dropping source records or hiding partial extraction.
Advisor: concise grounded prose, avoid repeating card fields, preserve explicit
confirmation and distinguish monthly allocation from recurring instructions.

Alternatives rejected: prompt-only visual/accuracy fixes; replacing the model
without measurements; copying sample financial values into runtime UI.
No planned API/schema/auth/deployment change. Any such need is reported before
expanding scope. No new ADR anticipated for restoration of existing contracts.

## Ownership and verification

Independent workers own extraction backend, advisor backend and native UI.
Coordinator owns integration, review, local runtime checks and master docs.
Regression checks are authored alongside fixes; combined backend/native test
and build run follows integration per owner preference. Compare simulator to
the reference, exercise jar proposal/follow-up and file import with live local
provider. Record actual latency/errors; never count schema tests as AI quality.
Do not mark verified until build, relevant checks and rendered comparison pass.
Owner visual acceptance remains separate. Existing local fixture stays alive.

## Reconciliation pending

Create-jar confirmation regression, 2026-11-01: the live local path returned
HTTP 500 with PostgreSQL SQLSTATE 42804 while marking the proposal applying.
Root cause was `Update("parts", []AdvisorPart)`, which bypassed GORM's typed
`serializer:json` field and bound a SQL record to JSONB. `MarkAction` and
`ReplaceAction` now restrict a typed model update to `Parts`. Before/after real
PostgreSQL tests reproduce 42804 and pass 2/2 after the fix. A fresh live-provider
proposal then confirmed with HTTP 201 and persisted jar id
`deba089d-b2c2-4e87-af54-4cfb22c50c54`; the reloaded native card shows applied
and no confirm button. Persisted rejected/superseded cards are unavailable rather
than deceptively retryable. Native source copy hides timezone/internal IDs and
formats month as `11/2026`; 116/116 DesignSystem tests pass with zero warnings.

2026-11-01 native integration evidence: Debug LOCAL_API_UI build succeeded on
simulator EC80B092-CEF1-4A8E-A5A3-7420298EFD4D with derived data
`/private/tmp/MyPocketRichSimulator`. Installed and opened the populated
Assistant: header is one row, compact composer remains pinned, user/assistant
bubbles and jar proposal are distinct, bottom navigation/feedback absent.
`xcodebuild ... -only-testing:MyPocketIOSTests/DesignSystemTests test` passed
115 tests, zero failures/skips/runtime warnings, verified with xcresulttool:
`Test-MyPocketIOS-2026.11.01_21-55-17-+0700.xcresult`.
`rtk proxy git diff --check` passes. Owner prioritizes native UI completion;
backend changes are not restarted or live-verified. Existing conversation prose
is historical. Reference fidelity still needs greeting/keyboard/card visual
acceptance; do not report full reference parity from this smoke check.

2026-11-01 follow-up: Assistant confirmation cards now use one shared compact
proposal anatomy (semantic catalog icon, friendly title/subtitle, aligned
details and a full-width primary confirmation action). Wallet/category UUIDs
are no longer exposed as labels. Follow-up edits such as “đổi thành 3 triệu”
receive a bounded server-validated pending-proposal context and call
`revise_pending_proposal`; persistence resolves the patch against the original
card, supersedes it and commits the replacement card atomically. The edit does
not write financial data and the replacement still requires confirmation.
Backend and signed LOCAL_API_UI native builds succeeded. The signed app was
installed on the rich-data simulator and opened to the full-screen Assistant;
shell navigation and feedback are absent and the revised applied-jar card is
visually present. Per explicit owner instruction, no additional automated tests
were run; conversational revision and final visual acceptance are left for the
owner's manual session now open in Device Hub.

Update screen spec, context/backlog, validation matrix and changelog with actual
changes and evidence. API/ERD/architecture remain unchanged unless investigation
demonstrates and documents an approved contract change.
