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
}
