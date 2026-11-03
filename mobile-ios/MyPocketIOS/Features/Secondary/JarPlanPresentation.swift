import SwiftUI

/// Pure presentation rules for the jar plan figures (JAR-12/13/14). A jar's headline number is the
/// cumulative plan remaining, never a "balance": it is not money held in a wallet.
enum JarPlanPresentation {
    static func headlineTitle(_ item: JarItemModel) -> String {
        guard let remaining = item.remainingPlan else { return "Đã chi cộng dồn" }
        return remaining < 0 ? "Vượt kế hoạch" : "Còn có thể chi"
    }

    /// A negative plan is shown as the overrun amount under "Vượt kế hoạch" rather than a negative balance.
    static func headlineAmount(_ item: JarItemModel) -> Int64 {
        if let remaining = item.remainingPlan { return abs(remaining) }
        return item.cumulativeSpent ?? item.spent
    }

    /// "Tháng này 10.000.000 / 15.000.000 đ" or the no-allocation variant (JAR-08).
    static func monthLine(_ item: JarItemModel, masked: Bool) -> String {
        if let allocation = item.calculatedAllocation {
            return "Tháng này \(MPFormat.money(item.spent, masked: masked)) / \(MPFormat.money(allocation, masked: masked))"
        }
        return "Chưa đặt mức · đã chi \(MPFormat.money(item.spent, masked: masked)) tháng này"
    }

    static func monthProgress(_ item: JarItemModel) -> Double? {
        JarsReferencePresentation.progress(spent: item.spent, allocation: item.calculatedAllocation)
    }

    static func isOverMonth(_ item: JarItemModel) -> Bool {
        JarsReferencePresentation.isOverspent(spent: item.spent, allocation: item.calculatedAllocation)
    }

    /// Plan left before this month's allocation and spending: `remaining − (allocation − spent)`.
    static func carriedOver(_ item: JarItemModel) -> Int64? {
        guard let remaining = item.remainingPlan else { return nil }
        return remaining - ((item.calculatedAllocation ?? 0) - item.spent)
    }

    /// The owner's caption: planned this month, carried over, and what can still be spent by plan.
    static func note(_ item: JarItemModel, masked: Bool) -> String {
        guard let remaining = item.remainingPlan else {
            return "Tháng này chưa đặt mức; chi vẫn trừ vào kế hoạch còn lại khi có mức."
        }
        var parts: [String] = []
        if let allocation = item.calculatedAllocation {
            parts.append("Kế hoạch tháng này: \(MPFormat.money(allocation, masked: masked))")
        } else {
            parts.append("Tháng này chưa đặt mức")
        }
        if let carried = carriedOver(item), carried != 0 {
            parts.append("Dư từ tháng trước: \(MPFormat.signedMoney(carried, masked: masked))")
        }
        parts.append("Còn có thể chi theo kế hoạch: \(MPFormat.money(remaining, masked: masked))")
        return parts.joined(separator: " · ")
    }

    struct Reconciliation: Equatable {
        let text: String
        let exceeds: Bool
    }

    /// JAR-13: plan remaining versus the real money in wallets that count toward the total.
    static func reconciliation(_ summary: JarMonthSummaryModel, masked: Bool) -> Reconciliation? {
        guard let plan = summary.remainingPlanTotal, let wallets = summary.walletsInTotalBalance,
              summary.items.contains(where: { $0.remainingPlan != nil }) else { return nil }
        if let exceeds = summary.planExceedsWalletsBy, exceeds > 0, wallets < 0 {
            return Reconciliation(
                text: "Kế hoạch còn lại \(MPFormat.money(plan, masked: masked)) trong khi các ví tính vào tổng đang âm \(MPFormat.money(-wallets, masked: masked)). Đây là cảnh báo, không chặn giao dịch.",
                exceeds: true
            )
        }
        if let exceeds = summary.planExceedsWalletsBy, exceeds > 0 {
            return Reconciliation(
                text: "Kế hoạch còn lại \(MPFormat.money(plan, masked: masked)) vượt tiền thực có \(MPFormat.money(wallets, masked: masked)) là \(MPFormat.money(exceeds, masked: masked)). Đây là cảnh báo, không chặn giao dịch.",
                exceeds: true
            )
        }
        return Reconciliation(
            text: "Kế hoạch còn lại của các hũ: \(MPFormat.money(plan, masked: masked)) · Tiền trong ví: \(MPFormat.money(wallets, masked: masked))",
            exceeds: false
        )
    }

    static func unassignedLine(_ summary: JarMonthSummaryModel, masked: Bool) -> String {
        if let cumulative = summary.unassignedCumulative {
            return "Cộng dồn: \(MPFormat.money(cumulative, masked: masked))"
        }
        return "Chi thường chưa gắn hũ"
    }
}

/// Jar filter of the transaction ledger (Tất cả · each jar · Chưa gắn hũ) and the editor's informational jar hint.
enum JarLedgerFilter {
    static let all = "__all__"
    static let unassigned = "unassigned"

    /// "Unassigned" is an ordinary expense without a jar (JAR-14); income, transfers, credit rows and adjustments never match it.
    static func isOrdinaryExpense(_ transaction: TransactionModel) -> Bool {
        transaction.type == "expense" && transaction.transferID == nil && transaction.creditKind == nil
    }

    static func matches(_ transaction: TransactionModel, selection: String) -> Bool {
        switch selection {
        case all: return true
        case unassigned: return isOrdinaryExpense(transaction) && (transaction.jarID ?? "").isEmpty
        default: return transaction.jarID == selection
        }
    }

    /// Jars of the viewed month plus any jar still referenced by loaded rows, so a filter never hides a used jar.
    static func options(summary: JarMonthSummaryModel?, transactions: [TransactionModel]) -> [MPSelectionOption] {
        var seen = Set<String>()
        var jars: [MPSelectionOption] = []
        for jar in summary?.jars ?? [] where seen.insert(jar.jarID).inserted {
            jars.append(MPSelectionOption(id: jar.jarID, label: jar.name))
        }
        for row in transactions {
            guard let id = row.jarID, !id.isEmpty, seen.insert(id).inserted else { continue }
            jars.append(MPSelectionOption(id: id, label: row.jarName ?? "Hũ đã gỡ"))
        }
        return [MPSelectionOption(id: all, label: "Tất cả")] + jars + [MPSelectionOption(id: unassigned, label: "Chưa gắn hũ")]
    }

    /// Jar name chip shown on ordinary expense rows only.
    static func chip(_ transaction: TransactionModel) -> String? {
        guard isOrdinaryExpense(transaction), let name = transaction.jarName, !name.isEmpty else { return nil }
        return name
    }
}

extension JarPlanPresentation {
    struct EditorHint: Equatable {
        let text: String
        let exceeds: Bool
    }

    /// Informational line under the editor's jar selector; never blocks saving (JAR-05).
    /// `existingAmount` is this transaction's own amount when it is already counted in the jar's spent.
    static func editorHint(summary: JarMonthSummaryModel?, jarID: String, amount: Int64, existingAmount: Int64 = 0) -> EditorHint? {
        guard let item = summary?.items.first(where: { $0.jarID == jarID && $0.active }) else { return nil }
        let delta = max(amount, 0) - existingAmount
        if let allocation = item.calculatedAllocation, allocation - item.spent - delta < 0 {
            let over = item.spent + delta - allocation
            return EditorHint(text: "Hũ \(item.name) sẽ vượt mức tháng \(MPFormat.money(over))", exceeds: true)
        }
        guard let remaining = item.remainingPlan else {
            return EditorHint(text: "Hũ \(item.name) chưa đặt mức tháng này", exceeds: false)
        }
        let after = remaining - delta
        if after < 0 {
            return EditorHint(text: "Hũ \(item.name) sẽ vượt kế hoạch \(MPFormat.money(-after))", exceeds: true)
        }
        return EditorHint(text: "Hũ \(item.name) còn có thể chi \(MPFormat.money(after))", exceeds: false)
    }
}

/// Where a jar detail screen reads from: one jar or the unassigned bucket.
enum JarDetailTarget: Equatable {
    case jar(JarItemModel)
    case unassigned

    static func == (lhs: JarDetailTarget, rhs: JarDetailTarget) -> Bool {
        switch (lhs, rhs) {
        case (.unassigned, .unassigned): return true
        case (.jar(let a), .jar(let b)): return a.jarID == b.jarID
        default: return false
        }
    }
}

struct JarDetailView: View {
    let api: FinanceAPI
    let target: JarDetailTarget
    let summary: JarMonthSummaryModel
    let month: String
    let timeZone: TimeZone
    let masked: Bool
    let onBack: () -> Void
    let onEdit: (JarItemModel) -> Void
    let onRemove: (JarItemModel) -> Void

    @State private var report: JarCumulativeSummaryModel?
    @State private var transactions: [TransactionModel] = []
    @State private var categories: [CategoryModel] = []
    @State private var loading = false
    @State private var error: String?

    private var jarID: String {
        switch target {
        case .jar(let item): return item.jarID
        case .unassigned: return "unassigned"
        }
    }

    private var item: JarItemModel? {
        if case .jar(let value) = target {
            // Prefer the freshest summary row for this jar so edits show without reopening.
            return summary.items.first(where: { $0.jarID == value.jarID }) ?? value
        }
        return nil
    }

    private var monthTransactions: [TransactionModel] {
        var calendar = Calendar(identifier: .gregorian)
        calendar.timeZone = timeZone
        return transactions.filter { row in
            guard let date = MPFormat.parseInstant(row.occurredAt) else { return false }
            let parts = calendar.dateComponents([.year, .month], from: date)
            return String(format: "%04d-%02d", parts.year ?? 0, parts.month ?? 0) == month
        }
    }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: JarsReferencePresentation.sectionSpacing) {
                MPScreenHeader(
                    title: item?.name ?? "Chưa phân bổ",
                    leading: { MPIconButton(systemName: "chevron.left", label: "Quay lại", action: onBack) },
                    trailing: {
                        if let item, item.active {
                            MPOverflowMenu(label: "Tuỳ chọn hũ", actions: [
                                MPOverflowAction(title: "Sửa mức tháng này") { onEdit(item) },
                                MPOverflowAction(title: "Gỡ khỏi tháng") { onRemove(item) }
                            ])
                        } else { EmptyView() }
                    }
                )
                MPText(MPFormat.monthLabel(month), style: .secondary)
                if let item {
                    MPFinanceCard {
                        MPMetricGrid(items: [
                            MPMetricItem(title: JarPlanPresentation.headlineTitle(item), value: MPFormat.money(JarPlanPresentation.headlineAmount(item)), masked: masked),
                            MPMetricItem(title: "Mức tháng này", value: item.calculatedAllocation.map { MPFormat.money($0) } ?? "Chưa đặt", masked: masked),
                            MPMetricItem(title: "Đã chi tháng này", value: MPFormat.money(item.spent), masked: masked)
                        ], masked: false)
                        MPText(JarPlanPresentation.note(item, masked: masked), style: .secondary)
                            .accessibilityIdentifier("jar-detail-note")
                        if JarPlanPresentation.isOverMonth(item) {
                            MPStatusMessage(text: "Chi vượt mức tháng này. Đây là cảnh báo, không chặn giao dịch.", tone: .danger)
                        }
                    }
                    .accessibilityIdentifier("jar-detail-summary")
                } else {
                    MPFinanceCard {
                        MPMetricGrid(items: [
                            MPMetricItem(title: "Chưa gắn hũ tháng này", value: MPFormat.money(summary.unassignedSpent), masked: masked),
                            MPMetricItem(title: "Cộng dồn", value: MPFormat.money(summary.unassignedCumulative ?? summary.unassignedSpent), masked: masked)
                        ], masked: false)
                        MPText("Khoản chi thường chưa gắn hũ vẫn hợp lệ và được tính trong chi tháng.", style: .secondary)
                    }
                    .accessibilityIdentifier("jar-detail-summary")
                }
                if let report, item != nil {
                    MPFinanceCard {
                        MPText("Diễn biến theo tháng", style: .emphasis)
                        MPText("Mức có ở \(report.allocationMonths)/\(report.monthsInRange) tháng. Kế hoạch còn lại không phải số dư ví.", style: .secondary)
                        ForEach(report.months, id: \.month) { row in
                            HStack {
                                MPText(MPFormat.monthLabel(row.month))
                                Spacer()
                                if let allocation = row.calculatedAllocation {
                                    MPText("\(MPFormat.money(row.spent, masked: masked)) / \(MPFormat.money(allocation, masked: masked))", style: .secondary)
                                } else {
                                    MPText("\(MPFormat.money(row.spent, masked: masked)) · không đặt mức", style: .secondary)
                                }
                            }
                        }
                    }
                }
                MPText("Giao dịch tháng này", style: .title)
                if loading && transactions.isEmpty { MPLoadingView(label: "Đang tải giao dịch...") }
                if let error {
                    MPStatusMessage(text: error, tone: .danger)
                    MPButton("Thử lại", style: .plain, fullWidth: false) { Task { await load() } }
                }
                if monthTransactions.isEmpty, !loading {
                    MPEmptyView(title: "Chưa có giao dịch", detail: "Chưa có khoản chi nào trong tháng này.")
                } else {
                    MPSurfaceCard(style: .grouped) {
                        VStack(alignment: .leading, spacing: 0) {
                            ForEach(Array(monthTransactions.enumerated()), id: \.element.id) { index, row in
                                if index > 0 { MPInsetDivider() }
                                MPGroupedRow {
                                    MPTransactionRow(
                                        transaction: row,
                                        category: categories.first(where: { $0.id == row.categoryID }),
                                        timeZone: timeZone,
                                        masked: masked
                                    )
                                }
                            }
                        }
                    }
                    .accessibilityIdentifier("jar-detail-transactions")
                }
            }
            .padding(MPMetrics.pagePadding)
        }
        .task { await load() }
        .onChange(of: month) { _, _ in Task { await load() } }
    }

    private func load() async {
        guard !loading else { return }
        loading = true; defer { loading = false }
        do {
            async let rows = api.transactions(jarID: jarID)
            async let cats = api.categories()
            transactions = try await rows.sorted { $0.occurredAt > $1.occurredAt }
            categories = (try? await cats) ?? categories
            if case .jar(let value) = target {
                report = try? await api.jarReport(jarID: value.jarID, to: month)
            }
            error = nil
        } catch { self.error = "Không thể tải giao dịch của hũ." }
    }
}
