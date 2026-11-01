import SwiftUI

/// Renders a catalog asset or a platform symbol through one base boundary.
/// Callers can continue passing legacy system names while migrated semantic
/// roles use the bundled Phosphor asset names.
struct MPIconImage: View {
    let systemName: String
    var size: CGFloat = 22

    var body: some View {
        if MPIconCatalog.isBundledAssetName(systemName) {
            Image(systemName)
                .resizable()
                .renderingMode(.template)
                .scaledToFit()
                .frame(width: size, height: size)
        } else {
            Image(systemName: systemName)
                .frame(width: size, height: size)
        }
    }
}

/// Native SF Symbol contract for reference surfaces. Feature screens consume
/// these names instead of embedding icon strings in their own layout code.
enum MPReferenceIcon {
    static let account = MPIconCatalog.assetName(for: .account)
    static let wallet = MPIconCatalog.assetName(for: .wallet)
    static let group = MPIconCatalog.assetName(for: .group)
    static let bankLink = MPIconCatalog.assetName(for: .bankLink)
    static let recurring = MPIconCatalog.assetName(for: .recurring)
    static let feedback = MPIconCatalog.assetName(for: .feedback)
}

struct MPIcon: View {
    let systemName: String
    let label: String?
    let color: Color
    let size: CGFloat?
    let semanticID: MPIconID?
    let style: MPIconStyle?

    init(systemName: String, label: String? = nil, color: Color = MPColor.ink, size: CGFloat? = nil) {
        self.systemName = systemName
        self.label = label
        self.color = color
        self.size = size
        self.semanticID = nil
        self.style = nil
    }

    /// Semantic initializer for catalog-backed icons. The system-name
    /// initializer above remains available for platform controls and scoped
    /// migration work; neither initializer changes a persisted API icon key.
    init(iconID: MPIconID, style: MPIconStyle = .regular, label: String? = nil, color: Color = MPColor.ink, size: CGFloat? = nil) {
        self.systemName = MPIconCatalog.assetName(for: iconID, style: style)
        self.label = label
        self.color = color
        self.size = size
        self.semanticID = iconID
        self.style = style
    }

    var body: some View {
        iconImage
            .foregroundStyle(color)
            .frame(width: size ?? 22, height: size ?? 22)
            .contentShape(Rectangle())
            .modifier(MPIconAccessibility(label: label))
    }

    @ViewBuilder private var iconImage: some View {
        if MPIconCatalog.isBundledAssetName(systemName) {
            Image(systemName)
                .resizable()
                .renderingMode(.template)
                .scaledToFit()
        } else {
            Image(systemName: systemName)
        }
    }
}

private struct MPIconAccessibility: ViewModifier {
    let label: String?

    func body(content: Content) -> some View {
        if let label, !label.isEmpty {
            content.accessibilityLabel(label)
        } else {
            content.accessibilityHidden(true)
        }
    }
}
