import Foundation
import LocalAuthentication
import Security
import SwiftUI

#if DEBUG
struct LocalAuthFixtureExchangeRequest: Codable, Sendable {
    let code: String
    let codeVerifier: String
    let redirectURI: String
    let platform: String

    enum CodingKeys: String, CodingKey {
        case code
        case codeVerifier = "code_verifier"
        case redirectURI = "redirect_uri"
        case platform
    }
}
#endif

enum AuthPhase: Equatable {
    case checking
    case unauthenticated
    case locked(AccessSession)
    case authenticated(AccessSession)
}

@MainActor
final class AuthViewModel: ObservableObject {
    @Published private(set) var phase: AuthPhase = .checking
    @Published var errorMessage: String?
    @Published var isBusy = false
    @Published var biometricEnabled = false
    @Published private(set) var rememberedDevice: DeviceSignInMetadata?

    private let store: SessionStore
    private let deviceStore: DeviceSignInStore
    private let preferences: UserDefaults
    let financeAPI: FinanceAPI
    let financialCache: FinancialSnapshotCache?
    let financialQueue: FinancialMutationQueue?
    private let biometricPreferenceKey = "mypocket.biometric.enabled"

    init(store: SessionStore? = nil, apiClient: APIClient? = nil, deviceStore: DeviceSignInStore? = nil, preferences: UserDefaults = .standard, financialCache: FinancialSnapshotCache? = nil, financialQueue: FinancialMutationQueue? = nil) {
        let resolvedStore = store ?? Self.makeDefaultStore()
        self.store = resolvedStore
        let resolvedDeviceStore = deviceStore ?? KeychainDeviceSignInStore(defaults: preferences)
        self.deviceStore = resolvedDeviceStore
        self.preferences = preferences
        self.rememberedDevice = resolvedDeviceStore.metadata
        let resolvedAPI = apiClient ?? APIClient(sessionStore: resolvedStore)
        self.financeAPI = FinanceAPI(client: resolvedAPI)
        // Custom transports/environments must explicitly supply their matching cache.
        self.financialCache = financialCache ?? (apiClient == nil ? try? FinancialSnapshotCache.makeDefault() : nil)
        self.financialQueue = financialQueue ?? (apiClient == nil ? try? FinancialMutationQueue.makeDefault(api: self.financeAPI, sessions: resolvedStore) : nil)
        self.biometricEnabled = preferences.bool(forKey: biometricPreferenceKey) || resolvedDeviceStore.metadata?.isValid == true
#if DEBUG
        let isUITestAuthenticated = ProcessInfo.processInfo.arguments.contains("-UITestAuthenticated")
#else
        let isUITestAuthenticated = false
#endif
        if !isUITestAuthenticated {
            resolvedAPI.onSessionInvalidated = { [weak self] in
                Task { @MainActor [weak self] in
                    self?.expireSession()
                }
            }
        }
    }

    private static func makeDefaultStore() -> SessionStore {
#if DEBUG
        if ProcessInfo.processInfo.arguments.contains("-UITestAuthenticated") ||
            ProcessInfo.processInfo.arguments.contains("-UITestAuthenticatedAPI") {
            return InMemorySessionStore()
        }
#endif
        return KeychainSessionStore()
    }

    func restore() async {
#if DEBUG
        if ProcessInfo.processInfo.arguments.contains("-UITestUnauthenticated") {
            store.clear()
            biometricEnabled = false
            UserDefaults.standard.set(false, forKey: biometricPreferenceKey)
            phase = .unauthenticated
            return
        }
        if ProcessInfo.processInfo.arguments.contains("-UITestAuthenticatedAPI") {
            do {
                let requestBody = try JSONEncoder.mypocket.encode(
                    LocalAuthFixtureExchangeRequest(
                        code: "fixture",
                        codeVerifier: "fixture",
                        redirectURI: "mypocket://oauth/callback",
                        platform: "ios"
                    )
                )
                let output = try await financeAPI.client.send(
                    APIRequest<LoginOutput>(
                        path: "/auth/mobile/google/exchange",
                        method: .post,
                        body: requestBody,
                        auth: .none
                    )
                )
                try store.save(output.session)
                phase = .authenticated(output.session)
            } catch {
#if DEBUG
                print("[AuthFixture] local exchange failed: \(APIError.diagnosticDescription(error))")
                errorMessage = "Không thể khởi tạo phiên simulator local: \(APIError.diagnosticDescription(error))"
#else
                errorMessage = "Không thể khởi tạo phiên simulator local."
#endif
                phase = .unauthenticated
            }
            return
        }
        if ProcessInfo.processInfo.arguments.contains("-UITestAuthenticated") {
            phase = .authenticated(
                AccessSession(
                    token: "ui-test-token",
                    expiresAt: "2099-01-01T00:00:00Z",
                    refreshToken: "ui-test-refresh",
                    refreshExpiresAt: "2099-01-02T00:00:00Z",
                    user: AuthUser(id: "ui-test-user", name: "UI Test", email: "ui@example.com", timezone: "Asia/Ho_Chi_Minh")
                )
            )
            return
        }
#endif
        guard let session = store.load() else {
            phase = .unauthenticated
            return
        }
        if session.hasValidAccessToken() {
            apply(session)
            return
        }
        guard session.hasValidRefreshToken() else {
            expireSession()
            return
        }
        do {
            apply(try await financeAPI.client.refreshSessionForRestore())
        } catch {
            expireSession()
        }
    }

    private func apply(_ session: AccessSession) {
        phase = biometricEnabled ? .locked(session) : .authenticated(session)
    }

    private func expireSession() {
        store.clear()
        biometricEnabled = canSignInWithBiometrics
        preferences.set(biometricEnabled, forKey: biometricPreferenceKey)
        phase = .unauthenticated
    }

    func signIn(email: String, password: String) async {
        guard !isBusy else { return }
        let normalizedEmail = email.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !normalizedEmail.isEmpty, !password.isEmpty else {
            errorMessage = "Nhập email và mật khẩu để đăng nhập."
            return
        }
        isBusy = true
        errorMessage = nil
        defer { isBusy = false }
        do {
            let output = try await financeAPI.login(email: normalizedEmail, password: password)
            if let rememberedDevice, rememberedDevice.ownerID != output.user.id {
                clearDeviceEnrollment()
            }
            try store.save(output.session)
            phase = .authenticated(output.session)
        } catch APIError.unauthorized {
            errorMessage = "Email hoặc mật khẩu không đúng."
        } catch {
            errorMessage = "Không thể đăng nhập. Vui lòng thử lại."
        }
    }

    var canSignInWithBiometrics: Bool { rememberedDevice?.isValid == true }

    func logout() async {
        guard !isBusy else { return }
        isBusy = true
        errorMessage = nil
        defer { isBusy = false }
        do {
            try await endCurrentSession()
        } catch {
            errorMessage = "Chưa thể đăng xuất trên máy chủ. Kiểm tra kết nối và thử lại."
        }
    }

    private func endCurrentSession() async throws {
        if let session = store.load() {
            do { try await financeAPI.logout(refreshToken: session.refreshToken) }
            catch APIError.unauthorized { /* Session has already been revoked. */ }
        }
        store.clear()
        phase = .unauthenticated
        biometricEnabled = canSignInWithBiometrics
        preferences.set(biometricEnabled, forKey: biometricPreferenceKey)
    }

    func signInWithBiometrics() async {
        guard !isBusy, canSignInWithBiometrics else { return }
        isBusy = true
        errorMessage = nil
        phase = .unauthenticated
        defer { isBusy = false }
        do {
            let credential = try await deviceStore.unlock()
            guard credential.metadata == rememberedDevice else { throw KeychainError.status(errSecItemNotFound) }
            let output = try await financeAPI.loginWithDevice(id: credential.id, secret: credential.secret)
            guard output.user.id == credential.ownerID else { throw APIError.unauthorized }
            try store.save(output.session)
            biometricEnabled = true
            preferences.set(true, forKey: biometricPreferenceKey)
            phase = .authenticated(output.session)
        } catch APIError.unauthorized {
            clearDeviceEnrollment()
            errorMessage = "Đăng nhập Face ID đã hết hạn hoặc bị thu hồi. Hãy đăng nhập bằng mật khẩu để bật lại."
        } catch KeychainError.status(errSecItemNotFound) {
            clearDeviceEnrollment()
            errorMessage = "Cần đăng nhập bằng mật khẩu để thiết lập lại Face ID."
        } catch KeychainError.status(errSecUserCanceled) {
            // Cancellation is a normal choice; keep remembered sign-in.
        } catch let error as LAError where [.userCancel, .appCancel, .systemCancel].contains(error.code) {
            // Keep the password option and the protected device credential.
        } catch {
            errorMessage = "Chưa thể đăng nhập bằng Face ID. Bạn có thể thử lại hoặc dùng mật khẩu."
        }
    }

    @discardableResult
    func enrollBiometric(password: String, deviceName: String) async -> Bool {
        guard !isBusy, case .authenticated(let session) = phase else { return false }
        guard !password.isEmpty else { errorMessage = "Nhập mật khẩu để bật đăng nhập Face ID."; return false }
        guard deviceStore.isAvailable else { errorMessage = "Hãy thiết lập Face ID trong Cài đặt iPhone trước."; return false }
        isBusy = true
        errorMessage = nil
        defer { isBusy = false }
        let installationKey = "mypocket.biometric.installation-id"
        let installationID = preferences.string(forKey: installationKey) ?? UUID().uuidString
        preferences.set(installationID, forKey: installationKey)
        do {
            let output = try await financeAPI.enrollDevice(password: password, deviceID: installationID, deviceName: deviceName)
            do {
                try deviceStore.save(DeviceSignInCredential(id: output.id, secret: output.secret, expiresAt: output.expiresAt, ownerID: session.user.id, email: session.user.email ?? ""))
            } catch {
                try? await financeAPI.revokeDevice(id: output.id)
                clearDeviceEnrollment()
                throw error
            }
            rememberedDevice = deviceStore.metadata
            biometricEnabled = true
            preferences.set(true, forKey: biometricPreferenceKey)
            return true
        } catch APIError.problem(let problem) where problem.code == "device_enrollment_password_invalid" {
            errorMessage = "Mật khẩu xác nhận không đúng."
        } catch APIError.unauthorized {
            errorMessage = "Phiên đã hết hạn. Hãy đăng nhập lại."
        } catch {
            errorMessage = "Chưa thể bật đăng nhập Face ID. Vui lòng thử lại."
        }
        return false
    }

    private func clearDeviceEnrollment() {
        deviceStore.clear()
        rememberedDevice = nil
        biometricEnabled = false
        preferences.set(false, forKey: biometricPreferenceKey)
    }

    private func revokeDeviceEnrollment() async throws {
        if let rememberedDevice {
            do { try await financeAPI.revokeDevice(id: rememberedDevice.id) }
            catch APIError.http(status: 404) { /* Already forgotten on the server. */ }
            catch APIError.problem(let problem) where problem.status == 404 { }
        }
        clearDeviceEnrollment()
    }

    func disableBiometric() async {
        guard !isBusy else { return }
        isBusy = true
        defer { isBusy = false }
        do { try await revokeDeviceEnrollment(); errorMessage = nil }
        catch { errorMessage = "Chưa thể tắt đăng nhập Face ID. Vui lòng thử lại." }
    }

    func forgetDevice() async {
        guard !isBusy else { return }
        isBusy = true
        errorMessage = nil
        defer { isBusy = false }
        do {
            try await revokeDeviceEnrollment()
            try await endCurrentSession()
        } catch { errorMessage = "Chưa thể gỡ tài khoản khỏi thiết bị. Vui lòng thử lại." }
    }

    /// Keeps the persisted session's user context in sync after the account
    /// timezone API succeeds. Screens that derive calendar formatting from the
    /// authenticated session (for example notifications) must not keep using
    /// the pre-update timezone until the next app restore.
    func updateAccountTimezone(_ timezone: String) {
        guard case .authenticated(let session) = phase else { return }
        let updatedUser = AuthUser(id: session.user.id, name: session.user.name, email: session.user.email, timezone: timezone)
        let updatedSession = AccessSession(token: session.token, expiresAt: session.expiresAt, refreshToken: session.refreshToken, refreshExpiresAt: session.refreshExpiresAt, user: updatedUser)
        do {
            try store.save(updatedSession)
            phase = .authenticated(updatedSession)
        } catch {
            errorMessage = "Không thể lưu múi giờ cho phiên hiện tại."
        }
    }

    func unlock() async {
        guard case .locked(let session) = phase else { return }
        let context = LAContext()
        var error: NSError?
        guard context.canEvaluatePolicy(.deviceOwnerAuthentication, error: &error) else {
            errorMessage = "Thiết bị chưa sẵn sàng xác thực."
            return
        }
        do {
            let success = try await context.evaluatePolicy(.deviceOwnerAuthentication, localizedReason: "Mở khóa MyPocket")
            if success { phase = .authenticated(session) }
        } catch { errorMessage = "Không xác thực được Face ID hoặc mã khóa." }
    }
}
