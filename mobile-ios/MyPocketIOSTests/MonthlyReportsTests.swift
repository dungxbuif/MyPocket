import Foundation
import XCTest
@testable import MyPocketIOS

final class MonthlyReportsTests: XCTestCase {
    func testCanonicalReportPayloadDecodesSnapshotIdentityAndAccountingTotals() throws {
        let payload = """
        {
          "scope": {"month": "2026-10", "granularity": "week"},
          "timezone": "Asia/Ho_Chi_Minh",
          "snapshot_id": "snapshot-1",
          "data_revision": 7,
          "totals": {"income": 1200000, "expense": 450000, "net": 750000, "transaction_count": 4},
          "series": [{"label": "2026-10-05", "income": 1200000, "expense": 450000, "net": 750000, "transaction_count": 4}],
          "categories": [{"category_id": "food", "name": "Ăn uống", "amount": 450000, "transaction_count": 4}],
          "insights": [{"id": "insight-1", "kind": "category_contribution", "title": "Chi tiêu", "body": "Theo dõi nhóm này."}]
        }
        """.data(using: .utf8)!

        let report = try JSONDecoder().decode(MonthlyReportModel.self, from: payload)

        XCTAssertEqual(report.month, "2026-10")
        XCTAssertEqual(report.timezone, "Asia/Ho_Chi_Minh")
        XCTAssertEqual(report.snapshotID, "snapshot-1")
        XCTAssertEqual(report.dataRevision, 7)
        XCTAssertEqual(report.income, 1_200_000)
        XCTAssertEqual(report.expense, 450_000)
        XCTAssertEqual(report.net, 750_000)
        XCTAssertEqual(report.transactionCount, 4)
        XCTAssertEqual(report.weekly.count, 1)
        XCTAssertTrue(report.daily.isEmpty)
        XCTAssertEqual(report.categories.first?.id, "food")
        XCTAssertEqual(report.insight, "Theo dõi nhóm này.")
    }

    func testLegacyReportPayloadStillDecodesWithoutLedgerRows() throws {
        let payload = """
        {"month":"2026-09","income":100,"expense":40,"net":60,"transactionCount":2,"weekly":[],"daily":[],"categories":[]}
        """.data(using: .utf8)!

        let report = try JSONDecoder().decode(MonthlyReportModel.self, from: payload)

        XCTAssertEqual(report.month, "2026-09")
        XCTAssertEqual(report.income, 100)
        XCTAssertEqual(report.expense, 40)
        XCTAssertEqual(report.net, 60)
        XCTAssertEqual(report.transactionCount, 2)
        XCTAssertTrue(report.weekly.isEmpty)
        XCTAssertTrue(report.daily.isEmpty)
    }
}
