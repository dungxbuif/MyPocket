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

struct MonthlyReportModel: Decodable, Sendable, Equatable {
    var month: String
    var timezone: String?
    var snapshotID: String?
    var dataRevision: Int64?
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
        case snapshotID = "snapshot_id"
        case dataRevision = "data_revision"
        case scope, totals, series, insights
    }

    init(month: String, timezone: String? = nil, snapshotID: String? = nil, dataRevision: Int64? = nil, income: Int64, expense: Int64, net: Int64, transactionCount: Int, weekly: [ReportPoint] = [], daily: [ReportPoint] = [], categories: [ReportCategoryTotal] = [], insight: String? = nil, revision: String? = nil) {
        self.month = month; self.timezone = timezone; self.snapshotID = snapshotID; self.dataRevision = dataRevision; self.income = income; self.expense = expense; self.net = net
        self.transactionCount = transactionCount; self.weekly = weekly; self.daily = daily; self.categories = categories
        self.insight = insight; self.revision = revision
    }

    init(from decoder: Decoder) throws {
        let values = try decoder.container(keyedBy: CodingKeys.self)
        timezone = try values.decodeIfPresent(String.self, forKey: .timezone)
        snapshotID = try values.decodeIfPresent(String.self, forKey: .snapshotID)
        dataRevision = try values.decodeIfPresent(Int64.self, forKey: .dataRevision)
        if let legacyMonth = try values.decodeIfPresent(String.self, forKey: .month) {
            month = legacyMonth
            income = try values.decodeIfPresent(Int64.self, forKey: .income) ?? 0
            expense = try values.decodeIfPresent(Int64.self, forKey: .expense) ?? 0
            net = try values.decodeIfPresent(Int64.self, forKey: .net) ?? income - expense
            transactionCount = try values.decodeIfPresent(Int.self, forKey: .transactionCount) ?? 0
            weekly = try values.decodeIfPresent([ReportPoint].self, forKey: .weekly) ?? []
            daily = try values.decodeIfPresent([ReportPoint].self, forKey: .daily) ?? []
            categories = try values.decodeIfPresent([ReportCategoryTotal].self, forKey: .categories) ?? []
            insight = try values.decodeIfPresent(String.self, forKey: .insight)
            revision = try values.decodeIfPresent(String.self, forKey: .revision)
            return
        }

        struct ScopePayload: Decodable { var month: String; var granularity: String? }
        struct TotalsPayload: Decodable {
            var income: Int64
            var expense: Int64
            var net: Int64
            var transactionCount: Int64

            enum CodingKeys: String, CodingKey {
                case income, expense, net
                case transactionCount = "transaction_count"
                case legacyTransactionCount = "transactionCount"
            }

            init(from decoder: Decoder) throws {
                let values = try decoder.container(keyedBy: CodingKeys.self)
                income = try values.decode(Int64.self, forKey: .income)
                expense = try values.decode(Int64.self, forKey: .expense)
                net = try values.decode(Int64.self, forKey: .net)
                transactionCount = try values.decodeIfPresent(Int64.self, forKey: .transactionCount)
                    ?? values.decode(Int64.self, forKey: .legacyTransactionCount)
            }
        }
        struct SeriesPayload: Decodable {
            var label: String
            var income: Int64
            var expense: Int64
            var transactionCount: Int64

            enum CodingKeys: String, CodingKey {
                case label, income, expense
                case transactionCount = "transaction_count"
                case legacyTransactionCount = "transactionCount"
            }

            init(from decoder: Decoder) throws {
                let values = try decoder.container(keyedBy: CodingKeys.self)
                label = try values.decode(String.self, forKey: .label)
                income = try values.decode(Int64.self, forKey: .income)
                expense = try values.decode(Int64.self, forKey: .expense)
                transactionCount = try values.decodeIfPresent(Int64.self, forKey: .transactionCount)
                    ?? values.decode(Int64.self, forKey: .legacyTransactionCount)
            }
        }
        struct CategoryPayload: Decodable {
            var categoryID: String?
            var name: String
            var amount: Int64
            var transactionCount: Int64

            enum CodingKeys: String, CodingKey {
                case name, amount
                case categoryID = "category_id"
                case legacyCategoryID = "categoryID"
                case transactionCount = "transaction_count"
                case legacyTransactionCount = "transactionCount"
            }

            init(from decoder: Decoder) throws {
                let values = try decoder.container(keyedBy: CodingKeys.self)
                categoryID = try values.decodeIfPresent(String.self, forKey: .categoryID)
                    ?? values.decodeIfPresent(String.self, forKey: .legacyCategoryID)
                name = try values.decode(String.self, forKey: .name)
                amount = try values.decode(Int64.self, forKey: .amount)
                transactionCount = try values.decodeIfPresent(Int64.self, forKey: .transactionCount)
                    ?? values.decode(Int64.self, forKey: .legacyTransactionCount)
            }
        }
        let scope = try values.decode(ScopePayload.self, forKey: .scope)
        let totals = try values.decode(TotalsPayload.self, forKey: .totals)
        let points = try values.decodeIfPresent([SeriesPayload].self, forKey: .series) ?? []
        let mappedPoints = points.map { point in
            ReportPoint(label: point.label, value: point.expense, comparison: point.income, detail: "Thu \(point.income) · Chi \(point.expense)")
        }
        month = scope.month
        income = totals.income
        expense = totals.expense
        net = totals.net
        transactionCount = Int(clamping: totals.transactionCount)
        if scope.granularity == "day" { weekly = []; daily = mappedPoints } else { weekly = mappedPoints; daily = [] }
        categories = try (values.decodeIfPresent([CategoryPayload].self, forKey: .categories) ?? []).map {
            ReportCategoryTotal(categoryID: $0.categoryID, name: $0.name, amount: $0.amount, count: Int(clamping: $0.transactionCount))
        }
        let insights = try values.decodeIfPresent([ReportInsightModel].self, forKey: .insights) ?? []
        insight = insights.first?.body
        revision = dataRevision.map(String.init)
    }
}

struct ReportInsightModel: Codable, Identifiable, Sendable, Equatable {
    var id: String
    var kind: String?
    var title: String?
    var body: String
}

struct ReportAnalysisRequestModel: Encodable, Sendable {
    var snapshotID: String
    var clientRequestID: String

    enum CodingKeys: String, CodingKey {
        case snapshotID = "snapshot_id"
        case clientRequestID = "client_request_id"
    }
}

struct ReportAnalysisModel: Decodable, Sendable {
    var id: String
    var snapshotID: String?
    var status: String
    var clientRequestID: String?
    var serverTime: String?
    var revision: Int64?
    var summary: String?
    var observations: [ReportInsightModel]
    var suggestions: [ReportInsightModel]
    var article: String?
    var expiresAt: String?
    var error: String?

    enum CodingKeys: String, CodingKey {
        case id
        case snapshotID = "snapshot_id"
        case status
        case clientRequestID = "client_request_id"
        case serverTime = "server_time"
        case revision
        case summary, observations, suggestions, article, expiresAt, error
    }

    init(from decoder: Decoder) throws {
        let values = try decoder.container(keyedBy: CodingKeys.self)
        id = try values.decodeIfPresent(String.self, forKey: .id) ?? ""
        snapshotID = try values.decodeIfPresent(String.self, forKey: .snapshotID)
        status = try values.decode(String.self, forKey: .status)
        clientRequestID = try values.decodeIfPresent(String.self, forKey: .clientRequestID)
        serverTime = try values.decodeIfPresent(String.self, forKey: .serverTime)
        if let numericRevision = try? values.decodeIfPresent(Int64.self, forKey: .revision) { revision = numericRevision } else { revision = nil }
        summary = try values.decodeIfPresent(String.self, forKey: .summary)
        observations = try values.decodeIfPresent([ReportInsightModel].self, forKey: .observations) ?? []
        suggestions = try values.decodeIfPresent([ReportInsightModel].self, forKey: .suggestions) ?? []
        article = try values.decodeIfPresent(String.self, forKey: .article)
        expiresAt = try values.decodeIfPresent(String.self, forKey: .expiresAt)
        error = try values.decodeIfPresent(String.self, forKey: .error)
    }

    var renderedArticle: String {
        if let article, !article.isEmpty { return article }
        var sections: [String] = []
        if let summary, !summary.isEmpty { sections.append(summary) }
        if !observations.isEmpty { sections.append("Quan sát\n" + observations.map { "• " + $0.body }.joined(separator: "\n")) }
        if !suggestions.isEmpty { sections.append("Gợi ý\n" + suggestions.map { "• " + $0.body }.joined(separator: "\n")) }
        return sections.joined(separator: "\n\n")
    }
}
