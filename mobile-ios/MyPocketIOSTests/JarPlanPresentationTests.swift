import XCTest
@testable import MyPocketIOS

final class JarPlanPresentationTests: XCTestCase {
    private func item(remaining: Int64?, allocation: Int64?, spent: Int64, variance: Int64? = nil, cumulativeSpent: Int64? = nil) -> JarItemModel {
        JarItemModel(jarID: "j", name: "Du lịch", allocationMode: allocation == nil ? "none" : "fixed", allocationAmount: allocation,
                     allocationPercent: nil, calculatedAllocation: allocation, spent: spent, active: true,
                     cumulativeAllocated: nil, cumulativeSpent: cumulativeSpent, remainingPlan: remaining, allocationMonths: 2,
                     monthsInRange: 2, firstConfigMonth: "2026-10", monthVariance: variance)
    }

    // Owner example: 15M monthly, 10M spent in October, nothing yet in November -> 20M can still be planned.
    func testOwnerExampleShowsCarryOverAndRemainingPlan() {
        let november = item(remaining: 20_000_000, allocation: 15_000_000, spent: 0, variance: 15_000_000)
        XCTAssertEqual(JarPlanPresentation.headlineTitle(november), "Còn có thể chi")
        XCTAssertEqual(JarPlanPresentation.headlineAmount(november), 20_000_000)
        XCTAssertEqual(JarPlanPresentation.carriedOver(november), 5_000_000)
        let note = JarPlanPresentation.note(november, masked: false)
        XCTAssertTrue(note.contains("Kế hoạch tháng này"), note)
        XCTAssertTrue(note.contains("Dư từ tháng trước: +"), note)
        XCTAssertTrue(note.contains("Còn có thể chi theo kế hoạch"), note)
        XCTAssertFalse(note.lowercased().contains("số dư hũ"))
    }

    func testOverspendingKeepsMonthWarningAndNoNegativeCarryOver() {
        let overspent = item(remaining: 3_000_000, allocation: 15_000_000, spent: 17_000_000, variance: -2_000_000)
        XCTAssertEqual(JarPlanPresentation.carriedOver(overspent), 5_000_000)
        XCTAssertTrue(JarPlanPresentation.isOverMonth(overspent))
    }

    func testOverrunIsShownAsAnOverrunNotANegativeBalance() {
        let over = item(remaining: -1_228_000, allocation: 5_500_000, spent: 1_264_000)
        XCTAssertEqual(JarPlanPresentation.headlineTitle(over), "Vượt kế hoạch")
        XCTAssertEqual(JarPlanPresentation.headlineAmount(over), 1_228_000)
    }

    func testJarWithoutAllocationShowsCumulativeSpendAndNoFakeWarning() {
        let none = item(remaining: nil, allocation: nil, spent: 400_000, cumulativeSpent: 900_000)
        XCTAssertEqual(JarPlanPresentation.headlineTitle(none), "Đã chi cộng dồn")
        XCTAssertEqual(JarPlanPresentation.headlineAmount(none), 900_000)
        XCTAssertNil(JarPlanPresentation.carriedOver(none))
        XCTAssertFalse(JarPlanPresentation.isOverMonth(none))
        XCTAssertTrue(JarPlanPresentation.monthLine(none, masked: false).hasPrefix("Chưa đặt mức"))
    }

    func testReconciliationWarnsOnlyWhenPlanExceedsWallets() {
        func summary(plan: Int64, wallets: Int64, exceeds: Int64) -> JarMonthSummaryModel {
            JarMonthSummaryModel(month: "2026-11", timezone: "Asia/Ho_Chi_Minh", actualIncome: 0, totalSpent: 0, unassignedSpent: 0,
                                 totalAllocated: 0, totalAllocationPercent: 0, overIncome: false, overOneHundredPercent: false,
                                 jars: [], items: [item(remaining: plan, allocation: 1, spent: 0)],
                                 remainingPlanTotal: plan, unassignedCumulative: nil, walletsInTotalBalance: wallets, planExceedsWalletsBy: exceeds)
        }
        let ok = JarPlanPresentation.reconciliation(summary(plan: 20_000_000, wallets: 30_000_000, exceeds: 0), masked: false)
        XCTAssertEqual(ok?.exceeds, false)
        let warn = JarPlanPresentation.reconciliation(summary(plan: 20_000_000, wallets: 2_000_000, exceeds: 18_000_000), masked: false)
        XCTAssertEqual(warn?.exceeds, true)
        XCTAssertTrue(warn?.text.contains("không chặn") == true)
    }

    func testOlderPayloadsWithoutPlanFieldsStillDecode() throws {
        let json = #"{"month":"2026-11","timezone":"Asia/Ho_Chi_Minh","actual_income":0,"total_spent":0,"unassigned_spent":0,"total_allocated":0,"total_allocation_percent":0,"over_income":false,"over_one_hundred_percent":false,"jars":[],"items":[{"jar_id":"j","name":"A","allocation_mode":"none","spent":1,"active":true}]}"#
        let decoded = try JSONDecoder.mypocket.decode(JarMonthSummaryModel.self, from: Data(json.utf8))
        XCTAssertNil(decoded.items.first?.remainingPlan)
        XCTAssertNil(decoded.remainingPlanTotal)
    }

    // MARK: Ledger filter, row chip, editor hint, overview rows

    private func tx(type: String = "expense", jarID: String? = nil, jarName: String? = nil, transferID: String? = nil, amount: Int64 = 1_000) -> TransactionModel {
        TransactionModel(id: UUID().uuidString, ownerID: "o", walletID: "w", categoryID: nil, jarID: jarID, jarName: jarName, type: type,
                         amount: amount, adjustmentDirection: nil, creditKind: nil, creditPaymentID: nil, travelEventID: nil,
                         occurredAt: "2026-11-02T03:00:00Z", note: nil, includedInReports: true, transferID: transferID,
                         createdAt: "", updatedAt: "")
    }

    private func summary(_ items: [JarItemModel], unassigned: Int64 = 0) -> JarMonthSummaryModel {
        JarMonthSummaryModel(month: "2026-11", timezone: "Asia/Ho_Chi_Minh", actualIncome: 0, totalSpent: 0, unassignedSpent: unassigned,
                             totalAllocated: 0, totalAllocationPercent: 0, overIncome: false, overOneHundredPercent: false,
                             jars: items.map { JarIdentityModel(jarID: $0.jarID, name: $0.name) }, items: items,
                             remainingPlanTotal: items.compactMap(\.remainingPlan).reduce(0, +), unassignedCumulative: nil,
                             walletsInTotalBalance: nil, planExceedsWalletsBy: nil)
    }

    func testLedgerFilterMatchesJarAndUnassignedOrdinaryExpensesOnly() {
        let withJar = tx(jarID: "j1", jarName: "Ăn uống")
        let noJar = tx()
        let income = tx(type: "income")
        let transfer = tx(transferID: "t1")
        XCTAssertTrue(JarLedgerFilter.matches(withJar, selection: "j1"))
        XCTAssertFalse(JarLedgerFilter.matches(noJar, selection: "j1"))
        XCTAssertTrue(JarLedgerFilter.matches(noJar, selection: JarLedgerFilter.unassigned))
        XCTAssertFalse(JarLedgerFilter.matches(withJar, selection: JarLedgerFilter.unassigned))
        XCTAssertFalse(JarLedgerFilter.matches(income, selection: JarLedgerFilter.unassigned))
        XCTAssertFalse(JarLedgerFilter.matches(transfer, selection: JarLedgerFilter.unassigned))
        XCTAssertTrue(JarLedgerFilter.matches(income, selection: JarLedgerFilter.all))
    }

    func testLedgerOptionsListMonthJarsThenUsedJarsAndUnassigned() {
        let options = JarLedgerFilter.options(summary: summary([item(remaining: 1, allocation: 1, spent: 0)]),
                                              transactions: [tx(jarID: "old", jarName: "Cũ"), tx(jarID: "j", jarName: "Du lịch")])
        XCTAssertEqual(options.map(\.label), ["Tất cả", "Du lịch", "Cũ", "Chưa gắn hũ"])
    }

    func testJarChipOnlyOnOrdinaryExpenses() {
        XCTAssertEqual(JarLedgerFilter.chip(tx(jarID: "j", jarName: "Ăn uống")), "Ăn uống")
        XCTAssertNil(JarLedgerFilter.chip(tx(type: "income", jarID: "j", jarName: "Ăn uống")))
        XCTAssertNil(JarLedgerFilter.chip(tx()))
    }

    @MainActor func testApplyFiltersCombinesWalletPeriodAndJar() {
        let rows = [tx(jarID: "j1", jarName: "A"), tx(), tx(jarID: "j2", jarName: "B")]
        let result = TransactionsViewModel.applyFilters(rows, walletSelection: "__all__", periodSelection: "all", customStart: Date(),
                                                        customEnd: Date(), accountTimezone: "Asia/Ho_Chi_Minh", jarSelection: "j2")
        XCTAssertEqual(result.count, 1)
        XCTAssertEqual(result.first?.jarID, "j2")
    }

    func testEditorHintShowsRemainingPlanThenOverrunWithoutBlocking() {
        let jar = item(remaining: 20_000_000, allocation: 15_000_000, spent: 0)
        let s = summary([jar])
        XCTAssertEqual(JarPlanPresentation.editorHint(summary: s, jarID: "j", amount: 0)?.exceeds, false)
        let fine = JarPlanPresentation.editorHint(summary: s, jarID: "j", amount: 5_000_000)
        XCTAssertTrue(fine?.text.hasPrefix("Hũ Du lịch còn có thể chi 15.000.000") == true, fine?.text ?? "")
        let over = JarPlanPresentation.editorHint(summary: s, jarID: "j", amount: 17_000_000)
        XCTAssertEqual(over?.exceeds, true)
        XCTAssertTrue(over?.text.contains("sẽ vượt mức tháng 2.000.000") == true, over?.text ?? "")
        XCTAssertNil(JarPlanPresentation.editorHint(summary: s, jarID: "missing", amount: 1))
    }

    func testEditorHintDoesNotDoubleCountTheEditedTransaction() {
        let jar = item(remaining: 5_000_000, allocation: 15_000_000, spent: 10_000_000)
        let hint = JarPlanPresentation.editorHint(summary: summary([jar]), jarID: "j", amount: 10_000_000, existingAmount: 10_000_000)
        XCTAssertEqual(hint?.exceeds, false)
        XCTAssertTrue(hint?.text.contains("5.000.000") == true, hint?.text ?? "")
    }

    @MainActor func testOverviewJarRowsShowPlanTotalAndUnassigned() {
        let s = summary([item(remaining: 20_000_000, allocation: 15_000_000, spent: 0)], unassigned: 3_000_000)
        let rows = OverviewViewModel.jarRows(s, masked: false)
        XCTAssertEqual(rows.map(\.title), ["Hũ: còn có thể chi tổng", "Chưa phân bổ tháng này"])
        XCTAssertTrue(rows[0].value.contains("20.000.000"))
        XCTAssertTrue(rows[1].value.contains("3.000.000"))
        XCTAssertTrue(OverviewViewModel.jarRows(nil, masked: false).isEmpty)
    }
}
