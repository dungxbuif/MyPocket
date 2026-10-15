import Foundation
import XCTest
@testable import MyPocketIOS

final class AppRouteTests: XCTestCase {
    func testParsesUniversalLinkOverview() {
        let url = URL(string: "https://money.dungxbuif.com/app/overview")!
        XCTAssertEqual(AppRouteParser.parse(url), .overview)
    }

    func testParsesSchemeTransactionsAndWalletQuery() {
        let url = URL(string: "mypocket://app/transactions?wallet_id=wallet-1")!
        XCTAssertEqual(AppRouteParser.parse(url), .transactions(walletID: "wallet-1"))
    }

    func testParsesNewTransactionWithoutWallet() {
        let url = URL(string: "https://money.dungxbuif.com/app/transactions/new")!
        XCTAssertEqual(AppRouteParser.parse(url), .newTransaction(walletID: nil))
    }

    func testParsesEverySecondaryScene() {
        let cases: [(String, AppRoute)] = [
            ("categories", .categories),
            ("jars", .jars),
            ("months", .months),
            ("recurring", .recurring),
            ("travel", .travel),
            ("portfolio", .portfolio),
            ("credit", .credit),
            ("notifications", .notifications),
            ("ai-entry", .aiEntry)
        ]
        for (path, expected) in cases {
            XCTAssertEqual(
                AppRouteParser.parse(URL(string: "mypocket://app/\(path)")!),
                expected,
                "route /\(path)"
            )
        }
    }

    func testParsesWalletDetailAndRejectsEmptyIdentifier() {
        let valid = URL(string: "https://money.dungxbuif.com/app/wallets/wallet-1")!
        XCTAssertEqual(AppRouteParser.parse(valid), .wallet(id: "wallet-1"))
        let empty = URL(string: "https://money.dungxbuif.com/app/wallets/")!
        XCTAssertNil(AppRouteParser.parse(empty))
    }

    func testRejectsUnknownPathAndForeignHost() {
        XCTAssertNil(AppRouteParser.parse(URL(string: "https://money.dungxbuif.com/app/admin")!))
        XCTAssertNil(AppRouteParser.parse(URL(string: "https://evil.example/app/overview")!))
        XCTAssertNil(AppRouteParser.parse(URL(string: "mypocket://other/overview")!))
    }

    func testRejectsCredentialsAndSensitiveQueryValues() {
        XCTAssertNil(AppRouteParser.parse(URL(string: "mypocket://app/overview?token=secret")!))
        XCTAssertNil(AppRouteParser.parse(URL(string: "https://user:pass@money.dungxbuif.com/app/overview")!))
    }

    func testRejectsUnsupportedSchemeAndFragments() {
        XCTAssertNil(AppRouteParser.parse(URL(string: "http://money.dungxbuif.com/app/overview")!))
        XCTAssertNil(AppRouteParser.parse(URL(string: "mypocket://app/overview#secret")!))
    }
}
