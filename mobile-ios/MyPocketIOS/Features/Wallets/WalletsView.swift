import SwiftUI

@MainActor
final class WalletsViewModel: ObservableObject {
    @Published private(set) var wallets: [WalletModel] = []
    @Published private(set) var isLoading = false
    @Published private(set) var errorMessage: String?
    private let api: FinanceAPI

    init(api: FinanceAPI) { self.api = api }

    func load() async {
        guard !isLoading else { return }
        isLoading = true
        errorMessage = nil
        defer { isLoading = false }
        do { wallets = try await api.wallets() }
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

    init(api: FinanceAPI) { _model = StateObject(wrappedValue: WalletsViewModel(api: api)) }

    var body: some View {
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
                                MPWalletRow(name: wallet.name, type: wallet.type, balance: MPFormat.money(wallet.currentBalance))
                                if wallet.type == "goal", let target = wallet.targetAmount {
                                    MPProgressBar(progress: Double(wallet.currentBalance) / Double(max(1, target)), label: "Mục tiêu \(MPFormat.money(target))")
                                }
                                if wallet.type == "credit", let available = wallet.availableCredit {
                                    MPText("Khả dụng: \(MPFormat.money(available))", style: .secondary)
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
        .task { await model.load() }
        .refreshable { await model.load() }
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
}

private extension WalletsViewModel {
    var apiForChild: FinanceAPI { api }
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
    @State private var creditLimit: String
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
        _creditLimit = State(initialValue: wallet?.creditLimit.map(String.init) ?? "")
        _description = State(initialValue: wallet?.description ?? "")
        _isInTotal = State(initialValue: wallet?.isInTotal == false ? "Không" : "Có")
    }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                MPScreenHeader(title: wallet == nil ? "Thêm ví" : "Sửa ví", leading: { MPIconButton(systemName: "xmark", label: "Đóng", action: onCancel) }, trailing: { EmptyView() })
                MPTextField(title: "Tên ví", text: $name, placeholder: "Ví tiền mặt")
                MPWalletSelector(title: "Loại ví", options: ["basic", "goal", "credit"], selection: $type)
                MPAmountField(title: "Số dư ban đầu", text: $openingBalance)
                if type == "goal" {
                    MPAmountField(title: "Mục tiêu", text: $targetAmount)
                    MPDateField(title: "Ngày mục tiêu", date: $targetDate)
                }
                if type == "credit" {
                    MPAmountField(title: "Hạn mức tín dụng", text: $creditLimit)
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
        saving = true
        defer { saving = false }
        let input = WalletModelInput(name: name.trimmingCharacters(in: .whitespacesAndNewlines), type: type, openingBalance: Int64(openingBalance.filter(\.isNumber)) ?? 0, isInTotal: isInTotal == "Có", description: description.isEmpty ? nil : description, targetAmount: Int64(targetAmount.filter(\.isNumber)), targetDate: type == "goal" ? MPFormat.todayISO : nil, creditLimit: Int64(creditLimit.filter(\.isNumber)), lastStatementBalance: nil, statementDay: nil, paymentDueDay: nil)
        do {
            if let wallet { _ = try await api.updateWallet(id: wallet.id, input: input) }
            else { _ = try await api.createWallet(input) }
            onSaved()
        } catch { errorMessage = "Không thể lưu ví. Kiểm tra thông tin và thử lại." }
    }
}
