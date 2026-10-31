import SwiftUI

enum AssistantEntryBatchDecision: Equatable, Sendable {
    case apply
    case reject
}

enum AssistantEntryBatchPolicy {
    static func initialState(for item: AdvisorEntryBatchItemModel) -> AssistantEntryRowState {
        switch item.status.trimmingCharacters(in: .whitespacesAndNewlines).lowercased() {
        case "applied", "saved", "completed": return .applied
        case "rejected", "skipped", "cancelled": return .rejected
        case "saving", "processing": return .saving
        case "conflict": return .conflict
        case "invalid", "error": return .invalid
        default: return item.draft == nil ? .invalid : .pending
        }
    }

    static func state(after decision: AssistantEntryBatchDecision) -> AssistantEntryRowState {
        switch decision {
        case .apply: return .saving
        case .reject: return .rejected
        }
    }

    static func canApply(_ item: AdvisorEntryBatchItemModel) -> Bool {
        guard let draft = item.draft,
              draft.amount > 0,
              !draft.walletID.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty else {
            return false
        }
        switch item.status.trimmingCharacters(in: .whitespacesAndNewlines).lowercased() {
        case "applied", "saved", "completed", "rejected", "skipped", "cancelled", "conflict":
            return false
        default:
            return true
        }
    }

    static func validItems(_ items: [AdvisorEntryBatchItemModel]) -> [AdvisorEntryBatchItemModel] {
        items.filter(canApply)
    }
}

enum AssistantEntryBatchPresentation {
    @MainActor static func transactionRow(
        for item: AdvisorEntryBatchItemModel,
        timeZone: TimeZone,
        masked: Bool
    ) -> MPTransactionRow? {
        guard let draft = item.draft else { return nil }
        return MPTransactionRow(
            type: draft.type,
            rawAmount: draft.amount,
            note: draft.note,
            occurredAt: draft.occurredAt,
            timeZone: timeZone,
            masked: masked
        )
    }
}

enum AssistantPartPolicy {
    static func isReadOnly(_ part: AdvisorPartModel) -> Bool {
        !allowsMutation(for: part)
    }

    static func allowsMutation(for part: AdvisorPartModel) -> Bool {
        if part.entryBatchEnvelope != nil { return true }
        switch AssistantCardType.normalized(part.type) {
        case "wallet_confirmation", "jar_confirmation", "budget_confirmation":
            return true
        default:
            return false
        }
    }
}

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
                    if AssistantEntryBatchPresentation.transactionRow(for: item, timeZone: timeZone, masked: masked) != nil {
                        VStack(alignment: .leading, spacing: MPMetrics.rowSpacing) {
                            if let row = AssistantEntryBatchPresentation.transactionRow(for: item, timeZone: timeZone, masked: masked) {
                                row
                            }
                            AssistantActionCard(
                                title: item.sourceLabel ?? "Giao dịch đề xuất",
                                summary: summary(for: item.draft!),
                                state: states[item.id] ?? AssistantEntryBatchPolicy.initialState(for: item),
                                onPrimary: {
                                    if AssistantEntryBatchPolicy.canApply(item) {
                                        states[item.id] = AssistantEntryBatchPolicy.state(after: .apply)
                                        onSaveOne(item)
                                    } else {
                                        onEdit(item)
                                    }
                                },
                                onSecondary: {
                                    states[item.id] = AssistantEntryBatchPolicy.state(after: .reject)
                                    onSkip(item)
                                }
                            )
                        }
                        .onTapGesture { onEdit(item) }
                    } else {
                        AssistantActionCard(
                            title: item.sourceLabel ?? "Giao dịch đề xuất",
                            summary: item.error ?? "Chưa đủ dữ liệu để ghi giao dịch.",
                            state: states[item.id] ?? AssistantEntryBatchPolicy.initialState(for: item),
                            onPrimary: { onEdit(item) },
                            onSecondary: {
                                states[item.id] = AssistantEntryBatchPolicy.state(after: .reject)
                                onSkip(item)
                            }
                        )
                    }
                }
                let valid = AssistantEntryBatchPolicy.validItems(batch.items)
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
