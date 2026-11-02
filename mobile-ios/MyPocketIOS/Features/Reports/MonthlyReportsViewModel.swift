import Foundation
import Combine

@MainActor
final class MonthlyReportsViewModel: ObservableObject {
    @Published private(set) var report: MonthlyReportModel?
    @Published private(set) var isLoading = false
    @Published private(set) var error: String?
    @Published private(set) var usedFallback = false
    @Published private(set) var wallets: [WalletModel] = []
    private var cache: [ReportScope: MonthlyReportModel] = [:]
    private var requestGeneration = 0
    private var walletRequestGeneration = 0
    private var walletsOwnerID: String?

    func load(api: FinanceAPI, month: String, walletID: String? = nil, ownerID: String? = nil, granularity: String = "day") async {
        let scope = ReportScope(ownerID: ownerID, month: month, walletID: walletID, granularity: granularity)
        requestGeneration += 1
        let generation = requestGeneration
        // A cached snapshot is only presentation while the revision-aware
        // server read refreshes it; never early-return stale financial data.
        report = cache[scope]
        isLoading = true
        error = nil
        usedFallback = false
        defer { if generation == requestGeneration { isLoading = false } }
        do {
            let loaded = try await api.monthlyReport(month: month, walletID: walletID, granularity: granularity, ownerID: ownerID)
            try Task.checkCancellation()
            guard generation == requestGeneration, loaded.month == month, loaded.walletID == walletID else { return }
            cache[scope] = loaded
            report = loaded
        } catch {
            guard generation == requestGeneration, !Task.isCancelled else { return }
            if walletID == nil { do {
                let fallback = try await loadAggregateFallback(api: api, month: month, ownerID: ownerID)
                guard generation == requestGeneration, !Task.isCancelled else { return }
                cache[scope] = fallback
                report = fallback
                usedFallback = true
            } catch {
                guard generation == requestGeneration, !Task.isCancelled else { return }
                if cache[scope] == nil { report = nil }
                self.error = "Không thể tải báo cáo cho tháng này."
            } } else {
                if cache[scope] == nil { report = nil }
                self.error = "Không thể tải báo cáo cho ví này."
            }
        }
    }

    func loadWallets(api: FinanceAPI, ownerID: String?) async {
        guard wallets.isEmpty || walletsOwnerID != ownerID else { return }
        walletRequestGeneration += 1
        let generation = walletRequestGeneration
        if walletsOwnerID != ownerID { wallets = [] }
        walletsOwnerID = ownerID
        let loaded = (try? await api.wallets(ownerID: ownerID)) ?? []
        guard generation == walletRequestGeneration, !Task.isCancelled else { return }
        wallets = loaded
    }

    func clear() {
        report = nil
        error = nil
        usedFallback = false
        requestGeneration += 1
    }

    private func loadAggregateFallback(api: FinanceAPI, month: String, ownerID: String?) async throws -> MonthlyReportModel {
        let summaryValue = try await api.monthSummary(month: month, ownerID: ownerID)
        return MonthlyReportModel(
            month: month,
            timezone: summaryValue.timezone,
            income: summaryValue.income,
            expense: summaryValue.expense,
            net: summaryValue.net,
            transactionCount: summaryValue.transactionCount,
            categories: summaryValue.categories.map {
                ReportCategoryTotal(categoryID: $0.categoryID, name: $0.name, amount: $0.amount, count: $0.count)
            },
            insight: "Đang dùng số liệu tổng hợp tương thích với máy chủ hiện tại.",
            revision: nil
        )
    }
}
