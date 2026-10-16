import SwiftUI

struct SavingsReportsView: View {
    let api: FinanceAPI
    let ownerID: String?
    let masked: Bool
    let onBack: () -> Void
    @State var month: String
    @State private var walletID = "__all__"
    @State private var wallets: [WalletModel] = []
    @State private var report: SavingsReportModel?
    @State private var loading = false
    @State private var error: String?
    @State private var drilldown: ReportTransactionScope?

    var body: some View {
        Group {
            if let drilldown {
                ReportTransactionsView(api: api, ownerID: ownerID, masked: masked, scope: drilldown, onBack: { self.drilldown = nil })
            } else {
                ScrollView {
                    VStack(alignment: .leading, spacing: MPMetrics.sectionSpacing) {
                        MPScreenHeader(title: "Báo cáo tiết kiệm", leading: { MPIconButton(systemName: "chevron.left", label: "Quay lại", action: onBack) }, trailing: { EmptyView() })
                        MPSurfaceCard(style: .reference) {
                            VStack(alignment: .leading, spacing: 8) {
                                MPTextField(title: "Tháng (YYYY-MM)", text: $month, accessibilityIdentifier: "savings-report-month")
                                MPOptionSelector(title: "Ví mục tiêu", options: [MPSelectionOption(id: "__all__", label: "Tất cả ví mục tiêu")] + wallets.map { MPSelectionOption(id: $0.id, label: $0.name) }, selectionID: $walletID, emptyLabel: "Tất cả ví mục tiêu")
                                MPButton("Xem báo cáo", isLoading: loading) { Task { await load() } }
                            }
                            .disabled(loading)
                        }
                        if loading { MPLoadingView(label: "Đang tải báo cáo tiết kiệm...") }
                        if let error { MPErrorView(title: "Không tải được báo cáo tiết kiệm", detail: error) { Task { await load() } } }
                        if let report {
                            MPText("\(report.month) · \(report.timezone) · Cập nhật \(report.asOf)", style: .secondary)
                            MPSurfaceCard(style: .reference) {
                                VStack(alignment: .leading, spacing: 8) {
                                    MPText("Ví mục tiêu", style: .emphasis)
                                    MPMetricGrid(items: goalSummaryMetrics(report), masked: masked)
                                    MPText("Chuyển giữa các ví mục tiêu đang chọn bù trừ nhau. Mục tiêu dưới đây dùng cấu hình hiện tại.", style: .secondary)
                                    MPButton("Giao dịch ví mục tiêu", style: .secondary) { openTransactions(section: "goals") }
                                }
                            }
                            .accessibilityIdentifier("savings-report-goals")
                            if report.goals.items.isEmpty { MPEmptyView(title: "Chưa có ví mục tiêu", detail: "Tạo ví mục tiêu để theo dõi số dư tiết kiệm.") }
                            ForEach(report.goals.items) { goal in
                                MPSurfaceCard(style: .reference) {
                                    VStack(alignment: .leading, spacing: 8) {
                                        MPText(goal.name, style: .emphasis)
                                        MPMetricGrid(items: goalMetrics(goal), masked: masked)
                                        if let progress = goal.progressBPS {
                                            MPProgressBar(progress: Double(progress) / 10_000, label: masked ? MPFormat.maskedValue : String(format: "%.2f%%", Double(progress) / 100))
                                        }
                                        if let date = goal.targetDate { MPText("Ngày mục tiêu hiện tại: \(date)", style: .secondary) }
                                        if goal.reached { MPStatusBadge(text: "Đã đạt mục tiêu", tone: .success) }
                                        MPText("\(goal.transactionCount) giao dịch trong kỳ", style: .secondary)
                                        MPButton("Xem giao dịch", style: .secondary) {
                                            drilldown = .savings(month: report.month, walletID: goal.walletID, section: "goals")
                                        }
                                    }
                                }
                            }
                            MPSurfaceCard(style: .reference) {
                                VStack(alignment: .leading, spacing: 8) {
                                    MPText("Thặng dư thu / chi · Toàn tài khoản", style: .emphasis)
                                    MPMetricGrid(items: surplusMetrics(report), masked: masked)
                                    if report.surplus.savingsRateBPS == nil {
                                        MPText("Chưa có thu nhập dương để tính tỷ lệ.", style: .secondary)
                                    }
                                    MPText("Không cộng thặng dư vào số dư ví mục tiêu. Bộ lọc ví chỉ áp dụng phần mục tiêu.", style: .secondary)
                                    MPButton("Giao dịch thu / chi", style: .secondary) { openTransactions(section: "surplus") }
                                }
                            }
                            .accessibilityIdentifier("savings-report-surplus")
                        }
                    }
                    .padding(MPMetrics.pagePadding)
                }
            }
        }
        .background(MPColor.canvas.ignoresSafeArea())
        .task { await load() }
        .onChange(of: month) { _, _ in report = nil; error = nil }
        .onChange(of: walletID) { _, _ in report = nil; error = nil }
    }

    private func metric(_ title: String, _ value: Int64) -> MPMetricItem {
        MPMetricItem(title: title, value: MPFormat.money(value))
    }

    private func goalSummaryMetrics(_ report: SavingsReportModel) -> [MPMetricItem] {
        let goal = report.goals
        var items = [metric("Số dư đầu tháng", goal.openingBalance), metric("Số dư cuối kỳ", goal.closingBalance)]
        if let initial = goal.initialBalanceAdded, initial != 0 { items.append(metric("Số dư khi tạo ví trong kỳ", initial)) }
        items += [
            metric("Nạp ngoài chuyển ví", goal.deposits), metric("Rút ngoài chuyển ví", goal.withdrawals),
            metric("Điều chỉnh", goal.adjustments), metric("Chuyển ví (thuần)", goal.internalTransfers),
            metric("Thay đổi số dư", goal.netChange)
        ]
        return items
    }

    private func goalMetrics(_ goal: SavingsGoalItemModel) -> [MPMetricItem] {
        var items = [metric("Số dư đầu tháng", goal.openingBalance), metric("Số dư", goal.closingBalance)]
        if let initial = goal.initialBalanceAdded, initial != 0 { items.append(metric("Số dư khi tạo ví trong kỳ", initial)) }
        items += [
            metric("Nạp", goal.deposits), metric("Rút", goal.withdrawals), metric("Điều chỉnh", goal.adjustments),
            metric("Chuyển ví (thuần)", goal.internalTransfers), metric("Thay đổi số dư", goal.netChange)
        ]
        if let target = goal.targetAmount {
            items += [metric("Mục tiêu hiện tại", target), metric("Còn thiếu", goal.remaining)]
        }
        return items
    }

    private func surplusMetrics(_ report: SavingsReportModel) -> [MPMetricItem] {
        var items = [metric("Thu nhập", report.surplus.income), metric("Chi tiêu", report.surplus.expense), metric("Thặng dư", report.surplus.net)]
        if let rate = report.surplus.savingsRateBPS {
            items.append(MPMetricItem(title: "Tỷ lệ thặng dư / thu nhập", value: String(format: "%.2f%%", Double(rate) / 100)))
        }
        return items
    }

    private func openTransactions(section: String) {
        guard let report else { return }
        drilldown = .savings(month: report.month, walletID: section == "goals" && walletID != "__all__" ? walletID : nil, section: section)
    }

    private func load() async {
        guard !loading else { return }
        let requestedMonth = month.trimmingCharacters(in: .whitespacesAndNewlines)
        let requestedWallet = walletID
        report = nil; error = nil
        guard ReportsMonthPolicy.isValid(requestedMonth) else { error = "Nhập tháng hợp lệ theo YYYY-MM."; return }
        loading = true; defer { loading = false }
        do {
            async let result = api.savingsReport(month: requestedMonth, walletID: requestedWallet == "__all__" ? nil : requestedWallet, ownerID: ownerID)
            async let walletResult = api.wallets(ownerID: ownerID)
            let loaded = try await result
            let loadedWallets = try await walletResult
            try Task.checkCancellation()
            guard requestedMonth == month.trimmingCharacters(in: .whitespacesAndNewlines), requestedWallet == walletID else { return }
            wallets = loadedWallets.filter { $0.type == "goal" }
            report = loaded
        } catch is CancellationError { }
        catch { self.error = (error as? APIError)?.userMessage ?? "Báo cáo chưa khả dụng. Thử lại khi máy chủ hỗ trợ tính năng này." }
    }
}

enum ReportTransactionScope: Equatable, Sendable {
    case savings(month: String, walletID: String?, section: String)
    case budget(id: String)

    var title: String {
        switch self {
        case .savings(_, _, let section): section == "goals" ? "Giao dịch ví mục tiêu" : "Thu / chi toàn tài khoản"
        case .budget: "Giao dịch ngân sách"
        }
    }
}

struct ReportTransactionsView: View {
    let api: FinanceAPI
    let ownerID: String?
    let masked: Bool
    let scope: ReportTransactionScope
    let onBack: () -> Void
    @State private var kind = "all"
    @State private var page: ReportTransactionPageModel?
    @State private var categories: [CategoryModel] = []
    @State private var loading = false
    @State private var error: String?
    @State private var failedOffset = 0

    private var kindOptions: [MPSelectionOption] {
        switch scope {
        case .savings(_, _, "goals"):
            [("all", "Tất cả"), ("deposit", "Nạp"), ("withdrawal", "Rút"), ("adjustment", "Điều chỉnh"), ("transfer", "Chuyển ví")].map { MPSelectionOption(id: $0.0, label: $0.1) }
        case .savings:
            [("all", "Tất cả"), ("income", "Thu"), ("expense", "Chi")].map { MPSelectionOption(id: $0.0, label: $0.1) }
        case .budget: []
        }
    }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 8) {
                MPScreenHeader(title: scope.title, leading: { MPIconButton(systemName: "chevron.left", label: "Quay lại", action: onBack) }, trailing: { EmptyView() })
                if !kindOptions.isEmpty {
                    MPOptionSelector(title: "Loại giao dịch", options: kindOptions, selectionID: $kind, emptyLabel: "Tất cả")
                        .disabled(loading)
                }
                if loading { MPLoadingView(label: "Đang tải giao dịch...") }
                if let error { MPErrorView(title: "Không tải được giao dịch", detail: error) { Task { await load(offset: failedOffset) } } }
                if let page {
                    if page.items.isEmpty { MPEmptyView(title: "Không có giao dịch phù hợp", detail: "Phạm vi giống báo cáo đã chọn.") }
                    else {
                        MPSurfaceCard(style: .reference) {
                            VStack(spacing: 8) {
                                ForEach(page.items) { item in
                                    MPTransactionRow(transaction: item, category: categories.first { $0.id == item.categoryID }, timeZone: TimeZone(identifier: page.timezone ?? "") ?? .current, masked: masked)
                                }
                            }
                        }
                    }
                    MPText("\(page.items.isEmpty ? 0 : page.offset + 1)–\(page.offset + page.items.count) / \(page.total) giao dịch", style: .secondary)
                    HStack {
                        MPButton("Trang trước", style: .secondary) { Task { await load(offset: max(0, page.offset - page.limit)) } }
                            .disabled(loading || page.offset == 0)
                        MPButton("Trang tiếp", style: .secondary) { Task { await load(offset: page.offset + page.limit) } }
                            .disabled(loading || page.items.isEmpty || page.offset + page.items.count >= page.total)
                    }
                }
            }
            .padding(MPMetrics.pagePadding)
        }
        .background(MPColor.canvas.ignoresSafeArea())
        .task(id: kind) { page = nil; await load(offset: 0) }
    }

    private func load(offset: Int) async {
        guard !loading else { return }
        let requestedKind = kind
        loading = true; error = nil; failedOffset = offset
        defer { loading = false }
        do {
            let result: ReportTransactionPageModel
            switch scope {
            case .savings(let month, let walletID, let section):
                result = try await api.savingsTransactions(month: month, walletID: walletID, section: section, kind: requestedKind, offset: offset, ownerID: ownerID)
            case .budget(let id): result = try await api.budgetTransactions(id: id, offset: offset, ownerID: ownerID)
            }
            try Task.checkCancellation()
            guard kind == requestedKind else { return }
            page = result
            if categories.isEmpty { categories = (try? await api.categories(ownerID: ownerID)) ?? [] }
        } catch is CancellationError { }
        catch { self.error = (error as? APIError)?.userMessage ?? "Không thể tải giao dịch. Thử lại." }
    }
}
