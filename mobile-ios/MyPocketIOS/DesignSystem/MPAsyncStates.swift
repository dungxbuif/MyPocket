import SwiftUI

struct MPLoadingView: View {
    let label: String

    var body: some View {
        ProgressView(label)
            .frame(maxWidth: .infinity, minHeight: 120)
            .accessibilityIdentifier("loading-state")
    }
}

struct MPEmptyView: View {
    let title: String
    let detail: String

    var body: some View {
        MPSurfaceCard {
            VStack(alignment: .leading, spacing: 6) {
                Text(title).font(.headline)
                Text(detail).font(.footnote).foregroundStyle(MPColor.secondary)
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
                Text(title).font(.headline).foregroundStyle(MPColor.danger)
                Text(detail).font(.footnote).foregroundStyle(MPColor.secondary)
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
            .font(.footnote)
            .foregroundStyle(tone.color)
            .padding(10)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(tone.color.opacity(0.1), in: RoundedRectangle(cornerRadius: 8))
    }
}

enum MPStatusTone {
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
