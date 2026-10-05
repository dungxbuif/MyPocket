import SwiftUI

@MainActor
final class TransactionsViewModel: ObservableObject {
    @Published private(set) var transactions: [TransactionModel] = []
    @Published private(set) var wallets: [WalletModel] = []
    @Published private(set) var categories: [CategoryModel] = []
    @Published private(set) var isLoading = false
    @Published private(set) var errorMessage: String?
    @Published var walletSelection = "__all__"
    @Published var periodSelection = "all"
    @Published var customStart = Calendar.current.date(byAdding: .day, value: -30, to: Date()) ?? Date()
    @Published var customEnd = Date()
    @Published private(set) var accountTimezone = TimeZone.current.identifier

    private let api: FinanceAPI
    init(api: FinanceAPI) { self.api = api }

    var filteredTransactions: [TransactionModel] {
        let walletFiltered = walletSelection == "__all__" ? transactions : transactions.filter { $0.walletID == walletSelection }
        if periodSelection == "all" { return walletFiltered.sorted { $0.occurredAt > $1.occurredAt } }
        var calendar = Calendar(identifier: .gregorian)
        calendar.timeZone = TimeZone(identifier: accountTimezone) ?? .current
        let start: Date
        let end: Date
        if periodSelection == "week" {
            start = calendar.dateInterval(of: .weekOfYear, for: Date())?.start ?? Date()
            end = calendar.date(byAdding: .weekOfYear, value: 1, to: start) ?? Date()
        } else {
            start = calendar.startOfDay(for: customStart)
            end = calendar.date(byAdding: .day, value: 1, to: calendar.startOfDay(for: customEnd)) ?? Date()
        }
        return walletFiltered.filter {
            guard let date = MPFormat.parseInstant($0.occurredAt) else { return true }
            return date >= start && date < end
        }.sorted { $0.occurredAt > $1.occurredAt }
    }

    var totalExpense: Int64 { filteredTransactions.filter { $0.type == "expense" }.reduce(0) { $0 + $1.amount } }
    var totalIncome: Int64 { filteredTransactions.filter { $0.type == "income" }.reduce(0) { $0 + $1.amount } }
    var groupedTransactions: [(String, [TransactionModel])] {
        var calendar = Calendar(identifier: .gregorian)
        calendar.timeZone = TimeZone(identifier: accountTimezone) ?? .current
        let formatter = DateFormatter()
        formatter.calendar = calendar
        formatter.timeZone = calendar.timeZone
        formatter.locale = Locale(identifier: "en_US_POSIX")
        formatter.dateFormat = "yyyy-MM-dd"
        return Dictionary(grouping: filteredTransactions) {
            guard let date = MPFormat.parseInstant($0.occurredAt) else { return String($0.occurredAt.prefix(10)) }
            return formatter.string(from: date)
        }
            .sorted { $0.key > $1.key }
    }

    func load() async {
        guard !isLoading else { return }
        isLoading = true
        errorMessage = nil
        defer { isLoading = false }
        do {
            async let transactions = api.transactions()
            async let wallets = api.wallets()
            async let categories = api.categories()
            async let profile = api.profile()
            self.transactions = try await transactions
            self.wallets = try await wallets
            self.categories = try await categories
            if let profile = try? await profile { accountTimezone = profile.timezone }
        } catch {
            errorMessage = "Không thể tải giao dịch."
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

    init(api: FinanceAPI) { _model = StateObject(wrappedValue: TransactionsViewModel(api: api)) }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                MPScreenHeader(title: "Giao dịch", leading: { EmptyView() }, trailing: {
                    HStack(spacing: 4) {
                        MPIconButton(systemName: "arrow.left.arrow.right", label: "Chuyển tiền") { transferPresented = true }
                        MPIconButton(systemName: "slider.horizontal.3", label: "Điều chỉnh số dư") { adjustmentPresented = true }
                        MPIconButton(systemName: "plus", label: "Thêm giao dịch") { editTarget = nil; editorPresented = true }
                    }
                })
                MPSegmentedControl(options: [("all", "Tất cả"), ("week", "Tuần này"), ("custom", "Tùy chọn")], selection: $model.periodSelection)
                if model.periodSelection == "custom" {
                    MPDateRangePicker(start: $model.customStart, end: $model.customEnd)
                }
                MPOptionSelector(title: "Ví", options: [MPSelectionOption(id: "__all__", label: "Tổng cộng")] + model.wallets.map { MPSelectionOption(id: $0.id, label: $0.name) }, selectionID: $model.walletSelection, emptyLabel: "Tổng cộng")
                if model.isLoading && model.transactions.isEmpty {
                    MPLoadingView(label: "Đang tải giao dịch...")
                } else if let errorMessage = model.errorMessage, model.transactions.isEmpty {
                    MPErrorView(title: "Không tải được giao dịch", detail: errorMessage) { Task { await model.load() } }
                } else {
                    LazyVGrid(columns: [GridItem(.flexible()), GridItem(.flexible())], spacing: 12) {
                        MPMetricCard(title: "Chi tiêu", value: MPFormat.money(model.totalExpense))
                        MPMetricCard(title: "Thu nhập", value: MPFormat.money(model.totalIncome))
                    }
                    if model.filteredTransactions.isEmpty {
                        MPEmptyView(title: "Không có giao dịch", detail: "Thử đổi khoảng thời gian hoặc thêm giao dịch mới.")
                    } else {
                        ForEach(model.groupedTransactions, id: \.0) { day, transactions in
                            MPSurfaceCard {
                                VStack(alignment: .leading, spacing: 0) {
                                    MPText(day, style: .secondary)
                                        .padding(.bottom, 4)
                                    ForEach(transactions) { transaction in
                                        MPButton(style: .plain) {
                                            selectedTransaction = transaction
                                        } label: {
                                            MPTransactionRow(
                                                title: transaction.note?.isEmpty == false ? transaction.note! : (transaction.type == "income" ? "Thu nhập" : "Chi tiêu"),
                                                subtitle: MPFormat.shortDate(transaction.occurredAt, timeZone: TimeZone(identifier: model.accountTimezone) ?? .current),
                                                amount: MPFormat.money(transaction.amount),
                                                isIncome: transaction.type == "income"
                                            )
                                        }
                                        .contextMenu {
                                            MPButton("Xóa giao dịch", style: .destructive) { deleteTarget = transaction }
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
                TransferEditorView(api: modelAPI, wallets: model.wallets) {
                    transferPresented = false
                    Task { await model.load() }
                } onCancel: { transferPresented = false }
                .background(MPColor.canvas.ignoresSafeArea())
            } else if adjustmentPresented {
                AdjustmentEditorView(api: modelAPI, wallets: model.wallets) {
                    adjustmentPresented = false
                    Task { await model.load() }
                } onCancel: { adjustmentPresented = false }
                .background(MPColor.canvas.ignoresSafeArea())
            } else if let transferEditID {
                TransferEditorView(api: modelAPI, wallets: model.wallets, transactions: model.transactions, existingTransferID: transferEditID) {
                    self.transferEditID = nil
                    Task { await model.load() }
                } onCancel: { self.transferEditID = nil }
                .background(MPColor.canvas.ignoresSafeArea())
            } else if let selectedTransaction {
                TransactionDetailView(api: modelAPI, transaction: selectedTransaction, allTransactions: model.transactions, onEdit: {
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
        .accessibilityIdentifier("transactions-screen")
    }

    private var modelAPI: FinanceAPI { model.apiForChild }
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
    @State private var date = Date()
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
        _walletID = State(initialValue: transaction?.walletID ?? wallets.first?.id ?? "")
        _categoryID = State(initialValue: transaction?.categoryID ?? "")
        if let raw = transaction?.occurredAt, let date = MPFormat.parseInstant(raw) { _date = State(initialValue: date) }
    }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                MPScreenHeader(title: transaction == nil ? "Thêm giao dịch" : "Sửa giao dịch", leading: { MPIconButton(systemName: "xmark", label: "Đóng", action: onCancel) }, trailing: { EmptyView() })
                MPSegmentedControl(options: [("expense", "Chi tiêu"), ("income", "Thu nhập")], selection: $type)
                MPOptionSelector(title: "Ví", options: walletOptions, selectionID: $walletID)
                MPOptionSelector(title: "Danh mục", options: categoryOptions, selectionID: $categoryID, emptyLabel: "Không chọn")
                MPAmountField(title: "Số tiền", text: $amount)
                MPDateField(title: "Ngày giao dịch", date: $date)
                MPTextField(title: "Ghi chú", text: $note, placeholder: "Ví dụ: Ăn trưa")
                if let errorMessage { MPStatusMessage(text: errorMessage, tone: .danger) }
                MPButton(transaction == nil ? "Lưu giao dịch" : "Lưu thay đổi", isLoading: isSaving) { Task { await save() } }
            }
            .padding(MPMetrics.pagePadding)
        }
        .environment(\.timeZone, TimeZone(identifier: timezone) ?? .current)
        .task { if walletID.isEmpty { walletID = wallets.first?.id ?? "" } }
    }

    private func save() async {
        guard let wallet = wallets.first(where: { $0.id == walletID }), let parsed = Int64(amount.filter(\.isNumber)), parsed > 0 else {
            errorMessage = "Chọn ví và nhập số tiền hợp lệ."
            return
        }
        let selectedCategory = categories.first(where: { $0.id == categoryID })
        if let selectedCategory, !selectedCategory.applies(to: type, walletID: wallet.id, walletType: wallet.type) {
            errorMessage = "Danh mục không phù hợp với loại giao dịch hoặc ví đã chọn."
            return
        }
        isSaving = true
        defer { isSaving = false }
        do {
            let input = TransactionInputModel(walletID: wallet.id, categoryID: selectedCategory?.id, type: type, amount: parsed, occurredAt: Self.dayString(from: date, timezone: timezone), note: note.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty ? nil : note.trimmingCharacters(in: .whitespacesAndNewlines), includedInReports: true, travelEventID: transaction?.travelEventID)
            if let transaction { _ = try await api.updateTransaction(id: transaction.id, input: input) }
            else { _ = try await api.createTransaction(input) }
            onSaved()
        } catch {
            errorMessage = "Không thể lưu giao dịch."
        }
    }

    private static func dayString(from date: Date, timezone: String) -> String {
        let formatter = DateFormatter()
        formatter.dateFormat = "yyyy-MM-dd'T'HH:mm:ssXXXXX"
        formatter.locale = Locale(identifier: "en_US_POSIX")
        formatter.timeZone = TimeZone(identifier: timezone) ?? .current
        return formatter.string(from: date)
    }

    private var selectedWallet: WalletModel? { wallets.first(where: { $0.id == walletID }) }
    private var walletOptions: [MPSelectionOption] { wallets.map { MPSelectionOption(id: $0.id, label: $0.name) } }
    private var categoryOptions: [MPSelectionOption] {
        let valid = categories.filter { $0.applies(to: type, walletID: selectedWallet?.id ?? "", walletType: selectedWallet?.type ?? "basic") }
        var options = [MPSelectionOption(id: "", label: "Không chọn")]
        options += valid.map { MPSelectionOption(id: $0.id, label: $0.name) }
        if !categoryID.isEmpty, !options.contains(where: { $0.id == categoryID }), let current = categories.first(where: { $0.id == categoryID }) {
            options.append(MPSelectionOption(id: current.id, label: current.name))
        }
        return options
    }
}

struct TransactionDetailView: View {
    let api: FinanceAPI
    let transaction: TransactionModel
    let allTransactions: [TransactionModel]
    let onEdit: () -> Void
    let onEditTransfer: () -> Void
    let onDelete: () -> Void
    let onDeleteTransfer: () -> Void
    let onClose: () -> Void
    @State private var travelEvents: [TravelEventModel] = []
    @State private var travelEventName = "Không gắn chuyến"
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
                    if transaction.transferID == nil {
                        MPWalletSelector(title: "Travel Mode", options: ["Không gắn chuyến"] + travelEvents.map(\.name), selection: $travelEventName)
                        MPButton("Lưu liên kết chuyến", style: .secondary) {
                            Task {
                                let event = travelEvents.first(where: { $0.name == travelEventName })
                                do { _ = try await api.updateTransactionTravel(id: transaction.id, eventID: event?.id); travelError = nil }
                                catch { travelError = "Không thể cập nhật liên kết chuyến." }
                            }
                        }
                        if let travelError { MPStatusMessage(text: travelError, tone: .danger) }
                    }
                    MPText(transaction.includedInReports ? "Được tính vào báo cáo" : "Không tính vào báo cáo", style: .secondary)
                }
            }
            if transaction.transferID == nil {
                MPButton("Sửa giao dịch", style: .secondary, action: onEdit)
                MPButton("Xóa giao dịch", style: .destructive, action: onDelete)
            } else {
                MPStatusMessage(text: "Đây là cặp chuyển ví; sửa/xóa phải áp dụng cho cả hai giao dịch.", tone: .info)
                MPButton("Sửa cả cặp chuyển ví", style: .secondary, action: onEditTransfer)
                MPButton("Xóa cả cặp chuyển ví", style: .destructive, action: onDeleteTransfer)
            }
            Spacer()
        }
        .padding(MPMetrics.pagePadding)
        .task {
            guard transaction.transferID == nil else { return }
            do {
                travelEvents = try await api.travelEvents()
                travelEventName = travelEvents.first(where: { $0.id == transaction.travelEventID })?.name ?? "Không gắn chuyến"
            } catch { travelError = "Không thể tải Travel Mode." }
        }
    }
}

struct TransferEditorView: View {
    let api: FinanceAPI
    let wallets: [WalletModel]
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

    init(api: FinanceAPI, wallets: [WalletModel], transactions: [TransactionModel] = [], existingTransferID: String? = nil, onSaved: @escaping () -> Void, onCancel: @escaping () -> Void) {
        self.api = api
        self.wallets = wallets
        self.transactions = transactions
        self.existingTransferID = existingTransferID
        self.onSaved = onSaved
        self.onCancel = onCancel
        let source = transactions.first(where: { $0.transferID == existingTransferID && $0.type == "expense" })
        let destination = transactions.first(where: { $0.transferID == existingTransferID && $0.type == "income" })
        _sourceID = State(initialValue: source?.walletID ?? wallets.first?.id ?? "")
        _destinationID = State(initialValue: destination?.walletID ?? wallets.dropFirst().first?.id ?? wallets.first?.id ?? "")
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
        .task { if sourceID.isEmpty { sourceID = wallets.first?.id ?? "" }; if destinationID.isEmpty { destinationID = wallets.dropFirst().first?.id ?? wallets.first?.id ?? "" } }
    }

    private func save() async {
        guard let from = wallets.first(where: { $0.id == sourceID }), let to = wallets.first(where: { $0.id == destinationID }), from.id != to.id, let parsed = Int64(amount.filter(\.isNumber)), parsed > 0 else { error = "Chọn hai ví khác nhau và nhập số tiền hợp lệ."; return }
        saving = true; defer { saving = false }
        do {
            if let existingTransferID { _ = try await api.updateTransfer(id: existingTransferID, input: TransferUpdateInputModel(amount: parsed, occurredAt: Self.timestamp(date), note: note.isEmpty ? nil : note)) }
            else { _ = try await api.createTransfer(TransferInputModel(sourceWalletID: from.id, destinationWalletID: to.id, amount: parsed, occurredAt: Self.timestamp(date), note: note.isEmpty ? nil : note)) }
            onSaved()
        }
        catch { self.error = "Không thể chuyển tiền. Số dư hoặc ví có thể không hợp lệ." }
    }
    private var walletOptions: [MPSelectionOption] { wallets.map { MPSelectionOption(id: $0.id, label: $0.name) } }
    private static func timestamp(_ date: Date) -> String { let formatter = ISO8601DateFormatter(); return formatter.string(from: date) }
}

struct AdjustmentEditorView: View {
    let api: FinanceAPI
    let wallets: [WalletModel]
    let onSaved: () -> Void
    let onCancel: () -> Void
    @State private var walletID = ""
    @State private var direction = "increase"
    @State private var amount = ""
    @State private var note = ""
    @State private var saving = false
    @State private var error: String?

    var body: some View {
        ScrollView { VStack(alignment: .leading, spacing: 16) {
            MPScreenHeader(title: "Điều chỉnh số dư", leading: { MPIconButton(systemName: "xmark", label: "Đóng", action: onCancel) }, trailing: { EmptyView() })
            MPOptionSelector(title: "Ví", options: wallets.map { MPSelectionOption(id: $0.id, label: $0.name) }, selectionID: $walletID)
            MPWalletSelector(title: "Hướng", options: ["increase", "decrease"], selection: $direction)
            MPAmountField(title: "Số tiền", text: $amount)
            MPTextField(title: "Lý do", text: $note)
            if let error { MPStatusMessage(text: error, tone: .danger) }
            MPButton("Lưu điều chỉnh", isLoading: saving) { Task { await save() } }
        }.padding(MPMetrics.pagePadding) }
        .task { if walletID.isEmpty { walletID = wallets.first?.id ?? "" } }
    }
    private func save() async { guard let selected = wallets.first(where: { $0.id == walletID }), let parsed = Int64(amount.filter(\.isNumber)), parsed > 0 else { error = "Chọn ví và nhập số tiền hợp lệ."; return }; saving = true; defer { saving = false }; do { _ = try await api.createAdjustment(AdjustmentInputModel(walletID: selected.id, amount: parsed, direction: direction, occurredAt: ISO8601DateFormatter().string(from: Date()), note: note.isEmpty ? nil : note)); onSaved() } catch { self.error = "Không thể điều chỉnh số dư." } }
}
