import SwiftUI
import Combine

/// Stable composition contract for the Stitch wallet-ledger screen.
enum WalletLedgerReferenceSection: String, CaseIterable {
    case header
    case balance
    case summary
    case allTransactions

    static let stitchOrder: [WalletLedgerReferenceSection] = [
        .header,
        .balance,
        .summary,
        .allTransactions
    ]
}

/// Stable composition contract for the Stitch wallet editor form.
enum WalletEditorReferenceSection: String, CaseIterable {
    case header
    case identity
    case type
    case specialized
    case settings
    case save

    static let stitchOrder: [WalletEditorReferenceSection] = [
        .header,
        .identity,
        .type,
        .specialized,
        .settings,
        .save
    ]
}

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

    func delete(_ wallet: WalletModel) async -> Bool {
        do {
            try await api.deleteWallet(id: wallet.id)
            wallets.removeAll { $0.id == wallet.id }
            return true
        } catch {
            errorMessage = "Không thể xóa ví. Ví có thể còn giao dịch liên kết."
            return false
        }
    }
}

struct WalletsView: View {
    @StateObject private var model: WalletsViewModel
    let onSelectWallet: (String?) -> Void
    let onBack: (() -> Void)?
    let onAIEntry: (() -> Void)?
    let initialWalletID: String?
    @State private var editorPresentation: WalletEditorPresentation?
    @State private var deleteTarget: WalletModel?
    @State private var selectedWalletID: String?
    @State private var isEditing = false

    init(
        api: FinanceAPI,
        initialWalletID: String? = nil,
        onSelectWallet: @escaping (String?) -> Void = { _ in },
        onBack: (() -> Void)? = nil,
        onAIEntry: (() -> Void)? = nil
    ) {
        _model = StateObject(wrappedValue: WalletsViewModel(api: api))
        self.initialWalletID = initialWalletID
        self.onSelectWallet = onSelectWallet
        self.onBack = onBack
        self.onAIEntry = onAIEntry
    }

    var body: some View {
        Group {
            if let editorPresentation {
                WalletEditorView(api: model.apiForChild, wallet: editorPresentation.wallet, timezone: model.accountTimezone, onDelete: {
                    if let wallet = editorPresentation.wallet { deleteTarget = wallet }
                }, onSaved: {
                    self.editorPresentation = nil
                    Task { await model.load() }
                }, onCancel: { self.editorPresentation = nil })
                .background(MPColor.canvas.ignoresSafeArea())
            } else if let selectedWalletID, let selectedWallet = model.wallets.first(where: { $0.id == selectedWalletID }) {
                if selectedWallet.type == "credit" {
                    CreditView(api: model.apiForChild, onBack: { self.selectedWalletID = nil }, preferredWalletID: selectedWallet.id)
                } else {
                        WalletDetailView(api: model.apiForChild, wallet: selectedWallet, onAIEntry: onAIEntry) { self.selectedWalletID = nil }
                }
            } else {
                walletList
            }
        }
        .task {
            await model.load()
            if selectedWalletID == nil, let initialWalletID,
               model.wallets.contains(where: { $0.id == initialWalletID }) {
                selectedWalletID = initialWalletID
            }
        }
        .onChange(of: initialWalletID) { _, walletID in
            // A second deep link can arrive while the wallet scene is already
            // mounted; update the detail selection without reloading the API.
            selectedWalletID = walletID
        }
        .refreshable { await model.load() }
        .environment(\.timeZone, TimeZone(identifier: model.accountTimezone) ?? .current)
        .mpConfirmDialog(isPresented: Binding(get: { deleteTarget != nil }, set: { if !$0 { deleteTarget = nil } }), title: "Xóa ví?", message: "Các giao dịch liên quan có thể ngăn việc xóa ví.", confirmTitle: "Xóa") {
            if let target = deleteTarget {
                Task {
                    _ = await model.delete(target)
                    editorPresentation = nil
                    isEditing = false
                }
            }
            deleteTarget = nil
        }
    }

    private var walletList: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                MPSelectionScreenHeader(
                    title: "Chọn Ví",
                    closeTitle: "Đóng",
                    closeAccessibilityLabel: "Quay lại",
                    closeAccessibilityIdentifier: "screen-back",
                    actionTitle: isEditing ? "Xong" : "Sửa",
                    actionAccessibilityIdentifier: "wallet-edit-mode",
                    onClose: onBack.map { back in
                        {
                            isEditing = false
                            back()
                        }
                    },
                    onAction: { isEditing.toggle() }
                )
                if !model.wallets.isEmpty, let errorMessage = model.errorMessage {
                    MPStatusMessage(text: errorMessage, tone: .danger)
                }
                if model.isLoading && model.wallets.isEmpty {
                    MPLoadingView(label: "Đang tải ví...")
                } else if let errorMessage = model.errorMessage, model.wallets.isEmpty {
                    MPErrorView(title: "Không tải được ví", detail: errorMessage) { Task { await model.load() } }
                } else {
                    if model.wallets.isEmpty {
                        MPEmptyView(title: "Chưa có ví", detail: "Tạo ví tiền mặt, tiết kiệm hoặc tín dụng để bắt đầu.")
                    }
                    MPWalletSelectionList(
                        wallets: model.wallets,
                        isEditing: isEditing,
                        onSelect: { wallet in
                            isEditing = false
                            onSelectWallet(wallet?.id)
                        },
                        onDetail: { selectedWalletID = $0.id },
                        onEdit: { editorPresentation = .edit($0) },
                        onAdd: { editorPresentation = .create }
                    )
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

enum WalletQueueEditPolicy {
    static func canEdit(_ payload: OutboxPayload, walletType: String, detailsReady: Bool) -> Bool {
        guard walletType != "credit" else { return false }
        switch payload {
        case .transfer:
            // Transfer editing only needs the wallet list. A transient empty
            // category response must not hide recovery for a savings wallet.
            return true
        case .transaction:
            return detailsReady
        }
    }
}

struct WalletDetailView: View {
    @Environment(\.mpFinancialQueue) private var financialQueue
    let api: FinanceAPI
    let onBack: () -> Void
    let onAIEntry: (() -> Void)?
    let onOpenReports: (() -> Void)?
    let onSelectWallet: ((String?) -> Void)?
    let embedded: Bool
    @State private var wallet: WalletModel
    @State private var wallets: [WalletModel] = []
    @State private var transactions: [TransactionModel] = []
    @State private var categories: [CategoryModel] = []
    @State private var statement: CreditStatementModel?
    @State private var loading = false
    @State private var detailsReady = false
    @State private var error: String?
    @State private var editorPresented = false
    @State private var editTarget: TransactionModel?
    @State private var queuedEditTarget: OutboxMutation?
    @State private var selectedTransaction: TransactionModel?
    @State private var transferEditID: String?
    @State private var deleteTarget: TransactionModel?
    @State private var searchQuery = ""
    @State private var searchPresented = false
    @State private var helpPresented = false
    @Environment(\.timeZone) private var timeZone
    @Environment(\.mpFinancialValuesMasked) private var masked

    init(api: FinanceAPI, wallet: WalletModel, embedded: Bool = false, onAIEntry: (() -> Void)? = nil, onOpenReports: (() -> Void)? = nil, onSelectWallet: ((String?) -> Void)? = nil, onBack: @escaping () -> Void) {
        self.api = api
        self.onBack = onBack
        self.onAIEntry = onAIEntry
        self.onOpenReports = onOpenReports
        self.onSelectWallet = onSelectWallet
        self.embedded = embedded
        _wallet = State(initialValue: wallet)
    }

    @ViewBuilder
    var body: some View {
        Group {
            if wallet.type != "credit", let transferEditID {
                TransferEditorView(
                    api: api,
                    wallets: wallets.isEmpty ? [wallet] : wallets,
                    timezone: timeZone.identifier,
                    transactions: transactions,
                    existingTransferID: transferEditID,
                    onSaved: {
                        self.transferEditID = nil
                        Task { await load() }
                    },
                    onCancel: { self.transferEditID = nil }
                )
                .background(MPColor.canvas.ignoresSafeArea())
            } else if wallet.type != "credit", let selectedTransaction {
                TransactionDetailView(
                    api: api,
                    transaction: selectedTransaction,
                    allTransactions: transactions,
                    isCreditWallet: false,
                    onEdit: {
                        editTarget = selectedTransaction
                        self.selectedTransaction = nil
                        editorPresented = true
                    },
                    onEditTransfer: {
                        transferEditID = selectedTransaction.transferID
                        self.selectedTransaction = nil
                    },
                    onDelete: {
                        deleteTarget = selectedTransaction
                        self.selectedTransaction = nil
                    },
                    onDeleteTransfer: {
                        deleteTarget = selectedTransaction
                        self.selectedTransaction = nil
                    },
                    onClose: { self.selectedTransaction = nil }
                )
                .background(MPColor.canvas.ignoresSafeArea())
            } else {
                walletContent
            }
        }
        .onReceive(financialQueue?.$revision.dropFirst().eraseToAnyPublisher() ?? Empty<Int, Never>().eraseToAnyPublisher()) { _ in
            Task<Void, Never> { @MainActor in await load() }
        }
        .mpFullScreenCover(isPresented: $editorPresented) { editorContent }
    }

    @ViewBuilder
    private var editorContent: some View {
        if let queuedEditTarget, case .transfer = queuedEditTarget.payload {
            TransferEditorView(api: api, wallets: wallets.isEmpty ? [wallet] : wallets, timezone: timeZone.identifier,
                queuedMutation: queuedEditTarget, onSaved: editorSaved, onCancel: closeEditor)
            .background(MPColor.canvas.ignoresSafeArea())
        } else {
            TransactionEditorView(
                api: api,
                wallets: queuedEditTarget == nil ? [wallet] : (wallets.isEmpty ? [wallet] : wallets),
                categories: categories,
                timezone: timeZone.identifier,
                transaction: editTarget,
                queuedMutation: queuedEditTarget,
                onSaved: editorSaved,
                onCancel: closeEditor,
                onAIEntry: aiEntryAction
            )
            .background(MPColor.canvas.ignoresSafeArea())
        }
    }

    private var aiEntryAction: (() -> Void)? {
        guard queuedEditTarget == nil else { return nil }
        return { enterAIEntry() }
    }

    private func editorSaved() {
        closeEditor()
        Task<Void, Never> { @MainActor in await load() }
    }

    private func closeEditor() {
        editorPresented = false; editTarget = nil; queuedEditTarget = nil
    }

    private func enterAIEntry() { closeEditor(); onAIEntry?() }

    @ViewBuilder
    private var pendingQueue: some View {
        if wallet.type != "credit", let financialQueue, !financialQueue.items.isEmpty {
            MPFinancialQueueStatus(queue: financialQueue, onEdit: pendingEditAction,
                canEdit: { WalletQueueEditPolicy.canEdit($0.payload, walletType: wallet.type, detailsReady: detailsReady) },
                onDiscarded: { Task<Void, Never> { @MainActor in await load() } })
        }
    }

    private var pendingEditAction: ((OutboxMutation) -> Void)? {
        return { item in editTarget = nil; queuedEditTarget = item; editorPresented = true }
    }

    @ViewBuilder
    private var walletContent: some View {
        if embedded && wallet.type == "goal" {
            goalLedgerContent
        } else {
            ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                if !embedded {
                    MPScreenHeader(title: wallet.name, leading: { MPIconButton(systemName: "chevron.left", label: "Quay lại", action: onBack) }, trailing: { EmptyView() })
                        .accessibilityIdentifier("wallet-detail-header")
                }
                pendingQueue
                if wallet.type != "goal" {
                    MPMetricCard(title: "Số dư hiện tại", value: MPFormat.money(wallet.currentBalance, masked: masked))
                        .accessibilityIdentifier("wallet-detail-balance")
                }
                if wallet.type == "goal", let target = wallet.targetAmount {
                    MPSavingsSummary(
                        balance: wallet.currentBalance,
                        target: target,
                        targetDate: wallet.targetDate,
                        timeZone: timeZone,
                        layout: .card,
                        masked: masked
                    )
                    .accessibilityIdentifier("wallet-detail-goal-progress")
                }
                if wallet.type == "credit" {
                    if let statement {
                        LazyVGrid(columns: [GridItem(.flexible()), GridItem(.flexible())], spacing: 12) {
                            MPMetricCard(title: "Dư nợ", value: MPFormat.money(abs(statement.balance), masked: masked))
                            MPMetricCard(title: "Khả dụng", value: MPFormat.money(statement.availableCredit, masked: masked))
                            MPMetricCard(title: "Hạn mức", value: MPFormat.money(statement.creditLimit, masked: masked))
                            MPMetricCard(title: "Đến hạn", value: MPFormat.money(statement.amountDue, masked: masked))
                            MPMetricCard(title: "Trạng thái", value: MPSelectionLabels.display(for: statement.paymentStatus))
                        }
                        if statement.paymentStatus == "not_configured" {
                            MPText("Chưa cấu hình dư nợ sao kê và ngày đến hạn.", style: .secondary)
                        } else if let dueAt = statement.paymentDueAt {
                            MPText("Hạn thanh toán: \(MPFormat.shortDate(dueAt, timeZone: timeZone))", style: .secondary)
                        }
                        MPText("Giao dịch tín dụng", style: .title)
                            .accessibilityIdentifier("wallet-detail-credit-transactions")
                        transactionCard(statement.items)
                    }
                } else if wallet.type != "goal" {
                    MPSurfaceCard(style: .reference) {
                        HStack(spacing: 12) {
                            VStack(alignment: .leading, spacing: 4) {
                                MPText("Tiền vào", style: .secondary)
                                MPText(MPFormat.money(totalIncome, masked: masked), style: .title)
                            }
                            Spacer()
                            VStack(alignment: .trailing, spacing: 4) {
                                MPText("Tiền ra", style: .secondary)
                                MPText(MPFormat.money(totalExpense, masked: masked), style: .title)
                            }
                        }
                    }
                    .accessibilityIdentifier("wallet-detail-summary")
                }
                if wallet.type != "credit" {
                    MPText("TẤT CẢ CÁC GIAO DỊCH", style: .title)
                        .accessibilityIdentifier("wallet-detail-all-transactions")
                    if groupedTransactions.isEmpty {
                        MPSurfaceCard {
                            MPEmptyView(title: "Chưa có giao dịch", detail: "Các giao dịch của ví sẽ xuất hiện ở đây.")
                        }
                    } else {
                        ForEach(groupedTransactions, id: \.0) { month, rows in
                            MPSurfaceCard(style: .reference) {
                                VStack(alignment: .leading, spacing: 8) {
                                    HStack {
                                        MPText(monthLabel(month), style: .secondary)
                                            .accessibilityIdentifier("wallet-detail-ledger-\(month)")
                                        Spacer()
                                        MPText(MPFormat.signedMoney(rows.reduce(0) { $0 + signedAmount($1) }, masked: masked), style: .secondary)
                                    }
                                    ForEach(rows) { row in
                                        walletTransactionRow(row)
                                    }
                                }
                            }
                        }
                    }
                    if !detailsReady, !loading, error == nil {
                        MPStatusMessage(text: "Chưa tải được danh mục giao dịch.", tone: .danger)
                    }
                }
                if loading { MPLoadingView(label: "Đang tải chi tiết...") }
                if let error { MPErrorView(title: "Không tải được chi tiết ví", detail: error) { Task { await load() } } }
            }
            .padding(MPMetrics.pagePadding)
        }
        .background(MPColor.canvas.ignoresSafeArea())
        // Keep the wallet CTA above the shell navigation and out of the
        // scrollable ledger. A nested safe-area inset reserves a real hit
        // region, so the shell's raised add button cannot intercept it on
        // compact iPhones.
        .safeAreaInset(edge: .bottom, spacing: 8) {
            if detailsReady {
                MPButton("Thêm giao dịch") { editTarget = nil; editorPresented = true }
                    .accessibilityIdentifier("wallet-detail-add-transaction")
                    .padding(.horizontal, MPMetrics.pagePadding)
                    // The shell adds its own raised navigation inset. Keep
                    // this nested CTA comfortably above that hit region
                    // instead of relying on exact safe-area arithmetic.
                    .padding(.bottom, 64)
                    .background(MPColor.canvas)
            }
        }
        .task { await load() }
            .mpConfirmDialog(
            isPresented: Binding(get: { deleteTarget != nil }, set: { if !$0 { deleteTarget = nil } }),
            title: deleteTarget?.transferID == nil ? "Xóa giao dịch?" : "Xóa cả cặp chuyển ví?",
            message: deleteTarget?.transferID == nil
                ? "Hành động này không thể hoàn tác."
                : "Cả hai dòng của cặp chuyển ví sẽ bị xóa.",
            confirmTitle: "Xóa"
        ) {
            if let target = deleteTarget {
                Task {
                    if target.transferID != nil {
                        await deleteTransfer(target)
                    } else {
                        await deleteTransaction(target)
                    }
                }
            }
            deleteTarget = nil
            }
        }
    }

    @ViewBuilder
    private var goalLedgerContent: some View {
        ScrollView {
            VStack(spacing: 16) {
                goalLedgerHeader
                pendingQueue
                if searchPresented {
                    MPTextField(title: "Tìm kiếm giao dịch", text: $searchQuery, placeholder: "Tên danh mục hoặc ghi chú")
                        .accessibilityIdentifier("wallet-detail-goal-search")
                }
                MPSavingsBalanceBlock(balance: wallet.currentBalance, masked: masked)
                HStack {
                    MPText("TẤT CẢ CÁC GIAO DỊCH", style: .title)
                    Spacer()
                }
                .padding(.top, 4)
                .accessibilityIdentifier("wallet-detail-all-transactions")
                MPSavingsProgressBlock(
                    progress: MPFormat.savingsProgress(balance: wallet.currentBalance, target: wallet.targetAmount ?? 0),
                    deadlineLabel: goalDeadlineLabel,
                    masked: masked
                )
                .accessibilityIdentifier("wallet-detail-goal-progress")
                goalSavingsAdvice
                if let onOpenReports {
                    MPTextLinkButton("Xem báo cáo cho giai đoạn này", action: onOpenReports)
                        .accessibilityIdentifier("wallet-detail-goal-report")
                }
                goalLedgerGroups
            }
            .padding(MPMetrics.pagePadding)
            .padding(.bottom, 16)
        }
        .background(MPColor.canvas.ignoresSafeArea())
        // Keep the goal CTA in a reserved safe-area region. The persistent
        // shell owns a raised add affordance, so an in-scroll button can be
        // visually visible yet still have its label/hit target covered.
        .safeAreaInset(edge: .bottom, spacing: 8) {
                MPButton("Thêm giao dịch") { editTarget = nil; editorPresented = true }
                    .accessibilityIdentifier("wallet-detail-add-transaction")
                .padding(.horizontal, MPMetrics.pagePadding)
                .padding(.bottom, 104)
                .background(MPColor.canvas)
                .offset(y: -56)
        }
        .task { await load() }
        .mpInfoAlert(
            isPresented: $helpPresented,
            title: "Ví tiết kiệm",
            message: "Theo dõi số dư, tiến độ và các giao dịch đã dành cho mục tiêu này."
        )
        .mpConfirmDialog(
            isPresented: Binding(get: { deleteTarget != nil }, set: { if !$0 { deleteTarget = nil } }),
            title: deleteTarget?.transferID == nil ? "Xóa giao dịch?" : "Xóa cả cặp chuyển ví?",
            message: deleteTarget?.transferID == nil
                ? "Hành động này không thể hoàn tác."
                : "Cả hai dòng của cặp chuyển ví sẽ bị xóa.",
            confirmTitle: "Xóa"
        ) {
            if let target = deleteTarget {
                Task {
                    if target.transferID != nil {
                        await deleteTransfer(target)
                    } else {
                        await deleteTransaction(target)
                    }
                }
            }
            deleteTarget = nil
        }
    }

    private var goalLedgerHeader: some View {
        HStack(spacing: 8) {
            MPHeaderIconButton(systemName: "questionmark", label: "Trợ giúp") { helpPresented = true }
            Spacer(minLength: 4)
            MPMenuSelector(
                options: wallets.filter { $0.type == "goal" }.map { MPSelectionOption(id: $0.id, label: $0.name) }
                    + [MPSelectionOption(id: "__all__", label: "Tất cả giao dịch")],
                onSelect: { selection in onSelectWallet?(selection == "__all__" ? nil : selection) }
            ) {
                HStack(spacing: 8) {
                    MPWalletGlyph(type: wallet.type, size: 32)
                    MPText(wallet.name, style: .emphasis)
                        .lineLimit(1)
                    MPIcon(systemName: "chevron.up.chevron.down", color: MPColor.ink, size: 18)
                }
                .padding(.horizontal, 12)
                .frame(minHeight: 48)
                .background(MPColor.surface, in: Capsule())
                .shadow(color: MPColor.ink.opacity(0.04), radius: 8, y: 3)
            }
            .accessibilityLabel("Ví: \(wallet.name)")
            .accessibilityIdentifier("wallet-detail-goal-wallet-selector")
            Spacer(minLength: 4)
            HStack(spacing: 0) {
                MPIconButton(systemName: "magnifyingglass", label: "Tìm kiếm") {
                    withAnimation(.easeInOut(duration: 0.18)) { searchPresented.toggle() }
                }
                MPOverflowMenu(label: "Tùy chọn ví", actions: [
                    MPOverflowAction(title: "Tải lại giao dịch") { Task { await load() } },
                    MPOverflowAction(title: "Chọn ví khác") { onBack() }
                ])
            }
            .padding(.horizontal, 4)
            .background(MPColor.surface, in: Capsule())
            .shadow(color: MPColor.ink.opacity(0.04), radius: 8, y: 3)
        }
        .frame(maxWidth: .infinity)
    }

    private var goalSavingsAdvice: some View {
        let progress = MPFormat.savingsProgress(balance: wallet.currentBalance, target: wallet.targetAmount ?? 0)
        return Group {
            if progress.remaining > 0 {
                MPSurfaceCard {
                    HStack(alignment: .top, spacing: 12) {
                        MPIcon(systemName: "lightbulb", color: MPColor.ink, size: 28)
                        MPText("Bạn nên tiết kiệm thêm \(MPFormat.money(progress.remaining, masked: masked)) nữa trong tháng này.", style: .body)
                            .fixedSize(horizontal: false, vertical: true)
                    }
                }
                .background(MPColor.successSoft)
                .accessibilityIdentifier("wallet-detail-goal-advice")
            }
        }
    }

    private var goalDeadlineLabel: String {
        guard let days = MPFormat.calendarDayDifference(to: wallet.targetDate, timeZone: timeZone) else {
            return "Chưa đặt ngày kết thúc"
        }
        if days < 0 { return "Đã qua hạn \(-days) ngày" }
        if days == 0 { return "Đến hạn hôm nay" }
        return "Còn \(days) ngày"
    }

    @ViewBuilder
    private var goalLedgerGroups: some View {
        if filteredGoalTransactions.isEmpty {
            MPSurfaceCard {
                MPEmptyView(title: "Chưa có giao dịch", detail: "Các giao dịch của ví sẽ xuất hiện ở đây.")
            }
        } else {
            ForEach(groupedGoalTransactions, id: \.0) { month, rows in
                MPSurfaceCard(style: .reference) {
                    VStack(alignment: .leading, spacing: 8) {
                        HStack {
                            MPText(monthLabel(month), style: .title)
                            Spacer()
                            MPText(MPFormat.signedMoney(rows.reduce(0) { $0 + signedAmount($1) }, masked: masked), style: .emphasis)
                        }
                        ForEach(rows) { row in
                            walletTransactionRow(row)
                        }
                    }
                }
            }
        }
    }

    @ViewBuilder
    private func transactionCard(_ rows: [TransactionModel]) -> some View {
        MPSurfaceCard {
            if rows.isEmpty {
                MPEmptyView(title: "Chưa có giao dịch", detail: "Các giao dịch của ví sẽ xuất hiện ở đây.")
            } else {
                VStack(spacing: 0) {
                    ForEach(rows) { row in
                        walletTransactionRow(row)
                    }
                }
            }
        }
    }

    @ViewBuilder
    private func walletTransactionRow(_ row: TransactionModel) -> some View {
        let category = categories.first(where: { $0.id == row.categoryID })
        let content = MPTransactionRow(transaction: row, category: category, timeZone: timeZone, masked: masked)
        MPButton(style: .plain) {
            selectedTransaction = row
        } label: {
            content
        }
        // Expose the row as one actionable element with its persisted amount
        // in the semantic label. This keeps VoiceOver and UI automation
        // independent from SwiftUI's nested text aggregation.
        .accessibilityIdentifier("wallet-transaction-\(row.id)")
        .accessibilityLabel("\(content.title), \(content.subtitle), \(content.amount)")
    }

    private var walletTransactions: [TransactionModel] { transactions.filter { $0.walletID == wallet.id } }
    private var filteredGoalTransactions: [TransactionModel] {
        let query = searchQuery.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !query.isEmpty else { return walletTransactions }
        return walletTransactions.filter { row in
            let category = categories.first(where: { $0.id == row.categoryID })?.name ?? ""
            return category.localizedCaseInsensitiveContains(query)
                || (row.note ?? "").localizedCaseInsensitiveContains(query)
                || row.type.localizedCaseInsensitiveContains(query)
        }
    }
    private var totalIncome: Int64 { walletTransactions.filter { $0.type == "income" }.reduce(0) { $0 + $1.amount } }
    private var totalExpense: Int64 { walletTransactions.filter { $0.type == "expense" }.reduce(0) { $0 + $1.amount } }
    private var groupedTransactions: [(String, [TransactionModel])] {
        let grouped = Dictionary(grouping: walletTransactions) { row in
            MPFormat.accountMonth(of: row.occurredAt, timeZone: timeZone) ?? "unknown"
        }
        return grouped.map { ($0.key, $0.value.sorted { $0.occurredAt > $1.occurredAt }) }.sorted { $0.0 > $1.0 }
    }
    private var groupedGoalTransactions: [(String, [TransactionModel])] {
        let grouped = Dictionary(grouping: filteredGoalTransactions) { row in
            MPFormat.accountMonth(of: row.occurredAt, timeZone: timeZone) ?? "unknown"
        }
        return grouped.map { ($0.key, $0.value.sorted { $0.occurredAt > $1.occurredAt }) }.sorted { $0.0 > $1.0 }
    }

    private func monthLabel(_ value: String) -> String {
        MPFormat.monthLabel(value)
    }

    private func signedAmount(_ row: TransactionModel) -> Int64 {
        if row.type == "income" { return row.amount }
        if row.type == "adjustment" { return row.adjustmentDirection == "increase" ? row.amount : -row.amount }
        return -row.amount
    }

    private func deleteTransaction(_ transaction: TransactionModel) async {
        do {
            try await api.deleteTransaction(id: transaction.id)
            await load()
        } catch {
            self.error = "Không thể xóa giao dịch."
        }
    }

    private func deleteTransfer(_ transaction: TransactionModel) async {
        guard let transferID = transaction.transferID else { return }
        do {
            try await api.deleteTransfer(id: transferID)
            await load()
        } catch {
            self.error = "Không thể xóa cặp giao dịch chuyển ví."
        }
    }

    private func load() async {
        guard !loading else { return }
        loading = true
        detailsReady = false
        defer { loading = false }
        do {
            async let walletResult = api.wallets()
            if wallet.type == "credit" {
                statement = try await api.creditStatement(walletID: wallet.id)
            } else {
                async let transactionResult = api.transactions()
                async let categoryResult = api.categories()
                transactions = try await transactionResult
                categories = try await categoryResult
            }
            if let refreshedWallets = try? await walletResult {
                wallets = refreshedWallets
                if let refreshed = refreshedWallets.first(where: { $0.id == wallet.id }) {
                    wallet = refreshed
                }
            }
            detailsReady = wallet.type == "credit" || !categories.isEmpty
            error = nil
        } catch {
            detailsReady = false
            self.error = "Vui lòng thử lại."
        }
    }
}

struct WalletEditorView: View {
    let api: FinanceAPI
    let wallet: WalletModel?
    let timezone: String
    let onDelete: (() -> Void)?
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
    @State private var goalNotificationsEnabled: Bool
    @State private var saving = false
    @State private var errorMessage: String?

    init(api: FinanceAPI, wallet: WalletModel?, timezone: String = TimeZone.current.identifier, onDelete: (() -> Void)? = nil, onSaved: @escaping () -> Void, onCancel: @escaping () -> Void) {
        self.api = api
        self.wallet = wallet
        self.timezone = timezone
        self.onDelete = onDelete
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
        _goalNotificationsEnabled = State(initialValue: wallet?.goalNotificationsEnabled ?? false)
    }

    var body: some View {
        VStack(spacing: 0) {
            ScrollView {
                VStack(alignment: .leading, spacing: 16) {
                    MPReferenceScreenHeader(title: wallet == nil ? "Thêm ví" : "Sửa ví", onClose: onCancel)
                        .accessibilityIdentifier("wallet-editor-header")
                    MPSurfaceCard(style: .reference) {
                        VStack(alignment: .leading, spacing: 14) {
                            MPTextField(title: "Tên ví", text: $name, placeholder: "Ví tiền mặt", accessibilityIdentifier: "wallet-editor-name")
                            if wallet == nil {
                                MPWalletSelector(title: "Loại ví", options: ["basic", "goal", "credit"], selection: $type)
                                    .accessibilityIdentifier("wallet-editor-type")
                                MPAmountField(title: "Số dư ban đầu", text: $openingBalance)
                                    .accessibilityIdentifier("wallet-editor-opening-balance")
                            } else {
                                VStack(alignment: .leading, spacing: 6) {
                                    MPText("Loại ví: \(type)", style: .secondary)
                                    MPText("Số dư ban đầu: \(MPFormat.money(wallet?.openingBalance ?? 0))", style: .secondary)
                                }
                                .accessibilityIdentifier("wallet-editor-type-readonly")
                            }
                            if type == "goal" {
                                MPAmountField(title: "Mục tiêu", text: $targetAmount)
                                    .accessibilityIdentifier("wallet-editor-target")
                                MPWalletSelector(title: "Có ngày mục tiêu", options: ["Có", "Không"], selection: $targetDateEnabled)
                                if targetDateEnabled == "Có" { MPDateField(title: "Ngày mục tiêu", date: $targetDate, timeZone: timezone) }
                            }
                            if type == "credit" {
                                MPAmountField(title: "Hạn mức tín dụng", text: $creditLimit)
                                MPAmountField(title: "Dư nợ sao kê", text: $lastStatementBalance)
                                MPTextField(title: "Ngày sao kê (1-31)", text: $statementDay, placeholder: "Không bắt buộc")
                                MPTextField(title: "Ngày đến hạn (1-31)", text: $paymentDueDay, placeholder: "Không bắt buộc")
                            }
                            if type == "goal" {
                                MPBooleanSelector(title: "Bật thông báo tiến độ", value: $goalNotificationsEnabled)
                                MPText("Thông báo trong ứng dụng sẽ xuất hiện khi có sự kiện phù hợp.", style: .secondary)
                            }
                            MPWalletSelector(title: "Tính vào tổng cộng", options: ["Có", "Không"], selection: $isInTotal)
                            MPTextField(title: "Mô tả", text: $description, placeholder: "Không bắt buộc", accessibilityIdentifier: "wallet-editor-description")
                        }
                    }
                    if let errorMessage { MPStatusMessage(text: errorMessage, tone: .danger) }
                    if let onDelete {
                        MPButton("Xóa ví", style: .destructive, action: onDelete)
                            .accessibilityIdentifier("wallet-editor-delete")
                    }
                }
                .padding(MPMetrics.pagePadding)
            }
            MPButton("Lưu", isLoading: saving) {
                Task { await save() }
            }
            .accessibilityIdentifier("wallet-editor-save")
            // Keep the established route-test labels while the visible
            // control follows the Stitch copy.
            .accessibilityLabel(wallet == nil ? "Tạo ví" : "Lưu thay đổi")
            .padding(.horizontal, MPMetrics.pagePadding)
            .padding(.vertical, 12)
            .background(MPColor.canvas)
            .safeAreaPadding(.bottom, 4)
        }
        .environment(\.timeZone, TimeZone(identifier: timezone) ?? .current)
        .safeAreaInset(edge: .bottom) {
            MPBottomNavigationInset()
        }
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
        let input = WalletModelInput(name: name.trimmingCharacters(in: .whitespacesAndNewlines), type: wallet?.type ?? type, openingBalance: wallet?.openingBalance ?? (parsedOpeningBalance ?? 0), isInTotal: isInTotal == "Có", description: trimmedDescription.isEmpty ? nil : trimmedDescription, targetAmount: type == "goal" ? parsedTarget : nil, targetDate: type == "goal" && targetDateEnabled == "Có" ? MPFormat.calendarDateString(targetDate, timeZone: TimeZone(identifier: timezone) ?? .current) : nil, creditLimit: type == "credit" ? parsedCreditLimit : nil, lastStatementBalance: type == "credit" ? parsedStatementBalance : nil, statementDay: type == "credit" ? parsedStatementDay : nil, paymentDueDay: type == "credit" ? parsedPaymentDueDay : nil, goalNotificationsEnabled: type == "goal" ? goalNotificationsEnabled : false)
        do {
            if let wallet {
                _ = try await api.updateWallet(id: wallet.id, input: input)
            } else {
                _ = try await api.createWallet(input)
            }
            onSaved()
        } catch {
            errorMessage = "Không thể lưu ví. Kiểm tra thông tin và thử lại."
        }
    }
}
