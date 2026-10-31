import SwiftUI

struct MPReportBarChart: View {
    let title: String
    let points: [ReportPoint]
    let masked: Bool
    let tone: Color
    let onSelect: (ReportPoint) -> Void

    init(title: String, points: [ReportPoint], masked: Bool, tone: Color = MPColor.action, onSelect: @escaping (ReportPoint) -> Void = { _ in }) {
        self.title = title; self.points = points; self.masked = masked; self.tone = tone; self.onSelect = onSelect
    }

    var body: some View {
        VStack(alignment: .leading, spacing: MPMetrics.rowSpacing) {
            MPText(title, style: .emphasis)
            if points.isEmpty {
                MPText("Chưa có chuỗi dữ liệu cho phạm vi này.", style: .secondary)
            } else {
                GeometryReader { geometry in
                    let maximum = max(points.map(\.value).max() ?? 1, 1)
                    HStack(alignment: .bottom, spacing: max(4, geometry.size.width / CGFloat(max(points.count * 5, 1)))) {
                        ForEach(points) { point in
                            MPButton(style: .plain, fullWidth: false, action: { onSelect(point) }) {
                                VStack(spacing: 4) {
                                    RoundedRectangle(cornerRadius: 4)
                                        .fill(tone)
                                        .frame(height: max(6, geometry.size.height * CGFloat(point.value) / CGFloat(maximum)))
                                    Text(point.label)
                                        .font(MPFont.labelSmall)
                                        .foregroundStyle(MPColor.secondary)
                                        .lineLimit(1)
                                }
                                .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .bottom)
                            }
                            .accessibilityLabel(point.label)
                            .accessibilityValue(masked ? MPFormat.maskedValue : MPFormat.money(point.value))
                        }
                    }
                }
                .frame(minHeight: 128)
            }
        }
    }
}

struct MPReportCategoryBars: View {
    let categories: [ReportCategoryTotal]
    let masked: Bool
    let onSelect: (ReportCategoryTotal) -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: MPMetrics.rowSpacing) {
            MPText("Theo nhóm", style: .emphasis)
            if categories.isEmpty {
                MPText("Chưa có nhóm chi tiêu trong phạm vi này.", style: .secondary)
            } else {
                let maximum = max(categories.map(\.amount).max() ?? 1, 1)
                ForEach(categories) { category in
                    MPButton(style: .plain, action: { onSelect(category) }) {
                        VStack(alignment: .leading, spacing: 4) {
                            HStack {
                                MPText(category.name, style: .body)
                                Spacer(minLength: MPMetrics.rowSpacing)
                                MPText(masked ? MPFormat.maskedValue : MPFormat.money(category.amount), style: .emphasis)
                            }
                            GeometryReader { geometry in
                                Capsule()
                                    .fill(MPColor.row)
                                    .overlay(alignment: .leading) {
                                        Capsule()
                                            .fill(MPColor.categoryTeal)
                                            .frame(width: geometry.size.width * CGFloat(category.amount) / CGFloat(maximum))
                                    }
                            }
                            .frame(height: 6)
                        }
                        .padding(.vertical, 4)
                    }
                    .accessibilityLabel(category.name)
                    .accessibilityValue(masked ? MPFormat.maskedValue : MPFormat.money(category.amount))
                }
            }
        }
    }
}
