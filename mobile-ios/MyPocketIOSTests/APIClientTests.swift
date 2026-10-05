import Foundation
import XCTest
@testable import MyPocketIOS

final class APIClientTests: XCTestCase {
    func testSessionRoundTripUsesCodableContract() throws {
        let user = AuthUser(id: "u1", name: "Test", email: "test@example.com", timezone: "Asia/Ho_Chi_Minh")
        let session = AccessSession(token: "access", expiresAt: "2026-10-05T00:00:00Z", refreshToken: "refresh", refreshExpiresAt: "2026-11-05T00:00:00Z", user: user)
        let data = try JSONEncoder.mypocket.encode(session)
        let decoded = try JSONDecoder.mypocket.decode(AccessSession.self, from: data)
        XCTAssertEqual(decoded, session)
    }

    func testProblemDetailsUsesSnakeCaseRequestID() throws {
        let data = Data(#"{"title":"Bad Request","status":400,"detail":"invalid","code":"INVALID","request_id":"req-1"}"#.utf8)
        let problem = try JSONDecoder.mypocket.decode(ProblemDetails.self, from: data)
        XCTAssertEqual(problem.requestID, "req-1")
        XCTAssertEqual(problem.code, "INVALID")
    }
}

final class RefreshCoordinatorTests: XCTestCase {
    func testTwoSequentialCallsCanUseTheSameContract() async throws {
        let coordinator = RefreshCoordinator {
            AccessSession(token: "a", expiresAt: "now", refreshToken: "r", refreshExpiresAt: "later", user: AuthUser(id: "u", name: nil, email: nil, timezone: nil))
        }
        let first = try await coordinator.refreshIfNeeded()
        let second = try await coordinator.refreshIfNeeded()
        XCTAssertEqual(first.token, second.token)
    }
}
