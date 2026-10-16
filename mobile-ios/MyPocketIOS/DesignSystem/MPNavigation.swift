import SwiftUI
import UIKit

private struct MPKeyboardVisibilityObserver: ViewModifier {
    @Binding var visible: Bool
    func body(content: Content) -> some View {
        content
            .onReceive(NotificationCenter.default.publisher(for: UIResponder.keyboardWillShowNotification).receive(on: RunLoop.main)) { _ in visible = true }
            .onReceive(NotificationCenter.default.publisher(for: UIResponder.keyboardDidHideNotification).receive(on: RunLoop.main)) { _ in visible = false }
    }
}

extension View {
    func mpObserveKeyboardVisibility(_ visible: Binding<Bool>) -> some View {
        modifier(MPKeyboardVisibilityObserver(visible: visible))
    }
}

/// Long-form presentation shared by management features. Callers own draft,
/// validation and requests; this base owns one scroll region and pinned actions.
struct MPEditorScaffold<Header: View, Fields: View, Footer: View>: View {
    let isBusy: Bool
    @ViewBuilder var header: () -> Header
    @ViewBuilder var fields: () -> Fields
    @ViewBuilder var footer: () -> Footer
    @State private var fieldsViewport: CGRect?

    init(isBusy: Bool, @ViewBuilder header: @escaping () -> Header,
         @ViewBuilder fields: @escaping () -> Fields,
         @ViewBuilder footer: @escaping () -> Footer) {
        self.isBusy = isBusy; self.header = header
        self.fields = fields; self.footer = footer
    }

    init(title: String, isBusy: Bool, onCancel: @escaping () -> Void,
         @ViewBuilder fields: @escaping () -> Fields,
         @ViewBuilder footer: @escaping () -> Footer) where Header == MPReferenceScreenHeader {
        self.init(isBusy: isBusy, header: {
            MPReferenceScreenHeader(title: title, closeLabel: "Hủy", onClose: onCancel)
        }, fields: fields, footer: footer)
    }

    var body: some View {
        VStack(spacing: 0) {
            header()
                .disabled(isBusy)
                .padding(.horizontal, MPMetrics.pagePadding)
                .frame(minHeight: 52)
            ScrollView {
                VStack(alignment: .leading, spacing: MPMetrics.sectionSpacing, content: fields)
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .padding(MPMetrics.pagePadding)
                    .disabled(isBusy)
                    .environment(\.mpFeedbackObstacleClip, fieldsViewport)
            }
            .scrollDismissesKeyboard(.interactively)
            .background {
                GeometryReader { geometry in
                    Color.clear.preference(key: MPFeedbackFieldsViewportKey.self,
                        value: geometry.frame(in: .named(MPFeedbackBubbleLayout.coordinateSpace)))
                }
            }
            .onPreferenceChange(MPFeedbackFieldsViewportKey.self) { fieldsViewport = $0 }
            VStack(spacing: 8, content: footer)
                .padding(.horizontal, MPMetrics.pagePadding)
                .padding(.vertical, 8)
                .background(MPColor.canvas)
                .background {
                    GeometryReader { geometry in
                        Color.clear.preference(key: MPFloatingActionClearanceKey.self, value: geometry.size.height)
                    }
                }
        }
        .background(MPColor.canvas.ignoresSafeArea())
        .interactiveDismissDisabled(isBusy)
    }
}

/// A pinned input reports its actual height to the shell's floating action.
/// This moves only the bubble, never reserves a lane in scroll content.
struct MPFloatingActionClearanceKey: PreferenceKey {
    static let defaultValue: CGFloat = 0
    static func reduce(value: inout CGFloat, nextValue: () -> CGFloat) {
        value = max(value, nextValue())
    }
}

enum MPFeedbackBubbleLayout {
    static let coordinateSpace = "mp-feedback-viewport"

    static func frame(in viewport: CGRect, clearance: CGFloat, obstacles: [CGRect]) -> CGRect? {
        let side = MPShellNavigation.feedbackBubbleDiameter
        guard viewport.width >= side, viewport.height >= side,
              viewport.minX.isFinite, viewport.minY.isFinite,
              viewport.maxX.isFinite, viewport.maxY.isFinite else { return nil }
        let gutter = min(MPMetrics.pagePadding, (viewport.width - side) / 2)
        let footer = clearance.isFinite ? max(0, clearance) : 0
        let bottom = max(viewport.minY, viewport.maxY - footer - 8 - side)
        // The shell owns one stable dock position. Obstacle-driven movement
        // made the bubble jump between rows and could still cover the very
        // control it was trying to avoid after a layout update. Protected
        // content is kept reachable by the footer clearance/scroll inset; the
        // feedback trigger itself never reserves a lane or changes sides.
        _ = obstacles
        return CGRect(x: viewport.maxX - gutter - side, y: bottom, width: side, height: side)
    }
}

private struct MPFeedbackFieldsViewportKey: PreferenceKey {
    static var defaultValue: CGRect? { nil }
    static func reduce(value: inout CGRect?, nextValue: () -> CGRect?) {
        if let next = nextValue() { value = next }
    }
}

private struct MPFeedbackObstacleClipKey: EnvironmentKey {
    static var defaultValue: CGRect? { nil }
}

extension EnvironmentValues {
    var mpFeedbackObstacleClip: CGRect? {
        get { self[MPFeedbackObstacleClipKey.self] }
        set { self[MPFeedbackObstacleClipKey.self] = newValue }
    }
}

struct MPFeedbackObstacleKey: PreferenceKey {
    static var defaultValue: [CGRect] { [] }
    static func reduce(value: inout [CGRect], nextValue: () -> [CGRect]) {
        value.append(contentsOf: nextValue())
    }
}

private struct MPFeedbackObstacleModifier: ViewModifier {
    let enabled: Bool
    @Environment(\.mpFeedbackObstacleClip) private var clip
    func body(content: Content) -> some View {
        content.background {
            if enabled {
                GeometryReader { geometry in
                    let frame = geometry.frame(in: .named(MPFeedbackBubbleLayout.coordinateSpace))
                    Color.clear.preference(key: MPFeedbackObstacleKey.self,
                        value: clip.map { $0.contains(frame) } == false ? [] : [frame])
                }
                .allowsHitTesting(false)
            }
        }
    }
}

extension View {
    func mpAvoidFloatingFeedback(_ enabled: Bool = true) -> some View {
        modifier(MPFeedbackObstacleModifier(enabled: enabled))
    }
}

struct MPScreenHeader<Leading: View, Trailing: View>: View {
    let title: String
    let accessibilityIdentifier: String
    private let leading: () -> Leading
    private let trailing: () -> Trailing

    init(
        title: String,
        accessibilityIdentifier: String? = nil,
        @ViewBuilder leading: @escaping () -> Leading,
        @ViewBuilder trailing: @escaping () -> Trailing
    ) {
        self.title = title
        self.accessibilityIdentifier = accessibilityIdentifier ?? "screen-title-\(title)"
        self.leading = leading
        self.trailing = trailing
    }

    var body: some View {
        HStack {
            leading()
            if !title.isEmpty { Text(title)
                .font(MPFont.headline)
                .frame(maxWidth: .infinity, alignment: .leading)
                .accessibilityIdentifier(accessibilityIdentifier) }
            trailing()
        }
        .frame(minHeight: MPMetrics.hitTarget)
        // Keep leading/trailing controls discoverable as separate elements in
        // long-form screens (notably Reports) while exposing the title as the
        // header's own content.
        .accessibilityElement(children: .contain)
    }
}

/// Centered wallet-selection header with balanced capsule actions from the
/// Stitch reference. Navigation/edit behavior remains owned by the caller.
struct MPSelectionScreenHeader: View {
    let title: String
    let closeTitle: String
    let closeAccessibilityLabel: String
    let closeAccessibilityIdentifier: String
    let actionTitle: String
    let actionAccessibilityIdentifier: String
    let onClose: (() -> Void)?
    let onAction: () -> Void

    var body: some View {
        ZStack {
            MPText(title, style: .title)
                .fixedSize()
                .accessibilityIdentifier("screen-title-\(title)")
            HStack {
                if let onClose {
                    capsuleAction(
                        closeTitle,
                        accessibilityLabel: closeAccessibilityLabel,
                        identifier: closeAccessibilityIdentifier,
                        action: onClose
                    )
                } else {
                    Color.clear
                        .frame(width: 80, height: MPMetrics.hitTarget)
                        .accessibilityHidden(true)
                }
                Spacer(minLength: 0)
                capsuleAction(
                    actionTitle,
                    accessibilityLabel: actionTitle,
                    identifier: actionAccessibilityIdentifier,
                    action: onAction
                )
            }
        }
        .frame(minHeight: MPMetrics.hitTarget)
        .accessibilityElement(children: .contain)
    }

    private func capsuleAction(
        _ title: String,
        accessibilityLabel: String,
        identifier: String,
        action: @escaping () -> Void
    ) -> some View {
        MPButton(style: .plain, fullWidth: false, action: action) {
            Text(title)
                .font(MPFont.body)
                .foregroundStyle(MPColor.ink)
        }
        .padding(.horizontal, 12)
        .background(MPColor.surface, in: RoundedRectangle(cornerRadius: MPMetrics.buttonCornerRadius))
        .shadow(color: MPColor.ink.opacity(0.06), radius: 10, y: 4)
        .accessibilityLabel(accessibilityLabel)
        .accessibilityIdentifier(identifier)
    }
}

/// Stitch-style support action used by the More/account landing surface. The
/// pill geometry belongs here so the screen does not invent a one-off header
/// button while keeping the existing route/accessibility contract.
struct MPHeaderSupportButton: View {
    let action: () -> Void

    var body: some View {
        MPButton(style: .plain, fullWidth: false, action: action) {
            HStack(spacing: 6) {
                Text("Hỗ trợ")
                Image(systemName: "questionmark.circle")
                    .font(MPFont.bodyEmphasis)
            }
        }
        .padding(.horizontal, 10)
        .background(MPColor.surface, in: RoundedRectangle(cornerRadius: MPMetrics.buttonCornerRadius))
        .shadow(color: MPColor.ink.opacity(0.04), radius: 8, y: 3)
        .accessibilityLabel("Hỗ trợ")
        .accessibilityIdentifier("more-profile-help")
    }
}

/// Stitch-style sheet header used by high-frequency forms. The close action
/// stays at a 44pt hit target while the title is visually centered regardless
/// of trailing content. Screens may override the accessibility identifier to
/// preserve a stable route contract while changing the visible copy.
struct MPReferenceScreenHeader: View {
    let title: String
    let closeLabel: String
    let accessibilityIdentifier: String
    let onClose: () -> Void

    init(
        title: String,
        closeLabel: String = "Hủy",
        accessibilityIdentifier: String? = nil,
        onClose: @escaping () -> Void
    ) {
        self.title = title
        self.closeLabel = closeLabel
        self.accessibilityIdentifier = accessibilityIdentifier ?? "screen-title-\(title)"
        self.onClose = onClose
    }

    var body: some View {
        HStack(spacing: 12) {
            MPButton(style: .plain, fullWidth: false, action: onClose) {
                Text(closeLabel)
                    .font(MPFont.body)
                    .foregroundStyle(MPColor.ink)
            }
            .padding(.horizontal, 12)
            .background(MPColor.surface, in: RoundedRectangle(cornerRadius: MPMetrics.buttonCornerRadius))
            .shadow(color: MPColor.ink.opacity(0.06), radius: 10, y: 4)
            .accessibilityLabel(closeLabel)
            .accessibilityIdentifier("screen-close")
            Spacer(minLength: 0)
            MPText(title, style: .title)
                .accessibilityIdentifier(accessibilityIdentifier)
            Spacer(minLength: 0)
            Color.clear.frame(width: 44, height: MPMetrics.hitTarget)
                .accessibilityHidden(true)
        }
        .frame(minHeight: MPMetrics.hitTarget)
    }
}

/// Header whose title remains centered when the trailing action has a
/// different width. This is the More/account landing contract from Stitch;
/// regular secondary headers continue using `MPScreenHeader`.
struct MPCenteredScreenHeader<Trailing: View>: View {
    let title: String
    let accessibilityIdentifier: String
    private let trailing: () -> Trailing

    init(
        title: String,
        accessibilityIdentifier: String? = nil,
        @ViewBuilder trailing: @escaping () -> Trailing
    ) {
        self.title = title
        self.accessibilityIdentifier = accessibilityIdentifier ?? "screen-title-(title)"
        self.trailing = trailing
    }

    var body: some View {
        ZStack {
            MPText(title, style: .title)
                .accessibilityIdentifier(accessibilityIdentifier)
            HStack {
                Spacer(minLength: 0)
                trailing()
            }
        }
        .frame(minHeight: MPMetrics.hitTarget)
        .accessibilityElement(children: .contain)
    }
}

/// Base presentation wrapper for reference forms. Feature screens should use
/// this instead of invoking SwiftUI's modal modifier directly so modal
/// coverage, dismissal and future platform styling stay centralized.
struct MPFullScreenCover<Presented: View>: ViewModifier {
    @Binding private var isPresented: Bool
    private let presented: () -> Presented

    init(isPresented: Binding<Bool>, @ViewBuilder presented: @escaping () -> Presented) {
        self._isPresented = isPresented
        self.presented = presented
    }

    func body(content: Content) -> some View {
        content.fullScreenCover(isPresented: $isPresented, content: presented)
    }
}

extension View {
    func mpFullScreenCover<Presented: View>(
        isPresented: Binding<Bool>,
        @ViewBuilder content: @escaping () -> Presented
    ) -> some View {
        modifier(MPFullScreenCover(isPresented: isPresented, presented: content))
    }
}

/// Shared native presentation for short transactional forms that are sheets
/// on the web. The base owns the reference header, scroll region, pinned
/// footer, canvas and iOS sheet presentation; feature views provide only
/// domain fields and actions.
struct MPFormBottomSheet<Fields: View, Footer: View>: ViewModifier {
    @Binding private var isPresented: Bool
    private let title: String
    private let accessibilityIdentifier: String
    private let isDismissDisabled: Bool
    private let onClose: () -> Void
    private let fields: () -> Fields
    private let footer: () -> Footer

    init(
        isPresented: Binding<Bool>,
        title: String,
        accessibilityIdentifier: String,
        isDismissDisabled: Bool = false,
        onClose: @escaping () -> Void,
        @ViewBuilder fields: @escaping () -> Fields,
        @ViewBuilder footer: @escaping () -> Footer
    ) {
        self._isPresented = isPresented
        self.title = title
        self.accessibilityIdentifier = accessibilityIdentifier
        self.isDismissDisabled = isDismissDisabled
        self.onClose = onClose
        self.fields = fields
        self.footer = footer
    }

    func body(content: Content) -> some View {
        content.sheet(isPresented: $isPresented) {
            VStack(spacing: 16) {
                MPReferenceScreenHeader(
                    title: title,
                    closeLabel: "Hủy",
                    accessibilityIdentifier: accessibilityIdentifier,
                    onClose: onClose
                )
                ScrollView {
                    VStack(alignment: .leading, spacing: 12, content: fields)
                        .frame(maxWidth: .infinity, alignment: .leading)
                }
                .scrollDismissesKeyboard(.interactively)
            }
            .padding(.horizontal, MPMetrics.pagePadding)
            .padding(.top, MPMetrics.pagePadding)
            .safeAreaInset(edge: .bottom, spacing: 0) {
                footer()
                    .padding(.top, 12)
                    .padding(.bottom, 8)
                    .background(MPColor.canvas)
            }
            .background(MPColor.canvas.ignoresSafeArea())
            .presentationDetents([.large])
            .presentationDragIndicator(.visible)
            .presentationBackground(MPColor.canvas)
            .interactiveDismissDisabled(isDismissDisabled)
        }
    }
}

extension View {
    func mpFormBottomSheet<Fields: View, Footer: View>(
        isPresented: Binding<Bool>,
        title: String,
        accessibilityIdentifier: String,
        isDismissDisabled: Bool = false,
        onClose: @escaping () -> Void,
        @ViewBuilder fields: @escaping () -> Fields,
        @ViewBuilder footer: @escaping () -> Footer
    ) -> some View {
        modifier(MPFormBottomSheet(
            isPresented: isPresented,
            title: title,
            accessibilityIdentifier: accessibilityIdentifier,
            isDismissDisabled: isDismissDisabled,
            onClose: onClose,
            fields: fields,
            footer: footer
        ))
    }
}

struct MPInfoAlert: ViewModifier {
    @Binding var isPresented: Bool
    let title: String
    let message: String

    func body(content: Content) -> some View {
        content.alert(title, isPresented: $isPresented) {
            Button("Đóng", role: .cancel) {}
        } message: {
            Text(message)
        }
    }
}

extension View {
    func mpInfoAlert(isPresented: Binding<Bool>, title: String, message: String) -> some View {
        modifier(MPInfoAlert(isPresented: isPresented, title: title, message: message))
    }
}

/// Canonical primary shell navigation. The identifiers and order mirror the
/// web MobileAppShell; secondary finance surfaces remain reachable from
/// Tài khoản so the bottom bar stays stable across platforms.
enum MPShellNavigation {
    enum TabSelectionStyle: Equatable {
        case surfacePill
    }

    static let primaryItems: [(id: String, title: String, systemName: String)] = [
        ("overview", "Tổng quan", "house.fill"),
        ("transactions", "Giao dịch", "list.bullet.rectangle"),
        ("assistant", "Trợ lý", "message.fill"),
        ("account", "Tài khoản", "ellipsis.circle.fill")
    ]

    static let addActionAccessibilityID = "add-transaction-button"
    static let selectedTabStyle: TabSelectionStyle = .surfacePill
    static let bottomNavigationCornerRadius: CGFloat = 28
    static let selectedTabCornerRadius: CGFloat = 22
    static let addButtonDiameter: CGFloat = 56
    static let feedbackBubbleDiameter: CGFloat = 52
}

struct MPBalanceHeader: View {
    let totalBalance: String
    let masked: Bool
    let onToggleMask: () -> Void
    let onSearch: () -> Void
    let onNotifications: () -> Void

    init(
        totalBalance: String,
        masked: Bool,
        onToggleMask: @escaping () -> Void,
        onSearch: @escaping () -> Void = {},
        onNotifications: @escaping () -> Void = {}
    ) {
        self.totalBalance = totalBalance
        self.masked = masked
        self.onToggleMask = onToggleMask
        self.onSearch = onSearch
        self.onNotifications = onNotifications
    }

    var body: some View {
        HStack(alignment: .top, spacing: 12) {
            VStack(alignment: .leading, spacing: 2) {
                HStack(alignment: .firstTextBaseline, spacing: 4) {
                    MPText(masked ? "••••••••" : totalBalance, style: .emphasis)
                        .font(MPFont.display.monospacedDigit())
                        .accessibilityLabel(masked ? "Đã ẩn tổng số dư" : "Tổng số dư: \(totalBalance)")
                    MPIconButton(
                        systemName: masked ? "eye.slash" : "eye",
                        label: masked ? "Hiện số dư" : "Ẩn số dư",
                        action: onToggleMask
                    )
                    .accessibilityIdentifier("icon-\(masked ? "Hiện số dư" : "Ẩn số dư")")
                }
                HStack(spacing: 4) {
                    MPText("Tổng số dư", style: .body)
                    Image(systemName: "questionmark.circle.fill")
                        .font(MPFont.labelSmall)
                        .foregroundStyle(MPColor.muted)
                        .accessibilityLabel("Thông tin tổng số dư")
                }
            }
            Spacer()
            HStack(spacing: 0) {
                MPIconButton(systemName: "magnifyingglass", label: "Tìm kiếm", action: onSearch)
                MPIconButton(systemName: "bell", label: "Thông báo", action: onNotifications)
            }
        }
        .padding(.horizontal, MPMetrics.pagePadding)
        .padding(.top, 18)
        .padding(.bottom, 18)
        .background(MPColor.canvas.opacity(0.96))
    }
}

struct MPBottomNavigation: View {
    @Environment(\.dynamicTypeSize) private var dynamicTypeSize
    let items: [(id: String, title: String, systemName: String)]
    @Binding var selection: String
    let onAdd: () -> Void
    let onLongPressAdd: (() -> Void)?
    let isAddEnabled: Bool

    init(
        items: [(id: String, title: String, systemName: String)],
        selection: Binding<String>,
        onAdd: @escaping () -> Void,
        onLongPressAdd: (() -> Void)? = nil,
        isAddEnabled: Bool = true
    ) {
        self.items = items
        _selection = selection
        self.onAdd = onAdd
        self.onLongPressAdd = onLongPressAdd
        self.isAddEnabled = isAddEnabled
    }

    var body: some View {
        if dynamicTypeSize.isAccessibilitySize {
            HStack(spacing: 8) {
                MPMenuSelector(options: items.map { MPSelectionOption(id: $0.id, label: $0.title) }, onSelect: { selection = $0 }) {
                    HStack(spacing: 8) {
                        Text(items.first(where: { $0.id == selection })?.title ?? "Chọn màn hình")
                            .font(MPFont.labelSmall)
                            .fixedSize(horizontal: false, vertical: true)
                        Spacer(minLength: 0)
                        Image(systemName: "chevron.down").accessibilityHidden(true)
                    }
                    .padding(.horizontal, 12)
                    .frame(maxWidth: .infinity, minHeight: MPMetrics.hitTarget)
                    .background(MPColor.surface, in: RoundedRectangle(cornerRadius: MPMetrics.buttonCornerRadius))
                }
                .accessibilityIdentifier("bottom-navigation-menu")
                .accessibilityLabel("Chọn màn hình, \(items.first(where: { $0.id == selection })?.title ?? "")")
                if isAddEnabled { MPBottomAddButton(action: onAdd, onLongPress: onLongPressAdd) }
            }
            .padding(.horizontal, MPMetrics.pagePadding)
        } else {
        ZStack(alignment: .top) {
            HStack(spacing: 0) {
                ForEach(Array(items.enumerated()), id: \.element.id) { index, item in
                    if index == 2 {
                        Color.clear
                            .frame(maxWidth: .infinity, minHeight: MPMetrics.hitTarget)
                    }
                    Button {
                        selection = item.id
                    } label: {
                        VStack(spacing: 4) {
                            Image(systemName: item.systemName)
                            Text(item.title).font(MPFont.labelSmall)
                        }
                        .frame(maxWidth: .infinity, minHeight: MPMetrics.hitTarget)
                        .padding(.horizontal, 4)
                        .padding(.vertical, 2)
                        .background(
                            selection == item.id ? MPColor.row : .clear,
                            in: RoundedRectangle(
                                cornerRadius: MPShellNavigation.selectedTabCornerRadius,
                                style: .continuous
                            )
                        )
                    }
                    .buttonStyle(.plain)
                    .foregroundStyle(selection == item.id ? MPColor.action : MPColor.secondary)
                    .accessibilityLabel(item.title)
                    .accessibilityIdentifier("bottom-tab-\(item.id)")
                    .accessibilityAddTraits(selection == item.id ? .isSelected : [])
                }
            }
            .frame(height: MPMetrics.hitTarget)
            .frame(maxWidth: .infinity)
            .padding(.horizontal, 8)
            .padding(.top, 6)
            .padding(.bottom, 2)
            .background(
                MPColor.surface,
                in: RoundedRectangle(
                    cornerRadius: MPShellNavigation.bottomNavigationCornerRadius,
                    style: .continuous
                )
            )
            .shadow(color: MPColor.ink.opacity(0.08), radius: 14, y: 4)
            .padding(.horizontal, 16)
            .padding(.bottom, 4)
            .overlay(alignment: .top) {
                if isAddEnabled {
                    MPBottomAddButton(action: onAdd, onLongPress: onLongPressAdd)
                        .offset(y: -24)
                }
            }
        }
        }
    }
}

/// Shared clearance for an independently presented editor/modal that remains
/// above the shell navigation. Shell-hosted routes use MainTabView's bounded
/// viewport instead of adding this spacer to every screen.
struct MPBottomNavigationInset: View {
    var body: some View {
        Color.clear
            .frame(height: MPMetrics.hitTarget * 2)
            .accessibilityIdentifier("bottom-navigation-inset")
    }
}


/// Shared semantic row used by account/secondary navigation surfaces.
/// Consumers provide only the destination metadata and action; this base owns
/// the 44pt hit target, row padding, chevron treatment and accessibility ID.
struct MPMenuRow: View {
    let title: String
    let subtitle: String?
    let systemName: String
    let accessibilityIdentifier: String
    let isEnabled: Bool
    let action: () -> Void

    init(
        title: String,
        subtitle: String? = nil,
        systemName: String,
        accessibilityIdentifier: String,
        isEnabled: Bool = true,
        action: @escaping () -> Void
    ) {
        self.title = title
        self.subtitle = subtitle
        self.systemName = systemName
        self.accessibilityIdentifier = accessibilityIdentifier
        self.isEnabled = isEnabled
        self.action = action
    }

    var body: some View {
        MPButton(style: .plain, accessibilityIdentifier: accessibilityIdentifier, action: action) {
            HStack(spacing: 12) {
                MPIcon(systemName: systemName, size: 32)
                    .accessibilityHidden(true)
                if let subtitle, !subtitle.isEmpty {
                    VStack(alignment: .leading, spacing: 2) {
                        MPText(title, style: .emphasis)
                        MPText(subtitle, style: .secondary)
                    }
                } else {
                    MPText(title, style: .emphasis)
                }
                Spacer()
                MPIcon(systemName: "chevron.right", color: MPColor.secondary)
                    .accessibilityHidden(true)
            }
            .frame(maxWidth: .infinity, minHeight: MPMetrics.controlHeight, alignment: .leading)
            .contentShape(Rectangle())
        }
        .contentShape(Rectangle())
        .disabled(!isEnabled)
        .opacity(isEnabled ? 1 : 0.55)
        .accessibilityLabel(title)
        .accessibilityIdentifier(accessibilityIdentifier)
    }
}

struct MPBottomAddButton: View {
    let action: () -> Void
    let onLongPress: (() -> Void)?
    @State private var longPressActivated = false

    init(action: @escaping () -> Void, onLongPress: (() -> Void)? = nil) {
        self.action = action
        self.onLongPress = onLongPress
    }

    var body: some View {
        Button {
            // SwiftUI may deliver the Button action after a successful
            // long-press gesture. Guard that follow-up tap so the AI route is
            // not immediately overwritten by the normal transaction route.
            guard !longPressActivated else {
                longPressActivated = false
                return
            }
            action()
        } label: {
            Image(systemName: "plus")
                .font(MPFont.headline)
                .frame(
                    width: MPShellNavigation.addButtonDiameter,
                    height: MPShellNavigation.addButtonDiameter
                )
                .foregroundStyle(.white)
                .background(MPColor.brand, in: Circle())
                .shadow(color: MPColor.brandHover.opacity(0.25), radius: 10, y: 5)
        }
        .buttonStyle(.plain)
        .highPriorityGesture(
            LongPressGesture(minimumDuration: 0.5)
                .onEnded { _ in
                    guard onLongPress != nil else { return }
                    longPressActivated = true
                    onLongPress?()
                    // The Button's release action is delivered after this
                    // gesture callback. Clear the one-shot guard on the next
                    // run-loop turn so it cannot suppress a later, genuine
                    // short tap if the platform does not deliver that
                    // release action.
                    DispatchQueue.main.async {
                        longPressActivated = false
                    }
                }
        )
        .accessibilityLabel("Thêm giao dịch")
        .accessibilityIdentifier(MPShellNavigation.addActionAccessibilityID)
    }
}

struct MPSegmentedControl: View {
    let options: [(id: String, title: String)]
    @Binding var selection: String
    let accessibilityIdentifier: String?
    @Environment(\.accessibilityReduceMotion) private var reduceMotion

    init(
        options: [(id: String, title: String)],
        selection: Binding<String>,
        accessibilityIdentifier: String? = nil
    ) {
        self.options = options
        self._selection = selection
        self.accessibilityIdentifier = accessibilityIdentifier
    }

    var body: some View {
        HStack(spacing: 0) {
            ForEach(options, id: \.id) { option in
                MPButton(style: Self.buttonStyle(isSelected: selection == option.id)) {
                    let duration = Self.transitionDuration(reduceMotion: reduceMotion)
                    if duration == 0 {
                        selection = option.id
                    } else {
                        withAnimation(.easeInOut(duration: duration)) {
                            selection = option.id
                        }
                    }
                } label: {
                    Text(option.title)
                        .font(MPFont.label)
                        .frame(maxWidth: .infinity)
                }
                .accessibilityLabel(option.title)
                .accessibilityIdentifier(option.id == options.first?.id ? accessibilityIdentifier ?? "segment-\(option.id)" : "segment-\(option.id)")
                .accessibilityAddTraits(selection == option.id ? .isSelected : [])
            }
        }
        .padding(4)
        .background(MPColor.line, in: RoundedRectangle(cornerRadius: MPMetrics.buttonCornerRadius))
    }

    static func buttonStyle(isSelected: Bool) -> MPButtonStyle {
        isSelected ? .segmentedSelected : .plain
    }

    static func transitionDuration(reduceMotion: Bool) -> Double {
        reduceMotion ? 0 : 0.18
    }
}

/// Account-local week navigation used by the ledger. The base owns hit targets
/// and accessibility labels so feature screens do not create ad-hoc controls.
struct MPWeekNavigator: View {
    let rangeLabel: String
    let isNextDisabled: Bool
    let onPrevious: () -> Void
    let onNext: () -> Void

    var body: some View {
        HStack(spacing: 12) {
            MPIconButton(systemName: "chevron.left", label: "Tuần trước", action: onPrevious)
            MPText(rangeLabel, style: .body)
                .frame(maxWidth: .infinity)
                .multilineTextAlignment(.center)
                .accessibilityAddTraits(.updatesFrequently)
            MPIconButton(systemName: "chevron.right", label: "Tuần sau", action: onNext)
                .opacity(isNextDisabled ? 0.4 : 1)
                .disabled(isNextDisabled)
        }
        .padding(.horizontal, 4)
        .frame(minHeight: MPMetrics.hitTarget)
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
        MPIconButton(systemName: "bubble.left.and.bubble.right.fill", label: "Gửi phản hồi", avoidsFeedback: false, action: action)
            .accessibilityIdentifier("floating-feedback-bubble")
            .frame(
                width: MPShellNavigation.feedbackBubbleDiameter,
                height: MPShellNavigation.feedbackBubbleDiameter
            )
            .background(MPColor.surface, in: Circle())
            .overlay(Circle().stroke(MPColor.line, lineWidth: MPMetrics.inputBorderWidth))
            .foregroundStyle(MPColor.secondary)
            .shadow(radius: 6, y: 3)
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
