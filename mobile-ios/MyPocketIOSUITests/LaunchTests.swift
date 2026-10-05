import XCTest

@MainActor
final class LaunchTests: XCTestCase {
    func testAppLaunches() throws {
        let app = XCUIApplication()
        app.launch()
        XCTAssertTrue(
            app.buttons["Đăng nhập bằng Google"].waitForExistence(timeout: 5),
            "Auth screen should expose the Google sign-in action after launch."
        )
    }

    func testAuthenticatedShellExposesPrimaryRoutes() throws {
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated"]
        app.launch()
        XCTAssertTrue(app.buttons["Tổng quan"].waitForExistence(timeout: 5))
        XCTAssertTrue(app.buttons["Giao dịch"].exists)
        XCTAssertTrue(app.buttons["Ví"].exists)
        app.buttons["Giao dịch"].tap()
        XCTAssertTrue(app.staticTexts["Giao dịch"].firstMatch.waitForExistence(timeout: 5))
        app.buttons["Ví"].tap()
        XCTAssertTrue(app.staticTexts["Ví"].firstMatch.waitForExistence(timeout: 5))
    }

    func testAuthenticatedMoreRoutesAndFeedbackEntry() throws {
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated"]
        app.launch()
        XCTAssertTrue(app.buttons["Thêm"].waitForExistence(timeout: 5))
        app.buttons["Thêm"].tap()
        XCTAssertTrue(app.buttons["Báo cáo"].waitForExistence(timeout: 5))
        app.buttons["Báo cáo"].tap()
        XCTAssertTrue(app.staticTexts["Báo cáo"].firstMatch.waitForExistence(timeout: 5))
        XCTAssertTrue(app.buttons["Xem báo cáo"].waitForExistence(timeout: 5))
        XCTAssertTrue(app.buttons["Gửi phản hồi"].waitForExistence(timeout: 5))
        app.buttons["Gửi phản hồi"].firstMatch.tap()
        XCTAssertTrue(app.staticTexts["Phản hồi"].firstMatch.waitForExistence(timeout: 5))
    }

    func testAuthenticatedMoreMenuOpensEveryDestination() throws {
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated"]
        app.launch()

        let destinations = [
            "Ngân sách",
            "Báo cáo",
            "Danh mục",
            "Hũ chi tiêu",
            "Tháng",
            "Giao dịch định kỳ",
            "Travel Mode",
            "Danh mục đầu tư",
            "Sổ tín dụng",
            "Finance Assistant",
            "Nhập giao dịch bằng AI",
            "Phản hồi & changelog",
            "Tài khoản"
        ]

        for (index, title) in destinations.enumerated() {
            if index > 0 {
                app.terminate()
                app.launchArguments = ["-UITestAuthenticated"]
                app.launch()
            }
            XCTAssertTrue(app.buttons["Thêm"].waitForExistence(timeout: 5))
            app.buttons["Thêm"].tap()

            let entry = app.buttons[title].firstMatch
            XCTAssertTrue(entry.waitForExistence(timeout: 5), "Missing More destination: \(title)")
            for _ in 0..<5 where !entry.isHittable {
                app.swipeUp()
            }
            XCTAssertTrue(entry.isHittable, "More destination is not reachable: \(title)")
            entry.tap()
            XCTAssertTrue(app.staticTexts[title].firstMatch.waitForExistence(timeout: 5), "Destination did not render: \(title)")

            let back = app.buttons["Quay lại"].firstMatch
            guard back.waitForExistence(timeout: 5) else {
                XCTFail("Missing back action for: \(title)")
                return
            }
            back.tap()
            XCTAssertTrue(app.buttons["Thêm"].waitForExistence(timeout: 5), "Could not return from: \(title)")
        }
    }
}
