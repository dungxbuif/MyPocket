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
    private let google: GooglePKCEClient
    let financeAPI: FinanceAPI
    private let biometricPreferenceKey = "mypocket.biometric.enabled"

    init(store: SessionStore = KeychainSessionStore(), apiClient: APIClient = APIClient()) {
        self.store = store
        self.google = GooglePKCEClient(apiClient: apiClient)
        self.financeAPI = FinanceAPI(client: apiClient)
        self.biometricEnabled = UserDefaults.standard.bool(forKey: biometricPreferenceKey)
#if DEBUG
        let isUITestAuthenticated = ProcessInfo.processInfo.arguments.contains("-UITestAuthenticated")
#else
        let isUITestAuthenticated = false
#endif
        if !isUITestAuthenticated {
            apiClient.onSessionInvalidated = { [weak self] in
                Task { @MainActor [weak self] in
                    self?.expireSession()
                }
            }
        }
    }

    func restore() async {
#if DEBUG
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
                errorMessage = "Không thể khởi tạo phiên simulator local."
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

    func signIn() async {
        guard !isBusy else { return }
        isBusy = true
        errorMessage = nil
        defer { isBusy = false }
        do {
            let output = try await google.signIn()
            try store.save(output.session)
            phase = .authenticated(output.session)
        } catch GooglePKCEError.cancelled {
            errorMessage = "Đăng nhập đã được hủy."
        } catch GooglePKCEError.invalidConfiguration {
            errorMessage = "Google OAuth chưa được cấu hình cho bản build này."
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
