import SwiftUI

enum MPButtonStyle: Equatable {
    case primary
    case outline
    case secondary
    /// Transparent action-colored button for low-emphasis actions such as
    /// clearing an optional selection, matching the web ghost-button base.
    case ghost
    /// Muted calculator key surface used by the shared VND amount keypad.
    /// Keeping the key geometry in the button base prevents transaction,
    /// budget and wallet forms from inventing separate numeric controls.
    case key
    /// White active pill used inside the shared reference segmented-control
    /// track. Keeping it in the button base prevents forms and report cards
    /// from recreating the Stitch/web selected-segment geometry locally.
    case segmentedSelected
    case destructive
    case plain
}

struct MPButton<Label: View>: View {
    private let style: MPButtonStyle
    private let isLoading: Bool
    private let isDisabled: Bool
    private let fullWidth: Bool
    private let accessibilityIdentifier: String?
    private let action: () -> Void
    private let label: () -> Label

    init(
        style: MPButtonStyle = .primary,
        isLoading: Bool = false,
        isDisabled: Bool = false,
        fullWidth: Bool = true,
        accessibilityIdentifier: String? = nil,
        action: @escaping () -> Void,
        @ViewBuilder label: @escaping () -> Label
    ) {
        self.style = style
        self.isLoading = isLoading
        self.isDisabled = isDisabled
        self.fullWidth = fullWidth
        self.accessibilityIdentifier = accessibilityIdentifier
        self.action = action
        self.label = label
    }

    @ViewBuilder
    var body: some View {
        let control = Button(action: action) {
            HStack(spacing: 8) {
                if isLoading {
                    ProgressView()
                        .tint(foreground)
                        .accessibilityLabel("Đang xử lý")
                }
                label()
            }
            .font(MPFont.label)
            .frame(minWidth: MPMetrics.hitTarget, minHeight: MPMetrics.controlHeight)
            .frame(maxWidth: fullWidth ? .infinity : nil)
            .contentShape(Rectangle())
            .padding(.horizontal, style == .plain ? 0 : 16)
            .background(background, in: RoundedRectangle(cornerRadius: MPMetrics.buttonCornerRadius))
            .overlay {
                if style == .outline {
                    RoundedRectangle(cornerRadius: MPMetrics.buttonCornerRadius)
                        .stroke(MPColor.action, lineWidth: MPMetrics.inputBorderWidth)
                }
            }
            .foregroundStyle(foreground)
        }
        .buttonStyle(.plain)
        .disabled(isLoading || isDisabled)
        .accessibilityAddTraits(.isButton)
        .mpAvoidFloatingFeedback()
        if let accessibilityIdentifier {
            control.accessibilityIdentifier(accessibilityIdentifier)
        } else {
            control
        }
    }

    private var background: Color {
        switch style {
        case .primary: return MPColor.action
        case .outline: return MPColor.surface
        case .secondary: return MPColor.successSoft
        case .ghost: return .clear
        case .segmentedSelected: return MPColor.surface
        case .key: return MPColor.row
        case .destructive: return MPColor.danger
        case .plain: return .clear
        }
    }

    private var foreground: Color {
        switch style {
        case .primary, .destructive: return .white
        case .outline, .secondary, .ghost, .segmentedSelected: return MPColor.action
        case .key: return MPColor.ink
        case .plain: return MPColor.ink
        }
    }
}

extension MPButton where Label == Text {
    init(
        _ title: String,
        style: MPButtonStyle = .primary,
        isLoading: Bool = false,
        isDisabled: Bool = false,
        fullWidth: Bool = true,
        action: @escaping () -> Void
    ) {
        self.init(style: style, isLoading: isLoading, isDisabled: isDisabled, fullWidth: fullWidth, action: action) {
            // The enclosing base applies MPFont.label. Do not reintroduce a
            // system-font override here; all button labels follow the bundled
            // Manrope/token contract.
            Text(title)
        }
    }
}

extension MPButton where Label == AnyView {
    init(
        _ title: String,
        systemName: String,
        style: MPButtonStyle = .primary,
        isLoading: Bool = false,
        isDisabled: Bool = false,
        fullWidth: Bool = true,
        action: @escaping () -> Void
    ) {
        self.init(style: style, isLoading: isLoading, isDisabled: isDisabled, fullWidth: fullWidth, action: action) {
            AnyView(SwiftUI.Label(title, systemImage: systemName))
        }
    }
}

struct MPIconButton: View {
    let systemName: String
    let label: String
    var avoidsFeedback = true
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
        .accessibilityIdentifier(label == "Quay lại" ? "screen-back" : "icon-\(label)")
        .mpAvoidFloatingFeedback(avoidsFeedback)
    }
}

/// Filled circular action used for compact, secondary actions in form
/// footers (for example, opening AI receipt capture from transaction entry).
/// Keeping the geometry here prevents screens from creating one-off icon
/// buttons that drift from the reference surface.
struct MPActionIconButton: View {
    let systemName: String
    let label: String
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            Image(systemName: systemName)
                .font(MPFont.bodyEmphasis)
                .frame(width: 56, height: 56)
                .foregroundStyle(.white)
                .background(MPColor.brand, in: Circle())
                .shadow(color: MPColor.brandHover.opacity(0.22), radius: 8, y: 4)
        }
        .buttonStyle(.plain)
        .accessibilityLabel(label)
    }
}

/// Small circular header action used by reference wallet/ledger surfaces.
/// The component keeps icon affordances compact without reusing the large
/// primary action geometry intended for form footers.
struct MPHeaderIconButton: View {
    let systemName: String
    let label: String
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            Image(systemName: systemName)
                .font(MPFont.bodyEmphasis)
                .frame(width: 52, height: 52)
                .foregroundStyle(MPColor.ink)
                .background(MPColor.surface, in: Circle())
                .shadow(color: MPColor.ink.opacity(0.05), radius: 8, y: 3)
        }
        .buttonStyle(.plain)
        .accessibilityLabel(label)
    }
}

struct MPTextLinkButton: View {
    let title: String
    let label: String?
    let action: () -> Void

    init(_ title: String, label: String? = nil, action: @escaping () -> Void) {
        self.title = title
        self.label = label
        self.action = action
    }

    var body: some View {
        Button(action: action) {
            Text(title)
                .font(MPFont.bodyEmphasis)
                .foregroundStyle(MPColor.brand)
                .frame(minHeight: MPMetrics.hitTarget)
        }
        .buttonStyle(.plain)
        .accessibilityLabel(label ?? title)
    }
}
