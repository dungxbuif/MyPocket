import SwiftUI

struct MonthlyReportsView: View {
    @Environment(\.mpTabIsActive) private var tabIsActive
    let api: FinanceAPI
    let onBack: () -> Void
    let masked: Bool
    let ownerID: String?
    let showsBack: Bool
    @Environment(\.reportAssistantHandoff) private var onReportHandoff
    @State private var insiderRequest = UUID()
    @StateObject private var model = MonthlyReportsViewModel()
    @State private var month = MPFormat.currentMonth
    @State private var walletSelection = "__all__"
    @State private var selectedPoint: ReportPoint?
    @State private var drilldown: ReportTransactionScope?
    @State private var evidenceError: String?
    @State private var accountTimezone = TimeZone.current.identifier

    init(api: FinanceAPI, onBack: @escaping () -> Void, masked: Bool, ownerID: String?, showsBack: Bool = true, onChatHandoff: @escaping (String) -> Void = { _ in }) {
        self.api = api; self.onBack = onBack; self.masked = masked; self.ownerID = ownerID
        self.showsBack = showsBack
        // The shell environment carries structured report handoffs for every entry route.
    }

    private var walletID: String? { walletSelection == "__all__" ? nil : walletSelection }
    private var scopeKey: String { (ownerID ?? "") + ":" + month + ":" + (walletID ?? "all") }

    var body: some View {
        Group {
            if let drilldown {
                ReportTransactionsView(api: api, ownerID: ownerID, masked: masked, scope: drilldown, onBack: { self.drilldown = nil })
            } else {
                ScrollView {
                    VStack(alignment: .leading, spacing: MPMetrics.sectionSpacing) {
                        MPScreenHeader(title: "Báo cáo", leading: {
                            if showsBack { MPIconButton(systemName: "chevron.left", label: "Quay lại", action: onBack) }
                        }, trailing: {
                            MPIconButton(systemName: "arrow.clockwise", label: "Tải lại") { Task { await load() } }
                        })

                        MPSurfaceCard(style: .grouped) {
                            MPReportPeriodPicker(month: $month) { Task { await load() } }
                            MPWalletScopeSelector(wallets: model.wallets, selectionID: $walletSelection)
                        }

                        if let evidenceError { MPStatusMessage(text: evidenceError, tone: .info) }
                        if let report = model.report, report.month == month, report.walletID == walletID {
                            if model.usedFallback {
                                MPStatusMessage(text: "Đang hiển thị bản tổng hợp tương thích; biểu đồ chi tiết cần máy chủ mới.", tone: .info)
                            }
                            summary(report)
                            sectionTitle("Xu hướng", detail: comparisonCaption(report))
                            MPSurfaceCard(style: .grouped) {
                                MPCumulativeReportChart(current: report.daily, previous: report.comparison?.comparable == true ? report.previousDaily : [], masked: masked) { selectedPoint = $0 }
                            }
                            if let selectedPoint {
                                MPStatusMessage(text: masked ? AssistantPrivacyPolicy.maskedText : "\(selectedPoint.label) · Chi \(MPFormat.money(selectedPoint.expense ?? selectedPoint.value))", tone: .info)
                            }

                            sectionTitle("Chi vào đâu?", detail: "Tỷ trọng trên tổng chi của kỳ đang xem")
                            MPSurfaceCard(style: .grouped) {
                                MPReportCategoryDistribution(categories: report.categories, masked: masked) { category in
                                    openTarget(snapshotID: report.snapshotID, targetID: category.targetID, title: category.name)
                                }
                            }

                            sectionTitle("Điều gì thay đổi?", detail: comparisonCaption(report))
                            MPSurfaceCard(style: .grouped) {
                                MPReportContributorChart(contributors: report.comparison?.comparable == true ? report.contributors : [], masked: masked) { item in
                                    openTarget(snapshotID: report.snapshotID, targetID: item.targetID, title: item.name)
                                }
                            }

                            if !report.insights.isEmpty {
                                sectionTitle("Money Insider", detail: "Từ cùng dữ liệu của báo cáo")
                                MPSurfaceCard(style: .grouped) {
                                    VStack(spacing: 0) {
                                        ForEach(Array(report.insights.prefix(3)).indices, id: \.self) { index in
                                            let insight = report.insights[index]
                                            MPReportInsightCard(insight: insight, masked: masked, compact: true, onOpen: {
                                                openTarget(snapshotID: report.snapshotID, targetID: insight.drilldown?.targetID, title: evidenceTitle(insight))
                                            }, onAsk: {
                                                Task { await askInsider(insight, report: report) }
                                            })
                                            if index < min(3, report.insights.count) - 1 { MPInsetDivider() }
                                        }
                                    }
                                }
                            }
                            ReportAnalysisView(api: api, report: report, ownerID: ownerID, masked: masked,
                                walletLabel: walletLabel,
                                onOpenEvidence: { savedSnapshotID, insight in
                                    openTarget(snapshotID: savedSnapshotID, targetID: insight.drilldown?.targetID, title: evidenceTitle(insight))
                                })
                                .id(scopeKey)
                        } else if model.isLoading {
                            MPLoadingView(label: "Đang tải báo cáo…")
                        } else if let error = model.error {
                            MPErrorView(title: "Không tải được báo cáo", detail: error) { Task { await load() } }
                        } else {
                            MPEmptyView(title: "Chưa có báo cáo", detail: "Thêm giao dịch để xem xu hướng và cơ cấu chi tiêu.")
                        }
                    }
                    .padding(MPMetrics.pagePadding)
                }
                .background(MPColor.canvas.ignoresSafeArea())
                .environment(\.timeZone, TimeZone(identifier: accountTimezone) ?? .current)
                .task(id: ownerID) { await model.loadWallets(api: api, ownerID: ownerID) }
                .task(id: scopeKey) {
                    insiderRequest = UUID()
                    selectedPoint = nil; evidenceError = nil; drilldown = nil
                    await load()
                }
                .mpRefreshOnTabActivation(tabIsActive) { await load() }
            }
        }
    }

    @ViewBuilder private func summary(_ report: MonthlyReportModel) -> some View {
        MPFinanceSummaryCard(
            title: "Chênh lệch thu–chi",
            context: displayMonth(report.month),
            primaryLabel: "Thu trừ chi",
            primaryValue: masked ? MPFormat.maskedValue : MPFormat.signedMoney(report.net),
            primaryTone: report.net < 0 ? .expense : .income,
            masked: false,
            accessibilityKey: "reports"
        ) { EmptyView() } content: {
            MPMetricGrid(items: [
                MPMetricItem(title: "Thu", value: MPFormat.money(report.income), masked: masked),
                MPMetricItem(title: "Chi", value: MPFormat.money(report.expense), masked: masked),
                MPMetricItem(title: "Giao dịch", value: "\(report.transactionCount)", masked: false)
            ], masked: false)
        }
    }

    @ViewBuilder private func sectionTitle(_ title: String, detail: String) -> some View {
        VStack(alignment: .leading, spacing: 3) {
            MPText(title, style: .emphasis)
            MPText(detail, style: .secondary)
        }
    }

    private func comparisonCaption(_ report: MonthlyReportModel) -> String {
        guard let comparison = report.comparison, comparison.comparable, comparison.commonDays > 0 else { return "Chưa đủ dữ liệu tương ứng để so sánh" }
        return "So với cùng \(comparison.commonDays) ngày của kỳ trước"
    }

    private func displayMonth(_ value: String) -> String {
        let parts = value.split(separator: "-")
        return parts.count == 2 ? "Tháng \(parts[1])/\(parts[0])" : value
    }

    private var walletLabel: String {
        walletID == nil ? "Tất cả ví" : (model.wallets.first { $0.id == walletID }?.name ?? "Ví đã chọn")
    }

    private func askInsider(_ insight: ReportInsightModel, report: MonthlyReportModel) async {
        let token = UUID()
        insiderRequest = token
        guard let snapshotID = report.snapshotID, let onReportHandoff else {
            evidenceError = "Hãy tải lại báo cáo và mở phân tích sâu để hỏi tiếp theo số liệu đã lưu."
            return
        }
        do {
            // Read only. Insider must never silently generate an analysis or
            // invent a saved finding identity for the Assistant endpoint.
            let latest = try await api.latestReportAnalysis(snapshotID: snapshotID)
            guard insiderRequest == token, !Task.isCancelled else { return }
            guard latest.freshness == "current", let analysis = latest.completed,
                  let savedFinding = analysis.observations.first(where: { $0.id == insight.id }),
                  let handoff = ReportAssistantHandoff.make(analysis: analysis, finding: savedFinding, month: report.month, walletLabel: walletLabel) else {
                evidenceError = "Điểm này chưa có trong bản phân tích hiện hành đã lưu. Hãy mở hoặc cập nhật Phân tích sâu bên dưới để chọn điểm cần hỏi tiếp."
                return
            }
            onReportHandoff(handoff)
        } catch {
            guard insiderRequest == token, !Task.isCancelled else { return }
            evidenceError = "Chưa tải được ngữ cảnh đã lưu. Hãy thử lại khi có kết nối hoặc mở Phân tích sâu."
        }
    }

    private func evidenceTitle(_ insight: ReportInsightModel) -> String {
        let title = masked ? "Giao dịch liên quan" : (insight.title ?? "Giao dịch liên quan")
        guard let period = insight.periodLabel, !period.isEmpty else { return title }
        return title + " · " + period
    }

    private func openTarget(snapshotID: String?, targetID: String?, title: String) {
        if let snapshotID, let targetID, !snapshotID.isEmpty, !targetID.isEmpty {
            drilldown = .monthlyTarget(snapshotID: snapshotID, targetID: targetID, title: title)
        } else {
            evidenceError = "Bản báo cáo này chưa có phạm vi giao dịch được lưu. Hãy tải lại báo cáo và cập nhật phân tích trước khi xem bằng chứng."
        }
    }

    private func load() async {
        await model.load(api: api, month: month, walletID: walletID, ownerID: ownerID, granularity: "day")
        if let timezone = model.report?.timezone { accountTimezone = timezone }
    }
}
