import Foundation

actor RefreshCoordinator {
    private var active: Task<AccessSession, Error>?
    private let refresh: @Sendable () async throws -> AccessSession

    init(refresh: @escaping @Sendable () async throws -> AccessSession) {
        self.refresh = refresh
    }

    func refreshIfNeeded() async throws -> AccessSession {
        if let active { return try await active.value }
        let task = Task { try await refresh() }
        active = task
        defer { active = nil }
        return try await task.value
    }
}
