import SwiftUI

struct MainTabView: View {
    @ObservedObject var auth: AuthViewModel
    let externalRoute: AppRoute?
    @State private var selection = "overview"
    @AppStorage("mypocket.privacy.masked") private var masked = false
    @State private var requestedAccountDestination: MoreDestination?
    @State private var scopedWalletID: String?
    @State private var addTransactionRequested = false
    @State private var feedbackPresented = false
    @State private var feedbackCapturePending = false
    @State private var feedbackScreenshot: Data?
    @State private var floatingActionClearance: CGFloat = 0
    @State private var keyboardVisible = false
    @State private var assistantDraft = ""
    @State private var requestedWalletID: String?
    @Environment(\.dynamicTypeSize) private var dynamicTypeSize

    init(auth: AuthViewModel, externalRoute: AppRoute? = nil) {
        self.auth = auth
        self.externalRoute = externalRoute
    }

    var body: some View {
        GeometryReader { geometry in
        let navigationVisible = !(selection == "assistant" && keyboardVisible)
        let navigationHeight = navigationVisible ? MPMetrics.bottomNavigationHeight : 0
        let feedbackFrame = MPFeedbackBubbleLayout.frame(
            in: CGRect(x: 0, y: 0, width: geometry.size.width,
                       height: max(0, geometry.size.height - navigationHeight)),
            clearance: floatingActionClearance, obstacles: [])
        VStack(spacing: 0) {
        ZStack(alignment: .bottom) {
            ZStack(alignment: .bottomTrailing) {
                Group {
                    switch selection {
                    case "overview":
                        OverviewView(
                            api: auth.financeAPI,
                            ownerID: authenticatedUserID,
                            snapshotCache: auth.financialCache,
                            masked: $masked,
                            onOpenTransactions: { selection = "transactions" },
                            onOpenWallets: { requestedAccountDestination = .wallets; selection = "account" },
                            onOpenMonth: { requestedAccountDestination = .months; selection = "account" },
                            onOpenJars: { requestedAccountDestination = .jars; selection = "account" },
                            onOpenNotifications: { requestedAccountDestination = .notifications; selection = "account" }
                        )
                    case "transactions":
                        TransactionsView(
                            api: auth.financeAPI,
                            initialWalletID: scopedWalletID,
                            ownerID: authenticatedUserID,
                            snapshotCache: auth.financialCache,
                            addRequested: $addTransactionRequested,
                            onAIEntry: {
                                addTransactionRequested = false
                                requestedAccountDestination = .aiEntry
                                selection = "account"
                            },
                            onOpenReports: {
                                requestedAccountDestination = .reports
                                selection = "account"
                            }
                        )
                        .id(scopedWalletID ?? "all-wallets")
                    case "assistant":
                        AssistantView(api: auth.financeAPI, onBack: { selection = "overview" }, ownerID: assistantOwnerID, masked: masked, draft: $assistantDraft)
                    case "account":
                        MoreView(
                            api: auth.financeAPI,
                            auth: auth,
                            masked: masked,
                            requestedDestination: $requestedAccountDestination,
                            requestedWalletID: requestedWalletID,
                            onOpenTransactions: { walletID in
                                scopedWalletID = walletID
                                selection = "transactions"
                            },
                            onOpenAIEntry: {
                                requestedAccountDestination = .aiEntry
                                selection = "account"
                            }
                        )
                    default:
                        OverviewView(
                            api: auth.financeAPI,
                            ownerID: authenticatedUserID,
                            snapshotCache: auth.financialCache,
                            masked: $masked,
                            onOpenTransactions: { selection = "transactions" },
                            onOpenNotifications: { requestedAccountDestination = .notifications; selection = "account" }
                        )
                    }
                }
                .frame(width: geometry.size.width, height: max(0, geometry.size.height - navigationHeight), alignment: .top)
                .clipped()
                .accessibilityElement(children: .contain)
                .accessibilityIdentifier("shell-content-viewport")

                if !feedbackPresented && !feedbackCapturePending && !keyboardVisible, let feedbackFrame {
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
                .position(x: feedbackFrame.midX, y: feedbackFrame.midY)
                }
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .top)
            .coordinateSpace(name: MPFeedbackBubbleLayout.coordinateSpace)

            if feedbackPresented {
                FeedbackOverlay(api: auth.financeAPI, screenshot: feedbackScreenshot) { feedbackPresented = false }
                    .padding(MPMetrics.pagePadding)
                    .background(.ultraThinMaterial)
                    .transition(.move(edge: .bottom).combined(with: .opacity))
            }
        }
        .frame(height: max(0, geometry.size.height - navigationHeight))
            if navigationVisible {
            ZStack(alignment: .bottom) {
                    MPBottomNavigation(
                        items: MPShellNavigation.primaryItems,
                        selection: $selection,
                        onAdd: {
                            selection = "transactions"
                            addTransactionRequested = true
                        },
                        onLongPressAdd: {
                            addTransactionRequested = false
                            requestedAccountDestination = .aiEntry
                            selection = "account"
                        },
                        // Account destinations own their own create actions. Hiding
                        // the shell add affordance there prevents its raised hit
                        // region from covering wallet/jar/category menu rows while
                        // the shell navigation remains above feature overlays.
                        isAddEnabled: selection != "account" && selection != "assistant"
                    )
                    // Route content is bounded above navigation. Keep this narrow
                    // blocker because the selected Account tab can otherwise
                    // fall through to a More row near the bottom edge.
                    .zIndex(2)
                    if selection == "account" && !dynamicTypeSize.isAccessibilitySize {
                        // This blocker lives inside the safe-area inset (unlike a
                        // normal overlay, which ends above the inset), so taps on
                        // the already-selected account tab cannot fall through to
                        // the last More row. Keep it narrow: a full-width blocker
                        // would also cover menu rows that are intentionally
                        // scrolled close to the shell inset.
                        HStack(spacing: 0) {
                            Spacer(minLength: 0)
                            Color.clear
                                .frame(width: 100, height: MPMetrics.hitTarget + 12)
                                .contentShape(Rectangle())
                        }
                        .frame(maxWidth: .infinity, alignment: .trailing)
                        .allowsHitTesting(true)
                        .zIndex(3)
                    }
            }
            .frame(height: MPMetrics.bottomNavigationHeight)
            }
        }
        }
        .environment(\.mpFinancialValuesMasked, masked)
        .mpObserveKeyboardVisibility($keyboardVisible)
        .onChange(of: authenticatedUserID) { _, _ in assistantDraft = "" }
        .onAppear { apply(externalRoute) }
        .onChange(of: externalRoute) { _, route in apply(route) }
        .onPreferenceChange(MPFloatingActionClearanceKey.self) { floatingActionClearance = $0 }
        .animation(.easeInOut(duration: 0.2), value: feedbackPresented)
    }

    private var assistantOwnerID: String {
        if case .authenticated(let session) = auth.phase { return session.user.id }
        return "unknown"
    }

    private var authenticatedUserID: String? {
        if case .authenticated(let session) = auth.phase { return session.user.id }
        return nil
    }

    private func apply(_ route: AppRoute?) {
        guard let route else { return }
        switch route {
        case .overview:
            requestedWalletID = nil
            requestedAccountDestination = nil
            selection = "overview"
        case .transactions(let walletID):
            requestedWalletID = nil
            scopedWalletID = walletID
            addTransactionRequested = false
            selection = "transactions"
        case .newTransaction(let walletID):
            requestedWalletID = nil
            scopedWalletID = walletID
            addTransactionRequested = true
            selection = "transactions"
        case .wallets:
            requestedWalletID = nil
            requestedAccountDestination = .wallets
            selection = "account"
        case .wallet(let id):
            requestedWalletID = id
            requestedAccountDestination = .wallets
            selection = "account"
        case .budgets:
            requestedWalletID = nil
            requestedAccountDestination = .budgets
            selection = "account"
        case .categories:
            requestedWalletID = nil
            requestedAccountDestination = .categories
            selection = "account"
        case .jars:
            requestedWalletID = nil
            requestedAccountDestination = .jars
            selection = "account"
        case .months:
            requestedWalletID = nil
            requestedAccountDestination = .months
            selection = "account"
        case .recurring:
            requestedWalletID = nil
            requestedAccountDestination = .recurring
            selection = "account"
        case .travel:
            requestedWalletID = nil
            requestedAccountDestination = .travel
            selection = "account"
        case .reports:
            requestedWalletID = nil
            requestedAccountDestination = .reports
            selection = "account"
        case .portfolio:
            requestedWalletID = nil
            requestedAccountDestination = .portfolio
            selection = "account"
        case .credit:
            requestedWalletID = nil
            requestedAccountDestination = .credit
            selection = "account"
        case .notifications:
            requestedWalletID = nil
            requestedAccountDestination = .notifications
            selection = "account"
        case .assistant:
            requestedWalletID = nil
            requestedAccountDestination = nil
            selection = "assistant"
        case .aiEntry:
            requestedWalletID = nil
            requestedAccountDestination = .aiEntry
            selection = "account"
        case .account:
            requestedWalletID = nil
            requestedAccountDestination = .account
            selection = "account"
        case .feedback:
            requestedWalletID = nil
            requestedAccountDestination = .feedback
            selection = "account"
        }
    }
}
