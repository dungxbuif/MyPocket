import Foundation
import XCTest
@testable import MyPocketIOS

final class FinanceModelTests: XCTestCase {
    func testSnakeCaseFinancialPayloadDecodesAcronyms() throws {
        let data = Data(#"{"id":"tx-1","owner_id":"u-1","wallet_id":"w-1","category_id":"c-1","type":"expense","amount":50000,"occurred_at":"2026-10-05T10:00:00Z","included_in_reports":true,"created_at":"2026-10-05T10:00:00Z","updated_at":"2026-10-05T10:00:00Z"}"#.utf8)
        let transaction = try JSONDecoder.mypocket.decode(TransactionModel.self, from: data)
        XCTAssertEqual(transaction.ownerID, "u-1")
        XCTAssertEqual(transaction.walletID, "w-1")
        XCTAssertEqual(transaction.categoryID, "c-1")
        XCTAssertTrue(transaction.includedInReports)
    }

    func testTransactionInputEncodesBackendSnakeCase() throws {
        let input = TransactionInputModel(walletID: "w-1", categoryID: "c-1", jarID: nil, type: "expense", amount: 1000, occurredAt: "2026-10-05", note: nil, includedInReports: true, travelEventID: nil)
        let data = try JSONEncoder.mypocket.encode(input)
        let json = String(decoding: data, as: UTF8.self)
        XCTAssertTrue(json.contains("wallet_id"))
        XCTAssertTrue(json.contains("included_in_reports"))
    }

    func testPortfolioSummaryDecodesTradesAndNullableValuation() throws {
        let data = Data(#"{"assets":[{"id":"asset-1","symbol":"VNM","name":"Vinamilk","latest_price":null}],"positions":[],"trades":[{"id":"trade-1","asset_id":"asset-1","side":"buy","quantity":"1.25","unit_price":72000,"fee":1000,"occurred_at":"2026-10-05T10:00:00Z"}],"total_cost_basis":91000,"total_market_value":null,"total_unrealized_pnl":null}"#.utf8)
        let summary = try JSONDecoder.mypocket.decode(PortfolioSummaryModel.self, from: data)
        XCTAssertEqual(summary.trades.first?.quantity, "1.25")
        XCTAssertNil(summary.totalMarketValue)
    }

    func testAIEntryDraftEncodesAllEditableFields() throws {
        let draft = EntryDraftModel(type: "expense", amount: 50000, walletID: "wallet-1", categoryID: "category-1", jarID: nil, occurredAt: "2026-10-05T10:00:00Z", note: "Cà phê", includedInReports: true)
        let json = String(decoding: try JSONEncoder.mypocket.encode(draft), as: UTF8.self)
        XCTAssertTrue(json.contains("wallet_id"))
        XCTAssertTrue(json.contains("included_in_reports"))
        XCTAssertTrue(json.contains("Cà phê"))
    }

    func testWalletInputEncodesGoalAndCreditMetadata() throws {
        let input = WalletModelInput(name: "Credit", type: "credit", openingBalance: 0, isInTotal: true, description: nil, targetAmount: nil, targetDate: nil, creditLimit: 3_000_000, lastStatementBalance: 250_000, statementDay: 15, paymentDueDay: 5)
        let json = String(decoding: try JSONEncoder.mypocket.encode(input), as: UTF8.self)
        XCTAssertTrue(json.contains("credit_limit"))
        XCTAssertTrue(json.contains("last_statement_balance"))
        XCTAssertTrue(json.contains("statement_day"))
        XCTAssertTrue(json.contains("payment_due_day"))
    }

    func testCalendarDateRoundTripDoesNotUseDeviceTimezone() {
        let date = MPFormat.parseCalendarDate("2026-12-31")
        XCTAssertEqual(date.map(MPFormat.calendarDateString), "2026-12-31")
    }
}
