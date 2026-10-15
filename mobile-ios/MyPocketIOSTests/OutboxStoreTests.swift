import CryptoKit
import Foundation
import Security
import SQLite3
import XCTest
@testable import MyPocketIOS

private struct FixtureOutboxKeys: OutboxKeyProviding {
    let bytes: Data
    init(bytes: Data = Data(repeating: 0x4a, count: 32)) { self.bytes = bytes }
    func key(for namespace: String, createIfMissing: Bool) throws -> Data { bytes }
}

private actor LegacyCategoryRejectionTransport: HTTPTransport {
    private(set) var requests: [URLRequest] = []
    func data(for request: URLRequest) async throws -> (Data, URLResponse) {
        requests.append(request)
        return (Data(#"{"status":400,"code":"TRANSACTION_INVALID","detail":"Category required"}"#.utf8),
            HTTPURLResponse(url: request.url!, statusCode: 400, httpVersion: nil, headerFields: nil)!)
    }
}

private actor OutboxReceiptTransport: HTTPTransport {
    var requests: [URLRequest] = []
    let resources: [TransactionModel]
    let key: String
    let operation: String
    let includeReceipt: Bool
    let timeoutFirstCreate: Bool
    var creates = 0
    init(resources: [TransactionModel], key: String, operation: String = "transaction.create", includeReceipt: Bool = true, timeoutFirstCreate: Bool = false) {
        self.resources = resources; self.key = key; self.operation = operation; self.includeReceipt = includeReceipt
        self.timeoutFirstCreate = timeoutFirstCreate
    }
    func data(for request: URLRequest) async throws -> (Data, URLResponse) {
        requests.append(request)
        if request.url?.path.hasSuffix("/sync/capabilities") == true {
            let data = Data(#"{"data":{"contract_version":1,"enabled":true,"operations":["transaction.create","transfer.create"],"receipt_metadata":"/meta/mutation","idempotency_key_required_for_offline":true,"legacy_unkeyed_online":true,"transaction_create_context":"explicit_travel_v1"},"meta":null}"#.utf8)
            return (data, HTTPURLResponse(url: request.url!, statusCode: 200, httpVersion: nil, headerFields: nil)!)
        }
        creates += 1
        if timeoutFirstCreate && creates == 1 { throw URLError(.timedOut) }
        let encoded = operation == "transfer.create" ? try JSONEncoder.mypocket.encode(resources) : try JSONEncoder.mypocket.encode(resources[0])
        var response: [String: Any] = ["data": try JSONSerialization.jsonObject(with: encoded)]
        if includeReceipt {
            response["meta"] = ["mutation": ["operation": operation, "idempotency_key": key, "status": "applied", "replayed": creates > 1]]
        }
        return (try JSONSerialization.data(withJSONObject: response), HTTPURLResponse(url: request.url!, statusCode: creates == 1 ? 201 : 200, httpVersion: nil, headerFields: nil)!)
    }
}

final class OutboxStoreTests: XCTestCase {
    func testLegacyUncategorizedLeasedIntentRetainsWireKeyAndReceivesRecoverableServerRejection() async throws {
        let input = TransactionInputModel(walletID: "wallet-1", categoryID: nil, jarID: nil,
            type: "expense", amount: 12500, occurredAt: "2026-09-01T03:00:00Z", note: "legacy draft",
            includedInReports: true, travelEventID: nil, autoLinkTravel: false)
        let mutation = sending(.transaction(input))
        let transport = LegacyCategoryRejectionTransport()
        let api = try adapter(transport)
        do { _ = try await api.applyOutbox(mutation); XCTFail("Server rejection must remain a rejection") }
        catch APIError.problem(let problem) { XCTAssertEqual(problem.status, 400) }
        catch { XCTFail("Legacy intent must not become a local storage error: \(error)") }
        let requests = await transport.requests
        XCTAssertEqual(requests.count, 1)
        guard let request = requests.first else { return }
        XCTAssertEqual(request.value(forHTTPHeaderField: "Idempotency-Key"), mutation.idempotencyKey)
        let body = try XCTUnwrap(JSONSerialization.jsonObject(with: XCTUnwrap(request.httpBody)) as? [String: Any])
        XCTAssertNil(body["category_id"])
        XCTAssertEqual(body["note"] as? String, "legacy draft")
        XCTAssertEqual(body["amount"] as? Int, 12500)
    }

    func testUncategorizedOrdinaryDraftCannotBecomeDurableAndValidCategorizedDraftStillCan() async throws {
        let store = try OutboxStore(directory: directory(), serverURL: server, keys: FixtureOutboxKeys())
        for kind in ["income", "expense"] {
            for category in [nil, "", "  ", " category ", "bad\0id"] as [String?] {
                let input = TransactionInputModel(walletID: "wallet-1", categoryID: category, jarID: nil,
                    type: kind, amount: 12500, occurredAt: "2026-09-01T03:00:00Z", note: nil,
                    includedInReports: true, travelEventID: nil)
                do {
                    _ = try await store.enqueue(accountID: "owner-a", payload: .transaction(input), key: UUID())
                    XCTFail("Invalid ordinary category must not become a durable financial intent")
                } catch { XCTAssertEqual(error as? OutboxStoreError, .invalidPayload) }
            }
        }
        var rows = try await store.items(accountID: "owner-a")
        XCTAssertTrue(rows.isEmpty, "Validation must precede any insert")
        _ = try await store.enqueue(accountID: "owner-a", payload: payload(kind: "income"), key: UUID())
        _ = try await store.enqueue(accountID: "owner-a", payload: payload(), key: UUID())
        rows = try await store.items(accountID: "owner-a")
        XCTAssertEqual(rows.count, 2)
    }

    func testUncategorizedReplacementCannotCancelRejectedOriginalOrCreateChild() async throws {
        let store = try OutboxStore(directory: directory(), serverURL: server, keys: FixtureOutboxKeys())
        let original = try await store.enqueue(accountID: "owner-a", payload: payload(), key: UUID())
        let leased = try await store.leaseNext(accountID: "owner-a")
        try await store.markOutcome(accountID: "owner-a", id: original.id, leaseID: XCTUnwrap(leased?.leaseID), outcome: .rejected(errorCode: "HTTP_422"))
        let input = TransactionInputModel(walletID: "wallet-1", categoryID: nil, jarID: nil,
            type: "expense", amount: 222, occurredAt: "2026-09-01T03:00:00Z", note: nil,
            includedInReports: true, travelEventID: nil)
        do {
            _ = try await store.replaceRejected(accountID: "owner-a", id: original.id, payload: .transaction(input), key: UUID())
            XCTFail("Invalid category cannot replace an existing financial intent")
        } catch { XCTAssertEqual(error as? OutboxStoreError, .invalidPayload) }
        let rows = try await store.items(accountID: "owner-a")
        XCTAssertEqual(rows.count, 1); XCTAssertEqual(rows.first?.state, .rejected)
        XCTAssertEqual(rows.first?.idempotencyKey, original.idempotencyKey)
    }

    func testTimezoneEquivalentDraftIsNotAChangedReplacementIntent() async throws {
        let store = try OutboxStore(directory: directory(), serverURL: server, keys: FixtureOutboxKeys())
        let original = try await store.enqueue(accountID: "owner-a", payload: payload(), key: UUID())
        let leased = try await store.leaseNext(accountID: "owner-a")
        try await store.markOutcome(accountID: "owner-a", id: original.id, leaseID: XCTUnwrap(leased?.leaseID), outcome: .rejected(errorCode: "HTTP_422"))
        do {
            _ = try await store.replaceRejected(accountID: "owner-a", id: original.id, payload: payload(date: "2026-09-01T10:00:00+07:00"), key: UUID())
            XCTFail("Editor timezone formatting alone is not a user-edited intent")
        } catch { XCTAssertEqual(error as? OutboxStoreError, .invalidPayload) }
        let rows = try await store.items(accountID: "owner-a")
        XCTAssertEqual(rows.count, 1); XCTAssertEqual(rows[0].state, .rejected)
    }
    func testFailedReplacementInsertRollsBackCancellationAndConcurrentReplacementCreatesOneChild() async throws {
        let root = try directory()
        let store = try OutboxStore(directory: root, serverURL: server, keys: FixtureOutboxKeys())
        let other = try OutboxStore(directory: root, serverURL: server, keys: FixtureOutboxKeys())
        let original = try await store.enqueue(accountID: "owner-a", payload: payload(), key: UUID())
        let leased = try await store.leaseNext(accountID: "owner-a")
        try await store.markOutcome(accountID: "owner-a", id: original.id, leaseID: XCTUnwrap(leased?.leaseID), outcome: .rejected(errorCode: "HTTP_400"))
        try executeSQL("CREATE TRIGGER reject_replacement BEFORE INSERT ON outbox BEGIN SELECT RAISE(ABORT, 'fixture_insert_failure'); END", root: root)
        do { _ = try await store.replaceRejected(accountID: "owner-a", id: original.id, payload: payload(amount: 222), key: UUID()); XCTFail("Insert failure must fail transaction") } catch {}
        var rows = try await store.items(accountID: "owner-a")
        XCTAssertEqual(rows.count, 1); XCTAssertEqual(rows[0].state, .rejected)
        try executeSQL("DROP TRIGGER reject_replacement", root: root)
        let key = UUID(), replacementPayload = payload(amount: 222)
        async let first = store.replaceRejected(accountID: "owner-a", id: original.id, payload: replacementPayload, key: key)
        async let second = other.replaceRejected(accountID: "owner-a", id: original.id, payload: replacementPayload, key: key)
        let results = try await (first, second)
        XCTAssertEqual(results.0.id, results.1.id)
        rows = try await store.items(accountID: "owner-a")
        XCTAssertEqual(rows.map(\.state), [.cancelled, .pending])
    }
    func testRetryNowPreservesKeyPayloadAndAttemptAndRefusesOtherStates() async throws {
        let store = try OutboxStore(directory: directory(), serverURL: server, keys: FixtureOutboxKeys())
        let original = try await store.enqueue(accountID: "owner-a", payload: payload(), key: UUID())
        do { try await store.retryNow(accountID: "owner-a", id: original.id); XCTFail("Pending is not a failed attempt") } catch {}
        let leased = try await store.leaseNext(accountID: "owner-a")
        let token = try XCTUnwrap(leased?.leaseID)
        do { try await store.retryNow(accountID: "owner-a", id: original.id); XCTFail("Cannot overwrite active sending lease") } catch {}
        try await store.markOutcome(accountID: "owner-a", id: original.id, leaseID: token, outcome: .retry(at: Date().addingTimeInterval(900), errorCode: "NETWORK_UNAVAILABLE"))
        let now = Date()
        try await store.retryNow(accountID: "owner-a", id: original.id, now: now)
        let rows = try await store.items(accountID: "owner-a")
        XCTAssertEqual(rows.first?.nextAttemptAt, now)
        XCTAssertEqual(rows.first?.idempotencyKey, original.idempotencyKey)
        XCTAssertEqual(rows.first?.attemptCount, 1)
        XCTAssertEqual(try XCTUnwrap(rows.first).payload.canonicalData(), try original.payload.canonicalData())
        do { try await store.retryNow(accountID: "owner-b", id: original.id); XCTFail("Foreign owner retry") } catch {}
    }

    func testDiscardIsLocalOnlyIdempotentAndCannotEraseSendingOrAppliedReceipt() async throws {
        let store = try OutboxStore(directory: directory(), serverURL: server, keys: FixtureOutboxKeys())
        let first = try await store.enqueue(accountID: "owner-a", payload: payload(), key: UUID())
        let leased = try await store.leaseNext(accountID: "owner-a")
        let token = try XCTUnwrap(leased?.leaseID)
        do { try await store.discard(accountID: "owner-a", id: first.id); XCTFail("Cannot discard in-flight write") } catch {}
        try await store.markOutcome(accountID: "owner-a", id: first.id, leaseID: token, outcome: .applied(resourceIDs: ["server-existing"]))
        do { try await store.discard(accountID: "owner-a", id: first.id); XCTFail("Cannot cancel acknowledged write") } catch {}
        let pending = try await store.enqueue(accountID: "owner-a", payload: payload(amount: 999), key: UUID())
        try await store.discard(accountID: "owner-a", id: pending.id)
        try await store.discard(accountID: "owner-a", id: pending.id)
        let rows = try await store.items(accountID: "owner-a")
        XCTAssertEqual(rows.count, 2)
        XCTAssertEqual(rows[0].state, .applied); XCTAssertEqual(rows[0].serverResourceIDs, ["server-existing"])
        XCTAssertEqual(rows[1].state, .cancelled); XCTAssertEqual(rows[1].idempotencyKey, pending.idempotencyKey)
        do { try await store.discard(accountID: "owner-b", id: pending.id); XCTFail("Foreign discard") } catch {}
    }

    func testRejectedReplacementAtomicallyCancelsOldIntentAndKeepsStableNewKey() async throws {
        let store = try OutboxStore(directory: directory(), serverURL: server, keys: FixtureOutboxKeys())
        let first = try await store.enqueue(accountID: "owner-a", payload: payload(), key: UUID())
        let leased = try await store.leaseNext(accountID: "owner-a")
        try await store.markOutcome(accountID: "owner-a", id: first.id, leaseID: XCTUnwrap(leased?.leaseID), outcome: .rejected(errorCode: "HTTP_422"))
        let key = UUID()
        do { _ = try await store.replaceRejected(accountID: "owner-a", id: first.id, payload: payload(), key: key); XCTFail("Unchanged draft cannot get a new key") } catch {}
        let replacement = try await store.replaceRejected(accountID: "owner-a", id: first.id, payload: payload(amount: 222), key: key)
        let repeated = try await store.replaceRejected(accountID: "owner-a", id: first.id, payload: payload(amount: 222), key: key)
        XCTAssertEqual(repeated.id, replacement.id)
        let rows = try await store.items(accountID: "owner-a")
        XCTAssertEqual(rows.map(\.state), [.cancelled, .pending])
        XCTAssertEqual(rows[1].supersedesID, first.id)
        XCTAssertEqual(rows[1].idempotencyKey, key.uuidString)
        XCTAssertNotEqual(rows[0].idempotencyKey, key.uuidString)
        do { _ = try await store.replaceRejected(accountID: "owner-a", id: first.id, payload: payload(amount: 333), key: key); XCTFail("Cannot mutate replacement key") } catch {}
        do { _ = try await store.replaceRejected(accountID: "owner-b", id: first.id, payload: payload(amount: 222), key: UUID()); XCTFail("Foreign replacement") } catch {}
    }

    func testUncertainRetryCannotBeReplacedAndInvalidReplacementPreservesRejectedOriginal() async throws {
        let store = try OutboxStore(directory: directory(), serverURL: server, keys: FixtureOutboxKeys())
        let first = try await store.enqueue(accountID: "owner-a", payload: payload(), key: UUID())
        let leased = try await store.leaseNext(accountID: "owner-a")
        let token = try XCTUnwrap(leased?.leaseID)
        try await store.markOutcome(accountID: "owner-a", id: first.id, leaseID: token, outcome: .retry(at: Date(), errorCode: "ACK_UNVERIFIED"))
        do { _ = try await store.replaceRejected(accountID: "owner-a", id: first.id, payload: payload(amount: 222), key: UUID()); XCTFail("Uncertain result must use same key") } catch {}
        let retried = try await store.leaseNext(accountID: "owner-a")
        try await store.markOutcome(accountID: "owner-a", id: first.id, leaseID: XCTUnwrap(retried?.leaseID), outcome: .conflict(errorCode: "HTTP_409"))
        do { _ = try await store.replaceRejected(accountID: "owner-a", id: first.id, payload: payload(amount: 0), key: UUID()); XCTFail("Invalid replacement") } catch {}
        let rows = try await store.items(accountID: "owner-a")
        XCTAssertEqual(rows.count, 1); XCTAssertEqual(rows[0].state, .conflict)
    }
    func testSyncEngineThroughRealHTTPAdapterReplaysTimeoutAndPersistsAcknowledgedIDs() async throws {
        let store = try OutboxStore(directory: directory(), serverURL: server, keys: FixtureOutboxKeys())
        let original = try await store.enqueue(accountID: "owner-a", payload: payload(), key: UUID())
        var row = transaction(); row.jarID = nil; row.travelEventID = nil
        let transport = OutboxReceiptTransport(resources: [row], key: original.idempotencyKey, timeoutFirstCreate: true)
        let api = try adapter(transport)
        let sync = SyncEngine(store: store, api: api, owner: { "owner-a" }, jitter: { 0 })
        let first = try await sync.trigger(accountID: "owner-a", reason: .foreground)
        XCTAssertEqual(first, .waiting)
        let waiting = try await store.items(accountID: "owner-a")
        XCTAssertEqual(waiting.first?.state, .retryWait)
        let due = try XCTUnwrap(waiting.first?.nextAttemptAt)
        // Clock advances only the scheduler; HTTP lease expiry still uses real current time.
        let resumed = SyncEngine(store: store, api: api, owner: { "owner-a" }, now: { due }, jitter: { 0 })
        let result = try await resumed.trigger(accountID: "owner-a", reason: .network)
        XCTAssertEqual(result, .drained)
        let applied = try await store.items(accountID: "owner-a")
        XCTAssertEqual(applied.first?.state, .applied)
        XCTAssertEqual(applied.first?.serverResourceIDs, ["tx-1"])
        let requests = await transport.requests
        XCTAssertEqual(requests.map { $0.url?.path }, ["/api/v1/sync/capabilities", "/api/v1/transactions", "/api/v1/sync/capabilities", "/api/v1/transactions"])
        XCTAssertEqual(requests[1].httpBody, requests[3].httpBody)
        XCTAssertEqual(requests[1].value(forHTTPHeaderField: "Idempotency-Key"), original.idempotencyKey)
        XCTAssertEqual(requests[3].value(forHTTPHeaderField: "Idempotency-Key"), original.idempotencyKey)
    }
    func testHTTPAdapterUsesPersistedLeaseIntentAndOnlyAcknowledgesMatchingReceipt() async throws {
        let store = try OutboxStore(directory: directory(), serverURL: server, keys: FixtureOutboxKeys())
        _ = try await store.enqueue(accountID: "owner-a", payload: payload(), key: UUID())
        let next = try await store.leaseNext(accountID: "owner-a")
        let lease = try XCTUnwrap(next)
        var row = transaction(); row.jarID = nil; row.travelEventID = nil
        let transport = OutboxReceiptTransport(resources: [row], key: lease.idempotencyKey)
        let api = try adapter(transport)
        let first = try await api.applyOutbox(lease)
        let replay = try await api.applyOutbox(lease)
        XCTAssertEqual(first, ["tx-1"]); XCTAssertEqual(replay, ["tx-1"])
        let requests = await transport.requests
        XCTAssertEqual(requests.count, 2)
        XCTAssertEqual(requests.map { $0.url?.path }, ["/api/v1/transactions", "/api/v1/transactions"])
        XCTAssertEqual(requests[0].httpBody, requests[1].httpBody)
        XCTAssertEqual(requests[0].value(forHTTPHeaderField: "Idempotency-Key"), lease.idempotencyKey)
        let object = try XCTUnwrap(JSONSerialization.jsonObject(with: XCTUnwrap(requests[0].httpBody)) as? [String: Any])
        XCTAssertEqual(requests[0].httpBody, try JSONSerialization.data(withJSONObject: object, options: [.sortedKeys]), "Persisted intent wire encoding must be deterministic, not Dictionary iteration order")
        XCTAssertEqual(object["auto_link_travel"] as? Bool, false)
        XCTAssertEqual(object["occurred_at"] as? String, "2026-09-01T03:00:00Z")
        XCTAssertNil(object["owner_id"])
        // HTTP acknowledgement does not mutate SQLite: only the engine may CAS the lease.
        let rows = try await store.items(accountID: "owner-a")
        XCTAssertEqual(rows.first?.state, .sending)
        let missing = try adapter(OutboxReceiptTransport(resources: [row], key: lease.idempotencyKey, includeReceipt: false))
        do { _ = try await missing.applyOutbox(lease); XCTFail("Bare 201 cannot mark applied") }
        catch { XCTAssertEqual(error as? OutboxAcknowledgementError, .invalidReceipt) }
    }

    func testHTTPAdapterRefusesUnleasedOrUnfrozenIntentBeforeTransport() async throws {
        let transport = OutboxReceiptTransport(resources: [transaction()], key: "saved-intent-key")
        let api = try adapter(transport)
        var mutation = sending(payload())
        do { _ = try await api.applyOutbox(mutation); XCTFail("Unfrozen ordinary intent cannot be sent") } catch {}
        mutation = sending(payload().frozenForOutbox()); mutation.state = .pending
        do { _ = try await api.applyOutbox(mutation); XCTFail("Unleased intent cannot be sent") } catch {}
        mutation = sending(payload().frozenForOutbox()); mutation.leaseExpiresAt = Date(timeIntervalSince1970: 1)
        do { _ = try await api.applyOutbox(mutation); XCTFail("Expired lease cannot be sent") } catch {}
        let count = await transport.requests.count
        XCTAssertEqual(count, 0)
    }

    func testHTTPAdapterValidatesTransferArrayFromPersistedIntent() async throws {
        let input = TransferInputModel(sourceWalletID: "wallet-1", destinationWalletID: "wallet-2", amount: 12500, occurredAt: "2026-09-01T03:00:00Z", note: nil)
        let mutation = sending(.transfer(input))
        var source = transaction(); source.transferID = "pair"
        var destination = transaction(id: "tx-2"); destination.walletID = "wallet-2"; destination.type = "income"; destination.transferID = "pair"
        let transport = OutboxReceiptTransport(resources: [destination, source], key: mutation.idempotencyKey, operation: "transfer.create")
        let ids = try await adapter(transport).applyOutbox(mutation)
        XCTAssertEqual(ids, ["tx-1", "tx-2"])
        let requests = await transport.requests
        XCTAssertEqual(requests.first?.url?.path, "/api/v1/transactions/transfer")
        XCTAssertEqual(requests.first?.value(forHTTPHeaderField: "Idempotency-Key"), "saved-intent-key")
    }

    private func adapter(_ transport: any HTTPTransport) throws -> FinanceAPI {
        let session = AccessSession(token: "fixture-token", expiresAt: "2099-01-01T00:00:00Z", refreshToken: "fixture-refresh", refreshExpiresAt: "2099-02-01T00:00:00Z", user: AuthUser(id: "owner-a", name: nil, email: nil, timezone: nil))
        let sessions = InMemorySessionStore()
        try sessions.save(session)
        return FinanceAPI(client: APIClient(baseURL: server, transport: transport, sessionStore: sessions))
    }
    // A missing/mismatched receipt or changed financial row must never acknowledge an intent.
    func testOrdinaryAcknowledgementRequiresExactReceiptAndFrozenResource() throws {
        let input = TransactionInputModel(walletID: "wallet-1", categoryID: "category-1", jarID: "jar-1", type: "expense", amount: 12500,
            occurredAt: "2026-09-01T10:00:00+07:00", note: " note ", includedInReports: false, travelEventID: "trip-1", autoLinkTravel: false)
        let mutation = sending(.transaction(input))
        let row = transaction()
        XCTAssertEqual(try OutboxAcknowledgement.validate(mutation, resources: [row], meta: receipt()), ["tx-1"])
        XCTAssertEqual(try OutboxAcknowledgement.validate(mutation, resources: [row], meta: receipt(replayed: true)), ["tx-1"])
        for meta in [nil, .null, receipt(key: "other-key"), receipt(operation: "transfer.create"), receipt(status: "pending"), receipt(replayedValue: .string("false"))] as [JSONValue?] {
            XCTAssertThrowsError(try OutboxAcknowledgement.validate(mutation, resources: [row], meta: meta))
        }
        var invalidRows: [TransactionModel] = []
        var wrong = row; wrong.amount += 1; invalidRows.append(wrong)
        wrong = row; wrong.walletID = "other"; invalidRows.append(wrong)
        wrong = row; wrong.type = "income"; invalidRows.append(wrong)
        wrong = row; wrong.categoryID = nil; invalidRows.append(wrong)
        wrong = row; wrong.jarID = nil; invalidRows.append(wrong)
        wrong = row; wrong.travelEventID = nil; invalidRows.append(wrong)
        wrong = row; wrong.includedInReports = true; invalidRows.append(wrong)
        wrong = row; wrong.occurredAt = "2026-09-02T03:00:00Z"; invalidRows.append(wrong)
        wrong = row; wrong.transferID = "pair"; invalidRows.append(wrong)
        invalidRows.append(transaction(owner: "other-owner"))
        invalidRows.append(transaction(id: ""))
        for invalid in invalidRows { XCTAssertThrowsError(try OutboxAcknowledgement.validate(mutation, resources: [invalid], meta: receipt())) }
        XCTAssertThrowsError(try OutboxAcknowledgement.validate(mutation, resources: [], meta: receipt()))
        XCTAssertThrowsError(try OutboxAcknowledgement.validate(mutation, resources: [row, row], meta: receipt()))
        var unleased = mutation; unleased.state = .pending
        XCTAssertThrowsError(try OutboxAcknowledgement.validate(unleased, resources: [row], meta: receipt()))
    }

    func testTransferAcknowledgementRequiresAnAtomicExcludedPairRegardlessOfArrayOrder() throws {
        let mutation = sending(.transfer(TransferInputModel(sourceWalletID: "wallet-1", destinationWalletID: "wallet-2", amount: 12500,
            occurredAt: "2026-09-01T03:00:00Z", note: nil)))
        var source = transaction(); source.categoryID = nil; source.jarID = nil; source.travelEventID = nil; source.transferID = "pair-1"
        var destination = transaction(id: "tx-2"); destination.walletID = "wallet-2"; destination.type = "income"
        destination.categoryID = nil; destination.jarID = nil; destination.travelEventID = nil; destination.transferID = "pair-1"
        let meta = receipt(operation: "transfer.create")
        XCTAssertEqual(try OutboxAcknowledgement.validate(mutation, resources: [destination, source], meta: meta), ["tx-1", "tx-2"])
        for rows in [[source], [source, source], [source, destination, destination]] {
            XCTAssertThrowsError(try OutboxAcknowledgement.validate(mutation, resources: rows, meta: meta))
        }
        var invalid = destination; invalid.transferID = "pair-2"
        XCTAssertThrowsError(try OutboxAcknowledgement.validate(mutation, resources: [source, invalid], meta: meta))
        invalid = destination; invalid.includedInReports = true
        XCTAssertThrowsError(try OutboxAcknowledgement.validate(mutation, resources: [source, invalid], meta: meta))
        invalid = destination; invalid.amount = 1
        XCTAssertThrowsError(try OutboxAcknowledgement.validate(mutation, resources: [source, invalid], meta: meta))
    }

    private func sending(_ payload: OutboxPayload) -> OutboxMutation {
        OutboxMutation(id: UUID(), accountID: "owner-a", idempotencyKey: "saved-intent-key", payload: payload, createdAt: Date(),
            state: .sending, attemptCount: 1, leaseID: UUID(), leaseExpiresAt: Date().addingTimeInterval(60))
    }
    private func transaction(id: String = "tx-1", owner: String = "owner-a") -> TransactionModel {
        TransactionModel(id: id, ownerID: owner, walletID: "wallet-1", categoryID: "category-1", jarID: "jar-1", type: "expense", amount: 12500,
            travelEventID: "trip-1", occurredAt: "2026-09-01T03:00:00.000Z", note: "note", includedInReports: false,
            createdAt: "2026-09-01T03:00:00Z", updatedAt: "2026-09-01T03:00:00Z")
    }
    private func receipt(key: String = "saved-intent-key", operation: String = "transaction.create", status: String = "applied", replayed: Bool = false, replayedValue: JSONValue? = nil) -> JSONValue {
        .object(["mutation": .object(["idempotency_key": .string(key), "operation": .string(operation), "status": .string(status), "replayed": replayedValue ?? .bool(replayed)])])
    }
    func testQueuedOrdinaryIntentFreezesNoTripOrSelectedTripBeforeSending() async throws {
        let store = try OutboxStore(directory: directory(), serverURL: server, keys: FixtureOutboxKeys())
        for trip in [nil, "original-trip"] as [String?] {
            var input = TransactionInputModel(walletID: "wallet", categoryID: "category", jarID: nil, type: "expense", amount: 100,
                occurredAt: "2026-09-01T03:00:00Z", note: nil, includedInReports: true, travelEventID: trip)
            input.note = "entry-time-context"
            let key = UUID()
            let queued = try await store.enqueue(accountID: "owner-a", payload: .transaction(input), key: key)
            guard case .transaction(let frozen) = queued.payload else { return XCTFail("Wrong operation") }
            let object = try XCTUnwrap(JSONSerialization.jsonObject(with: JSONEncoder.mypocket.encode(frozen)) as? [String: Any])
            XCTAssertEqual(object["auto_link_travel"] as? Bool, false)
            XCTAssertEqual(frozen.travelEventID, trip)
            let repeated = try await store.enqueue(accountID: "owner-a", payload: .transaction(input), key: key)
            XCTAssertEqual(repeated.id, queued.id)
        }
    }
    private let server = URL(string: "https://example.test/api/v1")!
    private func directory() throws -> URL {
        let url = FileManager.default.temporaryDirectory.appendingPathComponent("outbox-tests-" + UUID().uuidString, isDirectory: true)
        try FileManager.default.createDirectory(at: url, withIntermediateDirectories: true)
        addTeardownBlock { try FileManager.default.removeItem(at: url) }
        return url
    }
    private func payload(amount: Int64 = 12500, kind: String = "expense", date: String = "2026-09-01T03:00:00Z") -> OutboxPayload {
        .transaction(TransactionInputModel(walletID: "wallet-1", categoryID: "category-1", jarID: nil, type: kind, amount: amount, occurredAt: date, note: "synthetic-private-note-DO-NOT-STORE-PLAINTEXT", includedInReports: false, travelEventID: nil))
    }

    func testEnqueuePersistsEncryptedIntentAndStableKeyAcrossRestart() async throws {
        let root = try directory(), key = UUID()
        let store = try OutboxStore(directory: root, serverURL: server, keys: FixtureOutboxKeys())
        let first = try await store.enqueue(accountID: "owner-a", payload: payload(), key: key)
        let restarted = try OutboxStore(directory: root, serverURL: server, keys: FixtureOutboxKeys())
        let rows = try await restarted.items(accountID: "owner-a")
        XCTAssertEqual(rows.count, 1)
        XCTAssertEqual(rows.first?.id, first.id)
        XCTAssertEqual(rows.first?.idempotencyKey, key.uuidString)
        XCTAssertEqual(rows.first?.state, .pending)
        let count = try await restarted.pendingCount(accountID: "owner-a")
        XCTAssertEqual(count, 1)
        try assertEncryptedFiles(in: root, key: key)
    }

    private func assertEncryptedFiles(in root: URL, key: UUID) throws {
        let files = try XCTUnwrap(FileManager.default.enumerator(at: root, includingPropertiesForKeys: [.isRegularFileKey]))
        var seenFile = false
        for case let url as URL in files {
            guard try url.resourceValues(forKeys: [.isRegularFileKey]).isRegularFile == true else { continue }
            seenFile = true
            let data = try Data(contentsOf: url)
            XCTAssertNil(data.range(of: Data("synthetic-private-note-DO-NOT-STORE-PLAINTEXT".utf8)))
            XCTAssertNil(data.range(of: Data("owner-a".utf8)))
            XCTAssertNil(data.range(of: Data(key.uuidString.utf8)))
        }
        XCTAssertTrue(seenFile, "The intent must actually be persisted before a network send could begin.")
    }

    func testOwnerServerIsolationFIFOAndSameKeyConflict() async throws {
        let root = try directory(), key = UUID()
        let store = try OutboxStore(directory: root, serverURL: server, keys: FixtureOutboxKeys())
        let first = try await store.enqueue(accountID: "owner-a", payload: payload(), key: key)
        let second = try await store.enqueue(accountID: "owner-a", payload: payload(amount: 23000), key: UUID())
        let repeated = try await store.enqueue(accountID: "owner-a", payload: payload(), key: key)
        XCTAssertEqual(first.id, repeated.id)
        do { _ = try await store.enqueue(accountID: "owner-a", payload: payload(amount: 999), key: key); XCTFail("Changed intent must not overwrite the original") }
        catch { XCTAssertEqual(error as? OutboxStoreError, .intentConflict) }
        let other = try await store.enqueue(accountID: "owner-b", payload: payload(), key: key)
        XCTAssertNotEqual(first.id, other.id)
        let rows = try await store.items(accountID: "owner-a")
        XCTAssertEqual(rows.map(\.id), [first.id, second.id])
        let differentServer = try OutboxStore(directory: root, serverURL: URL(string: "https://other.test/api/v1")!, keys: FixtureOutboxKeys())
        let foreignRows = try await differentServer.items(accountID: "owner-a")
        XCTAssertTrue(foreignRows.isEmpty)
    }

    func testWrongEncryptionKeyRefusesReadWithoutOverwritingOriginalIntent() async throws {
        let root = try directory()
        let store = try OutboxStore(directory: root, serverURL: server, keys: FixtureOutboxKeys())
        let first = try await store.enqueue(accountID: "owner-a", payload: payload(), key: UUID())
        let wrong = try OutboxStore(directory: root, serverURL: server, keys: FixtureOutboxKeys(bytes: Data(repeating: 0x7b, count: 32)))
        do { _ = try await wrong.items(accountID: "owner-a"); XCTFail("Wrong key must fail, not return empty or delete data") }
        catch { XCTAssertEqual(error as? OutboxStoreError, .corruptStore) }
        let intact = try await store.items(accountID: "owner-a")
        XCTAssertEqual(intact.map(\.id), [first.id])
    }

    func testTwoStoresCommitOnlyOneLocalIntentForSameKey() async throws {
        let root = try directory(), key = UUID(), input = payload()
        let one = try OutboxStore(directory: root, serverURL: server, keys: FixtureOutboxKeys())
        let two = try OutboxStore(directory: root, serverURL: server, keys: FixtureOutboxKeys())
        async let first = one.enqueue(accountID: "owner-a", payload: input, key: key)
        async let second = two.enqueue(accountID: "owner-a", payload: input, key: key)
        let (a, b) = try await (first, second)
        XCTAssertEqual(a.id, b.id)
        let count = try await one.pendingCount(accountID: "owner-a")
        XCTAssertEqual(count, 1)
    }

    func testInvalidFinancialPayloadDoesNotCreatePendingIntent() async throws {
        let store = try OutboxStore(directory: directory(), serverURL: server, keys: FixtureOutboxKeys())
        for input in [payload(amount: 0), payload(kind: "adjustment"), payload(date: "2026-09-01")] {
            do { _ = try await store.enqueue(accountID: "owner-a", payload: input, key: UUID()); XCTFail("Invalid/unsupported financial payload accepted") }
            catch { XCTAssertEqual(error as? OutboxStoreError, .invalidPayload) }
        }
        let rows = try await store.items(accountID: "owner-a")
        XCTAssertTrue(rows.isEmpty)
    }

    func testKeychainOutboxKeyIsDeviceOnlyStableAndNotRecreatedWhenMissing() throws {
        let service = "test.mypocket.outbox." + UUID().uuidString
        let query: [String: Any] = [kSecClass as String: kSecClassGenericPassword, kSecAttrService as String: service]
        addTeardownBlock {
            let cleanup: [String: Any] = [kSecClass as String: kSecClassGenericPassword, kSecAttrService as String: service]
            SecItemDelete(cleanup as CFDictionary)
        }
        let keys = KeychainOutboxKeyProvider(service: service)
        let first = try keys.key(for: "owner-namespace", createIfMissing: true)
        XCTAssertEqual(first.count, 32)
        XCTAssertEqual(try keys.key(for: "owner-namespace", createIfMissing: false), first)
        var attributesQuery = query
        attributesQuery[kSecAttrAccount as String] = "owner-namespace"
        attributesQuery[kSecReturnAttributes as String] = true
        var result: CFTypeRef?
        XCTAssertEqual(SecItemCopyMatching(attributesQuery as CFDictionary, &result), errSecSuccess)
        let attributes = try XCTUnwrap(result as? [String: Any])
        XCTAssertEqual(attributes[kSecAttrAccessible as String] as? String, kSecAttrAccessibleAfterFirstUnlockThisDeviceOnly as String)
        XCTAssertNotEqual(attributes[kSecAttrSynchronizable as String] as? Bool, true)
        XCTAssertEqual(SecItemDelete(query as CFDictionary), errSecSuccess)
        XCTAssertThrowsError(try keys.key(for: "owner-namespace", createIfMissing: false))
    }

    func testMissingKeyWithExistingRowsDoesNotGenerateReplacementOrLoseIntent() async throws {
        let root = try directory(), service = "test.mypocket.outbox." + UUID().uuidString
        defer { deleteKeychainService(service) }
        let keys = KeychainOutboxKeyProvider(service: service)
        let store = try OutboxStore(directory: root, serverURL: server, keys: keys)
        let first = try await store.enqueue(accountID: "owner-a", payload: payload(), key: UUID())
        let original = try keychainRecord(service)
        deleteKeychainService(service)
        do { _ = try await store.items(accountID: "owner-a"); XCTFail("Missing key must not read as an empty queue") }
        catch { XCTAssertEqual(error as? OutboxStoreError, .keyUnavailable) }
        do { _ = try await store.enqueue(accountID: "owner-a", payload: payload(), key: UUID()); XCTFail("Missing key must not be recreated over existing rows") }
        catch { XCTAssertEqual(error as? OutboxStoreError, .keyUnavailable) }
        XCTAssertThrowsError(try keychainRecord(service))
        let restoration: [String: Any] = [kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service, kSecAttrAccount as String: original.account,
            kSecValueData as String: original.bytes, kSecAttrSynchronizable as String: false,
            kSecAttrAccessible as String: kSecAttrAccessibleAfterFirstUnlockThisDeviceOnly]
        XCTAssertEqual(SecItemAdd(restoration as CFDictionary, nil), errSecSuccess)
        let intact = try await store.items(accountID: "owner-a")
        XCTAssertEqual(intact.map(\.id), [first.id])
    }

    func testTransferKeepsFixedDateAndOneAtomicIntentAcrossRestart() async throws {
        let root = try directory(), key = UUID()
        let store = try OutboxStore(directory: root, serverURL: server, keys: FixtureOutboxKeys())
        let input = TransferInputModel(sourceWalletID: "source", destinationWalletID: "destination", amount: 50000, occurredAt: "2026-09-01T03:00:00Z", note: "transfer note")
        let transfer = try await store.enqueue(accountID: "owner-a", payload: .transfer(input), key: key)
        _ = try await store.enqueue(accountID: "owner-a", payload: payload(), key: key)
        let restarted = try OutboxStore(directory: root, serverURL: server.appendingPathComponent(""), keys: FixtureOutboxKeys())
        let rows = try await restarted.items(accountID: "owner-a")
        XCTAssertEqual(rows.count, 2, "Distinct operations may use the same key without merging")
        XCTAssertEqual(rows.first?.id, transfer.id)
        guard case .transfer(let restored) = rows.first?.payload else { return XCTFail("Transfer became another operation") }
        XCTAssertEqual(restored.sourceWalletID, "source")
        XCTAssertEqual(restored.destinationWalletID, "destination")
        XCTAssertEqual(restored.amount, 50000)
        XCTAssertEqual(restored.occurredAt, "2026-09-01T03:00:00Z")
        XCTAssertEqual(restored.note, "transfer note")
    }

    func testInvalidTransferAndAccountCannotEnterOutbox() async throws {
        let store = try OutboxStore(directory: directory(), serverURL: server, keys: FixtureOutboxKeys())
        for owner in ["", " owner-a", "owner-a\0"] {
            do { _ = try await store.enqueue(accountID: owner, payload: payload(), key: UUID()); XCTFail("Invalid account accepted") }
            catch { XCTAssertEqual(error as? OutboxStoreError, .invalidAccount) }
        }
        for input in [
            TransferInputModel(sourceWalletID: "same", destinationWalletID: "same", amount: 100, occurredAt: "2026-09-01T03:00:00Z", note: nil),
            TransferInputModel(sourceWalletID: "source", destinationWalletID: "", amount: 100, occurredAt: "2026-09-01T03:00:00Z", note: nil),
            TransferInputModel(sourceWalletID: "source", destinationWalletID: "destination", amount: -100, occurredAt: "2026-09-01T03:00:00Z", note: nil)
        ] {
            do { _ = try await store.enqueue(accountID: "owner-a", payload: .transfer(input), key: UUID()); XCTFail("Invalid transfer accepted") }
            catch { XCTAssertEqual(error as? OutboxStoreError, .invalidPayload) }
        }
        let rows = try await store.items(accountID: "owner-a")
        XCTAssertTrue(rows.isEmpty)
    }

    func testCiphertextCannotBeMovedBetweenOwnersEvenWithSameKeyBytes() async throws {
        let root = try directory()
        let store = try OutboxStore(directory: root, serverURL: server, keys: FixtureOutboxKeys())
        _ = try await store.enqueue(accountID: "owner-a", payload: payload(), key: UUID())
        _ = try await store.enqueue(accountID: "owner-b", payload: payload(), key: UUID())
        try executeSQL("UPDATE outbox SET sealed=(SELECT sealed FROM outbox ORDER BY ordinal LIMIT 1) WHERE ordinal=(SELECT MAX(ordinal) FROM outbox)", root: root)
        do { _ = try await store.items(accountID: "owner-b"); XCTFail("Copied ciphertext must fail owner/record authentication") }
        catch { XCTAssertEqual(error as? OutboxStoreError, .corruptStore) }
        let unaffected = try await store.items(accountID: "owner-a")
        XCTAssertEqual(unaffected.count, 1)
    }

    func testUnknownSchemaDoesNotResetOrOverwriteOriginalData() async throws {
        let root = try directory()
        let store = try OutboxStore(directory: root, serverURL: server, keys: FixtureOutboxKeys())
        let first = try await store.enqueue(accountID: "owner-a", payload: payload(), key: UUID())
        try executeSQL("PRAGMA user_version=99", root: root)
        XCTAssertThrowsError(try OutboxStore(directory: root, serverURL: server, keys: FixtureOutboxKeys())) {
            XCTAssertEqual($0 as? OutboxStoreError, .unsupportedVersion)
        }
        try executeSQL("PRAGMA user_version=1", root: root)
        let restarted = try OutboxStore(directory: root, serverURL: server, keys: FixtureOutboxKeys())
        let rows = try await restarted.items(accountID: "owner-a")
        XCTAssertEqual(rows.map(\.id), [first.id])
    }

    private func deleteKeychainService(_ service: String) {
        SecItemDelete([kSecClass as String: kSecClassGenericPassword, kSecAttrService as String: service] as CFDictionary)
    }

    func testLeaseIsExclusiveAcrossStoresAndRetryBlocksLaterIntent() async throws {
        let root = try directory(), now = Date(timeIntervalSince1970: 1000)
        let one = try OutboxStore(directory: root, serverURL: server, keys: FixtureOutboxKeys())
        let two = try OutboxStore(directory: root, serverURL: server, keys: FixtureOutboxKeys())
        let first = try await one.enqueue(accountID: "owner-a", payload: payload(), key: UUID())
        _ = try await one.enqueue(accountID: "owner-a", payload: payload(amount: 200), key: UUID())
        async let a = one.leaseNext(accountID: "owner-a", now: now)
        async let b = two.leaseNext(accountID: "owner-a", now: now)
        let candidates = try await [a, b].compactMap { $0 }
        XCTAssertEqual(candidates.count, 1)
        let lease = try XCTUnwrap(candidates.first)
        XCTAssertEqual(lease.id, first.id)
        XCTAssertEqual(lease.state, .sending)
        XCTAssertEqual(lease.attemptCount, 1)
        XCTAssertEqual(lease.leaseExpiresAt, now.addingTimeInterval(60))
        let token = try XCTUnwrap(lease.leaseID)
        try await one.markOutcome(accountID: "owner-a", id: first.id, leaseID: token, outcome: .retry(at: now.addingTimeInterval(120), errorCode: "NETWORK_TIMEOUT"))
        let tooSoon = try await two.leaseNext(accountID: "owner-a", now: now.addingTimeInterval(119))
        XCTAssertNil(tooSoon, "FIFO must not send later intent ahead of delayed first intent")
        let retry = try await two.leaseNext(accountID: "owner-a", now: now.addingTimeInterval(120))
        XCTAssertEqual(retry?.id, first.id)
        XCTAssertEqual(retry?.idempotencyKey, first.idempotencyKey)
        XCTAssertEqual(retry?.attemptCount, 2)
        XCTAssertNotEqual(retry?.leaseID, token)
    }

    func testExpiredLeaseRecoversAcrossRestartAndOldResponseCannotApply() async throws {
        let root = try directory(), now = Date(timeIntervalSince1970: 1000)
        let original = try OutboxStore(directory: root, serverURL: server, keys: FixtureOutboxKeys())
        let input = try await original.enqueue(accountID: "owner-a", payload: payload(), key: UUID())
        let leased = try await original.leaseNext(accountID: "owner-a", now: now)
        let oldToken = try XCTUnwrap(leased?.leaseID)
        let restarted = try OutboxStore(directory: root, serverURL: server, keys: FixtureOutboxKeys())
        try await restarted.resumeLeases(accountID: "owner-a", now: now.addingTimeInterval(59))
        let premature = try await restarted.leaseNext(accountID: "owner-a", now: now.addingTimeInterval(59))
        XCTAssertNil(premature)
        try await restarted.resumeLeases(accountID: "owner-a", now: now.addingTimeInterval(60))
        let resumed = try await restarted.leaseNext(accountID: "owner-a", now: now.addingTimeInterval(60))
        XCTAssertEqual(resumed?.idempotencyKey, input.idempotencyKey)
        XCTAssertEqual(resumed?.attemptCount, 2)
        let currentToken = try XCTUnwrap(resumed?.leaseID)
        XCTAssertNotEqual(currentToken, oldToken)
        do { try await original.markOutcome(accountID: "owner-a", id: input.id, leaseID: oldToken, outcome: .applied(resourceIDs: ["stale"])); XCTFail("Stale sender applied result") }
        catch { XCTAssertEqual(error as? OutboxStoreError, .leaseLost) }
        try await restarted.markOutcome(accountID: "owner-a", id: input.id, leaseID: currentToken, outcome: .applied(resourceIDs: ["server-transaction"]))
        let final = try await original.items(accountID: "owner-a")
        XCTAssertEqual(final.first?.state, .applied)
        XCTAssertEqual(final.first?.serverResourceIDs, ["server-transaction"])
        XCTAssertNil(final.first?.leaseID)
        let count = try await original.pendingCount(accountID: "owner-a")
        XCTAssertEqual(count, 0)
    }

    func testTerminalOutcomesPersistAndAreNotAutomaticallyRetried() async throws {
        let root = try directory(), now = Date(timeIntervalSince1970: 1000)
        let store = try OutboxStore(directory: root, serverURL: server, keys: FixtureOutboxKeys())
        let first = try await store.enqueue(accountID: "owner-a", payload: payload(), key: UUID())
        let second = try await store.enqueue(accountID: "owner-a", payload: payload(amount: 200), key: UUID())
        let conflict = try await store.leaseNext(accountID: "owner-a", now: now)
        try await store.markOutcome(accountID: "owner-a", id: first.id, leaseID: XCTUnwrap(conflict?.leaseID), outcome: .conflict(errorCode: "IDEMPOTENCY_KEY_REUSED"))
        let rejected = try await store.leaseNext(accountID: "owner-a", now: now)
        XCTAssertEqual(rejected?.id, second.id)
        try await store.markOutcome(accountID: "owner-a", id: second.id, leaseID: XCTUnwrap(rejected?.leaseID), outcome: .rejected(errorCode: "WALLET_NOT_FOUND"))
        let restarted = try OutboxStore(directory: root, serverURL: server, keys: FixtureOutboxKeys())
        let next = try await restarted.leaseNext(accountID: "owner-a", now: now.addingTimeInterval(1000))
        XCTAssertNil(next)
        let rows = try await restarted.items(accountID: "owner-a")
        XCTAssertEqual(rows.map(\.state), [.conflict, .rejected])
        XCTAssertEqual(rows.map(\.lastErrorCode), ["IDEMPOTENCY_KEY_REUSED", "WALLET_NOT_FOUND"])
    }

    func testLogoutRevokesLeaseAndOnlySameAccountResumeReleasesQueue() async throws {
        let store = try OutboxStore(directory: directory(), serverURL: server, keys: FixtureOutboxKeys())
        let first = try await store.enqueue(accountID: "owner-a", payload: payload(), key: UUID())
        _ = try await store.enqueue(accountID: "owner-b", payload: payload(), key: UUID())
        let lease = try await store.leaseNext(accountID: "owner-a")
        let token = try XCTUnwrap(lease?.leaseID)
        try await store.blockForAuth(accountID: "owner-a")
        do { try await store.markOutcome(accountID: "owner-a", id: first.id, leaseID: token, outcome: .applied(resourceIDs: ["late"])); XCTFail("Logged-out lease still valid") }
        catch { XCTAssertEqual(error as? OutboxStoreError, .leaseLost) }
        let blocked = try await store.leaseNext(accountID: "owner-a")
        XCTAssertNil(blocked)
        try await store.resumeAuth(accountID: "owner-b")
        let stillBlocked = try await store.leaseNext(accountID: "owner-a")
        XCTAssertNil(stillBlocked)
        let other = try await store.leaseNext(accountID: "owner-b")
        XCTAssertNotNil(other)
        try await store.resumeAuth(accountID: "owner-a")
        let resumed = try await store.leaseNext(accountID: "owner-a")
        XCTAssertEqual(resumed?.idempotencyKey, first.idempotencyKey)
        XCTAssertEqual(resumed?.attemptCount, 2)
    }

    func testForeignOwnerAndInvalidOutcomeCannotOverwriteLease() async throws {
        let store = try OutboxStore(directory: directory(), serverURL: server, keys: FixtureOutboxKeys())
        let first = try await store.enqueue(accountID: "owner-a", payload: payload(), key: UUID())
        let lease = try await store.leaseNext(accountID: "owner-a")
        let token = try XCTUnwrap(lease?.leaseID)
        do { try await store.markOutcome(accountID: "owner-b", id: first.id, leaseID: token, outcome: .applied(resourceIDs: ["foreign"])); XCTFail("Foreign owner applied result") }
        catch { XCTAssertEqual(error as? OutboxStoreError, .leaseLost) }
        for outcome in [OutboxOutcome.applied(resourceIDs: []), .rejected(errorCode: "raw sensitive error text")] {
            do { try await store.markOutcome(accountID: "owner-a", id: first.id, leaseID: token, outcome: outcome); XCTFail("Invalid outcome accepted") }
            catch { XCTAssertEqual(error as? OutboxStoreError, .invalidPayload) }
        }
        let rows = try await store.items(accountID: "owner-a")
        XCTAssertEqual(rows.first?.state, .sending)
        XCTAssertEqual(rows.first?.leaseID, token)
        XCTAssertEqual(rows.first?.serverResourceIDs, [])
        try await store.markOutcome(accountID: "owner-a", id: first.id, leaseID: token, outcome: .applied(resourceIDs: ["safe-id"]))
        do { try await store.markOutcome(accountID: "owner-a", id: first.id, leaseID: token, outcome: .conflict(errorCode: "LATE")); XCTFail("Terminal acknowledgement overwritten") }
        catch { XCTAssertEqual(error as? OutboxStoreError, .leaseLost) }
    }

    private func keychainRecord(_ service: String) throws -> (account: String, bytes: Data) {
        let query: [String: Any] = [kSecClass as String: kSecClassGenericPassword, kSecAttrService as String: service,
                                  kSecReturnAttributes as String: true, kSecReturnData as String: true]
        var result: CFTypeRef?
        guard SecItemCopyMatching(query as CFDictionary, &result) == errSecSuccess,
              let record = result as? [String: Any], let account = record[kSecAttrAccount as String] as? String,
              let bytes = record[kSecValueData as String] as? Data else { throw OutboxStoreError.keyUnavailable }
        return (account, bytes)
    }

    private func executeSQL(_ sql: String, root: URL) throws {
        let files = try XCTUnwrap(FileManager.default.enumerator(at: root, includingPropertiesForKeys: nil))
        let url = try XCTUnwrap(files.allObjects.compactMap { $0 as? URL }.first { $0.lastPathComponent == "outbox.sqlite" })
        var connection: OpaquePointer?
        guard sqlite3_open(url.path, &connection) == SQLITE_OK else { throw OutboxStoreError.storageUnavailable }
        defer { sqlite3_close_v2(connection) }
        guard sqlite3_exec(connection, sql, nil, nil, nil) == SQLITE_OK else { throw OutboxStoreError.storageUnavailable }
    }
}
