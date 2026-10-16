import SwiftUI
import Combine

enum TransactionEditorPolicy {
    static func applicableCategories(type: String, walletID: String, walletType: String, categories: [CategoryModel]) -> [CategoryModel] {
        categories.filter { $0.applies(to: type, walletID: walletID, walletType: walletType) }
    }

    static func requiredCategoryID(type: String, walletID: String, walletType: String, categories: [CategoryModel], currentID: String) -> String {
        let valid = applicableCategories(type: type, walletID: walletID, walletType: walletType, categories: categories)
        if valid.contains(where: { $0.id == currentID }) { return currentID }
        if let retained = categories.first(where: { $0.id == currentID && $0.isActive == false }),
           retained.appliesIgnoringActivity(to: type, walletID: walletID, walletType: walletType) {
            return retained.id
        }
        return valid.first?.id ?? ""
    }

    static func canUseJar(type: String, category: CategoryModel?, walletType: String) -> Bool {
        type == "expense"
            && category?.systemKey != "expense_transfer_out"
            && (walletType == "basic" || walletType == "goal")
    }

    static func jarOptions(
        summary: JarMonthSummaryModel?,
        selectedJarID: String,
        selectedJarName: String?
    ) -> [MPSelectionOption] {
        var options = [MPSelectionOption(id: "", label: "Không chọn")]
        guard let summary else { return options }
        let activeIDs = Set(summary.items.filter(\.active).map(\.jarID))
        options += summary.jars
            .filter { activeIDs.contains($0.jarID) }
            .map { MPSelectionOption(id: $0.jarID, label: $0.name) }
        if !selectedJarID.isEmpty, !activeIDs.contains(selectedJarID) {
            let name = summary.jars.first(where: { $0.jarID == selectedJarID })?.name
                ?? selectedJarName
                ?? "Hũ không có cấu hình tháng"
            options.append(MPSelectionOption(id: selectedJarID, label: "\(name) (đã gỡ khỏi tháng)"))
        }
        return options
    }

    static func hasActiveJar(_ jarID: String, summary: JarMonthSummaryModel?) -> Bool {
        guard !jarID.isEmpty, let summary else { return false }
        return summary.items.contains { $0.jarID == jarID && $0.active }
    }
}

enum TransactionDetailPolicy {
    static func canEditOrdinary(_ transaction: TransactionModel, isCreditWallet: Bool) -> Bool {
        !isCreditWallet && transaction.transferID == nil && transaction.type != "adjustment" && transaction.creditKind == nil
    }

    static func canLinkTravel(_ transaction: TransactionModel, isCreditWallet: Bool) -> Bool {
        canEditOrdinary(transaction, isCreditWallet: isCreditWallet)
    }

    static func actionKind(_ transaction: TransactionModel, isCreditWallet: Bool) -> String {
        if isCreditWallet || transaction.creditKind != nil { return "credit" }
        if transaction.transferID != nil { return "transfer" }
        if transaction.type == "adjustment" { return "adjustment" }
        return "ordinary"
    }
}

enum TransactionsPresentation {
    struct DayGroup: Equatable {
        let dayNumber: String
        let weekday: String
        let monthYear: String
    }

    static func scopeBalance(wallets: [WalletModel], selectionID: String) -> Int64? {
        guard !wallets.isEmpty else { return nil }
        if selectionID == "__all__" {
            return MPWalletManagementPolicy.totalBalance(from: wallets)
        }
        return wallets.first(where: { $0.id == selectionID })?.currentBalance
    }

    static func dayHeading(_ key: String, timeZone: TimeZone, now: Date = Date()) -> String {
        if key == MPFormat.calendarDateString(now, timeZone: timeZone) { return "Hôm nay" }
        return dayGroup(key, timeZone: timeZone)?.weekday ?? key
    }

    static func categoryRowDate(_ occurredAt: String, timeZone: TimeZone, now: Date = Date()) -> String {
        guard let instant = MPFormat.parseInstant(occurredAt) else { return MPFormat.shortDate(occurredAt, timeZone: timeZone) }
        let key = MPFormat.calendarDateString(instant, timeZone: timeZone)
        guard let day = dayGroup(key, timeZone: timeZone) else { return key }
        return "\(day.dayNumber) \(day.monthYear), \(dayHeading(key, timeZone: timeZone, now: now))"
    }

    static func dayGroup(_ key: String, timeZone: TimeZone) -> DayGroup? {
        guard let date = MPFormat.parseCalendarDate(key, timeZone: timeZone) else { return nil }
        func formatter(_ format: String) -> DateFormatter {
            let formatter = DateFormatter()
            formatter.calendar = Calendar(identifier: .gregorian)
            formatter.locale = Locale(identifier: "vi_VN")
            formatter.timeZone = timeZone
            formatter.dateFormat = format
            return formatter
        }
        return DayGroup(
            dayNumber: formatter("d").string(from: date),
            weekday: formatter("EEEE").string(from: date),
            monthYear: formatter("MMMM yyyy").string(from: date)
        )
    }

    static func newestWeekOffset(
        transactions: [TransactionModel],
        walletSelection: String,
        accountTimezone: String,
        now: Date = Date()
    ) -> Int {
        let timeZone = TimeZone(identifier: accountTimezone) ?? .current
        let currentWeekStart = MPFormat.weekRange(now: now, timeZone: timeZone).start
        var calendar = Calendar(identifier: .gregorian)
        calendar.timeZone = timeZone

        return transactions
            .filter { walletSelection == "__all__" || $0.walletID == walletSelection }
            .compactMap { transaction -> Int? in
                guard let date = MPFormat.parseInstant(transaction.occurredAt) else { return nil }
                let transactionWeekStart = MPFormat.weekRange(now: date, timeZone: timeZone).start
                guard transactionWeekStart <= currentWeekStart else { return nil }
                let days = calendar.dateComponents([.day], from: currentWeekStart, to: transactionWeekStart).day ?? 0
                return min(0, days / 7)
            }
            .max() ?? 0
    }
}

enum MPWalletScopePreference {
    private static let keyPrefix = "mypocket.transaction-wallet-scope."

    static func load(ownerID: String?, defaults: UserDefaults = .standard) -> String? {
        guard let key = preferenceKey(ownerID: ownerID),
              let walletID = defaults.string(forKey: key),
              !walletID.isEmpty,
              walletID != "__all__" else { return nil }
        return walletID
    }

    static func store(_ walletID: String, ownerID: String?, defaults: UserDefaults = .standard) {
        guard let key = preferenceKey(ownerID: ownerID) else { return }
        if walletID.isEmpty || walletID == "__all__" {
            defaults.removeObject(forKey: key)
        } else {
            defaults.set(walletID, forKey: key)
        }
    }

    private static func preferenceKey(ownerID: String?) -> String? {
        guard let ownerID = ownerID?.trimmingCharacters(in: .whitespacesAndNewlines),
              !ownerID.isEmpty else { return nil }
        return keyPrefix + ownerID
    }
}

@MainActor
final class TransactionsViewModel: ObservableObject {
    @Published private(set) var transactions: [TransactionModel] = []
    @Published private(set) var wallets: [WalletModel] = []
    @Published private(set) var categories: [CategoryModel] = []
    @Published private(set) var isLoading = false
    @Published private(set) var isBulkDeleting = false
    @Published private(set) var bulkDeleteErrorMessage: String?
    @Published private(set) var errorMessage: String?
    @Published private(set) var cachedAt: Date?
    @Published private(set) var cacheWarning: String?
    @Published var walletSelection: String {
        didSet { MPWalletScopePreference.store(walletSelection, ownerID: ownerID, defaults: preferences) }
    }
    @Published var periodSelection = "week"
    @Published var weekOffset = 0
    @Published var groupByCategory = false
    @Published var customStart = Calendar.current.date(byAdding: .day, value: -30, to: Date()) ?? Date()
    @Published var customEnd = Date()
    @Published private(set) var accountTimezone = TimeZone.current.identifier

    private let api: FinanceAPI
    private let ownerID: String?
    private let snapshotCache: FinancialSnapshotCache?
    private let preferences: UserDefaults
    private let bulkDeleteOperation: @Sendable ([String]) async throws -> Void
    private var hasResolvedInitialWeek = false

    init(
        api: FinanceAPI,
        initialWalletID: String? = nil,
        ownerID: String? = nil,
        snapshotCache: FinancialSnapshotCache? = nil,
        preferences: UserDefaults = .standard,
        bulkDeleteOperation: (@Sendable ([String]) async throws -> Void)? = nil
    ) {
        self.api = api
        self.ownerID = ownerID
        self.snapshotCache = snapshotCache
        self.preferences = preferences
        self.bulkDeleteOperation = bulkDeleteOperation ?? { ids in
            try await api.deleteTransactions(ids: ids)
        }
        let routeWalletID = initialWalletID?.isEmpty == false ? initialWalletID : nil
        self.walletSelection = routeWalletID ?? MPWalletScopePreference.load(ownerID: ownerID, defaults: preferences) ?? "__all__"
        if let routeWalletID {
            MPWalletScopePreference.store(routeWalletID, ownerID: ownerID, defaults: preferences)
        }
    }

    var filteredTransactions: [TransactionModel] {
        Self.applyFilters(
            transactions,
            walletSelection: walletSelection,
            periodSelection: periodSelection,
            customStart: customStart,
            customEnd: customEnd,
            accountTimezone: accountTimezone,
            weekOffset: weekOffset
        )
    }

    static func applyFilters(
        _ transactions: [TransactionModel],
        walletSelection: String,
        periodSelection: String,
        customStart: Date,
        customEnd: Date,
        accountTimezone: String,
        weekOffset: Int = 0,
        now: Date = Date()
    ) -> [TransactionModel] {
        let walletFiltered = walletSelection == "__all__" ? transactions : transactions.filter { $0.walletID == walletSelection }
        if periodSelection == "all" { return walletFiltered.sorted { $0.occurredAt > $1.occurredAt } }
        let timeZone = TimeZone(identifier: accountTimezone) ?? .current
        var calendar = Calendar(identifier: .gregorian)
        calendar.timeZone = timeZone
        let start: Date
        let end: Date
        if periodSelection == "week" {
            let range = MPFormat.weekRange(now: now, offset: weekOffset, timeZone: timeZone)
            start = range.start
            end = range.endExclusive
        } else {
            start = calendar.startOfDay(for: customStart)
            end = calendar.date(byAdding: .day, value: 1, to: calendar.startOfDay(for: customEnd)) ?? Date()
        }
        guard end > start else { return [] }
        return walletFiltered.filter {
            guard let date = MPFormat.parseInstant($0.occurredAt) else { return false }
            return date >= start && date < end
        }.sorted { $0.occurredAt > $1.occurredAt }
    }

    var totalExpense: Int64 { filteredTransactions.filter { $0.type == "expense" }.reduce(0) { $0 + $1.amount } }
    var totalIncome: Int64 { filteredTransactions.filter { $0.type == "income" }.reduce(0) { $0 + $1.amount } }
    var weekRangeLabel: String {
        let timeZone = TimeZone(identifier: accountTimezone) ?? .current
        let range = MPFormat.weekRange(offset: weekOffset, timeZone: timeZone)
        return MPFormat.weekRangeLabel(start: range.start, endExclusive: range.endExclusive, timeZone: timeZone)
    }
    var customRangeIsInvalid: Bool { customEnd < customStart }
    var selectedWalletType: String? {
        guard walletSelection != "__all__" else { return nil }
        return wallets.first(where: { $0.id == walletSelection })?.type
    }
    static func bulkDeleteSelectable(_ transaction: TransactionModel) -> Bool {
        transaction.transferID == nil && transaction.type != "adjustment" && transaction.creditKind == nil
    }

    static func ordinaryTransactionWallets(_ wallets: [WalletModel]) -> [WalletModel] {
        wallets.filter { $0.type != "credit" }
    }
    var groupedTransactions: [(String, [TransactionModel])] {
        var calendar = Calendar(identifier: .gregorian)
        calendar.timeZone = TimeZone(identifier: accountTimezone) ?? .current
        let formatter = DateFormatter()
        formatter.calendar = calendar
        formatter.timeZone = calendar.timeZone
        formatter.locale = Locale(identifier: "en_US_POSIX")
        formatter.dateFormat = "yyyy-MM-dd"
        if groupByCategory {
            return Dictionary(grouping: filteredTransactions) { transaction in
                transaction.categoryID ?? transaction.type
            }
            .sorted { $0.key < $1.key }
            .map { ($0.key, $0.value.sorted { $0.occurredAt > $1.occurredAt }) }
        }
        return Dictionary(grouping: filteredTransactions) {
            guard let date = MPFormat.parseInstant($0.occurredAt) else { return String($0.occurredAt.prefix(10)) }
            return formatter.string(from: date)
        }
        .sorted { $0.key > $1.key }
        .map { ($0.key, $0.value.sorted { $0.occurredAt > $1.occurredAt }) }
    }

    func groupLabel(_ key: String) -> String {
        if groupByCategory {
            if let category = categories.first(where: { $0.id == key }) { return category.name }
            switch key {
            case "income": return "Khoản thu"
            case "expense": return "Khoản chi"
            case "adjustment": return "Điều chỉnh số dư"
            default: return "Không phân loại"
            }
        }
        let timeZone = TimeZone(identifier: accountTimezone) ?? .current
        let today = MPFormat.todayISO(timeZone: timeZone)
        if key == today { return "Hôm nay" }
        if let todayDate = MPFormat.parseCalendarDate(today) {
            var calendar = Calendar(identifier: .gregorian)
            calendar.timeZone = timeZone
            if let yesterday = calendar.date(byAdding: .day, value: -1, to: todayDate), MPFormat.calendarDateString(yesterday) == key { return "Hôm qua" }
        }
        return MPFormat.shortDate(key, timeZone: timeZone)
    }

    func categoryName(for transaction: TransactionModel) -> String? {
        categories.first(where: { $0.id == transaction.categoryID })?.name
    }

    func walletName(for transaction: TransactionModel) -> String {
        wallets.first(where: { $0.id == transaction.walletID })?.name ?? "Ví đã xóa"
    }

    func signedAmount(for transaction: TransactionModel) -> Int64 {
        if transaction.type == "income" { return transaction.amount }
        if transaction.type == "adjustment" {
            return transaction.adjustmentDirection == "increase" ? transaction.amount : -transaction.amount
        }
        return -transaction.amount
    }

    func load() async {
        guard !isLoading else { return }
        isLoading = true
        errorMessage = nil
        cacheWarning = nil
        defer { isLoading = false }
        do {
            self.transactions = try await api.transactions(ownerID: ownerID)
        } catch {
            await recoverLoad(error, message: "Không thể tải giao dịch.")
            return
        }
        do {
            self.wallets = try await api.wallets(ownerID: ownerID)
        } catch {
            await recoverLoad(error, message: "Không thể tải danh sách ví.")
            return
        }
        if walletSelection != "__all__", !wallets.contains(where: { $0.id == walletSelection }) {
            walletSelection = "__all__"
        }
        var referencesComplete = true
        do {
            self.categories = try await api.categories(ownerID: ownerID)
        } catch {
            if SyncPolicy.isAuthFailure(error) { await recoverLoad(error, message: "Cần đăng nhập lại để tải dữ liệu."); return }
            referencesComplete = false
            // Category labels are supplementary metadata for historical rows; the
            // ordinary editor still requires an applicable category on save.
            self.categories = []
        }
        do {
            let profile = try await api.profile(ownerID: ownerID)
            accountTimezone = profile.timezone
        } catch {
            if SyncPolicy.isAuthFailure(error) { await recoverLoad(error, message: "Cần đăng nhập lại để tải dữ liệu."); return }
            referencesComplete = false
        }
        cachedAt = nil
        if referencesComplete, let snapshotCache, let ownerID {
            do {
                try await snapshotCache.save(FinancialSnapshot(ownerID: ownerID, fetchedAt: Date(), transactions: transactions,
                    wallets: wallets, categories: categories, timezone: accountTimezone))
            } catch { cacheWarning = "Đã tải dữ liệu, nhưng chưa lưu được bản xem khi mất mạng." }
        }
        resolveInitialWeek()
    }

    private func recoverLoad(_ error: Error, message: String) async {
        if FinancialSnapshotCache.allowsFallback(error), let snapshotCache, let ownerID {
            do {
                if let snapshot = try await snapshotCache.load(ownerID: ownerID) {
                    transactions = snapshot.transactions; wallets = snapshot.wallets; categories = snapshot.categories
                    accountTimezone = snapshot.timezone; cachedAt = snapshot.fetchedAt
                    errorMessage = nil
                    resolveInitialWeek()
                    return
                }
            } catch { cacheWarning = "Không thể đọc bản dữ liệu đã lưu trên thiết bị." }
        }
        transactions = []; wallets = []; categories = []; cachedAt = nil
        errorMessage = message
    }

    private func resolveInitialWeek() {
        if !hasResolvedInitialWeek, periodSelection == "week" {
            weekOffset = TransactionsPresentation.newestWeekOffset(
                transactions: transactions,
                walletSelection: walletSelection,
                accountTimezone: accountTimezone
            )
            hasResolvedInitialWeek = true
        }
    }

    func delete(_ transaction: TransactionModel) async {
        do {
            try await api.deleteTransaction(id: transaction.id)
            transactions.removeAll { $0.id == transaction.id }
        } catch {
            errorMessage = "Không thể xóa giao dịch."
        }
    }

    func deleteMany(_ ids: [String]) async -> Bool {
        bulkDeleteErrorMessage = nil
        let uniqueIDs = Array(Set(ids))
        guard !uniqueIDs.isEmpty, uniqueIDs.count <= 100 else {
            bulkDeleteErrorMessage = "Chỉ có thể chọn từ 1 đến 100 giao dịch."
            return false
        }
        isBulkDeleting = true
        defer { isBulkDeleting = false }
        do {
            try await bulkDeleteOperation(uniqueIDs.sorted())
            transactions.removeAll { uniqueIDs.contains($0.id) }
            return true
        } catch {
            bulkDeleteErrorMessage = "Không thể xóa hàng loạt. Không dòng nào bị xóa."
            return false
        }
    }

    func deleteTransfer(_ transaction: TransactionModel) async {
        guard let transferID = transaction.transferID else { return }
        do {
            try await api.deleteTransfer(id: transferID)
            transactions.removeAll { $0.transferID == transferID }
        } catch {
            errorMessage = "Không thể xóa cặp giao dịch chuyển ví."
        }
    }
}

struct TransactionsView: View {
    @Environment(\.mpFinancialQueue) private var financialQueue
    @StateObject private var model: TransactionsViewModel
    @Binding private var addRequested: Bool
    @Environment(\.mpFinancialValuesMasked) private var masked
    @Environment(\.dynamicTypeSize) private var dynamicTypeSize
    let onAIEntry: (() -> Void)?
    let onOpenReports: (() -> Void)?
    @State private var editorPresented = false
    @State private var transferPresented = false
    @State private var transferEditID: String?
    @State private var adjustmentPresented = false
    @State private var selectedTransaction: TransactionModel?
    @State private var editTarget: TransactionModel?
    @State private var queuedEditTarget: OutboxMutation?
    @State private var deleteTarget: TransactionModel?
    @State private var selectionMode = false
    @State private var selectedIDs = Set<String>()
    @State private var bulkDeleteConfirmation = false

    init(api: FinanceAPI, initialWalletID: String? = nil, ownerID: String? = nil, snapshotCache: FinancialSnapshotCache? = nil, addRequested: Binding<Bool> = .constant(false), onAIEntry: (() -> Void)? = nil, onOpenReports: (() -> Void)? = nil) {
        _model = StateObject(wrappedValue: TransactionsViewModel(api: api, initialWalletID: initialWalletID, ownerID: ownerID, snapshotCache: snapshotCache))
        _addRequested = addRequested
        self.onAIEntry = onAIEntry
        self.onOpenReports = onOpenReports
    }

    var body: some View {
        let selectedWallet = model.wallets.first(where: { $0.id == model.walletSelection })
        Group {
            if let selectedWallet, selectedWallet.type == "goal" {
                // Goal detail owns its own savings-ledger ScrollView and fixed
                // CTA. Keep it out of the transactions screen's outer scroller
                // so its safe-area inset is anchored to the visible viewport.
                AnyView(WalletDetailView(
                    api: modelAPI,
                    wallet: selectedWallet,
                    embedded: true,
                    onAIEntry: onAIEntry,
                    onOpenReports: onOpenReports,
                    onSelectWallet: { model.walletSelection = $0 ?? "__all__" },
                    onBack: { model.walletSelection = "__all__" }
                ))
            } else {
                AnyView(ScrollView {
                    VStack(alignment: .leading, spacing: MPMetrics.sectionSpacing) {
                if let financialQueue {
                    MPFinancialQueueStatus(queue: financialQueue, onEdit: { item in
                        queuedEditTarget = item; editorPresented = true
                    }, onDiscarded: { Task { await model.load() } })
                }
                if model.selectedWalletType != "goal" {
                    MPScreenHeader(title: "", leading: {
                        if model.selectedWalletType != "credit" {
                            MPIconButton(systemName: "plus", label: "Thêm giao dịch") { editTarget = nil; editorPresented = true }
                        }
                    }, trailing: {
                        MPWalletScopeSelector(wallets: model.wallets, selectionID: $model.walletSelection)
                        if model.selectedWalletType != "credit" {
                            MPOverflowMenu(label: "Tùy chọn giao dịch", actions: transactionMenuActions)
                        }
                    })
                }
                if model.selectedWalletType != "goal", model.selectedWalletType != "credit" {
                    if let balance = TransactionsPresentation.scopeBalance(wallets: model.wallets, selectionID: model.walletSelection) {
                        VStack(spacing: 4) {
                            MPText("Số dư", style: .secondary)
                            MPText(MPFormat.money(balance, masked: masked), style: .emphasis)
                        }
                        .frame(maxWidth: .infinity)
                        .accessibilityElement(children: .combine)
                        .accessibilityIdentifier("transactions-scope-balance")
                    }
                    MPSegmentedControl(options: [("week", "Theo tuần"), ("custom", "Tùy chọn")], selection: $model.periodSelection)
                    if model.periodSelection == "week" {
                        MPWeekNavigator(
                            rangeLabel: model.weekRangeLabel,
                            isNextDisabled: model.weekOffset >= 0,
                            onPrevious: { model.weekOffset -= 1 },
                            onNext: { model.weekOffset = min(model.weekOffset + 1, 0) }
                        )
                    }
                }
                if model.selectedWalletType != "goal", model.selectedWalletType != "credit", model.periodSelection == "custom" {
                    MPDateRangePicker(start: $model.customStart, end: $model.customEnd, timeZone: model.accountTimezone)
                    if model.customRangeIsInvalid {
                        MPStatusMessage(text: "Ngày kết thúc phải sau hoặc bằng ngày bắt đầu.", tone: .danger)
                    }
                }
                if let selectedWallet, selectedWallet.type == "credit" {
                    CreditView(api: modelAPI, onBack: { model.walletSelection = "__all__" }, preferredWalletID: selectedWallet.id)
                } else {
                    if selectionMode {
                    MPSurfaceCard(style: .reference) {
                        VStack(alignment: .leading, spacing: 8) {
                            HStack {
                                MPText("Đã chọn \(selectedIDs.count)/100")
                                Spacer()
                                MPButton("Xóa", style: .destructive, isLoading: model.isBulkDeleting, isDisabled: selectedIDs.isEmpty || selectedIDs.count > 100) {
                                    bulkDeleteConfirmation = true
                                }
                            }
                            MPText("Chỉ giao dịch thu/chi thường được chọn; chuyển ví, tín dụng và điều chỉnh số dư phải dùng luồng riêng.", style: .secondary)
                            if let bulkDeleteErrorMessage = model.bulkDeleteErrorMessage {
                                MPStatusMessage(text: bulkDeleteErrorMessage, tone: .danger)
                            }
                            if selectedIDs.count > 100 {
                                MPStatusMessage(text: "Không thể chọn quá 100 giao dịch.", tone: .danger)
                            }
                        }
                    }
                    }
                    if model.isLoading && model.transactions.isEmpty {
                    MPLoadingView(label: "Đang tải giao dịch...")
                    } else if let errorMessage = model.errorMessage, model.transactions.isEmpty {
                    MPErrorView(title: "Không tải được giao dịch", detail: errorMessage) { Task { await model.load() } }
                    } else {
                    if let cachedAt = model.cachedAt {
                        MPStatusMessage(text: "Dữ liệu đã lưu · \(cachedAt.formatted(date: .abbreviated, time: .shortened)). Chưa cập nhật từ máy chủ.")
                            .accessibilityIdentifier("transactions-stale-data")
                        MPButton("Tải lại", style: .secondary, isLoading: model.isLoading) { Task { await model.load() } }
                    }
                    if let warning = model.cacheWarning { MPStatusMessage(text: warning) }
                    VStack(alignment: .leading, spacing: 4) {
                        MPMetricRow(title: "Tiền vào", value: MPFormat.signedMoney(model.totalIncome, masked: masked), masked: masked, tone: .income)
                        MPMetricRow(title: "Tiền ra", value: MPFormat.signedMoney(-model.totalExpense, masked: masked), masked: masked, tone: .expense)
                        MPDivider()
                        MPMetricRow(title: "Chênh lệch", value: MPFormat.signedMoney(model.totalIncome - model.totalExpense, masked: masked), masked: masked)
                        if let onOpenReports {
                            MPTextLinkButton("Xem báo cáo", action: onOpenReports)
                                .frame(maxWidth: .infinity)
                                .accessibilityIdentifier("transactions-open-reports")
                        }
                    }
                    .accessibilityElement(children: .contain)
                    .accessibilityIdentifier("transactions-cashflow-summary")
                    if model.filteredTransactions.isEmpty {
                        MPEmptyView(title: "Không có giao dịch", detail: "Thử đổi khoảng thời gian hoặc thêm giao dịch mới.")
                    } else {
                        ForEach(model.groupedTransactions, id: \.0) { day, transactions in
                            MPSurfaceCard(style: .reference) {
                                VStack(alignment: .leading, spacing: 0) {
                                    transactionGroupHeader(day, transactions: transactions)
                                        .padding(.bottom, 8)
                                    ForEach(transactions) { transaction in
                                        transactionRow(transaction)
                                            .padding(.vertical, 4)
                                    }
                                }
                            }
                            .accessibilityIdentifier("transactions-group-\(day)")
                        }
                    }
                }
                }
                }
                .padding(MPMetrics.pagePadding)
                })
            }
        }
        .background(MPColor.canvas.ignoresSafeArea())
        .environment(\.timeZone, TimeZone(identifier: model.accountTimezone) ?? .current)
        .task {
            if addRequested {
                editTarget = nil
                editorPresented = true
                addRequested = false
            }
            await model.load()
        }
        .onChange(of: model.walletSelection) { _, newValue in
            let selectedType = model.wallets.first(where: { $0.id == newValue })?.type
            model.periodSelection = selectedType == "goal" || selectedType == "credit" ? "all" : "week"
            model.weekOffset = TransactionsPresentation.newestWeekOffset(
                transactions: model.transactions,
                walletSelection: newValue,
                accountTimezone: model.accountTimezone
            )
            model.groupByCategory = false
            selectedIDs.removeAll()
        }
        .onChange(of: model.periodSelection) { _, newValue in
            if newValue == "week" { model.weekOffset = 0 }
        }
        .onChange(of: addRequested) { _, requested in
            guard requested else { return }
            editTarget = nil
            editorPresented = true
            addRequested = false
        }
        // The reference add/edit flow is a modal form surface. Present it
        // above the shell so the persistent bottom navigation cannot occlude
        // the save footer or make the editor look like part of the ledger.
        .mpFullScreenCover(isPresented: $editorPresented) {
            if let queuedEditTarget {
                switch queuedEditTarget.payload {
                case .transaction:
                    TransactionEditorView(api: modelAPI, wallets: model.wallets, categories: model.categories, timezone: model.accountTimezone, queuedMutation: queuedEditTarget,
                        onSaved: closeQueuedEditor, onCancel: closeQueuedEditor)
                case .transfer:
                    TransferEditorView(api: modelAPI, wallets: model.wallets, timezone: model.accountTimezone, queuedMutation: queuedEditTarget,
                        onSaved: closeQueuedEditor, onCancel: closeQueuedEditor)
                }
            } else {
            TransactionEditorView(api: modelAPI, wallets: model.wallets, categories: model.categories, timezone: model.accountTimezone, transaction: editTarget) {
                editorPresented = false
                editTarget = nil
                Task { await model.load() }
            } onCancel: {
                editorPresented = false
                editTarget = nil
            } onAIEntry: {
                editorPresented = false
                editTarget = nil
                onAIEntry?()
            }
            .background(MPColor.canvas.ignoresSafeArea())
            }
        }
        .refreshable { await model.load() }
        .onReceive(financialQueue?.$revision.dropFirst().eraseToAnyPublisher() ?? Empty<Int, Never>().eraseToAnyPublisher()) { _ in
            Task { await model.load() }
        }
        .overlay {
            if transferPresented {
                TransferEditorView(api: modelAPI, wallets: model.wallets, timezone: model.accountTimezone) {
                    transferPresented = false
                    Task { await model.load() }
                } onCancel: { transferPresented = false }
                .background(MPColor.canvas.ignoresSafeArea())
            } else if adjustmentPresented {
                AdjustmentEditorView(api: modelAPI, wallets: model.wallets, timezone: model.accountTimezone) {
                    adjustmentPresented = false
                    Task { await model.load() }
                } onCancel: { adjustmentPresented = false }
                .background(MPColor.canvas.ignoresSafeArea())
            } else if let transferEditID {
                TransferEditorView(api: modelAPI, wallets: model.wallets, timezone: model.accountTimezone, transactions: model.transactions, existingTransferID: transferEditID) {
                    self.transferEditID = nil
                    Task { await model.load() }
                } onCancel: { self.transferEditID = nil }
                .background(MPColor.canvas.ignoresSafeArea())
            } else if let selectedTransaction {
                TransactionDetailView(api: modelAPI, transaction: selectedTransaction, allTransactions: model.transactions, isCreditWallet: model.wallets.first(where: { $0.id == selectedTransaction.walletID })?.type == "credit", onEdit: {
                    editTarget = selectedTransaction
                    editorPresented = true
                    self.selectedTransaction = nil
                }, onEditTransfer: {
                    transferEditID = selectedTransaction.transferID
                    self.selectedTransaction = nil
                }, onDelete: {
                    deleteTarget = selectedTransaction
                    self.selectedTransaction = nil
                }, onDeleteTransfer: {
                    deleteTarget = selectedTransaction
                    self.selectedTransaction = nil
                }, onClose: { self.selectedTransaction = nil })
                .background(MPColor.canvas.ignoresSafeArea())
            }
        }
        .mpConfirmDialog(isPresented: Binding(get: { deleteTarget != nil }, set: { if !$0 { deleteTarget = nil } }), title: "Xóa giao dịch?", message: "Hành động này không thể hoàn tác.", confirmTitle: "Xóa") {
            if let target = deleteTarget { Task { if target.transferID != nil { await model.deleteTransfer(target) } else { await model.delete(target) } } }
            deleteTarget = nil
        }
        .mpConfirmDialog(isPresented: $bulkDeleteConfirmation, title: "Xóa giao dịch đã chọn?", message: "Các dòng hợp lệ sẽ bị xóa nguyên tử; nếu một dòng không hợp lệ, không dòng nào bị xóa.", confirmTitle: "Xóa tất cả") {
            Task {
                if await model.deleteMany(Array(selectedIDs)) {
                    selectionMode = false
                    selectedIDs.removeAll()
                }
            }
        }
    }

    private var modelAPI: FinanceAPI { model.apiForChild }

    @ViewBuilder
    private func transactionGroupHeader(_ key: String, transactions: [TransactionModel]) -> some View {
        let total = MPFormat.signedMoney(transactions.reduce(0) { $0 + model.signedAmount(for: $1) }, masked: masked)
        if model.groupByCategory {
            HStack {
                VStack(alignment: .leading, spacing: 2) {
                    MPText(model.groupLabel(key), style: .emphasis)
                    MPText("\(transactions.count) giao dịch", style: .secondary)
                }
                Spacer()
                MPText(total, style: .emphasis)
            }
        } else if let group = TransactionsPresentation.dayGroup(key, timeZone: TimeZone(identifier: model.accountTimezone) ?? .current) {
            let heading = TransactionsPresentation.dayHeading(key, timeZone: TimeZone(identifier: model.accountTimezone) ?? .current)
            VStack(alignment: .leading, spacing: 4) {
            HStack(alignment: .center, spacing: 12) {
                MPText(group.dayNumber, style: .display)
                VStack(alignment: .leading, spacing: 2) {
                    MPText(heading, style: .secondary)
                    MPText(group.monthYear, style: .secondary)
                }
                Spacer()
                if !dynamicTypeSize.isAccessibilitySize { MPText(total, style: .emphasis) }
            }
            if dynamicTypeSize.isAccessibilitySize { MPText(total, style: .emphasis) }
            }
            .accessibilityElement(children: .combine)
            .accessibilityLabel("\(group.weekday), \(group.dayNumber) \(group.monthYear), tổng \(total)")
        } else {
            HStack {
                MPText(key, style: .secondary)
                Spacer()
                MPText(total, style: .secondary)
            }
        }
    }

    private var transactionMenuActions: [MPOverflowAction] {
        var actions: [MPOverflowAction] = []
        if model.selectedWalletType != "goal", model.selectedWalletType != "credit" {
            actions.append(MPOverflowAction(title: model.groupByCategory ? "Xem theo ngày" : "Xem theo nhóm") { model.groupByCategory.toggle() })
            actions.append(MPOverflowAction(title: "Khoảng thời gian") { model.periodSelection = "custom" })
        }
        actions.append(MPOverflowAction(title: "Tải lại giao dịch") { Task { await model.load() } })
        if model.selectedWalletType != "credit" {
            actions.append(MPOverflowAction(title: selectionMode ? "Hủy chọn" : "Xóa nhiều giao dịch") {
                selectionMode.toggle()
                if !selectionMode { selectedIDs.removeAll() }
            })
            actions.append(MPOverflowAction(title: "Chuyển tiền đến ví khác") { transferPresented = true })
            actions.append(MPOverflowAction(title: "Điều chỉnh số dư") { adjustmentPresented = true })
        }
        return actions
    }

    @ViewBuilder
    private func transactionRow(_ transaction: TransactionModel) -> some View {
        let selectable = TransactionsViewModel.bulkDeleteSelectable(transaction)
        let presentedRow = MPTransactionRow(
            transaction: transaction,
            category: model.categories.first(where: { $0.id == transaction.categoryID }),
            timeZone: TimeZone(identifier: model.accountTimezone) ?? .current,
            masked: masked
        )
        let row = model.groupByCategory ? MPTransactionRow(
            title: TransactionsPresentation.categoryRowDate(transaction.occurredAt, timeZone: TimeZone(identifier: model.accountTimezone) ?? .current),
            subtitle: presentedRow.subtitle, amount: presentedRow.amount,
            iconName: presentedRow.iconName, iconTone: presentedRow.iconTone,
            amountTone: presentedRow.amountTone
        ) : presentedRow
        if selectionMode {
            MPButton(style: .plain, isDisabled: !selectable) {
                if selectable {
                    if selectedIDs.contains(transaction.id) { selectedIDs.remove(transaction.id) }
                    else { selectedIDs.insert(transaction.id) }
                }
            } label: {
                HStack(spacing: 8) {
                    MPIcon(
                        systemName: selectedIDs.contains(transaction.id) ? "checkmark.circle.fill" : "circle",
                        label: selectedIDs.contains(transaction.id) ? "Đã chọn" : "Chưa chọn",
                        color: selectable ? MPColor.action : MPColor.secondary
                    )
                    row
                }
            }
            .accessibilityIdentifier("transaction-row-\(transaction.id)")
        } else {
            MPButton(style: .plain) {
                selectedTransaction = transaction
            } label: { row }
                .contextMenu {
                    MPButton("Xóa giao dịch", style: .destructive) { deleteTarget = transaction }
                }
                .accessibilityIdentifier("transaction-row-\(transaction.id)")
        }
    }

    private func closeQueuedEditor() {
        editorPresented = false; queuedEditTarget = nil
        Task { await model.load() }
    }
}

private extension TransactionsViewModel {
    var apiForChild: FinanceAPI { api }
}

struct TransactionEditorView: View {
    @Environment(\.mpFinancialQueue) private var financialQueue
    @State private var intentKey = UUID()
    let api: FinanceAPI
    let wallets: [WalletModel]
    let categories: [CategoryModel]
    let timezone: String
    let transaction: TransactionModel?
    let queuedMutation: OutboxMutation?
    let onSaved: () -> Void
    let onCancel: () -> Void
    let onAIEntry: (() -> Void)?
    @State private var type = "expense"
    @State private var amount = ""
    @State private var note = ""
    @State private var walletID = ""
    @State private var categoryID = ""
    @State private var categoryPickerPresented = false
    @State private var jarID = ""
    @State private var includedInReports = true
    @State private var date = Date()
    @State private var jarSummary: JarMonthSummaryModel?
    @State private var jarsLoading = false
    @State private var jarLoadError: String?
    @State private var isSaving = false
    @State private var errorMessage: String?
    @State private var showDetails = false
    @State private var noteEditorPresented = false
    @State private var noteDraft = ""

    init(api: FinanceAPI, wallets: [WalletModel], categories: [CategoryModel], timezone: String = TimeZone.current.identifier, transaction: TransactionModel? = nil, queuedMutation: OutboxMutation? = nil, onSaved: @escaping () -> Void, onCancel: @escaping () -> Void, onAIEntry: (() -> Void)? = nil) {
        self.api = api
        self.wallets = wallets
        self.categories = categories
        self.timezone = timezone
        self.transaction = transaction
        self.queuedMutation = queuedMutation
        let draft: TransactionInputModel?
        if let queuedMutation, case .transaction(let input) = queuedMutation.payload { draft = input } else { draft = nil }
        self.onSaved = onSaved
        self.onCancel = onCancel
        self.onAIEntry = onAIEntry
        _type = State(initialValue: draft?.type ?? transaction?.type ?? "expense")
        _amount = State(initialValue: draft.map { String($0.amount) } ?? transaction.map { String($0.amount) } ?? "")
        _note = State(initialValue: draft?.note ?? transaction?.note ?? "")
        let initialWalletID = draft?.walletID ?? transaction?.walletID ?? TransactionsViewModel.ordinaryTransactionWallets(wallets).first?.id ?? ""
        let initialWalletType = wallets.first(where: { $0.id == initialWalletID })?.type ?? "basic"
        _walletID = State(initialValue: initialWalletID)
        _categoryID = State(initialValue: TransactionEditorPolicy.requiredCategoryID(type: draft?.type ?? transaction?.type ?? "expense", walletID: initialWalletID, walletType: initialWalletType, categories: categories, currentID: draft?.categoryID ?? transaction?.categoryID ?? ""))
        _jarID = State(initialValue: draft?.jarID ?? transaction?.jarID ?? "")
        _includedInReports = State(initialValue: draft?.includedInReports ?? transaction?.includedInReports ?? true)
        _showDetails = State(initialValue: transaction != nil || queuedMutation != nil)
        if let raw = draft?.occurredAt ?? transaction?.occurredAt, let date = MPFormat.parseInstant(raw) { _date = State(initialValue: date) }
    }

    var body: some View {
        VStack(spacing: 0) {
        ScrollView {
            VStack(alignment: .leading, spacing: 12) {
                MPReferenceScreenHeader(
                    title: queuedMutation != nil ? "Sửa bản nháp" : transaction == nil ? "Thêm giao dịch" : "Sửa giao dịch",
                    closeLabel: "Hủy",
                    onClose: onCancel
                )
                if queuedMutation != nil {
                    MPStatusMessage(text: "Hãy thay đổi thông tin rồi xác nhận. Bản cũ sẽ ngừng gửi; bản sửa dùng key mới và không sửa giao dịch đã có trên máy chủ.")
                }
                MPSurfaceCard(style: .reference) {
                    VStack(alignment: .leading, spacing: 0) {
                        MPSegmentedControl(
                            options: [("expense", "Khoản chi"), ("income", "Khoản thu")],
                            selection: $type,
                            accessibilityIdentifier: "transaction-type-selector"
                        )
                        Divider().padding(.vertical, 8)
                        MPOptionSelector(
                            title: "Ví",
                            options: walletOptions,
                            selectionID: $walletID,
                            iconName: "wallet.pass.fill",
                            accessibilityIdentifier: "transaction-wallet-row"
                        )
                        Divider()
                        MPAmountField(title: "Số tiền", text: $amount, presentation: .transaction)
                        Divider().padding(.vertical, 8)
                        MPFormRow(
                            title: "Danh mục",
                            iconName: selectedCategoryGlyph,
                            value: selectedCategory?.name ?? "Chọn danh mục",
                            accessibilityIdentifier: "transaction-category-row"
                        ) { categoryPickerPresented = true }
                        Divider()
                        MPFormRow(
                            title: "Ghi chú",
                            iconName: "text.alignleft",
                            value: note.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty ? "Thêm ghi chú" : note,
                            accessibilityIdentifier: "transaction-note-row"
                        ) {
                            noteDraft = note
                            noteEditorPresented = true
                        }
                        Divider()
                        MPDateStepper(
                            date: date,
                            timeZone: timezone,
                            onPrevious: { shiftDate(by: -1) },
                            onNext: { shiftDate(by: 1) },
                            onSelect: { date = $0 },
                            accessibilityIdentifier: "transaction-date-stepper"
                        )
                        if canUseJar {
                            Divider().padding(.vertical, 8)
                            MPOptionSelector(
                                title: "Hũ chi tiêu",
                                options: TransactionEditorPolicy.jarOptions(
                                    summary: jarSummary,
                                    selectedJarID: jarID,
                                    selectedJarName: transaction?.jarName
                                ),
                                selectionID: $jarID,
                                emptyLabel: jarsLoading ? "Đang tải hũ..." : "Không chọn",
                                iconName: "shippingbox.fill",
                                accessibilityIdentifier: "transaction-jar-row"
                            )
                            if let jarLoadError { MPStatusMessage(text: jarLoadError, tone: .danger) }
                        }
                    }
                }
                MPButton(showDetails ? "Ẩn chi tiết" : "Thêm chi tiết", style: .secondary) {
                    withAnimation(.easeInOut(duration: 0.18)) { showDetails.toggle() }
                }
                .accessibilityIdentifier("transaction-details-toggle")
                if showDetails {
                    MPSurfaceCard(style: .reference) {
                        VStack(alignment: .leading, spacing: 12) {
                            MPCheckboxRow(
                                title: "Tính vào báo cáo",
                                isSelected: includedInReports,
                                accessibilityIdentifier: "transaction-include-reports-checkbox"
                            ) {
                                includedInReports.toggle()
                            }
                        }
                    }
                }
                if let errorMessage { MPStatusMessage(text: errorMessage, tone: .danger) }
                Color.clear.frame(height: 16)
            }
            .padding(.horizontal, MPMetrics.pagePadding)
            .padding(.top, 8)
            .padding(.bottom, 24)
        }
        HStack(spacing: 12) {
            MPButton(queuedMutation != nil ? "Xác nhận bản sửa" : transaction == nil ? "Lưu giao dịch" : "Lưu thay đổi", isLoading: isSaving) {
                Task { await save() }
            }
            .accessibilityIdentifier("transaction-save-footer")
            if transaction == nil, let onAIEntry {
                MPActionIconButton(systemName: "photo.on.rectangle.angled", label: "Nhập bằng AI từ ảnh hoặc nội dung", action: onAIEntry)
                    .accessibilityIdentifier("transaction-ai-entry-button")
            }
        }
        .padding(.horizontal, MPMetrics.pagePadding)
        .padding(.vertical, 12)
        .background(MPColor.canvas)
        .safeAreaPadding(.bottom, 4)
        }
        .environment(\.timeZone, TimeZone(identifier: timezone) ?? .current)
        .background(MPColor.canvas.ignoresSafeArea())
        .task {
            if walletID.isEmpty { walletID = TransactionsViewModel.ordinaryTransactionWallets(wallets).first?.id ?? "" }
            clearInvalidCategory()
            await loadJarSummary()
        }
        .onChange(of: type) { _, _ in
            clearInvalidCategory()
            if !canUseJar { jarID = "" }
            Task { await loadJarSummary() }
        }
        .onChange(of: walletID) { _, _ in
            clearInvalidCategory()
            if !canUseJar { jarID = "" }
            Task { await loadJarSummary() }
        }
        .onChange(of: categoryID) { _, _ in
            if !canUseJar { jarID = "" }
        }
        .onChange(of: date) { _, _ in Task { await loadJarSummary() } }
        .mpFormBottomSheet(
            isPresented: $noteEditorPresented,
            title: "Ghi chú giao dịch",
            accessibilityIdentifier: "transaction-note-editor",
            onClose: { noteEditorPresented = false }
        ) {
            MPTextEditor(title: "Ghi chú", text: $noteDraft, placeholder: "Ví dụ: Ăn trưa")
        } footer: {
            MPButton("Lưu ghi chú") {
                note = noteDraft
                noteEditorPresented = false
            }
            .accessibilityIdentifier("transaction-note-save")
        }
        .mpFormBottomSheet(
            isPresented: $categoryPickerPresented,
            title: "Chọn danh mục",
            accessibilityIdentifier: "transaction-category-picker-sheet",
            onClose: { categoryPickerPresented = false },
            fields: {
                MPCategoryPicker(
                    categories: categories,
                    kind: type,
                    selectableIDs: Set(TransactionEditorPolicy.applicableCategories(
                        type: type,
                        walletID: selectedWallet?.id ?? "",
                        walletType: selectedWallet?.type ?? "basic",
                        categories: categories
                    ).map(\.id)),
                    selectedID: categoryID,
                    clearLabel: nil,
                    clearID: nil,
                    onSelect: { categoryID = $0; categoryPickerPresented = false }
                )
            },
            footer: {
                MPButton("Đóng", style: .secondary) { categoryPickerPresented = false }
            }
        )
    }

    private func save() async {
        guard !isSaving else { return }
        guard let wallet = wallets.first(where: { $0.id == walletID }), let parsed = MPFormat.parsePositiveInteger(amount) else {
            errorMessage = "Chọn ví và nhập số tiền hợp lệ."
            return
        }
        guard let selectedCategory = categories.first(where: { $0.id == categoryID }) else {
            errorMessage = "Hãy chọn nhóm cho giao dịch."
            return
        }
        if !selectedCategory.applies(to: type, walletID: wallet.id, walletType: wallet.type) {
            errorMessage = "Danh mục không phù hợp với loại giao dịch hoặc ví đã chọn."
            return
        }
        if !jarID.isEmpty {
            guard canUseJar else {
                errorMessage = "Hũ chỉ áp dụng cho khoản chi thường hoặc ví tiết kiệm."
                return
            }
            guard TransactionEditorPolicy.hasActiveJar(jarID, summary: jarSummary) else {
                errorMessage = "Hũ đã gỡ khỏi tháng này. Hãy bỏ chọn hoặc chọn hũ đang hoạt động."
                return
            }
        }
        isSaving = true
        defer { isSaving = false }
        do {
            let occurredAt = queuedTransactionInput.flatMap { MPFormat.parseInstant($0.occurredAt) == date ? $0.occurredAt : nil } ?? Self.timestamp(from: date, timezone: timezone)
            let input = TransactionInputModel(walletID: wallet.id, categoryID: selectedCategory.id, jarID: jarID.isEmpty ? nil : jarID, type: type, amount: parsed, occurredAt: occurredAt, note: note.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty ? nil : note.trimmingCharacters(in: .whitespacesAndNewlines), includedInReports: includedInReports, travelEventID: queuedTransactionInput?.travelEventID ?? transaction?.travelEventID)
            if let transaction { _ = try await api.updateTransaction(id: transaction.id, input: input) }
            else if let queuedMutation {
                guard let financialQueue else { throw APIError.missingSession }
                _ = try await financialQueue.replaceRejected(id: queuedMutation.id, payload: .transaction(input), key: intentKey)
            }
            else if let financialQueue { _ = try await financialQueue.submit(.transaction(input), key: intentKey) }
            else { _ = try await api.createTransaction(input) }
            onSaved()
        } catch {
            errorMessage = queuedMutation == nil ? "Không thể lưu giao dịch." : "Hãy thay đổi thông tin hợp lệ và kiểm tra đúng tài khoản trước khi xác nhận bản sửa."
        }
    }

    private var selectedCategory: CategoryModel? { categories.first(where: { $0.id == categoryID }) }
    private var queuedTransactionInput: TransactionInputModel? {
        guard let queuedMutation, case .transaction(let input) = queuedMutation.payload else { return nil }
        return input
    }
    private var selectedWallet: WalletModel? { wallets.first(where: { $0.id == walletID }) }
    private var canUseJar: Bool {
        TransactionEditorPolicy.canUseJar(type: type, category: selectedCategory, walletType: selectedWallet?.type ?? "")
    }
    private static func timestamp(from date: Date, timezone: String) -> String {
        let formatter = DateFormatter()
        formatter.dateFormat = "yyyy-MM-dd'T'HH:mm:ssXXXXX"
        formatter.locale = Locale(identifier: "en_US_POSIX")
        formatter.timeZone = TimeZone(identifier: timezone) ?? .current
        return formatter.string(from: date)
    }

    private func loadJarSummary() async {
        guard canUseJar else {
            jarSummary = nil
            jarsLoading = false
            jarLoadError = nil
            return
        }
        let month = String(MPFormat.calendarDateString(date, timeZone: TimeZone(identifier: timezone) ?? .current).prefix(7))
        guard month.count == 7 else { return }
        jarsLoading = true
        jarLoadError = nil
        defer { jarsLoading = false }
        do {
            jarSummary = try await api.jarMonth(month: month)
        } catch {
            jarSummary = nil
            jarLoadError = "Không thể tải hũ của tháng này."
        }
    }

    private var walletOptions: [MPSelectionOption] { TransactionsViewModel.ordinaryTransactionWallets(wallets).map { MPSelectionOption(id: $0.id, label: $0.name) } }
    private var selectedCategoryGlyph: String {
        guard !categoryID.isEmpty else { return "circle.fill" }
        let category = categories.first(where: { $0.id == categoryID })
        return MPTransactionPresentation.glyph(
            systemKey: category?.systemKey ?? category?.iconKey,
            type: type
        ).systemName
    }

    private func shiftDate(by dayDelta: Int) {
        var calendar = Calendar(identifier: .gregorian)
        calendar.timeZone = TimeZone(identifier: timezone) ?? .current
        if let shifted = calendar.date(byAdding: .day, value: dayDelta, to: date) {
            date = shifted
        }
    }

    private func clearInvalidCategory() {
        guard let selectedWallet else { return }
        categoryID = TransactionEditorPolicy.requiredCategoryID(type: type, walletID: selectedWallet.id, walletType: selectedWallet.type, categories: categories, currentID: categoryID)
    }
}

struct TransactionDetailView: View {
    let api: FinanceAPI
    let transaction: TransactionModel
    let allTransactions: [TransactionModel]
    let isCreditWallet: Bool
    let onEdit: () -> Void
    let onEditTransfer: () -> Void
    let onDelete: () -> Void
    let onDeleteTransfer: () -> Void
    let onClose: () -> Void
    @State private var travelEvents: [TravelEventModel] = []
    @State private var travelEventID = ""
    @State private var travelError: String?
    @Environment(\.timeZone) private var timeZone
    @Environment(\.mpFinancialValuesMasked) private var masked

    var body: some View {
        VStack(alignment: .leading, spacing: MPMetrics.sectionSpacing) {
            MPScreenHeader(title: "Chi tiết giao dịch", leading: { MPIconButton(systemName: "xmark", label: "Đóng", action: onClose) }, trailing: { EmptyView() })
            MPSurfaceCard(style: .reference) {
                MPMetric(title: transaction.type == "income" ? "Thu nhập" : "Chi tiêu", value: MPFormat.money(transaction.amount, masked: masked))
            }
            MPSurfaceCard(style: .reference) {
                VStack(alignment: .leading, spacing: 8) {
                    MPText("Ngày: \(MPFormat.shortDate(transaction.occurredAt, timeZone: timeZone))")
                    if let note = transaction.note, !note.isEmpty { MPText("Ghi chú: \(note)") }
                    if TransactionDetailPolicy.canLinkTravel(transaction, isCreditWallet: isCreditWallet) {
                        MPOptionSelector(
                            title: "Travel Mode",
                            options: [MPSelectionOption(id: "", label: "Không gắn chuyến")] + travelEvents.map { MPSelectionOption(id: $0.id, label: $0.name) },
                            selectionID: $travelEventID,
                            emptyLabel: "Không gắn chuyến"
                        )
                        MPButton("Lưu liên kết chuyến", style: .secondary) {
                            Task {
                                let eventID = travelEventID.isEmpty ? nil : travelEventID
                                do { _ = try await api.updateTransactionTravel(id: transaction.id, eventID: eventID); travelError = nil }
                                catch { travelError = "Không thể cập nhật liên kết chuyến." }
                            }
                        }
                        if let travelError { MPStatusMessage(text: travelError, tone: .danger) }
                    }
                    MPText(transaction.includedInReports ? "Được tính vào báo cáo" : "Không tính vào báo cáo", style: .secondary)
                }
            }
            if TransactionDetailPolicy.canEditOrdinary(transaction, isCreditWallet: isCreditWallet) {
                MPButton("Sửa giao dịch", style: .secondary, action: onEdit)
                MPButton("Xóa giao dịch", style: .destructive, action: onDelete)
            } else if transaction.type == "adjustment" {
                MPStatusMessage(text: "Điều chỉnh số dư không thể sửa hoặc xóa. Hãy ghi một điều chỉnh ngược để bù trừ.", tone: .info)
            } else if isCreditWallet {
                MPStatusMessage(text: "Bút toán tín dụng chỉ được sửa hoặc hoàn tác trong Sổ tín dụng.", tone: .info)
            } else {
                MPStatusMessage(text: "Đây là cặp chuyển ví; sửa/xóa phải áp dụng cho cả hai giao dịch.", tone: .info)
                MPButton("Sửa cả cặp chuyển ví", style: .secondary, action: onEditTransfer)
                MPButton("Xóa cả cặp chuyển ví", style: .destructive, action: onDeleteTransfer)
            }
            Spacer()
        }
        .padding(MPMetrics.pagePadding)
        .accessibilityElement(children: .contain)
        .accessibilityIdentifier("transaction-detail-overlay")
        .safeAreaInset(edge: .bottom) {
            MPBottomNavigationInset()
                .accessibilityIdentifier("transaction-detail-shell-inset")
        }
        .task {
            guard TransactionDetailPolicy.canLinkTravel(transaction, isCreditWallet: isCreditWallet) else { return }
            do {
                travelEvents = try await api.travelEvents()
                travelEventID = transaction.travelEventID ?? ""
            } catch { travelError = "Không thể tải Travel Mode." }
        }
    }
}

struct TransferEditorView: View {
    @Environment(\.mpFinancialQueue) private var financialQueue
    @State private var intentKey = UUID()
    let api: FinanceAPI
    let wallets: [WalletModel]
    let timezone: String
    let transactions: [TransactionModel]
    let existingTransferID: String?
    let queuedMutation: OutboxMutation?
    let onSaved: () -> Void
    let onCancel: () -> Void
    @State private var sourceID = ""
    @State private var destinationID = ""
    @State private var amount = ""
    @State private var note = ""
    @State private var date = Date()
    @State private var saving = false
    @State private var error: String?

    init(api: FinanceAPI, wallets: [WalletModel], timezone: String = TimeZone.current.identifier, transactions: [TransactionModel] = [], existingTransferID: String? = nil, queuedMutation: OutboxMutation? = nil, onSaved: @escaping () -> Void, onCancel: @escaping () -> Void) {
        self.api = api
        self.wallets = wallets
        self.timezone = timezone
        self.transactions = transactions
        self.existingTransferID = existingTransferID
        self.queuedMutation = queuedMutation
        let draft: TransferInputModel?
        if let queuedMutation, case .transfer(let input) = queuedMutation.payload { draft = input } else { draft = nil }
        self.onSaved = onSaved
        self.onCancel = onCancel
        let source = transactions.first(where: { $0.transferID == existingTransferID && $0.type == "expense" })
        let destination = transactions.first(where: { $0.transferID == existingTransferID && $0.type == "income" })
        let eligibleWallets = Self.transferWallets(wallets)
        _sourceID = State(initialValue: draft?.sourceWalletID ?? source?.walletID ?? eligibleWallets.first?.id ?? "")
        _destinationID = State(initialValue: draft?.destinationWalletID ?? destination?.walletID ?? eligibleWallets.dropFirst().first?.id ?? eligibleWallets.first?.id ?? "")
        _amount = State(initialValue: draft.map { String($0.amount) } ?? source.map { String($0.amount) } ?? "")
        _note = State(initialValue: draft?.note ?? source?.note ?? "")
        _date = State(initialValue: draft.flatMap { MPFormat.parseInstant($0.occurredAt) } ?? source.flatMap { MPFormat.parseInstant($0.occurredAt) } ?? Date())
    }

    var body: some View {
        ScrollView { VStack(alignment: .leading, spacing: 16) {
            MPScreenHeader(title: queuedMutation != nil ? "Sửa bản nháp chuyển ví" : existingTransferID == nil ? "Chuyển tiền" : "Sửa chuyển tiền", leading: { MPIconButton(systemName: "xmark", label: "Đóng", action: onCancel) }, trailing: { EmptyView() })
            MPText("Chuyển tiền tạo cặp giao dịch nguyên tử giữa hai ví; không tính vào báo cáo.", style: .secondary)
            if queuedMutation != nil { MPStatusMessage(text: "Hãy thay đổi thông tin rồi xác nhận bản sửa với key mới. Không sửa cặp giao dịch đã có trên máy chủ.") }
            MPOptionSelector(
                title: "Từ ví",
                options: walletOptions,
                selectionID: $sourceID,
                accessibilityIdentifier: "transfer-source-wallet",
                isEnabled: existingTransferID == nil
            )
            MPOptionSelector(
                title: "Đến ví",
                options: walletOptions,
                selectionID: $destinationID,
                accessibilityIdentifier: "transfer-destination-wallet",
                isEnabled: existingTransferID == nil
            )
            MPAmountField(title: "Số tiền", text: $amount)
            MPDateField(title: "Ngày chuyển", date: $date, timeZone: timezone)
            MPTextField(title: "Ghi chú", text: $note)
            if let error { MPStatusMessage(text: error, tone: .danger) }
            MPButton(queuedMutation != nil ? "Xác nhận bản sửa" : existingTransferID == nil ? "Xác nhận chuyển" : "Lưu thay đổi", isLoading: saving) { Task { await save() } }
        }.padding(MPMetrics.pagePadding) }
        .environment(\.timeZone, TimeZone(identifier: timezone) ?? .current)
        .safeAreaInset(edge: .bottom) {
            MPBottomNavigationInset()
        }
        .task {
            let eligibleWallets = Self.transferWallets(wallets)
            if sourceID.isEmpty { sourceID = eligibleWallets.first?.id ?? "" }
            if destinationID.isEmpty { destinationID = eligibleWallets.dropFirst().first?.id ?? eligibleWallets.first?.id ?? "" }
        }
    }

    private func save() async {
        guard !saving else { return }
        guard let from = wallets.first(where: { $0.id == sourceID }), let to = wallets.first(where: { $0.id == destinationID }), from.id != to.id, let parsed = MPFormat.parsePositiveInteger(amount) else { error = "Chọn hai ví khác nhau và nhập số tiền hợp lệ."; return }
        saving = true; defer { saving = false }
        do {
            var occurredAt = Self.timestamp(date, timezone: timezone)
            if let queuedMutation, case .transfer(let draft) = queuedMutation.payload, let original = MPFormat.parseInstant(draft.occurredAt) {
                var calendar = Calendar(identifier: .gregorian)
                calendar.timeZone = TimeZone(identifier: timezone) ?? .current
                if calendar.isDate(original, inSameDayAs: date) { occurredAt = draft.occurredAt }
            }
            let trimmedNote = note.trimmingCharacters(in: .whitespacesAndNewlines)
            if let existingTransferID { _ = try await api.updateTransfer(id: existingTransferID, input: TransferUpdateInputModel(amount: parsed, occurredAt: occurredAt, note: trimmedNote.isEmpty ? nil : trimmedNote)) }
            else {
                let input = TransferInputModel(sourceWalletID: from.id, destinationWalletID: to.id, amount: parsed, occurredAt: occurredAt, note: trimmedNote.isEmpty ? nil : trimmedNote)
                if let queuedMutation {
                    guard let financialQueue else { throw APIError.missingSession }
                    _ = try await financialQueue.replaceRejected(id: queuedMutation.id, payload: .transfer(input), key: intentKey)
                } else if let financialQueue { _ = try await financialQueue.submit(.transfer(input), key: intentKey) }
                else { _ = try await api.createTransfer(input) }
            }
            onSaved()
        }
        catch { self.error = queuedMutation == nil ? "Không thể chuyển tiền. Số dư hoặc ví có thể không hợp lệ." : "Hãy thay đổi thông tin hợp lệ và kiểm tra đúng tài khoản trước khi xác nhận bản sửa." }
    }
    private var walletOptions: [MPSelectionOption] { Self.transferWallets(wallets).map { MPSelectionOption(id: $0.id, label: $0.name) } }
    static func transferWallets(_ wallets: [WalletModel]) -> [WalletModel] { wallets.filter { $0.type != "credit" } }
    static func timestamp(_ date: Date, timezone: String) -> String {
        let zone = TimeZone(identifier: timezone) ?? .current
        var calendar = Calendar(identifier: .gregorian)
        calendar.timeZone = zone
        let components = calendar.dateComponents([.year, .month, .day], from: date)
        let localNoon = calendar.date(bySettingHour: 12, minute: 0, second: 0, of: calendar.date(from: components) ?? date) ?? date
        return MPFormat.instantString(localNoon)
    }
}

struct AdjustmentEditorView: View {
    let api: FinanceAPI
    let wallets: [WalletModel]
    let timezone: String
    let onSaved: () -> Void
    let onCancel: () -> Void
    @State private var walletID = ""
    @State private var direction = "increase"
    @State private var amount = ""
    @State private var note = ""
    @State private var saving = false
    @State private var error: String?

    init(api: FinanceAPI, wallets: [WalletModel], timezone: String = TimeZone.current.identifier, onSaved: @escaping () -> Void, onCancel: @escaping () -> Void) {
        self.api = api
        self.wallets = wallets
        self.timezone = timezone
        self.onSaved = onSaved
        self.onCancel = onCancel
    }

    var body: some View {
        ScrollView { VStack(alignment: .leading, spacing: 16) {
            MPScreenHeader(title: "Điều chỉnh số dư", leading: { MPIconButton(systemName: "xmark", label: "Đóng", action: onCancel) }, trailing: { EmptyView() })
            MPOptionSelector(title: "Ví", options: ledgerWallets.map { MPSelectionOption(id: $0.id, label: $0.name) }, selectionID: $walletID)
            MPWalletSelector(title: "Hướng", options: ["increase", "decrease"], selection: $direction)
            MPAmountField(title: "Số tiền", text: $amount)
            MPTextField(title: "Lý do", text: $note)
            if let error { MPStatusMessage(text: error, tone: .danger) }
            MPButton("Lưu điều chỉnh", isLoading: saving) { Task { await save() } }
        }.padding(MPMetrics.pagePadding) }
        .safeAreaInset(edge: .bottom) {
            MPBottomNavigationInset()
        }
        .task { if walletID.isEmpty { walletID = ledgerWallets.first?.id ?? "" } }
    }
    private var ledgerWallets: [WalletModel] { wallets.filter { $0.type != "credit" } }
    private func save() async { guard let selected = ledgerWallets.first(where: { $0.id == walletID }), let parsed = MPFormat.parsePositiveInteger(amount) else { error = "Chọn ví và nhập số tiền hợp lệ."; return }; saving = true; defer { saving = false }; do { _ = try await api.createAdjustment(AdjustmentInputModel(walletID: selected.id, amount: parsed, direction: direction, occurredAt: Self.timestamp(Date(), timezone: timezone), note: note.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty ? nil : note.trimmingCharacters(in: .whitespacesAndNewlines))); onSaved() } catch { self.error = "Không thể điều chỉnh số dư." } }

    static func timestamp(_ date: Date, timezone: String) -> String {
        let zone = TimeZone(identifier: timezone) ?? .current
        var calendar = Calendar(identifier: .gregorian)
        calendar.timeZone = zone
        let components = calendar.dateComponents([.year, .month, .day], from: date)
        let localNoon = calendar.date(bySettingHour: 12, minute: 0, second: 0, of: calendar.date(from: components) ?? date) ?? date
        return MPFormat.instantString(localNoon)
    }
}
