import SwiftUI

/// Holds an external route until the current auth phase is safe to navigate.
/// It deliberately does not perform navigation or data loading itself.
@MainActor
final class AppRouteCoordinator: ObservableObject {
    @Published private(set) var pendingRoute: AppRoute?

    func receive(_ route: AppRoute) {
        pendingRoute = route
    }

    func consumeIfAuthenticated(_ phase: AuthPhase) -> AppRoute? {
        guard case .authenticated = phase, let route = pendingRoute else { return nil }
        pendingRoute = nil
        return route
    }
}
