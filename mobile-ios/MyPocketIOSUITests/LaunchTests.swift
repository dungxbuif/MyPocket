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

    func testTransactionOverflowMenuKeepsTransferAndAdjustmentActions() throws {
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated"]
        app.launch()
        XCTAssertTrue(app.buttons["Giao dịch"].waitForExistence(timeout: 5))
        app.buttons["Giao dịch"].tap()
        let menu = app.buttons["Tùy chọn giao dịch"]
        XCTAssertTrue(menu.waitForExistence(timeout: 5))
        menu.tap()
        XCTAssertTrue(app.buttons["Chuyển tiền đến ví khác"].waitForExistence(timeout: 5))
        XCTAssertTrue(app.buttons["Điều chỉnh số dư"].exists)
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

    func testAPIBackedAssistantLoadsOverview() throws {
#if !LOCAL_API_UI
        throw XCTSkip("Build with SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI and run against the disposable local API fixture.")
#else
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        XCTAssertTrue(app.buttons["Thêm"].waitForExistence(timeout: 10))
        app.buttons["Thêm"].tap()
        XCTAssertTrue(app.buttons["Finance Assistant"].waitForExistence(timeout: 10))
        app.buttons["Finance Assistant"].tap()
        XCTAssertTrue(app.staticTexts["Finance Assistant"].firstMatch.waitForExistence(timeout: 10))
        XCTAssertTrue(app.staticTexts["Finance Assistant chưa được bật."].waitForExistence(timeout: 10))
        XCTAssertFalse(app.staticTexts["Không thể tải tổng quan trợ lý tài chính."].exists)
#endif
    }

    func testAPIBackedWalletDetailExposesTransactionEditor() throws {
#if !LOCAL_API_UI
        throw XCTSkip("Build with SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI and seed the disposable fixture wallet.")
#else
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        XCTAssertTrue(app.buttons["Ví"].waitForExistence(timeout: 10))
        app.buttons["Ví"].tap()
        XCTAssertTrue(app.staticTexts["Fixture Wallet"].waitForExistence(timeout: 10))
        app.staticTexts["Fixture Wallet"].tap()
        XCTAssertTrue(app.buttons["Thêm giao dịch"].waitForExistence(timeout: 10))
        app.buttons["Thêm giao dịch"].tap()
        XCTAssertTrue(app.staticTexts["Thêm giao dịch"].waitForExistence(timeout: 10))
        XCTAssertTrue(app.staticTexts["Danh mục"].exists)
        XCTAssertTrue(app.staticTexts["Ghi chú"].exists)
#endif
    }

    func testAPIBackedCreditWalletOpensLedgerActions() throws {
#if !LOCAL_API_UI
        throw XCTSkip("Build with SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI and seed the disposable credit wallet.")
#else
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        XCTAssertTrue(app.buttons["Ví"].waitForExistence(timeout: 10))
        app.buttons["Ví"].tap()
        XCTAssertTrue(app.staticTexts["Fixture Credit"].waitForExistence(timeout: 10))
        app.staticTexts["Fixture Credit"].tap()
        XCTAssertTrue(app.staticTexts["Sổ tín dụng"].waitForExistence(timeout: 10))
        let addEntry = app.buttons["Thêm giao dịch"]
        if !addEntry.waitForExistence(timeout: 30) {
            print(app.debugDescription)
            XCTFail("Credit ledger actions did not render")
            return
        }
        XCTAssertTrue(app.buttons["Thanh toán"].waitForExistence(timeout: 5))
#endif
    }
}
