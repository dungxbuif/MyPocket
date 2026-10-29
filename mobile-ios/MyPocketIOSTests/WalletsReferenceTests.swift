import XCTest
@testable import MyPocketIOS

@MainActor
final class WalletsReferenceTests: XCTestCase {
    func testWalletSelectorAndManagementTitlesMatchReferenceStates() {
        XCTAssertEqual(WalletsReferenceCopy.title(isEditing: false), "Chọn Ví")
        XCTAssertEqual(WalletsReferenceCopy.title(isEditing: true), "Ví Của Tôi")
    }

    func testPopulatedWalletManagementKeepsAggregateAndIncludedGroupsSeparate() {
        let wallets = [
            WalletModel(id: "cash", name: "Tiền mặt", type: "basic", currency: "VND", openingBalance: 0,
                        currentBalance: 2_300_000, availableCredit: nil, isInTotal: true, description: nil,
                        targetAmount: nil, targetDate: nil, creditLimit: nil, lastStatementBalance: nil,
                        statementDay: nil, paymentDueDay: nil),
            WalletModel(id: "trip", name: "Du lịch", type: "goal", currency: "VND", openingBalance: 0,
                        currentBalance: 1_000_000, availableCredit: nil, isInTotal: false, description: nil,
                        targetAmount: 10_000_000, targetDate: "2027-01-01", creditLimit: nil,
                        lastStatementBalance: nil, statementDay: nil, paymentDueDay: nil)
        ]

        let sections = MPWalletManagementPolicy.sections(from: wallets)
        XCTAssertEqual(sections.map(\.id), ["included", "excluded"])
        XCTAssertEqual(MPWalletManagementPolicy.totalBalance(from: wallets), 2_300_000)
        XCTAssertEqual(MPWalletSelectionList.balanceLabel(2_300_000, masked: false), "2.300.000\u{00a0}₫")
    }

    func testWalletDetailScopesEachSideOfPairedTransferWithoutChangingAccountingFlags() {
        let source = transaction(
            id: "source", walletID: "cash", type: "expense", amount: 700_000,
            includedInReports: false, transferID: "pair-1"
        )
        let destination = transaction(
            id: "destination", walletID: "savings", type: "income", amount: 700_000,
            includedInReports: false, transferID: "pair-1"
        )

        let cashRows = WalletDetailTransactionPolicy.rows([source, destination], walletID: "cash")
        let savingsRows = WalletDetailTransactionPolicy.rows([source, destination], walletID: "savings")

        XCTAssertEqual(cashRows.map(\.id), ["source"])
        XCTAssertEqual(savingsRows.map(\.id), ["destination"])
        XCTAssertEqual(cashRows.first?.transferID, "pair-1")
        XCTAssertEqual(savingsRows.first?.transferID, "pair-1")
        XCTAssertEqual(cashRows.first?.includedInReports, false)
        XCTAssertEqual(savingsRows.first?.includedInReports, false)
        XCTAssertEqual(WalletDetailTransactionPolicy.signedAmount(source), -700_000)
        XCTAssertEqual(WalletDetailTransactionPolicy.signedAmount(destination), 700_000)
    }

    func testOfflineTransferRecoveryDoesNotDependOnCategoryHydration() {
        let transfer = OutboxPayload.transfer(TransferInputModel(
            sourceWalletID: "cash", destinationWalletID: "savings", amount: 100_000,
            occurredAt: "2026-10-26T10:00:00Z", note: nil
        ))

        XCTAssertTrue(WalletQueueEditPolicy.canEdit(transfer, walletType: "goal", detailsReady: false))
        XCTAssertFalse(WalletQueueEditPolicy.canEdit(transfer, walletType: "credit", detailsReady: true))
    }

    func testGoalWalletSelectorResolvesOnlyAnAvailableGoalWallet() {
        let basic = wallet(id: "cash", type: "basic")
        let firstGoal = wallet(id: "rainy-day", type: "goal")
        let secondGoal = wallet(id: "home", type: "goal")

        XCTAssertEqual(
            WalletSelectionPolicy.goalWalletID(for: "home", wallets: [basic, firstGoal, secondGoal]),
            "home"
        )
        XCTAssertNil(WalletSelectionPolicy.goalWalletID(for: "__all__", wallets: [firstGoal, secondGoal]))
        XCTAssertNil(WalletSelectionPolicy.goalWalletID(for: "cash", wallets: [basic, firstGoal]))
        XCTAssertNil(WalletSelectionPolicy.goalWalletID(for: "missing", wallets: [firstGoal]))
    }

    func testWalletDetailAndLedgerUseTheCanonicalTransactionSignPolicy() {
        let income = transaction(id: "income", walletID: "cash", type: "income", amount: 80_000, includedInReports: true, transferID: nil)
        let expense = transaction(id: "expense", walletID: "cash", type: "expense", amount: 30_000, includedInReports: true, transferID: nil)
        var decrease = transaction(id: "adjustment", walletID: "cash", type: "adjustment", amount: 10_000, includedInReports: false, transferID: nil)
        decrease.adjustmentDirection = "decrease"

        XCTAssertEqual(MPTransactionPresentation.signedAmount(for: income), 80_000)
        XCTAssertEqual(MPTransactionPresentation.signedAmount(for: expense), -30_000)
        XCTAssertEqual(MPTransactionPresentation.signedAmount(for: decrease), -10_000)
        XCTAssertEqual(WalletDetailTransactionPolicy.signedAmount(expense), MPTransactionPresentation.signedAmount(for: expense))
        XCTAssertEqual(TransactionsPresentation.signedAmount(for: decrease), MPTransactionPresentation.signedAmount(for: decrease))
    }

    private func wallet(id: String, type: String) -> WalletModel {
        WalletModel(
            id: id, name: id, type: type, currency: "VND", openingBalance: 0,
            currentBalance: 0, availableCredit: nil, isInTotal: true,
            description: nil, targetAmount: nil, targetDate: nil,
            creditLimit: nil, lastStatementBalance: nil, statementDay: nil,
            paymentDueDay: nil
        )
    }

    private func transaction(
        id: String,
        walletID: String,
        type: String,
        amount: Int64,
        includedInReports: Bool,
        transferID: String?
    ) -> TransactionModel {
        TransactionModel(
            id: id, ownerID: "owner", walletID: walletID, categoryID: nil, jarID: nil,
            jarName: nil, type: type, amount: amount, adjustmentDirection: nil,
            creditKind: nil, creditPaymentID: nil, travelEventID: nil,
            occurredAt: "2026-10-26T10:00:00Z", note: nil,
            includedInReports: includedInReports, transferID: transferID,
            createdAt: "2026-10-26T10:00:00Z", updatedAt: "2026-10-26T10:00:00Z"
        )
    }
}
