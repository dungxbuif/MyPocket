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
    @State private var editorPresentation: WalletEditorPresentation?
    @State private var deleteTarget: WalletModel?
    @State private var selectedWalletID: String?

    init(api: FinanceAPI) { _model = StateObject(wrappedValue: WalletsViewModel(api: api)) }

    var body: some View {
        Group {
            if let selectedWalletID, let selectedWallet = model.wallets.first(where: { $0.id == selectedWalletID }) {
                if selectedWallet.type == "credit" {
                    CreditView(api: model.apiForChild, onBack: { self.selectedWalletID = nil }, preferredWalletID: selectedWallet.id)
                } else {
                    WalletDetailView(api: model.apiForChild, wallet: selectedWallet) { self.selectedWalletID = nil }
                }
            } else {
                walletList
            }
        }
        .task { await model.load() }
        .refreshable { await model.load() }
        .environment(\.timeZone, TimeZone(identifier: model.accountTimezone) ?? .current)
        .overlay {
            if let editorPresentation {
                WalletEditorView(api: model.apiForChild, wallet: editorPresentation.wallet, timezone: model.accountTimezone, onSaved: {
                    self.editorPresentation = nil
                    Task { await model.load() }
                }, onCancel: { self.editorPresentation = nil })
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
                        editorPresentation = .create
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
                                    MPButton("Sửa", style: .secondary) { editorPresentation = .edit(wallet) }
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

private enum WalletEditorPresentation: Identifiable {
    case create
    case edit(WalletModel)

    var id: String {
        switch self {
        case .create: return "create"
        case .edit(let wallet): return "edit-\(wallet.id)"
        }
    }

    var wallet: WalletModel? {
        switch self {
        case .create: return nil
        case .edit(let wallet): return wallet
        }
    }
}

private extension WalletsViewModel {
    var apiForChild: FinanceAPI { api }
}

struct WalletDetailView: View {
    let api: FinanceAPI
    let onBack: () -> Void
    @State private var wallet: WalletModel
    @State private var transactions: [TransactionModel] = []
    @State private var categories: [CategoryModel] = []
    @State private var statement: CreditStatementModel?
    @State private var loading = false
    @State private var error: String?
    @State private var editorPresented = false
    @State private var editTarget: TransactionModel?
    @Environment(\.timeZone) private var timeZone

    init(api: FinanceAPI, wallet: WalletModel, onBack: @escaping () -> Void) {
        self.api = api
        self.onBack = onBack
        _wallet = State(initialValue: wallet)
    }

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
                    MPSurfaceCard {
                        HStack(spacing: 12) {
                            VStack(alignment: .leading, spacing: 4) {
                                MPText("Tiền vào", style: .secondary)
                                MPText(MPFormat.money(totalIncome), style: .title)
                            }
                            Spacer()
                            VStack(alignment: .trailing, spacing: 4) {
                                MPText("Tiền ra", style: .secondary)
                                MPText(MPFormat.money(totalExpense), style: .title)
                            }
                        }
                    }
                    MPText("TẤT CẢ CÁC GIAO DỊCH", style: .title)
                    if groupedTransactions.isEmpty {
                        MPSurfaceCard {
                            MPEmptyView(title: "Chưa có giao dịch", detail: "Các giao dịch của ví sẽ xuất hiện ở đây.")
                        }
                    } else {
                        ForEach(groupedTransactions, id: \.0) { month, rows in
                            MPSurfaceCard {
                                VStack(alignment: .leading, spacing: 8) {
                                    HStack {
                                        MPText(monthLabel(month), style: .secondary)
                                        Spacer()
                                        MPText(MPFormat.signedMoney(rows.reduce(0) { $0 + signedAmount($1) }), style: .secondary)
                                    }
                                    ForEach(rows) { row in
                                        walletTransactionRow(row)
                                    }
                                }
                            }
                        }
                    }
                    MPButton("Thêm giao dịch") { editTarget = nil; editorPresented = true }
                }
                if loading { MPLoadingView(label: "Đang tải chi tiết...") }
                if let error { MPErrorView(title: "Không tải được chi tiết ví", detail: error) { Task { await load() } } }
            }
            .padding(MPMetrics.pagePadding)
        }
        .background(MPColor.canvas.ignoresSafeArea())
        .task { await load() }
        .overlay {
            if wallet.type != "credit", editorPresented {
                TransactionEditorView(
                    api: api,
                    wallets: [wallet],
                    categories: categories,
                    timezone: timeZone.identifier,
                    transaction: editTarget,
                    onSaved: {
                        editorPresented = false
                        editTarget = nil
                        Task { await load() }
                    },
                    onCancel: { editorPresented = false; editTarget = nil }
                )
                .background(MPColor.canvas.ignoresSafeArea())
            }
        }
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

    @ViewBuilder
    private func walletTransactionRow(_ row: TransactionModel) -> some View {
        let category = categories.first(where: { $0.id == row.categoryID })
        let title = category?.name ?? (row.type == "income" ? "Khoản thu" : "Khoản chi")
        let subtitle = [MPFormat.shortDate(row.occurredAt, timeZone: timeZone), row.note].compactMap { $0?.isEmpty == false ? $0 : nil }.joined(separator: " · ")
        MPButton(style: .plain) {
            editTarget = row
            editorPresented = true
        } label: {
            MPTransactionRow(title: title, subtitle: subtitle, amount: MPFormat.signedMoney(signedAmount(row)), isIncome: row.type == "income")
        }
    }

    private var walletTransactions: [TransactionModel] { transactions.filter { $0.walletID == wallet.id } }
    private var totalIncome: Int64 { walletTransactions.filter { $0.type == "income" }.reduce(0) { $0 + $1.amount } }
    private var totalExpense: Int64 { walletTransactions.filter { $0.type == "expense" }.reduce(0) { $0 + $1.amount } }
    private var groupedTransactions: [(String, [TransactionModel])] {
        let grouped = Dictionary(grouping: walletTransactions) { row in
            MPFormat.accountMonth(of: row.occurredAt, timeZone: timeZone) ?? "unknown"
        }
        return grouped.map { ($0.key, $0.value.sorted { $0.occurredAt > $1.occurredAt }) }.sorted { $0.0 > $1.0 }
    }

    private func monthLabel(_ value: String) -> String {
        MPFormat.monthLabel(value)
    }

    private func signedAmount(_ row: TransactionModel) -> Int64 {
        row.type == "income" ? row.amount : -row.amount
    }

    private func load() async {
        guard !loading else { return }
        loading = true
        defer { loading = false }
        do {
            async let walletResult = api.wallets()
            if wallet.type == "credit" {
                statement = try await api.creditStatement(walletID: wallet.id)
            } else {
                async let transactionResult = api.transactions()
                async let categoryResult = api.categories()
                transactions = try await transactionResult
                categories = (try? await categoryResult) ?? []
            }
            if let refreshedWallets = try? await walletResult,
               let refreshed = refreshedWallets.first(where: { $0.id == wallet.id }) {
                wallet = refreshed
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
    let timezone: String
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

    init(api: FinanceAPI, wallet: WalletModel?, timezone: String = TimeZone.current.identifier, onSaved: @escaping () -> Void, onCancel: @escaping () -> Void) {
        self.api = api
        self.wallet = wallet
        self.timezone = timezone
        self.onSaved = onSaved
        self.onCancel = onCancel
        _name = State(initialValue: wallet?.name ?? "")
        _type = State(initialValue: wallet?.type ?? "basic")
        _openingBalance = State(initialValue: wallet.map { String($0.openingBalance) } ?? "")
        _targetAmount = State(initialValue: wallet?.targetAmount.map(String.init) ?? "")
        _targetDate = State(initialValue: MPFormat.parseCalendarDate(wallet?.targetDate, timeZone: TimeZone(identifier: timezone) ?? .current) ?? Date())
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
        .environment(\.timeZone, TimeZone(identifier: timezone) ?? .current)
    }

    private func save() async {
        guard !name.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty else {
            errorMessage = "Tên ví không được để trống."
            return
        }
        let trimmedTarget = targetAmount.trimmingCharacters(in: .whitespacesAndNewlines)
        let trimmedCreditLimit = creditLimit.trimmingCharacters(in: .whitespacesAndNewlines)
        let trimmedStatementBalance = lastStatementBalance.trimmingCharacters(in: .whitespacesAndNewlines)
        let trimmedStatementDay = statementDay.trimmingCharacters(in: .whitespacesAndNewlines)
        let trimmedPaymentDueDay = paymentDueDay.trimmingCharacters(in: .whitespacesAndNewlines)
        let parsedOpeningBalance = MPFormat.parseNonNegativeInteger(openingBalance, allowingEmpty: true)
        let parsedTarget = MPFormat.parseNonNegativeInteger(trimmedTarget)
        let parsedCreditLimit = MPFormat.parseNonNegativeInteger(trimmedCreditLimit)
        let parsedStatementBalance = MPFormat.parseNonNegativeInteger(trimmedStatementBalance)
        let parsedStatementDay = trimmedStatementDay.isEmpty ? nil : MPFormat.parseNonNegativeInteger(trimmedStatementDay).flatMap(Int.init(exactly:))
        let parsedPaymentDueDay = trimmedPaymentDueDay.isEmpty ? nil : MPFormat.parseNonNegativeInteger(trimmedPaymentDueDay).flatMap(Int.init(exactly:))
        if wallet == nil, parsedOpeningBalance == nil {
            errorMessage = "Số dư ban đầu phải là số nguyên không âm."
            return
        }
        if (!trimmedTarget.isEmpty && parsedTarget == nil) || (!trimmedCreditLimit.isEmpty && parsedCreditLimit == nil) || (!trimmedStatementBalance.isEmpty && parsedStatementBalance == nil) || (!trimmedStatementDay.isEmpty && parsedStatementDay == nil) || (!trimmedPaymentDueDay.isEmpty && parsedPaymentDueDay == nil) {
            errorMessage = "Các số tiền và ngày thanh toán phải là giá trị hợp lệ."
            return
        }
        if type == "goal", (parsedTarget ?? 0) <= 0 { errorMessage = "Mục tiêu phải lớn hơn 0."; return }
        if type == "credit", (parsedCreditLimit ?? 0) <= 0 { errorMessage = "Hạn mức tín dụng phải lớn hơn 0."; return }
        if type == "credit", (parsedStatementBalance ?? 0) < 0 { errorMessage = "Dư nợ sao kê không được âm."; return }
        if type == "credit", [parsedStatementDay, parsedPaymentDueDay].compactMap({ $0 }).contains(where: { $0 < 1 || $0 > 31 }) { errorMessage = "Ngày sao kê và ngày đến hạn phải từ 1 đến 31."; return }
        saving = true
        defer { saving = false }
        let trimmedDescription = description.trimmingCharacters(in: .whitespacesAndNewlines)
        let input = WalletModelInput(name: name.trimmingCharacters(in: .whitespacesAndNewlines), type: wallet?.type ?? type, openingBalance: wallet?.openingBalance ?? (parsedOpeningBalance ?? 0), isInTotal: isInTotal == "Có", description: trimmedDescription.isEmpty ? nil : trimmedDescription, targetAmount: type == "goal" ? parsedTarget : nil, targetDate: type == "goal" && targetDateEnabled == "Có" ? MPFormat.calendarDateString(targetDate, timeZone: TimeZone(identifier: timezone) ?? .current) : nil, creditLimit: type == "credit" ? parsedCreditLimit : nil, lastStatementBalance: type == "credit" ? parsedStatementBalance : nil, statementDay: type == "credit" ? parsedStatementDay : nil, paymentDueDay: type == "credit" ? parsedPaymentDueDay : nil)
        do {
            if let wallet { _ = try await api.updateWallet(id: wallet.id, input: input) }
            else { _ = try await api.createWallet(input) }
            onSaved()
        } catch { errorMessage = "Không thể lưu ví. Kiểm tra thông tin và thử lại." }
    }
}
