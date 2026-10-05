import SwiftUI

@MainActor
final class OverviewViewModel: ObservableObject {
    @Published private(set) var wallets: [WalletModel] = []
    @Published private(set) var transactions: [TransactionModel] = []
    @Published private(set) var isLoading = false
    @Published private(set) var errorMessage: String?
    @Published private(set) var accountTimezone = TimeZone.current.identifier
    @Published var masked = false

    private let api: FinanceAPI

    init(api: FinanceAPI) { self.api = api }

    var totalBalance: Int64 { wallets.filter(\.isInTotal).reduce(0) { $0 + $1.currentBalance } }
    var monthExpense: Int64 {
        let month = MPFormat.currentMonth(timeZone: TimeZone(identifier: accountTimezone) ?? .current)
        return transactions.filter { $0.type == "expense" && $0.includedInReports && MPFormat.accountMonth(of: $0.occurredAt, timeZone: TimeZone(identifier: accountTimezone) ?? .current) == month }.reduce(0) { $0 + $1.amount }
    }
    var monthIncome: Int64 {
        let month = MPFormat.currentMonth(timeZone: TimeZone(identifier: accountTimezone) ?? .current)
        return transactions.filter { $0.type == "income" && $0.includedInReports && MPFormat.accountMonth(of: $0.occurredAt, timeZone: TimeZone(identifier: accountTimezone) ?? .current) == month }.reduce(0) { $0 + $1.amount }
    }

    func load() async {
        guard !isLoading else { return }
        isLoading = true
        errorMessage = nil
        defer { isLoading = false }
        do {
            async let wallets = api.wallets()
            async let transactions = api.transactions()
            async let profile = api.profile()
            self.wallets = try await wallets
            self.transactions = try await transactions.sorted { $0.occurredAt > $1.occurredAt }
            if let profile = try? await profile { accountTimezone = profile.timezone }
        } catch {
            errorMessage = Self.message(for: error)
        }
    }

    private static func message(for error: Error) -> String {
        if let apiError = error as? APIError {
            switch apiError {
            case .missingSession: return "Phiên đăng nhập đã hết hạn."
            case .transport: return "Không thể kết nối máy chủ."
            case .problem(let problem): return problem.detail ?? "Không thể tải dữ liệu tài chính."
            default: return "Không thể tải dữ liệu tài chính."
            }
        }
        return "Không thể tải dữ liệu tài chính."
    }
}

struct OverviewView: View {
    @StateObject private var model: OverviewViewModel
    let onOpenTransactions: () -> Void

    init(api: FinanceAPI, onOpenTransactions: @escaping () -> Void) {
        _model = StateObject(wrappedValue: OverviewViewModel(api: api))
        self.onOpenTransactions = onOpenTransactions
    }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                MPScreenHeader(title: "Tổng quan", leading: { EmptyView() }, trailing: {
                    MPIconButton(systemName: model.masked ? "eye.slash" : "eye", label: model.masked ? "Hiện số tiền" : "Ẩn số tiền") {
                        model.masked.toggle()
                    }
                })
                if model.isLoading && model.wallets.isEmpty {
                    MPLoadingView(label: "Đang tải tổng quan...")
                } else if let errorMessage = model.errorMessage, model.wallets.isEmpty {
                    MPErrorView(title: "Không tải được tổng quan", detail: errorMessage) {
                        Task { await model.load() }
                    }
                } else {
                    LazyVGrid(columns: [GridItem(.flexible()), GridItem(.flexible())], spacing: 12) {
                        MPMetricCard(title: "Tổng số dư", value: MPFormat.money(model.totalBalance), masked: model.masked)
                        MPMetricCard(title: "Chi tháng này", value: MPFormat.money(model.monthExpense), masked: model.masked)
                        MPMetricCard(title: "Thu tháng này", value: MPFormat.money(model.monthIncome), masked: model.masked)
                        MPMetricCard(title: "Ví đang dùng", value: "\(model.wallets.count)")
                    }
                    MPAssistantCard(title: "Ví của bạn") {
                        if model.wallets.isEmpty {
                            MPEmptyView(title: "Chưa có ví", detail: "Tạo ví đầu tiên trong mục Ví.")
                        } else {
                            ForEach(model.wallets.prefix(5)) { wallet in
                                MPWalletRow(name: wallet.name, type: wallet.type, balance: model.masked ? "••••••" : MPFormat.money(wallet.currentBalance))
                            }
                        }
                    }
                    MPAssistantCard(title: "Giao dịch gần đây") {
                        if model.transactions.isEmpty {
                            MPEmptyView(title: "Chưa có giao dịch", detail: "Thêm giao dịch để bắt đầu theo dõi.")
                        } else {
                            ForEach(model.transactions.prefix(5)) { transaction in
                                MPTransactionRow(
                                    title: transaction.note?.isEmpty == false ? transaction.note! : (transaction.type == "income" ? "Thu nhập" : "Chi tiêu"),
                                    subtitle: MPFormat.shortDate(transaction.occurredAt, timeZone: TimeZone(identifier: model.accountTimezone) ?? .current),
                                    amount: model.masked ? "••••••" : MPFormat.money(transaction.amount),
                                    isIncome: transaction.type == "income"
                                )
                            }
                            MPButton("Xem toàn bộ giao dịch", style: .secondary, action: onOpenTransactions)
                        }
                    }
                }
            }
            .padding(MPMetrics.pagePadding)
        }
        .background(MPColor.canvas.ignoresSafeArea())
        .environment(\.timeZone, TimeZone(identifier: model.accountTimezone) ?? .current)
        .task { await model.load() }
        .refreshable { await model.load() }
        .accessibilityIdentifier("overview-screen")
    }
}
