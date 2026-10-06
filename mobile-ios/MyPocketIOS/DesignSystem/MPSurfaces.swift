import SwiftUI

struct MPSurfaceCard<Content: View>: View {
    private let content: () -> Content

    init(@ViewBuilder content: @escaping () -> Content) {
        self.content = content
    }

    var body: some View {
        content()
            .padding(16)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(MPColor.surface, in: RoundedRectangle(cornerRadius: MPMetrics.cornerRadius))
    }
}

struct MPMetricCard: View {
    let title: String
    let value: String
    var detail: String?
    var masked: Bool = false

    var body: some View {
        MPSurfaceCard {
            VStack(alignment: .leading, spacing: 8) {
                Text(title)
                    .font(.subheadline)
                    .foregroundStyle(MPColor.secondary)
                Text(masked ? "••••••" : value)
                    .font(.title3.weight(.bold).monospacedDigit())
                    .foregroundStyle(MPColor.ink)
                    .accessibilityLabel(masked ? "Đã ẩn số tiền" : "\(title): \(value)")
                if let detail {
                    Text(detail)
                        .font(.footnote)
                        .foregroundStyle(MPColor.secondary)
                }
            }
        }
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
                .font(.footnote)
                .foregroundStyle(MPColor.secondary)
        }
    }
}

struct MPTransactionRow: View {
    let title: String
    let subtitle: String
    let amount: String
    let isIncome: Bool

    var body: some View {
        HStack(spacing: 12) {
            Image(systemName: isIncome ? "arrow.down.left" : "arrow.up.right")
                .foregroundStyle(isIncome ? MPColor.action : MPColor.danger)
                .frame(width: MPMetrics.hitTarget, height: MPMetrics.hitTarget)
            VStack(alignment: .leading, spacing: 3) {
                Text(title).font(.body.weight(.semibold))
                Text(subtitle).font(.footnote).foregroundStyle(MPColor.secondary)
            }
            Spacer()
            Text(amount)
                .font(.body.monospacedDigit())
                .foregroundStyle(isIncome ? MPColor.action : MPColor.ink)
        }
        .accessibilityElement(children: .combine)
    }
}

struct MPWalletRow: View {
    let name: String
    let type: String
    let balance: String

    var body: some View {
        HStack {
            VStack(alignment: .leading, spacing: 4) {
                Text(name).font(.body.weight(.semibold))
                Text(MPSelectionLabels.display(for: type)).font(.footnote).foregroundStyle(MPColor.secondary)
            }
            Spacer()
            Text(balance).font(.body.monospacedDigit())
        }
        .padding(.vertical, 8)
        .accessibilityElement(children: .combine)
    }
}
