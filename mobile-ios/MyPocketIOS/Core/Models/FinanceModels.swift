import Foundation

struct OfflineSyncCapabilitiesModel: Codable, Equatable, Sendable {
    let contractVersion: Int
    let enabled: Bool
    let operations: [String]
    let receiptMetadata: String
    let idempotencyKeyRequiredForOffline: Bool
    let legacyUnkeyedOnline: Bool
    var transactionCreateContext: String? = nil

    func supports(operation: String) -> Bool {
        guard enabled, contractVersion == 1, idempotencyKeyRequiredForOffline,
              receiptMetadata == "/meta/mutation",
              operation == "transaction.create" || operation == "transfer.create" else { return false }
        return operations.contains(operation) && (operation != "transaction.create" || transactionCreateContext == "explicit_travel_v1")
    }
}

struct WalletModel: Codable, Identifiable, Equatable, Sendable {
    let id: String
    var name: String
    var type: String
    var currency: String
    var openingBalance: Int64
    var currentBalance: Int64
    var availableCredit: Int64?
    var isInTotal: Bool
    var description: String?
    var targetAmount: Int64?
    var targetDate: String?
    var creditLimit: Int64?
    var lastStatementBalance: Int64?
    var statementDay: Int?
    var paymentDueDay: Int?
    var goalNotificationsEnabled: Bool = false
}

struct NotificationModel: Codable, Identifiable, Equatable, Sendable {
    let id: String
    let kind: String
    let title: String
    let body: String
    let readAt: String?
    let createdAt: String
}

struct CategoryModel: Codable, Identifiable, Equatable, Sendable {
    let id: String
    var ownerID: String?
    var parentID: String?
    var kind: String
    var name: String
    var systemKey: String?
    var iconKey: String
    var isSystem: Bool
    var isActive: Bool? = true
    var walletIDs: [String]
}

enum CategoryEditPolicy {
    static func metadataIsReadOnly(_ category: CategoryModel) -> Bool { category.isSystem }
    static func canDelete(_ category: CategoryModel) -> Bool { !category.isSystem }
    static func canArchive(_ category: CategoryModel) -> Bool { !category.isSystem }
    static func canSelect(_ category: CategoryModel) -> Bool { category.isActive != false }
}

struct CategoryListNode: Identifiable, Equatable {
    let item: CategoryModel
    let depth: Int
    let isSelectable: Bool
    let isHierarchyContext: Bool
    var continuesBranch: Bool

    var id: String { item.id }
}

struct CategorySelectionGroup: Identifiable, Equatable {
    let root: CategoryListNode
    var children: [CategoryListNode]

    var id: String { root.id }
    var nodes: [CategoryListNode] { [root] + children }
}

enum CategoryListPresentation {
    static func nodes(
        items: [CategoryModel],
        kind: String,
        walletID: String?,
        search: String,
        selectableIDs: Set<String>? = nil,
        includeInactive: Bool = false
    ) -> [CategoryListNode] {
        let candidates = items.filter { $0.kind == kind }
        let knownIDs = Set(candidates.map(\.id))
        let childPairs = candidates.compactMap { item -> (String, CategoryModel)? in
            guard let parentID = item.parentID, knownIDs.contains(parentID) else { return nil }
            return (parentID, item)
        }
        let childMap = Dictionary(grouping: childPairs, by: { $0.0 }).mapValues { $0.map { $0.1 } }
        let roots = candidates.filter { $0.parentID == nil || !knownIDs.contains($0.parentID ?? "") }
        let normalizedQuery = normalize(search)

        func matchesSearch(_ item: CategoryModel) -> Bool {
            normalizedQuery.isEmpty || normalize(item.name).contains(normalizedQuery)
        }

        func appliesToWallet(_ item: CategoryModel) -> Bool {
            guard let walletID, !walletID.isEmpty else { return true }
            return item.walletIDs.isEmpty || item.walletIDs.contains(walletID)
        }

        func build(_ item: CategoryModel, depth: Int, inheritedSearchMatch: Bool) -> CategoryListNode? {
            let itemMatches = matchesSearch(item)
            let showAllDescendants = inheritedSearchMatch || (normalizedQuery.isEmpty || itemMatches)
            let children = (childMap[item.id] ?? []).compactMap {
                build($0, depth: depth + 1, inheritedSearchMatch: showAllDescendants)
            }
            let statusEligible = includeInactive || CategoryEditPolicy.canSelect(item)
            let eligible = appliesToWallet(item) && statusEligible && (selectableIDs?.contains(item.id) ?? true)
            guard (eligible && showAllDescendants) || !children.isEmpty else { return nil }
            return CategoryListNode(
                item: item,
                depth: depth,
                isSelectable: eligible,
                isHierarchyContext: !eligible || (!normalizedQuery.isEmpty && !itemMatches && !inheritedSearchMatch),
                continuesBranch: false
            )
        }

        func appendChildren(_ nodes: [CategoryListNode], into output: inout [CategoryListNode]) {
            for (index, node) in nodes.enumerated() {
                var node = node
                node.continuesBranch = index < nodes.count - 1
                output.append(node)
                let childNodes = (childMap[node.item.id] ?? []).compactMap {
                    build($0, depth: node.depth + 1, inheritedSearchMatch: normalizedQuery.isEmpty || matchesSearch(node.item))
                }
                appendChildren(childNodes, into: &output)
            }
        }

        var output: [CategoryListNode] = []
        appendChildren(roots.compactMap { build($0, depth: 0, inheritedSearchMatch: false) }, into: &output)
        return output
    }

    private static func normalize(_ value: String) -> String {
        value.folding(options: [.caseInsensitive, .diacriticInsensitive], locale: Locale(identifier: "vi_VN"))
            .trimmingCharacters(in: .whitespacesAndNewlines)
    }
}

enum CategorySelectionPresentation {
    static func subtitle(isSystem: Bool, isSelectable: Bool) -> String {
        guard isSelectable else { return "Không áp dụng cho lựa chọn này" }
        return isSystem ? "Nhóm mặc định" : "Nhóm cá nhân"
    }

    static func nodes(
        items: [CategoryModel],
        kind: String,
        selectableIDs: Set<String>? = nil,
        search: String
    ) -> [CategoryListNode] {
        CategoryListPresentation.nodes(
            items: items,
            kind: kind,
            walletID: nil,
            search: search,
            selectableIDs: selectableIDs
        )
    }

    static func groups(
        items: [CategoryModel],
        kind: String,
        selectableIDs: Set<String>? = nil,
        search: String
    ) -> [CategorySelectionGroup] {
        let nodes = nodes(items: items, kind: kind, selectableIDs: selectableIDs, search: search)
        var groups: [CategorySelectionGroup] = []

        for node in nodes {
            if node.depth == 0 {
                groups.append(CategorySelectionGroup(root: node, children: []))
            } else if !groups.isEmpty {
                groups[groups.count - 1].children.append(node)
            }
        }

        return groups
    }
}

extension CategoryModel {
    func applies(to transactionType: String, walletID: String, walletType: String) -> Bool {
        isActive != false && appliesIgnoringActivity(to: transactionType, walletID: walletID, walletType: walletType)
    }

    func appliesIgnoringActivity(to transactionType: String, walletID: String, walletType: String) -> Bool {
        let savingsTransfer = transactionType == "income"
            ? systemKey == "income_transfer_in" || systemKey == "income_interest"
            : systemKey == "expense_transfer_out"
        return kind == transactionType
            && (walletType != "goal" || savingsTransfer)
            && (walletIDs.isEmpty || walletIDs.contains(walletID))
    }
}

struct TransactionModel: Codable, Identifiable, Equatable, Sendable {
    let id: String
    let ownerID: String
    var walletID: String
    var categoryID: String?
    var jarID: String?
    var jarName: String?
    var type: String
    var amount: Int64
    var adjustmentDirection: String?
    var creditKind: String?
    var creditPaymentID: String?
    var travelEventID: String?
    var occurredAt: String
    var note: String?
    var includedInReports: Bool
    var transferID: String?
    var createdAt: String
    var updatedAt: String
}

struct TransactionInputModel: Codable, Sendable {
    var walletID: String
    var categoryID: String?
    var jarID: String?
    var type: String
    var amount: Int64
    var occurredAt: String
    var note: String?
    var includedInReports: Bool
    var travelEventID: String?
    var autoLinkTravel: Bool? = nil
}

struct BulkDeleteInputModel: Codable, Sendable {
    var transactionIDs: [String]

    enum CodingKeys: String, CodingKey {
        case transactionIDs = "transaction_ids"
    }
}

struct TransferInputModel: Codable, Sendable {
    var sourceWalletID: String
    var destinationWalletID: String
    var amount: Int64
    var occurredAt: String
    var note: String?
}

struct TransferUpdateInputModel: Codable, Sendable {
    var amount: Int64
    var occurredAt: String
    var note: String?
}

struct AdjustmentInputModel: Codable, Sendable {
    var walletID: String
    var amount: Int64
    var direction: String
    var occurredAt: String?
    var note: String?
}

struct BudgetModel: Codable, Identifiable, Equatable, Sendable {
    let id: String
    var name: String
    var limitAmount: Int64
    var walletID: String?
    var categoryID: String?
    var startDate: String
    var endDate: String
    var spent: Int64
    var daysRemaining: Int
    var ended: Bool
    var seriesID: String? = nil
    var baseLimitAmount: Int64? = nil
    var carryoverAmount: Int64? = nil
    var effectiveLimitAmount: Int64? = nil

    var displayLimitAmount: Int64 { effectiveLimitAmount ?? limitAmount }
    var isSeriesManaged: Bool { seriesID?.isEmpty == false }
}

enum BudgetCadence: String, Codable, CaseIterable, Sendable {
    case weekly, monthly, quarterly, yearly

    var label: String {
        switch self {
        case .weekly: "Hàng tuần"
        case .monthly: "Hàng tháng"
        case .quarterly: "Hàng quý"
        case .yearly: "Hàng năm"
        }
    }

    func canonicalStart(_ date: Date, timeZone: TimeZone) -> String {
        var calendar = Calendar(identifier: .gregorian)
        calendar.timeZone = timeZone
        var components = calendar.dateComponents([.year, .month, .day], from: date)
        switch self {
        case .weekly:
            let weekday = calendar.component(.weekday, from: date)
            let monday = calendar.date(byAdding: .day, value: -((weekday + 5) % 7), to: date) ?? date
            return MPFormat.calendarDateString(monday, timeZone: timeZone)
        case .monthly: components.day = 1
        case .quarterly: components.month = (((components.month ?? 1) - 1) / 3) * 3 + 1; components.day = 1
        case .yearly: components.month = 1; components.day = 1
        }
        return MPFormat.calendarDateString(calendar.date(from: components) ?? date, timeZone: timeZone)
    }
}

struct SavingsReportModel: Decodable, Sendable {
    let month: String
    let timezone: String
    let asOf: String
    let goals: SavingsGoalsModel
    let surplus: SavingsSurplusModel
}

struct SavingsGoalsModel: Decodable, Sendable {
    let openingBalance: Int64
    let closingBalance: Int64
    let deposits: Int64
    let withdrawals: Int64
    let adjustments: Int64
    let netChange: Int64
    let internalTransfers: Int64
    let initialBalanceAdded: Int64?
    let items: [SavingsGoalItemModel]
}

struct SavingsGoalItemModel: Decodable, Identifiable, Sendable {
    var id: String { walletID }
    let walletID: String
    let name: String
    let openingBalance: Int64
    let closingBalance: Int64
    let deposits: Int64
    let withdrawals: Int64
    let adjustments: Int64
    let netChange: Int64
    let internalTransfers: Int64
    let initialBalanceAdded: Int64?
    let targetAmount: Int64?
    let targetDate: String?
    let remaining: Int64
    let progressBPS: Int64?
    let reached: Bool
    let transactionCount: Int
}

struct SavingsSurplusModel: Decodable, Sendable {
    let scope: String
    let income: Int64
    let expense: Int64
    let net: Int64
    let savingsRateBPS: Int64?
    let transactionCount: Int
}

struct ReportTransactionPageModel: Decodable, Sendable {
    let items: [TransactionModel]
    let total: Int
    let limit: Int
    let offset: Int
    let month: String?
    let timezone: String?
    let asOf: String?
}

struct BudgetReportModel: Decodable, Sendable {
    let budgetID: String
    let timezone: String
    let asOf: String
    let startDate: String
    let endDate: String
    let baseLimitAmount: Int64
    let carryoverAmount: Int64
    let effectiveLimitAmount: Int64
    let spent: Int64
    let remaining: Int64
    let forecast: BudgetForecastModel
    let daily: [BudgetDailyModel]
}

struct BudgetForecastModel: Decodable, Sendable {
    let status: String
    let method: String
    let elapsedDays: Int
    let totalDays: Int
    let projectedSpent: Int64?
    let projectedRemaining: Int64?
    let overBudget: Bool?

    var statusLabel: String {
        switch status {
        case "estimated": "Ước tính theo nhịp chi hiện tại"
        case "actual": "Đã kết thúc · Chi tiêu thực tế"
        case "not_started": "Chưa bắt đầu · Chưa có dự báo"
        default: "Dự báo chưa khả dụng"
        }
    }
}

struct BudgetDailyModel: Decodable, Identifiable, Sendable {
    var id: String { date }
    let date: String
    let spent: Int64
}

struct BudgetSeriesListModel: Decodable, Sendable {
    let items: [BudgetSeriesModel]
}

struct BudgetSeriesModel: Decodable, Identifiable, Sendable {
    let id: String
    let ownerID: String
    let cadence: BudgetCadence
    let startDate: String
    let active: Bool
    let version: Int64
    let generationError: String?
    let currentConfig: BudgetSeriesConfigModel
    let pendingConfig: BudgetSeriesConfigModel?

    var editableConfig: BudgetSeriesConfigModel { pendingConfig ?? currentConfig }
}

struct BudgetSeriesConfigModel: Decodable, Sendable {
    let effectiveStartDate: String
    let name: String
    let limitAmount: Int64
    let walletID: String?
    let categoryID: String?
    let carryoverEnabled: Bool
}

struct BudgetSeriesCreateInput: Encodable, Sendable {
    let name: String
    let limitAmount: Int64
    let walletID: String?
    let categoryID: String?
    let cadence: BudgetCadence
    let startDate: String
    let carryoverEnabled: Bool
}

struct BudgetSeriesUpdateInput: Encodable, Sendable {
    let version: Int64
    let name: String
    let limitAmount: Int64
    let walletID: String?
    let categoryID: String?
    let carryoverEnabled: Bool
}

struct BudgetSummaryModel: Codable, Sendable {
    var items: [BudgetModel]
    var limitAmount: Int64
    var spent: Int64
}

struct BudgetInputModel: Codable, Sendable {
    var name: String
    var limitAmount: Int64
    var walletID: String?
    var categoryID: String?
    var startDate: String
    var endDate: String
}

struct MonthCategoryTotalModel: Codable, Sendable {
    var stableID: String { "\(type):\(categoryID ?? name)" }
    var categoryID: String?
    var name: String
    var type: String
    var amount: Int64
    var count: Int
}

struct MonthSummaryModel: Codable, Sendable {
    var month: String
    var timezone: String
    var startAt: String
    var nextStartAt: String
    var isCurrent: Bool
    var isComplete: Bool
    var income: Int64
    var expense: Int64
    var net: Int64
    var transactionCount: Int
    var categories: [MonthCategoryTotalModel]
    var note: String
    var jarSummary: JarMonthSummaryModel?
    var calculatedAt: String
}

struct InsiderCategoryTotalModel: Codable, Sendable {
    var categoryID: String?
    var name: String
    var amount: Int64
    var count: Int
}

struct InsiderExpenseRowModel: Codable, Identifiable, Sendable {
    let id: String
    var walletID: String
    var walletName: String?
    var categoryID: String?
    var categoryName: String?
    var amount: Int64
    var occurredAt: String
    var note: String?
}

struct InsiderSummaryModel: Codable, Sendable {
    var month: String
    var timezone: String
    var startAt: String
    var nextStartAt: String
    var isCurrent: Bool
    var income: Int64
    var expense: Int64
    var net: Int64
    var previousExpense: Int64
    var expenseDelta: Int64
    var expenseChangeBPS: Int64?
    var averageDailyExpense: Int64
    var daysConsidered: Int
    var spendingIncomeRatioBPS: Int64?
    var topCategories: [InsiderCategoryTotalModel]
    var topExpenses: [InsiderExpenseRowModel]
    var estimated: Bool
}

struct RecurringModel: Codable, Identifiable, Sendable {
    let id: String
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

struct TravelEventModel: Codable, Identifiable, Sendable {
    let id: String
    var name: String
    var context: String?
    var startsOn: String?
    var endsOn: String?
    var active: Bool
    var createdAt: String
    var updatedAt: String
}

struct PortfolioAssetModel: Codable, Identifiable, Sendable {
    let id: String
    var symbol: String
    var name: String
    var latestPrice: Int64?
    var latestPriceAt: String?

    enum CodingKeys: String, CodingKey {
        case id
        case symbol
        case name
        case latestPrice
        case latestPriceAt
    }
}

struct PortfolioPositionModel: Codable, Identifiable, Sendable {
    var id: String { assetID }
    let assetID: String
    var symbol: String?
    var name: String?
    var quantity: String
    var costBasis: Int64
    var averageCost: Int64
    var realizedPnL: Int64
    var marketValue: Int64?
    var unrealizedPnL: Int64?
    var latestPrice: Int64?

    enum CodingKeys: String, CodingKey {
        case assetID
        case symbol
        case name
        case quantity
        case costBasis
        case averageCost
        case realizedPnL
        case marketValue
        case unrealizedPnL
        case latestPrice
    }
}

struct PortfolioSummaryModel: Codable, Sendable {
    var assets: [PortfolioAssetModel]
    var positions: [PortfolioPositionModel]
    var trades: [PortfolioTradeModel]
    var totalCostBasis: Int64
    var totalMarketValue: Int64?
    var totalUnrealizedPnL: Int64?

    enum CodingKeys: String, CodingKey {
        case assets
        case positions
        case trades
        case totalCostBasis
        case totalMarketValue
        case totalUnrealizedPnL
    }
}

struct CreditStatementModel: Codable, Sendable {
    var walletID: String
    var creditLimit: Int64
    var balance: Int64
    var availableCredit: Int64
    var amountDue: Int64
    var paymentStatus: String
    var paymentDueAt: String?
    var items: [TransactionModel]
}

struct JarItemModel: Codable, Identifiable, Sendable {
    var id: String { jarID }
    let jarID: String
    var name: String
    var allocationMode: String
    var allocationAmount: Int64?
    var allocationPercent: Double?
    var calculatedAllocation: Int64?
    var spent: Int64
    var active: Bool
}

struct JarMonthSummaryModel: Codable, Sendable {
    var month: String
    var timezone: String
    var actualIncome: Int64
    var totalSpent: Int64
    var unassignedSpent: Int64
    var totalAllocated: Int64
    var totalAllocationPercent: Double
    var overIncome: Bool
    var overOneHundredPercent: Bool
    var jars: [JarIdentityModel]
    var items: [JarItemModel]
}

struct JarIdentityModel: Codable, Identifiable, Sendable {
    var id: String { jarID }
    let jarID: String
    var name: String
}

struct JarConfigInputModel: Codable, Sendable {
    var month: String
    var name: String
    var allocationMode: String
    var allocationAmount: Int64?
    var allocationPercent: Double?
}

struct JarConfigModel: Codable, Sendable {
    var month: String
    var jarID: String
    var name: String
    var allocationMode: String
    var allocationAmount: Int64?
    var allocationPercent: Double?
    var active: Bool
}

struct JarCumulativeMonthModel: Codable, Sendable {
    var month: String
    var name: String?
    var spent: Int64
    var calculatedAllocation: Int64?
    var allocationCovered: Bool
}

struct JarCumulativeSummaryModel: Codable, Sendable {
    var jarID: String
    var name: String
    var fromMonth: String
    var toMonth: String
    var totalSpent: Int64
    var totalAllocated: Int64
    var allocationMonths: Int
    var monthsInRange: Int
    var allocationVariance: Int64?
    var months: [JarCumulativeMonthModel]
}

struct ProfileModel: Codable, Sendable {
    var id: String
    var name: String
    var email: String
    var timezone: String
    var timezoneConfirmed: Bool
    var createdAt: String
}

struct APIKeyModel: Codable, Identifiable, Sendable {
    let id: String
    var name: String
    var scopes: [String]
    var expiresAt: String?
    var revokedAt: String?
    var lastUsedAt: String?
    var createdAt: String
}

struct APIKeyCreatedModel: Codable, Sendable {
    var id: String
    var name: String
    var scopes: [String]
    var expiresAt: String?
    var secret: String
}

enum APIKeySecretClipboard {
    /// Copies the exact one-time secret only in response to an explicit caller
    /// action. The injected writer keeps the policy deterministic in tests.
    @discardableResult
    static func copy(_ secret: String, write: (String) -> Void) -> Bool {
        guard !secret.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty else { return false }
        write(secret)
        return true
    }
}
