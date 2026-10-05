import SwiftUI

struct MPScreenHeader<Leading: View, Trailing: View>: View {
    let title: String
    private let leading: () -> Leading
    private let trailing: () -> Trailing

    init(
        title: String,
        @ViewBuilder leading: @escaping () -> Leading,
        @ViewBuilder trailing: @escaping () -> Trailing
    ) {
        self.title = title
        self.leading = leading
        self.trailing = trailing
    }

    var body: some View {
        HStack {
            leading()
            Text(title)
                .font(.title3.weight(.bold))
                .frame(maxWidth: .infinity, alignment: .leading)
            trailing()
        }
        .frame(minHeight: MPMetrics.hitTarget)
    }
}

struct MPBottomNavigation: View {
    let items: [(id: String, title: String, systemName: String)]
    @Binding var selection: String

    var body: some View {
        HStack {
            ForEach(items, id: \.id) { item in
                Button {
                    selection = item.id
                } label: {
                    VStack(spacing: 4) {
                        Image(systemName: item.systemName)
                        Text(item.title).font(.caption2)
                    }
                    .frame(maxWidth: .infinity, minHeight: MPMetrics.hitTarget)
                }
                .buttonStyle(.plain)
                .foregroundStyle(selection == item.id ? MPColor.action : MPColor.secondary)
                .accessibilityLabel(item.title)
                .accessibilityAddTraits(selection == item.id ? .isSelected : [])
            }
        }
        .padding(.horizontal, 8)
        .padding(.top, 8)
        .background(.bar)
    }
}

struct MPSegmentedControl: View {
    let options: [(id: String, title: String)]
    @Binding var selection: String

    var body: some View {
        HStack(spacing: 4) {
            ForEach(options, id: \.id) { option in
                MPButton(style: selection == option.id ? .primary : .plain) {
                    selection = option.id
                } label: {
                    Text(option.title)
                        .font(.footnote.weight(.semibold))
                        .frame(maxWidth: .infinity)
                }
                .accessibilityLabel(option.title)
                .accessibilityAddTraits(selection == option.id ? .isSelected : [])
            }
        }
        .padding(4)
        .background(MPColor.surface, in: RoundedRectangle(cornerRadius: MPMetrics.cornerRadius))
    }
}

/// Shared overflow action trigger for screen-level secondary actions.
/// Consumers provide semantic menu actions while this base owns the hit target
/// and accessible label, keeping the three-dot affordance consistent.
struct MPOverflowAction: Identifiable {
    let id = UUID()
    let title: String
    let action: () -> Void
}

struct MPOverflowMenu: View {
    let label: String
    let actions: [MPOverflowAction]

    init(label: String, actions: [MPOverflowAction]) {
        self.label = label
        self.actions = actions
    }

    var body: some View {
        Menu {
            ForEach(actions) { item in
                Button(item.title, action: item.action)
            }
        } label: {
            Image(systemName: "ellipsis")
                .frame(width: MPMetrics.hitTarget, height: MPMetrics.hitTarget)
                .contentShape(Rectangle())
        }
        .foregroundStyle(MPColor.ink)
        .accessibilityLabel(label)
    }
}

struct MPFeedbackBubble: View {
    let action: () -> Void

    var body: some View {
        MPIconButton(systemName: "bubble.left.and.bubble.right.fill", label: "Gửi phản hồi", action: action)
            .padding(8)
            .background(MPColor.action, in: Circle())
            .foregroundStyle(.white)
            .shadow(radius: 6, y: 3)
            .accessibilityIdentifier("feedback-bubble")
    }
}

struct MPConfirmDialog: ViewModifier {
    @Binding var isPresented: Bool
    let title: String
    let message: String
    let confirmTitle: String
    let confirm: () -> Void

    func body(content: Content) -> some View {
        content.alert(title, isPresented: $isPresented) {
            Button("Hủy", role: .cancel) {}
            Button(confirmTitle, role: .destructive, action: confirm)
        } message: {
            Text(message)
        }
    }
}

extension View {
    func mpConfirmDialog(
        isPresented: Binding<Bool>,
        title: String,
        message: String,
        confirmTitle: String = "Xác nhận",
        confirm: @escaping () -> Void
    ) -> some View {
        modifier(MPConfirmDialog(isPresented: isPresented, title: title, message: message, confirmTitle: confirmTitle, confirm: confirm))
    }
}
