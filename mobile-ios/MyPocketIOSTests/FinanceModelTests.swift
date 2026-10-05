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
}
