import Charts
import SwiftUI

private struct MPCumulativePoint: Identifiable {
    let id: Int
    let label: String
    let current: Int64
    let previous: Int64?
    let daily: Int64
}

struct MPCumulativeReportChart: View {
    let current: [ReportPoint]
    let previous: [ReportPoint]
    let masked: Bool
    let onSelect: (ReportPoint) -> Void
    @State private var selectedIndex: Int?
    @Environment(\.accessibilityReduceMotion) private var reduceMotion

    private var points: [MPCumulativePoint] {
        var currentTotal: Int64 = 0
        var previousTotal: Int64 = 0
        return current.enumerated().map { index, point in
            currentTotal += point.expense ?? point.value
            let previousValue: Int64?
            if previous.indices.contains(index) {
                previousTotal += previous[index].expense ?? previous[index].value
                previousValue = previousTotal
            } else { previousValue = nil }
            return MPCumulativePoint(id: index, label: point.label, current: currentTotal, previous: previousValue, daily: point.expense ?? point.value)
        }
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 10) {
            HStack(spacing: 16) {
                legend("Kỳ này", color: MPColor.action, dashed: false)
                legend("Kỳ trước", color: MPColor.secondary, dashed: true)
            }
            if masked {
                MPText("Hiện số tiền để xem hình dạng xu hướng.", style: .secondary)
                    .frame(maxWidth: .infinity, minHeight: 180, alignment: .center)
            } else if points.isEmpty {
                MPText("Chưa có dữ liệu xu hướng trong kỳ này.", style: .secondary)
                    .frame(maxWidth: .infinity, minHeight: 180, alignment: .center)
            } else {
                Chart(points) { point in
                    AreaMark(x: .value("Ngày", point.id), y: .value("Chi lũy kế", point.current))
                        .foregroundStyle(LinearGradient(colors: [MPColor.action.opacity(0.20), MPColor.action.opacity(0.02)], startPoint: .top, endPoint: .bottom))
                    LineMark(x: .value("Ngày", point.id), y: .value("Chi lũy kế", point.current), series: .value("Kỳ", "Kỳ này"))
                        .foregroundStyle(MPColor.action)
                        .lineStyle(StrokeStyle(lineWidth: 2.5, lineCap: .round, lineJoin: .round))
                    if let previous = point.previous {
                        LineMark(x: .value("Ngày", point.id), y: .value("Kỳ trước", previous), series: .value("Kỳ", "Kỳ trước"))
                            .foregroundStyle(MPColor.secondary)
                            .lineStyle(StrokeStyle(lineWidth: 1.8, dash: [5, 4]))
                    }
                    if selectedIndex == point.id {
                        RuleMark(x: .value("Ngày chọn", point.id)).foregroundStyle(MPColor.line)
                        PointMark(x: .value("Ngày", point.id), y: .value("Chi lũy kế", point.current)).foregroundStyle(MPColor.action)
                            .annotation(position: .top, spacing: 8) {
                                VStack(alignment: .leading, spacing: 2) {
                                    MPText(reportDayLabel(point.label), style: .secondary)
                                    MPText(MPFormat.money(point.current), style: .emphasis)
                                    MPText("Riêng ngày: \(MPFormat.money(point.daily))", style: .secondary)
                                }
                                .padding(8)
                                .background(MPColor.surface, in: RoundedRectangle(cornerRadius: 10, style: .continuous))
                            }
                    }
                }
                .chartXAxis {
                    AxisMarks(values: .automatic(desiredCount: min(points.count, 5))) { value in
                        AxisGridLine().foregroundStyle(MPColor.line.opacity(0.45))
                        AxisValueLabel {
                            if let index = value.as(Int.self), current.indices.contains(index) { Text(String(current[index].label.suffix(2))) }
                        }
                    }
                }
                .chartYAxis { AxisMarks(position: .leading) { _ in AxisGridLine().foregroundStyle(MPColor.line.opacity(0.45)); AxisValueLabel() } }
                .chartXSelection(value: $selectedIndex)
                .frame(height: 210)
                .animation(reduceMotion ? nil : .easeOut(duration: 0.2), value: current.map(\.value))
                .onChange(of: selectedIndex) { _, value in
                    guard let value, current.indices.contains(value) else { return }
                    onSelect(current[value])
                }
                .accessibilityElement(children: .ignore)
                .accessibilityLabel("Xu hướng chi lũy kế")
                .accessibilityValue(points.map { "\(reportDayLabel($0.label)): \(MPFormat.money($0.current))" }.joined(separator: "; "))
            }
        }
    }

    private func legend(_ title: String, color: Color, dashed: Bool) -> some View {
        HStack(spacing: 6) {
            Capsule().stroke(color, style: StrokeStyle(lineWidth: 2, dash: dashed ? [4, 3] : [])).frame(width: 18, height: 2)
            MPText(title, style: .secondary)
        }
    }
}

struct MPReportBarChart: View {
    let title: String
    let points: [ReportPoint]
    let masked: Bool
    let tone: Color
    let onSelect: (ReportPoint) -> Void
    @State private var selectedLabel: String?

    init(title: String, points: [ReportPoint], masked: Bool, tone: Color = MPColor.action, onSelect: @escaping (ReportPoint) -> Void = { _ in }) {
        self.title = title; self.points = points; self.masked = masked; self.tone = tone; self.onSelect = onSelect
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 10) {
            MPText(title, style: .emphasis)
            if masked {
                MPText("Hiện số tiền để xem biểu đồ.", style: .secondary).frame(maxWidth: .infinity, minHeight: 150)
            } else if points.isEmpty {
                MPText("Chưa có chuỗi dữ liệu cho phạm vi này.", style: .secondary).frame(maxWidth: .infinity, minHeight: 150)
            } else {
                Chart(points) { point in
                    BarMark(x: .value("Kỳ", point.label), y: .value("Chi", max(0, point.expense ?? point.value)))
                        .foregroundStyle(MPColor.danger.opacity(0.82)).position(by: .value("Loại", "Chi"))
                    BarMark(x: .value("Kỳ", point.label), y: .value("Thu", max(0, point.income ?? point.comparison ?? 0)))
                        .foregroundStyle(tone).position(by: .value("Loại", "Thu"))
                }
                .chartXSelection(value: $selectedLabel)
                .chartYAxis { AxisMarks(position: .leading) }
                .frame(height: 170)
                .onChange(of: selectedLabel) { _, label in
                    guard let point = points.first(where: { $0.label == label }) else { return }
                    onSelect(point)
                }
                .accessibilityElement(children: .ignore)
                .accessibilityLabel(title)
                .accessibilityValue(points.map { "\($0.label): Thu \(MPFormat.money($0.income ?? $0.comparison ?? 0)), Chi \(MPFormat.money($0.expense ?? $0.value))" }.joined(separator: "; "))
            }
        }
    }
}

struct MPReportCategoryDistribution: View {
    let categories: [ReportCategoryTotal]
    let masked: Bool
    let onSelect: (ReportCategoryTotal) -> Void
    @State private var selectedAngle: Double?
    private let colors: [Color] = [MPColor.action, MPColor.categoryTeal, MPColor.warning, MPColor.danger, MPColor.secondary, MPColor.line]

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            if masked {
                MPText("Hiện số tiền để xem cơ cấu chi tiêu.", style: .secondary).frame(maxWidth: .infinity, minHeight: 160)
            } else if categories.isEmpty {
                MPText("Chưa có nhóm chi tiêu trong phạm vi này.", style: .secondary).frame(maxWidth: .infinity, minHeight: 160)
            } else {
                Chart(Array(categories.enumerated()), id: \.element.id) { index, category in
                    SectorMark(angle: .value("Chi tiêu", max(0, category.amount)), innerRadius: .ratio(0.62), angularInset: 1.5)
                        .cornerRadius(4).foregroundStyle(colors[index % colors.count])
                }
                .chartAngleSelection(value: $selectedAngle)
                .frame(height: 175)
                .chartBackground { _ in
                    VStack(spacing: 2) {
                        MPText("Tổng chi", style: .secondary)
                        MPText(MPFormat.money(categories.reduce(Int64(0)) { $0 + $1.amount }), style: .emphasis)
                    }
                }
                .onChange(of: selectedAngle) { _, angle in
                    guard let angle else { return }
                    var cursor = 0.0
                    for category in categories {
                        cursor += Double(max(0, category.amount))
                        if angle <= cursor { onSelect(category); return }
                    }
                }
                VStack(spacing: 4) {
                    ForEach(Array(categories.enumerated()), id: \.element.id) { index, category in
                        MPButton(style: .plain, action: { onSelect(category) }) {
                            HStack(spacing: 10) {
                                RoundedRectangle(cornerRadius: 3).fill(colors[index % colors.count]).frame(width: 10, height: 10)
                                MPText(category.name, style: .body)
                                Spacer(minLength: 8)
                                MPText(String(format: "%.1f%%", Double(category.shareBPS) / 100), style: .secondary)
                                MPText(MPFormat.money(category.amount), style: .emphasis)
                            }.frame(minHeight: 44)
                        }
                    }
                }
            }
        }
    }
}

struct MPReportCategoryBars: View {
    let categories: [ReportCategoryTotal]
    let masked: Bool
    let onSelect: (ReportCategoryTotal) -> Void
    var body: some View { MPReportCategoryDistribution(categories: categories, masked: masked, onSelect: onSelect) }
}

struct MPReportContributorChart: View {
    let contributors: [ReportContributorModel]
    let masked: Bool
    let onSelect: (ReportContributorModel) -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 10) {
            if masked { MPText("Hiện số tiền để xem các thay đổi chính.", style: .secondary) }
            else if contributors.isEmpty { MPText("Chưa đủ dữ liệu để so sánh các nhóm.", style: .secondary) }
            else {
                Chart(Array(contributors.prefix(3))) { item in
                    BarMark(x: .value("Thay đổi", item.delta), y: .value("Nhóm", item.name))
                        .foregroundStyle(item.delta >= 0 ? MPColor.danger : MPColor.action)
                        .annotation(position: item.delta >= 0 ? .trailing : .leading) { MPText(MPFormat.signedMoney(item.delta), style: .secondary) }
                }
                .chartXAxis { AxisMarks(position: .bottom); AxisMarks(values: [0]) { _ in AxisGridLine().foregroundStyle(MPColor.secondary) } }
                .frame(height: CGFloat(max(1, min(3, contributors.count))) * 58)
                ForEach(Array(contributors.prefix(3))) { item in
                    MPButton(style: .plain, action: { onSelect(item) }) {
                        HStack { MPText(item.name, style: .body); Spacer(); MPText(item.delta >= 0 ? "Tăng" : "Giảm", style: .secondary) }.frame(minHeight: 44)
                    }
                }
            }
        }
    }
}

private func reportDayLabel(_ value: String) -> String {
    let parts = value.prefix(10).split(separator: "-")
    guard parts.count == 3 else { return value }
    return "\(parts[2])/\(parts[1])"
}

struct MPReportPeriodPicker: View {
    @Binding var month: String
    let onChange: () -> Void

    var body: some View {
        HStack(spacing: 8) {
            MPHeaderIconButton(systemName: "chevron.left", label: "Tháng trước") { move(-1) }
            Spacer(minLength: 4)
            VStack(spacing: 2) {
                MPText("Kỳ báo cáo", style: .secondary)
                MPText(displayMonth, style: .emphasis)
            }
            Spacer(minLength: 4)
            MPHeaderIconButton(systemName: "chevron.right", label: "Tháng sau") { move(1) }
                .disabled(month >= MPFormat.currentMonth)
        }
        .frame(minHeight: 48)
    }

    private var displayMonth: String {
        let parts = month.split(separator: "-")
        guard parts.count == 2 else { return month }
        return "Tháng \(parts[1])/\(parts[0])"
    }

    private func move(_ value: Int) {
        let parts = month.split(separator: "-").compactMap { Int($0) }
        guard parts.count == 2 else { return }
        var components = DateComponents(); components.year = parts[0]; components.month = parts[1]; components.day = 1
        guard let date = Calendar(identifier: .gregorian).date(from: components), let next = Calendar(identifier: .gregorian).date(byAdding: .month, value: value, to: date) else { return }
        let formatter = DateFormatter(); formatter.locale = Locale(identifier: "en_US_POSIX"); formatter.dateFormat = "yyyy-MM"
        month = min(formatter.string(from: next), MPFormat.currentMonth)
        onChange()
    }
}

struct MPReportInsightCard: View {
    let insight: ReportInsightModel
    let masked: Bool
    let onOpen: () -> Void

    var body: some View {
        MPButton(style: .plain, action: onOpen) {
            VStack(alignment: .leading, spacing: 6) {
                MPText(insight.title ?? "Điểm đáng chú ý", style: .emphasis)
                MPText(masked ? AssistantPrivacyPolicy.maskedText : insight.body, style: .body)
                MPText("Xem số liệu", style: .action)
            }
            .frame(maxWidth: .infinity, minHeight: 44, alignment: .leading)
        }
    }
}

struct MPReportAnalysisArticle: View {
    let analysis: ReportAnalysisModel
    let masked: Bool

    var body: some View {
        VStack(alignment: .leading, spacing: 10) {
            if analysis.retainedUntil != nil {
                MPText("Bản phân tích đã lưu", style: .secondary)
            }
            if masked {
                MPText(AssistantPrivacyPolicy.maskedText, style: .body)
            } else {
                if let summary = analysis.summary, !summary.isEmpty { MPText(summary, style: .body) }
                ForEach(analysis.observations) { item in
                    VStack(alignment: .leading, spacing: 3) {
                        MPText(item.title ?? "Quan sát", style: .emphasis)
                        MPText(item.body, style: .body)
                    }
                }
                if !analysis.suggestions.isEmpty {
                    MPText("Gợi ý theo dõi", style: .emphasis)
                    ForEach(analysis.suggestions) { item in MPText("• " + item.body, style: .body) }
                }
            }
        }
    }
}

struct MPReportInsiderCard: View {
    let insights: [ReportInsightModel]
    let masked: Bool
    let onOpen: () -> Void

    var body: some View {
        MPSurfaceCard(style: .reference) {
            VStack(alignment: .leading, spacing: 10) {
                MPSectionHeader(title: "Money Insider", actionTitle: "Xem báo cáo", action: onOpen)
                if let first = insights.first {
                    MPText(first.title ?? "Điểm đáng chú ý", style: .title)
                    MPText(masked ? AssistantPrivacyPolicy.maskedText : first.body, style: .body)
                    ForEach(Array(insights.dropFirst().prefix(2))) { insight in
                        MPInsetDivider()
                        MPText(insight.title ?? insight.body, style: .emphasis)
                    }
                } else {
                    MPText("Chưa có dữ liệu đủ để đưa ra nhận xét.", style: .secondary)
                }
                MPButton("Mở báo cáo", style: .secondary, fullWidth: false, action: onOpen)
            }
        }
    }
}
