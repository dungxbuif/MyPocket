import Foundation
import XCTest
@testable import MyPocketIOS

private struct ReferenceSnapshotKeys: OutboxKeyProviding {
    func key(for namespace: String, createIfMissing: Bool) throws -> Data {
        Data(repeating: 0x52, count: 32)
    }
}

@MainActor
final class JarsMonthsReferenceTests: XCTestCase {
    func testFinanceComparisonChartNormalizesSharedReportValuesWithoutInventingBars() {
        XCTAssertEqual(MPFinanceComparisonChart.normalizedHeights(values: [300, 150]), [100, 50])
        XCTAssertEqual(MPFinanceComparisonChart.normalizedHeights(values: [0, 250]), [0, 100])
        XCTAssertEqual(MPFinanceComparisonChart.normalizedHeights(values: [0, 0]), [0, 0])
        XCTAssertEqual(MPFinanceComparisonChart.normalizedHeights(values: [-50, 150]), [0, 100])
    }

    func testFinanceComparisonChartMasksEveryAccessibilityAmount() {
        let items = [
            MPFinanceComparisonItem(label: "Thu nhập", value: 500_000, tone: .income),
            MPFinanceComparisonItem(label: "Chi tiêu", value: 120_000, tone: .expense)
        ]

        XCTAssertEqual(
            MPFinanceComparisonChart.accessibilityValue(items: items, masked: false),
            "Thu nhập \(MPFormat.money(500_000)); Chi tiêu \(MPFormat.money(120_000))"
        )
        let masked = MPFinanceComparisonChart.accessibilityValue(items: items, masked: true)
        XCTAssertEqual(masked, "Thu nhập \(MPFormat.maskedValue); Chi tiêu \(MPFormat.maskedValue)")
        XCTAssertFalse(masked.contains("500.000"))
        XCTAssertFalse(masked.contains("120.000"))
    }

    func testFinanceSummaryUsesSafeBudgetPercentageAndStableSharedContract() {
        XCTAssertEqual(MPFinanceSummaryMath.percentage(numerator: 250, denominator: 500), 50)
        XCTAssertEqual(MPFinanceSummaryMath.percentage(numerator: 625, denominator: 500), 125)
        XCTAssertEqual(MPFinanceSummaryMath.percentage(numerator: -20, denominator: 500), 0)
        XCTAssertEqual(MPFinanceSummaryMath.percentage(numerator: 20, denominator: 0), 0)
        XCTAssertEqual(MPFinanceSummaryCardPolicy.surfaceStyle, .reference)
        XCTAssertEqual(MPFinanceSummaryCardPolicy.accessibilityID("budget"), "finance-summary-budget")
        XCTAssertEqual(MPFinanceSummaryCardPolicy.accessibilityID("savings-goals"), "finance-summary-savings-goals")
    }

    func testPlanningAndReportScreensUseOneReferenceFinanceCardBase() {
        XCTAssertEqual(MPFinanceCardPolicy.surfaceStyle, .reference)
        XCTAssertEqual(MPFinanceCardPolicy.contentSpacing, MPMetrics.sectionSpacing)
    }

    func testJarsKeepStableCreateActionAndMonthReportCopy() {
        XCTAssertEqual(JarsReferenceCopy.addAction, "Thêm hũ")
        XCTAssertEqual(JarsReferenceCopy.addActionAccessibilityID, "jar-add-month")
        XCTAssertEqual(MonthsPresentation.timezoneLabel("Asia/Ho_Chi_Minh"), "theo múi giờ tài khoản")
    }

    func testJarsReferenceCopyMakesMonthlyFlowAndRemovalScopeExplicit() {
        XCTAssertEqual(JarsFlowCopy.monthSelectorTitle, "Tháng đang xem")
        XCTAssertEqual(JarsFlowCopy.summaryTitle, "Tổng quan tháng")
        XCTAssertEqual(JarsFlowCopy.listTitle, "Hũ tháng")
        XCTAssertEqual(JarsFlowCopy.unassignedTitle, "Chi chưa gắn hũ")
        XCTAssertEqual(JarsFlowCopy.cumulativeTitle, "Cộng dồn")
        XCTAssertEqual(JarsFlowCopy.removalTitle, "Gỡ hũ khỏi tháng?")
        XCTAssertTrue(JarsFlowCopy.removalMessage.lowercased().contains("chỉ cấu hình của tháng này"))
        XCTAssertEqual(JarsFlowCopy.removalConfirm, "Gỡ khỏi tháng")
        XCTAssertEqual(JarsReferencePresentation.monthContext(month: "2026-10", currentMonth: "2026-10"), "Tháng 10 năm 2026 · Tháng hiện tại")
        XCTAssertEqual(JarsReferencePresentation.monthContext(month: "2026-09", currentMonth: "2026-10"), "Tháng 09 năm 2026 · Lịch sử")
    }

    func testJarReferenceProgressIsOptionalAndClampedForOverspend() {
        XCTAssertNil(JarsReferencePresentation.progress(spent: 100, allocation: nil))
        XCTAssertEqual(JarsReferencePresentation.progress(spent: 250, allocation: 200), 1)
        XCTAssertEqual(JarsReferencePresentation.progress(spent: 50, allocation: 200), 0.25)
        XCTAssertTrue(JarsReferencePresentation.isOverspent(spent: 250, allocation: 200))
        XCTAssertFalse(JarsReferencePresentation.isOverspent(spent: 50, allocation: 200))
    }

    func testJarAllocationDraftMatchesBackendModesAndZeroPercentContract() {
        XCTAssertTrue(JarAllocationDraftValidator.isValid(mode: "none", amount: nil, percent: nil))
        XCTAssertTrue(JarAllocationDraftValidator.isValid(mode: "fixed", amount: 1, percent: nil))
        XCTAssertTrue(JarAllocationDraftValidator.isValid(mode: "percent", amount: nil, percent: 0))
        XCTAssertFalse(JarAllocationDraftValidator.isValid(mode: "amount", amount: 1, percent: nil))
    }

    func testPairedTransferLegsAreExcludedFromReportTrendAndWeekTotals() {
        let occurredAt = "2026-10-20T03:00:00Z"
        let transactions = [
            reportTransaction(id: "transfer-out", type: "expense", amount: 500_000, occurredAt: occurredAt, includedInReports: false, transferID: "pair-1"),
            reportTransaction(id: "transfer-in", type: "income", amount: 500_000, occurredAt: occurredAt, includedInReports: false, transferID: "pair-1"),
            reportTransaction(id: "ordinary-expense", type: "expense", amount: 12_500, occurredAt: occurredAt, includedInReports: true)
        ]

        XCTAssertEqual(
            MPTrendData.points(transactions: transactions, month: "2026-10", timeZone: TimeZone(secondsFromGMT: 0)!),
            [MPTrendPoint(dayKey: "2026-10-20", income: 0, expense: 12_500)]
        )
        let week = MPWeekReportData.totals(
            transactions: transactions,
            now: MPFormat.parseInstant("2026-10-20T12:00:00Z")!,
            timeZone: TimeZone(secondsFromGMT: 0)!
        )
        XCTAssertEqual(week.income, 0)
        XCTAssertEqual(week.expense, 12_500)
    }

    func testJarEditorDraftRehydratesWhenMonthScopeChanges() {
        var session = SecondaryEditorDraftSession()

        XCTAssertTrue(session.begin(itemID: "jar-1", scopeID: "2026-10"))
        session.cancel()
        XCTAssertFalse(session.begin(itemID: "jar-1", scopeID: "2026-10"))
        XCTAssertTrue(session.begin(itemID: "jar-1", scopeID: "2026-11"))
    }

    func testReportsOfflineSnapshotRequiresBothPayloadsForTheExactMonth() {
        let october = reportSnapshot(monthSummaryMonth: "2026-10", insiderMonth: "2026-10")
        XCTAssertTrue(ReportsOfflinePolicy.accepts(october, month: "2026-10"))
        XCTAssertFalse(ReportsOfflinePolicy.accepts(october, month: "2026-11"))
        XCTAssertFalse(ReportsOfflinePolicy.accepts(reportSnapshot(monthSummaryMonth: "2026-10", insiderMonth: "2026-09"), month: "2026-10"))
    }

    func testReportsOfflinePolicyAcceptsPersistedExactMonthAfterCacheRestart() async throws {
        let directory = FileManager.default.temporaryDirectory
            .appendingPathComponent("reference-report-cache-" + UUID().uuidString, isDirectory: true)
        defer { try? FileManager.default.removeItem(at: directory) }
        let server = URL(string: "https://reference-report.test/api/v1")!
        let cache = try FinancialSnapshotCache(directory: directory, serverURL: server, keys: ReferenceSnapshotKeys())

        try await cache.save(reportSnapshot(monthSummaryMonth: "2026-10", insiderMonth: "2026-10"))

        let restarted = try FinancialSnapshotCache(directory: directory, serverURL: server, keys: ReferenceSnapshotKeys())
        let persisted = try await restarted.load(ownerID: "owner")
        XCTAssertTrue(ReportsOfflinePolicy.accepts(persisted, month: "2026-10"))
        XCTAssertFalse(ReportsOfflinePolicy.accepts(persisted, month: "2026-11"))
    }

    func testMonthReportPeriodUsesAccountTimezoneAndInclusiveEnd() {
        let period = MonthsPresentation.reportPeriod(
            startAt: "2026-09-01T00:00:00Z",
            nextStartAt: "2026-10-01T00:00:00Z",
            timeZone: TimeZone(identifier: "Asia/Ho_Chi_Minh")!
        )
        XCTAssertEqual(period, "01/09/2026 – 30/09/2026")
    }

    private func reportTransaction(
        id: String,
        type: String,
        amount: Int64,
        occurredAt: String,
        includedInReports: Bool,
        transferID: String? = nil
    ) -> TransactionModel {
        TransactionModel(
            id: id, ownerID: "owner", walletID: "wallet", categoryID: nil,
            jarID: nil, jarName: nil, type: type, amount: amount,
            adjustmentDirection: nil, creditKind: nil, creditPaymentID: nil,
            travelEventID: nil, occurredAt: occurredAt, note: nil,
            includedInReports: includedInReports, transferID: transferID,
            createdAt: occurredAt, updatedAt: occurredAt
        )
    }

    private func reportSnapshot(monthSummaryMonth: String, insiderMonth: String) -> FinancialSnapshot {
        let month = MonthSummaryModel(
            month: monthSummaryMonth, timezone: "Asia/Ho_Chi_Minh",
            startAt: "2026-10-01T00:00:00Z", nextStartAt: "2026-11-01T00:00:00Z",
            isCurrent: true, isComplete: false, income: 0, expense: 0, net: 0,
            transactionCount: 0, categories: [], note: "", jarSummary: nil,
            calculatedAt: "2026-10-20T00:00:00Z"
        )
        let insider = InsiderSummaryModel(
            month: insiderMonth, timezone: "Asia/Ho_Chi_Minh",
            startAt: "2026-10-01T00:00:00Z", nextStartAt: "2026-11-01T00:00:00Z",
            isCurrent: true, income: 0, expense: 0, net: 0, previousExpense: 0,
            expenseDelta: 0, expenseChangeBPS: nil, averageDailyExpense: 0,
            daysConsidered: 20, spendingIncomeRatioBPS: nil, topCategories: [],
            topExpenses: [], estimated: false
        )
        return FinancialSnapshot(
            ownerID: "owner", fetchedAt: Date(timeIntervalSince1970: 1_793_000_000),
            transactions: [], wallets: [], categories: [], timezone: "Asia/Ho_Chi_Minh",
            monthSummary: month, insider: insider
        )
    }
}
