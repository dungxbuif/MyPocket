import SwiftUI

struct MPDivider: View {
    var body: some View {
        Rectangle()
            .fill(MPColor.line)
            .frame(height: MPMetrics.inputBorderWidth)
            .accessibilityHidden(true)
    }
}

/// Shared accessible checkbox row. Feature screens provide only the label,
/// selected state and toggle action; hit target and checkmark treatment stay
/// consistent across multi-select forms.
struct MPCheckboxRow: View {
    let title: String
    let isSelected: Bool
    let accessibilityIdentifier: String
    let action: () -> Void

    var body: some View {
        MPButton(style: .plain, action: action) {
            HStack(spacing: 12) {
                MPText(title, style: .body)
                Spacer(minLength: 8)
                MPIcon(
                    systemName: isSelected ? "checkmark.square.fill" : "square",
                    color: isSelected ? MPColor.action : MPColor.secondary,
                    size: 22
                )
            }
            .frame(minHeight: MPMetrics.hitTarget)
            .contentShape(Rectangle())
        }
        .accessibilityLabel(title)
        .accessibilityValue(isSelected ? "Đã chọn" : "Chưa chọn")
        .accessibilityIdentifier(accessibilityIdentifier)
    }
}

/// Visible multi-select card for category-to-wallet applicability. In this
/// domain an empty API `wallet_ids` set means all wallets, so the card starts
/// with every row selected and normalizes an all-selected set back to empty.
struct MPWalletApplicabilityCard: View {
    let wallets: [WalletModel]
    @Binding var selectedWalletIDs: Set<String>
    var title: String = "Ví áp dụng"
    var allWalletsLabel: String = "Áp dụng tất cả ví"

    var body: some View {
        MPSurfaceCard {
            VStack(alignment: .leading, spacing: 8) {
                MPText(title, style: .emphasis)
                MPText(selectionSummary, style: .secondary)
                    .accessibilityIdentifier("category-wallet-selection-summary")
                if wallets.isEmpty {
                    MPText("Chưa có ví nào.", style: .secondary)
                } else {
                    ForEach(Array(wallets.enumerated()), id: \.element.id) { index, wallet in
                        if index > 0 { MPDivider() }
                        MPCheckboxRow(
                            title: wallet.name,
                            isSelected: selectedWalletIDs.isEmpty || selectedWalletIDs.contains(wallet.id),
                            accessibilityIdentifier: "category-wallet-option-\(wallet.id)"
                        ) {
                            toggle(wallet.id)
                        }
                    }
                }
            }
        }
    }

    private var selectionSummary: String {
        selectedWalletIDs.isEmpty ? allWalletsLabel : "Đã chọn \(selectedWalletIDs.count) ví"
    }

    private func toggle(_ walletID: String) {
        let allWalletIDs = Set(wallets.map(\.id))
        if selectedWalletIDs.isEmpty {
            selectedWalletIDs = allWalletIDs.subtracting([walletID])
        } else if selectedWalletIDs.remove(walletID) == nil {
            selectedWalletIDs.insert(walletID)
        }
        if selectedWalletIDs.isEmpty || selectedWalletIDs == allWalletIDs {
            selectedWalletIDs = []
        }
    }
}

struct MPWalletManagementSection: Equatable, Identifiable {
    let id: String
    let title: String
    let wallets: [WalletModel]
}

enum MPWalletManagementPolicy {
    static func sections(from wallets: [WalletModel]) -> [MPWalletManagementSection] {
        let included = wallets.filter(\.isInTotal)
        let excluded = wallets.filter { !$0.isInTotal }
        return [
            MPWalletManagementSection(id: "included", title: "TÍNH VÀO TỔNG", wallets: included),
            MPWalletManagementSection(id: "excluded", title: "KHÔNG TÍNH VÀO TỔNG", wallets: excluded)
        ].filter { !$0.wallets.isEmpty }
    }

    static func totalBalance(from wallets: [WalletModel]) -> Int64 {
        wallets.filter(\.isInTotal).reduce(0) { $0 + $1.currentBalance }
    }
}

/// Shared wallet-management composition. The screen owns navigation and
/// presentation state; this base owns aggregate/group rows and their stable
/// accessibility contract.
struct MPWalletSelectionList: View {
    let wallets: [WalletModel]
    let isEditing: Bool
    let onSelect: (WalletModel?) -> Void
    let onDetail: (WalletModel) -> Void
    let onEdit: (WalletModel) -> Void
    let onAdd: () -> Void
    @Environment(\.mpFinancialValuesMasked) private var masked
    @State private var selectedWalletID: String?

    static func balanceLabel(_ value: Int64, masked: Bool) -> String {
        MPFormat.money(value, masked: masked)
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 16) {
            MPSurfaceCard(style: .reference) {
                MPButton(style: .plain, isDisabled: isEditing, action: {
                    selectedWalletID = nil
                    onSelect(nil)
                }) {
                    HStack(spacing: 14) {
                        MPWalletAggregateGlyph()
                        VStack(alignment: .leading, spacing: 4) {
                            MPText("Tổng cộng", style: .emphasis)
                                .accessibilityIdentifier("wallet-total-title")
                            MPText(Self.balanceLabel(MPWalletManagementPolicy.totalBalance(from: wallets), masked: masked), style: .secondary)
                                .accessibilityIdentifier("wallet-total-balance")
                        }
                        Spacer(minLength: 8)
                        if selectedWalletID == nil && !isEditing {
                            selectionCheck(identifier: "wallet-selection-total-check")
                        }
                    }
                }
                .accessibilityIdentifier("wallet-total-row")
                .accessibilityValue(selectedWalletID == nil ? "Đang chọn" : "Chưa chọn")
            }

            ForEach(MPWalletManagementPolicy.sections(from: wallets)) { section in
                VStack(alignment: .leading, spacing: 8) {
                    MPText(section.title, style: .secondary)
                        .accessibilityIdentifier("wallet-group-\(section.id)")
                    MPSurfaceCard {
                        VStack(spacing: 0) {
                            ForEach(Array(section.wallets.enumerated()), id: \.element.id) { index, wallet in
                                if index > 0 {
                                    MPDivider()
                                }
                                walletRow(wallet)
                            }
                        }
                    }
                }
            }

            MPSurfaceCard {
                VStack(spacing: 0) {
                    MPButton(style: .plain, fullWidth: true, action: onAdd) {
                        HStack(spacing: 12) {
                            Image(systemName: "plus.circle")
                            MPText("Thêm ví", style: .action)
                            Spacer()
                        }
                    }
                    .accessibilityIdentifier("wallet-add-button")
                    MPDivider()
                    MPButton(style: .plain, isDisabled: true, action: {}) {
                        HStack(spacing: 12) {
                            Image(systemName: "link")
                            VStack(alignment: .leading, spacing: 2) {
                                MPText("Liên kết dịch vụ", style: .secondary)
                                MPText("Sắp ra mắt", style: .secondary)
                            }
                            Spacer()
                        }
                    }
                    .accessibilityIdentifier("wallet-link-service")
                }
            }
        }
    }

    @ViewBuilder
    private func walletRow(_ wallet: WalletModel) -> some View {
        VStack(alignment: .leading, spacing: 4) {
            HStack(spacing: 4) {
                MPButton(style: .plain, fullWidth: true, action: {
                    if isEditing {
                        onEdit(wallet)
                    } else {
                        selectedWalletID = wallet.id
                        onSelect(wallet)
                    }
                }) {
                    MPWalletSelectionRow(
                        walletID: wallet.id,
                        name: wallet.name,
                        type: wallet.type,
                        balance: Self.balanceLabel(wallet.currentBalance, masked: masked),
                        isNegative: wallet.currentBalance < 0,
                        isSelected: selectedWalletID == wallet.id && !isEditing,
                        checkIdentifier: "wallet-selection-check-\(wallet.id)"
                    )
                }
                .accessibilityIdentifier("wallet-scope-\(wallet.id)")
                .accessibilityValue(selectedWalletID == wallet.id && !isEditing ? "Đang chọn" : "Chưa chọn")

                MPTextLinkButton("Chi tiết") { onDetail(wallet) }
                    .accessibilityIdentifier("wallet-detail-\(wallet.id)")
            }
            if wallet.type == "goal", let target = wallet.targetAmount {
                MPProgressBar(progress: Double(wallet.currentBalance) / Double(max(1, target)), label: "Mục tiêu \(MPFormat.money(target, masked: masked))")
            }
            if wallet.type == "credit", let available = wallet.availableCredit {
                MPText("Khả dụng: \(MPFormat.money(available, masked: masked))", style: .secondary)
            }
        }
        .padding(.vertical, 4)
    }

    private func selectionCheck(identifier: String) -> some View {
        Image(systemName: "checkmark")
            .font(MPFont.bodyEmphasis)
            .foregroundStyle(MPColor.action)
            .accessibilityIdentifier(identifier)
            .accessibilityHidden(true)
    }
}

/// Selection-list anatomy follows the Stitch/web row: wallet identity at the
/// leading edge, title over current balance, and a trailing scope check.
private struct MPWalletSelectionRow: View {
    let walletID: String
    let name: String
    let type: String
    let balance: String
    let isNegative: Bool
    let isSelected: Bool
    let checkIdentifier: String

    var body: some View {
        HStack(spacing: 12) {
            MPWalletGlyph(type: type, size: 40)
            VStack(alignment: .leading, spacing: 4) {
                MPText(name, style: .emphasis)
                    .lineLimit(2)
                    .fixedSize(horizontal: false, vertical: true)
                    .accessibilityIdentifier("wallet-name-\(walletID)")
                MPText(balance, style: isNegative ? .danger : .secondary)
                    .lineLimit(1)
                    .fixedSize(horizontal: true, vertical: false)
                    .accessibilityIdentifier("wallet-balance-\(walletID)")
            }
            Spacer(minLength: 4)
            if isSelected {
                Image(systemName: "checkmark")
                    .font(MPFont.bodyEmphasis)
                    .foregroundStyle(MPColor.action)
                    .accessibilityIdentifier(checkIdentifier)
                    .accessibilityHidden(true)
            }
        }
    }

}

/// Wallet-management row keeps the amount readable at narrow phone widths;
/// secondary actions live on their own line in the selection-list base.
struct MPWalletManagementRow: View {
    let name: String
    let type: String
    let balance: String
    let subtitle: String?
    let isNegative: Bool

    var body: some View {
        HStack(spacing: 12) {
            MPWalletGlyph(type: type, size: 40)
            VStack(alignment: .leading, spacing: 4) {
                MPText(name, style: .emphasis)
                    .lineLimit(2)
                    .fixedSize(horizontal: false, vertical: true)
                MPText(MPWalletRowPolicy.subtitle(description: subtitle, type: type), style: .secondary)
                    .lineLimit(1)
                    .fixedSize(horizontal: false, vertical: true)
            }
            Spacer(minLength: 4)
            MPText(balance, style: isNegative ? .danger : .emphasis)
                .lineLimit(1)
                .fixedSize(horizontal: true, vertical: false)
        }
    }
}
