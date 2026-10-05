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

    private let api: FinanceAPI
    init(api: FinanceAPI) { self.api = api }

    var filteredTransactions: [TransactionModel] {
        let walletFiltered = walletSelection == "__all__" ? transactions : transactions.filter { $0.walletID == walletSelection }
        guard periodSelection != "all" else { return walletFiltered.sorted { $0.occurredAt > $1.occurredAt } }
        let calendar = Calendar.current
        let start = calendar.dateInterval(of: .weekOfYear, for: Date())?.start ?? Date()
        let formatter = ISO8601DateFormatter()
        return walletFiltered.filter {
            guard let date = formatter.date(from: $0.occurredAt) else { return true }
            return date >= start
        }.sorted { $0.occurredAt > $1.occurredAt }
    }

    var totalExpense: Int64 { filteredTransactions.filter { $0.type == "expense" }.reduce(0) { $0 + $1.amount } }
    var totalIncome: Int64 { filteredTransactions.filter { $0.type == "income" }.reduce(0) { $0 + $1.amount } }

    func load() async {
        guard !isLoading else { return }
        isLoading = true
        errorMessage = nil
        defer { isLoading = false }
        do {
            async let transactions = api.transactions()
            async let wallets = api.wallets()
            async let categories = api.categories()
            self.transactions = try await transactions
            self.wallets = try await wallets
            self.categories = try await categories
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
}

struct TransactionsView: View {
    @StateObject private var model: TransactionsViewModel
    @State private var editorPresented = false
    @State private var transferPresented = false
    @State private var adjustmentPresented = false
    @State private var selectedTransaction: TransactionModel?
    @State private var deleteTarget: TransactionModel?

    init(api: FinanceAPI) { _model = StateObject(wrappedValue: TransactionsViewModel(api: api)) }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                MPScreenHeader(title: "Giao dịch", leading: { EmptyView() }, trailing: {
                    HStack(spacing: 4) {
                        MPIconButton(systemName: "arrow.left.arrow.right", label: "Chuyển tiền") { transferPresented = true }
                        MPIconButton(systemName: "slider.horizontal.3", label: "Điều chỉnh số dư") { adjustmentPresented = true }
                        MPIconButton(systemName: "plus", label: "Thêm giao dịch") { editorPresented = true }
                    }
                })
                MPSegmentedControl(options: [("all", "Tất cả"), ("week", "Tuần này")], selection: $model.periodSelection)
                MPWalletSelector(title: "Ví", options: ["Tổng cộng"] + model.wallets.map(\.name), selection: Binding(
                    get: { model.walletSelection == "__all__" ? "Tổng cộng" : model.wallets.first(where: { $0.id == model.walletSelection })?.name ?? "Tổng cộng" },
                    set: { name in model.walletSelection = model.wallets.first(where: { $0.name == name })?.id ?? "__all__" }
                ))
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
                        MPSurfaceCard {
                            VStack(spacing: 0) {
                                ForEach(model.filteredTransactions) { transaction in
                                    MPButton(style: .plain) {
                                        selectedTransaction = transaction
                                    } label: {
                                        MPTransactionRow(
                                            title: transaction.note?.isEmpty == false ? transaction.note! : (transaction.type == "income" ? "Thu nhập" : "Chi tiêu"),
                                            subtitle: MPFormat.shortDate(transaction.occurredAt),
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
            .padding(MPMetrics.pagePadding)
        }
        .background(MPColor.canvas.ignoresSafeArea())
        .task { await model.load() }
        .refreshable { await model.load() }
        .overlay {
            if editorPresented {
                TransactionEditorView(api: modelAPI, wallets: model.wallets, categories: model.categories) {
                    editorPresented = false
                    Task { await model.load() }
                } onCancel: { editorPresented = false }
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
            } else if let selectedTransaction {
                TransactionDetailView(transaction: selectedTransaction, onDelete: {
                    deleteTarget = selectedTransaction
                    self.selectedTransaction = nil
                }, onClose: { self.selectedTransaction = nil })
                .background(MPColor.canvas.ignoresSafeArea())
            }
        }
        .mpConfirmDialog(isPresented: Binding(get: { deleteTarget != nil }, set: { if !$0 { deleteTarget = nil } }), title: "Xóa giao dịch?", message: "Hành động này không thể hoàn tác.", confirmTitle: "Xóa") {
            if let target = deleteTarget { Task { await model.delete(target) } }
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
    let onSaved: () -> Void
    let onCancel: () -> Void
    @State private var type = "expense"
    @State private var amount = ""
    @State private var note = ""
    @State private var walletName = ""
    @State private var categoryName = "Không chọn"
    @State private var date = Date()
    @State private var isSaving = false
    @State private var errorMessage: String?

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                MPScreenHeader(title: "Thêm giao dịch", leading: { MPIconButton(systemName: "xmark", label: "Đóng", action: onCancel) }, trailing: { EmptyView() })
                MPSegmentedControl(options: [("expense", "Chi tiêu"), ("income", "Thu nhập")], selection: $type)
                MPWalletSelector(title: "Ví", options: wallets.map(\.name), selection: $walletName)
                MPWalletSelector(title: "Danh mục", options: ["Không chọn"] + categories.map(\.name), selection: $categoryName)
                MPAmountField(title: "Số tiền", text: $amount)
                MPDateField(title: "Ngày giao dịch", date: $date)
                MPTextField(title: "Ghi chú", text: $note, placeholder: "Ví dụ: Ăn trưa")
                if let errorMessage { MPStatusMessage(text: errorMessage, tone: .danger) }
                MPButton("Lưu giao dịch", isLoading: isSaving) { Task { await save() } }
            }
            .padding(MPMetrics.pagePadding)
        }
        .task { if walletName.isEmpty { walletName = wallets.first?.name ?? "" } }
    }

    private func save() async {
        guard let wallet = wallets.first(where: { $0.name == walletName }), let parsed = Int64(amount.filter(\.isNumber)), parsed > 0 else {
            errorMessage = "Chọn ví và nhập số tiền hợp lệ."
            return
        }
        isSaving = true
        defer { isSaving = false }
        do {
            let input = TransactionInputModel(walletID: wallet.id, categoryID: categories.first(where: { $0.name == categoryName })?.id, type: type, amount: parsed, occurredAt: Self.dayString(from: date), note: note.isEmpty ? nil : note, includedInReports: true, travelEventID: nil)
            _ = try await api.createTransaction(input)
            onSaved()
        } catch {
            errorMessage = "Không thể lưu giao dịch."
        }
    }

    private static func dayString(from date: Date) -> String {
        let formatter = DateFormatter()
        formatter.dateFormat = "yyyy-MM-dd'T'HH:mm:ssXXXXX"
        formatter.locale = Locale(identifier: "en_US_POSIX")
        return formatter.string(from: date)
    }
}

struct TransactionDetailView: View {
    let transaction: TransactionModel
    let onDelete: () -> Void
    let onClose: () -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 16) {
            MPScreenHeader(title: "Chi tiết giao dịch", leading: { MPIconButton(systemName: "xmark", label: "Đóng", action: onClose) }, trailing: { EmptyView() })
            MPMetricCard(title: transaction.type == "income" ? "Thu nhập" : "Chi tiêu", value: MPFormat.money(transaction.amount))
            MPSurfaceCard {
                VStack(alignment: .leading, spacing: 8) {
                    MPText("Ngày: \(MPFormat.shortDate(transaction.occurredAt))")
                    if let note = transaction.note, !note.isEmpty { MPText("Ghi chú: \(note)") }
                    MPText(transaction.includedInReports ? "Được tính vào báo cáo" : "Không tính vào báo cáo", style: .secondary)
                }
            }
            MPButton("Xóa giao dịch", style: .destructive, action: onDelete)
            Spacer()
        }
        .padding(MPMetrics.pagePadding)
    }
}

struct TransferEditorView: View {
    let api: FinanceAPI
    let wallets: [WalletModel]
    let onSaved: () -> Void
    let onCancel: () -> Void
    @State private var source = ""
    @State private var destination = ""
    @State private var amount = ""
    @State private var note = ""
    @State private var date = Date()
    @State private var saving = false
    @State private var error: String?

    var body: some View {
        ScrollView { VStack(alignment: .leading, spacing: 16) {
            MPScreenHeader(title: "Chuyển tiền", leading: { MPIconButton(systemName: "xmark", label: "Đóng", action: onCancel) }, trailing: { EmptyView() })
            MPText("Chuyển tiền tạo cặp giao dịch nguyên tử giữa hai ví; không tính vào báo cáo.", style: .secondary)
            MPWalletSelector(title: "Từ ví", options: wallets.map(\.name), selection: $source)
            MPWalletSelector(title: "Đến ví", options: wallets.map(\.name), selection: $destination)
            MPAmountField(title: "Số tiền", text: $amount)
            MPDateField(title: "Ngày chuyển", date: $date)
            MPTextField(title: "Ghi chú", text: $note)
            if let error { MPStatusMessage(text: error, tone: .danger) }
            MPButton("Xác nhận chuyển", isLoading: saving) { Task { await save() } }
        }.padding(MPMetrics.pagePadding) }
        .task { if source.isEmpty { source = wallets.first?.name ?? "" }; if destination.isEmpty { destination = wallets.dropFirst().first?.name ?? wallets.first?.name ?? "" } }
    }

    private func save() async {
        guard let from = wallets.first(where: { $0.name == source }), let to = wallets.first(where: { $0.name == destination }), from.id != to.id, let parsed = Int64(amount.filter(\.isNumber)), parsed > 0 else { error = "Chọn hai ví khác nhau và nhập số tiền hợp lệ."; return }
        saving = true; defer { saving = false }
        do { _ = try await api.createTransfer(TransferInputModel(sourceWalletID: from.id, destinationWalletID: to.id, amount: parsed, occurredAt: Self.timestamp(date), note: note.isEmpty ? nil : note)); onSaved() }
        catch { self.error = "Không thể chuyển tiền. Số dư hoặc ví có thể không hợp lệ." }
    }
    private static func timestamp(_ date: Date) -> String { let formatter = ISO8601DateFormatter(); return formatter.string(from: date) }
}

struct AdjustmentEditorView: View {
    let api: FinanceAPI
    let wallets: [WalletModel]
    let onSaved: () -> Void
    let onCancel: () -> Void
    @State private var wallet = ""
    @State private var direction = "increase"
    @State private var amount = ""
    @State private var note = ""
    @State private var saving = false
    @State private var error: String?

    var body: some View {
        ScrollView { VStack(alignment: .leading, spacing: 16) {
            MPScreenHeader(title: "Điều chỉnh số dư", leading: { MPIconButton(systemName: "xmark", label: "Đóng", action: onCancel) }, trailing: { EmptyView() })
            MPWalletSelector(title: "Ví", options: wallets.map(\.name), selection: $wallet)
            MPWalletSelector(title: "Hướng", options: ["increase", "decrease"], selection: $direction)
            MPAmountField(title: "Số tiền", text: $amount)
            MPTextField(title: "Lý do", text: $note)
            if let error { MPStatusMessage(text: error, tone: .danger) }
            MPButton("Lưu điều chỉnh", isLoading: saving) { Task { await save() } }
        }.padding(MPMetrics.pagePadding) }
        .task { if wallet.isEmpty { wallet = wallets.first?.name ?? "" } }
    }
    private func save() async { guard let selected = wallets.first(where: { $0.name == wallet }), let parsed = Int64(amount.filter(\.isNumber)), parsed > 0 else { error = "Chọn ví và nhập số tiền hợp lệ."; return }; saving = true; defer { saving = false }; do { _ = try await api.createAdjustment(AdjustmentInputModel(walletID: selected.id, amount: parsed, direction: direction, occurredAt: ISO8601DateFormatter().string(from: Date()), note: note.isEmpty ? nil : note)); onSaved() } catch { self.error = "Không thể điều chỉnh số dư." } }
}
