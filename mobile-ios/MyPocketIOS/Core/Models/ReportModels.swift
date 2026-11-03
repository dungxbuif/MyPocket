import Foundation

struct ReportScope: Hashable, Sendable {
    var ownerID: String?
    var month: String
    var walletID: String?
    var granularity: String
}

struct ReportPoint: Codable, Identifiable, Sendable, Equatable {
    var id: String { from ?? label }
    var label: String
    var value: Int64
    var comparison: Int64?
    var detail: String?
    var from: String?
    var endExclusive: String?
    var income: Int64?
    var expense: Int64?
    var net: Int64?
    var transactionCount: Int?

    enum CodingKeys: String, CodingKey {
        case label, value, comparison, detail, from, income, expense, net
        case endExclusive, transactionCount
    }

    init(label: String, value: Int64, comparison: Int64? = nil, detail: String? = nil, from: String? = nil, endExclusive: String? = nil, income: Int64? = nil, expense: Int64? = nil, net: Int64? = nil, transactionCount: Int? = nil) {
        self.label = label; self.value = value; self.comparison = comparison; self.detail = detail
        self.from = from; self.endExclusive = endExclusive; self.income = income; self.expense = expense
        self.net = net; self.transactionCount = transactionCount
    }
}

struct ReportCategoryTotal: Codable, Identifiable, Sendable, Equatable {
    var id: String { targetID ?? categoryID ?? name }
    var categoryID: String?
    var name: String
    var amount: Int64
    var count: Int
    var shareBPS: Int64 = 0
    var bucket: String? = nil
    var targetID: String? = nil
    var drilldown: ReportDrilldownModel? = nil

    enum CodingKeys: String, CodingKey {
        case name, amount, bucket, drilldown
        case transactionCount, count
        case categoryID, shareBPS, targetID
    }

    init(categoryID: String? = nil, name: String, amount: Int64, count: Int, shareBPS: Int64 = 0, bucket: String? = nil, targetID: String? = nil, drilldown: ReportDrilldownModel? = nil) {
        self.categoryID = categoryID; self.name = name; self.amount = amount; self.count = count
        self.shareBPS = shareBPS; self.bucket = bucket; self.targetID = targetID; self.drilldown = drilldown
    }

    init(from decoder: Decoder) throws {
        let values = try decoder.container(keyedBy: CodingKeys.self)
        categoryID = try values.decodeIfPresent(String.self, forKey: .categoryID)
        name = try values.decode(String.self, forKey: .name)
        amount = try values.decode(Int64.self, forKey: .amount)
        count = try values.decodeIfPresent(Int.self, forKey: .transactionCount) ?? values.decodeIfPresent(Int.self, forKey: .count) ?? 0
        shareBPS = try values.decodeIfPresent(Int64.self, forKey: .shareBPS) ?? 0
        bucket = try values.decodeIfPresent(String.self, forKey: .bucket)
        targetID = try values.decodeIfPresent(String.self, forKey: .targetID)
        drilldown = try values.decodeIfPresent(ReportDrilldownModel.self, forKey: .drilldown)
    }

    func encode(to encoder: Encoder) throws {
        var values = encoder.container(keyedBy: CodingKeys.self)
        try values.encodeIfPresent(categoryID, forKey: .categoryID)
        try values.encode(name, forKey: .name)
        try values.encode(amount, forKey: .amount)
        try values.encode(count, forKey: .transactionCount)
        try values.encode(shareBPS, forKey: .shareBPS)
        try values.encodeIfPresent(bucket, forKey: .bucket)
        try values.encodeIfPresent(targetID, forKey: .targetID)
        try values.encodeIfPresent(drilldown, forKey: .drilldown)
    }
}

struct ReportRangeModel: Codable, Sendable, Equatable {
    var from: String
    var endExclusive: String
    var days: Int
    var isPartial: Bool

    enum CodingKeys: String, CodingKey {
        case from, days
        case endExclusive, isPartial
    }
}

struct ReportDrilldownModel: Codable, Sendable, Equatable {
    var targetID: String?
    var from: String?
    var endExclusive: String?
    var walletID: String?
    var categoryIDs: [String]
    var categoryBucket: String?
    var type: String?

    enum CodingKeys: String, CodingKey {
        case from, type
        case targetID, endExclusive, walletID, categoryIDs, categoryBucket
    }

    init(targetID: String? = nil, from: String? = nil, endExclusive: String? = nil, walletID: String? = nil, categoryIDs: [String] = [], categoryBucket: String? = nil, type: String? = nil) {
        self.targetID = targetID; self.from = from; self.endExclusive = endExclusive; self.walletID = walletID
        self.categoryIDs = categoryIDs; self.categoryBucket = categoryBucket; self.type = type
    }

    init(from decoder: Decoder) throws {
        let values = try decoder.container(keyedBy: CodingKeys.self)
        targetID = try values.decodeIfPresent(String.self, forKey: .targetID)
        from = try values.decodeIfPresent(String.self, forKey: .from)
        endExclusive = try values.decodeIfPresent(String.self, forKey: .endExclusive)
        walletID = try values.decodeIfPresent(String.self, forKey: .walletID)
        categoryIDs = try values.decodeIfPresent([String].self, forKey: .categoryIDs) ?? []
        categoryBucket = try values.decodeIfPresent(String.self, forKey: .categoryBucket)
        type = try values.decodeIfPresent(String.self, forKey: .type)
    }
}

struct ReportComparisonModel: Codable, Sendable, Equatable {
    var comparable: Bool
    var commonDays: Int
    var incomeDelta: Int64
    var expenseDelta: Int64
    var netDelta: Int64
    var expenseChangeBPS: Int64?
    var incomeChangeBPS: Int64?

    enum CodingKeys: String, CodingKey {
        case comparable
        case commonDays, incomeDelta, expenseDelta, netDelta, expenseChangeBPS, incomeChangeBPS
    }
}

struct ReportContributorModel: Codable, Identifiable, Sendable, Equatable {
    var id: String { targetID ?? factRef }
    var categoryID: String?
    var name: String
    var currentAmount: Int64
    var previousAmount: Int64
    var delta: Int64
    var currentCount: Int64
    var previousCount: Int64
    var changeBPS: Int64?
    var factRef: String
    var targetID: String?
    var drilldown: ReportDrilldownModel?

    enum CodingKeys: String, CodingKey {
        case name, delta, drilldown
        case categoryID, currentAmount, previousAmount, currentCount, previousCount, changeBPS, factRef, targetID
    }
}

struct MonthlyReportModel: Decodable, Sendable, Equatable {
    var month: String
    var timezone: String?
    var walletID: String?
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
    var previousDaily: [ReportPoint]
    var contributors: [ReportContributorModel]
    var insights: [ReportInsightModel]
    var currentRange: ReportRangeModel?
    var comparisonRange: ReportRangeModel?
    var comparison: ReportComparisonModel?

    enum CodingKeys: String, CodingKey {
        case month, timezone, income, expense, net, transactionCount
        case weekly, daily, categories, insight, revision
        case snapshotID, dataRevision
        case scope, totals, series, insights, contributors, comparison
        case previousSeries, currentRange, comparisonRange
    }

    init(month: String, timezone: String? = nil, walletID: String? = nil, snapshotID: String? = nil, dataRevision: Int64? = nil, income: Int64, expense: Int64, net: Int64, transactionCount: Int, weekly: [ReportPoint] = [], daily: [ReportPoint] = [], categories: [ReportCategoryTotal] = [], insight: String? = nil, revision: String? = nil, previousDaily: [ReportPoint] = [], contributors: [ReportContributorModel] = [], insights: [ReportInsightModel] = [], currentRange: ReportRangeModel? = nil, comparisonRange: ReportRangeModel? = nil, comparison: ReportComparisonModel? = nil) {
        self.month = month; self.timezone = timezone; self.walletID = walletID; self.snapshotID = snapshotID; self.dataRevision = dataRevision; self.income = income; self.expense = expense; self.net = net
        self.transactionCount = transactionCount; self.weekly = weekly; self.daily = daily; self.categories = categories
        self.insight = insight; self.revision = revision
        self.previousDaily = previousDaily; self.contributors = contributors; self.insights = insights
        self.currentRange = currentRange; self.comparisonRange = comparisonRange; self.comparison = comparison
    }

    init(from decoder: Decoder) throws {
        let values = try decoder.container(keyedBy: CodingKeys.self)
        timezone = try values.decodeIfPresent(String.self, forKey: .timezone)
        walletID = nil
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
            previousDaily = []; contributors = []; insights = []
            currentRange = nil; comparisonRange = nil; comparison = nil
            return
        }

        struct ScopePayload: Decodable {
            var month: String
            var walletID: String?
            var granularity: String?
            enum CodingKeys: String, CodingKey { case month, granularity, walletID }
        }
        struct TotalsPayload: Decodable {
            var income: Int64
            var expense: Int64
            var net: Int64
            var transactionCount: Int64

            enum CodingKeys: String, CodingKey {
                case income, expense, net
                case transactionCount
            }

            init(from decoder: Decoder) throws {
                let values = try decoder.container(keyedBy: CodingKeys.self)
                income = try values.decode(Int64.self, forKey: .income)
                expense = try values.decode(Int64.self, forKey: .expense)
                net = try values.decode(Int64.self, forKey: .net)
                transactionCount = try values.decodeIfPresent(Int64.self, forKey: .transactionCount) ?? 0
            }
        }
        struct SeriesPayload: Decodable {
            var label: String
            var from: String?
            var endExclusive: String?
            var income: Int64
            var expense: Int64
            var net: Int64?
            var transactionCount: Int64

            enum CodingKeys: String, CodingKey {
                case label, from, income, expense, net
                case endExclusive, transactionCount
            }

            init(from decoder: Decoder) throws {
                let values = try decoder.container(keyedBy: CodingKeys.self)
                label = try values.decode(String.self, forKey: .label)
                from = try values.decodeIfPresent(String.self, forKey: .from)
                endExclusive = try values.decodeIfPresent(String.self, forKey: .endExclusive)
                income = try values.decode(Int64.self, forKey: .income)
                expense = try values.decode(Int64.self, forKey: .expense)
                transactionCount = try values.decodeIfPresent(Int64.self, forKey: .transactionCount) ?? 0
            }
        }
        struct CategoryPayload: Decodable {
            var categoryID: String?
            var name: String
            var amount: Int64
            var transactionCount: Int64

            enum CodingKeys: String, CodingKey {
                case name, amount
                case categoryID, transactionCount
            }

            init(from decoder: Decoder) throws {
                let values = try decoder.container(keyedBy: CodingKeys.self)
                categoryID = try values.decodeIfPresent(String.self, forKey: .categoryID)
                name = try values.decode(String.self, forKey: .name)
                amount = try values.decode(Int64.self, forKey: .amount)
                transactionCount = try values.decodeIfPresent(Int64.self, forKey: .transactionCount) ?? 0
            }
        }
        let scope = try values.decode(ScopePayload.self, forKey: .scope)
        let totals = try values.decode(TotalsPayload.self, forKey: .totals)
        let points = try values.decodeIfPresent([SeriesPayload].self, forKey: .series) ?? []
        let mappedPoints = points.map { point in
            ReportPoint(label: point.label, value: point.expense, comparison: point.income, detail: "Thu \(point.income) · Chi \(point.expense)", from: point.from, endExclusive: point.endExclusive, income: point.income, expense: point.expense, net: point.net, transactionCount: Int(clamping: point.transactionCount))
        }
        month = scope.month
        walletID = scope.walletID
        income = totals.income
        expense = totals.expense
        net = totals.net
        transactionCount = Int(clamping: totals.transactionCount)
        if scope.granularity == "day" { weekly = []; daily = mappedPoints } else { weekly = mappedPoints; daily = [] }
        categories = try values.decodeIfPresent([ReportCategoryTotal].self, forKey: .categories) ?? []
        insights = try values.decodeIfPresent([ReportInsightModel].self, forKey: .insights) ?? []
        insight = insights.first?.body
        revision = dataRevision.map(String.init)
        let previous = try values.decodeIfPresent([SeriesPayload].self, forKey: .previousSeries) ?? []
        previousDaily = previous.map { ReportPoint(label: $0.label, value: $0.expense, comparison: $0.income, detail: nil, from: $0.from, endExclusive: $0.endExclusive, income: $0.income, expense: $0.expense, net: $0.net, transactionCount: Int(clamping: $0.transactionCount)) }
        contributors = try values.decodeIfPresent([ReportContributorModel].self, forKey: .contributors) ?? []
        currentRange = try values.decodeIfPresent(ReportRangeModel.self, forKey: .currentRange)
        comparisonRange = try values.decodeIfPresent(ReportRangeModel.self, forKey: .comparisonRange)
        comparison = try values.decodeIfPresent(ReportComparisonModel.self, forKey: .comparison)
    }
}

struct ReportInsightMetricModel: Codable, Sendable, Equatable {
    var label: String
    var amount: Int64
}

struct ReportInsightModel: Codable, Identifiable, Sendable, Equatable {
    var id: String
    var kind: String?
    var title: String?
    var body: String
    var factRefs: [String] = []
    var drilldown: ReportDrilldownModel? = nil

    var metrics: [ReportInsightMetricModel] = []
    var periodLabel: String?
    var actionHint: String?

    enum CodingKeys: String, CodingKey {
        case id, kind, title, body, drilldown
        case factRefs, metrics, periodLabel, actionHint
    }

    init(id: String, kind: String? = nil, title: String? = nil, body: String, factRefs: [String] = [], drilldown: ReportDrilldownModel? = nil, metrics: [ReportInsightMetricModel] = [], periodLabel: String? = nil, actionHint: String? = nil) {
        self.id = id; self.kind = kind; self.title = title; self.body = body; self.factRefs = factRefs; self.drilldown = drilldown
        self.metrics = metrics; self.periodLabel = periodLabel; self.actionHint = actionHint
    }

    init(from decoder: Decoder) throws {
        let values = try decoder.container(keyedBy: CodingKeys.self)
        id = try values.decodeIfPresent(String.self, forKey: .id) ?? UUID().uuidString
        kind = try values.decodeIfPresent(String.self, forKey: .kind)
        title = try values.decodeIfPresent(String.self, forKey: .title)
        body = try values.decodeIfPresent(String.self, forKey: .body) ?? ""
        factRefs = try values.decodeIfPresent([String].self, forKey: .factRefs) ?? []
        drilldown = try values.decodeIfPresent(ReportDrilldownModel.self, forKey: .drilldown)
        metrics = try values.decodeIfPresent([ReportInsightMetricModel].self, forKey: .metrics) ?? []
        periodLabel = try values.decodeIfPresent(String.self, forKey: .periodLabel)
        actionHint = try values.decodeIfPresent(String.self, forKey: .actionHint)
    }
}

struct ReportAnalysisRequestModel: Encodable, Sendable {
    var snapshotID: String
    var clientRequestID: String

    enum CodingKeys: String, CodingKey {
        case snapshotID, clientRequestID
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
    var retainedUntil: String?
    var analysisFingerprint: String?
    var error: String?
    var providerCalled: Bool?
    var ruleVersion: String?
    var promptVersion: String?

    enum CodingKeys: String, CodingKey {
        case id
        case snapshotID
        case status
        case clientRequestID, serverTime
        case revision
        case summary, observations, suggestions, article
        case errorCode, expiresAt, retainedUntil, analysisFingerprint, providerCalled, ruleVersion, promptVersion
    }

    init(id: String, snapshotID: String? = nil, status: String, clientRequestID: String? = nil, serverTime: String? = nil, revision: Int64? = nil, summary: String? = nil, observations: [ReportInsightModel] = [], suggestions: [ReportInsightModel] = [], article: String? = nil, expiresAt: String? = nil, retainedUntil: String? = nil, analysisFingerprint: String? = nil, error: String? = nil, providerCalled: Bool? = nil, ruleVersion: String? = nil, promptVersion: String? = nil) {
        self.id = id; self.snapshotID = snapshotID; self.status = status; self.clientRequestID = clientRequestID; self.serverTime = serverTime
        self.revision = revision; self.summary = summary; self.observations = observations; self.suggestions = suggestions; self.article = article
        self.expiresAt = expiresAt; self.retainedUntil = retainedUntil; self.analysisFingerprint = analysisFingerprint; self.error = error
        self.providerCalled = providerCalled; self.ruleVersion = ruleVersion; self.promptVersion = promptVersion
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
        retainedUntil = try values.decodeIfPresent(String.self, forKey: .retainedUntil)
        analysisFingerprint = try values.decodeIfPresent(String.self, forKey: .analysisFingerprint)
        error = try values.decodeIfPresent(String.self, forKey: .errorCode)
        providerCalled = try values.decodeIfPresent(Bool.self, forKey: .providerCalled)
        ruleVersion = try values.decodeIfPresent(String.self, forKey: .ruleVersion)
        promptVersion = try values.decodeIfPresent(String.self, forKey: .promptVersion)
    }

    // Historical provider_called=false was a persistence bug, not provenance.
    var hasCurrentProvenance: Bool {
        Self.version(ruleVersion, prefix: "report-analysis-v") >= 3
            && Self.version(promptVersion, prefix: "report-analysis-prompt-v") >= 3
            && providerCalled != nil
    }

    /// Provenance is trustworthy from v3 onward; a later backend version must not make every new
    /// analysis look stale until an app release.
    private static func version(_ value: String?, prefix: String) -> Int {
        guard let value, value.hasPrefix(prefix) else { return 0 }
        return Int(value.dropFirst(prefix.count)) ?? 0
    }

    var provenanceLabel: String {
        guard hasCurrentProvenance, let providerCalled else {
            return "Bản cũ chưa có thông tin xác thực về cách tạo nhận xét."
        }
        if error == "provider_output_invalid" || error == "provider_empty" {
            return "Nhận xét theo quy tắc từ số liệu; lựa chọn của AI chưa đạt yêu cầu."
        }
        return providerCalled
            ? "AI chọn điểm đáng chú ý; số liệu và nội dung được đối chiếu từ báo cáo."
            : "Nhận xét theo quy tắc từ số liệu báo cáo, không gọi AI."
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

struct ReportAnalysisLatestModel: Decodable, Sendable {
    var completed: ReportAnalysisModel?
    var active: ReportAnalysisModel?
    var freshness: String
    var staleReasons: [String]
    var currentSnapshotID: String

    enum CodingKeys: String, CodingKey {
        case completed, active, freshness
        case staleReasons, currentSnapshotID
    }
}
