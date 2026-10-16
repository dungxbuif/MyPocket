import XCTest
@testable import MyPocketIOS

final class TransactionsReferenceTests: XCTestCase {
    func testLedgerBalanceUsesIncludedAggregateAndSelectedWalletBalance() {
        let wallets = [wallet("included", balance: 900, included: true), wallet("excluded", balance: -200, included: false)]
        XCTAssertEqual(TransactionsPresentation.scopeBalance(wallets: wallets, selectionID: "__all__"), 900)
        XCTAssertEqual(TransactionsPresentation.scopeBalance(wallets: wallets, selectionID: "excluded"), -200)
        XCTAssertNil(TransactionsPresentation.scopeBalance(wallets: wallets, selectionID: "missing"))
        XCTAssertNil(TransactionsPresentation.scopeBalance(wallets: [], selectionID: "__all__"))
    }

    func testDayHeadingUsesAccountLocalTodayAndFallsBackToWeekday() {
        let timeZone = TimeZone(identifier: "Asia/Ho_Chi_Minh")!
        let now = MPFormat.parseInstant("2026-10-15T18:00:00Z")!
        XCTAssertEqual(TransactionsPresentation.dayHeading("2026-10-16", timeZone: timeZone, now: now), "Hôm nay")
        XCTAssertEqual(TransactionsPresentation.dayHeading("2026-10-15", timeZone: timeZone, now: now), "Thứ Năm")
        XCTAssertEqual(TransactionsPresentation.dayHeading("invalid", timeZone: timeZone, now: now), "invalid")
    }

    func testCategoryGroupedRowUsesAccountLocalDateInsteadOfRepeatingCategory() {
        let timeZone = TimeZone(identifier: "Asia/Ho_Chi_Minh")!
        let now = MPFormat.parseInstant("2026-10-15T18:00:00Z")!
        XCTAssertEqual(TransactionsPresentation.categoryRowDate("2026-10-15T18:00:00Z", timeZone: timeZone, now: now), "16 tháng 10 2026, Hôm nay")
    }

    private func wallet(_ id: String, balance: Int64, included: Bool) -> WalletModel {
        WalletModel(id: id, name: id, type: "basic", currency: "VND", openingBalance: 0,
                    currentBalance: balance, availableCredit: nil, isInTotal: included,
                    description: nil, targetAmount: nil, targetDate: nil, creditLimit: nil,
                    lastStatementBalance: nil, statementDay: nil, paymentDueDay: nil)
    }
}
