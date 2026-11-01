# Native Navigation UX Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Chuyển màn đúng ngữ cảnh mobile; Assistant toàn màn hình ẩn navigation chính và bubble, đóng về đúng nguồn.

**Architecture:** Shell sở hữu tab, path từng tab và presentation. Feature sử dụng typed routing và base hiện có; native NavigationStack/toolbar/sheet sở hữu chuyển màn và safe area. Migrate từng flow, giữ deep links và business contracts.

**Tech Stack:** SwiftUI, UIKit bridge khi cần root edge-dismiss, XCTest/XCUITest; không thêm dependency.

**Spec:** [NATIVE-NAVIGATION-UX-20261101](../../work/tickets/NATIVE-NAVIGATION-UX-20261101-DETAIL_DESIGN.md).

## Global Constraints

- Assistant là màn riêng, ẩn navigation chính và bubble; finding yêu cầu hiện bubble là false positive.
- Giữ base style/palette/card/font; mọi input tiền dùng MPAmountField/MPAmountKeypad.
- Portrait; target chạm tối thiểu 44pt; Dynamic Type; không yêu cầu dark mode.
- Không đổi API/DB/auth/financial semantics hoặc tự deploy.
- Tab mới Tổng quan/Giao dịch/Báo cáo/Tài khoản là đề xuất chờ duyệt.
- Đóng Assistant không tự cancel task; dữ liệu UI thuộc owner hiện hành.
- Implementation mới chưa được bắt đầu trong lượt planning này.

## Review Focus

- Đổi tab và mở overlay không làm mất filter/scroll/draft: Task 1, 2, 4.
- Hai entry Assistant không tạo hai session/presentation: Task 2.
- Deep link đến trong khi sửa form không làm mất draft: Task 1, 4.
- Record bị xóa hoặc logout khi đang xem detail: Task 3.
- Interactive dismissal cancel không pop hai tầng hoặc gửi lại mutation: Task 2, 4.

## Task 1: Shell routing và native containers

**Files:** Create `mobile-ios/MyPocketIOS/Core/Routing/MPNavigationState.swift`;
modify `Features/Shell/MainTabView.swift`, `DesignSystem/MPNavigation.swift`,
`Core/Routing/AppRouteCoordinator.swift`; test `MyPocketIOSTests/NavigationStateTests.swift`.
Các path không có prefix dùng gốc `mobile-ios/MyPocketIOS/` (test dùng gốc `mobile-ios/`).

**Interfaces:** `MPRootTab: Hashable { overview, transactions, reports, account }`;
`MPNavigationState` @MainActor ObservableObject sở hữu selectedTab và
`[MPRootTab: [AppRoute]]`; `push(_ route: AppRoute)`, `pop()`,
`receiveExternal(_ route: AppRoute)`. Reuse AppRoute để giữ URL compatibility.
Thêm typed detail cases vào AppRoute khi Task 3 cần; không tự thêm URI parser.

- [ ] Viết `testTabSwitchPreservesPaths`, `testPopRemovesOnlyTop`, `testExternalRouteWaitsForEditorDismissal`; assert độc lập từng path và pending route.
- [ ] Tạo state/router, map deep links sang tab canonical theo spec; logout clear owner state.
- [ ] Dùng TabView + NavigationStack riêng từng tab; giữ view-model identity/scroll anchors. Add mở presentation từ origin, bỏ đổi selection sang transactions để mở form.
- [ ] Native toolbar adapter dùng base title/actions; bỏ duplicate custom header trong flow migrated. Xóa hit-test blocker/clipping chỉ sau khi geometry tương ứng được chứng minh.
- [ ] Kiểm tra không mất navigation state qua đổi tab, không có tab click xuyên vào content.

## Task 2: Một Assistant presentation thống nhất

**Files:** Create `Features/Assistant/AssistantPresentation.swift`;
modify `Features/Assistant/AssistantViews.swift`, `Features/Shell/MainTabView.swift`,
`Features/More/MoreView.swift`, `DesignSystem/MPNavigation.swift`;
test `MyPocketIOSUITests/LaunchTests.swift`.

**Interfaces:** `presentAssistant(draft: String?)` và `dismissAssistant()` trên
MPNavigationState. Presentation session giữ origin tab/path, session ID và draft
theo owner; source không thay selection. AssistantPresentation sở hữu một stack
nội bộ. Existing report handoff chỉ prefill theo contract, không tự submit.

- [ ] Viết `testAssistantClosesToEachOrigin` cho Tổng quan/Báo cáo/Tài khoản;
  assert tab/bubble/Add không tồn tại, dismiss trả đúng path/filter.
- [ ] Hợp nhất các entry vào một full-screen presentation; loại `.assistant`
  destination của More khỏi render riêng nhưng giữ compatibility adapter cho deep link.
- [ ] Root một Close, child native Back; shared edge-dismiss bridge chỉ nhận
  gesture cạnh khi ở root và không có dirty editor. Test hoàn tất/cancel gesture.
- [ ] Draft/session/scroll giữ khi đóng rồi mở trong phiên; đóng không gọi cancel
  AI. Logout reset. Bàn phím không che composer; không phục hồi bubble tại Assistant.
- [ ] Giữ attachment/action integration findings riêng; không đánh dấu xong chỉ
  vì chuyển thành full-screen. Reference parity là gate riêng còn mở.

## Task 3: Detail flows trong stack đang dùng

**Files:** Modify `Features/Overview/OverviewView.swift`,
`Features/More/MoreView.swift`, `Features/Reports/MonthlyReportsView.swift`,
`Features/Secondary/SecondaryFeatureViews.swift`, `Features/Transactions/TransactionsView.swift`,
`Core/Routing/AppRoute.swift`; create `Core/Routing/MPRouteContext.swift`;
test `MyPocketIOSUITests/LaunchTests.swift`.

**Interfaces:** `MPRouteContext: Hashable` giữ walletID, jarID, categoryID,
account-local month và origin; giá trị optional chỉ có khi flow cần.
Detail cases: wallet detail, scoped ledger, transaction detail, jar detail,
asset detail, report category drilldown. Dùng ID domain hiện có, không copy model.

- [ ] Map toàn bộ bảng flow trong spec thành checklist có route owner và API
  dependency; đánh dấu missing backend capability, không giả đủ dữ liệu.
- [ ] Ví → giao dịch của ví mở ledger scoped trong cùng stack; giữ tab ledger
  tổng độc lập. Báo cáo drilldown truyền cùng kỳ/loại tiền/scope báo cáo.
- [ ] Hũ/Tháng preserve account month; portfolio giữ asset/unit/source; Back pop
  một tầng. Notifications/categories/budget/recurring/travel/credit/profile dùng
  cùng header/stack contract, không dựng router riêng.
- [ ] Viết UI cases `testWalletLedgerReturnsToWallet`,
  `testReportDrilldownKeepsScope`, `testMonthJarBackPreservesMonth`,
  `testDeletedDetailCanGoBack`; API fixture chứng minh scope đúng, không chỉ title.
- [ ] Verify không còn callback chi tiết ép đổi tab Tài khoản; entry trực tiếp
  vào tab canonical vẫn hỗ trợ external deep links.

## Task 4: Editor, picker và dismissal dùng chung

**Files:** Modify `DesignSystem/MPNavigation.swift`, `DesignSystem/MPFields.swift`,
`Features/Transactions/TransactionsView.swift`, `Features/Secondary/SecondaryFeatureViews.swift`,
`Features/Assistant/AssistantEntryBatchView.swift`; test `MyPocketIOSUITests/LaunchTests.swift`.

**Interfaces:** Shared editor presentation nhận origin/context, dirty/submitting
state, `onSaved`/`onCancel`; consumes MPNavigationState của Task 1. Không duplicate
save orchestration; reuse canonical editor và recovery hiện có.

- [ ] Long editor dùng large sheet và stack riêng; selector ngắn dùng sheet
  thích ứng, selector bên trong editor push cùng editor stack.
- [ ] Áp dụng một dismissal policy: dirty → confirm discard/continue; submitting
  → protected until known result/recovery. Cancel không reset parent filter.
- [ ] AI preview edit mở cùng transaction editor, quay lại đúng card; nếu entry
  metadata/contract chưa đủ, ghi dependency thay vì nối bằng dữ liệu đoán.
- [ ] Viết `testDirtyEditorSwipeDoesNotLoseDraft`, `testSaveReturnsToOriginOnce`,
  `testPickerBackKeepsAmountExpression`, `testDeepLinkWaitsForDirtyEditor`;
  assert một mutation tối đa và amount/quantity đúng parser hiện có.

## Task 5: Gộp, kiểm chứng và reconcile

**Files:** `mobile-ios/MyPocketIOSUITests/LaunchTests.swift`,
`docs/design/platforms/ios-screen-registry.md`, `docs/design/platforms/ios-base-guideline.md`,
`docs/design/pages/finance-assistant/README.md`, `docs/CONTEXT.md`,
`docs/work/BACKLOG.md`, `docs/work/VALIDATION_MATRIX.md`, `docs/releases/CHANGELOG.md`.

- [ ] Gộp toàn bộ flow trước lượt test tổng hợp theo yêu cầu owner; mỗi task
  vẫn định nghĩa test từ trước. Không cho nhiều reviewer chạy cùng simulator.
- [ ] Chạy `rtk proxy bash mobile-ios/scripts/verify-project.sh`,
  `rtk proxy bash mobile-ios/scripts/check-base-components.sh`,
  `rtk proxy bash mobile-ios/scripts/check-icon-catalog.sh`, `rtk proxy git diff --check`.
- [ ] Liệt kê destinations bằng `rtk proxy xcodebuild -showdestinations -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS`;
  chạy `rtk proxy xcodebuild -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS -destination 'platform=iOS Simulator,id=<available-UDID>' test` với signing settings đã hoạt động trong repo. Record exact resolved command/result, không dùng UDID đoán.
- [ ] Populated UAT nhỏ/lớn, keyboard/Dynamic Type/VoiceOver, save/cancel/error,
  origin restoration và swipe; chụp cùng data trước/sau. Không coi static guards
  là proof cho safe-area hoặc hit targets.
- [ ] Ba reviewer độc lập cho **từng task/flow**: product contract, routing/state,
  UX/runtime. Bảng coverage từng hàng phải có ba verdict và evidence; uncovered
  còn pending. Reviewers phải đọc owner correction về bubble trước khi audit.
- [ ] Reconcile docs/ADR navigation và ghi missing backend/reference gates rõ.
  Runtime pass/owner acceptance chỉ cập nhật sau bằng chứng. Không deploy trong plan.

## Thứ tự và phân công

Task 1 → Task 2/3 → Task 4 → Task 5. Shell/router/MPNavigation do một integrator
sở hữu; agents có thể làm screens disjoint sau interface freeze. Không để hai
agents cùng sửa SecondaryFeatureViews.swift hoặc LaunchTests.swift. Có thể chia
review độc lập song song; kiểm thử simulator theo hàng đợi.

## Self-review

Spec flow matrix được phân vào Tasks 2–4; origin/deep-link/logout/dismissal có
test owners cụ thể. Plan chỉ native navigation, không thay scope backend audit.
Status: draft; planning hoàn thành, implementation chưa chạy; không claim test pass.
