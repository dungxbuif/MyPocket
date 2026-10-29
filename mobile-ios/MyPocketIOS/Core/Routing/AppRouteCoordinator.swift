import SwiftUI

/// Holds an external route until the current auth phase is safe to navigate.
/// It deliberately does not perform navigation or data loading itself.
@MainActor
final class AppRouteCoordinator: ObservableObject {
    @Published private(set) var pendingRoute: AppRoute?

    func receive(_ route: AppRoute) {
        pendingRoute = route
    }

    /// Drops a route that was waiting for authentication when the protected
    /// session was explicitly ended or invalidated. A URL received while no
    /// session exists is intentionally retained until the first successful
    /// login; the app root decides when a real session transition invalidates
    /// it.
    func discardPendingRoute() {
        pendingRoute = nil
    }

    func consumeIfAuthenticated(_ phase: AuthPhase) -> AppRoute? {
        guard case .authenticated = phase, let route = pendingRoute else { return nil }
        pendingRoute = nil
        return route
    }
}
