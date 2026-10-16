import XCTest
@testable import MyPocketIOS

@MainActor
final class MoreFeedbackReferenceTests: XCTestCase {
    func testMoreReferenceSectionsKeepAccountGroupsAndDestinationOrder() {
        XCTAssertEqual(
            MoreReferencePresentation.sections.map(\.title),
            ["Phản hồi", "Ví của tôi", "Kế hoạch", "Giao dịch", "Trợ lý"]
        )
        XCTAssertEqual(MoreReferencePresentation.sections[1].items, [.wallets, .credit, .portfolio])
        XCTAssertEqual(MoreReferencePresentation.sections[2].items, [.budgets, .jars, .months])
    }

    func testMoreReferenceSectionsUseNamedGroupedSurfaceAndCompactRowRhythm() {
        XCTAssertEqual(MoreReferencePresentation.surfaceStyle, .reference)
        XCTAssertEqual(MoreReferencePresentation.rowSpacing, 0)
        XCTAssertEqual(MoreReferencePresentation.sectionSpacing, MPMetrics.sectionSpacing)
    }

    func testFeedbackTabsKeepOwnerResolutionMeaningAndLabels() {
        XCTAssertEqual(FeedbackReferencePresentation.tabID(forResolved: false), "unresolved")
        XCTAssertEqual(FeedbackReferencePresentation.tabID(forResolved: true), "resolved")
        XCTAssertEqual(FeedbackReferencePresentation.tabTitle(forResolved: false, count: 2), "Chưa giải quyết (2)")
        XCTAssertEqual(FeedbackReferencePresentation.tabTitle(forResolved: true, count: 1), "Đã giải quyết (1)")
    }

    func testFeedbackListUsesOneReferenceGroupAndKeepsResolutionCopyExplicit() {
        XCTAssertEqual(FeedbackReferencePresentation.surfaceStyle, .reference)
        XCTAssertEqual(FeedbackReferencePresentation.dividerSpacing, 0)
        XCTAssertEqual(
            FeedbackReferencePresentation.resolutionHint,
            "Tick khi bạn xác nhận đã hết lỗi. Bỏ tick nếu vấn đề xuất hiện lại."
        )
    }

    func testReferenceIconRegistryKeepsAccountGlyphsOnNativeSymbols() {
        XCTAssertEqual(MPReferenceIcon.account, "person.crop.circle")
        XCTAssertEqual(MPReferenceIcon.wallet, "wallet.pass.fill")
        XCTAssertEqual(MPReferenceIcon.group, "shippingbox.fill")
        XCTAssertEqual(MPReferenceIcon.bankLink, "link.circle.fill")
        XCTAssertEqual(MPReferenceIcon.recurring, "repeat")
        XCTAssertEqual(MPReferenceIcon.feedback, "bubble.left.and.bubble.right.fill")
    }
}
