import SwiftUI

struct AppRootView: View {
    @StateObject private var auth: AuthViewModel
    @StateObject private var routeCoordinator: AppRouteCoordinator
    @State private var activeRoute: AppRoute?
    @Environment(\.scenePhase) private var scenePhase

    init(auth: AuthViewModel = AuthViewModel()) {
        _auth = StateObject(wrappedValue: auth)
        _routeCoordinator = StateObject(wrappedValue: AppRouteCoordinator())
    }

    var body: some View {
        ZStack {
            MPColor.canvas.ignoresSafeArea()
            Group {
                switch auth.phase {
                case .checking:
                    MPLoadingView(label: "Đang kiểm tra phiên...")
                case .unauthenticated:
                    AuthView(model: auth)
                case .locked:
                    BiometricUnlockView(model: auth)
                case .authenticated:
                    MainTabView(auth: auth, externalRoute: activeRoute)
                }
            }
        }
        .font(MPFont.body)
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .top)
        .task { await auth.restore() }
        .onOpenURL { url in
            guard let route = AppRouteParser.parse(url) else { return }
            routeCoordinator.receive(route)
            activeRoute = routeCoordinator.consumeIfAuthenticated(auth.phase)
        }
        .onChange(of: auth.phase) { _, phase in
            switch phase {
            case .authenticated:
                activeRoute = routeCoordinator.consumeIfAuthenticated(phase)
            case .checking, .locked, .unauthenticated:
                // A consumed route must not be replayed after logout or a new
                // login; only a fresh URL may create a new pending route.
                activeRoute = nil
            }
        }
        .environmentObject(routeCoordinator)
        .environment(\.mpFinancialQueue, auth.financialQueue)
        .task(id: FinancialQueueLifecycle(ownerID: authenticatedOwnerID, foreground: scenePhase == .active)) {
            try? await auth.financialQueue?.updateLifecycle(ownerID: authenticatedOwnerID, isForeground: scenePhase == .active)
        }
    }

    private var authenticatedOwnerID: String? {
        if case .authenticated(let session) = auth.phase { return session.user.id }
        return nil
    }
}

private struct FinancialQueueLifecycle: Hashable {
    let ownerID: String?
    let foreground: Bool
}

#Preview {
    AppRootView()
}
