import SwiftUI
import XCTest
@testable import MyPocketIOS

final class DesignSystemTests: XCTestCase {
    func testSemanticTokensAreStable() {
        XCTAssertEqual(MPMetrics.hitTarget, 44)
        XCTAssertEqual(MPMetrics.cornerRadius, 12)
    }

    func testAttachmentLimitIsExplicitInThePickerContract() {
        let picker = MPAttachmentPicker(title: "Đính kèm", count: 20, maxCount: 20, add: {})
        XCTAssertNotNil(picker)
    }
}
