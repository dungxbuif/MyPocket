import Foundation

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
    var walletIDs: [String]
}

enum CategoryEditPolicy {
    static func metadataIsReadOnly(_ category: CategoryModel) -> Bool { category.isSystem }
    static func canDelete(_ category: CategoryModel) -> Bool { !category.isSystem }
}

extension CategoryModel {
    func applies(to transactionType: String, walletID: String, walletType: String) -> Bool {
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
