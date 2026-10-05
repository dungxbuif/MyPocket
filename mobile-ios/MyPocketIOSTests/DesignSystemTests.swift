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

    func testSelectionLabelsPreserveUnknownValuesAndLocalizeEnums() {
        XCTAssertEqual(MPSelectionLabels.display(for: "expense"), "Chi")
        XCTAssertEqual(MPSelectionLabels.display(for: "percent"), "Phần trăm thu nhập")
        XCTAssertEqual(MPSelectionLabels.display(for: "unknown-provider-value"), "unknown-provider-value")
    }

    func testJarAllocationRejectsZeroPercentAndStaleModeValues() {
        XCTAssertTrue(JarAllocationDraftValidator.isValid(mode: "none", amount: nil, percent: nil))
        XCTAssertTrue(JarAllocationDraftValidator.isValid(mode: "fixed", amount: 1, percent: nil))
        XCTAssertTrue(JarAllocationDraftValidator.isValid(mode: "percent", amount: nil, percent: 100))
        XCTAssertFalse(JarAllocationDraftValidator.isValid(mode: "percent", amount: nil, percent: 0))
        XCTAssertTrue(JarAllocationDraftValidator.isValid(mode: "percent", amount: nil, percent: 12.34))
        XCTAssertFalse(JarAllocationDraftValidator.isValid(mode: "percent", amount: nil, percent: 12.345))
        XCTAssertFalse(JarAllocationDraftValidator.isValid(mode: "percent", amount: nil, percent: .infinity))
        XCTAssertFalse(JarAllocationDraftValidator.isValid(mode: "fixed", amount: 100, percent: 10))
        XCTAssertTrue(JarAllocationDraftValidator.isValid(mode: "fixed", amount: JarAllocationDraftValidator.maximumFixedAmount, percent: nil))
        XCTAssertFalse(JarAllocationDraftValidator.isValid(mode: "fixed", amount: JarAllocationDraftValidator.maximumFixedAmount + 1, percent: nil))
        XCTAssertFalse(JarAllocationDraftValidator.isValid(mode: "none", amount: 100, percent: nil))
    }
}
