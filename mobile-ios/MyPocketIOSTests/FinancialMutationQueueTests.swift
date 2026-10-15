import Foundation
import XCTest
@testable import MyPocketIOS

private struct QueueFixtureKeys: OutboxKeyProviding {
    func key(for namespace: String, createIfMissing: Bool) throws -> Data { Data(repeating: 0x46, count: 32) }
}

private actor QueueBoundary: OfflineSyncAPI {
    let store: OutboxStore
    let enabled: Bool
    var observedPersistedSending = false
    var calls = 0
    var errors: [APIError] = []
    func failNext(_ error: APIError) { errors.append(error) }
    let holdFirstApply: Bool
    private var releaseSend: CheckedContinuation<Void, Never>?
    private var waiters: [CheckedContinuation<Void, Never>] = []
    init(store: OutboxStore, enabled: Bool, holdFirstApply: Bool = false) { self.store = store; self.enabled = enabled; self.holdFirstApply = holdFirstApply }
    func waitForFirstSend() async {
        if calls > 0 { return }
        await withCheckedContinuation { waiters.append($0) }
    }
    func release() { releaseSend?.resume(); releaseSend = nil }
    func syncCapabilities(ownerID: String?) async throws -> OfflineSyncCapabilitiesModel {
        OfflineSyncCapabilitiesModel(contractVersion: 1, enabled: enabled, operations: ["transaction.create", "transfer.create"], receiptMetadata: "/meta/mutation", idempotencyKeyRequiredForOffline: true, legacyUnkeyedOnline: true, transactionCreateContext: "explicit_travel_v1")
    }
    func applyOutbox(_ mutation: OutboxMutation) async throws -> [String] {
        calls += 1
        let rows = try await store.items(accountID: mutation.accountID)
        observedPersistedSending = rows.contains { $0.id == mutation.id && $0.state == .sending && $0.leaseID == mutation.leaseID }
        if holdFirstApply && calls == 1 {
            await withCheckedContinuation { continuation in
                releaseSend = continuation
                for waiter in waiters { waiter.resume() }; waiters.removeAll()
            }
        }
        if !errors.isEmpty { throw errors.removeFirst() }
        return ["server-row"]
    }
}

private final class QueueTestClock: @unchecked Sendable {
    private let lock = NSLock()
    private var value = Date()
    func now() -> Date { lock.withLock { value } }
    func advance(_ interval: TimeInterval) { lock.withLock { value.addTimeInterval(interval) } }
}

private actor QueueTestSleeper {
    private(set) var delays: [TimeInterval] = []
    private var pending: [CheckedContinuation<Void, Never>] = []
    func sleep(_ delay: TimeInterval) async {
        delays.append(delay)
        await withCheckedContinuation { pending.append($0) }
    }
    // Deliberately noncooperative cancellation tests the coordinator's wake token.
    func releaseAll() { let old = pending; pending.removeAll(); for continuation in old { continuation.resume() } }
}

@MainActor
final class FinancialMutationQueueTests: XCTestCase {
    func testBackgroundPausesInFlightDrainWithoutAuthBlockingOrAcknowledgingLateResponse() async throws {
        let store = try store(), sessions = InMemorySessionStore(), clock = QueueTestClock(), sleeper = QueueTestSleeper()
        addTeardownBlock { await sleeper.releaseAll() }
        try sessions.save(session("owner-a"))
        let api = QueueBoundary(store: store, enabled: true, holdFirstApply: true)
        let queue = FinancialMutationQueue(store: store, api: api, sessions: sessions, now: { clock.now() }, sleep: { await sleeper.sleep($0) })
        try await queue.activate(ownerID: "owner-a")
        let firstKey = UUID()
        let submission = Task { try await queue.submit(payload(), key: firstKey) }
        await api.waitForFirstSend()
        let second = try await store.enqueue(accountID: "owner-a", payload: payload(amount: 456), key: UUID())
        try await queue.updateLifecycle(ownerID: "owner-a", isForeground: false)
        await api.release()
        let accepted = try await submission.value
        XCTAssertEqual(accepted, .queued)
        var rows = try await store.items(accountID: "owner-a")
        XCTAssertEqual(rows.map(\.state), [.sending, .pending], "Pause must not acknowledge an uncertain response or drain the next row")
        XCTAssertEqual(queue.revision, 0)
        let pausedCalls = await api.calls; XCTAssertEqual(pausedCalls, 1)
        try await queue.drain(reason: .manualRetry)
        let backgroundCalls = await api.calls; XCTAssertEqual(backgroundCalls, 1)
        // No lease steal: only resume after the original uncertain lease expires.
        try await queue.updateLifecycle(ownerID: "owner-a", isForeground: true)
        rows = try await store.items(accountID: "owner-a")
        XCTAssertEqual(rows.map(\.state), [.sending, .pending], "Returning early must not steal the outstanding lease")
        let earlyCalls = await api.calls; XCTAssertEqual(earlyCalls, 1)
        let wakeDeadline = Date().addingTimeInterval(2)
        while await sleeper.delays.isEmpty && Date() < wakeDeadline { await Task.yield() }
        let leaseDelays = await sleeper.delays
        XCTAssertEqual(try XCTUnwrap(leaseDelays.first), 60, accuracy: 0.01)
        clock.advance(61); await sleeper.releaseAll()
        let appliedDeadline = Date().addingTimeInterval(2)
        while queue.revision < 2 && Date() < appliedDeadline { await Task.yield() }
        rows = try await store.items(accountID: "owner-a")
        XCTAssertEqual(rows.map(\.state), [.applied, .applied])
        XCTAssertEqual(rows[0].idempotencyKey, firstKey.uuidString)
        XCTAssertEqual(rows[1].id, second.id)
        XCTAssertEqual(queue.revision, 2)
        try await queue.stop()
    }

    func testForegroundTimerRespectsFirstRetryBeforeLaterPendingIntent() async throws {
        let store = try store(), sessions = InMemorySessionStore(), clock = QueueTestClock(), sleeper = QueueTestSleeper()
        addTeardownBlock { await sleeper.releaseAll() }
        try sessions.save(session("owner-a"))
        let first = try await store.enqueue(accountID: "owner-a", payload: payload(), key: UUID())
        let second = try await store.enqueue(accountID: "owner-a", payload: payload(amount: 456), key: UUID())
        let leased = try await store.leaseNext(accountID: "owner-a", now: clock.now())
        try await store.markOutcome(accountID: "owner-a", id: first.id, leaseID: XCTUnwrap(leased?.leaseID),
            outcome: .retry(at: clock.now().addingTimeInterval(100), errorCode: "HTTP_503"))
        let api = QueueBoundary(store: store, enabled: true)
        let queue = FinancialMutationQueue(store: store, api: api, sessions: sessions, now: { clock.now() }, sleep: { await sleeper.sleep($0) })
        try await queue.activate(ownerID: "owner-a")
        let initialDeadline = Date().addingTimeInterval(2)
        while await sleeper.delays.isEmpty && Date() < initialDeadline { await Task.yield() }
        let initialDelays = await sleeper.delays
        XCTAssertEqual(try XCTUnwrap(initialDelays.first), 100, accuracy: 0.01)
        clock.advance(99); await sleeper.releaseAll()
        let rescheduleDeadline = Date().addingTimeInterval(2)
        while await sleeper.delays.count < 2 && Date() < rescheduleDeadline { await Task.yield() }
        let earlyCalls = await api.calls; XCTAssertEqual(earlyCalls, 0)
        let stillWaiting = try await store.items(accountID: "owner-a")
        XCTAssertEqual(stillWaiting.map(\.state), [.retryWait, .pending])
        clock.advance(1); await sleeper.releaseAll()
        let appliedDeadline = Date().addingTimeInterval(2)
        while queue.revision < 2 && Date() < appliedDeadline { await Task.yield() }
        let rows = try await store.items(accountID: "owner-a")
        XCTAssertEqual(rows.map(\.id), [first.id, second.id])
        XCTAssertEqual(rows.map(\.state), [.applied, .applied])
        try await queue.stop()
    }

    func testUnsupportedPendingCapabilityUsesCooldownWithoutLeasingFinancialWrite() async throws {
        let store = try store(), sessions = InMemorySessionStore(), sleeper = QueueTestSleeper()
        addTeardownBlock { await sleeper.releaseAll() }
        try sessions.save(session("owner-a"))
        let api = QueueBoundary(store: store, enabled: false)
        let queue = FinancialMutationQueue(store: store, api: api, sessions: sessions, sleep: { await sleeper.sleep($0) })
        try await queue.activate(ownerID: "owner-a")
        _ = try await queue.submit(payload(), key: UUID())
        let deadline = Date().addingTimeInterval(2)
        while await sleeper.delays.isEmpty && Date() < deadline { await Task.yield() }
        let delays = await sleeper.delays
        XCTAssertEqual(try XCTUnwrap(delays.first), 30, accuracy: 0.01)
        let rows = try await store.items(accountID: "owner-a")
        XCTAssertEqual(rows.first?.state, .pending); XCTAssertEqual(rows.first?.attemptCount, 0)
        let calls = await api.calls; XCTAssertEqual(calls, 0)
        try await queue.stop()
    }

    func testCancelledWakeCannotResumeLoggedOutOwnerEvenWhenSleepIgnoresCancellation() async throws {
        let store = try store(), sessions = InMemorySessionStore(), clock = QueueTestClock(), sleeper = QueueTestSleeper()
        addTeardownBlock { await sleeper.releaseAll() }
        try sessions.save(session("owner-a"))
        let api = QueueBoundary(store: store, enabled: true)
        await api.failNext(.http(status: 503))
        let queue = FinancialMutationQueue(store: store, api: api, sessions: sessions, now: { clock.now() }, sleep: { await sleeper.sleep($0) })
        try await queue.activate(ownerID: "owner-a")
        _ = try await queue.submit(payload(), key: UUID())
        let deadline = Date().addingTimeInterval(2)
        while await sleeper.delays.isEmpty && Date() < deadline { await Task.yield() }
        sessions.clear(); try await queue.stop()
        try sessions.save(session("owner-b")); try await queue.activate(ownerID: "owner-b")
        clock.advance(900); await sleeper.releaseAll()
        for _ in 0..<20 { await Task.yield() }
        let old = try await store.items(accountID: "owner-a")
        XCTAssertEqual(old.first?.state, .blockedAuth)
        let calls = await api.calls; XCTAssertEqual(calls, 1)
        XCTAssertTrue(queue.items.isEmpty); XCTAssertEqual(queue.revision, 0)
        try await queue.stop()
    }

    func testForegroundAndNetworkWakeCannotSilentlyResumeAuthBlockedIntent() async throws {
        let store = try store(), sessions = InMemorySessionStore()
        try sessions.save(session("owner-a"))
        let api = QueueBoundary(store: store, enabled: true)
        await api.failNext(.unauthorized)
        let queue = FinancialMutationQueue(store: store, api: api, sessions: sessions)
        try await queue.activate(ownerID: "owner-a")
        _ = try await queue.submit(payload(), key: UUID())
        try await queue.updateLifecycle(ownerID: "owner-a", isForeground: false)
        try await queue.updateLifecycle(ownerID: "owner-a", isForeground: true)
        try await queue.networkChanged(isAvailable: false)
        try await queue.networkChanged(isAvailable: true)
        let rows = try await store.items(accountID: "owner-a")
        XCTAssertEqual(rows.first?.state, .blockedAuth)
        let calls = await api.calls; XCTAssertEqual(calls, 1)
        try await queue.stop()
    }

    func testBackgroundCreateStaysDurableWithoutNetworkAndForegroundAppliesOriginalKey() async throws {
        let store = try store(), sessions = InMemorySessionStore()
        try sessions.save(session("owner-a"))
        let api = QueueBoundary(store: store, enabled: true)
        let queue = FinancialMutationQueue(store: store, api: api, sessions: sessions)
        try await queue.updateLifecycle(ownerID: "owner-a", isForeground: false)
        let key = UUID()
        let result = try await queue.submit(payload(), key: key)
        XCTAssertEqual(result, .queued)
        var rows = try await store.items(accountID: "owner-a")
        XCTAssertEqual(rows.first?.state, .pending)
        let beforeCalls = await api.calls; XCTAssertEqual(beforeCalls, 0)
        try await queue.updateLifecycle(ownerID: "owner-a", isForeground: true)
        rows = try await store.items(accountID: "owner-a")
        XCTAssertEqual(rows.first?.state, .applied); XCTAssertEqual(rows.first?.idempotencyKey, key.uuidString)
        XCTAssertEqual(queue.revision, 1)
    }

    func testNetworkReturnDrainsQueuedIntentButUnavailablePathDoesNotSend() async throws {
        let store = try store(), sessions = InMemorySessionStore()
        try sessions.save(session("owner-a"))
        let api = QueueBoundary(store: store, enabled: true)
        let queue = FinancialMutationQueue(store: store, api: api, sessions: sessions)
        try await queue.activate(ownerID: "owner-a")
        try await queue.networkChanged(isAvailable: false)
        let key = UUID()
        let result = try await queue.submit(payload(), key: key)
        XCTAssertEqual(result, .queued)
        let beforeCalls = await api.calls; XCTAssertEqual(beforeCalls, 0)
        try await queue.networkChanged(isAvailable: true)
        let rows = try await store.items(accountID: "owner-a")
        XCTAssertEqual(rows.first?.state, .applied); XCTAssertEqual(rows.first?.idempotencyKey, key.uuidString)
        XCTAssertEqual(queue.revision, 1)
    }

    func testDueForegroundTimerRetriesSameIntentWithoutUserTap() async throws {
        let store = try store(), sessions = InMemorySessionStore(), clock = QueueTestClock(), sleeper = QueueTestSleeper()
        addTeardownBlock { await sleeper.releaseAll() }
        try sessions.save(session("owner-a"))
        let api = QueueBoundary(store: store, enabled: true)
        await api.failNext(.http(status: 503))
        let queue = FinancialMutationQueue(store: store, api: api, sessions: sessions, now: { clock.now() }, sleep: { await sleeper.sleep($0) })
        try await queue.activate(ownerID: "owner-a")
        let key = UUID()
        _ = try await queue.submit(payload(), key: key)
        let deadline = Date().addingTimeInterval(2)
        while await sleeper.delays.isEmpty && Date() < deadline { await Task.yield() }
        let delays = await sleeper.delays
        XCTAssertFalse(delays.isEmpty, "A transient failure must schedule a foreground wake")
        if let delay = delays.first { XCTAssertGreaterThanOrEqual(delay, 1); XCTAssertLessThanOrEqual(delay, 1.25) }
        clock.advance(900); await sleeper.releaseAll()
        let appliedDeadline = Date().addingTimeInterval(2)
        while queue.revision == 0 && Date() < appliedDeadline { await Task.yield() }
        let rows = try await store.items(accountID: "owner-a")
        XCTAssertEqual(rows.first?.state, .applied); XCTAssertEqual(rows.first?.idempotencyKey, key.uuidString)
        XCTAssertEqual(rows.count, 1); XCTAssertEqual(queue.revision, 1)
        try await queue.stop()
    }

    func testManualRetryBecomesDueWithOriginalKeyAndAppliesOnce() async throws {
        let store = try store(), sessions = InMemorySessionStore()
        try sessions.save(session("owner-a"))
        let original = try await store.enqueue(accountID: "owner-a", payload: payload(), key: UUID())
        let leased = try await store.leaseNext(accountID: "owner-a")
        try await store.markOutcome(accountID: "owner-a", id: original.id, leaseID: XCTUnwrap(leased?.leaseID), outcome: .retry(at: Date().addingTimeInterval(900), errorCode: "HTTP_503"))
        let api = QueueBoundary(store: store, enabled: true)
        let queue = FinancialMutationQueue(store: store, api: api, sessions: sessions)
        try await queue.activate(ownerID: "owner-a")
        XCTAssertEqual(queue.items.first?.state, .retryWait)
        try await queue.retryNow(id: original.id)
        let rows = try await store.items(accountID: "owner-a")
        XCTAssertEqual(rows.count, 1); XCTAssertEqual(rows.first?.state, .applied)
        XCTAssertEqual(rows.first?.idempotencyKey, original.idempotencyKey)
        let calls = await api.calls
        XCTAssertEqual(calls, 1); XCTAssertTrue(queue.items.isEmpty)
    }

    func testDiscardRemovesOnlyLocalQueueAndRefusesChangedSession() async throws {
        let store = try store(), sessions = InMemorySessionStore()
        try sessions.save(session("owner-a"))
        let api = QueueBoundary(store: store, enabled: false)
        let queue = FinancialMutationQueue(store: store, api: api, sessions: sessions)
        try await queue.activate(ownerID: "owner-a")
        _ = try await queue.submit(payload(), key: UUID())
        let id = try XCTUnwrap(queue.items.first?.id)
        try await queue.discard(id: id)
        XCTAssertTrue(queue.items.isEmpty)
        let rows = try await store.items(accountID: "owner-a")
        XCTAssertEqual(rows.count, 1); XCTAssertEqual(rows.first?.state, .cancelled)
        let calls = await api.calls
        XCTAssertEqual(calls, 0)
        _ = try await queue.submit(payload(), key: UUID())
        let second = try XCTUnwrap(queue.items.first?.id)
        try sessions.save(session("owner-b"))
        do { try await queue.discard(id: second); XCTFail("Wrong active session cannot discard old owner") }
        catch { XCTAssertEqual(error as? APIError, .accountMismatch) }
    }

    func testConfirmedChangedRejectedDraftCreatesOneNewIntentAndNeverMutatesOriginalKey() async throws {
        let store = try store(), sessions = InMemorySessionStore()
        try sessions.save(session("owner-a"))
        let old = try await store.enqueue(accountID: "owner-a", payload: payload(), key: UUID())
        let leased = try await store.leaseNext(accountID: "owner-a")
        try await store.markOutcome(accountID: "owner-a", id: old.id, leaseID: XCTUnwrap(leased?.leaseID), outcome: .rejected(errorCode: "HTTP_422"))
        let api = QueueBoundary(store: store, enabled: true)
        let queue = FinancialMutationQueue(store: store, api: api, sessions: sessions)
        try await queue.activate(ownerID: "owner-a")
        let key = UUID()
        let first = try await queue.replaceRejected(id: old.id, payload: payload(amount: 222), key: key)
        let replay = try await queue.replaceRejected(id: old.id, payload: payload(amount: 222), key: key)
        XCTAssertEqual(first, .applied); XCTAssertEqual(replay, .applied)
        let rows = try await store.items(accountID: "owner-a")
        XCTAssertEqual(rows.map(\.state), [.cancelled, .applied])
        XCTAssertEqual(rows[0].idempotencyKey, old.idempotencyKey)
        XCTAssertEqual(rows[1].idempotencyKey, key.uuidString)
        let calls = await api.calls
        XCTAssertEqual(calls, 1)
    }
    func testNewAccountCanDrainWhileCancelledOldAccountTransportStillReturnsLate() async throws {
        let store = try store(), sessions = InMemorySessionStore()
        try sessions.save(session("owner-a"))
        let api = QueueBoundary(store: store, enabled: true, holdFirstApply: true)
        let queue = FinancialMutationQueue(store: store, api: api, sessions: sessions)
        try await queue.activate(ownerID: "owner-a")
        let old = Task { try await queue.submit(payload(), key: UUID()) }
        await api.waitForFirstSend()
        try sessions.save(session("owner-b"))
        try await queue.activate(ownerID: "owner-b")
        let result = try await queue.submit(payload(amount: 45600), key: UUID())
        XCTAssertEqual(result, .applied, "Old-owner activity must not suppress the new owner's drain")
        await api.release()
        do { _ = try await old.value; XCTFail("Old form must not publish into new owner") }
        catch { XCTAssertEqual(error as? APIError, .accountMismatch) }
        XCTAssertTrue(queue.items.isEmpty)
        XCTAssertEqual(queue.revision, 1)
        let oldRows = try await store.items(accountID: "owner-a"), newRows = try await store.items(accountID: "owner-b")
        XCTAssertEqual(oldRows.first?.state, .blockedAuth)
        XCTAssertEqual(newRows.first?.state, .applied)
    }
    private func store() throws -> OutboxStore {
        let root = FileManager.default.temporaryDirectory.appendingPathComponent("financial-queue-" + UUID().uuidString)
        addTeardownBlock { if FileManager.default.fileExists(atPath: root.path) { try FileManager.default.removeItem(at: root) } }
        return try OutboxStore(directory: root, serverURL: URL(string: "https://queue.test/api/v1")!, keys: QueueFixtureKeys())
    }
    private func session(_ owner: String) -> AccessSession {
        AccessSession(token: "fixture-token", expiresAt: "2099-01-01T00:00:00Z", refreshToken: "fixture-refresh", refreshExpiresAt: "2099-02-01T00:00:00Z", user: AuthUser(id: owner, name: nil, email: nil, timezone: nil))
    }
    private func payload(amount: Int64 = 12300) -> OutboxPayload {
        .transaction(TransactionInputModel(walletID: "wallet", categoryID: "category", jarID: nil, type: "expense", amount: amount, occurredAt: "2026-09-01T03:00:00Z", note: nil, includedInReports: true, travelEventID: nil))
    }

    func testSubmitPersistsBeforeSendAndSameFormKeyCannotCreateAnotherIntent() async throws {
        let store = try store(), sessions = InMemorySessionStore()
        try sessions.save(session("owner-a"))
        let api = QueueBoundary(store: store, enabled: true)
        let queue = FinancialMutationQueue(store: store, api: api, sessions: sessions)
        try await queue.activate(ownerID: "owner-a")
        let key = UUID()
        let first = try await queue.submit(payload(), key: key)
        let repeated = try await queue.submit(payload(), key: key)
        XCTAssertEqual(first, .applied); XCTAssertEqual(repeated, .applied)
        let rows = try await store.items(accountID: "owner-a")
        XCTAssertEqual(rows.count, 1)
        XCTAssertEqual(rows.first?.state, .applied)
        XCTAssertEqual(rows.first?.idempotencyKey, key.uuidString)
        let persisted = await api.observedPersistedSending, calls = await api.calls
        XCTAssertTrue(persisted); XCTAssertEqual(calls, 1)
        XCTAssertEqual(queue.revision, 1)
        do { _ = try await queue.submit(payload(amount: 999), key: key); XCTFail("Cannot mutate previously queued form key") }
        catch { XCTAssertEqual(error as? OutboxStoreError, .intentConflict) }
    }

    func testUnsupportedServerKeepsDraftQueuedAndPublishedWithoutClaimingConfirmedWrite() async throws {
        let store = try store(), sessions = InMemorySessionStore()
        try sessions.save(session("owner-a"))
        let api = QueueBoundary(store: store, enabled: false)
        let queue = FinancialMutationQueue(store: store, api: api, sessions: sessions)
        try await queue.activate(ownerID: "owner-a")
        let result = try await queue.submit(payload(), key: UUID())
        XCTAssertEqual(result, .queued)
        XCTAssertEqual(queue.items.count, 1)
        XCTAssertEqual(queue.items.first?.state, .pending)
        XCTAssertEqual(queue.result, .unsupported)
        XCTAssertEqual(queue.revision, 0)
        let calls = await api.calls
        XCTAssertEqual(calls, 0)
    }

    func testAccountSwitchHidesOldQueueAndNeverReassignsItsMutation() async throws {
        let store = try store(), sessions = InMemorySessionStore()
        try sessions.save(session("owner-a"))
        let api = QueueBoundary(store: store, enabled: false)
        let queue = FinancialMutationQueue(store: store, api: api, sessions: sessions)
        try await queue.activate(ownerID: "owner-a")
        _ = try await queue.submit(payload(), key: UUID())
        try sessions.save(session("owner-b"))
        do { _ = try await queue.submit(payload(), key: UUID()); XCTFail("Stale form owner cannot enqueue into new account") }
        catch { XCTAssertEqual(error as? APIError, .accountMismatch) }
        try await queue.activate(ownerID: "owner-b")
        XCTAssertTrue(queue.items.isEmpty)
        let original = try await store.items(accountID: "owner-a")
        XCTAssertEqual(original.first?.accountID, "owner-a")
        XCTAssertEqual(original.first?.state, .blockedAuth)
        let other = try await store.items(accountID: "owner-b")
        XCTAssertTrue(other.isEmpty)
    }

    func testLogoutBlocksDurableQueueAndSameOwnerActivationRestoresIt() async throws {
        let store = try store(), sessions = InMemorySessionStore()
        try sessions.save(session("owner-a"))
        let queue = FinancialMutationQueue(store: store, api: QueueBoundary(store: store, enabled: false), sessions: sessions)
        try await queue.activate(ownerID: "owner-a")
        _ = try await queue.submit(payload(), key: UUID())
        sessions.clear()
        try await queue.stop()
        XCTAssertTrue(queue.items.isEmpty)
        let rows = try await store.items(accountID: "owner-a")
        XCTAssertEqual(rows.first?.state, .blockedAuth)
        try sessions.save(session("owner-a"))
        try await queue.activate(ownerID: "owner-a")
        XCTAssertEqual(queue.items.first?.state, .pending)
    }
}
