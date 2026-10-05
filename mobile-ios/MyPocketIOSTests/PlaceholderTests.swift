import XCTest
@testable import MyPocketIOS

final class PlaceholderTests: XCTestCase {
    func testEnvironmentHasAPIBaseURL() {
        XCTAssertFalse(EnvironmentConfig.apiBaseURL.absoluteString.isEmpty)
    }
}
