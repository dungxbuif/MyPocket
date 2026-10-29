import Foundation
import Combine
import Network

enum FinancialSubmission: Equatable, Sendable { case applied, queued }

@MainActor
final class FinancialMutationQueue: ObservableObject {
    @Published private(set) var items: [OutboxMutation] = []
    @Published private(set) var revision = 0
    @Published private(set) var errorMessage: String?
    @Published private(set) var result: SyncDrainResult?
    @Published private(set) var isSyncing = false

    static func makeDefault(api: FinanceAPI, sessions: any SessionStore) throws -> FinancialMutationQueue {
        let support = try FileManager.default.url(for: .applicationSupportDirectory, in: .userDomainMask, appropriateFor: nil, create: true)
        let store = try OutboxStore(directory: support.appendingPathComponent("MyPocket/outbox", isDirectory: true),
            serverURL: EnvironmentConfig.apiBaseURL, keys: KeychainOutboxKeyProvider(service: "com.mypocket.outbox.keys"))
        let queue = FinancialMutationQueue(store: store, api: api, sessions: sessions)
        queue.isForeground = false
        queue.networkAvailable = false
        queue.reachability = FinancialReachability { [weak queue] available in
            Task { @MainActor in try? await queue?.networkChanged(isAvailable: available) }
        }
        return queue
    }

    private let store: OutboxStore
    private let api: any OfflineSyncAPI
    private let sessions: any SessionStore
    private let now: @Sendable () -> Date
    private let sleep: @Sendable (TimeInterval) async throws -> Void
    private var ownerID: String?
    private var epoch: UInt64 = 0
    private var drainEpoch: UInt64?
    private var drainID: UUID?
    private var wakeID: UUID?
    private var wakeTask: Task<Void, Never>?
    private var reachability: FinancialReachability?
    private var isForeground = true
    private var networkAvailable = true
    private var resumeAuthPending = false
    private lazy var engine = SyncEngine(store: store, api: api, owner: { [sessions] in sessions.load()?.user.id }, now: now,
        invalidate: { [weak self] owner, _ in await self?.acknowledged(owner: owner) })

    init(store: OutboxStore, api: any OfflineSyncAPI, sessions: any SessionStore,
         now: @escaping @Sendable () -> Date = { Date() },
         sleep: @escaping @Sendable (TimeInterval) async throws -> Void = { try await Task.sleep(for: .seconds($0)) }) {
        self.store = store; self.api = api; self.sessions = sessions
        self.now = now; self.sleep = sleep
    }

    func updateLifecycle(ownerID: String?, isForeground: Bool) async throws {
        guard !Task.isCancelled else { return }
        let wasForeground = self.isForeground
        self.isForeground = isForeground
        if !isForeground { pausePublication() }
        await engine.setSuspended(!self.isForeground || !networkAvailable)
        guard !Task.isCancelled, self.isForeground == isForeground else { return }
        if self.ownerID != ownerID {
            try await activate(ownerID: ownerID)
        } else if isForeground && !wasForeground, let ownerID {
            let run = epoch
            try await refresh(owner: ownerID, epoch: run)
            if matches(owner: ownerID, epoch: run), !items.isEmpty { try await drain() }
        }
    }

    func networkChanged(isAvailable: Bool) async throws {
        let changed = networkAvailable != isAvailable
        networkAvailable = isAvailable
        if !isAvailable { pausePublication() }
        await engine.setSuspended(!isForeground || !networkAvailable)
        guard !Task.isCancelled, networkAvailable == isAvailable else { return }
        if changed && isAvailable && isForeground && !items.isEmpty {
            try await drain(reason: .network)
        }
    }

    deinit { wakeTask?.cancel() }

    func activate(ownerID: String?) async throws {
        guard !Task.isCancelled else { return }
        let previous = self.ownerID
        epoch &+= 1
        let activation = epoch
        cancelWake()
        self.ownerID = nil; items = []; errorMessage = nil; result = nil
        if let previous, previous != ownerID { try await engine.stop(accountID: previous) }
        guard epoch == activation, !Task.isCancelled else { return }
        guard let ownerID else { return }
        guard sessions.load()?.user.id == ownerID else { throw APIError.accountMismatch }
        self.ownerID = ownerID
        resumeAuthPending = true
        try await refresh(owner: ownerID, epoch: activation)
        if !items.isEmpty { try await drain(reason: .sameOwnerLogin) }
    }

    func submit(_ payload: OutboxPayload, key: UUID) async throws -> FinancialSubmission {
        guard let ownerID else { throw APIError.missingSession }
        guard sessions.load()?.user.id == ownerID else { throw APIError.accountMismatch }
        let submission = epoch
        let mutation = try await store.enqueue(accountID: ownerID, payload: payload, key: key)
        return try await finishPersisted(mutation, owner: ownerID, submission: submission)
    }

    private func finishPersisted(_ mutation: OutboxMutation, owner ownerID: String, submission: UInt64) async throws -> FinancialSubmission {
        guard matches(owner: ownerID, epoch: submission) else { throw APIError.accountMismatch }
        try await refresh(owner: ownerID, epoch: submission)
        // Once persisted, no online fallback/new key is permitted even if sync is unavailable.
        do { try await drain() }
        catch {
            guard matches(owner: ownerID, epoch: submission) else { throw APIError.accountMismatch }
            errorMessage = "Bản nháp đã lưu trên thiết bị, nhưng chưa thể đồng bộ."
        }
        guard matches(owner: ownerID, epoch: submission) else { throw APIError.accountMismatch }
        let persisted = try await store.items(accountID: ownerID)
        guard matches(owner: ownerID, epoch: submission) else { throw APIError.accountMismatch }
        return persisted.first(where: { $0.id == mutation.id })?.state == .applied ? .applied : .queued
    }

    func drain(reason: SyncTriggerReason = .foreground) async throws {
        guard !Task.isCancelled, isForeground, networkAvailable, !items.isEmpty else { return }
        guard drainEpoch != epoch else { return }
        guard let ownerID else { return }
        guard sessions.load()?.user.id == ownerID else { throw APIError.accountMismatch }
        let run = epoch
        let id = UUID()
        cancelWake()
        drainID = id
        drainEpoch = run
        isSyncing = true
        var completed = false
        defer {
            if drainID == id {
                drainID = nil; drainEpoch = nil; isSyncing = false
                if completed { scheduleNextWake() }
            }
        }
        errorMessage = nil
        let result = try await engine.trigger(accountID: ownerID, reason: resumeAuthPending ? .sameOwnerLogin : reason)
        guard matches(owner: ownerID, epoch: run), drainID == id else { return }
        if result != .stopped { resumeAuthPending = false }
        self.result = result
        try await refresh(owner: ownerID, epoch: run)
        completed = true
    }

    func retryNow(id: UUID) async throws {
        let owner = try boundOwner(), run = epoch
        try await store.retryNow(accountID: owner, id: id)
        guard matches(owner: owner, epoch: run) else { throw APIError.accountMismatch }
        try await refresh(owner: owner, epoch: run)
        try await drain(reason: .manualRetry)
    }
    func discard(id: UUID) async throws {
        let owner = try boundOwner(), run = epoch
        try await store.discard(accountID: owner, id: id)
        guard matches(owner: owner, epoch: run) else { throw APIError.accountMismatch }
        try await refresh(owner: owner, epoch: run)
        scheduleNextWake()
    }
    func replaceRejected(id: UUID, payload: OutboxPayload, key: UUID) async throws -> FinancialSubmission {
        let owner = try boundOwner(), run = epoch
        let child = try await store.replaceRejected(accountID: owner, id: id, payload: payload, key: key)
        return try await finishPersisted(child, owner: owner, submission: run)
    }

    private func boundOwner() throws -> String {
        guard let ownerID else { throw APIError.missingSession }
        guard sessions.load()?.user.id == ownerID else { throw APIError.accountMismatch }
        return ownerID
    }

    func stop() async throws {
        let previous = ownerID
        epoch &+= 1
        cancelWake()
        drainID = nil; drainEpoch = nil; isSyncing = false
        ownerID = nil; items = []; result = nil; errorMessage = nil
        if let previous { try await engine.stop(accountID: previous) }
    }

    private func cancelWake() {
        wakeID = nil
        wakeTask?.cancel(); wakeTask = nil
    }

    private func pausePublication() {
        cancelWake()
        drainID = nil; drainEpoch = nil; isSyncing = false
    }

    private func scheduleNextWake() {
        cancelWake()
        guard isForeground, networkAvailable, drainID == nil, let ownerID,
              matches(owner: ownerID, epoch: epoch),
              let head = items.first(where: { [.pending, .sending, .retryWait, .blockedAuth].contains($0.state) }),
              head.state != .blockedAuth else { return }
        let due: Date
        switch head.state {
        case .sending: guard let at = head.leaseExpiresAt else { return }; due = at
        case .retryWait: guard let at = head.nextAttemptAt else { return }; due = at
        default: due = now()
        }
        var delay = max(0.25, due.timeIntervalSince(now()))
        if result == .unsupported { delay = max(30, delay) }
        guard delay.isFinite else { return }
        let id = UUID(), run = epoch, sleep = self.sleep
        wakeID = id
        wakeTask = Task { [weak self] in
            do { try await sleep(delay) } catch { return }
            guard !Task.isCancelled, let self, self.wakeID == id,
                  self.matches(owner: ownerID, epoch: run), self.isForeground, self.networkAvailable else { return }
            self.wakeID = nil; self.wakeTask = nil
            do { try await self.drain() }
            catch {
                if self.matches(owner: ownerID, epoch: run) {
                    self.errorMessage = "Bản nháp vẫn được giữ; chưa thể đồng bộ. Hãy thử lại."
                }
            }
        }
    }

    private func matches(owner: String, epoch: UInt64) -> Bool {
        self.epoch == epoch && ownerID == owner && sessions.load()?.user.id == owner
    }

    private func refresh(owner: String, epoch: UInt64) async throws {
        let rows = try await store.items(accountID: owner)
        guard matches(owner: owner, epoch: epoch) else { return }
        items = rows.filter { $0.state != .applied && $0.state != .cancelled }
    }

    private func acknowledged(owner: String) {
        guard ownerID == owner, sessions.load()?.user.id == owner else { return }
        revision &+= 1
    }
}

/// SDK reachability is a wake hint only; capability/auth/receipt remain authoritative.
private final class FinancialReachability: @unchecked Sendable {
    private let monitor = NWPathMonitor()
    init(onChange: @escaping @Sendable (Bool) -> Void) {
        monitor.pathUpdateHandler = { path in onChange(path.status == .satisfied) }
        monitor.start(queue: DispatchQueue(label: "com.mypocket.offline.reachability"))
    }
    deinit { monitor.cancel() }
}
