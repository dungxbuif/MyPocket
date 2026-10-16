import SwiftUI

enum MoreDestination: String, CaseIterable, Identifiable, Hashable {
    // Keep account near the top of the menu. The shell owns a persistent
    // bottom tab bar, so placing it last makes its hit target overlap that
    // bar on compact phones.
    case wallets, budgets, reports, categories, jars, months, recurring, travel, portfolio, credit, aiEntry, account, feedback, notifications, assistant

    var id: String { rawValue }
    var title: String {
        switch self {
        case .wallets: return "Ví của tôi"
        case .budgets: return "Ngân sách"
        case .reports: return "Báo cáo"
        case .categories: return "Danh mục"
        case .jars: return "Hũ chi tiêu"
        case .months: return "Tháng"
        case .recurring: return "Giao dịch định kỳ"
        case .travel: return "Travel Mode"
        case .portfolio: return "Danh mục đầu tư"
        case .credit: return "Sổ tín dụng"
        case .notifications: return "Thông báo"
        case .assistant: return "Finance Assistant"
        case .aiEntry: return "Nhập giao dịch bằng AI"
        case .feedback: return "Phản hồi & changelog"
        case .account: return "Tài khoản"
        }
    }

    var icon: String {
        switch self {
        case .wallets: return "wallet.pass.fill"
        case .budgets: return "chart.pie.fill"
        case .reports: return "chart.bar.xaxis"
        case .categories: return "folder.fill"
        case .jars: return "shippingbox.fill"
        case .months: return "calendar"
        case .recurring: return "repeat"
        case .travel: return "airplane"
        case .portfolio: return "chart.line.uptrend.xyaxis"
        case .credit: return "creditcard.fill"
        case .notifications: return "bell.fill"
        case .assistant: return "sparkles"
        case .aiEntry: return "doc.text.image"
        case .feedback: return "bubble.left.and.bubble.right.fill"
        case .account: return "person.crop.circle"
        }
    }
}

/// Reference account navigation contract. The More route uses one named
/// grouped surface per section; feature destinations remain unchanged so deep
/// links and owner-scoped secondary routes keep their existing behavior.
enum MoreReferencePresentation {
    struct Section: Identifiable, Equatable {
        let title: String
        let items: [MoreDestination]
        var id: String { title }
    }

    static let sections: [Section] = [
        Section(title: "Phản hồi", items: [.feedback]),
        Section(title: "Ví của tôi", items: [.wallets, .credit, .portfolio]),
        Section(title: "Kế hoạch", items: [.budgets, .jars, .months]),
        Section(title: "Giao dịch", items: [.categories, .recurring, .travel, .reports]),
        Section(title: "Trợ lý", items: [.aiEntry])
    ]

    static let surfaceStyle: MPSurfaceCardStyle = .reference
    static let rowSpacing: CGFloat = 0
    static let sectionSpacing: CGFloat = MPMetrics.sectionSpacing
}

struct MoreView: View {
    let api: FinanceAPI
    @ObservedObject var auth: AuthViewModel
    let masked: Bool
    let onOpenTransactions: (String?) -> Void
    let onOpenAIEntry: () -> Void
    let requestedWalletID: String?
    @Binding private var requestedDestination: MoreDestination?
    @State private var navigationPath: [MoreDestination] = []

    init(
        api: FinanceAPI,
        auth: AuthViewModel,
        masked: Bool = false,
        requestedDestination: Binding<MoreDestination?> = .constant(nil),
        requestedWalletID: String? = nil,
        onOpenTransactions: @escaping (String?) -> Void = { _ in },
        onOpenAIEntry: @escaping () -> Void = {}
    ) {
        self.api = api
        self.auth = auth
        self.masked = masked
        self.onOpenTransactions = onOpenTransactions
        self.onOpenAIEntry = onOpenAIEntry
        self.requestedWalletID = requestedWalletID
        _requestedDestination = requestedDestination
    }

    var body: some View {
        NavigationStack(path: $navigationPath) {
            ScrollView {
                VStack(alignment: .leading, spacing: MPMetrics.sectionSpacing) {
                    MPCenteredScreenHeader(
                        title: "Mở rộng",
                        accessibilityIdentifier: "screen-title-Thêm",
                        trailing: {
                            MPHeaderSupportButton {
                                navigationPath = [.feedback]
                            }
                        }
                    )
                    MPProfileCard(
                        name: accountName,
                        email: accountEmail,
                        accountSubtitle: "Hồ sơ tài khoản",
                        accessibilityIdentifier: "more-profile-account"
                    ) {
                        navigationPath = [.account]
                    }
                    ForEach(MoreReferencePresentation.sections) { section in
                        VStack(alignment: .leading, spacing: 4) {
                            MPText(section.title, style: .secondary)
                                .padding(.horizontal, 4)
                            MPSurfaceCard(style: MoreReferencePresentation.surfaceStyle) {
                                VStack(alignment: .leading, spacing: MoreReferencePresentation.rowSpacing) {
                                    ForEach(Array(section.items.enumerated()), id: \.element) { index, item in
                                        if index > 0 { MPDivider() }
                                        MPMenuRow(
                                            title: item.title,
                                            systemName: item.icon,
                                            accessibilityIdentifier: "more-destination-\(item.id)"
                                        ) {
                                            navigationPath = [item]
                                        }
                                    }
                                }
                            }
                        }
                    }
                    MPButton("Đăng xuất", style: .destructive, isLoading: auth.isBusy) { Task { await auth.logout() } }
                    if let message = auth.errorMessage { MPStatusMessage(text: message, tone: .danger) }
                }
                .padding(MPMetrics.pagePadding)
            }
            .background(MPColor.canvas.ignoresSafeArea())
            .navigationDestination(for: MoreDestination.self) { destination in
                SecondaryScreen(
                    destination: destination,
                    api: api,
                    auth: auth,
                    masked: masked,
                    initialWalletID: requestedWalletID,
                    onOpenJars: {
                        navigationPath = [.jars]
                    },
                    onOpenTransactions: { walletID in
                        navigationPath.removeAll()
                        requestedDestination = nil
                        onOpenTransactions(walletID)
                    },
                    onOpenAIEntry: onOpenAIEntry
                ) {
                    closeDestination()
                }
            }
            .toolbar(.hidden, for: .navigationBar)
        }
        .onAppear {
            if let requestedDestination, navigationPath.last != requestedDestination {
                navigationPath = [requestedDestination]
            }
        }
        .onChange(of: requestedDestination) { _, destination in
            if let destination, navigationPath.last != destination {
                navigationPath = [destination]
            } else if destination == nil {
                if !navigationPath.isEmpty { navigationPath.removeAll() }
            }
        }
        .onChange(of: navigationPath) { _, path in
            let destination = path.last
            if requestedDestination != destination {
                requestedDestination = destination
            }
        }
    }

    private var accountName: String {
        if case .authenticated(let session) = auth.phase, let name = session.user.name, !name.isEmpty { return name }
        return "Tài khoản MyPocket"
    }

    private var accountEmail: String {
        if case .authenticated(let session) = auth.phase { return session.user.email ?? "" }
        return ""
    }

    private func closeDestination() {
        navigationPath.removeAll()
        requestedDestination = nil
    }
}

struct SecondaryScreen: View {
    let destination: MoreDestination
    let api: FinanceAPI
    @ObservedObject var auth: AuthViewModel
    let masked: Bool
    let initialWalletID: String?
    let onOpenJars: () -> Void
    let onOpenTransactions: (String?) -> Void
    let onOpenAIEntry: () -> Void
    let onBack: () -> Void

    var body: some View {
        Group {
            switch destination {
            case .wallets:
                WalletsView(
                    api: api,
                    initialWalletID: initialWalletID,
                    onSelectWallet: { onOpenTransactions($0) },
                    onBack: onBack,
                    onAIEntry: onOpenAIEntry
                )
            case .budgets: BudgetsView(api: api, onBack: onBack, masked: masked, ownerID: assistantOwnerID)
            case .reports: ReportsView(api: api, onBack: onBack, masked: masked, ownerID: assistantOwnerID, snapshotCache: auth.financialCache)
            case .categories: CategoriesView(api: api, onBack: onBack)
            case .jars: JarsView(api: api, onBack: onBack, masked: masked)
            case .months: MonthsView(api: api, onBack: onBack, onOpenJars: onOpenJars, masked: masked)
            case .recurring: RecurringView(api: api, onBack: onBack)
            case .travel: TravelView(api: api, onBack: onBack)
            case .portfolio: PortfolioView(api: api, onBack: onBack)
            case .credit: CreditView(api: api, onBack: onBack)
            case .notifications: NotificationsView(api: api, onBack: onBack, timeZone: accountTimeZone)
            case .assistant: AssistantView(api: api, onBack: onBack, ownerID: assistantOwnerID, masked: masked)
            case .aiEntry: AIEntryView(api: api, onBack: onBack)
            case .feedback: FeedbackView(api: api, onBack: onBack)
            case .account: AccountView(auth: auth, onBack: onBack)
            }
        }
        // Each secondary screen owns its own scroll position. Without an
        // explicit identity SwiftUI may reuse the previous route's offset,
        // hiding the header/back action after navigating from a long screen.
        .id(destination.id)
        // Secondary routes own the Stitch/web reference back capsule. Do not
        // render a second native navigation bar back button above it.
        .navigationBarBackButtonHidden(true)
    }

    private var assistantOwnerID: String {
        if case .authenticated(let session) = auth.phase { return session.user.id }
        return "unknown"
    }

    private var accountTimeZone: TimeZone {
        if case .authenticated(let session) = auth.phase {
            return TimeZone(identifier: session.user.timezone ?? "") ?? .current
        }
        return .current
    }
}
