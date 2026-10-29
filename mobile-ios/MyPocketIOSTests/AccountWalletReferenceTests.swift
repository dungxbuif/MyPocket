import XCTest
@testable import MyPocketIOS

@MainActor
final class AccountWalletReferenceTests: XCTestCase {
    func testAccountAndWalletRoutesKeepStableReferenceCopy() {
        XCTAssertEqual(MoreDestination.account.title, "Tài khoản")
        XCTAssertEqual(MoreDestination.wallets.title, "Ví của tôi")
        XCTAssertEqual(WalletsReferenceCopy.selectionTitle, "Chọn Ví")
        XCTAssertEqual(WalletsReferenceCopy.managementTitle, "Ví Của Tôi")
    }

    func testWalletLedgerSectionsKeepReferenceOrder() {
        XCTAssertEqual(WalletLedgerReferenceSection.stitchOrder, [
            .header, .balance, .summary, .allTransactions
        ])
        XCTAssertEqual(WalletEditorReferenceSection.stitchOrder, [
            .header, .identity, .type, .specialized, .settings, .save
        ])
    }
}
