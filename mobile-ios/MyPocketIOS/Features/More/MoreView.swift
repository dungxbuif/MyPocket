import SwiftUI

enum MoreDestination: String, CaseIterable, Identifiable {
    case budgets, reports, categories, jars, months, recurring, travel, portfolio, credit, assistant, aiEntry, feedback, account

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
                                            Image(systemName: item.icon).frame(width: MPMetrics.hitTarget)
                                            Text(item.title).font(.body.weight(.semibold))
                                            Spacer()
                                            Image(systemName: "chevron.right").foregroundStyle(MPColor.secondary)
                                        }
                                    }
                                }
                            }
                        }
                        MPButton("Đăng xuất", style: .destructive, action: auth.logout)
                    }
                    .padding(MPMetrics.pagePadding)
                }
                .background(MPColor.canvas.ignoresSafeArea())
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
        case .assistant: AssistantView(api: api, onBack: onBack, ownerID: assistantOwnerID)
        case .aiEntry: AIEntryView(api: api, onBack: onBack)
        case .feedback: FeedbackView(api: api, onBack: onBack)
        case .account: AccountView(auth: auth, onBack: onBack)
        }
    }

    private var assistantOwnerID: String {
        if case .authenticated(let session) = auth.phase { return session.user.id }
        return "unknown"
    }
}
