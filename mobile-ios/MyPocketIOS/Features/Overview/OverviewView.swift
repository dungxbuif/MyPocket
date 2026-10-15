import SwiftUI
import Combine

/// Stable composition contract for the Stitch overview hierarchy. Keeping the
/// order explicit prevents a later feature card from silently moving the
/// primary wallet/report surfaces below secondary detail content.
enum OverviewReferenceSection: String, CaseIterable {
    case wallets
    case insider
    case report
    case monthSummary
    case recentTransactions

    static let stitchOrder: [OverviewReferenceSection] = [
        .wallets,
        .insider,
        .report,
        .monthSummary,
        .recentTransactions
    ]
}

enum OverviewReferenceCopy {
    static let walletAction = "Quản lý"
}

@MainActor
final class OverviewViewModel: ObservableObject {
    @Published private(set) var wallets: [WalletModel] = []
    @Published private(set) var transactions: [TransactionModel] = []
    @Published private(set) var categories: [CategoryModel] = []
    @Published private(set) var monthSummary: MonthSummaryModel?
    @Published private(set) var insider: InsiderSummaryModel?
    @Published private(set) var isLoading = false
    @Published private(set) var monthLoading = false
    @Published private(set) var errorMessage: String?
    @Published private(set) var monthErrorMessage: String?
    @Published private(set) var accountTimezone = TimeZone.current.identifier
    @Published private(set) var cachedAt: Date?
    private let api: FinanceAPI
    private let ownerID: String?
    private let snapshotCache: FinancialSnapshotCache?

    init(api: FinanceAPI, ownerID: String? = nil, snapshotCache: FinancialSnapshotCache? = nil) {
        self.api = api
        self.ownerID = ownerID
        self.snapshotCache = snapshotCache ?? (try? FinancialSnapshotCache.makeDefault())
    }

    var timeZone: TimeZone { TimeZone(identifier: accountTimezone) ?? .current }
    var currentMonth: String { MPFormat.currentMonth(timeZone: timeZone) }
    var totalBalance: Int64 { wallets.filter(\.isInTotal).reduce(0) { $0 + $1.currentBalance } }

    func load() async {
        guard !isLoading else { return }
        isLoading = true
        errorMessage = nil
        cachedAt = nil
        defer { isLoading = false }
        do {
            async let walletResult = api.wallets()
            async let transactionResult = api.transactions()
            async let categoryResult = api.categories()
            async let profileResult = api.profile()
            wallets = try await walletResult
            transactions = try await transactionResult.sorted { $0.occurredAt > $1.occurredAt }
            categories = (try? await categoryResult) ?? []
            if let profile = try? await profileResult { accountTimezone = profile.timezone }
        } catch {
            await recoverLoad(error)
            return
        }
        await loadMonthSummary()
        await loadInsider()
        await persistSnapshot()
    }

    private func recoverLoad(_ error: Error) async {
        guard FinancialSnapshotCache.allowsFallback(error), let ownerID, let snapshotCache else {
            wallets = []; transactions = []; categories = []; monthSummary = nil; insider = nil
            errorMessage = Self.message(for: error)
            return
        }
        do {
            guard let snapshot = try await snapshotCache.load(ownerID: ownerID) else {
                throw OutboxStoreError.storageUnavailable
            }
            wallets = snapshot.wallets
            transactions = snapshot.transactions.sorted { $0.occurredAt > $1.occurredAt }
            categories = snapshot.categories
            accountTimezone = snapshot.timezone
            monthSummary = snapshot.monthSummary
            insider = snapshot.insider
            cachedAt = snapshot.fetchedAt
            errorMessage = nil
            monthErrorMessage = snapshot.monthSummary == nil ? "Chưa có bản lưu tổng kết tháng." : nil
        } catch {
            wallets = []; transactions = []; categories = []; monthSummary = nil; insider = nil; cachedAt = nil
            errorMessage = Self.message(for: error)
        }
    }

    private func persistSnapshot() async {
        guard let ownerID, let snapshotCache else { return }
        do {
            try await snapshotCache.save(FinancialSnapshot(
                ownerID: ownerID,
                fetchedAt: Date(),
                transactions: transactions,
                wallets: wallets,
                categories: categories,
                timezone: accountTimezone,
                monthSummary: monthSummary,
                insider: insider
            ))
        } catch {
            // A successful online read remains usable; cache failures are
            // surfaced only if a later offline read needs the cache.
        }
    }

    func loadInsider() async {
        do {
            insider = try await api.insider(month: currentMonth)
        } catch {
            insider = nil
        }
    }

    func loadMonthSummary() async {
        monthLoading = true
        monthErrorMessage = nil
        defer { monthLoading = false }
        do {
            monthSummary = try await api.monthSummary(month: currentMonth)
        } catch {
            monthErrorMessage = "Không tải được tổng kết tháng."
        }
    }

    static func walletSubtitle(description: String?, type: String) -> String {
        if let description = description?.trimmingCharacters(in: .whitespacesAndNewlines), !description.isEmpty {
            return description
        }
        return MPSelectionLabels.display(for: type)
    }

    var currentWeekTotals: (income: Int64, expense: Int64) {
        MPWeekReportData.totals(transactions: transactions, timeZone: timeZone)
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
    @Environment(\.mpFinancialQueue) private var financialQueue
    @StateObject private var model: OverviewViewModel
    @Binding private var masked: Bool
    let onOpenTransactions: () -> Void
    let onOpenWallets: () -> Void
    let onOpenMonth: () -> Void
    let onOpenJars: () -> Void
    let onOpenNotifications: () -> Void

    init(
        api: FinanceAPI,
        ownerID: String? = nil,
        snapshotCache: FinancialSnapshotCache? = nil,
        masked: Binding<Bool> = .constant(false),
        onOpenTransactions: @escaping () -> Void,
        onOpenWallets: @escaping () -> Void = {},
        onOpenMonth: @escaping () -> Void = {},
        onOpenJars: @escaping () -> Void = {},
        onOpenNotifications: @escaping () -> Void = {}
    ) {
        _model = StateObject(wrappedValue: OverviewViewModel(api: api, ownerID: ownerID, snapshotCache: snapshotCache))
        _masked = masked
        self.onOpenTransactions = onOpenTransactions
        self.onOpenWallets = onOpenWallets
        self.onOpenMonth = onOpenMonth
        self.onOpenJars = onOpenJars
        self.onOpenNotifications = onOpenNotifications
    }

    var body: some View {
        VStack(spacing: 0) {
            MPBalanceHeader(
                totalBalance: MPFormat.money(model.totalBalance),
                masked: masked,
                onToggleMask: { masked.toggle() },
                onNotifications: onOpenNotifications
            )
            ScrollView {
                VStack(alignment: .leading, spacing: MPMetrics.sectionSpacing) {
                    if model.isLoading && model.wallets.isEmpty {
                        MPLoadingView(label: "Đang tải tổng quan...")
                    } else if let errorMessage = model.errorMessage, model.wallets.isEmpty {
                        MPErrorView(title: "Không tải được tổng quan", detail: errorMessage) {
                            Task { await model.load() }
                        }
                    } else {
                        if let cachedAt = model.cachedAt {
                            MPStatusMessage(text: "Bản lưu \(cachedAt.formatted(date: .abbreviated, time: .shortened)) · Chưa cập nhật từ máy chủ.", tone: .info)
                        }
                        walletsCard
                        if let insider = model.insider {
                            MPMoneyInsiderCard(value: insider, masked: masked, onOpen: onOpenMonth)
                                .accessibilityIdentifier("overview-section-insider")
                        }
                        if let summary = model.monthSummary {
                            MPOverviewReportCard(summary: summary, weekIncome: model.currentWeekTotals.income, weekExpense: model.currentWeekTotals.expense, masked: masked, onOpen: onOpenMonth)
                                .accessibilityIdentifier("overview-section-report")
                        }
                        monthSummaryCard
                        recentTransactionsCard
                    }
                }
                .padding(.horizontal, MPMetrics.pagePadding)
                .padding(.top, 4)
                .padding(.bottom, 20)
            }
            .refreshable { await model.load() }
        }
        .background(MPColor.canvas.ignoresSafeArea())
        .environment(\.timeZone, model.timeZone)
        .task { await model.load() }
        .onReceive(financialQueue?.$revision.dropFirst().eraseToAnyPublisher() ?? Empty<Int, Never>().eraseToAnyPublisher()) { _ in
            Task { await model.load() }
        }
    }

    private var monthSummaryCard: some View {
        MPSurfaceCard(style: .reference) {
            VStack(alignment: .leading, spacing: 10) {
                MPSectionHeader(title: "Tổng kết tháng", actionTitle: "Xem chi tiết", action: onOpenMonth)
                if model.monthLoading {
                    MPStatusMessage(text: "Đang tải tổng kết tháng.", tone: .info)
                } else if let monthErrorMessage = model.monthErrorMessage {
                    MPStatusMessage(text: monthErrorMessage, tone: .danger)
                    MPButton("Thử lại", style: .plain, fullWidth: false) {
                        Task { await model.loadMonthSummary() }
                    }
                } else if let summary = model.monthSummary {
                    MPText(
                        "\(MPFormat.monthLabel(summary.month)) · \(summary.isComplete ? "Đã hoàn tất" : "Đang diễn ra")",
                        style: .secondary
                    )
                    HStack(spacing: 24) {
                        MPMetric(title: "Thu", value: MPFormat.money(summary.income), masked: masked, compact: true)
                        MPMetric(title: "Chi", value: MPFormat.money(summary.expense), masked: masked, compact: true)
                    }
                    if !summary.note.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
                        MPText(summary.note, style: .secondary).lineLimit(2)
                    }
                    MPButton("Hũ tháng", style: .plain, fullWidth: false, action: onOpenJars)
                }
            }
        }
        .accessibilityIdentifier("overview-section-month-summary")
    }

    private var walletsCard: some View {
        MPSurfaceCard(style: .reference) {
            VStack(alignment: .leading, spacing: MPMetrics.rowSpacing) {
                MPSectionHeader(title: "Ví của tôi", actionTitle: OverviewReferenceCopy.walletAction, action: onOpenWallets)
                if model.wallets.isEmpty {
                    MPText("Chưa có ví.", style: .secondary)
                } else {
                    ForEach(Array(model.wallets.prefix(3).enumerated()), id: \.element.id) { index, wallet in
                        if index > 0 { Divider().overlay(MPColor.line) }
                        MPWalletRow(
                            name: wallet.name,
                            type: wallet.type,
                            balance: masked ? MPFormat.maskedValue : MPFormat.money(wallet.currentBalance),
                            subtitle: OverviewViewModel.walletSubtitle(description: wallet.description, type: wallet.type),
                            isNegative: wallet.currentBalance < 0
                        )
                    }
                }
            }
        }
        .accessibilityIdentifier("overview-section-wallets")
    }

    private var recentTransactionsCard: some View {
        MPSurfaceCard(style: .reference) {
            VStack(alignment: .leading, spacing: MPMetrics.rowSpacing) {
                    MPSectionHeader(title: "Giao dịch gần đây", actionTitle: "Xem tất cả", action: onOpenTransactions)
                if model.transactions.isEmpty {
                    MPText("Chưa có giao dịch.", style: .secondary)
                } else {
                    ForEach(model.transactions.prefix(4)) { transaction in
                        let category = model.categories.first(where: { $0.id == transaction.categoryID })
                        MPTransactionRow(
                            transaction: transaction, category: category,
                            timeZone: model.timeZone, masked: masked
                        )
                        .accessibilityIdentifier("overview-transaction-\(transaction.id)")
                    }
                }
            }
        }
        .accessibilityIdentifier("overview-section-recent-transactions")
    }
}
