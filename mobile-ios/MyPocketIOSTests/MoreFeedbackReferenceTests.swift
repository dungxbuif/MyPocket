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
        XCTAssertEqual(MoreReferencePresentation.surfaceStyle, .grouped)
        XCTAssertEqual(MoreReferencePresentation.rowSpacing, 0)
    }

    func testMoreReferenceSectionsKeepSinglePrimaryAssistantAndNotificationEntries() {
        let visibleItems = MoreReferencePresentation.sections.flatMap(\.items)

        XCTAssertFalse(visibleItems.contains(.assistant))
        XCTAssertFalse(visibleItems.contains(.notifications))
        XCTAssertEqual(visibleItems.filter { $0 == .aiEntry }.count, 1)
    }

    func testAccountReferenceAnatomyKeepsProfileBadgeAndBankLinkInformational() {
        XCTAssertEqual(MoreReferencePresentation.profileBadgeTitle, "TÀI KHOẢN MYPOCKET")
        XCTAssertEqual(MoreReferencePresentation.bankLinkTitle, "Liên kết ngân hàng")
        XCTAssertFalse(MoreReferencePresentation.bankLinkStyle.isEnabled)
        XCTAssertEqual(MoreReferencePresentation.logoutAccessibilityIdentifier, "more-logout")
        XCTAssertEqual(MoreReferencePresentation.profileAccessibilityIdentifier, "more-profile-account")
    }

    func testFeedbackListUsesOneReferenceGroupAndKeepsResolutionCopyExplicit() {
        XCTAssertEqual(FeedbackReferencePresentation.surfaceStyle, .reference)
        XCTAssertEqual(FeedbackReferencePresentation.dividerSpacing, 0)
        XCTAssertEqual(
            FeedbackReferencePresentation.resolutionHint,
            "Tick khi bạn xác nhận đã hết lỗi. Bỏ tick nếu vấn đề xuất hiện lại."
        )
    }

    func testAuthAndFeedbackSurfacesUseSharedReferenceContracts() {
        XCTAssertEqual(AuthReferencePresentation.cardStyle, .reference)
        XCTAssertEqual(AuthReferencePresentation.contentSpacing, MPMetrics.sectionSpacing)
        XCTAssertEqual(FeedbackReferencePresentation.surfaceStyle, .reference)
    }

    func testAIEntryAndAccountUseReferenceSurfacesWithoutLocalSpacing() {
        XCTAssertEqual(AIEntryReferencePresentation.proposalSurfaceStyle, .reference)
        XCTAssertEqual(AIEntryReferencePresentation.contentSpacing, MPMetrics.sectionSpacing)
        XCTAssertFalse(AIEntryReferencePresentation.saveAllButtonFullWidth)
        XCTAssertEqual(AIEntryReferencePresentation.saveAllSystemName, "checkmark")
        XCTAssertEqual(AccountReferencePresentation.sectionSurfaceStyle, .reference)
        XCTAssertEqual(AccountReferencePresentation.contentSpacing, MPMetrics.sectionSpacing)
    }

    func testAIEntryProposalMetadataKeepsReviewTypeAndAccountLocalDateExplicit() {
        let timeZone = TimeZone(identifier: "Asia/Ho_Chi_Minh")!
        let occurredAt = "2026-10-07T08:00:00+07:00"
        let known = proposal(type: "expense", occurredAt: occurredAt)
        let unknown = proposal(type: "unknown", occurredAt: occurredAt)
        let expectedDate = MPFormat.shortDate(occurredAt, timeZone: timeZone)

        XCTAssertEqual(
            AIEntryProposalReview.reviewMetadata(for: known, timeZone: timeZone),
            "Loại: Chi tiêu · Ngày: \(expectedDate)"
        )
        XCTAssertEqual(
            AIEntryProposalReview.reviewMetadata(for: unknown, timeZone: timeZone),
            "Ngày: \(expectedDate)"
        )
        XCTAssertEqual(
            AIEntryProposalReview.reviewMetadata(for: proposal(type: "unknown", occurredAt: ""), timeZone: timeZone),
            "Ngày: Chưa chọn"
        )
    }

    func testFeedbackDockUsesStableTrailingAnchorAndIgnoresScrollingObstacles() {
        let viewport = CGRect(x: 0, y: 0, width: 320, height: 60)
        let obstacle = CGRect(x: 250, y: 0, width: 60, height: 60)

        XCTAssertEqual(
            MPFeedbackBubbleLayout.frame(in: viewport, clearance: 0, obstacles: []),
            MPFeedbackBubbleLayout.frame(in: viewport, clearance: 0, obstacles: [obstacle])
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

    private func proposal(type: String, occurredAt: String) -> EntryProposalModel {
        EntryProposalModel(
            id: "proposal-\(type)",
            processID: "process",
            version: 1,
            status: "pending",
            draft: EntryDraftModel(
                type: type,
                amount: 45_000,
                walletID: "wallet",
                categoryID: nil,
                jarID: nil,
                occurredAt: occurredAt,
                note: "Bữa sáng",
                includedInReports: true
            ),
            questions: [],
            transactionID: nil
        )
    }
}
