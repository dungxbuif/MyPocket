import SwiftUI

struct MPIcon: View {
    let systemName: String
    let label: String?
    let color: Color
    let size: CGFloat?

    init(systemName: String, label: String? = nil, color: Color = MPColor.ink, size: CGFloat? = nil) {
        self.systemName = systemName
        self.label = label
        self.color = color
        self.size = size
    }

    var body: some View {
        Image(systemName: systemName)
            .foregroundStyle(color)
            .frame(width: size, height: size)
            .contentShape(Rectangle())
            .modifier(MPIconAccessibility(label: label))
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
