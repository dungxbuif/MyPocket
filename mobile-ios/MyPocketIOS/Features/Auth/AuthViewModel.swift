import Foundation
import LocalAuthentication
import SwiftUI

enum AuthPhase: Equatable {
    case checking
    case unauthenticated
    case authenticated(AccessSession)
}

@MainActor
final class AuthViewModel: ObservableObject {
    @Published private(set) var phase: AuthPhase = .checking
    @Published var errorMessage: String?
    @Published var isBusy = false
    @Published var biometricEnabled = false

    private let store: KeychainSessionStore
    private let google: GooglePKCEClient
    let financeAPI: FinanceAPI

    init(store: KeychainSessionStore = KeychainSessionStore(), apiClient: APIClient = APIClient()) {
        self.store = store
        self.google = GooglePKCEClient(apiClient: apiClient)
        self.financeAPI = FinanceAPI(client: apiClient)
    }

    func restore() {
#if DEBUG
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
        if let session = store.load() {
            phase = .authenticated(session)
        } else {
            phase = .unauthenticated
        }
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
        } catch {
            errorMessage = "Không thể đăng nhập. Vui lòng thử lại."
        }
    }

    func logout() {
        store.clear()
        biometricEnabled = false
        phase = .unauthenticated
    }

    func toggleBiometric() async {
        guard case .authenticated = phase else { return }
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
            if !biometricEnabled { errorMessage = "Không xác thực được Face ID." }
        } catch {
            biometricEnabled = false
            errorMessage = "Không xác thực được Face ID."
        }
    }
}
