import XCTest
@testable import MyPocketIOS

@MainActor
final class OverviewReferenceTests: XCTestCase {
    func testOverviewBalanceUsesDedicatedReferenceHierarchyToken() {
        XCTAssertEqual(MPMetrics.balanceTextSize, 34)
    }

    func testOverviewMonthSummaryUsesCompactMetricLabels() {
        let items = [
            MPMetricItem(title: "Thu", value: MPFormat.money(32_500_000)),
            MPMetricItem(title: "Chi", value: MPFormat.money(18_700_000))
        ]

        XCTAssertEqual(items.map(\.title), ["Thu", "Chi"])
        XCTAssertEqual(items.map(\.value), ["32.500.000\u{00a0}₫", "18.700.000\u{00a0}₫"])
    }

    func testOverviewWalletSubtitleKeepsPopulatedDescriptionAndTypeFallback() {
        XCTAssertEqual(
            OverviewViewModel.walletSubtitle(description: "  Tài khoản chính  ", type: "basic"),
            "Tài khoản chính"
        )
        XCTAssertEqual(
            OverviewViewModel.walletSubtitle(description: " ", type: "goal"),
            "Tiết kiệm"
        )
    }
}
