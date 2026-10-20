import XCTest
@testable import MyPocketIOS

@MainActor
final class AppRouteCoordinatorTests: XCTestCase {
    func testProtectedRouteWaitsForAuthenticationAndIsConsumedOnce() {
        let coordinator = AppRouteCoordinator()
        coordinator.receive(.reports)
        XCTAssertNil(coordinator.consumeIfAuthenticated(.checking))
        XCTAssertNil(coordinator.consumeIfAuthenticated(.unauthenticated))
        XCTAssertNil(coordinator.consumeIfAuthenticated(.locked(testSession)))
        XCTAssertEqual(coordinator.consumeIfAuthenticated(.authenticated(testSession)), .reports)
        XCTAssertNil(coordinator.consumeIfAuthenticated(.authenticated(testSession)))
        XCTAssertNil(coordinator.pendingRoute)
    }

    func testNewRouteReplacesPendingRouteBeforeAuthentication() {
        let coordinator = AppRouteCoordinator()
        coordinator.receive(.reports)
        coordinator.receive(.assistant)
        XCTAssertEqual(coordinator.pendingRoute, .assistant)
        XCTAssertEqual(coordinator.consumeIfAuthenticated(.authenticated(testSession)), .assistant)
    }

    func testLogoutDiscardsRouteWaitingBehindProtectedSession() {
        let coordinator = AppRouteCoordinator()
        coordinator.receive(.reports)

        coordinator.discardPendingRoute()

        XCTAssertNil(coordinator.pendingRoute)
        XCTAssertNil(coordinator.consumeIfAuthenticated(.authenticated(testSession)))
    }

    private var testSession: AccessSession {
        AccessSession(
            token: "access",
            expiresAt: "2099-01-01T00:00:00Z",
            refreshToken: "refresh",
            refreshExpiresAt: "2099-01-02T00:00:00Z",
            user: AuthUser(id: "user-1", name: "Test", email: "test@example.com", timezone: "Asia/Ho_Chi_Minh")
        )
    }
}
