# Native UI System Consolidation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Apply five common component families across all native screens while retaining the app palette, financial behavior and owner data.

**Architecture:** Extend existing DesignSystem bases rather than create a second hierarchy. Features keep API/domain state; shared bases own geometry, field semantics and presentation. Migrate each registered screen and opening control with real UI/action evidence.

**Tech Stack:** SwiftUI/iOS17+, current Manrope, existing Go API, XCTest/XCUITest, PostgreSQL fixture.

**Spec:** `docs/design/platforms/ios-base-guideline.md` (owner approved); screen scope in `docs/design/platforms/ios-screen-registry.md`.

## Global Constraints

- Keep current semantic green/neutral palette, Manrope and tokenized text styles.
- Page horizontal gutter16; card horizontal inset12, vertical8. Section gap8, internal row gap4. Card corners12/button8.
- Minimum interactive target44pt; no fixed maximum row height/font shrinking.
- No card-in-card; one section surface with flat internal rows/metrics.
- Long forms use dedicated editors; short choices/confirmations use shared sheet/dialog.
- Financial/AI/auth semantics and existing per-owner isolation must not change.
- No archived web work, new UI package, automatic feedback acceptance or release publication.
- Preserve dirty worktree edits. Existing branch is codex/native-ios-port; never reset or stage unrelated work.

## Review Focus

1. Long Vietnamese labels/accessibility text must grow without hidden amounts/actions (Tasks1/6).
2. Cancelled editor/network failure must not lose draft or silently save/close (Tasks2/3).
3. Investment decimals/counts/percent must not be corrupted by currency formatting (Task2).
4. Photo cancel/partial failure/20-file cap must preserve already selected documents (Task4).
5. Keyboard/history refresh/slow AI response must keep composer and navigation usable (Task5).

## Verification commands

Run from repository root; all shell commands are prefixed with `rtk`.

Unit gate:
`rtk proxy xcodebuild test -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS -destination 'platform=iOS Simulator,id=182D41A2-4FC6-445B-A592-70FDCB5C662E' -derivedDataPath /tmp/mypocket-ui-system-tests -only-testing:MyPocketIOSTests -collect-test-diagnostics never SWIFT_TREAT_WARNINGS_AS_ERRORS=YES -quiet`

UI gate uses the same command with `-only-testing:MyPocketIOSUITests/LaunchTests/<test>` replacing the unit selector; API-backed tests additionally set `SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI` and `MP_API_BASE_URL=http://127.0.0.1:8081/api/v1`. Only disposable local API/DB, never production fixtures. Each run gets a unique resultBundlePath.

Build gate:
`rtk proxy xcodebuild build -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS -configuration Release -destination 'generic/platform=iOS' -derivedDataPath /tmp/mypocket-owner-device-20261014 DEVELOPMENT_TEAM=AU442S3MN3 CODE_SIGN_STYLE=Automatic MP_API_BASE_URL=https://money.dungxbuif.com/api/v1 SWIFT_TREAT_WARNINGS_AS_ERRORS=YES -quiet`

Guards: `rtk proxy bash mobile-ios/scripts/check-base-components.sh`, its `.test.sh`, `verify-project.sh`, and `rtk git diff --check` must exit0.

### Task 1: Common layout, surfaces and rows

**Files:** DesignSystem/Tokens.swift, MPSurfaces.swift, MPFinanceCards.swift, MPNavigation.swift; native feature consumers; DesignSystemTests.swift. Paths are under `mobile-ios/MyPocketIOS/` unless tests.

**Interfaces:** Keep MPSurfaceCard content API; add `MPMetrics.cardHorizontalPadding=12`, `cardVerticalPadding=8`. Keep MPMetric plain content, MPMetricCard only standalone. Shared section/header owns alignment; consumers provide layout, not styling.

- [ ] Add failing rendered tests: one-point surface height17, inset left/right12; shared menu/finance rows >=44; long names and accessibility3 grow without losing amount. Palette assertions remain unchanged.
- [ ] Run unit gate; record actual RED, not a test-helper compile error.
- [ ] Apply asymmetric padding; remove nested MPMetricCard/MPAssistantCard from same surface, replacing inner nodes with MPMetric/plain section content. Remove local heading/negative inset overrides.
- [ ] Run full unit gate and populated Overview/ledger/wallet/category/feedback screenshots on small phone; inspect alignment and final-item reachability.
- [ ] Reconcile TOKENS/BASE_COMPONENTS and touched screen specs; commit only this task's verified changes.

### Task 2: Typed fields and dedicated note editing

**Files:** DesignSystem/MPFields.swift, MPFormat.swift; Features/Transactions/TransactionsView.swift; Features/Secondary/SecondaryFeatureViews.swift; DesignSystemTests/FinanceModelTests/APIClientTests.

**Interfaces:** Keep MPAmountField(text: Binding<String>) money contract. Add MPNumberField(title:text:kind:) using `MPNumberInputKind.decimal/integer/percentage` and shared field styling. Normalized parser returns decimal canonical string or nil, never a currency integer for asset quantity. Note editor uses existing mpFormBottomSheet.

- [x] Add RED cases: quantity1.5 preserved, integer rejects1.5, percentage rejects101/negative, empty remains empty; money1234567 formats consistently without changing saved value. UI note tap opens editable note sheet; details contains report checkbox only.
- [x] Run unit/UI RED selectors; implement typed base and use quantity/interval/percent consumers. Do not change Go payload types or financial validators.
- [x] Move note entry out of details; save/dismiss note updates draft, transaction cancel never sends API write.
- [x] Run full units and API-backed transaction/report exclusion + fractional trade persistence tests. Check actual values, not only text existence.
- [x] Update field/transaction/portfolio/recurring/jar specs; commit deferred under the recorded dirty-worktree ruling (verified local source, not committed/released).

### Task 3: One editor workflow across management screens

**Files:** DesignSystem/MPNavigation.swift; Features/Secondary/SecondaryFeatureViews.swift, Wallets/WalletsView.swift, Assistant/AssistantViews.swift; LaunchTests.swift.

**Interfaces:** Add shared `MPEditorScaffold(title:isBusy:onCancel:fields:footer:)`; it owns one scroll region, header and pinned actions. Existing mpFullScreenCover/mpFormBottomSheet own presentation. Screens provide fields and existing validated save callbacks only.

- [ ] RED UI: portfolio list is not preceded by asset/trade forms; add opens editor, cancel returns with no created asset, failed save stays open. Repeat create/edit/cancel for budgets, jars, recurring and travel.
- [ ] Implement scaffold first, then move inline editors behind existing add/edit actions. Preserve draft state, delete confirmations, eligibility and API success/failure behavior.
- [ ] Use shared date fields for travel dates; keep wire YYYY-MM-DD semantics and account timezone. Use flat asset metrics/history, not card-in-card.
- [ ] Run units and API-backed management CRUD/trade/history/credit tests; verify server state after save, unchanged state after cancel. Capture keyboard/footer reachability.
- [ ] Reconcile each management screen spec and registry entry; commit verified task changes.

### Task 4: Photos-first attachment base

**Files:** DesignSystem/MPAttachmentViews.swift; Features/Assistant/AssistantViews.swift; DesignSystemTests/APIClientTests/LaunchTests.

**Interfaces:** Preserve MPDocumentPicker Files behavior and `onImport([URL])`. Add Photos-first source selector to shared attachment presentation, `MPPickedImage(name:data:mimeType:)` and `onPhotoImport([MPPickedImage])`; existing AI-entry handler maps to MultipartFile. No new upload API or usage limits.

- [x] RED UI: Files-first reproduced; source default/capacity20-currentCount/cancel/removal verified with native UI.
- [x] Source selector/PhotosPicker in DesignSystem; sequential bounded loading and visible per-item errors. JPEG/PNG pass through; supported other representations convert within5MiB or report error.
- [x] Thumbnail/remove and selection retention verified, including real stock-photo append and partial data-loader failure. Mixed JPEG/PDF uses existing multipart adapter with raw-byte contract proof.
- [x] Full187 units plus6UI pass193/193 twice, final `/tmp/mypocket-task4-complete-20261014.xcresult`. Actual native Photos/Files open+cancel automation fulfills platform check; live OCR/device proof remains in full-goal gate, not claimed here.
- [x] AI-entry/feedback/base/registry docs reconciled. Commit deferred under recorded dirty-worktree ruling; completion is verified local source, not release/install.

### Task 5: Full-height financial conversation

**Files:** DesignSystem/MPFields.swift, MPFinanceCards.swift; Features/Assistant/AssistantViews.swift; DesignSystemTests/LaunchTests.

**Interfaces:** Add `MPChatComposer(text:isBusy:isEnabled:onSend:onCancel:)` (optional cancel callback) within existing field family; no new orchestration/message schema. Text grows1–4 lines then scrolls. Overview remains compact pinned content; message ScrollView consumes remaining height.

- [ ] RED layout/UI: scroll chat does not move overview; keyboard does not cover send; history/suggestions share chat region; input not a110pt note editor. Long message and failed send retain text and reachable retry.
- [ ] Implement shared composer and conversation geometry. Reuse MPMarkdown/finance structured cards; move suggestions into empty conversation, avoid extra empty card beneath them.
- [ ] Preserve cancellation, paging, saved conversationID/user scoping, read-only tool boundary, privacy masking and actual provider errors.
- [ ] Run units and API-backed conversation history/state tests plus normal/accessibility keyboard captures. Do not alter provider settings or introduce streaming in this UI task.
- [ ] Reconcile advisor/AI-entry specs; commit verified task changes.

### Task 6: All-screen conformance, review and owner build

**Files:** Remaining registry consumers, mobile-ios/scripts/check-base-components.sh and its tests, all touched design docs, CONTEXT/BACKLOG/VALIDATION_MATRIX and task audit.

**Interfaces:** Tasks1–5 bases are authoritative. Registry becomes per-screen entry/back/save/cancel/error/empty/pending/visual proof ledger; no separate component families added by feature screens.

- [ ] Extend guard with behavioral fixtures for forbidden competing surfaces/presentation overrides; do not claim regex alone proves layout. Remove obsolete base variants only after reference search shows zero consumers.
- [ ] Apply family rules to all registry screens, including auth/account/security/key controls, months/reports/savings/credit and notification states. Preserve profile→account and bell→inbox single entries; remove redundant advisor More entry while keeping deep-link destination support.
- [ ] Run full native units, guards, API-backed critical business flows and route/action sweep; inspect actual populated small-phone captures, Dynamic Type, keyboard, landscape and theme constraints. Record failures/skips honestly.
- [ ] Reproduce newest Face ID report on physical device/account state; preserve protected device credential. If root cause requires new security behavior outside accepted auth design, stop for security approval instead of weakening authentication.
- [ ] One independent whole-change review, then TDD fix pass for Important/Critical findings; no broad UI-completion claim before registry evidence is complete.
- [ ] Finish already-approved feedback API activation32 after SSH approval; verify health/route/schema and unchanged data. Do not deploy unrelated dirty finance code. No additional migrations in this UI plan.
- [ ] Build final production-API Release once, strict signature check, install without uninstall, launch owner's iPhone. Keychain/SSH approvals remain manual; no password requests in chat. Record owner acceptance before release notes/feedback resolution changes.

## Plan self-review and handoff

All five families, eight fresh feedback items and registered screens map to
Tasks1–6. Shared interface order: layout/fields before editor consumers;
attachment and composer consume existing field/presentation families; final
sweep depends on all five. No new domain APIs, palette/font replacements or
bank integrations. Every Review Focus class has an explicit owning test above.

Execution method: inline/native, preserving existing worktree. Written plan is
owner approved; execution starts with Task5 (priority chat), then remaining tasks.
The full goal includes offline/AI actions/RAG/APNs and production gaps tracked in
[completion matrix](../../work/tickets/NATIVE-FULL-COMPLETION-20261014.md). Previously accepted
feedback feature/deployment remains separate and may finish independently.
