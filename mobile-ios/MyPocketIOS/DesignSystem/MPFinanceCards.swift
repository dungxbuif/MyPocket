import SwiftUI

struct MPChart: View {
    let values: [(label: String, value: Double)]

    var body: some View {
        HStack(alignment: .bottom, spacing: 8) {
            ForEach(Array(values.enumerated()), id: \.offset) { _, item in
                VStack(spacing: 4) {
                    RoundedRectangle(cornerRadius: 4)
                        .fill(MPColor.action)
                        .frame(height: max(4, CGFloat(item.value)))
                    Text(item.label).font(.caption2).foregroundStyle(MPColor.secondary)
                }
                .frame(maxWidth: .infinity)
            }
        }
        .frame(minHeight: 100, alignment: .bottom)
        .accessibilityElement(children: .ignore)
        .accessibilityLabel("Biểu đồ tài chính")
    }
}

struct MPAssistantCard<Content: View>: View {
    let title: String
    private let content: () -> Content

    init(title: String, @ViewBuilder content: @escaping () -> Content) {
        self.title = title
        self.content = content
    }

    var body: some View {
        MPSurfaceCard {
            VStack(alignment: .leading, spacing: 10) {
                Text(title).font(.headline)
                content()
            }
        }
    }
}

struct MPActionCard: View {
    let title: String
    let detail: String
    let actionTitle: String
    let action: () -> Void

    var body: some View {
        MPSurfaceCard {
            VStack(alignment: .leading, spacing: 10) {
                Text(title).font(.headline)
                Text(detail).font(.footnote).foregroundStyle(MPColor.secondary)
                MPButton(actionTitle, action: action)
            }
        }
    }
}
