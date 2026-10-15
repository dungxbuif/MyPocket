import Foundation
import XCTest
@testable import MyPocketIOS

private struct SyncFixtureKeys: OutboxKeyProviding {
    func key(for namespace: String, createIfMissing: Bool) throws -> Data { Data(repeating: 0x39, count: 32) }
}

private actor SyncBoundary: OfflineSyncAPI {
    var mutations: [OutboxMutation] = []
    var capabilitiesCalls = 0
    var errors: [APIError]
    let supported: Bool
    let holdApply: Bool
    private var releaseSend: CheckedContinuation<Void, Never>?
    private var sendWaiters: [CheckedContinuation<Void, Never>] = []
    init(errors: [APIError] = [], supported: Bool = true, holdApply: Bool = false) {
        self.errors = errors; self.supported = supported; self.holdApply = holdApply
    }
    func waitForSend() async {
        if !mutations.isEmpty { return }
        await withCheckedContinuation { sendWaiters.append($0) }
    }
    func release() { releaseSend?.resume(); releaseSend = nil }
    func syncCapabilities(ownerID: String?) async throws -> OfflineSyncCapabilitiesModel {
        capabilitiesCalls += 1
        return OfflineSyncCapabilitiesModel(contractVersion: 1, enabled: supported, operations: ["transaction.create", "transfer.create"], receiptMetadata: "/meta/mutation", idempotencyKeyRequiredForOffline: true, legacyUnkeyedOnline: true, transactionCreateContext: "explicit_travel_v1")
    }
    func applyOutbox(_ mutation: OutboxMutation) async throws -> [String] {
        mutations.append(mutation)
        if holdApply {
            await withCheckedContinuation { continuation in
                releaseSend = continuation
                for waiter in sendWaiters { waiter.resume() }
                sendWaiters.removeAll()
            }
        }
        if !errors.isEmpty { throw errors.removeFirst() }
        return ["server-" + mutation.id.uuidString]
    }
}

private actor SyncInvalidations {
    var owners: [String] = []
    func record(_ owner: String) { owners.append(owner) }
}

final class SyncEngineTests: XCTestCase {
    func testLogoutBlocksLeaseAndIgnoresNoncooperativeLateSuccess() async throws {
        let store = try store(), api = SyncBoundary(holdApply: true), invalidations = SyncInvalidations()
        _ = try await enqueue(store)
        let sync = engine(store, api, invalidations: invalidations)
        let task = Task { try await sync.trigger(accountID: "owner-a", reason: .foreground) }
        await api.waitForSend()
        try await sync.stop(accountID: "owner-a")
        await api.release()
        let result = try await task.value
        XCTAssertEqual(result, .stopped)
        let rows = try await store.items(accountID: "owner-a")
        XCTAssertEqual(rows.first?.state, .blockedAuth)
        XCTAssertTrue(rows.first?.serverResourceIDs.isEmpty == true)
        let owners = await invalidations.owners
        XCTAssertTrue(owners.isEmpty)
    }

    func testAuthBlockedQueueRequiresExplicitSameOwnerLoginToResumeOriginalKey() async throws {
        let store = try store(), api = SyncBoundary(errors: [.unauthorized])
        let original = try await enqueue(store)
        let sync = engine(store, api)
        let first = try await sync.trigger(accountID: "owner-a", reason: .foreground)
        XCTAssertEqual(first, .blockedAuth)
        let ordinaryWake = try await sync.trigger(accountID: "owner-a", reason: .network)
        XCTAssertEqual(ordinaryWake, .blockedAuth)
        var sent = await api.mutations
        XCTAssertEqual(sent.count, 1)
        let resumed = try await sync.trigger(accountID: "owner-a", reason: .sameOwnerLogin)
        XCTAssertEqual(resumed, .drained)
        sent = await api.mutations
        XCTAssertEqual(sent.map(\.idempotencyKey), [original.idempotencyKey, original.idempotencyKey])
    }

    func testBackoffIsExponentialJitteredAndCappedAtFifteenMinutes() throws {
        for (attempt, jitter, delay) in [(1, 0.0, 1.0), (2, 0.0, 2.0), (3, 1.0, 5.0), (Int.max, 1.0, 900.0)] {
            guard case .retry(let at, let code) = SyncPolicy.outcome(APIError.http(status: 429), attempt: attempt, now: instant, jitter: jitter) else { return XCTFail("Expected scheduled retry") }
            XCTAssertEqual(at.timeIntervalSince(instant), delay)
            XCTAssertEqual(code, "HTTP_429")
        }
    }
    private let instant = Date(timeIntervalSince1970: 1790000000)
    private func store() throws -> OutboxStore {
        let root = FileManager.default.temporaryDirectory.appendingPathComponent("sync-engine-" + UUID().uuidString)
        addTeardownBlock { if FileManager.default.fileExists(atPath: root.path) { try FileManager.default.removeItem(at: root) } }
        return try OutboxStore(directory: root, serverURL: URL(string: "https://fixture.test/api/v1")!, keys: SyncFixtureKeys())
    }
    private func enqueue(_ store: OutboxStore) async throws -> OutboxMutation {
        try await store.enqueue(accountID: "owner-a", payload: .transaction(TransactionInputModel(walletID: "wallet", categoryID: "category", jarID: nil, type: "expense", amount: 12500, occurredAt: "2026-09-01T03:00:00Z", note: nil, includedInReports: true, travelEventID: nil)), key: UUID(), now: instant)
    }
    private func engine(_ store: OutboxStore, _ api: SyncBoundary, now: Date? = nil, invalidations: SyncInvalidations = SyncInvalidations()) -> SyncEngine {
        let date = now ?? instant
        return SyncEngine(store: store, api: api, owner: { "owner-a" }, now: { date }, jitter: { 0 }, invalidate: { owner, _ in await invalidations.record(owner) })
    }

    func testDrainsFIFOAndInvalidatesOnlyAfterDurableAppliedOutcome() async throws {
        let store = try store(), api = SyncBoundary(), invalidations = SyncInvalidations()
        let first = try await enqueue(store), second = try await enqueue(store)
        let result = try await engine(store, api, invalidations: invalidations).trigger(accountID: "owner-a", reason: .foreground)
        XCTAssertEqual(result, .drained)
        let rows = try await store.items(accountID: "owner-a")
        XCTAssertEqual(rows.map(\.state), [.applied, .applied])
        let sent = await api.mutations
        XCTAssertEqual(sent.map(\.id), [first.id, second.id])
        let owners = await invalidations.owners
        XCTAssertEqual(owners, ["owner-a", "owner-a"])
    }

    func testUnsupportedCapabilityPreservesPendingIntentWithoutLeasingOrSending() async throws {
        let store = try store(), api = SyncBoundary(supported: false)
        _ = try await enqueue(store)
        let result = try await engine(store, api).trigger(accountID: "owner-a", reason: .network)
        XCTAssertEqual(result, .unsupported)
        let rows = try await store.items(accountID: "owner-a")
        XCTAssertEqual(rows.first?.state, .pending); XCTAssertEqual(rows.first?.attemptCount, 0)
        let sent = await api.mutations
        XCTAssertTrue(sent.isEmpty)
    }

    func testTimeoutPersistsRetryAndReplaysSameIntentAfterDueTime() async throws {
        let store = try store(), api = SyncBoundary(errors: [.transport("private-body-must-not-be-stored")])
        let original = try await enqueue(store)
        let first = try await engine(store, api).trigger(accountID: "owner-a", reason: .foreground)
        XCTAssertEqual(first, .waiting)
        var rows = try await store.items(accountID: "owner-a")
        XCTAssertEqual(rows.first?.state, .retryWait)
        XCTAssertEqual(rows.first?.lastErrorCode, "NETWORK_UNAVAILABLE")
        let due = try XCTUnwrap(rows.first?.nextAttemptAt)
        XCTAssertGreaterThan(due, instant)
        let result = try await engine(store, api, now: due).trigger(accountID: "owner-a", reason: .network)
        XCTAssertEqual(result, .drained)
        rows = try await store.items(accountID: "owner-a")
        XCTAssertEqual(rows.first?.state, .applied)
        let sent = await api.mutations
        XCTAssertEqual(sent.map(\.idempotencyKey), [original.idempotencyKey, original.idempotencyKey])
        XCTAssertEqual(try sent[0].payload.canonicalData(), try sent[1].payload.canonicalData())
    }

    func testHTTPFailuresBecomeRetryConflictRejectionOrAuthBlockWithoutRawMessage() async throws {
        for (status, expected) in [(408, OutboxState.retryWait), (429, .retryWait), (503, .retryWait), (400, .rejected), (403, .rejected), (404, .rejected), (409, .conflict), (422, .rejected), (401, .blockedAuth)] {
            let store = try store(), api = SyncBoundary(errors: [.http(status: status)])
            _ = try await enqueue(store)
            _ = try await engine(store, api).trigger(accountID: "owner-a", reason: .foreground)
            let rows = try await store.items(accountID: "owner-a")
            XCTAssertEqual(rows.first?.state, expected, "HTTP \(status)")
            XCTAssertNil(rows.first?.leaseID)
            let sent = await api.mutations
            XCTAssertEqual(sent.count, 1)
        }
    }

    func testWrongAccountBlocksQueueAndNeverSends() async throws {
        let store = try store(), api = SyncBoundary()
        _ = try await enqueue(store)
        let engine = SyncEngine(store: store, api: api, owner: { "owner-b" })
        let result = try await engine.trigger(accountID: "owner-a", reason: .foreground)
        XCTAssertEqual(result, .blockedAuth)
        let rows = try await store.items(accountID: "owner-a")
        XCTAssertEqual(rows.first?.state, .blockedAuth)
        let sent = await api.mutations
        XCTAssertTrue(sent.isEmpty)
    }
}
