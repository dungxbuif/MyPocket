import XCTest
@testable import MyPocketIOS

@MainActor
final class JarsMonthsReferenceTests: XCTestCase {
    func testJarsKeepStableCreateActionAndMonthReportCopy() {
        XCTAssertEqual(JarsReferenceCopy.addAction, "Thêm hũ")
        XCTAssertEqual(JarsReferenceCopy.addActionAccessibilityID, "jar-add-month")
        XCTAssertEqual(MonthsPresentation.timezoneLabel("Asia/Ho_Chi_Minh"), "theo múi giờ tài khoản")
    }

    func testMonthReportPeriodUsesAccountTimezoneAndInclusiveEnd() {
        let period = MonthsPresentation.reportPeriod(
            startAt: "2026-09-01T00:00:00Z",
            nextStartAt: "2026-10-01T00:00:00Z",
            timeZone: TimeZone(identifier: "Asia/Ho_Chi_Minh")!
        )
        XCTAssertEqual(period, "01/09/2026 – 30/09/2026")
    }
}
