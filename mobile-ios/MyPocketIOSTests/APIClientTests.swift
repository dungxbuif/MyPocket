import Foundation
import XCTest
@testable import MyPocketIOS

final class APIClientTests: XCTestCase {
    func testGoogleConfigurationRejectsPlaceholderAndAcceptsCallbackScheme() {
        XCTAssertFalse(GooglePKCEClient.isConfigured(clientID: "REPLACE_WITH_IOS_CLIENT_ID", redirectURI: "mypocket://oauth/callback"))
        XCTAssertFalse(GooglePKCEClient.isConfigured(clientID: "client-id", redirectURI: "oauth/callback"))
        XCTAssertTrue(GooglePKCEClient.isConfigured(clientID: "client-id", redirectURI: "mypocket://oauth/callback"))
    }

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

private final class NoContentTransport: HTTPTransport, @unchecked Sendable {
    func data(for request: URLRequest) async throws -> (Data, URLResponse) {
        (Data(), HTTPURLResponse(url: request.url!, statusCode: 204, httpVersion: nil, headerFields: nil)!)
    }
}

final class EmptyMutationResponseTests: XCTestCase {
    func testNoContentMutationDoesNotAttemptEnvelopeDecoding() async throws {
        let client = APIClient(
            baseURL: URL(string: "https://example.test/api/v1")!,
            transport: NoContentTransport(),
            sessionStore: TestSessionStore(session: AccessSession(token: "access", expiresAt: "2099-01-01T00:00:00Z", refreshToken: "refresh", refreshExpiresAt: "2099-02-01T00:00:00Z", user: AuthUser(id: "u", name: nil, email: nil, timezone: nil)))
        )
        let _: EmptyResponse = try await client.send(APIRequest(path: "/transactions/bulk-delete", method: .post, body: Data(#"{"transaction_ids":["tx-1"]}"#.utf8)))
    }
}

private final class RestoreSessionStore: SessionStore, @unchecked Sendable {
    var session: AccessSession?

    init(session: AccessSession?) { self.session = session }
    func load() -> AccessSession? { session }
    func save(_ session: AccessSession) throws { self.session = session }
    func clear() { session = nil }
}

private final class RestoreTransport: HTTPTransport, @unchecked Sendable {
    private(set) var refreshRequests = 0

    func data(for request: URLRequest) async throws -> (Data, URLResponse) {
        XCTAssertTrue(request.url?.path.hasSuffix("/auth/refresh") == true)
        refreshRequests += 1
        let payload = Data(#"{"data":{"token":"restored-access","expires_at":"2099-01-01T00:00:00Z","refresh_token":"restored-refresh","refresh_expires_at":"2099-02-01T00:00:00Z","user":{"id":"u-1","name":"Test","email":"test@example.com","timezone":"Asia/Ho_Chi_Minh"}}}"#.utf8)
        return (payload, HTTPURLResponse(url: request.url!, statusCode: 200, httpVersion: nil, headerFields: nil)!)
    }
}

private final class InvalidatingTransport: HTTPTransport, @unchecked Sendable {
    private(set) var requestCount = 0

    func data(for request: URLRequest) async throws -> (Data, URLResponse) {
        requestCount += 1
        return (Data(#"{"title":"Unauthorized","status":401,"detail":"expired","code":"TOKEN_INVALID"}"#.utf8), HTTPURLResponse(url: request.url!, statusCode: 401, httpVersion: nil, headerFields: nil)!)
    }
}

private struct InvalidResponse: Decodable {}

@MainActor
final class AuthRestoreTests: XCTestCase {
    private let biometricKey = "mypocket.biometric.enabled"

    override func tearDown() {
        UserDefaults.standard.removeObject(forKey: biometricKey)
        super.tearDown()
    }

    func testExpiredAccessRefreshesBeforeBiometricGate() async throws {
        UserDefaults.standard.set(true, forKey: biometricKey)
        let initial = AccessSession(token: "expired", expiresAt: "2020-01-01T00:00:00Z", refreshToken: "refresh", refreshExpiresAt: "2099-01-01T00:00:00Z", user: AuthUser(id: "u-1", name: "Test", email: nil, timezone: "Asia/Ho_Chi_Minh"))
        let store = RestoreSessionStore(session: initial)
        let transport = RestoreTransport()
        let apiClient = APIClient(baseURL: URL(string: "https://example.test/api/v1")!, transport: transport, sessionStore: store)
        let auth = AuthViewModel(store: store, apiClient: apiClient)

        await auth.restore()

        XCTAssertEqual(transport.refreshRequests, 1)
        XCTAssertEqual(store.session?.token, "restored-access")
        guard case .locked(let restored) = auth.phase else {
            return XCTFail("an enabled biometric gate should use the refreshed session")
        }
        XCTAssertEqual(restored.token, "restored-access")
    }

    func testExpiredAccessAndRefreshClearsSessionWithoutShowingBiometricGate() async {
        UserDefaults.standard.set(true, forKey: biometricKey)
        let expired = AccessSession(token: "expired", expiresAt: "2020-01-01T00:00:00Z", refreshToken: "refresh", refreshExpiresAt: "2020-01-01T00:00:00Z", user: AuthUser(id: "u-1", name: nil, email: nil, timezone: nil))
        let store = RestoreSessionStore(session: expired)
        let transport = RestoreTransport()
        let apiClient = APIClient(baseURL: URL(string: "https://example.test/api/v1")!, transport: transport, sessionStore: store)
        let auth = AuthViewModel(store: store, apiClient: apiClient)

        await auth.restore()

        XCTAssertEqual(transport.refreshRequests, 0)
        XCTAssertNil(store.load())
        XCTAssertEqual(auth.phase, .unauthenticated)
        XCTAssertFalse(auth.biometricEnabled)
    }

    func testRefreshRejectionInvalidatesAuthenticatedSession() async throws {
        let session = AccessSession(token: "access", expiresAt: "2099-01-01T00:00:00Z", refreshToken: "refresh", refreshExpiresAt: "2099-02-01T00:00:00Z", user: AuthUser(id: "u-1", name: nil, email: nil, timezone: nil))
        let store = RestoreSessionStore(session: session)
        let transport = InvalidatingTransport()
        let client = APIClient(baseURL: URL(string: "https://example.test/api/v1")!, transport: transport, sessionStore: store)
        let auth = AuthViewModel(store: store, apiClient: client)
        await auth.restore()
        guard case .authenticated = auth.phase else {
            return XCTFail("the valid access session should start authenticated")
        }

        do {
            let _: InvalidResponse = try await client.send(APIRequest(path: "/auth/profile"))
            XCTFail("an unauthorized request should not succeed")
        } catch APIError.unauthorized {
            // Expected: the refresh endpoint also returned 401.
        } catch {
            XCTFail("unexpected error: \(error)")
        }
        try await Task.sleep(for: .milliseconds(50))

        XCTAssertNil(store.load())
        XCTAssertEqual(auth.phase, .unauthenticated)
        XCTAssertFalse(auth.biometricEnabled)
    }
}
