import SwiftUI

struct AppRootView: View {
    @StateObject private var auth: AuthViewModel

    init(auth: AuthViewModel = AuthViewModel()) {
        _auth = StateObject(wrappedValue: auth)
    }

    var body: some View {
        Group {
            switch auth.phase {
            case .checking:
                MPLoadingView(label: "Đang kiểm tra phiên...")
            case .unauthenticated:
                AuthView(model: auth)
            case .authenticated:
                BiometricUnlockView(model: auth)
            }
        }
        .task { auth.restore() }
        .accessibilityIdentifier("app-root")
    }
}

#Preview {
    AppRootView()
}
