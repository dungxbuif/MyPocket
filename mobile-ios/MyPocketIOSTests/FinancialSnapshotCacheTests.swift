import Foundation
import XCTest
@testable import MyPocketIOS

private struct SnapshotFixtureKeys: OutboxKeyProviding {
    let byte: UInt8
    init(byte: UInt8 = 0x58) { self.byte = byte }
    func key(for namespace: String, createIfMissing: Bool) throws -> Data { Data(repeating: byte, count: 32) }
}

private struct SnapshotOfflineTransport: HTTPTransport {
    let status: Int?
    var body = "{}"
    func data(for request: URLRequest) async throws -> (Data, URLResponse) {
        guard let status else { throw URLError(.notConnectedToInternet) }
        return (Data(body.utf8), HTTPURLResponse(url: request.url!, statusCode: status, httpVersion: nil, headerFields: nil)!)
    }
}

final class FinancialSnapshotCacheTests: XCTestCase {
    func testHTTPStatusControlsRecoveryEvenWhenProblemBodyOmitsOrContradictsIt() async throws {
        for body in ["{}", #"{"status":400,"detail":"fixture"}"#] {
            let client = APIClient(baseURL: server, transport: SnapshotOfflineTransport(status: 503, body: body), sessionStore: InMemorySessionStore())
            do {
                let _: EmptyResponse = try await client.send(APIRequest(path: "/fixture", auth: .none))
                XCTFail("503 must not return success")
            } catch {
                guard case APIError.problem(let problem) = error else { return XCTFail("Expected typed problem") }
                XCTAssertEqual(problem.status, 503)
                XCTAssertTrue(FinancialSnapshotCache.allowsFallback(error))
            }
        }
    }
    @MainActor
    func testTransactionScreenRestoresMatchingSnapshotOnlyForTransientLoadFailure() async throws {
        let cache = try FinancialSnapshotCache(directory: root(), serverURL: server, keys: SnapshotFixtureKeys())
        try await cache.save(snapshot())
        for status in [nil, 503, 401, 403, 400] as [Int?] {
            let sessions = InMemorySessionStore()
            try sessions.save(AccessSession(token: "fixture-token", expiresAt: "2099-01-01T00:00:00Z", refreshToken: "fixture-refresh", refreshExpiresAt: "2099-02-01T00:00:00Z", user: AuthUser(id: "owner-private", name: nil, email: nil, timezone: nil)))
            let api = FinanceAPI(client: APIClient(baseURL: server, transport: SnapshotOfflineTransport(status: status), sessionStore: sessions))
            let model = TransactionsViewModel(api: api, ownerID: "owner-private", snapshotCache: cache)
            await model.load()
            if status == nil || status == 503 {
                XCTAssertEqual(model.transactions.first?.id, "tx-private")
                XCTAssertEqual(model.accountTimezone, "Asia/Ho_Chi_Minh")
                XCTAssertEqual(model.cachedAt, Date(timeIntervalSince1970: 1790000000))
                XCTAssertNil(model.errorMessage)
            } else {
                XCTAssertTrue(model.transactions.isEmpty)
                XCTAssertNil(model.cachedAt)
                XCTAssertNotNil(model.errorMessage)
            }
        }
    }
    private let server = URL(string: "https://cache.test/api/v1")!
    private func root() throws -> URL {
        let directory = FileManager.default.temporaryDirectory.appendingPathComponent("financial-cache-" + UUID().uuidString)
        try FileManager.default.createDirectory(at: directory, withIntermediateDirectories: true)
        addTeardownBlock { try FileManager.default.removeItem(at: directory) }
        return directory
    }
    private func snapshot(owner: String = "owner-private", rowOwner: String? = nil) -> FinancialSnapshot {
        let row = TransactionModel(id: "tx-private", ownerID: rowOwner ?? owner, walletID: "wallet-private", type: "expense", amount: 120000,
            occurredAt: "2026-09-01T03:00:00Z", note: "private-financial-note", includedInReports: true, createdAt: "2026-09-01T03:00:00Z", updatedAt: "2026-09-01T03:00:00Z")
        return FinancialSnapshot(ownerID: owner, fetchedAt: Date(timeIntervalSince1970: 1790000000), transactions: [row], wallets: [], categories: [], timezone: "Asia/Ho_Chi_Minh")
    }

    func testReportFieldsRoundTripAndNilUpdatePreservesVerifiedReports() async throws {
        let cache = try FinancialSnapshotCache(directory: root(), serverURL: server, keys: SnapshotFixtureKeys())
        var first = snapshot()
        first.monthSummary = MonthSummaryModel(month: "2026-10", timezone: "Asia/Ho_Chi_Minh", startAt: "2026-09-30T17:00:00Z", nextStartAt: "2026-10-31T17:00:00Z", isCurrent: true, isComplete: false, income: 100, expense: 10, net: 90, transactionCount: 1, categories: [], note: "", jarSummary: nil, calculatedAt: "2026-10-15T00:00:00Z")
        first.insider = InsiderSummaryModel(month: "2026-10", timezone: "Asia/Ho_Chi_Minh", startAt: "2026-09-30T17:00:00Z", nextStartAt: "2026-10-31T17:00:00Z", isCurrent: true, income: 100, expense: 10, net: 90, previousExpense: 20, expenseDelta: -10, expenseChangeBPS: -5000, averageDailyExpense: 1, daysConsidered: 15, spendingIncomeRatioBPS: 1000, topCategories: [], topExpenses: [], estimated: false)
        try await cache.save(first)

        let baseRefresh = snapshot()
        let ledgerRefresh = FinancialSnapshot(ownerID: baseRefresh.ownerID, fetchedAt: Date(timeIntervalSince1970: 1790000100), transactions: baseRefresh.transactions, wallets: baseRefresh.wallets, categories: baseRefresh.categories, timezone: baseRefresh.timezone)
        try await cache.save(ledgerRefresh)

        let loaded = try await cache.load(ownerID: "owner-private")
        XCTAssertEqual(loaded?.monthSummary?.month, "2026-10")
        XCTAssertEqual(loaded?.monthSummary?.expense, 10)
        XCTAssertEqual(loaded?.insider?.expense, 10)
        XCTAssertEqual(loaded?.fetchedAt, ledgerRefresh.fetchedAt)
    }

    @MainActor
    func testOverviewRestoresOwnerScopedReportSnapshotOnTransientFailure() async throws {
        let cache = try FinancialSnapshotCache(directory: root(), serverURL: server, keys: SnapshotFixtureKeys())
        let wallet = WalletModel(id: "wallet-private", name: "Ví chính", type: "basic", currency: "VND", openingBalance: 0, currentBalance: 900, availableCredit: nil, isInTotal: true, description: nil, targetAmount: nil, targetDate: nil, creditLimit: nil, lastStatementBalance: nil, statementDay: nil, paymentDueDay: nil)
        var cached = snapshot()
        cached = FinancialSnapshot(ownerID: cached.ownerID, fetchedAt: cached.fetchedAt, transactions: cached.transactions, wallets: [wallet], categories: cached.categories, timezone: cached.timezone, monthSummary: nil, insider: nil)
        try await cache.save(cached)
        let sessions = InMemorySessionStore()
        try sessions.save(AccessSession(token: "fixture-token", expiresAt: "2099-01-01T00:00:00Z", refreshToken: "fixture-refresh", refreshExpiresAt: "2099-02-01T00:00:00Z", user: AuthUser(id: "owner-private", name: nil, email: nil, timezone: nil)))
        let api = FinanceAPI(client: APIClient(baseURL: server, transport: SnapshotOfflineTransport(status: 503), sessionStore: sessions))
        let model = OverviewViewModel(api: api, ownerID: "owner-private", snapshotCache: cache)

        await model.load()

        XCTAssertEqual(model.wallets.first?.id, "wallet-private")
        XCTAssertEqual(model.totalBalance, 900)
        XCTAssertEqual(model.cachedAt, Date(timeIntervalSince1970: 1790000000))
        XCTAssertNil(model.errorMessage)
    }

    func testReportsOfflinePolicyRequiresExactReportMonth() async throws {
        var value = snapshot()
        value.monthSummary = MonthSummaryModel(month: "2026-10", timezone: "Asia/Ho_Chi_Minh", startAt: "2026-09-30T17:00:00Z", nextStartAt: "2026-10-31T17:00:00Z", isCurrent: true, isComplete: false, income: 100, expense: 10, net: 90, transactionCount: 1, categories: [], note: "", jarSummary: nil, calculatedAt: "2026-10-15T00:00:00Z")
        XCTAssertFalse(ReportsOfflinePolicy.accepts(value, month: "2026-09"))
        XCTAssertFalse(ReportsOfflinePolicy.accepts(value, month: "2026-10"))
        value.insider = InsiderSummaryModel(month: "2026-10", timezone: "Asia/Ho_Chi_Minh", startAt: "2026-09-30T17:00:00Z", nextStartAt: "2026-10-31T17:00:00Z", isCurrent: true, income: 100, expense: 10, net: 90, previousExpense: 20, expenseDelta: -10, expenseChangeBPS: -5000, averageDailyExpense: 1, daysConsidered: 15, spendingIncomeRatioBPS: 1000, topCategories: [], topExpenses: [], estimated: false)
        XCTAssertTrue(ReportsOfflinePolicy.accepts(value, month: "2026-10"))
    }

    func testEncryptedSnapshotSurvivesRestartWithoutPlaintextOrCrossOwnerServerRead() async throws {
        let directory = try root()
        let cache = try FinancialSnapshotCache(directory: directory, serverURL: server, keys: SnapshotFixtureKeys())
        let missing = try await cache.load(ownerID: "owner-private")
        XCTAssertNil(missing)
        try await cache.save(snapshot())
        let restarted = try FinancialSnapshotCache(directory: directory, serverURL: server, keys: SnapshotFixtureKeys())
        let value = try await restarted.load(ownerID: "owner-private")
        XCTAssertEqual(value?.transactions.first?.id, "tx-private")
        XCTAssertEqual(value?.transactions.first?.amount, 120000)
        XCTAssertEqual(value?.timezone, "Asia/Ho_Chi_Minh")
        XCTAssertEqual(value?.fetchedAt, Date(timeIntervalSince1970: 1790000000))
        let foreignOwner = try await cache.load(ownerID: "other-owner")
        XCTAssertNil(foreignOwner)
        let other = try FinancialSnapshotCache(directory: directory, serverURL: URL(string: "https://other.test/api/v1")!, keys: SnapshotFixtureKeys())
        let foreignServer = try await other.load(ownerID: "owner-private")
        XCTAssertNil(foreignServer)
        let files = try FileManager.default.contentsOfDirectory(at: directory, includingPropertiesForKeys: nil)
        XCTAssertEqual(files.count, 1)
        let bytes = try Data(contentsOf: XCTUnwrap(files.first))
        for secret in ["owner-private", "wallet-private", "tx-private", "private-financial-note"] {
            XCTAssertNil(bytes.range(of: Data(secret.utf8)))
            XCTAssertFalse(files[0].lastPathComponent.contains(secret))
        }
    }

    func testWrongKeyOrCorruptSnapshotFailsClosedAndPreservesOriginalFile() async throws {
        let directory = try root()
        let cache = try FinancialSnapshotCache(directory: directory, serverURL: server, keys: SnapshotFixtureKeys())
        try await cache.save(snapshot())
        let file = try XCTUnwrap(FileManager.default.contentsOfDirectory(at: directory, includingPropertiesForKeys: nil).first)
        let original = try Data(contentsOf: file)
        let wrong = try FinancialSnapshotCache(directory: directory, serverURL: server, keys: SnapshotFixtureKeys(byte: 0x70))
        do { _ = try await wrong.load(ownerID: "owner-private"); XCTFail("Wrong key must not look like empty cache") } catch {}
        do { try await wrong.save(snapshot()); XCTFail("Wrong key must not overwrite previous cache") } catch {}
        XCTAssertEqual(try Data(contentsOf: file), original)
        try Data([1, 2, 3]).write(to: file)
        do { _ = try await cache.load(ownerID: "owner-private"); XCTFail("Corruption must not look like empty cache") } catch {}
        XCTAssertEqual(try Data(contentsOf: file), Data([1, 2, 3]))
    }

    func testForeignRowsInvalidIdentityAndUnknownVersionCannotReplaceKnownSnapshot() async throws {
        let cache = try FinancialSnapshotCache(directory: root(), serverURL: server, keys: SnapshotFixtureKeys())
        try await cache.save(snapshot())
        for value in [snapshot(rowOwner: "other-owner"), snapshot(owner: " ")] {
            do { try await cache.save(value); XCTFail("Invalid snapshot must be rejected") } catch {}
        }
        var unsupported = snapshot(); unsupported.version = 2
        do { try await cache.save(unsupported); XCTFail("Unknown schema must be rejected") } catch {}
        let retained = try await cache.load(ownerID: "owner-private")
        XCTAssertEqual(retained?.transactions.first?.id, "tx-private")
    }

    func testFallbackOnlyAllowsTransientFailuresNeverAuthOrMalformedResponses() {
        for error in [APIError.transport("offline"), .http(status: 408), .http(status: 429), .http(status: 503)] {
            XCTAssertTrue(FinancialSnapshotCache.allowsFallback(error))
        }
        for error in [APIError.unauthorized, .missingSession, .accountMismatch, .decoding("bad"), .http(status: 400), .http(status: 403), .http(status: 404), .http(status: 409), .http(status: 422)] {
            XCTAssertFalse(FinancialSnapshotCache.allowsFallback(error))
        }
    }
}
