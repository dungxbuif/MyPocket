import SwiftUI

enum MPFinanceCardPolicy {
    static let surfaceStyle: MPSurfaceCardStyle = .reference
    static let contentSpacing = MPMetrics.sectionSpacing
}

/// Domain card wrapper for budgets, reports and specialized wallet surfaces.
/// Feature screens provide content only; spacing and reference surface geometry
/// stay centralized here so a report cannot drift into a one-off card style.
struct MPFinanceCard<Content: View>: View {
    @ViewBuilder let content: () -> Content

    init(@ViewBuilder content: @escaping () -> Content) {
        self.content = content
    }

    var body: some View {
        MPSurfaceCard(style: MPFinanceCardPolicy.surfaceStyle) {
            VStack(alignment: .leading, spacing: MPFinanceCardPolicy.contentSpacing, content: content)
        }
    }
}

/// User prose has a trailing bubble; assistant prose and structured finance
/// cards remain flat so a response never adds a second card around each card.
struct MPChatMessage<Content: View>: View {
    let isUser: Bool
    @ViewBuilder var content: () -> Content

    var body: some View {
        HStack(alignment: .top, spacing: 0) {
            if isUser { Spacer(minLength: 32) }
            VStack(alignment: .leading, spacing: 8, content: content)
                .padding(isUser ? 12 : 0)
                .background(isUser ? MPColor.successSoft : .clear, in: RoundedRectangle(cornerRadius: MPMetrics.cornerRadius))
                .frame(maxWidth: .infinity, alignment: .leading)
            if !isUser { Spacer(minLength: 0) }
        }
        .accessibilityElement(children: .contain)
        .accessibilityLabel(isUser ? "Bạn" : "Trợ lý")
    }
}

struct MPChatOverview: View {
    let income: String
    let expense: String
    let onDetails: () -> Void

    var body: some View {
        Button(action: onDetails) {
            HStack(spacing: 12) {
                VStack(alignment: .leading, spacing: 2) {
                    MPText("Thu nhập", style: .secondary)
                    MPText(income, style: .emphasis)
                }
                Spacer(minLength: 0)
                VStack(alignment: .trailing, spacing: 2) {
                    MPText("Chi tiêu", style: .secondary)
                    MPText(expense, style: .emphasis)
                }
                Image(systemName: "chevron.right").foregroundStyle(MPColor.secondary)
            }
            .padding(.horizontal, 12)
            .padding(.vertical, 8)
            .frame(minHeight: MPMetrics.hitTarget)
            .background(MPColor.surface, in: RoundedRectangle(cornerRadius: MPMetrics.cornerRadius))
        }
        .buttonStyle(.plain)
        .accessibilityLabel("Tổng quan tài chính")
        .accessibilityHint("Mở số liệu và nguồn dữ liệu")
        .accessibilityIdentifier("assistant-pinned-overview")
    }
}

/// Copy and identity rules shared by the native auth surface and the web
/// AuthGateScreen. Keeping these values in the design system prevents the
/// platform implementations from silently drifting apart.
enum AuthReferenceCopy {
    static let title = "MyPocket"
    static let subtitle = "Đăng nhập bằng email và mật khẩu để tiếp tục."
    static let emailPlaceholder = "you@example.com"
    static let passwordPlaceholder = "Nhập mật khẩu"
}

enum AuthReferencePresentation {
    static let cardStyle: MPSurfaceCardStyle = .reference
    static let contentSpacing = MPMetrics.sectionSpacing
}

/// The brand mark is a base component so feature screens do not invent their
/// own logo geometry, color, or accessibility contract.
struct MPBrandMark: View {
    let accessibilityIdentifier: String

    init(accessibilityIdentifier: String = "brand-mark") {
        self.accessibilityIdentifier = accessibilityIdentifier
    }

    var body: some View {
        Text("M")
            .font(MPFont.display)
            .foregroundStyle(MPColor.surface)
            .frame(width: 64, height: 64)
            .background(MPColor.action, in: Circle())
            .accessibilityElement(children: .ignore)
            .accessibilityLabel("MyPocket")
            .accessibilityIdentifier(accessibilityIdentifier)
    }
}

struct MPSectionHeader: View {
    let title: String
    let actionTitle: String?
    let action: (() -> Void)?

    init(title: String, actionTitle: String? = nil, action: (() -> Void)? = nil) {
        self.title = title
        self.actionTitle = actionTitle
        self.action = action
    }

    var body: some View {
        HStack(alignment: .firstTextBaseline, spacing: 8) {
            MPText(title, style: .title)
            Spacer(minLength: 8)
            if let actionTitle, let action {
                MPButton(actionTitle, style: .plain, fullWidth: false, action: action)
                    .font(MPFont.bodySmall.weight(.semibold))
                    .foregroundStyle(MPColor.action)
                    .frame(minHeight: MPMetrics.hitTarget)
            }
        }
    }
}

struct MPChart: View {
    let values: [(label: String, value: Double)]

    var body: some View {
        HStack(alignment: .bottom, spacing: 8) {
            ForEach(Array(values.enumerated()), id: \.offset) { _, item in
                VStack(spacing: 4) {
                    RoundedRectangle(cornerRadius: 4)
                        .fill(MPColor.action)
                        .frame(height: max(4, CGFloat(item.value)))
                    Text(item.label).font(MPFont.labelSmall).foregroundStyle(MPColor.secondary)
                }
                .frame(maxWidth: .infinity)
            }
        }
        .frame(minHeight: 100, alignment: .bottom)
        .accessibilityElement(children: .ignore)
        .accessibilityLabel("Biểu đồ tài chính")
    }
}

/// Shared report summary chart. Heights are normalized from the positive
/// income/expense totals so an empty month and malformed negative values stay
/// finite and render as zero-height bars instead of inventing a minimum spend.
struct MPReportComparisonChart: View {
    static let accessibilityID = "reports-comparison-chart"

    let income: Int64
    let expense: Int64

    static func normalizedHeights(income: Int64, expense: Int64) -> [Double] {
        let safeIncome = max(0, income)
        let safeExpense = max(0, expense)
        let maximum = max(1, max(safeIncome, safeExpense))
        return [safeIncome, safeExpense].map { Double($0) / Double(maximum) * 100 }
    }

    var body: some View {
        let heights = Self.normalizedHeights(income: income, expense: expense)
        HStack(alignment: .bottom, spacing: 8) {
            ForEach(Array(heights.enumerated()), id: \.offset) { index, height in
                GeometryReader { geometry in
                    ZStack(alignment: .bottom) {
                        RoundedRectangle(cornerRadius: 6)
                            .fill(MPColor.line.opacity(0.45))
                        RoundedRectangle(cornerRadius: 6)
                            .fill(index == 0 ? MPColor.action : MPColor.danger)
                            .frame(height: geometry.size.height * CGFloat(height / 100))
                    }
                }
                .frame(maxWidth: .infinity)
            }
        }
        .frame(height: 128, alignment: .bottom)
        .accessibilityElement(children: .ignore)
        .accessibilityLabel("Thu và chi trong tháng")
        .accessibilityValue("Thu \(Int(heights[0].rounded()))%, Chi \(Int(heights[1].rounded()))%")
        .accessibilityIdentifier(Self.accessibilityID)
    }
}

struct MPTrendPoint: Equatable, Identifiable, Sendable {
    let dayKey: String
    let income: Int64
    let expense: Int64

    var id: String { dayKey }
}

enum MPWeekReportData {
    static func totals(
        transactions: [TransactionModel],
        now: Date = Date(),
        timeZone: TimeZone
    ) -> (income: Int64, expense: Int64) {
        let range = MPFormat.weekRange(now: now, timeZone: timeZone)
        return transactions.reduce(into: (income: Int64(0), expense: Int64(0))) { totals, transaction in
            guard transaction.includedInReports,
                  transaction.type == "income" || transaction.type == "expense",
                  let date = MPFormat.parseInstant(transaction.occurredAt),
                  date >= range.start, date < range.endExclusive else { return }
            if transaction.type == "income" { totals.income += transaction.amount }
            else { totals.expense += transaction.amount }
        }
    }
}

enum MPTrendData {
    static func accessibilityValue(points: [MPTrendPoint], masked: Bool) -> String {
        points.map { point in
            "Ngày \(point.dayKey.suffix(2)): Thu \(MPFormat.money(point.income, masked: masked)), Chi \(MPFormat.money(point.expense, masked: masked))"
        }.joined(separator: "; ")
    }

    static func points(transactions: [TransactionModel], month: String, timeZone: TimeZone) -> [MPTrendPoint] {
        var calendar = Calendar(identifier: .gregorian)
        calendar.timeZone = timeZone
        var totals: [String: (income: Int64, expense: Int64)] = [:]
        let dayFormatter = DateFormatter()
        dayFormatter.calendar = calendar
        dayFormatter.locale = Locale(identifier: "en_US_POSIX")
        dayFormatter.timeZone = timeZone
        dayFormatter.dateFormat = "yyyy-MM-dd"

        for transaction in transactions where transaction.includedInReports && (transaction.type == "income" || transaction.type == "expense") {
            guard let date = MPFormat.parseInstant(transaction.occurredAt),
                  MPFormat.accountMonth(of: transaction.occurredAt, timeZone: timeZone) == month else { continue }
            let dayKey = dayFormatter.string(from: date)
            var total = totals[dayKey, default: (income: 0, expense: 0)]
            if transaction.type == "income" {
                total.income += transaction.amount
            } else {
                total.expense += transaction.amount
            }
            totals[dayKey] = total
        }

        return totals.keys.sorted().compactMap { dayKey in
            guard let total = totals[dayKey] else { return nil }
            return MPTrendPoint(dayKey: dayKey, income: total.income, expense: total.expense)
        }
    }
}

struct MPTrendChart: View {
    static let accessibilityID = "reports-trend-chart"

    let points: [MPTrendPoint]
    let masked: Bool

    var body: some View {
        if points.isEmpty {
            MPText("Chưa có giao dịch được tính vào xu hướng.", style: .secondary)
                .frame(maxWidth: .infinity, minHeight: 120, alignment: .center)
                .accessibilityIdentifier(Self.accessibilityID)
        } else {
            VStack(alignment: .leading, spacing: 8) {
                HStack(spacing: 16) {
                    legend(title: "Thu", color: MPColor.action)
                    legend(title: "Chi", color: MPColor.danger)
                }
                GeometryReader { proxy in
                    let width = max(proxy.size.width, 1)
                    let height = max(proxy.size.height, 1)
                    let maximum = max(1, points.map { max($0.income, $0.expense) }.max() ?? 1)
                    ZStack {
                        VStack(spacing: height / 3) {
                            ForEach(0..<4, id: \.self) { _ in
                                Rectangle().fill(MPColor.line.opacity(0.45)).frame(height: 1)
                            }
                        }
                        trendPath(points: points, value: { $0.income }, maximum: maximum, width: width, height: height)
                            .stroke(MPColor.action, style: StrokeStyle(lineWidth: 2, lineCap: .round, lineJoin: .round))
                        trendPath(points: points, value: { $0.expense }, maximum: maximum, width: width, height: height)
                            .stroke(MPColor.danger, style: StrokeStyle(lineWidth: 2, lineCap: .round, lineJoin: .round))
                    }
                }
                .frame(height: 150)
                HStack {
                    ForEach(points, id: \.dayKey) { point in
                        Text(String(point.dayKey.suffix(2)))
                            .font(MPFont.labelSmall)
                            .foregroundStyle(MPColor.secondary)
                            .frame(maxWidth: .infinity)
                    }
                }
                .accessibilityHidden(true)
            }
            .accessibilityElement(children: .ignore)
            .accessibilityLabel("Xu hướng thu chi")
            .accessibilityValue(MPTrendData.accessibilityValue(points: points, masked: masked))
            .accessibilityIdentifier(Self.accessibilityID)
        }
    }

    private func legend(title: String, color: Color) -> some View {
        HStack(spacing: 5) {
            Circle().fill(color).frame(width: 7, height: 7)
            MPText(title, style: .secondary)
        }
    }

    private func trendPath(points: [MPTrendPoint], value: (MPTrendPoint) -> Int64, maximum: Int64, width: CGFloat, height: CGFloat) -> Path {
        Path { path in
            let denominator = CGFloat(max(1, points.count - 1))
            for (index, point) in points.enumerated() {
                let x = CGFloat(index) / denominator * width
                let ratio = CGFloat(value(point)) / CGFloat(maximum)
                let y = height - min(1, max(0, ratio)) * height
                if index == 0 { path.move(to: CGPoint(x: x, y: y)) }
                else { path.addLine(to: CGPoint(x: x, y: y)) }
            }
        }
    }
}

struct MPAssistantCard<Content: View>: View {
    let title: String
    private let content: () -> Content

    static var surfaceStyle: MPSurfaceCardStyle { .elevated }

    init(title: String, @ViewBuilder content: @escaping () -> Content) {
        self.title = title
        self.content = content
    }

    var body: some View {
        MPSurfaceCard(style: Self.surfaceStyle) {
            VStack(alignment: .leading, spacing: 10) {
                Text(title).font(MPFont.subheadline)
                content()
            }
        }
    }
}

struct MPActionCard: View {
    let title: String
    let detail: String
    let actionTitle: String
    let action: () -> Void

    var body: some View {
        MPSurfaceCard {
            VStack(alignment: .leading, spacing: 10) {
                Text(title).font(MPFont.subheadline)
                Text(detail).font(MPFont.bodySmall).foregroundStyle(MPColor.secondary)
                MPButton(actionTitle, action: action)
            }
        }
    }
}

/// Small comparison ring used by the Money Insider card. The percentage is
/// kept unbounded for the center label (so a 214% increase remains honest),
/// while the visual stroke is clamped to the available circular geometry.
struct MPDonutChart: View {
    let percentage: Double?
    let label: String
    let tone: Color

    init(percentage: Double?, label: String, tone: Color = MPColor.action) {
        self.percentage = percentage
        self.label = label
        self.tone = tone
    }

    private var displayValue: String {
        guard let percentage else { return "Chưa có dữ liệu" }
        return String(format: "%.1f%%", percentage)
    }

    private var progress: Double {
        guard let percentage, percentage.isFinite else { return 0 }
        return min(1, max(0, abs(percentage) / 100))
    }

    var body: some View {
        ZStack {
            Circle()
                .stroke(MPColor.line, lineWidth: 8)
            Circle()
                .trim(from: 0, to: progress)
                .stroke(tone, style: StrokeStyle(lineWidth: 8, lineCap: .round))
                .rotationEffect(.degrees(-90))
            MPText(displayValue, style: .emphasis)
                .minimumScaleFactor(0.65)
        }
        .frame(width: 82, height: 82)
        .accessibilityElement(children: .ignore)
        .accessibilityLabel(label)
        .accessibilityValue(displayValue)
        .accessibilityIdentifier("money-insider-comparison-ring")
    }
}

enum MPMoneyInsiderRing {
    static func percentage(changeBPS: Int64?) -> Double? {
        guard let changeBPS else { return nil }
        return Double(changeBPS) / 100.0
    }

    static func tone(changeBPS: Int64?) -> Color {
        guard let changeBPS else { return MPColor.line }
        if changeBPS > 0 { return MPColor.danger }
        if changeBPS < 0 { return MPColor.action }
        return MPColor.secondary
    }

    static func label(changeBPS: Int64?) -> String {
        guard let changeBPS else { return "So sánh chi tiêu chưa có dữ liệu" }
        if changeBPS > 0 { return "Chi tiêu cao hơn kỳ trước" }
        if changeBPS < 0 { return "Chi tiêu thấp hơn kỳ trước" }
        return "Chi tiêu không đổi so với kỳ trước"
    }
}

/// Money Lover-inspired insight surface. The values are always supplied by
/// the real reports endpoint; this component owns only the shared hierarchy.
struct MPMoneyInsiderCard: View {
    let value: InsiderSummaryModel
    let masked: Bool
    let onOpen: () -> Void

    var body: some View {
        MPSurfaceCard(style: .reference) {
            VStack(alignment: .leading, spacing: 12) {
                MPSectionHeader(title: "Money Insider", actionTitle: "↻", action: onOpen)
                HStack(spacing: 8) {
                    MPText(value.topCategories.first?.name ?? "Chi tiêu", style: .title)
                    Image(systemName: "info.circle")
                        .foregroundStyle(MPColor.secondary)
                }
                HStack(alignment: .bottom) {
                    VStack(alignment: .leading, spacing: 4) {
                        MPText("Tổng đã chi", style: .secondary)
                        MPText(masked ? "••••••" : MPFormat.money(value.topCategories.first?.amount ?? value.expense), style: .title)
                            .foregroundStyle(MPColor.danger)
                    }
                    Spacer()
                    MPText(changeText, style: .secondary)
                        .multilineTextAlignment(.trailing)
                }
                Divider().overlay(MPColor.line)
                HStack(alignment: .top, spacing: 16) {
                    VStack(alignment: .leading, spacing: 4) {
                        MPText("Tháng này", style: .emphasis)
                        MPText("Trung bình", style: .secondary)
                        MPText(masked ? "••••••" : "\(MPFormat.money(value.averageDailyExpense))/ngày", style: .emphasis)
                    }
                    Spacer()
                    VStack(alignment: .trailing, spacing: 6) {
                        if value.expenseChangeBPS != nil {
                            MPDonutChart(
                                percentage: MPMoneyInsiderRing.percentage(changeBPS: value.expenseChangeBPS),
                                label: MPMoneyInsiderRing.label(changeBPS: value.expenseChangeBPS),
                                tone: MPMoneyInsiderRing.tone(changeBPS: value.expenseChangeBPS)
                            )
                        }
                        MPText(ratioText, style: .secondary)
                            .multilineTextAlignment(.trailing)
                    }
                }
                HStack(spacing: 10) {
                    MPButton("Xem chi tiết", style: .secondary, fullWidth: false, action: onOpen)
                    MPButton("Mở báo cáo", fullWidth: false, action: onOpen)
                }
            }
        }
    }

    private var changeText: String {
        guard let value = value.expenseChangeBPS else { return "Chưa có kỳ gốc" }
        return "\(value >= 0 ? "+" : "")\(MPFormat.percentageFromBPS(value)) so với kỳ trước"
    }

    private var ratioText: String {
        guard let ratio = value.spendingIncomeRatioBPS else { return "Chưa có thu nhập để tính tỷ lệ" }
        return "Chi/thu \(MPFormat.percentageFromBPS(ratio))"
    }
}

/// Stable copy and formatting rules for the detailed reports Money Insider
/// card. Keeping these decisions next to the shared card prevents the report
/// screen from inventing a second interpretation of the API response.
enum MPReportsInsiderCopy {
    static func change(expenseDelta: Int64, changeBPS: Int64?) -> String {
        guard let changeBPS else { return "Chi tiêu chưa có kỳ gốc để so sánh." }
        if expenseDelta > 0 {
            return "Chi tiêu tăng \(MPFormat.percentageFromBPS(changeBPS)) so với kỳ trước."
        }
        if expenseDelta < 0 {
            return "Chi tiêu giảm \(MPFormat.percentageFromBPS(changeBPS).replacingOccurrences(of: "+", with: "")) so với kỳ trước."
        }
        return "Chi tiêu không đổi so với kỳ trước."
    }

    static func ratio(spendingIncomeRatioBPS: Int64?) -> String {
        guard let spendingIncomeRatioBPS else { return "Chưa có thu nhập để tính tỷ lệ." }
        return "Chi/thu \(MPFormat.percentageFromBPS(spendingIncomeRatioBPS))"
    }

    static func averageDailyExpense(_ value: Int64, days: Int, masked: Bool) -> String {
        let amount = masked ? MPFormat.maskedValue : MPFormat.money(value)
        return "Trung bình \(amount)/ngày trong \(days) ngày."
    }
}

/// Full reports counterpart of the web Money Insider card. It deliberately
/// composes the shared surface/text primitives and renders only values from
/// the owner-scoped `InsiderSummaryModel` response.
struct MPReportsInsiderCard: View {
    let value: InsiderSummaryModel
    let masked: Bool

    var body: some View {
        MPSurfaceCard(style: .reference) {
            VStack(alignment: .leading, spacing: 12) {
                HStack(alignment: .firstTextBaseline, spacing: 8) {
                    MPText("Money Insider", style: .title)
                    Spacer(minLength: 8)
                    MPText(value.estimated ? "Ước tính" : "Từ sổ thật", style: .secondary)
                }
                MPText(
                    "\(MPReportsInsiderCopy.change(expenseDelta: value.expenseDelta, changeBPS: value.expenseChangeBPS)) \(MPReportsInsiderCopy.averageDailyExpense(value.averageDailyExpense, days: value.daysConsidered, masked: masked))",
                    style: .secondary
                )
                HStack(alignment: .top, spacing: 12) {
                    VStack(alignment: .leading, spacing: 4) {
                        MPText("So với kỳ trước", style: .secondary)
                        MPText(MPReportsInsiderCopy.ratio(spendingIncomeRatioBPS: value.spendingIncomeRatioBPS), style: .secondary)
                    }
                    Spacer(minLength: 8)
                    if value.expenseChangeBPS != nil {
                        MPDonutChart(
                            percentage: MPMoneyInsiderRing.percentage(changeBPS: value.expenseChangeBPS),
                            label: MPMoneyInsiderRing.label(changeBPS: value.expenseChangeBPS),
                            tone: MPMoneyInsiderRing.tone(changeBPS: value.expenseChangeBPS)
                        )
                    }
                }

                Divider().overlay(MPColor.line)
                MPText("Top danh mục", style: .emphasis)
                if value.topCategories.isEmpty {
                    MPText("Chưa có khoản chi.", style: .secondary)
                } else {
                    VStack(alignment: .leading, spacing: 8) {
                        ForEach(value.topCategories, id: \.name) { category in
                            HStack(spacing: 12) {
                                MPText(category.name)
                                Spacer(minLength: 8)
                                MPText(MPFormat.money(category.amount, masked: masked), style: .emphasis)
                            }
                        }
                    }
                }

                MPText("Top khoản chi", style: .emphasis)
                if value.topExpenses.isEmpty {
                    MPText("Chưa có khoản chi.", style: .secondary)
                } else {
                    VStack(alignment: .leading, spacing: 8) {
                        ForEach(value.topExpenses) { expense in
                            HStack(spacing: 12) {
                                VStack(alignment: .leading, spacing: 3) {
                                    MPText(expense.categoryName ?? "Khoản chi")
                                    MPText(expense.walletName ?? "Ví", style: .secondary)
                                }
                                Spacer(minLength: 8)
                                MPText(MPFormat.money(expense.amount, masked: masked), style: .emphasis)
                            }
                        }
                    }
                }
            }
        }
        .accessibilityElement(children: .contain)
        .accessibilityIdentifier("reports-section-insider")
    }
}

/// Shared copy rules for the budget summary and per-budget status rows. The
/// backend remains the source of truth for `spent`, `limitAmount` and
/// `daysRemaining`; this helper only keeps web/native wording and safe math in
/// one place.
enum MPBudgetCopy {
    static func timing(ended: Bool, startDate: String, today: String, daysRemaining: Int) -> String {
        if ended { return "Đã kết thúc" }
        if startDate > today { return "Sắp bắt đầu" }
        return "Còn \(max(0, daysRemaining)) ngày"
    }

    static func dailySuggestion(limit: Int64, spent: Int64, daysRemaining: Int, masked: Bool) -> String {
        let remaining = max(0, limit - spent)
        let days = max(1, daysRemaining)
        return "Gợi ý chi/ngày: \(MPFormat.money(remaining / Int64(days), masked: masked))"
    }

    static func threshold(spent: Int64, limit: Int64) -> (text: String, tone: MPStatusTone)? {
        guard limit > 0 else { return nil }
        if spent > limit { return ("Đã vượt ngân sách", .danger) }
        if spent >= limit - limit / 4 { return ("Đã dùng ít nhất 75% ngân sách", .info) }
        return nil
    }
}

enum BudgetReferenceCopy {
    static let addAction = "Thêm"
    static let addActionAccessibilityID = "budget-add"

    static func shouldShowSummary(itemCount: Int) -> Bool {
        itemCount > 0
    }
}

enum CategoriesReferenceCopy {
    static let addAction = "Nhóm mới"
    static let addActionAccessibilityID = "category-add-group"
    static let kindFilterAccessibilityID = "category-kind-filter"
    static let walletFilterAccessibilityID = "category-wallet-filter"
    static let searchAccessibilityID = "category-search"

    static func scopeSubtitle(walletIDs: [String]) -> String {
        walletIDs.isEmpty ? "Áp dụng tất cả ví" : "Hoạt động trong \(walletIDs.count) ví"
    }
}

struct MPCategoryTreeItem: Identifiable, Equatable {
    let id: String
    let title: String
    let subtitle: String
    let systemName: String
    let tone: String
    let depth: Int
    let isSystem: Bool
    let isSelectable: Bool
    let isHierarchyContext: Bool
    let continuesBranch: Bool
}

/// Shared category hierarchy for management lists. The consumer supplies
/// domain labels and callbacks; this base owns card, connector and row layout.
struct MPCategoryTree: View {
    let items: [MPCategoryTreeItem]
    var selectedID: String? = nil
    var groupID: String? = nil
    let onSelect: (String) -> Void

    var body: some View {
        if let groupID {
            card.accessibilityIdentifier(MPCategoryPickerCopy.groupAccessibilityID(groupID))
        } else {
            card
        }
    }

    private var card: some View {
        MPSurfaceCard {
            LazyVStack(spacing: 0) {
                ForEach(Array(items.enumerated()), id: \.element.id) { index, item in
                    if index > 0 { MPDivider() }
                    categoryRow(item)
                }
            }
        }
        .accessibilityElement(children: .contain)
    }

    private func categoryRow(_ item: MPCategoryTreeItem) -> some View {
        MPButton(
            style: .plain,
            isDisabled: selectedID != nil ? !item.isSelectable : item.isHierarchyContext,
            accessibilityIdentifier: selectedID != nil ? MPCategoryPickerCopy.choiceAccessibilityID(item.id) : "category-row-\(item.id)"
        ) {
            onSelect(item.id)
        } label: {
            HStack(spacing: 10) {
                MPTransactionGlyph(systemName: item.systemName, tone: item.tone, size: item.depth == 0 ? 40 : 32)
                VStack(alignment: .leading, spacing: 3) {
                    MPText(item.title, style: item.depth == 0 ? .emphasis : .body)
                        .lineLimit(2)
                    MPText(item.isHierarchyContext ? "Nhóm cha" : item.subtitle, style: .secondary)
                        .lineLimit(2)
                }
                Spacer(minLength: 4)
                if let selectedID {
                    if selectedID == item.id {
                        MPIcon(systemName: "checkmark.circle.fill", label: "Đã chọn", color: MPColor.action, size: 20)
                            .accessibilityHidden(true)
                    }
                } else if !item.isHierarchyContext {
                    HStack(spacing: 6) {
                        if item.isSystem {
                            MPIcon(systemName: "lock.fill", label: "Nhóm hệ thống", color: MPColor.secondary, size: 13)
                                .accessibilityHidden(true)
                        }
                        MPIcon(systemName: "chevron.right", label: "Mở nhóm", color: MPColor.secondary, size: 14)
                            .accessibilityHidden(true)
                    }
                }
            }
            .padding(.leading, CGFloat(max(0, item.depth)) * 20)
            .frame(maxWidth: .infinity, minHeight: MPMetrics.hitTarget, alignment: .leading)
            .background(alignment: .leading) {
                if item.depth > 0 {
                    MPCategoryTreeBranch(continues: item.continuesBranch)
                        .frame(width: 20)
                        .padding(.leading, CGFloat(item.depth - 1) * 20)
                        .accessibilityHidden(true)
                }
            }
        }
        .accessibilityElement(children: .ignore)
        .accessibilityLabel(item.isHierarchyContext ? "Nhóm cha: \(item.title)" : item.title)
        .accessibilityValue(selectedID == item.id ? "Đã chọn" : item.subtitle)
        .accessibilityHint(item.isSystem ? "Danh mục hệ thống" : "")
        .accessibilityAddTraits(selectedID == item.id ? .isSelected : [])
    }
}

enum MPCategoryPickerCopy {
    static let searchAccessibilityID = "category-picker-search"
    static let treeAccessibilityID = "category-picker-tree"
    static let clearAccessibilityID = "category-picker-clear"
    static let clearButtonStyle: MPButtonStyle = .ghost
    static func groupAccessibilityID(_ id: String) -> String { "category-picker-group-" + id }

    static func choiceAccessibilityID(_ id: String) -> String { "category-choice-\(id)" }
}


/// Searchable tree selector shared by transaction and budget forms. The caller
/// supplies candidate categories and the eligible IDs; API ownership and
/// persistence remain outside this component.
struct MPCategoryPicker: View {
    let categories: [CategoryModel]
    let kind: String
    let selectableIDs: Set<String>?
    let selectedID: String
    let clearLabel: String?
    let clearID: String?
    let onSelect: (String) -> Void
    @State private var query = ""

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            MPTextField(
                title: "Tìm kiếm nhóm",
                text: $query,
                placeholder: "Nhập tên nhóm",
                accessibilityIdentifier: MPCategoryPickerCopy.searchAccessibilityID
            )
            if let clearLabel, let clearID {
                MPButton(style: MPCategoryPickerCopy.clearButtonStyle, isDisabled: selectedID == clearID, action: { onSelect(clearID) }) {
                    HStack {
                        Text(clearLabel)
                        Spacer(minLength: 0)
                    }
                }
                .accessibilityIdentifier(MPCategoryPickerCopy.clearAccessibilityID)
            }
            if treeGroups.isEmpty {
                MPEmptyView(title: "Không có nhóm phù hợp", detail: "Thử từ khóa hoặc phạm vi khác.")
            } else {
                VStack(spacing: 12) {
                    ForEach(treeGroups) { group in
                    MPCategoryTree(
                            items: group.nodes.map(treeItem(for:)),
                            selectedID: selectedID,
                            groupID: group.id,
                            onSelect: onSelect
                        )
                    }
                }
                .accessibilityElement(children: .contain)
                .accessibilityIdentifier(MPCategoryPickerCopy.treeAccessibilityID)
            }
        }
    }

    private var treeGroups: [CategorySelectionGroup] {
        CategorySelectionPresentation.groups(
            items: categories,
            kind: kind,
            selectableIDs: selectableIDs,
            search: query
        )
    }

    private func treeItem(for node: CategoryListNode) -> MPCategoryTreeItem {
        let category = node.item
        let glyph = MPTransactionPresentation.glyph(systemKey: category.systemKey ?? category.iconKey, type: category.kind)
        return MPCategoryTreeItem(
            id: category.id,
            title: category.name,
            subtitle: CategorySelectionPresentation.subtitle(isSystem: category.isSystem, isSelectable: node.isSelectable),
            systemName: glyph.systemName,
            tone: glyph.tone,
            depth: node.depth,
            isSystem: category.isSystem,
            isSelectable: node.isSelectable,
            isHierarchyContext: node.isHierarchyContext,
            continuesBranch: node.continuesBranch
        )
    }
}

private struct MPCategoryTreeBranch: View {
    let continues: Bool

    var body: some View {
        GeometryReader { geometry in
            let mid = geometry.size.height / 2
            Path { path in
                path.move(to: CGPoint(x: 7, y: 0))
                path.addLine(to: CGPoint(x: 7, y: mid))
                path.addLine(to: CGPoint(x: max(7, geometry.size.width - 4), y: mid))
                if continues {
                    path.move(to: CGPoint(x: 7, y: mid))
                    path.addLine(to: CGPoint(x: 7, y: geometry.size.height))
                }
            }
            .stroke(MPColor.line, style: StrokeStyle(lineWidth: 2, lineCap: .round, lineJoin: .round))
        }
    }
}

enum JarsReferenceCopy {
    static let addAction = "Thêm hũ"
    static let addActionAccessibilityID = "jar-add-month"
}

enum RecurringReferenceCopy {
    static let addAction = "Thêm"
    static let addActionAccessibilityID = "recurring-add"
}

enum TravelReferenceCopy {
    static let addAction = "Thêm"
    static let addActionAccessibilityID = "travel-add"
}

enum PortfolioReferenceCopy {
    static let subtitle = "Theo dõi mua bán độc lập với số dư ví"
    static let summaryTitle = "Tổng danh mục"
    static let refreshAction = "Làm mới"
    static let refreshActionAccessibilityID = "portfolio-refresh"
    static let addAssetAction = "Thêm tài sản"
    static let addAssetActionAccessibilityID = "portfolio-add-asset"
    static let saveTradeAction = "Lưu giao dịch"
    static let saveTradeActionAccessibilityID = "portfolio-save-trade"
}

/// Stable labels and identity rules for the immutable portfolio correction
/// flow. The editor uses the same overflow/menu and action vocabulary across
/// asset cards; feature code must not invent one-off correction controls.
enum PortfolioCorrectionPresentation {
    static let replaceAction = "Sửa giao dịch"
    static let voidAction = "Hủy giao dịch"
    static let previewAction = "Xem thay đổi"
    static let confirmAction = "Xác nhận sửa"
    static let reasonLabel = "Lý do sửa"
    static let previewTitle = "Xem lại thay đổi"
    static let confirmationMessage = "Lịch sử gốc vẫn được giữ để đối soát."
    static let previewAccessibilityID = "portfolio-correction-preview"
    static let confirmAccessibilityID = "portfolio-correction-confirm"

    static func operationLabel(_ operation: String) -> String {
        operation == "void" ? voidAction : replaceAction
    }

    static func menuAccessibilityID(tradeID: String) -> String {
        "portfolio-trade-menu-\(tradeID)"
    }
}

struct MPOverviewReportCard: View {
    let summary: MonthSummaryModel
    let weekIncome: Int64
    let weekExpense: Int64
    let masked: Bool
    let onOpen: () -> Void
    @State private var period = "month"

    private var values: (income: Int64, expense: Int64) {
        period == "month" ? (summary.income, summary.expense) : (weekIncome, weekExpense)
    }

    var body: some View {
        MPSurfaceCard(style: .reference) {
            VStack(alignment: .leading, spacing: 12) {
                MPSectionHeader(title: "Báo cáo tháng này", actionTitle: "Xem báo cáo", action: onOpen)
                MPSegmentedControl(options: [(id: "week", title: "Tuần"), (id: "month", title: "Tháng")], selection: $period)
                let maxValue = max(max(values.income, values.expense), 1)
                MPChart(values: [(label: "Thu", value: Double(values.income) / Double(maxValue) * 100), (label: "Chi", value: Double(values.expense) / Double(maxValue) * 100)])
                HStack(spacing: 10) {
                    reportMetric(label: "Thu", value: values.income, color: MPColor.action)
                    reportMetric(label: "Chi", value: values.expense, color: MPColor.danger)
                }
            }
        }
    }

    private func reportMetric(label: String, value: Int64, color: Color) -> some View {
        VStack(alignment: .leading, spacing: 4) {
            MPText(label, style: .secondary)
            MPText(masked ? "••••••" : MPFormat.money(value), style: .emphasis).foregroundStyle(color)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(12)
        .background(MPColor.row, in: RoundedRectangle(cornerRadius: MPMetrics.cornerRadius))
    }
}
