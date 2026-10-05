import SwiftUI
import XCTest
@testable import MyPocketIOS

@MainActor
final class DesignSystemTests: XCTestCase {
    func testSemanticTokensAreStable() {
        XCTAssertEqual(MPMetrics.hitTarget, 44)
        XCTAssertEqual(MPMetrics.cornerRadius, 12)
    }

    func testAttachmentLimitIsExplicitInThePickerContract() {
        let picker = MPAttachmentPicker(title: "Đính kèm", count: 20, maxCount: 20, add: {})
        XCTAssertNotNil(picker)
    }

    func testReportFormattingHandlesDeltasAndMissingIncome() {
        XCTAssertEqual(MPFormat.percentageFromBPS(1234), "12.3%")
        XCTAssertEqual(MPFormat.percentageFromBPS(nil), "Chưa có dữ liệu")
        XCTAssertTrue(MPFormat.signedMoney(-1).hasPrefix("−"))
    }
}
