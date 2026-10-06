import SwiftUI

struct MainTabView: View {
    @ObservedObject var auth: AuthViewModel
    @State private var selection = "overview"
    @State private var feedbackPresented = false
    @State private var feedbackCapturePending = false
    @State private var feedbackScreenshot: Data?

    var body: some View {
        ZStack(alignment: .bottomTrailing) {
            Group {
                switch selection {
                case "overview":
                    OverviewView(api: auth.financeAPI) { selection = "transactions" }
                case "transactions":
                    TransactionsView(api: auth.financeAPI)
                case "wallets":
                    WalletsView(api: auth.financeAPI)
                default:
                    MoreView(api: auth.financeAPI, auth: auth)
                }
            }
            if !feedbackPresented && !feedbackCapturePending {
                MPFeedbackBubble {
                    // Hide the trigger for a rendered frame before taking the
                    // screenshot so the attached context is the underlying
                    // screen, not the feedback control itself.
                    feedbackCapturePending = true
                    Task { @MainActor in
                        try? await Task.sleep(for: .milliseconds(100))
                        guard feedbackCapturePending else { return }
                        feedbackScreenshot = ScreenSnapshot.capture()
                        feedbackCapturePending = false
                        feedbackPresented = true
                    }
                }
                .padding(.trailing, 16)
                .padding(.bottom, 82)
            }
            if feedbackPresented {
                FeedbackOverlay(api: auth.financeAPI, screenshot: feedbackScreenshot) { feedbackPresented = false }
                    .padding(MPMetrics.pagePadding)
                    .background(.ultraThinMaterial)
                    .transition(.move(edge: .bottom).combined(with: .opacity))
            }
        }
        .safeAreaInset(edge: .bottom) {
            MPBottomNavigation(
                items: [
                    ("overview", "Tổng quan", "house.fill"),
                    ("transactions", "Giao dịch", "list.bullet.rectangle"),
                    ("wallets", "Ví", "wallet.pass.fill"),
                    ("more", "Thêm", "ellipsis.circle.fill")
                ],
                selection: $selection
            )
        }
        .animation(.easeInOut(duration: 0.2), value: feedbackPresented)
        .accessibilityIdentifier("main-tab-screen")
    }
}
