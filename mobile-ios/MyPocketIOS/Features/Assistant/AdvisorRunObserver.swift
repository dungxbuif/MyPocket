import Foundation
import Combine

enum AdvisorRunConnectionState: Equatable {
    case idle
    case observing
    case reconnecting
    case suspended
    case finished
    case failed
}

/// Owns GET-only observation of one server-side advisor run. It deliberately
/// has no submit capability: a lost response can be recovered without ever
/// creating a second provider request.
@MainActor
final class AdvisorRunObserver: ObservableObject {
    typealias FetchRun = (String) async throws -> AdvisorRunModel
    typealias Sleep = (Duration) async throws -> Void

    @Published private(set) var activeRun: AdvisorRunModel?
    @Published private(set) var terminalResult: AdvisorRunModel?
    @Published private(set) var connectionState: AdvisorRunConnectionState = .idle
    @Published private(set) var observedRunID: String?
    @Published private(set) var observedOwnerID: String?
    @Published private(set) var lastError: String?

    private let fetchRun: FetchRun
    private let sleep: Sleep
    private var generation = UUID()
    private let terminalStatuses: Set<String> = ["completed", "failed", "cancelled", "interrupted", "purged"]

    init(fetchRun: @escaping FetchRun, sleep: @escaping Sleep = { duration in
        try await Task.sleep(for: duration)
    }) {
        self.fetchRun = fetchRun
        self.sleep = sleep
    }

    func observe(ownerID: String, runID: String) async {
        let owner = ownerID.trimmingCharacters(in: .whitespacesAndNewlines)
        let run = runID.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !owner.isEmpty, !run.isEmpty else { return }

        let token = UUID()
        generation = token
        observedOwnerID = owner
        observedRunID = run
        activeRun = nil
        terminalResult = nil
        lastError = nil
        connectionState = .observing

        var normalIndex = 0
        var networkIndex = 0
        while generation == token {
            do {
                let result = try await fetchRun(run)
                guard generation == token else { return }
                guard result.id == run else {
                    lastError = "run identity mismatch"
                    connectionState = .failed
                    return
                }
                guard result.ownerID == nil || result.ownerID == owner else {
                    lastError = "run owner mismatch"
                    connectionState = .reconnecting
                    try? await waitForRetry(index: &networkIndex, token: token, schedule: [.seconds(3), .seconds(5), .seconds(10)])
                    continue
                }
                lastError = nil
                networkIndex = 0
                activeRun = result
                if terminalStatuses.contains(result.status) {
                    terminalResult = result
                    connectionState = .finished
                    return
                }
                connectionState = .observing
                try await waitForRetry(index: &normalIndex, token: token, schedule: [.seconds(1), .seconds(2), .seconds(3)])
            } catch is CancellationError {
                return
            } catch {
                guard generation == token else { return }
                connectionState = .reconnecting
                lastError = error.localizedDescription
                try? await waitForRetry(index: &networkIndex, token: token, schedule: [.seconds(3), .seconds(5), .seconds(10)])
            }
        }
    }

    func suspend() {
        generation = UUID()
        connectionState = .suspended
    }

    func reset() {
        generation = UUID()
        activeRun = nil
        terminalResult = nil
        observedRunID = nil
        observedOwnerID = nil
        lastError = nil
        connectionState = .idle
    }

    private func waitForRetry(index: inout Int, token: UUID, schedule: [Duration]) async throws {
        guard generation == token else { return }
        let duration = schedule[min(index, schedule.count - 1)]
        index += 1
        try await sleep(duration)
    }
}
