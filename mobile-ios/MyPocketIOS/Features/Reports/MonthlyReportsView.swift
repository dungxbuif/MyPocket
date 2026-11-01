import SwiftUI

struct MonthlyReportsView: View {
    let api: FinanceAPI
    let onBack: () -> Void
    let masked: Bool
    let ownerID: String?
    let showsBack: Bool
    let onChatHandoff: (String) -> Void
    @StateObject private var model = MonthlyReportsViewModel()
    @State private var month = MPFormat.currentMonth
    @State private var walletSelection = "__all__"
    @State private var selectedPoint: ReportPoint?
    @State private var drilldown: ReportTransactionScope?
    @State private var accountTimezone = TimeZone.current.identifier

    init(api: FinanceAPI, onBack: @escaping () -> Void, masked: Bool, ownerID: String?, showsBack: Bool = true, onChatHandoff: @escaping (String) -> Void = { _ in }) {
        self.api = api; self.onBack = onBack; self.masked = masked; self.ownerID = ownerID
        self.showsBack = showsBack; self.onChatHandoff = onChatHandoff
    }

    private var walletID: String? { walletSelection == "__all__" ? nil : walletSelection }
    private var scopeKey: String { month + ":" + (walletID ?? "all") }

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

                        if let report = model.report {
                            if model.usedFallback {
                                MPStatusMessage(text: "Đang hiển thị bản tổng hợp tương thích; biểu đồ chi tiết cần máy chủ mới.", tone: .info)
                            }
                            summary(report)
                            sectionTitle("Xu hướng", detail: comparisonCaption(report))
                            MPSurfaceCard(style: .grouped) {
                                MPCumulativeReportChart(current: report.daily, previous: report.previousDaily, masked: masked) { selectedPoint = $0 }
                            }
                            if let selectedPoint {
                                MPStatusMessage(text: masked ? AssistantPrivacyPolicy.maskedText : "\(selectedPoint.label) · Chi \(MPFormat.money(selectedPoint.expense ?? selectedPoint.value))", tone: .info)
                            }

                            sectionTitle("Chi vào đâu?", detail: "Tỷ trọng trên tổng chi của kỳ đang xem")
                            MPSurfaceCard(style: .grouped) {
                                MPReportCategoryDistribution(categories: report.categories, masked: masked) { category in
                                    openTarget(snapshotID: report.snapshotID, targetID: category.targetID, legacyCategoryID: category.categoryID, title: category.name)
                                }
                            }

                            sectionTitle("Điều gì thay đổi?", detail: comparisonCaption(report))
                            MPSurfaceCard(style: .grouped) {
                                MPReportContributorChart(contributors: report.contributors, masked: masked) { item in
                                    openTarget(snapshotID: report.snapshotID, targetID: item.targetID, legacyCategoryID: item.categoryID, title: item.name)
                                }
                            }

                            if !report.insights.isEmpty {
                                sectionTitle("Money Insider", detail: "Từ cùng dữ liệu của báo cáo")
                                MPSurfaceCard(style: .grouped) {
                                    VStack(spacing: 0) {
                                        ForEach(Array(report.insights.prefix(3)).indices, id: \.self) { index in
                                            let insight = report.insights[index]
                                            MPReportInsightCard(insight: insight, masked: masked) {
                                                openTarget(snapshotID: report.snapshotID, targetID: insight.drilldown?.targetID, legacyCategoryID: insight.drilldown?.categoryIDs.first, title: insight.title ?? "Chi tiết")
                                            }
                                            if index < min(3, report.insights.count) - 1 { MPInsetDivider() }
                                        }
                                    }
                                }
                            }
                            ReportAnalysisView(api: api, report: report, ownerID: ownerID, masked: masked, onChatHandoff: onChatHandoff)
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
                .task { await model.loadWallets(api: api, ownerID: ownerID); await load() }
                .task(id: scopeKey) { await load() }
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
        guard let comparison = report.comparison, comparison.commonDays > 0 else { return "Chưa đủ ngày hoàn tất để so sánh" }
        return "So với cùng \(comparison.commonDays) ngày của kỳ trước"
    }

    private func displayMonth(_ value: String) -> String {
        let parts = value.split(separator: "-")
        return parts.count == 2 ? "Tháng \(parts[1])/\(parts[0])" : value
    }

    private func openTarget(snapshotID: String?, targetID: String?, legacyCategoryID: String?, title: String) {
        if let snapshotID, let targetID, !snapshotID.isEmpty, !targetID.isEmpty {
            drilldown = .monthlyTarget(snapshotID: snapshotID, targetID: targetID, title: title)
        } else {
            drilldown = .monthly(month: month, walletID: walletID, categoryID: legacyCategoryID, title: title)
        }
    }

    private func load() async {
        await model.load(api: api, month: month, walletID: walletID, ownerID: ownerID, granularity: "day")
        if let timezone = model.report?.timezone { accountTimezone = timezone }
    }
}
