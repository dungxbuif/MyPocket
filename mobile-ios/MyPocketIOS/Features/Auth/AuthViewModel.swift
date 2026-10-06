import Foundation
import LocalAuthentication
import SwiftUI

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

    private let store: SessionStore
    let financeAPI: FinanceAPI
    private let biometricPreferenceKey = "mypocket.biometric.enabled"

    init(store: SessionStore? = nil, apiClient: APIClient? = nil) {
        let resolvedStore = store ?? Self.makeDefaultStore()
        self.store = resolvedStore
        let resolvedAPI = apiClient ?? APIClient(sessionStore: resolvedStore)
        self.financeAPI = FinanceAPI(client: resolvedAPI)
        self.biometricEnabled = UserDefaults.standard.bool(forKey: biometricPreferenceKey)
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
                    MobileGoogleExchangeRequest(
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
                print("[AuthFixture] local exchange failed: \(String(reflecting: error))")
                errorMessage = "Không thể khởi tạo phiên simulator local: \(String(describing: error))"
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
        biometricEnabled = false
        UserDefaults.standard.set(false, forKey: biometricPreferenceKey)
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
            try store.save(output.session)
            phase = .authenticated(output.session)
        } catch APIError.unauthorized {
            errorMessage = "Email hoặc mật khẩu không đúng."
        } catch {
            errorMessage = "Không thể đăng nhập. Vui lòng thử lại."
        }
    }

    func logout() {
        store.clear()
        biometricEnabled = false
        UserDefaults.standard.set(false, forKey: biometricPreferenceKey)
        phase = .unauthenticated
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

    func toggleBiometric() async {
        guard case .authenticated = phase else { return }
        if biometricEnabled {
            biometricEnabled = false
            UserDefaults.standard.set(false, forKey: biometricPreferenceKey)
            return
        }
        let context = LAContext()
        var error: NSError?
        guard context.canEvaluatePolicy(.deviceOwnerAuthenticationWithBiometrics, error: &error) else {
            errorMessage = "Thiết bị chưa hỗ trợ Face ID."
            biometricEnabled = false
            return
        }
        do {
            biometricEnabled = try await context.evaluatePolicy(
                .deviceOwnerAuthenticationWithBiometrics,
                localizedReason: "Mở khóa MyPocket"
            )
            UserDefaults.standard.set(biometricEnabled, forKey: biometricPreferenceKey)
            if !biometricEnabled { errorMessage = "Không xác thực được Face ID." }
        } catch {
            biometricEnabled = false
            errorMessage = "Không xác thực được Face ID."
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
