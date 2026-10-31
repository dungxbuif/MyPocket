import Foundation
import Combine

@MainActor
final class MonthlyReportsViewModel: ObservableObject {
    @Published private(set) var report: MonthlyReportModel?
    @Published private(set) var isLoading = false
    @Published private(set) var error: String?
    @Published private(set) var usedFallback = false
    @Published private(set) var insider: InsiderSummaryModel?

    private var cache: [ReportScope: MonthlyReportModel] = [:]

    func load(api: FinanceAPI, month: String, walletID: String? = nil, ownerID: String? = nil, granularity: String = "day") async {
        let scope = ReportScope(ownerID: ownerID, month: month, walletID: walletID, granularity: granularity)
        guard !isLoading else { return }
        if let cached = cache[scope] {
            report = cached
            usedFallback = false
            error = nil
            return
        }
        isLoading = true
        error = nil
        usedFallback = false
        insider = nil
        defer { isLoading = false }
        async let insiderResult: InsiderSummaryModel? = try? api.insider(month: month, ownerID: ownerID)
        do {
            let loaded = try await api.monthlyReport(month: month, walletID: walletID, granularity: granularity, ownerID: ownerID)
            guard loaded.month == month else { throw APIError.decoding("Report scope mismatch") }
            cache[scope] = loaded
            report = loaded
        } catch {
            do {
                let fallback = try await loadAggregateFallback(api: api, month: month, ownerID: ownerID)
                cache[scope] = fallback
                report = fallback
                usedFallback = true
            } catch {
                report = nil
                self.error = "Không thể tải báo cáo cho tháng này."
            }
        }
        insider = await insiderResult
    }

    func clear() {
        report = nil
        error = nil
        usedFallback = false
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
