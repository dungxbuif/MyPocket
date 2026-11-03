import SwiftUI

private struct MPTabIsActiveKey: EnvironmentKey { static let defaultValue = true }

extension EnvironmentValues {
    /// False while a retained tab is mounted but not on screen. Screens use it to
    /// run a silent background refresh when the tab becomes active again.
    var mpTabIsActive: Bool {
        get { self[MPTabIsActiveKey.self] }
        set { self[MPTabIsActiveKey.self] = newValue }
    }
}

/// Keeps every tab that was visited alive (state, scroll position and loaded data)
/// instead of destroying it on each switch. Tabs mount lazily on first visit; an
/// inactive tab is invisible, not hit-testable and hidden from accessibility.
struct MPTabKeepAlive<Content: View>: View {
    let selection: String
    let tabs: [String]
    /// Tabs mounted shortly after launch so their first visit is already warm.
    var preload: [String] = []
    @ViewBuilder let content: (String) -> Content
    @State private var visited: Set<String> = []

    var body: some View {
        ZStack(alignment: .top) {
            ForEach(tabs, id: \.self) { tab in
                if visited.contains(tab) || tab == selection {
                    let active = tab == selection
                    content(tab)
                        .environment(\.mpTabIsActive, active)
                        .opacity(active ? 1 : 0)
                        .allowsHitTesting(active)
                        .disabled(!active)
                        // Ignoring children as well as hiding makes the inactive subtree leave the
                        // accessibility tree entirely without changing view identity (state is kept).
                        .accessibilityElement(children: active ? .contain : .ignore)
                        .accessibilityHidden(!active)
                        .zIndex(active ? 1 : 0)
                }
            }
        }
        .onAppear { visited.insert(selection) }
        .task {
            guard !preload.isEmpty else { return }
            try? await Task.sleep(nanoseconds: 800_000_000)
            visited.formUnion(preload)
        }
        .onChange(of: selection) { _, tab in visited.insert(tab) }
    }
}

extension View {
    /// Runs `refresh` each time a retained tab returns to the foreground. The
    /// existing content stays on screen; `refresh` must be non-blocking.
    func mpRefreshOnTabActivation(_ active: Bool, perform refresh: @escaping () async -> Void) -> some View {
        onChange(of: active) { _, isActive in
            if isActive { Task { await refresh() } }
        }
    }
}
