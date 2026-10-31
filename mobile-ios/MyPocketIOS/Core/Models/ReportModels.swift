import Foundation

struct ReportScope: Hashable, Sendable {
    var ownerID: String?
    var month: String
    var walletID: String?
    var granularity: String
}

struct ReportPoint: Codable, Identifiable, Sendable, Equatable {
    var id: String { label }
    var label: String
    var value: Int64
    var comparison: Int64?
    var detail: String?
}

struct ReportCategoryTotal: Codable, Identifiable, Sendable, Equatable {
    var id: String { categoryID ?? name }
    var categoryID: String?
    var name: String
    var amount: Int64
    var count: Int
}

struct MonthlyReportModel: Codable, Sendable, Equatable {
    var month: String
    var timezone: String?
    var income: Int64
    var expense: Int64
    var net: Int64
    var transactionCount: Int
    var weekly: [ReportPoint]
    var daily: [ReportPoint]
    var categories: [ReportCategoryTotal]
    var insight: String?
    var revision: String?

    enum CodingKeys: String, CodingKey {
        case month, timezone, income, expense, net, transactionCount
        case weekly, daily, categories, insight, revision
    }

    init(month: String, timezone: String? = nil, income: Int64, expense: Int64, net: Int64, transactionCount: Int, weekly: [ReportPoint] = [], daily: [ReportPoint] = [], categories: [ReportCategoryTotal] = [], insight: String? = nil, revision: String? = nil) {
        self.month = month; self.timezone = timezone; self.income = income; self.expense = expense; self.net = net
        self.transactionCount = transactionCount; self.weekly = weekly; self.daily = daily; self.categories = categories
        self.insight = insight; self.revision = revision
    }

    init(from decoder: Decoder) throws {
        let values = try decoder.container(keyedBy: CodingKeys.self)
        month = try values.decode(String.self, forKey: .month)
        timezone = try values.decodeIfPresent(String.self, forKey: .timezone)
        income = try values.decodeIfPresent(Int64.self, forKey: .income) ?? 0
        expense = try values.decodeIfPresent(Int64.self, forKey: .expense) ?? 0
        net = try values.decodeIfPresent(Int64.self, forKey: .net) ?? income - expense
        transactionCount = try values.decodeIfPresent(Int.self, forKey: .transactionCount) ?? 0
        weekly = try values.decodeIfPresent([ReportPoint].self, forKey: .weekly) ?? []
        daily = try values.decodeIfPresent([ReportPoint].self, forKey: .daily) ?? []
        categories = try values.decodeIfPresent([ReportCategoryTotal].self, forKey: .categories) ?? []
        insight = try values.decodeIfPresent(String.self, forKey: .insight)
        revision = try values.decodeIfPresent(String.self, forKey: .revision)
    }
}

struct ReportAnalysisRequestModel: Codable, Sendable {
    var requestID: String
    var month: String
    var ownerID: String?
    var revision: String?
}

struct ReportAnalysisModel: Codable, Sendable {
    var requestID: String
    var status: String
    var revision: String?
    var article: String?
    var expiresAt: String?
    var error: String?
}
