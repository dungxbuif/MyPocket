import SwiftUI

struct BudgetReportView: View {
    let api: FinanceAPI
    let budget: BudgetModel
    let ownerID: String?
    let masked: Bool
    let onBack: () -> Void
    @State private var report: BudgetReportModel?
    @State private var loading = false
    @State private var error: String?
    @State private var showingTransactions = false

    var body: some View {
        Group {
            if showingTransactions {
                ReportTransactionsView(api: api, ownerID: ownerID, masked: masked, scope: .budget(id: budget.id), onBack: { showingTransactions = false })
            } else {
                ScrollView {
                    VStack(alignment: .leading, spacing: 8) {
                        MPScreenHeader(title: "Báo cáo ngân sách", leading: { MPIconButton(systemName: "chevron.left", label: "Quay lại", action: onBack) }, trailing: { EmptyView() })
                        MPText(budget.name, style: .emphasis)
                        if loading { MPLoadingView(label: "Đang tải báo cáo...") }
                        if let error { MPErrorView(title: "Không tải được báo cáo", detail: error) { Task { await load() } } }
                        if let report {
                            MPText("\(report.startDate) → \(report.endDate) · \(report.timezone)", style: .secondary)
                            MPSurfaceCard {
                                VStack(alignment: .leading, spacing: 8) {
                                    metric("Giới hạn gốc", report.baseLimitAmount)
                                    metric("Chuyển dư từ kỳ trước", report.carryoverAmount)
                                    metric("Giới hạn hiệu lực", report.effectiveLimitAmount)
                                    metric("Đã chi", report.spent)
                                    metric("Còn lại", report.remaining)
                                }
                            }
                            MPSurfaceCard {
                                VStack(alignment: .leading, spacing: 8) {
                                    MPText(report.forecast.statusLabel, style: .emphasis)
                                    if let amount = report.forecast.projectedSpent { metric("Chi tiêu cuối kỳ", amount) }
                                    if let amount = report.forecast.projectedRemaining { metric("Còn lại cuối kỳ", amount) }
                                    if report.forecast.overBudget == true { MPStatusBadge(text: report.forecast.status == "actual" ? "Đã vượt ngân sách" : "Có thể vượt ngân sách", tone: .danger) }
                                    MPText("\(report.forecast.elapsedDays) / \(report.forecast.totalDays) ngày · Bao gồm ngày không chi.", style: .secondary)
                                    if report.forecast.status == "estimated" { MPText("Ước tính tuyến tính từ chi tiêu thực tế; không phải cam kết hoặc tư vấn AI.", style: .secondary) }
                                }
                            }
                            .accessibilityIdentifier("budget-report-forecast")
                            MPButton("Xem giao dịch ngân sách", style: .secondary) { showingTransactions = true }
                            MPSurfaceCard {
                                VStack(alignment: .leading, spacing: 8) {
                                    MPText("Chi tiêu từng ngày", style: .emphasis)
                                    if report.daily.isEmpty { MPText("Chưa có ngày đã bắt đầu trong kỳ.", style: .secondary) }
                                    ForEach(report.daily) { day in metric(day.date, day.spent) }
                                }
                            }
                            MPText("Cập nhật: \(report.asOf)", style: .secondary)
                        }
                    }
                    .padding(MPMetrics.pagePadding)
                }
            }
        }
        .background(MPColor.canvas.ignoresSafeArea())
        .task { await load() }
    }

    private func metric(_ title: String, _ value: Int64) -> some View {
        MPMetric(title: title, value: MPFormat.money(value, masked: masked))
    }

    private func load() async {
        guard !loading else { return }
        loading = true; error = nil; report = nil
        defer { loading = false }
        do {
            let loaded = try await api.budgetReport(id: budget.id, ownerID: ownerID)
            try Task.checkCancellation()
            report = loaded
        } catch is CancellationError { }
        catch { self.error = (error as? APIError)?.userMessage ?? "Báo cáo chưa khả dụng. Thử lại khi máy chủ hỗ trợ tính năng này." }
    }
}
