import SwiftUI

struct AssistantEntryBatchView: View {
    let batch: AdvisorEntryBatchEnvelopeModel
    let timeZone: TimeZone
    let masked: Bool
    let onEdit: (AdvisorEntryBatchItemModel) -> Void
    let onSaveOne: (AdvisorEntryBatchItemModel) -> Void
    let onSaveValid: ([AdvisorEntryBatchItemModel]) -> Void
    let onSkip: (AdvisorEntryBatchItemModel) -> Void
    @State private var states: [String: AssistantEntryRowState] = [:]

    var body: some View {
        VStack(alignment: .leading, spacing: MPMetrics.sectionSpacing) {
                MPSectionHeader(title: "Đề xuất giao dịch")
                MPText(batch.summary ?? (String(batch.items.count) + " giao dịch cần xem lại"), style: .secondary)
                ForEach(batch.items) { item in
                    if let draft = item.draft {
                        VStack(alignment: .leading, spacing: MPMetrics.rowSpacing) {
                            MPTransactionRow(
                                title: draft.type == "income" ? "Khoản thu" : "Khoản chi",
                                subtitle: draft.note,
                                amount: masked ? MPFormat.maskedValue : MPFormat.money(draft.amount),
                                isIncome: draft.type == "income"
                            )
                            AssistantActionCard(
                                title: item.sourceLabel ?? "Giao dịch đề xuất",
                                summary: summary(for: draft),
                                state: states[item.id] ?? .pending,
                                onPrimary: { onSaveOne(item) },
                                onSecondary: { states[item.id] = .rejected; onSkip(item) }
                            )
                        }
                        .onTapGesture { onEdit(item) }
                    } else {
                        AssistantActionCard(
                            title: item.sourceLabel ?? "Giao dịch đề xuất",
                            summary: item.error ?? "Chưa đủ dữ liệu để ghi giao dịch.",
                            state: .invalid,
                            onPrimary: { onEdit(item) },
                            onSecondary: { states[item.id] = .rejected; onSkip(item) }
                        )
                    }
                }
                let valid = batch.items.filter { item in
                    guard let draft = item.draft else { return false }
                    return draft.amount > 0 && !draft.walletID.isEmpty
                }
                MPButton("Ghi các giao dịch hợp lệ", style: .primary, isDisabled: valid.isEmpty) {
                    onSaveValid(valid)
                }
        }
        .accessibilityIdentifier("assistant-entry-batch")
    }

    private func summary(for draft: EntryDraftModel) -> String {
        let date = MPFormat.shortDate(draft.occurredAt, timeZone: timeZone)
        let scope = draft.jarID == nil ? "Không gắn hũ" : "Có gắn hũ"
        return date + " · " + scope + " · " + (draft.note.isEmpty ? "Chưa có ghi chú" : draft.note)
    }
}
