import Foundation
import XCTest
import UIKit
@testable import MyPocketIOS

final class APIClientTests: XCTestCase {
#if DEBUG
    func testInMemorySessionStoreRoundTripAndClear() throws {
        let store = InMemorySessionStore()
        XCTAssertNil(store.load())

        let session = AccessSession(
            token: "fixture-access",
            expiresAt: "2099-01-01T00:00:00Z",
            refreshToken: "fixture-refresh",
            refreshExpiresAt: "2099-02-01T00:00:00Z",
            user: AuthUser(id: "fixture-user", name: nil, email: nil, timezone: nil)
        )
        try store.save(session)
        XCTAssertEqual(store.load(), session)
        store.clear()
        XCTAssertNil(store.load())
    }
#endif

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

    func testQueryInAPIRequestPathRemainsAQuery() async throws {
        let transport = URLCaptureTransport()
        let client = APIClient(
            baseURL: URL(string: "https://example.test/api/v1")!,
            transport: transport,
            sessionStore: TestSessionStore(session: AccessSession(token: "access", expiresAt: "2099-01-01T00:00:00Z", refreshToken: "refresh", refreshExpiresAt: "2099-02-01T00:00:00Z", user: AuthUser(id: "u", name: nil, email: nil, timezone: nil)))
        )

        let _: EmptyResponse = try await client.send(APIRequest(path: "/credit/wallets/w-1/statement?from=2026-10-01&to=2026-11-01"))

        XCTAssertEqual(transport.url?.path, "/api/v1/credit/wallets/w-1/statement")
        XCTAssertEqual(transport.url?.query, "from=2026-10-01&to=2026-11-01")
        XCTAssertNotNil(transport.requestID)
        XCTAssertFalse(transport.requestID?.isEmpty ?? true)
    }

    func testFeedbackScreenshotUsesBackendPNGContract() async throws {
        let transport = FeedbackMultipartTransport()
        let client = APIClient(
            baseURL: URL(string: "https://example.test/api/v1")!,
            transport: transport,
            sessionStore: TestSessionStore(session: AccessSession(token: "access", expiresAt: "2099-01-01T00:00:00Z", refreshToken: "refresh", refreshExpiresAt: "2099-02-01T00:00:00Z", user: AuthUser(id: "u", name: nil, email: nil, timezone: nil)))
        )
        let api = FinanceAPI(client: client)

        _ = try await api.createFeedback(
            FeedbackInputModel(type: "bug", title: "Ảnh lỗi", description: "Không tải được ảnh"),
            screenshot: Data("fixture-image".utf8)
        )

        let body = String(decoding: transport.request?.httpBody ?? Data(), as: UTF8.self)
        XCTAssertTrue(body.contains("filename=\"screen.png\""))
        XCTAssertTrue(body.contains("Content-Type: image/png"))
    }

    @MainActor
    func testScreenSnapshotEncodesPNGBytes() {
        let image = UIGraphicsImageRenderer(size: CGSize(width: 1, height: 1)).image { _ in }
        let data = ScreenSnapshot.pngData(for: image)
        XCTAssertEqual(data?.prefix(8), Data([137, 80, 78, 71, 13, 10, 26, 10]))
    }

    @MainActor
    func testScreenSnapshotKeepsComplexCaptureUnderFeedbackLimit() {
        let width = 960
        let height = 1_600
        var bytes = Data(count: width * height * 4)
        bytes.withUnsafeMutableBytes { rawBuffer in
            var state: UInt32 = 0x9E3779B9
            for index in 0..<rawBuffer.count {
                state = state &* 1_664_525 &+ 1_013_904_223
                rawBuffer[index] = UInt8(truncatingIfNeeded: state >> 24)
            }
        }
        let provider = CGDataProvider(data: bytes as CFData)!
        let image = UIImage(cgImage: CGImage(
            width: width,
            height: height,
            bitsPerComponent: 8,
            bitsPerPixel: 32,
            bytesPerRow: width * 4,
            space: CGColorSpaceCreateDeviceRGB(),
            bitmapInfo: CGBitmapInfo(rawValue: CGImageAlphaInfo.premultipliedLast.rawValue),
            provider: provider,
            decode: nil,
            shouldInterpolate: false,
            intent: .defaultIntent
        )!)

        let data = ScreenSnapshot.pngData(for: image)
        XCTAssertNotNil(data)
        XCTAssertLessThanOrEqual(data?.count ?? Int.max, 1 << 20)
    }
}

private final class URLCaptureTransport: HTTPTransport, @unchecked Sendable {
    private(set) var url: URL?
    private(set) var requestID: String?

    func data(for request: URLRequest) async throws -> (Data, URLResponse) {
        url = request.url
        requestID = request.value(forHTTPHeaderField: "X-Request-ID")
        return (Data(#"{"data":{}}"#.utf8), HTTPURLResponse(url: request.url!, statusCode: 200, httpVersion: nil, headerFields: nil)!)
    }
}

private final class FeedbackMultipartTransport: HTTPTransport, @unchecked Sendable {
    private(set) var request: URLRequest?

    func data(for request: URLRequest) async throws -> (Data, URLResponse) {
        self.request = request
        let response = #"{"data":{"id":"feedback-1","type":"bug","title":"Ảnh lỗi","description":"Không tải được ảnh","status":"open","fixed_at":null,"changelog_id":null,"created_at":"2026-10-06T00:00:00Z","updated_at":"2026-10-06T00:00:00Z","screenshot_available":true},"meta":null}"#
        return (Data(response.utf8), HTTPURLResponse(url: request.url!, statusCode: 201, httpVersion: nil, headerFields: nil)!)
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
    private(set) var multipartRequestIDs: [String] = []

    func data(for request: URLRequest) async throws -> (Data, URLResponse) {
        let isRefresh = request.url?.path.hasSuffix("/auth/refresh") == true
        if isRefresh {
            refreshRequests += 1
            let data = Data(#"{"data":{"token":"new-access","expires_at":"2099-01-01T00:00:00Z","refresh_token":"new-refresh","refresh_expires_at":"2099-02-01T00:00:00Z","user":{"id":"u-1","name":"Test","email":"test@example.com","timezone":"Asia/Ho_Chi_Minh"}}}"#.utf8)
            return (data, HTTPURLResponse(url: request.url!, statusCode: 200, httpVersion: nil, headerFields: nil)!)
        }
        multipartRequests += 1
        multipartRequestIDs.append(request.value(forHTTPHeaderField: "X-Request-ID") ?? "")
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
        XCTAssertEqual(transport.multipartRequestIDs.count, 2)
        XCTAssertFalse(transport.multipartRequestIDs[0].isEmpty)
        XCTAssertEqual(transport.multipartRequestIDs[0], transport.multipartRequestIDs[1])
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
