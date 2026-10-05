import SwiftUI

enum MPButtonStyle {
    case primary
    case secondary
    case destructive
    case plain
}

struct MPButton<Label: View>: View {
    private let style: MPButtonStyle
    private let isLoading: Bool
    private let isDisabled: Bool
    private let action: () -> Void
    private let label: () -> Label

    init(
        style: MPButtonStyle = .primary,
        isLoading: Bool = false,
        isDisabled: Bool = false,
        action: @escaping () -> Void,
        @ViewBuilder label: @escaping () -> Label
    ) {
        self.style = style
        self.isLoading = isLoading
        self.isDisabled = isDisabled
        self.action = action
        self.label = label
    }

    var body: some View {
        Button(action: action) {
            HStack(spacing: 8) {
                if isLoading {
                    ProgressView()
                        .tint(foreground)
                        .accessibilityLabel("Đang xử lý")
                }
                label()
            }
            .frame(minWidth: MPMetrics.hitTarget, minHeight: MPMetrics.controlHeight)
            .frame(maxWidth: .infinity)
            .contentShape(Rectangle())
            .padding(.horizontal, 16)
            .background(background, in: RoundedRectangle(cornerRadius: MPMetrics.cornerRadius))
            .foregroundStyle(foreground)
        }
        .buttonStyle(.plain)
        .disabled(isLoading || isDisabled)
        .accessibilityAddTraits(.isButton)
    }

    private var background: Color {
        switch style {
        case .primary: return MPColor.action
        case .secondary: return MPColor.surface
        case .destructive: return MPColor.danger
        case .plain: return .clear
        }
    }

    private var foreground: Color {
        switch style {
        case .primary, .destructive: return .white
        case .secondary, .plain: return MPColor.ink
        }
    }
}

extension MPButton where Label == Text {
    init(
        _ title: String,
        style: MPButtonStyle = .primary,
        isLoading: Bool = false,
        isDisabled: Bool = false,
        action: @escaping () -> Void
    ) {
        self.init(style: style, isLoading: isLoading, isDisabled: isDisabled, action: action) {
            Text(title).font(.body.weight(.semibold))
        }
    }
}

struct MPIconButton: View {
    let systemName: String
    let label: String
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            Image(systemName: systemName)
                .frame(width: MPMetrics.hitTarget, height: MPMetrics.hitTarget)
                .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
        .foregroundStyle(MPColor.ink)
        .accessibilityLabel(label)
    }
}
