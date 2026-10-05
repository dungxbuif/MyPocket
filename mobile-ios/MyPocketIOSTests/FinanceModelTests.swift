import Foundation
import XCTest
@testable import MyPocketIOS

@MainActor
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

    func testBulkDeleteInputEncodesUniqueIDContractShape() throws {
        let input = BulkDeleteInputModel(transactionIDs: ["tx-1", "tx-2"])
        let payload = try JSONSerialization.jsonObject(with: JSONEncoder.mypocket.encode(input)) as? [String: [String]]
        XCTAssertEqual(payload?["transaction_ids"], ["tx-1", "tx-2"])
    }

    func testBulkDeleteSelectionExcludesImmutableRows() {
        let ordinary = transaction(id: "ordinary", occurredAt: "2026-10-06T00:00:00Z")
        var transfer = ordinary
        transfer.transferID = "transfer-1"
        var adjustment = ordinary
        adjustment.type = "adjustment"
        var credit = ordinary
        credit.creditKind = "purchase"

        XCTAssertTrue(TransactionsViewModel.bulkDeleteSelectable(ordinary))
        XCTAssertFalse(TransactionsViewModel.bulkDeleteSelectable(transfer))
        XCTAssertFalse(TransactionsViewModel.bulkDeleteSelectable(adjustment))
        XCTAssertFalse(TransactionsViewModel.bulkDeleteSelectable(credit))
    }

    func testOrdinaryTransactionWalletsExcludeCreditWallets() {
        let wallets = [
            WalletModel(id: "basic", name: "Tiền mặt", type: "basic", currency: "VND", openingBalance: 0, currentBalance: 0, availableCredit: nil, isInTotal: true, description: nil, targetAmount: nil, targetDate: nil, creditLimit: nil, lastStatementBalance: nil, statementDay: nil, paymentDueDay: nil),
            WalletModel(id: "credit", name: "Thẻ", type: "credit", currency: "VND", openingBalance: 0, currentBalance: 0, availableCredit: 10_000_000, isInTotal: true, description: nil, targetAmount: nil, targetDate: nil, creditLimit: 10_000_000, lastStatementBalance: nil, statementDay: 15, paymentDueDay: 5)
        ]
        XCTAssertEqual(TransactionsViewModel.ordinaryTransactionWallets(wallets).map(\.id), ["basic"])
    }

    func testAdvisorOverviewDecodesStructuredViewAndSource() throws {
        let data = Data(#"{"view_kind":"metric_group","view":{"income":1200000,"expense":500000},"source":{"tool":"get_finance_summary"},"status":"ok"}"#.utf8)
        let overview = try JSONDecoder.mypocket.decode(AdvisorOverviewModel.self, from: data)
        XCTAssertEqual(overview.viewKind, "metric_group")
        XCTAssertEqual(overview.status, "ok")
        if case .object(let view)? = overview.view {
            XCTAssertEqual(view["income"]?.prettyDescription, "1200000")
        } else {
            XCTFail("overview view should remain structured JSON")
        }
    }

    func testCreditEntryCategoriesRespectKindAndWalletScope() {
        let creditWalletID = "credit-wallet"
        let purchaseCategory = CategoryModel(id: "purchase", ownerID: nil, parentID: nil, kind: "expense", name: "Mua sắm", systemKey: nil, iconKey: "tag", isSystem: false, walletIDs: [creditWalletID])
        let refundCategory = CategoryModel(id: "refund", ownerID: nil, parentID: nil, kind: "income", name: "Hoàn tiền", systemKey: nil, iconKey: "tag", isSystem: false, walletIDs: [creditWalletID])
        XCTAssertTrue(purchaseCategory.applies(to: "expense", walletID: creditWalletID, walletType: "credit"))
        XCTAssertFalse(purchaseCategory.applies(to: "income", walletID: creditWalletID, walletType: "credit"))
        XCTAssertTrue(refundCategory.applies(to: "income", walletID: creditWalletID, walletType: "credit"))
        XCTAssertFalse(refundCategory.applies(to: "income", walletID: "other-wallet", walletType: "credit"))
    }

    func testNativeGoogleFixtureExchangeEncodesIOSPKCEContract() throws {
        let request = MobileGoogleExchangeRequest(
            code: "fixture",
            codeVerifier: "fixture",
            redirectURI: "mypocket://oauth/callback",
            platform: "ios"
        )
        let payload = try JSONSerialization.jsonObject(with: JSONEncoder.mypocket.encode(request)) as? [String: String]
        XCTAssertEqual(payload?["code"], "fixture")
        XCTAssertEqual(payload?["code_verifier"], "fixture")
        XCTAssertEqual(payload?["redirect_uri"], "mypocket://oauth/callback")
        XCTAssertEqual(payload?["platform"], "ios")
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

    func testCategoryInputPreservesParentID() throws {
        let input = CategoryModelInput(name: "Highlands", kind: "expense", parentID: "food", walletIDs: [], iconKey: "tag")
        let json = String(decoding: try JSONEncoder.mypocket.encode(input), as: UTF8.self)
        XCTAssertTrue(json.contains("\"parent_id\":\"food\""))
    }

    func testCategoryInputPreservesMultipleWalletScopes() throws {
        let input = CategoryModelInput(name: "Highlands", kind: "expense", parentID: "food", walletIDs: ["wallet-1", "wallet-2"], iconKey: "tag")
        let decoded = try JSONDecoder.mypocket.decode(CategoryModelInput.self, from: JSONEncoder.mypocket.encode(input))
        XCTAssertEqual(decoded.walletIDs, ["wallet-1", "wallet-2"])
    }

    func testRecurringInputUsesRFC3339InstantAndNote() throws {
        let input = RecurringInputModel(name: "Tiền nhà", walletID: "wallet-1", categoryID: nil, type: "expense", amount: 5_000_000, note: "Hợp đồng", frequency: "monthly", interval: 1, nextRunAt: "2026-10-05T02:30:00Z", endsAt: nil, active: true)
        let json = String(decoding: try JSONEncoder.mypocket.encode(input), as: UTF8.self)
        XCTAssertTrue(json.contains("2026-10-05T02:30:00Z"))
        XCTAssertTrue(json.contains("Hợp đồng"))
    }

    func testCategoryApplicabilityRespectsKindGoalWalletAndScope() {
        let food = CategoryModel(id: "food", ownerID: "u", parentID: nil, kind: "expense", name: "Ăn uống", systemKey: nil, iconKey: "fork.knife", isSystem: false, walletIDs: [])
        XCTAssertTrue(food.applies(to: "expense", walletID: "basic", walletType: "basic"))
        XCTAssertFalse(food.applies(to: "income", walletID: "basic", walletType: "basic"))
        XCTAssertFalse(food.applies(to: "expense", walletID: "goal", walletType: "goal"))

        let scoped = CategoryModel(id: "scoped", ownerID: "u", parentID: nil, kind: "expense", name: "Ví riêng", systemKey: nil, iconKey: "tag", isSystem: false, walletIDs: ["wallet-1"])
        XCTAssertTrue(scoped.applies(to: "expense", walletID: "wallet-1", walletType: "basic"))
        XCTAssertFalse(scoped.applies(to: "expense", walletID: "wallet-2", walletType: "basic"))
    }

    func testInstantFormatSupportsFractionalSecondsAndAccountTimezone() {
        let instant = "2026-10-01T00:30:00.123Z"
        XCTAssertNotNil(MPFormat.parseInstant(instant))
        XCTAssertEqual(MPFormat.accountMonth(of: instant, timeZone: TimeZone(identifier: "Asia/Ho_Chi_Minh")!), "2026-10")
        XCTAssertEqual(MPFormat.accountMonth(of: instant, timeZone: TimeZone(identifier: "America/Los_Angeles")!), "2026-09")
    }

    func testTransactionPeriodFilterExcludesMalformedTimestamps() {
        let valid = transaction(id: "valid", occurredAt: "2026-10-06T03:00:00Z")
        let malformed = transaction(id: "malformed", occurredAt: "not-a-timestamp")
        let day = MPFormat.parseCalendarDate("2026-10-06")!

        let result = TransactionsViewModel.applyFilters(
            [valid, malformed],
            walletSelection: "__all__",
            periodSelection: "custom",
            customStart: day,
            customEnd: day,
            accountTimezone: "Asia/Ho_Chi_Minh"
        )

        XCTAssertEqual(result.map(\.id), ["valid"])
    }

    private func transaction(id: String, occurredAt: String) -> TransactionModel {
        TransactionModel(
            id: id,
            ownerID: "owner",
            walletID: "wallet",
            categoryID: nil,
            jarID: nil,
            jarName: nil,
            type: "expense",
            amount: 10_000,
            adjustmentDirection: nil,
            creditKind: nil,
            creditPaymentID: nil,
            travelEventID: nil,
            occurredAt: occurredAt,
            note: nil,
            includedInReports: true,
            transferID: nil,
            createdAt: occurredAt,
            updatedAt: occurredAt
        )
    }
}
