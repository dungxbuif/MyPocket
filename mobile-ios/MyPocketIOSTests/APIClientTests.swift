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

private final class TestSessionStore: SessionStore, @unchecked Sendable {
    var session: AccessSession?
    init(session: AccessSession?) { self.session = session }
    func load() -> AccessSession? { session }
    func save(_ session: AccessSession) throws { self.session = session }
    func clear() { session = nil }
}

private final class MultipartRefreshTransport: HTTPTransport, @unchecked Sendable {
    private(set) var multipartRequests = 0
    private(set) var refreshRequests = 0

    func data(for request: URLRequest) async throws -> (Data, URLResponse) {
        let isRefresh = request.url?.path.hasSuffix("/auth/refresh") == true
        if isRefresh {
            refreshRequests += 1
            let data = Data(#"{"data":{"token":"new-access","expires_at":"2099-01-01T00:00:00Z","refresh_token":"new-refresh","refresh_expires_at":"2099-02-01T00:00:00Z","user":{"id":"u-1","name":"Test","email":"test@example.com","timezone":"Asia/Ho_Chi_Minh"}}}"#.utf8)
            return (data, HTTPURLResponse(url: request.url!, statusCode: 200, httpVersion: nil, headerFields: nil)!)
        }
        multipartRequests += 1
        if multipartRequests == 1 {
            return (Data(), HTTPURLResponse(url: request.url!, statusCode: 401, httpVersion: nil, headerFields: nil)!)
        }
        let data = Data(#"{"data":{"id":"request-1","processing":false,"proposals":[]}}"#.utf8)
        return (data, HTTPURLResponse(url: request.url!, statusCode: 200, httpVersion: nil, headerFields: nil)!)
    }
}

final class MultipartAPIClientTests: XCTestCase {
    func testMultipartSessionRefreshesAndRetriesOnce() async throws {
        let initial = AccessSession(token: "expired", expiresAt: "2020-01-01T00:00:00Z", refreshToken: "refresh", refreshExpiresAt: "2099-01-01T00:00:00Z", user: AuthUser(id: "u-1", name: "Test", email: "test@example.com", timezone: "Asia/Ho_Chi_Minh"))
        let store = TestSessionStore(session: initial)
        let transport = MultipartRefreshTransport()
        let client = APIClient(baseURL: URL(string: "https://example.test/api/v1")!, transport: transport, sessionStore: store)
        let result: EntryProcessModel = try await client.sendMultipart(path: "/ai/entry/process", fields: ["request_id": "r-1", "text": "test"], idempotencyKey: "idempotent-1")
        XCTAssertEqual(result.id, "request-1")
        XCTAssertEqual(transport.multipartRequests, 2)
        XCTAssertEqual(transport.refreshRequests, 1)
        XCTAssertEqual(store.session?.token, "new-access")
    }
}
