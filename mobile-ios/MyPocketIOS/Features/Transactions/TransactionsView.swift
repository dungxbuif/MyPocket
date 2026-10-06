import SwiftUI

enum TransactionEditorPolicy {
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

@MainActor
final class TransactionsViewModel: ObservableObject {
    @Published private(set) var transactions: [TransactionModel] = []
    @Published private(set) var wallets: [WalletModel] = []
    @Published private(set) var categories: [CategoryModel] = []
    @Published private(set) var isLoading = false
    @Published private(set) var isBulkDeleting = false
    @Published private(set) var bulkDeleteErrorMessage: String?
    @Published private(set) var errorMessage: String?
    @Published var walletSelection = "__all__"
    @Published var periodSelection = "week"
    @Published var weekOffset = 0
    @Published var groupByCategory = false
    @Published var customStart = Calendar.current.date(byAdding: .day, value: -30, to: Date()) ?? Date()
    @Published var customEnd = Date()
    @Published private(set) var accountTimezone = TimeZone.current.identifier

    private let api: FinanceAPI
    init(api: FinanceAPI) { self.api = api }

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
        defer { isLoading = false }
        do {
            self.transactions = try await api.transactions()
        } catch {
            errorMessage = "Không thể tải giao dịch."
            return
        }
        do {
            self.wallets = try await api.wallets()
        } catch {
            errorMessage = "Không thể tải danh sách ví."
            return
        }
        do {
            self.categories = try await api.categories()
        } catch {
            // Category labels are optional for the ledger; rows remain usable.
            self.categories = []
        }
        if let profile = try? await api.profile() {
            accountTimezone = profile.timezone
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
            try await api.deleteTransactions(ids: uniqueIDs)
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
    @StateObject private var model: TransactionsViewModel
    @State private var editorPresented = false
    @State private var transferPresented = false
    @State private var transferEditID: String?
    @State private var adjustmentPresented = false
    @State private var selectedTransaction: TransactionModel?
    @State private var editTarget: TransactionModel?
    @State private var deleteTarget: TransactionModel?
    @State private var selectionMode = false
    @State private var selectedIDs = Set<String>()
    @State private var bulkDeleteConfirmation = false

    init(api: FinanceAPI) { _model = StateObject(wrappedValue: TransactionsViewModel(api: api)) }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                MPScreenHeader(title: "Giao dịch", leading: { EmptyView() }, trailing: {
                    HStack(spacing: 4) {
                        if model.selectedWalletType == nil {
                            MPOverflowMenu(label: "Tùy chọn giao dịch", actions: transactionMenuActions)
                            MPIconButton(systemName: "plus", label: "Thêm giao dịch") { editTarget = nil; editorPresented = true }
                        }
                    }
                })
                if model.selectedWalletType != "goal", model.selectedWalletType != "credit" {
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
                    MPDateRangePicker(start: $model.customStart, end: $model.customEnd)
                    if model.customRangeIsInvalid {
                        MPStatusMessage(text: "Ngày kết thúc phải sau hoặc bằng ngày bắt đầu.", tone: .danger)
                    }
                }
                MPOptionSelector(title: "Ví", options: [MPSelectionOption(id: "__all__", label: "Tổng cộng")] + model.wallets.map { MPSelectionOption(id: $0.id, label: $0.name) }, selectionID: $model.walletSelection, emptyLabel: "Tổng cộng")
                if let selectedWallet = model.wallets.first(where: { $0.id == model.walletSelection }), selectedWallet.type == "credit" {
                    CreditView(api: modelAPI, onBack: { model.walletSelection = "__all__" }, preferredWalletID: selectedWallet.id)
                } else if let selectedWallet = model.wallets.first(where: { $0.id == model.walletSelection }), selectedWallet.type == "goal" {
                    WalletDetailView(api: modelAPI, wallet: selectedWallet) { model.walletSelection = "__all__" }
                } else if selectionMode {
                    MPSurfaceCard {
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
                } else {
                    if model.isLoading && model.transactions.isEmpty {
                    MPLoadingView(label: "Đang tải giao dịch...")
                    } else if let errorMessage = model.errorMessage, model.transactions.isEmpty {
                    MPErrorView(title: "Không tải được giao dịch", detail: errorMessage) { Task { await model.load() } }
                    } else {
                    MPSurfaceCard {
                        VStack(alignment: .leading, spacing: 10) {
                            HStack {
                                MPText("Tiền vào")
                                Spacer()
                                MPText(MPFormat.money(model.totalIncome), style: .title)
                            }
                            HStack {
                                MPText("Tiền ra")
                                Spacer()
                                MPText(MPFormat.money(model.totalExpense), style: .title)
                            }
                        }
                    }
                    if model.filteredTransactions.isEmpty {
                        MPEmptyView(title: "Không có giao dịch", detail: "Thử đổi khoảng thời gian hoặc thêm giao dịch mới.")
                    } else {
                        ForEach(model.groupedTransactions, id: \.0) { day, transactions in
                            MPSurfaceCard {
                                VStack(alignment: .leading, spacing: 0) {
                                    HStack {
                                        MPText(model.groupLabel(day), style: .secondary)
                                        Spacer()
                                        MPText(MPFormat.signedMoney(transactions.reduce(0) { $0 + model.signedAmount(for: $1) }), style: .secondary)
                                    }
                                        .padding(.bottom, 4)
                                    ForEach(transactions) { transaction in
                                        transactionRow(transaction)
                                    }
                                }
                            }
                        }
                    }
                }
                }
            }
            .padding(MPMetrics.pagePadding)
        }
        .background(MPColor.canvas.ignoresSafeArea())
        .environment(\.timeZone, TimeZone(identifier: model.accountTimezone) ?? .current)
        .task { await model.load() }
        .onChange(of: model.walletSelection) { _, newValue in
            model.periodSelection = model.wallets.first(where: { $0.id == newValue })?.type == "goal" || model.wallets.first(where: { $0.id == newValue })?.type == "credit" ? "all" : "week"
            model.weekOffset = 0
            model.groupByCategory = false
            selectedIDs.removeAll()
        }
        .onChange(of: model.periodSelection) { _, newValue in
            if newValue == "week" { model.weekOffset = 0 }
        }
        .refreshable { await model.load() }
        .overlay {
            if editorPresented {
                TransactionEditorView(api: modelAPI, wallets: model.wallets, categories: model.categories, timezone: model.accountTimezone, transaction: editTarget) {
                    editorPresented = false
                    editTarget = nil
                    Task { await model.load() }
                } onCancel: { editorPresented = false; editTarget = nil }
                .background(MPColor.canvas.ignoresSafeArea())
            } else if transferPresented {
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
        .accessibilityIdentifier("transactions-screen")
    }

    private var modelAPI: FinanceAPI { model.apiForChild }

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
        let title = model.categoryName(for: transaction) ?? (transaction.type == "income" ? "Khoản thu" : transaction.type == "adjustment" ? "Điều chỉnh số dư" : "Khoản chi")
        let subtitle = [model.walletName(for: transaction), transaction.note].compactMap { $0?.isEmpty == false ? $0 : nil }.joined(separator: " · ")
        let row = MPTransactionRow(
            title: title,
            subtitle: subtitle.isEmpty ? MPFormat.shortDate(transaction.occurredAt, timeZone: TimeZone(identifier: model.accountTimezone) ?? .current) : subtitle,
            amount: MPFormat.signedMoney(model.signedAmount(for: transaction)),
            isIncome: transaction.type == "income"
        )
        if selectionMode {
            MPButton(style: .plain, isDisabled: !selectable) {
                if selectable {
                    if selectedIDs.contains(transaction.id) { selectedIDs.remove(transaction.id) }
                    else { selectedIDs.insert(transaction.id) }
                }
            } label: {
                HStack(spacing: 8) {
                    Image(systemName: selectedIDs.contains(transaction.id) ? "checkmark.circle.fill" : "circle")
                        .foregroundStyle(selectable ? MPColor.action : MPColor.secondary)
                    row
                }
            }
        } else {
            MPButton(style: .plain) {
                selectedTransaction = transaction
            } label: { row }
                .contextMenu {
                    MPButton("Xóa giao dịch", style: .destructive) { deleteTarget = transaction }
                }
        }
    }
}

private extension TransactionsViewModel {
    var apiForChild: FinanceAPI { api }
}

struct TransactionEditorView: View {
    let api: FinanceAPI
    let wallets: [WalletModel]
    let categories: [CategoryModel]
    let timezone: String
    let transaction: TransactionModel?
    let onSaved: () -> Void
    let onCancel: () -> Void
    @State private var type = "expense"
    @State private var amount = ""
    @State private var note = ""
    @State private var walletID = ""
    @State private var categoryID = ""
    @State private var jarID = ""
    @State private var includedInReports = true
    @State private var date = Date()
    @State private var jarSummary: JarMonthSummaryModel?
    @State private var jarsLoading = false
    @State private var jarLoadError: String?
    @State private var isSaving = false
    @State private var errorMessage: String?

    init(api: FinanceAPI, wallets: [WalletModel], categories: [CategoryModel], timezone: String = TimeZone.current.identifier, transaction: TransactionModel? = nil, onSaved: @escaping () -> Void, onCancel: @escaping () -> Void) {
        self.api = api
        self.wallets = wallets
        self.categories = categories
        self.timezone = timezone
        self.transaction = transaction
        self.onSaved = onSaved
        self.onCancel = onCancel
        _type = State(initialValue: transaction?.type ?? "expense")
        _amount = State(initialValue: transaction.map { String($0.amount) } ?? "")
        _note = State(initialValue: transaction?.note ?? "")
        _walletID = State(initialValue: transaction?.walletID ?? TransactionsViewModel.ordinaryTransactionWallets(wallets).first?.id ?? "")
        _categoryID = State(initialValue: transaction?.categoryID ?? "")
        _jarID = State(initialValue: transaction?.jarID ?? "")
        _includedInReports = State(initialValue: transaction?.includedInReports ?? true)
        if let raw = transaction?.occurredAt, let date = MPFormat.parseInstant(raw) { _date = State(initialValue: date) }
    }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                MPScreenHeader(title: transaction == nil ? "Thêm giao dịch" : "Sửa giao dịch", leading: { MPIconButton(systemName: "xmark", label: "Đóng", action: onCancel) }, trailing: { EmptyView() })
                MPSegmentedControl(options: [("expense", "Chi tiêu"), ("income", "Thu nhập")], selection: $type)
                MPOptionSelector(title: "Ví", options: walletOptions, selectionID: $walletID)
                MPOptionSelector(title: "Danh mục", options: categoryOptions, selectionID: $categoryID, emptyLabel: "Không chọn")
                if canUseJar {
                    MPOptionSelector(
                        title: "Hũ chi tiêu",
                        options: TransactionEditorPolicy.jarOptions(
                            summary: jarSummary,
                            selectedJarID: jarID,
                            selectedJarName: transaction?.jarName
                        ),
                        selectionID: $jarID,
                        emptyLabel: jarsLoading ? "Đang tải hũ..." : "Không chọn"
                    )
                    if let jarLoadError { MPStatusMessage(text: jarLoadError, tone: .danger) }
                }
                MPAmountField(title: "Số tiền", text: $amount)
                MPDateTimeField(title: "Ngày và giờ", date: $date)
                MPTextField(title: "Ghi chú", text: $note, placeholder: "Ví dụ: Ăn trưa")
                MPBooleanSelector(title: "Tính vào báo cáo", value: $includedInReports, trueLabel: "Có", falseLabel: "Không")
                if let errorMessage { MPStatusMessage(text: errorMessage, tone: .danger) }
                MPButton(transaction == nil ? "Lưu giao dịch" : "Lưu thay đổi", isLoading: isSaving) { Task { await save() } }
            }
            .padding(MPMetrics.pagePadding)
        }
        .environment(\.timeZone, TimeZone(identifier: timezone) ?? .current)
        .task {
            if walletID.isEmpty { walletID = TransactionsViewModel.ordinaryTransactionWallets(wallets).first?.id ?? "" }
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
    }

    private func save() async {
        guard let wallet = wallets.first(where: { $0.id == walletID }), let parsed = MPFormat.parsePositiveInteger(amount) else {
            errorMessage = "Chọn ví và nhập số tiền hợp lệ."
            return
        }
        let selectedCategory = categories.first(where: { $0.id == categoryID })
        if let selectedCategory, !selectedCategory.applies(to: type, walletID: wallet.id, walletType: wallet.type) {
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
            let input = TransactionInputModel(walletID: wallet.id, categoryID: selectedCategory?.id, jarID: jarID.isEmpty ? nil : jarID, type: type, amount: parsed, occurredAt: Self.timestamp(from: date, timezone: timezone), note: note.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty ? nil : note.trimmingCharacters(in: .whitespacesAndNewlines), includedInReports: includedInReports, travelEventID: transaction?.travelEventID)
            if let transaction { _ = try await api.updateTransaction(id: transaction.id, input: input) }
            else { _ = try await api.createTransaction(input) }
            onSaved()
        } catch {
            errorMessage = "Không thể lưu giao dịch."
        }
    }

    private var selectedCategory: CategoryModel? { categories.first(where: { $0.id == categoryID }) }
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
    private var categoryOptions: [MPSelectionOption] {
        let valid = categories.filter { $0.applies(to: type, walletID: selectedWallet?.id ?? "", walletType: selectedWallet?.type ?? "basic") }
        var options = [MPSelectionOption(id: "", label: "Không chọn")]
        options += valid.map { MPSelectionOption(id: $0.id, label: $0.name) }
        if !categoryID.isEmpty, !options.contains(where: { $0.id == categoryID }), let current = categories.first(where: { $0.id == categoryID }) {
            options.append(MPSelectionOption(id: current.id, label: current.name))
        }
        return options
    }

    private func clearInvalidCategory() {
        guard let selectedCategory = categories.first(where: { $0.id == categoryID }), let selectedWallet else { return }
        if !selectedCategory.applies(to: type, walletID: selectedWallet.id, walletType: selectedWallet.type) {
            categoryID = ""
        }
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

    var body: some View {
        VStack(alignment: .leading, spacing: 16) {
            MPScreenHeader(title: "Chi tiết giao dịch", leading: { MPIconButton(systemName: "xmark", label: "Đóng", action: onClose) }, trailing: { EmptyView() })
            MPMetricCard(title: transaction.type == "income" ? "Thu nhập" : "Chi tiêu", value: MPFormat.money(transaction.amount))
            MPSurfaceCard {
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
    let api: FinanceAPI
    let wallets: [WalletModel]
    let timezone: String
    let transactions: [TransactionModel]
    let existingTransferID: String?
    let onSaved: () -> Void
    let onCancel: () -> Void
    @State private var sourceID = ""
    @State private var destinationID = ""
    @State private var amount = ""
    @State private var note = ""
    @State private var date = Date()
    @State private var saving = false
    @State private var error: String?

    init(api: FinanceAPI, wallets: [WalletModel], timezone: String = TimeZone.current.identifier, transactions: [TransactionModel] = [], existingTransferID: String? = nil, onSaved: @escaping () -> Void, onCancel: @escaping () -> Void) {
        self.api = api
        self.wallets = wallets
        self.timezone = timezone
        self.transactions = transactions
        self.existingTransferID = existingTransferID
        self.onSaved = onSaved
        self.onCancel = onCancel
        let source = transactions.first(where: { $0.transferID == existingTransferID && $0.type == "expense" })
        let destination = transactions.first(where: { $0.transferID == existingTransferID && $0.type == "income" })
        let eligibleWallets = Self.transferWallets(wallets)
        _sourceID = State(initialValue: source?.walletID ?? eligibleWallets.first?.id ?? "")
        _destinationID = State(initialValue: destination?.walletID ?? eligibleWallets.dropFirst().first?.id ?? eligibleWallets.first?.id ?? "")
        _amount = State(initialValue: source.map { String($0.amount) } ?? "")
        _note = State(initialValue: source?.note ?? "")
        _date = State(initialValue: source.flatMap { MPFormat.parseInstant($0.occurredAt) } ?? Date())
    }

    var body: some View {
        ScrollView { VStack(alignment: .leading, spacing: 16) {
            MPScreenHeader(title: existingTransferID == nil ? "Chuyển tiền" : "Sửa chuyển tiền", leading: { MPIconButton(systemName: "xmark", label: "Đóng", action: onCancel) }, trailing: { EmptyView() })
            MPText("Chuyển tiền tạo cặp giao dịch nguyên tử giữa hai ví; không tính vào báo cáo.", style: .secondary)
            MPOptionSelector(title: "Từ ví", options: walletOptions, selectionID: $sourceID)
            MPOptionSelector(title: "Đến ví", options: walletOptions, selectionID: $destinationID)
            MPAmountField(title: "Số tiền", text: $amount)
            MPDateField(title: "Ngày chuyển", date: $date)
            MPTextField(title: "Ghi chú", text: $note)
            if let error { MPStatusMessage(text: error, tone: .danger) }
            MPButton(existingTransferID == nil ? "Xác nhận chuyển" : "Lưu thay đổi", isLoading: saving) { Task { await save() } }
        }.padding(MPMetrics.pagePadding) }
        .environment(\.timeZone, TimeZone(identifier: timezone) ?? .current)
        .task {
            let eligibleWallets = Self.transferWallets(wallets)
            if sourceID.isEmpty { sourceID = eligibleWallets.first?.id ?? "" }
            if destinationID.isEmpty { destinationID = eligibleWallets.dropFirst().first?.id ?? eligibleWallets.first?.id ?? "" }
        }
    }

    private func save() async {
        guard let from = wallets.first(where: { $0.id == sourceID }), let to = wallets.first(where: { $0.id == destinationID }), from.id != to.id, let parsed = MPFormat.parsePositiveInteger(amount) else { error = "Chọn hai ví khác nhau và nhập số tiền hợp lệ."; return }
        saving = true; defer { saving = false }
        do {
            let occurredAt = Self.timestamp(date, timezone: timezone)
            let trimmedNote = note.trimmingCharacters(in: .whitespacesAndNewlines)
            if let existingTransferID { _ = try await api.updateTransfer(id: existingTransferID, input: TransferUpdateInputModel(amount: parsed, occurredAt: occurredAt, note: trimmedNote.isEmpty ? nil : trimmedNote)) }
            else { _ = try await api.createTransfer(TransferInputModel(sourceWalletID: from.id, destinationWalletID: to.id, amount: parsed, occurredAt: occurredAt, note: trimmedNote.isEmpty ? nil : trimmedNote)) }
            onSaved()
        }
        catch { self.error = "Không thể chuyển tiền. Số dư hoặc ví có thể không hợp lệ." }
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
