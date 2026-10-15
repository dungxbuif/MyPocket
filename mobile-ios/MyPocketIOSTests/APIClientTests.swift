import Foundation
import XCTest
import UIKit
import Darwin
@testable import MyPocketIOS

private struct PrivateErrorTransport: HTTPTransport {
    func data(for request: URLRequest) async throws -> (Data, URLResponse) {
        (Data(#"{"status":422,"detail":"PRIVATE_BODY_CANARY"}"#.utf8), HTTPURLResponse(url: request.url!, statusCode: 422, httpVersion: nil, headerFields: nil)!)
    }
}

private func captureHTTPStandardOutput(_ action: () async throws -> Void) async throws -> String {
    fflush(stdout)
    let saved = dup(STDOUT_FILENO)
    guard saved >= 0 else { throw NSError(domain: "HTTPDiagnosticCapture", code: 1) }
    let pipe = Pipe()
    guard dup2(pipe.fileHandleForWriting.fileDescriptor, STDOUT_FILENO) >= 0 else {
        close(saved)
        throw NSError(domain: "HTTPDiagnosticCapture", code: 2)
    }
    do {
        try await action()
    } catch {
        fflush(stdout)
        dup2(saved, STDOUT_FILENO)
        close(saved)
        pipe.fileHandleForWriting.closeFile()
        throw error
    }
    fflush(stdout)
    dup2(saved, STDOUT_FILENO)
    close(saved)
    pipe.fileHandleForWriting.closeFile()
    return String(decoding: pipe.fileHandleForReading.readDataToEndOfFile(), as: UTF8.self)
}

final class APIClientTests: XCTestCase {
    func testAdvisorHistoryPreservesCursorAndChronologicalStructuredMessages() async throws {
        let transport = AdvisorContractTransport(payloads: [
            try advisorHistoryPayload(sequences: [54, 53, 52]),
            try advisorHistoryPayload(sequences: [51, 50])
        ])
        let client = APIClient(baseURL: URL(string: "https://example.test/api/v1")!,
            transport: transport, sessionStore: TestSessionStore(session: testSession))
        let api = FinanceAPI(client: client)
        let conversationID = "00000000-0000-0000-0000-000000000001"

        let latest = try await api.advisorMessages(conversationID: conversationID)
        XCTAssertEqual(latest.map(\.seq), [52, 53, 54])
        XCTAssertEqual(latest.map(\.id), ["message-52", "message-53", "message-54"])
        XCTAssertTrue(latest.allSatisfy { $0.conversationID == conversationID })
        XCTAssertEqual(latest.last?.parts.map(\.type), ["text", "metric_group"])
        XCTAssertEqual(latest.last?.parts.first?.text, "Tin nhắn 54")
        guard case .object(let card) = latest.last?.parts.last?.data,
              case .object(let view) = card["view"], case .number(let income) = view["income"] else {
            return XCTFail("Persisted structured parts must survive native history decoding")
        }
        XCTAssertEqual(income, 125_000)

        let older = try await api.advisorMessages(conversationID: conversationID, beforeSequence: 52)
        XCTAssertEqual(older.map(\.seq), [50, 51])
        XCTAssertEqual(older.map(\.id), ["message-50", "message-51"])
        let requests = await transport.requests
        XCTAssertEqual(requests.count, 2)
        guard requests.count == 2 else { return }
        for request in requests {
            XCTAssertEqual(request.httpMethod, "GET")
            XCTAssertEqual(request.url?.path, "/api/v1/ai/advisor/conversation/messages")
            XCTAssertEqual(request.value(forHTTPHeaderField: "Authorization"), "Bearer access")
            XCTAssertNil(request.httpBody)
            let query = try XCTUnwrap(URLComponents(url: try XCTUnwrap(request.url), resolvingAgainstBaseURL: false)?.queryItems)
            XCTAssertEqual(query.filter { $0.name == "conversation_id" }.map(\.value), [conversationID])
            XCTAssertFalse(query.contains { $0.name == "owner_id" }, "Owner comes from session, not a caller query")
        }
        let firstQuery = URLComponents(url: try XCTUnwrap(requests[0].url), resolvingAgainstBaseURL: false)?.queryItems ?? []
        XCTAssertFalse(firstQuery.contains { $0.name == "before_seq" })
        let olderQuery = URLComponents(url: try XCTUnwrap(requests[1].url), resolvingAgainstBaseURL: false)?.queryItems ?? []
        XCTAssertEqual(olderQuery.filter { $0.name == "before_seq" }.map(\.value), ["52"])
    }

    func testAdvisorHistoryDecodesServerLifecycleMetadataWithoutChangingArrayContract() async throws {
        let legacyPayload = String(decoding: try advisorHistoryPayload(sequences: [1]), as: UTF8.self)
        let payload = Data(legacyPayload.replacingOccurrences(
            of: #""meta":null"#,
            with: #""meta":{"conversation_id":"00000000-0000-0000-0000-000000000001","generation":4,"last_user_message_at":"2026-10-15T10:00:00Z","expires_at":"2026-10-15T10:30:00Z","server_time":"2026-10-15T10:05:00Z"}"#
        ).utf8)
        let transport = AdvisorContractTransport(payloads: [payload])
        let client = APIClient(baseURL: URL(string: "https://example.test/api/v1")!, transport: transport, sessionStore: TestSessionStore(session: testSession))

        let page = try await FinanceAPI(client: client).advisorHistory(conversationID: "00000000-0000-0000-0000-000000000001")

        XCTAssertEqual(page.messages.map(\.seq), [1])
        XCTAssertEqual(page.meta?.conversationID, "00000000-0000-0000-0000-000000000001")
        XCTAssertEqual(page.meta?.generation, 4)
        XCTAssertEqual(page.meta?.expiresAt, "2026-10-15T10:30:00Z")
        XCTAssertEqual(page.meta?.serverTime, "2026-10-15T10:05:00Z")
    }

    func testAdvisorCancellationUsesExplicitRunEndpointWithoutResubmission() async throws {
        let transport = AdvisorContractTransport(payloads: [Data(#"{"data":{},"meta":null}"#.utf8)])
        let client = APIClient(baseURL: URL(string: "https://example.test/api/v1")!,
            transport: transport, sessionStore: TestSessionStore(session: testSession))
        try await FinanceAPI(client: client).cancelAdvisorRun(id: "run-fixture")
        let requests = await transport.requests
        XCTAssertEqual(requests.count, 1, "Cancellation must not submit another user message")
        let request = try XCTUnwrap(requests.first)
        XCTAssertEqual(request.url?.path, "/api/v1/ai/advisor/runs/run-fixture/cancel")
        XCTAssertEqual(request.httpMethod, "POST")
        XCTAssertEqual(request.value(forHTTPHeaderField: "Authorization"), "Bearer access")
        XCTAssertNotNil(UUID(uuidString: try XCTUnwrap(request.value(forHTTPHeaderField: "Idempotency-Key"))))
        XCTAssertNil(request.httpBody)
    }

    func testAdvisorResetUsesStableIdempotencyKeyAndDoesNotSubmitChatText() async throws {
        let transport = AdvisorContractTransport(payloads: [Data(#"{"data":{"conversation_id":"conversation-1","generation":2,"server_time":"2026-10-15T10:00:00Z","replayed":false},"meta":null}"#.utf8)])
        let client = APIClient(baseURL: URL(string: "https://example.test/api/v1")!,
            transport: transport, sessionStore: TestSessionStore(session: testSession))
        let result = try await FinanceAPI(client: client).resetAdvisorConversation(idempotencyKey: "reset-key-1")
        XCTAssertEqual(result.conversationID, "conversation-1")
        XCTAssertEqual(result.generation, 2)
        let requests = await transport.requests
        let request = try XCTUnwrap(requests.first)
        XCTAssertEqual(request.url?.path, "/api/v1/ai/advisor/conversation/reset")
        XCTAssertEqual(request.httpMethod, "POST")
        XCTAssertEqual(request.value(forHTTPHeaderField: "Idempotency-Key"), "reset-key-1")
        XCTAssertNil(request.httpBody)
    }

    private func advisorHistoryPayload(sequences: [Int]) throws -> Data {
        let rows: [[String: Any]] = sequences.map { seq in
            var parts: [[String: Any]] = [["type": "text", "text": "Tin nhắn \(seq)"]]
            if seq.isMultiple(of: 2) {
                parts.append(["type": "metric_group", "data": [
                    "view": ["income": 125_000, "expense": 25_000, "net": 100_000, "count": 2],
                    "source": ["timezone": "Asia/Ho_Chi_Minh"]
                ]])
            }
            return ["id": "message-\(seq)", "conversation_id": "00000000-0000-0000-0000-000000000001",
                "generation": 1, "seq": seq, "run_id": "run-\(seq)",
                "role": seq.isMultiple(of: 2) ? "assistant" : "user", "parts_version": 1,
                "parts": parts, "created_at": "2026-10-15T00:00:00Z"]
        }
        return try JSONSerialization.data(withJSONObject: ["data": rows, "meta": NSNull()])
    }

    func testFeatureDiagnosticsDoNotPrintPrivateProblemFieldsAndKeepHTTPStatus() async throws {
        let error = APIError.problem(ProblemDetails(title: "PRIVATE_TITLE", status: 422, detail: "PRIVATE_DETAIL", instance: "PRIVATE_INSTANCE", code: "PRIVATE_CODE", requestID: "PRIVATE_REQUEST_ID"))
        let output = try await captureHTTPStandardOutput { print(APIError.diagnosticDescription(error)) }
        XCTAssertFalse(output.contains("PRIVATE_"), "Feature diagnostics must not reflect problem fields")
        XCTAssertTrue(output.contains("422"), "Keep numeric status for diagnosis")
        XCTAssertEqual(error.userMessage, "PRIVATE_DETAIL", "Diagnostic policy must not change user-facing error semantics")
    }

    func testFeatureDiagnosticsDoNotPrintTransportDecodingOrUnknownErrorDetails() async throws {
        let cases: [(Error, String)] = [
            (APIError.transport("PRIVATE_TRANSPORT"), "transport"),
            (APIError.decoding("PRIVATE_DECODING"), "decoding"),
            (NSError(domain: "PRIVATE_DOMAIN", code: 123, userInfo: [NSLocalizedDescriptionKey: "PRIVATE_USER_INFO"]), "unexpected")
        ]
        for (error, category) in cases {
            let output = try await captureHTTPStandardOutput { print(APIError.diagnosticDescription(error)) }
            XCTAssertFalse(output.contains("PRIVATE_"), "No associated descriptions or NSError metadata may enter diagnostics")
            XCTAssertTrue(output.contains(category), "Keep a bounded failure category")
        }
    }

    func testFeatureDiagnosticsClassifyAuthAndHTTPFailuresWithoutAssociatedData() {
        let cases: [(APIError, String)] = [(.invalidURL, "invalidURL"), (.unauthorized, "unauthorized"), (.missingSession, "missingSession"), (.accountMismatch, "accountMismatch"), (.http(status: 503), "503"), (.problem(ProblemDetails(title: nil, status: nil, detail: "PRIVATE_DETAIL", instance: nil, code: nil, requestID: nil)), "problem")]
        for (error, category) in cases {
            let summary = APIError.diagnosticDescription(error)
            XCTAssertFalse(summary.contains("PRIVATE_"))
            XCTAssertTrue(summary.contains(category))
        }
    }

    func testHTTPDiagnosticsDoNotExposeQueryOrErrorPayloadButKeepStatusCorrelation() async throws {
        let client = APIClient(baseURL: URL(string: "https://example.test/api/v1")!, transport: PrivateErrorTransport(), sessionStore: TestSessionStore(session: nil))
        let output = try await captureHTTPStandardOutput {
            do {
                let _: EmptyResponse = try await client.send(APIRequest(path: "/transactions?note=PRIVATE_QUERY_CANARY", auth: .none))
                XCTFail("Expected the real typed HTTP failure")
            } catch APIError.problem(let problem) {
                XCTAssertEqual(problem.status, 422)
                XCTAssertEqual(problem.detail, "PRIVATE_BODY_CANARY")
            }
        }
        XCTAssertFalse(output.contains("PRIVATE_QUERY_CANARY"), "Diagnostic URL must not expose query values")
        XCTAssertFalse(output.contains("PRIVATE_BODY_CANARY"), "Diagnostic must not expose server error body")
        XCTAssertTrue(output.contains("GET"))
        XCTAssertTrue(output.contains("422"))
        XCTAssertNotNil(output.range(of: #"\[[0-9A-Fa-f]{8}-[0-9A-Fa-f-]{27,}\]"#, options: .regularExpression), "Keep locally generated request correlation")
    }

    func testSuccessfulHTTPDiagnosticsDoNotExposeQueryValues() async throws {
        let client = APIClient(baseURL: URL(string: "https://example.test/api/v1")!, transport: URLCaptureTransport(), sessionStore: TestSessionStore(session: nil))
        let output = try await captureHTTPStandardOutput {
            let _: EmptyResponse = try await client.send(APIRequest(path: "/transactions?note=PRIVATE_SUCCESS_CANARY", auth: .none))
        }
        XCTAssertFalse(output.contains("PRIVATE_SUCCESS_CANARY"))
        XCTAssertTrue(output.contains("200"))
    }

    func testSessionStoresAtomicallyRefuseStaleRefreshAndClearAcrossInstances() throws {
        let service = "test.mypocket.session-cas." + UUID().uuidString
        let keychain = KeychainSessionStore(service: service)
        let second = KeychainSessionStore(service: service)
        defer { keychain.clear() }
        var stores: [(any SessionStore, any SessionStore)] = [(keychain, second)]
#if DEBUG
        let memory = InMemorySessionStore()
        stores.append((memory, memory))
#endif
        for (writer, reader) in stores {
            let initial = testSession
            let refreshed = AccessSession(token: "refreshed", expiresAt: initial.expiresAt, refreshToken: "refreshed-refresh", refreshExpiresAt: initial.refreshExpiresAt, user: initial.user)
            let newerLogin = AccessSession(token: "new-login", expiresAt: initial.expiresAt, refreshToken: "new-login-refresh", refreshExpiresAt: initial.refreshExpiresAt, user: initial.user)
            let foreign = AccessSession(token: "foreign", expiresAt: initial.expiresAt, refreshToken: "foreign-refresh", refreshExpiresAt: initial.refreshExpiresAt, user: AuthUser(id: "owner-2", name: nil, email: nil, timezone: nil))
            try writer.save(initial)
            XCTAssertTrue(try reader.replace(refreshed, ifCredentialsMatch: initial))
            XCTAssertEqual(writer.load()?.token, "refreshed")
            XCTAssertFalse(reader.clear(ifCredentialsMatch: initial))
            try writer.save(newerLogin)
            XCTAssertFalse(try reader.replace(refreshed, ifCredentialsMatch: initial))
            XCTAssertFalse(reader.clear(ifCredentialsMatch: initial))
            XCTAssertEqual(writer.load()?.token, "new-login")
            XCTAssertFalse(try reader.replace(foreign, ifCredentialsMatch: newerLogin))
            XCTAssertEqual(writer.load()?.token, "new-login")
            XCTAssertTrue(reader.clear(ifCredentialsMatch: newerLogin))
            XCTAssertNil(writer.load())
            XCTAssertFalse(try reader.replace(refreshed, ifCredentialsMatch: initial))
            XCTAssertNil(writer.load(), "A refresh cannot resurrect a cleared session")
        }
    }

    func testConcurrentKeychainRefreshCandidatesHaveOnlyOneWinner() async throws {
        let service = "test.mypocket.session-cas-race." + UUID().uuidString
        let store = KeychainSessionStore(service: service)
        defer { store.clear() }
        let initial = testSession
        try store.save(initial)
        let winners = try await withThrowingTaskGroup(of: Bool.self, returning: Int.self) { group in
            for index in 0..<16 {
                group.addTask {
                    let candidate = AccessSession(token: "candidate-\(index)", expiresAt: initial.expiresAt, refreshToken: "candidate-refresh-\(index)", refreshExpiresAt: initial.refreshExpiresAt, user: initial.user)
                    return try KeychainSessionStore(service: service).replace(candidate, ifCredentialsMatch: initial)
                }
            }
            var total = 0
            for try await won in group { if won { total += 1 } }
            return total
        }
        XCTAssertEqual(winners, 1)
        XCTAssertTrue(store.load()?.token.hasPrefix("candidate-") == true)
    }
    func testSendEnvelopePreservesReceiptAndStableIntentThroughOneRefresh() async throws {
        let store = TestSessionStore(session: testSession)
        let transport = OwnerEnvelopeTransport(store: store, mode: .refresh(owner: "owner-1"))
        let client = APIClient(baseURL: URL(string: "https://example.test/api/v1")!, transport: transport, sessionStore: store)
        let body = Data(#"{"fixture":true}"#.utf8)
        let envelope: APIEnvelope<EmptyResponse> = try await client.sendEnvelope(APIRequest(path: "/transactions", method: .post, body: body, idempotencyKey: "persisted-key", expectedOwnerID: "owner-1"))
        guard case .object(let meta) = envelope.meta, case .object(let receipt) = meta["mutation"], case .string(let key) = receipt["idempotency_key"] else { return XCTFail("Receipt metadata discarded") }
        XCTAssertEqual(key, "persisted-key")
        let requests = await transport.requests
        XCTAssertEqual(requests.map { $0.url?.path }, ["/api/v1/transactions", "/api/v1/auth/refresh", "/api/v1/transactions"])
        XCTAssertEqual(requests[0].httpBody, body)
        XCTAssertEqual(requests[2].httpBody, body)
        XCTAssertEqual(requests[0].value(forHTTPHeaderField: "Idempotency-Key"), "persisted-key")
        XCTAssertEqual(requests[2].value(forHTTPHeaderField: "Idempotency-Key"), "persisted-key")
        XCTAssertEqual(requests[0].value(forHTTPHeaderField: "X-Request-ID"), requests[2].value(forHTTPHeaderField: "X-Request-ID"))
        XCTAssertEqual(requests[2].value(forHTTPHeaderField: "Authorization"), "Bearer access-new")
    }

    func testOwnerBoundRequestRejectsAnotherAccountBeforeTransport() async throws {
        let store = TestSessionStore(session: testSession)
        let transport = OwnerEnvelopeTransport(store: store, mode: .envelope)
        let client = APIClient(baseURL: URL(string: "https://example.test/api/v1")!, transport: transport, sessionStore: store)
        do {
            let _: EmptyResponse = try await client.send(APIRequest(path: "/transactions", expectedOwnerID: "other-owner"))
            XCTFail("Another owner's queue used current session")
        } catch { XCTAssertEqual(error as? APIError, .accountMismatch) }
        let requests = await transport.requests
        XCTAssertTrue(requests.isEmpty)
    }

    func testOwnerBoundRequestCannotAcceptResultAfterAccountChanges() async throws {
        for mode in [OwnerEnvelopeTransport.Mode.swapDuringResponse, .swapAndUnauthorized] {
            let store = TestSessionStore(session: testSession)
            let transport = OwnerEnvelopeTransport(store: store, mode: mode)
            let client = APIClient(baseURL: URL(string: "https://example.test/api/v1")!, transport: transport, sessionStore: store)
            do {
                let _: EmptyResponse = try await client.send(APIRequest(path: "/transactions", expectedOwnerID: "owner-1"))
                XCTFail("Response accepted after account switch")
            } catch { XCTAssertEqual(error as? APIError, .accountMismatch) }
            let requests = await transport.requests
            XCTAssertEqual(requests.count, 1, "Do not refresh another account for an old queue")
            XCTAssertEqual(store.load()?.user.id, "owner-2")
        }
    }

    func testRefreshRefusesForeignServerIdentityOrOverwritingNewLocalAccount() async throws {
        let cases: [(OwnerEnvelopeTransport.Mode, String?)] = [(.refresh(owner: "owner-2"), "access"), (.swapDuringRefresh, "owner-two-access"), (.swapDuringRefreshRejected, "owner-two-access"), (.clearDuringRefresh, nil), (.replaceSameOwnerDuringRefresh, "new-login-access")]
        for (mode, token) in cases {
            let store = TestSessionStore(session: testSession)
            let transport = OwnerEnvelopeTransport(store: store, mode: mode)
            let client = APIClient(baseURL: URL(string: "https://example.test/api/v1")!, transport: transport, sessionStore: store)
            do {
                let _: EmptyResponse = try await client.send(APIRequest(path: "/transactions", expectedOwnerID: "owner-1"))
                XCTFail("Refresh changed account identity")
            } catch { XCTAssertEqual(error as? APIError, .accountMismatch) }
            XCTAssertEqual(store.load()?.token, token)
        }
    }

    func testOwnerBoundRequestRequiresSessionNotUnattestedKeyOrAnonymousAuth() async throws {
        let store = TestSessionStore(session: testSession)
        let transport = OwnerEnvelopeTransport(store: store, mode: .envelope)
        let client = APIClient(baseURL: URL(string: "https://example.test/api/v1")!, transport: transport, sessionStore: store)
        for auth in [APIAuth.none, .apiKey("mpk_fixture")] {
            do {
                let _: EmptyResponse = try await client.send(APIRequest(path: "/transactions", auth: auth, expectedOwnerID: "owner-1"))
                XCTFail("Queue owner not attested by current session")
            } catch { XCTAssertEqual(error as? APIError, .accountMismatch) }
        }
        let requests = await transport.requests
        XCTAssertTrue(requests.isEmpty)
    }
    func testOldTravelContextCapabilityCannotAuthorizeOrdinaryOfflineWrites() throws {
        let legacy = #"{"contract_version":1,"enabled":true,"operations":["transaction.create","transfer.create"],"receipt_metadata":"/meta/mutation","idempotency_key_required_for_offline":true,"legacy_unkeyed_online":true}"#
        for suffix in ["", #", "transaction_create_context":"unknown_v2""#] {
            let body = String(legacy.dropLast()) + suffix + "}"
            let capability = try JSONDecoder.mypocket.decode(OfflineSyncCapabilitiesModel.self, from: Data(body.utf8))
            XCTAssertFalse(capability.supports(operation: "transaction.create"), "Old defaults may attach the trip active at delayed send time")
            XCTAssertTrue(capability.supports(operation: "transfer.create"), "Transfer has no automatic trip association")
        }
    }
    func testOfflineSyncCapabilityEnablesOnlyKnownAdvertisedCreateOperations() {
        let capability = OfflineSyncCapabilitiesModel(contractVersion: 1, enabled: true,
            operations: ["transaction.create", "transfer.create", "wallet.delete"],
            receiptMetadata: "/meta/mutation", idempotencyKeyRequiredForOffline: true, legacyUnkeyedOnline: true, transactionCreateContext: "explicit_travel_v1")
        XCTAssertTrue(capability.supports(operation: "transaction.create"))
        XCTAssertTrue(capability.supports(operation: "transfer.create"))
        XCTAssertFalse(capability.supports(operation: "wallet.delete"), "Server advertising a new operation cannot expand offline financial authority.")
    }

    func testOfflineSyncCapabilityRejectsDisabledUnknownOrMalformedPolicy() throws {
        func capability(version: Int = 1, enabled: Bool = true, operations: [String] = ["transaction.create"], metadata: String = "/meta/mutation", requiredKey: Bool = true) -> OfflineSyncCapabilitiesModel {
            OfflineSyncCapabilitiesModel(contractVersion: version, enabled: enabled, operations: operations,
                receiptMetadata: metadata, idempotencyKeyRequiredForOffline: requiredKey, legacyUnkeyedOnline: true, transactionCreateContext: "explicit_travel_v1")
        }
        for value in [capability(version: 0), capability(version: 2), capability(enabled: false),
                      capability(operations: []), capability(operations: ["transfer.create"]),
                      capability(metadata: "/data/mutation"), capability(requiredKey: false)] {
            XCTAssertFalse(value.supports(operation: "transaction.create"))
        }
        XCTAssertThrowsError(try JSONDecoder.mypocket.decode(OfflineSyncCapabilitiesModel.self, from: Data(#"{"enabled":true}"#.utf8)), "Missing contract fields cannot default to safe support.")
    }

    func testOfflineSyncCapabilitiesUsesAuthenticatedReadOnlyEndpoint() async throws {
        let transport = CapabilityCaptureTransport()
        let client = APIClient(baseURL: URL(string: "https://example.test/api/v1")!, transport: transport, sessionStore: TestSessionStore(session: testSession))
        let capability = try await FinanceAPI(client: client).syncCapabilities()
        let request = try XCTUnwrap(transport.request)
        XCTAssertEqual(request.httpMethod, "GET")
        XCTAssertEqual(request.url?.path, "/api/v1/sync/capabilities")
        XCTAssertEqual(request.value(forHTTPHeaderField: "Authorization"), "Bearer access")
        XCTAssertNil(request.httpBody)
        XCTAssertNil(request.value(forHTTPHeaderField: "Idempotency-Key"), "Capability checks must not be financial mutations.")
        XCTAssertTrue(capability.supports(operation: "transaction.create"))
    }

    func testOfflineSyncCapabilitiesOldServer404DoesNotFallbackToMutation() async throws {
        let transport = CapabilityCaptureTransport(statusCode: 404)
        let client = APIClient(baseURL: URL(string: "https://example.test/api/v1")!, transport: transport, sessionStore: TestSessionStore(session: testSession))
        do {
            _ = try await FinanceAPI(client: client).syncCapabilities()
            XCTFail("Old server must not be treated as offline-capable")
        } catch {
            XCTAssertEqual(error as? APIError, .http(status: 404))
        }
        XCTAssertEqual(transport.callCount, 1)
        XCTAssertEqual(transport.request?.httpMethod, "GET")
        XCTAssertEqual(transport.request?.url?.path, "/api/v1/sync/capabilities")
    }
    func testMutationMetadataPreservesTransactionObjectAndTransferArrayDecoding() throws {
        let transaction = #"{"id":"row-1","owner_id":"owner-1","wallet_id":"wallet-1","type":"expense","amount":12500,"occurred_at":"2026-09-01T03:00:00Z","included_in_reports":false,"created_at":"2026-09-01T03:00:00Z","updated_at":"2026-09-01T03:00:00Z"}"#
        let metadata = #"{"mutation":{"operation":"transaction.create","idempotency_key":"intent-1","status":"applied","replayed":true}}"#
        let objectData = Data("{\"data\":\(transaction),\"meta\":\(metadata)}".utf8)
        let object = try JSONDecoder.mypocket.decode(APIEnvelope<TransactionModel>.self, from: objectData)
        XCTAssertEqual(object.data.id, "row-1")
        XCTAssertEqual(object.data.amount, 12500)
        XCTAssertFalse(object.data.includedInReports)
        XCTAssertNotNil(object.meta)
        let pairMetadata = metadata.replacingOccurrences(of: "transaction.create", with: "transfer.create")
        let source = transaction.replacingOccurrences(of: "\"type\":\"expense\"", with: "\"transfer_id\":\"pair-1\",\"type\":\"expense\"")
        let destination = source.replacingOccurrences(of: "row-1", with: "row-2")
            .replacingOccurrences(of: "wallet-1", with: "wallet-2")
            .replacingOccurrences(of: "\"expense\"", with: "\"income\"")
        let arrayData = Data("{\"data\":[\(source),\(destination)],\"meta\":\(pairMetadata)}".utf8)
        let array = try JSONDecoder.mypocket.decode(APIEnvelope<[TransactionModel]>.self, from: arrayData)
        XCTAssertEqual(array.data.map(\.id), ["row-1", "row-2"])
        XCTAssertEqual(array.data.map(\.transferID), ["pair-1", "pair-1"])
        XCTAssertEqual(array.data.map(\.type), ["expense", "income"])
        XCTAssertEqual(array.data.first?.amount, 12500)
        XCTAssertNotNil(array.meta)
    }
    @MainActor
    func testAIEntryMixedNormalizedPhotoAndPDFUseRawMultipartAndPreserveInstruction() async throws {
        let renderer = UIGraphicsImageRenderer(size: CGSize(width: 4, height: 4))
        let bytes = try XCTUnwrap(renderer.image { context in
            UIColor.blue.setFill(); context.fill(CGRect(x: 0, y: 0, width: 4, height: 4))
        }.jpegData(compressionQuality: 0.8))
        let photo = try XCTUnwrap(MPDocumentImagePolicy.normalized(bytes, name: "photo.heic"))
        let pdf = Data("%PDF-1.4\nSynthetic document\n%%EOF".utf8)
        let transport = EntryMultipartCaptureTransport()
        let client = APIClient(baseURL: URL(string: "https://example.test/api/v1")!, transport: transport,
                               sessionStore: TestSessionStore(session: testSession))
        let result = try await FinanceAPI(client: client).processEntry(requestID: "photo-pdf-1",
            text: "Chỉ lấy giao dịch tháng 9", timezone: "Asia/Ho_Chi_Minh", files: [
                MultipartFile(fieldName: "files", fileName: photo.name, mimeType: photo.mimeType, data: photo.data),
                MultipartFile(fieldName: "files", fileName: "document.pdf", mimeType: "application/pdf", data: pdf)
            ])
        XCTAssertEqual(result.id, "batch-1")
        let request = try XCTUnwrap(transport.request)
        XCTAssertEqual(request.url?.path, "/api/v1/ai/entry/process")
        XCTAssertEqual(request.httpMethod, "POST")
        XCTAssertEqual(request.value(forHTTPHeaderField: "Authorization"), "Bearer access")
        XCTAssertTrue(request.value(forHTTPHeaderField: "Content-Type")?.hasPrefix("multipart/form-data; boundary=") == true)
        let body = try XCTUnwrap(request.httpBody)
        XCTAssertNotNil(body.range(of: photo.data))
        XCTAssertNotNil(body.range(of: pdf))
        let headers = String(decoding: body, as: UTF8.self)
        XCTAssertEqual(headers.components(separatedBy: "name=\"files\"; filename=").count - 1, 2)
        XCTAssertTrue(headers.contains("filename=\"photo.jpg\"\r\nContent-Type: image/jpeg"))
        XCTAssertTrue(headers.contains("filename=\"document.pdf\"\r\nContent-Type: application/pdf"))
        XCTAssertTrue(headers.contains("name=\"text\"\r\n\r\nChỉ lấy giao dịch tháng 9\r\n"))
        XCTAssertTrue(headers.contains("name=\"timezone\"\r\n\r\nAsia/Ho_Chi_Minh\r\n"))
        XCTAssertTrue(headers.contains("name=\"request_id\"\r\n\r\nphoto-pdf-1\r\n"))
        XCTAssertFalse(headers.contains("name=\"owner_id\""))
        XCTAssertFalse(headers.contains("base64"))
    }
    func testTravelFullFormUpdateExplicitlyClearsOptionalFields() async throws {
        let transport = TravelUpdateTransport()
        let client = APIClient(baseURL: URL(string: "https://example.test/api/v1")!, transport: transport, sessionStore: TestSessionStore(session: testSession))
        _ = try await FinanceAPI(client: client).updateTravelEvent(id: "trip-1", input: TravelEventInputModel(name: "Trip", context: nil, startsOn: nil, endsOn: "2026-10-14", active: false))
        let request = try XCTUnwrap(transport.request)
        XCTAssertEqual(request.httpMethod, "PATCH")
        XCTAssertEqual(request.url?.path, "/api/v1/travel/events/trip-1")
        XCTAssertEqual(request.value(forHTTPHeaderField: "Authorization"), "Bearer access")
        let data = try XCTUnwrap(request.httpBody)
        let body = try XCTUnwrap(JSONSerialization.jsonObject(with: data) as? [String: Any])
        XCTAssertEqual(body["starts_on"] as? String, "")
        XCTAssertEqual(body["context"] as? String, "")
        XCTAssertEqual(body["ends_on"] as? String, "2026-10-14")
        XCTAssertNil(body["owner_id"])
    }
    @MainActor
    func testFeedbackRefreshAndResolutionCannotOverwriteEachOther() async throws {
        let transport = DelayedFeedbackTransport()
        let session = AccessSession(token: "access", expiresAt: "2099-01-01T00:00:00Z", refreshToken: "refresh", refreshExpiresAt: "2099-02-01T00:00:00Z", user: AuthUser(id: "owner", name: nil, email: nil, timezone: nil))
        let client = APIClient(baseURL: URL(string: "https://example.test/api/v1")!, transport: transport, sessionStore: TestSessionStore(session: session))
        let model = FeedbackViewModel(api: FinanceAPI(client: client))
        await model.load()
        let item = try XCTUnwrap(model.items.first)
        await transport.holdNextPatch()
        await transport.holdNextList()
        let mutation = Task { await model.setResolved(item, resolved: true) }
        await transport.waitUntilPatchHeld()
        let refresh = Task { await model.load() }
        await transport.waitUntilHeld()
        await transport.releasePatch()
        await mutation.value
        let duringRefresh = await transport.patchCount
        XCTAssertEqual(duringRefresh, 1, "Resolution must not be dropped while an older refresh is pending")
        await transport.releaseList()
        await refresh.value
        XCTAssertTrue(try XCTUnwrap(model.items.first).isResolved)
        XCTAssertNil(model.errorMessage)
    }
#if DEBUG
    func testInMemorySessionStoreRoundTripAndClear() throws {
        let store = InMemorySessionStore()
        XCTAssertNil(store.load())

        let session = AccessSession(
            token: "fixture-access",
            expiresAt: "2099-01-01T00:00:00Z",
            refreshToken: "fixture-refresh",
            refreshExpiresAt: "2099-02-01T00:00:00Z",
            user: AuthUser(id: "fixture-user", name: nil, email: nil, timezone: nil)
        )
        try store.save(session)
        XCTAssertEqual(store.load(), session)
        store.clear()
        XCTAssertNil(store.load())
    }
#endif

    func testSessionRoundTripUsesCodableContract() throws {
        let user = AuthUser(id: "u1", name: "Test", email: "test@example.com", timezone: "Asia/Ho_Chi_Minh")
        let session = AccessSession(token: "access", expiresAt: "2026-10-05T00:00:00Z", refreshToken: "refresh", refreshExpiresAt: "2026-11-05T00:00:00Z", user: user)
        let data = try JSONEncoder.mypocket.encode(session)
        let decoded = try JSONDecoder.mypocket.decode(AccessSession.self, from: data)
        XCTAssertEqual(decoded, session)
    }

    func testProblemDetailsUsesSnakeCaseRequestID() throws {
        let data = Data(#"{"title":"Bad Request","status":400,"detail":"invalid","code":"INVALID","request_id":"req-1"}"#.utf8)
        let problem = try JSONDecoder.mypocket.decode(ProblemDetails.self, from: data)
        XCTAssertEqual(problem.requestID, "req-1")
        XCTAssertEqual(problem.code, "INVALID")
    }

    func testQueryInAPIRequestPathRemainsAQuery() async throws {
        let transport = URLCaptureTransport()
        let client = APIClient(
            baseURL: URL(string: "https://example.test/api/v1")!,
            transport: transport,
            sessionStore: TestSessionStore(session: AccessSession(token: "access", expiresAt: "2099-01-01T00:00:00Z", refreshToken: "refresh", refreshExpiresAt: "2099-02-01T00:00:00Z", user: AuthUser(id: "u", name: nil, email: nil, timezone: nil)))
        )

        let _: EmptyResponse = try await client.send(APIRequest(path: "/credit/wallets/w-1/statement?from=2026-10-01&to=2026-11-01"))

        XCTAssertEqual(transport.url?.path, "/api/v1/credit/wallets/w-1/statement")
        XCTAssertEqual(transport.url?.query, "from=2026-10-01&to=2026-11-01")
        XCTAssertNotNil(transport.requestID)
        XCTAssertFalse(transport.requestID?.isEmpty ?? true)
    }

    func testPasswordLoginUsesEmailEndpointAndReturnsSession() async throws {
        let transport = PasswordLoginTransport()
        let client = APIClient(
            baseURL: URL(string: "https://example.test/api/v1")!,
            transport: transport,
            sessionStore: TestSessionStore(session: nil)
        )
        let api = FinanceAPI(client: client)

        let output = try await api.login(email: "user@example.com", password: "test-password")

        XCTAssertEqual(output.session.user.email, "user@example.com")
        XCTAssertEqual(transport.request?.url?.path, "/api/v1/login")
        XCTAssertNil(transport.request?.value(forHTTPHeaderField: "Authorization"))
        let body = try XCTUnwrap(transport.request?.httpBody)
        let json = try XCTUnwrap(JSONSerialization.jsonObject(with: body) as? [String: String])
        XCTAssertEqual(json["email"], "user@example.com")
        XCTAssertEqual(json["password"], "test-password")
    }

    func testNotificationReadUsesOwnerSessionAndNoBody() async throws {
        let transport = AccountActionTransport()
        let client = APIClient(
            baseURL: URL(string: "https://example.test/api/v1")!,
            transport: transport,
            sessionStore: TestSessionStore(session: testSession)
        )

        let api = FinanceAPI(client: client)
        try await api.markNotificationRead(id: "notice-1")

        let request = try XCTUnwrap(transport.requests.first)
        XCTAssertEqual(request.httpMethod, "POST")
        XCTAssertEqual(request.url?.path, "/api/v1/notifications/notice-1/read")
        XCTAssertEqual(request.value(forHTTPHeaderField: "Authorization"), "Bearer access")
        XCTAssertNil(request.httpBody)
        XCTAssertNil(request.value(forHTTPHeaderField: "Idempotency-Key"))
    }

    func testAPIKeyCreateUsesOwnerSessionSnakeCaseBodyAndIdempotency() async throws {
        let transport = AccountActionTransport()
        let client = APIClient(
            baseURL: URL(string: "https://example.test/api/v1")!,
            transport: transport,
            sessionStore: TestSessionStore(session: testSession)
        )

        let api = FinanceAPI(client: client)
        let created = try await api.createAPIKey(name: "iPhone", expiresAt: "2027-01-01")

        XCTAssertEqual(created.id, "key-1")
        let request = try XCTUnwrap(transport.requests.first)
        XCTAssertEqual(request.httpMethod, "POST")
        XCTAssertEqual(request.url?.path, "/api/v1/api-keys")
        XCTAssertEqual(request.value(forHTTPHeaderField: "Authorization"), "Bearer access")
        XCTAssertFalse(request.value(forHTTPHeaderField: "Idempotency-Key")?.isEmpty ?? true)
        let json = try XCTUnwrap(JSONSerialization.jsonObject(with: try XCTUnwrap(request.httpBody)) as? [String: Any])
        XCTAssertEqual(json["name"] as? String, "iPhone")
        XCTAssertNil(json["scopes"], "Full-owner keys have no per-operation scopes")
        XCTAssertEqual(json["expires_at"] as? String, "2027-01-01")
    }

    func testAPIKeyRevokeAcceptsOwnerSessionNoContentResponse() async throws {
        let transport = AccountActionTransport()
        let client = APIClient(
            baseURL: URL(string: "https://example.test/api/v1")!,
            transport: transport,
            sessionStore: TestSessionStore(session: testSession)
        )

        let api = FinanceAPI(client: client)
        try await api.revokeAPIKey(id: "key-1")

        let request = try XCTUnwrap(transport.requests.first)
        XCTAssertEqual(request.httpMethod, "DELETE")
        XCTAssertEqual(request.url?.path, "/api/v1/api-keys/key-1")
        XCTAssertEqual(request.value(forHTTPHeaderField: "Authorization"), "Bearer access")
    }

    func testCreateWalletUsesCallerSuppliedIdempotencyKeyForActionRetry() async throws {
        let transport = WalletCreateCaptureTransport()
        let client = APIClient(
            baseURL: URL(string: "https://example.test/api/v1")!,
            transport: transport,
            sessionStore: TestSessionStore(session: testSession)
        )
        let input = WalletModelInput(
            name: "Du lịch", type: "goal", openingBalance: 0, isInTotal: true,
            description: nil, targetAmount: 10_000_000, targetDate: "2027-01-01",
            creditLimit: nil, lastStatementBalance: nil, statementDay: nil, paymentDueDay: nil
        )

        _ = try await FinanceAPI(client: client).createWallet(input, idempotencyKey: "assistant-action-1")

        let request = try XCTUnwrap(transport.request)
        XCTAssertEqual(request.httpMethod, "POST")
        XCTAssertEqual(request.url?.path, "/api/v1/wallets")
        XCTAssertEqual(request.value(forHTTPHeaderField: "Idempotency-Key"), "assistant-action-1")
    }

    func testConfirmAdvisorWalletActionUsesBoundActionEndpoint() async throws {
        let transport = WalletCreateCaptureTransport()
        let client = APIClient(
            baseURL: URL(string: "https://example.test/api/v1")!,
            transport: transport,
            sessionStore: TestSessionStore(session: testSession)
        )

        _ = try await FinanceAPI(client: client).confirmAdvisorWalletAction(
            actionID: "assistant-action-1", conversationID: "conversation-1", generation: 4
        )

        let request = try XCTUnwrap(transport.request)
        XCTAssertEqual(request.httpMethod, "POST")
        XCTAssertEqual(request.url?.path, "/api/v1/ai/advisor/actions/assistant-action-1/confirm")
        XCTAssertEqual(request.value(forHTTPHeaderField: "Authorization"), "Bearer access")
        XCTAssertEqual(request.value(forHTTPHeaderField: "Idempotency-Key"), nil, "The server derives the mutation key from action_id")
        let body = try XCTUnwrap(request.httpBody)
        let json = try XCTUnwrap(JSONSerialization.jsonObject(with: body) as? [String: Any])
        XCTAssertEqual(json["conversation_id"] as? String, "conversation-1")
        XCTAssertEqual(json["generation"] as? Int, 4)
        XCTAssertNil(json["name"], "Financial proposal fields must not be client-controlled")
    }

    func testConfirmAdvisorJarActionUsesBoundActionEndpoint() async throws {
        let transport = JarActionCaptureTransport()
        let client = APIClient(
            baseURL: URL(string: "https://example.test/api/v1")!,
            transport: transport,
            sessionStore: TestSessionStore(session: testSession)
        )

        _ = try await FinanceAPI(client: client).confirmAdvisorJarAction(
            actionID: "jar-action-1", conversationID: "conversation-1", generation: 4
        )

        let request = try XCTUnwrap(transport.request)
        XCTAssertEqual(request.httpMethod, "POST")
        XCTAssertEqual(request.url?.path, "/api/v1/ai/advisor/actions/jar-action-1/confirm")
        let body = try XCTUnwrap(request.httpBody)
        let json = try XCTUnwrap(JSONSerialization.jsonObject(with: body) as? [String: Any])
        XCTAssertEqual(json["conversation_id"] as? String, "conversation-1")
        XCTAssertEqual(json["generation"] as? Int, 4)
        XCTAssertNil(json["name"])
    }

    func testCategoryArchiveUsesAuthenticatedStatusPatchContract() async throws {
        let transport = CategoryStatusTransport()
        let client = APIClient(
            baseURL: URL(string: "https://example.test/api/v1")!,
            transport: transport,
            sessionStore: TestSessionStore(session: testSession)
        )
        let category = try await FinanceAPI(client: client).setCategoryActive(id: "category-1", active: false)
        XCTAssertEqual(category.isActive, false)
        let request = try XCTUnwrap(transport.request)
        XCTAssertEqual(request.httpMethod, "PATCH")
        XCTAssertEqual(request.url?.path, "/api/v1/categories/category-1/status")
        XCTAssertEqual(request.value(forHTTPHeaderField: "Authorization"), "Bearer access")
        let body = try XCTUnwrap(request.httpBody)
        let json = try XCTUnwrap(JSONSerialization.jsonObject(with: body) as? [String: Bool])
        XCTAssertEqual(json["is_active"], false)
    }

    func testPortfolioTradeUsesOwnerPathSnakeCaseBodyAndIdempotency() async throws {
        let transport = PortfolioTradeTransport()
        let client = APIClient(
            baseURL: URL(string: "https://example.test/api/v1")!,
            transport: transport,
            sessionStore: TestSessionStore(session: testSession)
        )

        let api = FinanceAPI(client: client)
        let trade = try await api.createPortfolioTrade(
            assetID: "asset-1",
            input: PortfolioTradeInputModel(
                side: "buy",
                quantity: "1.25000000",
                unitPrice: 72_000,
                fee: 1_000,
                occurredAt: "2026-10-07T05:00:00Z",
                note: "DCA"
            )
        )

        XCTAssertEqual(trade.id, "trade-1")
        let request = try XCTUnwrap(transport.request)
        XCTAssertEqual(request.httpMethod, "POST")
        XCTAssertEqual(request.url?.path, "/api/v1/portfolio/assets/asset-1/trades")
        XCTAssertEqual(request.value(forHTTPHeaderField: "Authorization"), "Bearer access")
        XCTAssertFalse(request.value(forHTTPHeaderField: "Idempotency-Key")?.isEmpty ?? true)
        let json = try XCTUnwrap(JSONSerialization.jsonObject(with: try XCTUnwrap(request.httpBody)) as? [String: Any])
        XCTAssertEqual(json["side"] as? String, "buy")
        XCTAssertEqual(json["quantity"] as? String, "1.25000000")
        XCTAssertEqual(json["unit_price"] as? Int, 72_000)
        XCTAssertEqual(json["fee"] as? Int, 1_000)
        XCTAssertEqual(json["occurred_at"] as? String, "2026-10-07T05:00:00Z")
        XCTAssertEqual(json["note"] as? String, "DCA")
    }

    private var testSession: AccessSession {
        AccessSession(
            token: "access",
            expiresAt: "2099-01-01T00:00:00Z",
            refreshToken: "refresh",
            refreshExpiresAt: "2099-02-01T00:00:00Z",
            user: AuthUser(id: "owner-1", name: "Test", email: "test@example.com", timezone: "Asia/Ho_Chi_Minh")
        )
    }

    func testFeedbackScreenshotUsesBackendPNGContract() async throws {
        let transport = FeedbackMultipartTransport()
        let client = APIClient(
            baseURL: URL(string: "https://example.test/api/v1")!,
            transport: transport,
            sessionStore: TestSessionStore(session: AccessSession(token: "access", expiresAt: "2099-01-01T00:00:00Z", refreshToken: "refresh", refreshExpiresAt: "2099-02-01T00:00:00Z", user: AuthUser(id: "u", name: nil, email: nil, timezone: nil)))
        )
        let api = FinanceAPI(client: client)

        _ = try await api.createFeedback(
            FeedbackInputModel(type: "bug", title: "Ảnh lỗi", description: "Không tải được ảnh"),
            screenshot: Data("fixture-image".utf8)
        )

        let body = String(decoding: transport.request?.httpBody ?? Data(), as: UTF8.self)
        XCTAssertTrue(body.contains("filename=\"screen.png\""))
        XCTAssertTrue(body.contains("Content-Type: image/png"))
    }

    @MainActor
    func testFeedbackScreenshotStorageErrorsFallBackToTextOnly() {
        let invalid = APIError.problem(ProblemDetails(
            title: "Bad Request",
            status: 400,
            detail: "invalid screenshot",
            instance: "/api/v1/feedback",
            code: "FEEDBACK_SCREENSHOT_INVALID",
            requestID: "req-1"
        ))
        let unavailable = APIError.problem(ProblemDetails(
            title: "Service Unavailable",
            status: 503,
            detail: "storage unavailable",
            instance: "/api/v1/feedback",
            code: "FEEDBACK_SCREENSHOT_STORAGE_UNAVAILABLE",
            requestID: "req-2"
        ))
        let transport = APIError.transport("offline")

        XCTAssertTrue(FeedbackViewModel.shouldFallbackToTextOnly(invalid))
        XCTAssertTrue(FeedbackViewModel.shouldFallbackToTextOnly(unavailable))
        XCTAssertFalse(FeedbackViewModel.shouldFallbackToTextOnly(transport))
    }

    @MainActor
    func testScreenSnapshotEncodesPNGBytes() {
        let image = UIGraphicsImageRenderer(size: CGSize(width: 1, height: 1)).image { _ in }
        let data = ScreenSnapshot.pngData(for: image)
        XCTAssertEqual(data?.prefix(8), Data([137, 80, 78, 71, 13, 10, 26, 10]))
    }

    @MainActor
    func testScreenSnapshotKeepsComplexCaptureUnderFeedbackLimit() {
        let width = 960
        let height = 1_600
        var bytes = Data(count: width * height * 4)
        bytes.withUnsafeMutableBytes { rawBuffer in
            var state: UInt32 = 0x9E3779B9
            for index in 0..<rawBuffer.count {
                state = state &* 1_664_525 &+ 1_013_904_223
                rawBuffer[index] = UInt8(truncatingIfNeeded: state >> 24)
            }
        }
        let provider = CGDataProvider(data: bytes as CFData)!
        let image = UIImage(cgImage: CGImage(
            width: width,
            height: height,
            bitsPerComponent: 8,
            bitsPerPixel: 32,
            bytesPerRow: width * 4,
            space: CGColorSpaceCreateDeviceRGB(),
            bitmapInfo: CGBitmapInfo(rawValue: CGImageAlphaInfo.premultipliedLast.rawValue),
            provider: provider,
            decode: nil,
            shouldInterpolate: false,
            intent: .defaultIntent
        )!)

        let data = ScreenSnapshot.pngData(for: image)
        XCTAssertNotNil(data)
        XCTAssertLessThanOrEqual(data?.count ?? Int.max, 1 << 20)
    }
}

private actor AdvisorContractTransport: HTTPTransport {
    let payloads: [Data]
    private(set) var requests: [URLRequest] = []
    init(payloads: [Data]) { self.payloads = payloads }
    func data(for request: URLRequest) async throws -> (Data, URLResponse) {
        let index = requests.count
        requests.append(request)
        guard index < payloads.count else { throw APIError.transport("Unexpected additional advisor request") }
        return (payloads[index], HTTPURLResponse(url: request.url!, statusCode: 200, httpVersion: nil, headerFields: nil)!)
    }
}

private actor DelayedFeedbackTransport: HTTPTransport {
    private var holdNextPatchRequest = false
    private var holdNext = false
    private var patchHeld = false
    private var held = false
    private var patchWaiter: CheckedContinuation<Void, Never>?
    private var waiter: CheckedContinuation<Void, Never>?
    private var patchRelease: CheckedContinuation<Void, Never>?
    private var release: CheckedContinuation<Void, Never>?
    private var saved = false
    private(set) var patchCount = 0
    func holdNextPatch() { holdNextPatchRequest = true }
    func waitUntilPatchHeld() async {
        if patchHeld { return }
        await withCheckedContinuation { patchWaiter = $0 }
    }
    func releasePatch() { patchRelease?.resume(); patchRelease = nil; patchHeld = false }
    func holdNextList() { holdNext = true }
    func waitUntilHeld() async {
        if held { return }
        await withCheckedContinuation { waiter = $0 }
    }
    func releaseList() { release?.resume(); release = nil; held = false }
    func data(for request: URLRequest) async throws -> (Data, URLResponse) {
        let data: Data
        if request.url!.path.hasSuffix("/changelog") {
            data = Data(#"{"data":[],"meta":null}"#.utf8)
        } else {
            if request.httpMethod == "PATCH" {
                patchCount += 1
                if holdNextPatchRequest {
                    holdNextPatchRequest = false
                    patchHeld = true
                    patchWaiter?.resume(); patchWaiter = nil
                    await withCheckedContinuation { patchRelease = $0 }
                }
                saved = true
            }
            // Capture the list response before the optional hold. The delayed
            // GET must remain stale after a concurrent PATCH; otherwise this
            // regression test can pass merely because the fake server echoes
            // the mutation that happened while the request was suspended.
            let listResolvedAtRequestStart = saved
            let row: [String: Any] = ["id":"fb", "type":"bug", "title":"Issue", "description":"Details", "status":"open", "resolved": listResolvedAtRequestStart, "created_at":"2026-10-14T00:00:00Z", "updated_at":"2026-10-14T00:00:00Z", "screenshot_available":false]
            let payload: Any = request.httpMethod == "PATCH" ? row as Any : [row] as Any
            data = try JSONSerialization.data(withJSONObject: ["data": payload])
            if request.httpMethod != "PATCH" && holdNext {
                holdNext = false; held = true; waiter?.resume(); waiter = nil
                await withCheckedContinuation { release = $0 }
            }
        }
        return (data, HTTPURLResponse(url: request.url!, statusCode: 200, httpVersion: nil, headerFields: nil)!)
    }
}

private final class CapabilityCaptureTransport: HTTPTransport, @unchecked Sendable {
    private(set) var request: URLRequest?
    private(set) var callCount = 0
    let statusCode: Int
    init(statusCode: Int = 200) { self.statusCode = statusCode }
    func data(for request: URLRequest) async throws -> (Data, URLResponse) {
        self.request = request
        callCount += 1
        let body = #"{"data":{"contract_version":1,"enabled":true,"operations":["transaction.create","transfer.create"],"receipt_metadata":"/meta/mutation","idempotency_key_required_for_offline":true,"legacy_unkeyed_online":true,"transaction_create_context":"explicit_travel_v1"},"meta":null}"#
        return (Data((statusCode == 404 ? "404 page not found" : body).utf8), HTTPURLResponse(url: request.url!, statusCode: statusCode, httpVersion: nil, headerFields: nil)!)
    }
}

private final class URLCaptureTransport: HTTPTransport, @unchecked Sendable {
    private(set) var url: URL?
    private(set) var requestID: String?

    func data(for request: URLRequest) async throws -> (Data, URLResponse) {
        url = request.url
        requestID = request.value(forHTTPHeaderField: "X-Request-ID")
        return (Data(#"{"data":{}}"#.utf8), HTTPURLResponse(url: request.url!, statusCode: 200, httpVersion: nil, headerFields: nil)!)
    }
}

private final class PasswordLoginTransport: HTTPTransport, @unchecked Sendable {
    private(set) var request: URLRequest?

    func data(for request: URLRequest) async throws -> (Data, URLResponse) {
        self.request = request
        let response = #"{"data":{"token":"access","expires_at":"2099-01-01T00:00:00Z","refresh_token":"refresh","refresh_expires_at":"2099-02-01T00:00:00Z","user":{"id":"u-1","name":"Test","email":"user@example.com","timezone":"Asia/Ho_Chi_Minh"}},"meta":null}"#
        return (Data(response.utf8), HTTPURLResponse(url: request.url!, statusCode: 200, httpVersion: nil, headerFields: nil)!)
    }
}

private final class AccountActionTransport: HTTPTransport, @unchecked Sendable {
    private(set) var requests: [URLRequest] = []

    func data(for request: URLRequest) async throws -> (Data, URLResponse) {
        requests.append(request)
        let path = request.url?.path ?? ""
        if request.httpMethod == "POST", path.hasSuffix("/api-keys") {
            let body = #"{"data":{"id":"key-1","name":"iPhone","scopes":["finance:read"],"expires_at":"2027-01-01","secret":"mpk_secret"},"meta":null}"#
            return (Data(body.utf8), HTTPURLResponse(url: request.url!, statusCode: 201, httpVersion: nil, headerFields: nil)!)
        }
        if path.contains("/notifications/") || request.httpMethod == "DELETE" {
            return (Data(), HTTPURLResponse(url: request.url!, statusCode: 204, httpVersion: nil, headerFields: nil)!)
        }
        throw URLError(.badServerResponse)
    }
}

private final class WalletCreateCaptureTransport: HTTPTransport, @unchecked Sendable {
    private(set) var request: URLRequest?

    func data(for request: URLRequest) async throws -> (Data, URLResponse) {
        self.request = request
        let body = #"{"data":{"id":"wallet-1","name":"Du lịch","type":"goal","currency":"VND","opening_balance":0,"current_balance":0,"available_credit":null,"is_in_total":true,"description":null,"target_amount":10000000,"target_date":"2027-01-01","credit_limit":null,"last_statement_balance":null,"statement_day":null,"payment_due_day":null,"goal_notifications_enabled":false},"meta":null}"#
        return (Data(body.utf8), HTTPURLResponse(url: request.url!, statusCode: 201, httpVersion: nil, headerFields: nil)!)
    }
}

private final class JarActionCaptureTransport: HTTPTransport, @unchecked Sendable {
    private(set) var request: URLRequest?

    func data(for request: URLRequest) async throws -> (Data, URLResponse) {
        self.request = request
        let body = #"{"data":{"month":"2026-09","jar_id":"jar-1","name":"Ăn uống","allocation_mode":"fixed","allocation_amount":400000,"allocation_percent":null,"active":true},"meta":null}"#
        return (Data(body.utf8), HTTPURLResponse(url: request.url!, statusCode: 201, httpVersion: nil, headerFields: nil)!)
    }
}

private final class EntryMultipartCaptureTransport: HTTPTransport, @unchecked Sendable {
    private(set) var request: URLRequest?
    func data(for request: URLRequest) async throws -> (Data, URLResponse) {
        self.request = request
        let body = Data(#"{"data":{"id":"batch-1","processing":false,"proposals":[],"reply":null,"error":null},"meta":null}"#.utf8)
        return (body, HTTPURLResponse(url: request.url!, statusCode: 200, httpVersion: nil, headerFields: nil)!)
    }
}

private final class TravelUpdateTransport: HTTPTransport, @unchecked Sendable {
    private(set) var request: URLRequest?
    func data(for request: URLRequest) async throws -> (Data, URLResponse) {
        self.request = request
        let response = #"{"data":{"id":"trip-1","owner_id":"owner","name":"Trip","context":null,"starts_on":null,"ends_on":"2026-10-14","active":false,"created_at":"2026-10-14T00:00:00Z","updated_at":"2026-10-14T00:00:00Z"},"meta":null}"#
        return (Data(response.utf8), HTTPURLResponse(url: request.url!, statusCode: 200, httpVersion: nil, headerFields: nil)!)
    }
}

private final class PortfolioTradeTransport: HTTPTransport, @unchecked Sendable {
    private(set) var request: URLRequest?

    func data(for request: URLRequest) async throws -> (Data, URLResponse) {
        self.request = request
        let response = #"{"data":{"id":"trade-1","owner_id":"owner-1","asset_id":"asset-1","side":"buy","quantity":"1.25000000","unit_price":72000,"fee":1000,"occurred_at":"2026-10-07T05:00:00Z","note":"DCA"},"meta":null}"#
        return (Data(response.utf8), HTTPURLResponse(url: request.url!, statusCode: 201, httpVersion: nil, headerFields: nil)!)
    }
}

private final class CategoryStatusTransport: HTTPTransport, @unchecked Sendable {
    private(set) var request: URLRequest?

    func data(for request: URLRequest) async throws -> (Data, URLResponse) {
        self.request = request
        let response = #"{"data":{"id":"category-1","name":"Cafe","kind":"expense","icon_key":"tag","is_system":false,"is_active":false,"wallet_ids":[]},"meta":null}"#
        return (Data(response.utf8), HTTPURLResponse(url: request.url!, statusCode: 200, httpVersion: nil, headerFields: nil)!)
    }
}

private final class FeedbackMultipartTransport: HTTPTransport, @unchecked Sendable {
    private(set) var request: URLRequest?

    func data(for request: URLRequest) async throws -> (Data, URLResponse) {
        self.request = request
        let response = #"{"data":{"id":"feedback-1","type":"bug","title":"Ảnh lỗi","description":"Không tải được ảnh","status":"open","fixed_at":null,"changelog_id":null,"created_at":"2026-10-06T00:00:00Z","updated_at":"2026-10-06T00:00:00Z","screenshot_available":true},"meta":null}"#
        return (Data(response.utf8), HTTPURLResponse(url: request.url!, statusCode: 201, httpVersion: nil, headerFields: nil)!)
    }
}

final class RefreshCoordinatorTests: XCTestCase {
    func testTwoSequentialCallsCanUseTheSameContract() async throws {
        let coordinator = RefreshCoordinator {
            AccessSession(token: "a", expiresAt: "now", refreshToken: "r", refreshExpiresAt: "later", user: AuthUser(id: "u", name: nil, email: nil, timezone: nil))
        }
        let first = try await coordinator.refreshIfNeeded()
        let second = try await coordinator.refreshIfNeeded()
        XCTAssertEqual(first.token, second.token)
    }
}

private actor OwnerEnvelopeTransport: HTTPTransport {
    enum Mode: Equatable, Sendable { case envelope, refresh(owner: String), swapDuringResponse, swapAndUnauthorized, swapDuringRefresh, swapDuringRefreshRejected, clearDuringRefresh, replaceSameOwnerDuringRefresh }
    let store: any SessionStore
    let mode: Mode
    private(set) var requests: [URLRequest] = []
    init(store: any SessionStore, mode: Mode) { self.store = store; self.mode = mode }
    func data(for request: URLRequest) async throws -> (Data, URLResponse) {
        requests.append(request)
        let foreign = AccessSession(token: "owner-two-access", expiresAt: "2099-01-01T00:00:00Z", refreshToken: "owner-two-refresh", refreshExpiresAt: "2099-02-01T00:00:00Z", user: AuthUser(id: "owner-2", name: nil, email: nil, timezone: nil))
        var status = 200
        let body: String
        if request.url?.path.hasSuffix("/auth/refresh") == true {
            if mode == .swapDuringRefresh { try store.save(foreign) }
            if mode == .swapDuringRefreshRejected { try store.save(foreign); status = 401 }
            if mode == .clearDuringRefresh { store.clear() }
            if mode == .replaceSameOwnerDuringRefresh {
                try store.save(AccessSession(token: "new-login-access", expiresAt: "2099-01-01T00:00:00Z", refreshToken: "new-login-refresh", refreshExpiresAt: "2099-02-01T00:00:00Z", user: AuthUser(id: "owner-1", name: nil, email: nil, timezone: nil)))
            }
            let owner: String
            if case .refresh(let requested) = mode { owner = requested } else { owner = "owner-1" }
            body = "{\"data\":{\"token\":\"access-new\",\"expires_at\":\"2099-01-01T00:00:00Z\",\"refresh_token\":\"refresh-new\",\"refresh_expires_at\":\"2099-02-01T00:00:00Z\",\"user\":{\"id\":\"\(owner)\"}},\"meta\":null}"
        } else {
            if mode == .swapDuringResponse || mode == .swapAndUnauthorized { try store.save(foreign) }
            if mode == .swapAndUnauthorized || (requests.count == 1 && mode != .envelope && mode != .swapDuringResponse) { status = 401 }
            body = #"{"data":{},"meta":{"mutation":{"operation":"transaction.create","idempotency_key":"persisted-key","status":"applied","replayed":true}}}"#
        }
        return (Data(body.utf8), HTTPURLResponse(url: request.url!, statusCode: status, httpVersion: nil, headerFields: nil)!)
    }
}

private final class TestSessionStore: SessionStore, @unchecked Sendable {
    private let lock = NSRecursiveLock()
    private var stored: AccessSession?
    var session: AccessSession? {
        get { load() }
        set { lock.withLock { stored = newValue } }
    }
    init(session: AccessSession?) { stored = session }
    func load() -> AccessSession? { lock.withLock { stored } }
    func save(_ session: AccessSession) throws { lock.withLock { stored = session } }
    func clear() { lock.withLock { stored = nil } }
    func replace(_ session: AccessSession, ifCredentialsMatch original: AccessSession) throws -> Bool {
        lock.withLock {
            guard stored?.hasSameCredentials(as: original) == true, session.user.id == original.user.id else { return false }
            stored = session; return true
        }
    }
    func clear(ifCredentialsMatch original: AccessSession) -> Bool {
        lock.withLock {
            guard stored?.hasSameCredentials(as: original) == true else { return false }
            stored = nil; return true
        }
    }
}

private final class MultipartRefreshTransport: HTTPTransport, @unchecked Sendable {
    private(set) var multipartRequests = 0
    private(set) var refreshRequests = 0
    private(set) var multipartRequestIDs: [String] = []

    func data(for request: URLRequest) async throws -> (Data, URLResponse) {
        let isRefresh = request.url?.path.hasSuffix("/auth/refresh") == true
        if isRefresh {
            refreshRequests += 1
            let data = Data(#"{"data":{"token":"new-access","expires_at":"2099-01-01T00:00:00Z","refresh_token":"new-refresh","refresh_expires_at":"2099-02-01T00:00:00Z","user":{"id":"u-1","name":"Test","email":"test@example.com","timezone":"Asia/Ho_Chi_Minh"}}}"#.utf8)
            return (data, HTTPURLResponse(url: request.url!, statusCode: 200, httpVersion: nil, headerFields: nil)!)
        }
        multipartRequests += 1
        multipartRequestIDs.append(request.value(forHTTPHeaderField: "X-Request-ID") ?? "")
        if multipartRequests == 1 {
            return (Data(), HTTPURLResponse(url: request.url!, statusCode: 401, httpVersion: nil, headerFields: nil)!)
        }
        let data = Data(#"{"data":{"id":"request-1","processing":false,"proposals":[]}}"#.utf8)
        return (data, HTTPURLResponse(url: request.url!, statusCode: 200, httpVersion: nil, headerFields: nil)!)
    }
}

final class MultipartAPIClientTests: XCTestCase {
    func testMultipartSessionRefreshesAndRetriesOnce() async throws {
        let initial = AccessSession(token: "expired", expiresAt: "2020-01-01T00:00:00Z", refreshToken: "refresh", refreshExpiresAt: "2099-01-01T00:00:00Z", user: AuthUser(id: "u-1", name: "Test", email: "test@example.com", timezone: "Asia/Ho_Chi_Minh"))
        let store = TestSessionStore(session: initial)
        let transport = MultipartRefreshTransport()
        let client = APIClient(baseURL: URL(string: "https://example.test/api/v1")!, transport: transport, sessionStore: store)
        let result: EntryProcessModel = try await client.sendMultipart(path: "/ai/entry/process", fields: ["request_id": "r-1", "text": "test"], idempotencyKey: "idempotent-1")
        XCTAssertEqual(result.id, "request-1")
        XCTAssertEqual(transport.multipartRequests, 2)
        XCTAssertEqual(transport.refreshRequests, 1)
        XCTAssertEqual(transport.multipartRequestIDs.count, 2)
        XCTAssertFalse(transport.multipartRequestIDs[0].isEmpty)
        XCTAssertEqual(transport.multipartRequestIDs[0], transport.multipartRequestIDs[1])
        XCTAssertEqual(store.session?.token, "new-access")
    }
}

private final class NoContentTransport: HTTPTransport, @unchecked Sendable {
    func data(for request: URLRequest) async throws -> (Data, URLResponse) {
        (Data(), HTTPURLResponse(url: request.url!, statusCode: 204, httpVersion: nil, headerFields: nil)!)
    }
}

final class EmptyMutationResponseTests: XCTestCase {
    func testNoContentMutationDoesNotAttemptEnvelopeDecoding() async throws {
        let client = APIClient(
            baseURL: URL(string: "https://example.test/api/v1")!,
            transport: NoContentTransport(),
            sessionStore: TestSessionStore(session: AccessSession(token: "access", expiresAt: "2099-01-01T00:00:00Z", refreshToken: "refresh", refreshExpiresAt: "2099-02-01T00:00:00Z", user: AuthUser(id: "u", name: nil, email: nil, timezone: nil)))
        )
        let _: EmptyResponse = try await client.send(APIRequest(path: "/transactions/bulk-delete", method: .post, body: Data(#"{"transaction_ids":["tx-1"]}"#.utf8)))
    }
}

private typealias RestoreSessionStore = TestSessionStore

private final class RestoreTransport: HTTPTransport, @unchecked Sendable {
    private(set) var refreshRequests = 0

    func data(for request: URLRequest) async throws -> (Data, URLResponse) {
        XCTAssertTrue(request.url?.path.hasSuffix("/auth/refresh") == true)
        refreshRequests += 1
        let payload = Data(#"{"data":{"token":"restored-access","expires_at":"2099-01-01T00:00:00Z","refresh_token":"restored-refresh","refresh_expires_at":"2099-02-01T00:00:00Z","user":{"id":"u-1","name":"Test","email":"test@example.com","timezone":"Asia/Ho_Chi_Minh"}}}"#.utf8)
        return (payload, HTTPURLResponse(url: request.url!, statusCode: 200, httpVersion: nil, headerFields: nil)!)
    }
}

private final class InvalidatingTransport: HTTPTransport, @unchecked Sendable {
    private(set) var requestCount = 0

    func data(for request: URLRequest) async throws -> (Data, URLResponse) {
        requestCount += 1
        return (Data(#"{"title":"Unauthorized","status":401,"detail":"expired","code":"TOKEN_INVALID"}"#.utf8), HTTPURLResponse(url: request.url!, statusCode: 401, httpVersion: nil, headerFields: nil)!)
    }
}

private struct InvalidResponse: Decodable {}

/// Restore tests model an unenrolled device explicitly. Never consult or clear
/// the simulator's actual protected credential left by device-auth UI coverage.
private struct UnenrolledDeviceTestStore: DeviceSignInStore {
    var metadata: DeviceSignInMetadata? { nil }
    var isAvailable: Bool { false }
    func save(_ credential: DeviceSignInCredential) throws { throw APIError.missingSession }
    func unlock() async throws -> DeviceSignInCredential { throw APIError.missingSession }
    func clear() {}
}

@MainActor
final class AuthRestoreTests: XCTestCase {
    private let biometricKey = "mypocket.biometric.enabled"
    private let preferencesSuite = "MyPocketAuthRestoreTests.\(UUID().uuidString)"
    private lazy var preferences = UserDefaults(suiteName: preferencesSuite)!

    func testExpiredAccessRefreshesBeforeBiometricGate() async throws {
        defer { preferences.removePersistentDomain(forName: preferencesSuite) }
        preferences.set(true, forKey: biometricKey)
        let initial = AccessSession(token: "expired", expiresAt: "2020-01-01T00:00:00Z", refreshToken: "refresh", refreshExpiresAt: "2099-01-01T00:00:00Z", user: AuthUser(id: "u-1", name: "Test", email: nil, timezone: "Asia/Ho_Chi_Minh"))
        let store = RestoreSessionStore(session: initial)
        let transport = RestoreTransport()
        let apiClient = APIClient(baseURL: URL(string: "https://example.test/api/v1")!, transport: transport, sessionStore: store)
        let auth = AuthViewModel(store: store, apiClient: apiClient, deviceStore: UnenrolledDeviceTestStore(), preferences: preferences)

        await auth.restore()

        XCTAssertEqual(transport.refreshRequests, 1)
        XCTAssertEqual(store.session?.token, "restored-access")
        guard case .locked(let restored) = auth.phase else {
            return XCTFail("an enabled biometric gate should use the refreshed session")
        }
        XCTAssertEqual(restored.token, "restored-access")
    }

    func testExpiredAccessAndRefreshClearsSessionWithoutShowingBiometricGate() async {
        defer { preferences.removePersistentDomain(forName: preferencesSuite) }
        preferences.set(true, forKey: biometricKey)
        let expired = AccessSession(token: "expired", expiresAt: "2020-01-01T00:00:00Z", refreshToken: "refresh", refreshExpiresAt: "2020-01-01T00:00:00Z", user: AuthUser(id: "u-1", name: nil, email: nil, timezone: nil))
        let store = RestoreSessionStore(session: expired)
        let transport = RestoreTransport()
        let apiClient = APIClient(baseURL: URL(string: "https://example.test/api/v1")!, transport: transport, sessionStore: store)
        let auth = AuthViewModel(store: store, apiClient: apiClient, deviceStore: UnenrolledDeviceTestStore(), preferences: preferences)

        await auth.restore()

        XCTAssertEqual(transport.refreshRequests, 0)
        XCTAssertNil(store.load())
        XCTAssertEqual(auth.phase, .unauthenticated)
        XCTAssertFalse(auth.biometricEnabled)
    }

    func testRefreshRejectionInvalidatesAuthenticatedSession() async throws {
        defer { preferences.removePersistentDomain(forName: preferencesSuite) }
        let session = AccessSession(token: "access", expiresAt: "2099-01-01T00:00:00Z", refreshToken: "refresh", refreshExpiresAt: "2099-02-01T00:00:00Z", user: AuthUser(id: "u-1", name: nil, email: nil, timezone: nil))
        let store = RestoreSessionStore(session: session)
        let transport = InvalidatingTransport()
        let client = APIClient(baseURL: URL(string: "https://example.test/api/v1")!, transport: transport, sessionStore: store)
        let auth = AuthViewModel(store: store, apiClient: client, deviceStore: UnenrolledDeviceTestStore(), preferences: preferences)
        await auth.restore()
        guard case .authenticated = auth.phase else {
            return XCTFail("the valid access session should start authenticated")
        }

        do {
            let _: InvalidResponse = try await client.send(APIRequest(path: "/auth/profile"))
            XCTFail("an unauthorized request should not succeed")
        } catch APIError.unauthorized {
            // Expected: the refresh endpoint also returned 401.
        } catch {
            XCTFail("unexpected error: \(error)")
        }
        try await Task.sleep(for: .milliseconds(50))

        XCTAssertNil(store.load())
        XCTAssertEqual(auth.phase, .unauthenticated)
        XCTAssertFalse(auth.biometricEnabled)
    }

    func testTimezoneUpdateKeepsAuthenticatedSessionContextInSync() async {
        defer { preferences.removePersistentDomain(forName: preferencesSuite) }
        let session = AccessSession(token: "access", expiresAt: "2099-01-01T00:00:00Z", refreshToken: "refresh", refreshExpiresAt: "2099-02-01T00:00:00Z", user: AuthUser(id: "u-1", name: "Test", email: "test@example.com", timezone: "Asia/Ho_Chi_Minh"))
        let store = RestoreSessionStore(session: session)
        let client = APIClient(baseURL: URL(string: "https://example.test/api/v1")!, transport: RestoreTransport(), sessionStore: store)
        let auth = AuthViewModel(store: store, apiClient: client, deviceStore: UnenrolledDeviceTestStore(), preferences: preferences)

        await auth.restore()
        auth.updateAccountTimezone("UTC")

        XCTAssertEqual(store.session?.user.timezone, "UTC")
        guard case .authenticated(let updated) = auth.phase else {
            return XCTFail("a valid authenticated session should remain authenticated")
        }
        XCTAssertEqual(updated.user.timezone, "UTC")
        XCTAssertEqual(updated.token, "access")
    }
}

@MainActor
final class PasswordAuthViewModelTests: XCTestCase {
    func testSignInWithEmailAndPasswordPersistsSession() async {
        let store = TestSessionStore(session: nil)
        let client = APIClient(
            baseURL: URL(string: "https://example.test/api/v1")!,
            transport: PasswordLoginTransport(),
            sessionStore: store
        )
        let auth = AuthViewModel(store: store, apiClient: client)

        await auth.signIn(email: "user@example.com", password: "test-password")

        guard case .authenticated(let session) = auth.phase else {
            return XCTFail("a successful password login should authenticate the session")
        }
        XCTAssertEqual(session.user.email, "user@example.com")
        XCTAssertEqual(store.load()?.token, "access")
        XCTAssertNil(auth.errorMessage)
    }

}
