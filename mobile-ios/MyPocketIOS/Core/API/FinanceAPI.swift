import Foundation

final class FinanceAPI: @unchecked Sendable {
    let client: APIClient

    init(client: APIClient = APIClient()) {
        self.client = client
    }

    private func body<T: Encodable>(_ value: T) throws -> Data {
        try JSONEncoder.mypocket.encode(value)
    }

    func wallets() async throws -> [WalletModel] {
        try await client.send(APIRequest(path: "/wallets"))
    }

    func createWallet(_ input: WalletModelInput) async throws -> WalletModel {
        try await client.send(APIRequest(path: "/wallets", method: .post, body: try body(input), idempotencyKey: UUID().uuidString))
    }

    func updateWallet(id: String, input: WalletModelInput) async throws -> WalletModel {
        try await client.send(APIRequest(path: "/wallets/\(id)", method: .patch, body: try body(input)))
    }

    func deleteWallet(id: String) async throws {
        let _: EmptyResponse = try await client.send(APIRequest(path: "/wallets/\(id)", method: .delete))
    }

    func categories() async throws -> [CategoryModel] {
        try await client.send(APIRequest(path: "/categories"))
    }

    func createCategory(_ input: CategoryModelInput) async throws -> CategoryModel {
        try await client.send(APIRequest(path: "/categories", method: .post, body: try body(input), idempotencyKey: UUID().uuidString))
    }

    func updateCategory(id: String, input: CategoryModelInput) async throws -> CategoryModel {
        try await client.send(APIRequest(path: "/categories/\(id)", method: .patch, body: try body(input)))
    }

    func deleteCategory(id: String) async throws {
        let _: EmptyResponse = try await client.send(APIRequest(path: "/categories/\(id)", method: .delete))
    }

    func transactions() async throws -> [TransactionModel] {
        try await client.send(APIRequest(path: "/transactions"))
    }

    func createTransaction(_ input: TransactionInputModel) async throws -> TransactionModel {
        try await client.send(APIRequest(path: "/transactions", method: .post, body: try body(input), idempotencyKey: UUID().uuidString))
    }

    func updateTransaction(id: String, input: TransactionInputModel) async throws -> TransactionModel {
        try await client.send(APIRequest(path: "/transactions/\(id)", method: .patch, body: try body(input)))
    }

    func updateTransactionTravel(id: String, eventID: String?) async throws -> TransactionModel {
        try await client.send(APIRequest(path: "/transactions/\(id)/travel", method: .patch, body: try body(["event_id": eventID])))
    }

    func deleteTransaction(id: String) async throws {
        let _: EmptyResponse = try await client.send(APIRequest(path: "/transactions/\(id)", method: .delete))
    }

    func createTransfer(_ input: TransferInputModel) async throws -> [TransactionModel] {
        try await client.send(APIRequest(path: "/transactions/transfer", method: .post, body: try body(input), idempotencyKey: UUID().uuidString))
    }

    func updateTransfer(id: String, input: TransferUpdateInputModel) async throws -> [TransactionModel] {
        try await client.send(APIRequest(path: "/transactions/transfer/\(id)", method: .patch, body: try body(input), idempotencyKey: UUID().uuidString))
    }

    func deleteTransfer(id: String) async throws {
        let _: EmptyResponse = try await client.send(APIRequest(path: "/transactions/transfer/\(id)", method: .delete, idempotencyKey: UUID().uuidString))
    }

    func createAdjustment(_ input: AdjustmentInputModel) async throws -> TransactionModel {
        try await client.send(APIRequest(path: "/transactions/adjustment", method: .post, body: try body(input), idempotencyKey: UUID().uuidString))
    }

    func budgets() async throws -> BudgetSummaryModel {
        try await client.send(APIRequest(path: "/budgets"))
    }

    func createBudget(_ input: BudgetInputModel) async throws -> BudgetModel {
        try await client.send(APIRequest(path: "/budgets", method: .post, body: try body(input), idempotencyKey: UUID().uuidString))
    }

    func updateBudget(id: String, input: BudgetInputModel) async throws -> BudgetModel {
        try await client.send(APIRequest(path: "/budgets/\(id)", method: .patch, body: try body(input)))
    }

    func deleteBudget(id: String) async throws {
        let _: EmptyResponse = try await client.send(APIRequest(path: "/budgets/\(id)", method: .delete))
    }

    func monthSummary(month: String) async throws -> MonthSummaryModel {
        try await client.send(APIRequest(path: "/months/\(month)"))
    }

    func saveMonthNote(month: String, note: String) async throws {
        let _: EmptyResponse = try await client.send(APIRequest(path: "/months/\(month)/note", method: .put, body: try body(["note": note])))
    }

    func deleteMonthNote(month: String) async throws {
        let _: EmptyResponse = try await client.send(APIRequest(path: "/months/\(month)/note", method: .delete))
    }

    func insider(month: String, walletID: String? = nil, categoryID: String? = nil) async throws -> InsiderSummaryModel {
        var query = URLComponents()
        var items = [URLQueryItem(name: "month", value: month)]
        if let walletID { items.append(URLQueryItem(name: "wallet_id", value: walletID)) }
        if let categoryID { items.append(URLQueryItem(name: "category_id", value: categoryID)) }
        query.queryItems = items
        return try await client.send(APIRequest(path: "/reports/insider?\(query.percentEncodedQuery ?? "")"))
    }

    func jarMonth(month: String) async throws -> JarMonthSummaryModel {
        try await client.send(APIRequest(path: "/jars?month=\(month)"))
    }

    func createJar(_ input: JarConfigInputModel) async throws -> JarConfigModel {
        try await client.send(APIRequest(path: "/jars", method: .post, body: try body(input), idempotencyKey: UUID().uuidString))
    }

    func updateJarMonth(jarID: String, month: String, input: JarConfigInputModel) async throws -> JarConfigModel {
        try await client.send(APIRequest(path: "/jars/\(jarID)/months/\(month)", method: .put, body: try body(input)))
    }

    func removeJarMonth(jarID: String, month: String) async throws {
        let _: EmptyResponse = try await client.send(APIRequest(path: "/jars/\(jarID)/months/\(month)", method: .delete))
    }

    func jarReport(jarID: String, from: String? = nil, to: String? = nil) async throws -> JarCumulativeSummaryModel {
        var components = URLComponents()
        var items: [URLQueryItem] = []
        if let from { items.append(URLQueryItem(name: "from", value: from)) }
        if let to { items.append(URLQueryItem(name: "to", value: to)) }
        components.queryItems = items
        let suffix = components.percentEncodedQuery.map { "?\($0)" } ?? ""
        return try await client.send(APIRequest(path: "/jars/\(jarID)/report\(suffix)"))
    }

    func recurring() async throws -> [RecurringModel] {
        try await client.send(APIRequest(path: "/recurring"))
    }

    func createRecurring(_ input: RecurringInputModel) async throws -> RecurringModel {
        try await client.send(APIRequest(path: "/recurring", method: .post, body: try body(input), idempotencyKey: UUID().uuidString))
    }

    func updateRecurring(id: String, input: RecurringInputModel) async throws -> RecurringModel {
        try await client.send(APIRequest(path: "/recurring/\(id)", method: .patch, body: try body(input)))
    }

    func deleteRecurring(id: String) async throws {
        let _: EmptyResponse = try await client.send(APIRequest(path: "/recurring/\(id)", method: .delete))
    }

    func runDueRecurring() async throws -> RunDueRecurringResponse {
        try await client.send(APIRequest(path: "/recurring/run-due", method: .post, idempotencyKey: UUID().uuidString))
    }

    func travelEvents() async throws -> [TravelEventModel] {
        try await client.send(APIRequest(path: "/travel/events"))
    }

    func createTravelEvent(_ input: TravelEventInputModel) async throws -> TravelEventModel {
        try await client.send(APIRequest(path: "/travel/events", method: .post, body: try body(input), idempotencyKey: UUID().uuidString))
    }

    func setTravelEventActive(id: String, active: Bool) async throws -> TravelEventModel {
        try await client.send(APIRequest(path: "/travel/events/\(id)/\(active ? "activate" : "deactivate")", method: .post, idempotencyKey: UUID().uuidString))
    }

    func updateTravelEvent(id: String, input: TravelEventInputModel) async throws -> TravelEventModel {
        try await client.send(APIRequest(path: "/travel/events/\(id)", method: .patch, body: try body(input)))
    }

    func deleteTravelEvent(id: String) async throws {
        let _: EmptyResponse = try await client.send(APIRequest(path: "/travel/events/\(id)", method: .delete))
    }

    func portfolioSummary() async throws -> PortfolioSummaryModel {
        try await client.send(APIRequest(path: "/portfolio/summary"))
    }

    func createPortfolioAsset(symbol: String, name: String) async throws -> PortfolioAssetModel {
        try await client.send(APIRequest(path: "/portfolio/assets", method: .post, body: try body(["symbol": symbol, "name": name]), idempotencyKey: UUID().uuidString))
    }

    func updatePortfolioPrice(assetID: String, latestPrice: Int64?) async throws -> PortfolioAssetModel {
        try await client.send(APIRequest(path: "/portfolio/assets/\(assetID)/price", method: .patch, body: try body(["latest_price": latestPrice])))
    }

    func createPortfolioTrade(assetID: String, input: PortfolioTradeInputModel) async throws -> PortfolioTradeModel {
        try await client.send(APIRequest(path: "/portfolio/assets/\(assetID)/trades", method: .post, body: try body(input), idempotencyKey: UUID().uuidString))
    }

    func portfolioTrades(assetID: String) async throws -> [PortfolioTradeModel] {
        try await client.send(APIRequest(path: "/portfolio/assets/\(assetID)/trades"))
    }

    func creditStatement(walletID: String, from: String? = nil, to: String? = nil) async throws -> CreditStatementModel {
        var components = URLComponents()
        var items: [URLQueryItem] = []
        if let from { items.append(URLQueryItem(name: "from", value: from)) }
        if let to { items.append(URLQueryItem(name: "to", value: to)) }
        components.queryItems = items
        let suffix = components.percentEncodedQuery.map { "?\($0)" } ?? ""
        return try await client.send(APIRequest(path: "/credit/wallets/\(walletID)/statement\(suffix)"))
    }

    func createCreditEntry(walletID: String, input: CreditEntryInputModel) async throws -> TransactionModel {
        try await client.send(APIRequest(path: "/credit/wallets/\(walletID)/entries", method: .post, body: try body(input), idempotencyKey: UUID().uuidString))
    }

    func createCreditPayment(walletID: String, input: CreditPaymentInputModel) async throws -> [TransactionModel] {
        try await client.send(APIRequest(path: "/credit/wallets/\(walletID)/payments", method: .post, body: try body(input), idempotencyKey: UUID().uuidString))
    }

    func feedback() async throws -> [FeedbackModel] {
        try await client.send(APIRequest(path: "/feedback"))
    }

    func createFeedback(_ input: FeedbackInputModel) async throws -> FeedbackModel {
        try await client.send(APIRequest(path: "/feedback", method: .post, body: try body(input), idempotencyKey: UUID().uuidString))
    }

    func createFeedback(_ input: FeedbackInputModel, screenshot: Data?) async throws -> FeedbackModel {
        guard let screenshot else { return try await createFeedback(input) }
        return try await client.sendMultipart(
            path: "/feedback",
            fields: ["type": input.type, "title": input.title, "description": input.description],
            files: [MultipartFile(fieldName: "screenshot", fileName: "screen.jpg", mimeType: "image/jpeg", data: screenshot)]
        )
    }

    func changelog() async throws -> [ChangelogModel] {
        try await client.send(APIRequest(path: "/changelog?limit=20", auth: .none))
    }

    func entryCapabilities() async throws -> EntryCapabilitiesModel {
        try await client.send(APIRequest(path: "/ai/entry/capabilities"))
    }

    func entryRequest(id: String) async throws -> EntryProcessModel {
        try await client.send(APIRequest(path: "/ai/entry/requests/\(id)"))
    }

    func processEntry(requestID: String, text: String, timezone: String, files: [MultipartFile] = []) async throws -> EntryProcessModel {
        try await client.sendMultipart(path: "/ai/entry/process", fields: ["request_id": requestID, "text": text, "timezone": timezone], files: files)
    }

    func decideEntryProposal(id: String, version: Int, decision: String) async throws -> EntryProposalModel {
        try await client.send(APIRequest(path: "/ai/entry/proposals/\(id)/\(decision)", method: .post, body: try body(["version": version]), idempotencyKey: UUID().uuidString))
    }

    func advisorCapabilities() async throws -> AdvisorCapabilitiesModel {
        try await client.send(APIRequest(path: "/ai/advisor/capabilities"))
    }

    func advisorMessages(conversationID: String) async throws -> [AdvisorMessageModel] {
        try await client.send(APIRequest(path: "/ai/advisor/conversation/messages?conversation_id=\(conversationID)"))
    }

    func submitAdvisorMessage(_ input: AdvisorMessageInputModel) async throws -> AdvisorSubmitModel {
        try await client.send(APIRequest(path: "/ai/advisor/messages", method: .post, body: try body(input), idempotencyKey: UUID().uuidString))
    }

    func editEntryProposal(id: String, version: Int, draft: EntryDraftModel) async throws -> EntryProposalModel {
        try await client.send(APIRequest(path: "/ai/entry/proposals/\(id)", method: .patch, body: try body(EntryProposalEditBody(version: version, draft: draft))))
    }

    func profile() async throws -> ProfileModel {
        try await client.send(APIRequest(path: "/auth/profile"))
    }

    func updateTimezone(_ timezone: String) async throws -> ProfileModel {
        try await client.send(APIRequest(path: "/auth/profile", method: .patch, body: try body(["timezone": timezone])))
    }

    func apiKeys() async throws -> [APIKeyModel] {
        try await client.send(APIRequest(path: "/api-keys"))
    }

    func createAPIKey(name: String, scopes: [String], expiresAt: String? = nil) async throws -> APIKeyCreatedModel {
        try await client.send(APIRequest(path: "/api-keys", method: .post, body: try body(APIKeyCreateBody(name: name, scopes: scopes, expiresAt: expiresAt)), idempotencyKey: UUID().uuidString))
    }

    func revokeAPIKey(id: String) async throws {
        let _: EmptyResponse = try await client.send(APIRequest(path: "/api-keys/\(id)", method: .delete))
    }
}

struct EmptyResponse: Decodable, Sendable {}

private struct EntryProposalEditBody: Encodable, Sendable {
    var version: Int
    var draft: EntryDraftModel
}

private struct APIKeyCreateBody: Encodable, Sendable {
    var name: String
    var scopes: [String]
    var expiresAt: String?
}

struct WalletModelInput: Codable, Sendable {
    var name: String
    var type: String
    var openingBalance: Int64
    var isInTotal: Bool
    var description: String?
    var targetAmount: Int64?
    var targetDate: String?
    var creditLimit: Int64?
    var lastStatementBalance: Int64?
    var statementDay: Int?
    var paymentDueDay: Int?
}

struct CategoryModelInput: Codable, Sendable {
    var name: String
    var kind: String
    var parentID: String?
    var walletIDs: [String]
    var iconKey: String
}

struct RecurringInputModel: Codable, Sendable {
    var name: String
    var walletID: String
    var categoryID: String?
    var type: String
    var amount: Int64
    var note: String?
    var frequency: String
    var interval: Int
    var nextRunAt: String
    var endsAt: String?
    var active: Bool
}

struct RunDueRecurringResponse: Codable, Sendable {
    var created: Int
}

struct TravelEventInputModel: Codable, Sendable {
    var name: String
    var context: String?
    var startsOn: String?
    var endsOn: String?
    var active: Bool
}

struct PortfolioTradeInputModel: Codable, Sendable {
    var side: String
    var quantity: String
    var unitPrice: Int64
    var fee: Int64
    var occurredAt: String
    var note: String?
}

struct PortfolioTradeModel: Codable, Identifiable, Sendable {
    let id: String
    var assetID: String
    var side: String
    var quantity: String
    var unitPrice: Int64
    var fee: Int64
    var occurredAt: String
    var note: String?
}

struct CreditEntryInputModel: Codable, Sendable {
    var kind: String
    var categoryID: String?
    var amount: Int64
    var occurredAt: String?
    var note: String?
}

struct CreditPaymentInputModel: Codable, Sendable {
    var sourceWalletID: String
    var amount: Int64
    var occurredAt: String?
    var note: String?
}

struct AdvisorMessageInputModel: Codable, Sendable {
    var clientRequestID: String
    var text: String
}
