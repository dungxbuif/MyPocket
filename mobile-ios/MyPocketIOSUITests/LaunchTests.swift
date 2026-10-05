import XCTest

final class LaunchTests: XCTestCase {
    func testAppLaunches() throws {
        let app = XCUIApplication()
        app.launch()
        XCTAssertTrue(app.otherElements["app-root"].waitForExistence(timeout: 5))
    }
}
