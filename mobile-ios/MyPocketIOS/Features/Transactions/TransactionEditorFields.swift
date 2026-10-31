import SwiftUI

/// Shared field language for ordinary transaction proposals and the native
/// transaction editor. The parent owns persistence and jar loading; this base
/// owns only policy-filtered controls and their density.
struct TransactionEditorFields: View {
    @Binding var amount: String
    @Binding var type: String
    @Binding var walletID: String
    @Binding var categoryID: String
    @Binding var jarID: String
    @Binding var includedInReports: Bool
    @Binding var date: Date
    @Binding var note: String
    let wallets: [WalletModel]
    let categories: [CategoryModel]
    let jarSummary: JarMonthSummaryModel?
    let jarsLoading: Bool
    let jarLoadError: String?
    let timeZone: TimeZone

    private var selectedWallet: WalletModel? { wallets.first(where: { $0.id == walletID }) }
    private var selectedCategory: CategoryModel? { categories.first(where: { $0.id == categoryID }) }
    private var canUseJar: Bool {
        TransactionEditorPolicy.canUseJar(type: type, category: selectedCategory, walletType: selectedWallet?.type ?? "")
    }
    private var categoryOptions: [MPSelectionOption] {
        var options = TransactionEditorPolicy
            .applicableCategories(type: type, walletID: walletID, walletType: selectedWallet?.type ?? "basic", categories: categories)
            .map { MPSelectionOption(id: $0.id, label: $0.name) }
        if !categoryID.isEmpty, !options.contains(where: { $0.id == categoryID }), let current = selectedCategory {
            options.append(MPSelectionOption(id: current.id, label: current.name))
        }
        return options
    }

    var body: some View {
        MPAmountField(title: "Số tiền", text: $amount)
        MPOptionSelector(title: "Loại giao dịch", options: [MPSelectionOption(id: "income", label: "Thu nhập"), MPSelectionOption(id: "expense", label: "Chi tiêu")], selectionID: $type, emptyLabel: "Chưa xác định — chọn loại")
        MPOptionSelector(title: "Ví", options: TransactionsViewModel.ordinaryTransactionWallets(wallets).map { MPSelectionOption(id: $0.id, label: $0.name) }, selectionID: $walletID)
        MPOptionSelector(title: "Danh mục", options: categoryOptions, selectionID: $categoryID)
        if canUseJar {
            MPOptionSelector(
                title: "Hũ chi tiêu",
                options: TransactionEditorPolicy.jarOptions(summary: jarSummary, selectedJarID: jarID, selectedJarName: nil),
                selectionID: $jarID,
                emptyLabel: jarsLoading ? "Đang tải hũ..." : "Không chọn"
            )
            if let jarLoadError { MPStatusMessage(text: jarLoadError, tone: .danger) }
        }
        MPBooleanSelector(title: "Tính vào báo cáo", value: $includedInReports, trueLabel: "Có", falseLabel: "Không")
        MPDateTimeField(title: "Ngày và giờ", date: $date, timeZone: timeZone.identifier)
        MPTextEditor(title: "Ghi chú", text: $note)
    }
}
