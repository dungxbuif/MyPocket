import XCTest
@testable import MyPocketIOS

final class TransactionsReferenceTests: XCTestCase {
    func testTransactionDetailRowsUseCategoryWalletNoteLocalDateAndReportStatus() {
        let transaction = transaction(note: "  Cà phê với Mai  ", includedInReports: false)
        let rows = TransactionDetailPresentation.rows(
            transaction: transaction,
            categoryName: "Ăn uống",
            walletName: "Ví chính",
            timeZone: TimeZone(identifier: "Asia/Ho_Chi_Minh")!
        )

        XCTAssertEqual(rows, [
            TransactionDetailPresentation.Row(title: "Danh mục", value: "Ăn uống"),
            TransactionDetailPresentation.Row(title: "Ví", value: "Ví chính"),
            TransactionDetailPresentation.Row(title: "Ghi chú", value: "Cà phê với Mai"),
            TransactionDetailPresentation.Row(title: "Ngày", value: "16/10/2026"),
            TransactionDetailPresentation.Row(title: "Báo cáo", value: "Không tính vào báo cáo")
        ])
    }

    func testTransactionDetailRowsOmitBlankNoteAndUseFallbackLabels() {
        let rows = TransactionDetailPresentation.rows(
            transaction: transaction(note: " \n ", includedInReports: true),
            categoryName: nil,
            walletName: nil,
            timeZone: TimeZone(identifier: "Asia/Ho_Chi_Minh")!
        )

        XCTAssertEqual(rows, [
            TransactionDetailPresentation.Row(title: "Danh mục", value: "Khoản chi"),
            TransactionDetailPresentation.Row(title: "Ví", value: "Ví đã xóa"),
            TransactionDetailPresentation.Row(title: "Ngày", value: "16/10/2026"),
            TransactionDetailPresentation.Row(title: "Báo cáo", value: "Được tính vào báo cáo")
        ])
    }

    func testTransactionDetailRowsCanOmitUnavailableCallerContext() {
        let rows = TransactionDetailPresentation.rows(
            transaction: transaction(note: "Cà phê", includedInReports: true),
            categoryName: nil,
            walletName: nil,
            timeZone: TimeZone(identifier: "Asia/Ho_Chi_Minh")!,
            includesContext: false
        )

        XCTAssertEqual(rows.map(\.title), ["Ghi chú", "Ngày", "Báo cáo"])
        XCTAssertFalse(rows.contains { $0.value == "Ví đã xóa" })
    }

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

    private func transaction(note: String?, includedInReports: Bool) -> TransactionModel {
        TransactionModel(
            id: "transaction", ownerID: "owner", walletID: "wallet", categoryID: nil,
            jarID: nil, jarName: nil, type: "expense", amount: 120_000,
            adjustmentDirection: nil, creditKind: nil, creditPaymentID: nil,
            travelEventID: nil, occurredAt: "2026-10-15T18:00:00Z", note: note,
            includedInReports: includedInReports, transferID: nil,
            createdAt: "2026-10-15T18:00:00Z", updatedAt: "2026-10-15T18:00:00Z"
        )
    }
}
