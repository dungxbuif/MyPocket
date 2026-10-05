import XCTest

final class LaunchTests: XCTestCase {
    func testAppLaunches() throws {
        let app = XCUIApplication()
        app.launch()
        XCTAssertTrue(
            app.buttons["Đăng nhập bằng Google"].waitForExistence(timeout: 5),
            "Auth screen should expose the Google sign-in action after launch."
        )
    }
}
