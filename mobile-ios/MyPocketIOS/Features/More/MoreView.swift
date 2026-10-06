import SwiftUI

enum MoreDestination: String, CaseIterable, Identifiable {
    // Keep account near the top of the menu. The shell owns a persistent
    // bottom tab bar, so placing it last makes its hit target overlap that
    // bar on compact phones.
    case budgets, reports, categories, jars, months, recurring, travel, portfolio, credit, aiEntry, account, feedback, notifications, assistant

    var id: String { rawValue }
    var title: String {
        switch self {
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

struct MoreView: View {
    let api: FinanceAPI
    @ObservedObject var auth: AuthViewModel
    @State private var destination: MoreDestination?

    var body: some View {
        Group {
            if let destination {
                SecondaryScreen(destination: destination, api: api, auth: auth, onOpenJars: { self.destination = .jars }) { self.destination = nil }
            } else {
                ScrollView {
                    VStack(alignment: .leading, spacing: 16) {
                        MPScreenHeader(title: "Thêm", leading: { EmptyView() }, trailing: { EmptyView() })
                        MPSurfaceCard {
                            VStack(spacing: 0) {
                                ForEach(MoreDestination.allCases) { item in
                                    MPButton(style: .plain) { destination = item } label: {
                                        HStack(spacing: 12) {
                                            MPIcon(systemName: item.icon, size: MPMetrics.hitTarget)
                                            MPText(item.title, style: .emphasis)
                                            Spacer()
                                            MPIcon(systemName: "chevron.right", color: MPColor.secondary)
                                        }
                                        .accessibilityElement(children: .ignore)
                                        .accessibilityLabel(item.title)
                                    }
                                }
                            }
                        }
                        MPButton("Đăng xuất", style: .destructive, action: auth.logout)
                    }
                    .padding(MPMetrics.pagePadding)
                    // Leave room for the shell's bottom navigation so the
                    // final destinations remain tappable instead of being
                    // covered by the persistent tab bar.
                    .padding(.bottom, MPMetrics.hitTarget * 2)
                }
                .background(MPColor.canvas.ignoresSafeArea())
                // The shell's persistent tab bar sits outside this screen's
                // scroll view. Reserve its hit area so the last menu rows are
                // reported as non-hittable until XCTest/users scroll them
                // into view, rather than receiving a tap on the tab bar.
                .safeAreaInset(edge: .bottom) {
                    Color.clear.frame(height: MPMetrics.hitTarget * 2)
                }
            }
        }
        .accessibilityIdentifier("more-screen")
    }
}

struct SecondaryScreen: View {
    let destination: MoreDestination
    let api: FinanceAPI
    @ObservedObject var auth: AuthViewModel
    let onOpenJars: () -> Void
    let onBack: () -> Void

    var body: some View {
        Group {
            switch destination {
            case .budgets: BudgetsView(api: api, onBack: onBack)
            case .reports: ReportsView(api: api, onBack: onBack)
            case .categories: CategoriesView(api: api, onBack: onBack)
            case .jars: JarsView(api: api, onBack: onBack)
            case .months: MonthsView(api: api, onBack: onBack, onOpenJars: onOpenJars)
            case .recurring: RecurringView(api: api, onBack: onBack)
            case .travel: TravelView(api: api, onBack: onBack)
            case .portfolio: PortfolioView(api: api, onBack: onBack)
            case .credit: CreditView(api: api, onBack: onBack)
            case .notifications: NotificationsView(api: api, onBack: onBack, timeZone: accountTimeZone)
            case .assistant: AssistantView(api: api, onBack: onBack, ownerID: assistantOwnerID)
            case .aiEntry: AIEntryView(api: api, onBack: onBack)
            case .feedback: FeedbackView(api: api, onBack: onBack)
            case .account: AccountView(auth: auth, onBack: onBack)
            }
        }
        // Each secondary screen owns its own scroll position. Without an
        // explicit identity SwiftUI may reuse the previous route's offset,
        // hiding the header/back action after navigating from a long screen.
        .id(destination.id)
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
