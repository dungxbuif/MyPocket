import SwiftUI

enum MPSurfaceCardStyle: Equatable {
    case standard
    case elevated
    case reference

    var cornerRadius: CGFloat {
        self == .reference ? MPMetrics.referenceCornerRadius : MPMetrics.cornerRadius
    }
}

struct MPSurfaceCard<Content: View>: View {
    typealias Style = MPSurfaceCardStyle

    private let content: () -> Content
    private let style: Style

    init(style: Style = .standard, @ViewBuilder content: @escaping () -> Content) {
        self.content = content
        self.style = style
    }

    var body: some View {
        content()
            .padding(.horizontal, MPMetrics.cardHorizontalPadding)
            .padding(.vertical, MPMetrics.cardVerticalPadding)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(MPColor.surface, in: RoundedRectangle(cornerRadius: style.cornerRadius))
            .accessibilityElement(children: .contain)
    }
}

struct MPMetric: View {
    let title: String
    let value: String
    var detail: String?
    var masked: Bool = false
    var compact: Bool = false

    var displayValue: String { masked ? MPFormat.maskedValue : value }
    var spokenValue: String { masked ? "Đã ẩn số tiền" : value }

    var body: some View {
            VStack(alignment: .leading, spacing: 4) {
                Text(title)
                    .font(MPFont.bodySmall)
                    .foregroundStyle(MPColor.secondary)
                Text(displayValue)
                    .font(compact ? MPFont.bodyEmphasis.monospacedDigit() : MPFont.currency)
                    .foregroundStyle(MPColor.ink)
                if let detail {
                    Text(detail)
                        .font(MPFont.bodySmall)
                        .foregroundStyle(MPColor.secondary)
                }
            }
            .fixedSize(horizontal: false, vertical: true)
            .frame(maxWidth: .infinity, alignment: .leading)
            .accessibilityElement(children: .ignore)
            .accessibilityLabel(title)
            .accessibilityValue([spokenValue, detail].compactMap { $0 }.joined(separator: ". "))
    }
}

/// Features supply labels and formatted values; this flat base owns density,
/// reflow and reading order without creating an individual card per metric.
struct MPMetricItem: Identifiable {
    let title: String
    let value: String
    var detail: String? = nil
    var id: String { title }
}

struct MPMetricGrid: View {
    let items: [MPMetricItem]
    var masked: Bool = false
    @Environment(\.dynamicTypeSize) private var dynamicTypeSize

    var body: some View {
        MPMetricGridLayout(accessibility: dynamicTypeSize.isAccessibilitySize) {
            ForEach(items) { item in
                MPMetric(title: item.title, value: item.value, detail: item.detail, masked: masked, compact: true)
            }
        }
        .accessibilityElement(children: .contain)
    }
}

private struct MPMetricGridLayout: Layout {
    let accessibility: Bool

    private func columns(for width: CGFloat) -> Int {
        accessibility || width < MPMetrics.hitTarget * 6 + MPMetrics.sectionSpacing ? 1 : 2
    }

    private func rowHeights(width: CGFloat, subviews: Subviews) -> [CGFloat] {
        let columns = columns(for: width)
        let cellWidth = max(0, (width - CGFloat(columns - 1) * MPMetrics.sectionSpacing) / CGFloat(columns))
        return stride(from: 0, to: subviews.count, by: columns).map { start in
            (start..<min(start + columns, subviews.count)).map { index in
                subviews[index].sizeThatFits(ProposedViewSize(width: cellWidth, height: nil)).height
            }.max() ?? 0
        }
    }

    func sizeThatFits(proposal: ProposedViewSize, subviews: Subviews, cache: inout ()) -> CGSize {
        guard !subviews.isEmpty else { return .zero }
        let width = proposal.width ?? subviews.map { $0.sizeThatFits(.unspecified).width }.max() ?? 0
        let heights = rowHeights(width: width, subviews: subviews)
        return CGSize(width: width, height: heights.reduce(0, +) + CGFloat(max(0, heights.count - 1)) * MPMetrics.sectionSpacing)
    }

    func placeSubviews(in bounds: CGRect, proposal: ProposedViewSize, subviews: Subviews, cache: inout ()) {
        let columns = columns(for: bounds.width)
        let cellWidth = max(0, (bounds.width - CGFloat(columns - 1) * MPMetrics.sectionSpacing) / CGFloat(columns))
        let heights = rowHeights(width: bounds.width, subviews: subviews)
        var y = bounds.minY
        for (row, height) in heights.enumerated() {
            for column in 0..<columns {
                let index = row * columns + column
                guard index < subviews.count else { break }
                subviews[index].place(
                    at: CGPoint(x: bounds.minX + CGFloat(column) * (cellWidth + MPMetrics.sectionSpacing), y: y),
                    anchor: .topLeading,
                    proposal: ProposedViewSize(width: cellWidth, height: height)
                )
            }
            y += height + MPMetrics.sectionSpacing
        }
    }
}

/// Flat compact label/value row used by daily reports and ledger summaries.
/// The full value participates in fitting, so long money never shrinks/clips.
struct MPMetricRow: View {
    let title: String
    let value: String
    var masked: Bool = false
    var tone: MPTransactionAmountTone = .neutral
    @Environment(\.dynamicTypeSize) private var dynamicTypeSize

    var displayValue: String { masked ? MPFormat.maskedValue : value }
    var spokenValue: String { masked ? "Đã ẩn số tiền" : value }

    var body: some View {
        Group {
            if dynamicTypeSize.isAccessibilitySize {
                stacked
            } else {
                ViewThatFits(in: .horizontal) {
                    HStack(alignment: .firstTextBaseline, spacing: MPMetrics.sectionSpacing) {
                        MPText(title, style: .secondary).fixedSize()
                        Spacer(minLength: 0)
                        amount.fixedSize()
                    }
                    stacked
                }
            }
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .accessibilityElement(children: .ignore)
        .accessibilityLabel(title)
        .accessibilityValue(spokenValue)
    }

    private var amount: some View {
        Text(displayValue)
            .font(MPFont.bodyEmphasis.monospacedDigit())
            .foregroundStyle(tone == .neutral ? MPColor.ink : tone.color)
            .fixedSize(horizontal: false, vertical: true)
    }

    private var stacked: some View {
        VStack(alignment: .leading, spacing: MPMetrics.rowSpacing) {
            MPText(title, style: .secondary).fixedSize(horizontal: false, vertical: true)
            amount
        }
    }
}

struct MPMetricCard: View {
    let title: String
    let value: String
    var detail: String?
    var masked: Bool = false

    var body: some View {
        MPSurfaceCard { MPMetric(title: title, value: value, detail: detail, masked: masked) }
    }
}

struct MPProgressBar: View {
    let progress: Double
    let label: String

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            ProgressView(value: min(max(progress, 0), 1))
                .tint(MPColor.action)
                .accessibilityValue("\(Int(progress * 100)) phần trăm")
            Text(label)
                .font(MPFont.bodySmall)
                .foregroundStyle(MPColor.secondary)
        }
    }
}

/// Shared arc gauge for budget summaries. The feature supplies the real
/// spent/limit ratio; this base owns clamping, geometry, color and VoiceOver
/// semantics so the web and native budget surfaces cannot drift.
struct MPBudgetGauge: View {
    static let accessibilityID = "budget-summary-gauge"

    let value: Double
    let label: String

    static func normalizedProgress(value: Double) -> Double {
        guard value.isFinite else { return 0 }
        return min(1, max(0, value / 100))
    }

    private var percent: Int {
        guard value.isFinite else { return 0 }
        return Int(value.rounded())
    }

    var body: some View {
        GeometryReader { proxy in
            let width = max(proxy.size.width, 1)
            let height = max(proxy.size.height, 1)
            ZStack {
                gaugePath(width: width, height: height)
                    .stroke(MPColor.line, style: StrokeStyle(lineWidth: 16, lineCap: .round))
                gaugePath(width: width, height: height)
                    .trim(from: 0, to: Self.normalizedProgress(value: value))
                    .stroke(value > 100 ? MPColor.danger : MPColor.action, style: StrokeStyle(lineWidth: 16, lineCap: .round))
                MPText("\(percent)%", style: .title)
                    .accessibilityHidden(true)
            }
        }
        .frame(height: 130)
        .accessibilityElement(children: .ignore)
        .accessibilityLabel(label)
        .accessibilityValue("\(percent)%")
        .accessibilityIdentifier(Self.accessibilityID)
    }

    private func gaugePath(width: CGFloat, height: CGFloat) -> Path {
        Path { path in
            let inset: CGFloat = 12
            let radius = max(1, min((width - inset * 2) / 2, height - inset))
            let center = CGPoint(x: width / 2, y: height - inset)
            path.addArc(center: center, radius: radius, startAngle: .degrees(180), endAngle: .degrees(360), clockwise: false)
        }
    }
}

/// Shared savings-goal summary used by the wallet ledger and the explicit
/// wallet detail route. Keeping the calculation and typography here prevents
/// the two entry points from drifting while still allowing the reference
/// ledger layout to stay flat (the web/Stitch goal panel is not a generic
/// metric card).
struct MPSavingsSummary: View {
    enum Layout {
        case card
        case ledger
    }

    let balance: Int64
    let target: Int64
    let targetDate: String?
    let timeZone: TimeZone
    let layout: Layout
    var masked: Bool = false

    private var progress: MPSavingsProgress {
        MPFormat.savingsProgress(balance: balance, target: target)
    }

    var body: some View {
        Group {
            if layout == .card {
                MPSurfaceCard(style: .reference) { content }
            } else {
                content
                    .padding(.vertical, 4)
            }
        }
    }

    @ViewBuilder
    private var content: some View {
        if layout == .ledger {
            VStack(spacing: 16) {
                MPSavingsBalanceBlock(balance: balance, masked: masked)
                MPSavingsProgressBlock(progress: progress, deadlineLabel: deadlineLabel, masked: masked)
            }
            .frame(maxWidth: .infinity)
        } else {
            VStack(alignment: .leading, spacing: 8) {
                MPText("Số dư", style: .secondary)
                MPText(MPFormat.money(balance, masked: masked), style: .title)
                MPText(progress.reached ? "ĐÃ ĐẠT MỤC TIÊU" : "CẦN THÊM", style: .secondary)
                MPText(MPFormat.money(progress.remaining, masked: masked), style: .title)
                MPText(deadlineLabel, style: .secondary)
                MPProgressBar(progress: progress.percentage, label: "Tiến độ tiết kiệm")
                MPText("Mục tiêu \(MPFormat.money(target, masked: masked))", style: .secondary)
            }
        }
    }

    private var deadlineLabel: String {
        guard let days = MPFormat.calendarDayDifference(to: targetDate, timeZone: timeZone) else {
            return "Chưa đặt ngày kết thúc"
        }
        if days < 0 { return "Đã qua hạn \(-days) ngày" }
        if days == 0 { return "Đến hạn hôm nay" }
        return "Còn \(days) ngày"
    }
}

/// Base blocks for the embedded Stitch savings ledger. The balance and the
/// target/progress block are separate so the ledger can place the transaction
/// scope tab between them, matching the reference anatomy without duplicating
/// typography or money formatting in the feature screen.
struct MPSavingsBalanceBlock: View {
    let balance: Int64
    var masked: Bool = false

    var body: some View {
        VStack(spacing: 4) {
            MPText("Số dư", style: .secondary)
            MPText(MPFormat.money(balance, masked: masked), style: .display)
        }
    }
}

struct MPSavingsProgressBlock: View {
    let progress: MPSavingsProgress
    let deadlineLabel: String
    var masked: Bool = false

    var body: some View {
        VStack(spacing: 4) {
            MPText(progress.reached ? "ĐÃ ĐẠT MỤC TIÊU" : "Cần thêm", style: .secondary)
            MPText(MPFormat.money(progress.remaining, masked: masked), style: .display)
            MPText(deadlineLabel, style: .body)
            MPProgressBar(progress: progress.percentage, label: "Tiến độ tiết kiệm")
                .accessibilityIdentifier("wallet-detail-savings-progress")
        }
    }
}

enum MPTransactionAmountTone: Equatable {
    case income
    case expense
    case neutral

    var color: Color {
        switch self {
        case .income: return MPColor.action
        case .expense: return MPColor.danger
        case .neutral: return MPColor.secondary
        }
    }
}

struct MPTransactionGlyph: View {
    let systemName: String
    let tone: String
    var size: CGFloat = MPMetrics.hitTarget

    var body: some View {
        Image(systemName: systemName)
            .font(MPFont.bodyEmphasis)
            .foregroundStyle(foreground)
            .frame(width: size, height: size)
            .background(background, in: Circle())
            .overlay(Circle().stroke(foreground.opacity(0.16), lineWidth: MPMetrics.inputBorderWidth))
            .accessibilityHidden(true)
    }

    private var foreground: Color {
        switch tone {
        case "categoryOrange", "categoryAmber": return MPColor.warning
        case "categoryBlue", "categoryIndigo", "categoryCyan": return MPColor.categoryTeal
        case "categoryViolet", "categoryFuchsia": return MPColor.actionHover
        case "categoryRose", "categoryRed", "categoryPink": return MPColor.danger
        case "categorySky", "categoryTeal": return MPColor.categoryTeal
        case "categoryEmerald", "categoryIncome": return MPColor.action
        case "categoryLime": return MPColor.warning
        case "categoryDebt": return MPColor.categoryTeal
        default: return MPColor.secondary
        }
    }

    private var background: Color {
        switch tone {
        case "categoryOrange", "categoryAmber", "categoryLime": return MPColor.warningSoft
        case "categoryBlue", "categoryIndigo", "categoryCyan", "categorySky", "categoryTeal", "categoryDebt": return MPColor.categoryTealSoft
        case "categoryViolet", "categoryFuchsia": return MPColor.successSoft
        case "categoryRose", "categoryRed", "categoryPink": return MPColor.dangerSoft
        case "categoryEmerald", "categoryIncome": return MPColor.successSoft
        default: return MPColor.row
        }
    }
}

/// One semantic presentation for a persisted transaction on every native screen.
/// This is data mapping, not a second row component.
enum MPTransactionPresentation {
    static func rowSubtitle(note: String?, occurredAt: String, timeZone: TimeZone) -> String {
        guard let note = note?.trimmingCharacters(in: .whitespacesAndNewlines), !note.isEmpty else {
            return MPFormat.shortDate(occurredAt, timeZone: timeZone)
        }
        return note
    }

    struct Glyph: Equatable {
        let systemName: String
        let tone: String
    }
    static func glyph(systemKey: String?, type: String) -> Glyph {
        if type == "adjustment" { return Glyph(systemName: "arrow.up.and.down", tone: "categorySlate") }
        let key = systemKey ?? ""
        let presentations: [String: Glyph] = [
            // User-created categories keep their canonical iconKey when no
            // systemKey is available, so the editor and ledger share one
            // visual identity instead of falling back to a generic tag.
            "fork.knife": Glyph(systemName: "fork.knife", tone: "categoryOrange"),
            "leaf.fill": Glyph(systemName: "leaf.fill", tone: "categoryOrange"),
            "cup.and.saucer.fill": Glyph(systemName: "cup.and.saucer.fill", tone: "categoryAmber"),
            "bowl.fill": Glyph(systemName: "bowl.fill", tone: "categoryOrange"),
            "receipt": Glyph(systemName: "receipt", tone: "categoryBlue"),
            "bag.fill": Glyph(systemName: "bag.fill", tone: "categoryViolet"),
            "car.fill": Glyph(systemName: "car.fill", tone: "categorySky"),
            "airplane": Glyph(systemName: "airplane", tone: "categoryCyan"),
            "expense_food": Glyph(systemName: "fork.knife", tone: "categoryOrange"),
            "expense_food_snacks": Glyph(systemName: "leaf.fill", tone: "categoryOrange"),
            "expense_food_coffee": Glyph(systemName: "cup.and.saucer.fill", tone: "categoryAmber"),
            "expense_food_meals": Glyph(systemName: "bowl.fill", tone: "categoryOrange"),
            "expense_bills": Glyph(systemName: "receipt", tone: "categoryBlue"),
            "expense_shopping": Glyph(systemName: "bag.fill", tone: "categoryViolet"),
            "expense_family": Glyph(systemName: "house.fill", tone: "categoryRose"),
            "expense_transport": Glyph(systemName: "car.fill", tone: "categorySky"),
            "expense_health": Glyph(systemName: "heart.text.square.fill", tone: "categoryRed"),
            "expense_education": Glyph(systemName: "graduationcap.fill", tone: "categoryIndigo"),
            "expense_entertainment": Glyph(systemName: "gamecontroller.fill", tone: "categoryFuchsia"),
            "expense_gifts": Glyph(systemName: "gift.fill", tone: "categoryPink"),
            "expense_insurance": Glyph(systemName: "checkmark.shield.fill", tone: "categoryTeal"),
            "expense_investment": Glyph(systemName: "chart.xyaxis.line", tone: "categoryEmerald"),
            "expense_travel": Glyph(systemName: "airplane", tone: "categoryCyan"),
            "expense_savings": Glyph(systemName: "banknote.fill", tone: "categoryLime"),
            "expense_business": Glyph(systemName: "briefcase.fill", tone: "categoryAmber"),
            "expense_transfer_out": Glyph(systemName: "building.columns.fill", tone: "categorySlate"),
            "expense_uncategorized": Glyph(systemName: "tag.fill", tone: "categorySlate"),
            "income_root": Glyph(systemName: "banknote.fill", tone: "categoryIncome"),
            "income_salary": Glyph(systemName: "banknote.fill", tone: "categoryIncome"),
            "income_interest": Glyph(systemName: "banknote.fill", tone: "categoryIncome"),
            "income_gift": Glyph(systemName: "gift.fill", tone: "categoryIncome"),
            "income_bonus": Glyph(systemName: "gift.fill", tone: "categoryIncome"),
            "income_transfer_in": Glyph(systemName: "building.columns.fill", tone: "categoryIncome"),
            "income_uncategorized": Glyph(systemName: "tag.fill", tone: "categoryIncome"),
            "debt_root": Glyph(systemName: "hand.raised.fill", tone: "categoryDebt")
        ]
        if let exact = presentations[key] { return exact }
        let familyPrefixes: [(String, Glyph)] = [
            ("expense_food", Glyph(systemName: "fork.knife", tone: "categoryOrange")),
            ("expense_bills", Glyph(systemName: "receipt", tone: "categoryBlue")),
            ("expense_shopping", Glyph(systemName: "bag.fill", tone: "categoryViolet")),
            ("expense_family", Glyph(systemName: "house.fill", tone: "categoryRose")),
            ("expense_transport", Glyph(systemName: "car.fill", tone: "categorySky")),
            ("expense_health", Glyph(systemName: "heart.text.square.fill", tone: "categoryRed")),
            ("expense_education", Glyph(systemName: "graduationcap.fill", tone: "categoryIndigo")),
            ("expense_entertainment", Glyph(systemName: "gamecontroller.fill", tone: "categoryFuchsia")),
            ("expense_gifts", Glyph(systemName: "gift.fill", tone: "categoryPink")),
            ("expense_insurance", Glyph(systemName: "checkmark.shield.fill", tone: "categoryTeal")),
            ("expense_investment", Glyph(systemName: "chart.xyaxis.line", tone: "categoryEmerald")),
            ("expense_travel", Glyph(systemName: "airplane", tone: "categoryCyan")),
            ("expense_savings", Glyph(systemName: "banknote.fill", tone: "categoryLime")),
            ("expense_business", Glyph(systemName: "briefcase.fill", tone: "categoryAmber")),
            ("debt_", Glyph(systemName: "hand.raised.fill", tone: "categoryDebt"))
        ]
        if let family = familyPrefixes.first(where: { key.hasPrefix($0.0) }) { return family.1 }
        if type == "income" { return Glyph(systemName: "banknote.fill", tone: "categoryIncome") }
        return Glyph(systemName: "tag.fill", tone: "categorySlate")
    }
}

struct MPTransactionRow: View {
    @Environment(\.dynamicTypeSize) private var dynamicTypeSize
    let title: String
    let subtitle: String
    let amount: String
    let iconName: String
    let iconTone: String
    let amountTone: MPTransactionAmountTone

    init(transaction: TransactionModel, category: CategoryModel? = nil, timeZone: TimeZone, masked: Bool) {
        self.init(
            type: transaction.type, rawAmount: transaction.amount,
            categoryName: category?.name, systemKey: category?.systemKey,
            iconKey: category?.iconKey, note: transaction.note,
            occurredAt: transaction.occurredAt, timeZone: timeZone, masked: masked,
            transferID: transaction.transferID,
            adjustmentDirection: transaction.adjustmentDirection,
            creditKind: transaction.creditKind
        )
    }

    /// Also accepts validated structured Assistant results without inventing a
    /// TransactionModel or a screen-specific transaction card.
    init(type: String, rawAmount: Int64, categoryName: String? = nil,
         systemKey: String? = nil, iconKey: String? = nil, note: String? = nil,
         occurredAt: String, timeZone: TimeZone, masked: Bool,
         transferID: String? = nil, adjustmentDirection: String? = nil,
         creditKind: String? = nil) {
        let isTransfer = transferID?.isEmpty == false || type == "transfer"
            || systemKey == "income_transfer_in" || systemKey == "expense_transfer_out"
        let isAdjustment = type == "adjustment"
        let isCreditReturn = creditKind == "refund" || creditKind == "payment"
        let isIncoming = isAdjustment ? adjustmentDirection == "increase" : type == "income" || isCreditReturn
        let categoryTitle = categoryName?.trimmingCharacters(in: .whitespacesAndNewlines)
        let title: String
        if creditKind == "payment" { title = MPSelectionLabels.creditKind(creditKind) }
        else if isTransfer { title = type == "income" ? "Nhận chuyển ví" : type == "expense" ? "Chuyển sang ví khác" : "Chuyển ví" }
        else if isAdjustment { title = "Điều chỉnh số dư" }
        else if isCreditReturn, let creditKind { title = MPSelectionLabels.creditKind(creditKind) }
        else if let categoryTitle, !categoryTitle.isEmpty { title = categoryTitle }
        else if let creditKind { title = MPSelectionLabels.creditKind(creditKind) }
        else { title = type == "income" ? "Khoản thu" : "Khoản chi" }

        let subtitle = MPTransactionPresentation.rowSubtitle(note: note, occurredAt: occurredAt, timeZone: timeZone)
        let glyph: MPTransactionPresentation.Glyph
        if isTransfer { glyph = .init(systemName: "arrow.left.arrow.right", tone: "categorySlate") }
        else if isAdjustment { glyph = .init(systemName: "arrow.up.and.down", tone: "categorySlate") }
        else if systemKey == nil, let iconKey,
                let icon = MPCategoryIconCatalog.options.first(where: { $0.id == iconKey || $0.systemName == iconKey }) {
            glyph = .init(systemName: icon.systemName, tone: icon.tone)
        } else { glyph = MPTransactionPresentation.glyph(systemKey: systemKey ?? iconKey, type: type) }
        self.init(title: title, subtitle: subtitle,
                  amount: masked ? MPFormat.maskedValue : (rawAmount == 0 ? "" : isIncoming ? "+" : "−") + MPFormat.money(rawAmount),
                  iconName: glyph.systemName, iconTone: glyph.tone,
                  amountTone: isTransfer || isAdjustment ? .neutral : isIncoming ? .income : .expense)
    }

    init(
        title: String,
        subtitle: String,
        amount: String,
        iconName: String = "arrow.up.right",
        iconTone: String = "categorySlate",
        amountTone: MPTransactionAmountTone = .expense
    ) {
        self.title = title
        self.subtitle = subtitle
        self.amount = amount
        self.iconName = iconName
        self.iconTone = iconTone
        self.amountTone = amountTone
    }

    init(title: String, subtitle: String, amount: String, isIncome: Bool) {
        self.init(
            title: title,
            subtitle: subtitle,
            amount: amount,
            iconName: isIncome ? "arrow.down.left" : "arrow.up.right",
            iconTone: isIncome ? "categoryIncome" : "categorySlate",
            amountTone: isIncome ? .income : .expense
        )
    }

    var body: some View {
        HStack(spacing: 8) {
            MPTransactionGlyph(systemName: iconName, tone: iconTone, size: 32)
            VStack(alignment: .leading, spacing: 2) {
                Text(title).font(MPFont.bodyEmphasis).lineLimit(2)
                if !subtitle.isEmpty {
                    Text(subtitle).font(MPFont.bodySmall).foregroundStyle(MPColor.secondary).lineLimit(2)
                }
                if dynamicTypeSize.isAccessibilitySize { amountLabel }
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            if !dynamicTypeSize.isAccessibilitySize { amountLabel }
        }
        .frame(minHeight: MPMetrics.hitTarget)
        .accessibilityElement(children: .combine)
    }

    private var amountLabel: some View {
        Text(amount)
            .font(MPFont.body.monospacedDigit())
            .foregroundStyle(amountTone.color)
            .lineLimit(1)
            .minimumScaleFactor(0.85)
            .fixedSize(horizontal: !dynamicTypeSize.isAccessibilitySize, vertical: false)
    }
}

/// Wallet identity badge shared by Overview, wallet lists and selectors. The
/// type-to-tone mapping mirrors the web WalletCard/IconBadge contract so a
/// wallet never changes visual meaning between routes.
struct MPWalletGlyph: View {
    let type: String
    let size: CGFloat

    init(type: String, size: CGFloat = 48) {
        self.type = type
        self.size = size
    }

    static func iconName(for type: String) -> String {
        switch type {
        case "goal": return "banknote.fill"
        case "credit": return "creditcard.fill"
        default: return "wallet.bifold.fill"
        }
    }

    var body: some View {
        MPIcon(systemName: iconName, color: iconForeground, size: size)
            .font(MPFont.bodyEmphasis)
            .background(iconBackground, in: Circle())
            .overlay(Circle().stroke(iconLine, lineWidth: MPMetrics.inputBorderWidth))
    }

    private var iconName: String {
        Self.iconName(for: type)
    }

    private var iconBackground: Color {
        switch type {
        case "goal": return MPColor.successSoft
        case "credit": return MPColor.categoryTealSoft
        default: return MPColor.warningSoft
        }
    }

    private var iconForeground: Color {
        switch type {
        case "goal": return MPColor.action
        case "credit": return MPColor.categoryTeal
        default: return MPColor.warning
        }
    }

    private var iconLine: Color {
        switch type {
        case "goal": return MPColor.successLine
        case "credit": return MPColor.categoryTeal.opacity(0.2)
        default: return MPColor.warning.opacity(0.2)
        }
    }
}

/// Distinct aggregate-scope identity used by the wallet-management total row.
/// The web selector uses a Globe IconBadge in the shared categoryTeal tone;
/// this base keeps the same meaning without misrepresenting the aggregate as
/// an individual basic wallet.
struct MPWalletAggregateGlyph: View {
    static let accessibilityID = "wallet-total-glyph"

    let size: CGFloat

    init(size: CGFloat = 42) {
        self.size = size
    }

    var body: some View {
        ZStack {
            Circle()
                .fill(MPColor.categoryTealSoft)
            Circle()
                .stroke(MPColor.categoryTeal.opacity(0.2), lineWidth: MPMetrics.inputBorderWidth)
            Image(systemName: "globe")
                .font(MPFont.bodyEmphasis)
                .foregroundStyle(MPColor.categoryTeal)
        }
        .frame(width: size, height: size)
        .accessibilityElement(children: .ignore)
        .accessibilityLabel("Ví tổng hợp")
        .accessibilityIdentifier(Self.accessibilityID)
    }
}

/// Small wallet identity badge for the ledger scope capsule. Unlike wallet
/// cards, the basic-wallet selector uses the neutral slate treatment from the
/// web `WalletScopeSelector`; goal and credit keep their semantic accents.
struct MPWalletScopeGlyph: View {
    let type: String?

    static func iconName(for type: String?) -> String {
        guard let type else { return "globe" }
        return MPWalletGlyph.iconName(for: type)
    }

    static func tone(for type: String?) -> String {
        switch type {
        case "goal": return "success"
        case "credit": return "categoryTeal"
        default: return "categorySlate"
        }
    }

    private var foreground: Color {
        switch Self.tone(for: type) {
        case "success": return MPColor.action
        case "categoryTeal": return MPColor.categoryTeal
        default: return MPColor.secondary
        }
    }

    private var background: Color {
        switch Self.tone(for: type) {
        case "success": return MPColor.successSoft
        case "categoryTeal": return MPColor.categoryTealSoft
        default: return MPColor.row
        }
    }

    var body: some View {
        Image(systemName: Self.iconName(for: type))
            .font(MPFont.labelSmall)
            .foregroundStyle(foreground)
            .frame(width: 32, height: 32)
            .background(background, in: Circle())
            .overlay(Circle().stroke(foreground.opacity(0.16), lineWidth: MPMetrics.inputBorderWidth))
            .accessibilityHidden(true)
    }
}

enum MPWalletBalanceTone: Equatable {
    case ink
    case danger
}

enum MPWalletRowPolicy {
    static func subtitle(description: String?, type: String) -> String {
        if let description = description?.trimmingCharacters(in: .whitespacesAndNewlines), !description.isEmpty {
            return description
        }
        return MPSelectionLabels.display(for: type)
    }

    static func balanceTone(isNegative: Bool) -> MPWalletBalanceTone {
        isNegative ? .danger : .ink
    }
}

struct MPWalletRow: View {
    let name: String
    let type: String
    let balance: String
    let subtitle: String?
    let isNegative: Bool

    init(name: String, type: String, balance: String, subtitle: String? = nil, isNegative: Bool = false) {
        self.name = name
        self.type = type
        self.balance = balance
        self.subtitle = subtitle
        self.isNegative = isNegative
    }

    var body: some View {
        HStack(spacing: 8) {
            MPWalletGlyph(type: type, size: 32)
            VStack(alignment: .leading, spacing: 2) {
                Text(name).font(MPFont.bodyEmphasis).foregroundStyle(MPColor.ink)
                Text(MPWalletRowPolicy.subtitle(description: subtitle, type: type)).font(MPFont.bodySmall).foregroundStyle(MPColor.secondary)
            }
            Spacer()
            Text(balance).font(MPFont.bodyEmphasis.monospacedDigit()).foregroundStyle(isNegative ? MPColor.danger : MPColor.ink)
        }
        .frame(minHeight: MPMetrics.hitTarget)
        .accessibilityElement(children: .combine)
    }
}

/// Shared identity surface for the More/account landing screen. The data is
/// deliberately presentation-only: session identity remains owned by the
/// authenticated API/session layer and the action routes to AccountView.
struct MPProfileCard: View {
    let name: String
    let email: String
    let badgeTitle: String?
    let accountSubtitle: String
    let onManage: () -> Void
    let accessibilityIdentifier: String

    init(
        name: String,
        email: String,
        badgeTitle: String? = nil,
        accountSubtitle: String = "Hồ sơ tài khoản",
        accessibilityIdentifier: String = "more-profile-account",
        onManage: @escaping () -> Void
    ) {
        self.name = name
        self.email = email
        self.badgeTitle = badgeTitle
        self.accountSubtitle = accountSubtitle
        self.accessibilityIdentifier = accessibilityIdentifier
        self.onManage = onManage
    }

    private var initial: String {
        String(name.trimmingCharacters(in: .whitespacesAndNewlines).first ?? "M").uppercased()
    }

    var body: some View {
        MPSurfaceCard(style: .reference) {
            if badgeTitle == nil { compactBody }
            else { referenceBody }
        }
    }

    private var identityRow: some View {
        HStack(spacing: 8) {
            Text(initial)
                .font(MPFont.headline)
                .foregroundStyle(.white)
                .frame(width: badgeTitle == nil ? 36 : 40, height: badgeTitle == nil ? 36 : 40)
                .background(MPColor.brand, in: Circle())
                .accessibilityHidden(true)
            VStack(alignment: .leading, spacing: 2) {
                MPText(name, style: .emphasis).lineLimit(1)
                if badgeTitle == nil {
                    HStack(spacing: 4) {
                        MPText(accountSubtitle, style: .secondary).font(MPFont.labelSmall).lineLimit(1)
                        Text("·").font(MPFont.labelSmall).foregroundStyle(MPColor.secondary).accessibilityHidden(true)
                        MPText(email.isEmpty ? "Chưa có email" : email, style: .secondary).font(MPFont.labelSmall).lineLimit(1)
                    }
                } else {
                    MPText(email.isEmpty ? "Chưa có email" : email, style: .secondary).lineLimit(1)
                }
            }
            .frame(maxWidth: .infinity, alignment: .leading)
        }
    }

    private var compactBody: some View {
        MPButton(style: .plain, accessibilityIdentifier: accessibilityIdentifier, action: onManage) {
            HStack {
                identityRow
                MPIcon(systemName: "chevron.right", color: MPColor.secondary).accessibilityHidden(true)
            }
        }
        .accessibilityLabel("Quản lý tài khoản: \(name), \(accountSubtitle), \(email)")
    }

    private var referenceBody: some View {
        VStack(alignment: .leading, spacing: 10) {
            identityRow
            if let badgeTitle {
                Text(badgeTitle)
                    .font(MPFont.labelSmall)
                    .foregroundStyle(.white)
                    .padding(.horizontal, 10)
                    .padding(.vertical, 5)
                    .background(MPColor.warning, in: Capsule())
                    .accessibilityIdentifier("more-profile-badge")
            }
            MPDivider()
            MPButton(style: .plain, accessibilityIdentifier: accessibilityIdentifier, action: onManage) {
                HStack {
                    MPText("Quản lý tài khoản", style: .emphasis)
                    Spacer()
                    MPIcon(systemName: "chevron.right", color: MPColor.secondary).accessibilityHidden(true)
                }
            }
            .accessibilityLabel("Quản lý tài khoản: \(name), \(accountSubtitle), \(email)")
        }
    }
}
