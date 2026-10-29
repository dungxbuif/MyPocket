import XCTest
@testable import MyPocketIOS

@MainActor
final class CreditNotificationsReferenceTests: XCTestCase {
    func testCreditReferencePopulatedSummaryUsesCompactFourMetricContract() {
        let statement = CreditStatementModel(
            walletID: "credit",
            creditLimit: 10_000_000,
            balance: 2_400_000,
            availableCredit: 7_600_000,
            amountDue: 900_000,
            paymentStatus: "upcoming",
            paymentDueAt: "2026-11-05T00:00:00Z",
            items: []
        )

        XCTAssertEqual(CreditLedgerPresentation.statementCountLabel(1), "1 giao dịch")
        XCTAssertEqual(CreditLedgerPresentation.statementCountLabel(3), "3 giao dịch")
        XCTAssertEqual(CreditLedgerPresentation.summaryItems(statement, masked: false).map(\.title), [
            "Dư nợ", "Còn khả dụng", "Hạn mức", "Kỳ này cần trả"
        ])
        XCTAssertTrue(CreditLedgerPresentation.summaryItems(statement, masked: true).allSatisfy { $0.value == MPFormat.maskedValue })

        let supporting = CreditLedgerPresentation.supportingSummaryItems(statement, masked: true)
        XCTAssertEqual(supporting.map(\.value), [
            MPFormat.money(7_600_000), MPFormat.money(10_000_000), MPFormat.money(900_000)
        ])
        XCTAssertTrue(supporting.allSatisfy(\.masked))
    }

    func testCreditActionFormUsesSharedSectionSpacing() {
        XCTAssertEqual(CreditReferencePresentation.actionFormSpacing, MPMetrics.sectionSpacing)
    }

    func testCreditActionSaveEligibilityRequiresActionSpecificCurrentSelection() {
        XCTAssertTrue(CreditActionPolicy.saveIsDisabled(
            action: "entry", amount: "12000", categoryID: "",
            applicableCategoryIDs: ["food"], sourceWalletID: "cash",
            eligibleSourceWalletIDs: ["cash"]
        ))
        XCTAssertFalse(CreditActionPolicy.saveIsDisabled(
            action: "entry", amount: "12000", categoryID: "food",
            applicableCategoryIDs: ["food"], sourceWalletID: "",
            eligibleSourceWalletIDs: []
        ))
        XCTAssertTrue(CreditActionPolicy.saveIsDisabled(
            action: "payment", amount: "12000", categoryID: "",
            applicableCategoryIDs: [], sourceWalletID: "missing",
            eligibleSourceWalletIDs: ["cash"]
        ))
        XCTAssertFalse(CreditActionPolicy.saveIsDisabled(
            action: "payment", amount: "12000", categoryID: "",
            applicableCategoryIDs: [], sourceWalletID: "cash",
            eligibleSourceWalletIDs: ["cash"]
        ))
    }

    func testNotificationsReferenceSeparatesUnreadAndReadWithoutChangingOrder() {
        let items = [
            NotificationModel(id: "read-1", kind: "budget", title: "Đã đọc", body: "A", readAt: "2026-10-01T00:00:00Z", createdAt: "2026-10-01T00:00:00Z"),
            NotificationModel(id: "new-1", kind: "goal", title: "Chưa đọc", body: "B", readAt: nil, createdAt: "2026-10-02T00:00:00Z"),
            NotificationModel(id: "new-2", kind: "feedback", title: "Chưa đọc 2", body: "C", readAt: nil, createdAt: "2026-10-03T00:00:00Z")
        ]

        XCTAssertEqual(NotificationsPresentation.unread(items).map(\.id), ["new-1", "new-2"])
        XCTAssertEqual(NotificationsPresentation.read(items).map(\.id), ["read-1"])
        XCTAssertEqual(NotificationsPresentation.sectionTitle(unread: true, count: 2), "Chưa đọc · 2")
        XCTAssertEqual(NotificationsPresentation.sectionTitle(unread: false, count: 1), "Đã đọc · 1")
    }

    func testNotificationsReferenceUsesOneSharedGroupedSurfacePerSection() {
        XCTAssertEqual(NotificationsPresentation.surfaceStyle, .grouped)
        XCTAssertEqual(NotificationsPresentation.rowSpacing, 0)
        XCTAssertEqual(NotificationsPresentation.sectionSpacing, MPMetrics.sectionSpacing)
    }
}
