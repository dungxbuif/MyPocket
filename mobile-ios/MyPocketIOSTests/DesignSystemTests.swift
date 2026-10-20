import SwiftUI
import UIKit
import XCTest
import ImageIO
import UniformTypeIdentifiers
@testable import MyPocketIOS

@MainActor
final class DesignSystemTests: XCTestCase {
    func testApplicationSupportsPortraitOnly() {
        let delegate = PortraitOnlyApplicationDelegate()
        XCTAssertEqual(
            delegate.application(UIApplication.shared, supportedInterfaceOrientationsFor: nil),
            .portrait
        )
    }

    func testOverlappingBackBridgesRestoreOriginalDelegate() throws {
        let original = NSObjectGestureDelegate()
        let first = MPInteractiveBack.Controller()
        let second = MPInteractiveBack.Controller()
        let navigation = UINavigationController()
        navigation.setViewControllers([UIViewController(), first, second], animated: false)
        navigation.loadViewIfNeeded()
        let gesture = try XCTUnwrap(navigation.interactivePopGestureRecognizer)
        gesture.delegate = original
        first.install()
        second.install()
        first.restore()
        XCTAssertTrue(gesture.delegate === second, "An older route cannot detach the current route's gesture owner.")
        second.restore()
        XCTAssertTrue(gesture.delegate === original, "Route replacement must restore UIKit's original delegate, not a detached bridge.")
        first.install()
        XCTAssertTrue(gesture.delegate === first, "Returning to a prior route reattaches that route only.")
        first.restore()
        second.restore()
        XCTAssertTrue(gesture.delegate === original, "Returning to root and repeated cleanup preserve the original delegate.")
    }

    private final class NSObjectGestureDelegate: NSObject, UIGestureRecognizerDelegate {}

    func testInteractiveBackRefusesRootEditorAndTransition() {
        XCTAssertTrue(MPInteractiveBackPolicy.canBegin(depth: 2, isBlocked: false, isTransitioning: false))
        XCTAssertFalse(MPInteractiveBackPolicy.canBegin(depth: 1, isBlocked: false, isTransitioning: false))
        XCTAssertFalse(MPInteractiveBackPolicy.canBegin(depth: 2, isBlocked: true, isTransitioning: false))
        XCTAssertFalse(MPInteractiveBackPolicy.canBegin(depth: 2, isBlocked: false, isTransitioning: true))
    }
    func testTransactionRowsHaveBreathingRoomAndExpandForAccessibility() {
        let row = MPTransactionRow(title: "Ăn uống", subtitle: "", amount: "−50.000 ₫")
        let compact = UIHostingController(rootView: row.environment(\.dynamicTypeSize, .large))
            .sizeThatFits(in: CGSize(width: 343, height: 1000)).height
        XCTAssertGreaterThanOrEqual(compact, 56, "Rows need 6pt separation above and below their 44pt content.")
        let expanded = UIHostingController(rootView: row.environment(\.dynamicTypeSize, .accessibility3))
            .sizeThatFits(in: CGSize(width: 343, height: 1000)).height
        XCTAssertGreaterThan(expanded, compact)
    }
    func testFloatingFeedbackUsesStableLowerRightDockAndNeverReservesObstacleLane() throws {
        let viewport = CGRect(x: 0, y: 0, width: 375, height: 611)
        let ordinary = try XCTUnwrap(MPFeedbackBubbleLayout.frame(in: viewport, clearance: 60, obstacles: []))
        XCTAssertEqual(ordinary, CGRect(x: 307, y: 491, width: 52, height: 52))
        let rightControl = CGRect(x: 311, y: 495, width: 44, height: 44)
        let shifted = try XCTUnwrap(MPFeedbackBubbleLayout.frame(in: viewport, clearance: 60, obstacles: [rightControl]))
        XCTAssertEqual(shifted, ordinary)
        let leftControl = CGRect(x: 20, y: 495, width: 44, height: 44)
        let raised = try XCTUnwrap(MPFeedbackBubbleLayout.frame(in: viewport, clearance: 60, obstacles: [rightControl, leftControl]))
        XCTAssertEqual(raised, ordinary)
        XCTAssertTrue(viewport.contains(raised))
        XCTAssertEqual(raised.size, CGSize(width: 52, height: 52))
        XCTAssertNil(MPFeedbackBubbleLayout.frame(in: .zero, clearance: 0, obstacles: []))
    }
    func testPhotoBatchContinuesAfterInvalidAndUnavailableItemsWithinRemainingCapacity() async throws {
        let renderer = UIGraphicsImageRenderer(size: CGSize(width: 4, height: 4))
        let first = try XCTUnwrap(renderer.image { context in
            UIColor.red.setFill(); context.fill(CGRect(x: 0, y: 0, width: 4, height: 4))
        }.pngData())
        let last = try XCTUnwrap(renderer.image { context in
            UIColor.blue.setFill(); context.fill(CGRect(x: 0, y: 0, width: 4, height: 4))
        }.pngData())
        var calls: [Int] = []
        var active = 0
        var maximumActive = 0
        let result = await MPDocumentImagePolicy.loadBatch(count: 5, capacity: 4) { index in
            calls.append(index)
            active += 1; maximumActive = max(maximumActive, active)
            defer { active -= 1 }
            try await Task.sleep(for: .milliseconds(2))
            switch index {
            case 0: return first
            case 1: throw URLError(.notConnectedToInternet)
            case 2: return Data("invalid image".utf8)
            case 3: return last
            default: XCTFail("Must not load beyond remaining capacity"); return nil
            }
        }
        XCTAssertEqual(result.images.map(\.data), [first, last])
        XCTAssertEqual(result.errors.count, 2)
        XCTAssertEqual(calls, [0, 1, 2, 3])
        XCTAssertEqual(maximumActive, 1, "Cloud/media loads must be sequential, not20 concurrent buffers.")
        let full = await MPDocumentImagePolicy.loadBatch(count: 2, capacity: 0) { _ in
            XCTFail("A full selection must not load another photo"); return first
        }
        XCTAssertTrue(full.images.isEmpty)
        XCTAssertTrue(full.errors.isEmpty)
    }
    func testDocumentPhotosPreserveValidJPEGAndPNGBytes() throws {
        let renderer = UIGraphicsImageRenderer(size: CGSize(width: 64, height: 48))
        let image = renderer.image { context in
            UIColor.blue.setFill(); context.fill(CGRect(x: 0, y: 0, width: 64, height: 48))
        }
        let jpeg = try XCTUnwrap(image.jpegData(compressionQuality: 0.8))
        let png = try XCTUnwrap(image.pngData())
        let first = try XCTUnwrap(MPDocumentImagePolicy.normalized(jpeg, name: "receipt.heic"))
        XCTAssertEqual(first.data, jpeg)
        XCTAssertEqual(first.mimeType, "image/jpeg")
        XCTAssertEqual(first.name, "receipt.jpg")
        let second = try XCTUnwrap(MPDocumentImagePolicy.normalized(png, name: "receipt.png"))
        XCTAssertEqual(second.data, png)
        XCTAssertEqual(second.mimeType, "image/png")
    }

    func testUnsupportedPhotoRepresentationConvertsToBoundedJPEG() throws {
        let image = UIGraphicsImageRenderer(size: CGSize(width: 128, height: 96)).image { context in
            UIColor.green.setFill(); context.fill(CGRect(x: 0, y: 0, width: 128, height: 96))
        }
        let raw = NSMutableData()
        let destination = try XCTUnwrap(CGImageDestinationCreateWithData(raw, UTType.tiff.identifier as CFString, 1, nil))
        CGImageDestinationAddImage(destination, try XCTUnwrap(image.cgImage), nil)
        XCTAssertTrue(CGImageDestinationFinalize(destination))
        let normalized = try XCTUnwrap(MPDocumentImagePolicy.normalized(raw as Data, name: "scan.tiff"))
        XCTAssertEqual(normalized.name, "scan.jpg")
        XCTAssertEqual(normalized.mimeType, "image/jpeg")
        XCTAssertEqual(Array(normalized.data.prefix(3)), [255, 216, 255])
        XCTAssertLessThanOrEqual(normalized.data.count, 5 << 20)
        XCTAssertNotNil(UIImage(data: normalized.data))
        var oversizedSource = raw as Data
        oversizedSource.append(Data(count: (25 << 20) + 1 - oversizedSource.count))
        XCTAssertNil(MPDocumentImagePolicy.normalized(oversizedSource, name: "large.tiff"))
    }

    func testDocumentPhotoPolicyRejectsInvalidAndOversizedWithoutFabricatingAnImage() {
        XCTAssertNil(MPDocumentImagePolicy.normalized(Data(), name: "empty.jpg"))
        XCTAssertNil(MPDocumentImagePolicy.normalized(Data("not an image".utf8), name: "bad.png"))
        XCTAssertNil(MPDocumentImagePolicy.normalized(Data(count: (25 << 20) + 1), name: "large.heic"))
    }

    func testDocumentPhotoUploadBoundAcceptsFiveMiBButRejectsOneMoreByte() throws {
        let image = UIGraphicsImageRenderer(size: CGSize(width: 2, height: 2)).image { context in
            UIColor.red.setFill(); context.fill(CGRect(x: 0, y: 0, width: 2, height: 2))
        }
        var png = try XCTUnwrap(image.pngData())
        png.append(Data(count: (5 << 20) - png.count))
        XCTAssertEqual(try XCTUnwrap(MPDocumentImagePolicy.normalized(png, name: "boundary.png")).data, png)
        png.append(0)
        XCTAssertNil(MPDocumentImagePolicy.normalized(png, name: "too-large.png"))
    }
    func testTransferBackedCreditPaymentKeepsSpecificTitleAndNeutralMoneySemantics() {
        let payment = MPTransactionRow(type: "income", rawAmount: 7500,
            systemKey: "income_transfer_in", occurredAt: "2026-10-14T00:00:00Z",
            timeZone: .gmt, masked: false, transferID: "pair", creditKind: "payment")
        XCTAssertEqual(payment.title, "Thanh toán")
        XCTAssertEqual(payment.amountTone, .neutral)
        XCTAssertEqual(payment.amount, "+7.500\u{00a0}₫")
        let incoming = MPTransactionRow(type: "income", rawAmount: 7500,
            systemKey: "income_transfer_in", occurredAt: "2026-10-14T00:00:00Z",
            timeZone: .gmt, masked: false, transferID: "pair")
        XCTAssertEqual(incoming.title, "Nhận chuyển ví")
    }
    func testMoneyGroupingNeverChangesTheBoundIntegerDraft() async throws {
        var raw = "1234567"
        let host = UIHostingController(rootView: MPAmountField(title: "Amount", text: Binding(get: { raw }, set: { raw = $0 }), presentation: .transaction))
        let window = UIWindow(frame: CGRect(x: 0, y: 0, width: 375, height: 667))
        window.rootViewController = host
        window.makeKeyAndVisible()
        defer { window.isHidden = true }
        host.view.layoutIfNeeded()
        try await Task.sleep(for: .milliseconds(50))
        func fields(in view: UIView) -> [UITextField] {
            (view as? UITextField).map { [$0] } ?? view.subviews.flatMap { fields(in: $0) }
        }
        let field = try XCTUnwrap(fields(in: host.view).first)
        XCTAssertEqual(field.text, "1.234.567")
        XCTAssertEqual(raw, "1234567")
        XCTAssertEqual(MPFormat.parsePositiveInteger(raw), 1_234_567)
        XCTAssertEqual(MPFormat.money(1_234_567), "1.234.567\u{00a0}₫")
    }
    func testTypedNumberNormalizationPreservesDecimalsAndRejectsAmbiguousDrafts() {
        XCTAssertEqual(MPNumberInputKind.decimal.normalized(" 1,25000000 "), "1.25000000")
        XCTAssertEqual(MPNumberInputKind.decimal.normalized("1.5"), "1.5")
        XCTAssertEqual(MPNumberInputKind.decimal.normalized("0.00000001"), "0.00000001")
        XCTAssertEqual(MPNumberInputKind.decimal.normalized(""), "")
        for value in ["1,000.5", "1..5", "1e2", "-1", "+1", "١.٥", "1 000", ".5", "1."] {
            XCTAssertNil(MPNumberInputKind.decimal.normalized(value), value)
        }
        XCTAssertEqual(MPNumberInputKind.integer.normalized("12"), "12")
        XCTAssertNil(MPNumberInputKind.integer.normalized("1.5"))
        XCTAssertNil(MPNumberInputKind.integer.normalized("1,5"))
        XCTAssertNil(MPNumberInputKind.integer.normalized("1.000"))
    }

    func testTypedPercentageNormalizationChecksBoundsWithoutFloatingPointRounding() {
        XCTAssertEqual(MPNumberInputKind.percentage.normalized("12,50"), "12.50")
        XCTAssertEqual(MPNumberInputKind.percentage.normalized("100.00"), "100.00")
        XCTAssertEqual(MPNumberInputKind.percentage.normalized("0"), "0")
        XCTAssertEqual(MPNumberInputKind.percentage.normalized(""), "")
        for value in ["101", "-1", "100.00000000000000000000000000000000000000001", "NaN", "Infinity"] {
            XCTAssertNil(MPNumberInputKind.percentage.normalized(value), value)
        }
    }
    func testDisabledTextFieldCannotRemainEditableThroughUIKitBridge() async throws {
        let host = UIHostingController(rootView: MPTextField(title: "Disabled input", text: .constant("draft")).disabled(true))
        let window = UIWindow(frame: CGRect(x: 0, y: 0, width: 375, height: 667))
        window.rootViewController = host
        window.makeKeyAndVisible()
        defer { window.isHidden = true }
        host.view.frame = CGRect(x: 0, y: 0, width: 343, height: 120)
        host.view.setNeedsLayout()
        host.view.layoutIfNeeded()
        try await Task.sleep(for: .milliseconds(50))
        func fields(in view: UIView) -> [UITextField] {
            (view as? UITextField).map { [$0] } ?? view.subviews.flatMap { fields(in: $0) }
        }
        let inputs = fields(in: host.view)
        XCTAssertEqual(inputs.count, 1)
        XCTAssertFalse(inputs.first?.isEnabled ?? true, "SwiftUI disabled must reach the actual UIKit input.")
    }
    func testFeedbackImageNormalizationRejectsInvalidDataAndOversizedSource() {
        XCTAssertNil(MPImageAttachmentPolicy.normalizedPNG(Data("not an image".utf8)))
        XCTAssertNil(MPImageAttachmentPolicy.normalizedPNG(Data(count: MPImageAttachmentPolicy.maxSourceBytes + 1)))
    }

    func testFeedbackImageNormalizationProducesBoundedPNG() throws {
        let format = UIGraphicsImageRendererFormat()
        format.scale = 1
        let source = UIGraphicsImageRenderer(size: CGSize(width: 2400, height: 1600), format: format).jpegData(withCompressionQuality: 0.8) { context in
            UIColor.blue.setFill()
            context.fill(CGRect(x: 0, y: 0, width: 2400, height: 1600))
        }
        let png = try XCTUnwrap(MPImageAttachmentPolicy.normalizedPNG(source))
        XCTAssertEqual(Array(png.prefix(8)), [137, 80, 78, 71, 13, 10, 26, 10])
        XCTAssertLessThanOrEqual(png.count, ScreenSnapshot.maxPNGBytes)
        XCTAssertLessThanOrEqual(try XCTUnwrap(UIImage(data: png)).size.width, 1600)
    }

    func testProfileCardUsesCompactIdentityLayout() {
        let host = UIHostingController(rootView: MPProfileCard(name: "UI Test", email: "ui@example.com", onManage: {}))
        XCTAssertLessThanOrEqual(host.sizeThatFits(in: CGSize(width: 343, height: 1000)).height, 65)
    }

    func testCardUsesCompactSharedPadding() {
        let host = UIHostingController(rootView: MPSurfaceCard { Color.clear.frame(height: 1) })
        let size = host.sizeThatFits(in: CGSize(width: 375, height: 1000))
        XCTAssertEqual(size.height, 17, accuracy: 0.5, "One point of content plus 8pt padding on each edge")
    }

    func testCardAddsHorizontalBreathingRoomWithoutMakingRowsTaller() {
        let host = UIHostingController(rootView:
            MPSurfaceCard { Color.clear.frame(width: 100, height: 1) }
                .fixedSize(horizontal: true, vertical: true)
        )
        let size = host.sizeThatFits(in: CGSize(width: 375, height: 1000))
        XCTAssertEqual(size.width, 124, accuracy: 0.5, "100pt content plus12pt leading/trailing inset")
        XCTAssertEqual(size.height, 17, accuracy: 0.5, "Horizontal breathing room must not increase vertical density")
    }

    func testAssistantSummaryDoesNotAddAnotherSurfaceInsetAroundEveryMetric() {
        let host = UIHostingController(rootView: AssistantStructuredCard(
            type: "finance_summary",
            data: .object(["view": .object([
                "income": .number(1000), "expense": .number(150),
                "net": .number(850), "count": .number(3)
            ])])
        ))
        let reference = UIHostingController(rootView: MPAssistantCard(title: "Tổng quan tài chính") {
            LazyVGrid(columns: [GridItem(.flexible()), GridItem(.flexible())], spacing: 8) {
                MPMetric(title: "Thu nhập", value: "1.000 ₫")
                MPMetric(title: "Chi tiêu", value: "150 ₫")
                MPMetric(title: "Ròng", value: "850 ₫")
                MPMetric(title: "Giao dịch", value: "3")
            }
            MPText("Nguồn dữ liệu tài chính", style: .secondary)
        })
        let proposal = CGSize(width: 343, height: 1000)
        XCTAssertEqual(host.sizeThatFits(in: proposal).height, reference.sizeThatFits(in: proposal).height, accuracy: 1,
            "Summary must match a flat section, not add another inset around each metric.")
    }

    func testTransactionRowUsesCompactGeometryAndExpandsForAccessibility() {
        let row = MPTransactionRow(title: "Ăn uống", subtitle: "Cà phê", amount: "−45.000 ₫")
        let host = UIHostingController(rootView: row.environment(\.dynamicTypeSize, .large))
        let compact = host.sizeThatFits(in: CGSize(width: 343, height: 1000))
        XCTAssertLessThanOrEqual(compact.height, 56.5)
        let accessible = UIHostingController(rootView: row.environment(\.dynamicTypeSize, .accessibility3))
        XCTAssertGreaterThan(accessible.sizeThatFits(in: CGSize(width: 343, height: 1000)).height, compact.height)
    }

    func testAPIKeySecretClipboardCopiesExactSecretWhenRequested() {
        var clipboard: String?
        let copied = APIKeySecretClipboard.copy("mpk_lookup.secret-value") { clipboard = $0 }

        XCTAssertTrue(copied)
        XCTAssertEqual(clipboard, "mpk_lookup.secret-value")
    }

    func testAPIKeySecretClipboardDoesNotReplaceClipboardForEmptySecret() {
        var clipboard = "existing clipboard value"
        let copied = APIKeySecretClipboard.copy("   ") { clipboard = $0 }

        XCTAssertFalse(copied)
        XCTAssertEqual(clipboard, "existing clipboard value")
    }

    func testSemanticColorsMirrorWebThemeTokens() {
        assertRGB(MPColor.canvas, red: 251.0 / 255.0, green: 249.0 / 255.0, blue: 249.0 / 255.0)
        assertRGB(MPColor.surface, red: 1.0, green: 1.0, blue: 1.0)
        assertRGB(MPColor.brand, red: 76.0 / 255.0, green: 175.0 / 255.0, blue: 80.0 / 255.0)
        assertRGB(MPColor.action, red: 0.0, green: 110.0 / 255.0, blue: 28.0 / 255.0)
    }

    func testPrimaryNavigationMatchesWebShellAndExposesGlobalAddAction() {
        XCTAssertEqual(
            MPShellNavigation.primaryItems.map(\.id),
            ["overview", "transactions", "assistant", "account"]
        )
        XCTAssertEqual(
            MPShellNavigation.primaryItems.map(\.title),
            ["Tổng quan", "Giao dịch", "Trợ lý", "Tài khoản"]
        )
        XCTAssertEqual(MPShellNavigation.addActionAccessibilityID, "add-transaction-button")
    }

    func testBottomNavigationUsesReferenceSelectedPillAndSharedGeometry() {
        XCTAssertEqual(MPShellNavigation.selectedTabStyle, .surfacePill)
        XCTAssertEqual(MPShellNavigation.bottomNavigationCornerRadius, 28)
        XCTAssertEqual(MPShellNavigation.selectedTabCornerRadius, 22)
        XCTAssertEqual(MPShellNavigation.addButtonDiameter, 56)
        XCTAssertEqual(MPShellNavigation.feedbackBubbleDiameter, 52)
    }

    func testReferenceSegmentedControlUsesSelectedSurfacePill() {
        XCTAssertEqual(MPSegmentedControl.buttonStyle(isSelected: true), .segmentedSelected)
        XCTAssertEqual(MPSegmentedControl.buttonStyle(isSelected: false), .plain)
    }

    func testNativeCategoryPickerUsesStableAccessibleControls() {
        XCTAssertEqual(MPCategoryPickerCopy.searchAccessibilityID, "category-picker-search")
        XCTAssertEqual(MPCategoryPickerCopy.treeAccessibilityID, "category-picker-tree")
        XCTAssertEqual(MPCategoryPickerCopy.clearAccessibilityID, "category-picker-clear")
        XCTAssertEqual(MPCategoryPickerCopy.choiceAccessibilityID("food"), "category-choice-food")
        XCTAssertEqual(MPCategoryPickerCopy.clearButtonStyle, .ghost)
    }

    func testCategoryTreeSubtitleMatchesWalletScopeContract() {
        XCTAssertEqual(CategoriesReferenceCopy.scopeSubtitle(walletIDs: []), "Áp dụng tất cả ví")
        XCTAssertEqual(CategoriesReferenceCopy.scopeSubtitle(walletIDs: ["wallet-1"]), "Hoạt động trong 1 ví")
        XCTAssertEqual(CategoriesReferenceCopy.scopeSubtitle(walletIDs: ["wallet-1", "wallet-2"]), "Hoạt động trong 2 ví")
    }

    func testOverviewReferenceSectionsKeepStitchOrder() {
        XCTAssertEqual(
            OverviewReferenceSection.stitchOrder,
            [.wallets, .insider, .report, .monthSummary, .recentTransactions]
        )
    }

    func testOverviewRecentTransactionTitlePrefersCategoryAndFallsBackByType() {
        func title(_ type: String, _ name: String?) -> String {
            MPTransactionRow(type: type, rawAmount: 10_000, categoryName: name, occurredAt: "2026-10-13T00:00:00Z", timeZone: .gmt, masked: false).title
        }
        XCTAssertEqual(title("expense", "Ăn uống"), "Ăn uống")
        XCTAssertEqual(title("income", nil), "Khoản thu")
        XCTAssertEqual(title("expense", "  "), "Khoản chi")
    }

    func testOverviewWalletSubtitlePrefersDescriptionAndFallsBackToType() {
        XCTAssertEqual(OverviewViewModel.walletSubtitle(description: "  Chi tiêu hàng ngày ", type: "basic"), "Chi tiêu hàng ngày")
        XCTAssertEqual(OverviewViewModel.walletSubtitle(description: "", type: "goal"), "Tiết kiệm")
        XCTAssertEqual(OverviewViewModel.walletSubtitle(description: nil, type: "credit"), "Tín dụng")
    }

    func testOverviewWalletActionMatchesWebCopy() {
        XCTAssertEqual(OverviewReferenceCopy.walletAction, "Quản lý")
    }

    func testCreditLedgerEmptyStateMatchesWebCopy() {
        XCTAssertEqual(CreditLedgerCopy.emptyStatement, "Chưa có giao dịch tín dụng.")
    }

    func testCreditSummaryActionsMatchWebPanelContract() {
        XCTAssertEqual(CreditLedgerCopy.addEntryAction, "Thêm giao dịch")
        XCTAssertEqual(CreditLedgerCopy.addEntryActionAccessibilityID, "credit-add-entry")
        XCTAssertEqual(CreditLedgerCopy.paymentAction, "Thanh toán thẻ")
        XCTAssertEqual(CreditLedgerCopy.paymentActionAccessibilityID, "credit-add-payment")
        XCTAssertEqual(CreditLedgerReferenceSection.stitchOrder, [.header, .walletSelector, .summary, .actions, .statement])
    }

    func testWalletLedgerReferenceSectionsKeepStitchOrder() {
        XCTAssertEqual(
            WalletLedgerReferenceSection.stitchOrder,
            [.header, .balance, .summary, .allTransactions]
        )
    }

    func testWalletEditorReferenceSectionsKeepStitchOrder() {
        XCTAssertEqual(
            WalletEditorReferenceSection.stitchOrder,
            [.header, .identity, .type, .specialized, .settings, .save]
        )
    }

    func testReportsReferenceSectionsKeepStitchOrder() {
        XCTAssertEqual(
            ReportsReferenceSection.stitchOrder,
            [.header, .filter, .summary, .trend, .breakdown, .activity, .insider]
        )
    }

    func testReportsSummaryActionMatchesWebMonthCopy() {
        XCTAssertEqual(ReportsReferenceCopy.summaryAction(month: "2026-10"), "2026-10")
    }

    func testTrendDataUsesAccountTimezoneAndExcludesNonReportRows() {
        let timeZone = TimeZone(identifier: "Asia/Ho_Chi_Minh")!
        let rows = [
            TransactionModel(id: "expense-local-day", ownerID: "owner", walletID: "wallet", categoryID: nil, jarID: nil, jarName: nil, type: "expense", amount: 120_000, adjustmentDirection: nil, creditKind: nil, creditPaymentID: nil, travelEventID: nil, occurredAt: "2026-10-01T00:30:00+07:00", note: nil, includedInReports: true, transferID: nil, createdAt: "2026-10-01T00:30:00+07:00", updatedAt: "2026-10-01T00:30:00+07:00"),
            TransactionModel(id: "income-boundary", ownerID: "owner", walletID: "wallet", categoryID: nil, jarID: nil, jarName: nil, type: "income", amount: 500_000, adjustmentDirection: nil, creditKind: nil, creditPaymentID: nil, travelEventID: nil, occurredAt: "2026-09-30T17:30:00Z", note: nil, includedInReports: true, transferID: nil, createdAt: "2026-09-30T17:30:00Z", updatedAt: "2026-09-30T17:30:00Z"),
            TransactionModel(id: "excluded", ownerID: "owner", walletID: "wallet", categoryID: nil, jarID: nil, jarName: nil, type: "expense", amount: 900_000, adjustmentDirection: nil, creditKind: nil, creditPaymentID: nil, travelEventID: nil, occurredAt: "2026-10-01T03:00:00+07:00", note: nil, includedInReports: false, transferID: nil, createdAt: "2026-10-01T03:00:00+07:00", updatedAt: "2026-10-01T03:00:00+07:00"),
            TransactionModel(id: "transfer", ownerID: "owner", walletID: "wallet", categoryID: nil, jarID: nil, jarName: nil, type: "transfer", amount: 700_000, adjustmentDirection: nil, creditKind: nil, creditPaymentID: nil, travelEventID: nil, occurredAt: "2026-10-02T10:00:00+07:00", note: nil, includedInReports: true, transferID: "transfer-1", createdAt: "2026-10-02T10:00:00+07:00", updatedAt: "2026-10-02T10:00:00+07:00")
        ]

        let points = MPTrendData.points(transactions: rows, month: "2026-10", timeZone: timeZone)

        XCTAssertEqual(points.map(\.dayKey), ["2026-10-01"])
        XCTAssertEqual(points.first?.income, 500_000)
        XCTAssertEqual(points.first?.expense, 120_000)
        XCTAssertEqual(
            MPTrendData.accessibilityValue(points: points, masked: false),
            "Ngày 01: Thu \(MPFormat.money(500_000)), Chi \(MPFormat.money(120_000))"
        )
        let maskedValue = MPTrendData.accessibilityValue(points: points, masked: true)
        XCTAssertEqual(maskedValue, "Ngày 01: Thu \(MPFormat.maskedValue), Chi \(MPFormat.maskedValue)")
        XCTAssertFalse(maskedValue.contains("500.000"))
        XCTAssertFalse(maskedValue.contains("120.000"))
    }

    func testTrendChartUsesStableAccessibilityIdentifier() {
        XCTAssertEqual(MPTrendChart.accessibilityID, "reports-trend-chart")
    }

    func testReportsComparisonChartNormalizesIncomeAndExpenseSafely() {
        XCTAssertEqual(MPReportComparisonChart.normalizedHeights(income: 300, expense: 150), [100, 50])
        XCTAssertEqual(MPReportComparisonChart.normalizedHeights(income: 0, expense: 250), [0, 100])
        XCTAssertEqual(MPReportComparisonChart.normalizedHeights(income: 0, expense: 0), [0, 0])
        XCTAssertEqual(MPReportComparisonChart.normalizedHeights(income: -50, expense: 150), [0, 100])
        XCTAssertEqual(MPReportComparisonChart.accessibilityID, "reports-comparison-chart")
    }

    func testBudgetGaugeClampsAndExposesStableAccessibilityContract() {
        XCTAssertEqual(MPBudgetGauge.normalizedProgress(value: -10), 0)
        XCTAssertEqual(MPBudgetGauge.normalizedProgress(value: 42.5), 0.425, accuracy: 0.0001)
        XCTAssertEqual(MPBudgetGauge.normalizedProgress(value: 125), 1)
        XCTAssertEqual(MPBudgetGauge.accessibilityID, "budget-summary-gauge")
    }

    func testStatusBadgeUsesNamedSemanticContract() {
        XCTAssertEqual(MPStatusBadge.accessibilityID, "status-badge")
        XCTAssertEqual(MPStatusBadge.tone(for: true), .success)
        XCTAssertEqual(MPStatusBadge.tone(for: false), .muted)
    }

    func testBudgetsSummaryVisibilityMatchesWebEmptyBranch() {
        XCTAssertFalse(BudgetReferenceCopy.shouldShowSummary(itemCount: 0))
        XCTAssertTrue(BudgetReferenceCopy.shouldShowSummary(itemCount: 1))
    }

    func testBudgetsCreateActionMatchesWebCopy() {
        XCTAssertEqual(BudgetReferenceCopy.addAction, "Thêm")
        XCTAssertEqual(BudgetReferenceCopy.addActionAccessibilityID, "budget-add")
    }

    func testCategoriesCreateActionMatchesWebGroupsContract() {
        XCTAssertEqual(CategoriesReferenceCopy.addAction, "Nhóm mới")
        XCTAssertEqual(CategoriesReferenceCopy.addActionAccessibilityID, "category-add-group")
        XCTAssertEqual(CategoriesReferenceCopy.kindFilterAccessibilityID, "category-kind-filter")
        XCTAssertEqual(CategoriesReferenceCopy.walletFilterAccessibilityID, "category-wallet-filter")
        XCTAssertEqual(CategoriesReferenceCopy.searchAccessibilityID, "category-search")
    }

    func testJarsCreateActionMatchesWebMonthlySectionContract() {
        XCTAssertEqual(JarsReferenceCopy.addAction, "Thêm hũ")
        XCTAssertEqual(JarsReferenceCopy.addActionAccessibilityID, "jar-add-month")
    }

    func testRecurringCreateActionMatchesWebSummaryCardContract() {
        XCTAssertEqual(RecurringReferenceCopy.addAction, "Thêm")
        XCTAssertEqual(RecurringReferenceCopy.addActionAccessibilityID, "recurring-add")
    }

    func testTravelCreateActionMatchesWebSummaryCardContract() {
        XCTAssertEqual(TravelReferenceCopy.addAction, "Thêm")
        XCTAssertEqual(TravelReferenceCopy.addActionAccessibilityID, "travel-add")
    }

    func testPortfolioActionAnatomyMatchesWebPanelContract() {
        XCTAssertEqual(PortfolioReferenceCopy.subtitle, "Theo dõi mua bán độc lập với số dư ví")
        XCTAssertEqual(PortfolioReferenceCopy.summaryTitle, "Tổng danh mục")
        XCTAssertEqual(PortfolioReferenceCopy.refreshAction, "Làm mới")
        XCTAssertEqual(PortfolioReferenceCopy.refreshActionAccessibilityID, "portfolio-refresh")
        XCTAssertEqual(PortfolioReferenceCopy.addAssetAction, "Thêm tài sản")
        XCTAssertEqual(PortfolioReferenceCopy.addAssetActionAccessibilityID, "portfolio-add-asset")
        XCTAssertEqual(PortfolioReferenceCopy.saveTradeAction, "Lưu giao dịch")
        XCTAssertEqual(PortfolioReferenceCopy.saveTradeActionAccessibilityID, "portfolio-save-trade")
    }

    func testPortfolioCorrectionActionsExposeStableCopyAndAccessibilityIDs() {
        XCTAssertEqual(PortfolioCorrectionPresentation.replaceAction, "Sửa giao dịch")
        XCTAssertEqual(PortfolioCorrectionPresentation.voidAction, "Hủy giao dịch")
        XCTAssertEqual(PortfolioCorrectionPresentation.previewAction, "Xem thay đổi")
        XCTAssertEqual(PortfolioCorrectionPresentation.confirmAction, "Xác nhận sửa")
        XCTAssertEqual(PortfolioCorrectionPresentation.operationLabel("replace"), "Sửa giao dịch")
        XCTAssertEqual(PortfolioCorrectionPresentation.operationLabel("void"), "Hủy giao dịch")
        XCTAssertEqual(PortfolioCorrectionPresentation.operationLabel("unknown"), "Sửa giao dịch")
        XCTAssertEqual(PortfolioCorrectionPresentation.menuAccessibilityID(tradeID: "trade-1"), "portfolio-trade-menu-trade-1")
        XCTAssertEqual(PortfolioCorrectionPresentation.previewAccessibilityID, "portfolio-correction-preview")
        XCTAssertEqual(PortfolioCorrectionPresentation.confirmAccessibilityID, "portfolio-correction-confirm")
    }

    func testSemanticTokensAreStable() {
        XCTAssertEqual(MPMetrics.hitTarget, 44)
        XCTAssertEqual(MPMetrics.controlHeight, 44)
        XCTAssertEqual(MPMetrics.cornerRadius, 12)
        XCTAssertEqual(MPMetrics.buttonCornerRadius, 8)
        XCTAssertEqual(MPMetrics.pagePadding, 16)
        XCTAssertEqual(MPFont.uiBody.pointSize, 16, accuracy: 0.01)
    }

    func testWalletGlyphUsesWebSemanticToneContract() {
        XCTAssertEqual(MPWalletGlyph.iconName(for: "basic"), "wallet.bifold.fill")
        XCTAssertEqual(MPWalletGlyph.iconName(for: "goal"), "banknote.fill")
        XCTAssertEqual(MPWalletGlyph.iconName(for: "credit"), "creditcard.fill")
        assertRGB(MPColor.warning, red: 217.0 / 255.0, green: 119.0 / 255.0, blue: 6.0 / 255.0)
        assertRGB(MPColor.categoryTeal, red: 13.0 / 255.0, green: 148.0 / 255.0, blue: 136.0 / 255.0)
        assertRGB(MPColor.categoryTealSoft, red: 240.0 / 255.0, green: 253.0 / 255.0, blue: 250.0 / 255.0)
        XCTAssertEqual(MPMetrics.hitTarget, 44)
    }

    func testWalletScopeGlyphMatchesWebSelectorIdentityAndTone() {
        XCTAssertEqual(MPWalletScopeGlyph.iconName(for: nil), "globe")
        XCTAssertEqual(MPWalletScopeGlyph.iconName(for: "basic"), "wallet.bifold.fill")
        XCTAssertEqual(MPWalletScopeGlyph.iconName(for: "goal"), "banknote.fill")
        XCTAssertEqual(MPWalletScopeGlyph.iconName(for: "credit"), "creditcard.fill")
        XCTAssertEqual(MPWalletScopeGlyph.tone(for: nil), "categorySlate")
        XCTAssertEqual(MPWalletScopeGlyph.tone(for: "basic"), "categorySlate")
        XCTAssertEqual(MPWalletScopeGlyph.tone(for: "goal"), "success")
        XCTAssertEqual(MPWalletScopeGlyph.tone(for: "credit"), "categoryTeal")
    }

    func testAuthCopyMatchesWebMobileContract() {
        XCTAssertEqual(AuthReferenceCopy.title, "MyPocket")
        XCTAssertEqual(AuthReferenceCopy.subtitle, "Đăng nhập bằng email và mật khẩu để tiếp tục.")
        XCTAssertEqual(AuthReferenceCopy.emailPlaceholder, "you@example.com")
        XCTAssertEqual(AuthReferenceCopy.passwordPlaceholder, "Nhập mật khẩu")
    }

    func testTransactionFormBasesKeepStitchContracts() {
        XCTAssertEqual(MPFormRow.minimumHitTarget, MPMetrics.hitTarget)
        var components = DateComponents()
        components.calendar = Calendar(identifier: .gregorian)
        components.timeZone = TimeZone(identifier: "Asia/Ho_Chi_Minh")
        components.year = 2026
        components.month = 10
        components.day = 5
        let date = components.date ?? .distantPast
        XCTAssertEqual(
            MPDateStepper.displayLabel(for: date, timeZone: "Asia/Ho_Chi_Minh"),
            "Thứ Hai, 05/10/2026"
        )
    }

    func testCalendarGridMatchesStitchMondayFirstDateOnlyContract() {
        let timeZone = TimeZone(identifier: "Asia/Ho_Chi_Minh")!
        var components = DateComponents()
        components.calendar = Calendar(identifier: .gregorian)
        components.timeZone = timeZone
        components.year = 2026
        components.month = 10
        components.day = 1
        let month = components.date!

        XCTAssertEqual(MPCalendarGrid.weekdayLabels, ["Thứ 2", "Thứ 3", "Thứ 4", "Thứ 5", "Thứ 6", "Thứ 7", "CN"])
        XCTAssertEqual(MPCalendarGrid.monthTitle(for: month, timeZone: timeZone), "thg 10 2026")

        let days = MPCalendarGrid.days(for: month, timeZone: timeZone)
        XCTAssertEqual(days.count, 42)
        XCTAssertEqual(days.first?.key, "2026-09-28")
        XCTAssertEqual(days.last?.key, "2026-11-08")
        XCTAssertEqual(days.filter(\.isInDisplayedMonth).count, 31)
        XCTAssertEqual(MPCalendarGrid.dateKey(days[3].date, timeZone: timeZone), "2026-10-01")
    }

    func testDateTimePickerPreservesClockWhenSelectingAnotherAccountDay() {
        let timeZone = TimeZone(identifier: "Asia/Ho_Chi_Minh")!
        let source = MPFormat.parseInstant("2026-10-05T09:45:12+07:00")!
        let targetDay = MPFormat.parseCalendarDate("2026-10-08", timeZone: timeZone)!
        let selected = MPCalendarGrid.datePreservingClock(source, on: targetDay, timeZone: timeZone)
        XCTAssertEqual(MPCalendarGrid.dateKey(selected, timeZone: timeZone), "2026-10-08")
        let calendar = MPCalendarGrid.calendar(timeZone: timeZone)
        XCTAssertEqual(calendar.component(.hour, from: selected), 9)
        XCTAssertEqual(calendar.component(.minute, from: selected), 45)
        XCTAssertEqual(calendar.component(.second, from: selected), 12)
    }

    func testDateTimeFieldUsesAccountTimezoneDisplayContract() {
        let date = MPFormat.parseInstant("2026-10-07T09:45:00+07:00")!
        let label = MPDateTimeField.displayLabel(for: date, timeZone: "Asia/Ho_Chi_Minh")
        XCTAssertTrue(label.contains("07/10/2026 09:45"), label)
    }

    func testAssistantCardAliasesMatchWebVocabulary() {
        XCTAssertEqual(AssistantCardType.normalized("finance_summary"), "metric_group")
        XCTAssertEqual(AssistantCardType.normalized("spending_comparison"), "period_comparison")
        XCTAssertEqual(AssistantCardType.normalized("transaction_search"), "transaction_list")
        XCTAssertEqual(AssistantCardType.normalized("budget_progress"), "budget_status")
        XCTAssertEqual(AssistantCardType.normalized("wallet_creation_proposal"), "wallet_confirmation")
        XCTAssertEqual(AssistantCardType.normalized("jar_creation_proposal"), "jar_confirmation")
        XCTAssertEqual(AssistantCardType.normalized("product_help"), "product_help")
        XCTAssertEqual(AssistantCardType.normalized("transaction_detail"), "transaction_detail")
        XCTAssertEqual(AssistantCardType.title(for: "transaction_detail"), "Chi tiết giao dịch")
        XCTAssertEqual(AssistantCardType.title(for: "product_help"), "Hướng dẫn trong ứng dụng")
    }

    func testAssistantCardsUseTheSameBoundedItemPrefixesAsWeb() {
        let items = (0..<10).map { JSONValue.string("row-\($0)") }

        XCTAssertEqual(AssistantCardType.visibleItems(items, for: "transaction_list").count, 5)
        XCTAssertEqual(AssistantCardType.visibleItems(items, for: "transaction_search").count, 5)
        for type in ["wallet_balances", "goal_progress", "budget_status", "budget_progress", "jar_progress"] {
            XCTAssertEqual(AssistantCardType.visibleItems(items, for: type).count, 8, type)
        }
        XCTAssertEqual(
            AssistantCardType.visibleItems(items, for: "transaction_search").map(\.prettyDescription),
            Array(items.prefix(5)).map(\.prettyDescription)
        )
        XCTAssertEqual(AssistantCardType.transactionCount(["total_count": .number(42), "items": .array(items)]), 42)
        XCTAssertEqual(AssistantCardType.visibleItems(items, for: "unknown_card").count, 10)
    }

    func testAssistantWalletConfirmationPreventsDuplicateAndLocksAfterSuccess() {
        var confirmation: AssistantWalletConfirmationState = .idle

        XCTAssertTrue(confirmation.begin())
        XCTAssertEqual(confirmation, .saving)
        XCTAssertFalse(confirmation.canConfirm)
        XCTAssertFalse(confirmation.begin(), "A second tap during the request must not submit another wallet create.")

        confirmation.finish(success: true)
        XCTAssertEqual(confirmation, .succeeded)
        XCTAssertFalse(confirmation.canConfirm, "A completed confirmation card must stay disabled.")

        var retryable: AssistantWalletConfirmationState = .idle
        XCTAssertTrue(retryable.begin())
        retryable.finish(success: false)
        XCTAssertEqual(retryable, .failed)
        XCTAssertTrue(retryable.canConfirm, "A failed request must remain retryable.")
    }

    func testAssistantMalformedWalletProposalRendersInvalidStatusInsteadOfConfirmation() {
        let reference = UIHostingController(rootView: MPAssistantCard(title: "Đề xuất tạo ví") {
            MPStatusMessage(text: "Đề xuất tạo ví chưa đủ dữ liệu để xác nhận.", tone: .danger)
        })
        let size = CGSize(width: 343, height: 1000)
        let expected = reference.sizeThatFits(in: size).height
        let malformed: [[String: JSONValue]] = [
            ["name": .string("Ví đề xuất"), "type": .string("unsupported")],
            ["name": .string("Ví đề xuất"), "type": .string("basic"), "opening_balance": .number(1.5)],
            ["name": .string("Ví đề xuất"), "type": .string("goal")],
            ["name": .string("Ví đề xuất"), "type": .string("credit")]
        ]
        for view in malformed {
            let host = UIHostingController(rootView: AssistantStructuredCard(
                type: "wallet_creation_proposal", data: .object(["view": .object(view)])))
            XCTAssertEqual(host.sizeThatFits(in: size).height, expected, accuracy: 1,
                "Malformed proposal must render the shared invalid status, not a fabricated wallet confirmation: \(view)")
        }
    }

    func testAssistantWalletProposalPreservesValidInputsAndOnlyDocumentedDefaults() throws {
        let basic = try XCTUnwrap(AssistantWalletProposal.decode(["name": .string("  Ví chính  "), "type": .string("basic")]))
        XCTAssertEqual(basic.name, "Ví chính")
        XCTAssertEqual(basic.type, "basic")
        XCTAssertEqual(basic.openingBalance, 0)
        XCTAssertTrue(basic.isInTotal)
        let goal = try XCTUnwrap(AssistantWalletProposal.decode([
            "name": .string("Du lịch"), "type": .string("goal"), "action": .string("create_wallet"),
            "opening_balance": .number(12500), "is_in_total": .bool(false),
            "target_amount": .number(1000000), "target_date": .string("2028-02-29"),
            "description": .string(" Kỳ nghỉ "), "credit_limit": .null
        ]))
        XCTAssertEqual(goal.openingBalance, 12500)
        XCTAssertFalse(goal.isInTotal)
        XCTAssertEqual(goal.targetAmount, 1000000)
        XCTAssertEqual(goal.targetDate, "2028-02-29")
        XCTAssertEqual(goal.description, "Kỳ nghỉ")
        XCTAssertNil(goal.creditLimit)
        let credit = try XCTUnwrap(AssistantWalletProposal.decode([
            "name": .string("Thẻ"), "type": .string("credit"), "credit_limit": .number(5000000),
            "last_statement_balance": .number(200000), "statement_day": .number(31),
            "payment_due_day": .number(1), "description": .null, "target_date": .null
        ]))
        XCTAssertEqual(credit.creditLimit, 5000000)
        XCTAssertEqual(credit.lastStatementBalance, 200000)
        XCTAssertEqual(credit.statementDay, 31)
        XCTAssertEqual(credit.paymentDueDay, 1)
        XCTAssertNil(credit.targetAmount)
    }

    func testAssistantWalletProposalUsesPersistedActionIDForConfirmationReplay() throws {
        let view: [String: JSONValue] = [
            "action_id": .string("wallet-action-123"),
            "name": .string("Ví chính"),
            "type": .string("basic")
        ]
        XCTAssertNotNil(try XCTUnwrap(AssistantWalletProposal.decode(view)))
        XCTAssertEqual(AssistantWalletProposal.actionID(view), "wallet-action-123")
        XCTAssertNil(AssistantWalletProposal.actionID(["action_id": .string(" ")]))
    }

    func testAssistantJarProposalDecodesOnlyValidatedAllocationFields() throws {
        let fixed = try XCTUnwrap(AssistantJarProposal.decode([
            "action": .string("create_jar"), "action_id": .string("jar-action-1"),
            "month": .string("2026-09"), "name": .string(" Ăn uống "),
            "allocation_mode": .string("fixed"), "allocation_amount": .number(400000)
        ]))
        XCTAssertEqual(fixed.month, "2026-09")
        XCTAssertEqual(fixed.name, "Ăn uống")
        XCTAssertEqual(fixed.amount, 400000)
        let update = try XCTUnwrap(AssistantJarProposal.decode([
            "action": .string("update_jar"), "jar_id": .string("11111111-1111-4111-8111-111111111111"),
            "month": .string("2026-09"), "name": .string("Ăn uống mới"),
            "allocation_mode": .string("percent"), "allocation_percent": .number(20)
        ]))
        XCTAssertEqual(update.action, "update_jar")
        XCTAssertEqual(update.jarID, "11111111-1111-4111-8111-111111111111")
        XCTAssertEqual(update.percent, 20)
        XCTAssertEqual(AssistantCardType.title(for: "jar_update_proposal"), "Đề xuất cập nhật hũ")
        XCTAssertNil(AssistantJarProposal.decode([
            "action": .string("update_jar"), "month": .string("2026-09"), "name": .string("Hũ"),
            "allocation_mode": .string("none")
        ]))

        let budget = try XCTUnwrap(AssistantBudgetProposal.decode([
            "action": .string("create_budget"), "action_id": .string("budget-action-1"),
            "name": .string(" Ăn uống "), "limit_amount": .number(4_000_000),
            "start_date": .string("2026-09-01"), "end_date": .string("2026-09-30")
        ]))
        XCTAssertEqual(budget.name, "Ăn uống")
        XCTAssertEqual(budget.limitAmount, 4_000_000)
        XCTAssertEqual(AssistantCardType.title(for: "budget_creation_proposal"), "Đề xuất tạo ngân sách")
        let budgetUpdate = try XCTUnwrap(AssistantBudgetProposal.decode([
            "action": .string("update_budget"), "budget_id": .string("11111111-1111-4111-8111-111111111111"),
            "name": .string(" Ăn uống mới "), "limit_amount": .number(5_000_000),
            "start_date": .string("2026-09-01"), "end_date": .string("2026-09-30")
        ]))
        XCTAssertEqual(budgetUpdate.action, "update_budget")
        XCTAssertEqual(budgetUpdate.budgetID, "11111111-1111-4111-8111-111111111111")
        XCTAssertEqual(budgetUpdate.name, "Ăn uống mới")
        XCTAssertEqual(AssistantCardType.title(for: "budget_update_proposal"), "Đề xuất cập nhật ngân sách")
        XCTAssertNil(AssistantBudgetProposal.decode([
            "action": .string("create_budget"), "name": .string("Hũ"), "limit_amount": .number(1),
            "start_date": .string("2026-09-30"), "end_date": .string("2026-09-01")
        ]))
        XCTAssertNil(AssistantBudgetProposal.decode([
            "action": .string("update_budget"), "name": .string("Hũ"), "limit_amount": .number(1),
            "start_date": .string("2026-09-01"), "end_date": .string("2026-09-30")
        ]))
        XCTAssertEqual(AssistantJarProposal.actionID(["action_id": .string("jar-action-1")]), "jar-action-1")
        XCTAssertNil(AssistantJarProposal.decode(["month": .string("2026-09"), "name": .string("Hũ"), "allocation_mode": .string("fixed")]))
        XCTAssertNil(AssistantJarProposal.decode(["month": .string("2026-09"), "name": .string("Hũ"), "allocation_mode": .string("none"), "allocation_amount": .number(1)]))
    }

    func testAssistantWalletProposalRequiresBoundConversationGenerationForDispatcher() {
        let view: [String: JSONValue] = [
            "conversation_id": .string("conversation-1"),
            "generation": .number(4)
        ]
        XCTAssertEqual(AssistantWalletProposal.actionContext(view)?.conversationID, "conversation-1")
        XCTAssertEqual(AssistantWalletProposal.actionContext(view)?.generation, 4)
        XCTAssertNil(AssistantWalletProposal.actionContext(["conversation_id": .string("conversation-1"), "generation": .number(0)]))
        XCTAssertNil(AssistantWalletProposal.actionContext(["conversation_id": .string(" "), "generation": .number(4)]))
    }

    func testAssistantWalletConfirmationSummaryShowsGoalCreditAndTotalFields() throws {
        let goal = try XCTUnwrap(AssistantWalletProposal.decode([
            "name": .string("Du lịch"), "type": .string("goal"), "opening_balance": .number(12500),
            "is_in_total": .bool(false), "target_amount": .number(1000000), "target_date": .string("2028-02-29"),
            "description": .string("Kỳ nghỉ")
        ]))
        let goalLines = AssistantWalletConfirmationSummary.lines(for: goal, masked: false)
        XCTAssertTrue(goalLines.contains { $0.contains("Không tính vào tổng") })
        XCTAssertTrue(goalLines.contains { $0.contains("Mục tiêu") && $0.contains("1") })
        XCTAssertTrue(goalLines.contains { $0.contains("Hạn mục tiêu: 2028-02-29") })
        XCTAssertTrue(goalLines.contains { $0.contains("Kỳ nghỉ") })

        let credit = try XCTUnwrap(AssistantWalletProposal.decode([
            "name": .string("Thẻ"), "type": .string("credit"), "credit_limit": .number(5000000),
            "last_statement_balance": .number(200000), "statement_day": .number(31), "payment_due_day": .number(1)
        ]))
        let creditLines = AssistantWalletConfirmationSummary.lines(for: credit, masked: false)
        XCTAssertTrue(creditLines.contains { $0.contains("Hạn mức") && $0.contains("5") })
        XCTAssertTrue(creditLines.contains { $0.contains("Sao kê ngày 31") })
        XCTAssertTrue(creditLines.contains { $0.contains("Đến hạn ngày 1") })
        XCTAssertTrue(creditLines.contains { $0.contains("Dư nợ sao kê") && $0.contains("200") })
    }

    func testAssistantWalletProposalRejectsMalformedAndUnsafeValuesWithoutCoercion() {
        let base: [String: JSONValue] = ["name": .string("Ví"), "type": .string("basic")]
        for key in ["opening_balance", "target_amount", "credit_limit", "last_statement_balance", "statement_day", "payment_due_day"] {
            for value in [JSONValue.number(1.5), .number(-1), .number(.infinity), .number(.nan), .number(Double(Int64.max)), .number(9_007_199_254_740_992), .string("12500")] {
                var proposal = base; proposal[key] = value
                XCTAssertNil(AssistantWalletProposal.decode(proposal), key)
            }
        }
        for (key, value) in [("name", JSONValue.string(" ")), ("name", .string(String(repeating: "a", count: 81))),
                             ("type", .string("other")), ("type", .null), ("action", .string("delete_wallet")),
                             ("is_in_total", .string("false")), ("description", .number(12)),
                             ("description", .string(String(repeating: "a", count: 513)))] {
            var proposal = base; proposal[key] = value
            XCTAssertNil(AssistantWalletProposal.decode(proposal), key)
        }
        for date in ["2027-02-29", "2026-13-01", "2026-1-1", "tomorrow"] {
            XCTAssertNil(AssistantWalletProposal.decode(["name": .string("Mục tiêu"), "type": .string("goal"),
                                                        "target_amount": .number(1000), "target_date": .string(date)]))
        }
        for day in [0.0, 32.0] {
            XCTAssertNil(AssistantWalletProposal.decode(["name": .string("Thẻ"), "type": .string("credit"),
                                                        "credit_limit": .number(1000), "statement_day": .number(day)]))
        }
        XCTAssertNil(AssistantWalletProposal.decode(["name": .string("Ví")]))
        XCTAssertNil(AssistantWalletProposal.decode(["name": .string("Ví"), "type": .string("goal"), "target_amount": .number(0)]))
    }

    func testAssistantTransactionDetailSummaryHasSafeFallbacks() {
        let data: [String: JSONValue] = [
            "wallet_name": .string("Ví chính"),
            "category_name": .string("Ăn uống"),
            "included_in_reports": .bool(false),
            "transfer_id": .string("transfer-1")
        ]
        XCTAssertEqual(AssistantCardType.transactionDetailSummary(data), "Ví chính · Ăn uống")
        XCTAssertEqual(AssistantCardType.transactionDetailFlags(data), "Không tính vào báo cáo · Chuyển ví")
        XCTAssertEqual(AssistantCardType.transactionDetailSummary([:]), "Ví")
        XCTAssertEqual(AssistantCardType.transactionDetailFlags([:]), "Tính vào báo cáo")
    }

    func testAssistantComposerRequiresEnabledCapabilityAndNonEmptyText() {
        XCTAssertTrue(AssistantComposerPolicy.canSend(enabled: true, text: "  tháng này  ", activeRunID: nil))
        XCTAssertFalse(AssistantComposerPolicy.canSend(enabled: false, text: "tháng này", activeRunID: nil))
        XCTAssertFalse(AssistantComposerPolicy.canSend(enabled: true, text: "   ", activeRunID: nil))
        XCTAssertFalse(AssistantComposerPolicy.canSend(enabled: true, text: "tháng này", activeRunID: "run-1"))
        XCTAssertFalse(AssistantComposerPolicy.canSend(enabled: true, text: "tháng này", activeRunID: nil, isLocallySending: true))
    }

func testAssistantPrivacyMaskRedactsMoneyAndProse() {
        XCTAssertEqual(MPFormat.money(123_000, masked: true), MPFormat.maskedValue)
        XCTAssertEqual(MPFormat.signedMoney(-123_000, masked: true), MPFormat.maskedValue)
        XCTAssertEqual(
            AssistantPrivacyPolicy.text("Bạn đã chi 123.000 ₫", masked: true),
            "Nội dung ẩn khi bật chế độ che số dư."
        )
        XCTAssertEqual(AssistantPrivacyPolicy.text("Tóm tắt an toàn", masked: false), "Tóm tắt an toàn")
    }

    func testAssistantGoalProgressClampsSafely() {
        XCTAssertEqual(AssistantCardType.goalProgressPercent(current: 300, target: 1_000), 30)
        XCTAssertEqual(AssistantCardType.goalProgressPercent(current: 1_100, target: 1_000), 100)
        XCTAssertEqual(AssistantCardType.goalProgressPercent(current: -10, target: 1_000), 0)
        XCTAssertNil(AssistantCardType.goalProgressPercent(current: 100, target: 0))
        XCTAssertNil(AssistantCardType.goalProgressPercent(current: nil, target: 1_000))
    }

    func testAssistantProgressTargetPrefersJarAllocationAndRejectsInvalidTargets() {
        XCTAssertEqual(AssistantCardType.progressTarget(["limit_amount": .number(4_000_000), "effective_limit_amount": .number(5_000_000)]), 5_000_000)
        XCTAssertEqual(
            AssistantCardType.progressTarget([
                "allocation_amount": .number(2_000),
                "limit_amount": .number(9_000),
                "calculated_allocation": .number(8_000)
            ]),
            2_000
        )
        XCTAssertEqual(
            AssistantCardType.progressTarget([
                "limit_amount": .number(9_000),
                "calculated_allocation": .number(8_000)
            ]),
            9_000
        )
        XCTAssertEqual(
            AssistantCardType.progressTarget([
                "calculated_allocation": .number(8_000)
            ]),
            8_000
        )
        XCTAssertEqual(AssistantCardType.progressTarget(["allocation_amount": .number(0)]), 0)
        XCTAssertNil(AssistantCardType.progressTarget(["allocation_amount": .number(-1)]))
        XCTAssertNil(AssistantCardType.progressTarget([:]))
    }

    func testAssistantGoalProgressDoesNotFabricateCompletionFromUnsafeMoney() {
        XCTAssertNil(AssistantCardType.goalProgressPercent(current: 1e100, target: 1000))
        XCTAssertNil(AssistantCardType.goalProgressPercent(current: 1000, target: 1e100))
        XCTAssertNil(AssistantCardType.goalProgressPercent(current: 1.5, target: 1000))
        XCTAssertNil(AssistantCardType.goalProgressPercent(current: 1000, target: 1.5))
    }

    func testAssistantProgressTargetRejectsAmountsOutsideFinanceSafeIntegerRange() {
        let safeMaximum = 9_007_199_254_740_991.0
        XCTAssertEqual(AssistantCardType.progressTarget(["allocation_amount": .number(0)]), 0)
        XCTAssertEqual(AssistantCardType.progressTarget(["allocation_amount": .number(safeMaximum)]), safeMaximum)
        for amount in [safeMaximum + 1, Double(Int64.max), 1e100, .infinity, .nan] {
            XCTAssertNil(
                AssistantCardType.progressTarget(["allocation_amount": .number(amount)]),
                "An unsafe amount must never reach the progress renderer's Int64 conversion: \(amount)"
            )
        }
    }

    func testAssistantCardMetadataUsesSafeSourceAndCountFallbacks() {
        XCTAssertEqual(
            AssistantCardType.sourceLabel(.object(["timezone": .string("Asia/Ho_Chi_Minh")])),
            "Nguồn dữ liệu · Asia/Ho_Chi_Minh"
        )
        XCTAssertEqual(
            AssistantCardType.sourceLabel(.object(["timezone": .string("  ")])),
            "Nguồn dữ liệu tài chính"
        )
        XCTAssertEqual(AssistantCardType.transactionCount(["total_count": .number(42)]), 42)
        XCTAssertEqual(AssistantCardType.transactionCount(["items": .array([.null, .null])]), 0)
        XCTAssertEqual(
            AssistantCardType.transactionCount(["items": .array((0..<7).map { .object(["id": .string("\($0)")]) })]),
            5
        )
        XCTAssertEqual(AssistantCardType.transactionCount(["total_count": .number(-1)]), 0)
    }

    func testAssistantMoneyNeverInventsZeroOrTruncatesInvalidAmounts() {
        for value: JSONValue? in [nil, .null, .string("100"), .number(1.5), .number(1e100), .number(.infinity), .number(.nan), .number(9_007_199_254_740_992)] {
            XCTAssertEqual(AssistantCardType.money(value, masked: false), "Chưa có dữ liệu")
            XCTAssertEqual(AssistantCardType.money(value, masked: true), MPFormat.maskedValue)
        }
        XCTAssertEqual(AssistantCardType.monetaryValue(.number(-9_007_199_254_740_991)), -9_007_199_254_740_991)
        XCTAssertEqual(AssistantCardType.monetaryValue(.number(9_007_199_254_740_991)), 9_007_199_254_740_991)
        XCTAssertEqual(AssistantCardType.money(.number(0), masked: false), "0\u{00A0}₫")
        XCTAssertEqual(AssistantCardType.money(.number(-12500), masked: false), "-12.500\u{00A0}₫")
        XCTAssertEqual(AssistantCardType.money(.number(12500), masked: false), "12.500\u{00A0}₫")
    }

    func testAssistantCountOverflowFallsBackToVisibleTransactions() {
        let items: JSONValue = .array([.object(["id": .string("one")]), .null, .object(["id": .string("two")])])
        for amount in [Double(Int.max), 1e100, .infinity, .nan, 1.5, -1] {
            XCTAssertEqual(AssistantCardType.transactionCount(["total_count": .number(amount), "items": items]), 2)
        }
    }

    func testJSONFallbackFormatsUnrepresentableNumbersWithoutIntegerConversionTrap() throws {
        let decoded = try JSONDecoder().decode(JSONValue.self, from: Data("1e100".utf8))
        XCTAssertEqual(decoded.prettyDescription, "1e+100")
        XCTAssertEqual(JSONValue.number(Double(Int.max)).prettyDescription, "9.223372036854776e+18")
        XCTAssertEqual(JSONValue.number(12500).prettyDescription, "12500")
        XCTAssertEqual(JSONValue.number(1.5).prettyDescription, "1.5")
    }

    func testAssistantMalformedCardsRenderMaskedFallbackInsteadOfPrivatePayload() {
        let secret = String(repeating: "Giao dịch riêng tư 123456789 ", count: 30)
        let data: JSONValue = .object(["view": .object(["items": .string(secret)])])
        let size = CGSize(width: 343, height: 10000)
        for type in ["unrecognized_card", "transaction_list", "wallet_balances", "budget_status", "jar_progress"] {
            let reference = UIHostingController(rootView: MPAssistantCard(title: AssistantCardType.title(for: type)) {
                MPText("Nội dung ẩn khi bật chế độ che số dư.", style: .secondary)
                if type == "transaction_list" {
                    MPText("Nguồn dữ liệu tài chính", style: .secondary)
                }
            })
            // Transaction lists also render a count above the fallback; inspect
            // their fallback text directly rather than equating unequal layouts.
            if type == "transaction_list" {
                XCTAssertEqual(AssistantCardType.fallbackText(data, masked: true), "Nội dung ẩn khi bật chế độ che số dư.")
                continue
            }
            let host = UIHostingController(rootView: AssistantStructuredCard(type: type, data: data, masked: true))
            XCTAssertEqual(host.sizeThatFits(in: size).height, reference.sizeThatFits(in: size).height, accuracy: 1,
                           "Masked fallback must not lay out the long private payload for \(type)")
        }
        XCTAssertEqual(AssistantCardType.fallbackText(.string(secret), masked: false), secret)
    }

    private func assertRGB(_ color: Color, red expectedRed: CGFloat, green expectedGreen: CGFloat, blue expectedBlue: CGFloat) {
        var red: CGFloat = 0
        var green: CGFloat = 0
        var blue: CGFloat = 0
        var alpha: CGFloat = 0
        XCTAssertTrue(UIColor(color).getRed(&red, green: &green, blue: &blue, alpha: &alpha))
        XCTAssertEqual(red, expectedRed, accuracy: 0.002)
        XCTAssertEqual(green, expectedGreen, accuracy: 0.002)
        XCTAssertEqual(blue, expectedBlue, accuracy: 0.002)
    }

    func testAttachmentLimitIsExplicitInThePickerContract() {
        let picker = MPAttachmentPicker(title: "Đính kèm", count: 20, maxCount: 20, add: {})
        XCTAssertNotNil(picker)
    }

    func testReportFormattingHandlesDeltasAndMissingIncome() {
        XCTAssertEqual(MPFormat.percentageFromBPS(1234), "12.3%")
        XCTAssertEqual(MPFormat.percentageFromBPS(nil), "Chưa có dữ liệu")
        XCTAssertTrue(MPFormat.signedMoney(-1).hasPrefix("−"))
    }

    func testNumericDraftParserRejectsMalformedValuesWithoutFiltering() {
        XCTAssertEqual(MPFormat.parseNonNegativeInteger(" 1000 "), 1000)
        XCTAssertEqual(MPFormat.parseNonNegativeInteger(""), nil)
        XCTAssertEqual(MPFormat.parseNonNegativeInteger("", allowingEmpty: true), 0)
        XCTAssertEqual(MPFormat.parsePositiveInteger("0"), nil)
        XCTAssertEqual(MPFormat.parsePositiveInteger("1000"), 1000)
        XCTAssertNil(MPFormat.parseNonNegativeInteger("1_000"))
        XCTAssertNil(MPFormat.parseNonNegativeInteger("12abc3"))
        XCTAssertNil(MPFormat.parseNonNegativeInteger("١٢٣"))
        XCTAssertNil(MPFormat.parseNonNegativeInteger("9223372036854775808"))
    }

    func testSelectionLabelsPreserveUnknownValuesAndLocalizeEnums() {
        XCTAssertEqual(MPSelectionLabels.display(for: "expense"), "Chi")
        XCTAssertEqual(MPSelectionLabels.display(for: "percent"), "Phần trăm thu nhập")
        XCTAssertEqual(MPSelectionLabels.display(for: "unknown-provider-value"), "unknown-provider-value")
    }

    func testJarAllocationRejectsZeroPercentAndStaleModeValues() {
        XCTAssertTrue(JarAllocationDraftValidator.isValid(mode: "none", amount: nil, percent: nil))
        XCTAssertTrue(JarAllocationDraftValidator.isValid(mode: "fixed", amount: 1, percent: nil))
        XCTAssertTrue(JarAllocationDraftValidator.isValid(mode: "percent", amount: nil, percent: 100))
        XCTAssertFalse(JarAllocationDraftValidator.isValid(mode: "percent", amount: nil, percent: 0))
        XCTAssertTrue(JarAllocationDraftValidator.isValid(mode: "percent", amount: nil, percent: 12.34))
        XCTAssertFalse(JarAllocationDraftValidator.isValid(mode: "percent", amount: nil, percent: 12.345))
        XCTAssertFalse(JarAllocationDraftValidator.isValid(mode: "percent", amount: nil, percent: .infinity))
        XCTAssertFalse(JarAllocationDraftValidator.isValid(mode: "fixed", amount: 100, percent: 10))
        XCTAssertTrue(JarAllocationDraftValidator.isValid(mode: "fixed", amount: JarAllocationDraftValidator.maximumFixedAmount, percent: nil))
        XCTAssertFalse(JarAllocationDraftValidator.isValid(mode: "fixed", amount: JarAllocationDraftValidator.maximumFixedAmount + 1, percent: nil))
        XCTAssertFalse(JarAllocationDraftValidator.isValid(mode: "none", amount: 100, percent: nil))
    }

    func testPrivacyMaskSecondaryDisplaysUseSharedMoneyFormatter() {
        let values: [Int64] = [0, 1, 123_000, -45_000]
        for value in values {
            XCTAssertEqual(MPFormat.money(value, masked: true), MPFormat.maskedValue)
            XCTAssertEqual(MPFormat.signedMoney(value, masked: true), MPFormat.maskedValue)
        }
        XCTAssertNotEqual(MPFormat.money(123_000, masked: false), MPFormat.maskedValue)
    }

    func testWalletManagementBalanceLabelsRespectGlobalPrivacyMask() {
        XCTAssertEqual(MPWalletSelectionList.balanceLabel(100_000, masked: true), MPFormat.maskedValue)
        XCTAssertEqual(MPWalletSelectionList.balanceLabel(-75_000, masked: true), MPFormat.maskedValue)
        XCTAssertEqual(MPWalletSelectionList.balanceLabel(100_000, masked: false), MPFormat.money(100_000))
    }

    func testSegmentedControlTransitionHonorsReducedMotion() {
        XCTAssertEqual(MPSegmentedControl.transitionDuration(reduceMotion: false), 0.18, accuracy: 0.001)
        XCTAssertEqual(MPSegmentedControl.transitionDuration(reduceMotion: true), 0)
    }

    func testAssistantCardUsesSharedElevatedSurface() {
        XCTAssertEqual(MPAssistantCard<Text>.surfaceStyle, .elevated)
    }

    func testReportsInsiderCopyUsesSafeFallbacksAndMasking() {
        XCTAssertEqual(
            MPReportsInsiderCopy.change(expenseDelta: 10, changeBPS: 1250),
            "Chi tiêu tăng 12.5% so với kỳ trước."
        )
        XCTAssertEqual(
            MPReportsInsiderCopy.change(expenseDelta: -10, changeBPS: -850),
            "Chi tiêu giảm -8.5% so với kỳ trước."
        )
        XCTAssertEqual(
            MPReportsInsiderCopy.change(expenseDelta: 0, changeBPS: nil),
            "Chi tiêu chưa có kỳ gốc để so sánh."
        )
        XCTAssertEqual(MPReportsInsiderCopy.ratio(spendingIncomeRatioBPS: nil), "Chưa có thu nhập để tính tỷ lệ.")
        XCTAssertEqual(MPReportsInsiderCopy.ratio(spendingIncomeRatioBPS: 3250), "Chi/thu 32.5%")
        XCTAssertEqual(MPReportsInsiderCopy.averageDailyExpense(100_000, days: 7, masked: true), "Trung bình \(MPFormat.maskedValue)/ngày trong 7 ngày.")
        XCTAssertEqual(MPReportsInsiderCopy.averageDailyExpense(100_000, days: 7, masked: false), "Trung bình 100.000 ₫/ngày trong 7 ngày.")
    }

    func testMoneyInsiderRingKeepsActualPercentageAndSafeMissingState() {
        XCTAssertEqual(MPMoneyInsiderRing.percentage(changeBPS: 21_400), 214)
        XCTAssertEqual(MPMoneyInsiderRing.percentage(changeBPS: -850), -8.5)
        XCTAssertEqual(MPMoneyInsiderRing.percentage(changeBPS: 0), 0)
        XCTAssertNil(MPMoneyInsiderRing.percentage(changeBPS: nil))
        XCTAssertEqual(MPMoneyInsiderRing.label(changeBPS: 21_400), "Chi tiêu cao hơn kỳ trước")
        XCTAssertEqual(MPMoneyInsiderRing.label(changeBPS: -850), "Chi tiêu thấp hơn kỳ trước")
        XCTAssertEqual(MPMoneyInsiderRing.label(changeBPS: 0), "Chi tiêu không đổi so với kỳ trước")
        XCTAssertEqual(MPMoneyInsiderRing.label(changeBPS: nil), "So sánh chi tiêu chưa có dữ liệu")
    }

    func testBudgetCopyMatchesWebTimingSuggestionsAndThresholds() {
        XCTAssertEqual(MPBudgetCopy.timing(ended: true, startDate: "2026-10-01", today: "2026-10-07", daysRemaining: 0), "Đã kết thúc")
        XCTAssertEqual(MPBudgetCopy.timing(ended: false, startDate: "2026-10-10", today: "2026-10-07", daysRemaining: 3), "Sắp bắt đầu")
        XCTAssertEqual(MPBudgetCopy.timing(ended: false, startDate: "2026-10-01", today: "2026-10-07", daysRemaining: -2), "Còn 0 ngày")
        XCTAssertEqual(MPBudgetCopy.dailySuggestion(limit: 1_000_000, spent: 250_000, daysRemaining: 5, masked: false), "Gợi ý chi/ngày: 150.000 ₫")
        XCTAssertEqual(MPBudgetCopy.dailySuggestion(limit: 1_000_000, spent: 250_000, daysRemaining: 5, masked: true), "Gợi ý chi/ngày: \(MPFormat.maskedValue)")
        XCTAssertEqual(MPBudgetCopy.threshold(spent: 750_000, limit: 1_000_000)?.text, "Đã dùng ít nhất 75% ngân sách")
        XCTAssertEqual(MPBudgetCopy.threshold(spent: 1_000_001, limit: 1_000_000)?.text, "Đã vượt ngân sách")
        XCTAssertNil(MPBudgetCopy.threshold(spent: 1, limit: 0))
    }

    func testAIEntryPendingSelectionKeepsOnlyPendingProposals() {
        let draft = EntryDraftModel(
            type: "expense",
            amount: 10,
            walletID: "wallet",
            categoryID: nil,
            jarID: nil,
            occurredAt: "2026-10-07T08:00:00+07:00",
            note: "fixture",
            includedInReports: true
        )
        let proposals = [
            EntryProposalModel(id: "pending", processID: "process", version: 1, status: "pending", draft: draft, questions: [], transactionID: nil),
            EntryProposalModel(id: "approved", processID: "process", version: 1, status: "approved", draft: draft, questions: [], transactionID: "transaction"),
            EntryProposalModel(id: "rejected", processID: "process", version: 1, status: "rejected", draft: draft, questions: [], transactionID: nil)
        ]

        XCTAssertEqual(AIEntryProposalReview.pending(proposals).map(\.id), ["pending"])
    }

    func testAIEntryProposalPresentationUsesCanonicalTransactionRow() {
        let draft = EntryDraftModel(
            type: "expense",
            amount: 45_000,
            walletID: "wallet",
            categoryID: "coffee",
            jarID: nil,
            occurredAt: "2026-10-07T08:00:00+07:00",
            note: "Cà phê sáng",
            includedInReports: true
        )
        let proposal = EntryProposalModel(id: "proposal", processID: "process", version: 1, status: "pending", draft: draft, questions: [], transactionID: nil)
        let category = CategoryModel(id: "coffee", ownerID: "owner", parentID: nil, kind: "expense", name: "Cà phê", systemKey: nil, iconKey: "cup.and.saucer.fill", isSystem: false, walletIDs: [])

        let row = AIEntryProposalReview.transactionRow(for: proposal, category: category, timeZone: TimeZone(identifier: "Asia/Ho_Chi_Minh")!, masked: false)

        XCTAssertEqual(row.title, "Cà phê")
        XCTAssertEqual(row.subtitle, "Cà phê sáng")
        XCTAssertEqual(row.amount, "−45.000 ₫")
        XCTAssertEqual(row.amountTone, .expense)
    }

    func testAIEntryDecisionPolicyLocksOneProposalAndGlobalSave() {
        XCTAssertTrue(AIEntryDecisionPolicy.canDecide(proposalID: "p1", deciding: [], savingAll: false))
        XCTAssertFalse(AIEntryDecisionPolicy.canDecide(proposalID: "p1", deciding: ["p1"], savingAll: false))
        XCTAssertFalse(AIEntryDecisionPolicy.canDecide(proposalID: "p2", deciding: [], savingAll: true))
        XCTAssertFalse(AIEntryDecisionPolicy.canSaveAll(savingAll: false, processing: false, deciding: ["p1"]))
        XCTAssertTrue(AIEntryDecisionPolicy.canSaveAll(savingAll: false, processing: false, deciding: []))
    }
}
