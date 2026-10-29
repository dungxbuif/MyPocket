import Foundation

protocol OfflineSyncAPI: Sendable {
    func syncCapabilities(ownerID: String?) async throws -> OfflineSyncCapabilitiesModel
    func applyOutbox(_ mutation: OutboxMutation) async throws -> [String]
}

extension FinanceAPI: OfflineSyncAPI {}

enum SyncTriggerReason: Sendable { case foreground, network, manualRetry, sameOwnerLogin }
enum SyncDrainResult: Equatable, Sendable { case drained, waiting, unsupported, blockedAuth, stopped }

actor SyncEngine {
    private let store: OutboxStore
    private let api: any OfflineSyncAPI
    private let owner: @Sendable () -> String?
    private let now: @Sendable () -> Date
    private let jitter: @Sendable () -> Double
    private let invalidate: @Sendable (String, [String]) async -> Void
    private var generation: UInt64 = 0
    private var suspended = false
    private var active: (owner: String, generation: UInt64, task: Task<SyncDrainResult, Error>)?

    init(store: OutboxStore, api: any OfflineSyncAPI,
         owner: @escaping @Sendable () -> String?,
         now: @escaping @Sendable () -> Date = { Date() },
         jitter: @escaping @Sendable () -> Double = { Double.random(in: 0...1) },
         invalidate: @escaping @Sendable (String, [String]) async -> Void = { _, _ in }) {
        self.store = store; self.api = api; self.owner = owner; self.now = now
        self.jitter = jitter; self.invalidate = invalidate
    }

    func trigger(accountID: String, reason: SyncTriggerReason) async throws -> SyncDrainResult {
        guard !suspended, !Task.isCancelled else { return .stopped }
        if let active {
            guard active.owner == accountID else { throw APIError.accountMismatch }
            return try await active.task.value
        }
        generation &+= 1
        let run = generation
        let task = Task { try await self.drain(accountID: accountID, reason: reason, run: run) }
        active = (accountID, run, task)
        defer { if active?.generation == run { active = nil } }
        return try await task.value
    }

    func setSuspended(_ value: Bool) {
        guard suspended != value else { return }
        suspended = value
        if value {
            generation &+= 1
            active?.task.cancel()
            active = nil
        }
    }

    func stop(accountID: String) async throws {
        if let active, active.owner == accountID {
            generation &+= 1
            active.task.cancel()
            self.active = nil
        }
        try await store.blockForAuth(accountID: accountID)
    }

    private func current(_ run: UInt64) -> Bool { generation == run && !suspended && !Task.isCancelled }

    private func requireOwner(_ accountID: String) async throws -> Bool {
        guard owner() == accountID else {
            try await store.blockForAuth(accountID: accountID)
            return false
        }
        return true
    }

    private func drain(accountID: String, reason: SyncTriggerReason, run: UInt64) async throws -> SyncDrainResult {
        guard current(run) else { return .stopped }
        guard try await requireOwner(accountID) else { return .blockedAuth }
        if case .sameOwnerLogin = reason { try await store.resumeAuth(accountID: accountID) }
        guard current(run) else { return .stopped }
        let capabilities: OfflineSyncCapabilitiesModel
        do { capabilities = try await api.syncCapabilities(ownerID: accountID) }
        catch {
            guard current(run) else { return .stopped }
            if SyncPolicy.isAuthFailure(error) {
                try await store.blockForAuth(accountID: accountID)
                return .blockedAuth
            }
            return .unsupported
        }
        while current(run) {
            guard try await requireOwner(accountID) else { return .blockedAuth }
            let items = try await store.items(accountID: accountID)
            guard current(run) else { return .stopped }
            guard let next = items.first(where: { [.pending, .sending, .retryWait, .blockedAuth].contains($0.state) }) else { return .drained }
            if next.state == .blockedAuth { return .blockedAuth }
            guard capabilities.supports(operation: next.payload.operation) else { return .unsupported }
            guard let mutation = try await store.leaseNext(accountID: accountID, now: now()), let leaseID = mutation.leaseID else { return .waiting }
            guard current(run) else { return .stopped }
            // Another process/store may have advanced FIFO between inspection and lease.
            guard capabilities.supports(operation: mutation.payload.operation) else { throw OutboxStoreError.invalidPayload }
            do {
                let ids = try await api.applyOutbox(mutation)
                guard current(run) else { return .stopped }
                guard try await requireOwner(accountID) else { return .blockedAuth }
                try await store.markOutcome(accountID: accountID, id: mutation.id, leaseID: leaseID, outcome: .applied(resourceIDs: ids))
                guard current(run), owner() == accountID else { return .stopped }
                await invalidate(accountID, ids)
            } catch {
                guard current(run) else { return .stopped }
                guard try await requireOwner(accountID) else { return .blockedAuth }
                if SyncPolicy.isAuthFailure(error) {
                    try await store.blockForAuth(accountID: accountID)
                    return .blockedAuth
                }
                // Lease/storage failure is not a provider error and must not be overwritten.
                if error is OutboxStoreError { throw error }
                let outcome = SyncPolicy.outcome(error, attempt: mutation.attemptCount, now: now(), jitter: jitter())
                try await store.markOutcome(accountID: accountID, id: mutation.id, leaseID: leaseID, outcome: outcome)
                if case .retry = outcome { return .waiting }
            }
        }
        return .stopped
    }
}

enum SyncPolicy {
    static func isAuthFailure(_ error: Error) -> Bool {
        guard let api = error as? APIError else { return false }
        switch api {
        case .unauthorized, .missingSession, .accountMismatch: return true
        case .http(let status): return status == 401
        case .problem(let problem): return problem.status == 401
        default: return false
        }
    }

    static func outcome(_ error: Error, attempt: Int, now: Date, jitter: Double) -> OutboxOutcome {
        func retry(_ code: String) -> OutboxOutcome {
            let exponent = min(max(attempt - 1, 0), 10)
            let random = jitter.isFinite ? min(max(jitter, 0), 1) : 0
            let delay = min(900, pow(2, Double(exponent)) * (1 + random * 0.25))
            return .retry(at: now.addingTimeInterval(delay), errorCode: code)
        }
        func http(_ status: Int) -> OutboxOutcome {
            if status == 408 || status == 429 || (500...599).contains(status) { return retry("HTTP_\(status)") }
            if status == 409 { return .conflict(errorCode: "HTTP_409") }
            return .rejected(errorCode: "HTTP_\(max(0, status))")
        }
        if error is OutboxAcknowledgementError { return retry("ACK_UNVERIFIED") }
        guard let api = error as? APIError else { return .rejected(errorCode: "CLIENT_FAILURE") }
        switch api {
        case .transport: return retry("NETWORK_UNAVAILABLE")
        case .decoding: return retry("ACK_UNVERIFIED")
        case .http(let status): return http(status)
        case .problem(let problem): return http(problem.status ?? 400)
        default: return .rejected(errorCode: "CLIENT_FAILURE")
        }
    }
}
