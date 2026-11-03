import SwiftUI

/// Wallet-style row for a jar (or the "Chưa phân bổ" bucket): a plan figure, not money in a wallet.
/// `amount` is the headline figure and `amountCaption` names it ("Còn có thể chi"); `subtitle` gives the
/// month line. Consumers resolve the figures and copy through `JarPlanPresentation`.
struct MPJarPlanRow: View {
    let title: String
    let amount: String
    let amountCaption: String
    let subtitle: String
    var progress: Double? = nil
    var progressLabel: String = "Đã chi / mức tháng"
    var badge: (text: String, tone: MPStatusBadge.Tone)? = nil
    let accessibilityIdentifier: String
    let action: () -> Void

    var body: some View {
        MPButton(style: .plain, accessibilityIdentifier: accessibilityIdentifier, action: action) {
            VStack(alignment: .leading, spacing: MPMetrics.rowSpacing) {
                HStack(alignment: .top, spacing: MPMetrics.rowSpacing) {
                    VStack(alignment: .leading, spacing: 4) {
                        MPText(title, style: .emphasis).lineLimit(2)
                        if let badge { MPStatusBadge(text: badge.text, tone: badge.tone) }
                    }
                    Spacer(minLength: MPMetrics.sectionSpacing)
                    VStack(alignment: .trailing, spacing: 4) {
                        MPText(amountCaption, style: .secondary)
                        MPText(amount, style: .emphasis)
                    }
                    MPIcon(systemName: "chevron.right", color: MPColor.secondary)
                        .accessibilityHidden(true)
                }
                MPText(subtitle, style: .secondary)
                if let progress { MPProgressBar(progress: progress, label: progressLabel) }
            }
        }
    }
}
