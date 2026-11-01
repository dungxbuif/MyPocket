import SwiftUI

struct MPChatAvatar: View {
    var isUser = false
    var size: CGFloat = 32

    var body: some View {
        MPIconImage(systemName: isUser ? "person" : MPIconCatalog.assetName(for: .assistant))
            .font(MPFont.bodyEmphasis)
            .foregroundStyle(isUser ? Color.white : MPColor.action)
            .frame(width: size, height: size)
            .background(isUser ? MPColor.action : MPColor.successSoft, in: Circle())
            .accessibilityHidden(true)
    }
}

/// Fixed action budget keeps long localized titles out of the button layout.
struct MPChatHeader: View {
    let status: String
    let onBack: () -> Void
    let actions: [MPOverflowAction]
    @Environment(\.dynamicTypeSize) private var dynamicTypeSize

    var body: some View {
        VStack(alignment: .leading, spacing: 4) {
            HStack(spacing: 4) {
                MPIconButton(systemName: "chevron.left", label: "Quay lại", action: onBack)
                    .accessibilityIdentifier("assistant-close")
                MPChatAvatar()
                if !dynamicTypeSize.isAccessibilitySize { identity }
                Spacer(minLength: 0)
                MPOverflowMenu(label: "Tùy chọn hội thoại", actions: actions)
                MPChatAvatar(isUser: true, size: 28)
            }
            if dynamicTypeSize.isAccessibilitySize { identity }
        }
        .padding(.horizontal, MPMetrics.pagePadding)
        .padding(.vertical, 4)
        .frame(minHeight: 56)
        .background(MPColor.canvas)
        .accessibilityIdentifier("assistant-header")
    }

    private var identity: some View {
        VStack(alignment: .leading, spacing: 2) {
            MPText("Trợ lý tài chính", style: .emphasis)
                .lineLimit(dynamicTypeSize.isAccessibilitySize ? nil : 1)
                .accessibilityIdentifier("assistant-header-title")
            Text(status)
                .font(MPFont.labelSmall)
                .foregroundStyle(MPColor.secondary)
                .lineLimit(dynamicTypeSize.isAccessibilitySize ? nil : 1)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
    }
}

struct MPChatGreeting: View {
    let income: String
    let expense: String
    let onDetails: () -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack(alignment: .top, spacing: 10) {
                MPChatAvatar(size: 40)
                VStack(alignment: .leading, spacing: 4) {
                    MPText("Chào bạn", style: .title)
                    MPText("Cùng xem lại tài chính của bạn tháng này.", style: .secondary)
                }
            }
            MPChatOverview(income: income, expense: expense, onDetails: onDetails)
        }
        .padding(12)
        .background(MPColor.successSoft.opacity(0.45), in: RoundedRectangle(cornerRadius: MPMetrics.referenceCornerRadius))
        .accessibilityIdentifier("assistant-greeting")
    }
}

enum MPChatFooterLayout {
    /// Reserved breathing room for the floating feedback control on every
    /// screen that hosts chat. Assistant is the deliberate exception because
    /// the shell hides feedback while the full-screen chat is visible.
    static let feedbackClearance: CGFloat = 60
}
