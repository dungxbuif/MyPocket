import XCTest
@testable import MyPocketIOS

@MainActor
final class BillsDebtTravelReferenceTests: XCTestCase {
    func testSecondaryRoutesExposeReferenceDensityContracts() {
        XCTAssertEqual(TravelReferenceCopy.addAction, "Thêm")
        XCTAssertEqual(PortfolioReferenceCopy.addAssetAction, "Thêm tài sản")
        XCTAssertEqual(PortfolioReferenceCopy.saveTradeAction, "Lưu giao dịch")
        XCTAssertEqual(PortfolioView.PortfolioTradePresentation.countLabel(0), "0 giao dịch")
        XCTAssertEqual(PortfolioView.PortfolioTradePresentation.countLabel(1), "1 giao dịch")
        XCTAssertEqual(PortfolioView.PortfolioTradePresentation.countLabel(2), "2 giao dịch")
    }

    func testCreditSummaryKeepsPopulatedStatementValuesCompactAndMasked() {
        let statement = CreditStatementModel(
            walletID: "credit",
            creditLimit: 30_000_000,
            balance: 7_200_000,
            availableCredit: 22_800_000,
            amountDue: 1_100_000,
            paymentStatus: "upcoming",
            paymentDueAt: "2026-11-05T00:00:00Z",
            items: []
        )

        let items = CreditLedgerPresentation.summaryItems(statement, masked: false)
        XCTAssertEqual(items.count, 4)
        XCTAssertEqual(items.map(\.title), ["Dư nợ", "Còn khả dụng", "Hạn mức", "Kỳ này cần trả"])
        XCTAssertTrue(CreditLedgerPresentation.summaryItems(statement, masked: true).allSatisfy { $0.value == MPFormat.maskedValue })
    }

    func testTravelValidationPreservesDateRangeAndNameRules() {
        XCTAssertTrue(TravelEventDraftValidator.isValid(name: "Tokyo", startsOn: "2026-11-01", endsOn: "2026-11-07"))
        XCTAssertTrue(TravelEventDraftValidator.isValid(name: "Open ended", startsOn: "2026-11-01", endsOn: ""))
        XCTAssertFalse(TravelEventDraftValidator.isValid(name: "", startsOn: "2026-11-01", endsOn: "2026-11-07"))
        XCTAssertFalse(TravelEventDraftValidator.isValid(name: "Reversed", startsOn: "2026-11-07", endsOn: "2026-11-01"))
    }
}
