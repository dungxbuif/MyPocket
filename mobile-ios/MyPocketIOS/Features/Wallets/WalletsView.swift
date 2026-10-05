import SwiftUI

@MainActor
final class WalletsViewModel: ObservableObject {
    @Published private(set) var wallets: [WalletModel] = []
    @Published private(set) var isLoading = false
    @Published private(set) var errorMessage: String?
    @Published private(set) var accountTimezone = TimeZone.current.identifier
    private let api: FinanceAPI

    init(api: FinanceAPI) { self.api = api }

    func load() async {
        guard !isLoading else { return }
        isLoading = true
        errorMessage = nil
        defer { isLoading = false }
        do {
            async let walletResult = api.wallets()
            async let profile = api.profile()
            wallets = try await walletResult
            if let profile = try? await profile { accountTimezone = profile.timezone }
        }
        catch { errorMessage = "Không thể tải ví." }
    }

    func delete(_ wallet: WalletModel) async {
        do {
            try await api.deleteWallet(id: wallet.id)
            wallets.removeAll { $0.id == wallet.id }
        } catch { errorMessage = "Không thể xóa ví. Ví có thể còn giao dịch liên kết." }
    }
}

struct WalletsView: View {
    @StateObject private var model: WalletsViewModel
    @State private var editorPresented = false
    @State private var editTarget: WalletModel?
    @State private var deleteTarget: WalletModel?
    @State private var selectedWalletID: String?

    init(api: FinanceAPI) { _model = StateObject(wrappedValue: WalletsViewModel(api: api)) }

    var body: some View {
        Group {
            if let selectedWalletID, let selectedWallet = model.wallets.first(where: { $0.id == selectedWalletID }) {
                WalletDetailView(api: model.apiForChild, wallet: selectedWallet) { self.selectedWalletID = nil }
            } else {
                walletList
            }
        }
        .task { await model.load() }
        .refreshable { await model.load() }
        .environment(\.timeZone, TimeZone(identifier: model.accountTimezone) ?? .current)
        .overlay {
            if editorPresented {
                WalletEditorView(api: model.apiForChild, wallet: editTarget, onSaved: {
                    editorPresented = false
                    Task { await model.load() }
                }, onCancel: { editorPresented = false })
                .background(MPColor.canvas.ignoresSafeArea())
            }
        }
        .mpConfirmDialog(isPresented: Binding(get: { deleteTarget != nil }, set: { if !$0 { deleteTarget = nil } }), title: "Xóa ví?", message: "Các giao dịch liên quan có thể ngăn việc xóa ví.", confirmTitle: "Xóa") {
            if let target = deleteTarget { Task { await model.delete(target) } }
            deleteTarget = nil
        }
        .accessibilityIdentifier("wallets-screen")
    }

    private var walletList: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                MPScreenHeader(title: "Ví", leading: { EmptyView() }, trailing: {
                    MPIconButton(systemName: "plus", label: "Thêm ví") {
                        editTarget = nil
                        editorPresented = true
                    }
                })
                if model.isLoading && model.wallets.isEmpty {
                    MPLoadingView(label: "Đang tải ví...")
                } else if let errorMessage = model.errorMessage, model.wallets.isEmpty {
                    MPErrorView(title: "Không tải được ví", detail: errorMessage) { Task { await model.load() } }
                } else if model.wallets.isEmpty {
                    MPEmptyView(title: "Chưa có ví", detail: "Tạo ví tiền mặt, tiết kiệm hoặc tín dụng để bắt đầu.")
                } else {
                    ForEach(model.wallets) { wallet in
                        MPSurfaceCard {
                            VStack(alignment: .leading, spacing: 10) {
                                    MPButton(style: .plain, action: { selectedWalletID = wallet.id }) {
                                    VStack(alignment: .leading, spacing: 10) {
                                        MPWalletRow(name: wallet.name, type: wallet.type, balance: MPFormat.money(wallet.currentBalance))
                                        if wallet.type == "goal", let target = wallet.targetAmount {
                                            MPProgressBar(progress: Double(wallet.currentBalance) / Double(max(1, target)), label: "Mục tiêu \(MPFormat.money(target))")
                                        }
                                        if wallet.type == "credit", let available = wallet.availableCredit {
                                            MPText("Khả dụng: \(MPFormat.money(available))", style: .secondary)
                                        }
                                    }
                                }
                                HStack {
                                    MPButton("Sửa", style: .secondary) { editTarget = wallet; editorPresented = true }
                                    MPButton("Xóa", style: .destructive) { deleteTarget = wallet }
                                }
                            }
                        }
                    }
                }
            }
            .padding(MPMetrics.pagePadding)
        }
        .background(MPColor.canvas.ignoresSafeArea())
    }
}

private extension WalletsViewModel {
    var apiForChild: FinanceAPI { api }
}

struct WalletDetailView: View {
    let api: FinanceAPI
    let wallet: WalletModel
    let onBack: () -> Void
    @State private var transactions: [TransactionModel] = []
    @State private var statement: CreditStatementModel?
    @State private var loading = false
    @State private var error: String?
    @Environment(\.timeZone) private var timeZone

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                MPScreenHeader(title: wallet.name, leading: { MPIconButton(systemName: "chevron.left", label: "Quay lại", action: onBack) }, trailing: { EmptyView() })
                MPMetricCard(title: "Số dư hiện tại", value: MPFormat.money(wallet.currentBalance))
                if wallet.type == "goal", let target = wallet.targetAmount {
                    let remaining = max(0, target - wallet.currentBalance)
                    MPSurfaceCard {
                        VStack(alignment: .leading, spacing: 8) {
                            MPText("Mục tiêu tiết kiệm", style: .title)
                            MPProgressBar(progress: Double(wallet.currentBalance) / Double(max(1, target)), label: remaining == 0 ? "Đã đạt mục tiêu" : "Còn thiếu \(MPFormat.money(remaining))")
                            if let targetDate = wallet.targetDate { MPText("Hạn mục tiêu: \(targetDate)", style: .secondary) }
                        }
                    }
                }
                if wallet.type == "credit" {
                    if let statement {
                        LazyVGrid(columns: [GridItem(.flexible()), GridItem(.flexible())], spacing: 12) {
                            MPMetricCard(title: "Dư nợ", value: MPFormat.money(statement.balance))
                            MPMetricCard(title: "Khả dụng", value: MPFormat.money(statement.availableCredit))
                            MPMetricCard(title: "Đến hạn", value: MPFormat.money(statement.amountDue))
                            MPMetricCard(title: "Trạng thái", value: statement.paymentStatus)
                        }
                        MPText("Giao dịch tín dụng", style: .title)
                        transactionCard(statement.items)
                    }
                } else {
                    MPText("Lịch sử giao dịch", style: .title)
                    transactionCard(transactions)
                }
                if loading { MPLoadingView(label: "Đang tải chi tiết...") }
                if let error { MPErrorView(title: "Không tải được chi tiết ví", detail: error) { Task { await load() } } }
            }
            .padding(MPMetrics.pagePadding)
        }
        .background(MPColor.canvas.ignoresSafeArea())
        .task { await load() }
        .accessibilityIdentifier("wallet-detail-screen")
    }

    @ViewBuilder
    private func transactionCard(_ rows: [TransactionModel]) -> some View {
        MPSurfaceCard {
            if rows.isEmpty {
                MPEmptyView(title: "Chưa có giao dịch", detail: "Các giao dịch của ví sẽ xuất hiện ở đây.")
            } else {
                VStack(spacing: 0) {
                    ForEach(rows) { row in
                        MPTransactionRow(title: row.note?.isEmpty == false ? row.note! : (row.type == "income" ? "Thu nhập" : "Chi tiêu"), subtitle: MPFormat.shortDate(row.occurredAt, timeZone: timeZone), amount: MPFormat.money(row.amount), isIncome: row.type == "income")
                    }
                }
            }
        }
    }

    private func load() async {
        guard !loading else { return }
        loading = true
        defer { loading = false }
        do {
            if wallet.type == "credit" {
                statement = try await api.creditStatement(walletID: wallet.id)
            } else {
                transactions = try await api.transactions().filter { $0.walletID == wallet.id }.sorted { $0.occurredAt > $1.occurredAt }
            }
            error = nil
        } catch {
            self.error = "Vui lòng thử lại."
        }
    }
}

struct WalletEditorView: View {
    let api: FinanceAPI
    let wallet: WalletModel?
    let onSaved: () -> Void
    let onCancel: () -> Void
    @State private var name: String
    @State private var type: String
    @State private var openingBalance: String
    @State private var targetAmount: String
    @State private var targetDate = Date()
    @State private var targetDateEnabled: String
    @State private var creditLimit: String
    @State private var lastStatementBalance: String
    @State private var statementDay: String
    @State private var paymentDueDay: String
    @State private var description: String
    @State private var isInTotal: String
    @State private var saving = false
    @State private var errorMessage: String?

    init(api: FinanceAPI, wallet: WalletModel?, onSaved: @escaping () -> Void, onCancel: @escaping () -> Void) {
        self.api = api
        self.wallet = wallet
        self.onSaved = onSaved
        self.onCancel = onCancel
        _name = State(initialValue: wallet?.name ?? "")
        _type = State(initialValue: wallet?.type ?? "basic")
        _openingBalance = State(initialValue: wallet.map { String($0.openingBalance) } ?? "")
        _targetAmount = State(initialValue: wallet?.targetAmount.map(String.init) ?? "")
        _targetDate = State(initialValue: MPFormat.parseCalendarDate(wallet?.targetDate) ?? Date())
        _targetDateEnabled = State(initialValue: wallet?.targetDate == nil ? "Không" : "Có")
        _creditLimit = State(initialValue: wallet?.creditLimit.map(String.init) ?? "")
        _lastStatementBalance = State(initialValue: wallet?.lastStatementBalance.map(String.init) ?? "")
        _statementDay = State(initialValue: wallet?.statementDay.map(String.init) ?? "")
        _paymentDueDay = State(initialValue: wallet?.paymentDueDay.map(String.init) ?? "")
        _description = State(initialValue: wallet?.description ?? "")
        _isInTotal = State(initialValue: wallet?.isInTotal == false ? "Không" : "Có")
    }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                MPScreenHeader(title: wallet == nil ? "Thêm ví" : "Sửa ví", leading: { MPIconButton(systemName: "xmark", label: "Đóng", action: onCancel) }, trailing: { EmptyView() })
                MPTextField(title: "Tên ví", text: $name, placeholder: "Ví tiền mặt")
                if wallet == nil {
                    MPWalletSelector(title: "Loại ví", options: ["basic", "goal", "credit"], selection: $type)
                    MPAmountField(title: "Số dư ban đầu", text: $openingBalance)
                } else {
                    MPSurfaceCard {
                        VStack(alignment: .leading, spacing: 6) {
                            MPText("Loại ví: \(type)", style: .secondary)
                            MPText("Số dư ban đầu: \(MPFormat.money(wallet?.openingBalance ?? 0))", style: .secondary)
                        }
                    }
                }
                if type == "goal" {
                    MPAmountField(title: "Mục tiêu", text: $targetAmount)
                    MPWalletSelector(title: "Có ngày mục tiêu", options: ["Có", "Không"], selection: $targetDateEnabled)
                    if targetDateEnabled == "Có" { MPDateField(title: "Ngày mục tiêu", date: $targetDate) }
                }
                if type == "credit" {
                    MPAmountField(title: "Hạn mức tín dụng", text: $creditLimit)
                    MPAmountField(title: "Dư nợ sao kê", text: $lastStatementBalance)
                    MPTextField(title: "Ngày sao kê (1-31)", text: $statementDay, placeholder: "Không bắt buộc")
                    MPTextField(title: "Ngày đến hạn (1-31)", text: $paymentDueDay, placeholder: "Không bắt buộc")
                }
                MPWalletSelector(title: "Tính vào tổng cộng", options: ["Có", "Không"], selection: $isInTotal)
                MPTextField(title: "Mô tả", text: $description, placeholder: "Không bắt buộc")
                if let errorMessage { MPStatusMessage(text: errorMessage, tone: .danger) }
                MPButton(wallet == nil ? "Tạo ví" : "Lưu thay đổi", isLoading: saving) { Task { await save() } }
            }
            .padding(MPMetrics.pagePadding)
        }
    }

    private func save() async {
        guard !name.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty else {
            errorMessage = "Tên ví không được để trống."
            return
        }
        let parsedTarget = Int64(targetAmount.filter(\.isNumber))
        let parsedCreditLimit = Int64(creditLimit.filter(\.isNumber))
        let parsedStatementBalance = Int64(lastStatementBalance.filter(\.isNumber))
        let parsedStatementDay = Int(statementDay.filter(\.isNumber))
        let parsedPaymentDueDay = Int(paymentDueDay.filter(\.isNumber))
        if type == "goal", (parsedTarget ?? 0) <= 0 { errorMessage = "Mục tiêu phải lớn hơn 0."; return }
        if type == "credit", (parsedCreditLimit ?? 0) <= 0 { errorMessage = "Hạn mức tín dụng phải lớn hơn 0."; return }
        if type == "credit", (parsedStatementBalance ?? 0) < 0 { errorMessage = "Dư nợ sao kê không được âm."; return }
        if type == "credit", [parsedStatementDay, parsedPaymentDueDay].compactMap({ $0 }).contains(where: { $0 < 1 || $0 > 31 }) { errorMessage = "Ngày sao kê và ngày đến hạn phải từ 1 đến 31."; return }
        saving = true
        defer { saving = false }
        let trimmedDescription = description.trimmingCharacters(in: .whitespacesAndNewlines)
        let input = WalletModelInput(name: name.trimmingCharacters(in: .whitespacesAndNewlines), type: wallet?.type ?? type, openingBalance: wallet?.openingBalance ?? Int64(openingBalance.filter(\.isNumber)) ?? 0, isInTotal: isInTotal == "Có", description: trimmedDescription.isEmpty ? nil : trimmedDescription, targetAmount: type == "goal" ? parsedTarget : nil, targetDate: type == "goal" && targetDateEnabled == "Có" ? MPFormat.calendarDateString(targetDate) : nil, creditLimit: type == "credit" ? parsedCreditLimit : nil, lastStatementBalance: type == "credit" ? parsedStatementBalance : nil, statementDay: type == "credit" ? parsedStatementDay : nil, paymentDueDay: type == "credit" ? parsedPaymentDueDay : nil)
        do {
            if let wallet { _ = try await api.updateWallet(id: wallet.id, input: input) }
            else { _ = try await api.createWallet(input) }
            onSaved()
        } catch { errorMessage = "Không thể lưu ví. Kiểm tra thông tin và thử lại." }
    }
}
