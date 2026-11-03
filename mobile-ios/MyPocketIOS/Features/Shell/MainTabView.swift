import SwiftUI

enum MPShellContentLayout {
    static func routeHeight(containerHeight: CGFloat, navigationHeight: CGFloat) -> CGFloat {
        max(0, containerHeight - max(0, navigationHeight))
    }
}

struct MainTabView: View {
    @ObservedObject var auth: AuthViewModel
    let externalRoute: AppRoute?
    @State private var selection = "overview"
    @AppStorage("mypocket.privacy.masked") private var masked = false
    @State private var requestedAccountDestination: MoreDestination?
    @State private var scopedWalletID: String?
    @State private var addTransactionRequested = false
    @State private var transactionReturnTab: String?
    @State private var feedbackPresented = false
    @State private var feedbackCapturePending = false
    @State private var feedbackScreenshot: Data?
    @State private var keyboardVisible = false
    @State private var assistantDraft = ""
    @State private var assistantReportHandoff: ReportAssistantHandoff?
    @State private var assistantPendingMessage: AdvisorPendingMessageModel?
    @State private var assistantPresented = false
    @State private var footerClearance: CGFloat = 0
    @State private var requestedWalletID: String?
    @State private var overviewPath: [MoreDestination] = []
    @Environment(\.dynamicTypeSize) private var dynamicTypeSize

    init(auth: AuthViewModel, externalRoute: AppRoute? = nil) {
        self.auth = auth
        self.externalRoute = externalRoute
    }

    @Environment(\.mpBackgroundWrites) private var backgroundWrites

    var body: some View {
        GeometryReader { geometry in
        // A full-screen Assistant is a separate experience, not another shell
        // route. Remove shell controls from both rendering and accessibility
        // while it is presented so VoiceOver and UI automation cannot reach
        // navigation or Feedback through the cover.
        let navigationVisible = !keyboardVisible && !assistantPresented
        let feedbackVisible = !keyboardVisible && !assistantPresented && !feedbackPresented && !feedbackCapturePending
        let navigationHeight = navigationVisible ? MPMetrics.bottomNavigationHeight : 0
        let contentHeight = MPShellContentLayout.routeHeight(
            containerHeight: geometry.size.height,
            navigationHeight: navigationHeight
        )
        ZStack(alignment: .bottomTrailing) {
            VStack(spacing: 0) {
            VStack(spacing: 0) {
                MPTabKeepAlive(selection: selection, tabs: ["overview", "transactions", "reports", "account"], preload: ["transactions", "reports"]) { tab in
                    switch tab {
                    case "overview":
                        NavigationStack(path: $overviewPath) {
                        OverviewView(
                            api: auth.financeAPI,
                            ownerID: authenticatedUserID,
                            snapshotCache: auth.financialCache,
                            masked: $masked,
                            onOpenTransactions: { selection = "transactions" },
                            onOpenWallets: { overviewPath.append(.wallets) },
                            onOpenMonth: { overviewPath.append(.months) },
                            onOpenJars: { overviewPath.append(.jars) },
                            onOpenNotifications: { overviewPath.append(.notifications) }
                        )
                        .navigationDestination(for: MoreDestination.self) { destination in
                            SecondaryScreen(
                                destination: destination,
                                api: auth.financeAPI,
                                auth: auth,
                                masked: masked,
                                initialWalletID: requestedWalletID,
                                onOpenJars: { overviewPath = [.jars] },
                                onOpenTransactions: { walletID in
                                    overviewPath.removeAll()
                                    scopedWalletID = walletID
                                    selection = "transactions"
                                },
                                onOpenAIEntry: {
                                    overviewPath.removeAll()
                                    requestedAccountDestination = .aiEntry
                                    selection = "account"
                                },
                                onOpenAssistant: presentAssistant,
                                onBack: { if !overviewPath.isEmpty { overviewPath.removeLast() } }
                            )
                        }
                        .toolbar(.hidden, for: .navigationBar)
                        }
                    case "transactions":
                        TransactionsView(
                            api: auth.financeAPI,
                            initialWalletID: scopedWalletID,
                            ownerID: authenticatedUserID,
                            snapshotCache: auth.financialCache,
                            addRequested: $addTransactionRequested,
                            onAIEntry: {
                                transactionReturnTab = nil
                                addTransactionRequested = false
                                requestedAccountDestination = .aiEntry
                                selection = "account"
                            },
                            onOpenReports: {
                                requestedAccountDestination = .reports
                                selection = "account"
                            },
                            onEditorDismissed: {
                                if let returnTab = transactionReturnTab {
                                    transactionReturnTab = nil
                                    selection = returnTab
                                }
                            }
                        )
                        .id(scopedWalletID ?? "all-wallets")
                    case "reports":
                        MonthlyReportsView(
                            api: auth.financeAPI,
                            onBack: {},
                            masked: masked,
                            ownerID: authenticatedUserID,
                            showsBack: false,
                            onChatHandoff: presentAssistant
                        )
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
                            },
                            onOpenAssistant: presentAssistant
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
                .overlay(alignment: .top) { MPBackgroundWriteStatus(center: backgroundWrites) }
                .frame(width: geometry.size.width, height: contentHeight, alignment: .top)
                .clipped()
                .accessibilityHidden(assistantPresented)
                .accessibilityElement(children: .contain)
                .accessibilityIdentifier("shell-content-viewport")
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .top)
            if navigationVisible {
                ZStack(alignment: .bottom) {
                    MPBottomNavigation(
                        items: MPShellNavigation.primaryItems,
                        selection: $selection,
                        onAdd: {
                            transactionReturnTab = selection == "account" ? "account" : nil
                            selection = "transactions"
                            addTransactionRequested = true
                        },
                        onLongPressAdd: {
                            transactionReturnTab = nil
                            addTransactionRequested = false
                            requestedAccountDestination = .aiEntry
                            selection = "account"
                        },
                        // Account root shares global Add; child routes and chat
                        // retain their own actions and protected input space.
                        isAddEnabled: selection != "assistant" &&
                            (selection != "account" || requestedAccountDestination == nil)
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
            if feedbackVisible {
                MPFeedbackBubble {
                    // Hide the trigger for one rendered frame before capture.
                    feedbackCapturePending = true
                    Task { @MainActor in
                        try? await Task.sleep(for: .milliseconds(100))
                        guard feedbackCapturePending else { return }
                        feedbackScreenshot = ScreenSnapshot.capture()
                        feedbackCapturePending = false
                        feedbackPresented = true
                    }
                }
                .padding(.trailing, MPMetrics.pagePadding)
                // Remain a true overlay while clearing any pinned editor
                // footer. This moves the bubble; it never reserves route
                // height or inserts a dock/spacer into the shell.
                .padding(.bottom, navigationHeight + footerClearance + 8)
                .zIndex(4)
            }
        }
        if feedbackPresented {
            FeedbackOverlay(api: auth.financeAPI, screenshot: feedbackScreenshot) { feedbackPresented = false }
                .padding(MPMetrics.pagePadding)
                .background(.ultraThinMaterial)
                .transition(.move(edge: .bottom).combined(with: .opacity))
        }
        }
        .environment(\.mpFinancialValuesMasked, masked)
        .onPreferenceChange(MPFloatingActionClearanceKey.self) { footerClearance = $0 }
        .mpObserveKeyboardVisibility($keyboardVisible)
        .environment(\.reportAssistantHandoff, presentReportAssistant)
        .onChange(of: authenticatedUserID) { _, _ in
            assistantDraft = ""; assistantReportHandoff = nil; assistantPendingMessage = nil
            assistantPresented = false
        }
        .mpFullScreenCover(isPresented: $assistantPresented) {
            AssistantView(
                api: auth.financeAPI,
                onBack: { assistantPresented = false },
                ownerID: assistantOwnerID,
                masked: masked,
                draft: $assistantDraft,
                reportHandoff: $assistantReportHandoff,
                pendingMessage: $assistantPendingMessage
            )
        }
        .onAppear { apply(externalRoute) }
        .onChange(of: externalRoute) { _, route in apply(route) }
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

    private func presentAssistant(_ draft: String) {
        if !draft.isEmpty {
            if assistantDraft.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty { assistantDraft = draft }
            else if assistantDraft != draft { assistantDraft += "\n\n" + draft }
        }
        assistantPresented = true
    }

    private func presentReportAssistant(_ handoff: ReportAssistantHandoff) {
        assistantReportHandoff = handoff
        // Context is visible beside the composer; an existing user draft is
        // preserved verbatim, including while an older request awaits retry.
        if assistantDraft.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
            assistantDraft = handoff.prompt
        }
        assistantPresented = true
    }

    private func apply(_ route: AppRoute?) {
        guard let route else { return }
        switch route {
        case .overview:
            requestedWalletID = nil
            requestedAccountDestination = nil
            overviewPath.removeAll()
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
            requestedAccountDestination = nil
            selection = "reports"
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
            assistantPresented = true
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
