import SwiftUI

struct MPLoadingView: View {
    let label: String

    var body: some View {
        ProgressView(label)
            .frame(maxWidth: .infinity, minHeight: 120)
            .accessibilityIdentifier("loading-state")
    }
}

/// Loading policy shared by finance screens. A refresh must not replace an
/// already-rendered ledger with a full-screen spinner; the existing snapshot
/// remains visible while the request is in flight.
enum MPLoadingPresentationPolicy {
    static func showsBlockingLoader(isLoading: Bool, hasContent: Bool) -> Bool {
        isLoading && !hasContent
    }

    /// A refresh keeps the last successful content on screen and only adds a
    /// compact progress affordance when there is something meaningful to
    /// preserve. This prevents a fast retry/foreground refresh from flashing
    /// a full-screen loading state.
    static func showsInlineProgress(isLoading: Bool, hasContent: Bool) -> Bool {
        isLoading && hasContent
    }
}

struct MPInlineRefreshProgress: View {
    let label: String

    var body: some View {
        HStack(spacing: 8) {
            ProgressView()
                .controlSize(.small)
            Text(label)
                .font(MPFont.bodySmall)
                .foregroundStyle(MPColor.secondary)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .accessibilityElement(children: .combine)
        .accessibilityIdentifier("refresh-progress")
    }
}

struct MPEmptyView: View {
    let title: String
    let detail: String

    var body: some View {
        MPSurfaceCard {
            VStack(alignment: .leading, spacing: 6) {
                Text(title).font(MPFont.headline)
                Text(detail).font(MPFont.bodySmall).foregroundStyle(MPColor.secondary)
            }
        }
        .accessibilityIdentifier("empty-state")
    }
}

struct MPErrorView: View {
    let title: String
    let detail: String
    let retry: () -> Void

    var body: some View {
        MPSurfaceCard {
            VStack(alignment: .leading, spacing: 10) {
                Text(title).font(MPFont.headline).foregroundStyle(MPColor.danger)
                Text(detail).font(MPFont.bodySmall).foregroundStyle(MPColor.secondary)
                MPButton("Thử lại", style: .secondary, action: retry)
            }
        }
        .accessibilityIdentifier("error-state")
    }
}

struct MPStatusMessage: View {
    let text: String
    var tone: MPStatusTone = .info

    var body: some View {
        Text(text)
            .font(MPFont.bodySmall)
            .foregroundStyle(tone.color)
            .padding(10)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(tone.color.opacity(0.1), in: RoundedRectangle(cornerRadius: 8))
    }
}

/// Compact semantic status/percentage badge shared by list rows and budget
/// thresholds. It deliberately accepts already-resolved copy; no business
/// status is guessed in the design system.
struct MPStatusBadge: View {
    enum Tone: Equatable {
        case success
        case muted
        case danger

        var color: Color {
            switch self {
            case .success: return MPColor.action
            case .muted: return MPColor.secondary
            case .danger: return MPColor.danger
            }
        }

        var fill: Color { color.opacity(0.10) }
    }

    static let accessibilityID = "status-badge"

    let text: String
    let tone: Tone

    static func tone(for active: Bool) -> Tone { active ? .success : .muted }

    var body: some View {
        Text(text)
            .font(MPFont.labelSmall)
            .foregroundStyle(tone.color)
            .padding(.horizontal, 9)
            .padding(.vertical, 5)
            .background(tone.fill, in: Capsule())
            .accessibilityElement(children: .ignore)
            .accessibilityLabel(text)
            .accessibilityIdentifier(Self.accessibilityID)
    }
}

/// Shared feedback status presentation. The web contract shows a concise
/// status label and, for fixed items, the linked release version underneath.
struct MPFeedbackStatus: View {
    let statusLabel: String
    let isFixed: Bool
    let changelogVersion: String?

    var body: some View {
        VStack(alignment: .trailing, spacing: 2) {
            MPText(statusLabel, style: .secondary)
            if isFixed, let changelogVersion {
                MPText("Fixed in \(changelogVersion)", style: .secondary)
            }
        }
        .multilineTextAlignment(.trailing)
        .accessibilityElement(children: .combine)
        .accessibilityLabel(accessibilityLabel)
    }

    private var accessibilityLabel: String {
        if isFixed, let changelogVersion {
            return "\(statusLabel), Fixed in \(changelogVersion)"
        }
        return statusLabel
    }
}

enum MPStatusTone: Equatable {
    case info
    case success
    case danger

    var color: Color {
        switch self {
        case .info: return MPColor.secondary
        case .success: return MPColor.action
        case .danger: return MPColor.danger
        }
    }
}
