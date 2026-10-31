# MyPocket Typography, Chat and Icons Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Giữ phong cách MyPocket chủ đang thích, làm chữ gọn/dễ đọc, tăng vùng hội thoại hữu ích, mở rộng vùng viết đúng lúc và dùng icon nhẹ, dễ nhận diện.

**Architecture:** Thay đổi qua DesignSystem và shell hiện có, không custom lại từng screen. Typography, chat, Hũ và icon là các đợt có nghiệm thu riêng; quyền gửi phản hồi, lifecycle AI và API-backed jar behavior vẫn do owner hiện tại quản lý.

**Tech Stack:** Swift 6, SwiftUI/UIKit, iOS 17+, XCTest/XCUITest, vector asset catalog.

**Spec:** [NATIVE-CHAT-COMPOSER-20261031-DETAIL_DESIGN](../../work/tickets/NATIVE-CHAT-COMPOSER-20261031-DETAIL_DESIGN.md).

> **Reconciliation 2026-11-01:** The shared typography, adaptive full-screen
> Assistant, Hũ surfaces, shared transaction preview cards, Phosphor Regular
> icon catalog and base guards are implemented. Native tests pass 442/442 and
> the debug build is installed on simulator `EC80B092-CEF1-4A8E-A5A3-7420298EFD4D`.
> Dark mode and large Dynamic Type remain explicitly unclaimed acceptance
> gates; no screen-specific styling exception was added.

## Global Constraints

- Người dùng đã chuyển từ lập plan sang thực thi; chỉ triển khai các thay đổi UI/UX trong scope, không đổi API/DB/auth/deploy.
- Giữ phong cách card, màu sắc và tổng thể bản đang cài trên máy ảo.
- Giữ bong bóng chat/phản hồi. Không xóa, thay bằng menu, hoặc làm mất tính năng.
- Chữ mặc định: body/chat 15pt regular, phụ 13pt regular, nhãn nhỏ 12pt regular, tiêu đề 17pt semibold, tiền dòng 16pt medium/tabular, số dư 34pt semibold/tabular.
- AI Assistant và mọi preview/result item của AI phải dùng cùng card language với màn `Thêm giao dịch`: shared reference/grouped surface, inset/divider/row geometry, icon footprint, money hierarchy và action row; không tạo card/bubble variant riêng cho cùng một transaction.
- Cỡ chữ là ở Dynamic Type mặc định; không fixed-height/cắt chữ để đạt mật độ. Hit target tối thiểu 44pt.
- Giữ binding draft, không tự gửi gợi ý, không đổi dữ liệu iconKey/systemKey, không đổi API/DB/auth/model/deploy.
- Native portrait; không thay thanh tab/card toàn app, không thêm dependency renderer SVG.
- Đóng gói subset icon có nguồn/version/license, không copy asset Money Lover.
- Bản trên simulator `MyPocket-iPhone-16` là baseline người dùng thích; dùng simulator preview riêng khi thực thi.

## Review Focus

1. Chữ Việt dài, số tiền âm/nhiều chữ số và Dynamic Type lớn: không bị cắt; Task 1/5 kiểm tra.
2. Máy 375pt và bàn phím mở: feedback 52pt + send 44pt đều bấm được; Task 2 kiểm tra hình học.
3. Gõ nhiều dòng, bấm gợi ý/mở feedback/chuyển tab: draft không bị mất hoặc gửi ngoài ý muốn; Task 2 kiểm tra.
4. Tin mới và tải lịch sử khi đang đọc phía trên: không giật xuống, không mất mốc đọc; Task 3 kiểm tra.
5. Category tự tạo, key cũ/không biết, dark mode: icon không biến thành chấm đen hoặc sai danh tính; Task 4 kiểm tra.

## Hiện trạng và thứ tự

Chỉ plan/docs được viết trong lượt lập plan. Có nháp cũ chưa build/test ở bốn
file `Tokens.swift`, `MPFields.swift`, `AssistantViews.swift`, `MainTabView.swift`.
Không coi nháp là thiết kế cuối: composer luôn lớn, chữ 17pt trong nháp và bubble
đẩy lên trên không đáp ứng spec mới. Khi bắt đầu thực thi, review từng hunk và
giữ mọi thay đổi của người khác; không dùng reset/checkout để dọn toàn file.

Đợt A: Task 1. Đợt B: Task 2–3. Đợt C: Task 4 (Hũ) và Task 5 (icon) theo
write-scope riêng. Task 6 nghiệm thu tổng thể. Các agent chỉ sửa vùng file được
giao; phần tích hợp shared base và kiểm thử cuối do agent điều phối rà lại.

### Task 1: Typography thực, baseline và mật độ nội dung

**Files**
- Modify: `mobile-ios/MyPocketIOS/DesignSystem/Tokens.swift`, `MPText.swift`, `MPSurfaces.swift`, `MPFinanceCards.swift`, `MPNavigation.swift`, `MPFields.swift`.
- Test: `mobile-ios/MyPocketIOSTests/DesignSystemTests.swift`; `mobile-ios/MyPocketIOSUITests/LaunchTests.swift`.
- Docs: `docs/design/system/TOKENS.md`, `docs/design/platforms/ios-base-guideline.md`.

**Interfaces**
- Consume: existing `MPFont`, `MPTextStyle`, shared row/header components.
- Produce: explicit native semantic fonts in `MPFont`, plus `static let amount: Font`; stable public role names. Remove the misleading ignored-size helper rather than retaining numeric values that do nothing.

- [ ] Capture baseline Overview/Transactions/Assistant with identical local fixture data, default text and keyboard closed; also capture one Assistant draft with keyboard open. Record device viewport, font category and visible complete rows/lines.
- [ ] In `Tokens.swift`, map display→title/28, balance→largeTitle/34, headline→headline/17, body/chatBody→subheadline/15, bodySmall→footnote/13, labelSmall→caption/12, amount→callout/16, currency→title3/20. Apply spec weights. Keep system Dynamic Type rather than custom fixed UIFont sizes.
- [ ] Wire amount role to money rows only; remove inherited semibold from ordinary prose. Read each consumer before changing so privacy/formatting and symbol alignment remain intact.
- [ ] Run existing `DesignSystemTests`; add only layout/behavior coverage for long Vietnamese names and large/negative amounts, not tests that merely compare token constants. At accessibility size assert full strings remain present and layout grows.
- [ ] Compare same-content screenshots on the three pilot screens and inspect Wallet/Transaction editor/Account for shared-style regressions. Record measured density change, including cases that did not improve; 20–30% is a goal for chat reading, not a fabricated pass threshold.
- [ ] Update typography contract with actual default sizes and scaling behavior; hand off preview A before combining with icon changes.

### Task 2: Adaptive composer and preserved feedback bubble

**Files**
- Modify: `mobile-ios/MyPocketIOS/DesignSystem/MPFields.swift`, `MPNavigation.swift`.
- Modify: `mobile-ios/MyPocketIOS/Features/Assistant/AssistantViews.swift`, `Features/Shell/MainTabView.swift`.
- Test: `mobile-ios/MyPocketIOSUITests/LaunchTests.swift`.
- Docs: `docs/design/pages/finance-assistant/README.md`, `docs/design/system/BASE_COMPONENTS.md`.

**Interfaces**
- Consume: Task 1 fonts; `MPChatComposer` draft/lifecycle props; shell `feedbackPresented`, `feedbackCapturePending`, snapshot action.
- Produce: `enum MPChatComposerMode { case welcome, conversation }`; add `var mode: MPChatComposerMode = .conversation` to `MPChatComposer`.
- Produce: `MPChatFooterLayout.feedbackClearance: CGFloat = 60` (52pt bubble + 8pt gap), centralized in `MPNavigation.swift`. Shell owns exactly one `MPFeedbackBubble`; composer reserves its trailing footprint. Do not duplicate capture/network ownership in AssistantView.

- [ ] Add `testAssistantAdaptiveComposerKeepsFeedbackAndSendReachable`: for welcome, nonempty draft and populated conversation, assert feedback/send hit targets ≥44pt, no pairwise intersection with input, and all controls inside the viewport/above keyboard. On 375pt assert conversation height ≥44pt while focused.
- [ ] Update the existing compact composer test to verify welcome has room for three default-size lines and ongoing conversation has a smaller minimum. Do not retain the old single-line-only bound as the product requirement.
- [ ] Implement welcome 3-line minimum; conversation 1–4; normal focused 2–4; accessibility 1–2. Derive line/font height, not fixed pt editor height. Sending indicator and stop action reuse the same control slot; no extra permanent label row.
- [ ] Replace the old draft's upward feedback offset with a protected horizontal footer: reserve trailing 60pt beside composer, align bubble to its lower edge. Keep it visible while typing on Assistant unless feedback overlay/capture owns the interaction; leave other routes unchanged. Use keyboard safe area and shell's existing tab hiding.
- [ ] Keep Send/Stop 44pt and circular feedback 52pt. Preserve their accessibility identifiers/labels; feedback opens the existing form. Add a visible-focus affordance without changing the app palette.
- [ ] Align Assistant transaction/result previews with the Add Transaction reference card: reuse the existing transaction row, category glyph, amount/date hierarchy and shared action controls; conversation chrome may remain conversational, but its finance content must not use a one-off card language.
- [ ] Extend `testAssistantFeedbackBubbleDoesNotCoverComposerAndDraftSurvivesDisabledSend`: enter ≥6 lines, open/close feedback, dismiss keyboard and switch tabs; assert exact draft retained, no request submitted while unavailable, feedback returns to its dedicated slot.
- [ ] Run the existing largest-text test, suggestion-prefill test and composer test on iPhone 16 and SE. Inspect dark mode at rest and with keyboard. Do not install over the liked baseline.

### Task 3: More room to read, no forced scrolling

**Files**
- Modify: `mobile-ios/MyPocketIOS/Features/Assistant/AssistantViews.swift`.
- Modify: `mobile-ios/MyPocketIOS/DesignSystem/MPFinanceCards.swift`, `MPText.swift` (only named base variants if necessary).
- Create: `mobile-ios/MyPocketIOS/Features/Assistant/AssistantScrollPolicy.swift`.
- Test: `mobile-ios/MyPocketIOSTests/AssistantScrollPolicyTests.swift`, `mobile-ios/MyPocketIOSUITests/LaunchTests.swift`.

**Interfaces**
- Consume: `MPChatOverview`, existing overview sheet, messages IDs/older cursor, draft binding and mode from Task 2.
- Produce: `AssistantScrollPolicy.shouldFollowLatest(distanceFromBottom: CGFloat, isLoadingOlder: Bool) -> Bool`; follow only when distance ≤64pt and not loading older. Track the visible message anchor for prepend restoration; do not infer user position solely from message count.

- [ ] Move summary access to an always-available header icon when overview exists; keep the same sheet, exact data and privacy. Remove the permanently pinned summary from transcript height.
- [ ] Render welcome copy + one horizontal suggestion row only when history and draft are empty. Press pre-fills and focuses without sending. Keep full accessible text, allow horizontal reveal, and do not use ellipsis as the only accessible label.
- [ ] Add unit boundary cases: distance 0/64→true; 65→false; loading older→false regardless of distance. Add UI fixture messages long enough to require scroll, with no live model dependency.
- [ ] Replace unconditional latest-ID scrolling with policy-gated scroll. When not following, show shared-base “Tin mới” action; tapping goes to bottom. Prepending history preserves first visible message ID and relative position. Do not add fake streaming or change polling.
- [ ] Add `testAssistantNewMessagesPreserveReadingAnchor` and `testAssistantOlderHistoryPreservesReadingAnchor`; assert current message still visible after updates, “Tin mới” works, and following at bottom still works. Keep retry, warnings, references and confirmation cards visible and accurate.
- [ ] Capture populated chat before/after using exactly the same answer. Record useful visible lines at rest and keyboard-open; avoid gaining space by hiding financial caveats or data.

### Task 4: Đưa Hũ thành một flow rõ ràng, không chỉ là một dòng trong Mở rộng

**Files**
- Modify: `mobile-ios/MyPocketIOS/Features/Secondary/SecondaryFeatureViews.swift`.
- Modify: `mobile-ios/MyPocketIOS/Features/Overview/OverviewView.swift` only when needed to clarify the existing month/Hũ entry point.
- Test: `mobile-ios/MyPocketIOSTests/JarsMonthsReferenceTests.swift`, `mobile-ios/MyPocketIOSUITests/LaunchTests.swift`.
- Docs: `docs/design/pages/jars/README.md`, `docs/design/pages/month-detail/README.md`, relevant screen registry entries.

**Interfaces**
- Consume: existing `FinanceAPI` jar/month endpoints, `JarMonthSummaryModel`, `JarCumulativeSummaryModel`, `JarsView`, `MonthsView` and transaction/AI jar selectors.
- Produce: a visually coherent Hũ flow with explicit month list, add/edit configuration, cumulative detail and month-summary handoff; preserve API contracts, stable jar IDs and existing accessibility identifiers.

- [ ] Treat Hũ as a first-class feature in the UI review: verify entry from Tổng quan’s `Hũ tháng`, Tài khoản’s `Hũ chi tiêu`, and Tháng detail all land on the selected month rather than an unrelated tab or generic Mở rộng row.
- [ ] Keep the existing behavior: real API loading/error/empty states; monthly allocation modes; active vs historical rows; remove-from-month semantics; cumulative report; no fabricated balance/carryover; no backend or model changes.
- [ ] Make the list hierarchy scannable: month selector → monthly figures → jar rows/progress → unassigned spending → cumulative history. Use shared cards/rows/tokens only; do not create a second jar card system or hard-code a new palette.
- [ ] Make add/edit a clear native form with the amount/percent field as the primary focus, visible allocation mode, preserved draft on validation/API failure, and a stable 44pt Save/Cancel path. Add an explicit confirmation for removal if the current shared action has none.
- [ ] Add/adjust UI coverage for empty month, populated month, add/edit, historical jar, cumulative selection, and returning to the originating route. Verify long Vietnamese names, overspend warning, masked money and Dynamic Type without clipping.
- [ ] Audit all jar touchpoints: transaction editor and AI proposal editor must continue offering only eligible active month jars; changing month/type must clear or preserve values according to the existing contract.

### Task 5: Money Lover-inspired icon family, with compatibility

**Files**
- Create: `docs/design/references/mypocket-icon-study/README.md` and `comparison.html` (explicit visual-reference bundle).
- Create: `mobile-ios/MyPocketIOS/DesignSystem/MPIconCatalog.swift`.
- Create: selected vector entries under `mobile-ios/MyPocketIOS/Assets.xcassets/Icons/`; source manifest and `mobile-ios/MyPocketIOS/Resources/Phosphor-LICENSE.txt`.
- Modify: `DesignSystem/MPIcon.swift`, `MPButton.swift`, `MPNavigation.swift`, `MPSurfaces.swift`, `MPFields.swift`, `MPFinanceCards.swift`.
- Test: `mobile-ios/MyPocketIOSTests/DesignSystemTests.swift` plus UI catalog/ledger/editor coverage in `LaunchTests.swift`.

**Interfaces**
- Produce: `enum MPIconID: String, CaseIterable`; `enum MPIconStyle { case regular, duotone }`.
- Produce: `MPIconCatalog.icon(forLegacyKey: String) -> MPIconID` and `assetName(for: MPIconID, style: MPIconStyle) -> String`.
- Extend `MPIcon` with an initializer consuming semantic icon ID/style while retaining the system-name initializer for scoped platform controls during migration. Both apply label/hidden semantics. Persisted API icon keys do not change.

- [ ] Build a 20-icon comparison board from the spec: current SF glyph, Phosphor Regular/Duotone, Lucide. Show at actual 20/24/32pt beside 15pt Vietnamese text and on light/dark backgrounds. Record source commit/version, verified asset names and licenses. Select Phosphor as proposed default, subject to actual preview legibility; do not automatically use Thin.
- [ ] Inventory every current catalog option and `MPTransactionPresentation` exact/prefix mapping. Give all supported keys a mapped asset, including credit/goal/basic wallets. Keep explicit old aliases such as `bag.fill`; unknown keys get a visible tag icon.
- [ ] Pin/download only selected upstream vectors; preserve regular/duotone distinction and export reproducibly to iOS vector PDF assets if required. Keep two tintable layers for duotone so dark appearance stays readable. Include source/license manifest and no runtime network fetch.
- [ ] Use `MPIconCatalog` in tab/action, wallet, transaction, category picker and structured-card bases. Category recognition must remain identical across picker, editor and ledger. Selected tabs use existing green/selection treatment, not an arbitrary switch to black filled glyphs.
- [ ] Test all `MPIconID` assets resolve, all known keys/aliases map, unknown key maps to tag, and choosing an icon still round-trips the existing API key. Render light/dark; assert actionable icons retain labels and ≥44pt targets.
- [ ] If a Vietnamese category lacks a recognizable stock glyph, add a small original vector in the same grid/weight. Do not mix downloaded styles ad hoc or copy Money Lover's illustrations. Record custom provenance.
- [ ] Present preview C on Overview, Transactions, AI and category picker. Keep global palette/card structure from baseline.

### Task 6: Verification, reconciliation and handoff

**Files**
- Update: the spec linked above, `docs/CONTEXT.md`, `docs/work/BACKLOG.md`, `docs/work/VALIDATION_MATRIX.md`, `docs/design/platforms/ios-screen-registry.md`, relevant native page/base/token contracts.
- Release notes: local-candidate record only after verification; no production/physical acceptance claim.

- [ ] At execution time create dedicated iPhone 16 and SE preview simulators and use their returned UUIDs as `CHAT_PREVIEW_PHONE_ID`/`CHAT_PREVIEW_SMALL_ID`. Use a dedicated temp derived-data directory `CHAT_PREVIEW_DD`; retain the original installed app and fixture baseline.
- [ ] Build/run focused tests with local-only configuration. Example command (variables are populated from actual created devices/paths, never guessed):

```sh
rtk proxy xcodebuild -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS \
  -destination "id=$CHAT_PREVIEW_PHONE_ID" -derivedDataPath "$CHAT_PREVIEW_DD" \
  -only-testing:MyPocketIOSTests/DesignSystemTests \
  -only-testing:MyPocketIOSTests/AssistantScrollPolicyTests \
  -only-testing:MyPocketIOSUITests/LaunchTests/testAssistantAdaptiveComposerKeepsFeedbackAndSendReachable \
  -only-testing:MyPocketIOSUITests/LaunchTests/testAssistantLargestTextKeepsMultilineComposerAndConversationAboveKeyboard \
  -only-testing:MyPocketIOSUITests/LaunchTests/testAssistantFeedbackBubbleDoesNotCoverComposerAndDraftSurvivesDisabledSend \
  -only-testing:MyPocketIOSUITests/LaunchTests/testAssistantShowsSuggestedQuestionsBeforeFirstMessage \
  -only-testing:MyPocketIOSUITests/LaunchTests/testAssistantNewMessagesPreserveReadingAnchor \
  -only-testing:MyPocketIOSUITests/LaunchTests/testAssistantOlderHistoryPreservesReadingAnchor \
  test MP_API_BASE_URL=http://127.0.0.1:8080/api/v1
rtk proxy bash mobile-ios/scripts/check-base-components.sh
rtk proxy bash mobile-ios/scripts/verify-project.sh
rtk proxy git diff --check
```

- [ ] Run compact-phone UI subset with the second UUID; inspect default/largest text, light/dark, keyboard on/off, long Vietnamese names, 12-digit/negative money, disabled/busy/retry and empty/populated chat. Reuse established simulator signing settings; report exact command and failures rather than silently dropping tests.
- [ ] Full native unit target after shared typography/icon changes; Release build for asset inclusion. Web build/tests are not applicable to this SwiftUI-only slice. Do not broaden to backend/model evals unless later scope changes.
- [ ] Record screenshots and density measurements, no-overlap geometry, test counts and residuals in the work item/validation matrix. Check source references and license inclusions. Request owner visual review of preview; this is not automatically a request to release.
- [ ] Reconcile design contracts only to verified behavior; leave unimplemented/deferred items marked as such. No new ADR needed for cosmetics; document icon provenance in the catalog contract. Stop at a reviewable local candidate.

## Deferred, not silently included

- Full-screen writing editor; model/system-prompt changes for shorter answers.
- New attachment/voice/model selection buttons, AI streaming, backend features.
- General redesign/removal of cards; production release or replacing the liked simulator build.

## Plan self-review

Spec coverage: typography→1; welcome/composer/bubble/drafts→2; density/summary/
suggestions/reading anchor→3; Hũ entry/list/editor/cumulative/assignment→4;
icon research/assets/key compatibility→5; device, accessibility, measurement and
docs→6. The Review Focus cases have owning checks. No task assumes that the old
runtime draft passed tests. Draft sizes are based on line counts, bubble placement
is explicit, icon choice is a proposal supported by a comparison deliverable, and
Hũ behavior remains API-backed rather than mocked.
