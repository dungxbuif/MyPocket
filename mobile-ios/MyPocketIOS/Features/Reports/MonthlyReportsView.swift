import SwiftUI

struct MonthlyReportsView: View {
    let api: FinanceAPI
    let onBack: () -> Void
    let masked: Bool
    let ownerID: String?
    let showsBack: Bool
    @StateObject private var model = MonthlyReportsViewModel()
    @State private var month = MPFormat.currentMonth
    @State private var selectedPoint: ReportPoint?
    @State private var selectedCategory: ReportCategoryTotal?
    @State private var drilldown: ReportTransactionScope?
    @State private var accountTimezone = TimeZone.current.identifier
    let onChatHandoff: (String) -> Void

    init(api: FinanceAPI, onBack: @escaping () -> Void, masked: Bool, ownerID: String?, showsBack: Bool = true, onChatHandoff: @escaping (String) -> Void = { _ in }) {
        self.api = api
        self.onBack = onBack
        self.masked = masked
        self.ownerID = ownerID
        self.showsBack = showsBack
        self.onChatHandoff = onChatHandoff
    }

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
                    MPIconButton(systemName: MPIconCatalog.assetName(for: .report), label: "Tải lại") { Task { await load() } }
                })
                MPText("Tổng hợp theo tháng, không mở sổ giao dịch mặc định.", style: .secondary)
                MPFinanceCard {
                    MPTextField(title: "Tháng (YYYY-MM)", text: $month, accessibilityIdentifier: "reports-month")
                    MPButton("Xem báo cáo", style: .secondary, isLoading: model.isLoading) { Task { await load() } }
                        .accessibilityIdentifier("reports-load-button")
                }
                if let report = model.report {
                    if model.usedFallback {
                        MPStatusMessage(text: "Máy chủ chưa có báo cáo chi tiết; đang hiển thị số liệu tổng hợp tương thích.", tone: .info)
                    }
                    MPFinanceSummaryCard(
                        title: "Tổng quan báo cáo",
                        context: report.month,
                        primaryLabel: "Chênh lệch",
                        primaryValue: masked ? MPFormat.maskedValue : MPFormat.signedMoney(report.net),
                        primaryTone: report.net < 0 ? .expense : .income,
                        masked: false,
                        accessibilityKey: "reports"
                    ) {
                        MPReportComparisonChart(income: report.income, expense: report.expense)
                    } content: {
                        MPMetricGrid(items: [
                            MPMetricItem(title: "Thu nhập", value: MPFormat.money(report.income), masked: masked),
                            MPMetricItem(title: "Chi tiêu", value: MPFormat.money(report.expense), masked: masked),
                            MPMetricItem(title: "Giao dịch", value: "\(report.transactionCount)", masked: false)
                        ], masked: false)
                    }
                    if !report.daily.isEmpty {
                        MPSurfaceCard(style: .grouped) {
                            MPTrendChart(
                                points: report.daily.compactMap { point in
                                    let income = point.comparison ?? 0
                                    guard income != 0 || point.value != 0 else { return nil }
                                    return MPTrendPoint(dayKey: point.label, income: income, expense: point.value)
                                },
                                masked: masked
                            )
                        }
                    } else if !report.weekly.isEmpty {
                        MPSurfaceCard(style: .grouped) {
                            MPReportBarChart(title: "Xu hướng theo tuần", points: report.weekly, masked: masked) { selectedPoint = $0 }
                        }
                    }
                    MPSurfaceCard(style: .grouped) {
                        MPReportCategoryBars(categories: report.categories, masked: masked) { selectedCategory = $0 }
                    }
                    if let selectedPoint {
                        let value = masked ? MPFormat.maskedValue : MPFormat.money(selectedPoint.value)
                        MPStatusMessage(text: selectedPoint.label + ": " + value, tone: .info)
                    }
                    if let selectedCategory {
                        let value = masked ? MPFormat.maskedValue : MPFormat.money(selectedCategory.amount)
                        MPStatusMessage(text: selectedCategory.name + " · " + String(selectedCategory.count) + " giao dịch · " + value, tone: .info)
                        MPButton("Xem giao dịch \(selectedCategory.name)", style: .secondary) {
                            drilldown = .monthly(month: report.month, walletID: nil, categoryID: selectedCategory.categoryID, title: selectedCategory.name)
                        }
                    }
                    if let insight = report.insight, !insight.isEmpty {
                        MPSurfaceCard(style: .grouped) {
                            MPText("Nhận định", style: .emphasis)
                            MPText(masked ? AssistantPrivacyPolicy.maskedText : insight, style: .body)
                        }
                    }
                    ReportAnalysisView(api: api, report: report, ownerID: ownerID, masked: masked, onChatHandoff: onChatHandoff)
                } else if model.isLoading {
                    MPLoadingView(label: "Đang tải báo cáo…")
                } else if let error = model.error {
                    MPErrorView(title: "Không tải được báo cáo", detail: error) { Task { await load() } }
                } else {
                    MPEmptyView(title: "Chọn tháng để xem báo cáo", detail: "Số liệu được tính theo múi giờ tài khoản.")
                }
            }
            .padding(MPMetrics.pagePadding)
        }
        .background(MPColor.canvas.ignoresSafeArea())
        .environment(\.timeZone, TimeZone(identifier: accountTimezone) ?? .current)
        .task { await load() }
        }
        }
    }

    private func load() async {
        let normalized = month.trimmingCharacters(in: .whitespacesAndNewlines)
        guard ReportsMonthPolicy.isValid(normalized) else { return }
        await model.load(api: api, month: normalized, ownerID: ownerID)
    }
}
