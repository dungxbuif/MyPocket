import Foundation
import LocalAuthentication
import XCTest
@testable import MyPocketIOS

@MainActor
final class DeviceSignInTests: XCTestCase {
    private var preferences: UserDefaults!

    override func setUp() async throws {
        preferences = UserDefaults(suiteName: "device-sign-in-tests")!
        preferences.removePersistentDomain(forName: "device-sign-in-tests")
    }

    override func tearDown() async throws {
        preferences.removePersistentDomain(forName: "device-sign-in-tests")
    }

    private func model(session: AccessSession? = nil, device: TestDeviceStore, transport: DeviceAuthTransport) -> (AuthViewModel, DeviceTestSessionStore) {
        let sessions = DeviceTestSessionStore(session: session)
        let client = APIClient(baseURL: URL(string: "https://example.test/api/v1")!, transport: transport, sessionStore: sessions)
        return (AuthViewModel(store: sessions, apiClient: client, deviceStore: device, preferences: preferences), sessions)
    }

    func testLogoutRetainsFaceIDEnrollmentAndDeviceLoginCreatesNewSession() async throws {
        let device = TestDeviceStore()
        let transport = DeviceAuthTransport()
        let (auth, sessions) = model(session: testSession, device: device, transport: transport)
        await auth.restore()
        await auth.logout()
        XCTAssertEqual(auth.phase, .unauthenticated)
        XCTAssertNil(sessions.load())
        XCTAssertTrue(auth.canSignInWithBiometrics)
        XCTAssertNotNil(device.credential)
        await auth.signInWithBiometrics()
        guard case .authenticated(let session) = auth.phase else { return XCTFail("Face ID should create a new authenticated session after logout") }
        XCTAssertEqual(session.token, "new-device-access")
        XCTAssertEqual(sessions.load()?.token, session.token)
        let login = try XCTUnwrap(transport.requests.last)
        XCTAssertEqual(login.url?.path, "/api/v1/auth/device-login")
        XCTAssertNil(login.value(forHTTPHeaderField: "Authorization"))
        let body = try XCTUnwrap(JSONSerialization.jsonObject(with: try XCTUnwrap(login.httpBody)) as? [String: String])
        XCTAssertEqual(body["id"], "device-id")
        XCTAssertEqual(body["secret"], "mpd_synthetic-device-secret")
    }

    func testCancellingFaceIDPreservesEnrollmentAndDoesNotCallAPI() async {
        let device = TestDeviceStore()
        device.unlockError = LAError(.userCancel)
        let transport = DeviceAuthTransport()
        let (auth, _) = model(device: device, transport: transport)
        await auth.signInWithBiometrics()
        XCTAssertEqual(auth.phase, .unauthenticated)
        XCTAssertTrue(auth.canSignInWithBiometrics)
        XCTAssertTrue(transport.requests.isEmpty)
        XCTAssertNil(auth.errorMessage)
    }

    func testRevokedCredentialIsClearedButNetworkFailureKeepsFaceIDOption() async {
        for status in [401, 503] {
            let device = TestDeviceStore()
            let transport = DeviceAuthTransport()
            transport.deviceLoginStatus = status
            let (auth, _) = model(device: device, transport: transport)
            await auth.signInWithBiometrics()
            XCTAssertEqual(auth.phase, .unauthenticated)
            XCTAssertEqual(auth.canSignInWithBiometrics, status != 401)
            XCTAssertEqual(device.credential == nil, status == 401)
        }
    }

    func testExpiredSessionStillOffersIndependentDeviceSignIn() async {
        let expired = AccessSession(token: "expired", expiresAt: "2020-01-01T00:00:00Z", refreshToken: "expired-refresh", refreshExpiresAt: "2020-02-01T00:00:00Z", user: testSession.user)
        let device = TestDeviceStore()
        let transport = DeviceAuthTransport()
        let (auth, sessions) = model(session: expired, device: device, transport: transport)
        await auth.restore()
        XCTAssertEqual(auth.phase, .unauthenticated)
        XCTAssertNil(sessions.load())
        XCTAssertTrue(auth.canSignInWithBiometrics)
        XCTAssertTrue(transport.requests.isEmpty)
    }

    func testEnrollmentRequiresPasswordAndForgettingRevokesDeviceAndSession() async throws {
        let device = TestDeviceStore()
        device.clear()
        let transport = DeviceAuthTransport()
        let (auth, _) = model(session: testSession, device: device, transport: transport)
        await auth.restore()
        let enrolled = await auth.enrollBiometric(password: "confirmed-password", deviceName: "Test iPhone")
        XCTAssertTrue(enrolled)
        XCTAssertTrue(auth.canSignInWithBiometrics)
        let request = try XCTUnwrap(transport.requests.first)
        XCTAssertEqual(request.url?.path, "/api/v1/auth/devices")
        XCTAssertEqual(request.value(forHTTPHeaderField: "Authorization"), "Bearer current-access")
        let body = try XCTUnwrap(JSONSerialization.jsonObject(with: try XCTUnwrap(request.httpBody)) as? [String: String])
        XCTAssertEqual(body["password"], "confirmed-password")
        XCTAssertNotNil(body["device_id"])
        await auth.forgetDevice()
        XCTAssertNil(device.credential)
        XCTAssertFalse(auth.canSignInWithBiometrics)
        XCTAssertEqual(auth.phase, .unauthenticated)
        XCTAssertTrue(transport.requests.contains { $0.httpMethod == "DELETE" && $0.url?.path == "/api/v1/auth/devices/device-id" })
    }

    private var testSession: AccessSession {
        AccessSession(token: "current-access", expiresAt: "2099-01-01T00:00:00Z", refreshToken: "current-refresh", refreshExpiresAt: "2099-02-01T00:00:00Z", user: AuthUser(id: "owner", name: "Owner", email: "owner@example.test", timezone: "Asia/Ho_Chi_Minh"))
    }
}

private final class DeviceTestSessionStore: SessionStore, @unchecked Sendable {
    private let lock = NSRecursiveLock()
    private var stored: AccessSession?
    var session: AccessSession? {
        get { load() }
        set { lock.withLock { stored = newValue } }
    }
    init(session: AccessSession?) { stored = session }
    func load() -> AccessSession? { lock.withLock { stored } }
    func save(_ session: AccessSession) throws { lock.withLock { stored = session } }
    func clear() { lock.withLock { stored = nil } }
    func replace(_ session: AccessSession, ifCredentialsMatch original: AccessSession) throws -> Bool {
        lock.withLock {
            guard stored?.hasSameCredentials(as: original) == true, session.user.id == original.user.id else { return false }
            stored = session; return true
        }
    }
    func clear(ifCredentialsMatch original: AccessSession) -> Bool {
        lock.withLock {
            guard stored?.hasSameCredentials(as: original) == true else { return false }
            stored = nil; return true
        }
    }
}

private final class TestDeviceStore: DeviceSignInStore, @unchecked Sendable {
    var credential: DeviceSignInCredential? = .init(id: "device-id", secret: "mpd_synthetic-device-secret", expiresAt: "2099-01-01T00:00:00Z", ownerID: "owner", email: "owner@example.test")
    var unlockError: Error?
    var metadata: DeviceSignInMetadata? { credential?.metadata }
    var isAvailable: Bool { true }
    func save(_ credential: DeviceSignInCredential) throws { self.credential = credential }
    func unlock() async throws -> DeviceSignInCredential {
        if let unlockError { throw unlockError }
        return try XCTUnwrap(credential)
    }
    func clear() { credential = nil }
}

private final class DeviceAuthTransport: HTTPTransport, @unchecked Sendable {
    var requests: [URLRequest] = []
    var deviceLoginStatus = 200
    func data(for request: URLRequest) async throws -> (Data, URLResponse) {
        requests.append(request)
        let path = request.url!.path
        let status: Int
        let json: String
        if path.hasSuffix("/auth/device-login") {
            status = deviceLoginStatus
            json = #"{"data":{"token":"new-device-access","expires_at":"2099-01-01T00:00:00Z","refresh_token":"new-refresh","refresh_expires_at":"2099-02-01T00:00:00Z","user":{"id":"owner","email":"owner@example.test"}}}"#
        } else if path.hasSuffix("/auth/devices"), request.httpMethod == "POST" {
            status = 200
            json = #"{"data":{"id":"device-id","secret":"mpd_synthetic-device-secret","expires_at":"2099-01-01T00:00:00Z"}}"#
        } else {
            status = 204
            json = ""
        }
        return (Data(json.utf8), HTTPURLResponse(url: request.url!, statusCode: status, httpVersion: nil, headerFields: nil)!)
    }
}
