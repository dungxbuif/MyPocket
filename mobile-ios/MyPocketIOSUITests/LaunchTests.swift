import XCTest
import LocalAuthentication

#if LOCAL_API_UI
private enum OfflineFaultFixture {
    struct Stats: Decodable { let blocked: Int; let dropped: Int; let keys: [String]; let upstreamStatuses: [Int] }
    static func configure(mode: String, walletID: String, reset: Bool = false) async throws {
        var request = URLRequest(url: URL(string: "http://127.0.0.1:18081/__fault")!)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.setValue("offline-e2e", forHTTPHeaderField: "X-MP-Local-Test")
        request.httpBody = try JSONSerialization.data(withJSONObject: ["mode": mode, "walletID": walletID, "reset": reset])
        let (_, response) = try await URLSession.shared.data(for: request)
        guard (response as? HTTPURLResponse)?.statusCode == 200 else { throw NSError(domain: "OfflineFaultFixture", code: 1) }
    }
    static func stats() async throws -> Stats {
        var request = URLRequest(url: URL(string: "http://127.0.0.1:18081/__fault")!)
        request.setValue("offline-e2e", forHTTPHeaderField: "X-MP-Local-Test")
        let (data, response) = try await URLSession.shared.data(for: request)
        guard (response as? HTTPURLResponse)?.statusCode == 200 else { throw NSError(domain: "OfflineFaultFixture", code: 2) }
        return try JSONDecoder().decode(Stats.self, from: data)
    }
}

private struct FixtureEnvelope<Value: Decodable>: Decodable {
    let data: Value
}

private struct FixtureLogin: Decodable {
    let token: String
}

private struct FixtureNotification: Decodable, Sendable {
    let id: String
    let title: String
    let readAt: String?

    enum CodingKeys: String, CodingKey {
        case id
        case title
        case readAt = "read_at"
    }
}

private struct FixtureWallet: Decodable, Sendable {
    let id: String
    let name: String
    let isInTotal: Bool?
    let currentBalance: Int64?

    enum CodingKeys: String, CodingKey {
        case id
        case name
        case isInTotal = "is_in_total"
        case currentBalance = "current_balance"
    }
}

private struct FixtureWalletHandle: Sendable {
    let token: String
    let wallet: FixtureWallet
}

private struct FixturePortfolioAsset: Decodable, Sendable {
    let id: String
    let symbol: String
    let name: String
}

private struct FixturePortfolioAssetList: Decodable, Sendable {
    let id: String
    let symbol: String
    let name: String
    let latestPrice: Int64?

    enum CodingKeys: String, CodingKey {
        case id, symbol, name
        case latestPrice = "latest_price"
    }
}

private struct FixturePortfolioAssetHandle: Sendable {
    let token: String
    let asset: FixturePortfolioAsset
}

private struct FixturePortfolioTrade: Decodable, Sendable {
    let id: String
    let assetID: String
    let side: String
    let quantity: String
    let unitPrice: Int64
    let fee: Int64

    enum CodingKeys: String, CodingKey {
        case id
        case assetID = "asset_id"
        case side
        case quantity
        case unitPrice = "unit_price"
        case fee
    }
}

private struct FixtureBudget: Decodable, Sendable {
    let id: String
    let name: String
    let categoryID: String?
    let ended: Bool?

    enum CodingKeys: String, CodingKey {
        case id
        case name
        case categoryID = "category_id"
        case ended
    }
}

private struct FixtureCategory: Decodable, Sendable {
    let id: String
    let name: String
    let kind: String?
    let systemKey: String?
    let walletIDs: [String]?
    let iconKey: String?
    let isSystem: Bool?
    let isActive: Bool?

    enum CodingKeys: String, CodingKey {
        case id
        case name
        case kind
        case systemKey = "system_key"
        case walletIDs = "wallet_ids"
        case iconKey = "icon_key"
        case isSystem = "is_system"
        case isActive = "is_active"
    }
}

private struct FixtureAPIKey: Decodable, Sendable {
    let id: String
    let name: String
}

private struct FixtureJarItem: Decodable, Sendable {
    let jarID: String
    let name: String

    enum CodingKeys: String, CodingKey {
        case jarID = "jar_id"
        case name
    }
}

private struct FixtureJarMonth: Decodable, Sendable {
    let items: [FixtureJarItem]
}

private struct FixtureRecurring: Decodable, Sendable {
    let id: String
    let name: String
    let walletID: String
    let categoryID: String?
    let interval: Int

    enum CodingKeys: String, CodingKey {
        case id
        case name
        case walletID = "wallet_id"
        case categoryID = "category_id"
        case interval
    }
}

private struct FixtureTravelEvent: Decodable, Sendable {
    let id: String
    let name: String
    let startsOn: String?
    let endsOn: String?
    let context: String?
    enum CodingKeys: String, CodingKey {
        case id, name, context
        case startsOn = "starts_on"
        case endsOn = "ends_on"
    }
}

private struct FixtureTransaction: Decodable, Sendable, Equatable {
    let id: String
    let walletID: String
    let categoryID: String?
    let amount: Int64
    let note: String?
    let travelEventID: String?
    let type: String?
    let occurredAt: String?
    let includedInReports: Bool?
    let transferID: String?
    let creditKind: String?
    let creditPaymentID: String?

    enum CodingKeys: String, CodingKey {
        case id
        case walletID = "wallet_id"
        case categoryID = "category_id"
        case amount
        case note
        case travelEventID = "travel_event_id"
        case type
        case occurredAt = "occurred_at"
        case includedInReports = "included_in_reports"
        case transferID = "transfer_id"
        case creditKind = "credit_kind"
        case creditPaymentID = "credit_payment_id"
    }
}

private struct FixtureCreditStatement: Decodable, Sendable {
    let items: [FixtureTransaction]
}

private struct FixtureMonthSummary: Decodable, Sendable {
    let income: Int64
    let expense: Int64
    let transactionCount: Int

    enum CodingKeys: String, CodingKey {
        case income
        case expense
        case transactionCount = "transaction_count"
    }
}

private struct FixtureAdvisorHistoryMessage: Decodable, Equatable {
    struct Part: Decodable, Equatable { let type: String; let text: String }
    let id: String
    let conversation_id: String
    let generation: Int
    let seq: Int
    let run_id: String
    let role: String
    let parts_version: Int
    let parts: [Part]
    let created_at: String
}

private enum LocalFixtureError: Error {
    case invalidResponse
    case http(Int, String)
}

private enum LocalAPIFixture {
    struct OwnerProfile: Decodable { let id: String; let timezone: String }
    static func ownerID() async throws -> String {
        try await request(path: "/auth/profile", method: "GET", token: loginToken(), body: nil, decode: FixtureEnvelope<OwnerProfile>.self).data.id
    }
    static func advisorHistory(conversationID: String, beforeSequence: Int? = nil) async throws -> [FixtureAdvisorHistoryMessage] {
        var query = URLComponents()
        query.queryItems = [URLQueryItem(name: "conversation_id", value: conversationID)]
        if let beforeSequence { query.queryItems?.append(URLQueryItem(name: "before_seq", value: String(beforeSequence))) }
        return try await request(path: "/ai/advisor/conversation/messages?\(query.percentEncodedQuery ?? "")",
            method: "GET", token: loginToken(), body: nil, decode: FixtureEnvelope<[FixtureAdvisorHistoryMessage]>.self).data
    }
    static func ownerTimezone() async throws -> String {
        try await request(path: "/auth/profile", method: "GET", token: loginToken(), body: nil, decode: FixtureEnvelope<OwnerProfile>.self).data.timezone
    }
    struct FeedbackState: Decodable { let id: String; let resolved: Bool; let status: String }

    static func createResolutionFeedback(title: String) async throws -> FeedbackState {
        let token = try await loginToken()
        let body = try JSONSerialization.data(withJSONObject: ["type":"bug", "title":title, "description":"Synthetic resolution/reopen UI proof"])
        return try await request(path: "/feedback", method: "POST", token: token, body: body, decode: FixtureEnvelope<FeedbackState>.self).data
    }

    static func resolutionFeedback(id: String) async throws -> FeedbackState {
        let token = try await loginToken()
        return try await request(path: "/feedback/\(id)", method: "GET", token: token, body: nil, decode: FixtureEnvelope<FeedbackState>.self).data
    }
    static let baseURL = URL(
        string: ProcessInfo.processInfo.environment["MYPOCKET_TEST_API_BASE_URL"] ?? "http://127.0.0.1:8081/api/v1"
    )!

    static func wallet(name: String, type: String, isInTotal: Bool = true) async throws -> FixtureWalletHandle {
        let token = try await loginToken()
        let existing = try await request(
            path: "/wallets",
            method: "GET",
            token: token,
            body: nil,
            decode: FixtureEnvelope<[FixtureWallet]>.self
        ).data
        for item in existing where item.name == name {
            try await deleteWallet(id: item.id, token: token)
        }

        var payload: [String: Any] = [
            "name": name,
            "type": type,
            "opening_balance": 100000,
            "is_in_total": NSNumber(value: isInTotal),
        ]
        if type == "credit" {
            payload["credit_limit"] = 5_000_000
            payload["last_statement_balance"] = 0
            payload["statement_day"] = 15
            payload["payment_due_day"] = 25
        }
        if type == "goal" {
            payload["target_amount"] = 1_000_000
            payload["target_date"] = "2026-12-31"
            payload["goal_notifications_enabled"] = true
        }
        let data = try JSONSerialization.data(withJSONObject: payload)
        let created = try await request(
            path: "/wallets",
            method: "POST",
            token: token,
            body: data,
            decode: FixtureEnvelope<FixtureWallet>.self
        ).data
        return FixtureWalletHandle(token: token, wallet: created)
    }

    static func delete(_ fixture: FixtureWalletHandle) async throws {
        try await deleteWallet(id: fixture.wallet.id, token: fixture.token)
    }

    static func creditStatement(for fixture: FixtureWalletHandle) async throws -> FixtureCreditStatement {
        try await request(
            path: "/credit/wallets/\(fixture.wallet.id)/statement",
            method: "GET",
            token: fixture.token,
            body: nil,
            decode: FixtureEnvelope<FixtureCreditStatement>.self
        ).data
    }

    static func deleteWallets(named name: String) async throws {
        let token = try await loginToken()
        let existing = try await request(
            path: "/wallets",
            method: "GET",
            token: token,
            body: nil,
            decode: FixtureEnvelope<[FixtureWallet]>.self
        ).data
        for item in existing where item.name == name {
            try await deleteWallet(id: item.id, token: token)
        }
    }

    static func wallet(named name: String) async throws -> FixtureWallet {
        let token = try await loginToken()
        let existing = try await request(
            path: "/wallets",
            method: "GET",
            token: token,
            body: nil,
            decode: FixtureEnvelope<[FixtureWallet]>.self
        ).data
        guard let item = existing.first(where: { $0.name == name }) else {
            throw LocalFixtureError.http(404, "Wallet fixture not found: \(name)")
        }
        return item
    }

    static func deleteBudgets(named name: String) async throws {
        let token = try await loginToken()
        let existing = try await request(
            path: "/budgets",
            method: "GET",
            token: token,
            body: nil,
            decode: FixtureEnvelope<BudgetList>.self
        ).data.items
        for item in existing where item.name == name {
            _ = try await request(
                path: "/budgets/\(item.id)",
                method: "DELETE",
                token: token,
                body: nil,
                decode: EmptyFixtureResponse.self
            )
        }
    }

    static func budget(named name: String) async throws -> FixtureBudget {
        let token = try await loginToken()
        let existing = try await request(
            path: "/budgets",
            method: "GET",
            token: token,
            body: nil,
            decode: FixtureEnvelope<BudgetList>.self
        ).data.items
        guard let item = existing.first(where: { $0.name == name }) else {
            throw LocalFixtureError.http(404, "Budget fixture not found: \(name)")
        }
        return item
    }

    static func endedBudget(name: String) async throws -> FixtureBudget {
        let token = try await loginToken()
        let body = try JSONSerialization.data(withJSONObject: [
            "name": name, "limit_amount": 120000,
            "start_date": "2020-01-01", "end_date": "2020-01-31"
        ])
        return try await request(path: "/budgets", method: "POST", token: token,
                                 body: body, decode: FixtureEnvelope<FixtureBudget>.self).data
    }

    static func deleteCategories(named name: String) async throws {
        let token = try await loginToken()
        let existing = try await request(
            path: "/categories",
            method: "GET",
            token: token,
            body: nil,
            decode: FixtureEnvelope<[FixtureCategory]>.self
        ).data
        for item in existing where item.name == name {
            _ = try await request(
                path: "/categories/\(item.id)",
                method: "DELETE",
                token: token,
                body: nil,
                decode: EmptyFixtureResponse.self
            )
        }
    }

    static func allCategories() async throws -> [FixtureCategory] {
        let token = try await loginToken()
        return try await request(path: "/categories", method: "GET", token: token, body: nil,
                                 decode: FixtureEnvelope<[FixtureCategory]>.self).data
    }

    static func category(named name: String) async throws -> FixtureCategory {
        let token = try await loginToken()
        let existing = try await request(
            path: "/categories",
            method: "GET",
            token: token,
            body: nil,
            decode: FixtureEnvelope<[FixtureCategory]>.self
        ).data
        guard let item = existing.first(where: { $0.name == name }) else {
            throw LocalFixtureError.http(404, "Category fixture not found: \(name)")
        }
        return item
    }

    static func category(kind: String, walletID: String) async throws -> FixtureCategory {
        let token = try await loginToken()
        let existing = try await request(
            path: "/categories",
            method: "GET",
            token: token,
            body: nil,
            decode: FixtureEnvelope<[FixtureCategory]>.self
        ).data
        guard let item = existing.first(where: {
            $0.kind == kind && (($0.walletIDs?.isEmpty ?? true) || $0.walletIDs?.contains(walletID) == true)
        }) else {
            throw LocalFixtureError.http(404, "Applicable \(kind) category fixture not found for wallet \(walletID)")
        }
        return item
    }

    static func category(systemKey: String, walletID: String) async throws -> FixtureCategory {
        let token = try await loginToken()
        let existing = try await request(
            path: "/categories",
            method: "GET",
            token: token,
            body: nil,
            decode: FixtureEnvelope<[FixtureCategory]>.self
        ).data
        guard let item = existing.first(where: {
            $0.systemKey == systemKey
                && $0.isActive != false
                && ($0.walletIDs?.isEmpty ?? true || $0.walletIDs?.contains(walletID) == true)
        }) else {
            throw LocalFixtureError.http(404, "Applicable category fixture not found for system key \(systemKey)")
        }
        return item
    }

    static func category(kind: String) async throws -> FixtureCategory {
        let token = try await loginToken()
        let existing = try await request(
            path: "/categories",
            method: "GET",
            token: token,
            body: nil,
            decode: FixtureEnvelope<[FixtureCategory]>.self
        ).data
        guard let item = existing.first(where: { $0.kind == kind }) else {
            throw LocalFixtureError.http(404, "Category fixture not found for kind \(kind)")
        }
        return item
    }

    static func deleteAPIKeys(named name: String) async throws {
        let token = try await loginToken()
        let existing = try await request(
            path: "/api-keys",
            method: "GET",
            token: token,
            body: nil,
            decode: FixtureEnvelope<[FixtureAPIKey]>.self
        ).data
        for item in existing where item.name == name {
            _ = try await request(
                path: "/api-keys/\(item.id)",
                method: "DELETE",
                token: token,
                body: nil,
                decode: EmptyFixtureResponse.self
            )
        }
    }

    static func apiKey(named name: String) async throws -> FixtureAPIKey {
        let token = try await loginToken()
        let existing = try await request(
            path: "/api-keys",
            method: "GET",
            token: token,
            body: nil,
            decode: FixtureEnvelope<[FixtureAPIKey]>.self
        ).data
        guard let item = existing.first(where: { $0.name == name }) else {
            throw LocalFixtureError.http(404, "API key fixture not found: \(name)")
        }
        return item
    }

    static func deleteJars(named name: String, month: String) async throws {
        let token = try await loginToken()
        let summary = try await request(
            path: "/jars?month=\(month)",
            method: "GET",
            token: token,
            body: nil,
            decode: FixtureEnvelope<FixtureJarMonth>.self
        ).data
        for item in summary.items where item.name == name {
            _ = try await request(
                path: "/jars/\(item.jarID)/months/\(month)",
                method: "DELETE",
                token: token,
                body: nil,
                decode: EmptyFixtureResponse.self
            )
        }
    }

    static func jar(named name: String, month: String) async throws -> FixtureJarItem {
        let token = try await loginToken()
        let summary = try await request(
            path: "/jars?month=\(month)", method: "GET", token: token, body: nil,
            decode: FixtureEnvelope<FixtureJarMonth>.self
        ).data
        guard let item = summary.items.first(where: { $0.name == name }) else {
            throw LocalFixtureError.http(404, "Jar fixture not found: \(name)")
        }
        return item
    }

    static func deleteRecurring(named name: String) async throws {
        let token = try await loginToken()
        let existing = try await request(
            path: "/recurring",
            method: "GET",
            token: token,
            body: nil,
            decode: FixtureEnvelope<[FixtureRecurring]>.self
        ).data
        for item in existing where item.name == name {
            _ = try await request(
                path: "/recurring/\(item.id)",
                method: "DELETE",
                token: token,
                body: nil,
                decode: EmptyFixtureResponse.self
            )
        }
    }

    static func recurring(named name: String) async throws -> FixtureRecurring {
        let token = try await loginToken()
        let existing = try await request(
            path: "/recurring",
            method: "GET",
            token: token,
            body: nil,
            decode: FixtureEnvelope<[FixtureRecurring]>.self
        ).data
        guard let item = existing.first(where: { $0.name == name }) else {
            throw LocalFixtureError.http(404, "Recurring fixture not found: \(name)")
        }
        return item
    }

    static func deleteTravelEvents(named name: String) async throws {
        let token = try await loginToken()
        let existing = try await request(
            path: "/travel/events",
            method: "GET",
            token: token,
            body: nil,
            decode: FixtureEnvelope<[FixtureTravelEvent]>.self
        ).data
        for item in existing where item.name == name {
            _ = try await request(
                path: "/travel/events/\(item.id)",
                method: "DELETE",
                token: token,
                body: nil,
                decode: EmptyFixtureResponse.self
            )
        }
    }

    static func travelEvent(named name: String) async throws -> FixtureTravelEvent {
        let token = try await loginToken()
        let existing = try await request(
            path: "/travel/events",
            method: "GET",
            token: token,
            body: nil,
            decode: FixtureEnvelope<[FixtureTravelEvent]>.self
        ).data
        guard let item = existing.first(where: { $0.name == name }) else {
            throw LocalFixtureError.http(404, "Travel fixture not found: \(name)")
        }
        return item
    }

    static func deleteTransactions(walletID: String) async throws {
        let existing = try await transactions()
        let token = try await loginToken()
        for item in existing where item.walletID == walletID {
            _ = try await request(
                path: "/transactions/\(item.id)",
                method: "DELETE",
                token: token,
                body: nil,
                decode: EmptyFixtureResponse.self
            )
        }
    }

    static func deleteTransferPairs(walletIDs: Set<String>) async throws {
        guard baseURL.scheme == "http", ["127.0.0.1", "::1", "[::1]"].contains(baseURL.host ?? "") else { throw LocalFixtureError.invalidResponse }
        let rows = try await transactions().filter { walletIDs.contains($0.walletID) }
        let pairs = Set(rows.compactMap(\.transferID))
        let token = try await loginToken()
        for pair in pairs {
            _ = try await request(path: "/transactions/transfer/\(pair)", method: "DELETE", token: token,
                body: nil, decode: EmptyFixtureResponse.self)
        }
    }

    static func transactions() async throws -> [FixtureTransaction] {
        let token = try await loginToken()
        return try await request(
            path: "/transactions",
            method: "GET",
            token: token,
            body: nil,
            decode: FixtureEnvelope<[FixtureTransaction]>.self
        ).data
    }

    static func transaction(
        walletID: String,
        categoryID: String,
        amount: Int64,
        note: String,
        type: String = "expense",
        occurredAt: String? = nil,
        includedInReports: Bool = true
    ) async throws -> FixtureTransaction {
        let token = try await loginToken()
        let payload: [String: Any] = [
            "wallet_id": walletID,
            "category_id": categoryID,
            "type": type,
            "amount": amount,
            "occurred_at": occurredAt ?? ISO8601DateFormatter().string(from: Date()),
            "note": note,
            "included_in_reports": includedInReports,
        ]
        return try await request(
            path: "/transactions",
            method: "POST",
            token: token,
            body: JSONSerialization.data(withJSONObject: payload),
            decode: FixtureEnvelope<FixtureTransaction>.self
        ).data
    }

    static func monthSummary(month: String) async throws -> FixtureMonthSummary {
        let token = try await loginToken()
        return try await request(
            path: "/months/\(month)",
            method: "GET",
            token: token,
            body: nil,
            decode: FixtureEnvelope<FixtureMonthSummary>.self
        ).data
    }

    static func currentMonth() -> String {
        let formatter = DateFormatter()
        formatter.calendar = Calendar(identifier: .gregorian)
        formatter.locale = Locale(identifier: "en_US_POSIX")
        formatter.timeZone = TimeZone(identifier: "Asia/Ho_Chi_Minh")!
        formatter.dateFormat = "yyyy-MM"
        return formatter.string(from: Date())
    }

    static func deleteMonthNote(month: String) async throws {
        let token = try await loginToken()
        _ = try await request(
            path: "/months/\(month)/note",
            method: "DELETE",
            token: token,
            body: nil,
            decode: EmptyFixtureResponse.self
        )
    }

    static func portfolioAsset(symbol: String, name: String) async throws -> FixturePortfolioAssetHandle {
        let token = try await loginToken()
        let existing = try await request(
            path: "/portfolio/assets",
            method: "GET",
            token: token,
            body: nil,
            decode: FixtureEnvelope<[FixturePortfolioAssetList]>.self
        ).data
        for item in existing where item.symbol == symbol {
            _ = try await request(
                path: "/portfolio/assets/\(item.id)",
                method: "DELETE",
                token: token,
                body: nil,
                decode: EmptyFixtureResponse.self
            )
        }
        let data = try JSONSerialization.data(withJSONObject: ["symbol": symbol, "name": name])
        let created = try await request(
            path: "/portfolio/assets",
            method: "POST",
            token: token,
            body: data,
            decode: FixtureEnvelope<FixturePortfolioAsset>.self
        ).data
        return FixturePortfolioAssetHandle(token: token, asset: created)
    }

    static func portfolioAssets() async throws -> [FixturePortfolioAssetList] {
        try await request(path: "/portfolio/assets", method: "GET", token: loginToken(), body: nil,
                          decode: FixtureEnvelope<[FixturePortfolioAssetList]>.self).data
    }

    static func deletePortfolioAsset(id: String, token: String) async throws {
        _ = try await request(path: "/portfolio/assets/\(id)", method: "DELETE", token: token,
                              body: nil, decode: EmptyFixtureResponse.self)
    }

    static func delete(_ fixture: FixturePortfolioAssetHandle) async throws {
        _ = try await request(
            path: "/portfolio/assets/\(fixture.asset.id)",
            method: "DELETE",
            token: fixture.token,
            body: nil,
            decode: EmptyFixtureResponse.self
        )
    }

    static func portfolioTrades(for fixture: FixturePortfolioAssetHandle) async throws -> [FixturePortfolioTrade] {
        try await request(
            path: "/portfolio/assets/\(fixture.asset.id)/trades",
            method: "GET",
            token: fixture.token,
            body: nil,
            decode: FixtureEnvelope<[FixturePortfolioTrade]>.self
        ).data
    }

    static func notification(
        named title: String,
        apiBaseURL: URL = baseURL
    ) async throws -> FixtureNotification? {
        let token = try await loginToken(baseURL: apiBaseURL)
        let notifications = try await request(
            path: "/notifications",
            method: "GET",
            token: token,
            body: nil,
            decode: FixtureEnvelope<[FixtureNotification]>.self,
            baseURL: apiBaseURL
        ).data
        return notifications.first { $0.title == title }
    }

    private static func loginToken(baseURL: URL = baseURL) async throws -> String {
        let body = try JSONSerialization.data(withJSONObject: [
            "code": "fixture",
            "code_verifier": "fixture",
            "redirect_uri": "mypocket://oauth/callback",
            "platform": "ios",
        ])
        return try await request(
            path: "/auth/mobile/google/exchange",
            method: "POST",
            token: nil,
            body: body,
            decode: FixtureEnvelope<FixtureLogin>.self,
            baseURL: baseURL
        ).data.token
    }

    private static func deleteWallet(id: String, token: String) async throws {
        _ = try await request(
            path: "/wallets/\(id)",
            method: "DELETE",
            token: token,
            body: nil,
            decode: EmptyFixtureResponse.self
        )
    }

    private static func request<Value: Decodable>(
        path: String,
        method: String,
        token: String?,
        body: Data?,
        decode: Value.Type,
        baseURL requestBaseURL: URL = baseURL
    ) async throws -> Value {
        let rawPath = String(path.dropFirst())
        let url: URL
        if let queryIndex = rawPath.firstIndex(of: "?") {
            let route = String(rawPath[..<queryIndex])
            let query = String(rawPath[rawPath.index(after: queryIndex)...])
            url = URL(string: "\(requestBaseURL.appendingPathComponent(route).absoluteString)?\(query)")!
        } else {
            url = requestBaseURL.appendingPathComponent(rawPath)
        }
        var request = URLRequest(url: url)
        request.httpMethod = method
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        if let token { request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization") }
        request.httpBody = body
        let (data, response) = try await URLSession.shared.data(for: request)
        guard let response = response as? HTTPURLResponse else { throw LocalFixtureError.invalidResponse }
        guard (200..<300).contains(response.statusCode) else {
            throw LocalFixtureError.http(response.statusCode, String(data: data, encoding: .utf8) ?? "")
        }
        if response.statusCode == 204 { return try JSONDecoder().decode(Value.self, from: Data("{}".utf8)) }
        return try JSONDecoder().decode(Value.self, from: data)
    }
}

private struct BudgetList: Decodable, Sendable {
    let items: [FixtureBudget]
}

private struct EmptyFixtureResponse: Decodable {}

// Raw wire fixtures intentionally stay separate from app models: API readback
// must independently verify what the native UI persisted and displayed.
private struct SavingsBudgetSeriesFixture: Decodable, Sendable {
    struct Config: Decodable, Sendable {
        let name: String
        let limit_amount: Int64
        let wallet_id: String?
        let carryover_enabled: Bool
        let effective_start_date: String
    }
    let id: String
    let cadence: String
    let active: Bool
    let version: Int64
    let current_config: Config
    let pending_config: Config?
}

private struct SavingsBudgetPeriodFixture: Decodable, Sendable {
    let id: String
    let name: String
    let series_id: String?
    let limit_amount: Int64
    let effective_limit_amount: Int64?
}

private struct SavingsBudgetListFixture<Item: Decodable & Sendable>: Decodable, Sendable {
    let items: [Item]
}

private struct SavingsReportFixture: Decodable, Sendable {
    struct Goals: Decodable, Sendable {
        let opening_balance: Int64
        let closing_balance: Int64
        let initial_balance_added: Int64
        let deposits: Int64
        let internal_transfers: Int64
    }
    struct Surplus: Decodable, Sendable {
        let income: Int64
        let expense: Int64
        let net: Int64
    }
    let goals: Goals
    let surplus: Surplus
}

private struct SavingsBudgetReportFixture: Decodable, Sendable {
    struct Forecast: Decodable, Sendable {
        let status: String
        let elapsed_days: Int64
        let total_days: Int64
        let projected_spent: Int64?
    }
    struct Day: Decodable, Sendable { let spent: Int64 }
    let base_limit_amount: Int64
    let carryover_amount: Int64
    let effective_limit_amount: Int64
    let spent: Int64
    let remaining: Int64
    let forecast: Forecast
    let daily: [Day]
}

private struct SavingsBudgetPageFixture: Decodable, Sendable {
    let items: [FixtureTransaction]
    let total: Int
}

private extension LocalAPIFixture {
    static func savingsBudgetRead<Value: Decodable & Sendable>(_ path: String, as type: Value.Type) async throws -> Value {
        try await request(path: path, method: "GET", token: loginToken(), body: nil,
                          decode: FixtureEnvelope<Value>.self).data
    }

    static func savingsBudgetTransfer(source: String, destination: String, amount: Int64, note: String) async throws {
        _ = try await request(path: "/transactions/transfer", method: "POST", token: loginToken(),
            body: JSONSerialization.data(withJSONObject: ["source_wallet_id": source, "destination_wallet_id": destination,
                "amount": amount, "occurred_at": ISO8601DateFormatter().string(from: Date().addingTimeInterval(-60)), "note": note]),
            decode: FixtureEnvelope<[FixtureTransaction]>.self)
    }
}

private extension XCUIElement {
    func clearAndType(_ text: String) {
        // SwiftUI can leave a different field focused after a scroll/reveal.
        // Focus the target before inspecting or mutating its value so the
        // helper never dispatches keyboard events to the wrong control.
        tap()
        let existingValue = value as? String
        let isPlaceholder = existingValue == placeholderValue
        if existingValue?.isEmpty ?? true || isPlaceholder {
            typeText(text)
            return
        }
        let clearButton = XCUIApplication().buttons["clear-field-\(label)"]
        if clearButton.waitForExistence(timeout: 2) {
            clearButton.tap()
            if (value as? String)?.isEmpty == true {
                typeText(text)
                return
            }
        }
        press(forDuration: 1.0)
        let selectAll = XCUIApplication().menuItems.matching(
            NSPredicate(format: "label ==[c] 'Select All' OR label ==[c] 'Chọn tất cả'")
        ).firstMatch
        if selectAll.waitForExistence(timeout: 2) {
            selectAll.tap()
            typeText(XCUIKeyboardKey.delete.rawValue)
            typeText(text)
            return
        }
        // Tap near the trailing edge so the insertion point is at the end,
        // then backspace past the complete existing value. This is more
        // deterministic than relying on the localized edit menu.
        coordinate(withNormalizedOffset: CGVector(dx: 0.98, dy: 0.5)).tap()
        for _ in 0..<(existingValue?.count ?? 0) + 32 {
            typeText(XCUIKeyboardKey.delete.rawValue)
        }
        typeText(text)
    }
}
#endif

@MainActor
final class LaunchTests: XCTestCase {
    func testOfflineRejectedTransferRequiresChangedConfirmationAndKeepsReportTotals() async throws {
#if LOCAL_API_UI && OFFLINE_FAULT_UI
        try await runOfflineTransferRecovery(mode: "reject422", editRejectedTransfer: true)
#else
        throw XCTSkip("Requires dedicated loopback fault proxy and OFFLINE_FAULT_UI build")
#endif
    }

    func testOfflineSavingsRejectedTransferCanBeEditedAndConfirmedWithoutChangingReportTotals() async throws {
#if LOCAL_API_UI && OFFLINE_FAULT_UI
        try await runOfflineTransferRecovery(mode: "reject422", verifyGoalScope: true, editRejectedTransfer: true)
#else
        throw XCTSkip("Requires dedicated loopback fault proxy and OFFLINE_FAULT_UI build")
#endif
    }

    func testOfflineSavingsScopeKeepsRecoveryAccessibleAndRefreshesConfirmedBalance() async throws {
#if LOCAL_API_UI && OFFLINE_FAULT_UI
        try await runOfflineTransferRecovery(mode: "reject503", verifyGoalScope: true)
#else
        throw XCTSkip("Requires dedicated loopback fault proxy and OFFLINE_FAULT_UI build")
#endif
    }

    func testOfflineConflictDraftRequiresChangedConfirmationAndPreservesPreviousDay() async throws {
#if LOCAL_API_UI && OFFLINE_FAULT_UI
        try await runOfflineCreateRecovery(mode: "reject409", recovery: "edit")
#else
        throw XCTSkip("Requires dedicated loopback fault proxy and OFFLINE_FAULT_UI build")
#endif
    }

    func testOfflineTransfer503RestartRetryCreatesOneAtomicExcludedPair() async throws {
#if LOCAL_API_UI && OFFLINE_FAULT_UI
        try await runOfflineTransferRecovery(mode: "reject503")
#else
        throw XCTSkip("Requires dedicated loopback fault proxy and OFFLINE_FAULT_UI build")
#endif
    }

    func testOfflineTransferCommittedLostResponseReplaysOriginalPairAndBalancesOnce() async throws {
#if LOCAL_API_UI && OFFLINE_FAULT_UI
        try await runOfflineTransferRecovery(mode: "commit-then-drop")
#else
        throw XCTSkip("Requires dedicated loopback fault proxy and OFFLINE_FAULT_UI build")
#endif
    }

    func testOfflineRejectedDraftRequiresChangedConfirmationAndPreservesPreviousDay() async throws {
#if LOCAL_API_UI && OFFLINE_FAULT_UI
        try await runOfflineCreateRecovery(mode: "reject422", recovery: "edit")
#else
        throw XCTSkip("Requires dedicated loopback fault proxy and OFFLINE_FAULT_UI build")
#endif
    }

    func testOfflineDiscardRejectedDraftRequiresConfirmationAndNeverWritesServer() async throws {
#if LOCAL_API_UI && OFFLINE_FAULT_UI
        try await runOfflineCreateRecovery(mode: "reject422", recovery: "discard")
#else
        throw XCTSkip("Requires dedicated loopback fault proxy and OFFLINE_FAULT_UI build")
#endif
    }

    func testOfflineDiscardUncertainCommittedDraftNeverDeletesServerTransaction() async throws {
#if LOCAL_API_UI && OFFLINE_FAULT_UI
        try await runOfflineCreateRecovery(mode: "commit-then-drop", recovery: "discard")
#else
        throw XCTSkip("Requires dedicated loopback fault proxy and OFFLINE_FAULT_UI build")
#endif
    }

    func testOffline503DoesNotSendInBackgroundAndResumesAutomaticallyInForeground() async throws {
#if LOCAL_API_UI && OFFLINE_FAULT_UI
        try await runOfflineCreateRecovery(mode: "reject503", resumeViaForeground: true)
#else
        throw XCTSkip("Requires dedicated loopback fault proxy and OFFLINE_FAULT_UI build")
#endif
    }

    func testOffline503DraftSurvivesRestartAndManualRetryCreatesExactlyOneTransaction() async throws {
#if LOCAL_API_UI && OFFLINE_FAULT_UI
        try await runOfflineCreateRecovery(mode: "reject503")
#else
        throw XCTSkip("Requires dedicated loopback fault proxy and OFFLINE_FAULT_UI build")
#endif
    }

    func testOfflineLostCommittedResponseReplaysSameKeyAfterRestartWithoutDuplicateTransaction() async throws {
#if LOCAL_API_UI && OFFLINE_FAULT_UI
        try await runOfflineCreateRecovery(mode: "commit-then-drop")
#else
        throw XCTSkip("Requires dedicated loopback fault proxy and OFFLINE_FAULT_UI build")
#endif
    }

#if LOCAL_API_UI && OFFLINE_FAULT_UI
    private func runOfflineTransferRecovery(mode: String, verifyGoalScope: Bool = false, editRejectedTransfer: Bool = false) async throws {
        let suffix = String(UUID().uuidString.prefix(8))
        let source = try await LocalAPIFixture.wallet(name: "Fixture Offline Source \(suffix)", type: "basic")
        let destination = try await LocalAPIFixture.wallet(name: "Fixture Offline Destination \(suffix)", type: verifyGoalScope ? "goal" : "basic")
        addTeardownBlock {
            try? await OfflineFaultFixture.configure(mode: "pass", walletID: source.wallet.id)
            try? await LocalAPIFixture.deleteTransferPairs(walletIDs: [source.wallet.id, destination.wallet.id])
            try? await LocalAPIFixture.delete(source)
            try? await LocalAPIFixture.delete(destination)
        }
        try await OfflineFaultFixture.configure(mode: mode, walletID: source.wallet.id, reset: true)
        let timezone = try await LocalAPIFixture.ownerTimezone()
        let monthFormatter = DateFormatter(); monthFormatter.timeZone = TimeZone(identifier: timezone); monthFormatter.dateFormat = "yyyy-MM"
        let month = monthFormatter.string(from: Date())
        let reportBefore = try await LocalAPIFixture.monthSummary(month: month)
        let note = "Offline transfer \(suffix)"
        let app = XCUIApplication(); app.launchArguments = ["-UITestAuthenticatedAPI"]; app.launch()
        XCTAssertTrue(app.buttons["bottom-tab-transactions"].waitForExistence(timeout: 10))
        app.buttons["bottom-tab-transactions"].tap()
        let options = app.buttons["Tùy chọn giao dịch"]
        XCTAssertTrue(options.waitForExistence(timeout: 10)); reveal(options, in: app); options.tap()
        app.buttons["Chuyển tiền đến ví khác"].tap()
        let from = app.buttons["transfer-source-wallet"]
        XCTAssertTrue(from.waitForExistence(timeout: 10)); from.tap(); app.buttons[source.wallet.name].tap()
        app.buttons["transfer-destination-wallet"].tap(); app.buttons[destination.wallet.name].tap()
        let amount = app.textFields["Số tiền"]
        reveal(amount, in: app); amount.clearAndType("12500"); dismissKeyboard(app)
        let noteField = app.textFields["Ghi chú"]
        reveal(noteField, in: app); noteField.clearAndType(note); dismissKeyboard(app)
        let save = app.buttons["Xác nhận chuyển"]
        reveal(save, in: app); save.tap(); XCTAssertTrue(save.waitForNonExistence(timeout: 20))
        if verifyGoalScope { selectOfflineGoalScope(app, name: destination.wallet.name) }
        let section = app.descendants(matching: .any)["financial-outbox-section"].firstMatch
        let draftNote = section.staticTexts[note]
        XCTAssertTrue(draftNote.waitForExistence(timeout: 10), "Savings scope must expose the same pending recovery queue")
        guard draftNote.exists else { return }
        reveal(draftNote, in: app)
        if verifyGoalScope {
            XCTAssertTrue(app.staticTexts["100.000 ₫"].exists, "Pending transfer must not inflate confirmed savings")
            XCTAssertTrue(app.staticTexts["900.000 ₫"].exists, "Pending transfer must not change remaining target")
        }
        let items = section.descendants(matching: .any).matching(NSPredicate(format: "identifier BEGINSWITH 'outbox-item-'"))
        let item = try XCTUnwrap(items.allElementsBoundByIndex.first(where: { $0.staticTexts[note].exists }))
        let intentID = item.identifier
        let before = try await LocalAPIFixture.transactions().filter { [source.wallet.id, destination.wallet.id].contains($0.walletID) }
        XCTAssertEqual(before.count, mode == "commit-then-drop" ? 2 : 0)
        let originalIDs = Set(before.map(\.id))
        attachScreenshot(app, named: "offline-transfer-\(mode)-pending")
        app.terminate(); app.launch()
        XCTAssertTrue(app.buttons["bottom-tab-transactions"].waitForExistence(timeout: 10))
        app.buttons["bottom-tab-transactions"].tap()
        if verifyGoalScope { selectOfflineGoalScope(app, name: destination.wallet.name) }
        openOfflineDraftMenu(app, intentID: intentID)
        if editRejectedTransfer {
            XCTAssertFalse(app.buttons["Thử lại ngay"].exists)
            app.buttons["Sửa bản nháp"].tap()
            let confirm = app.buttons["Xác nhận bản sửa"]
            XCTAssertTrue(confirm.waitForExistence(timeout: 10)); reveal(confirm, in: app); confirm.tap()
            XCTAssertTrue(app.staticTexts["Hãy thay đổi thông tin hợp lệ và kiểm tra đúng tài khoản trước khi xác nhận bản sửa."].waitForExistence(timeout: 5))
            let unchanged = try await OfflineFaultFixture.stats(); XCTAssertEqual(unchanged.keys.count, 1)
            let editedAmount = app.textFields["Số tiền"]
            reveal(editedAmount, in: app); editedAmount.clearAndType("22500"); dismissKeyboard(app)
            try await OfflineFaultFixture.configure(mode: "pass", walletID: source.wallet.id)
            reveal(confirm, in: app); confirm.tap(); XCTAssertTrue(confirm.waitForNonExistence(timeout: 15))
        } else {
        let retry = app.buttons["Thử lại ngay"]
        XCTAssertTrue(retry.waitForExistence(timeout: 5))
        try await OfflineFaultFixture.configure(mode: "pass", walletID: source.wallet.id)
        retry.tap()
        }
        XCTAssertTrue(app.descendants(matching: .any)[intentID].firstMatch.waitForNonExistence(timeout: 20))
        let rows = try await LocalAPIFixture.transactions().filter { [source.wallet.id, destination.wallet.id].contains($0.walletID) }
        XCTAssertEqual(rows.count, 2)
        XCTAssertEqual(Set(rows.compactMap(\.transferID)).count, 1)
        XCTAssertTrue(rows.allSatisfy { $0.transferID != nil && $0.includedInReports == false && $0.amount == (editRejectedTransfer ? 22500 : 12500) && $0.note == note })
        XCTAssertEqual(rows.first(where: { $0.type == "expense" })?.walletID, source.wallet.id)
        XCTAssertEqual(rows.first(where: { $0.type == "income" })?.walletID, destination.wallet.id)
        if mode == "commit-then-drop" { XCTAssertEqual(Set(rows.map(\.id)), originalIDs) }
        let sourceWallet = try await LocalAPIFixture.wallet(named: source.wallet.name)
        let destinationWallet = try await LocalAPIFixture.wallet(named: destination.wallet.name)
        XCTAssertEqual(sourceWallet.currentBalance, editRejectedTransfer ? 77500 : 87500)
        XCTAssertEqual(destinationWallet.currentBalance, editRejectedTransfer ? 122500 : 112500)
        let reportAfter = try await LocalAPIFixture.monthSummary(month: month)
        XCTAssertEqual(reportAfter.income, reportBefore.income)
        XCTAssertEqual(reportAfter.expense, reportBefore.expense)
        XCTAssertEqual(reportAfter.transactionCount, reportBefore.transactionCount, "Transfer pair must not enter monthly report aggregates")
        let stats = try await OfflineFaultFixture.stats(); XCTAssertEqual(stats.keys.count, editRejectedTransfer ? 2 : 1)
        if mode == "commit-then-drop" { XCTAssertGreaterThan(stats.dropped, 0); XCTAssertTrue(stats.upstreamStatuses.contains(200)) }
        else { XCTAssertGreaterThan(stats.blocked, 0) }
        if verifyGoalScope {
            let balance = app.staticTexts[editRejectedTransfer ? "122.500 ₫" : "112.500 ₫"]
            XCTAssertTrue(balance.waitForExistence(timeout: 10), "Receipt must reload savings balance without manual refresh")
            XCTAssertTrue(app.staticTexts[editRejectedTransfer ? "877.500 ₫" : "887.500 ₫"].exists, "Remaining goal must derive from the confirmed wallet balance")
            let cta = app.buttons["wallet-detail-add-transaction"]
            for _ in 0..<4 where balance.frame.maxY >= cta.frame.minY { app.scrollViews.firstMatch.swipeUp(velocity: .slow) }
            XCTAssertLessThanOrEqual(balance.frame.maxY, cta.frame.minY, "Confirmed savings balance must be reachable above the fixed CTA")
        } else {
            openConfirmedOfflineDetail(app, rowID: try XCTUnwrap(rows.first(where: { $0.type == "expense" })?.id), note: note)
        }
        attachScreenshot(app, named: "offline-transfer-\(mode)-applied")
    }

    private func selectOfflineGoalScope(_ app: XCUIApplication, name: String) {
        let goalSelector = app.buttons["wallet-detail-goal-wallet-selector"]
        if goalSelector.exists && goalSelector.label == "Ví: \(name)" { return }
        let selector = goalSelector.exists ? goalSelector : app.buttons["transaction-wallet-selector"]
        XCTAssertTrue(selector.waitForExistence(timeout: 10)); reveal(selector, in: app); selector.tap()
        app.buttons[name].tap()
        XCTAssertTrue(goalSelector.waitForExistence(timeout: 10))
        XCTAssertEqual(goalSelector.label, "Ví: \(name)", "Recovery checks must run in the intended savings wallet, not another retained scope")
    }

    private func runOfflineCreateRecovery(mode: String, resumeViaForeground: Bool = false, recovery: String? = nil) async throws {
        let suffix = String(UUID().uuidString.prefix(8))
        let fixture = try await LocalAPIFixture.wallet(name: "Fixture Offline \(suffix)", type: "basic")
        addTeardownBlock {
            try? await OfflineFaultFixture.configure(mode: "pass", walletID: fixture.wallet.id)
            try? await LocalAPIFixture.deleteTransactions(walletID: fixture.wallet.id)
            try? await LocalAPIFixture.delete(fixture)
        }
        let category = try await LocalAPIFixture.category(kind: "expense", walletID: fixture.wallet.id)
        try await OfflineFaultFixture.configure(mode: mode, walletID: fixture.wallet.id, reset: true)
        let note = "Offline draft \(suffix)"
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI", "-mypocket.privacy.masked", "NO"]
        app.launch()
        openWallets(app)
        XCTAssertTrue(app.staticTexts[fixture.wallet.name].waitForExistence(timeout: 10))
        app.buttons["Chi tiết"].firstMatch.tap()
        let add = app.buttons["wallet-detail-add-transaction"]
        XCTAssertTrue(add.waitForExistence(timeout: 10)); add.tap()
        let picker = app.buttons["transaction-category-row"]
        XCTAssertTrue(picker.waitForExistence(timeout: 10)); picker.tap()
        app.textFields["category-picker-search"].clearAndType(category.name)
        let choice = app.buttons["category-choice-\(category.id)"]
        XCTAssertTrue(choice.waitForExistence(timeout: 5)); reveal(choice, in: app); choice.tap()
        let amount = app.textFields["Số tiền"]
        amount.clearAndType("12500"); dismissKeyboard(app)
        app.buttons["transaction-note-row"].tap()
        let text = app.textViews.firstMatch
        XCTAssertTrue(text.waitForExistence(timeout: 5)); text.clearAndType(note); dismissKeyboard(app)
        app.buttons["transaction-note-save"].tap()
        let timezone = try await LocalAPIFixture.ownerTimezone()
        var calendar = Calendar(identifier: .gregorian)
        calendar.timeZone = try XCTUnwrap(TimeZone(identifier: timezone))
        let previousDay = try XCTUnwrap(calendar.date(byAdding: .day, value: -1, to: Date()))
        if recovery == "edit" {
            let previous = app.buttons["Ngày trước"]
            reveal(previous, in: app); previous.tap()
        }
        let save = app.buttons["Lưu giao dịch"]
        reveal(save, in: app); save.tap()
        XCTAssertTrue(save.waitForNonExistence(timeout: 20), "Durable queue acceptance should close form")
        app.buttons["bottom-tab-transactions"].tap()
        let section = app.descendants(matching: .any)["financial-outbox-section"].firstMatch
        XCTAssertTrue(section.waitForExistence(timeout: 10))
        let draftNote = section.staticTexts[note]
        XCTAssertTrue(draftNote.waitForExistence(timeout: 5), "Queue needs identifying note, not indistinguishable amount-only drafts")
        guard draftNote.exists else { return }
        reveal(draftNote, in: app)
        let items = section.descendants(matching: .any).matching(NSPredicate(format: "identifier BEGINSWITH 'outbox-item-'"))
        let item = try XCTUnwrap(items.allElementsBoundByIndex.first(where: { $0.staticTexts[note].exists }))
        let intentID = item.identifier
        let before = try await LocalAPIFixture.transactions().filter { $0.walletID == fixture.wallet.id }
        XCTAssertEqual(before.count, mode == "commit-then-drop" ? 1 : 0)
        attachScreenshot(app, named: "offline-\(mode)-pending-before-restart")
        if let recovery {
            app.terminate(); app.launch()
            XCTAssertTrue(app.buttons["bottom-tab-transactions"].waitForExistence(timeout: 10))
            app.buttons["bottom-tab-transactions"].tap()
            if recovery == "edit" {
                try await editRejectedOfflineDraft(app, intentID: intentID, fixture: fixture,
                    category: category, note: note, expectedDay: previousDay, timezone: timezone)
            } else {
                try await discardOfflineDraft(app, intentID: intentID, fixture: fixture, note: note, wasCommitted: mode == "commit-then-drop")
            }
            return
        }
        if resumeViaForeground {
            XCUIDevice.shared.press(.home)
            XCTAssertTrue(app.wait(for: .runningBackground, timeout: 5))
            try await OfflineFaultFixture.configure(mode: "pass", walletID: fixture.wallet.id)
            for _ in 0..<4 {
                try await Task.sleep(for: .milliseconds(500))
                let backgroundRows = try await LocalAPIFixture.transactions().filter { $0.walletID == fixture.wallet.id }
                XCTAssertTrue(backgroundRows.isEmpty, "Background app must not send a financial write after server recovery")
            }
            app.activate()
        } else {
            app.terminate(); app.launch()
            XCTAssertTrue(app.buttons["bottom-tab-transactions"].waitForExistence(timeout: 10))
            app.buttons["bottom-tab-transactions"].tap()
            let restored = app.descendants(matching: .any)[intentID].firstMatch
            XCTAssertTrue(restored.waitForExistence(timeout: 10), "Same intent must survive app restart without uninstall")
            let menu = restored.buttons["Thao tác bản nháp"]
            XCTAssertTrue(menu.waitForExistence(timeout: 10), "Recovery actions must be reachable from the compact item's overflow menu")
            guard menu.exists else { return }
            reveal(menu, in: app)
            // The separately blocked floating Feedback can intercept a visible menu.
            // Scroll the real row clear; do not hide the overlay or treat placement as passed.
            let feedback = app.buttons["Gửi phản hồi"].firstMatch
            if feedback.exists && menu.frame.intersects(feedback.frame) {
                app.scrollViews["shell-content-viewport"].swipeUp(velocity: .slow)
                reveal(menu, in: app)
            }
            menu.tap()
            let retry = app.buttons["Thử lại ngay"]
            XCTAssertTrue(retry.waitForExistence(timeout: 5))
            try await OfflineFaultFixture.configure(mode: "pass", walletID: fixture.wallet.id)
            retry.tap()
        }
        let restored = app.descendants(matching: .any)[intentID].firstMatch
        XCTAssertTrue(restored.waitForNonExistence(timeout: 20), "Receipt-confirmed item should leave pending queue")
        let after = try await LocalAPIFixture.transactions().filter { $0.walletID == fixture.wallet.id }
        XCTAssertEqual(after.count, 1); XCTAssertEqual(after.first?.amount, 12500)
        XCTAssertEqual(after.first?.note, note); XCTAssertEqual(after.first?.categoryID, category.id)
        let stats = try await OfflineFaultFixture.stats()
        XCTAssertEqual(stats.keys.count, 1, "All attempts/restart/replay must use one key")
        if mode == "commit-then-drop" {
            XCTAssertGreaterThan(stats.dropped, 0); XCTAssertTrue(stats.upstreamStatuses.contains(200))
        } else { XCTAssertGreaterThan(stats.blocked, 0) }
        attachScreenshot(app, named: "offline-\(mode)-applied")
    }

    private func openOfflineDraftMenu(_ app: XCUIApplication, intentID: String) {
        let item = app.descendants(matching: .any)[intentID].firstMatch
        XCTAssertTrue(item.waitForExistence(timeout: 10))
        let menu = item.buttons["Thao tác bản nháp"]
        XCTAssertTrue(menu.waitForExistence(timeout: 10))
        reveal(menu, in: app)
        let feedback = app.buttons["Gửi phản hồi"].firstMatch
        if feedback.exists && menu.frame.intersects(feedback.frame) {
            app.scrollViews["shell-content-viewport"].swipeUp(velocity: .slow)
            reveal(menu, in: app)
        }
        menu.tap()
    }

    private func openConfirmedOfflineDetail(_ app: XCUIApplication, rowID: String, note: String) {
        let row = app.buttons["transaction-row-\(rowID)"]
        XCTAssertTrue(row.waitForExistence(timeout: 10), "Receipt must invalidate/refetch the actual ledger row")
        reveal(row, in: app)
        let feedback = app.buttons["Gửi phản hồi"].firstMatch
        if feedback.exists && row.frame.intersects(feedback.frame) {
            app.scrollViews["shell-content-viewport"].swipeUp(velocity: .slow)
            reveal(row, in: app)
        }
        row.tap()
        XCTAssertTrue(app.staticTexts["Chi tiết giao dịch"].waitForExistence(timeout: 5))
        let detailContainers = app.descendants(matching: .any).matching(identifier: "transaction-detail-overlay")
        XCTAssertEqual(detailContainers.count, 1, "Screen identity must belong to one container, not leak into child buttons/cards")
        let detail = detailContainers.firstMatch
        XCTAssertTrue(detail.staticTexts["Ghi chú: \(note)"].waitForExistence(timeout: 5), "Verify detail content, not the underlying ledger label")
    }

    private func editRejectedOfflineDraft(_ app: XCUIApplication, intentID: String, fixture: FixtureWalletHandle,
        category: FixtureCategory, note: String, expectedDay: Date, timezone: String) async throws {
        openOfflineDraftMenu(app, intentID: intentID)
        XCTAssertFalse(app.buttons["Thử lại ngay"].exists, "Permanent rejection is not a transient retry")
        app.buttons["Sửa bản nháp"].tap()
        let confirm = app.buttons["Xác nhận bản sửa"]
        XCTAssertTrue(confirm.waitForExistence(timeout: 10)); reveal(confirm, in: app); confirm.tap()
        let error = app.staticTexts["Hãy thay đổi thông tin hợp lệ và kiểm tra đúng tài khoản trước khi xác nhận bản sửa."]
        XCTAssertTrue(error.waitForExistence(timeout: 5), "Unchanged draft must not mint another intent/key")
        XCTAssertTrue(confirm.exists)
        var stats = try await OfflineFaultFixture.stats()
        XCTAssertEqual(stats.keys.count, 1)
        let amount = app.textFields["Số tiền"]
        reveal(amount, in: app); amount.clearAndType("22500"); dismissKeyboard(app)
        try await OfflineFaultFixture.configure(mode: "pass", walletID: fixture.wallet.id)
        reveal(confirm, in: app); confirm.tap()
        XCTAssertTrue(confirm.waitForNonExistence(timeout: 15))
        let old = app.descendants(matching: .any)[intentID].firstMatch
        XCTAssertTrue(old.waitForNonExistence(timeout: 10))
        let rows = try await LocalAPIFixture.transactions().filter { $0.walletID == fixture.wallet.id }
        XCTAssertEqual(rows.count, 1); XCTAssertEqual(rows.first?.amount, 22500)
        XCTAssertEqual(rows.first?.note, note); XCTAssertEqual(rows.first?.categoryID, category.id)
        let formatter = DateFormatter()
        formatter.calendar = Calendar(identifier: .gregorian); formatter.timeZone = TimeZone(identifier: timezone)
        formatter.dateFormat = "yyyy-MM-dd"
        let iso = ISO8601DateFormatter(); iso.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        let wire = try XCTUnwrap(rows.first?.occurredAt)
        let parsed = iso.date(from: wire) ?? { iso.formatOptions = [.withInternetDateTime]; return iso.date(from: wire) }()
        XCTAssertEqual(formatter.string(from: try XCTUnwrap(parsed)), formatter.string(from: expectedDay), "Editing amount must preserve the selected account-local day")
        stats = try await OfflineFaultFixture.stats(); XCTAssertEqual(stats.keys.count, 2)
        openConfirmedOfflineDetail(app, rowID: try XCTUnwrap(rows.first?.id), note: note)
        attachScreenshot(app, named: "offline-rejected-edit-applied")
    }

    private func discardOfflineDraft(_ app: XCUIApplication, intentID: String, fixture: FixtureWalletHandle,
        note: String, wasCommitted: Bool) async throws {
        openOfflineDraftMenu(app, intentID: intentID)
        app.buttons["Bỏ bản nháp"].tap()
        let warning = app.alerts["Bỏ bản nháp trên thiết bị?"]
        XCTAssertTrue(warning.waitForExistence(timeout: 5))
        XCTAssertTrue(warning.staticTexts.matching(NSPredicate(format: "label CONTAINS %@", "không xóa giao dịch đã lưu")).firstMatch.exists)
        warning.buttons["Hủy"].tap()
        let original = app.descendants(matching: .any)[intentID].firstMatch
        XCTAssertTrue(original.waitForExistence(timeout: 5), "Cancel must preserve the local intent")
        openOfflineDraftMenu(app, intentID: intentID)
        app.buttons["Bỏ bản nháp"].tap()
        XCTAssertTrue(warning.waitForExistence(timeout: 5)); warning.buttons["Bỏ bản nháp"].tap()
        XCTAssertTrue(original.waitForNonExistence(timeout: 10))
        try await OfflineFaultFixture.configure(mode: "pass", walletID: fixture.wallet.id)
        app.terminate(); app.launch()
        XCTAssertTrue(app.buttons["bottom-tab-transactions"].waitForExistence(timeout: 10))
        app.buttons["bottom-tab-transactions"].tap()
        XCTAssertFalse(original.exists, "Discarded local row must not resume after restart")
        let rows = try await LocalAPIFixture.transactions().filter { $0.walletID == fixture.wallet.id }
        XCTAssertEqual(rows.count, wasCommitted ? 1 : 0)
        if wasCommitted { XCTAssertEqual(rows.first?.amount, 12500); XCTAssertEqual(rows.first?.note, note) }
        let wallet = try await LocalAPIFixture.wallet(named: fixture.wallet.name)
        XCTAssertEqual(wallet.currentBalance, wasCommitted ? 87500 : 100000)
        let stats = try await OfflineFaultFixture.stats(); XCTAssertEqual(stats.keys.count, 1)
        attachScreenshot(app, named: "offline-discard-\(wasCommitted ? "committed" : "rejected")")
    }
#endif

    private func attachScreenshot(_ app: XCUIApplication, named name: String) {
        let attachment = XCTAttachment(screenshot: app.screenshot())
        attachment.name = name
        attachment.lifetime = .keepAlways
        add(attachment)
    }

    private func openMore(_ app: XCUIApplication) {
        let accountTab = app.buttons["bottom-tab-account"]
        XCTAssertTrue(accountTab.waitForExistence(timeout: 10))
        accountTab.tap()
        XCTAssertTrue(app.staticTexts["screen-title-Thêm"].waitForExistence(timeout: 10))
    }

    private func openWallets(_ app: XCUIApplication) {
        openMore(app)
        let walletsEntry = app.buttons["more-destination-wallets"]
        XCTAssertTrue(walletsEntry.waitForExistence(timeout: 10))
        reveal(walletsEntry, in: app)
        XCTAssertTrue(walletsEntry.isHittable)
        walletsEntry.tap()
        XCTAssertTrue(app.staticTexts["screen-title-Chọn Ví"].waitForExistence(timeout: 10))
    }

    private func dismissKeyboard(_ app: XCUIApplication) {
        let toolbarDone = app.buttons["Xong"]
        if toolbarDone.exists {
            toolbarDone.tap()
            return
        }
        let key = app.keyboards.buttons.matching(
            NSPredicate(format: "label ==[c] 'return' OR label ==[c] 'done' OR label ==[c] 'xong'")
        ).firstMatch
        if key.exists { key.tap() }
    }

    private func reveal(_ element: XCUIElement, in app: XCUIApplication) {
        let window = app.windows.firstMatch
        let viewport = app.descendants(matching: .any)["shell-content-viewport"].firstMatch
        func fail(_ reason: String) {
            let geometry = XCTAttachment(string: "\(reason)\nwindow=\(window.frame)\nviewport=\(viewport.exists ? String(describing: viewport.frame) : "absent")\ntarget=\(element.exists ? String(describing: element.frame) : "absent")\nscrolls=\(app.scrollViews.allElementsBoundByIndex.map { "\($0.identifier):\($0.frame)" })")
            geometry.name = "reveal-container-geometry"
            geometry.lifetime = .keepAlways
            add(geometry)
            XCTFail(reason)
        }
        for attempt in 0...20 {
            // A global firstMatch can be the keyboard's QuickType ScrollView.
            // Resolve the target's actual content ancestor, never keyboard UI.
            let scrollView: XCUIElement?
            if element.exists {
                let identity = element.identifier.isEmpty ? element.label : element.identifier
                let containers = identity.isEmpty ? [] : app.scrollViews.containing(
                    element.elementType, identifier: identity
                ).allElementsBoundByIndex
                if containers.count <= 1 {
                    scrollView = containers.first
                } else {
                    // A SwiftUI form may expose both the shell viewport and
                    // the nested editor viewport. The smallest visible
                    // containing scroll view is the innermost owner that can
                    // reveal the target without moving an unrelated parent.
                    // Equal-sized candidates remain an explicit failure so
                    // this helper never guesses between peer surfaces.
                    let visible = containers.filter { candidate in
                        let frame = candidate.frame
                        return frame.width > 0 && frame.height > 16 && window.frame.intersects(frame)
                    }
                    let ranked = visible.sorted { lhs, rhs in
                        lhs.frame.width * lhs.frame.height < rhs.frame.width * rhs.frame.height
                    }
                    guard let innermost = ranked.first else {
                        fail("Ambiguous scroll containers for \(identity)"); return
                    }
                    let area = innermost.frame.width * innermost.frame.height
                    let tied = ranked.dropFirst().contains { candidate in
                        abs(candidate.frame.width * candidate.frame.height - area) < 1
                    }
                    guard !tied else {
                        fail("Ambiguous scroll containers for \(identity)"); return
                    }
                    scrollView = innermost
                }
            } else if viewport.exists {
                let containers = viewport.elementType == .scrollView
                    ? [viewport] : viewport.scrollViews.allElementsBoundByIndex
                guard containers.count == 1 else {
                    fail("Missing lazy target has no unique app content container"); return
                }
                scrollView = containers[0]
            } else {
                fail("Missing lazy target and app content viewport"); return
            }
            // Pinned controls need no scroll ancestor if already in bounds.
            if scrollView == nil, element.exists {
                let frame = element.frame
                let keyboard = app.keyboards.firstMatch
                guard window.frame.contains(frame), frame.width > 0, frame.height > 0,
                      element.isHittable,
                      !keyboard.exists || !keyboard.frame.intersects(frame) else {
                    fail("Pinned control is not visible/hittable in the keyboard-free app window"); return
                }
                return
            }
            let bounds = (scrollView?.frame ?? window.frame).intersection(window.frame)
            guard !bounds.isNull && bounds.height > 16 else {
                fail("Content container has no usable viewport"); return
            }
            let topContent = bounds.minY + 8
            let bottomContent = bounds.maxY - 8
            func scroll(up: Bool) {
                // Recheck after the twentieth drag; never declare timeout
                // before observing the position produced by the last drag.
                guard attempt < 20, let scrollView else { return }
                // Full-page swipes overshoot a field in a keyboard-compressed
                // editor. Move a bounded fraction of this content viewport.
                let start = scrollView.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: up ? 0.7 : 0.3))
                let end = scrollView.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: up ? 0.3 : 0.7))
                start.press(forDuration: 0.01, thenDragTo: end)
            }
            guard element.exists else { scroll(up: true); continue }
            let frame = element.frame
            // SwiftUI buttons inside a long ScrollView can report `isHittable`
            // false while their rendered hit shape is visible (the parent
            // scroll container owns the accessibility frame). Bounds are the
            // stable signal for this helper; the following tap/assertion still
            // verifies the control can actually be activated.
            if frame.minY >= topContent && frame.maxY <= bottomContent {
                return
            }
            guard scrollView != nil else {
                fail("Offscreen control has no identified scroll ancestor"); return
            }
            scroll(up: frame.minY >= topContent)
        }
        fail("Control did not enter its content viewport after bounded scrolling")
    }

    private func moreEntry(_ id: String, in app: XCUIApplication) -> XCUIElement {
        let scrollView = app.scrollViews.firstMatch
        // Returning from a secondary screen intentionally preserves the menu
        // offset. Start each route assertion at the menu's top, then let the
        // shared reveal helper move to the requested section.
        for _ in 0..<8 { scrollView.swipeDown() }
        let entry = app.buttons[id].firstMatch
        if entry.waitForExistence(timeout: 2) { return entry }
        for _ in 0..<16 {
            scrollView.swipeUp()
            if entry.exists { return entry }
        }
        return entry
    }

    private func moreEntryInScreenOrder(_ id: String, in app: XCUIApplication) -> XCUIElement {
        let scrollView = app.scrollViews.firstMatch
        let entry = app.buttons[id].firstMatch
        // This helper is only for a sweep ordered exactly like MoreView's
        // sections. Reuse the preserved scroll position instead of swiping to
        // the top before every route; each next target is at or below the last.
        for _ in 0..<16 {
            if entry.exists {
                reveal(entry, in: app)
                if entry.exists && entry.isHittable { return entry }
            }
            scrollView.swipeUp()
        }
        return entry
    }

    func testAppLaunches() throws {
        let app = XCUIApplication()
        app.launchArguments = ["-UITestUnauthenticated"]
        app.launch()
        XCTAssertTrue(
            app.buttons["Đăng nhập"].waitForExistence(timeout: 5),
            "Auth screen should expose the email sign-in action after launch."
        )
        XCTAssertTrue(app.staticTexts["MyPocket"].waitForExistence(timeout: 2))
        XCTAssertTrue(app.staticTexts["Đăng nhập bằng email và mật khẩu để tiếp tục."].exists)
        XCTAssertTrue(app.textFields["Email"].exists)
        XCTAssertTrue(app.secureTextFields["Mật khẩu"].exists)
        XCTAssertFalse(app.buttons["auth-google"].exists)
        XCTAssertFalse(app.staticTexts["hoặc"].exists)
    }

    func testAuthenticatedShellExposesPrimaryRoutes() throws {
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated"]
        app.launch()
        XCTAssertTrue(app.buttons["Tổng quan"].waitForExistence(timeout: 5))
        XCTAssertTrue(app.buttons["Giao dịch"].exists)
        XCTAssertTrue(app.buttons["bottom-tab-account"].exists)
        app.buttons["Giao dịch"].tap()
        XCTAssertTrue(app.staticTexts["Giao dịch"].firstMatch.waitForExistence(timeout: 5))
        app.buttons["bottom-tab-account"].tap()
        XCTAssertTrue(app.staticTexts["Tài khoản"].firstMatch.waitForExistence(timeout: 5))
    }

    func testFeedbackBubbleHidesBeforeFeedbackOverlayCapture() throws {
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated"]
        app.launch()

        let feedbackButtons = app.buttons.matching(
            NSPredicate(format: "label == %@", "Gửi phản hồi")
        )
        XCTAssertEqual(feedbackButtons.count, 1)
        feedbackButtons.firstMatch.tap()
        XCTAssertTrue(app.staticTexts["Phản hồi"].firstMatch.waitForExistence(timeout: 5))
        XCTAssertFalse(app.buttons["floating-feedback-bubble"].exists)
        XCTAssertTrue(app.buttons["feedback-attachment-add"].exists, "User can attach or replace an image, not only accept an automatic capture")
        XCTAssertEqual(
            feedbackButtons.count,
            1,
            "The floating trigger must not be captured or remain over the feedback sheet."
        )
    }

    func testAccountFeedbackBubbleIsTappableAtItsLowerEdge() throws {
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated"]
        app.launch()
        openMore(app)
        let feedback = app.buttons["floating-feedback-bubble"]
        XCTAssertTrue(feedback.waitForExistence(timeout: 5))
        feedback.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: 0.9)).tap()
        XCTAssertTrue(app.staticTexts["Phản hồi"].firstMatch.waitForExistence(timeout: 5), "The selected Account tab blocker must not intercept the feedback hit area")
    }

    func testAPIBackedFaceIDEnrollmentSurvivesLogout() throws {
#if !LOCAL_API_UI
        throw XCTSkip("Requires disposable LOCAL_API_UI and an enrolled simulator biometric")
#else
        guard ["localhost", "127.0.0.1"].contains(LocalAPIFixture.baseURL.host ?? "") else {
            throw XCTSkip("Biometric enrollment test must only use the disposable loopback API")
        }
        guard LAContext().canEvaluatePolicy(.deviceOwnerAuthenticationWithBiometrics, error: nil) else {
            throw XCTSkip("Enable simulator Face ID enrollment before this hardware-specific test")
        }
        let app = XCUIApplication()
        app.launchArguments = ["-UITestUnauthenticated"]
        app.launch()
        let email = app.textFields["auth-email"]
        XCTAssertTrue(email.waitForExistence(timeout: 10))
        email.clearAndType("native-fixture@example.test")
        XCTAssertEqual(email.value as? String, "native-fixture@example.test", "Email input must preserve @ and every typed character")
        app.secureTextFields["auth-password"].clearAndType("synthetic-fixture-password")
        dismissKeyboard(app)
        app.buttons["Đăng nhập"].tap()
        XCTAssertTrue(app.buttons["bottom-tab-account"].waitForExistence(timeout: 10), app.debugDescription)
        openMore(app)
        app.buttons["more-profile-account"].tap()
        let toggle = app.buttons["account-face-id-toggle"]
        reveal(toggle, in: app)
        XCTAssertTrue(toggle.waitForExistence(timeout: 10))
        if toggle.label.contains("Tắt") {
            toggle.tap()
            XCTAssertTrue(app.buttons["Bật đăng nhập Face ID"].waitForExistence(timeout: 10))
        }
        toggle.tap()
        let confirmation = app.secureTextFields["face-id-enrollment-password"]
        XCTAssertTrue(confirmation.waitForExistence(timeout: 5))
        confirmation.clearAndType("synthetic-fixture-password")
        dismissKeyboard(app)
        app.buttons["face-id-enrollment-confirm"].tap()
        XCTAssertTrue(app.buttons["account-forget-device"].waitForExistence(timeout: 10), app.debugDescription)
        attachScreenshot(app, named: "face-id-enrolled-account")
        let logout = app.buttons["account-logout"]
        XCTAssertTrue(XCTWaiter.wait(for: [XCTNSPredicateExpectation(predicate: NSPredicate(format: "exists == false"), object: confirmation)], timeout: 5) == .completed)
        reveal(logout, in: app)
        XCTAssertTrue(logout.isHittable)
        XCTAssertLessThanOrEqual(logout.frame.maxY, app.buttons["bottom-tab-account"].frame.minY, "Logout must be visibly above navigation, not merely present in the accessibility tree")
        logout.tap()
        XCTAssertTrue(app.buttons["auth-face-id-login"].waitForExistence(timeout: 10), "Logging out must retain the independently protected device sign-in credential")
        XCTAssertEqual(app.textFields["auth-email"].value as? String, "native-fixture@example.test")
        attachScreenshot(app, named: "face-id-after-logout")
        app.buttons["auth-face-id-login"].tap()
        XCTAssertTrue(app.buttons["bottom-tab-overview"].waitForExistence(timeout: 10), "Protected Keychain credential must obtain a new authenticated session")
        attachScreenshot(app, named: "face-id-fresh-session")
#endif
    }

    func testPrivacyMaskHidesOverviewBalanceFromAccessibility() throws {
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated"]
        app.launch()
        let hideBalances = app.buttons["icon-Ẩn số dư"]
        if hideBalances.waitForExistence(timeout: 5) {
            hideBalances.tap()
        } else {
            XCTAssertTrue(app.buttons["icon-Hiện số dư"].exists, "The test account may already have privacy masking enabled from an earlier launch.")
        }
        XCTAssertTrue(app.staticTexts["Đã ẩn tổng số dư"].waitForExistence(timeout: 5))
    }

    func testTransactionOverflowMenuKeepsTransferAndAdjustmentActions() throws {
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated"]
        app.launch()
        XCTAssertTrue(app.buttons["Giao dịch"].waitForExistence(timeout: 5))
        app.buttons["Giao dịch"].tap()
        let menu = app.buttons["Tùy chọn giao dịch"]
        XCTAssertTrue(menu.waitForExistence(timeout: 5))
        menu.tap()
        XCTAssertTrue(app.buttons["Chuyển tiền đến ví khác"].waitForExistence(timeout: 5))
        XCTAssertTrue(app.buttons["Điều chỉnh số dư"].exists)
    }

    func testLedgerHeaderUsesWalletScopeWithoutDuplicateTitle() throws {
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated"]
        app.launch()
        app.buttons["bottom-tab-transactions"].tap()
        XCTAssertTrue(app.buttons["transaction-wallet-selector"].waitForExistence(timeout: 5))
        XCTAssertFalse(app.staticTexts["screen-title-Giao dịch"].exists, "Wallet selection replaces the redundant ledger title")
        XCTAssertTrue(app.buttons["Tùy chọn giao dịch"].exists)
    }

    func testGlobalAddLongPressOpensAIEntry() throws {
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated"]
        app.launch()
        let add = app.buttons["add-transaction-button"]
        XCTAssertTrue(add.waitForExistence(timeout: 5))
        add.press(forDuration: 0.7)
        XCTAssertTrue(app.staticTexts["screen-title-Nhập giao dịch bằng AI"].waitForExistence(timeout: 5))
        app.buttons["screen-back"].tap()
        XCTAssertTrue(app.staticTexts["Mở rộng"].waitForExistence(timeout: 5))
        XCTAssertTrue(
            add.exists,
            "Account root keeps the shared transaction Add action; only Account child routes hide it."
        )
        app.buttons["bottom-tab-transactions"].tap()
        let transactionAdd = app.buttons["add-transaction-button"]
        XCTAssertTrue(transactionAdd.isEnabled)
        transactionAdd.tap()
        XCTAssertTrue(app.staticTexts["screen-title-Thêm giao dịch"].waitForExistence(timeout: 5))
    }

    func testTransactionNoteHasDedicatedEditorAndCancelDoesNotChangeDraft() throws {
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated"]
        app.launch()
        app.buttons["add-transaction-button"].tap()
        let noteRow = app.buttons["transaction-note-row"]
        XCTAssertTrue(noteRow.waitForExistence(timeout: 5))
        noteRow.tap()
        let saveNote = app.buttons["transaction-note-save"]
        guard saveNote.waitForExistence(timeout: 5) else {
            XCTFail("Tapping note must open its own editor, not expand report details.")
            return
        }
        let editor = app.textViews.firstMatch
        editor.tap()
        editor.typeText("Ca phe\nGhi chu rieng")
        saveNote.tap()
        XCTAssertTrue(noteRow.waitForExistence(timeout: 5))
        XCTAssertTrue(noteRow.label.contains("Ca phe"))
        noteRow.tap()
        editor.tap()
        editor.typeText(" discard")
        app.buttons["screen-close"].firstMatch.tap()
        XCTAssertTrue(noteRow.waitForExistence(timeout: 5))
        XCTAssertFalse(noteRow.label.contains("discard"))
        let details = app.buttons["transaction-details-toggle"]
        reveal(details, in: app)
        details.tap()
        XCTAssertTrue(app.buttons["transaction-include-reports-checkbox"].waitForExistence(timeout: 5))
        XCTAssertFalse(app.textFields["Ghi chú"].exists, "Report details must not contain a duplicate note editor.")
    }

    func testAddTransactionUsesStitchFormBaseControls() throws {
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated"]
        app.launch()

        let add = app.buttons["add-transaction-button"]
        XCTAssertTrue(add.waitForExistence(timeout: 5))
        add.tap()

        XCTAssertTrue(app.staticTexts["screen-title-Thêm giao dịch"].waitForExistence(timeout: 5))
        XCTAssertTrue(app.buttons["transaction-type-selector"].waitForExistence(timeout: 5))
        XCTAssertTrue(app.buttons["transaction-wallet-row"].waitForExistence(timeout: 5))
        XCTAssertTrue(app.staticTexts["Ví"].exists)
        XCTAssertTrue(app.buttons["transaction-category-row"].waitForExistence(timeout: 5))
        XCTAssertTrue(app.staticTexts["Danh mục"].exists)
        XCTAssertTrue(app.staticTexts["Số tiền"].exists)
        let amountCurrency = app.staticTexts["Số tiền-currency"]
        let amountInput = app.textFields["Số tiền"]
        XCTAssertTrue(amountCurrency.waitForExistence(timeout: 5))
        XCTAssertTrue(amountInput.waitForExistence(timeout: 5))
        XCTAssertLessThan(amountCurrency.frame.midX, amountInput.frame.minX, "Stitch transaction form places the VND chip before the large amount input")
        XCTAssertTrue(app.buttons["transaction-note-row"].waitForExistence(timeout: 5))
        XCTAssertTrue(app.buttons["transaction-date-stepper"].waitForExistence(timeout: 5))
        let calendar = app.buttons["transaction-date-stepper-calendar-button"]
        XCTAssertTrue(calendar.waitForExistence(timeout: 5))
        calendar.tap()
        XCTAssertTrue(app.staticTexts["screen-title-Chọn ngày"].waitForExistence(timeout: 5))
        XCTAssertTrue(app.descendants(matching: .any)["date-calendar-picker"].exists)
        attachScreenshot(app, named: "transaction-calendar-stitch-parity")
        app.buttons.matching(identifier: "screen-close").firstMatch.tap()
        XCTAssertTrue(app.buttons["transaction-date-stepper"].waitForExistence(timeout: 5))
        calendar.tap()
        XCTAssertTrue(app.buttons["date-calendar-done"].waitForExistence(timeout: 5))
        app.buttons["date-calendar-done"].tap()
        XCTAssertTrue(app.buttons["transaction-date-stepper"].waitForExistence(timeout: 5))
        XCTAssertTrue(app.buttons["transaction-details-toggle"].waitForExistence(timeout: 5))
        app.buttons["transaction-details-toggle"].tap()
        let reportCheckbox = app.buttons["transaction-include-reports-checkbox"]
        XCTAssertTrue(reportCheckbox.waitForExistence(timeout: 5))
        XCTAssertEqual(reportCheckbox.value as? String, "Đã chọn", "New transactions should be included in reports by default.")
        reportCheckbox.tap()
        XCTAssertEqual(reportCheckbox.value as? String, "Chưa chọn", "The shared checkbox should expose its toggled state to accessibility.")
        XCTAssertTrue(app.buttons["transaction-save-footer"].waitForExistence(timeout: 5))
        XCTAssertTrue(app.buttons["Lưu giao dịch"].waitForExistence(timeout: 5))
        XCTAssertTrue(app.buttons["transaction-ai-entry-button"].waitForExistence(timeout: 5))
        let calculator = app.buttons["amount-calculator-Số tiền"]
        XCTAssertTrue(calculator.waitForExistence(timeout: 5))
        calculator.tap()
        let keypadDisplay = app.descendants(matching: .any)["amount-keypad-display"].firstMatch
        XCTAssertTrue(keypadDisplay.waitForExistence(timeout: 5))
        for key in ["1", "0", "0", "0", "0", "0", "0", "+", "5", "0", "0", "0", "0", "0"] {
            XCTAssertTrue(app.buttons["amount-key-\(key)"].waitForExistence(timeout: 2), "Missing amount key \(key)")
            app.buttons["amount-key-\(key)"].tap()
        }
        app.buttons["amount-key-="].tap()
        XCTAssertTrue(keypadDisplay.label.contains("1500000"))
        app.buttons["amount-key-xong"].tap()
        XCTAssertTrue(app.buttons["transaction-save-footer"].waitForExistence(timeout: 5))
        XCTAssertEqual(amountInput.value as? String, "1.500.000", "The visible VND value uses the same Vietnamese grouping as web and Stitch")
        attachScreenshot(app, named: "transaction-editor-stitch-parity")
    }

    func testAuthenticatedMoreRoutesAndFeedbackEntry() throws {
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated"]
        app.launch()
        XCTAssertTrue(app.buttons["bottom-tab-account"].waitForExistence(timeout: 5))
        app.buttons["bottom-tab-account"].tap()
        let reports = app.buttons["more-destination-reports"]
        XCTAssertTrue(reports.waitForExistence(timeout: 5))
        reveal(reports, in: app)
        XCTAssertTrue(reports.isHittable, "Reports entry is not hittable: \(reports.frame)")
        reports.tap()
        XCTAssertTrue(app.staticTexts["screen-title-Báo cáo"].firstMatch.waitForExistence(timeout: 5))
        XCTAssertTrue(app.buttons["Gửi phản hồi"].waitForExistence(timeout: 5))
        app.buttons["Gửi phản hồi"].firstMatch.tap()
        XCTAssertTrue(app.staticTexts["Phản hồi"].firstMatch.waitForExistence(timeout: 5))
    }

    func testMoreProfileCardOpensAccountSettings() throws {
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated"]
        app.launch()
        XCTAssertTrue(app.buttons["bottom-tab-account"].waitForExistence(timeout: 5))
        app.buttons["bottom-tab-account"].tap()

        XCTAssertTrue(app.staticTexts["UI Test"].waitForExistence(timeout: 5), "More should render the centered profile identity.")
        XCTAssertTrue(app.staticTexts["Hồ sơ tài khoản"].waitForExistence(timeout: 5), "Profile card should render the account subtitle.")
        attachScreenshot(app, named: "more-account-stitch-parity")
        let manageAccount = app.buttons["more-profile-account"]
        XCTAssertTrue(manageAccount.waitForExistence(timeout: 5), "Profile card should expose account management.")
        manageAccount.tap()
        XCTAssertTrue(app.staticTexts["screen-title-Tài khoản"].waitForExistence(timeout: 5))
    }

    func testMoreProfileHelpOpensFeedback() throws {
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated"]
        app.launch()
        XCTAssertTrue(app.buttons["bottom-tab-account"].waitForExistence(timeout: 5))
        app.buttons["bottom-tab-account"].tap()

        XCTAssertTrue(app.buttons["more-profile-help"].waitForExistence(timeout: 5), "More header should expose the shared support pill.")
        let help = app.buttons["more-profile-help"]
        XCTAssertTrue(help.waitForExistence(timeout: 5), "Profile card should expose help.")
        help.tap()
        XCTAssertTrue(app.staticTexts["screen-title-Phản hồi"].waitForExistence(timeout: 5))
    }

    func testOverviewNotificationBellOpensNotifications() throws {
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated"]
        app.launch()

        let notificationButton = app.buttons["icon-Thông báo"]
        XCTAssertTrue(notificationButton.waitForExistence(timeout: 5), "Overview should expose the notification bell.")
        notificationButton.tap()
        XCTAssertTrue(app.staticTexts["screen-title-Thông báo"].waitForExistence(timeout: 5))
    }

    func testAPIBackedFeedbackBubbleSubmitsWhenScreenshotStorageUnavailable() async throws {
#if !LOCAL_API_UI
        throw XCTSkip("Build with SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI and run against the disposable local API fixture.")
#else
        let title = "UI feedback \(UUID().uuidString.prefix(8))"
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()

        let bubble = app.buttons["Gửi phản hồi"]
        XCTAssertTrue(bubble.waitForExistence(timeout: 10))
        bubble.tap()
        XCTAssertTrue(app.staticTexts["Phản hồi"].firstMatch.waitForExistence(timeout: 10))

        let titleField = app.textFields["Tiêu đề"]
        XCTAssertTrue(titleField.waitForExistence(timeout: 10))
        titleField.clearAndType(title)
        let detail = app.textViews["Chi tiết"]
        XCTAssertTrue(detail.waitForExistence(timeout: 10))
        detail.clearAndType("Ảnh chụp không có storage trong môi trường local; nội dung vẫn phải được lưu.")
        let submit = app.buttons["Gửi phản hồi"]
        XCTAssertTrue(submit.waitForExistence(timeout: 10))
        submit.tap()
        XCTAssertFalse(app.staticTexts["Phản hồi"].firstMatch.waitForExistence(timeout: 5), "Overlay should close after text-only fallback succeeds.")

        openMore(app)
        let feedbackRoute = app.buttons["Phản hồi & changelog"].firstMatch
        XCTAssertTrue(feedbackRoute.waitForExistence(timeout: 10))
        reveal(feedbackRoute, in: app)
        feedbackRoute.tap()
        XCTAssertTrue(app.staticTexts["screen-title-Phản hồi"].waitForExistence(timeout: 10))
        XCTAssertTrue(app.staticTexts[title].waitForExistence(timeout: 15))
#endif
    }

    func testAuthenticatedMoreMenuOpensEveryDestination() throws {
        let app = XCUIApplication()
#if LOCAL_API_UI
        app.launchArguments = ["-UITestAuthenticatedAPI"]
#else
        app.launchArguments = ["-UITestAuthenticated"]
#endif
        app.terminate()
        app.launch()

        let destinations: [(title: String, id: String)] = [
            ("Phản hồi", "feedback"),
            ("Ví", "wallets"),
            ("Sổ tín dụng", "credit"),
            ("Danh mục đầu tư", "portfolio"),
            ("Ngân sách", "budgets"),
            ("Hũ chi tiêu", "jars"),
            ("Tháng", "months"),
            ("Danh mục", "categories"),
            ("Giao dịch định kỳ", "recurring"),
            ("Travel Mode", "travel"),
            ("Báo cáo", "reports"),
            ("Nhập giao dịch bằng AI", "aiEntry")
        ]

        for destination in destinations {
            let title = destination.title
            XCTAssertTrue(app.buttons["bottom-tab-account"].waitForExistence(timeout: 5))
            app.buttons["bottom-tab-account"].tap()
            XCTAssertTrue(
                app.staticTexts["screen-title-Thêm"].waitForExistence(timeout: 5),
                "More menu did not render before selecting: \(title)"
            )
            let entry = moreEntry("more-destination-\(destination.id)", in: app)
            XCTAssertTrue(entry.waitForExistence(timeout: 5), "Missing More destination: \(title)")
            // The grouped account surface is intentionally scrollable. Use
            // the shared reveal helper so route coverage remains stable when
            // the visual grouping adds a section label or profile card.
            reveal(entry, in: app)
            XCTAssertTrue(entry.exists, "More destination disappeared: \(title)")
            entry.tap()
            let expectedScreenTitle = destination.id == "wallets" ? "Chọn Ví" : title
            let screenTitle = app.staticTexts["screen-title-\(expectedScreenTitle)"].firstMatch
            XCTAssertTrue(screenTitle.waitForExistence(timeout: 5), "Destination did not render: \(title)")
            if ["categories", "jars", "months", "portfolio", "notifications"].contains(destination.id) {
                let viewport = app.descendants(matching: .any)["shell-content-viewport"]
                XCTAssertTrue(
                    viewport.waitForExistence(timeout: 2) && viewport.frame.intersects(app.buttons["floating-feedback-bubble"].frame),
                    "Scrollable secondary screen extends behind the bubble overlay, not a reserved lane: \(title)"
                )
            }
            attachScreenshot(app, named: "secondary-route-\(destination.id)")

            let back = app.buttons["screen-back"].firstMatch
            guard back.waitForExistence(timeout: 5) else {
                XCTFail("Missing back action for: \(title)")
                return
            }
            back.tap()
            XCTAssertTrue(
                app.staticTexts["screen-title-Thêm"].waitForExistence(timeout: 5),
                "Could not return from: \(title)"
            )
        }
    }

    func testAPIBackedFeedbackResolveMovesTabsAndReopensPersistedItem() async throws {
#if !LOCAL_API_UI
        throw XCTSkip("Requires disposable local API fixture")
#else
        let title = "Resolution UI \(UUID().uuidString.prefix(8))"
        let item = try await LocalAPIFixture.createResolutionFeedback(title: title)
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        openMore(app)
        let route = app.buttons["more-destination-feedback"]
        XCTAssertTrue(route.waitForExistence(timeout: 10))
        reveal(route, in: app)
        route.tap()
        let checkbox = app.buttons["feedback-resolution-\(item.id)"]
        XCTAssertTrue(checkbox.waitForExistence(timeout: 10))
        checkbox.tap()
        XCTAssertTrue(checkbox.waitForNonExistence(timeout: 10), "Saved item leaves unresolved tab")
        let saved = try await LocalAPIFixture.resolutionFeedback(id: item.id)
        XCTAssertTrue(saved.resolved)
        XCTAssertEqual(saved.status, "open", "Owner confirmation does not alter dev lifecycle")
        app.buttons["segment-resolved"].tap()
        XCTAssertTrue(checkbox.waitForExistence(timeout: 10))
        checkbox.tap()
        XCTAssertTrue(checkbox.waitForNonExistence(timeout: 10))
        let reopened = try await LocalAPIFixture.resolutionFeedback(id: item.id)
        XCTAssertFalse(reopened.resolved)
        app.buttons["feedback-tabs"].tap()
        XCTAssertTrue(checkbox.waitForExistence(timeout: 10))
        let screenshot = XCTAttachment(screenshot: app.screenshot())
        screenshot.name = "Feedback unresolved after reopen"
        screenshot.lifetime = .keepAlways
        add(screenshot)
#endif
    }

    func testMoreDoesNotDuplicateProfileAndOverviewNotificationDestinations() {
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated"]
        app.launch()
        openMore(app)
        XCTAssertTrue(app.buttons["more-profile-account"].waitForExistence(timeout: 5))
        XCTAssertFalse(app.buttons["more-destination-account"].exists, "Profile is the sole account entry")
        XCTAssertFalse(app.buttons["more-destination-notifications"].exists, "Overview bell owns inbox entry")
    }

    func testAssistantTabIsSolePrimaryEntryAndRetainsUnsentDraftAcrossTabs() {
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated"]
        app.launch()
        openMore(app)
        for _ in 0..<4 { app.scrollViews.firstMatch.swipeUp() }
        XCTAssertFalse(app.buttons["more-destination-assistant"].exists, "Advisor tab must not compete with a duplicate More route.")
        XCTAssertTrue(app.buttons["more-destination-aiEntry"].exists, "Transaction extraction stays a separate workflow.")
        let tab = app.buttons["bottom-tab-assistant"]
        XCTAssertTrue(tab.isHittable)
        tab.tap()
        let input = app.textFields["assistant-composer-input"]
        XCTAssertTrue(input.waitForExistence(timeout: 5))
        input.tap()
        input.typeText("Chi tieu thang nay")
        XCTAssertFalse(tab.exists)
        app.buttons["assistant-dismiss-keyboard"].tap()
        XCTAssertTrue(tab.waitForExistence(timeout: 5))
        openMore(app)
        tab.tap()
        XCTAssertTrue(input.waitForExistence(timeout: 5))
        XCTAssertEqual(input.value as? String, "Chi tieu thang nay", "Switching tabs must preserve the unsent financial question.")
        attachScreenshot(app, named: "assistant-primary-tab-retained-draft")
    }

    func testAIEntryScreenBackActionHasStableIdentifier() throws {
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated"]
        app.launch()
        openMore(app)

        let entry = moreEntry("more-destination-aiEntry", in: app)
        XCTAssertTrue(entry.waitForExistence(timeout: 5))
        reveal(entry, in: app)
        entry.tap()
        XCTAssertTrue(app.staticTexts["screen-title-Nhập giao dịch bằng AI"].waitForExistence(timeout: 10))

        let back = app.buttons["screen-back"]
        XCTAssertTrue(back.waitForExistence(timeout: 5), "AI entry must expose the shared secondary-route back identifier.")
        XCTAssertTrue(back.isHittable)
        back.tap()
        XCTAssertTrue(app.staticTexts["screen-title-Thêm"].waitForExistence(timeout: 10))
    }

    func testScrollableCategoryRouteUsesBoundedShellViewport() throws {
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated"]
        app.launch()
        openMore(app)
        let entry = moreEntry("more-destination-categories", in: app)
        XCTAssertTrue(entry.waitForExistence(timeout: 5))
        reveal(entry, in: app)
        entry.tap()
        XCTAssertTrue(app.staticTexts["screen-title-Danh mục"].waitForExistence(timeout: 5))
        XCTAssertFalse(app.buttons["Back"].exists, "Secondary routes must not show the native back button above the Stitch header.")
        let viewport = app.descendants(matching: .any)["shell-content-viewport"]
        XCTAssertTrue(
            viewport.waitForExistence(timeout: 2) && viewport.frame.intersects(app.buttons["floating-feedback-bubble"].frame),
            "Categories use the full scroll viewport and bubble overlay."
        )
        attachScreenshot(app, named: "categories-shell-clearance")
    }

    func testScrollableNotificationsRouteUsesBoundedShellViewport() throws {
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated"]
        app.launch()
        let entry = app.buttons["icon-Thông báo"]
        XCTAssertTrue(entry.waitForExistence(timeout: 5))
        entry.tap()
        XCTAssertTrue(app.staticTexts["screen-title-Thông báo"].waitForExistence(timeout: 5))
        XCTAssertFalse(app.buttons["Back"].exists, "Secondary routes must not show the native back button above the Stitch header.")
        let viewport = app.descendants(matching: .any)["shell-content-viewport"]
        XCTAssertTrue(
            viewport.waitForExistence(timeout: 2) && viewport.frame.intersects(app.buttons["floating-feedback-bubble"].frame),
            "Notifications use the full scroll viewport and bubble overlay."
        )
        attachScreenshot(app, named: "notifications-shell-clearance")
    }

    func testAPIBackedNotificationCanBeMarkedRead() async throws {
#if !LOCAL_API_UI
        throw XCTSkip("Build with SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI and run against the disposable local API fixture.")
#else
        let title = "Native UI notification fixture"
        let notification = try await LocalAPIFixture.notification(named: title)
        XCTAssertNotNil(notification, "The local stage DB must be seeded with the documented synthetic notification.")
        guard let notification else { return }
        XCTAssertNil(notification.readAt, "Fixture should begin unread.")

        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        openMore(app)
        app.buttons["bottom-tab-overview"].tap()
        let entry = app.buttons["icon-Thông báo"]
        XCTAssertTrue(entry.waitForExistence(timeout: 10))
        reveal(entry, in: app)
        entry.tap()

        XCTAssertTrue(app.staticTexts[title].waitForExistence(timeout: 10))
        let markRead = app.buttons["notification-mark-read-\(notification.id)"]
        XCTAssertTrue(markRead.waitForExistence(timeout: 5), "Unread notification must expose its read action.")
        markRead.tap()
        XCTAssertFalse(markRead.waitForExistence(timeout: 3), "Read notification must no longer expose the unread action.")

        let updated = try await LocalAPIFixture.notification(named: title)
        XCTAssertNotNil(updated?.readAt, "The read action must persist to the owner-scoped API.")
#endif
    }

    func testRemainingScrollableSecondaryRoutesReserveShellInset() throws {
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated"]
        app.launch()

        let routes: [(id: String, title: String)] = [
            ("aiEntry", "Nhập giao dịch bằng AI"),
            ("feedback", "Phản hồi")
        ]

        for route in routes {
            openMore(app)
            let entry = moreEntry("more-destination-\(route.id)", in: app)
            XCTAssertTrue(entry.waitForExistence(timeout: 5), "Missing route: \(route.id)")
            reveal(entry, in: app)
            entry.tap()
            XCTAssertTrue(
                app.staticTexts["screen-title-\(route.title)"].waitForExistence(timeout: 5),
                "Route did not render: \(route.id)"
            )
            XCTAssertFalse(app.buttons["Back"].exists, "Secondary routes must not show the native back button: \(route.id)")
            let viewport = app.descendants(matching: .any)["shell-content-viewport"]
            XCTAssertTrue(
                viewport.waitForExistence(timeout: 2) && viewport.frame.intersects(app.buttons["floating-feedback-bubble"].frame),
                "Scrollable secondary route must use the bounded shell viewport: \(route.id)"
            )
            let back = app.buttons["screen-back"].firstMatch
            XCTAssertTrue(back.waitForExistence(timeout: 5), "Missing custom back action: \(route.id)")
            back.tap()
            XCTAssertTrue(app.staticTexts["screen-title-Thêm"].waitForExistence(timeout: 5))
        }
    }

    func testWalletEditorReservesShellInset() async throws {
#if !LOCAL_API_UI
        throw XCTSkip("Build with SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI and run against the disposable local API fixture.")
#else
        let wallet = try await LocalAPIFixture.wallet(
            name: "Fixture Wallet Editor (UUID().uuidString.prefix(8))",
            type: "basic"
        )
        addTeardownBlock { try? await LocalAPIFixture.delete(wallet) }
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        openMore(app)
        let walletsEntry = moreEntry("more-destination-wallets", in: app)
        XCTAssertTrue(walletsEntry.waitForExistence(timeout: 5))
        reveal(walletsEntry, in: app)
        walletsEntry.tap()
        XCTAssertTrue(app.staticTexts["screen-title-Chọn Ví"].waitForExistence(timeout: 5))
        let addWallet = app.buttons["wallet-add-button"]
        XCTAssertTrue(addWallet.waitForExistence(timeout: 5))
        addWallet.tap()
        XCTAssertTrue(app.staticTexts["Thêm ví"].waitForExistence(timeout: 5))
        XCTAssertTrue(
            app.descendants(matching: .any)["bottom-navigation-inset"].waitForExistence(timeout: 2),
            "Wallet editor must reserve the persistent shell navigation."
        )
        let save = app.buttons["wallet-editor-save"]
        XCTAssertTrue(save.waitForExistence(timeout: 5))
        let shellTop = app.buttons["bottom-tab-account"].frame.minY
        XCTAssertLessThanOrEqual(save.frame.maxY, shellTop, "Wallet editor save footer must stay above the shell navigation.")
#endif
    }

    func testTransferAndAdjustmentEditorsReserveShellInset() throws {
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated"]
        app.launch()
        app.buttons["bottom-tab-transactions"].tap()
        XCTAssertTrue(app.buttons["transaction-wallet-selector"].waitForExistence(timeout: 5))

        let overflow = app.buttons["Tùy chọn giao dịch"]
        XCTAssertTrue(overflow.waitForExistence(timeout: 5))
        overflow.tap()
        let transfer = app.buttons["Chuyển tiền đến ví khác"]
        XCTAssertTrue(transfer.waitForExistence(timeout: 5))
        transfer.tap()
        XCTAssertTrue(app.staticTexts["Chuyển tiền"].waitForExistence(timeout: 5))
        XCTAssertTrue(
            app.descendants(matching: .any)["bottom-navigation-inset"].waitForExistence(timeout: 2),
            "Transfer editor must reserve the persistent shell navigation."
        )
        let transferClose = app.buttons["Đóng"]
        XCTAssertTrue(transferClose.waitForExistence(timeout: 5))
        transferClose.tap()

        overflow.tap()
        let adjustment = app.buttons["Điều chỉnh số dư"]
        XCTAssertTrue(adjustment.waitForExistence(timeout: 5))
        adjustment.tap()
        XCTAssertTrue(app.staticTexts["Điều chỉnh số dư"].waitForExistence(timeout: 5))
        XCTAssertTrue(
            app.descendants(matching: .any)["bottom-navigation-inset"].waitForExistence(timeout: 2),
            "Adjustment editor must reserve the persistent shell navigation."
        )
    }

    func testAPIBackedAssistantLoadsOlderPersistedHistoryWithoutJumpingToLatest() async throws {
#if !LOCAL_API_UI
        throw XCTSkip("Requires the disposable local history fixture")
#else
        let conversationID = "native-history-e2e-20261015-e8080c47"
        let ownerID = try await LocalAPIFixture.ownerID()
        XCTAssertEqual(ownerID, "e4aedf4a-1aa7-4918-ba09-0287c95b154c", "Use only the isolated fixture owner")
        let latestBefore = try await LocalAPIFixture.advisorHistory(conversationID: conversationID)
        let oldestBefore = try await LocalAPIFixture.advisorHistory(conversationID: conversationID, beforeSequence: 2)
        XCTAssertEqual(latestBefore.map(\.seq), Array((2...51).reversed()), "Seed51messages before running; API must page50 newest")
        XCTAssertEqual(oldestBefore.map(\.seq), [1])
        guard latestBefore.count == 50, oldestBefore.count == 1 else { return }

        let app = XCUIApplication()
        // Foundation argument-domain preference is launch-scoped, not a
        // persisted override or a new production fixture hook.
        app.launchArguments = ["-UITestAuthenticatedAPI", "-mypocket.advisor.conversation.\(ownerID)", conversationID]
        app.launch()
        app.buttons["bottom-tab-assistant"].tap()
        let latest = app.staticTexts["Native History Fixture 51"]
        XCTAssertTrue(latest.waitForExistence(timeout: 15))
        let conversation = app.scrollViews["assistant-conversation"]
        XCTAssertTrue(conversation.exists)
        let loadOlder = app.buttons["Tải tin cũ"]
        for _ in 0..<30 {
            if loadOlder.exists && loadOlder.isHittable && conversation.frame.contains(loadOlder.frame) { break }
            conversation.swipeDown()
        }
        XCTAssertTrue(loadOlder.exists && loadOlder.isHittable)
        guard loadOlder.exists && loadOlder.isHittable else { attachScreenshot(app, named: "history-load-older-unreachable"); return }
        loadOlder.tap()
        let oldest = app.staticTexts["Native History Fixture 01"]
        XCTAssertTrue(oldest.waitForExistence(timeout: 10))
        XCTAssertTrue(loadOlder.waitForNonExistence(timeout: 5), "The final older page must stop pagination")
        XCTAssertEqual(app.staticTexts.matching(identifier: "Native History Fixture 01").count, 1)
        XCTAssertTrue(conversation.frame.intersects(oldest.frame), "Prepending must keep the older history visible")
        XCTAssertFalse(latest.exists && conversation.frame.intersects(latest.frame), "Loading older messages must not jump to latest")
        attachScreenshot(app, named: "assistant-persisted-older-history")

        // Offscreen rows can leave the accessibility snapshot. Check the
        // latest row's uniqueness only after bringing it back into view.
        for _ in 0..<30 {
            if latest.exists && latest.isHittable && conversation.frame.contains(latest.frame) { break }
            conversation.swipeUp()
        }
        XCTAssertTrue(latest.exists && latest.isHittable)
        XCTAssertEqual(app.staticTexts.matching(identifier: "Native History Fixture 51").count, 1)

        app.terminate()
        app.launch()
        app.buttons["bottom-tab-assistant"].tap()
        XCTAssertTrue(latest.waitForExistence(timeout: 15), "Persisted history must reload after restart")
        XCTAssertEqual(app.staticTexts.matching(identifier: "Native History Fixture 51").count, 1)
        let latestAfter = try await LocalAPIFixture.advisorHistory(conversationID: conversationID)
        let oldestAfter = try await LocalAPIFixture.advisorHistory(conversationID: conversationID, beforeSequence: 2)
        XCTAssertEqual(latestAfter, latestBefore, "History navigation must not mutate persisted messages")
        XCTAssertEqual(oldestAfter, oldestBefore)
#endif
    }

    func testAPIBackedAssistantLoadsOverview() throws {
#if !LOCAL_API_UI
        throw XCTSkip("Build with SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI and run against the disposable local API fixture.")
#else
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        let assistantEntry = app.buttons["bottom-tab-assistant"]
        XCTAssertTrue(assistantEntry.waitForExistence(timeout: 10))
        assistantEntry.tap()
        XCTAssertTrue(app.staticTexts["Finance Assistant"].firstMatch.waitForExistence(timeout: 10))
        let disabledMessage = app.staticTexts["Finance Assistant chưa được bật."]
        let firstQuestionPrompt = app.staticTexts["Hỏi về tài chính của bạn"]
        XCTAssertTrue(
            disabledMessage.waitForExistence(timeout: 10) || firstQuestionPrompt.waitForExistence(timeout: 10),
            "Assistant should render either its disabled state or the ready conversation empty state."
        )
        XCTAssertFalse(app.staticTexts["Không thể tải tổng quan trợ lý tài chính."].exists)
        let overview = app.buttons["assistant-pinned-overview"]
        XCTAssertTrue(overview.waitForExistence(timeout: 10))
        let top = overview.frame.minY
        let conversation = app.scrollViews["assistant-conversation"]
        XCTAssertTrue(conversation.exists)
        conversation.swipeUp()
        XCTAssertEqual(overview.frame.minY, top, accuracy: 1, "History scrolling must not move the financial overview.")
        overview.tap()
        XCTAssertTrue(app.staticTexts["Tổng quan tài chính"].firstMatch.waitForExistence(timeout: 5))
        app.buttons["Đóng"].firstMatch.tap()
        XCTAssertTrue(overview.waitForExistence(timeout: 5))
        let screenshot = XCTAttachment(screenshot: app.screenshot())
        screenshot.name = "assistant-api-pinned-overview"
        screenshot.lifetime = .keepAlways
        add(screenshot)
#endif
    }

    func testAPIBackedOverviewFinalSectionStaysAboveShellNavigation() throws {
#if !LOCAL_API_UI
        throw XCTSkip("Build with SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI and run against the disposable local API fixture.")
#else
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()

        let scrollView = app.scrollViews.firstMatch
        XCTAssertTrue(scrollView.waitForExistence(timeout: 10))
        let recent = app.descendants(matching: .any)["overview-section-recent-transactions"].firstMatch
        XCTAssertTrue(recent.waitForExistence(timeout: 10))

        reveal(recent, in: app)
        XCTAssertTrue(recent.isHittable, "Recent transactions section must be reachable above the shell")
        let shellTop = app.buttons["bottom-tab-account"].frame.minY
        XCTAssertLessThanOrEqual(
            recent.frame.maxY,
            shellTop,
            "Recent transactions section is obscured by persistent shell navigation"
        )
#endif
    }

    func testAssistantComposerIsCompactAndReachableWithKeyboard() throws {
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated"]
        app.launch()
        app.buttons["bottom-tab-assistant"].tap()
        XCTAssertTrue(app.staticTexts["Finance Assistant"].firstMatch.waitForExistence(timeout: 10))
        let input = app.textViews.firstMatch.exists ? app.textViews.firstMatch : app.textFields.firstMatch
        XCTAssertTrue(input.waitForExistence(timeout: 5))
        XCTAssertLessThanOrEqual(input.frame.height, 76, "A chat input must not consume a note editor's height at rest.")
        input.tap()
        XCTAssertTrue(app.keyboards.firstMatch.waitForExistence(timeout: 5))
        let send = app.buttons.matching(NSPredicate(format: "label == %@", "Gửi")).firstMatch
        XCTAssertTrue(send.exists)
        XCTAssertGreaterThanOrEqual(send.frame.height, 44)
        XCTAssertLessThanOrEqual(send.frame.maxY, app.keyboards.firstMatch.frame.minY + 1, "Keyboard must not cover send.")
        let shot = XCTAttachment(screenshot: app.screenshot())
        shot.name = "assistant-compact-keyboard"
        shot.lifetime = .keepAlways
        add(shot)
    }

    func testAssistantLargestTextKeepsMultilineComposerAndConversationAboveKeyboard() {
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated"]
        app.launch()
        app.buttons["bottom-tab-assistant"].tap()
        let input = app.textFields["assistant-composer-input"]
        XCTAssertTrue(input.waitForExistence(timeout: 10))
        let baseline = input.frame.height
        app.terminate()
        app.launchArguments += ["-UIPreferredContentSizeCategoryName", "UICTContentSizeCategoryAccessibilityXXXL"]
        app.launch()
        let navigation = app.buttons["bottom-navigation-menu"]
        XCTAssertTrue(navigation.waitForExistence(timeout: 5))
        navigation.tap()
        app.buttons["Trợ lý"].firstMatch.tap()
        XCTAssertTrue(input.waitForExistence(timeout: 10))
        XCTAssertGreaterThan(input.frame.height, baseline + 10, "The test must exercise actual system text scaling, not the normal layout")
        input.tap()
        XCTAssertTrue(app.keyboards.firstMatch.waitForExistence(timeout: 5))
        input.typeText("Dong dau\nDong hai\nDong ba\nDong cuoi can giu")
        XCTAssertTrue((input.value as? String)?.contains("Dong cuoi can giu") == true)
        let send = app.buttons["assistant-composer-action"]
        let composer = app.descendants(matching: .any)["assistant-composer"].firstMatch
        let conversation = app.scrollViews["assistant-conversation"]
        XCTAssertTrue(send.exists)
        XCTAssertGreaterThanOrEqual(send.frame.width, 44)
        XCTAssertGreaterThanOrEqual(send.frame.height, 44)
        XCTAssertTrue(app.frame.contains(composer.frame), "The complete focused composer must remain inside the app viewport")
        XCTAssertLessThanOrEqual(composer.frame.maxY, app.keyboards.firstMatch.frame.minY + 1, "Keyboard must not obscure large-text composer")
        XCTAssertGreaterThanOrEqual(conversation.frame.height, 44, "Large drafts must not consume the entire conversation viewport")
        attachScreenshot(app, named: "assistant-largest-text-multiline-keyboard")
        app.buttons["assistant-dismiss-keyboard"].tap()
        XCTAssertTrue(navigation.waitForExistence(timeout: 5))
        XCTAssertTrue((input.value as? String)?.contains("Dong cuoi can giu") == true)
        attachScreenshot(app, named: "assistant-largest-text-navigation-restored")
        navigation.tap()
        app.buttons["Tài khoản"].firstMatch.tap()
        XCTAssertTrue(navigation.waitForExistence(timeout: 5))
        XCTAssertTrue(navigation.label.contains("Tài khoản"))
        navigation.tap()
        app.buttons["Trợ lý"].firstMatch.tap()
        XCTAssertTrue(input.waitForExistence(timeout: 5))
        XCTAssertTrue((input.value as? String)?.contains("Dong cuoi can giu") == true, "Accessibility menu switching must retain the full draft")
    }

    func testAppRemainsPortraitWhenDeviceRotates() {
        let device = XCUIDevice.shared
        defer { device.orientation = .portrait }
        device.orientation = .portrait
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated"]
        app.launch()
        XCTAssertTrue(app.buttons["bottom-tab-overview"].waitForExistence(timeout: 10))
        device.orientation = .landscapeLeft
        let portrait = NSPredicate { _, _ in app.frame.height >= app.frame.width }
        let remainsPortrait = XCTNSPredicateExpectation(predicate: portrait, object: nil)
        XCTAssertEqual(XCTWaiter.wait(for: [remainsPortrait], timeout: 5), .completed)
        XCTAssertGreaterThanOrEqual(app.frame.height, app.frame.width, "MyPocket supports portrait only")
        attachScreenshot(app, named: "portrait-lock-after-landscape-request")
    }

    func testAssistantShowsSuggestedQuestionsBeforeFirstMessage() throws {
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated"]
        app.launch()
        let assistantEntry = app.buttons["bottom-tab-assistant"]
        XCTAssertTrue(assistantEntry.waitForExistence(timeout: 10))
        assistantEntry.tap()
        XCTAssertTrue(app.staticTexts["Finance Assistant"].firstMatch.waitForExistence(timeout: 10))
        XCTAssertTrue(
            app.buttons["assistant-suggestion-0"].waitForExistence(timeout: 10),
            "The assistant should offer a deterministic first question while history is empty."
        )
        app.buttons["assistant-suggestion-0"].tap()
        let composer = app.textFields["assistant-composer-input"]
        XCTAssertTrue(composer.exists)
        XCTAssertEqual(composer.value as? String, "Tháng này tôi tiêu thế nào?")
    }

    func testAssistantFeedbackBubbleDoesNotCoverComposerAndDraftSurvivesDisabledSend() throws {
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated"]
        app.launch()
        app.buttons["bottom-tab-assistant"].tap()
        let input = app.textFields["assistant-composer-input"]
        XCTAssertTrue(input.waitForExistence(timeout: 10))
        let send = app.buttons["assistant-composer-action"]
        let bubble = app.buttons["floating-feedback-bubble"]
        XCTAssertFalse(bubble.frame.intersects(send.frame), "Feedback must float above chat input, not cover send.")
        input.tap()
        input.typeText("Chi tieu thang nay\nDong hai\nDong ba\nDong bon\nDong nam\nDong sau")
        XCTAssertTrue(app.keyboards.firstMatch.exists)
        XCTAssertLessThan(input.frame.height, 140, "Long drafts should scroll inside a four-line input, not grow indefinitely.")
        XCTAssertFalse(send.isEnabled, "Unavailable advisor must not send, even when a local draft exists.")
        XCTAssertTrue((input.value as? String)?.contains("Dong sau") == true)
        XCTAssertFalse(bubble.exists, "Feedback stays hidden while typing and restores after keyboard dismissal.")
        let shot = XCTAttachment(screenshot: app.screenshot())
        shot.name = "assistant-multiline-draft"
        shot.lifetime = .keepAlways
        add(shot)
    }

    func testReportsExposeMonthFilterAndReloadAction() throws {
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated"]
        app.launch()
        openMore(app)
        let reportsEntry = app.buttons["more-destination-reports"]
        XCTAssertTrue(reportsEntry.waitForExistence(timeout: 10))
        reveal(reportsEntry, in: app)
        reportsEntry.tap()
        XCTAssertTrue(app.staticTexts["screen-title-Báo cáo"].waitForExistence(timeout: 10))
        XCTAssertTrue(app.textFields["reports-month"].waitForExistence(timeout: 10))
        let reload = app.buttons["reports-load-button"]
        XCTAssertTrue(reload.waitForExistence(timeout: 10))
        reload.tap()
        XCTAssertTrue(app.staticTexts["screen-title-Báo cáo"].exists)
    }

#if LOCAL_API_UI
    func testReportsRejectInvalidMonthBeforeNetworkAndShowRetryState() throws {
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated"]
        app.launch()
        openMore(app)
        let reportsEntry = app.buttons["more-destination-reports"]
        XCTAssertTrue(reportsEntry.waitForExistence(timeout: 10))
        reveal(reportsEntry, in: app)
        reportsEntry.tap()
        XCTAssertTrue(app.staticTexts["screen-title-Báo cáo"].waitForExistence(timeout: 10))

        let monthField = app.textFields["reports-month"]
        XCTAssertTrue(monthField.waitForExistence(timeout: 10))
        reveal(monthField, in: app)
        monthField.clearAndType("2026-99")
        dismissKeyboard(app)
        let reload = app.buttons["reports-load-button"]
        reveal(reload, in: app)
        reload.tap()

        XCTAssertTrue(
            app.staticTexts["Nhập tháng theo định dạng YYYY-MM (tháng từ 01 đến 12)."].waitForExistence(timeout: 5),
            "Invalid month must be rejected locally and surfaced through the shared error state."
        )
        XCTAssertFalse(app.descendants(matching: .any)["reports-section-summary"].exists)
    }
#endif

    func testAPIBackedReportsExposeMoneyInsiderDetails() async throws {
#if !LOCAL_API_UI
        throw XCTSkip("Build with SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI and run against the disposable local API fixture.")
#else
        let suffix = String(UUID().uuidString.prefix(8))
        let wallet = try await LocalAPIFixture.wallet(name: "Fixture Reports Wallet \(suffix)", type: "basic")
        addTeardownBlock {
            try? await LocalAPIFixture.deleteTransactions(walletID: wallet.wallet.id)
            try? await LocalAPIFixture.delete(wallet)
        }
        let expenseCategory = try await LocalAPIFixture.category(kind: "expense", walletID: wallet.wallet.id)
        let currentExpense = try await LocalAPIFixture.transaction(
            walletID: wallet.wallet.id,
            categoryID: expenseCategory.id,
            amount: 120_000,
            note: "Fixture current report expense",
            type: "expense",
            occurredAt: "2031-10-06T03:00:00Z"
        )
        let previousExpense = try await LocalAPIFixture.transaction(
            walletID: wallet.wallet.id,
            categoryID: expenseCategory.id,
            amount: 80_000,
            note: "Fixture previous report expense",
            type: "expense",
            occurredAt: "2031-09-06T03:00:00Z"
        )
        XCTAssertNotEqual(currentExpense.id, previousExpense.id)

        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        openMore(app)
        let reportsEntry = app.buttons["more-destination-reports"]
        XCTAssertTrue(reportsEntry.waitForExistence(timeout: 10))
        reveal(reportsEntry, in: app)
        reportsEntry.tap()
        XCTAssertTrue(app.staticTexts["screen-title-Báo cáo"].waitForExistence(timeout: 10))
        XCTAssertTrue(
            app.descendants(matching: .any)["reports-section-insider"].waitForExistence(timeout: 15),
            "Reports must render the full Money Insider card from the real insider response."
        )
        let monthField = app.textFields["reports-month"]
        XCTAssertTrue(monthField.waitForExistence(timeout: 10))
        reveal(monthField, in: app)
        monthField.clearAndType("2031-10")
        dismissKeyboard(app)
        let reload = app.buttons["reports-load-button"]
        reveal(reload, in: app)
        reload.tap()
        let comparisonRing = app.descendants(matching: .any)["money-insider-comparison-ring"].firstMatch
        XCTAssertTrue(
            comparisonRing.waitForExistence(timeout: 10),
            "A selected report month with a previous-month baseline must render its Money Insider comparison ring. UI: \(app.debugDescription)"
        )
        XCTAssertTrue(
            app.descendants(matching: .any)["reports-comparison-chart"].waitForExistence(timeout: 5),
            "Reports must show the income/expense comparison chart in the summary."
        )
        let trendChart = app.descendants(matching: .any).matching(
            NSPredicate(format: "label == %@", "Xu hướng thu chi")
        ).firstMatch
        XCTAssertTrue(
            trendChart.waitForExistence(timeout: 5),
            "The report trend must be mounted without switching modes."
        )
        XCTAssertFalse(app.buttons["Xu hướng"].exists, "Reports must not hide the trend behind a mode toggle.")
        attachScreenshot(app, named: "reports-summary-with-always-visible-trend")

        reveal(comparisonRing, in: app)
        let shellViewport = app.descendants(matching: .any)["shell-content-viewport"].firstMatch
        for _ in 0..<3 where comparisonRing.frame.maxY > shellViewport.frame.maxY {
            shellViewport.swipeUp()
        }
        XCTAssertLessThanOrEqual(
            comparisonRing.frame.maxY,
            shellViewport.frame.maxY,
            "The full comparison ring must be reachable inside the app's clipped content viewport."
        )
        XCTAssertTrue(comparisonRing.isHittable, "The comparison ring must not remain obscured by the persistent shell.")
        attachScreenshot(app, named: "money-insider-comparison-ring")
#endif
    }

    func testAPIBackedReportsTrendExposesPersistedIncomeAndExpenseValues() async throws {
#if !LOCAL_API_UI
        throw XCTSkip("Build with SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI and run against the disposable local API fixture.")
#else
        let suffix = String(UUID().uuidString.prefix(8))
        let walletName = "Fixture Trend Wallet \(suffix)"
        let wallet = try await LocalAPIFixture.wallet(name: walletName, type: "basic")
        addTeardownBlock {
            try? await LocalAPIFixture.deleteTransactions(walletID: wallet.wallet.id)
            try? await LocalAPIFixture.delete(wallet)
        }
        let incomeCategory = try await LocalAPIFixture.category(kind: "income", walletID: wallet.wallet.id)
        let expenseCategory = try await LocalAPIFixture.category(kind: "expense", walletID: wallet.wallet.id)
        let transactionDate = "2031-10-06T03:00:00Z" // 10:00 on Oct 6 in the fixture account timezone.
        let income = try await LocalAPIFixture.transaction(
            walletID: wallet.wallet.id,
            categoryID: incomeCategory.id,
            amount: 123_456,
            note: "Fixture report trend income",
            type: "income",
            occurredAt: transactionDate
        )
        let expense = try await LocalAPIFixture.transaction(
            walletID: wallet.wallet.id,
            categoryID: expenseCategory.id,
            amount: 34_567,
            note: "Fixture report trend expense",
            occurredAt: transactionDate
        )
        let excludedExpense = try await LocalAPIFixture.transaction(
            walletID: wallet.wallet.id,
            categoryID: expenseCategory.id,
            amount: 888_888,
            note: "Fixture excluded report trend expense",
            occurredAt: transactionDate,
            includedInReports: false
        )
        let persistedRows = try await LocalAPIFixture.transactions()
        XCTAssertTrue(persistedRows.contains(where: { $0.id == income.id && $0.includedInReports == true }))
        XCTAssertTrue(persistedRows.contains(where: { $0.id == expense.id && $0.includedInReports == true }))
        XCTAssertTrue(
            persistedRows.contains(where: { $0.id == excludedExpense.id && $0.includedInReports == false }),
            "The transaction API must return the persisted report-exclusion choice."
        )

        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        openMore(app)
        let reportsEntry = app.buttons["more-destination-reports"]
        XCTAssertTrue(reportsEntry.waitForExistence(timeout: 10))
        reveal(reportsEntry, in: app)
        reportsEntry.tap()
        XCTAssertTrue(app.staticTexts["screen-title-Báo cáo"].waitForExistence(timeout: 10))

        let monthField = app.textFields["reports-month"]
        XCTAssertTrue(monthField.waitForExistence(timeout: 10))
        reveal(monthField, in: app)
        monthField.clearAndType("2031-10")
        dismissKeyboard(app)
        let reload = app.buttons["reports-load-button"]
        reveal(reload, in: app)
        reload.tap()
        XCTAssertTrue(app.staticTexts["Tổng quan báo cáo"].waitForExistence(timeout: 15))

        let chart = app.descendants(matching: .any).matching(
            NSPredicate(format: "label == %@", "Xu hướng thu chi")
        ).firstMatch
        XCTAssertTrue(chart.waitForExistence(timeout: 10), app.debugDescription)
        XCTAssertEqual(
            chart.value as? String,
            "Ngày 06: Thu 123.456 ₫, Chi 34.567 ₫",
            "The populated trend chart should expose included amounts for the account-local day, excluding the 888,888 VND row."
        )
#endif
    }

    func testAccountAPIKeyEditorUsesStableBaseControls() throws {
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated"]
        app.launch()
        openMore(app)
        let accountEntry = app.buttons["more-profile-account"]
        XCTAssertTrue(accountEntry.waitForExistence(timeout: 10))
        reveal(accountEntry, in: app)
        accountEntry.tap()
        XCTAssertTrue(app.staticTexts["screen-title-Tài khoản"].waitForExistence(timeout: 10))
        let addKey = app.buttons["account-api-key-add"]
        XCTAssertTrue(addKey.waitForExistence(timeout: 10))
        addKey.tap()
        XCTAssertTrue(app.textFields["Tên key"].waitForExistence(timeout: 10))
        XCTAssertTrue(app.buttons["account-api-key-create"].waitForExistence(timeout: 10))
        XCTAssertFalse(app.staticTexts["Scope"].exists, "Keys must not offer obsolete limited permissions")
        XCTAssertTrue(app.staticTexts["account-api-key-full-access"].exists, "Full read/write authority must be visible before creation")
        attachScreenshot(app, named: "account-api-key-full-access")
    }

    func testFeedbackBubbleDoesNotCoverAccountSecurityDescription() throws {
        let app = XCUIApplication()
#if LOCAL_API_UI
        app.launchArguments = ["-UITestAuthenticatedAPI"]
#else
        app.launchArguments = ["-UITestAuthenticated"]
#endif
        app.launch()
        openMore(app)
        let accountEntry = app.buttons["more-profile-account"]
        XCTAssertTrue(accountEntry.waitForExistence(timeout: 10))
        reveal(accountEntry, in: app)
        accountEntry.tap()

        let explanation = app.staticTexts["Đăng nhập lại bằng Face ID, kể cả sau khi đăng xuất. Chỉ dùng trên iPhone này."]
        XCTAssertTrue(explanation.waitForExistence(timeout: 10))
        let feedback = app.buttons["floating-feedback-bubble"]
        XCTAssertTrue(feedback.waitForExistence(timeout: 5))
        XCTAssertFalse(
            explanation.frame.intersects(feedback.frame),
            "Floating feedback must not cover the Face ID explanation."
        )
        attachScreenshot(app, named: "account-feedback-collision")
    }

    func testFeedbackBubbleDoesNotCoverMoreRowsOrCategoryEditorActions() throws {
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated"]
        app.launch()
        openMore(app)

        let feedback = app.buttons["floating-feedback-bubble"]
        XCTAssertTrue(feedback.waitForExistence(timeout: 5))
        let viewport = app.descendants(matching: .any)["shell-content-viewport"]
        XCTAssertTrue(viewport.exists, "Authenticated route content must have a bounded shell viewport.")
        XCTAssertTrue(viewport.frame.intersects(feedback.frame), "Feedback overlays, rather than reduces, the viewport")
        let visibleMoreRows = app.buttons.matching(
            NSPredicate(format: "identifier BEGINSWITH %@", "more-destination-")
        ).allElementsBoundByIndex.filter { $0.exists && viewport.frame.intersects($0.frame) }
        XCTAssertFalse(visibleMoreRows.isEmpty, app.debugDescription)
        XCTAssertNil(
            visibleMoreRows.first { feedback.frame.contains(CGPoint(x: $0.frame.midX, y: $0.frame.midY)) },
            "Feedback must leave row action centers reachable."
        )
        attachScreenshot(app, named: "more-feedback-dock-before-scroll")

        app.swipeUp()
        let scrolledMoreRows = app.buttons.matching(
            NSPredicate(format: "identifier BEGINSWITH %@", "more-destination-")
        ).allElementsBoundByIndex.filter { $0.exists && viewport.frame.intersects($0.frame) }
        XCTAssertFalse(scrolledMoreRows.isEmpty, app.debugDescription)
        XCTAssertNil(
            scrolledMoreRows.first { feedback.frame.contains(CGPoint(x: $0.frame.midX, y: $0.frame.midY)) },
            "Scrolled More row centers must stay reachable."
        )

        let categories = moreEntry("more-destination-categories", in: app)
        XCTAssertTrue(categories.waitForExistence(timeout: 5))
        reveal(categories, in: app)
        categories.tap()
        let addCategory = app.buttons["category-add-group"]
        XCTAssertTrue(addCategory.waitForExistence(timeout: 10))
        reveal(addCategory, in: app)
        addCategory.tap()
        XCTAssertTrue(app.descendants(matching: .any)["category-editor-screen"].waitForExistence(timeout: 5))

        let save = app.buttons["Lưu nhóm"].firstMatch
        let cancel = app.buttons["Hủy"].firstMatch
        XCTAssertTrue(save.waitForExistence(timeout: 5))
        XCTAssertTrue(cancel.waitForExistence(timeout: 5))
        // On compact phones the applicability card pushes the editor actions
        // below the bounded shell viewport. Scroll them into the actually
        // visible viewport before comparing their visible frames with the dock;
        // an offscreen accessibility frame can extend behind the dock without
        // anything being painted or tappable there.
        for _ in 0..<20 where !viewport.frame.contains(save.frame) || !viewport.frame.contains(cancel.frame) {
            app.swipeUp()
        }
        XCTAssertTrue(viewport.frame.contains(save.frame), "Category save must be scrollable into the bounded viewport. frame=\(save.frame) viewport=\(viewport.frame)")
        XCTAssertTrue(viewport.frame.contains(cancel.frame), "Category cancel must be scrollable into the bounded viewport. frame=\(cancel.frame) viewport=\(viewport.frame)")
        XCTAssertFalse(feedback.frame.contains(CGPoint(x: save.frame.midX, y: save.frame.midY)), "Bubble must not cover the save action center")
        XCTAssertFalse(feedback.frame.contains(CGPoint(x: cancel.frame.midX, y: cancel.frame.midY)), "Bubble must not cover the cancel action center")
        cancel.tap()
        XCTAssertTrue(addCategory.waitForExistence(timeout: 5), "Category cancel must actually work above the navbar")
        attachScreenshot(app, named: "category-editor-feedback-clearance")
    }

    func testAccountRootAddOpensTransactionAndCancelReturnsToAccount() throws {
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated"]
        app.launch()
        openMore(app)
        let add = app.buttons["add-transaction-button"]
        XCTAssertTrue(add.waitForExistence(timeout: 3), "Account root keeps the global transaction Add action.")
        guard add.exists else { return }
        add.tap()
        let close = app.buttons["screen-close"]
        XCTAssertTrue(close.waitForExistence(timeout: 5))
        guard close.exists else { return }
        close.tap()
        XCTAssertTrue(app.staticTexts["screen-title-Thêm"].waitForExistence(timeout: 5), "Cancel returns to the originating Account tab.")
        app.buttons["more-profile-account"].tap()
        XCTAssertFalse(app.buttons["add-transaction-button"].exists, "Secondary routes keep their own create actions.")
    }

    func testSecondaryRouteSupportsLeadingEdgeSwipeBack() throws {
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated"]
        app.launch()
        openMore(app)
        app.buttons["more-profile-account"].tap()
        XCTAssertTrue(app.staticTexts["screen-title-Tài khoản"].waitForExistence(timeout: 5))
        let start = app.coordinate(withNormalizedOffset: CGVector(dx: 0.005, dy: 0.5))
        let end = app.coordinate(withNormalizedOffset: CGVector(dx: 0.85, dy: 0.5))
        start.press(forDuration: 0.1, thenDragTo: end)
        XCTAssertTrue(app.staticTexts["screen-title-Thêm"].waitForExistence(timeout: 5), "A leading-edge swipe pops the secondary route.")
    }

    func testSecondaryEdgeSwipeDoesNotDiscardCategoryEditorDraft() throws {
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated"]
        app.launch()
        openMore(app)
        let categories = app.buttons["more-destination-categories"]
        reveal(categories, in: app)
        categories.tap()
        let add = app.buttons["category-add-group"]
        XCTAssertTrue(add.waitForExistence(timeout: 5))
        add.tap()
        let name = app.textFields["Tên nhóm"]
        XCTAssertTrue(name.waitForExistence(timeout: 5))
        name.tap()
        name.typeText("Draft preserved")
        app.coordinate(withNormalizedOffset: CGVector(dx: 0.005, dy: 0.35))
            .press(forDuration: 0.1, thenDragTo: app.coordinate(withNormalizedOffset: CGVector(dx: 0.85, dy: 0.35)))
        XCTAssertTrue(app.staticTexts["category-editor-title"].exists)
        XCTAssertEqual(name.value as? String, "Draft preserved", "Back swipe cannot silently discard an editor draft.")
    }

    func testSelectedAccountTabDoesNotActivateLastMoreRow() throws {
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated"]
        app.launch()
        openMore(app)

        let accountTab = app.buttons["bottom-tab-account"]
        let lastMoreRow = app.buttons["more-destination-aiEntry"]
        XCTAssertTrue(lastMoreRow.waitForExistence(timeout: 5))
        reveal(lastMoreRow, in: app)
        accountTab.tap()

        XCTAssertTrue(
            app.staticTexts["screen-title-Thêm"].waitForExistence(timeout: 5),
            "Tapping the already-selected Account tab must not activate the last More row beneath it."
        )
        XCTAssertFalse(app.staticTexts["screen-title-Phản hồi"].exists)
    }

    func testMoreLastDestinationScrollsToTheDockBoundary() throws {
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated"]
        app.launch()
        openMore(app)

        let viewport = app.descendants(matching: .any)["shell-content-viewport"]
        let lastContentAction = app.buttons["Đăng xuất"]
        XCTAssertTrue(viewport.waitForExistence(timeout: 5))
        XCTAssertTrue(lastContentAction.waitForExistence(timeout: 5))
        for _ in 0..<12 { app.swipeUp() }

        let visibleLastAction = lastContentAction.frame.intersection(viewport.frame)
        XCTAssertFalse(visibleLastAction.isNull, "The final More action must remain reachable inside its viewport.")
        XCTAssertLessThanOrEqual(lastContentAction.frame.maxY, viewport.frame.maxY + 1, "The final More action must scroll into the bounded shell viewport.")
        let trailingClearance = viewport.frame.maxY - lastContentAction.frame.maxY
        XCTAssertLessThanOrEqual(
            trailingClearance,
            120,
            "The shell must not leave duplicate navigation/dock spacer after the final More row (gap: \(trailingClearance)pt)."
        )
    }

    func testFeedbackOverlayDoesNotReserveAScrollLane() throws {
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated"]
        app.launch()
        openMore(app)
        let viewport = app.descendants(matching: .any)["shell-content-viewport"]
        let tab = app.buttons["bottom-tab-account"]
        let bubble = app.buttons["floating-feedback-bubble"]
        XCTAssertTrue(viewport.waitForExistence(timeout: 5))
        XCTAssertTrue(bubble.exists)
        XCTAssertLessThanOrEqual(tab.frame.minY - viewport.frame.maxY, 12, "Content must extend to the navbar; allow its 6pt inset and the native label frame, not a 52pt feedback lane")
        XCTAssertLessThanOrEqual(viewport.frame.maxY, tab.frame.minY + 1, "Content bounds must not extend behind the navbar")
        XCTAssertTrue(viewport.frame.intersects(bubble.frame), "Feedback is an overlay within the content viewport")
        XCTAssertFalse(app.descendants(matching: .any)["feedback-dock"].exists)
        attachScreenshot(app, named: "compact-more-overlay-full-viewport")
    }

    func testAPIBackedAccountAPIKeyCreateAndRevoke() async throws {
#if !LOCAL_API_UI
        throw XCTSkip("Build with SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI and run against the disposable local API fixture.")
#else
        let suffix = String(UUID().uuidString.prefix(8))
        let keyName = "Fixture API Key \(suffix)"
        try? await LocalAPIFixture.deleteAPIKeys(named: keyName)
        addTeardownBlock { try? await LocalAPIFixture.deleteAPIKeys(named: keyName) }

        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        openMore(app)
        let accountEntry = app.buttons["more-profile-account"]
        XCTAssertTrue(accountEntry.waitForExistence(timeout: 10))
        reveal(accountEntry, in: app)
        accountEntry.tap()
        XCTAssertTrue(app.staticTexts["screen-title-Tài khoản"].waitForExistence(timeout: 10))

        let addKey = app.buttons["account-api-key-add"]
        XCTAssertTrue(addKey.waitForExistence(timeout: 10))
        reveal(addKey, in: app)
        addKey.tap()
        let name = app.textFields["Tên key"]
        XCTAssertTrue(name.waitForExistence(timeout: 10))
        reveal(name, in: app)
        name.clearAndType(keyName)
        dismissKeyboard(app)
        let create = app.buttons["account-api-key-create"]
        XCTAssertTrue(create.waitForExistence(timeout: 10))
        reveal(create, in: app)
        create.tap()

        let created = app.staticTexts[keyName]
        XCTAssertTrue(created.waitForExistence(timeout: 15), "API key did not render after POST: \(keyName)")
        guard created.exists else { return }

        let secret = app.staticTexts["account-api-key-secret-value"]
        XCTAssertTrue(secret.waitForExistence(timeout: 5), "One-time API key secret should be visible after creation.")
        let copy = app.buttons["account-api-key-copy"]
        XCTAssertTrue(copy.waitForExistence(timeout: 5), "The one-time secret needs an explicit copy action.")
        copy.tap()
        XCTAssertTrue(app.staticTexts["Đã sao chép API key."].waitForExistence(timeout: 5))

        let createdKey = try await LocalAPIFixture.apiKey(named: keyName)
        let revoke = app.buttons["account-api-key-revoke-\(createdKey.id)"]
        XCTAssertTrue(revoke.waitForExistence(timeout: 10))
        reveal(revoke, in: app)
        revoke.tap()
        XCTAssertFalse(app.staticTexts[keyName].waitForExistence(timeout: 5))
        XCTAssertFalse(app.staticTexts["account-api-key-secret-value"].exists, "Revoking the just-created key must clear its one-time secret")
        XCTAssertFalse(app.buttons["account-api-key-copy"].exists, "Do not offer copying an already revoked secret")
        app.buttons["Quay lại"].tap()
        XCTAssertTrue(app.buttons["more-profile-account"].waitForExistence(timeout: 10))
        app.buttons["more-profile-account"].tap()
        XCTAssertTrue(app.textFields["IANA timezone"].waitForExistence(timeout: 10))
        XCTAssertFalse(app.staticTexts[keyName].waitForExistence(timeout: 5), "Revoked metadata must not resurrect the active key row after reload")
        attachScreenshot(app, named: "account-api-key-revocation-retained")
#endif
    }

    func testAPIBackedAIEntryDisablesAttachmentsWhenFilesUnavailable() throws {
#if !LOCAL_API_UI
        throw XCTSkip("Build with SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI and run against the disposable local API fixture.")
#else
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        openMore(app)
        XCTAssertTrue(app.buttons["Nhập giao dịch bằng AI"].waitForExistence(timeout: 10))
        let aiEntry = app.buttons["Nhập giao dịch bằng AI"]
        reveal(aiEntry, in: app)
        aiEntry.tap()
        XCTAssertTrue(app.staticTexts["Nhập giao dịch bằng AI"].firstMatch.waitForExistence(timeout: 10))
        XCTAssertTrue(app.staticTexts["OCR chưa được cấu hình."].waitForExistence(timeout: 10))
        let attachmentPicker = app.buttons["Đính kèm ảnh hoặc PDF"]
        XCTAssertTrue(attachmentPicker.waitForExistence(timeout: 10))
        XCTAssertFalse(attachmentPicker.isEnabled)
#endif
    }

    func testAIEntryConfiguredAttachmentLimitAndThumbnails() throws {
#if !LOCAL_API_UI
        throw XCTSkip("Build with SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI to run the deterministic attachment fixture.")
#else
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated", "-UITestAIEntryAttachments"]
        app.launch()
        openMore(app)
        XCTAssertTrue(app.buttons["Nhập giao dịch bằng AI"].waitForExistence(timeout: 10))
        let aiEntry = app.buttons["Nhập giao dịch bằng AI"]
        reveal(aiEntry, in: app)
        aiEntry.tap()
        XCTAssertTrue(app.staticTexts["Nhập giao dịch bằng AI"].firstMatch.waitForExistence(timeout: 10))

        let removeButtons = app.buttons.matching(
            NSPredicate(format: "label == %@", "Bỏ tệp")
        )
        XCTAssertEqual(removeButtons.count, 20)
        XCTAssertFalse(app.buttons["Đính kèm ảnh hoặc PDF"].isEnabled)
        removeButtons.firstMatch.tap()
        XCTAssertEqual(removeButtons.count, 19)
        XCTAssertTrue(app.buttons["Đính kèm ảnh hoặc PDF"].isEnabled)
#endif
    }

    func testAIEntryUncertainRequestRequiresGETResolutionBeforeResubmission() throws {
#if !LOCAL_API_UI
        throw XCTSkip("Build with SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI to run the deterministic recovery fixture.")
#else
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated", "-UITestAIEntryRecovery404"]
        app.launch()
        openMore(app)
        let aiEntry = app.buttons["Nhập giao dịch bằng AI"]
        XCTAssertTrue(aiEntry.waitForExistence(timeout: 10))
        reveal(aiEntry, in: app)
        aiEntry.tap()

        XCTAssertTrue(app.staticTexts["Yêu cầu có thể đã được tiếp nhận. Tải lại kết quả để kiểm tra trước khi gửi nội dung khác. Không tự động gửi lại."].waitForExistence(timeout: 10))
        let analyze = app.buttons["ai-entry-process"]
        let description = app.textViews["ai-entry-description"]
        XCTAssertTrue(analyze.waitForExistence(timeout: 5))
        XCTAssertTrue(description.exists)
        XCTAssertFalse(analyze.isEnabled, "An uncertain request must not be posted again before its request ID is checked.")
        XCTAssertFalse(description.isEnabled, "The original input must stay unchanged until its request is resolved.")
        XCTAssertTrue(app.buttons["ai-entry-refresh-result"].isEnabled)

        app.buttons["ai-entry-refresh-result"].tap()
        XCTAssertTrue(app.staticTexts["Chưa tìm thấy yêu cầu đã tiếp nhận. Bạn có thể gửi lại nội dung đang soạn."].waitForExistence(timeout: 5))
        XCTAssertTrue(analyze.isEnabled, "Only a GET 404 resolution may reopen explicit submission.")
        XCTAssertTrue(description.isEnabled)
        XCTAssertEqual(description.value as? String, "Chi 50.000 cho cà phê hôm nay")
#endif
    }

    func testAIEntryAttachmentSourcesDefaultToPhotosAndCancelPreservesFiles() throws {
#if !LOCAL_API_UI
        throw XCTSkip("Requires deterministic attachment fixture.")
#else
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated", "-UITestAIEntryAttachments"]
        app.launch()
        openMore(app)
        let entry = app.buttons["Nhập giao dịch bằng AI"]
        reveal(entry, in: app); entry.tap()
        let removals = app.buttons.matching(NSPredicate(format: "label == %@", "Bỏ tệp"))
        XCTAssertEqual(removals.count, 20)
        removals.firstMatch.tap()
        let add = app.buttons["Đính kèm ảnh hoặc PDF"]
        reveal(add, in: app); add.tap()
        guard app.staticTexts["attachment-source-picker"].waitForExistence(timeout: 5) else {
            XCTFail("Attachments must show the common Photos-first source chooser, not open Files directly.")
            return
        }
        XCTAssertTrue(app.buttons["segment-photos"].isSelected)
        XCTAssertTrue(app.buttons["attachment-open-photos"].exists)
        XCTAssertEqual(app.staticTexts["attachment-remaining-capacity"].label, "Còn 1 tệp")
        app.buttons["segment-files"].tap()
        XCTAssertTrue(app.buttons["attachment-open-files"].waitForExistence(timeout: 5), app.debugDescription)
        XCTAssertFalse(app.buttons["attachment-open-photos"].exists)
        app.buttons["segment-photos"].tap()
        XCTAssertTrue(app.buttons["attachment-open-photos"].waitForExistence(timeout: 5), app.debugDescription)
        attachScreenshot(app, named: "photos-first-source-and-remaining-capacity")
        app.buttons["screen-close"].firstMatch.tap()
        XCTAssertTrue(add.waitForExistence(timeout: 5))
        XCTAssertEqual(removals.count, 19, "Canceling source selection must preserve existing documents.")
#endif
    }

    func testAIEntrySystemPhotosAndFilesCancelPreserveExistingAttachments() throws {
#if !LOCAL_API_UI
        throw XCTSkip("Requires deterministic attachment fixture.")
#else
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated", "-UITestAIEntryAttachments"]
        app.launch()
        openMore(app)
        let entry = app.buttons["Nhập giao dịch bằng AI"]
        reveal(entry, in: app); entry.tap()
        let removals = app.buttons.matching(NSPredicate(format: "label == %@", "Bỏ tệp"))
        removals.firstMatch.tap()
        let add = app.buttons["Đính kèm ảnh hoặc PDF"]
        reveal(add, in: app); add.tap()
        XCTAssertTrue(app.buttons["attachment-open-photos"].waitForExistence(timeout: 5))
        app.buttons["attachment-open-photos"].tap()
        let photoCancel = app.buttons["Cancel"].firstMatch
        guard photoCancel.waitForExistence(timeout: 10) else {
            attachScreenshot(app, named: "system-photos-cancel-missing")
            XCTFail("System Photos picker must open with a reachable Cancel action. \(app.debugDescription)")
            return
        }
        attachScreenshot(app, named: "system-photos-picker")
        let pickerHierarchy = XCTAttachment(string: app.debugDescription)
        pickerHierarchy.name = "system-photos-picker-hierarchy"
        pickerHierarchy.lifetime = .keepAlways
        self.add(pickerHierarchy)
        photoCancel.tap()
        XCTAssertTrue(app.buttons["attachment-open-photos"].waitForExistence(timeout: 5))
        app.buttons["segment-files"].tap()
        XCTAssertTrue(app.buttons["attachment-open-files"].waitForExistence(timeout: 5))
        app.buttons["attachment-open-files"].tap()
        let fileCancel = app.buttons["Cancel"].firstMatch
        guard fileCancel.waitForExistence(timeout: 10) else {
            attachScreenshot(app, named: "system-files-cancel-missing")
            XCTFail("System Files picker must open with a reachable Cancel action. \(app.debugDescription)")
            return
        }
        attachScreenshot(app, named: "system-files-picker")
        fileCancel.tap()
        XCTAssertTrue(app.buttons["attachment-open-files"].waitForExistence(timeout: 5))
        app.buttons["screen-close"].firstMatch.tap()
        XCTAssertTrue(add.waitForExistence(timeout: 5))
        XCTAssertEqual(removals.count, 19, "Canceling both system pickers must not drop existing files.")
#endif
    }

    func testAIEntryProposalReviewRequiresExplicitType() throws {
#if !LOCAL_API_UI
        throw XCTSkip("Build with SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI to run the deterministic review fixture.")
#else
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated", "-UITestAIEntryProposal"]
        app.launch()
        openMore(app)
        XCTAssertTrue(app.buttons["Nhập giao dịch bằng AI"].waitForExistence(timeout: 10))
        let aiEntry = app.buttons["Nhập giao dịch bằng AI"]
        reveal(aiEntry, in: app)
        aiEntry.tap()
        XCTAssertTrue(app.staticTexts["Chưa xác định — chọn loại"].waitForExistence(timeout: 10))
        XCTAssertTrue(app.buttons["ai-entry-save-all"].waitForExistence(timeout: 10))
        app.buttons["Sửa"].tap()
        XCTAssertTrue(app.buttons["Loại giao dịch: Chưa xác định — chọn loại"].waitForExistence(timeout: 10))
        app.buttons["Lưu đề xuất"].tap()
        XCTAssertTrue(app.staticTexts["Chọn loại giao dịch là thu nhập hoặc chi tiêu trước khi lưu."].waitForExistence(timeout: 5))
        app.buttons["Loại giao dịch: Chưa xác định — chọn loại"].tap()
        XCTAssertTrue(app.buttons["Chi tiêu"].waitForExistence(timeout: 5))
        app.buttons["Chi tiêu"].tap()
        XCTAssertTrue(app.buttons["Hũ chi tiêu: Không chọn"].waitForExistence(timeout: 5))
        XCTAssertTrue(app.staticTexts["Ngày và giờ"].exists)
        let dateTimeField = app.buttons["datetime-field-Ngày và giờ"]
        XCTAssertTrue(dateTimeField.waitForExistence(timeout: 5))
        dateTimeField.tap()
        XCTAssertTrue(app.staticTexts["screen-title-Chọn ngày và giờ"].waitForExistence(timeout: 5))
        XCTAssertTrue(app.descendants(matching: .any)["date-time-calendar-time"].waitForExistence(timeout: 5))
        attachScreenshot(app, named: "ai-entry-date-time-calendar-parity")
        app.buttons["date-calendar-cancel"].tap()
        XCTAssertTrue(app.buttons["datetime-field-Ngày và giờ"].waitForExistence(timeout: 5))
        XCTAssertTrue(app.buttons["Tính vào báo cáo: Có"].waitForExistence(timeout: 5))
#endif
    }

    func testAIEntrySelectingSystemPhotoAppendsWithoutReplacingExistingFiles() throws {
#if !LOCAL_API_UI
        throw XCTSkip("Requires deterministic existing attachments and Simulator stock Photos.")
#else
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated", "-UITestAIEntryAttachments"]
        app.launch()
        openMore(app)
        let entry = app.buttons["Nhập giao dịch bằng AI"]
        reveal(entry, in: app); entry.tap()
        let removals = app.buttons.matching(NSPredicate(format: "label == %@", "Bỏ tệp"))
        XCTAssertEqual(removals.count, 20)
        removals.firstMatch.tap()
        let attachment = app.buttons["Đính kèm ảnh hoặc PDF"]
        reveal(attachment, in: app); attachment.tap()
        XCTAssertTrue(app.buttons["attachment-open-photos"].waitForExistence(timeout: 5))
        app.buttons["attachment-open-photos"].tap()
        // Observed from the actual native picker hierarchy, independent of its
        // localized photo dates. The Simulator has stock, non-user images.
        let photo = app.images.matching(identifier: "PXGGridLayout-Info").firstMatch
        guard photo.waitForExistence(timeout: 10) else {
            XCTFail("Simulator stock photo must be selectable. \(app.debugDescription)")
            return
        }
        photo.tap()
        let complete = XCTNSPredicateExpectation(predicate: NSPredicate { _, _ in
            removals.count == 20 && !app.staticTexts["attachment-source-picker"].exists
        }, object: app)
        XCTAssertEqual(XCTWaiter.wait(for: [complete], timeout: 15), .completed,
                       "A real selected photo must append and close the source chooser. \(app.debugDescription)")
        XCTAssertFalse(attachment.isEnabled, "The20-file selection must be full after importing one photo.")
        for index in 2...20 {
            XCTAssertTrue(app.staticTexts[String(format: "IMG_%02d.PNG", index)].exists,
                          "Import must not replace previously selected files.")
        }
        let imported = app.staticTexts.matching(NSPredicate(format: "label MATCHES %@", "IMG_[0-9A-Fa-f-]+\\.(jpg|png)")).firstMatch
        XCTAssertTrue(imported.waitForExistence(timeout: 5), "The real image must have a supported normalized filename.")
        reveal(imported, in: app)
        attachScreenshot(app, named: "system-photo-imported-next-to-existing-files")
#endif
    }

    func testAIEntryHeaderAndAnalyzeRemainPinnedWhileAttachmentsScroll() throws {
#if !LOCAL_API_UI
        throw XCTSkip("Requires deterministic20 attachment fixture.")
#else
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated", "-UITestAIEntryAttachments"]
        app.launch()
        openMore(app)
        let entry = app.buttons["Nhập giao dịch bằng AI"]
        reveal(entry, in: app); entry.tap()
        let header = app.staticTexts["screen-title-Nhập giao dịch bằng AI"]
        XCTAssertTrue(header.waitForExistence(timeout: 5))
        let headerBefore = header.frame
        let analyze = app.buttons["ai-entry-process"]
        let viewport = app.descendants(matching: .any)["shell-content-viewport"]
        XCTAssertTrue(viewport.frame.contains(analyze.frame), "Analyze must be pinned, not below20 attachments.")
        for _ in 0..<4 { app.scrollViews.firstMatch.swipeUp() }
        XCTAssertEqual(header.frame.minY, headerBefore.minY, accuracy: 1, "AI entry header must not scroll away.")
        XCTAssertTrue(header.isHittable)
        XCTAssertTrue(viewport.frame.contains(analyze.frame))
        let feedback = app.buttons["floating-feedback-bubble"]
        XCTAssertFalse(feedback.frame.intersects(analyze.frame), "Feedback must not cover the pinned Analyze action.")
        let lastRemove = app.buttons.matching(NSPredicate(format: "label == %@", "Bỏ tệp")).element(boundBy: 19)
        XCTAssertFalse(feedback.frame.intersects(lastRemove.frame), "Moving Feedback above the footer must not block the last attachment's remove control.")
        let preview = try XCTUnwrap(app.images.allElementsBoundByIndex.first {
            abs($0.frame.midY - lastRemove.frame.midY) < 1
                && abs($0.frame.width - 44) < 1 && $0.frame.minX < 100
        })
        XCTAssertFalse(feedback.frame.intersects(preview.frame), "Feedback must not hide the thumbnail needed to identify the last document.")
        attachScreenshot(app, named: "ai-entry-pinned-header-and-process-footer")
#endif
    }

    func testAIEntrySaveAllStaysReachableAfterProposalScroll() throws {
#if !LOCAL_API_UI
        throw XCTSkip("Requires deterministic pending proposal fixture.")
#else
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated", "-UITestAIEntryProposal"]
        app.launch()
        openMore(app)
        let entry = app.buttons["Nhập giao dịch bằng AI"]
        reveal(entry, in: app); entry.tap()
        let save = app.buttons["ai-entry-save-all"]
        XCTAssertTrue(save.waitForExistence(timeout: 5))
        let before = save.frame
        app.scrollViews.firstMatch.swipeUp()
        XCTAssertEqual(save.frame.minY, before.minY, accuracy: 1)
        XCTAssertTrue(save.isHittable, "Save-all must stay reachable while reviewing proposals.")
        let feedback = app.buttons["floating-feedback-bubble"]
        XCTAssertFalse(feedback.frame.intersects(save.frame))
        XCTAssertFalse(feedback.frame.intersects(app.buttons["Lưu giao dịch"].frame), "Feedback must not cover proposal approval.")
        attachScreenshot(app, named: "ai-entry-save-all-pinned-during-review")
#endif
    }

    func testAIEntryProposalEditorReservesShellInset() throws {
#if !LOCAL_API_UI
        throw XCTSkip("Build with SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI to run the deterministic proposal editor fixture.")
#else
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated", "-UITestAIEntryProposal"]
        app.launch()
        openMore(app)
        let aiEntry = app.buttons["Nhập giao dịch bằng AI"]
        XCTAssertTrue(aiEntry.waitForExistence(timeout: 10))
        reveal(aiEntry, in: app)
        aiEntry.tap()
        XCTAssertTrue(app.staticTexts["Nhập giao dịch bằng AI"].firstMatch.waitForExistence(timeout: 10))

        let edit = app.buttons["Sửa"].firstMatch
        XCTAssertTrue(edit.waitForExistence(timeout: 10))
        edit.tap()
        XCTAssertTrue(app.staticTexts["Sửa đề xuất"].waitForExistence(timeout: 10))
        XCTAssertTrue(
            app.descendants(matching: .any)["ai-entry-proposal-editor-shell-inset"].waitForExistence(timeout: 2),
            "AI entry proposal editor must reserve the persistent shell navigation."
        )
        let save = app.buttons["ai-entry-proposal-save"]
        XCTAssertTrue(save.waitForExistence(timeout: 10))
        reveal(save, in: app)
        let shellTop = app.buttons["bottom-tab-account"].frame.minY
        XCTAssertLessThanOrEqual(
            save.frame.maxY,
            shellTop,
            "AI entry proposal editor save action must stay above the shell navigation."
        )
        attachScreenshot(app, named: "ai-entry-proposal-editor-shell-clearance")
#endif
    }

    func testAPIBackedLedgerUsesWeekModeAndCashflowCard() async throws {
#if !LOCAL_API_UI
        throw XCTSkip("Build with SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI and run against the disposable local API fixture.")
#else
        let suffix = String(UUID().uuidString.prefix(8))
        let wallet = try await LocalAPIFixture.wallet(name: "Fixture Ledger \(suffix)", type: "basic")
        addTeardownBlock {
            try? await LocalAPIFixture.deleteTransactions(walletID: wallet.wallet.id)
            try? await LocalAPIFixture.delete(wallet)
        }
        let category = try await LocalAPIFixture.category(kind: "expense", walletID: wallet.wallet.id)
        for index in 0..<12 {
            _ = try await LocalAPIFixture.transaction(
                walletID: wallet.wallet.id,
                categoryID: category.id,
                amount: Int64(10_000 + index),
                note: "Fixture ledger row (suffix)-(index)"
            )
        }

        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        XCTAssertTrue(app.buttons["Giao dịch"].waitForExistence(timeout: 10))
        app.buttons["Giao dịch"].tap()
        XCTAssertTrue(app.staticTexts["Giao dịch"].firstMatch.waitForExistence(timeout: 10))
        XCTAssertTrue(app.buttons["Theo tuần"].waitForExistence(timeout: 10))
        XCTAssertTrue(app.buttons["Tùy chọn"].exists)
        XCTAssertTrue(app.buttons["Tuần trước"].exists)
        app.buttons["Tùy chọn"].tap()
        let customStart = app.buttons["date-field-Từ ngày"]
        XCTAssertTrue(customStart.waitForExistence(timeout: 10))
        customStart.tap()
        XCTAssertTrue(app.staticTexts["screen-title-Chọn ngày"].waitForExistence(timeout: 5))
        XCTAssertTrue(app.descendants(matching: .any)["date-calendar-picker"].exists)
        app.buttons["date-calendar-cancel"].tap()
        XCTAssertTrue(app.buttons["date-field-Đến ngày"].waitForExistence(timeout: 5))
        XCTAssertTrue(app.staticTexts["Tiền vào"].exists)
        XCTAssertTrue(app.staticTexts["Tiền ra"].exists)
        XCTAssertFalse(app.buttons["Tất cả"].exists)
        let ledgerRows = app.buttons.matching(NSPredicate(format: "identifier BEGINSWITH 'transaction-row-'")).allElementsBoundByIndex
        XCTAssertFalse(ledgerRows.isEmpty, "The ledger should expose stable transaction rows for shell-clearance verification.")
        for _ in 0..<8 { app.scrollViews.firstMatch.swipeUp() }
        let shellTop = app.buttons["bottom-tab-account"].frame.minY
        let visibleBottomRow = ledgerRows.filter(\.exists).max { $0.frame.maxY < $1.frame.maxY }
        XCTAssertNotNil(visibleBottomRow, "The ledger should retain a visible row at the end of the scroll range.")
        if let visibleBottomRow {
            XCTAssertLessThanOrEqual(
                visibleBottomRow.frame.maxY,
                shellTop,
                "The last transaction rows must remain above persistent shell navigation."
            )
        }
        attachScreenshot(app, named: "transaction-ledger-stitch-parity")
#endif
    }

    func testAPIBackedTransactionPresentationIsConsistentAcrossOverviewLedgerAndWallet() async throws {
#if !LOCAL_API_UI
        throw XCTSkip("Requires disposable LOCAL_API_UI fixture")
#else
        let fixture = try await LocalAPIFixture.wallet(name: "Consistent rows \(UUID().uuidString.prefix(8))", type: "basic")
        let category = try await LocalAPIFixture.category(kind: "expense", walletID: fixture.wallet.id)
        let note = "Cà phê buổi sáng cùng đồng nghiệp; ghi chú dài để kiểm tra cách xuống dòng trên iPhone nhỏ"
        let transaction = try await LocalAPIFixture.transaction(walletID: fixture.wallet.id, categoryID: category.id, amount: 45_000, note: note)
        addTeardownBlock {
            try? await LocalAPIFixture.deleteTransactions(walletID: fixture.wallet.id)
            try? await LocalAPIFixture.delete(fixture)
        }
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        let overviewRow = app.descendants(matching: .any)["overview-transaction-\(transaction.id)"].firstMatch
        let showAmounts = app.buttons["icon-Hiện số dư"]
        if showAmounts.waitForExistence(timeout: 2) { showAmounts.tap() }
        reveal(overviewRow, in: app)
        XCTAssertTrue(overviewRow.waitForExistence(timeout: 10), app.debugDescription)
        for part in [note, "45.000", "−"] { XCTAssertTrue(overviewRow.label.contains(part), "Overview must use the shared transaction presentation: \(part)") }
        attachScreenshot(app, named: "shared-transaction-overview")
        app.buttons["bottom-tab-transactions"].tap()
        let ledgerRow = app.buttons["transaction-row-\(transaction.id)"]
        XCTAssertTrue(ledgerRow.waitForExistence(timeout: 10))
        reveal(ledgerRow, in: app)
        for part in [note, "45.000", "−"] { XCTAssertTrue(ledgerRow.label.contains(part), "Ledger must use the shared transaction presentation: \(part)") }
        XCTAssertLessThanOrEqual(ledgerRow.frame.maxY, app.buttons["bottom-tab-account"].frame.minY)
        attachScreenshot(app, named: "shared-transaction-ledger")
        openWallets(app)
        let detail = app.buttons["wallet-detail-\(fixture.wallet.id)"]
        XCTAssertTrue(detail.waitForExistence(timeout: 10))
        reveal(detail, in: app)
        detail.tap()
        let walletRow = app.buttons["wallet-transaction-\(transaction.id)"]
        XCTAssertTrue(walletRow.waitForExistence(timeout: 10))
        reveal(walletRow, in: app)
        for part in [note, "45.000", "−"] { XCTAssertTrue(walletRow.label.contains(part), "Wallet must use the shared transaction presentation: \(part)") }
        attachScreenshot(app, named: "shared-transaction-wallet")
#endif
    }

    func testManagementTypingHidesFeedbackUntilKeyboardDismissal() throws {
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated"]
        app.launch()
        openMore(app)
        let budgets = app.buttons["Ngân sách"]
        reveal(budgets, in: app); budgets.tap()
        app.buttons["budget-add"].tap()
        let name = app.textFields["Tên"]
        XCTAssertTrue(name.waitForExistence(timeout: 5))
        name.tap()
        name.typeText("Draft")
        XCTAssertTrue(app.keyboards.firstMatch.waitForExistence(timeout: 5))
        XCTAssertFalse(app.buttons["floating-feedback-bubble"].exists, "Auxiliary bubble must not intercept input/clear targets while typing.")
        app.buttons["clear-field-Tên"].tap()
        name.tap(); name.typeText("Retained draft")
        XCTAssertFalse(app.staticTexts["screen-title-Phản hồi"].exists)
        dismissKeyboard(app)
        let bubble = app.buttons["floating-feedback-bubble"]
        XCTAssertTrue(bubble.waitForExistence(timeout: 5))
        XCTAssertEqual(name.value as? String, "Retained draft")
        XCTAssertFalse(bubble.frame.intersects(app.buttons["Lưu"].frame))
    }

    func testAPIBackedManagementEditorsReplaceListAndKeepCancelReachable() throws {
#if !LOCAL_API_UI
        throw XCTSkip("Requires disposable local API fixture.")
#else
        let cases = [("Ngân sách", "budget-add", "Tạo ngân sách"),
                     ("Hũ chi tiêu", "jar-add-month", "Tạo hũ"),
                     ("Giao dịch định kỳ", "recurring-add", "Tạo lịch"),
                     ("Travel Mode", "travel-add", "Chuyến đi mới")]
        for (title, addID, editorTitle) in cases {
            let app = XCUIApplication()
            app.launchArguments = ["-UITestAuthenticatedAPI"]
            app.launch()
            openMore(app)
            let entry = app.buttons[title]
            XCTAssertTrue(entry.waitForExistence(timeout: 10))
            reveal(entry, in: app)
            entry.tap()
            let add = app.buttons[addID]
            XCTAssertTrue(add.waitForExistence(timeout: 10), title)
            reveal(add, in: app)
            add.tap()
            XCTAssertTrue(app.staticTexts[editorTitle].waitForExistence(timeout: 5))
            XCTAssertFalse(app.staticTexts["screen-title-\(title)"].exists, "\(title) must replace its list while editing.")
            let save = app.buttons["Lưu"].firstMatch
            XCTAssertTrue(save.exists)
            XCTAssertFalse(app.buttons["floating-feedback-bubble"].frame.intersects(save.frame), "Feedback must not cover \(title) editor save.")
            let close = app.buttons["screen-close"]
            if close.exists {
                XCTAssertTrue(close.isHittable)
                close.tap()
                XCTAssertTrue(app.staticTexts["screen-title-\(title)"].waitForExistence(timeout: 5))
            } else { XCTFail("\(title) lacks common editor cancel header") }
            app.terminate()
        }
#endif
    }

    func testAPIBackedSavingsReportSeparatesGoalTransfersAndAccountSurplus() async throws {
#if !LOCAL_API_UI
        throw XCTSkip("Requires disposable local API with LOCAL_API_UI.")
#else
        XCUIDevice.shared.orientation = .portrait
        let suffix = String(UUID().uuidString.prefix(8))
        let source = try await LocalAPIFixture.wallet(name: "SB Cash \(suffix)", type: "basic")
        let goal = try await LocalAPIFixture.wallet(name: "SB Goal \(suffix)", type: "goal")
        addTeardownBlock {
            try? await LocalAPIFixture.deleteTransferPairs(walletIDs: [source.wallet.id, goal.wallet.id])
            try? await LocalAPIFixture.deleteTransactions(walletID: source.wallet.id)
            try? await LocalAPIFixture.delete(source)
            try? await LocalAPIFixture.delete(goal)
        }
        let incomeCategory = try await LocalAPIFixture.category(kind: "income", walletID: source.wallet.id)
        let expenseCategory = try await LocalAPIFixture.category(kind: "expense", walletID: source.wallet.id)
        let income = try await LocalAPIFixture.transaction(walletID: source.wallet.id, categoryID: incomeCategory.id,
            amount: 20_000_000, note: "SB Income \(suffix)", type: "income")
        let expense = try await LocalAPIFixture.transaction(walletID: source.wallet.id, categoryID: expenseCategory.id,
            amount: 12_000_000, note: "SB Expense \(suffix)")
        let transferNote = "SB Transfer \(suffix)"
        try await LocalAPIFixture.savingsBudgetTransfer(source: source.wallet.id, destination: goal.wallet.id,
                                                       amount: 5_000_000, note: transferNote)
        let month = LocalAPIFixture.currentMonth()
        let report = try await LocalAPIFixture.savingsBudgetRead("/reports/savings?month=\(month)", as: SavingsReportFixture.self)
        // A goal wallet created during the selected month starts at zero for
        // the period opening; its configured seed is reported separately as
        // initial_balance_added. This matches the savings report contract.
        XCTAssertEqual(report.goals.opening_balance, 0)
        XCTAssertEqual(report.goals.closing_balance, 5_100_000)
        XCTAssertEqual(report.goals.initial_balance_added, 100_000)
        XCTAssertEqual(report.goals.deposits, 0, "Paired transfer is separate from ordinary deposits")
        XCTAssertEqual(report.goals.internal_transfers, 5_000_000)
        XCTAssertEqual(report.surplus.income, 20_000_000)
        XCTAssertEqual(report.surplus.expense, 12_000_000)
        XCTAssertEqual(report.surplus.net, 8_000_000)
        let goalRows = try await LocalAPIFixture.savingsBudgetRead("/reports/savings/transactions?month=\(month)&section=goals", as: SavingsBudgetPageFixture.self)
        let surplusRows = try await LocalAPIFixture.savingsBudgetRead("/reports/savings/transactions?month=\(month)&section=surplus", as: SavingsBudgetPageFixture.self)
        XCTAssertEqual(goalRows.total, 1)
        XCTAssertEqual(goalRows.items.map(\.amount), [5_000_000])
        XCTAssertEqual(Set(surplusRows.items.map(\.id)), Set([income.id, expense.id]))
        XCTAssertEqual(surplusRows.total, 2)

        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI", "-mypocket.privacy.masked", "NO"]
        app.launch()
        openMore(app)
        let reports = app.buttons["more-destination-reports"]
        XCTAssertTrue(reports.waitForExistence(timeout: 10)); reveal(reports, in: app); reports.tap()
        let savings = app.buttons["reports-open-savings"]
        XCTAssertTrue(savings.waitForExistence(timeout: 10)); reveal(savings, in: app); savings.tap()
        XCTAssertTrue(app.textFields["savings-report-month"].waitForExistence(timeout: 10))
        XCTAssertEqual(app.textFields["savings-report-month"].value as? String, month)
        assertSavingsBudgetMetric("Số dư cuối kỳ", amount: report.goals.closing_balance, in: app)
        assertSavingsBudgetMetric("Chuyển ví (thuần)", amount: report.goals.internal_transfers, in: app)
        attachScreenshot(app, named: "savings-goal-balances-and-transfer-separation")
        let goalDrilldown = app.buttons["Giao dịch ví mục tiêu"]
        reveal(goalDrilldown, in: app); goalDrilldown.tap()
        assertSavingsBudgetNote(transferNote, in: app)
        XCTAssertTrue(app.staticTexts["1–1 / 1 giao dịch"].exists)
        attachScreenshot(app, named: "savings-goal-api-matching-transfer-row")
        app.buttons["Quay lại"].firstMatch.tap()
        let surplusTitle = app.staticTexts["Thặng dư thu / chi · Toàn tài khoản"]
        XCTAssertTrue(surplusTitle.waitForExistence(timeout: 10)); reveal(surplusTitle, in: app)
        assertSavingsBudgetMetric("Thặng dư", amount: report.surplus.net, in: app)
        attachScreenshot(app, named: "savings-account-surplus-distinct-section")
        let surplusDrilldown = app.buttons["Giao dịch thu / chi"]
        reveal(surplusDrilldown, in: app); surplusDrilldown.tap()
        assertSavingsBudgetNote(try XCTUnwrap(income.note), in: app)
        assertSavingsBudgetNote(try XCTUnwrap(expense.note), in: app)
        XCTAssertTrue(app.staticTexts["1–2 / 2 giao dịch"].exists)
        XCTAssertFalse(app.staticTexts.matching(NSPredicate(format: "label CONTAINS %@", transferNote)).firstMatch.exists)
        attachScreenshot(app, named: "savings-surplus-api-matching-income-expense-rows")
#endif
    }

    func testAPIBackedRecurringBudgetCreateForecastEditNextAndStop() async throws {
#if !LOCAL_API_UI
        throw XCTSkip("Requires disposable local API with LOCAL_API_UI.")
#else
        XCUIDevice.shared.orientation = .portrait
        let suffix = String(UUID().uuidString.prefix(8))
        let wallet = try await LocalAPIFixture.wallet(name: "SB Budget Wallet \(suffix)", type: "basic")
        addTeardownBlock {
            try? await LocalAPIFixture.deleteTransactions(walletID: wallet.wallet.id)
            try? await LocalAPIFixture.delete(wallet)
        }
        let category = try await LocalAPIFixture.category(kind: "expense", walletID: wallet.wallet.id)
        let first = try await LocalAPIFixture.transaction(walletID: wallet.wallet.id, categoryID: category.id,
            amount: 120_000, note: "SB Budget First \(suffix)")
        let second = try await LocalAPIFixture.transaction(walletID: wallet.wallet.id, categoryID: category.id,
            amount: 180_000, note: "SB Budget Second \(suffix)")
        let excluded = try await LocalAPIFixture.transaction(walletID: wallet.wallet.id, categoryID: category.id,
            amount: 90_000, note: "SB Budget Excluded \(suffix)", includedInReports: false)
        let name = "SB Monthly \(suffix)"
        let nextName = "SB Next \(suffix)"
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI", "-mypocket.privacy.masked", "NO"]
        app.launch(); openMore(app)
        let budgets = app.buttons["more-destination-budgets"]
        XCTAssertTrue(budgets.waitForExistence(timeout: 10)); reveal(budgets, in: app); budgets.tap()
        let add = app.buttons["budget-add"]
        XCTAssertTrue(add.waitForExistence(timeout: 10)); reveal(add, in: app); add.tap()
        let nameField = app.textFields["Tên"]
        XCTAssertTrue(nameField.waitForExistence(timeout: 10)); nameField.clearAndType(name); dismissKeyboard(app)
        let limit = app.textFields["Giới hạn"]
        reveal(limit, in: app); limit.clearAndType("1000000"); dismissKeyboard(app)
        let recurring = app.buttons["budget-recurring-toggle"]
        reveal(recurring, in: app); recurring.tap()
        let carry = app.buttons["budget-carryover-toggle"]
        XCTAssertTrue(carry.waitForExistence(timeout: 5)); reveal(carry, in: app); carry.tap()
        let walletScope = app.buttons["form-row-Phạm vi ví"]
        reveal(walletScope, in: app); walletScope.tap()
        let walletChoice = app.buttons[wallet.wallet.name]
        XCTAssertTrue(walletChoice.waitForExistence(timeout: 5)); walletChoice.tap()
        XCTAssertTrue(walletScope.label.contains(wallet.wallet.name))
        app.buttons["Lưu"].firstMatch.tap()
        XCTAssertTrue(app.staticTexts[name].firstMatch.waitForExistence(timeout: 15), "Native create must leave editor after real API success")

        let seriesList = try await LocalAPIFixture.savingsBudgetRead("/budget-series", as: SavingsBudgetListFixture<SavingsBudgetSeriesFixture>.self)
        let series = try XCTUnwrap(seriesList.items.first { $0.current_config.name == name })
        XCTAssertEqual(series.cadence, "monthly"); XCTAssertTrue(series.active)
        XCTAssertTrue(series.current_config.carryover_enabled)
        XCTAssertEqual(series.current_config.wallet_id, wallet.wallet.id)
        XCTAssertEqual(series.current_config.limit_amount, 1_000_000)
        let periods = try await LocalAPIFixture.savingsBudgetRead("/budgets", as: SavingsBudgetListFixture<SavingsBudgetPeriodFixture>.self)
        let period = try XCTUnwrap(periods.items.first { $0.series_id == series.id })
        let report = try await LocalAPIFixture.savingsBudgetRead("/budgets/\(period.id)/report", as: SavingsBudgetReportFixture.self)
        let rows = try await LocalAPIFixture.savingsBudgetRead("/budgets/\(period.id)/transactions", as: SavingsBudgetPageFixture.self)
        XCTAssertEqual(report.base_limit_amount, 1_000_000); XCTAssertEqual(report.carryover_amount, 0)
        XCTAssertEqual(report.effective_limit_amount, 1_000_000); XCTAssertEqual(report.spent, 300_000)
        XCTAssertEqual(report.remaining, 700_000)
        XCTAssertEqual(report.daily.reduce(0) { $0 + $1.spent }, report.spent)
        XCTAssertEqual(report.forecast.status, "estimated")
        let elapsed = report.forecast.elapsed_days
        XCTAssertGreaterThan(elapsed, 0)
        XCTAssertEqual(report.forecast.projected_spent, (report.spent * report.forecast.total_days + elapsed - 1) / elapsed)
        XCTAssertEqual(rows.total, 2); XCTAssertEqual(Set(rows.items.map(\.id)), Set([first.id, second.id]))
        XCTAssertEqual(rows.items.reduce(0) { $0 + $1.amount }, report.spent)
        let reportButton = app.buttons["budget-report-\(period.id)"]
        XCTAssertTrue(reportButton.waitForExistence(timeout: 10)); reveal(reportButton, in: app); reportButton.tap()
        assertSavingsBudgetMetric("Đã chi", amount: report.spent, in: app)
        assertSavingsBudgetMetric("Chi tiêu cuối kỳ", amount: try XCTUnwrap(report.forecast.projected_spent), in: app)
        attachScreenshot(app, named: "recurring-budget-api-matching-forecast")
        let transactions = app.buttons["Xem giao dịch ngân sách"]
        reveal(transactions, in: app); transactions.tap()
        assertSavingsBudgetNote(try XCTUnwrap(first.note), in: app)
        assertSavingsBudgetNote(try XCTUnwrap(second.note), in: app)
        XCTAssertFalse(app.staticTexts.matching(NSPredicate(format: "label CONTAINS %@", try XCTUnwrap(excluded.note))).firstMatch.exists)
        XCTAssertTrue(app.staticTexts["1–2 / 2 giao dịch"].exists)
        attachScreenshot(app, named: "recurring-budget-drilldown-matches-spent-excludes-hidden")
        app.buttons["Quay lại"].firstMatch.tap()
        app.buttons["Quay lại"].firstMatch.tap()
        let edit = app.buttons["Sửa chu kỳ"].firstMatch
        XCTAssertTrue(edit.waitForExistence(timeout: 10)); reveal(edit, in: app); edit.tap()
        XCTAssertTrue(app.staticTexts["Sửa chu kỳ ngân sách"].waitForExistence(timeout: 5))
        app.textFields["Tên"].clearAndType(nextName); dismissKeyboard(app)
        reveal(app.textFields["Giới hạn"], in: app)
        app.textFields["Giới hạn"].clearAndType("2000000"); dismissKeyboard(app)
        app.buttons["Lưu"].firstMatch.tap()
        XCTAssertTrue(app.buttons["budget-report-\(period.id)"].waitForExistence(timeout: 15))
        let editedList = try await LocalAPIFixture.savingsBudgetRead("/budget-series", as: SavingsBudgetListFixture<SavingsBudgetSeriesFixture>.self)
        let edited = try XCTUnwrap(editedList.items.first { $0.id == series.id })
        XCTAssertEqual(edited.version, series.version + 1)
        XCTAssertEqual(edited.current_config.name, name)
        XCTAssertEqual(edited.current_config.limit_amount, 1_000_000)
        XCTAssertEqual(edited.pending_config?.name, nextName)
        XCTAssertEqual(edited.pending_config?.limit_amount, 2_000_000)
        let pendingLabel = app.staticTexts.matching(NSPredicate(format: "label BEGINSWITH %@ AND label CONTAINS %@", "Thay đổi từ", nextName)).firstMatch
        XCTAssertTrue(pendingLabel.waitForExistence(timeout: 10)); reveal(pendingLabel, in: app)
        attachScreenshot(app, named: "recurring-budget-next-boundary-pending-config")
        let stop = app.buttons["Dừng chu kỳ"].firstMatch
        reveal(stop, in: app); stop.tap()
        let confirm = app.alerts.buttons["Dừng"]
        XCTAssertTrue(confirm.waitForExistence(timeout: 5)); confirm.tap()
        let stoppedLabel = app.staticTexts.matching(NSPredicate(format: "label CONTAINS %@", "Đã dừng")).firstMatch
        XCTAssertTrue(stoppedLabel.waitForExistence(timeout: 15), "Confirm stop must persist inactive series and update UI")
        let stoppedList = try await LocalAPIFixture.savingsBudgetRead("/budget-series", as: SavingsBudgetListFixture<SavingsBudgetSeriesFixture>.self)
        let stopped = try XCTUnwrap(stoppedList.items.first { $0.id == series.id })
        XCTAssertFalse(stopped.active); XCTAssertEqual(stopped.version, edited.version + 1)
        let after = try await LocalAPIFixture.savingsBudgetRead("/budgets", as: SavingsBudgetListFixture<SavingsBudgetPeriodFixture>.self)
        XCTAssertEqual(after.items.first { $0.id == period.id }?.limit_amount, 1_000_000)
        let unchangedRows = try await LocalAPIFixture.savingsBudgetRead("/budgets/\(period.id)/transactions", as: SavingsBudgetPageFixture.self)
        XCTAssertEqual(Set(unchangedRows.items.map(\.id)), Set(rows.items.map(\.id)))
        attachScreenshot(app, named: "recurring-budget-stop-preserves-generated-period")
#endif
    }

#if LOCAL_API_UI
    private func assertSavingsBudgetMetric(_ title: String, amount: Int64, in app: XCUIApplication) {
        let formatter = NumberFormatter()
        formatter.numberStyle = .currency; formatter.currencyCode = "VND"; formatter.currencySymbol = "₫"
        formatter.maximumFractionDigits = 0; formatter.locale = Locale(identifier: "vi_VN")
        let formatted = formatter.string(from: NSNumber(value: amount))!
        // MPMetricGrid exposes each metric as an accessibility container with
        // label=title and value=formatted amount (the value is not a child
        // StaticText joined into one identifier). Query that stable contract
        // instead of depending on the visual text composition.
        let metric = app.descendants(matching: .any)
            .matching(NSPredicate(format: "label == %@", title)).firstMatch
        guard metric.waitForExistence(timeout: 10) else {
            XCTFail("Expected server metric label \(title) with amount \(formatted)")
            return
        }
        reveal(metric, in: app)
        XCTAssertEqual(metric.value as? String, formatted, "Expected server amount \(title): \(formatted)")
    }

    private func assertSavingsBudgetNote(_ note: String, in app: XCUIApplication) {
        let row = app.descendants(matching: .any).matching(NSPredicate(format: "label CONTAINS %@", note)).firstMatch
        guard row.waitForExistence(timeout: 10) else {
            XCTFail("Drilldown must expose real API row: \(note)")
            return
        }
        reveal(row, in: app)
    }
#endif

    func testAPIBackedBudgetCreateUpdateDeleteFromNativeUI() async throws {
#if !LOCAL_API_UI
        throw XCTSkip("Build with SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI and run against the disposable local API fixture.")
#else
        let suffix = String(UUID().uuidString.prefix(8))
        let nameValue = "Fixture Budget \(suffix)"
        let updatedNameValue = "Fixture Budget Updated \(suffix)"
        try? await LocalAPIFixture.deleteBudgets(named: nameValue)
        try? await LocalAPIFixture.deleteBudgets(named: updatedNameValue)
        addTeardownBlock {
            try? await LocalAPIFixture.deleteBudgets(named: nameValue)
            try? await LocalAPIFixture.deleteBudgets(named: updatedNameValue)
        }
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        openMore(app)
        let budgetsEntry = app.buttons["Ngân sách"]
        XCTAssertTrue(budgetsEntry.waitForExistence(timeout: 10))
        reveal(budgetsEntry, in: app)
        budgetsEntry.tap()
        XCTAssertTrue(app.staticTexts["Ngân sách"].firstMatch.waitForExistence(timeout: 10))

        let addBudget = app.buttons["budget-add"]
        XCTAssertTrue(addBudget.waitForExistence(timeout: 10))
        reveal(addBudget, in: app)
        addBudget.tap()
        XCTAssertTrue(app.staticTexts["Tạo ngân sách"].waitForExistence(timeout: 10))
        let name = app.textFields["Tên"]
        XCTAssertTrue(name.waitForExistence(timeout: 10))
        reveal(name, in: app)
        name.clearAndType(nameValue)
        dismissKeyboard(app)
        let limit = app.textFields["Giới hạn"]
        XCTAssertTrue(limit.waitForExistence(timeout: 10))
        reveal(limit, in: app)
        limit.clearAndType("500000")
        if app.buttons["Xong"].exists { app.buttons["Xong"].tap() }
        let budgetCategory = try await LocalAPIFixture.category(kind: "expense")
        let budgetCategoryRow = app.buttons["budget-category-scope-row"]
        XCTAssertTrue(budgetCategoryRow.waitForExistence(timeout: 5))
        reveal(budgetCategoryRow, in: app)
        budgetCategoryRow.tap()
        let budgetCategorySearch = app.textFields["category-picker-search"]
        XCTAssertTrue(budgetCategorySearch.waitForExistence(timeout: 5))
        let rootCategoryCards = app.descendants(matching: .any).matching(
            NSPredicate(format: "identifier BEGINSWITH %@", "category-picker-group-")
        )
        XCTAssertGreaterThan(
            rootCategoryCards.count,
            1,
            "The web picker renders each root category in its own card."
        )
        attachScreenshot(app, named: "Native category picker - root cards")
        budgetCategorySearch.clearAndType(budgetCategory.name)
        dismissKeyboard(app)
        let budgetCategoryChoice = app.buttons["category-choice-\(budgetCategory.id)"]
        XCTAssertTrue(budgetCategoryChoice.waitForExistence(timeout: 5))
        let expectedCategorySubtitle = budgetCategory.isSystem == true ? "Nhóm mặc định" : "Nhóm cá nhân"
        XCTAssertEqual(
            budgetCategoryChoice.label,
            "\(budgetCategory.name), \(expectedCategorySubtitle)",
            "The category choice should expose the web system/personal subtitle to assistive technology."
        )
        reveal(budgetCategoryChoice, in: app)
        budgetCategoryChoice.tap()
        XCTAssertTrue(app.buttons["budget-category-scope-row"].label.contains(budgetCategory.name))
        app.buttons["Lưu"].tap()
        let created = app.staticTexts[nameValue]
        XCTAssertTrue(created.waitForExistence(timeout: 15), "Budget did not render after POST: \(nameValue)")
        guard created.exists else { return }
        XCTAssertTrue(app.staticTexts["Tổng quan ngân sách"].waitForExistence(timeout: 10))
        XCTAssertTrue(app.staticTexts["Còn lại"].exists)
        XCTAssertTrue(
            app.staticTexts.matching(NSPredicate(format: "label BEGINSWITH %@", "Gợi ý chi/ngày:")).firstMatch.waitForExistence(timeout: 10),
            "Active budgets must expose the remaining daily-spend suggestion."
        )

        let createdBudget = try await LocalAPIFixture.budget(named: nameValue)
        XCTAssertEqual(createdBudget.categoryID, budgetCategory.id)
        let budgetEdit = app.buttons["budget-edit-\(createdBudget.id)"]
        XCTAssertTrue(budgetEdit.waitForExistence(timeout: 10))
        reveal(budgetEdit, in: app)
        budgetEdit.tap()
        XCTAssertTrue(app.staticTexts["Sửa ngân sách"].waitForExistence(timeout: 10))
        let editedName = app.textFields["Tên"]
        XCTAssertTrue(editedName.waitForExistence(timeout: 10))
        reveal(editedName, in: app)
        editedName.clearAndType(updatedNameValue)
        dismissKeyboard(app)
        let budgetEditSave = app.buttons["Lưu"].firstMatch
        app.buttons["screen-close"].tap()
        let unchangedBudget = try await LocalAPIFixture.budget(named: nameValue)
        XCTAssertEqual(unchangedBudget.id, createdBudget.id, "Cancel must not write the budget draft.")
        reveal(budgetEdit, in: app); budgetEdit.tap()
        XCTAssertTrue(app.staticTexts["Sửa ngân sách"].waitForExistence(timeout: 5))
        XCTAssertEqual(app.textFields["Tên"].value as? String, updatedNameValue, "Reopening the same budget must retain the canceled draft.")
        reveal(budgetEditSave, in: app)
        XCTAssertTrue(budgetEditSave.isHittable)
        budgetEditSave.tap()
        XCTAssertTrue(app.staticTexts[updatedNameValue].waitForExistence(timeout: 15))

        let budgetDelete = app.buttons["budget-delete-\(createdBudget.id)"]
        XCTAssertTrue(budgetDelete.waitForExistence(timeout: 10))
        reveal(budgetDelete, in: app)
        budgetDelete.tap()
        XCTAssertTrue(app.alerts.buttons["Xóa"].waitForExistence(timeout: 5))
        app.alerts.buttons["Xóa"].tap()
        XCTAssertFalse(app.staticTexts[updatedNameValue].waitForExistence(timeout: 5))
#endif
    }

    func testAPIBackedEndedBudgetPreviewHasNoWriteControlsAndDeleteCancelPreservesData() async throws {
#if !LOCAL_API_UI
        throw XCTSkip("Requires disposable local API fixture.")
#else
        let name = "Fixture Ended Budget \(UUID().uuidString.prefix(8))"
        addTeardownBlock { try? await LocalAPIFixture.deleteBudgets(named: name) }
        let created = try await LocalAPIFixture.endedBudget(name: name)
        let before = try await LocalAPIFixture.budget(named: name)
        XCTAssertEqual(before.ended, true)
        let beforeLedger = try await LocalAPIFixture.transactions().sorted { $0.id < $1.id }
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        openMore(app)
        let entry = app.buttons["Ngân sách"]
        reveal(entry, in: app); entry.tap()
        let preview = app.buttons["budget-preview-\(created.id)"]
        XCTAssertTrue(preview.waitForExistence(timeout: 10))
        reveal(preview, in: app); preview.tap()
        XCTAssertTrue(app.staticTexts["Xem ngân sách"].waitForExistence(timeout: 5))
        XCTAssertFalse(app.textFields["Tên"].exists)
        XCTAssertFalse(app.textFields["Giới hạn"].exists)
        XCTAssertFalse(app.buttons["Lưu"].exists)
        let remove = app.buttons["Xóa"]
        XCTAssertTrue(remove.isHittable)
        attachScreenshot(app, named: "ended-budget-read-only-preview")
        remove.tap()
        XCTAssertTrue(app.alerts.firstMatch.waitForExistence(timeout: 5))
        app.alerts.buttons["Hủy"].tap()
        XCTAssertTrue(preview.waitForExistence(timeout: 5))
        let unchanged = try await LocalAPIFixture.budget(named: name)
        XCTAssertEqual(unchanged.id, created.id)
        let afterLedger = try await LocalAPIFixture.transactions().sorted { $0.id < $1.id }
        XCTAssertEqual(afterLedger, beforeLedger, "Preview/delete cancel must not mutate financial transactions.")
#endif
    }

    func testAPIBackedCategoryCreateUpdateArchiveRestoreDeleteFromNativeUI() async throws {
#if !LOCAL_API_UI
        throw XCTSkip("Build with SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI and run against the disposable local API fixture.")
#else
        let suffix = String(UUID().uuidString.prefix(8))
        let nameValue = "Fixture Category \(suffix)"
        let updatedNameValue = "Fixture Category Updated \(suffix)"
        let scopeWalletA = try await LocalAPIFixture.wallet(name: "Fixture Category Scope A \(suffix)", type: "basic")
        let scopeWalletB = try await LocalAPIFixture.wallet(name: "Fixture Category Scope B \(suffix)", type: "basic")
        addTeardownBlock {
            try? await LocalAPIFixture.delete(scopeWalletA)
            try? await LocalAPIFixture.delete(scopeWalletB)
        }
        try? await LocalAPIFixture.deleteCategories(named: nameValue)
        try? await LocalAPIFixture.deleteCategories(named: updatedNameValue)
        addTeardownBlock {
            try? await LocalAPIFixture.deleteCategories(named: nameValue)
            try? await LocalAPIFixture.deleteCategories(named: updatedNameValue)
        }
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        openMore(app)
        let feedbackBubbleOnMore = app.buttons["floating-feedback-bubble"]
        XCTAssertTrue(feedbackBubbleOnMore.waitForExistence(timeout: 5))
        let visibleMoreRows = app.buttons.matching(
            NSPredicate(format: "identifier BEGINSWITH %@", "more-destination-")
        ).allElementsBoundByIndex.filter { $0.exists && app.frame.intersects($0.frame) }
        XCTAssertFalse(visibleMoreRows.isEmpty)
        XCTAssertNil(
            visibleMoreRows.first { feedbackBubbleOnMore.frame.contains(CGPoint(x: $0.frame.midX, y: $0.frame.midY)) },
            "The feedback bubble must not cover a More row action center."
        )
        let categoriesEntry = app.buttons["Danh mục"]
        XCTAssertTrue(categoriesEntry.waitForExistence(timeout: 10))
        reveal(categoriesEntry, in: app)
        categoriesEntry.tap()
        XCTAssertTrue(app.staticTexts["Danh mục"].firstMatch.waitForExistence(timeout: 10))
        let feedbackBubble = app.buttons["floating-feedback-bubble"]
        XCTAssertTrue(feedbackBubble.waitForExistence(timeout: 5))
        let visibleCategoryRows = app.buttons.matching(
            NSPredicate(format: "identifier BEGINSWITH %@", "category-row-")
        ).allElementsBoundByIndex.filter { $0.exists && app.frame.intersects($0.frame) }
        XCTAssertFalse(visibleCategoryRows.isEmpty, "The category list should expose visible row actions for the overlap check.")
        attachScreenshot(app, named: "category-list-stitch-parity")
        let overlappingCategoryRow = visibleCategoryRows.first { feedbackBubble.frame.contains(CGPoint(x: $0.frame.midX, y: $0.frame.midY)) }
        XCTAssertNil(
            overlappingCategoryRow,
            "The global feedback bubble must not cover a visible category row action. bubble=\(feedbackBubble.frame), row=\(String(describing: overlappingCategoryRow?.frame))"
        )

        let addCategory = app.buttons["category-add-group"]
        XCTAssertTrue(addCategory.waitForExistence(timeout: 10))
        reveal(addCategory, in: app)
        addCategory.tap()
        XCTAssertTrue(app.staticTexts["Nhóm mới"].waitForExistence(timeout: 10), "Create should navigate to a dedicated category editor screen.")
        XCTAssertTrue(app.descendants(matching: .any)["category-editor-screen"].firstMatch.waitForExistence(timeout: 5))
        attachScreenshot(app, named: "category-editor-stitch-parity")
        let walletAOption = app.buttons["category-wallet-option-\(scopeWalletA.wallet.id)"]
        let walletBOption = app.buttons["category-wallet-option-\(scopeWalletB.wallet.id)"]
        XCTAssertTrue(walletAOption.waitForExistence(timeout: 10), "Applicable wallets should be visible as selectable rows, not a menu.")
        XCTAssertTrue(walletBOption.waitForExistence(timeout: 10))
        XCTAssertEqual(walletAOption.value as? String, "Đã chọn", "An empty wallet scope means all wallets are selected.")
        XCTAssertEqual(walletBOption.value as? String, "Đã chọn")
        walletBOption.tap()
        XCTAssertEqual(walletBOption.value as? String, "Chưa chọn")
        // The disposable owner can have other wallets from independent tests.
        // Select exactly A rather than assuming deselecting B leaves one wallet.
        let walletOptionIDs = app.buttons.matching(NSPredicate(format: "identifier BEGINSWITH %@", "category-wallet-option-"))
            .allElementsBoundByIndex.map(\.identifier)
        for optionID in walletOptionIDs where optionID != walletAOption.identifier {
            let option = app.buttons[optionID]
            reveal(option, in: app)
            if option.value as? String == "Đã chọn" { option.tap() }
            XCTAssertEqual(option.value as? String, "Chưa chọn")
        }
        XCTAssertEqual(walletAOption.value as? String, "Đã chọn")
        reveal(walletAOption, in: app)
        attachScreenshot(app, named: "category-editor-wallet-scope-stitch-parity")
        let name = app.textFields["Tên nhóm"]
        XCTAssertTrue(name.waitForExistence(timeout: 10))
        reveal(name, in: app)
        name.clearAndType(nameValue)
        XCTAssertEqual(name.value as? String, nameValue, "Category name must remain exactly the typed fixture before any save")
        guard name.value as? String == nameValue else { attachScreenshot(app, named: "category-name-input-mismatch"); return }
        // The selector is below the keyboard-compressed editor viewport.
        // Dismiss before bounded scrolling so the helper never overscrolls a
        // valid control out of the reduced content area.
        dismissKeyboard(app)
        let iconSelector = app.buttons["category-icon-selector"]
        XCTAssertTrue(iconSelector.waitForExistence(timeout: 10))
        reveal(iconSelector, in: app)
        iconSelector.tap()
        let foodIcon = app.buttons["category-icon-expense_food"]
        XCTAssertTrue(foodIcon.waitForExistence(timeout: 10))
        foodIcon.tap()
        let categorySave = app.buttons["Lưu nhóm"].firstMatch
        dismissKeyboard(app)
        reveal(categorySave, in: app)
        categorySave.tap()
        let categoryEditor = app.descendants(matching: .any)["category-editor-screen"].firstMatch
        let editorClosed = await XCTWaiter.fulfillment(of: [XCTNSPredicateExpectation(predicate: NSPredicate(format: "exists == false"), object: categoryEditor)], timeout: 15)
        XCTAssertEqual(editorClosed, .completed, "Successful save must return from the editor before reading the persisted category")
        guard editorClosed == .completed else { return }

        let createdCategory = try await LocalAPIFixture.category(named: nameValue)
        XCTAssertEqual(createdCategory.iconKey, "expense_food")
        XCTAssertTrue(createdCategory.walletIDs?.contains(scopeWalletA.wallet.id) == true)
        XCTAssertFalse(createdCategory.walletIDs?.contains(scopeWalletB.wallet.id) == true)
        XCTAssertEqual(Set(createdCategory.walletIDs ?? []), Set([scopeWalletA.wallet.id]))
        let categoryRow = app.buttons["category-row-\(createdCategory.id)"]
        reveal(categoryRow, in: app)
        XCTAssertTrue(app.staticTexts[nameValue].waitForExistence(timeout: 10), "Persisted category must render its exact literal name after revealing its row")
        let kindFilter = app.buttons["category-kind-filter"]
        let walletFilter = app.buttons["category-wallet-filter"]
        let searchField = app.textFields["category-search"]
        XCTAssertLessThan(kindFilter.frame.minY, addCategory.frame.minY, "Kind selection should precede the add action, as on web.")
        XCTAssertLessThan(addCategory.frame.minY, walletFilter.frame.minY)
        XCTAssertLessThan(walletFilter.frame.minY, searchField.frame.minY)
        XCTAssertTrue(categoryRow.exists)
        reveal(categoryRow, in: app)
        XCTAssertFalse(app.buttons["category-edit-\(createdCategory.id)"].exists, "Edit is opened by selecting the row, not a trailing pencil.")
        let scopeMatches = categoryRow.label.contains("Hoạt động trong 1 ví")
        XCTAssertTrue(scopeMatches, "Accessible subtitle should reflect the API-backed wallet scope; actual: \(categoryRow.debugDescription)")
        guard scopeMatches else { attachScreenshot(app, named: "category-scope-value-failure"); return }
        let incomeTab = app.buttons["segment-income"]
        XCTAssertTrue(incomeTab.waitForExistence(timeout: 5))
        reveal(incomeTab, in: app)
        incomeTab.tap()
        XCTAssertEqual(XCTWaiter.wait(for: [XCTNSPredicateExpectation(predicate: NSPredicate(format: "exists == false"), object: categoryRow)], timeout: 5), .completed,
                       "The expense category must disappear after the actual income-filter transition")
        let expenseTab = app.buttons["category-kind-filter"]
        XCTAssertTrue(expenseTab.waitForExistence(timeout: 5))
        reveal(expenseTab, in: app)
        expenseTab.tap()
        XCTAssertEqual(XCTWaiter.wait(for: [XCTNSPredicateExpectation(predicate: NSPredicate(format: "isSelected == true"), object: expenseTab)], timeout: 5), .completed)
        let categorySearch = app.textFields["category-search"]
        XCTAssertTrue(categorySearch.waitForExistence(timeout: 5))
        reveal(categorySearch, in: app)
        categorySearch.clearAndType(nameValue)
        dismissKeyboard(app)
        reveal(app.buttons["category-row-\(createdCategory.id)"], in: app)
        XCTAssertTrue(app.buttons["category-row-\(createdCategory.id)"].waitForExistence(timeout: 5))
        XCTAssertTrue(app.buttons["category-wallet-filter"].exists)
        let categoryEdit = app.buttons["category-row-\(createdCategory.id)"]
        XCTAssertTrue(categoryEdit.waitForExistence(timeout: 10))
        reveal(categoryEdit, in: app)
        categoryEdit.tap()
        XCTAssertTrue(app.staticTexts["Sửa nhóm"].waitForExistence(timeout: 10))
        XCTAssertTrue(app.descendants(matching: .any)["category-editor-screen"].firstMatch.waitForExistence(timeout: 5))
        let editedName = app.textFields["Tên nhóm"]
        XCTAssertTrue(editedName.waitForExistence(timeout: 10))
        reveal(editedName, in: app)
        editedName.clearAndType(updatedNameValue)
        dismissKeyboard(app)
        let categoryEditSave = app.buttons["Lưu nhóm"].firstMatch
        reveal(categoryEditSave, in: app)
        categoryEditSave.tap()
        categorySearch.clearAndType(updatedNameValue)
        dismissKeyboard(app)
        reveal(app.buttons["category-row-\(createdCategory.id)"], in: app)
        XCTAssertTrue(app.staticTexts[updatedNameValue].waitForExistence(timeout: 15))

        let categoryEditAgain = app.buttons["category-row-\(createdCategory.id)"]
        XCTAssertTrue(categoryEditAgain.waitForExistence(timeout: 10))
        reveal(categoryEditAgain, in: app)
        categoryEditAgain.tap()
        let archiveCategory = app.buttons["category-active-toggle"]
        XCTAssertTrue(archiveCategory.waitForExistence(timeout: 10))
        XCTAssertEqual(archiveCategory.label, "Lưu trữ nhóm")
        reveal(archiveCategory, in: app)
        archiveCategory.tap()
        XCTAssertTrue(app.staticTexts["category-empty-filter-result"].waitForExistence(timeout: 10), "Archived categories should be hidden by default.")
        let archivedCategory = try await LocalAPIFixture.category(named: updatedNameValue)
        XCTAssertEqual(archivedCategory.isActive, false, "The native archive action should persist inactive state through the API.")
        XCTAssertEqual(Set(archivedCategory.walletIDs ?? []), Set([scopeWalletA.wallet.id]))

        let showInactive = app.buttons["category-show-inactive"]
        XCTAssertTrue(showInactive.waitForExistence(timeout: 10))
        reveal(showInactive, in: app)
        showInactive.tap()
        let archivedRow = app.buttons["category-row-\(createdCategory.id)"]
        XCTAssertTrue(archivedRow.waitForExistence(timeout: 10))
        XCTAssertTrue(archivedRow.label.contains("Đã lưu trữ"), "The identified row button combines its title and archived subtitle in its accessible label.")
        reveal(archivedRow, in: app)
        archivedRow.tap()
        let restoreCategory = app.buttons["category-active-toggle"]
        XCTAssertTrue(restoreCategory.waitForExistence(timeout: 10))
        XCTAssertEqual(restoreCategory.label, "Khôi phục nhóm")
        reveal(restoreCategory, in: app)
        restoreCategory.tap()
        XCTAssertTrue(app.buttons["category-row-\(createdCategory.id)"].waitForExistence(timeout: 10))
        let restoredCategory = try await LocalAPIFixture.category(named: updatedNameValue)
        XCTAssertEqual(restoredCategory.isActive, true, "The native restore action should persist active state through the API.")
        XCTAssertEqual(Set(restoredCategory.walletIDs ?? []), Set([scopeWalletA.wallet.id]))

        let categoryEditForDelete = app.buttons["category-row-\(createdCategory.id)"]
        XCTAssertTrue(categoryEditForDelete.waitForExistence(timeout: 10))
        reveal(categoryEditForDelete, in: app)
        categoryEditForDelete.tap()
        let categoryDelete = app.buttons["category-delete"]
        XCTAssertTrue(categoryDelete.waitForExistence(timeout: 10))
        reveal(categoryDelete, in: app)
        categoryDelete.tap()
        XCTAssertTrue(app.buttons["Xóa"].waitForExistence(timeout: 5))
        app.buttons["Xóa"].tap()
        XCTAssertTrue(app.staticTexts["category-empty-filter-result"].waitForExistence(timeout: 5))
        let categoriesAfterDelete = try await LocalAPIFixture.allCategories()
        XCTAssertFalse(categoriesAfterDelete.contains { $0.id == createdCategory.id }, "Delete must remove the persisted category, not only hide its native row")
#endif
    }

    func testAPIBackedJarCreateUpdateRemoveFromNativeUI() async throws {
#if !LOCAL_API_UI
        throw XCTSkip("Build with SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI and run against the disposable local API fixture.")
#else
        let month = LocalAPIFixture.currentMonth()
        let suffix = String(UUID().uuidString.prefix(8))
        let nameValue = "Fixture Jar \(suffix)"
        let updatedNameValue = "Fixture Jar Updated \(suffix)"
        try? await LocalAPIFixture.deleteJars(named: nameValue, month: month)
        try? await LocalAPIFixture.deleteJars(named: updatedNameValue, month: month)
        addTeardownBlock {
            try? await LocalAPIFixture.deleteJars(named: nameValue, month: month)
            try? await LocalAPIFixture.deleteJars(named: updatedNameValue, month: month)
        }

        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        openMore(app)
        let jarsEntry = app.buttons["Hũ chi tiêu"]
        XCTAssertTrue(jarsEntry.waitForExistence(timeout: 10))
        reveal(jarsEntry, in: app)
        jarsEntry.tap()
        XCTAssertTrue(app.staticTexts["Hũ chi tiêu"].firstMatch.waitForExistence(timeout: 10))

        app.buttons["Thêm hũ"].tap()
        XCTAssertTrue(app.staticTexts["Tạo hũ"].waitForExistence(timeout: 10))
        let name = app.textFields["Tên"]
        XCTAssertTrue(name.waitForExistence(timeout: 10))
        reveal(name, in: app)
        name.clearAndType(nameValue)
        dismissKeyboard(app)
        let save = app.buttons["Lưu"].firstMatch
        XCTAssertTrue(save.waitForExistence(timeout: 10))
        reveal(save, in: app)
        XCTAssertTrue(save.isHittable)
        save.tap()
        let created = app.staticTexts[nameValue]
        XCTAssertTrue(created.waitForExistence(timeout: 15))
        guard created.exists else { return }

        let edit = app.buttons["Sửa tháng này"].firstMatch
        XCTAssertTrue(edit.waitForExistence(timeout: 10))
        reveal(edit, in: app)
        edit.tap()
        XCTAssertTrue(app.staticTexts["Sửa hũ"].waitForExistence(timeout: 10))
        let editedName = app.textFields["Tên"]
        XCTAssertTrue(editedName.waitForExistence(timeout: 10))
        reveal(editedName, in: app)
        editedName.clearAndType(updatedNameValue)
        let editSave = app.buttons["Lưu"].firstMatch
        dismissKeyboard(app)
        app.buttons["screen-close"].tap()
        let unchangedJar = try await LocalAPIFixture.jar(named: nameValue, month: month)
        XCTAssertFalse(unchangedJar.jarID.isEmpty, "Cancel must not write the jar draft.")
        reveal(edit, in: app); edit.tap()
        XCTAssertTrue(app.staticTexts["Sửa hũ"].waitForExistence(timeout: 5))
        XCTAssertEqual(app.textFields["Tên"].value as? String, updatedNameValue, "Reopening the same jar/month must retain the canceled draft.")
        reveal(editSave, in: app)
        XCTAssertTrue(editSave.isHittable, "Jar edit save is not hittable: \(editSave.frame)")
        editSave.tap()
        XCTAssertTrue(app.staticTexts[updatedNameValue].waitForExistence(timeout: 15))

        // A canceled monthly config must not survive a period change, even if
        // the user later returns to the same stable jar identity.
        reveal(edit, in: app); edit.tap()
        app.textFields["Tên"].clearAndType("Discarded month draft")
        dismissKeyboard(app)
        app.buttons["screen-close"].tap()
        let parts = month.split(separator: "-")
        let year = try XCTUnwrap(Int(parts[0]))
        let number = try XCTUnwrap(Int(parts[1]))
        let previousNumber = number == 1 ? 12 : number - 1
        let previousYear = number == 1 ? year - 1 : year
        let monthSelector = app.buttons["form-row-Tháng"]
        reveal(monthSelector, in: app); monthSelector.tap()
        let previousChoice = app.buttons[String(format: "Tháng %02d năm %04d", previousNumber, previousYear)]
        for _ in 0..<4 {
            if previousChoice.exists && previousChoice.isHittable { break }
            app.collectionViews.firstMatch.swipeUp()
        }
        XCTAssertTrue(previousChoice.waitForExistence(timeout: 5))
        previousChoice.tap()
        XCTAssertTrue(monthSelector.label.contains(String(format: "Tháng %02d năm %04d", previousNumber, previousYear)))
        monthSelector.tap()
        let currentChoice = app.buttons[String(format: "Tháng %02d năm %04d", number, year)]
        for _ in 0..<4 {
            if currentChoice.exists && currentChoice.isHittable { break }
            app.collectionViews.firstMatch.swipeUp()
        }
        XCTAssertTrue(currentChoice.waitForExistence(timeout: 5))
        currentChoice.tap()
        XCTAssertTrue(app.staticTexts[updatedNameValue].waitForExistence(timeout: 10))
        reveal(edit, in: app); edit.tap()
        XCTAssertEqual(app.textFields["Tên"].value as? String, updatedNameValue, "Changing jar month must invalidate the canceled edit draft.")
        app.buttons["screen-close"].tap()

        let remove = app.buttons["Gỡ khỏi tháng"].firstMatch
        XCTAssertTrue(remove.waitForExistence(timeout: 10))
        reveal(remove, in: app)
        remove.tap()
        // The stable jar identity remains available in the cumulative selector;
        // removing a month only removes the active month configuration. Assert
        // that the month row and its mutation controls disappear.
        XCTAssertFalse(app.buttons["Sửa tháng này"].waitForExistence(timeout: 10))
#endif
    }

    func testAPIBackedMonthSummaryRecalculatesFromOwnerLedger() async throws {
#if !LOCAL_API_UI
        throw XCTSkip("Build with SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI and run against the disposable local API fixture.")
#else
        let month = LocalAPIFixture.currentMonth()
        let suffix = String(UUID().uuidString.prefix(8))
        let wallet = try await LocalAPIFixture.wallet(name: "Fixture Month Wallet \(suffix)", type: "basic")
        addTeardownBlock {
            try? await LocalAPIFixture.deleteTransactions(walletID: wallet.wallet.id)
            try? await LocalAPIFixture.delete(wallet)
        }

        let before = try await LocalAPIFixture.monthSummary(month: month)
        let amount: Int64 = 321_000
        let category = try await LocalAPIFixture.category(kind: "expense", walletID: wallet.wallet.id)
        _ = try await LocalAPIFixture.transaction(
            walletID: wallet.wallet.id,
            categoryID: category.id,
            amount: amount,
            note: "Fixture month summary \(suffix)"
        )
        let after = try await LocalAPIFixture.monthSummary(month: month)
        XCTAssertEqual(after.expense, before.expense + amount)
        XCTAssertEqual(after.transactionCount, before.transactionCount + 1)

        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        openMore(app)
        let monthsEntry = app.buttons["Tháng"]
        XCTAssertTrue(monthsEntry.waitForExistence(timeout: 10))
        reveal(monthsEntry, in: app)
        monthsEntry.tap()

        let expectedFormatter = NumberFormatter()
        expectedFormatter.numberStyle = .currency
        expectedFormatter.currencyCode = "VND"
        expectedFormatter.currencySymbol = "₫"
        expectedFormatter.maximumFractionDigits = 0
        expectedFormatter.locale = Locale(identifier: "vi_VN")
        let expectedExpense = expectedFormatter.string(from: NSNumber(value: after.expense))!
        let expenseMetric = app.staticTexts.matching(
            NSPredicate(format: "label == %@", "Chi: \(expectedExpense)")
        ).firstMatch
        XCTAssertTrue(
            expenseMetric.waitForExistence(timeout: 15),
            "Month detail must render the updated API expense total \(expectedExpense). UI: \(app.debugDescription)"
        )
        attachScreenshot(app, named: "month-summary-ledger-state")
#endif
    }

    func testAPIBackedMonthNoteSaveAndDeleteFromNativeUI() async throws {
#if !LOCAL_API_UI
        throw XCTSkip("Build with SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI and run against the disposable local API fixture.")
#else
        let month = LocalAPIFixture.currentMonth()
        let noteValue = "UI month note \(UUID().uuidString.prefix(8))"
        try? await LocalAPIFixture.deleteMonthNote(month: month)
        addTeardownBlock { try? await LocalAPIFixture.deleteMonthNote(month: month) }

        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        openMore(app)
        let monthsEntry = app.buttons["Tháng"]
        XCTAssertTrue(monthsEntry.waitForExistence(timeout: 10))
        reveal(monthsEntry, in: app)
        monthsEntry.tap()
        XCTAssertTrue(app.staticTexts["Tháng"].firstMatch.waitForExistence(timeout: 10))

        let note = app.textViews["Ghi chú tháng"]
        XCTAssertTrue(note.waitForExistence(timeout: 15))
        reveal(note, in: app)
        note.tap()
        note.typeText(noteValue)
        XCTAssertTrue(
            (note.value as? String)?.contains(noteValue) == true,
            "Typing into the month note editor must update its bound value before save. UI: \(app.debugDescription)"
        )
        dismissKeyboard(app)
        let save = app.buttons["Lưu ghi chú"]
        reveal(save, in: app)
        XCTAssertTrue(save.isEnabled, "Month-note Save must enable after changing the note.")
        save.tap()

        app.buttons["Quay lại"].tap()
        openMore(app)
        let persistedMonthsEntry = app.buttons["Tháng"]
        reveal(persistedMonthsEntry, in: app)
        persistedMonthsEntry.tap()
        XCTAssertTrue(app.staticTexts["Tháng"].firstMatch.waitForExistence(timeout: 10))
        let persisted = app.textViews["Ghi chú tháng"]
        XCTAssertTrue(persisted.waitForExistence(timeout: 10))
        XCTAssertEqual(persisted.value as? String, noteValue)
        let delete = app.buttons["Xóa ghi chú"]
        reveal(delete, in: app)
        delete.tap()
        XCTAssertFalse(app.buttons["Xóa ghi chú"].waitForExistence(timeout: 5))
#endif
    }

    func testAPIBackedRecurringCreateUpdateDeleteFromNativeUI() async throws {
#if !LOCAL_API_UI
        throw XCTSkip("Build with SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI and run against the disposable local API fixture.")
#else
        let suffix = String(UUID().uuidString.prefix(8))
        let nameValue = "Fixture Recurring \(suffix)"
        let updatedNameValue = "Fixture Recurring Updated \(suffix)"
        let wallet = try await LocalAPIFixture.wallet(name: "Fixture Recurring Wallet \(suffix)", type: "basic")
        let recurringCategory = try await LocalAPIFixture.category(kind: "expense")
        addTeardownBlock {
            try? await LocalAPIFixture.deleteRecurring(named: nameValue)
            try? await LocalAPIFixture.deleteRecurring(named: updatedNameValue)
            try? await LocalAPIFixture.deleteTransactions(walletID: wallet.wallet.id)
            try? await LocalAPIFixture.delete(wallet)
        }

        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        openMore(app)
        let recurringEntry = app.buttons["Giao dịch định kỳ"]
        XCTAssertTrue(recurringEntry.waitForExistence(timeout: 10))
        reveal(recurringEntry, in: app)
        recurringEntry.tap()
        XCTAssertTrue(app.staticTexts["Giao dịch định kỳ"].firstMatch.waitForExistence(timeout: 10))

        let addRecurring = app.buttons["recurring-add"]
        XCTAssertTrue(addRecurring.waitForExistence(timeout: 10))
        XCTAssertEqual(addRecurring.label, "Thêm")
        addRecurring.tap()
        XCTAssertTrue(app.staticTexts["Tạo lịch"].waitForExistence(timeout: 10))
        let name = app.textFields["Tên"]
        XCTAssertTrue(name.waitForExistence(timeout: 10))
        reveal(name, in: app)
        name.clearAndType(nameValue)
        let amount = app.textFields["Số tiền"].firstMatch
        XCTAssertTrue(amount.waitForExistence(timeout: 10))
        // The amount field is below the fold; dismiss the name keyboard before
        // revealing it so XCTest does not dispatch the next edit to the still
        // focused name field while SwiftUI reflows the scroll view.
        dismissKeyboard(app)
        let walletSelector = app.buttons["form-row-Ví"]
        XCTAssertTrue(walletSelector.waitForExistence(timeout: 10))
        reveal(walletSelector, in: app)
        walletSelector.tap()
        let walletChoice = app.buttons[wallet.wallet.name]
        XCTAssertTrue(walletChoice.waitForExistence(timeout: 5), "Recurring editor must expose the fixture wallet as a selectable option.")
        walletChoice.tap()
        XCTAssertTrue(walletSelector.label.contains(wallet.wallet.name), "Recurring schedule must use the fixture wallet selected in the form.")
        reveal(amount, in: app)
        amount.clearAndType("15000")
        if app.buttons["Xong"].exists { app.buttons["Xong"].firstMatch.tap() }
        let categoryRow = app.buttons["recurring-category-row"]
        XCTAssertTrue(categoryRow.waitForExistence(timeout: 5))
        categoryRow.tap()
        let categorySearch = app.textFields["category-picker-search"]
        XCTAssertTrue(categorySearch.waitForExistence(timeout: 5))
        categorySearch.clearAndType(recurringCategory.name)
        let categoryChoice = app.buttons["category-choice-\(recurringCategory.id)"]
        XCTAssertTrue(categoryChoice.waitForExistence(timeout: 5))
        reveal(categoryChoice, in: app)
        categoryChoice.tap()
        let save = app.buttons["Lưu"].firstMatch
        XCTAssertTrue(save.waitForExistence(timeout: 10))
        reveal(save, in: app)
        XCTAssertTrue(save.isHittable)
        save.tap()
        let created = app.staticTexts[nameValue]
        XCTAssertTrue(created.waitForExistence(timeout: 15), "Recurring create did not render after POST: \(nameValue). UI: \(app.debugDescription)")
        guard created.exists else { return }

        let createdRecurring = try await LocalAPIFixture.recurring(named: nameValue)
        XCTAssertEqual(createdRecurring.categoryID, recurringCategory.id)
        XCTAssertEqual(createdRecurring.interval, 1)
        XCTAssertEqual(createdRecurring.walletID, wallet.wallet.id, "Recurring schedule must persist the wallet selected in the editor.")
        let edit = app.buttons["recurring-edit-\(createdRecurring.id)"]
        XCTAssertTrue(edit.waitForExistence(timeout: 10))
        reveal(edit, in: app)
        XCTAssertTrue(edit.isHittable)
        edit.tap()
        XCTAssertTrue(app.staticTexts["Sửa lịch"].waitForExistence(timeout: 10))
        let editedName = app.textFields["Tên"]
        XCTAssertTrue(editedName.waitForExistence(timeout: 10))
        reveal(editedName, in: app)
        editedName.clearAndType(updatedNameValue)
        dismissKeyboard(app)
        let interval = app.textFields["Mỗi bao nhiêu kỳ"]
        reveal(interval, in: app)
        interval.clearAndType("2")
        XCTAssertFalse(app.buttons["amount-calculator-Mỗi bao nhiêu kỳ"].exists, "Recurrence count must not use a money calculator.")
        let editSave = app.buttons["Lưu"].firstMatch
        dismissKeyboard(app)
        app.buttons["screen-close"].tap()
        let unchangedSchedule = try await LocalAPIFixture.recurring(named: nameValue)
        XCTAssertEqual(unchangedSchedule.interval, 1, "Cancel must not write the schedule draft.")
        reveal(edit, in: app); edit.tap()
        XCTAssertTrue(app.staticTexts["Sửa lịch"].waitForExistence(timeout: 5))
        XCTAssertEqual(app.textFields["Tên"].value as? String, updatedNameValue, "Reopening the same schedule must retain the canceled draft.")
        XCTAssertEqual(app.textFields["Mỗi bao nhiêu kỳ"].value as? String, "2")
        reveal(editSave, in: app)
        XCTAssertTrue(editSave.isHittable, "Recurring edit save is not hittable: \(editSave.frame)")
        editSave.tap()
        do {
            let updatedRecurring = try await LocalAPIFixture.recurring(named: updatedNameValue)
            XCTAssertEqual(updatedRecurring.name, updatedNameValue)
            XCTAssertEqual(updatedRecurring.interval, 2)
        } catch {
            XCTFail("Recurring update did not persist through API: \(error). UI: \(app.debugDescription)")
        }
        XCTAssertTrue(
            app.staticTexts[updatedNameValue].waitForExistence(timeout: 15),
            "Updated recurring row did not render after API update. UI: \(app.debugDescription)"
        )

        app.buttons["Chạy các giao dịch đến hạn"].tap()
        XCTAssertTrue(app.staticTexts["Đã tạo 1 giao dịch."].waitForExistence(timeout: 15))
        let generated = try await LocalAPIFixture.transactions().first {
            $0.walletID == wallet.wallet.id && $0.note == "Giao dịch định kỳ — \(updatedNameValue)"
        }
        XCTAssertNotNil(generated, "Run-due must persist the ordinary ledger row using the recurring schedule's default note.")
        XCTAssertEqual(generated?.amount, 15000)
        XCTAssertEqual(generated?.categoryID, recurringCategory.id)
        if let generated {
            app.buttons["bottom-tab-transactions"].tap()
            XCTAssertTrue(
                app.buttons["transaction-row-\(generated.id)"].waitForExistence(timeout: 15),
                "The generated transaction must appear in the native transaction ledger after switching tabs."
            )
        }

        app.buttons["bottom-tab-account"].tap()
        XCTAssertTrue(app.staticTexts["Giao dịch định kỳ"].waitForExistence(timeout: 10))
        let delete = app.buttons["recurring-delete-\(createdRecurring.id)"]
        XCTAssertTrue(delete.waitForExistence(timeout: 10))
        reveal(delete, in: app)
        delete.tap()
        XCTAssertTrue(app.alerts.buttons["Xóa"].waitForExistence(timeout: 5))
        app.alerts.buttons["Xóa"].tap()
        XCTAssertFalse(app.staticTexts[updatedNameValue].waitForExistence(timeout: 10))
#endif
    }

    func testAPIBackedTravelOptionalDatesUseCalendarAndPreserveNulls() async throws {
#if !LOCAL_API_UI
        throw XCTSkip("Requires disposable local API fixture.")
#else
        let name = "Fixture Date Travel \(UUID().uuidString.prefix(8))"
        addTeardownBlock { try? await LocalAPIFixture.deleteTravelEvents(named: name) }
        let timezone = try await LocalAPIFixture.ownerTimezone()
        let formatter = DateFormatter()
        formatter.dateFormat = "yyyy-MM-dd"
        formatter.timeZone = TimeZone(identifier: timezone)
        let expectedDay = formatter.string(from: Date())
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        openMore(app)
        let entry = app.buttons["Travel Mode"]
        reveal(entry, in: app); entry.tap()
        app.buttons["travel-add"].tap()
        let start = app.buttons["optional-date-toggle-travel-start"]
        guard start.waitForExistence(timeout: 5) else { XCTFail("Travel dates must use optional calendar controls, not raw text inputs."); return }
        let field = app.textFields["Tên"]
        field.clearAndType(name)
        dismissKeyboard(app)
        let description = app.textViews.firstMatch
        reveal(description, in: app)
        description.tap()
        description.typeText("Synthetic travel context")
        XCTAssertEqual(description.value as? String, "Synthetic travel context")
        // Do not swipe an outer scroll region through the predictive keyboard.
        // Transfer focus to the native name field's explicit Done accessory.
        field.tap()
        dismissKeyboard(app)
        XCTAssertFalse(app.textFields["Ngày bắt đầu (YYYY-MM-DD)"].exists)
        reveal(start, in: app); start.tap()
        let calendar = app.buttons["date-field-Ngày bắt đầu"]
        reveal(calendar, in: app); calendar.tap()
        XCTAssertTrue(app.descendants(matching: .any)["date-calendar-picker"].waitForExistence(timeout: 5))
        app.buttons["date-calendar-done"].tap()
        app.buttons["Lưu"].tap()
        XCTAssertTrue(app.staticTexts[name].waitForExistence(timeout: 10))
        let created = try await LocalAPIFixture.travelEvent(named: name)
        XCTAssertEqual(created.startsOn, expectedDay)
        XCTAssertNil(created.endsOn)
        XCTAssertEqual(created.context, "Synthetic travel context")
        app.buttons["travel-edit-\(created.id)"].tap()
        let editDescription = app.textViews.firstMatch
        reveal(editDescription, in: app)
        editDescription.tap()
        XCTAssertTrue(app.keyboards.firstMatch.waitForExistence(timeout: 5))
        editDescription.press(forDuration: 1.1)
        let selectAllPredicate = NSPredicate(format: "label ==[c] 'Select All' OR label ==[c] 'Chọn tất cả'")
        let selectAllButton = app.buttons.matching(selectAllPredicate).firstMatch
        let selectAllMenu = app.menuItems.matching(selectAllPredicate).firstMatch
        if selectAllButton.waitForExistence(timeout: 2) {
            selectAllButton.tap()
        } else if selectAllMenu.waitForExistence(timeout: 2) {
            selectAllMenu.tap()
        } else {
            XCTFail("The system selection menu must allow clearing the existing description.")
            return
        }
        editDescription.typeText(XCUIKeyboardKey.delete.rawValue)
        XCTAssertEqual(editDescription.value as? String, "")
        app.textFields["Tên"].tap()
        dismissKeyboard(app)
        let startAgain = app.buttons["optional-date-toggle-travel-start"]
        reveal(startAgain, in: app); startAgain.tap()
        let end = app.buttons["optional-date-toggle-travel-end"]
        reveal(end, in: app); end.tap()
        app.buttons["Lưu"].tap()
        XCTAssertTrue(app.staticTexts[name].waitForExistence(timeout: 10))
        let updated = try await LocalAPIFixture.travelEvent(named: name)
        XCTAssertNil(updated.startsOn)
        XCTAssertEqual(updated.endsOn, expectedDay)
        XCTAssertNil(updated.context)
#endif
    }

    func testAPIBackedTravelCreateToggleUpdateDeleteFromNativeUI() async throws {
#if !LOCAL_API_UI
        throw XCTSkip("Build with SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI and run against the disposable local API fixture.")
#else
        let suffix = String(UUID().uuidString.prefix(8))
        let nameValue = "Fixture Travel \(suffix)"
        let updatedNameValue = "Fixture Travel Updated \(suffix)"
        try? await LocalAPIFixture.deleteTravelEvents(named: nameValue)
        try? await LocalAPIFixture.deleteTravelEvents(named: updatedNameValue)
        addTeardownBlock {
            try? await LocalAPIFixture.deleteTravelEvents(named: nameValue)
            try? await LocalAPIFixture.deleteTravelEvents(named: updatedNameValue)
        }

        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        openMore(app)
        let travelEntry = app.buttons["Travel Mode"]
        XCTAssertTrue(travelEntry.waitForExistence(timeout: 10))
        reveal(travelEntry, in: app)
        travelEntry.tap()
        XCTAssertTrue(app.staticTexts["Travel Mode"].firstMatch.waitForExistence(timeout: 10))

        app.buttons["travel-add"].tap()
        XCTAssertTrue(app.staticTexts["Chuyến đi mới"].waitForExistence(timeout: 10))
        let name = app.textFields["Tên"]
        XCTAssertTrue(name.waitForExistence(timeout: 10))
        reveal(name, in: app)
        name.clearAndType(nameValue)
        dismissKeyboard(app)
        let save = app.buttons["Lưu"].firstMatch
        reveal(save, in: app)
        XCTAssertTrue(save.isHittable)
        save.tap()
        XCTAssertTrue(app.staticTexts[nameValue].waitForExistence(timeout: 15))

        let createdTravel = try await LocalAPIFixture.travelEvent(named: nameValue)
        let enable = app.buttons["travel-toggle-\(createdTravel.id)"]
        XCTAssertTrue(enable.waitForExistence(timeout: 10))
        reveal(enable, in: app)
        enable.tap()
        XCTAssertTrue(app.buttons["Tắt"].waitForExistence(timeout: 15))

        let edit = app.buttons["travel-edit-\(createdTravel.id)"]
        XCTAssertTrue(edit.waitForExistence(timeout: 10))
        reveal(edit, in: app)
        edit.tap()
        XCTAssertTrue(app.staticTexts["Sửa chuyến đi"].waitForExistence(timeout: 10))
        let editedName = app.textFields["Tên"]
        XCTAssertTrue(editedName.waitForExistence(timeout: 10))
        reveal(editedName, in: app)
        editedName.clearAndType(updatedNameValue)
        dismissKeyboard(app)
        let editSave = app.buttons["Lưu"].firstMatch
        app.buttons["screen-close"].tap()
        let unchangedTravel = try await LocalAPIFixture.travelEvent(named: nameValue)
        XCTAssertEqual(unchangedTravel.id, createdTravel.id, "Cancel must not write the travel draft.")
        reveal(edit, in: app); edit.tap()
        XCTAssertTrue(app.staticTexts["Sửa chuyến đi"].waitForExistence(timeout: 5))
        XCTAssertEqual(app.textFields["Tên"].value as? String, updatedNameValue, "Reopening the same trip must retain the canceled draft.")
        reveal(editSave, in: app)
        editSave.tap()
        XCTAssertTrue(app.staticTexts[updatedNameValue].waitForExistence(timeout: 15))

        let delete = app.buttons["travel-delete-\(createdTravel.id)"]
        XCTAssertTrue(delete.waitForExistence(timeout: 10))
        reveal(delete, in: app)
        delete.tap()
        XCTAssertTrue(app.alerts.buttons["Xóa"].waitForExistence(timeout: 5))
        app.alerts.buttons["Xóa"].tap()
        XCTAssertFalse(app.staticTexts[updatedNameValue].waitForExistence(timeout: 10))
#endif
    }

    func testAPIBackedTravelLinkAndUnlinkTransactionFromNativeUI() async throws {
#if !LOCAL_API_UI
        throw XCTSkip("Build with SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI and run against the disposable local API fixture.")
#else
        let suffix = String(UUID().uuidString.prefix(8))
        let walletName = "Fixture Travel Wallet \(suffix)"
        let eventName = "Fixture Travel Link \(suffix)"
        try? await LocalAPIFixture.deleteTravelEvents(named: eventName)
        let wallet = try await LocalAPIFixture.wallet(name: walletName, type: "basic")
        addTeardownBlock {
            try? await LocalAPIFixture.deleteTransactions(walletID: wallet.wallet.id)
            try? await LocalAPIFixture.delete(wallet)
            try? await LocalAPIFixture.deleteTravelEvents(named: eventName)
        }
        let category = try await LocalAPIFixture.category(kind: "expense", walletID: wallet.wallet.id)
        let transaction = try await LocalAPIFixture.transaction(
            walletID: wallet.wallet.id,
            categoryID: category.id,
            amount: 31_000,
            note: "Fixture travel-link transaction"
        )

        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        openMore(app)
        let travelEntry = app.buttons["Travel Mode"]
        XCTAssertTrue(travelEntry.waitForExistence(timeout: 10))
        reveal(travelEntry, in: app)
        travelEntry.tap()
        XCTAssertTrue(app.staticTexts["screen-title-Travel Mode"].waitForExistence(timeout: 10))
        let addTravel = app.buttons["travel-add"]
        XCTAssertTrue(addTravel.waitForExistence(timeout: 10))
        addTravel.tap()
        let name = app.textFields["Tên"]
        XCTAssertTrue(name.waitForExistence(timeout: 10))
        reveal(name, in: app)
        name.clearAndType(eventName)
        dismissKeyboard(app)
        let saveTravel = app.buttons["Lưu"].firstMatch
        reveal(saveTravel, in: app)
        saveTravel.tap()
        XCTAssertTrue(app.staticTexts[eventName].waitForExistence(timeout: 15))

        let backToMore = app.buttons["screen-back"]
        XCTAssertTrue(backToMore.waitForExistence(timeout: 5))
        backToMore.tap()
        openWallets(app)
        XCTAssertTrue(app.staticTexts[walletName].waitForExistence(timeout: 10))
        let walletDetail = app.buttons["wallet-detail-\(wallet.wallet.id)"]
        XCTAssertTrue(walletDetail.waitForExistence(timeout: 10))
        reveal(walletDetail, in: app)
        walletDetail.tap()
        let transactionRow = app.buttons["wallet-transaction-\(transaction.id)"]
        XCTAssertTrue(transactionRow.waitForExistence(timeout: 10))
        transactionRow.tap()

        let travelSelector = app.buttons.matching(
            NSPredicate(format: "label BEGINSWITH %@", "Travel Mode: ")
        ).firstMatch
        XCTAssertTrue(travelSelector.waitForExistence(timeout: 10), app.debugDescription)
        reveal(travelSelector, in: app)
        travelSelector.tap()
        let eventOption = app.buttons[eventName]
        XCTAssertTrue(eventOption.waitForExistence(timeout: 5), app.debugDescription)
        eventOption.tap()
        let saveLink = app.buttons["Lưu liên kết chuyến"]
        reveal(saveLink, in: app)
        saveLink.tap()

        let createdTravelEvent = try await LocalAPIFixture.travelEvent(named: eventName)
        var linked: FixtureTransaction?
        for _ in 0..<20 {
            linked = try await LocalAPIFixture.transactions().first(where: { $0.id == transaction.id })
            if linked?.travelEventID == createdTravelEvent.id { break }
            try await Task.sleep(nanoseconds: 250_000_000)
        }
        XCTAssertEqual(linked?.travelEventID, createdTravelEvent.id)

        let selectorAfterLink = app.buttons.matching(
            NSPredicate(format: "label BEGINSWITH %@", "Travel Mode: ")
        ).firstMatch
        XCTAssertTrue(selectorAfterLink.waitForExistence(timeout: 5))
        selectorAfterLink.tap()
        let unlinkOption = app.buttons["Không gắn chuyến"]
        XCTAssertTrue(unlinkOption.waitForExistence(timeout: 5), app.debugDescription)
        unlinkOption.tap()
        reveal(saveLink, in: app)
        saveLink.tap()

        var unlinked: FixtureTransaction?
        for _ in 0..<20 {
            unlinked = try await LocalAPIFixture.transactions().first(where: { $0.id == transaction.id })
            if unlinked?.travelEventID == nil { break }
            try await Task.sleep(nanoseconds: 250_000_000)
        }
        XCTAssertNil(unlinked?.travelEventID, "Unlinking should persist a null travel_event_id.")
#endif
    }

    func testAPIBackedWalletDetailExposesTransactionEditor() async throws {
#if !LOCAL_API_UI
        throw XCTSkip("Build with SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI and seed the disposable fixture wallet.")
#else
        // Previous interrupted UI runs can leave the named fixture behind;
        // remove all stale copies so the first matching wallet is deterministic.
        try? await LocalAPIFixture.deleteWallets(named: "Fixture Wallet")
        let fixture = try await LocalAPIFixture.wallet(name: "Fixture Wallet", type: "basic")
        addTeardownBlock { try? await LocalAPIFixture.delete(fixture) }
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        openWallets(app)
        XCTAssertTrue(app.staticTexts["Fixture Wallet"].waitForExistence(timeout: 10))
        XCTAssertTrue(app.buttons["Chi tiết"].firstMatch.waitForExistence(timeout: 10))
        app.buttons["Chi tiết"].firstMatch.tap()
        let addTransaction = app.buttons["wallet-detail-add-transaction"]
        XCTAssertTrue(addTransaction.waitForExistence(timeout: 10))
        addTransaction.tap()
        XCTAssertTrue(app.staticTexts["Thêm giao dịch"].waitForExistence(timeout: 10))
        XCTAssertTrue(app.staticTexts["Danh mục"].exists)
        XCTAssertTrue(app.staticTexts["Hũ chi tiêu"].exists)
        XCTAssertTrue(app.buttons["Ngày giao dịch"].exists)
        let details = app.buttons["transaction-details-toggle"]
        XCTAssertTrue(details.waitForExistence(timeout: 10))
        attachScreenshot(app, named: "transaction-editor-api-fixture-parity")
        details.tap()
        XCTAssertTrue(app.staticTexts["Tính vào báo cáo"].exists)
        XCTAssertTrue(app.staticTexts["Ghi chú"].exists)
#endif
    }

    func testAPIBackedGoalWalletDetailRendersSavingsProgress() async throws {
#if !LOCAL_API_UI
        throw XCTSkip("Build with SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI and seed the disposable goal wallet.")
#else
        try? await LocalAPIFixture.deleteWallets(named: "Fixture Goal Wallet")
        let fixture = try await LocalAPIFixture.wallet(name: "Fixture Goal Wallet", type: "goal")
        let category = try await LocalAPIFixture.category(systemKey: "income_transfer_in", walletID: fixture.wallet.id)
        let transaction = try await LocalAPIFixture.transaction(
            walletID: fixture.wallet.id,
            categoryID: category.id,
            amount: 25_000,
            note: "Fixture goal ledger deposit",
            type: "income"
        )
        addTeardownBlock {
            try? await LocalAPIFixture.deleteTransactions(walletID: fixture.wallet.id)
            try? await LocalAPIFixture.delete(fixture)
        }

        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        openWallets(app)
        XCTAssertTrue(app.staticTexts["Fixture Goal Wallet"].waitForExistence(timeout: 10))
        let detail = app.buttons["wallet-detail-\(fixture.wallet.id)"]
        XCTAssertTrue(detail.waitForExistence(timeout: 10), app.debugDescription)
        reveal(detail, in: app)
        detail.tap()
        XCTAssertTrue(app.descendants(matching: .any)["wallet-detail-header"].waitForExistence(timeout: 10))
        XCTAssertTrue(app.descendants(matching: .any)["wallet-detail-goal-progress"].waitForExistence(timeout: 10))
        XCTAssertTrue(app.descendants(matching: .any)["wallet-detail-all-transactions"].waitForExistence(timeout: 10))
        let ledgerRow = app.buttons["wallet-transaction-\(transaction.id)"]
        XCTAssertTrue(ledgerRow.waitForExistence(timeout: 10), "Goal detail should render the persisted ledger row.")
        XCTAssertTrue(ledgerRow.label.contains("25.000"), "The accessible ledger row should show the persisted amount.")
        attachScreenshot(app, named: "goal-wallet-detail-stitch-parity")
#endif
    }

    func testAPIBackedGoalWalletScopedLedgerRendersSavingsPanel() async throws {
#if !LOCAL_API_UI
        throw XCTSkip("Build with SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI and seed the disposable goal wallet.")
#else
        try? await LocalAPIFixture.deleteWallets(named: "Fixture Goal Ledger")
        let fixture = try await LocalAPIFixture.wallet(name: "Fixture Goal Ledger", type: "goal")
        let category = try await LocalAPIFixture.category(systemKey: "income_transfer_in", walletID: fixture.wallet.id)
        _ = try await LocalAPIFixture.transaction(
            walletID: fixture.wallet.id,
            categoryID: category.id,
            amount: 25_000,
            note: "Fixture goal scoped deposit",
            type: "income"
        )
        addTeardownBlock {
            try? await LocalAPIFixture.deleteTransactions(walletID: fixture.wallet.id)
            try? await LocalAPIFixture.delete(fixture)
        }

        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        openWallets(app)
        XCTAssertTrue(app.staticTexts["Fixture Goal Ledger"].waitForExistence(timeout: 10))
        // SwiftUI may compose the descendant identifier into the outer card button
        // when accessibility children are flattened. Match the wallet label as the
        // stable user-facing contract instead of depending on that implementation
        // detail.
        let scope = app.buttons.matching(
            NSPredicate(format: "label CONTAINS[c] %@", fixture.wallet.name)
        ).firstMatch
        XCTAssertTrue(scope.waitForExistence(timeout: 10), app.debugDescription)
        reveal(scope, in: app)
        scope.tap()
        XCTAssertTrue(app.buttons["Ví: Fixture Goal Ledger"].waitForExistence(timeout: 10))
        XCTAssertTrue(app.descendants(matching: .any)["wallet-detail-goal-wallet-selector"].exists)
        XCTAssertTrue(app.descendants(matching: .any)["wallet-detail-goal-progress"].waitForExistence(timeout: 10))
        XCTAssertTrue(app.descendants(matching: .any)["wallet-detail-all-transactions"].waitForExistence(timeout: 10))
        XCTAssertTrue(app.descendants(matching: .any)["wallet-detail-goal-advice"].exists)
        let ledgerRow = app.buttons.matching(
            NSPredicate(format: "label CONTAINS %@", "Fixture goal scoped deposit")
        ).firstMatch
        XCTAssertTrue(ledgerRow.waitForExistence(timeout: 5), "Goal ledger should render its persisted transaction.")
        XCTAssertTrue(ledgerRow.label.contains("25.000"))
        XCTAssertTrue(app.descendants(matching: .any)["wallet-detail-add-transaction"].exists)
        let addTransaction = app.descendants(matching: .any)["wallet-detail-add-transaction"].firstMatch
        let shellTransactions = app.buttons["bottom-tab-transactions"]
        XCTAssertTrue(shellTransactions.waitForExistence(timeout: 10))
        XCTAssertLessThan(
            addTransaction.frame.maxY,
            shellTransactions.frame.minY,
            "Goal ledger CTA must remain above the persistent shell navigation."
        )
        let shellAdd = app.buttons["add-transaction-button"]
        XCTAssertTrue(shellAdd.waitForExistence(timeout: 10))
        XCTAssertLessThan(
            addTransaction.frame.maxY,
            shellAdd.frame.minY,
            "Goal ledger CTA must remain above the shell add affordance."
        )
        let allTransactions = app.descendants(matching: .any)["wallet-detail-all-transactions"].firstMatch
        let savingsProgress = app.descendants(matching: .any)["wallet-detail-goal-progress"].firstMatch
        XCTAssertLessThan(
            allTransactions.frame.minY,
            savingsProgress.frame.minY,
            "Stitch goal ledger places the transaction scope before savings progress."
        )
        XCTAssertFalse(app.descendants(matching: .any)["wallet-detail-header"].exists)
        for _ in 0..<3 where ledgerRow.frame.maxY >= addTransaction.frame.minY {
            app.swipeUp()
        }
        XCTAssertTrue(ledgerRow.isHittable, "The populated goal transaction should scroll into the visible ledger.")
        XCTAssertLessThanOrEqual(
            ledgerRow.frame.maxY,
            addTransaction.frame.minY,
            "The full transaction row must be reachable above the fixed wallet CTA."
        )
        XCTAssertLessThan(
            addTransaction.frame.maxY,
            shellAdd.frame.minY,
            "Scrolling the ledger must not move the fixed wallet CTA under the shell add affordance."
        )
        attachScreenshot(app, named: "goal-wallet-scoped-ledger-stitch-parity")
#endif
    }

    func testAPIBackedWalletSelectionRoutesToScopedLedger() async throws {
#if !LOCAL_API_UI
        throw XCTSkip("Build with SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI and seed the disposable local API fixture.")
#else
        let fixture = try await LocalAPIFixture.wallet(name: "Fixture Scope Wallet", type: "basic")
        addTeardownBlock { try? await LocalAPIFixture.delete(fixture) }

        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        XCTAssertTrue(app.buttons["bottom-tab-account"].waitForExistence(timeout: 10))
        app.buttons["bottom-tab-account"].tap()
        let walletsEntry = app.buttons["more-destination-wallets"]
        XCTAssertTrue(walletsEntry.waitForExistence(timeout: 10))
        reveal(walletsEntry, in: app)
        walletsEntry.tap()
        XCTAssertTrue(app.staticTexts["screen-title-Chọn Ví"].waitForExistence(timeout: 10))
        XCTAssertTrue(app.staticTexts["Fixture Scope Wallet"].waitForExistence(timeout: 10))
        let walletRow = app.buttons.matching(NSPredicate(format: "label CONTAINS[c] %@", "Fixture Scope Wallet")).firstMatch
        XCTAssertTrue(walletRow.waitForExistence(timeout: 10))
        walletRow.tap()
        XCTAssertTrue(app.staticTexts["Giao dịch"].firstMatch.waitForExistence(timeout: 10))
        let walletSelector = app.buttons["transaction-wallet-selector"]
        XCTAssertTrue(walletSelector.waitForExistence(timeout: 10))
        XCTAssertEqual(walletSelector.label, "Ví: Fixture Scope Wallet")
        let periodControl = app.buttons["segment-week"]
        XCTAssertTrue(periodControl.waitForExistence(timeout: 5))
        XCTAssertLessThan(
            walletSelector.frame.minY,
            periodControl.frame.minY,
            "Wallet scope must appear before the period controls, matching the web ledger hierarchy."
        )
        attachScreenshot(app, named: "ordinary-wallet-scope-capsule")
#endif
    }

    func testAPIBackedWalletManagementMatchesGroupedWebContract() async throws {
#if !LOCAL_API_UI
        throw XCTSkip("Build with SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI and seed the disposable local API fixture.")
#else
        let included = try await LocalAPIFixture.wallet(name: "Fixture Included Wallet", type: "basic")
        let excluded = try await LocalAPIFixture.wallet(name: "Fixture Excluded Wallet", type: "basic", isInTotal: false)
        XCTAssertEqual(excluded.wallet.isInTotal, false)
        addTeardownBlock {
            try? await LocalAPIFixture.delete(excluded)
            try? await LocalAPIFixture.delete(included)
        }

        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        openMore(app)
        let walletsEntry = app.buttons["more-destination-wallets"]
        XCTAssertTrue(walletsEntry.waitForExistence(timeout: 10))
        reveal(walletsEntry, in: app)
        walletsEntry.tap()

        XCTAssertTrue(app.staticTexts["screen-title-Chọn Ví"].waitForExistence(timeout: 10), app.debugDescription)
        let totalRow = app.buttons["wallet-total-row"]
        XCTAssertTrue(totalRow.exists)
        XCTAssertTrue(totalRow.label.contains("100.000"), "The total must omit the excluded wallet: \(totalRow.label)")
        XCTAssertTrue(
            app.descendants(matching: .any)["wallet-total-glyph"].exists,
            "The aggregate row needs its distinct globe identity, not the orange basic-wallet glyph."
        )
        XCTAssertTrue(app.staticTexts["TÍNH VÀO TỔNG"].exists)
        XCTAssertTrue(app.descendants(matching: .any)["wallet-group-excluded"].exists, app.debugDescription)
        XCTAssertTrue(app.staticTexts[included.wallet.name].exists)
        XCTAssertTrue(app.staticTexts[excluded.wallet.name].exists)
        let totalTitle = app.staticTexts["wallet-total-title"]
        let totalBalance = app.staticTexts["wallet-total-balance"]
        let totalCheck = app.descendants(matching: .any)["wallet-selection-total-check"]
        XCTAssertTrue(totalTitle.exists)
        XCTAssertTrue(totalBalance.exists)
        XCTAssertTrue(totalCheck.exists, "The aggregate scope is selected when wallet management opens.")
        XCTAssertLessThan(totalTitle.frame.minY, totalBalance.frame.minY, "Stitch stacks aggregate title above its balance.")
        XCTAssertEqual(totalTitle.frame.minX, totalBalance.frame.minX, accuracy: 2, "Aggregate title and balance share a leading edge.")
        let includedName = app.staticTexts["wallet-name-\(included.wallet.id)"]
        let includedBalance = app.staticTexts["wallet-balance-\(included.wallet.id)"]
        XCTAssertTrue(includedName.exists)
        XCTAssertTrue(includedBalance.exists)
        XCTAssertLessThan(includedName.frame.minY, includedBalance.frame.minY, "Stitch stacks wallet name above its balance.")
        XCTAssertEqual(includedName.frame.minX, includedBalance.frame.minX, accuracy: 2, "Wallet name and balance share a leading edge.")
        let includedRow = app.buttons["wallet-scope-\(included.wallet.id)"]
        let includedDetail = app.buttons["wallet-detail-\(included.wallet.id)"]
        XCTAssertTrue(includedDetail.exists, "The native detail extension must remain available.")
        XCTAssertLessThan(
            abs(includedRow.frame.midY - includedDetail.frame.midY),
            8,
            "The native detail extension should stay on the compact wallet row instead of adding a second action row."
        )
        XCTAssertTrue(app.buttons["wallet-add-button"].exists)
        let linkService = app.buttons["wallet-link-service"]
        XCTAssertTrue(linkService.exists)
        XCTAssertFalse(linkService.isEnabled)

        totalRow.tap()
        let totalScope = app.buttons["transaction-wallet-selector"]
        XCTAssertTrue(totalScope.waitForExistence(timeout: 10))
        XCTAssertEqual(totalScope.label, "Ví: Tổng cộng")
        openMore(app)
        let reopenedWallets = app.buttons["more-destination-wallets"]
        XCTAssertTrue(reopenedWallets.waitForExistence(timeout: 10))
        reveal(reopenedWallets, in: app)
        reopenedWallets.tap()
        XCTAssertTrue(app.staticTexts["screen-title-Chọn Ví"].waitForExistence(timeout: 10))
        attachScreenshot(app, named: "wallet-management-grouped-list")

        let editMode = app.buttons["wallet-edit-mode"]
        XCTAssertTrue(editMode.exists)
        editMode.tap()
        XCTAssertTrue(app.buttons["wallet-edit-mode"].label.contains("Xong"))
        app.buttons["wallet-scope-\(included.wallet.id)"].tap()
        XCTAssertTrue(app.descendants(matching: .any)["wallet-editor-header"].waitForExistence(timeout: 10))
        let deleteWallet = app.buttons["wallet-editor-delete"]
        XCTAssertTrue(deleteWallet.exists)
        attachScreenshot(app, named: "wallet-management-editor")
        deleteWallet.tap()
        let confirmation = app.alerts["Xóa ví?"]
        XCTAssertTrue(confirmation.waitForExistence(timeout: 5))
        confirmation.buttons["Xóa"].tap()
        XCTAssertTrue(app.staticTexts["screen-title-Chọn Ví"].waitForExistence(timeout: 10))
        XCTAssertFalse(app.buttons["wallet-scope-\(included.wallet.id)"].waitForExistence(timeout: 5))
#endif
    }

    func testWalletManagementHeaderMatchesStitchCenteredActionAnatomy() throws {
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated"]
        app.launch()
        openMore(app)
        let walletsEntry = app.buttons["more-destination-wallets"]
        XCTAssertTrue(walletsEntry.waitForExistence(timeout: 10))
        reveal(walletsEntry, in: app)
        walletsEntry.tap()

        let title = app.staticTexts["screen-title-Chọn Ví"]
        let close = app.buttons["screen-back"]
        let edit = app.buttons["wallet-edit-mode"]
        XCTAssertTrue(title.waitForExistence(timeout: 10))
        XCTAssertTrue(close.exists)
        XCTAssertTrue(edit.exists)
        XCTAssertLessThan(
            abs(title.frame.midX - app.frame.midX),
            4,
            "Wallet title must be centered independently of the action labels."
        )
        XCTAssertGreaterThan(close.frame.width, 50, "The visible close capsule must exceed an icon-only hit target.")
        XCTAssertGreaterThan(edit.frame.width, 50, "The edit action must use the reference capsule geometry.")
        XCTAssertLessThan(close.frame.maxX, title.frame.minX)
        XCTAssertGreaterThan(edit.frame.minX, title.frame.maxX)
        attachScreenshot(app, named: "wallet-header-stitch-parity")
    }

    func testAPIBackedPortfolioPositionUsesSameCompactMetricScaleAsSummary() async throws {
#if !LOCAL_API_UI
        throw XCTSkip("Requires disposable local API fixture.")
#else
        let symbol = "D\(UUID().uuidString.replacingOccurrences(of: "-", with: "").prefix(8))"
        let fixture = try await LocalAPIFixture.portfolioAsset(symbol: symbol, name: "Compact Position")
        addTeardownBlock { try await LocalAPIFixture.delete(fixture) }
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        openMore(app)
        let entry = app.buttons["more-destination-portfolio"]
        reveal(entry, in: app)
        entry.tap()
        let costs = app.staticTexts.matching(NSPredicate(format: "label BEGINSWITH %@", "Giá vốn: "))
        XCTAssertTrue(costs.firstMatch.waitForExistence(timeout: 10))
        XCTAssertGreaterThanOrEqual(costs.count, 2, "Summary and at least one asset must expose cost values.")
        guard costs.count >= 2 else { return }
        let compactHeight = costs.element(boundBy: 0).frame.height
        let assetCost = costs.element(boundBy: costs.count - 1)
        reveal(assetCost, in: app)
        XCTAssertLessThanOrEqual(assetCost.frame.height, compactHeight + 1,
            "Secondary asset metrics must use the same compact scale, not hero currency typography.")
        let price = app.buttons["portfolio-price-update-\(fixture.asset.id)"]
        XCTAssertTrue(price.exists)
        reveal(price, in: app)
        XCTAssertTrue(price.isHittable)
        XCTAssertGreaterThanOrEqual(price.frame.height, 44)
        attachScreenshot(app, named: "portfolio-compact-position-metrics")
#endif
    }

    func testAPIBackedPortfolioEditorsPreserveDraftOnFailureAndCancelWithoutWrites() async throws {
#if !LOCAL_API_UI
        throw XCTSkip("Requires disposable local API fixture.")
#else
        let symbol = "E\(UUID().uuidString.replacingOccurrences(of: "-", with: "").prefix(8))"
        let fixture = try await LocalAPIFixture.portfolioAsset(symbol: symbol, name: "Fixture Editor")
        addTeardownBlock { try? await LocalAPIFixture.delete(fixture) }
        let before = try await LocalAPIFixture.portfolioAssets().map(\.id).sorted()
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        openMore(app)
        let entry = app.buttons["more-destination-portfolio"]
        reveal(entry, in: app)
        entry.tap()
        XCTAssertTrue(app.staticTexts["screen-title-Danh mục đầu tư"].waitForExistence(timeout: 10))
        XCTAssertFalse(app.textFields["Mã"].exists, "Portfolio list must not contain an asset creation form.")
        XCTAssertFalse(app.textFields["Khối lượng"].exists, "Portfolio list must not contain a trade form.")
        let add = app.buttons["portfolio-open-asset-editor"]
        guard add.waitForExistence(timeout: 5) else { XCTFail("Missing dedicated asset editor entry"); return }
        add.tap()
        let code = app.textFields["Mã"]
        XCTAssertTrue(code.waitForExistence(timeout: 5))
        code.clearAndType(symbol)
        dismissKeyboard(app)
        let name = app.textFields["Tên"]
        name.clearAndType("Duplicate draft retained")
        dismissKeyboard(app)
        app.buttons["portfolio-add-asset"].tap()
        XCTAssertTrue(app.staticTexts["Không thể tạo tài sản."].waitForExistence(timeout: 10))
        XCTAssertEqual(code.value as? String, symbol)
        XCTAssertEqual(name.value as? String, "Duplicate draft retained")
        attachScreenshot(app, named: "portfolio-asset-editor-retryable-error")
        app.buttons["screen-close"].firstMatch.tap()
        XCTAssertTrue(app.staticTexts["screen-title-Danh mục đầu tư"].waitForExistence(timeout: 5))
        let after = try await LocalAPIFixture.portfolioAssets().map(\.id).sorted()
        XCTAssertEqual(after, before)
        add.tap()
        XCTAssertEqual(app.textFields["Mã"].value as? String, symbol, "Cancel must not silently lose the local editor draft.")
        app.textFields["Mã"].clearAndType(symbol + "N")
        dismissKeyboard(app)
        app.buttons["portfolio-add-asset"].tap()
        XCTAssertTrue(app.staticTexts["screen-title-Danh mục đầu tư"].waitForExistence(timeout: 10))
        let createdAssets = try await LocalAPIFixture.portfolioAssets()
        let created = try XCTUnwrap(createdAssets.first(where: { $0.symbol == symbol + "N" }))
        addTeardownBlock { try await LocalAPIFixture.deletePortfolioAsset(id: created.id, token: fixture.token) }
        attachScreenshot(app, named: "portfolio-flat-list-after-editor-save")
        add.tap()
        XCTAssertEqual(app.textFields["Mã"].value as? String, app.textFields["Mã"].placeholderValue, "Successful save resets the create draft.")
        app.buttons["screen-close"].firstMatch.tap()
#endif
    }

    func testAPIBackedPortfolioUnusedAssetCanBeDeletedWithConfirmation() async throws {
#if !LOCAL_API_UI
        throw XCTSkip("Requires disposable local API fixture.")
#else
        let suffix = String(UUID().uuidString.prefix(8))
        let fixture = try await LocalAPIFixture.portfolioAsset(symbol: "X\(suffix)", name: "Fixture Delete Asset \(suffix)")
        addTeardownBlock { try? await LocalAPIFixture.delete(fixture) }

        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        openMore(app)
        let entry = app.buttons["more-destination-portfolio"]
        XCTAssertTrue(entry.waitForExistence(timeout: 10))
        reveal(entry, in: app)
        entry.tap()
        XCTAssertTrue(app.staticTexts["screen-title-Danh mục đầu tư"].waitForExistence(timeout: 10))

        let delete = app.buttons["portfolio-delete-\(fixture.asset.id)"]
        XCTAssertTrue(delete.waitForExistence(timeout: 10), "Unused portfolio assets must expose a delete action.")
        reveal(delete, in: app)
        delete.tap()
        let confirmation = app.alerts["Xóa tài sản?"]
        XCTAssertTrue(confirmation.waitForExistence(timeout: 5), "Deleting an unused asset must require confirmation.")
        confirmation.buttons["Xóa"].tap()

        XCTAssertFalse(delete.waitForExistence(timeout: 10), "The deleted asset must leave the native portfolio list.")
        let remaining = try await LocalAPIFixture.portfolioAssets()
        XCTAssertFalse(remaining.contains(where: { $0.id == fixture.asset.id }), "The owner-scoped API must delete the unused asset.")
#endif
    }

    func testAPIBackedPortfolioAssetCreateAndPriceRefresh() async throws {
#if !LOCAL_API_UI
        throw XCTSkip("Build with SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI and seed the disposable local API fixture.")
#else
        let suffix = UUID().uuidString.replacingOccurrences(of: "-", with: "").prefix(8)
        let symbol = "F\(suffix)"
        let name = "Fixture Portfolio \(suffix)"
        let fixture = try await LocalAPIFixture.portfolioAsset(symbol: symbol, name: name)
        addTeardownBlock { try? await LocalAPIFixture.delete(fixture) }

        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        openMore(app)
        let portfolioEntry = app.buttons["more-destination-portfolio"]
        XCTAssertTrue(portfolioEntry.waitForExistence(timeout: 10))
        reveal(portfolioEntry, in: app)
        portfolioEntry.tap()
        XCTAssertTrue(app.staticTexts["screen-title-Danh mục đầu tư"].waitForExistence(timeout: 10), app.debugDescription)
        // The asset cards and their controls are below the first viewport on a
        // phone-sized simulator. Scroll the real screen before asserting the
        // owner-scoped asset and interacting with its price field.
        app.swipeUp()
        XCTAssertTrue(app.staticTexts["\(symbol) · \(name)"].waitForExistence(timeout: 10))

        let updatePriceButton = app.buttons["portfolio-price-update-\(fixture.asset.id)"]
        XCTAssertTrue(updatePriceButton.waitForExistence(timeout: 5))
        reveal(updatePriceButton, in: app)
        updatePriceButton.tap()
        let price = app.textFields["portfolio-price-\(fixture.asset.id)"]
        XCTAssertTrue(price.waitForExistence(timeout: 10))
        reveal(price, in: app)
        XCTAssertTrue(price.isHittable, "Portfolio price field must be hittable after reveal (frame: \(price.frame), app: \(app.debugDescription))")
        price.clearAndType("120000")
        XCTAssertTrue(app.staticTexts["portfolio-price-editor"].exists, app.debugDescription)
        dismissKeyboard(app)
        let savePrice = app.buttons["portfolio-price-save"]
        reveal(savePrice, in: app)
        XCTAssertTrue(savePrice.isHittable)
        savePrice.tap()
        XCTAssertTrue(app.staticTexts["screen-title-Danh mục đầu tư"].waitForExistence(timeout: 10), app.debugDescription)
        // The card title is present before the request completes, so waiting for
        // that title would make this assertion race the refresh. Wait for the
        // formatted latest-price value rendered from the refreshed API state.
        let refreshedPrice = app.staticTexts.matching(NSPredicate(format: "label CONTAINS[c] %@", "Giá hiện tại: 120.000")).firstMatch
        XCTAssertTrue(refreshedPrice.waitForExistence(timeout: 15))
        let priceReadback = try await LocalAPIFixture.portfolioAssets()
        XCTAssertEqual(priceReadback.first(where: { $0.id == fixture.asset.id })?.latestPrice, 120_000)
        attachScreenshot(app, named: "portfolio-manual-price-persisted")
#endif
    }

    func testAPIBackedPortfolioTradeRefreshesPositionAndHistory() async throws {
#if !LOCAL_API_UI
        throw XCTSkip("Build with SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI and run against the disposable local API fixture.")
#else
        let suffix = UUID().uuidString.replacingOccurrences(of: "-", with: "").prefix(8)
        let symbol = "T\(suffix)"
        let fixture = try await LocalAPIFixture.portfolioAsset(symbol: symbol, name: "Fixture Trade Asset")
        addTeardownBlock { try? await LocalAPIFixture.delete(fixture) }

        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        openMore(app)
        let portfolioEntry = app.buttons["more-destination-portfolio"]
        XCTAssertTrue(portfolioEntry.waitForExistence(timeout: 10))
        reveal(portfolioEntry, in: app)
        portfolioEntry.tap()
        XCTAssertTrue(app.staticTexts["screen-title-Danh mục đầu tư"].waitForExistence(timeout: 10), app.debugDescription)

        let assetSelector = app.buttons["portfolio-trade-asset"]
        let openTrade = app.buttons["portfolio-open-trade-editor"]
        XCTAssertTrue(openTrade.waitForExistence(timeout: 10))
        let ready = XCTNSPredicateExpectation(predicate: NSPredicate(format: "enabled == true"), object: openTrade)
        XCTAssertEqual(XCTWaiter.wait(for: [ready], timeout: 10), .completed)
        openTrade.tap()
        XCTAssertTrue(assetSelector.waitForExistence(timeout: 10))
        reveal(assetSelector, in: app)
        assetSelector.tap()
        let assetOption = app.buttons["\(symbol) · Fixture Trade Asset"]
        XCTAssertTrue(assetOption.waitForExistence(timeout: 5), app.debugDescription)
        assetOption.tap()

        let quantity = app.textFields["Khối lượng"]
        XCTAssertTrue(quantity.waitForExistence(timeout: 5))
        reveal(quantity, in: app)
        quantity.clearAndType("1,25")
        let saveWithKeyboard = app.buttons["portfolio-save-trade"]
        let keyboardTop = app.buttons["Xong"].exists ? app.buttons["Xong"].frame.minY : app.keyboards.firstMatch.frame.minY
        XCTAssertTrue(saveWithKeyboard.exists)
        XCTAssertLessThanOrEqual(saveWithKeyboard.frame.maxY, keyboardTop + 1, "Shared editor footer must remain above keyboard/accessory.")
        attachScreenshot(app, named: "portfolio-trade-editor-keyboard-clearance")
        dismissKeyboard(app)

        let unitPrice = app.textFields["Giá đơn vị"]
        XCTAssertTrue(unitPrice.waitForExistence(timeout: 5))
        reveal(unitPrice, in: app)
        unitPrice.clearAndType("72000")
        dismissKeyboard(app)

        let fee = app.textFields["Phí"]
        XCTAssertTrue(fee.waitForExistence(timeout: 5))
        reveal(fee, in: app)
        fee.clearAndType("2500")
        dismissKeyboard(app)

        let save = app.buttons["portfolio-save-trade"]
        XCTAssertTrue(save.waitForExistence(timeout: 5), app.debugDescription)
        reveal(save, in: app)
        XCTAssertTrue(save.isHittable, "Save trade must be reachable above the shell: \(save.frame)")
        save.tap()

        let position = app.staticTexts["Số lượng: 1.25"]
        XCTAssertTrue(position.waitForExistence(timeout: 15), "Persisted buy quantity must refresh the position: \(app.debugDescription)")
        let history = app.staticTexts["Mua · 1.25"]
        XCTAssertTrue(history.waitForExistence(timeout: 10), "The saved trade must appear in asset history: \(app.debugDescription)")
        XCTAssertFalse(
            app.buttons["portfolio-delete-\(fixture.asset.id)"].exists,
            "A traded asset must remain append-only and must not expose destructive deletion."
        )
        let persistedTrades = try await LocalAPIFixture.portfolioTrades(for: fixture)
        XCTAssertEqual(persistedTrades.count, 1)
        XCTAssertEqual(persistedTrades.first?.assetID, fixture.asset.id)
        XCTAssertEqual(persistedTrades.first?.side, "buy")
        XCTAssertEqual(persistedTrades.first?.quantity, "1.25")
        XCTAssertEqual(persistedTrades.first?.unitPrice, 72_000)
        XCTAssertEqual(persistedTrades.first?.fee, 2_500)
#endif
    }

    func testAPIBackedTransactionCancelDoesNotPersistNoteDraft() async throws {
#if !LOCAL_API_UI
        throw XCTSkip("Requires disposable local API fixture.")
#else
        let before = try await LocalAPIFixture.transactions().map(\.id).sorted()
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        app.buttons["add-transaction-button"].tap()
        let note = app.buttons["transaction-note-row"]
        XCTAssertTrue(note.waitForExistence(timeout: 10))
        note.tap()
        let editor = app.textViews.firstMatch
        XCTAssertTrue(editor.waitForExistence(timeout: 5))
        editor.tap()
        editor.typeText("Synthetic draft must not persist")
        app.buttons["transaction-note-save"].tap()
        XCTAssertTrue(note.waitForExistence(timeout: 5))
        XCTAssertTrue(note.label.contains("Synthetic draft must not persist"))
        app.buttons["screen-close"].firstMatch.tap()
        XCTAssertTrue(app.buttons["bottom-tab-transactions"].waitForExistence(timeout: 5))
        let after = try await LocalAPIFixture.transactions().map(\.id).sorted()
        XCTAssertEqual(after, before, "Note-sheet save and outer cancel must not create a ledger transaction.")
#endif
    }

    func testAPIBackedWalletDetailCreatesAndUpdatesTransaction() async throws {
#if !LOCAL_API_UI
        throw XCTSkip("Build with SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI and seed the disposable fixture wallet.")
#else
        // Keep retries idempotent after a test runner crash or timeout.
        try? await LocalAPIFixture.deleteWallets(named: "Fixture Mutation Wallet")
        let fixture = try await LocalAPIFixture.wallet(name: "Fixture Mutation Wallet", type: "basic")
        addTeardownBlock { try? await LocalAPIFixture.delete(fixture) }
        let fixtureCategory = try await LocalAPIFixture.category(kind: "expense", walletID: fixture.wallet.id)

        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        openWallets(app)
        XCTAssertTrue(app.staticTexts["Fixture Mutation Wallet"].waitForExistence(timeout: 10))
        let fixtureDetail = app.buttons["wallet-detail-\(fixture.wallet.id)"]
        XCTAssertTrue(fixtureDetail.waitForExistence(timeout: 10))
        reveal(fixtureDetail, in: app)
        fixtureDetail.tap()

        let addTransaction = app.buttons["wallet-detail-add-transaction"]
        XCTAssertTrue(addTransaction.waitForExistence(timeout: 10))
        addTransaction.tap()
        XCTAssertTrue(app.staticTexts["Thêm giao dịch"].waitForExistence(timeout: 10))

        let amount = app.textFields["Số tiền"]
        XCTAssertTrue(amount.waitForExistence(timeout: 10))
        reveal(amount, in: app)
        let categorySelector = app.buttons["transaction-category-row"]
        XCTAssertTrue(categorySelector.waitForExistence(timeout: 5))
        categorySelector.tap()
        let categorySearch = app.textFields["category-picker-search"]
        XCTAssertTrue(categorySearch.waitForExistence(timeout: 5))
        categorySearch.clearAndType(fixtureCategory.name)
        let categoryChoice = app.buttons["category-choice-\(fixtureCategory.id)"]
        XCTAssertTrue(categoryChoice.waitForExistence(timeout: 5))
        reveal(categoryChoice, in: app)
        categoryChoice.tap()
        amount.clearAndType("25000")
        dismissKeyboard(app)
        let saveTransaction = app.buttons["Lưu giao dịch"]
        reveal(saveTransaction, in: app)
        saveTransaction.tap()

        // Ordinary rows render the persisted category name when metadata is
        // available; "Khoản chi" is only the uncategorized fallback. Assert
        // the stable persisted amount instead of coupling the test to a label
        // that is intentionally category-specific.
        let createdRow = app.buttons.matching(NSPredicate(format: "identifier BEGINSWITH[c] 'wallet-transaction-' AND label CONTAINS[c] %@", "25.000")).firstMatch
        XCTAssertTrue(createdRow.waitForExistence(timeout: 15), "Created transaction amount should be visible in wallet detail.")
        let persistedTransaction = try await LocalAPIFixture.transactions().first(where: {
            $0.walletID == fixture.wallet.id && $0.amount == 25_000
        })
        XCTAssertEqual(persistedTransaction?.categoryID, fixtureCategory.id)
        createdRow.tap()
        XCTAssertTrue(
            app.descendants(matching: .any)["transaction-detail-shell-inset"].waitForExistence(timeout: 2),
            "Transaction detail overlay must reserve the persistent shell navigation."
        )
        attachScreenshot(app, named: "transaction-detail-shell-clearance")
        let editTransaction = app.buttons["Sửa giao dịch"]
        XCTAssertTrue(editTransaction.waitForExistence(timeout: 10), "Ordinary wallet rows should open a detail action before editing.")
        editTransaction.tap()
        XCTAssertTrue(app.staticTexts["Sửa giao dịch"].waitForExistence(timeout: 10))

        let noteRow = app.buttons["transaction-note-row"]
        XCTAssertTrue(noteRow.waitForExistence(timeout: 5))
        reveal(noteRow, in: app)
        noteRow.tap()
        let editableNote = app.textViews.firstMatch
        XCTAssertTrue(editableNote.waitForExistence(timeout: 10))
        reveal(editableNote, in: app)
        XCTAssertTrue(editableNote.isHittable)
        editableNote.clearAndType("UI Fixture Expense Updated")
        dismissKeyboard(app)
        app.buttons["transaction-note-save"].tap()
        let saveEdit = app.buttons["Lưu thay đổi"]
        reveal(saveEdit, in: app)
        XCTAssertTrue(saveEdit.isHittable, "Wallet transaction save must be hittable after editing note.")
        saveEdit.tap()

        let updatedRow = app.staticTexts.matching(NSPredicate(format: "label CONTAINS[c] %@", "UI Fixture Expense Updated")).firstMatch
        XCTAssertTrue(updatedRow.waitForExistence(timeout: 15), "Updated transaction should be visible after saving.")
        let updatedTransactions = try await LocalAPIFixture.transactions()
        XCTAssertEqual(updatedTransactions.first(where: { $0.id == persistedTransaction?.id })?.note, "UI Fixture Expense Updated")
#endif
    }

    func testAPIBackedCreditWalletOpensLedgerActions() async throws {
#if !LOCAL_API_UI
        throw XCTSkip("Build with SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI and seed the disposable credit wallet.")
#else
        let fixture = try await LocalAPIFixture.wallet(name: "Fixture Credit", type: "credit")
        addTeardownBlock { try? await LocalAPIFixture.delete(fixture) }
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        XCTAssertTrue(app.buttons["Giao dịch"].waitForExistence(timeout: 10))
        app.buttons["Giao dịch"].tap()
        let walletSelector = app.buttons["Ví: Tổng cộng"]
        XCTAssertTrue(walletSelector.waitForExistence(timeout: 10))
        walletSelector.tap()
        XCTAssertTrue(app.buttons["Fixture Credit"].waitForExistence(timeout: 10))
        app.buttons["Fixture Credit"].tap()
        XCTAssertTrue(app.staticTexts["Sổ tín dụng"].waitForExistence(timeout: 10))
        XCTAssertTrue(app.staticTexts["Chưa có giao dịch tín dụng."].waitForExistence(timeout: 10), "An empty credit statement should explain that no credit transactions exist yet.")
        XCTAssertTrue(app.buttons["credit-add-entry"].waitForExistence(timeout: 30))
        XCTAssertTrue(app.buttons["credit-add-payment"].waitForExistence(timeout: 5))
        XCTAssertTrue(app.staticTexts["Hạn mức"].waitForExistence(timeout: 5))

        app.buttons["credit-add-entry"].tap()
        let entrySheetTitle = app.staticTexts["credit-action-sheet"]
        XCTAssertTrue(entrySheetTitle.waitForExistence(timeout: 5), "Adding a credit transaction should open the shared form sheet, not insert a form into the ledger.")
        XCTAssertTrue(app.buttons["screen-close"].exists)
        attachScreenshot(app, named: "credit-entry-shared-form-sheet")
        app.buttons["screen-close"].tap()
        XCTAssertFalse(entrySheetTitle.waitForExistence(timeout: 1), "Canceling the credit form should dismiss it without saving.")
        app.buttons["credit-add-payment"].tap()
        XCTAssertTrue(app.staticTexts["credit-action-sheet"].waitForExistence(timeout: 5), "Paying a credit card should use the same shared form sheet.")
        XCTAssertTrue(app.buttons["screen-close"].exists)
        app.buttons["screen-close"].tap()
#endif
    }

    func testAPIBackedWalletNameDraftStaysLiteralAcrossFocusWithoutSaving() throws {
#if !LOCAL_API_UI
        throw XCTSkip("Requires disposable local API fixture")
#else
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        openWallets(app)
        app.buttons["Thêm ví"].tap()
        defer {
            let cancel = app.buttons["Hủy"]
            if cancel.exists && cancel.isHittable { cancel.tap() }
            else { app.terminate() }
        }
        let name = app.textFields["wallet-editor-name"]
        XCTAssertTrue(name.waitForExistence(timeout: 10))
        name.clearAndType("Fixture UI Wallet")
        XCTAssertEqual(name.value as? String, "Fixture UI Wallet", "Boundary1: direct keyboard entry")
        attachScreenshot(app, named: "wallet-name-before-focus-change")
        let balance = app.textFields["wallet-editor-opening-balance"]
        let frames = XCTAttachment(string: "app=\(app.frame)\nkeyboard=\(app.keyboards.firstMatch.frame)\nscrolls=\(app.scrollViews.allElementsBoundByIndex.map { "\($0.identifier):\($0.frame)" })\nname=\(name.frame)")
        frames.name = "wallet-name-before-reveal-geometry"; frames.lifetime = .keepAlways; add(frames)
        reveal(balance, in: app)
        XCTAssertEqual(name.value as? String, "Fixture UI Wallet", "Boundary1b: reveal/scroll alone must not enter keyboard predictions")
        attachScreenshot(app, named: "wallet-name-after-reveal-before-focus")
        balance.tap()
        XCTAssertEqual(name.value as? String, "Fixture UI Wallet", "Boundary2: focusing another field must not append text")
        attachScreenshot(app, named: "wallet-name-after-focus-change")
        balance.clearAndType("100000")
        let done = app.buttons["Xong"]
        XCTAssertTrue(done.waitForExistence(timeout: 5))
        done.tap()
        XCTAssertEqual(name.value as? String, "Fixture UI Wallet", "Boundary3: closing keyboard must not append text")
        attachScreenshot(app, named: "wallet-name-after-keyboard-dismiss")
#endif
    }

    func testAPIBackedWalletCreateAndUpdateFromNativeEditor() async throws {
#if !LOCAL_API_UI
        throw XCTSkip("Build with SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI and run against the disposable local API fixture.")
#else
        try? await LocalAPIFixture.deleteWallets(named: "Fixture UI Wallet")
        try? await LocalAPIFixture.deleteWallets(named: "Fixture UI Wallet Updated")
        addTeardownBlock {
            try? await LocalAPIFixture.deleteWallets(named: "Fixture UI Wallet")
            try? await LocalAPIFixture.deleteWallets(named: "Fixture UI Wallet Updated")
        }

        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        openWallets(app)
        XCTAssertTrue(app.buttons["Thêm ví"].waitForExistence(timeout: 10))
        app.buttons["Thêm ví"].tap()
        XCTAssertTrue(app.staticTexts["Thêm ví"].waitForExistence(timeout: 10))

        let name = app.textFields["Tên ví"]
        XCTAssertTrue(name.waitForExistence(timeout: 10))
        reveal(name, in: app)
        name.clearAndType("Fixture UI Wallet")
        XCTAssertEqual(name.value as? String, "Fixture UI Wallet", "Wallet name must remain exactly the typed fixture before any save")
        guard name.value as? String == "Fixture UI Wallet" else { attachScreenshot(app, named: "wallet-name-input-mismatch"); return }
        let openingBalance = app.textFields["Số dư ban đầu"]
        XCTAssertTrue(openingBalance.waitForExistence(timeout: 10))
        reveal(openingBalance, in: app)
        openingBalance.clearAndType("100000")
        XCTAssertEqual(openingBalance.value as? String, "100000", "Opening balance input must reach the UIKit-backed binding before save.")
        let openingBalanceDone = app.buttons["Xong"]
        XCTAssertTrue(openingBalanceDone.waitForExistence(timeout: 5))
        openingBalanceDone.tap()
        let create = app.buttons["Tạo ví"]
        XCTAssertTrue(create.waitForExistence(timeout: 10), "Wallet create action is missing: \(app.debugDescription)")
        reveal(create, in: app)
        XCTAssertTrue(create.isHittable, "Wallet create action is not hittable: \(create.frame). UI: \(app.debugDescription)")
        create.tap()
        XCTAssertTrue(app.staticTexts["Fixture UI Wallet"].waitForExistence(timeout: 15))

        let editMode = app.buttons["wallet-edit-mode"]
        XCTAssertTrue(editMode.waitForExistence(timeout: 10))
        reveal(editMode, in: app)
        editMode.tap()
        let createdWallet = try await LocalAPIFixture.wallet(named: "Fixture UI Wallet")
        let walletRow = app.buttons["wallet-scope-\(createdWallet.id)"]
        XCTAssertTrue(walletRow.waitForExistence(timeout: 10))
        walletRow.tap()
        XCTAssertTrue(app.staticTexts["Sửa ví"].waitForExistence(timeout: 10))
        let editedName = app.textFields["Tên ví"]
        XCTAssertTrue(editedName.waitForExistence(timeout: 10))
        reveal(editedName, in: app)
        let currentNameValue = editedName.value as? String ?? "<nil>"
        XCTAssertEqual(currentNameValue, "Fixture UI Wallet")
        editedName.clearAndType("Fixture UI Wallet Updated")
        XCTAssertEqual(editedName.value as? String, "Fixture UI Wallet Updated")
        try await Task.sleep(for: .milliseconds(500))
        let returnKey = app.keyboards.buttons.matching(NSPredicate(format: "label ==[c] 'return' OR label ==[c] 'done'" )).firstMatch
        if returnKey.exists { returnKey.tap() }
        dismissKeyboard(app)
        let save = app.buttons["Lưu thay đổi"]
        XCTAssertTrue(save.isHittable)
        save.tap()
        XCTAssertTrue(app.staticTexts["Fixture UI Wallet Updated"].waitForExistence(timeout: 15))
#endif
    }

    func testAPIBackedTransferCreatesEditsAndDeletesAtomicPairFromNativeUI() async throws {
#if !LOCAL_API_UI
        throw XCTSkip("Build with SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI and run against the disposable local API fixture.")
#else
        let source = try await LocalAPIFixture.wallet(name: "Fixture Transfer Source", type: "basic")
        let destination = try await LocalAPIFixture.wallet(name: "Fixture Transfer Destination", type: "basic")
        addTeardownBlock {
            try? await LocalAPIFixture.deleteTransactions(walletID: source.wallet.id)
            try? await LocalAPIFixture.deleteTransactions(walletID: destination.wallet.id)
            try? await LocalAPIFixture.delete(source)
            try? await LocalAPIFixture.delete(destination)
        }

        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        XCTAssertTrue(app.buttons["Giao dịch"].waitForExistence(timeout: 10))
        app.buttons["Giao dịch"].tap()
        let menu = app.buttons["Tùy chọn giao dịch"]
        XCTAssertTrue(menu.waitForExistence(timeout: 10))
        menu.tap()
        XCTAssertTrue(app.buttons["Chuyển tiền đến ví khác"].waitForExistence(timeout: 10))
        app.buttons["Chuyển tiền đến ví khác"].tap()
        XCTAssertTrue(app.staticTexts["screen-title-Chuyển tiền"].waitForExistence(timeout: 10), "Transfer editor did not open after overflow action. UI: \(app.debugDescription)")

        app.buttons["transfer-source-wallet"].tap()
        app.buttons[source.wallet.name].tap()
        app.buttons["transfer-destination-wallet"].tap()
        app.buttons[destination.wallet.name].tap()

        let amount = app.textFields["Số tiền"]
        XCTAssertTrue(amount.waitForExistence(timeout: 10))
        reveal(amount, in: app)
        amount.clearAndType("12500")
        let amountDone = app.buttons["Xong"]
        if amountDone.exists { amountDone.tap() }
        let note = app.textFields["Ghi chú"]
        XCTAssertTrue(note.waitForExistence(timeout: 10))
        reveal(note, in: app)
        note.clearAndType("UI Transfer")
        dismissKeyboard(app)
        let confirmTransfer = app.buttons["Xác nhận chuyển"]
        XCTAssertTrue(confirmTransfer.waitForExistence(timeout: 10), "Transfer confirmation is missing. UI: \(app.debugDescription)")
        reveal(confirmTransfer, in: app)
        XCTAssertTrue(confirmTransfer.isHittable, "Transfer confirmation is not hittable: \(confirmTransfer.frame). UI: \(app.debugDescription)")
        confirmTransfer.tap()

        let transferRow = app.staticTexts.matching(NSPredicate(format: "label CONTAINS[c] %@", "UI Transfer")).firstMatch
        XCTAssertTrue(transferRow.waitForExistence(timeout: 15), "The atomic transfer pair should be visible after save. UI: \(app.debugDescription)")
        let initialRows = try await LocalAPIFixture.transactions().filter { $0.note == "UI Transfer" }
        XCTAssertEqual(initialRows.count, 2)
        let transferID = try XCTUnwrap(initialRows.compactMap(\.transferID).first)
        XCTAssertEqual(Set(initialRows.map(\.walletID)), Set([source.wallet.id, destination.wallet.id]))
        XCTAssertTrue(initialRows.allSatisfy { $0.includedInReports == false })
        XCTAssertEqual(initialRows.first(where: { $0.type == "expense" })?.walletID, source.wallet.id)
        XCTAssertEqual(initialRows.first(where: { $0.type == "income" })?.walletID, destination.wallet.id)
        let sourceAfterCreate = try await LocalAPIFixture.wallet(named: source.wallet.name)
        let destinationAfterCreate = try await LocalAPIFixture.wallet(named: destination.wallet.name)
        XCTAssertEqual(sourceAfterCreate.currentBalance, 87500)
        XCTAssertEqual(destinationAfterCreate.currentBalance, 112500)

        transferRow.tap()
        XCTAssertTrue(app.staticTexts["Chi tiết giao dịch"].waitForExistence(timeout: 10))
        let editTransfer = app.buttons["Sửa cả cặp chuyển ví"]
        XCTAssertTrue(editTransfer.waitForExistence(timeout: 10))
        editTransfer.tap()
        XCTAssertTrue(app.staticTexts["screen-title-Sửa chuyển tiền"].waitForExistence(timeout: 10))

        let sourceSelector = app.descendants(matching: .any).matching(identifier: "transfer-source-wallet").firstMatch
        let destinationSelector = app.descendants(matching: .any).matching(identifier: "transfer-destination-wallet").firstMatch
        XCTAssertTrue(sourceSelector.waitForExistence(timeout: 5))
        XCTAssertTrue(destinationSelector.waitForExistence(timeout: 5))
        XCTAssertFalse(sourceSelector.isEnabled, "The transfer update API cannot change its source wallet.")
        XCTAssertFalse(destinationSelector.isEnabled, "The transfer update API cannot change its destination wallet.")

        let editedAmount = app.textFields["Số tiền"]
        XCTAssertTrue(editedAmount.waitForExistence(timeout: 10))
        reveal(editedAmount, in: app)
        editedAmount.clearAndType("22500")
        let editedAmountDone = app.buttons["Xong"]
        if editedAmountDone.exists { editedAmountDone.tap() }
        let editedNote = app.textFields["Ghi chú"]
        XCTAssertTrue(editedNote.waitForExistence(timeout: 10))
        reveal(editedNote, in: app)
        editedNote.clearAndType("UI Transfer Updated")
        dismissKeyboard(app)
        let saveTransfer = app.buttons["Lưu thay đổi"]
        XCTAssertTrue(saveTransfer.waitForExistence(timeout: 10))
        reveal(saveTransfer, in: app)
        XCTAssertTrue(saveTransfer.isHittable)
        saveTransfer.tap()

        let updatedTransferRow = app.staticTexts.matching(NSPredicate(format: "label CONTAINS[c] %@", "UI Transfer Updated")).firstMatch
        XCTAssertTrue(updatedTransferRow.waitForExistence(timeout: 15), "The updated pair should return to the refreshed ledger. UI: \(app.debugDescription)")
        let updatedRows = try await LocalAPIFixture.transactions().filter { $0.transferID == transferID }
        XCTAssertEqual(updatedRows.count, 2)
        XCTAssertEqual(Set(updatedRows.map(\.walletID)), Set([source.wallet.id, destination.wallet.id]))
        XCTAssertEqual(Set(updatedRows.map(\.amount)), Set([22_500]))
        XCTAssertEqual(Set(updatedRows.compactMap(\.note)), Set(["UI Transfer Updated"]))
        XCTAssertEqual(Set(updatedRows.compactMap(\.type)), Set(["expense", "income"]))
        XCTAssertTrue(updatedRows.allSatisfy { $0.includedInReports == false })
        let sourceAfterEdit = try await LocalAPIFixture.wallet(named: source.wallet.name)
        let destinationAfterEdit = try await LocalAPIFixture.wallet(named: destination.wallet.name)
        XCTAssertEqual(sourceAfterEdit.currentBalance, 77500)
        XCTAssertEqual(destinationAfterEdit.currentBalance, 122500)

        updatedTransferRow.tap()
        XCTAssertTrue(app.staticTexts["Chi tiết giao dịch"].waitForExistence(timeout: 10))
        XCTAssertTrue(app.buttons["Xóa cả cặp chuyển ví"].waitForExistence(timeout: 10))
        app.buttons["Xóa cả cặp chuyển ví"].tap()
        XCTAssertTrue(app.buttons["Xóa"].waitForExistence(timeout: 5))
        app.buttons["Xóa"].tap()
        XCTAssertFalse(updatedTransferRow.waitForExistence(timeout: 5), "Deleting a transfer must remove both rows from the ledger.")
        let remainingPair = try await LocalAPIFixture.transactions().filter { $0.transferID == transferID }
        XCTAssertTrue(remainingPair.isEmpty, "Server must delete both rows, not just hide the UI")
        let sourceAfterDelete = try await LocalAPIFixture.wallet(named: source.wallet.name)
        let destinationAfterDelete = try await LocalAPIFixture.wallet(named: destination.wallet.name)
        XCTAssertEqual(sourceAfterDelete.currentBalance, 100000)
        XCTAssertEqual(destinationAfterDelete.currentBalance, 100000)
#endif
    }

    func testAPIBackedCreditEntryAndPaymentRefreshStatementFromNativeUI() async throws {
#if !LOCAL_API_UI
        throw XCTSkip("Build with SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI and run against the disposable local API fixture.")
#else
        let suffix = String(UUID().uuidString.prefix(8))
        let source = try await LocalAPIFixture.wallet(name: "Fixture Credit Source \(suffix)", type: "basic")
        let credit = try await LocalAPIFixture.wallet(name: "Fixture Credit Mutation \(suffix)", type: "credit")
        addTeardownBlock {
            try? await LocalAPIFixture.deleteTransferPairs(walletIDs: [source.wallet.id, credit.wallet.id])
            try? await LocalAPIFixture.deleteTransactions(walletID: credit.wallet.id)
            try? await LocalAPIFixture.delete(source)
            try? await LocalAPIFixture.delete(credit)
        }

        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        openWallets(app)
        XCTAssertTrue(app.staticTexts[credit.wallet.name].waitForExistence(timeout: 10))
        let creditRow = app.buttons.matching(NSPredicate(format: "label CONTAINS[c] %@", credit.wallet.name)).firstMatch
        XCTAssertTrue(creditRow.waitForExistence(timeout: 10))
        creditRow.tap()
        XCTAssertTrue(app.staticTexts["Sổ tín dụng"].waitForExistence(timeout: 10))

        let creditAddEntry = app.buttons["credit-add-entry"]
        reveal(creditAddEntry, in: app)
        creditAddEntry.tap()
        XCTAssertTrue(app.staticTexts["Thêm giao dịch tín dụng"].waitForExistence(timeout: 10))
        let entryAmount = app.textFields["Số tiền"]
        XCTAssertTrue(entryAmount.waitForExistence(timeout: 10))
        reveal(entryAmount, in: app)
        entryAmount.clearAndType("12000")
        let entryAmountDone = app.buttons["Xong"]
        XCTAssertTrue(entryAmountDone.waitForExistence(timeout: 5))
        entryAmountDone.tap()
        let note = app.textFields["Ghi chú"]
        XCTAssertTrue(note.waitForExistence(timeout: 10))
        reveal(note, in: app)
        note.clearAndType("Credit category draft")
        dismissKeyboard(app)

        let categoryRow = app.buttons["credit-entry-category"]
        XCTAssertTrue(categoryRow.waitForExistence(timeout: 5))
        reveal(categoryRow, in: app)
        categoryRow.tap()
        XCTAssertTrue(app.textFields["category-picker-search"].waitForExistence(timeout: 5), "Credit entry should use the shared searchable category tree.")
        let choices = app.buttons.matching(NSPredicate(format: "identifier BEGINSWITH %@", "category-choice-"))
        let categoryChoice = (0..<choices.count).map { choices.element(boundBy: $0) }.first(where: \.isEnabled)
        XCTAssertNotNil(categoryChoice, "Credit entry should expose at least one eligible category.")
        let categoryIdentifier = try XCTUnwrap(categoryChoice?.identifier)
        let expectedCategoryID = String(categoryIdentifier.dropFirst("category-choice-".count))
        categoryChoice?.tap()
        XCTAssertTrue(app.buttons["credit-entry-category"].waitForExistence(timeout: 5), "Selecting a category should return to the preserved entry form.")
        XCTAssertEqual(entryAmount.value as? String, "12000", "Choosing a category must preserve the amount draft.")
        XCTAssertEqual(note.value as? String, "Credit category draft", "Choosing a category must preserve the note draft.")
        app.buttons["Lưu"].tap()
        let entryRow = app.staticTexts.matching(NSPredicate(format: "label CONTAINS %@ AND label CONTAINS %@", "Credit category draft", "12.000")).firstMatch
        reveal(entryRow, in: app)
        XCTAssertTrue(entryRow.waitForExistence(timeout: 15), "Credit entry must render the shared category/note/amount row.")
        let persistedStatement = try await LocalAPIFixture.creditStatement(for: credit)
        XCTAssertEqual(persistedStatement.items.first?.categoryID, expectedCategoryID, "Credit entry should persist the category selected from the shared tree.")
        XCTAssertEqual(persistedStatement.items.first?.amount, 12_000)
        XCTAssertEqual(persistedStatement.items.first?.creditKind, "purchase")

        let sourceBefore = try await LocalAPIFixture.wallet(named: source.wallet.name)
        let creditBefore = try await LocalAPIFixture.wallet(named: credit.wallet.name)
        let monthFormatter = DateFormatter()
        monthFormatter.timeZone = TimeZone(identifier: try await LocalAPIFixture.ownerTimezone())
        monthFormatter.dateFormat = "yyyy-MM"
        let month = monthFormatter.string(from: Date())
        let reportBeforePayment = try await LocalAPIFixture.monthSummary(month: month)

        app.buttons["credit-add-payment"].tap()
        XCTAssertTrue(app.staticTexts["Thanh toán thẻ"].waitForExistence(timeout: 10))
        let paymentSource = app.buttons["form-row-Ví thanh toán"]
        XCTAssertTrue(paymentSource.waitForExistence(timeout: 10))
        reveal(paymentSource, in: app)
        paymentSource.tap()
        app.buttons[source.wallet.name].tap()
        XCTAssertTrue(paymentSource.label.contains(source.wallet.name), "Payment must use this test's source, not another wallet's default")
        let paymentAmount = app.textFields["Số tiền"]
        XCTAssertTrue(paymentAmount.waitForExistence(timeout: 10))
        reveal(paymentAmount, in: app)
        paymentAmount.clearAndType("7500")
        let paymentAmountDone = app.buttons["Xong"]
        XCTAssertTrue(paymentAmountDone.waitForExistence(timeout: 5))
        paymentAmountDone.tap()
        // Payment note is also optional; keep this flow focused on the
        // statement mutation and refresh contract.
        app.buttons["Lưu"].tap()
        let paymentRow = app.staticTexts.matching(NSPredicate(format: "label CONTAINS[c] %@", "Thanh toán")).firstMatch
        for _ in 0..<5 where !paymentRow.exists { app.swipeUp() }
        XCTAssertTrue(paymentRow.waitForExistence(timeout: 15), "Credit payment should refresh the statement.")
        let paidStatement = try await LocalAPIFixture.creditStatement(for: credit)
        let payment = try XCTUnwrap(paidStatement.items.first(where: { $0.creditKind == "payment" }))
        XCTAssertEqual(payment.amount, 7500)
        XCTAssertEqual(payment.includedInReports, false)
        XCTAssertNotNil(payment.creditPaymentID)
        XCTAssertEqual(payment.creditPaymentID, payment.transferID)
        let pairID = try XCTUnwrap(payment.transferID)
        let pair = try await LocalAPIFixture.transactions().filter { $0.transferID == pairID }
        XCTAssertEqual(pair.count, 2)
        XCTAssertEqual(Set(pair.map(\.walletID)), Set([source.wallet.id, credit.wallet.id]))
        XCTAssertTrue(pair.allSatisfy { $0.amount == 7500 && $0.includedInReports == false })
        XCTAssertEqual(pair.first(where: { $0.type == "expense" })?.walletID, source.wallet.id)
        XCTAssertEqual(pair.first(where: { $0.type == "income" })?.walletID, credit.wallet.id)
        let sourceAfter = try await LocalAPIFixture.wallet(named: source.wallet.name)
        let creditAfter = try await LocalAPIFixture.wallet(named: credit.wallet.name)
        XCTAssertEqual(sourceAfter.currentBalance, try XCTUnwrap(sourceBefore.currentBalance) - 7500)
        XCTAssertEqual(creditAfter.currentBalance, try XCTUnwrap(creditBefore.currentBalance) + 7500)
        let reportAfterPayment = try await LocalAPIFixture.monthSummary(month: month)
        XCTAssertEqual(reportAfterPayment.income, reportBeforePayment.income)
        XCTAssertEqual(reportAfterPayment.expense, reportBeforePayment.expense)
        XCTAssertEqual(reportAfterPayment.transactionCount, reportBeforePayment.transactionCount,
                       "Paying credit debt must not count the purchase twice in monthly reports")
        attachScreenshot(app, named: "credit-entry-and-payment-shared-rows")
#endif
    }
}
